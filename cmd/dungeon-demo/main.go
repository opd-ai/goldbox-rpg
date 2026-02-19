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

	// Create a logger for the demo
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create a dungeon generator
	generator := pcg.NewDungeonGenerator(logger)

	// Create a simple world for context
	world := &game.World{
		// Add minimal world state for context
	}

	// Set up generation parameters
	params := pcg.GenerationParams{
		Seed:        12345, // Fixed seed for reproducible results
		Difficulty:  2,
		PlayerLevel: 3,
		WorldState:  world,
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"dungeon_params": pcg.DungeonParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        12345,
					Difficulty:  2,
					PlayerLevel: 3,
					WorldState:  world,
					Timeout:     30 * time.Second,
					Constraints: make(map[string]interface{}),
				},
				LevelCount:    3,
				LevelWidth:    40,
				LevelHeight:   30,
				RoomsPerLevel: 6,
				Theme:         pcg.ThemeClassic,
				Connectivity:  pcg.ConnectivityModerate,
				Density:       0.6,
				Difficulty: pcg.DifficultyProgression{
					BaseDifficulty:  2,
					ScalingFactor:   1.5,
					MaxDifficulty:   10,
					ProgressionType: "linear",
				},
			},
		},
	}

	fmt.Printf("🎲 Generating dungeon with seed: %d\n", params.Seed)
	fmt.Printf("📏 Dimensions: %dx%d per level\n", 40, 30)
	fmt.Printf("🏠 Rooms per level: %d\n", 6)
	fmt.Printf("🎭 Theme: %s\n", pcg.ThemeClassic)
	fmt.Println()

	// Generate the dungeon
	start := time.Now()
	result, err := generator.Generate(context.Background(), params)
	duration := time.Since(start)

	if err != nil {
		return fmt.Errorf("dungeon generation failed: %w", err)
	}

	dungeon, ok := result.(*pcg.DungeonComplex)
	if !ok {
		return fmt.Errorf("unexpected result type: expected *pcg.DungeonComplex, got %T", result)
	}

	// Display results
	fmt.Printf("✅ Generation completed in %v\n", duration)
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
	fmt.Printf("💾 Generation seed: %d (use this for reproducible results)\n", params.Seed)

	return nil
}
