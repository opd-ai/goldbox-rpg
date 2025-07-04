package game

import (
	"testing"
)

// TestCharacterClass_String tests the String() method for all character classes
func TestCharacterClass_String(t *testing.T) {
	tests := []struct {
		name     string
		class    CharacterClass
		expected string
	}{
		{
			name:     "Fighter class returns correct string",
			class:    ClassFighter,
			expected: "Fighter",
		},
		{
			name:     "Mage class returns correct string",
			class:    ClassMage,
			expected: "Mage",
		},
		{
			name:     "Cleric class returns correct string",
			class:    ClassCleric,
			expected: "Cleric",
		},
		{
			name:     "Thief class returns correct string",
			class:    ClassThief,
			expected: "Thief",
		},
		{
			name:     "Ranger class returns correct string",
			class:    ClassRanger,
			expected: "Ranger",
		},
		{
			name:     "Paladin class returns correct string",
			class:    ClassPaladin,
			expected: "Paladin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.class.String()
			if result != tt.expected {
				t.Errorf("CharacterClass.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCharacterClass_Constants tests that all class constants have the expected values
func TestCharacterClass_Constants(t *testing.T) {
	tests := []struct {
		name     string
		class    CharacterClass
		expected int
	}{
		{
			name:     "ClassFighter has value 0",
			class:    ClassFighter,
			expected: 0,
		},
		{
			name:     "ClassMage has value 1",
			class:    ClassMage,
			expected: 1,
		},
		{
			name:     "ClassCleric has value 2",
			class:    ClassCleric,
			expected: 2,
		},
		{
			name:     "ClassThief has value 3",
			class:    ClassThief,
			expected: 3,
		},
		{
			name:     "ClassRanger has value 4",
			class:    ClassRanger,
			expected: 4,
		},
		{
			name:     "ClassPaladin has value 5",
			class:    ClassPaladin,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := int(tt.class)
			if result != tt.expected {
				t.Errorf("CharacterClass constant = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCharacterClass_StringConsistency verifies String() method returns consistent results
func TestCharacterClass_StringConsistency(t *testing.T) {
	// Test that calling String() multiple times returns the same result
	class := ClassMage
	first := class.String()
	second := class.String()
	
	if first != second {
		t.Errorf("CharacterClass.String() inconsistent: first=%v, second=%v", first, second)
	}
}

// TestClassConfig_StructFields tests that ClassConfig struct can be instantiated properly
func TestClassConfig_StructFields(t *testing.T) {
	config := ClassConfig{
		Type:        ClassFighter,
		Name:        "Fighter",
		Description: "A warrior trained in combat",
		HitDice:     "1d10",
		BaseSkills:  []string{"Sword", "Shield"},
		Abilities:   []string{"Power Attack", "Cleave"},
	}
	
	// Test basic field assignment
	if config.Type != ClassFighter {
		t.Errorf("ClassConfig.Type = %v, want %v", config.Type, ClassFighter)
	}
	
	if config.Name != "Fighter" {
		t.Errorf("ClassConfig.Name = %v, want %v", config.Name, "Fighter")
	}
	
	if config.Description != "A warrior trained in combat" {
		t.Errorf("ClassConfig.Description = %v, want %v", config.Description, "A warrior trained in combat")
	}
	
	if config.HitDice != "1d10" {
		t.Errorf("ClassConfig.HitDice = %v, want %v", config.HitDice, "1d10")
	}
	
	if len(config.BaseSkills) != 2 {
		t.Errorf("ClassConfig.BaseSkills length = %v, want %v", len(config.BaseSkills), 2)
	}
	
	if len(config.Abilities) != 2 {
		t.Errorf("ClassConfig.Abilities length = %v, want %v", len(config.Abilities), 2)
	}
}

// TestClassConfig_Requirements tests the nested Requirements struct
func TestClassConfig_Requirements(t *testing.T) {
	config := ClassConfig{
		Type: ClassPaladin,
		Name: "Paladin",
	}
	
	// Set requirements
	config.Requirements.MinStr = 13
	config.Requirements.MinDex = 10
	config.Requirements.MinCon = 12
	config.Requirements.MinInt = 9
	config.Requirements.MinWis = 13
	config.Requirements.MinCha = 17
	
	tests := []struct {
		name     string
		actual   int
		expected int
	}{
		{"MinStr requirement", config.Requirements.MinStr, 13},
		{"MinDex requirement", config.Requirements.MinDex, 10},
		{"MinCon requirement", config.Requirements.MinCon, 12},
		{"MinInt requirement", config.Requirements.MinInt, 9},
		{"MinWis requirement", config.Requirements.MinWis, 13},
		{"MinCha requirement", config.Requirements.MinCha, 17},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

// TestClassProficiencies_StructFields tests ClassProficiencies struct initialization
func TestClassProficiencies_StructFields(t *testing.T) {
	proficiencies := ClassProficiencies{
		Class:            ClassRanger,
		WeaponTypes:      []string{"bow", "sword", "dagger"},
		ArmorTypes:       []string{"light", "medium"},
		ShieldProficient: true,
		Restrictions:     []string{"no heavy armor"},
	}
	
	if proficiencies.Class != ClassRanger {
		t.Errorf("ClassProficiencies.Class = %v, want %v", proficiencies.Class, ClassRanger)
	}
	
	if len(proficiencies.WeaponTypes) != 3 {
		t.Errorf("ClassProficiencies.WeaponTypes length = %v, want %v", len(proficiencies.WeaponTypes), 3)
	}
	
	if proficiencies.WeaponTypes[0] != "bow" {
		t.Errorf("ClassProficiencies.WeaponTypes[0] = %v, want %v", proficiencies.WeaponTypes[0], "bow")
	}
	
	if len(proficiencies.ArmorTypes) != 2 {
		t.Errorf("ClassProficiencies.ArmorTypes length = %v, want %v", len(proficiencies.ArmorTypes), 2)
	}
	
	if !proficiencies.ShieldProficient {
		t.Errorf("ClassProficiencies.ShieldProficient = %v, want %v", proficiencies.ShieldProficient, true)
	}
	
	if len(proficiencies.Restrictions) != 1 {
		t.Errorf("ClassProficiencies.Restrictions length = %v, want %v", len(proficiencies.Restrictions), 1)
	}
}

// TestClassConfig_EmptyConfiguration tests ClassConfig with default/empty values
func TestClassConfig_EmptyConfiguration(t *testing.T) {
	var config ClassConfig
	
	// Test default values
	if config.Type != ClassFighter { // 0 value should be ClassFighter
		t.Errorf("Empty ClassConfig.Type = %v, want %v", config.Type, ClassFighter)
	}
	
	if config.Name != "" {
		t.Errorf("Empty ClassConfig.Name = %v, want empty string", config.Name)
	}
	
	if len(config.BaseSkills) != 0 {
		t.Errorf("Empty ClassConfig.BaseSkills length = %v, want 0", len(config.BaseSkills))
	}
	
	if len(config.Abilities) != 0 {
		t.Errorf("Empty ClassConfig.Abilities length = %v, want 0", len(config.Abilities))
	}
	
	// Test that Requirements struct is also initialized to zero values
	if config.Requirements.MinStr != 0 {
		t.Errorf("Empty ClassConfig.Requirements.MinStr = %v, want 0", config.Requirements.MinStr)
	}
}

// TestClassProficiencies_EmptyConfiguration tests ClassProficiencies with default values
func TestClassProficiencies_EmptyConfiguration(t *testing.T) {
	var proficiencies ClassProficiencies
	
	// Test default values
	if proficiencies.Class != ClassFighter { // 0 value should be ClassFighter
		t.Errorf("Empty ClassProficiencies.Class = %v, want %v", proficiencies.Class, ClassFighter)
	}
	
	if len(proficiencies.WeaponTypes) != 0 {
		t.Errorf("Empty ClassProficiencies.WeaponTypes length = %v, want 0", len(proficiencies.WeaponTypes))
	}
	
	if len(proficiencies.ArmorTypes) != 0 {
		t.Errorf("Empty ClassProficiencies.ArmorTypes length = %v, want 0", len(proficiencies.ArmorTypes))
	}
	
	if proficiencies.ShieldProficient != false {
		t.Errorf("Empty ClassProficiencies.ShieldProficient = %v, want false", proficiencies.ShieldProficient)
	}
	
	if len(proficiencies.Restrictions) != 0 {
		t.Errorf("Empty ClassProficiencies.Restrictions length = %v, want 0", len(proficiencies.Restrictions))
	}
}
