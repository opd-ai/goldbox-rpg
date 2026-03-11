package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCombatInitiation tests starting combat scenarios
func TestCombatInitiation(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("CombatTester")
	require.NoError(t, err)

	charID, err := client.CreateCharacter(sessionID, "Warrior", "fighter")
	require.NoError(t, err)
	assert.NotEmpty(t, charID)

	result, err := client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "combat_state")
}

// TestAttackAction tests basic attack mechanics
func TestAttackAction(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("Attacker")
	require.NoError(t, err)

	_, err = client.CreateCharacter(sessionID, "Fighter", "fighter")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	testCases := []struct {
		name          string
		targetID      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "attack_with_valid_target",
			targetID:    "enemy1",
			expectError: false,
		},
		{
			name:          "attack_without_target",
			targetID:      "",
			expectError:   true,
			errorContains: "target",
		},
		{
			name:          "attack_invalid_target",
			targetID:      "nonexistent",
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("attack", map[string]interface{}{
				"session_id": sessionID,
				"target_id":  tc.targetID,
			})

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestTurnBasedCombat tests turn progression in combat
func TestTurnBasedCombat(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("TurnPlayer")
	require.NoError(t, err)

	_, err = client.CreateCharacter(sessionID, "Mage", "mage")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	gameState, err := client.Call("getGameState", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, gameState)

	result, err := client.Call("endTurn", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "next_turn")
}

// TestCombatEffects tests effect application during combat
func TestCombatEffects(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("EffectTester")
	require.NoError(t, err)

	charID, err := client.CreateCharacter(sessionID, "Cleric", "cleric")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	testCases := []struct {
		name       string
		effectType string
		duration   int
		magnitude  int
	}{
		{
			name:       "apply_stun_effect",
			effectType: "stun",
			duration:   2,
			magnitude:  1,
		},
		{
			name:       "apply_burning_effect",
			effectType: "burning",
			duration:   3,
			magnitude:  5,
		},
		{
			name:       "apply_poison_effect",
			effectType: "poison",
			duration:   4,
			magnitude:  3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("applyEffect", map[string]interface{}{
				"session_id":   sessionID,
				"character_id": charID,
				"effect_type":  tc.effectType,
				"duration":     tc.duration,
				"magnitude":    tc.magnitude,
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestCombatWithWebSocketEvents tests real-time combat events
func TestCombatWithWebSocketEvents(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("WSCombatTester")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	_, err = client.CreateCharacter(sessionID, "Ranger", "ranger")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	event, err := client.WaitForEvent("combat_start", 5*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "combat_start", event["type"])
}

// TestCombatSequence tests a full combat sequence
func TestCombatSequence(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("SequenceTester")
	require.NoError(t, err)

	_, err = client.CreateCharacter(sessionID, "Paladin", "paladin")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		gameState, err := client.Call("getGameState", map[string]interface{}{
			"session_id": sessionID,
		})
		require.NoError(t, err)
		assert.NotNil(t, gameState)

		_, err = client.Call("endTurn", map[string]interface{}{
			"session_id": sessionID,
		})
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
	}
}
