package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpellManager_LoadSpells(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "spell_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test spell file
	testSpell := `spells:
  - spell_id: "test_spell"
    spell_name: "Test Spell"
    spell_level: 1
    spell_school: 4
    spell_range: 30
    spell_duration: 5
    spell_components: [0, 1]
    spell_description: "A test spell"
    damage_type: "fire"
    damage_dice: "1d6"
    healing_dice: ""
    area_effect: false
    save_type: ""
    effect_keywords: ["test"]`

	testFile := filepath.Join(tempDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte(testSpell), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test loading
	manager := NewSpellManager(tempDir)
	if err := manager.LoadSpells(); err != nil {
		t.Fatalf("Failed to load spells: %v", err)
	}

	// Verify spell was loaded
	spell, err := manager.GetSpell("test_spell")
	if err != nil {
		t.Fatalf("Failed to get spell: %v", err)
	}

	if spell.Name != "Test Spell" {
		t.Errorf("Expected name 'Test Spell', got %s", spell.Name)
	}
	if spell.Level != 1 {
		t.Errorf("Expected level 1, got %d", spell.Level)
	}
	if spell.School != SchoolEvocation {
		t.Errorf("Expected school %d, got %d", SchoolEvocation, spell.School)
	}
}

func TestSpellManager_SaveSpell(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "spell_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewSpellManager(tempDir)

	// Create test spell
	spell := &Spell{
		ID:             "save_test",
		Name:           "Save Test",
		Level:          2,
		School:         SchoolIllusion,
		Range:          60,
		Duration:       10,
		Components:     []SpellComponent{ComponentVerbal, ComponentSomatic},
		Description:    "A spell for testing save functionality",
		DamageType:     "",
		DamageDice:     "",
		HealingDice:    "",
		AreaEffect:     false,
		SaveType:       "",
		EffectKeywords: []string{"test", "save"},
	}

	// Save spell
	if err := manager.SaveSpell(spell, "save_test.yaml"); err != nil {
		t.Fatalf("Failed to save spell: %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "save_test.yaml")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Spell file was not created")
	}

	// Verify spell can be retrieved
	retrievedSpell, err := manager.GetSpell("save_test")
	if err != nil {
		t.Fatalf("Failed to get saved spell: %v", err)
	}

	if retrievedSpell.Name != spell.Name {
		t.Errorf("Expected name %s, got %s", spell.Name, retrievedSpell.Name)
	}
}

func TestSpellManager_GetSpellsByLevel(t *testing.T) {
	manager := NewSpellManager("")

	// Add test spells
	spells := []*Spell{
		{ID: "level0_1", Name: "Cantrip One", Level: 0, School: SchoolEvocation},
		{ID: "level0_2", Name: "Cantrip Two", Level: 0, School: SchoolIllusion},
		{ID: "level1_1", Name: "First Level One", Level: 1, School: SchoolEvocation},
		{ID: "level2_1", Name: "Second Level One", Level: 2, School: SchoolEnchantment},
	}

	for _, spell := range spells {
		if err := manager.AddSpell(spell); err != nil {
			t.Fatalf("Failed to add spell: %v", err)
		}
	}

	// Test getting cantrips
	cantrips := manager.GetSpellsByLevel(0)
	if len(cantrips) != 2 {
		t.Errorf("Expected 2 cantrips, got %d", len(cantrips))
	}

	// Test getting level 1 spells
	level1 := manager.GetSpellsByLevel(1)
	if len(level1) != 1 {
		t.Errorf("Expected 1 level 1 spell, got %d", len(level1))
	}

	// Test getting non-existent level
	level5 := manager.GetSpellsByLevel(5)
	if len(level5) != 0 {
		t.Errorf("Expected 0 level 5 spells, got %d", len(level5))
	}
}

func TestSpellManager_GetSpellsBySchool(t *testing.T) {
	manager := NewSpellManager("")

	// Add test spells
	spells := []*Spell{
		{ID: "evocation1", Name: "Evocation One", Level: 1, School: SchoolEvocation},
		{ID: "evocation2", Name: "Evocation Two", Level: 2, School: SchoolEvocation},
		{ID: "illusion1", Name: "Illusion One", Level: 1, School: SchoolIllusion},
	}

	for _, spell := range spells {
		if err := manager.AddSpell(spell); err != nil {
			t.Fatalf("Failed to add spell: %v", err)
		}
	}

	// Test getting evocation spells
	evocationSpells := manager.GetSpellsBySchool(SchoolEvocation)
	if len(evocationSpells) != 2 {
		t.Errorf("Expected 2 evocation spells, got %d", len(evocationSpells))
	}

	// Test getting illusion spells
	illusionSpells := manager.GetSpellsBySchool(SchoolIllusion)
	if len(illusionSpells) != 1 {
		t.Errorf("Expected 1 illusion spell, got %d", len(illusionSpells))
	}

	// Test getting spells from empty school
	necromancySpells := manager.GetSpellsBySchool(SchoolNecromancy)
	if len(necromancySpells) != 0 {
		t.Errorf("Expected 0 necromancy spells, got %d", len(necromancySpells))
	}
}

func TestSpellManager_SearchSpells(t *testing.T) {
	manager := NewSpellManager("")

	// Add test spells
	spells := []*Spell{
		{
			ID:             "fire_spell",
			Name:           "Fireball",
			Description:    "A spell that creates a ball of fire",
			EffectKeywords: []string{"fire", "damage"},
		},
		{
			ID:             "heal_spell",
			Name:           "Cure Wounds",
			Description:    "A healing spell",
			EffectKeywords: []string{"healing", "restoration"},
		},
		{
			ID:             "ice_spell",
			Name:           "Ice Shard",
			Description:    "Creates sharp projectiles of ice",
			EffectKeywords: []string{"ice", "damage", "cold"},
		},
	}

	for _, spell := range spells {
		if err := manager.AddSpell(spell); err != nil {
			t.Fatalf("Failed to add spell: %v", err)
		}
	}

	// Test searching by name
	results := manager.SearchSpells("fire")
	if len(results) != 1 || results[0].ID != "fire_spell" {
		t.Errorf("Expected 1 fire spell, got %d", len(results))
	}

	// Test searching by keyword
	results = manager.SearchSpells("damage")
	if len(results) != 2 {
		t.Errorf("Expected 2 damage spells, got %d", len(results))
	}

	// Test searching by description
	results = manager.SearchSpells("healing")
	if len(results) != 1 || results[0].ID != "heal_spell" {
		t.Errorf("Expected 1 healing spell, got %d", len(results))
	}

	// Test case insensitive search
	results = manager.SearchSpells("FIRE")
	if len(results) != 1 {
		t.Errorf("Case insensitive search failed, expected 1 result, got %d", len(results))
	}
}

func TestSpellSchool_String(t *testing.T) {
	tests := []struct {
		school   SpellSchool
		expected string
	}{
		{SchoolAbjuration, "Abjuration"},
		{SchoolConjuration, "Conjuration"},
		{SchoolDivination, "Divination"},
		{SchoolEnchantment, "Enchantment"},
		{SchoolEvocation, "Evocation"},
		{SchoolIllusion, "Illusion"},
		{SchoolNecromancy, "Necromancy"},
		{SchoolTransmutation, "Transmutation"},
	}

	for _, test := range tests {
		result := test.school.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestParseSpellSchool(t *testing.T) {
	tests := []struct {
		input    string
		expected SpellSchool
	}{
		{"Abjuration", SchoolAbjuration},
		{"abjuration", SchoolAbjuration},
		{"Evocation", SchoolEvocation},
		{"evocation", SchoolEvocation},
		{"Unknown", SchoolEvocation}, // Default
		{"", SchoolEvocation},        // Default
	}

	for _, test := range tests {
		result := ParseSpellSchool(test.input)
		if result != test.expected {
			t.Errorf("ParseSpellSchool(%s): expected %d, got %d", test.input, test.expected, result)
		}
	}
}

func TestSpellComponent_String(t *testing.T) {
	tests := []struct {
		component SpellComponent
		expected  string
	}{
		{ComponentVerbal, "Verbal"},
		{ComponentSomatic, "Somatic"},
		{ComponentMaterial, "Material"},
	}

	for _, test := range tests {
		result := test.component.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestParseSpellComponent(t *testing.T) {
	tests := []struct {
		input    string
		expected SpellComponent
	}{
		{"Verbal", ComponentVerbal},
		{"verbal", ComponentVerbal},
		{"V", ComponentVerbal},
		{"Somatic", ComponentSomatic},
		{"somatic", ComponentSomatic},
		{"S", ComponentSomatic},
		{"Material", ComponentMaterial},
		{"material", ComponentMaterial},
		{"M", ComponentMaterial},
		{"Unknown", ComponentVerbal}, // Default
	}

	for _, test := range tests {
		result := ParseSpellComponent(test.input)
		if result != test.expected {
			t.Errorf("ParseSpellComponent(%s): expected %d, got %d", test.input, test.expected, result)
		}
	}
}

// TestSpellManager_GetAllSpells tests the GetAllSpells method
func TestSpellManager_GetAllSpells(t *testing.T) {
	sm := NewSpellManager("")

	// Add test spells
	spell1 := &Spell{ID: "spell1", Name: "Magic Missile", Level: 1}
	spell2 := &Spell{ID: "spell2", Name: "Fireball", Level: 3}
	spell3 := &Spell{ID: "spell3", Name: "Cure Light Wounds", Level: 1}

	sm.AddSpell(spell1)
	sm.AddSpell(spell2)
	sm.AddSpell(spell3)

	allSpells := sm.GetAllSpells()

	// Should return all spells sorted by level, then name
	if len(allSpells) != 3 {
		t.Errorf("GetAllSpells() returned %d spells; want 3", len(allSpells))
	}

	// Check sorting: level 1 spells first (alphabetical), then level 3
	if allSpells[0].Level != 1 {
		t.Errorf("First spell level = %d; want 1", allSpells[0].Level)
	}
	if allSpells[2].Level != 3 {
		t.Errorf("Last spell level = %d; want 3", allSpells[2].Level)
	}
}

// TestSpellManager_UpdateSpell tests the UpdateSpell method
func TestSpellManager_UpdateSpell(t *testing.T) {
	sm := NewSpellManager("")

	// Add initial spell
	original := &Spell{ID: "spell1", Name: "Magic Missile", Level: 1}
	sm.AddSpell(original)

	// Update the spell
	updated := &Spell{ID: "spell1", Name: "Greater Magic Missile", Level: 2}
	err := sm.UpdateSpell(updated)
	if err != nil {
		t.Errorf("UpdateSpell() unexpected error: %v", err)
	}

	// Verify update
	result, _ := sm.GetSpell("spell1")
	if result.Name != "Greater Magic Missile" {
		t.Errorf("Spell name = %s; want 'Greater Magic Missile'", result.Name)
	}
	if result.Level != 2 {
		t.Errorf("Spell level = %d; want 2", result.Level)
	}
}

// TestSpellManager_UpdateSpell_NotFound tests updating non-existent spell
func TestSpellManager_UpdateSpell_NotFound(t *testing.T) {
	sm := NewSpellManager("")

	spell := &Spell{ID: "nonexistent", Name: "Unknown Spell", Level: 1}
	err := sm.UpdateSpell(spell)

	if err == nil {
		t.Error("UpdateSpell() expected error for non-existent spell, got nil")
	}
}

// TestSpellManager_RemoveSpell tests the RemoveSpell method
func TestSpellManager_RemoveSpell(t *testing.T) {
	sm := NewSpellManager("")

	// Add spell
	spell := &Spell{ID: "spell1", Name: "Magic Missile", Level: 1}
	sm.AddSpell(spell)

	// Remove spell
	err := sm.RemoveSpell("spell1")
	if err != nil {
		t.Errorf("RemoveSpell() unexpected error: %v", err)
	}

	// Verify removal
	_, err = sm.GetSpell("spell1")
	if err == nil {
		t.Error("GetSpell() expected error after removal, got nil")
	}
}

// TestSpellManager_RemoveSpell_NotFound tests removing non-existent spell
func TestSpellManager_RemoveSpell_NotFound(t *testing.T) {
	sm := NewSpellManager("")

	err := sm.RemoveSpell("nonexistent")

	if err == nil {
		t.Error("RemoveSpell() expected error for non-existent spell, got nil")
	}
}

// TestSpellManager_GetSpellCount tests the GetSpellCount method
func TestSpellManager_GetSpellCount(t *testing.T) {
	sm := NewSpellManager("")

	if sm.GetSpellCount() != 0 {
		t.Errorf("GetSpellCount() = %d; want 0 for empty manager", sm.GetSpellCount())
	}

	sm.AddSpell(&Spell{ID: "spell1", Name: "Spell 1", Level: 1})
	sm.AddSpell(&Spell{ID: "spell2", Name: "Spell 2", Level: 2})

	if sm.GetSpellCount() != 2 {
		t.Errorf("GetSpellCount() = %d; want 2", sm.GetSpellCount())
	}
}

// TestSpellManager_GetSpellCountByLevel tests the GetSpellCountByLevel method
func TestSpellManager_GetSpellCountByLevel(t *testing.T) {
	sm := NewSpellManager("")

	sm.AddSpell(&Spell{ID: "spell1", Name: "Spell 1", Level: 1})
	sm.AddSpell(&Spell{ID: "spell2", Name: "Spell 2", Level: 1})
	sm.AddSpell(&Spell{ID: "spell3", Name: "Spell 3", Level: 2})
	sm.AddSpell(&Spell{ID: "spell4", Name: "Spell 4", Level: 3})

	counts := sm.GetSpellCountByLevel()

	if counts[1] != 2 {
		t.Errorf("Count for level 1 = %d; want 2", counts[1])
	}
	if counts[2] != 1 {
		t.Errorf("Count for level 2 = %d; want 1", counts[2])
	}
	if counts[3] != 1 {
		t.Errorf("Count for level 3 = %d; want 1", counts[3])
	}
}
