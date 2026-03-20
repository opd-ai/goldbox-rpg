// Package wasmui provides the Ebitengine/WASM-based game UI client types.
// This file contains core game state types for player, combat, and world.
package wasmui

// Position represents a 2D coordinate.
// Uses uppercase field names to match server's default JSON encoding.
type Position struct {
	X      int `json:"X"`
	Y      int `json:"Y"`
	Level  int `json:"Level,omitempty"`
	Facing int `json:"Facing,omitempty"`
}

// PlayerAttributes represents the six core D&D-style attributes.
// Field names match the server's public data format.
type PlayerAttributes struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// PlayerState represents the current state of a player character.
// Field names match the server's session PublicData format.
type PlayerState struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Position   Position         `json:"position"`
	HP         int              `json:"hp"`
	MaxHP      int              `json:"max_hp"`
	AP         int              `json:"ap"`
	MaxAP      int              `json:"max_ap"`
	Level      int              `json:"level"`
	Experience int              `json:"experience"`
	Class      string           `json:"class"`
	Attributes PlayerAttributes `json:"attributes"`
	Appearance *Appearance      `json:"appearance,omitempty"`
	Effects    []EffectData     `json:"effects,omitempty"`
	Equipment  []EquippedItem   `json:"equipment,omitempty"`
	Immunities []string         `json:"immunities,omitempty"`
}

// Appearance holds cosmetic and biographical character properties.
// Mirrors game.Appearance for the WASM UI layer.
type Appearance struct {
	SkinTone            int    `json:"skin_tone,omitempty"`
	HairStyle           string `json:"hair_style,omitempty"`
	HairColor           string `json:"hair_color,omitempty"`
	BodyType            int    `json:"body_type,omitempty"`
	GenderExpression    string `json:"gender_expression,omitempty"`
	Pronouns            string `json:"pronouns,omitempty"`
	RomanticOrientation string `json:"romantic_orientation,omitempty"`
}

// EquippedItem represents an item in an equipment slot.
type EquippedItem struct {
	Slot   string `json:"slot"`
	ItemID string `json:"item_id"`
	Name   string `json:"name"`
}

// EffectData represents an active effect on a character.
type EffectData struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Duration      int    `json:"duration"`
	Remaining     int    `json:"remaining"`
	Magnitude     int    `json:"magnitude"`
	Source        string `json:"source"`
	EffectKeyword string `json:"effect_keyword,omitempty"`
}

// CombatState represents the current combat state.
type CombatState struct {
	Active       bool              `json:"active"`
	CurrentTurn  string            `json:"currentTurn"`
	Initiative   []InitiativeEntry `json:"initiative"`
	Round        int               `json:"round"`
	InCombat     bool              `json:"inCombat"`
	IsPlayerTurn bool              `json:"isPlayerTurn"`
}

// InitiativeEntry represents an entry in the combat initiative order.
type InitiativeEntry struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Initiative  int          `json:"initiative"`
	IsPlayer    bool         `json:"isPlayer"`
	HP          int          `json:"hp,omitempty"`
	MaxHP       int          `json:"max_hp,omitempty"`
	MoraleState string       `json:"morale_state,omitempty"`
	Effects     []EffectData `json:"effects,omitempty"`
}

// GameStateData represents the complete game state from server.
type GameStateData struct {
	Player       *PlayerState  `json:"player"`
	PartyMembers []PlayerState `json:"party_members,omitempty"`
	Combat       *CombatState  `json:"combat"`
	World        interface{}   `json:"world"`
	Session      *SessionData  `json:"session"`
}

// SessionData represents the current session information.
type SessionData struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// CombatModifiers holds cover and flanking information for attack targeting.
type CombatModifiers struct {
	CoverType     string `json:"cover_type"`
	CoverBonus    int    `json:"cover_bonus"`
	IsFlanking    bool   `json:"is_flanking"`
	FlankingBonus int    `json:"flanking_bonus"`
	AttackerPos   struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"attacker_pos"`
	DefenderPos struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"defender_pos"`
}

// VictoryData holds statistics displayed on the victory screen.
type VictoryData struct {
	AdventureTitle  string
	TimePlayed      string
	QuestsComplete  int
	QuestsTotal     int
	EnemiesDefeated int
	GoldEarned      int
	XPEarned        int
	LevelFrom       int
	LevelTo         int
}

// DefeatData holds information displayed on the defeat screen.
type DefeatData struct {
	LastLocation string
	CauseOfDeath string
}

// AttributeModifier calculates the D&D-style modifier for an attribute score.
func AttributeModifier(score int) int {
	return (score - 10) / 2
}

// StandardArray is the fixed attribute array per §4 Step 3.
var StandardArray = [6]int{15, 14, 13, 12, 10, 8}

// PointBuyCost returns the point cost for a given attribute score.
func PointBuyCost(score int) int {
	if score < 8 || score > 15 {
		return -1
	}
	costs := map[int]int{8: 0, 9: 1, 10: 2, 11: 3, 12: 4, 13: 5, 14: 7, 15: 9}
	return costs[score]
}

// PointBuyTotal is the total points available for point-buy.
const PointBuyTotal = 27
