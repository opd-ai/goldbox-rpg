# Implementation Plan: Critical Bug Fixes & Gold Box UI Authenticity

## Project Context
- **What it does**: A modern Go-based RPG engine providing turn-based combat, character management, and world interactions through JSON-RPC/WebSocket APIs, inspired by the classic SSI Gold Box series.
- **Current goal**: Fix critical behavioral bugs in combat effects (Stun/Root) and WebSocket concurrency, then achieve "Gold Box faithful" UI visual authenticity.
- **Estimated Scope**: Large (204 functions above complexity threshold 9.0)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Combat conditions (Stun, Root) work correctly | ❌ Empty case statements | Yes - P0 |
| WebSocket session thread safety | ⚠️ Race conditions documented | Yes - P0 |
| Effect system resistance handling | ❌ API exists but non-functional | Yes - P1 |
| Healing modifier stacking | ❌ Incorrect initialization | Yes - P1 |
| Multiplicative modifier math | ❌ Formula bug | Yes - P1 |
| Gold Box visual panel borders | ⚠️ Partial | Yes |
| First-person dungeon viewport | ❌ Not implemented | Yes |
| Movement/attack range highlighting | ❌ Not implemented | Yes |
| Combat damage flash animation | ❌ Not implemented | Yes |
| Message log integration | ⚠️ Partial | Yes |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 204 functions above threshold (>9.0)
  - `wasmui`: 52 high-complexity functions (UI-critical)
  - `server`: 35 high-complexity functions (includes WebSocket handlers)
  - `pcg`: 33 high-complexity functions
  - `game`: 23 high-complexity functions (effect system)
- **Duplication ratio**: 1.26% (44 clone pairs, 918 duplicated lines — healthy)
- **Doc coverage**: 87.8% overall (functions: 94.2%, methods: 82.9%)
- **Package coupling**: server→game (expected), wasmui→server (RPC client), pcg self-contained

## Implementation Steps

---

### Step 1: Implement Stun Effect Behavior (P0)

- **Deliverable**: Stun effect prevents all actions (move, attack, cast spell) when applied
- **Dependencies**: None
- **Goal Impact**: Fixes critical combat mechanic — Stun condition will actually stun
- **Files to modify**:
  - `pkg/server/handlers.go`: Add `HasEffect(game.EffectStun)` check in `handleMove()`, `handleAttack()`, `handleCastSpell()`
  - `pkg/game/effectbehavior.go`: Populate the empty `EffectStun` case at line 497 (set action points to 0)
  - `pkg/game/effectbehavior_test.go`: Add `TestStunPreventsActions`
- **Acceptance**: Test passes verifying stunned entities cannot perform actions
- **Validation**: `go test -run TestStun ./pkg/game/... ./pkg/server/...`

---

### Step 2: Implement Root Effect Behavior (P0)

- **Deliverable**: Root effect prevents movement but allows other actions
- **Dependencies**: None (parallel with Step 1)
- **Goal Impact**: Fixes critical crowd-control mechanic — Root condition will actually root
- **Files to modify**:
  - `pkg/server/handlers.go`: Add `HasEffect(game.EffectRoot)` check in `handleMove()` only
  - `pkg/game/effectbehavior.go`: Populate the empty `EffectRoot` case at line 494
  - `pkg/server/handlers_test.go`: Add `TestRootPreventsMovement`, `TestRootAllowsAttack`
- **Acceptance**: Test passes verifying rooted entities cannot move but can attack/cast
- **Validation**: `go test -run TestRoot ./pkg/server/...`

---

### Step 3: Fix WebSocket Session Thread Safety (P0)

- **Deliverable**: Eliminate race conditions in WebSocket session management
- **Dependencies**: None (parallel with Steps 1-2)
- **Goal Impact**: Prevents production crashes during concurrent WebSocket operations
- **Files to modify**:
  - `pkg/server/websocket.go`: Wrap `session.Connected = false; session.WSConn = nil` modifications with mutex at lines 254-267
  - `pkg/server/session.go`: Add reference counting for broadcast snapshot safety
  - `pkg/server/websocket_test.go`: Add `TestConcurrentDisconnectDuringBroadcast`
- **Acceptance**: Race detector passes with 100 iterations of concurrent disconnect test
- **Validation**: `go test -race -count=100 ./pkg/server/...`

---

### Step 4: Initialize Healing Modifier Correctly (P1)

- **Deliverable**: Healing-over-time effects work correctly without healing debuffs present
- **Dependencies**: None
- **Goal Impact**: Fixes effect system stacking — HoT effects will apply properly
- **Files to modify**:
  - `pkg/game/effects.go`: Set `healingModifier = 1.0` in `NewEffectManager()` at line 236
  - `pkg/game/effectbehavior.go`: Remove conditional check at line 484-485, apply modifier directly
  - `pkg/game/effectbehavior_test.go`: Add `TestHealOverTimeWithoutDebuff`
- **Acceptance**: Test passes verifying 100% healing when no debuff is active
- **Validation**: `go test -run TestHeal ./pkg/game/...`

---

### Step 5: Fix Multiplicative Modifier Stacking (P1)

- **Deliverable**: Multiple multiplicative buffs stack correctly (1.2x × 1.2x = 1.44x, not 2.64x)
- **Dependencies**: Step 4 (effect system consistency)
- **Goal Impact**: Fixes game balance — buff stacking won't produce wildly inflated stats
- **Files to modify**:
  - `pkg/game/effectmanager.go`: Fix formula at line 341 — initialize `multMods[stat] = 1.0`, use `multMods[mod.Stat] *= mod.Value`
  - `pkg/game/effectmanager_test.go`: Add `TestMultiplicativeStackingCorrect` with two 1.2x buffs → 1.44x
- **Acceptance**: Test passes verifying correct multiplicative stacking math
- **Validation**: `go test -run TestMultiplicative ./pkg/game/...`

---

### Step 6: Implement Resistance API (P1)

- **Deliverable**: Public API to set and retrieve resistance values; resistance actually reduces damage
- **Dependencies**: Steps 4-5 (effect system fixes)
- **Goal Impact**: Makes "Immunity and resistance handling" functional as documented
- **Files to modify**:
  - `pkg/game/effects.go`: Add `SetResistance(effectType EffectType, value float64) error` and `GetResistance(effectType EffectType) float64`
  - `pkg/game/effectbehavior.go`: Verify `getResistanceForDamageType()` uses populated map
  - `pkg/game/effectbehavior_test.go`: Add `TestResistanceReducesDamage`
- **Acceptance**: Test passes verifying 50% fire resistance halves fire damage
- **Validation**: `go test -run TestResistance ./pkg/game/...`

---

### Step 7: EGA-Style Bold Panel Borders

- **Deliverable**: All UI panels have double-pixel bright-colored borders matching Gold Box aesthetic
- **Dependencies**: None (UI parallel track)
- **Goal Impact**: Foundation visual fix for Gold Box authenticity — applies to all screens
- **Files to modify**:
  - `pkg/wasmui/types_ui.go`: Define `BorderThickness = 2`, verify `ColorPanelBorderHi` is bright
  - `pkg/wasmui/exploration.go`: Update `drawViewport()`, `drawMinimap()`, `drawCombatLog()` to use double-border
  - `pkg/wasmui/combat_screen.go`: Update `drawCombatGrid()`, `drawInitiativePanel()` borders
  - `pkg/wasmui/overlays.go`: Update overlay panels
- **Acceptance**: Visual inspection shows all panels with Gold Box-style double borders
- **Validation**: `make wasm && make test-browser` (screenshot comparison)

---

### Step 8: Route All Feedback to Message Log

- **Deliverable**: All game feedback (damage, misses, spells, errors) appears in the scrolling log panel
- **Dependencies**: Step 7 (panel structure)
- **Goal Impact**: Eliminates floating text, enforces Gold Box information paradigm
- **Files to modify**:
  - `pkg/wasmui/exploration.go`: Ensure `addLogMessage()` called for all combat/interaction results
  - `pkg/wasmui/combat_screen.go`: Route attack results, spell results to log via `addLogMessage()`
  - `pkg/wasmui/overlays.go`: Remove any floating text overlays, use log instead
- **Acceptance**: No floating text anywhere; all feedback appears in log panel
- **Validation**: `make test-browser` playthrough captures showing log-only feedback

---

### Step 9: Damage Flash Animation

- **Deliverable**: Entities flash red when damaged, green when healed (200ms duration)
- **Dependencies**: None
- **Goal Impact**: High-impact visual feedback for combat, simple implementation
- **Files to modify**:
  - `pkg/wasmui/types_game.go`: Add `DamageFlash` struct with `entityID`, `startTime`, `duration`, `color`
  - `pkg/wasmui/combat_screen.go`: Add `damageFlashes []DamageFlash` to Game; in `executeAttack()` add flash on hit; in `drawCombatGrid()` apply color tint to flashing entities
- **Acceptance**: Combat hits produce visible red flash on target for ~200ms
- **Validation**: `make test-browser` combat screenshot showing entity flash state

---

### Step 10: Movement Range Highlighting

- **Deliverable**: Move mode highlights all reachable tiles in blue overlay based on AP
- **Dependencies**: None
- **Goal Impact**: Core combat UX — players see where they can move
- **Files to modify**:
  - `pkg/wasmui/combat_screen.go`: Add `getMovementRange(playerX, playerY, ap int) []Position`; in `drawCombatGrid()`, if `CombatActionMove`, draw blue overlay on reachable tiles
- **Acceptance**: Entering move mode shows blue-highlighted tiles within movement range
- **Validation**: `make test-browser` screenshot showing move mode highlights

---

### Step 11: Attack Range Highlighting

- **Deliverable**: Attack mode highlights tiles within weapon range in red overlay
- **Dependencies**: Step 10 (similar implementation)
- **Goal Impact**: Combat UX — players see attack range for equipped weapon
- **Files to modify**:
  - `pkg/wasmui/combat_screen.go`: Add `getAttackRange()` similar to movement; draw red overlay when `CombatActionAttack`
  - `pkg/wasmui/types_rpc.go`: Ensure `WeaponRange` available from player state
- **Acceptance**: Entering attack mode shows red-highlighted tiles within weapon range
- **Validation**: `make test-browser` screenshot showing attack mode highlights

---

### Step 12: Active Character Tile Highlight

- **Deliverable**: Current turn character has pulsing gold border on their tile
- **Dependencies**: Steps 9-11 (combat grid rendering)
- **Goal Impact**: Combat clarity — immediately visible whose turn it is
- **Files to modify**:
  - `pkg/wasmui/combat_screen.go`: In `drawCombatGrid()`, identify current turn entity, draw oscillating gold border around tile
- **Acceptance**: Current turn entity has visible pulsing highlight
- **Validation**: `make test-browser` screenshot showing active character highlight

---

### Step 13: NPC Morale Indicator

- **Deliverable**: Initiative panel shows morale state icons for NPCs; players can see when enemies are close to fleeing
- **Dependencies**: Step 12 (UI polish consistency)
- **Goal Impact**: Surfaces hidden backend system (morale exists in `pkg/game/morale.go` but invisible)
- **Files to modify**:
  - `pkg/wasmui/types_game.go`: Add `MoraleState string` to `InitiativeEntry`
  - `pkg/wasmui/rpc_methods.go`: Include morale in combat state response
  - `pkg/wasmui/combat_screen.go`: In `drawInitiativePanel()`, show morale icon by NPC name
- **Acceptance**: NPC initiative entries show Steadfast/Shaken/Broken/Panicked indicators
- **Validation**: `make test-browser` combat screenshot showing morale icons

---

### Step 14: First-Person Dungeon Viewport (Large)

- **Deliverable**: Exploration viewport renders first-person corridor view with walls at 3 depth levels
- **Dependencies**: Steps 7-8 (panel infrastructure)
- **Goal Impact**: Major visual overhaul achieving core Gold Box exploration aesthetic
- **Files to modify**:
  - `pkg/wasmui/exploration.go`: Replace grid rendering with `drawFirstPersonView()`; implement raycasting or depth slices; add facing direction state
  - `pkg/wasmui/types_game.go`: Add `Facing int` to `PlayerState`
  - `pkg/wasmui/exploration.go`: Add Q/E keybindings for turn-left/turn-right
- **Acceptance**: Exploration shows first-person view with visible walls and doors
- **Validation**: `make test-browser` exploration screenshot showing first-person perspective

---

### Step 15: Reduce UI Complexity Hotspots (Maintenance)

- **Deliverable**: `drawCombatGrid`, `updateCharCreationName`, `Draw` (adventure_ui) complexity < 15
- **Dependencies**: Steps 9-14 (combat grid refactoring will touch these)
- **Goal Impact**: Maintainability — easier to modify and extend UI code
- **Files to modify**:
  - `pkg/wasmui/combat_screen.go`: Extract `drawCombatFloor()`, `drawCombatEntities()`, `drawCombatHighlights()` from `drawCombatGrid`
  - `pkg/wasmui/character_creation.go`: Extract keyboard handling helper from `updateCharCreationName`
  - `pkg/wasmui/adventure_ui.go`: Split `Draw` into `drawAdventureList()`, `drawAdventureDetail()`
- **Acceptance**: No function in wasmui exceeds complexity 15
- **Validation**: `go-stats-generator analyze ./pkg/wasmui --skip-tests --format json | jq '[.functions[] | select(.complexity.overall > 15)] | length'` returns 0

---

### Step 16: Improve Server Package Test Coverage (Maintenance)

- **Deliverable**: `pkg/server` test coverage ≥ 85%
- **Dependencies**: Steps 1-3 (effect checks and WebSocket fixes add testable code)
- **Goal Impact**: Critical network layer reliability for production
- **Files to modify**:
  - `pkg/server/handlers_editor_test.go`: Add tests for `handleQuestEditorUpdate`
  - `pkg/server/session_test.go`: Add concurrent session tests
  - `pkg/server/websocket_test.go`: Add reconnection edge case tests
- **Acceptance**: Coverage report shows ≥ 85% for pkg/server
- **Validation**: `go test -cover ./pkg/server/... | grep coverage`

---

## Tiebreaker Rationale

The most impactful unachieved goals are:

1. **Critical bugs** (Steps 1-6): Stun/Root/resistance are core combat mechanics that don't work. These break tactical gameplay expectations and must be fixed before any UI polish.

2. **Gold Box visual authenticity** (Steps 7-14): The project explicitly targets "Gold Box faithful" UI. The ROADMAP.md provides detailed specifications for first-person viewport, panel borders, combat highlighting, and morale visibility.

3. **Maintainability** (Steps 15-16): High complexity in wasmui (52 functions > 9.0) and lower server coverage (78.3%) create technical debt that slows future development.

The plan orders work by:
1. **Dependency**: Bug fixes first (independent), then UI infrastructure, then advanced features
2. **Impact**: P0 bugs → P1 effect system → Foundation visuals → Combat UX → Polish → Maintenance

---

## Validation Commands Summary

```bash
# After Steps 1-6 (Bug Fixes)
go test -race ./pkg/game/... ./pkg/server/...

# After Step 7-14 (UI)
make wasm && make test-browser

# After Step 15 (Complexity)
go-stats-generator analyze ./pkg/wasmui --skip-tests --format json --sections functions | jq '[.functions[] | select(.complexity.overall > 15)] | length'

# After Step 16 (Coverage)
go test -cover ./pkg/server/... | grep -E '^ok.*coverage:'
```

---

*Generated: 2026-03-19 | Based on go-stats-generator metrics and project GAPS.md/ROADMAP.md*
