package persistence

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileSessionStore(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewFileSessionStore(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, store)
}

func TestFileSessionStore_SaveLoadSession(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileSessionStore(tmpDir)
	require.NoError(t, err)

	session := &SessionData{
		SessionID:  "test-session-1",
		PlayerID:   "player-123",
		PlayerName: "TestPlayer",
		LastActive: time.Now(),
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		Connected:  true,
	}

	err = store.SaveSession(session)
	require.NoError(t, err)

	filename := filepath.Join(tmpDir, "sessions", "test-session-1.yaml")
	_, err = os.Stat(filename)
	require.NoError(t, err, "session file should exist")

	loaded, err := store.LoadSession("test-session-1")
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, session.SessionID, loaded.SessionID)
	assert.Equal(t, session.PlayerID, loaded.PlayerID)
	assert.Equal(t, session.PlayerName, loaded.PlayerName)
	assert.Equal(t, session.Connected, loaded.Connected)
}

func TestFileSessionStore_LoadNonExistentSession(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileSessionStore(tmpDir)
	require.NoError(t, err)

	_, err = store.LoadSession("non-existent")
	assert.Error(t, err)
}

func TestFileSessionStore_DeleteSession(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileSessionStore(tmpDir)
	require.NoError(t, err)

	session := &SessionData{
		SessionID:  "test-session-2",
		PlayerID:   "player-456",
		PlayerName: "TestPlayer2",
		LastActive: time.Now(),
		CreatedAt:  time.Now(),
		Connected:  false,
	}

	err = store.SaveSession(session)
	require.NoError(t, err)

	err = store.DeleteSession("test-session-2")
	require.NoError(t, err)

	filename := filepath.Join(tmpDir, "sessions", "test-session-2.yaml")
	_, err = os.Stat(filename)
	assert.True(t, os.IsNotExist(err), "session file should not exist after deletion")
}

func TestFileSessionStore_ListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileSessionStore(tmpDir)
	require.NoError(t, err)

	sessions := []*SessionData{
		{
			SessionID:  "session-1",
			PlayerID:   "player-1",
			PlayerName: "Player1",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  true,
		},
		{
			SessionID:  "session-2",
			PlayerID:   "player-2",
			PlayerName: "Player2",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  false,
		},
		{
			SessionID:  "session-3",
			PlayerID:   "player-3",
			PlayerName: "Player3",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  true,
		},
	}

	for _, session := range sessions {
		err := store.SaveSession(session)
		require.NoError(t, err)
	}

	sessionIDs, err := store.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessionIDs, 3)

	expectedIDs := map[string]bool{
		"session-1": true,
		"session-2": true,
		"session-3": true,
	}

	for _, id := range sessionIDs {
		assert.True(t, expectedIDs[id], "unexpected session ID: %s", id)
	}
}

func TestFileSessionStore_SaveAllSessions(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileSessionStore(tmpDir)
	require.NoError(t, err)

	sessions := map[string]*SessionData{
		"session-1": {
			SessionID:  "session-1",
			PlayerID:   "player-1",
			PlayerName: "Player1",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  true,
		},
		"session-2": {
			SessionID:  "session-2",
			PlayerID:   "player-2",
			PlayerName: "Player2",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  false,
		},
	}

	err = store.SaveAllSessions(sessions)
	require.NoError(t, err)

	sessionIDs, err := store.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessionIDs, 2)
}

func TestFileSessionStore_LoadAllSessions(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileSessionStore(tmpDir)
	require.NoError(t, err)

	originalSessions := map[string]*SessionData{
		"session-1": {
			SessionID:  "session-1",
			PlayerID:   "player-1",
			PlayerName: "Player1",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  true,
		},
		"session-2": {
			SessionID:  "session-2",
			PlayerID:   "player-2",
			PlayerName: "Player2",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  false,
		},
	}

	err = store.SaveAllSessions(originalSessions)
	require.NoError(t, err)

	loadedSessions, err := store.LoadAllSessions()
	require.NoError(t, err)
	assert.Len(t, loadedSessions, 2)

	for id, session := range originalSessions {
		loaded, exists := loadedSessions[id]
		require.True(t, exists, "session %s should exist", id)
		assert.Equal(t, session.SessionID, loaded.SessionID)
		assert.Equal(t, session.PlayerID, loaded.PlayerID)
		assert.Equal(t, session.PlayerName, loaded.PlayerName)
		assert.Equal(t, session.Connected, loaded.Connected)
	}
}

func TestFileSessionStore_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileSessionStore(tmpDir)
	require.NoError(t, err)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			session := &SessionData{
				SessionID:  "session-" + string(rune(index)),
				PlayerID:   "player-" + string(rune(index)),
				PlayerName: "Player" + string(rune(index)),
				LastActive: time.Now(),
				CreatedAt:  time.Now(),
				Connected:  index%2 == 0,
			}
			_ = store.SaveSession(session)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	sessionIDs, err := store.ListSessions()
	require.NoError(t, err)
	assert.Greater(t, len(sessionIDs), 0)
}
