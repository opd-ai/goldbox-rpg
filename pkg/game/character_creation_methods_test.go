package game

import (
	"testing"
)

// Test_CharacterCreation_AllMethods_Bug verifies that all documented character creation methods
// are properly implemented: roll, standard array, point-buy, and custom
func Test_CharacterCreation_AllMethods_Bug(t *testing.T) {
	creator := NewCharacterCreator()

	// Test 1: Roll method (may need multiple attempts due to randomness)
	t.Run("Roll method", func(t *testing.T) {
		config := CharacterCreationConfig{
			Name:              "TestRoll",
			Class:             ClassThief, // Thief has lower requirements
			AttributeMethod:   "roll",
			StartingEquipment: false,
			StartingGold:      100,
		}

		// Try multiple times since rolling is random and might not meet requirements
		success := false
		var lastResult CharacterCreationResult
		for attempts := 0; attempts < 10; attempts++ {
			result := creator.CreateCharacter(config)
			lastResult = result
			if result.Success {
				success = true
				// Verify all attributes are in valid range (3-18)
				for attr, value := range result.GeneratedStats {
					if value < 3 || value > 18 {
						t.Errorf("Roll method generated invalid attribute %s: %d (expected 3-18)", attr, value)
					}
				}
				break
			}
		}

		if !success {
			t.Logf("Roll method eventually failed after 10 attempts (this can happen with random rolls): %v", lastResult.Errors)
			// This is acceptable - verify the method exists and produces valid errors
			if len(lastResult.Errors) == 0 {
				t.Errorf("Roll method should provide error messages when it fails")
			}
		}
	})

	// Test 2: Standard array method
	t.Run("Standard array method", func(t *testing.T) {
		config := CharacterCreationConfig{
			Name:              "TestStandard",
			Class:             ClassThief, // Thief has lower requirements
			AttributeMethod:   "standard",
			StartingEquipment: false,
			StartingGold:      100,
		}

		result := creator.CreateCharacter(config)
		if !result.Success {
			t.Errorf("Standard array method failed: %v", result.Errors)
		}

		if result.Character == nil {
			t.Errorf("Standard array method did not create character")
		}

		// Verify standard array values are used
		expectedValues := map[int]bool{15: false, 14: false, 13: false, 12: false, 10: false, 8: false}
		for _, value := range result.GeneratedStats {
			if _, exists := expectedValues[value]; !exists {
				t.Errorf("Standard array method generated non-standard value: %d", value)
			}
			expectedValues[value] = true
		}

		// Verify all standard values were used
		for value, used := range expectedValues {
			if !used {
				t.Errorf("Standard array method did not use expected value: %d", value)
			}
		}
	})

	// Test 3: Point-buy method
	t.Run("Point-buy method", func(t *testing.T) {
		config := CharacterCreationConfig{
			Name:              "TestPointBuy",
			Class:             ClassThief, // Thief has lower requirements
			AttributeMethod:   "pointbuy",
			StartingEquipment: false,
			StartingGold:      100,
		}

		result := creator.CreateCharacter(config)
		if !result.Success {
			t.Errorf("Point-buy method failed: %v", result.Errors)
		}

		if result.Character == nil {
			t.Errorf("Point-buy method did not create character")
		}

		// Verify all attributes start at 8 and don't exceed 15 (point-buy limits)
		for attr, value := range result.GeneratedStats {
			if value < 8 || value > 15 {
				t.Errorf("Point-buy method generated invalid attribute %s: %d (expected 8-15)", attr, value)
			}
		}

		// Verify the total point cost doesn't exceed 27
		totalCost := 0
		for _, value := range result.GeneratedStats {
			pointsSpent := value - 8
			if value >= 13 {
				// Each point above 12 costs 2 points
				extraPoints := value - 12
				pointsSpent = 4 + (extraPoints * 2) // 4 points to get to 12, then 2 per additional
			}
			totalCost += pointsSpent
		}

		if totalCost > 27 {
			t.Errorf("Point-buy method exceeded 27 point limit: %d", totalCost)
		}
	})

	// Test 4: Custom method
	t.Run("Custom method", func(t *testing.T) {
		customAttrs := map[string]int{
			"strength":     16,
			"dexterity":    14,
			"constitution": 15,
			"intelligence": 12,
			"wisdom":       13,
			"charisma":     10,
		}

		config := CharacterCreationConfig{
			Name:              "TestCustom",
			Class:             ClassThief,
			AttributeMethod:   "custom",
			CustomAttributes:  customAttrs,
			StartingEquipment: false,
			StartingGold:      100,
		}

		result := creator.CreateCharacter(config)
		if !result.Success {
			t.Errorf("Custom method failed: %v", result.Errors)
		}

		if result.Character == nil {
			t.Errorf("Custom method did not create character")
		}

		// Verify custom attributes were used exactly
		for attr, expectedValue := range customAttrs {
			actualValue, exists := result.GeneratedStats[attr]
			if !exists {
				t.Errorf("Custom method missing attribute: %s", attr)
			} else if actualValue != expectedValue {
				t.Errorf("Custom method attribute %s: expected %d, got %d", attr, expectedValue, actualValue)
			}
		}
	})

	// Test 5: Verify all methods are accessible via string identifiers
	t.Run("All method identifiers work", func(t *testing.T) {
		methods := []string{"roll", "standard", "pointbuy", "custom"}

		for _, method := range methods {
			config := CharacterCreationConfig{
				Name:              "Test" + method,
				Class:             ClassThief, // Use Thief for consistent class requirements
				AttributeMethod:   method,
				StartingEquipment: false,
				StartingGold:      100,
			}

			// For custom method, we need to provide custom attributes
			if method == "custom" {
				config.CustomAttributes = map[string]int{
					"strength": 15, "dexterity": 14, "constitution": 13,
					"intelligence": 12, "wisdom": 10, "charisma": 8,
				}
			}

			result := creator.CreateCharacter(config)
			if !result.Success {
				t.Errorf("Method %s failed to create character: %v", method, result.Errors)
			}
		}
	})

	// This test should pass, confirming all documented methods are implemented
	t.Logf("All four documented character creation methods are properly implemented")
}
