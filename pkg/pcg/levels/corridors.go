package levels

import (
	"fmt"
	"math"
	"math/rand"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// CorridorPlanner handles corridor generation between rooms
type CorridorPlanner struct {
	style pcg.CorridorStyle
	rng   *rand.Rand
}

// NewCorridorPlanner creates a new corridor planner with specified style
func NewCorridorPlanner(style pcg.CorridorStyle, rng *rand.Rand) *CorridorPlanner {
	return &CorridorPlanner{
		style: style,
		rng:   rng,
	}
}

// CreateCorridor generates a corridor between two points
func (cp *CorridorPlanner) CreateCorridor(id string, start, end game.Position, theme pcg.LevelTheme) (*pcg.Corridor, error) {
	corridor := &pcg.Corridor{
		ID:       id,
		Start:    start,
		End:      end,
		Width:    cp.determineCorridorWidth(),
		Style:    cp.style,
		Features: []pcg.CorridorFeature{},
	}

	// Generate path based on style
	var err error
	corridor.Path, err = cp.generatePath(start, end, theme)
	if err != nil {
		return nil, fmt.Errorf("failed to generate corridor path: %w", err)
	}

	// Add corridor features
	corridor.Features = cp.generateCorridorFeatures(corridor.Path, theme)

	return corridor, nil
}

// determineCorridorWidth calculates appropriate corridor width
func (cp *CorridorPlanner) determineCorridorWidth() int {
	switch cp.style {
	case pcg.CorridorMinimal:
		return 1
	case pcg.CorridorStraight:
		return 1 + cp.rng.Intn(2)
	case pcg.CorridorWindy:
		return 1 + cp.rng.Intn(2)
	case pcg.CorridorMaze:
		return 1
	case pcg.CorridorOrganic:
		return 2 + cp.rng.Intn(2)
	default:
		return 1
	}
}

// generatePath creates the corridor path based on style
func (cp *CorridorPlanner) generatePath(start, end game.Position, theme pcg.LevelTheme) ([]game.Position, error) {
	switch cp.style {
	case pcg.CorridorStraight:
		return cp.generateStraightPath(start, end)
	case pcg.CorridorWindy:
		return cp.generateWindyPath(start, end)
	case pcg.CorridorMaze:
		return cp.generateMazePath(start, end)
	case pcg.CorridorOrganic:
		return cp.generateOrganicPath(start, end)
	case pcg.CorridorMinimal:
		return cp.generateMinimalPath(start, end)
	default:
		return cp.generateStraightPath(start, end)
	}
}

// generateStraightPath creates direct L-shaped corridors
func (cp *CorridorPlanner) generateStraightPath(start, end game.Position) ([]game.Position, error) {
	var path []game.Position

	// Determine if we go horizontal first or vertical first
	horizontalFirst := cp.rng.Float64() < 0.5

	current := start
	path = append(path, current)

	if horizontalFirst {
		// Move horizontally first
		for current.X != end.X {
			if current.X < end.X {
				current.X++
			} else {
				current.X--
			}
			path = append(path, current)
		}

		// Then move vertically
		for current.Y != end.Y {
			if current.Y < end.Y {
				current.Y++
			} else {
				current.Y--
			}
			path = append(path, current)
		}
	} else {
		// Move vertically first
		for current.Y != end.Y {
			if current.Y < end.Y {
				current.Y++
			} else {
				current.Y--
			}
			path = append(path, current)
		}

		// Then move horizontally
		for current.X != end.X {
			if current.X < end.X {
				current.X++
			} else {
				current.X--
			}
			path = append(path, current)
		}
	}

	return path, nil
}

// generateWindyPath creates corridors with random turns
func (cp *CorridorPlanner) generateWindyPath(start, end game.Position) ([]game.Position, error) {
	var path []game.Position
	current := start
	path = append(path, current)

	// Add some randomness to the path
	for current.X != end.X || current.Y != end.Y {
		// Decide direction with bias toward target
		dx := end.X - current.X
		dy := end.Y - current.Y

		// Choose direction with weighted probability
		directions := []game.Position{}
		weights := []float64{}

		// Right
		if dx > 0 {
			directions = append(directions, game.Position{X: 1, Y: 0})
			weights = append(weights, 0.6)
		}
		// Left
		if dx < 0 {
			directions = append(directions, game.Position{X: -1, Y: 0})
			weights = append(weights, 0.6)
		}
		// Down
		if dy > 0 {
			directions = append(directions, game.Position{X: 0, Y: 1})
			weights = append(weights, 0.6)
		}
		// Up
		if dy < 0 {
			directions = append(directions, game.Position{X: 0, Y: -1})
			weights = append(weights, 0.6)
		}

		// Add some random directions
		if len(directions) > 0 {
			// Pick weighted random direction
			dir := cp.weightedRandomDirection(directions, weights)
			current.X += dir.X
			current.Y += dir.Y
			path = append(path, current)
		} else {
			break
		}

		// Add occasional random turns
		if cp.rng.Float64() < 0.2 {
			// Random side step
			if cp.rng.Float64() < 0.5 && dx != 0 {
				current.Y += cp.rng.Intn(3) - 1
			} else if dy != 0 {
				current.X += cp.rng.Intn(3) - 1
			}
			path = append(path, current)
		}
	}

	return path, nil
}

// generateMazePath creates maze-like corridor paths
func (cp *CorridorPlanner) generateMazePath(start, end game.Position) ([]game.Position, error) {
	// For now, create a more complex path with multiple turns
	var path []game.Position
	current := start
	path = append(path, current)

	// Create intermediate waypoints
	waypoints := cp.generateWaypoints(start, end, 2+cp.rng.Intn(3))

	for _, waypoint := range waypoints {
		// Generate straight path to each waypoint
		for current.X != waypoint.X || current.Y != waypoint.Y {
			if current.X != waypoint.X {
				if current.X < waypoint.X {
					current.X++
				} else {
					current.X--
				}
			} else if current.Y != waypoint.Y {
				if current.Y < waypoint.Y {
					current.Y++
				} else {
					current.Y--
				}
			}
			path = append(path, current)
		}
	}

	return path, nil
}

// generateOrganicPath creates natural, flowing corridors
func (cp *CorridorPlanner) generateOrganicPath(start, end game.Position) ([]game.Position, error) {
	var path []game.Position
	current := start
	path = append(path, current)

	distance := math.Sqrt(float64((end.X-start.X)*(end.X-start.X) + (end.Y-start.Y)*(end.Y-start.Y)))
	steps := int(distance * 1.5) // Make path longer for organic feel

	for i := 0; i < steps && (current.X != end.X || current.Y != end.Y); i++ {
		// Use sine wave for organic movement
		progress := float64(i) / float64(steps)

		// Calculate ideal position along direct line
		idealX := start.X + int(float64(end.X-start.X)*progress)
		idealY := start.Y + int(float64(end.Y-start.Y)*progress)

		// Add organic deviation
		deviation := math.Sin(progress*math.Pi*4) * 2 // Sine wave deviation

		// Move toward ideal position with organic curves
		if current.X < idealX {
			current.X++
		} else if current.X > idealX {
			current.X--
		}

		if current.Y < idealY {
			current.Y++
		} else if current.Y > idealY {
			current.Y--
		}

		// Apply organic deviation
		if cp.rng.Float64() < 0.3 {
			current.X += int(deviation)
		}

		path = append(path, current)
	}

	// Ensure we reach the end
	for current.X != end.X || current.Y != end.Y {
		if current.X < end.X {
			current.X++
		} else if current.X > end.X {
			current.X--
		} else if current.Y < end.Y {
			current.Y++
		} else if current.Y > end.Y {
			current.Y--
		}
		path = append(path, current)
	}

	return path, nil
}

// generateMinimalPath creates the shortest possible path
func (cp *CorridorPlanner) generateMinimalPath(start, end game.Position) ([]game.Position, error) {
	var path []game.Position
	current := start
	path = append(path, current)

	// Move diagonally when possible, then orthogonally
	for current.X != end.X || current.Y != end.Y {
		moved := false

		// Try diagonal movement first
		if current.X != end.X && current.Y != end.Y {
			if current.X < end.X {
				current.X++
			} else {
				current.X--
			}
			if current.Y < end.Y {
				current.Y++
			} else {
				current.Y--
			}
			moved = true
		} else {
			// Orthogonal movement
			if current.X != end.X {
				if current.X < end.X {
					current.X++
				} else {
					current.X--
				}
				moved = true
			} else if current.Y != end.Y {
				if current.Y < end.Y {
					current.Y++
				} else {
					current.Y--
				}
				moved = true
			}
		}

		if moved {
			path = append(path, current)
		} else {
			break
		}
	}

	return path, nil
}

// generateWaypoints creates intermediate points for complex paths
func (cp *CorridorPlanner) generateWaypoints(start, end game.Position, count int) []game.Position {
	var waypoints []game.Position

	for i := 0; i < count; i++ {
		progress := float64(i+1) / float64(count+1)

		x := start.X + int(float64(end.X-start.X)*progress)
		y := start.Y + int(float64(end.Y-start.Y)*progress)

		// Add some randomness
		x += cp.rng.Intn(6) - 3
		y += cp.rng.Intn(6) - 3

		waypoints = append(waypoints, game.Position{X: x, Y: y})
	}

	// Always end at the target
	waypoints = append(waypoints, end)

	return waypoints
}

// weightedRandomDirection selects a direction based on weights
func (cp *CorridorPlanner) weightedRandomDirection(directions []game.Position, weights []float64) game.Position {
	if len(directions) == 0 {
		return game.Position{X: 0, Y: 0}
	}

	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	if totalWeight == 0 {
		return directions[cp.rng.Intn(len(directions))]
	}

	target := cp.rng.Float64() * totalWeight
	current := 0.0

	for i, weight := range weights {
		current += weight
		if current >= target {
			return directions[i]
		}
	}

	return directions[len(directions)-1]
}

// generateCorridorFeatures adds special features to corridors
func (cp *CorridorPlanner) generateCorridorFeatures(path []game.Position, theme pcg.LevelTheme) []pcg.CorridorFeature {
	var features []pcg.CorridorFeature

	// Add features spaced throughout the corridor
	featureSpacing := 8 + cp.rng.Intn(5)

	for i := featureSpacing; i < len(path); i += featureSpacing {
		if cp.rng.Float64() < 0.4 { // 40% chance for feature
			featureType := cp.selectCorridorFeatureType(theme)

			feature := pcg.CorridorFeature{
				Type:     featureType,
				Position: path[i],
				Properties: map[string]interface{}{
					"theme": theme,
				},
			}

			features = append(features, feature)
		}
	}

	return features
}

// selectCorridorFeatureType chooses appropriate corridor features based on theme
func (cp *CorridorPlanner) selectCorridorFeatureType(theme pcg.LevelTheme) string {
	features := map[pcg.LevelTheme][]string{
		pcg.ThemeClassic:    {"torch", "banner", "statue"},
		pcg.ThemeHorror:     {"blood_stain", "scratch_marks", "bone_pile"},
		pcg.ThemeNatural:    {"moss", "root", "mushroom"},
		pcg.ThemeMechanical: {"gear", "pipe", "console"},
		pcg.ThemeMagical:    {"rune", "crystal", "floating_orb"},
		pcg.ThemeUndead:     {"coffin", "skeleton", "tomb"},
		pcg.ThemeElemental:  {"flame", "ice_crystal", "water_pool"},
	}

	themeFeatures, exists := features[theme]
	if !exists {
		themeFeatures = features[pcg.ThemeClassic]
	}

	return themeFeatures[cp.rng.Intn(len(themeFeatures))]
}
