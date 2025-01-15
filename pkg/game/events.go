package game

import (
	"sync"
)

// EventType represents different types of game events
type EventType int

const (
	EventLevelUp EventType = iota
	EventDamage
	EventDeath
	EventItemPickup
	EventItemDrop
	EventSpellCast
	EventQuestUpdate
)

// GameEvent represents an event in the game
// Contains all metadata and payload for event processing
type GameEvent struct {
	Type      EventType              `yaml:"event_type"`      // Type of the event
	SourceID  string                 `yaml:"source_id"`       // ID of the event originator
	TargetID  string                 `yaml:"target_id"`       // ID of the event target
	Data      map[string]interface{} `yaml:"event_data"`      // Additional event data
	Timestamp int64                  `yaml:"event_timestamp"` // When the event occurred
}

// EventHandler represents a function that handles game events
type EventHandler func(event GameEvent)

// EventSystem manages game event subscriptions and dispatching
// Provides thread-safe event handling infrastructure
type EventSystem struct {
	mu       sync.RWMutex                 `yaml:"mutex,omitempty"`          // Mutex for thread safety
	handlers map[EventType][]EventHandler `yaml:"event_handlers,omitempty"` // Map of event handlers
}

// EventSystemConfig represents serializable configuration for the event system
type EventSystemConfig struct {
	RegisteredTypes []EventType       `yaml:"registered_event_types"` // List of registered event types
	HandlerCount    map[EventType]int `yaml:"handler_counts"`         // Number of handlers per type
	AsyncHandling   bool              `yaml:"async_handling"`         // Whether events are handled asynchronously
}

var defaultEventSystem = NewEventSystem()

// NewEventSystem creates a new event system
func NewEventSystem() *EventSystem {
	return &EventSystem{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe registers a handler for a specific event type
func (es *EventSystem) Subscribe(eventType EventType, handler EventHandler) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.handlers[eventType] = append(es.handlers[eventType], handler)
}

// Emit sends an event to all registered handlers
func (es *EventSystem) Emit(event GameEvent) {
	es.mu.RLock()
	handlers := es.handlers[event.Type]
	es.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event) // Async event handling
	}
}

// emitLevelUpEvent is a helper function to emit level up events
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

// getCurrentGameTick returns the current game tick (implement based on your game loop)
func getCurrentGameTick() int64 {
	// Implementation depends on your game loop timing system
	// This is a placeholder
	return 0
}

// Example event handlers
func init() {
	// Register default event handlers
	defaultEventSystem.Subscribe(EventLevelUp, func(event GameEvent) {
		// Log level up
		oldLevel := event.Data["oldLevel"].(int)
		newLevel := event.Data["newLevel"].(int)
		logger.Printf("Player %s leveled up from %d to %d",
			event.SourceID, oldLevel, newLevel)
	})
}
