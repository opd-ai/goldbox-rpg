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

### CircuitBreakerManager

Centralized management of multiple circuit breakers with:
- GetOrCreate for on-demand circuit breaker creation
- Lifecycle management (Get, Remove)
- Statistics collection (GetAllStats, GetStats)
- Bulk operations (ResetAll, GetBreakerNames)

## Usage

### Basic Circuit Breaker

```go
import (
    "context"
    "goldbox-rpg/pkg/resilience"
)

// Create a circuit breaker with configuration
config := resilience.CircuitBreakerConfig{
    Name:        "external_api",
    MaxFailures: 5,
    Timeout:     30 * time.Second,
    MaxRequests: 3,
}
cb := resilience.NewCircuitBreaker(config)

// Use the circuit breaker (requires context)
ctx := context.Background()
err := cb.Execute(ctx, func(ctx context.Context) error {
    // Call external service
    return callExternalAPI(ctx)
})

if err != nil {
    if errors.Is(err, resilience.ErrCircuitBreakerOpen) {
        log.Warn("Circuit breaker is open, service unavailable")
    } else {
        log.Error("Request failed:", err)
    }
}
```

### Using Default Configuration

```go
// Create with sensible defaults
config := resilience.DefaultCircuitBreakerConfig("my_service")
cb := resilience.NewCircuitBreaker(config)
```

### Circuit Breaker Manager

```go
// Create manager
manager := resilience.NewCircuitBreakerManager()

// Get or create circuit breakers with custom config
dbConfig := &resilience.CircuitBreakerConfig{
    MaxFailures: 3,
    Timeout:     10 * time.Second,
    MaxRequests: 2,
}
dbBreaker := manager.GetOrCreate("database", dbConfig)

// Get or create with default config (pass nil)
cacheBreaker := manager.GetOrCreate("cache", nil)

// Use the circuit breaker directly
ctx := context.Background()
err := dbBreaker.Execute(ctx, func(ctx context.Context) error {
    return queryDatabase(ctx)
})

// Check if a breaker exists
if cb, exists := manager.Get("database"); exists {
    stats := cb.GetStats()
    log.Info("Database circuit breaker stats:", stats)
}

// Get all statistics
allStats := manager.GetAllStats()

// Reset all circuit breakers
manager.ResetAll()

// Remove a circuit breaker
manager.Remove("cache")
```

### Global Manager and Helper Functions

```go
// Access the global circuit breaker manager
globalManager := resilience.GetGlobalCircuitBreakerManager()

// Use predefined helper functions for common dependencies
ctx := context.Background()

// File system operations
err := resilience.ExecuteWithFileSystemCircuitBreaker(ctx, func(ctx context.Context) error {
    return readConfigFile(ctx)
})

// WebSocket operations
err = resilience.ExecuteWithWebSocketCircuitBreaker(ctx, func(ctx context.Context) error {
    return sendWebSocketMessage(ctx)
})

// Config loader operations
err = resilience.ExecuteWithConfigLoaderCircuitBreaker(ctx, func(ctx context.Context) error {
    return loadConfiguration(ctx)
})
```

## Configuration

### CircuitBreakerConfig

```go
type CircuitBreakerConfig struct {
    Name        string        // Identifier for this circuit breaker
    MaxFailures int           // Number of failures before opening circuit
    Timeout     time.Duration // Time to wait before transitioning to half-open
    MaxRequests int           // Max requests allowed in half-open state
}
```

### Predefined Configurations

The package provides predefined configurations for common use cases:

```go
// File system operations (conservative)
FileSystemConfig = CircuitBreakerConfig{
    Name:        "filesystem",
    MaxFailures: 3,
    Timeout:     10 * time.Second,
    MaxRequests: 2,
}

// WebSocket operations (moderate)
WebSocketConfig = CircuitBreakerConfig{
    Name:        "websocket",
    MaxFailures: 5,
    Timeout:     30 * time.Second,
    MaxRequests: 3,
}

// Config loader operations (strict)
ConfigLoaderConfig = CircuitBreakerConfig{
    Name:        "config_loader",
    MaxFailures: 2,
    Timeout:     15 * time.Second,
    MaxRequests: 1,
}
```

### Integration with Game Systems

```go
// Example: Protecting spell validation
config := resilience.CircuitBreakerConfig{
    Name:        "spell_validator",
    MaxFailures: 3,
    Timeout:     5 * time.Second,
    MaxRequests: 2,
}
spellValidatorCB := resilience.NewCircuitBreaker(config)

func ValidateSpell(ctx context.Context, spell *game.Spell) error {
    return spellValidatorCB.Execute(ctx, func(ctx context.Context) error {
        return expensiveSpellValidation(ctx, spell)
    })
}
```

## Circuit Breaker States

```go
const (
    StateClosed   CircuitBreakerState = iota // Normal operation
    StateOpen                                 // Failing fast
    StateHalfOpen                             // Testing recovery
)
```

## Thread Safety

All components in this package are thread-safe and can be used safely in concurrent environments, following the established patterns from the main game engine.

## Error Handling

The package provides the following error type:

```go
var ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
```

When the circuit breaker is open, Execute returns an error wrapping `ErrCircuitBreakerOpen` with the circuit breaker name:

```go
err := cb.Execute(ctx, fn)
if errors.Is(err, resilience.ErrCircuitBreakerOpen) {
    // Handle circuit open state
}
```

## Circuit Breaker Methods

### CircuitBreaker

| Method | Description |
|--------|-------------|
| `Execute(ctx, fn)` | Execute function with circuit breaker protection |
| `GetState()` | Get current state (Closed, Open, HalfOpen) |
| `GetStats()` | Get statistics map (name, state, failures, etc.) |
| `Reset()` | Force circuit breaker back to closed state |

### CircuitBreakerManager

| Method | Description |
|--------|-------------|
| `GetOrCreate(name, config)` | Get existing or create new circuit breaker |
| `Get(name)` | Get circuit breaker by name (returns bool exists) |
| `Remove(name)` | Remove circuit breaker from manager |
| `GetAllStats()` | Get statistics for all managed circuit breakers |
| `ResetAll()` | Reset all circuit breakers to closed state |
| `GetBreakerNames()` | Get list of all circuit breaker names |

## Testing

```go
func TestCircuitBreaker(t *testing.T) {
    config := resilience.CircuitBreakerConfig{
        Name:        "test",
        MaxFailures: 2,
        Timeout:     100 * time.Millisecond,
        MaxRequests: 1,
    }
    cb := resilience.NewCircuitBreaker(config)
    ctx := context.Background()
    
    // Test normal operation
    err := cb.Execute(ctx, func(ctx context.Context) error { return nil })
    assert.NoError(t, err)
    
    // Test failure threshold
    for i := 0; i < 3; i++ {
        cb.Execute(ctx, func(ctx context.Context) error { 
            return errors.New("failure") 
        })
    }
    
    // Circuit should be open
    err = cb.Execute(ctx, func(ctx context.Context) error { return nil })
    assert.Error(t, err)
    assert.True(t, errors.Is(err, resilience.ErrCircuitBreakerOpen))
    
    // Wait for timeout and test half-open
    time.Sleep(150 * time.Millisecond)
    err = cb.Execute(ctx, func(ctx context.Context) error { return nil })
    assert.NoError(t, err)
}
```

## Dependencies

- `context`: Context-aware execution with cancellation support
- `sync`: Thread-safe operations
- `time`: Timeout and recovery timing
- `github.com/sirupsen/logrus`: Structured logging

Last Updated: 2026-02-19
