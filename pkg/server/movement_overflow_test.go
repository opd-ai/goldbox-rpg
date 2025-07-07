package server

import (
	"goldbox-rpg/pkg/game"
	"math"
	"testing"
)

// TestCalculateNewPositionOverflowScenarios tests potential overflow scenarios
func TestCalculateNewPositionOverflowScenarios(t *testing.T) {
	tests := []struct {
		name        string
		current     game.Position
		direction   game.Direction
		worldWidth  int
		worldHeight int
		expected    game.Position
	}{
		{
			name:        "Large world coordinates stay within bounds",
			current:     game.Position{X: 1000000, Y: 1000000},
			direction:   game.DirectionNorth,
			worldWidth:  2000000,
			worldHeight: 2000000,
			expected:    game.Position{X: 1000000, Y: 999999}, // Normal movement
		},
		{
			name:        "Near max int coordinates are handled safely",
			current:     game.Position{X: math.MaxInt32 - 1000, Y: math.MaxInt32 - 1000},
			direction:   game.DirectionEast,
			worldWidth:  math.MaxInt32,
			worldHeight: math.MaxInt32,
			expected:    game.Position{X: math.MaxInt32 - 999, Y: math.MaxInt32 - 1000}, // Normal movement
		},
		{
			name:        "Movement at world edge stays within bounds",
			current:     game.Position{X: 99, Y: 99},
			direction:   game.DirectionEast,
			worldWidth:  100,
			worldHeight: 100,
			expected:    game.Position{X: 99, Y: 99}, // Stays at edge
		},
		{
			name:        "Movement at coordinate 0 boundary",
			current:     game.Position{X: 0, Y: 0},
			direction:   game.DirectionWest,
			worldWidth:  100,
			worldHeight: 100,
			expected:    game.Position{X: 0, Y: 0}, // Stays at boundary
		},
		{
			name:        "Movement at coordinate 0 boundary North",
			current:     game.Position{X: 0, Y: 0},
			direction:   game.DirectionNorth,
			worldWidth:  100,
			worldHeight: 100,
			expected:    game.Position{X: 0, Y: 0}, // Stays at boundary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNewPosition(tt.current, tt.direction, tt.worldWidth, tt.worldHeight)
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("calculateNewPosition() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateNewPositionBoundsSafety tests that bounds checking prevents invalid coordinates
func TestCalculateNewPositionBoundsSafety(t *testing.T) {
	// Test with very large world dimensions
	largeWorld := 1000000
	pos := game.Position{X: largeWorld / 2, Y: largeWorld / 2}

	// Test all directions
	directions := []game.Direction{
		game.DirectionNorth,
		game.DirectionSouth,
		game.DirectionEast,
		game.DirectionWest,
	}

	for _, dir := range directions {
		result := calculateNewPosition(pos, dir, largeWorld, largeWorld)

		// Verify result is always within bounds
		if result.X < 0 || result.X >= largeWorld {
			t.Errorf("X coordinate %d out of bounds for world width %d", result.X, largeWorld)
		}
		if result.Y < 0 || result.Y >= largeWorld {
			t.Errorf("Y coordinate %d out of bounds for world height %d", result.Y, largeWorld)
		}
	}
}
