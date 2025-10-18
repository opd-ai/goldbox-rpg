/**
 * Standardized error handling utility with TypeScript support
 * Provides unified error management across the GoldBox RPG client application
 */
import type { ErrorMetadata } from '../types/GameTypes';
import type { EventEmitterInterface } from '../types/UITypes';
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
export declare class ErrorHandler {
    private readonly component;
    private readonly eventEmitter?;
    private readonly userMessageCallback?;
    private readonly enableStackTrace;
    private readonly enableMetadataLogging;
    private readonly componentLogger;
    constructor(options: ErrorHandlerOptions);
    /**
     * Handles recoverable errors that should not stop execution
     * Logs error, emits event, and optionally shows user message
     */
    handleRecoverableError(error: Error | string, context: string, userMessage?: string, metadata?: ErrorMetadata): void;
    /**
     * Handles critical errors that should stop execution
     * Logs error and throws it to stop execution flow
     */
    handleCriticalError(error: Error | string, context: string, metadata?: ErrorMetadata): never;
    /**
     * Handles initialization errors with cleanup
     * Logs error, attempts cleanup, and throws
     */
    handleInitializationError(error: Error | string, context: string, cleanupFn?: () => void, metadata?: ErrorMetadata): never;
    /**
     * Wraps async operations with standardized error handling
     */
    wrapAsync<T, TArgs extends readonly unknown[]>(asyncFn: (...args: TArgs) => Promise<T>, context: string, options?: {
        readonly userMessage?: string;
        readonly critical?: boolean;
        readonly metadata?: ErrorMetadata;
        readonly onError?: (error: Error) => void;
    }): (...args: TArgs) => Promise<T>;
    /**
     * Creates a safe wrapper for synchronous operations
     */
    wrapSync<T, TArgs extends readonly unknown[]>(syncFn: (...args: TArgs) => T, context: string, options?: {
        readonly userMessage?: string;
        readonly critical?: boolean;
        readonly metadata?: ErrorMetadata;
        readonly defaultValue?: T;
    }): (...args: TArgs) => T;
    /**
     * Normalizes different error types to Error objects
     */
    private normalizeError;
    /**
     * Creates error context with metadata
     */
    private createErrorContext;
    /**
     * Logs error with appropriate level and formatting
     */
    private logError;
    /**
     * Gets the component name this error handler is associated with
     */
    getComponent(): string;
    /**
     * Checks if error handling is configured for user messages
     */
    hasUserMessageHandler(): boolean;
    /**
     * Checks if error handling is configured for event emission
     */
    hasEventEmitter(): boolean;
}
/**
 * Creates a new error handler with the specified configuration
 */
export declare function createErrorHandler(options: ErrorHandlerOptions): ErrorHandler;
/**
 * Global error handler factory for common use cases
 */
export declare class GlobalErrorHandler {
    private static readonly handlers;
    /**
     * Gets or creates an error handler for a component
     */
    static getHandler(component: string, options?: Partial<ErrorHandlerOptions>): ErrorHandler;
    /**
     * Sets up global error handlers for unhandled errors
     */
    static setupGlobalHandlers(): void;
    /**
     * Clears all cached error handlers
     */
    static clearHandlers(): void;
}
//# sourceMappingURL=ErrorHandler.d.ts.map