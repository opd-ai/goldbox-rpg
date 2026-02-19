// Package server implements a JSON-RPC 2.0 game server for the GoldBox RPG engine.
//
// This package provides complete backend infrastructure for turn-based RPG gameplay
// including session management, real-time WebSocket communication, combat handling,
// spell casting, and comprehensive operational monitoring.
//
// # Server Architecture
//
// RPCServer is the main server instance that coordinates:
//
//   - Game state management (WorldState, TurnManager, TimeManager)
//
//   - Player session tracking with automatic cleanup
//
//   - WebSocket broadcasting for real-time updates
//
//   - Request validation, rate limiting, and metrics collection
//
//   - Procedural content generation integration
//
//     cfg, _ := config.Load()
//     srv := server.NewRPCServer(cfg)
//     srv.Start()
//
// # Session Management
//
// PlayerSession represents an active player connection with session ID,
// player reference, activity tracking, and WebSocket connection handling.
// Sessions are automatically cleaned up after configurable timeout periods.
//
// # JSON-RPC Methods
//
// The server handles standard RPG operations via JSON-RPC 2.0:
//   - Player/Character management: createPlayer, getPlayer, createCharacter
//   - Movement and positioning: move, getPosition
//   - Combat actions: attack, castSpell, getSpells
//   - Equipment: equipItem, unequipItem, getInventory
//   - World state: getWorld, getWorldState
//
// # Real-time Communication
//
// WebSocket connections enable bi-directional communication for:
//   - Combat event broadcasting
//   - Turn notifications
//   - State synchronization across multiple clients
//
// # Operational Features
//
//   - Health checks at /health, /ready, /live endpoints
//   - Prometheus metrics at /metrics
//   - Request rate limiting with configurable thresholds
//   - Pprof profiling when enabled
//   - File-based auto-save with configurable intervals
//
// # Thread Safety
//
// All server operations are mutex-protected for safe concurrent access.
// Session cleanup and state updates use proper locking patterns.
package server
