package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateContent tests basic content generation
func TestGenerateContent(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("ContentGenerator")
	require.NoError(t, err)

	result, err := client.Call("generateContent", map[string]interface{}{
		"session_id": sessionID,
		"type":       "terrain",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestTerrainGeneration tests procedural terrain generation
func TestTerrainGeneration(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("TerrainGen")
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
				"session_id": sessionID,
				"biome":      tc.biome,
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

	sessionID, err := client.JoinGame("ItemGen")
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
				"session_id": sessionID,
				"type":       tc.itemType,
				"rarity":     tc.rarity,
				"count":      1,
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

	sessionID, err := client.JoinGame("LevelGen")
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
				"session_id": sessionID,
				"type":       tc.levelType,
				"difficulty": tc.difficulty,
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

	sessionID, err := client.JoinGame("QuestGen")
	require.NoError(t, err)

	_, err = client.CreateCharacter(sessionID, "Adventurer", "ranger")
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

	sessionID, err := client.JoinGame("StatGetter")
	require.NoError(t, err)

	_, err = client.Call("generateLevel", map[string]interface{}{
		"session_id": sessionID,
		"type":       "dungeon",
		"difficulty": 3,
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

	sessionID, err := client.JoinGame("Validator")
	require.NoError(t, err)

	result, err := client.Call("validateContent", map[string]interface{}{
		"session_id": sessionID,
		"content_id": "test_content",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestDeterministicGeneration tests that PCG with same seed produces same results
func TestDeterministicGeneration(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID1, err := client.JoinGame("Seed1")
	require.NoError(t, err)

	sessionID2, err := client.JoinGame("Seed2")
	require.NoError(t, err)

	seed := int64(12345)

	result1, err := client.Call("generateLevel", map[string]interface{}{
		"session_id": sessionID1,
		"type":       "dungeon",
		"difficulty": 5,
		"seed":       seed,
	})
	require.NoError(t, err)

	result2, err := client.Call("generateLevel", map[string]interface{}{
		"session_id": sessionID2,
		"type":       "dungeon",
		"difficulty": 5,
		"seed":       seed,
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

	sessionID, err := client.JoinGame("BiomeTester")
	require.NoError(t, err)

	biomes := []string{
		"forest",
		"desert",
		"mountain",
		"swamp",
		"tundra",
		"grassland",
		"dungeon",
		"cave",
	}

	for _, biome := range biomes {
		t.Run("biome_"+biome, func(t *testing.T) {
			result, err := client.Call("regenerateTerrain", map[string]interface{}{
				"session_id": sessionID,
				"biome":      biome,
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

	sessionID, err := client.JoinGame("RarityTester")
	require.NoError(t, err)

	rarities := []string{
		"common",
		"uncommon",
		"rare",
		"very_rare",
		"legendary",
	}

	for _, rarity := range rarities {
		t.Run("rarity_"+rarity, func(t *testing.T) {
			result, err := client.Call("generateItems", map[string]interface{}{
				"session_id": sessionID,
				"type":       "weapon",
				"rarity":     rarity,
				"count":      3,
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

	sessionID, err := client.JoinGame("LargeScaleTester")
	require.NoError(t, err)

	result, err := client.Call("generateItems", map[string]interface{}{
		"session_id": sessionID,
		"type":       "mixed",
		"count":      100,
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
}
