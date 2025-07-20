package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInputValidator(t *testing.T) {
	validator := NewInputValidator(1024)

	assert.NotNil(t, validator)
	assert.Equal(t, int64(1024), validator.maxRequestSize)
	assert.NotEmpty(t, validator.validators)

	// Check that all expected methods are registered
	expectedMethods := []string{
		"ping", "createPlayer", "getPlayer", "listPlayers",
		"createCharacter", "getCharacter", "updateCharacter", "listCharacters",
		"move", "getPosition", "attack", "castSpell", "getSpells",
		"getWorld", "getWorldState", "equipItem", "unequipItem", "getInventory",
	}

	for _, method := range expectedMethods {
		_, exists := validator.validators[method]
		assert.True(t, exists, "method %s should be registered", method)
	}
}

func TestValidateRPCRequest(t *testing.T) {
	validator := NewInputValidator(100)

	tests := []struct {
		name          string
		method        string
		params        interface{}
		requestSize   int64
		expectError   bool
		errorContains string
	}{
		{
			name:          "request too large",
			method:        "ping",
			params:        nil,
			requestSize:   200,
			expectError:   true,
			errorContains: "exceeds maximum",
		},
		{
			name:          "unknown method",
			method:        "unknownMethod",
			params:        nil,
			requestSize:   50,
			expectError:   true,
			errorContains: "unknown method",
		},
		{
			name:        "valid ping request",
			method:      "ping",
			params:      nil,
			requestSize: 50,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRPCRequest(tt.method, tt.params, tt.requestSize)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCreatePlayer(t *testing.T) {
	validator := NewInputValidator(1024)

	tests := []struct {
		name          string
		params        interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid player creation",
			params:      map[string]interface{}{"name": "TestPlayer"},
			expectError: false,
		},
		{
			name:          "missing name parameter",
			params:        map[string]interface{}{},
			expectError:   true,
			errorContains: "requires 'name' parameter",
		},
		{
			name:          "non-string name",
			params:        map[string]interface{}{"name": 123},
			expectError:   true,
			errorContains: "must be a string",
		},
		{
			name:          "empty name",
			params:        map[string]interface{}{"name": ""},
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "name too long",
			params:        map[string]interface{}{"name": strings.Repeat("a", 51)},
			expectError:   true,
			errorContains: "cannot exceed 50 characters",
		},
		{
			name:          "invalid characters in name",
			params:        map[string]interface{}{"name": "Player<script>"},
			expectError:   true,
			errorContains: "invalid characters",
		},
		{
			name:          "non-object parameters",
			params:        "not an object",
			expectError:   true,
			errorContains: "expects object parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateCreatePlayer(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCreateCharacter(t *testing.T) {
	validator := NewInputValidator(1024)
	validSessionID := "12345678-1234-1234-1234-123456789abc"

	tests := []struct {
		name          string
		params        interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "valid character creation",
			params: map[string]interface{}{
				"sessionId": validSessionID,
				"name":      "TestCharacter",
				"class":     "fighter",
			},
			expectError: false,
		},
		{
			name: "missing session ID",
			params: map[string]interface{}{
				"name":  "TestCharacter",
				"class": "fighter",
			},
			expectError:   true,
			errorContains: "sessionId",
		},
		{
			name: "missing name",
			params: map[string]interface{}{
				"sessionId": validSessionID,
				"class":     "fighter",
			},
			expectError:   true,
			errorContains: "requires 'name' parameter",
		},
		{
			name: "missing class",
			params: map[string]interface{}{
				"sessionId": validSessionID,
				"name":      "TestCharacter",
			},
			expectError:   true,
			errorContains: "requires 'class' parameter",
		},
		{
			name: "invalid character class",
			params: map[string]interface{}{
				"sessionId": validSessionID,
				"name":      "TestCharacter",
				"class":     "invalidclass",
			},
			expectError:   true,
			errorContains: "invalid character class",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateCreateCharacter(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMove(t *testing.T) {
	validator := NewInputValidator(1024)
	validSessionID := "12345678-1234-1234-1234-123456789abc"

	tests := []struct {
		name          string
		params        interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "valid move",
			params: map[string]interface{}{
				"sessionId": validSessionID,
				"x":         100.0,
				"y":         200.0,
			},
			expectError: false,
		},
		{
			name: "missing coordinates",
			params: map[string]interface{}{
				"sessionId": validSessionID,
			},
			expectError:   true,
			errorContains: "requires 'x' and 'y' coordinates",
		},
		{
			name: "coordinates out of range",
			params: map[string]interface{}{
				"sessionId": validSessionID,
				"x":         15000.0,
				"y":         200.0,
			},
			expectError:   true,
			errorContains: "out of valid range",
		},
		{
			name: "non-numeric coordinates",
			params: map[string]interface{}{
				"sessionId": validSessionID,
				"x":         "invalid",
				"y":         200.0,
			},
			expectError:   true,
			errorContains: "must be a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateMove(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		expectError bool
	}{
		{
			name:        "valid UUID",
			uuid:        "12345678-1234-1234-1234-123456789abc",
			expectError: false,
		},
		{
			name:        "valid UUID with uppercase",
			uuid:        "12345678-1234-1234-1234-123456789ABC",
			expectError: false,
		},
		{
			name:        "invalid UUID format - too short",
			uuid:        "12345678-1234-1234-1234-123456789ab",
			expectError: true,
		},
		{
			name:        "invalid UUID format - missing dashes",
			uuid:        "123456781234123412341234123456789abc",
			expectError: true,
		},
		{
			name:        "invalid UUID format - invalid characters",
			uuid:        "12345678-1234-1234-1234-123456789abg",
			expectError: true,
		},
		{
			name:        "empty UUID",
			uuid:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUUID(tt.uuid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePlayerName(t *testing.T) {
	tests := []struct {
		name        string
		playerName  string
		expectError bool
	}{
		{
			name:        "valid name",
			playerName:  "TestPlayer",
			expectError: false,
		},
		{
			name:        "valid name with spaces",
			playerName:  "Test Player",
			expectError: false,
		},
		{
			name:        "valid name with numbers",
			playerName:  "TestPlayer123",
			expectError: false,
		},
		{
			name:        "valid name with allowed punctuation",
			playerName:  "Test-Player_42.0",
			expectError: false,
		},
		{
			name:        "empty name",
			playerName:  "",
			expectError: true,
		},
		{
			name:        "name too long",
			playerName:  strings.Repeat("a", 51),
			expectError: true,
		},
		{
			name:        "name with invalid characters",
			playerName:  "Test<Player>",
			expectError: true,
		},
		{
			name:        "name with only whitespace",
			playerName:  "   ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePlayerName(tt.playerName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCharacterClass(t *testing.T) {
	tests := []struct {
		name        string
		class       string
		expectError bool
	}{
		{
			name:        "valid class - fighter",
			class:       "fighter",
			expectError: false,
		},
		{
			name:        "valid class - wizard",
			class:       "wizard",
			expectError: false,
		},
		{
			name:        "valid class with uppercase",
			class:       "FIGHTER",
			expectError: false,
		},
		{
			name:        "valid class with whitespace",
			class:       " fighter ",
			expectError: false,
		},
		{
			name:        "invalid class",
			class:       "invalidclass",
			expectError: true,
		},
		{
			name:        "empty class",
			class:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCharacterClass(tt.class)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSpellID(t *testing.T) {
	tests := []struct {
		name        string
		spellID     string
		expectError bool
	}{
		{
			name:        "valid spell ID",
			spellID:     "magic-missile",
			expectError: false,
		},
		{
			name:        "valid spell ID with numbers",
			spellID:     "fireball-lvl3",
			expectError: false,
		},
		{
			name:        "valid spell ID with underscores",
			spellID:     "healing_light",
			expectError: false,
		},
		{
			name:        "invalid spell ID with uppercase",
			spellID:     "Magic-Missile",
			expectError: true,
		},
		{
			name:        "invalid spell ID with spaces",
			spellID:     "magic missile",
			expectError: true,
		},
		{
			name:        "empty spell ID",
			spellID:     "",
			expectError: true,
		},
		{
			name:        "spell ID too long",
			spellID:     strings.Repeat("a", 101),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSpellID(tt.spellID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEquipmentSlot(t *testing.T) {
	tests := []struct {
		name        string
		slot        string
		expectError bool
	}{
		{
			name:        "valid slot - head",
			slot:        "head",
			expectError: false,
		},
		{
			name:        "valid slot - main-hand",
			slot:        "main-hand",
			expectError: false,
		},
		{
			name:        "valid slot with uppercase",
			slot:        "HEAD",
			expectError: false,
		},
		{
			name:        "valid slot with whitespace",
			slot:        " chest ",
			expectError: false,
		},
		{
			name:        "invalid slot",
			slot:        "invalid-slot",
			expectError: true,
		},
		{
			name:        "empty slot",
			slot:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEquipmentSlot(tt.slot)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
