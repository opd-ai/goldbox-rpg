// Package cliutil provides common utilities for CLI tools, including a
// WebSocket-based preview server for live content editing.
package cliutil

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// PreviewServer manages WebSocket connections for live preview in CLI tools.
// It broadcasts JSON-encoded data to all connected clients whenever content changes.
type PreviewServer struct {
	mu         sync.RWMutex
	clients    map[*websocket.Conn]struct{}
	port       int
	previewFS  fs.FS
	previewDir string
}

// NewPreviewServer creates a new preview server on the specified port.
// previewFS should contain the embedded preview.html file.
// previewDir is the directory name within the FS (e.g., "." for root).
func NewPreviewServer(port int, previewFS fs.FS, previewDir string) *PreviewServer {
	return &PreviewServer{
		clients:    make(map[*websocket.Conn]struct{}),
		port:       port,
		previewFS:  previewFS,
		previewDir: previewDir,
	}
}

// AddClient registers a new WebSocket client.
func (ps *PreviewServer) AddClient(conn *websocket.Conn) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.clients[conn] = struct{}{}
}

// RemoveClient unregisters a WebSocket client.
func (ps *PreviewServer) RemoveClient(conn *websocket.Conn) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.clients, conn)
}

// Broadcast sends the given JSON data to all connected clients.
func (ps *PreviewServer) Broadcast(data []byte) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for conn := range ps.clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		conn.Write(ctx, websocket.MessageText, data)
		cancel()
	}
}

// Start begins serving the preview HTTP server.
// contentType is used in the startup message (e.g., "map", "quest").
func (ps *PreviewServer) Start(contentType string) error {
	mux := http.NewServeMux()

	// Serve the embedded preview HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := fs.ReadFile(ps.previewFS, "preview.html")
		if err != nil {
			http.Error(w, "Preview page not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})

	// WebSocket endpoint for live updates
	mux.HandleFunc("/ws", ps.handleWebSocket)

	addr := fmt.Sprintf(":%d", ps.port)
	fmt.Printf("Preview server starting at http://localhost%s\n", addr)
	fmt.Printf("Open this URL in a browser to see live %s updates\n", contentType)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Preview server error: %v\n", err)
		}
	}()

	return nil
}

// handleWebSocket handles WebSocket connection upgrades and message handling.
func (ps *PreviewServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "connection closed")

	ps.AddClient(conn)
	defer ps.RemoveClient(conn)

	// Keep connection open until client disconnects
	for {
		_, _, err := conn.Read(context.Background())
		if err != nil {
			return
		}
	}
}

// Port returns the port number the server is configured to use.
func (ps *PreviewServer) Port() int {
	return ps.port
}
