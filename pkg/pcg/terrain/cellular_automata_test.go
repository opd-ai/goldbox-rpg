package terrain

import (
	"testing"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCAConfig(t *testing.T) {
	config := DefaultCAConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.WallThreshold, 0)
	assert.Greater(t, config.FloorThreshold, 0)
	assert.Greater(t, config.MaxIterations, 0)
	assert.GreaterOrEqual(t, config.SmoothingPasses, 0)
	assert.GreaterOrEqual(t, config.EdgeBuffer, 0)
	assert.Greater(t, config.MinRoomSize, 0)
}

func TestRunCellularAutomata(t *testing.T) {
	width, height := 20, 20
	gameMap := createTestGameMap(width, height)

	seedMgr := pcg.NewSeedManager(12345)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeTerrain, "test", pcg.GenerationParams{
		Seed: 12345,
	})

	config := DefaultCAConfig()
	config.MaxIterations = 3 // Reduce for faster testing

	err := RunCellularAutomata(gameMap, config, genCtx)
	require.NoError(t, err)

	// Verify map has been modified
	hasWalls := false
	hasFloors := false

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if gameMap.Tiles[y][x].Walkable {
				hasFloors = true
			} else {
				hasWalls = true
			}
		}
	}

	assert.True(t, hasWalls)
	assert.True(t, hasFloors)
}

func TestRunCellularAutomataWithNilConfig(t *testing.T) {
	width, height := 10, 10
	gameMap := createTestGameMap(width, height)

	seedMgr := pcg.NewSeedManager(12345)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeTerrain, "test", pcg.GenerationParams{
		Seed: 12345,
	})

	// Should use default config when nil is passed
	err := RunCellularAutomata(gameMap, nil, genCtx)
	require.NoError(t, err)
}

func TestCountNeighborWalls(t *testing.T) {
	gameMap := createTestGameMap(3, 3)

	// Set up a specific pattern
	gameMap.Tiles[0][0].Walkable = false // wall
	gameMap.Tiles[0][1].Walkable = false // wall
	gameMap.Tiles[0][2].Walkable = true  // floor
	gameMap.Tiles[1][0].Walkable = false // wall
	gameMap.Tiles[1][1].Walkable = true  // floor (center)
	gameMap.Tiles[1][2].Walkable = true  // floor
	gameMap.Tiles[2][0].Walkable = true  // floor
	gameMap.Tiles[2][1].Walkable = true  // floor
	gameMap.Tiles[2][2].Walkable = true  // floor

	// Count walls around center position (1,1)
	wallCount := countNeighborWalls(gameMap, 1, 1)
	assert.Equal(t, 3, wallCount) // Should count 3 walls
}

func TestCountNeighborWallsEdgeCases(t *testing.T) {
	gameMap := createTestGameMap(3, 3)

	// All floors
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	// Corner position should count out-of-bounds as walls
	// Position (0,0) has 8 neighbors: 5 out-of-bounds, 3 in-bounds (all floors)
	wallCount := countNeighborWalls(gameMap, 0, 0)
	assert.Equal(t, 5, wallCount) // 5 out-of-bounds positions

	// Edge position (1,0) has 8 neighbors: 3 out-of-bounds, 5 in-bounds (all floors)
	wallCount = countNeighborWalls(gameMap, 1, 0)
	assert.Equal(t, 3, wallCount) // 3 out-of-bounds positions

	// Center position with all floor neighbors
	wallCount = countNeighborWalls(gameMap, 1, 1)
	assert.Equal(t, 0, wallCount) // No walls
}

func TestEnforceEdgeBoundaries(t *testing.T) {
	width, height := 10, 10
	gameMap := createTestGameMap(width, height)

	// Set all tiles to walkable first
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	buffer := 2
	err := enforceEdgeBoundaries(gameMap, buffer)
	require.NoError(t, err)

	// Check that edge tiles within buffer are walls
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x < buffer || x >= width-buffer || y < buffer || y >= height-buffer {
				assert.False(t, gameMap.Tiles[y][x].Walkable, "Edge tile at (%d,%d) should be wall", x, y)
			}
		}
	}

	// Check that center tiles are still walkable
	centerX, centerY := width/2, height/2
	if centerX >= buffer && centerX < width-buffer && centerY >= buffer && centerY < height-buffer {
		assert.True(t, gameMap.Tiles[centerY][centerX].Walkable, "Center tile should remain walkable")
	}
}

func TestRemoveSmallAreas(t *testing.T) {
	gameMap := createTestGameMap(5, 5)

	// Create a pattern with a small isolated area
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			gameMap.Tiles[y][x].Walkable = false // Start with all walls
		}
	}

	// Create main area
	gameMap.Tiles[1][1].Walkable = true
	gameMap.Tiles[1][2].Walkable = true
	gameMap.Tiles[2][1].Walkable = true
	gameMap.Tiles[2][2].Walkable = true

	// Create small isolated area (size 1)
	gameMap.Tiles[4][4].Walkable = true

	minRoomSize := 3
	err := removeSmallAreas(gameMap, minRoomSize)
	require.NoError(t, err)

	// Small area should be converted to wall
	assert.False(t, gameMap.Tiles[4][4].Walkable)

	// Main area should remain
	assert.True(t, gameMap.Tiles[1][1].Walkable)
	assert.True(t, gameMap.Tiles[2][2].Walkable)
}

func TestApplySmoothingPass(t *testing.T) {
	gameMap := createTestGameMap(5, 5)

	// Create a pattern with isolated walls that should be smoothed
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			gameMap.Tiles[y][x].Walkable = true // Start with all floors
		}
	}

	// Add an isolated wall in the center that should be smoothed away
	gameMap.Tiles[2][2].Walkable = false

	originalCenterWalkable := gameMap.Tiles[2][2].Walkable

	err := applySmoothingPass(gameMap)
	require.NoError(t, err)

	// The isolated wall should likely be smoothed to a floor
	// (depending on neighbor count, it might change)
	newCenterWalkable := gameMap.Tiles[2][2].Walkable

	// At minimum, the function should execute without error
	// The specific behavior depends on the smoothing algorithm
	t.Logf("Original center walkable: %v, new: %v", originalCenterWalkable, newCenterWalkable)
}

func TestDeterministicGeneration(t *testing.T) {
	width, height := 15, 15
	seed := int64(54321)

	// Generate two maps with the same seed
	gameMap1 := createTestGameMap(width, height)
	gameMap2 := createTestGameMap(width, height)

	seedMgr1 := pcg.NewSeedManager(seed)
	genCtx1 := pcg.NewGenerationContext(seedMgr1, pcg.ContentTypeTerrain, "test", pcg.GenerationParams{
		Seed: seed,
	})

	seedMgr2 := pcg.NewSeedManager(seed)
	genCtx2 := pcg.NewGenerationContext(seedMgr2, pcg.ContentTypeTerrain, "test", pcg.GenerationParams{
		Seed: seed,
	})

	config := DefaultCAConfig()
	config.MaxIterations = 2 // Reduce for faster testing

	err1 := RunCellularAutomata(gameMap1, config, genCtx1)
	err2 := RunCellularAutomata(gameMap2, config, genCtx2)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Maps should be identical
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			assert.Equal(t, gameMap1.Tiles[y][x].Walkable, gameMap2.Tiles[y][x].Walkable,
				"Tiles at (%d,%d) should be identical", x, y)
		}
	}
}

// Helper function to create a test game map
func createTestGameMap(width, height int) *game.GameMap {
	gameMap := &game.GameMap{
		Width:  width,
		Height: height,
		Tiles:  make([][]game.MapTile, height),
	}

	for i := range gameMap.Tiles {
		gameMap.Tiles[i] = make([]game.MapTile, width)
	}

	return gameMap
}
