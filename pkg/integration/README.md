# Integration Package

This package provides integration utilities that combine resilience and validation patterns for robust API endpoints in the GoldBox RPG Engine.

## Overview

The integration package combines the resilience (circuit breaker) and validation frameworks to provide comprehensive protection for critical game operations. It offers utilities that apply both validation and fault tolerance patterns in a unified, easy-to-use interface.

## Features

- **Unified Protection**: Combines input validation and circuit breaker patterns
- **Pre-configured Patterns**: Common integration patterns for game operations
- **Consistent Error Handling**: Unified error responses across systems
- **Performance Optimized**: Minimal overhead for protection layers
- **Thread-Safe**: Safe for concurrent use across multiple sessions
- **Monitoring Integration**: Built-in metrics and logging

## Components

### ResilientValidator

The main integration component that combines validation and circuit breaker protection:

```go
type ResilientValidator struct {
    validator       *validation.InputValidator
    circuitBreaker  *resilience.CircuitBreaker
    config          Config
}
```

### Integration Config

Configuration for the integrated protection:

```go
type Config struct {
    ValidationConfig  validation.Config
    ResilienceConfig  resilience.Config
    EnableMetrics     bool
    LoggingEnabled    bool
}
```

## Usage

### Basic Setup

```go
import "goldbox-rpg/pkg/integration"

// Create integrated protection
config := integration.Config{
    ValidationConfig: validation.Config{
        MaxRequestSize: 1024 * 1024, // 1MB
    },
    ResilienceConfig: resilience.Config{
        FailureThreshold: 5,
        RecoveryTimeout:  30 * time.Second,
        MaxRequests:      3,
    },
    EnableMetrics:    true,
    LoggingEnabled:   true,
}

resilientValidator := integration.NewResilientValidator("game_api", config)
```

### JSON-RPC Handler Integration

```go
func (s *Server) handleRequestWithProtection(method string, params interface{}) (interface{}, error) {
    // Apply both validation and circuit breaker protection
    return s.resilientValidator.ExecuteWithValidation(method, params, func() (interface{}, error) {
        // Your protected operation here
        return s.processGameAction(method, params)
    })
}
```

### Game Operation Protection

```go
// Protect character creation
func (s *Server) createCharacterProtected(params interface{}) (interface{}, error) {
    return s.resilientValidator.ExecuteWithValidation("create_character", params, func() (interface{}, error) {
        // Parse validated parameters
        characterParams := params.(map[string]interface{})
        
        // Create character
        character, err := s.gameEngine.CreateCharacter(characterParams)
        if err != nil {
            return nil, err
        }
        
        return map[string]interface{}{
            "character_id": character.ID,
            "success":      true,
        }, nil
    })
}
```

## Pre-configured Patterns

### Database Operations

```go
// Pre-configured protection for database operations
dbProtection := integration.NewDatabaseProtection(integration.DatabaseConfig{
    ValidationTimeout: 5 * time.Second,
    FailureThreshold:  3,
    RecoveryTimeout:   10 * time.Second,
})

err := dbProtection.ExecuteQuery("save_character", params, func() error {
    return database.SaveCharacter(character)
})
```

### External API Calls

```go
// Pre-configured protection for external APIs
apiProtection := integration.NewAPIProtection(integration.APIConfig{
    MaxRequestSize:    512 * 1024, // 512KB
    FailureThreshold:  5,
    RecoveryTimeout:   30 * time.Second,
    RequestTimeout:    10 * time.Second,
})

response, err := apiProtection.CallAPI("spell_validation", spellData, func() (interface{}, error) {
    return callSpellValidationService(spellData)
})
```

### Game State Operations

```go
// Pre-configured protection for critical game state changes
gameStateProtection := integration.NewGameStateProtection(integration.GameStateConfig{
    ValidationRules:   gameStateValidationRules,
    FailureThreshold:  2, // Lower threshold for critical operations
    RecoveryTimeout:   5 * time.Second,
    EnableAuditLog:    true,
})

err := gameStateProtection.UpdateGameState("player_move", moveData, func() error {
    return gameEngine.ProcessPlayerMove(moveData)
})
```

## Advanced Features

### Conditional Circuit Breaking

```go
// Different circuit breaker behavior based on operation type
validator := integration.NewResilientValidator("adaptive", config)

// Configure different thresholds for different operations
validator.SetOperationConfig("critical_operation", integration.OperationConfig{
    FailureThreshold: 1, // Fail fast for critical operations
    RecoveryTimeout:  60 * time.Second,
})

validator.SetOperationConfig("non_critical_operation", integration.OperationConfig{
    FailureThreshold: 10, // More tolerant for non-critical operations
    RecoveryTimeout:  5 * time.Second,
})
```

### Custom Validation + Resilience

```go
// Combine custom validation with resilience
validator := integration.NewResilientValidator("custom", config)

// Register custom validation
validator.RegisterCustomValidator("special_action", func(params interface{}) error {
    // Custom validation logic
    return validateSpecialAction(params)
})

// Use with automatic resilience
result, err := validator.ExecuteWithValidation("special_action", actionParams, func() (interface{}, error) {
    return processSpecialAction(actionParams)
})
```

## Monitoring and Metrics

### Built-in Metrics

The integration package automatically tracks:
- Validation success/failure rates
- Circuit breaker state changes
- Operation latencies
- Error frequencies by type

### Custom Metrics

```go
// Add custom metrics
validator.RegisterMetric("custom_operation_duration", func(duration time.Duration) {
    customMetrics.ObserveDuration(duration)
})

validator.RegisterMetric("custom_validation_errors", func(errorType string) {
    customMetrics.IncrementValidationErrors(errorType)
})
```

### Health Checks

```go
// Check health of integrated systems
health := validator.GetHealthStatus()
if !health.IsHealthy {
    log.Warn("Integrated protection system is unhealthy:", health.Issues)
}

// Example health check response
type HealthStatus struct {
    IsHealthy           bool                    `json:"is_healthy"`
    ValidationStatus    validation.Health      `json:"validation_status"`
    CircuitBreakerStatus resilience.Health     `json:"circuit_breaker_status"`
    Issues              []string               `json:"issues,omitempty"`
    LastChecked         time.Time              `json:"last_checked"`
}
```

## Error Handling

### Unified Error Response

```go
type IntegrationError struct {
    Type          string    `json:"type"`           // "validation" or "resilience"
    Code          string    `json:"code"`           // Specific error code
    Message       string    `json:"message"`        // Human-readable message
    OriginalError error     `json:"-"`              // Original error (not serialized)
    Timestamp     time.Time `json:"timestamp"`      // When error occurred
    Context       map[string]interface{} `json:"context,omitempty"` // Additional context
}
```

### Error Categories

- **Validation Errors**: Input validation failures
- **Circuit Breaker Errors**: Circuit open, too many requests
- **Operation Errors**: Errors from the protected operation itself
- **Configuration Errors**: Misconfiguration of protection layers

## Testing

### Unit Testing

```go
func TestResilientValidator(t *testing.T) {
    config := integration.Config{
        ValidationConfig: validation.Config{MaxRequestSize: 1024},
        ResilienceConfig: resilience.Config{FailureThreshold: 2},
    }
    
    validator := integration.NewResilientValidator("test", config)
    
    // Test successful operation
    result, err := validator.ExecuteWithValidation("test_method", validParams, func() (interface{}, error) {
        return "success", nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "success", result)
}
```

### Integration Testing

```go
func TestFullStackProtection(t *testing.T) {
    // Test the full protection stack
    server := setupTestServer()
    
    // Send valid request
    response := sendJSONRPCRequest("create_character", validCharacterData)
    assert.Equal(t, "success", response["status"])
    
    // Send invalid request
    response = sendJSONRPCRequest("create_character", invalidCharacterData)
    assert.Equal(t, "error", response["status"])
    assert.Contains(t, response["error"], "validation")
    
    // Trigger circuit breaker
    for i := 0; i < 10; i++ {
        sendJSONRPCRequest("failing_operation", nil)
    }
    
    response = sendJSONRPCRequest("failing_operation", nil)
    assert.Contains(t, response["error"], "circuit")
}
```

## Performance Considerations

- **Minimal Overhead**: Optimized for low-latency game operations
- **Memory Efficient**: Reuses validation and circuit breaker instances
- **Concurrent Safe**: All components are thread-safe
- **Configurable Trade-offs**: Balance between protection and performance

## Dependencies

- `goldbox-rpg/pkg/validation`: Input validation framework
- `goldbox-rpg/pkg/resilience`: Circuit breaker patterns
- `goldbox-rpg/pkg/retry`: Retry mechanisms (optional)
- `github.com/sirupsen/logrus`: Structured logging
- `time`: Timing and timeout operations

## Best Practices

1. **Layer Protection**: Use integration patterns for all external-facing APIs
2. **Configure Appropriately**: Different operations need different protection levels
3. **Monitor Health**: Regular health checks of protection systems
4. **Test Thoroughly**: Include protection testing in integration tests
5. **Document Configurations**: Clear documentation of protection configurations

Last Updated: 2025-08-20
