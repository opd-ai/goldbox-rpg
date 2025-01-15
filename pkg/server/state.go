package server

import (
	"sync"
	"time"

	"goldbox-rpg/pkg/game"
)

// PlayerSession represents an active player connection
type PlayerSession struct {
	SessionID  string       `yaml:"session_id"`  // Unique session identifier
	Player     *game.Player `yaml:"player"`      // Associated player
	LastActive time.Time    `yaml:"last_active"` // Last activity timestamp
	Connected  bool         `yaml:"connected"`   // Connection status
}

// GameState represents the complete server-side game state
type GameState struct {
	WorldState  *game.World               `yaml:"state_world"`    // Current world state
	TurnManager *TurnManager              `yaml:"state_turns"`    // Turn management
	TimeManager *TimeManager              `yaml:"state_time"`     // Time tracking
	Sessions    map[string]*PlayerSession `yaml:"state_sessions"` // Active player sessions
	mu          sync.RWMutex              `yaml:"-"`              // State mutex
	updates     chan StateUpdate          `yaml:"-"`              // Update channel
}

// TurnManager handles combat turns and initiative ordering
type TurnManager struct {
	CurrentRound   int                 `yaml:"turn_current_round"`    // Active combat round
	Initiative     []string            `yaml:"turn_initiative_order"` // Turn order by entity ID
	CurrentIndex   int                 `yaml:"turn_current_index"`    // Current actor index
	IsInCombat     bool                `yaml:"turn_in_combat"`        // Combat state flag
	CombatGroups   map[string][]string `yaml:"turn_combat_groups"`    // Allied entities
	DelayedActions []DelayedAction     `yaml:"turn_delayed_actions"`  // Pending actions
}

// TimeManager handles game time progression and scheduled events
type TimeManager struct {
	CurrentTime     game.GameTime    `yaml:"time_current"`          // Current game time
	TimeScale       float64          `yaml:"time_scale"`            // Time progression rate
	LastTick        time.Time        `yaml:"time_last_tick"`        // Last update time
	ScheduledEvents []ScheduledEvent `yaml:"time_scheduled_events"` // Pending events
}

// CombatState tracks active combat information
type CombatState struct {
	ActiveCombatants []string                 `yaml:"combat_active_entities"` // Entities in combat
	RoundCount       int                      `yaml:"combat_round_count"`     // Number of rounds
	CombatZone       game.Position            `yaml:"combat_zone_center"`     // Combat area center
	StatusEffects    map[string][]game.Effect `yaml:"combat_status_effects"`  // Active effects
}

// ScriptContext represents the NPC behavior script state
type ScriptContext struct {
	ScriptID     string                 `yaml:"script_id"`            // Script identifier
	Variables    map[string]interface{} `yaml:"script_variables"`     // Script state
	LastExecuted time.Time              `yaml:"script_last_executed"` // Last run timestamp
	IsActive     bool                   `yaml:"script_is_active"`     // Execution state
}

// StateUpdate represents a game state change notification
type StateUpdate struct {
	UpdateType string                 `yaml:"update_type"`      // Type of update
	EntityID   string                 `yaml:"update_entity_id"` // Affected entity
	ChangeData map[string]interface{} `yaml:"update_data"`      // Update details
	Timestamp  time.Time              `yaml:"update_timestamp"` // When it occurred
}

// DelayedAction represents a pending combat action
type DelayedAction struct {
	ActorID     string        `yaml:"action_actor_id"`     // Entity performing action
	ActionType  string        `yaml:"action_type"`         // Type of action
	Target      game.Position `yaml:"action_target_pos"`   // Target location
	TriggerTime game.GameTime `yaml:"action_trigger_time"` // When to execute
	Parameters  []string      `yaml:"action_parameters"`   // Additional data
}

// ScheduledEvent represents a future game event
type ScheduledEvent struct {
	EventID     string        `yaml:"event_id"`           // Event identifier
	EventType   string        `yaml:"event_type"`         // Type of event
	TriggerTime game.GameTime `yaml:"event_trigger_time"` // When to trigger
	Parameters  []string      `yaml:"event_parameters"`   // Event data
	Repeating   bool          `yaml:"event_is_repeating"` // Whether it repeats
}
