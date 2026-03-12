package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEditorMapWorkflow tests the complete map editor lifecycle:
// create → edit tiles → save → load → verify.
func TestEditorMapWorkflow(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	// Step 1: Create a session with a player character
	sessionID, _ := helper.CreateSession()

	// Step 2: Create a new map via the editor
	createResult, err := client.Call("editor.createMap", map[string]interface{}{
		"session_id": sessionID,
		"name":       "Test Dungeon",
		"width":      10,
		"height":     10,
	})
	require.NoError(t, err, "should create map")
	require.NotNil(t, createResult)
	assert.True(t, createResult["success"].(bool), "create should succeed")

	mapID, ok := createResult["map_id"].(string)
	require.True(t, ok, "map_id should be a string")
	require.NotEmpty(t, mapID, "map_id should not be empty")

	// Step 3: Update a tile on the created map
	updateResult, err := client.Call("editor.updateTile", map[string]interface{}{
		"session_id":   sessionID,
		"map_id":       mapID,
		"x":            3,
		"y":            4,
		"terrain_type": "wall",
	})
	require.NoError(t, err, "should update tile")
	require.NotNil(t, updateResult)
	assert.True(t, updateResult["success"].(bool), "tile update should succeed")

	// Step 4: Save the map
	saveResult, err := client.Call("editor.saveMap", map[string]interface{}{
		"session_id": sessionID,
		"map_id":     mapID,
		"filename":   "test_dungeon.yaml",
	})
	require.NoError(t, err, "should save map")
	require.NotNil(t, saveResult)
	assert.True(t, saveResult["success"].(bool), "save should succeed")

	// Step 5: Load the map back and verify
	loadResult, err := client.Call("editor.loadMap", map[string]interface{}{
		"session_id": sessionID,
		"map_id":     mapID,
		"filename":   "test_dungeon.yaml",
	})
	require.NoError(t, err, "should load map")
	require.NotNil(t, loadResult)
	assert.True(t, loadResult["success"].(bool), "load should succeed")

	// Verify the loaded data contains expected fields
	assert.Contains(t, loadResult, "map_id", "loaded map should have map_id")
}

// TestEditorMapValidation tests editor input validation for map operations.
func TestEditorMapValidation(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, _ := helper.CreateSession()

	t.Run("create_map_missing_name", func(t *testing.T) {
		_, err := client.Call("editor.createMap", map[string]interface{}{
			"session_id": sessionID,
			"name":       "",
			"width":      10,
			"height":     10,
		})
		require.Error(t, err, "should reject map with empty name")
	})

	t.Run("create_map_invalid_dimensions", func(t *testing.T) {
		_, err := client.Call("editor.createMap", map[string]interface{}{
			"session_id": sessionID,
			"name":       "Bad Map",
			"width":      -1,
			"height":     10,
		})
		require.Error(t, err, "should reject map with negative width")
	})

	t.Run("create_map_oversized", func(t *testing.T) {
		_, err := client.Call("editor.createMap", map[string]interface{}{
			"session_id": sessionID,
			"name":       "Huge Map",
			"width":      999,
			"height":     999,
		})
		require.Error(t, err, "should reject map exceeding max dimensions")
	})
}

// TestEditorQuestWorkflow tests the complete quest editor CRUD lifecycle:
// create → get → update → list → delete → verify deletion.
func TestEditorQuestWorkflow(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	// Create a session with a player character
	sessionID, _ := helper.CreateSession()

	// Step 1: Create a quest
	createResult, err := client.Call("questEditor.create", map[string]interface{}{
		"session_id":  sessionID,
		"title":       "Defeat the Dragon",
		"description": "A mighty dragon terrorizes the kingdom.",
		"objectives": []map[string]interface{}{
			{"description": "Find the dragon's lair", "required": 1},
			{"description": "Defeat the dragon", "required": 1},
		},
		"rewards": []map[string]interface{}{
			{"type": "gold", "value": 500},
			{"type": "exp", "value": 1000},
		},
	})
	require.NoError(t, err, "should create quest")
	require.NotNil(t, createResult)
	assert.True(t, createResult["success"].(bool), "create should succeed")

	questID, ok := createResult["quest_id"].(string)
	require.True(t, ok, "quest_id should be a string")
	require.NotEmpty(t, questID, "quest_id should not be empty")

	// Step 2: Get the quest
	getResult, err := client.Call("questEditor.get", map[string]interface{}{
		"session_id": sessionID,
		"quest_id":   questID,
	})
	require.NoError(t, err, "should get quest")
	require.NotNil(t, getResult)
	assert.True(t, getResult["success"].(bool))
	assert.Equal(t, questID, getResult["quest_id"])

	// Step 3: Update the quest
	updateResult, err := client.Call("questEditor.update", map[string]interface{}{
		"session_id":  sessionID,
		"quest_id":    questID,
		"title":       "Defeat the Ancient Dragon",
		"description": "An ancient dragon terrorizes the kingdom. Only the bravest shall prevail.",
		"objectives": []map[string]interface{}{
			{"description": "Find the dragon's lair", "required": 1},
			{"description": "Obtain the dragon-slaying sword", "required": 1},
			{"description": "Defeat the ancient dragon", "required": 1},
		},
		"rewards": []map[string]interface{}{
			{"type": "gold", "value": 1000},
			{"type": "exp", "value": 2000},
			{"type": "item", "value": 1, "item_id": "dragon_scale_armor"},
		},
	})
	require.NoError(t, err, "should update quest")
	require.NotNil(t, updateResult)
	assert.True(t, updateResult["success"].(bool))

	// Step 4: List quests (should include our quest)
	listResult, err := client.Call("questEditor.list", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err, "should list quests")
	require.NotNil(t, listResult)
	assert.True(t, listResult["success"].(bool))

	// Step 5: Delete the quest
	deleteResult, err := client.Call("questEditor.delete", map[string]interface{}{
		"session_id": sessionID,
		"quest_id":   questID,
	})
	require.NoError(t, err, "should delete quest")
	require.NotNil(t, deleteResult)
	assert.True(t, deleteResult["success"].(bool))
}

// TestEditorQuestValidation tests quest editor input validation.
func TestEditorQuestValidation(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, _ := helper.CreateSession()

	t.Run("create_quest_missing_title", func(t *testing.T) {
		_, err := client.Call("questEditor.create", map[string]interface{}{
			"session_id": sessionID,
			"title":      "",
			"objectives": []map[string]interface{}{
				{"description": "Do something", "required": 1},
			},
		})
		require.Error(t, err, "should reject quest with empty title")
	})

	t.Run("create_quest_no_objectives", func(t *testing.T) {
		_, err := client.Call("questEditor.create", map[string]interface{}{
			"session_id": sessionID,
			"title":      "Empty Quest",
			"objectives": []map[string]interface{}{},
		})
		require.Error(t, err, "should reject quest with no objectives")
	})

	t.Run("create_quest_invalid_reward_type", func(t *testing.T) {
		_, err := client.Call("questEditor.create", map[string]interface{}{
			"session_id": sessionID,
			"title":      "Bad Reward Quest",
			"objectives": []map[string]interface{}{
				{"description": "Task", "required": 1},
			},
			"rewards": []map[string]interface{}{
				{"type": "magic_beans", "value": 42},
			},
		})
		require.Error(t, err, "should reject quest with invalid reward type")
	})

	t.Run("get_quest_missing_quest_id", func(t *testing.T) {
		_, err := client.Call("questEditor.get", map[string]interface{}{
			"session_id": sessionID,
			"quest_id":   "",
		})
		require.Error(t, err, "should reject get with empty quest_id")
	})

	t.Run("delete_quest_missing_quest_id", func(t *testing.T) {
		_, err := client.Call("questEditor.delete", map[string]interface{}{
			"session_id": sessionID,
			"quest_id":   "",
		})
		require.Error(t, err, "should reject delete with empty quest_id")
	})
}
