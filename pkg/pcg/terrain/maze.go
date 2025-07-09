package terrain

import (
	"context"
	"fmt"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// MazeGenerator creates maze-like terrain structures
type MazeGenerator struct {
	version string
}

// NewMazeGenerator creates a new maze terrain generator
func NewMazeGenerator() *MazeGenerator {
	return &MazeGenerator{version: "1.0.0"}
}

// Generate implements the Generator interface for maze terrain
func (mg *MazeGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
	// Extract terrain-specific parameters
	terrainParams, ok := params.Constraints["terrain_params"].(pcg.TerrainParams)
	if !ok {
		return nil, fmt.Errorf("missing or invalid terrain parameters")
	}

	// Extract dimensions from constraints
	width, ok := params.Constraints["width"].(int)
	if !ok {
		width = 50 // Default width
	}

	height, ok := params.Constraints["height"].(int)
	if !ok {
		height = 50 // Default height
	}

	return mg.GenerateTerrain(ctx, width, height, terrainParams)
}

// GetType implements the Generator interface
func (mg *MazeGenerator) GetType() pcg.ContentType {
	return pcg.ContentTypeTerrain
}

// GetVersion implements the Generator interface
func (mg *MazeGenerator) GetVersion() string {
	return mg.version
}

// Validate implements the Generator interface
func (mg *MazeGenerator) Validate(params pcg.GenerationParams) error {
	// Check if required constraints are present
	if params.Constraints == nil {
		return fmt.Errorf("constraints required for maze generation")
	}

	// Validate dimensions
	if width, ok := params.Constraints["width"].(int); ok && width < 5 {
		return fmt.Errorf("width must be at least 5 for maze generation")
	}

	if height, ok := params.Constraints["height"].(int); ok && height < 5 {
		return fmt.Errorf("height must be at least 5 for maze generation")
	}

	return nil
}

// GenerateTerrain creates maze-style terrain using recursive backtracking
func (mg *MazeGenerator) GenerateTerrain(ctx context.Context, width, height int, params pcg.TerrainParams) (*game.GameMap, error) {
	// Create generation context
	seedMgr := pcg.NewSeedManager(params.Seed)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeTerrain, "maze", params.GenerationParams)

	// Apply biome modifications
	if err := ApplyBiomeModifications(&params, params.BiomeType); err != nil {
		return nil, fmt.Errorf("failed to apply biome modifications: %w", err)
	}

	// Initialize game map
	gameMap := &game.GameMap{
		Width:  width,
		Height: height,
		Tiles:  make([][]game.MapTile, height),
	}

	for i := range gameMap.Tiles {
		gameMap.Tiles[i] = make([]game.MapTile, width)
	}

	// Step 1: Create grid with all walls
	if err := mg.initializeAllWalls(gameMap); err != nil {
		return nil, fmt.Errorf("failed to initialize walls: %w", err)
	}

	// Step 2: Use recursive backtracking to carve passages
	if err := mg.recursiveBacktrackMaze(gameMap, genCtx); err != nil {
		return nil, fmt.Errorf("failed to generate maze: %w", err)
	}

	// Step 3: Add rooms and special features based on biome
	if err := mg.addSpecialFeatures(gameMap, params, genCtx); err != nil {
		return nil, fmt.Errorf("failed to add special features: %w", err)
	}

	// Step 4: Apply biome-specific modifications
	if err := mg.applyBiomeSpecificFeatures(gameMap, params, genCtx); err != nil {
		return nil, fmt.Errorf("failed to apply biome features: %w", err)
	}

	return gameMap, nil
}

// ValidateConnectivity implements the TerrainGenerator interface
func (mg *MazeGenerator) ValidateConnectivity(terrain *game.GameMap) bool {
	// Find first walkable tile
	var start *game.Position
	for y := 0; y < terrain.Height && start == nil; y++ {
		for x := 0; x < terrain.Width && start == nil; x++ {
			if terrain.Tiles[y][x].Walkable {
				start = &game.Position{X: x, Y: y}
			}
		}
	}

	if start == nil {
		return false // No walkable tiles
	}

	// Count total walkable tiles
	totalWalkable := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.Tiles[y][x].Walkable {
				totalWalkable++
			}
		}
	}

	// Flood fill from start position
	visited := make([][]bool, terrain.Height)
	for i := range visited {
		visited[i] = make([]bool, terrain.Width)
	}

	reachable := mg.floodFillCount(terrain, start.X, start.Y, visited)

	// All walkable tiles should be reachable
	return reachable == totalWalkable
}

// GenerateBiome implements the TerrainGenerator interface
func (mg *MazeGenerator) GenerateBiome(ctx context.Context, biome pcg.BiomeType, bounds pcg.Rectangle, params pcg.TerrainParams) (*game.GameMap, error) {
	// Set biome type and generate terrain for the specified bounds
	params.BiomeType = biome
	return mg.GenerateTerrain(ctx, bounds.Width, bounds.Height, params)
}

// initializeAllWalls fills the entire map with walls
func (mg *MazeGenerator) initializeAllWalls(gameMap *game.GameMap) error {
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			gameMap.Tiles[y][x].Walkable = false
			gameMap.Tiles[y][x].Transparent = false
			gameMap.Tiles[y][x].SpriteX = 1 // Wall sprite
			gameMap.Tiles[y][x].SpriteY = 0
		}
	}
	return nil
}

// recursiveBacktrackMaze generates maze using recursive backtracking algorithm
func (mg *MazeGenerator) recursiveBacktrackMaze(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	// Stack for backtracking
	var stack []game.Position
	visited := make([][]bool, gameMap.Height)
	for i := range visited {
		visited[i] = make([]bool, gameMap.Width)
	}

	// Start from an odd coordinate to ensure proper maze structure
	startX, startY := 1, 1
	if startX >= gameMap.Width {
		startX = 0
	}
	if startY >= gameMap.Height {
		startY = 0
	}

	// Mark starting position as passage
	gameMap.Tiles[startY][startX].Walkable = true
	gameMap.Tiles[startY][startX].Transparent = true
	gameMap.Tiles[startY][startX].SpriteX = 0 // Floor sprite
	gameMap.Tiles[startY][startX].SpriteY = 0
	visited[startY][startX] = true

	stack = append(stack, game.Position{X: startX, Y: startY})

	for len(stack) > 0 {
		current := stack[len(stack)-1]

		// Get unvisited neighbors (2 steps away to maintain wall thickness)
		neighbors := mg.getUnvisitedNeighbors(current, visited, gameMap)

		if len(neighbors) > 0 {
			// Choose random neighbor
			neighbor := neighbors[genCtx.RandomIntRange(0, len(neighbors)-1)]

			// Remove wall between current and neighbor
			wallX := (current.X + neighbor.X) / 2
			wallY := (current.Y + neighbor.Y) / 2

			// Carve passage to neighbor
			gameMap.Tiles[neighbor.Y][neighbor.X].Walkable = true
			gameMap.Tiles[neighbor.Y][neighbor.X].Transparent = true
			gameMap.Tiles[neighbor.Y][neighbor.X].SpriteX = 0
			gameMap.Tiles[neighbor.Y][neighbor.X].SpriteY = 0

			// Carve wall between
			gameMap.Tiles[wallY][wallX].Walkable = true
			gameMap.Tiles[wallY][wallX].Transparent = true
			gameMap.Tiles[wallY][wallX].SpriteX = 0
			gameMap.Tiles[wallY][wallX].SpriteY = 0

			visited[neighbor.Y][neighbor.X] = true
			stack = append(stack, neighbor)
		} else {
			// No unvisited neighbors, backtrack
			stack = stack[:len(stack)-1]
		}
	}

	return nil
}

// getUnvisitedNeighbors returns unvisited neighbors that are 2 steps away
func (mg *MazeGenerator) getUnvisitedNeighbors(pos game.Position, visited [][]bool, gameMap *game.GameMap) []game.Position {
	var neighbors []game.Position

	// Check all four directions, 2 steps away
	directions := []struct{ dx, dy int }{
		{0, -2}, // North
		{2, 0},  // East
		{0, 2},  // South
		{-2, 0}, // West
	}

	for _, dir := range directions {
		nx, ny := pos.X+dir.dx, pos.Y+dir.dy

		if nx >= 0 && nx < gameMap.Width && ny >= 0 && ny < gameMap.Height {
			if !visited[ny][nx] {
				neighbors = append(neighbors, game.Position{X: nx, Y: ny})
			}
		}
	}

	return neighbors
}

// addSpecialFeatures adds rooms and special features to the maze
func (mg *MazeGenerator) addSpecialFeatures(gameMap *game.GameMap, params pcg.TerrainParams, genCtx *pcg.GenerationContext) error {
	// Add some larger open areas (rooms) based on difficulty
	roomCount := params.Difficulty / 3
	if roomCount < 1 {
		roomCount = 1
	}
	if roomCount > 5 {
		roomCount = 5
	}

	for i := 0; i < roomCount; i++ {
		roomSize := genCtx.RandomIntRange(3, 7)
		roomX := genCtx.RandomIntRange(1, gameMap.Width-roomSize-1)
		roomY := genCtx.RandomIntRange(1, gameMap.Height-roomSize-1)

		// Create room
		for y := roomY; y < roomY+roomSize; y++ {
			for x := roomX; x < roomX+roomSize; x++ {
				if x < gameMap.Width && y < gameMap.Height {
					gameMap.Tiles[y][x].Walkable = true
					gameMap.Tiles[y][x].Transparent = true
					gameMap.Tiles[y][x].SpriteX = 0
					gameMap.Tiles[y][x].SpriteY = 0
				}
			}
		}
	}

	return nil
}

// applyBiomeSpecificFeatures adds biome-specific elements to the maze
func (mg *MazeGenerator) applyBiomeSpecificFeatures(gameMap *game.GameMap, params pcg.TerrainParams, genCtx *pcg.GenerationContext) error {
	features, err := GetBiomeFeatures(params.BiomeType)
	if err != nil {
		return err
	}

	// Apply each feature with some probability
	for _, feature := range features {
		if genCtx.RandomFloat() < 0.3 { // 30% chance for each feature
			switch feature {
			case pcg.FeatureWater:
				mg.addWaterFeatures(gameMap, genCtx)
			case pcg.FeatureTraps:
				mg.addTrapFeatures(gameMap, genCtx)
			case pcg.FeatureSecretDoors:
				mg.addSecretDoors(gameMap, genCtx)
			}
		}
	}

	return nil
}

// addWaterFeatures adds water tiles to some floor areas
func (mg *MazeGenerator) addWaterFeatures(gameMap *game.GameMap, genCtx *pcg.GenerationContext) {
	waterCount := genCtx.RandomIntRange(2, 8)

	for i := 0; i < waterCount; i++ {
		x := genCtx.RandomIntRange(0, gameMap.Width-1)
		y := genCtx.RandomIntRange(0, gameMap.Height-1)

		if gameMap.Tiles[y][x].Walkable {
			gameMap.Tiles[y][x].SpriteX = 2 // Water sprite
			gameMap.Tiles[y][x].SpriteY = 0
		}
	}
}

// addTrapFeatures marks some floor tiles as potentially dangerous
func (mg *MazeGenerator) addTrapFeatures(gameMap *game.GameMap, genCtx *pcg.GenerationContext) {
	trapCount := genCtx.RandomIntRange(1, 5)

	for i := 0; i < trapCount; i++ {
		x := genCtx.RandomIntRange(0, gameMap.Width-1)
		y := genCtx.RandomIntRange(0, gameMap.Height-1)

		if gameMap.Tiles[y][x].Walkable {
			gameMap.Tiles[y][x].SpriteX = 3 // Trap sprite
			gameMap.Tiles[y][x].SpriteY = 0
		}
	}
}

// addSecretDoors converts some walls to secret passages
func (mg *MazeGenerator) addSecretDoors(gameMap *game.GameMap, genCtx *pcg.GenerationContext) {
	secretCount := genCtx.RandomIntRange(1, 3)

	for i := 0; i < secretCount; i++ {
		x := genCtx.RandomIntRange(1, gameMap.Width-2)
		y := genCtx.RandomIntRange(1, gameMap.Height-2)

		if !gameMap.Tiles[y][x].Walkable {
			// Check if this wall has passages on both sides
			if (gameMap.Tiles[y][x-1].Walkable && gameMap.Tiles[y][x+1].Walkable) ||
				(gameMap.Tiles[y-1][x].Walkable && gameMap.Tiles[y+1][x].Walkable) {
				gameMap.Tiles[y][x].Walkable = true
				gameMap.Tiles[y][x].Transparent = true
				gameMap.Tiles[y][x].SpriteX = 4 // Secret door sprite
				gameMap.Tiles[y][x].SpriteY = 0
			}
		}
	}
}

// floodFillCount performs flood fill and returns the count of reachable tiles
func (mg *MazeGenerator) floodFillCount(gameMap *game.GameMap, startX, startY int, visited [][]bool) int {
	if startX < 0 || startX >= gameMap.Width || startY < 0 || startY >= gameMap.Height {
		return 0
	}

	if visited[startY][startX] || !gameMap.Tiles[startY][startX].Walkable {
		return 0
	}

	visited[startY][startX] = true
	count := 1

	// Check 4-connected neighbors
	count += mg.floodFillCount(gameMap, startX+1, startY, visited)
	count += mg.floodFillCount(gameMap, startX-1, startY, visited)
	count += mg.floodFillCount(gameMap, startX, startY+1, visited)
	count += mg.floodFillCount(gameMap, startX, startY-1, visited)

	return count
}
