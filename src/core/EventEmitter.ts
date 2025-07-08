/**
 * Type-safe EventEmitter implementation for GoldBox RPG Engine
 * Provides memory leak prevention and proper TypeScript support
 */

import type { EventCallback, EventEmitterInterface, EventUnsubscriber } from '../types/UITypes';

export class EventEmitter implements EventEmitterInterface {
  private readonly events = new Map<string, Set<EventCallback>>();

  /**
   * Register an event listener
   * @param event - Event name to listen for
   * @param callback - Function to call when event is emitted
   * @returns Unsubscriber function to remove the listener
   */
  on<T = unknown>(event: string, callback: EventCallback<T>): EventUnsubscriber {
    if (!this.events.has(event)) {
      this.events.set(event, new Set());
    }
    
    const listeners = this.events.get(event)!;
    listeners.add(callback as EventCallback);

    // Return unsubscriber function
    return () => {
      listeners.delete(callback as EventCallback);
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
  emit<T = unknown>(event: string, data?: T): void {
    const listeners = this.events.get(event);
    if (!listeners) {
      return;
    }

    // Create a copy of listeners to prevent issues if listeners are modified during iteration
    const listenersCopy = Array.from(listeners);
    
    for (const callback of listenersCopy) {
      try {
        callback(data);
      } catch (error) {
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
  off(event: string, callback: EventCallback): boolean {
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
  removeAllListeners(event?: string): void {
    if (event !== undefined) {
      this.events.delete(event);
    } else {
      this.events.clear();
    }
  }

  /**
   * Clear all events and listeners (memory leak prevention)
   */
  clear(): void {
    this.events.clear();
  }

  /**
   * Get the number of listeners for a specific event
   * @param event - Event name to count listeners for
   * @returns Number of registered listeners
   */
  listenerCount(event: string): number {
    const listeners = this.events.get(event);
    return listeners ? listeners.size : 0;
  }

  /**
   * Get an array of all event names that have listeners
   * @returns Array of event names
   */
  eventNames(): string[] {
    return Array.from(this.events.keys());
  }

  /**
   * Check if there are any listeners for a specific event
   * @param event - Event name to check
   * @returns True if there are listeners, false otherwise
   */
  hasListeners(event: string): boolean {
    return this.listenerCount(event) > 0;
  }

  /**
   * Get debugging information about current event listeners
   * @returns Object with event listener statistics
   */
  getDebugInfo(): Record<string, unknown> {
    const events: Record<string, number> = {};
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
export class TypedEventEmitter<TEventMap extends Record<string, unknown>> {
  private readonly emitter = new EventEmitter();

  /**
   * Register a typed event listener
   */
  on<K extends keyof TEventMap>(
    event: K,
    callback: (data: TEventMap[K]) => void
  ): EventUnsubscriber {
    return this.emitter.on(event as string, callback as EventCallback);
  }

  /**
   * Emit a typed event
   */
  emit<K extends keyof TEventMap>(event: K, data: TEventMap[K]): void {
    this.emitter.emit(event as string, data);
  }

  /**
   * Remove a typed event listener
   */
  off<K extends keyof TEventMap>(
    event: K,
    callback: (data: TEventMap[K]) => void
  ): boolean {
    return this.emitter.off(event as string, callback as EventCallback);
  }

  /**
   * Remove all listeners for an event
   */
  removeAllListeners<K extends keyof TEventMap>(event?: K): void {
    this.emitter.removeAllListeners(event as string);
  }

  /**
   * Clear all events
   */
  clear(): void {
    this.emitter.clear();
  }

  /**
   * Get listener count for an event
   */
  listenerCount<K extends keyof TEventMap>(event: K): number {
    return this.emitter.listenerCount(event as string);
  }

  /**
   * Get all event names
   */
  eventNames(): Array<keyof TEventMap> {
    return this.emitter.eventNames() as Array<keyof TEventMap>;
  }

  /**
   * Get debug information
   */
  getDebugInfo(): Record<string, unknown> {
    return this.emitter.getDebugInfo();
  }
}
