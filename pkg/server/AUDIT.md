# Audit: goldbox-rpg/pkg/server
**Date**: 2026-02-18
**Status**: Complete

## Summary
The server package implements the JSON-RPC 2.0 API layer, WebSocket broadcasting, session management, and game state orchestration. Overall health is good with robust concurrency patterns and integration with resilience/validation subsystems. Critical issues include mutex copy in tests, missing package documentation, and non-deterministic PCG seeding.

## Issues Found
- [x] high concurrency — Mutex copy in test code causes race condition (`handlers_test.go:45`)
- [x] med documentation — No `doc.go` file for package-level documentation
- [x] med determinism — Direct `time.Now()` usage in PCG seeding breaks reproducibility (`server.go:180`)
- [x] low stub — TODO comment for version info hardcoded instead of build-time injection (`health.go:81`)
- [x] low error — Intentionally suppressed session variables in handlers (`handlers.go:2786, 2977, 3099, 3177, 3324, 3410, 3451`)
- [x] low determinism — Direct `time.Now()` usage in TimeManager initialization (`state.go:437, 442`)

## Test Coverage
55.6% (target: 65%)
**Status**: Below target. Coverage gap of 9.4 percentage points.

## Dependencies
**Internal:**
- `goldbox-rpg/pkg/game` - Core game mechanics
- `goldbox-rpg/pkg/pcg` - Procedural content generation
- `goldbox-rpg/pkg/config` - Configuration management
- `goldbox-rpg/pkg/validation` - Input validation
- `goldbox-rpg/pkg/resilience` - Circuit breaker patterns
- `goldbox-rpg/pkg/retry` - Retry mechanisms
- `goldbox-rpg/pkg/persistence` - File-based persistence

**External:**
- `github.com/gorilla/websocket` - WebSocket connections
- `github.com/sirupsen/logrus` - Structured logging
- `github.com/google/uuid` - Session ID generation
- `github.com/prometheus/client_golang` - Metrics collection
- `golang.org/x/time/rate` - Rate limiting

## Recommendations
1. **Fix mutex copy bug**: Change `Character: *character` to `Character: character` in `handlers_test.go:45` to pass pointer, not dereference
2. **Add package documentation**: Create `doc.go` with package overview, architecture, and usage examples
3. **Make PCG seeding configurable**: Add optional seed parameter to `NewRPCServer()` or config, fallback to `time.Now().UnixNano()` only in non-test mode
4. **Improve test coverage**: Add tests for error paths in handlers, WebSocket upgrade failures, and session cleanup edge cases to reach 65%+ target
5. **Inject version at build time**: Replace hardcoded version in `health.go:81` with `-ldflags` build injection
