package server

import (
	"encoding/json"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/gorilla/websocket"
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
	character.Position = game.Position{X: 10, Y: 10, Level: 0}

	player := &game.Player{
		Character: *character,
	}

	session := &PlayerSession{
		SessionID:   "test-session-001",
		Player:      player,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   true,
		MessageChan: make(chan []byte, 500),
		WSConn:      &websocket.Conn{}, // Mock WebSocket connection for tests
	}

	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()
	
	// Initialize world bounds if not set
	if server.state.WorldState.Width == 0 {
		server.state.WorldState.Width = 100
		server.state.WorldState.Height = 100
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
				assert.Equal(t, "move successful", resultMap["message"])

				// Check that Y position decreased (north = -Y)
				pos := session.Player.Character.GetPosition()
				assert.Equal(t, 9, pos.Y)
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
			name: "missing direction",
			params: map[string]interface{}{
				"session_id": "test-session-001",
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				return createTestSessionForHandlers(t, server)
			},
			expectError: true,
		},
		{
			name: "invalid direction (out of range)",
			params: map[string]interface{}{
				"session_id": "test-session-001",
				"direction":  999, // Invalid direction value
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				return createTestSessionForHandlers(t, server)
			},
			expectError: true,
		},
		{
			name: "insufficient action points",
			params: map[string]interface{}{
				"session_id": "test-session-001",
				"direction":  0, // DirectionNorth
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				session := createTestSessionForHandlers(t, server)
				// Set action points to zero
				session.Player.Character.SetActionPoints(0)
				return session
			},
			expectError: true,
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
				assert.Equal(t, "TestPlayer123", resultMap["player_name"])
			},
		},
		{
			name: "valid join generates default name",
			params: map[string]interface{}{
				"player_name": "",
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, resultMap["session_id"])
				// Default name starts with "Player-"
				playerName, ok := resultMap["player_name"].(string)
				assert.True(t, ok)
				assert.Contains(t, playerName, "Player-")
			},
		},
		{
			name: "missing player name uses default",
			params: map[string]interface{}{
				"other_field": "value",
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				playerName, ok := resultMap["player_name"].(string)
				assert.True(t, ok)
				assert.Contains(t, playerName, "Player-")
			},
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
			setupServer:  func(server *RPCServer) {},
			expectError:  true,
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
			name: "valid character creation with standard array",
			params: map[string]interface{}{
				"name":   "Warrior",
				"class":  "fighter",
				"method": "standard_array",
				"attributes": map[string]interface{}{
					"strength":     15,
					"dexterity":    14,
					"constitution": 13,
					"intelligence": 12,
					"wisdom":       10,
					"charisma":     8,
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, resultMap["session_id"])
				
				character, ok := resultMap["character"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Warrior", character["name"])
			},
		},
		{
			name: "missing character name",
			params: map[string]interface{}{
				"class":  "fighter",
				"method": "standard_array",
			},
			expectError: true,
		},
		{
			name: "invalid class",
			params: map[string]interface{}{
				"name":   "Test",
				"class":  "invalid_class",
				"method": "standard_array",
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
			name:        "valid weapon main slot",
			slotName:    "weapon-main",
			expected:    game.SlotWeaponMain,
			expectError: false,
		},
		{
			name:        "case insensitive",
			slotName:    "HEAD",
			expected:    game.SlotHead,
			expectError: false,
		},
		{
			name:        "whitespace trimmed",
			slotName:    "  head  ",
			expected:    game.SlotHead,
			expectError: false,
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
			expected: "weapon-main",
		},
		{
			name:     "weapon off slot",
			slot:     game.SlotWeaponOff,
			expected: "weapon-off",
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
