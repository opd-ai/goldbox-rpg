# Implementation Plan: Code Quality and Security Hardening

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by classic SSI Gold Box series, providing character management, turn-based combat, and world interactions via JSON-RPC/WebSocket API.
- **Current goal**: Upgrade Go toolchain to address security vulnerabilities and reduce code duplication in server package
- **Estimated Scope**: Medium

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Character Management | ✅ Achieved | No |
| Combat & Effects System | ✅ Achieved | No |
| World Management + Spatial Indexing | ✅ Achieved | No |
| Event-Driven Architecture | ✅ Achieved | No |
| WebSocket Real-time Communication | ✅ Achieved | No |
| Health Monitoring & Prometheus Metrics | ✅ Achieved | No |
| Procedural Content Generation | ✅ Achieved | No |
| Circuit Breakers & Resilience | ✅ Achieved | No |
| Input Validation Framework | ✅ Achieved | No |
| Asset Generation Pipeline | ⚠️ Partial (48% assets) | No |
| Advanced NPC AI | ✅ Achieved | No |
| Enhanced Combat Mechanics | ✅ Achieved | No |
| Spell System (levels 0-9) | ✅ Achieved | No |
| World Editor Tools | ⚠️ Partial (CLI only) | No |
| Network Optimization | ✅ Achieved | No |
| Player Progression Persistence | ✅ Achieved | No |
| Guild and Faction Systems | ✅ Achieved | No |
| Embedded Adventures | ✅ Achieved | No |
| Go Toolchain Security | ❌ Vulnerable | **Yes** |
| Code Maintainability | ⚠️ 1.6% duplication | **Yes** |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 4 functions above threshold (complexity >9)
- **Duplication ratio**: 1.6% (1,010 duplicate lines in 40 clone pairs)
- **Doc coverage**: 88.7% overall
- **Test coverage**: 79.1% (above 60% CI threshold)
- **Package coupling**: Clean (0 circular dependencies)

### High-Duplication Hotspots
| File | Clone Size | Occurrences | Severity |
|------|------------|-------------|----------|
| `pkg/server/server.go` (lines 1027-1106) | 14-35 lines | 2-5 instances | Violation |
| `pkg/game/guild.go` (lines 375-511) | 14-19 lines | 4 instances | Violation |
| `cmd/map-editor/main.go` + `cmd/quest-builder/main.go` | 18-33 lines | 2 instances | Violation |
| `pkg/server/websocket.go` + `pkg/server/websocket_editor.go` | 21 lines | 2 instances | Violation |
| `pkg/server/handlers_guild.go` | 14-16 lines | 2 instances | Warning |

---

## Implementation Steps

### Step 1: Upgrade Go Toolchain to 1.25+
- **Deliverable**: Update `go.mod` to require Go 1.25.6+; update CI workflows to use Go 1.25.6
- **Dependencies**: None
- **Goal Impact**: Closes 6 CVEs affecting `crypto/tls`, `net/http`, `net/url`, `html/template`, and `os` packages documented in CHANGELOG.md
- **Acceptance**: `govulncheck ./...` reports 0 vulnerabilities in Go standard library
- **Validation**: 
  ```bash
  go version  # Should show 1.25.6+
  govulncheck ./...  # Should report no stdlib vulnerabilities
  go test -race ./...  # All tests pass
  ```

### Step 2: Extract Common CLI Initialization Pattern
- **Deliverable**: Create `pkg/cliutil/init.go` with shared CLI initialization code; refactor `cmd/map-editor/main.go` and `cmd/quest-builder/main.go` to use it
- **Dependencies**: None
- **Goal Impact**: Reduces 33-line duplicate block across CLI tools
- **Acceptance**: Duplication ratio in `cmd/` directory drops by ≥30 lines
- **Validation**:
  ```bash
  go-stats-generator analyze . --skip-tests --format json --sections duplication | jq '.duplication.duplicated_lines'
  # Should be <980 (down from 1010)
  ```

### Step 3: Extract Guild Permission Validation Helper
- **Deliverable**: Add `validateGuildMemberOperation()` helper in `pkg/game/guild.go` to consolidate 4 duplicate 14-line permission check blocks (lines 375-388, 420-433, 464-477, 498-511)
- **Dependencies**: None
- **Goal Impact**: Improves maintainability of guild system; reduces violation-severity duplication
- **Acceptance**: `pkg/game/guild.go` has ≤2 instances of permission validation pattern
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/game/guild.go --skip-tests --format json --sections duplication | jq '.duplication.clone_pairs'
  # Should decrease by ≥3
  ```

### Step 4: Extract WebSocket Message Handling Pattern
- **Deliverable**: Create shared `processWSMessage()` function in `pkg/server/websocket_common.go` used by both `websocket.go` and `websocket_editor.go` (currently 21-line duplicate at lines 218-238 and 164-184)
- **Dependencies**: None
- **Goal Impact**: Single point of maintenance for WebSocket message handling; reduces violation-severity duplication
- **Acceptance**: WebSocket message processing logic exists in exactly one location
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/server --skip-tests --format json --sections duplication | jq '[.duplication.clones[] | select(.instances[].file | contains("websocket"))] | length'
  # Should decrease by ≥1
  ```

### Step 5: Extract RPC Handler Session Validation Pattern
- **Deliverable**: Consolidate duplicate session validation in `pkg/server/handlers_guild.go` (lines 58-73, 92-107 and 182-195, 260-273) into reusable `validateGuildSession()` helper
- **Dependencies**: None
- **Goal Impact**: Reduces warning-severity duplication in guild handlers
- **Acceptance**: Session validation code appears once in handlers_guild.go
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/server/handlers_guild.go --skip-tests --format json --sections duplication | jq '.duplication.clone_pairs'
  # Should be ≤2 (down from 4+)
  ```

### Step 6: Extract Faction Handler Diplomacy Pattern
- **Deliverable**: Create `validateFactionDiplomacy()` helper in `pkg/server/handlers_faction.go` to consolidate 19-line duplicate at lines 143-161 and 209-227
- **Dependencies**: None
- **Goal Impact**: Single maintenance point for faction diplomacy validation
- **Acceptance**: Diplomacy validation logic exists in exactly one location
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/server/handlers_faction.go --skip-tests --format json --sections duplication | jq '.duplication.clone_pairs'
  # Should decrease by ≥1
  ```

### Step 7: Refactor Content Creator YAML Output Pattern
- **Deliverable**: Extract shared YAML output helpers from `cmd/content-creator/main.go` (duplicate blocks at lines 222-229/395-402 and 296-314/450-468) and `cmd/quest-builder/main.go` (lines 608-627) into `pkg/cliutil/yaml.go`
- **Dependencies**: Step 2 (cliutil package exists)
- **Goal Impact**: Consolidates 3-location 19-line duplicate and 2-location 8-line duplicate
- **Acceptance**: YAML output logic in CLI tools calls shared helper
- **Validation**:
  ```bash
  go-stats-generator analyze cmd --skip-tests --format json --sections duplication | jq '.duplication.duplicated_lines'
  # Should decrease by ≥50 lines
  ```

### Step 8: Extract PCG Handler Request Parsing Pattern
- **Deliverable**: Create `parsePCGRequest()` helper in `pkg/server/handlers_pcg.go` to consolidate duplicates at lines 40-53/260-273 and 450-464/486-500
- **Dependencies**: None
- **Goal Impact**: Reduces 4 duplicate blocks totaling ~58 lines
- **Acceptance**: PCG request parsing exists in ≤2 locations
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/server/handlers_pcg.go --skip-tests --format json --sections duplication | jq '.duplication.clone_pairs'
  # Should be ≤1 (down from 4)
  ```

### Step 9: Consolidate RPC Handler Request Parsing in handlers.go
- **Deliverable**: Extract shared request parsing pattern from `pkg/server/handlers.go` (lines 49-67 and 443-461) into `parseStandardRequest()` helper
- **Dependencies**: None
- **Goal Impact**: Reduces warning-severity 19-line duplicate
- **Acceptance**: Standard request parsing pattern exists once
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/server/handlers.go --skip-tests --format json --sections duplication | jq '.duplication.clone_pairs'
  # Should decrease by ≥1
  ```

### Step 10: Final Duplication Audit and Documentation
- **Deliverable**: Run full duplication analysis; update ROADMAP.md with completion status; verify all tests pass
- **Dependencies**: Steps 1-9
- **Goal Impact**: Confirms maintainability improvements; documents achieved state
- **Acceptance**: Overall duplication ratio <1.2% (down from 1.6%)
- **Validation**:
  ```bash
  go-stats-generator analyze . --skip-tests --format json --sections duplication | jq '.duplication.duplication_ratio'
  # Should be <0.012
  go test -race ./...  # All tests pass
  ```

---

## Scope Assessment Calibration

Based on this project's codebase (30,991 LOC, 189 files):

| Metric | Current Value | Threshold | Assessment |
|--------|--------------|-----------|------------|
| Functions above complexity 9.0 | 4 | <5 = Small | ✅ Small |
| Duplication ratio | 1.6% | <3% = Small | ✅ Small |
| Doc coverage gap | 11.3% | 10–25% = Medium | ⚠️ Medium |
| Clone pairs (violations) | 15 | 5–15 = Medium | ⚠️ Medium |

**Overall Scope**: Medium — primarily refactoring work with one security upgrade.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Go 1.25 breaks dependencies | Medium | High | Test in CI first; keep go 1.24 branch |
| Refactoring introduces bugs | Low | Medium | Existing 79% test coverage; run `go test -race` after each step |
| Duplication extraction changes behavior | Low | Medium | Table-driven tests verify identical behavior |

---

## Notes

- **Not addressed in this plan**: Asset generation (requires external AI tooling), GUI editor (significant new feature), doc coverage (already 88.7%)
- **Security priority**: Step 1 should be executed first due to 6 CVEs in current Go version
- **Duplication in server.go lines 1027-1106**: This is method registry initialization — structurally repetitive but not extractable without adding complexity. Acknowledged as false positive for duplication metrics.

---

*Generated: 2026-03-13*  
*Analysis Tool: go-stats-generator v1.0.0*  
*Files Analyzed: 189 Go files, 30,991 lines of code*
