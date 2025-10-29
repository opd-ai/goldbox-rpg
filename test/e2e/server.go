package e2e

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// TestServer manages a test server instance for E2E tests
type TestServer struct {
	cmd        *exec.Cmd
	port       int
	baseURL    string
	dataDir    string
	webDir     string
	logFile    *os.File
	log        *logrus.Logger
	cancelFunc context.CancelFunc
}

// NewTestServer creates a new test server instance
func NewTestServer() (*TestServer, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Find available port
	port, err := findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}

	// Create temporary directories
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("goldbox-e2e-%d", time.Now().UnixNano()))
	dataDir := filepath.Join(tmpDir, "data")
	webDir := filepath.Join(tmpDir, "web")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}
	if err := os.MkdirAll(webDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create web dir: %w", err)
	}

	// Create minimal web directory structure
	staticDir := filepath.Join(webDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create static dir: %w", err)
	}

	// Create a simple index.html for testing
	indexHTML := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><h1>GoldBox RPG Test Server</h1></body>
</html>`
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte(indexHTML), 0644); err != nil {
		return nil, fmt.Errorf("failed to create index.html: %w", err)
	}

	// Create log file
	logFile, err := os.Create(filepath.Join(tmpDir, "server.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &TestServer{
		port:    port,
		baseURL: fmt.Sprintf("http://localhost:%d", port),
		dataDir: dataDir,
		webDir:  webDir,
		logFile: logFile,
		log:     logger,
	}, nil
}

// Start starts the test server
func (ts *TestServer) Start() error {
	// Build the server binary if it doesn't exist
	serverBin := filepath.Join(".", "bin", "server")
	if _, err := os.Stat(serverBin); os.IsNotExist(err) {
		ts.log.Info("Building server binary...")
		buildCmd := exec.Command("make", "build")
		buildCmd.Stdout = ts.logFile
		buildCmd.Stderr = ts.logFile
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("failed to build server: %w", err)
		}
	}

	// Create context for server lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	ts.cancelFunc = cancel

	// Start server with test configuration
	ts.cmd = exec.CommandContext(ctx, serverBin)
	ts.cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOLDBOX_PORT=%d", ts.port),
		fmt.Sprintf("GOLDBOX_DATA_DIR=%s", ts.dataDir),
		fmt.Sprintf("GOLDBOX_WEB_DIR=%s", ts.webDir),
		"GOLDBOX_LOG_LEVEL=info",
		"GOLDBOX_AUTO_SAVE_INTERVAL=5s",
		"GOLDBOX_SESSION_TIMEOUT=30s",
		"GOLDBOX_DEV_MODE=true",
	)
	ts.cmd.Stdout = ts.logFile
	ts.cmd.Stderr = ts.logFile

	// Set process group ID so we can kill the entire process tree
	ts.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	ts.log.Infof("Starting test server on port %d", ts.port)
	ts.log.Infof("Data directory: %s", ts.dataDir)
	ts.log.Infof("Web directory: %s", ts.webDir)

	if err := ts.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for server to be ready
	client := NewClient(ts.baseURL)
	if err := client.WaitForHealth(30 * time.Second); err != nil {
		ts.Stop()
		return fmt.Errorf("server did not become healthy: %w", err)
	}

	ts.log.Info("Test server is ready")
	return nil
}

// Stop stops the test server and cleans up resources
func (ts *TestServer) Stop() error {
	ts.log.Info("Stopping test server...")

	if ts.cancelFunc != nil {
		ts.cancelFunc()
	}

	if ts.cmd != nil && ts.cmd.Process != nil {
		// Kill the process group to ensure all child processes are terminated
		pgid, err := syscall.Getpgid(ts.cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		}

		// Wait for process to exit with timeout
		done := make(chan error, 1)
		go func() {
			done <- ts.cmd.Wait()
		}()

		select {
		case <-done:
			ts.log.Info("Server stopped gracefully")
		case <-time.After(5 * time.Second):
			ts.log.Warn("Server did not stop gracefully, forcing kill")
			if pgid, err := syscall.Getpgid(ts.cmd.Process.Pid); err == nil {
				syscall.Kill(-pgid, syscall.SIGKILL)
			}
			ts.cmd.Process.Kill()
		}
	}

	if ts.logFile != nil {
		ts.logFile.Close()
	}

	// Clean up temporary directories
	if ts.dataDir != "" {
		tmpDir := filepath.Dir(ts.dataDir)
		os.RemoveAll(tmpDir)
	}

	return nil
}

// BaseURL returns the server's base URL
func (ts *TestServer) BaseURL() string {
	return ts.baseURL
}

// DataDir returns the server's data directory
func (ts *TestServer) DataDir() string {
	return ts.dataDir
}

// GetLogContents returns the contents of the server log
func (ts *TestServer) GetLogContents() (string, error) {
	if ts.logFile == nil {
		return "", fmt.Errorf("log file not available")
	}

	// Sync and seek to beginning
	ts.logFile.Sync()
	ts.logFile.Seek(0, 0)

	content, err := io.ReadAll(ts.logFile)
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %w", err)
	}

	return string(content), nil
}

// Restart restarts the test server
func (ts *TestServer) Restart() error {
	ts.log.Info("Restarting test server...")

	// Stop the server
	if err := ts.Stop(); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	// Wait a moment for cleanup
	time.Sleep(1 * time.Second)

	// Start the server again
	if err := ts.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// findAvailablePort finds an available TCP port
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
