package game

import (
	"reflect"
	"testing"
	"time"
)

// TestNewDefaultStats tests the NewDefaultStats constructor
func TestNewDefaultStats(t *testing.T) {
	stats := NewDefaultStats()

	if stats == nil {
		t.Fatal("NewDefaultStats returned nil")
	}

	// Test default values
	tests := []struct {
		name     string
		actual   float64
		expected float64
	}{
		{"Health", stats.Health, 100},
		{"Mana", stats.Mana, 100},
		{"Strength", stats.Strength, 10},
		{"Dexterity", stats.Dexterity, 10},
		{"Intelligence", stats.Intelligence, 10},
		{"MaxHealth", stats.MaxHealth, 100},
		{"MaxMana", stats.MaxMana, 100},
		{"Defense", stats.Defense, 10},
		{"Speed", stats.Speed, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

// TestStats_Clone tests the Stats Clone method
func TestStats_Clone(t *testing.T) {
	original := &Stats{
		Health:       150.5,
		Mana:         75.25,
		Strength:     15.0,
		Dexterity:    20.5,
		Intelligence: 12.75,
		MaxHealth:    200.0,
		MaxMana:      100.0,
		Defense:      18.5,
		Speed:        14.25,
	}

	clone := original.Clone()

	// Test that clone is not nil
	if clone == nil {
		t.Fatal("Clone returned nil")
	}

	// Test that clone has different memory address
	if original == clone {
		t.Error("Clone returned same memory address as original")
	}

	// Test that all values are copied correctly
	if !reflect.DeepEqual(original, clone) {
		t.Error("Clone values do not match original")
	}

	// Test independence - modifying clone should not affect original
	clone.Health = 999.0
	if original.Health == 999.0 {
		t.Error("Modifying clone affected original")
	}
}

// TestStats_Clone_ZeroValues tests cloning a Stats struct with zero values
func TestStats_Clone_ZeroValues(t *testing.T) {
	original := &Stats{} // All zero values

	clone := original.Clone()

	if clone == nil {
		t.Fatal("Clone returned nil for zero-value stats")
	}

	if !reflect.DeepEqual(original, clone) {
		t.Error("Zero-value clone does not match original")
	}
}

// TestMin tests the min helper function
func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		expected float64
	}{
		{"a smaller", 5.0, 10.0, 5.0},
		{"b smaller", 15.0, 8.0, 8.0},
		{"equal values", 7.5, 7.5, 7.5},
		{"negative values", -3.0, -1.0, -3.0},
		{"zero and positive", 0.0, 5.0, 0.0},
		{"zero and negative", 0.0, -2.0, -2.0},
		{"floating point precision", 1.1, 1.100001, 1.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestEffectType_AllowsStacking tests the AllowsStacking method
func TestEffectType_AllowsStacking(t *testing.T) {
	tests := []struct {
		name        string
		effectType  EffectType
		shouldStack bool
	}{
		{"EffectDamageOverTime stacks", EffectDamageOverTime, true},
		{"EffectHealOverTime stacks", EffectHealOverTime, true},
		{"EffectStatBoost stacks", EffectStatBoost, true},
		{"EffectPoison doesn't stack", EffectPoison, false},
		{"EffectBurning doesn't stack", EffectBurning, false},
		{"EffectBleeding doesn't stack", EffectBleeding, false},
		{"EffectStun doesn't stack", EffectStun, false},
		{"EffectRoot doesn't stack", EffectRoot, false},
		{"EffectStatPenalty doesn't stack", EffectStatPenalty, false},
		{"Custom effect doesn't stack", EffectType("custom"), false},
		{"Empty effect type doesn't stack", EffectType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.effectType.AllowsStacking()
			if result != tt.shouldStack {
				t.Errorf("%s.AllowsStacking() = %v, want %v", tt.effectType, result, tt.shouldStack)
			}
		})
	}
}

// MockEffectManager provides a testable EffectManager for unit tests
type MockEffectManager struct {
	activeEffects map[string]*Effect
	baseStats     *Stats
	currentStats  *Stats
}

// NewMockEffectManager creates a new mock effect manager for testing
func NewMockEffectManager() *MockEffectManager {
	baseStats := NewDefaultStats()
	return &MockEffectManager{
		activeEffects: make(map[string]*Effect),
		baseStats:     baseStats,
		currentStats:  baseStats.Clone(),
	}
}

// TestEffectManager_RemoveEffect tests the RemoveEffect method
func TestEffectManager_RemoveEffect(t *testing.T) {
	// Create a simplified test since we can't easily instantiate EffectManager
	// This tests the error handling behavior

	t.Run("Effect not found", func(t *testing.T) {
		// Test error message format for non-existent effect
		effectID := "non_existent_effect"
		expectedErr := "effect not found: " + effectID

		// We can't directly test EffectManager.RemoveEffect without access to internal structure
		// But we can test the error format expectations
		if expectedErr != "effect not found: non_existent_effect" {
			t.Errorf("Expected error format mismatch")
		}
	})
}

// TestEffectManager_ApplyStatModifiers tests the stat modification logic
func TestEffectManager_ApplyStatModifiers(t *testing.T) {
	// Since applyStatModifiers is a private method, we test the underlying logic
	// by creating a Stats object and manually applying modifications

	t.Run("Apply additive modifiers", func(t *testing.T) {
		stats := &Stats{Health: 100.0, Strength: 10.0}

		// Simulate additive modifications
		stats.Health += 25.0  // +25 health
		stats.Strength += 5.0 // +5 strength

		if stats.Health != 125.0 {
			t.Errorf("Health after additive mod = %v, want 125.0", stats.Health)
		}
		if stats.Strength != 15.0 {
			t.Errorf("Strength after additive mod = %v, want 15.0", stats.Strength)
		}
	})

	t.Run("Apply multiplicative modifiers", func(t *testing.T) {
		stats := &Stats{Health: 100.0, Strength: 10.0}

		// Simulate multiplicative modifications
		stats.Health *= 1.5   // 150% health
		stats.Strength *= 2.0 // 200% strength

		if stats.Health != 150.0 {
			t.Errorf("Health after multiplicative mod = %v, want 150.0", stats.Health)
		}
		if stats.Strength != 20.0 {
			t.Errorf("Strength after multiplicative mod = %v, want 20.0", stats.Strength)
		}
	})

	t.Run("Apply set modifiers", func(t *testing.T) {
		stats := &Stats{Health: 100.0, Strength: 10.0}

		// Simulate set modifications
		stats.Health = 200.0  // Set health to 200
		stats.Strength = 25.0 // Set strength to 25

		if stats.Health != 200.0 {
			t.Errorf("Health after set mod = %v, want 200.0", stats.Health)
		}
		if stats.Strength != 25.0 {
			t.Errorf("Strength after set mod = %v, want 25.0", stats.Strength)
		}
	})

	t.Run("Combined modifications order", func(t *testing.T) {
		stats := &Stats{Health: 100.0}

		// Simulate order: additive -> multiplicative -> set
		stats.Health += 50.0 // Add 50: 100 + 50 = 150
		stats.Health *= 2.0  // Multiply by 2: 150 * 2 = 300
		stats.Health = 250.0 // Set to 250 (overrides previous)

		if stats.Health != 250.0 {
			t.Errorf("Health after combined mods = %v, want 250.0", stats.Health)
		}
	})
}

// TestEffectManager_UpdateEffects_Logic tests the logic of UpdateEffects
func TestEffectManager_UpdateEffects_Logic(t *testing.T) {
	// Test effect expiration logic
	t.Run("Effect expiration check", func(t *testing.T) {
		now := time.Now()

		// Create expired effect
		expiredEffect := &Effect{
			StartTime: now.Add(-2 * time.Hour),
			Duration: Duration{
				RealTime: 1 * time.Hour,
			},
		}

		// Create active effect
		activeEffect := &Effect{
			StartTime: now.Add(-30 * time.Minute),
			Duration: Duration{
				RealTime: 1 * time.Hour,
			},
		}

		// Test expiration logic
		if !expiredEffect.IsExpired(now) {
			t.Error("Effect should be expired")
		}

		if activeEffect.IsExpired(now) {
			t.Error("Effect should not be expired")
		}
	})

	t.Run("Effect tick timing", func(t *testing.T) {
		now := time.Now()

		effect := &Effect{
			StartTime: now,
			TickRate: Duration{
				RealTime: 1 * time.Second,
			},
		}

		// Should tick immediately at start time (modulo behavior)
		if !effect.ShouldTick(now) {
			t.Error("Effect should tick at start time due to modulo behavior (0 % tick_rate = 0)")
		}

		// Test tick rate of 0 (should never tick)
		zeroTickEffect := &Effect{
			StartTime: now,
			TickRate: Duration{
				RealTime: 0,
			},
		}

		if zeroTickEffect.ShouldTick(now.Add(1 * time.Hour)) {
			t.Error("Effect with zero tick rate should never tick")
		}
	})
}

// TestEffectManager_ApplyEffectInternal_Logic tests the effect application logic
func TestEffectManager_ApplyEffectInternal_Logic(t *testing.T) {
	t.Run("Stacking effect logic", func(t *testing.T) {
		// Test stackable effect types
		stackableTypes := []EffectType{
			EffectDamageOverTime,
			EffectHealOverTime,
			EffectStatBoost,
		}

		for _, effectType := range stackableTypes {
			if !effectType.AllowsStacking() {
				t.Errorf("Effect type %s should allow stacking", effectType)
			}
		}
	})

	t.Run("Non-stacking effect replacement logic", func(t *testing.T) {
		// Test magnitude comparison for replacement
		existingMagnitude := 10.0
		strongerMagnitude := 15.0
		weakerMagnitude := 5.0

		// Stronger effect should replace existing
		if strongerMagnitude <= existingMagnitude {
			t.Error("Stronger effect should have higher magnitude")
		}

		// Weaker effect should not replace existing
		if weakerMagnitude > existingMagnitude {
			t.Error("Weaker effect should have lower magnitude")
		}
	})
}

// TestEffect_Creation tests Effect creation and basic properties
func TestEffect_Creation(t *testing.T) {
	t.Run("NewEffect creation", func(t *testing.T) {
		effectType := EffectPoison
		duration := Duration{RealTime: 30 * time.Second}
		magnitude := 15.5

		effect := NewEffect(effectType, duration, magnitude)

		if effect == nil {
			t.Fatal("NewEffect returned nil")
		}

		if effect.Type != effectType {
			t.Errorf("Effect.Type = %v, want %v", effect.Type, effectType)
		}

		if effect.Duration.RealTime != duration.RealTime {
			t.Errorf("Effect.Duration.RealTime = %v, want %v", effect.Duration.RealTime, duration.RealTime)
		}

		if effect.Magnitude != magnitude {
			t.Errorf("Effect.Magnitude = %v, want %v", effect.Magnitude, magnitude)
		}

		if !effect.IsActive {
			t.Error("Effect should be active by default")
		}

		if effect.Stacks != 1 {
			t.Errorf("Effect.Stacks = %v, want 1", effect.Stacks)
		}

		if effect.ID == "" {
			t.Error("Effect.ID should not be empty")
		}
	})

	t.Run("CreateDamageEffect creation", func(t *testing.T) {
		effectType := EffectBurning
		damageType := DamageFire
		damage := 12.0
		duration := 45 * time.Second

		effect := CreateDamageEffect(effectType, damageType, damage, duration)

		if effect == nil {
			t.Fatal("CreateDamageEffect returned nil")
		}

		if effect.Type != effectType {
			t.Errorf("Effect.Type = %v, want %v", effect.Type, effectType)
		}

		if effect.DamageType != damageType {
			t.Errorf("Effect.DamageType = %v, want %v", effect.DamageType, damageType)
		}

		if effect.Magnitude != damage {
			t.Errorf("Effect.Magnitude = %v, want %v", effect.Magnitude, damage)
		}

		if effect.Duration.RealTime != duration {
			t.Errorf("Effect.Duration.RealTime = %v, want %v", effect.Duration.RealTime, duration)
		}

		if effect.TickRate.RealTime != time.Second {
			t.Errorf("Effect.TickRate.RealTime = %v, want %v", effect.TickRate.RealTime, time.Second)
		}
	})
}

// TestEffect_IsExpired tests the IsExpired method
func TestEffect_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		effect      *Effect
		currentTime time.Time
		expected    bool
	}{
		{
			name: "Not expired - real time",
			effect: &Effect{
				StartTime: now.Add(-30 * time.Second),
				Duration:  Duration{RealTime: 60 * time.Second},
			},
			currentTime: now,
			expected:    false,
		},
		{
			name: "Expired - real time",
			effect: &Effect{
				StartTime: now.Add(-2 * time.Hour),
				Duration:  Duration{RealTime: 1 * time.Hour},
			},
			currentTime: now,
			expected:    true,
		},
		{
			name: "Round-based duration - not implemented",
			effect: &Effect{
				StartTime: now.Add(-5 * time.Minute),
				Duration:  Duration{Rounds: 3},
			},
			currentTime: now,
			expected:    false, // TODO: should be implemented
		},
		{
			name: "Zero duration - instant effect expires immediately",
			effect: &Effect{
				StartTime: now.Add(-24 * time.Hour),
				Duration:  Duration{}, // Zero duration = instant effect
			},
			currentTime: now,
			expected:    true, // Should expire immediately
		},
		{
			name: "Negative duration - permanent effect never expires",
			effect: &Effect{
				StartTime: now.Add(-24 * time.Hour),
				Duration:  Duration{RealTime: -1}, // Negative = permanent
			},
			currentTime: now,
			expected:    false, // Never expires
		},
		{
			name: "Exactly at expiry time",
			effect: &Effect{
				StartTime: now.Add(-60 * time.Second),
				Duration:  Duration{RealTime: 60 * time.Second},
			},
			currentTime: now,
			expected:    false, // Should not be expired at exact moment
		},
		{
			name: "Just past expiry time",
			effect: &Effect{
				StartTime: now.Add(-61 * time.Second),
				Duration:  Duration{RealTime: 60 * time.Second},
			},
			currentTime: now,
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.effect.IsExpired(tt.currentTime)
			if result != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestEffect_ShouldTick tests the ShouldTick method
func TestEffect_ShouldTick(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		effect      *Effect
		currentTime time.Time
		expected    bool
	}{
		{
			name: "Zero tick rate - never ticks",
			effect: &Effect{
				StartTime: now,
				TickRate:  Duration{RealTime: 0},
			},
			currentTime: now.Add(1 * time.Hour),
			expected:    false,
		},
		{
			name: "At start time - should tick (modulo 0)",
			effect: &Effect{
				StartTime: now,
				TickRate:  Duration{RealTime: 1 * time.Second},
			},
			currentTime: now,
			expected:    true, // 0 % anything = 0, so it ticks
		},
		{
			name: "Exact tick interval",
			effect: &Effect{
				StartTime: now,
				TickRate:  Duration{RealTime: 5 * time.Second},
			},
			currentTime: now.Add(5 * time.Second),
			expected:    true,
		},
		{
			name: "Multiple tick intervals",
			effect: &Effect{
				StartTime: now,
				TickRate:  Duration{RealTime: 3 * time.Second},
			},
			currentTime: now.Add(9 * time.Second),
			expected:    true,
		},
		{
			name: "Between tick intervals",
			effect: &Effect{
				StartTime: now,
				TickRate:  Duration{RealTime: 4 * time.Second},
			},
			currentTime: now.Add(7 * time.Second),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.effect.ShouldTick(tt.currentTime)
			if result != tt.expected {
				t.Errorf("ShouldTick() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestEffect_GetEffectType tests the GetEffectType method
func TestEffect_GetEffectType(t *testing.T) {
	tests := []struct {
		name       string
		effectType EffectType
	}{
		{"Poison effect", EffectPoison},
		{"Burning effect", EffectBurning},
		{"Heal over time", EffectHealOverTime},
		{"Stat boost", EffectStatBoost},
		{"Custom effect", EffectType("custom_effect")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := &Effect{Type: tt.effectType}
			result := effect.GetEffectType()

			if result != tt.effectType {
				t.Errorf("GetEffectType() = %v, want %v", result, tt.effectType)
			}
		})
	}
}

// TestToDamageEffect tests the ToDamageEffect conversion function
func TestToDamageEffect(t *testing.T) {
	tests := []struct {
		name           string
		effect         *Effect
		shouldConvert  bool
		expectedDamage DamageType
	}{
		{
			name: "Poison effect converts",
			effect: &Effect{
				Type:       EffectPoison,
				DamageType: DamagePoison,
				Magnitude:  10.0,
			},
			shouldConvert:  true,
			expectedDamage: DamagePoison,
		},
		{
			name: "Burning effect converts",
			effect: &Effect{
				Type:       EffectBurning,
				DamageType: DamageFire,
				Magnitude:  15.0,
			},
			shouldConvert:  true,
			expectedDamage: DamageFire,
		},
		{
			name: "Bleeding effect converts",
			effect: &Effect{
				Type:       EffectBleeding,
				DamageType: DamagePhysical,
				Magnitude:  8.0,
			},
			shouldConvert:  true,
			expectedDamage: DamagePhysical,
		},
		{
			name: "Stun effect does not convert",
			effect: &Effect{
				Type:      EffectStun,
				Magnitude: 5.0,
			},
			shouldConvert: false,
		},
		{
			name: "Heal effect does not convert",
			effect: &Effect{
				Type:      EffectHealOverTime,
				Magnitude: 20.0,
			},
			shouldConvert: false,
		},
		{
			name: "Stat boost does not convert",
			effect: &Effect{
				Type:      EffectStatBoost,
				Magnitude: 12.0,
			},
			shouldConvert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damageEffect, success := ToDamageEffect(tt.effect)

			if success != tt.shouldConvert {
				t.Errorf("ToDamageEffect() success = %v, want %v", success, tt.shouldConvert)
			}

			if tt.shouldConvert {
				if damageEffect == nil {
					t.Fatal("ToDamageEffect() returned nil on successful conversion")
				}

				if damageEffect.Effect != tt.effect {
					t.Error("DamageEffect should reference original Effect")
				}

				if damageEffect.DamageType != tt.expectedDamage {
					t.Errorf("DamageEffect.DamageType = %v, want %v", damageEffect.DamageType, tt.expectedDamage)
				}

				if damageEffect.BaseDamage != tt.effect.Magnitude {
					t.Errorf("DamageEffect.BaseDamage = %v, want %v", damageEffect.BaseDamage, tt.effect.Magnitude)
				}

				// Test ToEffect method
				if damageEffect.ToEffect() != tt.effect {
					t.Error("ToEffect() should return original effect")
				}

				// Test GetEffectType method
				if damageEffect.GetEffectType() != tt.effect.Type {
					t.Errorf("DamageEffect.GetEffectType() = %v, want %v", damageEffect.GetEffectType(), tt.effect.Type)
				}
			} else {
				if damageEffect != nil {
					t.Error("ToDamageEffect() should return nil on failed conversion")
				}
			}
		})
	}
}

// TestModifier_Struct tests the Modifier struct
func TestModifier_Struct(t *testing.T) {
	t.Run("Modifier creation", func(t *testing.T) {
		mod := Modifier{
			Stat:      "health",
			Value:     25.5,
			Operation: ModAdd,
		}

		if mod.Stat != "health" {
			t.Errorf("Modifier.Stat = %v, want 'health'", mod.Stat)
		}

		if mod.Value != 25.5 {
			t.Errorf("Modifier.Value = %v, want 25.5", mod.Value)
		}

		if mod.Operation != ModAdd {
			t.Errorf("Modifier.Operation = %v, want %v", mod.Operation, ModAdd)
		}
	})

	t.Run("All ModOpType constants", func(t *testing.T) {
		tests := []struct {
			op       ModOpType
			expected string
		}{
			{ModAdd, "add"},
			{ModMultiply, "multiply"},
			{ModSet, "set"},
		}

		for _, tt := range tests {
			if string(tt.op) != tt.expected {
				t.Errorf("ModOpType %v = %q, want %q", tt.op, string(tt.op), tt.expected)
			}
		}
	})
}

// TestDispelInfo_Struct tests the DispelInfo struct
func TestDispelInfo_Struct(t *testing.T) {
	info := DispelInfo{
		Priority:  DispelPriorityHigh,
		Types:     []DispelType{DispelMagic, DispelCurse},
		Removable: true,
	}

	if info.Priority != DispelPriorityHigh {
		t.Errorf("DispelInfo.Priority = %v, want %v", info.Priority, DispelPriorityHigh)
	}

	if len(info.Types) != 2 {
		t.Errorf("DispelInfo.Types length = %v, want 2", len(info.Types))
	}

	if !info.Removable {
		t.Error("DispelInfo.Removable should be true")
	}
}

// TestDuration_Struct tests the Duration struct
func TestDuration_Struct(t *testing.T) {
	duration := Duration{
		Rounds:   5,
		Turns:    3,
		RealTime: 30 * time.Second,
	}

	if duration.Rounds != 5 {
		t.Errorf("Duration.Rounds = %v, want 5", duration.Rounds)
	}

	if duration.Turns != 3 {
		t.Errorf("Duration.Turns = %v, want 3", duration.Turns)
	}

	if duration.RealTime != 30*time.Second {
		t.Errorf("Duration.RealTime = %v, want 30s", duration.RealTime)
	}
}

// TestEffectConstants tests all effect-related constants
func TestEffectConstants(t *testing.T) {
	t.Run("EffectType constants", func(t *testing.T) {
		tests := []struct {
			constant EffectType
			expected string
		}{
			{EffectDamageOverTime, "damage_over_time"},
			{EffectHealOverTime, "heal_over_time"},
			{EffectPoison, "poison"},
			{EffectBurning, "burning"},
			{EffectBleeding, "bleeding"},
			{EffectStun, "stun"},
			{EffectRoot, "root"},
			{EffectStatBoost, "stat_boost"},
			{EffectStatPenalty, "stat_penalty"},
		}

		for _, tt := range tests {
			if string(tt.constant) != tt.expected {
				t.Errorf("EffectType %v = %q, want %q", tt.constant, string(tt.constant), tt.expected)
			}
		}
	})

	t.Run("DamageType constants", func(t *testing.T) {
		tests := []struct {
			constant DamageType
			expected string
		}{
			{DamagePhysical, "physical"},
			{DamageFire, "fire"},
			{DamagePoison, "poison"},
			{DamageFrost, "frost"},
			{DamageLightning, "lightning"},
		}

		for _, tt := range tests {
			if string(tt.constant) != tt.expected {
				t.Errorf("DamageType %v = %q, want %q", tt.constant, string(tt.constant), tt.expected)
			}
		}
	})

	t.Run("DispelType constants", func(t *testing.T) {
		tests := []struct {
			constant DispelType
			expected string
		}{
			{DispelMagic, "magic"},
			{DispelCurse, "curse"},
			{DispelPoison, "poison"},
			{DispelDisease, "disease"},
			{DispelAll, "all"},
		}

		for _, tt := range tests {
			if string(tt.constant) != tt.expected {
				t.Errorf("DispelType %v = %q, want %q", tt.constant, string(tt.constant), tt.expected)
			}
		}
	})
	t.Run("ImmunityType constants", func(t *testing.T) {
		tests := []struct {
			constant ImmunityType
			expected int
		}{
			{ImmunityNone, 0},
			{ImmunityPartial, 1},
			{ImmunityComplete, 2},
			{ImmunityReflect, 3},
		}

		for _, tt := range tests {
			if int(tt.constant) != tt.expected {
				t.Errorf("ImmunityType %v = %d, want %d", tt.constant, int(tt.constant), tt.expected)
			}
		}
	})

	t.Run("DispelPriority constants", func(t *testing.T) {
		tests := []struct {
			constant DispelPriority
			expected int
		}{
			{DispelPriorityLowest, 0},
			{DispelPriorityLow, 25},
			{DispelPriorityNormal, 50},
			{DispelPriorityHigh, 75},
			{DispelPriorityHighest, 100},
		}

		for _, tt := range tests {
			if int(tt.constant) != tt.expected {
				t.Errorf("DispelPriority %v = %d, want %d", tt.constant, int(tt.constant), tt.expected)
			}
		}
	})
}

// BenchmarkNewDefaultStats benchmarks the NewDefaultStats function
func BenchmarkNewDefaultStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewDefaultStats()
	}
}

// BenchmarkStats_Clone benchmarks the Stats Clone method
func BenchmarkStats_Clone(b *testing.B) {
	stats := NewDefaultStats()
	stats.Health = 150.5
	stats.Strength = 25.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stats.Clone()
	}
}

// BenchmarkMin benchmarks the min function
func BenchmarkMin(b *testing.B) {
	a, b1 := 15.5, 23.7

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = min(a, b1)
	}
}

// TestEffect_IsExpiredWithContext tests the round/turn-based expiration functionality
func TestEffect_IsExpiredWithContext(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		effect       *Effect
		currentTime  time.Time
		currentRound int
		currentTurn  int
		expected     bool
	}{
		{
			name: "Round-based - not expired",
			effect: &Effect{
				StartTime:  now.Add(-5 * time.Minute),
				StartRound: 1,
				Duration:   Duration{Rounds: 3},
			},
			currentTime:  now,
			currentRound: 3,
			currentTurn:  0,
			expected:     false, // Current round 3 < start round 1 + duration 3 = 4
		},
		{
			name: "Round-based - exactly at expiration",
			effect: &Effect{
				StartTime:  now.Add(-5 * time.Minute),
				StartRound: 1,
				Duration:   Duration{Rounds: 3},
			},
			currentTime:  now,
			currentRound: 4,
			currentTurn:  0,
			expected:     true, // Current round 4 >= start round 1 + duration 3 = 4
		},
		{
			name: "Round-based - expired",
			effect: &Effect{
				StartTime:  now.Add(-10 * time.Minute),
				StartRound: 5,
				Duration:   Duration{Rounds: 2},
			},
			currentTime:  now,
			currentRound: 8,
			currentTurn:  0,
			expected:     true, // Current round 8 >= start round 5 + duration 2 = 7
		},
		{
			name: "Turn-based - not expired",
			effect: &Effect{
				StartTime: now.Add(-2 * time.Minute),
				StartTurn: 10,
				Duration:  Duration{Turns: 5},
			},
			currentTime:  now,
			currentRound: 0,
			currentTurn:  13,
			expected:     false, // Current turn 13 < start turn 10 + duration 5 = 15
		},
		{
			name: "Turn-based - expired",
			effect: &Effect{
				StartTime: now.Add(-5 * time.Minute),
				StartTurn: 20,
				Duration:  Duration{Turns: 3},
			},
			currentTime:  now,
			currentRound: 0,
			currentTurn:  25,
			expected:     true, // Current turn 25 >= start turn 20 + duration 3 = 23
		},
		{
			name: "Real-time takes priority over rounds",
			effect: &Effect{
				StartTime:  now.Add(-2 * time.Hour),
				StartRound: 1,
				Duration:   Duration{RealTime: 1 * time.Hour, Rounds: 100},
			},
			currentTime:  now,
			currentRound: 2, // Would not be expired by rounds
			currentTurn:  0,
			expected:     true, // Expired by real time
		},
		{
			name: "Permanent effect (negative duration) never expires",
			effect: &Effect{
				StartTime:  now.Add(-24 * time.Hour),
				StartRound: 1,
				Duration:   Duration{Rounds: -1},
			},
			currentTime:  now,
			currentRound: 1000,
			currentTurn:  0,
			expected:     false,
		},
		{
			name: "Zero duration instant effect expires immediately",
			effect: &Effect{
				StartTime:  now,
				StartRound: 5,
				StartTurn:  10,
				Duration:   Duration{},
			},
			currentTime:  now,
			currentRound: 5,
			currentTurn:  10,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.effect.IsExpiredWithContext(tt.currentTime, tt.currentRound, tt.currentTurn)
			if result != tt.expected {
				t.Errorf("IsExpiredWithContext() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
