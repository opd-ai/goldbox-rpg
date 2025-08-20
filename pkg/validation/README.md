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
- **Extensible Framework**: Easy to add new validation rules

## Components

### InputValidator

The main validation engine that:
- Maintains a registry of validation functions
- Enforces request size limits
- Provides method-specific validation
- Handles validation errors consistently

## Usage

### Basic Setup

```go
import "goldbox-rpg/pkg/validation"

// Create validator with size limit
validator := validation.NewInputValidator(1024 * 1024) // 1MB limit

// Register validation rules
validator.RegisterValidator("move", validation.ValidateMoveRequest)
validator.RegisterValidator("cast_spell", validation.ValidateSpellCastRequest)
validator.RegisterValidator("create_character", validation.ValidateCharacterCreationRequest)
```

### Validating Requests

```go
// In JSON-RPC handler
func (s *Server) handleRequest(method string, params interface{}) (interface{}, error) {
    // Validate input first
    if err := s.validator.ValidateInput(method, params); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Process validated request
    return s.processRequest(method, params)
}
```

### Custom Validators

```go
// Register custom validator
validator.RegisterValidator("custom_action", func(params interface{}) error {
    paramMap, ok := params.(map[string]interface{})
    if !ok {
        return validation.ErrInvalidParameterType
    }
    
    // Validate required fields
    if _, exists := paramMap["required_field"]; !exists {
        return validation.ErrMissingRequiredField
    }
    
    // Validate field values
    if value, ok := paramMap["numeric_field"].(float64); ok {
        if value < 0 || value > 100 {
            return validation.ErrValueOutOfRange
        }
    }
    
    return nil
})
```

## Built-in Validators

### Move Request Validation

```go
func ValidateMoveRequest(params interface{}) error {
    // Validates session_id and direction parameters
    // Ensures direction is valid (north, south, east, west)
    // Validates session ID format
}
```

### Spell Cast Validation

```go
func ValidateSpellCastRequest(params interface{}) error {
    // Validates spell_id, target coordinates, caster_id
    // Ensures spell exists and is valid
    // Validates target coordinates are within bounds
}
```

### Character Creation Validation

```go
func ValidateCharacterCreationRequest(params interface{}) error {
    // Validates character name, class, attributes
    // Ensures attribute values are within valid ranges
    // Validates class exists and is available
}
```

## Validation Rules

### String Validation

- **Length Limits**: Prevents buffer overflow attacks
- **Character Sets**: Allows only safe characters
- **Format Validation**: Ensures proper formatting (UUIDs, names, etc.)

### Numeric Validation

- **Range Checking**: Ensures values are within expected bounds
- **Type Validation**: Confirms numeric types are correct
- **Overflow Protection**: Prevents integer overflow attacks

### Collection Validation

- **Size Limits**: Prevents memory exhaustion
- **Content Validation**: Validates all elements in collections
- **Nesting Limits**: Prevents deeply nested structure attacks

## Error Types

```go
var (
    ErrInvalidParameterType   = errors.New("invalid parameter type")
    ErrMissingRequiredField   = errors.New("missing required field")
    ErrValueOutOfRange        = errors.New("value out of valid range")
    ErrInvalidFormat          = errors.New("invalid format")
    ErrRequestTooLarge        = errors.New("request too large")
    ErrValidationFailed       = errors.New("validation failed")
)
```

## Security Considerations

### Injection Prevention

- **SQL Injection**: Validates and sanitizes database queries
- **Script Injection**: Prevents script execution in user inputs
- **Path Traversal**: Validates file paths and prevents directory traversal

### DoS Prevention

- **Request Size Limits**: Prevents memory exhaustion attacks
- **Input Complexity**: Limits nested structures and array sizes
- **Rate Limiting**: Works with rate limiting system for comprehensive protection

### Data Integrity

- **Type Safety**: Ensures data types match expectations
- **Range Validation**: Prevents invalid game state values
- **Format Consistency**: Maintains consistent data formats

## Integration with Game Systems

### JSON-RPC Server Integration

```go
// In server initialization
func (s *Server) initializeValidation() {
    s.validator = validation.NewInputValidator(s.config.MaxRequestSize)
    
    // Register all validators
    s.validator.RegisterValidator("move", validation.ValidateMoveRequest)
    s.validator.RegisterValidator("attack", validation.ValidateAttackRequest)
    s.validator.RegisterValidator("cast_spell", validation.ValidateSpellCastRequest)
    // ... more validators
}
```

### Event System Integration

```go
// Validate event data
func (es *EventSystem) EmitEvent(event *Event) error {
    if err := es.validator.ValidateEventData(event.Type, event.Data); err != nil {
        return fmt.Errorf("event validation failed: %w", err)
    }
    
    return es.doEmitEvent(event)
}
```

## Testing

```go
func TestInputValidator(t *testing.T) {
    validator := validation.NewInputValidator(1024)
    
    // Test valid input
    validParams := map[string]interface{}{
        "session_id": "valid-uuid-here",
        "direction":  "north",
    }
    
    err := validator.ValidateInput("move", validParams)
    assert.NoError(t, err)
    
    // Test invalid input
    invalidParams := map[string]interface{}{
        "session_id": "invalid",
        "direction":  "invalid_direction",
    }
    
    err = validator.ValidateInput("move", invalidParams)
    assert.Error(t, err)
}
```

## Performance

- **Fast Validation**: Optimized for low-latency validation
- **Memory Efficient**: Minimal memory allocation during validation
- **Cacheable Results**: Validation rules can be cached for repeated use

## Dependencies

- Standard library packages: `fmt`, `regexp`, `strings`, `unicode/utf8`
- No external dependencies for security and reliability

Last Updated: 2025-08-20
