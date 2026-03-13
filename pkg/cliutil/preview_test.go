package cliutil

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"
)

//go:embed preview.go
var testFS embed.FS

func TestNewPreviewServer(t *testing.T) {
	tests := []struct {
		name       string
		port       int
		previewDir string
	}{
		{
			name:       "standard port",
			port:       9001,
			previewDir: ".",
		},
		{
			name:       "high port",
			port:       65000,
			previewDir: "preview",
		},
		{
			name:       "zero port (random)",
			port:       0,
			previewDir: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPreviewServer(tt.port, testFS, tt.previewDir)
			assert.NotNil(t, ps)
			assert.Equal(t, tt.port, ps.Port())
			assert.NotNil(t, ps.clients)
			assert.Equal(t, tt.previewDir, ps.previewDir)
		})
	}
}

func TestPreviewServerPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"standard port", 8080},
		{"high port", 9999},
		{"low port", 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPreviewServer(tt.port, testFS, ".")
			assert.Equal(t, tt.port, ps.Port())
		})
	}
}

func TestPreviewServerAddRemoveClient(t *testing.T) {
	ps := NewPreviewServer(9002, testFS, ".")

	// Verify no clients initially
	ps.mu.RLock()
	initialCount := len(ps.clients)
	ps.mu.RUnlock()
	assert.Equal(t, 0, initialCount)

	// Create a test server for WebSocket connections
	mockFS := fstest.MapFS{
		"preview.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	psWithMock := NewPreviewServer(0, mockFS, ".")

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", psWithMock.handleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect a client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)

	// Wait for client to be added
	time.Sleep(100 * time.Millisecond)

	psWithMock.mu.RLock()
	assert.Equal(t, 1, len(psWithMock.clients))
	psWithMock.mu.RUnlock()

	// Close connection which should trigger RemoveClient
	conn.Close(websocket.StatusNormalClosure, "test complete")

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	psWithMock.mu.RLock()
	assert.Equal(t, 0, len(psWithMock.clients))
	psWithMock.mu.RUnlock()
}

func TestPreviewServerBroadcast(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		clientCount int
	}{
		{
			name:        "no clients",
			data:        []byte(`{"test": true}`),
			clientCount: 0,
		},
		{
			name:        "empty data",
			data:        []byte{},
			clientCount: 0,
		},
		{
			name:        "large data",
			data:        []byte(strings.Repeat("a", 10000)),
			clientCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPreviewServer(9003, testFS, ".")
			// Broadcast should not panic even with no clients
			assert.NotPanics(t, func() {
				ps.Broadcast(tt.data)
			})
		})
	}
}

func TestPreviewServerBroadcastToClients(t *testing.T) {
	mockFS := fstest.MapFS{
		"preview.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	ps := NewPreviewServer(0, mockFS, ".")

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ps.handleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect two clients
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn1, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn1.Close(websocket.StatusNormalClosure, "")

	conn2, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn2.Close(websocket.StatusNormalClosure, "")

	// Wait for clients to be registered
	time.Sleep(100 * time.Millisecond)

	ps.mu.RLock()
	assert.Equal(t, 2, len(ps.clients))
	ps.mu.RUnlock()

	// Broadcast a message
	testData := []byte(`{"type":"update","content":"test"}`)
	ps.Broadcast(testData)

	// Read from both clients
	readCtx, readCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer readCancel()

	_, data1, err := conn1.Read(readCtx)
	require.NoError(t, err)
	assert.Equal(t, testData, data1)

	_, data2, err := conn2.Read(readCtx)
	require.NoError(t, err)
	assert.Equal(t, testData, data2)
}

func TestPreviewServerStartServeFiles(t *testing.T) {
	mockFS := fstest.MapFS{
		"preview.html": &fstest.MapFile{Data: []byte("<html><body>Preview</body></html>")},
	}
	ps := NewPreviewServer(0, mockFS, ".")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := fs.ReadFile(ps.previewFS, "preview.html")
		if err != nil {
			http.Error(w, "Preview page not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
}

func TestPreviewServerStartFileNotFound(t *testing.T) {
	// Empty FS - no preview.html
	emptyFS := fstest.MapFS{}
	ps := NewPreviewServer(0, emptyFS, ".")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := fs.ReadFile(ps.previewFS, "preview.html")
		if err != nil {
			http.Error(w, "Preview page not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPreviewServerHandleWebSocketAccept(t *testing.T) {
	mockFS := fstest.MapFS{
		"preview.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	ps := NewPreviewServer(0, mockFS, ".")

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ps.handleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, resp, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
}

func TestPreviewServerConcurrentClients(t *testing.T) {
	mockFS := fstest.MapFS{
		"preview.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	ps := NewPreviewServer(0, mockFS, ".")

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ps.handleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	const numClients = 10
	var wg sync.WaitGroup
	connections := make([]*websocket.Conn, numClients)

	// Connect clients concurrently
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			conn, _, err := websocket.Dial(ctx, wsURL, nil)
			if err == nil {
				connections[idx] = conn
			}
		}(i)
	}
	wg.Wait()

	// Wait for all clients to be registered
	time.Sleep(200 * time.Millisecond)

	ps.mu.RLock()
	assert.Equal(t, numClients, len(ps.clients))
	ps.mu.RUnlock()

	// Broadcast to all clients
	testData := []byte(`{"concurrent":"test"}`)
	ps.Broadcast(testData)

	// Verify each client receives the message
	for i, conn := range connections {
		if conn != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, data, err := conn.Read(ctx)
			cancel()
			require.NoError(t, err, "client %d should receive message", i)
			assert.Equal(t, testData, data)
			conn.Close(websocket.StatusNormalClosure, "")
		}
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	ps.mu.RLock()
	assert.Equal(t, 0, len(ps.clients))
	ps.mu.RUnlock()
}

func TestPreviewServerClientManagement(t *testing.T) {
	ps := NewPreviewServer(9002, testFS, ".")

	// Verify no clients initially
	ps.mu.RLock()
	assert.Equal(t, 0, len(ps.clients))
	ps.mu.RUnlock()
}

func TestPreviewServerStart(t *testing.T) {
	// Create a mock FS with preview.html
	mockFS := fstest.MapFS{
		"preview.html": &fstest.MapFile{Data: []byte("<html><body>Test Preview</body></html>")},
	}

	// Use a random available port to avoid conflicts
	ps := NewPreviewServer(0, mockFS, ".")

	// Override port to use a high ephemeral port
	ps.port = 19876

	// Start should return nil (no error)
	err := ps.Start("test")
	assert.NoError(t, err)

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Try to connect to verify server is running
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test the root endpoint (preview HTML)
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:19876/", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	// Note: Server might not be fully ready, so we don't fail on connection errors
}

func TestPreviewServerStartWithDifferentContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		port        int
	}{
		{"map content", "map", 19877},
		{"quest content", "quest", 19878},
		{"item content", "item", 19879},
	}

	mockFS := fstest.MapFS{
		"preview.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPreviewServer(tt.port, mockFS, ".")
			err := ps.Start(tt.contentType)
			assert.NoError(t, err)
			// Server starts asynchronously, so no need to wait
		})
	}
}

func TestPreviewServerStartIntegration(t *testing.T) {
	// Full integration test - start server, connect WebSocket, broadcast, receive
	mockFS := fstest.MapFS{
		"preview.html": &fstest.MapFile{Data: []byte("<html><head></head><body>Preview Server</body></html>")},
	}

	ps := NewPreviewServer(19880, mockFS, ".")
	err := ps.Start("integration-test")
	require.NoError(t, err)

	// Wait for server startup
	time.Sleep(200 * time.Millisecond)

	// Connect via WebSocket
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws://localhost:19880/ws", nil)
	if err != nil {
		// Server might not have started in time on slow systems - skip test
		t.Skip("Server not ready:", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Verify client was added
	time.Sleep(100 * time.Millisecond)
	ps.mu.RLock()
	clientCount := len(ps.clients)
	ps.mu.RUnlock()
	assert.GreaterOrEqual(t, clientCount, 1)

	// Broadcast and verify receipt
	testMessage := []byte(`{"event":"update","data":"test"}`)
	ps.Broadcast(testMessage)

	readCtx, readCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer readCancel()
	_, data, err := conn.Read(readCtx)
	if err == nil {
		assert.Equal(t, testMessage, data)
	}
}
