package server

import (
	"goldbox-rpg/pkg/game"
)

// calculateNewPosition calculates a new position based on the current position and movement direction
//
// Parameters:
//   - current: The current Position containing X and Y coordinates
//   - direction: The Direction to move (North, South, East, or West)
//
// Returns:
//   - A new Position with updated coordinates based on the direction of movement
//
// Notes:
//   - Movement increments/decrements X or Y by 1 unit in the specified direction
//   - Does not check for boundary conditions or invalid positions
//   - Related to game.Position and game.Direction types
func calculateNewPosition(current game.Position, direction game.Direction) game.Position {
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
