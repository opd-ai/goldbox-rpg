package game

import (
	"fmt"
	"sync"
	"time"
)

// EffectHolder represents an entity that can have effects applied
type EffectHolder interface {
	// Effect management
	AddEffect(effect *Effect) error
	RemoveEffect(effectID string) error
	HasEffect(effectType EffectType) bool
	GetEffects() []*Effect

	// Stats that can be modified by effects
	GetStats() *Stats
	SetStats(*Stats)

	// Base stats before effects
	GetBaseStats() *Stats
}

// Stats represents an entity's modifiable attributes
type Stats struct {
	Health       float64
	Mana         float64
	Strength     float64
	Dexterity    float64
	Intelligence float64
	// Add other stats as needed

	// Calculated stats
	MaxHealth float64
	MaxMana   float64
	Defense   float64
	Speed     float64
}

// EffectManager handles effect application and management for an entity
type EffectManager struct {
	activeEffects map[string]*Effect     // All current effects
	baseStats     *Stats                 // Original stats
	currentStats  *Stats                 // Stats after effects
	immunities    []EffectType           // Effect immunities
	resistances   map[EffectType]float64 // Effect resistance percentages
	mu            sync.RWMutex
}

// NewEffectManager creates a new effect manager for an entity
func NewEffectManager(baseStats *Stats) *EffectManager {
	return &EffectManager{
		activeEffects: make(map[string]*Effect),
		baseStats:     baseStats,
		currentStats:  baseStats.Clone(), // Implement Clone method for Stats
		resistances:   make(map[EffectType]float64),
	}
}

// ApplyEffect adds and applies an effect to the entity
func (em *EffectManager) ApplyEffect(effect *Effect) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Check immunities
	for _, immunity := range em.immunities {
		if immunity == effect.Type {
			return fmt.Errorf("entity is immune to effect type: %s", effect.Type)
		}
	}

	// Apply resistance if any
	if resistance, exists := em.resistances[effect.Type]; exists {
		effect.Magnitude *= (1 - resistance)
	}

	// Check for existing effect of same type
	for _, existing := range em.activeEffects {
		if existing.Type == effect.Type {
			switch {
			case existing.Type.AllowsStacking():
				existing.AddStack()
				return nil
			case effect.Magnitude > existing.Magnitude:
				// Replace if new effect is stronger
				delete(em.activeEffects, existing.ID)
				break
			default:
				return fmt.Errorf("cannot apply weaker effect of same type")
			}
		}
	}

	// Add new effect
	effect.StartTime = time.Now()
	effect.IsActive = true
	em.activeEffects[effect.ID] = effect

	// Recalculate stats
	em.recalculateStats()

	return nil
}

// RemoveEffect removes an effect by ID
func (em *EffectManager) RemoveEffect(effectID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if effect, exists := em.activeEffects[effectID]; exists {
		effect.IsActive = false
		delete(em.activeEffects, effectID)
		em.recalculateStats()
		return nil
	}
	return fmt.Errorf("effect not found: %s", effectID)
}

// UpdateEffects processes all active effects
func (em *EffectManager) UpdateEffects(currentTime time.Time) {
	em.mu.Lock()
	defer em.mu.Unlock()

	needsRecalc := false

	for id, effect := range em.activeEffects {
		// Check expiration
		if effect.IsExpired(currentTime) {
			delete(em.activeEffects, id)
			needsRecalc = true
			continue
		}

		// Process periodic effects
		if effect.ShouldTick(currentTime) {
			em.processEffectTick(effect)
		}
	}

	if needsRecalc {
		em.recalculateStats()
	}
}

// recalculateStats applies all active effects to base stats
func (em *EffectManager) recalculateStats() {
	// Start with base stats
	newStats := em.baseStats.Clone()

	// First pass: collect all modifiers
	addMods := make(map[string]float64)
	multMods := make(map[string]float64)
	setMods := make(map[string]float64)

	for _, effect := range em.activeEffects {
		magnitude := effect.Magnitude * float64(effect.Stacks)

		for _, mod := range effect.Modifiers {
			switch mod.Operation {
			case ModAdd:
				addMods[mod.Stat] += mod.Value * magnitude
			case ModMultiply:
				multMods[mod.Stat] = (multMods[mod.Stat] + 1) * (mod.Value * magnitude)
			case ModSet:
				if current, exists := setMods[mod.Stat]; !exists || mod.Value > current {
					setMods[mod.Stat] = mod.Value * magnitude
				}
			}
		}
	}

	// Apply modifications in order: add -> multiply -> set
	em.applyStatModifiers(newStats, addMods, multMods, setMods)

	em.currentStats = newStats
}

// processEffectTick handles periodic effect updates
func (em *EffectManager) processEffectTick(effect *Effect) {
	switch effect.Type {
	case EffectDamageOverTime:
		em.currentStats.Health -= effect.Magnitude * float64(effect.Stacks)
	case EffectHealOverTime:
		em.currentStats.Health = min(
			em.currentStats.Health+effect.Magnitude*float64(effect.Stacks),
			em.currentStats.MaxHealth,
		)
		// Add other effect type processing
	}
}

// Helper methods

func (em *EffectManager) applyStatModifiers(stats *Stats, addMods, multMods, setMods map[string]float64) {
	// Helper function to apply mods to a stat
	applyStat := func(current *float64, statName string) {
		if add, ok := addMods[statName]; ok {
			*current += add
		}
		if mult, ok := multMods[statName]; ok {
			*current *= mult
		}
		if set, ok := setMods[statName]; ok {
			*current = set
		}
	}

	// Apply to each stat
	applyStat(&stats.Health, "health")
	applyStat(&stats.Mana, "mana")
	applyStat(&stats.Strength, "strength")
	applyStat(&stats.Dexterity, "dexterity")
	applyStat(&stats.Intelligence, "intelligence")
	// Apply to other stats
}

// Stats Clone method
func (s *Stats) Clone() *Stats {
	return &Stats{
		Health:       s.Health,
		Mana:         s.Mana,
		Strength:     s.Strength,
		Dexterity:    s.Dexterity,
		Intelligence: s.Intelligence,
		MaxHealth:    s.MaxHealth,
		MaxMana:      s.MaxMana,
		Defense:      s.Defense,
		Speed:        s.Speed,
	}
}

// Helper function for min value
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Method to check if effect type allows stacking
func (et EffectType) AllowsStacking() bool {
	switch et {
	case EffectDamageOverTime, EffectHealOverTime, EffectStatBoost:
		return true
	default:
		return false
	}
}
