// Package cliutil provides common utilities for CLI tools in the GoldBox RPG Engine.
//
// This package contains shared functionality used across multiple CLI applications
// including map-editor, quest-builder, and content-creator tools.
//
// # Preview Server
//
// PreviewServer enables live content editing with WebSocket-based real-time updates:
//
//	server := cliutil.NewPreviewServer(9001, previewFS, ".")
//	server.Start("map")
//
//	// Broadcast updates to all connected browsers
//	data, _ := json.Marshal(mapData)
//	server.Broadcast(data)
//
// The preview server:
//   - Serves embedded HTML preview pages
//   - Manages WebSocket connections for live updates
//   - Broadcasts JSON-encoded content changes to all clients
//   - Supports multiple concurrent preview connections
//
// # Thread Safety
//
// All PreviewServer methods are safe for concurrent use. Client connections
// are managed with proper mutex protection for reliable multi-client operation.
package cliutil
