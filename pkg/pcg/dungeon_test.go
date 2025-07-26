package pcg

import (
	"context"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDungeonGenerator(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "with nil logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewDungeonGenerator(tt.logger)

			assert.NotNil(t, generator)
			assert.Equal(t, "1.0.0", generator.GetVersion())
			assert.Equal(t, ContentTypeDungeon, generator.GetType())
			assert.NotNil(t, generator.logger)
			assert.NotNil(t, generator.rng)
		})
	}
}

func TestDungeonGenerator_GetType(t *testing.T) {
	generator := NewDungeonGenerator(nil)
	assert.Equal(t, ContentTypeDungeon, generator.GetType())
}

func TestDungeonGenerator_GetVersion(t *testing.T) {
	generator := NewDungeonGenerator(nil)
	assert.Equal(t, "1.0.0", generator.GetVersion())
}

func TestDungeonGenerator_Validate(t *testing.T) {
	generator := NewDungeonGenerator(nil)

	tests := []struct {
		name        string
		params      GenerationParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid parameters",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Timeout:     30 * time.Second,
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount:    3,
						LevelWidth:    50,
						LevelHeight:   50,
						RoomsPerLevel: 10,
						Theme:         ThemeClassic,
						Connectivity:  ConnectivityModerate,
						Density:       0.5,
						Difficulty: DifficultyProgression{
							BaseDifficulty:  1,
							ScalingFactor:   1.5,
							MaxDifficulty:   10,
							ProgressionType: "linear",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing dungeon_params",
			params: GenerationParams{
				Seed:        12345,
				Constraints: map[string]interface{}{},
			},
			expectError: true,
			errorMsg:    "expected dungeon_params in constraints",
		},
		{
			name: "invalid level count - too low",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount: 0,
					},
				},
			},
			expectError: true,
			errorMsg:    "level count must be between 1 and 20",
		},
		{
			name: "invalid level count - too high",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount: 25,
					},
				},
			},
			expectError: true,
			errorMsg:    "level count must be between 1 and 20",
		},
		{
			name: "invalid level width - too small",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount:  3,
						LevelWidth:  15,
						LevelHeight: 50,
					},
				},
			},
			expectError: true,
			errorMsg:    "level width must be between 20 and 200",
		},
		{
			name: "invalid level height - too large",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount:  3,
						LevelWidth:  50,
						LevelHeight: 250,
					},
				},
			},
			expectError: true,
			errorMsg:    "level height must be between 20 and 200",
		},
		{
			name: "invalid rooms per level - too few",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount:    3,
						LevelWidth:    50,
						LevelHeight:   50,
						RoomsPerLevel: 2,
					},
				},
			},
			expectError: true,
			errorMsg:    "rooms per level must be between 3 and 50",
		},
		{
			name: "invalid scaling factor - negative",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount:    3,
						LevelWidth:    50,
						LevelHeight:   50,
						RoomsPerLevel: 10,
						Difficulty: DifficultyProgression{
							ScalingFactor: -1.0,
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "scaling factor must be between 0 and 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.Validate(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDungeonGenerator_Generate(t *testing.T) {
	generator := NewDungeonGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		params      GenerationParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful generation - small dungeon",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  3,
				PlayerLevel: 2,
				WorldState:  &game.World{},
				Timeout:     30 * time.Second,
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount:    2,
						LevelWidth:    30,
						LevelHeight:   30,
						RoomsPerLevel: 5,
						Theme:         ThemeClassic,
						Connectivity:  ConnectivityModerate,
						Density:       0.4,
						Difficulty: DifficultyProgression{
							BaseDifficulty:  1,
							ScalingFactor:   1.0,
							MaxDifficulty:   5,
							ProgressionType: "linear",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "successful generation - large dungeon",
			params: GenerationParams{
				Seed:        54321,
				Difficulty:  8,
				PlayerLevel: 10,
				WorldState:  &game.World{},
				Timeout:     30 * time.Second,
				Constraints: map[string]interface{}{
					"dungeon_params": DungeonParams{
						LevelCount:    5,
						LevelWidth:    80,
						LevelHeight:   80,
						RoomsPerLevel: 15,
						Theme:         ThemeHorror,
						Connectivity:  ConnectivityHigh,
						Density:       0.7,
						Difficulty: DifficultyProgression{
							BaseDifficulty:  3,
							ScalingFactor:   2.0,
							MaxDifficulty:   15,
							ProgressionType: "exponential",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid parameters",
			params: GenerationParams{
				Seed:        12345,
				Constraints: map[string]interface{}{},
			},
			expectError: true,
			errorMsg:    "expected dungeon_params in constraints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.Generate(ctx, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify the result is a DungeonComplex
				dungeon, ok := result.(*DungeonComplex)
				require.True(t, ok, "result should be a DungeonComplex")

				// Extract dungeon params for validation
				dungeonParams := tt.params.Constraints["dungeon_params"].(DungeonParams)

				// Validate dungeon structure
				assert.NotEmpty(t, dungeon.ID)
				assert.NotEmpty(t, dungeon.Name)
				assert.Equal(t, dungeonParams.LevelCount, len(dungeon.Levels))
				assert.Equal(t, dungeonParams.Theme, dungeon.Theme)
				assert.Equal(t, dungeonParams.Difficulty, dungeon.Difficulty)
				assert.NotZero(t, dungeon.Generated)

				// Validate levels
				for i := 1; i <= dungeonParams.LevelCount; i++ {
					level, exists := dungeon.Levels[i]
					assert.True(t, exists, "Level %d should exist", i)
					assert.Equal(t, i, level.Level)
					assert.NotNil(t, level.Map)
					assert.Equal(t, dungeonParams.LevelWidth, level.Map.Width)
					assert.Equal(t, dungeonParams.LevelHeight, level.Map.Height)
					assert.NotEmpty(t, level.Rooms)
					assert.Equal(t, dungeonParams.Theme, level.Theme)
				}

				// Validate connections between levels
				if dungeonParams.LevelCount > 1 {
					assert.NotEmpty(t, dungeon.Connections, "Multi-level dungeon should have connections")
				}

				// Validate metadata
				assert.Contains(t, dungeon.Metadata, "total_rooms")
				assert.Contains(t, dungeon.Metadata, "connection_count")
				assert.Contains(t, dungeon.Metadata, "generation_seed")
				assert.Equal(t, tt.params.Seed, dungeon.Metadata["generation_seed"])
			}
		})
	}
}

func TestDungeonGenerator_DifficultyProgression(t *testing.T) {
	generator := NewDungeonGenerator(nil)

	tests := []struct {
		name        string
		level       int
		progression DifficultyProgression
		expected    int
	}{
		{
			name:  "linear progression level 1",
			level: 1,
			progression: DifficultyProgression{
				BaseDifficulty: 2,
				ScalingFactor:  1.5,
				MaxDifficulty:  10,
			},
			expected: 2,
		},
		{
			name:  "linear progression level 3",
			level: 3,
			progression: DifficultyProgression{
				BaseDifficulty: 2,
				ScalingFactor:  1.5,
				MaxDifficulty:  10,
			},
			expected: 5, // 2 + (3-1)*1.5 = 5
		},
		{
			name:  "capped at max difficulty",
			level: 10,
			progression: DifficultyProgression{
				BaseDifficulty: 1,
				ScalingFactor:  2.0,
				MaxDifficulty:  8,
			},
			expected: 8, // Would be 19 but capped at 8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.calculateLevelDifficulty(tt.level, tt.progression)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDungeonGenerator_RoomGeneration(t *testing.T) {
	generator := NewDungeonGenerator(nil)

	gameMap := &game.GameMap{
		Width:  50,
		Height: 50,
		Tiles:  make([][]game.MapTile, 50),
	}

	// Initialize map
	for y := 0; y < 50; y++ {
		gameMap.Tiles[y] = make([]game.MapTile, 50)
	}

	dungeonParams := DungeonParams{
		LevelWidth:    50,
		LevelHeight:   50,
		RoomsPerLevel: 8,
		Theme:         ThemeClassic,
	}

	rooms := generator.generateRoomsForLevel(gameMap, dungeonParams, 5)

	// Validate room generation
	assert.NotEmpty(t, rooms)
	assert.LessOrEqual(t, len(rooms), dungeonParams.RoomsPerLevel)

	// Check that rooms don't overlap
	for i, room1 := range rooms {
		for j, room2 := range rooms {
			if i != j {
				assert.False(t, room1.Bounds.Intersects(room2.Bounds),
					"Room %d and %d should not overlap", i, j)
			}
		}
	}

	// Validate room properties
	for i, room := range rooms {
		assert.NotEmpty(t, room.ID)
		assert.NotEqual(t, "", room.Type)
		assert.GreaterOrEqual(t, room.Bounds.Width, 5)
		assert.GreaterOrEqual(t, room.Bounds.Height, 5)
		assert.LessOrEqual(t, room.Bounds.Width, 12)
		assert.LessOrEqual(t, room.Bounds.Height, 12)
		assert.NotNil(t, room.Tiles)
		assert.NotNil(t, room.Properties)

		// First room should be entrance
		if i == 0 {
			assert.Equal(t, RoomTypeEntrance, room.Type)
		}
	}
}

func TestDungeonGenerator_ConnectionTypes(t *testing.T) {
	generator := NewDungeonGenerator(nil)

	tests := []struct {
		name        string
		theme       LevelTheme
		expectTypes []ConnectionType
	}{
		{
			name:  "classic theme",
			theme: ThemeClassic,
			expectTypes: []ConnectionType{
				ConnectionStairs, ConnectionLadder, ConnectionPit, ConnectionTunnel,
			},
		},
		{
			name:  "mechanical theme",
			theme: ThemeMechanical,
			expectTypes: []ConnectionType{
				ConnectionStairs, ConnectionLadder, ConnectionPit, ConnectionTunnel, ConnectionElevator,
			},
		},
		{
			name:  "magical theme",
			theme: ThemeMagical,
			expectTypes: []ConnectionType{
				ConnectionStairs, ConnectionLadder, ConnectionPit, ConnectionTunnel, ConnectionPortal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the weighted random selection multiple times
			foundTypes := make(map[ConnectionType]bool)

			for i := 0; i < 100; i++ {
				// Create dummy levels for connection type selection
				fromLevel := &DungeonLevel{Level: 1, Theme: tt.theme}
				toLevel := &DungeonLevel{Level: 2, Theme: tt.theme}

				connType := generator.chooseConnectionType(fromLevel, toLevel, tt.theme)
				foundTypes[connType] = true
			}

			// Verify that at least the basic connection types are possible
			assert.True(t, foundTypes[ConnectionStairs], "Stairs should be a possible connection type")
		})
	}
}

func TestDungeonGenerator_DeterministicGeneration(t *testing.T) {
	generator1 := NewDungeonGenerator(nil)
	generator2 := NewDungeonGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        98765,
		Difficulty:  4,
		PlayerLevel: 3,
		WorldState:  &game.World{},
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"dungeon_params": DungeonParams{
				LevelCount:    2,
				LevelWidth:    40,
				LevelHeight:   40,
				RoomsPerLevel: 6,
				Theme:         ThemeNatural,
				Connectivity:  ConnectivityModerate,
				Density:       0.5,
				Difficulty: DifficultyProgression{
					BaseDifficulty:  2,
					ScalingFactor:   1.5,
					MaxDifficulty:   8,
					ProgressionType: "linear",
				},
			},
		},
	}

	// Generate with same parameters
	result1, err1 := generator1.Generate(ctx, params)
	result2, err2 := generator2.Generate(ctx, params)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	dungeon1 := result1.(*DungeonComplex)
	dungeon2 := result2.(*DungeonComplex)

	// Verify deterministic generation (same structure with same seed)
	assert.Equal(t, len(dungeon1.Levels), len(dungeon2.Levels))

	for levelNum := range dungeon1.Levels {
		level1 := dungeon1.Levels[levelNum]
		level2 := dungeon2.Levels[levelNum]

		assert.Equal(t, level1.Level, level2.Level)
		assert.Equal(t, level1.Map.Width, level2.Map.Width)
		assert.Equal(t, level1.Map.Height, level2.Map.Height)
		assert.Equal(t, len(level1.Rooms), len(level2.Rooms))
	}
}

// Benchmark tests for performance validation

func BenchmarkDungeonGeneration_Small(b *testing.B) {
	generator := NewDungeonGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  3,
		PlayerLevel: 2,
		WorldState:  &game.World{},
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"dungeon_params": DungeonParams{
				LevelCount:    2,
				LevelWidth:    30,
				LevelHeight:   30,
				RoomsPerLevel: 5,
				Theme:         ThemeClassic,
				Connectivity:  ConnectivityModerate,
				Density:       0.4,
				Difficulty: DifficultyProgression{
					BaseDifficulty:  1,
					ScalingFactor:   1.0,
					MaxDifficulty:   5,
					ProgressionType: "linear",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDungeonGeneration_Large(b *testing.B) {
	generator := NewDungeonGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  8,
		PlayerLevel: 10,
		WorldState:  &game.World{},
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"dungeon_params": DungeonParams{
				LevelCount:    5,
				LevelWidth:    80,
				LevelHeight:   80,
				RoomsPerLevel: 15,
				Theme:         ThemeHorror,
				Connectivity:  ConnectivityHigh,
				Density:       0.7,
				Difficulty: DifficultyProgression{
					BaseDifficulty:  3,
					ScalingFactor:   2.0,
					MaxDifficulty:   15,
					ProgressionType: "exponential",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate(ctx, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}
