package game

import (
	"math"
	"math/rand"
	"sync"
)

// CombatAI provides tactical decision-making for NPCs in combat.
// It selects targets, chooses actions, and determines retreat conditions
// based on the NPC's behavior type and difficulty tier.
type CombatAI struct {
	mu         sync.RWMutex
	difficulty AIDifficulty
	pathfinder *PathFinder
}

// AIDifficulty represents the AI skill level affecting decision quality.
type AIDifficulty int

const (
	// AIDifficultyEasy makes predictable, suboptimal choices
	AIDifficultyEasy AIDifficulty = iota
	// AIDifficultyMedium makes reasonable tactical choices
	AIDifficultyMedium
	// AIDifficultyHard makes optimal tactical decisions
	AIDifficultyHard
)

// NewCombatAI creates a new combat AI instance with the specified difficulty.
func NewCombatAI(difficulty AIDifficulty, world *World) *CombatAI {
	return &CombatAI{
		difficulty: difficulty,
		pathfinder: NewPathFinder(world),
	}
}

// SelectTarget chooses the best target from available enemies based on tactical priorities.
// Returns nil if no valid targets are available.
func (ai *CombatAI) SelectTarget(npc *NPC, enemies []*Character, world *World) *Character {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	if len(enemies) == 0 {
		return nil
	}

	var candidates []*Character
	for _, enemy := range enemies {
		if enemy.HP > 0 && ai.canReach(npc.Position, enemy.Position) {
			candidates = append(candidates, enemy)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	switch ai.difficulty {
	case AIDifficultyEasy:
		return ai.selectRandomTarget(candidates)
	case AIDifficultyMedium:
		return ai.selectNearestTarget(npc, candidates)
	case AIDifficultyHard:
		return ai.selectOptimalTarget(npc, candidates)
	default:
		return ai.selectRandomTarget(candidates)
	}
}

// ChooseAction determines the best action for the NPC to take this turn.
// Returns the action type and target position.
func (ai *CombatAI) ChooseAction(npc *NPC, enemies []*Character, world *World) (ActionType, Position) {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	// Check if should retreat
	if ai.ShouldRetreat(npc, enemies) {
		return ActionRetreat, ai.findRetreatPosition(npc, enemies, world)
	}

	target := ai.SelectTarget(npc, enemies, world)
	if target == nil {
		return ActionIdle, npc.Position
	}

	// Calculate distance to target
	distance := ai.manhattanDistance(npc.Position, target.Position)

	// If in melee range (adjacent), attack
	if distance <= 1 {
		return ActionAttack, target.Position
	}

	// Move towards target
	return ActionMove, ai.findMoveTowardsPosition(npc.Position, target.Position)
}

// ShouldRetreat determines if the NPC should flee from combat.
func (ai *CombatAI) ShouldRetreat(npc *NPC, threats []*Character) bool {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	// Never retreat if behavior is aggressive and at medium/hard difficulty
	if npc.Behavior == "aggressive" && ai.difficulty >= AIDifficultyMedium {
		return false
	}

	// Health-based retreat threshold varies by difficulty
	healthPercent := float64(npc.HP) / float64(npc.MaxHP)

	var retreatThreshold float64
	switch ai.difficulty {
	case AIDifficultyEasy:
		retreatThreshold = 0.5 // Retreat at 50% health
	case AIDifficultyMedium:
		retreatThreshold = 0.3 // Retreat at 30% health
	case AIDifficultyHard:
		retreatThreshold = 0.2 // Retreat at 20% health
	}

	if healthPercent < retreatThreshold {
		return true
	}

	// Consider overwhelming odds (only for medium/hard difficulty)
	if ai.difficulty >= AIDifficultyMedium {
		activeThreats := 0
		for _, threat := range threats {
			if threat.HP > 0 {
				activeThreats++
			}
		}

		if activeThreats >= 3 && healthPercent < 0.7 {
			return true
		}
	}

	return false
}

// ActionType represents the type of action an NPC can take.
type ActionType int

const (
	// ActionIdle means no action (wait/pass turn)
	ActionIdle ActionType = iota
	// ActionMove means move to a position
	ActionMove
	// ActionAttack means attack a target
	ActionAttack
	// ActionRetreat means flee from combat
	ActionRetreat
)

// selectRandomTarget picks a random target from candidates (easy difficulty).
func (ai *CombatAI) selectRandomTarget(candidates []*Character) *Character {
	if len(candidates) == 0 {
		return nil
	}
	return candidates[rand.Intn(len(candidates))]
}

// selectNearestTarget picks the closest target (medium difficulty).
func (ai *CombatAI) selectNearestTarget(npc *NPC, candidates []*Character) *Character {
	if len(candidates) == 0 {
		return nil
	}

	nearest := candidates[0]
	minDistance := ai.manhattanDistance(npc.Position, nearest.Position)

	for _, candidate := range candidates[1:] {
		distance := ai.manhattanDistance(npc.Position, candidate.Position)
		if distance < minDistance {
			minDistance = distance
			nearest = candidate
		}
	}

	return nearest
}

// selectOptimalTarget picks the best tactical target (hard difficulty).
// Prioritizes: low HP targets > nearby targets > high threat targets.
func (ai *CombatAI) selectOptimalTarget(npc *NPC, candidates []*Character) *Character {
	if len(candidates) == 0 {
		return nil
	}

	type targetScore struct {
		character *Character
		score     float64
	}

	scores := make([]targetScore, 0, len(candidates))

	for _, candidate := range candidates {
		score := ai.calculateTargetScore(npc, candidate)
		scores = append(scores, targetScore{character: candidate, score: score})
	}

	// Find highest scoring target
	best := scores[0]
	for _, ts := range scores[1:] {
		if ts.score > best.score {
			best = ts
		}
	}

	return best.character
}

// calculateTargetScore assigns a tactical score to a potential target.
func (ai *CombatAI) calculateTargetScore(npc *NPC, target *Character) float64 {
	score := 0.0

	// Prioritize low HP targets (finish off wounded enemies)
	healthPercent := float64(target.HP) / float64(target.MaxHP)
	if healthPercent < 0.3 {
		score += 50.0
	} else if healthPercent < 0.5 {
		score += 25.0
	}

	// Favor closer targets
	distance := ai.manhattanDistance(npc.Position, target.Position)
	score += (10.0 - math.Min(distance, 10.0))

	// Prioritize high-damage threats (based on attributes)
	threatLevel := float64(target.Strength+target.Dexterity) / 2.0
	score += threatLevel / 4.0

	return score
}

// canReach checks if there's a path from start to end.
func (ai *CombatAI) canReach(start, end Position) bool {
	return ai.pathfinder.CanReach(start, end)
}

// manhattanDistance calculates the Manhattan distance between two positions.
func (ai *CombatAI) manhattanDistance(a, b Position) float64 {
	return math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y))
}

// findMoveTowardsPosition determines the next position to move towards a target.
func (ai *CombatAI) findMoveTowardsPosition(current, target Position) Position {
	path, found := ai.pathfinder.FindPath(current, target)
	if !found || len(path) == 0 {
		return current
	}

	// Return the first step in the path (path includes current position)
	if len(path) > 1 {
		return path[1]
	}

	return path[0]
}

// findRetreatPosition finds a safe position away from threats.
func (ai *CombatAI) findRetreatPosition(npc *NPC, threats []*Character, world *World) Position {
	// Calculate centroid of threats
	avgX, avgY := 0.0, 0.0
	activeThreats := 0

	for _, threat := range threats {
		if threat.HP > 0 {
			avgX += float64(threat.Position.X)
			avgY += float64(threat.Position.Y)
			activeThreats++
		}
	}

	if activeThreats == 0 {
		return npc.Position
	}

	avgX /= float64(activeThreats)
	avgY /= float64(activeThreats)

	// Move in opposite direction from threats
	dx := float64(npc.Position.X) - avgX
	dy := float64(npc.Position.Y) - avgY

	// Normalize and scale
	length := math.Sqrt(dx*dx + dy*dy)
	if length > 0 {
		dx = (dx / length) * 3.0
		dy = (dy / length) * 3.0
	}

	retreatPos := Position{
		X:     npc.Position.X + int(dx),
		Y:     npc.Position.Y + int(dy),
		Level: npc.Position.Level,
	}

	// Validate retreat position is walkable
	if ai.pathfinder.isWalkable(retreatPos.X, retreatPos.Y) {
		return retreatPos
	}

	// Fallback: try to move in cardinal directions away from threats
	directions := []Position{
		{X: npc.Position.X, Y: npc.Position.Y - 1, Level: npc.Position.Level}, // North
		{X: npc.Position.X + 1, Y: npc.Position.Y, Level: npc.Position.Level}, // East
		{X: npc.Position.X, Y: npc.Position.Y + 1, Level: npc.Position.Level}, // South
		{X: npc.Position.X - 1, Y: npc.Position.Y, Level: npc.Position.Level}, // West
	}

	// Find direction that increases distance from threat centroid
	bestPos := npc.Position
	maxDistance := ai.euclideanDistance(
		float64(npc.Position.X), float64(npc.Position.Y),
		avgX, avgY,
	)

	for _, dir := range directions {
		if ai.pathfinder.isWalkable(dir.X, dir.Y) {
			dist := ai.euclideanDistance(float64(dir.X), float64(dir.Y), avgX, avgY)
			if dist > maxDistance {
				maxDistance = dist
				bestPos = dir
			}
		}
	}

	return bestPos
}

// euclideanDistance calculates the straight-line distance between two points.
func (ai *CombatAI) euclideanDistance(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}
