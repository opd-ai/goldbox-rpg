# Implementation Plan: Complete Browser Editor Features & Code Quality

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by the classic SSI Gold Box series, providing character management, turn-based combat systems, and world interactions through JSON-RPC API with WebSocket support.
- **Current goal**: Complete the only documented-but-unimplemented feature (keyboard shortcuts in browser editors) and address code quality improvements.
- **Estimated Scope**: Small

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Six core attributes (STR/DEX/CON/INT/WIS/CHA) | ✅ Achieved | No |
| Class-based system (6 classes) | ✅ Achieved | No |
| Multiple character creation methods | ✅ Achieved | No |
| Equipment with class proficiency | ✅ Achieved | No |
| Comprehensive Effect System | ✅ Achieved | No |
| Advanced spatial indexing (R-tree) | ✅ Achieved | No |
| WebSocket real-time communication | ✅ Achieved | No |
| Health monitoring endpoints | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Circuit breaker patterns | ✅ Achieved | No |
| Input validation framework | ✅ Achieved | No |
| Asset generation pipeline (521 assets) | ✅ Achieved | No |
| Complete spell system (levels 0-9) | ✅ Achieved | No |
| Embedded Adventures (10 packs) | ✅ Achieved | No |
| Guild and faction systems | ✅ Achieved | No |
| Advanced NPC AI | ✅ Achieved | No |
| **Browser-based editor keyboard shortcuts** | ⚠️ Documented but unimplemented | **Yes** |
| ≥60% test coverage | ✅ 82.5% achieved | No |
| Migrate test WebSocket client from archived gorilla/websocket | ⚠️ Low priority debt | **Yes** |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 11 functions above threshold (cyclomatic > 9), 0 above 15
- **Duplication ratio**: 0.0% (no significant duplication detected)
- **Doc coverage**: Not fully parsed; existing docs meet standards per GAPS.md (89.9%)
- **Package coupling**: 19 packages, well-separated; largest: `pkg/game` (463 funcs), `pkg/wasmui` (278 funcs)

### High Complexity Functions (relevant to plan)
| CC | File | Function |
|----|------|----------|
| 14 | pkg/wasmui/character_creation.go | updateCharCreationAttributes |
| 13 | pkg/wasmui/exploration.go | updateExploration |
| 12 | pkg/wasmui/overlays.go | updateInventory |
| 11 | pkg/wasmui/combat_screen.go | updateCombat |
| 11 | pkg/wasmui/overlays.go | updateSpellbook |
| 10 | pkg/wasmui/game.go | connectAndJoin |

Note: These are WASM UI functions. Complexity is acceptable for game loop state machines but should not increase further.

## Implementation Steps

### Step 1: Implement Map Editor Keyboard Shortcuts
- **Deliverable**: Add keyboard event handling to `pkg/wasmui/editor.go` for documented shortcuts
  - `Ctrl+S`: Save current map (calls existing save RPC)
  - `Ctrl+Z`: Undo last edit
  - `Ctrl+Y`: Redo last undo
  - `G`, `W`, `S`, `D`: Quick-select terrain types (grass, water, stone, dirt)
- **Dependencies**: None
- **Goal Impact**: Completes the only documented-but-unimplemented feature (100% feature parity)
- **Acceptance**: 
  - Manual test: keyboard shortcuts functional at `http://localhost:8080/editor.html`
  - No increase in cyclomatic complexity of `editor.go` beyond +2 per new handler
- **Validation**: 
  ```bash
  make wasm && make run
  # Then manually test keyboard shortcuts in browser
  go test ./pkg/wasmui/... -v -run TestEditor
  ```

### Step 2: Implement Quest Editor Keyboard Shortcuts
- **Deliverable**: Add keyboard event handling to `pkg/wasmui/quest_editor.go`
  - `Ctrl+S`: Save current quest
  - `Ctrl+Z`: Undo last change
  - `Ctrl+Y`: Redo
- **Dependencies**: Step 1 (shared keyboard handling pattern)
- **Goal Impact**: Completes editor feature parity
- **Acceptance**: 
  - Manual test: keyboard shortcuts functional at `http://localhost:8080/quest-builder.html`
- **Validation**:
  ```bash
  make wasm && make run
  # Then manually test keyboard shortcuts in browser
  go test ./pkg/wasmui/... -v -run TestQuestEditor
  ```

### Step 3: Update Editor Documentation
- **Deliverable**: Remove "*(To be implemented in future versions)*" disclaimer from `docs/EDITOR_GUIDE.md:175`
- **Dependencies**: Steps 1 and 2 complete
- **Goal Impact**: Documentation accuracy
- **Acceptance**: Documentation matches implemented behavior
- **Validation**:
  ```bash
  grep -c "To be implemented" docs/EDITOR_GUIDE.md
  # Should return 0
  ```

### Step 4: Migrate E2E Test WebSocket Client
- **Deliverable**: 
  - Update `test/e2e/client.go` to use `github.com/coder/websocket` instead of `gorilla/websocket`
  - Remove `gorilla/websocket` from `go.mod` after migration
- **Dependencies**: None (independent of Steps 1-3)
- **Goal Impact**: Eliminates archived dependency; aligns with server's WebSocket library choice
- **Acceptance**: 
  - `go mod graph | grep gorilla` returns empty
  - All E2E tests pass
- **Validation**:
  ```bash
  go test ./test/e2e/... -v
  go mod tidy
  go mod graph | grep -c gorilla
  # Should return 0
  ```

### Step 5: Review Open Dependabot PRs
- **Deliverable**: Evaluate and merge or close:
  - PR #38: `golang` Docker image upgrade (1.23-bookworm → 1.26-bookworm)
  - PR #35: `actions/upload-artifact` upgrade (v4 → v7)
- **Dependencies**: Step 4 (ensures test suite is fully updated before CI changes)
- **Goal Impact**: CI modernization; dependency freshness
- **Acceptance**: 
  - PRs reviewed and either merged (if CI passes) or closed with comment
  - CI pipeline passes on master after merge
- **Validation**:
  ```bash
  # After merge, verify CI status
  gh pr view 38 --json state
  gh pr view 35 --json state
  ```

## Scope Assessment Rationale

| Metric | Value | Category |
|--------|-------|----------|
| Functions above complexity 9.0 | 11 | Medium |
| Duplication ratio | 0.0% | Small |
| Doc coverage gap | <10% (estimated) | Small |
| New code required | ~100-150 lines | Small |
| Files to modify | 4 | Small |

**Overall: Small** — The primary work is implementing documented keyboard shortcuts (UI polish) and migrating one test file. No complex refactoring required.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Ebitengine keyboard API changes | Low | Medium | Pin Ebiten version; test on current v2.9.9 |
| WebSocket client API differences | Low | Low | coder/websocket is API-compatible with nhooyr; migration is straightforward |
| CI breakage from Dependabot PRs | Medium | Low | Review PR diffs carefully; run full test suite before merge |

## Success Criteria

1. ✅ All keyboard shortcuts documented in `docs/EDITOR_GUIDE.md:173-183` are functional
2. ✅ Documentation contains no "To be implemented" disclaimers for implemented features
3. ✅ `gorilla/websocket` removed from dependency graph
4. ✅ All tests pass: `go test ./... -race`
5. ✅ CI pipeline green on master

---

*Generated: 2026-03-16*  
*Tool: go-stats-generator v1.0.0*  
*Based on: ROADMAP.md, GAPS.md, CHANGELOG.md, CLIENT_SPEC.md, go-stats-generator metrics*
