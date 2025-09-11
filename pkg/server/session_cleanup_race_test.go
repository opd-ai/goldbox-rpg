package server

import (
	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/game"
	"sync"
	"testing"
	"time"
)

// TestSessionCleanupRaceCondition tests that session cleanup doesn't interfere with active handlers
func TestSessionCleanupRaceCondition(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
		config: &config.Config{
			SessionTimeout: 30 * time.Minute, // Use default timeout
		},
	}

	// Create a session that would normally be expired
	expiredSession := &PlayerSession{
		SessionID:   "expired-session",
		LastActive:  time.Now().Add(-2 * time.Hour), // Very old
		MessageChan: make(chan []byte, 100),
		Player: &game.Player{
			Character: game.Character{
				ID:   "test-player",
				Name: "Test Player",
			},
		},
	}
	server.sessions["expired-session"] = expiredSession

	// Step 1: Handler acquires the session (simulates what a real handler would do)
	handlerSession, exists := server.getSession("expired-session")
	if !exists {
		t.Fatal("Session should exist before cleanup")
	}

	// Step 2: Run cleanup while handler is "using" the session
	// The cleanup should skip this session because it's in use
	server.cleanupExpiredSessions()

	// Step 3: Verify session still exists (wasn't cleaned up)
	if _, exists := server.sessions["expired-session"]; !exists {
		t.Error("Session was cleaned up while in use - race condition detected!")
	}

	// Step 4: Simulate handler accessing session data (this would panic if session was deleted)
	playerID := handlerSession.Player.GetID()
	if playerID != "test-player" {
		t.Errorf("Session data corrupted: expected 'test-player', got '%s'", playerID)
	}

	// Step 5: Handler releases the session
	handlerSession.release()

	// Step 6: Now cleanup should be able to delete the session
	server.cleanupExpiredSessions()

	// Step 7: Verify session was cleaned up
	if _, exists := server.sessions["expired-session"]; exists {
		t.Error("Session should have been cleaned up after being released")
	}
}

// TestSessionCleanupRespectsInUse tests that cleanup skips sessions currently in use
func TestSessionCleanupRespectsInUse(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
		config: &config.Config{
			SessionTimeout: 30 * time.Minute, // Use default timeout
		},
	}

	// Create an expired session
	expiredSession := &PlayerSession{
		SessionID:   "in-use-session",
		LastActive:  time.Now().Add(-2 * time.Hour), // Very old
		MessageChan: make(chan []byte, 100),
		Player: &game.Player{
			Character: game.Character{
				ID:   "test-player",
				Name: "Test Player",
			},
		},
	}
	server.sessions["in-use-session"] = expiredSession

	// Mark session as in use
	expiredSession.addRef()

	// Try cleanup - should skip the in-use session
	server.cleanupExpiredSessions()

	// Session should still exist because it's in use
	if _, exists := server.sessions["in-use-session"]; !exists {
		t.Error("In-use session was incorrectly cleaned up")
	}

	// Release the session
	expiredSession.release()

	// Now cleanup should work
	server.cleanupExpiredSessions()

	// Session should now be cleaned up
	if _, exists := server.sessions["in-use-session"]; exists {
		t.Error("Session should have been cleaned up after being released")
	}
}
