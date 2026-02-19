// Package terrain provides procedural terrain generation for the GoldBox RPG engine.
//
// The terrain package implements multiple terrain generation algorithms optimized for
// different biome types and gameplay scenarios. It produces GameMap structures with
// tile-based terrain suitable for dungeon crawling, exploration, and combat.
//
// # Generators
//
// The package provides several terrain generators:
//
//   - CellularAutomataGenerator: Creates organic, cave-like terrain using cellular automata
//     rules. Best suited for natural caves and irregular dungeon layouts.
//
//   - MazeGenerator: Produces traditional maze structures using recursive backtracking.
//     Ideal for labyrinths, puzzle dungeons, and structured corridors.
//
// All generators implement the pcg.TerrainGenerator interface:
//
//	type TerrainGenerator interface {
//		Generator
//		GenerateTerrain(ctx context.Context, width, height int, params TerrainParams) (*game.GameMap, error)
//	}
//
// # Biome System
//
// Terrain generation is influenced by biome definitions that control visual and
// gameplay characteristics:
//
//   - BiomeCave: Natural cave systems with stalactites and underground rivers
//   - BiomeDungeon: Structured dungeons with doors, traps, and secret passages
//   - BiomeForest: Outdoor forested areas with vegetation
//   - BiomeCrypt: Dark tombs with coffins and undead-themed features
//
// Each biome defines:
//   - Tile distribution (wall, floor, water percentages)
//   - Feature placement (decorations, special tiles)
//   - Connectivity requirements (how connected areas should be)
//   - Roughness and density parameters
//
// # Cellular Automata Generation
//
// The CellularAutomataGenerator uses iterative refinement to create natural-looking
// terrain:
//
//	generator := terrain.NewCellularAutomataGenerator()
//	params := pcg.TerrainParams{
//		Seed:       12345,
//		Biome:      pcg.BiomeCave,
//		Density:    0.45,
//		Roughness:  0.7,
//	}
//	gameMap, err := generator.GenerateTerrain(ctx, 50, 50, params)
//
// The algorithm:
//  1. Initializes a grid with random wall placement based on density
//  2. Applies cellular automata rules to smooth terrain
//  3. Detects isolated regions using flood-fill
//  4. Connects regions with corridor carving
//  5. Adds biome-specific features
//
// # Maze Generation
//
// The MazeGenerator creates perfect mazes with guaranteed solutions:
//
//	generator := terrain.NewMazeGenerator()
//	gameMap, err := generator.GenerateTerrain(ctx, 51, 51, params)
//
// Note: Maze dimensions should be odd numbers for proper wall/corridor alignment.
//
// # Feature Placement
//
// After base terrain generation, biome-specific features are added:
//
//   - Cave features: Decorations placed near walls based on roughness
//   - Dungeon doors: Placed at narrow passages between rooms
//   - Torch positions: Wall-mounted lighting with spacing enforcement
//   - Vegetation: Trees and plants placed based on biome density
//
// # Connectivity System
//
// The terrain system ensures all walkable areas are reachable:
//
//   - findWalkableRegions(): Uses flood-fill to identify isolated areas
//   - connectRegions(): Carves L-shaped corridors between disconnected regions
//
// Connectivity levels control how many connections are created:
//   - ConnectivityMinimal: Single path between regions
//   - ConnectivityModerate: Adds redundant connections
//   - ConnectivityHigh: Connects nearest neighbors
//   - ConnectivityComplete: Connects all regions within threshold
//
// # Thread Safety
//
// Terrain generators are safe for concurrent use. Each generation call uses its own
// seeded RNG via the GenerationContext, ensuring deterministic and isolated results.
//
// # Integration with PCG System
//
// Register terrain generators with the central manager:
//
//	manager := pcg.NewManager(12345)
//	manager.RegisterGenerator(pcg.ContentTypeTerrain, terrain.NewCellularAutomataGenerator())
//
// Generate terrain through the unified interface:
//
//	result, err := manager.GenerateContent(ctx, pcg.ContentTypeTerrain, params)
//	gameMap := result.(*game.GameMap)
package terrain
