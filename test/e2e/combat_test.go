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

	_, err := client.JoinGame("CombatTester")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "Warrior", "fighter")
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

	_, err := client.JoinGame("Attacker")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Fighter", "fighter")
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
			name:          "attack_without_target",
			targetID:      "",
			expectError:   true,
			errorContains: "target", // "target ID cannot be empty" from validation
		},
		{
			name:          "attack_invalid_target",
			targetID:      "nonexistent",
			expectError:   true,
			errorContains: "invalid target", // From processCombatAction
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

	_, err := client.JoinGame("TurnPlayer")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Mage", "mage")
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

	_, err := client.JoinGame("EffectTester")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "Cleric", "cleric")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	testCases := []struct {
		name       string
		effectType string
		rounds     int
		magnitude  int
	}{
		{
			name:       "apply_stun_effect",
			effectType: "stun",
			rounds:     2,
			magnitude:  1,
		},
		{
			name:       "apply_burning_effect",
			effectType: "burning",
			rounds:     3,
			magnitude:  5,
		},
		{
			name:       "apply_poison_effect",
			effectType: "poison",
			rounds:     4,
			magnitude:  3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// The applyEffect handler expects duration as a game.Duration struct
			// with fields: duration_rounds, duration_turns, duration_real
			result, err := client.Call("applyEffect", map[string]interface{}{
				"session_id":  sessionID,
				"target_id":   charID, // Apply effect to the character itself for testing
				"effect_type": tc.effectType,
				"duration": map[string]interface{}{
					"duration_rounds": tc.rounds,
					"duration_turns":  0,
					"duration_real":   0,
				},
				"magnitude": tc.magnitude,
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

	_, err := client.JoinGame("WSCombatTester")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	sessionID, _, err := client.CreateCharacter("", "Ranger", "ranger")
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
// Note: This test validates the single-player combat flow without hostile opponents.
func TestCombatSequence(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("SequenceTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Paladin", "paladin")
	require.NoError(t, err)

	combatResult, err := client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, combatResult)

	// Verify combat started correctly
	if cr, ok := combatResult["combat_state"].(map[string]interface{}); ok {
		assert.Equal(t, true, cr["is_in_combat"])
	}

	// Get initial game state during combat
	gameState, err := client.Call("getGameState", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, gameState)

	// Verify combat state before ending turn
	if turns, ok := gameState["turns"].(map[string]interface{}); ok {
		assert.Equal(t, true, turns["in_combat"])
	}

	// End the first turn - validates turn progression works
	result, err := client.Call("endTurn", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "next_turn")
	assert.Contains(t, result, "success")
}
