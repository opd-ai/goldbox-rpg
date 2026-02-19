# Validation Package

This package provides comprehensive input validation for JSON-RPC requests in the GoldBox RPG Engine, ensuring security and data integrity.

## Overview

The validation package implements a robust framework for validating all user inputs before processing. It prevents security vulnerabilities like injection attacks and maintains data integrity throughout the system.

## Features

- **Comprehensive Input Validation**: Validates all JSON-RPC parameters
- **Security Focus**: Prevents injection attacks and DoS conditions
- **Method-Specific Validators**: Custom validation rules per API method
- **Size Limiting**: Request size limits to prevent memory exhaustion
- **Thread-Safe Operations**: Safe for concurrent use
- **Pre-registered Validators**: All JSON-RPC methods have built-in validators

## Components

### InputValidator

The main validation engine that:
- Maintains a registry of validation functions per method
- Enforces request size limits
- Provides method-specific validation
- Handles validation errors consistently

## Usage

### Basic Setup

```go
import "goldbox-rpg/pkg/validation"

// Create validator with size limit (validators are auto-registered)
validator := validation.NewInputValidator(1024 * 1024) // 1MB limit
```

### Validating Requests

```go
// In JSON-RPC handler
func (s *Server) handleRequest(method string, params interface{}, requestSize int64) (interface{}, error) {
    // Validate input first
    if err := s.validator.ValidateRPCRequest(method, params, requestSize); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Process validated request
    return s.processRequest(method, params)
}
```

## Built-in Validators

The following JSON-RPC methods have pre-registered validators:

### Game Session Methods
- `ping` - No parameters required
- `createPlayer` - Validates player name (string, 1-50 chars, safe characters)
- `getPlayer` - Validates session_id (UUID format)
- `listPlayers` - Validates session_id (UUID format)

### Character Management Methods
- `createCharacter` - Validates session_id, name, and class (fighter, mage, cleric, thief, ranger, paladin)
- `getCharacter` - Validates session_id and optional characterId (UUID)
- `updateCharacter` - Validates session_id and characterId (UUID)
- `listCharacters` - Validates session_id

### Movement Methods
- `move` - Validates session_id and x/y coordinates (-10000 to 10000 range)
- `getPosition` - Validates session_id

### Combat Methods
- `attack` - Validates session_id and targetId (UUID)
- `castSpell` - Validates session_id and spellId (lowercase alphanumeric with hyphens/underscores)
- `getSpells` - Validates session_id

### World Interaction Methods
- `getWorld` - Validates session_id
- `getWorldState` - Validates session_id

### Equipment Methods
- `equipItem` - Validates session_id and itemId (UUID)
- `unequipItem` - Validates session_id and optional slot (head, chest, main-hand, etc.)
- `getInventory` - Validates session_id
- `useItem` - Validates session_id, item_id, and optional target_id

### Other Methods
- `leaveGame` - Validates session_id

## Validation Rules

### String Validation

- **Length Limits**: Player/character names limited to 50 characters
- **Character Sets**: Names allow only letters, numbers, spaces, hyphens, underscores, apostrophes, and periods
- **Format Validation**: UUIDs must match 8-4-4-4-12 hex digit format

### Numeric Validation

- **Range Checking**: Coordinates must be within -10000 to 10000
- **Type Validation**: JSON numbers converted to float64 as expected

### ID Validation

- **UUID Format**: session_id, characterId, targetId, itemId must be valid UUIDs
- **Spell ID Format**: Lowercase alphanumeric with hyphens/underscores, max 100 chars

## Security Considerations

### DoS Prevention

- **Request Size Limits**: Configurable maximum request size (default 1MB)
- **Input Length Limits**: All string inputs have maximum length constraints

### Data Integrity

- **Type Safety**: Ensures data types match expectations
- **Range Validation**: Prevents invalid coordinate values
- **Format Consistency**: Validates ID formats before processing

## Integration with Server

The validation package is integrated into the JSON-RPC server pipeline:

```go
// In server.go - ValidateRPCRequest is called before processing
if err := s.validator.ValidateRPCRequest(method, params, requestSize); err != nil {
    return nil, err
}
```

## Testing

```go
func TestInputValidator(t *testing.T) {
    validator := validation.NewInputValidator(1024)
    
    // Test valid move request
    validParams := map[string]interface{}{
        "session_id": "12345678-1234-1234-1234-123456789abc",
        "x":          10.0,
        "y":          20.0,
    }
    
    err := validator.ValidateRPCRequest("move", validParams, 100)
    assert.NoError(t, err)
    
    // Test invalid session ID
    invalidParams := map[string]interface{}{
        "session_id": "invalid",
        "x":          10.0,
        "y":          20.0,
    }
    
    err = validator.ValidateRPCRequest("move", invalidParams, 100)
    assert.Error(t, err)
}
```

## Dependencies

- Standard library packages: `fmt`, `regexp`, `strings`, `unicode/utf8`
- `github.com/sirupsen/logrus`: Structured logging with caller context

Last Updated: 2026-02-19
