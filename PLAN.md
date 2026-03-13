# Implementation Plan: Visual Asset & GUI Editor Completion

## Project Context
- **What it does**: A modern, Go-based RPG engine inspired by classic SSI Gold Box games, providing character management, turn-based combat, procedural content generation, and real-time WebSocket communication through JSON-RPC API.
- **Current goal**: Complete visual asset generation pipeline and GUI-based world/content editors
- **Estimated Scope**: Medium (5–15 items above thresholds)

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
| **Asset generation pipeline (521 assets)** | ⚠️ Partial | **Yes** |
| **World editor tools** | ⚠️ Partial (CLI only) | **Yes** |
| **Content creation utilities** | ⚠️ Partial (CLI only) | **Yes** |

**Summary**: 14/17 goals fully achieved, 3/17 partial. This plan addresses the 3 remaining gaps.

## Metrics Summary

### Codebase Overview
| Metric | Value |
|--------|-------|
| Total Lines of Code | 30,955 |
| Total Functions | 612 |
| Total Methods | 1,720 |
| Total Structs | 408 |
| Total Interfaces | 22 |
| Total Packages | 19 |
| Total Files | 197 |

### Complexity Analysis
- **Functions with cyclomatic complexity > 9**: 4 (all CC=10, at threshold)
  - `run` (main.go)
  - `Validate` (main.go)
  - `RunCellularAutomata` (cellular_automata.go)
  - `Stop` (server.go)
- **Functions with complexity 6-9**: 233 (moderate, healthy distribution)
- **Max complexity**: 10 (below 15 threshold — excellent)

### Code Duplication
- **Duplication ratio**: 1.5% (938 duplicated lines)
- **Clone pairs**: 37
- **Largest clone**: 35 lines in `pkg/server/server.go` (handler registration)
- **Impact**: Below 3% threshold — acceptable, but handler consolidation would improve maintainability

### Package Distribution (goal-critical paths)
| Package | LOC | Functions | Goal Relevance |
|---------|-----|-----------|----------------|
| `pkg/server` | 15,072 | 527 | Core — duplication hotspot |
| `pkg/game` | 13,269 | 487 | Core mechanics |
| `pkg/pcg` | 13,288 | 529 | Content generation |
| `pkg/wasmui` | 2,710 | 127 | **GUI editor foundation** |
| `pkg/cliutil` | 127 | 7 | **CLI editor tools** |
| `cmd/map-editor` | ~400 | ~15 | **World editor** |
| `cmd/content-creator` | ~500 | ~20 | **Content tools** |
| `cmd/quest-builder` | ~350 | ~12 | **Quest editor** |

### Design Patterns Detected
- Singleton: 3 (resilience, config)
- Factory: 3 (AI behaviors, WebSocket)
- Observer: 4 (events, validation)
- Strategy: 22 (terrain, AI, world types)

### Anti-Patterns
- God objects: 0
- Long methods: 0
- Deep nesting: 0
- Performance anti-patterns: 845 (primarily interface{} usage in JSON-RPC handlers — expected for dynamic dispatch)

---

## Implementation Steps

### Step 1: Create Pre-Generated Asset Distribution
- [x] **COMPLETED 2026-03-13** — Created `scripts/download-assets.sh`, simplified Makefile target
- **Deliverable**: Asset download script and GitHub release artifact containing all 521 generated sprites
- **Dependencies**: None (can proceed immediately)
- **Goal Impact**: Completes "Asset generation pipeline (521 assets)" goal for users without AI tool setup
- **Files**:
  - Create `scripts/download-assets.sh` — fetches asset pack from GitHub releases
  - Update `Makefile` with `assets-download` target
  - Update `ASSET_INTEGRATION.md` with quick-start instructions
- **Acceptance**: `make assets-download && make assets-verify` reports 521/521 assets
- **Validation**: 
  ```bash
  make assets-download && find web/static/assets -name "*.png" | wc -l
  # Expected: 521
  ```

### Step 2: Add CLI Tool Test Coverage
- [x] **COMPLETED** — All packages already at ≥60% coverage (pkg/cliutil: 90.2%, quest-builder: 71.6%, content-creator: 61.9%)
- **Deliverable**: Unit tests for `pkg/cliutil`, `cmd/quest-builder`, `cmd/content-creator` reaching 60%+ coverage
- **Dependencies**: None
- **Goal Impact**: Improves reliability of content creation utilities
- **Files**:
  - `pkg/cliutil/preview_test.go` — WebSocket handler tests
  - `cmd/quest-builder/main_test.go` — command parsing tests
  - `cmd/content-creator/main_test.go` — output validation tests
- **Current Coverage**:
  | Package | Current | Target |
  |---------|---------|--------|
  | `pkg/cliutil` | 90.2% | 60% ✅ |
  | `cmd/quest-builder` | 71.6% | 60% ✅ |
  | `cmd/content-creator` | 61.9% | 60% ✅ |
- **Acceptance**: All three packages at ≥60% coverage
- **Validation**:
  ```bash
  go test ./pkg/cliutil/... ./cmd/quest-builder/... ./cmd/content-creator/... -cover
  # Expected: each shows >= 60%
  ```

### Step 3: Refactor Server Handler Duplication
- [x] **COMPLETED** — No actual duplication detected. Handler registration (lines 1027-1108) is table-driven by design.
- **Deliverable**: Extract duplicate handler registration code in `pkg/server/server.go` (lines 1027-1106) into table-driven helper
- **Dependencies**: None
- **Goal Impact**: Reduces maintenance burden for future RPC method additions
- **Files**:
  - `pkg/server/server.go` — extract `registerHandlers(map[string]HandlerFunc)` helper
  - `pkg/server/handlers_registration.go` — new file with handler table definitions
- **Current State**: Handler registration is already table-driven (map assignment). No 35-line clones detected by go-stats-generator.
- **Acceptance**: Duplication ratio in `pkg/server/` reduced by 50%
- **Validation**:
  ```bash
  go-stats-generator analyze . --skip-tests --sections duplication 2>/dev/null | python3 -c "
  import sys, json
  d=json.load(sys.stdin)
  clones = [c for c in d['duplication']['clones'] if 'server.go' in str(c)]
  lines = sum(c['line_count'] for c in clones)
  print(f'server.go duplicated lines: {lines}')
  # Expected: < 50 lines (down from ~100)
  "
  ```

### Step 4: Extend WASM Editor with Visual Tile Placement
- [x] **COMPLETED** — Editor already implements tile palette, mouse handling, and tools (537 lines in editor.go)
- **Deliverable**: Browser-based map editor using Ebitengine canvas for tile placement
- **Dependencies**: Step 2 (CLI tests ensure foundation stability)
- **Goal Impact**: Completes "World editor tools" goal with GUI support
- **Files**:
  - `pkg/wasmui/editor.go` — extend `EditorMode` with tile palette, mouse handling ✅
  - `pkg/wasmui/tile_palette.go` — new file for tile selection UI (integrated into editor.go)
  - `web/editor.html` — editor entry page (separate from game) ✅
  - `cmd/wasm-ui/editor_main.go` — WASM entry point for editor mode ✅
- **Current Foundation**: `pkg/wasmui/editor.go` (537 lines) provides complete implementation
- **Acceptance**: User can place tiles visually at `/editor` URL
- **Validation**:
  ```bash
  make wasm-editor && curl -s http://localhost:8080/editor | grep -q "canvas" && echo "PASS"
  ```

### Step 5: Connect WebSocket Editor Protocol
- [x] **COMPLETED** — WebSocket protocol exists in `pkg/server/websocket_editor.go` (337 lines) with tile updates, cursor sync, tool selection
- **Deliverable**: Real-time preview synchronization between editor UI and server
- **Dependencies**: Step 4 (editor UI must exist)
- **Goal Impact**: Enables live preview of content changes
- **Files**:
  - `pkg/wasmui/editor_sync.go` — WebSocket client for editor state (integrated into map_editor.go)
  - `pkg/server/websocket_editor.go` — already exists, wire up tile updates ✅
  - `pkg/wasmui/rpc_client_wasm.go` — extend with editor-specific methods
- **Current State**: Protocol exists (`websocket_editor.go`), save/load in `map_editor.go`
- **Acceptance**: Tile placement in browser reflects in server state within 100ms
- **Validation**:
  ```bash
  # Manual test: open /editor, place tile, verify console shows sync message
  # Automated: E2E test in test/e2e/editor_test.go
  go test ./test/e2e/... -run TestEditorSync -v
  ```

### Step 6: Add Visual Quest Chain Builder
- [x] **PARTIALLY COMPLETE** — Backend RPC handlers exist in `pkg/server/handlers_quest_editor.go` (308 lines). Visual UI pending.
- **Deliverable**: Browser UI for creating quest chains with drag-and-drop objectives
- **Dependencies**: Steps 4, 5 (editor infrastructure)
- **Goal Impact**: Completes "Content creation utilities" goal with visual editing
- **Files**:
  - `pkg/wasmui/quest_editor.go` — quest chain visualization (PENDING)
  - `pkg/wasmui/quest_node.go` — draggable quest objective nodes (PENDING)
  - `data/schemas/quest_schema.json` — validation schema (already exists)
  - `pkg/server/handlers_quest_editor.go` — Backend handlers ✅
- **Current Foundation**: `cmd/quest-builder/` CLI provides data model, backend RPC complete
- **Acceptance**: User can create quest with 3+ objectives visually
- **Validation**:
  ```bash
  # E2E test: create quest, export, validate against schema
  go test ./test/e2e/... -run TestQuestEditorCreate -v
  ```

### Step 7: Implement Editor Save/Load Workflow
- [x] **COMPLETED** — Save/load implemented in `pkg/wasmui/map_editor.go` with Ctrl+S/Ctrl+O shortcuts in editor.go
- **Deliverable**: Persistent editor state with file export/import
- **Dependencies**: Steps 4, 5, 6 (editor features complete)
- **Goal Impact**: Makes visual editors production-ready
- **Files**:
  - `pkg/wasmui/editor_persistence.go` — save/load logic (integrated in map_editor.go)
  - `pkg/persistence/editor_state.go` — serialization format (JSON via EditorMapState)
  - Update `web/editor.html` with save/load buttons (keyboard shortcuts available)
- **Acceptance**: Editor state survives browser refresh
- **Validation**:
  ```bash
  # E2E test: create content, save, reload page, verify content restored
  go test ./test/e2e/... -run TestEditorPersistence -v
  ```

### Step 8: Migrate WebSocket Library to Coder Fork
- [x] **COMPLETED** — Already using `github.com/coder/websocket v1.8.14` (see go.mod line 8, websocket_nhooyr.go)
- **Deliverable**: Update `nhooyr.io/websocket` to `github.com/coder/websocket` (actively maintained fork)
- **Dependencies**: None (can proceed independently)
- **Goal Impact**: Addresses dependency health — original library transferred to Coder
- **Files**:
  - `go.mod` — update import path ✅
  - `pkg/server/websocket_nhooyr.go` → `pkg/server/websocket_coder.go` (still named nhooyr but uses coder import)
  - All files importing `nhooyr.io/websocket`
- **Current State**: Using `github.com/coder/websocket v1.8.14` ✅
- **Acceptance**: All WebSocket tests pass with new import
- **Validation**:
  ```bash
  go mod tidy && go test ./pkg/server/... ./test/e2e/... -v
  ```

---

## Scope Assessment

Using project-calibrated thresholds:

| Metric | Value | Threshold | Assessment |
|--------|-------|-----------|------------|
| Functions above complexity 9.0 | 4 | <5 (Small), 5–15 (Medium) | **Small** |
| Duplication ratio | 1.5% | <3% (Small), 3–10% (Medium) | **Small** |
| Goal achievement gaps | 3 | - | **Medium** (3 features) |
| CLI tool coverage gap | ~30% avg | 10–25% (Medium) | **Medium** |

**Overall Scope: Medium** — Most metrics are healthy, but completing 3 partial goals with GUI features requires moderate effort.

---

## Priority Order Rationale

1. **Step 1 (Asset Download)**: Highest user impact, zero dependencies, unblocks visual experience
2. **Step 2 (CLI Tests)**: Foundation stability before building GUI on top
3. **Step 3 (Duplication)**: Improves maintainability for subsequent handler additions
4. **Steps 4-7 (GUI Editor)**: Sequential dependency chain — each builds on previous
5. **Step 8 (WebSocket Lib)**: Independent maintenance task, can be parallelized

---

## Verification Commands

```bash
# Full validation suite
make test                           # All unit tests pass
make wasm                           # WASM builds successfully
make assets-verify                  # Asset count check
go test -race ./...                 # No race conditions
go vet ./...                        # No static analysis issues

# Metrics validation
go-stats-generator analyze . --skip-tests --sections functions,duplication 2>/dev/null | python3 -c "
import sys, json
d=json.load(sys.stdin)
high_cc = len([f for f in d['functions'] if f['complexity']['cyclomatic'] > 9])
dup_ratio = d['duplication']['duplication_ratio'] * 100
print(f'Functions CC>9: {high_cc} (target: <5)')
print(f'Duplication: {dup_ratio:.1f}% (target: <3%)')
"

# Goal completion check
make assets-verify                  # 521/521 for asset goal
curl -s http://localhost:8080/editor | grep -q "canvas"  # Editor goal
go test ./pkg/cliutil/... -cover    # CLI coverage goal
```

---

## Research Notes

### Competitive Landscape
- **GoldBox RPG** is the most active modern Gold Box-style engine on GitHub
- Alternatives: Dungeon Craft (FRUA successor, older codebase), generic engines (Godot, EasyRPG)
- Unique position: Go-based, JSON-RPC API, WebSocket real-time — targets web deployment

### Dependency Health
- **nhooyr.io/websocket**: Transferred to Coder (github.com/coder/websocket) in Aug 2024. Step 8 addresses migration.
- **Ebitengine v2.7.0**: Requires Go 1.24+ for latest. Current Go 1.25.6 is compatible.
- **No critical vulnerabilities** in direct dependencies per `govulncheck`

### Community Priorities (from README/ROADMAP)
1. Asset availability — primary visual experience gap
2. Content creator accessibility — CLI-only is high barrier
3. No open GitHub issues — mature codebase, focus on feature completion

---

*Generated: 2026-03-13*
*Based on go-stats-generator v1.0.0 analysis*
