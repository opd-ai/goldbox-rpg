package game

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacterError(t *testing.T) {
	charID := "char-123"
	baseErr := ErrNegativeExperience
	err := NewCharacterError(charID, "addExperience", baseErr)

	// Test error message
	assert.Contains(t, err.Error(), charID)
	assert.Contains(t, err.Error(), "addExperience")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var charErr *CharacterError
	require.True(t, errors.As(err, &charErr))
	assert.Equal(t, charID, charErr.CharacterID)
	assert.Equal(t, "addExperience", charErr.Operation)
}

func TestInventoryError(t *testing.T) {
	itemID := "sword-01"
	baseErr := ErrCarryingCapacity
	invErr := NewInventoryError(itemID, "add", baseErr)
	invErr.CurrentLoad = 150
	invErr.MaxLoad = 100

	// Test error message
	assert.Contains(t, invErr.Error(), itemID)
	assert.Contains(t, invErr.Error(), "add")
	assert.Contains(t, invErr.Error(), "150")
	assert.Contains(t, invErr.Error(), "100")

	// Test error unwrapping
	assert.True(t, errors.Is(invErr, baseErr))

	// Test error.As
	var ie *InventoryError
	require.True(t, errors.As(invErr, &ie))
	assert.Equal(t, itemID, ie.ItemID)
	assert.Equal(t, 150, ie.CurrentLoad)
	assert.Equal(t, 100, ie.MaxLoad)
}

func TestInventoryErrorEmptyID(t *testing.T) {
	err := NewInventoryError("", "add", ErrEmptyItemID)

	assert.Contains(t, err.Error(), "add")
	assert.True(t, errors.Is(err, ErrEmptyItemID))
}

func TestCombatError(t *testing.T) {
	attackerID := "player-1"
	targetID := "enemy-5"
	baseErr := ErrTargetOutOfRange
	err := NewCombatError(attackerID, targetID, "attack", baseErr)

	// Test error message
	assert.Contains(t, err.Error(), attackerID)
	assert.Contains(t, err.Error(), targetID)
	assert.Contains(t, err.Error(), "attack")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var combatErr *CombatError
	require.True(t, errors.As(err, &combatErr))
	assert.Equal(t, attackerID, combatErr.AttackerID)
	assert.Equal(t, targetID, combatErr.TargetID)
}

func TestEffectError(t *testing.T) {
	effectID := "stun-01"
	targetID := "player-1"
	effectType := "stun"
	baseErr := ErrEffectImmunity
	err := NewEffectError(effectID, targetID, effectType, "apply", baseErr)

	// Test error message
	assert.Contains(t, err.Error(), effectID)
	assert.Contains(t, err.Error(), targetID)
	assert.Contains(t, err.Error(), effectType)
	assert.Contains(t, err.Error(), "apply")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var effErr *EffectError
	require.True(t, errors.As(err, &effErr))
	assert.Equal(t, effectID, effErr.EffectID)
	assert.Equal(t, targetID, effErr.TargetID)
	assert.Equal(t, effectType, effErr.EffectType)
}

func TestSpatialError(t *testing.T) {
	objectID := "obj-123"
	position := &Position{X: 150, Y: 200}
	baseErr := ErrOutOfBounds
	err := NewSpatialError(objectID, position, "insert", baseErr)

	// Test error message
	assert.Contains(t, err.Error(), objectID)
	assert.Contains(t, err.Error(), "150")
	assert.Contains(t, err.Error(), "200")
	assert.Contains(t, err.Error(), "insert")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var spatialErr *SpatialError
	require.True(t, errors.As(err, &spatialErr))
	assert.Equal(t, objectID, spatialErr.ObjectID)
	assert.Equal(t, 150, spatialErr.Position.X)
	assert.Equal(t, 200, spatialErr.Position.Y)
}

func TestSpatialErrorNilPosition(t *testing.T) {
	objectID := "obj-123"
	baseErr := ErrObjectNotFound
	err := NewSpatialError(objectID, nil, "remove", baseErr)

	assert.Contains(t, err.Error(), objectID)
	assert.Contains(t, err.Error(), "remove")
	assert.True(t, errors.Is(err, baseErr))
}

func TestSpellError(t *testing.T) {
	spellID := "fireball"
	casterID := "wizard-1"
	targetID := "enemy-1"
	baseErr := ErrInsufficientMP
	err := NewSpellError(spellID, casterID, targetID, "cast", baseErr)

	// Test error message
	assert.Contains(t, err.Error(), spellID)
	assert.Contains(t, err.Error(), casterID)
	assert.Contains(t, err.Error(), "cast")

	// Test error unwrapping
	assert.True(t, errors.Is(err, baseErr))

	// Test error.As
	var spellErr *SpellError
	require.True(t, errors.As(err, &spellErr))
	assert.Equal(t, spellID, spellErr.SpellID)
	assert.Equal(t, casterID, spellErr.CasterID)
}

func TestErrorChaining(t *testing.T) {
	// Test error chain: CharacterError -> InventoryError -> ErrItemNotFound
	itemErr := NewInventoryError("sword-01", "remove", ErrItemNotFound)
	charErr := NewCharacterError("char-123", "unequipItem", itemErr)

	// Should be able to unwrap through chain
	assert.True(t, errors.Is(charErr, ErrItemNotFound))

	// Should be able to extract both error types
	var ie *InventoryError
	require.True(t, errors.As(charErr, &ie))
	assert.Equal(t, "sword-01", ie.ItemID)

	var ce *CharacterError
	require.True(t, errors.As(charErr, &ce))
	assert.Equal(t, "char-123", ce.CharacterID)
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrCharacterNotFound", ErrCharacterNotFound},
		{"ErrInvalidCharacterName", ErrInvalidCharacterName},
		{"ErrInsufficientHP", ErrInsufficientHP},
		{"ErrInsufficientMP", ErrInsufficientMP},
		{"ErrInsufficientAP", ErrInsufficientAP},
		{"ErrExperienceOverflow", ErrExperienceOverflow},
		{"ErrNegativeExperience", ErrNegativeExperience},
		{"ErrInvalidTarget", ErrInvalidTarget},
		{"ErrTargetOutOfRange", ErrTargetOutOfRange},
		{"ErrCannotReceiveDamage", ErrCannotReceiveDamage},
		{"ErrEffectReflected", ErrEffectReflected},
		{"ErrItemNotFound", ErrItemNotFound},
		{"ErrEmptyItemID", ErrEmptyItemID},
		{"ErrCarryingCapacity", ErrCarryingCapacity},
		{"ErrCannotEquipItem", ErrCannotEquipItem},
		{"ErrInvalidSlot", ErrInvalidSlot},
		{"ErrEffectNotFound", ErrEffectNotFound},
		{"ErrWeakerEffect", ErrWeakerEffect},
		{"ErrInvalidEffectType", ErrInvalidEffectType},
		{"ErrEffectImmunity", ErrEffectImmunity},
		{"ErrCannotCastSpells", ErrCannotCastSpells},
		{"ErrInvalidSpellLevel", ErrInvalidSpellLevel},
		{"ErrSpellNotKnown", ErrSpellNotKnown},
		{"ErrInvalidDiceExpression", ErrInvalidDiceExpression},
		{"ErrObjectNotFound", ErrObjectNotFound},
		{"ErrOutOfBounds", ErrOutOfBounds},
		{"ErrInvalidPosition", ErrInvalidPosition},
		{"ErrWorldNotInitialized", ErrWorldNotInitialized},
		{"ErrInvalidWorldState", ErrInvalidWorldState},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All sentinel errors should be non-nil and have a message
			assert.NotNil(t, tt.err)
			assert.NotEmpty(t, tt.err.Error())

			// Each sentinel error should match itself
			assert.True(t, errors.Is(tt.err, tt.err))
		})
	}
}

func TestMultiLevelErrorUnwrapping(t *testing.T) {
	// Create a multi-level error chain
	baseErr := ErrItemNotFound
	invErr := NewInventoryError("sword-99", "find", baseErr)
	charErr := NewCharacterError("char-456", "equipItem", invErr)

	// Should unwrap through all levels
	assert.True(t, errors.Is(charErr, ErrItemNotFound))
	assert.True(t, errors.Is(invErr, ErrItemNotFound))
}

func TestErrorsAreUnique(t *testing.T) {
	// Each sentinel error should be unique
	assert.False(t, errors.Is(ErrItemNotFound, ErrCharacterNotFound))
	assert.False(t, errors.Is(ErrInsufficientHP, ErrInsufficientMP))
	assert.False(t, errors.Is(ErrObjectNotFound, ErrItemNotFound))
}
