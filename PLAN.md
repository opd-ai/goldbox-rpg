# Implementation Plan: Phase 1 - Critical Path Optimization

## Project Context
- **What it does**: A modern Go-based RPG engine providing turn-based RPG games with combat systems, character management, and world interactions through a JSON-RPC API with WebSocket support.
- **Current milestone**: Phase 1 - Critical Path Optimization (from ROADMAP.md Remediation Roadmap)
- **Estimated Scope**: **Large** (16 functions above complexity threshold 9, 189 functions >30 lines)

## Metrics Summary
- Complexity hotspots: **16 functions** above threshold (complexity >9)
- Duplication ratio: **1.19%** (well below 5% threshold ✅)
- Doc coverage: **83.8%** (above 80% threshold ✅)
- Package coupling: Clean layered architecture with zero circular dependencies ✅

### Current Metrics Baseline (go-stats-generator 2026-03-11)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Functions complexity >9 | 16 | 0 | ⚠️ FAIL |
| Functions >30 lines | 189 | 0 | ⚠️ FAIL |
| Documentation coverage | 83.8% | ≥80% | ✅ PASS |
| Duplication ratio | 1.19% | <5% | ✅ PASS |
| Circular dependencies | 0 | 0 | ✅ PASS |

### Priority Functions Requiring Remediation
| Function | File | Complexity | Lines | Priority |
|----------|------|------------|-------|----------|
| `extractMethods` | cmd/openapi-gen/main.go:79 | 16 | 76 | P3 (tooling) |
| `ModifyReputation` | pkg/pcg/reputation.go:197 | 14 | 89 | **P1** |
| `addVegetation` | pkg/pcg/terrain/generator.go:674 | 14 | 57 | P2 |
| `refreshGameState` | pkg/wasmui/game.go:135 | 14 | 51 | **P1** |
| `GenerateLevel` | pkg/pcg/levels/generator.go:147 | 13 | 74 | P2 |
| `LoadFromFile` | pkg/pcg/items/templates.go:112 | 12 | 48 | P2 |
| `NewRPCServer` | pkg/server/server.go:439 | 12 | 63 | **P1** |
| `handleMethod` | pkg/server/server.go:897 | 6 | 140 | **P1** |
| `handleAttack` | pkg/server/handlers.go:284 | 8 | 80 | **P1** |
| `NewMetrics` | pkg/server/metrics.go:49 | 1 | 148 | P2 |

---

## Implementation Steps

### Step 1: Refactor Server Request Handler
- **Deliverable**: Split `handleMethod` (140 lines) into focused helper functions
- **Dependencies**: None (entry point)
- **Acceptance**: `handleMethod` ≤30 lines code, all existing tests pass
- **Validation**: 
```bash
go-stats-generator analyze . --skip-tests --format json --output /tmp/metrics.json --sections functions && \
jq '.functions[] | select(.name == "handleMethod" and .file | contains("server.go")) | {name, lines: .lines.code, complexity: .complexity.cyclomatic}' /tmp/metrics.json
```

**Refactoring Plan**:
1. Extract `routeToHandler(method string) (HandlerFunc, error)` - method routing lookup
2. Extract `validateSession(sessionID string) (*Session, error)` - session validation
3. Extract `sendJSONResponse(w http.ResponseWriter, data interface{})` - response serialization
4. Reduce `handleMethod` to orchestration logic only

---

### Step 2: Refactor RPC Attack Handler
- **Deliverable**: Split `handleAttack` (80 lines, complexity 8) into testable combat functions
- **Dependencies**: Step 1 (handler pattern established)
- **Acceptance**: `handleAttack` ≤30 lines, each extracted function has unit tests
- **Validation**:
```bash
go-stats-generator analyze . --skip-tests --format json --output /tmp/metrics.json --sections functions && \
jq '.functions[] | select(.name == "handleAttack") | {name, lines: .lines.code, complexity: .complexity.cyclomatic}' /tmp/metrics.json
```

**Refactoring Plan**:
1. Extract `validateCombatAction(attacker, target *Character) error`
2. Extract `calculateDamageResult(attacker, target *Character, weapon *Item) DamageResult`
3. Extract `applyAttackEffects(target *Character, effects []Effect)`
4. Extract `emitCombatEvents(attackResult AttackResult)`

---

### Step 3: Refactor Server Initialization
- **Deliverable**: Split `NewRPCServer` (63 lines, complexity 12) into initialization methods
- **Dependencies**: None (can be done in parallel with Step 2)
- **Acceptance**: `NewRPCServer` ≤25 lines, complexity ≤5
- **Validation**:
```bash
go-stats-generator analyze . --skip-tests --format json --output /tmp/metrics.json --sections functions && \
jq '.functions[] | select(.name == "NewRPCServer") | {name, lines: .lines.code, complexity: .complexity.cyclomatic}' /tmp/metrics.json
```

**Refactoring Plan**:
1. Extract `(s *RPCServer) registerRoutes()`
2. Extract `(s *RPCServer) setupMiddleware()`
3. Extract `(s *RPCServer) registerHealthChecks()`
4. Extract `(s *RPCServer) registerWebSocketHandler()`

---

### Step 4: Refactor WASM UI State Refresh
- **Deliverable**: Split `refreshGameState` (51 lines, complexity 14) into UI update methods
- **Dependencies**: None (isolated package)
- **Acceptance**: `refreshGameState` ≤20 lines, complexity ≤5
- **Validation**:
```bash
go-stats-generator analyze . --skip-tests --format json --output /tmp/metrics.json --sections functions && \
jq '.functions[] | select(.name == "refreshGameState") | {name, lines: .lines.code, complexity: .complexity.cyclomatic}' /tmp/metrics.json
```

**Refactoring Plan**:
1. Extract `(g *Game) updateCharacterState(state *GameState)`
2. Extract `(g *Game) updateMapView(state *GameState)`
3. Extract `(g *Game) updateCombatLog(events []Event)`
4. Extract `(g *Game) updateUIPanels(state *GameState)`

---

### Step 5: Refactor Reputation Modifier
- **Deliverable**: Split `ModifyReputation` (89 lines, complexity 14) into calculation helpers
- **Dependencies**: None (isolated function)
- **Acceptance**: `ModifyReputation` ≤40 lines, complexity ≤8, 100% test coverage for reputation logic
- **Validation**:
```bash
go-stats-generator analyze . --skip-tests --format json --output /tmp/metrics.json --sections functions && \
jq '.functions[] | select(.name == "ModifyReputation") | {name, lines: .lines.code, complexity: .complexity.cyclomatic}' /tmp/metrics.json
```

**Refactoring Plan**:
1. Extract `calculateReputationTier(current, delta int) (newTier int, crossed bool)`
2. Extract `updateFactionRelationships(faction *Faction, delta int) []RelationshipChange`
3. Extract `generateReputationEvents(changes []RelationshipChange) []PCGEvent`
4. Extract `recordReputationMetrics(faction *Faction, oldTier, newTier int)`

---

## Phase 1 Acceptance Criteria (All Steps)

Run full validation after completing all steps:

```bash
# Generate fresh metrics
go-stats-generator analyze . --skip-tests --format json --output /tmp/phase1_metrics.json --sections functions,packages

# Verify complexity targets met
echo "=== Complexity Check (should be 0 P1 functions >10) ==="
jq '[.functions[] | select(.complexity.cyclomatic > 10) | select(.file | contains("server") or contains("wasmui") or contains("reputation"))] | length' /tmp/phase1_metrics.json

# Verify line count targets met  
echo "=== Line Count Check (P1 functions) ==="
jq '.functions[] | select(.name == "handleMethod" or .name == "handleAttack" or .name == "NewRPCServer" or .name == "refreshGameState" or .name == "ModifyReputation") | {name, lines: .lines.code, complexity: .complexity.cyclomatic}' /tmp/phase1_metrics.json

# Run tests with race detector
go test -race ./pkg/server/... ./pkg/wasmui/... ./pkg/pcg/...

# Verify no regressions in E2E
go test -v ./test/e2e/...
```

### Success Metrics
| Function | Current | Target | 
|----------|---------|--------|
| `handleMethod` | 140 lines | ≤30 lines |
| `handleAttack` | 80 lines | ≤30 lines |
| `NewRPCServer` | 63 lines, complexity 12 | ≤25 lines, complexity ≤5 |
| `refreshGameState` | 51 lines, complexity 14 | ≤20 lines, complexity ≤5 |
| `ModifyReputation` | 89 lines, complexity 14 | ≤40 lines, complexity ≤8 |

---

## Estimated Effort
| Step | Hours | Dependencies |
|------|-------|--------------|
| Step 1: handleMethod | 16 | None |
| Step 2: handleAttack | 12 | Step 1 |
| Step 3: NewRPCServer | 8 | None |
| Step 4: refreshGameState | 12 | None |
| Step 5: ModifyReputation | 12 | None |
| **Total** | **60** | 2-3 weeks |

---

## Out of Scope (Deferred to Later Phases)

The following items are documented in ROADMAP.md but not included in Phase 1:

### Phase 2: Metrics Initialization Cleanup
- `NewMetrics` (148 lines) - P2 priority, post-launch
- `LoadWithSecrets` (87 lines) - P2 priority
- `getQuestTemplates` (144 lines) - P3 priority

### Phase 3: PCG Algorithm Refinement
- `Generate` (109 lines) in pkg/pcg/character.go
- `GenerateLevel` (74 lines, complexity 13)
- `addVegetation` (57 lines, complexity 14)

### Phase 4: Concurrency Lifecycle
- Context cancellation for server goroutines
- Graceful WebSocket shutdown

### Phase 5: Code Hygiene
- BUG annotation cleanup (5 annotations)
- Package naming conventions (41 low-severity violations)

---

## Validation Commands Reference

```bash
# Full codebase analysis
go-stats-generator analyze . --skip-tests --format json --output metrics.json --sections functions,duplication,documentation,packages,patterns

# Check specific function complexity
jq '.functions[] | select(.complexity.cyclomatic > 9) | {name, file, complexity: .complexity.cyclomatic, lines: .lines.code}' metrics.json

# Check functions over line threshold
jq '[.functions[] | select(.lines.code > 30)] | sort_by(-.lines.code) | .[0:20] | .[] | {name, file, lines: .lines.code}' metrics.json

# Documentation coverage
jq '.documentation.coverage' metrics.json

# Duplication ratio
jq '.duplication.duplication_ratio' metrics.json
```

---

**Generated**: 2026-03-11  
**Source**: go-stats-generator metrics + ROADMAP.md remediation phases  
**Next Review**: After Phase 1 completion
