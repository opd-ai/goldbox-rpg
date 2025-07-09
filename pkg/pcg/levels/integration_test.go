package levels

import (
	"context"
	"testing"

	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
)

// TestLevelGeneratorRegistryIntegration tests using the level generator through the registry
func TestLevelGeneratorRegistryIntegration(t *testing.T) {
	// Create registry and register level generator
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce test noise

	registry := pcg.NewRegistry(logger)
	generator := NewRoomCorridorGenerator()

	err := registry.RegisterGenerator("room_corridor", generator)
	if err != nil {
		t.Fatalf("Failed to register level generator: %v", err)
	}

	// Create factory
	factory := pcg.NewFactory(registry, logger)

	// Test level generation through factory
	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        12345,
			Difficulty:  5,
			PlayerLevel: 8,
		},
		MinRooms:      3,
		MaxRooms:      5,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
		HasBoss:       false,
		SecretRooms:   0,
	}

	ctx := context.Background()
	level, err := factory.GenerateLevel(ctx, "room_corridor", levelParams)
	if err != nil {
		t.Fatalf("Factory level generation failed: %v", err)
	}

	if level == nil {
		t.Fatal("Factory returned nil level")
	}

	// Validate the generated level
	if level.Width <= 0 || level.Height <= 0 {
		t.Error("Level dimensions must be positive")
	}

	if level.ID == "" {
		t.Error("Level should have an ID")
	}

	if level.Name == "" {
		t.Error("Level should have a name")
	}

	// Verify factory integration properties
	if _, exists := level.Properties["generator"]; !exists {
		t.Error("Level should have generator property set")
	}
}

func TestCorridorFeatureGeneration(t *testing.T) {
	generator := NewRoomCorridorGenerator()

	// Test different corridor styles
	styles := []pcg.CorridorStyle{
		pcg.CorridorStraight,
		pcg.CorridorWindy,
		pcg.CorridorMaze,
		pcg.CorridorOrganic,
		pcg.CorridorMinimal,
	}

	for _, style := range styles {
		t.Run(string(style), func(t *testing.T) {
			levelParams := pcg.LevelParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        54321,
					Difficulty:  6,
					PlayerLevel: 10,
				},
				MinRooms:      3,
				MaxRooms:      4,
				CorridorStyle: style,
				LevelTheme:    pcg.ThemeClassic,
			}

			ctx := context.Background()
			level, err := generator.GenerateLevel(ctx, levelParams)
			if err != nil {
				t.Fatalf("Level generation failed for style %s: %v", style, err)
			}

			if level == nil {
				t.Fatalf("Level is nil for style %s", style)
			}

			// Basic validation
			if level.Width <= 0 || level.Height <= 0 {
				t.Errorf("Invalid dimensions for style %s: %dx%d", style, level.Width, level.Height)
			}
		})
	}
}

func TestDifferentThemes(t *testing.T) {
	generator := NewRoomCorridorGenerator()

	themes := []pcg.LevelTheme{
		pcg.ThemeClassic,
		pcg.ThemeHorror,
		pcg.ThemeNatural,
		pcg.ThemeMechanical,
		pcg.ThemeMagical,
		pcg.ThemeUndead,
		pcg.ThemeElemental,
	}

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			levelParams := pcg.LevelParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        98765,
					Difficulty:  7,
					PlayerLevel: 12,
				},
				MinRooms:      4,
				MaxRooms:      6,
				RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure, pcg.RoomTypeBoss},
				CorridorStyle: pcg.CorridorStraight,
				LevelTheme:    theme,
				HasBoss:       true,
				SecretRooms:   1,
			}

			ctx := context.Background()
			level, err := generator.GenerateLevel(ctx, levelParams)
			if err != nil {
				t.Fatalf("Level generation failed for theme %s: %v", theme, err)
			}

			if level == nil {
				t.Fatalf("Level is nil for theme %s", theme)
			}

			// Verify theme is recorded
			if themeProperty, exists := level.Properties["theme"]; !exists || themeProperty != theme {
				t.Errorf("Theme not properly recorded for %s", theme)
			}
		})
	}
}
