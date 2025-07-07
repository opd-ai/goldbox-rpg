package server

import (
	"encoding/json"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/gorilla/websocket"
)

// TestTurnBasedCombatEnforcement tests that combat actions are properly restricted to the current turn
func TestTurnBasedCombatEnforcement(t *testing.T) {
	server := NewRPCServer(":8080")

	// Create test world
	server.state.WorldState = game.NewWorldWithSize(10, 10, 1)

	// Create two test players
	player1 := &game.Player{
		Character: game.Character{
			ID:              "player1",
			Name:            "Test Player 1",
			Position:        game.Position{X: 1, Y: 1},
			HP:              100,
			MaxHP:           100,
			Level:           1,
			Dexterity:       12, // Normal dexterity (no bonus)
			MaxActionPoints: 2,  // Base action points for level 1
			ActionPoints:    2,  // Start with full action points
		},
		Level:      1,
		Experience: 0,
	}

	player2 := &game.Player{
		Character: game.Character{
			ID:              "player2",
			Name:            "Test Player 2",
			Position:        game.Position{X: 2, Y: 2},
			HP:              100,
			MaxHP:           100,
			Level:           1,
			Dexterity:       12, // Normal dexterity (no bonus)
			MaxActionPoints: 2,  // Base action points for level 1
			ActionPoints:    2,  // Start with full action points
		},
		Level:      1,
		Experience: 0,
	}

	// Create sessions for both players
	session1 := &PlayerSession{
		SessionID: "session1",
		Player:    player1,
		Connected: true,
		WSConn:    &websocket.Conn{},
	}

	session2 := &PlayerSession{
		SessionID: "session2",
		Player:    player2,
		Connected: true,
		WSConn:    &websocket.Conn{},
	}

	// Add sessions to server
	server.mu.Lock()
	server.sessions["session1"] = session1
	server.sessions["session2"] = session2
	server.mu.Unlock()

	// Add players to world
	server.state.WorldState.AddObject(player1)
	server.state.WorldState.AddObject(player2)

	t.Run("Actions allowed when not in combat", func(t *testing.T) {
		// Test that actions work normally when not in combat
		attackParams := map[string]interface{}{
			"session_id": "session1",
			"target_id":  "player2",
			"weapon_id":  "",
		}
		paramsJSON, _ := json.Marshal(attackParams)

		// Attack should fail because not in combat, but not due to turn restriction
		_, err := server.handleAttack(paramsJSON)
		if err == nil {
			t.Error("Expected attack to fail when not in combat")
		}
		if err.Error() != "not in combat" {
			t.Errorf("Expected 'not in combat' error, got: %v", err)
		}
	})

	t.Run("Combat actions restricted by turn order", func(t *testing.T) {
		// Start combat with player1 going first
		server.state.TurnManager.StartCombat([]string{"player1", "player2"})

		// Verify player1 can act (it's their turn)
		attackParams := map[string]interface{}{
			"session_id": "session1",
			"target_id":  "player2",
			"weapon_id":  "",
		}
		paramsJSON, _ := json.Marshal(attackParams)

		result, err := server.handleAttack(paramsJSON)
		if err != nil {
			t.Errorf("Player1 attack should succeed when it's their turn, got error: %v", err)
		}
		if result == nil {
			t.Error("Expected attack result when it's player's turn")
		}

		// Verify player2 cannot act (not their turn)
		attackParams2 := map[string]interface{}{
			"session_id": "session2",
			"target_id":  "player1",
			"weapon_id":  "",
		}
		paramsJSON2, _ := json.Marshal(attackParams2)

		_, err = server.handleAttack(paramsJSON2)
		if err == nil {
			t.Error("Player2 attack should fail when it's not their turn")
		}
		if err.Error() != "not your turn" {
			t.Errorf("Expected 'not your turn' error, got: %v", err)
		}

		// Clean up combat state
		server.state.TurnManager.EndCombat()
	})

	t.Run("Spell casting restricted by turn order", func(t *testing.T) {
		// Start combat again for this test
		server.state.TurnManager.StartCombat([]string{"player1", "player2"})

		// Test spell casting turn restrictions
		spellParams := map[string]interface{}{
			"session_id": "session2",
			"spell_id":   "fireball",
			"target_id":  "player1",
		}
		paramsJSON, _ := json.Marshal(spellParams)

		_, err := server.handleCastSpell(paramsJSON)
		if err == nil {
			t.Error("Player2 spell cast should fail when it's not their turn")
		}
		if err.Error() != "not your turn" {
			t.Errorf("Expected 'not your turn' error for spell cast, got: %v", err)
		}

		// Test that player1 can cast spells (it's their turn)
		spellParams1 := map[string]interface{}{
			"session_id": "session1",
			"spell_id":   "fireball",
			"target_id":  "player2",
		}
		paramsJSON1, _ := json.Marshal(spellParams1)

		// This will fail due to spell not found, but not due to turn restriction
		_, err = server.handleCastSpell(paramsJSON1)
		if err != nil && err.Error() == "not your turn" {
			t.Error("Player1 should be able to attempt spell cast when it's their turn")
		}

		// Clean up combat state
		server.state.TurnManager.EndCombat()
	})

	t.Run("Item usage restricted by turn order", func(t *testing.T) {
		// Start combat again for this test
		server.state.TurnManager.StartCombat([]string{"player1", "player2"})

		// Test item usage turn restrictions
		itemParams := map[string]interface{}{
			"session_id": "session2",
			"item_id":    "healing_potion",
			"target_id":  "player2",
		}
		paramsJSON, _ := json.Marshal(itemParams)

		_, err := server.handleUseItem(paramsJSON)
		if err == nil {
			t.Error("Player2 item use should fail when it's not their turn")
		}
		if err.Error() != "not your turn" {
			t.Errorf("Expected 'not your turn' error for item use, got: %v", err)
		}

		// Test that player1 can use items (it's their turn)
		itemParams1 := map[string]interface{}{
			"session_id": "session1",
			"item_id":    "healing_potion",
			"target_id":  "player1",
		}
		paramsJSON1, _ := json.Marshal(itemParams1)

		// This will fail due to item not found, but not due to turn restriction
		result, err := server.handleUseItem(paramsJSON1)
		if err != nil && err.Error() == "not your turn" {
			t.Error("Player1 should be able to attempt item use when it's their turn")
		}
		// Check that we get a proper response (item not found)
		if result != nil {
			resultMap := result.(map[string]interface{})
			if success, ok := resultMap["success"].(bool); ok && success {
				t.Error("Expected item use to fail due to item not found")
			}
		}

		// Clean up combat state
		server.state.TurnManager.EndCombat()
	})

	t.Run("Turn advancement allows next player to act", func(t *testing.T) {
		// Start combat for this test
		server.state.TurnManager.StartCombat([]string{"player1", "player2"})

		// Advance to player2's turn
		nextPlayer := server.state.TurnManager.AdvanceTurn()
		if nextPlayer != "player2" {
			t.Errorf("Expected player2's turn after advancement, got: %s", nextPlayer)
		}

		// Now player2 should be able to act
		attackParams := map[string]interface{}{
			"session_id": "session2",
			"target_id":  "player1",
			"weapon_id":  "",
		}
		paramsJSON, _ := json.Marshal(attackParams)

		result, err := server.handleAttack(paramsJSON)
		if err != nil {
			t.Errorf("Player2 attack should succeed when it's their turn, got error: %v", err)
		}
		if result == nil {
			t.Error("Expected attack result when it's player2's turn")
		}

		// Player1 should no longer be able to act
		attackParams1 := map[string]interface{}{
			"session_id": "session1",
			"target_id":  "player2",
			"weapon_id":  "",
		}
		paramsJSON1, _ := json.Marshal(attackParams1)

		_, err = server.handleAttack(paramsJSON1)
		if err == nil {
			t.Error("Player1 attack should fail when it's not their turn")
		}
		if err.Error() != "not your turn" {
			t.Errorf("Expected 'not your turn' error, got: %v", err)
		}

		// Clean up combat state
		server.state.TurnManager.EndCombat()
	})

	t.Run("EndTurn works regardless of current turn", func(t *testing.T) {
		// Start combat for this test
		server.state.TurnManager.StartCombat([]string{"player1", "player2"})

		// End turn should work for any player (they can pass their turn)
		endTurnParams := map[string]interface{}{
			"session_id": "session1",
		}
		paramsJSON, _ := json.Marshal(endTurnParams)

		result, err := server.handleEndTurn(paramsJSON)
		if err != nil {
			t.Errorf("EndTurn should work for any player, got error: %v", err)
		}
		if result == nil {
			t.Error("Expected result from endTurn")
		}

		// Clean up combat state
		server.state.TurnManager.EndCombat()
	})
}

// TestCombatTurnValidationEdgeCases tests edge cases in turn validation
func TestCombatTurnValidationEdgeCases(t *testing.T) {
	server := NewRPCServer(":8080")
	server.state.WorldState = game.NewWorldWithSize(5, 5, 1)

	// Test with invalid session
	t.Run("Invalid session handled gracefully", func(t *testing.T) {
		attackParams := map[string]interface{}{
			"session_id": "invalid_session",
			"target_id":  "target",
			"weapon_id":  "",
		}
		paramsJSON, _ := json.Marshal(attackParams)

		_, err := server.handleAttack(paramsJSON)
		if err == nil {
			t.Error("Expected error for invalid session")
		}
		if err.Error() != "invalid session" {
			t.Errorf("Expected 'invalid session' error, got: %v", err)
		}
	})

	t.Run("Combat state transitions handled properly", func(t *testing.T) {
		// Create test session
		player := &game.Player{
			Character: game.Character{
				ID:   "test_player",
				Name: "Test Player",
				HP:   100,
			},
		}

		session := &PlayerSession{
			SessionID: "test_session",
			Player:    player,
			Connected: true,
			WSConn:    &websocket.Conn{},
		}

		server.mu.Lock()
		server.sessions["test_session"] = session
		server.mu.Unlock()

		// Start combat
		server.state.TurnManager.StartCombat([]string{"test_player"})

		attackParams := map[string]interface{}{
			"session_id": "test_session",
			"target_id":  "some_target",
			"weapon_id":  "",
		}
		paramsJSON, _ := json.Marshal(attackParams)

		// Should be able to act when it's their turn
		_, err := server.handleAttack(paramsJSON)
		// This may fail for other reasons (target not found), but not turn validation
		if err != nil && err.Error() == "not your turn" {
			t.Error("Should not get 'not your turn' error when player is current turn")
		}

		// End combat properly
		server.state.TurnManager.EndCombat()

		// Actions should work when not in combat (may fail for other reasons)
		_, err = server.handleAttack(paramsJSON)
		if err != nil && err.Error() == "not your turn" {
			t.Error("Should not get turn validation error when not in combat")
		}
	})
}
