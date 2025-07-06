package server

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// GameState represents the core game state container managing all dynamic game elements.
// It provides thread-safe access to the world state, turn sequencing, time tracking,
// and player session management.
//
// Fields:
//   - WorldState: Holds the current state of the game world including entities, items, etc
//   - TurnManager: Manages turn order and action resolution for game entities
//   - TimeManager: Tracks game time progression and scheduling
//   - Sessions: Maps session IDs to active PlayerSession objects
//   - mu: Provides thread-safe access to state
//   - updates: Channel for broadcasting state changes to listeners
//
// Thread Safety:
// All public methods are protected by mutex to ensure thread-safe concurrent access.
// The updates channel allows for non-blocking notifications of state changes.
//
// Related Types:
//   - game.World
//   - TurnManager
//   - TimeManager
//   - PlayerSession
type GameState struct {
	WorldState  *game.World               `yaml:"state_world"`
	TurnManager *TurnManager              `yaml:"state_turns"`
	TimeManager *TimeManager              `yaml:"state_time"`
	Sessions    map[string]*PlayerSession `yaml:"state_sessions"`
	Version     int                       `yaml:"state_version"`

	// Locking implementation
	stateMu   sync.RWMutex `yaml:"-"` // Primary state mutex
	worldMu   sync.RWMutex `yaml:"-"` // World state mutex
	sessionMu sync.RWMutex `yaml:"-"` // Session mutex
	turnMu    sync.RWMutex `yaml:"-"` // Turn manager mutex

	// State caching
	cachedState  atomic.Value `yaml:"-"` // Cached state data
	cacheVersion int32        `yaml:"-"` // Atomic cache version

	updates chan StateUpdate `yaml:"-"` // Update channel
}

// AddPlayer initializes a new player in the game state
func (gs *GameState) AddPlayer(session *PlayerSession) {
	if session == nil || session.Player == nil {
		return
	}

	gs.worldMu.Lock()
	defer gs.worldMu.Unlock()

	// Initialize WorldState if nil
	if gs.WorldState == nil {
		gs.WorldState = &game.World{}
	}

	// Initialize Objects map if nil
	if gs.WorldState.Objects == nil {
		gs.WorldState.Objects = make(map[string]game.GameObject)
	}

	// Add player to world state
	gs.WorldState.Objects[session.Player.GetID()] = session.Player
}

// GetState returns the current game state as a map.
func (gs *GameState) GetState() map[string]interface{} {
	// Try to get cached state first
	if cached := gs.cachedState.Load(); cached != nil {
		if state, ok := cached.(map[string]interface{}); ok {
			if atomic.LoadInt32(&gs.cacheVersion) == int32(gs.Version) {
				return state
			}
		}
	}

	// Cache miss - generate new state with minimal locking
	gs.stateMu.RLock()
	version := gs.Version
	gs.stateMu.RUnlock()

	state := make(map[string]interface{})

	// Get world state with separate lock
	gs.worldMu.RLock()
	state["world"] = gs.WorldState.Serialize()
	gs.worldMu.RUnlock()

	// Get time state
	state["time"] = gs.TimeManager.Serialize()

	// Get turn state with separate lock
	gs.turnMu.RLock()
	state["turns"] = gs.TurnManager.Serialize()
	gs.turnMu.RUnlock()

	// Get session data with separate lock
	gs.sessionMu.RLock()
	sessions := make(map[string]interface{})
	for id, session := range gs.Sessions {
		sessions[id] = session.PublicData()
	}
	gs.sessionMu.RUnlock()
	state["sessions"] = sessions

	state["version"] = version

	// Update cache
	gs.cachedState.Store(state)
	atomic.StoreInt32(&gs.cacheVersion, int32(version))

	return state
}

func (gs *GameState) validate() error {
	if gs.WorldState == nil ||
		gs.TimeManager == nil ||
		gs.TurnManager == nil ||
		gs.Sessions == nil {
		return fmt.Errorf("missing required state components")
	}
	return nil
}

// UpdateState applies updates to the game state.
func (gs *GameState) UpdateState(updates map[string]interface{}) error {
	// Create snapshot for rollback under read lock
	gs.stateMu.RLock()
	snapshot := gs.createSnapshot()
	version := gs.Version
	gs.stateMu.RUnlock()

	// Acquire locks in consistent order to prevent deadlocks
	gs.worldMu.Lock()
	gs.sessionMu.Lock()
	gs.turnMu.Lock()
	gs.stateMu.Lock()
	defer func() {
		gs.stateMu.Unlock()
		gs.turnMu.Unlock()
		gs.sessionMu.Unlock()
		gs.worldMu.Unlock()
	}()

	// Verify version hasn't changed
	if gs.Version != version {
		return fmt.Errorf("state version changed during update")
	}

	// Apply updates with timeout
	done := make(chan error, 1)
	go func() {
		done <- gs.applyUpdates(updates)
	}()

	select {
	case err := <-done:
		if err != nil {
			gs.rollback(snapshot)
			return err
		}
	case <-time.After(5 * time.Second):
		gs.rollback(snapshot)
		return fmt.Errorf("update timed out")
	}

	gs.Version++
	// Invalidate cache
	atomic.StoreInt32(&gs.cacheVersion, -1)

	return nil
}

func (gs *GameState) rollback(snapshot any) {
	if snapshotState, ok := snapshot.(*GameState); ok {
		// Restore all state components from snapshot
		gs.WorldState = snapshotState.WorldState
		gs.TimeManager = snapshotState.TimeManager
		gs.TurnManager = snapshotState.TurnManager
		gs.Sessions = snapshotState.Sessions
		gs.Version = snapshotState.Version

		logrus.WithField("version", gs.Version).Info("rolled back game state to previous snapshot")
	} else {
		logrus.Error("invalid snapshot type for rollback")
	}
}

func (gs *GameState) applyUpdates(updates map[string]interface{}) error {
	// Handle world state updates
	if worldUpdates, ok := updates["world"].(map[string]interface{}); ok {
		if err := gs.WorldState.Update(worldUpdates); err != nil {
			return fmt.Errorf("world update failed: %w", err)
		}
	}

	// Handle time manager updates
	if timeUpdates, ok := updates["time"].(map[string]interface{}); ok {
		if currentTime, ok := timeUpdates["current_time"].(map[string]interface{}); ok {
			if scale, ok := currentTime["time_scale"].(float64); ok {
				gs.TimeManager.TimeScale = scale
			}
		}
	}

	// Handle turn manager updates
	if turnUpdates, ok := updates["turns"].(map[string]interface{}); ok {
		if err := gs.TurnManager.Update(turnUpdates); err != nil {
			return fmt.Errorf("turn update failed: %v", err)
		}
	}

	// Handle session updates
	if sessionUpdates, ok := updates["sessions"].(map[string]interface{}); ok {
		for id, update := range sessionUpdates {
			if session, exists := gs.Sessions[id]; exists {
				if updateMap, ok := update.(map[string]interface{}); ok {
					if err := session.Update(updateMap); err != nil {
						return fmt.Errorf("session update failed for %s: %v", id, err)
					}
				}
			}
		}
	}

	return nil
}

func (gs *GameState) createSnapshot() any {
	// Create a deep copy of the game state for rollback purposes
	snapshot := &GameState{
		WorldState: gs.WorldState.Clone(), // Assuming World has a Clone method
		TimeManager: &TimeManager{
			CurrentTime:     gs.TimeManager.CurrentTime,
			TimeScale:       gs.TimeManager.TimeScale,
			LastTick:        gs.TimeManager.LastTick,
			ScheduledEvents: make([]ScheduledEvent, len(gs.TimeManager.ScheduledEvents)),
		},
		TurnManager: gs.TurnManager.Clone(), // Assuming TurnManager has a Clone method
		Sessions:    make(map[string]*PlayerSession),
	}

	// Copy scheduled events
	copy(snapshot.TimeManager.ScheduledEvents, gs.TimeManager.ScheduledEvents)

	// Copy sessions
	for id, session := range gs.Sessions {
		snapshot.Sessions[id] = session.Clone() // Assuming PlayerSession has a Clone method
	}

	return snapshot
}

// TimeManager handles game time progression and scheduled event management.
// It maintains the current game time, controls time progression speed,
// and manages a queue of scheduled future events.
//
// Fields:
//   - CurrentTime: The current in-game time represented as a GameTime struct
//   - TimeScale: Multiplier that controls how fast game time progresses relative to real time (e.g. 2.0 = twice as fast)
//   - LastTick: Real-world timestamp of the most recent time update
//   - ScheduledEvents: Slice of pending events to be triggered at specific game times
//
// Related types:
//   - game.GameTime - Represents a point in game time
//   - ScheduledEvent - Defines a future event to occur at a specific game time
type TimeManager struct {
	CurrentTime     game.GameTime    `yaml:"time_current"`          // Current game time
	TimeScale       float64          `yaml:"time_scale"`            // Time progression rate
	LastTick        time.Time        `yaml:"time_last_tick"`        // Last update time
	ScheduledEvents []ScheduledEvent `yaml:"time_scheduled_events"` // Pending events
}

// Serialize returns a map representation of the TimeManager state
func (t *TimeManager) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"current_time": map[string]interface{}{
			"real_time":  t.CurrentTime.RealTime,
			"game_ticks": t.CurrentTime.GameTicks,
			"time_scale": t.CurrentTime.TimeScale,
		},
		"time_scale": t.TimeScale,
		"last_tick":  t.LastTick,
		"events": func() []map[string]interface{} {
			events := make([]map[string]interface{}, len(t.ScheduledEvents))
			for i, event := range t.ScheduledEvents {
				events[i] = map[string]interface{}{
					"id":           event.EventID,
					"type":         event.EventType,
					"trigger_time": event.TriggerTime,
					"parameters":   event.Parameters,
					"repeating":    event.Repeating,
				}
			}
			return events
		}(),
	}
}

// ScheduledEvent represents a future event that will be triggered at a specific game time.
// It is used to schedule in-game events like monster spawns, weather changes, or quest updates.
//
// Fields:
//   - EventID: Unique string identifier for the event
//   - EventType: Category/type of the event (e.g. "spawn", "weather", etc)
//   - TriggerTime: The game.GameTime when this event should execute
//   - Parameters: Additional string data needed for the event execution
//   - Repeating: If true, the event will reschedule itself after triggering
//
// Related types:
//   - game.GameTime: Represents the in-game time when event triggers
type ScheduledEvent struct {
	EventID     string        `yaml:"event_id"`           // Event identifier
	EventType   string        `yaml:"event_type"`         // Type of event
	TriggerTime game.GameTime `yaml:"event_trigger_time"` // When to trigger
	Parameters  []string      `yaml:"event_parameters"`   // Event data
	Repeating   bool          `yaml:"event_is_repeating"` // Whether it repeats
}

// ScriptContext represents the execution state and variables of a running script in the game.
// It maintains context between script executions including variables and timing.
//
// Fields:
//   - ScriptID: Unique identifier string for the script
//   - Variables: Map storing script state variables and their values
//   - LastExecuted: Timestamp of when the script was last run
//   - IsActive: Boolean flag indicating if script is currently executing
//
// Related types:
//   - Server.Scripts (map[string]*ScriptContext)
//   - ScriptEngine interface
//
// Thread-safety: This struct should be protected by a mutex when accessed concurrently
type ScriptContext struct {
	ScriptID     string                 `yaml:"script_id"`            // Script identifier
	Variables    map[string]interface{} `yaml:"script_variables"`     // Script state
	LastExecuted time.Time              `yaml:"script_last_executed"` // Last run timestamp
	IsActive     bool                   `yaml:"script_is_active"`     // Execution state
}

// NewTimeManager creates and initializes a new TimeManager instance.
//
// The TimeManager handles game time tracking, time scaling, and scheduled event management.
// It maintains the current game time, real time mapping, and a list of scheduled events.
//
// Returns:
//   - *TimeManager: A new TimeManager instance initialized with:
//   - Current time set to now
//   - Game ticks starting at 0
//   - Default time scale of 1.0
//   - Empty scheduled events list
//
// Related types:
//   - game.GameTime
//   - ScheduledEvent
func NewTimeManager() *TimeManager {
	logrus.WithFields(logrus.Fields{
		"function": "NewTimeManager",
	}).Debug("creating new time manager")

	tm := &TimeManager{
		CurrentTime: game.GameTime{
			RealTime:  time.Now(),
			GameTicks: 0,
			TimeScale: 1.0,
		},
		TimeScale:       1.0,
		LastTick:        time.Now(),
		ScheduledEvents: make([]ScheduledEvent, 0),
	}

	logrus.WithFields(logrus.Fields{
		"function":  "NewTimeManager",
		"timeScale": tm.TimeScale,
		"lastTick":  tm.LastTick,
	}).Info("time manager initialized")

	return tm
}
