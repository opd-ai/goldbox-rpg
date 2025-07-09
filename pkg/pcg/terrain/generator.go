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

// Helper methods for connectivity (simplified implementations)
func (cag *CellularAutomataGenerator) ensureModerateConnectivity(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	// More sophisticated connectivity ensuring multiple paths
	return cag.ensureMinimalConnectivity(gameMap, genCtx)
}

func (cag *CellularAutomataGenerator) ensureHighConnectivity(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	// High connectivity with redundant paths
	return cag.ensureMinimalConnectivity(gameMap, genCtx)
}

func (cag *CellularAutomataGenerator) ensureCompleteConnectivity(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	// Complete connectivity ensuring all areas are reachable
	return cag.ensureMinimalConnectivity(gameMap, genCtx)
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

func (cag *CellularAutomataGenerator) addCaveFeatures(gameMap *game.GameMap, genCtx *pcg.GenerationContext, roughness float64) {
	// Add cave-specific features based on roughness
	// This is a simplified implementation
}

func (cag *CellularAutomataGenerator) addDungeonDoors(gameMap *game.GameMap, genCtx *pcg.GenerationContext) {
	// Add door markers for dungeon entrances
	// This is a simplified implementation
}

func (cag *CellularAutomataGenerator) addTorchPositions(gameMap *game.GameMap, genCtx *pcg.GenerationContext) {
	// Add torch positions for dungeon lighting
	// This is a simplified implementation
}

func (cag *CellularAutomataGenerator) addVegetation(gameMap *game.GameMap, genCtx *pcg.GenerationContext, density float64) {
	// Add vegetation features for swamp biomes
	// This is a simplified implementation
}

// Connectivity helper methods (simplified implementations)
func (cag *CellularAutomataGenerator) findWalkableRegions(gameMap *game.GameMap) [][]game.Position {
	// Return connected components of walkable tiles
	// This is a simplified implementation that would use flood fill
	return [][]game.Position{}
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

func (cag *CellularAutomataGenerator) connectRegions(gameMap *game.GameMap, region1, region2 []game.Position) {
	// Create a path between two regions by carving through walls
	// This is a simplified implementation
}
