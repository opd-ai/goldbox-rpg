# Implementation Plan: WASM UI Complexity Reduction and Test Coverage Alignment

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by the classic SSI Gold Box series, providing character management, combat systems, and world interactions through JSON-RPC API with WebSocket support.
- **Current goal**: Reduce high-complexity functions in the WASM UI layer to improve maintainability and align actual test coverage with documented claims (78-96%).
- **Estimated Scope**: Medium (14 functions above complexity threshold 9, 4 packages below 78% coverage)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Core RPG mechanics | ✅ Achieved | No |
| Combat and effect systems | ✅ Achieved | No |
| WebSocket real-time communication | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| System resilience patterns | ✅ Achieved | No |
| Health monitoring and metrics | ✅ Achieved | No |
| Asset pipeline (521 assets) | ✅ Achieved | No |
| Advanced NPC AI behaviors | ✅ Achieved | No |
| Enhanced combat mechanics | ✅ Achieved | No |
| Complete spell system | ✅ Achieved | No |
| Guild and faction systems | ✅ Achieved | No |
| Embedded adventures (10 packs) | ✅ Achieved | No |
| Browser-based editors | ✅ Achieved | No |
| Test coverage 78-96% | ⚠️ Partial (65-96%) | **Yes** |
| Maintainable codebase | ⚠️ Partial (14 high-complexity functions) | **Yes** |

## Metrics Summary (go-stats-generator 2026-03-17)

| Metric | Value | Threshold | Assessment |
|--------|-------|-----------|------------|
| Total LOC | 34,415 | — | — |
| Total Functions/Methods | 2,569 | — | — |
| Functions above complexity 9 | 18 | <15 | ⚠️ Needs attention |
| Functions above complexity 15 | 7 | <5 | ⚠️ Needs attention |
| Long functions (>50 lines) | 111 | <100 | ⚠️ Moderate |
| Duplication ratio | 0.01% | <5% | ✅ Excellent |
| Package circular dependencies | 0 | 0 | ✅ Excellent |

### High-Complexity Hotspots (complexity > 15)

| Function | Package | File | Complexity | Lines |
|----------|---------|------|------------|-------|
| `Update` | wasmui | adventure_screen.go | 23 | 92 |
| `updateInventory` | wasmui | overlays.go | 23 | 106 |
| `updateSpellbook` | wasmui | overlays.go | 21 | 105 |
| `updateQuestLogOverlay` | wasmui | overlays.go | 20 | 116 |
| `updateCombat` | wasmui | combat_screen.go | 18 | 118 |
| `updateCharCreationClass` | wasmui | character_creation.go | 18 | 74 |
| `updateExploration` | wasmui | exploration.go | 15 | 116 |

### Test Coverage Below Target (78%)

| Package | Current | Target | Gap |
|---------|---------|--------|-----|
| `test/e2e` | 65.5% | 78% | 12.5% |
| `scripts` | 68.8% | 78% | 9.2% |
| `cmd/quest-builder` | 71.6% | 78% | 6.4% |
| `cmd/server` | 71.8% | 78% | 6.2% |

---

## Implementation Steps

### Step 1: Refactor `pkg/wasmui/overlays.go` High-Complexity Functions

- **Deliverable**: Split `updateInventory` (complexity 23), `updateSpellbook` (complexity 21), and `updateQuestLogOverlay` (complexity 20) into focused helper functions, each with complexity ≤10.
- **Dependencies**: None
- **Goal Impact**: Addresses maintainability concern; reduces bug risk in user-facing overlay code
- **Acceptance**: 
  - Each refactored function has cyclomatic complexity ≤10
  - All existing wasmui tests pass (`go test -race ./pkg/wasmui/...`)
  - No functional regression in overlay behavior
- **Validation**: 
  ```bash
  go-stats-generator analyze ./pkg/wasmui --skip-tests --format json | \
    jq '[.functions[] | select(.name | test("updateInventory|updateSpellbook|updateQuestLogOverlay")) | {name: .name, complexity: .complexity.cyclomatic}]'
  ```

**Suggested decomposition for `updateInventory`:**
- `handleInventoryInput()` — process keyboard/touch input for item selection
- `updateInventoryDragDrop()` — handle drag-and-drop state transitions
- `updateInventoryTooltip()` — manage tooltip visibility and positioning

**Suggested decomposition for `updateSpellbook`:**
- `handleSpellbookNavigation()` — spell level tabs and scrolling
- `handleSpellSelection()` — spell selection and casting initiation
- `filterSpellList()` — filter spells by level, school, or search term

**Suggested decomposition for `updateQuestLogOverlay`:**
- `handleQuestLogInput()` — navigation and selection
- `updateQuestProgress()` — sync quest state with server response
- `renderQuestDetails()` — populate detail panel when quest selected

---

### Step 2: Refactor `pkg/wasmui/adventure_screen.go` Update Function

- **Deliverable**: Split `Update` (complexity 23) into focused state-update helpers with complexity ≤10.
- **Dependencies**: None (can be done in parallel with Step 1)
- **Goal Impact**: The main game loop update function is critical path; high complexity increases regression risk during feature additions.
- **Acceptance**:
  - `Update` complexity ≤10
  - All adventure screen state transitions work correctly
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/wasmui --skip-tests --format json | \
    jq '[.functions[] | select(.file | contains("adventure_screen")) | select(.name == "Update") | {name: .name, complexity: .complexity.cyclomatic}]'
  ```

**Suggested decomposition:**
- `updateAdventureMenuState()` — handle menu navigation
- `updateAdventureActionState()` — handle action confirmations
- `processAdventureInput()` — dispatch input to appropriate handler

---

### Step 3: Refactor `pkg/wasmui/combat_screen.go` updateCombat

- **Deliverable**: Split `updateCombat` (complexity 18, 118 lines) into ≤3 focused functions.
- **Dependencies**: None (can be done in parallel with Steps 1-2)
- **Goal Impact**: Combat is the most complex gameplay system; clear update logic prevents combat bugs.
- **Acceptance**:
  - `updateCombat` complexity ≤12 (acceptable for combat complexity)
  - All combat E2E tests pass
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/wasmui --skip-tests --format json | \
    jq '[.functions[] | select(.name == "updateCombat") | {name: .name, complexity: .complexity.cyclomatic}]'
  ```

**Suggested decomposition:**
- `processCombatInput()` — handle player input during combat
- `updateCombatAnimations()` — manage attack/spell animation state
- `handleCombatTurnEnd()` — process turn resolution and state transition

---

### Step 4: Refactor `pkg/wasmui/exploration.go` updateExploration

- **Deliverable**: Split `updateExploration` (complexity 15, 116 lines) into focused functions.
- **Dependencies**: None (can be done in parallel with Steps 1-3)
- **Goal Impact**: Exploration is the primary non-combat game state; reducing complexity improves feature velocity.
- **Acceptance**:
  - `updateExploration` complexity ≤10
  - All exploration-related tests pass
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/wasmui --skip-tests --format json | \
    jq '[.functions[] | select(.name == "updateExploration") | {name: .name, complexity: .complexity.cyclomatic}]'
  ```

**Suggested decomposition:**
- `handleExplorationMovement()` — process WASD/arrow/touch movement
- `handleExplorationInteraction()` — NPC/object interaction
- `triggerExplorationEvents()` — emit events for world interactions

---

### Step 5: Increase `test/e2e` Coverage to 78%

- **Deliverable**: Add E2E test scenarios to increase coverage from 65.5% to ≥78%.
- **Dependencies**: Steps 1-4 should be completed first to ensure stable refactored code.
- **Goal Impact**: Aligns actual coverage with README badge claim (78-96%).
- **Acceptance**:
  - `go test -cover ./test/e2e/...` reports ≥78% coverage
  - New tests cover PCG integration, guild operations, and faction diplomacy E2E flows
- **Validation**:
  ```bash
  go test -cover ./test/e2e/... 2>/dev/null | grep coverage
  ```

**Suggested new test scenarios:**
- `TestPCGDungeonGeneration_E2E` — verify procedural dungeon generation via RPC
- `TestGuildCreation_E2E` — create guild, add members, manage treasury
- `TestFactionDiplomacy_E2E` — modify faction standings, trigger alliance/war
- `TestAdventureProgression_E2E` — load adventure, complete quest objectives

---

### Step 6: Increase `scripts` Coverage to 78%

- **Deliverable**: Add tests for `scripts/verify_adventures.go` and `scripts/find_untested_files.go`.
- **Dependencies**: None
- **Goal Impact**: Aligns coverage with stated minimum; scripts are CI-critical.
- **Acceptance**:
  - `go test -cover ./scripts/...` reports ≥78% coverage
- **Validation**:
  ```bash
  go test -cover ./scripts/... 2>/dev/null | grep coverage
  ```

**Suggested test additions:**
- Test `verify_adventures.go` with valid and invalid adventure YAML
- Test `find_untested_files.go` with mock package structure

---

### Step 7: Increase `cmd/quest-builder` Coverage to 78%

- **Deliverable**: Add tests for edge cases in quest-builder CLI.
- **Dependencies**: None
- **Goal Impact**: CLI tools should meet same quality bar as library code.
- **Acceptance**:
  - `go test -cover ./cmd/quest-builder/...` reports ≥78% coverage
- **Validation**:
  ```bash
  go test -cover ./cmd/quest-builder/... 2>/dev/null | grep coverage
  ```

**Suggested test additions:**
- Test circular quest dependency detection
- Test invalid objective types
- Test missing required fields in quest YAML

---

### Step 8: Increase `cmd/server` Coverage to 78%

- **Deliverable**: Add tests for server startup edge cases and graceful shutdown.
- **Dependencies**: None
- **Goal Impact**: Server entry point is production-critical.
- **Acceptance**:
  - `go test -cover ./cmd/server/...` reports ≥78% coverage
- **Validation**:
  ```bash
  go test -cover ./cmd/server/... 2>/dev/null | grep coverage
  ```

**Suggested test additions:**
- Test graceful shutdown on SIGTERM/SIGINT
- Test invalid configuration handling
- Test port binding failure recovery

---

### Step 9: Update README Coverage Badge

- **Deliverable**: Update README.md coverage badge from "78-96%" to accurate range after coverage improvements.
- **Dependencies**: Steps 5-8 completed
- **Goal Impact**: Documentation accuracy; prevents misleading contributors.
- **Acceptance**:
  - Badge reflects actual minimum-maximum coverage range
  - All packages ≥78% coverage (verified by CI)
- **Validation**:
  ```bash
  go test -cover ./... 2>/dev/null | grep -oP 'coverage: \K[0-9.]+' | sort -n | head -1
  ```

---

## Summary Metrics Targets

| Metric | Current | Target | Validation Command |
|--------|---------|--------|-------------------|
| Functions complexity >15 | 7 | ≤3 | `go-stats-generator analyze . --skip-tests --format json \| jq '[.functions[] \| select(.complexity.cyclomatic > 15)] \| length'` |
| Functions complexity >9 | 18 | ≤10 | `go-stats-generator analyze . --skip-tests --format json \| jq '[.functions[] \| select(.complexity.cyclomatic > 9)] \| length'` |
| Minimum package coverage | 65.5% | 78% | `go test -cover ./... 2>/dev/null \| grep coverage \| grep -oP '[0-9.]+' \| sort -n \| head -1` |
| README coverage badge | "78-96%" | Accurate | Manual verification |

---

## Dependency Graph

```
Step 1 (overlays.go) ─┐
Step 2 (adventure_screen.go) ─┼─→ Step 5 (E2E tests) ─→ Step 9 (README)
Step 3 (combat_screen.go) ─┤
Step 4 (exploration.go) ─┘

Step 6 (scripts) ─────────────┐
Step 7 (quest-builder) ───────┼─→ Step 9 (README)
Step 8 (cmd/server) ──────────┘
```

Steps 1-4 can be parallelized. Steps 6-8 can be parallelized. Step 5 depends on Steps 1-4. Step 9 depends on Steps 5-8.

---

## Estimated Effort

| Step | Scope | Effort |
|------|-------|--------|
| Steps 1-4 (Refactoring) | 4 files, ~10 functions | 2-3 days |
| Steps 5-8 (Test coverage) | 4 packages | 2-3 days |
| Step 9 (Documentation) | 1 file | 0.5 day |
| **Total** | | **4-6 days** |

---

## Notes

- **No open GitHub issues**: The project has no open issues or PRs, so this plan is derived entirely from metrics analysis and stated roadmap goals.
- **WASM UI is the complexity hotspot**: 11 of 18 high-complexity functions are in `pkg/wasmui`. This is expected for game UI code but warrants focused attention.
- **Coverage claims are aspirational**: The README claims 78-96% but actual minimum is 65.5%. This plan closes that gap.
- **All major features achieved**: 31/32 stated goals are fully implemented. This plan addresses the remaining maintainability and quality debt.

---

*Generated: 2026-03-17 | Tool: go-stats-generator v1.0.0*
