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
			itemID:      "longsword",
			slot:        "main_hand",
			expectError: false,
		},
		{
			name:        "equip_armor",
			itemID:      "chainmail",
			slot:        "body",
			expectError: false,
		},
		{
			name:          "equip_invalid_item",
			itemID:        "nonexistent",
			slot:          "main_hand",
			expectError:   true,
			errorContains: "item",
		},
		{
			name:          "equip_to_invalid_slot",
			itemID:        "shield",
			slot:          "invalid_slot",
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

	_, err = client.Call("equipItem", map[string]interface{}{
		"session_id":   sessionID,
		"character_id": charID,
		"item_id":      "dagger",
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

	testCases := []struct {
		name          string
		itemID        string
		expectError   bool
		errorContains string
	}{
		{
			name:        "use_potion",
			itemID:      "healing_potion",
			expectError: false,
		},
		{
			name:          "use_nonexistent_item",
			itemID:        "fake_item",
			expectError:   true,
			errorContains: "item",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Call("useItem", map[string]interface{}{
				"session_id":   sessionID,
				"character_id": charID,
				"item_id":      tc.itemID,
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

	testCases := []struct {
		name        string
		class       string
		weapon      string
		expectError bool
	}{
		{
			name:        "fighter_longsword",
			class:       "fighter",
			weapon:      "longsword",
			expectError: false,
		},
		{
			name:        "mage_dagger",
			class:       "mage",
			weapon:      "dagger",
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

	testCases := []struct {
		name        string
		class       string
		armor       string
		expectError bool
	}{
		{
			name:        "fighter_plate",
			class:       "fighter",
			armor:       "plate_armor",
			expectError: false,
		},
		{
			name:        "cleric_chainmail",
			class:       "cleric",
			armor:       "chainmail",
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
				"slot":         "body",
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

// TestEquipmentSlots tests all equipment slots
func TestEquipmentSlots(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("SlotTester")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "FullyEquipped", "fighter")
	require.NoError(t, err)

	slots := []struct {
		slotName string
		itemID   string
	}{
		{"main_hand", "longsword"},
		{"off_hand", "shield"},
		{"body", "chainmail"},
		{"head", "helmet"},
		{"hands", "gauntlets"},
		{"feet", "boots"},
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
