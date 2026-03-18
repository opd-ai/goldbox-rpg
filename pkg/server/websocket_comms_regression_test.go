package server

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewErrorResponse_PreservesJSONRPCErrorCode verifies that NewErrorResponse
// preserves the code, message, and data from a *JSONRPCError instead of
// hardcoding -32000.  This is a regression test for a bug where the WebSocket
// path always returned error code -32000 regardless of the specific error type.
func TestNewErrorResponse_PreservesJSONRPCErrorCode(t *testing.T) {
	tests := []struct {
		name         string
		id           interface{}
		err          error
		expectedCode int
		expectedMsg  string
		expectedData interface{}
	}{
		{
			name:         "plain error uses default code",
			id:           1,
			err:          errors.New("something went wrong"),
			expectedCode: -32000,
			expectedMsg:  "something went wrong",
			expectedData: nil,
		},
		{
			name:         "JSONRPCError method not found",
			id:           2,
			err:          NewJSONRPCError(JSONRPCMethodNotFound, "Method not found: badMethod", nil),
			expectedCode: JSONRPCMethodNotFound,
			expectedMsg:  "Method not found: badMethod",
			expectedData: nil,
		},
		{
			name:         "JSONRPCError invalid params with data",
			id:           3,
			err:          NewJSONRPCError(JSONRPCInvalidParams, "Invalid method parameters", "missing field: session_id"),
			expectedCode: JSONRPCInvalidParams,
			expectedMsg:  "Invalid method parameters",
			expectedData: "missing field: session_id",
		},
		{
			name:         "JSONRPCError parse error",
			id:           "req-4",
			err:          NewJSONRPCError(JSONRPCParseError, "Parse error", nil),
			expectedCode: JSONRPCParseError,
			expectedMsg:  "Parse error",
			expectedData: nil,
		},
		{
			name:         "JSONRPCError internal error",
			id:           nil,
			err:          NewJSONRPCError(JSONRPCInternalError, "Internal JSON-RPC error", nil),
			expectedCode: JSONRPCInternalError,
			expectedMsg:  "Internal JSON-RPC error",
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewErrorResponse(tt.id, tt.err)
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatal("NewErrorResponse should return map[string]interface{}")
			}

			if resultMap["jsonrpc"] != "2.0" {
				t.Errorf("expected jsonrpc '2.0', got %v", resultMap["jsonrpc"])
			}
			if !reflect.DeepEqual(tt.id, resultMap["id"]) {
				t.Errorf("expected id %v, got %v", tt.id, resultMap["id"])
			}

			errorObj, ok := resultMap["error"].(map[string]interface{})
			if !ok {
				t.Fatal("error field should be a map")
			}

			if errorObj["code"] != tt.expectedCode {
				t.Errorf("expected error code %d, got %v", tt.expectedCode, errorObj["code"])
			}
			if errorObj["message"] != tt.expectedMsg {
				t.Errorf("expected error message %q, got %v", tt.expectedMsg, errorObj["message"])
			}
			if tt.expectedData != nil {
				if errorObj["data"] != tt.expectedData {
					t.Errorf("expected error data %v, got %v", tt.expectedData, errorObj["data"])
				}
			} else {
				if _, hasData := errorObj["data"]; hasData {
					t.Errorf("expected no data field, got %v", errorObj["data"])
				}
			}
		})
	}
}

// TestNewErrorResponse_JSONRoundTrip verifies that a *JSONRPCError survives
// JSON marshal→unmarshal with the correct code preserved (not -32000).
func TestNewErrorResponse_JSONRoundTrip(t *testing.T) {
	original := NewJSONRPCError(JSONRPCMethodNotFound, "Method not found: move", nil)
	response := NewErrorResponse(42, original)

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	errorObj, ok := unmarshaled["error"].(map[string]interface{})
	if !ok {
		t.Fatal("error field should be present")
	}

	// JSON numbers become float64 after unmarshal
	if errorObj["code"] != float64(JSONRPCMethodNotFound) {
		t.Errorf("expected code %d after roundtrip, got %v", JSONRPCMethodNotFound, errorObj["code"])
	}
	if errorObj["message"] != "Method not found: move" {
		t.Errorf("expected message 'Method not found: move', got %v", errorObj["message"])
	}
}

// TestNewErrorResponse_JSONRPCErrorWithData checks that the optional data field
// is included when the JSONRPCError carries it.
func TestNewErrorResponse_JSONRPCErrorWithData(t *testing.T) {
	errWithData := NewJSONRPCError(JSONRPCInvalidParams, "bad params", map[string]string{"field": "session_id"})
	response := NewErrorResponse(99, errWithData)

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	errorObj := unmarshaled["error"].(map[string]interface{})

	if errorObj["code"] != float64(JSONRPCInvalidParams) {
		t.Errorf("expected code %d, got %v", JSONRPCInvalidParams, errorObj["code"])
	}

	dataObj, ok := errorObj["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", errorObj["data"])
	}
	if dataObj["field"] != "session_id" {
		t.Errorf("expected data.field 'session_id', got %v", dataObj["field"])
	}
}

// TestHandleJoinGame_ReturnsPlayerID verifies that handleJoinGame includes
// the player_id field in the response so the WASM client can populate
// JoinGameResult.PlayerID. This is a regression test for a bug where
// player_id was missing from the joinGame response.
func TestHandleJoinGame_ReturnsPlayerID(t *testing.T) {
	server := createTestServerForHandlers(t)

	// ---- WebSocket path: attach to existing session ----
	existingSessionID := "ws-session-pid-001"
	server.mu.Lock()
	server.sessions[existingSessionID] = &PlayerSession{
		SessionID:   existingSessionID,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}
	server.mu.Unlock()

	joinParams, err := json.Marshal(map[string]interface{}{
		"player_name": "PlayerIdTestWS",
		"session_id":  existingSessionID,
	})
	require.NoError(t, err)

	result, err := server.handleJoinGame(joinParams)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, true, resultMap["success"])
	assert.Equal(t, existingSessionID, resultMap["session_id"])

	// player_id must be present and non-empty
	playerID, ok := resultMap["player_id"].(string)
	assert.True(t, ok, "player_id should be a string")
	assert.NotEmpty(t, playerID, "player_id should not be empty")

	// ---- HTTP path: no existing session ----
	httpParams, err := json.Marshal(map[string]interface{}{
		"player_name": "PlayerIdTestHTTP",
	})
	require.NoError(t, err)

	result2, err := server.handleJoinGame(httpParams)
	require.NoError(t, err)

	resultMap2, ok := result2.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, true, resultMap2["success"])
	assert.NotEmpty(t, resultMap2["session_id"])

	playerID2, ok := resultMap2["player_id"].(string)
	assert.True(t, ok, "player_id should be a string")
	assert.NotEmpty(t, playerID2, "player_id should not be empty on HTTP path")
}
