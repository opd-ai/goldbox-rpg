/**
 * Type-safe EventEmitter implementation for GoldBox RPG Engine
 * Provides memory leak prevention and proper TypeScript support
 */
import type { EventCallback, EventEmitterInterface, EventUnsubscriber } from '../types/UITypes';
export declare class EventEmitter implements EventEmitterInterface {
    private readonly events;
    /**
     * Register an event listener
     * @param event - Event name to listen for
     * @param callback - Function to call when event is emitted
     * @returns Unsubscriber function to remove the listener
     */
    on<T = unknown>(event: string, callback: EventCallback<T>): EventUnsubscriber;
    /**
     * Emit an event to all registered listeners
     * @param event - Event name to emit
     * @param data - Data to pass to event listeners
     */
    emit<T = unknown>(event: string, data?: T): void;
    /**
     * Remove a specific event listener
     * @param event - Event name
     * @param callback - Specific callback to remove
     * @returns True if listener was found and removed, false otherwise
     */
    off(event: string, callback: EventCallback): boolean;
    /**
     * Remove all listeners for a specific event, or all events if no event specified
     * @param event - Optional specific event to clear, or undefined to clear all
     */
    removeAllListeners(event?: string): void;
    /**
     * Clear all events and listeners (memory leak prevention)
     */
    clear(): void;
    /**
     * Get the number of listeners for a specific event
     * @param event - Event name to count listeners for
     * @returns Number of registered listeners
     */
    listenerCount(event: string): number;
    /**
     * Get an array of all event names that have listeners
     * @returns Array of event names
     */
    eventNames(): string[];
    /**
     * Check if there are any listeners for a specific event
     * @param event - Event name to check
     * @returns True if there are listeners, false otherwise
     */
    hasListeners(event: string): boolean;
    /**
     * Get debugging information about current event listeners
     * @returns Object with event listener statistics
     */
    getDebugInfo(): Record<string, unknown>;
}
/**
 * Typed EventEmitter that enforces specific event types
 * @template TEventMap - Object type mapping event names to their data types
 */
export declare class TypedEventEmitter<TEventMap extends Record<string, unknown>> {
    private readonly emitter;
    /**
     * Register a typed event listener
     */
    on<K extends keyof TEventMap>(event: K, callback: (data: TEventMap[K]) => void): EventUnsubscriber;
    /**
     * Emit a typed event
     */
    emit<K extends keyof TEventMap>(event: K, data: TEventMap[K]): void;
    /**
     * Remove a typed event listener
     */
    off<K extends keyof TEventMap>(event: K, callback: (data: TEventMap[K]) => void): boolean;
    /**
     * Remove all listeners for an event
     */
    removeAllListeners<K extends keyof TEventMap>(event?: K): void;
    /**
     * Clear all events
     */
    clear(): void;
    /**
     * Get listener count for an event
     */
    listenerCount<K extends keyof TEventMap>(event: K): number;
    /**
     * Get all event names
     */
    eventNames(): Array<keyof TEventMap>;
    /**
     * Get debug information
     */
    getDebugInfo(): Record<string, unknown>;
}
//# sourceMappingURL=EventEmitter.d.ts.map