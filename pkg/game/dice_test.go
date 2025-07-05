package game

import (
	"testing"
)

func TestDiceRoller_Roll(t *testing.T) {
	// Use a fixed seed for reproducible tests
	roller := NewDiceRollerWithSeed(12345)

	tests := []struct {
		expression    string
		expectError   bool
		expectedRolls int // Number of dice rolls expected
	}{
		{"1d6", false, 1},
		{"2d8", false, 2},
		{"3d6+2", false, 3},
		{"1d20-1", false, 1},
		{"4d4+4", false, 4},
		{"", false, 0},
		{"invalid", true, 0},
		{"d6", true, 0},
		{"1d", true, 0},
		{"0d6", true, 0},
		{"1d0", true, 0},
	}

	for _, test := range tests {
		result, err := roller.Roll(test.expression)

		if test.expectError {
			if err == nil {
				t.Errorf("Expected error for expression '%s', but got none", test.expression)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for expression '%s': %v", test.expression, err)
			continue
		}

		if len(result.Rolls) != test.expectedRolls {
			t.Errorf("Expression '%s': expected %d rolls, got %d", test.expression, test.expectedRolls, len(result.Rolls))
		}

		// Verify rolls are within valid range
		for i, roll := range result.Rolls {
			if roll < 1 {
				t.Errorf("Expression '%s': roll %d is less than 1: %d", test.expression, i, roll)
			}
		}

		// Verify total is sum of rolls
		expectedTotal := 0
		for _, roll := range result.Rolls {
			expectedTotal += roll
		}
		if result.Total != expectedTotal {
			t.Errorf("Expression '%s': total mismatch. Expected %d, got %d", test.expression, expectedTotal, result.Total)
		}

		// Verify final is total + modifier
		expectedFinal := result.Total + result.Modifier
		if result.Final != expectedFinal {
			t.Errorf("Expression '%s': final mismatch. Expected %d, got %d", test.expression, expectedFinal, result.Final)
		}
	}
}

func TestDiceRoller_RollWithModifiers(t *testing.T) {
	roller := NewDiceRollerWithSeed(54321)

	// Test positive modifier
	result, err := roller.Roll("1d6+3")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Modifier != 3 {
		t.Errorf("Expected modifier 3, got %d", result.Modifier)
	}
	if result.Final != result.Total+3 {
		t.Errorf("Final should be total + modifier: %d + %d = %d, got %d", result.Total, result.Modifier, result.Total+result.Modifier, result.Final)
	}

	// Test negative modifier
	result, err = roller.Roll("1d6-2")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Modifier != -2 {
		t.Errorf("Expected modifier -2, got %d", result.Modifier)
	}
	if result.Final != result.Total-2 {
		t.Errorf("Final should be total + modifier: %d + %d = %d, got %d", result.Total, result.Modifier, result.Total+result.Modifier, result.Final)
	}
}

func TestDiceRoller_RollMultiple(t *testing.T) {
	roller := NewDiceRollerWithSeed(98765)

	expressions := []string{"1d6", "2d4+1", "1d8-1"}
	result, err := roller.RollMultiple(expressions)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have 1 + 2 + 1 = 4 individual rolls
	expectedRolls := 4
	if len(result.Rolls) != expectedRolls {
		t.Errorf("Expected %d total rolls, got %d", expectedRolls, len(result.Rolls))
	}

	// Verify total and modifier calculations
	expectedTotal := 0
	for _, roll := range result.Rolls {
		expectedTotal += roll
	}
	if result.Total != expectedTotal {
		t.Errorf("Total mismatch. Expected %d, got %d", expectedTotal, result.Total)
	}

	// Expected modifier: 0 + 1 + (-1) = 0
	expectedModifier := 0
	if result.Modifier != expectedModifier {
		t.Errorf("Expected modifier %d, got %d", expectedModifier, result.Modifier)
	}
}

func TestCalculateDiceAverage(t *testing.T) {
	tests := []struct {
		expression string
		expected   float64
		expectErr  bool
	}{
		{"1d6", 3.5, false},    // (6+1)/2 = 3.5
		{"2d6", 7.0, false},    // 2 * 3.5 = 7.0
		{"1d20", 10.5, false},  // (20+1)/2 = 10.5
		{"3d6+3", 13.5, false}, // 3 * 3.5 + 3 = 13.5
		{"1d8-1", 3.5, false},  // (8+1)/2 - 1 = 3.5
		{"", 0, false},         // Empty expression = 0
		{"invalid", 0, true},   // Invalid expression
		{"1d0", 0, true},       // Invalid die size
	}

	for _, test := range tests {
		result, err := CalculateDiceAverage(test.expression)

		if test.expectErr {
			if err == nil {
				t.Errorf("Expected error for expression '%s', but got none", test.expression)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for expression '%s': %v", test.expression, err)
			continue
		}

		if result != test.expected {
			t.Errorf("Expression '%s': expected average %f, got %f", test.expression, test.expected, result)
		}
	}
}

func TestDiceRoll_String(t *testing.T) {
	tests := []struct {
		roll     *DiceRoll
		expected string
	}{
		{
			&DiceRoll{Rolls: []int{3, 4, 2}, Total: 9, Modifier: 0, Final: 9},
			"[3 4 2] = 9",
		},
		{
			&DiceRoll{Rolls: []int{5}, Total: 5, Modifier: 3, Final: 8},
			"[5] + 3 = 8",
		},
		{
			&DiceRoll{Rolls: []int{6, 1}, Total: 7, Modifier: -2, Final: 5},
			"[6 1] - 2 = 5",
		},
		{
			&DiceRoll{Rolls: []int{}, Total: 0, Modifier: 0, Final: 0},
			"0",
		},
	}

	for i, test := range tests {
		result := test.roll.String()
		if result != test.expected {
			t.Errorf("Test %d: expected '%s', got '%s'", i, test.expected, result)
		}
	}
}

func TestDiceRollerDeterministic(t *testing.T) {
	// Test that the same seed produces the same results
	seed := int64(42)

	roller1 := NewDiceRollerWithSeed(seed)
	roller2 := NewDiceRollerWithSeed(seed)

	expression := "3d6+2"

	result1, err1 := roller1.Roll(expression)
	if err1 != nil {
		t.Fatalf("Roller 1 error: %v", err1)
	}

	result2, err2 := roller2.Roll(expression)
	if err2 != nil {
		t.Fatalf("Roller 2 error: %v", err2)
	}

	// Results should be identical
	if len(result1.Rolls) != len(result2.Rolls) {
		t.Errorf("Different number of rolls: %d vs %d", len(result1.Rolls), len(result2.Rolls))
	}

	for i := 0; i < len(result1.Rolls); i++ {
		if result1.Rolls[i] != result2.Rolls[i] {
			t.Errorf("Roll %d differs: %d vs %d", i, result1.Rolls[i], result2.Rolls[i])
		}
	}

	if result1.Final != result2.Final {
		t.Errorf("Final results differ: %d vs %d", result1.Final, result2.Final)
	}
}
