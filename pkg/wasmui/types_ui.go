// Package wasmui provides the Ebitengine/WASM-based game UI client types.
// This file contains UI component types for screens, modes, and overlays.
package wasmui

import "image/color"

// Gold Box UI palette — EGA-inspired colors from ASSET_ANALYSIS.md.
// Used for panel borders, titles, and text throughout the UI.
var (
	// Panel framing
	ColorPanelBG       = color.RGBA{R: 20, G: 18, B: 28, A: 255}    // deep near-black
	ColorPanelBorder   = color.RGBA{R: 90, G: 80, B: 130, A: 255}   // vivid blue-purple border
	ColorPanelBorderHi = color.RGBA{R: 130, G: 115, B: 180, A: 255} // bright highlight border

	// Title / header text
	ColorGold   = color.RGBA{R: 191, G: 165, B: 74, A: 255}  // gold accent (#BFA54A)
	ColorGoldHi = color.RGBA{R: 220, G: 200, B: 100, A: 255} // bright gold for active items

	// Character panel text
	ColorPlayerName = color.RGBA{R: 100, G: 220, B: 100, A: 255} // green for player names
	ColorEnemyName  = color.RGBA{R: 220, G: 90, B: 90, A: 255}   // red for enemy names
	ColorStatLabel  = color.RGBA{R: 160, G: 160, B: 180, A: 255} // muted label text
	ColorStatValue  = color.RGBA{R: 220, G: 220, B: 220, A: 255} // bright value text
)

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

// ScreenState tracks the current screen within ModeNormal.
type ScreenState int

const (
	ScreenSplash ScreenState = iota
	ScreenMainMenu
	ScreenExploration
	ScreenVictory
	ScreenDefeat
)

// OverlayState tracks which overlay panels are open.
// These are drawn on top of the current mode per §2 note.
type OverlayState struct {
	ShowQuestLog   bool
	ShowGuildPanel bool
	ShowSettings   bool
}

// CharCreationStep tracks the current step within character creation.
type CharCreationStep int

const (
	CharStepName CharCreationStep = iota
	CharStepClass
	CharStepAttributes
	CharStepReview
)

// AttributeMethod determines how attributes are generated.
type AttributeMethod int

const (
	AttrMethodRoll AttributeMethod = iota
	AttrMethodStandard
	AttrMethodPointBuy
	AttrMethodCustom
)

// String returns the display name of the attribute method.
func (m AttributeMethod) String() string {
	switch m {
	case AttrMethodRoll:
		return "Roll"
	case AttrMethodStandard:
		return "Standard"
	case AttrMethodPointBuy:
		return "Point Buy"
	case AttrMethodCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// CharCreationState holds temporary state for the character creation flow.
type CharCreationState struct {
	Step           CharCreationStep
	Name           string
	SelectedClass  int // index into ClassInfoList
	AttrMethod     AttributeMethod
	Attributes     PlayerAttributes
	SelectedAttr   int // index 0-5 for attribute selection in point-buy / adjustment
	PointBuyPoints int
}

// GetAttr returns the attribute value at index i (0=STR..5=CHA).
func (c *CharCreationState) GetAttr(i int) int {
	switch i {
	case 0:
		return c.Attributes.Strength
	case 1:
		return c.Attributes.Dexterity
	case 2:
		return c.Attributes.Constitution
	case 3:
		return c.Attributes.Intelligence
	case 4:
		return c.Attributes.Wisdom
	case 5:
		return c.Attributes.Charisma
	}
	return 0
}

// SetAttr sets the attribute value at index i (0=STR..5=CHA).
func (c *CharCreationState) SetAttr(i, val int) {
	switch i {
	case 0:
		c.Attributes.Strength = val
	case 1:
		c.Attributes.Dexterity = val
	case 2:
		c.Attributes.Constitution = val
	case 3:
		c.Attributes.Intelligence = val
	case 4:
		c.Attributes.Wisdom = val
	case 5:
		c.Attributes.Charisma = val
	}
}

// SetStandardArray loads the standard array into the attributes.
func (c *CharCreationState) SetStandardArray() {
	c.Attributes.Strength = StandardArray[0]
	c.Attributes.Dexterity = StandardArray[1]
	c.Attributes.Constitution = StandardArray[2]
	c.Attributes.Intelligence = StandardArray[3]
	c.Attributes.Wisdom = StandardArray[4]
	c.Attributes.Charisma = StandardArray[5]
}

// ResetAttributes sets all attribute values to the given value.
func (c *CharCreationState) ResetAttributes(val int) {
	c.Attributes.Strength = val
	c.Attributes.Dexterity = val
	c.Attributes.Constitution = val
	c.Attributes.Intelligence = val
	c.Attributes.Wisdom = val
	c.Attributes.Charisma = val
}

// ClassInfo describes a character class for the creation screen.
type ClassInfo struct {
	Name                string
	Description         string
	HitDice             string
	WeaponProficiencies []string
	ArmorProficiencies  []string
	ShieldProficiency   bool
	Restrictions        string
}

// ClassInfoList is the ordered list of selectable classes per §4.
var ClassInfoList = []ClassInfo{
	{Name: "Fighter", Description: "Melee combat specialist with high HP", HitDice: "d10", WeaponProficiencies: []string{"sword", "axe", "mace", "bow", "dagger", "spear", "wand", "staff"}, ArmorProficiencies: []string{"light", "medium", "heavy"}, ShieldProficiency: true},
	{Name: "Mage", Description: "Arcane spellcaster with powerful magic", HitDice: "d4", WeaponProficiencies: []string{"staff", "dagger", "wand"}, ArmorProficiencies: []string{}, ShieldProficiency: false, Restrictions: "No armor, no shields"},
	{Name: "Cleric", Description: "Divine healer and support caster", HitDice: "d8", WeaponProficiencies: []string{"mace", "staff", "dagger"}, ArmorProficiencies: []string{"light", "medium", "heavy"}, ShieldProficiency: true, Restrictions: "No edged weapons"},
	{Name: "Thief", Description: "Stealthy rogue with trap expertise", HitDice: "d6", WeaponProficiencies: []string{"dagger", "sword", "bow"}, ArmorProficiencies: []string{"light"}, ShieldProficiency: false},
	{Name: "Ranger", Description: "Wilderness warrior with tracking skills", HitDice: "d8", WeaponProficiencies: []string{"bow", "sword", "dagger", "spear"}, ArmorProficiencies: []string{"light", "medium"}, ShieldProficiency: true},
	{Name: "Paladin", Description: "Holy warrior combining arms and faith", HitDice: "d10", WeaponProficiencies: []string{"sword", "mace", "spear", "bow", "dagger"}, ArmorProficiencies: []string{"light", "medium", "heavy"}, ShieldProficiency: true},
}

// CombatAction represents the currently selected combat action.
type CombatAction int

const (
	CombatActionNone CombatAction = iota
	CombatActionMove
	CombatActionAttack
	CombatActionCast
	CombatActionItem
	CombatActionDefend
	CombatActionFlee
)

// String returns the display name of the combat action.
func (a CombatAction) String() string {
	switch a {
	case CombatActionNone:
		return "None"
	case CombatActionMove:
		return "Move"
	case CombatActionAttack:
		return "Attack"
	case CombatActionCast:
		return "Cast"
	case CombatActionItem:
		return "Item"
	case CombatActionDefend:
		return "Defend"
	case CombatActionFlee:
		return "Flee"
	default:
		return "Unknown"
	}
}
