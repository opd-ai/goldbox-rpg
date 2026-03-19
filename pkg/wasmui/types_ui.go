// Package wasmui provides the Ebitengine/WASM-based game UI client types.
// This file contains UI component types for screens, modes, and overlays.
package wasmui

import (
	"image/color"
	"time"
)

// Gold Box UI palette — EGA-inspired colors from ASSET_ANALYSIS.md.
// Used for panel borders, titles, and text throughout the UI.
var (
	// Border thickness for Gold Box-style double-pixel borders
	BorderThickness = 2

	// Panel framing — Bold, bright colors for authentic Gold Box aesthetic
	ColorPanelBG       = color.RGBA{R: 20, G: 18, B: 28, A: 255}    // deep near-black
	ColorPanelBorder   = color.RGBA{R: 120, G: 100, B: 180, A: 255} // vivid blue-purple border (bolder)
	ColorPanelBorderHi = color.RGBA{R: 180, G: 160, B: 255, A: 255} // bright highlight border (bolder)
	ColorPanelShadow   = color.RGBA{R: 50, G: 40, B: 70, A: 255}    // inner shadow line

	// Title / header text
	ColorGold   = color.RGBA{R: 191, G: 165, B: 74, A: 255}  // gold accent (#BFA54A)
	ColorGoldHi = color.RGBA{R: 220, G: 200, B: 100, A: 255} // bright gold for active items

	// Character panel text
	ColorPlayerName = color.RGBA{R: 100, G: 220, B: 100, A: 255} // green for player names
	ColorEnemyName  = color.RGBA{R: 220, G: 90, B: 90, A: 255}   // red for enemy names
	ColorStatLabel  = color.RGBA{R: 160, G: 160, B: 180, A: 255} // muted label text
	ColorStatValue  = color.RGBA{R: 220, G: 220, B: 220, A: 255} // bright value text

	// Resource state colors
	ColorAPDepleted = color.RGBA{R: 160, G: 80, B: 80, A: 255} // red-ish for zero AP

	// Effect type colors
	ColorEffectDebuff  = color.RGBA{R: 255, G: 100, B: 100, A: 255} // red for damage effects
	ColorEffectControl = color.RGBA{R: 255, G: 200, B: 0, A: 255}   // yellow for CC effects
	ColorEffectBuff    = color.RGBA{R: 100, G: 220, B: 100, A: 255} // green for beneficial effects
	ColorEffectDefault = color.RGBA{R: 200, G: 150, B: 255, A: 255} // purple default
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

// EncounterOverlay represents a Gold Box-style encounter/dialogue overlay panel.
// Renders centered over the viewport with optional NPC portrait and choices.
type EncounterOverlay struct {
	Visible        bool     // Whether the overlay is currently shown
	Title          string   // Title text (shown in gold)
	Text           string   // Body text (shown in white)
	PortraitPath   string   // Optional path to NPC portrait sprite
	Choices        []string // Optional multiple-choice options
	SelectedChoice int      // Currently selected choice index
}

// HasChoices returns true if the overlay has multiple choices.
func (e *EncounterOverlay) HasChoices() bool {
	return len(e.Choices) > 0
}

// SelectNext moves to the next choice in the list.
func (e *EncounterOverlay) SelectNext() {
	if len(e.Choices) > 0 {
		e.SelectedChoice = (e.SelectedChoice + 1) % len(e.Choices)
	}
}

// SelectPrev moves to the previous choice in the list.
func (e *EncounterOverlay) SelectPrev() {
	if len(e.Choices) > 0 {
		e.SelectedChoice = (e.SelectedChoice - 1 + len(e.Choices)) % len(e.Choices)
	}
}

// GetSelectedChoice returns the currently selected choice text, or empty string.
func (e *EncounterOverlay) GetSelectedChoice() string {
	if e.SelectedChoice >= 0 && e.SelectedChoice < len(e.Choices) {
		return e.Choices[e.SelectedChoice]
	}
	return ""
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

// DamageFlash represents a visual flash effect when an entity takes damage or heals.
// Used to provide Gold Box-style visual feedback during combat.
type DamageFlash struct {
	EntityID  string        // ID of the entity that was hit/healed
	StartTime time.Time     // When the flash started
	Duration  time.Duration // How long the flash lasts (typically ~200ms)
	Color     color.RGBA    // Flash color (red for damage, green for healing)
}

// IsActive returns true if the flash effect is still visible.
func (f *DamageFlash) IsActive() bool {
	return time.Since(f.StartTime) < f.Duration
}

// Alpha returns the flash intensity (0.0 to 1.0), fading over duration.
func (f *DamageFlash) Alpha() float32 {
	elapsed := time.Since(f.StartTime)
	if elapsed >= f.Duration {
		return 0
	}
	// Fade out linearly
	remaining := float32(f.Duration-elapsed) / float32(f.Duration)
	return remaining * 0.6 // Max alpha of 0.6 for visibility
}

// SpellEffect represents a visual spell effect animation on the combat grid.
// Used to show fireball explosions, lightning bolts, healing auras, etc.
type SpellEffect struct {
	SpellID      string        // ID of the spell for effect lookup
	SpellSchool  string        // School (Evocation, Necromancy, etc.) for color
	TargetPos    Position      // Grid position where effect plays
	StartTime    time.Time     // When the effect started
	Duration     time.Duration // Total duration of the effect (~400ms)
	CurrentFrame int           // Current animation frame (0-3)
	TotalFrames  int           // Total frames in animation
}

// IsActive returns true if the spell effect is still animating.
func (e *SpellEffect) IsActive() bool {
	return time.Since(e.StartTime) < e.Duration
}

// GetFrame returns the current animation frame index (0 to TotalFrames-1).
func (e *SpellEffect) GetFrame() int {
	elapsed := time.Since(e.StartTime)
	frameTime := e.Duration / time.Duration(e.TotalFrames)
	frame := int(elapsed / frameTime)
	if frame >= e.TotalFrames {
		return e.TotalFrames - 1
	}
	return frame
}

// GetRadius returns the current effect radius for expanding circle fallback.
func (e *SpellEffect) GetRadius() float64 {
	elapsed := float64(time.Since(e.StartTime)) / float64(e.Duration)
	if elapsed > 1.0 {
		elapsed = 1.0
	}
	// Expand from 5 to 25 pixels over duration
	return 5.0 + (elapsed * 20.0)
}

// GetAlpha returns effect alpha for fade-out in second half.
func (e *SpellEffect) GetAlpha() float32 {
	elapsed := float64(time.Since(e.StartTime)) / float64(e.Duration)
	if elapsed < 0.5 {
		return 1.0
	}
	// Fade out in second half
	return float32(1.0 - ((elapsed - 0.5) * 2.0))
}

// SpellSchoolColor returns the color associated with a spell school.
func SpellSchoolColor(school string) color.RGBA {
	switch school {
	case "Evocation":
		return color.RGBA{R: 255, G: 100, B: 50, A: 255} // Fire orange
	case "Necromancy":
		return color.RGBA{R: 80, G: 40, B: 120, A: 255} // Dark purple
	case "Conjuration":
		return color.RGBA{R: 50, G: 200, B: 255, A: 255} // Cyan
	case "Abjuration":
		return color.RGBA{R: 255, G: 255, B: 150, A: 255} // Yellow glow
	case "Enchantment":
		return color.RGBA{R: 200, G: 100, B: 255, A: 255} // Pink/purple
	case "Illusion":
		return color.RGBA{R: 150, G: 150, B: 200, A: 255} // Silver/gray
	case "Transmutation":
		return color.RGBA{R: 100, G: 255, B: 100, A: 255} // Green
	case "Divination":
		return color.RGBA{R: 255, G: 255, B: 255, A: 255} // White
	default:
		return color.RGBA{R: 150, G: 150, B: 255, A: 255} // Default blue
	}
}
