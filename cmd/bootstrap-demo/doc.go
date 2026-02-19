// Package main provides a demonstration application for the zero-configuration
// game generation system in the GoldBox RPG engine.
//
// The bootstrap-demo application showcases how to use the PCG (Procedural Content
// Generation) bootstrap system to automatically generate a complete, playable RPG
// experience without requiring any manual configuration files.
//
// # Features
//
// The demo demonstrates several key capabilities:
//
//   - Zero-configuration game generation with sensible defaults
//   - Template-based configuration using predefined game profiles
//   - Custom configuration through command-line flags
//   - Automatic content generation including world, factions, characters, quests
//   - Quick-start scenario generation for immediate gameplay
//
// # Usage
//
// Run the demo with default settings:
//
//	go run ./cmd/bootstrap-demo
//
// List available templates:
//
//	go run ./cmd/bootstrap-demo -list-templates
//
// Use a predefined template:
//
//	go run ./cmd/bootstrap-demo -template epic_campaign
//
// Custom configuration:
//
//	go run ./cmd/bootstrap-demo -length long -complexity advanced -genre grimdark
//
// # Command-Line Options
//
//	-template string    Template name from bootstrap_templates.yaml (overrides other options)
//	-list-templates     List available templates and exit
//	-length string      Game length: short, medium, long (default "medium")
//	-complexity string  Complexity level: simple, standard, advanced (default "standard")
//	-genre string       Genre variant: classic_fantasy, grimdark, high_magic, low_fantasy (default "classic_fantasy")
//	-players int        Maximum number of players (default 4)
//	-level int          Starting character level (default 1)
//	-seed int           World seed for deterministic generation (0 = random)
//	-output string      Output directory for generated files (default "demo_output")
//	-quick              Enable quick start scenario (default true)
//	-verbose            Enable verbose logging (default false)
//
// # Game Length Settings
//
// The game length affects the scope and duration of the generated content:
//
//   - short: Quick adventure, minimal world complexity, few quests
//   - medium: Standard campaign, balanced world size, moderate quest chains
//   - long: Epic campaign, large world, extensive quest networks
//
// # Complexity Levels
//
// Complexity controls the depth of game mechanics:
//
//   - simple: Basic combat, straightforward quests, minimal faction interactions
//   - standard: Full combat system, branching quests, faction dynamics
//   - advanced: Complex tactical options, multi-layered quests, political intrigue
//
// # Genre Variants
//
// Genre affects the tone and flavor of generated content:
//
//   - classic_fantasy: Traditional high fantasy with magic and heroics
//   - grimdark: Dark, gritty setting with moral ambiguity
//   - high_magic: Magic-heavy world with powerful artifacts
//   - low_fantasy: Subtle magic, realistic combat, political focus
//
// # Output Structure
//
// The generated content is written to the output directory (default: demo_output)
// with the following structure:
//
//	demo_output/
//	├── pcg/
//	│   ├── bootstrap_config.yaml  # Generated configuration
//	│   └── ...                    # Other PCG data files
//	└── ...                        # Additional game data
//
// # Integration Example
//
// The bootstrap system can be used programmatically:
//
//	config := &pcg.BootstrapConfig{
//	    GameLength:       pcg.GameLengthMedium,
//	    ComplexityLevel:  pcg.ComplexityStandard,
//	    GenreVariant:     pcg.GenreClassicFantasy,
//	    MaxPlayers:       4,
//	    StartingLevel:    1,
//	    EnableQuickStart: true,
//	    DataDirectory:    "output",
//	}
//
//	world := game.NewWorld()
//	bootstrap := pcg.NewBootstrap(config, world, logger)
//
//	generatedWorld, err := bootstrap.GenerateCompleteGame(ctx)
//	if err != nil {
//	    // Handle generation error
//	}
package main
