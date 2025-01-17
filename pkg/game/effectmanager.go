package game

import (
	"fmt"
	"time"
)

// EffectHolder represents an entity that can have effects applied
// EffectHolder defines an interface for entities that can have effects applied to them.
// An effect holder maintains both current stats (which include effect modifications)
// and base stats (original values before effects).
//
// Implementations must handle:
// - Effect management (add/remove/query)
// - Current stats that can be modified by effects
// - Base stats that represent original unmodified values
//
// Related types:
// - Effect: Represents a single effect that can be applied
// - Stats: Contains the actual stat values
// - EffectType: Enumeration of possible effect types
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
// Stats represents a character's base and derived statistics in the game.
// It contains both primary attributes that can be directly modified
// and secondary (calculated) attributes that are derived from the primary ones.
//
// Primary attributes:
//   - Health: Current health points
//   - Mana: Current mana points
//   - Strength: Physical power and carrying capacity
//   - Dexterity: Agility and precision
//   - Intelligence: Mental capability and magical aptitude
//
// Calculated attributes:
//   - MaxHealth: Maximum possible health points
//   - MaxMana: Maximum possible mana points
//   - Defense: Damage reduction capability
//   - Speed: Movement and action speed
//
// The Stats struct is used throughout the game systems including:
// - Combat calculations
// - Character progression
// - Status effect application
// - Equipment bonuses
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

// NewDefaultStats creates and returns a new Stats structure initialized with default values.
// It sets baseline stats that are commonly used as a starting point for new game entities.
//
// Returns:
//   - *Stats: A pointer to a new Stats structure with the following default values:
//     Health: 100, Mana: 100, Strength: 10, Dexterity: 10, Intelligence: 10,
//     MaxHealth: 100, MaxMana: 100, Defense: 10, Speed: 10
//
// Related types:
//   - Stats struct: The base structure containing all stat fields
func NewDefaultStats() *Stats {
	return &Stats{
		Health:       100,
		Mana:         100,
		Strength:     10,
		Dexterity:    10,
		Intelligence: 10,
		MaxHealth:    100,
		MaxMana:      100,
		Defense:      10,
		Speed:        10,
	}
}

// RemoveEffect removes an effect by ID
// RemoveEffect deactivates and removes an effect from the active effects list by its ID.
//
// Parameters:
//   - effectID: string - The unique identifier of the effect to remove
//
// Returns:
//   - error: Returns nil if effect was successfully removed, or an error if effect was not found
//
// Notable behavior:
// - Locks the EffectManager mutex during operation to ensure thread safety
// - Sets effect's IsActive flag to false before removal
// - Triggers recalculation of stats after removing the effect
// - Returns error if effect ID does not exist in activeEffects map
//
// Related:
// - recalculateStats() - Called after effect removal to update stats
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
// UpdateEffects processes and maintains active effects based on the current time.
// It handles effect expiration, periodic effect ticks, and stat recalculation.
//
// Parameters:
//   - currentTime time.Time: The current game time to check effects against
//
// The method performs the following:
// - Removes expired effects from activeEffects
// - Triggers periodic effect ticks when appropriate
// - Recalculates stats if any effects were removed
//
// Thread-safety: Uses mutex locking to safely modify shared state
//
// Related:
// - EffectManager.processEffectTick()
// - EffectManager.recalculateStats()
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
// recalculateStats recalculates entity stats by applying all active effect modifiers.
// It processes effects in the following order:
// 1. Collects all additive, multiplicative and set modifiers from active effects
// 2. Applies modifiers in order: additive -> multiplicative -> set
//
// The method updates em.currentStats with the newly calculated stats.
// Base stats are preserved in em.baseStats.
//
// Related types:
// - ModOperation (pkg/game/effect.go)
// - Stats (pkg/game/stats.go)
//
// Note: Effect magnitudes are multiplied by stack count when applying modifiers.
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

func (em *EffectManager) applyEffectInternal(effect *Effect) error {
	// Check for existing effect of same type
	for _, existing := range em.activeEffects {
		if existing.Type == effect.Type {
			switch {
			case effect.Type.AllowsStacking():
				existing.Stacks++
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
