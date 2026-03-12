# Implementation Plan: GUI World Editor Development

## Project Context
- **What it does**: Modern Go-based RPG engine for turn-based games inspired by classic SSI Gold Box series, providing JSON-RPC API with WebSocket support for real-time gameplay
- **Current goal**: GUI World Editor — the most significant unachieved enhancement from the project roadmap
- **Estimated Scope**: Large (cross-cutting enhancement across multiple packages)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character Management (6 attributes, classes) | ✅ Achieved | No |
| Combat & Effects System | ✅ Achieved | No |
| World Management + Spatial Indexing | ✅ Achieved | No |
| Event-Driven Architecture | ✅ Achieved | No |
| WebSocket Real-time Communication | ✅ Achieved | No |
| Health Monitoring & Prometheus Metrics | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Circuit Breakers & Resilience | ✅ Achieved | No |
| Input Validation Framework | ✅ Achieved | No |
| Asset Generation Pipeline | ✅ Achieved (252/252 placeholders) | No |
| Advanced NPC AI (A*, behavior trees) | ✅ Achieved | No |
| Combat Mechanics (opportunity attacks, flanking) | ✅ Achieved | No |
| Spell System (levels 0-9) | ✅ Achieved (60 spells) | No |
| Guild & Faction Systems | ✅ Achieved | No |
| World Editor Tools | ⚠️ Partial (CLI only) | **Yes** |
| Network Optimization | ✅ Achieved (delta compression) | No |
| Content Creation Utilities | ⚠️ Partial (CLI only) | Prerequisite |

## Metrics Summary

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 29,215 | — |
| Total Functions/Methods | 545 + 1,574 | — |
| Total Packages | 18 | — |
| Test Coverage | 77.9% | ✅ Above 60% CI threshold |
| High Complexity Functions (>10) | 4 | ✅ Low risk |
| Duplication Instances | 52 clones | ⚠️ Moderate |

### Complexity Hotspots (functions >10 cyclomatic complexity)
| Function | File | Complexity |
|----------|------|------------|
| `drawRect` | `cmd/map-editor/main.go` | 12 |
| `CalculateDelta` | `pkg/server/websocket_delta.go` | 12 |
| `validateQuest` | `cmd/quest-builder/main.go` | 11 |
| `interactiveEdit` | `cmd/map-editor/main.go` | 11 |

### Coverage Gaps on Goal-Critical Paths
| Package | Coverage | Impact on GUI Editor |
|---------|----------|---------------------|
| `pkg/server` | 70.5% | WebSocket handlers for editor |
| `scripts` | 68.8% | Asset validation utilities |
| `test/e2e` | 65.3% | Editor integration tests |

## Implementation Steps

### Step 1: Restore Test Coverage to CI Threshold
- **Deliverable**: Ensure overall test coverage remains ≥60%
- **Dependencies**: None
- **Goal Impact**: Prerequisite — CI must pass before any new features
- **Acceptance**: `go test ./... -coverprofile=c.out && go tool cover -func=c.out | grep total` shows ≥60%
- **Files**:
  - `pkg/server/handlers_test.go` — add tests for uncovered RPC handlers
  - `pkg/server/websocket_delta_test.go` — improve delta compression coverage
- **Validation**: `go test ./... -short && go tool cover -func=coverage.out | grep total`

### Step 2: Extract Editor-Specific RPC Methods
- **Deliverable**: New RPC methods in `pkg/server/handlers_editor.go` for map CRUD operations
- **Dependencies**: Step 1 (coverage must pass CI)
- **Goal Impact**: Foundation for browser-based editor — separates editor concerns from gameplay
- **Acceptance**: 
  - `Editor.CreateMap`, `Editor.UpdateTile`, `Editor.SaveMap`, `Editor.LoadMap` methods implemented
  - Each method has corresponding test with ≥60% coverage
- **Files**:
  - `pkg/server/handlers_editor.go` — new file with editor-specific handlers
  - `pkg/server/handlers_editor_test.go` — test coverage
  - `pkg/server/constants.go` — add new method constants
- **Validation**: `go test ./pkg/server/... -run TestEditor -v`

### Step 3: Create WebSocket Editor Protocol
- **Deliverable**: Bidirectional WebSocket protocol for real-time tile updates
- **Dependencies**: Step 2 (editor RPC methods exist)
- **Goal Impact**: Enables live preview while editing — core editor UX requirement
- **Acceptance**:
  - Editor clients receive tile change broadcasts
  - Changes persist to YAML within 100ms latency
- **Files**:
  - `pkg/server/websocket_editor.go` — editor-specific WebSocket handler
  - `pkg/server/websocket_editor_test.go` — protocol tests
- **Validation**: `go test ./pkg/server/... -run TestWebSocketEditor -v`

### Step 4: Implement Browser-Based Map Editor Frontend
- **Deliverable**: WASM-based map editor using existing Ebitengine infrastructure
- **Dependencies**: Step 3 (WebSocket protocol ready)
- **Goal Impact**: Primary deliverable — user can create maps without CLI
- **Acceptance**:
  - User can create, load, edit, and save maps via browser
  - All terrain types selectable from palette
  - Grid-based tile placement with click/drag
- **Files**:
  - `pkg/wasmui/editor.go` — editor UI implementation
  - `pkg/wasmui/editor_state.go` — editor state management
  - `cmd/wasm-editor/main.go` — WASM entry point for editor
  - `web/editor.html` — editor HTML shell
- **Validation**: `make wasm-editor && curl -s http://localhost:8080/editor.html | grep -q 'editor'`

### Step 5: Add Quest Builder Visual Interface
- **Deliverable**: Browser UI for quest creation using existing quest schema
- **Dependencies**: Step 4 (map editor provides UI patterns)
- **Goal Impact**: Completes content creation utilities goal
- **Acceptance**:
  - Quest objectives can be added/edited visually
  - Quest rewards configurable via dropdowns
  - Validation feedback shown inline
- **Files**:
  - `pkg/wasmui/quest_editor.go` — quest editor UI
  - `pkg/server/handlers_quest_editor.go` — quest CRUD RPC methods
  - `web/quest-builder.html` — quest builder HTML shell
- **Validation**: `go test ./pkg/server/... -run TestQuestEditor -v`

### Step 6: Reduce Code Duplication in Validation Package
- **Deliverable**: Extract common validation patterns into shared helpers
- **Dependencies**: None (can run in parallel with Steps 1-5)
- **Goal Impact**: Maintainability — duplication creates bug surface area
- **Acceptance**: 
  - Duplication in `pkg/validation/validation.go` reduced by extracting helper functions
  - No functional changes (tests still pass)
- **Files**:
  - `pkg/validation/validation.go` — refactor duplicate code blocks (lines 117-188)
  - `pkg/validation/helpers.go` — new file for extracted helpers
- **Validation**: `go-stats-generator analyze ./pkg/validation --sections duplication --format json | jq '.duplication.clones | length'` shows reduced count

### Step 7: Integration Testing for Editor Workflow
- **Deliverable**: E2E tests covering complete editor workflow
- **Dependencies**: Steps 4-5 (editor functionality complete)
- **Goal Impact**: Confidence in editor stability for users
- **Acceptance**:
  - Test creates map, edits tiles, saves, reloads, verifies
  - Test creates quest, configures objectives, validates
- **Files**:
  - `test/e2e/editor_test.go` — editor integration tests
- **Validation**: `go test ./test/e2e/... -run TestEditor -v`

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| WASM editor performance | Use existing Ebitengine patterns from `pkg/wasmui/game.go` |
| WebSocket complexity | Build on proven `pkg/server/websocket.go` infrastructure |
| Test coverage regression | Run `make test-coverage` before each merge |
| Breaking existing CLI tools | Additive changes only — don't modify `cmd/map-editor/` |

## Dependencies (Pending PRs)

The following open Dependabot PRs should be considered but are **not blockers**:

| PR | Description | Priority |
|----|-------------|----------|
| #15 | Bump golang Docker image to 1.25 | Medium — security improvements |
| #14 | Bump Go dependencies (prometheus, testify, time) | Low — compatible updates |
| #12 | Bump golangci-lint-action to v8 | Low — CI enhancement |

## Success Criteria

1. **Functional**: User can create a playable map without using command line
2. **Performance**: Editor operations complete in <100ms
3. **Quality**: Test coverage remains ≥60%
4. **Compatibility**: Existing CLI tools continue to work unchanged

---

*Generated: 2026-03-12*  
*Analysis Tool: go-stats-generator v1.0.0*  
*Files Analyzed: 174 Go files, 29,215 lines of code*
