# Audit: goldbox-rpg/pkg/server
**Date**: 2026-02-19
**Status**: Needs Work

## Summary
JSON-RPC 2.0 game server implementation handling player sessions, combat mechanics, spell casting, movement, and game state management. Features WebSocket support, session management with reference counting, rate limiting, circuit breakers, health checks, metrics, and profiling. The package is the second largest in the codebase (10900+ lines, 34 test files) but has race conditions in spell handling, incomplete session reference counting, and goroutine leaks in state updates.

## Issues Found
- [ ] **high** Race Condition — `applySpellDamage()` releases RLock early then accesses `session.Player` without protection; session could be deleted/modified concurrently (`spells.go:416-449`)
- [ ] **high** Error Handling — `close(session.MessageChan)` called without checking if already closed; concurrent goroutines could panic on double close (`session.go`)
- [ ] **high** Nil Dereference — `session.Player` accessed without nil check after session release in combat handlers (`handlers.go:284`)
- [ ] **high** Resource Leak — State update timeout goroutine spawned without cleanup; if update completes, goroutine blocks indefinitely on channel read (`state.go:164-177`)
- [ ] **high** API Design — Session reference counting via `addRef()`/`releaseSession()` inconsistently applied; some code paths never release, blocking session cleanup (`types.go:151-162`, `session.go:119`)
- [ ] **med** Test Coverage — No concurrent handler stress tests; session reference counting untested under load
- [ ] **med** Documentation — Four separate mutexes in state.go without documented lock ordering or domain separation (`state.go:43-46`)
- [ ] **med** Error Handling — Multiple `json.NewEncoder().Encode()` calls ignore encoding errors; failed responses silently dropped (`server.go`)
- [ ] **med** API Design — `getSessionSafely()` returns session without documented contract requiring corresponding `releaseSession()` call (`websocket.go:387`)
- [ ] **low** Documentation — Missing package-level doc.go file; doc.md exists but is empty (`pkg/server/`)
- [ ] **low** Performance — Debug logging in hot-path utility functions like `min()` called thousands of times per game loop (`util.go:570-583`)
- [ ] **low** Naming — Inconsistent method receiver names: mix of `s`, `server`, `gs`, `m` for same types
- [ ] **low** Code Quality — 500-element buffered session message channel without backpressure mechanism (`constants.go:32`)

## Test Coverage
55.6% (target: 65%) — ⚠️ BELOW TARGET (tests also have failures)

34 test files covering handlers, WebSocket, sessions, combat, spells, circuit breakers, rate limiting. Coverage gaps in concurrent handler scenarios and session lifecycle edge cases.

## Dependencies
**External:**
- `github.com/google/uuid`: Session ID generation
- `github.com/gorilla/websocket`: WebSocket connections
- `github.com/sirupsen/logrus`: Structured logging
- `golang.org/x/exp/rand`: Initiative rolling

**Internal:**
- `goldbox-rpg/pkg/game`: Game mechanics and entities
- `goldbox-rpg/pkg/config`: Configuration
- `goldbox-rpg/pkg/pcg`: Procedural content generation
- `goldbox-rpg/pkg/validation`: Input validation
- `goldbox-rpg/pkg/persistence`: File storage

## Recommendations
1. **CRITICAL**: Fix race condition in applySpellDamage() — hold lock throughout or copy session reference safely
2. **CRITICAL**: Implement consistent session reference counting — audit all session acquisitions for matching releases
3. **HIGH**: Replace custom timeout goroutine with context.WithTimeout in state updates to prevent goroutine leaks
4. **HIGH**: Add concurrent handler stress tests to reach 65% coverage target
5. **MEDIUM**: Document mutex purposes, lock ordering, and domain separation in state.go
6. **MEDIUM**: Remove debug logging from utility functions in hot paths
7. **LOW**: Standardize method receiver naming conventions
