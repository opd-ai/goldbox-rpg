package game

import (
	"testing"
)

// TestDefaultWorldConstants tests the default world size constants
func TestDefaultWorldConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{
			name:     "DefaultWorldWidth should be 10",
			constant: DefaultWorldWidth,
			expected: 10,
		},
		{
			name:     "DefaultWorldHeight should be 10",
			constant: DefaultWorldHeight,
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("constant = %d, want %d", tt.constant, tt.expected)
			}
		})
	}
}

// TestCreateDefaultWorld_BasicStructure tests the basic structure of the created world
func TestCreateDefaultWorld_BasicStructure(t *testing.T) {
	world := CreateDefaultWorld()

	if world == nil {
		t.Fatal("CreateDefaultWorld() returned nil")
	}

	if len(world.Levels) != 1 {
		t.Errorf("world.Levels length = %d, want 1", len(world.Levels))
	}

	level := world.Levels[0]

	if level.ID != "default_level" {
		t.Errorf("level.ID = %q, want %q", level.ID, "default_level")
	}

	if level.Name != "Test Chamber" {
		t.Errorf("level.Name = %q, want %q", level.Name, "Test Chamber")
	}

	if level.Width != DefaultWorldWidth {
		t.Errorf("level.Width = %d, want %d", level.Width, DefaultWorldWidth)
	}

	if level.Height != DefaultWorldHeight {
		t.Errorf("level.Height = %d, want %d", level.Height, DefaultWorldHeight)
	}

	if level.Properties == nil {
		t.Error("level.Properties is nil, want initialized map")
	}

	if len(level.Properties) != 0 {
		t.Errorf("level.Properties length = %d, want 0", len(level.Properties))
	}
}

// TestCreateDefaultWorld_TileArrayDimensions tests the tile array dimensions
func TestCreateDefaultWorld_TileArrayDimensions(t *testing.T) {
	world := CreateDefaultWorld()
	level := world.Levels[0]

	if len(level.Tiles) != DefaultWorldHeight {
		t.Errorf("level.Tiles height = %d, want %d", len(level.Tiles), DefaultWorldHeight)
	}

	for y := 0; y < DefaultWorldHeight; y++ {
		if len(level.Tiles[y]) != DefaultWorldWidth {
			t.Errorf("level.Tiles[%d] width = %d, want %d", y, len(level.Tiles[y]), DefaultWorldWidth)
		}
	}
}

// TestCreateDefaultWorld_WallTiles tests wall tiles around the edges
func TestCreateDefaultWorld_WallTiles(t *testing.T) {
	world := CreateDefaultWorld()
	level := world.Levels[0]

	tests := []struct {
		name string
		x, y int
	}{
		// Top edge
		{"top-left corner", 0, 0},
		{"top-middle", DefaultWorldWidth / 2, 0},
		{"top-right corner", DefaultWorldWidth - 1, 0},
		// Bottom edge
		{"bottom-left corner", 0, DefaultWorldHeight - 1},
		{"bottom-middle", DefaultWorldWidth / 2, DefaultWorldHeight - 1},
		{"bottom-right corner", DefaultWorldWidth - 1, DefaultWorldHeight - 1},
		// Left edge
		{"left-middle", 0, DefaultWorldHeight / 2},
		// Right edge
		{"right-middle", DefaultWorldWidth - 1, DefaultWorldHeight / 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := level.Tiles[tt.y][tt.x]

			if tile.Type != TileWall {
				t.Errorf("tile at (%d,%d).Type = %v, want %v", tt.x, tt.y, tile.Type, TileWall)
			}

			if tile.Walkable != false {
				t.Errorf("tile at (%d,%d).Walkable = %v, want false", tt.x, tt.y, tile.Walkable)
			}

			if tile.Transparent != false {
				t.Errorf("tile at (%d,%d).Transparent = %v, want false", tt.x, tt.y, tile.Transparent)
			}

			if tile.Sprite != "wall" {
				t.Errorf("tile at (%d,%d).Sprite = %q, want %q", tt.x, tt.y, tile.Sprite, "wall")
			}

			expectedColor := RGB{128, 128, 128}
			if tile.Color != expectedColor {
				t.Errorf("tile at (%d,%d).Color = %v, want %v", tt.x, tt.y, tile.Color, expectedColor)
			}

			if tile.BlocksSight != true {
				t.Errorf("tile at (%d,%d).BlocksSight = %v, want true", tt.x, tt.y, tile.BlocksSight)
			}

			if tile.Properties == nil {
				t.Errorf("tile at (%d,%d).Properties is nil, want initialized map", tt.x, tt.y)
			}
		})
	}
}

// TestCreateDefaultWorld_FloorTiles tests floor tiles in the interior
func TestCreateDefaultWorld_FloorTiles(t *testing.T) {
	world := CreateDefaultWorld()
	level := world.Levels[0]

	tests := []struct {
		name string
		x, y int
	}{
		{"center tile", DefaultWorldWidth / 2, DefaultWorldHeight / 2},
		{"near top-left", 1, 1},
		{"near top-right", DefaultWorldWidth - 2, 1},
		{"near bottom-left", 1, DefaultWorldHeight - 2},
		{"near bottom-right", DefaultWorldWidth - 2, DefaultWorldHeight - 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := level.Tiles[tt.y][tt.x]

			if tile.Type != TileFloor {
				t.Errorf("tile at (%d,%d).Type = %v, want %v", tt.x, tt.y, tile.Type, TileFloor)
			}

			if tile.Walkable != true {
				t.Errorf("tile at (%d,%d).Walkable = %v, want true", tt.x, tt.y, tile.Walkable)
			}

			if tile.Transparent != true {
				t.Errorf("tile at (%d,%d).Transparent = %v, want true", tt.x, tt.y, tile.Transparent)
			}

			if tile.Sprite != "floor" {
				t.Errorf("tile at (%d,%d).Sprite = %q, want %q", tt.x, tt.y, tile.Sprite, "floor")
			}

			expectedColor := RGB{200, 200, 200}
			if tile.Color != expectedColor {
				t.Errorf("tile at (%d,%d).Color = %v, want %v", tt.x, tt.y, tile.Color, expectedColor)
			}

			if tile.BlocksSight != false {
				t.Errorf("tile at (%d,%d).BlocksSight = %v, want false", tt.x, tt.y, tile.BlocksSight)
			}

			if tile.Properties == nil {
				t.Errorf("tile at (%d,%d).Properties is nil, want initialized map", tt.x, tt.y)
			}
		})
	}
}

// TestCreateDefaultWorld_TilePropertiesInitialization tests that all tiles have initialized properties
func TestCreateDefaultWorld_TilePropertiesInitialization(t *testing.T) {
	world := CreateDefaultWorld()
	level := world.Levels[0]

	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			tile := level.Tiles[y][x]

			if tile.Properties == nil {
				t.Errorf("tile at (%d,%d).Properties is nil, want initialized map", x, y)
				continue
			}

			// Properties should be empty but not nil
			if len(tile.Properties) != 0 {
				t.Errorf("tile at (%d,%d).Properties length = %d, want 0", x, y, len(tile.Properties))
			}
		}
	}
}

// TestCreateDefaultWorld_BoundaryConditions tests the boundary between walls and floors
func TestCreateDefaultWorld_BoundaryConditions(t *testing.T) {
	world := CreateDefaultWorld()
	level := world.Levels[0]

	// Test that tiles just inside the border are floors
	innerBoundaryTests := []struct {
		name string
		x, y int
	}{
		{"inner top-left", 1, 1},
		{"inner top-right", DefaultWorldWidth - 2, 1},
		{"inner bottom-left", 1, DefaultWorldHeight - 2},
		{"inner bottom-right", DefaultWorldWidth - 2, DefaultWorldHeight - 2},
	}

	for _, tt := range innerBoundaryTests {
		t.Run("floor_"+tt.name, func(t *testing.T) {
			tile := level.Tiles[tt.y][tt.x]
			if tile.Type != TileFloor {
				t.Errorf("inner boundary tile at (%d,%d) should be floor, got %v", tt.x, tt.y, tile.Type)
			}
		})
	}

	// Test that all edge tiles are walls
	// Top and bottom edges
	for x := 0; x < DefaultWorldWidth; x++ {
		// Top edge
		if level.Tiles[0][x].Type != TileWall {
			t.Errorf("top edge tile at (%d,0) should be wall, got %v", x, level.Tiles[0][x].Type)
		}
		// Bottom edge
		bottomY := DefaultWorldHeight - 1
		if level.Tiles[bottomY][x].Type != TileWall {
			t.Errorf("bottom edge tile at (%d,%d) should be wall, got %v", x, bottomY, level.Tiles[bottomY][x].Type)
		}
	}

	// Left and right edges
	for y := 0; y < DefaultWorldHeight; y++ {
		// Left edge
		if level.Tiles[y][0].Type != TileWall {
			t.Errorf("left edge tile at (0,%d) should be wall, got %v", y, level.Tiles[y][0].Type)
		}
		// Right edge
		rightX := DefaultWorldWidth - 1
		if level.Tiles[y][rightX].Type != TileWall {
			t.Errorf("right edge tile at (%d,%d) should be wall, got %v", rightX, y, level.Tiles[y][rightX].Type)
		}
	}
}

// TestCreateDefaultWorld_MultipleCallsIndependence tests that multiple calls create independent worlds
func TestCreateDefaultWorld_MultipleCallsIndependence(t *testing.T) {
	world1 := CreateDefaultWorld()
	world2 := CreateDefaultWorld()

	// Worlds should be independent objects
	if world1 == world2 {
		t.Error("multiple calls to CreateDefaultWorld() returned the same object")
	}

	// Levels should be independent
	if &world1.Levels[0] == &world2.Levels[0] {
		t.Error("levels from different worlds share the same memory address")
	}

	// Modify a tile in world1 and ensure world2 is unaffected
	world1.Levels[0].Tiles[1][1].Type = TileWater
	if world2.Levels[0].Tiles[1][1].Type != TileFloor {
		t.Error("modification to world1 affected world2")
	}

	// Modify properties in world1 and ensure world2 is unaffected
	world1.Levels[0].Properties["test"] = "value"
	if len(world2.Levels[0].Properties) != 0 {
		t.Error("modification to world1 properties affected world2")
	}
}

// TestCreateDefaultWorld_ValidWorldType tests that NewWorld() creates a valid World type
func TestCreateDefaultWorld_ValidWorldType(t *testing.T) {
	world := CreateDefaultWorld()

	// Test that the world has expected World type characteristics
	// This tests integration with NewWorld()
	if world.Objects == nil {
		t.Error("world.Objects should be initialized by NewWorld()")
	}

	if world.Players == nil {
		t.Error("world.Players should be initialized by NewWorld()")
	}

	if world.NPCs == nil {
		t.Error("world.NPCs should be initialized by NewWorld()")
	}

	if world.SpatialGrid == nil {
		t.Error("world.SpatialGrid should be initialized by NewWorld()")
	}
}
