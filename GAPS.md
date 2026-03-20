# Implementation Gaps — 2026-03-20

This document identifies gaps between stated goals in the README/documentation and the current implementation state.

---

## 1. API Documentation Coverage Gap

- **Stated Goal**: The README and `pkg/README-RPC.md` imply comprehensive API documentation for all JSON-RPC methods
- **Current State**: 48 of 74 RPC methods (65%) are documented. 26 methods are fully implemented but lack documentation:
  - **Quest System (8)**: `completeQuest`, `failQuest`, `getActiveQuests`, `getCompletedQuests`, `getQuest`, `getQuestLog`, `startQuest`, `updateObjective`
  - **Quest Editor (5)**: `questEditor.create`, `questEditor.delete`, `questEditor.get`, `questEditor.list`, `questEditor.update`
  - **Map Editor (4)**: `editor.createMap`, `editor.loadMap`, `editor.saveMap`, `editor.updateTile`
  - **Spatial Operations (5)**: `findPath`, `getObjectsInRadius`, `getObjectsInRange`, `getNearestObjects`, `getVisibleTiles`
  - **Adventure System (2)**: `adventure.list`, `adventure.load`
  - **Combat Helpers (2)**: `getCombatModifiers`, `rest`
- **Impact**: Developers integrating with the API must read source code to understand undocumented endpoints. This increases onboarding time and error risk.
- **Closing the Gap**: 
  1. Add documentation sections for all 26 undocumented methods in `pkg/README-RPC.md`
  2. Follow the existing format: method name, description, parameters table, response schema, example request/response
  3. Estimated effort: 4-6 hours
  4. Validation: `grep -c '^### ' pkg/README-RPC.md` should return 74

---

## 2. REST Endpoint HP Restoration

- **Stated Goal**: The `rest` RPC method should handle player resting mechanics per D&D conventions
- **Current State**: `handleRest()` at `pkg/server/handlers.go:1319` contains TODO: "Could also restore some HP here based on game rules." The handler currently only validates and returns success without HP restoration.
- **Impact**: Players cannot restore HP through resting, breaking expected RPG gameplay loop. Must rely on healing items/spells exclusively.
- **Closing the Gap**:
  1. Implement HP restoration in `handleRest()`:
     ```go
     // Calculate HP restoration (e.g., 1 HP per character level)
     restoreAmount := player.Level
     player.Heal(restoreAmount)
     ```
  2. Add response field for HP restored
  3. Add test case in `handlers_test.go`
  4. Estimated effort: 1-2 hours

---

## 3. First-Person View Server Query Integration

- **Stated Goal**: The first-person dungeon view should display accurate wall/door visibility from server-authoritative game state
- **Current State**: `drawFirstPersonViewAt()` at `pkg/wasmui/exploration.go:500` uses cached `visibleTiles` from prior queries. TODO at line 742: "Query server for visible walls via getVisibleWalls RPC". The client may show stale tile data during rapid movement.
- **Impact**: Visual desync between client rendering and server state during exploration. Players may see walls that have changed or miss newly revealed areas.
- **Closing the Gap**:
  1. In `maybeRefreshVisibleTiles()`, call the existing `getVisibleTiles` RPC endpoint
  2. The server handler already exists at `pkg/server/handlers_spatial.go:355`
  3. Add debouncing to prevent excessive RPC calls (e.g., max 1 request per 100ms)
  4. Estimated effort: 2-3 hours

---

## 4. High-Complexity UI Function

- **Stated Goal**: Code quality standards suggest functions should have cyclomatic complexity ≤15 and ≤50 lines
- **Current State**: `drawFirstPersonViewAt()` at `pkg/wasmui/exploration.go:500` has cyclomatic complexity of 22 and 154 lines. It handles wall rendering at three depth levels (far/mid/near) with door detection and gradient drawing.
- **Impact**: High complexity increases bug risk and makes the function difficult to maintain or modify. Any change to perspective rendering requires understanding 150+ lines of nested conditionals.
- **Closing the Gap**:
  1. Extract helper functions:
     - `drawFarDepthLayer(screen, vpX, vpY, vpWidth, vpHeight, tiles)` 
     - `drawMidDepthLayer(screen, vpX, vpY, vpWidth, vpHeight, tiles)`
     - `drawNearDepthLayer(screen, vpX, vpY, vpWidth, vpHeight, tiles)`
  2. Extract wall/door detection into `getTileAtDepth(tiles, relX, depth) TileInfo`
  3. Target: each function ≤40 lines, complexity ≤10
  4. Estimated effort: 3-4 hours

---

## 5. Adventure Map Counting Precision

- **Stated Goal**: README claims "100 maps" across 10 adventure packs
- **Current State**: Maps are embedded within `adventure.yaml` files without explicit `map_id:` keys. The structure uses `maps:` arrays with `name:` fields, but automated counting is imprecise. Visual inspection suggests ~10 maps per adventure (100 total), but this cannot be programmatically verified.
- **Impact**: Documentation accuracy cannot be automatically validated. CI/CD cannot verify map count claims.
- **Closing the Gap**:
  1. Add explicit `map_id:` field to each map definition in adventure YAML files
  2. Create validation script: `scripts/count_adventure_maps.sh`
  3. Add CI check to verify map count matches README claims
  4. Estimated effort: 2-3 hours

---

## 6. WASM UI Test Coverage

- **Stated Goal**: Project maintains ≥60% test coverage (all packages meet this)
- **Current State**: `pkg/wasmui/` has 71.3% coverage—above threshold but lowest among all packages. Complex drawing functions in `exploration.go`, `combat_screen.go`, and `adventure_screen.go` have limited test coverage due to ebiten dependency.
- **Impact**: UI bugs may go undetected. Changes to rendering logic have higher regression risk.
- **Closing the Gap**:
  1. Create mock ebiten.Image for testing draw functions
  2. Add table-driven tests for:
     - `drawFirstPersonViewAt()` with various tile configurations
     - `drawCombatActionBar()` with different game states
     - `drawMinimapOverlay()` boundary conditions
  3. Target: 80% coverage for pkg/wasmui/
  4. Estimated effort: 4-6 hours

---

## 7. Morale System UI Integration

- **Stated Goal**: Per ROADMAP.md (Group: Combat Screen, Item 1), enemy morale state should display in the combat UI
- **Current State**: The morale system is fully implemented in `pkg/game/morale.go` with states (Steadfast, Shaken, Broken, Panicked). The `InitiativeEntry` struct has a `MoraleState` field. However, `pkg/wasmui/combat_screen.go` never displays morale, and server handlers don't populate the field.
- **Impact**: Players cannot make tactical decisions based on enemy morale despite the backend fully supporting it. This is a Gold Box authenticity gap.
- **Closing the Gap**:
  1. In server combat handlers, populate `InitiativeEntry.MoraleState` from `MoraleSystem.GetMoraleState()`
  2. In `drawInitiativeEntry()` at `pkg/wasmui/combat_screen.go:597`, add morale state display
  3. Add message log entries for morale changes
  4. Estimated effort: 2-3 hours

---

## 8. Effect Display on Combat Tokens

- **Stated Goal**: Per ROADMAP.md (Group: Combat Screen, Item 2), active effects should display on combat tokens
- **Current State**: `PlayerState.Effects` contains active effect data. Combat tokens are drawn in `drawPlayerToken()` and `drawSingleEnemyToken()` but show no effect indicators. Effect display only appears in exploration mode character panel.
- **Impact**: During combat, players must remember which enemies have DoTs, stuns, or other effects. This reduces tactical clarity.
- **Closing the Gap**:
  1. Add `drawEffectIndicators()` helper function
  2. Call from `drawPlayerToken()` and `drawSingleEnemyToken()`
  3. Display small colored squares (8x8) above tokens using effect type colors from `types_ui.go`
  4. Limit to 4 icons with overflow indicator
  5. Estimated effort: 2-3 hours

---

## Summary

| Gap | Severity | Effort | Category |
|-----|----------|--------|----------|
| API Documentation (35% missing) | HIGH | 4-6h | Documentation |
| REST HP Restoration | MEDIUM | 1-2h | Feature |
| First-Person Server Query | MEDIUM | 2-3h | Feature |
| High-Complexity UI Function | MEDIUM | 3-4h | Code Quality |
| Adventure Map Counting | LOW | 2-3h | Documentation |
| WASM UI Test Coverage | LOW | 4-6h | Testing |
| Morale System UI | LOW | 2-3h | Feature (Roadmap) |
| Effect Display on Tokens | LOW | 2-3h | Feature (Roadmap) |

**Total Estimated Effort**: 21-32 hours to close all gaps

---

## Notes

- No CRITICAL gaps were identified. All core gameplay features work as documented.
- The HIGH-priority API documentation gap is the most impactful for developer experience.
- ROADMAP.md items (Morale UI, Effect Display) are acknowledged future work, not broken promises.
- All gaps have clear remediation paths with specific file locations and validation methods.
