package server

import (
	"encoding/json"
	"testing"

	"goldbox-rpg/pkg/game"
)

func TestSpatialIndexingRPCIntegration(t *testing.T) {
	server, err := NewRPCServer(":8080")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a world with spatial indexing enabled
	server.state.WorldState = game.NewWorldWithSize(200, 200, 25)

	// Create test session and player
	player := &game.Player{
		Character: game.Character{
			ID:       "test_player",
			Name:     "Test Player",
			Position: game.Position{X: 100, Y: 100},
		},
	}

	session := &PlayerSession{
		SessionID: "test_session",
		Player:    player,
		Connected: true,
	}

	server.sessions["test_session"] = session
	server.state.WorldState.AddObject(player)

	// Add some test objects around the player
	npc1 := &game.NPC{
		Character: game.Character{
			ID:       "npc_close",
			Name:     "Close NPC",
			Position: game.Position{X: 105, Y: 105},
		},
	}

	npc2 := &game.NPC{
		Character: game.Character{
			ID:       "npc_far",
			Name:     "Far NPC",
			Position: game.Position{X: 150, Y: 150},
		},
	}

	enemy := &game.NPC{
		Character: game.Character{
			ID:       "enemy",
			Name:     "Enemy",
			Position: game.Position{X: 108, Y: 108},
		},
	}

	server.state.WorldState.AddObject(npc1)
	server.state.WorldState.AddObject(npc2)
	server.state.WorldState.AddObject(enemy)

	// Test getObjectsInRange RPC call
	t.Run("GetObjectsInRange", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "test_session",
			"min_x":      95,
			"min_y":      95,
			"max_x":      110,
			"max_y":      110,
		}

		paramsJSON, _ := json.Marshal(params)
		result, err := server.handleGetObjectsInRange(paramsJSON)

		if err != nil {
			t.Errorf("GetObjectsInRange failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if !resultMap["success"].(bool) {
			t.Errorf("GetObjectsInRange was not successful: %v", resultMap["error"])
		}

		count := resultMap["count"].(int)
		if count != 3 { // player, npc_close, enemy
			t.Errorf("Expected 3 objects in range, got %d", count)
		}
	})

	// Test getObjectsInRadius RPC call
	t.Run("GetObjectsInRadius", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "test_session",
			"center_x":   100,
			"center_y":   100,
			"radius":     10.0,
		}

		paramsJSON, _ := json.Marshal(params)
		result, err := server.handleGetObjectsInRadius(paramsJSON)

		if err != nil {
			t.Errorf("GetObjectsInRadius failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if !resultMap["success"].(bool) {
			t.Errorf("GetObjectsInRadius was not successful: %v", resultMap["error"])
		}

		count := resultMap["count"].(int)
		if count != 2 { // player (distance 0), npc_close (distance ~7.07)
			t.Errorf("Expected 2 objects in radius, got %d", count)
		}
	})

	// Test getNearestObjects RPC call
	t.Run("GetNearestObjects", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "test_session",
			"center_x":   100,
			"center_y":   100,
			"k":          2,
		}

		paramsJSON, _ := json.Marshal(params)
		result, err := server.handleGetNearestObjects(paramsJSON)

		if err != nil {
			t.Errorf("GetNearestObjects failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if !resultMap["success"].(bool) {
			t.Errorf("GetNearestObjects was not successful: %v", resultMap["error"])
		}

		count := resultMap["count"].(int)
		if count != 2 {
			t.Errorf("Expected 2 nearest objects, got %d", count)
		}
	})

	// Test error handling with invalid session
	t.Run("InvalidSession", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "invalid_session",
			"min_x":      0,
			"min_y":      0,
			"max_x":      10,
			"max_y":      10,
		}

		paramsJSON, _ := json.Marshal(params)
		result, err := server.handleGetObjectsInRange(paramsJSON)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["success"].(bool) {
			t.Error("Expected failure with invalid session")
		}

		if resultMap["error"].(string) != "invalid session" {
			t.Errorf("Expected 'invalid session' error, got: %v", resultMap["error"])
		}
	})
}
