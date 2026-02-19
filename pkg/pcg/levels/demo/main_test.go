package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"goldbox-rpg/pkg/pcg"
	"goldbox-rpg/pkg/pcg/levels"
)

// TestGenerateLevelSuccess verifies that level generation completes successfully
// with the default demo parameters.
func TestGenerateLevelSuccess(t *testing.T) {
	generator := levels.NewRoomCorridorGeneratorWithSeed(42)

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

	ctx := context.Background()
	level, err := generator.GenerateLevel(ctx, levelParams)
	if err != nil {
		t.Fatalf("GenerateLevel failed: %v", err)
	}

	if level == nil {
		t.Fatal("GenerateLevel returned nil level")
	}

	if level.Width <= 0 || level.Height <= 0 {
		t.Errorf("Invalid level dimensions: %dx%d", level.Width, level.Height)
	}

	if len(level.Tiles) == 0 {
		t.Error("Level has no tiles")
	}

	if level.Name == "" {
		t.Error("Level has no name")
	}
}

// TestGenerateLevelDeterminism verifies that the same seed produces identical levels.
func TestGenerateLevelDeterminism(t *testing.T) {
	seed := int64(12345)
	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        seed,
			Difficulty:  5,
			PlayerLevel: 5,
		},
		MinRooms:      3,
		MaxRooms:      5,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
		HasBoss:       false,
		SecretRooms:   0,
	}

	ctx := context.Background()

	// Generate first level
	gen1 := levels.NewRoomCorridorGeneratorWithSeed(seed)
	level1, err := gen1.GenerateLevel(ctx, levelParams)
	if err != nil {
		t.Fatalf("First GenerateLevel failed: %v", err)
	}

	// Generate second level with same seed
	gen2 := levels.NewRoomCorridorGeneratorWithSeed(seed)
	level2, err := gen2.GenerateLevel(ctx, levelParams)
	if err != nil {
		t.Fatalf("Second GenerateLevel failed: %v", err)
	}

	// Compare dimensions
	if level1.Width != level2.Width || level1.Height != level2.Height {
		t.Errorf("Dimensions differ: %dx%d vs %dx%d",
			level1.Width, level1.Height, level2.Width, level2.Height)
	}

	// Compare tile walkability in overlapping region
	maxY := min(len(level1.Tiles), len(level2.Tiles))
	for y := 0; y < maxY; y++ {
		maxX := min(len(level1.Tiles[y]), len(level2.Tiles[y]))
		for x := 0; x < maxX; x++ {
			if level1.Tiles[y][x].Walkable != level2.Tiles[y][x].Walkable {
				t.Errorf("Tile at (%d,%d) differs: walkable=%v vs %v",
					x, y, level1.Tiles[y][x].Walkable, level2.Tiles[y][x].Walkable)
			}
		}
	}
}

// TestLevelMapVisualization verifies that level tiles can be visualized
// similar to the main() output.
func TestLevelMapVisualization(t *testing.T) {
	generator := levels.NewRoomCorridorGeneratorWithSeed(42)

	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        42,
			Difficulty:  5,
			PlayerLevel: 5,
		},
		MinRooms:      2,
		MaxRooms:      4,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
		HasBoss:       false,
		SecretRooms:   0,
	}

	ctx := context.Background()
	level, err := generator.GenerateLevel(ctx, levelParams)
	if err != nil {
		t.Fatalf("GenerateLevel failed: %v", err)
	}

	// Generate ASCII visualization
	var buf bytes.Buffer
	maxY := min(20, level.Height)
	maxX := min(20, level.Width)

	for y := 0; y < maxY; y++ {
		for x := 0; x < maxX; x++ {
			if level.Tiles[y][x].Walkable {
				buf.WriteByte('.')
			} else {
				buf.WriteByte('#')
			}
		}
		buf.WriteByte('\n')
	}

	output := buf.String()

	// Verify output contains expected characters
	if !strings.Contains(output, ".") {
		t.Error("Visualization should contain walkable tiles (.)")
	}

	if !strings.Contains(output, "#") {
		t.Error("Visualization should contain wall tiles (#)")
	}

	// Verify dimensions
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != maxY {
		t.Errorf("Expected %d lines, got %d", maxY, len(lines))
	}
}

// TestLevelPropertiesWithBoss verifies levels with boss rooms have correct properties.
func TestLevelPropertiesWithBoss(t *testing.T) {
	generator := levels.NewRoomCorridorGeneratorWithSeed(42)

	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        42,
			Difficulty:  8,
			PlayerLevel: 10,
		},
		MinRooms:      4,
		MaxRooms:      7,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
		HasBoss:       true,
		SecretRooms:   2,
	}

	ctx := context.Background()
	level, err := generator.GenerateLevel(ctx, levelParams)
	if err != nil {
		t.Fatalf("GenerateLevel failed: %v", err)
	}

	// Check that Properties map exists
	if level.Properties == nil {
		t.Error("Level properties should not be nil")
	}

	// Verify boss room property if set
	if hasBoss, ok := level.Properties["has_boss"]; ok {
		if hasBoss != true {
			t.Errorf("Expected has_boss=true, got %v", hasBoss)
		}
	}
}

// TestLevelGenerationWithCanceledContext verifies context cancellation is handled.
func TestLevelGenerationWithCanceledContext(t *testing.T) {
	generator := levels.NewRoomCorridorGeneratorWithSeed(42)

	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        42,
			Difficulty:  5,
			PlayerLevel: 5,
		},
		MinRooms:      2,
		MaxRooms:      4,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
	}

	// Create and immediately cancel context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Generation may or may not fail depending on implementation
	// At minimum, verify it doesn't panic
	_, _ = generator.GenerateLevel(ctx, levelParams)
}

// TestVariousRoomTypes verifies generation works with different room type combinations.
func TestVariousRoomTypes(t *testing.T) {
	testCases := []struct {
		name      string
		roomTypes []pcg.RoomType
	}{
		{"CombatOnly", []pcg.RoomType{pcg.RoomTypeCombat}},
		{"TreasureOnly", []pcg.RoomType{pcg.RoomTypeTreasure}},
		{"PuzzleOnly", []pcg.RoomType{pcg.RoomTypePuzzle}},
		{"AllTypes", []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure, pcg.RoomTypePuzzle}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generator := levels.NewRoomCorridorGeneratorWithSeed(42)

			levelParams := pcg.LevelParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        42,
					Difficulty:  5,
					PlayerLevel: 5,
				},
				MinRooms:      2,
				MaxRooms:      4,
				RoomTypes:     tc.roomTypes,
				CorridorStyle: pcg.CorridorStraight,
				LevelTheme:    pcg.ThemeClassic,
			}

			ctx := context.Background()
			level, err := generator.GenerateLevel(ctx, levelParams)
			if err != nil {
				t.Fatalf("GenerateLevel failed for %s: %v", tc.name, err)
			}

			if level == nil {
				t.Fatalf("GenerateLevel returned nil for %s", tc.name)
			}
		})
	}
}

// TestVariousCorridorStyles verifies generation works with different corridor styles.
func TestVariousCorridorStyles(t *testing.T) {
	corridorStyles := []pcg.CorridorStyle{
		pcg.CorridorStraight,
		pcg.CorridorWindy,
	}

	for _, style := range corridorStyles {
		t.Run(fmt.Sprintf("Style%v", style), func(t *testing.T) {
			generator := levels.NewRoomCorridorGeneratorWithSeed(42)

			levelParams := pcg.LevelParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        42,
					Difficulty:  5,
					PlayerLevel: 5,
				},
				MinRooms:      3,
				MaxRooms:      5,
				RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat},
				CorridorStyle: style,
				LevelTheme:    pcg.ThemeClassic,
			}

			ctx := context.Background()
			level, err := generator.GenerateLevel(ctx, levelParams)
			if err != nil {
				t.Fatalf("GenerateLevel failed for style %v: %v", style, err)
			}

			if level == nil {
				t.Fatalf("GenerateLevel returned nil for style %v", style)
			}
		})
	}
}

// TestLevelThemes verifies generation works with different themes.
func TestLevelThemes(t *testing.T) {
	themes := []pcg.LevelTheme{
		pcg.ThemeClassic,
		pcg.ThemeHorror,
		pcg.ThemeNatural,
	}

	for _, theme := range themes {
		t.Run(fmt.Sprintf("Theme%s", theme), func(t *testing.T) {
			generator := levels.NewRoomCorridorGeneratorWithSeed(42)

			levelParams := pcg.LevelParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        42,
					Difficulty:  5,
					PlayerLevel: 5,
				},
				MinRooms:      2,
				MaxRooms:      4,
				RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat},
				CorridorStyle: pcg.CorridorStraight,
				LevelTheme:    theme,
			}

			ctx := context.Background()
			level, err := generator.GenerateLevel(ctx, levelParams)
			if err != nil {
				t.Fatalf("GenerateLevel failed for theme %s: %v", theme, err)
			}

			if level == nil {
				t.Fatalf("GenerateLevel returned nil for theme %s", theme)
			}
		})
	}
}

// TestDifficultyRange verifies generation works across difficulty spectrum.
func TestDifficultyRange(t *testing.T) {
	difficulties := []int{1, 5, 10}

	for _, diff := range difficulties {
		t.Run(fmt.Sprintf("Difficulty%d", diff), func(t *testing.T) {
			generator := levels.NewRoomCorridorGeneratorWithSeed(42)

			levelParams := pcg.LevelParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        42,
					Difficulty:  diff,
					PlayerLevel: diff,
				},
				MinRooms:      2,
				MaxRooms:      4,
				RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat},
				CorridorStyle: pcg.CorridorStraight,
				LevelTheme:    pcg.ThemeClassic,
			}

			ctx := context.Background()
			level, err := generator.GenerateLevel(ctx, levelParams)
			if err != nil {
				t.Fatalf("GenerateLevel failed for difficulty %d: %v", diff, err)
			}

			if level == nil {
				t.Fatalf("GenerateLevel returned nil for difficulty %d", diff)
			}
		})
	}
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
