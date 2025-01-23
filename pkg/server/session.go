package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *RPCServer) getOrCreateSession(w http.ResponseWriter, r *http.Request) (*PlayerSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cookie, err := r.Cookie("session_id")
	if err == nil {
		if session, exists := s.sessions[cookie.Value]; exists {
			session.LastActive = time.Now()
			return session, nil
		}
	}

	sessionID := uuid.New().String()
	session := &PlayerSession{
		SessionID:   sessionID,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, 100),
	}
	s.sessions[sessionID] = session

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})

	return session, nil
}

func (s *RPCServer) startSessionCleanup() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			s.mu.Lock()
			for id, session := range s.sessions {
				if time.Since(session.LastActive) > 30*time.Minute {
					if session.WSConn != nil {
						session.WSConn.Close()
					}
					close(session.MessageChan)
					delete(s.sessions, id)
				}
			}
			s.mu.Unlock()
		}
	}()
}
