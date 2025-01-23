package server

import (
	"fmt"
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

// processEffectTick handles the processing of a single effect tick in the game state.
// It determines the effect type and routes the processing to the appropriate handler.
//
// Parameters:
//   - effect: *game.Effect - The effect to process. Must not be nil.
//
// Returns:
//   - error: Returns nil on success, or an error if:
//   - The effect parameter is nil
//   - The effect type is unknown/unsupported
//
// Related:
//   - processDamageEffect
//   - processHealEffect
//   - processStatEffect
//
// Handles effect types:
//   - EffectDamageOverTime
//   - EffectHealOverTime
//   - EffectStatBoost
//   - EffectStatPenalty
func (gs *GameState) processEffectTick(effect *game.Effect) error {
	if effect == nil {
		return fmt.Errorf("nil effect")
	}

	switch effect.Type {
	case game.EffectDamageOverTime:
		return gs.processDamageEffect(effect)
	case game.EffectHealOverTime:
		return gs.processHealEffect(effect)
	case game.EffectStatBoost, game.EffectStatPenalty:
		return gs.processStatEffect(effect)
	default:
		return fmt.Errorf("unknown effect type: %s", effect.Type)
	}
}

// processDamageEffect applies damage to a target character based on the provided effect.
// It locates the target in the world state and reduces their HP by the effect magnitude.
//
// Parameters:
//   - effect: *game.Effect - Contains target ID and damage magnitude to apply
//
// Returns:
//   - error - Returns nil if damage was successfully applied, or an error if:
//   - Target ID does not exist in world state
//   - Target is not a Character type that can receive damage
//
// Related:
//   - game.Character
//   - game.Effect
//   - GameState.WorldState
func (gs *GameState) processDamageEffect(effect *game.Effect) error {
	target, exists := gs.WorldState.Objects[effect.TargetID]
	if !exists {
		return fmt.Errorf("invalid effect target")
	}

	if char, ok := target.(*game.Character); ok {
		damage := int(effect.Magnitude)
		char.HP -= damage
		if char.HP < 0 {
			char.HP = 0
		}
		return nil
	}
	return fmt.Errorf("target cannot receive damage")
}

// processHealEffect applies a healing effect to a target character in the game world.
// It increases the target's HP by the effect magnitude, up to their max HP.
//
// Parameters:
//   - effect: *game.Effect - The healing effect to process, must contain:
//   - TargetID: ID of the character to heal
//   - Magnitude: Amount of HP to heal
//
// Returns:
//   - error: Returns nil on success, or an error if:
//   - Target does not exist in world state
//   - Target is not a Character type
//
// Related:
//   - game.Character
//   - game.Effect
//   - GameState.WorldState
func (gs *GameState) processHealEffect(effect *game.Effect) error {
	target, exists := gs.WorldState.Objects[effect.TargetID]
	if !exists {
		return fmt.Errorf("invalid effect target")
	}

	if char, ok := target.(*game.Character); ok {
		healAmount := int(effect.Magnitude)
		char.HP = min(char.HP+healAmount, char.MaxHP)
		return nil
	}
	return fmt.Errorf("target cannot be healed")
}

// ProcessStatEffect applies a stat modification effect to a character target.
//
// Parameters:
//   - effect: *game.Effect - Contains the target ID, stat to modify, and magnitude
//     of the modification. Must have valid StatAffected and Magnitude fields.
//
// Returns:
//
//	error - Returns nil if successful, or an error if:
//	- Target ID doesn't exist in WorldState
//	- Target is not a Character type
//	- StatAffected is not a valid stat name
//
// StatAffected must be one of: strength, dexterity, constitution, intelligence,
// wisdom, charisma
//
// Related types:
//   - game.Effect
//   - game.Character
func (gs *GameState) processStatEffect(effect *game.Effect) error {
	target, exists := gs.WorldState.Objects[effect.TargetID]
	if !exists {
		return fmt.Errorf("invalid effect target")
	}

	if char, ok := target.(*game.Character); ok {
		// Apply stat modification based on effect type
		magnitude := int(effect.Magnitude)
		switch effect.StatAffected {
		case "strength":
			char.Strength += magnitude
		case "dexterity":
			char.Dexterity += magnitude
		case "constitution":
			char.Constitution += magnitude
		case "intelligence":
			char.Intelligence += magnitude
		case "wisdom":
			char.Wisdom += magnitude
		case "charisma":
			char.Charisma += magnitude
		default:
			return fmt.Errorf("unknown stat type: %s", effect.StatAffected)
		}
		return nil
	}
	return fmt.Errorf("target cannot receive stat effects")
}
