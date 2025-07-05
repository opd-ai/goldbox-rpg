package game

import (
	"testing"
)

// TestCharacter_EquipItem tests the EquipItem functionality
func TestCharacter_EquipItem(t *testing.T) {
	character := &Character{
		ID:        "test-char-1",
		Name:      "Test Character",
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{
			{
				ID:         "sword001",
				Name:       "Iron Sword",
				Type:       "weapon",
				Properties: []string{"sharp"},
			},
			{
				ID:         "helmet001",
				Name:       "Iron Helmet",
				Type:       "helmet",
				Properties: []string{"protective"},
			},
		},
	}

	tests := []struct {
		name        string
		itemID      string
		slot        EquipmentSlot
		expectError bool
		errorMsg    string
	}{
		{
			name:        "equip valid weapon",
			itemID:      "sword001",
			slot:        SlotWeaponMain,
			expectError: false,
		},
		{
			name:        "equip valid helmet",
			itemID:      "helmet001",
			slot:        SlotHead,
			expectError: false,
		},
		{
			name:        "equip non-existent item",
			itemID:      "nonexistent",
			slot:        SlotWeaponMain,
			expectError: true,
			errorMsg:    "item not found in inventory",
		},
		{
			name:        "equip wrong item type",
			itemID:      "sword001",
			slot:        SlotHead,
			expectError: true,
			errorMsg:    "cannot be equipped in slot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset character state
			character.Equipment = make(map[EquipmentSlot]Item)
			character.Inventory = []Item{
				{ID: "sword001", Name: "Iron Sword", Type: "weapon", Properties: []string{"sharp"}},
				{ID: "helmet001", Name: "Iron Helmet", Type: "helmet", Properties: []string{"protective"}},
			}

			err := character.EquipItem(tt.itemID, tt.slot)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != "" {
					// Check if error message contains expected substring
					found := false
					for i := 0; i <= len(err.Error())-len(tt.errorMsg); i++ {
						if err.Error()[i:i+len(tt.errorMsg)] == tt.errorMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Error message %q does not contain %q", err.Error(), tt.errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify item was equipped
				if equipped, exists := character.Equipment[tt.slot]; !exists {
					t.Errorf("Item was not equipped in slot %s", tt.slot.String())
				} else if equipped.ID != tt.itemID {
					t.Errorf("Wrong item equipped: got %s, want %s", equipped.ID, tt.itemID)
				}

				// Verify item was removed from inventory
				for _, item := range character.Inventory {
					if item.ID == tt.itemID {
						t.Errorf("Item %s was not removed from inventory", tt.itemID)
					}
				}
			}
		})
	}
}

// TestCharacter_UnequipItem tests the UnequipItem functionality
func TestCharacter_UnequipItem(t *testing.T) {
	character := &Character{
		ID:   "test-char-1",
		Name: "Test Character",
		Equipment: map[EquipmentSlot]Item{
			SlotWeaponMain: {
				ID:   "sword001",
				Name: "Iron Sword",
				Type: "weapon",
			},
		},
		Inventory: []Item{},
	}

	// Test successful unequip
	item, err := character.UnequipItem(SlotWeaponMain)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if item == nil {
		t.Errorf("Expected item to be returned")
	} else if item.ID != "sword001" {
		t.Errorf("Wrong item returned: got %s, want sword001", item.ID)
	}

	// Verify item was removed from equipment
	if _, exists := character.Equipment[SlotWeaponMain]; exists {
		t.Errorf("Item was not removed from equipment slot")
	}

	// Verify item was added to inventory
	found := false
	for _, invItem := range character.Inventory {
		if invItem.ID == "sword001" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Item was not added to inventory")
	}

	// Test unequip empty slot
	_, err = character.UnequipItem(SlotWeaponOff)
	if err == nil {
		t.Errorf("Expected error when unequipping empty slot")
	}
}

// TestCharacter_CanEquipItem tests the CanEquipItem functionality
func TestCharacter_CanEquipItem(t *testing.T) {
	character := &Character{
		ID:        "test-char-1",
		Name:      "Test Character",
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{
			{ID: "sword001", Name: "Iron Sword", Type: "weapon"},
			{ID: "helmet001", Name: "Iron Helmet", Type: "helmet"},
		},
	}

	tests := []struct {
		name     string
		itemID   string
		slot     EquipmentSlot
		expected bool
	}{
		{"weapon in weapon slot", "sword001", SlotWeaponMain, true},
		{"helmet in head slot", "helmet001", SlotHead, true},
		{"weapon in head slot", "sword001", SlotHead, false},
		{"helmet in weapon slot", "helmet001", SlotWeaponMain, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canEquip, err := character.CanEquipItem(tt.itemID, tt.slot)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if canEquip != tt.expected {
				t.Errorf("CanEquipItem() = %v, want %v", canEquip, tt.expected)
			}
		})
	}

	// Test with non-existent item
	_, err := character.CanEquipItem("nonexistent", SlotWeaponMain)
	if err == nil {
		t.Errorf("Expected error for non-existent item")
	}
}

// TestCharacter_AddItemToInventory tests the AddItemToInventory functionality
func TestCharacter_AddItemToInventory(t *testing.T) {
	character := &Character{
		ID:        "test-char-1",
		Name:      "Test Character",
		Strength:  15, // Should allow decent carrying capacity
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{},
	}

	// Test adding valid item
	item := Item{
		ID:     "potion001",
		Name:   "Health Potion",
		Type:   "potion",
		Weight: 1,
	}

	err := character.AddItemToInventory(item)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify item was added
	if len(character.Inventory) != 1 {
		t.Errorf("Expected 1 item in inventory, got %d", len(character.Inventory))
	}

	if character.Inventory[0].ID != "potion001" {
		t.Errorf("Wrong item added: got %s, want potion001", character.Inventory[0].ID)
	}

	// Test adding item with empty ID
	invalidItem := Item{
		ID:   "",
		Name: "Invalid Item",
	}

	err = character.AddItemToInventory(invalidItem)
	if err == nil {
		t.Errorf("Expected error for item with empty ID")
	}

	// Test weight limit
	heavyItem := Item{
		ID:     "anvil001",
		Name:   "Heavy Anvil",
		Type:   "misc",
		Weight: 1000, // Should exceed capacity
	}

	err = character.AddItemToInventory(heavyItem)
	if err == nil {
		t.Errorf("Expected error for item exceeding weight capacity")
	}
}

// TestCharacter_RemoveItemFromInventory tests the RemoveItemFromInventory functionality
func TestCharacter_RemoveItemFromInventory(t *testing.T) {
	character := &Character{
		ID:        "test-char-1",
		Name:      "Test Character",
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{
			{ID: "potion001", Name: "Health Potion", Type: "potion"},
			{ID: "sword001", Name: "Iron Sword", Type: "weapon"},
		},
	}

	// Test removing existing item
	item, err := character.RemoveItemFromInventory("potion001")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if item == nil {
		t.Errorf("Expected item to be returned")
	} else if item.ID != "potion001" {
		t.Errorf("Wrong item returned: got %s, want potion001", item.ID)
	}

	// Verify item was removed
	if len(character.Inventory) != 1 {
		t.Errorf("Expected 1 item remaining, got %d", len(character.Inventory))
	}

	// Test removing non-existent item
	_, err = character.RemoveItemFromInventory("nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent item")
	}
}

// TestCharacter_TransferItemTo tests the TransferItemTo functionality
func TestCharacter_TransferItemTo(t *testing.T) {
	sourceChar := &Character{
		ID:        "source-char",
		Name:      "Source Character",
		Strength:  15,
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{
			{ID: "potion001", Name: "Health Potion", Type: "potion", Weight: 1},
		},
	}

	targetChar := &Character{
		ID:        "target-char",
		Name:      "Target Character",
		Strength:  15,
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{},
	}

	// Test successful transfer
	err := sourceChar.TransferItemTo("potion001", targetChar)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify item was removed from source
	if len(sourceChar.Inventory) != 0 {
		t.Errorf("Expected source inventory to be empty, got %d items", len(sourceChar.Inventory))
	}

	// Verify item was added to target
	if len(targetChar.Inventory) != 1 {
		t.Errorf("Expected target inventory to have 1 item, got %d", len(targetChar.Inventory))
	}

	if targetChar.Inventory[0].ID != "potion001" {
		t.Errorf("Wrong item transferred: got %s, want potion001", targetChar.Inventory[0].ID)
	}

	// Test transfer non-existent item
	err = sourceChar.TransferItemTo("nonexistent", targetChar)
	if err == nil {
		t.Errorf("Expected error for non-existent item")
	}
}

// TestCharacter_GetInventoryWeight tests weight calculation
func TestCharacter_GetInventoryWeight(t *testing.T) {
	character := &Character{
		ID:   "test-char-1",
		Name: "Test Character",
		Equipment: map[EquipmentSlot]Item{
			SlotWeaponMain: {ID: "sword001", Name: "Iron Sword", Type: "weapon", Weight: 5},
		},
		Inventory: []Item{
			{ID: "potion001", Name: "Health Potion", Type: "potion", Weight: 1},
			{ID: "armor001", Name: "Leather Armor", Type: "armor", Weight: 10},
		},
	}

	weight := character.GetInventoryWeight()
	expectedWeight := 16 // 1 + 10 + 5 (equipped sword)
	if weight != expectedWeight {
		t.Errorf("GetInventoryWeight() = %d, want %d", weight, expectedWeight)
	}
}

// TestCharacter_CalculateEquipmentBonuses tests equipment stat bonus calculation
func TestCharacter_CalculateEquipmentBonuses(t *testing.T) {
	character := &Character{
		ID:   "test-char-1",
		Name: "Test Character",
		Equipment: map[EquipmentSlot]Item{
			SlotWeaponMain: {
				ID:         "magic_sword",
				Name:       "Magic Sword",
				Type:       "weapon",
				Properties: []string{"strength+2", "sharp"},
			},
			SlotChest: {
				ID:   "leather_armor",
				Name: "Leather Armor",
				Type: "armor",
				AC:   12,
			},
		},
		Inventory: []Item{},
	}

	bonuses := character.CalculateEquipmentBonuses()

	if bonuses["strength"] != 2 {
		t.Errorf("Expected strength bonus of 2, got %d", bonuses["strength"])
	}

	if bonuses["armor_class"] != 2 { // 12 - 10 (base AC)
		t.Errorf("Expected armor_class bonus of 2, got %d", bonuses["armor_class"])
	}
}

// TestCharacter_CalculateEquipmentBonuses_MultiDigit tests equipment parsing with multi-digit modifiers
func TestCharacter_CalculateEquipmentBonuses_MultiDigit(t *testing.T) {
	character := &Character{
		ID:   "test-char-1",
		Name: "Test Character",
		Equipment: map[EquipmentSlot]Item{
			SlotWeaponMain: {
				ID:         "cursed_sword",
				Name:       "Cursed Sword",
				Type:       "weapon",
				Properties: []string{"strength-10", "dexterity+15"},
			},
		},
		Inventory: []Item{},
	}

	bonuses := character.CalculateEquipmentBonuses()

	if bonuses["strength"] != -10 {
		t.Errorf("Expected strength bonus of -10, got %d", bonuses["strength"])
	}

	if bonuses["dexterity"] != 15 {
		t.Errorf("Expected dexterity bonus of 15, got %d", bonuses["dexterity"])
	}
}

// TestCharacter_HasItem tests the HasItem functionality
func TestCharacter_HasItem(t *testing.T) {
	character := &Character{
		ID:        "test-char-1",
		Name:      "Test Character",
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{
			{ID: "potion001", Name: "Health Potion", Type: "potion"},
		},
	}

	if !character.HasItem("potion001") {
		t.Errorf("HasItem() = false, want true for existing item")
	}

	if character.HasItem("nonexistent") {
		t.Errorf("HasItem() = true, want false for non-existent item")
	}
}

// TestCharacter_CountItems tests the CountItems functionality
func TestCharacter_CountItems(t *testing.T) {
	character := &Character{
		ID:        "test-char-1",
		Name:      "Test Character",
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{
			{ID: "potion001", Name: "Health Potion", Type: "potion"},
			{ID: "potion002", Name: "Mana Potion", Type: "potion"},
			{ID: "sword001", Name: "Iron Sword", Type: "weapon"},
		},
	}

	potionCount := character.CountItems("potion")
	if potionCount != 2 {
		t.Errorf("CountItems('potion') = %d, want 2", potionCount)
	}

	weaponCount := character.CountItems("weapon")
	if weaponCount != 1 {
		t.Errorf("CountItems('weapon') = %d, want 1", weaponCount)
	}

	armorCount := character.CountItems("armor")
	if armorCount != 0 {
		t.Errorf("CountItems('armor') = %d, want 0", armorCount)
	}
}
