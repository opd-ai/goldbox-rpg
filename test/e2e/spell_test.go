package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpellRetrieval tests getting spell information
func TestSpellRetrieval(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("SpellReader")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		method     string
		params     map[string]interface{}
		expectData bool
	}{
		{
			name:   "get_all_spells",
			method: "getAllSpells",
			params: map[string]interface{}{
				"session_id": sessionID,
			},
			expectData: true,
		},
		{
			name:   "get_spells_by_level",
			method: "getSpellsByLevel",
			params: map[string]interface{}{
				"session_id": sessionID,
				"level":      1,
			},
			expectData: true,
		},
		{
			name:   "get_spells_by_school",
			method: "getSpellsBySchool",
			params: map[string]interface{}{
				"session_id": sessionID,
				"school":     "evocation",
			},
			expectData: true,
		},
		{
			name:   "search_spells",
			method: "searchSpells",
			params: map[string]interface{}{
				"session_id": sessionID,
				"query":      "fire",
			},
			expectData: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call(tc.method, tc.params)
			require.NoError(t, err)
			if tc.expectData {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestSpellCasting tests casting spells in various scenarios
func TestSpellCasting(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("SpellCaster")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Wizard", "mage")
	require.NoError(t, err)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	testCases := []struct {
		name          string
		spellID       string
		targetID      string
		expectError   bool
		errorContains string
	}{
		{
			// Character doesn't know this spell yet
			name:          "cast_unknown_spell",
			spellID:       "magic_missile",
			targetID:      "enemy1",
			expectError:   true,
			errorContains: "spell",
		},
		{
			// Character doesn't know this spell yet
			name:          "cast_unknown_spell_no_target",
			spellID:       "fireball",
			targetID:      "",
			expectError:   true,
			errorContains: "spell",
		},
		{
			name:          "cast_invalid_spell",
			spellID:       "nonexistent_spell",
			targetID:      "enemy1",
			expectError:   true,
			errorContains: "spell",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("castSpell", map[string]interface{}{
				"session_id": sessionID,
				"spell_id":   tc.spellID,
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

// TestSpellSlots tests spell slot management
func TestSpellSlots(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("SlotManager")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Sorcerer", "mage")
	require.NoError(t, err)

	gameState, err := client.Call("getGameState", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, gameState)

	if character, ok := gameState["character"].(map[string]interface{}); ok {
		assert.NotNil(t, character)
	}
}

// TestSpellsByClass tests class-specific spell access
func TestSpellsByClass(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	classes := []struct {
		name      string
		className string
	}{
		{"mage_spells", "mage"},
		{"cleric_spells", "cleric"},
	}

	for _, tc := range classes {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.JoinGame("ClassSpellTester_" + tc.className)
			require.NoError(t, err)

			sessionID, _, err := client.CreateCharacter("", "Caster", tc.className)
			require.NoError(t, err)

			result, err := client.Call("getAllSpells", map[string]interface{}{
				"session_id": sessionID,
			})
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestSpellSchools tests different schools of magic
func TestSpellSchools(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("SchoolTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Enchanter", "mage")
	require.NoError(t, err)

	schools := []string{
		"abjuration",
		"conjuration",
		"divination",
		"enchantment",
		"evocation",
		"illusion",
		"necromancy",
		"transmutation",
	}

	for _, school := range schools {
		t.Run("school_"+school, func(t *testing.T) {
			result, err := client.Call("getSpellsBySchool", map[string]interface{}{
				"session_id": sessionID,
				"school":     school,
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestSpellLevels tests spells by level
func TestSpellLevels(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("LevelTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Arcanist", "mage")
	require.NoError(t, err)

	for level := 0; level <= 9; level++ {
		t.Run("level_"+string(rune('0'+level)), func(t *testing.T) {
			result, err := client.Call("getSpellsByLevel", map[string]interface{}{
				"session_id": sessionID,
				"level":      level,
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestSpellSearch tests spell search functionality
func TestSpellSearch(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("SearchTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Researcher", "mage")
	require.NoError(t, err)

	searchTerms := []string{
		"fire",
		"heal",
		"shield",
		"bolt",
		"light",
	}

	for _, term := range searchTerms {
		t.Run("search_"+term, func(t *testing.T) {
			result, err := client.Call("searchSpells", map[string]interface{}{
				"session_id": sessionID,
				"query":      term,
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}
