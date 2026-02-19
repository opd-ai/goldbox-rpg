// Package server tests for process.go effect processing functions
package server

import (
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
)

// createGameStateForProcessTests creates a properly initialized GameState for process tests
func createGameStateForProcessTests() *GameState {
	return &GameState{
		WorldState: &game.World{
			Objects: make(map[string]game.GameObject),
		},
	}
}

// TestProcessEffectTick tests the processEffectTick function
func TestProcessEffectTick(t *testing.T) {
	tests := []struct {
		name        string
		effect      *game.Effect
		expectError bool
	}{
		{
			name:        "nil effect returns error",
			effect:      nil,
			expectError: true,
		},
		{
			name: "unsupported effect type returns error",
			effect: &game.Effect{
				ID:   "test_effect",
				Type: game.EffectType("invalid_type"), // Invalid type
			},
			expectError: true,
		},
		{
			name: "damage over time effect without target",
			effect: &game.Effect{
				ID:        "dot_effect",
				Type:      game.EffectDamageOverTime,
				TargetID:  "nonexistent_target",
				Magnitude: 5,
			},
			expectError: true,
		},
		{
			name: "heal over time effect without target",
			effect: &game.Effect{
				ID:        "hot_effect",
				Type:      game.EffectHealOverTime,
				TargetID:  "nonexistent_target",
				Magnitude: 5,
			},
			expectError: true,
		},
		{
			name: "stat boost effect without target",
			effect: &game.Effect{
				ID:        "stat_effect",
				Type:      game.EffectStatBoost,
				TargetID:  "nonexistent_target",
				Magnitude: 5,
			},
			expectError: true,
		},
		{
			name: "stat penalty effect without target",
			effect: &game.Effect{
				ID:        "stat_penalty",
				Type:      game.EffectStatPenalty,
				TargetID:  "nonexistent_target",
				Magnitude: -5,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := createGameStateForProcessTests()

			err := gs.processEffectTick(tt.effect)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateEffectNotNil tests the validateEffectNotNil function
func TestValidateEffectNotNil(t *testing.T) {
	gs := createGameStateForProcessTests()

	// Test with nil effect
	err := gs.validateEffectNotNil(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "effect is nil")

	// Test with valid effect
	effect := &game.Effect{ID: "test"}
	err = gs.validateEffectNotNil(effect)
	assert.NoError(t, err)
}

// TestHandleDamageOverTimeEffect tests the handleDamageOverTimeEffect function
func TestHandleDamageOverTimeEffect(t *testing.T) {
	gs := createGameStateForProcessTests()

	// Test with effect that has no valid target
	effect := &game.Effect{
		ID:        "test_dot",
		Type:      game.EffectDamageOverTime,
		TargetID:  "nonexistent",
		Magnitude: 10,
	}

	err := gs.handleDamageOverTimeEffect(effect)
	assert.Error(t, err)
}

// TestHandleHealingOverTimeEffect tests the handleHealingOverTimeEffect function
func TestHandleHealingOverTimeEffect(t *testing.T) {
	gs := createGameStateForProcessTests()

	// Test with effect that has no valid target
	effect := &game.Effect{
		ID:        "test_hot",
		Type:      game.EffectHealOverTime,
		TargetID:  "nonexistent",
		Magnitude: 10,
	}

	err := gs.handleHealingOverTimeEffect(effect)
	assert.Error(t, err)
}

// TestHandleStatModificationEffect tests the handleStatModificationEffect function
func TestHandleStatModificationEffect(t *testing.T) {
	gs := createGameStateForProcessTests()

	// Test stat boost with no valid target
	effect := &game.Effect{
		ID:        "test_stat",
		Type:      game.EffectStatBoost,
		TargetID:  "nonexistent",
		Magnitude: 5,
	}

	err := gs.handleStatModificationEffect(effect)
	assert.Error(t, err)

	// Test stat penalty with no valid target
	effect2 := &game.Effect{
		ID:        "test_stat2",
		Type:      game.EffectStatPenalty,
		TargetID:  "nonexistent",
		Magnitude: -5,
	}

	err = gs.handleStatModificationEffect(effect2)
	assert.Error(t, err)
}

// TestProcessDamageEffect tests the processDamageEffect function
func TestProcessDamageEffect(t *testing.T) {
	gs := createGameStateForProcessTests()

	// Test with effect but no valid target
	effect := &game.Effect{
		ID:        "test_damage",
		Type:      game.EffectDamageOverTime,
		TargetID:  "nonexistent",
		Magnitude: 10,
	}

	err := gs.processDamageEffect(effect)
	assert.Error(t, err)
}

// TestProcessHealEffect tests the processHealEffect function
func TestProcessHealEffect(t *testing.T) {
	gs := createGameStateForProcessTests()

	// Test with effect but no valid target
	effect := &game.Effect{
		ID:        "test_heal",
		Type:      game.EffectHealOverTime,
		TargetID:  "nonexistent",
		Magnitude: 10,
	}

	err := gs.processHealEffect(effect)
	assert.Error(t, err)
}

// TestProcessStatEffect tests the processStatEffect function
func TestProcessStatEffect(t *testing.T) {
	gs := createGameStateForProcessTests()

	// Test with effect but no valid target
	effect := &game.Effect{
		ID:        "test_stat_proc",
		Type:      game.EffectStatBoost,
		TargetID:  "nonexistent",
		Magnitude: 5,
	}

	err := gs.processStatEffect(effect)
	assert.Error(t, err)
}

// TestProcessEffectTickWithValidTarget tests effect processing with a valid target
func TestProcessEffectTickWithValidTarget(t *testing.T) {
	gs := createGameStateForProcessTests()

	// Add a character to the world
	char := &game.Character{
		ID:    "test_char",
		HP:    100,
		MaxHP: 100,
	}
	gs.WorldState.Objects["test_char"] = char

	// Test damage over time
	damageEffect := &game.Effect{
		ID:        "damage_effect",
		Type:      game.EffectDamageOverTime,
		TargetID:  "test_char",
		Magnitude: 10,
	}
	err := gs.processEffectTick(damageEffect)
	assert.NoError(t, err)
	assert.Equal(t, 90, char.HP)

	// Test heal over time
	healEffect := &game.Effect{
		ID:        "heal_effect",
		Type:      game.EffectHealOverTime,
		TargetID:  "test_char",
		Magnitude: 5,
	}
	err = gs.processEffectTick(healEffect)
	assert.NoError(t, err)
	assert.Equal(t, 95, char.HP)

	// Test stat boost
	statEffect := &game.Effect{
		ID:           "stat_effect",
		Type:         game.EffectStatBoost,
		TargetID:     "test_char",
		Magnitude:    3,
		StatAffected: "strength",
	}
	err = gs.processEffectTick(statEffect)
	assert.NoError(t, err)
}
