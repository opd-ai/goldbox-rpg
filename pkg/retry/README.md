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

## Components

### RetryConfig

Configuration for retry behavior:

```go
type RetryConfig struct {
    MaxAttempts     int           // Maximum number of retry attempts
    InitialDelay    time.Duration // Initial delay before first retry
    MaxDelay        time.Duration // Maximum delay between retries
    BackoffFactor   float64       // Multiplier for exponential backoff
    JitterEnabled   bool          // Whether to add randomization
    RetryableErrors []error       // Specific errors that should trigger retries
}
```

### Retry Function

The main retry execution function with exponential backoff:

```go
func WithRetry(ctx context.Context, config RetryConfig, operation func() error) error
```

## Usage

### Basic Retry

```go
import "goldbox-rpg/pkg/retry"

config := retry.RetryConfig{
    MaxAttempts:   3,
    InitialDelay:  100 * time.Millisecond,
    MaxDelay:      5 * time.Second,
    BackoffFactor: 2.0,
    JitterEnabled: true,
}

ctx := context.Background()
err := retry.WithRetry(ctx, config, func() error {
    // Operation that might fail transiently
    return callExternalService()
})

if err != nil {
    log.Error("Operation failed after retries:", err)
}
```

### Database Operations

```go
// Retry database operations
dbConfig := retry.RetryConfig{
    MaxAttempts:   5,
    InitialDelay:  50 * time.Millisecond,
    MaxDelay:      2 * time.Second,
    BackoffFactor: 1.5,
    JitterEnabled: true,
    RetryableErrors: []error{
        sql.ErrConnDone,
        context.DeadlineExceeded,
    },
}

err := retry.WithRetry(ctx, dbConfig, func() error {
    return saveCharacterToDatabase(character)
})
```

### Network Operations

```go
// Retry network calls
networkConfig := retry.RetryConfig{
    MaxAttempts:   4,
    InitialDelay:  200 * time.Millisecond,
    MaxDelay:      10 * time.Second,
    BackoffFactor: 2.0,
    JitterEnabled: true,
}

err := retry.WithRetry(ctx, networkConfig, func() error {
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

## Advanced Usage

### Conditional Retries

```go
// Only retry specific errors
config := retry.RetryConfig{
    MaxAttempts:   3,
    InitialDelay:  100 * time.Millisecond,
    MaxDelay:      1 * time.Second,
    BackoffFactor: 2.0,
    RetryableErrors: []error{
        ErrTemporaryFailure,
        ErrRateLimited,
    },
}

err := retry.WithRetry(ctx, config, func() error {
    result, err := processGameAction()
    if err != nil {
        // Only certain errors will trigger retries
        return err
    }
    return nil
})
```

### With Context Timeout

```go
// Combine with context timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

config := retry.RetryConfig{
    MaxAttempts:   10, // May not reach 10 if context times out
    InitialDelay:  100 * time.Millisecond,
    MaxDelay:      2 * time.Second,
    BackoffFactor: 1.8,
    JitterEnabled: true,
}

err := retry.WithRetry(ctx, config, func() error {
    return longRunningOperation()
})
```

## Integration with Game Systems

### Spell Validation

```go
func ValidateSpellWithRetry(ctx context.Context, spell *game.Spell) error {
    config := retry.RetryConfig{
        MaxAttempts:   3,
        InitialDelay:  50 * time.Millisecond,
        MaxDelay:      500 * time.Millisecond,
        BackoffFactor: 2.0,
        JitterEnabled: true,
    }
    
    return retry.WithRetry(ctx, config, func() error {
        return validateSpellRules(spell)
    })
}
```

### Character Save Operations

```go
func SaveCharacterWithRetry(ctx context.Context, character *game.Character) error {
    config := retry.RetryConfig{
        MaxAttempts:   5,
        InitialDelay:  100 * time.Millisecond,
        MaxDelay:      3 * time.Second,
        BackoffFactor: 2.0,
        JitterEnabled: true,
    }
    
    return retry.WithRetry(ctx, config, func() error {
        return persistCharacter(character)
    })
}
```

### Event Processing

```go
func ProcessEventWithRetry(ctx context.Context, event *game.Event) error {
    config := retry.RetryConfig{
        MaxAttempts:   3,
        InitialDelay:  25 * time.Millisecond,
        MaxDelay:      1 * time.Second,
        BackoffFactor: 2.0,
        JitterEnabled: true,
    }
    
    return retry.WithRetry(ctx, config, func() error {
        return processGameEvent(event)
    })
}
```

## Backoff Algorithm

The exponential backoff algorithm with jitter:

```
delay = min(initialDelay * (backoffFactor ^ attempt), maxDelay)
if jitterEnabled:
    delay = delay * (0.5 + random(0, 0.5))
```

This prevents thundering herd problems where multiple clients retry simultaneously.

## Error Handling

### Retry Decision Logic

1. Check if maximum attempts reached
2. Check if error is in retryable errors list (if specified)
3. Check if context is cancelled or timed out
4. Calculate next delay and schedule retry

### Error Types

```go
var (
    ErrMaxAttemptsReached = errors.New("maximum retry attempts reached")
    ErrContextCancelled   = errors.New("context cancelled during retry")
    ErrNonRetryableError  = errors.New("error is not retryable")
)
```

## Logging

The retry mechanism provides structured logging:

```go
logrus.WithFields(logrus.Fields{
    "attempt":    attempt,
    "max_attempts": config.MaxAttempts,
    "delay":      delay,
    "error":      err,
}).Warn("Retry attempt failed, retrying...")
```

## Testing

```go
func TestRetryWithSuccess(t *testing.T) {
    attemptCount := 0
    config := retry.RetryConfig{
        MaxAttempts:   3,
        InitialDelay:  10 * time.Millisecond,
        MaxDelay:      100 * time.Millisecond,
        BackoffFactor: 2.0,
    }
    
    ctx := context.Background()
    err := retry.WithRetry(ctx, config, func() error {
        attemptCount++
        if attemptCount < 3 {
            return errors.New("temporary failure")
        }
        return nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 3, attemptCount)
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

Last Updated: 2025-08-20
