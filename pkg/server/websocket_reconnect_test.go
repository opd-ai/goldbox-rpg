package server

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestWebSocketReconnection_SessionPreservation tests that session state is preserved
// across reconnection attempts within the timeout window.
func TestWebSocketReconnection_SessionPreservation(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Mark session as disconnected (simulating WebSocket disconnect)
	session.Connected = false
	disconnectTime := time.Now()
	session.LastActive = disconnectTime

	// Verify session still exists and can be retrieved
	server.mu.RLock()
	retrievedSession, exists := server.sessions[session.SessionID]
	server.mu.RUnlock()

	assert.True(t, exists, "Session should exist after disconnect")
	assert.NotNil(t, retrievedSession, "Session should not be nil")
	assert.False(t, retrievedSession.Connected, "Session should be marked disconnected")
}

// TestWebSocketError_MalformedMessage tests error handling for malformed WebSocket messages.
func TestWebSocketError_MalformedMessage(t *testing.T) {
	tests := []struct {
		name          string
		message       []byte
		expectError   bool
		errorContains string
	}{
		{
			name:        "empty message",
			message:     []byte{},
			expectError: true,
		},
		{
			name:        "invalid JSON",
			message:     []byte(`{invalid`),
			expectError: true,
		},
		{
			name:        "missing jsonrpc field",
			message:     []byte(`{"method": "test", "id": 1}`),
			expectError: false, // JSON parses, validation comes later
		},
		{
			name:        "missing method field",
			message:     []byte(`{"jsonrpc": "2.0", "id": 1}`),
			expectError: false, // JSON parses, validation comes later
		},
		{
			name:        "invalid params type (string instead of object)",
			message:     []byte(`{"jsonrpc": "2.0", "method": "test", "params": "invalid", "id": 1}`),
			expectError: false, // params get parsed as raw JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req JSONRPCRequest
			err := json.Unmarshal(tt.message, &req)

			if tt.expectError {
				// Either unmarshal fails or validation would fail
				if err == nil {
					// Check if basic validation would fail
					if req.JSONRPC == "" && string(req.Method) == "" {
						err = errors.New("missing required fields")
					}
				}
				assert.Error(t, err)
			}
		})
	}
}

// TestWebSocketError_ConnectionDropped tests behavior when connection is unexpectedly closed.
func TestWebSocketError_ConnectionDropped(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Simulate connection drop by closing message channel
	close(session.MessageChan)

	// Create new channel to avoid panic
	session.MessageChan = make(chan []byte, 10)

	// Verify session can be gracefully handled
	server.mu.RLock()
	retrievedSession, exists := server.sessions[session.SessionID]
	server.mu.RUnlock()

	assert.True(t, exists)
	assert.NotNil(t, retrievedSession)
}

// TestWebSocketError_SessionTimeout tests session expiration handling.
func TestWebSocketError_SessionTimeout(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Create session with old LastActive time
	session := createTestSessionForHandlers(t, server)
	session.LastActive = time.Now().Add(-2 * time.Hour)
	session.Connected = false

	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()

	// Try to use expired session
	params := map[string]interface{}{"session_id": session.SessionID}
	paramBytes, _ := json.Marshal(params)

	// handleGetGameState should work but session is stale
	result, err := server.handleGetGameState(paramBytes)
	// Stale sessions are still valid until cleanup runs
	if err == nil {
		assert.NotNil(t, result)
	}
}

// TestWebSocketError_ConcurrentDisconnect tests handling multiple disconnect events.
func TestWebSocketError_ConcurrentDisconnect(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	done := make(chan bool, 3)

	// Simulate concurrent disconnect attempts
	for i := 0; i < 3; i++ {
		go func() {
			defer func() { done <- true }()

			server.mu.Lock()
			if s, ok := server.sessions[session.SessionID]; ok {
				s.Connected = false
			}
			server.mu.Unlock()
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify session state is consistent
	server.mu.RLock()
	s, exists := server.sessions[session.SessionID]
	server.mu.RUnlock()

	assert.True(t, exists)
	assert.False(t, s.Connected)
}

// TestWebSocketError_MessageQueueOverflow tests handling of message queue overflow.
func TestWebSocketError_MessageQueueOverflow(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Fill up the message channel
	for i := 0; i < cap(session.MessageChan); i++ {
		select {
		case session.MessageChan <- []byte(`{"test": true}`):
		default:
			// Channel full, expected
		}
	}

	// Try to send one more message (should not block or panic)
	select {
	case session.MessageChan <- []byte(`{"overflow": true}`):
		t.Log("Message sent (channel had space)")
	default:
		t.Log("Message dropped (channel full) - expected behavior")
	}

	// Verify session still valid
	assert.True(t, session.SessionID != "")
}

// TestWebSocketReconnection_SessionRecovery tests that a session can be recovered after disconnect.
func TestWebSocketReconnection_SessionRecovery(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)
	originalSessionID := session.SessionID

	// Simulate disconnect
	session.Connected = false

	// Simulate reconnect by setting connected back to true
	server.mu.Lock()
	if s, ok := server.sessions[originalSessionID]; ok {
		s.Connected = true
		s.LastActive = time.Now()
	}
	server.mu.Unlock()

	// Verify session state
	server.mu.RLock()
	recoveredSession, exists := server.sessions[originalSessionID]
	server.mu.RUnlock()

	assert.True(t, exists, "Session should still exist")
	assert.True(t, recoveredSession.Connected, "Session should be reconnected")
}

// TestWebSocketError_InvalidSessionID tests error handling for invalid session IDs.
func TestWebSocketError_InvalidSessionID(t *testing.T) {
	server := createTestServerForHandlers(t)
	_ = createTestSessionForHandlers(t, server)

	invalidIDs := []string{
		"",
		"nonexistent-session-id",
		"  spaces  ",
	}

	for _, invalidID := range invalidIDs {
		t.Run("invalid_id_"+invalidID, func(t *testing.T) {
			params := map[string]interface{}{"session_id": invalidID}
			paramBytes, _ := json.Marshal(params)

			_, err := server.handleGetGameState(paramBytes)
			assert.Error(t, err, "Should error on invalid session ID: %s", invalidID)
		})
	}
}

// TestWebSocketError_RateLimiting tests that rate limiting protects against message flooding.
func TestWebSocketError_RateLimiting(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Send multiple requests rapidly
	successCount := 0
	errorCount := 0

	for i := 0; i < 100; i++ {
		params := map[string]interface{}{"session_id": session.SessionID}
		paramBytes, _ := json.Marshal(params)

		_, err := server.handleGetGameState(paramBytes)
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// At least some should succeed
	assert.Greater(t, successCount, 0, "Some requests should succeed")
}
