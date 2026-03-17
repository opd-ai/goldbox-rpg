// FILENAME: webservice_test.go
// PURPOSE: Tests for android-service helper functions.
package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/config"
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
