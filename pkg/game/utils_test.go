package game

import (
	"regexp"
	"testing"
)

func TestNewUID(t *testing.T) {
	t.Run("GeneratesUniqueIdentifiers", func(t *testing.T) {
		// Generate multiple UIDs to test uniqueness
		uids := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			uid := NewUID()
			if uids[uid] {
				t.Errorf("Generated duplicate UID: %s", uid)
			}
			uids[uid] = true
		}
	})

	t.Run("ReturnsCorrectFormat", func(t *testing.T) {
		uid := NewUID()

		// Should be 36 characters long (UUID format)
		if len(uid) != 36 {
			t.Errorf("Expected UID length 36, got %d", len(uid))
		}

		// Should match UUID v4 format
		uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
		if !uuidPattern.MatchString(uid) {
			t.Errorf("UID does not match UUID v4 format: %s", uid)
		}
	})

	t.Run("GeneratesNonEmptyUID", func(t *testing.T) {
		uid := NewUID()
		if uid == "" {
			t.Error("Generated UID should not be empty")
		}
	})
}

func TestIsValidPosition(t *testing.T) {
	tests := []struct {
		name     string
		position Position
		width    int
		height   int
		maxLevel int
		expected bool
	}{
		{
			name:     "ValidPositionAllPositive",
			position: Position{X: 5, Y: 9, Level: 0},
			width:    10, height: 10, maxLevel: 1,
			expected: true,
		},
		{
			name:     "ValidPositionZeroValues",
			position: Position{X: 0, Y: 0, Level: 0},
			width:    10, height: 10, maxLevel: 1,
			expected: true,
		},
		{
			name:     "InvalidPositionNegativeX",
			position: Position{X: -1, Y: 5, Level: 0},
			width:    10, height: 10, maxLevel: 1,
			expected: false,
		},
		{
			name:     "InvalidPositionNegativeY",
			position: Position{X: 5, Y: -1, Level: 0},
			width:    10, height: 10, maxLevel: 1,
			expected: false,
		},
		{
			name:     "InvalidPositionNegativeLevel",
			position: Position{X: 5, Y: 9, Level: -1},
			width:    10, height: 10, maxLevel: 1,
			expected: false,
		},
		{
			name:     "InvalidPositionAllNegative",
			position: Position{X: -1, Y: -1, Level: -1},
			width:    10, height: 10, maxLevel: 1,
			expected: false,
		},
		{
			name:     "InvalidPositionTooLarge",
			position: Position{X: 9999, Y: 9999, Level: 9999},
			width:    10, height: 10, maxLevel: 1,
			expected: false,
		},
		{
			name:     "InvalidPositionXTooLarge",
			position: Position{X: 10, Y: 5, Level: 0},
			width:    10, height: 10, maxLevel: 1,
			expected: false,
		},
		{
			name:     "InvalidPositionYTooLarge",
			position: Position{X: 5, Y: 10, Level: 0},
			width:    10, height: 10, maxLevel: 1,
			expected: false,
		},
		{
			name:     "InvalidPositionLevelTooLarge",
			position: Position{X: 5, Y: 5, Level: 1},
			width:    10, height: 10, maxLevel: 1,
			expected: false,
		},
		{
			name:     "ValidPositionAtMaxMinusOne",
			position: Position{X: 9, Y: 9, Level: 0},
			width:    10, height: 10, maxLevel: 1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPosition(tt.position, tt.width, tt.height, tt.maxLevel)
			if result != tt.expected {
				t.Errorf("isValidPosition(%+v, %d, %d, %d) = %v, expected %v", tt.position, tt.width, tt.height, tt.maxLevel, result, tt.expected)
			}
		})
	}
}

func TestCalculateLevel(t *testing.T) {
	tests := []struct {
		name     string
		exp      int64
		expected int
	}{
		{
			name:     "Level1_MinimumExperience",
			exp:      0,
			expected: 1,
		},
		{
			name:     "Level1_MaximumExperience",
			exp:      1999,
			expected: 1,
		},
		{
			name:     "Level2_MinimumExperience",
			exp:      2000,
			expected: 2,
		},
		{
			name:     "Level2_MaximumExperience",
			exp:      3999,
			expected: 2,
		},
		{
			name:     "Level3_MinimumExperience",
			exp:      4000,
			expected: 3,
		},
		{
			name:     "Level4_MinimumExperience",
			exp:      8000,
			expected: 4,
		},
		{
			name:     "Level5_MinimumExperience",
			exp:      16000,
			expected: 5,
		},
		{
			name:     "Level6_MinimumExperience",
			exp:      32000,
			expected: 6,
		},
		{
			name:     "Level7_MinimumExperience",
			exp:      64000,
			expected: 7,
		},
		{
			name:     "MaxLevel_HighExperience",
			exp:      100000,
			expected: 7,
		},
		{
			name:     "NegativeExperience_ReturnsLevel0",
			exp:      -100,
			expected: 0,
		},
		{
			name:     "BoundaryTest_JustBelowLevel2",
			exp:      1999,
			expected: 1,
		},
		{
			name:     "BoundaryTest_ExactlyLevel2",
			exp:      2000,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateLevel(tt.exp)
			if result != tt.expected {
				t.Errorf("calculateLevel(%d) = %d, expected %d", tt.exp, result, tt.expected)
			}
		})
	}
}

func TestCalculateHealthGain(t *testing.T) {
	tests := []struct {
		name         string
		class        CharacterClass
		constitution int
		expected     int
	}{
		{
			name:         "Fighter_AverageConstitution",
			class:        ClassFighter,
			constitution: 10,
			expected:     10, // base 10 + 0 con bonus
		},
		{
			name:         "Fighter_HighConstitution",
			class:        ClassFighter,
			constitution: 16,
			expected:     13, // base 10 + 3 con bonus
		},
		{
			name:         "Fighter_LowConstitution",
			class:        ClassFighter,
			constitution: 8,
			expected:     9, // base 10 + (-1) con bonus
		},
		{
			name:         "Mage_AverageConstitution",
			class:        ClassMage,
			constitution: 10,
			expected:     4, // base 4 + 0 con bonus
		},
		{
			name:         "Mage_HighConstitution",
			class:        ClassMage,
			constitution: 18,
			expected:     8, // base 4 + 4 con bonus
		},
		{
			name:         "Mage_LowConstitution",
			class:        ClassMage,
			constitution: 6,
			expected:     2, // base 4 + (-2) con bonus
		},
		{
			name:         "Cleric_AverageConstitution",
			class:        ClassCleric,
			constitution: 10,
			expected:     8, // base 8 + 0 con bonus
		},
		{
			name:         "Thief_AverageConstitution",
			class:        ClassThief,
			constitution: 10,
			expected:     6, // base 6 + 0 con bonus
		},
		{
			name:         "Ranger_AverageConstitution",
			class:        ClassRanger,
			constitution: 10,
			expected:     8, // base 8 + 0 con bonus
		},
		{
			name:         "Paladin_AverageConstitution",
			class:        ClassPaladin,
			constitution: 10,
			expected:     10, // base 10 + 0 con bonus
		},
		{
			name:         "Paladin_VeryHighConstitution",
			class:        ClassPaladin,
			constitution: 20,
			expected:     15, // base 10 + 5 con bonus
		},
		{
			name:         "Thief_VeryLowConstitution",
			class:        ClassThief,
			constitution: 3,
			expected:     3, // base 6 + (-3) con bonus
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateHealthGain(tt.class, tt.constitution)
			if result != tt.expected {
				t.Errorf("calculateHealthGain(%v, %d) = %d, expected %d", tt.class, tt.constitution, result, tt.expected)
			}
		})
	}
}

func TestCalculateHealthGain_AllClasses(t *testing.T) {
	// Test that all classes have defined base health gains
	classes := []CharacterClass{
		ClassFighter,
		ClassMage,
		ClassCleric,
		ClassThief,
		ClassRanger,
		ClassPaladin,
	}

	constitution := 10 // neutral constitution
	for _, class := range classes {
		t.Run("Class_"+class.String(), func(t *testing.T) {
			result := calculateHealthGain(class, constitution)
			// All classes should have positive health gain with neutral constitution
			if result <= 0 {
				t.Errorf("calculateHealthGain(%v, %d) = %d, expected positive value", class, constitution, result)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkNewUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewUID()
	}
}

func BenchmarkIsValidPosition(b *testing.B) {
	pos := Position{X: 10, Y: 20, Level: 1}
	for i := 0; i < b.N; i++ {
		isValidPosition(pos, 100, 100, 10)
	}
}

func BenchmarkCalculateLevel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculateLevel(25000)
	}
}

func BenchmarkCalculateHealthGain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculateHealthGain(ClassFighter, 14)
	}
}
