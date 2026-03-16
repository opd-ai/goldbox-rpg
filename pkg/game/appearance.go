package game

// BodyType represents a character's body type (cosmetic only).
type BodyType int

const (
	BodyDefault  BodyType = iota // zero-value; treated as "average" at render time
	BodySlim                     // slim build
	BodyAverage                  // average build
	BodyMuscular                 // muscular build
	BodyStocky                   // stocky build
	BodyLarge                    // large build
)

// SkinTonePalette maps numeric skin tone scale (1-10) to named entries for UI display.
// Decoupled from fantasy race; purely cosmetic.
var SkinTonePalette = map[int]string{
	1:  "porcelain",
	2:  "ivory",
	3:  "beige",
	4:  "sand",
	5:  "tan",
	6:  "bronze",
	7:  "umber",
	8:  "chestnut",
	9:  "espresso",
	10: "obsidian",
}

// Appearance holds cosmetic and biographical character properties.
// None of these fields affect attributes, combat, or class eligibility.
// All fields are optional; zero values are valid defaults.
type Appearance struct {
	SkinTone            int      `yaml:"skin_tone"            json:"skin_tone,omitempty"`
	HairStyle           string   `yaml:"hair_style"           json:"hair_style,omitempty"`
	HairColor           string   `yaml:"hair_color"           json:"hair_color,omitempty"`
	BodyType            BodyType `yaml:"body_type"            json:"body_type,omitempty"`
	GenderExpression    string   `yaml:"gender_expression"    json:"gender_expression,omitempty"`
	Pronouns            string   `yaml:"pronouns"             json:"pronouns,omitempty"`
	RomanticOrientation string   `yaml:"romantic_orientation" json:"romantic_orientation,omitempty"`
}

// DefaultAppearance returns an Appearance with sensible mid-range defaults.
func DefaultAppearance() Appearance {
	return Appearance{
		SkinTone: 5,
		BodyType: BodyAverage,
	}
}

// SkinToneName returns the named palette entry for the tone value,
// or "unknown" if out of range.
func SkinToneName(tone int) string {
	if name, ok := SkinTonePalette[tone]; ok {
		return name
	}
	return "unknown"
}

// SkinToneGroup returns "light" (1-3), "medium" (4-7), or "dark" (8-10).
// Used for portrait asset lookup.
func SkinToneGroup(tone int) string {
	switch {
	case tone <= 0:
		return "medium"
	case tone <= 3:
		return "light"
	case tone <= 7:
		return "medium"
	default:
		return "dark"
	}
}

// PortraitTag returns the gender-expression tag for asset lookup.
// Maps common expressions to "a" / "b" / "nb" portrait sets;
// unknown values default to "nb".
func (a Appearance) PortraitTag() string {
	switch a.GenderExpression {
	case "masculine", "male":
		return "a"
	case "feminine", "female":
		return "b"
	default:
		return "nb"
	}
}
