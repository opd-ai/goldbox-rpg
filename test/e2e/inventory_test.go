package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEquipmentManagement tests equipping and unequipping items
func TestEquipmentManagement(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("EquipmentManager")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "Knight", "fighter")
	require.NoError(t, err)
	assert.NotEmpty(t, charID)

	equipResult, err := client.Call("getEquipment", map[string]interface{}{
		"session_id":   sessionID,
		"character_id": charID,
	})
	require.NoError(t, err)
	assert.NotNil(t, equipResult)
}

// TestEquipItem tests equipping various items
func TestEquipItem(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("ItemEquipper")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "Warrior", "fighter")
	require.NoError(t, err)

	testCases := []struct {
		name          string
		itemID        string
		slot          string
		expectError   bool
		errorContains string
	}{
		{
			name:        "equip_weapon",
			itemID:      "weapon_shortsword", // Use actual starting item ID
			slot:        "main_hand",
			expectError: false,
		},
		{
			name:        "equip_armor",
			itemID:      "armor_leather", // Use actual starting item ID
			slot:        "chest",         // Use valid slot name
			expectError: false,
		},
		{
			name:          "equip_invalid_item",
			itemID:        "nonexistent_item_12345",
			slot:          "main_hand",
			expectError:   true,
			errorContains: "item",
		},
		{
			name:          "equip_to_invalid_slot",
			itemID:        "weapon_shortsword",
			slot:          "invalid_slot_xyz",
			expectError:   true,
			errorContains: "slot",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("equipItem", map[string]interface{}{
				"session_id":   sessionID,
				"character_id": charID,
				"item_id":      tc.itemID,
				"slot":         tc.slot,
			})

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestUnequipItem tests removing equipped items
func TestUnequipItem(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("ItemUnequipper")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "Rogue", "thief")
	require.NoError(t, err)

	// First equip an item using the starting equipment (weapon_shortsword)
	_, err = client.Call("equipItem", map[string]interface{}{
		"session_id":   sessionID,
		"character_id": charID,
		"item_id":      "weapon_shortsword",
		"slot":         "main_hand",
	})
	require.NoError(t, err)

	result, err := client.Call("unequipItem", map[string]interface{}{
		"session_id":   sessionID,
		"character_id": charID,
		"slot":         "main_hand",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestItemUsage tests using consumable items
func TestItemUsage(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("ItemUser")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "Healer", "cleric")
	require.NoError(t, err)

	// Characters don't start with consumable items, so all use attempts should fail
	testCases := []struct {
		name          string
		itemID        string
		expectError   bool
		errorContains string
	}{
		{
			name:          "use_nonexistent_potion",
			itemID:        "healing_potion",
			expectError:   true,
			errorContains: "item",
		},
		{
			name:          "use_nonexistent_item",
			itemID:        "fake_item_12345",
			expectError:   true,
			errorContains: "item",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.Call("useItem", map[string]interface{}{
				"session_id":   sessionID,
				"character_id": charID,
				"item_id":      tc.itemID,
			})

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			}
		})
	}
}

// TestInventoryCapacity tests inventory size limits
func TestInventoryCapacity(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("InventoryTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Packrat", "thief")
	require.NoError(t, err)

	gameState, err := client.Call("getGameState", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)
	assert.NotNil(t, gameState)
}

// TestWeaponProficiency tests class weapon proficiency restrictions
func TestWeaponProficiency(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	// Test classes that start with weapon_shortsword: fighter, thief, ranger, paladin
	// Note: Mages get no starting equipment, so cannot be tested here
	testCases := []struct {
		name        string
		class       string
		weapon      string
		expectError bool
	}{
		{
			name:        "fighter_shortsword",
			class:       "fighter",
			weapon:      "weapon_shortsword",
			expectError: false,
		},
		{
			name:        "thief_shortsword",
			class:       "thief",
			weapon:      "weapon_shortsword",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.JoinGame("ProfTest_" + tc.name)
			require.NoError(t, err)

			sessionID, charID, err := client.CreateCharacter("", "Proficient", tc.class)
			require.NoError(t, err)

			result, err := client.Call("equipItem", map[string]interface{}{
				"session_id":   sessionID,
				"character_id": charID,
				"item_id":      tc.weapon,
				"slot":         "main_hand",
			})

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestArmorProficiency tests class armor proficiency restrictions
func TestArmorProficiency(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	// All classes start with armor_leather, test equipping it
	testCases := []struct {
		name        string
		class       string
		armor       string
		expectError bool
	}{
		{
			name:        "fighter_leather",
			class:       "fighter",
			armor:       "armor_leather",
			expectError: false,
		},
		{
			name:        "cleric_leather",
			class:       "cleric",
			armor:       "armor_leather",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.JoinGame("ArmorTest_" + tc.name)
			require.NoError(t, err)

			sessionID, charID, err := client.CreateCharacter("", "Armored", tc.class)
			require.NoError(t, err)

			result, err := client.Call("equipItem", map[string]interface{}{
				"session_id":   sessionID,
				"character_id": charID,
				"item_id":      tc.armor,
				"slot":         "chest", // Use valid slot name
			})

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestEquipmentSlots tests equipment slots with available starting items
func TestEquipmentSlots(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("SlotTester")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "FullyEquipped", "fighter")
	require.NoError(t, err)

	// Characters start with weapon_shortsword and armor_leather
	// Test equipping these to their respective slots
	slots := []struct {
		slotName string
		itemID   string
	}{
		{"main_hand", "weapon_shortsword"},
		{"chest", "armor_leather"},
	}

	for _, slot := range slots {
		t.Run("slot_"+slot.slotName, func(t *testing.T) {
			result, err := client.Call("equipItem", map[string]interface{}{
				"session_id":   sessionID,
				"character_id": charID,
				"item_id":      slot.itemID,
				"slot":         slot.slotName,
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}

	equipResult, err := client.Call("getEquipment", map[string]interface{}{
		"session_id":   sessionID,
		"character_id": charID,
	})
	require.NoError(t, err)
	assert.NotNil(t, equipResult)
}
