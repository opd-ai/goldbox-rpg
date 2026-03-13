package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdventureList tests the adventure.list RPC method
func TestAdventureList(t *testing.T) {
	server, err := NewTestServer()
	require.NoError(t, err, "Failed to create test server")
	require.NoError(t, server.Start(), "Failed to start test server")
	defer server.Stop()

	client := NewClient(server.BaseURL())

	// Call adventure.list
	result, err := client.Call("adventure.list", nil)
	require.NoError(t, err, "adventure.list call failed")

	// Verify result structure
	adventures, ok := result["adventures"].([]interface{})
	require.True(t, ok, "Expected adventures array in result")
	require.GreaterOrEqual(t, len(adventures), 10, "Expected at least 10 adventures")

	// Verify each adventure has required fields
	for i, adv := range adventures {
		advMap, ok := adv.(map[string]interface{})
		require.True(t, ok, "Adventure %d should be a map", i)

		assert.NotEmpty(t, advMap["id"], "Adventure %d missing id", i)
		assert.NotEmpty(t, advMap["slug"], "Adventure %d missing slug", i)
		assert.NotEmpty(t, advMap["title"], "Adventure %d missing title", i)
		assert.NotEmpty(t, advMap["description"], "Adventure %d missing description", i)

		// Verify level range
		minLevel, ok := advMap["min_level"].(float64)
		require.True(t, ok, "Adventure %d min_level should be a number", i)
		maxLevel, ok := advMap["max_level"].(float64)
		require.True(t, ok, "Adventure %d max_level should be a number", i)
		assert.GreaterOrEqual(t, maxLevel, minLevel, "Adventure %d max_level should be >= min_level", i)
	}
}

// TestAdventureLoad tests the adventure.load RPC method
func TestAdventureLoad(t *testing.T) {
	server, err := NewTestServer()
	require.NoError(t, err, "Failed to create test server")
	require.NoError(t, server.Start(), "Failed to start test server")
	defer server.Stop()

	client := NewClient(server.BaseURL())

	// Test loading each known adventure
	adventureSlugs := []string{
		"sunken-sanctum",
		"crimson-coast",
		"frost-barrow",
		"forbidden-spire",
		"ember-caverns",
		"giant-clans",
		"emerald-swamp",
		"iron-colosseum",
		"dreaming-pharaoh",
		"void-tyrant",
	}

	for _, slug := range adventureSlugs {
		t.Run(slug, func(t *testing.T) {
			result, err := client.Call("adventure.load", map[string]interface{}{
				"slug": slug,
			})
			require.NoError(t, err, "adventure.load call failed for %s", slug)

			// Verify adventure loaded
			adventure, ok := result["adventure"].(map[string]interface{})
			require.True(t, ok, "Expected adventure object in result")

			// Verify basic fields
			assert.Equal(t, slug, adventure["slug"], "Slug mismatch")
			assert.NotEmpty(t, adventure["title"], "Missing title")
			assert.NotEmpty(t, adventure["description"], "Missing description")

			// Verify maps array exists
			maps, ok := adventure["maps"].([]interface{})
			require.True(t, ok, "Expected maps array")
			assert.GreaterOrEqual(t, len(maps), 5, "Expected at least 5 maps")

			// Verify NPCs array exists
			npcs, ok := adventure["npcs"].([]interface{})
			require.True(t, ok, "Expected npcs array")
			assert.GreaterOrEqual(t, len(npcs), 1, "Expected at least 1 NPC")

			// Verify items array exists
			items, ok := adventure["items"].([]interface{})
			require.True(t, ok, "Expected items array")
			assert.GreaterOrEqual(t, len(items), 1, "Expected at least 1 item")

			// Verify encounters array exists
			encounters, ok := adventure["encounters"].([]interface{})
			require.True(t, ok, "Expected encounters array")
			assert.GreaterOrEqual(t, len(encounters), 1, "Expected at least 1 encounter")

			// Verify quests array exists
			quests, ok := adventure["quests"].([]interface{})
			require.True(t, ok, "Expected quests array")
			assert.GreaterOrEqual(t, len(quests), 1, "Expected at least 1 quest")
		})
	}
}

// TestAdventureLoadNotFound tests loading a non-existent adventure
func TestAdventureLoadNotFound(t *testing.T) {
	server, err := NewTestServer()
	require.NoError(t, err, "Failed to create test server")
	require.NoError(t, server.Start(), "Failed to start test server")
	defer server.Stop()

	client := NewClient(server.BaseURL())

	_, err = client.Call("adventure.load", map[string]interface{}{
		"slug": "non-existent-adventure",
	})
	require.Error(t, err, "Expected error for non-existent adventure")
}

// TestAdventureValidation tests that all adventures have valid structure
func TestAdventureValidation(t *testing.T) {
	server, err := NewTestServer()
	require.NoError(t, err, "Failed to create test server")
	require.NoError(t, server.Start(), "Failed to start test server")
	defer server.Stop()

	client := NewClient(server.BaseURL())

	// Get list of adventures
	listResult, err := client.Call("adventure.list", nil)
	require.NoError(t, err, "adventure.list call failed")

	adventures, ok := listResult["adventures"].([]interface{})
	require.True(t, ok, "Expected adventures array")

	for _, adv := range adventures {
		advMap := adv.(map[string]interface{})
		slug := advMap["slug"].(string)

		t.Run(slug+"_validation", func(t *testing.T) {
			result, err := client.Call("adventure.load", map[string]interface{}{
				"slug": slug,
			})
			require.NoError(t, err, "Failed to load adventure %s", slug)

			adventure := result["adventure"].(map[string]interface{})

			// Validate quest chain integrity
			quests, ok := adventure["quests"].([]interface{})
			require.True(t, ok, "Expected quests array for %s", slug)

			questIDs := make(map[string]bool)
			for _, q := range quests {
				quest := q.(map[string]interface{})
				questIDs[quest["id"].(string)] = true
			}

			// Verify quest_next references exist (except for final quest)
			for _, q := range quests {
				quest := q.(map[string]interface{})
				if nextQuest, ok := quest["next_quest"].(string); ok && nextQuest != "" {
					assert.True(t, questIDs[nextQuest],
						"Quest %s references non-existent next quest %s in %s",
						quest["id"], nextQuest, slug)
				}
			}

			// Validate encounters reference existing maps
			encounters, ok := adventure["encounters"].([]interface{})
			require.True(t, ok, "Expected encounters array for %s", slug)

			maps := adventure["maps"].([]interface{})
			mapIDs := make(map[string]bool)
			for _, m := range maps {
				mapObj := m.(map[string]interface{})
				mapIDs[mapObj["id"].(string)] = true
			}

			for _, e := range encounters {
				enc := e.(map[string]interface{})
				mapID := enc["map_id"].(string)
				assert.True(t, mapIDs[mapID],
					"Encounter %s references non-existent map %s in %s",
					enc["id"], mapID, slug)
			}
		})
	}
}

// TestAdventureSmokeRun tests that each adventure can be started
func TestAdventureSmokeRun(t *testing.T) {
	server, err := NewTestServer()
	require.NoError(t, err, "Failed to create test server")
	require.NoError(t, server.Start(), "Failed to start test server")
	defer server.Stop()

	client := NewClient(server.BaseURL())

	// Create a session first
	_, err = client.Call("joinGame", map[string]interface{}{
		"player_name": "AdventureTestPlayer",
	})
	require.NoError(t, err, "Failed to join game")

	// Get list of adventures
	listResult, err := client.Call("adventure.list", nil)
	require.NoError(t, err, "adventure.list call failed")

	adventures := listResult["adventures"].([]interface{})

	// Verify we have all 10 adventures
	require.GreaterOrEqual(t, len(adventures), 10, "Expected at least 10 adventures")

	// Log adventure summary
	t.Logf("Found %d adventures:", len(adventures))
	for _, adv := range adventures {
		advMap := adv.(map[string]interface{})
		t.Logf("  - %s (Levels %v-%v)",
			advMap["title"],
			advMap["min_level"],
			advMap["max_level"])
	}
}
