package server

import (
	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// ADDED: calculateNewPosition computes a new position from current position and movement direction.
// It enforces world boundary constraints to prevent invalid coordinates.
//
// Movement rules:
// - Coordinates are clamped to world bounds [0, worldWidth) x [0, worldHeight)
// - Invalid movements (out of bounds) are ignored, returning current position
// - Direction mapping: North=+Y, South=-Y, East=+X, West=-X
//
// Parameters:
//   - current: Current position with X, Y coordinates
//   - direction: Movement direction (North, South, East, West)
//   - worldWidth: Maximum X coordinate (exclusive upper bound)
//   - worldHeight: Maximum Y coordinate (exclusive upper bound)
//
// Returns:
//   - game.Position: New position with boundary-constrained coordinates
//
// Boundary enforcement prevents characters from moving outside the game world.
func calculateNewPosition(current game.Position, direction game.Direction, worldWidth, worldHeight int) game.Position {
	logrus.WithFields(logrus.Fields{
		"function":    "calculateNewPosition",
		"current":     current,
		"direction":   direction,
		"worldWidth":  worldWidth,
		"worldHeight": worldHeight,
	}).Debug("entering calculateNewPosition")

	newPos := current

	logrus.WithFields(logrus.Fields{
		"function": "calculateNewPosition",
	}).Info("calculating new position with bounds checking")

	switch direction {
	case game.North:
		if newPos.Y+1 < worldHeight {
			newPos.Y++
		}
	case game.South:
		if newPos.Y-1 >= 0 {
			newPos.Y--
		}
	case game.East:
		if newPos.X+1 < worldWidth {
			newPos.X++
		}
	case game.West:
		if newPos.X-1 >= 0 {
			newPos.X--
		}
	}

	logrus.WithFields(logrus.Fields{
		"function": "calculateNewPosition",
		"newPos":   newPos,
	}).Debug("exiting calculateNewPosition")

	return newPos
}

// calculateNewPositionUnchecked calculates a new position without bounds checking.
// This function is preserved for testing purposes and backward compatibility.
// Production code should use calculateNewPosition with proper bounds.
func calculateNewPositionUnchecked(current game.Position, direction game.Direction) game.Position {
	newPos := current

	switch direction {
	case game.North:
		newPos.Y++
	case game.South:
		newPos.Y--
	case game.East:
		newPos.X++
	case game.West:
		newPos.X--
	}

	return newPos
}
