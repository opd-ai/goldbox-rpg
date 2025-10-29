package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCharacterCreation tests character creation workflows
func TestCharacterCreation(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	// Create session first
	sessionID, err := client.JoinGame("TestPlayer")
	require.NoError(t, err)

	testCases := []struct {
		name          string
		charName      string
		charClass     string
		expectError   bool
		errorContains string
	}{
		{
			name:        "create_fighter",
			charName:    "Aldric",
			charClass:   "fighter",
			expectError: false,
		},
		{
			name:        "create_mage",
			charName:    "Eldrin",
			charClass:   "mage",
			expectError: false,
		},
		{
			name:        "create_cleric",
			charName:    "Helena",
			charClass:   "cleric",
			expectError: false,
		},
		{
			name:          "create_with_invalid_class",
			charName:      "Invalid",
			charClass:     "ninja",
			expectError:   true,
			errorContains: "class",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			charID, err := client.CreateCharacter(sessionID, tc.charName, tc.charClass)

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					ErrorContains(t, err, tc.errorContains)
				}
			} else {
				require.NoError(t, err, "should create character successfully")
				AssertCharacterID(t, charID)
			}
		})
	}
}

// TestCharacterAttributes tests that characters have valid attributes
func TestCharacterAttributes(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()
	sessionID, charID := helper.CreateSession()

	// Get game state to check character
	state, err := client.GetGameState(sessionID)
	require.NoError(t, err)

	// Find character in state
	player, ok := state["player"].(map[string]interface{})
	require.True(t, ok, "should have player in state")

	character, ok := player["character"].(map[string]interface{})
	require.True(t, ok, "should have character in player")

	// Verify character ID matches
	stateCharID, ok := character["id"].(string)
	require.True(t, ok, "character should have ID")
	assert.Equal(t, charID, stateCharID, "character ID should match")

	// Verify attributes are within valid ranges
	attributes := []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}
	for _, attr := range attributes {
		value, ok := character[attr].(float64)
		require.True(t, ok, "attribute %s should be numeric", attr)
		assert.GreaterOrEqual(t, value, float64(3), "attribute %s should be at least 3", attr)
		assert.LessOrEqual(t, value, float64(18), "attribute %s should be at most 18", attr)
	}

	// Verify HP
	currentHP, ok := character["current_hp"].(float64)
	require.True(t, ok, "should have current HP")
	assert.Greater(t, currentHP, float64(0), "current HP should be positive")

	maxHP, ok := character["max_hp"].(float64)
	require.True(t, ok, "should have max HP")
	assert.Greater(t, maxHP, float64(0), "max HP should be positive")
	assert.LessOrEqual(t, currentHP, maxHP, "current HP should not exceed max HP")

	// Verify level
	level, ok := character["level"].(float64)
	require.True(t, ok, "should have level")
	assert.Equal(t, float64(1), level, "new character should be level 1")
}

// TestCharacterWithoutSession tests that character operations require valid session
func TestCharacterWithoutSession(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	// Try to create character without session
	_, err := client.CreateCharacter("invalid-session", "TestChar", "fighter")
	require.Error(t, err, "should fail without valid session")
	ErrorContains(t, err, "session")
}

// TestMultipleCharactersPerSession tests creating multiple characters in same session
func TestMultipleCharactersPerSession(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("TestPlayer")
	require.NoError(t, err)

	// Create first character
	char1ID, err := client.CreateCharacter(sessionID, "Character1", "fighter")
	require.NoError(t, err)
	AssertCharacterID(t, char1ID)

	// Create second character (this may or may not be allowed based on game rules)
	// This test documents current behavior
	_, err = client.CreateCharacter(sessionID, "Character2", "mage")
	// Note: behavior depends on game implementation
	// Some games allow multiple characters, others don't
	if err != nil {
		t.Logf("Multiple characters per session not allowed: %v", err)
	}
}
