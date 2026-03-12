// Package game provides combat modifier calculations for tactical combat.
package game

import (
	"math"
	"sync"

	"github.com/sirupsen/logrus"
)

// CoverType represents the degree of cover provided by terrain or objects.
type CoverType int

const (
	// CoverNone indicates no cover bonus.
	CoverNone CoverType = iota
	// CoverHalf provides +2 AC bonus (partial obstruction).
	CoverHalf
	// CoverThreeQuarters provides +5 AC bonus (significant obstruction).
	CoverThreeQuarters
	// CoverFull provides immunity to ranged attacks (complete obstruction).
	CoverFull
)

// CoverBonus returns the AC bonus for the given cover type.
func CoverBonus(cover CoverType) int {
	switch cover {
	case CoverHalf:
		return 2
	case CoverThreeQuarters:
		return 5
	case CoverFull:
		return 10 // Effectively blocks attacks
	default:
		return 0
	}
}

// CombatModifiers calculates tactical bonuses for combat encounters.
// It provides cover and flanking calculations using spatial awareness.
type CombatModifiers struct {
	mu    sync.RWMutex
	world *World
}

// NewCombatModifiers creates a new combat modifier calculator for the given world.
func NewCombatModifiers(world *World) *CombatModifiers {
	return &CombatModifiers{world: world}
}

// CalculateCover determines the cover bonus between attacker and defender positions.
// Cover is provided by walls, obstacles, or other entities between combatants.
//
// Parameters:
//   - attacker: Position of the attacking entity
//   - defender: Position of the defending entity
//
// Returns:
//   - CoverType: The degree of cover the defender has from the attacker
func (cm *CombatModifiers) CalculateCover(attacker, defender Position) CoverType {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.world == nil {
		return CoverNone
	}

	logrus.WithFields(logrus.Fields{
		"function": "CalculateCover",
		"attacker": attacker,
		"defender": defender,
	}).Debug("calculating cover")

	// Check tiles along line of sight for obstacles
	obstacleCount := cm.countObstaclesInLine(attacker, defender)

	switch {
	case obstacleCount >= 3:
		return CoverFull
	case obstacleCount >= 2:
		return CoverThreeQuarters
	case obstacleCount >= 1:
		return CoverHalf
	default:
		return CoverNone
	}
}

// countObstaclesInLine counts blocking tiles between two positions using Bresenham's line.
func (cm *CombatModifiers) countObstaclesInLine(from, to Position) int {
	obstacles := 0
	points := cm.getLinePoints(from, to)

	for _, p := range points {
		// Skip start and end positions
		if (p.X == from.X && p.Y == from.Y) || (p.X == to.X && p.Y == to.Y) {
			continue
		}

		tile := cm.getTileAt(p.X, p.Y, from.Level)
		if tile == nil {
			continue
		}

		// Walls and non-walkable tiles provide cover
		if tile.Type == TileWall || !tile.Walkable {
			obstacles++
		}
	}

	return obstacles
}

// getTileAt retrieves a tile from the world at the given coordinates and level.
func (cm *CombatModifiers) getTileAt(x, y, level int) *Tile {
	if cm.world == nil {
		return nil
	}

	cm.world.mu.RLock()
	defer cm.world.mu.RUnlock()

	// Check level bounds
	if level < 0 || level >= len(cm.world.Levels) {
		return nil
	}

	lvl := &cm.world.Levels[level]
	if x < 0 || x >= lvl.Width || y < 0 || y >= lvl.Height {
		return nil
	}

	if y >= len(lvl.Tiles) || x >= len(lvl.Tiles[y]) {
		return nil
	}

	return &lvl.Tiles[y][x]
}

// getLinePoints returns all positions along a line using Bresenham's algorithm.
func (cm *CombatModifiers) getLinePoints(from, to Position) []Position {
	var points []Position

	dx := absInt(to.X - from.X)
	dy := absInt(to.Y - from.Y)

	sx := 1
	if from.X > to.X {
		sx = -1
	}

	sy := 1
	if from.Y > to.Y {
		sy = -1
	}

	err := dx - dy
	x, y := from.X, from.Y

	for {
		points = append(points, Position{X: x, Y: y, Level: from.Level})

		if x == to.X && y == to.Y {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}

	return points
}

// CalculateFlanking determines if the attacker is flanking the defender.
// Flanking occurs when an ally is on the opposite side of the defender.
//
// Parameters:
//   - attacker: Position of the attacking entity
//   - defender: Position of the defending entity
//   - allies: Positions of the attacker's allies
//
// Returns:
//   - bool: True if flanking bonus applies
//   - int: The attack bonus from flanking (0 if not flanking, +2 otherwise)
func (cm *CombatModifiers) CalculateFlanking(attacker, defender Position, allies []Position) (bool, int) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	const flankingBonus = 2

	logrus.WithFields(logrus.Fields{
		"function":  "CalculateFlanking",
		"attacker":  attacker,
		"defender":  defender,
		"allyCount": len(allies),
	}).Debug("calculating flanking")

	// Must be adjacent to defender to flank
	if !cm.isAdjacent(attacker, defender) {
		return false, 0
	}

	// Check if any ally is on opposite side of defender
	for _, ally := range allies {
		if !cm.isAdjacent(ally, defender) {
			continue
		}

		if cm.isOpposite(attacker, defender, ally) {
			logrus.WithFields(logrus.Fields{
				"function": "CalculateFlanking",
				"ally":     ally,
			}).Debug("flanking bonus applies")
			return true, flankingBonus
		}
	}

	return false, 0
}

// isAdjacent checks if two positions are orthogonally or diagonally adjacent.
func (cm *CombatModifiers) isAdjacent(a, b Position) bool {
	dx := absInt(a.X - b.X)
	dy := absInt(a.Y - b.Y)
	return dx <= 1 && dy <= 1 && !(dx == 0 && dy == 0)
}

// isOpposite determines if attacker and ally are on opposite sides of defender.
func (cm *CombatModifiers) isOpposite(attacker, defender, ally Position) bool {
	// Calculate vectors from defender to attacker and ally
	attackerDX := attacker.X - defender.X
	attackerDY := attacker.Y - defender.Y
	allyDX := ally.X - defender.X
	allyDY := ally.Y - defender.Y

	// Opposite means vectors point in opposite directions
	// Dot product of opposite vectors is negative
	dotProduct := attackerDX*allyDX + attackerDY*allyDY

	return dotProduct < 0
}

// GetCombatModifiers calculates all applicable combat modifiers for an attack.
//
// Parameters:
//   - attackerPos: Position of the attacker
//   - defenderPos: Position of the defender
//   - allyPositions: Positions of the attacker's allies
//
// Returns:
//   - attackBonus: Total bonus to attack roll
//   - defenseBonus: Total bonus to defender's AC
func (cm *CombatModifiers) GetCombatModifiers(attackerPos, defenderPos Position, allyPositions []Position) (attackBonus, defenseBonus int) {
	// Calculate cover bonus for defender
	cover := cm.CalculateCover(attackerPos, defenderPos)
	defenseBonus = CoverBonus(cover)

	// Calculate flanking bonus for attacker
	isFlanking, flankBonus := cm.CalculateFlanking(attackerPos, defenderPos, allyPositions)
	if isFlanking {
		attackBonus = flankBonus
	}

	logrus.WithFields(logrus.Fields{
		"function":     "GetCombatModifiers",
		"attackBonus":  attackBonus,
		"defenseBonus": defenseBonus,
		"cover":        cover,
		"flanking":     isFlanking,
	}).Debug("combat modifiers calculated")

	return attackBonus, defenseBonus
}

// HighGroundBonus returns attack bonus for having higher elevation.
// Returns 0 until MapTile gains an Elevation field to support terrain height differences.
// Planned bonus: +2 for elevation advantage, -2 for disadvantage.
func (cm *CombatModifiers) HighGroundBonus(attackerPos, defenderPos Position) int {
	// Elevation not yet supported in terrain system - MapTile needs Elevation field
	return 0
}

// GetAdjacentPositions returns all positions adjacent to the given position.
func (cm *CombatModifiers) GetAdjacentPositions(pos Position) []Position {
	var adjacent []Position
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			adjacent = append(adjacent, Position{
				X:     pos.X + dx,
				Y:     pos.Y + dy,
				Level: pos.Level,
			})
		}
	}
	return adjacent
}

// absInt returns the absolute value of an integer.
// Named absInt to avoid collision with other packages.
func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// DistanceBetween calculates the Euclidean distance between two positions.
func DistanceBetween(a, b Position) float64 {
	dx := float64(b.X - a.X)
	dy := float64(b.Y - a.Y)
	return math.Sqrt(dx*dx + dy*dy)
}
