package persistence

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SessionData represents the serializable session information.
type SessionData struct {
	SessionID  string    `yaml:"session_id"`
	PlayerID   string    `yaml:"player_id"`
	PlayerName string    `yaml:"player_name"`
	LastActive time.Time `yaml:"last_active"`
	CreatedAt  time.Time `yaml:"created_at"`
	Connected  bool      `yaml:"connected"`
}

// SessionStore defines the interface for session persistence.
type SessionStore interface {
	// SaveSession persists a session to storage
	SaveSession(session *SessionData) error

	// LoadSession retrieves a session from storage
	LoadSession(sessionID string) (*SessionData, error)

	// DeleteSession removes a session from storage
	DeleteSession(sessionID string) error

	// ListSessions returns all stored session IDs
	ListSessions() ([]string, error)

	// SaveAllSessions persists multiple sessions atomically
	SaveAllSessions(sessions map[string]*SessionData) error

	// LoadAllSessions retrieves all sessions from storage
	LoadAllSessions() (map[string]*SessionData, error)
}

// FileSessionStore implements SessionStore using the file-based persistence layer.
type FileSessionStore struct {
	fileStore *FileStore
	mu        sync.RWMutex
}

// NewFileSessionStore creates a new file-based session store.
func NewFileSessionStore(dataDir string) (*FileSessionStore, error) {
	logrus.WithFields(logrus.Fields{
		"function": "NewFileSessionStore",
		"dataDir":  dataDir,
	}).Info("creating file-based session store")

	fileStore, err := NewFileStore(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create file store: %w", err)
	}

	return &FileSessionStore{
		fileStore: fileStore,
	}, nil
}

// SaveSession persists a single session to disk.
func (fss *FileSessionStore) SaveSession(session *SessionData) error {
	fss.mu.Lock()
	defer fss.mu.Unlock()

	filename := fmt.Sprintf("sessions/%s.yaml", session.SessionID)

	logrus.WithFields(logrus.Fields{
		"function":   "SaveSession",
		"sessionID":  session.SessionID,
		"playerName": session.PlayerName,
	}).Debug("saving session to file")

	return fss.fileStore.Save(filename, session)
}

// LoadSession retrieves a session from disk.
func (fss *FileSessionStore) LoadSession(sessionID string) (*SessionData, error) {
	fss.mu.RLock()
	defer fss.mu.RUnlock()

	filename := fmt.Sprintf("sessions/%s.yaml", sessionID)

	logrus.WithFields(logrus.Fields{
		"function":  "LoadSession",
		"sessionID": sessionID,
	}).Debug("loading session from file")

	var session SessionData
	if err := fss.fileStore.Load(filename, &session); err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	return &session, nil
}

// DeleteSession removes a session from disk.
func (fss *FileSessionStore) DeleteSession(sessionID string) error {
	fss.mu.Lock()
	defer fss.mu.Unlock()

	filename := fmt.Sprintf("sessions/%s.yaml", sessionID)

	logrus.WithFields(logrus.Fields{
		"function":  "DeleteSession",
		"sessionID": sessionID,
	}).Debug("deleting session file")

	return fss.fileStore.Delete(filename)
}

// ListSessions returns all stored session IDs.
func (fss *FileSessionStore) ListSessions() ([]string, error) {
	fss.mu.RLock()
	defer fss.mu.RUnlock()

	files, err := fss.fileStore.List("sessions/*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessionIDs := make([]string, 0, len(files))
	for _, file := range files {
		sessionID := file[9 : len(file)-5]
		sessionIDs = append(sessionIDs, sessionID)
	}

	logrus.WithFields(logrus.Fields{
		"function": "ListSessions",
		"count":    len(sessionIDs),
	}).Debug("listed sessions")

	return sessionIDs, nil
}

// SaveAllSessions persists all sessions atomically.
func (fss *FileSessionStore) SaveAllSessions(sessions map[string]*SessionData) error {
	fss.mu.Lock()
	defer fss.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function": "SaveAllSessions",
		"count":    len(sessions),
	}).Info("saving all sessions")

	for sessionID, session := range sessions {
		filename := fmt.Sprintf("sessions/%s.yaml", sessionID)
		if err := fss.fileStore.Save(filename, session); err != nil {
			return fmt.Errorf("failed to save session %s: %w", sessionID, err)
		}
	}

	return nil
}

// LoadAllSessions retrieves all sessions from disk.
func (fss *FileSessionStore) LoadAllSessions() (map[string]*SessionData, error) {
	fss.mu.RLock()
	defer fss.mu.RUnlock()

	sessionIDs, err := fss.ListSessions()
	if err != nil {
		return nil, err
	}

	sessions := make(map[string]*SessionData, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		fss.mu.RUnlock()
		session, err := fss.LoadSession(sessionID)
		fss.mu.RLock()

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"function":  "LoadAllSessions",
				"sessionID": sessionID,
				"error":     err,
			}).Warn("failed to load session, skipping")
			continue
		}

		sessions[sessionID] = session
	}

	logrus.WithFields(logrus.Fields{
		"function": "LoadAllSessions",
		"count":    len(sessions),
	}).Info("loaded all sessions")

	return sessions, nil
}
