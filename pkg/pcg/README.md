# Procedural Content Generation (PCG) System

## Overview

The Procedural Content Generation (PCG) subsystem provides comprehensive tools for generating game content dynamically in the GoldBox RPG Engine. This system follows the established architectural patterns of the engine with thread-safe operations, deterministic seeding, and YAML-based configuration.

## Features

- **Multi-Type Content Generation**: Terrain, items, levels, quests, and NPCs
- **Deterministic Seeding**: Reproducible content using seed-based generation
- **Biome-Aware Terrain**: Context-sensitive terrain generation for different environments
- **Template-Based Items**: Flexible item generation with rarity tiers and enchantments
- **Dungeon Layout**: Complete level generation with rooms, corridors, and features
- **Quest Generation**: Dynamic quest creation with objectives and rewards
- **Validation System**: Content validation before world integration
- **Performance Monitoring**: Timeout handling and generation statistics

## Architecture

### Package Structure

```
pkg/pcg/
├── interfaces.go        # Core generator interfaces
├── registry.go          # Generator registration and management
├── types.go             # Type definitions and enums
├── seed.go              # Deterministic seeding system
├── validation.go        # Content validation
├── manager.go           # Main PCG coordinator
├── terrain/             # Terrain generation implementations
├── items/               # Item generation implementations
├── levels/              # Level/dungeon generation implementations
├── quests/              # Quest generation implementations
└── utils/               # Utility functions and algorithms
```

### Core Interfaces

#### Generator Interface
All PCG implementations must satisfy the `Generator` interface:

```go
type Generator interface {
    Generate(ctx context.Context, params GenerationParams) (interface{}, error)
    GetType() ContentType
    GetVersion() string
    Validate(params GenerationParams) error
}
```

#### Specialized Generators
- `TerrainGenerator`: Generates 2D terrain maps with biome awareness
- `ItemGenerator`: Creates items with templates and procedural properties
- `LevelGenerator`: Builds complete dungeon levels with rooms and corridors
- `QuestGenerator`: Generates quests with objectives and narratives

### Thread Safety

The PCG system follows the established thread-safety patterns from the main game engine:

- **Registry**: Uses `sync.RWMutex` for concurrent generator access
- **Seed Manager**: Deterministic seed derivation without locks
- **Generation Context**: Per-generation RNG instances for thread isolation
- **Validation**: Stateless validators for concurrent content checking

## Usage Examples

### Basic Setup

```go
// Initialize PCG manager with world reference
world := &game.World{...}
logger := logrus.New()
pcgManager := pcg.NewPCGManager(world, logger)

// Set deterministic seed for reproducible generation
pcgManager.InitializeWithSeed(12345)

// Register default generators
err := pcgManager.RegisterDefaultGenerators()
if err != nil {
    log.Fatal("Failed to register generators:", err)
}
```

### Terrain Generation

```go
ctx := context.Background()

// Generate a cave system for a dungeon level
gameMap, err := pcgManager.GenerateTerrainForLevel(
    ctx,
    "dungeon_level_1",  // Unique level ID
    50, 50,             // Dimensions
    pcg.BiomeCave,      // Biome type
    5,                  // Difficulty level
)

if err != nil {
    log.Fatal("Terrain generation failed:", err)
}

// Validate and integrate into world
if err := pcgManager.IntegrateContentIntoWorld(gameMap, "dungeon_level_1"); err != nil {
    log.Fatal("Integration failed:", err)
}
```

### Item Generation

```go
// Generate treasure items for a specific location
items, err := pcgManager.GenerateItemsForLocation(
    ctx,
    "treasure_room_A",     // Location ID
    5,                     // Number of items
    pcg.RarityUncommon,    // Minimum rarity
    pcg.RarityEpic,        // Maximum rarity
    8,                     // Player level for scaling
)

if err != nil {
    log.Fatal("Item generation failed:", err)
}

// Integrate all items into the world
for _, item := range items {
    if err := pcgManager.IntegrateContentIntoWorld(item, "treasure_room_A"); err != nil {
        log.Error("Failed to integrate item:", err)
    }
}
```

### Complete Dungeon Generation

```go
// Generate a complete dungeon level
level, err := pcgManager.GenerateDungeonLevel(
    ctx,
    "main_dungeon_floor_2", // Level ID
    8,                      // Minimum rooms
    15,                     // Maximum rooms
    pcg.ThemeClassic,       // Dungeon theme
    10,                     // Difficulty
)

if err != nil {
    log.Fatal("Level generation failed:", err)
}
```

### Custom Generator Registration

```go
// Create a custom terrain generator
customGenerator := &MyCustomTerrainGenerator{
    version: "1.0.0",
}

// Register with the PCG system
registry := pcgManager.GetRegistry()
err := registry.RegisterGenerator("my_custom_terrain", customGenerator)
if err != nil {
    log.Error("Failed to register custom generator:", err)
}

// Use the custom generator
params := pcg.TerrainParams{
    GenerationParams: pcg.GenerationParams{
        Seed:       12345,
        Difficulty: 5,
        Timeout:    30 * time.Second,
        Constraints: map[string]interface{}{
            "width":          80,
            "height":         60,
            "terrain_params": terrainParams,
        },
    },
    BiomeType:    pcg.BiomeForest,
    Density:      0.6,
    Connectivity: pcg.ConnectivityHigh,
}

factory := pcgManager.GetFactory()
result, err := factory.GenerateTerrain(ctx, "my_custom_terrain", params)
```

## Integration with Game Systems

### Event System Integration

```go
// PCG operations can emit events through the existing event system
func (pcg *PCGManager) GenerateWithEvents(ctx context.Context, contentType pcg.ContentType) {
    // Emit generation start event
    startEvent := &game.Event{
        Type: "pcg_generation_started",
        Data: map[string]interface{}{
            "content_type": contentType,
            "timestamp":    time.Now(),
        },
    }
    pcg.world.EventSystem.Emit(startEvent)
    
    // ... perform generation ...
    
    // Emit completion event
    completeEvent := &game.Event{
        Type: "pcg_generation_completed",
        Data: map[string]interface{}{
            "content_type": contentType,
            "success":      true,
        },
    }
    pcg.world.EventSystem.Emit(completeEvent)
}
```

### Spatial Index Integration

Generated content automatically integrates with the existing spatial indexing system:

```go
// Items are automatically added to spatial index during integration
item := &game.Item{
    ID:       "generated_sword_001",
    Name:     "Enchanted Blade",
    Position: game.Position{X: 10, Y: 15, Level: 1},
}

// Integration automatically handles spatial index updates
pcgManager.IntegrateContentIntoWorld(item, "dungeon_room_5")
```

### JSON-RPC API Integration

Add PCG endpoints to the existing JSON-RPC server:

```go
// In pkg/server/handlers.go
func (s *Server) handleGenerateContent(sessionID string, params map[string]interface{}) (interface{}, error) {
    contentType := params["content_type"].(string)
    locationID := params["location_id"].(string)
    
    ctx := context.Background()
    content, err := s.pcgManager.RegenerateContentForLocation(ctx, locationID, pcg.ContentType(contentType))
    if err != nil {
        return nil, err
    }
    
    // Integrate into world
    if err := s.pcgManager.IntegrateContentIntoWorld(content, locationID); err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "success":     true,
        "content_id":  extractContentID(content),
        "location_id": locationID,
    }, nil
}
```

## Configuration

### YAML Configuration Example

```yaml
# data/pcg/terrain_config.yaml
terrain_generators:
  cellular_automata:
    version: "1.0.0"
    default_params:
      density: 0.45
      connectivity: moderate
      iterations: 5
    biome_settings:
      cave:
        density: 0.45
        water_level: 0.1
        roughness: 0.6
      dungeon:
        density: 0.4
        water_level: 0.05
        roughness: 0.3

item_generators:
  template_based:
    version: "1.0.0"
    templates_file: "data/pcg/item_templates.yaml"
    enchantment_rate: 0.2
    unique_chance: 0.05
```

### Loading Configuration

```go
// Load PCG configuration during server startup
func loadPCGConfig(pcgManager *pcg.PCGManager) error {
    configFiles := []string{
        "data/pcg/terrain_config.yaml",
        "data/pcg/item_config.yaml",
        "data/pcg/level_config.yaml",
    }
    
    for _, configFile := range configFiles {
        if err := loadConfigFile(pcgManager, configFile); err != nil {
            return fmt.Errorf("failed to load %s: %w", configFile, err)
        }
    }
    
    return nil
}
```

## Performance Considerations

### Timeout Management

All generation operations support context-based timeouts:

```go
// Set generation timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Generation will be cancelled if it exceeds timeout
content, err := pcgManager.GenerateContent(ctx, params)
if err == context.DeadlineExceeded {
    log.Warn("Generation timed out")
}
```

### Memory Management

- **Streaming Generation**: Large content can be generated in chunks
- **Validation Caching**: Validation results cached for repeated content
- **Generator Pooling**: Reuse generator instances for performance

### Deterministic Performance

- **Seed-Based Caching**: Cache generated content by seed for instant retrieval
- **Lazy Loading**: Generate content only when needed
- **Background Generation**: Pre-generate content for upcoming areas

## Testing

### Unit Testing Example

```go
func TestCellularAutomataGenerator(t *testing.T) {
    generator := terrain.NewCellularAutomataGenerator()
    
    params := pcg.TerrainParams{
        GenerationParams: pcg.GenerationParams{
            Seed:       12345,
            Difficulty: 5,
            Timeout:    10 * time.Second,
        },
        BiomeType:    pcg.BiomeCave,
        Density:      0.45,
        Connectivity: pcg.ConnectivityModerate,
    }
    
    ctx := context.Background()
    gameMap, err := generator.GenerateTerrain(ctx, 20, 20, params)
    
    assert.NoError(t, err)
    assert.NotNil(t, gameMap)
    assert.Equal(t, 20, gameMap.Width)
    assert.Equal(t, 20, gameMap.Height)
    
    // Test deterministic generation
    gameMap2, err := generator.GenerateTerrain(ctx, 20, 20, params)
    assert.NoError(t, err)
    assert.Equal(t, gameMap.Tiles, gameMap2.Tiles)
}
```

### Integration Testing

```go
func TestPCGManagerIntegration(t *testing.T) {
    world := &game.World{
        Objects: make(map[string]game.GameObject),
        Levels:  []game.Level{},
    }
    
    logger := logrus.New()
    pcgManager := pcg.NewPCGManager(world, logger)
    pcgManager.InitializeWithSeed(54321)
    
    // Test terrain generation and integration
    ctx := context.Background()
    gameMap, err := pcgManager.GenerateTerrainForLevel(ctx, "test_level", 10, 10, pcg.BiomeCave, 3)
    
    assert.NoError(t, err)
    assert.NotNil(t, gameMap)
    
    // Test integration
    err = pcgManager.IntegrateContentIntoWorld(gameMap, "test_level")
    assert.NoError(t, err)
}
```

## Error Handling

The PCG system follows the established error handling patterns:

```go
// Structured error types
type GenerationError struct {
    Type        string
    Generator   string
    ContentType pcg.ContentType
    Cause       error
}

func (e *GenerationError) Error() string {
    return fmt.Sprintf("PCG %s error in %s generator for %s: %v", 
        e.Type, e.Generator, e.ContentType, e.Cause)
}

// Error handling in generators
func (g *MyGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
    if err := g.Validate(params); err != nil {
        return nil, &GenerationError{
            Type:        "validation",
            Generator:   "my_generator",
            ContentType: g.GetType(),
            Cause:       err,
        }
    }
    
    // ... generation logic ...
}
```

## Dependencies

The PCG system uses the following dependencies:

- **Core Dependencies**: Standard library (`context`, `crypto/sha256`, `encoding/binary`)
- **Game Engine**: `goldbox-rpg/pkg/game` for integration with existing systems
- **Logging**: `github.com/sirupsen/logrus` for structured logging
- **YAML**: `gopkg.in/yaml.v3` for configuration files
- **UUID**: `github.com/google/uuid` for unique content identification

## Future Enhancements

### Planned Features

1. **Machine Learning Integration**: AI-driven content generation
2. **Biome Transition**: Smooth transitions between different biomes
3. **Dynamic Difficulty**: Adaptive difficulty based on player performance
4. **Content Evolution**: Content that changes over time
5. **Multi-threaded Generation**: Parallel generation for complex content
6. **Content Marketplace**: Sharing and importing community-generated content

### API Extensions

1. **WebSocket Streaming**: Real-time generation progress updates
2. **Batch Operations**: Generate multiple content pieces in one request
3. **Preview Mode**: Generate content for preview without world integration
4. **Undo/Redo**: Revert generation changes
5. **Content Versioning**: Track and manage content versions
