package game

import (
	"reflect"
	"testing"
	"time"
)

// TestCreatePoisonEffect tests the creation of poison damage effects
func TestCreatePoisonEffect(t *testing.T) {
	tests := []struct {
		name       string
		baseDamage float64
		duration   time.Duration
		wantType   EffectType
		wantDamage DamageType
		wantScale  float64
		wantPen    float64
	}{
		{
			name:       "Basic poison effect",
			baseDamage: 10.0,
			duration:   30 * time.Second,
			wantType:   EffectPoison,
			wantDamage: DamagePoison,
			wantScale:  0.8,
			wantPen:    0,
		},
		{
			name:       "High damage poison",
			baseDamage: 25.5,
			duration:   60 * time.Second,
			wantType:   EffectPoison,
			wantDamage: DamagePoison,
			wantScale:  0.8,
			wantPen:    0,
		},
		{
			name:       "Zero damage poison",
			baseDamage: 0,
			duration:   5 * time.Second,
			wantType:   EffectPoison,
			wantDamage: DamagePoison,
			wantScale:  0.8,
			wantPen:    0,
		},
		{
			name:       "Long duration poison",
			baseDamage: 15.0,
			duration:   5 * time.Minute,
			wantType:   EffectPoison,
			wantDamage: DamagePoison,
			wantScale:  0.8,
			wantPen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreatePoisonEffect(tt.baseDamage, tt.duration)

			// Verify return value is not nil
			if result == nil {
				t.Fatal("CreatePoisonEffect returned nil")
			}

			// Verify underlying Effect is not nil
			if result.Effect == nil {
				t.Fatal("DamageEffect.Effect is nil")
			}

			// Test effect type
			if result.Effect.Type != tt.wantType {
				t.Errorf("Effect.Type = %v, want %v", result.Effect.Type, tt.wantType)
			}

			// Test damage type
			if result.DamageType != tt.wantDamage {
				t.Errorf("DamageType = %v, want %v", result.DamageType, tt.wantDamage)
			}

			// Test base damage
			if result.BaseDamage != tt.baseDamage {
				t.Errorf("BaseDamage = %v, want %v", result.BaseDamage, tt.baseDamage)
			}

			// Test damage scale
			if result.DamageScale != tt.wantScale {
				t.Errorf("DamageScale = %v, want %v", result.DamageScale, tt.wantScale)
			}

			// Test penetration percentage
			if result.PenetrationPct != tt.wantPen {
				t.Errorf("PenetrationPct = %v, want %v", result.PenetrationPct, tt.wantPen)
			}

			// Test duration
			if result.Effect.Duration.RealTime != tt.duration {
				t.Errorf("Duration.RealTime = %v, want %v", result.Effect.Duration.RealTime, tt.duration)
			}

			// Test magnitude matches base damage
			if result.Effect.Magnitude != tt.baseDamage {
				t.Errorf("Effect.Magnitude = %v, want %v", result.Effect.Magnitude, tt.baseDamage)
			}

			// Test effect is active
			if !result.Effect.IsActive {
				t.Error("Effect should be active")
			}

			// Test stacks is 1
			if result.Effect.Stacks != 1 {
				t.Errorf("Effect.Stacks = %v, want 1", result.Effect.Stacks)
			}
		})
	}
}

// TestCreateBurningEffect tests the creation of burning damage effects
func TestCreateBurningEffect(t *testing.T) {
	tests := []struct {
		name       string
		baseDamage float64
		duration   time.Duration
		wantType   EffectType
		wantDamage DamageType
		wantScale  float64
		wantPen    float64
	}{
		{
			name:       "Basic burning effect",
			baseDamage: 8.0,
			duration:   20 * time.Second,
			wantType:   EffectBurning,
			wantDamage: DamageFire,
			wantScale:  1.2,
			wantPen:    0,
		},
		{
			name:       "Intense burning effect",
			baseDamage: 50.0,
			duration:   45 * time.Second,
			wantType:   EffectBurning,
			wantDamage: DamageFire,
			wantScale:  1.2,
			wantPen:    0,
		},
		{
			name:       "Minimal burning effect",
			baseDamage: 1.0,
			duration:   time.Second,
			wantType:   EffectBurning,
			wantDamage: DamageFire,
			wantScale:  1.2,
			wantPen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateBurningEffect(tt.baseDamage, tt.duration)

			// Verify return value is not nil
			if result == nil {
				t.Fatal("CreateBurningEffect returned nil")
			}

			// Verify underlying Effect is not nil
			if result.Effect == nil {
				t.Fatal("DamageEffect.Effect is nil")
			}

			// Test effect type
			if result.Effect.Type != tt.wantType {
				t.Errorf("Effect.Type = %v, want %v", result.Effect.Type, tt.wantType)
			}

			// Test damage type
			if result.DamageType != tt.wantDamage {
				t.Errorf("DamageType = %v, want %v", result.DamageType, tt.wantDamage)
			}

			// Test base damage
			if result.BaseDamage != tt.baseDamage {
				t.Errorf("BaseDamage = %v, want %v", result.BaseDamage, tt.baseDamage)
			}

			// Test damage scale (burning has 1.2x scale)
			if result.DamageScale != tt.wantScale {
				t.Errorf("DamageScale = %v, want %v", result.DamageScale, tt.wantScale)
			}

			// Test penetration percentage
			if result.PenetrationPct != tt.wantPen {
				t.Errorf("PenetrationPct = %v, want %v", result.PenetrationPct, tt.wantPen)
			}

			// Test duration
			if result.Effect.Duration.RealTime != tt.duration {
				t.Errorf("Duration.RealTime = %v, want %v", result.Effect.Duration.RealTime, tt.duration)
			}

			// Test magnitude matches base damage
			if result.Effect.Magnitude != tt.baseDamage {
				t.Errorf("Effect.Magnitude = %v, want %v", result.Effect.Magnitude, tt.baseDamage)
			}
		})
	}
}

// TestCreateBleedingEffect tests the creation of bleeding damage effects
func TestCreateBleedingEffect(t *testing.T) {
	tests := []struct {
		name       string
		baseDamage float64
		duration   time.Duration
		wantType   EffectType
		wantDamage DamageType
		wantScale  float64
		wantPen    float64
	}{
		{
			name:       "Basic bleeding effect",
			baseDamage: 12.0,
			duration:   25 * time.Second,
			wantType:   EffectBleeding,
			wantDamage: DamagePhysical,
			wantScale:  1.0,
			wantPen:    0.5,
		},
		{
			name:       "Severe bleeding effect",
			baseDamage: 35.0,
			duration:   90 * time.Second,
			wantType:   EffectBleeding,
			wantDamage: DamagePhysical,
			wantScale:  1.0,
			wantPen:    0.5,
		},
		{
			name:       "Light bleeding effect",
			baseDamage: 2.5,
			duration:   10 * time.Second,
			wantType:   EffectBleeding,
			wantDamage: DamagePhysical,
			wantScale:  1.0,
			wantPen:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateBleedingEffect(tt.baseDamage, tt.duration)

			// Verify return value is not nil
			if result == nil {
				t.Fatal("CreateBleedingEffect returned nil")
			}

			// Verify underlying Effect is not nil
			if result.Effect == nil {
				t.Fatal("DamageEffect.Effect is nil")
			}

			// Test effect type
			if result.Effect.Type != tt.wantType {
				t.Errorf("Effect.Type = %v, want %v", result.Effect.Type, tt.wantType)
			}

			// Test damage type
			if result.DamageType != tt.wantDamage {
				t.Errorf("DamageType = %v, want %v", result.DamageType, tt.wantDamage)
			}

			// Test base damage
			if result.BaseDamage != tt.baseDamage {
				t.Errorf("BaseDamage = %v, want %v", result.BaseDamage, tt.baseDamage)
			}

			// Test damage scale (bleeding has 1.0x scale)
			if result.DamageScale != tt.wantScale {
				t.Errorf("DamageScale = %v, want %v", result.DamageScale, tt.wantScale)
			}

			// Test penetration percentage (bleeding ignores 50% of armor)
			if result.PenetrationPct != tt.wantPen {
				t.Errorf("PenetrationPct = %v, want %v", result.PenetrationPct, tt.wantPen)
			}

			// Test duration
			if result.Effect.Duration.RealTime != tt.duration {
				t.Errorf("Duration.RealTime = %v, want %v", result.Effect.Duration.RealTime, tt.duration)
			}

			// Test magnitude matches base damage
			if result.Effect.Magnitude != tt.baseDamage {
				t.Errorf("Effect.Magnitude = %v, want %v", result.Effect.Magnitude, tt.baseDamage)
			}
		})
	}
}

// TestAsDamageEffect tests conversion of Effects to DamageEffects
func TestAsDamageEffect(t *testing.T) {
	tests := []struct {
		name           string
		effect         *Effect
		wantSuccess    bool
		wantDamageType DamageType
		wantBaseDamage float64
		wantScale      float64
		wantPen        float64
	}{
		{
			name: "Convert poison effect",
			effect: &Effect{
				Type:       EffectPoison,
				DamageType: DamagePoison,
				Magnitude:  15.0,
			},
			wantSuccess:    true,
			wantDamageType: DamagePoison,
			wantBaseDamage: 15.0,
			wantScale:      0,
			wantPen:        0,
		},
		{
			name: "Convert burning effect",
			effect: &Effect{
				Type:       EffectBurning,
				DamageType: DamageFire,
				Magnitude:  20.0,
			},
			wantSuccess:    true,
			wantDamageType: DamageFire,
			wantBaseDamage: 20.0,
			wantScale:      0,
			wantPen:        0,
		},
		{
			name: "Convert bleeding effect",
			effect: &Effect{
				Type:       EffectBleeding,
				DamageType: DamagePhysical,
				Magnitude:  8.5,
			},
			wantSuccess:    true,
			wantDamageType: DamagePhysical,
			wantBaseDamage: 8.5,
			wantScale:      0,
			wantPen:        0,
		},
		{
			name: "Cannot convert stun effect",
			effect: &Effect{
				Type:      EffectStun,
				Magnitude: 5.0,
			},
			wantSuccess: false,
		},
		{
			name: "Cannot convert heal effect",
			effect: &Effect{
				Type:      EffectHealOverTime,
				Magnitude: 10.0,
			},
			wantSuccess: false,
		},
		{
			name: "Cannot convert stat boost",
			effect: &Effect{
				Type:      EffectStatBoost,
				Magnitude: 3.0,
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, success := AsDamageEffect(tt.effect)

			// Test success/failure expectation
			if success != tt.wantSuccess {
				t.Errorf("AsDamageEffect success = %v, want %v", success, tt.wantSuccess)
			}

			if tt.wantSuccess {
				// Test successful conversions
				if result == nil {
					t.Fatal("AsDamageEffect returned nil on successful conversion")
				}

				// Verify the underlying effect is preserved
				if result.Effect != tt.effect {
					t.Error("Converted DamageEffect should reference the original Effect")
				}

				// Test damage type
				if result.DamageType != tt.wantDamageType {
					t.Errorf("DamageType = %v, want %v", result.DamageType, tt.wantDamageType)
				}

				// Test base damage matches magnitude
				if result.BaseDamage != tt.wantBaseDamage {
					t.Errorf("BaseDamage = %v, want %v", result.BaseDamage, tt.wantBaseDamage)
				}

				// Test damage scale is 0 (as per specification)
				if result.DamageScale != tt.wantScale {
					t.Errorf("DamageScale = %v, want %v", result.DamageScale, tt.wantScale)
				}

				// Test penetration percentage is 0 (as per specification)
				if result.PenetrationPct != tt.wantPen {
					t.Errorf("PenetrationPct = %v, want %v", result.PenetrationPct, tt.wantPen)
				}
			} else {
				// Test failed conversions
				if result != nil {
					t.Error("AsDamageEffect should return nil on failed conversion")
				}
			}
		})
	}
}

// TestAsDamageEffect_NilInput tests AsDamageEffect with nil input
func TestAsDamageEffect_NilInput(t *testing.T) {
	result, success := AsDamageEffect(nil)

	if success {
		t.Error("AsDamageEffect should fail with nil input")
	}

	if result != nil {
		t.Error("AsDamageEffect should return nil with nil input")
	}
}

// TestDamageEffect_GetEffect tests the GetEffect method
func TestDamageEffect_GetEffect(t *testing.T) {
	// Create a poison effect to test with
	poisonEffect := CreatePoisonEffect(10.0, 30*time.Second)

	// Test GetEffect method
	effect := poisonEffect.GetEffect()

	// Verify the returned effect is the same as the underlying effect
	if effect != poisonEffect.Effect {
		t.Error("GetEffect should return the underlying Effect pointer")
	}

	// Verify it's not nil
	if effect == nil {
		t.Error("GetEffect should not return nil")
	}

	// Verify properties are accessible through the returned effect
	if effect.Type != EffectPoison {
		t.Errorf("Effect.Type = %v, want %v", effect.Type, EffectPoison)
	}
}

// TestCreateEffects_DurationRoundsAndTurns tests that created effects have correct duration structure
func TestCreateEffects_DurationRoundsAndTurns(t *testing.T) {
	duration := 45 * time.Second

	tests := []struct {
		name   string
		create func() *DamageEffect
	}{
		{
			name:   "Poison effect duration",
			create: func() *DamageEffect { return CreatePoisonEffect(10.0, duration) },
		},
		{
			name:   "Burning effect duration",
			create: func() *DamageEffect { return CreateBurningEffect(10.0, duration) },
		},
		{
			name:   "Bleeding effect duration",
			create: func() *DamageEffect { return CreateBleedingEffect(10.0, duration) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := tt.create()

			// Test that rounds and turns are set to 0 (real-time based effect)
			if effect.Effect.Duration.Rounds != 0 {
				t.Errorf("Duration.Rounds = %v, want 0", effect.Effect.Duration.Rounds)
			}

			if effect.Effect.Duration.Turns != 0 {
				t.Errorf("Duration.Turns = %v, want 0", effect.Effect.Duration.Turns)
			}

			// Test that RealTime is set correctly
			if effect.Effect.Duration.RealTime != duration {
				t.Errorf("Duration.RealTime = %v, want %v", effect.Effect.Duration.RealTime, duration)
			}
		})
	}
}

// TestCreateEffects_CompareAttributes tests that different effect types have different attributes
func TestCreateEffects_CompareAttributes(t *testing.T) {
	baseDamage := 10.0
	duration := 30 * time.Second

	poison := CreatePoisonEffect(baseDamage, duration)
	burning := CreateBurningEffect(baseDamage, duration)
	bleeding := CreateBleedingEffect(baseDamage, duration)

	// Test that each effect has different properties as expected
	effects := []struct {
		name           string
		effect         *DamageEffect
		expectedType   EffectType
		expectedDamage DamageType
		expectedScale  float64
		expectedPen    float64
	}{
		{"Poison", poison, EffectPoison, DamagePoison, 0.8, 0},
		{"Burning", burning, EffectBurning, DamageFire, 1.2, 0},
		{"Bleeding", bleeding, EffectBleeding, DamagePhysical, 1.0, 0.5},
	}

	for _, e := range effects {
		t.Run(e.name, func(t *testing.T) {
			if e.effect.Effect.Type != e.expectedType {
				t.Errorf("%s Type = %v, want %v", e.name, e.effect.Effect.Type, e.expectedType)
			}
			if e.effect.DamageType != e.expectedDamage {
				t.Errorf("%s DamageType = %v, want %v", e.name, e.effect.DamageType, e.expectedDamage)
			}
			if e.effect.DamageScale != e.expectedScale {
				t.Errorf("%s DamageScale = %v, want %v", e.name, e.effect.DamageScale, e.expectedScale)
			}
			if e.effect.PenetrationPct != e.expectedPen {
				t.Errorf("%s PenetrationPct = %v, want %v", e.name, e.effect.PenetrationPct, e.expectedPen)
			}
		})
	}
}

// TestCreateEffects_EdgeCases tests edge cases for effect creation
func TestCreateEffects_EdgeCases(t *testing.T) {
	t.Run("Zero duration effects", func(t *testing.T) {
		poison := CreatePoisonEffect(10.0, 0)
		burning := CreateBurningEffect(10.0, 0)
		bleeding := CreateBleedingEffect(10.0, 0)

		effects := []*DamageEffect{poison, burning, bleeding}
		for i, effect := range effects {
			if effect.Effect.Duration.RealTime != 0 {
				t.Errorf("Effect %d should have zero duration", i)
			}
		}
	})

	t.Run("Negative damage values", func(t *testing.T) {
		poison := CreatePoisonEffect(-5.0, 30*time.Second)
		burning := CreateBurningEffect(-10.0, 30*time.Second)
		bleeding := CreateBleedingEffect(-2.5, 30*time.Second)

		effects := []*DamageEffect{poison, burning, bleeding}
		expectedDamages := []float64{-5.0, -10.0, -2.5}

		for i, effect := range effects {
			if effect.BaseDamage != expectedDamages[i] {
				t.Errorf("Effect %d BaseDamage = %v, want %v", i, effect.BaseDamage, expectedDamages[i])
			}
			if effect.Effect.Magnitude != expectedDamages[i] {
				t.Errorf("Effect %d Magnitude = %v, want %v", i, effect.Effect.Magnitude, expectedDamages[i])
			}
		}
	})

	t.Run("Very large damage values", func(t *testing.T) {
		largeDamage := 999999.99
		poison := CreatePoisonEffect(largeDamage, time.Hour)

		if poison.BaseDamage != largeDamage {
			t.Errorf("BaseDamage = %v, want %v", poison.BaseDamage, largeDamage)
		}
		if poison.Effect.Magnitude != largeDamage {
			t.Errorf("Magnitude = %v, want %v", poison.Effect.Magnitude, largeDamage)
		}
	})
}

// TestAsDamageEffect_DeepEquality tests that AsDamageEffect preserves all effect data
func TestAsDamageEffect_DeepEquality(t *testing.T) {
	// Create a complex effect with all fields populated
	originalEffect := &Effect{
		ID:           "test-effect-123",
		Type:         EffectPoison,
		Name:         "Test Poison",
		Description:  "A test poison effect",
		DamageType:   DamagePoison,
		Magnitude:    25.5,
		IsActive:     true,
		Stacks:       3,
		Tags:         []string{"damage", "poison", "test"},
		SourceID:     "source-123",
		TargetID:     "target-456",
		StatAffected: "health",
	}

	// Convert to DamageEffect
	damageEffect, success := AsDamageEffect(originalEffect)

	if !success {
		t.Fatal("AsDamageEffect should succeed for poison effect")
	}

	// Verify the underlying effect is exactly the same reference
	if damageEffect.Effect != originalEffect {
		t.Error("DamageEffect should reference the original Effect, not a copy")
	}

	// Verify all fields are preserved through the reference
	if damageEffect.Effect.ID != originalEffect.ID {
		t.Error("Effect ID should be preserved")
	}
	if damageEffect.Effect.Name != originalEffect.Name {
		t.Error("Effect Name should be preserved")
	}
	if damageEffect.Effect.Description != originalEffect.Description {
		t.Error("Effect Description should be preserved")
	}
	if !reflect.DeepEqual(damageEffect.Effect.Tags, originalEffect.Tags) {
		t.Error("Effect Tags should be preserved")
	}
}

// TestEffectManager_calculateDamageWithResistance_DivisionByZero tests the critical edge case
// where division by zero would occur in damage calculation
func TestEffectManager_calculateDamageWithResistance_DivisionByZero(t *testing.T) {
	// Test the specific case mentioned in the audit: defense = -100 causing division by zero
	em := &EffectManager{
		currentStats: &Stats{
			Defense: -100.0,
		},
	}

	effect := &DamageEffect{
		DamageType:     DamageFire,
		PenetrationPct: 0.0,
	}

	// This should not panic - the key test is that it doesn't crash
	result := em.calculateDamageWithResistance(100.0, effect)

	// Verify we get a valid number (not NaN or Inf)
	if result != result { // NaN check
		t.Error("calculateDamageWithResistance() returned NaN")
	}
	if result == result+1 { // Inf check
		t.Error("calculateDamageWithResistance() returned Infinity")
	}
	if result < 0 {
		t.Errorf("calculateDamageWithResistance() returned negative damage: %v", result)
	}

	// For defense = -100, effectiveDefense = -100, denominator = 0
	// Our fix should set damageReduction = 1.0, so result should be baseDamage * 1.0 * resistanceMultiplier
	// Since resistance is 0 for fire damage by default, resistanceMultiplier should be 1.0
	// So result should equal baseDamage = 100.0
	if result != 100.0 {
		t.Errorf("calculateDamageWithResistance() = %v, want 100.0 (full damage when defense causes division by zero)", result)
	}
}

// TestEffectManager_calculateDamageWithResistance_PenetrationCausesDivisionByZero tests
// the case where penetration creates the division by zero condition
func TestEffectManager_calculateDamageWithResistance_PenetrationCausesDivisionByZero(t *testing.T) {
	// Test case where penetration causes division by zero: defense = -200, penetration = 0.5
	// effectiveDefense = -200 * (1 - 0.5) = -200 * 0.5 = -100
	em := &EffectManager{
		currentStats: &Stats{
			Defense: -200.0,
		},
	}

	effect := &DamageEffect{
		DamageType:     DamageFire,
		PenetrationPct: 0.5,
	}

	// This should not panic
	result := em.calculateDamageWithResistance(100.0, effect)

	// Verify we get a valid number (not NaN or Inf)
	if result != result { // NaN check
		t.Error("calculateDamageWithResistance() returned NaN")
	}
	if result == result+1 { // Inf check
		t.Error("calculateDamageWithResistance() returned Infinity")
	}
	if result < 0 {
		t.Errorf("calculateDamageWithResistance() returned negative damage: %v", result)
	}
}
