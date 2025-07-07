package game

// Direction constants represent the four cardinal directions.
// These values are used throughout the game for movement, facing, and orientation.
// The values increment clockwise starting from North (0).
// Moved from: types.go
const (
	DirectionNorth Direction = iota // North direction (0 degrees)
	DirectionEast                   // East direction (90 degrees)
	DirectionSouth                  // South direction (180 degrees)
	DirectionWest                   // West direction (270 degrees)
)

// Legacy constants for backward compatibility
// Moved from: types.go
const (
	North = DirectionNorth
	East  = DirectionEast
	South = DirectionSouth
	West  = DirectionWest
)

// TileType constants represent different types of tiles in the game world.
// Each constant is assigned a unique integer value through iota.
// Moved from: tile.go
const (
	TileFloor  TileType = iota // Basic floor tile that can be walked on
	TileWall                   // Solid wall that blocks movement and sight
	TileDoor                   // Door that can be opened/closed
	TileWater                  // Water tile that may affect movement
	TileLava                   // Dangerous lava tile that causes damage
	TilePit                    // Pit that entities may fall into
	TileStairs                 // Stairs for level transitions
)

// Effect constants define types, damage types, and related game mechanics.
// Related effects: EffectPoison, EffectBurning, EffectBleeding
// Related damage types: DamagePhysical, DamageFire, DamagePoison
// Moved from: effects.go
const (
	// Effect Types
	EffectDamageOverTime EffectType = "damage_over_time"
	EffectHealOverTime   EffectType = "heal_over_time"
	EffectPoison         EffectType = "poison"
	EffectBurning        EffectType = "burning"
	EffectBleeding       EffectType = "bleeding"
	EffectStun           EffectType = "stun"
	EffectRoot           EffectType = "root"
	EffectStatBoost      EffectType = "stat_boost"
	EffectStatPenalty    EffectType = "stat_penalty"

	// Damage Types
	DamagePhysical  DamageType = "physical"
	DamageFire      DamageType = "fire"
	DamagePoison    DamageType = "poison"
	DamageFrost     DamageType = "frost"
	DamageLightning DamageType = "lightning"

	// Dispel Types
	DispelMagic   DispelType = "magic"
	DispelCurse   DispelType = "curse"
	DispelPoison  DispelType = "poison"
	DispelDisease DispelType = "disease"
	DispelAll     DispelType = "all"

	// Immunity Types
	ImmunityNone ImmunityType = iota
	ImmunityPartial
	ImmunityComplete
	ImmunityReflect

	// Dispel Priorities
	DispelPriorityLowest  DispelPriority = 0
	DispelPriorityLow     DispelPriority = 25
	DispelPriorityNormal  DispelPriority = 50
	DispelPriorityHigh    DispelPriority = 75
	DispelPriorityHighest DispelPriority = 100
)

// ModOpType constants define modifier operations for effects.
// Moved from: effects.go
const (
	ModAdd      ModOpType = "add"      // Adds the modifier value to the base stat
	ModMultiply ModOpType = "multiply" // Multiplies the base stat by the modifier value
	ModSet      ModOpType = "set"      // Sets the stat directly to the modifier value
)

// EquipmentSlot constants represent the different slots where equipment/items can be equipped on a character.
// This type is used as an enum to identify valid equipment positions (e.g. weapon slot, armor slot, etc).
// Moved from: equipment.go
const (
	SlotHead EquipmentSlot = iota
	SlotNeck
	SlotChest
	SlotHands
	SlotRings
	SlotLegs
	SlotFeet
	SlotWeaponMain
	SlotWeaponOff
)

// CharacterClass constants represent different character classes in the game.
// These classes define the role and abilities of player characters.
// Moved from: classes.go
const (
	ClassFighter CharacterClass = iota
	ClassMage
	ClassCleric
	ClassThief
	ClassRanger
	ClassPaladin
)

// SpellSchool constants represent different schools of magic in the game.
// These schools categorize spells by their magical disciplines and effects.
// Moved from: spell.go
const (
	SchoolAbjuration SpellSchool = iota
	SchoolConjuration
	SchoolDivination
	SchoolEnchantment
	SchoolEvocation
	SchoolIllusion
	SchoolNecromancy
	SchoolTransmutation
)

// SpellComponent constants represent components required for casting spells.
// ComponentVerbal represents the verbal component required for casting spells.
// It indicates that the spell requires specific words or phrases to be spoken
// to be successfully cast. This is one of the fundamental spell components
// alongside Somatic and Material components.
// Moved from: spell.go
const (
	ComponentVerbal SpellComponent = iota
	ComponentSomatic
	ComponentMaterial
)

// QuestStatus constants represent the state of a quest in the game.
// QuestNotStarted indicates that a quest has not yet been started by the player.
// This is the initial state of any quest when first created or discovered.
// Related: QuestActive, QuestCompleted, QuestFailed.
// Moved from: quest.go
const (
	QuestNotStarted QuestStatus = iota
	QuestActive
	QuestCompleted
	QuestFailed
)

// EventType constants represent different types of events in the game system.
// Related events and experience gain, quest requirements.
// Moved from: events.go
const (
	EventLevelUp EventType = iota
	EventDamage
	EventDeath
	EventItemPickup
	EventItemDrop
	EventMovement
	EventSpellCast
	EventQuestUpdate
)

// ItemType constants represent different categories of items in the game.
// ItemTypeWeapon represents a weapon item type constant used for categorizing items
// in the game inventory and equipment system. This type is used when creating or
// identifying weapon items.
// Moved from: item.go
const (
	ItemTypeWeapon = "weapon"
	ItemTypeArmor  = "armor"
)

// DefaultWorld constants define the dimensions of the default test world.
// Moved from: default_world.go
const (
	DefaultWorldWidth  = 10
	DefaultWorldHeight = 10
)

// Action Point constants define the cost of different actions in combat.
// Simple system: 2 points per turn, 1 for move, 1 for attack/spell.
const (
	ActionPointsPerTurn = 2 // Total action points available per turn
	ActionCostMove      = 1 // Cost to move one tile
	ActionCostAttack    = 1 // Cost to perform a melee/ranged attack
	ActionCostSpell     = 1 // Cost to cast a spell
)
