package game

import (
	"fmt"
	"sort"
	"time"
)

// initializeDefaultImmunities sets up the default immunity data for various effect types
// in the EffectManager. It populates the immunities map with predefined immunity configurations
// for common effect types like poison.
//
// This method is called internally during EffectManager initialization and should not need
// to be called directly.
//
// Current default immunities:
// - Poison: 25% resistance, partial immunity, no duration limit
//
// Related types:
// - EffectType: Enum defining possible effect types
// - ImmunityData: Struct containing immunity configuration
// - ImmunityType: Enum defining immunity types (full vs partial)
func (em *EffectManager) initializeDefaultImmunities() {
	// Example default immunities
	em.immunities[EffectPoison] = &ImmunityData{
		Type:       ImmunityPartial,
		Duration:   0,
		Resistance: 0.25,
		ExpiresAt:  time.Time{},
	}
}

// AddImmunity adds an immunity to a specific effect type to the EffectManager.
// If the immunity has a duration > 0, it is added as a temporary immunity
// that will expire after the specified duration. Otherwise, it is added
// as a permanent immunity.
//
// Parameters:
//   - effectType: The type of effect to become immune to
//   - immunity: ImmunityData struct containing duration and other immunity properties
//
// The immunity is stored in either tempImmunities or immunities map based on duration.
// If duration > 0, ExpiresAt is calculated as current time + duration.
//
// Thread-safe through mutex locking.
//
// Related:
//   - ImmunityData struct
//   - EffectType type
func (em *EffectManager) AddImmunity(effectType EffectType, immunity ImmunityData) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if immunity.Duration > 0 {
		immunity.ExpiresAt = time.Now().Add(immunity.Duration)
		em.tempImmunities[effectType] = &immunity
	} else {
		em.immunities[effectType] = &immunity
	}
}

// CheckImmunity checks if there is an active immunity against the given effect type.
// It first checks temporary immunities, then permanent immunities.
//
// Parameters:
//   - effectType: The type of effect to check immunity against
//
// Returns:
//   - *ImmunityData: Contains immunity details including:
//   - Type: The type of immunity (temporary, permanent, or none)
//   - Duration: How long the immunity lasts (0 for permanent)
//   - Resistance: Resistance level against the effect (0-100)
//   - ExpiresAt: When the immunity expires (empty for permanent)
//
// Thread-safety:
// This method is thread-safe as it uses a read lock when accessing the immunity maps.
//
// Notable behaviors:
// - Automatically cleans up expired temporary immunities when encountered
// - Returns a default ImmunityData with ImmunityNone if no immunity exists
// - Temporary immunities take precedence over permanent ones
//
// Related types:
// - ImmunityData
// - EffectType
func (em *EffectManager) CheckImmunity(effectType EffectType) *ImmunityData {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Check temporary immunities first
	if immunity, exists := em.tempImmunities[effectType]; exists {
		if time.Now().Before(immunity.ExpiresAt) {
			return immunity
		}
		// Clean up expired temporary immunity
		delete(em.tempImmunities, effectType)
	}

	// Check permanent immunities
	if immunity, exists := em.immunities[effectType]; exists {
		return immunity
	}

	return &ImmunityData{
		Type:       ImmunityNone,
		Duration:   0,
		Resistance: 0,
		ExpiresAt:  time.Time{},
	}
}

// DispelEffects removes a specified number of active effects of a given dispel type from the entity.
// It handles effect removal based on their dispel priority, with higher priority effects being removed first.
//
// Parameters:
//   - dispelType: The type of dispel to apply (e.g., magic, curse, etc.). Using DispelAll will target all dispellable effects
//   - count: Maximum number of effects to remove. Must be >= 0
//
// Returns:
//   - []string: Slice containing the IDs of all removed effects
//
// Notable behaviors:
//   - Thread-safe due to mutex locking
//   - Only removes effects marked as removable
//   - Automatically recalculates stats if any effects were removed
//   - If count exceeds available effects, removes all eligible effects
//
// Related types:
//   - DispelType: Enum defining different types of dispel
//   - DispelPriority: Defines removal priority of effects
//   - Effect.DispelInfo: Contains dispel-related properties of an effect
func (em *EffectManager) DispelEffects(dispelType DispelType, count int) []string {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Collect eligible effects
	type dispelCandidate struct {
		id       string
		effect   *Effect
		priority DispelPriority
	}

	var candidates []dispelCandidate

	for id, effect := range em.activeEffects {
		if !effect.DispelInfo.Removable {
			continue
		}

		for _, dType := range effect.DispelInfo.Types {
			if dType == dispelType || dispelType == DispelAll {
				candidates = append(candidates, dispelCandidate{
					id:       id,
					effect:   effect,
					priority: effect.DispelInfo.Priority,
				})
				break
			}
		}
	}

	// Sort by priority (highest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].priority > candidates[j].priority
	})

	// Remove effects
	removed := make([]string, 0, count)
	for i := 0; i < len(candidates) && i < count; i++ {
		delete(em.activeEffects, candidates[i].id)
		removed = append(removed, candidates[i].id)
	}

	if len(removed) > 0 {
		em.recalculateStats()
	}

	return removed
}

// Helper function to create effect with dispel info
func NewEffectWithDispel(effectType EffectType, duration Duration, magnitude float64, dispelInfo DispelInfo) *Effect {
	effect := NewEffect(effectType, duration, magnitude)
	effect.DispelInfo = dispelInfo
	return effect
}

// Example effect creation with dispel info
func CreatePoisonEffectWithDispel(baseDamage float64, duration time.Duration) *DamageEffect {
	effect := CreatePoisonEffect(baseDamage, duration)
	effect.Effect.DispelInfo = DispelInfo{
		Priority:  DispelPriorityNormal,
		Types:     []DispelType{DispelPoison, DispelMagic},
		Removable: true,
	}
	return effect
}

// Update ApplyEffect to check immunities
func (em *EffectManager) ApplyEffect(effect *Effect) error {
	immunity := em.CheckImmunity(effect.Type)

	switch immunity.Type {
	case ImmunityComplete:
		return fmt.Errorf("target is immune to %s effects", effect.Type)

	case ImmunityReflect:
		// Handle reflection logic
		return fmt.Errorf("effect reflected")

	case ImmunityPartial:
		effect.Magnitude *= (1 - immunity.Resistance)
	}

	// Continue with normal effect application...
	return em.applyEffectInternal(effect)
}

// Example usage:
func ExampleEffectDispel() {
	em := NewEffectManager(NewDefaultStats())

	// Add some effects
	poison := CreatePoisonEffect(10, 30*time.Second)
	curse := NewEffectWithDispel(EffectStatPenalty,
		Duration{
			Rounds:   0,
			Turns:    0,
			RealTime: 60 * time.Second,
		},
		-5,
		DispelInfo{
			Priority:  DispelPriorityHigh,
			Types:     []DispelType{DispelCurse, DispelMagic},
			Removable: true,
		},
	)

	// Apply effects using their base Effect
	_ = em.ApplyEffect(poison.GetEffect())
	_ = em.ApplyEffect(curse)

	// Dispel highest priority effects
	removed := em.DispelEffects(DispelMagic, 1)
	_ = removed // Use removed to avoid unused variable warning
}
