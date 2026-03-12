package pcg

import (
	"goldbox-rpg/pkg/game"
)

// Rectangle represents a rectangular area for spatial operations
type Rectangle struct {
	X, Y          int // Top-left corner coordinates
	Width, Height int // Dimensions
}

// Contains checks if a position is within the rectangle
func (r Rectangle) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

// Intersects checks if this rectangle intersects with another
func (r Rectangle) Intersects(other Rectangle) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}

// RoomLayout represents the layout of a generated room
type RoomLayout struct {
	ID         string                 `yaml:"id"`         // Unique room identifier
	Type       RoomType               `yaml:"type"`       // Room type classification
	Bounds     Rectangle              `yaml:"bounds"`     // Room dimensions and position
	Tiles      [][]game.Tile          `yaml:"tiles"`      // Room tile data
	Doors      []game.Position        `yaml:"doors"`      // Door/entrance positions
	Features   []RoomFeature          `yaml:"features"`   // Special room features
	Difficulty int                    `yaml:"difficulty"` // Challenge rating
	Properties map[string]interface{} `yaml:"properties"` // Additional room data
	Connected  []string               `yaml:"connected"`  // IDs of connected rooms
}

// Corridor represents a connection between rooms
type Corridor struct {
	ID       string            `yaml:"id"`       // Unique corridor identifier
	Start    game.Position     `yaml:"start"`    // Starting position
	End      game.Position     `yaml:"end"`      // Ending position
	Path     []game.Position   `yaml:"path"`     // Corridor path tiles
	Width    int               `yaml:"width"`    // Corridor width
	Style    CorridorStyle     `yaml:"style"`    // Generation style used
	Features []CorridorFeature `yaml:"features"` // Special corridor features
}

// RoomFeature represents special features within rooms
type RoomFeature struct {
	Type       string                 `yaml:"type"`       // Feature type (chest, altar, etc.)
	Position   game.Position          `yaml:"position"`   // Location within room
	Properties map[string]interface{} `yaml:"properties"` // Feature-specific data
}

// CorridorFeature represents special features within corridors
type CorridorFeature struct {
	Type       string                 `yaml:"type"`       // Feature type (trap, secret door, etc.)
	Position   game.Position          `yaml:"position"`   // Location within corridor
	Properties map[string]interface{} `yaml:"properties"` // Feature-specific data
}

// ItemTemplate represents a template for procedural item generation
type ItemTemplate struct {
	BaseType   string                `yaml:"base_type"`   // Base item type (sword, armor, etc.)
	NameParts  []string              `yaml:"name_parts"`  // Name generation components
	StatRanges map[string]StatRange  `yaml:"stat_ranges"` // Stat generation ranges
	Properties []string              `yaml:"properties"`  // Possible item properties
	Enchants   []EnchantmentTemplate `yaml:"enchants"`    // Available enchantments
	Materials  []string              `yaml:"materials"`   // Possible materials
	Rarities   []RarityTier          `yaml:"rarities"`    // Applicable rarity tiers
}

// StatRange represents a range for procedural stat generation
type StatRange struct {
	Min     int     `yaml:"min"`     // Minimum value
	Max     int     `yaml:"max"`     // Maximum value
	Scaling float64 `yaml:"scaling"` // Level scaling factor
}

// EnchantmentTemplate represents a template for procedural enchantments
type EnchantmentTemplate struct {
	Name         string                 `yaml:"name"`         // Enchantment name
	Type         string                 `yaml:"type"`         // Enchantment type
	MinLevel     int                    `yaml:"min_level"`    // Minimum required level
	MaxLevel     int                    `yaml:"max_level"`    // Maximum applicable level
	Effects      []game.Effect          `yaml:"effects"`      // Enchantment effects
	Restrictions map[string]interface{} `yaml:"restrictions"` // Usage restrictions
}

// QuestObjective represents a single quest objective
type QuestObjective struct {
	ID          string                 `yaml:"id"`          // Unique objective ID
	Type        string                 `yaml:"type"`        // Objective type
	Description string                 `yaml:"description"` // Human-readable description
	Target      string                 `yaml:"target"`      // Target entity/location
	Quantity    int                    `yaml:"quantity"`    // Required quantity
	Progress    int                    `yaml:"progress"`    // Current progress
	Complete    bool                   `yaml:"complete"`    // Completion status
	Optional    bool                   `yaml:"optional"`    // Whether objective is optional
	Conditions  map[string]interface{} `yaml:"conditions"`  // Completion conditions
}
