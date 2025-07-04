package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestItem_FromJSON_Valid tests FromJSON with valid JSON input
func TestItem_FromJSON_Valid(t *testing.T) {
	// Use Go struct field names for JSON keys, since only yaml tags are present
	jsonData := []byte(`{"ID":"sword_001","Name":"Sword","Type":"weapon","Damage":"1d6","AC":0,"Weight":3,"Value":50,"Properties":["sharp","metal"]}`)
	var item Item
	err := item.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}
	if item.ID != "sword_001" || item.Name != "Sword" || item.Type != "weapon" {
		t.Errorf("Unexpected item fields: %+v", item)
	}
	if len(item.Properties) != 2 || item.Properties[0] != "sharp" {
		t.Errorf("Properties not parsed correctly: %+v", item.Properties)
	}
}

// TestItem_FromJSON_Invalid tests FromJSON with invalid JSON
func TestItem_FromJSON_Invalid(t *testing.T) {
	var item Item
	err := item.FromJSON([]byte(`{"item_id":`))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestItem_GetDescription tests GetDescription returns correct format
func TestItem_GetDescription(t *testing.T) {
	item := Item{Name: "Potion", Type: "potion"}
	desc := item.GetDescription()
	if desc != "Potion (potion)" {
		t.Errorf("GetDescription = %q, want %q", desc, "Potion (potion)")
	}
}

// TestItem_Getters test all simple getter methods
func TestItem_Getters(t *testing.T) {
	item := Item{
		ID:         "item42",
		Name:       "Ring of Power",
		Type:       "ring",
		Properties: []string{"magic", "unique"},
	}
	if item.GetID() != "item42" {
		t.Errorf("GetID = %q, want %q", item.GetID(), "item42")
	}
	if item.GetName() != "Ring of Power" {
		t.Errorf("GetName = %q, want %q", item.GetName(), "Ring of Power")
	}
	if !reflect.DeepEqual(item.GetTags(), []string{"magic", "unique"}) {
		t.Errorf("GetTags = %+v, want %+v", item.GetTags(), []string{"magic", "unique"})
	}
}

// TestItem_GetHealth_AlwaysZero tests GetHealth always returns 0
func TestItem_GetHealth_AlwaysZero(t *testing.T) {
	item := Item{}
	if item.GetHealth() != 0 {
		t.Errorf("GetHealth = %d, want 0", item.GetHealth())
	}
}

// TestItem_IsActive_AlwaysTrue tests IsActive always returns true
func TestItem_IsActive_AlwaysTrue(t *testing.T) {
	item := Item{}
	if !item.IsActive() {
		t.Error("IsActive = false, want true")
	}
}

// TestItem_IsObstacle_AlwaysFalse tests IsObstacle always returns false
func TestItem_IsObstacle_AlwaysFalse(t *testing.T) {
	item := Item{}
	if item.IsObstacle() {
		t.Error("IsObstacle = true, want false")
	}
}

// TestItem_SetHealth_NoOp tests SetHealth does not panic or change state
func TestItem_SetHealth_NoOp(t *testing.T) {
	item := Item{}
	item.SetHealth(99) // Should not panic or error
}

// TestItem_SetPosition_NoOp tests SetPosition always returns nil
func TestItem_SetPosition_NoOp(t *testing.T) {
	item := Item{}
	err := item.SetPosition(Position{X: 1, Y: 2, Level: 3})
	if err != nil {
		t.Errorf("SetPosition returned error: %v", err)
	}
}

// TestItem_ToJSON_RoundTrip tests ToJSON and FromJSON together (table-driven)
func TestItem_ToJSON_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		item Item
	}{
		{"Minimal", Item{ID: "id1", Name: "A", Type: "misc"}},
		{"WithProperties", Item{ID: "id2", Name: "B", Type: "weapon", Properties: []string{"sharp"}}},
		{"WithAllFields", Item{ID: "id3", Name: "C", Type: "armor", Damage: "1d4", AC: 2, Weight: 5, Value: 10, Properties: []string{"heavy"}, Position: Position{X: 1, Y: 2, Level: 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.item.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON failed: %v", err)
			}
			var out Item
			err = json.Unmarshal(data, &out)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if out.ID != tt.item.ID || out.Name != tt.item.Name || out.Type != tt.item.Type {
				t.Errorf("Round-trip mismatch: got %+v, want %+v", out, tt.item)
			}
		})
	}
}

// TestItem_GetPosition_Default tests GetPosition always returns zero value
func TestItem_GetPosition_Default(t *testing.T) {
	item := Item{Position: Position{X: 5, Y: 5, Level: 1}}
	pos := item.GetPosition()
	if pos != (Position{}) {
		t.Errorf("GetPosition = %+v, want zero value", pos)
	}
}
