package pcg

import "goldbox-rpg/pkg/game"

// BiomeType represents different terrain biomes for generation
type BiomeType string

const (
	BiomeForest    BiomeType = "forest"
	BiomeMountain  BiomeType = "mountain"
	BiomeDesert    BiomeType = "desert"
	BiomeSwamp     BiomeType = "swamp"
	BiomeCave      BiomeType = "cave"
	BiomeDungeon   BiomeType = "dungeon"
	BiomeCoastal   BiomeType = "coastal"
	BiomeUrban     BiomeType = "urban"
	BiomeWasteland BiomeType = "wasteland"
)

// RarityTier represents item rarity levels
type RarityTier string

const (
	RarityCommon    RarityTier = "common"
	RarityUncommon  RarityTier = "uncommon"
	RarityRare      RarityTier = "rare"
	RarityEpic      RarityTier = "epic"
	RarityLegendary RarityTier = "legendary"
	RarityArtifact  RarityTier = "artifact"
)

// RoomType represents different types of rooms in generated levels
type RoomType string

const (
	RoomTypeEntrance RoomType = "entrance"
	RoomTypeExit     RoomType = "exit"
	RoomTypeCombat   RoomType = "combat"
	RoomTypeTreasure RoomType = "treasure"
	RoomTypePuzzle   RoomType = "puzzle"
	RoomTypeBoss     RoomType = "boss"
	RoomTypeSecret   RoomType = "secret"
	RoomTypeShop     RoomType = "shop"
	RoomTypeRest     RoomType = "rest"
	RoomTypeTrap     RoomType = "trap"
	RoomTypeStory    RoomType = "story"
)

// CorridorStyle represents different corridor generation approaches
type CorridorStyle string

const (
	CorridorStraight CorridorStyle = "straight"
	CorridorWindy    CorridorStyle = "windy"
	CorridorMaze     CorridorStyle = "maze"
	CorridorOrganic  CorridorStyle = "organic"
	CorridorMinimal  CorridorStyle = "minimal"
)

// LevelTheme represents thematic constraints for level generation
type LevelTheme string

const (
	ThemeClassic    LevelTheme = "classic"
	ThemeHorror     LevelTheme = "horror"
	ThemeNatural    LevelTheme = "natural"
	ThemeMechanical LevelTheme = "mechanical"
	ThemeMagical    LevelTheme = "magical"
	ThemeUndead     LevelTheme = "undead"
	ThemeElemental  LevelTheme = "elemental"
)

// QuestType represents different categories of quests
type QuestType string

const (
	QuestTypeFetch    QuestType = "fetch"
	QuestTypeKill     QuestType = "kill"
	QuestTypeEscort   QuestType = "escort"
	QuestTypeExplore  QuestType = "explore"
	QuestTypeDefend   QuestType = "defend"
	QuestTypePuzzle   QuestType = "puzzle"
	QuestTypeDelivery QuestType = "delivery"
	QuestTypeSurvival QuestType = "survival"
	QuestTypeStory    QuestType = "story"
)

// NarrativeType represents different story generation styles
type NarrativeType string

const (
	NarrativeLinear    NarrativeType = "linear"
	NarrativeBranching NarrativeType = "branching"
	NarrativeOpen      NarrativeType = "open"
	NarrativeEpisodic  NarrativeType = "episodic"
)

// ConnectivityLevel represents how connected terrain features should be
type ConnectivityLevel string

const (
	ConnectivityNone     ConnectivityLevel = "none"
	ConnectivityLow      ConnectivityLevel = "low"
	ConnectivityMinimal  ConnectivityLevel = "minimal"
	ConnectivityModerate ConnectivityLevel = "moderate"
	ConnectivityHigh     ConnectivityLevel = "high"
	ConnectivityComplete ConnectivityLevel = "complete"
)

// TerrainFeature represents special features that can be included in terrain
type TerrainFeature string

const (
	FeatureWater            TerrainFeature = "water"
	FeatureMountain         TerrainFeature = "mountain"
	FeatureForest           TerrainFeature = "forest"
	FeatureCave             TerrainFeature = "cave"
	FeatureRuins            TerrainFeature = "ruins"
	FeatureRoad             TerrainFeature = "road"
	FeatureBridge           TerrainFeature = "bridge"
	FeatureTown             TerrainFeature = "town"
	FeatureShrine           TerrainFeature = "shrine"
	FeatureStalactites      TerrainFeature = "stalactites"
	FeatureUndergroundRiver TerrainFeature = "underground_river"
	FeatureSecretDoors      TerrainFeature = "secret_doors"
	FeatureTraps            TerrainFeature = "traps"
	FeatureTrees            TerrainFeature = "trees"
	FeatureStreams          TerrainFeature = "streams"
	FeatureCliffs           TerrainFeature = "cliffs"
	FeatureCrevasses        TerrainFeature = "crevasses"
	FeatureBogs             TerrainFeature = "bogs"
	FeatureVines            TerrainFeature = "vines"
	FeatureDunes            TerrainFeature = "dunes"
	FeatureOasis            TerrainFeature = "oasis"
)

// ItemSetType represents collections of related items
type ItemSetType string

const (
	ItemSetArmor    ItemSetType = "armor"
	ItemSetWeapons  ItemSetType = "weapons"
	ItemSetJewelry  ItemSetType = "jewelry"
	ItemSetTools    ItemSetType = "tools"
	ItemSetConsumab ItemSetType = "consumables"
	ItemSetMagical  ItemSetType = "magical"
	ItemSetCrafting ItemSetType = "crafting"
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
