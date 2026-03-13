# Implementation Plan: WebSocket Library Migration & Code Consolidation

## Project Context
- **What it does**: GoldBox RPG Engine is a modern Go-based framework for creating turn-based RPG games inspired by the classic SSI Gold Box series, providing character management, combat systems, and world interactions through a JSON-RPC API with WebSocket support.
- **Current goal**: Migrate from archived gorilla/websocket to actively-maintained nhooyr.io/websocket for long-term security and maintainability.
- **Estimated Scope**: Medium

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character Management (6 attributes, classes, creation) | ✅ Achieved | No |
| Combat & Effects System | ✅ Achieved | No |
| World Management + Spatial Indexing | ✅ Achieved | No |
| Event-Driven Architecture | ✅ Achieved | No |
| WebSocket Real-time Communication | ✅ Achieved | **Yes** (security migration) |
| Health Monitoring & Prometheus Metrics | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Circuit Breakers & Resilience | ✅ Achieved | No |
| Input Validation Framework | ✅ Achieved | No |
| Asset Generation Pipeline | ⚠️ Partial (requires external AI tool) | No |
| Advanced NPC AI | ✅ Achieved | No |
| Spell System (levels 0-9) | ✅ Achieved | No |
| World Editor Tools | ⚠️ Partial (CLI only) | No |
| Network Optimization (delta compression) | ✅ Achieved | No |
| Player Progression Persistence | ✅ Achieved | No |
| Guild and Faction Systems | ✅ Achieved | No |
| Embedded Adventures (10 packs) | ✅ Achieved | No |
| **Code Maintainability** | ⚠️ Duplication hotspots | **Yes** |

**Overall: 15/17 goals fully achieved, 2 partial (asset generation, editor GUI)**

## Metrics Summary

| Metric | Value | Assessment |
|--------|-------|------------|
| **Total Lines of Code** | 30,736 | — |
| **Total Functions** | 596 functions + 1,694 methods | — |
| **Total Packages** | 18 | — |
| **High Complexity Functions (>9)** | 4 | ✅ Low risk |
| **Documentation Coverage** | 89.0% (2038/2290) | ✅ Excellent |
| **Code Duplication** | 1.98% (1,207 lines, 46 clone pairs) | ✅ Acceptable |
| **Test Coverage** | 70-95% per package | ✅ Above 60% CI threshold |
| **Circular Dependencies** | 0 | ✅ Clean |

### Duplication Hotspots Requiring Consolidation

| Package | Clone Locations | Lines | Priority |
|---------|----------------|-------|----------|
| `pkg/game/faction_relations.go` | 10 locations | 11+ lines each | Medium |
| `pkg/server/handlers_guild.go` | 4 locations | 14-16 lines | Medium |
| `pkg/server/handlers_pcg.go` | 4 locations | 14-15 lines | Low |
| `pkg/game/guild.go` | 4 locations | 14 lines | Low |
| `pkg/server/server.go` | 5 locations | 14-18 lines | Low |

### Security Considerations from Research

| Issue | Status | Risk | Action |
|-------|--------|------|--------|
| gorilla/websocket archived (2022) | ⚠️ | Medium | Migration planned |
| CVE-2020-27813 (integer overflow) | Patched in v1.4.1 | Low | Current v1.5.3 not affected |
| Go stdlib vulnerabilities (18 CVEs) | Known | Medium | Requires Go 1.24.12+ when available |
| WebSocket origin validation | ✅ Implemented | — | Production ready |

## Implementation Steps

### Step 1: Create WebSocket Adapter Interface

- **Deliverable**: New file `pkg/server/websocket_adapter.go` with `WebSocketConn` interface abstracting library-specific APIs
- **Dependencies**: None
- **Goal Impact**: Enables gradual WebSocket library migration without breaking changes
- **Acceptance**: Interface compiles; no behavior change to existing code
- **Validation**: `go build ./pkg/server/...`

```go
// pkg/server/websocket_adapter.go
type WebSocketConn interface {
    ReadMessage(ctx context.Context) (messageType int, p []byte, err error)
    WriteMessage(ctx context.Context, messageType int, data []byte) error
    Close(code int, reason string) error
    RemoteAddr() net.Addr
}
```

### Step 2: Add nhooyr.io/websocket Dependency

- **Deliverable**: Updated `go.mod` with both websocket libraries; feature flag `USE_NHOOYR_WEBSOCKET` environment variable
- **Dependencies**: Step 1
- **Goal Impact**: Prepares codebase for dual-library operation during migration
- **Acceptance**: `go mod tidy` succeeds; tests pass with feature flag disabled
- **Validation**: `go mod tidy && go test ./... -race -short`

### Step 3: Implement gorilla Adapter

- **Deliverable**: `pkg/server/websocket_gorilla.go` implementing `WebSocketConn` interface using existing gorilla code
- **Dependencies**: Step 1
- **Goal Impact**: Isolates gorilla-specific code behind interface
- **Acceptance**: All E2E WebSocket tests pass with gorilla adapter
- **Validation**: `go test ./test/e2e/... -v -run WebSocket`

### Step 4: Implement nhooyr Adapter

- **Deliverable**: `pkg/server/websocket_nhooyr.go` implementing `WebSocketConn` interface using nhooyr.io/websocket
- **Dependencies**: Steps 1, 2
- **Goal Impact**: Provides modern, actively-maintained WebSocket implementation
- **Acceptance**: All E2E WebSocket tests pass with `USE_NHOOYR_WEBSOCKET=true`
- **Validation**: `USE_NHOOYR_WEBSOCKET=true go test ./test/e2e/... -v -run WebSocket`

### Step 5: Migrate Primary WebSocket Handler

- **Deliverable**: Updated `pkg/server/websocket.go` using `WebSocketConn` interface; feature flag selects implementation
- **Dependencies**: Steps 3, 4
- **Goal Impact**: Core real-time communication uses abstracted interface
- **Acceptance**: Both adapters work; delta compression preserved; origin validation functional
- **Validation**: 
  ```bash
  go test ./pkg/server/... -v -run WebSocket
  USE_NHOOYR_WEBSOCKET=true go test ./test/e2e/... -v
  ```

### Step 6: Migrate Editor WebSocket

- **Deliverable**: Updated `pkg/server/websocket_editor.go` using `WebSocketConn` interface
- **Dependencies**: Step 5
- **Goal Impact**: Map editor real-time collaboration uses modern library
- **Acceptance**: Editor protocol (`tile_update`, `map_created`, etc.) works with both adapters
- **Validation**: `go test ./pkg/server/... -v -run Editor`

### Step 7: Update E2E Test Client

- **Deliverable**: Updated `test/e2e/client.go` using `WebSocketConn` interface
- **Dependencies**: Step 5
- **Goal Impact**: Tests can verify both implementations
- **Acceptance**: E2E tests run with either adapter via environment variable
- **Validation**: 
  ```bash
  go test ./test/e2e/... -v
  USE_NHOOYR_WEBSOCKET=true go test ./test/e2e/... -v
  ```

### Step 8: WebSocket Benchmark Comparison

- **Deliverable**: Updated `pkg/server/benchmark_test.go` with comparative benchmarks for both libraries
- **Dependencies**: Steps 5, 6, 7
- **Goal Impact**: Ensures migration doesn't regress performance
- **Acceptance**: nhooyr within 10% of gorilla benchmark baseline
- **Validation**: 
  ```bash
  go test ./pkg/server/... -bench=BenchmarkWebSocket -benchmem | tee /tmp/gorilla.txt
  USE_NHOOYR_WEBSOCKET=true go test ./pkg/server/... -bench=BenchmarkWebSocket -benchmem | tee /tmp/nhooyr.txt
  ```

### Step 9: Extract Faction Relations Helpers

- **Deliverable**: New helper functions in `pkg/game/faction_helpers.go` consolidating 10+ duplicated reputation modification blocks in `faction_relations.go`
- **Dependencies**: None (can run in parallel with WebSocket migration)
- **Goal Impact**: Reduces duplication ratio; improves maintainability
- **Acceptance**: Duplication ratio in `faction_relations.go` reduced by 50%+
- **Validation**: 
  ```bash
  go test ./pkg/game/... -v -run Faction
  go-stats-generator analyze ./pkg/game/faction_relations.go --sections duplication | grep ratio
  ```

### Step 10: Extract Guild Handler Pattern

- **Deliverable**: New `pkg/server/rpc_helpers.go` with common RPC response pattern used across guild handlers
- **Dependencies**: None (can run in parallel)
- **Goal Impact**: Reduces 4 x 14-16 line duplications in `handlers_guild.go`
- **Acceptance**: Handler tests pass; duplication reduced
- **Validation**: 
  ```bash
  go test ./pkg/server/... -v -run Guild
  go-stats-generator analyze ./pkg/server/handlers_guild.go --sections duplication | grep ratio
  ```

### Step 11: Remove gorilla Dependency (Final)

- **Deliverable**: 
  - Remove gorilla/websocket from `go.mod`
  - Delete `pkg/server/websocket_gorilla.go`
  - Remove feature flag code
  - Update vendor directory
- **Dependencies**: Steps 1-8 complete; validation period passed
- **Goal Impact**: Single, actively-maintained WebSocket library
- **Acceptance**: All tests pass; no gorilla imports remain
- **Validation**: 
  ```bash
  grep -r "gorilla/websocket" . --include="*.go" | wc -l  # Should be 0
  go mod tidy && go test -race ./...
  ```

### Step 12: Update Documentation

- **Deliverable**: 
  - Update `docs/WEBSOCKET_MIGRATION.md` status to "Complete"
  - Update `CHANGELOG.md` with migration notes
  - Update `GAPS.md` to remove WebSocket maintenance concern
- **Dependencies**: Step 11
- **Goal Impact**: Documentation reflects current state
- **Acceptance**: All status markers accurate
- **Validation**: Manual review; no outdated references to gorilla

## Dependency Graph

```
Step 1 (Adapter Interface)
    │
    ├── Step 2 (Add nhooyr dependency)
    │       │
    │       └── Step 4 (nhooyr adapter)
    │               │
    └── Step 3 (gorilla adapter)
            │
            └── Step 5 (Migrate websocket.go)
                    │
                    ├── Step 6 (Migrate editor WebSocket)
                    │
                    ├── Step 7 (Update E2E client)
                    │       │
                    │       └── Step 8 (Benchmarks)
                    │               │
                    │               └── Step 11 (Remove gorilla)
                    │                       │
                    │                       └── Step 12 (Documentation)

[Parallel Track - Duplication Reduction]
Step 9 (Faction helpers) ─────────┐
                                  │
Step 10 (Guild handler pattern) ──┘
```

## Rollback Plan

If issues are discovered post-migration:

```bash
# Tag pre-migration state before starting
git tag pre-websocket-migration

# Rollback if needed
git checkout pre-websocket-migration -- go.mod go.sum pkg/server/ test/e2e/client.go
go mod vendor
go test -race ./...
```

## Future Enhancements (Not This Plan)

These items are documented for future consideration but are out of scope:

1. **GUI World Editor** — Browser-based map editor using existing WASM infrastructure (see `docs/WEBSOCKET_MIGRATION.md` Priority 7)
2. **AI Art Generation** — Complete 521-asset generation requires external Stable Diffusion/DALL-E setup
3. **Go Toolchain Upgrade** — Go 1.24.12+ when available to resolve 18 stdlib CVEs

## Validation Summary

| Step | Validation Command |
|------|-------------------|
| 1 | `go build ./pkg/server/...` |
| 2 | `go mod tidy && go test ./... -race -short` |
| 3 | `go test ./test/e2e/... -v -run WebSocket` |
| 4 | `USE_NHOOYR_WEBSOCKET=true go test ./test/e2e/... -v -run WebSocket` |
| 5 | `go test ./pkg/server/... -v -run WebSocket` |
| 6 | `go test ./pkg/server/... -v -run Editor` |
| 7 | `go test ./test/e2e/... -v` |
| 8 | `go test ./pkg/server/... -bench=BenchmarkWebSocket` |
| 9 | `go test ./pkg/game/... -v -run Faction` |
| 10 | `go test ./pkg/server/... -v -run Guild` |
| 11 | `grep -r "gorilla/websocket" . --include="*.go" \| wc -l` (expect 0) |
| 12 | Manual documentation review |

---

*Generated: 2026-03-13*  
*Tool: go-stats-generator v1.0.0*  
*Files Analyzed: 184 Go files, 30,736 lines of code*
