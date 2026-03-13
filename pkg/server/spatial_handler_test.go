package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHandleFindPathInvalidJSON tests findPath with invalid JSON
func TestHandleFindPathInvalidJSON(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	params := json.RawMessage(`{invalid}`)
	result, err := server.handleFindPath(params)

	assert.NoError(t, err) // Returns error in response, not as error
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "invalid")
}

// TestHandleFindPathInvalidSession tests findPath with invalid session
func TestHandleFindPathInvalidSession(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	params := json.RawMessage(`{"session_id": "nonexistent", "start_x": 0, "start_y": 0, "end_x": 5, "end_y": 5}`)
	result, err := server.handleFindPath(params)

	assert.NoError(t, err)
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "invalid session")
}

// TestHandleFindPathValid tests valid findPath request
func TestHandleFindPathValid(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)

	params := json.RawMessage(`{"session_id": "` + session.SessionID + `", "start_x": 0, "start_y": 0, "end_x": 5, "end_y": 5}`)
	result, err := server.handleFindPath(params)

	assert.NoError(t, err)
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	// Should have success field (true or false depending on pathfinding)
	_, hasSuccess := response["success"]
	assert.True(t, hasSuccess)
}

// TestHandleGetNearestObjectsInvalidJSON tests getNearestObjects with invalid JSON
func TestHandleGetNearestObjectsInvalidJSON(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	params := json.RawMessage(`{invalid}`)
	result, err := server.handleGetNearestObjects(params)

	assert.NoError(t, err)
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, false, response["success"])
}

// TestHandleGetNearestObjectsInvalidSession tests getNearestObjects with invalid session
func TestHandleGetNearestObjectsInvalidSession(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	params := json.RawMessage(`{"session_id": "nonexistent"}`)
	result, err := server.handleGetNearestObjects(params)

	assert.NoError(t, err)
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, false, response["success"])
}

// TestHandleGetNearestObjectsValid tests valid getNearestObjects
func TestHandleGetNearestObjectsValid(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)

	params := json.RawMessage(`{"session_id": "` + session.SessionID + `", "count": 5}`)
	result, err := server.handleGetNearestObjects(params)

	assert.NoError(t, err)
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	_, hasSuccess := response["success"]
	assert.True(t, hasSuccess)
}

// TestHandleGetObjectsInRangeValid tests getObjectsInRange
func TestHandleGetObjectsInRangeValid(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)

	params := json.RawMessage(`{"session_id": "` + session.SessionID + `", "range": 10}`)
	result, err := server.handleGetObjectsInRange(params)

	assert.NoError(t, err)
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	_, hasSuccess := response["success"]
	assert.True(t, hasSuccess)
}

// TestHandleGetObjectsInRadiusValid tests getObjectsInRadius
func TestHandleGetObjectsInRadiusValid(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)

	params := json.RawMessage(`{"session_id": "` + session.SessionID + `", "center_x": 5, "center_y": 5, "radius": 10}`)
	result, err := server.handleGetObjectsInRadius(params)

	assert.NoError(t, err)
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	_, hasSuccess := response["success"]
	assert.True(t, hasSuccess)
}
