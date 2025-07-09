package utils

import (
	"container/heap"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAStarPathfind(t *testing.T) {
	// Create a simple 5x5 map with a clear path
	gameMap := createTestMap(5, 5)

	// Set all tiles as walkable
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	start := game.Position{X: 0, Y: 0}
	goal := game.Position{X: 4, Y: 4}

	result := AStarPathfind(gameMap, start, goal)

	require.NotNil(t, result)
	assert.True(t, result.Found)
	assert.NotEmpty(t, result.Path)
	assert.Equal(t, start, result.Path[0])
	assert.Equal(t, goal, result.Path[len(result.Path)-1])
	assert.Greater(t, result.Distance, 0)
}

func TestAStarPathfindWithObstacles(t *testing.T) {
	gameMap := createTestMap(5, 5)

	// Set all tiles as walkable first
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	// Add a wall in the middle
	gameMap.Tiles[2][1].Walkable = false
	gameMap.Tiles[2][2].Walkable = false
	gameMap.Tiles[2][3].Walkable = false

	start := game.Position{X: 0, Y: 2}
	goal := game.Position{X: 4, Y: 2}

	result := AStarPathfind(gameMap, start, goal)

	require.NotNil(t, result)
	assert.True(t, result.Found)
	assert.NotEmpty(t, result.Path)

	// Path should go around the obstacle
	assert.Greater(t, result.Distance, 4) // Direct distance would be 4
}

func TestAStarPathfindNoPath(t *testing.T) {
	gameMap := createTestMap(5, 5)

	// Set all tiles as walls except start and goal
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}

	// Make start and goal walkable but isolated
	gameMap.Tiles[0][0].Walkable = true
	gameMap.Tiles[4][4].Walkable = true

	start := game.Position{X: 0, Y: 0}
	goal := game.Position{X: 4, Y: 4}

	result := AStarPathfind(gameMap, start, goal)

	require.NotNil(t, result)
	assert.False(t, result.Found)
}

func TestAStarPathfindInvalidPositions(t *testing.T) {
	gameMap := createTestMap(3, 3)

	// Test out-of-bounds start
	result := AStarPathfind(gameMap, game.Position{X: -1, Y: 0}, game.Position{X: 1, Y: 1})
	assert.False(t, result.Found)

	// Test out-of-bounds goal
	result = AStarPathfind(gameMap, game.Position{X: 0, Y: 0}, game.Position{X: 5, Y: 5})
	assert.False(t, result.Found)

	// Test unwalkable start
	gameMap.Tiles[0][0].Walkable = false
	gameMap.Tiles[1][1].Walkable = true
	result = AStarPathfind(gameMap, game.Position{X: 0, Y: 0}, game.Position{X: 1, Y: 1})
	assert.False(t, result.Found)
}

func TestFloodFill(t *testing.T) {
	gameMap := createTestMap(4, 4)

	// Create a connected area
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	// Add some walls to create two disconnected areas
	gameMap.Tiles[1][1].Walkable = false
	gameMap.Tiles[1][2].Walkable = false
	gameMap.Tiles[2][1].Walkable = false
	gameMap.Tiles[2][2].Walkable = false

	start := game.Position{X: 0, Y: 0}
	result := FloodFill(gameMap, start)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, start)

	// Should not contain wall positions
	wallPos1 := game.Position{X: 1, Y: 1}
	wallPos2 := game.Position{X: 1, Y: 2}

	found1, found2 := false, false
	for _, pos := range result {
		if pos.X == wallPos1.X && pos.Y == wallPos1.Y {
			found1 = true
		}
		if pos.X == wallPos2.X && pos.Y == wallPos2.Y {
			found2 = true
		}
	}
	assert.False(t, found1, "Should not contain wall position (1,1)")
	assert.False(t, found2, "Should not contain wall position (1,2)")
}

func TestFloodFillInvalidStart(t *testing.T) {
	gameMap := createTestMap(3, 3)

	// Test out-of-bounds start
	result := FloodFill(gameMap, game.Position{X: -1, Y: 0})
	assert.Nil(t, result)

	// Test unwalkable start
	gameMap.Tiles[0][0].Walkable = false
	result = FloodFill(gameMap, game.Position{X: 0, Y: 0})
	assert.Nil(t, result)
}

func TestValidateConnectivity(t *testing.T) {
	gameMap := createTestMap(4, 4)

	// Test fully connected map
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	assert.True(t, ValidateConnectivity(gameMap))

	// Test disconnected map - create two separate areas
	// Area 1: top-left corner
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}

	// Area 1: (0,0) and (0,1)
	gameMap.Tiles[0][0].Walkable = true
	gameMap.Tiles[0][1].Walkable = true

	// Area 2: (3,3) (isolated)
	gameMap.Tiles[3][3].Walkable = true

	// This creates two separate areas, so connectivity should fail
	assert.False(t, ValidateConnectivity(gameMap))
}

func TestValidateConnectivityEmptyMap(t *testing.T) {
	gameMap := createTestMap(3, 3)

	// All walls - should be considered connected (vacuously true)
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			gameMap.Tiles[y][x].Walkable = false
		}
	}

	assert.True(t, ValidateConnectivity(gameMap))
}

func TestManhattanDistance(t *testing.T) {
	a := game.Position{X: 0, Y: 0}
	b := game.Position{X: 3, Y: 4}

	distance := manhattanDistance(a, b)
	assert.Equal(t, 7, distance) // |3-0| + |4-0| = 7

	// Test same position
	distance = manhattanDistance(a, a)
	assert.Equal(t, 0, distance)

	// Test negative coordinates
	c := game.Position{X: -2, Y: -3}
	distance = manhattanDistance(a, c)
	assert.Equal(t, 5, distance) // |0-(-2)| + |0-(-3)| = 5
}

func TestGetNeighbors(t *testing.T) {
	gameMap := createTestMap(3, 3)

	// Set all tiles as walkable
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			gameMap.Tiles[y][x].Walkable = true
		}
	}

	// Test center position
	neighbors := getNeighbors(gameMap, game.Position{X: 1, Y: 1})
	assert.Len(t, neighbors, 4) // Should have 4 neighbors

	// Test corner position
	neighbors = getNeighbors(gameMap, game.Position{X: 0, Y: 0})
	assert.Len(t, neighbors, 2) // Should have 2 neighbors

	// Test with obstacles
	gameMap.Tiles[0][1].Walkable = false
	neighbors = getNeighbors(gameMap, game.Position{X: 0, Y: 0})
	assert.Len(t, neighbors, 1) // Should have 1 neighbor now
}

func TestIsValidPosition(t *testing.T) {
	gameMap := createTestMap(3, 3)

	// Valid positions
	assert.True(t, isValidPosition(gameMap, game.Position{X: 0, Y: 0}))
	assert.True(t, isValidPosition(gameMap, game.Position{X: 2, Y: 2}))
	assert.True(t, isValidPosition(gameMap, game.Position{X: 1, Y: 1}))

	// Invalid positions
	assert.False(t, isValidPosition(gameMap, game.Position{X: -1, Y: 0}))
	assert.False(t, isValidPosition(gameMap, game.Position{X: 0, Y: -1}))
	assert.False(t, isValidPosition(gameMap, game.Position{X: 3, Y: 0}))
	assert.False(t, isValidPosition(gameMap, game.Position{X: 0, Y: 3}))
}

func TestPriorityQueue(t *testing.T) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	// Add nodes with different F values
	node1 := &Node{Position: game.Position{X: 0, Y: 0}, F: 10}
	node2 := &Node{Position: game.Position{X: 1, Y: 1}, F: 5}
	node3 := &Node{Position: game.Position{X: 2, Y: 2}, F: 15}

	heap.Push(pq, node1)
	heap.Push(pq, node2)
	heap.Push(pq, node3)

	// Should pop in order of F value (lowest first)
	popped := heap.Pop(pq).(*Node)
	assert.Equal(t, 5, popped.F)

	popped = heap.Pop(pq).(*Node)
	assert.Equal(t, 10, popped.F)

	popped = heap.Pop(pq).(*Node)
	assert.Equal(t, 15, popped.F)
}

// Helper function to create a test game map
func createTestMap(width, height int) *game.GameMap {
	gameMap := &game.GameMap{
		Width:  width,
		Height: height,
		Tiles:  make([][]game.MapTile, height),
	}

	for i := range gameMap.Tiles {
		gameMap.Tiles[i] = make([]game.MapTile, width)
		// Initialize all tiles as walkable by default
		for j := range gameMap.Tiles[i] {
			gameMap.Tiles[i][j].Walkable = true
		}
	}

	return gameMap
}
