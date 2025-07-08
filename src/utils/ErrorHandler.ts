/**
 * Standardized error handling utility with TypeScript support
 * Provides unified error management across the GoldBox RPG client application
 */

import type { ErrorMetadata } from '../types/GameTypes';
import type { EventEmitterInterface } from '../types/UITypes';
import { logger } from './Logger';

export interface ErrorHandlerOptions {
  readonly component: string;
  readonly eventEmitter?: EventEmitterInterface | undefined;
  readonly userMessageCallback?: ((message: string, type: 'error' | 'warning' | 'info') => void) | undefined;
  readonly enableStackTrace?: boolean;
  readonly enableMetadataLogging?: boolean;
}

export interface ErrorContext {
  readonly method: string;
  readonly timestamp: number;
  readonly metadata?: ErrorMetadata | undefined;
  readonly stackTrace?: string | undefined;
}

export class ErrorHandler {
  private readonly component: string;
  private readonly eventEmitter?: EventEmitterInterface | undefined;
  private readonly userMessageCallback?: ((message: string, type: 'error' | 'warning' | 'info') => void) | undefined;
  private readonly enableStackTrace: boolean;
  private readonly enableMetadataLogging: boolean;
  private readonly componentLogger: typeof logger;

  constructor(options: ErrorHandlerOptions) {
    this.component = options.component;
    this.eventEmitter = options.eventEmitter;
    this.userMessageCallback = options.userMessageCallback;
    this.enableStackTrace = options.enableStackTrace ?? true;
    this.enableMetadataLogging = options.enableMetadataLogging ?? true;
    
    // Create component-specific logger
    this.componentLogger = logger.createChildLogger(this.component);
  }

  /**
   * Handles recoverable errors that should not stop execution
   * Logs error, emits event, and optionally shows user message
   */
  handleRecoverableError(
    error: Error | string,
    context: string,
    userMessage?: string,
    metadata: ErrorMetadata = {}
  ): void {
    const errorObj = this.normalizeError(error);
    const errorContext = this.createErrorContext(context, metadata);
    
    // Always log with context
    this.logError(errorObj, errorContext, 'warn');
    
    // Emit error event if event emitter is available
    if (this.eventEmitter) {
      this.eventEmitter.emit('error', {
        error: errorObj,
        context: errorContext,
        recoverable: true,
        userMessage
      });
    }
    
    // Show user message if callback is provided
    if (userMessage && this.userMessageCallback) {
      this.userMessageCallback(userMessage, 'warning');
    }
  }

  /**
   * Handles critical errors that should stop execution
   * Logs error and throws it to stop execution flow
   */
  handleCriticalError(
    error: Error | string,
    context: string,
    metadata: ErrorMetadata = {}
  ): never {
    const errorObj = this.normalizeError(error);
    const errorContext = this.createErrorContext(context, metadata);
    
    // Log as error level
    this.logError(errorObj, errorContext, 'error');
    
    // Emit error event if event emitter is available
    if (this.eventEmitter) {
      this.eventEmitter.emit('error', {
        error: errorObj,
        context: errorContext,
        recoverable: false
      });
    }
    
    // Show critical error message
    if (this.userMessageCallback) {
      this.userMessageCallback(
        `Critical error in ${this.component}: ${errorObj.message}`,
        'error'
      );
    }
    
    // Always throw to stop execution
    throw errorObj;
  }

  /**
   * Handles initialization errors with cleanup
   * Logs error, attempts cleanup, and throws
   */
  handleInitializationError(
    error: Error | string,
    context: string,
    cleanupFn?: () => void,
    metadata: ErrorMetadata = {}
  ): never {
    const errorObj = this.normalizeError(error);
    const errorContext = this.createErrorContext(context, metadata);
    
    this.componentLogger.error(
      `Initialization failed in ${context}:`,
      errorObj,
      errorContext
    );
    
    // Attempt cleanup if provided
    if (cleanupFn) {
      try {
        cleanupFn();
        this.componentLogger.info('Cleanup completed after initialization failure');
      } catch (cleanupError) {
        this.componentLogger.error('Cleanup failed:', cleanupError);
      }
    }
    
    // Emit initialization error event
    if (this.eventEmitter) {
      this.eventEmitter.emit('initializationError', {
        error: errorObj,
        context: errorContext
      });
    }
    
    throw errorObj;
  }

  /**
   * Wraps async operations with standardized error handling
   */
  wrapAsync<T, TArgs extends readonly unknown[]>(
    asyncFn: (...args: TArgs) => Promise<T>,
    context: string,
    options: {
      readonly userMessage?: string;
      readonly critical?: boolean;
      readonly metadata?: ErrorMetadata;
      readonly onError?: (error: Error) => void;
    } = {}
  ): (...args: TArgs) => Promise<T> {
    return async (...args: TArgs): Promise<T> => {
      try {
        return await asyncFn(...args);
      } catch (error) {
        const normalizedError = this.normalizeError(error);
        
        // Call custom error handler if provided
        if (options.onError) {
          try {
            options.onError(normalizedError);
          } catch (handlerError) {
            this.componentLogger.error('Error in custom error handler:', handlerError);
          }
        }
        
        // Handle based on criticality
        if (options.critical) {
          this.handleCriticalError(normalizedError, context, options.metadata);
        } else {
          this.handleRecoverableError(
            normalizedError,
            context,
            options.userMessage,
            options.metadata
          );
          throw normalizedError; // Re-throw for async flow control
        }
      }
    };
  }

  /**
   * Creates a safe wrapper for synchronous operations
   */
  wrapSync<T, TArgs extends readonly unknown[]>(
    syncFn: (...args: TArgs) => T,
    context: string,
    options: {
      readonly userMessage?: string;
      readonly critical?: boolean;
      readonly metadata?: ErrorMetadata;
      readonly defaultValue?: T;
    } = {}
  ): (...args: TArgs) => T {
    return (...args: TArgs): T => {
      try {
        return syncFn(...args);
      } catch (error) {
        const normalizedError = this.normalizeError(error);
        
        if (options.critical) {
          this.handleCriticalError(normalizedError, context, options.metadata);
        } else {
          this.handleRecoverableError(
            normalizedError,
            context,
            options.userMessage,
            options.metadata
          );
          
          if (options.defaultValue !== undefined) {
            return options.defaultValue;
          }
          throw normalizedError;
        }
      }
    };
  }

  /**
   * Normalizes different error types to Error objects
   */
  private normalizeError(error: unknown): Error {
    if (error instanceof Error) {
      return error;
    }
    
    if (typeof error === 'string') {
      return new Error(error);
    }
    
    if (error && typeof error === 'object' && 'message' in error) {
      return new Error(String(error.message));
    }
    
    return new Error(`Unknown error: ${String(error)}`);
  }

  /**
   * Creates error context with metadata
   */
  private createErrorContext(method: string, metadata: ErrorMetadata = {}): ErrorContext {
    const stackTrace = this.enableStackTrace ? new Error().stack : undefined;
    
    return {
      method,
      timestamp: Date.now(),
      metadata: this.enableMetadataLogging ? metadata : undefined,
      stackTrace
    };
  }

  /**
   * Logs error with appropriate level and formatting
   */
  private logError(
    error: Error,
    context: ErrorContext,
    level: 'warn' | 'error' = 'error'
  ): void {
    const logMessage = `${this.component}.${context.method}: ${error.message}`;
    
    const logData: Record<string, unknown> = {
      error: {
        name: error.name,
        message: error.message,
        stack: this.enableStackTrace ? error.stack : undefined
      },
      context
    };
    
    if (level === 'error') {
      this.componentLogger.error(logMessage, logData);
    } else {
      this.componentLogger.warn(logMessage, logData);
    }
  }

  /**
   * Gets the component name this error handler is associated with
   */
  getComponent(): string {
    return this.component;
  }

  /**
   * Checks if error handling is configured for user messages
   */
  hasUserMessageHandler(): boolean {
    return this.userMessageCallback !== undefined;
  }

  /**
   * Checks if error handling is configured for event emission
   */
  hasEventEmitter(): boolean {
    return this.eventEmitter !== undefined;
  }
}

/**
 * Creates a new error handler with the specified configuration
 */
export function createErrorHandler(options: ErrorHandlerOptions): ErrorHandler {
  return new ErrorHandler(options);
}

/**
 * Global error handler factory for common use cases
 */
export class GlobalErrorHandler {
  private static readonly handlers = new Map<string, ErrorHandler>();
  
  /**
   * Gets or creates an error handler for a component
   */
  static getHandler(component: string, options?: Partial<ErrorHandlerOptions>): ErrorHandler {
    if (!this.handlers.has(component)) {
      this.handlers.set(component, new ErrorHandler({
        component,
        ...options
      }));
    }
    
    return this.handlers.get(component)!;
  }
  
  /**
   * Sets up global error handlers for unhandled errors
   */
  static setupGlobalHandlers(): void {
    if (typeof window !== 'undefined') {
      // Handle unhandled promise rejections
      window.addEventListener('unhandledrejection', (event) => {
        const handler = this.getHandler('GlobalPromiseRejection');
        handler.handleRecoverableError(
          event.reason,
          'unhandledPromiseRejection',
          'An unexpected error occurred',
          { promise: event.promise }
        );
      });
      
      // Handle uncaught errors
      window.addEventListener('error', (event) => {
        const handler = this.getHandler('GlobalError');
        handler.handleRecoverableError(
          event.error || new Error(event.message),
          'uncaughtError',
          'An unexpected error occurred',
          {
            filename: event.filename,
            lineno: event.lineno,
            colno: event.colno
          }
        );
      });
    }
  }
  
  /**
   * Clears all cached error handlers
   */
  static clearHandlers(): void {
    this.handlers.clear();
  }
}
