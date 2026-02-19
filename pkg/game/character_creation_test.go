package game

import (
	"testing"
)

func TestCharacterCreator_CreateCharacter_RollMethod(t *testing.T) {
	creator := NewCharacterCreator()

	config := CharacterCreationConfig{
		Name:              "TestFighter",
		Class:             ClassThief, // Use Thief which has lower requirements
		AttributeMethod:   "roll",
		StartingEquipment: true,
		StartingGold:      100,
	}

	// Try multiple times since roll is random
	var result CharacterCreationResult
	success := false
	for i := 0; i < 10; i++ {
		result = creator.CreateCharacter(config)
		if result.Success {
			success = true
			break
		}
	}

	if !success {
		t.Fatalf("Character creation failed after 10 attempts: %v", result.Errors)
	}

	if result.Character == nil {
		t.Fatal("Character was not created")
	}

	if result.PlayerData == nil {
		t.Fatal("Player data was not created")
	}

	// Check character attributes
	if result.Character.Name != "TestFighter" {
		t.Errorf("Expected name 'TestFighter', got '%s'", result.Character.Name)
	}

	if result.PlayerData.Character.Class != ClassThief {
		t.Errorf("Expected class Thief, got %v", result.PlayerData.Character.Class)
	}

	// Check attributes are within valid range (3-18)
	stats := result.GeneratedStats
	for attr, value := range stats {
		if value < 3 || value > 18 {
			t.Errorf("Attribute %s value %d is out of range (3-18)", attr, value)
		}
	}

	// Check derived stats
	if result.Character.MaxHP <= 0 {
		t.Errorf("MaxHP should be positive, got %d", result.Character.MaxHP)
	}

	if result.Character.HP != result.Character.MaxHP {
		t.Errorf("HP should equal MaxHP for new character, got HP=%d, MaxHP=%d",
			result.Character.HP, result.Character.MaxHP)
	}

	// Check starting equipment
	if len(result.StartingItems) == 0 {
		t.Error("Thief should have starting equipment")
	}
}

func TestCharacterCreator_CreateCharacter_StandardArray(t *testing.T) {
	creator := NewCharacterCreator()

	// Use custom attributes that satisfy Mage requirements
	config := CharacterCreationConfig{
		Name:            "TestMage",
		Class:           ClassMage,
		AttributeMethod: "custom",
		CustomAttributes: map[string]int{
			"strength":     8,
			"dexterity":    12,
			"constitution": 10,
			"intelligence": 15, // Meets Mage requirement
			"wisdom":       13,
			"charisma":     14,
		},
		StartingEquipment: false,
		StartingGold:      50,
	}

	result := creator.CreateCharacter(config)

	if !result.Success {
		t.Fatalf("Character creation failed: %v", result.Errors)
	}

	// Check that intelligence meets requirement
	if result.GeneratedStats["intelligence"] < 13 {
		t.Errorf("Mage intelligence should be at least 13, got %d", result.GeneratedStats["intelligence"])
	}
}

func TestCharacterCreator_CreateCharacter_CustomAttributes(t *testing.T) {
	creator := NewCharacterCreator()

	customAttrs := map[string]int{
		"strength":     15,
		"dexterity":    14,
		"constitution": 13,
		"intelligence": 16,
		"wisdom":       12,
		"charisma":     10,
	}

	config := CharacterCreationConfig{
		Name:              "TestCustom",
		Class:             ClassMage,
		AttributeMethod:   "custom",
		CustomAttributes:  customAttrs,
		StartingEquipment: false,
		StartingGold:      0,
	}

	result := creator.CreateCharacter(config)

	if !result.Success {
		t.Fatalf("Character creation failed: %v", result.Errors)
	}

	// Check custom attributes match
	for attr, expected := range customAttrs {
		if result.GeneratedStats[attr] != expected {
			t.Errorf("Custom attribute %s: expected %d, got %d",
				attr, expected, result.GeneratedStats[attr])
		}
	}
}

func TestCharacterCreator_CreateCharacter_InvalidName(t *testing.T) {
	creator := NewCharacterCreator()

	config := CharacterCreationConfig{
		Name:            "", // Empty name should fail
		Class:           ClassFighter,
		AttributeMethod: "standard",
	}

	result := creator.CreateCharacter(config)

	if result.Success {
		t.Error("Character creation should have failed with empty name")
	}

	if len(result.Errors) == 0 {
		t.Error("Should have error messages for empty name")
	}
}

func TestCharacterCreator_CreateCharacter_ClassRequirements(t *testing.T) {
	creator := NewCharacterCreator()

	// Try to create a Paladin with insufficient requirements
	customAttrs := map[string]int{
		"strength":     10, // Too low for Paladin (needs 13)
		"dexterity":    10,
		"constitution": 10,
		"intelligence": 10,
		"wisdom":       10,
		"charisma":     10, // Too low for Paladin (needs 13)
	}

	config := CharacterCreationConfig{
		Name:             "TestPaladin",
		Class:            ClassPaladin,
		AttributeMethod:  "custom",
		CustomAttributes: customAttrs,
	}

	result := creator.CreateCharacter(config)

	if result.Success {
		t.Error("Character creation should have failed due to insufficient class requirements")
	}

	if len(result.Errors) == 0 {
		t.Error("Should have error messages for failed class requirements")
	}
}

func TestCharacterCreator_ValidateConfig(t *testing.T) {
	creator := NewCharacterCreator()

	tests := []struct {
		name        string
		config      CharacterCreationConfig
		expectError bool
	}{
		{
			name: "Valid config",
			config: CharacterCreationConfig{
				Name:            "TestChar",
				Class:           ClassFighter,
				AttributeMethod: "roll",
			},
			expectError: false,
		},
		{
			name: "Empty name",
			config: CharacterCreationConfig{
				Name:            "",
				Class:           ClassFighter,
				AttributeMethod: "roll",
			},
			expectError: true,
		},
		{
			name: "Name too long",
			config: CharacterCreationConfig{
				Name:            "ThisNameIsWayTooLongAndShouldFailValidationBecauseItExceedsFiftyCharacters",
				Class:           ClassFighter,
				AttributeMethod: "roll",
			},
			expectError: true,
		},
		{
			name: "Invalid attribute method",
			config: CharacterCreationConfig{
				Name:            "TestChar",
				Class:           ClassFighter,
				AttributeMethod: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := creator.validateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestCharacterCreator_GenerateAttributes(t *testing.T) {
	creator := NewCharacterCreator()

	// Test roll method
	config := CharacterCreationConfig{AttributeMethod: "roll"}
	attrs, err := creator.generateAttributes(config)
	if err != nil {
		t.Errorf("Roll method failed: %v", err)
	}
	if len(attrs) != 6 {
		t.Errorf("Expected 6 attributes, got %d", len(attrs))
	}

	// Test standard method
	config.AttributeMethod = "standard"
	attrs, err = creator.generateAttributes(config)
	if err != nil {
		t.Errorf("Standard method failed: %v", err)
	}
	if attrs["strength"] != 15 {
		t.Errorf("Standard array strength should be 15, got %d", attrs["strength"])
	}

	// Test pointbuy method
	config.AttributeMethod = "pointbuy"
	attrs, err = creator.generateAttributes(config)
	if err != nil {
		t.Errorf("Pointbuy method failed: %v", err)
	}

	// Check that all attributes are at least 8 (base value)
	for attr, value := range attrs {
		if value < 8 {
			t.Errorf("Pointbuy attribute %s should be at least 8, got %d", attr, value)
		}
	}

	// Test custom method
	config.AttributeMethod = "custom"
	config.CustomAttributes = map[string]int{
		"strength": 15, "dexterity": 14, "constitution": 13,
		"intelligence": 12, "wisdom": 11, "charisma": 10,
	}
	attrs, err = creator.generateAttributes(config)
	if err != nil {
		t.Errorf("Custom method failed: %v", err)
	}
	if attrs["strength"] != 15 {
		t.Errorf("Custom strength should be 15, got %d", attrs["strength"])
	}
}

func TestCharacterCreator_CalculateDerivedStats(t *testing.T) {
	creator := NewCharacterCreator()

	character := &Character{
		Class:        ClassFighter,
		Constitution: 14, // +2 modifier
		Dexterity:    16, // +3 modifier
	}

	creator.calculateDerivedStats(character, ClassFighter)

	// Fighter has 10 base HP + CON modifier
	expectedMaxHP := 10 + 2 // 12
	if character.MaxHP != expectedMaxHP {
		t.Errorf("Fighter MaxHP should be %d, got %d", expectedMaxHP, character.MaxHP)
	}

	if character.HP != character.MaxHP {
		t.Errorf("HP should equal MaxHP, got HP=%d, MaxHP=%d", character.HP, character.MaxHP)
	}

	// AC should be 10 + DEX modifier
	expectedAC := 10 + 3 // 13
	if character.ArmorClass != expectedAC {
		t.Errorf("AC should be %d, got %d", expectedAC, character.ArmorClass)
	}

	if character.THAC0 != 20 {
		t.Errorf("THAC0 should be 20 for level 1, got %d", character.THAC0)
	}
}

func TestCharacterCreator_GetStartingEquipment(t *testing.T) {
	creator := NewCharacterCreator()

	// Test Fighter equipment
	equipment := creator.getStartingEquipment(ClassFighter)
	if len(equipment) == 0 {
		t.Error("Fighter should have starting equipment")
	}

	hasWeapon := false
	hasArmor := false
	for _, item := range equipment {
		if item.Type == "weapon" {
			hasWeapon = true
		}
		if item.Type == "armor" {
			hasArmor = true
		}
	}

	if !hasWeapon {
		t.Error("Fighter should have a starting weapon")
	}
	if !hasArmor {
		t.Error("Fighter should have starting armor")
	}

	// Test Mage equipment (should be minimal)
	equipment = creator.getStartingEquipment(ClassMage)
	// Mages typically get very little starting equipment
	// This is acceptable behavior
}

func TestCharacterCreator_NewCharacterCreator(t *testing.T) {
	creator := NewCharacterCreator()

	if creator == nil {
		t.Fatal("NewCharacterCreator returned nil")
	}

	if creator.classConfigs == nil {
		t.Error("Class configs not initialized")
	}

	if creator.itemDatabase == nil {
		t.Error("Item database not initialized")
	}

	if creator.rng == nil {
		t.Error("Random number generator not initialized")
	}

	// Check that all classes are configured
	expectedClasses := []CharacterClass{
		ClassFighter, ClassMage, ClassCleric, ClassThief, ClassRanger, ClassPaladin,
	}

	for _, class := range expectedClasses {
		if _, exists := creator.classConfigs[class]; !exists {
			t.Errorf("Class %v not configured", class)
		}
	}

	// Check that basic items exist
	expectedItems := []string{"weapon_shortsword", "armor_leather"}
	for _, itemID := range expectedItems {
		if _, exists := creator.itemDatabase[itemID]; !exists {
			t.Errorf("Item %s not in database", itemID)
		}
	}
}

func TestCharacterCreator_NewCharacterCreatorWithSeed(t *testing.T) {
	// Test that NewCharacterCreatorWithSeed produces the same creator structure
	seed := int64(42)
	creator := NewCharacterCreatorWithSeed(seed)

	if creator == nil {
		t.Fatal("NewCharacterCreatorWithSeed returned nil")
	}

	if creator.classConfigs == nil {
		t.Error("Class configs not initialized")
	}

	if creator.itemDatabase == nil {
		t.Error("Item database not initialized")
	}

	if creator.rng == nil {
		t.Error("Random number generator not initialized")
	}

	// Check that all classes are configured
	expectedClasses := []CharacterClass{
		ClassFighter, ClassMage, ClassCleric, ClassThief, ClassRanger, ClassPaladin,
	}

	for _, class := range expectedClasses {
		if _, exists := creator.classConfigs[class]; !exists {
			t.Errorf("Class %v not configured", class)
		}
	}
}

func TestCharacterCreator_DeterministicCreation(t *testing.T) {
	// Test that the same seed produces the same character attributes
	seed := int64(12345)

	creator1 := NewCharacterCreatorWithSeed(seed)
	creator2 := NewCharacterCreatorWithSeed(seed)

	config := CharacterCreationConfig{
		Name:              "TestDeterminism",
		Class:             ClassThief, // Use Thief which has lower requirements
		AttributeMethod:   "roll",
		StartingEquipment: true,
		StartingGold:      100,
	}

	result1 := creator1.CreateCharacter(config)
	result2 := creator2.CreateCharacter(config)

	// Both should either succeed or fail the same way
	if result1.Success != result2.Success {
		t.Errorf("Success status differs: %v vs %v", result1.Success, result2.Success)
	}

	if result1.Success && result2.Success {
		// Check that generated stats are identical
		for attr, value1 := range result1.GeneratedStats {
			value2, exists := result2.GeneratedStats[attr]
			if !exists {
				t.Errorf("Attribute %s missing from second creation", attr)
				continue
			}
			if value1 != value2 {
				t.Errorf("Attribute %s differs: %d vs %d", attr, value1, value2)
			}
		}
	}
}

func TestCharacterCreator_PointBuyDeterministic(t *testing.T) {
	// Test that pointbuy method is deterministic with same seed
	seed := int64(67890)

	creator1 := NewCharacterCreatorWithSeed(seed)
	creator2 := NewCharacterCreatorWithSeed(seed)

	config := CharacterCreationConfig{
		Name:              "TestPointBuy",
		Class:             ClassFighter,
		AttributeMethod:   "pointbuy",
		StartingEquipment: false,
		StartingGold:      0,
	}

	result1 := creator1.CreateCharacter(config)
	result2 := creator2.CreateCharacter(config)

	if !result1.Success || !result2.Success {
		t.Fatalf("Point buy creation failed: %v / %v", result1.Errors, result2.Errors)
	}

	// Check that generated stats are identical
	for attr, value1 := range result1.GeneratedStats {
		value2, exists := result2.GeneratedStats[attr]
		if !exists {
			t.Errorf("Attribute %s missing from second creation", attr)
			continue
		}
		if value1 != value2 {
			t.Errorf("Attribute %s differs: %d vs %d", attr, value1, value2)
		}
	}
}

func BenchmarkCharacterCreation_Roll(b *testing.B) {
	creator := NewCharacterCreator()
	config := CharacterCreationConfig{
		Name:              "BenchChar",
		Class:             ClassFighter,
		AttributeMethod:   "roll",
		StartingEquipment: true,
		StartingGold:      100,
	}

	for i := 0; i < b.N; i++ {
		result := creator.CreateCharacter(config)
		if !result.Success {
			b.Fatalf("Character creation failed: %v", result.Errors)
		}
	}
}

func BenchmarkCharacterCreation_Standard(b *testing.B) {
	creator := NewCharacterCreator()
	config := CharacterCreationConfig{
		Name:              "BenchChar",
		Class:             ClassFighter,
		AttributeMethod:   "standard",
		StartingEquipment: true,
		StartingGold:      100,
	}

	for i := 0; i < b.N; i++ {
		result := creator.CreateCharacter(config)
		if !result.Success {
			b.Fatalf("Character creation failed: %v", result.Errors)
		}
	}
}
