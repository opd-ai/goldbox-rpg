package game

// Spell represents a magical ability that can be cast in the game.
// It contains all the necessary information about a spell's properties and effects.
//
// Fields:
//   - ID: Unique string identifier for the spell
//   - Name: Display name shown to players
//   - Level: Required caster level (must be >= 0)
//   - School: Magic school classification (e.g. Abjuration, Evocation)
//   - Range: Distance in game units the spell can reach (must be >= 0)
//   - Duration: Number of game turns the spell effects last (must be >= 0)
//   - Components: Required components needed to cast the spell
//   - Description: Detailed text describing the spell's effects and usage
//
// Related types:
//   - SpellSchool: Enum defining valid magic schools
//   - SpellComponent: Struct defining spell component requirements
type Spell struct {
	ID          string           `yaml:"spell_id"`          // Unique identifier for the spell
	Name        string           `yaml:"spell_name"`        // Display name of the spell
	Level       int              `yaml:"spell_level"`       // Required caster level for the spell
	School      SpellSchool      `yaml:"spell_school"`      // Magic school classification
	Range       int              `yaml:"spell_range"`       // Range in game units
	Duration    int              `yaml:"spell_duration"`    // Duration in game turns
	Components  []SpellComponent `yaml:"spell_components"`  // Required components for casting
	Description string           `yaml:"spell_description"` // Full spell description and effects
}

// SpellSchool represents the different schools of magic available in the game
// SpellSchool represents the school/category of magic that a spell belongs to.
// It is implemented as an integer type for efficient storage and comparison.
// The specific values are defined as constants representing different magical disciplines
// like Abjuration, Conjuration, Evocation etc.
//
// Related types:
// - Spell struct - Contains SpellSchool as one of its properties
// - SpellEffect interface - Implemented by specific spell effects
type SpellSchool int

// SchoolAbjuration represents the school of Abjuration magic in the game world.
// Abjuration spells are protective in nature, creating barriers, negating harmful
// effects, or banishing creatures to other planes of existence.
//
// This is one of the eight classical schools of magic defined in the game system.
//
// Related constants:
// - SchoolConjuration
// - SchoolDivination
// - SchoolEnchantment
// - SchoolEvocation
// - SchoolIllusion
// - SchoolNecromancy
// - SchoolTransmutation
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

// SpellComponent represents a component of a spell in the game.
// It is implemented as an integer type that can be used to classify
// different aspects or parts of a spell, such as verbal, somatic,
// or material components.
//
// Related types:
//   - Spell (not shown in provided code)
type SpellComponent int

// ComponentVerbal represents the verbal component required for casting spells.
// It indicates that the spell requires specific words or phrases to be spoken
// to be successfully cast. This is one of the fundamental spell components
// alongside Somatic and Material components.
//
// Related constants:
// - ComponentSomatic
// - ComponentMaterial
const (
	ComponentVerbal SpellComponent = iota
	ComponentSomatic
	ComponentMaterial
)
