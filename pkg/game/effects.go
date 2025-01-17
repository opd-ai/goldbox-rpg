// pkg/game/effect.go
package game

import (
	"sync"
	"time"
)

// Core types
type (
	EffectType     string
	DamageType     string
	DispelType     string
	ImmunityType   int
	DispelPriority int
)

// Constants
const (
	// Effect Types
	EffectDamageOverTime EffectType = "damage_over_time"
	EffectHealOverTime   EffectType = "heal_over_time"
	EffectPoison         EffectType = "poison"
	EffectBurning        EffectType = "burning"
	EffectBleeding       EffectType = "bleeding"
	EffectStun           EffectType = "stun"
	EffectRoot           EffectType = "root"
	EffectStatBoost      EffectType = "stat_boost"
	EffectStatPenalty    EffectType = "stat_penalty"

	// Damage Types
	DamagePhysical  DamageType = "physical"
	DamageFire      DamageType = "fire"
	DamagePoison    DamageType = "poison"
	DamageFrost     DamageType = "frost"
	DamageLightning DamageType = "lightning"

	// Dispel Types
	DispelMagic   DispelType = "magic"
	DispelCurse   DispelType = "curse"
	DispelPoison  DispelType = "poison"
	DispelDisease DispelType = "disease"
	DispelAll     DispelType = "all"

	// Immunity Types
	ImmunityNone ImmunityType = iota
	ImmunityPartial
	ImmunityComplete
	ImmunityReflect

	// Dispel Priorities
	DispelPriorityLowest  DispelPriority = 0
	DispelPriorityLow     DispelPriority = 25
	DispelPriorityNormal  DispelPriority = 50
	DispelPriorityHigh    DispelPriority = 75
	DispelPriorityHighest DispelPriority = 100
)

// Duration represents a game time duration
type Duration struct {
	Rounds   int           `yaml:"duration_rounds"`
	Turns    int           `yaml:"duration_turns"`
	RealTime time.Duration `yaml:"duration_real"`
}

// Effect represents a game effect
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

	IsActive bool     `yaml:"effect_active"`
	Stacks   int      `yaml:"effect_stacks"`
	Tags     []string `yaml:"effect_tags"`

	DispelInfo DispelInfo `yaml:"dispel_info"`
	Modifiers  []Modifier `yaml:"effect_modifiers"`
}

type Modifier struct {
	Stat      string    `yaml:"mod_stat"`
	Value     float64   `yaml:"mod_value"`
	Operation ModOpType `yaml:"mod_operation"`
}

type ModOpType string

const (
	ModAdd      ModOpType = "add"
	ModMultiply ModOpType = "multiply"
	ModSet      ModOpType = "set"
)

// DispelInfo contains metadata about effect dispelling
type DispelInfo struct {
	Priority  DispelPriority `yaml:"dispel_priority"`
	Types     []DispelType   `yaml:"dispel_types"`
	Removable bool           `yaml:"dispel_removable"`
}

// ImmunityData represents immunity information
type ImmunityData struct {
	Type       ImmunityType
	Duration   time.Duration
	Resistance float64
	ExpiresAt  time.Time
}

// EffectManager handles effect application and management
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

// NewEffectManager creates a new effect manager
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

func CreateDamageEffect(effectType EffectType, damageType DamageType, damage float64, duration time.Duration) *Effect {
	effect := NewEffect(effectType, Duration{RealTime: duration}, damage)
	effect.DamageType = damageType
	effect.TickRate = Duration{RealTime: time.Second}
	return effect
}

// Add to Effect type in effects.go
func (e *Effect) IsExpired(currentTime time.Time) bool {
	if e.Duration.RealTime > 0 {
		return currentTime.After(e.StartTime.Add(e.Duration.RealTime))
	}
	if e.Duration.Rounds > 0 {
		// Handle round-based expiration
		return false // TODO: Implement round-based expiration
	}
	return false
}

func (e *Effect) ShouldTick(currentTime time.Time) bool {
	if e.TickRate.RealTime == 0 {
		return false
	}
	timeSinceStart := currentTime.Sub(e.StartTime)
	return timeSinceStart%e.TickRate.RealTime == 0
}

// EffectTyper interface for getting effect type
type EffectTyper interface {
	GetEffectType() EffectType
}

// Implement EffectTyper for Effect
func (e *Effect) GetEffectType() EffectType {
	return e.Type
}

// Implement EffectTyper for DamageEffect
func (de *DamageEffect) GetEffectType() EffectType {
	return de.Effect.Type
}

// Helper method to convert DamageEffect to Effect
func (de *DamageEffect) ToEffect() *Effect {
	return de.Effect
}

// Helper method to check and convert Effect to DamageEffect
func ToDamageEffect(e *Effect) (*DamageEffect, bool) {
	switch e.Type {
	case EffectPoison, EffectBurning, EffectBleeding:
		return &DamageEffect{
			Effect:     e,
			DamageType: e.DamageType,
			BaseDamage: e.Magnitude,
		}, true
	default:
		return nil, false
	}
}
