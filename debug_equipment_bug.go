package main

import (
	"fmt"
	"goldbox-rpg/pkg/game"
)

func main() {
	// Test the exact scenario from the failing test
	character := &game.Character{
		ID:        "test-char-1",
		Name:      "Test Character",
		Class:     game.ClassFighter, // Default class should be fighter
		Equipment: make(map[game.EquipmentSlot]game.Item),
		Inventory: []game.Item{
			{
				ID:         "sword001",
				Name:       "Iron Sword",
				Type:       "weapon", // This is the issue!
				Properties: []string{"sharp"},
			},
		},
	}

	fmt.Println("=== Debugging Equipment Test Failure ===")

	// Check proficiencies for Fighter
	prof := game.GetClassProficiencies(game.ClassFighter)
	fmt.Printf("Fighter weapon proficiencies: %v\n", prof.WeaponTypes)

	// Try to equip the weapon
	err := character.EquipItem("sword001", game.SlotWeaponMain)
	if err != nil {
		fmt.Printf("❌ Failed to equip: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully equipped weapon\n")
	}

	// Test with correct weapon type
	character.Inventory = []game.Item{
		{
			ID:         "sword001",
			Name:       "Iron Sword",
			Type:       "sword", // Correct specific type
			Properties: []string{"sharp"},
		},
	}

	err = character.EquipItem("sword001", game.SlotWeaponMain)
	if err != nil {
		fmt.Printf("❌ Failed to equip sword: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully equipped sword\n")
	}

	// Test shield issue
	character.Inventory = []game.Item{
		{
			ID:   "shield001",
			Name: "Iron Shield",
			Type: "shield",
		},
	}

	err = character.EquipItem("shield001", game.SlotWeaponOff)
	if err != nil {
		fmt.Printf("❌ Failed to equip shield: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully equipped shield\n")
	}

	// Test cleric with mace
	cleric := &game.Character{
		ID:        "test-cleric",
		Name:      "Test Cleric",
		Class:     game.ClassCleric,
		Equipment: make(map[game.EquipmentSlot]game.Item),
		Inventory: []game.Item{
			{
				ID:   "mace001",
				Name: "Iron Mace",
				Type: "mace",
			},
		},
	}

	err = cleric.EquipItem("mace001", game.SlotWeaponMain)
	if err != nil {
		fmt.Printf("❌ Cleric failed to equip mace: %v\n", err)
	} else {
		fmt.Printf("✅ Cleric successfully equipped mace\n")
	}
}
