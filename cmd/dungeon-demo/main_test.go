package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// TestDungeonGeneratorBasic tests basic dungeon generation functionality.
func TestDungeonGeneratorBasic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	generator := pcg.NewDungeonGenerator(logger)

	world := &game.World{}
	params := pcg.GenerationParams{
		Seed:        12345,
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
				LevelCount:    2,
				LevelWidth:    20,
				LevelHeight:   25,
				RoomsPerLevel: 3,
				Theme:         pcg.ThemeClassic,
				Connectivity:  pcg.ConnectivityModerate,
				Density:       0.5,
				Difficulty: pcg.DifficultyProgression{
					BaseDifficulty:  2,
					ScalingFactor:   1.5,
					MaxDifficulty:   10,
					ProgressionType: "linear",
				},
			},
		},
	}

	result, err := generator.Generate(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)

	dungeon, ok := result.(*pcg.DungeonComplex)
	require.True(t, ok, "Expected result to be *pcg.DungeonComplex")
	assert.NotEmpty(t, dungeon.ID)
	assert.NotEmpty(t, dungeon.Name)
	assert.Len(t, dungeon.Levels, 2)
}

// TestDungeonGeneratorDeterminism tests that same seed produces same dungeon.
func TestDungeonGeneratorDeterminism(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	generator := pcg.NewDungeonGenerator(logger)

	world := &game.World{}
	makeParams := func(seed int64) pcg.GenerationParams {
		return pcg.GenerationParams{
			Seed:        seed,
			Difficulty:  2,
			PlayerLevel: 3,
			WorldState:  world,
			Timeout:     30 * time.Second,
			Constraints: map[string]interface{}{
				"dungeon_params": pcg.DungeonParams{
					GenerationParams: pcg.GenerationParams{
						Seed:        seed,
						Difficulty:  2,
						PlayerLevel: 3,
						WorldState:  world,
						Timeout:     30 * time.Second,
						Constraints: make(map[string]interface{}),
					},
					LevelCount:    2,
					LevelWidth:    20,
					LevelHeight:   25,
					RoomsPerLevel: 3,
					Theme:         pcg.ThemeClassic,
					Connectivity:  pcg.ConnectivityModerate,
					Density:       0.5,
					Difficulty: pcg.DifficultyProgression{
						BaseDifficulty:  2,
						ScalingFactor:   1.5,
						MaxDifficulty:   10,
						ProgressionType: "linear",
					},
				},
			},
		}
	}

	// Generate twice with same seed
	result1, err := generator.Generate(context.Background(), makeParams(99999))
	require.NoError(t, err)
	result2, err := generator.Generate(context.Background(), makeParams(99999))
	require.NoError(t, err)

	dungeon1 := result1.(*pcg.DungeonComplex)
	dungeon2 := result2.(*pcg.DungeonComplex)

	// Names and room counts should match
	assert.Equal(t, dungeon1.Name, dungeon2.Name)
	assert.Equal(t, len(dungeon1.Levels), len(dungeon2.Levels))
}

// TestDungeonGeneratorThemes tests different dungeon themes.
func TestDungeonGeneratorThemes(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	generator := pcg.NewDungeonGenerator(logger)

	themes := []pcg.LevelTheme{
		pcg.ThemeClassic,
		pcg.ThemeHorror,
		pcg.ThemeNatural,
		pcg.ThemeMechanical,
	}

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			world := &game.World{}
			params := pcg.GenerationParams{
				Seed:        42,
				Difficulty:  1,
				PlayerLevel: 1,
				WorldState:  world,
				Timeout:     30 * time.Second,
				Constraints: map[string]interface{}{
					"dungeon_params": pcg.DungeonParams{
						GenerationParams: pcg.GenerationParams{
							Seed:        42,
							Difficulty:  1,
							PlayerLevel: 1,
							WorldState:  world,
							Timeout:     30 * time.Second,
							Constraints: make(map[string]interface{}),
						},
						LevelCount:    1,
						LevelWidth:    20,
						LevelHeight:   25,
						RoomsPerLevel: 3,
						Theme:         theme,
						Connectivity:  pcg.ConnectivityLow,
						Density:       0.4,
						Difficulty: pcg.DifficultyProgression{
							BaseDifficulty:  1,
							ScalingFactor:   1.0,
							MaxDifficulty:   5,
							ProgressionType: "linear",
						},
					},
				},
			}

			result, err := generator.Generate(context.Background(), params)
			require.NoError(t, err)
			require.NotNil(t, result)

			dungeon := result.(*pcg.DungeonComplex)
			assert.NotEmpty(t, dungeon.Levels)
		})
	}
}

// TestDungeonGeneratorConnectivity tests different connectivity levels.
func TestDungeonGeneratorConnectivity(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	generator := pcg.NewDungeonGenerator(logger)

	connectivities := []pcg.ConnectivityLevel{
		pcg.ConnectivityLow,
		pcg.ConnectivityModerate,
		pcg.ConnectivityHigh,
		pcg.ConnectivityComplete,
	}

	for _, connectivity := range connectivities {
		t.Run(string(connectivity), func(t *testing.T) {
			world := &game.World{}
			params := pcg.GenerationParams{
				Seed:        123,
				Difficulty:  1,
				PlayerLevel: 1,
				WorldState:  world,
				Timeout:     30 * time.Second,
				Constraints: map[string]interface{}{
					"dungeon_params": pcg.DungeonParams{
						GenerationParams: pcg.GenerationParams{
							Seed:        123,
							Difficulty:  1,
							PlayerLevel: 1,
							WorldState:  world,
							Timeout:     30 * time.Second,
							Constraints: make(map[string]interface{}),
						},
						LevelCount:    2,
						LevelWidth:    20,
						LevelHeight:   25,
						RoomsPerLevel: 4,
						Theme:         pcg.ThemeClassic,
						Connectivity:  connectivity,
						Density:       0.5,
						Difficulty: pcg.DifficultyProgression{
							BaseDifficulty:  1,
							ScalingFactor:   1.0,
							MaxDifficulty:   5,
							ProgressionType: "linear",
						},
					},
				},
			}

			result, err := generator.Generate(context.Background(), params)
			require.NoError(t, err)
			require.NotNil(t, result)

			dungeon := result.(*pcg.DungeonComplex)
			assert.NotEmpty(t, dungeon.Levels)
		})
	}
}

// TestDungeonGeneratorMultipleLevels tests multi-level generation.
func TestDungeonGeneratorMultipleLevels(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	generator := pcg.NewDungeonGenerator(logger)

	tests := []struct {
		name       string
		levelCount int
	}{
		{"single_level", 1},
		{"two_levels", 2},
		{"five_levels", 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			world := &game.World{}
			params := pcg.GenerationParams{
				Seed:        555,
				Difficulty:  2,
				PlayerLevel: 2,
				WorldState:  world,
				Timeout:     60 * time.Second,
				Constraints: map[string]interface{}{
					"dungeon_params": pcg.DungeonParams{
						GenerationParams: pcg.GenerationParams{
							Seed:        555,
							Difficulty:  2,
							PlayerLevel: 2,
							WorldState:  world,
							Timeout:     60 * time.Second,
							Constraints: make(map[string]interface{}),
						},
						LevelCount:    tc.levelCount,
						LevelWidth:    20,
						LevelHeight:   25,
						RoomsPerLevel: 3,
						Theme:         pcg.ThemeClassic,
						Connectivity:  pcg.ConnectivityModerate,
						Density:       0.5,
						Difficulty: pcg.DifficultyProgression{
							BaseDifficulty:  2,
							ScalingFactor:   1.2,
							MaxDifficulty:   10,
							ProgressionType: "linear",
						},
					},
				},
			}

			result, err := generator.Generate(context.Background(), params)
			require.NoError(t, err)
			require.NotNil(t, result)

			dungeon := result.(*pcg.DungeonComplex)
			assert.Len(t, dungeon.Levels, tc.levelCount)

			// Verify each level exists
			for i := 1; i <= tc.levelCount; i++ {
				level, ok := dungeon.Levels[i]
				assert.True(t, ok, "Level %d should exist", i)
				if ok {
					assert.Equal(t, i, level.Level)
					assert.NotEmpty(t, level.Rooms)
				}
			}

			// Multi-level dungeons should have connections
			if tc.levelCount > 1 {
				assert.NotEmpty(t, dungeon.Connections, "Multi-level dungeon should have connections")
			}
		})
	}
}

// TestDungeonGeneratorRoomTypes verifies room types are generated.
func TestDungeonGeneratorRoomTypes(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	generator := pcg.NewDungeonGenerator(logger)

	world := &game.World{}
	params := pcg.GenerationParams{
		Seed:        77777,
		Difficulty:  3,
		PlayerLevel: 5,
		WorldState:  world,
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"dungeon_params": pcg.DungeonParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        77777,
					Difficulty:  3,
					PlayerLevel: 5,
					WorldState:  world,
					Timeout:     30 * time.Second,
					Constraints: make(map[string]interface{}),
				},
				LevelCount:    3,
				LevelWidth:    30,
				LevelHeight:   25,
				RoomsPerLevel: 6,
				Theme:         pcg.ThemeClassic,
				Connectivity:  pcg.ConnectivityHigh,
				Density:       0.6,
				Difficulty: pcg.DifficultyProgression{
					BaseDifficulty:  3,
					ScalingFactor:   1.5,
					MaxDifficulty:   15,
					ProgressionType: "linear",
				},
			},
		},
	}

	result, err := generator.Generate(context.Background(), params)
	require.NoError(t, err)

	dungeon := result.(*pcg.DungeonComplex)

	// Collect all room types across all levels
	roomTypes := make(map[pcg.RoomType]int)
	for _, level := range dungeon.Levels {
		for _, room := range level.Rooms {
			roomTypes[room.Type]++
		}
	}

	// Should have multiple room types
	assert.Greater(t, len(roomTypes), 1, "Dungeon should have varied room types")
}

// TestDungeonGeneratorMetadata tests that metadata is populated.
func TestDungeonGeneratorMetadata(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	generator := pcg.NewDungeonGenerator(logger)

	world := &game.World{}
	params := pcg.GenerationParams{
		Seed:        88888,
		Difficulty:  2,
		PlayerLevel: 3,
		WorldState:  world,
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"dungeon_params": pcg.DungeonParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        88888,
					Difficulty:  2,
					PlayerLevel: 3,
					WorldState:  world,
					Timeout:     30 * time.Second,
					Constraints: make(map[string]interface{}),
				},
				LevelCount:    2,
				LevelWidth:    20,
				LevelHeight:   25,
				RoomsPerLevel: 4,
				Theme:         pcg.ThemeClassic,
				Connectivity:  pcg.ConnectivityModerate,
				Density:       0.5,
				Difficulty: pcg.DifficultyProgression{
					BaseDifficulty:  2,
					ScalingFactor:   1.5,
					MaxDifficulty:   10,
					ProgressionType: "linear",
				},
			},
		},
	}

	result, err := generator.Generate(context.Background(), params)
	require.NoError(t, err)

	dungeon := result.(*pcg.DungeonComplex)
	assert.NotNil(t, dungeon.Metadata)
	assert.Contains(t, dungeon.Metadata, "total_rooms")
}

// TestDungeonGeneratorWithContext tests context handling.
func TestDungeonGeneratorWithContext(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	generator := pcg.NewDungeonGenerator(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	world := &game.World{}
	params := pcg.GenerationParams{
		Seed:        11111,
		Difficulty:  1,
		PlayerLevel: 1,
		WorldState:  world,
		Timeout:     5 * time.Second,
		Constraints: map[string]interface{}{
			"dungeon_params": pcg.DungeonParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        11111,
					Difficulty:  1,
					PlayerLevel: 1,
					WorldState:  world,
					Timeout:     5 * time.Second,
					Constraints: make(map[string]interface{}),
				},
				LevelCount:    1,
				LevelWidth:    25,
				LevelHeight:   25,
				RoomsPerLevel: 3,
				Theme:         pcg.ThemeClassic,
				Connectivity:  pcg.ConnectivityLow,
				Density:       0.3,
				Difficulty: pcg.DifficultyProgression{
					BaseDifficulty:  1,
					ScalingFactor:   1.0,
					MaxDifficulty:   5,
					ProgressionType: "linear",
				},
			},
		},
	}

	result, err := generator.Generate(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestMainOutputIntegration tests that main produces expected output.
func TestMainOutputIntegration(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	done := make(chan bool)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				t.Logf("main() panicked: %v", rec)
			}
			done <- true
		}()
		main()
	}()

	<-done

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify expected output sections
	assert.Contains(t, output, "GoldBox RPG")
	assert.Contains(t, output, "Dungeon Generator")
	assert.Contains(t, output, "Generation completed")
	assert.Contains(t, output, "Level Details")
}
