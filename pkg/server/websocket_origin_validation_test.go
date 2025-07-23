package server

import (
	"net/http/httptest"
	"os"
	"testing"

	"goldbox-rpg/pkg/config"
)

// TestWebSocketOriginValidation_ProductionMode tests WebSocket origin validation
// in production mode with explicit allowed origins configuration.
func TestWebSocketOriginValidation_ProductionMode(t *testing.T) {
	// Create a config for production mode with specific allowed origins
	cfg := &config.Config{
		EnableDevMode:  false, // Production mode
		AllowedOrigins: []string{"https://game.example.com", "https://secure.game.com"},
		ServerPort:     8080,
		WebDir:         "./test_web",
	}

	server, cleanup, err := createTestServerWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer cleanup()

	upgrader := server.upgrader()

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{
			name:     "allowed origin should pass",
			origin:   "https://game.example.com",
			expected: true,
		},
		{
			name:     "another allowed origin should pass",
			origin:   "https://secure.game.com",
			expected: true,
		},
		{
			name:     "disallowed origin should fail",
			origin:   "https://malicious.com",
			expected: false,
		},
		{
			name:     "localhost should fail in production",
			origin:   "http://localhost:8080",
			expected: false,
		},
		{
			name:     "empty origin should fail",
			origin:   "",
			expected: false,
		},
		{
			name:     "case mismatch should fail",
			origin:   "https://GAME.EXAMPLE.COM",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tt.origin)

			result := upgrader.CheckOrigin(req)
			if result != tt.expected {
				t.Errorf("CheckOrigin(%q) = %v, want %v", tt.origin, result, tt.expected)
			}
		})
	}
}

// TestWebSocketOriginValidation_DevelopmentMode tests WebSocket origin validation
// in development mode which should allow all origins for convenience.
func TestWebSocketOriginValidation_DevelopmentMode(t *testing.T) {
	// Create a config for development mode
	cfg := &config.Config{
		EnableDevMode:  true,       // Development mode
		AllowedOrigins: []string{}, // Empty - should allow all in dev mode
		ServerPort:     8080,
		WebDir:         "./test_web",
	}

	server, cleanup, err := createTestServerWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer cleanup()

	upgrader := server.upgrader()

	tests := []struct {
		name   string
		origin string
	}{
		{
			name:   "localhost should be allowed in dev mode",
			origin: "http://localhost:8080",
		},
		{
			name:   "any domain should be allowed in dev mode",
			origin: "https://random.example.com",
		},
		{
			name:   "even suspicious origins should be allowed in dev mode",
			origin: "https://malicious.com",
		},
		{
			name:   "empty origin should be allowed in dev mode",
			origin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tt.origin)

			result := upgrader.CheckOrigin(req)
			if !result {
				t.Errorf("CheckOrigin(%q) = false, want true (dev mode should allow all)", tt.origin)
			}
		})
	}
}

// TestWebSocketOriginValidation_EnvironmentOverride tests that the configuration
// properly respects environment variable overrides for allowed origins.
func TestWebSocketOriginValidation_EnvironmentOverride(t *testing.T) {
	// Set environment variable for allowed origins
	originalValue := os.Getenv("ALLOWED_ORIGINS")
	defer os.Setenv("ALLOWED_ORIGINS", originalValue)

	os.Setenv("ALLOWED_ORIGINS", "https://env.example.com,https://another.example.com")

	// Load config which should pick up the environment variable
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Force production mode for strict validation
	cfg.EnableDevMode = false

	server, cleanup, err := createTestServerWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer cleanup()

	upgrader := server.upgrader()

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{
			name:     "origin from env var should be allowed",
			origin:   "https://env.example.com",
			expected: true,
		},
		{
			name:     "second origin from env var should be allowed",
			origin:   "https://another.example.com",
			expected: true,
		},
		{
			name:     "origin not in env var should be rejected",
			origin:   "https://notallowed.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tt.origin)

			result := upgrader.CheckOrigin(req)
			if result != tt.expected {
				t.Errorf("CheckOrigin(%q) = %v, want %v", tt.origin, result, tt.expected)
			}
		})
	}
}

// TestOrderHosts tests the orderHosts function which sorts hostnames by priority.
func TestOrderHosts(t *testing.T) {
	hosts := map[string]string{
		"192.168.1.1":   "192.168.1.1",
		"localhost":     "localhost",
		"game.test.com": "game.test.com",
		"api.test.com":  "api.test.com",
		"127.0.0.1":     "127.0.0.1",
	}

	result := orderHosts(hosts)

	// Expected order: custom hostnames first (alphabetically), then localhost, then IPs (alphabetically)
	expected := []string{"api.test.com", "game.test.com", "localhost", "127.0.0.1", "192.168.1.1"}

	if len(result) != len(expected) {
		t.Fatalf("orderHosts() returned %d hosts, want %d", len(result), len(expected))
	}

	for i, host := range result {
		if host != expected[i] {
			t.Errorf("orderHosts()[%d] = %q, want %q", i, host, expected[i])
		}
	}
}

// Helper function to create a test server with custom config
func createTestServerWithConfig(cfg *config.Config) (*RPCServer, func(), error) {
	// Create temporary directory for web files
	tempDir := "./test_web"
	os.MkdirAll(tempDir, 0755)

	server := &RPCServer{
		config: cfg,
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return server, cleanup, nil
}

// TestWebSocketOriginValidation_Integration tests the complete WebSocket upgrade
// process with origin validation in a more realistic scenario.
func TestWebSocketOriginValidation_Integration(t *testing.T) {
	cfg := &config.Config{
		EnableDevMode:  false,
		AllowedOrigins: []string{"https://trusted.example.com"},
		ServerPort:     8080,
		WebDir:         "./test_web",
	}

	server, cleanup, err := createTestServerWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer cleanup()

	// Test with allowed origin
	t.Run("allowed origin should allow upgrade", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Origin", "https://trusted.example.com")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
		req.Header.Set("Sec-WebSocket-Version", "13")

		upgrader := server.upgrader()
		allowed := upgrader.CheckOrigin(req)
		if !allowed {
			t.Error("Expected allowed origin to pass CheckOrigin")
		}
	})

	// Test with disallowed origin
	t.Run("disallowed origin should reject upgrade", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Origin", "https://malicious.example.com")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
		req.Header.Set("Sec-WebSocket-Version", "13")

		upgrader := server.upgrader()
		allowed := upgrader.CheckOrigin(req)
		if allowed {
			t.Error("Expected disallowed origin to fail CheckOrigin")
		}
	})
}
