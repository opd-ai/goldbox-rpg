/**
 * Type-safe EventEmitter implementation for GoldBox RPG Engine
 * Provides memory leak prevention and proper TypeScript support
 */
export class EventEmitter {
    constructor() {
        this.events = new Map();
    }
    /**
     * Register an event listener
     * @param event - Event name to listen for
     * @param callback - Function to call when event is emitted
     * @returns Unsubscriber function to remove the listener
     */
    on(event, callback) {
        if (!this.events.has(event)) {
            this.events.set(event, new Set());
        }
        const listeners = this.events.get(event);
        listeners.add(callback);
        // Return unsubscriber function
        return () => {
            listeners.delete(callback);
            if (listeners.size === 0) {
                this.events.delete(event);
            }
        };
    }
    /**
     * Emit an event to all registered listeners
     * @param event - Event name to emit
     * @param data - Data to pass to event listeners
     */
    emit(event, data) {
        const listeners = this.events.get(event);
        if (!listeners) {
            return;
        }
        // Create a copy of listeners to prevent issues if listeners are modified during iteration
        const listenersCopy = Array.from(listeners);
        for (const callback of listenersCopy) {
            try {
                callback(data);
            }
            catch (error) {
                console.error(`Error in event listener for '${event}':`, error);
                // Continue with other listeners even if one fails
            }
        }
    }
    /**
     * Remove a specific event listener
     * @param event - Event name
     * @param callback - Specific callback to remove
     * @returns True if listener was found and removed, false otherwise
     */
    off(event, callback) {
        const listeners = this.events.get(event);
        if (!listeners) {
            return false;
        }
        const removed = listeners.delete(callback);
        if (listeners.size === 0) {
            this.events.delete(event);
        }
        return removed;
    }
    /**
     * Remove all listeners for a specific event, or all events if no event specified
     * @param event - Optional specific event to clear, or undefined to clear all
     */
    removeAllListeners(event) {
        if (event !== undefined) {
            this.events.delete(event);
        }
        else {
            this.events.clear();
        }
    }
    /**
     * Clear all events and listeners (memory leak prevention)
     */
    clear() {
        this.events.clear();
    }
    /**
     * Get the number of listeners for a specific event
     * @param event - Event name to count listeners for
     * @returns Number of registered listeners
     */
    listenerCount(event) {
        const listeners = this.events.get(event);
        return listeners ? listeners.size : 0;
    }
    /**
     * Get an array of all event names that have listeners
     * @returns Array of event names
     */
    eventNames() {
        return Array.from(this.events.keys());
    }
    /**
     * Check if there are any listeners for a specific event
     * @param event - Event name to check
     * @returns True if there are listeners, false otherwise
     */
    hasListeners(event) {
        return this.listenerCount(event) > 0;
    }
    /**
     * Get debugging information about current event listeners
     * @returns Object with event listener statistics
     */
    getDebugInfo() {
        const events = {};
        let totalListeners = 0;
        for (const [event, listeners] of Array.from(this.events.entries())) {
            events[event] = listeners.size;
            totalListeners += listeners.size;
        }
        return {
            totalEvents: this.events.size,
            totalListeners,
            events
        };
    }
}
/**
 * Typed EventEmitter that enforces specific event types
 * @template TEventMap - Object type mapping event names to their data types
 */
export class TypedEventEmitter {
    constructor() {
        this.emitter = new EventEmitter();
    }
    /**
     * Register a typed event listener
     */
    on(event, callback) {
        return this.emitter.on(event, callback);
    }
    /**
     * Emit a typed event
     */
    emit(event, data) {
        this.emitter.emit(event, data);
    }
    /**
     * Remove a typed event listener
     */
    off(event, callback) {
        return this.emitter.off(event, callback);
    }
    /**
     * Remove all listeners for an event
     */
    removeAllListeners(event) {
        this.emitter.removeAllListeners(event);
    }
    /**
     * Clear all events
     */
    clear() {
        this.emitter.clear();
    }
    /**
     * Get listener count for an event
     */
    listenerCount(event) {
        return this.emitter.listenerCount(event);
    }
    /**
     * Get all event names
     */
    eventNames() {
        return this.emitter.eventNames();
    }
    /**
     * Get debug information
     */
    getDebugInfo() {
        return this.emitter.getDebugInfo();
    }
}
//# sourceMappingURL=EventEmitter.js.map