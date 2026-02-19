package game

import (
	"sync"
	"testing"
	"time"
)

// TestNewEventSystem tests the NewEventSystem constructor function
func TestNewEventSystem_ReturnsInitializedEventSystem(t *testing.T) {
	eventSystem := NewEventSystem()

	if eventSystem == nil {
		t.Fatal("NewEventSystem() returned nil")
	}

	if eventSystem.handlers == nil {
		t.Error("NewEventSystem() returned EventSystem with nil handlers map")
	}

	if len(eventSystem.handlers) != 0 {
		t.Errorf("NewEventSystem() returned EventSystem with non-empty handlers map, got %d handlers", len(eventSystem.handlers))
	}
}

// TestEventSystem_Subscribe tests the Subscribe method
func TestEventSystem_Subscribe(t *testing.T) {
	tests := []struct {
		name        string
		eventType   EventType
		handlerFunc EventHandler
	}{
		{
			name:      "Subscribe_LevelUpEvent_SuccessfulRegistration",
			eventType: EventLevelUp,
			handlerFunc: func(event GameEvent) {
				// Test handler function
			},
		},
		{
			name:      "Subscribe_DamageEvent_SuccessfulRegistration",
			eventType: EventDamage,
			handlerFunc: func(event GameEvent) {
				// Test handler function
			},
		},
		{
			name:      "Subscribe_MovementEvent_SuccessfulRegistration",
			eventType: EventMovement,
			handlerFunc: func(event GameEvent) {
				// Test handler function
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventSystem := NewEventSystem()

			// Subscribe handler
			eventSystem.Subscribe(tt.eventType, tt.handlerFunc)

			// Verify handler was registered
			eventSystem.mu.RLock()
			handlers := eventSystem.handlers[tt.eventType]
			eventSystem.mu.RUnlock()

			if len(handlers) != 1 {
				t.Errorf("Subscribe() failed to register handler, expected 1 handler, got %d", len(handlers))
			}
		})
	}
}

// TestEventSystem_Subscribe_MultipleHandlers tests subscribing multiple handlers to the same event type
func TestEventSystem_Subscribe_MultipleHandlers(t *testing.T) {
	eventSystem := NewEventSystem()
	eventType := EventLevelUp

	handler1 := func(event GameEvent) {}
	handler2 := func(event GameEvent) {}
	handler3 := func(event GameEvent) {}

	// Subscribe multiple handlers
	eventSystem.Subscribe(eventType, handler1)
	eventSystem.Subscribe(eventType, handler2)
	eventSystem.Subscribe(eventType, handler3)

	// Verify all handlers were registered
	eventSystem.mu.RLock()
	handlers := eventSystem.handlers[eventType]
	eventSystem.mu.RUnlock()

	if len(handlers) != 3 {
		t.Errorf("Subscribe() failed to register multiple handlers, expected 3 handlers, got %d", len(handlers))
	}
}

// TestEventSystem_Subscribe_DifferentEventTypes tests subscribing to different event types
func TestEventSystem_Subscribe_DifferentEventTypes(t *testing.T) {
	eventSystem := NewEventSystem()

	handler := func(event GameEvent) {}

	// Subscribe to different event types
	eventSystem.Subscribe(EventLevelUp, handler)
	eventSystem.Subscribe(EventDamage, handler)
	eventSystem.Subscribe(EventDeath, handler)

	// Verify handlers were registered for each event type
	eventSystem.mu.RLock()
	defer eventSystem.mu.RUnlock()

	if len(eventSystem.handlers[EventLevelUp]) != 1 {
		t.Errorf("Subscribe() failed for EventLevelUp, expected 1 handler, got %d", len(eventSystem.handlers[EventLevelUp]))
	}

	if len(eventSystem.handlers[EventDamage]) != 1 {
		t.Errorf("Subscribe() failed for EventDamage, expected 1 handler, got %d", len(eventSystem.handlers[EventDamage]))
	}

	if len(eventSystem.handlers[EventDeath]) != 1 {
		t.Errorf("Subscribe() failed for EventDeath, expected 1 handler, got %d", len(eventSystem.handlers[EventDeath]))
	}
}

// TestEventSystem_Emit tests the Emit method
func TestEventSystem_Emit(t *testing.T) {
	tests := []struct {
		name      string
		event     GameEvent
		setupFunc func(*EventSystem) chan bool
	}{
		{
			name: "Emit_LevelUpEvent_HandlerCalled",
			event: GameEvent{
				Type:      EventLevelUp,
				SourceID:  "player1",
				TargetID:  "",
				Data:      map[string]interface{}{"oldLevel": 1, "newLevel": 2},
				Timestamp: 12345,
			},
			setupFunc: func(es *EventSystem) chan bool {
				called := make(chan bool, 1)
				es.Subscribe(EventLevelUp, func(event GameEvent) {
					called <- true
				})
				return called
			},
		},
		{
			name: "Emit_DamageEvent_HandlerCalled",
			event: GameEvent{
				Type:      EventDamage,
				SourceID:  "enemy1",
				TargetID:  "player1",
				Data:      map[string]interface{}{"damage": 10, "damageType": "physical"},
				Timestamp: 12346,
			},
			setupFunc: func(es *EventSystem) chan bool {
				called := make(chan bool, 1)
				es.Subscribe(EventDamage, func(event GameEvent) {
					called <- true
				})
				return called
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventSystem := NewEventSystem()
			called := tt.setupFunc(eventSystem)

			// Emit the event
			eventSystem.Emit(tt.event)

			// Wait for handler to be called (async)
			select {
			case <-called:
				// Handler was called successfully
			case <-time.After(100 * time.Millisecond):
				t.Error("Emit() failed to call registered handler within timeout")
			}
		})
	}
}

// TestEventSystem_Emit_MultipleHandlers tests emitting events to multiple handlers
func TestEventSystem_Emit_MultipleHandlers(t *testing.T) {
	eventSystem := NewEventSystem()
	eventType := EventLevelUp

	// Create channels for tracking handler calls
	handler1Called := make(chan bool, 1)
	handler2Called := make(chan bool, 1)
	handler3Called := make(chan bool, 1)

	// Subscribe multiple handlers
	eventSystem.Subscribe(eventType, func(event GameEvent) {
		handler1Called <- true
	})
	eventSystem.Subscribe(eventType, func(event GameEvent) {
		handler2Called <- true
	})
	eventSystem.Subscribe(eventType, func(event GameEvent) {
		handler3Called <- true
	})

	// Emit event
	event := GameEvent{
		Type:      eventType,
		SourceID:  "player1",
		Data:      map[string]interface{}{"test": "data"},
		Timestamp: 12345,
	}
	eventSystem.Emit(event)

	// Verify all handlers were called
	timeout := time.After(200 * time.Millisecond)
	handlersCalled := 0

	for handlersCalled < 3 {
		select {
		case <-handler1Called:
			handlersCalled++
		case <-handler2Called:
			handlersCalled++
		case <-handler3Called:
			handlersCalled++
		case <-timeout:
			t.Errorf("Emit() failed to call all handlers, only %d out of 3 handlers were called", handlersCalled)
			return
		}
	}
}

// TestEventSystem_Emit_NoHandlers tests emitting events when no handlers are registered
func TestEventSystem_Emit_NoHandlers(t *testing.T) {
	eventSystem := NewEventSystem()

	event := GameEvent{
		Type:      EventLevelUp,
		SourceID:  "player1",
		Data:      map[string]interface{}{"test": "data"},
		Timestamp: 12345,
	}

	// This should not panic or cause errors
	eventSystem.Emit(event)

	// Verify no handlers exist
	eventSystem.mu.RLock()
	handlers := eventSystem.handlers[EventLevelUp]
	eventSystem.mu.RUnlock()

	if len(handlers) != 0 {
		t.Errorf("Expected no handlers, but found %d", len(handlers))
	}
}

// TestEventSystem_Emit_HandlerReceivesCorrectEvent tests that handlers receive the correct event data
func TestEventSystem_Emit_HandlerReceivesCorrectEvent(t *testing.T) {
	eventSystem := NewEventSystem()

	expectedEvent := GameEvent{
		Type:      EventDamage,
		SourceID:  "enemy1",
		TargetID:  "player1",
		Data:      map[string]interface{}{"damage": 25, "weapon": "sword"},
		Timestamp: 98765,
	}

	receivedEvent := make(chan GameEvent, 1)

	// Subscribe handler that captures the received event
	eventSystem.Subscribe(EventDamage, func(event GameEvent) {
		receivedEvent <- event
	})

	// Emit event
	eventSystem.Emit(expectedEvent)

	// Verify handler received correct event
	select {
	case event := <-receivedEvent:
		if event.Type != expectedEvent.Type {
			t.Errorf("Handler received wrong event type, expected %v, got %v", expectedEvent.Type, event.Type)
		}
		if event.SourceID != expectedEvent.SourceID {
			t.Errorf("Handler received wrong SourceID, expected %s, got %s", expectedEvent.SourceID, event.SourceID)
		}
		if event.TargetID != expectedEvent.TargetID {
			t.Errorf("Handler received wrong TargetID, expected %s, got %s", expectedEvent.TargetID, event.TargetID)
		}
		if event.Timestamp != expectedEvent.Timestamp {
			t.Errorf("Handler received wrong Timestamp, expected %d, got %d", expectedEvent.Timestamp, event.Timestamp)
		}
		if event.Data["damage"] != expectedEvent.Data["damage"] {
			t.Errorf("Handler received wrong damage data, expected %v, got %v", expectedEvent.Data["damage"], event.Data["damage"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Handler was not called within timeout")
	}
}

// TestEventSystem_ThreadSafety tests concurrent access to the event system
func TestEventSystem_ThreadSafety(t *testing.T) {
	eventSystem := NewEventSystem()

	// Number of concurrent operations
	numGoroutines := 50
	numEvents := 10

	var wg sync.WaitGroup

	// Concurrently subscribe handlers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eventSystem.Subscribe(EventLevelUp, func(event GameEvent) {
				// Handler that does minimal work
			})
		}()
	}

	// Concurrently emit events
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numEvents; j++ {
				event := GameEvent{
					Type:      EventLevelUp,
					SourceID:  "player1",
					Data:      map[string]interface{}{"iteration": j, "goroutine": id},
					Timestamp: int64(j),
				}
				eventSystem.Emit(event)
			}
		}(i)
	}

	// Wait for all operations to complete
	wg.Wait()

	// Verify final state
	eventSystem.mu.RLock()
	handlerCount := len(eventSystem.handlers[EventLevelUp])
	eventSystem.mu.RUnlock()

	if handlerCount != numGoroutines {
		t.Errorf("Thread safety test failed, expected %d handlers, got %d", numGoroutines, handlerCount)
	}
}

// TestEventTypes tests the event type constants
func TestEventTypes_ValidConstants(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		expected  EventType
	}{
		{"EventLevelUp", EventLevelUp, 0},
		{"EventDamage", EventDamage, 1},
		{"EventDeath", EventDeath, 2},
		{"EventItemPickup", EventItemPickup, 3},
		{"EventItemDrop", EventItemDrop, 4},
		{"EventMovement", EventMovement, 5},
		{"EventSpellCast", EventSpellCast, 6},
		{"EventQuestUpdate", EventQuestUpdate, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.eventType != tt.expected {
				t.Errorf("Event type %s has wrong value, expected %d, got %d", tt.name, tt.expected, tt.eventType)
			}
		})
	}
}

// TestGameEvent_StructInitialization tests GameEvent struct initialization
func TestGameEvent_StructInitialization(t *testing.T) {
	tests := []struct {
		name  string
		event GameEvent
	}{
		{
			name:  "EmptyGameEvent_InitializedCorrectly",
			event: GameEvent{},
		},
		{
			name: "PopulatedGameEvent_InitializedCorrectly",
			event: GameEvent{
				Type:      EventLevelUp,
				SourceID:  "player1",
				TargetID:  "target1",
				Data:      map[string]interface{}{"key": "value"},
				Timestamp: 123456789,
			},
		},
		{
			name: "GameEventWithComplexData_InitializedCorrectly",
			event: GameEvent{
				Type:     EventDamage,
				SourceID: "enemy_orc",
				TargetID: "player_warrior",
				Data: map[string]interface{}{
					"damage":      50,
					"damageType":  "physical",
					"isCritical":  true,
					"weaponUsed":  "battle_axe",
					"resistances": []string{"fire", "poison"},
				},
				Timestamp: 987654321,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the struct can be created and fields are accessible
			_ = tt.event.Type
			_ = tt.event.SourceID
			_ = tt.event.TargetID
			_ = tt.event.Data
			_ = tt.event.Timestamp

			// Test data access if present
			if tt.event.Data != nil {
				for key, value := range tt.event.Data {
					if value == nil {
						t.Errorf("GameEvent data key %s has nil value", key)
					}
				}
			}
		})
	}
}

// TestGetCurrentGameTick tests the getCurrentGameTick function and SetCurrentGameTick
func TestGetCurrentGameTick_ReturnsConsistentValue(t *testing.T) {
	// Store original value to restore after test
	originalTick := getCurrentGameTick()
	defer SetCurrentGameTick(originalTick)

	// Test that multiple reads return consistent values
	tick1 := getCurrentGameTick()
	tick2 := getCurrentGameTick()
	if tick1 != tick2 {
		t.Errorf("getCurrentGameTick() returned inconsistent values: %d and %d", tick1, tick2)
	}

	// Test SetCurrentGameTick updates the value
	testValue := int64(12345)
	SetCurrentGameTick(testValue)
	tick3 := getCurrentGameTick()
	if tick3 != testValue {
		t.Errorf("getCurrentGameTick() expected %d after SetCurrentGameTick, got %d", testValue, tick3)
	}

	// Verify GetCurrentGameTick (exported) matches getCurrentGameTick (unexported)
	if GetCurrentGameTick() != getCurrentGameTick() {
		t.Error("GetCurrentGameTick() and getCurrentGameTick() return different values")
	}
}

// TestGameTimeTracker_ThreadSafety tests concurrent access to game time tracker
func TestGameTimeTracker_ThreadSafety(t *testing.T) {
	// Store original value to restore after test
	originalTick := getCurrentGameTick()
	defer SetCurrentGameTick(originalTick)

	SetCurrentGameTick(0)
	done := make(chan bool)
	iterations := 1000

	// Concurrent writers
	go func() {
		for i := 0; i < iterations; i++ {
			SetCurrentGameTick(int64(i))
		}
		done <- true
	}()

	// Concurrent readers
	go func() {
		for i := 0; i < iterations; i++ {
			_ = getCurrentGameTick()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// No panic or race condition means success
}

// TestDefaultEventSystem tests that the default event system is properly initialized
func TestDefaultEventSystem_Initialization(t *testing.T) {
	if defaultEventSystem == nil {
		t.Fatal("defaultEventSystem is nil")
	}

	if defaultEventSystem.handlers == nil {
		t.Error("defaultEventSystem has nil handlers map")
	}

	// Test that we can subscribe to the default event system
	called := make(chan bool, 1)
	defaultEventSystem.Subscribe(EventMovement, func(event GameEvent) {
		called <- true
	})

	// Emit an event to test the subscription
	event := GameEvent{
		Type:      EventMovement,
		SourceID:  "test_entity",
		Data:      map[string]interface{}{"x": 10, "y": 20},
		Timestamp: getCurrentGameTick(),
	}
	defaultEventSystem.Emit(event)

	// Verify handler was called
	select {
	case <-called:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Default event system failed to call handler")
	}
}

// TestEmitLevelUpEvent tests the emitLevelUpEvent function
func TestEmitLevelUpEvent_CreatesCorrectEvent(t *testing.T) {
	// Save the original default event system
	originalEventSystem := defaultEventSystem

	// Create a new event system for testing
	testEventSystem := NewEventSystem()

	// Temporarily replace the default event system
	defaultEventSystem = testEventSystem

	// Restore the original default event system after test
	defer func() {
		defaultEventSystem = originalEventSystem
	}()

	// Set up a channel to capture the emitted event
	eventReceived := make(chan GameEvent, 1)

	// Subscribe to level up events
	testEventSystem.Subscribe(EventLevelUp, func(event GameEvent) {
		eventReceived <- event
	})

	// Test data
	playerID := "test_player_123"
	oldLevel := 5
	newLevel := 6

	// Call the function under test
	emitLevelUpEvent(playerID, oldLevel, newLevel)

	// Verify the event was emitted correctly
	select {
	case event := <-eventReceived:
		if event.Type != EventLevelUp {
			t.Errorf("emitLevelUpEvent() created wrong event type, expected %v, got %v", EventLevelUp, event.Type)
		}
		if event.SourceID != playerID {
			t.Errorf("emitLevelUpEvent() created wrong SourceID, expected %s, got %s", playerID, event.SourceID)
		}

		// Check event data
		if event.Data["oldLevel"] != oldLevel {
			t.Errorf("emitLevelUpEvent() created wrong oldLevel data, expected %d, got %v", oldLevel, event.Data["oldLevel"])
		}
		if event.Data["newLevel"] != newLevel {
			t.Errorf("emitLevelUpEvent() created wrong newLevel data, expected %d, got %v", newLevel, event.Data["newLevel"])
		}

		// Verify timestamp is set
		if event.Timestamp != getCurrentGameTick() {
			t.Errorf("emitLevelUpEvent() created wrong timestamp, expected %d, got %d", getCurrentGameTick(), event.Timestamp)
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("emitLevelUpEvent() failed to emit event within timeout")
	}
}

// TestEmitLevelUpEvent_MultipleEvents tests emitting multiple level up events
func TestEmitLevelUpEvent_MultipleEvents(t *testing.T) {
	// Save the original default event system
	originalEventSystem := defaultEventSystem

	// Create a new event system for testing
	testEventSystem := NewEventSystem()

	// Temporarily replace the default event system
	defaultEventSystem = testEventSystem

	// Restore the original default event system after test
	defer func() {
		defaultEventSystem = originalEventSystem
	}()

	// Set up a channel to capture emitted events
	eventsReceived := make(chan GameEvent, 10)

	// Subscribe to level up events
	testEventSystem.Subscribe(EventLevelUp, func(event GameEvent) {
		eventsReceived <- event
	})

	// Test data for multiple events
	testCases := []struct {
		playerID string
		oldLevel int
		newLevel int
	}{
		{"player1", 1, 2},
		{"player2", 10, 11},
		{"player3", 25, 26},
	}

	// Emit multiple events
	for _, tc := range testCases {
		emitLevelUpEvent(tc.playerID, tc.oldLevel, tc.newLevel)
	}

	// Verify all events were received
	receivedCount := 0
	timeout := time.After(200 * time.Millisecond)

	for receivedCount < len(testCases) {
		select {
		case event := <-eventsReceived:
			receivedCount++

			// Find the matching test case
			var matchedCase *struct {
				playerID string
				oldLevel int
				newLevel int
			}

			for i := range testCases {
				if testCases[i].playerID == event.SourceID {
					matchedCase = &testCases[i]
					break
				}
			}

			if matchedCase == nil {
				t.Errorf("Received event for unexpected player: %s", event.SourceID)
				continue
			}

			if event.Data["oldLevel"] != matchedCase.oldLevel {
				t.Errorf("Wrong oldLevel for player %s, expected %d, got %v",
					matchedCase.playerID, matchedCase.oldLevel, event.Data["oldLevel"])
			}
			if event.Data["newLevel"] != matchedCase.newLevel {
				t.Errorf("Wrong newLevel for player %s, expected %d, got %v",
					matchedCase.playerID, matchedCase.newLevel, event.Data["newLevel"])
			}

		case <-timeout:
			t.Fatalf("Only received %d out of %d expected events", receivedCount, len(testCases))
		}
	}
}

// TestInitFunction tests that the init function properly sets up the default event handler
func TestInitFunction_DefaultLevelUpHandler(t *testing.T) {
	// Create a test event that would trigger the default handler
	testEvent := GameEvent{
		Type:     EventLevelUp,
		SourceID: "test_player",
		Data: map[string]interface{}{
			"oldLevel": 5,
			"newLevel": 6,
		},
		Timestamp: getCurrentGameTick(),
	}

	// Since the init function registers a handler that logs,
	// we can't easily test the log output without modifying the logger.
	// Instead, we verify that handlers are registered for EventLevelUp
	defaultEventSystem.mu.RLock()
	handlers := defaultEventSystem.handlers[EventLevelUp]
	defaultEventSystem.mu.RUnlock()

	if len(handlers) == 0 {
		t.Error("init() function failed to register default EventLevelUp handler")
	}

	// Test that emitting an event doesn't panic (which would indicate handler issues)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Default EventLevelUp handler panicked: %v", r)
		}
	}()

	// Emit event to test the default handler doesn't crash
	defaultEventSystem.Emit(testEvent)

	// Give the async handler time to execute
	time.Sleep(10 * time.Millisecond)
}

// TestEventSystemConfig_StructInitialization tests EventSystemConfig struct initialization
func TestEventSystemConfig_StructInitialization(t *testing.T) {
	tests := []struct {
		name   string
		config EventSystemConfig
	}{
		{
			name:   "EmptyEventSystemConfig_InitializedCorrectly",
			config: EventSystemConfig{},
		},
		{
			name: "PopulatedEventSystemConfig_InitializedCorrectly",
			config: EventSystemConfig{
				RegisteredTypes: []EventType{EventLevelUp, EventDamage, EventDeath},
				HandlerCount: map[EventType]int{
					EventLevelUp: 2,
					EventDamage:  1,
					EventDeath:   0,
				},
				AsyncHandling: true,
			},
		},
		{
			name: "EventSystemConfigWithAllEventTypes_InitializedCorrectly",
			config: EventSystemConfig{
				RegisteredTypes: []EventType{
					EventLevelUp, EventDamage, EventDeath, EventItemPickup,
					EventItemDrop, EventMovement, EventSpellCast, EventQuestUpdate,
				},
				HandlerCount: map[EventType]int{
					EventLevelUp:     1,
					EventDamage:      2,
					EventDeath:       1,
					EventItemPickup:  3,
					EventItemDrop:    1,
					EventMovement:    5,
					EventSpellCast:   2,
					EventQuestUpdate: 1,
				},
				AsyncHandling: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the struct can be created and fields are accessible
			_ = tt.config.RegisteredTypes
			_ = tt.config.HandlerCount
			_ = tt.config.AsyncHandling

			// Verify RegisteredTypes slice
			if tt.config.RegisteredTypes != nil {
				for _, eventType := range tt.config.RegisteredTypes {
					if eventType < EventLevelUp || eventType > EventQuestUpdate {
						t.Errorf("Invalid event type in RegisteredTypes: %v", eventType)
					}
				}
			}

			// Verify HandlerCount map
			if tt.config.HandlerCount != nil {
				for eventType, count := range tt.config.HandlerCount {
					if count < 0 {
						t.Errorf("Negative handler count for event type %v: %d", eventType, count)
					}
					if eventType < EventLevelUp || eventType > EventQuestUpdate {
						t.Errorf("Invalid event type in HandlerCount: %v", eventType)
					}
				}
			}
		})
	}
}
