// Package quests provides procedural quest generation for the GoldBox RPG engine.
//
// The quests package implements a sophisticated quest generation system that creates
// dynamic, context-aware quests with objectives, narratives, and rewards. It integrates
// with the core PCG system to ensure consistent procedural generation across the game.
//
// # Architecture
//
// The package consists of three main components:
//
//   - ObjectiveBasedGenerator: Creates complete quests using objective templates
//   - ObjectiveGenerator: Generates specific quest objectives (kill, collect, explore, etc.)
//   - NarrativeEngine: Produces quest stories, dialogue, and lore
//
// # Quest Generation
//
// Quests are generated using the ObjectiveBasedGenerator, which implements the
// pcg.Generator interface:
//
//	generator := quests.NewObjectiveBasedGenerator()
//	params := pcg.GenerationParams{
//		Seed:       12345,
//		Difficulty: 5,
//		Constraints: map[string]interface{}{
//			"quest_type":     pcg.QuestTypeMainStory,
//			"min_objectives": 2,
//			"max_objectives": 4,
//		},
//	}
//	result, err := generator.Generate(ctx, params)
//
// # Objective Types
//
// The ObjectiveGenerator supports multiple objective types for different gameplay styles:
//
//   - Kill objectives: Defeat enemies of specific types
//   - Collect objectives: Gather items from locations
//   - Explore objectives: Discover and visit areas
//   - Escort objectives: Protect NPCs during travel
//   - Deliver objectives: Transport items between locations
//
// Each objective type scales with difficulty and adapts to the game world context.
//
// # Narrative Generation
//
// The NarrativeEngine creates contextual stories using template-based generation:
//
//	engine := quests.NewNarrativeEngine()
//	narrative := engine.GenerateNarrative(quest, pcg.QuestTypeMainStory, rng)
//
// Narratives include:
//   - Quest titles and descriptions
//   - Quest giver characterization
//   - Start and end dialogue
//   - Contextual lore elements
//
// # Templates
//
// Quest generation uses configurable templates for objectives and stories:
//
//	template := &quests.ObjectiveTemplate{
//		Type:        "kill",
//		Description: "Defeat the %s threatening the village",
//		Targets:     []string{"goblin", "orc", "troll"},
//		Quantities:  [2]int{5, 15},
//		Rewards:     []string{"gold", "experience"},
//	}
//
// # Thread Safety
//
// Quest generators are safe for concurrent use. Each generation call creates its own
// RNG state from the provided seed, ensuring deterministic results without shared
// mutable state.
//
// # Integration with PCG System
//
// The quests package integrates with the central PCG manager:
//
//	manager := pcg.NewManager(12345)
//	manager.RegisterGenerator(pcg.ContentTypeQuests, quests.NewObjectiveBasedGenerator())
//
// This allows quests to be generated alongside other content types while maintaining
// consistent seed management across the game.
package quests
