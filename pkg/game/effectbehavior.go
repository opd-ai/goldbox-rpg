package game

import (
	"time"
)

// DamageEffect represents effects that deal damage
type DamageEffect struct {
	Effect         *Effect    `yaml:",inline"` // Change to pointer
	DamageType     DamageType `yaml:"damage_type"`
	BaseDamage     float64    `yaml:"base_damage"`
	DamageScale    float64    `yaml:"damage_scale"`
	PenetrationPct float64    `yaml:"penetration_pct"`
}

// Add methods to properly access Effect fields
func (de *DamageEffect) GetEffect() *Effect {
	return de.Effect
}

// Status effect creation functions
func CreatePoisonEffect(baseDamage float64, duration time.Duration) *DamageEffect {
	return &DamageEffect{
		Effect:      NewEffect(EffectPoison, Duration{RealTime: duration}, baseDamage),
		DamageType:  DamagePoison,
		BaseDamage:  baseDamage,
		DamageScale: 0.8, // Each stack does 80% of base damage
	}
}

func CreateBurningEffect(baseDamage float64, duration time.Duration) *DamageEffect {
	return &DamageEffect{
		Effect:      NewEffect(EffectBurning, Duration{RealTime: duration}, baseDamage),
		DamageType:  DamageFire,
		BaseDamage:  baseDamage,
		DamageScale: 1.2, // Fire stacks more intensely
	}
}

func CreateBleedingEffect(baseDamage float64, duration time.Duration) *DamageEffect {
	return &DamageEffect{
		Effect:         NewEffect(EffectBleeding, Duration{RealTime: duration}, baseDamage),
		DamageType:     DamagePhysical,
		BaseDamage:     baseDamage,
		DamageScale:    1.0,
		PenetrationPct: 0.5, // Bleeding ignores 50% of armor
	}
}

// Add method to check if Effect is DamageEffect
func AsDamageEffect(e *Effect) (*DamageEffect, bool) {
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

// Extend EffectManager with specific effect processing

func (em *EffectManager) processDamageEffect(effect *DamageEffect, currentTime time.Time) {
	if !effect.Effect.ShouldTick(currentTime) {
		return
	}

	// Calculate damage based on stacks and scaling
	baseDamage := effect.BaseDamage * effect.DamageScale * float64(effect.Effect.Stacks)

	// Apply damage type modifiers
	finalDamage := em.calculateDamageWithResistance(baseDamage, effect)

	// Apply the damage
	em.currentStats.Health -= finalDamage

	// Handle additional effect-specific behaviors
	switch effect.Effect.Type {
	case EffectBurning:
		em.currentStats.MaxMana *= 0.95
	case EffectBleeding:
		em.applyHealingDebuff(0.5)
	case EffectPoison:
		em.applyStatDebuff(0.98)
	}
}

// Helper methods for damage calculation

func (em *EffectManager) calculateDamageWithResistance(baseDamage float64, effect *DamageEffect) float64 {
	// Get defense and resistances
	defense := em.currentStats.Defense
	resistance := em.getResistanceForDamageType(effect.DamageType)

	// Apply penetration
	effectiveDefense := defense * (1 - effect.PenetrationPct)

	// Calculate damage reduction
	damageReduction := 1 - (effectiveDefense / (effectiveDefense + 100))
	resistanceMultiplier := 1 - resistance

	return baseDamage * damageReduction * resistanceMultiplier
}

func (em *EffectManager) getResistanceForDamageType(dmgType DamageType) float64 {
	// This could be expanded to check equipment, buffs, etc.
	switch dmgType {
	case DamageFire:
		return em.resistances[EffectBurning]
	case DamagePoison:
		return em.resistances[EffectPoison]
	default:
		return 0
	}
}

// Status effect utility methods

func (em *EffectManager) applyStatDebuff(multiplier float64) {
	em.currentStats.Strength *= multiplier
	em.currentStats.Dexterity *= multiplier
	em.currentStats.Intelligence *= multiplier
}

func (em *EffectManager) applyHealingDebuff(multiplier float64) {
	// Store healing modifier in the manager
	em.healingModifier = multiplier
}

// Update the main effect processing to handle damage effects
func (em *EffectManager) processEffectTick(effect *Effect) {
	if damageEffect, ok := ToDamageEffect(effect); ok {
		em.processDamageEffect(damageEffect, time.Now())
		return
	}

	// Handle other effect types...
	switch effect.Type {
	case EffectDamageOverTime:
		em.currentStats.Health -= effect.Magnitude * float64(effect.Stacks)
	case EffectHealOverTime:
		healing := effect.Magnitude * float64(effect.Stacks)
		if em.healingModifier != 0 {
			healing *= em.healingModifier
		}
		em.currentStats.Health = min(
			em.currentStats.Health+healing,
			em.currentStats.MaxHealth,
		)
	}
}
