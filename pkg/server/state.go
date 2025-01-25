package server

import (
	"sync"
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
	WorldState  *game.World               `yaml:"state_world"`    // Current world state
	TurnManager *TurnManager              `yaml:"state_turns"`    // Turn management
	TimeManager *TimeManager              `yaml:"state_time"`     // Time tracking
	Sessions    map[string]*PlayerSession `yaml:"state_sessions"` // Active player sessions
	mu          sync.RWMutex              `yaml:"-"`              // State mutex
	updates     chan StateUpdate          `yaml:"-"`              // Update channel
}

// AddPlayer initializes a new player in the game state
func (gs *GameState) AddPlayer(session *PlayerSession) {
	// Add player to world state and initialize their state
	gs.WorldState.Objects[session.Player.GetID()] = session.Player
}

// GetState returns the current game state in a format suitable for client consumption
func (s *GameState) GetState() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state := make(map[string]interface{})
	state["world"] = s.WorldState
	state["time"] = s.TimeManager.CurrentTime
	state["turns"] = s.TurnManager
	state["sessions"] = s.Sessions
	return state
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
