package game

import "time"

// EffectType defines the category of game effect
type EffectType string

const (
	// Combat effects
	EffectDamageOverTime EffectType = "damage_over_time"
	EffectHealOverTime   EffectType = "heal_over_time"
	EffectStun           EffectType = "stun"
	EffectRoot           EffectType = "root"

	// Stat modification effects
	EffectStatBoost   EffectType = "stat_boost"
	EffectStatPenalty EffectType = "stat_penalty"

	// Status effects
	EffectPoison    EffectType = "poison"
	EffectBlind     EffectType = "blind"
	EffectInvisible EffectType = "invisible"
	EffectHaste     EffectType = "haste"
	EffectSlow      EffectType = "slow"

	// Special effects
	EffectAura    EffectType = "aura"
	EffectShield  EffectType = "shield"
	EffectReflect EffectType = "reflect"
)

// Effect represents a temporary modification to an entity's state or attributes
type Effect struct {
	ID          string     `yaml:"effect_id"`   // Unique identifier
	Type        EffectType `yaml:"effect_type"` // Category of effect
	Name        string     `yaml:"effect_name"` // Display name
	Description string     `yaml:"effect_desc"` // Effect description

	// Timing
	StartTime time.Time `yaml:"effect_start"`     // When effect began
	Duration  Duration  `yaml:"effect_duration"`  // How long it lasts
	TickRate  Duration  `yaml:"effect_tick_rate"` // How often it updates

	// Effect values
	Magnitude float64    `yaml:"effect_magnitude"` // Primary effect strength
	Modifiers []Modifier `yaml:"effect_modifiers"` // Additional modifications

	// Source tracking
	SourceID   string `yaml:"effect_source"`      // ID of effect creator
	SourceType string `yaml:"effect_source_type"` // Type of source (spell, item, etc)

	// State
	IsActive bool     `yaml:"effect_active"` // Whether effect is applied
	Stacks   int      `yaml:"effect_stacks"` // Number of accumulated stacks
	Tags     []string `yaml:"effect_tags"`   // Categorization/filtering
}

// Modifier represents a single stat or attribute modification
type Modifier struct {
	Stat      string    `yaml:"mod_stat"`      // Stat being modified
	Value     float64   `yaml:"mod_value"`     // Modification amount
	Operation ModOpType `yaml:"mod_operation"` // How to apply modification
}

// ModOpType defines how a modifier value is applied
type ModOpType string

const (
	ModAdd      ModOpType = "add"      // Direct addition
	ModMultiply ModOpType = "multiply" // Percentage increase
	ModSet      ModOpType = "set"      // Override value
	ModMin      ModOpType = "min"      // Minimum value
	ModMax      ModOpType = "max"      // Maximum value
)

// Duration represents a game time duration
type Duration struct {
	Rounds   int           `yaml:"duration_rounds"` // Combat rounds
	Turns    int           `yaml:"duration_turns"`  // Game turns
	RealTime time.Duration `yaml:"duration_real"`   // Real-world time
}

// Methods for Effect
func (e *Effect) IsExpired(currentTime time.Time) bool {
	if e.Duration.RealTime > 0 {
		return currentTime.After(e.StartTime.Add(e.Duration.RealTime))
	}
	return false // Handle round/turn based duration
}

func (e *Effect) ShouldTick(currentTime time.Time) bool {
	if e.TickRate.RealTime == 0 {
		return false
	}
	timeSinceStart := currentTime.Sub(e.StartTime)
	return timeSinceStart%e.TickRate.RealTime == 0
}

func (e *Effect) AddStack() bool {
	if e.Stacks < 99 { // Arbitrary max stacks
		e.Stacks++
		return true
	}
	return false
}

func (e *Effect) RemoveStack() bool {
	if e.Stacks > 0 {
		e.Stacks--
		return true
	}
	return false
}

// Helper functions
func NewEffect(effectType EffectType, duration Duration, magnitude float64) *Effect {
	return &Effect{
		ID:        NewUID(),
		Type:      effectType,
		StartTime: time.Now(),
		Duration:  duration,
		Magnitude: magnitude,
		IsActive:  true,
		Stacks:    1,
	}
}

// Create common effect types
func CreateDamageOverTimeEffect(dps float64, duration time.Duration) *Effect {
	return NewEffect(EffectDamageOverTime, Duration{RealTime: duration}, dps)
}

func CreateStatBoostEffect(stat string, bonus float64, duration Duration) *Effect {
	effect := NewEffect(EffectStatBoost, duration, bonus)
	effect.Modifiers = []Modifier{{
		Stat:      stat,
		Value:     bonus,
		Operation: ModAdd,
	}}
	return effect
}
