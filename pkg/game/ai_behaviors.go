// Package game provides behavior tree framework for NPC AI.
package game

import (
	"math"

	"github.com/sirupsen/logrus"
)

// BehaviorStatus represents the result of a behavior node tick.
type BehaviorStatus int

const (
	// StatusSuccess indicates the behavior completed successfully.
	StatusSuccess BehaviorStatus = iota
	// StatusFailure indicates the behavior failed.
	StatusFailure
	// StatusRunning indicates the behavior is still in progress.
	StatusRunning
)

// String returns the string representation of a behavior status.
func (s BehaviorStatus) String() string {
	switch s {
	case StatusSuccess:
		return "Success"
	case StatusFailure:
		return "Failure"
	case StatusRunning:
		return "Running"
	default:
		return "Unknown"
	}
}

// BehaviorContext provides context for behavior tree execution.
// It contains references to game systems needed by behavior nodes.
type BehaviorContext struct {
	World      *World
	PathFinder *PathFinder
	CombatAI   *CombatAI
	Enemies    []*Character
	Allies     []*Character
	Target     *Character
	TargetPos  *Position
	DeltaTime  float64
}

// BehaviorNode represents a node in a behavior tree.
// All behavior tree nodes must implement this interface.
type BehaviorNode interface {
	// Tick executes the behavior for one frame.
	// Returns the status of the behavior after this tick.
	Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus
}

// SequenceNode executes children in order until one fails.
// Returns Success only if all children succeed.
// Returns Failure as soon as any child fails.
type SequenceNode struct {
	Name     string
	Children []BehaviorNode
}

// NewSequenceNode creates a new sequence node with the given children.
func NewSequenceNode(name string, children ...BehaviorNode) *SequenceNode {
	return &SequenceNode{
		Name:     name,
		Children: children,
	}
}

// Tick executes children in sequence until one fails or all succeed.
func (s *SequenceNode) Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	logrus.WithFields(logrus.Fields{
		"function": "SequenceNode.Tick",
		"node":     s.Name,
		"npc_id":   npc.ID,
	}).Debug("executing sequence node")

	for i, child := range s.Children {
		status := child.Tick(npc, ctx)
		if status != StatusSuccess {
			logrus.WithFields(logrus.Fields{
				"function":    "SequenceNode.Tick",
				"node":        s.Name,
				"child_index": i,
				"status":      status.String(),
			}).Debug("child returned non-success")
			return status
		}
	}
	return StatusSuccess
}

// SelectorNode executes children until one succeeds.
// Returns Success as soon as any child succeeds.
// Returns Failure only if all children fail.
type SelectorNode struct {
	Name     string
	Children []BehaviorNode
}

// NewSelectorNode creates a new selector node with the given children.
func NewSelectorNode(name string, children ...BehaviorNode) *SelectorNode {
	return &SelectorNode{
		Name:     name,
		Children: children,
	}
}

// Tick executes children in order until one succeeds.
func (s *SelectorNode) Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	logrus.WithFields(logrus.Fields{
		"function": "SelectorNode.Tick",
		"node":     s.Name,
		"npc_id":   npc.ID,
	}).Debug("executing selector node")

	for i, child := range s.Children {
		status := child.Tick(npc, ctx)
		if status != StatusFailure {
			logrus.WithFields(logrus.Fields{
				"function":    "SelectorNode.Tick",
				"node":        s.Name,
				"child_index": i,
				"status":      status.String(),
			}).Debug("child returned non-failure")
			return status
		}
	}
	return StatusFailure
}

// InverterNode inverts the result of its child node.
// Success becomes Failure and Failure becomes Success.
// Running status is passed through unchanged.
type InverterNode struct {
	Child BehaviorNode
}

// NewInverterNode creates an inverter that wraps the given child.
func NewInverterNode(child BehaviorNode) *InverterNode {
	return &InverterNode{Child: child}
}

// Tick executes the child and inverts its result.
func (i *InverterNode) Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	status := i.Child.Tick(npc, ctx)
	switch status {
	case StatusSuccess:
		return StatusFailure
	case StatusFailure:
		return StatusSuccess
	default:
		return status
	}
}

// RepeatNode repeats its child a specified number of times.
type RepeatNode struct {
	Child       BehaviorNode
	RepeatCount int
	current     int
}

// NewRepeatNode creates a node that repeats its child n times.
func NewRepeatNode(child BehaviorNode, count int) *RepeatNode {
	return &RepeatNode{
		Child:       child,
		RepeatCount: count,
		current:     0,
	}
}

// Tick executes the child repeatedly until count is reached.
func (r *RepeatNode) Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	if r.current >= r.RepeatCount {
		r.current = 0
		return StatusSuccess
	}

	status := r.Child.Tick(npc, ctx)
	if status == StatusFailure {
		r.current = 0
		return StatusFailure
	}

	if status == StatusSuccess {
		r.current++
	}
	return StatusRunning
}

// ConditionNode evaluates a condition and returns Success or Failure.
type ConditionNode struct {
	Name      string
	Condition func(npc *NPC, ctx *BehaviorContext) bool
}

// NewConditionNode creates a condition node with the given predicate.
func NewConditionNode(name string, condition func(npc *NPC, ctx *BehaviorContext) bool) *ConditionNode {
	return &ConditionNode{
		Name:      name,
		Condition: condition,
	}
}

// Tick evaluates the condition and returns Success or Failure.
func (c *ConditionNode) Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	result := c.Condition(npc, ctx)
	logrus.WithFields(logrus.Fields{
		"function": "ConditionNode.Tick",
		"node":     c.Name,
		"result":   result,
	}).Debug("condition evaluated")

	if result {
		return StatusSuccess
	}
	return StatusFailure
}

// ActionNode executes an action and returns its result.
type ActionNode struct {
	Name   string
	Action func(npc *NPC, ctx *BehaviorContext) BehaviorStatus
}

// NewActionNode creates an action node with the given action function.
func NewActionNode(name string, action func(npc *NPC, ctx *BehaviorContext) BehaviorStatus) *ActionNode {
	return &ActionNode{
		Name:   name,
		Action: action,
	}
}

// Tick executes the action and returns its status.
func (a *ActionNode) Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	logrus.WithFields(logrus.Fields{
		"function": "ActionNode.Tick",
		"node":     a.Name,
	}).Debug("executing action")

	return a.Action(npc, ctx)
}

// Built-in condition factories

// HealthBelowThreshold creates a condition that checks if NPC health is below a threshold.
func HealthBelowThreshold(threshold float64) func(npc *NPC, ctx *BehaviorContext) bool {
	return func(npc *NPC, ctx *BehaviorContext) bool {
		maxHP := npc.Character.MaxHP
		currentHP := npc.Character.HP
		if maxHP == 0 {
			return false
		}
		ratio := float64(currentHP) / float64(maxHP)
		return ratio < threshold
	}
}

// DistanceToTargetBelow creates a condition that checks distance to target.
func DistanceToTargetBelow(maxDist float64) func(npc *NPC, ctx *BehaviorContext) bool {
	return func(npc *NPC, ctx *BehaviorContext) bool {
		if ctx.Target == nil {
			return false
		}
		npcPos := npc.Character.Position
		targetPos := ctx.Target.Position
		dist := math.Sqrt(float64((npcPos.X-targetPos.X)*(npcPos.X-targetPos.X) +
			(npcPos.Y-targetPos.Y)*(npcPos.Y-targetPos.Y)))
		return dist <= maxDist
	}
}

// DistanceToTargetAbove creates a condition that checks distance to target is above threshold.
func DistanceToTargetAbove(minDist float64) func(npc *NPC, ctx *BehaviorContext) bool {
	return func(npc *NPC, ctx *BehaviorContext) bool {
		if ctx.Target == nil {
			return false
		}
		npcPos := npc.Character.Position
		targetPos := ctx.Target.Position
		dist := math.Sqrt(float64((npcPos.X-targetPos.X)*(npcPos.X-targetPos.X) +
			(npcPos.Y-targetPos.Y)*(npcPos.Y-targetPos.Y)))
		return dist > minDist
	}
}

// HasEnemiesNearby creates a condition that checks if living enemies are within range.
func HasEnemiesNearby(range_ float64) func(npc *NPC, ctx *BehaviorContext) bool {
	return func(npc *NPC, ctx *BehaviorContext) bool {
		npcPos := npc.Character.Position
		for _, enemy := range ctx.Enemies {
			if enemy.HP <= 0 {
				continue // Skip dead enemies
			}
			dist := math.Sqrt(float64((npcPos.X-enemy.Position.X)*(npcPos.X-enemy.Position.X) +
				(npcPos.Y-enemy.Position.Y)*(npcPos.Y-enemy.Position.Y)))
			if dist <= range_ {
				return true
			}
		}
		return false
	}
}

// AllyCountNearby creates a condition that checks if enough allies are nearby.
func AllyCountNearby(minCount int, range_ float64) func(npc *NPC, ctx *BehaviorContext) bool {
	return func(npc *NPC, ctx *BehaviorContext) bool {
		count := 0
		npcPos := npc.Character.Position
		for _, ally := range ctx.Allies {
			dist := math.Sqrt(float64((npcPos.X-ally.Position.X)*(npcPos.X-ally.Position.X) +
				(npcPos.Y-ally.Position.Y)*(npcPos.Y-ally.Position.Y)))
			if dist <= range_ {
				count++
			}
		}
		return count >= minCount
	}
}

// HasTarget creates a condition that checks if a target is set.
func HasTarget() func(npc *NPC, ctx *BehaviorContext) bool {
	return func(npc *NPC, ctx *BehaviorContext) bool {
		return ctx.Target != nil
	}
}

// IsTargetAlive creates a condition that checks if the target is alive.
func IsTargetAlive() func(npc *NPC, ctx *BehaviorContext) bool {
	return func(npc *NPC, ctx *BehaviorContext) bool {
		return ctx.Target != nil && ctx.Target.HP > 0
	}
}

// Built-in action factories

// MoveToTarget creates an action that moves the NPC toward its target.
func MoveToTarget() func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	return func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
		if ctx.Target == nil || ctx.PathFinder == nil {
			return StatusFailure
		}

		start := npc.Character.Position
		end := ctx.Target.Position

		path, found := ctx.PathFinder.FindPath(start, end)
		if !found || len(path) < 2 {
			return StatusFailure
		}

		// Move to the next step in path
		npc.Character.Position = path[1]

		logrus.WithFields(logrus.Fields{
			"function": "MoveToTarget",
			"npc_id":   npc.ID,
			"new_pos":  path[1],
		}).Debug("NPC moved toward target")

		return StatusSuccess
	}
}

// MoveToPosition creates an action that moves the NPC toward a specific position.
func MoveToPosition() func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	return func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
		if ctx.TargetPos == nil || ctx.PathFinder == nil {
			return StatusFailure
		}

		start := npc.Character.Position
		end := *ctx.TargetPos

		path, found := ctx.PathFinder.FindPath(start, end)
		if !found || len(path) < 2 {
			return StatusFailure
		}

		// Move to the next step in path
		npc.Character.Position = path[1]

		return StatusSuccess
	}
}

// AttackTarget creates an action that attacks the current target.
func AttackTarget() func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	return func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
		if ctx.Target == nil {
			return StatusFailure
		}

		// Check if in melee range (adjacent)
		npcPos := npc.Character.Position
		targetPos := ctx.Target.Position
		dx := abs(npcPos.X - targetPos.X)
		dy := abs(npcPos.Y - targetPos.Y)

		if dx > 1 || dy > 1 {
			// Not in range
			return StatusFailure
		}

		// Perform attack (simplified - in real implementation would use combat system)
		logrus.WithFields(logrus.Fields{
			"function":  "AttackTarget",
			"npc_id":    npc.ID,
			"target_id": ctx.Target.ID,
		}).Debug("NPC attacked target")

		return StatusSuccess
	}
}

// Flee creates an action that moves the NPC away from threats.
func Flee() func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	return func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
		if ctx.PathFinder == nil || ctx.CombatAI == nil || ctx.World == nil {
			return StatusFailure
		}

		// Use combat AI to find retreat position
		retreatPos := ctx.CombatAI.findRetreatPosition(npc, ctx.Enemies, ctx.World)

		start := npc.Character.Position
		path, found := ctx.PathFinder.FindPath(start, retreatPos)
		if !found || len(path) < 2 {
			return StatusFailure
		}

		npc.Character.Position = path[1]

		logrus.WithFields(logrus.Fields{
			"function":   "Flee",
			"npc_id":     npc.ID,
			"retreat_to": path[1],
		}).Debug("NPC fleeing")

		return StatusSuccess
	}
}

// Patrol creates an action that makes the NPC patrol between waypoints.
func Patrol(waypoints []Position) func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	currentWaypoint := 0
	return func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
		if len(waypoints) == 0 || ctx.PathFinder == nil {
			return StatusFailure
		}

		target := waypoints[currentWaypoint]
		npcPos := npc.Character.Position

		// Check if at waypoint
		if npcPos.X == target.X && npcPos.Y == target.Y {
			currentWaypoint = (currentWaypoint + 1) % len(waypoints)
			target = waypoints[currentWaypoint]
		}

		// Move toward waypoint
		path, found := ctx.PathFinder.FindPath(npcPos, target)
		if !found || len(path) < 2 {
			return StatusFailure
		}

		npc.Character.Position = path[1]
		return StatusRunning
	}
}

// Idle creates an action that does nothing (succeeds immediately).
func Idle() func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	return func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
		return StatusSuccess
	}
}

// SelectNearestEnemy creates an action that selects the nearest living enemy as target.
func SelectNearestEnemy() func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	return func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
		if len(ctx.Enemies) == 0 {
			ctx.Target = nil
			return StatusFailure
		}

		npcPos := npc.Character.Position
		var nearest *Character
		minDist := math.MaxFloat64

		for _, enemy := range ctx.Enemies {
			if enemy.HP <= 0 {
				continue // Skip dead enemies
			}
			dist := math.Sqrt(float64((npcPos.X-enemy.Position.X)*(npcPos.X-enemy.Position.X) +
				(npcPos.Y-enemy.Position.Y)*(npcPos.Y-enemy.Position.Y)))
			if dist < minDist {
				minDist = dist
				nearest = enemy
			}
		}

		if nearest == nil {
			ctx.Target = nil
			return StatusFailure
		}

		ctx.Target = nearest
		return StatusSuccess
	}
}

// BehaviorTreeBuilder provides a fluent API for building behavior trees.
type BehaviorTreeBuilder struct {
	root BehaviorNode
}

// NewBehaviorTreeBuilder creates a new builder.
func NewBehaviorTreeBuilder() *BehaviorTreeBuilder {
	return &BehaviorTreeBuilder{}
}

// Sequence starts building a sequence node.
func (b *BehaviorTreeBuilder) Sequence(name string, children ...BehaviorNode) *BehaviorTreeBuilder {
	b.root = NewSequenceNode(name, children...)
	return b
}

// Selector starts building a selector node.
func (b *BehaviorTreeBuilder) Selector(name string, children ...BehaviorNode) *BehaviorTreeBuilder {
	b.root = NewSelectorNode(name, children...)
	return b
}

// Build returns the constructed behavior tree.
func (b *BehaviorTreeBuilder) Build() BehaviorNode {
	return b.root
}

// StandardBehaviorTrees provides pre-built behavior trees for common NPC types.
type StandardBehaviorTrees struct{}

// AggressiveTree returns a behavior tree for aggressive NPCs.
// They seek enemies and attack, retreating when low on health.
func (StandardBehaviorTrees) AggressiveTree() BehaviorNode {
	return NewSelectorNode("Aggressive",
		// If low health and enemies nearby, flee
		NewSequenceNode("FleeWhenHurt",
			NewConditionNode("HealthLow", HealthBelowThreshold(0.25)),
			NewConditionNode("EnemiesNear", HasEnemiesNearby(5)),
			NewActionNode("Flee", Flee()),
		),
		// If have target in range, attack
		NewSequenceNode("AttackNearTarget",
			NewConditionNode("HasTarget", HasTarget()),
			NewConditionNode("TargetAlive", IsTargetAlive()),
			NewConditionNode("InRange", DistanceToTargetBelow(1.5)),
			NewActionNode("Attack", AttackTarget()),
		),
		// If have target out of range, approach
		NewSequenceNode("ApproachTarget",
			NewConditionNode("HasTarget", HasTarget()),
			NewConditionNode("TargetAlive", IsTargetAlive()),
			NewActionNode("MoveToTarget", MoveToTarget()),
		),
		// Find a target
		NewSequenceNode("FindTarget",
			NewConditionNode("EnemiesExist", HasEnemiesNearby(20)),
			NewActionNode("SelectEnemy", SelectNearestEnemy()),
		),
		// Default: idle
		NewActionNode("Idle", Idle()),
	)
}

// GuardTree returns a behavior tree for guard NPCs.
// They stay in position until enemies approach, then engage.
func (StandardBehaviorTrees) GuardTree() BehaviorNode {
	return NewSelectorNode("Guard",
		// If enemies are close, engage
		NewSequenceNode("EngageIntruders",
			NewConditionNode("IntrudersNear", HasEnemiesNearby(10)),
			NewActionNode("SelectEnemy", SelectNearestEnemy()),
			NewSelectorNode("CombatOrChase",
				NewSequenceNode("Attack",
					NewConditionNode("InRange", DistanceToTargetBelow(1.5)),
					NewActionNode("Attack", AttackTarget()),
				),
				NewSequenceNode("Chase",
					NewConditionNode("NotTooFar", DistanceToTargetBelow(15)),
					NewActionNode("Chase", MoveToTarget()),
				),
			),
		),
		// Default: stand guard (idle)
		NewActionNode("StandGuard", Idle()),
	)
}

// PatrolTree returns a behavior tree for patrol NPCs with given waypoints.
func (StandardBehaviorTrees) PatrolTree(waypoints []Position) BehaviorNode {
	return NewSelectorNode("Patrol",
		// If enemies spotted, engage
		NewSequenceNode("EngageEnemies",
			NewConditionNode("EnemiesNear", HasEnemiesNearby(8)),
			NewActionNode("SelectEnemy", SelectNearestEnemy()),
			NewSelectorNode("FightOrChase",
				NewSequenceNode("Fight",
					NewConditionNode("InRange", DistanceToTargetBelow(1.5)),
					NewActionNode("Attack", AttackTarget()),
				),
				NewSequenceNode("Chase",
					NewConditionNode("NotTooFar", DistanceToTargetBelow(12)),
					NewActionNode("Chase", MoveToTarget()),
				),
			),
		),
		// Default: patrol
		NewActionNode("Patrol", Patrol(waypoints)),
	)
}

// CowardTree returns a behavior tree for cowardly NPCs.
// They flee from combat and only fight when cornered.
func (StandardBehaviorTrees) CowardTree() BehaviorNode {
	return NewSelectorNode("Coward",
		// Always try to flee if enemies nearby
		NewSequenceNode("FleeFromDanger",
			NewConditionNode("EnemiesNear", HasEnemiesNearby(10)),
			NewActionNode("Flee", Flee()),
		),
		// If cornered (can't flee) and enemy in range, fight
		NewSequenceNode("FightWhenCornered",
			NewConditionNode("EnemyClose", HasEnemiesNearby(2)),
			NewActionNode("SelectEnemy", SelectNearestEnemy()),
			NewConditionNode("InRange", DistanceToTargetBelow(1.5)),
			NewActionNode("DesperateAttack", AttackTarget()),
		),
		// Otherwise idle
		NewActionNode("Hide", Idle()),
	)
}
