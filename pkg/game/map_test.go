package game

import (
	"encoding/json"
	"testing"
)

func TestGameMap_GetTile_ValidCoordinates_ReturnsCorrectTile(t *testing.T) {
	// Setup test map
	gameMap := &GameMap{
		Width:  3,
		Height: 2,
		Tiles: [][]MapTile{
			{
				{SpriteX: 0, SpriteY: 0, Walkable: true, Transparent: true},
				{SpriteX: 1, SpriteY: 0, Walkable: false, Transparent: false},
				{SpriteX: 2, SpriteY: 0, Walkable: true, Transparent: true},
			},
			{
				{SpriteX: 0, SpriteY: 1, Walkable: true, Transparent: false},
				{SpriteX: 1, SpriteY: 1, Walkable: false, Transparent: true},
				{SpriteX: 2, SpriteY: 1, Walkable: true, Transparent: true},
			},
		},
	}

	tests := []struct {
		name     string
		x, y     int
		expected *MapTile
	}{
		{
			name: "top-left corner",
			x:    0, y: 0,
			expected: &MapTile{SpriteX: 0, SpriteY: 0, Walkable: true, Transparent: true},
		},
		{
			name: "top-right corner",
			x:    2, y: 0,
			expected: &MapTile{SpriteX: 2, SpriteY: 0, Walkable: true, Transparent: true},
		},
		{
			name: "bottom-left corner",
			x:    0, y: 1,
			expected: &MapTile{SpriteX: 0, SpriteY: 1, Walkable: true, Transparent: false},
		},
		{
			name: "center tile",
			x:    1, y: 1,
			expected: &MapTile{SpriteX: 1, SpriteY: 1, Walkable: false, Transparent: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gameMap.GetTile(tt.x, tt.y)
			if result == nil {
				t.Fatalf("GetTile(%d, %d) returned nil, expected tile", tt.x, tt.y)
			}

			if result.SpriteX != tt.expected.SpriteX {
				t.Errorf("SpriteX = %d, expected %d", result.SpriteX, tt.expected.SpriteX)
			}
			if result.SpriteY != tt.expected.SpriteY {
				t.Errorf("SpriteY = %d, expected %d", result.SpriteY, tt.expected.SpriteY)
			}
			if result.Walkable != tt.expected.Walkable {
				t.Errorf("Walkable = %t, expected %t", result.Walkable, tt.expected.Walkable)
			}
			if result.Transparent != tt.expected.Transparent {
				t.Errorf("Transparent = %t, expected %t", result.Transparent, tt.expected.Transparent)
			}
		})
	}
}

func TestGameMap_GetTile_InvalidCoordinates_ReturnsNil(t *testing.T) {
	gameMap := &GameMap{
		Width:  2,
		Height: 2,
		Tiles: [][]MapTile{
			{{SpriteX: 0, SpriteY: 0}, {SpriteX: 1, SpriteY: 0}},
			{{SpriteX: 0, SpriteY: 1}, {SpriteX: 1, SpriteY: 1}},
		},
	}

	tests := []struct {
		name string
		x, y int
	}{
		{"negative x", -1, 0},
		{"negative y", 0, -1},
		{"both negative", -1, -1},
		{"x out of bounds", 2, 0},
		{"y out of bounds", 0, 2},
		{"both out of bounds", 3, 3},
		{"x equals width", 2, 1},
		{"y equals height", 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gameMap.GetTile(tt.x, tt.y)
			if result != nil {
				t.Errorf("GetTile(%d, %d) returned %+v, expected nil", tt.x, tt.y, result)
			}
		})
	}
}

func TestGameMap_GetTile_EmptyMap_ReturnsNil(t *testing.T) {
	gameMap := &GameMap{
		Width:  0,
		Height: 0,
		Tiles:  [][]MapTile{},
	}

	result := gameMap.GetTile(0, 0)
	if result != nil {
		t.Errorf("GetTile(0, 0) on empty map returned %+v, expected nil", result)
	}
}

func TestGameMap_MarshalJSON_ValidMap_ReturnsCorrectJSON(t *testing.T) {
	gameMap := &GameMap{
		Width:  2,
		Height: 1,
		Tiles: [][]MapTile{
			{
				{SpriteX: 10, SpriteY: 20, Walkable: true, Transparent: false},
				{SpriteX: 30, SpriteY: 40, Walkable: false, Transparent: true},
			},
		},
	}

	jsonData, err := gameMap.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() failed: %v", err)
	}

	// Parse the JSON to verify structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result JSON: %v", err)
	}

	// Check basic fields
	if width, ok := result["width"].(float64); !ok || int(width) != 2 {
		t.Errorf("width = %v, expected 2", result["width"])
	}
	if height, ok := result["height"].(float64); !ok || int(height) != 1 {
		t.Errorf("height = %v, expected 1", result["height"])
	}

	// Check tiles array
	tiles, ok := result["tiles"].([]interface{})
	if !ok {
		t.Fatalf("tiles field is not an array: %T", result["tiles"])
	}
	if len(tiles) != 1 {
		t.Errorf("tiles length = %d, expected 1", len(tiles))
	}

	// Check first row
	row0, ok := tiles[0].([]interface{})
	if !ok {
		t.Fatalf("first row is not an array: %T", tiles[0])
	}
	if len(row0) != 2 {
		t.Errorf("first row length = %d, expected 2", len(row0))
	}

	// Check first tile
	tile0, ok := row0[0].(map[string]interface{})
	if !ok {
		t.Fatalf("first tile is not an object: %T", row0[0])
	}
	if spriteX, ok := tile0["spriteX"].(float64); !ok || int(spriteX) != 10 {
		t.Errorf("first tile spriteX = %v, expected 10", tile0["spriteX"])
	}
	if walkable, ok := tile0["walkable"].(bool); !ok || !walkable {
		t.Errorf("first tile walkable = %v, expected true", tile0["walkable"])
	}

	// Check for getTile function
	if getTile, ok := result["getTile"].(string); !ok || getTile == "" {
		t.Errorf("getTile function not present or empty")
	}
}

func TestGameMap_MarshalJSON_EmptyMap_ReturnsValidJSON(t *testing.T) {
	gameMap := &GameMap{
		Width:  0,
		Height: 0,
		Tiles:  [][]MapTile{},
	}

	jsonData, err := gameMap.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() failed on empty map: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty map JSON: %v", err)
	}

	if width, ok := result["width"].(float64); !ok || int(width) != 0 {
		t.Errorf("width = %v, expected 0", result["width"])
	}
	if height, ok := result["height"].(float64); !ok || int(height) != 0 {
		t.Errorf("height = %v, expected 0", result["height"])
	}
	if tiles, ok := result["tiles"].([]interface{}); !ok || len(tiles) != 0 {
		t.Errorf("tiles = %v, expected empty array", result["tiles"])
	}
	if getTile, ok := result["getTile"].(string); !ok || getTile == "" {
		t.Errorf("getTile function not present or empty")
	}
}

func TestMapTile_FieldsAccess_AllFieldsReadable(t *testing.T) {
	tile := MapTile{
		SpriteX:     42,
		SpriteY:     24,
		Walkable:    true,
		Transparent: false,
	}

	if tile.SpriteX != 42 {
		t.Errorf("SpriteX = %d, expected 42", tile.SpriteX)
	}
	if tile.SpriteY != 24 {
		t.Errorf("SpriteY = %d, expected 24", tile.SpriteY)
	}
	if !tile.Walkable {
		t.Errorf("Walkable = %t, expected true", tile.Walkable)
	}
	if tile.Transparent {
		t.Errorf("Transparent = %t, expected false", tile.Transparent)
	}
}

func TestMapTile_JSONMarshaling_CorrectFieldNames(t *testing.T) {
	tile := MapTile{
		SpriteX:     1,
		SpriteY:     2,
		Walkable:    true,
		Transparent: false,
	}

	jsonData, err := json.Marshal(tile)
	if err != nil {
		t.Fatalf("Failed to marshal MapTile: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal MapTile JSON: %v", err)
	}

	expectedFields := []string{"spriteX", "spriteY", "walkable", "transparent"}
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("JSON missing field: %s", field)
		}
	}

	if len(result) != len(expectedFields) {
		t.Errorf("JSON has %d fields, expected %d", len(result), len(expectedFields))
	}
}
