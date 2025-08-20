# Zero-Configuration Bootstrap System

The GoldBox RPG Engine Bootstrap System enables instant deployment of complete, playable RPG experiences without requiring any manual configuration files. This system automatically generates all necessary game content including world structure, NPCs, quests, dialogue, items, and spells.

## Overview

The bootstrap system addresses the "cold start" problem in RPG game deployment by providing:

- **Instant Deployment**: Complete games generated in under 1 second
- **Zero Configuration**: No manual setup or configuration files required
- **Deterministic Generation**: Reproducible worlds using seed-based generation
- **Scalable Complexity**: From simple tavern adventures to epic multi-region campaigns
- **Genre Variants**: Support for different fantasy themes and mechanical emphasis

## Architecture

### Core Components

1. **Bootstrap Engine** (`pkg/pcg/bootstrap.go`)
   - Main orchestrator for zero-configuration game generation
   - Coordinates all content generation systems
   - Manages configuration templates and parameter scaling

2. **Configuration Detection** (`DetectConfigurationPresence`)
   - Automatically detects if manual configuration files exist
   - Triggers bootstrap only when needed
   - Preserves existing configurations

3. **Parameter Templates** (`data/pcg/bootstrap_templates.yaml`)
   - Pre-configured settings for different game types
   - Balanced parameters for various play styles
   - Easy customization and extension

4. **Server Integration** (`cmd/server/main.go`)
   - Seamless integration with existing server startup
   - Automatic bootstrap on first run
   - No impact on normal operation

## Usage

### Automatic Server Bootstrap

The bootstrap system automatically activates when starting the server without existing configuration:

```bash
# First run - triggers automatic bootstrap
go run cmd/server/main.go

# Subsequent runs - uses existing configuration
go run cmd/server/main.go
```

### Manual Bootstrap Demo

Test different bootstrap configurations using the demo command:

```bash
# Basic usage with defaults
go run cmd/bootstrap-demo/main.go

# Custom configuration
go run cmd/bootstrap-demo/main.go \
  -length long \
  -complexity advanced \
  -genre grimdark \
  -players 6 \
  -level 3 \
  -seed 42 \
  -output my_campaign

# Help for all options
go run cmd/bootstrap-demo/main.go -help
```

### Programmatic Usage

```go
package main

import (
    "context"
    "goldbox-rpg/pkg/game"
    "goldbox-rpg/pkg/pcg"
)

func createGame() error {
    // Check if bootstrap is needed
    if !pcg.DetectConfigurationPresence("data") {
        // Create world and bootstrap configuration
        world := game.NewWorld()
        config := pcg.DefaultBootstrapConfig()
        
        // Customize configuration
        config.GameLength = pcg.GameLengthLong
        config.ComplexityLevel = pcg.ComplexityAdvanced
        config.GenreVariant = pcg.GenreGrimdark
        config.WorldSeed = 12345
        
        // Generate complete game
        bootstrap := pcg.NewBootstrap(config, world, logger)
        ctx := context.Background()
        
        _, err := bootstrap.GenerateCompleteGame(ctx)
        return err
    }
    
    return nil
}
```

## Configuration Options

### Game Length Types

- **Short** (`GameLengthShort`): 3-5 hours gameplay
  - 1 region, 2 factions, 10 NPCs, 5 quests
  - Single location focus, linear progression
  - Perfect for one-shot adventures

- **Medium** (`GameLengthMedium`): 8-12 hours gameplay
  - 3 regions, 4 factions, 20 NPCs, 12 quests
  - Regional scope, moderate branching
  - Standard campaign length

- **Long** (`GameLengthLong`): 20+ hours gameplay
  - 5 regions, 6 factions, 30 NPCs, 25 quests
  - Multi-region epic campaigns
  - Complex interconnected storylines

### Complexity Levels

- **Simple** (`ComplexitySimple`): Basic mechanics
  - Linear progression, minimal branching
  - Straightforward mechanics and interactions
  - Ideal for new players or quick sessions

- **Standard** (`ComplexityStandard`): Full mechanics
  - Moderate branching and complexity
  - All core RPG systems active
  - Balanced challenge and accessibility

- **Advanced** (`ComplexityAdvanced`): Maximum complexity
  - Complex interactions and systems
  - Multiple solution paths
  - Rich interconnected content

### Genre Variants

- **Classic Fantasy** (`GenreClassicFantasy`): Standard D&D-style
  - Balanced magic and combat
  - Traditional fantasy tropes
  - Medium difficulty and magic levels

- **Grimdark** (`GenreGrimdark`): Dark, harsh world
  - Low magic, high conflict
  - Survival-focused gameplay
  - Moral complexity and difficult choices

- **High Magic** (`GenreHighMagic`): Magic-saturated world
  - Fantastical environments
  - Magic-focused problem solving
  - Lower physical danger, higher mystery

- **Low Fantasy** (`GenreLowFantasy`): Minimal magic
  - Realistic tone and consequences
  - Limited magical elements
  - Focus on politics and intrigue

## Generated Content

### World Structure
- Procedural regions with distinct characteristics
- Settlement networks and trade routes
- Geographic features and resource distribution
- Climate and environmental systems

### Political Systems
- Faction relationships and conflicts
- Diplomatic status and trade agreements
- Territory control and strategic importance
- Economic networks and resource flows

### Population
- NPCs with unique personalities and motivations
- Background systems affecting character traits
- Social relationships and group dynamics
- Settlement populations and demographics

### Quest Systems
- Multiple quest types (fetch, escort, elimination, discovery, diplomatic)
- Dynamic difficulty scaling based on party level
- Branching objectives and multiple solutions
- Faction-aware quest generation

### Dialogue Systems
- Template-based conversation frameworks
- Personality-driven responses
- Context-aware dialogue (faction relationships, quest states)
- Markov chain enhancement for natural variety

### Game Assets
- Balanced spell progression for all levels
- Item sets appropriate for difficulty and theme
- Equipment scaling with character progression
- Treasure distribution matching world economy

## Performance Characteristics

- **Generation Time**: Under 1 second for most configurations
- **Memory Usage**: <10MB for complete game generation
- **File Size**: <5MB for generated configuration files
- **Startup Impact**: <1 second additional server startup time

## Quality Assurance

### Content Validation
- Automatic consistency checking
- Balance verification across systems
- Logical coherence validation
- Fallback mechanisms for edge cases

### Testing Coverage
- Unit tests: >90% coverage
- Integration tests: End-to-end generation
- Performance tests: Generation speed benchmarks
- Deterministic tests: Seed-based reproducibility

### Error Handling
- Graceful failure recovery
- Descriptive error messages
- Fallback to default configurations
- Automatic retry mechanisms

## Customization and Extension

### Template Modification
Edit `data/pcg/bootstrap_templates.yaml` to create custom configurations:

```yaml
my_custom_campaign:
  game_length: "medium"
  complexity_level: "advanced"
  genre_variant: "grimdark"
  max_players: 6
  starting_level: 3
  world_seed: 999
  enable_quick_start: true
  data_directory: "data"
```

### Parameter Adjustment
Modify generation parameters by editing the bootstrap configuration:

```go
config := pcg.DefaultBootstrapConfig()
config.MaxPlayers = 8           // Large party
config.StartingLevel = 5        // Veteran characters
config.WorldSeed = 42           // Deterministic world
config.EnableQuickStart = false // No starting scenario
```

### Content Generator Extension
Add new content types by implementing the `Generator` interface:

```go
type CustomGenerator struct {
    // Implementation details
}

func (g *CustomGenerator) Generate(ctx context.Context, params GenerationParams) (interface{}, error) {
    // Custom generation logic
}
```

## Best Practices

### Production Deployment
1. **Seed Management**: Use consistent seeds for reproducible environments
2. **Configuration Backup**: Save generated configurations for later modification
3. **Performance Monitoring**: Track generation times and resource usage
4. **Content Validation**: Verify generated content quality before deployment

### Development Workflow
1. **Rapid Prototyping**: Use bootstrap for quick campaign testing
2. **Content Iteration**: Generate multiple variants for comparison
3. **Parameter Tuning**: Adjust configurations based on player feedback
4. **Quality Testing**: Validate generated content meets requirements

### Troubleshooting
- **Slow Generation**: Reduce complexity level or content scope
- **Memory Issues**: Use streaming generation for large campaigns
- **Consistency Problems**: Check seed management and parameter validation
- **File Conflicts**: Ensure proper cleanup of existing configurations

## Future Enhancements

### Planned Features
- **Template Marketplace**: Community-shared configuration templates
- **Dynamic Adjustment**: Runtime parameter modification based on player behavior
- **Content Migration**: Tools for converting existing configurations to bootstrap
- **Advanced Analytics**: Detailed metrics on generated content quality

### Extension Points
- **Custom Generators**: Plugin system for specialized content types
- **External Data**: Integration with community databases and content
- **AI Enhancement**: Machine learning-driven content optimization
- **Multiplayer Scenarios**: Bootstrap for multi-party campaigns

## API Reference

### Core Types

```go
type BootstrapConfig struct {
    GameLength       GameLengthType `yaml:"game_length"`
    ComplexityLevel  ComplexityType `yaml:"complexity_level"`
    GenreVariant     GenreType      `yaml:"genre_variant"`
    MaxPlayers       int            `yaml:"max_players"`
    StartingLevel    int            `yaml:"starting_level"`
    WorldSeed        int64          `yaml:"world_seed"`
    EnableQuickStart bool           `yaml:"enable_quick_start"`
    DataDirectory    string         `yaml:"data_directory"`
}

type Bootstrap struct {
    // Private fields
}
```

### Key Functions

```go
// Detection and configuration
func DetectConfigurationPresence(dataDir string) bool
func DefaultBootstrapConfig() *BootstrapConfig

// Bootstrap creation and execution
func NewBootstrap(config *BootstrapConfig, world *game.World, logger *logrus.Logger) *Bootstrap
func (b *Bootstrap) GenerateCompleteGame(ctx context.Context) (*game.World, error)
```

For detailed API documentation, see the GoDoc comments in the source code.
