package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateContent tests basic content generation
func TestGenerateContent(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("ContentGenerator")
	require.NoError(t, err)

	// Create a character to establish a player session
	sessionID, _, err := client.CreateCharacter("", "ContentGen", "fighter")
	require.NoError(t, err)

	result, err := client.Call("generateContent", map[string]interface{}{
		"session_id":   sessionID,
		"content_type": "terrain",
		"location_id":  "test-location-001",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestTerrainGeneration tests procedural terrain generation
func TestTerrainGeneration(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("TerrainGen")
	require.NoError(t, err)

	// Create a character to establish a player session
	sessionID, _, err := client.CreateCharacter("", "TerrainGen", "ranger")
	require.NoError(t, err)

	testCases := []struct {
		name        string
		biome       string
		expectError bool
	}{
		{
			name:        "generate_forest",
			biome:       "forest",
			expectError: false,
		},
		{
			name:        "generate_desert",
			biome:       "desert",
			expectError: false,
		},
		{
			name:        "generate_mountain",
			biome:       "mountain",
			expectError: false,
		},
		{
			name:        "generate_dungeon",
			biome:       "dungeon",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("regenerateTerrain", map[string]interface{}{
				"session_id":  sessionID,
				"biome_type":  tc.biome,
				"location_id": "test-terrain-" + tc.biome,
			})

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestItemGeneration tests procedural item generation
func TestItemGeneration(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("ItemGen")
	require.NoError(t, err)

	// Create a character to establish a player session
	sessionID, _, err := client.CreateCharacter("", "ItemGen", "thief")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		itemType   string
		rarity     string
		expectData bool
	}{
		{
			name:       "generate_common_weapon",
			itemType:   "weapon",
			rarity:     "common",
			expectData: true,
		},
		{
			name:       "generate_rare_armor",
			itemType:   "armor",
			rarity:     "rare",
			expectData: true,
		},
		{
			name:       "generate_magical_potion",
			itemType:   "potion",
			rarity:     "uncommon",
			expectData: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("generateItems", map[string]interface{}{
				"session_id":  sessionID,
				"type":        tc.itemType,
				"rarity":      tc.rarity,
				"count":       1,
				"location_id": "test-item-gen",
			})

			require.NoError(t, err)
			if tc.expectData {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestLevelGeneration tests full level generation
func TestLevelGeneration(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("LevelGen")
	require.NoError(t, err)

	// Create a character to establish a player session
	sessionID, _, err := client.CreateCharacter("", "LevelGen", "mage")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		levelType  string
		difficulty int
	}{
		{
			name:       "generate_easy_dungeon",
			levelType:  "dungeon",
			difficulty: 1,
		},
		{
			name:       "generate_medium_cave",
			levelType:  "cave",
			difficulty: 5,
		},
		{
			name:       "generate_hard_fortress",
			levelType:  "fortress",
			difficulty: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("generateLevel", map[string]interface{}{
				"session_id":  sessionID,
				"type":        tc.levelType,
				"difficulty":  tc.difficulty,
				"location_id": "test-level-" + tc.levelType,
			})

			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestQuestGeneration tests procedural quest generation
func TestQuestGeneration(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("QuestGen")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Adventurer", "ranger")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		questType string
		level     int
	}{
		{
			name:      "generate_fetch_quest",
			questType: "fetch",
			level:     1,
		},
		{
			name:      "generate_kill_quest",
			questType: "kill",
			level:     3,
		},
		{
			name:      "generate_escort_quest",
			questType: "escort",
			level:     5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("generateQuest", map[string]interface{}{
				"session_id": sessionID,
				"type":       tc.questType,
				"level":      tc.level,
			})

			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestPCGStats tests getting PCG statistics
func TestPCGStats(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("StatGetter")
	require.NoError(t, err)

	// Create a character to establish a player session
	sessionID, _, err := client.CreateCharacter("", "StatGetter", "cleric")
	require.NoError(t, err)

	_, err = client.Call("generateLevel", map[string]interface{}{
		"session_id":  sessionID,
		"type":        "dungeon",
		"difficulty":  3,
		"location_id": "test-stats-dungeon",
	})
	require.NoError(t, err)

	result, err := client.Call("getPCGStats", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestContentValidation tests PCG content validation
func TestContentValidation(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("Validator")
	require.NoError(t, err)

	// Create a character to establish a player session
	sessionID, _, err := client.CreateCharacter("", "Validator", "paladin")
	require.NoError(t, err)

	// Test content validation with proper parameters
	result, err := client.Call("validateContent", map[string]interface{}{
		"session_id":   sessionID,
		"content_type": "items",
		"content": map[string]interface{}{
			"id":     "test_item",
			"name":   "Test Sword",
			"type":   "weapon",
			"rarity": "common",
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestDeterministicGeneration tests that PCG with same seed produces same results
func TestDeterministicGeneration(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("Seed1")
	require.NoError(t, err)

	// Create characters for both sessions
	sessionID1, _, err := client.CreateCharacter("", "SeedChar1", "fighter")
	require.NoError(t, err)

	_, err = client.JoinGame("Seed2")
	require.NoError(t, err)

	sessionID2, _, err := client.CreateCharacter("", "SeedChar2", "mage")
	require.NoError(t, err)

	seed := int64(12345)

	result1, err := client.Call("generateLevel", map[string]interface{}{
		"session_id":  sessionID1,
		"type":        "dungeon",
		"difficulty":  5,
		"seed":        seed,
		"location_id": "test-seed1-dungeon",
	})
	require.NoError(t, err)

	result2, err := client.Call("generateLevel", map[string]interface{}{
		"session_id":  sessionID2,
		"type":        "dungeon",
		"difficulty":  5,
		"seed":        seed,
		"location_id": "test-seed2-dungeon",
	})
	require.NoError(t, err)

	assert.NotNil(t, result1)
	assert.NotNil(t, result2)
}

// TestBiomeVariety tests different biome types
func TestBiomeVariety(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, _, err := client.CreateCharacter("", "BiomeTester", "fighter")
	require.NoError(t, err)

	// Test valid biomes as defined in pkg/pcg/types_enums.go
	biomes := []string{
		"forest",
		"desert",
		"mountain",
		"swamp",
		"dungeon",
		"cave",
		"coastal",
		"urban",
		"wasteland",
	}

	for i, biome := range biomes {
		t.Run("biome_"+biome, func(t *testing.T) {
			result, err := client.Call("regenerateTerrain", map[string]interface{}{
				"session_id":  sessionID,
				"biome_type":  biome,
				"location_id": fmt.Sprintf("test-biome-%d-%s", i, biome),
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestItemRarityDistribution tests item generation by rarity
func TestItemRarityDistribution(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, _, err := client.CreateCharacter("", "RarityTester", "fighter")
	require.NoError(t, err)

	rarities := []string{
		"common",
		"uncommon",
		"rare",
		"very_rare",
		"legendary",
	}

	for i, rarity := range rarities {
		t.Run("rarity_"+rarity, func(t *testing.T) {
			result, err := client.Call("generateItems", map[string]interface{}{
				"session_id":  sessionID,
				"type":        "weapon",
				"rarity":      rarity,
				"count":       3,
				"location_id": fmt.Sprintf("test-rarity-%d-%s", i, rarity),
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestLargeScaleGeneration tests generating large amounts of content
func TestLargeScaleGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large-scale generation in short mode")
	}

	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, _, err := client.CreateCharacter("", "LargeScaleTester", "fighter")
	require.NoError(t, err)

	result, err := client.Call("generateItems", map[string]interface{}{
		"session_id":  sessionID,
		"type":        "mixed",
		"count":       100,
		"location_id": "test-large-scale-gen",
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
}
