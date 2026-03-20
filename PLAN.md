# Implementation Plan: Gold Box UI Authenticity

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by classic SSI Gold Box games, providing character management, turn-based combat, and world interactions through JSON-RPC API with WebSocket support.
- **Current goal**: Achieve Gold Box UI faithfulness—the project's stated reference standard requires distinctive SSI visual language with fixed panels, EGA palette, chunky sprites, and tactical feedback.
- **Estimated Scope**: Medium (12 functions above complexity threshold, 35% API documentation gap)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|----------------|---------------------|
| Gold Box visual authenticity | ⚠️ Partial (palette done, panels incomplete) | Yes |
| API documentation completeness | ❌ 65% coverage (26 methods undocumented) | Yes |
| Combat tactical feedback | ⚠️ Morale/effects not surfaced in UI | Yes |
| Effect system functionality | ❌ Stun/Root effects non-functional | Yes |
| First-person view accuracy | ❌ Shows placeholder, not actual map | Yes |
| Code maintainability | ⚠️ 1 function at complexity 22 | Yes |

## Metrics Summary

| Metric | Current Value | Threshold | Status |
|--------|---------------|-----------|--------|
| Complexity hotspots (>9) | 14 functions | <5 Small, 5-15 Medium | Medium |
| Doc coverage (exports) | 100% | ≥90% | ✅ |
| API doc coverage | 65% | 100% | ❌ |
| Total LOC | 38,877 | - | - |
| Total functions | 2,911 | - | - |
| Duplication ratio | 0% | <3% | ✅ |

### Complexity Hotspots on Goal-Critical Paths

| Complexity | File | Function | Impact |
|------------|------|----------|--------|
| 22 | exploration.go:500 | `drawFirstPersonViewAt` | First-person rendering |
| 12 | overlays.go:1483 | `drawMinimapOverlay` | Map display |
| 11 | websocket.go:336 | `processWebSocketRequest` | Real-time comms |
| 10 | handlers.go:324 | `handleAttack` | Combat actions |
| 10 | handlers.go:525 | `executeCastSpellAction` | Spell casting |
| 10 | exploration.go:50 | `handleExplorationOverlayKeys` | UI input |
| 10 | overlays.go:640 | `drawSpellbookScreen` | Spell UI |

---

## Implementation Steps

### Step 1: Fix Critical Effect Processing (Stun/Root)

- **Deliverable**: Functional crowd-control effects that prevent actions
- **Dependencies**: None
- **Goal Impact**: Combat tactical feedback, Gold Box authenticity (CC was fundamental)
- **Files to modify**:
  - `pkg/server/handlers.go` - Add stun/root checks to `handleMove`, `handleAttack`, `handleCastSpell`
  - `pkg/game/effectbehavior.go` - Implement `EffectStun` and `EffectRoot` case bodies at lines 494-497
- **Acceptance**: 
  - `go test ./pkg/game/... -v -run TestStun` passes
  - `go test ./pkg/server/... -v -run TestEffect` passes
  - Stunned character returns error on action attempt
- **Validation**: 
  ```bash
  go test ./pkg/game/... ./pkg/server/... -v -run "Stun|Root" && echo "PASS"
  ```

---

### Step 2: Wire Morale System to Combat UI

- **Deliverable**: Enemy morale state visible in initiative panel and message log
- **Dependencies**: None
- **Goal Impact**: Gold Box authenticity (morale was tactically visible), combat feedback
- **Files to modify**:
  - `pkg/wasmui/combat_screen.go` - Add morale display in `drawInitiativeEntry()` at line 597
  - `pkg/server/handlers.go` - Populate `InitiativeEntry.MoraleState` from `MoraleSystem.GetMoraleState()`
  - `pkg/wasmui/types_ui.go` - Add `getMoraleColor(state string) color.RGBA` helper
- **Acceptance**: 
  - Morale state displays when not "Steadfast"
  - Message log announces morale changes
- **Validation**: 
  ```bash
  grep -q "MoraleState" pkg/wasmui/combat_screen.go && echo "UI wired"
  go build ./pkg/wasmui/... && echo "WASM builds"
  ```

---

### Step 3: Display Active Effects on Combat Tokens

- **Deliverable**: Visual effect indicators (colored icons) above combat tokens
- **Dependencies**: None
- **Goal Impact**: Combat tactical feedback, Gold Box authenticity (status icons on tokens)
- **Files to modify**:
  - `pkg/wasmui/combat_screen.go` - Add `drawEffectIndicators()` helper, call from `drawPlayerToken()` and `drawSingleEnemyToken()`
- **Acceptance**: 
  - Active effects show as 8x8 colored squares above tokens
  - Uses ColorEffectDebuff/ColorEffectControl/ColorEffectBuff from types_ui.go
  - Maximum 4 icons with overflow indicator
- **Validation**: 
  ```bash
  grep -q "drawEffectIndicators" pkg/wasmui/combat_screen.go && echo "PASS"
  GOOS=js GOARCH=wasm go build -o /dev/null ./cmd/wasm-ui && echo "WASM builds"
  ```

---

### Step 4: Implement Attack Roll Narration

- **Deliverable**: Rich textual combat narration in message log (hits, misses, criticals)
- **Dependencies**: None
- **Goal Impact**: Gold Box authenticity (message log was primary feedback channel)
- **Files to modify**:
  - `pkg/wasmui/combat_screen.go` - Format attack results as narration
  - `pkg/server/handlers.go` - Enhance `handleAttack` response with `AttackResult` struct
  - `pkg/wasmui/types_rpc.go` - Add `AttackResult` struct
- **Acceptance**: 
  - Every attack produces message log entry
  - Critical hits emphasized with "CRITICAL HIT"
  - Misses explicitly announced
- **Validation**: 
  ```bash
  grep -q "AttackResult" pkg/wasmui/types_rpc.go && echo "Struct defined"
  go build ./pkg/... && echo "Builds"
  ```

---

### Step 5: Refactor `drawFirstPersonViewAt()` for Maintainability

- **Deliverable**: Reduced complexity from 22 to ≤10 per extracted function
- **Dependencies**: None (prerequisite for Step 6)
- **Goal Impact**: Code maintainability, enables future first-person view improvements
- **Files to modify**:
  - `pkg/wasmui/exploration.go` - Extract helpers:
    - `drawFarDepthLayer(screen, vpX, vpY, vpWidth, vpHeight, tiles)`
    - `drawMidDepthLayer(screen, vpX, vpY, vpWidth, vpHeight, tiles)`
    - `drawNearDepthLayer(screen, vpX, vpY, vpWidth, vpHeight, tiles)`
    - `getTileAtDepth(tiles, relX, depth) TileInfo`
- **Acceptance**: 
  - `drawFirstPersonViewAt` complexity ≤10
  - Each extracted function ≤40 lines, complexity ≤10
  - No visual rendering changes (regression test)
- **Validation**: 
  ```bash
  go-stats-generator analyze ./pkg/wasmui/ --skip-tests --format json 2>/dev/null | \
    python3 -c "import json,sys; d=json.load(sys.stdin); f=[x for x in d['functions'] if x['name']=='drawFirstPersonViewAt']; print('Complexity:', f[0]['complexity']['cyclomatic'] if f else 'NOT_FOUND')"
  ```

---

### Step 6: Wire First-Person View to Actual Map Data

- **Deliverable**: First-person view reflects real dungeon geometry from server state
- **Dependencies**: Step 5 (refactored function is easier to modify)
- **Goal Impact**: Gold Box authenticity (exploration showed real map), core gameplay
- **Files to modify**:
  - `pkg/wasmui/exploration.go` - Replace placeholder rendering with map-driven logic
  - `pkg/wasmui/rpc_methods.go` - Add `GetVisibleTiles` RPC call
  - `pkg/server/handlers_spatial.go` - Handler already exists at line 355, verify it returns view cone data
- **Acceptance**: 
  - First-person view shows walls where map has walls
  - Doors visible at correct positions
  - View updates on player movement
- **Validation**: 
  ```bash
  grep -q "getVisibleTiles\|GetVisibleTiles" pkg/wasmui/exploration.go && echo "RPC wired"
  go test ./pkg/server/... -v -run TestGetVisibleTiles && echo "Server test passes"
  ```

---

### Step 7: Document Undocumented RPC Methods (Quest System)

- **Deliverable**: Full documentation for 8 quest system methods in `pkg/README-RPC.md`
- **Dependencies**: None
- **Goal Impact**: API documentation completeness (35% → ~24% gap reduction)
- **Methods to document**:
  - `completeQuest`, `failQuest`, `getActiveQuests`, `getCompletedQuests`
  - `getQuest`, `getQuestLog`, `startQuest`, `updateObjective`
- **Acceptance**: 
  - Each method has: description, parameters table, response schema, example request/response
  - Follows existing format in `pkg/README-RPC.md`
- **Validation**: 
  ```bash
  grep -c "^### " pkg/README-RPC.md  # Should increase by 8
  grep -E "completeQuest|failQuest|getActiveQuests" pkg/README-RPC.md | wc -l  # Should be ≥8
  ```

---

### Step 8: Document Undocumented RPC Methods (Editors)

- **Deliverable**: Full documentation for 9 editor methods in `pkg/README-RPC.md`
- **Dependencies**: Step 7 (sequential documentation work)
- **Goal Impact**: API documentation completeness (enables content creators)
- **Methods to document**:
  - Quest Editor: `questEditor.create`, `questEditor.delete`, `questEditor.get`, `questEditor.list`, `questEditor.update`
  - Map Editor: `editor.createMap`, `editor.loadMap`, `editor.saveMap`, `editor.updateTile`
- **Acceptance**: 
  - Each method documented with full spec
  - Examples use realistic data
- **Validation**: 
  ```bash
  grep -c "questEditor\." pkg/README-RPC.md  # Should be ≥5
  grep -c "editor\." pkg/README-RPC.md  # Should be ≥4
  ```

---

### Step 9: Document Remaining Undocumented RPC Methods

- **Deliverable**: Full documentation for 9 remaining methods
- **Dependencies**: Step 8
- **Goal Impact**: API documentation completeness (gap → 0%)
- **Methods to document**:
  - Spatial: `findPath`, `getObjectsInRadius`, `getObjectsInRange`, `getNearestObjects`, `getVisibleTiles`
  - Adventure: `adventure.list`, `adventure.load`
  - Combat: `getCombatModifiers`, `rest`
- **Acceptance**: 
  - All 74 RPC methods documented
  - `grep -c '^### ' pkg/README-RPC.md` returns 74
- **Validation**: 
  ```bash
  # Count documented methods
  grep -c "^### " pkg/README-RPC.md
  # Verify specific methods
  grep -E "findPath|getObjectsInRadius|adventure.list" pkg/README-RPC.md | wc -l
  ```

---

### Step 10: Implement REST HP Restoration

- **Deliverable**: `handleRest()` restores HP based on character level
- **Dependencies**: Step 9 (rest method needs documentation)
- **Goal Impact**: Core gameplay loop (rest/recovery is RPG fundamental)
- **Files to modify**:
  - `pkg/server/handlers.go` - Implement HP restoration in `handleRest()` at line 1319
  - Add test case in `handlers_test.go`
- **Acceptance**: 
  - Rest restores HP (e.g., 1 HP per character level)
  - Response includes HP restored amount
  - Test verifies HP changes
- **Validation**: 
  ```bash
  go test ./pkg/server/... -v -run TestRest && echo "PASS"
  grep -q "Heal\|RestoreHP" pkg/server/handlers.go && echo "HP restoration implemented"
  ```

---

### Step 11: Surface Faction Relations in UI

- **Deliverable**: Faction panel showing diplomatic standings
- **Dependencies**: None
- **Goal Impact**: Gold Box authenticity (faction visibility), game system wiring
- **Files to modify**:
  - `pkg/wasmui/overlays.go` - Add `drawFactionPanel(screen)`, add F key binding
  - `pkg/wasmui/types_rpc.go` - Add `FactionStanding` struct
  - `pkg/wasmui/rpc_methods.go` - Add `GetPlayerFactions` RPC call
- **Acceptance**: 
  - F key opens faction relations panel
  - Faction standings displayed with color coding (War=red → Allied=gold)
  - Reputation changes appear in message log
- **Validation**: 
  ```bash
  grep -q "drawFactionPanel" pkg/wasmui/overlays.go && echo "Panel implemented"
  GOOS=js GOARCH=wasm go build -o /dev/null ./cmd/wasm-ui && echo "WASM builds"
  ```

---

### Step 12: Wire Guild System to UI

- **Deliverable**: Guild panel showing membership, ranks, treasury, perks
- **Dependencies**: None
- **Goal Impact**: Game system wiring (guild backend exists, UI missing)
- **Files to modify**:
  - `pkg/wasmui/overlays.go` - Implement `drawGuildPanel(screen)`
  - `pkg/wasmui/exploration.go` - Implement `loadGuildData()`
  - `pkg/wasmui/types_rpc.go` - Add `GuildData` struct
- **Acceptance**: 
  - G key opens guild panel
  - Guild info displays correctly
  - Member list and treasury visible
- **Validation**: 
  ```bash
  grep -q "loadGuildData" pkg/wasmui/exploration.go && echo "Load implemented"
  grep -q "drawGuildPanel" pkg/wasmui/overlays.go && echo "Panel implemented"
  ```

---

## Priority Order Summary

| Step | Priority | Effort | Goal Impact |
|------|----------|--------|-------------|
| 1. Fix Stun/Root Effects | Critical | 2-3h | Combat functionality |
| 2. Wire Morale to UI | High | 2-3h | Gold Box authenticity |
| 3. Effect Display on Tokens | High | 2-3h | Combat feedback |
| 4. Attack Roll Narration | High | 2-3h | Gold Box authenticity |
| 5. Refactor First-Person View | Medium | 3-4h | Maintainability |
| 6. Wire First-Person to Map | High | 4-6h | Core gameplay |
| 7-9. API Documentation | High | 4-6h | Developer experience |
| 10. REST HP Restoration | Medium | 1-2h | Gameplay loop |
| 11. Faction UI | Medium | 2-3h | Game system wiring |
| 12. Guild UI | Medium | 2-3h | Game system wiring |

**Total Estimated Effort**: 27-38 hours

---

## Validation Commands Summary

```bash
# Run all tests
go test ./pkg/... -race

# Check complexity after refactoring
go-stats-generator analyze ./pkg/wasmui/ --skip-tests --format json 2>/dev/null | \
  python3 -c "import json,sys; d=json.load(sys.stdin); high=[f for f in d['functions'] if f['complexity']['cyclomatic']>9]; print(f'High complexity: {len(high)}')"

# Verify WASM builds
GOOS=js GOARCH=wasm go build -o /dev/null ./cmd/wasm-ui

# Count documented RPC methods
grep -c "^### " pkg/README-RPC.md

# Verify effect handlers have checks
grep -E "HasEffect.*Stun|HasEffect.*Root" pkg/server/handlers.go
```

---

## Notes

- All steps are independently testable and can be parallelized except Step 6 (depends on Step 5).
- Steps 7-9 (documentation) can be done in parallel with Steps 1-4 (code changes).
- The project has excellent test coverage (65-96%); maintain this by adding tests for all changes.
- WASM builds should be verified after each UI change: `GOOS=js GOARCH=wasm go build -o /dev/null ./cmd/wasm-ui`

---

*Generated: 2026-03-20 | Based on go-stats-generator v1.0.0 analysis*
