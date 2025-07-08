/**
 * Base component class providing common lifecycle and error handling patterns
 * Implements standardized initialization, update, and cleanup for game components
 */

import type { ComponentLifecycle, Service } from '../types/GameTypes';
import type { EventEmitterInterface } from '../types/UITypes';
import { EventEmitter } from './EventEmitter';
import { createErrorHandler, type ErrorHandler } from '../utils/ErrorHandler';
import { logger } from '../utils/Logger';

export interface BaseComponentOptions {
  readonly name: string;
  readonly enableEventEmission?: boolean;
  readonly enableErrorHandling?: boolean;
  readonly autoInitialize?: boolean;
}

/**
 * Base component class that provides:
 * - Standardized lifecycle management
 * - Error handling integration
 * - Event emission capabilities
 * - Initialization state tracking
 */
export abstract class BaseComponent implements ComponentLifecycle {
  protected readonly componentName: string;
  protected readonly eventEmitter: EventEmitterInterface;
  protected readonly errorHandler: ErrorHandler;
  protected readonly componentLogger: typeof logger;
  
  private _initialized = false;
  private _destroyed = false;

  constructor(options: BaseComponentOptions) {
    this.componentName = options.name;
    this.componentLogger = logger.createChildLogger(this.componentName);
    
    // Initialize event emitter if enabled
    this.eventEmitter = options.enableEventEmission !== false 
      ? new EventEmitter() 
      : this.createNoOpEventEmitter();

    // Initialize error handler
    this.errorHandler = createErrorHandler({
      component: this.componentName,
      eventEmitter: this.eventEmitter,
      enableStackTrace: true,
      enableMetadataLogging: true
    });

    this.componentLogger.debug(`${this.componentName} component created`);

    // Auto-initialize if requested
    if (options.autoInitialize === true) {
      this.safeInitialize();
    }
  }

  /**
   * Gets the component name
   */
  get name(): string {
    return this.componentName;
  }

  /**
   * Gets the initialization state
   */
  get initialized(): boolean {
    return this._initialized;
  }

  /**
   * Gets the destruction state
   */
  get destroyed(): boolean {
    return this._destroyed;
  }

  /**
   * Public initialize method with error handling and state management
   */
  async initialize(): Promise<void> {
    if (this._initialized) {
      this.componentLogger.warn('Component already initialized');
      return;
    }

    if (this._destroyed) {
      throw new Error('Cannot initialize destroyed component');
    }

    try {
      this.componentLogger.info('Initializing component');
      await this.onInitialize();
      this._initialized = true;
      this.eventEmitter.emit('initialized', { component: this.componentName });
      this.componentLogger.info('Component initialized successfully');
    } catch (error) {
      this.errorHandler.handleInitializationError(
        error instanceof Error ? error : new Error(String(error)),
        'initialize',
        () => this.onCleanup(),
        { componentName: this.componentName }
      );
    }
  }

  /**
   * Public cleanup method with error handling and state management
   */
  async cleanup(): Promise<void> {
    if (this._destroyed) {
      this.componentLogger.warn('Component already destroyed');
      return;
    }

    try {
      this.componentLogger.info('Cleaning up component');
      await this.onCleanup();
      this._destroyed = true;
      this._initialized = false;
      this.eventEmitter.emit('destroyed', { component: this.componentName });
      this.eventEmitter.clear(); // Prevent memory leaks
      this.componentLogger.info('Component cleaned up successfully');
    } catch (error) {
      this.errorHandler.handleRecoverableError(
        error instanceof Error ? error : new Error(String(error)),
        'cleanup',
        undefined,
        { componentName: this.componentName }
      );
    }
  }

  /**
   * Update method for components that need regular updates
   * Override in subclasses that need update functionality
   */
  update(deltaTime: number): void {
    if (!this._initialized || this._destroyed) {
      return;
    }

    try {
      this.onUpdate(deltaTime);
    } catch (error) {
      this.errorHandler.handleRecoverableError(
        error instanceof Error ? error : new Error(String(error)),
        'update',
        undefined,
        { componentName: this.componentName, deltaTime }
      );
    }
  }

  /**
   * Safe event emission that catches errors
   */
  protected emit<T = unknown>(event: string, data?: T): void {
    try {
      this.eventEmitter.emit(event, data);
    } catch (error) {
      this.errorHandler.handleRecoverableError(
        error instanceof Error ? error : new Error(String(error)),
        'emit',
        undefined,
        { event, data }
      );
    }
  }

  /**
   * Safe event listener registration
   */
  protected on<T = unknown>(event: string, callback: (data: T) => void): () => void {
    const wrappedCallback = this.errorHandler.wrapSync(
      callback,
      `eventListener:${event}`,
      { 
        userMessage: 'An error occurred while handling an event',
        metadata: { event }
      }
    );

    return this.eventEmitter.on(event, wrappedCallback);
  }

  /**
   * Assert component is initialized before operations
   */
  protected assertInitialized(): void {
    if (!this._initialized) {
      throw new Error(`Component ${this.componentName} is not initialized`);
    }
    if (this._destroyed) {
      throw new Error(`Component ${this.componentName} has been destroyed`);
    }
  }

  /**
   * Wraps component methods with error handling
   */
  protected wrapMethod<T extends readonly unknown[]>(
    method: (...args: T) => void,
    methodName: string
  ): (...args: T) => void {
    return this.errorHandler.wrapSync(
      method,
      methodName,
      {
        metadata: { componentName: this.componentName }
      }
    );
  }

  /**
   * Wraps async component methods with error handling
   */
  protected wrapAsyncMethod<T extends readonly unknown[], R>(
    method: (...args: T) => Promise<R>,
    methodName: string
  ): (...args: T) => Promise<R> {
    return this.errorHandler.wrapAsync(
      method,
      methodName,
      {
        metadata: { componentName: this.componentName }
      }
    );
  }

  // Abstract methods that subclasses must implement

  /**
   * Override to implement component-specific initialization logic
   */
  protected abstract onInitialize(): Promise<void> | void;

  /**
   * Override to implement component-specific cleanup logic
   */
  protected abstract onCleanup(): Promise<void> | void;

  /**
   * Override to implement component-specific update logic
   * Only called if component is initialized and not destroyed
   */
  protected onUpdate(_deltaTime: number): void {
    // Default implementation does nothing
    // Subclasses can override if they need update functionality
  }

  // Private helper methods

  private async safeInitialize(): Promise<void> {
    try {
      await this.initialize();
    } catch (error) {
      this.componentLogger.error('Auto-initialization failed:', error);
    }
  }

  private createNoOpEventEmitter(): EventEmitterInterface {
    return {
      on: () => () => {},
      emit: () => {},
      off: () => false,
      removeAllListeners: () => {},
      clear: () => {},
      listenerCount: () => 0,
      eventNames: () => []
    };
  }
}

/**
 * Service base class that extends BaseComponent with service-specific features
 */
export abstract class BaseService extends BaseComponent implements Service {
  constructor(name: string, options: Partial<BaseComponentOptions> = {}) {
    super({
      name,
      enableEventEmission: true,
      enableErrorHandling: true,
      ...options
    });
  }

  /**
   * Service-specific start method
   * Override to implement service startup logic
   */
  async start(): Promise<void> {
    if (!this.initialized) {
      await this.initialize();
    }
    await this.onStart();
    this.emit('started', { service: this.name });
  }

  /**
   * Service-specific stop method
   * Override to implement service shutdown logic
   */
  async stop(): Promise<void> {
    await this.onStop();
    await this.cleanup();
    this.emit('stopped', { service: this.name });
  }

  /**
   * Override to implement service-specific start logic
   */
  protected async onStart(): Promise<void> {
    // Default implementation does nothing
  }

  /**
   * Override to implement service-specific stop logic
   */
  protected async onStop(): Promise<void> {
    // Default implementation does nothing
  }
}

/**
 * Component manager for handling multiple components
 */
export class ComponentManager {
  private readonly components = new Map<string, BaseComponent>();
  private readonly logger = logger.createChildLogger('ComponentManager');

  /**
   * Register a component
   */
  register(component: BaseComponent): void {
    if (this.components.has(component.name)) {
      throw new Error(`Component ${component.name} is already registered`);
    }

    this.components.set(component.name, component);
    this.logger.debug(`Registered component: ${component.name}`);
  }

  /**
   * Get a component by name
   */
  get<T extends BaseComponent>(name: string): T | undefined {
    return this.components.get(name) as T | undefined;
  }

  /**
   * Initialize all registered components
   */
  async initializeAll(): Promise<void> {
    this.logger.info('Initializing all components');
    
    const initPromises = Array.from(this.components.values()).map(
      component => component.initialize()
    );

    await Promise.all(initPromises);
    this.logger.info('All components initialized');
  }

  /**
   * Cleanup all registered components
   */
  async cleanupAll(): Promise<void> {
    this.logger.info('Cleaning up all components');
    
    const cleanupPromises = Array.from(this.components.values()).map(
      component => component.cleanup()
    );

    await Promise.all(cleanupPromises);
    this.components.clear();
    this.logger.info('All components cleaned up');
  }

  /**
   * Update all components
   */
  updateAll(deltaTime: number): void {
    for (const component of Array.from(this.components.values())) {
      component.update(deltaTime);
    }
  }

  /**
   * Get all registered component names
   */
  getComponentNames(): string[] {
    return Array.from(this.components.keys());
  }

  /**
   * Get component count
   */
  getComponentCount(): number {
    return this.components.size;
  }
}
