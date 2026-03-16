# Implementation Gaps — 2026-03-16

## Browser Editor Keyboard Shortcuts

- **Stated Goal**: Documentation at `docs/EDITOR_GUIDE.md:173-183` lists keyboard shortcuts including Ctrl+S (save), Ctrl+Z (undo), Ctrl+Y (redo), and terrain shortcuts (G=grass, W=water, S=stone, D=dirt).
- **Current State**: The documentation explicitly states "To be implemented in future versions". No keyboard event handlers exist in `pkg/wasmui/editor.go` or `pkg/wasmui/quest_editor.go` for these shortcuts.
- **Impact**: Users expecting the documented shortcuts will find them non-functional. This affects workflow efficiency for map creators who rely on keyboard shortcuts for rapid editing.
- **Closing the Gap**:
  1. Add keyboard event handlers in `pkg/wasmui/editor.go` using `ebiten.IsKeyPressed()` and `inpututil.IsKeyJustPressed()` patterns already used elsewhere in the WASM UI
  2. Implement undo/redo stack for map operations (requires tile history tracking)
  3. Wire terrain shortcuts to tile palette selection
  4. Remove "To be implemented" disclaimer from `docs/EDITOR_GUIDE.md:175`
  5. Validation: Manual testing at `/editor.html` and `/quest-builder.html`

## High-Complexity WASM UI Functions

- **Stated Goal**: Project guidelines specify cyclomatic complexity should not exceed 15 for maintainability.
- **Current State**: Six functions in `pkg/wasmui/` exceed this threshold:
  - `updateCharCreationAttributes` (CC=19.2) at `pkg/wasmui/character_creation.go:70`
  - `updateExploration` (CC=17.9) at `pkg/wasmui/exploration.go`
  - `updateInventory` (CC=17.1) at `pkg/wasmui/overlay_screens.go`
  - `updateSpellbook` (CC=15.8) at `pkg/wasmui/overlay_screens.go`
  - `updateMainMenu` (CC=15.8) at `pkg/wasmui/screens.go`
  - `updateCombat` (CC=15.3) at `pkg/wasmui/combat_screen.go`
- **Impact**: High complexity increases bug risk and maintenance burden. These are user-facing UI functions where bugs directly affect player experience.
- **Closing the Gap**:
  1. Extract state-machine branches into dedicated handler functions (e.g., `handleAttributeRoll()`, `handleAttributeReroll()`, `handleAttributeConfirm()`)
  2. Move input processing logic into separate methods with single responsibilities
  3. Use strategy pattern for different menu/screen states instead of large switch statements
  4. Target: All functions below CC=15
  5. Validation: `go-stats-generator analyze . --skip-tests | grep "High Complexity"` returns fewer than 3 functions

## Oversized Handler File

- **Stated Goal**: Project conventions suggest file sizes that enable easy navigation and review.
- **Current State**: `pkg/server/handlers.go` contains 1,797 lines and 56 functions, making it the second-largest burden score (2.40) in the codebase.
- **Impact**: Large file size impedes code review efficiency, increases merge conflict likelihood, and makes it harder to locate specific handler logic.
- **Closing the Gap**:
  1. Split handlers by RPC category into focused files:
     - `handlers_character.go` — createCharacter, getGameState, move, attack
     - `handlers_combat.go` — startCombat, endTurn, castSpell, applyEffect
     - `handlers_quest.go` — quest management methods (8 handlers)
     - `handlers_guild.go` — guild and faction methods (18 handlers)
     - `handlers_editor.go` — editor.* and questEditor.* methods (9 handlers)
     - `handlers_pcg.go` — procedural content generation methods (7 handlers)
  2. Keep shared helper functions in `handlers_common.go`
  3. Update imports in each new file
  4. Validation: `go build ./pkg/server && go test ./pkg/server/...`

## WASM Character Creation Duplication

- **Stated Goal**: DRY principle — avoid code duplication to reduce maintenance burden.
- **Current State**: `pkg/wasmui/character_creation.go` contains a 6-line escape-key handling pattern duplicated 4 times at lines 72-77, 133-138, 214-219, and at least one more location.
- **Impact**: Bug fixes or behavior changes to escape handling must be applied in 4+ places, risking inconsistency.
- **Closing the Gap**:
  1. Extract common escape handling to:
     ```go
     func (g *Game) handleCharCreationEscape() bool {
         if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
             g.mu.Lock()
             g.mode = ModeNormal
             g.screenState = ScreenMainMenu
             g.charCreation = CharCreationState{}
             g.mu.Unlock()
             return true
         }
         return false
     }
     ```
  2. Replace duplicated blocks with `if g.handleCharCreationEscape() { return }`
  3. Validation: `go-stats-generator analyze . --skip-tests | grep character_creation` shows reduced duplication

## Archived Dependency in Test Suite

- **Stated Goal**: Use maintained dependencies; project explicitly documents gorilla/websocket as test-only.
- **Current State**: `go.mod:15` retains `github.com/gorilla/websocket v1.5.3` which has been archived since 2022. The go.mod includes a comment acknowledging this is acceptable for test-only usage.
- **Impact**: No security risk (test-only), but represents technical debt and may trigger dependency scanner alerts.
- **Closing the Gap**:
  1. Migrate `test/e2e/client.go` WebSocket client from gorilla/websocket to coder/websocket API
  2. Update test client connection logic:
     ```go
     // Before (gorilla)
     conn, _, err := websocket.DefaultDialer.Dial(url, nil)
     // After (coder)
     conn, _, err := websocket.Dial(ctx, url, nil)
     ```
  3. Migrate `pkg/server/benchmark_test.go` if it uses gorilla directly
  4. Run `go mod tidy` to remove unused dependency
  5. Validation: `go mod graph | grep gorilla` returns empty

## Types File Organization

- **Stated Goal**: Cohesive file organization with related types grouped together.
- **Current State**: `pkg/wasmui/types.go` contains 409 lines and 31 type definitions with a burden score of 2.59 (highest in codebase). The file mixes game state types, UI component types, and RPC structures.
- **Impact**: Difficult to locate specific type definitions; high merge conflict probability when multiple developers modify types.
- **Closing the Gap**:
  1. Split into domain-focused files:
     - `types_game.go` — GameState, Character, Position, Combat-related types
     - `types_ui.go` — Screen states, UI component configs, overlay types
     - `types_rpc.go` — Request/response structures for JSON-RPC communication
  2. Maintain a single `types.go` with shared primitives and constants if needed
  3. Validation: `wc -l pkg/wasmui/types*.go` shows each file under 150 lines

## Method Documentation Gap

- **Stated Goal**: Comprehensive godoc coverage for exported APIs.
- **Current State**: go-stats-generator reports 81.9% method documentation coverage vs 94.3% for functions. The gap primarily exists in `pkg/server/` helper methods and `pkg/pcg/` internal methods.
- **Impact**: Developers integrating with or maintaining these packages have less guidance on method behavior and parameters.
- **Closing the Gap**:
  1. Identify undocumented methods: `grep -rn "^func.*\).*(" pkg/server/ pkg/pcg/ | grep -v "^.*//"`
  2. Add godoc comments following project conventions:
     ```go
     // methodName performs X for Y.
     // Parameters:
     //   - param1: description
     // Returns: description
     ```
  3. Focus on methods with 2+ parameters or non-obvious return values
  4. Validation: `go-stats-generator analyze . --skip-tests | grep "Method Coverage"` shows >90%

---

## Summary

| Gap | Severity | Effort | Priority |
|-----|----------|--------|----------|
| Keyboard shortcuts not implemented | MEDIUM | 4-8 hours | 1 |
| High-complexity WASM functions | HIGH | 4-6 hours | 2 |
| Oversized handlers.go | HIGH | 2-3 hours | 3 |
| Character creation duplication | MEDIUM | 1 hour | 4 |
| Archived test dependency | MEDIUM | 2-3 hours | 5 |
| Types file organization | LOW | 1-2 hours | 6 |
| Method documentation gap | LOW | 2-4 hours | 7 |

**Total estimated effort to close all gaps: 16-27 hours**

---

*Generated: 2026-03-16*
