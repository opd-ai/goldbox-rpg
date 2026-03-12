package pcg

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
