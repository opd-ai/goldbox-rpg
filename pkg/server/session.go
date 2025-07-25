package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Session configuration constants are defined in constants.go

// safeSendMessage attempts to send a message to a session's MessageChan without blocking.
// If the channel is full, it logs a warning and drops the message to prevent resource exhaustion.
//
// Parameters:
//   - session: The player session to send the message to
//   - message: The message bytes to send
//
// Returns:
//   - bool: true if message was sent successfully, false if dropped due to full channel
func safeSendMessage(session *PlayerSession, message []byte) bool {
	if session == nil || session.MessageChan == nil {
		return false
	}

	select {
	case session.MessageChan <- message:
		return true
	case <-time.After(MessageSendTimeout):
		logrus.WithFields(logrus.Fields{
			"sessionID": session.SessionID,
			"function":  "safeSendMessage",
		}).Warn("Message dropped: channel full or timeout reached")
		return false
	}
}

// getOrCreateSession handles session management for HTTP requests by either retrieving an existing
// session or creating a new one. It maintains user sessions through cookies and ensures thread-safe
// access to the sessions map.
//
// Parameters:
//   - w http.ResponseWriter: The response writer to set session cookies
//   - r *http.Request: The incoming HTTP request containing potential session cookies
//
// Returns:
//   - *PlayerSession: A pointer to either the existing or newly created session
//   - error: Error if session handling fails
//
// The function performs the following:
// 1. Checks for existing session cookie
// 2. If found and valid, returns the existing session
// 3. If not found or invalid, creates new session with UUID
// 4. Sets session cookie in response
// 5. Updates LastActive timestamp
//
// Thread-safety is ensured via mutex locking of the sessions map.
// Sessions expire after 30 minutes as set in cookie MaxAge and sessionTimeout constant.
//
// Related types:
//   - PlayerSession struct
//   - RPCServer struct
func (s *RPCServer) getOrCreateSession(w http.ResponseWriter, r *http.Request) (*PlayerSession, error) {
	logrus.WithFields(logrus.Fields{
		"func": "getOrCreateSession",
		"path": r.URL.Path,
	}).Debug("Starting session handling")

	s.mu.Lock()
	defer s.mu.Unlock()

	cookie, err := r.Cookie("session_id")
	if err == nil {
		if session, exists := s.sessions[cookie.Value]; exists {
			session.LastActive = time.Now()
			session.addRef() // Increment reference count to prevent cleanup
			logrus.WithFields(logrus.Fields{
				"func":      "getOrCreateSession",
				"sessionID": cookie.Value,
			}).Debug("Existing session found and updated")
			return session, nil
		}
		logrus.WithFields(logrus.Fields{
			"func":      "getOrCreateSession",
			"sessionID": cookie.Value,
		}).Warn("Cookie exists but session not found")
	}

	sessionID := uuid.New().String()
	session := &PlayerSession{
		SessionID:   sessionID,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}
	session.addRef() // Increment reference count for new session
	s.sessions[sessionID] = session

	// Update metrics for new session
	if s.metrics != nil {
		s.metrics.UpdateActiveSessions(len(s.sessions))
	}

	// Determine if connection is secure for proper cookie settings
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(sessionTimeout.Seconds()), // Use consistent session timeout
		SameSite: http.SameSiteStrictMode,
		Secure:   isSecure,
	})

	logrus.WithFields(logrus.Fields{
		"func":      "getOrCreateSession",
		"sessionID": sessionID,
	}).Info("New session created")

	return session, nil
}

// startSessionCleanup starts a background goroutine that periodically cleans up expired sessions.
// It runs every 5 minutes and removes sessions that have been inactive for more than 30 minutes.
//
// The cleanup process:
// 1. Iterates through all sessions under a mutex lock
// 2. Checks each session's LastActive timestamp
// 3. For expired sessions:
//   - Closes the websocket connection if present
//   - Closes the message channel
//   - Removes the session from the sessions map
//
// The function logs:
// - Debug messages when starting and during each cleanup cycle
// - Info messages for removed sessions and cleanup completion
// - Error messages if websocket connections fail to close
//
// Related types:
// - RPCServer - The server instance this runs on
// - Session - The session objects being cleaned up
//
// Note: This is a non-blocking function as it launches the cleanup routine in a separate goroutine
/*func (s *RPCServer) startSessionCleanup() {
	logrus.WithFields(logrus.Fields{
		"func": "startSessionCleanup",
	}).Debug("Starting session cleanup routine")

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			logrus.WithFields(logrus.Fields{
				"func": "startSessionCleanup",
			}).Debug("Running cleanup cycle")

			s.mu.Lock()
			expiredCount := 0
			for id, session := range s.sessions {
				if time.Since(session.LastActive) > 30*time.Minute {
					logrus.WithFields(logrus.Fields{
						"func":      "startSessionCleanup",
						"sessionID": id,
						"inactive":  time.Since(session.LastActive),
					}).Info("Removing expired session")

					if session.WSConn != nil {
						if err := session.WSConn.Close(); err != nil {
							logrus.WithFields(logrus.Fields{
								"func":      "startSessionCleanup",
								"sessionID": id,
								"error":     err,
							}).Error("Failed to close websocket connection")
						}
					}
					close(session.MessageChan)
					delete(s.sessions, id)
					expiredCount++
				}
			}
			s.mu.Unlock()

			logrus.WithFields(logrus.Fields{
				"func":         "startSessionCleanup",
				"expiredCount": expiredCount,
				"totalActive":  len(s.sessions),
			}).Info("Cleanup cycle completed")
		}
	}()
}
*/
func (s *RPCServer) startSessionCleanup() {
	ticker := time.NewTicker(sessionCleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.cleanupExpiredSessions()
			case <-s.done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *RPCServer) cleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.Sub(session.LastActive) > sessionTimeout {
			// Check if session is currently in use by a handler
			if session.isInUse() {
				// Skip cleanup for now, will be retried in next cycle
				continue
			}

			if session.WSConn != nil {
				session.WSConn.Close()
			}
			delete(s.sessions, id)

			// Update metrics for session removal
			if s.metrics != nil {
				s.metrics.UpdateActiveSessions(len(s.sessions))
			}
		}
	}
}

// getSession safely retrieves a session by ID with proper mutex protection and reference counting
func (s *RPCServer) getSession(sessionID string) (*PlayerSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if exists {
		session.addRef() // Increment reference count to prevent cleanup
	}
	return session, exists
}

// setSession safely sets a session with proper mutex protection
func (s *RPCServer) setSession(sessionID string, session *PlayerSession) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = session
}

// releaseSession decrements the reference count for a session after handler use
func (s *RPCServer) releaseSession(session *PlayerSession) {
	if session != nil {
		session.release()
	}
}
