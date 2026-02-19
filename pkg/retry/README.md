# Retry Package

This package provides retry mechanisms with exponential backoff for handling transient failures in the GoldBox RPG Engine.

## Overview

The retry package implements robust retry strategies to handle temporary failures in external dependencies and critical operations. It uses exponential backoff with jitter to prevent thundering herd problems and provides configurable retry policies.

## Features

- **Exponential Backoff**: Progressively longer delays between retries
- **Jitter Support**: Randomization to prevent synchronized retries
- **Configurable Policies**: Customizable retry behavior per use case
- **Context Support**: Proper cancellation and timeout handling
- **Thread-Safe**: Safe for concurrent use
- **Structured Logging**: Detailed retry attempt logging
- **Pre-configured Retriers**: Default, Network, and FileSystem optimized configurations

## Components

### RetryConfig

Configuration for retry behavior:

```go
type RetryConfig struct {
    MaxAttempts       int           // Maximum number of retry attempts (including initial)
    InitialDelay      time.Duration // Initial delay before first retry
    MaxDelay          time.Duration // Maximum delay between retries
    BackoffMultiplier float64       // Multiplier for exponential backoff (typically 2.0)
    JitterMaxPercent  int           // Maximum percentage of jitter to add (0-100)
    RetryableErrors   []error       // Specific errors that should trigger retries
}
```

### Retrier

The `Retrier` type provides retry functionality:

```go
type Retrier struct {
    config RetryConfig
    logger *logrus.Entry
}

// Execute runs the given function with retry logic
func (r *Retrier) Execute(ctx context.Context, operation func(context.Context) error) error

// ExecuteWithResult runs the given function and returns error (result discarded)
func (r *Retrier) ExecuteWithResult(ctx context.Context, operation func(context.Context) (interface{}, error)) error
```

### Pre-configured Retriers

```go
var (
    DefaultRetrier    = NewRetrier(DefaultRetryConfig())    // General purpose
    NetworkRetrier    = NewRetrier(NetworkRetryConfig())    // Network operations
    FileSystemRetrier = NewRetrier(FileSystemRetryConfig()) // File system operations
)
```

## Usage

### Basic Retry with Convenience Function

```go
import (
    "context"
    "goldbox-rpg/pkg/retry"
)

ctx := context.Background()

// Use the default retrier via convenience function
err := retry.Execute(ctx, func(ctx context.Context) error {
    // Operation that might fail transiently
    return callExternalService()
})

if err != nil {
    log.Error("Operation failed after retries:", err)
}
```

### Custom Retry Configuration

```go
import "goldbox-rpg/pkg/retry"

config := retry.RetryConfig{
    MaxAttempts:       5,
    InitialDelay:      100 * time.Millisecond,
    MaxDelay:          5 * time.Second,
    BackoffMultiplier: 2.0,
    JitterMaxPercent:  15,
    RetryableErrors:   []error{context.DeadlineExceeded},
}

retrier := retry.NewRetrier(config)

ctx := context.Background()
err := retrier.Execute(ctx, func(ctx context.Context) error {
    return performOperation()
})
```

### Network Operations

```go
// Use the network-optimized retrier
ctx := context.Background()

err := retry.ExecuteNetwork(ctx, func(ctx context.Context) error {
    response, err := http.Get("https://api.example.com/data")
    if err != nil {
        return err
    }
    defer response.Body.Close()
    
    if response.StatusCode >= 500 {
        return fmt.Errorf("server error: %d", response.StatusCode)
    }
    
    return nil
})
```

### File System Operations

```go
// Use the file system-optimized retrier
ctx := context.Background()

err := retry.ExecuteFileSystem(ctx, func(ctx context.Context) error {
    return saveCharacterToFile(character)
})
```

## Pre-configured Configurations

### DefaultRetryConfig

```go
RetryConfig{
    MaxAttempts:       3,
    InitialDelay:      100 * time.Millisecond,
    MaxDelay:          30 * time.Second,
    BackoffMultiplier: 2.0,
    JitterMaxPercent:  10,
    RetryableErrors:   []error{context.DeadlineExceeded},
}
```

### NetworkRetryConfig

```go
RetryConfig{
    MaxAttempts:       5,
    InitialDelay:      200 * time.Millisecond,
    MaxDelay:          60 * time.Second,
    BackoffMultiplier: 2.0,
    JitterMaxPercent:  15,
    RetryableErrors:   []error{context.DeadlineExceeded},
}
```

### FileSystemRetryConfig

```go
RetryConfig{
    MaxAttempts:       3,
    InitialDelay:      50 * time.Millisecond,
    MaxDelay:          5 * time.Second,
    BackoffMultiplier: 1.5,
    JitterMaxPercent:  5,
    RetryableErrors:   []error{context.DeadlineExceeded},
}
```

## Advanced Usage

### With Context Timeout

```go
// Combine with context timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

config := retry.RetryConfig{
    MaxAttempts:       10, // May not reach 10 if context times out
    InitialDelay:      100 * time.Millisecond,
    MaxDelay:          2 * time.Second,
    BackoffMultiplier: 1.8,
    JitterMaxPercent:  10,
}

retrier := retry.NewRetrier(config)
err := retrier.Execute(ctx, func(ctx context.Context) error {
    return longRunningOperation()
})
```

### Conditional Retries with Specific Errors

```go
// Only retry specific errors
config := retry.RetryConfig{
    MaxAttempts:       3,
    InitialDelay:      100 * time.Millisecond,
    MaxDelay:          1 * time.Second,
    BackoffMultiplier: 2.0,
    JitterMaxPercent:  10,
    RetryableErrors: []error{
        ErrTemporaryFailure,
        context.DeadlineExceeded,
    },
}

retrier := retry.NewRetrier(config)
err := retrier.Execute(ctx, func(ctx context.Context) error {
    result, err := processGameAction()
    return err
})
```

## Integration with Game Systems

### Character Save Operations

```go
func SaveCharacterWithRetry(ctx context.Context, character *game.Character) error {
    config := retry.RetryConfig{
        MaxAttempts:       5,
        InitialDelay:      100 * time.Millisecond,
        MaxDelay:          3 * time.Second,
        BackoffMultiplier: 2.0,
        JitterMaxPercent:  10,
    }
    
    retrier := retry.NewRetrier(config)
    return retrier.Execute(ctx, func(ctx context.Context) error {
        return persistCharacter(character)
    })
}
```

### Event Processing

```go
func ProcessEventWithRetry(ctx context.Context, event *game.Event) error {
    // Use the default retrier for general operations
    return retry.Execute(ctx, func(ctx context.Context) error {
        return processGameEvent(event)
    })
}
```

## Backoff Algorithm

The exponential backoff algorithm with jitter:

```
delay = min(initialDelay * (backoffMultiplier ^ (attempt-1)), maxDelay)
if jitterMaxPercent > 0:
    jitterRange = delay * jitterMaxPercent / 100
    jitter = random(-jitterRange, +jitterRange)
    delay = delay + jitter
```

This prevents thundering herd problems where multiple clients retry simultaneously.

## Error Handling

### Retry Decision Logic

1. Check if context is cancelled or timed out
2. Execute the operation
3. If successful, return immediately
4. Check if maximum attempts reached
5. Check if error is retryable (all errors are retryable by default unless RetryableErrors is specified)
6. Calculate next delay with exponential backoff and jitter
7. Wait for delay (respecting context cancellation)
8. Retry operation

### Final Error Format

When all retries are exhausted, the error is wrapped:

```go
fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
```

## Logging

The retry mechanism provides structured logging via logrus:

- **Debug**: Entry/exit of Execute, attempt starts, operation success/failure
- **Info**: Successful recovery after retries
- **Warn**: All retry attempts exhausted
- **Error**: Context validation failures, operation execution failures

## Testing

```go
func TestRetryWithSuccess(t *testing.T) {
    attemptCount := 0
    config := retry.RetryConfig{
        MaxAttempts:       3,
        InitialDelay:      10 * time.Millisecond,
        MaxDelay:          100 * time.Millisecond,
        BackoffMultiplier: 2.0,
        JitterMaxPercent:  0, // No jitter for deterministic tests
    }
    
    retrier := retry.NewRetrier(config)
    ctx := context.Background()
    
    err := retrier.Execute(ctx, func(ctx context.Context) error {
        attemptCount++
        if attemptCount < 3 {
            return errors.New("temporary failure")
        }
        return nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 3, attemptCount)
}

func TestRetryWithContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately
    
    err := retry.Execute(ctx, func(ctx context.Context) error {
        return errors.New("should not be called")
    })
    
    assert.Error(t, err)
    assert.Equal(t, context.Canceled, err)
}
```

## Performance Considerations

- **Memory Efficient**: Minimal memory allocation during retry operations
- **CPU Friendly**: Exponential backoff reduces CPU usage under failure conditions
- **Network Considerate**: Jitter prevents network congestion from synchronized retries

## Dependencies

- `context`: For cancellation and timeout support
- `time`: For delay calculations and timing
- `math/rand`: For jitter randomization
- `github.com/sirupsen/logrus`: For structured logging

Last Updated: 2026-02-19
