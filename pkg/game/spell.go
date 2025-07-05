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
//   - DamageType: Type of damage dealt (fire, cold, etc.)
//   - DamageDice: Dice expression for damage (e.g., "3d6+2")
//   - HealingDice: Dice expression for healing (e.g., "2d4+1")
//   - AreaEffect: Whether the spell affects an area
//   - SaveType: Type of saving throw required
//   - EffectKeywords: Tags describing spell effects
//
// Related types:
//   - SpellSchool: Enum defining valid magic schools
//   - SpellComponent: Struct defining spell component requirements
type Spell struct {
	ID             string           `yaml:"spell_id"`          // Unique identifier for the spell
	Name           string           `yaml:"spell_name"`        // Display name of the spell
	Level          int              `yaml:"spell_level"`       // Required caster level for the spell
	School         SpellSchool      `yaml:"spell_school"`      // Magic school classification
	Range          int              `yaml:"spell_range"`       // Range in game units
	Duration       int              `yaml:"spell_duration"`    // Duration in game turns
	Components     []SpellComponent `yaml:"spell_components"`  // Required components for casting
	Description    string           `yaml:"spell_description"` // Full spell description and effects
	DamageType     string           `yaml:"damage_type"`       // Type of damage (fire, cold, etc.)
	DamageDice     string           `yaml:"damage_dice"`       // Damage dice expression
	HealingDice    string           `yaml:"healing_dice"`      // Healing dice expression
	AreaEffect     bool             `yaml:"area_effect"`       // Whether spell affects an area
	SaveType       string           `yaml:"save_type"`         // Required saving throw type
	EffectKeywords []string         `yaml:"effect_keywords"`   // Tags describing spell effects
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

// String returns the string representation of a SpellSchool
func (s SpellSchool) String() string {
	switch s {
	case SchoolAbjuration:
		return "Abjuration"
	case SchoolConjuration:
		return "Conjuration"
	case SchoolDivination:
		return "Divination"
	case SchoolEnchantment:
		return "Enchantment"
	case SchoolEvocation:
		return "Evocation"
	case SchoolIllusion:
		return "Illusion"
	case SchoolNecromancy:
		return "Necromancy"
	case SchoolTransmutation:
		return "Transmutation"
	default:
		return "Unknown"
	}
}

// ParseSpellSchool converts a string to a SpellSchool enum
func ParseSpellSchool(s string) SpellSchool {
	switch s {
	case "Abjuration", "abjuration":
		return SchoolAbjuration
	case "Conjuration", "conjuration":
		return SchoolConjuration
	case "Divination", "divination":
		return SchoolDivination
	case "Enchantment", "enchantment":
		return SchoolEnchantment
	case "Evocation", "evocation":
		return SchoolEvocation
	case "Illusion", "illusion":
		return SchoolIllusion
	case "Necromancy", "necromancy":
		return SchoolNecromancy
	case "Transmutation", "transmutation":
		return SchoolTransmutation
	default:
		return SchoolEvocation // Default to Evocation
	}
}

// String returns the string representation of a SpellComponent
func (c SpellComponent) String() string {
	switch c {
	case ComponentVerbal:
		return "Verbal"
	case ComponentSomatic:
		return "Somatic"
	case ComponentMaterial:
		return "Material"
	default:
		return "Unknown"
	}
}

// ParseSpellComponent converts a string to a SpellComponent enum
func ParseSpellComponent(s string) SpellComponent {
	switch s {
	case "Verbal", "verbal", "V":
		return ComponentVerbal
	case "Somatic", "somatic", "S":
		return ComponentSomatic
	case "Material", "material", "M":
		return ComponentMaterial
	default:
		return ComponentVerbal // Default to Verbal
	}
}

// SpellCollection represents a collection of spells loaded from YAML
type SpellCollection struct {
	Spells []Spell `yaml:"spells"`
}
