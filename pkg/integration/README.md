# Integration Package

This package provides integration utilities that combine retry and circuit breaker patterns for robust resilience in the GoldBox RPG Engine.

## Overview

The integration package combines the retry and circuit breaker (resilience) frameworks to provide comprehensive fault tolerance for external dependency operations. It offers utilities that apply both patterns in a unified, easy-to-use interface.

## Features

- **Unified Resilience**: Combines retry and circuit breaker patterns
- **Pre-configured Executors**: Common integration patterns for different operation types
- **Functional Options**: Customize behavior with option functions
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Monitoring Support**: Built-in statistics and logging

## Components

### ResilientExecutor

The main integration component that combines retry and circuit breaker protection:

```go
type ResilientExecutor struct {
    circuitBreaker *resilience.CircuitBreaker
    retrier        *retry.Retrier
    logger         *logrus.Entry
}
```

## Usage

### Basic Setup

```go
import "goldbox-rpg/pkg/integration"

// Create custom resilient executor
executor := integration.NewResilientExecutor(
    resilience.CircuitBreakerConfig{
        Name:        "my_service",
        MaxFailures: 5,
        Timeout:     30 * time.Second,
        MaxRequests: 3,
    },
    retry.RetryConfig{
        MaxAttempts:       3,
        InitialDelay:      100 * time.Millisecond,
        MaxDelay:          5 * time.Second,
        BackoffMultiplier: 2.0,
    },
)
```

### Executing Operations

```go
// Execute an operation with both circuit breaker and retry protection
err := executor.Execute(ctx, func(ctx context.Context) error {
    return externalService.Call()
})
```

### Using Pre-configured Executors

The package provides pre-configured executors for common operations:

```go
// File system operations
err := integration.ExecuteFileSystemOperation(ctx, func(ctx context.Context) error {
    return os.WriteFile("test.txt", data, 0644)
})

// Network operations
err := integration.ExecuteNetworkOperation(ctx, func(ctx context.Context) error {
    return sendHTTPRequest()
})

// Configuration loading
err := integration.ExecuteConfigOperation(ctx, func(ctx context.Context) error {
    return loadConfig()
})
```

## Pre-configured Executors

### FileSystemExecutor
For file system operations with appropriate retry and circuit breaker settings.

### NetworkExecutor
For network operations with settings tuned for transient network failures.

### ConfigLoaderExecutor
For configuration loading operations during startup.

## Custom Executors

### Create Custom Executor

```go
// Create with custom configuration
executor := integration.CreateCustomExecutor(
    "my_custom_service",
    resilience.CircuitBreakerConfig{
        MaxFailures: 3,
        Timeout:     10 * time.Second,
        MaxRequests: 1,
    },
    retry.RetryConfig{
        MaxAttempts:       5,
        InitialDelay:      50 * time.Millisecond,
        MaxDelay:          2 * time.Second,
        BackoffMultiplier: 2.0,
    },
)
```

### Retry-Only Executor

```go
// Create executor with circuit breaker disabled
executor := integration.WithCircuitBreakerDisabled(myRetryConfig)
```

### Circuit Breaker-Only Executor

```go
// Create executor with retry disabled
executor := integration.WithRetryDisabled(myCircuitBreakerConfig)
```

## Ad-hoc Resilient Execution

For one-off operations with optional customization:

```go
// Basic ad-hoc execution with defaults
err := integration.ExecuteResilient(ctx, myOperation)

// With custom retry configuration
err := integration.ExecuteResilient(ctx, myOperation,
    integration.ConfigureRetry(customRetryConfig),
)

// With custom circuit breaker configuration
err := integration.ExecuteResilient(ctx, myOperation,
    integration.ConfigureCircuitBreaker(customCBConfig),
)

// With both customized
err := integration.ExecuteResilient(ctx, myOperation,
    integration.ConfigureRetry(customRetryConfig),
    integration.ConfigureCircuitBreaker(customCBConfig),
)
```

## Monitoring and Statistics

```go
// Get combined statistics from circuit breaker and retry
stats := executor.GetStats()
// Returns map with keys like:
//   "circuit_breaker_state"
//   "circuit_breaker_failures"
//   "circuit_breaker_successes"
```

## Testing Support

```go
// Reset all global executors between tests
integration.ResetExecutorsForTesting()
```

## Example Usage Patterns

### Basic Resilient File Write

```go
err := integration.ExecuteFileSystemOperation(ctx, func(ctx context.Context) error {
    return os.WriteFile("test.txt", data, 0644)
})
```

### Custom Service Integration

```go
executor := integration.CreateCustomExecutor("my_service", myCircuitConfig, myRetryConfig)
err := executor.Execute(ctx, func(ctx context.Context) error {
    return myService.DoOperation()
})
```

### Ad-hoc with Options

```go
err := integration.ExecuteResilient(ctx, myOperation,
    integration.ConfigureRetry(retry.RetryConfig{
        MaxAttempts:  5,
        InitialDelay: 100 * time.Millisecond,
    }),
    integration.ConfigureCircuitBreaker(resilience.CircuitBreakerConfig{
        MaxFailures: 3,
        Timeout:     30 * time.Second,
    }),
)
```

## Dependencies

- `goldbox-rpg/pkg/resilience`: Circuit breaker patterns
- `goldbox-rpg/pkg/retry`: Retry mechanisms
- `github.com/sirupsen/logrus`: Structured logging

## Notes

- The circuit breaker is applied first (inner), then retry logic wraps around it
- This means retries will occur even if the circuit breaker trips during operation
- Use `WithRetryDisabled` if you want circuit breaker-only protection
- Use `WithCircuitBreakerDisabled` if you want retry-only protection
- Global executors are shared across the application - use `CreateCustomExecutor` for isolated instances

Last Updated: 2026-02-19
