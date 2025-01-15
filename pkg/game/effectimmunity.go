package game

import (
	"fmt"
	"sort"
	"time"
)

func (em *EffectManager) initializeDefaultImmunities() {
	// Example default immunities
	em.immunities[EffectPoison] = &ImmunityData{
		Type:       ImmunityPartial,
		Resistance: 0.25, // 25% poison resistance
	}
}

// AddImmunity adds or updates an immunity
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

// CheckImmunity returns immunity status for an effect type
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

	return &ImmunityData{Type: ImmunityNone}
}

// DispelEffects removes effects based on type and count
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
		Duration{RealTime: 60 * time.Second},
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
