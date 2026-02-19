// Package main provides a demonstration application for the multi-level dungeon
// generation system in the GoldBox RPG engine.
//
// The dungeon-demo application showcases the procedural content generation (PCG)
// system's ability to create complete, interconnected dungeon complexes with
// multiple levels, themed rooms, and proper level-to-level connectivity.
//
// # Usage
//
// Run the demo directly:
//
//	go run ./cmd/dungeon-demo
//
// Or build and execute:
//
//	go build -o dungeon-demo ./cmd/dungeon-demo
//	./dungeon-demo
//
// # Generation Features
//
// The demo generates a multi-level dungeon with:
//
//   - Multiple dungeon levels with configurable dimensions (default: 40x30 per level)
//   - Room generation with varied room types (Entry, Combat, Treasure, Boss, etc.)
//   - Level-to-level connections via stairs and ladders
//   - Theme support (Classic, Cave, Temple, Crypt)
//   - Configurable connectivity levels (Low, Moderate, High, Complete)
//   - Progressive difficulty scaling across levels
//
// # Generation Parameters
//
// The dungeon generator accepts parameters controlling:
//
//   - Seed: Fixed seed for reproducible dungeon generation
//   - LevelCount: Number of dungeon levels to generate
//   - LevelWidth/Height: Dimensions of each level
//   - RoomsPerLevel: Target number of rooms per level
//   - Theme: Visual and structural theme for the dungeon
//   - Connectivity: How interconnected rooms should be
//   - Density: How much of the level should be filled
//   - Difficulty: Base difficulty and scaling progression
//
// # Output
//
// The demo outputs:
//
//   - Generation timing information
//   - Dungeon name and unique identifier
//   - Per-level statistics including room counts and types
//   - Connection details between levels
//   - Total room counts and dungeon metadata
//
// # Integration Example
//
// The dungeon generator can be used in game initialization:
//
//	generator := pcg.NewDungeonGenerator(logger)
//	params := pcg.GenerationParams{
//	    Seed:       time.Now().UnixNano(),
//	    Difficulty: playerLevel,
//	    Constraints: map[string]interface{}{
//	        "dungeon_params": pcg.DungeonParams{
//	            LevelCount:    5,
//	            RoomsPerLevel: 8,
//	            Theme:         pcg.ThemeCrypt,
//	        },
//	    },
//	}
//	result, err := generator.Generate(ctx, params)
//	dungeon := result.(*pcg.DungeonComplex)
package main
