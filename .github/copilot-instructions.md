# Project Overview

GoldBox RPG Engine is a modern Go-based framework for creating turn-based RPG games inspired by the classic SSI Gold Box series. This engine provides comprehensive character management, combat systems, and world interactions through a JSON-RPC API with WebSocket support for real-time communication. The project targets game developers building web-based RPG experiences with classical tabletop RPG mechanics including D&D-inspired attribute systems, turn-based combat, spell casting, and character progression focused on tactical gameplay with grid-based movement and positioning.

The engine features a complete character system with six core attributes (Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma), multiple character classes (Fighter, Mage, Cleric, Thief, Ranger, Paladin), and an advanced effect system for combat conditions (Stun, Root, Burning, Bleeding, Poison) and status modifications. The architecture emphasizes thread-safe concurrent operations, event-driven gameplay mechanics, and spatial indexing for efficient world queries through an R-tree-like structure. The project includes comprehensive procedural content generation for terrain, items, quests, and NPCs, along with robust system resilience patterns (circuit breakers, retry mechanisms, input validation).

The frontend is an Ebitengine/WASM client compiled from Go. Browser-based visual editors for map creation (`/editor.html`) and quest building (`/quest-builder.html`) enable content authoring. The project includes 10 embedded adventure packs with 100 maps and 37 quests providing 30+ hours of gameplay content.

## Technical Stack

- **Primary Language**: Go 1.25.6 with toolchain 1.25.8
- **Web Framework**: Native Go HTTP server with JSON-RPC 2.0 protocol
- **Real-time Communication**: Coder WebSocket v1.8.14 (nhooyr.io/websocket fork) for production; gorilla/websocket v1.5.3 for E2E test client only
- **Data Format**: YAML v3.0.1 for game data configuration (spells, items, PCG templates)
- **Logging**: Sirupsen Logrus v1.9.4 for structured logging with caller context
- **Utilities**: Google UUID v1.6.0 for entity identification, golang.org/x/exp for extended functionality
- **Testing**: Go built-in testing framework with Testify v1.11.1 for assertions, test coverage analysis scripts, chromedp v0.14.2 for browser playtests
- **Build System**: Makefile with gofumpt formatting, Docker support with health checks, asset generation pipeline
- **Frontend**: Ebitengine v2.9.9/WASM (Go compiled to WebAssembly), launched via splash-screen HTML page
- **Development Tools**: gofumpt for formatting, godocdown for documentation
- **Monitoring**: Prometheus client v1.23.2 for metrics collection, health check endpoints (`/health`, `/ready`, `/live`, `/metrics`)
- **Rate Limiting**: golang.org/x/time v0.15.0 for API throttling
- **Markov Chains**: mb-14/gomarkov for procedural text generation

## Code Assistance Guidelines

1. **Thread Safety First**: All Character and game state modifications must use proper mutex locking (`mu.Lock()` for writes, `mu.RLock()` for reads). Follow the established pattern in `pkg/game/character.go` where concurrent access is protected with `sync.RWMutex`. Example: Character struct uses `mu sync.RWMutex yaml:"-"` and all field modifications require proper locking.

2. **YAML-First Configuration**: Game data (spells, items, character classes) should be defined in YAML files under `/data/` directory. Use struct tags like `yaml:"spell_id"` for proper serialization. Reference `data/spells/cantrips.yaml` for structure examples with fields like `spell_level: 0`, `spell_school: 5`, `damage_type: ""`.

3. **Event-Driven Architecture**: Implement game actions through the event system in `pkg/game/events.go`. Create GameEvent structs with EventType enums and emit events using the EventSystem pattern. Events must include Type, SourceID, TargetID, Data map, and Timestamp for proper game state synchronization.

4. **JSON-RPC Method Pattern**: New server endpoints must follow JSON-RPC 2.0 specification in `pkg/server/handlers.go`. Pattern: validate session with `getSessionForMove()`, process game logic, emit events, return structured response. See `handleMove` implementation with parseMoveRequest, validateCombatConstraints, executePlayerMovement sequence.

5. **Spatial Awareness**: Use the spatial indexing system (`pkg/game/spatial_index.go`) for efficient world queries. Implement position-based operations through the R-tree-like SpatialIndex structure with Rectangle bounds and SpatialNode children rather than brute-force iteration over game objects.

6. **Error Handling Strategy**: Return descriptive errors rather than panicking. Use domain-specific error types from `pkg/game/errors.go` and `pkg/server/errors.go` with sentinel errors and custom error types (CharacterError, CombatError, SessionError, etc.). Use `logrus.WithFields()` for contextual logging. See `docs/ERROR_HANDLING.md` for complete patterns.

7. **Table-Driven Testing**: Write table-driven tests for all business logic functions using Go's testing framework. Follow pattern in `pkg/game/effectbehavior_test.go` with test structs containing name, input parameters, and expected outputs. Include integration tests for API endpoints and maintain ≥60% code coverage using `make test-coverage`.

8. **Procedural Content Generation**: Use the PCG system in `pkg/pcg/` for dynamic content creation. Follow the established Generator interface pattern with proper seeding for deterministic results. PCG content must validate against game schemas before integration. Reference `pkg/pcg/README.md` for complete implementation guidelines.

9. **Resilience Patterns**: Implement circuit breakers from `pkg/resilience/` for external dependencies and critical operations. Use the retry mechanisms in `pkg/retry/` with exponential backoff for transient failures. Critical game operations should be wrapped with resilience patterns to prevent cascade failures.

10. **Input Validation Security**: All JSON-RPC endpoints must use the validation framework in `pkg/validation/` to sanitize user inputs. Validate request size limits, parameter types, and ranges to prevent injection attacks and DoS conditions. Follow the established validation patterns with method-specific validators.

11. **Session Management**: RPCServer.getSession() increments session refcount via session.addRef(); handlers must call RPCServer.releaseSession(session) (usually via defer) on all paths to avoid leaks. Session IDs use cryptographic randomness.

12. **Configuration Management**: Use the configuration system in `pkg/config/` for all application settings. Server config uses env vars: `SERVER_PORT`, `WEB_DIR`, `DATA_DIR`, `LOG_LEVEL`, `ENABLE_DEV_MODE` (no GOLDBOX_ prefix). WebSocket origin validation via `WEBSOCKET_ALLOWED_ORIGINS`. Session timeout configurable via `GOLDBOX_SESSION_TIMEOUT`.

13. **Structured Logging**: Use `logrus.WithFields()` consistently for all logging with function names and relevant context. Initialize logrus with `SetReportCaller(true)` for automatic caller tracking. Log levels: Debug for entry/exit of functions, Info for significant events, Warn for retry attempts, Error for failures.

14. **Character Class Proficiency**: All equipment operations must check class proficiency restrictions defined in `pkg/game/classes.go`. Each CharacterClass has WeaponProficiencies and ArmorProficiencies that must be validated before equipping items. Use the Character.CanEquipItem() method for validation.

15. **WASM Client Patterns**: WASM RPCClient.sessionID access is guarded by a package-level sessionIDMu (RWMutex): RLock for reads (GetSessionID/Call), Lock for writes (SetSessionID/captureSessionID/JoinGame/autoReconnect). Use drawColoredText() for colored debug-font text rendering.

16. **WebSocket Concurrency**: All WebSocket writes must go through writeWSJSON() helper with session.WSWriteMu to prevent data races with concurrent broadcastToAll writes.

## Project Context

- **Domain**: Classical tabletop RPG mechanics digitized with D&D-inspired attribute systems, turn-based combat, spell casting, and character progression. Focus on tactical gameplay with grid-based movement and positioning. Complete spell system with levels 0-9 (60 spells across 10 YAML files).

- **Architecture**: Monolithic server with clear package separation (`game/` for mechanics, `server/` for network layer). Event-driven state management with concurrent session handling. WebSocket connections for real-time updates alongside HTTP JSON-RPC for actions.

- **Key Directories**:
  - `pkg/game/`: Core RPG mechanics (character, combat, spells, world management, effects, equipment, quests, spatial indexing, pathfinding with A*)
  - `pkg/server/`: Network layer (HTTP handlers, WebSocket, session management, health checks, rate limiting, circuit breakers)
  - `pkg/pcg/`: Procedural Content Generation system with terrain, item, quest, NPC, and faction generation. Includes biome-aware algorithms, template systems, and reputation tracking
  - `pkg/resilience/`: Circuit breaker patterns, graceful degradation, and fallback mechanisms for fault tolerance
  - `pkg/validation/`: Comprehensive input validation for JSON-RPC security and DoS prevention
  - `pkg/retry/`: Retry mechanisms with exponential backoff and jitter for reliability
  - `pkg/integration/`: Integration utilities combining resilience and validation patterns
  - `pkg/config/`: Configuration management with environment variable support and YAML loading
  - `pkg/wasmui/`: Ebitengine/WASM game UI client (game.go, rpc_client_wasm.go, adventure_screen.go, character_creation.go, combat_screen.go, map_editor.go, quest_editor.go)
  - `pkg/persistence/`: Game state persistence to YAML files
  - `pkg/secrets/`: Secret management with provider interfaces (Vault, AWS Secrets Manager support)
  - `data/`: YAML configuration files for game content (spells in data/spells/, items in data/items/, PCG templates in data/pcg/, adventures in data/adventures/)
  - `cmd/`: Multiple applications (server/, wasm-ui/, wasm-editor/, map-editor/, quest-builder/, content-creator/, dungeon-demo/, events-demo/, metrics-demo/, validator-demo/, bootstrap-demo/, pcg-demo/, openapi-gen/)
  - `scripts/`: Build automation (generate-all.sh, analyze_test_coverage.sh), asset pipeline, coverage analysis
  - `web/static/`: Static web assets including 521 generated sprites across 6 categories
  - `test/`: Integration tests (e2e/, browser/ with chromedp-based playtests)
  - `docs/`: Technical documentation (ERROR_HANDLING.md, EDITOR_GUIDE.md, NPC_AI.md, WEBSOCKET_MIGRATION.md, CI_CD.md, SECRETS_MANAGEMENT.md)

- **Configuration**: Game content loaded from YAML files at startup using struct tags (e.g., `yaml:"spell_id"`). Server configuration through environment variables. WebSocket origin validation required for production via `WEBSOCKET_ALLOWED_ORIGINS`. Session timeout defaults to 30 minutes. Docker support includes health checks and multi-stage builds.

- **Asset Pipeline**: 521 production-ready sprite assets in `web/static/assets/sprites/` across 6 categories (characters, monsters, items, terrain, effects, UI). Asset generation via `make assets` (4-6 hours with AI tool) or `make assets-download` for pre-generated assets. Configuration in `game-assets.yaml`.

## Quality Standards

- **Testing Requirements**: Maintain ≥60% code coverage (current: 65-96%) with Go's built-in testing framework. Write table-driven tests for business logic with test structs containing name, input parameters, and expected outputs. Include integration tests for API endpoints. Use `go test -race` to detect race conditions in concurrent code. Run coverage analysis with `make test-coverage`. Use `make find-untested` to identify files without tests. Browser playtests run via `make test-browser` using chromedp with SwiftShader WebGL rendering.

- **Code Review Criteria**: All Character state modifications must use proper mutex locking. New game mechanics require corresponding event types. API endpoints must validate session IDs and input parameters. YAML configuration changes need validation against existing schema.

- **Documentation Standards**: Use Go doc comments for all exported functions. Update `pkg/README-RPC.md` for new API endpoints with complete examples. Maintain inline code documentation for complex game mechanics like effect stacking and spatial queries.

- **Security Considerations**: Validate all user inputs in RPC handlers using `pkg/validation/` framework. Implement proper session timeout (currently 30 minutes, configurable). WebSocket origin validation must be enabled for production via `WEBSOCKET_ALLOWED_ORIGINS`. Prevent denial-of-service through input validation and request size limits (default 1MB). Rate limiting via golang.org/x/time protects API endpoints. Use controlled error returns (ErrInvalidSession, etc.) instead of panic().

- **Performance Standards**: Use spatial indexing (`pkg/game/spatial_index.go`) with R-tree-like SpatialIndex structure for world queries instead of linear searches. Implement proper connection pooling for concurrent sessions. Monitor memory usage in effect system to prevent accumulation of expired effects. WebSocket connections use goroutine-per-connection model with proper cleanup. HTTP handlers use timeouts to prevent resource exhaustion.

- **Formatting**: Go code must be formatted with gofumpt (run `make fmt` before committing).

## Networking Best Practices

When declaring network variables, always use interface types for testability and flexibility:

- Never use `net.UDPAddr`, `net.IPAddr`, or `net.TCPAddr`. Use `net.Addr` only instead.
- Never use `net.UDPConn`, use `net.PacketConn` instead.
- Never use `net.TCPConn`, use `net.Conn` instead.
- Never use `net.UDPListener` or `net.TCPListener`, use `net.Listener` instead.
- Never use a type switch or type assertion to convert from an interface type to a concrete type. Use the interface methods instead.

This approach enhances testability and flexibility when working with different network implementations or mocks.

## Build and Run Commands

```bash
# Build and run
make build              # Build server binary
make run                # Build and run server
make wasm               # Build WASM UI
make build-all          # Build server + WASM UI + WASM editor

# Testing
make test               # Run all tests
make test-coverage      # Run coverage analysis
make test-e2e           # Run E2E integration tests
make test-browser       # Run headless browser playtests
go test -race ./...     # Run with race detector

# Docker
make docker-build       # Build Docker image
make docker-run         # Run container
make docker-health      # Check container health

# Assets
make assets-download    # Download pre-generated assets (~50MB)
make assets-verify      # Verify all required assets present

# Code Quality
make fmt                # Format code with gofumpt
make openapi-gen        # Generate OpenAPI spec from code
```

## API Endpoints

- **JSON-RPC**: `POST /rpc` - All game actions via JSON-RPC 2.0
- **WebSocket**: `ws://host/ws` - Real-time game events
- **Health**: `/health`, `/ready`, `/live` - Health probes
- **Metrics**: `/metrics` - Prometheus metrics
- **Editors**: `/editor.html`, `/quest-builder.html` - Visual content editors
- **Swagger**: `/swagger/` - API documentation UI
