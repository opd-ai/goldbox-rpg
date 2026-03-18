# Implementation Gaps — 2026-03-18

This document identifies gaps between the GoldBox RPG Engine's stated goals and its actual implementation. Each gap includes the impact on users and specific steps to close it.

---

## Android Service Test Coverage Below Minimum

- **Stated Goal**: README badge claims "coverage-65-96%" implying all packages meet the 65% minimum threshold.
- **Current State**: `cmd/android-service/webservice.go` has only 36.1% test coverage, containing:
  - `main()` (64 lines, complexity 17.9) — server bootstrap and signal handling
  - `bootstrapGame()` (13 lines) — game initialization
  - `getLANIP()` (24 lines, complexity 11.9) — network interface enumeration
  - `configureLogging()` (5 lines) — logging setup
- **Impact**: The coverage claim in README is technically inaccurate. Changes to the Android service entry point cannot be validated automatically, risking Android-specific regressions.
- **Closing the Gap**:
  1. Add unit tests for `bootstrapGame()` function with mocked server
  2. Add unit tests for `getLANIP()` with mocked network interfaces
  3. Add signal handling tests verifying graceful shutdown on SIGINT/SIGTERM
  4. Target: ≥65% coverage to match README claim
  5. **Validation**: `go test -cover ./cmd/android-service/...` reports ≥65%

---

## Browser Playtest Timeout in Headless Chrome

- **Stated Goal**: Automated browser playtests capture screenshots at key game states (splash screen, main menu, character creation, gameplay) for nightly release artifacts.
- **Current State**: `test/browser/browser_playtest_test.go` fails with "context deadline exceeded" when detecting the Ebitengine canvas:
  ```
  browser_playtest_test.go:362: screenshot 03-canvas-fail failed: context deadline exceeded
  browser_playtest_test.go:363: Ebitengine canvas not found: err=context deadline exceeded exists=false
  ```
  The WASM binary builds successfully and the server reports healthy, but the canvas element is not detected within the timeout window.
- **Impact**: CI cannot reliably capture gameplay screenshots. Nightly releases may not include screenshot artifacts. Manual testing is required to verify the WASM client works correctly.
- **Closing the Gap**:
  1. Increase canvas detection timeout from current value to 60 seconds
  2. Add explicit WebSocket connection state polling before canvas checks
  3. Add retry logic for canvas element detection with exponential backoff
  4. Consider adding a WASM-side health indicator (e.g., global variable set after Ebitengine init)
  5. Test in non-headless Chrome to isolate headless-specific issues
  6. **Validation**: `go test ./test/browser/... -v -timeout 5m` passes consistently

---

## WASM UI Complexity Hotspots

- **Stated Goal**: The project aims for maintainable, well-structured code as evidenced by its comprehensive CI checks, linting, and quality gates.
- **Current State**: Five functions in `pkg/wasmui/` have cyclomatic complexity >15:
  | Function | File | Lines | Complexity |
  |----------|------|-------|------------|
  | `updateCharCreationClass` | character_creation.go | 74 | 25.9 |
  | `drawQuestLogOverlay` | overlays.go | 91 | 20.7 |
  | `updateCharCreationAttributes` | character_creation.go | 78 | 20.2 |
  | `updateMainMenu` | screens.go | 72 | 19.7 |
  | `drawCharCreationReview` | character_creation.go | 99 | 18.4 |
- **Impact**: High complexity in user-facing UI code creates:
  - Difficult bug diagnosis when users report UI issues
  - Higher regression risk when adding features
  - Harder onboarding for new contributors to the frontend
- **Closing the Gap**:
  1. **Refactor `updateCharCreationClass`**: Extract `handleClassSelection()`, `validateClassChoice()`, `updateClassPreview()` as helper functions
  2. **Refactor `drawQuestLogOverlay`**: Extract `drawQuestList()`, `drawQuestDetails()`, `handleQuestLogScroll()`
  3. **Refactor `updateCharCreationAttributes`**: Extract `handleAttributeInput()`, `validateAttributeRange()`, `calculateDerivedStats()`
  4. **Refactor `updateMainMenu`**: Extract `handleMenuNavigation()`, `processMenuSelection()`, `updateMenuAnimations()`
  5. Target: Each function should have complexity ≤15
  6. **Validation**: `go-stats-generator analyze ./pkg/wasmui --skip-tests | grep -c "complexity > 15"` returns ≤2

---

## Low Package Cohesion

- **Stated Goal**: Well-organized package structure with clear separation of concerns (per Architecture section in README).
- **Current State**: Three packages have cohesion scores significantly below the 2.0 threshold:
  | Package | Cohesion | Files | Functions |
  |---------|----------|-------|-----------|
  | `pkg/secrets` | 0.7 | 3 | 7 |
  | `pkg/cliutil` | 0.8 | 2 | 7 |
  | `pkg/persistence` | 1.0 | 8 | 34 |
- **Impact**: Low cohesion indicates functions within a package may not be strongly related, making the codebase harder to navigate for new contributors.
- **Closing the Gap**:
  1. **pkg/secrets**: Evaluate whether functionality should merge into `pkg/config`. If secret management is distinct from configuration, add `doc.go` explaining the boundary.
  2. **pkg/cliutil**: Verify each utility is used by ≥2 CLI tools. If not, inline helpers into the specific command. If yes, expand scope and document.
  3. **pkg/persistence**: Review 8 files for 34 functions. Consolidate platform-specific lock files with main lock logic. Merge related save/load file pairs.
  4. **Validation**: Each package should have a `doc.go` explaining its purpose and boundaries. `go build ./...` succeeds.

---

## Functions Exceeding Length Guidelines

- **Stated Goal**: Maintainable codebase with reasonable function lengths for readability and testability.
- **Current State**: 111 functions (4.3%) exceed 50 lines. The guideline target is <3% (78 functions). Longest function: `saveItemFiles` at 118 lines.
- **Impact**: Long functions are harder to test in isolation, more prone to bugs, and increase cognitive load for reviewers.
- **Closing the Gap**:
  1. Audit the 33 functions that need reduction to reach 3% target
  2. Prioritize functions with BOTH high length AND high complexity
  3. Document intentionally long functions (parsers, generators) with inline comments explaining why extraction isn't beneficial
  4. Extract logical sub-operations into helper functions
  5. **Validation**: `go-stats-generator analyze . --skip-tests | grep "Functions > 50 lines"` shows <3%

---

## CLI Tool Code Duplication

- **Stated Goal**: DRY (Don't Repeat Yourself) codebase with shared utilities for common patterns.
- **Current State**: 40 clone pairs detected totaling 915 duplicated lines (1.33% ratio). Notable duplications in CLI tools:
  - `cmd/bootstrap-demo/main.go:195-200` ↔ `cmd/map-editor/main.go:80-85` (6 lines)
  - `cmd/events-demo/main.go:349-355` ↔ `cmd/map-editor/main.go:511-517` ↔ `cmd/metrics-demo/main.go:257-263` (7 lines × 3)
  - `cmd/map-editor/main.go:82-88` ↔ `cmd/quest-builder/main.go:45-51` (7 lines)
  - `pkg/wasmui/editor.go:320-328` ↔ `pkg/wasmui/editor.go:333-341` (9 lines internal clone)
- **Impact**: Duplicated code creates maintenance burden—changes must be made in multiple places, increasing risk of inconsistency.
- **Closing the Gap**:
  1. Extract common CLI initialization patterns to `pkg/cliutil/`:
     - Configuration loading boilerplate
     - Logging setup patterns
     - Flag parsing conventions
     - Graceful shutdown handling
  2. DRY up `pkg/wasmui/editor.go` internal clones by extracting common pattern to helper
  3. Target duplication ratio: ≤1.2%
  4. **Validation**: `go-stats-generator analyze . --skip-tests | grep "Duplication Ratio"` shows ≤1.2%

---

## Summary Table

| Gap | Severity | Effort | Impact |
|-----|----------|--------|--------|
| Android Service Test Coverage | HIGH | 1 day | Coverage claim accuracy |
| Browser Playtest Timeout | CRITICAL | 1-2 days | CI reliability |
| WASM UI Complexity | HIGH | 2-3 days | Maintainability |
| Low Package Cohesion | MEDIUM | 1 day | Code navigation |
| Long Functions | MEDIUM | 2-3 days | Testability |
| CLI Code Duplication | LOW | 0.5 days | Maintenance burden |

---

## Priority Recommendation

1. **Immediate** (blocks CI): Browser Playtest Timeout
2. **Short-term** (accuracy): Android Service Test Coverage
3. **Medium-term** (maintainability): WASM UI Complexity, Package Cohesion
4. **Long-term** (code quality): Long Functions, CLI Duplication

---

*Companion to AUDIT.md | Generated: 2026-03-18*
