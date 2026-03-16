// Package wasmui provides the Ebitengine/WASM-based game UI client types.
package wasmui

import "image/color"

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
	Level      int              `json:"level"`
	Experience int              `json:"experience"`
	Class      string           `json:"class"`
	Attributes PlayerAttributes `json:"attributes"`
	Appearance *Appearance      `json:"appearance,omitempty"`
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

// CombatState represents the current combat state.
type CombatState struct {
	Active      bool              `json:"active"`
	CurrentTurn string            `json:"currentTurn"`
	Initiative  []InitiativeEntry `json:"initiative"`
	Round       int               `json:"round"`
	InCombat    bool              `json:"inCombat"`
}

// InitiativeEntry represents an entry in the combat initiative order.
type InitiativeEntry struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Initiative int    `json:"initiative"`
	IsPlayer   bool   `json:"isPlayer"`
}

// GameStateData represents the complete game state from server.
type GameStateData struct {
	Player  *PlayerState `json:"player"`
	Combat  *CombatState `json:"combat"`
	World   interface{}  `json:"world"`
	Session *SessionData `json:"session"`
}

// SessionData represents the current session information.
type SessionData struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// LogMessage represents a message in the combat/game log.
type LogMessage struct {
	Text      string
	Type      MessageType
	Timestamp int64
}

// MessageType defines the type of log message for styling.
type MessageType int

const (
	MessageInfo MessageType = iota
	MessageWarning
	MessageError
	MessageCombat
	MessageSystem
)

// Color returns the appropriate color for the message type.
func (mt MessageType) Color() color.RGBA {
	switch mt {
	case MessageWarning:
		return color.RGBA{R: 255, G: 200, B: 0, A: 255}
	case MessageError:
		return color.RGBA{R: 255, G: 100, B: 100, A: 255}
	case MessageCombat:
		return color.RGBA{R: 200, G: 150, B: 255, A: 255}
	case MessageSystem:
		return color.RGBA{R: 150, G: 200, B: 255, A: 255}
	default:
		return color.RGBA{R: 220, G: 220, B: 220, A: 255}
	}
}

// UIMode represents the current UI mode.
type UIMode int

const (
	ModeNormal UIMode = iota
	ModeCombat
	ModeInventory
	ModeSpellcasting
	ModeAdventureSelect
	ModeCharacterCreation
)
