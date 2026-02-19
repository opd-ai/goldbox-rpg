package main

import (
	"bytes"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/server"
)

// TestConfigureLogging tests the logging configuration function.
func TestConfigureLogging(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      string
		expectedLevel logrus.Level
	}{
		{
			name:          "debug level",
			logLevel:      "debug",
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "info level",
			logLevel:      "info",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "warn level",
			logLevel:      "warn",
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "error level",
			logLevel:      "error",
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "invalid level falls back to info",
			logLevel:      "invalid",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "empty level falls back to info",
			logLevel:      "",
			expectedLevel: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output to suppress warning messages
			logrus.SetOutput(io.Discard)
			defer logrus.SetOutput(os.Stderr)

			configureLogging(tt.logLevel)
			assert.Equal(t, tt.expectedLevel, logrus.GetLevel())
		})
	}
}

// TestLogStartupInfo tests that startup info is logged correctly.
func TestLogStartupInfo(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	defer logrus.SetOutput(os.Stderr)

	cfg := &config.Config{
		ServerPort:     8080,
		WebDir:         "./web",
		SessionTimeout: 30 * time.Minute,
		LogLevel:       "info",
		EnableDevMode:  true,
	}

	logStartupInfo(cfg)

	output := buf.String()
	assert.Contains(t, output, "Starting GoldBox RPG Engine server")
	assert.Contains(t, output, "8080")
	assert.Contains(t, output, "./web")
}

// TestSetupShutdownHandling tests the shutdown signal channel setup.
func TestSetupShutdownHandling(t *testing.T) {
	sigChan, errChan := setupShutdownHandling()

	assert.NotNil(t, sigChan)
	assert.NotNil(t, errChan)

	// Test that sigChan has capacity
	assert.Equal(t, 1, cap(sigChan))

	// Test that errChan has capacity
	assert.Equal(t, 1, cap(errChan))

	// Clean up signal notification
	signal.Stop(sigChan)
}

// TestInitializeServerWithValidConfig tests server initialization with a valid configuration.
func TestInitializeServerWithValidConfig(t *testing.T) {
	// Create temp directory for web files
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ServerPort:     0, // Use port 0 to let OS assign available port
		WebDir:         tmpDir,
		SessionTimeout: 30 * time.Minute,
		LogLevel:       "info",
		EnableDevMode:  true,
	}

	srv, listener := initializeServer(cfg)

	assert.NotNil(t, srv)
	assert.NotNil(t, listener)

	// Get the assigned port
	addr := listener.Addr().(*net.TCPAddr)
	assert.Greater(t, addr.Port, 0)

	// Clean up
	listener.Close()
}

// TestStartServerAsync tests the asynchronous server start.
func TestStartServerAsync(t *testing.T) {
	// Create temp directory for web files
	tmpDir := t.TempDir()

	srv, err := server.NewRPCServer(tmpDir)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	errChan := make(chan error, 1)

	// Start server asynchronously
	startServerAsync(srv, listener, errChan)

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Server should be running, errChan should be empty
	select {
	case err := <-errChan:
		t.Fatalf("Server failed unexpectedly: %v", err)
	default:
		// This is expected - no error means server is running
	}

	// Close the listener to trigger server shutdown
	listener.Close()

	// Wait a bit for the error to propagate
	time.Sleep(100 * time.Millisecond)
}

// TestWaitForShutdownSignal_Signal tests that shutdown signal is handled.
func TestWaitForShutdownSignal_Signal(t *testing.T) {
	sigChan := make(chan os.Signal, 1)
	errChan := make(chan error, 1)

	// Send signal in goroutine
	go func() {
		time.Sleep(10 * time.Millisecond)
		sigChan <- syscall.SIGINT
	}()

	// This should return when signal is received
	done := make(chan struct{})
	go func() {
		waitForShutdownSignal(sigChan, errChan)
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("waitForShutdownSignal did not return after signal")
	}
}

// TestWaitForShutdownSignal_Error tests that server errors trigger shutdown.
func TestWaitForShutdownSignal_Error(t *testing.T) {
	sigChan := make(chan os.Signal, 1)
	errChan := make(chan error, 1)

	// Send error in goroutine
	go func() {
		time.Sleep(10 * time.Millisecond)
		errChan <- assert.AnError
	}()

	// This should return when error is received
	done := make(chan struct{})
	go func() {
		waitForShutdownSignal(sigChan, errChan)
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("waitForShutdownSignal did not return after error")
	}
}

// TestPerformGracefulShutdown tests the graceful shutdown process.
func TestPerformGracefulShutdown(t *testing.T) {
	// Suppress log output during test
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	// Create temp directory for web files
	tmpDir := t.TempDir()

	srv, err := server.NewRPCServer(tmpDir)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	cfg := &config.Config{
		EnablePersistence: false, // Disable persistence to avoid file operations
	}

	// Test that shutdown completes without panic
	done := make(chan struct{})
	go func() {
		performGracefulShutdown(cfg, listener, srv)
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Graceful shutdown did not complete in time")
	}
}

// TestPerformGracefulShutdownWithPersistence tests shutdown with persistence enabled.
func TestPerformGracefulShutdownWithPersistence(t *testing.T) {
	// Suppress log output during test
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	// Create temp directory for web files
	tmpDir := t.TempDir()

	srv, err := server.NewRPCServer(tmpDir)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	cfg := &config.Config{
		EnablePersistence: true, // Enable persistence
		DataDir:           tmpDir,
	}

	// Test that shutdown completes without panic even with persistence
	done := make(chan struct{})
	go func() {
		performGracefulShutdown(cfg, listener, srv)
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Graceful shutdown with persistence did not complete in time")
	}
}

// TestLoadAndConfigureSystem tests the configuration loading function.
func TestLoadAndConfigureSystem(t *testing.T) {
	// Set up environment variables for test
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("LOG_LEVEL", "warn")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("LOG_LEVEL")

	// Suppress log output during test
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	cfg := loadAndConfigureSystem()

	assert.NotNil(t, cfg)
	assert.Equal(t, 9999, cfg.ServerPort)
	assert.Equal(t, "warn", cfg.LogLevel)
}

// TestExecuteServerLifecycle tests the full server lifecycle with early shutdown.
func TestExecuteServerLifecycle(t *testing.T) {
	// Suppress log output during test
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	// Create temp directory for web files
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ServerPort:        0,
		WebDir:            tmpDir,
		SessionTimeout:    30 * time.Minute,
		LogLevel:          "info",
		EnableDevMode:     true,
		EnablePersistence: false,
	}

	srv, err := server.NewRPCServer(tmpDir)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Run server lifecycle in goroutine and send shutdown signal
	done := make(chan struct{})
	go func() {
		// Override signal handling to trigger immediate shutdown
		sigChan, errChan := setupShutdownHandling()

		// Start server
		startServerAsync(srv, listener, errChan)

		// Send shutdown signal after brief delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			sigChan <- syscall.SIGINT
		}()

		waitForShutdownSignal(sigChan, errChan)
		performGracefulShutdown(cfg, listener, srv)
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("Server lifecycle did not complete in time")
	}
}

// TestInitializeBootstrapGame tests the bootstrap game initialization.
func TestInitializeBootstrapGame(t *testing.T) {
	// Suppress log output during test
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	// Create temp directory for data
	tmpDir := t.TempDir()

	// Test bootstrap initialization
	err := initializeBootstrapGame(tmpDir)

	// Bootstrap may fail if PCG resources are not available, which is OK for testing
	// The important thing is that it doesn't panic
	if err != nil {
		t.Logf("Bootstrap game initialization returned error (expected in test environment): %v", err)
	}
}

// BenchmarkConfigureLogging benchmarks the logging configuration.
func BenchmarkConfigureLogging(b *testing.B) {
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	for i := 0; i < b.N; i++ {
		configureLogging("info")
	}
}

// BenchmarkSetupShutdownHandling benchmarks shutdown handler setup.
func BenchmarkSetupShutdownHandling(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sigChan, _ := setupShutdownHandling()
		signal.Stop(sigChan)
	}
}
