package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetMaxActionPoints tests the GetMaxActionPoints method
func TestCharacter_GetMaxActionPoints(t *testing.T) {
	char := &Character{
		ID:              "test-char",
		Name:            "Test Fighter",
		MaxActionPoints: 5,
	}

	maxAP := char.GetMaxActionPoints()
	assert.Equal(t, 5, maxAP)
}

// TestConsumeActionPoints tests the ConsumeActionPoints method
func TestCharacter_ConsumeActionPoints(t *testing.T) {
	tests := []struct {
		name            string
		initialAP       int
		cost            int
		expectSuccess   bool
		expectedFinalAP int
	}{
		{
			name:            "Sufficient points",
			initialAP:       5,
			cost:            3,
			expectSuccess:   true,
			expectedFinalAP: 2,
		},
		{
			name:            "Exact amount",
			initialAP:       3,
			cost:            3,
			expectSuccess:   true,
			expectedFinalAP: 0,
		},
		{
			name:            "Insufficient points",
			initialAP:       2,
			cost:            5,
			expectSuccess:   false,
			expectedFinalAP: 2,
		},
		{
			name:            "Zero cost",
			initialAP:       5,
			cost:            0,
			expectSuccess:   true,
			expectedFinalAP: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char := &Character{
				ID:           "test-char",
				ActionPoints: tt.initialAP,
			}

			result := char.ConsumeActionPoints(tt.cost)
			assert.Equal(t, tt.expectSuccess, result)
			assert.Equal(t, tt.expectedFinalAP, char.ActionPoints)
		})
	}
}

// TestRestoreActionPoints tests the RestoreActionPoints method
func TestCharacter_RestoreActionPoints(t *testing.T) {
	char := &Character{
		ID:              "test-char",
		ActionPoints:    2,
		MaxActionPoints: 5,
	}

	char.RestoreActionPoints()
	assert.Equal(t, 5, char.ActionPoints)
}

// TestGetEquippedItem tests the GetEquippedItem method
func TestCharacter_GetEquippedItem(t *testing.T) {
	sword := Item{
		ID:   "sword-1",
		Name: "Iron Sword",
		Type: "weapon",
	}

	char := &Character{
		ID:   "test-char",
		Name: "Test Fighter",
		Equipment: map[EquipmentSlot]Item{
			SlotWeaponMain: sword,
		},
	}

	// Test getting equipped item
	item, exists := char.GetEquippedItem(SlotWeaponMain)
	require.True(t, exists)
	require.NotNil(t, item)
	assert.Equal(t, "sword-1", item.ID)

	// Test empty slot
	item, exists = char.GetEquippedItem(SlotHead)
	assert.False(t, exists)
	assert.Nil(t, item)
}

// TestGetAllEquippedItems tests the GetAllEquippedItems method
func TestCharacter_GetAllEquippedItems(t *testing.T) {
	sword := Item{ID: "sword-1", Name: "Iron Sword", Type: "weapon"}
	helmet := Item{ID: "helmet-1", Name: "Iron Helmet", Type: "armor"}

	char := &Character{
		ID:   "test-char",
		Name: "Test Fighter",
		Equipment: map[EquipmentSlot]Item{
			SlotWeaponMain: sword,
			SlotHead:       helmet,
		},
	}

	equipped := char.GetAllEquippedItems()
	assert.Len(t, equipped, 2)
	assert.Equal(t, "sword-1", equipped[SlotWeaponMain].ID)
	assert.Equal(t, "helmet-1", equipped[SlotHead].ID)
}

// TestGetEquipmentSlots tests the GetEquipmentSlots method
func TestCharacter_GetEquipmentSlots(t *testing.T) {
	char := &Character{ID: "test-char"}

	slots := char.GetEquipmentSlots()
	assert.Len(t, slots, 9)

	// Verify all expected slots are present
	expectedSlots := []EquipmentSlot{
		SlotHead, SlotNeck, SlotChest, SlotHands, SlotRings,
		SlotLegs, SlotFeet, SlotWeaponMain, SlotWeaponOff,
	}
	for _, slot := range expectedSlots {
		assert.Contains(t, slots, slot)
	}
}

// TestFindItemInInventory tests the FindItemInInventory method
func TestCharacter_FindItemInInventory(t *testing.T) {
	items := []Item{
		{ID: "item-1", Name: "Potion"},
		{ID: "item-2", Name: "Scroll"},
		{ID: "item-3", Name: "Key"},
	}

	char := &Character{
		ID:        "test-char",
		Inventory: items,
	}

	// Test finding existing item
	item, idx := char.FindItemInInventory("item-2")
	require.NotNil(t, item)
	assert.Equal(t, 1, idx)
	assert.Equal(t, "Scroll", item.Name)

	// Test finding non-existent item
	item, idx = char.FindItemInInventory("item-999")
	assert.Nil(t, item)
	assert.Equal(t, -1, idx)
}

// TestGetInventory tests the GetInventory method
func TestCharacter_GetInventory(t *testing.T) {
	items := []Item{
		{ID: "item-1", Name: "Potion"},
		{ID: "item-2", Name: "Scroll"},
	}

	char := &Character{
		ID:        "test-char",
		Inventory: items,
	}

	inventory := char.GetInventory()

	// Verify it's a copy, not the original
	assert.Len(t, inventory, 2)
	assert.Equal(t, "item-1", inventory[0].ID)
	assert.Equal(t, "item-2", inventory[1].ID)

	// Verify modifications don't affect original
	inventory[0].Name = "Modified"
	assert.Equal(t, "Potion", char.Inventory[0].Name)
}

// TestGetEffectManager tests the GetEffectManager method
func TestCharacter_GetEffectManager(t *testing.T) {
	char := &Character{
		ID:       "test-char",
		Strength: 14,
	}

	// Initially nil
	assert.Nil(t, char.EffectManager)

	// GetEffectManager should initialize it
	em := char.GetEffectManager()
	require.NotNil(t, em)

	// Second call should return same instance
	em2 := char.GetEffectManager()
	assert.Same(t, em, em2)
}

// TestCharacterEffectMethods tests AddEffect, RemoveEffect, HasEffect, GetEffects
func TestCharacter_EffectMethods(t *testing.T) {
	char := &Character{
		ID:       "test-char",
		Name:     "Test Fighter",
		Strength: 14,
		MaxHP:    100,
		HP:       100,
	}

	// Initial state - no effects
	assert.False(t, char.HasEffect(EffectPoison))
	assert.Empty(t, char.GetEffects())

	// Add an effect
	effect := NewEffect(EffectPoison, NewDuration(5, 0, 0), 10.0)
	err := char.AddEffect(effect)
	require.NoError(t, err)

	// Verify effect was added
	assert.True(t, char.HasEffect(EffectPoison))
	effects := char.GetEffects()
	assert.Len(t, effects, 1)

	// Remove the effect
	err = char.RemoveEffect(effect.ID)
	require.NoError(t, err)

	// Verify effect was removed
	assert.False(t, char.HasEffect(EffectPoison))
	assert.Empty(t, char.GetEffects())
}

// TestCharacterGetStats tests GetStats method
func TestCharacter_GetStats(t *testing.T) {
	char := &Character{
		ID:           "test-char",
		Strength:     15,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
		Wisdom:       13,
		Charisma:     11,
	}

	stats := char.GetStats()
	require.NotNil(t, stats)

	// Stats should reflect character attributes (converted to float64)
	assert.Equal(t, float64(15), stats.Strength)
	assert.Equal(t, float64(12), stats.Dexterity)
}

// TestCharacterSetStats tests SetStats method
func TestCharacter_SetStats(t *testing.T) {
	char := &Character{
		ID:       "test-char",
		Strength: 10,
	}

	newStats := &Stats{
		Strength:     18,
		Dexterity:    16,
		Intelligence: 12,
	}

	char.SetStats(newStats)

	stats := char.GetStats()
	assert.Equal(t, float64(18), stats.Strength)
}

// TestCharacterGetBaseStats tests GetBaseStats method
func TestCharacter_GetBaseStats(t *testing.T) {
	char := &Character{
		ID:           "test-char",
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
	}

	baseStats := char.GetBaseStats()
	require.NotNil(t, baseStats)
	assert.Equal(t, float64(15), baseStats.Strength)
}

// TestActionPointsConcurrency tests thread safety of action point methods
func TestCharacter_ActionPointsConcurrency(t *testing.T) {
	char := &Character{
		ID:              "test-char",
		ActionPoints:    100,
		MaxActionPoints: 100,
	}

	done := make(chan bool)

	// Concurrently consume action points
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				char.ConsumeActionPoints(1)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have consumed 100 points total
	assert.Equal(t, 0, char.ActionPoints)
}
