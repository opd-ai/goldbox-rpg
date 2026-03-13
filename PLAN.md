# Implementation Plan: Content Creation Tooling Maturity

## Project Context
- **What it does**: A modern, Go-based RPG engine inspired by SSI Gold Box series with character management, combat systems, and world interactions through a JSON-RPC API with WebSocket support.
- **Current goal**: Complete asset generation pipeline and mature CLI content creation tools
- **Estimated Scope**: Medium

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Core RPG mechanics and character system | ✅ Achieved | No |
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
| Embedded Adventures (10 packs) | ✅ Achieved | No |
| Asset generation pipeline (521 assets) | ⚠️ Partial (252/521) | Yes |
| World editor tools | ⚠️ Partial (CLI only) | Yes |
| Content creation utilities | ⚠️ Partial (low coverage) | Yes |

**Summary**: 14/17 goals fully achieved, 3/17 partial

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 5 functions at complexity 10 (threshold 9)
  - `terrain.RunCellularAutomata` (pkg/pcg/terrain/cellular_automata.go:56)
  - `main.Validate` (cmd/bootstrap-demo/main.go:96)
  - `main.run` (cmd/quest-builder/main.go:102)
  - `main.promptRewards` (cmd/quest-builder/main.go:399)
  - `e2e.Stop` (test/e2e/server.go:164)
- **Duplication ratio**: 1.52% (37 clone pairs, 938 duplicated lines)
- **Doc coverage**: 88.65% overall, 94.07% function coverage
- **Total code coverage**: 79.1% (threshold: 60%)
- **Package coupling concerns**: 
  - `pkg/cliutil`: 12.2% coverage (critical for content creation)
  - `cmd/quest-builder`: 27.6% coverage
  - `cmd/content-creator`: 37.0% coverage
  - `pkg/server`: 70.5% coverage (10% below game package)

## Implementation Steps

### Step 1: Improve CLI Utility Test Coverage
- **Deliverable**: Add unit tests for `pkg/cliutil/preview.go` (9 functions, 126 LOC)
- **Dependencies**: None
- **Goal Impact**: Advances "Content creation utilities" — preview server is core infrastructure for visual content validation
- **Acceptance**: `pkg/cliutil` coverage reaches ≥70% (currently 12.2%)
- **Validation**: 
  ```bash
  go test ./pkg/cliutil/... -coverprofile=c.out && go tool cover -func=c.out | grep cliutil
  ```

### Step 2: Add Quest Builder Test Suite
- **Deliverable**: Table-driven tests for `cmd/quest-builder/main.go` covering:
  - Command parsing and validation
  - Quest template loading
  - Reward prompt handling (`promptRewards` complexity 10)
  - Output file generation
- **Dependencies**: Step 1 (shares test patterns)
- **Goal Impact**: Advances "Content creation utilities" — quest builder is primary tool for non-technical content creators
- **Acceptance**: `cmd/quest-builder` coverage reaches ≥60% (currently 27.6%)
- **Validation**:
  ```bash
  go test ./cmd/quest-builder/... -coverprofile=c.out && go tool cover -func=c.out | grep quest-builder
  ```

### Step 3: Add Content Creator Integration Tests
- **Deliverable**: Integration tests for `cmd/content-creator/main.go` validating:
  - Template-based content generation
  - Output schema compliance
  - Error handling for invalid inputs
- **Dependencies**: Step 1
- **Goal Impact**: Advances "Content creation utilities" — validates PCG content meets game schema requirements
- **Acceptance**: `cmd/content-creator` coverage reaches ≥60% (currently 37.0%)
- **Validation**:
  ```bash
  go test ./cmd/content-creator/... -coverprofile=c.out && go tool cover -func=c.out | grep content-creator
  ```

### Step 4: Add Asset Download Automation
- **Deliverable**: 
  - New Makefile target `assets-download` to fetch pre-generated asset pack
  - Update `ASSET_INTEGRATION.md` with simplified quick-start using asset pack
  - GitHub Release asset with pre-generated 521-asset pack
- **Dependencies**: None (parallel with Steps 1-3)
- **Goal Impact**: Directly closes "Asset generation pipeline" gap — eliminates 4-6 hour AI generation requirement
- **Acceptance**: `make assets-verify` reports 521/521 after `make assets-download`
- **Validation**:
  ```bash
  make assets-download && make assets-verify
  ```

### Step 5: Improve Server Package Test Coverage
- **Deliverable**: Add tests for edge cases in `pkg/server/`:
  - Session timeout and cleanup (`session.go`)
  - WebSocket reconnection error handling (`websocket_nhooyr.go`)
  - Rate limiting boundary conditions (`websocket.go`)
- **Dependencies**: Steps 1-3 completed (pattern consistency)
- **Goal Impact**: Strengthens production reliability — server is the runtime for all content
- **Acceptance**: `pkg/server` coverage reaches ≥75% (currently 70.5%)
- **Validation**:
  ```bash
  go test ./pkg/server/... -coverprofile=c.out && go tool cover -func=c.out | grep server
  ```

### Step 6: Refactor High-Complexity Quest Builder Function
- **Deliverable**: Refactor `main.promptRewards()` in `cmd/quest-builder/main.go:399` (complexity 10) into smaller functions:
  - `parseRewardInput()` for input validation
  - `buildRewardStruct()` for data construction
  - `validateReward()` for constraint checking
- **Dependencies**: Step 2 (tests must exist before refactoring)
- **Goal Impact**: Advances "Content creation utilities" — reduces cognitive load for maintainers
- **Acceptance**: `promptRewards` complexity drops to ≤6, no test regressions
- **Validation**:
  ```bash
  go-stats-generator analyze ./cmd/quest-builder --format json --sections functions | jq '.functions[] | select(.name == "promptRewards") | .complexity.cyclomatic'
  ```

### Step 7: Visual Editor Foundation (Optional Enhancement)
- **Deliverable**: Extend `pkg/wasmui/editor.go` (537 lines) with:
  - Tile palette selection UI
  - Click-to-place tile interaction
  - Save/load via existing WebSocket editor protocol
- **Dependencies**: Steps 1-6 completed
- **Goal Impact**: Advances "World editor tools" — bridges gap from CLI-only to visual editors
- **Acceptance**: User can place tiles at `/editor` route without CLI
- **Validation**: Manual verification — load editor, place tiles, verify save/load cycle

## Scope Assessment Details

| Metric | Value | Threshold | Assessment |
|--------|-------|-----------|------------|
| Functions above complexity 9 | 5 | <5 Small, 5-15 Medium | Medium |
| Duplication ratio | 1.52% | <3% Small | Small ✅ |
| Doc coverage gap | 11.35% | <10% Small, 10-25% Medium | Medium |
| Test coverage gap (CLI tools) | 45-88% | Varies | Medium |

**Overall Scope**: Medium — 5-15 functions need attention, coverage gaps in specific packages

## Research Findings

### Dependency Security
- **Go 1.25.6+**: Project already on secure Go version (go.mod shows 1.25.6)
- **nhooyr.io/websocket**: No critical CVEs for 2025-2026; known DoS via double pong (mitigated by rate limiting)
- **gorilla/websocket**: Archived but acceptable for test-only usage (documented in go.mod)
- **Ebitengine v2.7**: Deprecated APIs present (`DeviceScaleFactor`, `Image.Dispose`) — future migration advisable

### Community Context
- No critical open issues on GitHub
- Bug hunt documentation exists (`BUG_HUNT_REPORT_2025.md`)
- Active maintenance with recent commits

---

*Generated: 2026-03-13 by go-stats-generator + project analysis*
