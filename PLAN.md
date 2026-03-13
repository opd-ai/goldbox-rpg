# Implementation Plan: Project Maturity & Enhancement

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by the classic SSI Gold Box series, providing turn-based RPG gameplay via JSON-RPC API with WebSocket support
- **Current goal**: Transition from feature-complete to production-ready with improved maintainability and developer experience
- **Estimated Scope**: Medium

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character Management (6 attributes, classes) | ✅ Achieved | No |
| Combat & Effects System | ✅ Achieved | No |
| World Management + Spatial Indexing | ✅ Achieved | No |
| Event-Driven Architecture | ✅ Achieved | No |
| WebSocket Real-time Communication | ✅ Achieved | No |
| Health Monitoring & Prometheus Metrics | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Circuit Breakers & Resilience | ✅ Achieved | No |
| Input Validation Framework | ✅ Achieved | No |
| Asset Generation Pipeline | ✅ Achieved (252/252 placeholders) | No |
| Advanced NPC AI (A*, behavior trees) | ✅ Achieved | No |
| Opportunity Attacks, Cover/Flanking, Morale | ✅ Achieved | No |
| Spell System (levels 0-9, 60 spells) | ✅ Achieved | No |
| World Editor Tools (CLI) | ✅ Achieved | Yes — GUI enhancement |
| Network Delta Compression | ✅ Achieved | No |
| Player Progression Persistence | ✅ Achieved | No |
| Guild and Faction Systems | ✅ Achieved | No |
| Embedded Adventures (10 packs, 30+ hours) | ✅ Achieved | No |
| Go Toolchain Security Updates | ⚠️ Blocked | Yes — future planning |
| Documentation Coverage | ⚠️ Partial (86.5%) | Yes |
| Code Maintainability | ⚠️ Partial (6 high-complexity functions) | Yes |

**Summary: 18/20 goals fully achieved, 2 partial (documentation gaps, security updates blocked on Go release)**

## Metrics Summary

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 30,677 | — |
| Total Functions | 592 functions + 1,658 methods | — |
| Total Packages | 18 | — |
| Average Function Complexity | 4.0 cyclomatic | ✅ Good |
| Functions with complexity > 9 | 6 (0.27%) | ✅ Excellent |
| Duplication Ratio | 2.30% (1,395 lines) | ✅ Good (<3%) |
| Documentation Coverage | 86.5% | ✅ Good |
| Undocumented Exported Functions | 32 | ⚠️ Needs attention |
| Test Coverage | 79.4% | ✅ Above 60% CI threshold |
| Circular Dependencies | 0 | ✅ Clean |
| All Tests Passing | Yes (30/30 packages) | ✅ |

### High-Complexity Functions (Cyclomatic > 9)

| Function | File | Complexity | Recommendation |
|----------|------|------------|----------------|
| `promptRewards` | `cmd/quest-builder/main.go:326` | 10 | CLI tool — acceptable |
| `Validate` | `cmd/bootstrap-demo/main.go:96` | 10 | Demo tool — acceptable |
| `RunCellularAutomata` | `pkg/pcg/terrain/cellular_automata.go:56` | 10 | Algorithmic — acceptable |
| `registerDungeonRules` | `pkg/pcg/validator.go:416` | 10 | Rule registration — extract helpers |
| `Stop` | `test/e2e/server.go:164` | 10 | Test infrastructure — acceptable |
| `Draw` | `pkg/wasmui/adventure_screen.go:84` | 10 | UI rendering — extract sub-renderers |

### Package Coupling Analysis

| Package | Cohesion | Coupling | Assessment |
|---------|----------|----------|------------|
| `pkg/pcg` | 5.62 | 1.5 | ✅ High cohesion |
| `pkg/wasmui` | 3.44 | 1.5 | ✅ Good |
| `pkg/game` | 3.06 | 1.5 | ✅ Good |
| `pkg/server` | 2.90 | 5.5 | ⚠️ High coupling (expected for server) |
| `pkg/validation` | 3.07 | 0.5 | ✅ Low coupling |

### Research Findings

1. **Gorilla WebSocket**: Actively maintained as of 2026; v1.5.3 released June 2024 with ongoing community support — no migration needed
2. **Ebitengine**: Stable v2.7.0 in use; v2.9.x series available but optional upgrade — no breaking changes anticipated
3. **Go Stdlib Vulnerabilities**: 18 CVEs require Go 1.24.12+ or Go 1.25.8 — blocked on Go release cycle
4. **Community**: No open issues on GitHub; documentation and internal roadmap are primary drivers

---

## Implementation Steps

### Step 1: Document Undocumented Exported Functions
- **Deliverable**: Add godoc comments to 32 undocumented exported functions across `pkg/`
- **Dependencies**: None
- **Goal Impact**: Improves documentation coverage from 86.5% to >95%; better developer onboarding
- **Acceptance**: Documentation coverage ≥95% for exported functions
- **Validation**: `go-stats-generator analyze . --skip-tests --format json --sections documentation | jq '.documentation.coverage.functions'` shows ≥95

**Files to document** (prioritized by package importance):
1. `pkg/game/` — Core mechanics API surface
2. `pkg/server/` — RPC handlers and session management
3. `pkg/pcg/` — Content generation interfaces

---

### Step 2: Extract Helper Functions from `registerDungeonRules`
- **Deliverable**: Refactor `pkg/pcg/validator.go:registerDungeonRules` from 78 lines/complexity 10 into 3-4 focused helpers
- **Dependencies**: None
- **Goal Impact**: Improves code maintainability; reduces cognitive load for PCG contributors
- **Acceptance**: `registerDungeonRules` complexity ≤7; no test regressions
- **Validation**: `go-stats-generator analyze . --skip-tests --format json --sections functions | jq '[.functions[] | select(.name == "registerDungeonRules")] | .[0].complexity.cyclomatic'` shows ≤7

**Suggested extraction pattern**:
```go
func (v *ContentValidator) registerDungeonRules() {
    v.registerDungeonSizeRules()
    v.registerDungeonConnectivityRules()
    v.registerDungeonEntityRules()
}
```

---

### Step 3: Extract Sub-Renderers from AdventureScreen.Draw
- **Deliverable**: Refactor `pkg/wasmui/adventure_screen.go:Draw` (64 lines/complexity 10) into composable render methods
- **Dependencies**: None
- **Goal Impact**: Improves UI code maintainability; enables easier theming and customization
- **Acceptance**: `Draw` complexity ≤6; no visual regressions
- **Validation**: `go-stats-generator analyze . --skip-tests --format json --sections functions | jq '[.functions[] | select(.file | contains("adventure_screen.go")) | select(.name == "Draw")] | .[0].complexity.cyclomatic'` shows ≤6

**Suggested extraction**:
```go
func (a *AdventureScreen) Draw(screen *ebiten.Image, g *Game) {
    a.drawBackground(screen)
    a.drawAdventureList(screen, g)
    a.drawSelectedDetails(screen, g)
    a.drawControls(screen)
}
```

---

### Step 4: Add Go Toolchain Upgrade Tracking
- **Deliverable**: Create `.github/ISSUE_TEMPLATE/security-upgrade.md` template; add `scripts/check-go-vuln.sh` automation
- **Dependencies**: None
- **Goal Impact**: Prepares project for Go 1.24.12+/1.25.8 security update when available
- **Acceptance**: Script runs in CI; outputs clear upgrade recommendation
- **Validation**: `./scripts/check-go-vuln.sh` returns non-zero if vulnerable Go version detected

**Script outline**:
```bash
#!/bin/bash
# Check if current Go version has known vulnerabilities
govulncheck ./...
if [ $? -ne 0 ]; then
    echo "⚠️ Vulnerabilities detected - upgrade Go when fix available"
    exit 1
fi
```

---

### Step 5: Browser-Based Map Editor MVP (Enhancement)
- **Deliverable**: Extend `pkg/wasmui/editor.go` with clickable tile placement and export functionality
- **Dependencies**: Steps 1-3 (clean codebase)
- **Goal Impact**: Addresses "World editor tools (CLI only)" limitation noted in README roadmap
- **Acceptance**: User can create a 20x20 map in browser, place terrain tiles, and export as YAML
- **Validation**: `go test ./pkg/wasmui/... -v -run TestEditorExport` passes

**Implementation approach**:
1. Add `EditorScreen` state to WASM UI game loop
2. Implement tile palette sidebar with terrain types
3. Add click-to-place tile functionality
4. Add YAML export via browser download API

---

### Step 6: Reduce Code Duplication in Guild/Faction Handlers
- **Deliverable**: Extract common RPC response patterns from `pkg/server/handlers_guild.go` (14-line clones × 7)
- **Dependencies**: None
- **Goal Impact**: Reduces duplication ratio from 2.3% toward 2.0%; DRY principle
- **Acceptance**: Duplication ratio <2.2%; handlers use shared helper
- **Validation**: `go-stats-generator analyze . --skip-tests --format json --sections duplication | jq '.duplication.duplication_ratio'` shows <0.022

**Pattern to extract**:
```go
func (s *RPCServer) sendGuildResponse(sessionID string, result interface{}, err error) map[string]interface{} {
    if err != nil {
        return map[string]interface{}{"success": false, "error": err.Error()}
    }
    return map[string]interface{}{"success": true, "result": result, "session_id": sessionID}
}
```

---

### Step 7: Add Integration Test for Adventure Loading
- **Deliverable**: Add `test/e2e/adventure_integration_test.go` verifying all 10 adventures load and pass smoke test
- **Dependencies**: None
- **Goal Impact**: Ensures bundled content integrity; prevents adventure regression
- **Acceptance**: All 10 adventures load without error; basic quest chain validates
- **Validation**: `go test ./test/e2e/... -v -run TestAdventureIntegration` passes

**Test coverage**:
- Adventure discovery via `adventure.list` RPC
- Adventure loading via `adventure.load` RPC for each of 10 adventures
- Schema validation for loaded adventure data
- Basic playthrough smoke test (enter adventure, complete first objective)

---

## Scope Assessment

Using project-calibrated thresholds:

| Metric | Current | Target | Items to Address |
|--------|---------|--------|------------------|
| Functions > complexity 9 | 6 | <5 | 2 functions (Steps 2, 3) |
| Duplication ratio | 2.3% | <2.2% | ~100 lines (Step 6) |
| Doc coverage gap | 13.5% | <5% | 32 functions (Step 1) |

**Scope: Medium** (5-15 items above threshold across categories)

---

## Dependency Graph

```
Step 1 (Documentation) ──────────────────────────────┐
                                                     │
Step 2 (Refactor registerDungeonRules) ──────────────┼──> Step 5 (Browser Editor MVP)
                                                     │
Step 3 (Refactor AdventureScreen.Draw) ──────────────┘

Step 4 (Go Upgrade Tracking) ── Independent

Step 6 (DRY Guild Handlers) ── Independent

Step 7 (Adventure Integration Tests) ── Independent
```

Steps 1-3 are prerequisites for Step 5 (browser editor) to ensure clean codebase before new feature work.
Steps 4, 6, 7 can be executed independently in parallel.

---

## Future Considerations (Out of Scope)

1. **GUI Quest Builder**: Lower priority than map editor; CLI tool is functional
2. **Full AI Asset Generation**: Requires external tool setup; placeholders are sufficient for development
3. **Ebitengine v2.9 Upgrade**: Optional; v2.7.0 is stable and maintained
4. **Additional Adventures**: 10 adventures with 30+ hours of content is substantial; community contributions welcome

---

*Generated: 2026-03-13*  
*Analysis Tool: go-stats-generator v1.0.0*  
*Files Analyzed: 184 Go files, 30,677 lines of code*
