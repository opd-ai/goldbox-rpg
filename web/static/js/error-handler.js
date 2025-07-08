/**
 * Standardized error handling utility for consistent error management
 * across the GoldBox RPG client application.
 * 
 * Provides unified methods for:
 * - Error logging with context
 * - Error event emission
 * - User-friendly error messages
 * - Error classification and handling strategies
 */
class ErrorHandler {
  /**
   * Creates an error handler instance
   * @param {string} component - The component name (e.g., "GameState", "CombatManager")
   * @param {EventEmitter} [eventEmitter] - Optional event emitter for error events
   * @param {Function} [userMessageCallback] - Optional callback for user messages
   */
  constructor(component, eventEmitter = null, userMessageCallback = null) {
    this.component = component;
    this.eventEmitter = eventEmitter;
    this.userMessageCallback = userMessageCallback;
  }

  /**
   * Handles recoverable errors that should not stop execution
   * Logs error, emits event, and optionally shows user message
   * 
   * @param {Error|string} error - The error to handle
   * @param {string} context - Context where error occurred
   * @param {string} [userMessage] - Optional user-friendly message
   * @param {Object} [metadata] - Additional context data
   */
  handleRecoverableError(error, context, userMessage = null, metadata = {}) {
    const errorObj = error instanceof Error ? error : new Error(error);
    
    // Always log with context
    console.error(`${this.component}.${context}:`, errorObj, metadata);
    
    // Emit error event if event emitter available
    if (this.eventEmitter && typeof this.eventEmitter.emit === 'function') {
      this.eventEmitter.emit('error', {
        error: errorObj,
        context,
        component: this.component,
        metadata
      });
    }
    
    // Show user message if callback available and message provided
    if (this.userMessageCallback && userMessage) {
      this.userMessageCallback(userMessage, 'error');
    }
    
    return { success: false, error: errorObj };
  }

  /**
   * Handles critical errors that should stop execution
   * Logs error and throws it to stop execution flow
   * 
   * @param {Error|string} error - The error to handle
   * @param {string} context - Context where error occurred
   * @param {Object} [metadata] - Additional context data
   * @throws {Error} Always throws the error
   */
  handleCriticalError(error, context, metadata = {}) {
    const errorObj = error instanceof Error ? error : new Error(error);
    
    console.error(`${this.component}.${context}: CRITICAL ERROR`, errorObj, metadata);
    
    // Always throw critical errors to stop execution
    throw errorObj;
  }

  /**
   * Handles initialization errors with cleanup
   * Logs error, attempts cleanup, and throws
   * 
   * @param {Error|string} error - The error to handle
   * @param {string} context - Context where error occurred
   * @param {Function} [cleanupFn] - Optional cleanup function
   * @param {Object} [metadata] - Additional context data
   * @throws {Error} Always throws the error after cleanup
   */
  handleInitializationError(error, context, cleanupFn = null, metadata = {}) {
    const errorObj = error instanceof Error ? error : new Error(error);
    
    console.error(`${this.component}.${context}: INITIALIZATION ERROR`, errorObj, metadata);
    
    // Attempt cleanup if provided
    if (cleanupFn && typeof cleanupFn === 'function') {
      try {
        cleanupFn();
        console.info(`${this.component}.${context}: cleanup completed`);
      } catch (cleanupError) {
        console.error(`${this.component}.${context}: cleanup failed`, cleanupError);
      }
    }
    
    throw errorObj;
  }

  /**
   * Wraps async operations with standardized error handling
   * 
   * @param {Function} asyncFn - The async function to wrap
   * @param {string} context - Context for error reporting
   * @param {Object} [options] - Options for error handling
   * @param {string} [options.userMessage] - User-friendly error message
   * @param {boolean} [options.critical=false] - Whether errors should be critical
   * @returns {Function} Wrapped async function
   */
  wrapAsync(asyncFn, context, options = {}) {
    return async (...args) => {
      try {
        return await asyncFn(...args);
      } catch (error) {
        if (options.critical) {
          return this.handleCriticalError(error, context);
        } else {
          return this.handleRecoverableError(error, context, options.userMessage);
        }
      }
    };
  }
}
