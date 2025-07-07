// Package game provides core RPG mechanics and systems for the GoldBox RPG Engine.
// This includes character management, effects, combat, spells, equipment, and world interactions.
package game

import (
	"sync"
	"time"
)

// Core types

// EffectType represents a type of effect that can be applied to a game entity
// in the RPG system. It is implemented as a string to allow for easy extensibility
// and readable effect definitions.
type EffectType string

// DamageType represents different types of damage that can be dealt in combat
type DamageType string

// DispelType represents different methods of dispelling effects
type DispelType string

// ImmunityType represents different types of immunity that characters can have
type ImmunityType int

// DispelPriority represents the priority level for effect dispelling
type DispelPriority int

// Constants
// EffectDamageOverTime represents an effect that deals damage to a target over a period of time.
// It is commonly used for effects like poison, burning, or bleeding that deal periodic damage.
// Effect constants are defined in constants.go
// Related effects: EffectPoison, EffectBurning, EffectBleeding
// Related damage types: DamagePhysical, DamageFire, DamagePoison

// Duration struct is defined in duration.go

// Effect represents a game effect
// Effect represents a game effect that can be applied to entities, modifying their stats or behavior over time.
// It contains all the information needed to track, apply and manage status effects in the game.
//
// Fields:
//   - ID: Unique identifier for the effect
//   - Type: Category/type of the effect (e.g. buff, debuff, dot)
//   - Name: Display name of the effect
//   - Description: Detailed description of what the effect does
//   - StartTime: When the effect was applied
//   - Duration: How long the effect lasts
//   - TickRate: How often the effect triggers/updates
//   - Magnitude: Strength/value of the effect
//   - DamageType: Type of damage if effect deals damage
//   - SourceID: ID of entity that applied the effect
//   - SourceType: Type of entity that applied the effect
//   - TargetID: ID of entity the effect is applied to
//   - StatAffected: Which stat the effect modifies
//   - IsActive: Whether effect is currently active
//   - Stacks: Number of times effect has stacked
//   - Tags: Labels for categorizing/filtering effects
//   - DispelInfo: Rules for removing/dispelling the effect
//   - Modifiers: List of stat/attribute modifications
//
// Related types:
//   - EffectType: Type definition for effect categories
//   - Duration: Custom time duration type
//   - DamageType: Enumeration of damage types
//   - DispelInfo: Rules for dispelling effects
//   - Modifier: Definition of stat modifications
type Effect struct {
	ID          string     `yaml:"effect_id"`
	Type        EffectType `yaml:"effect_type"`
	Name        string     `yaml:"effect_name"`
	Description string     `yaml:"effect_desc"`

	StartTime time.Time `yaml:"effect_start"`
	Duration  Duration  `yaml:"effect_duration"`
	TickRate  Duration  `yaml:"effect_tick_rate"`

	Magnitude  float64    `yaml:"effect_magnitude"`
	DamageType DamageType `yaml:"damage_type,omitempty"`

	SourceID   string `yaml:"effect_source"`
	SourceType string `yaml:"effect_source_type"`

	TargetID     string `yaml:"effect_target"`
	StatAffected string `yaml:"effect_stat_affected"`

	IsActive bool     `yaml:"effect_active"`
	Stacks   int      `yaml:"effect_stacks"`
	Tags     []string `yaml:"effect_tags"`

	DispelInfo DispelInfo `yaml:"dispel_info"`
	Modifiers  []Modifier `yaml:"effect_modifiers"`
}

// Modifier represents a modification to a game statistic or attribute.
// It defines how a specific stat should be modified through a mathematical operation.
//
// Fields:
//   - Stat: The name/identifier of the stat being modified
//   - Value: The numeric value to apply in the modification
//   - Operation: The type of mathematical operation to perform (e.g. add, multiply)
//
// Related types:
//   - ModOpType: Enum defining valid modification operations
//
// Usage example:
//
//	mod := Modifier{
//	  Stat: "health",
//	  Value: 10,
//	  Operation: ModAdd,
//	}
// Modifier struct is defined in modifier.go

// ModOpType represents the type of modification operation that can be applied to game attributes.
// It is implemented as a string type to allow for extensible operation types while maintaining
// type safety through constant definitions.
type ModOpType string

// ModOpType constants define supported mathematical operations for modifying stats/attributes.
// These are used by the Modifier type to specify how a stat value should be changed.
//
// Supported operations:
// - ModAdd: Adds the modifier value to the base stat
// - ModMultiply: Multiplies the base stat by the modifier value
// - ModSet: Sets the stat directly to the modifier value, ignoring the base value
//
// Related types:
// - Modifier: Uses these operations to define stat modifications
// ModOpType constants are defined in constants.go
// - Effect: Contains Modifiers that use these operations

// DispelInfo contains metadata about how a game effect can be dispelled or removed.
//
// Fields:
//   - Priority: Determines the order in which effects are dispelled (higher priority = dispelled first)
//   - Types: List of dispel types that can remove this effect (e.g. magic, poison, curse)
//   - Removable: Whether the effect can be removed at all
//
// Related types:
//   - DispelPriority: Priority level constants (0-100)
//   - DispelType: Type of dispel (magic, curse, poison, etc)
//   - Effect: Contains DispelInfo as a field
//
// Example usage:
//
//	info := DispelInfo{
//	    Priority: DispelPriorityNormal,
//	    Types: []DispelType{DispelMagic},
//	    Removable: true,
//	}
// DispelInfo struct is defined in dispel_info.go

// ImmunityData represents immunity effects that can be applied to game entities.
// It tracks the type, duration, resistance level and expiration time of immunities.
//
// Fields:
//   - Type: The type/category of immunity effect (ImmunityType)
//   - Duration: How long the immunity lasts (time.Duration)
//   - Resistance: A value between 0-1 representing immunity strength
//   - ExpiresAt: Timestamp when immunity effect ends
//
// Related types:
//   - ImmunityType: Enumeration of possible immunity types
//
// The immunity effect expires when current time exceeds ExpiresAt.
// Resistance of 1.0 means complete immunity, while 0.0 means no immunity.
type ImmunityData struct {
	Type       ImmunityType
	Duration   time.Duration
	Resistance float64
	ExpiresAt  time.Time
}

// EffectManager handles all temporary and permanent effects applied to an entity in the game.
// It manages active effects, base and current stats, immunities, resistances, and healing modifiers.
//
// The manager maintains thread-safe access to its data structures through a mutex.
//
// Fields:
//   - activeEffects: Maps effect IDs to Effect instances currently applied
//   - baseStats: Original unmodified stats of the entity
//   - currentStats: Current stats after applying all active effects
//   - immunities: Permanent immunity data mapped by effect type
//   - tempImmunities: Temporary immunity data mapped by effect type
//   - resistances: Damage/effect resistance multipliers (0-1) mapped by effect type
//   - healingModifier: Multiplier affecting all healing received (1.0 = normal healing)
//
// Related types:
//   - Effect: Represents a single effect instance
//   - Stats: Contains all modifiable entity statistics
//   - EffectType: Enumeration of possible effect types
//   - ImmunityData: Contains immunity duration and source information
type EffectManager struct {
	activeEffects   map[string]*Effect
	baseStats       *Stats
	currentStats    *Stats
	immunities      map[EffectType]*ImmunityData
	tempImmunities  map[EffectType]*ImmunityData
	resistances     map[EffectType]float64
	healingModifier float64
	mu              sync.RWMutex
}

// NewEffectManager creates and initializes a new EffectManager instance.
//
// Parameters:
//   - baseStats: A pointer to Stats representing the base statistics that will be modified by effects.
//     Must not be nil as it serves as the foundation for all stat calculations.
//
// Returns:
//   - *EffectManager: A new EffectManager instance with initialized maps for active effects,
//     immunities, temporary immunities, and resistances. The current stats are initialized
//     as a clone of the base stats.
//
// Related types:
//   - Stats: Base statistical values
//   - Effect: Individual effect instances
//   - EffectType: Types of effects that can be applied
//   - ImmunityData: Immunity information for effect types
//
// Note:
//   - Automatically initializes default immunities via initializeDefaultImmunities()
//   - All maps are initialized as empty but non-nil
func NewEffectManager(baseStats *Stats) *EffectManager {
	em := &EffectManager{
		activeEffects:  make(map[string]*Effect),
		baseStats:      baseStats,
		currentStats:   baseStats.Clone(),
		immunities:     make(map[EffectType]*ImmunityData),
		tempImmunities: make(map[EffectType]*ImmunityData),
		resistances:    make(map[EffectType]float64),
	}
	em.initializeDefaultImmunities()
	return em
}

// Effect creation helpers

// NewEffect creates a new Effect instance with the specified type, duration and magnitude.
//
// Parameters:
//   - effectType: The type of effect to create (EffectType)
//   - duration: How long the effect lasts (Duration struct with Rounds, Turns, RealTime)
//   - magnitude: The strength/amount of the effect (float64)
//
// Returns:
//   - *Effect: A pointer to the newly created Effect instance with default values
//
// The effect is initialized with:
//   - A new unique ID
//   - Active status
//   - 1 stack
//   - Default dispel info (lowest priority, not removable)
//   - Empty slices for tags and modifiers
//   - Current time as start time
//   - Empty strings for name, description and other string fields
//
// Related types:
//   - Effect struct
//   - EffectType type
//   - Duration struct
//   - DispelInfo struct
func NewEffect(effectType EffectType, duration Duration, magnitude float64) *Effect {
	return &Effect{
		ID:          NewUID(),
		Type:        effectType,
		Name:        "",
		Description: "",
		StartTime:   time.Now(),
		Duration:    duration,
		TickRate: Duration{
			Rounds:   0,
			Turns:    0,
			RealTime: 0,
		},
		Magnitude:    magnitude,
		DamageType:   "",
		SourceID:     "",
		SourceType:   "",
		TargetID:     "",
		StatAffected: "",
		IsActive:     true,
		Stacks:       1,
		Tags:         []string{},
		DispelInfo: DispelInfo{
			Priority:  DispelPriorityLowest,
			Types:     []DispelType{},
			Removable: false,
		},
		Modifiers: []Modifier{},
	}
}

// CreateDamageEffect creates a new damage-dealing Effect with the specified parameters.
//
// Parameters:
//   - effectType: The type of effect being created (e.g. poison, bleed, etc)
//   - damageType: The type of damage this effect deals (e.g. physical, magic, etc)
//   - damage: Amount of damage dealt per tick (must be >= 0)
//   - duration: How long the effect lasts in real time
//
// Returns:
//
//	A new *Effect configured to deal periodic damage of the specified type
//
// The effect will tick once per second (defined in TickRate).
// Related types:
//   - Effect
//   - EffectType
//   - DamageType
//   - Duration
func CreateDamageEffect(effectType EffectType, damageType DamageType, damage float64, duration time.Duration) *Effect {
	effect := NewEffect(effectType, Duration{
		Rounds:   0,
		Turns:    0,
		RealTime: duration,
	}, damage)
	effect.DamageType = damageType
	effect.TickRate = Duration{
		Rounds:   0,
		Turns:    0,
		RealTime: time.Second,
	}
	return effect
}

// IsExpired checks if the effect has expired based on either real time duration or number of rounds.
//
// Parameters:
//   - currentTime time.Time: The current time to check against the effect's start time
//
// Returns:
//   - bool: true if the effect has expired, false otherwise
//
// Notes:
// - For real-time based effects (Duration.RealTime > 0), checks if currentTime is after startTime + duration
// - For round-based effects (Duration.Rounds > 0), currently returns false (TODO: implementation needed)
// - If neither duration type is set, effect never expires (returns false)
//
// Related:
// - Duration struct containing RealTime and Rounds fields
// - Effect struct containing StartTime and Duration fields
func (e *Effect) IsExpired(currentTime time.Time) bool {
	if e.Duration.RealTime > 0 {
		return currentTime.After(e.StartTime.Add(e.Duration.RealTime))
	}
	if e.Duration.Rounds > 0 {
		// Handle round-based expiration
		return false // TODO: Implement round-based expiration
	}
	if e.Duration.Turns > 0 {
		// Handle turn-based expiration
		return false // TODO: Implement turn-based expiration
	}

	// Negative durations are permanent effects (never expire)
	if e.Duration.RealTime < 0 || e.Duration.Rounds < 0 || e.Duration.Turns < 0 {
		return false
	}

	// Zero duration = instant effect (expires immediately)
	if e.Duration.RealTime == 0 && e.Duration.Rounds == 0 && e.Duration.Turns == 0 {
		return true
	}

	return false
}

// ShouldTick determines if the effect should trigger based on its tick rate.
// It checks if enough real time has elapsed since the effect started for the next tick to occur.
//
// Parameters:
//   - currentTime time.Time: The current timestamp to check against
//
// Returns:
//   - bool: true if the effect should tick, false otherwise
//
// Edge cases:
//   - Returns false if TickRate.RealTime is 0 to prevent infinite ticking
//   - Uses modulo operation to determine regular intervals based on TickRate.RealTime
//
// Related:
//   - Effect.StartTime field
//   - Effect.TickRate struct
func (e *Effect) ShouldTick(currentTime time.Time) bool {
	if e.TickRate.RealTime == 0 {
		return false
	}
	timeSinceStart := currentTime.Sub(e.StartTime)
	return timeSinceStart%e.TickRate.RealTime == 0
}

// EffectTyper is an interface that defines a contract for types that have an associated effect type.
// It provides a common way to identify and categorize different types of effects in the game.
//
// Returns:
//   - EffectType: The type classification of the effect
//
// EffectTyper interface is defined in types.go
// Related types:
//   - EffectType: The enumeration of possible effect types

// GetEffectType returns the type of the Effect.
//
// Returns:
//   - EffectType: The type classification of this effect.
//
// Related types:
//   - EffectType: An enumeration defining different effect categories
//   - Effect: The parent struct containing effect data
func (e *Effect) GetEffectType() EffectType {
	return e.Type
}

// GetEffectType returns the type of this DamageEffect
//
// Returns:
//   - EffectType: The type of effect this DamageEffect represents
//
// Related:
//   - EffectType interface
//   - Effect.Type field
func (de *DamageEffect) GetEffectType() EffectType {
	return de.Effect.Type
}

// ToEffect converts a DamageEffect to an Effect by returning the underlying Effect field.
// This method allows DamageEffect to be used as an Effect type.
//
// Returns:
//   - *Effect: The underlying Effect pointer contained in the DamageEffect struct
//
// Related Types:
//   - Effect
//   - DamageEffect
func (de *DamageEffect) ToEffect() *Effect {
	return de.Effect
}

// ToDamageEffect attempts to convert a generic Effect to a DamageEffect.
//
// Parameters:
//   - e *Effect: The effect to convert. Must not be nil.
//
// Returns:
//   - *DamageEffect: The converted damage effect if successful, nil otherwise
//   - bool: true if conversion was successful, false if effect type is not convertible
//
// The function only converts poison, burning and bleeding effect types.
// All other effect types will return nil and false.
//
// Related types:
//   - Effect
//   - DamageEffect
//   - EffectType (EffectPoison, EffectBurning, EffectBleeding)
func ToDamageEffect(e *Effect) (*DamageEffect, bool) {
	switch e.Type {
	case EffectPoison, EffectBurning, EffectBleeding:
		return &DamageEffect{
			Effect:         e,
			DamageType:     e.DamageType,
			BaseDamage:     e.Magnitude,
			DamageScale:    0,
			PenetrationPct: 0,
		}, true
	default:
		return nil, false
	}
}
