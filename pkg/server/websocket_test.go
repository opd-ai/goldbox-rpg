package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestNewResponse tests the NewResponse function with various input types
func TestNewResponse(t *testing.T) {
	tests := []struct {
		name     string
		id       interface{}
		result   interface{}
		expected map[string]interface{}
	}{
		{
			name:   "string id and result",
			id:     "test-id",
			result: "success",
			expected: map[string]interface{}{
				"jsonrpc": "2.0",
				"result":  "success",
				"id":      "test-id",
			},
		},
		{
			name:   "numeric id with map result",
			id:     123,
			result: map[string]string{"status": "ok"},
			expected: map[string]interface{}{
				"jsonrpc": "2.0",
				"result":  map[string]string{"status": "ok"},
				"id":      123,
			},
		},
		{
			name:   "nil id with slice result",
			id:     nil,
			result: []string{"item1", "item2"},
			expected: map[string]interface{}{
				"jsonrpc": "2.0",
				"result":  []string{"item1", "item2"},
				"id":      nil,
			},
		},
		{
			name:   "float id with nil result",
			id:     123.45,
			result: nil,
			expected: map[string]interface{}{
				"jsonrpc": "2.0",
				"result":  nil,
				"id":      123.45,
			},
		},
		{
			name: "complex object result",
			id:   "complex-test",
			result: struct {
				Name string
				Age  int
			}{"John", 30},
			expected: map[string]interface{}{
				"jsonrpc": "2.0",
				"result": struct {
					Name string
					Age  int
				}{"John", 30},
				"id": "complex-test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewResponse(tt.id, tt.result)

			// Type assertion to map[string]interface{}
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatal("NewResponse should return map[string]interface{}")
			}

			if !reflect.DeepEqual(tt.expected, resultMap) {
				t.Errorf("Expected %v, got %v", tt.expected, resultMap)
			}

			// Verify JSON-RPC 2.0 compliance
			if resultMap["jsonrpc"] != "2.0" {
				t.Errorf("Expected jsonrpc to be '2.0', got %v", resultMap["jsonrpc"])
			}
			if !reflect.DeepEqual(tt.id, resultMap["id"]) {
				t.Errorf("Expected id %v, got %v", tt.id, resultMap["id"])
			}
			if !reflect.DeepEqual(tt.result, resultMap["result"]) {
				t.Errorf("Expected result %v, got %v", tt.result, resultMap["result"])
			}
		})
	}
}

// TestNewErrorResponse tests the NewErrorResponse function with various error types
func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name        string
		id          interface{}
		err         error
		expectedErr map[string]interface{}
	}{
		{
			name: "simple error with string id",
			id:   "error-test",
			err:  errors.New("test error message"),
			expectedErr: map[string]interface{}{
				"code":    -32000,
				"message": "test error message",
			},
		},
		{
			name: "error with numeric id",
			id:   456,
			err:  errors.New("database connection failed"),
			expectedErr: map[string]interface{}{
				"code":    -32000,
				"message": "database connection failed",
			},
		},
		{
			name: "error with nil id",
			id:   nil,
			err:  errors.New("unauthorized access"),
			expectedErr: map[string]interface{}{
				"code":    -32000,
				"message": "unauthorized access",
			},
		},
		{
			name: "empty error message",
			id:   "empty-error",
			err:  errors.New(""),
			expectedErr: map[string]interface{}{
				"code":    -32000,
				"message": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewErrorResponse(tt.id, tt.err)

			// Type assertion to map[string]interface{}
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatal("NewErrorResponse should return map[string]interface{}")
			}

			// Verify JSON-RPC 2.0 compliance
			if resultMap["jsonrpc"] != "2.0" {
				t.Errorf("Expected jsonrpc to be '2.0', got %v", resultMap["jsonrpc"])
			}
			if !reflect.DeepEqual(tt.id, resultMap["id"]) {
				t.Errorf("Expected id %v, got %v", tt.id, resultMap["id"])
			}

			// Verify error structure
			errorObj, ok := resultMap["error"].(map[string]interface{})
			if !ok {
				t.Fatal("error field should be a map")
			}
			if !reflect.DeepEqual(tt.expectedErr, errorObj) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, errorObj)
			}
		})
	}
}

// TestNewResponseJSONSerialization tests that NewResponse output can be properly JSON serialized
func TestNewResponseJSONSerialization(t *testing.T) {
	response := NewResponse("test-123", map[string]interface{}{
		"status": "success",
		"data":   []int{1, 2, 3},
	})

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("NewResponse output should be JSON serializable: %v", err)
	}

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("JSON should unmarshal correctly: %v", err)
	}

	if unmarshaled["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", unmarshaled["jsonrpc"])
	}
	if unmarshaled["id"] != "test-123" {
		t.Errorf("Expected id 'test-123', got %v", unmarshaled["id"])
	}
	if unmarshaled["result"] == nil {
		t.Error("Expected result to be present")
	}
}

// TestNewErrorResponseJSONSerialization tests that NewErrorResponse output can be properly JSON serialized
func TestNewErrorResponseJSONSerialization(t *testing.T) {
	response := NewErrorResponse("error-456", errors.New("validation failed"))

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("NewErrorResponse output should be JSON serializable: %v", err)
	}

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("JSON should unmarshal correctly: %v", err)
	}

	if unmarshaled["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", unmarshaled["jsonrpc"])
	}
	if unmarshaled["id"] != "error-456" {
		t.Errorf("Expected id 'error-456', got %v", unmarshaled["id"])
	}

	errorObj, ok := unmarshaled["error"].(map[string]interface{})
	if !ok {
		t.Fatal("error field should be present")
	}
	if errorObj["code"] != float64(-32000) { // JSON numbers become float64
		t.Errorf("Expected error code -32000, got %v", errorObj["code"])
	}
	if errorObj["message"] != "validation failed" {
		t.Errorf("Expected error message 'validation failed', got %v", errorObj["message"])
	}
}

// TestRPCRequestStructure tests the RPCRequest struct and its JSON tags
func TestRPCRequestStructure(t *testing.T) {
	tests := []struct {
		name        string
		jsonInput   string
		expected    RPCRequest
		shouldError bool
	}{
		{
			name:      "valid complete request",
			jsonInput: `{"jsonrpc":"2.0","method":"test.method","params":{"key":"value"},"id":"123"}`,
			expected: RPCRequest{
				JSONRPC: "2.0",
				Method:  "test.method",
				Params:  map[string]interface{}{"key": "value"},
				ID:      "123",
			},
			shouldError: false,
		},
		{
			name:      "request without params",
			jsonInput: `{"jsonrpc":"2.0","method":"simple.method","id":456}`,
			expected: RPCRequest{
				JSONRPC: "2.0",
				Method:  "simple.method",
				Params:  nil,
				ID:      float64(456), // JSON numbers become float64
			},
			shouldError: false,
		},
		{
			name:      "request with null id",
			jsonInput: `{"jsonrpc":"2.0","method":"notification","id":null}`,
			expected: RPCRequest{
				JSONRPC: "2.0",
				Method:  "notification",
				Params:  nil,
				ID:      nil,
			},
			shouldError: false,
		},
		{
			name:        "invalid json",
			jsonInput:   `{"jsonrpc":"2.0","method":}`,
			expected:    RPCRequest{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req RPCRequest
			err := json.Unmarshal([]byte(tt.jsonInput), &req)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if req.JSONRPC != tt.expected.JSONRPC {
				t.Errorf("Expected JSONRPC %q, got %q", tt.expected.JSONRPC, req.JSONRPC)
			}
			if req.Method != tt.expected.Method {
				t.Errorf("Expected Method %q, got %q", tt.expected.Method, req.Method)
			}
			if !reflect.DeepEqual(req.ID, tt.expected.ID) {
				t.Errorf("Expected ID %v, got %v", tt.expected.ID, req.ID)
			}

			if tt.expected.Params != nil {
				if !reflect.DeepEqual(req.Params, tt.expected.Params) {
					t.Errorf("Expected Params %v, got %v", tt.expected.Params, req.Params)
				}
			}
		})
	}
}

// TestWSConnectionStructure tests the wsConnection struct initialization
func TestWSConnectionStructure(t *testing.T) {
	// This test mainly verifies the struct can be created and has expected fields
	conn := &wsConnection{}

	// Test that the struct has the expected fields by checking they can be accessed
	// We can't test the mutex directly, but we can verify the struct is properly defined
	if conn == nil {
		t.Error("wsConnection should be creatable")
	}

	// Test that we can assign values to the struct fields
	conn.conn = nil // This should work if the field exists
	// The mutex field exists if we can take its address without compiler error
}

// TestUpgraderConfiguration tests the upgrader method configuration
func TestUpgraderConfiguration(t *testing.T) {
	server := &RPCServer{}
	upgrader := server.upgrader()

	if upgrader.ReadBufferSize != 1024 {
		t.Errorf("Expected ReadBufferSize 1024, got %d", upgrader.ReadBufferSize)
	}
	if upgrader.WriteBufferSize != 1024 {
		t.Errorf("Expected WriteBufferSize 1024, got %d", upgrader.WriteBufferSize)
	}
	if upgrader.CheckOrigin == nil {
		t.Error("CheckOrigin function should be set")
	}

	// Test CheckOrigin function allows localhost origins (default allowed origins)
	req := &http.Request{
		Header: http.Header{
			"Origin": []string{"http://localhost:8080"},
		},
	}
	if !upgrader.CheckOrigin(req) {
		t.Error("CheckOrigin should allow localhost origins")
	}

	// Test with disallowed origin
	req.Header.Set("Origin", "https://malicious.site")
	if upgrader.CheckOrigin(req) {
		t.Error("CheckOrigin should reject unauthorized origins")
	}
}

// TestRPCRequestMarshalUnmarshal tests round-trip JSON serialization
func TestRPCRequestMarshalUnmarshal(t *testing.T) {
	original := RPCRequest{
		JSONRPC: "2.0",
		Method:  "test.method",
		Params:  map[string]interface{}{"key": "value", "number": 42},
		ID:      "test-id",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal RPCRequest: %v", err)
	}

	// Unmarshal back
	var unmarshaled RPCRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal RPCRequest: %v", err)
	}

	// Compare fields
	if original.JSONRPC != unmarshaled.JSONRPC {
		t.Errorf("JSONRPC mismatch: expected %q, got %q", original.JSONRPC, unmarshaled.JSONRPC)
	}
	if original.Method != unmarshaled.Method {
		t.Errorf("Method mismatch: expected %q, got %q", original.Method, unmarshaled.Method)
	}
	if original.ID != unmarshaled.ID {
		t.Errorf("ID mismatch: expected %v, got %v", original.ID, unmarshaled.ID)
	}
}

// TestNewResponseEdgeCases tests edge cases for NewResponse
func TestNewResponseEdgeCases(t *testing.T) {
	// Test with zero values
	response := NewResponse(0, "")
	resultMap := response.(map[string]interface{})

	if resultMap["jsonrpc"] != "2.0" {
		t.Error("jsonrpc should always be '2.0'")
	}
	if resultMap["id"] != 0 {
		t.Errorf("Expected id 0, got %v", resultMap["id"])
	}
	if resultMap["result"] != "" {
		t.Errorf("Expected empty string result, got %v", resultMap["result"])
	}

	// Test with large data structures
	largeData := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largeData[string(rune('a'+i%26))] = i
	}

	response = NewResponse("large-test", largeData)
	resultMap = response.(map[string]interface{})

	if !reflect.DeepEqual(resultMap["result"], largeData) {
		t.Error("Large data structure should be preserved")
	}
}

// TestNewErrorResponseEdgeCases tests edge cases for NewErrorResponse
func TestNewErrorResponseEdgeCases(t *testing.T) {
	// Test with custom error implementation
	customErr := customError{msg: "custom error message"}
	response := NewErrorResponse("custom", customErr)
	resultMap := response.(map[string]interface{})

	errorObj := resultMap["error"].(map[string]interface{})
	if errorObj["message"] != "custom error message" {
		t.Errorf("Expected custom error message, got %v", errorObj["message"])
	}
	if errorObj["code"] != -32000 {
		t.Errorf("Expected error code -32000, got %v", errorObj["code"])
	}
}

// Define custom error type at package level
type customError struct {
	msg string
}

func (e customError) Error() string {
	return e.msg
}

// BenchmarkNewResponse benchmarks the NewResponse function
func BenchmarkNewResponse(b *testing.B) {
	id := "test-id"
	result := map[string]interface{}{
		"status": "success",
		"data":   []int{1, 2, 3, 4, 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewResponse(id, result)
	}
}

// BenchmarkNewErrorResponse benchmarks the NewErrorResponse function
func BenchmarkNewErrorResponse(b *testing.B) {
	id := "error-id"
	err := errors.New("benchmark error message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewErrorResponse(id, err)
	}
}

// TestGetAllowedOrigins tests the getAllowedOrigins method
func TestGetAllowedOrigins(t *testing.T) {
	server := &RPCServer{}

	// Test default origins (when env var is not set)
	origins := server.getAllowedOrigins()
	expectedDefaults := []string{
		"http://localhost:8080",
		"https://localhost:8080",
		"http://127.0.0.1:8080",
		"https://127.0.0.1:8080",
	}

	if len(origins) != len(expectedDefaults) {
		t.Errorf("Expected %d default origins, got %d", len(expectedDefaults), len(origins))
	}

	for i, expected := range expectedDefaults {
		if i >= len(origins) || origins[i] != expected {
			t.Errorf("Expected default origin %s at index %d, got %s", expected, i, origins[i])
		}
	}
}

// TestIsOriginAllowed tests the isOriginAllowed function
func TestIsOriginAllowed(t *testing.T) {
	server := &RPCServer{}
	allowedOrigins := []string{
		"https://example.com",
		"http://localhost:8080",
		"https://app.example.com",
	}

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"allowed origin", "https://example.com", true},
		{"localhost allowed", "http://localhost:8080", true},
		{"subdomain allowed", "https://app.example.com", true},
		{"disallowed origin", "https://malicious.site", false},
		{"case sensitive", "HTTPS://EXAMPLE.COM", false},
		{"empty origin", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := server.isOriginAllowed(test.origin, allowedOrigins)
			if result != test.expected {
				t.Errorf("isOriginAllowed(%s) = %v, expected %v", test.origin, result, test.expected)
			}
		})
	}
}

// TestGetSessionSafely_TOCTOU tests that getSessionSafely prevents time-of-check-time-of-use issues
func TestGetSessionSafely_TOCTOU(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	// Create a test session
	sessionID := "test-session-toctou"
	session := &PlayerSession{
		SessionID:   sessionID,
		LastActive:  time.Now().Add(-time.Minute),
		MessageChan: make(chan []byte, 500),
		WSConn:      &websocket.Conn{}, // Mock WebSocket connection
	}
	server.sessions[sessionID] = session

	// Test concurrent access - one goroutine retrieves session, another deletes it
	var wg sync.WaitGroup
	var retrievedSession *PlayerSession
	var retrievalError error

	// Goroutine 1: Retrieve session safely
	wg.Add(1)
	go func() {
		defer wg.Done()
		retrievedSession, retrievalError = server.getSessionSafely(sessionID)
	}()

	// Goroutine 2: Delete session (simulating cleanup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.mu.Lock()
		delete(server.sessions, sessionID)
		server.mu.Unlock()
	}()

	wg.Wait()

	// Either the session was retrieved successfully (before deletion) or not found (after deletion)
	// Both outcomes are acceptable - the important thing is no panic or inconsistent state
	if retrievalError == nil {
		if retrievedSession == nil {
			t.Error("Expected non-nil session when no error returned")
		}
		if retrievedSession.SessionID != sessionID {
			t.Errorf("Expected session ID %s, got %s", sessionID, retrievedSession.SessionID)
		}
	} else {
		if retrievedSession != nil {
			t.Error("Expected nil session when error returned")
		}
	}
}

// TestGetSessionSafely_ValidSession tests successful session retrieval
func TestGetSessionSafely_ValidSession(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	sessionID := "valid-session"
	originalTime := time.Now().Add(-time.Hour)
	session := &PlayerSession{
		SessionID:   sessionID,
		LastActive:  originalTime,
		MessageChan: make(chan []byte, 500),
		WSConn:      &websocket.Conn{}, // Mock WebSocket connection
	}
	server.sessions[sessionID] = session

	retrievedSession, err := server.getSessionSafely(sessionID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrievedSession == nil {
		t.Fatal("Expected non-nil session")
	}
	if retrievedSession.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, retrievedSession.SessionID)
	}
	// LastActive should be updated
	if !retrievedSession.LastActive.After(originalTime) {
		t.Error("Expected LastActive to be updated")
	}
}

// TestGetSessionSafely_InvalidSession tests handling of non-existent sessions
func TestGetSessionSafely_InvalidSession(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	retrievedSession, err := server.getSessionSafely("non-existent")

	if err != ErrInvalidSession {
		t.Errorf("Expected ErrInvalidSession, got %v", err)
	}
	if retrievedSession != nil {
		t.Error("Expected nil session for non-existent session ID")
	}
}

// TestGetSessionSafely_EmptySessionID tests handling of empty session ID
func TestGetSessionSafely_EmptySessionID(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	retrievedSession, err := server.getSessionSafely("")

	if err != ErrInvalidSession {
		t.Errorf("Expected ErrInvalidSession, got %v", err)
	}
	if retrievedSession != nil {
		t.Error("Expected nil session for empty session ID")
	}
}

// TestGetSessionSafely_NoWebSocketConnection tests handling of sessions without WebSocket connection
func TestGetSessionSafely_NoWebSocketConnection(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	sessionID := "no-websocket"
	session := &PlayerSession{
		SessionID:   sessionID,
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, 500),
		WSConn:      nil, // No WebSocket connection
	}
	server.sessions[sessionID] = session

	retrievedSession, err := server.getSessionSafely(sessionID)

	if err != ErrInvalidSession {
		t.Errorf("Expected ErrInvalidSession, got %v", err)
	}
	if retrievedSession != nil {
		t.Error("Expected nil session for session without WebSocket connection")
	}
}
