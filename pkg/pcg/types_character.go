package pcg

// CharacterType represents different categories of NPCs
type CharacterType string

const (
	CharacterTypeGeneric  CharacterType = "generic"  // General purpose NPC
	CharacterTypeMerchant CharacterType = "merchant" // Shop keeper or trader
	CharacterTypeGuard    CharacterType = "guard"    // Security or military
	CharacterTypeNoble    CharacterType = "noble"    // Aristocracy or leadership
	CharacterTypePeasant  CharacterType = "peasant"  // Common folk or laborers
	CharacterTypeCrafter  CharacterType = "crafter"  // Artisan or specialist
	CharacterTypeCleric   CharacterType = "cleric"   // Religious figure
	CharacterTypeMage     CharacterType = "mage"     // Magic user or scholar
	CharacterTypeRogue    CharacterType = "rogue"    // Thief or scoundrel
	CharacterTypeBard     CharacterType = "bard"     // Entertainer or storyteller
)

// BackgroundType represents character background categories
type BackgroundType string

const (
	BackgroundUrban      BackgroundType = "urban"      // City dweller
	BackgroundRural      BackgroundType = "rural"      // Countryside origin
	BackgroundNomadic    BackgroundType = "nomadic"    // Travel-oriented background
	BackgroundNoble      BackgroundType = "noble"      // Aristocratic upbringing
	BackgroundCriminal   BackgroundType = "criminal"   // Unlawful background
	BackgroundMilitary   BackgroundType = "military"   // Armed forces background
	BackgroundReligious  BackgroundType = "religious"  // Religious/monastic background
	BackgroundScholar    BackgroundType = "scholar"    // Academic background
	BackgroundWilderness BackgroundType = "wilderness" // Outdoor/survival background
)

// SocialClass represents character's social standing
type SocialClass string

const (
	SocialClassSlave    SocialClass = "slave"    // Lowest social position
	SocialClassSerf     SocialClass = "serf"     // Bound peasant
	SocialClassPeasant  SocialClass = "peasant"  // Free commoner
	SocialClassCrafter  SocialClass = "crafter"  // Skilled artisan
	SocialClassMerchant SocialClass = "merchant" // Trade class
	SocialClassGentry   SocialClass = "gentry"   // Minor nobility
	SocialClassNoble    SocialClass = "noble"    // Aristocracy
	SocialClassRoyalty  SocialClass = "royalty"  // Ruling class
)

// AgeRange represents character age categories
type AgeRange string

const (
	AgeRangeChild      AgeRange = "child"       // Young character (5-12)
	AgeRangeAdolescent AgeRange = "adolescent"  // Teenage character (13-17)
	AgeRangeYoungAdult AgeRange = "young_adult" // Young adult (18-25)
	AgeRangeAdult      AgeRange = "adult"       // Mature adult (26-40)
	AgeRangeMiddleAged AgeRange = "middle_aged" // Middle-aged (41-60)
	AgeRangeElderly    AgeRange = "elderly"     // Elderly character (61+)
	AgeRangeAncient    AgeRange = "ancient"     // Very old character (special cases)
)

// NPCGroupType represents different types of NPC groups
type NPCGroupType string

const (
	NPCGroupFamily    NPCGroupType = "family"    // Related family members
	NPCGroupGuards    NPCGroupType = "guards"    // Security patrol or unit
	NPCGroupMerchants NPCGroupType = "merchants" // Trading group or caravan
	NPCGroupCultists  NPCGroupType = "cultists"  // Religious or cult group
	NPCGroupBandits   NPCGroupType = "bandits"   // Criminal organization
	NPCGroupScholars  NPCGroupType = "scholars"  // Academic or research group
	NPCGroupCrafters  NPCGroupType = "crafters"  // Guild or workshop group
)

// PersonalityTrait represents individual personality characteristics
type PersonalityTrait struct {
	Name        string  `json:"name"`        // Trait name (e.g., "brave", "greedy")
	Intensity   float64 `json:"intensity"`   // Trait strength (0.0-1.0)
	Description string  `json:"description"` // Descriptive text
}

// Motivation represents character goals and drives
type Motivation struct {
	Type        string  `json:"type"`        // Motivation category (power, wealth, love, etc.)
	Target      string  `json:"target"`      // What the motivation is directed toward
	Intensity   float64 `json:"intensity"`   // How strongly motivated (0.0-1.0)
	Description string  `json:"description"` // Detailed description
}

// PersonalityProfile represents a complete character personality system
type PersonalityProfile struct {
	Traits      []PersonalityTrait `json:"traits"`      // Individual personality traits
	Motivations []Motivation       `json:"motivations"` // Character goals and drives
	Alignment   string             `json:"alignment"`   // Moral alignment
	Temperament string             `json:"temperament"` // General disposition
	Values      []string           `json:"values"`      // What the character values most
	Fears       []string           `json:"fears"`       // Character's primary fears
	Speech      SpeechPattern      `json:"speech"`      // How the character speaks
}

// SpeechPattern represents how a character communicates
type SpeechPattern struct {
	Formality   string   `json:"formality"`   // Level of formality (formal, casual, crude)
	Vocabulary  string   `json:"vocabulary"`  // Complexity level (simple, moderate, complex)
	Accent      string   `json:"accent"`      // Regional or cultural accent
	Mannerisms  []string `json:"mannerisms"`  // Speech habits or quirks
	Catchphrase string   `json:"catchphrase"` // Signature phrase (optional)
}
