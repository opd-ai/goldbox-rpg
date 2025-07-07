package game

import (
	"testing"
)

func TestCalculateMaxActionPoints(t *testing.T) {
	tests := []struct {
		level     int
		dexterity int
		expected  int
		desc      string
	}{
		{1, 10, 2, "Base level, normal dexterity"},
		{1, 15, 3, "Base level, high dexterity (+1 bonus)"},
		{1, 16, 3, "Base level, very high dexterity (+1 bonus)"},
		{1, 14, 2, "Base level, dexterity 14 (no bonus)"},
		{2, 10, 2, "Even level, normal dexterity"},
		{2, 15, 3, "Even level, high dexterity (+1 bonus)"},
		{3, 10, 3, "First odd level bonus, normal dexterity"},
		{3, 15, 4, "First odd level bonus, high dexterity (+1 bonus)"},
		{4, 10, 3, "Even level, same as level 3, normal dexterity"},
		{4, 15, 4, "Even level, same as level 3, high dexterity"},
		{5, 10, 4, "Second odd level bonus, normal dexterity"},
		{5, 15, 5, "Second odd level bonus, high dexterity"},
		{6, 10, 4, "Even level, same as level 5, normal dexterity"},
		{7, 10, 5, "Third odd level bonus, normal dexterity"},
		{8, 10, 5, "Even level, same as level 7, normal dexterity"},
		{9, 10, 6, "Fourth odd level bonus, normal dexterity"},
		{9, 15, 7, "Fourth odd level bonus, high dexterity"},
		{10, 10, 6, "Even level, same as level 9, normal dexterity"},
		{0, 10, 2, "Invalid level, should default to 1"},
		{-1, 10, 2, "Invalid level, should default to 1"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			actual := calculateMaxActionPoints(test.level, test.dexterity)
			if actual != test.expected {
				t.Errorf("calculateMaxActionPoints(%d, %d) = %d, expected %d",
					test.level, test.dexterity, actual, test.expected)
			}
		})
	}
}

func TestCharacterCreationWithActionPoints(t *testing.T) {
	creator := NewCharacterCreator()

	config := CharacterCreationConfig{
		Name:              "TestCharacter",
		Class:             ClassFighter,
		AttributeMethod:   "standard",
		StartingEquipment: false,
		StartingGold:      100,
	}

	result := creator.CreateCharacter(config)

	if !result.Success {
		t.Fatalf("Character creation failed: %v", result.Errors)
	}

	char := result.Character

	// New characters should start at level 1 with base action points
	// The exact amount depends on their dexterity (2 base + 1 if dex > 14)
	if char.Level != 1 {
		t.Errorf("New character level = %d, expected 1", char.Level)
	}

	expectedActionPoints := 2
	if char.Dexterity > 14 {
		expectedActionPoints = 3
	}

	if char.MaxActionPoints != expectedActionPoints {
		t.Errorf("New character MaxActionPoints = %d, expected %d (dex: %d)",
			char.MaxActionPoints, expectedActionPoints, char.Dexterity)
	}

	if char.ActionPoints != expectedActionPoints {
		t.Errorf("New character ActionPoints = %d, expected %d (dex: %d)",
			char.ActionPoints, expectedActionPoints, char.Dexterity)
	}
}

func TestPlayerLevelUpActionPoints(t *testing.T) {
	// Create a base character with dexterity <= 14 (no dex bonus)
	char := &Character{
		ID:              "test-char",
		Name:            "Test Character",
		Class:           ClassFighter,
		Level:           1,
		HP:              100,
		MaxHP:           100,
		Strength:        15,
		Constitution:    14,
		Dexterity:       14, // No dexterity bonus
		ActionPoints:    2,
		MaxActionPoints: 2,
	}

	player := &Player{
		Character:  *char.Clone(),
		Level:      1,
		Experience: 0,
	}

	// Test level up to level 3 (should gain action point)
	err := player.levelUp(3)
	if err != nil {
		t.Fatalf("levelUp failed: %v", err)
	}

	if player.Level != 3 {
		t.Errorf("Player level = %d, expected 3", player.Level)
	}

	if player.Character.MaxActionPoints != 3 {
		t.Errorf("Level 3 MaxActionPoints = %d, expected 3", player.Character.MaxActionPoints)
	}

	if player.Character.ActionPoints != 3 {
		t.Errorf("Level 3 ActionPoints = %d, expected 3 (should be restored on level up)",
			player.Character.ActionPoints)
	}

	// Test level up to level 4 (even level, no action point gain)
	err = player.levelUp(4)
	if err != nil {
		t.Fatalf("levelUp failed: %v", err)
	}

	if player.Character.MaxActionPoints != 3 {
		t.Errorf("Level 4 MaxActionPoints = %d, expected 3 (no change from level 3)",
			player.Character.MaxActionPoints)
	}

	// Test level up to level 5 (should gain another action point)
	err = player.levelUp(5)
	if err != nil {
		t.Fatalf("levelUp failed: %v", err)
	}

	if player.Character.MaxActionPoints != 4 {
		t.Errorf("Level 5 MaxActionPoints = %d, expected 4", player.Character.MaxActionPoints)
	}
}

func TestPlayerDexterityBonusActionPoints(t *testing.T) {
	// Create a character with high dexterity
	char := &Character{
		ID:              "test-char-dex",
		Name:            "High Dex Character",
		Class:           ClassThief,
		Level:           1,
		HP:              100,
		MaxHP:           100,
		Strength:        12,
		Constitution:    14,
		Dexterity:       16, // High dexterity for +1 bonus
		ActionPoints:    0,  // Will be calculated
		MaxActionPoints: 0,  // Will be calculated
	}

	// Properly calculate action points based on level and dexterity
	char.MaxActionPoints = calculateMaxActionPoints(char.Level, char.Dexterity)
	char.ActionPoints = char.MaxActionPoints

	player := &Player{
		Character:  *char.Clone(),
		Level:      1,
		Experience: 0,
	}

	// Test that high dex character starts with 3 action points (2 base + 1 dex bonus)
	if player.Character.MaxActionPoints != 3 {
		t.Errorf("High dex character MaxActionPoints = %d, expected 3", player.Character.MaxActionPoints)
	}

	// Test level up to level 3 (odd level bonus + dex bonus)
	err := player.levelUp(3)
	if err != nil {
		t.Fatalf("levelUp failed: %v", err)
	}

	// Should have 4 action points: 2 base + 1 odd level bonus + 1 dex bonus
	if player.Character.MaxActionPoints != 4 {
		t.Errorf("Level 3 high dex MaxActionPoints = %d, expected 4", player.Character.MaxActionPoints)
	}

	// Test level up to level 5 (more odd level bonuses + dex bonus)
	err = player.levelUp(5)
	if err != nil {
		t.Fatalf("levelUp failed: %v", err)
	}

	// Should have 5 action points: 2 base + 2 odd level bonuses + 1 dex bonus
	if player.Character.MaxActionPoints != 5 {
		t.Errorf("Level 5 high dex MaxActionPoints = %d, expected 5", player.Character.MaxActionPoints)
	}
}
