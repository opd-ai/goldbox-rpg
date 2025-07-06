package game

// TileType represents the type of a tile in the game world.
// It is implemented as an integer enum to efficiently store and compare different tile types.
type TileType int

// TileType constants are defined in constants.go
// Each constant is assigned a unique integer value through iota.

// Tile represents a single cell in the game map. It encapsulates all properties
// that define a tile's behavior, appearance, and interaction capabilities within the game world.
//
// Related types:
// - TileType: Defines the base classification of the tile
// - RGB: Defines the color properties
//
// Fields:
// - Type: Base classification of the tile (floor, wall, etc.)
// - Walkable: Determines if entities can traverse this tile
// - Transparent: Controls if light can pass through the tile
// - Properties: Custom key-value store for additional tile attributes
// - Sprite: Identifier for the tile's visual representation
// - Color: Base RGB color tint for rendering
// - BlocksSight: Specifically controls line of sight behavior
// - Dangerous: Indicates if the tile can cause damage
// - DamageType: Classification of damage (e.g., "fire", "poison")
// - Damage: Integer amount of damage dealt per turn if dangerous
//
// Note: Properties map allows for dynamic extension of tile attributes
// without modifying the core structure.
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
// RGB represents a color in the RGB color space with 8-bit components.
// Each component (R,G,B) ranges from 0-255.
//
// The struct is designed to be YAML serializable with custom field tags.
//
// This is used throughout the game engine for defining colors of tiles,
// sprites and other visual elements.
//
// Related types:
//   - Tile - Uses RGB for foreground/background colors
type RGB struct {
	R uint8 `yaml:"color_red"`   // Red component
	G uint8 `yaml:"color_green"` // Green component
	B uint8 `yaml:"color_blue"`  // Blue component
}

// Common tile factory functions
// NewFloorTile creates and returns a new floor tile with default properties.
// The floor tile is walkable and transparent with a light gray color (RGB: 200,200,200).
// Returns a Tile struct configured as a basic floor tile with:
// - Type: TileFloor
// - Walkable: true
// - Transparent: true
// - Empty properties map
// - Light gray color
//
// Related types:
// - Tile struct
// - TileFloor constant
func NewFloorTile() Tile {
	return Tile{
		Type:        TileFloor,
		Walkable:    true,
		Transparent: true,
		Properties:  make(map[string]interface{}),
		Sprite:      "",
		Color:       RGB{200, 200, 200},
		BlocksSight: false,
		Dangerous:   false,
		DamageType:  "",
		Damage:      0,
	}
}

// NewWallTile creates and returns a new wall tile with default properties.
// It initializes an impassable, opaque wall with gray coloring that blocks line of sight.
//
// Returns:
//   - Tile: A new wall tile instance with the following default properties:
//   - Type: TileWall
//   - Walkable: false (cannot be walked through)
//   - Transparent: false (blocks vision)
//   - Properties: empty map for custom properties
//   - Sprite: empty string (no sprite assigned)
//   - Color: gray RGB(128,128,128)
//   - BlocksSight: true (blocks line of sight)
//   - Dangerous: false (does not cause damage)
//   - DamageType: empty string (no damage type)
//   - Damage: 0 (no damage value)
//
// Related types:
//   - Tile
//   - RGB
//   - TileWall (constant)
func NewWallTile() Tile {
	return Tile{
		Type:        TileWall,
		Walkable:    false,
		Transparent: false,
		Properties:  make(map[string]interface{}),
		Sprite:      "",
		Color:       RGB{128, 128, 128},
		BlocksSight: true,
		Dangerous:   false,
		DamageType:  "",
		Damage:      0,
	}
}
