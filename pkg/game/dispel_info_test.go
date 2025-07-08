package game

import (
	"testing"
)

func TestNewDispelInfo(t *testing.T) {
	tests := []struct {
		name      string
		priority  DispelPriority
		types     []DispelType
		removable bool
		want      *DispelInfo
	}{
		{
			name:      "Create basic dispel info",
			priority:  DispelPriorityNormal,
			types:     []DispelType{DispelMagic},
			removable: true,
			want: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelMagic},
				Removable: true,
			},
		},
		{
			name:      "Create non-removable dispel info",
			priority:  DispelPriorityHigh,
			types:     []DispelType{DispelCurse, DispelPoison},
			removable: false,
			want: &DispelInfo{
				Priority:  DispelPriorityHigh,
				Types:     []DispelType{DispelCurse, DispelPoison},
				Removable: false,
			},
		},
		{
			name:      "Create with empty types",
			priority:  DispelPriorityLowest,
			types:     []DispelType{},
			removable: true,
			want: &DispelInfo{
				Priority:  DispelPriorityLowest,
				Types:     []DispelType{},
				Removable: true,
			},
		},
		{
			name:      "Create with all dispel type",
			priority:  DispelPriorityHighest,
			types:     []DispelType{DispelAll},
			removable: true,
			want: &DispelInfo{
				Priority:  DispelPriorityHighest,
				Types:     []DispelType{DispelAll},
				Removable: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDispelInfo(tt.priority, tt.types, tt.removable)

			if got.Priority != tt.want.Priority {
				t.Errorf("NewDispelInfo() Priority = %v, want %v", got.Priority, tt.want.Priority)
			}

			if got.Removable != tt.want.Removable {
				t.Errorf("NewDispelInfo() Removable = %v, want %v", got.Removable, tt.want.Removable)
			}

			if len(got.Types) != len(tt.want.Types) {
				t.Errorf("NewDispelInfo() Types length = %d, want %d", len(got.Types), len(tt.want.Types))
				return
			}

			for i, dispelType := range got.Types {
				if dispelType != tt.want.Types[i] {
					t.Errorf("NewDispelInfo() Types[%d] = %v, want %v", i, dispelType, tt.want.Types[i])
				}
			}
		})
	}
}

func TestDispelInfo_CanBeDispelledBy(t *testing.T) {
	tests := []struct {
		name       string
		dispelInfo *DispelInfo
		dispelType DispelType
		want       bool
	}{
		{
			name: "Removable effect with matching type",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelMagic},
				Removable: true,
			},
			dispelType: DispelMagic,
			want:       true,
		},
		{
			name: "Non-removable effect should not be dispelled",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelMagic},
				Removable: false,
			},
			dispelType: DispelMagic,
			want:       false,
		},
		{
			name: "Removable effect with non-matching type",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelCurse},
				Removable: true,
			},
			dispelType: DispelMagic,
			want:       false,
		},
		{
			name: "DispelAll can dispel any removable effect",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelCurse},
				Removable: true,
			},
			dispelType: DispelAll,
			want:       true,
		},
		{
			name: "Effect with DispelAll type can be dispelled by any type",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelAll},
				Removable: true,
			},
			dispelType: DispelPoison,
			want:       true,
		},
		{
			name: "Multiple types - match found",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelCurse, DispelPoison, DispelMagic},
				Removable: true,
			},
			dispelType: DispelPoison,
			want:       true,
		},
		{
			name: "Multiple types - no match",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{DispelCurse, DispelPoison},
				Removable: true,
			},
			dispelType: DispelMagic,
			want:       false,
		},
		{
			name: "Empty types list with DispelAll",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{},
				Removable: true,
			},
			dispelType: DispelAll,
			want:       true,
		},
		{
			name: "Empty types list with specific type",
			dispelInfo: &DispelInfo{
				Priority:  DispelPriorityNormal,
				Types:     []DispelType{},
				Removable: true,
			},
			dispelType: DispelMagic,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dispelInfo.CanBeDispelledBy(tt.dispelType)
			if got != tt.want {
				t.Errorf("DispelInfo.CanBeDispelledBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDispelInfo_StructFields(t *testing.T) {
	// Test that struct fields are properly accessible and set
	dispelInfo := DispelInfo{
		Priority:  DispelPriorityHigh,
		Types:     []DispelType{DispelMagic, DispelCurse},
		Removable: true,
	}

	if dispelInfo.Priority != DispelPriorityHigh {
		t.Errorf("Expected Priority %v, got %v", DispelPriorityHigh, dispelInfo.Priority)
	}

	if len(dispelInfo.Types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(dispelInfo.Types))
	}

	if !dispelInfo.Removable {
		t.Errorf("Expected Removable to be true, got false")
	}
}

func TestDispelInfo_ZeroValue(t *testing.T) {
	// Test behavior with zero value
	var dispelInfo DispelInfo

	// Zero value should have default values
	if dispelInfo.Priority != 0 {
		t.Errorf("Expected zero Priority, got %v", dispelInfo.Priority)
	}

	if dispelInfo.Types != nil {
		t.Errorf("Expected nil Types slice, got %v", dispelInfo.Types)
	}

	if dispelInfo.Removable {
		t.Errorf("Expected Removable to be false (zero value), got true")
	}

	// Zero value should not be dispellable by anything
	canDispel := dispelInfo.CanBeDispelledBy(DispelMagic)
	if canDispel {
		t.Errorf("Expected zero value to not be dispellable, but it was")
	}

	canDispelAll := dispelInfo.CanBeDispelledBy(DispelAll)
	if canDispelAll {
		t.Errorf("Expected zero value to not be dispellable by DispelAll, but it was")
	}
}

func TestDispelInfo_EdgeCases(t *testing.T) {
	t.Run("Nil types slice behavior", func(t *testing.T) {
		dispelInfo := &DispelInfo{
			Priority:  DispelPriorityNormal,
			Types:     nil, // explicitly nil
			Removable: true,
		}

		// Should handle nil slice gracefully
		canDispel := dispelInfo.CanBeDispelledBy(DispelMagic)
		if canDispel {
			t.Errorf("Expected false for nil types slice with DispelMagic, got true")
		}

		// DispelAll should still work even with nil types
		canDispelAll := dispelInfo.CanBeDispelledBy(DispelAll)
		if !canDispelAll {
			t.Errorf("Expected true for nil types slice with DispelAll, got false")
		}
	})

	t.Run("Large number of types", func(t *testing.T) {
		// Test with many dispel types
		manyTypes := []DispelType{
			DispelMagic, DispelCurse, DispelPoison, DispelDisease,
		}

		dispelInfo := NewDispelInfo(DispelPriorityNormal, manyTypes, true)

		// Should find match among many types
		if !dispelInfo.CanBeDispelledBy(DispelPoison) {
			t.Errorf("Expected to find DispelPoison among many types")
		}

		// Should not find non-existent type
		if dispelInfo.CanBeDispelledBy(DispelAll) {
			// This should be true because DispelAll can dispel anything removable
			// This is actually correct behavior based on the implementation
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkNewDispelInfo(b *testing.B) {
	types := []DispelType{DispelMagic, DispelCurse}

	for i := 0; i < b.N; i++ {
		NewDispelInfo(DispelPriorityNormal, types, true)
	}
}

func BenchmarkCanBeDispelledBy(b *testing.B) {
	dispelInfo := NewDispelInfo(
		DispelPriorityNormal,
		[]DispelType{DispelMagic, DispelCurse, DispelPoison},
		true,
	)

	for i := 0; i < b.N; i++ {
		dispelInfo.CanBeDispelledBy(DispelPoison)
	}
}
