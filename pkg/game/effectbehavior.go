package game

import (
	"time"

	"github.com/sirupsen/logrus"
)

// DamageEffect represents a damage-dealing effect in the game system.
// It extends the base Effect struct with damage-specific attributes.
//
// Fields:
//   - Effect: Pointer to the base Effect struct containing common effect properties
//   - DamageType: The type of damage dealt (e.g. physical, magical, etc)
//   - BaseDamage: The base amount of damage before scaling
//   - DamageScale: Multiplier applied to the base damage
//   - PenetrationPct: Percentage of target's defense that is ignored
//
// Related types:
//   - Effect: Base effect type this extends
//   - DamageType: Enum of possible damage types
//
// Example usage:
//
//	damageEffect := &DamageEffect{
//	  Effect: &Effect{},
//	  DamageType: Physical,
//	  BaseDamage: 10.0,
//	  DamageScale: 1.5,
//	  PenetrationPct: 0.25,
//	}
type DamageEffect struct {
	Effect         *Effect    `yaml:",inline"` // Change to pointer
	DamageType     DamageType `yaml:"damage_type"`
	BaseDamage     float64    `yaml:"base_damage"`
	DamageScale    float64    `yaml:"damage_scale"`
	PenetrationPct float64    `yaml:"penetration_pct"`
}

// GetEffect returns the Effect object associated with this DamageEffect.
// This is an accessor method that provides access to the underlying Effect field.
//
// Returns:
//   - *Effect: A pointer to the Effect object contained within this DamageEffect
//
// Related types:
//   - Effect type
//   - DamageEffect type
func (de *DamageEffect) GetEffect() *Effect {
	return de.Effect
}

func CreatePoisonEffect(baseDamage float64, duration time.Duration) *DamageEffect {
	return &DamageEffect{
		Effect: NewEffect(EffectPoison, Duration{
			Rounds:   0,
			Turns:    0,
			RealTime: duration,
		}, baseDamage),
		DamageType:     DamagePoison,
		BaseDamage:     baseDamage,
		DamageScale:    0.8,
		PenetrationPct: 0,
	}
}

// CreateBurningEffect creates a new fire-based damage effect that deals damage over time
//
// Parameters:
//   - baseDamage: Base damage per tick (float64) that will be dealt
//   - duration: How long the burning effect lasts (time.Duration)
//
// Returns:
//
//	*DamageEffect - A configured burning damage effect with:
//	- Fire damage type
//	- 20% damage scaling multiplier
//	- No armor penetration
//	- Real-time based duration tracking
//
// Related:
//   - DamageEffect
//   - EffectBurning constant
//   - DamageFire constant
func CreateBurningEffect(baseDamage float64, duration time.Duration) *DamageEffect {
	return &DamageEffect{
		Effect: NewEffect(EffectBurning, Duration{
			Rounds:   0,
			Turns:    0,
			RealTime: duration,
		}, baseDamage),
		DamageType:     DamageFire,
		BaseDamage:     baseDamage,
		DamageScale:    1.2,
		PenetrationPct: 0,
	}
}

// CreateBleedingEffect creates a new bleeding damage effect that deals continuous physical damage over time
//
// Parameters:
//   - baseDamage: Base amount of physical damage dealt per tick (float64, must be >= 0)
//   - duration: How long the bleeding effect lasts (time.Duration)
//
// Returns:
//
//	*DamageEffect - A configured bleeding damage effect that:
//	- Deals physical damage over time
//	- Ignores 50% of armor via penetration
//	- Scales at 1.0x base damage
//
// Related:
//   - DamageEffect struct
//   - NewEffect() - Base effect constructor
//   - EffectBleeding constant
//   - DamagePhysical constant
func CreateBleedingEffect(baseDamage float64, duration time.Duration) *DamageEffect {
	return &DamageEffect{
		Effect: NewEffect(EffectBleeding, Duration{
			Rounds:   0,
			Turns:    0,
			RealTime: duration,
		}, baseDamage),
		DamageType:     DamagePhysical,
		BaseDamage:     baseDamage,
		DamageScale:    1.0,
		PenetrationPct: 0.5, // Bleeding ignores 50% of armor
	}
}

// Add method to check if Effect is DamageEffect
// AsDamageEffect attempts to convert a generic Effect into a DamageEffect.
//
// Parameters:
//   - e: A pointer to the Effect to convert
//
// Returns:
//   - *DamageEffect: A pointer to the created DamageEffect if conversion was successful
//   - bool: True if conversion was successful, false otherwise
//
// The function will only convert effects of type EffectPoison, EffectBurning, or EffectBleeding.
// For all other effect types, it returns nil and false.
//
// The resulting DamageEffect will:
// - Inherit the base Effect properties
// - Use the original Effect's DamageType and Magnitude
// - Have DamageScale and PenetrationPct set to 0
//
// Related types:
//   - Effect
//   - DamageEffect
//   - EffectType (EffectPoison, EffectBurning, EffectBleeding)
func AsDamageEffect(e *Effect) (*DamageEffect, bool) {
	if e == nil {
		return nil, false
	}
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

// Extend EffectManager with specific effect processing

// processDamageEffect applies damage and additional effects from a DamageEffect to the target.
// It checks if the effect should tick based on the current time, calculates the final damage
// accounting for stacks, scaling and resistance, and applies specific status effects based on
// the effect type.
//
// Parameters:
//   - effect: *DamageEffect - Contains damage and effect properties to be applied
//   - currentTime: time.Time - Current game time used to check effect tick timing
//
// Notable effects by type:
//   - EffectBurning: Reduces max mana by 5%
//   - EffectBleeding: Applies 50% healing debuff
//   - EffectPoison: Applies 2% stat reduction
//   - Other effect types are reserved for future implementation
//
// Panics if an unexpected EffectType is provided.
//
// Related types:
//   - DamageEffect
//   - Effect
//   - EffectType
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
	case EffectDamageOverTime:
	case EffectHealOverTime:
	case EffectRoot:
	case EffectStatBoost:
	case EffectStatPenalty:
	case EffectStun:
	default:
		logrus.WithField("effectType", effect.Effect.Type).Error("unsupported effect type in processDamageEffect")
	}
}

// Helper methods for damage calculation

// calculateDamageWithResistance calculates the final damage value after applying defense and resistance.
//
// Parameters:
//   - baseDamage (float64): The initial damage amount to be modified
//   - effect (*DamageEffect): Effect struct containing DamageType and PenetrationPct
//
// Returns:
//
//	float64: The final calculated damage value after applying defenses and resistances
//
// The calculation follows these steps:
// 1. Gets defense from currentStats and resistance for the damage type
// 2. Applies penetration percentage to reduce effective defense
// 3. Calculates damage reduction using standard formula: 1 - (defense/(defense + 100))
// 4. Applies resistance as a direct multiplier
//
// Related:
//   - DamageEffect struct
//   - getResistanceForDamageType method
func (em *EffectManager) calculateDamageWithResistance(baseDamage float64, effect *DamageEffect) float64 {
	// Get defense and resistances
	defense := em.currentStats.Defense
	resistance := em.getResistanceForDamageType(effect.DamageType)

	// Apply penetration
	effectiveDefense := defense * (1 - effect.PenetrationPct)

	// Calculate damage reduction with protection against division by zero
	var damageReduction float64
	denominator := effectiveDefense + 100
	if denominator == 0 {
		// Handle edge case where effectiveDefense = -100
		// In this case, assume maximum damage (no reduction)
		damageReduction = 1.0
	} else {
		damageReduction = 1 - (effectiveDefense / denominator)
	}
	resistanceMultiplier := 1 - resistance

	return baseDamage * damageReduction * resistanceMultiplier
}

// getResistanceForDamageType returns the resistance value for a given damage type.
// The resistance value reduces damage taken of that type, with higher values providing more protection.
//
// Parameters:
//   - dmgType: The type of damage to get resistance for (e.g. DamageFire, DamagePoison)
//
// Returns:
//
//	float64 representing the resistance value (0.0 - 1.0) where:
//	- 0.0 means no resistance (full damage taken)
//	- Higher values mean more resistance (less damage taken)
//
// Notes:
// - Currently only handles Fire and Poison damage types
// - All other damage types return 0 (no resistance)
// - Could be expanded to check equipment buffs and other modifiers
//
// Related:
// - DamageType enum
// - EffectType enum (for resistance mapping)
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

// applyStatDebuff applies a multiplier to reduce all base stats (Strength, Dexterity, Intelligence)
// of the current stats object managed by the EffectManager.
//
// Parameters:
//   - multiplier: float64 - The multiplier to apply to all stats (should be < 1.0 for debuffs)
//
// The function directly modifies the currentStats object and does not return a value.
//
// Note: This method assumes positive multiplier values. Negative multipliers would invert stats.
// Stats are not clamped to any min/max values after multiplication.
//
// Related:
// - EffectManager.currentStats field
// - Stats struct containing the base stats
func (em *EffectManager) applyStatDebuff(multiplier float64) {
	em.currentStats.Strength *= multiplier
	em.currentStats.Dexterity *= multiplier
	em.currentStats.Intelligence *= multiplier
}

// applyHealingDebuff applies a healing modifier multiplier to the effect manager
// that will scale any future healing effects.
//
// Parameters:
//   - multiplier: float64 - Scaling factor to modify healing effects. Values less than 1.0
//     reduce healing, values greater than 1.0 increase healing.
//
// This method updates the internal healingModifier field which is used by other
// methods when calculating healing amounts. The modifier persists until changed again.
//
// Related:
//   - EffectManager.healingModifier field
//   - Methods that apply healing effects should check this modifier
func (em *EffectManager) applyHealingDebuff(multiplier float64) {
	// Store healing modifier in the manager
	em.healingModifier = multiplier
}

// Update the main effect processing to handle damage effects
// processEffectTick processes a single tick of an effect on the target entity.
//
// It handles different types of effects including:
// - Damage effects (via processDamageEffect)
// - Damage over time
// - Healing over time
// - Status effects (bleeding, burning, poison, root, stun)
// - Stat modifications (boosts and penalties)
//
// Parameters:
//   - effect: *Effect - The effect to process, containing type, magnitude, stacks etc.
//
// Notable behaviors:
// - For damage effects, delegates to processDamageEffect
// - For healing, applies healing modifier and caps at max health
// - Panics on unknown effect types
//
// Related types:
// - Effect
// - DamageEffect
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
	case EffectBleeding:
	case EffectBurning:
	case EffectPoison:
	case EffectRoot:
	case EffectStatBoost:
	case EffectStatPenalty:
	case EffectStun:
	default:
		logrus.WithField("effectType", effect.Type).Error("unsupported effect type in processEffectTick")
	}
}
