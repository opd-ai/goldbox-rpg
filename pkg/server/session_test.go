package server

import (
	"crypto/tls"
	"goldbox-rpg/pkg/config"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestGetOrCreateSession_CreateNewSession(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
		config: &config.Config{
			SessionTimeout: 30 * time.Minute,
		},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	session, err := server.getOrCreateSession(w, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if session == nil {
		t.Fatal("Expected session to be created, got nil")
	}

	if session.SessionID == "" {
		t.Error("Expected non-empty SessionID")
	}

	if session.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if session.LastActive.IsZero() {
		t.Error("Expected LastActive to be set")
	}

	if session.MessageChan == nil {
		t.Error("Expected MessageChan to be initialized")
	}

	// Check that session was stored in server
	if _, exists := server.sessions[session.SessionID]; !exists {
		t.Error("Expected session to be stored in server sessions map")
	}

	// Check cookie was set
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != "session_id" {
		t.Errorf("Expected cookie name 'session_id', got '%s'", cookie.Name)
	}

	if cookie.Value != session.SessionID {
		t.Errorf("Expected cookie value '%s', got '%s'", session.SessionID, cookie.Value)
	}

	if cookie.HttpOnly != true {
		t.Error("Expected cookie to be HttpOnly")
	}

	if cookie.MaxAge != 1800 {
		t.Errorf("Expected cookie MaxAge 1800, got %d", cookie.MaxAge)
	}
}

func TestGetOrCreateSession_RetrieveExistingSession(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	// Create an existing session
	existingSessionID := "test-session-123"
	existingSession := &PlayerSession{
		SessionID:   existingSessionID,
		CreatedAt:   time.Now().Add(-1 * time.Hour),
		LastActive:  time.Now().Add(-5 * time.Minute),
		MessageChan: make(chan []byte, 100),
	}
	server.sessions[existingSessionID] = existingSession

	// Create request with session cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: existingSessionID,
	})
	w := httptest.NewRecorder()

	session, err := server.getOrCreateSession(w, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if session == nil {
		t.Fatal("Expected session to be retrieved, got nil")
	}

	if session.SessionID != existingSessionID {
		t.Errorf("Expected SessionID '%s', got '%s'", existingSessionID, session.SessionID)
	}

	// Verify LastActive was updated
	if session.LastActive.Before(time.Now().Add(-1 * time.Minute)) {
		t.Error("Expected LastActive to be updated to recent time")
	}

	// Should be the same session object
	if session != existingSession {
		t.Error("Expected to get the same session object")
	}

	// No new cookie should be set since session exists
	cookies := w.Result().Cookies()
	if len(cookies) != 0 {
		t.Errorf("Expected no new cookies, got %d", len(cookies))
	}
}

func TestGetOrCreateSession_InvalidSessionCookie(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	// Create request with invalid session cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "invalid-session-id",
	})
	w := httptest.NewRecorder()

	session, err := server.getOrCreateSession(w, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if session == nil {
		t.Fatal("Expected new session to be created, got nil")
	}

	// Should create a new session since the cookie was invalid
	if session.SessionID == "invalid-session-id" {
		t.Error("Expected new SessionID, got the invalid one")
	}

	// Should set a new cookie
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	if cookies[0].Value == "invalid-session-id" {
		t.Error("Expected new cookie value, got the invalid one")
	}
}

func TestGetOrCreateSession_ConcurrentAccess(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	// Test concurrent access to session creation
	const numGoroutines = 10
	results := make(chan *PlayerSession, numGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			session, err := server.getOrCreateSession(w, req)
			if err != nil {
				t.Errorf("Unexpected error in goroutine: %v", err)
				return
			}
			results <- session
		}()
	}

	wg.Wait()
	close(results)

	// Collect all sessions
	var sessions []*PlayerSession
	for session := range results {
		sessions = append(sessions, session)
	}

	if len(sessions) != numGoroutines {
		t.Fatalf("Expected %d sessions, got %d", numGoroutines, len(sessions))
	}

	// All sessions should have unique IDs
	sessionIDs := make(map[string]bool)
	for _, session := range sessions {
		if sessionIDs[session.SessionID] {
			t.Errorf("Duplicate session ID found: %s", session.SessionID)
		}
		sessionIDs[session.SessionID] = true
	}

	// Server should have all sessions stored
	if len(server.sessions) != numGoroutines {
		t.Errorf("Expected %d sessions in server, got %d", numGoroutines, len(server.sessions))
	}
}

func TestStartSessionCleanup(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
		done:     make(chan struct{}),
	}

	// Create test sessions with different activity times
	activeSession := &PlayerSession{
		SessionID:   "active-session",
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, 100),
	}

	expiredSession := &PlayerSession{
		SessionID:   "expired-session",
		LastActive:  time.Now().Add(-2 * time.Hour), // Very old
		MessageChan: make(chan []byte, 100),
	}

	server.sessions["active-session"] = activeSession
	server.sessions["expired-session"] = expiredSession

	// Start cleanup
	server.startSessionCleanup()

	// Wait for cleanup to run (since cleanup runs every 5 minutes, we need to manually trigger it)
	// We'll call cleanupExpiredSessions directly to test the logic
	server.cleanupExpiredSessions()

	// Stop the cleanup
	close(server.done)

	// Check that expired session was removed and active session remains
	server.mu.Lock()
	defer server.mu.Unlock()

	if _, exists := server.sessions["active-session"]; !exists {
		t.Error("Expected active session to remain")
	}

	if _, exists := server.sessions["expired-session"]; exists {
		t.Error("Expected expired session to be removed")
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	// Create test sessions
	recentSession := &PlayerSession{
		SessionID:   "recent",
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, 100),
	}

	oldSession := &PlayerSession{
		SessionID:   "old",
		LastActive:  time.Now().Add(-2 * time.Hour), // Older than 30 minutes
		MessageChan: make(chan []byte, 100),
	}

	server.sessions["recent"] = recentSession
	server.sessions["old"] = oldSession

	// Run cleanup (uses the default 30 minute timeout)
	server.cleanupExpiredSessions()

	// Check results
	if _, exists := server.sessions["recent"]; !exists {
		t.Error("Expected recent session to remain")
	}

	if _, exists := server.sessions["old"]; exists {
		t.Error("Expected old session to be removed")
	}

	if len(server.sessions) != 1 {
		t.Errorf("Expected 1 session remaining, got %d", len(server.sessions))
	}
}

func TestCleanupExpiredSessions_WithWebSocketConnection(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	// Create a mock websocket connection (we can't easily test the actual closing behavior in unit tests)
	expiredSession := &PlayerSession{
		SessionID:   "expired-with-ws",
		LastActive:  time.Now().Add(-2 * time.Hour), // Older than 30 minutes
		MessageChan: make(chan []byte, 100),
		WSConn:      nil, // In real usage this would be a websocket connection
	}

	server.sessions["expired-with-ws"] = expiredSession

	// Run cleanup (uses the default 30 minute timeout)
	server.cleanupExpiredSessions()

	// Session should be removed
	if _, exists := server.sessions["expired-with-ws"]; exists {
		t.Error("Expected expired session with websocket to be removed")
	}
}

func TestGetOrCreateSession_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		existingSessionID string
		cookieValue       string
		expectNewSession  bool
		expectError       bool
	}{
		{
			name:              "No cookie provided",
			existingSessionID: "",
			cookieValue:       "",
			expectNewSession:  true,
			expectError:       false,
		},
		{
			name:              "Valid cookie with existing session",
			existingSessionID: "valid-session-123",
			cookieValue:       "valid-session-123",
			expectNewSession:  false,
			expectError:       false,
		},
		{
			name:              "Invalid cookie value",
			existingSessionID: "valid-session-123",
			cookieValue:       "invalid-session-456",
			expectNewSession:  true,
			expectError:       false,
		},
		{
			name:              "Empty cookie value",
			existingSessionID: "valid-session-123",
			cookieValue:       "",
			expectNewSession:  true,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &RPCServer{
				sessions: make(map[string]*PlayerSession),
				mu:       sync.RWMutex{},
			}

			// Set up existing session if specified
			var existingSession *PlayerSession
			if tt.existingSessionID != "" {
				existingSession = &PlayerSession{
					SessionID:   tt.existingSessionID,
					CreatedAt:   time.Now().Add(-1 * time.Hour),
					LastActive:  time.Now().Add(-5 * time.Minute),
					MessageChan: make(chan []byte, 100),
				}
				server.sessions[tt.existingSessionID] = existingSession
			}

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tt.cookieValue,
				})
			}
			w := httptest.NewRecorder()

			// Call function
			session, err := server.getOrCreateSession(w, req)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check session creation expectation
			if tt.expectNewSession {
				if session == nil {
					t.Fatal("Expected new session to be created")
				}
				if session == existingSession {
					t.Error("Expected new session, got existing session")
				}
				// Should have set a cookie
				cookies := w.Result().Cookies()
				if len(cookies) != 1 {
					t.Errorf("Expected 1 cookie for new session, got %d", len(cookies))
				}
			} else {
				if session == nil {
					t.Fatal("Expected existing session to be returned")
				}
				if session != existingSession {
					t.Error("Expected existing session, got different session")
				}
				// Should not have set a new cookie
				cookies := w.Result().Cookies()
				if len(cookies) != 0 {
					t.Errorf("Expected no new cookies for existing session, got %d", len(cookies))
				}
			}
		})
	}
}

// TestSecureCookieSettings tests that cookies are set with appropriate security settings
func TestSecureCookieSettings(t *testing.T) {
	tests := []struct {
		name            string
		useTLS          bool
		xForwardedProto string
		expectSecure    bool
	}{
		{
			name:            "HTTPS connection should set secure cookie",
			useTLS:          true,
			xForwardedProto: "",
			expectSecure:    true,
		},
		{
			name:            "HTTP with X-Forwarded-Proto https should set secure cookie",
			useTLS:          false,
			xForwardedProto: "https",
			expectSecure:    true,
		},
		{
			name:            "HTTP connection should not set secure cookie",
			useTLS:          false,
			xForwardedProto: "",
			expectSecure:    false,
		},
		{
			name:            "HTTP with X-Forwarded-Proto http should not set secure cookie",
			useTLS:          false,
			xForwardedProto: "http",
			expectSecure:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := &RPCServer{
				sessions: make(map[string]*PlayerSession),
				mu:       sync.RWMutex{},
			}

			req := httptest.NewRequest("GET", "/test", nil)
			if test.xForwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", test.xForwardedProto)
			}
			if test.useTLS {
				req.TLS = &tls.ConnectionState{}
			}

			w := httptest.NewRecorder()

			_, err := server.getOrCreateSession(w, req)
			if err != nil {
				t.Fatalf("getOrCreateSession failed: %v", err)
			}

			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("Expected 1 cookie, got %d", len(cookies))
			}

			cookie := cookies[0]
			if cookie.Secure != test.expectSecure {
				t.Errorf("Expected Secure=%v, got Secure=%v", test.expectSecure, cookie.Secure)
			}

			if cookie.SameSite != http.SameSiteStrictMode {
				t.Errorf("Expected SameSite=Strict, got SameSite=%v", cookie.SameSite)
			}

			if !cookie.HttpOnly {
				t.Error("Expected HttpOnly=true")
			}
		})
	}
}

// TestSafeSendMessage_ChannelFull tests backpressure handling when message channel is full
func TestSafeSendMessage_ChannelFull(t *testing.T) {
	// Create session with small buffer for testing
	session := &PlayerSession{
		SessionID:   "test-session",
		MessageChan: make(chan []byte, 2), // Small buffer to easily fill
	}

	// Fill the channel buffer
	message1 := []byte("message1")
	message2 := []byte("message2")

	// These should succeed
	if !safeSendMessage(session, message1) {
		t.Error("Expected first message to be sent successfully")
	}
	if !safeSendMessage(session, message2) {
		t.Error("Expected second message to be sent successfully")
	}

	// This should fail due to full buffer and timeout
	message3 := []byte("message3")
	if safeSendMessage(session, message3) {
		t.Error("Expected third message to be dropped due to full channel")
	}

	// Verify only 2 messages are in channel
	close(session.MessageChan)
	count := 0
	for range session.MessageChan {
		count++
	}
	if count != 2 {
		t.Errorf("Expected 2 messages in channel, got %d", count)
	}
}

// TestSafeSendMessage_NilSession tests handling of nil session
func TestSafeSendMessage_NilSession(t *testing.T) {
	if safeSendMessage(nil, []byte("test")) {
		t.Error("Expected safeSendMessage to return false for nil session")
	}
}

// TestSafeSendMessage_NilChannel tests handling of session with nil channel
func TestSafeSendMessage_NilChannel(t *testing.T) {
	session := &PlayerSession{
		SessionID:   "test-session",
		MessageChan: nil,
	}

	if safeSendMessage(session, []byte("test")) {
		t.Error("Expected safeSendMessage to return false for nil channel")
	}
}

// TestHandleCreateCharacter_SessionCollisionDetection tests that character creation handles session ID collisions properly
func TestHandleCreateCharacter_SessionCollisionDetection(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
	}

	// Pre-populate sessions map with a session to force potential collision detection
	existingSessionID := "collision-test-session"
	existingSession := &PlayerSession{
		SessionID:   existingSessionID,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		MessageChan: make(chan []byte, 100),
	}
	server.sessions[existingSessionID] = existingSession

	// Test character creation request that could potentially generate a collision
	// We can't easily force a UUID collision, but we can verify the collision detection logic
	// by checking that sessions are properly stored and don't overwrite existing ones

	params := []byte(`{
		"name": "TestCharacter",
		"class": "fighter",
		"attribute_method": "standard",
		"starting_equipment": true,
		"starting_gold": 100
	}`)

	response, err := server.handleCreateCharacter(params)

	if err != nil {
		t.Fatalf("Character creation failed: %v", err)
	}

	result, ok := response.(map[string]interface{})
	if !ok {
		t.Fatal("Response is not a map")
	}

	success, exists := result["success"]
	if !exists || success != true {
		t.Fatalf("Character creation was not successful: %v", result)
	}

	sessionID, exists := result["session_id"]
	if !exists {
		t.Fatal("No session_id in response")
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok {
		t.Fatal("session_id is not a string")
	}

	// Verify that the new session was created and doesn't overwrite the existing one
	server.mu.RLock()
	defer server.mu.RUnlock()

	if len(server.sessions) != 2 {
		t.Errorf("Expected 2 sessions (original + new), got %d", len(server.sessions))
	}

	// Verify original session still exists
	if _, exists := server.sessions[existingSessionID]; !exists {
		t.Error("Original session was overwritten - collision detection failed")
	}

	// Verify new session exists and is different
	newSession, exists := server.sessions[sessionIDStr]
	if !exists {
		t.Error("New session was not created")
	}

	if newSession.SessionID == existingSessionID {
		t.Error("New session has same ID as existing session - collision detection failed")
	}

	// Verify the new session has correct structure
	if newSession.Player == nil {
		t.Error("New session has no player data")
	}
}
