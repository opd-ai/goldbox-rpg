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
type SpellSchool int

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

// SpellComponent represents the physical or verbal components required to cast a spell
type SpellComponent int

const (
	ComponentVerbal SpellComponent = iota
	ComponentSomatic
	ComponentMaterial
)
