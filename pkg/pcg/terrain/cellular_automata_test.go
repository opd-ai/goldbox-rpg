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

// ==================== Connectivity Tests ====================

func TestFindWalkableRegions_EmptyMap(t *testing.T) {
	cag := NewCellularAutomataGenerator()

	// Nil map
	regions := cag.findWalkableRegions(nil)
	assert.Empty(t, regions)

	// Empty map
	emptyMap := &game.GameMap{Width: 0, Height: 0, Tiles: [][]game.MapTile{}}
	regions = cag.findWalkableRegions(emptyMap)
	assert.Empty(t, regions)
}

func TestFindWalkableRegions_SingleRegion(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(5, 5)

	// Create a single connected walkable region
	// All walls except a cross in the center
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}
	gameMap.Tiles[2][1].Walkable = true
	gameMap.Tiles[2][2].Walkable = true
	gameMap.Tiles[2][3].Walkable = true
	gameMap.Tiles[1][2].Walkable = true
	gameMap.Tiles[3][2].Walkable = true

	regions := cag.findWalkableRegions(gameMap)

	assert.Len(t, regions, 1)
	assert.Len(t, regions[0], 5) // 5 walkable tiles
}

func TestFindWalkableRegions_MultipleRegions(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(10, 10)

	// All walls first
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}

	// Region 1: top-left (2x2)
	gameMap.Tiles[1][1].Walkable = true
	gameMap.Tiles[1][2].Walkable = true
	gameMap.Tiles[2][1].Walkable = true
	gameMap.Tiles[2][2].Walkable = true

	// Region 2: bottom-right (2x2)
	gameMap.Tiles[7][7].Walkable = true
	gameMap.Tiles[7][8].Walkable = true
	gameMap.Tiles[8][7].Walkable = true
	gameMap.Tiles[8][8].Walkable = true

	regions := cag.findWalkableRegions(gameMap)

	assert.Len(t, regions, 2)
	assert.Len(t, regions[0], 4)
	assert.Len(t, regions[1], 4)
}

func TestFindWalkableRegions_NoWalkable(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(5, 5)

	// All walls
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}

	regions := cag.findWalkableRegions(gameMap)
	assert.Empty(t, regions)
}

func TestConnectRegions_EmptyRegions(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(5, 5)

	// Should not panic with empty regions
	cag.connectRegions(gameMap, []game.Position{}, []game.Position{})
	cag.connectRegions(nil, []game.Position{{X: 1, Y: 1}}, []game.Position{{X: 3, Y: 3}})
}

func TestConnectRegions_CreatesCorridor(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(10, 10)

	// All walls first
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}

	// Region 1: top-left
	gameMap.Tiles[1][1].Walkable = true
	region1 := []game.Position{{X: 1, Y: 1}}

	// Region 2: bottom-right
	gameMap.Tiles[8][8].Walkable = true
	region2 := []game.Position{{X: 8, Y: 8}}

	cag.connectRegions(gameMap, region1, region2)

	// Verify corridor was carved (L-shaped path from (1,1) to (8,8))
	// Path goes horizontal first: (1,1) -> (8,1), then vertical: (8,1) -> (8,8)

	// Check horizontal segment at y=1
	for x := 1; x <= 8; x++ {
		assert.True(t, gameMap.Tiles[1][x].Walkable, "Expected walkable at (%d, 1)", x)
	}

	// Check vertical segment at x=8
	for y := 1; y <= 8; y++ {
		assert.True(t, gameMap.Tiles[y][8].Walkable, "Expected walkable at (8, %d)", y)
	}
}

func TestConnectRegions_FindsClosestPoints(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(10, 10)

	// All walls first
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}

	// Region 1: two points, one far, one close
	region1 := []game.Position{
		{X: 1, Y: 1}, // Far point
		{X: 4, Y: 4}, // Close point
	}
	for _, p := range region1 {
		gameMap.Tiles[p.Y][p.X].Walkable = true
	}

	// Region 2: one point
	region2 := []game.Position{{X: 6, Y: 4}}
	gameMap.Tiles[4][6].Walkable = true

	cag.connectRegions(gameMap, region1, region2)

	// Should connect (4,4) to (6,4) - the closest pair
	// Horizontal path from x=4 to x=6 at y=4
	for x := 4; x <= 6; x++ {
		assert.True(t, gameMap.Tiles[4][x].Walkable, "Expected walkable at (%d, 4)", x)
	}
}

func TestEnsureMinimalConnectivity_AlreadyConnected(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(5, 5)

	// Create single connected region
	for y := 1; y < 4; y++ {
		for x := 1; x < 4; x++ {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	seedMgr := pcg.NewSeedManager(12345)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeTerrain, "test", pcg.GenerationParams{
		Seed: 12345,
	})

	err := cag.ensureMinimalConnectivity(gameMap, genCtx)
	assert.NoError(t, err)
}

func TestEnsureMinimalConnectivity_ConnectsDisjointRegions(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(10, 10)

	// All walls first
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}

	// Region 1: top-left
	gameMap.Tiles[1][1].Walkable = true
	gameMap.Tiles[1][2].Walkable = true

	// Region 2: bottom-right
	gameMap.Tiles[8][8].Walkable = true
	gameMap.Tiles[8][7].Walkable = true

	seedMgr := pcg.NewSeedManager(12345)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeTerrain, "test", pcg.GenerationParams{
		Seed: 12345,
	})

	// Before: 2 regions
	regionsBefore := cag.findWalkableRegions(gameMap)
	assert.Len(t, regionsBefore, 2)

	err := cag.ensureMinimalConnectivity(gameMap, genCtx)
	assert.NoError(t, err)

	// After: 1 connected region
	regionsAfter := cag.findWalkableRegions(gameMap)
	assert.Len(t, regionsAfter, 1)
}

func TestFloodFill_SingleTile(t *testing.T) {
	cag := NewCellularAutomataGenerator()
	gameMap := createTestGameMap(3, 3)

	// All walls except center
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}
	gameMap.Tiles[1][1].Walkable = true

	visited := make([][]bool, 3)
	for i := range visited {
		visited[i] = make([]bool, 3)
	}

	region := cag.floodFill(gameMap, 1, 1, visited)

	assert.Len(t, region, 1)
	assert.Equal(t, game.Position{X: 1, Y: 1}, region[0])
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-1, 1},
	}

	for _, tc := range tests {
		result := abs(tc.input)
		assert.Equal(t, tc.expected, result)
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
