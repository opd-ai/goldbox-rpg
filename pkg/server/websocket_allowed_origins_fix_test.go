package server

import (
	"net/http/httptest"
	"os"
	"testing"

	"goldbox-rpg/pkg/config"
)

// TestWebSocketOriginValidation_WebSocketAllowedOriginsBug demonstrates and validates the fix for
// Bug 4 from AUDIT.md where WEBSOCKET_ALLOWED_ORIGINS environment variable was documented in
// README.md but not actually used by the WebSocket upgrader implementation.
func TestWebSocketOriginValidation_WebSocketAllowedOriginsBug(t *testing.T) {
	// Save original environment
	originalWebSocketOrigins := os.Getenv("WEBSOCKET_ALLOWED_ORIGINS")
	originalAllowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	defer func() {
		os.Setenv("WEBSOCKET_ALLOWED_ORIGINS", originalWebSocketOrigins)
		os.Setenv("ALLOWED_ORIGINS", originalAllowedOrigins)
	}()

	// Clear both environment variables first
	os.Unsetenv("WEBSOCKET_ALLOWED_ORIGINS")
	os.Unsetenv("ALLOWED_ORIGINS")

	// Set WEBSOCKET_ALLOWED_ORIGINS as documented in README.md
	os.Setenv("WEBSOCKET_ALLOWED_ORIGINS", "https://documented.example.com")

	// Load config - this will NOT pick up WEBSOCKET_ALLOWED_ORIGINS
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Force production mode to test strict validation
	cfg.EnableDevMode = false

	// Create server with this config (reuse existing helper function)
	server, cleanup, err := createTestServerWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer cleanup()

	// Get the WebSocket upgrader
	upgrader := server.upgrader()

	// Test that WEBSOCKET_ALLOWED_ORIGINS is now properly honored
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "https://documented.example.com")

	allowed := upgrader.CheckOrigin(req)

	// After fix: This should be true because WEBSOCKET_ALLOWED_ORIGINS is now used
	if !allowed {
		t.Error("WEBSOCKET_ALLOWED_ORIGINS should be honored after bug fix")
	}

	// Test that unlisted origins are still rejected
	reqUnlisted := httptest.NewRequest("GET", "/ws", nil)
	reqUnlisted.Header.Set("Origin", "https://not.allowed.com")

	allowedUnlisted := upgrader.CheckOrigin(reqUnlisted)
	if allowedUnlisted {
		t.Error("Unlisted origins should still be rejected in production mode")
	}

	// Test fallback to ALLOWED_ORIGINS when WEBSOCKET_ALLOWED_ORIGINS is not set
	os.Unsetenv("WEBSOCKET_ALLOWED_ORIGINS")
	os.Setenv("ALLOWED_ORIGINS", "https://config.example.com")

	// Reload config to pick up ALLOWED_ORIGINS
	cfg2, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}
	cfg2.EnableDevMode = false

	server2, cleanup2, err := createTestServerWithConfig(cfg2)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer cleanup2()

	upgrader2 := server2.upgrader()

	// Test that ALLOWED_ORIGINS works as fallback
	reqFallback := httptest.NewRequest("GET", "/ws", nil)
	reqFallback.Header.Set("Origin", "https://config.example.com")

	allowedFallback := upgrader2.CheckOrigin(reqFallback)
	if !allowedFallback {
		t.Error("ALLOWED_ORIGINS should work as fallback when WEBSOCKET_ALLOWED_ORIGINS is not set")
	}
}
