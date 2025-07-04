package game

import (
	"testing"
)

// TestEquipmentSlot_String tests the String method for all valid EquipmentSlot values
func TestEquipmentSlot_String(t *testing.T) {
	tests := []struct {
		name     string
		slot     EquipmentSlot
		expected string
	}{
		{"Head slot", SlotHead, "Head"},
		{"Neck slot", SlotNeck, "Neck"},
		{"Chest slot", SlotChest, "Chest"},
		{"Hands slot", SlotHands, "Hands"},
		{"Rings slot", SlotRings, "Rings"},
		{"Legs slot", SlotLegs, "Legs"},
		{"Feet slot", SlotFeet, "Feet"},
		{"Main weapon slot", SlotWeaponMain, "MainHand"},
		{"Off weapon slot", SlotWeaponOff, "OffHand"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.slot.String()
			if result != tt.expected {
				t.Errorf("EquipmentSlot.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestEquipmentSlot_String_AllSlots tests that all enum values have string representations
func TestEquipmentSlot_String_AllSlots(t *testing.T) {
	slots := []EquipmentSlot{
		SlotHead, SlotNeck, SlotChest, SlotHands, SlotRings,
		SlotLegs, SlotFeet, SlotWeaponMain, SlotWeaponOff,
	}

	for _, slot := range slots {
		result := slot.String()
		if result == "" {
			t.Errorf("EquipmentSlot(%d).String() returned empty string", int(slot))
		}
	}
}

// TestEquipmentSlot_Constants tests that all slot constants have expected values
func TestEquipmentSlot_Constants(t *testing.T) {
	tests := []struct {
		name     string
		slot     EquipmentSlot
		expected int
	}{
		{"SlotHead should be 0", SlotHead, 0},
		{"SlotNeck should be 1", SlotNeck, 1},
		{"SlotChest should be 2", SlotChest, 2},
		{"SlotHands should be 3", SlotHands, 3},
		{"SlotRings should be 4", SlotRings, 4},
		{"SlotLegs should be 5", SlotLegs, 5},
		{"SlotFeet should be 6", SlotFeet, 6},
		{"SlotWeaponMain should be 7", SlotWeaponMain, 7},
		{"SlotWeaponOff should be 8", SlotWeaponOff, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.slot) != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, int(tt.slot), tt.expected)
			}
		})
	}
}

// TestEquipmentSlotConfig_StructFields tests EquipmentSlotConfig struct instantiation
func TestEquipmentSlotConfig_StructFields(t *testing.T) {
	config := EquipmentSlotConfig{
		Slot:         SlotChest,
		Name:         "Chest Armor",
		Description:  "Protective gear worn on the torso",
		AllowedTypes: []string{"armor", "robe"},
		Restricted:   false,
	}

	// Test field assignment
	if config.Slot != SlotChest {
		t.Errorf("Slot = %v, want %v", config.Slot, SlotChest)
	}
	if config.Name != "Chest Armor" {
		t.Errorf("Name = %q, want %q", config.Name, "Chest Armor")
	}
	if config.Description != "Protective gear worn on the torso" {
		t.Errorf("Description = %q, want %q", config.Description, "Protective gear worn on the torso")
	}
	if len(config.AllowedTypes) != 2 {
		t.Errorf("AllowedTypes length = %d, want 2", len(config.AllowedTypes))
	}
	if config.AllowedTypes[0] != "armor" || config.AllowedTypes[1] != "robe" {
		t.Errorf("AllowedTypes = %v, want [armor robe]", config.AllowedTypes)
	}
	if config.Restricted != false {
		t.Errorf("Restricted = %v, want false", config.Restricted)
	}
}

// TestEquipmentSlotConfig_EmptyValues tests EquipmentSlotConfig with default/empty values
func TestEquipmentSlotConfig_EmptyValues(t *testing.T) {
	config := EquipmentSlotConfig{}

	// Test zero values
	if config.Slot != SlotHead { // SlotHead is 0, the zero value
		t.Errorf("Default Slot = %v, want %v", config.Slot, SlotHead)
	}
	if config.Name != "" {
		t.Errorf("Default Name = %q, want empty string", config.Name)
	}
	if config.Description != "" {
		t.Errorf("Default Description = %q, want empty string", config.Description)
	}
	if config.AllowedTypes != nil {
		t.Errorf("Default AllowedTypes = %v, want nil", config.AllowedTypes)
	}
	if config.Restricted != false {
		t.Errorf("Default Restricted = %v, want false", config.Restricted)
	}
}

// TestEquipmentSet_StructFields tests EquipmentSet struct instantiation
func TestEquipmentSet_StructFields(t *testing.T) {
	slotConfig := EquipmentSlotConfig{
		Slot:         SlotWeaponMain,
		Name:         "Main Hand",
		Description:  "Primary weapon slot",
		AllowedTypes: []string{"sword", "axe", "staff"},
		Restricted:   true,
	}

	equipmentSet := EquipmentSet{
		CharacterID: "char123",
		Slots: map[EquipmentSlot]EquipmentSlotConfig{
			SlotWeaponMain: slotConfig,
		},
	}

	// Test field assignment
	if equipmentSet.CharacterID != "char123" {
		t.Errorf("CharacterID = %q, want %q", equipmentSet.CharacterID, "char123")
	}
	if len(equipmentSet.Slots) != 1 {
		t.Errorf("Slots length = %d, want 1", len(equipmentSet.Slots))
	}

	// Test slot configuration
	mainHandSlot, exists := equipmentSet.Slots[SlotWeaponMain]
	if !exists {
		t.Error("SlotWeaponMain not found in Slots map")
	}
	if mainHandSlot.Name != "Main Hand" {
		t.Errorf("Slot Name = %q, want %q", mainHandSlot.Name, "Main Hand")
	}
	if mainHandSlot.Restricted != true {
		t.Errorf("Slot Restricted = %v, want true", mainHandSlot.Restricted)
	}
}

// TestEquipmentSet_EmptyValues tests EquipmentSet with default/empty values
func TestEquipmentSet_EmptyValues(t *testing.T) {
	equipmentSet := EquipmentSet{}

	// Test zero values
	if equipmentSet.CharacterID != "" {
		t.Errorf("Default CharacterID = %q, want empty string", equipmentSet.CharacterID)
	}
	if equipmentSet.Slots != nil {
		t.Errorf("Default Slots = %v, want nil", equipmentSet.Slots)
	}
}

// TestEquipmentSet_MultipleSlots tests EquipmentSet with multiple equipment slots
func TestEquipmentSet_MultipleSlots(t *testing.T) {
	equipmentSet := EquipmentSet{
		CharacterID: "warrior001",
		Slots: map[EquipmentSlot]EquipmentSlotConfig{
			SlotHead: {
				Slot:         SlotHead,
				Name:         "Helmet",
				AllowedTypes: []string{"helmet", "hat"},
				Restricted:   false,
			},
			SlotChest: {
				Slot:         SlotChest,
				Name:         "Chest Armor",
				AllowedTypes: []string{"armor", "shirt"},
				Restricted:   false,
			},
			SlotWeaponMain: {
				Slot:         SlotWeaponMain,
				Name:         "Main Weapon",
				AllowedTypes: []string{"sword", "axe"},
				Restricted:   true,
			},
		},
	}

	// Test that all slots are present
	expectedSlots := []EquipmentSlot{SlotHead, SlotChest, SlotWeaponMain}
	for _, slot := range expectedSlots {
		config, exists := equipmentSet.Slots[slot]
		if !exists {
			t.Errorf("Slot %v not found in equipment set", slot)
			continue
		}
		if config.Slot != slot {
			t.Errorf("Slot %v has incorrect Slot field: got %v, want %v", slot, config.Slot, slot)
		}
	}

	// Test specific slot configurations
	if len(equipmentSet.Slots) != 3 {
		t.Errorf("Equipment set has %d slots, want 3", len(equipmentSet.Slots))
	}

	mainWeapon := equipmentSet.Slots[SlotWeaponMain]
	if !mainWeapon.Restricted {
		t.Error("Main weapon slot should be restricted")
	}

	helmet := equipmentSet.Slots[SlotHead]
	if helmet.Restricted {
		t.Error("Head slot should not be restricted")
	}
}
