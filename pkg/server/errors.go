package server

import (
	"errors"
	"fmt"
)

// ErrInvalidSession is defined in handlers.go for backward compatibility
// Keeping the definition here as a reference sentinel error

// Sentinel errors for server operations
var (
	// Session errors (ErrInvalidSession is in handlers.go)
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionCreation = errors.New("session creation failed")

	// Request validation errors
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidParams  = errors.New("invalid parameters")
	ErrMissingParams  = errors.New("missing required parameters")
	ErrInvalidMethod  = errors.New("invalid method")

	// Game state errors
	ErrGameStateNotFound  = errors.New("game state not found")
	ErrGameStateCorrupted = errors.New("game state corrupted")
	ErrGameStateNil       = errors.New("game state is nil")

	// Server health errors
	ErrServerShuttingDown = errors.New("server is shutting down")
	ErrServerNotReady     = errors.New("server not ready")
	ErrServerNil          = errors.New("server instance is nil")

	// Component initialization errors
	ErrWorldNil          = errors.New("world state is nil")
	ErrSpellManagerNil   = errors.New("spell manager not initialized")
	ErrEventSystemNil    = errors.New("event system not initialized")
	ErrPCGManagerNil     = errors.New("PCG manager not initialized")
	ErrCircuitBreakerNil = errors.New("circuit breaker manager not initialized")
	ErrConfigNil         = errors.New("configuration not initialized")

	// Persistence errors
	ErrPersistenceFailed = errors.New("persistence operation failed")
	ErrLoadFailed        = errors.New("load operation failed")
	ErrSaveFailed        = errors.New("save operation failed")

	// WebSocket errors
	ErrWebSocketClosed  = errors.New("websocket connection closed")
	ErrWebSocketUpgrade = errors.New("websocket upgrade failed")
	ErrInvalidOrigin    = errors.New("invalid websocket origin")
)

// SessionError represents errors related to session management
type SessionError struct {
	SessionID string
	Operation string
	Err       error
}

func (e *SessionError) Error() string {
	if e.SessionID != "" {
		return fmt.Sprintf("session %s: %s: %v", e.SessionID, e.Operation, e.Err)
	}
	return fmt.Sprintf("session: %s: %v", e.Operation, e.Err)
}

func (e *SessionError) Unwrap() error {
	return e.Err
}

// NewSessionError creates a new SessionError
func NewSessionError(sessionID, operation string, err error) *SessionError {
	return &SessionError{
		SessionID: sessionID,
		Operation: operation,
		Err:       err,
	}
}

// ValidationError represents request validation errors
type ValidationError struct {
	Method    string
	Parameter string
	Value     interface{}
	Err       error
}

func (e *ValidationError) Error() string {
	if e.Parameter != "" {
		return fmt.Sprintf("validation error: method %s: parameter %s (value: %v): %v",
			e.Method, e.Parameter, e.Value, e.Err)
	}
	return fmt.Sprintf("validation error: method %s: %v", e.Method, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new ValidationError
func NewValidationError(method, parameter string, value interface{}, err error) *ValidationError {
	return &ValidationError{
		Method:    method,
		Parameter: parameter,
		Value:     value,
		Err:       err,
	}
}

// PersistenceError represents errors during persistence operations
type PersistenceError struct {
	SessionID string
	FilePath  string
	Operation string
	Err       error
}

func (e *PersistenceError) Error() string {
	if e.SessionID != "" && e.FilePath != "" {
		return fmt.Sprintf("persistence: %s: session %s: file %s: %v",
			e.Operation, e.SessionID, e.FilePath, e.Err)
	}
	if e.SessionID != "" {
		return fmt.Sprintf("persistence: %s: session %s: %v", e.Operation, e.SessionID, e.Err)
	}
	return fmt.Sprintf("persistence: %s: %v", e.Operation, e.Err)
}

func (e *PersistenceError) Unwrap() error {
	return e.Err
}

// NewPersistenceError creates a new PersistenceError
func NewPersistenceError(sessionID, filePath, operation string, err error) *PersistenceError {
	return &PersistenceError{
		SessionID: sessionID,
		FilePath:  filePath,
		Operation: operation,
		Err:       err,
	}
}

// HealthCheckError represents errors during health check operations
type HealthCheckError struct {
	Component string
	Check     string
	Err       error
}

func (e *HealthCheckError) Error() string {
	return fmt.Sprintf("health check: %s: %s: %v", e.Component, e.Check, e.Err)
}

func (e *HealthCheckError) Unwrap() error {
	return e.Err
}

// NewHealthCheckError creates a new HealthCheckError
func NewHealthCheckError(component, check string, err error) *HealthCheckError {
	return &HealthCheckError{
		Component: component,
		Check:     check,
		Err:       err,
	}
}

// RPCError represents errors in RPC request processing
type RPCError struct {
	Method    string
	RequestID interface{}
	Err       error
}

func (e *RPCError) Error() string {
	if e.RequestID != nil {
		return fmt.Sprintf("RPC: method %s: request %v: %v", e.Method, e.RequestID, e.Err)
	}
	return fmt.Sprintf("RPC: method %s: %v", e.Method, e.Err)
}

func (e *RPCError) Unwrap() error {
	return e.Err
}

// NewRPCError creates a new RPCError
func NewRPCError(method string, requestID interface{}, err error) *RPCError {
	return &RPCError{
		Method:    method,
		RequestID: requestID,
		Err:       err,
	}
}

// WebSocketError represents errors in WebSocket operations
type WebSocketError struct {
	ClientID  string
	Operation string
	Err       error
}

func (e *WebSocketError) Error() string {
	if e.ClientID != "" {
		return fmt.Sprintf("websocket: client %s: %s: %v", e.ClientID, e.Operation, e.Err)
	}
	return fmt.Sprintf("websocket: %s: %v", e.Operation, e.Err)
}

func (e *WebSocketError) Unwrap() error {
	return e.Err
}

// NewWebSocketError creates a new WebSocketError
func NewWebSocketError(clientID, operation string, err error) *WebSocketError {
	return &WebSocketError{
		ClientID:  clientID,
		Operation: operation,
		Err:       err,
	}
}

// ErrorCategory represents categories of errors for monitoring and alerting
type ErrorCategory string

const (
	// ErrorCategoryValidation represents input validation errors
	ErrorCategoryValidation ErrorCategory = "validation"
	// ErrorCategorySession represents session management errors
	ErrorCategorySession ErrorCategory = "session"
	// ErrorCategoryPersistence represents data persistence errors
	ErrorCategoryPersistence ErrorCategory = "persistence"
	// ErrorCategoryHealth represents health check errors
	ErrorCategoryHealth ErrorCategory = "health"
	// ErrorCategoryRPC represents RPC protocol errors
	ErrorCategoryRPC ErrorCategory = "rpc"
	// ErrorCategoryWebSocket represents WebSocket errors
	ErrorCategoryWebSocket ErrorCategory = "websocket"
	// ErrorCategoryInternal represents internal server errors
	ErrorCategoryInternal ErrorCategory = "internal"
	// ErrorCategoryUnknown represents uncategorized errors
	ErrorCategoryUnknown ErrorCategory = "unknown"
)

// CategorizeError returns the error category for monitoring purposes
func CategorizeError(err error) ErrorCategory {
	if err == nil {
		return ErrorCategoryUnknown
	}

	// Check for typed errors using errors.As
	var sessErr *SessionError
	if errors.As(err, &sessErr) {
		return ErrorCategorySession
	}

	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return ErrorCategoryValidation
	}

	var persErr *PersistenceError
	if errors.As(err, &persErr) {
		return ErrorCategoryPersistence
	}

	var hcErr *HealthCheckError
	if errors.As(err, &hcErr) {
		return ErrorCategoryHealth
	}

	var rpcErr *RPCError
	if errors.As(err, &rpcErr) {
		return ErrorCategoryRPC
	}

	var wsErr *WebSocketError
	if errors.As(err, &wsErr) {
		return ErrorCategoryWebSocket
	}

	// Check for sentinel errors using errors.Is
	switch {
	case errors.Is(err, ErrInvalidSession),
		errors.Is(err, ErrSessionNotFound),
		errors.Is(err, ErrSessionExpired),
		errors.Is(err, ErrSessionCreation):
		return ErrorCategorySession

	case errors.Is(err, ErrInvalidRequest),
		errors.Is(err, ErrInvalidParams),
		errors.Is(err, ErrMissingParams),
		errors.Is(err, ErrInvalidMethod):
		return ErrorCategoryValidation

	case errors.Is(err, ErrPersistenceFailed),
		errors.Is(err, ErrLoadFailed),
		errors.Is(err, ErrSaveFailed):
		return ErrorCategoryPersistence

	case errors.Is(err, ErrWebSocketClosed),
		errors.Is(err, ErrWebSocketUpgrade),
		errors.Is(err, ErrInvalidOrigin):
		return ErrorCategoryWebSocket

	case errors.Is(err, ErrServerShuttingDown),
		errors.Is(err, ErrServerNotReady),
		errors.Is(err, ErrGameStateCorrupted),
		errors.Is(err, ErrGameStateNil),
		errors.Is(err, ErrServerNil),
		errors.Is(err, ErrWorldNil),
		errors.Is(err, ErrSpellManagerNil),
		errors.Is(err, ErrEventSystemNil),
		errors.Is(err, ErrPCGManagerNil),
		errors.Is(err, ErrCircuitBreakerNil),
		errors.Is(err, ErrConfigNil):
		return ErrorCategoryInternal

	default:
		return ErrorCategoryUnknown
	}
}
