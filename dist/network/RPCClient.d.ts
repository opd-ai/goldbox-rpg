/**
 * RPC Client for GoldBox RPG Engine
 * Provides WebSocket-based JSON-RPC 2.0 communication with the server
 */
import { TypedEventEmitter } from '../core/EventEmitter';
import type { RPCMethod, SessionInfo } from '../types/RPCTypes';
interface RPCEventMap extends Record<string, unknown> {
    connected: void;
    disconnected: {
        reason: string;
    };
    error: {
        error: Error;
    };
    sessionExpired: void;
    reconnecting: {
        attempt: number;
    };
    message: {
        data: unknown;
    };
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
export declare class RPCClient extends TypedEventEmitter<RPCEventMap> {
    private readonly config;
    private readonly clientLogger;
    private readonly requestQueue;
    private ws;
    private sessionId;
    private sessionExpiry;
    private requestId;
    private reconnectAttempts;
    private reconnectTimer;
    private isConnecting;
    private isDestroyed;
    constructor(config?: RPCClientConfig);
    /**
     * Connect to the RPC server via WebSocket
     */
    connect(): Promise<void>;
    /**
     * Disconnect from the RPC server
     */
    disconnect(reason?: string): void;
    /**
     * Destroy the RPC client and clean up resources
     */
    destroy(): void;
    /**
     * Check if currently connected to server
     */
    isConnected(): boolean;
    /**
     * Get current session information
     */
    getSession(): SessionInfo | null;
    /**
     * Send an RPC request to the server
     */
    call<T = unknown>(method: RPCMethod, params?: Record<string, unknown>, timeout?: number): Promise<T>;
    /**
     * Establish WebSocket connection with proper error handling
     */
    private establishConnection;
    /**
     * Set up WebSocket event handlers
     */
    private setupWebSocketHandlers;
    /**
     * Wait for WebSocket connection to be established
     */
    private waitForConnection;
    /**
     * Handle incoming WebSocket messages
     */
    private handleMessage;
    /**
     * Handle RPC response messages
     */
    private handleResponse;
    /**
     * Handle server notifications
     */
    private handleNotification;
    /**
     * Handle connection errors with exponential backoff
     */
    private handleConnectionError;
    /**
     * Reject all pending requests
     */
    private rejectPendingRequests;
}
export declare const rpcClient: RPCClient;
export {};
//# sourceMappingURL=RPCClient.d.ts.map