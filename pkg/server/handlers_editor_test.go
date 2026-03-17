package server

import (
	"encoding/json"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEditorHandlers tests the editor-specific RPC method handlers
func TestEditorHandlers(t *testing.T) {
	// Create a test server
	server, err := NewRPCServer(":8080")
	require.NoError(t, err, "Failed to create server")
	defer server.Stop()

	// Create a test session
	sessionID := "test_session_editor"
	testCharacter := game.Character{
		ID:       "test_player_editor",
		Name:     "Editor Test Player",
		Class:    game.ClassFighter,
		Position: game.Position{X: 5, Y: 5},
	}
	testPlayer := &game.Player{
		Character: *testCharacter.Clone(),
		Level:     1,
	}

	testSession := &PlayerSession{
		SessionID:   sessionID,
		Player:      testPlayer,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   true,
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}

	server.mu.Lock()
	server.sessions[sessionID] = testSession
	server.mu.Unlock()

	t.Run("TestHandleEditorCreateMap", func(t *testing.T) {
		tests := []struct {
			name        string
			params      map[string]interface{}
			expectError bool
			checkFields []string
		}{
			{
				name: "create map successfully",
				params: map[string]interface{}{
					"session_id": sessionID,
					"name":       "Test Map",
					"width":      20,
					"height":     15,
				},
				expectError: false,
				checkFields: []string{"success", "map_id", "width", "height"},
			},
			{
				name: "create map with template",
				params: map[string]interface{}{
					"session_id": sessionID,
					"name":       "Dungeon Map",
					"width":      30,
					"height":     25,
					"template":   "dungeon",
				},
				expectError: false,
				checkFields: []string{"success", "map_id"},
			},
			{
				name: "missing session_id",
				params: map[string]interface{}{
					"name":   "Test Map",
					"width":  20,
					"height": 15,
				},
				expectError: true,
			},
			{
				name: "missing name",
				params: map[string]interface{}{
					"session_id": sessionID,
					"width":      20,
					"height":     15,
				},
				expectError: true,
			},
			{
				name: "invalid width (too large)",
				params: map[string]interface{}{
					"session_id": sessionID,
					"name":       "Test Map",
					"width":      300,
					"height":     15,
				},
				expectError: true,
			},
			{
				name: "invalid height (negative)",
				params: map[string]interface{}{
					"session_id": sessionID,
					"name":       "Test Map",
					"width":      20,
					"height":     -5,
				},
				expectError: true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				paramsJSON, err := json.Marshal(tc.params)
				require.NoError(t, err, "Failed to marshal params")

				result, err := server.handleEditorCreateMap(paramsJSON)

				if tc.expectError {
					assert.Error(t, err, "Expected error but got none")
					return
				}

				require.NoError(t, err, "Unexpected error")
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok, "Expected result to be a map")

				for _, field := range tc.checkFields {
					assert.Contains(t, resultMap, field, "Missing expected field: %s", field)
				}

				if resultMap["success"] != nil {
					assert.True(t, resultMap["success"].(bool), "Expected success to be true")
				}
			})
		}
	})

	t.Run("TestHandleEditorUpdateTile", func(t *testing.T) {
		// First create a map to update
		createParams := map[string]interface{}{
			"session_id": sessionID,
			"name":       "Test Map for Tile Update",
			"width":      20,
			"height":     15,
		}
		createParamsJSON, _ := json.Marshal(createParams)
		createResult, err := server.handleEditorCreateMap(createParamsJSON)
		require.NoError(t, err, "Failed to create map")
		createdMapID := createResult.(map[string]interface{})["map_id"].(string)

		tests := []struct {
			name        string
			params      map[string]interface{}
			expectError bool
		}{
			{
				name: "update tile successfully",
				params: map[string]interface{}{
					"session_id":  sessionID,
					"map_id":      createdMapID,
					"x":           5,
					"y":           5,
					"sprite_x":    1,
					"sprite_y":    0,
					"walkable":    false,
					"transparent": false,
				},
				expectError: false,
			},
			{
				name: "missing map_id",
				params: map[string]interface{}{
					"session_id": sessionID,
					"x":          5,
					"y":          5,
				},
				expectError: true,
			},
			{
				name: "negative coordinates",
				params: map[string]interface{}{
					"session_id": sessionID,
					"map_id":     createdMapID,
					"x":          -1,
					"y":          5,
				},
				expectError: true,
			},
			{
				name: "map not found",
				params: map[string]interface{}{
					"session_id": sessionID,
					"map_id":     "nonexistent-map-id",
					"x":          5,
					"y":          5,
				},
				expectError: true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				paramsJSON, err := json.Marshal(tc.params)
				require.NoError(t, err, "Failed to marshal params")

				result, err := server.handleEditorUpdateTile(paramsJSON)

				if tc.expectError {
					assert.Error(t, err, "Expected error but got none")
					return
				}

				require.NoError(t, err, "Unexpected error")
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok, "Expected result to be a map")
				assert.True(t, resultMap["success"].(bool), "Expected success")
			})
		}
	})

	t.Run("TestHandleEditorSaveMap", func(t *testing.T) {
		// First create a map to save
		createParams := map[string]interface{}{
			"session_id": sessionID,
			"name":       "Test Map for Save",
			"width":      20,
			"height":     15,
		}
		createParamsJSON, _ := json.Marshal(createParams)
		createResult, err := server.handleEditorCreateMap(createParamsJSON)
		require.NoError(t, err, "Failed to create map")
		createdMapID := createResult.(map[string]interface{})["map_id"].(string)

		tests := []struct {
			name        string
			params      map[string]interface{}
			expectError bool
		}{
			{
				name: "save map successfully",
				params: map[string]interface{}{
					"session_id": sessionID,
					"map_id":     createdMapID,
					"filename":   "test_map.json",
				},
				expectError: false,
			},
			{
				name: "map not found",
				params: map[string]interface{}{
					"session_id": sessionID,
					"map_id":     "nonexistent-map-id",
					"filename":   "test_map.json",
				},
				expectError: true,
			},
			{
				name: "missing map_id",
				params: map[string]interface{}{
					"session_id": sessionID,
					"filename":   "test_map.json",
				},
				expectError: true,
			},
			{
				name: "missing filename",
				params: map[string]interface{}{
					"session_id": sessionID,
					"map_id":     createdMapID,
				},
				expectError: true,
			},
			{
				name: "path traversal attempt",
				params: map[string]interface{}{
					"session_id": sessionID,
					"map_id":     createdMapID,
					"filename":   "../../../etc/passwd",
				},
				expectError: true,
			},
			{
				name: "absolute path attempt",
				params: map[string]interface{}{
					"session_id": sessionID,
					"map_id":     createdMapID,
					"filename":   "/etc/passwd",
				},
				expectError: true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				paramsJSON, err := json.Marshal(tc.params)
				require.NoError(t, err, "Failed to marshal params")

				result, err := server.handleEditorSaveMap(paramsJSON)

				if tc.expectError {
					assert.Error(t, err, "Expected error but got none")
					return
				}

				require.NoError(t, err, "Unexpected error")
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok, "Expected result to be a map")
				assert.True(t, resultMap["success"].(bool), "Expected success")
			})
		}
	})

	t.Run("TestHandleEditorLoadMap", func(t *testing.T) {
		tests := []struct {
			name        string
			params      map[string]interface{}
			expectError bool
		}{
			{
				name: "load map successfully",
				params: map[string]interface{}{
					"session_id": sessionID,
					"filename":   "test_map.json",
				},
				expectError: false,
			},
			{
				name: "missing filename",
				params: map[string]interface{}{
					"session_id": sessionID,
				},
				expectError: true,
			},
			{
				name: "path traversal attempt",
				params: map[string]interface{}{
					"session_id": sessionID,
					"filename":   "../../secret.json",
				},
				expectError: true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				paramsJSON, err := json.Marshal(tc.params)
				require.NoError(t, err, "Failed to marshal params")

				result, err := server.handleEditorLoadMap(paramsJSON)

				if tc.expectError {
					assert.Error(t, err, "Expected error but got none")
					return
				}

				require.NoError(t, err, "Unexpected error")
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok, "Expected result to be a map")
				assert.True(t, resultMap["success"].(bool), "Expected success")
			})
		}
	})
}

// TestEditorMapStorage tests the editor map storage functionality
func TestEditorMapStorage(t *testing.T) {
	storage := NewEditorMapStorage()

	t.Run("SetAndGetMap", func(t *testing.T) {
		testMap := &game.GameMap{
			Width:  20,
			Height: 15,
			Tiles:  make([][]game.MapTile, 15),
		}
		for y := 0; y < 15; y++ {
			testMap.Tiles[y] = make([]game.MapTile, 20)
		}

		mapID := "test-map-1"
		storage.SetMap(mapID, "Test Map", testMap)

		retrieved, err := storage.GetMap(mapID)
		require.NoError(t, err, "Failed to get map")
		assert.Equal(t, testMap.Width, retrieved.Width)
		assert.Equal(t, testMap.Height, retrieved.Height)
	})

	t.Run("GetNonExistentMap", func(t *testing.T) {
		_, err := storage.GetMap("non-existent-map")
		assert.Error(t, err, "Expected error for non-existent map")
	})

	t.Run("DeleteMap", func(t *testing.T) {
		testMap := &game.GameMap{
			Width:  10,
			Height: 10,
			Tiles:  make([][]game.MapTile, 10),
		}
		for y := 0; y < 10; y++ {
			testMap.Tiles[y] = make([]game.MapTile, 10)
		}

		mapID := "test-map-delete"
		storage.SetMap(mapID, "Delete Test Map", testMap)

		// Verify it exists
		_, err := storage.GetMap(mapID)
		require.NoError(t, err)

		// Delete and verify it's gone
		storage.DeleteMap(mapID)
		_, err = storage.GetMap(mapID)
		assert.Error(t, err, "Expected error after deletion")
	})
}

// TestContainsPathTraversal tests the path traversal detection function
func TestContainsPathTraversal(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.json", false},
		{"maps/dungeon.json", false},
		{"my-map_v2.json", false},
		{"../etc/passwd", true},
		{"/etc/passwd", true},
		{"..\\windows\\system32", true},
		{"../../secret", true},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			result := containsPathTraversal(tc.filename)
			assert.Equal(t, tc.expected, result, "Unexpected result for %q", tc.filename)
		})
	}
}
