package persistence

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// MemorySessionStore implements SessionStore using in-memory storage.
// This is useful for development and testing where persistence is not required.
type MemorySessionStore struct {
	sessions map[string]*SessionData
	mu       sync.RWMutex
}

// NewMemorySessionStore creates a new in-memory session store.
func NewMemorySessionStore() *MemorySessionStore {
	logrus.WithField("function", "NewMemorySessionStore").Info("creating in-memory session store")

	return &MemorySessionStore{
		sessions: make(map[string]*SessionData),
	}
}

// SaveSession saves a session to memory.
func (mss *MemorySessionStore) SaveSession(session *SessionData) error {
	mss.mu.Lock()
	defer mss.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function":   "SaveSession",
		"sessionID":  session.SessionID,
		"playerName": session.PlayerName,
	}).Debug("saving session to memory")

	sessionCopy := *session
	mss.sessions[session.SessionID] = &sessionCopy

	return nil
}

// LoadSession retrieves a session from memory.
func (mss *MemorySessionStore) LoadSession(sessionID string) (*SessionData, error) {
	mss.mu.RLock()
	defer mss.mu.RUnlock()

	logrus.WithFields(logrus.Fields{
		"function":  "LoadSession",
		"sessionID": sessionID,
	}).Debug("loading session from memory")

	session, exists := mss.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	sessionCopy := *session
	return &sessionCopy, nil
}

// DeleteSession removes a session from memory.
func (mss *MemorySessionStore) DeleteSession(sessionID string) error {
	mss.mu.Lock()
	defer mss.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function":  "DeleteSession",
		"sessionID": sessionID,
	}).Debug("deleting session from memory")

	delete(mss.sessions, sessionID)
	return nil
}

// ListSessions returns all stored session IDs.
func (mss *MemorySessionStore) ListSessions() ([]string, error) {
	mss.mu.RLock()
	defer mss.mu.RUnlock()

	sessionIDs := make([]string, 0, len(mss.sessions))
	for sessionID := range mss.sessions {
		sessionIDs = append(sessionIDs, sessionID)
	}

	logrus.WithFields(logrus.Fields{
		"function": "ListSessions",
		"count":    len(sessionIDs),
	}).Debug("listed sessions from memory")

	return sessionIDs, nil
}

// SaveAllSessions saves all sessions to memory.
func (mss *MemorySessionStore) SaveAllSessions(sessions map[string]*SessionData) error {
	mss.mu.Lock()
	defer mss.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function": "SaveAllSessions",
		"count":    len(sessions),
	}).Info("saving all sessions to memory")

	for sessionID, session := range sessions {
		sessionCopy := *session
		mss.sessions[sessionID] = &sessionCopy
	}

	return nil
}

// LoadAllSessions retrieves all sessions from memory.
func (mss *MemorySessionStore) LoadAllSessions() (map[string]*SessionData, error) {
	mss.mu.RLock()
	defer mss.mu.RUnlock()

	sessions := make(map[string]*SessionData, len(mss.sessions))
	for sessionID, session := range mss.sessions {
		sessionCopy := *session
		sessions[sessionID] = &sessionCopy
	}

	logrus.WithFields(logrus.Fields{
		"function": "LoadAllSessions",
		"count":    len(sessions),
	}).Info("loaded all sessions from memory")

	return sessions, nil
}
