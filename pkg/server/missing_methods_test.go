package server

import (
	"encoding/json"
	"testing"

	"goldbox-rpg/pkg/game"
)

// TestHandleMissingMethods tests the previously missing RPC methods
func TestHandleMissingMethods(t *testing.T) {
	server := createTestServer()

	tests := []struct {
		name    string
		method  RPCMethod
		params  interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:   "useItem with valid parameters",
			method: MethodUseItem,
			params: map[string]interface{}{
				"session_id": "test-session",
				"item_id":    "test-item",
				"target_id":  "test-target",
			},
			wantErr: true, // Will fail because session doesn't exist, but handler exists
			errMsg:  "invalid session",
		},
		{
			name:   "leaveGame with valid parameters",
			method: MethodLeaveGame,
			params: map[string]interface{}{
				"session_id": "test-session",
			},
			wantErr: true, // Will fail because session doesn't exist, but handler exists
			errMsg:  "invalid session",
		},
		{
			name:   "useItem with missing item_id",
			method: MethodUseItem,
			params: map[string]interface{}{
				"session_id": "test-session",
				"target_id":  "test-target",
			},
			wantErr: true,
			errMsg:  "item ID is required",
		},
		{
			name:   "leaveGame with empty session_id",
			method: MethodLeaveGame,
			params: map[string]interface{}{
				"session_id": "",
			},
			wantErr: true,
			errMsg:  "invalid session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert params to json.RawMessage
			paramsJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("Failed to marshal params: %v", err)
			}

			// Call the handler
			result, err := server.handleMethod(tt.method, paramsJSON)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("Expected result but got nil")
				}
			}
		})
	}
}

// TestHandleUseItemWithValidSession tests useItem with a valid session
func TestHandleUseItemWithValidSession(t *testing.T) {
	server := createTestServer()

	// Create a test session with a player
	sessionID := "test-session-valid"
	session := &PlayerSession{
		SessionID:   sessionID,
		Player:      createTestPlayer(),
		MessageChan: make(chan []byte, 100),
	}
	server.sessions[sessionID] = session

	// Add a test item to the player's inventory
	testItem := game.Item{
		ID:   "test-potion",
		Name: "Test Potion",
		Type: "consumable",
	}
	session.Player.Character.Inventory = append(session.Player.Character.Inventory, testItem)

	params := map[string]interface{}{
		"session_id": sessionID,
		"item_id":    "test-potion",
		"target_id":  "player",
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}

	result, err := server.handleMethod(MethodUseItem, paramsJSON)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", result)
	}

	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Errorf("Expected success to be true, got %v", resultMap["success"])
	}

	if effect, ok := resultMap["effect"].(string); !ok || effect == "" {
		t.Errorf("Expected non-empty effect string, got %v", resultMap["effect"])
	}
}

// TestHandleLeaveGameWithValidSession tests leaveGame with a valid session
func TestHandleLeaveGameWithValidSession(t *testing.T) {
	server := createTestServer()

	// Create a test session
	sessionID := "test-session-leave"
	session := &PlayerSession{
		SessionID:   sessionID,
		Player:      createTestPlayer(),
		MessageChan: make(chan []byte, 100),
	}
	server.sessions[sessionID] = session

	params := map[string]interface{}{
		"session_id": sessionID,
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}

	result, err := server.handleMethod(MethodLeaveGame, paramsJSON)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", result)
	}

	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Errorf("Expected success to be true, got %v", resultMap["success"])
	}

	// Verify session was removed
	if _, exists := server.sessions[sessionID]; exists {
		t.Errorf("Expected session to be removed, but it still exists")
	}
}

func createTestServer() *RPCServer {
	return &RPCServer{
		sessions: make(map[string]*PlayerSession),
		state: &GameState{
			WorldState: &game.World{
				Objects: make(map[string]game.GameObject),
			},
		},
	}
}

func createTestPlayer() *game.Player {
	character := game.Character{
		ID:        "test-player",
		Name:      "Test Player",
		Inventory: []game.Item{},
	}
	return &game.Player{
		Character: character,
	}
}
