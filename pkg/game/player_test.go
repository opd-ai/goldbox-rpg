package game

import (
	"strings"
	"testing"
)

func TestPlayer_Update_ValidData_UpdatesFields(t *testing.T) {
	// Create a test player with embedded character
	char := &Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		Class:        ClassFighter,
		HP:           100,
		MaxHP:        100,
		Strength:     15,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
	}

	player := &Player{
		Character:   *char.Clone(),
		Level:       1,
		Experience:  0,
		QuestLog:    []Quest{},
		KnownSpells: []Spell{},
	}

	tests := []struct {
		name       string
		updateData map[string]interface{}
		verify     func(*testing.T, *Player)
	}{
		{
			name: "update class",
			updateData: map[string]interface{}{
				"class": ClassMage,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Class != ClassMage {
					t.Errorf("Class = %v, expected %v", p.Class, ClassMage)
				}
			},
		},
		{
			name: "update level",
			updateData: map[string]interface{}{
				"level": 5,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Level != 5 {
					t.Errorf("Level = %d, expected 5", p.Level)
				}
			},
		},
		{
			name: "update experience",
			updateData: map[string]interface{}{
				"experience": 1500,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Experience != 1500 {
					t.Errorf("Experience = %d, expected 1500", p.Experience)
				}
			},
		},
		{
			name: "update hp",
			updateData: map[string]interface{}{
				"hp": 75,
			},
			verify: func(t *testing.T, p *Player) {
				if p.HP != 75 {
					t.Errorf("HP = %d, expected 75", p.HP)
				}
			},
		},
		{
			name: "update max_hp",
			updateData: map[string]interface{}{
				"max_hp": 120,
			},
			verify: func(t *testing.T, p *Player) {
				if p.MaxHP != 120 {
					t.Errorf("MaxHP = %d, expected 120", p.MaxHP)
				}
			},
		},
		{
			name: "update strength",
			updateData: map[string]interface{}{
				"strength": 18,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Strength != 18 {
					t.Errorf("Strength = %d, expected 18", p.Strength)
				}
			},
		},
		{
			name: "multiple fields",
			updateData: map[string]interface{}{
				"level":      3,
				"experience": 900,
				"hp":         90,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Level != 3 {
					t.Errorf("Level = %d, expected 3", p.Level)
				}
				if p.Experience != 900 {
					t.Errorf("Experience = %d, expected 900", p.Experience)
				}
				if p.HP != 90 {
					t.Errorf("HP = %d, expected 90", p.HP)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset player to initial state
			player.Character.Class = ClassFighter
			player.Level = 1
			player.Experience = 0
			player.HP = 100
			player.MaxHP = 100
			player.Strength = 15

			player.Update(tt.updateData)
			tt.verify(t, player)
		})
	}
}

func TestPlayer_Update_InvalidFields_IgnoresUnknownFields(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		HP:           100,
		MaxHP:        100,
		Strength:     15,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
	}

	player := &Player{
		Character:  *char.Clone(),
		Level:      1,
		Experience: 0,
	}

	originalLevel := player.Level
	originalExp := player.Experience

	// Try to update with invalid field names
	updateData := map[string]interface{}{
		"invalidField":    "should be ignored",
		"anotherInvalid":  123,
		"unknownProperty": true,
	}

	player.Update(updateData)

	// Verify original values are unchanged
	if player.Level != originalLevel {
		t.Errorf("Level changed from %d to %d, should be unchanged", originalLevel, player.Level)
	}
	if player.Experience != originalExp {
		t.Errorf("Experience changed from %d to %d, should be unchanged", originalExp, player.Experience)
	}
}

func TestPlayer_Clone_CreatesIndependentCopy(t *testing.T) {
	// Create original player with complex data
	originalQuests := []Quest{
		{ID: "quest1", Title: "First Quest", Status: QuestActive},
		{ID: "quest2", Title: "Second Quest", Status: QuestCompleted},
	}
	originalSpells := []Spell{
		{ID: "spell1", Name: "Fireball", Level: 1},
		{ID: "spell2", Name: "Heal", Level: 1},
	}

	char := &Character{
		ID:           "test-char-1",
		Name:         "Original Character",
		Class:        ClassRanger,
		HP:           80,
		MaxHP:        100,
		Strength:     16,
		Dexterity:    14,
		Constitution: 15,
		Intelligence: 12,
	}

	original := &Player{
		Character:   *char.Clone(),
		Level:       3,
		Experience:  1200,
		QuestLog:    originalQuests,
		KnownSpells: originalSpells,
	}

	// Clone the player
	clone := original.Clone()

	// Verify clone is not nil
	if clone == nil {
		t.Fatal("Clone() returned nil")
	}

	// Verify basic fields are copied
	if clone.Class != original.Class {
		t.Errorf("Clone Class = %v, expected %v", clone.Class, original.Class)
	}
	if clone.Level != original.Level {
		t.Errorf("Clone Level = %d, expected %d", clone.Level, original.Level)
	}
	if clone.Experience != original.Experience {
		t.Errorf("Clone Experience = %d, expected %d", clone.Experience, original.Experience)
	}

	// Verify character data is copied
	if clone.Character.Name != original.Character.Name {
		t.Errorf("Clone Character Name = %s, expected %s", clone.Character.Name, original.Character.Name)
	}
	if clone.Character.HP != original.Character.HP {
		t.Errorf("Clone Character HP = %d, expected %d", clone.Character.HP, original.Character.HP)
	}

	// Verify quest log is deep copied
	if len(clone.QuestLog) != len(original.QuestLog) {
		t.Errorf("Clone QuestLog length = %d, expected %d", len(clone.QuestLog), len(original.QuestLog))
	}
	if len(clone.QuestLog) > 0 && &clone.QuestLog[0] == &original.QuestLog[0] {
		t.Error("QuestLog is not deep copied - shares memory with original")
	}

	// Verify known spells is deep copied
	if len(clone.KnownSpells) != len(original.KnownSpells) {
		t.Errorf("Clone KnownSpells length = %d, expected %d", len(clone.KnownSpells), len(original.KnownSpells))
	}
	if len(clone.KnownSpells) > 0 && &clone.KnownSpells[0] == &original.KnownSpells[0] {
		t.Error("KnownSpells is not deep copied - shares memory with original")
	}

	// Modify clone and verify original is unaffected
	clone.Level = 5
	clone.Experience = 2000
	if len(clone.QuestLog) > 0 {
		clone.QuestLog[0].Title = "Modified Quest"
	}

	if original.Level == clone.Level {
		t.Error("Modifying clone Level affected original")
	}
	if original.Experience == clone.Experience {
		t.Error("Modifying clone Experience affected original")
	}
	if len(original.QuestLog) > 0 && original.QuestLog[0].Title == "Modified Quest" {
		t.Error("Modifying clone QuestLog affected original")
	}
}

func TestPlayer_Clone_NilPlayer_ReturnsNil(t *testing.T) {
	var player *Player = nil
	clone := player.Clone()

	if clone != nil {
		t.Errorf("Clone() on nil player returned %v, expected nil", clone)
	}
}

func TestPlayer_PublicData_ReturnsCorrectStructure(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Public Test Character",
		Class:        ClassCleric,
		HP:           75,
		MaxHP:        100,
		Strength:     17,
		Dexterity:    13,
		Constitution: 16,
		Intelligence: 11,
	}

	player := &Player{
		Character:   *char.Clone(),
		Level:       4,
		Experience:  1800,
		QuestLog:    []Quest{{ID: "quest1", Title: "Secret Quest"}},
		KnownSpells: []Spell{{ID: "spell1", Name: "Secret Spell"}},
	}

	publicData := player.PublicData()

	// Verify the returned data is a map
	data := publicData
	if data == nil {
		t.Fatal("PublicData() returned nil")
	}

	// Verify expected fields are present
	expectedFields := []string{"name", "class", "hp", "max_hp", "strength", "constitution"}
	for _, field := range expectedFields {
		if _, exists := data[field]; !exists {
			t.Errorf("PublicData() missing field: %s", field)
		}
	}

	// Verify specific values
	if data["name"] != "Public Test Character" {
		t.Errorf("name = %v, expected Public Test Character", data["name"])
	}
	if data["class"] != ClassCleric {
		t.Errorf("class = %v, expected %v", data["class"], ClassCleric)
	}
	if data["hp"] != 75 {
		t.Errorf("hp = %v, expected 75", data["hp"])
	}
	if data["max_hp"] != 100 {
		t.Errorf("max_hp = %v, expected 100", data["max_hp"])
	}

	// Verify sensitive data is not included
	sensitiveFields := []string{"experience", "questLog", "knownSpells"}
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			t.Errorf("PublicData() should not include sensitive field: %s", field)
		}
	}
}

func TestPlayer_AddExperience_ValidValues_AddsExperience(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		Class:        ClassFighter,
		HP:           100,
		MaxHP:        100,
		Strength:     15,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
	}

	tests := []struct {
		name             string
		initialExp       int
		expToAdd         int
		expectedFinalExp int
		expectError      bool
	}{
		{
			name:             "add positive experience",
			initialExp:       0,
			expToAdd:         100,
			expectedFinalExp: 100,
			expectError:      false,
		},
		{
			name:             "add zero experience",
			initialExp:       500,
			expToAdd:         0,
			expectedFinalExp: 500,
			expectError:      false,
		},
		{
			name:             "add large amount",
			initialExp:       1000,
			expToAdd:         5000,
			expectedFinalExp: 6000,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{
				Character:  *char.Clone(),
				Level:      1,
				Experience: tt.initialExp,
			}

			err := player.AddExperience(tt.expToAdd)

			if tt.expectError && err == nil {
				t.Error("AddExperience() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("AddExperience() unexpected error: %v", err)
			}

			if player.Experience != tt.expectedFinalExp {
				t.Errorf("Experience = %d, expected %d", player.Experience, tt.expectedFinalExp)
			}
		})
	}
}

func TestPlayer_AddExperience_NegativeValue_ReturnsError(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		Class:        ClassFighter,
		HP:           100,
		MaxHP:        100,
		Strength:     15,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
	}

	player := &Player{
		Character:  *char.Clone(),
		Level:      1,
		Experience: 100,
	}

	initialExp := player.Experience
	err := player.AddExperience(-50)

	if err == nil {
		t.Error("AddExperience() with negative value should return error")
	}

	if player.Experience != initialExp {
		t.Errorf("Experience changed to %d, should remain %d when error occurs", player.Experience, initialExp)
	}

	// Verify error message
	expectedErrorSubstring := "cannot add negative experience"
	if err.Error() != "cannot add negative experience: -50" {
		t.Errorf("Error message = %q, expected to contain %q", err.Error(), expectedErrorSubstring)
	}
}

func TestPlayer_AddExperience_LevelUp_CallsLevelUpLogic(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		Class:        ClassFighter,
		HP:           100,
		MaxHP:        100,
		Strength:     15,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
	}

	player := &Player{
		Character:  *char.Clone(),
		Level:      0,    // Start at level 0
		Experience: 1800, // Just below level 1 threshold (2000)
	}

	initialHP := player.HP
	initialMaxHP := player.MaxHP
	initialLevel := player.Level

	// Add enough experience to trigger level up from 0 to 1
	err := player.AddExperience(400) // Should trigger level up to 1 (1800 + 400 = 2200 XP)

	if err != nil {
		t.Errorf("AddExperience() returned unexpected error: %v", err)
	}

	// Verify level increased
	if player.Level <= initialLevel {
		t.Errorf("Level = %d, expected > %d (level up should occur)", player.Level, initialLevel)
	}

	// Verify HP increased (level up should increase HP)
	if player.MaxHP <= initialMaxHP {
		t.Errorf("MaxHP = %d, expected > %d (level up should increase MaxHP)", player.MaxHP, initialMaxHP)
	}

	if player.HP <= initialHP {
		t.Errorf("HP = %d, expected > %d (level up should increase HP)", player.HP, initialHP)
	}

	// Verify experience was added
	expectedExp := 1800 + 400
	if player.Experience != expectedExp {
		t.Errorf("Experience = %d, expected %d", player.Experience, expectedExp)
	}
}

func TestPlayer_GetStats_ReturnsCorrectStats(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		Class:        ClassPaladin,
		HP:           85,
		MaxHP:        120,
		Strength:     18,
		Dexterity:    14,
		Constitution: 16,
		Intelligence: 13,
	}

	player := &Player{
		Character:  *char.Clone(),
		Level:      5,
		Experience: 2500,
	}

	stats := player.GetStats()

	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	// Verify all stat fields are correctly converted to float64
	if stats.Health != float64(player.HP) {
		t.Errorf("Health = %f, expected %f", stats.Health, float64(player.HP))
	}
	if stats.MaxHealth != float64(player.MaxHP) {
		t.Errorf("MaxHealth = %f, expected %f", stats.MaxHealth, float64(player.MaxHP))
	}
	if stats.Strength != float64(player.Strength) {
		t.Errorf("Strength = %f, expected %f", stats.Strength, float64(player.Strength))
	}
	if stats.Dexterity != float64(player.Dexterity) {
		t.Errorf("Dexterity = %f, expected %f", stats.Dexterity, float64(player.Dexterity))
	}
	if stats.Intelligence != float64(player.Intelligence) {
		t.Errorf("Intelligence = %f, expected %f", stats.Intelligence, float64(player.Intelligence))
	}

	// Verify mana is set to intelligence
	if stats.Mana != float64(player.Intelligence) {
		t.Errorf("Mana = %f, expected %f", stats.Mana, float64(player.Intelligence))
	}
	if stats.MaxMana != float64(player.Intelligence) {
		t.Errorf("MaxMana = %f, expected %f", stats.MaxMana, float64(player.Intelligence))
	}

	// Verify default values for unset fields
	if stats.Defense != 0 {
		t.Errorf("Defense = %f, expected 0", stats.Defense)
	}
	if stats.Speed != 0 {
		t.Errorf("Speed = %f, expected 0", stats.Speed)
	}
}

func TestPlayer_GetStats_ZeroValues_HandlesCorrectly(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		Class:        ClassFighter,
		HP:           0,
		MaxHP:        0,
		Strength:     0,
		Dexterity:    0,
		Constitution: 0,
		Intelligence: 0,
	}

	player := &Player{
		Character:  *char.Clone(),
		Level:      1,
		Experience: 0,
	}

	stats := player.GetStats()

	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	// Verify all stats are 0.0
	expectedZero := float64(0)
	if stats.Health != expectedZero {
		t.Errorf("Health = %f, expected %f", stats.Health, expectedZero)
	}
	if stats.MaxHealth != expectedZero {
		t.Errorf("MaxHealth = %f, expected %f", stats.MaxHealth, expectedZero)
	}
	if stats.Strength != expectedZero {
		t.Errorf("Strength = %f, expected %f", stats.Strength, expectedZero)
	}
	if stats.Dexterity != expectedZero {
		t.Errorf("Dexterity = %f, expected %f", stats.Dexterity, expectedZero)
	}
	if stats.Intelligence != expectedZero {
		t.Errorf("Intelligence = %f, expected %f", stats.Intelligence, expectedZero)
	}
	if stats.Mana != expectedZero {
		t.Errorf("Mana = %f, expected %f", stats.Mana, expectedZero)
	}
	if stats.MaxMana != expectedZero {
		t.Errorf("MaxMana = %f, expected %f", stats.MaxMana, expectedZero)
	}
}

// Table-driven test for comprehensive Update method testing
func TestPlayer_Update_ComprehensiveFields(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		Class:        ClassFighter,
		HP:           100,
		MaxHP:        100,
		Strength:     15,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
	}

	tests := []struct {
		name       string
		updateData map[string]interface{}
		expectFunc func(*testing.T, *Player)
	}{
		{
			name: "update intelligence and dexterity",
			updateData: map[string]interface{}{
				"intelligence": 16,
				"dexterity":    18,
			},
			expectFunc: func(t *testing.T, p *Player) {
				if p.Intelligence != 16 {
					t.Errorf("Intelligence = %d, expected 16", p.Intelligence)
				}
				if p.Dexterity != 18 {
					t.Errorf("Dexterity = %d, expected 18", p.Dexterity)
				}
			},
		},
		{
			name: "update constitution",
			updateData: map[string]interface{}{
				"constitution": 20,
			},
			expectFunc: func(t *testing.T, p *Player) {
				if p.Constitution != 20 {
					t.Errorf("Constitution = %d, expected 20", p.Constitution)
				}
			},
		},
		{
			name: "update all basic stats",
			updateData: map[string]interface{}{
				"class":        ClassMage,
				"level":        10,
				"experience":   5000,
				"hp":           80,
				"max_hp":       150,
				"strength":     12,
				"dexterity":    16,
				"constitution": 18,
				"intelligence": 20,
			},
			expectFunc: func(t *testing.T, p *Player) {
				if p.Class != ClassMage {
					t.Errorf("Class = %v, expected %v", p.Class, ClassMage)
				}
				if p.Level != 10 {
					t.Errorf("Level = %d, expected 10", p.Level)
				}
				if p.Experience != 5000 {
					t.Errorf("Experience = %d, expected 5000", p.Experience)
				}
				if p.HP != 80 {
					t.Errorf("HP = %d, expected 80", p.HP)
				}
				if p.MaxHP != 150 {
					t.Errorf("MaxHP = %d, expected 150", p.MaxHP)
				}
				if p.Strength != 12 {
					t.Errorf("Strength = %d, expected 12", p.Strength)
				}
				if p.Dexterity != 16 {
					t.Errorf("Dexterity = %d, expected 16", p.Dexterity)
				}
				if p.Constitution != 18 {
					t.Errorf("Constitution = %d, expected 18", p.Constitution)
				}
				if p.Intelligence != 20 {
					t.Errorf("Intelligence = %d, expected 20", p.Intelligence)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{
				Character:   *char.Clone(),
				Level:       1,
				Experience:  0,
				QuestLog:    []Quest{},
				KnownSpells: []Spell{},
			}

			player.Update(tt.updateData)
			tt.expectFunc(t, player)
		})
	}
}

// TestPlayer_Update_CharacterFields tests updating Character-specific fields
func TestPlayer_Update_CharacterFields(t *testing.T) {
	char := &Character{
		ID:           "test-char-1",
		Name:         "Original Name",
		Description:  "Original Description",
		Class:        ClassFighter,
		Position:     Position{X: 0, Y: 0, Level: 1, Facing: DirectionNorth},
		HP:           100,
		MaxHP:        100,
		Strength:     15,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
		Wisdom:       12,
		Charisma:     13,
		ArmorClass:   15,
		THAC0:        18,
		Gold:         100,
	}

	player := &Player{
		Character:  *char.Clone(),
		Level:      1,
		Experience: 0,
	}

	tests := []struct {
		name       string
		updateData map[string]interface{}
		verify     func(*testing.T, *Player)
	}{
		{
			name: "update character name",
			updateData: map[string]interface{}{
				"name": "New Character Name",
			},
			verify: func(t *testing.T, p *Player) {
				if p.Character.Name != "New Character Name" {
					t.Errorf("Character.Name = %q, expected %q", p.Character.Name, "New Character Name")
				}
			},
		},
		{
			name: "update character description",
			updateData: map[string]interface{}{
				"description": "Updated character description",
			},
			verify: func(t *testing.T, p *Player) {
				if p.Character.Description != "Updated character description" {
					t.Errorf("Character.Description = %q, expected %q", p.Character.Description, "Updated character description")
				}
			},
		},
		{
			name: "update position components",
			updateData: map[string]interface{}{
				"position_x":      5,
				"position_y":      10,
				"position_level":  2,
				"position_facing": DirectionEast,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Character.Position.X != 5 {
					t.Errorf("Position.X = %d, expected 5", p.Character.Position.X)
				}
				if p.Character.Position.Y != 10 {
					t.Errorf("Position.Y = %d, expected 10", p.Character.Position.Y)
				}
				if p.Character.Position.Level != 2 {
					t.Errorf("Position.Level = %d, expected 2", p.Character.Position.Level)
				}
				if p.Character.Position.Facing != DirectionEast {
					t.Errorf("Position.Facing = %v, expected %v", p.Character.Position.Facing, DirectionEast)
				}
			},
		},
		{
			name: "update wisdom and charisma",
			updateData: map[string]interface{}{
				"wisdom":   16,
				"charisma": 18,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Character.Wisdom != 16 {
					t.Errorf("Wisdom = %d, expected 16", p.Character.Wisdom)
				}
				if p.Character.Charisma != 18 {
					t.Errorf("Charisma = %d, expected 18", p.Character.Charisma)
				}
			},
		},
		{
			name: "update combat stats",
			updateData: map[string]interface{}{
				"armor_class": 12,
				"thac0":       16,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Character.ArmorClass != 12 {
					t.Errorf("ArmorClass = %d, expected 12", p.Character.ArmorClass)
				}
				if p.Character.THAC0 != 16 {
					t.Errorf("THAC0 = %d, expected 16", p.Character.THAC0)
				}
			},
		},
		{
			name: "update gold",
			updateData: map[string]interface{}{
				"gold": 500,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Character.Gold != 500 {
					t.Errorf("Gold = %d, expected 500", p.Character.Gold)
				}
			},
		},
		{
			name: "update multiple character and player fields",
			updateData: map[string]interface{}{
				"name":       "Multi Update Test",
				"wisdom":     14,
				"charisma":   15,
				"gold":       250,
				"level":      3,
				"experience": 1200,
			},
			verify: func(t *testing.T, p *Player) {
				if p.Character.Name != "Multi Update Test" {
					t.Errorf("Character.Name = %q, expected %q", p.Character.Name, "Multi Update Test")
				}
				if p.Character.Wisdom != 14 {
					t.Errorf("Wisdom = %d, expected 14", p.Character.Wisdom)
				}
				if p.Character.Charisma != 15 {
					t.Errorf("Charisma = %d, expected 15", p.Character.Charisma)
				}
				if p.Character.Gold != 250 {
					t.Errorf("Gold = %d, expected 250", p.Character.Gold)
				}
				if p.Level != 3 {
					t.Errorf("Level = %d, expected 3", p.Level)
				}
				if p.Experience != 1200 {
					t.Errorf("Experience = %d, expected 1200", p.Experience)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset player to initial state
			player.Character.Name = "Original Name"
			player.Character.Description = "Original Description"
			player.Character.Position = Position{X: 0, Y: 0, Level: 1, Facing: DirectionNorth}
			player.Character.Wisdom = 12
			player.Character.Charisma = 13
			player.Character.ArmorClass = 15
			player.Character.THAC0 = 18
			player.Character.Gold = 100
			player.Level = 1
			player.Experience = 0

			player.Update(tt.updateData)
			tt.verify(t, player)
		})
	}
}

func TestPlayer_AddExperience_IntegerOverflow_ReturnsError(t *testing.T) {
	char := &Character{
		ID:           "test-char-overflow",
		Name:         "Test Character",
		Class:        ClassFighter,
		HP:           100,
		MaxHP:        100,
		Strength:     15,
		Constitution: 14,
	}

	player := &Player{
		Character:  *char.Clone(),
		Level:      50,
		Experience: 1 << 62, // Very large experience value close to max int64
	}

	initialExp := player.Experience
	initialLevel := player.Level

	// Try to add experience that would cause overflow
	err := player.AddExperience(1 << 62)

	if err == nil {
		t.Error("AddExperience() with overflow value should return error")
	}

	// Verify experience and level remain unchanged when overflow error occurs
	if player.Experience != initialExp {
		t.Errorf("Experience changed to %d, should remain %d when overflow error occurs", player.Experience, initialExp)
	}

	if player.Level != initialLevel {
		t.Errorf("Level changed to %d, should remain %d when overflow error occurs", player.Level, initialLevel)
	}

	// Verify error message mentions overflow
	expectedErrorSubstring := "overflow"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Error message = %q, expected to contain %q", err.Error(), expectedErrorSubstring)
	}
}
