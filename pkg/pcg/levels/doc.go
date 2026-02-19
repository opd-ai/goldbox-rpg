// Package levels provides procedural dungeon level generation for the GoldBox RPG Engine.
//
// This package implements the room-corridor approach to level generation, creating
// interconnected dungeon layouts with varied room types, themed features, and
// strategic corridor placement for tactical gameplay.
//
// # Overview
//
// The levels package uses a multi-phase generation pipeline:
//
//  1. Space Partitioning: Divide level area into candidate room regions
//  2. Room Generation: Create themed rooms using specialized generators
//  3. Corridor Planning: Connect rooms with varied corridor styles
//  4. Feature Placement: Add special features, encounters, and loot
//  5. Validation: Ensure connectivity and gameplay balance
//
// # Room Types
//
// The package supports 11 distinct room types, each with specialized generators:
//
//   - Combat: Tactical encounter rooms with cover and enemy spawns
//   - Treasure: Loot rooms with valuable items and optional guardians
//   - Puzzle: Interactive challenge rooms requiring problem-solving
//   - Boss: Large encounter rooms for major battles
//   - Entrance: Level starting points with tutorial elements
//   - Exit: Level endpoints with completion rewards
//   - Secret: Hidden rooms requiring discovery
//   - Shop: Merchant areas for equipment trading
//   - Rest: Safe zones for healing and saving
//   - Trap: Hazard rooms requiring careful navigation
//   - Story: Narrative rooms with lore and dialogue
//
// # Creating a Level Generator
//
// For non-deterministic generation (typical gameplay):
//
//	gen := levels.NewRoomCorridorGenerator()
//	level, err := gen.GenerateLevel(ctx, params)
//
// For deterministic generation (testing, replays):
//
//	gen := levels.NewRoomCorridorGeneratorWithSeed(12345)
//	level, err := gen.GenerateLevel(ctx, params)
//
// # Level Parameters
//
// Configure generation via pcg.LevelParams:
//
//	params := pcg.LevelParams{
//	    Seed:             42,
//	    MinRooms:         5,
//	    MaxRooms:         10,
//	    LevelTheme:       pcg.ThemeDungeon,
//	    CorridorStyle:    pcg.CorridorWindy,
//	    GenerationParams: pcg.GenerationParams{Difficulty: 5},
//	}
//	level, err := gen.GenerateLevel(ctx, params)
//
// # Corridor Styles
//
// The CorridorPlanner supports multiple connection styles:
//
//   - Minimal: Direct single-tile connections
//   - Straight: Linear corridors with minimal turns
//   - Windy: Natural-looking meandering paths
//   - Maze: Complex labyrinthine connections
//
// # Level Themes
//
// Generation adapts to the following themes:
//
//   - Classic: Traditional dungeon with stone and torches
//   - Cave: Natural underground environments
//   - Castle: Structured fortress layouts
//   - Crypt: Undead-themed burial chambers
//   - Temple: Religious architecture with altars
//   - Sewer: Underground waterways and grates
//   - Forest: Natural outdoor dungeon variants
//
// # Integration with PCG System
//
// This package implements the pcg.Generator interface for integration with
// the broader procedural content generation system:
//
//	manager := pcg.NewManager()
//	manager.RegisterGenerator(pcg.ContentTypeLevels, levels.NewRoomCorridorGenerator())
//
// # Thread Safety
//
// Level generators maintain internal state (RNG) and should not be shared
// across goroutines without external synchronization. Create separate
// generator instances for concurrent generation.
//
// # Example: Complete Level Generation
//
//	ctx := context.Background()
//	gen := levels.NewRoomCorridorGeneratorWithSeed(time.Now().UnixNano())
//
//	params := pcg.LevelParams{
//	    Seed:       42,
//	    MinRooms:   8,
//	    MaxRooms:   12,
//	    LevelTheme: pcg.ThemeDungeon,
//	    GenerationParams: pcg.GenerationParams{
//	        Difficulty:  5,
//	        PlayerLevel: 3,
//	    },
//	}
//
//	level, err := gen.GenerateLevel(ctx, params)
//	if err != nil {
//	    log.Fatalf("level generation failed: %v", err)
//	}
//
//	// Use level.Tiles, level.Rooms, level.Corridors, etc.
package levels
