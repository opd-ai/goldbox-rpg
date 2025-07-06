package server

import (
	"testing"

	"goldbox-rpg/pkg/game"
)

// TestCalculateNewPosition_AllDirections tests position calculations for all cardinal directions
func TestCalculateNewPosition_AllDirections(t *testing.T) {
	tests := []struct {
		name      string
		current   game.Position
		direction game.Direction
		expected  game.Position
	}{
		{
			name:      "Move North from origin",
			current:   game.Position{X: 0, Y: 0},
			direction: game.North,
			expected:  game.Position{X: 0, Y: -1},
		},
		{
			name:      "Move South from origin",
			current:   game.Position{X: 0, Y: 0},
			direction: game.South,
			expected:  game.Position{X: 0, Y: 1},
		},
		{
			name:      "Move East from origin",
			current:   game.Position{X: 0, Y: 0},
			direction: game.East,
			expected:  game.Position{X: 1, Y: 0},
		},
		{
			name:      "Move West from origin",
			current:   game.Position{X: 0, Y: 0},
			direction: game.West,
			expected:  game.Position{X: -1, Y: 0},
		},
		{
			name:      "Move North from positive coordinates",
			current:   game.Position{X: 5, Y: 3},
			direction: game.North,
			expected:  game.Position{X: 5, Y: 2},
		},
		{
			name:      "Move South from positive coordinates",
			current:   game.Position{X: 5, Y: 3},
			direction: game.South,
			expected:  game.Position{X: 5, Y: 4},
		},
		{
			name:      "Move East from positive coordinates",
			current:   game.Position{X: 5, Y: 3},
			direction: game.East,
			expected:  game.Position{X: 6, Y: 3},
		},
		{
			name:      "Move West from positive coordinates",
			current:   game.Position{X: 5, Y: 3},
			direction: game.West,
			expected:  game.Position{X: 4, Y: 3},
		},
		{
			name:      "Move North from negative coordinates",
			current:   game.Position{X: -2, Y: -4},
			direction: game.North,
			expected:  game.Position{X: -2, Y: -5},
		},
		{
			name:      "Move South from negative coordinates",
			current:   game.Position{X: -2, Y: -4},
			direction: game.South,
			expected:  game.Position{X: -2, Y: -3},
		},
		{
			name:      "Move East from negative coordinates",
			current:   game.Position{X: -2, Y: -4},
			direction: game.East,
			expected:  game.Position{X: -1, Y: -4},
		},
		{
			name:      "Move West from negative coordinates",
			current:   game.Position{X: -2, Y: -4},
			direction: game.West,
			expected:  game.Position{X: -3, Y: -4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNewPositionUnchecked(tt.current, tt.direction)

			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("calculateNewPosition() = {X: %d, Y: %d}, want {X: %d, Y: %d}",
					result.X, result.Y, tt.expected.X, tt.expected.Y)
			}
		})
	}
}

// TestCalculateNewPosition_LargeCoordinates tests movement calculations with large coordinate values
func TestCalculateNewPosition_LargeCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		current   game.Position
		direction game.Direction
		expected  game.Position
	}{
		{
			name:      "Large positive coordinates North",
			current:   game.Position{X: 1000000, Y: 999999},
			direction: game.North,
			expected:  game.Position{X: 1000000, Y: 999998},
		},
		{
			name:      "Large negative coordinates South",
			current:   game.Position{X: -1000000, Y: -999999},
			direction: game.South,
			expected:  game.Position{X: -1000000, Y: -999998},
		},
		{
			name:      "Mixed large coordinates East",
			current:   game.Position{X: -500000, Y: 750000},
			direction: game.East,
			expected:  game.Position{X: -499999, Y: 750000},
		},
		{
			name:      "Mixed large coordinates West",
			current:   game.Position{X: 500000, Y: -750000},
			direction: game.West,
			expected:  game.Position{X: 499999, Y: -750000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNewPositionUnchecked(tt.current, tt.direction)

			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("calculateNewPosition() = {X: %d, Y: %d}, want {X: %d, Y: %d}",
					result.X, result.Y, tt.expected.X, tt.expected.Y)
			}
		})
	}
}

// TestCalculateNewPosition_MultipleMovements tests sequential movement calculations
func TestCalculateNewPosition_MultipleMovements(t *testing.T) {
	start := game.Position{X: 0, Y: 0}

	// Move in a square pattern: North -> East -> South -> West
	pos1 := calculateNewPositionUnchecked(start, game.North)
	expected1 := game.Position{X: 0, Y: -1}
	if pos1 != expected1 {
		t.Errorf("First move North: got {X: %d, Y: %d}, want {X: %d, Y: %d}",
			pos1.X, pos1.Y, expected1.X, expected1.Y)
	}

	pos2 := calculateNewPositionUnchecked(pos1, game.East)
	expected2 := game.Position{X: 1, Y: -1}
	if pos2 != expected2 {
		t.Errorf("Second move East: got {X: %d, Y: %d}, want {X: %d, Y: %d}",
			pos2.X, pos2.Y, expected2.X, expected2.Y)
	}

	pos3 := calculateNewPositionUnchecked(pos2, game.South)
	expected3 := game.Position{X: 1, Y: 0}
	if pos3 != expected3 {
		t.Errorf("Third move South: got {X: %d, Y: %d}, want {X: %d, Y: %d}",
			pos3.X, pos3.Y, expected3.X, expected3.Y)
	}

	pos4 := calculateNewPositionUnchecked(pos3, game.West)
	expected4 := game.Position{X: 0, Y: 0}
	if pos4 != expected4 {
		t.Errorf("Fourth move West: got {X: %d, Y: %d}, want {X: %d, Y: %d}",
			pos4.X, pos4.Y, expected4.X, expected4.Y)
	}

	// Verify we're back at the starting position
	if pos4 != start {
		t.Errorf("After square movement pattern: got {X: %d, Y: %d}, want original {X: %d, Y: %d}",
			pos4.X, pos4.Y, start.X, start.Y)
	}
}

// TestCalculateNewPosition_NoMutation tests that original position is not modified
func TestCalculateNewPosition_NoMutation(t *testing.T) {
	original := game.Position{X: 10, Y: 20}
	originalCopy := original // Store a copy for comparison

	// Perform movement calculation
	result := calculateNewPositionUnchecked(original, game.North)

	// Verify original position wasn't modified
	if original != originalCopy {
		t.Errorf("Original position was mutated: got {X: %d, Y: %d}, want {X: %d, Y: %d}",
			original.X, original.Y, originalCopy.X, originalCopy.Y)
	}

	// Verify result is different from original
	if result == original {
		t.Errorf("Result should be different from original position")
	}

	// Verify result has correct values
	expected := game.Position{X: 10, Y: 19}
	if result != expected {
		t.Errorf("calculateNewPosition() = {X: %d, Y: %d}, want {X: %d, Y: %d}",
			result.X, result.Y, expected.X, expected.Y)
	}
}

// TestCalculateNewPosition_BoundaryValues tests movement with integer boundary values
func TestCalculateNewPosition_BoundaryValues(t *testing.T) {
	tests := []struct {
		name      string
		current   game.Position
		direction game.Direction
		expected  game.Position
	}{
		{
			name:      "Zero coordinates North",
			current:   game.Position{X: 0, Y: 0},
			direction: game.North,
			expected:  game.Position{X: 0, Y: -1},
		},
		{
			name:      "Zero coordinates South",
			current:   game.Position{X: 0, Y: 0},
			direction: game.South,
			expected:  game.Position{X: 0, Y: 1},
		},
		{
			name:      "Move from Y=-1 to Y=-2",
			current:   game.Position{X: 5, Y: -1},
			direction: game.North,
			expected:  game.Position{X: 5, Y: -2},
		},
		{
			name:      "Move from Y=1 to Y=2",
			current:   game.Position{X: 5, Y: 1},
			direction: game.South,
			expected:  game.Position{X: 5, Y: 2},
		},
		{
			name:      "Move from X=-1 to X=0",
			current:   game.Position{X: -1, Y: 5},
			direction: game.East,
			expected:  game.Position{X: 0, Y: 5},
		},
		{
			name:      "Move from X=1 to X=0",
			current:   game.Position{X: 1, Y: 5},
			direction: game.West,
			expected:  game.Position{X: 0, Y: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNewPositionUnchecked(tt.current, tt.direction)

			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("calculateNewPosition() = {X: %d, Y: %d}, want {X: %d, Y: %d}",
					result.X, result.Y, tt.expected.X, tt.expected.Y)
			}
		})
	}
}

// TestCalculateNewPosition_DirectionConstants tests using both legacy and new direction constants
func TestCalculateNewPosition_DirectionConstants(t *testing.T) {
	start := game.Position{X: 5, Y: 5}

	// Test legacy constants
	resultNorth := calculateNewPositionUnchecked(start, game.North)
	expectedNorth := game.Position{X: 5, Y: 4}
	if resultNorth != expectedNorth {
		t.Errorf("Legacy North constant: got {X: %d, Y: %d}, want {X: %d, Y: %d}",
			resultNorth.X, resultNorth.Y, expectedNorth.X, expectedNorth.Y)
	}

	// Test new constants (should produce same results)
	resultDirectionNorth := calculateNewPositionUnchecked(start, game.DirectionNorth)
	if resultDirectionNorth != expectedNorth {
		t.Errorf("DirectionNorth constant: got {X: %d, Y: %d}, want {X: %d, Y: %d}",
			resultDirectionNorth.X, resultDirectionNorth.Y, expectedNorth.X, expectedNorth.Y)
	}

	// Verify both constants are equivalent
	if game.North != game.DirectionNorth {
		t.Errorf("Legacy North constant should equal DirectionNorth")
	}
}

// TestCalculateNewPosition_BoundsEnforcement tests that movement is properly constrained to world bounds
func TestCalculateNewPosition_BoundsEnforcement(t *testing.T) {
	tests := []struct {
		name        string
		current     game.Position
		direction   game.Direction
		worldWidth  int
		worldHeight int
		expected    game.Position
	}{
		{
			name:        "Move beyond north boundary",
			current:     game.Position{X: 5, Y: 0},
			direction:   game.North,
			worldWidth:  10,
			worldHeight: 10,
			expected:    game.Position{X: 5, Y: 0}, // Clamped to minimum Y
		},
		{
			name:        "Move beyond south boundary",
			current:     game.Position{X: 5, Y: 9},
			direction:   game.South,
			worldWidth:  10,
			worldHeight: 10,
			expected:    game.Position{X: 5, Y: 9}, // Clamped to maximum Y
		},
		{
			name:        "Move beyond east boundary",
			current:     game.Position{X: 9, Y: 5},
			direction:   game.East,
			worldWidth:  10,
			worldHeight: 10,
			expected:    game.Position{X: 9, Y: 5}, // Clamped to maximum X
		},
		{
			name:        "Move beyond west boundary",
			current:     game.Position{X: 0, Y: 5},
			direction:   game.West,
			worldWidth:  10,
			worldHeight: 10,
			expected:    game.Position{X: 0, Y: 5}, // Clamped to minimum X
		},
		{
			name:        "Valid movement within bounds",
			current:     game.Position{X: 5, Y: 5},
			direction:   game.North,
			worldWidth:  10,
			worldHeight: 10,
			expected:    game.Position{X: 5, Y: 4}, // Normal movement
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNewPosition(tt.current, tt.direction, tt.worldWidth, tt.worldHeight)

			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("calculateNewPosition() = {X: %d, Y: %d}, want {X: %d, Y: %d}",
					result.X, result.Y, tt.expected.X, tt.expected.Y)
			}
		})
	}
}
