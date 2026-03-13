// Package server provides common RPC handler patterns to reduce duplication.
package server

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// sessionRequest is an interface for request types that contain a session ID.
type sessionRequest interface {
	GetSessionID() string
}

// simpleSessionRequest is a minimal request containing only a session ID.
type simpleSessionRequest struct {
	SessionID string `json:"session_id"`
}

// GetSessionID returns the session ID from the request.
func (r simpleSessionRequest) GetSessionID() string {
	return r.SessionID
}

// sessionOp is a function that performs an operation with a validated session.
// It receives the character ID of the session's player and performs the operation.
type sessionOp func(characterID string) (interface{}, error)

// executeWithSession handles the common pattern of validating a session,
// getting the character ID, and executing an operation. It reduces duplication
// across guild, faction, and other session-based handlers.
func (s *RPCServer) executeWithSession(params json.RawMessage, opName string, fn sessionOp) (interface{}, error) {
	logrus.WithField("function", opName).Debug("entering " + opName)

	var req simpleSessionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	characterID := session.Player.GetID()
	return fn(characterID)
}

// guildSuccessMsg creates a standardized success response with a custom message.
func guildSuccessMsg(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"message": message,
	}
}
