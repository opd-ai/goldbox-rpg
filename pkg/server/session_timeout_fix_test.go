package server

import (
	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
	"goldbox-rpg/pkg/validation"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// createServerForTimeoutTest creates an RPCServer instance for testing with custom configuration.
// This bypasses the normal NewRPCServer constructor to allow custom config injection.
func createServerForTimeoutTest(cfg *config.Config) (*RPCServer, error) {
	validator := validation.NewInputValidator(cfg.MaxRequestSize)
	
	// Create minimal spell manager for testing
	spellManager := game.NewSpellManager("./test-data")
	
	// Create minimal PCG manager for testing  
	world := game.CreateDefaultWorld()
	logger := logrus.New()
	pcgManager := pcg.NewPCGManager(world, logger)
	
	server := createServerInstance("./web", cfg, validator, spellManager, pcgManager)
	return server, nil
}

// TestSessionTimeoutConfigurationUsage validates that session cleanup logic uses
// the configured SessionTimeout value instead of hardcoded constants.
// This addresses the audit recommendation to eliminate hardcoded timeouts.
func TestSessionTimeoutConfigurationUsage(t *testing.T) {
	tests := []struct {
		name           string
		sessionTimeout time.Duration
		description    string
	}{
		{
			name:           "custom_15_minute_timeout",
			sessionTimeout: 15 * time.Minute,
			description:    "Verifies session cleanup respects custom 15-minute timeout",
		},
		{
			name:           "custom_45_minute_timeout", 
			sessionTimeout: 45 * time.Minute,
			description:    "Verifies session cleanup respects custom 45-minute timeout",
		},
		{
			name:           "custom_60_minute_timeout",
			sessionTimeout: 60 * time.Minute,
			description:    "Verifies session cleanup respects custom 60-minute timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create server with custom session timeout
			cfg := &config.Config{
				ServerPort:     8080,
				WebDir:         "./web",
				SessionTimeout: tt.sessionTimeout,
				LogLevel:       "info",
				AllowedOrigins: []string{},
				MaxRequestSize: 1024 * 1024,
				EnableDevMode:  true,
				RequestTimeout: 30 * time.Second,
			}

			server, err := createServerForTimeoutTest(cfg)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			// Test cookie MaxAge uses configured timeout
			t.Run("cookie_max_age_uses_config", func(t *testing.T) {
				req := httptest.NewRequest("GET", "/", nil)
				w := httptest.NewRecorder()

				session, err := server.getOrCreateSession(w, req)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}

				// Check that cookie MaxAge matches configured timeout
				cookies := w.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "session_id" {
						sessionCookie = cookie
						break
					}
				}

				if sessionCookie == nil {
					t.Fatal("Session cookie not found")
				}

				expectedMaxAge := int(tt.sessionTimeout.Seconds())
				if sessionCookie.MaxAge != expectedMaxAge {
					t.Errorf("Cookie MaxAge = %d, want %d (from config.SessionTimeout %v)",
						sessionCookie.MaxAge, expectedMaxAge, tt.sessionTimeout)
				}

				// Clean up session for next test
				server.mu.Lock()
				delete(server.sessions, session.SessionID)
				server.mu.Unlock()
			})

			// Test session cleanup uses configured timeout
			t.Run("cleanup_uses_config_timeout", func(t *testing.T) {
				// Create a session manually for testing
				sessionID := "test-session-cleanup"
				testSession := &PlayerSession{
					SessionID:   sessionID,
					CreatedAt:   time.Now(),
					LastActive:  time.Now().Add(-tt.sessionTimeout).Add(-1*time.Minute), // Expired by 1 minute
					MessageChan: make(chan []byte, MessageChanBufferSize),
				}

				server.mu.Lock()
				server.sessions[sessionID] = testSession
				server.mu.Unlock()

				// Verify session exists before cleanup
				server.mu.RLock()
				_, exists := server.sessions[sessionID]
				server.mu.RUnlock()
				if !exists {
					t.Fatal("Test session not found before cleanup")
				}

				// Run cleanup
				server.cleanupExpiredSessions()

				// Verify session was removed after cleanup (since it's expired)
				server.mu.RLock()
				_, exists = server.sessions[sessionID]
				server.mu.RUnlock()
				if exists {
					t.Error("Expired session should have been removed by cleanup")
				}
			})

			// Test that non-expired sessions are preserved
			t.Run("active_sessions_preserved", func(t *testing.T) {
				// Create a session that's not expired
				sessionID := "test-session-active"
				testSession := &PlayerSession{
					SessionID:   sessionID,
					CreatedAt:   time.Now(),
					LastActive:  time.Now().Add(-tt.sessionTimeout).Add(5*time.Minute), // Not expired yet
					MessageChan: make(chan []byte, MessageChanBufferSize),
				}

				server.mu.Lock()
				server.sessions[sessionID] = testSession
				server.mu.Unlock()

				// Run cleanup
				server.cleanupExpiredSessions()

				// Verify session still exists (not expired)
				server.mu.RLock()
				_, exists := server.sessions[sessionID]
				server.mu.RUnlock()
				if !exists {
					t.Error("Active session should not have been removed by cleanup")
				}

				// Clean up for next test
				server.mu.Lock()
				delete(server.sessions, sessionID)
				server.mu.Unlock()
			})
		})
	}
}

// TestSessionTimeoutRegression is a regression test to ensure that session cleanup
// no longer uses hardcoded 30-minute timeouts and respects configuration.
func TestSessionTimeoutRegression(t *testing.T) {
	// Create a server with a non-default timeout to verify it's actually used
	customTimeout := 75 * time.Minute // Different from the old hardcoded 30 minutes
	cfg := &config.Config{
		ServerPort:     8080,
		WebDir:         "./web",
		SessionTimeout: customTimeout,
		LogLevel:       "info",
		AllowedOrigins: []string{},
		MaxRequestSize: 1024 * 1024,
		EnableDevMode:  true,
		RequestTimeout: 30 * time.Second,
	}

	server, err := createServerForTimeoutTest(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a session that would be expired under the old 30-minute hardcoded timeout
	// but should still be active under our 75-minute configured timeout
	sessionID := "regression-test-session"
	testSession := &PlayerSession{
		SessionID:   sessionID,
		CreatedAt:   time.Now(),
		LastActive:  time.Now().Add(-45 * time.Minute), // 45 minutes ago
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}

	server.mu.Lock()
	server.sessions[sessionID] = testSession
	server.mu.Unlock()

	// Run cleanup
	server.cleanupExpiredSessions()

	// Under the old hardcoded 30-minute timeout, this session would be removed
	// Under our new 75-minute configured timeout, it should still exist
	server.mu.RLock()
	_, exists := server.sessions[sessionID]
	server.mu.RUnlock()

	if !exists {
		t.Error("Session should still exist under 75-minute configured timeout, " +
			"but was removed (suggests hardcoded 30-minute timeout still in use)")
	}

	// Clean up
	server.mu.Lock()
	delete(server.sessions, sessionID)
	server.mu.Unlock()
}
