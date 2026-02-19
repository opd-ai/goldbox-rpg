// Package main implements the GoldBox RPG Engine server application.
//
// This is the main entry point for the GoldBox RPG Engine, a modern Go-based
// framework for creating turn-based RPG games inspired by the classic SSI
// Gold Box series. The server provides comprehensive character management,
// combat systems, and world interactions through a JSON-RPC API with WebSocket
// support for real-time communication.
//
// # Architecture
//
// The server application follows a clean separation of concerns:
//
//   - Configuration loading and validation (via pkg/config)
//   - Logging setup and initialization
//   - Zero-configuration bootstrap for new installations (via pkg/pcg)
//   - Server lifecycle management with graceful shutdown
//   - Signal handling for SIGINT and SIGTERM
//
// # Startup Sequence
//
// 1. Load configuration from environment variables with secure defaults
// 2. Configure logging based on LOG_LEVEL setting
// 3. Detect existing game configuration or bootstrap a new game
// 4. Initialize the RPC server with WebSocket support
// 5. Start listening for connections
// 6. Handle shutdown signals gracefully with state persistence
//
// # Environment Variables
//
// The server supports the following environment variables:
//
//   - SERVER_PORT: HTTP server port (default: 8080)
//   - WEB_DIR: Static web file directory (default: ./web)
//   - SESSION_TIMEOUT: Session expiration time (default: 30m)
//   - LOG_LEVEL: Logging verbosity (debug, info, warn, error; default: info)
//   - ENABLE_DEV_MODE: Development mode flag (default: true)
//   - ENABLE_PERSISTENCE: Auto-save game state (default: true)
//   - DATA_DIR: Persistence directory (default: ./data)
//
// # Usage
//
// Run the server with default settings:
//
//	./server
//
// Run with custom port and debug logging:
//
//	SERVER_PORT=9000 LOG_LEVEL=debug ./server
//
// # Graceful Shutdown
//
// The server handles SIGINT (Ctrl+C) and SIGTERM signals gracefully:
//
// 1. Stop accepting new connections
// 2. Save game state if persistence is enabled
// 3. Close all active connections
// 4. Exit cleanly
//
// The shutdown process has a 30-second timeout before forcing exit.
package main
