# Implementation Plan: Phase 2 - High Priority Improvements

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by the classic SSI Gold Box series, providing turn-based combat, character management, and world interactions through JSON-RPC with WebSocket support.
- **Current milestone**: Phase 2 - High Priority Issues (from ROADMAP.md)
- **Estimated Scope**: **Large** (>15 items across 8 sub-phases)

## Completed Phases
| Phase | Status | Evidence |
|-------|--------|----------|
| 1.1 Data Persistence | ✅ Complete | `pkg/persistence/` with filestore.go, lock.go, atomic.go |
| 1.2 CI/CD Pipeline | ✅ Complete | `.github/workflows/ci.yml`, `build.yml` |
| 1.3 Secrets Management | ⏳ Pending | Not started per ROADMAP |
| 1.4 Deployment Manifests | ⏳ Pending | Not started per ROADMAP |

## Metrics Summary (via go-stats-generator)

### Complexity Analysis
- **Functions above cyclomatic complexity 10**: 15 functions
- **Functions above cyclomatic complexity 9**: 36 functions
- **Highest complexity**: `addTorchPositions()` at 18 (pkg/pcg/terrain/generator.go)

| Package | High-Complexity Functions (≥10) |
|---------|--------------------------------|
| wasmui | 4 |
| terrain | 3 |
| pcg | 2 |
| main (demos) | 2 |
| game, items, levels, e2e | 1 each |

### Code Duplication
- **Duplication ratio**: 1.27% ✅ (below 3% threshold - excellent)
- **Clone pairs**: 20
- **Duplicated lines**: 606 of 23,993 total

### Documentation Coverage
- **Overall coverage**: 85.5%
- **Functions documented**: 97.8%
- **Undocumented functions**: 258 (primarily unexported)
- **TODO comments**: 2 (faction territory, build version)

### Package Health
| Package | Coupling | Cohesion | Functions | Assessment |
|---------|----------|----------|-----------|------------|
| server | 3.5 | 3.3 | 334 | High coupling - central coordinator |
| game | 1.5 | 2.6 | 296 | Balanced - core domain |
| pcg | 1.5 | 6.6 | 489 | High cohesion - well-encapsulated |
| wasmui | 1.5 | 2.8 | 47 | Contains 4 complex functions |

### Codebase Overview
- **Total LOC**: 23,993 (non-test)
- **Total functions**: 362 functions + 1,301 methods
- **Packages**: 17
- **Files analyzed**: 133

---

## Implementation Steps

### Step 1: Complete Test Coverage for Core Files (Phase 2.1) ✅ COMPLETE
- **Deliverable**: Add test files for 6 critical untested packages
- **Dependencies**: None (can start immediately)
- **Priority**: Highest - blocks all other quality improvements
- **Files Created/Enhanced**:
  - `pkg/game/character_test.go` - ✅ EXISTS (364 lines)
  - `pkg/game/effects_test.go` - ✅ EXISTS (effectbehavior_test.go, effectimmunity_test.go, effectmanager_test.go)
  - `pkg/server/handlers_test.go` - ✅ EXISTS (628 lines)
  - `pkg/server/server_test.go` - ✅ CREATED (329 lines) - NEW
  - `pkg/server/health_test.go` - ✅ EXISTS (health_comprehensive_test.go)
  - `pkg/pcg/manager_test.go` - ✅ EXISTS (380 lines)
- **Results**:
  - pkg/game coverage: 80.5%
  - pkg/server coverage: 65.2%
  - Overall project test coverage maintained at 74-78% range
- **Validation**:
  ```bash
  go test ./pkg/game ./pkg/server -cover
  # pkg/game: 80.5% ✅
  # pkg/server: 65.2% (improved from uncovered server.go)
  ```

### Step 2: Update Dependencies and Fix Vulnerabilities (Phase 2.2)
- **Deliverable**: All dependencies updated, zero vulnerabilities
- **Dependencies**: Step 1 (tests must pass before updating)
- **Files to Modify**:
  - `go.mod` - Update prometheus/client_golang v1.22.0 → latest
  - `go.sum` - Regenerated checksums
  - `.github/dependabot.yml` - New file for auto-updates
- **Acceptance Criteria**: `govulncheck ./...` reports zero vulnerabilities
- **Validation**:
  ```bash
  go install golang.org/x/vuln/cmd/govulncheck@latest
  govulncheck ./...
  # Target: 0 vulnerabilities found
  ```

### Step 3: Implement Structured Error Handling (Phase 2.3)
- **Deliverable**: Domain-specific error types with proper wrapping
- **Dependencies**: Step 1 (need tests before refactoring)
- **Files to Create**:
  - `pkg/game/errors.go` - Game-specific error types
  - `pkg/server/errors.go` - Server-specific error types
  - `docs/ERROR_HANDLING.md` - Error handling guide
- **Acceptance Criteria**: `errors.Is/As` usage increases from 8 to 50+ instances
- **Validation**:
  ```bash
  grep -r "errors\.\(Is\|As\|Unwrap\)" --include="*.go" pkg/ | wc -l
  # Target: >= 50 instances
  ```

### Step 4: Expand End-to-End Integration Tests (Phase 2.4)
- **Deliverable**: Complete E2E test coverage for all major workflows
- **Dependencies**: Steps 1-3
- **Current State**: E2E framework exists in `test/e2e/` with basic tests
- **Files to Create/Enhance**:
  - `test/e2e/combat_test.go` - Combat workflow tests
  - `test/e2e/spell_test.go` - Spell casting tests
  - `test/e2e/inventory_test.go` - Item/equipment tests
  - `test/e2e/pcg_test.go` - PCG content tests
  - `test/e2e/websocket_test.go` - Real-time event tests
- **Acceptance Criteria**: E2E tests cover 10+ major user workflows
- **Validation**:
  ```bash
  go test ./test/e2e/... -v -count=1
  # Target: All E2E tests pass
  ```

### Step 5: Implement Request Tracing (Phase 2.5)
- **Deliverable**: Correlation ID middleware for request tracing
- **Dependencies**: Step 4
- **Files to Create**:
  - `pkg/server/correlation.go` - Correlation ID middleware
  - `pkg/server/tracing.go` - Tracing utilities
- **Acceptance Criteria**: All log entries include correlation ID
- **Validation**:
  ```bash
  # Start server and make request
  curl -X POST http://localhost:8080/rpc -d '{"jsonrpc":"2.0","method":"getState","id":1}'
  # Check logs contain X-Correlation-ID
  grep -c "correlation_id" /tmp/server.log
  # Target: >= 1 per request
  ```

### Step 6: Enhance Dockerfile for Production (Phase 2.6)
- **Deliverable**: Optimized multi-stage Docker build
- **Dependencies**: None (independent)
- **Files to Modify**:
  - `Dockerfile` - Multi-stage build with distroless base
  - `.dockerignore` - Optimize build context
- **Files to Create**:
  - `Dockerfile.debug` - Debug image with tools
- **Acceptance Criteria**: Production image <50MB
- **Validation**:
  ```bash
  docker build -t goldbox-rpg:prod .
  docker images goldbox-rpg:prod --format "{{.Size}}"
  # Target: < 50MB
  ```

### Step 7: Add Performance Benchmarks (Phase 2.7)
- **Deliverable**: Benchmark tests for critical paths
- **Dependencies**: Step 1
- **Files to Create**:
  - `pkg/game/character_bench_test.go`
  - `pkg/game/spatial_bench_test.go`
  - `pkg/server/handlers_bench_test.go`
  - `pkg/pcg/terrain/bench_test.go`
- **Targets**:
  - RPC request: <10ms p95
  - Spatial query: <1ms p95
  - Combat round: <50ms p95
  - PCG level generation: <500ms
- **Validation**:
  ```bash
  go test -bench=. -benchmem ./pkg/game/... ./pkg/server/...
  # Analyze output for performance targets
  ```

### Step 8: Implement Session Persistence (Phase 2.8)
- **Deliverable**: Sessions persist across server restarts
- **Dependencies**: Step 1 (Phase 1.1 persistence already complete)
- **Files to Create**:
  - `pkg/persistence/session_store.go` - SessionStore interface
  - `pkg/persistence/memory_store.go` - In-memory implementation for dev
- **Files to Modify**:
  - `pkg/server/server.go` - Use session store
  - `pkg/config/config.go` - Session persistence config
- **Acceptance Criteria**: Sessions survive server restart
- **Validation**:
  ```bash
  # 1. Start server, create session
  # 2. Restart server
  # 3. Verify session still valid
  ls -la data/sessions/*.yaml | wc -l
  # Target: Files exist after restart
  ```

---

## Complexity Hotspots Requiring Attention

The following functions exceed cyclomatic complexity 10 and should be refactored during related work:

| Function | Package | File | Complexity | Suggested Action |
|----------|---------|------|------------|-----------------|
| `addTorchPositions` | terrain | generator.go | 18 | Extract helper functions |
| `ModifyReputation` | pcg | reputation.go | 14 | Decompose into smaller methods |
| `addVegetation` | terrain | generator.go | 14 | Extract biome-specific logic |
| `refreshGameState` | wasmui | game.go | 14 | Split state update logic |
| `GenerateLevel` | levels | generator.go | 13 | Use strategy pattern |
| `LoadFromFile` | items | templates.go | 12 | Separate parsing/validation |

**Note**: These should be addressed opportunistically during Steps 1-7, not as separate tasks.

---

## Metrics Validation Commands

### Overall Health Check
```bash
go-stats-generator analyze . --skip-tests --format json | jq '{
  complexity_hotspots: [.functions[] | select(.complexity.cyclomatic >= 10)] | length,
  duplication_ratio: .duplication.duplication_ratio,
  doc_coverage: .documentation.coverage.overall,
  total_functions: .overview.total_functions
}'
```

### Post-Step Verification
```bash
# After Step 1 (Test Coverage)
go test ./... -cover | grep -E "^ok|coverage"

# After Step 2 (Dependencies)
govulncheck ./... 2>&1 | grep -E "vulnerability|No vulnerabilities"

# After Step 3 (Error Handling)
grep -r "errors\.\(Is\|As\)" --include="*.go" pkg/ | wc -l

# After Step 6 (Docker)
docker images goldbox-rpg --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
```

---

## Scope Assessment Rationale

**Classification: Large** based on:
- 15 functions above complexity threshold (threshold: >15 = Large)
- 8 implementation steps in Phase 2
- ~2,500 lines of new test code required
- Multiple packages affected
- Estimated 3-4 weeks of effort

## Out of Scope (Phase 3+)

The following are documented in ROADMAP.md but NOT part of this plan:
- Phase 1.3: Secrets Management (Critical, but separate sprint)
- Phase 1.4: Deployment Manifests (Critical, but separate sprint)
- Phase 3.x: OpenAPI documentation, graceful degradation, audit logging
- Phase 4.x: API versioning, caching layer

See `ROADMAP.md` for complete backlog.

---

*Generated: 2026-03-11 | Tool: go-stats-generator v1.0.0 | Analyzed: 133 files, 23,993 LOC*
