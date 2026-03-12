# Implementation Plan: NPC AI Behaviors & Combat Enhancement

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by SSI Gold Box games, providing comprehensive character management, combat systems, and world interactions through JSON-RPC API with WebSocket support for real-time communication.
- **Current goal**: Implement Advanced NPC AI Behaviors (roadmap item marked ❌ Missing)
- **Estimated Scope**: Medium (13 functions above complexity threshold 9, 759 duplicated lines, 83.9% doc coverage)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Core RPG mechanics and character system | ✅ Achieved | No |
| Combat and effect systems | ✅ Achieved | Partially (enhanced combat) |
| WebSocket real-time communication | ✅ Achieved | No |
| Procedural Content Generation system | ✅ Achieved | No |
| Circuit breaker patterns and resilience | ✅ Achieved | No |
| Comprehensive input validation | ✅ Achieved | No |
| Health monitoring and metrics | ✅ Achieved | No |
| Asset generation pipeline | ⚠️ Partial (6/521 assets) | No |
| **Advanced NPC AI behaviors** | ❌ Missing | **Yes** |
| Enhanced combat mechanics | ⚠️ Partial | Yes |
| Additional spell effects | ⚠️ Partial (3/10 level files) | No |
| World editor tools | ❌ Missing | No |
| Network optimization | ⚠️ Partial | No |
| Content creation utilities | ⚠️ Partial | No |
| Player progression persistence | ✅ Achieved | No |
| Guild and faction systems | ⚠️ Partial | No |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 13 functions with cyclomatic > 9
  - Highest: `addVegetation` (14), `refreshGameState` (14), `GenerateLevel` (13)
  - Combat-related: None above threshold (existing combat.go is well-factored)
- **Duplication ratio**: 759 lines / 28 clone pairs (primarily in `pkg/validation/validation.go`)
- **Doc coverage**: 83.9% overall, 93.1% functions, 79.4% methods
- **Package coupling**: Clean package boundaries, no circular dependencies detected
- **TODOs in codebase**: 3 (territory generation, version from build info, vault provider)

## Implementation Steps

### Step 1: Create Pathfinding Foundation
- **Deliverable**: New file `pkg/game/pathfinding.go` implementing A* algorithm (~200-300 lines)
  - `type PathFinder struct` with reference to existing `SpatialIndex`
  - `func (pf *PathFinder) FindPath(start, end Position, world *World) []Position`
  - `func (pf *PathFinder) CanReach(start, end Position, world *World) bool`
  - Integrate with existing terrain walkability from `pkg/game/map.go`
- **Dependencies**: None (uses existing `SpatialIndex` from `pkg/game/spatial_index.go`)
- **Goal Impact**: Foundational for NPC AI movement, patrol routes, and combat positioning
- **Acceptance**: Cyclomatic complexity < 15 per function; 85%+ test coverage on new file
- **Validation**: 
  ```bash
  go-stats-generator analyze pkg/game/pathfinding.go --format json | jq '.functions[] | select(.complexity.cyclomatic > 15)'
  go test -cover ./pkg/game/... | grep pathfinding
  ```

### Step 2: Implement Combat AI Decision Engine
- **Deliverable**: New file `pkg/game/ai_combat.go` with tactical decision-making (~400-500 lines)
  - `type CombatAI struct` with difficulty tiers (Easy, Medium, Hard)
  - `func (ai *CombatAI) SelectTarget(npc *NPC, enemies []*Character, world *World) *Character`
  - `func (ai *CombatAI) ChooseAction(npc *NPC, gameState *GameState) Action`
  - `func (ai *CombatAI) ShouldRetreat(npc *NPC, threats []*Character) bool`
  - Use existing `NPCBehavior` enum (Idle, Patrol, Guard, Aggressive) from `pkg/game/world_types.go`
- **Dependencies**: Step 1 (pathfinding for retreat routes)
- **Goal Impact**: Directly addresses "Advanced NPC AI behaviors" roadmap item
- **Acceptance**: NPCs make tactical decisions in combat; difficulty affects target selection quality
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/game/ai_combat.go --format json | jq '.functions[] | select(.complexity.cyclomatic > 15)'
  go test -v ./pkg/game/... -run TestCombatAI
  ```

### Step 3: Add Behavior Tree Framework
- **Deliverable**: New file `pkg/game/ai_behaviors.go` with composable behavior nodes (~300-400 lines)
  - `type BehaviorNode interface { Tick(npc *NPC, ctx *BehaviorContext) Status }`
  - `type SequenceNode struct`, `type SelectorNode struct`, `type ConditionNode struct`
  - Built-in conditions: `HealthBelowThreshold`, `DistanceFromTarget`, `AllyCountNearby`
  - Built-in actions: `MoveToTarget`, `AttackTarget`, `Flee`, `Patrol`
- **Dependencies**: Steps 1 and 2 (uses pathfinding and combat decisions)
- **Goal Impact**: Enables designer-controllable NPC behaviors via composition
- **Acceptance**: Behavior trees execute correctly; NPC behavior changes based on conditions
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/game/ai_behaviors.go --format json | jq '.functions[] | select(.complexity.cyclomatic > 15)'
  go test -v ./pkg/game/... -run TestBehaviorTree
  ```

### Step 4: Implement Opportunity Attacks
- **Deliverable**: Modifications to `pkg/game/combat.go` and new file `pkg/game/combat_opportunity.go` (~150 lines)
  - `func CheckOpportunityAttack(mover *Character, adjacentEnemies []*Character) *Attack`
  - Add "Disengage" action to movement commands
  - Hook into existing turn system in `pkg/server/turn.go`
- **Dependencies**: Step 2 (combat AI needs to consider opportunity attacks)
- **Goal Impact**: Addresses "Enhanced combat mechanics" roadmap item
- **Acceptance**: Moving away from enemies triggers opportunity attacks unless disengaging
- **Validation**:
  ```bash
  go test -v ./pkg/game/... -run TestOpportunityAttack
  go test -v ./test/e2e/... -run TestCombatOpportunity
  ```

### Step 5: Add Cover and Flanking Mechanics
- **Deliverable**: New file `pkg/game/combat_modifiers.go` (~200 lines)
  - `func CalculateCoverBonus(attacker, defender *Character, world *World) int`
  - `func CalculateFlankingBonus(attacker, defender *Character, allies []*Character) int`
  - Integrate with existing attack roll calculations in `pkg/game/combat.go`
  - Use spatial index for efficient adjacent character detection
- **Dependencies**: Step 1 (spatial queries for positioning)
- **Goal Impact**: Addresses "Enhanced combat mechanics" with tactical positioning
- **Acceptance**: Cover provides AC bonus; flanking (2+ allies opposite sides) provides attack bonus
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/game/combat_modifiers.go --format json | jq '.functions[] | select(.complexity.cyclomatic > 15)'
  go test -v ./pkg/game/... -run TestCoverFlanking
  ```

### Step 6: Implement Morale System
- **Deliverable**: New file `pkg/game/morale.go` (~250 lines)
  - `type MoraleTracker struct` per NPC/party
  - `func (mt *MoraleTracker) UpdateMorale(event CombatEvent)`
  - Events: ally death (-), damage taken (-), enemy killed (+), overwhelming odds (-)
  - `func (mt *MoraleTracker) CheckMoraleBreak() bool` triggers flee behavior
  - Tie morale resistance to existing Wisdom/Charisma attributes
- **Dependencies**: Steps 2 and 3 (morale break triggers AI retreat behavior)
- **Goal Impact**: Adds tactical depth to combat; NPCs react realistically to battlefield conditions
- **Acceptance**: NPCs flee when morale breaks; morale recovers over time or with rally actions
- **Validation**:
  ```bash
  go test -v ./pkg/game/... -run TestMorale
  go test -v ./test/e2e/... -run TestMoraleInCombat
  ```

### Step 7: Create E2E Integration Tests
- **Deliverable**: New test file `test/e2e/ai_combat_test.go` (~400 lines)
  - `TestNPCPathfinding`: NPC navigates around obstacles to reach target
  - `TestCombatAITargetSelection`: AI selects optimal targets based on difficulty
  - `TestBehaviorTreeExecution`: Guard NPC patrols, detects player, attacks
  - `TestTacticalCombat`: Full combat with opportunity attacks, cover, flanking, morale
- **Dependencies**: Steps 1-6 completed
- **Goal Impact**: Validates all AI and combat features work together
- **Acceptance**: All E2E tests pass; demonstrates NPC winning tactical combat scenario
- **Validation**:
  ```bash
  go test -v -race ./test/e2e/... -run TestAICombat
  go test -cover ./test/e2e/... | grep ai_combat
  ```

### Step 8: Update Documentation
- **Deliverable**: Updates to `pkg/README-RPC.md` and new `docs/NPC_AI.md` (~300 lines)
  - Document new AI-related RPC methods if any
  - Explain behavior tree system for designers
  - Combat mechanics reference (opportunity attacks, cover, flanking, morale)
  - Integration guide for extending AI behaviors
- **Dependencies**: Steps 1-7 completed
- **Goal Impact**: Enables developers and designers to use and extend AI system
- **Acceptance**: Documentation covers all new features; includes code examples
- **Validation**:
  ```bash
  # Verify doc coverage improvement
  go-stats-generator analyze ./pkg/game --format json --sections documentation | jq '.documentation.coverage.overall'
  ```

## Default Thresholds (Calibrated to Project)
| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| Functions above complexity 9.0 | <5 | 5-15 | >15 |
| Duplication ratio | <3% | 3-10% | >10% |
| Doc coverage gap | <10% | 10-25% | >25% |

**Current Status**: Medium scope (13 high-complexity functions, ~3% duplication, 16.1% doc gap)

## Quality Gates (Per CI Configuration)
- ✅ All CI checks pass (tests, lint, format, security)
- ✅ Test coverage ≥78% maintained
- ✅ No new `go vet` warnings
- ✅ No new cyclomatic complexity >15 functions
- ✅ Race detector clean (`go test -race`)
- ✅ Docker health checks passing

## Timeline Estimate
| Step | Estimated Hours | Cumulative |
|------|-----------------|------------|
| 1. Pathfinding | 15-20 | 15-20 |
| 2. Combat AI | 25-30 | 40-50 |
| 3. Behavior Trees | 20-25 | 60-75 |
| 4. Opportunity Attacks | 8-10 | 68-85 |
| 5. Cover & Flanking | 10-15 | 78-100 |
| 6. Morale System | 15-20 | 93-120 |
| 7. E2E Tests | 15-20 | 108-140 |
| 8. Documentation | 8-12 | 116-152 |

**Total**: ~116-152 developer-hours

## Risk Mitigation
- **Complexity Risk**: Combat AI may exceed complexity threshold
  - Mitigation: Keep decision logic in small, focused functions; extract to helper methods early
- **Performance Risk**: Many NPCs with pathfinding may cause slowdown
  - Mitigation: Leverage existing spatial index; add pathfinding result caching; profile with 50+ NPCs
- **Integration Risk**: New combat mechanics may break existing tests
  - Mitigation: Run full test suite after each step; use feature flags if needed

## Success Criteria
1. NPC navigates maze using A* pathfinding (Step 1)
2. Combat AI defeats player in tactical combat demonstrating target selection (Step 2)
3. Guard NPC patrols, detects intruder, and pursues using behavior tree (Step 3)
4. Player taking opportunity attack when fleeing enemy (Step 4)
5. Combat shows cover AC bonus when behind terrain (Step 5)
6. NPC retreats when morale breaks from ally deaths (Step 6)
7. All E2E tests pass demonstrating integrated AI combat (Step 7)

---

*Generated: 2026-03-12 | Based on go-stats-generator metrics and ROADMAP.md goal assessment*
