package game

import (
	"fmt"
	"testing"
)

func TestSpatialIndex_BasicOperations(t *testing.T) {
	index := NewSpatialIndex(100, 100, 10)

	// Test empty index
	objects := index.GetObjectsAt(Position{X: 50, Y: 50})
	if len(objects) != 0 {
		t.Errorf("Empty index should return 0 objects, got %d", len(objects))
	}

	// Create test objects
	player := &TestGameObject{
		id:       "player1",
		name:     "Test Player",
		position: Position{X: 50, Y: 50},
	}

	npc := &TestGameObject{
		id:       "npc1",
		name:     "Test NPC",
		position: Position{X: 52, Y: 53},
	}

	enemy := &TestGameObject{
		id:       "enemy1",
		name:     "Test Enemy",
		position: Position{X: 80, Y: 80},
	}

	// Test insertion
	err := index.Insert(player)
	if err != nil {
		t.Errorf("Failed to insert player: %v", err)
	}

	err = index.Insert(npc)
	if err != nil {
		t.Errorf("Failed to insert npc: %v", err)
	}

	err = index.Insert(enemy)
	if err != nil {
		t.Errorf("Failed to insert enemy: %v", err)
	}

	// Test exact position query
	objects = index.GetObjectsAt(Position{X: 50, Y: 50})
	if len(objects) != 1 || objects[0].GetID() != "player1" {
		t.Errorf("Expected 1 object (player1) at (50,50), got %d objects", len(objects))
	}

	// Test range query
	rect := Rectangle{MinX: 48, MinY: 48, MaxX: 55, MaxY: 55}
	objects = index.GetObjectsInRange(rect)
	if len(objects) != 2 {
		t.Errorf("Expected 2 objects in range, got %d", len(objects))
	}

	// Test radius query
	objects = index.GetObjectsInRadius(Position{X: 50, Y: 50}, 5.0)
	if len(objects) != 2 {
		t.Errorf("Expected 2 objects in radius 5, got %d", len(objects))
	}

	// Test removal
	err = index.Remove("npc1")
	if err != nil {
		t.Errorf("Failed to remove npc1: %v", err)
	}

	objects = index.GetObjectsInRadius(Position{X: 50, Y: 50}, 5.0)
	if len(objects) != 1 || objects[0].GetID() != "player1" {
		t.Errorf("Expected 1 object after removal, got %d", len(objects))
	}
}

func TestSpatialIndex_Update(t *testing.T) {
	index := NewSpatialIndex(100, 100, 10)

	obj := &TestGameObject{
		id:       "obj1",
		name:     "Test Object",
		position: Position{X: 10, Y: 10},
	}

	// Insert object
	err := index.Insert(obj)
	if err != nil {
		t.Errorf("Failed to insert object: %v", err)
	}

	// Verify initial position
	objects := index.GetObjectsAt(Position{X: 10, Y: 10})
	if len(objects) != 1 {
		t.Errorf("Expected 1 object at initial position, got %d", len(objects))
	}

	// Update position
	obj.position = Position{X: 90, Y: 90}
	err = index.Update("obj1", Position{X: 90, Y: 90})
	if err != nil {
		t.Errorf("Failed to update object position: %v", err)
	}

	// Verify old position is empty
	objects = index.GetObjectsAt(Position{X: 10, Y: 10})
	if len(objects) != 0 {
		t.Errorf("Expected 0 objects at old position, got %d", len(objects))
	}

	// Verify new position has object
	objects = index.GetObjectsAt(Position{X: 90, Y: 90})
	if len(objects) != 1 || objects[0].GetID() != "obj1" {
		t.Errorf("Expected 1 object (obj1) at new position, got %d objects", len(objects))
	}
}

func TestSpatialIndex_NearestObjects(t *testing.T) {
	index := NewSpatialIndex(100, 100, 10)

	// Create objects at various distances
	objects := []*TestGameObject{
		{id: "obj1", position: Position{X: 50, Y: 50}}, // Distance 0
		{id: "obj2", position: Position{X: 53, Y: 50}}, // Distance 3
		{id: "obj3", position: Position{X: 50, Y: 55}}, // Distance 5
		{id: "obj4", position: Position{X: 60, Y: 60}}, // Distance ~14.14
		{id: "obj5", position: Position{X: 80, Y: 80}}, // Distance ~42.43
	}

	for _, obj := range objects {
		err := index.Insert(obj)
		if err != nil {
			t.Errorf("Failed to insert object %s: %v", obj.id, err)
		}
	}

	// Test getting 3 nearest objects
	center := Position{X: 50, Y: 50}
	nearest := index.GetNearestObjects(center, 3)

	if len(nearest) != 3 {
		t.Errorf("Expected 3 nearest objects, got %d", len(nearest))
	}

	// Verify order (should be obj1, obj2, obj3)
	expectedOrder := []string{"obj1", "obj2", "obj3"}
	for i, obj := range nearest {
		if obj.GetID() != expectedOrder[i] {
			t.Errorf("Expected object %s at position %d, got %s", expectedOrder[i], i, obj.GetID())
		}
	}
}

func TestSpatialIndex_Performance(t *testing.T) {
	index := NewSpatialIndex(1000, 1000, 50)

	// Insert many objects
	numObjects := 1000
	for i := 0; i < numObjects; i++ {
		obj := &TestGameObject{
			id:       fmt.Sprintf("obj%d", i),
			position: Position{X: i % 1000, Y: (i * 7) % 1000},
		}

		err := index.Insert(obj)
		if err != nil {
			t.Errorf("Failed to insert object %d: %v", i, err)
		}
	}

	// Test range query performance
	rect := Rectangle{MinX: 100, MinY: 100, MaxX: 200, MaxY: 200}
	objects := index.GetObjectsInRange(rect)

	// Should find objects efficiently without O(n) scan
	t.Logf("Found %d objects in range query", len(objects))

	// Test radius query performance
	objects = index.GetObjectsInRadius(Position{X: 500, Y: 500}, 50.0)
	t.Logf("Found %d objects in radius query", len(objects))

	// Test stats
	stats := index.GetStats()
	t.Logf("Spatial index stats: Objects=%d, Nodes=%d, MaxDepth=%d, LeafNodes=%d",
		stats.TotalObjects, stats.TotalNodes, stats.MaxDepth, stats.LeafNodes)

	if stats.TotalObjects != numObjects {
		t.Errorf("Expected %d total objects in stats, got %d", numObjects, stats.TotalObjects)
	}
}

func TestSpatialIndex_EdgeCases(t *testing.T) {
	index := NewSpatialIndex(100, 100, 10)

	// Test inserting out of bounds
	obj := &TestGameObject{
		id:       "out_of_bounds",
		position: Position{X: 150, Y: 150},
	}

	err := index.Insert(obj)
	if err == nil {
		t.Error("Expected error when inserting out of bounds object")
	}

	// Test removing non-existent object
	err = index.Remove("non_existent")
	if err == nil {
		t.Error("Expected error when removing non-existent object")
	}

	// Test updating non-existent object
	err = index.Update("non_existent", Position{X: 50, Y: 50})
	if err == nil {
		t.Error("Expected error when updating non-existent object")
	}

	// Test large radius query
	objects := index.GetObjectsInRadius(Position{X: 50, Y: 50}, 1000.0)
	if len(objects) != 0 {
		t.Errorf("Expected 0 objects for large radius on empty index, got %d", len(objects))
	}
}

func TestSpatialIndex_Clear(t *testing.T) {
	index := NewSpatialIndex(100, 100, 10)

	// Insert some objects
	for i := 0; i < 10; i++ {
		obj := &TestGameObject{
			id:       fmt.Sprintf("obj%d", i),
			position: Position{X: i * 10, Y: i * 10},
		}
		index.Insert(obj)
	}

	// Verify objects exist
	stats := index.GetStats()
	if stats.TotalObjects != 10 {
		t.Errorf("Expected 10 objects before clear, got %d", stats.TotalObjects)
	}

	// Clear index
	index.Clear()

	// Verify all objects removed
	stats = index.GetStats()
	if stats.TotalObjects != 0 {
		t.Errorf("Expected 0 objects after clear, got %d", stats.TotalObjects)
	}

	objects := index.GetObjectsInRange(Rectangle{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100})
	if len(objects) != 0 {
		t.Errorf("Expected 0 objects in range after clear, got %d", len(objects))
	}
}

// BenchmarkGetObjectsInRadius tests the performance of radius queries
func BenchmarkGetObjectsInRadius(b *testing.B) {
	index := NewSpatialIndex(1000, 1000, 50)

	// Create many objects for realistic performance testing
	numObjects := 1000
	for i := 0; i < numObjects; i++ {
		obj := &TestGameObject{
			id:       fmt.Sprintf("obj_%d", i),
			position: Position{X: i % 100 * 10, Y: i / 100 * 10},
			health:   100,
			active:   true,
		}
		err := index.Insert(obj)
		if err != nil {
			b.Fatalf("Failed to insert object %d: %v", i, err)
		}
	}

	center := Position{X: 500, Y: 500}
	radius := 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = index.GetObjectsInRadius(center, radius)
	}
}

// TestGameObject is a simple implementation for testing
type TestGameObject struct {
	id          string
	name        string
	description string
	position    Position
	active      bool
	health      int
	tags        []string
}

func (t *TestGameObject) GetID() string                  { return t.id }
func (t *TestGameObject) GetName() string                { return t.name }
func (t *TestGameObject) GetDescription() string         { return t.description }
func (t *TestGameObject) GetPosition() Position          { return t.position }
func (t *TestGameObject) SetPosition(pos Position) error { t.position = pos; return nil }
func (t *TestGameObject) IsActive() bool                 { return t.active }
func (t *TestGameObject) GetTags() []string              { return t.tags }
func (t *TestGameObject) ToJSON() ([]byte, error)        { return nil, nil }
func (t *TestGameObject) FromJSON([]byte) error          { return nil }
func (t *TestGameObject) GetHealth() int                 { return t.health }
func (t *TestGameObject) SetHealth(h int)                { t.health = h }
func (t *TestGameObject) IsObstacle() bool               { return false }
