package levels

import (
	"context"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

func TestNewRoomCorridorGenerator(t *testing.T) {
	generator := NewRoomCorridorGenerator()

	if generator == nil {
		t.Fatal("NewRoomCorridorGenerator returned nil")
	}

	if generator.GetVersion() != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", generator.GetVersion())
	}

	if generator.GetType() != pcg.ContentTypeLevels {
		t.Errorf("Expected content type %s, got %s", pcg.ContentTypeLevels, generator.GetType())
	}
}

func TestRoomCorridorGenerator_Validate(t *testing.T) {
	generator := NewRoomCorridorGenerator()

	tests := []struct {
		name        string
		params      pcg.GenerationParams
		expectError bool
	}{
		{
			name: "valid parameters",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 10,
				Constraints: map[string]interface{}{
					"level_params": pcg.LevelParams{
						GenerationParams: pcg.GenerationParams{
							Seed:        12345,
							Difficulty:  5,
							PlayerLevel: 10,
						},
						MinRooms:      3,
						MaxRooms:      8,
						RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure},
						CorridorStyle: pcg.CorridorStraight,
						LevelTheme:    pcg.ThemeClassic,
						HasBoss:       true,
						SecretRooms:   1,
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid min rooms",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 10,
				Constraints: map[string]interface{}{
					"level_params": pcg.LevelParams{
						GenerationParams: pcg.GenerationParams{
							Seed:        12345,
							Difficulty:  5,
							PlayerLevel: 10,
						},
						MinRooms:      0,
						MaxRooms:      8,
						CorridorStyle: pcg.CorridorStraight,
						LevelTheme:    pcg.ThemeClassic,
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid max rooms",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 10,
				Constraints: map[string]interface{}{
					"level_params": pcg.LevelParams{
						GenerationParams: pcg.GenerationParams{
							Seed:        12345,
							Difficulty:  5,
							PlayerLevel: 10,
						},
						MinRooms:      5,
						MaxRooms:      3,
						CorridorStyle: pcg.CorridorStraight,
						LevelTheme:    pcg.ThemeClassic,
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid difficulty",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  25,
				PlayerLevel: 10,
				Constraints: map[string]interface{}{
					"level_params": pcg.LevelParams{
						GenerationParams: pcg.GenerationParams{
							Seed:        12345,
							Difficulty:  25,
							PlayerLevel: 10,
						},
						MinRooms:      3,
						MaxRooms:      8,
						CorridorStyle: pcg.CorridorStraight,
						LevelTheme:    pcg.ThemeClassic,
					},
				},
			},
			expectError: true,
		},
		{
			name: "missing level params",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 10,
				Constraints: map[string]interface{}{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.Validate(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRoomCorridorGenerator_Generate(t *testing.T) {
	generator := NewRoomCorridorGenerator()

	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        54321,
			Difficulty:  8,
			PlayerLevel: 12,
		},
		MinRooms:      4,
		MaxRooms:      6,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure, pcg.RoomTypePuzzle},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
		HasBoss:       true,
		SecretRooms:   1,
	}

	params := pcg.GenerationParams{
		Seed:        54321,
		Difficulty:  8,
		PlayerLevel: 12,
		Constraints: map[string]interface{}{
			"level_params": levelParams,
		},
	}

	ctx := context.Background()
	result, err := generator.Generate(ctx, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	level, ok := result.(*game.Level)
	if !ok {
		t.Fatal("Generate did not return a *game.Level")
	}

	// Validate the generated level
	if level.Width <= 0 || level.Height <= 0 {
		t.Error("Level dimensions must be positive")
	}

	if len(level.Tiles) != level.Height {
		t.Errorf("Expected %d tile rows, got %d", level.Height, len(level.Tiles))
	}

	if len(level.Tiles) > 0 && len(level.Tiles[0]) != level.Width {
		t.Errorf("Expected %d tile columns, got %d", level.Width, len(level.Tiles[0]))
	}

	// Check for required properties
	if _, exists := level.Properties["theme"]; !exists {
		t.Error("Level should have a theme property")
	}

	if _, exists := level.Properties["difficulty"]; !exists {
		t.Error("Level should have a difficulty property")
	}

	if _, exists := level.Properties["room_count"]; !exists {
		t.Error("Level should have a room_count property")
	}
}

func TestRoomCorridorGenerator_GenerateLevel(t *testing.T) {
	generator := NewRoomCorridorGenerator()

	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        98765,
			Difficulty:  10,
			PlayerLevel: 15,
		},
		MinRooms:      3,
		MaxRooms:      5,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure},
		CorridorStyle: pcg.CorridorWindy,
		LevelTheme:    pcg.ThemeHorror,
		HasBoss:       false,
		SecretRooms:   0,
	}

	ctx := context.Background()
	level, err := generator.GenerateLevel(ctx, levelParams)
	if err != nil {
		t.Fatalf("GenerateLevel failed: %v", err)
	}

	if level == nil {
		t.Fatal("GenerateLevel returned nil level")
	}

	// Validate level structure
	if level.ID == "" {
		t.Error("Level should have an ID")
	}

	if level.Name == "" {
		t.Error("Level should have a name")
	}

	// Validate tiles are properly initialized
	if len(level.Tiles) == 0 {
		t.Error("Level should have tiles")
	}

	// Check that we have both walkable and non-walkable tiles
	hasWalkable := false
	hasWalls := false

	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			if level.Tiles[y][x].Walkable {
				hasWalkable = true
			} else {
				hasWalls = true
			}
		}
	}

	if !hasWalkable {
		t.Error("Level should have walkable tiles")
	}

	if !hasWalls {
		t.Error("Level should have wall tiles")
	}
}

func TestRoomCorridorGenerator_GenerateRoom(t *testing.T) {
	generator := NewRoomCorridorGenerator()

	bounds := pcg.Rectangle{X: 5, Y: 5, Width: 10, Height: 8}
	roomType := pcg.RoomTypeCombat
	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        11111,
			Difficulty:  6,
			PlayerLevel: 8,
		},
		LevelTheme: pcg.ThemeClassic,
	}

	ctx := context.Background()
	room, err := generator.GenerateRoom(ctx, bounds, roomType, levelParams)
	if err != nil {
		t.Fatalf("GenerateRoom failed: %v", err)
	}

	if room == nil {
		t.Fatal("GenerateRoom returned nil room")
	}

	// Validate room structure
	if room.Type != roomType {
		t.Errorf("Expected room type %s, got %s", roomType, room.Type)
	}

	if room.Bounds != bounds {
		t.Errorf("Expected bounds %+v, got %+v", bounds, room.Bounds)
	}

	// Validate room tiles
	if len(room.Tiles) != bounds.Height {
		t.Errorf("Expected %d tile rows, got %d", bounds.Height, len(room.Tiles))
	}

	if len(room.Tiles) > 0 && len(room.Tiles[0]) != bounds.Width {
		t.Errorf("Expected %d tile columns, got %d", bounds.Width, len(room.Tiles[0]))
	}
}

func TestCalculateLevelDimensions(t *testing.T) {
	generator := NewRoomCorridorGenerator()

	tests := []struct {
		name   string
		params pcg.LevelParams
	}{
		{
			name: "classic theme",
			params: pcg.LevelParams{
				MinRooms:   3,
				MaxRooms:   6,
				LevelTheme: pcg.ThemeClassic,
			},
		},
		{
			name: "horror theme",
			params: pcg.LevelParams{
				MinRooms:   4,
				MaxRooms:   8,
				LevelTheme: pcg.ThemeHorror,
			},
		},
		{
			name: "mechanical theme",
			params: pcg.LevelParams{
				MinRooms:   5,
				MaxRooms:   10,
				LevelTheme: pcg.ThemeMechanical,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := generator.calculateLevelDimensions(tt.params)

			if width < 30 || height < 30 {
				t.Errorf("Dimensions too small: %dx%d", width, height)
			}

			if width > 200 || height > 200 {
				t.Errorf("Dimensions too large: %dx%d", width, height)
			}
		})
	}
}

func TestDeterministicGeneration(t *testing.T) {
	// Use seeded generators for deterministic testing
	seed := int64(999999)
	generator1 := NewRoomCorridorGeneratorWithSeed(seed)
	generator2 := NewRoomCorridorGeneratorWithSeed(seed)

	levelParams := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        999999,
			Difficulty:  7,
			PlayerLevel: 10,
		},
		MinRooms:      4,
		MaxRooms:      6,
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
	}

	ctx := context.Background()

	level1, err1 := generator1.GenerateLevel(ctx, levelParams)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	level2, err2 := generator2.GenerateLevel(ctx, levelParams)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	// Compare basic properties
	if level1.Width != level2.Width {
		t.Errorf("Width mismatch: %d vs %d", level1.Width, level2.Width)
	}

	if level1.Height != level2.Height {
		t.Errorf("Height mismatch: %d vs %d", level1.Height, level2.Height)
	}

	// Compare some tile properties (basic smoke test for determinism)
	if len(level1.Tiles) > 0 && len(level2.Tiles) > 0 {
		if level1.Tiles[0][0].Type != level2.Tiles[0][0].Type {
			t.Error("First tile type should be identical for same seed")
		}
	}
}

func TestNewRoomCorridorGeneratorWithSeed(t *testing.T) {
	// Test that the seeded constructor works correctly
	seed := int64(12345)
	generator := NewRoomCorridorGeneratorWithSeed(seed)

	if generator == nil {
		t.Fatal("NewRoomCorridorGeneratorWithSeed returned nil")
	}

	if generator.GetVersion() != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", generator.GetVersion())
	}

	if generator.GetType() != pcg.ContentTypeLevels {
		t.Errorf("Expected content type %s, got %s", pcg.ContentTypeLevels, generator.GetType())
	}

	if generator.rng == nil {
		t.Error("Generator RNG should not be nil")
	}
}

func TestNewRoomCorridorGeneratorWithSeed_Determinism(t *testing.T) {
	// Test that the same seed produces the same initial state
	seed := int64(54321)

	generator1 := NewRoomCorridorGeneratorWithSeed(seed)
	generator2 := NewRoomCorridorGeneratorWithSeed(seed)

	// Generate some random numbers from each generator's RNG
	// They should be identical for the same seed
	val1 := generator1.rng.Intn(1000)
	val2 := generator2.rng.Intn(1000)

	if val1 != val2 {
		t.Errorf("Same seed should produce same RNG values, got %d vs %d", val1, val2)
	}
}

func TestRoomCorridorGenerator_GenerateLevel_ContextCancellation(t *testing.T) {
	generator := NewRoomCorridorGeneratorWithSeed(42)

	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		expectError   bool
		errorContains string
	}{
		{
			name: "cancelled context before start",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			expectError:   true,
			errorContains: "cancelled before start",
		},
		{
			name: "valid context completes successfully",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			expectError: false,
		},
	}

	params := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        42,
			Difficulty:  5,
			PlayerLevel: 10,
		},
		MinRooms:      3,
		MaxRooms:      5,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat, pcg.RoomTypeTreasure},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupContext()
			defer cancel()

			level, err := generator.GenerateLevel(ctx, params)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				if level != nil {
					t.Error("expected nil level when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if level == nil {
					t.Error("expected non-nil level")
				}
			}
		})
	}
}

func TestRoomCorridorGenerator_GenerateLevel_DeadlineExceeded(t *testing.T) {
	generator := NewRoomCorridorGeneratorWithSeed(42)

	// Create a context that has already expired
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()

	params := pcg.LevelParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        42,
			Difficulty:  5,
			PlayerLevel: 10,
		},
		MinRooms:      3,
		MaxRooms:      5,
		RoomTypes:     []pcg.RoomType{pcg.RoomTypeCombat},
		CorridorStyle: pcg.CorridorStraight,
		LevelTheme:    pcg.ThemeClassic,
	}

	level, err := generator.GenerateLevel(ctx, params)

	if err == nil {
		t.Error("expected deadline exceeded error but got none")
	}

	if level != nil {
		t.Error("expected nil level when deadline exceeded")
	}

	if !containsString(err.Error(), "cancelled") {
		t.Errorf("expected error about cancellation, got: %v", err)
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
