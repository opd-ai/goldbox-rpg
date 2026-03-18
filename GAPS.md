# Implementation Gaps — 2026-03-18

This document identifies gaps between the project's stated goals and current implementation. Overall, the GoldBox RPG Engine achieves excellent goal coverage with only minor documentation discrepancies and coverage improvement opportunities.

---

## Documentation Discrepancy: Visual Editor Status

- **Stated Goal**: README.md:458-459 states "⚠️ World editor tools (CLI tools only, no GUI editors)" in the roadmap section, suggesting visual editors are incomplete.

- **Current State**: Fully functional WASM-based visual editors exist:
  - **Map Editor** (`web/editor.html`): Loads `editor.wasm`, provides canvas-based tile editing with splash screen, progress loading, and full Ebiten integration
  - **Quest Builder** (`web/quest-builder.html`): Form-based quest creation UI with objectives, rewards, validation, and status tracking
  - **WASM Implementation** (`pkg/wasmui/editor.go`): Complete Ebiten game implementation with 64+ methods including input handling, tool application, undo/redo, and rendering
  - **Server Handlers** (`pkg/server/handlers_editor.go`): RPC handlers for editor operations

- **Impact**: Users reading the README roadmap may believe visual editors don't exist when they are actually functional. This could cause users to skip exploring the browser-based editors or assume CLI is the only option.

- **Closing the Gap**: Update README.md:458-459 to accurately reflect the implemented state:
  ```markdown
  # Change from:
  - ⚠️ World editor tools (CLI tools only, no GUI editors)
  - ⚠️ Content creation utilities (CLI tools only, no visual editors)
  
  # To:
  - ✅ World editor tools (CLI + browser-based visual editors at /editor.html)
  - ✅ Content creation utilities (CLI + visual Quest Builder at /quest-builder.html)
  ```

---

## Test Coverage Gap: Server Package

- **Stated Goal**: Maintain ≥60% test coverage with comprehensive integration tests (README.md:184-185, ROADMAP.md:29).

- **Current State**: `pkg/server/` has 78.0% coverage—above the 60% threshold but lowest among core packages (average is 87%). Complex handlers lack dedicated test coverage:
  - `handleJoinGame`: 118 lines, handles session creation and concurrent scenarios
  - `handleEditorLoadMap`: 89 lines, complexity 17.1, processes map loading edge cases
  - `handleQuestEditorUpdate`: 70 lines, complexity 14.0, manages concurrent quest editing

- **Impact**: The server package is critical infrastructure handling all client requests. Lower coverage in complex handlers increases risk of undetected regressions in session management, editor operations, and concurrent access patterns.

- **Closing the Gap**:
  1. Add table-driven integration tests for `handleJoinGame`:
     - Cover session creation with valid/invalid parameters
     - Test concurrent join scenarios with race detection
     - Validate error handling paths
  2. Add tests for `handleEditorLoadMap`:
     - Test map format validation
     - Cover loading edge cases (empty maps, corrupted data)
     - Verify session validation logic
  3. Add tests for `handleQuestEditorUpdate`:
     - Test concurrent edit detection
     - Validate quest update validation logic
  4. Target: Raise `pkg/server` coverage from 78% to 85%
  5. Validate: `go test -cover ./pkg/server/... | grep coverage`

---

## Test Coverage Gap: PCG Package

- **Stated Goal**: Procedural content generation should be "deterministic" with "validation system for generated content integrity" (README.md:71-72).

- **Current State**: `pkg/pcg/` has 78.9% coverage—above threshold but lowest among PCG-related packages. The deterministic seeding system (`pkg/pcg/seed.go`) and content validation (`pkg/pcg/validator.go`) could benefit from additional edge case testing.

- **Impact**: PCG is a core differentiator of the engine. Insufficient testing of edge cases in terrain generation, content validation, or seed reproducibility could lead to unpredictable content or validation failures in production.

- **Closing the Gap**:
  1. Add edge case tests for terrain generation:
     - Test boundary conditions (1x1 maps, maximum sizes)
     - Validate biome transitions
  2. Add validation tests for generated content:
     - Verify all generated items/quests pass schema validation
     - Test constraint handling
  3. Add deterministic seeding verification:
     - Confirm identical seeds produce identical output across runs
     - Test seed derivation with various context parameters
  4. Target: Raise `pkg/pcg` coverage from 78.9% to 85%
  5. Validate: `go test -cover ./pkg/pcg/... | grep coverage`

---

## Complexity Gap: UI State Machines

- **Stated Goal**: Maintainable, well-structured codebase (implicit in development guidelines and CI quality gates).

- **Current State**: Seven functions in `pkg/wasmui/` exceed complexity threshold 15:
  | Function | Lines | Complexity |
  |----------|-------|------------|
  | `drawQuestLogOverlay` | 91 | 20.7 |
  | `updateCharCreationAttributes` | 78 | 20.2 |
  | `updateMainMenu` | 72 | 19.7 |
  | `drawCharCreationReview` | 99 | 18.4 |
  | `updateCharCreationName` | 59 | 17.1 |
  | `Draw` (adventure_ui) | 91 | 15.8 |

- **Impact**: High complexity makes these functions harder to maintain, test, and extend. UI state machine complexity is common but these functions are outliers relative to the codebase average (4.0).

- **Closing the Gap**:
  1. Extract helper functions from `drawQuestLogOverlay`:
     - Split quest list rendering from detail rendering
     - Create `drawQuestListItem()` helper
  2. Simplify `updateCharCreationAttributes`:
     - Extract attribute adjustment logic to dedicated function
     - Consider state pattern for character creation steps
  3. Refactor `updateMainMenu`:
     - Extract menu option handlers to separate functions
     - Use table-driven approach for menu items
  4. Validate: `go-stats-generator analyze . --skip-tests | grep -A20 "Top Complex Functions"` shows no functions with complexity >15

---

## BUG Annotation Gap

- **Stated Goal**: Clean codebase without unresolved issues (implicit in CI/quality gates).

- **Current State**: `go-stats-generator` reports 5 BUG annotations in the codebase. These appear to be documentation/logging notes rather than active bugs:
  - Logging-related notes in documentation files
  - Reproduction tracking comments
  - Server debugging notes

- **Impact**: BUG annotations can confuse developers and create uncertainty about code quality. If these are not actual bugs, they should be categorized differently.

- **Closing the Gap**:
  1. Review each BUG annotation:
     - If fixed: Remove the annotation
     - If tracking a known issue: Convert to TODO with GitHub issue reference
     - If documentation: Use NOTE or different terminology
  2. Validate: `grep -rn "BUG" pkg/ docs/ | wc -l` returns 0 unaddressed items

---

## Summary

| Gap Category | Severity | Current State | Target State |
|--------------|----------|---------------|--------------|
| Documentation discrepancy (visual editors) | LOW | README says "CLI only" | Update to reflect functional editors |
| Server test coverage | MEDIUM | 78.0% | 85% |
| PCG test coverage | MEDIUM | 78.9% | 85% |
| UI complexity | MEDIUM | 7 functions >15 | 0 functions >15 |
| BUG annotations | LOW | 5 annotations | 0 unaddressed |

**Overall Assessment**: The GoldBox RPG Engine achieves 100% of its stated feature goals. Identified gaps are primarily:
1. One documentation discrepancy (editors marked incomplete but are functional)
2. Coverage improvement opportunities in two packages (still above 60% threshold)
3. Optional complexity reduction in UI code
4. Housekeeping for BUG annotations

None of these gaps represent missing functionality or broken features. The codebase is production-quality with comprehensive testing and clean architecture.
