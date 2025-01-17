package server

import (
	"fmt"
	"goldbox-rpg/pkg/game"
	"sync"
	"time"
)

// GameState represents the complete server-side game state
type GameState struct {
	WorldState  *game.World               `yaml:"state_world"`    // Current world state
	TurnManager *TurnManager              `yaml:"state_turns"`    // Turn management
	TimeManager *TimeManager              `yaml:"state_time"`     // Time tracking
	Sessions    map[string]*PlayerSession `yaml:"state_sessions"` // Active player sessions
	mu          sync.RWMutex              `yaml:"-"`              // State mutex
	updates     chan StateUpdate          `yaml:"-"`              // Update channel
}

// TimeManager handles game time progression and scheduled events
type TimeManager struct {
	CurrentTime     game.GameTime    `yaml:"time_current"`          // Current game time
	TimeScale       float64          `yaml:"time_scale"`            // Time progression rate
	LastTick        time.Time        `yaml:"time_last_tick"`        // Last update time
	ScheduledEvents []ScheduledEvent `yaml:"time_scheduled_events"` // Pending events
}

// ScheduledEvent represents a future game event
type ScheduledEvent struct {
	EventID     string        `yaml:"event_id"`           // Event identifier
	EventType   string        `yaml:"event_type"`         // Type of event
	TriggerTime game.GameTime `yaml:"event_trigger_time"` // When to trigger
	Parameters  []string      `yaml:"event_parameters"`   // Event data
	Repeating   bool          `yaml:"event_is_repeating"` // Whether it repeats
}

// ScriptContext represents the NPC behavior script state
type ScriptContext struct {
	ScriptID     string                 `yaml:"script_id"`            // Script identifier
	Variables    map[string]interface{} `yaml:"script_variables"`     // Script state
	LastExecuted time.Time              `yaml:"script_last_executed"` // Last run timestamp
	IsActive     bool                   `yaml:"script_is_active"`     // Execution state
}

func NewTimeManager() *TimeManager {
	return &TimeManager{
		CurrentTime: game.GameTime{
			RealTime:  time.Now(),
			GameTicks: 0,
			TimeScale: 1.0,
		},
		TimeScale:       1.0,
		LastTick:        time.Now(),
		ScheduledEvents: make([]ScheduledEvent, 0),
	}
}

// Add these methods to GameState
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
