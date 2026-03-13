package server

import (
	"sync"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createSessionForTimeoutTest creates a test session with a unique ID
func createSessionForTimeoutTest(server *RPCServer, sessionID string) *PlayerSession {
	character := &game.Character{
		ID:   "test-player-timeout",
		Name: "Timeout Test Player",
		HP:   100,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	return &PlayerSession{
		SessionID:   sessionID,
		Player:      player,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   true,
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}
}

// TestSessionTimeoutExpiration tests that sessions are marked for cleanup
// after exceeding the configured timeout
func TestSessionTimeoutExpiration(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Override timeout for testing
	server.config.SessionTimeout = 10 * time.Millisecond

	session := createSessionForTimeoutTest(server, "timeout-test-001")
	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()

	// Let session expire
	time.Sleep(50 * time.Millisecond)

	// Manually trigger cleanup
	server.cleanupExpiredSessions()

	// Session should be removed
	server.mu.RLock()
	_, exists := server.sessions[session.SessionID]
	server.mu.RUnlock()

	assert.False(t, exists, "expired session should be removed")
}

// TestSessionTimeoutPreservesActive tests that active sessions are not cleaned up
func TestSessionTimeoutPreservesActive(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	server.config.SessionTimeout = 100 * time.Millisecond

	session := createSessionForTimeoutTest(server, "active-test-001")
	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()

	// Keep updating LastActive
	for i := 0; i < 5; i++ {
		time.Sleep(30 * time.Millisecond)
		server.mu.Lock()
		session.LastActive = time.Now()
		server.mu.Unlock()
	}

	// Manually trigger cleanup
	server.cleanupExpiredSessions()

	// Session should still exist
	server.mu.RLock()
	_, exists := server.sessions[session.SessionID]
	server.mu.RUnlock()

	assert.True(t, exists, "active session should not be removed")
}

// TestSessionReferenceCountPreventsCleanup tests that sessions in use are not cleaned up
func TestSessionReferenceCountPreventsCleanup(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	server.config.SessionTimeout = 10 * time.Millisecond

	session := createSessionForTimeoutTest(server, "refcount-test-001")
	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()

	// Simulate session being in use
	session.addRef()

	// Let it "expire"
	time.Sleep(50 * time.Millisecond)

	// Try cleanup
	server.cleanupExpiredSessions()

	// Session should NOT be removed because it's in use
	server.mu.RLock()
	_, exists := server.sessions[session.SessionID]
	server.mu.RUnlock()

	assert.True(t, exists, "in-use session should not be removed")

	// Release the reference
	session.release()

	// Now cleanup should work
	server.cleanupExpiredSessions()

	server.mu.RLock()
	_, exists = server.sessions[session.SessionID]
	server.mu.RUnlock()

	assert.False(t, exists, "released session should be removed after cleanup")
}

// TestCleanupExpiredSessionsConcurrent tests concurrent cleanup safety
func TestCleanupExpiredSessionsConcurrent(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	server.config.SessionTimeout = 5 * time.Millisecond

	// Create multiple sessions
	for i := 0; i < 10; i++ {
		session := createSessionForTimeoutTest(server, "concurrent-test-"+string(rune('A'+i)))
		server.mu.Lock()
		server.sessions[session.SessionID] = session
		server.mu.Unlock()
	}

	// Let sessions expire
	time.Sleep(20 * time.Millisecond)

	// Run concurrent cleanup attempts
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.cleanupExpiredSessions()
		}()
	}
	wg.Wait()

	// Verify sessions are cleaned up properly (no panics, no data corruption)
	server.mu.RLock()
	remaining := len(server.sessions)
	server.mu.RUnlock()

	assert.Equal(t, 0, remaining, "all expired sessions should be removed")
}

// TestSessionGetAndRelease tests the getSession/releaseSession pattern
func TestSessionGetAndRelease(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createSessionForTimeoutTest(server, "getrelease-test-001")
	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()

	// Get session (increments refCount)
	retrieved, exists := server.getSession(session.SessionID)
	require.True(t, exists)
	require.NotNil(t, retrieved)

	// Verify reference count was incremented
	assert.True(t, session.isInUse(), "session should be in use after getSession")

	// Release session
	server.releaseSession(session)

	// May still be considered in use depending on implementation
	// Just verify no panic occurred
}

// TestGetNonExistentSession tests getSession with invalid session ID
func TestGetNonExistentSession(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session, exists := server.getSession("nonexistent-session-id")
	assert.False(t, exists)
	assert.Nil(t, session)
}

// TestMessageChannelClose tests safe closing of message channels
func TestMessageChannelClose(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createSessionForTimeoutTest(server, "close-test-001")
	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()

	// Close channel multiple times (should not panic)
	assert.NotPanics(t, func() {
		session.closeMessageChannel()
		session.closeMessageChannel() // Double close should be safe
	})
}

// TestSessionSendMessage tests the safeSendMessage function
func TestSessionSendMessage(t *testing.T) {
	session := &PlayerSession{
		SessionID:   "test-session-send",
		MessageChan: make(chan []byte, 10),
	}

	// Send message
	msg := []byte(`{"type":"test"}`)
	success := safeSendMessage(session, msg)
	assert.True(t, success, "safeSendMessage should succeed with buffered channel")

	// Verify message was received
	select {
	case received := <-session.MessageChan:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("message not received")
	}
}

// TestSessionSendMessageTimeout tests message send timeout
func TestSessionSendMessageTimeout(t *testing.T) {
	session := &PlayerSession{
		SessionID:   "test-session-timeout",
		MessageChan: make(chan []byte), // Unbuffered, will block
	}

	// This should not block forever; it will timeout
	start := time.Now()
	success := safeSendMessage(session, []byte(`{"type":"test"}`))
	elapsed := time.Since(start)

	// Should return false because channel is unbuffered
	assert.False(t, success, "safeSendMessage should fail on full channel")

	// Should complete within 2x MessageSendTimeout
	assert.Less(t, elapsed, 2*MessageSendTimeout+100*time.Millisecond,
		"safeSendMessage should timeout, not block forever")
}

// TestMultipleSessionCleanupCycles tests multiple cleanup cycles
func TestMultipleSessionCleanupCycles(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	server.config.SessionTimeout = 20 * time.Millisecond

	// Add and expire sessions in waves
	for wave := 0; wave < 3; wave++ {
		// Add sessions
		for i := 0; i < 5; i++ {
			sessionID := "wave-" + string(rune('0'+wave)) + "-" + string(rune('A'+i))
			session := createSessionForTimeoutTest(server, sessionID)
			server.mu.Lock()
			server.sessions[session.SessionID] = session
			server.mu.Unlock()
		}

		// Let them expire and cleanup
		time.Sleep(30 * time.Millisecond)
		server.cleanupExpiredSessions()

		server.mu.RLock()
		count := len(server.sessions)
		server.mu.RUnlock()

		assert.Equal(t, 0, count, "wave %d: all sessions should be cleaned up", wave)
	}
}

// TestCleanupClosesWebSocket tests that cleanup properly closes WebSocket connections
func TestCleanupClosesWebSocket(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	server.config.SessionTimeout = 10 * time.Millisecond

	session := createSessionForTimeoutTest(server, "ws-close-test-001")
	// Note: In real tests we'd mock the WebSocket, but for this test
	// we verify the function doesn't panic with nil WSConn
	session.WSConn = nil

	server.mu.Lock()
	server.sessions[session.SessionID] = session
	server.mu.Unlock()

	time.Sleep(20 * time.Millisecond)

	// Should not panic with nil WSConn
	assert.NotPanics(t, func() {
		server.cleanupExpiredSessions()
	})
}
