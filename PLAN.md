# Implementation Plan: Documentation & Tooling Refinement

**Generated:** 2026-03-13  
**Analysis Tool:** go-stats-generator v1.0.0  
**Files Analyzed:** 184 Go files, 30,553 lines of code

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by the classic SSI Gold Box series, providing character management, combat systems, and world interactions through a JSON-RPC API with WebSocket support.
- **Current goal**: Improve documentation accuracy, complete partial implementations, and enhance content creation tooling.
- **Estimated Scope**: Small (all major features implemented; refinement and polish work remains)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character Management (6 attributes, classes, creation) | ✅ Achieved | No |
| Combat & Effects System | ✅ Achieved | No |
| World Management + Spatial Indexing | ✅ Achieved | No |
| Event-Driven Architecture | ✅ Achieved | No |
| WebSocket Real-time Communication | ✅ Achieved | No |
| Health Monitoring & Prometheus Metrics | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Circuit Breakers & Resilience | ✅ Achieved | No |
| Input Validation Framework | ✅ Achieved | No |
| Asset Generation Pipeline | ⚠️ Partial (placeholders only) | Yes |
| Advanced NPC AI (A*, behavior trees) | ✅ Achieved | No |
| Opportunity Attacks, Cover/Flanking, Morale | ✅ Achieved | No |
| Spell System (levels 0-9, 60 spells) | ✅ Achieved | No |
| World Editor Tools | ⚠️ Partial (CLI only, no GUI) | Yes |
| Network Delta Compression | ✅ Achieved | No |
| Player Progression Persistence | ✅ Achieved | No |
| Guild & Faction Systems | ✅ Achieved | No |
| Embedded Adventures (10 complete) | ✅ Achieved | No |
| Vault Secrets Provider | ⚠️ Stub only | Yes |
| Documentation Accuracy | ⚠️ Inconsistencies exist | Yes |

**Overall: 16/20 goals fully achieved, 4 partial**

## Metrics Summary

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 30,553 | — |
| Total Functions + Methods | 592 + 1,641 | — |
| High Complexity Functions (>9) | 9 | ✅ Low risk |
| Maximum Complexity | 10 | ✅ Within acceptable limits |
| Documentation Coverage | 86.4% | ✅ Good |
| Code Duplication | 2.3% (1,395 lines) | ✅ Low |
| Test Coverage | 81.37% | ✅ Above 60% CI threshold |
| E2E Tests | All passing | ✅ Healthy |
| Circular Dependencies | 0 | ✅ Clean architecture |

### Complexity Hotspots (complexity = 10)
| Function | File | Lines |
|----------|------|-------|
| `promptRewards` | `cmd/content-creator/main.go` | 54 |
| `Validate` | `cmd/content-creator/main.go` | 44 |
| `RunCellularAutomata` | `pkg/pcg/terrain/cellular_automata.go` | 40 |
| `registerDungeonRules` | `pkg/pcg/validator.go` | 78 |
| `Draw` | `pkg/wasmui/adventure_screen.go` | 64 |
| `Stop` | `pkg/server/server.go` | 42 |
| `handleKeyboardInput` | `pkg/wasmui/editor.go` | 35 |
| `Connect` | `pkg/wasmui/rpc_client_wasm.go` | 72 |
| `handleMessage` | `pkg/wasmui/rpc_client_wasm.go` | 45 |

### Design Patterns Detected
- **Singleton** (3 instances): `pkg/resilience/degradation.go`, `pkg/resilience/manager.go`, `pkg/wasmui/rpc_client_wasm.go`
- **Factory Method** (3 instances): `pkg/game/ai_behaviors.go`, `pkg/server/websocket.go`
- **Observer Pattern** (4 instances): `pkg/game/events.go`, `pkg/pcg/validator.go`, `pkg/server/health.go`
- **Strategy Pattern**: Multiple instances in AI and game logic

---

## Implementation Steps

### Step 1: Update README Badges and Documentation Accuracy

- **Deliverable**: Update `README.md` to accurately reflect current implementation status
- **Dependencies**: None
- **Goal Impact**: Documentation Accuracy
- **Changes Required**:
  1. Update coverage badge from "60%" to "81%" (actual measured coverage)
  2. Update asset badge from "6/521 (1%)" to "37 placeholders / 521 defined"
  3. Verify ROADMAP items match actual implementation (guild system, spell system already complete)
  4. Add note that adventure system is fully functional with 10 adventures
- **Acceptance**: README claims match `go-stats-generator` metrics and codebase reality
- **Validation**: 
  ```bash
  go test ./... -cover 2>&1 | grep -E "^ok.*coverage:" | awk '{sum+=$5; count++} END {print "Coverage: " sum/count "%"}'
  # Expected: > 80%
  ```

---

### Step 2: Resolve Vault Provider Stub

- **Deliverable**: Either implement HashiCorp Vault integration or remove the stub and document limitations
- **Dependencies**: Step 1 (documentation must be updated to reflect decision)
- **Goal Impact**: Vault Secrets Provider (removes ⚠️ partial status)
- **Options**:
  - **Option A (Implement)**: Complete `pkg/secrets/vault_provider.go` using HashiCorp Vault Go client
  - **Option B (Remove)**: Delete stub, update `pkg/secrets/doc.go` to document env/file-only support
- **Acceptance**: 
  - If implemented: Vault provider passes integration test connecting to test Vault instance
  - If removed: No TODO comments remain in `pkg/secrets/vault_provider.go`
- **Validation**:
  ```bash
  grep -r "TODO.*Future implementation" pkg/secrets/ | wc -l
  # Expected: 0 (no stubs remaining)
  ```

---

### Step 3: Enhance Asset Pipeline Documentation

- **Deliverable**: Improve asset generation documentation to help users get real assets
- **Dependencies**: None
- **Goal Impact**: Asset Generation Pipeline
- **Changes Required**:
  1. Update `ASSET_INTEGRATION.md` with step-by-step Stable Diffusion local setup
  2. Add ComfyUI workflow file for batch asset generation
  3. Document alternative: purchasing/licensing CC0 pixel art packs
  4. Add `make assets-download` target to fetch pre-generated asset pack from releases
- **Acceptance**: User can follow documentation to generate or obtain real game assets
- **Validation**:
  ```bash
  ls web/static/assets/sprites/*.png | wc -l
  # Baseline: 37 placeholders
  # After real assets: 521+ files with size > 1KB
  ```

---

### Step 4: Add GUI Map Editor Foundation

- **Deliverable**: Browser-based map editor using existing WASM infrastructure
- **Dependencies**: None (extends existing `pkg/wasmui/editor.go`)
- **Goal Impact**: World Editor Tools (upgrades from CLI-only to GUI+CLI)
- **Implementation Approach**:
  1. Create `pkg/wasmui/map_editor.go` with Ebitengine-based tile palette and canvas
  2. Add WebSocket sync for real-time collaboration preview
  3. Integrate with existing `pkg/game/map.go` for save/load
  4. Add `cmd/wasm-editor/` entry point (already exists, needs enhancement)
- **Acceptance**: User can create, edit, and save a map visually in browser
- **Validation**:
  ```bash
  go build ./cmd/wasm-editor/...
  # Expected: builds successfully
  go test ./pkg/wasmui/... -run TestMapEditor
  # Expected: editor tests pass
  ```

---

### Step 5: Reduce Maximum Function Complexity

- **Deliverable**: Refactor functions with complexity = 10 to improve maintainability
- **Dependencies**: None
- **Goal Impact**: Code quality and maintainability
- **Priority Targets** (highest impact):
  1. `registerDungeonRules` (78 lines) → Extract rule registration helpers
  2. `Connect` in `rpc_client_wasm.go` (72 lines) → Extract connection setup phases
  3. `Draw` in `adventure_screen.go` (64 lines) → Extract rendering sub-functions
- **Acceptance**: No functions exceed complexity 9
- **Validation**:
  ```bash
  go-stats-generator analyze . --skip-tests --format json --sections functions 2>/dev/null | \
    jq '[.functions[] | select(.complexity.cyclomatic > 9)] | length'
  # Expected: 0
  ```

---

### Step 6: Add Handler-Validator Parity Test

- **Deliverable**: CI test ensuring every registered RPC method has a validator
- **Dependencies**: None
- **Goal Impact**: Prevents silent feature breakage (adventure methods issue prevention)
- **Implementation**:
  1. Create `pkg/server/handler_coverage_test.go`
  2. Parse `pkg/server/constants.go` for all `RPCMethod` constants
  3. Verify each method exists in `pkg/validation/validation.go` validators map
  4. Fail CI if any method lacks validation
- **Acceptance**: Test passes and catches any future validator omissions
- **Validation**:
  ```bash
  go test ./pkg/server/... -run TestHandlerValidatorParity -v
  # Expected: PASS
  ```

---

### Step 7: Enhance Duplication Hotspots (Optional Refinement)

- **Deliverable**: Extract common patterns from identified duplication areas
- **Dependencies**: Steps 1-6 completed
- **Goal Impact**: Code maintainability (duplication already low at 2.3%)
- **Priority Targets**:
  1. `pkg/server/handlers_guild.go` (14 lines × 7 clones) → Extract RPC response helper
  2. `pkg/game/guild.go` (14 lines × 4 clones) → Extract member validation helper
  3. `pkg/game/faction_relations.go` (11 lines × 10 clones) → Extract reputation modifier
- **Acceptance**: Duplication ratio < 2.0%
- **Validation**:
  ```bash
  go-stats-generator analyze . --skip-tests --format json --sections duplication 2>/dev/null | \
    jq '.duplication.duplication_ratio'
  # Expected: < 0.02
  ```

---

## Dependency Graph

```
Step 1 (README) ──────────────────────────────────────────────────┐
                                                                   │
Step 2 (Vault) ─── depends on ─── Step 1                          │
                                                                   │
Step 3 (Assets) ──────────────────────────────────────────────────│── All independent
                                                                   │
Step 4 (GUI Editor) ──────────────────────────────────────────────│
                                                                   │
Step 5 (Complexity) ──────────────────────────────────────────────│
                                                                   │
Step 6 (Parity Test) ─────────────────────────────────────────────┘
                                                                   
Step 7 (Duplication) ─── depends on ─── Steps 1-6 complete
```

---

## Scope Assessment Calibration

Based on this project's baseline:

| Metric | This Project | Threshold Assessment |
|--------|--------------|---------------------|
| Functions above complexity 9 | 9 | Small (< 15 threshold for Medium) |
| Duplication ratio | 2.3% | Small (< 3% threshold) |
| Doc coverage gap | 13.6% | Medium (10-25% range) |
| Test coverage | 81.37% | ✅ Exceeds 60% CI requirement |

**Overall Scope: Small** — The codebase is well-maintained with no critical technical debt. Remaining work is refinement and polish rather than fundamental implementation.

---

## External Dependencies & Risks

### Go Toolchain Vulnerabilities
Per `CHANGELOG.md`, 18 Go stdlib vulnerabilities exist requiring Go 1.24.12+ or Go 1.25.8:
- Affects: `crypto/tls`, `crypto/x509`, `net/http`, `net/url`, `html/template`, `os`
- **Recommendation**: Plan toolchain upgrade when Go 1.24.12+ releases
- **Risk**: Low for typical use cases; upgrade recommended for production

### Dependency Status
| Dependency | Version | Status |
|------------|---------|--------|
| gorilla/websocket | v1.5.3 | ✅ No known CVEs |
| hajimehoshi/ebiten | v2.7.0 | ⚠️ Deprecated APIs used |
| prometheus/client_golang | v1.23.2 | ✅ Current |

### Ebitengine Migration Notes
When upgrading to Ebitengine v2.8+:
- Replace `(*ebiten.Image).Dispose` → `(*ebiten.Image).Deallocate`
- Replace `ebiten.DeviceScaleFactor` → `ebiten.Monitor().DeviceScaleFactor()`
- Update mobile build pipeline if applicable

---

## Success Criteria

This plan is complete when:
1. ✅ README badges accurately reflect metrics (coverage, assets)
2. ✅ Vault provider is either implemented or cleanly removed
3. ✅ Asset pipeline has actionable documentation for real asset generation
4. ✅ Browser-based map editor exists and functions
5. ✅ No functions exceed complexity 9
6. ✅ Handler-validator parity test exists in CI
7. ✅ Duplication ratio < 2.0% (optional)

---

## Appendix: Raw Metrics

```json
{
  "overview": {
    "total_lines_of_code": 30553,
    "total_functions": 592,
    "total_methods": 1641,
    "total_structs": 402,
    "total_interfaces": 20,
    "total_packages": 18,
    "total_files": 184
  },
  "documentation": {
    "coverage": {
      "overall": 86.44
    }
  },
  "duplication": {
    "clone_pairs": 58,
    "duplicated_lines": 1395,
    "duplication_ratio": 0.023
  }
}
```
