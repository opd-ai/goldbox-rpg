package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestServerStartup is a diagnostic test to verify server can start
func TestServerStartup(t *testing.T) {
	server, err := NewTestServer()
	require.NoError(t, err, "should create test server")
	
	err = server.Start()
	if err != nil {
		// Print server logs for debugging
		logs, _ := server.GetLogContents()
		t.Logf("Server logs:\n%s", logs)
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Let server run for a bit
	time.Sleep(5 * time.Second)

	// Print logs
	logs, err := server.GetLogContents()
	require.NoError(t, err)
	t.Logf("Server logs after 5 seconds:\n%s", logs)

	// Try to check health
	client := NewClient(server.BaseURL())
	err = client.WaitForHealth(5 * time.Second)
	if err != nil {
		logs, _ := server.GetLogContents()
		t.Logf("Health check failed. Server logs:\n%s", logs)
		t.Fatalf("Health check failed: %v", err)
	}

	t.Log("Server started successfully and is healthy!")
}
