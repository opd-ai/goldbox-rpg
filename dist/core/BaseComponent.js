/**
 * Base component class providing common lifecycle and error handling patterns
 * Implements standardized initialization, update, and cleanup for game components
 */
import { EventEmitter } from './EventEmitter';
import { createErrorHandler } from '../utils/ErrorHandler';
import { logger } from '../utils/Logger';
/**
 * Base component class that provides:
 * - Standardized lifecycle management
 * - Error handling integration
 * - Event emission capabilities
 * - Initialization state tracking
 */
export class BaseComponent {
    constructor(options) {
        this._initialized = false;
        this._destroyed = false;
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
    get name() {
        return this.componentName;
    }
    /**
     * Gets the initialization state
     */
    get initialized() {
        return this._initialized;
    }
    /**
     * Gets the destruction state
     */
    get destroyed() {
        return this._destroyed;
    }
    /**
     * Public initialize method with error handling and state management
     */
    async initialize() {
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
        }
        catch (error) {
            this.errorHandler.handleInitializationError(error instanceof Error ? error : new Error(String(error)), 'initialize', () => this.onCleanup(), { componentName: this.componentName });
        }
    }
    /**
     * Public cleanup method with error handling and state management
     */
    async cleanup() {
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
        }
        catch (error) {
            this.errorHandler.handleRecoverableError(error instanceof Error ? error : new Error(String(error)), 'cleanup', undefined, { componentName: this.componentName });
        }
    }
    /**
     * Update method for components that need regular updates
     * Override in subclasses that need update functionality
     */
    update(deltaTime) {
        if (!this._initialized || this._destroyed) {
            return;
        }
        try {
            this.onUpdate(deltaTime);
        }
        catch (error) {
            this.errorHandler.handleRecoverableError(error instanceof Error ? error : new Error(String(error)), 'update', undefined, { componentName: this.componentName, deltaTime });
        }
    }
    /**
     * Safe event emission that catches errors
     */
    emit(event, data) {
        try {
            this.eventEmitter.emit(event, data);
        }
        catch (error) {
            this.errorHandler.handleRecoverableError(error instanceof Error ? error : new Error(String(error)), 'emit', undefined, { event, data });
        }
    }
    /**
     * Safe event listener registration
     */
    on(event, callback) {
        const wrappedCallback = this.errorHandler.wrapSync(callback, `eventListener:${event}`, {
            userMessage: 'An error occurred while handling an event',
            metadata: { event }
        });
        return this.eventEmitter.on(event, wrappedCallback);
    }
    /**
     * Assert component is initialized before operations
     */
    assertInitialized() {
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
    wrapMethod(method, methodName) {
        return this.errorHandler.wrapSync(method, methodName, {
            metadata: { componentName: this.componentName }
        });
    }
    /**
     * Wraps async component methods with error handling
     */
    wrapAsyncMethod(method, methodName) {
        return this.errorHandler.wrapAsync(method, methodName, {
            metadata: { componentName: this.componentName }
        });
    }
    /**
     * Override to implement component-specific update logic
     * Only called if component is initialized and not destroyed
     */
    onUpdate(_deltaTime) {
        // Default implementation does nothing
        // Subclasses can override if they need update functionality
    }
    // Private helper methods
    async safeInitialize() {
        try {
            await this.initialize();
        }
        catch (error) {
            this.componentLogger.error('Auto-initialization failed:', error);
        }
    }
    createNoOpEventEmitter() {
        return {
            on: () => () => { },
            emit: () => { },
            off: () => false,
            removeAllListeners: () => { },
            clear: () => { },
            listenerCount: () => 0,
            eventNames: () => []
        };
    }
}
/**
 * Service base class that extends BaseComponent with service-specific features
 */
export class BaseService extends BaseComponent {
    constructor(name, options = {}) {
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
    async start() {
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
    async stop() {
        await this.onStop();
        await this.cleanup();
        this.emit('stopped', { service: this.name });
    }
    /**
     * Override to implement service-specific start logic
     */
    async onStart() {
        // Default implementation does nothing
    }
    /**
     * Override to implement service-specific stop logic
     */
    async onStop() {
        // Default implementation does nothing
    }
}
/**
 * Component manager for handling multiple components
 */
export class ComponentManager {
    constructor() {
        this.components = new Map();
        this.logger = logger.createChildLogger('ComponentManager');
    }
    /**
     * Register a component
     */
    register(component) {
        if (this.components.has(component.name)) {
            throw new Error(`Component ${component.name} is already registered`);
        }
        this.components.set(component.name, component);
        this.logger.debug(`Registered component: ${component.name}`);
    }
    /**
     * Get a component by name
     */
    get(name) {
        return this.components.get(name);
    }
    /**
     * Initialize all registered components
     */
    async initializeAll() {
        this.logger.info('Initializing all components');
        const initPromises = Array.from(this.components.values()).map(component => component.initialize());
        await Promise.all(initPromises);
        this.logger.info('All components initialized');
    }
    /**
     * Cleanup all registered components
     */
    async cleanupAll() {
        this.logger.info('Cleaning up all components');
        const cleanupPromises = Array.from(this.components.values()).map(component => component.cleanup());
        await Promise.all(cleanupPromises);
        this.components.clear();
        this.logger.info('All components cleaned up');
    }
    /**
     * Update all components
     */
    updateAll(deltaTime) {
        for (const component of Array.from(this.components.values())) {
            component.update(deltaTime);
        }
    }
    /**
     * Get all registered component names
     */
    getComponentNames() {
        return Array.from(this.components.keys());
    }
    /**
     * Get component count
     */
    getComponentCount() {
        return this.components.size;
    }
}
//# sourceMappingURL=BaseComponent.js.map