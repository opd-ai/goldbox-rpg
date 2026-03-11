package game

import (
	"errors"
	"fmt"
)

// Sentinel errors for common game conditions
var (
	// Character errors
	ErrCharacterNotFound    = errors.New("character not found")
	ErrInvalidCharacterName = errors.New("invalid character name")
	ErrInsufficientHP       = errors.New("insufficient hit points")
	ErrInsufficientMP       = errors.New("insufficient mana points")
	ErrInsufficientAP       = errors.New("insufficient action points")
	ErrExperienceOverflow   = errors.New("experience overflow")
	ErrNegativeExperience   = errors.New("negative experience not allowed")

	// Combat errors
	ErrInvalidTarget       = errors.New("invalid target")
	ErrTargetOutOfRange    = errors.New("target out of range")
	ErrCannotReceiveDamage = errors.New("target cannot receive damage")
	ErrEffectReflected     = errors.New("effect reflected")

	// Inventory errors
	ErrItemNotFound     = errors.New("item not found in inventory")
	ErrEmptyItemID      = errors.New("item ID cannot be empty")
	ErrCarryingCapacity = errors.New("carrying capacity exceeded")
	ErrCannotEquipItem  = errors.New("cannot equip item")
	ErrInvalidSlot      = errors.New("invalid equipment slot")

	// Effect errors
	ErrEffectNotFound    = errors.New("effect not found")
	ErrWeakerEffect      = errors.New("cannot apply weaker effect of same type")
	ErrInvalidEffectType = errors.New("invalid effect type")
	ErrEffectImmunity    = errors.New("target is immune to effect")

	// Spell errors
	ErrCannotCastSpells      = errors.New("class cannot cast spells")
	ErrInvalidSpellLevel     = errors.New("invalid spell level")
	ErrSpellNotKnown         = errors.New("spell not known by character")
	ErrInvalidDiceExpression = errors.New("invalid dice expression")

	// Spatial errors
	ErrObjectNotFound  = errors.New("object not found")
	ErrOutOfBounds     = errors.New("position outside bounds")
	ErrInvalidPosition = errors.New("invalid position")

	// World errors
	ErrWorldNotInitialized = errors.New("world not initialized")
	ErrInvalidWorldState   = errors.New("invalid world state")
)

// CharacterError represents errors related to character operations
type CharacterError struct {
	CharacterID string
	Operation   string
	Err         error
}

func (e *CharacterError) Error() string {
	if e.CharacterID != "" {
		return fmt.Sprintf("character %s: %s: %v", e.CharacterID, e.Operation, e.Err)
	}
	return fmt.Sprintf("character: %s: %v", e.Operation, e.Err)
}

func (e *CharacterError) Unwrap() error {
	return e.Err
}

// NewCharacterError creates a new CharacterError
func NewCharacterError(characterID, operation string, err error) *CharacterError {
	return &CharacterError{
		CharacterID: characterID,
		Operation:   operation,
		Err:         err,
	}
}

// InventoryError represents errors related to inventory operations
type InventoryError struct {
	ItemID      string
	Operation   string
	CurrentLoad int
	MaxLoad     int
	Err         error
}

func (e *InventoryError) Error() string {
	if e.CurrentLoad > 0 {
		return fmt.Sprintf("inventory: %s: item %s: %v (load: %d/%d)",
			e.Operation, e.ItemID, e.Err, e.CurrentLoad, e.MaxLoad)
	}
	if e.ItemID != "" {
		return fmt.Sprintf("inventory: %s: item %s: %v", e.Operation, e.ItemID, e.Err)
	}
	return fmt.Sprintf("inventory: %s: %v", e.Operation, e.Err)
}

func (e *InventoryError) Unwrap() error {
	return e.Err
}

// NewInventoryError creates a new InventoryError
func NewInventoryError(itemID, operation string, err error) *InventoryError {
	return &InventoryError{
		ItemID:    itemID,
		Operation: operation,
		Err:       err,
	}
}

// CombatError represents errors during combat operations
type CombatError struct {
	AttackerID string
	TargetID   string
	Action     string
	Err        error
}

func (e *CombatError) Error() string {
	if e.AttackerID != "" && e.TargetID != "" {
		return fmt.Sprintf("combat: %s: attacker %s -> target %s: %v",
			e.Action, e.AttackerID, e.TargetID, e.Err)
	}
	return fmt.Sprintf("combat: %s: %v", e.Action, e.Err)
}

func (e *CombatError) Unwrap() error {
	return e.Err
}

// NewCombatError creates a new CombatError
func NewCombatError(attackerID, targetID, action string, err error) *CombatError {
	return &CombatError{
		AttackerID: attackerID,
		TargetID:   targetID,
		Action:     action,
		Err:        err,
	}
}

// EffectError represents errors related to effect management
type EffectError struct {
	EffectID   string
	TargetID   string
	EffectType string
	Operation  string
	Err        error
}

func (e *EffectError) Error() string {
	if e.EffectID != "" && e.TargetID != "" {
		return fmt.Sprintf("effect: %s: effect %s on target %s (type: %s): %v",
			e.Operation, e.EffectID, e.TargetID, e.EffectType, e.Err)
	}
	if e.EffectID != "" {
		return fmt.Sprintf("effect: %s: effect %s: %v", e.Operation, e.EffectID, e.Err)
	}
	return fmt.Sprintf("effect: %s: %v", e.Operation, e.Err)
}

func (e *EffectError) Unwrap() error {
	return e.Err
}

// NewEffectError creates a new EffectError
func NewEffectError(effectID, targetID, effectType, operation string, err error) *EffectError {
	return &EffectError{
		EffectID:   effectID,
		TargetID:   targetID,
		EffectType: effectType,
		Operation:  operation,
		Err:        err,
	}
}

// SpatialError represents errors in spatial indexing operations
type SpatialError struct {
	ObjectID  string
	Position  *Position
	Bounds    *Rectangle
	Operation string
	Err       error
}

func (e *SpatialError) Error() string {
	if e.ObjectID != "" && e.Position != nil {
		return fmt.Sprintf("spatial: %s: object %s at position (%d,%d): %v",
			e.Operation, e.ObjectID, e.Position.X, e.Position.Y, e.Err)
	}
	if e.ObjectID != "" {
		return fmt.Sprintf("spatial: %s: object %s: %v", e.Operation, e.ObjectID, e.Err)
	}
	return fmt.Sprintf("spatial: %s: %v", e.Operation, e.Err)
}

func (e *SpatialError) Unwrap() error {
	return e.Err
}

// NewSpatialError creates a new SpatialError
func NewSpatialError(objectID string, position *Position, operation string, err error) *SpatialError {
	return &SpatialError{
		ObjectID:  objectID,
		Position:  position,
		Operation: operation,
		Err:       err,
	}
}

// SpellError represents errors related to spell operations
type SpellError struct {
	SpellID   string
	CasterID  string
	TargetID  string
	Operation string
	Err       error
}

func (e *SpellError) Error() string {
	if e.SpellID != "" && e.CasterID != "" {
		return fmt.Sprintf("spell: %s: spell %s cast by %s: %v",
			e.Operation, e.SpellID, e.CasterID, e.Err)
	}
	if e.SpellID != "" {
		return fmt.Sprintf("spell: %s: spell %s: %v", e.Operation, e.SpellID, e.Err)
	}
	return fmt.Sprintf("spell: %s: %v", e.Operation, e.Err)
}

func (e *SpellError) Unwrap() error {
	return e.Err
}

// NewSpellError creates a new SpellError
func NewSpellError(spellID, casterID, targetID, operation string, err error) *SpellError {
	return &SpellError{
		SpellID:   spellID,
		CasterID:  casterID,
		TargetID:  targetID,
		Operation: operation,
		Err:       err,
	}
}
