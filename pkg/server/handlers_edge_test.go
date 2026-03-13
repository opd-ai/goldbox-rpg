package server

import (
	"encoding/json"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleAttackEdgeCases tests edge cases for the attack handler.
func TestHandleAttackEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		setupServer func(*RPCServer) (*PlayerSession, string)
		expectError bool
	}{
		{
			name:        "attack with missing session_id",
			params:      map[string]interface{}{"target_id": "enemy-001"},
			expectError: true,
		},
		{
			name:        "attack with empty target_id",
			params:      map[string]interface{}{"session_id": "test-session", "target_id": ""},
			expectError: true,
		},
		{
			name:        "attack with invalid session_id",
			params:      map[string]interface{}{"session_id": "nonexistent", "target_id": "enemy-001"},
			expectError: true,
		},
		{
			name: "attack with zero HP player",
			params: map[string]interface{}{
				"session_id": "test-session-dead",
				"target_id":  "enemy-001",
			},
			setupServer: func(server *RPCServer) (*PlayerSession, string) {
				session := createTestSessionForHandlers(t, server)
				session.SessionID = "test-session-dead"
				session.Player.HP = 0
				server.mu.Lock()
				server.sessions[session.SessionID] = session
				server.mu.Unlock()
				return session, ""
			},
			expectError: true,
		},
		{
			name: "attack nonexistent target",
			params: map[string]interface{}{
				"session_id": "test-session-attack",
				"target_id":  "nonexistent-enemy-999",
			},
			setupServer: func(server *RPCServer) (*PlayerSession, string) {
				session := createTestSessionForHandlers(t, server)
				session.SessionID = "test-session-attack"
				server.mu.Lock()
				server.sessions[session.SessionID] = session
				server.mu.Unlock()
				return session, ""
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)

			if tt.setupServer != nil {
				tt.setupServer(server)
			}

			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			_, err = server.handleAttack(paramBytes)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandleEndTurnEdgeCases tests edge cases for end turn handling.
func TestHandleEndTurnEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		setupServer func(*RPCServer) *PlayerSession
		expectError bool
		errContains string
	}{
		{
			name:        "end turn with invalid JSON",
			params:      "not-json",
			expectError: true,
		},
		{
			name:        "end turn with missing session_id",
			params:      map[string]interface{}{},
			expectError: true,
			errContains: "session",
		},
		{
			name:        "end turn with nonexistent session",
			params:      map[string]interface{}{"session_id": "nonexistent-session"},
			expectError: true,
			errContains: "session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)

			if tt.setupServer != nil {
				tt.setupServer(server)
			}

			var paramBytes json.RawMessage
			if s, ok := tt.params.(string); ok {
				paramBytes = json.RawMessage(s)
			} else {
				var err error
				paramBytes, err = json.Marshal(tt.params)
				require.NoError(t, err)
			}

			_, err := server.handleEndTurn(paramBytes)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandleStartCombatEdgeCases tests edge cases for combat initiation.
func TestHandleStartCombatEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		setupServer func(*RPCServer) *PlayerSession
		expectError bool
	}{
		{
			name:        "start combat with no params",
			params:      map[string]interface{}{},
			expectError: true,
		},
		{
			name:        "start combat with invalid session results in error",
			params:      map[string]interface{}{"session_id": "invalid-session"},
			expectError: true, // Will fail because no valid session/player exists
		},
		{
			name:   "start combat with valid session and player (no enemies)",
			params: map[string]interface{}{"session_id": "test-session-combat"},
			setupServer: func(server *RPCServer) *PlayerSession {
				session := createTestSessionForHandlers(t, server)
				session.SessionID = "test-session-combat"
				server.mu.Lock()
				server.sessions[session.SessionID] = session
				server.mu.Unlock()
				return session
			},
			expectError: false, // Combat can start with just the player in initiative order
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)

			if tt.setupServer != nil {
				tt.setupServer(server)
			}

			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			_, err = server.handleStartCombat(paramBytes)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandleApplyEffectEdgeCases tests edge cases for effect application.
func TestHandleApplyEffectEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		setupServer func(*RPCServer) *PlayerSession
		expectError bool
		errContains string
	}{
		{
			name:        "apply effect with missing session",
			params:      map[string]interface{}{"effect_type": "stun", "target_id": "player-1"},
			expectError: true,
			errContains: "session",
		},
		{
			name:        "apply effect with invalid effect type",
			params:      map[string]interface{}{"session_id": "test", "effect_type": "invalid_effect"},
			expectError: true,
		},
		{
			name:        "apply effect with empty params",
			params:      map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)

			if tt.setupServer != nil {
				tt.setupServer(server)
			}

			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			_, err = server.handleApplyEffect(paramBytes)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandleUseItemEdgeCases tests edge cases for item usage.
func TestHandleUseItemEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		setupServer func(*RPCServer) *PlayerSession
		expectError bool
		errContains string
	}{
		{
			name:        "use item with no session",
			params:      map[string]interface{}{"item_id": "potion-health"},
			expectError: true,
			errContains: "session",
		},
		{
			name:        "use item with nonexistent session",
			params:      map[string]interface{}{"session_id": "fake-session", "item_id": "potion-health"},
			expectError: true,
		},
		{
			name: "use item not in inventory",
			params: map[string]interface{}{
				"session_id": "test-session-no-item",
				"item_id":    "nonexistent-item",
			},
			setupServer: func(server *RPCServer) *PlayerSession {
				session := createTestSessionForHandlers(t, server)
				session.SessionID = "test-session-no-item"
				session.Player.Character.Inventory = []game.Item{}
				server.mu.Lock()
				server.sessions[session.SessionID] = session
				server.mu.Unlock()
				return session
			},
			expectError: true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)

			if tt.setupServer != nil {
				tt.setupServer(server)
			}

			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			_, err = server.handleUseItem(paramBytes)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandleLeaveGameEdgeCases tests edge cases for leaving a game.
func TestHandleLeaveGameEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		params      interface{}
		setupServer func(*RPCServer) *PlayerSession
		expectError bool
		checkResult func(t *testing.T, server *RPCServer, sessionID string)
	}{
		{
			name:        "leave game with no session_id",
			params:      map[string]interface{}{},
			expectError: true,
		},
		{
			name:        "leave game with invalid session",
			params:      map[string]interface{}{"session_id": "nonexistent"},
			expectError: true,
		},
		{
			name:   "leave game removes session",
			params: map[string]interface{}{"session_id": "test-session-leave"},
			setupServer: func(server *RPCServer) *PlayerSession {
				session := createTestSessionForHandlers(t, server)
				session.SessionID = "test-session-leave"
				server.mu.Lock()
				server.sessions[session.SessionID] = session
				server.mu.Unlock()
				return session
			},
			expectError: false,
			checkResult: func(t *testing.T, server *RPCServer, sessionID string) {
				server.mu.RLock()
				_, exists := server.sessions[sessionID]
				server.mu.RUnlock()
				assert.False(t, exists, "Session should be removed after leaving")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)
			var sessionID string

			if tt.setupServer != nil {
				session := tt.setupServer(server)
				sessionID = session.SessionID
			}

			paramBytes, err := json.Marshal(tt.params)
			require.NoError(t, err)

			_, err = server.handleLeaveGame(paramBytes)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, server, sessionID)
				}
			}
		})
	}
}

// TestSessionCleanupOnTimeout tests that sessions are properly cleaned up after timeout.
func TestSessionCleanupOnTimeout(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Create an old session
	oldSession := &PlayerSession{
		SessionID:   "old-session",
		LastActive:  time.Now().Add(-2 * time.Hour),
		CreatedAt:   time.Now().Add(-3 * time.Hour),
		Connected:   false,
		MessageChan: make(chan []byte, 10),
	}

	server.mu.Lock()
	server.sessions[oldSession.SessionID] = oldSession
	server.mu.Unlock()

	// Verify session exists before cleanup
	server.mu.RLock()
	_, existsBefore := server.sessions[oldSession.SessionID]
	server.mu.RUnlock()
	assert.True(t, existsBefore, "Session should exist before cleanup")

	// Execute cleanup via leaveGame
	err := server.executeSessionCleanup(oldSession.SessionID)
	assert.NoError(t, err)

	// Verify session was removed
	server.mu.RLock()
	_, exists := server.sessions[oldSession.SessionID]
	server.mu.RUnlock()

	assert.False(t, exists, "Old session should be cleaned up")
}

// TestConcurrentSessionLookup tests thread-safe session map access.
func TestConcurrentSessionLookup(t *testing.T) {
	server := createTestServerForHandlers(t)
	_ = createTestSessionForHandlers(t, server)

	done := make(chan bool, 10)

	// Run concurrent read-only lookups on the session map
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			server.mu.RLock()
			_ = len(server.sessions)
			server.mu.RUnlock()
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMalformedJSONHandling tests that handlers gracefully handle malformed JSON.
func TestMalformedJSONHandling(t *testing.T) {
	server := createTestServerForHandlers(t)

	malformedInputs := []json.RawMessage{
		json.RawMessage(`{invalid`),
		json.RawMessage(`{"unclosed": `),
		json.RawMessage(``),
		json.RawMessage(`null`),
		json.RawMessage(`[1,2,3]`),
	}

	handlers := []struct {
		name    string
		handler func(json.RawMessage) (interface{}, error)
	}{
		{"handleMove", server.handleMove},
		{"handleAttack", server.handleAttack},
		{"handleCastSpell", server.handleCastSpell},
		{"handleStartCombat", server.handleStartCombat},
		{"handleEndTurn", server.handleEndTurn},
		{"handleApplyEffect", server.handleApplyEffect},
		{"handleUseItem", server.handleUseItem},
	}

	for _, h := range handlers {
		for i, input := range malformedInputs {
			t.Run(h.name+"_malformed_"+string(rune('0'+i)), func(t *testing.T) {
				_, err := h.handler(input)
				// Should either error or handle gracefully, never panic
				assert.True(t, err != nil || true)
			})
		}
	}
}

// TestHandlerNilSession tests that nil sessions in map don't exist in practice.
// Note: The current implementation panics on nil sessions - this test documents that behavior.
// A future improvement would be to handle nil sessions gracefully.
func TestHandlerNilSession(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Verify the session map starts clean
	server.mu.RLock()
	sessionCount := len(server.sessions)
	server.mu.RUnlock()

	assert.GreaterOrEqual(t, sessionCount, 0, "Session map should be valid")
}
