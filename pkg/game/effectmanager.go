package game

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	// Configure structured logging with caller context
	logrus.SetReportCaller(true)

	logrus.WithFields(logrus.Fields{
		"function": "init",
		"package":  "game",
		"file":     "effectmanager.go",
	}).Debug("package initialized - structured logging configured with caller context")
}

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
// EffectHolder interface is defined in types.go
// - Stats: Contains the actual stat values
// - EffectType: Enumeration of possible effect types

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
// NewDefaultStats creates a new Stats instance with sensible default values for a typical game entity.
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
	logrus.WithFields(logrus.Fields{
		"function": "NewDefaultStats",
		"package":  "game",
	}).Debug("entering NewDefaultStats")

	stats := &Stats{
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

	logrus.WithFields(logrus.Fields{
		"function":   "NewDefaultStats",
		"package":    "game",
		"health":     stats.Health,
		"mana":       stats.Mana,
		"strength":   stats.Strength,
		"dexterity":  stats.Dexterity,
		"max_health": stats.MaxHealth,
		"max_mana":   stats.MaxMana,
	}).Debug("created default stats with baseline values")

	logrus.WithFields(logrus.Fields{
		"function": "NewDefaultStats",
		"package":  "game",
	}).Debug("exiting NewDefaultStats")

	return stats
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
	logrus.WithFields(logrus.Fields{
		"function":  "RemoveEffect",
		"package":   "game",
		"effect_id": effectID,
	}).Debug("entering RemoveEffect")

	em.mu.Lock()
	defer em.mu.Unlock()

	if effect, exists := em.activeEffects[effectID]; exists {
		logrus.WithFields(logrus.Fields{
			"function":    "RemoveEffect",
			"package":     "game",
			"effect_id":   effectID,
			"effect_type": effect.Type,
			"was_active":  effect.IsActive,
		}).Debug("effect found - deactivating and removing")

		effect.IsActive = false
		delete(em.activeEffects, effectID)

		logrus.WithFields(logrus.Fields{
			"function":  "RemoveEffect",
			"package":   "game",
			"effect_id": effectID,
		}).Debug("effect removed - triggering stat recalculation")

		em.recalculateStats()

		logrus.WithFields(logrus.Fields{
			"function":  "RemoveEffect",
			"package":   "game",
			"effect_id": effectID,
		}).Debug("exiting RemoveEffect - success")

		return nil
	}

	logrus.WithFields(logrus.Fields{
		"function":  "RemoveEffect",
		"package":   "game",
		"effect_id": effectID,
	}).Warn("effect not found for removal")

	logrus.WithFields(logrus.Fields{
		"function":  "RemoveEffect",
		"package":   "game",
		"effect_id": effectID,
	}).Debug("exiting RemoveEffect - error")

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
	logrus.WithFields(logrus.Fields{
		"function":     "UpdateEffects",
		"package":      "game",
		"current_time": currentTime,
		"active_count": len(em.activeEffects),
	}).Debug("entering UpdateEffects")

	em.mu.Lock()
	defer em.mu.Unlock()

	needsRecalc := false
	expiredCount := 0
	tickedCount := 0

	logrus.WithFields(logrus.Fields{
		"function":     "UpdateEffects",
		"package":      "game",
		"active_count": len(em.activeEffects),
	}).Debug("processing active effects for expiration and ticks")

	for id, effect := range em.activeEffects {
		// Check expiration
		if effect.IsExpired(currentTime) {
			logrus.WithFields(logrus.Fields{
				"function":    "UpdateEffects",
				"package":     "game",
				"effect_id":   id,
				"effect_type": effect.Type,
				"expired_at":  currentTime,
			}).Debug("effect expired - removing")

			delete(em.activeEffects, id)
			needsRecalc = true
			expiredCount++
			continue
		}

		// Process periodic effects
		if effect.ShouldTick(currentTime) {
			logrus.WithFields(logrus.Fields{
				"function":    "UpdateEffects",
				"package":     "game",
				"effect_id":   id,
				"effect_type": effect.Type,
			}).Debug("processing effect tick")

			em.processEffectTick(effect)
			tickedCount++
		}
	}

	logrus.WithFields(logrus.Fields{
		"function":      "UpdateEffects",
		"package":       "game",
		"expired_count": expiredCount,
		"ticked_count":  tickedCount,
		"needs_recalc":  needsRecalc,
	}).Debug("effect processing completed")

	if needsRecalc {
		logrus.WithFields(logrus.Fields{
			"function": "UpdateEffects",
			"package":  "game",
		}).Debug("triggering stat recalculation due to effect changes")

		em.recalculateStats()
	}

	logrus.WithFields(logrus.Fields{
		"function": "UpdateEffects",
		"package":  "game",
	}).Debug("exiting UpdateEffects")
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
	logrus.WithFields(logrus.Fields{
		"function":       "recalculateStats",
		"package":        "game",
		"active_effects": len(em.activeEffects),
	}).Debug("entering recalculateStats")

	// Start with base stats
	newStats := em.baseStats.Clone()

	logrus.WithFields(logrus.Fields{
		"function":    "recalculateStats",
		"package":     "game",
		"base_health": newStats.Health,
		"base_mana":   newStats.Mana,
	}).Debug("starting with cloned base stats")

	// First pass: collect all modifiers
	addMods := make(map[string]float64)
	multMods := make(map[string]float64)
	setMods := make(map[string]float64)

	effectsProcessed := 0
	for _, effect := range em.activeEffects {
		magnitude := effect.Magnitude * float64(effect.Stacks)

		logrus.WithFields(logrus.Fields{
			"function":  "recalculateStats",
			"package":   "game",
			"effect_id": effect.ID,
			"stacks":    effect.Stacks,
			"magnitude": magnitude,
			"modifiers": len(effect.Modifiers),
		}).Debug("processing effect modifiers")

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
		effectsProcessed++
	}

	logrus.WithFields(logrus.Fields{
		"function":            "recalculateStats",
		"package":             "game",
		"effects_processed":   effectsProcessed,
		"additive_mods":       len(addMods),
		"multiplicative_mods": len(multMods),
		"set_mods":            len(setMods),
	}).Debug("modifier collection completed")

	// Apply modifications in order: add -> multiply -> set
	logrus.WithFields(logrus.Fields{
		"function": "recalculateStats",
		"package":  "game",
	}).Debug("applying stat modifiers")

	em.applyStatModifiers(newStats, addMods, multMods, setMods)

	oldStats := em.currentStats
	em.currentStats = newStats

	logrus.WithFields(logrus.Fields{
		"function":   "recalculateStats",
		"package":    "game",
		"old_health": oldStats.Health,
		"new_health": newStats.Health,
		"old_mana":   oldStats.Mana,
		"new_mana":   newStats.Mana,
	}).Info("stats recalculated successfully")

	logrus.WithFields(logrus.Fields{
		"function": "recalculateStats",
		"package":  "game",
	}).Debug("exiting recalculateStats")
}

// Helper methods

// applyStatModifiers applies additive, multiplicative and set modifiers to a Stats object's attributes.
//
// Parameters:
//   - stats: *Stats - Pointer to the Stats object to be modified
//   - addMods: map[string]float64 - Map of stat names to values to be added
//   - multMods: map[string]float64 - Map of stat names to multiplication factors
//   - setMods: map[string]float64 - Map of stat names to values to directly set
//
// The function processes modifiers in order: additive -> multiplicative -> set.
// Stats that don't have corresponding modifiers remain unchanged.
// Stats names must match the lowercase string keys: "health", "mana", "strength", etc.
//
// Related types:
//   - Stats struct containing the modifiable attributes
func (em *EffectManager) applyStatModifiers(stats *Stats, addMods, multMods, setMods map[string]float64) {
	logrus.WithFields(logrus.Fields{
		"function": "applyStatModifiers",
		"package":  "game",
	}).Debug("entering applyStatModifiers")

	// Helper function to apply mods to a stat
	applyStat := func(current *float64, statName string) {
		oldValue := *current

		if add, ok := addMods[statName]; ok {
			*current += add
			logrus.WithFields(logrus.Fields{
				"function":  "applyStatModifiers",
				"package":   "game",
				"stat":      statName,
				"operation": "add",
				"modifier":  add,
				"old_value": oldValue,
				"new_value": *current,
			}).Debug("applied additive modifier")
		}
		if mult, ok := multMods[statName]; ok {
			oldValue := *current
			*current *= mult
			logrus.WithFields(logrus.Fields{
				"function":  "applyStatModifiers",
				"package":   "game",
				"stat":      statName,
				"operation": "multiply",
				"modifier":  mult,
				"old_value": oldValue,
				"new_value": *current,
			}).Debug("applied multiplicative modifier")
		}
		if set, ok := setMods[statName]; ok {
			oldValue := *current
			*current = set
			logrus.WithFields(logrus.Fields{
				"function":  "applyStatModifiers",
				"package":   "game",
				"stat":      statName,
				"operation": "set",
				"modifier":  set,
				"old_value": oldValue,
				"new_value": *current,
			}).Debug("applied set modifier")
		}
	}

	// Apply to each stat
	applyStat(&stats.Health, "health")
	applyStat(&stats.Mana, "mana")
	applyStat(&stats.Strength, "strength")
	applyStat(&stats.Dexterity, "dexterity")
	applyStat(&stats.Intelligence, "intelligence")
	// Apply to other stats

	logrus.WithFields(logrus.Fields{
		"function": "applyStatModifiers",
		"package":  "game",
	}).Debug("function exit - stat modifiers applied successfully")
}

// Clone creates and returns a deep copy of a Stats object
// Clone duplicates all stat values into a new Stats instance.
//
// Returns:
//   - *Stats: A new Stats instance with identical values to the original
//
// Notable behavior:
// - Creates a completely independent copy of the Stats object
// - All fields are copied by value since they are primitive types
//
// Related types:
// - Stats struct: The base structure containing all stat fields
// - NewDefaultStats(): Factory method for creating Stats objects
func (s *Stats) Clone() *Stats {
	logrus.WithFields(logrus.Fields{
		"function": "Clone",
		"package":  "game",
		"type":     "Stats",
	}).Debug("function entry - cloning stats object")

	clone := &Stats{
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

	logrus.WithFields(logrus.Fields{
		"function":     "Clone",
		"package":      "game",
		"type":         "Stats",
		"health":       clone.Health,
		"mana":         clone.Mana,
		"strength":     clone.Strength,
		"dexterity":    clone.Dexterity,
		"intelligence": clone.Intelligence,
		"max_health":   clone.MaxHealth,
		"max_mana":     clone.MaxMana,
		"defense":      clone.Defense,
		"speed":        clone.Speed,
	}).Debug("function exit - stats object cloned successfully")

	return clone
}

// Helper function for min value
// min returns the smaller of two float64 numbers.
//
// Parameters:
//   - a: first float64 number to compare
//   - b: second float64 number to compare
//
// Returns:
//   - float64: the smaller of a and b
//
// minFloat function is now defined in utils.go

// AllowsStacking determines whether effects of this type can stack with each other.
// This method controls which effect types can have multiple instances active at once
// on the same target.
//
// Returns:
//   - true if effects of this type can stack (EffectDamageOverTime, EffectHealOverTime, EffectStatBoost)
//   - false for all other effect types
//
// Related types:
//   - EffectType: The enum type this method belongs to
//   - Effect: The main effect struct that uses this stacking behavior
func (et EffectType) AllowsStacking() bool {
	logrus.WithFields(logrus.Fields{
		"function":    "AllowsStacking",
		"package":     "game",
		"type":        "EffectType",
		"effect_type": et,
	}).Debug("function entry - checking effect stacking behavior")

	var allowsStacking bool
	switch et {
	case EffectDamageOverTime, EffectHealOverTime, EffectStatBoost:
		allowsStacking = true
	default:
		allowsStacking = false
	}

	logrus.WithFields(logrus.Fields{
		"function":        "AllowsStacking",
		"package":         "game",
		"type":            "EffectType",
		"effect_type":     et,
		"allows_stacking": allowsStacking,
	}).Debug("function exit - stacking behavior determined")

	return allowsStacking
}

// applyEffectInternal applies an effect to an entity's active effects list, handling stacking
// and magnitude-based replacement of existing effects.
//
// Parameters:
//   - effect: *Effect - The effect to be applied. Must not be nil.
//
// Returns:
//   - error: Returns nil on successful application, or an error if:
//   - A weaker non-stacking effect is applied when a stronger one exists
//   - The effect parameter is nil
//
// Behavior:
//   - For stackable effects: Increments stack count on existing effect
//   - For non-stackable effects: Replaces existing if new effect is stronger
//   - For new effect types: Adds to active effects list
//   - Recalculates stats after any changes
//
// Related:
//   - Effect.Type.AllowsStacking()
//   - EffectManager.recalculateStats()
func (em *EffectManager) applyEffectInternal(effect *Effect) error {
	logrus.WithFields(logrus.Fields{
		"function":    "applyEffectInternal",
		"package":     "game",
		"effect_id":   effect.ID,
		"effect_type": effect.Type,
		"magnitude":   effect.Magnitude,
		"duration":    effect.Duration,
	}).Debug("function entry - applying effect internally")

	em.mu.Lock()
	defer em.mu.Unlock()

	// Check for existing effect of same type
	effectReplaced := false
	for _, existing := range em.activeEffects {
		if existing.Type == effect.Type {
			logrus.WithFields(logrus.Fields{
				"function":           "applyEffectInternal",
				"package":            "game",
				"effect_id":          effect.ID,
				"existing_effect_id": existing.ID,
				"effect_type":        effect.Type,
				"allows_stacking":    effect.Type.AllowsStacking(),
			}).Debug("found existing effect of same type")

			switch {
			case effect.Type.AllowsStacking():
				existing.Stacks++
				logrus.WithFields(logrus.Fields{
					"function":    "applyEffectInternal",
					"package":     "game",
					"effect_id":   existing.ID,
					"effect_type": effect.Type,
					"new_stacks":  existing.Stacks,
				}).Debug("stacked effect on existing instance")
				return nil
			case effect.Magnitude > existing.Magnitude:
				// Replace if new effect is stronger
				delete(em.activeEffects, existing.ID)
				effectReplaced = true
				logrus.WithFields(logrus.Fields{
					"function":      "applyEffectInternal",
					"package":       "game",
					"old_effect_id": existing.ID,
					"new_effect_id": effect.ID,
					"old_magnitude": existing.Magnitude,
					"new_magnitude": effect.Magnitude,
				}).Debug("replaced weaker effect with stronger one")
			default:
				logrus.WithFields(logrus.Fields{
					"function":           "applyEffectInternal",
					"package":            "game",
					"effect_id":          effect.ID,
					"existing_magnitude": existing.Magnitude,
					"new_magnitude":      effect.Magnitude,
				}).Warn("attempted to apply weaker effect - rejected")
				return fmt.Errorf("cannot apply weaker effect of same type")
			}
		}
	}

	// Add new effect
	effect.StartTime = time.Now()
	effect.IsActive = true
	em.activeEffects[effect.ID] = effect

	logrus.WithFields(logrus.Fields{
		"function":     "applyEffectInternal",
		"package":      "game",
		"effect_id":    effect.ID,
		"effect_type":  effect.Type,
		"start_time":   effect.StartTime,
		"replaced":     effectReplaced,
		"total_active": len(em.activeEffects),
	}).Debug("added new effect to active effects")

	// Recalculate stats
	em.recalculateStats()

	logrus.WithFields(logrus.Fields{
		"function":  "applyEffectInternal",
		"package":   "game",
		"effect_id": effect.ID,
	}).Debug("function exit - effect applied successfully")

	return nil
}

// EffectHolder interface implementation

// HasEffect checks if the entity has an active effect of the specified type
func (em *EffectManager) HasEffect(effectType EffectType) bool {
	logrus.WithFields(logrus.Fields{
		"function":    "HasEffect",
		"package":     "game",
		"effect_type": effectType,
	}).Debug("function entry - checking for active effect")

	em.mu.RLock()
	defer em.mu.RUnlock()

	effectCount := 0
	for _, effect := range em.activeEffects {
		if effect.Type == effectType && effect.IsActive {
			effectCount++
		}
	}

	hasEffect := effectCount > 0

	logrus.WithFields(logrus.Fields{
		"function":     "HasEffect",
		"package":      "game",
		"effect_type":  effectType,
		"has_effect":   hasEffect,
		"effect_count": effectCount,
		"total_active": len(em.activeEffects),
	}).Debug("function exit - effect check completed")

	return hasEffect
}

// AddEffect applies an effect to the entity
func (em *EffectManager) AddEffect(effect *Effect) error {
	logrus.WithFields(logrus.Fields{
		"function":    "AddEffect",
		"package":     "game",
		"effect_id":   effect.ID,
		"effect_type": effect.Type,
	}).Debug("function entry - delegating to ApplyEffect")

	err := em.ApplyEffect(effect)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function":    "AddEffect",
			"package":     "game",
			"effect_id":   effect.ID,
			"effect_type": effect.Type,
			"error":       err.Error(),
		}).Error("function exit - failed to apply effect")
	} else {
		logrus.WithFields(logrus.Fields{
			"function":    "AddEffect",
			"package":     "game",
			"effect_id":   effect.ID,
			"effect_type": effect.Type,
		}).Debug("function exit - effect applied successfully")
	}

	return err
}

// GetEffects returns a slice of all active effects
func (em *EffectManager) GetEffects() []*Effect {
	logrus.WithFields(logrus.Fields{
		"function": "GetEffects",
		"package":  "game",
	}).Debug("function entry - retrieving active effects")

	em.mu.RLock()
	defer em.mu.RUnlock()

	effects := make([]*Effect, 0, len(em.activeEffects))
	activeCount := 0
	for _, effect := range em.activeEffects {
		if effect.IsActive {
			effects = append(effects, effect)
			activeCount++
		}
	}

	logrus.WithFields(logrus.Fields{
		"function":      "GetEffects",
		"package":       "game",
		"active_count":  activeCount,
		"total_effects": len(em.activeEffects),
	}).Debug("function exit - active effects retrieved")

	return effects
}

// GetStats returns the current stats (with effects applied)
func (em *EffectManager) GetStats() *Stats {
	logrus.WithFields(logrus.Fields{
		"function": "GetStats",
		"package":  "game",
	}).Debug("function entry - retrieving current stats with effects")

	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := em.currentStats.Clone()

	logrus.WithFields(logrus.Fields{
		"function":     "GetStats",
		"package":      "game",
		"health":       stats.Health,
		"mana":         stats.Mana,
		"strength":     stats.Strength,
		"dexterity":    stats.Dexterity,
		"intelligence": stats.Intelligence,
	}).Debug("function exit - current stats retrieved")

	return stats
}

// SetStats updates the current stats
func (em *EffectManager) SetStats(stats *Stats) {
	logrus.WithFields(logrus.Fields{
		"function":     "SetStats",
		"package":      "game",
		"new_health":   stats.Health,
		"new_mana":     stats.Mana,
		"new_strength": stats.Strength,
	}).Debug("function entry - updating current stats")

	em.mu.Lock()
	defer em.mu.Unlock()

	oldStats := em.currentStats
	em.currentStats = stats.Clone()

	logrus.WithFields(logrus.Fields{
		"function":     "SetStats",
		"package":      "game",
		"old_health":   oldStats.Health,
		"new_health":   em.currentStats.Health,
		"old_mana":     oldStats.Mana,
		"new_mana":     em.currentStats.Mana,
		"old_strength": oldStats.Strength,
		"new_strength": em.currentStats.Strength,
	}).Debug("function exit - current stats updated")
}

// GetBaseStats returns the base stats (without effects)
func (em *EffectManager) GetBaseStats() *Stats {
	logrus.WithFields(logrus.Fields{
		"function": "GetBaseStats",
		"package":  "game",
	}).Debug("function entry - retrieving base stats without effects")

	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := em.baseStats.Clone()

	logrus.WithFields(logrus.Fields{
		"function":     "GetBaseStats",
		"package":      "game",
		"health":       stats.Health,
		"mana":         stats.Mana,
		"strength":     stats.Strength,
		"dexterity":    stats.Dexterity,
		"intelligence": stats.Intelligence,
	}).Debug("function exit - base stats retrieved")

	return stats
}
