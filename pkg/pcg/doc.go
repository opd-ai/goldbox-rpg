// Package pcg provides Procedural Content Generation for the GoldBox RPG Engine.
//
// This package dynamically generates game content including terrain, items,
// dungeon levels, quests, NPCs, and factions with deterministic seeding,
// validation, and world integration.
//
// # PCGManager
//
// PCGManager is the central coordinator for all content generation:
//
//	manager := pcg.NewPCGManager(world)
//
//	// Generate terrain
//	terrain, err := manager.GenerateTerrain(ctx, biome, width, height)
//
//	// Generate items
//	item, err := manager.GenerateItem(ctx, rarity, itemType)
//
//	// Generate dungeon level
//	level, err := manager.GenerateLevel(ctx, difficulty, template)
//
// # Content Types
//
// The system supports generation of:
//   - Terrain: Biome-aware landscapes with proper connectivity
//   - Items: Equipment with stats, enchantments, and rarity tiers
//   - Levels: Complete dungeon floors with rooms and corridors
//   - Quests: Multi-objective quest chains with narrative
//   - NPCs: Characters with personalities and behaviors
//   - Factions: Groups with relationships and reputations
//
// # Generator Registry
//
// Register custom generators for extensible content creation:
//
//	registry := pcg.NewRegistry()
//	registry.Register("custom-terrain", myGenerator)
//
//	factory := pcg.NewFactory(registry)
//	content, err := factory.Create(ctx, "custom-terrain", params)
//
// # Deterministic Seeding
//
// SeedManager provides reproducible generation:
//
//	seedMgr := pcg.NewSeedManager(baseSeed)
//	seed := seedMgr.GetSeed("terrain", x, y)
//
// The same seed produces identical content, enabling:
//   - Save/load consistency
//   - Multiplayer synchronization
//   - Bug reproduction
//
// # Validation
//
// Content is validated before world integration:
//
//	validator := pcg.NewValidator()
//	result := validator.Validate(content)
//	if !result.Valid {
//	    for _, err := range result.Errors {
//	        log.Error(err)
//	    }
//	}
//
// # World Integration
//
// Generated content integrates safely with game world:
//
//	err := manager.IntegrateContent(ctx, content)
//
// Integration handles:
//   - Spatial indexing updates
//   - Entity registration
//   - Event emission
//
// # Metrics
//
// Performance and quality metrics for monitoring:
//
//	metrics := manager.GetMetrics()
//	// Generation times, cache hits, validation failures
//
// # Subpackages
//
// Specialized generators in subpackages:
//   - pcg/terrain: Biome-aware terrain with noise algorithms
//   - pcg/items: Equipment generation with enchantments
//   - pcg/levels: Room and corridor dungeon layouts
//   - pcg/quests: Quest objective and narrative generation
//   - pcg/utils: Pathfinding, noise, and utility functions
package pcg
