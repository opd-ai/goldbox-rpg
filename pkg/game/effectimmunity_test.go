package game

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestEffectManager_initializeDefaultImmunities tests the initialization of default immunities
func TestEffectManager_initializeDefaultImmunities(t *testing.T) {
	em := NewEffectManager(NewDefaultStats())

	// Check if default immunities are properly set
	immunity := em.CheckImmunity(EffectPoison)

	if immunity.Type != ImmunityPartial {
		t.Errorf("Expected poison immunity type to be ImmunityPartial, got %v", immunity.Type)
	}

	if immunity.Resistance != 0.25 {
		t.Errorf("Expected poison resistance to be 0.25, got %v", immunity.Resistance)
	}

	if immunity.Duration != 0 {
		t.Errorf("Expected poison immunity duration to be 0, got %v", immunity.Duration)
	}

	if !immunity.ExpiresAt.IsZero() {
		t.Errorf("Expected poison immunity ExpiresAt to be zero time, got %v", immunity.ExpiresAt)
	}
}

// TestEffectManager_AddImmunity tests adding immunities with various configurations
func TestEffectManager_AddImmunity(t *testing.T) {
	tests := []struct {
		name         string
		effectType   EffectType
		immunity     ImmunityData
		expectedTemp bool
	}{
		{
			name:       "Permanent complete immunity",
			effectType: EffectBurning,
			immunity: ImmunityData{
				Type:       ImmunityComplete,
				Duration:   0,
				Resistance: 1.0,
				ExpiresAt:  time.Time{},
			},
			expectedTemp: false,
		},
		{
			name:       "Temporary partial immunity",
			effectType: EffectStun,
			immunity: ImmunityData{
				Type:       ImmunityPartial,
				Duration:   30 * time.Second,
				Resistance: 0.5,
				ExpiresAt:  time.Time{},
			},
			expectedTemp: true,
		},
		{
			name:       "Permanent reflection immunity",
			effectType: EffectRoot,
			immunity: ImmunityData{
				Type:       ImmunityReflect,
				Duration:   0,
				Resistance: 0,
				ExpiresAt:  time.Time{},
			},
			expectedTemp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := NewEffectManager(NewDefaultStats())

			em.AddImmunity(tt.effectType, tt.immunity)

			// Check immunity was added correctly
			immunity := em.CheckImmunity(tt.effectType)

			if immunity.Type != tt.immunity.Type {
				t.Errorf("Expected immunity type %v, got %v", tt.immunity.Type, immunity.Type)
			}

			if immunity.Resistance != tt.immunity.Resistance {
				t.Errorf("Expected resistance %v, got %v", tt.immunity.Resistance, immunity.Resistance)
			}

			// For temporary immunities, check ExpiresAt was set
			if tt.expectedTemp && immunity.ExpiresAt.IsZero() {
				t.Errorf("Expected ExpiresAt to be set for temporary immunity")
			}

			// For permanent immunities, check ExpiresAt is zero
			if !tt.expectedTemp && !immunity.ExpiresAt.IsZero() {
				t.Errorf("Expected ExpiresAt to be zero for permanent immunity")
			}
		})
	}
}

// TestEffectManager_CheckImmunity tests checking immunity for various scenarios
func TestEffectManager_CheckImmunity(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*EffectManager)
		effectType     EffectType
		expectedType   ImmunityType
		expectedResist float64
		checkExpired   bool
	}{
		{
			name: "No immunity exists",
			setupFunc: func(em *EffectManager) {
				// No setup needed
			},
			effectType:     EffectBleeding,
			expectedType:   ImmunityNone,
			expectedResist: 0,
		},
		{
			name: "Permanent immunity exists",
			setupFunc: func(em *EffectManager) {
				em.AddImmunity(EffectBurning, ImmunityData{
					Type:       ImmunityComplete,
					Duration:   0,
					Resistance: 1.0,
				})
			},
			effectType:     EffectBurning,
			expectedType:   ImmunityComplete,
			expectedResist: 1.0,
		},
		{
			name: "Temporary immunity active",
			setupFunc: func(em *EffectManager) {
				em.AddImmunity(EffectStun, ImmunityData{
					Type:       ImmunityPartial,
					Duration:   60 * time.Second,
					Resistance: 0.75,
				})
			},
			effectType:     EffectStun,
			expectedType:   ImmunityPartial,
			expectedResist: 0.75,
		},
		{
			name: "Temporary immunity expired",
			setupFunc: func(em *EffectManager) {
				em.AddImmunity(EffectRoot, ImmunityData{
					Type:       ImmunityPartial,
					Duration:   1 * time.Nanosecond, // Very short duration
					Resistance: 0.8,
				})
				time.Sleep(2 * time.Nanosecond) // Wait for expiration
			},
			effectType:     EffectRoot,
			expectedType:   ImmunityNone,
			expectedResist: 0,
			checkExpired:   true,
		},
		{
			name: "Default poison immunity",
			setupFunc: func(em *EffectManager) {
				// Use default immunities from initialization
			},
			effectType:     EffectPoison,
			expectedType:   ImmunityPartial,
			expectedResist: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := NewEffectManager(NewDefaultStats())
			tt.setupFunc(em)

			immunity := em.CheckImmunity(tt.effectType)

			if immunity.Type != tt.expectedType {
				t.Errorf("Expected immunity type %v, got %v", tt.expectedType, immunity.Type)
			}

			if immunity.Resistance != tt.expectedResist {
				t.Errorf("Expected resistance %v, got %v", tt.expectedResist, immunity.Resistance)
			}
		})
	}
}

// TestEffectManager_DispelEffects tests effect dispelling functionality
func TestEffectManager_DispelEffects(t *testing.T) {
	t.Run("Dispel effects by type and priority", func(t *testing.T) {
		em := NewEffectManager(NewDefaultStats())

		// Create test effects with different priorities
		highPriorityEffect := NewEffectWithDispel(
			EffectStatBoost,
			Duration{RealTime: 60 * time.Second},
			10,
			DispelInfo{
				Priority:  DispelPriorityHigh,
				Types:     []DispelType{DispelMagic},
				Removable: true,
			},
		)

		lowPriorityEffect := NewEffectWithDispel(
			EffectStatPenalty,
			Duration{RealTime: 60 * time.Second},
			-5,
			DispelInfo{
				Priority:  DispelPriorityLow,
				Types:     []DispelType{DispelMagic},
				Removable: true,
			},
		)

		nonRemovableEffect := NewEffectWithDispel(
			EffectStun,
			Duration{RealTime: 60 * time.Second},
			1,
			DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelMagic},
				Removable: false,
			},
		)

		// Add effects to manager
		em.activeEffects["high"] = highPriorityEffect
		em.activeEffects["low"] = lowPriorityEffect
		em.activeEffects["nonremovable"] = nonRemovableEffect

		// Dispel one magic effect
		removed := em.DispelEffects(DispelMagic, 1)

		// Should remove the highest priority effect first
		if len(removed) != 1 {
			t.Errorf("Expected 1 effect removed, got %d", len(removed))
		}

		if removed[0] != "high" {
			t.Errorf("Expected 'high' priority effect to be removed first, got %s", removed[0])
		}

		// Verify the high priority effect was removed
		if _, exists := em.activeEffects["high"]; exists {
			t.Errorf("High priority effect should have been removed")
		}

		// Verify other effects still exist
		if _, exists := em.activeEffects["low"]; !exists {
			t.Errorf("Low priority effect should still exist")
		}

		if _, exists := em.activeEffects["nonremovable"]; !exists {
			t.Errorf("Non-removable effect should still exist")
		}
	})

	t.Run("Dispel multiple effects", func(t *testing.T) {
		em := NewEffectManager(NewDefaultStats())

		// Add multiple removable effects
		for i := 0; i < 3; i++ {
			effect := NewEffectWithDispel(
				EffectStatBoost,
				Duration{RealTime: 60 * time.Second},
				float64(i+1),
				DispelInfo{
					Priority:  DispelPriorityNormal,
					Types:     []DispelType{DispelCurse},
					Removable: true,
				},
			)
			em.activeEffects[fmt.Sprintf("effect%d", i)] = effect
		}

		// Dispel all curse effects
		removed := em.DispelEffects(DispelCurse, 5) // Request more than available

		if len(removed) != 3 {
			t.Errorf("Expected 3 effects removed, got %d", len(removed))
		}

		if len(em.activeEffects) != 0 {
			t.Errorf("Expected all effects to be removed, %d remain", len(em.activeEffects))
		}
	})

	t.Run("Dispel with DispelAll", func(t *testing.T) {
		em := NewEffectManager(NewDefaultStats())

		// Add effects with different dispel types
		magicEffect := NewEffectWithDispel(
			EffectStatBoost,
			Duration{RealTime: 60 * time.Second},
			10,
			DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelMagic},
				Removable: true,
			},
		)

		curseEffect := NewEffectWithDispel(
			EffectStatPenalty,
			Duration{RealTime: 60 * time.Second},
			-5,
			DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelCurse},
				Removable: true,
			},
		)

		em.activeEffects["magic"] = magicEffect
		em.activeEffects["curse"] = curseEffect

		// Dispel all effects
		removed := em.DispelEffects(DispelAll, 10)

		if len(removed) != 2 {
			t.Errorf("Expected 2 effects removed, got %d", len(removed))
		}

		if len(em.activeEffects) != 0 {
			t.Errorf("Expected all effects to be removed")
		}
	})
}

// TestNewEffectWithDispel tests creating effects with dispel information
func TestNewEffectWithDispel(t *testing.T) {
	tests := []struct {
		name       string
		effectType EffectType
		duration   Duration
		magnitude  float64
		dispelInfo DispelInfo
	}{
		{
			name:       "Stat boost with high priority dispel",
			effectType: EffectStatBoost,
			duration:   Duration{RealTime: 30 * time.Second},
			magnitude:  15.0,
			dispelInfo: DispelInfo{
				Priority:  DispelPriorityHigh,
				Types:     []DispelType{DispelMagic, DispelCurse},
				Removable: true,
			},
		},
		{
			name:       "Poison effect with poison dispel",
			effectType: EffectPoison,
			duration:   Duration{RealTime: 60 * time.Second},
			magnitude:  5.0,
			dispelInfo: DispelInfo{
				Priority:  DispelPriorityLow,
				Types:     []DispelType{DispelPoison},
				Removable: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := NewEffectWithDispel(tt.effectType, tt.duration, tt.magnitude, tt.dispelInfo)

			if effect.Type != tt.effectType {
				t.Errorf("Expected effect type %v, got %v", tt.effectType, effect.Type)
			}

			if effect.Magnitude != tt.magnitude {
				t.Errorf("Expected magnitude %v, got %v", tt.magnitude, effect.Magnitude)
			}

			if !reflect.DeepEqual(effect.DispelInfo, tt.dispelInfo) {
				t.Errorf("Expected dispel info %+v, got %+v", tt.dispelInfo, effect.DispelInfo)
			}

			if effect.Duration != tt.duration {
				t.Errorf("Expected duration %+v, got %+v", tt.duration, effect.Duration)
			}
		})
	}
}

// TestCreatePoisonEffectWithDispel tests creating poison effects with dispel properties
func TestCreatePoisonEffectWithDispel(t *testing.T) {
	tests := []struct {
		name       string
		baseDamage float64
		duration   time.Duration
	}{
		{
			name:       "Low damage poison",
			baseDamage: 2.5,
			duration:   15 * time.Second,
		},
		{
			name:       "High damage poison",
			baseDamage: 10.0,
			duration:   60 * time.Second,
		},
		{
			name:       "Quick poison",
			baseDamage: 5.0,
			duration:   5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poisonEffect := CreatePoisonEffectWithDispel(tt.baseDamage, tt.duration)

			// Check it's a damage effect
			if poisonEffect.Effect.Type != EffectPoison {
				t.Errorf("Expected effect type EffectPoison, got %v", poisonEffect.Effect.Type)
			}

			// Check base damage
			if poisonEffect.BaseDamage != tt.baseDamage {
				t.Errorf("Expected base damage %v, got %v", tt.baseDamage, poisonEffect.BaseDamage)
			}

			// Check duration
			if poisonEffect.Effect.Duration.RealTime != tt.duration {
				t.Errorf("Expected duration %v, got %v", tt.duration, poisonEffect.Effect.Duration.RealTime)
			}

			// Check dispel info
			if poisonEffect.Effect.DispelInfo.Priority != DispelPriorityNormal {
				t.Errorf("Expected dispel priority Normal, got %v", poisonEffect.Effect.DispelInfo.Priority)
			}

			expectedTypes := []DispelType{DispelPoison, DispelMagic}
			if !reflect.DeepEqual(poisonEffect.Effect.DispelInfo.Types, expectedTypes) {
				t.Errorf("Expected dispel types %v, got %v", expectedTypes, poisonEffect.Effect.DispelInfo.Types)
			}

			if !poisonEffect.Effect.DispelInfo.Removable {
				t.Errorf("Expected poison effect to be removable")
			}
		})
	}
}

// TestEffectManager_ApplyEffect tests applying effects with immunity considerations
func TestEffectManager_ApplyEffect(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*EffectManager)
		effect         *Effect
		expectError    bool
		errorContains  string
		checkMagnitude bool
		expectedMag    float64
	}{
		{
			name: "Apply effect with no immunity",
			setupFunc: func(em *EffectManager) {
				// No immunity setup
			},
			effect:      NewEffect(EffectStatBoost, Duration{RealTime: 30 * time.Second}, 10.0),
			expectError: false,
		},
		{
			name: "Apply effect with complete immunity",
			setupFunc: func(em *EffectManager) {
				em.AddImmunity(EffectBurning, ImmunityData{
					Type:       ImmunityComplete,
					Duration:   0,
					Resistance: 1.0,
				})
			},
			effect:        NewEffect(EffectBurning, Duration{RealTime: 30 * time.Second}, 5.0),
			expectError:   true,
			errorContains: "immune to burning effects",
		},
		{
			name: "Apply effect with partial immunity",
			setupFunc: func(em *EffectManager) {
				em.AddImmunity(EffectStatPenalty, ImmunityData{
					Type:       ImmunityPartial,
					Duration:   0,
					Resistance: 0.5, // 50% resistance
				})
			},
			effect:         NewEffect(EffectStatPenalty, Duration{RealTime: 30 * time.Second}, -10.0),
			expectError:    false,
			checkMagnitude: true,
			expectedMag:    -5.0, // 50% of original magnitude
		},
		{
			name: "Apply effect with reflection immunity",
			setupFunc: func(em *EffectManager) {
				em.AddImmunity(EffectStun, ImmunityData{
					Type:       ImmunityReflect,
					Duration:   0,
					Resistance: 0,
				})
			},
			effect:        NewEffect(EffectStun, Duration{RealTime: 10 * time.Second}, 1.0),
			expectError:   true,
			errorContains: "effect reflected",
		},
		{
			name: "Apply effect with unknown immunity type",
			setupFunc: func(em *EffectManager) {
				// Manually add an immunity with an invalid type
				em.AddImmunity(EffectPoison, ImmunityData{
					Type:       ImmunityType(999), // Invalid immunity type
					Duration:   0,
					Resistance: 0,
				})
			},
			effect:        NewEffect(EffectPoison, Duration{RealTime: 10 * time.Second}, 5.0),
			expectError:   true,
			errorContains: "unknown immunity type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := NewEffectManager(NewDefaultStats())
			tt.setupFunc(em)

			originalMagnitude := tt.effect.Magnitude
			err := em.ApplyEffect(tt.effect)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if tt.checkMagnitude {
				if tt.effect.Magnitude != tt.expectedMag {
					t.Errorf("Expected effect magnitude to be %v after immunity, got %v", tt.expectedMag, tt.effect.Magnitude)
				}
			} else if !tt.expectError {
				// For successful applications without immunity checks, magnitude shouldn't change
				if tt.effect.Magnitude != originalMagnitude {
					t.Errorf("Expected effect magnitude to remain %v, got %v", originalMagnitude, tt.effect.Magnitude)
				}
			}
		})
	}
}

// TestExampleEffectDispel tests the example function to ensure it works correctly
func TestExampleEffectDispel(t *testing.T) {
	// This test ensures the example function runs without panics
	// and demonstrates expected usage patterns
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ExampleEffectDispel panicked: %v", r)
		}
	}()

	ExampleEffectDispel()
}

// TestExampleEffectDispelWithLogging tests that the example function properly
// logs errors when effects fail to apply (e.g., due to immunity)
func TestExampleEffectDispelWithLogging(t *testing.T) {
	// Capture log output using a custom buffer
	var logBuf strings.Builder
	customLogger := log.New(&logBuf, "[GAME] ", log.LstdFlags)

	// Save original logger and restore after test
	originalLogger := getLogger()
	SetLogger(customLogger)
	defer SetLogger(originalLogger)

	// Run the example - it should log the dispel result
	ExampleEffectDispel()

	logOutput := logBuf.String()

	// Verify that the dispel result was logged
	if !strings.Contains(logOutput, "dispelled") {
		t.Errorf("Expected log output to contain dispel result, got: %s", logOutput)
	}
}

// Additional helper tests

// TestEffectManager_ThreadSafety tests concurrent access to immunity methods
func TestEffectManager_ThreadSafety(t *testing.T) {
	em := NewEffectManager(NewDefaultStats())

	// Run multiple goroutines concurrently
	done := make(chan bool, 4)

	// Goroutine 1: Add immunities
	go func() {
		for i := 0; i < 100; i++ {
			em.AddImmunity(EffectBurning, ImmunityData{
				Type:       ImmunityPartial,
				Duration:   time.Duration(i) * time.Millisecond,
				Resistance: 0.1,
			})
		}
		done <- true
	}()

	// Goroutine 2: Check immunities
	go func() {
		for i := 0; i < 100; i++ {
			em.CheckImmunity(EffectBurning)
		}
		done <- true
	}()

	// Goroutine 3: Dispel effects
	go func() {
		for i := 0; i < 100; i++ {
			em.DispelEffects(DispelMagic, 1)
		}
		done <- true
	}()

	// Goroutine 4: Apply effects
	go func() {
		for i := 0; i < 100; i++ {
			effect := NewEffect(EffectStatBoost, Duration{RealTime: time.Second}, 1.0)
			em.ApplyEffect(effect)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Test passes if no race conditions occurred
}
