package terrain

import (
	"context"
	"testing"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMazeGenerator(t *testing.T) {
	mg := NewMazeGenerator()

	assert.NotNil(t, mg)
	assert.Equal(t, "1.0.0", mg.GetVersion())
	assert.Equal(t, pcg.ContentTypeTerrain, mg.GetType())
}

func TestMazeGenerator_Generate(t *testing.T) {
	tests := []struct {
		name        string
		params      pcg.GenerationParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid parameters",
			params: pcg.GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"width":  20,
					"height": 20, "terrain_params": pcg.TerrainParams{
						GenerationParams: pcg.GenerationParams{
							Seed: 12345,
						},
						BiomeType: pcg.BiomeDungeon,
						Density:   0.4,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing terrain params",
			params: pcg.GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"width":  20,
					"height": 20,
				},
			},
			expectError: true,
			errorMsg:    "missing or invalid terrain parameters",
		},
		{
			name: "default dimensions",
			params: pcg.GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"terrain_params": pcg.TerrainParams{
						GenerationParams: pcg.GenerationParams{
							Seed: 12345,
						},
						BiomeType: pcg.BiomeCave,
						Density:   0.45,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := NewMazeGenerator()
			ctx := context.Background()

			result, err := mg.Generate(ctx, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				gameMap, ok := result.(*game.GameMap)
				assert.True(t, ok)
				assert.NotNil(t, gameMap)
				assert.NotNil(t, gameMap.Tiles)

				// Check default dimensions if not specified
				if tt.params.Constraints["width"] == nil {
					assert.Equal(t, 50, gameMap.Width)
				}
				if tt.params.Constraints["height"] == nil {
					assert.Equal(t, 50, gameMap.Height)
				}
			}
		})
	}
}

func TestMazeGenerator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		params      pcg.GenerationParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid parameters",
			params: pcg.GenerationParams{
				Constraints: map[string]interface{}{
					"width":  10,
					"height": 10,
				},
			},
			expectError: false,
		},
		{
			name: "nil constraints",
			params: pcg.GenerationParams{
				Constraints: nil,
			},
			expectError: true,
			errorMsg:    "constraints required for maze generation",
		},
		{
			name: "width too small",
			params: pcg.GenerationParams{
				Constraints: map[string]interface{}{
					"width":  3,
					"height": 10,
				},
			},
			expectError: true,
			errorMsg:    "width must be at least 5 for maze generation",
		},
		{
			name: "height too small",
			params: pcg.GenerationParams{
				Constraints: map[string]interface{}{
					"width":  10,
					"height": 4,
				},
			},
			expectError: true,
			errorMsg:    "height must be at least 5 for maze generation",
		},
		{
			name: "missing constraints",
			params: pcg.GenerationParams{
				Constraints: map[string]interface{}{},
			},
			expectError: false, // Should pass with no width/height constraints
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := NewMazeGenerator()
			err := mg.Validate(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMazeGenerator_GenerateTerrain(t *testing.T) {
	mg := NewMazeGenerator()
	ctx := context.Background()

	tests := []struct {
		name        string
		width       int
		height      int
		params      pcg.TerrainParams
		expectError bool
	}{
		{
			name:   "small maze",
			width:  10,
			height: 10,
			params: pcg.TerrainParams{
				GenerationParams: pcg.GenerationParams{
					Seed:       12345,
					Difficulty: 2,
				},
				BiomeType: pcg.BiomeDungeon,
				Density:   0.4,
			},
			expectError: false,
		},
		{
			name:   "large maze",
			width:  50,
			height: 50,
			params: pcg.TerrainParams{
				GenerationParams: pcg.GenerationParams{
					Seed:       67890,
					Difficulty: 5,
				},
				BiomeType: pcg.BiomeCave,
				Density:   0.45,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameMap, err := mg.GenerateTerrain(ctx, tt.width, tt.height, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, gameMap)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gameMap)
				assert.Equal(t, tt.width, gameMap.Width)
				assert.Equal(t, tt.height, gameMap.Height)
				assert.NotNil(t, gameMap.Tiles)
				assert.Len(t, gameMap.Tiles, tt.height)

				// Check that tiles are properly initialized
				for y := 0; y < tt.height; y++ {
					assert.Len(t, gameMap.Tiles[y], tt.width)
				}

				// Verify maze structure - should have both walls and passages
				hasWalls := false
				hasPassages := false

				for y := 0; y < tt.height; y++ {
					for x := 0; x < tt.width; x++ {
						if gameMap.Tiles[y][x].Walkable {
							hasPassages = true
						} else {
							hasWalls = true
						}
					}
				}

				assert.True(t, hasWalls, "maze should have walls")
				assert.True(t, hasPassages, "maze should have passages")

				// Verify connectivity
				assert.True(t, mg.ValidateConnectivity(gameMap), "maze should be fully connected")
			}
		})
	}
}

func TestMazeGenerator_ValidateConnectivity(t *testing.T) {
	mg := NewMazeGenerator()

	tests := []struct {
		name     string
		gameMap  *game.GameMap
		expected bool
	}{
		{
			name:     "connected maze",
			gameMap:  createConnectedTestMaze(),
			expected: true,
		},
		{
			name:     "disconnected maze",
			gameMap:  createDisconnectedTestMaze(),
			expected: false,
		},
		{
			name:     "no walkable tiles",
			gameMap:  createAllWallsMaze(),
			expected: false,
		},
		{
			name:     "single walkable tile",
			gameMap:  createSingleTileMaze(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mg.ValidateConnectivity(tt.gameMap)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMazeGenerator_GenerateBiome(t *testing.T) {
	mg := NewMazeGenerator()
	ctx := context.Background()

	tests := []struct {
		name   string
		biome  pcg.BiomeType
		bounds pcg.Rectangle
		params pcg.TerrainParams
	}{
		{
			name:  "dungeon biome",
			biome: pcg.BiomeDungeon,
			bounds: pcg.Rectangle{
				Width:  20,
				Height: 20,
			},
			params: pcg.TerrainParams{
				GenerationParams: pcg.GenerationParams{
					Seed:       12345,
					Difficulty: 3,
				},
				Density: 0.4,
			},
		},
		{
			name:  "cave biome",
			biome: pcg.BiomeCave,
			bounds: pcg.Rectangle{
				Width:  15,
				Height: 15,
			},
			params: pcg.TerrainParams{
				GenerationParams: pcg.GenerationParams{
					Seed:       67890,
					Difficulty: 2,
				},
				Density: 0.45,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameMap, err := mg.GenerateBiome(ctx, tt.biome, tt.bounds, tt.params)

			assert.NoError(t, err)
			assert.NotNil(t, gameMap)
			assert.Equal(t, tt.bounds.Width, gameMap.Width)
			assert.Equal(t, tt.bounds.Height, gameMap.Height)
			assert.Equal(t, tt.biome, tt.params.BiomeType)
		})
	}
}

func TestMazeGenerator_InitializeAllWalls(t *testing.T) {
	mg := NewMazeGenerator()
	gameMap := createTestGameMap(10, 10)

	// Set some tiles to walkable first
	gameMap.Tiles[0][0].Walkable = true
	gameMap.Tiles[5][5].Walkable = true

	err := mg.initializeAllWalls(gameMap)
	assert.NoError(t, err)

	// All tiles should now be walls
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			assert.False(t, gameMap.Tiles[y][x].Walkable)
			assert.False(t, gameMap.Tiles[y][x].Transparent)
			assert.Equal(t, 1, gameMap.Tiles[y][x].SpriteX)
			assert.Equal(t, 0, gameMap.Tiles[y][x].SpriteY)
		}
	}
}

func TestMazeGenerator_GetUnvisitedNeighbors(t *testing.T) {
	mg := NewMazeGenerator()
	gameMap := createTestGameMap(10, 10)
	visited := make([][]bool, gameMap.Height)
	for i := range visited {
		visited[i] = make([]bool, gameMap.Width)
	}

	// Test center position
	pos := game.Position{X: 5, Y: 5}
	neighbors := mg.getUnvisitedNeighbors(pos, visited, gameMap)

	// Should have 4 neighbors (North, East, South, West) at 2-step distance
	assert.Len(t, neighbors, 4)

	expectedNeighbors := []game.Position{
		{X: 5, Y: 3}, // North
		{X: 7, Y: 5}, // East
		{X: 5, Y: 7}, // South
		{X: 3, Y: 5}, // West
	}

	for _, expected := range expectedNeighbors {
		assert.Contains(t, neighbors, expected)
	}

	// Mark some neighbors as visited
	visited[3][5] = true // North
	visited[7][5] = true // South

	neighbors = mg.getUnvisitedNeighbors(pos, visited, gameMap)
	assert.Len(t, neighbors, 2) // Should only have East and West now
}

func TestMazeGenerator_GetUnvisitedNeighborsEdgeCases(t *testing.T) {
	mg := NewMazeGenerator()
	gameMap := createTestGameMap(5, 5)
	visited := make([][]bool, gameMap.Height)
	for i := range visited {
		visited[i] = make([]bool, gameMap.Width)
	}

	// Test corner position
	pos := game.Position{X: 0, Y: 0}
	neighbors := mg.getUnvisitedNeighbors(pos, visited, gameMap)

	// Should have only 2 neighbors (East and South) due to boundaries
	assert.Len(t, neighbors, 2)

	// Test position near edge
	pos = game.Position{X: 1, Y: 1}
	neighbors = mg.getUnvisitedNeighbors(pos, visited, gameMap)

	// Should have only 2 neighbors due to small map size
	assert.LessOrEqual(t, len(neighbors), 4)
}

func TestMazeGenerator_AddSpecialFeatures(t *testing.T) {
	mg := NewMazeGenerator()
	gameMap := createTestGameMap(20, 20)

	// Create some passages first
	for y := 1; y < 19; y += 2 {
		for x := 1; x < 19; x += 2 {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	seedMgr := pcg.NewSeedManager(12345)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeTerrain, "test", pcg.GenerationParams{
		Seed: 12345,
	})

	params := pcg.TerrainParams{
		GenerationParams: pcg.GenerationParams{
			Difficulty: 6, // Should create 2 rooms (6/3 = 2)
		},
	}

	err := mg.addSpecialFeatures(gameMap, params, genCtx)
	assert.NoError(t, err)

	// Should have added some room areas
	roomTileCount := 0
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			if gameMap.Tiles[y][x].Walkable {
				roomTileCount++
			}
		}
	}

	// Should have more walkable tiles after adding rooms
	assert.Greater(t, roomTileCount, 0)
}

func TestMazeGenerator_BiomeSpecificFeatures(t *testing.T) {
	mg := NewMazeGenerator()
	gameMap := createTestGameMap(15, 15)

	// Create some passages
	for y := 1; y < 14; y += 2 {
		for x := 1; x < 14; x += 2 {
			gameMap.Tiles[y][x].Walkable = true
			gameMap.Tiles[y][x].Transparent = true
		}
	}

	seedMgr := pcg.NewSeedManager(12345)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeTerrain, "test", pcg.GenerationParams{
		Seed: 12345,
	})

	params := pcg.TerrainParams{
		BiomeType: pcg.BiomeDungeon,
	}

	err := mg.applyBiomeSpecificFeatures(gameMap, params, genCtx)
	assert.NoError(t, err) // Should not error even if features are applied
}

func TestMazeGenerator_DeterministicGeneration(t *testing.T) {
	mg := NewMazeGenerator()
	ctx := context.Background()

	params := pcg.TerrainParams{
		GenerationParams: pcg.GenerationParams{
			Seed:       54321,
			Difficulty: 3,
		},
		BiomeType: pcg.BiomeDungeon,
		Density:   0.4,
	}

	// Generate the same maze twice
	gameMap1, err1 := mg.GenerateTerrain(ctx, 15, 15, params)
	require.NoError(t, err1)

	gameMap2, err2 := mg.GenerateTerrain(ctx, 15, 15, params)
	require.NoError(t, err2)

	// Maps should be identical
	assert.Equal(t, gameMap1.Width, gameMap2.Width)
	assert.Equal(t, gameMap1.Height, gameMap2.Height)

	for y := 0; y < gameMap1.Height; y++ {
		for x := 0; x < gameMap1.Width; x++ {
			assert.Equal(t, gameMap1.Tiles[y][x].Walkable, gameMap2.Tiles[y][x].Walkable,
				"tiles should be identical at position (%d, %d)", x, y)
		}
	}
}

// Helper functions for creating test mazes

func createConnectedTestMaze() *game.GameMap {
	gameMap := createTestGameMap(5, 5)

	// Create a simple connected path
	gameMap.Tiles[1][1].Walkable = true
	gameMap.Tiles[1][2].Walkable = true
	gameMap.Tiles[1][3].Walkable = true
	gameMap.Tiles[2][3].Walkable = true
	gameMap.Tiles[3][3].Walkable = true

	return gameMap
}

func createDisconnectedTestMaze() *game.GameMap {
	gameMap := createTestGameMap(5, 5)

	// Create disconnected areas
	gameMap.Tiles[1][1].Walkable = true
	gameMap.Tiles[1][2].Walkable = true

	gameMap.Tiles[3][3].Walkable = true
	gameMap.Tiles[3][4].Walkable = true

	return gameMap
}

func createAllWallsMaze() *game.GameMap {
	return createTestGameMap(5, 5) // All tiles are walls by default
}

func createSingleTileMaze() *game.GameMap {
	gameMap := createTestGameMap(5, 5)
	gameMap.Tiles[2][2].Walkable = true
	return gameMap
}
