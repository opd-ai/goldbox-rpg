package game

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// TestNewWorld tests the creation of a new World instance
func TestNewWorld(t *testing.T) {
	world := NewWorld()

	// Verify the world is not nil
	if world == nil {
		t.Fatal("NewWorld() returned nil")
	}

	// Verify all maps are initialized
	if world.Objects == nil {
		t.Error("Objects map should be initialized")
	}

	if world.Players == nil {
		t.Error("Players map should be initialized")
	}

	if world.NPCs == nil {
		t.Error("NPCs map should be initialized")
	}

	if world.SpatialGrid == nil {
		t.Error("SpatialGrid map should be initialized")
	}

	// Verify maps are empty
	if len(world.Objects) != 0 {
		t.Error("Objects map should be empty initially")
	}

	if len(world.Players) != 0 {
		t.Error("Players map should be empty initially")
	}

	if len(world.NPCs) != 0 {
		t.Error("NPCs map should be empty initially")
	}

	if len(world.SpatialGrid) != 0 {
		t.Error("SpatialGrid map should be empty initially")
	}

	// Verify default values
	if world.Width != 0 {
		t.Error("Width should be 0 by default")
	}

	if world.Height != 0 {
		t.Error("Height should be 0 by default")
	}
}

// TestWorld_AddObject tests adding objects to the world
func TestWorld_AddObject(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name    string
		object  GameObject
		wantErr bool
		errMsg  string
	}{
		{
			name: "Add new player",
			object: &Player{
				Character: Character{
					ID:       "player1",
					Name:     "Test Player",
					Position: Position{X: 5, Y: 10},
				},
			},
			wantErr: false,
		},
		{
			name: "Add NPC",
			object: &NPC{
				Character: Character{
					ID:       "npc1",
					Name:     "Test NPC",
					Position: Position{X: 3, Y: 7},
				},
			},
			wantErr: false,
		},
		{
			name: "Add duplicate object",
			object: &Player{
				Character: Character{
					ID:       "player1", // Same ID as first test
					Name:     "Duplicate Player",
					Position: Position{X: 1, Y: 1},
				},
			},
			wantErr: true,
			errMsg:  "object with ID player1 already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := world.AddObject(tt.object)

			if tt.wantErr {
				if err == nil {
					t.Error("AddObject() should have returned an error")
				} else if err.Error() != tt.errMsg {
					t.Errorf("AddObject() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("AddObject() unexpected error = %v", err)
				}

				// Verify object was added to Objects map
				if _, exists := world.Objects[tt.object.GetID()]; !exists {
					t.Error("Object should be added to Objects map")
				}

				// Verify object was added to SpatialGrid
				pos := tt.object.GetPosition()
				found := false
				for _, id := range world.SpatialGrid[pos] {
					if id == tt.object.GetID() {
						found = true
						break
					}
				}
				if !found {
					t.Error("Object should be added to SpatialGrid")
				}
			}
		})
	}
}

// TestWorld_GetObjectsAt tests retrieving objects at a specific position
func TestWorld_GetObjectsAt(t *testing.T) {
	world := NewWorld()
	pos := Position{X: 5, Y: 5}

	// Test empty position
	t.Run("Empty position", func(t *testing.T) {
		objects := world.GetObjectsAt(pos)
		if len(objects) != 0 {
			t.Errorf("GetObjectsAt() returned %d objects, want 0", len(objects))
		}
	})

	// Add objects at the position
	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: pos,
		},
	}

	npc := &NPC{
		Character: Character{
			ID:       "npc1",
			Name:     "Test NPC",
			Position: pos,
		},
	}

	world.AddObject(player)
	world.AddObject(npc)

	t.Run("Position with objects", func(t *testing.T) {
		objects := world.GetObjectsAt(pos)
		if len(objects) != 2 {
			t.Errorf("GetObjectsAt() returned %d objects, want 2", len(objects))
		}

		// Verify correct objects are returned
		objectIDs := make(map[string]bool)
		for _, obj := range objects {
			objectIDs[obj.GetID()] = true
		}

		if !objectIDs["player1"] {
			t.Error("player1 should be in returned objects")
		}

		if !objectIDs["npc1"] {
			t.Error("npc1 should be in returned objects")
		}
	})

	t.Run("Different position", func(t *testing.T) {
		differentPos := Position{X: 10, Y: 10}
		objects := world.GetObjectsAt(differentPos)
		if len(objects) != 0 {
			t.Errorf("GetObjectsAt() returned %d objects at different position, want 0", len(objects))
		}
	})
}

// TestWorld_isPositionWithinBounds tests boundary checking
func TestWorld_isPositionWithinBounds(t *testing.T) {
	world := NewWorld()
	world.Width = 10
	world.Height = 10

	tests := []struct {
		name     string
		position Position
		want     bool
	}{
		{
			name:     "Valid position (0,0)",
			position: Position{X: 0, Y: 0},
			want:     true,
		},
		{
			name:     "Valid position (5,5)",
			position: Position{X: 5, Y: 5},
			want:     true,
		},
		{
			name:     "Valid position (9,9)",
			position: Position{X: 9, Y: 9},
			want:     true,
		},
		{
			name:     "Invalid position - negative X",
			position: Position{X: -1, Y: 5},
			want:     false,
		},
		{
			name:     "Invalid position - negative Y",
			position: Position{X: 5, Y: -1},
			want:     false,
		},
		{
			name:     "Invalid position - X too large",
			position: Position{X: 10, Y: 5},
			want:     false,
		},
		{
			name:     "Invalid position - Y too large",
			position: Position{X: 5, Y: 10},
			want:     false,
		},
		{
			name:     "Invalid position - both out of bounds",
			position: Position{X: 15, Y: 15},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := world.isPositionWithinBounds(tt.position)
			if got != tt.want {
				t.Errorf("isPositionWithinBounds() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWorld_ValidateMove tests move validation logic
func TestWorld_ValidateMove(t *testing.T) {
	world := NewWorld()
	world.Width = 10
	world.Height = 10

	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: Position{X: 5, Y: 5},
		},
	}

	// Add an obstacle
	obstacle := &MockObstacle{
		id:         "obstacle1",
		position:   Position{X: 3, Y: 3},
		isObstacle: true,
	}
	world.AddObject(obstacle)

	tests := []struct {
		name    string
		player  *Player
		newPos  Position
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid move within bounds",
			player:  player,
			newPos:  Position{X: 6, Y: 6},
			wantErr: false,
		},
		{
			name:    "Invalid move - out of bounds (negative)",
			player:  player,
			newPos:  Position{X: -1, Y: 5},
			wantErr: true,
			errMsg:  "position out of bounds",
		},
		{
			name:    "Invalid move - out of bounds (too large)",
			player:  player,
			newPos:  Position{X: 10, Y: 5},
			wantErr: true,
			errMsg:  "position out of bounds",
		},
		{
			name:    "Invalid move - position occupied by obstacle",
			player:  player,
			newPos:  Position{X: 3, Y: 3},
			wantErr: true,
			errMsg:  "position occupied by an obstacle",
		},
		{
			name:    "Valid move to edge position",
			player:  player,
			newPos:  Position{X: 0, Y: 0},
			wantErr: false,
		},
		{
			name:    "Valid move to opposite edge",
			player:  player,
			newPos:  Position{X: 9, Y: 9},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := world.ValidateMove(tt.player, tt.newPos)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateMove() should have returned an error")
				} else if err.Error() != tt.errMsg {
					t.Errorf("ValidateMove() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateMove() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestWorld_Clone tests deep cloning of World state
func TestWorld_Clone(t *testing.T) {
	original := NewWorld()
	original.Width = 20
	original.Height = 15
	original.CurrentTime = GameTime{
		RealTime:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		GameTicks: 53,
		TimeScale: 1.0,
	}

	// Add objects to original
	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: Position{X: 5, Y: 5},
		},
	}
	original.AddObject(player)

	npc := &NPC{
		Character: Character{
			ID:       "npc1",
			Name:     "Test NPC",
			Position: Position{X: 10, Y: 10},
		},
	}
	original.AddObject(npc)

	// Add to Players and NPCs maps
	original.Players["player1"] = player
	original.NPCs["npc1"] = npc

	// Clone the world
	clone := original.Clone()

	// Test that clone is not nil
	if clone == nil {
		t.Fatal("Clone() returned nil")
	}

	// Test that clone is a different instance
	if clone == original {
		t.Error("Clone() returned same instance, not a copy")
	}

	// Test basic fields are copied
	if clone.Width != original.Width {
		t.Errorf("Clone Width = %v, want %v", clone.Width, original.Width)
	}

	if clone.Height != original.Height {
		t.Errorf("Clone Height = %v, want %v", clone.Height, original.Height)
	}

	if !reflect.DeepEqual(clone.CurrentTime, original.CurrentTime) {
		t.Errorf("Clone CurrentTime = %v, want %v", clone.CurrentTime, original.CurrentTime)
	}

	// Test that maps are copied (not same reference)
	if &clone.Objects == &original.Objects {
		t.Error("Objects map should be a copy, not same reference")
	}

	if &clone.Players == &original.Players {
		t.Error("Players map should be a copy, not same reference")
	}

	if &clone.NPCs == &original.NPCs {
		t.Error("NPCs map should be a copy, not same reference")
	}

	if &clone.SpatialGrid == &original.SpatialGrid {
		t.Error("SpatialGrid map should be a copy, not same reference")
	}

	// Test that map contents are the same
	if len(clone.Objects) != len(original.Objects) {
		t.Errorf("Clone Objects length = %v, want %v", len(clone.Objects), len(original.Objects))
	}

	if len(clone.Players) != len(original.Players) {
		t.Errorf("Clone Players length = %v, want %v", len(clone.Players), len(original.Players))
	}

	if len(clone.NPCs) != len(original.NPCs) {
		t.Errorf("Clone NPCs length = %v, want %v", len(clone.NPCs), len(original.NPCs))
	}

	// Test that SpatialGrid is properly copied
	for pos, ids := range original.SpatialGrid {
		cloneIds, exists := clone.SpatialGrid[pos]
		if !exists {
			t.Errorf("Position %v missing in clone SpatialGrid", pos)
			continue
		}

		if !reflect.DeepEqual(ids, cloneIds) {
			t.Errorf("SpatialGrid at %v = %v, want %v", pos, cloneIds, ids)
		}

		// Verify it's a copy, not same slice
		if len(ids) > 0 && &ids[0] == &cloneIds[0] {
			t.Error("SpatialGrid slices should be copies, not same reference")
		}
	}
}

// TestWorld_Update tests updating world state
func TestWorld_Update(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: Position{X: 1, Y: 1},
		},
	}

	npc := &NPC{
		Character: Character{
			ID:       "npc1",
			Name:     "Test NPC",
			Position: Position{X: 2, Y: 2},
		},
	}

	tests := []struct {
		name      string
		updates   map[string]interface{}
		wantErr   bool
		errMsg    string
		checkFunc func(*World) error
	}{
		{
			name: "Update objects",
			updates: map[string]interface{}{
				"objects": map[string]GameObject{
					"player1": player,
				},
			},
			wantErr: false,
			checkFunc: func(w *World) error {
				if _, exists := w.Objects["player1"]; !exists {
					return fmt.Errorf("player1 should exist in Objects")
				}
				return nil
			},
		},
		{
			name: "Update players",
			updates: map[string]interface{}{
				"players": map[string]*Player{
					"player1": player,
				},
			},
			wantErr: false,
			checkFunc: func(w *World) error {
				if _, exists := w.Players["player1"]; !exists {
					return fmt.Errorf("player1 should exist in Players")
				}
				return nil
			},
		},
		{
			name: "Update NPCs",
			updates: map[string]interface{}{
				"npcs": map[string]*NPC{
					"npc1": npc,
				},
			},
			wantErr: false,
			checkFunc: func(w *World) error {
				if _, exists := w.NPCs["npc1"]; !exists {
					return fmt.Errorf("npc1 should exist in NPCs")
				}
				return nil
			},
		},
		{
			name: "Update current time",
			updates: map[string]interface{}{
				"current_time": GameTime{
					RealTime:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					GameTicks: 100,
					TimeScale: 1.0,
				},
			},
			wantErr: false,
			checkFunc: func(w *World) error {
				expected := GameTime{
					RealTime:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					GameTicks: 100,
					TimeScale: 1.0,
				}
				if !reflect.DeepEqual(w.CurrentTime, expected) {
					return fmt.Errorf("CurrentTime = %v, want %v", w.CurrentTime, expected)
				}
				return nil
			},
		},
		{
			name: "Invalid update key",
			updates: map[string]interface{}{
				"invalid_key": "some_value",
			},
			wantErr: true,
			errMsg:  "unknown update key: invalid_key",
		},
		{
			name: "Multiple valid updates",
			updates: map[string]interface{}{
				"objects": map[string]GameObject{
					"player2": player,
				},
				"current_time": GameTime{
					RealTime:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					GameTicks: 150,
					TimeScale: 1.0,
				},
			},
			wantErr: false,
			checkFunc: func(w *World) error {
				if _, exists := w.Objects["player2"]; !exists {
					return fmt.Errorf("player2 should exist in Objects")
				}
				expected := GameTime{
					RealTime:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					GameTicks: 150,
					TimeScale: 1.0,
				}
				if !reflect.DeepEqual(w.CurrentTime, expected) {
					return fmt.Errorf("CurrentTime = %v, want %v", w.CurrentTime, expected)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := world.Update(tt.updates)

			if tt.wantErr {
				if err == nil {
					t.Error("Update() should have returned an error")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Update() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Update() unexpected error = %v", err)
				}

				if tt.checkFunc != nil {
					if checkErr := tt.checkFunc(world); checkErr != nil {
						t.Error(checkErr)
					}
				}
			}
		})
	}
}

// TestWorld_Update_SpatialIndexIntegration tests that Update method updates both SpatialGrid and SpatialIndex
func TestWorld_Update_SpatialIndexIntegration(t *testing.T) {
	// Use NewWorldWithSize to ensure we have spatial index initialized
	world := NewWorldWithSize(100, 100, 25)

	if world.SpatialIndex == nil {
		t.Fatal("World should have spatial index initialized")
	}

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
			Position: Position{X: 75, Y: 75},
		},
	}

	// Test Update method with objects
	updates := map[string]interface{}{
		"objects": map[string]GameObject{
			"player1": player,
			"npc1":    npc,
		},
	}

	err := world.Update(updates)
	if err != nil {
		t.Errorf("Update() failed: %v", err)
	}

	// Verify objects are in the Objects map
	if _, exists := world.Objects["player1"]; !exists {
		t.Error("player1 should exist in Objects map")
	}
	if _, exists := world.Objects["npc1"]; !exists {
		t.Error("npc1 should exist in Objects map")
	}

	// Verify legacy SpatialGrid is updated
	playerPos := Position{X: 50, Y: 50}
	legacyObjects := world.GetObjectsAt(playerPos)
	found := false
	for _, obj := range legacyObjects {
		if obj.GetID() == "player1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("player1 should be found via legacy GetObjectsAt")
	}

	// Verify advanced SpatialIndex is updated
	if world.SpatialIndex != nil {
		advancedObjects := world.SpatialIndex.GetObjectsAt(playerPos)
		found = false
		for _, obj := range advancedObjects {
			if obj.GetID() == "player1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("player1 should be found via advanced SpatialIndex.GetObjectsAt")
		}

		// Test range query on spatial index
		rect := Rectangle{MinX: 40, MinY: 40, MaxX: 80, MaxY: 80}
		rangeObjects := world.SpatialIndex.GetObjectsInRange(rect)
		if len(rangeObjects) != 2 {
			t.Errorf("Expected 2 objects in range via SpatialIndex, got %d", len(rangeObjects))
		}

		// Verify both objects are found
		foundPlayer, foundNPC := false, false
		for _, obj := range rangeObjects {
			if obj.GetID() == "player1" {
				foundPlayer = true
			}
			if obj.GetID() == "npc1" {
				foundNPC = true
			}
		}
		if !foundPlayer {
			t.Error("player1 should be found in range query")
		}
		if !foundNPC {
			t.Error("npc1 should be found in range query")
		}
	}
}

// TestWorld_Serialize tests world serialization
func TestWorld_Serialize(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Character: Character{
			ID:       "player1",
			Name:     "Test Player",
			Position: Position{X: 5, Y: 5},
		},
	}
	world.AddObject(player)

	serialized := world.Serialize()

	// Test that serialized data contains expected keys
	if _, exists := serialized["objects"]; !exists {
		t.Error("Serialized data should contain 'objects' key")
	}

	// Test that objects are properly serialized
	if objects, ok := serialized["objects"].(map[string]GameObject); ok {
		if _, exists := objects["player1"]; !exists {
			t.Error("Serialized objects should contain player1")
		}
	} else {
		t.Error("Objects should be of type map[string]GameObject")
	}
}

// MockObstacle is a test helper that implements GameObject interface
type MockObstacle struct {
	id         string
	position   Position
	isObstacle bool
}

func (m *MockObstacle) GetID() string {
	return m.id
}

func (m *MockObstacle) GetName() string {
	return "Mock Obstacle"
}

func (m *MockObstacle) GetDescription() string {
	return "A test obstacle"
}

func (m *MockObstacle) GetPosition() Position {
	return m.position
}

func (m *MockObstacle) SetPosition(pos Position) error {
	m.position = pos
	return nil
}

func (m *MockObstacle) IsActive() bool {
	return true
}

func (m *MockObstacle) GetTags() []string {
	return []string{"obstacle", "test"}
}

func (m *MockObstacle) ToJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"id":"%s","position":{"x":%d,"y":%d}}`, m.id, m.position.X, m.position.Y)), nil
}

func (m *MockObstacle) FromJSON(data []byte) error {
	return nil // Simple implementation for testing
}

func (m *MockObstacle) GetHealth() int {
	return 100
}

func (m *MockObstacle) SetHealth(health int) {
	// No-op for obstacle
}

func (m *MockObstacle) IsObstacle() bool {
	return m.isObstacle
}

func (m *MockObstacle) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"id":       m.id,
		"position": m.position,
		"obstacle": m.isObstacle,
	}
}
