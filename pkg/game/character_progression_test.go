package game

import (
	"testing"
)

func TestCharacterExperienceAndLevel(t *testing.T) {
	// Create a test character
	char := &Character{
		ID:         "test-char",
		Name:       "Test Character",
		Level:      1,
		Experience: 0,
	}

	// Test initial state
	if char.GetLevel() != 1 {
		t.Errorf("Expected initial level 1, got %d", char.GetLevel())
	}
	if char.GetExperience() != 0 {
		t.Errorf("Expected initial experience 0, got %d", char.GetExperience())
	}

	// Test adding experience without level up
	leveledUp, err := char.AddExperience(500)
	if err != nil {
		t.Errorf("Unexpected error adding experience: %v", err)
	}
	if leveledUp {
		t.Errorf("Character should not have leveled up at 500 XP")
	}
	if char.GetExperience() != 500 {
		t.Errorf("Expected 500 experience, got %d", char.GetExperience())
	}
	if char.GetLevel() != 1 {
		t.Errorf("Expected level 1, got %d", char.GetLevel())
	}

	// Test level up
	leveledUp, err = char.AddExperience(500) // Total: 1000 XP
	if err != nil {
		t.Errorf("Unexpected error adding experience: %v", err)
	}
	if !leveledUp {
		t.Errorf("Character should have leveled up at 1000 XP")
	}
	if char.GetLevel() != 2 {
		t.Errorf("Expected level 2, got %d", char.GetLevel())
	}
	if char.GetExperience() != 1000 {
		t.Errorf("Expected 1000 experience, got %d", char.GetExperience())
	}

	// Test experience to next level
	toNext := char.GetExperienceToNextLevel()
	if toNext != 1000 { // Need 2000 total for level 3, have 1000
		t.Errorf("Expected 1000 XP to next level, got %d", toNext)
	}

	// Test multiple level ups at once
	leveledUp, err = char.AddExperience(3000) // Total: 4000 XP (should reach level 4)
	if err != nil {
		t.Errorf("Unexpected error adding experience: %v", err)
	}
	if !leveledUp {
		t.Errorf("Character should have leveled up with large XP gain")
	}
	if char.GetLevel() != 4 {
		t.Errorf("Expected level 4, got %d", char.GetLevel())
	}

	// Test setting level directly
	err = char.SetLevel(10)
	if err != nil {
		t.Errorf("Unexpected error setting level: %v", err)
	}
	if char.GetLevel() != 10 {
		t.Errorf("Expected level 10, got %d", char.GetLevel())
	}

	// Test setting experience directly
	err = char.SetExperience(50000)
	if err != nil {
		t.Errorf("Unexpected error setting experience: %v", err)
	}
	if char.GetExperience() != 50000 {
		t.Errorf("Expected 50000 experience, got %d", char.GetExperience())
	}
}

func TestCharacterExperienceValidation(t *testing.T) {
	char := &Character{
		ID:         "test-char",
		Level:      1,
		Experience: 0,
	}

	// Test negative experience
	_, err := char.AddExperience(-100)
	if err == nil {
		t.Errorf("Expected error when adding negative experience")
	}

	// Test setting negative experience
	err = char.SetExperience(-100)
	if err == nil {
		t.Errorf("Expected error when setting negative experience")
	}

	// Test setting invalid level
	err = char.SetLevel(0)
	if err == nil {
		t.Errorf("Expected error when setting level to 0")
	}

	err = char.SetLevel(-1)
	if err == nil {
		t.Errorf("Expected error when setting negative level")
	}
}

func TestCharacterCloneWithExperience(t *testing.T) {
	original := &Character{
		ID:         "original",
		Level:      5,
		Experience: 10000,
	}

	clone := original.Clone()

	if clone.GetLevel() != original.GetLevel() {
		t.Errorf("Clone level %d doesn't match original %d", clone.GetLevel(), original.GetLevel())
	}
	if clone.GetExperience() != original.GetExperience() {
		t.Errorf("Clone experience %d doesn't match original %d", clone.GetExperience(), original.GetExperience())
	}

	// Modify clone and ensure original is unchanged
	clone.AddExperience(1000)
	if original.GetExperience() == clone.GetExperience() {
		t.Errorf("Modifying clone should not affect original")
	}
}

func TestExperienceTable(t *testing.T) {
	char := &Character{ID: "test"}

	tests := []struct {
		level      int
		requiredXP int64
	}{
		{1, 0},
		{2, 1000},
		{3, 2000},
		{4, 4000},
		{5, 8000},
		{10, 200000},
	}

	for _, test := range tests {
		actual := char.getExperienceRequiredForLevel(test.level)
		if actual != test.requiredXP {
			t.Errorf("Level %d: expected %d XP, got %d", test.level, test.requiredXP, actual)
		}
	}

	// Test max level
	maxLevelXP := char.getExperienceRequiredForLevel(21)
	if maxLevelXP != -1 {
		t.Errorf("Expected -1 for level beyond max, got %d", maxLevelXP)
	}
}
