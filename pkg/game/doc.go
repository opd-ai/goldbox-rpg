// Package game implements the core RPG mechanics engine for the GoldBox RPG system.
//
// This package provides comprehensive game state management including character systems,
// combat mechanics, effect management, spell casting, world state, and entity interactions
// for turn-based dungeon-crawling gameplay.
//
// # Character System
//
// Characters have six core attributes (Strength, Dexterity, Constitution, Intelligence,
// Wisdom, Charisma) along with combat statistics (HP, AC, THAC0), equipment slots,
// inventory, and active effects. The Player type extends Character with progression
// mechanics including level, experience points, quest logs, and known spells.
//
//	char := game.NewCharacter("Hero", game.Fighter)
//	player := game.NewPlayer(char)
//
// # Combat System
//
// Combat uses THAC0-based hit calculations with armor class defense, action points
// for turn management, and initiative ordering. The TurnManager coordinates
// turn-based combat flow.
//
// # Effect System
//
// Effects represent status conditions with duration, magnitude, and tick-based updates.
// Supports damage-over-time, healing, stat buffs/debuffs, and custom behaviors.
// The EffectManager handles active effects on entities with proper stacking rules.
//
//	effect := game.NewEffect(game.EffectPoison, 5, 3) // 5 damage, 3 turns
//	char.AddEffect(effect)
//
// # Spell System
//
// Spells are organized by school (Abjuration, Evocation, Conjuration, etc.) with
// configurable range, damage/healing dice, area effects, and components.
// The SpellManager handles spell library and casting mechanics.
//
// # World Management
//
// The World type serves as the primary game state container, managing multiple
// dungeon levels, entities, and time tracking. It uses spatial indexing via
// SpatialIndex for efficient entity location queries and area searches.
//
//	world := game.NewWorld()
//	world.AddEntity(player)
//	nearby := world.GetEntitiesInRange(position, 10)
//
// # Thread Safety
//
// All core types support concurrent access via sync.RWMutex protection.
// Character state modifications require proper locking through the provided
// getter/setter methods.
//
// # Persistence
//
// Types support YAML serialization for save/load functionality using struct tags.
// The internal mutex is excluded from serialization via yaml:"-" tags.
package game
