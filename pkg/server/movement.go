package server

import (
	"goldbox-rpg/pkg/game"
)

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
