# Project Overview

GoldBox RPG Engine is a modern Go-based framework for creating turn-based RPG games inspired by the classic SSI Gold Box series. This engine provides comprehensive character management, combat systems, and world interactions through a JSON-RPC API with WebSocket support for real-time communication. The project targets game developers building web-based RPG experiences with robust backend systems and browser-based frontends.

The engine features a complete character system with six core attributes (Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma), multiple character classes, and an advanced effect system for combat conditions and status modifications. The architecture emphasizes thread-safe concurrent operations, event-driven gameplay mechanics, and spatial indexing for efficient world queries.

## Technical Stack

- **Primary Language**: Go 1.22.0 with toolchain 1.22.10
- **Web Framework**: Native Go HTTP server with JSON-RPC 2.0 protocol
- **Real-time Communication**: Gorilla WebSocket v1.5.3 for live game updates
- **Data Format**: YAML v3.0.1 for game data configuration (spells, items)
- **Logging**: Sirupsen Logrus v1.9.3 for structured logging
- **Utilities**: Google UUID v1.6.0 for entity identification, golang.org/x/exp for experimental features
- **Testing**: Go built-in testing framework with table-driven tests
- **Build System**: Makefile with gofumpt formatting and custom documentation generation
- **Frontend**: Vanilla JavaScript with EventEmitter pattern for state management

## Code Assistance Guidelines

1. **Thread Safety First**: All Character and game state modifications must use proper mutex locking (`mu.Lock()` for writes, `mu.RLock()` for reads). Follow the established pattern in `pkg/game/character.go` where concurrent access is protected with `sync.RWMutex`.

2. **YAML-First Configuration**: Game data (spells, items, character classes) should be defined in YAML files under `/data/` directory. Use struct tags like `yaml:"spell_id"` for proper serialization. Reference `data/spells/cantrips.yaml` for structure examples.

3. **Event-Driven Architecture**: Implement game actions through the event system in `pkg/game/events.go`. Create appropriate event types (Movement, Combat, SpellCasting) and emit events using the established EventSystem pattern for proper game state synchronization.

4. **JSON-RPC Method Pattern**: New server endpoints must follow the JSON-RPC 2.0 specification. Implement handler functions in `pkg/server/handlers.go` with pattern: validate session, process game logic, emit events, return structured response. See `handleMove` for reference implementation.

5. **Spatial Awareness**: Use the spatial indexing system (`pkg/game/spatial_index.go`) for efficient world queries. Implement position-based operations through the established R-tree-like structure rather than brute-force iteration over game objects.

6. **Error Handling Strategy**: Return descriptive errors rather than panicking. Use `logrus.WithFields()` for contextual logging. Critical game state corruption should use controlled error returns, not `panic()` statements that can crash the server.

7. **Testing Requirements**: Write table-driven tests for all business logic functions. Maintain >80% code coverage using Go's built-in testing package. Include integration tests for API endpoints using testify patterns. Test files should match `*_test.go` convention.

## Project Context

- **Domain**: Classical tabletop RPG mechanics digitized with D&D-inspired attribute systems, turn-based combat, spell casting, and character progression. Focus on tactical gameplay with grid-based movement and positioning.

- **Architecture**: Monolithic server with clear package separation (`game/` for mechanics, `server/` for network layer). Event-driven state management with concurrent session handling. WebSocket connections for real-time updates alongside HTTP JSON-RPC for actions.

- **Key Directories**:
  - `pkg/game/`: Core RPG mechanics (character, combat, spells, world management)
  - `pkg/server/`: Network layer (HTTP handlers, WebSocket, session management)
  - `data/`: YAML configuration files for game content (spells, items)
  - `web/static/`: Frontend JavaScript modules for game UI and state management
  - `cmd/server/`: Application entry point and server initialization

- **Configuration**: Game content loaded from YAML files at startup. Server configuration through environment variables. WebSocket origins validation required for production deployment (currently allows all origins for development).

## Quality Standards

- **Testing Requirements**: Maintain >80% code coverage with Go's built-in testing framework. Write table-driven tests for business logic, integration tests for API endpoints. Use `go test -race` to detect race conditions in concurrent code.

- **Code Review Criteria**: All Character state modifications must use proper mutex locking. New game mechanics require corresponding event types. API endpoints must validate session IDs and input parameters. YAML configuration changes need validation against existing schema.

- **Documentation Standards**: Use Go doc comments for all exported functions. Update `pkg/README-RPC.md` for new API endpoints with complete examples. Maintain inline code documentation for complex game mechanics like effect stacking and spatial queries.

- **Security Considerations**: Validate all user inputs in RPC handlers. Implement proper session timeout (currently 30 minutes). WebSocket origin validation must be enabled for production. Prevent denial-of-service through input validation rather than panic conditions.

- **Performance Standards**: Use spatial indexing for world queries instead of linear searches. Implement proper connection pooling for concurrent sessions. Monitor memory usage in effect system to prevent accumulation of expired effects.
