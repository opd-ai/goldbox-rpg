/**
 * Base component class providing common lifecycle and error handling patterns
 * Implements standardized initialization, update, and cleanup for game components
 */
import type { ComponentLifecycle, Service } from '../types/GameTypes';
import type { EventEmitterInterface } from '../types/UITypes';
import { type ErrorHandler } from '../utils/ErrorHandler';
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
export declare abstract class BaseComponent implements ComponentLifecycle {
    protected readonly componentName: string;
    protected readonly eventEmitter: EventEmitterInterface;
    protected readonly errorHandler: ErrorHandler;
    protected readonly componentLogger: typeof logger;
    private _initialized;
    private _destroyed;
    constructor(options: BaseComponentOptions);
    /**
     * Gets the component name
     */
    get name(): string;
    /**
     * Gets the initialization state
     */
    get initialized(): boolean;
    /**
     * Gets the destruction state
     */
    get destroyed(): boolean;
    /**
     * Public initialize method with error handling and state management
     */
    initialize(): Promise<void>;
    /**
     * Public cleanup method with error handling and state management
     */
    cleanup(): Promise<void>;
    /**
     * Update method for components that need regular updates
     * Override in subclasses that need update functionality
     */
    update(deltaTime: number): void;
    /**
     * Safe event emission that catches errors
     */
    protected emit<T = unknown>(event: string, data?: T): void;
    /**
     * Safe event listener registration
     */
    protected on<T = unknown>(event: string, callback: (data: T) => void): () => void;
    /**
     * Assert component is initialized before operations
     */
    protected assertInitialized(): void;
    /**
     * Wraps component methods with error handling
     */
    protected wrapMethod<T extends readonly unknown[]>(method: (...args: T) => void, methodName: string): (...args: T) => void;
    /**
     * Wraps async component methods with error handling
     */
    protected wrapAsyncMethod<T extends readonly unknown[], R>(method: (...args: T) => Promise<R>, methodName: string): (...args: T) => Promise<R>;
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
    protected onUpdate(_deltaTime: number): void;
    private safeInitialize;
    private createNoOpEventEmitter;
}
/**
 * Service base class that extends BaseComponent with service-specific features
 */
export declare abstract class BaseService extends BaseComponent implements Service {
    constructor(name: string, options?: Partial<BaseComponentOptions>);
    /**
     * Service-specific start method
     * Override to implement service startup logic
     */
    start(): Promise<void>;
    /**
     * Service-specific stop method
     * Override to implement service shutdown logic
     */
    stop(): Promise<void>;
    /**
     * Override to implement service-specific start logic
     */
    protected onStart(): Promise<void>;
    /**
     * Override to implement service-specific stop logic
     */
    protected onStop(): Promise<void>;
}
/**
 * Component manager for handling multiple components
 */
export declare class ComponentManager {
    private readonly components;
    private readonly logger;
    /**
     * Register a component
     */
    register(component: BaseComponent): void;
    /**
     * Get a component by name
     */
    get<T extends BaseComponent>(name: string): T | undefined;
    /**
     * Initialize all registered components
     */
    initializeAll(): Promise<void>;
    /**
     * Cleanup all registered components
     */
    cleanupAll(): Promise<void>;
    /**
     * Update all components
     */
    updateAll(deltaTime: number): void;
    /**
     * Get all registered component names
     */
    getComponentNames(): string[];
    /**
     * Get component count
     */
    getComponentCount(): number;
}
//# sourceMappingURL=BaseComponent.d.ts.map