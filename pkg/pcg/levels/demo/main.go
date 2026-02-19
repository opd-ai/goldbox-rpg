// Package main provides a demonstration of the PCG levels package.
// It showcases procedural dungeon level generation using the room-corridor
// approach with BSP space partitioning.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
	"goldbox-rpg/pkg/pcg/levels"
)

// Config holds configuration options for the demo.
type Config struct {
	// Timeout is the maximum duration for level generation.
	Timeout time.Duration
	// Output is the writer for demo output.
	Output io.Writer
}

// DefaultConfig returns the default configuration for the demo.
func DefaultConfig() *Config {
	return &Config{
		Timeout: 30 * time.Second,
		Output:  os.Stdout,
	}
}

func main() {
	if err := run(DefaultConfig()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the level generation demo with the given configuration.
// It returns an error if level generation fails, enabling graceful cleanup.
func run(cfg *Config) error {
	// Create a new level generator
	generator := levels.NewRoomCorridorGenerator()

	// Set up generation parameters
	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        42,
			Difficulty:  8,
			PlayerLevel: 10,
		},
		MinRooms:      4,
		MaxRooms:      7,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure, pcg.RoomTypePuzzle},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
		HasBoss:       true,
		SecretRooms:   1,
	}

	// Generate a level with a timeout context to demonstrate cancellation handling
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	level, err := generator.GenerateLevel(ctx, levelParams)
	if err != nil {
		return fmt.Errorf("failed to generate level: %w", err)
	}

	// Display level information
	displayLevelInfo(cfg.Output, level)

	return nil
}

// displayLevelInfo outputs level details and a map preview.
func displayLevelInfo(w io.Writer, level *game.Level) {
	fmt.Fprintf(w, "Generated Level: %s\n", level.Name)
	fmt.Fprintf(w, "Dimensions: %dx%d\n", level.Width, level.Height)
	fmt.Fprintf(w, "Properties: %+v\n", level.Properties)

	// Show a small section of the level map
	fmt.Fprintln(w, "\nLevel Map (first 20x20 section):")
	for y := 0; y < 20 && y < level.Height; y++ {
		for x := 0; x < 20 && x < level.Width; x++ {
			if level.Tiles[y][x].Walkable {
				fmt.Fprint(w, ".")
			} else {
				fmt.Fprint(w, "#")
			}
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "\nLevel generation completed successfully!")
}
