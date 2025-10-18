package game

import (
	"sync"
)

// EventType represents different types of game events
// EventType represents the type of an event in the game.
// It is implemented as an integer enum to allow for efficient comparison and switching.
// The specific event type values should be defined as constants using this type.
//
// Related types:
//   - Event interface (if exists)
//   - Any concrete event types that use this enum
type EventType int

// EventLevelUp represents a character gaining a level.
// This event is triggered when a character accumulates enough experience points
// to advance to the next level. The event carries information about:
// - The character that leveled up
// - The new level achieved
// - Any stat increases or new abilities gained
//
// Related events:
// EventType constants are defined in constants.go
// - EventDamage: May contribute to experience gain
// - EventQuestUpdate: Quests may require reaching certain levels

// GameEvent represents an occurrence or action within the game system that needs to be tracked or handled.
// It contains information about what happened, who/what was involved, and when it occurred.
//
// Fields:
//   - Type: The category/classification of the event (EventType)
//   - SourceID: Unique identifier for the entity that triggered/caused the event
//   - TargetID: Unique identifier for the entity that the event affects/targets
//   - Data: Additional contextual information about the event as key-value pairs
//   - Timestamp: Unix timestamp (in seconds) when the event occurred
//
// The GameEvent struct is used throughout the event system to standardize how
// game occurrences are represented and processed. Events can represent things like
// combat actions, item usage, movement, etc.
//
// Related types:
//   - EventType: Enumeration of possible event categories
type GameEvent struct {
	Type      EventType              `yaml:"event_type"`      // Type of the event
	SourceID  string                 `yaml:"source_id"`       // ID of the event originator
	TargetID  string                 `yaml:"target_id"`       // ID of the event target
	Data      map[string]interface{} `yaml:"event_data"`      // Additional event data
	Timestamp int64                  `yaml:"event_timestamp"` // When the event occurred
}

// EventHandler is a function type that handles game events in the game system.
// It takes a GameEvent parameter and processes it according to the specific event handling logic.
//
// Parameters:
//   - event GameEvent: The game event to be handled
//
// Note: EventHandler functions are typically used as callbacks registered to handle
// specific types of game events in an event-driven architecture.
//
// Related types:
//   - GameEvent (defined elsewhere in the codebase)
type EventHandler func(event GameEvent)

// EventSystem manages event handling and dispatching in the game.
// It provides a thread-safe way to register handlers for different event types
// and dispatch events to all registered handlers.
//
// Fields:
//   - mu: sync.RWMutex for ensuring thread-safe access to handlers
//   - handlers: Map storing event handlers organized by EventType
//
// Thread Safety:
// All methods on EventSystem are thread-safe and can be called concurrently
// from multiple goroutines.
//
// Related Types:
//   - EventType: Type definition for different kinds of game events
//   - EventHandler: Interface for handling dispatched events
type EventSystem struct {
	mu       sync.RWMutex                 `yaml:"mutex,omitempty"`          // Mutex for thread safety
	handlers map[EventType][]EventHandler `yaml:"event_handlers,omitempty"` // Map of event handlers
}

// EventSystemConfig defines the configuration settings for the event handling system.
// It manages event type registration, handler tracking, and processing behavior.
//
// Fields:
//   - RegisteredTypes: Slice of EventType that are registered in the system.
//   - HandlerCount: Map tracking number of handlers registered for each EventType.
//     A count of 0 indicates no handlers are registered for that type.
//   - AsyncHandling: Boolean flag determining if events are processed asynchronously.
//     When true, events are handled in separate goroutines.
//     When false, events are handled synchronously in the calling goroutine.
//
// The config should be initialized before registering any event handlers.
// AsyncHandling should be used with caution as it may affect event ordering.
//
// Related:
//   - EventType: Type definition for supported event types
//   - EventHandler: Interface for event handler implementations
type EventSystemConfig struct {
	RegisteredTypes []EventType       `yaml:"registered_event_types"` // List of registered event types
	HandlerCount    map[EventType]int `yaml:"handler_counts"`         // Number of handlers per type
	AsyncHandling   bool              `yaml:"async_handling"`         // Whether events are handled asynchronously
}

// defaultEventSystem is the global default event system instance used for managing game events.
// It is initialized using NewEventSystem() and serves as a centralized event bus for the game.
// This singleton instance allows components throughout the game to subscribe to and publish events
// without having to pass around an event system reference.
//
// Related types:
// - EventSystem: The underlying event management system type
// - Event: Base interface for all game events
//
// Usage:
// To publish events: defaultEventSystem.Publish(event)
// To subscribe: defaultEventSystem.Subscribe(eventType, handler)
var defaultEventSystem = NewEventSystem()

// NewEventSystem creates and initializes a new event system.
// It initializes an empty map of event handlers that can be registered
// to handle different event types.
//
// Returns:
//   - *EventSystem: A pointer to the newly created event system with an initialized
//     empty handlers map.
//
// Related types:
// - EventType: The type used to identify different kinds of events
// - EventHandler: Function type for handling specific events
func NewEventSystem() *EventSystem {
	return &EventSystem{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe registers a new event handler for a specific event type.
// The handler will be called when events of the specified type are published.
//
// Parameters:
//   - eventType: The type of event to subscribe to
//   - handler: The event handler function to be called when events occur
//
// Thread safety: This method is thread-safe as it uses mutex locking.
//
// Related:
//   - EventType
//   - EventHandler
//   - EventSystem.Publish
func (es *EventSystem) Subscribe(eventType EventType, handler EventHandler) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.handlers[eventType] = append(es.handlers[eventType], handler)
}

// Emit asynchronously distributes a game event to all registered handlers for that event type.
// It safely accesses the handlers map using a read lock to prevent concurrent map access issues.
//
// Parameters:
//   - event GameEvent: The game event to be processed. Must contain a valid Type field that
//     matches registered handler types.
//
// Thread-safety:
//   - Uses RWMutex to safely access handlers map
//   - Handlers are executed concurrently in separate goroutines
//
// Related types:
//   - GameEvent interface
//   - EventHandler func type
//   - EventType enum
func (es *EventSystem) Emit(event GameEvent) {
	es.mu.RLock()
	handlers := es.handlers[event.Type]
	es.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event) // Async event handling
	}
}

// emitLevelUpEvent sends a level up event to the default event system when a player levels up.
// It creates a GameEvent with the level up information and emits it.
//
// Parameters:
//   - playerID: string - Unique identifier for the player who leveled up
//   - oldLevel: int - The player's level before leveling up
//   - newLevel: int - The player's new level after leveling up
//
// Related:
//   - GameEvent struct
//   - defaultEventSystem
//   - EventLevelUp const
//   - getCurrentGameTick()
func emitLevelUpEvent(playerID string, oldLevel, newLevel int) {
	event := GameEvent{
		Type:     EventLevelUp,
		SourceID: playerID,
		Data: map[string]interface{}{
			"oldLevel": oldLevel,
			"newLevel": newLevel,
		},
		Timestamp: getCurrentGameTick(),
	}

	defaultEventSystem.Emit(event)
}

// getCurrentGameTick returns the current game tick count as an int64.
// This represents the number of game update cycles that have occurred.
//
// The actual implementation depends on the game loop timing system.
// Currently returns a placeholder value of 0.
//
// Returns:
//   - int64: The current game tick count
//
// Related:
//   - Game loop implementation (TBD)
//   - Time management system (TBD)
func getCurrentGameTick() int64 {
	// Implementation depends on your game loop timing system
	// This is a placeholder
	return 0
}

// init initializes the default event system by registering event handlers.
// It sets up a handler for the EventLevelUp event that logs when a player levels up.
//
// The handler processes the following event data:
// - "oldLevel" (int): The player's previous level
// - "newLevel" (int): The player's new level
// - SourceID (string): The player's ID/name
//
// This function is called automatically when the package is imported.
//
// Related:
// - GameEvent type
// - defaultEventSystem EventSystem
// - EventLevelUp constant
func init() {
	// Register default event handlers
	defaultEventSystem.Subscribe(EventLevelUp, func(event GameEvent) {
		// Log level up
		oldLevel := event.Data["oldLevel"].(int)
		newLevel := event.Data["newLevel"].(int)
		getLogger().Printf("Player %s leveled up from %d to %d",
			event.SourceID, oldLevel, newLevel)
	})
}
