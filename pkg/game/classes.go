package game

// CharacterClass represents available character classes
type CharacterClass int

const (
	ClassFighter CharacterClass = iota
	ClassMage
	ClassCleric
	ClassThief
	ClassRanger
	ClassPaladin
)

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
type ClassProficiencies struct {
	Class            CharacterClass `yaml:"class_type"`             // Associated character class
	WeaponTypes      []string       `yaml:"allowed_weapons"`        // Allowed weapon types
	ArmorTypes       []string       `yaml:"allowed_armor"`          // Allowed armor types
	ShieldProficient bool           `yaml:"can_use_shields"`        // Whether class can use shields
	Restrictions     []string       `yaml:"equipment_restrictions"` // Special equipment restrictions
}
