// pkg/game/default_world.go

package game

// DefaultWorld constants are defined in constants.go

// CreateDefaultWorld initializes a new world with a basic test level
func CreateDefaultWorld() *World {
	world := NewWorld()

	// Create default level
	level := &Level{
		ID:         "default_level",
		Name:       "Test Chamber",
		Width:      DefaultWorldWidth,
		Height:     DefaultWorldHeight,
		Tiles:      make([][]Tile, DefaultWorldHeight),
		Properties: make(map[string]interface{}),
	}

	// Initialize tiles
	for y := 0; y < level.Height; y++ {
		level.Tiles[y] = make([]Tile, level.Width)
		for x := 0; x < level.Width; x++ {
			// Create walls around the edges
			if x == 0 || x == level.Width-1 || y == 0 || y == level.Height-1 {
				level.Tiles[y][x] = Tile{
					Type:        TileWall,
					Walkable:    false,
					Transparent: false,
					Properties:  make(map[string]interface{}),
					Sprite:      "wall",
					Color:       RGB{128, 128, 128}, // Gray walls
					BlocksSight: true,
				}
			} else {
				// Create floor tiles in the middle
				level.Tiles[y][x] = Tile{
					Type:        TileFloor,
					Walkable:    true,
					Transparent: true,
					Properties:  make(map[string]interface{}),
					Sprite:      "floor",
					Color:       RGB{200, 200, 200}, // Light gray floor
					BlocksSight: false,
				}
			}
		}
	}

	// Add the level to the world
	world.Levels = []Level{*level}

	// Add some test objects
	// world.AddObject(testObj)

	return world
}
