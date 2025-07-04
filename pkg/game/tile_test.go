package game

import (
	"testing"
)

// TestTileType_Constants tests that all tile type constants have the expected values
func TestTileType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		tileType TileType
		expected int
	}{
		{
			name:     "TileFloor has value 0",
			tileType: TileFloor,
			expected: 0,
		},
		{
			name:     "TileWall has value 1",
			tileType: TileWall,
			expected: 1,
		},
		{
			name:     "TileDoor has value 2",
			tileType: TileDoor,
			expected: 2,
		},
		{
			name:     "TileWater has value 3",
			tileType: TileWater,
			expected: 3,
		},
		{
			name:     "TileLava has value 4",
			tileType: TileLava,
			expected: 4,
		},
		{
			name:     "TilePit has value 5",
			tileType: TilePit,
			expected: 5,
		},
		{
			name:     "TileStairs has value 6",
			tileType: TileStairs,
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := int(tt.tileType)
			if result != tt.expected {
				t.Errorf("TileType constant = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestRGB_StructFields tests RGB struct field assignment and initialization
func TestRGB_StructFields(t *testing.T) {
	tests := []struct {
		name     string
		rgb      RGB
		expectedR uint8
		expectedG uint8
		expectedB uint8
	}{
		{
			name:     "RGB with zero values",
			rgb:      RGB{R: 0, G: 0, B: 0},
			expectedR: 0,
			expectedG: 0,
			expectedB: 0,
		},
		{
			name:     "RGB with max values",
			rgb:      RGB{R: 255, G: 255, B: 255},
			expectedR: 255,
			expectedG: 255,
			expectedB: 255,
		},
		{
			name:     "RGB with mixed values",
			rgb:      RGB{R: 128, G: 64, B: 192},
			expectedR: 128,
			expectedG: 64,
			expectedB: 192,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rgb.R != tt.expectedR {
				t.Errorf("RGB.R = %v, want %v", tt.rgb.R, tt.expectedR)
			}
			if tt.rgb.G != tt.expectedG {
				t.Errorf("RGB.G = %v, want %v", tt.rgb.G, tt.expectedG)
			}
			if tt.rgb.B != tt.expectedB {
				t.Errorf("RGB.B = %v, want %v", tt.rgb.B, tt.expectedB)
			}
		})
	}
}

// TestRGB_DefaultValues tests RGB struct with default zero values
func TestRGB_DefaultValues(t *testing.T) {
	var rgb RGB
	
	if rgb.R != 0 {
		t.Errorf("Default RGB.R = %v, want 0", rgb.R)
	}
	if rgb.G != 0 {
		t.Errorf("Default RGB.G = %v, want 0", rgb.G)
	}
	if rgb.B != 0 {
		t.Errorf("Default RGB.B = %v, want 0", rgb.B)
	}
}

// TestTile_StructFields tests Tile struct field assignment and initialization
func TestTile_StructFields(t *testing.T) {
	properties := make(map[string]interface{})
	properties["test_key"] = "test_value"
	properties["numeric_key"] = 42
	
	tile := Tile{
		Type:        TileFloor,
		Walkable:    true,
		Transparent: true,
		Properties:  properties,
		Sprite:      "floor_sprite",
		Color:       RGB{100, 150, 200},
		BlocksSight: false,
		Dangerous:   false,
		DamageType:  "",
		Damage:      0,
	}
	
	// Test basic field assignment
	if tile.Type != TileFloor {
		t.Errorf("Tile.Type = %v, want %v", tile.Type, TileFloor)
	}
	
	if tile.Walkable != true {
		t.Errorf("Tile.Walkable = %v, want %v", tile.Walkable, true)
	}
	
	if tile.Transparent != true {
		t.Errorf("Tile.Transparent = %v, want %v", tile.Transparent, true)
	}
	
	if tile.Sprite != "floor_sprite" {
		t.Errorf("Tile.Sprite = %v, want %v", tile.Sprite, "floor_sprite")
	}
	
	if tile.Color.R != 100 || tile.Color.G != 150 || tile.Color.B != 200 {
		t.Errorf("Tile.Color = %v, want RGB{100, 150, 200}", tile.Color)
	}
	
	if tile.BlocksSight != false {
		t.Errorf("Tile.BlocksSight = %v, want %v", tile.BlocksSight, false)
	}
	
	if tile.Dangerous != false {
		t.Errorf("Tile.Dangerous = %v, want %v", tile.Dangerous, false)
	}
	
	if tile.DamageType != "" {
		t.Errorf("Tile.DamageType = %v, want empty string", tile.DamageType)
	}
	
	if tile.Damage != 0 {
		t.Errorf("Tile.Damage = %v, want %v", tile.Damage, 0)
	}
	
	// Test properties map
	if len(tile.Properties) != 2 {
		t.Errorf("Tile.Properties length = %v, want %v", len(tile.Properties), 2)
	}
	
	if tile.Properties["test_key"] != "test_value" {
		t.Errorf("Tile.Properties[\"test_key\"] = %v, want %v", tile.Properties["test_key"], "test_value")
	}
	
	if tile.Properties["numeric_key"] != 42 {
		t.Errorf("Tile.Properties[\"numeric_key\"] = %v, want %v", tile.Properties["numeric_key"], 42)
	}
}

// TestTile_DefaultValues tests Tile struct with default zero values
func TestTile_DefaultValues(t *testing.T) {
	var tile Tile
	
	// Test default values
	if tile.Type != TileFloor { // 0 value should be TileFloor
		t.Errorf("Default Tile.Type = %v, want %v", tile.Type, TileFloor)
	}
	
	if tile.Walkable != false {
		t.Errorf("Default Tile.Walkable = %v, want false", tile.Walkable)
	}
	
	if tile.Transparent != false {
		t.Errorf("Default Tile.Transparent = %v, want false", tile.Transparent)
	}
	
	if tile.Properties != nil {
		t.Errorf("Default Tile.Properties = %v, want nil", tile.Properties)
	}
	
	if tile.Sprite != "" {
		t.Errorf("Default Tile.Sprite = %v, want empty string", tile.Sprite)
	}
	
	if tile.Color.R != 0 || tile.Color.G != 0 || tile.Color.B != 0 {
		t.Errorf("Default Tile.Color = %v, want RGB{0, 0, 0}", tile.Color)
	}
	
	if tile.BlocksSight != false {
		t.Errorf("Default Tile.BlocksSight = %v, want false", tile.BlocksSight)
	}
	
	if tile.Dangerous != false {
		t.Errorf("Default Tile.Dangerous = %v, want false", tile.Dangerous)
	}
	
	if tile.DamageType != "" {
		t.Errorf("Default Tile.DamageType = %v, want empty string", tile.DamageType)
	}
	
	if tile.Damage != 0 {
		t.Errorf("Default Tile.Damage = %v, want 0", tile.Damage)
	}
}

// TestNewFloorTile tests the NewFloorTile constructor function
func TestNewFloorTile(t *testing.T) {
	tile := NewFloorTile()
	
	// Test that all properties are set correctly for a floor tile
	if tile.Type != TileFloor {
		t.Errorf("NewFloorTile().Type = %v, want %v", tile.Type, TileFloor)
	}
	
	if tile.Walkable != true {
		t.Errorf("NewFloorTile().Walkable = %v, want true", tile.Walkable)
	}
	
	if tile.Transparent != true {
		t.Errorf("NewFloorTile().Transparent = %v, want true", tile.Transparent)
	}
	
	if tile.Properties == nil {
		t.Errorf("NewFloorTile().Properties = nil, want initialized map")
	}
	
	if len(tile.Properties) != 0 {
		t.Errorf("NewFloorTile().Properties length = %v, want 0", len(tile.Properties))
	}
	
	if tile.Sprite != "" {
		t.Errorf("NewFloorTile().Sprite = %v, want empty string", tile.Sprite)
	}
	
	expectedColor := RGB{200, 200, 200}
	if tile.Color != expectedColor {
		t.Errorf("NewFloorTile().Color = %v, want %v", tile.Color, expectedColor)
	}
	
	if tile.BlocksSight != false {
		t.Errorf("NewFloorTile().BlocksSight = %v, want false", tile.BlocksSight)
	}
	
	if tile.Dangerous != false {
		t.Errorf("NewFloorTile().Dangerous = %v, want false", tile.Dangerous)
	}
	
	if tile.DamageType != "" {
		t.Errorf("NewFloorTile().DamageType = %v, want empty string", tile.DamageType)
	}
	
	if tile.Damage != 0 {
		t.Errorf("NewFloorTile().Damage = %v, want 0", tile.Damage)
	}
}

// TestNewWallTile tests the NewWallTile constructor function
func TestNewWallTile(t *testing.T) {
	tile := NewWallTile()
	
	// Test that all properties are set correctly for a wall tile
	if tile.Type != TileWall {
		t.Errorf("NewWallTile().Type = %v, want %v", tile.Type, TileWall)
	}
	
	if tile.Walkable != false {
		t.Errorf("NewWallTile().Walkable = %v, want false", tile.Walkable)
	}
	
	if tile.Transparent != false {
		t.Errorf("NewWallTile().Transparent = %v, want false", tile.Transparent)
	}
	
	if tile.Properties == nil {
		t.Errorf("NewWallTile().Properties = nil, want initialized map")
	}
	
	if len(tile.Properties) != 0 {
		t.Errorf("NewWallTile().Properties length = %v, want 0", len(tile.Properties))
	}
	
	if tile.Sprite != "" {
		t.Errorf("NewWallTile().Sprite = %v, want empty string", tile.Sprite)
	}
	
	expectedColor := RGB{128, 128, 128}
	if tile.Color != expectedColor {
		t.Errorf("NewWallTile().Color = %v, want %v", tile.Color, expectedColor)
	}
	
	if tile.BlocksSight != true {
		t.Errorf("NewWallTile().BlocksSight = %v, want true", tile.BlocksSight)
	}
	
	if tile.Dangerous != false {
		t.Errorf("NewWallTile().Dangerous = %v, want false", tile.Dangerous)
	}
	
	if tile.DamageType != "" {
		t.Errorf("NewWallTile().DamageType = %v, want empty string", tile.DamageType)
	}
	
	if tile.Damage != 0 {
		t.Errorf("NewWallTile().Damage = %v, want 0", tile.Damage)
	}
}

// TestNewFloorTile_PropertiesModification tests that the Properties map returned by NewFloorTile can be modified
func TestNewFloorTile_PropertiesModification(t *testing.T) {
	tile := NewFloorTile()
	
	// Test that we can modify the properties map
	tile.Properties["custom"] = "value"
	tile.Properties["number"] = 123
	
	if len(tile.Properties) != 2 {
		t.Errorf("Modified floor tile properties length = %v, want 2", len(tile.Properties))
	}
	
	if tile.Properties["custom"] != "value" {
		t.Errorf("tile.Properties[\"custom\"] = %v, want \"value\"", tile.Properties["custom"])
	}
	
	if tile.Properties["number"] != 123 {
		t.Errorf("tile.Properties[\"number\"] = %v, want 123", tile.Properties["number"])
	}
}

// TestNewWallTile_PropertiesModification tests that the Properties map returned by NewWallTile can be modified
func TestNewWallTile_PropertiesModification(t *testing.T) {
	tile := NewWallTile()
	
	// Test that we can modify the properties map
	tile.Properties["material"] = "stone"
	tile.Properties["height"] = 10
	
	if len(tile.Properties) != 2 {
		t.Errorf("Modified wall tile properties length = %v, want 2", len(tile.Properties))
	}
	
	if tile.Properties["material"] != "stone" {
		t.Errorf("tile.Properties[\"material\"] = %v, want \"stone\"", tile.Properties["material"])
	}
	
	if tile.Properties["height"] != 10 {
		t.Errorf("tile.Properties[\"height\"] = %v, want 10", tile.Properties["height"])
	}
}

// TestTileConstructors_Independence tests that different constructor calls return independent instances
func TestTileConstructors_Independence(t *testing.T) {
	floor1 := NewFloorTile()
	floor2 := NewFloorTile()
	wall1 := NewWallTile()
	wall2 := NewWallTile()
	
	// Modify properties of one instance
	floor1.Properties["test"] = "floor1"
	wall1.Properties["test"] = "wall1"
	
	// Verify other instances are not affected
	if len(floor2.Properties) != 0 {
		t.Errorf("floor2.Properties affected by floor1 modification: length = %v, want 0", len(floor2.Properties))
	}
	
	if len(wall2.Properties) != 0 {
		t.Errorf("wall2.Properties affected by wall1 modification: length = %v, want 0", len(wall2.Properties))
	}
	
	// Verify the modified instances have the expected values
	if floor1.Properties["test"] != "floor1" {
		t.Errorf("floor1.Properties[\"test\"] = %v, want \"floor1\"", floor1.Properties["test"])
	}
	
	if wall1.Properties["test"] != "wall1" {
		t.Errorf("wall1.Properties[\"test\"] = %v, want \"wall1\"", wall1.Properties["test"])
	}
}

// TestTile_DangerousConfiguration tests tiles configured as dangerous
func TestTile_DangerousConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		tileType   TileType
		dangerous  bool
		damageType string
		damage     int
		expectValid bool
	}{
		{
			name:       "Lava tile with fire damage",
			tileType:   TileLava,
			dangerous:  true,
			damageType: "fire",
			damage:     5,
			expectValid: true,
		},
		{
			name:       "Safe floor tile",
			tileType:   TileFloor,
			dangerous:  false,
			damageType: "",
			damage:     0,
			expectValid: true,
		},
		{
			name:       "Dangerous tile with zero damage",
			tileType:   TilePit,
			dangerous:  true,
			damageType: "fall",
			damage:     0,
			expectValid: true,
		},
		{
			name:       "Non-dangerous tile with damage configured",
			tileType:   TileWater,
			dangerous:  false,
			damageType: "drown",
			damage:     3,
			expectValid: true, // This should be valid but might be illogical
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := Tile{
				Type:       tt.tileType,
				Dangerous:  tt.dangerous,
				DamageType: tt.damageType,
				Damage:     tt.damage,
				Properties: make(map[string]interface{}),
			}
			
			if tile.Type != tt.tileType {
				t.Errorf("Tile.Type = %v, want %v", tile.Type, tt.tileType)
			}
			
			if tile.Dangerous != tt.dangerous {
				t.Errorf("Tile.Dangerous = %v, want %v", tile.Dangerous, tt.dangerous)
			}
			
			if tile.DamageType != tt.damageType {
				t.Errorf("Tile.DamageType = %v, want %v", tile.DamageType, tt.damageType)
			}
			
			if tile.Damage != tt.damage {
				t.Errorf("Tile.Damage = %v, want %v", tile.Damage, tt.damage)
			}
		})
	}
}
