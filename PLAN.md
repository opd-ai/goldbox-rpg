# Implementation Plan: Code Quality and Visual Editor Completion

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by SSI Gold Box games, providing turn-based combat, character management, procedural content generation, and real-time WebSocket communication for web-based RPG experiences.
- **Current goal**: Resolve documentation discrepancy regarding visual editors and reduce UI complexity hotspots
- **Estimated Scope**: Small (7 functions above complexity 9.0 threshold, 1.33% duplication ratio)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character System (6 attributes, 6 classes) | ✅ Achieved | No |
| Comprehensive Effect System | ✅ Achieved | No |
| Spatial Indexing (R-tree) | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Event-Driven Architecture | ✅ Achieved | No |
| WebSocket Real-time Communication | ✅ Achieved | No |
| Health Monitoring & Metrics | ✅ Achieved | No |
| System Resilience (circuit breakers) | ✅ Achieved | No |
| Asset Generation Pipeline (521 assets) | ✅ Achieved | No |
| Embedded Adventures (10 packs) | ✅ Achieved | No |
| Advanced NPC AI (A* pathfinding) | ✅ Achieved | No |
| Guild and Faction Systems | ✅ Achieved | No |
| Network Optimization | ✅ Achieved | No |
| Complete Spell System (60 spells) | ✅ Achieved | No |
| Visual Editors (Map/Quest) | ⚠️ Partial | **Yes** |
| Test Coverage ≥60% | ✅ 87% average | Yes (improve server to 85%) |

## Metrics Summary
- Complexity hotspots on goal-critical paths: **7 functions** above threshold (complexity >9.0)
- Duplication ratio: **1.33%** (40 clone pairs, 915 duplicated lines)
- Doc coverage: **87.5%** overall (94% functions, 91% types, 83% methods)
- Package coupling: Clean architecture with 0 circular dependencies
- Test coverage gaps: `pkg/server` (78%), `pkg/pcg` (78.9%)

## External Research Findings
- **coder/websocket**: Project has migrated to the recommended library. No known vulnerabilities in current version (v1.8.14). No action needed.
- **Ebitengine v2.9.9**: Current version is up-to-date. Best practices followed (Update/Draw/Layout pattern, go:embed for assets).
- **Go 1.25.6**: CHANGELOG notes 18 stdlib vulnerabilities requiring Go 1.24.12+ or 1.25.8. Consider upgrade when available.
- **Community**: No open issues found on GitHub. Project appears well-maintained.

---

## Implementation Steps

### Step 1: Clarify Visual Editor Status in README
- **Deliverable**: Update `README.md` roadmap section to accurately reflect visual editor state
- **Dependencies**: None
- **Goal Impact**: Resolves the only incomplete stated goal (visual editors documentation discrepancy)
- **Acceptance**: README roadmap items for visual editors match actual implementation; no contradictory statements about "CLI tools only"
- **Validation**: Manual review of `README.md` lines 329-356 and 458-461; confirm `web/editor.html` and `web/quest-builder.html` function as documented
- **Details**:
  - Files `editor.html`, `quest-builder.html` exist with functional WASM clients
  - README line 458-459 states "⚠️ World editor tools (CLI tools only, no GUI editors)"
  - README lines 329-356 document working browser-based editors
  - Either mark editors as complete (✅) or update the feature section to clarify partial state

### Step 2: Refactor `drawQuestLogOverlay` (Highest Complexity)
- **Deliverable**: Reduce complexity of `pkg/wasmui/overlays.go:drawQuestLogOverlay` from 20.7 to ≤9.0
- **Dependencies**: Step 1 (confirms editors are priority)
- **Goal Impact**: Maintainability of WASM frontend; establishes refactoring pattern for remaining UI functions
- **Acceptance**: Function complexity ≤9.0 as measured by go-stats-generator
- **Validation**: `go-stats-generator analyze . --skip-tests --format json | jq '.functions[] | select(.name == "drawQuestLogOverlay") | .complexity.overall'` returns ≤9.0
- **Details**:
  - Current: 91 lines, complexity 20.7
  - Split into: `drawQuestListSection()`, `drawQuestDetailSection()`, `drawQuestItemRow()`
  - Preserve existing behavior and visual output

### Step 3: Refactor `updateCharCreationAttributes`
- **Deliverable**: Reduce complexity of `pkg/wasmui/character_creation.go:updateCharCreationAttributes` from 20.2 to ≤9.0
- **Dependencies**: None (can parallelize with Step 2)
- **Goal Impact**: Maintainability of character creation flow
- **Acceptance**: Function complexity ≤9.0
- **Validation**: `go-stats-generator analyze . --skip-tests --format json | jq '.functions[] | select(.name == "updateCharCreationAttributes") | .complexity.overall'` returns ≤9.0
- **Details**:
  - Current: 78 lines, complexity 20.2
  - Extract: `handleAttributeIncrement()`, `handleAttributeDecrement()`, `validateAttributeBounds()`
  - Consider table-driven approach for attribute adjustment logic

### Step 4: Refactor `updateMainMenu`
- **Deliverable**: Reduce complexity of `pkg/wasmui/screens.go:updateMainMenu` from 19.7 to ≤9.0
- **Dependencies**: None (can parallelize with Steps 2-3)
- **Goal Impact**: Maintainability of main menu state machine
- **Acceptance**: Function complexity ≤9.0
- **Validation**: `go-stats-generator analyze . --skip-tests --format json | jq '.functions[] | select(.name == "updateMainMenu") | .complexity.overall'` returns ≤9.0
- **Details**:
  - Current: 72 lines, complexity 19.7
  - Use table-driven menu option handlers
  - Extract: `handleMenuSelection()`, menu option dispatch table

### Step 5: Refactor `drawCharCreationReview`
- **Deliverable**: Reduce complexity of `pkg/wasmui/character_creation.go:drawCharCreationReview` from 18.4 to ≤9.0
- **Dependencies**: Step 3 (same file, avoid conflicts)
- **Goal Impact**: Maintainability of character creation UI
- **Acceptance**: Function complexity ≤9.0
- **Validation**: `go-stats-generator analyze . --skip-tests --format json | jq '.functions[] | select(.name == "drawCharCreationReview") | .complexity.overall'` returns ≤9.0
- **Details**:
  - Current: 99 lines, complexity 18.4
  - Extract: `drawAttributeSummary()`, `drawClassSummary()`, `drawEquipmentSummary()`

### Step 6: Refactor `handleEditorLoadMap`
- **Deliverable**: Reduce complexity of `pkg/server/handlers_editor.go:handleEditorLoadMap` from 17.1 to ≤9.0
- **Dependencies**: Step 1 (confirms editor priority)
- **Goal Impact**: Server-side editor reliability; this is the only server function above threshold
- **Acceptance**: Function complexity ≤9.0
- **Validation**: `go-stats-generator analyze . --skip-tests --format json | jq '.functions[] | select(.name == "handleEditorLoadMap") | .complexity.overall'` returns ≤9.0
- **Details**:
  - Current: 89 lines, complexity 17.1
  - Extract: `validateMapRequest()`, `loadMapFromFile()`, `convertMapToResponse()`
  - Add corresponding unit tests (contributes to Step 8)

### Step 7: Add Tests for `handleEditorLoadMap`
- **Deliverable**: Unit tests for `pkg/server/handlers_editor.go:handleEditorLoadMap` and extracted helpers
- **Dependencies**: Step 6 (refactored functions to test)
- **Goal Impact**: Raise server test coverage toward 85% target
- **Acceptance**: New tests cover edge cases (missing file, invalid format, concurrent loads)
- **Validation**: `go test -cover ./pkg/server/... | grep handlers_editor` shows ≥80% coverage for file
- **Details**:
  - Test cases: valid map load, missing map file, invalid YAML, permission error
  - Use table-driven test pattern per project conventions

### Step 8: Raise Server Package Coverage to 85%
- **Deliverable**: Additional tests for `pkg/server/` to reach 85% coverage (currently 78%)
- **Dependencies**: Step 7 (contributes coverage)
- **Goal Impact**: Critical path reliability; server is lowest-coverage core package
- **Acceptance**: `go test -cover ./pkg/server/...` reports ≥85%
- **Validation**: `go test -cover ./pkg/server/...`
- **Details**:
  - Priority functions needing tests:
    - `handleJoinGame` (118 lines, long function)
    - `handleQuestEditorUpdate` (70 lines, complexity 14.0)
    - WebSocket session lifecycle handlers
  - ~7% coverage gap = ~35-40 new test cases estimated

### Step 9: Review and Triage BUG Annotations
- **Deliverable**: Triage 5 BUG annotations found by go-stats-generator
- **Dependencies**: None
- **Goal Impact**: Code quality hygiene
- **Acceptance**: Each BUG annotation is either fixed, converted to TODO with issue link, or documented as intentional
- **Validation**: `grep -r "BUG" pkg/ cmd/ --include="*.go" | wc -l` returns 0 or all are acknowledged
- **Details**:
  - `cmd/server/doc.go:46` - logging documentation
  - `cmd/server/doc.go:48` - server path documentation
  - `pkg/pcg/doc.go:52` - reproduction documentation
  - `pkg/server/session.go:160` - cleanup cycle messages
  - `pkg/server/server.go:760` - profiling endpoints
  - Most appear to be documentation annotations, not actual bugs

---

## Scope Assessment Calibration

| Metric | This Project | Threshold | Assessment |
|--------|-------------|-----------|------------|
| Functions above complexity 9.0 | 7 | <5 = Small | **Small-Medium** |
| Duplication ratio | 1.33% | <3% = Small | **Small** |
| Doc coverage gap | 12.5% | <10% = Small | **Small-Medium** |
| Server coverage gap | 7% | <10% = Small | **Small** |

**Overall Scope: Small** — The project is in excellent shape with only minor refinements needed.

---

## Non-Goals (Out of Scope for This Plan)
- TLS/HTTPS (handled by infrastructure/reverse proxy)
- Database integration (in-memory with file persistence by design)
- User authentication (delegated to hosting layer)
- Go toolchain upgrade to 1.25.8 (wait for stable release)
- PCG coverage increase (78.9% is acceptable; focus on server first)

---

## Summary

The GoldBox RPG Engine achieves **16 of 17 stated goals** with excellent code quality metrics:
- 87% test coverage (exceeds 60% CI requirement)
- 1.33% duplication ratio (excellent)
- 87.5% documentation coverage (strong)
- 0 circular dependencies (clean architecture)

This plan addresses the single incomplete goal (visual editor documentation) and improves maintainability through targeted complexity reduction in 7 UI functions. The estimated effort is **Small** scope with independently testable, ordered steps.

---

*Generated: 2026-03-18 by go-stats-generator analysis*
