# Procedural Content Generation Implementation Guide

## Current Implementation Status (Updated)

The PCG system has been substantially implemented with complete terrain and item generation systems:

**✅ Completed:**
- ✅ Core interfaces (`interfaces.go`)
- ✅ Type definitions and enums (`types.go`) 
- ✅ Registry and factory system (`registry.go`)
- ✅ Deterministic seeding system (`seed.go`)
- ✅ Content validation framework (`validation.go`)
- ✅ PCG manager and world integration (`manager.go`)
- ✅ **Complete terrain generation system:**
  - Biome definitions and feature distribution (`terrain/biomes.go`)
  - Cellular automata algorithm (`terrain/cellular_automata.go`)
  - Maze generation with recursive backtracking (`terrain/maze.go`)
  - Perlin and Simplex noise utilities (`utils/noise.go`)
  - A* pathfinding and connectivity validation (`utils/pathfinding.go`)
  - Comprehensive unit tests with >95% coverage
- ✅ **Complete item generation system:**
  - Template-based item generator (`items/generator.go`)
  - Item template registry with rarity modifiers (`items/templates.go`)
  - Enchantment system with magic schools (`items/enchantments.go`)
  - YAML configuration support (`data/pcg/item_templates.yaml`)
  - Full test suite covering all functionality
- ✅ **Complete level/dungeon generation system:**
  - Room-corridor level generator (`levels/generator.go`)
  - Specialized room generators for all room types (`levels/rooms.go`)
  - Advanced corridor planning with multiple styles (`levels/corridors.go`)
  - BSP-based room layout with theme support
  - Comprehensive test suite with 100% pass rate
- ✅ **API integration and server endpoints:**
  - PCG handler methods in `pkg/server/handlers.go`
  - Content generation, validation, and stats endpoints
  - Generator registration and management
  - Comprehensive test suite covering all endpoints
- ✅ **Quest generation system:**
  - Core generator, narratives, and objectives fully implemented
  - See `quests/generator.go`, `quests/narratives.go`, and `quests/objectives.go`
  - >95% test coverage with table-driven tests
- ✅ **Documentation and usage examples:**
  - See [`pkg/pcg/README.md`](README.md) for complete documentation and code examples
  - Includes setup, API usage, configuration, integration, and testing patterns

**✅ Recently Completed:**
- ✅ **Maze generation test suite (`terrain/maze_test.go`) - COMPLETED**
  - Comprehensive test coverage for MazeGenerator implementation
  - Tests for all public methods: Generate, Validate, GenerateTerrain, ValidateConnectivity, GenerateBiome
  - Edge case testing for connectivity validation and parameter validation
  - Deterministic generation verification with seed consistency
  - Helper functions for test maze creation and validation
  - Table-driven test patterns following project conventions
  - 100% test pass rate with proper error handling coverage
- ✅ **Performance metrics and monitoring system (`metrics.go`) - COMPLETED**
  - Full implementation of GenerationMetrics with comprehensive tracking
  - Thread-safe performance statistics collection for all content types
  - Integration with PCGManager for automatic timing and error recording
  - Cache hit/miss ratio tracking and detailed performance analytics
  - Complete test suite with >95% coverage including concurrency tests
  - Real-time generation statistics available through GetStats() API

## Implementation Roadmap

### Phase 1: Complete Core Terrain Generation

#### 1.1 Enhance Terrain Generator (`pkg/pcg/terrain/`)

**File: `pkg/pcg/terrain/biomes.go`**
```go
package terrain

import (
    "goldbox-rpg/pkg/pcg"
)

// BiomeDefinition defines the characteristics of a biome
type BiomeDefinition struct {
    Type              pcg.BiomeType `yaml:"type"`
    DefaultDensity    float64       `yaml:"default_density"`
    WaterLevelRange   [2]float64    `yaml:"water_level_range"`
    RoughnessRange    [2]float64    `yaml:"roughness_range"`
    ConnectivityLevel pcg.ConnectivityLevel `yaml:"connectivity_level"`
    Features          []pcg.TerrainFeature  `yaml:"features"`
    TileDistribution  map[string]float64    `yaml:"tile_distribution"`
}

// GetBiomeDefinition returns the definition for a specific biome
func GetBiomeDefinition(biome pcg.BiomeType) (*BiomeDefinition, error) {
    // Implementation: Load from YAML configuration
    // Return biome-specific parameters for generation
}

// ApplyBiomeModifications modifies generation parameters based on biome
func ApplyBiomeModifications(params *pcg.TerrainParams, biome pcg.BiomeType) error {
    // Implementation: Adjust density, water level, features based on biome
}
```

**File: `pkg/pcg/terrain/cellular_automata.go`**
```go
package terrain

import (
    "context"
    "goldbox-rpg/pkg/game"
    "goldbox-rpg/pkg/pcg"
)

// CellularAutomataConfig holds configuration for the algorithm
type CellularAutomataConfig struct {
    WallThreshold      int     `yaml:"wall_threshold"`      // Neighbor count for wall formation
    FloorThreshold     int     `yaml:"floor_threshold"`     // Neighbor count for floor formation
    MaxIterations      int     `yaml:"max_iterations"`      // Maximum CA iterations
    SmoothingPasses    int     `yaml:"smoothing_passes"`    // Post-processing smoothing
    EdgeBuffer         int     `yaml:"edge_buffer"`         // Border wall thickness
    MinRoomSize        int     `yaml:"min_room_size"`       // Minimum viable room size
}

// DefaultCAConfig returns default cellular automata configuration
func DefaultCAConfig() *CellularAutomataConfig {
    return &CellularAutomataConfig{
        WallThreshold:   5,
        FloorThreshold:  3,
        MaxIterations:   6,
        SmoothingPasses: 2,
        EdgeBuffer:      1,
        MinRoomSize:     16,
    }
}

// RunCellularAutomata executes the cellular automata algorithm
func RunCellularAutomata(gameMap *game.GameMap, config *CellularAutomataConfig, genCtx *pcg.GenerationContext) error {
    // Implementation:
    // 1. Initialize random noise based on density
    // 2. Apply cellular automata rules for specified iterations
    // 3. Remove small disconnected areas
    // 4. Apply smoothing passes
    // 5. Ensure proper edge boundaries
}
```

**File: `pkg/pcg/terrain/maze.go`**
```go
package terrain

import (
    "goldbox-rpg/pkg/game"
    "goldbox-rpg/pkg/pcg"
)

// MazeGenerator creates maze-like terrain structures
type MazeGenerator struct {
    version string
}

// NewMazeGenerator creates a new maze terrain generator
func NewMazeGenerator() *MazeGenerator {
    return &MazeGenerator{version: "1.0.0"}
}

// Generate implements the Generator interface for maze terrain
func (mg *MazeGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
    // Implementation: Extract parameters and delegate to GenerateTerrain
}

// GenerateTerrain creates maze-style terrain using recursive backtracking
func (mg *MazeGenerator) GenerateTerrain(ctx context.Context, width, height int, params pcg.TerrainParams) (*game.GameMap, error) {
    // Implementation:
    // 1. Create grid with all walls
    // 2. Use recursive backtracking to carve passages
    // 3. Add rooms and special features
    // 4. Apply biome-specific modifications
}
```

#### 1.2 Add Noise-Based Generation (`pkg/pcg/utils/`)

**File: `pkg/pcg/utils/noise.go`**
```go
package utils

import "math"

// PerlinNoise generates Perlin noise for terrain generation
type PerlinNoise struct {
    seed       int64
    permutation []int
}

// NewPerlinNoise creates a new Perlin noise generator with seed
func NewPerlinNoise(seed int64) *PerlinNoise {
    // Implementation: Initialize permutation table with seed
}

// Noise2D generates 2D Perlin noise value at coordinates
func (pn *PerlinNoise) Noise2D(x, y float64) float64 {
    // Implementation: Standard Perlin noise algorithm
}

// FractalNoise generates fractal noise by combining multiple octaves
func (pn *PerlinNoise) FractalNoise(x, y float64, octaves int, persistence, scale float64) float64 {
    // Implementation: Sum multiple noise octaves with decreasing amplitude
}

// SimplexNoise provides faster alternative to Perlin noise
type SimplexNoise struct {
    seed int64
    grad [][]float64
}

// NewSimplexNoise creates a new Simplex noise generator
func NewSimplexNoise(seed int64) *SimplexNoise {
    // Implementation: Initialize gradient table
}

// Noise2D generates 2D Simplex noise
func (sn *SimplexNoise) Noise2D(x, y float64) float64 {
    // Implementation: Simplex noise algorithm
}
```

**File: `pkg/pcg/utils/pathfinding.go`**
```go
package utils

import "goldbox-rpg/pkg/game"

// PathfindingResult represents the result of pathfinding
type PathfindingResult struct {
    Path     []game.Position `json:"path"`
    Found    bool           `json:"found"`
    Distance int            `json:"distance"`
}

// AStarPathfind finds optimal path using A* algorithm
func AStarPathfind(gameMap *game.GameMap, start, goal game.Position) *PathfindingResult {
    // Implementation:
    // 1. Initialize open and closed sets
    // 2. Calculate heuristic costs
    // 3. Find optimal path
    // 4. Return path or indicate failure
}

// FloodFill finds all connected walkable areas
func FloodFill(gameMap *game.GameMap, start game.Position) []game.Position {
    // Implementation: Standard flood fill algorithm to find connected components
}

// ValidateConnectivity checks if all walkable areas are connected
func ValidateConnectivity(gameMap *game.GameMap) bool {
    // Implementation:
    // 1. Find all walkable tiles
    // 2. Use flood fill from first walkable tile
    // 3. Check if all walkable tiles are reachable
}
```

### Phase 2: Implement Item Generation System

#### 2.1 Create Item Generator (`pkg/pcg/items/`)

**File: `pkg/pcg/items/generator.go`**
```go
package items

import (
    "context"
    "goldbox-rpg/pkg/game"
    "goldbox-rpg/pkg/pcg"
)

// TemplateBasedGenerator generates items using template system
type TemplateBasedGenerator struct {
    version   string
    templates map[string]*pcg.ItemTemplate
}

// NewTemplateBasedGenerator creates a new template-based item generator
func NewTemplateBasedGenerator() *TemplateBasedGenerator {
    return &TemplateBasedGenerator{
        version:   "1.0.0",
        templates: make(map[string]*pcg.ItemTemplate),
    }
}

// LoadTemplates loads item templates from YAML configuration
func (tbg *TemplateBasedGenerator) LoadTemplates(configPath string) error {
    // Implementation: Load and parse YAML template definitions
}

// Generate implements the Generator interface
func (tbg *TemplateBasedGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
    // Implementation: Generate items based on parameters and templates
}

// GenerateItem creates a single item from template
func (tbg *TemplateBasedGenerator) GenerateItem(ctx context.Context, template pcg.ItemTemplate, params pcg.ItemParams) (*game.Item, error) {
    // Implementation:
    // 1. Select base item type from template
    // 2. Generate procedural name
    // 3. Roll stats within template ranges
    // 4. Apply level scaling
    // 5. Add enchantments based on rarity
    // 6. Set appropriate value and weight
}

// GenerateItemSet creates a collection of related items
func (tbg *TemplateBasedGenerator) GenerateItemSet(ctx context.Context, setType pcg.ItemSetType, params pcg.ItemParams) ([]*game.Item, error) {
    // Implementation: Generate coordinated item sets with thematic consistency
}
```

**File: `pkg/pcg/items/templates.go`**
```go
package items

import (
    "goldbox-rpg/pkg/pcg"
)

// ItemTemplateRegistry manages available item templates
type ItemTemplateRegistry struct {
    templates map[string]*pcg.ItemTemplate
    rarityModifiers map[pcg.RarityTier]RarityModifier
}

// RarityModifier defines how rarity affects item generation
type RarityModifier struct {
    StatMultiplier     float64   `yaml:"stat_multiplier"`
    EnchantmentChance  float64   `yaml:"enchantment_chance"`
    MaxEnchantments    int       `yaml:"max_enchantments"`
    ValueMultiplier    float64   `yaml:"value_multiplier"`
    NamePrefixes       []string  `yaml:"name_prefixes"`
    NameSuffixes       []string  `yaml:"name_suffixes"`
}

// LoadDefaultTemplates loads built-in item templates
func (itr *ItemTemplateRegistry) LoadDefaultTemplates() error {
    // Implementation: Define default weapon, armor, consumable templates
}

// GetTemplate retrieves template by base type and rarity
func (itr *ItemTemplateRegistry) GetTemplate(baseType string, rarity pcg.RarityTier) (*pcg.ItemTemplate, error) {
    // Implementation: Return appropriate template with rarity modifications
}

// GenerateItemName creates procedural item names
func GenerateItemName(template *pcg.ItemTemplate, rarity pcg.RarityTier, genCtx *pcg.GenerationContext) string {
    // Implementation:
    // 1. Select base name from template
    // 2. Add rarity-appropriate prefixes/suffixes
    // 3. Include material qualifiers
    // 4. Ensure name uniqueness
}
```

**File: `pkg/pcg/items/enchantments.go`**
```go
package items

import (
    "goldbox-rpg/pkg/game"
    "goldbox-rpg/pkg/pcg"
)

// EnchantmentSystem manages procedural enchantments
type EnchantmentSystem struct {
    enchantments map[string]*pcg.EnchantmentTemplate
    schools      map[string][]string // Magic school -> enchantment list
}

// NewEnchantmentSystem creates a new enchantment system
func NewEnchantmentSystem() *EnchantmentSystem {
    // Implementation: Initialize with default enchantments
}

// ApplyEnchantments adds procedural enchantments to an item
func (es *EnchantmentSystem) ApplyEnchantments(item *game.Item, rarity pcg.RarityTier, playerLevel int, genCtx *pcg.GenerationContext) error {
    // Implementation:
    // 1. Determine number of enchantments based on rarity
    // 2. Select appropriate enchantments for item type
    // 3. Scale enchantment power to player level
    // 4. Apply enchantment effects to item
    // 5. Update item name and value
}

// GetAvailableEnchantments returns enchantments valid for item type
func (es *EnchantmentSystem) GetAvailableEnchantments(itemType string, minLevel, maxLevel int) []*pcg.EnchantmentTemplate {
    // Implementation: Filter enchantments by item type and level requirements
}
```

#### 2.2 Configuration System (`data/pcg/`)

**File: `data/pcg/item_templates.yaml`**
```yaml
weapon_templates:
  sword:
    base_type: "weapon"
    name_parts: ["Blade", "Sword", "Saber", "Falchion"]
    damage_range: [6, 8]
    stat_ranges:
      damage: {min: 1, max: 6, scaling: 0.1}
      critical: {min: 19, max: 20, scaling: 0.0}
    properties: ["slashing", "martial"]
    materials: ["iron", "steel", "mithril", "adamantine"]
    rarities: ["common", "uncommon", "rare", "epic", "legendary"]

  bow:
    base_type: "weapon"
    name_parts: ["Bow", "Longbow", "Shortbow", "Recurve"]
    damage_range: [6, 6]
    stat_ranges:
      damage: {min: 1, max: 6, scaling: 0.1}
      range: {min: 80, max: 150, scaling: 1.0}
    properties: ["ranged", "martial", "ammunition"]
    materials: ["wood", "yew", "ironwood", "dragonbone"]

armor_templates:
  leather_armor:
    base_type: "armor"
    name_parts: ["Leather", "Hide", "Studded"]
    stat_ranges:
      ac: {min: 11, max: 12, scaling: 0.05}
      max_dex: {min: 2, max: 6, scaling: 0.1}
    properties: ["light"]
    materials: ["leather", "studded_leather", "dragonskin"]

consumable_templates:
  healing_potion:
    base_type: "consumable"
    name_parts: ["Potion", "Elixir", "Draught"]
    stat_ranges:
      healing: {min: 8, max: 16, scaling: 0.5}
    properties: ["consumable", "magical"]
    materials: ["glass", "crystal", "vial"]

enchantment_templates:
  weapon_enhancement:
    name: "Enhancement"
    type: "weapon_bonus"
    min_level: 1
    max_level: 20
    effects:
      - type: "damage_bonus"
        range: [1, 5]
        scaling: 0.2

  elemental_damage:
    name: "Elemental"
    type: "damage_type"
    min_level: 3
    max_level: 20
    effects:
      - type: "elemental_damage"
        elements: ["fire", "cold", "lightning", "acid"]
        range: [2, 12]

rarity_modifiers:
  common:
    stat_multiplier: 1.0
    enchantment_chance: 0.0
    max_enchantments: 0
    value_multiplier: 1.0
    name_prefixes: []
    name_suffixes: []

  uncommon:
    stat_multiplier: 1.1
    enchantment_chance: 0.3
    max_enchantments: 1
    value_multiplier: 2.0
    name_prefixes: ["Fine", "Quality"]
    name_suffixes: []

  rare:
    stat_multiplier: 1.25
    enchantment_chance: 0.6
    max_enchantments: 2
    value_multiplier: 5.0
    name_prefixes: ["Superior", "Masterwork"]
    name_suffixes: ["of Power"]

  epic:
    stat_multiplier: 1.5
    enchantment_chance: 0.8
    max_enchantments: 3
    value_multiplier: 10.0
    name_prefixes: ["Legendary", "Epic"]
    name_suffixes: ["of the Masters", "of Legend"]

  legendary:
    stat_multiplier: 2.0
    enchantment_chance: 1.0
    max_enchantments: 4
    value_multiplier: 25.0
    name_prefixes: ["Mythic", "Divine"]
    name_suffixes: ["of the Gods", "of Infinity"]
```

### Phase 3: Implement Level/Dungeon Generation

#### 3.1 Create Level Generator (`pkg/pcg/levels/`)

**File: `pkg/pcg/levels/generator.go`**
```go
package levels

import (
    "context"
    "goldbox-rpg/pkg/game"
    "goldbox-rpg/pkg/pcg"
)

// RoomCorridorGenerator creates levels using room-corridor approach
type RoomCorridorGenerator struct {
    version string
    roomGenerators map[pcg.RoomType]RoomGenerator
}

// RoomGenerator interface for different room types
type RoomGenerator interface {
    GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error)
}

// NewRoomCorridorGenerator creates a new room-corridor level generator
func NewRoomCorridorGenerator() *RoomCorridorGenerator {
    rcg := &RoomCorridorGenerator{
        version: "1.0.0",
        roomGenerators: make(map[pcg.RoomType]RoomGenerator),
    }
    
    // Register default room generators
    rcg.registerDefaultRoomGenerators()
    return rcg
}

// Generate implements the Generator interface
func (rcg *RoomCorridorGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
    // Implementation: Extract level parameters and delegate to GenerateLevel
}

// GenerateLevel creates a complete dungeon level
func (rcg *RoomCorridorGenerator) GenerateLevel(ctx context.Context, params pcg.LevelParams) (*game.Level, error) {
    // Implementation:
    // 1. Plan room layout using space partitioning
    // 2. Generate individual rooms
    // 3. Create corridor connections
    // 4. Add special features and encounters
    // 5. Validate connectivity and balance
    // 6. Convert to game.Level format
}

// generateRoomLayout creates the spatial layout of rooms
func (rcg *RoomCorridorGenerator) generateRoomLayout(width, height int, params pcg.LevelParams, genCtx *pcg.GenerationContext) ([]*pcg.RoomLayout, error) {
    // Implementation: Use BSP (Binary Space Partitioning) or similar algorithm
}
```

**File: `pkg/pcg/levels/rooms.go`**
```go
package levels

import (
    "goldbox-rpg/pkg/pcg"
)

// CombatRoomGenerator creates combat encounter rooms
type CombatRoomGenerator struct{}

func (crg *CombatRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
    // Implementation:
    // 1. Create basic room shape
    // 2. Add tactical features (cover, elevation)
    // 3. Place enemy spawn points
    // 4. Add environmental hazards based on theme
}

// TreasureRoomGenerator creates treasure and loot rooms
type TreasureRoomGenerator struct{}

func (trg *TreasureRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
    // Implementation:
    // 1. Create secure room layout
    // 2. Add treasure containers
    // 3. Place guardian encounters
    // 4. Add trap mechanisms
}

// PuzzleRoomGenerator creates puzzle and challenge rooms
type PuzzleRoomGenerator struct{}

func (prg *PuzzleRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
    // Implementation:
    // 1. Design puzzle layout
    // 2. Place interactive elements
    // 3. Create success/failure conditions
    // 4. Add thematic decorations
}

// BossRoomGenerator creates climactic boss encounter rooms
type BossRoomGenerator struct{}

func (brg *BossRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
    // Implementation:
    // 1. Create large, dramatic space
    // 2. Add multi-phase encounter areas
    // 3. Place environmental interaction points
    // 4. Design escape/entry routes
}
```

**File: `pkg/pcg/levels/corridors.go`**
```go
package levels

import (
    "goldbox-rpg/pkg/game"
    "goldbox-rpg/pkg/pcg"
)

// CorridorPlanner handles corridor generation between rooms
type CorridorPlanner struct {
    style pcg.CorridorStyle
}

// NewCorridorPlanner creates a corridor planner with specified style
func NewCorridorPlanner(style pcg.CorridorStyle) *CorridorPlanner {
    return &CorridorPlanner{style: style}
}

// ConnectRooms creates corridors between all rooms
func (cp *CorridorPlanner) ConnectRooms(rooms []*pcg.RoomLayout, levelBounds pcg.Rectangle, genCtx *pcg.GenerationContext) ([]*pcg.Corridor, error) {
    // Implementation:
    // 1. Create minimum spanning tree of room connections
    // 2. Add additional connections for redundancy
    // 3. Generate corridor paths using specified style
    // 4. Handle elevation changes and special features
}

// generateCorridorPath creates a single corridor between two points
func (cp *CorridorPlanner) generateCorridorPath(start, end game.Position, obstacles []pcg.Rectangle, genCtx *pcg.GenerationContext) (*pcg.Corridor, error) {
    switch cp.style {
    case pcg.CorridorStraight:
        return cp.generateStraightCorridor(start, end, genCtx)
    case pcg.CorridorWindy:
        return cp.generateWindyCorridor(start, end, obstacles, genCtx)
    case pcg.CorridorMaze:
        return cp.generateMazeCorridor(start, end, obstacles, genCtx)
    case pcg.CorridorOrganic:
        return cp.generateOrganicCorridor(start, end, obstacles, genCtx)
    default:
        return cp.generateStraightCorridor(start, end, genCtx)
    }
}
```

### Phase 4: Implement Quest Generation System

#### 4.1 Create Quest Generator (`pkg/pcg/quests/`)

**File: `pkg/pcg/quests/generator.go`**
```go
package quests

import (
    "context"
    "goldbox-rpg/pkg/game"
    "goldbox-rpg/pkg/pcg"
)

// ObjectiveBasedGenerator creates quests using objective templates
type ObjectiveBasedGenerator struct {
    version           string
    objectiveTemplates map[pcg.QuestType][]*ObjectiveTemplate
    narrativeEngine   *NarrativeEngine
}

// ObjectiveTemplate defines the structure of quest objectives
type ObjectiveTemplate struct {
    Type         string   `yaml:"type"`
    Description  string   `yaml:"description"`
    Requirements []string `yaml:"requirements"`
    Targets      []string `yaml:"targets"`
    Quantities   [2]int   `yaml:"quantities"`
    Rewards      []string `yaml:"rewards"`
}

// NewObjectiveBasedGenerator creates a new objective-based quest generator
func NewObjectiveBasedGenerator() *ObjectiveBasedGenerator {
    return &ObjectiveBasedGenerator{
        version:           "1.0.0",
        objectiveTemplates: make(map[pcg.QuestType][]*ObjectiveTemplate),
        narrativeEngine:   NewNarrativeEngine(),
    }
}

// Generate implements the Generator interface
func (obg *ObjectiveBasedGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
    // Implementation: Extract quest parameters and delegate to GenerateQuest
}

// GenerateQuest creates a quest with objectives and narrative
func (obg *ObjectiveBasedGenerator) GenerateQuest(ctx context.Context, questType pcg.QuestType, params pcg.QuestParams) (*game.Quest, error) {
    // Implementation:
    // 1. Select appropriate objective templates
    // 2. Generate specific targets from world state
    // 3. Create narrative context
    // 4. Determine rewards based on difficulty
    // 5. Build quest structure
}

// GenerateQuestChain creates a series of connected quests
func (obg *ObjectiveBasedGenerator) GenerateQuestChain(ctx context.Context, chainLength int, params pcg.QuestParams) ([]*game.Quest, error) {
    // Implementation: Create narrative arc across multiple quests
}
```

**File: `pkg/pcg/quests/objectives.go`** ✅ **COMPLETED**
```go
package quests

import (
)

// ObjectiveGenerator creates specific quest objectives.
// ObjectiveGenerator is stateless and does not require world state.
type ObjectiveGenerator struct{}

// NewObjectiveGenerator creates an objective generator
func NewObjectiveGenerator() *ObjectiveGenerator {
    return &ObjectiveGenerator{}
}

// GenerateKillObjective creates kill/defeat objectives
func (og *ObjectiveGenerator) GenerateKillObjective(difficulty int, genCtx *pcg.GenerationContext) (*pcg.QuestObjective, error) {
    // ✅ IMPLEMENTED: Complete functionality for generating kill objectives
    // - Validates difficulty and generation context parameters
    // - Selects appropriate enemy types based on difficulty level (1-10)
    // - Determines quantity based on challenge rating
    // - Chooses random location from available areas
    // - Creates properly structured quest objective with conditions
}

// GenerateFetchObjective creates item retrieval objectives
func (og *ObjectiveGenerator) GenerateFetchObjective(playerLevel int, genCtx *pcg.GenerationContext) (*pcg.QuestObjective, error) {
    // ✅ IMPLEMENTED: Complete functionality for generating fetch objectives
    // - Validates player level (1-20) and generation context
    // - Selects appropriate item types for player level
    // - Handles common items with variable quantities (1-3)
    // - Ensures different pickup and delivery locations
    // - Creates objective with pickup/delivery conditions
}

// GenerateExploreObjective creates exploration objectives
func (og *ObjectiveGenerator) GenerateExploreObjective(genCtx *pcg.GenerationContext) (*pcg.QuestObjective, error) {
    // ✅ IMPLEMENTED: Complete functionality for generating exploration objectives
    // - Validates generation context
    // - Identifies unexplored areas from predefined list
    // - Sets discovery requirements (70-100% completion)
    // - Adds optional sub-objectives (hidden areas, secrets)
    // - Creates objective with area and percentage conditions
}

// ✅ IMPLEMENTED: Complete helper functions
// - selectEnemyTypesForDifficulty: Maps difficulty levels to enemy types
// - selectItemTypesForLevel: Maps player levels to appropriate item types
// - getAvailableLocations: Returns list of available quest locations
// - getUnexploredAreas: Returns list of unexplored areas
// - isCommonItem: Determines if item type is common for quantity scaling
```

**File: `pkg/pcg/quests/narratives.go`**
```go
package quests

import (
    "goldbox-rpg/pkg/pcg"
)

// NarrativeEngine generates quest stories and dialogue
type NarrativeEngine struct {
    storyTemplates map[pcg.QuestType][]*StoryTemplate
    characterPool  []*NPCTemplate
}

// StoryTemplate defines narrative structure
type StoryTemplate struct {
    Theme       string   `yaml:"theme"`
    Setup       string   `yaml:"setup"`
    Motivation  string   `yaml:"motivation"`
    Climax      string   `yaml:"climax"`
    Resolution  string   `yaml:"resolution"`
    Characters  []string `yaml:"characters"`
    Locations   []string `yaml:"locations"`
}

// NPCTemplate defines quest-giver characteristics
type NPCTemplate struct {
    Archetype   string   `yaml:"archetype"`
    Personality []string `yaml:"personality"`
    Motivations []string `yaml:"motivations"`
    Speech      []string `yaml:"speech_patterns"`
}

// NewNarrativeEngine creates a new narrative engine
func NewNarrativeEngine() *NarrativeEngine {
    return &NarrativeEngine{
        storyTemplates: make(map[pcg.QuestType][]*StoryTemplate),
        characterPool:  make([]*NPCTemplate, 0),
    }
}

// GenerateQuestNarrative creates story context for a quest
func (ne *NarrativeEngine) GenerateQuestNarrative(questType pcg.QuestType, objectives []*pcg.QuestObjective, genCtx *pcg.GenerationContext) (*QuestNarrative, error) {
    // Implementation:
    // 1. Select appropriate story template
    // 2. Generate quest-giver character
    // 3. Create dialogue and descriptions
    // 4. Tie narrative to objectives
    // 5. Generate completion dialogue
}

// QuestNarrative holds the complete story context
type QuestNarrative struct {
    Title         string `yaml:"title"`
    Description   string `yaml:"description"`
    QuestGiver    string `yaml:"quest_giver"`
    StartDialogue string `yaml:"start_dialogue"`
    EndDialogue   string `yaml:"end_dialogue"`
    Lore          string `yaml:"lore"`
}
```

### Phase 5: API Integration and Server Endpoints ✅

**Status: COMPLETED** - All PCG API endpoints have been implemented and tested.

The following endpoints were added to `pkg/server/handlers.go`:

- `handleGenerateContent` - Generates procedural content on demand
- `handleRegenerateTerrain` - Regenerates terrain for specific locations  
- `handleGenerateItems` - Generates items with specified parameters
- `handleGenerateLevel` - Generates complete level layouts
- `handleGenerateQuest` - Generates quest content
- `handleGetPCGStats` - Retrieves PCG system statistics
- `handleValidateContent` - Validates generated content

#### Implementation Details:

```go
// PCG handler methods added to RPCServer struct in pkg/server/handlers.go
// All methods follow the JSON-RPC 2.0 specification
// Comprehensive test suite in pkg/server/pcg_handlers_test.go

// handleGenerateContent generates procedural content on demand
func (s *Server) handleGenerateContent(sessionID string, params map[string]interface{}) (interface{}, error) {
    session := s.getSession(sessionID)
    if session == nil {
        return nil, fmt.Errorf("invalid session")
    }

    contentType, ok := params["content_type"].(string)
    if !ok {
        return nil, fmt.Errorf("content_type parameter required")
    }

    locationID, ok := params["location_id"].(string)
    if !ok {
        return nil, fmt.Errorf("location_id parameter required")
    }

    difficulty, _ := params["difficulty"].(float64)
    if difficulty == 0 {
        difficulty = 5 // Default difficulty
    }

    ctx := context.Background()
    content, err := s.pcgManager.RegenerateContentForLocation(
        ctx, 
        locationID, 
        pcg.ContentType(contentType),
    )
    if err != nil {
        return nil, fmt.Errorf("generation failed: %w", err)
    }

    // Integrate into world
    if err := s.pcgManager.IntegrateContentIntoWorld(content, locationID); err != nil {
        return nil, fmt.Errorf("integration failed: %w", err)
    }

    return map[string]interface{}{
        "success":     true,
        "content_id":  extractContentID(content),
        "location_id": locationID,
        "type":        contentType,
    }, nil
}

// handleRegenerateTerrain regenerates terrain for a specific area
func (s *Server) handleRegenerateTerrain(sessionID string, params map[string]interface{}) (interface{}, error) {
    session := s.getSession(sessionID)
    if session == nil {
        return nil, fmt.Errorf("invalid session")
    }

    levelID, ok := params["level_id"].(string)
    if !ok {
        return nil, fmt.Errorf("level_id parameter required")
    }

    width, _ := params["width"].(float64)
    height, _ := params["height"].(float64)
    biome, _ := params["biome"].(string)
    difficulty, _ := params["difficulty"].(float64)

    if width == 0 { width = 50 }
    if height == 0 { height = 50 }
    if biome == "" { biome = "dungeon" }
    if difficulty == 0 { difficulty = 5 }

    ctx := context.Background()
    gameMap, err := s.pcgManager.GenerateTerrainForLevel(
        ctx,
        levelID,
        int(width), int(height),
        pcg.BiomeType(biome),
        int(difficulty),
    )
    if err != nil {
        return nil, fmt.Errorf("terrain generation failed: %w", err)
    }

    return map[string]interface{}{
        "success":   true,
        "level_id":  levelID,
        "width":     gameMap.Width,
        "height":    gameMap.Height,
        "tile_data": gameMap.Tiles,
    }, nil
}

// handleGenerateItems generates items for a location
func (s *Server) handleGenerateItems(sessionID string, params map[string]interface{}) (interface{}, error) {
    session := s.getSession(sessionID)
    if session == nil {
        return nil, fmt.Errorf("invalid session")
    }

    locationID, ok := params["location_id"].(string)
    if !ok {
        return nil, fmt.Errorf("location_id parameter required")
    }

    count, _ := params["count"].(float64)
    minRarity, _ := params["min_rarity"].(string)
    maxRarity, _ := params["max_rarity"].(string)
    playerLevel, _ := params["player_level"].(float64)

    if count == 0 { count = 3 }
    if minRarity == "" { minRarity = "common" }
    if maxRarity == "" { maxRarity = "rare" }
    if playerLevel == 0 { playerLevel = 5 }

    ctx := context.Background()
    items, err := s.pcgManager.GenerateItemsForLocation(
        ctx,
        locationID,
        int(count),
        pcg.RarityTier(minRarity),
        pcg.RarityTier(maxRarity),
        int(playerLevel),
    )
    if err != nil {
        return nil, fmt.Errorf("item generation failed: %w", err)
    }

    // Convert items to response format
    itemData := make([]map[string]interface{}, len(items))
    for i, item := range items {
        itemData[i] = map[string]interface{}{
            "id":       item.ID,
            "name":     item.Name,
            "type":     item.Type,
            "value":    item.Value,
            "weight":   item.Weight,
            "properties": item.Properties,
        }
    }

    return map[string]interface{}{
        "success":     true,
        "location_id": locationID,
        "items":       itemData,
        "count":       len(items),
    }, nil
}
```

#### 5.2 Register PCG Methods (`pkg/server/server.go`)

```go
// Add PCG method registration to the server initialization

func (s *Server) registerPCGMethods() {
    s.registerMethod("generateContent", s.handleGenerateContent)
    s.registerMethod("regenerateTerrain", s.handleRegenerateTerrain)
    s.registerMethod("generateItems", s.handleGenerateItems)
    s.registerMethod("generateLevel", s.handleGenerateLevel)
    s.registerMethod("generateQuest", s.handleGenerateQuest)
    s.registerMethod("getPCGStats", s.handleGetPCGStats)
    s.registerMethod("validateContent", s.handleValidateContent)
}

// Call registerPCGMethods() in NewServer() after other method registrations
```

### Phase 6: Testing and Validation

#### 6.1 Unit Tests for Each Generator

**File: `pkg/pcg/terrain/generator_test.go`**
```go
package terrain

import (
    "context"
    "testing"
    "time"
    
    "goldbox-rpg/pkg/pcg"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCellularAutomataGenerator(t *testing.T) {
    tests := []struct {
        name     string
        width    int
        height   int
        biome    pcg.BiomeType
        seed     int64
        density  float64
        wantErr  bool
    }{
        {
            name:    "basic cave generation",
            width:   20,
            height:  20,
            biome:   pcg.BiomeCave,
            seed:    12345,
            density: 0.45,
            wantErr: false,
        },
        {
            name:    "dungeon generation",
            width:   30,
            height:  30,
            biome:   pcg.BiomeDungeon,
            seed:    54321,
            density: 0.4,
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            generator := NewCellularAutomataGenerator()
            
            params := pcg.TerrainParams{
                GenerationParams: pcg.GenerationParams{
                    Seed:       tt.seed,
                    Difficulty: 5,
                    Timeout:    10 * time.Second,
                    Constraints: map[string]interface{}{
                        "width":          tt.width,
                        "height":         tt.height,
                        "terrain_params": pcg.TerrainParams{},
                    },
                },
                BiomeType:    tt.biome,
                Density:      tt.density,
                Connectivity: pcg.ConnectivityModerate,
            }

            ctx := context.Background()
            gameMap, err := generator.GenerateTerrain(ctx, tt.width, tt.height, params)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.NotNil(t, gameMap)
            assert.Equal(t, tt.width, gameMap.Width)
            assert.Equal(t, tt.height, gameMap.Height)
            assert.Len(t, gameMap.Tiles, tt.height)
            
            // Test deterministic generation
            gameMap2, err := generator.GenerateTerrain(ctx, tt.width, tt.height, params)
            require.NoError(t, err)
            assert.Equal(t, gameMap.Tiles, gameMap2.Tiles)
        })
    }
}

func TestTerrainConnectivity(t *testing.T) {
    generator := NewCellularAutomataGenerator()
    
    params := pcg.TerrainParams{
        GenerationParams: pcg.GenerationParams{
            Seed:       12345,
            Difficulty: 5,
            Timeout:    10 * time.Second,
        },
        BiomeType:    pcg.BiomeCave,
        Density:      0.3, // Lower density for better connectivity
        Connectivity: pcg.ConnectivityHigh,
    }

    ctx := context.Background()
    gameMap, err := generator.GenerateTerrain(ctx, 25, 25, params)
    require.NoError(t, err)

    // Validate connectivity
    assert.True(t, generator.ValidateConnectivity(gameMap))
}
```

#### 6.2 Integration Tests

**File: `pkg/pcg/integration_test.go`**
```go
package pcg

import (
    "context"
    "testing"
    "time"
    
    "goldbox-rpg/pkg/game"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPCGManagerIntegration(t *testing.T) {
    // Create test world
    world := &game.World{
        Objects: make(map[string]game.GameObject),
        Levels:  []game.Level{},
        Players: make(map[string]*game.Player),
    }

    logger := logrus.New()
    logger.SetLevel(logrus.WarnLevel) // Reduce test noise

    pcgManager := NewPCGManager(world, logger)
    pcgManager.InitializeWithSeed(12345)

    // Test terrain generation and integration
    ctx := context.Background()
    gameMap, err := pcgManager.GenerateTerrainForLevel(
        ctx, "test_level", 15, 15, BiomeCave, 3,
    )
    require.NoError(t, err)
    assert.NotNil(t, gameMap)

    // Test content validation
    result, err := pcgManager.ValidateGeneratedContent(gameMap)
    require.NoError(t, err)
    assert.True(t, result.IsValid())

    // Test world integration
    err = pcgManager.IntegrateContentIntoWorld(gameMap, "test_level")
    assert.NoError(t, err)
}

func TestEndToEndGeneration(t *testing.T) {
    world := createTestWorld()
    pcgManager := NewPCGManager(world, logrus.New())
    pcgManager.InitializeWithSeed(54321)

    ctx := context.Background()

    // Generate terrain
    gameMap, err := pcgManager.GenerateTerrainForLevel(
        ctx, "dungeon_1", 20, 20, BiomeDungeon, 5,
    )
    require.NoError(t, err)

    // Generate items for the terrain
    items, err := pcgManager.GenerateItemsForLocation(
        ctx, "dungeon_1", 5, RarityCommon, RarityEpic, 8,
    )
    require.NoError(t, err)
    assert.Len(t, items, 5)

    // Integrate everything
    err = pcgManager.IntegrateContentIntoWorld(gameMap, "dungeon_1")
    require.NoError(t, err)

    for _, item := range items {
        err = pcgManager.IntegrateContentIntoWorld(item, "dungeon_1")
        require.NoError(t, err)
    }

    // Verify integration
    assert.True(t, len(world.Objects) >= 5)
}

func createTestWorld() *game.World {
    return &game.World{
        Objects: make(map[string]game.GameObject),
        Levels:  []game.Level{},
        Players: make(map[string]*game.Player),
        Width:   100,
        Height:  100,
    }
}
```

### Phase 7: Performance and Optimization

#### 7.1 Performance Monitoring

**File: `pkg/pcg/metrics.go`**
```go
package pcg

import (
    "sync"
    "time"
)

// GenerationMetrics tracks performance statistics
type GenerationMetrics struct {
    mu                sync.RWMutex
    GenerationCounts  map[ContentType]int64         `json:"generation_counts"`
    AverageTimings    map[ContentType]time.Duration `json:"average_timings"`
    ErrorCounts       map[ContentType]int64         `json:"error_counts"`
    CacheHits         int64                         `json:"cache_hits"`
    CacheMisses       int64                         `json:"cache_misses"`
    TotalGenerations  int64                         `json:"total_generations"`
}

// NewGenerationMetrics creates a new metrics tracker
func NewGenerationMetrics() *GenerationMetrics {
    return &GenerationMetrics{
        GenerationCounts: make(map[ContentType]int64),
        AverageTimings:   make(map[ContentType]time.Duration),
        ErrorCounts:      make(map[ContentType]int64),
    }
}

// RecordGeneration records a successful generation
func (gm *GenerationMetrics) RecordGeneration(contentType ContentType, duration time.Duration) {
    gm.mu.Lock()
    defer gm.mu.Unlock()
    
    gm.GenerationCounts[contentType]++
    gm.TotalGenerations++
    
    // Update rolling average
    if current, exists := gm.AverageTimings[contentType]; exists {
        count := gm.GenerationCounts[contentType]
        gm.AverageTimings[contentType] = (current*time.Duration(count-1) + duration) / time.Duration(count)
    } else {
        gm.AverageTimings[contentType] = duration
    }
}

// RecordError records a generation error
func (gm *GenerationMetrics) RecordError(contentType ContentType) {
    gm.mu.Lock()
    defer gm.mu.Unlock()
    
    gm.ErrorCounts[contentType]++
}

// GetStats returns current performance statistics
func (gm *GenerationMetrics) GetStats() map[string]interface{} {
    gm.mu.RLock()
    defer gm.mu.RUnlock()
    
    return map[string]interface{}{
        "generation_counts": gm.GenerationCounts,
        "average_timings":   gm.AverageTimings,
        "error_counts":      gm.ErrorCounts,
        "cache_hits":        gm.CacheHits,
        "cache_misses":      gm.CacheMisses,
        "total_generations": gm.TotalGenerations,
    }
}
```

### Implementation Priority

1. **High Priority (Complete First):**
   - ✅ Complete terrain generation system with all biomes
   - ✅ Implement item generation with templates and enchantments  
   - Add basic level generation (room-corridor approach)
   - Create unit tests for all generators

2. **Medium Priority:**
   - Implement quest generation system
   - Add API endpoints for PCG functionality
   - Create comprehensive configuration system
   - Add performance monitoring and metrics

3. **Low Priority (Future Enhancements):**
   - Advanced algorithms (Wave Function Collapse, etc.)
   - Machine learning integration
   - Real-time generation streaming
   - Content marketplace features

### Success Criteria

- [x] Terrain generation interfaces implemented with working examples
- [x] Item generation system with templates and enchantments
- [x] Deterministic generation (same seed = same output)
- [x] Thread-safe concurrent generation for completed systems
- [x] API integration and server endpoints implemented
- [x] Proper integration with existing game systems (PCG components)
- [x] >80% test coverage for terrain and item PCG code
- [x] Content validation system with proper type handling
- [ ] Performance benchmarks under 100ms for basic generation
- [ ] Complete API documentation with examples
- [x] YAML configuration system functional for items

### Notes for Implementation

1. **Follow Existing Patterns**: Use the same mutex patterns, logging style, and error handling as the existing codebase
2. **Maintain Compatibility**: Ensure all new code integrates seamlessly with existing game systems
3. **Test Extensively**: Write comprehensive tests including edge cases and performance tests
4. **Document Thoroughly**: Update README and create examples for each generator type
5. **Performance First**: Profile generation code and optimize hot paths
6. **Validation Always**: Never integrate content without validation checks

This guide provides a complete roadmap for implementing the remaining PCG functionality while maintaining consistency with the existing GoldBox RPG Engine architecture.
