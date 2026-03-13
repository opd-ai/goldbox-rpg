package cliutil

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed preview.go
var testFS embed.FS

func TestNewPreviewServer(t *testing.T) {
	ps := NewPreviewServer(9001, testFS, ".")
	assert.NotNil(t, ps)
	assert.Equal(t, 9001, ps.Port())
	assert.NotNil(t, ps.clients)
}

func TestPreviewServerClientManagement(t *testing.T) {
	ps := NewPreviewServer(9002, testFS, ".")

	// Verify no clients initially
	ps.mu.RLock()
	assert.Equal(t, 0, len(ps.clients))
	ps.mu.RUnlock()

	// AddClient is tested through the WebSocket handler in integration tests
}

func TestPreviewServerBroadcast(t *testing.T) {
	ps := NewPreviewServer(9003, testFS, ".")

	// Broadcast with no clients should not panic
	ps.Broadcast([]byte(`{"test": true}`))
}
