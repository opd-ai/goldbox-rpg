# Resilience Package

This package provides circuit breaker patterns and system resilience components for the GoldBox RPG Engine.

## Overview

The resilience package implements fault tolerance patterns to protect against cascade failures and improve system stability. It includes circuit breaker functionality to prevent overloading external dependencies and automatic recovery mechanisms.

## Features

- **Circuit Breaker Pattern**: Protects against cascade failures
- **Automatic Recovery**: Self-healing mechanisms for transient failures  
- **Configurable Thresholds**: Customizable failure detection and recovery
- **Thread-Safe Operations**: Safe for concurrent use
- **Metrics Integration**: Performance monitoring and alerting

## Components

### CircuitBreaker

The main circuit breaker implementation with three states:
- **Closed**: Normal operation, requests pass through
- **Open**: Failure mode, requests fail fast
- **Half-Open**: Recovery testing, limited requests allowed

### Manager

Centralized management of multiple circuit breakers with:
- Registration and lifecycle management
- Configuration management
- Monitoring and metrics collection

## Usage

### Basic Circuit Breaker

```go
import "goldbox-rpg/pkg/resilience"

// Create a circuit breaker
cb := resilience.NewCircuitBreaker("external_api", resilience.Config{
    FailureThreshold: 5,
    RecoveryTimeout:  30 * time.Second,
    MaxRequests:      3,
})

// Use the circuit breaker
err := cb.Execute(func() error {
    // Call external service
    return callExternalAPI()
})

if err != nil {
    // Handle failure (could be original error or circuit breaker error)
    log.Error("Request failed:", err)
}
```

### Circuit Breaker Manager

```go
// Create manager
manager := resilience.NewManager()

// Register circuit breakers
err := manager.RegisterCircuitBreaker("database", resilience.Config{
    FailureThreshold: 3,
    RecoveryTimeout:  10 * time.Second,
})

err = manager.RegisterCircuitBreaker("cache", resilience.Config{
    FailureThreshold: 5,
    RecoveryTimeout:  5 * time.Second,
})

// Use through manager
err = manager.Execute("database", func() error {
    return queryDatabase()
})
```

## Configuration

### CircuitBreaker Config

```go
type Config struct {
    FailureThreshold int           // Number of failures before opening
    RecoveryTimeout  time.Duration // Time to wait before half-open
    MaxRequests      int           // Max requests in half-open state
    OnStateChange    func(state CircuitBreakerState) // State change callback
}
```

### Integration with Game Systems

```go
// Example: Protecting spell validation
spellValidatorCB := resilience.NewCircuitBreaker("spell_validator", config)

func ValidateSpell(spell *game.Spell) error {
    return spellValidatorCB.Execute(func() error {
        return expensiveSpellValidation(spell)
    })
}
```

## Thread Safety

All components in this package are thread-safe and can be used safely in concurrent environments, following the established patterns from the main game engine.

## Error Handling

The package provides structured error types:
- `ErrCircuitOpen`: Circuit breaker is open
- `ErrTooManyRequests`: Too many requests in half-open state
- `ErrTimeout`: Operation timeout

## Testing

```go
func TestCircuitBreaker(t *testing.T) {
    cb := resilience.NewCircuitBreaker("test", resilience.Config{
        FailureThreshold: 2,
        RecoveryTimeout:  100 * time.Millisecond,
    })
    
    // Test normal operation
    err := cb.Execute(func() error { return nil })
    assert.NoError(t, err)
    
    // Test failure threshold
    for i := 0; i < 3; i++ {
        cb.Execute(func() error { return errors.New("failure") })
    }
    
    // Circuit should be open
    err = cb.Execute(func() error { return nil })
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circuit open")
}
```

## Dependencies

- `sync`: Thread-safe operations
- `time`: Timeout and recovery timing
- `github.com/sirupsen/logrus`: Structured logging

Last Updated: 2025-08-20
