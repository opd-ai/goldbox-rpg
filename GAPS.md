# Implementation Gaps — 2026-03-17

This document identifies gaps between the GoldBox RPG Engine's stated goals and its actual implementation. Each gap includes the impact on users and specific steps to close it.

---

## Browser Editor Persistence

- **Stated Goal**: The README and `docs/EDITOR_GUIDE.md` describe browser-based Map Editor and Quest Builder as functional tools for visual content creation with real-time preview and export capabilities.
- **Current State**: While the HTML pages (`web/editor.html`, `web/quest-builder.html`) and WASM code (`pkg/wasmui/editor.go`, `map_editor.go`, `quest_editor.go`) exist, the server-side handlers are stubs:
  - `handleEditorLoadMap` (`pkg/server/handlers_editor.go:310`) generates a new UUID instead of loading actual map data
  - `handleEditorSaveMap` (`pkg/server/handlers_editor.go:236`) acknowledges success but doesn't persist
  - `handleQuestEditorList` (`pkg/server/handlers_quest_editor.go:180`) always returns an empty array
  - `handleQuestEditorCreate/Update/Delete` lack actual data storage
- **Impact**: Users attempting to use the browser editors will find their work is not saved. The editors appear functional but all changes are lost on page refresh or when attempting to load previously "saved" content.
- **Closing the Gap**:
  1. Implement `fileStore` integration in `handleEditorSaveMap` to persist map data to disk or database
  2. Implement actual file loading in `handleEditorLoadMap` to retrieve previously saved maps
  3. Connect quest editor handlers to the game state's quest registry or a persistence layer
  4. Add integration tests verifying round-trip save/load functionality
  5. Update `docs/EDITOR_GUIDE.md` to clarify current limitations until persistence is implemented
  6. Validation: `curl` test that saves a map, then loads it, and verifies data integrity

---

## Spell File Count Documentation

- **Stated Goal**: README claims "60+ spells across 11 YAML files" for the spell system.
- **Current State**: The `data/spells/` directory contains exactly 10 YAML files:
  - `cantrips.yaml` (level 0)
  - `level1.yaml` through `level9.yaml`
  - Total: 60 spell definitions
- **Impact**: Minor documentation inaccuracy. Users expecting 11 files may be confused when counting only 10. The spell count (60) is accurate.
- **Closing the Gap**:
  1. Update README.md line 449 from "11 YAML files" to "10 YAML files"
  2. Alternatively, if an 11th file was intended (e.g., `special.yaml` for unique spells), create it
  3. Validation: `ls -la data/spells/*.yaml | wc -l` returns the documented count

---

## Test Coverage Below Stated Minimum

- **Stated Goal**: README badge claims "coverage-78-96%" implying all packages have at least 78% test coverage.
- **Current State**: Several packages fall below the 78% threshold:
  - `test/e2e`: 65.5%
  - `scripts`: 68.8%
  - `cmd/quest-builder`: 71.6%
  - `cmd/server`: 71.8%
  - `cmd/android-service`: 0.0%
- **Impact**: The coverage badge misleads contributors and users about code quality. Lower coverage in E2E tests means less confidence in integration behavior.
- **Closing the Gap**:
  1. **Option A (Improve Coverage)**:
     - Add E2E tests for PCG integration, guild operations, faction diplomacy
     - Add tests for `scripts/verify_adventures.go` and `scripts/find_untested_files.go`
     - Add edge case tests for `cmd/quest-builder` (circular dependencies, invalid objectives)
     - Add startup/shutdown tests for `cmd/server`
     - Add basic tests for `cmd/android-service` (bootstrap, signal handling)
  2. **Option B (Update Documentation)**: Change badge to "coverage-65-96%" to reflect actual range
  3. Validation: `go test -cover ./... | grep coverage | sort -t: -k2 -n | head -5`

---

## Adventure Map Count Documentation

- **Stated Goal**: README claims "51 maps" in the embedded adventures.
- **Current State**: Actual map count across all 10 adventure packs is 100 maps.
- **Impact**: Positive discrepancy—users get nearly double the content advertised. However, understating content may discourage potential users who compare this to competitors.
- **Closing the Gap**:
  1. Update README.md to claim "100 maps" instead of "51 maps"
  2. Add automated CI check to keep content counts current
  3. Consider adding a `data/adventures/README.md` with detailed content inventory
  4. Validation: Script that counts map definitions and compares to README claims

---

## WASM UI Complexity Hotspots

- **Stated Goal**: The project aims for maintainable, well-structured code as evidenced by its CI checks, linting, and quality gates.
- **Current State**: The `pkg/wasmui/` package contains 4 functions with cyclomatic complexity >20:
  - `Update` (adventure_screen.go:50): complexity 31.9, 92 lines
  - `updateInventory` (overlays.go:26): complexity 31.4, 106 lines
  - `updateSpellbook` (overlays.go:299): complexity 28.8, 105 lines
  - `updateQuestLogOverlay` (overlays.go:537): complexity 27.5, 116 lines
- **Impact**: High complexity in UI code leads to:
  - Difficult bug diagnosis in user-facing features
  - Higher regression risk when adding features
  - Harder onboarding for new contributors
- **Closing the Gap**:
  1. **Refactor `updateInventory`**: Extract `handleInventoryInput()`, `updateInventoryDragDrop()`, `updateInventoryTooltip()` as helper functions
  2. **Refactor `updateSpellbook`**: Extract `handleSpellbookNavigation()`, `handleSpellSelection()`, `filterSpellList()`
  3. **Refactor `updateQuestLogOverlay`**: Extract `handleQuestLogInput()`, `updateQuestProgress()`, `renderQuestDetails()`
  4. **Refactor `AdventureScreen.Update`**: Extract state-specific update functions for each adventure phase
  5. Target: Each function should have complexity ≤15
  6. Validation: `go-stats-generator analyze ./pkg/wasmui --skip-tests --format json | jq '[.functions[] | select(.complexity.cyclomatic > 15)] | length'` returns ≤4

---

## Android Service Test Coverage

- **Stated Goal**: The project maintains ≥60% test coverage as a quality baseline (per CI configuration).
- **Current State**: `cmd/android-service/webservice.go` has 0% test coverage despite containing:
  - Server bootstrap logic
  - Signal handling for graceful shutdown
  - Network configuration (LAN IP detection)
  - Logging configuration
- **Impact**: Changes to the Android service entry point cannot be validated automatically. Production bugs on Android may not be caught before release.
- **Closing the Gap**:
  1. Add unit tests for `bootstrapGame()` function
  2. Add unit tests for `getLANIP()` function
  3. Add integration tests for signal handling
  4. Mock the server to test initialization flow
  5. Validation: `go test -cover ./cmd/android-service/...` reports ≥60%

---

## Low Package Cohesion

- **Stated Goal**: Well-organized package structure with clear separation of concerns (per Architecture section in README).
- **Current State**: Three packages have cohesion scores below 2.0:
  - `pkg/secrets`: 0.7 cohesion (3 files, 7 functions)
  - `pkg/cliutil`: 0.8 cohesion (2 files, 7 functions)
  - `pkg/persistence`: 1.0 cohesion (8 files, 34 functions)
- **Impact**: Low cohesion indicates functions within a package may not be strongly related, making the codebase harder to navigate and understand.
- **Closing the Gap**:
  1. **pkg/secrets**: Review if functionality should merge into `pkg/config`, or document clear separation rationale in `doc.go`
  2. **pkg/cliutil**: Consider inlining helpers into CLI commands if shared utility value is low, or expand scope to justify standalone package
  3. **pkg/persistence**: Review 8 files for 34 functions—consider consolidating related functionality (e.g., merge platform-specific lock files with main lock logic)
  4. Validation: Each package should have a `doc.go` explaining its purpose and boundaries

---

## Summary Table

| Gap | Severity | Effort to Close | Impact on Users |
|-----|----------|-----------------|-----------------|
| Browser Editor Persistence | CRITICAL | 2-3 days | Editors appear broken |
| Spell File Count Doc | LOW | 5 minutes | Minor confusion |
| Test Coverage Claim | HIGH | 3-4 days | Misleading quality signal |
| Adventure Map Count Doc | LOW | 5 minutes | Undersells content |
| WASM UI Complexity | HIGH | 2-3 days | Bug risk in UI |
| Android Service Tests | MEDIUM | 1 day | Android regression risk |
| Low Package Cohesion | LOW | 1 day | Navigation difficulty |

---

## Priority Recommendation

1. **Immediate** (blocks stated functionality): Browser Editor Persistence
2. **Short-term** (quality/accuracy): Test Coverage Claim, Documentation fixes
3. **Medium-term** (maintainability): WASM UI Complexity, Android Service Tests
4. **Long-term** (code organization): Package Cohesion improvements

---

*Generated: 2026-03-17 | Companion to AUDIT.md*
