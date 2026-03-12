// Package server implements a JSON-RPC 2.0 server with WebSocket support for the GoldBox RPG Engine,
// providing real-time multiplayer game sessions, character management, combat coordination, and
// procedural content generation.
//
// This package provides complete backend infrastructure for turn-based RPG gameplay
// including session management, real-time WebSocket communication, combat handling,
// spell casting, and comprehensive operational monitoring.
//
// # Server Architecture
//
// RPCServer is the main server type that coordinates all game operations through a handler pattern:
//
//   - JSON-RPC 2.0 handlers: handleMove, handleAttack, handleCastSpell, handleEquipItem, handleJoinGame, etc.
//   - Session management: PlayerSession tracks active connections with automatic timeout cleanup
//   - WebSocket integration: Real-time event broadcasting for combat updates, turn notifications, and state sync
//   - Health endpoints: /health (comprehensive status), /ready (readiness probe), /live (liveness probe)
//   - Prometheus metrics: /metrics endpoint exposing combat, quest, session, PCG, and performance metrics
//   - Game state management: WorldState, TurnManager, TimeManager coordination
//   - Procedural content generation: Integrated PCG manager for dynamic content creation
//
// # Security
//
// The server implements multiple layers of security protection:
//
//   - Input validation: All JSON-RPC requests validated via pkg/validation framework to prevent injection attacks
//   - Rate limiting: Token bucket rate limiter via golang.org/x/time protects API endpoints from abuse
//   - WebSocket origin validation: WEBSOCKET_ALLOWED_ORIGINS environment variable enforces allowed client domains
//   - Request size limiting: Maximum request body size enforced (default 1MB) to prevent DoS attacks
//   - Session timeout: Configurable session expiration (default 30m via GOLDBOX_SESSION_TIMEOUT)
//   - Concurrent access protection: All shared state protected with sync.RWMutex for thread safety
//
// Production deployments MUST configure WEBSOCKET_ALLOWED_ORIGINS to restrict WebSocket connections
// to legitimate client domains. Development mode allows all origins but should never be used in production.
//
// # Configuration
//
// Server configuration is loaded from environment variables with the GOLDBOX_ prefix:
//
// Core Server Settings:
//   - GOLDBOX_PORT: HTTP server port (default: 8080)
//   - GOLDBOX_LOG_LEVEL: Logging verbosity - debug, info, warn, error (default: info)
//   - GOLDBOX_WEB_DIR: Static web assets directory path (default: "web/static")
//   - GOLDBOX_DATA_DIR: Game data storage directory (default: "data")
//   - GOLDBOX_DEV_MODE: Enable development mode with relaxed security (default: false)
//
// Session & Request Settings:
//   - GOLDBOX_SESSION_TIMEOUT: Player session inactivity timeout (default: 30m)
//   - GOLDBOX_REQUEST_TIMEOUT: HTTP request processing timeout (default: 30s)
//   - GOLDBOX_MAX_REQUEST_SIZE: Maximum request body size in bytes (default: 1048576 = 1MB)
//
// WebSocket Security (REQUIRED FOR PRODUCTION):
//   - WEBSOCKET_ALLOWED_ORIGINS: Comma-separated list of allowed WebSocket origins (e.g., "https://example.com,https://www.example.com")
//   - ALLOWED_ORIGINS: Alternative name for WEBSOCKET_ALLOWED_ORIGINS (deprecated)
//
// Rate Limiting:
//   - GOLDBOX_RATE_LIMIT_ENABLED: Enable request rate limiting (default: true)
//   - GOLDBOX_RATE_LIMIT_RPS: Requests per second limit (default: 10.0)
//   - GOLDBOX_RATE_LIMIT_BURST: Burst capacity allowance (default: 20)
//   - GOLDBOX_RATE_LIMIT_CLEANUP_INTERVAL: Rate limiter cleanup interval (default: 5m)
//
// Retry & Resilience:
//   - GOLDBOX_RETRY_ENABLED: Enable retry mechanisms for transient failures (default: true)
//   - GOLDBOX_RETRY_MAX_ATTEMPTS: Maximum retry attempts (default: 3)
//   - GOLDBOX_RETRY_INITIAL_DELAY: Initial retry delay (default: 100ms)
//   - GOLDBOX_RETRY_MAX_DELAY: Maximum retry delay (default: 10s)
//   - GOLDBOX_RETRY_BACKOFF_MULTIPLIER: Exponential backoff multiplier (default: 2.0)
//
// Observability:
//   - GOLDBOX_METRICS_INTERVAL: Metrics collection interval (default: 30s)
//   - GOLDBOX_ENABLE_PROFILING: Enable pprof profiling endpoints (default: false)
//   - GOLDBOX_PROFILING_PORT: Pprof HTTP port when profiling enabled (default: 0 = disabled)
//   - GOLDBOX_ALERTING_ENABLED: Enable alerting system (default: true)
//   - GOLDBOX_ALERTING_INTERVAL: Alerting evaluation interval (default: 1m)
//
// Persistence:
//   - GOLDBOX_ENABLE_PERSISTENCE: Enable game state persistence to disk (default: false)
//   - GOLDBOX_ENABLE_SESSION_PERSISTENCE: Enable session state persistence (default: false)
//   - GOLDBOX_AUTO_SAVE_INTERVAL: Automatic game state save frequency (default: 5m)
//
// # Session Management
//
// PlayerSession represents an active player connection with session ID,
// player reference, activity tracking, and WebSocket connection handling.
// Sessions are automatically cleaned up after configurable timeout periods
// (GOLDBOX_SESSION_TIMEOUT environment variable, default 30 minutes).
//
// # JSON-RPC API Reference
//
// The server implements a comprehensive JSON-RPC 2.0 API with 40+ methods organized into categories.
// For complete API documentation with request/response schemas and examples, see pkg/README-RPC.md.
//
// Core Game Methods:
//   - joinGame, leaveGame, getGameState, saveGame
//
// Character Management:
//   - createCharacter, getCharacter, updateCharacter, deleteCharacter
//   - levelUp, gainExperience, modifyHealth
//
// Movement & Positioning:
//   - move, getPosition, teleport
//
// Combat Actions:
//   - startCombat, endCombat, attack, defend
//   - castSpell, useItem, endTurn
//
// Spell System:
//   - getSpell, getAllSpells, searchSpells, getSpellsBySchool, getSpellsByLevel
//
// Equipment & Inventory:
//   - equipItem, unequipItem, getInventory, addItem, removeItem
//
// World State:
//   - getWorld, getWorldState, getArea, updateWorld
//
// Quest System:
//   - getQuest, getAllQuests, updateQuest, completeQuest
//
// Procedural Content Generation:
//   - generateDungeon, generateItem, generateQuest, generateNPC
//
// # Real-time Communication
//
// WebSocket connections enable bi-directional communication for:
//   - Combat event broadcasting (attacks, damage, status effects)
//   - Turn notifications (turn order, initiative changes)
//   - State synchronization across multiple clients
//   - Real-time updates for multiplayer sessions
//
// WebSocket connections are established at the /ws endpoint and require valid session IDs.
// All WebSocket messages are JSON-RPC 2.0 formatted for consistency with HTTP endpoints.
//
// # Operational Features
//
// Health Check Endpoints:
//   - /health: Comprehensive health status with component checks (database, cache, dependencies)
//   - /ready: Kubernetes-style readiness probe (returns 200 when server can accept traffic)
//   - /live: Basic liveness probe for load balancers (returns 200 when process is running)
//
// Metrics & Observability:
//   - /metrics: Prometheus metrics endpoint with combat, quest, session, PCG, and performance data
//   - Request rate limiting with configurable thresholds via GOLDBOX_RATE_LIMIT_* variables
//   - Pprof profiling endpoints when enabled via GOLDBOX_ENABLE_PROFILING=true
//   - File-based auto-save with configurable intervals via GOLDBOX_AUTO_SAVE_INTERVAL
//   - Structured logging with logrus including caller context and field-based metadata
//
// # Thread Safety
//
// All server operations are mutex-protected for safe concurrent access.
// Session cleanup and state updates use proper locking patterns with sync.RWMutex.
// WebSocket handlers use goroutine-per-connection model with channel-based synchronization.
// Character, World, and EffectManager types implement thread-safe concurrent operations.
//
// # Example Usage
//
// Basic server initialization and startup:
//
//	package main
//
//	import (
//	    "fmt"
//	    "net"
//	    "goldbox-rpg/pkg/config"
//	    "goldbox-rpg/pkg/server"
//	    "github.com/sirupsen/logrus"
//	)
//
//	func main() {
//	    // Load configuration from environment variables
//	    cfg, err := config.Load()
//	    if err != nil {
//	        logrus.WithError(err).Fatal("Failed to load configuration")
//	    }
//
//	    // Initialize server with web asset directory
//	    srv, err := server.NewRPCServer(cfg.WebDir)
//	    if err != nil {
//	        logrus.WithError(err).Fatal("Failed to initialize server")
//	    }
//
//	    // Create network listener
//	    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ServerPort))
//	    if err != nil {
//	        logrus.WithError(err).Fatal("Failed to start listener")
//	    }
//
//	    // Start server (blocks until shutdown)
//	    logrus.WithField("address", listener.Addr()).Info("Server listening")
//	    if err := srv.Serve(listener); err != nil {
//	        logrus.WithError(err).Fatal("Server failed")
//	    }
//	}
//
// For complete server lifecycle management including graceful shutdown, bootstrap initialization,
// and signal handling, see cmd/server/main.go.
package server
