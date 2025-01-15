package game

// Spell represents a magical ability that can be cast by characters.
// Contains all the core attributes and metadata needed to define a spell effect.
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
