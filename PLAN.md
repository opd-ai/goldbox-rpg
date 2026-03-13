# Implementation Plan: Security & Maintainability Hardening

**Generated:** 2026-03-13  
**Analysis Tool:** go-stats-generator v1.0.0  
**Files Analyzed:** 184 Go files, 30,736 lines of code

## Project Context

- **What it does**: GoldBox RPG Engine is a modern Go-based framework for creating turn-based RPG games with JSON-RPC API and WebSocket real-time communication, featuring comprehensive character management, combat systems, and procedural content generation.
- **Current goal**: Upgrade Go toolchain to address stdlib vulnerabilities and migrate WebSocket library for long-term maintainability.
- **Estimated Scope**: Medium (5–15 items above threshold)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character Management (6 attributes, classes, creation) | ✅ Achieved | No |
| Combat & Effects System | ✅ Achieved | No |
| World Management + Spatial Indexing | ✅ Achieved | No |
| Event-Driven Architecture | ✅ Achieved | No |
| WebSocket Real-time Communication | ✅ Achieved | Yes (library migration) |
| Health Monitoring & Prometheus Metrics | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Circuit Breakers & Resilience | ✅ Achieved | No |
| Input Validation Framework | ✅ Achieved | No |
| Asset Generation Pipeline | ⚠️ Partial (requires external AI) | No |
| Advanced NPC AI (A*, behavior trees) | ✅ Achieved | No |
| Spell System (levels 0-9) | ✅ Achieved | No |
| World Editor Tools | ⚠️ Partial (CLI only) | No |
| Network Optimization | ✅ Achieved (delta compression) | No |
| Player Progression Persistence | ✅ Achieved | No |
| Guild and Faction Systems | ✅ Achieved | No |
| Embedded Adventures (10 complete) | ✅ Achieved | No |
| Go stdlib security patches | ❌ Needs Go 1.24.13+ | Yes (toolchain upgrade) |

**Summary: 17/18 goals achieved, 1 security enhancement needed**

## Metrics Summary

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 30,736 | — |
| Total Functions | 596 + 1,694 methods | — |
| Total Packages | 18 | — |
| High Complexity (≥10) | 4 functions | ✅ Low risk |
| Duplication Ratio | 1.98% | ✅ Good (<3%) |
| Documentation Coverage | 88.7% | ✅ Good |
| Test Coverage | 79.6% | ✅ Exceeds 60% CI threshold |
| Circular Dependencies | 0 | ✅ Clean |

### Complexity Hotspots

| Function | File | Cyclomatic | Assessment |
|----------|------|------------|------------|
| `Validate` | `cmd/bootstrap-demo/main.go` | 10 | ⚠️ Edge of threshold |
| `promptRewards` | `cmd/quest-builder/main.go` | 10 | ⚠️ Edge of threshold |
| `RunCellularAutomata` | `pkg/pcg/terrain/cellular_automata.go` | 10 | ⚠️ Acceptable for PCG |
| `Stop` | `test/e2e/server.go` | 10 | ⚠️ Test infrastructure |

### Duplication Hotspots

| Clone Size | Location | Recommendation |
|------------|----------|----------------|
| 35 lines | `pkg/server/server.go` (internal) | Extract RPC handler registration pattern |
| 21 lines | `pkg/server/websocket.go`, `websocket_editor.go` | Extract WebSocket upgrade helper |
| 19 lines | `cmd/content-creator/main.go`, `quest-builder/main.go` | Extract CLI argument parsing pattern |
| 16 lines | `pkg/pcg/validator.go`, `pkg/validation/validation.go` | Extract validation helper to shared package |

### Package Coupling Analysis

| Package | Functions | Lines | Cohesion | Notes |
|---------|-----------|-------|----------|-------|
| `server` | 505 | 14,581 | 3.00 | Largest package, well-structured |
| `game` | 484 | 13,282 | 3.06 | Core mechanics, good cohesion |
| `pcg` | 529 | 13,249 | 5.66 | High cohesion, many utilities |
| `secrets` | 7 | 295 | 0.67 | Low cohesion—consider consolidation |
| `persistence` | 28 | 821 | 1.13 | Low cohesion—review structure |
| `integration` | 13 | 171 | 1.40 | Small utility package |

---

## Implementation Steps

### Step 1: Upgrade Go Toolchain to 1.24.13+

- **Deliverable**: Update `go.mod` to specify Go 1.24.13+ and verify all tests pass
- **Dependencies**: None
- **Goal Impact**: Resolves 18 known stdlib vulnerabilities in `crypto/tls`, `crypto/x509`, `net/http`, `net/url`, `html/template`, `os`
- **Acceptance**: 
  - `govulncheck ./...` reports 0 vulnerabilities
  - All tests pass with `go test -race ./...`
  - CI pipeline passes
- **Validation**: 
  ```bash
  go version  # Should show go1.24.13 or higher
  govulncheck ./... | grep -c "No vulnerabilities found"  # Should be 1
  ```

**Security Context**: Per CHANGELOG.md, 18 Go stdlib vulnerabilities require Go 1.24.12+ or Go 1.25.8. Research confirms Go 1.24.13 patches CVE-2025-68121 (crypto/tls auth bypass), CVE-2025-58189 (ALPN log injection), CVE-2025-61732 (cgo code injection), and multiple resource exhaustion issues.

---

### Step 2: Plan WebSocket Library Migration

- **Deliverable**: Create `docs/WEBSOCKET_MIGRATION.md` documenting migration path from `gorilla/websocket` to `nhooyr.io/websocket`
- **Dependencies**: None (documentation only)
- **Goal Impact**: Long-term maintainability—Gorilla WebSocket is archived and receives no updates since 2022
- **Acceptance**: 
  - Document covers API differences, migration steps, and rollback plan
  - Document identifies all affected files (6 files use websocket)
- **Validation**: 
  ```bash
  grep -r "gorilla/websocket" pkg/ cmd/ --include="*.go" -l | wc -l  # Baseline count
  test -f docs/WEBSOCKET_MIGRATION.md && echo "Migration doc exists"
  ```

**Affected Files**:
- `pkg/server/websocket.go` (primary WebSocket handling)
- `pkg/server/websocket_delta.go` (delta compression)
- `pkg/server/websocket_editor.go` (editor WebSocket)
- `test/e2e/client.go` (test client)

---

### Step 3: Extract Duplication in Server Package

- **Deliverable**: Refactor `pkg/server/server.go` to extract RPC handler registration pattern into helper function
- **Dependencies**: Step 1 (ensure tests pass baseline)
- **Goal Impact**: Reduces 35-line duplication block, improves maintainability
- **Acceptance**: 
  - Duplication ratio in `pkg/server/` decreases by ≥50 lines
  - All tests pass
  - No behavioral changes
- **Validation**: 
  ```bash
  go-stats-generator analyze ./pkg/server --skip-tests --format json | jq '.duplication.duplicated_lines'
  # Should be <100 (currently ~180)
  go test ./pkg/server/... -race
  ```

---

### Step 4: Consolidate WebSocket Helpers

- **Deliverable**: Extract shared WebSocket upgrade logic from `websocket.go` and `websocket_editor.go` into `pkg/server/websocket_common.go`
- **Dependencies**: Step 3
- **Goal Impact**: Reduces 21-line duplication, prepares for WebSocket migration
- **Acceptance**: 
  - Single WebSocket upgrader configuration
  - Tests pass for both regular and editor WebSocket endpoints
- **Validation**: 
  ```bash
  go test ./pkg/server/... -run TestWebSocket -race
  curl -I http://localhost:8080/ws  # After starting server
  ```

---

### Step 5: Consolidate CLI Argument Parsing

- **Deliverable**: Extract shared CLI parsing from `cmd/content-creator/` and `cmd/quest-builder/` into `pkg/cli/common.go`
- **Dependencies**: None
- **Goal Impact**: Reduces 19-line duplication across CLI tools
- **Acceptance**: 
  - Both CLI tools use shared parsing functions
  - CLI tools function identically (smoke test)
- **Validation**: 
  ```bash
  go build ./cmd/content-creator && ./content-creator --help
  go build ./cmd/quest-builder && ./quest-builder --help
  go-stats-generator analyze ./cmd --skip-tests --format json | jq '.duplication.duplicated_lines'
  ```

---

### Step 6: Review Low-Cohesion Packages

- **Deliverable**: Analyze `pkg/secrets/` (cohesion 0.67) and `pkg/persistence/` (cohesion 1.13); document whether consolidation is appropriate
- **Dependencies**: None
- **Goal Impact**: Code organization and maintainability
- **Acceptance**: 
  - Create `docs/PACKAGE_REVIEW.md` with analysis and recommendations
  - If consolidation is recommended, create TODO for follow-up
- **Validation**: 
  ```bash
  go-stats-generator analyze ./pkg/secrets ./pkg/persistence --skip-tests --format json \
    | jq '[.packages[] | {name: .name, cohesion: .cohesion_score}]'
  ```

---

### Step 7: Extract Shared Validation Helpers

- **Deliverable**: Move 16-line validation pattern duplicated between `pkg/pcg/validator.go` and `pkg/validation/validation.go` into shared location
- **Dependencies**: Step 6 (understand package structure)
- **Goal Impact**: Reduces duplication, single source of truth for validation patterns
- **Acceptance**: 
  - Both packages import shared validation helper
  - All PCG and validation tests pass
- **Validation**: 
  ```bash
  go test ./pkg/pcg/... ./pkg/validation/... -race
  go-stats-generator analyze ./pkg/pcg ./pkg/validation --skip-tests --format json \
    | jq '.duplication.clone_pairs'
  ```

---

### Step 8: Implement WebSocket Migration (Optional/Future)

- **Deliverable**: Migrate from `gorilla/websocket` to `nhooyr.io/websocket` using plan from Step 2
- **Dependencies**: Steps 2, 4 (migration plan and consolidated helpers)
- **Goal Impact**: Long-term security and maintainability
- **Acceptance**: 
  - All E2E WebSocket tests pass
  - Delta compression functionality preserved
  - Performance benchmarks within 10% of baseline
- **Validation**: 
  ```bash
  go test ./test/e2e/... -race -v
  go test ./pkg/server/... -bench=. -benchmem
  grep -r "gorilla/websocket" . --include="*.go" -l  # Should return empty
  ```

---

## Scope Assessment

Using project-calibrated thresholds:

| Metric | Value | Threshold | Assessment |
|--------|-------|-----------|------------|
| Functions above complexity 10 | 4 | <5 = Small | ✅ Small scope |
| Duplication ratio | 1.98% | <3% = Small | ✅ Small scope |
| Doc coverage gap | 11.3% | 10–25% = Medium | ⚠️ Medium scope |
| Security vulnerabilities | 18 stdlib | >0 = Priority | 🔴 Critical priority |

**Overall Scope: Medium** — Primary driver is security (Go toolchain upgrade), with maintainability improvements as secondary goals.

---

## Priority Ordering

1. **Step 1** (Go Toolchain) — Security-critical, blocks production deployment confidence
2. **Step 2** (Migration Plan) — Documentation only, low risk, enables future work
3. **Steps 3–5** (Duplication) — Maintainability, independent of each other, can parallelize
4. **Steps 6–7** (Package Review) — Lower priority, optional consolidation
5. **Step 8** (WebSocket Migration) — Optional enhancement for Q2 2026

---

## Appendix: Raw Metrics

### High-Complexity Function Details

```
10.0: Validate                  cmd/bootstrap-demo/main.go:44 lines
10.0: promptRewards             cmd/quest-builder/main.go:54 lines  
10.0: RunCellularAutomata       pkg/pcg/terrain/cellular_automata.go:40 lines
10.0: Stop                      test/e2e/server.go:42 lines
```

### Documentation Coverage by Package

```
packages coverage:  83.3%
functions coverage: 94.0%
overall coverage:   88.7%
```

### Test Coverage by Package

```
pkg/game:       85.7%
pkg/server:     70.8%
pkg/validation: 92.0%
pkg/wasmui:     94.1%
pkg/pcg:        78.2%
total:          79.6%
```

---

## Validation Commands

```bash
# Verify all steps completed
go version                                    # Step 1: Go 1.24.13+
govulncheck ./...                             # Step 1: No vulnerabilities
test -f docs/WEBSOCKET_MIGRATION.md           # Step 2: Migration doc exists

# Verify metrics improved
go-stats-generator analyze . --skip-tests --sections duplication --format json \
  | jq '.duplication.duplication_ratio'       # Should be <1.5%

# Verify no regressions
go test -race ./...                           # All tests pass
make test-coverage                            # Coverage ≥79%
```

---

*Generated by go-stats-generator analysis combined with project goal assessment*
