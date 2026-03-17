package server

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wsTestLogger() *logrus.Entry {
	return logrus.NewEntry(logrus.New())
}

// TestProcessWebSocketAdventureListRequest verifies that adventure.list works
// correctly through the WebSocket request processing path (processWebSocketRequest).
func TestProcessWebSocketAdventureListRequest(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	conn := newMockWebSocketConn()
	session := &PlayerSession{SessionID: "test-session-ws-adv"}

	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  "adventure.list",
		Params:  map[string]interface{}{},
		ID:      float64(1),
	}

	err := server.processWebSocketRequest(conn, session, req, wsTestLogger())
	require.NoError(t, err)

	messages := conn.GetMessages()
	require.Len(t, messages, 1)

	// Marshal and re-parse to get consistent types
	respBytes, err := json.Marshal(messages[0])
	require.NoError(t, err)

var respMap map[string]interface{}
require.NoError(t, json.Unmarshal(respBytes, &respMap))

assert.Equal(t, "2.0", respMap["jsonrpc"])
assert.Equal(t, float64(1), respMap["id"])

resultMap, ok := respMap["result"].(map[string]interface{})
require.True(t, ok)
assert.Equal(t, true, resultMap["success"])
}

// TestProcessWebSocketAdventureLoadRequest verifies that adventure.load works
// through the WebSocket request processing path.
func TestProcessWebSocketAdventureLoadRequest(t *testing.T) {
server := createTestServerForHandlers(t)
defer server.Close()

// Check if any adventures are available
advMgr := server.getAdventureManager()
if advMgr == nil || len(advMgr.List()) == 0 {
t.Skip("No adventures available for test")
}

slug := advMgr.List()[0].Slug

conn := newMockWebSocketConn()
session := &PlayerSession{SessionID: "test-session-ws-load"}

req := RPCRequest{
JSONRPC: "2.0",
Method:  "adventure.load",
Params:  map[string]interface{}{"slug": slug},
ID:      float64(2),
}

err := server.processWebSocketRequest(conn, session, req, wsTestLogger())
require.NoError(t, err)

messages := conn.GetMessages()
require.Len(t, messages, 1)

respBytes, err := json.Marshal(messages[0])
require.NoError(t, err)

var respMap map[string]interface{}
require.NoError(t, json.Unmarshal(respBytes, &respMap))

assert.Equal(t, "2.0", respMap["jsonrpc"])
assert.Equal(t, float64(2), respMap["id"])

resultMap, ok := respMap["result"].(map[string]interface{})
require.True(t, ok)
assert.Equal(t, true, resultMap["success"])
}

// TestProcessWebSocketSequentialRequests verifies that multiple requests processed
// sequentially over a WebSocket connection work correctly (no stall between requests).
func TestProcessWebSocketSequentialRequests(t *testing.T) {
server := createTestServerForHandlers(t)
defer server.Close()

conn := newMockWebSocketConn()
session := &PlayerSession{SessionID: "test-session-ws-seq"}

// Send adventure.list first
listReq := RPCRequest{
JSONRPC: "2.0",
Method:  "adventure.list",
Params:  map[string]interface{}{},
ID:      float64(1),
}
err := server.processWebSocketRequest(conn, session, listReq, wsTestLogger())
require.NoError(t, err)

// Then send getGameState (simulates concurrent request pattern)
stateReq := RPCRequest{
JSONRPC: "2.0",
Method:  "getGameState",
Params:  map[string]interface{}{},
ID:      float64(2),
}
err = server.processWebSocketRequest(conn, session, stateReq, wsTestLogger())
require.NoError(t, err)

// Then send another adventure.list
listReq2 := RPCRequest{
JSONRPC: "2.0",
Method:  "adventure.list",
Params:  map[string]interface{}{},
ID:      float64(3),
}
err = server.processWebSocketRequest(conn, session, listReq2, wsTestLogger())
require.NoError(t, err)

// All three requests should have produced responses
messages := conn.GetMessages()
assert.Len(t, messages, 3, "Expected 3 responses for 3 sequential requests")
}
