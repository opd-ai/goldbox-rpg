# Implementation Gaps — 2026-03-18

This document identifies gaps between the project's stated goals and current implementation. Overall, the GoldBox RPG Engine achieves excellent goal coverage with only maintenance-level gaps identified.

---

## Test Coverage Gap: Server Package

- **Stated Goal**: README claims "comprehensive integration tests" and CI enforces ≥60% coverage (README.md:184-185).

- **Current State**: `pkg/server/` has 78.3% coverage—above threshold but lowest among core packages (codebase average: 87%). Complex handlers lack dedicated edge-case tests:
  - `handleQuestEditorUpdate` (70 lines, complexity 14.0) — concurrent edit scenarios untested
  - `handleJoinGame` (118 lines) — WebSocket vs HTTP paths need differentiated tests
  - `handleEditorLoadMap` — map validation edge cases

- **Impact**: Server is critical infrastructure handling all client requests. Lower coverage increases risk of undetected regressions in session management and editor operations.

- **Closing the Gap**:
  1. Add table-driven tests for `handleJoinGame` in `pkg/server/handlers_test.go`:
     - Test concurrent session join scenarios
     - Test WebSocket session attachment path
     - Test HTTP new-session path
  2. Add tests for `handleQuestEditorUpdate`:
     - Test concurrent edit detection
     - Test quest validation failures
  3. **Target**: Raise `pkg/server/` coverage from 78.3% to 85%
  4. **Validate**: `go test -cover ./pkg/server/... | grep coverage`

---

## Test Coverage Gap: PCG Package

- **Stated Goal**: README states PCG should be "deterministic" with "validation system for generated content integrity" (README.md:71-72).

- **Current State**: `pkg/pcg/` has 78.9% coverage—above threshold but lowest among PCG-related packages. Related packages have better coverage:
  - `pkg/pcg/pcgutil`: 96.7%
  - `pkg/pcg/quests`: 92.5%
  - `pkg/pcg/levels`: 90.1%
  - `pkg/pcg/terrain`: 86.6%
  - `pkg/pcg/items`: 83.8%

- **Impact**: PCG is a core differentiator. Insufficient testing of seed reproducibility or content validation edge cases could lead to unpredictable content generation.

- **Closing the Gap**:
  1. Add deterministic seeding verification tests:
     - Confirm identical seeds produce identical output across runs
     - Test seed derivation with edge-case context parameters
  2. Add validation boundary tests:
     - Test constraint handling with conflicting requirements
     - Test validation with malformed content
  3. **Target**: Raise `pkg/pcg/` coverage from 78.9% to 85%
  4. **Validate**: `go test -cover ./pkg/pcg/... | grep coverage`

---

## Complexity Gap: UI State Machines

- **Stated Goal**: Maintainable, well-structured codebase (implicit in development guidelines).

- **Current State**: Three functions in `pkg/wasmui/` exceed complexity threshold 15:

  | Function | File | Lines | Complexity |
  |----------|------|-------|------------|
  | `drawCombatGrid` | `combat_screen.go` | 73 | 19.4 |
  | `updateCharCreationName` | `character_creation.go` | 59 | 17.1 |
  | `Draw` | `adventure_ui.go` | 99 | 15.8 |

  Codebase average complexity is 4.0, making these clear outliers.

- **Impact**: High complexity makes these functions harder to maintain, test, and extend. UI state machine complexity is common but manageable through refactoring.

- **Closing the Gap**:
  1. Refactor `drawCombatGrid` (`pkg/wasmui/combat_screen.go:73`):
     - Extract `drawGridCell()` for individual cell rendering
     - Extract highlight logic for selected/targeted cells
  2. Simplify `updateCharCreationName` (`pkg/wasmui/character_creation.go:59`):
     - Extract keyboard navigation to `handleNameKeyInput()`
     - Use table-driven approach for key mapping
  3. Extract helpers from `Draw` (`pkg/wasmui/adventure_ui.go:99`):
     - Split state-specific drawing into `drawAdventurePanel()`, `drawQuestPanel()`
  4. **Validate**: `go-stats-generator analyze . --skip-tests | grep -A5 "Top Complex"` shows no functions >15

---

## Code Organization Gap: Oversized Handler File

- **Stated Goal**: Clean package separation per project structure (README.md:279-305).

- **Current State**: `pkg/server/handlers.go` is 1,374 lines with 57 functions and maintenance burden score of 2.54 (highest in codebase). It contains handlers for combat, quests, spatial queries, equipment, and more.

- **Impact**: Large files with mixed concerns are harder to navigate, test in isolation, and maintain. Changes to combat handlers risk affecting quest handlers due to proximity.

- **Closing the Gap**:
  1. Split `handlers.go` into domain-specific files:
     - `handlers_combat.go` — attack, spell casting, turn management
     - `handlers_quest.go` — quest start, complete, update
     - `handlers_spatial.go` — object queries, range searches
     - `handlers_equipment.go` — equip, unequip, get equipment
  2. Keep core handlers (move, joinGame, getGameState) in `handlers.go`
  3. **Target**: `handlers.go` under 800 lines
  4. **Validate**: `wc -l pkg/server/handlers.go` shows <800

---

## Duplication Gap: Server Method Registration

- **Stated Goal**: Clean codebase with minimal duplication (1.34% ratio is excellent).

- **Current State**: `go-stats-generator` identifies duplicated blocks (29 lines) at `pkg/server/server.go:1031` in the method registration section. Similar patterns repeat for each RPC method registration.

- **Impact**: Duplicated registration patterns make it easy to introduce inconsistencies when adding new methods. Minor issue given overall low duplication.

- **Closing the Gap**:
  1. Consider extracting common registration pattern:
     ```go
     func (s *RPCServer) registerMethod(method RPCMethod, handler MethodHandler) {
         s.methodRegistry[method] = handler
         // Optional: log registration, add middleware hooks
     }
     ```
  2. Apply pattern to reduce repetitive `s.methodRegistry[Method...] = s.handle...` lines
  3. **Validate**: `go-stats-generator analyze . --skip-tests --sections duplication` shows no clone pairs >20 lines in server.go

---

## Documentation Gap: Go Version Badge

- **Stated Goal**: Accurate documentation (README badges).

- **Current State**: README.md:6 badge shows "Go Version 1.25.6" but `go.mod` specifies `go 1.25.6` with `toolchain go1.25.8`. Badge is correct but could include toolchain version for completeness.

- **Impact**: Minor. Version information is accurate but incomplete.

- **Closing the Gap**:
  1. Update badge to `go >=1.25.6 (toolchain 1.25.8)` or leave as-is
  2. **Validate**: Compare `go.mod` with README badge

---

## Summary

| Gap | Severity | Current State | Target State |
|-----|----------|---------------|--------------|
| Server test coverage | HIGH | 78.3% | 85% |
| PCG test coverage | HIGH | 78.9% | 85% |
| UI complexity hotspots | MEDIUM | 3 functions >15 | 0 functions >15 |
| Oversized handlers.go | MEDIUM | 1,374 lines | <800 lines |
| Server duplication | LOW | 29-line clone | No clones >20 |
| Go version badge | LOW | Correct but partial | Include toolchain |

**Overall Assessment**: The GoldBox RPG Engine achieves **100% of stated feature goals**. All identified gaps are maintenance improvements:

1. **Coverage gaps** are above the 60% threshold but below codebase average
2. **Complexity gaps** affect only UI code, not core game mechanics
3. **Organization gaps** are code hygiene, not functional issues

None of these gaps represent broken features or unmet promises. The codebase is production-quality with comprehensive testing (87% average coverage) and clean architecture (zero circular dependencies).

**Recommended Priority**:
1. Server test coverage (critical path)
2. PCG test coverage (core differentiator)
3. UI complexity refactoring (maintainability)
4. Handler file splitting (organization)
5. Other gaps (optional)
