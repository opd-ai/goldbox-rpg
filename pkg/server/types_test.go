package server

import (
	"reflect"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"
)

func TestPlayerSession_Update_ValidUpdates(t *testing.T) {
	// Create a test player
	testPlayer := &game.Player{
		Character: game.Character{
			ID:   "test-player",
			Name: "Test Player",
		},
		Level:      1,
		Experience: 0,
	}

	// Create a test session
	session := &PlayerSession{
		SessionID:  "test-session",
		Player:     testPlayer,
		Connected:  false,
		LastActive: time.Now().Add(-time.Hour),
	}

	tests := []struct {
		name       string
		updateData map[string]interface{}
		validator  func(*testing.T, *PlayerSession)
	}{
		{
			name: "Update connected status to true",
			updateData: map[string]interface{}{
				"connected": true,
			},
			validator: func(t *testing.T, ps *PlayerSession) {
				if !ps.Connected {
					t.Errorf("Expected Connected to be true, got %v", ps.Connected)
				}
			},
		},
		{
			name: "Update connected status to false",
			updateData: map[string]interface{}{
				"connected": false,
			},
			validator: func(t *testing.T, ps *PlayerSession) {
				if ps.Connected {
					t.Errorf("Expected Connected to be false, got %v", ps.Connected)
				}
			},
		},
		{
			name: "Update lastActive timestamp",
			updateData: map[string]interface{}{
				"lastActive": time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			validator: func(t *testing.T, ps *PlayerSession) {
				expected := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				if !ps.LastActive.Equal(expected) {
					t.Errorf("Expected LastActive to be %v, got %v", expected, ps.LastActive)
				}
			},
		},
		{
			name: "Update sessionId",
			updateData: map[string]interface{}{
				"sessionId": "new-session-id",
			},
			validator: func(t *testing.T, ps *PlayerSession) {
				if ps.SessionID != "new-session-id" {
					t.Errorf("Expected SessionID to be 'new-session-id', got %s", ps.SessionID)
				}
			},
		},
		{
			name: "Update player data",
			updateData: map[string]interface{}{
				"player": map[string]interface{}{
					"level":      5,
					"experience": 1000,
				},
			},
			validator: func(t *testing.T, ps *PlayerSession) {
				if ps.Player.Level != 5 {
					t.Errorf("Expected Player.Level to be 5, got %d", ps.Player.Level)
				}
				if ps.Player.Experience != 1000 {
					t.Errorf("Expected Player.Experience to be 1000, got %d", ps.Player.Experience)
				}
			},
		},
		{
			name: "Update multiple fields at once",
			updateData: map[string]interface{}{
				"connected":  true,
				"sessionId":  "multi-update-session",
				"lastActive": time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC),
			},
			validator: func(t *testing.T, ps *PlayerSession) {
				if !ps.Connected {
					t.Errorf("Expected Connected to be true, got %v", ps.Connected)
				}
				if ps.SessionID != "multi-update-session" {
					t.Errorf("Expected SessionID to be 'multi-update-session', got %s", ps.SessionID)
				}
				expected := time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC)
				if !ps.LastActive.Equal(expected) {
					t.Errorf("Expected LastActive to be %v, got %v", expected, ps.LastActive)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh session for each test
			testSession := &PlayerSession{
				SessionID:  session.SessionID,
				Player:     session.Player,
				Connected:  session.Connected,
				LastActive: session.LastActive,
			}

			err := testSession.Update(tt.updateData)
			if err != nil {
				t.Errorf("Update() returned unexpected error: %v", err)
			}

			tt.validator(t, testSession)
		})
	}
}

func TestPlayerSession_Update_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		session     *PlayerSession
		updateData  map[string]interface{}
		expectError bool
	}{
		{
			name:        "Update nil session",
			session:     nil,
			updateData:  map[string]interface{}{"connected": true},
			expectError: true,
		},
		{
			name: "Update with wrong type for connected",
			session: &PlayerSession{
				SessionID: "test",
				Player:    &game.Player{},
			},
			updateData:  map[string]interface{}{"connected": "not-a-bool"},
			expectError: false, // Should not error, just ignore invalid type
		},
		{
			name: "Update with wrong type for lastActive",
			session: &PlayerSession{
				SessionID: "test",
				Player:    &game.Player{},
			},
			updateData:  map[string]interface{}{"lastActive": "not-a-time"},
			expectError: false, // Should not error, just ignore invalid type
		},
		{
			name: "Update with wrong type for sessionId",
			session: &PlayerSession{
				SessionID: "test",
				Player:    &game.Player{},
			},
			updateData:  map[string]interface{}{"sessionId": 123},
			expectError: false, // Should not error, just ignore invalid type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Update(tt.updateData)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPlayerSession_Clone_ValidSession(t *testing.T) {
	originalTime := time.Date(2023, 5, 10, 14, 30, 0, 0, time.UTC)
	createdTime := time.Date(2023, 5, 10, 10, 0, 0, 0, time.UTC)

	// Create test player
	testPlayer := &game.Player{
		Character: game.Character{
			ID:   "test-player-123",
			Name: "Test Player",
		},
		Level:      3,
		Experience: 500,
	}

	// Create test session
	original := &PlayerSession{
		SessionID:   "original-session-id",
		Player:      testPlayer,
		LastActive:  originalTime,
		CreatedAt:   createdTime,
		Connected:   true,
		MessageChan: make(chan []byte, 10),
		WSConn:      nil, // Using nil for testing
	}

	// Test the clone
	cloned := original.Clone()

	// Verify clone is not nil
	if cloned == nil {
		t.Fatal("Clone() returned nil")
	}

	// Verify clone has same values
	if cloned.SessionID != original.SessionID {
		t.Errorf("SessionID mismatch: expected %s, got %s", original.SessionID, cloned.SessionID)
	}

	if cloned.Connected != original.Connected {
		t.Errorf("Connected mismatch: expected %v, got %v", original.Connected, cloned.Connected)
	}

	if !cloned.LastActive.Equal(original.LastActive) {
		t.Errorf("LastActive mismatch: expected %v, got %v", original.LastActive, cloned.LastActive)
	}

	if !cloned.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt mismatch: expected %v, got %v", original.CreatedAt, cloned.CreatedAt)
	}

	// Verify player is cloned (not same reference)
	if cloned.Player == original.Player {
		t.Error("Player should be cloned, not same reference")
	}

	// Verify player data is copied correctly
	if cloned.Player.ID != original.Player.ID {
		t.Errorf("Player ID mismatch: expected %s, got %s", original.Player.ID, cloned.Player.ID)
	}

	if cloned.Player.Level != original.Player.Level {
		t.Errorf("Player Level mismatch: expected %d, got %d", original.Player.Level, cloned.Player.Level)
	}

	// Verify MessageChan is a new channel
	if cloned.MessageChan == original.MessageChan {
		t.Error("MessageChan should be a new channel, not same reference")
	}

	// Verify WSConn is same reference (as expected)
	if cloned.WSConn != original.WSConn {
		t.Error("WSConn should be same reference")
	}
}

func TestPlayerSession_Clone_NilSession(t *testing.T) {
	var session *PlayerSession = nil

	cloned := session.Clone()

	if cloned != nil {
		t.Errorf("Clone() of nil session should return nil, got %v", cloned)
	}
}

func TestPlayerSession_PublicData(t *testing.T) {
	testTime := time.Date(2023, 7, 20, 16, 45, 0, 0, time.UTC)

	// Create test player
	testPlayer := &game.Player{
		Character: game.Character{
			ID:   "public-data-player",
			Name: "Public Player",
		},
		Level:      2,
		Experience: 250,
	}

	// Create test session
	session := &PlayerSession{
		SessionID:  "public-session",
		Player:     testPlayer,
		Connected:  true,
		LastActive: testTime,
	}

	// Get public data
	publicData := session.PublicData()

	// Type assertion to check structure
	data, ok := publicData.(struct {
		SessionID  string      `json:"sessionId"`
		PlayerData interface{} `json:"player"`
		Connected  bool        `json:"connected"`
		LastActive time.Time   `json:"lastActive"`
	})

	if !ok {
		t.Fatal("PublicData() returned wrong type structure")
	}

	// Verify data fields
	if data.SessionID != "public-session" {
		t.Errorf("SessionID mismatch: expected 'public-session', got %s", data.SessionID)
	}

	if data.Connected != true {
		t.Errorf("Connected mismatch: expected true, got %v", data.Connected)
	}

	if !data.LastActive.Equal(testTime) {
		t.Errorf("LastActive mismatch: expected %v, got %v", testTime, data.LastActive)
	}

	// Verify PlayerData is not nil
	if data.PlayerData == nil {
		t.Error("PlayerData should not be nil")
	}
}

func TestRPCMethod_Constants(t *testing.T) {
	// Test that RPC method constants have expected values
	tests := []struct {
		name     string
		method   RPCMethod
		expected string
	}{
		{"MethodMove", MethodMove, "move"},
		{"MethodAttack", MethodAttack, "attack"},
		{"MethodCastSpell", MethodCastSpell, "castSpell"},
		{"MethodUseItem", MethodUseItem, "useItem"},
		{"MethodApplyEffect", MethodApplyEffect, "applyEffect"},
		{"MethodStartCombat", MethodStartCombat, "startCombat"},
		{"MethodEndTurn", MethodEndTurn, "endTurn"},
		{"MethodGetGameState", MethodGetGameState, "getGameState"},
		{"MethodJoinGame", MethodJoinGame, "joinGame"},
		{"MethodLeaveGame", MethodLeaveGame, "leaveGame"},
		{"MethodCreateCharacter", MethodCreateCharacter, "createCharacter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.method) != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, string(tt.method))
			}
		})
	}
}

func TestRPCMethod_TypeConversion(t *testing.T) {
	// Test that RPCMethod can be converted to/from string
	original := "customMethod"
	method := RPCMethod(original)
	converted := string(method)

	if converted != original {
		t.Errorf("Type conversion failed: expected %s, got %s", original, converted)
	}
}

func TestPlayerSession_StructureIntegrity(t *testing.T) {
	// Test that PlayerSession struct has expected fields
	session := &PlayerSession{}

	// Use reflection to verify struct fields exist
	sessionType := reflect.TypeOf(*session)

	expectedFields := []string{
		"SessionID",
		"Player",
		"LastActive",
		"CreatedAt",
		"Connected",
		"MessageChan",
		"WSConn",
	}

	for _, fieldName := range expectedFields {
		_, found := sessionType.FieldByName(fieldName)
		if !found {
			t.Errorf("Expected field %s not found in PlayerSession struct", fieldName)
		}
	}
}

func TestPlayerSession_ZeroValues(t *testing.T) {
	// Test behavior with zero values
	session := &PlayerSession{
		Player: &game.Player{}, // Initialize Player to avoid nil pointer
	}

	// Test Update with zero value session (should not panic)
	err := session.Update(map[string]interface{}{
		"connected": true,
	})
	if err != nil {
		t.Errorf("Update on zero-value session should not error, got: %v", err)
	}

	// Test PublicData with zero values (should not panic)
	publicData := session.PublicData()
	if publicData == nil {
		t.Error("PublicData should not return nil even for zero-value session")
	}
}
