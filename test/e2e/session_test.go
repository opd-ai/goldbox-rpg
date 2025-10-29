package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionWorkflow tests the complete session lifecycle
func TestSessionWorkflow(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	t.Run("join_game_creates_session", func(t *testing.T) {
		sessionID, err := client.JoinGame("TestPlayer")
		require.NoError(t, err, "should join game successfully")
		AssertSessionID(t, sessionID)
	})

	t.Run("join_game_with_empty_name_generates_default", func(t *testing.T) {
		sessionID, err := client.JoinGame("")
		require.NoError(t, err, "should join game with empty name")
		AssertSessionID(t, sessionID)
	})

	t.Run("get_game_state_with_valid_session", func(t *testing.T) {
		sessionID, err := client.JoinGame("TestPlayer")
		require.NoError(t, err)

		state, err := client.GetGameState(sessionID)
		require.NoError(t, err, "should get game state successfully")
		AssertGameState(t, state)
	})

	t.Run("get_game_state_with_invalid_session", func(t *testing.T) {
		_, err := client.GetGameState("invalid-session-id")
		require.Error(t, err, "should fail with invalid session ID")
		ErrorContains(t, err, "session")
	})
}

// TestSessionConcurrency tests concurrent session creation
func TestSessionConcurrency(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	// Create multiple sessions concurrently
	numSessions := 5
	sessionCh := make(chan string, numSessions)
	errCh := make(chan error, numSessions)

	for i := 0; i < numSessions; i++ {
		go func(i int) {
			sessionID, err := client.JoinGame(RandomCharacterName())
			if err != nil {
				errCh <- err
				return
			}
			sessionCh <- sessionID
		}(i)
	}

	// Collect results
	sessions := make([]string, 0, numSessions)
	for i := 0; i < numSessions; i++ {
		select {
		case sessionID := <-sessionCh:
			sessions = append(sessions, sessionID)
		case err := <-errCh:
			t.Fatalf("error creating session: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatal("timeout waiting for sessions")
		}
	}

	// Verify all sessions are unique
	assert.Len(t, sessions, numSessions, "should create correct number of sessions")
	sessionMap := make(map[string]bool)
	for _, sessionID := range sessions {
		assert.False(t, sessionMap[sessionID], "session IDs should be unique")
		sessionMap[sessionID] = true
	}
}

// TestSessionTimeout tests session timeout behavior
func TestSessionTimeout(t *testing.T) {
	t.Skip("Requires shorter session timeout for testing")

	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	// Create session
	sessionID, err := client.JoinGame("TestPlayer")
	require.NoError(t, err)

	// Verify session is valid
	_, err = client.GetGameState(sessionID)
	require.NoError(t, err, "session should be valid initially")

	// Wait for session to timeout (this test would need a shorter timeout in config)
	time.Sleep(35 * time.Second)

	// Verify session is no longer valid
	_, err = client.GetGameState(sessionID)
	require.Error(t, err, "session should timeout")
	ErrorContains(t, err, "session")
}

// TestMultipleClients tests multiple clients connecting simultaneously
func TestMultipleClients(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	baseURL := helper.Server().BaseURL()

	// Create multiple clients
	numClients := 3
	clients := make([]*Client, numClients)
	sessions := make([]string, numClients)

	for i := 0; i < numClients; i++ {
		clients[i] = NewClient(baseURL)
		defer clients[i].Close()

		sessionID, err := clients[i].JoinGame(RandomCharacterName())
		require.NoError(t, err, "client %d should join game", i)
		sessions[i] = sessionID
	}

	// Verify each client can access their own session
	for i, client := range clients {
		state, err := client.GetGameState(sessions[i])
		require.NoError(t, err, "client %d should get game state", i)
		AssertGameState(t, state)
	}

	// Verify clients cannot access each other's sessions
	_, err := clients[0].GetGameState(sessions[1])
	require.Error(t, err, "client should not access another client's session")
}
