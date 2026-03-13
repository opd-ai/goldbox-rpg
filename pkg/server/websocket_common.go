package server

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// getSessionFromContext extracts and validates a PlayerSession from the HTTP request context.
// Returns nil and writes an HTTP error response if the session is not found or invalid.
//
// Parameters:
//   - w: HTTP response writer for error responses
//   - r: HTTP request containing the session context
//   - loggerName: Name of the calling function for logging
//
// Returns:
//   - *PlayerSession: The validated session, or nil if validation failed
//   - bool: true if session was successfully extracted, false otherwise
func (s *RPCServer) getSessionFromContext(w http.ResponseWriter, r *http.Request, loggerName string) (*PlayerSession, bool) {
	logger := logrus.WithField("function", loggerName)

	sessionVal := r.Context().Value(sessionKey)
	if sessionVal == nil {
		logger.Error("no session value in context")
		http.Error(w, "Session required", http.StatusUnauthorized)
		return nil, false
	}

	session, ok := sessionVal.(*PlayerSession)
	if !ok || session == nil {
		logger.Error("invalid session type in context")
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return nil, false
	}

	return session, true
}
