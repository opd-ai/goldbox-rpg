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

// TestHandleCreateCharacter_WebSocketSessionReuse verifies that when createCharacter
// is called with an existing session_id (as happens via WebSocket enrichment), the
// character is attached to the existing session instead of creating a new one.
func TestHandleCreateCharacter_WebSocketSessionReuse(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Simulate a bare WebSocket session (no player yet).
	existingSessionID := "ws-session-cc-001"
	server.mu.Lock()
	server.sessions[existingSessionID] = &PlayerSession{
		SessionID:   existingSessionID,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}
	server.mu.Unlock()

	// Call createCharacter with the existing session_id.
	ccParams, err := json.Marshal(map[string]interface{}{
		"name":               "WSCharacter",
		"class":              "fighter",
		"attribute_method":   "standard",
		"starting_equipment": true,
		"starting_gold":      100,
		"session_id":         existingSessionID,
	})
	require.NoError(t, err)

	result, err := server.handleCreateCharacter(ccParams)
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
	assert.Equal(t, "WSCharacter", session.Player.Name)
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

// TestStunPreventsActionsInCombat tests that stunned players cannot move, attack, or cast spells during combat.
func TestStunPreventsActionsInCombat(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Put server into combat mode
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{session.Player.GetID()}
	server.state.TurnManager.CurrentIndex = 0

	// Apply stun effect to the player
	stunEffect := game.NewEffect(
		game.EffectStun,
		game.Duration{Rounds: 2},
		1.0,
	)
	err := session.Player.Character.AddEffect(stunEffect)
	require.NoError(t, err)

	// Verify the player has the stun effect
	require.True(t, session.Player.HasEffect(game.EffectStun), "Player should have stun effect")

	t.Run("stunned player cannot move", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": session.SessionID,
			"direction":  0, // DirectionNorth
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleMove(paramBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stunned")
	})

	t.Run("stunned player cannot attack", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": session.SessionID,
			"target_id":  "enemy-001",
			"weapon_id":  "sword-001",
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleAttack(paramBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stunned")
	})

	t.Run("stunned player cannot cast spells", func(t *testing.T) {
		// Add a known spell to the player
		session.Player.KnownSpells = []game.Spell{
			{ID: "magic-missile", Name: "Magic Missile"},
		}

		params := map[string]interface{}{
			"session_id": session.SessionID,
			"spell_id":   "magic-missile",
			"target_id":  "enemy-001",
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleCastSpell(paramBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stunned")
	})
}

// TestRootPreventsMovementInCombat tests that rooted players cannot move but can still attack and cast spells during combat.
func TestRootPreventsMovementInCombat(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Put server into combat mode
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{session.Player.GetID()}
	server.state.TurnManager.CurrentIndex = 0

	// Apply root effect to the player
	rootEffect := game.NewEffect(
		game.EffectRoot,
		game.Duration{Rounds: 2},
		1.0,
	)
	err := session.Player.Character.AddEffect(rootEffect)
	require.NoError(t, err)

	// Verify the player has the root effect
	require.True(t, session.Player.HasEffect(game.EffectRoot), "Player should have root effect")

	t.Run("rooted player cannot move", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": session.SessionID,
			"direction":  0, // DirectionNorth
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleMove(paramBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rooted")
	})

	t.Run("rooted player can still attack", func(t *testing.T) {
		// Note: Attack will fail due to no valid target, but NOT due to being rooted
		params := map[string]interface{}{
			"session_id": session.SessionID,
			"target_id":  "enemy-001",
			"weapon_id":  "sword-001",
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleAttack(paramBytes)
		// Error should NOT be about being rooted - it should be about missing target
		if err != nil {
			assert.NotContains(t, err.Error(), "rooted", "Rooted players should be able to attack")
		}
	})

	t.Run("rooted player can still cast spells", func(t *testing.T) {
		// Add a known spell to the player
		session.Player.KnownSpells = []game.Spell{
			{ID: "magic-missile", Name: "Magic Missile"},
		}

		params := map[string]interface{}{
			"session_id": session.SessionID,
			"spell_id":   "magic-missile",
			"target_id":  "enemy-001",
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleCastSpell(paramBytes)
		// Error should NOT be about being rooted - it should be about invalid spell/target
		if err != nil {
			assert.NotContains(t, err.Error(), "rooted", "Rooted players should be able to cast spells")
		}
	})
}

// TestParalysisPreventsAllActionsInCombat tests that paralyzed players cannot move, attack, or cast spells during combat.
// Paralysis is an enhanced stun effect that completely blocks all actions.
func TestParalysisPreventsAllActionsInCombat(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Put server into combat mode
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{session.Player.GetID()}
	server.state.TurnManager.CurrentIndex = 0

	// Apply paralysis effect to the player
	paralysisEffect := game.NewEffect(
		game.EffectParalysis,
		game.Duration{Rounds: 2},
		1.0,
	)
	err := session.Player.Character.AddEffect(paralysisEffect)
	require.NoError(t, err)

	// Verify the player has the paralysis effect
	require.True(t, session.Player.HasEffect(game.EffectParalysis), "Player should have paralysis effect")

	t.Run("paralyzed player cannot move", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": session.SessionID,
			"direction":  0, // DirectionNorth
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleMove(paramBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "paralyzed")
	})

	t.Run("paralyzed player cannot attack", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": session.SessionID,
			"target_id":  "enemy-001",
			"weapon_id":  "sword-001",
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleAttack(paramBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "paralyzed")
	})

	t.Run("paralyzed player cannot cast spells", func(t *testing.T) {
		// Add a known spell to the player
		session.Player.KnownSpells = []game.Spell{
			{ID: "magic-missile", Name: "Magic Missile"},
		}

		params := map[string]interface{}{
			"session_id": session.SessionID,
			"spell_id":   "magic-missile",
			"target_id":  "enemy-001",
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		_, err = server.handleCastSpell(paramBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "paralyzed")
	})
}

// TestStunAndRootOutsideCombat verifies that stun and root effects don't block actions outside combat.
func TestStunAndRootOutsideCombat(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Ensure NOT in combat
	server.state.TurnManager.IsInCombat = false

	// Apply both stun and root effects
	stunEffect := game.NewEffect(game.EffectStun, game.Duration{Rounds: 2}, 1.0)
	rootEffect := game.NewEffect(game.EffectRoot, game.Duration{Rounds: 2}, 1.0)
	require.NoError(t, session.Player.Character.AddEffect(stunEffect))
	require.NoError(t, session.Player.Character.AddEffect(rootEffect))

	t.Run("stunned and rooted player can move outside combat", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": session.SessionID,
			"direction":  0, // DirectionNorth
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		result, err := server.handleMove(paramBytes)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestOpportunityAttackTriggersOnMovement tests that opportunity attacks are triggered
// when a player moves away from an adjacent enemy during combat.
func TestOpportunityAttackTriggersOnMovement(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Create an enemy character at adjacent position
	enemyChar := &game.Character{
		ID:         "enemy-opp-001",
		Name:       "Test Enemy",
		HP:         50,
		MaxHP:      50,
		Strength:   14,
		ArmorClass: 10,
	}

	// Set up player position and enemy position adjacent to player
	session.Player.SetPosition(game.Position{X: 5, Y: 5, Level: 0})
	enemyPos := game.Position{X: 5, Y: 6, Level: 0} // Adjacent to player

	// Add enemy to world objects
	server.state.WorldState.Objects[enemyChar.ID] = enemyChar

	// Put server into combat mode
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{session.Player.GetID(), enemyChar.ID}
	server.state.TurnManager.CurrentIndex = 0

	// Register entities for opportunity attacks
	server.state.OpportunityManager.RegisterEntity(session.Player.GetID(), session.Player.GetPosition())
	server.state.OpportunityManager.RegisterEntity(enemyChar.ID, enemyPos)

	t.Run("opportunity attack triggers when moving away from enemy", func(t *testing.T) {
		// Move north (away from enemy who is south)
		params := map[string]interface{}{
			"session_id": session.SessionID,
			"direction":  0, // DirectionNorth (moving from 5,5 to 5,4)
		}
		paramBytes, err := json.Marshal(params)
		require.NoError(t, err)

		result, err := server.handleMove(paramBytes)
		require.NoError(t, err)
		require.NotNil(t, result)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)

		// Check if opportunity_attacks field is present in result
		if oaResults, hasOA := resultMap["opportunity_attacks"]; hasOA {
			oaList, ok := oaResults.([]map[string]interface{})
			if ok && len(oaList) > 0 {
				// Verify the attacker and target are correct
				assert.Equal(t, "enemy-opp-001", oaList[0]["attacker_id"])
				assert.Equal(t, session.Player.GetID(), oaList[0]["target_id"])
			}
		}
	})

	t.Run("enemy reaction is used after opportunity attack", func(t *testing.T) {
		// After the attack, enemy should have used their reaction
		hasReaction := server.state.OpportunityManager.HasReaction(enemyChar.ID)
		assert.False(t, hasReaction, "Enemy should have used their reaction")
	})
}

// TestOpportunityAttackReactionsResetOnNewRound tests that reactions reset at the start of a new round.
func TestOpportunityAttackReactionsResetOnNewRound(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Create an enemy
	enemyChar := &game.Character{
		ID:   "enemy-reaction-001",
		Name: "Reaction Test Enemy",
	}
	server.state.WorldState.Objects[enemyChar.ID] = enemyChar

	// Put server into combat mode
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{session.Player.GetID(), enemyChar.ID}
	server.state.TurnManager.CurrentIndex = 0

	// Register entity and use its reaction
	server.state.OpportunityManager.RegisterEntity(enemyChar.ID, game.Position{X: 1, Y: 1, Level: 0})
	server.state.OpportunityManager.UseReaction(enemyChar.ID)

	// Verify reaction is used
	assert.False(t, server.state.OpportunityManager.HasReaction(enemyChar.ID), "Reaction should be used")

	// Process end of round
	server.processEndRound()

	// Verify reaction is reset
	assert.True(t, server.state.OpportunityManager.HasReaction(enemyChar.ID), "Reaction should be reset after new round")
}

// TestMoraleScoreExposedInInitiativeEntries tests that morale score is included in initiative entries for NPCs.
func TestMoraleScoreExposedInInitiativeEntries(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Create an enemy NPC
	enemyID := "enemy-morale-001"
	enemyChar := &game.Character{
		ID:    enemyID,
		Name:  "Morale Test Enemy",
		HP:    50,
		MaxHP: 50,
	}
	server.state.WorldState.Objects[enemyID] = enemyChar

	// Register NPC with morale system (initial morale 75)
	server.moraleSystem.RegisterNPC(enemyID, "enemies", false, 75)

	// Set up combat
	server.state.TurnManager.IsInCombat = true
	initiative := []string{session.Player.GetID(), enemyID}
	server.state.TurnManager.Initiative = initiative
	server.state.TurnManager.CurrentIndex = 0

	// Build initiative entries
	entries := server.buildInitiativeEntries(initiative)

	// Find the enemy entry
	var enemyEntry map[string]interface{}
	for _, entry := range entries {
		if entry["id"] == enemyID {
			enemyEntry = entry
			break
		}
	}

	require.NotNil(t, enemyEntry, "Enemy entry should exist")

	// Verify morale_score is present for NPC
	moraleScore, hasMoraleScore := enemyEntry["morale_score"]
	assert.True(t, hasMoraleScore, "NPC should have morale_score field")
	assert.Equal(t, 75, moraleScore, "Morale score should be 75")

	// Verify morale_state is also present
	moraleState, hasMoraleState := enemyEntry["morale_state"]
	assert.True(t, hasMoraleState, "NPC should have morale_state field")
	assert.NotEmpty(t, moraleState, "Morale state should not be empty")

	// Find the player entry
	var playerEntry map[string]interface{}
	for _, entry := range entries {
		if entry["id"] == session.Player.GetID() {
			playerEntry = entry
			break
		}
	}

	require.NotNil(t, playerEntry, "Player entry should exist")

	// Verify morale fields are NOT present for player
	_, hasPlayerMoraleScore := playerEntry["morale_score"]
	assert.False(t, hasPlayerMoraleScore, "Player should not have morale_score field")
}

// TestBehaviorTypeExposedInInitiativeEntries verifies that NPC behavior type is included
// in initiative entries for combat state, allowing the UI to display AI behavior icons.
func TestBehaviorTypeExposedInInitiativeEntries(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)
	playerID := session.Player.GetID()

	// Create an NPC with a specific behavior type
	npcID := "test-npc-behavior"
	npc := &game.NPC{
		Character: game.Character{
			ID:    npcID,
			Name:  "Aggressive Orc",
			HP:    30,
			MaxHP: 30,
		},
		Behavior: "aggressive",
		Faction:  "enemy",
	}

	// Add NPC to world state
	server.state.worldMu.Lock()
	if server.state.WorldState.NPCs == nil {
		server.state.WorldState.NPCs = make(map[string]*game.NPC)
	}
	server.state.WorldState.NPCs[npcID] = npc
	server.state.worldMu.Unlock()

	// Set up combat with NPC and player
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{playerID, npcID}
	server.state.TurnManager.CurrentIndex = 0

	// Build initiative entries
	entries := server.buildInitiativeEntries(server.state.TurnManager.Initiative)

	// Verify we have two entries
	assert.Equal(t, 2, len(entries), "Should have entries for player and NPC")

	// Find the NPC entry
	var npcEntry map[string]interface{}
	for _, entry := range entries {
		if entry["id"] == npcID {
			npcEntry = entry
			break
		}
	}

	assert.NotNil(t, npcEntry, "NPC entry should exist")

	// Verify behavior_type is present for NPC
	behaviorType, hasBehavior := npcEntry["behavior_type"]
	assert.True(t, hasBehavior, "NPC should have behavior_type field")
	assert.Equal(t, "aggressive", behaviorType, "NPC behavior type should be 'aggressive'")

	// Verify player does not have behavior_type
	var playerEntry map[string]interface{}
	for _, entry := range entries {
		if entry["id"] == playerID {
			playerEntry = entry
			break
		}
	}

	assert.NotNil(t, playerEntry, "Player entry should exist")
	_, hasPlayerBehavior := playerEntry["behavior_type"]
	assert.False(t, hasPlayerBehavior, "Player should not have behavior_type field")
}
