# Project Overview

GoldBox RPG Engine is a modern Go-based framework for creating turn-based RPG games inspired by the classic SSI Gold Box series. This engine provides comprehensive character management, combat systems, and world interactions through a JSON-RPC API with WebSocket support for real-time communication. The project targets game developers building web-based RPG experiences with classical tabletop RPG mechanics including D&D-inspired attribute systems, turn-based combat, spell casting, and character progression focused on tactical gameplay with grid-based movement and positioning.

The engine features a complete character system with six core attributes (Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma), multiple character classes (Fighter, Mage, Cleric, Thief, Ranger, Paladin), and an advanced effect system for combat conditions (Stun, Root, Burning, Bleeding, Poison) and status modifications. The architecture emphasizes thread-safe concurrent operations, event-driven gameplay mechanics, and spatial indexing for efficient world queries through an R-tree-like structure. The project includes comprehensive procedural content generation for terrain, items, quests, and NPCs, along with robust system resilience patterns (circuit breakers, retry mechanisms, input validation).

## Technical Stack

- **Primary Language**: Go 1.23.0 with toolchain 1.23.2
- **Web Framework**: Native Go HTTP server with JSON-RPC 2.0 protocol
- **Real-time Communication**: Gorilla WebSocket v1.5.3 for live game updates
- **Data Format**: YAML v3.0.1 for game data configuration (spells, items, PCG templates)
- **Logging**: Sirupsen Logrus v1.9.3 for structured logging with caller context
- **Utilities**: Google UUID v1.6.0 for entity identification, golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
- **Testing**: Go built-in testing framework with Testify v1.10.0 for assertions, test coverage analysis scripts
- **Build System**: Makefile with gofumpt formatting, Docker support, asset generation pipeline
- **Frontend**: TypeScript with ES2020 target, ESBuild v0.25.0 bundling, ES2020 module system
- **Development Tools**: Concurrently v8.0.0 for parallel tasks, ESLint for code quality
- **Monitoring**: Prometheus client v1.22.0 for metrics collection, health check endpoints (/health, /ready, /live)
- **Rate Limiting**: golang.org/x/time v0.12.0 for API throttling
- **Markov Chains**: mb-14/gomarkov v0.0.0-20231120193207-9cbdc8df67a8 for procedural text generation

## Code Assistance Guidelines

1. **Thread Safety First**: All Character and game state modifications must use proper mutex locking (`mu.Lock()` for writes, `mu.RLock()` for reads). Follow the established pattern in `pkg/game/character.go` where concurrent access is protected with `sync.RWMutex`. Example: Character struct uses `mu sync.RWMutex yaml:"-"` and all field modifications require proper locking.

2. **YAML-First Configuration**: Game data (spells, items, character classes) should be defined in YAML files under `/data/` directory. Use struct tags like `yaml:"spell_id"` for proper serialization. Reference `data/spells/cantrips.yaml` for structure examples with fields like `spell_level: 0`, `spell_school: 5`, `damage_type: ""`.

3. **Event-Driven Architecture**: Implement game actions through the event system in `pkg/game/events.go`. Create GameEvent structs with EventType enums and emit events using the EventSystem pattern. Events must include Type, SourceID, TargetID, Data map, and Timestamp for proper game state synchronization.

4. **JSON-RPC Method Pattern**: New server endpoints must follow JSON-RPC 2.0 specification in `pkg/server/handlers.go`. Pattern: validate session with `getSessionForMove()`, process game logic, emit events, return structured response. See `handleMove` implementation with parseMoveRequest, validateCombatConstraints, executePlayerMovement sequence.

5. **Spatial Awareness**: Use the spatial indexing system (`pkg/game/spatial_index.go`) for efficient world queries. Implement position-based operations through the R-tree-like SpatialIndex structure with Rectangle bounds and SpatialNode children rather than brute-force iteration over game objects.

6. **Error Handling Strategy**: Return descriptive errors rather than panicking. Use `logrus.WithFields()` for contextual logging with function names and relevant data. Critical game state corruption should use controlled error returns like `ErrInvalidSession`, not `panic()` statements that can crash the server.

7. **Table-Driven Testing**: Write table-driven tests for all business logic functions using Go's testing framework. Follow pattern in `pkg/game/effectbehavior_test.go` with test structs containing name, input parameters, and expected outputs. Include integration tests for API endpoints and maintain >80% code coverage using `make test-coverage`.

8. **Procedural Content Generation**: Use the PCG system in `pkg/pcg/` for dynamic content creation. Follow the established Generator interface pattern with proper seeding for deterministic results. PCG content must validate against game schemas before integration. Reference `pkg/pcg/README.md` for complete implementation guidelines.

9. **Resilience Patterns**: Implement circuit breakers from `pkg/resilience/` for external dependencies and critical operations. Use the retry mechanisms in `pkg/retry/` with exponential backoff for transient failures. Critical game operations should be wrapped with resilience patterns to prevent cascade failures.

10. **Input Validation Security**: All JSON-RPC endpoints must use the validation framework in `pkg/validation/` to sanitize user inputs. Validate request size limits, parameter types, and ranges to prevent injection attacks and DoS conditions. Follow the established validation patterns with method-specific validators.

11. **Integration Patterns**: Use utilities in `pkg/integration/` that combine resilience and validation for robust API endpoints. These patterns should be applied to all external communications and critical game state operations that require both validation and fault tolerance.

12. **Configuration Management**: Use the configuration system in `pkg/config/` for all application settings. Load configuration from environment variables (prefixed with `GOLDBOX_`) and YAML files. Support default values and validation. Example: `GOLDBOX_PORT`, `GOLDBOX_LOG_LEVEL`, `GOLDBOX_SESSION_TIMEOUT` for server configuration.

13. **Structured Logging**: Use `logrus.WithFields()` consistently for all logging with function names and relevant context. Initialize logrus with `SetReportCaller(true)` for automatic caller tracking. Log levels: Debug for entry/exit of functions, Info for significant events, Warn for retry attempts, Error for failures.

14. **Character Class Proficiency**: All equipment operations must check class proficiency restrictions defined in `pkg/game/classes.go`. Each CharacterClass has WeaponProficiencies and ArmorProficiencies that must be validated before equipping items. Use the Character.CanEquipItem() method for validation.

## Project Context

- **Domain**: Classical tabletop RPG mechanics digitized with D&D-inspired attribute systems, turn-based combat, spell casting, and character progression. Focus on tactical gameplay with grid-based movement and positioning.

- **Architecture**: Monolithic server with clear package separation (`game/` for mechanics, `server/` for network layer). Event-driven state management with concurrent session handling. WebSocket connections for real-time updates alongside HTTP JSON-RPC for actions.

- **Key Directories**:
  - `pkg/game/`: Core RPG mechanics (character, combat, spells, world management, effects, equipment, quests, spatial indexing)
  - `pkg/server/`: Network layer (HTTP handlers, WebSocket, session management, health checks)
  - `pkg/pcg/`: Procedural Content Generation system with terrain, item, quest, and NPC generation. Includes biome-aware algorithms and template systems
  - `pkg/resilience/`: Circuit breaker patterns and system resilience components for fault tolerance
  - `pkg/validation/`: Comprehensive input validation for JSON-RPC security and DoS prevention
  - `pkg/retry/`: Retry mechanisms with exponential backoff and jitter for reliability
  - `pkg/integration/`: Integration utilities combining resilience and validation patterns
  - `pkg/config/`: Configuration management with environment variable support (GOLDBOX_ prefix) and YAML loading
  - `data/`: YAML configuration files for game content (spells in data/spells/, items in data/items/, PCG templates in data/pcg/)
  - `src/`: TypeScript frontend modules (core/, game/, network/, ui/, utils/, types/, rendering/) with ES2020 target
  - `cmd/`: Multiple applications (server/, dungeon-demo/, events-demo/, metrics-demo/, validator-demo/, bootstrap-demo/)
  - `scripts/`: Build automation (generate-all.sh, analyze_test_coverage.sh), asset pipeline, coverage analysis
  - `web/static/`: Static web assets including generated sprites and UI elements

- **Configuration**: Game content loaded from YAML files at startup using struct tags (e.g., `yaml:"spell_id"`). Server configuration through environment variables with GOLDBOX_ prefix (e.g., GOLDBOX_PORT=8080, GOLDBOX_LOG_LEVEL=debug). WebSocket origin validation required for production deployment via WEBSOCKET_ALLOWED_ORIGINS environment variable (currently allows all origins in development mode). Session timeout defaults to 30 minutes (configurable via GOLDBOX_SESSION_TIMEOUT). Docker support includes health checks and multi-stage builds.

## Quality Standards

- **Testing Requirements**: Maintain >80% code coverage with Go's built-in testing framework (currently at 78% with 106 test files). Write table-driven tests for business logic with test structs containing name, input parameters, and expected outputs (see `pkg/game/effectbehavior_test.go` for examples). Include integration tests for API endpoints. Use `go test -race` to detect race conditions in concurrent code. Run coverage analysis with `make test-coverage` or `./scripts/analyze_test_coverage.sh`. Use `make find-untested` to identify files without tests.

- **Code Review Criteria**: All Character state modifications must use proper mutex locking. New game mechanics require corresponding event types. API endpoints must validate session IDs and input parameters. YAML configuration changes need validation against existing schema.

- **Documentation Standards**: Use Go doc comments for all exported functions. Update `pkg/README-RPC.md` for new API endpoints with complete examples. Maintain inline code documentation for complex game mechanics like effect stacking and spatial queries.

- **Security Considerations**: Validate all user inputs in RPC handlers using `pkg/validation/` framework. Implement proper session timeout (currently 30 minutes, configurable via GOLDBOX_SESSION_TIMEOUT). WebSocket origin validation must be enabled for production via WEBSOCKET_ALLOWED_ORIGINS environment variable. Prevent denial-of-service through input validation and request size limits (default 1MB) rather than panic conditions. Use ErrInvalidSession and other controlled error returns instead of panic() in server code. Rate limiting via golang.org/x/time protects API endpoints from abuse.

- **Performance Standards**: Use spatial indexing (`pkg/game/spatial_index.go`) with R-tree-like SpatialIndex structure for world queries instead of linear searches. Implement proper connection pooling for concurrent sessions. Monitor memory usage in effect system (`pkg/game/effects.go`) to prevent accumulation of expired effects. Use efficient YAML parsing with struct tags for game data loading. WebSocket connections use goroutine-per-connection model with proper cleanup. HTTP handlers use timeouts to prevent resource exhaustion.
