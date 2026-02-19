package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
)

// timeNow is the function used to get the current time.
// It defaults to time.Now but can be overridden in tests for reproducibility.
var timeNow = time.Now

// timeSince returns the duration since the given time.
// It defaults to time.Since but can be overridden in tests for reproducibility.
var timeSince = time.Since

// DemoConfig holds configuration for dungeon demo generation.
// It provides a reusable structure for customizing dungeon generation parameters.
type DemoConfig struct {
	// Seed for reproducible random generation. Use 0 for time-based seed.
	Seed int64
	// Difficulty affects enemy strength and trap frequency (1-10 scale).
	Difficulty int
	// PlayerLevel affects treasure quality and scaling (1-20 scale).
	PlayerLevel int
	// LevelCount specifies how many dungeon floors to generate.
	LevelCount int
	// LevelWidth is the width of each dungeon level in tiles.
	LevelWidth int
	// LevelHeight is the height of each dungeon level in tiles.
	LevelHeight int
	// RoomsPerLevel is the target number of rooms per dungeon level.
	RoomsPerLevel int
	// Theme affects the visual style and room types (classic, horror, natural, mechanical).
	Theme pcg.LevelTheme
	// Connectivity controls how connected rooms are (low, moderate, high, complete).
	Connectivity pcg.ConnectivityLevel
	// Density affects corridor and room placement density (0.0-1.0).
	Density float64
	// Timeout for generation operations.
	Timeout time.Duration
	// Logger for structured logging output. If nil, a default logger is created.
	Logger *logrus.Logger
}

// DefaultDemoConfig returns a DemoConfig with sensible defaults.
func DefaultDemoConfig() DemoConfig {
	return DemoConfig{
		Seed:          12345,
		Difficulty:    2,
		PlayerLevel:   3,
		LevelCount:    3,
		LevelWidth:    40,
		LevelHeight:   30,
		RoomsPerLevel: 6,
		Theme:         pcg.ThemeClassic,
		Connectivity:  pcg.ConnectivityModerate,
		Density:       0.6,
		Timeout:       30 * time.Second,
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the dungeon demo and returns any errors encountered.
func run() error {
	fmt.Println("🏰 GoldBox RPG - Multi-Level Dungeon Generator Demo")
	fmt.Println(strings.Repeat("=", 55))

	config := DefaultDemoConfig()
	result, err := GenerateDungeon(config)
	if err != nil {
		return err
	}

	DisplayDungeonResults(result, config)
	return nil
}

// GenerateDungeon creates a dungeon complex using the provided configuration.
// It returns the generated DungeonComplex and any error encountered.
// This function is exported for reusability by other packages.
func GenerateDungeon(config DemoConfig) (*pcg.DungeonComplex, error) {
	logger := config.Logger
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	logFields := logrus.Fields{
		"function":        "GenerateDungeon",
		"seed":            config.Seed,
		"difficulty":      config.Difficulty,
		"player_level":    config.PlayerLevel,
		"level_count":     config.LevelCount,
		"level_width":     config.LevelWidth,
		"level_height":    config.LevelHeight,
		"rooms_per_level": config.RoomsPerLevel,
		"theme":           config.Theme,
		"connectivity":    config.Connectivity,
		"density":         config.Density,
	}

	logger.WithFields(logFields).Info("Starting dungeon generation")

	generator := pcg.NewDungeonGenerator(logger)
	world := game.NewWorld()

	params := pcg.GenerationParams{
		Seed:        config.Seed,
		Difficulty:  config.Difficulty,
		PlayerLevel: config.PlayerLevel,
		WorldState:  world,
		Timeout:     config.Timeout,
		Constraints: map[string]interface{}{
			"dungeon_params": pcg.DungeonParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        config.Seed,
					Difficulty:  config.Difficulty,
					PlayerLevel: config.PlayerLevel,
					WorldState:  world,
					Timeout:     config.Timeout,
					Constraints: make(map[string]interface{}),
				},
				LevelCount:    config.LevelCount,
				LevelWidth:    config.LevelWidth,
				LevelHeight:   config.LevelHeight,
				RoomsPerLevel: config.RoomsPerLevel,
				Theme:         config.Theme,
				Connectivity:  config.Connectivity,
				Density:       config.Density,
				Difficulty: pcg.DifficultyProgression{
					BaseDifficulty:  config.Difficulty,
					ScalingFactor:   1.5,
					MaxDifficulty:   10,
					ProgressionType: "linear",
				},
			},
		},
	}

	start := timeNow()
	result, err := generator.Generate(context.Background(), params)
	duration := timeSince(start)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"function": "GenerateDungeon",
			"seed":     config.Seed,
			"duration": duration,
			"error":    err.Error(),
		}).Error("Dungeon generation failed")
		return nil, fmt.Errorf("dungeon generation failed: %w", err)
	}

	dungeon, ok := result.(*pcg.DungeonComplex)
	if !ok {
		logger.WithFields(logrus.Fields{
			"function":      "GenerateDungeon",
			"expected_type": "*pcg.DungeonComplex",
			"actual_type":   fmt.Sprintf("%T", result),
		}).Error("Unexpected result type from generator")
		return nil, fmt.Errorf("unexpected result type: expected *pcg.DungeonComplex, got %T", result)
	}

	logger.WithFields(logrus.Fields{
		"function":     "GenerateDungeon",
		"dungeon_id":   dungeon.ID,
		"dungeon_name": dungeon.Name,
		"levels":       len(dungeon.Levels),
		"connections":  len(dungeon.Connections),
		"duration":     duration,
	}).Info("Dungeon generation completed successfully")

	return dungeon, nil
}

// DisplayDungeonResults prints the dungeon generation results to stdout.
// This function is exported for reusability by other packages.
func DisplayDungeonResults(dungeon *pcg.DungeonComplex, config DemoConfig) {
	fmt.Printf("🎲 Generating dungeon with seed: %d\n", config.Seed)
	fmt.Printf("📏 Dimensions: %dx%d per level\n", config.LevelWidth, config.LevelHeight)
	fmt.Printf("🏠 Rooms per level: %d\n", config.RoomsPerLevel)
	fmt.Printf("🎭 Theme: %s\n", config.Theme)
	fmt.Println()

	fmt.Println("✅ Generation completed")
	fmt.Printf("🏰 Dungeon: %s (ID: %s)\n", dungeon.Name, dungeon.ID)
	fmt.Printf("📊 Levels: %d\n", len(dungeon.Levels))
	fmt.Printf("🔗 Connections: %d\n", len(dungeon.Connections))
	fmt.Printf("📈 Total rooms: %d\n", dungeon.Metadata["total_rooms"])
	fmt.Println()

	// Show level details
	fmt.Println("📋 Level Details:")
	for i := 1; i <= len(dungeon.Levels); i++ {
		level := dungeon.Levels[i]
		fmt.Printf("  Level %d: %d rooms, Difficulty %d, Theme: %s\n",
			level.Level, len(level.Rooms), level.Difficulty, level.Theme)

		// Show room types
		roomTypes := make(map[pcg.RoomType]int)
		for _, room := range level.Rooms {
			roomTypes[room.Type]++
		}

		fmt.Printf("    Room types: ")
		for roomType, count := range roomTypes {
			fmt.Printf("%s:%d ", roomType, count)
		}
		fmt.Println()

		// Show connections
		if len(level.Connections) > 0 {
			fmt.Printf("    Connections: %d to other levels\n", len(level.Connections))
		}
	}

	fmt.Println()

	// Show connection details
	if len(dungeon.Connections) > 0 {
		fmt.Println("🔗 Level Connections:")
		for _, conn := range dungeon.Connections {
			fmt.Printf("  %s: Level %d (%d,%d) ↔ Level %d (%d,%d)\n",
				conn.Type,
				conn.FromLevel, conn.FromPosition.X, conn.FromPosition.Y,
				conn.ToLevel, conn.ToPosition.X, conn.ToPosition.Y)
		}
	}

	fmt.Println()
	fmt.Printf("🎉 Demo completed! Dungeon ready for adventure.\n")
	fmt.Printf("💾 Generation seed: %d (use this for reproducible results)\n", config.Seed)
}
