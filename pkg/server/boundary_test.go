package server

import (
	"encoding/json"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/gorilla/websocket"
)

func TestMovementBoundaryEnforcement(t *testing.T) {
	server := NewRPCServer(":8080")

	// Create a small world for testing (5x5)
	server.state.WorldState = game.NewWorldWithSize(5, 5, 10)

	// Create test player at position (2,2)
	player := &game.Player{
		Character: game.Character{
			ID:       "test_player",
			Name:     "Test Player",
			Position: game.Position{X: 2, Y: 2},
		},
	}

	session := &PlayerSession{
		SessionID: "test_session",
		Player:    player,
		Connected: true,
		WSConn:    &websocket.Conn{}, // Mock WebSocket connection
	}

	// Properly add session with thread safety
	server.mu.Lock()
	server.sessions["test_session"] = session
	server.mu.Unlock()

	server.state.WorldState.AddObject(player)

	t.Run("Movement beyond north boundary should be prevented", func(t *testing.T) {
		// Move north multiple times to try to exceed boundary
		// In a 5x5 world, valid Y coordinates are 0-4
		// Starting at Y=2, should be able to move to Y=1, Y=0, but no further

		// Move to Y=1 (should succeed)
		moveParams := map[string]interface{}{
			"session_id": "test_session",
			"direction":  game.North,
		}
		paramsJSON, _ := json.Marshal(moveParams)

		result, err := server.handleMove(paramsJSON)
		if err != nil {
			t.Errorf("First north move should succeed, got error: %v", err)
		}

		if player.GetPosition().Y != 1 {
			t.Errorf("Expected Y=1 after first north move, got Y=%d", player.GetPosition().Y)
		}

		// Move to Y=0 (should succeed)
		result, err = server.handleMove(paramsJSON)
		if err != nil {
			t.Errorf("Second north move should succeed, got error: %v", err)
		}

		if player.GetPosition().Y != 0 {
			t.Errorf("Expected Y=0 after second north move, got Y=%d", player.GetPosition().Y)
		}

		// Try to move beyond Y=0 (should be prevented)
		result, err = server.handleMove(paramsJSON)
		if err != nil {
			t.Errorf("Third north move should succeed but be clamped, got error: %v", err)
		}

		// Position should remain at Y=0 (clamped to boundary)
		if player.GetPosition().Y != 0 {
			t.Errorf("Expected Y=0 after third north move (clamped), got Y=%d", player.GetPosition().Y)
		}

		// Verify result indicates no movement occurred
		resultMap := result.(map[string]interface{})
		position := resultMap["position"].(game.Position)
		if position.Y != 0 {
			t.Errorf("Expected result position Y=0, got Y=%d", position.Y)
		}
	})

	t.Run("Movement beyond south boundary should be prevented", func(t *testing.T) {
		// Reset player to center
		player.SetPosition(game.Position{X: 2, Y: 2})

		// Move south multiple times to try to go beyond Y=4
		moveParams := map[string]interface{}{
			"session_id": "test_session",
			"direction":  game.South,
		}
		paramsJSON, _ := json.Marshal(moveParams)

		// Move to Y=3 (should succeed)
		_, err := server.handleMove(paramsJSON)
		if err != nil {
			t.Errorf("First south move should succeed, got error: %v", err)
		}

		if player.GetPosition().Y != 3 {
			t.Errorf("Expected Y=3 after first south move, got Y=%d", player.GetPosition().Y)
		}

		// Move to Y=4 (should succeed)
		_, err = server.handleMove(paramsJSON)
		if err != nil {
			t.Errorf("Second south move should succeed, got error: %v", err)
		}

		if player.GetPosition().Y != 4 {
			t.Errorf("Expected Y=4 after second south move, got Y=%d", player.GetPosition().Y)
		}

		// Try to move beyond Y=4 (should be prevented)
		_, err = server.handleMove(paramsJSON)
		if err != nil {
			t.Errorf("Third south move should succeed but be clamped, got error: %v", err)
		}

		// Position should remain at Y=4 (clamped to boundary)
		if player.GetPosition().Y != 4 {
			t.Errorf("Expected Y=4 after third south move (clamped), got Y=%d", player.GetPosition().Y)
		}
	})

	t.Run("Movement beyond east boundary should be prevented", func(t *testing.T) {
		// Reset player to center
		player.SetPosition(game.Position{X: 2, Y: 2})

		// Move east to boundary
		moveParams := map[string]interface{}{
			"session_id": "test_session",
			"direction":  game.East,
		}
		paramsJSON, _ := json.Marshal(moveParams)

		// Move to X=3, then X=4, then try to go beyond
		server.handleMove(paramsJSON)
		server.handleMove(paramsJSON)

		if player.GetPosition().X != 4 {
			t.Errorf("Expected X=4 after two east moves, got X=%d", player.GetPosition().X)
		}

		// Try to move beyond X=4 (should be prevented)
		_, err := server.handleMove(paramsJSON)
		if err != nil {
			t.Errorf("East move beyond boundary should succeed but be clamped, got error: %v", err)
		}

		// Position should remain at X=4
		if player.GetPosition().X != 4 {
			t.Errorf("Expected X=4 after east move beyond boundary (clamped), got X=%d", player.GetPosition().X)
		}
	})

	t.Run("Movement beyond west boundary should be prevented", func(t *testing.T) {
		// Reset player to center
		player.SetPosition(game.Position{X: 2, Y: 2})

		// Move west to boundary
		moveParams := map[string]interface{}{
			"session_id": "test_session",
			"direction":  game.West,
		}
		paramsJSON, _ := json.Marshal(moveParams)

		// Move to X=1, then X=0, then try to go beyond
		server.handleMove(paramsJSON)
		server.handleMove(paramsJSON)

		if player.GetPosition().X != 0 {
			t.Errorf("Expected X=0 after two west moves, got X=%d", player.GetPosition().X)
		}

		// Try to move beyond X=0 (should be prevented)
		_, err := server.handleMove(paramsJSON)
		if err != nil {
			t.Errorf("West move beyond boundary should succeed but be clamped, got error: %v", err)
		}

		// Position should remain at X=0
		if player.GetPosition().X != 0 {
			t.Errorf("Expected X=0 after west move beyond boundary (clamped), got X=%d", player.GetPosition().X)
		}
	})
}
