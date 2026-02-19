package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"goldbox-rpg/pkg/pcg"
	"goldbox-rpg/pkg/pcg/levels"
)

func main() {
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	level, err := generator.GenerateLevel(ctx, levelParams)
	if err != nil {
		log.Fatalf("Failed to generate level: %v", err)
	}

	// Display basic level information
	fmt.Printf("Generated Level: %s\n", level.Name)
	fmt.Printf("Dimensions: %dx%d\n", level.Width, level.Height)
	fmt.Printf("Properties: %+v\n", level.Properties)

	// Show a small section of the level map
	fmt.Println("\nLevel Map (first 20x20 section):")
	for y := 0; y < 20 && y < level.Height; y++ {
		for x := 0; x < 20 && x < level.Width; x++ {
			if level.Tiles[y][x].Walkable {
				fmt.Print(".")
			} else {
				fmt.Print("#")
			}
		}
		fmt.Println()
	}

	fmt.Println("\nLevel generation completed successfully!")
}
