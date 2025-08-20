package game

import (
	"testing"
)

// Test for Character Class String() method panic bug
func Test_CharacterClass_String_Panic_Bug2(t *testing.T) {
	t.Run("Valid character classes should work", func(t *testing.T) {
		validClasses := []CharacterClass{
			ClassFighter,
			ClassMage,
			ClassCleric,
			ClassThief,
			ClassRanger,
			ClassPaladin,
		}

		expectedNames := []string{
			"Fighter",
			"Mage",
			"Cleric",
			"Thief",
			"Ranger",
			"Paladin",
		}

		for i, class := range validClasses {
			result := class.String()
			if result != expectedNames[i] {
				t.Errorf("Expected %s, got %s for class %d", expectedNames[i], result, class)
			}
		}
	})

	t.Run("Invalid character class should return Unknown (bug fixed)", func(t *testing.T) {
		// This should no longer panic with the fix
		invalidClass := CharacterClass(6) // Outside valid range 0-5
		result := invalidClass.String()

		if result != "Unknown" {
			t.Errorf("Expected 'Unknown' for invalid character class, got '%s'", result)
		} else {
			t.Logf("✅ Bug fixed: Invalid character class returns 'Unknown' instead of panicking")
		}
	})

	t.Run("Negative character class should return Unknown (bug fixed)", func(t *testing.T) {
		// This should no longer panic with the fix
		invalidClass := CharacterClass(-1)
		result := invalidClass.String()

		if result != "Unknown" {
			t.Errorf("Expected 'Unknown' for negative character class, got '%s'", result)
		} else {
			t.Logf("✅ Bug fixed: Negative character class returns 'Unknown' instead of panicking")
		}
	})

	t.Run("Very large character class should return Unknown (bug fixed)", func(t *testing.T) {
		// This should no longer panic with the fix
		invalidClass := CharacterClass(999)
		result := invalidClass.String()

		if result != "Unknown" {
			t.Errorf("Expected 'Unknown' for very large character class, got '%s'", result)
		} else {
			t.Logf("✅ Bug fixed: Very large character class returns 'Unknown' instead of panicking")
		}
	})
}
