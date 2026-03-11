package persistence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemorySessionStore(t *testing.T) {
	store := NewMemorySessionStore()
	require.NotNil(t, store)
	require.NotNil(t, store.sessions)
}

func TestMemorySessionStore_SaveLoadSession(t *testing.T) {
	store := NewMemorySessionStore()

	session := &SessionData{
		SessionID:  "test-session-1",
		PlayerID:   "player-123",
		PlayerName: "TestPlayer",
		LastActive: time.Now(),
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		Connected:  true,
	}

	err := store.SaveSession(session)
	require.NoError(t, err)

	loaded, err := store.LoadSession("test-session-1")
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, session.SessionID, loaded.SessionID)
	assert.Equal(t, session.PlayerID, loaded.PlayerID)
	assert.Equal(t, session.PlayerName, loaded.PlayerName)
	assert.Equal(t, session.Connected, loaded.Connected)
}

func TestMemorySessionStore_LoadNonExistentSession(t *testing.T) {
	store := NewMemorySessionStore()

	_, err := store.LoadSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestMemorySessionStore_DeleteSession(t *testing.T) {
	store := NewMemorySessionStore()

	session := &SessionData{
		SessionID:  "test-session-2",
		PlayerID:   "player-456",
		PlayerName: "TestPlayer2",
		LastActive: time.Now(),
		CreatedAt:  time.Now(),
		Connected:  false,
	}

	err := store.SaveSession(session)
	require.NoError(t, err)

	err = store.DeleteSession("test-session-2")
	require.NoError(t, err)

	_, err = store.LoadSession("test-session-2")
	assert.Error(t, err)
}

func TestMemorySessionStore_ListSessions(t *testing.T) {
	store := NewMemorySessionStore()

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

func TestMemorySessionStore_SaveAllSessions(t *testing.T) {
	store := NewMemorySessionStore()

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

	err := store.SaveAllSessions(sessions)
	require.NoError(t, err)

	sessionIDs, err := store.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessionIDs, 2)
}

func TestMemorySessionStore_LoadAllSessions(t *testing.T) {
	store := NewMemorySessionStore()

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

	err := store.SaveAllSessions(originalSessions)
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

func TestMemorySessionStore_SessionIsolation(t *testing.T) {
	store := NewMemorySessionStore()

	session := &SessionData{
		SessionID:  "test-session",
		PlayerID:   "player-1",
		PlayerName: "Player1",
		LastActive: time.Now(),
		CreatedAt:  time.Now(),
		Connected:  true,
	}

	err := store.SaveSession(session)
	require.NoError(t, err)

	session.PlayerName = "ModifiedPlayer"

	loaded, err := store.LoadSession("test-session")
	require.NoError(t, err)
	assert.Equal(t, "Player1", loaded.PlayerName, "session should be isolated from external modifications")
}

func TestMemorySessionStore_ConcurrentAccess(t *testing.T) {
	store := NewMemorySessionStore()

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

func TestMemorySessionStore_EmptyStore(t *testing.T) {
	store := NewMemorySessionStore()

	sessionIDs, err := store.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessionIDs, 0)

	sessions, err := store.LoadAllSessions()
	require.NoError(t, err)
	assert.Len(t, sessions, 0)
}
