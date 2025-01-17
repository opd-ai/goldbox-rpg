package game

// CharacterClass represents available character classes
// CharacterClass represents the character's role or profession in the game.
// It is implemented as an enumerated type using integers for efficient storage
// and comparison operations.
//
// The specific class values and their gameplay implications should be defined
// as constants using this type. Each class may have different abilities,
// starting stats, and progression paths.
//
// Related types:
// - Character struct (which likely contains this as a field)
// - Any class-specific ability or skill types
type CharacterClass int

const (
	ClassFighter CharacterClass = iota
	ClassMage
	ClassCleric
	ClassThief
	ClassRanger
	ClassPaladin
)

// String returns the string representation of a CharacterClass.
// It converts the CharacterClass enum value to its corresponding human-readable name.
//
// Returns:
//
//	string: The name of the character class as a string ("Fighter", "Mage", etc.)
//
// Notable Cases:
//   - Assumes valid enum values within array bounds
//   - Will panic if given an invalid CharacterClass value
//
// Related Types:
//   - CharacterClass type (enum)
func (cc CharacterClass) String() string {
	return [...]string{
		"Fighter",
		"Mage",
		"Cleric",
		"Thief",
		"Ranger",
		"Paladin",
	}[cc]
}

// ClassConfig represents the configuration for a character class
// Contains all metadata and attributes for a specific class
type ClassConfig struct {
	Type         CharacterClass `yaml:"class_type"`        // The class enumeration value
	Name         string         `yaml:"class_name"`        // Display name of the class
	Description  string         `yaml:"class_description"` // Class description and background
	HitDice      string         `yaml:"class_hit_dice"`    // Hit points per level (e.g., "1d10")
	BaseSkills   []string       `yaml:"class_base_skills"` // Default skills for the class
	Abilities    []string       `yaml:"class_abilities"`   // Special class abilities
	Requirements struct {
		MinStr int `yaml:"min_strength"`     // Minimum strength requirement
		MinDex int `yaml:"min_dexterity"`    // Minimum dexterity requirement
		MinCon int `yaml:"min_constitution"` // Minimum constitution requirement
		MinInt int `yaml:"min_intelligence"` // Minimum intelligence requirement
		MinWis int `yaml:"min_wisdom"`       // Minimum wisdom requirement
		MinCha int `yaml:"min_charisma"`     // Minimum charisma requirement
	} `yaml:"class_requirements"` // Minimum stat requirements
}

// ClassProficiencies represents weapon and armor proficiencies for a class
// ClassProficiencies defines what equipment and items a character class can use.
// It specifies allowed weapons, armor types and any special restrictions.
//
// Fields:
//   - Class: The character class these proficiencies apply to
//   - WeaponTypes: List of weapon categories this class can use (e.g. "sword", "bow")
//   - ArmorTypes: List of armor categories this class can wear (e.g. "light", "heavy")
//   - ShieldProficient: Whether the class is trained in shield usage
//   - Restrictions: Any special limitations on equipment usage
//
// Related types:
//   - CharacterClass: The class enum these proficiencies are linked to
//
// Example:
//
//	Fighter proficiencies would allow all weapons and armor types with shield use
//	Mage proficiencies would be limited to staves/wands and light armor with no shields
type ClassProficiencies struct {
	Class            CharacterClass `yaml:"class_type"`             // Associated character class
	WeaponTypes      []string       `yaml:"allowed_weapons"`        // Allowed weapon types
	ArmorTypes       []string       `yaml:"allowed_armor"`          // Allowed armor types
	ShieldProficient bool           `yaml:"can_use_shields"`        // Whether class can use shields
	Restrictions     []string       `yaml:"equipment_restrictions"` // Special equipment restrictions
}
