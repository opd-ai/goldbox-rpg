package server

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionError(t *testing.T) {
	sessionID := "sess-456"
	baseErr := ErrSessionExpired
	err := NewSessionError(sessionID, "validate", baseErr)

	// Test error message
	assert.Contains(t, err.Error(), sessionID)
	assert.Contains(t, err.Error(), "validate")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var sessErr *SessionError
	require.True(t, errors.As(err, &sessErr))
	assert.Equal(t, sessionID, sessErr.SessionID)
	assert.Equal(t, "validate", sessErr.Operation)
}

func TestValidationError(t *testing.T) {
	method := "handleMove"
	param := "targetX"
	value := -5
	baseErr := ErrInvalidParams
	err := NewValidationError(method, param, value, baseErr)

	// Test error message
	assert.Contains(t, err.Error(), method)
	assert.Contains(t, err.Error(), param)
	assert.Contains(t, err.Error(), "-5")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))
	assert.Equal(t, method, valErr.Method)
	assert.Equal(t, param, valErr.Parameter)
	assert.Equal(t, value, valErr.Value)
}

func TestPersistenceError(t *testing.T) {
	sessionID := "sess-789"
	filePath := "/data/session.yaml"
	baseErr := ErrSaveFailed
	err := NewPersistenceError(sessionID, filePath, "save", baseErr)

	// Test error message
	assert.Contains(t, err.Error(), sessionID)
	assert.Contains(t, err.Error(), filePath)
	assert.Contains(t, err.Error(), "save")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var persErr *PersistenceError
	require.True(t, errors.As(err, &persErr))
	assert.Equal(t, sessionID, persErr.SessionID)
	assert.Equal(t, filePath, persErr.FilePath)
}

func TestHealthCheckError(t *testing.T) {
	component := "spell_manager"
	check := "initialization"
	baseErr := ErrSpellManagerNil
	err := NewHealthCheckError(component, check, baseErr)

	// Test error message
	assert.Contains(t, err.Error(), component)
	assert.Contains(t, err.Error(), check)

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var hcErr *HealthCheckError
	require.True(t, errors.As(err, &hcErr))
	assert.Equal(t, component, hcErr.Component)
	assert.Equal(t, check, hcErr.Check)
}

func TestRPCError(t *testing.T) {
	method := "getState"
	requestID := 42
	baseErr := ErrInvalidSession
	err := NewRPCError(method, requestID, baseErr)

	// Test error message
	assert.Contains(t, err.Error(), method)
	assert.Contains(t, err.Error(), "42")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var rpcErr *RPCError
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, method, rpcErr.Method)
	assert.Equal(t, requestID, rpcErr.RequestID)
}

func TestWebSocketError(t *testing.T) {
	clientID := "ws-001"
	baseErr := ErrWebSocketClosed
	err := NewWebSocketError(clientID, "broadcast", baseErr)

	// Test error message
	assert.Contains(t, err.Error(), clientID)
	assert.Contains(t, err.Error(), "broadcast")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var wsErr *WebSocketError
	require.True(t, errors.As(err, &wsErr))
	assert.Equal(t, clientID, wsErr.ClientID)
	assert.Equal(t, "broadcast", wsErr.Operation)
}

func TestErrorChaining(t *testing.T) {
	// Test error chain: RPCError -> SessionError -> ErrSessionExpired
	sessErr := NewSessionError("sess-123", "lookup", ErrSessionExpired)
	rpcErr := NewRPCError("handleMove", 1, sessErr)

	// Should be able to unwrap through chain
	assert.True(t, errors.Is(rpcErr, ErrSessionExpired))

	// Should be able to extract both error types
	var se *SessionError
	require.True(t, errors.As(rpcErr, &se))
	assert.Equal(t, "sess-123", se.SessionID)

	var re *RPCError
	require.True(t, errors.As(rpcErr, &re))
	assert.Equal(t, "handleMove", re.Method)
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrSessionNotFound", ErrSessionNotFound},
		{"ErrSessionExpired", ErrSessionExpired},
		{"ErrSessionCreation", ErrSessionCreation},
		{"ErrInvalidRequest", ErrInvalidRequest},
		{"ErrInvalidParams", ErrInvalidParams},
		{"ErrMissingParams", ErrMissingParams},
		{"ErrInvalidMethod", ErrInvalidMethod},
		{"ErrGameStateNotFound", ErrGameStateNotFound},
		{"ErrGameStateCorrupted", ErrGameStateCorrupted},
		{"ErrGameStateNil", ErrGameStateNil},
		{"ErrServerShuttingDown", ErrServerShuttingDown},
		{"ErrServerNotReady", ErrServerNotReady},
		{"ErrServerNil", ErrServerNil},
		{"ErrWorldNil", ErrWorldNil},
		{"ErrSpellManagerNil", ErrSpellManagerNil},
		{"ErrEventSystemNil", ErrEventSystemNil},
		{"ErrPCGManagerNil", ErrPCGManagerNil},
		{"ErrCircuitBreakerNil", ErrCircuitBreakerNil},
		{"ErrConfigNil", ErrConfigNil},
		{"ErrPersistenceFailed", ErrPersistenceFailed},
		{"ErrLoadFailed", ErrLoadFailed},
		{"ErrSaveFailed", ErrSaveFailed},
		{"ErrWebSocketClosed", ErrWebSocketClosed},
		{"ErrWebSocketUpgrade", ErrWebSocketUpgrade},
		{"ErrInvalidOrigin", ErrInvalidOrigin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All sentinel errors should be non-nil and have a message
			assert.NotNil(t, tt.err)
			assert.NotEmpty(t, tt.err.Error())

			// Each sentinel error should match itself
			assert.True(t, errors.Is(tt.err, tt.err))
		})
	}
}

func TestCategorizeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: ErrorCategoryUnknown,
		},
		{
			name:     "session error typed",
			err:      NewSessionError("sess-123", "operation", ErrSessionExpired),
			expected: ErrorCategorySession,
		},
		{
			name:     "validation error typed",
			err:      NewValidationError("move", "param", 5, ErrInvalidParams),
			expected: ErrorCategoryValidation,
		},
		{
			name:     "persistence error typed",
			err:      NewPersistenceError("sess-456", "/data/file", "save", ErrSaveFailed),
			expected: ErrorCategoryPersistence,
		},
		{
			name:     "health check error typed",
			err:      NewHealthCheckError("database", "ping", errors.New("connection refused")),
			expected: ErrorCategoryHealth,
		},
		{
			name:     "rpc error typed",
			err:      NewRPCError("handleMove", "req-123", ErrInvalidRequest),
			expected: ErrorCategoryRPC,
		},
		{
			name:     "websocket error typed",
			err:      NewWebSocketError("client-789", "send", ErrWebSocketClosed),
			expected: ErrorCategoryWebSocket,
		},
		{
			name:     "session sentinel error",
			err:      ErrSessionExpired,
			expected: ErrorCategorySession,
		},
		{
			name:     "validation sentinel error",
			err:      ErrInvalidParams,
			expected: ErrorCategoryValidation,
		},
		{
			name:     "persistence sentinel error",
			err:      ErrPersistenceFailed,
			expected: ErrorCategoryPersistence,
		},
		{
			name:     "websocket sentinel error",
			err:      ErrWebSocketUpgrade,
			expected: ErrorCategoryWebSocket,
		},
		{
			name:     "internal error",
			err:      ErrServerNil,
			expected: ErrorCategoryInternal,
		},
		{
			name:     "unknown error",
			err:      errors.New("some random error"),
			expected: ErrorCategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := CategorizeError(tt.err)
			assert.Equal(t, tt.expected, category, 
				"Expected category %s for error %v, got %s", 
				tt.expected, tt.err, category)
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test that wrapped errors maintain Is/As relationships
	baseErr := ErrSessionExpired
	wrapped := NewSessionError("sess-123", "validate", baseErr)
	doubleWrapped := NewRPCError("handleMove", "req-456", wrapped)

	// errors.Is should find the sentinel error through the chain
	assert.True(t, errors.Is(doubleWrapped, baseErr),
		"errors.Is should find base error through wrapping chain")

	// errors.As should extract SessionError from RPC wrapper
	var sessErr *SessionError
	require.True(t, errors.As(doubleWrapped, &sessErr),
		"errors.As should extract SessionError from wrapped chain")
	assert.Equal(t, "sess-123", sessErr.SessionID)

	// errors.As should also extract RPCError
	var rpcErr *RPCError
	require.True(t, errors.As(doubleWrapped, &rpcErr),
		"errors.As should extract RPCError from chain")
	assert.Equal(t, "handleMove", rpcErr.Method)
}

func TestErrorWrappingPreservesContext(t *testing.T) {
	// Original error with context
	originalErr := NewValidationError("handleAttack", "targetID", "", ErrMissingParams)

	// Wrap it in RPC error
	wrappedErr := NewRPCError("handleAttack", 123, originalErr)

	// Should preserve both error types
	var valErr *ValidationError
	require.True(t, errors.As(wrappedErr, &valErr))
	assert.Equal(t, "handleAttack", valErr.Method)
	assert.Equal(t, "targetID", valErr.Parameter)

	var rpcErr *RPCError
	require.True(t, errors.As(wrappedErr, &rpcErr))
	assert.Equal(t, "handleAttack", rpcErr.Method)
	assert.Equal(t, 123, rpcErr.RequestID)

	// Should find base error
	assert.True(t, errors.Is(wrappedErr, ErrMissingParams))
}
