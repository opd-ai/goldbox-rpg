package pcg

import (
	"context"
	"time"

	"goldbox-rpg/pkg/game"
)

// Generator is the base interface for all procedural content generators
// All PCG implementations must satisfy this interface for registration and execution
type Generator interface {
	// Generate creates content based on the provided context and parameters
	// Returns the generated content and any error that occurred
	Generate(ctx context.Context, params GenerationParams) (interface{}, error)

	// GetType returns the content type this generator produces
	GetType() ContentType

	// GetVersion returns the generator version for compatibility checking
	GetVersion() string

	// Validate checks if the provided parameters are valid for this generator
	Validate(params GenerationParams) error
}

// TerrainGenerator specializes in generating terrain and map layouts
type TerrainGenerator interface {
	Generator

	// GenerateTerrain creates a 2D terrain map with the specified dimensions
	GenerateTerrain(ctx context.Context, width, height int, params TerrainParams) (*game.GameMap, error)

	// GenerateBiome creates terrain for a specific biome type
	GenerateBiome(ctx context.Context, biome BiomeType, bounds Rectangle, params TerrainParams) (*game.GameMap, error)

	// ValidateConnectivity ensures generated terrain has proper pathfinding connectivity
	ValidateConnectivity(terrain *game.GameMap) bool
}

// ItemGenerator specializes in generating items with procedural properties
type ItemGenerator interface {
	Generator

	// GenerateItem creates a single item based on templates and rarity
	GenerateItem(ctx context.Context, template ItemTemplate, params ItemParams) (*game.Item, error)

	// GenerateItemSet creates a collection of related items (e.g., armor set)
	GenerateItemSet(ctx context.Context, setType ItemSetType, params ItemParams) ([]*game.Item, error)

	// GenerateRandomItem creates an item appropriate for the given context
	GenerateRandomItem(ctx context.Context, level int, rarity RarityTier, params ItemParams) (*game.Item, error)
}

// LevelGenerator specializes in generating complete game levels/dungeons
type LevelGenerator interface {
	Generator

	// GenerateLevel creates a complete game level with rooms, corridors, and features
	GenerateLevel(ctx context.Context, params LevelParams) (*game.Level, error)

	// GenerateRoom creates a single room with specified constraints
	GenerateRoom(ctx context.Context, bounds Rectangle, roomType RoomType, params LevelParams) (*RoomLayout, error)

	// ConnectRooms generates corridors and passages between rooms
	ConnectRooms(ctx context.Context, rooms []RoomLayout, params LevelParams) ([]Corridor, error)
}

// QuestGenerator specializes in generating quests and storylines
type QuestGenerator interface {
	Generator

	// GenerateQuest creates a quest with objectives and rewards
	GenerateQuest(ctx context.Context, questType QuestType, params QuestParams) (*game.Quest, error)

	// GenerateQuestChain creates a series of connected quests
	GenerateQuestChain(ctx context.Context, chainLength int, params QuestParams) ([]*game.Quest, error)

	// GenerateObjectives creates quest objectives based on available content
	GenerateObjectives(ctx context.Context, world *game.World, params QuestParams) ([]QuestObjective, error)
}

// CharacterGenerator specializes in generating NPCs with personalities and motivations
type CharacterGenerator interface {
	Generator

	// GenerateNPC creates a single NPC with personality and motivations
	GenerateNPC(ctx context.Context, characterType CharacterType, params CharacterParams) (*game.NPC, error)

	// GenerateNPCGroup creates a collection of related NPCs (e.g., family, guards)
	GenerateNPCGroup(ctx context.Context, groupType NPCGroupType, params CharacterParams) ([]*game.NPC, error)

	// GeneratePersonality creates personality traits and motivations for an existing character
	GeneratePersonality(ctx context.Context, character *game.Character, params CharacterParams) (*PersonalityProfile, error)
}

// ContentType represents the type of content being generated
type ContentType string

const (
	ContentTypeTerrain    ContentType = "terrain"
	ContentTypeItems      ContentType = "items"
	ContentTypeLevels     ContentType = "levels"
	ContentTypeQuests     ContentType = "quests"
	ContentTypeCharacters ContentType = "characters"
	ContentTypeNPCs       ContentType = "npcs"
	ContentTypeEvents     ContentType = "events"
	ContentTypeDungeon    ContentType = "dungeon"
	ContentTypeNarrative  ContentType = "narrative"
	ContentTypeFactions   ContentType = "factions"
	ContentTypeDialogue   ContentType = "dialogue"
)

// GenerationParams provides common parameters for all generators
type GenerationParams struct {
	Seed        int64                  `yaml:"seed"`         // Deterministic seed for reproducible generation
	Difficulty  int                    `yaml:"difficulty"`   // Target difficulty level (1-20)
	PlayerLevel int                    `yaml:"player_level"` // Average party level for scaling
	WorldState  *game.World            `yaml:"-"`            // Current world state for context
	Constraints map[string]interface{} `yaml:"constraints"`  // Generator-specific constraints
	Metadata    map[string]interface{} `yaml:"metadata"`     // Additional context data
	Timeout     time.Duration          `yaml:"timeout"`      // Maximum generation time
}

// TerrainParams provides terrain-specific generation parameters
type TerrainParams struct {
	GenerationParams `yaml:",inline"`
	BiomeType        BiomeType         `yaml:"biome_type"`   // Target biome for generation
	Density          float64           `yaml:"density"`      // Feature density (0.0-1.0)
	Connectivity     ConnectivityLevel `yaml:"connectivity"` // Required connectivity level
	Features         []TerrainFeature  `yaml:"features"`     // Special features to include
	WaterLevel       float64           `yaml:"water_level"`  // Water coverage percentage
	Roughness        float64           `yaml:"roughness"`    // Terrain complexity
}

// ItemParams provides item-specific generation parameters
type ItemParams struct {
	GenerationParams `yaml:",inline"`
	MinRarity        RarityTier `yaml:"min_rarity"`       // Minimum item rarity
	MaxRarity        RarityTier `yaml:"max_rarity"`       // Maximum item rarity
	ItemTypes        []string   `yaml:"item_types"`       // Allowed item types
	EnchantmentRate  float64    `yaml:"enchantment_rate"` // Probability of enchantments
	UniqueChance     float64    `yaml:"unique_chance"`    // Probability of unique items
	LevelScaling     bool       `yaml:"level_scaling"`    // Whether to scale with player level
}

// LevelParams provides level-specific generation parameters
type LevelParams struct {
	GenerationParams `yaml:",inline"`
	MinRooms         int           `yaml:"min_rooms"`      // Minimum number of rooms
	MaxRooms         int           `yaml:"max_rooms"`      // Maximum number of rooms
	RoomTypes        []RoomType    `yaml:"room_types"`     // Allowed room types
	CorridorStyle    CorridorStyle `yaml:"corridor_style"` // Corridor generation style
	LevelTheme       LevelTheme    `yaml:"level_theme"`    // Thematic constraints
	HasBoss          bool          `yaml:"has_boss"`       // Whether to include a boss room
	SecretRooms      int           `yaml:"secret_rooms"`   // Number of secret rooms
}

// QuestParams provides quest-specific generation parameters
type QuestParams struct {
	GenerationParams `yaml:",inline"`
	QuestType        QuestType     `yaml:"quest_type"`      // Type of quest to generate
	MinObjectives    int           `yaml:"min_objectives"`  // Minimum quest objectives
	MaxObjectives    int           `yaml:"max_objectives"`  // Maximum quest objectives
	RewardTier       RarityTier    `yaml:"reward_tier"`     // Quality of quest rewards
	Narrative        NarrativeType `yaml:"narrative"`       // Story generation style
	RequiredItems    []string      `yaml:"required_items"`  // Items that must be involved
	ForbiddenItems   []string      `yaml:"forbidden_items"` // Items to exclude
}

// DungeonParams provides dungeon-specific generation parameters
type DungeonParams struct {
	GenerationParams `yaml:",inline"`
	LevelCount       int                   `yaml:"level_count"`     // Number of levels in the dungeon
	LevelWidth       int                   `yaml:"level_width"`     // Width of each level
	LevelHeight      int                   `yaml:"level_height"`    // Height of each level
	RoomsPerLevel    int                   `yaml:"rooms_per_level"` // Target number of rooms per level
	Theme            LevelTheme            `yaml:"theme"`           // Overall dungeon theme
	Connectivity     ConnectivityLevel     `yaml:"connectivity"`    // Level connectivity requirements
	Density          float64               `yaml:"density"`         // Feature density (0.0-1.0)
	Difficulty       DifficultyProgression `yaml:"difficulty"`      // Difficulty scaling across levels
}

// NarrativeParams provides narrative-specific generation parameters
type NarrativeParams struct {
	GenerationParams `yaml:",inline"`
	NarrativeType    NarrativeType `yaml:"narrative_type"`   // Type of narrative structure
	Theme            string        `yaml:"theme"`            // Overall narrative theme
	CampaignLength   string        `yaml:"campaign_length"`  // Length of campaign (short/medium/long)
	ComplexityLevel  int           `yaml:"complexity_level"` // Story complexity (1-5)
	CharacterFocus   bool          `yaml:"character_focus"`  // Whether to focus on character development
	MainAntagonist   string        `yaml:"main_antagonist"`  // Type of primary antagonist
	ConflictType     string        `yaml:"conflict_type"`    // Primary type of conflict
	TonePreference   string        `yaml:"tone_preference"`  // Preferred narrative tone
}

// CharacterParams provides character-specific generation parameters
type CharacterParams struct {
	GenerationParams `yaml:",inline"`
	CharacterType    CharacterType  `yaml:"character_type"`    // Type of character to generate
	PersonalityDepth int            `yaml:"personality_depth"` // Complexity of personality (1-5)
	MotivationCount  int            `yaml:"motivation_count"`  // Number of motivations
	BackgroundType   BackgroundType `yaml:"background_type"`   // Character background category
	Alignment        string         `yaml:"alignment"`         // Moral alignment preference
	SocialClass      SocialClass    `yaml:"social_class"`      // Character's social standing
	AgeRange         AgeRange       `yaml:"age_range"`         // Character age category
	Gender           string         `yaml:"gender"`            // Character gender (optional constraint)
	Faction          string         `yaml:"faction"`           // Associated faction (optional)
	Profession       string         `yaml:"profession"`        // Character's profession (optional)
	UniqueTraits     int            `yaml:"unique_traits"`     // Number of distinctive traits
}
