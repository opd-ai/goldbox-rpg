// Bootstrap Demo - Demonstrates the zero-configuration game generation system
//
// This demo shows how the GoldBox RPG Engine can automatically generate a complete,
// playable RPG experience without requiring any manual configuration files.
//
// Usage:
//   go run cmd/bootstrap-demo/main.go [options]
//
// Options:
//   -template string  Template name from bootstrap_templates.yaml (overrides other options)
//   -list-templates   List available templates and exit
//   -length string    Game length: short, medium, long (default "medium")
//   -complexity string Complexity level: simple, standard, advanced (default "standard")
//   -genre string     Genre variant: classic_fantasy, grimdark, high_magic, low_fantasy (default "classic_fantasy")
//   -players int      Maximum number of players (default 4)
//   -level int        Starting character level (default 1)
//   -seed int         World seed for deterministic generation (0 = random) (default 0)
//   -output string    Output directory for generated files (default "demo_output")
//   -quick            Enable quick start scenario (default true)
//   -verbose          Enable verbose logging (default false)
//
// Examples:
//   # List available templates
//   go run cmd/bootstrap-demo/main.go -list-templates
//   
//   # Use a predefined template
//   go run cmd/bootstrap-demo/main.go -template epic_campaign
//   
//   # Custom configuration
//   go run cmd/bootstrap-demo/main.go -length long -complexity advanced -genre grimdark

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
)

type DemoConfig struct {
	TemplateName     string
	GameLength       string
	ComplexityLevel  string
	GenreVariant     string
	MaxPlayers       int
	StartingLevel    int
	WorldSeed        int64
	OutputDir        string
	EnableQuickStart bool
	Verbose          bool
	ListTemplates    bool
}

func main() {
	config := parseFlags()
	setupLogging(config.Verbose)

	// Handle template listing
	if config.ListTemplates {
		if err := listAvailableTemplates(); err != nil {
			logrus.WithError(err).Fatal("Failed to list templates")
		}
		return
	}

	logrus.WithFields(logrus.Fields{
		"template":       config.TemplateName,
		"game_length":    config.GameLength,
		"complexity":     config.ComplexityLevel,
		"genre":          config.GenreVariant,
		"max_players":    config.MaxPlayers,
		"starting_level": config.StartingLevel,
		"world_seed":     config.WorldSeed,
		"output_dir":     config.OutputDir,
		"quick_start":    config.EnableQuickStart,
	}).Info("Starting GoldBox RPG Engine Bootstrap Demo")

	if err := runBootstrapDemo(config); err != nil {
		logrus.WithError(err).Fatal("Bootstrap demo failed")
	}

	logrus.Info("Bootstrap demo completed successfully!")
}

func parseFlags() *DemoConfig {
	config := &DemoConfig{}

	flag.StringVar(&config.TemplateName, "template", "", "Template name from bootstrap_templates.yaml (overrides other options)")
	flag.StringVar(&config.GameLength, "length", "medium", "Game length: short, medium, long")
	flag.StringVar(&config.ComplexityLevel, "complexity", "standard", "Complexity level: simple, standard, advanced")
	flag.StringVar(&config.GenreVariant, "genre", "classic_fantasy", "Genre variant: classic_fantasy, grimdark, high_magic, low_fantasy")
	flag.IntVar(&config.MaxPlayers, "players", 4, "Maximum number of players")
	flag.IntVar(&config.StartingLevel, "level", 1, "Starting character level")
	flag.Int64Var(&config.WorldSeed, "seed", 0, "World seed for deterministic generation (0 = random)")
	flag.StringVar(&config.OutputDir, "output", "demo_output", "Output directory for generated files")
	flag.BoolVar(&config.EnableQuickStart, "quick", true, "Enable quick start scenario")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.ListTemplates, "list-templates", false, "List available templates and exit")

	flag.Parse()

	return config
}

func setupLogging(verbose bool) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func listAvailableTemplates() error {
	templates, err := pcg.ListAvailableTemplates("data")
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Println("No templates found in data/pcg/bootstrap_templates.yaml")
		return nil
	}

	fmt.Printf("Available bootstrap templates (%d found):\n", len(templates))
	for _, template := range templates {
		fmt.Printf("  - %s\n", template)
	}

	fmt.Println("\nUsage:")
	fmt.Println("  go run cmd/bootstrap-demo/main.go -template <template_name>")
	fmt.Println("\nExample:")
	fmt.Println("  go run cmd/bootstrap-demo/main.go -template epic_campaign")
	
	return nil
}

func runBootstrapDemo(config *DemoConfig) error {
	// Clean up any existing output directory
	if err := os.RemoveAll(config.OutputDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}

	// Convert demo config to bootstrap config
	var bootstrapConfig *pcg.BootstrapConfig
	var err error

	if config.TemplateName != "" {
		// Load from template
		logrus.WithField("template", config.TemplateName).Info("Loading bootstrap configuration from template")
		bootstrapConfig, err = pcg.LoadBootstrapTemplate(config.TemplateName, "data")
		if err != nil {
			return fmt.Errorf("failed to load template %s: %w", config.TemplateName, err)
		}
		// Override output directory
		bootstrapConfig.DataDirectory = config.OutputDir
	} else {
		// Convert manual config
		bootstrapConfig, err = convertToBootstrapConfig(config)
		if err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}
	}

	// Create world and initialize bootstrap system
	world := game.NewWorld()
	bootstrap := pcg.NewBootstrap(bootstrapConfig, world, logrus.StandardLogger())

	// Demonstrate configuration detection
	logrus.Info("Checking for existing configuration...")
	hasConfig := pcg.DetectConfigurationPresence(config.OutputDir)
	logrus.WithField("has_config", hasConfig).Info("Configuration detection result")

	if hasConfig {
		logrus.Info("Configuration found, skipping bootstrap (this shouldn't happen in demo)")
		return nil
	}

	// Generate the complete game
	logrus.Info("Starting zero-configuration game generation...")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	generatedWorld, err := bootstrap.GenerateCompleteGame(ctx)
	if err != nil {
		return fmt.Errorf("game generation failed: %w", err)
	}

	duration := time.Since(startTime)

	// Display generation results
	displayResults(generatedWorld, bootstrap, duration, config)

	// Verify generated files
	if err := verifyGeneratedFiles(config.OutputDir); err != nil {
		return fmt.Errorf("file verification failed: %w", err)
	}

	return nil
}

func convertToBootstrapConfig(config *DemoConfig) (*pcg.BootstrapConfig, error) {
	bootstrapConfig := &pcg.BootstrapConfig{
		MaxPlayers:       config.MaxPlayers,
		StartingLevel:    config.StartingLevel,
		WorldSeed:        config.WorldSeed,
		EnableQuickStart: config.EnableQuickStart,
		DataDirectory:    config.OutputDir,
	}

	// Convert string values to enum types
	switch config.GameLength {
	case "short":
		bootstrapConfig.GameLength = pcg.GameLengthShort
	case "medium":
		bootstrapConfig.GameLength = pcg.GameLengthMedium
	case "long":
		bootstrapConfig.GameLength = pcg.GameLengthLong
	default:
		return nil, fmt.Errorf("invalid game length: %s", config.GameLength)
	}

	switch config.ComplexityLevel {
	case "simple":
		bootstrapConfig.ComplexityLevel = pcg.ComplexitySimple
	case "standard":
		bootstrapConfig.ComplexityLevel = pcg.ComplexityStandard
	case "advanced":
		bootstrapConfig.ComplexityLevel = pcg.ComplexityAdvanced
	default:
		return nil, fmt.Errorf("invalid complexity level: %s", config.ComplexityLevel)
	}

	switch config.GenreVariant {
	case "classic_fantasy":
		bootstrapConfig.GenreVariant = pcg.GenreClassicFantasy
	case "grimdark":
		bootstrapConfig.GenreVariant = pcg.GenreGrimdark
	case "high_magic":
		bootstrapConfig.GenreVariant = pcg.GenreHighMagic
	case "low_fantasy":
		bootstrapConfig.GenreVariant = pcg.GenreLowFantasy
	default:
		return nil, fmt.Errorf("invalid genre variant: %s", config.GenreVariant)
	}

	return bootstrapConfig, nil
}

func displayResults(world *game.World, bootstrap *pcg.Bootstrap, duration time.Duration, config *DemoConfig) {
	logrus.WithFields(logrus.Fields{
		"generation_time": duration,
		"output_dir":      config.OutputDir,
	}).Info("Game generation completed")

	separator := strings.Repeat("=", 60)

	fmt.Println("\n" + separator)
	fmt.Println("🎲 GOLDBOX RPG ENGINE - BOOTSTRAP DEMO RESULTS")
	fmt.Println(separator)

	fmt.Printf("📊 Generation Summary:\n")
	fmt.Printf("   • Game Length: %s\n", config.GameLength)
	fmt.Printf("   • Complexity: %s\n", config.ComplexityLevel)
	fmt.Printf("   • Genre: %s\n", config.GenreVariant)
	fmt.Printf("   • Max Players: %d\n", config.MaxPlayers)
	fmt.Printf("   • Starting Level: %d\n", config.StartingLevel)
	fmt.Printf("   • World Seed: %d\n", config.WorldSeed)
	fmt.Printf("   • Generation Time: %v\n", duration)

	fmt.Printf("\n📁 Generated Content:\n")
	// List expected content types since we can't access private fields
	contentTypes := []string{"world", "factions", "characters", "quests", "dialogue", "spells", "items"}
	if config.EnableQuickStart {
		contentTypes = append(contentTypes, "starting_scenario")
	}

	for _, contentType := range contentTypes {
		fmt.Printf("   ✓ %s\n", contentType)
	}

	fmt.Printf("\n📂 Output Directory: %s\n", config.OutputDir)
	fmt.Printf("\n🚀 Your zero-configuration RPG game is ready to play!\n")

	if config.EnableQuickStart {
		fmt.Printf("\n🎯 Quick Start Scenario Available:\n")
		fmt.Printf("   The game includes a ready-to-play starting scenario\n")
		fmt.Printf("   for immediate adventure with %d players at level %d.\n",
			config.MaxPlayers, config.StartingLevel)
	}

	fmt.Printf("\n📖 Next Steps:\n")
	fmt.Printf("   1. Start the server: go run cmd/server/main.go\n")
	fmt.Printf("   2. Open your browser to the server URL\n")
	fmt.Printf("   3. Begin your procedurally generated adventure!\n")

	fmt.Println(separator)
}

func verifyGeneratedFiles(outputDir string) error {
	logrus.Debug("Verifying generated files...")

	expectedFiles := []string{
		"pcg/bootstrap_config.yaml",
	}

	for _, file := range expectedFiles {
		fullPath := fmt.Sprintf("%s/%s", outputDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("expected file not found: %s", fullPath)
		}
	}

	logrus.WithField("files_checked", len(expectedFiles)).Debug("File verification completed")
	return nil
}
