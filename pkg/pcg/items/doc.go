// Package items provides procedural item generation for the GoldBox RPG Engine.
//
// This package implements template-based item generation with support for
// enchantments, rarity tiers, and level-scaled statistics. Items are generated
// using configurable templates loaded from YAML configuration files.
//
// # Template-Based Generation
//
// Items are generated using the TemplateBasedGenerator:
//
//	gen := items.NewTemplateBasedGenerator()
//	gen.SetSeed(12345) // For deterministic generation
//
//	params := pcg.GenerationParams{Seed: 12345, Difficulty: 3}
//	item, err := gen.Generate(ctx, params)
//
// # Item Templates
//
// Templates define base item properties and are loaded from YAML files:
//
//	gen.LoadTemplates("data/pcg/items.yaml")
//
// The ItemTemplateRegistry manages template storage and retrieval by item type
// and rarity tier.
//
// # Enchantment System
//
// The EnchantmentSystem adds magical properties to generated items:
//   - Stat bonuses (damage, armor, attributes)
//   - Special effects (fire damage, life steal)
//   - Rarity-appropriate enchantment selection
//   - Enchantment stacking and conflict resolution
//
// # Rarity Tiers
//
// Items are generated across rarity tiers:
//   - Common: Base stats, no enchantments
//   - Uncommon: Minor stat bonuses
//   - Rare: Moderate bonuses, possible enchantments
//   - Epic: Strong bonuses, guaranteed enchantments
//   - Legendary: Maximum stats, unique effects
//
// # Name Generation
//
// GenerateItemName creates contextual names based on template and rarity:
//
//	name := items.GenerateItemName(template, rarity, rng)
//	// Examples: "Iron Sword", "Enchanted Steel Blade", "Legendary Dragon Slayer"
//
// # Deterministic Seeding
//
// All generators support deterministic seeding for reproducible results,
// enabling save/load consistency and multiplayer synchronization.
package items
