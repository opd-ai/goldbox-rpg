package pcg

import (
	"goldbox-rpg/pkg/game"
	"time"
)

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

// Faction-related types

// FactionType represents different categories of factions
type FactionType string

const (
	FactionTypeMilitary  FactionType = "military"
	FactionTypeEconomic  FactionType = "economic"
	FactionTypeReligious FactionType = "religious"
	FactionTypeCriminal  FactionType = "criminal"
	FactionTypeScholarly FactionType = "scholarly"
	FactionTypePolitical FactionType = "political"
	FactionTypeMercenary FactionType = "mercenary"
	FactionTypeMagical   FactionType = "magical"
)

// RelationshipStatus represents the diplomatic status between factions
type RelationshipStatus string

const (
	RelationStatusAllied   RelationshipStatus = "allied"
	RelationStatusFriendly RelationshipStatus = "friendly"
	RelationStatusNeutral  RelationshipStatus = "neutral"
	RelationStatusTense    RelationshipStatus = "tense"
	RelationStatusHostile  RelationshipStatus = "hostile"
	RelationStatusWar      RelationshipStatus = "war"
)

// ResourceType represents different types of resources factions can control (already exists in world.go)

// TerritoryType represents different types of faction territories
type TerritoryType string

const (
	TerritoryTypeCapital     TerritoryType = "capital"
	TerritoryTypeCity        TerritoryType = "city"
	TerritoryTypeOutpost     TerritoryType = "outpost"
	TerritoryTypeFortress    TerritoryType = "fortress"
	TerritoryTypeTradingPost TerritoryType = "trading_post"
	TerritoryTypeResource    TerritoryType = "resource"
)

// ConflictType represents different types of conflicts between factions
type ConflictType string

const (
	ConflictTypeTrade      ConflictType = "trade"
	ConflictTypeTerritory  ConflictType = "territory"
	ConflictTypeReligious  ConflictType = "religious"
	ConflictTypeResource   ConflictType = "resource"
	ConflictTypeSuccession ConflictType = "succession"
	ConflictTypeRevenge    ConflictType = "revenge"
)

// DifficultyProgression already exists in dungeon.go as a struct

// PlotType already exists in narrative.go

// Faction-related structures

// Faction represents a political/social organization
type Faction struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       FactionType            `json:"type"`
	Government GovernmentType         `json:"government"`
	Ideology   string                 `json:"ideology"`
	Power      int                    `json:"power"`
	Wealth     int                    `json:"wealth"`
	Military   int                    `json:"military"`
	Influence  int                    `json:"influence"`
	Stability  float64                `json:"stability"`
	Goals      []string               `json:"goals"`
	Resources  []ResourceType         `json:"resources"`
	Leaders    []*FactionLeader       `json:"leaders"`
	Properties map[string]interface{} `json:"properties"`
}

// FactionLeader represents a leader within a faction
type FactionLeader struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Title      string                 `json:"title"`
	Age        int                    `json:"age"`
	Traits     []string               `json:"traits"`
	Loyalty    float64                `json:"loyalty"`
	Competence float64                `json:"competence"`
	Influence  float64                `json:"influence"`
	Properties map[string]interface{} `json:"properties"`
}

// FactionRelationship represents diplomatic relations between two factions
type FactionRelationship struct {
	ID          string                 `json:"id"`
	Faction1ID  string                 `json:"faction1_id"`
	Faction2ID  string                 `json:"faction2_id"`
	Status      RelationshipStatus     `json:"status"`
	Opinion     float64                `json:"opinion"`     // -1.0 to 1.0
	TrustLevel  float64                `json:"trust_level"` // 0.0 to 1.0
	TradeLevel  float64                `json:"trade_level"` // 0.0 to 1.0
	Hostility   float64                `json:"hostility"`   // 0.0 to 1.0
	History     []string               `json:"history"`
	LastChanged time.Time              `json:"last_changed"`
	Properties  map[string]interface{} `json:"properties"`
}

// Territory represents a geographic area controlled by a faction
type Territory struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         TerritoryType          `json:"type"`
	ControllerID string                 `json:"controller_id"`
	Position     game.Position          `json:"position"`
	Size         int                    `json:"size"`
	Population   int                    `json:"population"`
	Defenses     int                    `json:"defenses"`
	Resources    []ResourceType         `json:"resources"`
	Strategic    bool                   `json:"strategic"`
	Properties   map[string]interface{} `json:"properties"`
}

// TradeDeal represents economic agreements between factions
type TradeDeal struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Faction1ID string                 `json:"faction1_id"`
	Faction2ID string                 `json:"faction2_id"`
	Resource1  ResourceType           `json:"resource1"`
	Resource2  ResourceType           `json:"resource2"`
	Volume1    int                    `json:"volume1"`
	Volume2    int                    `json:"volume2"`
	Duration   int                    `json:"duration"` // Duration in days
	Profit1    int                    `json:"profit1"`
	Profit2    int                    `json:"profit2"`
	Active     bool                   `json:"active"`
	Properties map[string]interface{} `json:"properties"`
}

// Conflict represents ongoing conflicts between factions
type Conflict struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       ConflictType           `json:"type"`
	Factions   []string               `json:"factions"`
	Cause      string                 `json:"cause"`
	Intensity  float64                `json:"intensity"` // 0.0 to 1.0
	Duration   int                    `json:"duration"`  // Duration in days
	Territory  string                 `json:"territory"` // Disputed territory ID
	Resolution string                 `json:"resolution"`
	Active     bool                   `json:"active"`
	Properties map[string]interface{} `json:"properties"`
}

// GeneratedFactionSystem represents a complete faction system
type GeneratedFactionSystem struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Factions      []*Faction             `json:"factions"`
	Relationships []*FactionRelationship `json:"relationships"`
	Territories   []*Territory           `json:"territories"`
	TradeDeals    []*TradeDeal           `json:"trade_deals"`
	Conflicts     []*Conflict            `json:"conflicts"`
	Metadata      map[string]interface{} `json:"metadata"`
	Generated     time.Time              `json:"generated"`
}

// FactionParams provides faction-specific generation parameters
type FactionParams struct {
	GenerationParams   `yaml:",inline"`
	FactionCount       int     `yaml:"faction_count"`       // Number of factions to generate
	MinPower           int     `yaml:"min_power"`           // Minimum faction power level
	MaxPower           int     `yaml:"max_power"`           // Maximum faction power level
	ConflictLevel      float64 `yaml:"conflict_level"`      // Overall conflict intensity (0.0-1.0)
	EconomicFocus      float64 `yaml:"economic_focus"`      // Economic activity emphasis (0.0-1.0)
	MilitaryFocus      float64 `yaml:"military_focus"`      // Military emphasis (0.0-1.0)
	CulturalFocus      float64 `yaml:"cultural_focus"`      // Cultural/religious emphasis (0.0-1.0)
	TerritoryCount     int     `yaml:"territory_count"`     // Number of territories per faction
	TradeVolume        float64 `yaml:"trade_volume"`        // Overall trade activity (0.0-1.0)
	PoliticalStability float64 `yaml:"political_stability"` // Overall political stability (0.0-1.0)
}

// CharacterType represents different categories of NPCs
type CharacterType string

const (
	CharacterTypeGeneric  CharacterType = "generic"  // General purpose NPC
	CharacterTypeMerchant CharacterType = "merchant" // Shop keeper or trader
	CharacterTypeGuard    CharacterType = "guard"    // Security or military
	CharacterTypeNoble    CharacterType = "noble"    // Aristocracy or leadership
	CharacterTypePeasant  CharacterType = "peasant"  // Common folk or laborers
	CharacterTypeCrafter  CharacterType = "crafter"  // Artisan or specialist
	CharacterTypeCleric   CharacterType = "cleric"   // Religious figure
	CharacterTypeMage     CharacterType = "mage"     // Magic user or scholar
	CharacterTypeRogue    CharacterType = "rogue"    // Thief or scoundrel
	CharacterTypeBard     CharacterType = "bard"     // Entertainer or storyteller
)

// BackgroundType represents character background categories
type BackgroundType string

const (
	BackgroundUrban     BackgroundType = "urban"     // City dweller
	BackgroundRural     BackgroundType = "rural"     // Countryside origin
	BackgroundNomadic   BackgroundType = "nomadic"   // Travel-oriented background
	BackgroundNoble     BackgroundType = "noble"     // Aristocratic upbringing
	BackgroundCriminal  BackgroundType = "criminal"  // Unlawful background
	BackgroundMilitary  BackgroundType = "military"  // Armed forces background
	BackgroundReligious BackgroundType = "religious" // Religious/monastic background
	BackgroundScholar   BackgroundType = "scholar"   // Academic background
	BackgroundWilderness BackgroundType = "wilderness" // Outdoor/survival background
)

// SocialClass represents character's social standing
type SocialClass string

const (
	SocialClassSlave    SocialClass = "slave"     // Lowest social position
	SocialClassSerf     SocialClass = "serf"      // Bound peasant
	SocialClassPeasant  SocialClass = "peasant"   // Free commoner
	SocialClassCrafter  SocialClass = "crafter"   // Skilled artisan
	SocialClassMerchant SocialClass = "merchant"  // Trade class
	SocialClassGentry   SocialClass = "gentry"    // Minor nobility
	SocialClassNoble    SocialClass = "noble"     // Aristocracy
	SocialClassRoyalty  SocialClass = "royalty"   // Ruling class
)

// AgeRange represents character age categories
type AgeRange string

const (
	AgeRangeChild      AgeRange = "child"       // Young character (5-12)
	AgeRangeAdolescent AgeRange = "adolescent"  // Teenage character (13-17)
	AgeRangeYoungAdult AgeRange = "young_adult" // Young adult (18-25)
	AgeRangeAdult      AgeRange = "adult"       // Mature adult (26-40)
	AgeRangeMiddleAged AgeRange = "middle_aged" // Middle-aged (41-60)
	AgeRangeElderly    AgeRange = "elderly"     // Elderly character (61+)
	AgeRangeAncient    AgeRange = "ancient"     // Very old character (special cases)
)

// NPCGroupType represents different types of NPC groups
type NPCGroupType string

const (
	NPCGroupFamily    NPCGroupType = "family"    // Related family members
	NPCGroupGuards    NPCGroupType = "guards"    // Security patrol or unit
	NPCGroupMerchants NPCGroupType = "merchants" // Trading group or caravan
	NPCGroupCultists  NPCGroupType = "cultists"  // Religious or cult group
	NPCGroupBandits   NPCGroupType = "bandits"   // Criminal organization
	NPCGroupScholars  NPCGroupType = "scholars"  // Academic or research group
	NPCGroupCrafters  NPCGroupType = "crafters"  // Guild or workshop group
)

// PersonalityTrait represents individual personality characteristics
type PersonalityTrait struct {
	Name        string  `json:"name"`        // Trait name (e.g., "brave", "greedy")
	Intensity   float64 `json:"intensity"`   // Trait strength (0.0-1.0)
	Description string  `json:"description"` // Descriptive text
}

// Motivation represents character goals and drives
type Motivation struct {
	Type        string  `json:"type"`        // Motivation category (power, wealth, love, etc.)
	Target      string  `json:"target"`      // What the motivation is directed toward
	Intensity   float64 `json:"intensity"`   // How strongly motivated (0.0-1.0)
	Description string  `json:"description"` // Detailed description
}

// PersonalityProfile represents a complete character personality system
type PersonalityProfile struct {
	Traits      []PersonalityTrait `json:"traits"`      // Individual personality traits
	Motivations []Motivation       `json:"motivations"` // Character goals and drives
	Alignment   string             `json:"alignment"`   // Moral alignment
	Temperament string             `json:"temperament"` // General disposition
	Values      []string           `json:"values"`      // What the character values most
	Fears       []string           `json:"fears"`       // Character's primary fears
	Speech      SpeechPattern      `json:"speech"`      // How the character speaks
}

// SpeechPattern represents how a character communicates
type SpeechPattern struct {
	Formality   string   `json:"formality"`   // Level of formality (formal, casual, crude)
	Vocabulary  string   `json:"vocabulary"`  // Complexity level (simple, moderate, complex)
	Accent      string   `json:"accent"`      // Regional or cultural accent
	Mannerisms  []string `json:"mannerisms"`  // Speech habits or quirks
	Catchphrase string   `json:"catchphrase"` // Signature phrase (optional)
}
