# Implementation Plan: Codebase Quality Hardening

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by SSI Gold Box games, providing turn-based combat, character management, procedural content generation, and real-time communication through JSON-RPC/WebSocket APIs for web-based RPG experiences.
- **Current goal**: Strengthen codebase quality by reducing complexity hotspots, improving test coverage in critical packages, and maintaining clean architecture
- **Estimated Scope**: Medium (5–15 items above threshold for complexity >9.0 in goal-critical paths)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character System (6 attributes, classes, creation methods) | ✅ Achieved | No |
| Comprehensive Effect System (DoT, HoT, conditions, stacking) | ✅ Achieved | No |
| Spatial Indexing (R-tree-like structure) | ✅ Achieved | No |
| Procedural Content Generation (terrain, items, quests, NPCs) | ✅ Achieved | Indirectly (coverage) |
| Event-Driven Architecture | ✅ Achieved | No |
| WebSocket Real-time Communication | ✅ Achieved | No |
| Health Monitoring (/health, /ready, /live, /metrics) | ✅ Achieved | No |
| System Resilience (circuit breakers, retry, validation) | ✅ Achieved | No |
| Asset Generation Pipeline (521 assets) | ✅ Achieved | No |
| Embedded Adventures (10 packs, 100 maps) | ✅ Achieved | No |
| Test Coverage ≥60% | ✅ Achieved (87% avg) | Yes (improve server/PCG) |
| UI Maintainability | ⚠️ 3 functions >15 complexity | Yes |
| Code Organization | ⚠️ handlers.go 1,374 lines | Yes |

**Summary**: All 19 feature goals achieved. This plan targets maintenance improvements that further strengthen an already production-quality codebase.

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 186 functions above threshold 9.0 (3 above 15.0)
- **Duplication ratio**: 0% (932 duplicated lines total, 1.34% historical)
- **Doc coverage**: Varies by package (87% test coverage overall)
- **Package coupling**: Clean architecture with 0 circular dependencies
- **Total LOC**: 35,276 across 218 files and 19 packages

### Top Complexity Functions (>15.0)
| Function | Package | File | Lines | Complexity |
|----------|---------|------|-------|------------|
| `drawCombatGrid` | wasmui | combat_screen.go | 73 | 19.4 |
| `updateCharCreationName` | wasmui | character_creation.go | 59 | 17.1 |
| `Draw` | wasmui | adventure_screen.go | 99 | 15.8 |

### Server Package Coverage Target
Current: 78.3% | Target: 85% | Gap: 6.7%

### PCG Package Coverage Target
Current: 78.9% | Target: 85% | Gap: 6.1%

---

## Implementation Steps

### Step 1: Refactor `drawCombatGrid` Complexity
- **Deliverable**: Reduce complexity from 19.4 to <15.0 by extracting helper functions in `pkg/wasmui/combat_screen.go`
- **Dependencies**: None
- **Goal Impact**: UI maintainability (highest complexity function in codebase)
- **Specific Changes**:
  1. Extract `drawGridCell(x, y int, cell *GridCell)` for individual cell rendering
  2. Extract `highlightSelectedCell(x, y int, selected bool)` for selection state
  3. Extract `drawEntityOnCell(entity Entity, x, y int)` for entity overlay rendering
- **Acceptance**: Function complexity <15.0 as measured by go-stats-generator
- **Validation**: `go-stats-generator analyze ./pkg/wasmui/combat_screen.go --format json 2>/dev/null | jq '.functions[] | select(.name=="drawCombatGrid") | .complexity.overall'` returns value <15.0

### Step 2: Refactor `updateCharCreationName` Complexity
- **Deliverable**: Reduce complexity from 17.1 to <15.0 by extracting keyboard handling in `pkg/wasmui/character_creation.go`
- **Dependencies**: None
- **Goal Impact**: UI maintainability (second highest complexity function)
- **Specific Changes**:
  1. Extract `handleNameKeyInput(key ebiten.Key) rune` for key-to-character mapping
  2. Use table-driven key mapping instead of switch statements
  3. Extract `validateNameInput(name string) bool` for input validation
- **Acceptance**: Function complexity <15.0 as measured by go-stats-generator
- **Validation**: `go-stats-generator analyze ./pkg/wasmui/character_creation.go --format json 2>/dev/null | jq '.functions[] | select(.name=="updateCharCreationName") | .complexity.overall'` returns value <15.0

### Step 3: Refactor `Draw` in adventure_screen.go
- **Deliverable**: Reduce complexity from 15.8 to <15.0 by extracting state-specific drawing in `pkg/wasmui/adventure_screen.go`
- **Dependencies**: None
- **Goal Impact**: UI maintainability (third highest complexity function)
- **Specific Changes**:
  1. Extract `drawAdventureHeader(screen *ebiten.Image)` for header panel
  2. Extract `drawQuestPanel(screen *ebiten.Image)` for quest state display
  3. Extract `drawInventoryOverlay(screen *ebiten.Image)` for inventory rendering
- **Acceptance**: Function complexity <15.0 as measured by go-stats-generator
- **Validation**: `go-stats-generator analyze ./pkg/wasmui/adventure_screen.go --format json 2>/dev/null | jq '.functions[] | select(.name=="Draw") | .complexity.overall'` returns value <15.0

### Step 4: Add Server Handler Edge Case Tests
- **Deliverable**: Increase `pkg/server/` test coverage from 78.3% to 85%
- **Dependencies**: Steps 1-3 (complete UI refactoring first to establish patterns)
- **Goal Impact**: Server reliability (critical path for all client requests)
- **Specific Test Files**:
  1. `pkg/server/handlers_test.go` — add concurrent session join scenarios
  2. `pkg/server/handlers_quest_editor_test.go` — add concurrent edit detection tests
  3. `pkg/server/session_test.go` — add timeout edge case tests
- **Test Cases to Add**:
  - `TestHandleJoinGame_ConcurrentSessions` — verify race-free session attachment
  - `TestHandleQuestEditorUpdate_ConcurrentEdits` — verify conflict detection
  - `TestSessionTimeout_EdgeCases` — verify cleanup at exact timeout boundary
  - `TestWebSocketReconnection_SessionReuse` — verify session state preservation
- **Acceptance**: `go test -cover ./pkg/server/...` shows ≥85%
- **Validation**: `go test -cover ./pkg/server/... 2>&1 | grep 'coverage:' | awk '{print $2}'` returns value ≥85.0%

### Step 5: Add PCG Determinism Verification Tests
- **Deliverable**: Increase `pkg/pcg/` test coverage from 78.9% to 85%
- **Dependencies**: None (can run parallel to Step 4)
- **Goal Impact**: PCG reliability (core differentiator for procedural content)
- **Specific Test Files**:
  1. `pkg/pcg/seed_test.go` — add deterministic seeding verification
  2. `pkg/pcg/validator_test.go` — add malformed content validation tests
  3. `pkg/pcg/terrain/generator_test.go` — add boundary condition tests
- **Test Cases to Add**:
  - `TestSeedDeterminism_IdenticalOutput` — same seed produces identical content
  - `TestSeedDerivation_EdgeCases` — empty context, long context, special characters
  - `TestValidateContent_MalformedInput` — graceful handling of invalid content
  - `TestTerrainGeneration_BoundaryConditions` — 1x1 maps, maximum sizes
- **Acceptance**: `go test -cover ./pkg/pcg/...` shows ≥85%
- **Validation**: `go test -cover ./pkg/pcg/... 2>&1 | grep 'coverage:' | awk '{print $2}'` returns value ≥85.0%

### Step 6: Split handlers.go by Domain
- **Deliverable**: Reduce `pkg/server/handlers.go` from 1,374 lines to <800 lines
- **Dependencies**: Step 4 (test coverage first to validate refactoring)
- **Goal Impact**: Code organization (clean package separation per README promise)
- **New Files to Create**:
  1. `pkg/server/handlers_combat.go` — attack, spell casting, turn management handlers
  2. `pkg/server/handlers_quest.go` — quest start, complete, update handlers
  3. `pkg/server/handlers_spatial.go` — object queries, range search handlers
- **Handlers to Move**:
  - `handleAttack`, `handleSpellCast`, `handleEndTurn` → handlers_combat.go
  - `handleStartQuest`, `handleCompleteQuest`, `handleUpdateQuest` → handlers_quest.go
  - `handleGetObjectsInRange`, `handleQuerySpatialIndex` → handlers_spatial.go
- **Acceptance**: Main handlers.go under 800 lines, all tests pass
- **Validation**: `wc -l pkg/server/handlers.go` shows <800 && `go test ./pkg/server/...` passes

### Step 7: Reduce Server Method Registration Duplication
- **Deliverable**: Extract common registration pattern in `pkg/server/server.go`
- **Dependencies**: Step 6 (handler organization complete)
- **Goal Impact**: Maintainability (reduce risk of inconsistent method registration)
- **Specific Changes**:
  1. Create `registerMethod(method RPCMethod, handler MethodHandler)` helper
  2. Apply pattern to consolidate repetitive `s.methodRegistry[Method...] = s.handle...` lines
  3. Consider adding middleware hooks for cross-cutting concerns (logging, metrics)
- **Acceptance**: No clone pairs >20 lines in server.go registration section
- **Validation**: `go-stats-generator analyze ./pkg/server/server.go --format json 2>/dev/null | jq '.duplication.clone_pairs | length'` returns 0 or shows no clones >20 lines

---

## Validation Summary

After completing all steps, run full validation:

```bash
# 1. Verify no high-complexity functions
go-stats-generator analyze . --skip-tests --format json 2>/dev/null | \
  jq '[.functions[] | select(.complexity.overall > 15.0)] | length'
# Expected: 0

# 2. Verify server coverage
go test -cover ./pkg/server/... 2>&1 | grep 'coverage:' | awk '{print $2}'
# Expected: ≥85.0%

# 3. Verify PCG coverage
go test -cover ./pkg/pcg/... 2>&1 | grep 'coverage:' | awk '{print $2}'
# Expected: ≥85.0%

# 4. Verify handlers.go size
wc -l pkg/server/handlers.go
# Expected: <800

# 5. Verify all tests pass
go test ./...
# Expected: PASS

# 6. Verify build succeeds
go build ./...
# Expected: No errors
```

---

## Risk Assessment

| Step | Risk | Mitigation |
|------|------|------------|
| Steps 1-3 (UI refactoring) | May affect rendering behavior | Table-driven tests for visual output; manual verification |
| Step 4 (server tests) | Tests may reveal existing bugs | Fix bugs as discovered; document in bugfix report |
| Step 5 (PCG tests) | Seed changes may affect existing content | Maintain backward compatibility; version seeds if needed |
| Step 6 (handler split) | Import cycles possible | Use careful package boundaries; no new dependencies |
| Step 7 (registration) | May affect method discovery | Verify all methods registered; add registration tests |

---

## Non-Goals (Out of Scope)

- **TLS/HTTPS**: Transport security handled by infrastructure (reverse proxy)
- **Database integration**: Game state is in-memory with file persistence by design
- **User authentication**: Not part of engine scope (delegated to hosting layer)
- **New features**: This plan focuses on quality improvements, not new functionality

---

## Dependencies & Prerequisites

- Go 1.25.6+ with toolchain 1.25.8
- `go-stats-generator` installed: `go install github.com/opd-ai/go-stats-generator@latest`
- Existing tests must pass before starting: `go test ./...`
- CI pipeline must be green before merging any changes

---

## Timeline Estimate

| Step | Estimated Effort |
|------|-----------------|
| Step 1: drawCombatGrid | 2-3 hours |
| Step 2: updateCharCreationName | 2-3 hours |
| Step 3: Draw adventure_screen | 2-3 hours |
| Step 4: Server tests | 4-6 hours |
| Step 5: PCG tests | 4-6 hours |
| Step 6: Handler split | 3-4 hours |
| Step 7: Registration dedup | 1-2 hours |
| **Total** | **18-27 hours** |

---

## Research Findings

### Online Research Summary (Phase 1)

**GitHub Issues & Community**:
- Low issue volume indicates stable codebase or small user base
- Recent commits focus on browser playtest fixes, CI improvements, dependency updates
- No critical bugs reported; maintenance is active

**Dependency Status**:
- `coder/websocket v1.8.14`: Actively maintained (Coder took over from nhooyr); known DoS vulnerability (double channel close) patched in recent versions — verify current version is patched
- `gorilla/websocket`: Retained only for E2E test client (production uses coder/websocket)
- `ebiten v2.9.9`: Latest version with fullscreen/vector rendering fixes
- Go stdlib vulnerabilities: 18 identified by `govulncheck` requiring Go 1.24.12+ or 1.25.8 — documented in CHANGELOG as known issue

**Best Practices Alignment**:
- Project follows Go conventions (gofumpt, golangci-lint)
- CI gates comprehensive (race detection, coverage, vulnerability scanning)
- Documentation quality high (README, CHANGELOG, GAPS, ROADMAP current)

---

*Generated: 2026-03-18 | Tool: go-stats-generator v1.0.0*
