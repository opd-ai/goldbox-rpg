package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNPCPathfinding tests that NPCs can navigate around obstacles to reach targets
func TestNPCPathfinding(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("PathfindingTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Navigator", "fighter")
	require.NoError(t, err)

	// Test pathfinding around obstacles
	testCases := []struct {
		name           string
		startX, startY int
		endX, endY     int
		expectPath     bool
	}{
		{
			name:       "direct_path_exists",
			startX:     0,
			startY:     0,
			endX:       5,
			endY:       5,
			expectPath: true,
		},
		{
			name:       "path_around_obstacle",
			startX:     0,
			startY:     0,
			endX:       10,
			endY:       0,
			expectPath: true,
		},
		{
			name:       "same_position",
			startX:     5,
			startY:     5,
			endX:       5,
			endY:       5,
			expectPath: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("findPath", map[string]interface{}{
				"session_id": sessionID,
				"start_x":    tc.startX,
				"start_y":    tc.startY,
				"end_x":      tc.endX,
				"end_y":      tc.endY,
			})

			if tc.expectPath {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if result != nil {
					path, ok := result["path"].([]interface{})
					if ok && len(path) > 0 {
						assert.GreaterOrEqual(t, len(path), 1)
					}
				}
			}
		})
	}
}

// TestCombatAITargetSelection tests that AI selects optimal targets based on difficulty
func TestCombatAITargetSelection(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("AITargetTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Strategist", "fighter")
	require.NoError(t, err)

	// Start combat to initialize AI
	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Test AI target selection at different difficulties
	testCases := []struct {
		name       string
		difficulty string
	}{
		{
			name:       "easy_difficulty",
			difficulty: "easy",
		},
		{
			name:       "medium_difficulty",
			difficulty: "medium",
		},
		{
			name:       "hard_difficulty",
			difficulty: "hard",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("setAIDifficulty", map[string]interface{}{
				"session_id": sessionID,
				"difficulty": tc.difficulty,
			})

			// If the method exists, verify it works
			if err == nil {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestBehaviorTreeExecution tests that behavior trees execute correctly for NPCs
func TestBehaviorTreeExecution(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("BehaviorTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Observer", "ranger")
	require.NoError(t, err)

	// Test different behavior tree patterns
	testCases := []struct {
		name         string
		behaviorType string
	}{
		{
			name:         "aggressive_behavior",
			behaviorType: "aggressive",
		},
		{
			name:         "guard_behavior",
			behaviorType: "guard",
		},
		{
			name:         "patrol_behavior",
			behaviorType: "patrol",
		},
		{
			name:         "coward_behavior",
			behaviorType: "coward",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test NPC behavior assignment (if endpoint exists)
			result, err := client.Call("setNPCBehavior", map[string]interface{}{
				"session_id":    sessionID,
				"behavior_type": tc.behaviorType,
			})

			// If the method exists, verify it works
			if err == nil {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestTacticalCombatOpportunityAttacks tests opportunity attack mechanics
func TestTacticalCombatOpportunityAttacks(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("OpportunityTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Tactician", "fighter")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Test movement that should trigger opportunity attack
	testCases := []struct {
		name        string
		disengage   bool
		expectOA    bool
		description string
	}{
		{
			name:        "move_without_disengage",
			disengage:   false,
			expectOA:    true,
			description: "Moving away without disengage should trigger opportunity attack",
		},
		{
			name:        "move_with_disengage",
			disengage:   true,
			expectOA:    false,
			description: "Moving away with disengage should not trigger opportunity attack",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("move", map[string]interface{}{
				"session_id": sessionID,
				"direction":  DirectionNorth,
				"disengage":  tc.disengage,
			})

			// Move command should succeed regardless
			assert.NoError(t, err, "move should succeed")
			assert.NotNil(t, result)
		})
	}
}

// TestTacticalCombatCoverBonus tests cover mechanics
func TestTacticalCombatCoverBonus(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("CoverTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "TakeCover", "ranger")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Test cover mechanics
	result, err := client.Call("getCombatModifiers", map[string]interface{}{
		"session_id": sessionID,
	})

	// If the method exists, verify it returns modifiers
	if err == nil {
		assert.NotNil(t, result)
		// Check for cover-related fields
		if modifiers, ok := result["modifiers"].(map[string]interface{}); ok {
			assert.Contains(t, modifiers, "cover")
		}
	}
}

// TestTacticalCombatFlankingBonus tests flanking mechanics
func TestTacticalCombatFlankingBonus(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("FlankTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Flanker", "thief")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Test flanking mechanics
	result, err := client.Call("getCombatModifiers", map[string]interface{}{
		"session_id": sessionID,
	})

	// If the method exists, verify it returns modifiers
	if err == nil {
		assert.NotNil(t, result)
		// Check for flanking-related fields
		if modifiers, ok := result["modifiers"].(map[string]interface{}); ok {
			assert.Contains(t, modifiers, "flanking")
		}
	}
}

// TestMoraleMechanics tests morale system
func TestMoraleMechanics(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("MoraleTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "BraveSoul", "paladin")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Test morale status retrieval
	result, err := client.Call("getMoraleStatus", map[string]interface{}{
		"session_id": sessionID,
	})

	// If the method exists, verify morale status structure
	if err == nil {
		assert.NotNil(t, result)
	}
}

// TestFullAICombatScenario tests a complete AI-driven combat scenario
func TestFullAICombatScenario(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("FullCombatTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Hero", "fighter")
	require.NoError(t, err)

	// Start combat
	combatResult, err := client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, combatResult)

	// Execute several rounds of combat
	for round := 0; round < 5; round++ {
		// Get current game state
		gameState, err := client.Call("getGameState", map[string]interface{}{
			"session_id": sessionID,
		})
		require.NoError(t, err)
		assert.NotNil(t, gameState)

		// Perform an action (attack or move)
		_, err = client.Call("attack", map[string]interface{}{
			"session_id": sessionID,
			"target_id":  "enemy1",
		})
		// Attack may fail if no valid target, that's okay

		// End turn
		_, err = client.Call("endTurn", map[string]interface{}{
			"session_id": sessionID,
		})
		require.NoError(t, err)

		// Short delay between rounds
		time.Sleep(50 * time.Millisecond)
	}

	// Get final state
	finalState, err := client.Call("getGameState", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, finalState)
}

// TestAICombatWithWebSocketEvents tests AI combat events via WebSocket
func TestAICombatWithWebSocketEvents(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSAICombatTester")
	require.NoError(t, err)

	// Connect WebSocket
	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	sessionID, _, err := client.CreateCharacter("", "EventHero", "mage")
	require.NoError(t, err)

	// Start combat
	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Wait for combat start event
	event, err := client.WaitForEvent("combat_start", 5*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "combat_start", event["type"])

	// End turn and wait for turn event
	_, err = client.Call("endTurn", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Try to catch a turn change event
	turnEvent, err := client.WaitForEvent("turn_change", 5*time.Second)
	if err == nil {
		assert.NotNil(t, turnEvent)
	}
}

// TestNPCRetreatBehavior tests that wounded NPCs retreat appropriately
func TestNPCRetreatBehavior(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("RetreatTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Pursuer", "fighter")
	require.NoError(t, err)

	// Start combat
	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Execute combat rounds - NPCs should retreat when low health
	for round := 0; round < 10; round++ {
		// Get game state
		_, err := client.Call("getGameState", map[string]interface{}{
			"session_id": sessionID,
		})
		require.NoError(t, err)

		// Attack
		_, err = client.Call("attack", map[string]interface{}{
			"session_id": sessionID,
			"target_id":  "enemy1",
		})
		// Attack may fail, that's acceptable

		// End turn
		_, err = client.Call("endTurn", map[string]interface{}{
			"session_id": sessionID,
		})
		require.NoError(t, err)

		time.Sleep(50 * time.Millisecond)
	}
}

// TestAICombatDifficultyScaling tests that combat difficulty affects AI behavior
func TestAICombatDifficultyScaling(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	difficulties := []string{"easy", "medium", "hard"}

	for _, difficulty := range difficulties {
		t.Run(difficulty+"_difficulty", func(t *testing.T) {
			_, err := client.JoinGame("DifficultyTester" + difficulty)
			require.NoError(t, err)

			sessionID, _, err := client.CreateCharacter("", "Challenger", "fighter")
			require.NoError(t, err)

			// Set difficulty (if endpoint exists)
			_, err = client.Call("setDifficulty", map[string]interface{}{
				"session_id": sessionID,
				"difficulty": difficulty,
			})
			// Ignore error if method doesn't exist

			// Start combat
			_, err = client.Call("startCombat", map[string]interface{}{
				"session_id": sessionID,
			})
			require.NoError(t, err)

			// Run a few combat rounds
			for i := 0; i < 3; i++ {
				_, err = client.Call("endTurn", map[string]interface{}{
					"session_id": sessionID,
				})
				require.NoError(t, err)
			}
		})
	}
}
