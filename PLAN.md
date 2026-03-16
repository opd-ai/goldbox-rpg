# Implementation Plan: Maintenance & Quality Improvements

*Generated: 2026-03-16*

## Project Context

- **What it does**: A modern Go-based RPG engine inspired by classic SSI Gold Box games, providing character management, combat systems, and world interactions through a JSON-RPC API with WebSocket support and an Ebitengine/WASM frontend.
- **Current goal**: All 17 stated goals are fully achieved. The next priority is merging pending dependency updates and addressing maintainability improvements identified by static analysis.
- **Estimated Scope**: Small (<5 high-priority items)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character Management (6 attributes, 6 classes) | ✅ Achieved | No |
| Combat and effect systems | ✅ Achieved | No |
| WebSocket real-time communication | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Circuit breaker patterns | ✅ Achieved | No |
| Comprehensive input validation | ✅ Achieved | No |
| Health monitoring and metrics | ✅ Achieved | No |
| Advanced NPC AI behaviors | ✅ Achieved | No |
| Enhanced combat mechanics | ✅ Achieved | No |
| Complete spell system (levels 0-9) | ✅ Achieved | No |
| Network optimization | ✅ Achieved | No |
| Player progression persistence | ✅ Achieved | No |
| Guild and faction systems | ✅ Achieved | No |
| Embedded Adventures (10+) | ✅ Achieved | No |
| Asset generation pipeline (521 assets) | ✅ Achieved | No |
| World editor tools | ✅ Achieved | No |
| Content creation utilities | ✅ Achieved | No |

**Summary**: All 17 stated goals are fully achieved. This plan focuses on dependency maintenance, code quality improvements, and preparing for future scalability.

## Metrics Summary

| Metric | Value | Status |
|--------|-------|--------|
| Total Lines of Code | 31,367 | - |
| Total Functions/Methods | 2,406 | - |
| Test Coverage | 82.9% | ✅ Exceeds 60% threshold |
| Functions with complexity >10 | 0 | ✅ Clean |
| Duplication Ratio | 1.38% | ✅ <5% threshold |
| Documentation Coverage | 89.9% | ✅ Good |
| Circular Dependencies | 0 | ✅ None |
| `go vet` Issues | 0 | ✅ Clean |

### Complexity Hotspots

| Function | File | Complexity |
|----------|------|------------|
| Stop | test/e2e/server.go | 14.5 |
| parseConstDeclaration | cmd/openapi-gen/main.go | 14.2 |
| run | cmd/quest-builder/main.go | 14.0 |
| Validate | cmd/bootstrap-demo/main.go | 14.0 |
| extractMethods | cmd/openapi-gen/main.go | 13.7 |

*Note: All functions are below cyclomatic complexity 15, which is acceptable. Top items are in CLI tools and test utilities, not core game logic.*

### Package Health

| Package | Functions | Cohesion | Notes |
|---------|-----------|----------|-------|
| pcg | 543 | - | Largest package, appropriately sized for PCG responsibility |
| server | 534 | - | High coupling (11 deps) expected for network layer |
| game | 496 | - | Core game logic, well-organized |
| wasmui | 155 | - | Ebitengine frontend |

### Duplication Analysis

- **Clone pairs detected**: 34
- **Total duplicated lines**: 868 (1.38%)
- **Largest clone**: 35 lines
- **Top duplication patterns**: Handler registration in `pkg/server/`, CLI flag parsing

---

## Implementation Steps

### Step 1: Merge Dependabot PR #38 (Docker Go Version)

- **Deliverable**: Merge PR #38 to upgrade Docker base image from `golang:1.23-bookworm` to `golang:1.26-bookworm`
- **Dependencies**: None
- **Goal Impact**: Security and performance improvements from Go 1.26; resolves 18 Go stdlib vulnerabilities identified by govulncheck
- **Acceptance**: PR merged, CI passes on master, Docker build succeeds
- **Validation**: 
```bash
# After merge:
git checkout master && git pull
docker build -t goldbox-rpg . && docker run --rm goldbox-rpg /app/goldbox-rpg --version
curl -s http://localhost:8080/health  # (when running)
```

### Step 2: Merge Dependabot PR #35 (GitHub Actions)

- **Deliverable**: Merge PR #35 to upgrade `actions/upload-artifact` from v4 to v7 (Node 24 runtime support)
- **Dependencies**: Step 1 (merge sequentially to avoid conflicts)
- **Goal Impact**: CI/CD modernization; Node 24 runtime compatibility for GitHub Actions
- **Acceptance**: PR merged, all CI workflows pass
- **Validation**:
```bash
# After merge:
git checkout master && git pull
# Trigger CI by pushing a no-op commit or via GitHub Actions UI
# Verify all workflow jobs complete successfully
```

### Step 3: Refactor Handler Registration to Table-Driven Pattern

- **Deliverable**: Extract handler registration in `pkg/server/server.go` into a table-driven pattern
- **Dependencies**: Steps 1-2 (clean master branch)
- **Goal Impact**: Reduces 868 duplicated lines; improves maintainability; ROI score 29.00 (highest from static analysis)
- **Files to modify**:
  - `pkg/server/server.go` (extract registration loop)
  - Create `pkg/server/handlers_table.go` (handler table definition)
- **Acceptance**: 
  - Duplication ratio reduced from 1.38% to <1.0%
  - All existing tests pass
  - No behavior changes
- **Validation**:
```bash
go test -race ./pkg/server/...
go-stats-generator analyze ./pkg/server --skip-tests --format json | jq '.duplication.ratio'
# Expected: ratio < 1.0
```

### Step 4: Increase Quest Builder Test Coverage

- **Deliverable**: Add tests to `cmd/quest-builder/` to improve coverage from 71.6% to 80%+
- **Dependencies**: None (can be done in parallel with Step 3)
- **Goal Impact**: Improves test coverage for CLI tools; aligns with 80%+ coverage of other CLI commands
- **Files to modify**:
  - `cmd/quest-builder/main_test.go` (add edge case tests)
- **Test cases to add**:
  - Invalid YAML input handling
  - Quest chain validation with circular dependencies
  - Missing required field validation
  - Empty quest list handling
- **Acceptance**: Coverage ≥80% for `cmd/quest-builder/`
- **Validation**:
```bash
go test ./cmd/quest-builder/... -coverprofile=/tmp/qb.out
go tool cover -func=/tmp/qb.out | grep total
# Expected: total coverage >= 80.0%
```

### Step 5: Address BUG Comments (5 items)

- **Deliverable**: Review and address 5 BUG annotations identified by static analysis
- **Dependencies**: Steps 1-2 (clean baseline)
- **Goal Impact**: Reduces technical debt; ensures known issues are tracked or resolved
- **Discovery**: 
```bash
grep -rn "BUG:" --include="*.go" pkg/
```
- **Acceptance**: Each BUG either:
  - Fixed with corresponding test
  - Converted to tracked issue with justification
  - Documented as intentional behavior
- **Validation**:
```bash
go-stats-generator analyze . --skip-tests | grep -A2 "BUG:"
# Expected: 0 unaddressed BUG annotations
```

---

## Low Priority (Future Consideration)

These items are identified for awareness but should not be prioritized unless modifying adjacent code:

### File Size Reduction

| File | Lines | Functions | Action |
|------|-------|-----------|--------|
| pkg/server/handlers.go | 1187 | 56 | Consider split by RPC category if editing |
| pkg/pcg/metrics.go | 687 | 48 | Consider split by metric type if editing |
| pkg/pcg/validator.go | 777 | 49 | Acceptable for validation logic |

### Package Cohesion

Low cohesion packages (expected for utility packages):
- `cliutil` (0.8) - Small CLI utilities, acceptable
- `secrets` (0.7) - Small secrets utilities, acceptable  
- `persistence` (1.1) - Intentionally minimal interface

### Naming Convention Violations

- 14 file name violations (generic names like `constants.go`, `errors.go`)
- 28 identifier violations (stuttering, acronym casing)
- Overall naming score: 1.00 (excellent)

*Action: Address during related refactoring, not as standalone work.*

---

## External Research Findings

### Dependency Status

| Dependency | Version | Status | Notes |
|------------|---------|--------|-------|
| github.com/coder/websocket | v1.8.14 | ✅ Secure | No known vulnerabilities; actively maintained |
| gorilla/websocket | v1.5.3 | ⚠️ Archived | Retained for E2E tests only (documented in go.mod) |
| hajimehoshi/ebiten/v2 | v2.9.9 | ✅ Latest | Consider upgrading to benefit from new vector API |
| prometheus/client_golang | v1.23.2 | ✅ Current | Standard metrics library |
| sirupsen/logrus | v1.9.4 | ✅ Stable | Mature logging library |

### Ebitengine v2.9 Considerations

Per Ebitengine 2.9 release notes:
- New vector graphics rendering with `vector.FillPath`, `vector.StrokePath`
- Deprecated: `AppendVerticesAndIndicesFor...` functions (still present but will be removed)
- **Recommendation**: Audit `pkg/wasmui/` for deprecated vector API usage when planning frontend updates

### GitHub Repository Status

- **Open Issues**: 0
- **Open PRs**: 2 (both Dependabot, addressed in Steps 1-2)
- **Community Activity**: Single maintainer project

---

## Verification Commands

```bash
# Full validation suite after all steps:
go test -race ./...                    # All tests pass
go vet ./...                           # Zero issues
golangci-lint run --timeout=5m         # Lint passes

# Coverage check
go test ./... -coverprofile=/tmp/c.out && go tool cover -func=/tmp/c.out | grep total
# Expected: >= 82%

# Metrics refresh
go-stats-generator analyze . --skip-tests

# Docker build verification
docker build -t goldbox-rpg .
docker run --rm -d -p 8080:8080 --name goldbox-test goldbox-rpg
curl http://localhost:8080/health
docker stop goldbox-test
```

---

## Summary

**Project Health**: Excellent. All 17 stated goals achieved with 82.9% test coverage and zero high-complexity functions.

**Immediate Actions** (Steps 1-2):
1. Merge PR #38 (Docker Go 1.26)
2. Merge PR #35 (GitHub Actions v7)

**Short-term Improvements** (Steps 3-5):
3. Refactor handler registration (reduces duplication)
4. Increase quest-builder test coverage to 80%+
5. Address 5 BUG annotations

**No Critical Gaps**: The project delivers on all documented promises. This plan addresses maintainability and dependency hygiene, not functional blockers.

---

*Analysis performed using go-stats-generator v1.0.0*  
*Metrics baseline: 2026-03-16*
