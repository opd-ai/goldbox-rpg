package game

import (
	"testing"
)

func TestWorld_SpatialIndexingIntegration(t *testing.T) {
	world := NewWorldWithSize(200, 200, 25)

	// Test that world has spatial index initialized
	if world.SpatialIndex == nil {
		t.Fatal("World should have spatial index initialized")
	}

	// Create test objects
	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: Position{X: 50, Y: 50},
		},
	}

	npc1 := &NPC{
		Character: Character{
			ID:       "npc1",
			Name:     "Close NPC",
			Position: Position{X: 53, Y: 52},
		},
	}

	npc2 := &NPC{
		Character: Character{
			ID:       "npc2",
			Name:     "Far NPC",
			Position: Position{X: 150, Y: 150},
		},
	}

	enemy := &NPC{
		Character: Character{
			ID:       "enemy1",
			Name:     "Enemy",
			Position: Position{X: 55, Y: 55},
		},
	}

	// Test adding objects
	err := world.AddObject(player)
	if err != nil {
		t.Errorf("Failed to add player: %v", err)
	}

	err = world.AddObject(npc1)
	if err != nil {
		t.Errorf("Failed to add npc1: %v", err)
	}

	err = world.AddObject(npc2)
	if err != nil {
		t.Errorf("Failed to add npc2: %v", err)
	}

	err = world.AddObject(enemy)
	if err != nil {
		t.Errorf("Failed to add enemy: %v", err)
	}

	// Test GetObjectsInRange
	rect := Rectangle{MinX: 45, MinY: 45, MaxX: 60, MaxY: 60}
	objectsInRange := world.GetObjectsInRange(rect)
	if len(objectsInRange) != 3 {
		t.Errorf("Expected 3 objects in range, got %d", len(objectsInRange))
	}

	// Test GetObjectsInRadius
	center := Position{X: 50, Y: 50}
	objectsInRadius := world.GetObjectsInRadius(center, 10.0)
	if len(objectsInRadius) != 3 {
		t.Errorf("Expected 3 objects in radius 10, got %d", len(objectsInRadius))
	}

	// Test GetNearestObjects
	nearest := world.GetNearestObjects(center, 2)
	if len(nearest) != 2 {
		t.Errorf("Expected 2 nearest objects, got %d", len(nearest))
	}

	// First nearest should be the player itself
	if nearest[0].GetID() != "player1" {
		t.Errorf("Expected first nearest to be player1, got %s", nearest[0].GetID())
	}

	// Test UpdateObjectPosition
	err = world.UpdateObjectPosition("npc1", Position{X: 100, Y: 100})
	if err != nil {
		t.Errorf("Failed to update npc1 position: %v", err)
	}

	// Verify object moved
	objectsInRadius = world.GetObjectsInRadius(center, 10.0)
	if len(objectsInRadius) != 2 {
		t.Errorf("Expected 2 objects in radius after move, got %d", len(objectsInRadius))
	}

	// Test RemoveObject
	err = world.RemoveObject("enemy1")
	if err != nil {
		t.Errorf("Failed to remove enemy1: %v", err)
	}

	objectsInRadius = world.GetObjectsInRadius(center, 10.0)
	if len(objectsInRadius) != 1 {
		t.Errorf("Expected 1 object in radius after removal, got %d", len(objectsInRadius))
	}

	// Test spatial index stats
	stats := world.GetSpatialIndexStats()
	if stats == nil {
		t.Error("Expected spatial index stats, got nil")
	} else if stats.TotalObjects != 3 {
		t.Errorf("Expected 3 total objects in stats, got %d", stats.TotalObjects)
	}
}

func TestWorld_SpatialIndexingFallback(t *testing.T) {
	world := NewWorld()

	// Disable spatial index to test fallback
	world.SpatialIndex = nil

	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: Position{X: 50, Y: 50},
		},
	}

	npc := &NPC{
		Character: Character{
			ID:       "npc1",
			Name:     "Test NPC",
			Position: Position{X: 53, Y: 52},
		},
	}

	// Add objects (should use legacy method)
	world.Objects[player.GetID()] = player
	world.Objects[npc.GetID()] = npc

	// Test fallback methods
	rect := Rectangle{MinX: 45, MinY: 45, MaxX: 60, MaxY: 60}
	objectsInRange := world.GetObjectsInRange(rect)
	if len(objectsInRange) != 2 {
		t.Errorf("Expected 2 objects in range with fallback, got %d", len(objectsInRange))
	}

	center := Position{X: 50, Y: 50}
	objectsInRadius := world.GetObjectsInRadius(center, 10.0)
	if len(objectsInRadius) != 2 {
		t.Errorf("Expected 2 objects in radius with fallback, got %d", len(objectsInRadius))
	}

	nearest := world.GetNearestObjects(center, 1)
	if len(nearest) != 1 {
		t.Errorf("Expected 1 nearest object with fallback, got %d", len(nearest))
	}
}

func TestWorld_SpatialIndexingClone(t *testing.T) {
	original := NewWorldWithSize(100, 100, 20)

	// Add some objects
	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: Position{X: 25, Y: 25},
		},
	}

	npc := &NPC{
		Character: Character{
			ID:       "npc1",
			Name:     "Test NPC",
			Position: Position{X: 75, Y: 75},
		},
	}

	original.AddObject(player)
	original.AddObject(npc)

	// Clone the world
	clone := original.Clone()

	// Verify clone has spatial index
	if clone.SpatialIndex == nil {
		t.Error("Cloned world should have spatial index")
	}

	// Verify clone has same objects
	cloneObjects := clone.GetObjectsInRange(Rectangle{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100})
	if len(cloneObjects) != 2 {
		t.Errorf("Expected 2 objects in cloned world, got %d", len(cloneObjects))
	}

	// Verify modifications to clone don't affect original
	newNPC := &NPC{
		Character: Character{
			ID:       "npc2",
			Name:     "Clone NPC",
			Position: Position{X: 50, Y: 50},
		},
	}

	clone.AddObject(newNPC)

	originalObjects := original.GetObjectsInRange(Rectangle{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100})
	cloneObjects = clone.GetObjectsInRange(Rectangle{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100})

	if len(originalObjects) != 2 {
		t.Errorf("Original should still have 2 objects, got %d", len(originalObjects))
	}

	if len(cloneObjects) != 3 {
		t.Errorf("Clone should have 3 objects after addition, got %d", len(cloneObjects))
	}
}

func TestWorld_LegacyCompatibility(t *testing.T) {
	// Use NewWorldWithSize to ensure we have a spatial index for this test
	world := NewWorldWithSize(100, 100, 20)

	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: Position{X: 50, Y: 50},
		},
	}

	// Test that AddObject updates both legacy and advanced systems
	err := world.AddObject(player)
	if err != nil {
		t.Errorf("Failed to add object: %v", err)
	}

	// Check legacy spatial grid
	pos := Position{X: 50, Y: 50}
	legacyObjects := world.GetObjectsAt(pos)
	if len(legacyObjects) != 1 || legacyObjects[0].GetID() != "player1" {
		t.Error("Legacy GetObjectsAt should work")
	}

	// Check advanced spatial index
	if world.SpatialIndex != nil {
		advancedObjects := world.SpatialIndex.GetObjectsAt(pos)
		if len(advancedObjects) != 1 || advancedObjects[0].GetID() != "player1" {
			t.Error("Advanced spatial index GetObjectsAt should work")
		}

		// Test that both are synchronized
		if len(legacyObjects) != len(advancedObjects) {
			t.Error("Legacy and advanced spatial systems should be synchronized")
		}
	} else {
		t.Error("Expected spatial index to be initialized with NewWorldWithSize")
	}
}
