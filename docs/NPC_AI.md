# NPC AI System Documentation

This document describes the NPC AI system in GoldBox RPG Engine, including behavior trees, combat AI, pathfinding, and tactical mechanics.

## Overview

The NPC AI system provides intelligent, configurable behavior for non-player characters through a composable behavior tree architecture combined with tactical combat decision-making.

### Key Components

| Component | File | Purpose |
|-----------|------|---------|
| Pathfinding | `pkg/game/pathfinding.go` | A* pathfinding algorithm for navigation |
| Combat AI | `pkg/game/ai_combat.go` | Tactical decision engine with difficulty tiers |
| Behavior Trees | `pkg/game/ai_behaviors.go` | Composable behavior nodes |
| Opportunity Attacks | `pkg/game/combat_opportunity.go` | Movement-triggered attacks |
| Combat Modifiers | `pkg/game/combat_modifiers.go` | Cover and flanking mechanics |
| Morale System | `pkg/game/morale.go` | Morale tracking and break conditions |

## Behavior Tree System

### Core Concepts

Behavior trees are hierarchical structures that control NPC decision-making. Each node returns one of three statuses:

- **Success**: The behavior completed successfully
- **Failure**: The behavior failed
- **Running**: The behavior is still in progress

### Node Types

#### Composite Nodes

**SequenceNode** - Executes children in order until one fails (AND logic):
```go
sequence := NewSequenceNode("AttackSequence",
    NewConditionNode("HasTarget", HasTarget()),
    NewConditionNode("InRange", DistanceToTargetBelow(1.5)),
    NewActionNode("Attack", AttackTarget()),
)
```

**SelectorNode** - Executes children until one succeeds (OR logic):
```go
selector := NewSelectorNode("CombatOptions",
    NewSequenceNode("Attack", ...),
    NewSequenceNode("Chase", ...),
    NewActionNode("Idle", Idle()),
)
```

#### Decorator Nodes

**InverterNode** - Inverts child result:
```go
inverter := NewInverterNode(
    NewConditionNode("HealthLow", HealthBelowThreshold(0.25)),
)
```

**RepeatNode** - Repeats child N times:
```go
repeat := NewRepeatNode(patrolAction, 5)
```

#### Leaf Nodes

**ConditionNode** - Evaluates a condition:
```go
condition := NewConditionNode("HealthCheck", HealthBelowThreshold(0.25))
```

**ActionNode** - Executes an action:
```go
action := NewActionNode("Attack", AttackTarget())
```

### Built-in Conditions

| Condition | Parameters | Description |
|-----------|------------|-------------|
| `HealthBelowThreshold` | `threshold float64` | True if HP ratio < threshold |
| `DistanceToTargetBelow` | `maxDist float64` | True if distance to target < maxDist |
| `DistanceToTargetAbove` | `minDist float64` | True if distance to target > minDist |
| `HasEnemiesNearby` | `range float64` | True if living enemies within range |
| `AllyCountNearby` | `minCount int, range float64` | True if enough allies nearby |
| `HasTarget` | - | True if target is set |
| `IsTargetAlive` | - | True if target has HP > 0 |

### Built-in Actions

| Action | Description |
|--------|-------------|
| `MoveToTarget()` | Move toward current target |
| `MoveToPosition()` | Move toward TargetPos |
| `AttackTarget()` | Attack current target (melee) |
| `Flee()` | Retreat from enemies |
| `Patrol(waypoints)` | Patrol between waypoints |
| `Idle()` | Do nothing (always succeeds) |
| `SelectNearestEnemy()` | Set nearest living enemy as target |

### Standard Behavior Trees

The `StandardBehaviorTrees` struct provides pre-built trees:

```go
trees := StandardBehaviorTrees{}

// Aggressive: Attack enemies, flee when hurt
aggressive := trees.AggressiveTree()

// Guard: Engage intruders, stay at post
guard := trees.GuardTree()

// Patrol: Follow waypoints, engage nearby enemies
patrol := trees.PatrolTree([]Position{{X: 0, Y: 0}, {X: 10, Y: 0}})

// Coward: Always flee, fight only when cornered
coward := trees.CowardTree()
```

### Creating Custom Trees

Use the fluent builder API:

```go
tree := NewBehaviorTreeBuilder().
    Selector("MainBehavior",
        NewSequenceNode("CombatSequence",
            NewConditionNode("EnemyClose", HasEnemiesNearby(5)),
            NewActionNode("SelectTarget", SelectNearestEnemy()),
            NewActionNode("Attack", AttackTarget()),
        ),
        NewActionNode("DefaultIdle", Idle()),
    ).
    Build()
```

## Combat AI System

### Difficulty Tiers

The combat AI adapts its behavior based on difficulty:

| Difficulty | Target Selection | Tactics |
|------------|-----------------|---------|
| Easy | Random targeting | Basic positioning |
| Medium | Weakest enemy first | Some tactical awareness |
| Hard | Optimal target selection | Full tactical analysis |

### Target Selection

The AI considers multiple factors:
- Health (prefer low-HP targets)
- Distance (prefer closer targets)
- Threat level (high-damage enemies)
- Class (prioritize healers/mages)

### Action Selection

The `ChooseAction` method returns one of:
- `ActionAttack` - Attack current target
- `ActionMove` - Reposition tactically
- `ActionRetreat` - Flee from danger
- `ActionIdle` - Wait/defend

### Retreat Logic

NPCs retreat when:
- Health falls below threshold (default: 25%)
- Overwhelmed by multiple enemies
- Morale breaks

## Pathfinding System

### A* Algorithm

The pathfinding system uses A* with:
- Manhattan distance heuristic
- Terrain walkability from world data
- Multi-level support

### Usage

```go
pathfinder := NewPathFinder(world)

// Find complete path
path, found := pathfinder.FindPath(start, end)

// Check reachability
canReach := pathfinder.CanReach(start, end)
```

### Performance

- Uses priority queue for efficient node expansion
- Caches walkability checks
- Limits search depth for performance

## Tactical Combat Mechanics

### Opportunity Attacks

Moving away from an adjacent enemy triggers an opportunity attack:

```go
// Check for opportunity attacks
attacks := CheckOpportunityAttacks(mover, adjacentEnemies)

// Avoid with Disengage action
if action == ActionDisengage {
    // No opportunity attacks triggered
}
```

### Cover System

Cover provides AC bonuses:

| Cover Type | AC Bonus |
|------------|----------|
| Half Cover | +2 |
| Three-quarters | +5 |
| Full Cover | No attack possible |

```go
bonus := CalculateCoverBonus(attacker, defender, world)
```

### Flanking

Attackers gain bonuses when allies are positioned opposite the target:

```go
bonus := CalculateFlankingBonus(attacker, defender, allies)
// +2 attack bonus when flanking
```

### Requirements

- At least 2 attackers
- Positioned on opposite sides of target
- All combatants must be alive

## Morale System

### Morale Events

Events that affect morale:

| Event | Effect |
|-------|--------|
| Ally killed | -20 morale |
| Damage taken | -5 to -15 based on severity |
| Enemy killed | +10 morale |
| Rally action | +15 morale |
| Overwhelming odds | -10 morale |

### Morale Break

When morale drops to 0, the NPC may:
- Flee from combat
- Surrender
- Fight desperately

Resistance is based on:
- Wisdom score
- Charisma of nearby allies
- Character class bonuses

### Recovery

Morale recovers:
- +1 per turn out of combat
- +5 from rally actions
- Full recovery at combat end

## Integration Examples

### Complete NPC Setup

```go
// Create NPC with behavior tree
npc := &NPC{
    Character: Character{
        Name: "Guard Captain",
        HP: 100, MaxHP: 100,
        Position: Position{X: 10, Y: 10, Level: 0},
    },
    Behavior: "guard",
}

// Set up AI systems
pathfinder := NewPathFinder(world)
combatAI := NewCombatAI(DifficultyMedium)
behaviorTree := StandardBehaviorTrees{}.GuardTree()

// Create behavior context
ctx := &BehaviorContext{
    World:      world,
    PathFinder: pathfinder,
    CombatAI:   combatAI,
    Enemies:    playerParty,
    DeltaTime:  0.016, // 60 FPS
}

// Execute behavior each tick
func OnTick(npc *NPC, ctx *BehaviorContext) {
    status := behaviorTree.Tick(npc, ctx)
    // Handle status as needed
}
```

### Combat Turn Processing

```go
func ProcessNPCTurn(npc *NPC, ctx *BehaviorContext) {
    // 1. Check morale
    if moraleTracker.CheckMoraleBreak() {
        npc.SetBehavior("flee")
        return
    }

    // 2. Execute behavior tree
    status := behaviorTree.Tick(npc, ctx)

    // 3. Apply tactical modifiers
    if attacking {
        bonus := CalculateFlankingBonus(npc, target, allies)
        // Apply bonus to attack roll
    }

    // 4. Check for opportunity attacks
    if moved {
        oaAttacks := CheckOpportunityAttacks(npc, adjacentEnemies)
        // Process any attacks against NPC
    }
}
```

## Best Practices

### Behavior Tree Design

1. **Keep trees shallow** - Deep nesting hurts readability
2. **Use standard trees** - Customize only when needed
3. **Test incrementally** - Verify each node before combining
4. **Document complex trees** - Add comments for non-obvious logic

### Performance

1. **Cache pathfinding results** - Paths don't change often
2. **Limit vision checks** - Use spatial indexing
3. **Batch NPC updates** - Process groups together
4. **Profile regularly** - Monitor AI frame time

### Debugging

Enable AI logging:
```go
logrus.SetLevel(logrus.DebugLevel)
// AI decisions will be logged with node names and results
```

## Related Documentation

- [README-RPC.md](../pkg/README-RPC.md) - RPC method reference
- [ERROR_HANDLING.md](ERROR_HANDLING.md) - Error handling patterns
- [GRACEFUL_DEGRADATION.md](GRACEFUL_DEGRADATION.md) - Resilience patterns
