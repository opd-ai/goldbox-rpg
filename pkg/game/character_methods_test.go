package game

import (
	"encoding/json"
	"sync"
	"testing"
)

// TestCharacter_GetHealth tests the GetHealth method
func TestCharacter_GetHealth(t *testing.T) {
	tests := []struct {
		name           string
		characterHP    int
		expectedHealth int
	}{
		{
			name:           "PositiveHealth",
			characterHP:    100,
			expectedHealth: 100,
		},
		{
			name:           "ZeroHealth",
			characterHP:    0,
			expectedHealth: 0,
		},
		{
			name:           "NegativeHealth",
			characterHP:    -10,
			expectedHealth: -10,
		},
		{
			name:           "MaxHealth",
			characterHP:    999,
			expectedHealth: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				HP: tt.characterHP,
			}

			result := character.GetHealth()
			if result != tt.expectedHealth {
				t.Errorf("GetHealth() = %v, want %v", result, tt.expectedHealth)
			}
		})
	}
}

// TestCharacter_IsObstacle tests the IsObstacle method
func TestCharacter_IsObstacle(t *testing.T) {
	t.Run("CharacterIsAlwaysObstacle", func(t *testing.T) {
		character := &Character{
			ID:   "test-character",
			Name: "Test Character",
		}

		result := character.IsObstacle()
		if !result {
			t.Error("IsObstacle() should always return true for characters")
		}
	})

	t.Run("EmptyCharacterIsStillObstacle", func(t *testing.T) {
		character := &Character{}

		result := character.IsObstacle()
		if !result {
			t.Error("IsObstacle() should return true even for empty character")
		}
	})
}

// TestCharacter_SetHealth tests the SetHealth method
func TestCharacter_SetHealth(t *testing.T) {
	tests := []struct {
		name        string
		initialHP   int
		maxHP       int
		newHealth   int
		expectedHP  int
		description string
	}{
		{
			name:        "SetNormalHealth",
			initialHP:   50,
			maxHP:       100,
			newHealth:   75,
			expectedHP:  75,
			description: "Setting health within normal range",
		},
		{
			name:        "SetHealthToZero",
			initialHP:   50,
			maxHP:       100,
			newHealth:   0,
			expectedHP:  0,
			description: "Setting health to zero",
		},
		{
			name:        "SetNegativeHealthCapsAtZero",
			initialHP:   50,
			maxHP:       100,
			newHealth:   -20,
			expectedHP:  0,
			description: "Negative health should be capped at 0",
		},
		{
			name:        "SetHealthAboveMaxCapsAtMax",
			initialHP:   50,
			maxHP:       100,
			newHealth:   150,
			expectedHP:  100,
			description: "Health above max should be capped at max",
		},
		{
			name:        "SetHealthEqualToMax",
			initialHP:   50,
			maxHP:       100,
			newHealth:   100,
			expectedHP:  100,
			description: "Setting health equal to max",
		},
		{
			name:        "LargeNegativeValue",
			initialHP:   50,
			maxHP:       100,
			newHealth:   -1000,
			expectedHP:  0,
			description: "Very negative health should still cap at 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				HP:    tt.initialHP,
				MaxHP: tt.maxHP,
			}

			character.SetHealth(tt.newHealth)

			if character.HP != tt.expectedHP {
				t.Errorf("SetHealth(%v): HP = %v, want %v. %s",
					tt.newHealth, character.HP, tt.expectedHP, tt.description)
			}
		})
	}
}

// TestCharacter_SetHealth_Concurrent tests SetHealth method for thread safety
func TestCharacter_SetHealth_Concurrent(t *testing.T) {
	character := &Character{
		HP:    50,
		MaxHP: 100,
	}

	// Number of goroutines and iterations per goroutine
	numGoroutines := 10
	iterationsPerGoroutine := 100

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines that concurrently call SetHealth
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				// Alternate between different health values
				if j%2 == 0 {
					character.SetHealth(75)
				} else {
					character.SetHealth(25)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify character is in a valid state (HP should be either 25 or 75)
	finalHP := character.HP
	if finalHP != 25 && finalHP != 75 {
		t.Errorf("After concurrent SetHealth calls, HP = %v, expected 25 or 75", finalHP)
	}

	// Verify HP is within valid bounds
	if finalHP < 0 || finalHP > character.MaxHP {
		t.Errorf("After concurrent SetHealth calls, HP = %v is outside valid bounds [0, %v]", finalHP, character.MaxHP)
	}
}

// TestCharacter_GetName tests the GetName method
func TestCharacter_GetName(t *testing.T) {
	tests := []struct {
		name          string
		characterName string
	}{
		{
			name:          "RegularName",
			characterName: "Aragorn",
		},
		{
			name:          "EmptyName",
			characterName: "",
		},
		{
			name:          "NameWithSpaces",
			characterName: "Gandalf the Grey",
		},
		{
			name:          "NameWithSpecialCharacters",
			characterName: "Légolas-Elf",
		},
		{
			name:          "VeryLongName",
			characterName: "Thranduil-Oropherion-Elvenking-of-the-Woodland-Realm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				Name: tt.characterName,
			}

			result := character.GetName()
			if result != tt.characterName {
				t.Errorf("GetName() = %v, want %v", result, tt.characterName)
			}
		})
	}
}

// TestCharacter_GetDescription tests the GetDescription method
func TestCharacter_GetDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "RegularDescription",
			description: "A brave warrior from the North",
		},
		{
			name:        "EmptyDescription",
			description: "",
		},
		{
			name:        "LongDescription",
			description: "A tall, weathered ranger with dark hair and grey eyes. He carries himself with the bearing of nobility despite his rough appearance. Known to be the heir of Isildur.",
		},
		{
			name:        "DescriptionWithNewlines",
			description: "Line 1\nLine 2\nLine 3",
		},
		{
			name:        "DescriptionWithSpecialCharacters",
			description: "Description with éñ special çhàracters & symbols!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				Description: tt.description,
			}

			result := character.GetDescription()
			if result != tt.description {
				t.Errorf("GetDescription() = %v, want %v", result, tt.description)
			}
		})
	}
}

// TestCharacter_SetPosition tests the SetPosition method
func TestCharacter_SetPosition(t *testing.T) {
	tests := []struct {
		name        string
		position    Position
		expectError bool
		description string
	}{
		{
			name: "ValidPosition",
			position: Position{
				X:     10,
				Y:     20,
				Level: 1,
			},
			expectError: false,
			description: "Valid position should be set without error",
		},
		{
			name: "ZeroPosition",
			position: Position{
				X:     0,
				Y:     0,
				Level: 0,
			},
			expectError: false,
			description: "Zero coordinates should be valid",
		},
		{
			name: "PositivePosition",
			position: Position{
				X:     100,
				Y:     200,
				Level: 5,
			},
			expectError: false,
			description: "Large positive coordinates should be valid",
		},
		{
			name: "NegativeXPosition",
			position: Position{
				X:     -1,
				Y:     10,
				Level: 1,
			},
			expectError: true,
			description: "Negative X coordinate should be invalid",
		},
		{
			name: "NegativeYPosition",
			position: Position{
				X:     10,
				Y:     -1,
				Level: 1,
			},
			expectError: true,
			description: "Negative Y coordinate should be invalid",
		},
		{
			name: "NegativeLevelPosition",
			position: Position{
				X:     10,
				Y:     10,
				Level: -1,
			},
			expectError: true,
			description: "Negative level should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				Position: Position{X: 5, Y: 5, Level: 1},
			}

			err := character.SetPosition(tt.position)

			if tt.expectError {
				if err == nil {
					t.Errorf("SetPosition(%v) expected error, got nil. %s", tt.position, tt.description)
				}
				// Position should not change on error
				if character.Position == tt.position {
					t.Error("Position should not change when SetPosition returns error")
				}
			} else {
				if err != nil {
					t.Errorf("SetPosition(%v) unexpected error: %v. %s", tt.position, err, tt.description)
				}
				// Position should be updated on success
				if character.Position != tt.position {
					t.Errorf("Position = %v, want %v", character.Position, tt.position)
				}
			}
		})
	}
}

// TestCharacter_IsActive tests the IsActive method
func TestCharacter_IsActive(t *testing.T) {
	tests := []struct {
		name           string
		activeState    bool
		expectedResult bool
	}{
		{
			name:           "CharacterIsActive",
			activeState:    true,
			expectedResult: true,
		},
		{
			name:           "CharacterIsInactive",
			activeState:    false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				active: tt.activeState,
			}

			result := character.IsActive()
			if result != tt.expectedResult {
				t.Errorf("IsActive() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestCharacter_SetActive tests the SetActive method
func TestCharacter_SetActive(t *testing.T) {
	tests := []struct {
		name         string
		initialState bool
		newState     bool
	}{
		{
			name:         "SetActiveToTrue",
			initialState: false,
			newState:     true,
		},
		{
			name:         "SetActiveToFalse",
			initialState: true,
			newState:     false,
		},
		{
			name:         "SetActiveTrueWhenAlreadyTrue",
			initialState: true,
			newState:     true,
		},
		{
			name:         "SetActiveFalseWhenAlreadyFalse",
			initialState: false,
			newState:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				active: tt.initialState,
			}

			character.SetActive(tt.newState)

			if character.active != tt.newState {
				t.Errorf("SetActive(%v): active = %v, want %v", tt.newState, character.active, tt.newState)
			}

			// Verify IsActive() also returns the new state
			if character.IsActive() != tt.newState {
				t.Errorf("IsActive() = %v, want %v after SetActive(%v)", character.IsActive(), tt.newState, tt.newState)
			}
		})
	}
}

// TestCharacter_GetTags tests the GetTags method
func TestCharacter_GetTags(t *testing.T) {
	tests := []struct {
		name          string
		characterTags []string
		description   string
	}{
		{
			name:          "EmptyTags",
			characterTags: []string{},
			description:   "Character with no tags",
		},
		{
			name:          "SingleTag",
			characterTags: []string{"warrior"},
			description:   "Character with one tag",
		},
		{
			name:          "MultipleTags",
			characterTags: []string{"warrior", "noble", "human"},
			description:   "Character with multiple tags",
		},
		{
			name:          "TagsWithSpaces",
			characterTags: []string{"magic user", "spell caster"},
			description:   "Tags containing spaces",
		},
		{
			name:          "NilTags",
			characterTags: nil,
			description:   "Character with nil tags slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{
				tags: tt.characterTags,
			}

			result := character.GetTags()

			// Check that we get the correct values
			if len(result) != len(tt.characterTags) {
				t.Errorf("GetTags() length = %v, want %v. %s", len(result), len(tt.characterTags), tt.description)
			}

			for i, tag := range tt.characterTags {
				if i >= len(result) || result[i] != tag {
					t.Errorf("GetTags()[%d] = %v, want %v. %s", i, result[i], tag, tt.description)
				}
			}

			// Verify that modifying the returned slice doesn't affect the original
			if len(result) > 0 {
				result[0] = "modified"
				// Get tags again to ensure original wasn't modified
				result2 := character.GetTags()
				if len(result2) > 0 && result2[0] == "modified" {
					t.Error("GetTags() should return a copy, not a reference to internal slice")
				}
			}
		})
	}
}

// TestCharacter_ToJSON tests the ToJSON method
func TestCharacter_ToJSON(t *testing.T) {
	tests := []struct {
		name        string
		character   *Character
		expectError bool
		description string
	}{
		{
			name: "BasicCharacterSerialization",
			character: &Character{
				ID:          "char-001",
				Name:        "Test Hero",
				Description: "A brave adventurer",
				HP:          100,
				MaxHP:       100,
				Strength:    18,
				Dexterity:   15,
			},
			expectError: false,
			description: "Basic character should serialize successfully",
		},
		{
			name:        "EmptyCharacterSerialization",
			character:   &Character{},
			expectError: false,
			description: "Empty character should serialize successfully",
		},
		{
			name: "CharacterWithComplexData",
			character: &Character{
				ID:          "char-002",
				Name:        "Complex Hero",
				Description: "A character with special chars: éñüñ",
				Position:    Position{X: 10, Y: 20, Level: 2},
				tags:        []string{"warrior", "magic-user"},
				Equipment:   make(map[EquipmentSlot]Item),
				Inventory:   []Item{},
			},
			expectError: false,
			description: "Character with complex data should serialize successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := tt.character.ToJSON()

			if tt.expectError {
				if err == nil {
					t.Errorf("ToJSON() expected error, got nil. %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("ToJSON() unexpected error: %v. %s", err, tt.description)
				return
			}

			if len(jsonData) == 0 {
				t.Error("ToJSON() returned empty data")
				return
			}

			// Verify the JSON is valid by unmarshaling it
			var testChar Character
			if err := json.Unmarshal(jsonData, &testChar); err != nil {
				t.Errorf("ToJSON() produced invalid JSON: %v", err)
			}

			// Verify key fields are preserved
			if testChar.ID != tt.character.ID {
				t.Errorf("ToJSON() ID not preserved: got %v, want %v", testChar.ID, tt.character.ID)
			}
			if testChar.Name != tt.character.Name {
				t.Errorf("ToJSON() Name not preserved: got %v, want %v", testChar.Name, tt.character.Name)
			}
		})
	}
}

// TestCharacter_FromJSON tests the FromJSON method
func TestCharacter_FromJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		validator   func(*testing.T, *Character)
		description string
	}{
		{
			name:        "ValidJSONDeserialization",
			jsonData:    `{"ID":"test-001","Name":"Test Character","HP":75,"MaxHP":100}`,
			expectError: false,
			validator: func(t *testing.T, c *Character) {
				if c.ID != "test-001" {
					t.Errorf("ID = %v, want test-001", c.ID)
				}
				if c.Name != "Test Character" {
					t.Errorf("Name = %v, want Test Character", c.Name)
				}
				if c.HP != 75 {
					t.Errorf("HP = %v, want 75", c.HP)
				}
				if c.MaxHP != 100 {
					t.Errorf("MaxHP = %v, want 100", c.MaxHP)
				}
			},
			description: "Valid JSON should deserialize correctly",
		},
		{
			name:        "EmptyJSONObject",
			jsonData:    `{}`,
			expectError: false,
			validator: func(t *testing.T, c *Character) {
				if c.ID != "" {
					t.Errorf("ID = %v, want empty string", c.ID)
				}
				if c.Name != "" {
					t.Errorf("Name = %v, want empty string", c.Name)
				}
			},
			description: "Empty JSON object should create character with default values",
		},
		{
			name:        "InvalidJSON",
			jsonData:    `{"invalid": json}`,
			expectError: true,
			validator:   nil,
			description: "Invalid JSON should return error",
		},
		{
			name:        "PartialJSON",
			jsonData:    `{"Name":"Partial Character","Strength":18}`,
			expectError: false,
			validator: func(t *testing.T, c *Character) {
				if c.Name != "Partial Character" {
					t.Errorf("Name = %v, want Partial Character", c.Name)
				}
				if c.Strength != 18 {
					t.Errorf("Strength = %v, want 18", c.Strength)
				}
				if c.ID != "" {
					t.Errorf("ID = %v, want empty string (should remain default)", c.ID)
				}
			},
			description: "Partial JSON should set only provided fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			character := &Character{}
			err := character.FromJSON([]byte(tt.jsonData))

			if tt.expectError {
				if err == nil {
					t.Errorf("FromJSON() expected error, got nil. %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("FromJSON() unexpected error: %v. %s", err, tt.description)
				return
			}

			if tt.validator != nil {
				tt.validator(t, character)
			}
		})
	}
}

// TestCharacter_JSONRoundTrip tests serialization and deserialization together
func TestCharacter_JSONRoundTrip(t *testing.T) {
	originalCharacter := &Character{
		ID:           "test-round-trip",
		Name:         "Round Trip Hero",
		Description:  "Testing JSON round trip",
		HP:           85,
		MaxHP:        100,
		Strength:     16,
		Dexterity:    14,
		Constitution: 15,
		Intelligence: 12,
		Wisdom:       13,
		Charisma:     10,
		ArmorClass:   18,
		THAC0:        10,
		Gold:         500,
		Position:     Position{X: 15, Y: 25, Level: 3},
		active:       true,
		tags:         []string{"hero", "adventurer"},
	}

	// Serialize to JSON
	jsonData, err := originalCharacter.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	// Deserialize from JSON
	newCharacter := &Character{}
	err = newCharacter.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() failed: %v", err)
	}

	// Compare key fields
	if newCharacter.ID != originalCharacter.ID {
		t.Errorf("ID mismatch: got %v, want %v", newCharacter.ID, originalCharacter.ID)
	}
	if newCharacter.Name != originalCharacter.Name {
		t.Errorf("Name mismatch: got %v, want %v", newCharacter.Name, originalCharacter.Name)
	}
	if newCharacter.HP != originalCharacter.HP {
		t.Errorf("HP mismatch: got %v, want %v", newCharacter.HP, originalCharacter.HP)
	}
	if newCharacter.Strength != originalCharacter.Strength {
		t.Errorf("Strength mismatch: got %v, want %v", newCharacter.Strength, originalCharacter.Strength)
	}
	if newCharacter.Position.X != originalCharacter.Position.X {
		t.Errorf("Position.X mismatch: got %v, want %v", newCharacter.Position.X, originalCharacter.Position.X)
	}
}
