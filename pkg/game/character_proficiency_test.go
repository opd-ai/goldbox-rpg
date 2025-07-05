package game

import (
	"testing"
)

// TestClassProficiencyValidation tests that equipment proficiency validation works correctly
func TestClassProficiencyValidation(t *testing.T) {
	tests := []struct {
		name           string
		characterClass CharacterClass
		item           Item
		slot           EquipmentSlot
		expected       bool
		description    string
	}{
		{
			name:           "Fighter can equip sword",
			characterClass: ClassFighter,
			item:           Item{ID: "sword001", Name: "Iron Sword", Type: "sword"},
			slot:           SlotWeaponMain,
			expected:       true,
			description:    "Fighters should be able to equip swords",
		},
		{
			name:           "Mage cannot equip sword",
			characterClass: ClassMage,
			item:           Item{ID: "sword001", Name: "Iron Sword", Type: "sword"},
			slot:           SlotWeaponMain,
			expected:       false,
			description:    "Mages should not be able to equip swords",
		},
		{
			name:           "Mage can equip staff",
			characterClass: ClassMage,
			item:           Item{ID: "staff001", Name: "Magic Staff", Type: "staff"},
			slot:           SlotWeaponMain,
			expected:       true,
			description:    "Mages should be able to equip staves",
		},
		{
			name:           "Fighter can equip heavy armor",
			characterClass: ClassFighter,
			item:           Item{ID: "plate001", Name: "Plate Armor", Type: "armor", Properties: []string{"heavy"}},
			slot:           SlotChest,
			expected:       true,
			description:    "Fighters should be able to equip heavy armor",
		},
		{
			name:           "Mage cannot equip armor",
			characterClass: ClassMage,
			item:           Item{ID: "leather001", Name: "Leather Armor", Type: "armor", Properties: []string{"light"}},
			slot:           SlotChest,
			expected:       false,
			description:    "Mages should not be able to equip any armor",
		},
		{
			name:           "Thief can equip light armor",
			characterClass: ClassThief,
			item:           Item{ID: "leather001", Name: "Leather Armor", Type: "armor", Properties: []string{"light"}},
			slot:           SlotChest,
			expected:       true,
			description:    "Thieves should be able to equip light armor",
		},
		{
			name:           "Thief cannot equip heavy armor",
			characterClass: ClassThief,
			item:           Item{ID: "plate001", Name: "Plate Armor", Type: "armor", Properties: []string{"heavy"}},
			slot:           SlotChest,
			expected:       false,
			description:    "Thieves should not be able to equip heavy armor",
		},
		{
			name:           "Mage cannot equip shield",
			characterClass: ClassMage,
			item:           Item{ID: "shield001", Name: "Iron Shield", Type: "shield"},
			slot:           SlotWeaponOff,
			expected:       false,
			description:    "Mages should not be able to equip shields",
		},
		{
			name:           "Fighter can equip shield",
			characterClass: ClassFighter,
			item:           Item{ID: "shield001", Name: "Iron Shield", Type: "shield"},
			slot:           SlotWeaponOff,
			expected:       true,
			description:    "Fighters should be able to equip shields",
		},
		{
			name:           "Cleric can equip mace",
			characterClass: ClassCleric,
			item:           Item{ID: "mace001", Name: "Iron Mace", Type: "mace"},
			slot:           SlotWeaponMain,
			expected:       true,
			description:    "Clerics should be able to equip maces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				ID:        "test-char",
				Name:      "Test Character",
				Class:     tt.characterClass,
				Equipment: make(map[EquipmentSlot]Item),
				Inventory: []Item{tt.item},
			}

			result := character.canEquipItemInSlot(tt.item, tt.slot)
			if result != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.expected, result)
			}
		})
	}
}

// TestGetClassProficiencies tests the GetClassProficiencies function
func TestGetClassProficiencies(t *testing.T) {
	tests := []struct {
		name      string
		class     CharacterClass
		checkFunc func(t *testing.T, prof ClassProficiencies)
	}{
		{
			name:  "Fighter proficiencies",
			class: ClassFighter,
			checkFunc: func(t *testing.T, prof ClassProficiencies) {
				if !prof.ShieldProficient {
					t.Error("Fighter should be shield proficient")
				}
				if len(prof.WeaponTypes) == 0 {
					t.Error("Fighter should have weapon proficiencies")
				}
				if len(prof.ArmorTypes) != 3 {
					t.Errorf("Fighter should have 3 armor types, got %d", len(prof.ArmorTypes))
				}
			},
		},
		{
			name:  "Mage proficiencies",
			class: ClassMage,
			checkFunc: func(t *testing.T, prof ClassProficiencies) {
				if prof.ShieldProficient {
					t.Error("Mage should not be shield proficient")
				}
				if len(prof.ArmorTypes) != 0 {
					t.Errorf("Mage should have no armor proficiencies, got %d", len(prof.ArmorTypes))
				}
				// Check mage has staff proficiency
				hasStaff := false
				for _, weapon := range prof.WeaponTypes {
					if weapon == "staff" {
						hasStaff = true
						break
					}
				}
				if !hasStaff {
					t.Error("Mage should be proficient with staff")
				}
			},
		},
		{
			name:  "Thief proficiencies",
			class: ClassThief,
			checkFunc: func(t *testing.T, prof ClassProficiencies) {
				if prof.ShieldProficient {
					t.Error("Thief should not be shield proficient")
				}
				// Check thief can only wear light armor
				if len(prof.ArmorTypes) != 1 || prof.ArmorTypes[0] != "light" {
					t.Errorf("Thief should only have light armor proficiency, got %v", prof.ArmorTypes)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proficiencies := GetClassProficiencies(tt.class)
			if proficiencies.Class != tt.class {
				t.Errorf("Expected class %v, got %v", tt.class, proficiencies.Class)
			}
			tt.checkFunc(t, proficiencies)
		})
	}
}

// TestDetermineArmorType tests the armor type detection helper function
func TestDetermineArmorType(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected string
	}{
		{
			name:     "Light armor with property",
			item:     Item{Type: "armor", Properties: []string{"light"}},
			expected: "light",
		},
		{
			name:     "Heavy armor with property",
			item:     Item{Type: "armor", Properties: []string{"heavy"}},
			expected: "heavy",
		},
		{
			name:     "Leather armor by name",
			item:     Item{Type: "armor", Name: "Leather Armor"},
			expected: "light",
		},
		{
			name:     "Plate armor by name",
			item:     Item{Type: "armor", Name: "Plate Mail"},
			expected: "heavy",
		},
		{
			name:     "Chain armor by name",
			item:     Item{Type: "armor", Name: "Chain Mail"},
			expected: "medium",
		},
		{
			name:     "Default armor type",
			item:     Item{Type: "armor", Name: "Unknown Armor"},
			expected: "light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineArmorType(tt.item)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
