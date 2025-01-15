package game

import (
	"fmt"
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
