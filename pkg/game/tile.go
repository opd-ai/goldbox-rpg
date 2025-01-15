package game

// TileType represents different types of map tiles
type TileType int

const (
	TileFloor TileType = iota
	TileWall
	TileDoor
	TileWater
	TileLava
	TilePit
	TileStairs
)

// Tile represents a single map cell
// Contains all properties that define a tile's behavior and appearance
type Tile struct {
	Type        TileType               `yaml:"tile_type"`        // Base type of the tile
	Walkable    bool                   `yaml:"tile_walkable"`    // Whether entities can move through
	Transparent bool                   `yaml:"tile_transparent"` // Whether light passes through
	Properties  map[string]interface{} `yaml:"tile_properties"`  // Custom property map

	// Visual properties
	Sprite string `yaml:"tile_sprite"` // Sprite/texture identifier
	Color  RGB    `yaml:"tile_color"`  // Base color tint

	// Special properties
	BlocksSight bool   `yaml:"tile_blocks_sight"` // Whether blocks line of sight
	Dangerous   bool   `yaml:"tile_dangerous"`    // Whether causes damage
	DamageType  string `yaml:"tile_damage_type"`  // Type of damage dealt
	Damage      int    `yaml:"tile_damage"`       // Amount of damage per turn
}

// RGB represents a color in RGB format
// Each component ranges from 0-255
type RGB struct {
	R uint8 `yaml:"color_red"`   // Red component
	G uint8 `yaml:"color_green"` // Green component
	B uint8 `yaml:"color_blue"`  // Blue component
}

// Common tile factory functions
func NewFloorTile() Tile {
	return Tile{
		Type:        TileFloor,
		Walkable:    true,
		Transparent: true,
		Properties:  make(map[string]interface{}),
		Color:       RGB{200, 200, 200},
	}
}

func NewWallTile() Tile {
	return Tile{
		Type:        TileWall,
		Walkable:    false,
		Transparent: false,
		BlocksSight: true,
		Properties:  make(map[string]interface{}),
		Color:       RGB{128, 128, 128},
	}
}
