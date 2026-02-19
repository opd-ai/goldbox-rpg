package terrain

import (
	"context"
	"fmt"
	"math"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// CellularAutomataGenerator implements terrain generation using cellular automata
// Particularly effective for generating cave systems and natural-looking dungeons
type CellularAutomataGenerator struct {
	version string
}

// NewCellularAutomataGenerator creates a new cellular automata terrain generator
func NewCellularAutomataGenerator() *CellularAutomataGenerator {
	return &CellularAutomataGenerator{
		version: "1.0.0",
	}
}

// Generate implements the Generator interface
func (cag *CellularAutomataGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
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

	return cag.GenerateTerrain(ctx, width, height, terrainParams)
}

// GenerateTerrain implements the TerrainGenerator interface
func (cag *CellularAutomataGenerator) GenerateTerrain(ctx context.Context, width, height int, params pcg.TerrainParams) (*game.GameMap, error) {
	// Create generation context with seeded RNG
	seedMgr := pcg.NewSeedManager(params.Seed)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeTerrain, "cellular_automata", params.GenerationParams)

	// Initialize the map
	gameMap := &game.GameMap{
		Width:  width,
		Height: height,
		Tiles:  make([][]game.MapTile, height),
	}

	// Initialize tiles array
	for y := 0; y < height; y++ {
		gameMap.Tiles[y] = make([]game.MapTile, width)
	}

	// Generate initial random layout
	cag.generateInitialLayout(gameMap, genCtx, params)

	// Apply cellular automata iterations
	iterations := cag.calculateIterations(params.Difficulty)
	for i := 0; i < iterations; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		cag.applyCellularAutomataStep(gameMap, genCtx)
	}

	// Post-process the map based on biome and parameters
	if err := cag.postProcessMap(gameMap, genCtx, params); err != nil {
		return nil, fmt.Errorf("post-processing failed: %w", err)
	}

	// Apply connectivity requirements
	if params.Connectivity != pcg.ConnectivityNone {
		if err := cag.ensureConnectivity(gameMap, genCtx, params.Connectivity); err != nil {
			return nil, fmt.Errorf("connectivity enforcement failed: %w", err)
		}
	}

	return gameMap, nil
}

// GenerateBiome implements the TerrainGenerator interface
func (cag *CellularAutomataGenerator) GenerateBiome(ctx context.Context, biome pcg.BiomeType, bounds pcg.Rectangle, params pcg.TerrainParams) (*game.GameMap, error) {
	// Adjust parameters based on biome characteristics
	adjustedParams := cag.adjustParamsForBiome(params, biome)

	return cag.GenerateTerrain(ctx, bounds.Width, bounds.Height, adjustedParams)
}

// ValidateConnectivity implements the TerrainGenerator interface
func (cag *CellularAutomataGenerator) ValidateConnectivity(terrain *game.GameMap) bool {
	validator := pcg.NewValidator(true)
	result := validator.ValidateGameMap(terrain)
	return result.IsValid()
}

// GetType implements the Generator interface
func (cag *CellularAutomataGenerator) GetType() pcg.ContentType {
	return pcg.ContentTypeTerrain
}

// GetVersion implements the Generator interface
func (cag *CellularAutomataGenerator) GetVersion() string {
	return cag.version
}

// Validate implements the Generator interface
func (cag *CellularAutomataGenerator) Validate(params pcg.GenerationParams) error {
	validator := pcg.NewValidator(false)

	// Check if terrain parameters are provided
	if _, ok := params.Constraints["terrain_params"]; !ok {
		return fmt.Errorf("terrain_params must be provided in constraints")
	}

	terrainParams, ok := params.Constraints["terrain_params"].(pcg.TerrainParams)
	if !ok {
		return fmt.Errorf("terrain_params must be of type TerrainParams")
	}

	result := validator.ValidateTerrainParams(terrainParams)
	if !result.IsValid() {
		return fmt.Errorf("validation failed: %v", result.Errors)
	}

	return nil
}

// generateInitialLayout creates the initial random layout for cellular automata
func (cag *CellularAutomataGenerator) generateInitialLayout(gameMap *game.GameMap, genCtx *pcg.GenerationContext, params pcg.TerrainParams) {
	// Use density parameter to determine fill probability
	fillProbability := params.Density
	if fillProbability <= 0 {
		fillProbability = 0.45 // Default density for caves
	}

	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			// Edge tiles are always walls for boundary
			if x == 0 || y == 0 || x == gameMap.Width-1 || y == gameMap.Height-1 {
				gameMap.Tiles[y][x] = cag.createWallTile()
			} else {
				// Random fill based on density
				if genCtx.RandomFloat() < fillProbability {
					gameMap.Tiles[y][x] = cag.createWallTile()
				} else {
					gameMap.Tiles[y][x] = cag.createFloorTile()
				}
			}
		}
	}
}

// applyCellularAutomataStep applies one iteration of the cellular automata algorithm
func (cag *CellularAutomataGenerator) applyCellularAutomataStep(gameMap *game.GameMap, genCtx *pcg.GenerationContext) {
	newTiles := make([][]game.MapTile, gameMap.Height)
	for y := 0; y < gameMap.Height; y++ {
		newTiles[y] = make([]game.MapTile, gameMap.Width)
	}

	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			wallCount := cag.countAdjacentWalls(gameMap, x, y)

			// Apply cellular automata rules
			if wallCount >= 5 {
				newTiles[y][x] = cag.createWallTile()
			} else if wallCount <= 3 {
				newTiles[y][x] = cag.createFloorTile()
			} else {
				// Keep current state for borderline cases
				newTiles[y][x] = gameMap.Tiles[y][x]
			}
		}
	}

	gameMap.Tiles = newTiles
}

// countAdjacentWalls counts walls in the 3x3 neighborhood around a position
func (cag *CellularAutomataGenerator) countAdjacentWalls(gameMap *game.GameMap, centerX, centerY int) int {
	count := 0

	for y := centerY - 1; y <= centerY+1; y++ {
		for x := centerX - 1; x <= centerX+1; x++ {
			// Count out-of-bounds as walls
			if x < 0 || y < 0 || x >= gameMap.Width || y >= gameMap.Height {
				count++
			} else if !gameMap.Tiles[y][x].Walkable {
				count++
			}
		}
	}

	return count
}

// postProcessMap applies biome-specific post-processing
func (cag *CellularAutomataGenerator) postProcessMap(gameMap *game.GameMap, genCtx *pcg.GenerationContext, params pcg.TerrainParams) error {
	switch params.BiomeType {
	case pcg.BiomeCave:
		return cag.postProcessCave(gameMap, genCtx, params)
	case pcg.BiomeDungeon:
		return cag.postProcessDungeon(gameMap, genCtx, params)
	case pcg.BiomeSwamp:
		return cag.postProcessSwamp(gameMap, genCtx, params)
	default:
		return cag.postProcessGeneric(gameMap, genCtx, params)
	}
}

// postProcessCave applies cave-specific post-processing
func (cag *CellularAutomataGenerator) postProcessCave(gameMap *game.GameMap, genCtx *pcg.GenerationContext, params pcg.TerrainParams) error {
	// Add water pools based on water level parameter
	if params.WaterLevel > 0 {
		cag.addWaterFeatures(gameMap, genCtx, params.WaterLevel)
	}

	// Add stalactites/stalagmites based on roughness
	if params.Roughness > 0.5 {
		cag.addCaveFeatures(gameMap, genCtx, params.Roughness)
	}

	return nil
}

// postProcessDungeon applies dungeon-specific post-processing
func (cag *CellularAutomataGenerator) postProcessDungeon(gameMap *game.GameMap, genCtx *pcg.GenerationContext, params pcg.TerrainParams) error {
	// Add door positions for dungeon rooms
	cag.addDungeonDoors(gameMap, genCtx)

	// Add torch positions for lighting
	cag.addTorchPositions(gameMap, genCtx)

	return nil
}

// postProcessSwamp applies swamp-specific post-processing
func (cag *CellularAutomataGenerator) postProcessSwamp(gameMap *game.GameMap, genCtx *pcg.GenerationContext, params pcg.TerrainParams) error {
	// Add extensive water features
	cag.addWaterFeatures(gameMap, genCtx, math.Max(params.WaterLevel, 0.3))

	// Add vegetation density
	cag.addVegetation(gameMap, genCtx, 0.7)

	return nil
}

// postProcessGeneric applies generic post-processing
func (cag *CellularAutomataGenerator) postProcessGeneric(gameMap *game.GameMap, genCtx *pcg.GenerationContext, params pcg.TerrainParams) error {
	// Apply water level if specified
	if params.WaterLevel > 0 {
		cag.addWaterFeatures(gameMap, genCtx, params.WaterLevel)
	}

	return nil
}

// ensureConnectivity ensures the map meets connectivity requirements
func (cag *CellularAutomataGenerator) ensureConnectivity(gameMap *game.GameMap, genCtx *pcg.GenerationContext, level pcg.ConnectivityLevel) error {
	switch level {
	case pcg.ConnectivityMinimal:
		return cag.ensureMinimalConnectivity(gameMap, genCtx)
	case pcg.ConnectivityModerate:
		return cag.ensureModerateConnectivity(gameMap, genCtx)
	case pcg.ConnectivityHigh:
		return cag.ensureHighConnectivity(gameMap, genCtx)
	case pcg.ConnectivityComplete:
		return cag.ensureCompleteConnectivity(gameMap, genCtx)
	default:
		return nil
	}
}

// ensureMinimalConnectivity ensures basic connectivity between major areas
func (cag *CellularAutomataGenerator) ensureMinimalConnectivity(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	// Find all disconnected walkable regions
	regions := cag.findWalkableRegions(gameMap)

	if len(regions) <= 1 {
		return nil // Already connected or no walkable areas
	}

	// Connect the largest regions
	mainRegion := cag.findLargestRegion(regions)
	for i, region := range regions {
		if i != mainRegion {
			cag.connectRegions(gameMap, regions[mainRegion], region)
		}
	}

	return nil
}

// ensureModerateConnectivity connects all regions to the main region and adds
// 1-2 random redundant connections between smaller regions for basic path redundancy.
func (cag *CellularAutomataGenerator) ensureModerateConnectivity(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	regions := cag.findWalkableRegions(gameMap)
	if len(regions) <= 1 {
		return nil
	}

	// First, connect all regions to the main region (like minimal)
	mainRegion := cag.findLargestRegion(regions)
	for i, region := range regions {
		if i != mainRegion {
			cag.connectRegions(gameMap, regions[mainRegion], region)
		}
	}

	// Add 1-2 redundant connections between non-main regions
	if len(regions) > 2 {
		redundantCount := 1
		if len(regions) > 4 {
			redundantCount = 2
		}
		for r := 0; r < redundantCount; r++ {
			// Pick two random non-main regions
			idx1 := genCtx.RandomIntRange(0, len(regions)-1)
			idx2 := genCtx.RandomIntRange(0, len(regions)-1)
			for idx1 == mainRegion {
				idx1 = genCtx.RandomIntRange(0, len(regions)-1)
			}
			for idx2 == mainRegion || idx2 == idx1 {
				idx2 = genCtx.RandomIntRange(0, len(regions)-1)
			}
			cag.connectRegions(gameMap, regions[idx1], regions[idx2])
		}
	}

	return nil
}

// ensureHighConnectivity connects all regions to the main region and also connects
// each region to its nearest neighbor, creating a web of connections with multiple paths.
func (cag *CellularAutomataGenerator) ensureHighConnectivity(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	regions := cag.findWalkableRegions(gameMap)
	if len(regions) <= 1 {
		return nil
	}

	// Connect all regions to the main region
	mainRegion := cag.findLargestRegion(regions)
	for i, region := range regions {
		if i != mainRegion {
			cag.connectRegions(gameMap, regions[mainRegion], region)
		}
	}

	// Connect each region to its nearest neighbor (not just main)
	for i := range regions {
		nearestIdx := cag.findNearestRegion(regions, i)
		if nearestIdx != -1 && nearestIdx != i {
			cag.connectRegions(gameMap, regions[i], regions[nearestIdx])
		}
	}

	return nil
}

// ensureCompleteConnectivity connects all regions to the main region and then connects
// each region to all neighbors within a threshold distance for maximum traversability.
func (cag *CellularAutomataGenerator) ensureCompleteConnectivity(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	regions := cag.findWalkableRegions(gameMap)
	if len(regions) <= 1 {
		return nil
	}

	// Connect all regions to the main region
	mainRegion := cag.findLargestRegion(regions)
	for i, region := range regions {
		if i != mainRegion {
			cag.connectRegions(gameMap, regions[mainRegion], region)
		}
	}

	// Calculate average region-to-region distance for threshold
	threshold := cag.calculateConnectionThreshold(regions, gameMap.Width, gameMap.Height)

	// Connect each region to all neighbors within the threshold
	for i := 0; i < len(regions); i++ {
		for j := i + 1; j < len(regions); j++ {
			dist := cag.regionDistance(regions[i], regions[j])
			if dist <= threshold {
				cag.connectRegions(gameMap, regions[i], regions[j])
			}
		}
	}

	return nil
}

// findNearestRegion finds the index of the nearest region to the given region index.
// Returns -1 if no other region exists.
func (cag *CellularAutomataGenerator) findNearestRegion(regions [][]game.Position, sourceIdx int) int {
	if len(regions) < 2 || sourceIdx < 0 || sourceIdx >= len(regions) {
		return -1
	}

	nearestIdx := -1
	nearestDist := int(^uint(0) >> 1) // Max int

	for i, region := range regions {
		if i == sourceIdx {
			continue
		}
		dist := cag.regionDistance(regions[sourceIdx], region)
		if dist < nearestDist {
			nearestDist = dist
			nearestIdx = i
		}
	}

	return nearestIdx
}

// regionDistance calculates the minimum Manhattan distance between two regions.
func (cag *CellularAutomataGenerator) regionDistance(region1, region2 []game.Position) int {
	if len(region1) == 0 || len(region2) == 0 {
		return int(^uint(0) >> 1) // Max int
	}

	minDist := int(^uint(0) >> 1)
	for _, p1 := range region1 {
		for _, p2 := range region2 {
			dist := abs(p1.X-p2.X) + abs(p1.Y-p2.Y)
			if dist < minDist {
				minDist = dist
			}
		}
	}
	return minDist
}

// calculateConnectionThreshold returns a distance threshold for complete connectivity.
// Regions within this distance should be connected.
func (cag *CellularAutomataGenerator) calculateConnectionThreshold(regions [][]game.Position, mapWidth, mapHeight int) int {
	if len(regions) < 2 {
		return 0
	}

	// Use map diagonal / number of regions as a baseline threshold
	// This ensures we connect nearby regions without connecting everything to everything
	diagonal := int(math.Sqrt(float64(mapWidth*mapWidth + mapHeight*mapHeight)))
	threshold := diagonal / len(regions)

	// Minimum threshold to ensure some connectivity
	if threshold < 10 {
		threshold = 10
	}

	return threshold
}

// Utility methods for tile creation
func (cag *CellularAutomataGenerator) createWallTile() game.MapTile {
	return game.MapTile{
		SpriteX:     1, // Wall sprite coordinates
		SpriteY:     0,
		Walkable:    false,
		Transparent: false,
	}
}

func (cag *CellularAutomataGenerator) createFloorTile() game.MapTile {
	return game.MapTile{
		SpriteX:     0, // Floor sprite coordinates
		SpriteY:     0,
		Walkable:    true,
		Transparent: true,
	}
}

func (cag *CellularAutomataGenerator) createWaterTile() game.MapTile {
	return game.MapTile{
		SpriteX:     2, // Water sprite coordinates
		SpriteY:     0,
		Walkable:    false, // Water is not walkable by default
		Transparent: true,
	}
}

// Helper methods (simplified implementations)
func (cag *CellularAutomataGenerator) calculateIterations(difficulty int) int {
	// More difficult areas get more iterations for complexity
	return 4 + (difficulty / 5)
}

func (cag *CellularAutomataGenerator) adjustParamsForBiome(params pcg.TerrainParams, biome pcg.BiomeType) pcg.TerrainParams {
	adjusted := params

	switch biome {
	case pcg.BiomeCave:
		adjusted.Density = 0.45
		adjusted.WaterLevel = 0.1
	case pcg.BiomeDungeon:
		adjusted.Density = 0.4
		adjusted.WaterLevel = 0.05
	case pcg.BiomeSwamp:
		adjusted.Density = 0.3
		adjusted.WaterLevel = 0.4
	}

	return adjusted
}

// Feature addition methods (simplified implementations)
func (cag *CellularAutomataGenerator) addWaterFeatures(gameMap *game.GameMap, genCtx *pcg.GenerationContext, waterLevel float64) {
	// Add water tiles to low-lying floor areas
	for y := 1; y < gameMap.Height-1; y++ {
		for x := 1; x < gameMap.Width-1; x++ {
			if gameMap.Tiles[y][x].Walkable && genCtx.RandomFloat() < waterLevel {
				gameMap.Tiles[y][x] = cag.createWaterTile()
			}
		}
	}
}

// addCaveFeatures adds stalactites/stalagmites and rocky debris based on roughness.
// Higher roughness values result in more features being placed near walls.
func (cag *CellularAutomataGenerator) addCaveFeatures(gameMap *game.GameMap, genCtx *pcg.GenerationContext, roughness float64) {
	if gameMap == nil || genCtx == nil || roughness <= 0 {
		return
	}

	for y := 1; y < gameMap.Height-1; y++ {
		for x := 1; x < gameMap.Width-1; x++ {
			tile := &gameMap.Tiles[y][x]
			if !tile.Walkable {
				continue
			}

			// Count adjacent walls to determine if this is near a cave wall
			wallCount := cag.countAdjacentWalls(gameMap, x, y)
			if wallCount == 0 {
				continue // Not near any walls
			}

			// Probability of placing a feature increases with wall adjacency and roughness
			featureProb := roughness * float64(wallCount) * 0.05
			if genCtx.RandomFloat() < featureProb {
				// Mark tile as having cave decoration (rocky debris/stalagmite)
				// Uses sprite coordinates to indicate decorated floor
				tile.SpriteX = 3 // Decorated floor sprite
				tile.SpriteY = 1
			}
		}
	}
}

// addDungeonDoors places door markers at narrow passages between rooms.
// Doors are placed where corridors connect to larger open areas.
func (cag *CellularAutomataGenerator) addDungeonDoors(gameMap *game.GameMap, genCtx *pcg.GenerationContext) {
	if gameMap == nil || genCtx == nil {
		return
	}

	for y := 2; y < gameMap.Height-2; y++ {
		for x := 2; x < gameMap.Width-2; x++ {
			if !gameMap.Tiles[y][x].Walkable {
				continue
			}

			// Check for horizontal doorway pattern: wall above and below, open left and right
			isHorizontalDoorway := !gameMap.Tiles[y-1][x].Walkable &&
				!gameMap.Tiles[y+1][x].Walkable &&
				gameMap.Tiles[y][x-1].Walkable &&
				gameMap.Tiles[y][x+1].Walkable

			// Check for vertical doorway pattern: wall left and right, open above and below
			isVerticalDoorway := !gameMap.Tiles[y][x-1].Walkable &&
				!gameMap.Tiles[y][x+1].Walkable &&
				gameMap.Tiles[y-1][x].Walkable &&
				gameMap.Tiles[y+1][x].Walkable

			if isHorizontalDoorway || isVerticalDoorway {
				// Place door with some randomness (not every valid position gets a door)
				if genCtx.RandomFloat() < 0.4 {
					// Mark as door tile
					gameMap.Tiles[y][x].SpriteX = 4 // Door sprite
					gameMap.Tiles[y][x].SpriteY = 0
				}
			}
		}
	}
}

// addTorchPositions places torches on walls adjacent to walkable areas.
// Torches are spaced out to provide even lighting coverage.
func (cag *CellularAutomataGenerator) addTorchPositions(gameMap *game.GameMap, genCtx *pcg.GenerationContext) {
	if gameMap == nil || genCtx == nil {
		return
	}

	// Track torch positions to ensure minimum spacing
	const minTorchSpacing = 4

	for y := 1; y < gameMap.Height-1; y++ {
		for x := 1; x < gameMap.Width-1; x++ {
			tile := &gameMap.Tiles[y][x]
			// Only place torches on wall tiles adjacent to walkable areas
			if tile.Walkable {
				continue
			}

			// Check if this wall is adjacent to a walkable tile
			hasAdjacentFloor := false
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < gameMap.Width && ny >= 0 && ny < gameMap.Height {
						if gameMap.Tiles[ny][nx].Walkable {
							hasAdjacentFloor = true
							break
						}
					}
				}
				if hasAdjacentFloor {
					break
				}
			}

			if !hasAdjacentFloor {
				continue
			}

			// Check spacing from other torches (identified by sprite)
			tooClose := false
			for dy := -minTorchSpacing; dy <= minTorchSpacing && !tooClose; dy++ {
				for dx := -minTorchSpacing; dx <= minTorchSpacing && !tooClose; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < gameMap.Width && ny >= 0 && ny < gameMap.Height {
						if gameMap.Tiles[ny][nx].SpriteX == 5 && gameMap.Tiles[ny][nx].SpriteY == 0 {
							tooClose = true
						}
					}
				}
			}

			if tooClose {
				continue
			}

			// Place torch with some randomness
			if genCtx.RandomFloat() < 0.3 {
				tile.SpriteX = 5 // Torch sprite
				tile.SpriteY = 0
			}
		}
	}
}

// addVegetation places vegetation features (grass, reeds, vines) on floor tiles.
// Higher density values result in more vegetation coverage.
func (cag *CellularAutomataGenerator) addVegetation(gameMap *game.GameMap, genCtx *pcg.GenerationContext, density float64) {
	if gameMap == nil || genCtx == nil || density <= 0 {
		return
	}

	for y := 1; y < gameMap.Height-1; y++ {
		for x := 1; x < gameMap.Width-1; x++ {
			tile := &gameMap.Tiles[y][x]
			if !tile.Walkable {
				continue
			}

			// Skip water tiles (sprite coordinates 2,0)
			if tile.SpriteX == 2 && tile.SpriteY == 0 {
				// Place reeds near water with higher probability
				if genCtx.RandomFloat() < density*0.5 {
					// Check adjacent tiles for more water (creates clusters near water)
					waterCount := 0
					for dy := -1; dy <= 1; dy++ {
						for dx := -1; dx <= 1; dx++ {
							nx, ny := x+dx, y+dy
							if nx >= 0 && nx < gameMap.Width && ny >= 0 && ny < gameMap.Height {
								adjTile := &gameMap.Tiles[ny][nx]
								if adjTile.SpriteX == 2 && adjTile.SpriteY == 0 {
									waterCount++
								}
							}
						}
					}
					if waterCount > 0 && genCtx.RandomFloat() < float64(waterCount)*0.15 {
						tile.SpriteX = 6 // Reeds sprite
						tile.SpriteY = 1
					}
				}
				continue
			}

			// Place general vegetation on regular floor tiles
			if genCtx.RandomFloat() < density {
				// Vary vegetation type based on random value
				vegType := genCtx.RandomFloat()
				switch {
				case vegType < 0.5:
					// Light grass
					tile.SpriteX = 6 // Grass sprite
					tile.SpriteY = 0
				case vegType < 0.8:
					// Dense vegetation
					tile.SpriteX = 7 // Dense vegetation sprite
					tile.SpriteY = 0
				default:
					// Sparse vines/moss
					tile.SpriteX = 7 // Moss sprite
					tile.SpriteY = 1
				}
			}
		}
	}
}

// findWalkableRegions identifies all disconnected walkable regions using flood-fill.
// Returns a slice of regions, where each region is a slice of connected walkable positions.
func (cag *CellularAutomataGenerator) findWalkableRegions(gameMap *game.GameMap) [][]game.Position {
	if gameMap == nil || gameMap.Width == 0 || gameMap.Height == 0 {
		return [][]game.Position{}
	}

	visited := make([][]bool, gameMap.Height)
	for i := range visited {
		visited[i] = make([]bool, gameMap.Width)
	}

	var regions [][]game.Position

	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			if !visited[y][x] && gameMap.Tiles[y][x].Walkable {
				region := cag.floodFill(gameMap, x, y, visited)
				if len(region) > 0 {
					regions = append(regions, region)
				}
			}
		}
	}

	return regions
}

// floodFill performs iterative flood-fill starting from (startX, startY).
// Returns all connected walkable positions.
func (cag *CellularAutomataGenerator) floodFill(gameMap *game.GameMap, startX, startY int, visited [][]bool) []game.Position {
	var region []game.Position
	stack := []game.Position{{X: startX, Y: startY}}

	// 4-directional neighbors
	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}

	for len(stack) > 0 {
		// Pop from stack
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Skip if out of bounds or already visited
		if pos.X < 0 || pos.X >= gameMap.Width || pos.Y < 0 || pos.Y >= gameMap.Height {
			continue
		}
		if visited[pos.Y][pos.X] {
			continue
		}
		if !gameMap.Tiles[pos.Y][pos.X].Walkable {
			continue
		}

		visited[pos.Y][pos.X] = true
		region = append(region, pos)

		// Add neighbors to stack
		for i := 0; i < 4; i++ {
			nx, ny := pos.X+dx[i], pos.Y+dy[i]
			if nx >= 0 && nx < gameMap.Width && ny >= 0 && ny < gameMap.Height && !visited[ny][nx] {
				stack = append(stack, game.Position{X: nx, Y: ny})
			}
		}
	}

	return region
}

func (cag *CellularAutomataGenerator) findLargestRegion(regions [][]game.Position) int {
	maxSize := 0
	maxIndex := 0

	for i, region := range regions {
		if len(region) > maxSize {
			maxSize = len(region)
			maxIndex = i
		}
	}

	return maxIndex
}

// connectRegions creates a corridor between two disconnected regions.
// Uses L-shaped corridor through the closest points of each region.
func (cag *CellularAutomataGenerator) connectRegions(gameMap *game.GameMap, region1, region2 []game.Position) {
	if len(region1) == 0 || len(region2) == 0 || gameMap == nil {
		return
	}

	// Find closest pair of points between the two regions
	var bestP1, bestP2 game.Position
	bestDist := int(^uint(0) >> 1) // Max int

	for _, p1 := range region1 {
		for _, p2 := range region2 {
			dist := abs(p1.X-p2.X) + abs(p1.Y-p2.Y) // Manhattan distance
			if dist < bestDist {
				bestDist = dist
				bestP1 = p1
				bestP2 = p2
			}
		}
	}

	// Carve L-shaped corridor: horizontal first, then vertical
	cag.carveCorridor(gameMap, bestP1.X, bestP1.Y, bestP2.X, bestP2.Y)
}

// carveCorridor creates an L-shaped path between two points.
func (cag *CellularAutomataGenerator) carveCorridor(gameMap *game.GameMap, x1, y1, x2, y2 int) {
	// Carve horizontal segment from (x1, y1) to (x2, y1)
	startX, endX := x1, x2
	if startX > endX {
		startX, endX = endX, startX
	}
	for x := startX; x <= endX; x++ {
		cag.carveFloorAt(gameMap, x, y1)
	}

	// Carve vertical segment from (x2, y1) to (x2, y2)
	startY, endY := y1, y2
	if startY > endY {
		startY, endY = endY, startY
	}
	for y := startY; y <= endY; y++ {
		cag.carveFloorAt(gameMap, x2, y)
	}
}

// carveFloorAt sets a tile to walkable floor if within bounds.
func (cag *CellularAutomataGenerator) carveFloorAt(gameMap *game.GameMap, x, y int) {
	if x >= 0 && x < gameMap.Width && y >= 0 && y < gameMap.Height {
		gameMap.Tiles[y][x] = cag.createFloorTile()
	}
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
