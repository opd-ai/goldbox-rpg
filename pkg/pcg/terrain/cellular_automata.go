package terrain

import (
	"fmt"
	"math/rand"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
	"goldbox-rpg/pkg/pcg/utils"
)

// CellularAutomataConfig holds configuration for the algorithm
type CellularAutomataConfig struct {
	WallThreshold   int     `yaml:"wall_threshold"`   // Neighbor count for wall formation
	FloorThreshold  int     `yaml:"floor_threshold"`  // Neighbor count for floor formation
	MaxIterations   int     `yaml:"max_iterations"`   // Maximum CA iterations
	SmoothingPasses int     `yaml:"smoothing_passes"` // Post-processing smoothing
	EdgeBuffer      int     `yaml:"edge_buffer"`      // Border wall thickness
	MinRoomSize     int     `yaml:"min_room_size"`    // Minimum viable room size
	UsePerlinNoise  bool    `yaml:"use_perlin_noise"` // Use Perlin noise for initial layout (vs random)
	NoiseScale      float64 `yaml:"noise_scale"`      // Scale factor for noise sampling
	NoiseThreshold  float64 `yaml:"noise_threshold"`  // Threshold for wall placement from noise
}

// DefaultCAConfig returns default cellular automata configuration
func DefaultCAConfig() *CellularAutomataConfig {
	return &CellularAutomataConfig{
		WallThreshold:   5,
		FloorThreshold:  3,
		MaxIterations:   6,
		SmoothingPasses: 2,
		EdgeBuffer:      1,
		MinRoomSize:     16,
		UsePerlinNoise:  false,
		NoiseScale:      0.1,
		NoiseThreshold:  0.0,
	}
}

// NoiseBasedCAConfig returns a configuration that uses Perlin noise for more organic terrain
func NoiseBasedCAConfig() *CellularAutomataConfig {
	return &CellularAutomataConfig{
		WallThreshold:   5,
		FloorThreshold:  3,
		MaxIterations:   4,
		SmoothingPasses: 1,
		EdgeBuffer:      1,
		MinRoomSize:     16,
		UsePerlinNoise:  true,
		NoiseScale:      0.1,
		NoiseThreshold:  0.0,
	}
}

// RunCellularAutomata executes the cellular automata algorithm
func RunCellularAutomata(gameMap *game.GameMap, config *CellularAutomataConfig, genCtx *pcg.GenerationContext) error {
	if config == nil {
		config = DefaultCAConfig()
	}

	// Step 1: Initialize layout based on config
	var err error
	if config.UsePerlinNoise {
		err = initializePerlinNoise(gameMap, genCtx, config.NoiseScale, config.NoiseThreshold)
	} else {
		err = initializeRandomNoise(gameMap, genCtx)
	}
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Step 2: Apply cellular automata rules for specified iterations
	for i := 0; i < config.MaxIterations; i++ {
		if err := applyCellularAutomataStep(gameMap, config, genCtx.RNG); err != nil {
			return fmt.Errorf("failed CA iteration %d: %w", i, err)
		}
	}

	// Step 3: Remove small disconnected areas
	if err := removeSmallAreas(gameMap, config.MinRoomSize); err != nil {
		return fmt.Errorf("failed to remove small areas: %w", err)
	}

	// Step 4: Apply smoothing passes
	for i := 0; i < config.SmoothingPasses; i++ {
		if err := applySmoothingPass(gameMap); err != nil {
			return fmt.Errorf("failed smoothing pass %d: %w", i, err)
		}
	}

	// Step 5: Ensure proper edge boundaries
	if err := enforceEdgeBoundaries(gameMap, config.EdgeBuffer); err != nil {
		return fmt.Errorf("failed to enforce edge boundaries: %w", err)
	}

	return nil
}

// initializeRandomNoise fills the map with random noise based on density
func initializeRandomNoise(gameMap *game.GameMap, genCtx *pcg.GenerationContext) error {
	// For terrain generation, we need a density parameter
	// This should come from the terrain parameters
	density := 0.45 // Default density for cave generation

	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			if genCtx.RandomFloat() < density {
				gameMap.Tiles[y][x].Walkable = false
				gameMap.Tiles[y][x].Transparent = false
				gameMap.Tiles[y][x].SpriteX = 1 // Wall sprite
				gameMap.Tiles[y][x].SpriteY = 0
			} else {
				gameMap.Tiles[y][x].Walkable = true
				gameMap.Tiles[y][x].Transparent = true
				gameMap.Tiles[y][x].SpriteX = 0 // Floor sprite
				gameMap.Tiles[y][x].SpriteY = 0
			}
		}
	}

	return nil
}

// initializePerlinNoise fills the map using Perlin noise for more organic terrain patterns.
// The noise creates coherent clusters of walls and floors rather than pure random distribution.
// scale controls the frequency of noise features, threshold controls wall density.
func initializePerlinNoise(gameMap *game.GameMap, genCtx *pcg.GenerationContext, scale, threshold float64) error {
	if gameMap == nil || genCtx == nil {
		return fmt.Errorf("nil gameMap or generation context")
	}

	// Use seed from generation context for deterministic noise
	noise := utils.NewPerlinNoise(genCtx.Seed)

	// Default scale if not set
	if scale <= 0 {
		scale = 0.1
	}

	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			// Sample noise at this position
			noiseVal := noise.Noise2D(float64(x)*scale, float64(y)*scale)

			// Perlin noise returns values in roughly [-1, 1] range
			// Map to [0, 1] for threshold comparison
			normalizedNoise := (noiseVal + 1.0) / 2.0

			if normalizedNoise < threshold+0.45 { // ~45% wall density at threshold=0
				gameMap.Tiles[y][x].Walkable = false
				gameMap.Tiles[y][x].Transparent = false
				gameMap.Tiles[y][x].SpriteX = 1 // Wall sprite
				gameMap.Tiles[y][x].SpriteY = 0
			} else {
				gameMap.Tiles[y][x].Walkable = true
				gameMap.Tiles[y][x].Transparent = true
				gameMap.Tiles[y][x].SpriteX = 0 // Floor sprite
				gameMap.Tiles[y][x].SpriteY = 0
			}
		}
	}

	return nil
}

// applyCellularAutomataStep applies one iteration of the cellular automata rules
func applyCellularAutomataStep(gameMap *game.GameMap, config *CellularAutomataConfig, rng *rand.Rand) error {
	newTiles := make([][]game.MapTile, gameMap.Height)
	for i := range newTiles {
		newTiles[i] = make([]game.MapTile, gameMap.Width)
		copy(newTiles[i], gameMap.Tiles[i])
	}

	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			wallCount := countNeighborWalls(gameMap, x, y)

			if wallCount >= config.WallThreshold {
				newTiles[y][x].Walkable = false
				newTiles[y][x].Transparent = false
				newTiles[y][x].SpriteX = 1 // Wall sprite
				newTiles[y][x].SpriteY = 0
			} else if wallCount <= config.FloorThreshold {
				newTiles[y][x].Walkable = true
				newTiles[y][x].Transparent = true
				newTiles[y][x].SpriteX = 0 // Floor sprite
				newTiles[y][x].SpriteY = 0
			}
			// Tiles with neighbor counts between thresholds remain unchanged
		}
	}

	gameMap.Tiles = newTiles
	return nil
}

// countNeighborWalls counts wall tiles in the 8-neighborhood around a position
func countNeighborWalls(gameMap *game.GameMap, x, y int) int {
	wallCount := 0

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue // Skip the center tile
			}

			nx, ny := x+dx, y+dy

			// Treat out-of-bounds as walls
			if nx < 0 || nx >= gameMap.Width || ny < 0 || ny >= gameMap.Height {
				wallCount++
			} else if !gameMap.Tiles[ny][nx].Walkable {
				wallCount++
			}
		}
	}

	return wallCount
}

// removeSmallAreas removes disconnected floor areas smaller than minSize
func removeSmallAreas(gameMap *game.GameMap, minSize int) error {
	visited := make([][]bool, gameMap.Height)
	for i := range visited {
		visited[i] = make([]bool, gameMap.Width)
	}

	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			if !visited[y][x] && gameMap.Tiles[y][x].Walkable {
				area := floodFillArea(gameMap, x, y, visited)
				if len(area) < minSize {
					// Convert small area to walls
					for _, pos := range area {
						gameMap.Tiles[pos.Y][pos.X].Walkable = false
						gameMap.Tiles[pos.Y][pos.X].Transparent = false
						gameMap.Tiles[pos.Y][pos.X].SpriteX = 1 // Wall sprite
						gameMap.Tiles[pos.Y][pos.X].SpriteY = 0
					}
				}
			}
		}
	}

	return nil
}

// floodFillArea performs flood fill to find connected floor areas
func floodFillArea(gameMap *game.GameMap, startX, startY int, visited [][]bool) []game.Position {
	var area []game.Position
	var stack []game.Position

	stack = append(stack, game.Position{X: startX, Y: startY})

	for len(stack) > 0 {
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if pos.X < 0 || pos.X >= gameMap.Width || pos.Y < 0 || pos.Y >= gameMap.Height {
			continue
		}

		if visited[pos.Y][pos.X] || !gameMap.Tiles[pos.Y][pos.X].Walkable {
			continue
		}

		visited[pos.Y][pos.X] = true
		area = append(area, pos)

		// Add 4-connected neighbors
		stack = append(stack, game.Position{X: pos.X + 1, Y: pos.Y})
		stack = append(stack, game.Position{X: pos.X - 1, Y: pos.Y})
		stack = append(stack, game.Position{X: pos.X, Y: pos.Y + 1})
		stack = append(stack, game.Position{X: pos.X, Y: pos.Y - 1})
	}

	return area
}

// applySmoothingPass applies one smoothing iteration to reduce noise
func applySmoothingPass(gameMap *game.GameMap) error {
	newTiles := make([][]game.MapTile, gameMap.Height)
	for i := range newTiles {
		newTiles[i] = make([]game.MapTile, gameMap.Width)
		copy(newTiles[i], gameMap.Tiles[i])
	}

	for y := 1; y < gameMap.Height-1; y++ {
		for x := 1; x < gameMap.Width-1; x++ {
			wallCount := countNeighborWalls(gameMap, x, y)

			// Smooth isolated walls and floors
			if !gameMap.Tiles[y][x].Walkable && wallCount < 3 {
				newTiles[y][x].Walkable = true
				newTiles[y][x].Transparent = true
				newTiles[y][x].SpriteX = 0 // Floor sprite
				newTiles[y][x].SpriteY = 0
			} else if gameMap.Tiles[y][x].Walkable && wallCount > 5 {
				newTiles[y][x].Walkable = false
				newTiles[y][x].Transparent = false
				newTiles[y][x].SpriteX = 1 // Wall sprite
				newTiles[y][x].SpriteY = 0
			}
		}
	}

	gameMap.Tiles = newTiles
	return nil
}

// enforceEdgeBoundaries ensures map edges are walls with specified buffer
func enforceEdgeBoundaries(gameMap *game.GameMap, buffer int) error {
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			if x < buffer || x >= gameMap.Width-buffer || y < buffer || y >= gameMap.Height-buffer {
				gameMap.Tiles[y][x].Walkable = false
				gameMap.Tiles[y][x].Transparent = false
				gameMap.Tiles[y][x].SpriteX = 1 // Wall sprite
				gameMap.Tiles[y][x].SpriteY = 0
			}
		}
	}

	return nil
}
