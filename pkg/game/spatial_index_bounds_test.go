package game

import (
	"testing"
)

// Test for bounds checking bug in spatial index
func Test_SpatialIndex_BoundsCheck_Bug2(t *testing.T) {
	// Create spatial index with bounds 0-99, 0-99
	index := NewSpatialIndex(100, 100, 10)

	t.Run("GetObjectsAt with out-of-bounds position should not panic", func(t *testing.T) {
		// Position 150,150 is outside bounds (0-99, 0-99)
		outOfBoundsPos := Position{X: 150, Y: 150}

		// This currently doesn't validate bounds and could cause issues
		objects := index.GetObjectsAt(outOfBoundsPos)

		// Should return empty slice safely
		if objects == nil {
			t.Error("GetObjectsAt should return empty slice, not nil")
		}

		// Should ideally return empty for out-of-bounds position
		// but currently doesn't validate bounds
		t.Logf("GetObjectsAt(150,150) returned %d objects", len(objects))
	})

	t.Run("Update to out-of-bounds position should fail gracefully", func(t *testing.T) {
		// First insert a valid object
		validObj := &TestGameObject{
			id:       "test1",
			name:     "Test Object",
			position: Position{X: 50, Y: 50},
			active:   true,
			health:   100,
		}

		err := index.Insert(validObj)
		if err != nil {
			t.Fatalf("Failed to insert valid object: %v", err)
		}

		// Now try to update to out-of-bounds position
		err = index.Update("test1", Position{X: 200, Y: 200})
		if err == nil {
			t.Error("Update to out-of-bounds position should fail but succeeded")
		}
	})

	t.Run("Boundary position validation", func(t *testing.T) {
		// Test boundary positions (should be valid)
		boundaryObj := &TestGameObject{
			id:       "boundary",
			name:     "Boundary Object",
			position: Position{X: 99, Y: 99}, // Should be valid
			active:   true,
			health:   100,
		}

		err := index.Insert(boundaryObj)
		if err != nil {
			t.Errorf("Failed to insert boundary object: %v", err)
		}

		// Test just outside boundary
		outsideObj := &TestGameObject{
			id:       "outside",
			name:     "Outside Object",
			position: Position{X: 100, Y: 100}, // Should be invalid (bounds are 0-99)
			active:   true,
			health:   100,
		}

		err = index.Insert(outsideObj)
		if err == nil {
			t.Error("Insert at (100,100) should fail but succeeded")
		}
	})
}
