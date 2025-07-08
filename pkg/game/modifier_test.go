package game

import (
	"testing"
)

// TestNewModifier tests the NewModifier constructor function
func TestNewModifier(t *testing.T) {
	tests := []struct {
		name      string
		stat      string
		value     float64
		operation ModOpType
	}{
		{
			name:      "create strength add modifier",
			stat:      "strength",
			value:     5.0,
			operation: ModAdd,
		},
		{
			name:      "create dexterity multiply modifier",
			stat:      "dexterity",
			value:     1.5,
			operation: ModMultiply,
		},
		{
			name:      "create health set modifier",
			stat:      "health",
			value:     100.0,
			operation: ModSet,
		},
		{
			name:      "create negative value modifier",
			stat:      "armor",
			value:     -2.0,
			operation: ModAdd,
		},
		{
			name:      "create zero value modifier",
			stat:      "mana",
			value:     0.0,
			operation: ModAdd,
		},
		{
			name:      "create decimal value modifier",
			stat:      "speed",
			value:     2.75,
			operation: ModMultiply,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := NewModifier(tt.stat, tt.value, tt.operation)

			if modifier == nil {
				t.Fatal("NewModifier returned nil")
			}

			if modifier.Stat != tt.stat {
				t.Errorf("NewModifier().Stat = %v, want %v", modifier.Stat, tt.stat)
			}

			if modifier.Value != tt.value {
				t.Errorf("NewModifier().Value = %v, want %v", modifier.Value, tt.value)
			}

			if modifier.Operation != tt.operation {
				t.Errorf("NewModifier().Operation = %v, want %v", modifier.Operation, tt.operation)
			}
		})
	}
}

// TestModifier_Apply tests the Apply method for all operation types
func TestModifier_Apply(t *testing.T) {
	tests := []struct {
		name      string
		modifier  Modifier
		baseValue float64
		expected  float64
	}{
		// ModAdd operations
		{
			name: "add positive value",
			modifier: Modifier{
				Stat:      "strength",
				Value:     5.0,
				Operation: ModAdd,
			},
			baseValue: 10.0,
			expected:  15.0,
		},
		{
			name: "add negative value",
			modifier: Modifier{
				Stat:      "strength",
				Value:     -3.0,
				Operation: ModAdd,
			},
			baseValue: 10.0,
			expected:  7.0,
		},
		{
			name: "add zero value",
			modifier: Modifier{
				Stat:      "strength",
				Value:     0.0,
				Operation: ModAdd,
			},
			baseValue: 10.0,
			expected:  10.0,
		},
		{
			name: "add decimal value",
			modifier: Modifier{
				Stat:      "strength",
				Value:     2.5,
				Operation: ModAdd,
			},
			baseValue: 10.0,
			expected:  12.5,
		},
		// ModMultiply operations
		{
			name: "multiply by positive value",
			modifier: Modifier{
				Stat:      "damage",
				Value:     2.0,
				Operation: ModMultiply,
			},
			baseValue: 10.0,
			expected:  20.0,
		},
		{
			name: "multiply by decimal value",
			modifier: Modifier{
				Stat:      "damage",
				Value:     1.5,
				Operation: ModMultiply,
			},
			baseValue: 10.0,
			expected:  15.0,
		},
		{
			name: "multiply by zero",
			modifier: Modifier{
				Stat:      "damage",
				Value:     0.0,
				Operation: ModMultiply,
			},
			baseValue: 10.0,
			expected:  0.0,
		},
		{
			name: "multiply by one",
			modifier: Modifier{
				Stat:      "damage",
				Value:     1.0,
				Operation: ModMultiply,
			},
			baseValue: 10.0,
			expected:  10.0,
		},
		{
			name: "multiply by negative value",
			modifier: Modifier{
				Stat:      "damage",
				Value:     -2.0,
				Operation: ModMultiply,
			},
			baseValue: 10.0,
			expected:  -20.0,
		},
		// ModSet operations
		{
			name: "set to positive value",
			modifier: Modifier{
				Stat:      "health",
				Value:     100.0,
				Operation: ModSet,
			},
			baseValue: 50.0,
			expected:  100.0,
		},
		{
			name: "set to zero",
			modifier: Modifier{
				Stat:      "health",
				Value:     0.0,
				Operation: ModSet,
			},
			baseValue: 50.0,
			expected:  0.0,
		},
		{
			name: "set to negative value",
			modifier: Modifier{
				Stat:      "health",
				Value:     -10.0,
				Operation: ModSet,
			},
			baseValue: 50.0,
			expected:  -10.0,
		},
		{
			name: "set to decimal value",
			modifier: Modifier{
				Stat:      "health",
				Value:     75.5,
				Operation: ModSet,
			},
			baseValue: 50.0,
			expected:  75.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.modifier.Apply(tt.baseValue)
			if result != tt.expected {
				t.Errorf("Modifier.Apply(%v) = %v, want %v", tt.baseValue, result, tt.expected)
			}
		})
	}
}

// TestModifier_Apply_InvalidOperation tests Apply method with invalid operation
func TestModifier_Apply_InvalidOperation(t *testing.T) {
	modifier := Modifier{
		Stat:      "test",
		Value:     5.0,
		Operation: ModOpType("invalid"), // Invalid operation
	}

	baseValue := 10.0
	result := modifier.Apply(baseValue)

	// Should return base value unchanged for invalid operations
	if result != baseValue {
		t.Errorf("Modifier.Apply() with invalid operation = %v, want %v (base value unchanged)", result, baseValue)
	}
}

// TestModifier_Apply_EdgeCases tests edge cases for the Apply method
func TestModifier_Apply_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		modifier  Modifier
		baseValue float64
		expected  float64
	}{
		{
			name: "very large positive base value",
			modifier: Modifier{
				Stat:      "test",
				Value:     1.0,
				Operation: ModAdd,
			},
			baseValue: 1e10,
			expected:  1e10 + 1.0,
		},
		{
			name: "very large negative base value",
			modifier: Modifier{
				Stat:      "test",
				Value:     1.0,
				Operation: ModAdd,
			},
			baseValue: -1e10,
			expected:  -1e10 + 1.0,
		},
		{
			name: "very small decimal base value",
			modifier: Modifier{
				Stat:      "test",
				Value:     0.001,
				Operation: ModAdd,
			},
			baseValue: 0.001,
			expected:  0.002,
		},
		{
			name: "multiply very large number",
			modifier: Modifier{
				Stat:      "test",
				Value:     2.0,
				Operation: ModMultiply,
			},
			baseValue: 1e10,
			expected:  2e10,
		},
		{
			name: "multiply very small decimal",
			modifier: Modifier{
				Stat:      "test",
				Value:     0.5,
				Operation: ModMultiply,
			},
			baseValue: 0.0001,
			expected:  0.00005,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.modifier.Apply(tt.baseValue)
			if result != tt.expected {
				t.Errorf("Modifier.Apply(%v) = %v, want %v", tt.baseValue, result, tt.expected)
			}
		})
	}
}

// TestModifier_StructFields tests that Modifier struct fields are properly accessible
func TestModifier_StructFields(t *testing.T) {
	modifier := Modifier{
		Stat:      "test_stat",
		Value:     42.5,
		Operation: ModAdd,
	}

	// Test direct field access
	if modifier.Stat != "test_stat" {
		t.Errorf("Modifier.Stat = %v, want %v", modifier.Stat, "test_stat")
	}

	if modifier.Value != 42.5 {
		t.Errorf("Modifier.Value = %v, want %v", modifier.Value, 42.5)
	}

	if modifier.Operation != ModAdd {
		t.Errorf("Modifier.Operation = %v, want %v", modifier.Operation, ModAdd)
	}
}

// TestModifier_ZeroValues tests Modifier with zero values
func TestModifier_ZeroValues(t *testing.T) {
	var modifier Modifier // Zero value

	// Test zero value fields
	if modifier.Stat != "" {
		t.Errorf("Zero Modifier.Stat = %v, want empty string", modifier.Stat)
	}

	if modifier.Value != 0.0 {
		t.Errorf("Zero Modifier.Value = %v, want 0.0", modifier.Value)
	}

	// Test applying zero value modifier (should use default case)
	result := modifier.Apply(10.0)
	if result != 10.0 {
		t.Errorf("Zero Modifier.Apply(10.0) = %v, want 10.0", result)
	}
}

// TestModifier_StringComparison tests modifiers with various string stats
func TestModifier_StringComparison(t *testing.T) {
	tests := []struct {
		name string
		stat string
	}{
		{"empty string stat", ""},
		{"single character stat", "a"},
		{"normal stat name", "strength"},
		{"stat with spaces", "max health"},
		{"stat with numbers", "level1_bonus"},
		{"stat with special chars", "buff-effect_modifier"},
		{"very long stat name", "this_is_a_very_long_stat_name_that_might_be_used_in_some_edge_case"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := NewModifier(tt.stat, 1.0, ModAdd)
			if modifier.Stat != tt.stat {
				t.Errorf("NewModifier() with stat %q, got stat %q", tt.stat, modifier.Stat)
			}
		})
	}
}

// TestModifier_OperationConsistency tests that all operations are consistent
func TestModifier_OperationConsistency(t *testing.T) {
	baseValue := 10.0

	// Test that operations are consistent across multiple calls
	addModifier := NewModifier("test", 5.0, ModAdd)
	for i := 0; i < 5; i++ {
		result := addModifier.Apply(baseValue)
		if result != 15.0 {
			t.Errorf("Add operation inconsistent on call %d: got %v, want 15.0", i+1, result)
		}
	}

	multiplyModifier := NewModifier("test", 2.0, ModMultiply)
	for i := 0; i < 5; i++ {
		result := multiplyModifier.Apply(baseValue)
		if result != 20.0 {
			t.Errorf("Multiply operation inconsistent on call %d: got %v, want 20.0", i+1, result)
		}
	}

	setModifier := NewModifier("test", 100.0, ModSet)
	for i := 0; i < 5; i++ {
		result := setModifier.Apply(baseValue)
		if result != 100.0 {
			t.Errorf("Set operation inconsistent on call %d: got %v, want 100.0", i+1, result)
		}
	}
}
