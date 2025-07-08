/**
 * RPC Client for GoldBox RPG Engine
 * Provides WebSocket-based JSON-RPC 2.0 communication with the server
 */

import { TypedEventEmitter } from '../core/EventEmitter';
import type { 
  RPCRequest,
  RPCResponse,
  RPCMethod,
  SessionInfo
} from '../types/RPCTypes';
import { logger } from '../utils/Logger';

interface RPCEventMap extends Record<string, unknown> {
  connected: void;
  disconnected: { reason: string };
  error: { error: Error };
  sessionExpired: void;
  reconnecting: { attempt: number };
  message: { data: unknown };
}

interface PendingRequest {
  readonly resolve: (value: unknown) => void;
  readonly reject: (reason: Error) => void;
  readonly timestamp: number;
  readonly method: RPCMethod;
  readonly timeout: number;
}

/**
 * Configuration options for RPC client
 */
export interface RPCClientConfig {
  readonly baseUrl?: string;
  readonly maxReconnectAttempts?: number;
  readonly connectionTimeout?: number;
  readonly requestTimeout?: number;
  readonly reconnectDelay?: number;
  readonly enableLogging?: boolean;
}

/**
 * WebSocket-based JSON-RPC 2.0 client with automatic reconnection
 */
export class RPCClient extends TypedEventEmitter<RPCEventMap> {
  private readonly config: Required<RPCClientConfig>;
  private readonly clientLogger = logger.createChildLogger('RPCClient');
  private readonly requestQueue = new Map<number, PendingRequest>();
  
  private ws: WebSocket | null = null;
  private sessionId: string | null = null;
  private sessionExpiry: number | null = null;
  private requestId = 1;
  private reconnectAttempts = 0;
  private reconnectTimer: number | null = null;
  private isConnecting = false;
  private isDestroyed = false;

  constructor(config: RPCClientConfig = {}) {
    super();
    
    this.config = {
      baseUrl: './rpc',
      maxReconnectAttempts: 5,
      connectionTimeout: 10000,
      requestTimeout: 30000,
      reconnectDelay: 1000,
      enableLogging: true,
      ...config
    };

    this.clientLogger.info('RPC Client initialized', { config: this.config });
  }

  /**
   * Connect to the RPC server via WebSocket
   */
  async connect(): Promise<void> {
    if (this.isDestroyed) {
      throw new Error('Cannot connect destroyed RPC client');
    }

    if (this.isConnected() || this.isConnecting) {
      this.clientLogger.warn('Already connected or connecting');
      return;
    }

    this.isConnecting = true;
    this.clientLogger.info('Establishing WebSocket connection');

    try {
      await this.establishConnection();
      this.reconnectAttempts = 0;
      this.isConnecting = false;
      this.emit('connected', undefined);
      this.clientLogger.info('Successfully connected to RPC server');
    } catch (error) {
      this.isConnecting = false;
      this.clientLogger.error('Failed to connect:', error);
      await this.handleConnectionError(error as Error);
      throw error;
    }
  }

  /**
   * Disconnect from the RPC server
   */
  disconnect(reason = 'Client disconnect'): void {
    this.clientLogger.info('Disconnecting RPC client', { reason });

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close(1000, reason);
      this.ws = null;
    }

    this.sessionId = null;
    this.sessionExpiry = null;
    this.isConnecting = false;

    // Reject all pending requests
    this.rejectPendingRequests(new Error(`Connection closed: ${reason}`));

    this.emit('disconnected', { reason });
  }

  /**
   * Destroy the RPC client and clean up resources
   */
  destroy(): void {
    this.isDestroyed = true;
    this.disconnect('Client destroyed');
    this.removeAllListeners();
    this.clientLogger.info('RPC client destroyed');
  }

  /**
   * Check if currently connected to server
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Get current session information
   */
  getSession(): SessionInfo | null {
    if (!this.sessionId || !this.sessionExpiry) {
      return null;
    }

    return {
      sessionId: this.sessionId,
      expiresAt: new Date(this.sessionExpiry),
      isValid: Date.now() < this.sessionExpiry
    };
  }

  /**
   * Send an RPC request to the server
   */
  async call<T = unknown>(
    method: RPCMethod,
    params?: Record<string, unknown>,
    timeout?: number
  ): Promise<T> {
    if (this.isDestroyed) {
      throw new Error('Cannot call method on destroyed RPC client');
    }

    if (!this.isConnected()) {
      throw new Error('Not connected to RPC server');
    }

    const id = this.requestId++;
    const baseParams = params || {};
    
    // Add session ID if available
    const requestParams = this.sessionId 
      ? { ...baseParams, sessionId: this.sessionId }
      : baseParams;
    
    const request: RPCRequest = {
      jsonrpc: '2.0',
      method,
      params: requestParams,
      id
    };

    const requestTimeout = timeout || this.config.requestTimeout;

    return new Promise<T>((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.requestQueue.delete(id);
        reject(new Error(`Request timeout after ${requestTimeout}ms`));
      }, requestTimeout);

      this.requestQueue.set(id, {
        resolve: (value: unknown) => {
          clearTimeout(timeoutId);
          resolve(value as T);
        },
        reject: (error: Error) => {
          clearTimeout(timeoutId);
          reject(error);
        },
        timestamp: Date.now(),
        method,
        timeout: requestTimeout
      });

      try {
        this.ws!.send(JSON.stringify(request));
        this.clientLogger.debug('Sent RPC request', { method, id });
      } catch (error) {
        this.requestQueue.delete(id);
        clearTimeout(timeoutId);
        reject(error);
      }
    });
  }

  /**
   * Establish WebSocket connection with proper error handling
   */
  private async establishConnection(): Promise<void> {
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${location.host}/rpc/ws`;
    
    this.clientLogger.debug('Connecting to WebSocket', { wsUrl });
    
    this.ws = new WebSocket(wsUrl);
    this.setupWebSocketHandlers();
    
    return this.waitForConnection();
  }

  /**
   * Set up WebSocket event handlers
   */
  private setupWebSocketHandlers(): void {
    if (!this.ws) return;

    this.ws.addEventListener('open', () => {
      this.clientLogger.info('WebSocket connection opened');
    });

    this.ws.addEventListener('message', (event) => {
      this.handleMessage(event.data);
    });

    this.ws.addEventListener('close', (event) => {
      this.clientLogger.info('WebSocket connection closed', { 
        code: event.code, 
        reason: event.reason 
      });
      
      if (!this.isDestroyed && event.code !== 1000) {
        // Unexpected close - attempt reconnection
        this.handleConnectionError(new Error(`Connection closed unexpectedly: ${event.reason}`));
      }
    });

    this.ws.addEventListener('error', (event) => {
      this.clientLogger.error('WebSocket error:', event);
      this.emit('error', { error: new Error('WebSocket error') });
    });
  }

  /**
   * Wait for WebSocket connection to be established
   */
  private waitForConnection(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.ws) {
        reject(new Error('No WebSocket instance'));
        return;
      }

      if (this.ws.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      const timeout = setTimeout(() => {
        reject(new Error(`Connection timeout after ${this.config.connectionTimeout}ms`));
      }, this.config.connectionTimeout);

      const openHandler = () => {
        clearTimeout(timeout);
        resolve();
      };

      const errorHandler = () => {
        clearTimeout(timeout);
        reject(new Error('WebSocket connection failed'));
      };

      this.ws.addEventListener('open', openHandler, { once: true });
      this.ws.addEventListener('error', errorHandler, { once: true });
    });
  }

  /**
   * Handle incoming WebSocket messages
   */
  private handleMessage(data: string): void {
    try {
      const message = JSON.parse(data) as RPCResponse;
      this.clientLogger.debug('Received RPC message', { id: message.id });

      if ('id' in message && message.id !== null) {
        // Response to a request
        this.handleResponse(message);
      } else {
        // Server notification
        this.handleNotification(message);
      }
    } catch (error) {
      this.clientLogger.error('Failed to parse message:', error);
    }
  }

  /**
   * Handle RPC response messages
   */
  private handleResponse(response: RPCResponse): void {
    const pendingRequest = this.requestQueue.get(response.id as number);
    
    if (!pendingRequest) {
      this.clientLogger.warn('Received response for unknown request', { id: response.id });
      return;
    }

    this.requestQueue.delete(response.id as number);

    if ('error' in response && response.error) {
      const error = new Error(response.error.message);
      (error as any).code = response.error.code;
      (error as any).data = response.error.data;
      pendingRequest.reject(error);
    } else {
      pendingRequest.resolve(response.result);
    }
  }

  /**
   * Handle server notifications
   */
  private handleNotification(notification: RPCResponse): void {
    this.emit('message', { data: notification.result });
    
    // Handle specific server notifications
    if (typeof notification.result === 'object' && notification.result) {
      const result = notification.result as Record<string, unknown>;
      
      if (result.type === 'sessionExpired') {
        this.sessionId = null;
        this.sessionExpiry = null;
        this.emit('sessionExpired', undefined);
      }
    }
  }

  /**
   * Handle connection errors with exponential backoff
   */
  private async handleConnectionError(error: Error): Promise<void> {
    this.reconnectAttempts++;
    
    this.clientLogger.warn(`Connection error (attempt ${this.reconnectAttempts}):`, error);
    
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      this.clientLogger.error('Max reconnection attempts reached');
      this.emit('error', { error: new Error('Max reconnection attempts exceeded') });
      return;
    }

    const delay = this.config.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    this.clientLogger.info(`Reconnecting in ${delay}ms...`);
    this.emit('reconnecting', { attempt: this.reconnectAttempts });

    this.reconnectTimer = window.setTimeout(async () => {
      try {
        await this.connect();
      } catch (reconnectError) {
        this.clientLogger.error('Reconnection failed:', reconnectError);
      }
    }, delay);
  }

  /**
   * Reject all pending requests
   */
  private rejectPendingRequests(error: Error): void {
    for (const [, request] of this.requestQueue.entries()) {
      request.reject(error);
    }
    this.requestQueue.clear();
  }
}

// Export singleton instance for global use
export const rpcClient = new RPCClient();
