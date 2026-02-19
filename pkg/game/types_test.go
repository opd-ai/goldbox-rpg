package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestDirection_Constants tests that all direction constants have the expected values
func TestDirection_Constants(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		expected  int
	}{
		{"North direction", DirectionNorth, 0},
		{"East direction", DirectionEast, 1},
		{"South direction", DirectionSouth, 2},
		{"West direction", DirectionWest, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.direction) != tt.expected {
				t.Errorf("Expected %s to have value %d, got %d", tt.name, tt.expected, int(tt.direction))
			}
		})
	}
}

// TestDirection_LegacyConstants tests backward compatibility constants
func TestDirection_LegacyConstants(t *testing.T) {
	tests := []struct {
		name   string
		legacy Direction
		modern Direction
	}{
		{"North legacy", North, DirectionNorth},
		{"East legacy", East, DirectionEast},
		{"South legacy", South, DirectionSouth},
		{"West legacy", West, DirectionWest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.legacy != tt.modern {
				t.Errorf("Legacy constant %s should equal modern constant, got %d vs %d", tt.name, tt.legacy, tt.modern)
			}
		})
	}
}

// TestPosition_Creation tests Position struct creation and field assignment
func TestPosition_Creation(t *testing.T) {
	tests := []struct {
		name   string
		x      int
		y      int
		level  int
		facing Direction
	}{
		{"Origin position", 0, 0, 0, DirectionNorth},
		{"Positive coordinates", 10, 15, 1, DirectionEast},
		{"Negative coordinates", -5, -3, -1, DirectionSouth},
		{"Mixed coordinates", -2, 8, 2, DirectionWest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := Position{
				X:      tt.x,
				Y:      tt.y,
				Level:  tt.level,
				Facing: tt.facing,
			}

			if pos.X != tt.x {
				t.Errorf("Expected X to be %d, got %d", tt.x, pos.X)
			}
			if pos.Y != tt.y {
				t.Errorf("Expected Y to be %d, got %d", tt.y, pos.Y)
			}
			if pos.Level != tt.level {
				t.Errorf("Expected Level to be %d, got %d", tt.level, pos.Level)
			}
			if pos.Facing != tt.facing {
				t.Errorf("Expected Facing to be %d, got %d", tt.facing, pos.Facing)
			}
		})
	}
}

// TestPosition_YAMLTags tests that Position struct has correct YAML tags
func TestPosition_YAMLTags(t *testing.T) {
	posType := reflect.TypeOf(Position{})

	tests := []struct {
		fieldName string
		yamlTag   string
	}{
		{"X", "position_x"},
		{"Y", "position_y"},
		{"Level", "position_level"},
		{"Facing", "position_facing"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			field, ok := posType.FieldByName(tt.fieldName)
			if !ok {
				t.Errorf("Field %s not found in Position struct", tt.fieldName)
				return
			}

			yamlTag := field.Tag.Get("yaml")
			if yamlTag != tt.yamlTag {
				t.Errorf("Expected YAML tag for %s to be %q, got %q", tt.fieldName, tt.yamlTag, yamlTag)
			}
		})
	}
}

// TestDirectionConfig_Creation tests DirectionConfig struct creation
func TestDirectionConfig_Creation(t *testing.T) {
	tests := []struct {
		name        string
		value       Direction
		nameStr     string
		degreeAngle int
	}{
		{"North config", DirectionNorth, "North", 0},
		{"East config", DirectionEast, "East", 90},
		{"South config", DirectionSouth, "South", 180},
		{"West config", DirectionWest, "West", 270},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DirectionConfig{
				Value:       tt.value,
				Name:        tt.nameStr,
				DegreeAngle: tt.degreeAngle,
			}

			if config.Value != tt.value {
				t.Errorf("Expected Value to be %d, got %d", tt.value, config.Value)
			}
			if config.Name != tt.nameStr {
				t.Errorf("Expected Name to be %q, got %q", tt.nameStr, config.Name)
			}
			if config.DegreeAngle != tt.degreeAngle {
				t.Errorf("Expected DegreeAngle to be %d, got %d", tt.degreeAngle, config.DegreeAngle)
			}
		})
	}
}

// TestDirectionConfig_YAMLTags tests that DirectionConfig struct has correct YAML tags
func TestDirectionConfig_YAMLTags(t *testing.T) {
	configType := reflect.TypeOf(DirectionConfig{})

	tests := []struct {
		fieldName string
		yamlTag   string
	}{
		{"Value", "direction_value"},
		{"Name", "direction_name"},
		{"DegreeAngle", "direction_angle"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			field, ok := configType.FieldByName(tt.fieldName)
			if !ok {
				t.Errorf("Field %s not found in DirectionConfig struct", tt.fieldName)
				return
			}

			yamlTag := field.Tag.Get("yaml")
			if yamlTag != tt.yamlTag {
				t.Errorf("Expected YAML tag for %s to be %q, got %q", tt.fieldName, tt.yamlTag, yamlTag)
			}
		})
	}
}

// TestDirection_TypeProperties tests properties of the Direction type
func TestDirection_TypeProperties(t *testing.T) {
	t.Run("Direction type is integer", func(t *testing.T) {
		var d Direction
		dirType := reflect.TypeOf(d)
		if dirType.Kind() != reflect.Int {
			t.Errorf("Expected Direction to be int type, got %v", dirType.Kind())
		}
	})

	t.Run("Direction values are sequential", func(t *testing.T) {
		directions := []Direction{DirectionNorth, DirectionEast, DirectionSouth, DirectionWest}
		for i, direction := range directions {
			if int(direction) != i {
				t.Errorf("Expected direction at index %d to have value %d, got %d", i, i, int(direction))
			}
		}
	})
}

// TestGameObject_Interface tests that GameObject interface is properly defined
func TestGameObject_Interface(t *testing.T) {
	gameObjectType := reflect.TypeOf((*GameObject)(nil)).Elem()

	expectedMethods := []string{
		"GetID",
		"GetName",
		"GetDescription",
		"GetPosition",
		"SetPosition",
		"IsActive",
		"GetTags",
		"ToJSON",
		"FromJSON",
		"GetHealth",
		"SetHealth",
		"IsObstacle",
	}

	t.Run("GameObject has all required methods", func(t *testing.T) {
		for _, methodName := range expectedMethods {
			_, ok := gameObjectType.MethodByName(methodName)
			if !ok {
				t.Errorf("GameObject interface missing method: %s", methodName)
			}
		}
	})

	t.Run("GameObject method count", func(t *testing.T) {
		actualMethodCount := gameObjectType.NumMethod()
		expectedMethodCount := len(expectedMethods)
		if actualMethodCount != expectedMethodCount {
			t.Errorf("Expected GameObject to have %d methods, got %d", expectedMethodCount, actualMethodCount)
		}
	})
}

// TestGameObject_MethodSignatures tests GameObject interface method signatures
func TestGameObject_MethodSignatures(t *testing.T) {
	gameObjectType := reflect.TypeOf((*GameObject)(nil)).Elem()

	tests := []struct {
		methodName string
		numIn      int
		numOut     int
		outTypes   []reflect.Type
	}{
		{"GetID", 0, 1, []reflect.Type{reflect.TypeOf("")}},
		{"GetName", 0, 1, []reflect.Type{reflect.TypeOf("")}},
		{"GetDescription", 0, 1, []reflect.Type{reflect.TypeOf("")}},
		{"GetPosition", 0, 1, []reflect.Type{reflect.TypeOf(Position{})}},
		{"SetPosition", 1, 1, []reflect.Type{reflect.TypeOf((*error)(nil)).Elem()}},
		{"IsActive", 0, 1, []reflect.Type{reflect.TypeOf(true)}},
		{"GetTags", 0, 1, []reflect.Type{reflect.TypeOf([]string{})}},
		{"ToJSON", 0, 2, []reflect.Type{reflect.TypeOf([]byte{}), reflect.TypeOf((*error)(nil)).Elem()}},
		{"FromJSON", 1, 1, []reflect.Type{reflect.TypeOf((*error)(nil)).Elem()}},
		{"GetHealth", 0, 1, []reflect.Type{reflect.TypeOf(0)}},
		{"SetHealth", 1, 0, []reflect.Type{}},
		{"IsObstacle", 0, 1, []reflect.Type{reflect.TypeOf(true)}},
	}

	for _, tt := range tests {
		t.Run(tt.methodName, func(t *testing.T) {
			method, ok := gameObjectType.MethodByName(tt.methodName)
			if !ok {
				t.Errorf("Method %s not found", tt.methodName)
				return
			}

			methodType := method.Type

			// Check input parameter count (excluding receiver)
			if methodType.NumIn() != tt.numIn {
				t.Errorf("Method %s expected %d input parameters, got %d", tt.methodName, tt.numIn, methodType.NumIn())
			}

			// Check output parameter count
			if methodType.NumOut() != tt.numOut {
				t.Errorf("Method %s expected %d output parameters, got %d", tt.methodName, tt.numOut, methodType.NumOut())
			}

			// Check output types
			for i, expectedType := range tt.outTypes {
				if i < methodType.NumOut() {
					actualType := methodType.Out(i)
					if actualType != expectedType {
						t.Errorf("Method %s output %d expected type %v, got %v", tt.methodName, i, expectedType, actualType)
					}
				}
			}
		})
	}
}

// TestPosition_EdgeCases tests Position struct with edge case values
func TestPosition_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		position Position
	}{
		{
			"Maximum int values",
			Position{X: 2147483647, Y: 2147483647, Level: 2147483647, Facing: DirectionWest},
		},
		{
			"Minimum int values",
			Position{X: -2147483648, Y: -2147483648, Level: -2147483648, Facing: DirectionNorth},
		},
		{
			"Zero values",
			Position{X: 0, Y: 0, Level: 0, Facing: DirectionNorth},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the position can be created and accessed without issues
			pos := tt.position

			if pos.X != tt.position.X {
				t.Errorf("X value not preserved: expected %d, got %d", tt.position.X, pos.X)
			}
			if pos.Y != tt.position.Y {
				t.Errorf("Y value not preserved: expected %d, got %d", tt.position.Y, pos.Y)
			}
			if pos.Level != tt.position.Level {
				t.Errorf("Level value not preserved: expected %d, got %d", tt.position.Level, pos.Level)
			}
			if pos.Facing != tt.position.Facing {
				t.Errorf("Facing value not preserved: expected %d, got %d", tt.position.Facing, pos.Facing)
			}
		})
	}
}

// TestDirectionConfig_EdgeCases tests DirectionConfig with edge case values
func TestDirectionConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config DirectionConfig
	}{
		{
			"Empty name",
			DirectionConfig{Value: DirectionNorth, Name: "", DegreeAngle: 0},
		},
		{
			"Large angle",
			DirectionConfig{Value: DirectionEast, Name: "East", DegreeAngle: 450},
		},
		{
			"Negative angle",
			DirectionConfig{Value: DirectionWest, Name: "West", DegreeAngle: -90},
		},
		{
			"Long name",
			DirectionConfig{Value: DirectionSouth, Name: "VeryLongDirectionNameThatExceedsNormalLimits", DegreeAngle: 180},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config

			if config.Value != tt.config.Value {
				t.Errorf("Value not preserved: expected %d, got %d", tt.config.Value, config.Value)
			}
			if config.Name != tt.config.Name {
				t.Errorf("Name not preserved: expected %q, got %q", tt.config.Name, config.Name)
			}
			if config.DegreeAngle != tt.config.DegreeAngle {
				t.Errorf("DegreeAngle not preserved: expected %d, got %d", tt.config.DegreeAngle, config.DegreeAngle)
			}
		})
	}
}

// TestDirection_JSONSerialization tests JSON serialization/deserialization of Direction
func TestDirection_JSONSerialization(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
	}{
		{"North direction", DirectionNorth},
		{"East direction", DirectionEast},
		{"South direction", DirectionSouth},
		{"West direction", DirectionWest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.direction)
			if err != nil {
				t.Errorf("Failed to marshal Direction: %v", err)
				return
			}

			// Test unmarshaling
			var direction Direction
			err = json.Unmarshal(data, &direction)
			if err != nil {
				t.Errorf("Failed to unmarshal Direction: %v", err)
				return
			}

			if direction != tt.direction {
				t.Errorf("Direction not preserved through JSON: expected %d, got %d", tt.direction, direction)
			}
		})
	}
}

// TestPosition_JSONSerialization tests JSON serialization/deserialization of Position
func TestPosition_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		position Position
	}{
		{
			"Basic position",
			Position{X: 10, Y: 20, Level: 1, Facing: DirectionEast},
		},
		{
			"Zero position",
			Position{X: 0, Y: 0, Level: 0, Facing: DirectionNorth},
		},
		{
			"Negative position",
			Position{X: -5, Y: -10, Level: -2, Facing: DirectionWest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.position)
			if err != nil {
				t.Errorf("Failed to marshal Position: %v", err)
				return
			}

			// Test unmarshaling
			var position Position
			err = json.Unmarshal(data, &position)
			if err != nil {
				t.Errorf("Failed to unmarshal Position: %v", err)
				return
			}

			if !reflect.DeepEqual(position, tt.position) {
				t.Errorf("Position not preserved through JSON: expected %+v, got %+v", tt.position, position)
			}
		})
	}
}

// TestIdentifiable_Interface tests that Identifiable interface is properly defined
func TestIdentifiable_Interface(t *testing.T) {
	identifiableType := reflect.TypeOf((*Identifiable)(nil)).Elem()

	expectedMethods := []string{
		"GetID",
		"GetName",
		"GetDescription",
	}

	t.Run("Identifiable has all required methods", func(t *testing.T) {
		for _, methodName := range expectedMethods {
			_, ok := identifiableType.MethodByName(methodName)
			if !ok {
				t.Errorf("Identifiable interface missing method: %s", methodName)
			}
		}
	})

	t.Run("Identifiable method count", func(t *testing.T) {
		actualMethodCount := identifiableType.NumMethod()
		expectedMethodCount := len(expectedMethods)
		if actualMethodCount != expectedMethodCount {
			t.Errorf("Expected Identifiable to have %d methods, got %d", expectedMethodCount, actualMethodCount)
		}
	})
}

// TestPositionable_Interface tests that Positionable interface is properly defined
func TestPositionable_Interface(t *testing.T) {
	positionableType := reflect.TypeOf((*Positionable)(nil)).Elem()

	expectedMethods := []string{
		"GetPosition",
		"SetPosition",
	}

	t.Run("Positionable has all required methods", func(t *testing.T) {
		for _, methodName := range expectedMethods {
			_, ok := positionableType.MethodByName(methodName)
			if !ok {
				t.Errorf("Positionable interface missing method: %s", methodName)
			}
		}
	})

	t.Run("Positionable method count", func(t *testing.T) {
		actualMethodCount := positionableType.NumMethod()
		expectedMethodCount := len(expectedMethods)
		if actualMethodCount != expectedMethodCount {
			t.Errorf("Expected Positionable to have %d methods, got %d", expectedMethodCount, actualMethodCount)
		}
	})
}

// TestDamageable_Interface tests that Damageable interface is properly defined
func TestDamageable_Interface(t *testing.T) {
	damageableType := reflect.TypeOf((*Damageable)(nil)).Elem()

	expectedMethods := []string{
		"GetHealth",
		"SetHealth",
	}

	t.Run("Damageable has all required methods", func(t *testing.T) {
		for _, methodName := range expectedMethods {
			_, ok := damageableType.MethodByName(methodName)
			if !ok {
				t.Errorf("Damageable interface missing method: %s", methodName)
			}
		}
	})

	t.Run("Damageable method count", func(t *testing.T) {
		actualMethodCount := damageableType.NumMethod()
		expectedMethodCount := len(expectedMethods)
		if actualMethodCount != expectedMethodCount {
			t.Errorf("Expected Damageable to have %d methods, got %d", expectedMethodCount, actualMethodCount)
		}
	})
}

// TestSerializable_Interface tests that Serializable interface is properly defined
func TestSerializable_Interface(t *testing.T) {
	serializableType := reflect.TypeOf((*Serializable)(nil)).Elem()

	expectedMethods := []string{
		"ToJSON",
		"FromJSON",
	}

	t.Run("Serializable has all required methods", func(t *testing.T) {
		for _, methodName := range expectedMethods {
			_, ok := serializableType.MethodByName(methodName)
			if !ok {
				t.Errorf("Serializable interface missing method: %s", methodName)
			}
		}
	})

	t.Run("Serializable method count", func(t *testing.T) {
		actualMethodCount := serializableType.NumMethod()
		expectedMethodCount := len(expectedMethods)
		if actualMethodCount != expectedMethodCount {
			t.Errorf("Expected Serializable to have %d methods, got %d", expectedMethodCount, actualMethodCount)
		}
	})
}

// TestGameObject_ComposesSmallInterfaces tests that GameObject properly embeds smaller interfaces
func TestGameObject_ComposesSmallInterfaces(t *testing.T) {
	// Create a test type that implements all smaller interfaces
	// and verify it satisfies GameObject
	gameObjectType := reflect.TypeOf((*GameObject)(nil)).Elem()
	identifiableType := reflect.TypeOf((*Identifiable)(nil)).Elem()
	positionableType := reflect.TypeOf((*Positionable)(nil)).Elem()
	damageableType := reflect.TypeOf((*Damageable)(nil)).Elem()
	serializableType := reflect.TypeOf((*Serializable)(nil)).Elem()

	// Verify that all methods from smaller interfaces are in GameObject
	smallerInterfaces := []struct {
		name       string
		interface_ reflect.Type
	}{
		{"Identifiable", identifiableType},
		{"Positionable", positionableType},
		{"Damageable", damageableType},
		{"Serializable", serializableType},
	}

	for _, smaller := range smallerInterfaces {
		t.Run(smaller.name+" methods in GameObject", func(t *testing.T) {
			for i := 0; i < smaller.interface_.NumMethod(); i++ {
				method := smaller.interface_.Method(i)
				_, ok := gameObjectType.MethodByName(method.Name)
				if !ok {
					t.Errorf("GameObject missing method from %s: %s", smaller.name, method.Name)
				}
			}
		})
	}
}

// TestItem_ImplementsInterfaces tests that Item implements the required interfaces
func TestItem_ImplementsInterfaces(t *testing.T) {
	item := &Item{
		ID:   "test-item",
		Name: "Test Sword",
	}

	t.Run("Item implements Identifiable", func(t *testing.T) {
		var _ Identifiable = item
	})

	t.Run("Item implements Positionable", func(t *testing.T) {
		var _ Positionable = item
	})

	t.Run("Item implements Damageable", func(t *testing.T) {
		var _ Damageable = item
	})

	t.Run("Item implements Serializable", func(t *testing.T) {
		var _ Serializable = item
	})

	t.Run("Item implements GameObject", func(t *testing.T) {
		var _ GameObject = item
	})
}

// TestCharacter_ImplementsInterfaces tests that Character implements the required interfaces
func TestCharacter_ImplementsInterfaces(t *testing.T) {
	char := &Character{
		ID:   "test-char",
		Name: "Test Hero",
	}

	t.Run("Character implements Identifiable", func(t *testing.T) {
		var _ Identifiable = char
	})

	t.Run("Character implements Positionable", func(t *testing.T) {
		var _ Positionable = char
	})

	t.Run("Character implements Damageable", func(t *testing.T) {
		var _ Damageable = char
	})

	t.Run("Character implements Serializable", func(t *testing.T) {
		var _ Serializable = char
	})

	t.Run("Character implements GameObject", func(t *testing.T) {
		var _ GameObject = char
	})
}
