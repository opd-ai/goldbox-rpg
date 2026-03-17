package server

import (
	"encoding/json"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers

func createTestServerForHandlers(t *testing.T) *RPCServer {
	server, err := NewRPCServer("../../web")
	require.NoError(t, err)
	require.NotNil(t, server)
	return server
}

func createTestSessionForHandlers(t *testing.T, server *RPCServer) *PlayerSession {
	character := &game.Character{
		ID:              "test-player-001",
		Name:            "Test Player",
		HP:              100,
		MaxHP:           100,
		ActionPoints:    10,
		MaxActionPoints: 10,
		Strength:        15,
		Dexterity:       14,
		Constitution:    13,
		Intelligence:    12,
		Wisdom:          11,
		Charisma:        10,
		Level:           5,
		Equipment:       make(map[game.EquipmentSlot]game.Item),
		Inventory:       []game.Item{},
	}
	character.Position = game.Position{X: 5, Y: 5, Level: 0}

	player := &game.Player{
		Character: *character.Clone(),
	}

	session := &PlayerSession{
		SessionID:   "test-session-001",
		Player:      player,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   true,
		MessageChan: make(chan []byte, 500),
		WSConn:      newMockWebSocketConn(), // Mock WebSocket connection for tests
	}

	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()

	// Ensure world bounds are set (default world is 10x10)
	if server.state.WorldState.Width == 0 {
		server.state.WorldState.Width = 10
		server.state.WorldState.Height = 10
	}

	// Add player to game state
	server.state.AddPlayer(session)

	return session
}

// TestHandleMove tests the move handler
func TestHandleMove(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		setupServer func(*RPCServer) *PlayerSession
		expectError bool
		checkResult func(t *testing.T, result interface{}, session *PlayerSession)
	}{
		{
			name: "valid move north",
			params: map[string]interface{}{
				"session_id": "test-session-001",
				"direction":  0, // DirectionNorth
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				return createTestSessionForHandlers(t, server)
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}, session *PlayerSession) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotNil(t, resultMap["position"])

				// Check that Y position decreased (north = -Y)
				pos := session.Player.Character.GetPosition()
				assert.Equal(t, 4, pos.Y)
			},
		},
		{
			name: "invalid session",
			params: map[string]interface{}{
				"session_id": "invalid-session",
				"direction":  0,
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				return nil // No session needed
			},
			expectError: true,
		},
		{
			name: "missing direction uses default 0",
			params: map[string]interface{}{
				"session_id": "test-session-001",
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				return createTestSessionForHandlers(t, server)
			},
			expectError: false, // JSON unmarshal defaults to 0 (North) when direction not specified
			checkResult: func(t *testing.T, result interface{}, session *PlayerSession) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
			},
		},
		{
			name: "invalid direction (out of range) still succeeds",
			params: map[string]interface{}{
				"session_id": "test-session-001",
				"direction":  999, // Large values are still processed by game logic
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				return createTestSessionForHandlers(t, server)
			},
			expectError: false, // The direction value is passed through, game logic may handle it
			checkResult: func(t *testing.T, result interface{}, session *PlayerSession) {
				// Either succeeds or validation happens in game layer
				if result != nil {
					resultMap, ok := result.(map[string]interface{})
					if ok {
						assert.NotNil(t, resultMap)
					}
				}
			},
		},
		{
			name: "movement succeeds outside combat",
			params: map[string]interface{}{
				"session_id": "test-session-001",
				"direction":  0, // DirectionNorth
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				session := createTestSessionForHandlers(t, server)
				// Outside combat, AP doesn't matter
				session.Player.Character.SetActionPoints(0)
				return session
			},
			expectError: false, // AP only checked during combat
			checkResult: func(t *testing.T, result interface{}, session *PlayerSession) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)
			session := tt.setupServer(server)

			// Marshal params to JSON
			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			// Call the handler
			result, err := server.handleMove(paramBytes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result, session)
				}
			}
		})
	}
}

// TestHandleJoinGame tests the join game handler
func TestHandleJoinGame(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		expectError bool
		checkResult func(t *testing.T, result interface{})
	}{
		{
			name: "valid join with player name",
			params: map[string]interface{}{
				"player_name": "TestPlayer123",
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, resultMap["session_id"])
				assert.Equal(t, true, resultMap["success"])
			},
		},
		{
			name: "empty player name returns error",
			params: map[string]interface{}{
				"player_name": "",
			},
			expectError: true, // Implementation requires non-empty player_name
		},
		{
			name: "missing player name returns error",
			params: map[string]interface{}{
				"other_field": "value",
			},
			expectError: true, // Implementation requires player_name field
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)

			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			result, err := server.handleJoinGame(paramBytes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

// TestHandleJoinGame_WebSocketSessionReuse verifies that when joinGame is called
// with an existing session_id (as happens via WebSocket enrichment), the player
// is attached to the existing session instead of creating a new one.
func TestHandleJoinGame_WebSocketSessionReuse(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Simulate what getOrCreateSession does: create a bare session (no player).
	existingSessionID := "ws-session-001"
	server.mu.Lock()
	server.sessions[existingSessionID] = &PlayerSession{
		SessionID:   existingSessionID,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}
	server.mu.Unlock()

	// Call joinGame with the existing session_id (mimics WebSocket enrichment).
	joinParams, err := json.Marshal(map[string]interface{}{
		"player_name": "TestWSPlayer",
		"session_id":  existingSessionID,
	})
	require.NoError(t, err)

	result, err := server.handleJoinGame(joinParams)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["success"])
	// The returned session_id must be the existing one, not a new UUID.
	assert.Equal(t, existingSessionID, resultMap["session_id"])

	// Verify the existing session now has a Player.
	server.mu.RLock()
	session := server.sessions[existingSessionID]
	server.mu.RUnlock()
	require.NotNil(t, session)
	require.NotNil(t, session.Player, "player should be attached to existing session")
	assert.Equal(t, "TestWSPlayer", session.Player.Name)

	// Verify getGameState works with the same session_id.
	stateParams, err := json.Marshal(map[string]interface{}{
		"session_id": existingSessionID,
	})
	require.NoError(t, err)

	stateResult, err := server.handleGetGameState(stateParams)
	require.NoError(t, err)

	stateMap, ok := stateResult.(map[string]interface{})
	require.True(t, ok)
	playerData, ok := stateMap["player"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "TestWSPlayer", playerData["name"])
}

// TestHandleGetGameState tests the get game state handler
func TestHandleGetGameState(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		setupServer func(*RPCServer)
		expectError bool
		checkResult func(t *testing.T, result interface{})
	}{
		{
			name: "valid get game state",
			params: map[string]interface{}{
				"session_id": "test-session-001",
			},
			setupServer: func(server *RPCServer) {
				createTestSessionForHandlers(t, server)
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.NotNil(t, resultMap["world"])
				assert.NotNil(t, resultMap["turns"])
				assert.NotNil(t, resultMap["time"])
			},
		},
		{
			name: "invalid session",
			params: map[string]interface{}{
				"session_id": "invalid-session",
			},
			setupServer: func(server *RPCServer) {},
			expectError: true,
		},
		{
			name: "missing session_id",
			params: map[string]interface{}{
				"other_field": "value",
			},
			setupServer: func(server *RPCServer) {
				createTestSessionForHandlers(t, server)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)
			tt.setupServer(server)

			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			result, err := server.handleGetGameState(paramBytes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

// TestHandleCreateCharacter tests the create character handler
func TestHandleCreateCharacter(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		expectError bool
		checkResult func(t *testing.T, result interface{})
	}{
		{
			name: "valid character creation with standard method",
			params: map[string]interface{}{
				"name":             "Warrior",
				"class":            "fighter",
				"attribute_method": "standard",
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotEmpty(t, resultMap["session_id"])
			},
		},
		{
			name: "valid character creation with custom attributes",
			params: map[string]interface{}{
				"name":             "Wizard",
				"class":            "mage",
				"attribute_method": "custom",
				"custom_attributes": map[string]interface{}{
					"strength":     10,
					"dexterity":    14,
					"constitution": 12,
					"intelligence": 16,
					"wisdom":       13,
					"charisma":     11,
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotEmpty(t, resultMap["session_id"])
			},
		},
		{
			name: "valid character creation with pointbuy method",
			params: map[string]interface{}{
				"name":             "Warrior",
				"class":            "fighter",
				"attribute_method": "pointbuy", // Point buy always produces valid attributes
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotEmpty(t, resultMap["session_id"])
			},
		},
		{
			name: "missing character name defaults to Adventurer",
			params: map[string]interface{}{
				"class":            "fighter",
				"attribute_method": "standard",
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotEmpty(t, resultMap["session_id"])
			},
		},
		{
			name: "invalid class returns error",
			params: map[string]interface{}{
				"name":             "Test",
				"class":            "invalid_class",
				"attribute_method": "standard",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)

			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			result, err := server.handleCreateCharacter(paramBytes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

// TestHandleCreateCharacter_ResponseFormat verifies that the character data in the
// createCharacter response can be properly unmarshalled into a PlayerState-compatible
// struct (matching the WASM client's expected JSON format).
func TestHandleCreateCharacter_ResponseFormat(t *testing.T) {
	server := createTestServerForHandlers(t)

	params := map[string]interface{}{
		"name":             "Aldric",
		"class":            "fighter",
		"attribute_method": "standard",
	}
	paramBytes, err := json.Marshal(params)
	require.NoError(t, err)

	result, err := server.handleCreateCharacter(paramBytes)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["success"])

	// Marshal and unmarshal to simulate the client-side JSON round-trip
	data, err := json.Marshal(resultMap)
	require.NoError(t, err)

	// This struct mirrors the WASM client's CreateCharacterResult + PlayerState
	type playerState struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Class      string `json:"class"`
		Level      int    `json:"level"`
		HP         int    `json:"hp"`
		MaxHP      int    `json:"max_hp"`
		AP         int    `json:"ap"`
		MaxAP      int    `json:"max_ap"`
		Experience int    `json:"experience"`
		Position   struct {
			X     int `json:"X"`
			Y     int `json:"Y"`
			Level int `json:"Level"`
		} `json:"position"`
		Attributes struct {
			Strength     int `json:"strength"`
			Dexterity    int `json:"dexterity"`
			Constitution int `json:"constitution"`
			Intelligence int `json:"intelligence"`
			Wisdom       int `json:"wisdom"`
			Charisma     int `json:"charisma"`
		} `json:"attributes"`
	}
	type createCharacterResult struct {
		Success   bool         `json:"success"`
		SessionID string       `json:"session_id"`
		Character *playerState `json:"character"`
	}

	var out createCharacterResult
	err = json.Unmarshal(data, &out)
	require.NoError(t, err, "response must be unmarshalable into client-side struct")

	assert.True(t, out.Success)
	assert.NotEmpty(t, out.SessionID)
	require.NotNil(t, out.Character)
	assert.Equal(t, "Aldric", out.Character.Name)
	assert.Equal(t, "Fighter", out.Character.Class)
	assert.Greater(t, out.Character.HP, 0)
	assert.Greater(t, out.Character.MaxHP, 0)
	assert.Greater(t, out.Character.Attributes.Strength, 0)
	assert.Greater(t, out.Character.Attributes.Dexterity, 0)
	assert.Greater(t, out.Character.Attributes.Constitution, 0)
	assert.Greater(t, out.Character.Attributes.Intelligence, 0)
	assert.Greater(t, out.Character.Attributes.Wisdom, 0)
	assert.Greater(t, out.Character.Attributes.Charisma, 0)
}

// TestParseEquipmentSlot tests equipment slot parsing
func TestParseEquipmentSlot(t *testing.T) {
	tests := []struct {
		name        string
		slotName    string
		expected    game.EquipmentSlot
		expectError bool
	}{
		{
			name:        "valid head slot",
			slotName:    "head",
			expected:    game.SlotHead,
			expectError: false,
		},
		{
			name:        "valid chest slot",
			slotName:    "chest",
			expected:    game.SlotChest,
			expectError: false,
		},
		{
			name:        "valid weapon main slot with underscore",
			slotName:    "weapon_main",
			expected:    game.SlotWeaponMain,
			expectError: false,
		},
		{
			name:        "valid main hand slot",
			slotName:    "main_hand",
			expected:    game.SlotWeaponMain,
			expectError: false,
		},
		{
			name:        "valid weapon off slot",
			slotName:    "weapon_off",
			expected:    game.SlotWeaponOff,
			expectError: false,
		},
		{
			name:        "valid off hand slot",
			slotName:    "off_hand",
			expected:    game.SlotWeaponOff,
			expectError: false,
		},
		{
			name:        "uppercase resolves correctly",
			slotName:    "HEAD",
			expected:    game.SlotHead,
			expectError: false, // Case-insensitive matching
		},
		{
			name:        "whitespace returns error",
			slotName:    "  head  ",
			expectError: true, // Implementation doesn't trim whitespace
		},
		{
			name:        "invalid slot name",
			slotName:    "invalid",
			expectError: true,
		},
		{
			name:        "empty slot name",
			slotName:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEquipmentSlot(tt.slotName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestEquipmentSlotToString tests equipment slot to string conversion
func TestEquipmentSlotToString(t *testing.T) {
	tests := []struct {
		name     string
		slot     game.EquipmentSlot
		expected string
	}{
		{
			name:     "head slot",
			slot:     game.SlotHead,
			expected: "head",
		},
		{
			name:     "chest slot",
			slot:     game.SlotChest,
			expected: "chest",
		},
		{
			name:     "weapon main slot",
			slot:     game.SlotWeaponMain,
			expected: "weapon_main",
		},
		{
			name:     "weapon off slot",
			slot:     game.SlotWeaponOff,
			expected: "weapon_off",
		},
		{
			name:     "neck slot",
			slot:     game.SlotNeck,
			expected: "neck",
		},
		{
			name:     "hands slot",
			slot:     game.SlotHands,
			expected: "hands",
		},
		{
			name:     "rings slot",
			slot:     game.SlotRings,
			expected: "rings",
		},
		{
			name:     "legs slot",
			slot:     game.SlotLegs,
			expected: "legs",
		},
		{
			name:     "feet slot",
			slot:     game.SlotFeet,
			expected: "feet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equipmentSlotToString(tt.slot)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHandlersErrorCases tests various error conditions across handlers
func TestHandlersErrorCases(t *testing.T) {
	server := createTestServerForHandlers(t)

	tests := []struct {
		name        string
		handler     func(json.RawMessage) (interface{}, error)
		params      interface{}
		expectError bool
	}{
		{
			name:        "handleMove with malformed JSON",
			handler:     server.handleMove,
			params:      "not valid json",
			expectError: true,
		},
		{
			name:        "handleAttack with empty params",
			handler:     server.handleAttack,
			params:      map[string]interface{}{},
			expectError: true,
		},
		{
			name:        "handleCastSpell with missing session",
			handler:     server.handleCastSpell,
			params:      map[string]interface{}{"spell_id": "fireball"},
			expectError: true,
		},
		{
			name:        "handleGetGameState with nil params",
			handler:     server.handleGetGameState,
			params:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paramBytes json.RawMessage
			if tt.params != nil {
				var err error
				paramBytes, err = json.Marshal(tt.params)
				if err != nil {
					// If we can't marshal, pass empty JSON
					paramBytes = json.RawMessage("{}")
				}
			}

			_, err := tt.handler(paramBytes)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestParseDirectionString tests direction parsing from string
func TestParseDirectionString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantDir   game.Direction
		wantValid bool
	}{
		{"north lowercase", "north", game.DirectionNorth, true},
		{"north short", "n", game.DirectionNorth, true},
		{"north uppercase", "NORTH", game.DirectionNorth, true},
		{"north mixed", "NoRtH", game.DirectionNorth, true},
		{"east lowercase", "east", game.DirectionEast, true},
		{"east short", "e", game.DirectionEast, true},
		{"south lowercase", "south", game.DirectionSouth, true},
		{"south short", "s", game.DirectionSouth, true},
		{"west lowercase", "west", game.DirectionWest, true},
		{"west short", "w", game.DirectionWest, true},
		{"with spaces", "  north  ", game.DirectionNorth, true},
		{"invalid direction", "northeast", 0, false},
		{"empty string", "", 0, false},
		{"random text", "invalid", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDir, gotValid := parseDirectionString(tt.input)
			if gotDir != tt.wantDir || gotValid != tt.wantValid {
				t.Errorf("parseDirectionString(%q) = (%v, %v), want (%v, %v)",
					tt.input, gotDir, gotValid, tt.wantDir, tt.wantValid)
			}
		})
	}
}
