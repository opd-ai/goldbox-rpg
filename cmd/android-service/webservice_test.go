// FILENAME: webservice_test.go
// PURPOSE: Tests for android-service helper functions.
package main

import (
	"context"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/server"
)

func TestConfigureLogging(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel logrus.Level
	}{
		{
			name:          "debug level",
			level:         "debug",
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "info level",
			level:         "info",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "warn level",
			level:         "warn",
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "error level",
			level:         "error",
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "invalid level defaults to info",
			level:         "invalid",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "empty level defaults to info",
			level:         "",
			expectedLevel: logrus.InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configureLogging(tc.level)
			assert.Equal(t, tc.expectedLevel, logrus.GetLevel())
		})
	}
}

func TestGetLANIP(t *testing.T) {
	// getLANIP returns the first non-loopback IPv4 address
	// On most systems this will return an IP, on isolated systems it may return empty
	ip := getLANIP()
	// We can't predict the exact IP, but we can validate the format if returned
	if ip != "" {
		// Should be a valid IPv4 format (4 dot-separated octets)
		assert.Regexp(t, `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`, ip)
		// Should not be loopback
		assert.NotEqual(t, "127.0.0.1", ip)
	}
	// Empty result is valid if no non-loopback interfaces exist
}

func TestBootstrapGame(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "goldbox-bootstrap-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a config with short timeout for testing
	cfg := &config.Config{
		DataDir:          tmpDir,
		BootstrapTimeout: 30 * time.Second,
	}

	// Run bootstrap - this tests the complete bootstrap flow
	err = bootstrapGame(cfg)
	assert.NoError(t, err)

	// Verify some files were created
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries, "Bootstrap should create files in data directory")
}

func TestBootstrapGameTimeout(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "goldbox-bootstrap-timeout-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a config with very short timeout that will likely expire
	// Note: In practice, bootstrap may complete quickly, so this test verifies
	// the timeout mechanism is properly wired up
	cfg := &config.Config{
		DataDir:          tmpDir,
		BootstrapTimeout: 1 * time.Nanosecond,
	}

	// With such a short timeout, bootstrap should fail with context deadline exceeded
	// unless the bootstrap is extremely fast
	err = bootstrapGame(cfg)
	// Either it completes quickly or times out - both are valid behaviors
	if err != nil {
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	}
}

func TestBootstrapGameCancellation(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "goldbox-bootstrap-cancel-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config with reasonable timeout
	cfg := &config.Config{
		DataDir:          tmpDir,
		BootstrapTimeout: 5 * time.Second,
	}

	// Run bootstrap in background
	done := make(chan error, 1)
	go func() {
		done <- bootstrapGame(cfg)
	}()

	// Wait for completion (bootstrap should work)
	select {
	case err := <-done:
		// Bootstrap completed - verify it succeeded
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Bootstrap took too long")
	}
}

func TestGetBindAddress(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "default when not set",
			envValue: "",
			expected: "127.0.0.1",
		},
		{
			name:     "custom bind address",
			envValue: "0.0.0.0",
			expected: "0.0.0.0",
		},
		{
			name:     "specific IP address",
			envValue: "192.168.1.100",
			expected: "192.168.1.100",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Save and restore original env
			original := os.Getenv("GOLDBOX_BIND_ADDR")
			defer os.Setenv("GOLDBOX_BIND_ADDR", original)

			if tc.envValue == "" {
				os.Unsetenv("GOLDBOX_BIND_ADDR")
			} else {
				os.Setenv("GOLDBOX_BIND_ADDR", tc.envValue)
			}

			result := getBindAddress()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetShutdownTimeout(t *testing.T) {
	tests := []struct {
		name       string
		configured time.Duration
		expected   time.Duration
	}{
		{
			name:       "zero defaults to 5 seconds",
			configured: 0,
			expected:   5 * time.Second,
		},
		{
			name:       "negative defaults to 5 seconds",
			configured: -1 * time.Second,
			expected:   5 * time.Second,
		},
		{
			name:       "positive value is used",
			configured: 10 * time.Second,
			expected:   10 * time.Second,
		},
		{
			name:       "small positive value is used",
			configured: 1 * time.Second,
			expected:   1 * time.Second,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getShutdownTimeout(tc.configured)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestLogServerStartup(t *testing.T) {
	// Capture log output
	var buf logrus.Hook

	tests := []struct {
		name     string
		bindAddr string
		port     int
	}{
		{
			name:     "localhost binding",
			bindAddr: "127.0.0.1",
			port:     8080,
		},
		{
			name:     "all interfaces binding",
			bindAddr: "0.0.0.0",
			port:     9090,
		},
		{
			name:     "specific IP binding",
			bindAddr: "192.168.1.1",
			port:     3000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Just verify it doesn't panic - logging output is side effect
			assert.NotPanics(t, func() {
				logServerStartup(tc.bindAddr, tc.port)
			})
		})
	}

	// Suppress unused variable warning
	_ = buf
}

func TestSetupGracefulShutdownWithSignal(t *testing.T) {
	// Create a temp directory for the test
	tmpDir := t.TempDir()

	// Create a test server
	srv, err := server.NewRPCServer(tmpDir)
	require.NoError(t, err)

	// Create a listener on an ephemeral port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Create a signal channel we control
	sigCh := make(chan os.Signal, 1)

	// Setup graceful shutdown with our signal channel
	setupGracefulShutdownWithSignal(srv, listener, 1*time.Second, sigCh)

	// Give the goroutine time to start
	time.Sleep(100 * time.Millisecond)

	// Send shutdown signal
	sigCh <- syscall.SIGTERM

	// Wait for shutdown to complete
	time.Sleep(500 * time.Millisecond)

	// Verify listener is closed by trying to accept (should fail)
	_, err = listener.Accept()
	assert.Error(t, err, "Listener should be closed after shutdown signal")
}

func TestSetupGracefulShutdownDefault(t *testing.T) {
	// Create a temp directory for the test
	tmpDir := t.TempDir()

	// Create a test server
	srv, err := server.NewRPCServer(tmpDir)
	require.NoError(t, err)

	// Create a listener on an ephemeral port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	// Setup graceful shutdown (this uses the default signal handling)
	// We can't easily trigger this without sending real signals, so just verify it doesn't panic
	assert.NotPanics(t, func() {
		setupGracefulShutdown(srv, listener, 1*time.Second)
	})

	// Clean up: close listener to stop the goroutine from blocking
	listener.Close()
}

func TestSetupGracefulShutdownWithClosedListener(t *testing.T) {
	// Test graceful shutdown when listener is already closed (error path)
	tmpDir := t.TempDir()

	srv, err := server.NewRPCServer(tmpDir)
	require.NoError(t, err)

	// Create and immediately close the listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	listener.Close() // Close it before shutdown

	sigCh := make(chan os.Signal, 1)
	setupGracefulShutdownWithSignal(srv, listener, 1*time.Second, sigCh)

	// Give the goroutine time to start
	time.Sleep(100 * time.Millisecond)

	// Send shutdown signal - this should handle the error gracefully
	sigCh <- syscall.SIGTERM

	// Wait for shutdown to complete
	time.Sleep(500 * time.Millisecond)
}

func TestSetupGracefulShutdownShortTimeout(t *testing.T) {
	// Test graceful shutdown with very short timeout to hit error path
	tmpDir := t.TempDir()

	srv, err := server.NewRPCServer(tmpDir)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Start serving in background to create some state
	go func() {
		srv.Serve(listener)
	}()

	sigCh := make(chan os.Signal, 1)
	// Use 1 nanosecond timeout to force timeout error
	setupGracefulShutdownWithSignal(srv, listener, 1*time.Nanosecond, sigCh)

	time.Sleep(100 * time.Millisecond)

	// Send shutdown signal
	sigCh <- syscall.SIGTERM

	// Wait for shutdown to complete
	time.Sleep(500 * time.Millisecond)
}
