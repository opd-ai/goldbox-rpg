// Package retry provides configurable retry mechanisms with exponential backoff.
//
// This package handles transient failures by automatically retrying operations
// with configurable delays, backoff strategies, and jitter to prevent thundering
// herd problems.
//
// # Configuration
//
// Create a Retrier with custom retry policy:
//
//	config := retry.RetryConfig{
//	    MaxAttempts:       5,
//	    InitialDelay:      100 * time.Millisecond,
//	    MaxDelay:          30 * time.Second,
//	    BackoffMultiplier: 2.0,
//	    JitterMaxPercent:  25,
//	}
//	retrier := retry.NewRetrier(config)
//
// # Executing with Retry
//
// Wrap operations with automatic retry on failure:
//
//	err := retrier.Execute(ctx, func() error {
//	    return callUnreliableService()
//	})
//
// For operations that return a value:
//
//	result, err := retrier.ExecuteWithResult(ctx, func() (any, error) {
//	    return fetchData()
//	})
//
// # Backoff Strategy
//
// Delays increase exponentially between retries:
//
//	Attempt 1: InitialDelay (100ms)
//	Attempt 2: InitialDelay * BackoffMultiplier (200ms)
//	Attempt 3: Previous * BackoffMultiplier (400ms)
//	...up to MaxDelay
//
// Jitter is applied to prevent synchronized retries across clients.
//
// # Pre-configured Retriers
//
// Global retriers with common configurations:
//
//	// Default: 3 attempts, 100ms initial delay
//	err := retry.Execute(ctx, operation)
//
//	// Network: 5 attempts, 200ms initial, 60s max
//	err := retry.ExecuteNetwork(ctx, operation)
//
//	// File system: 3 attempts, 50ms initial, 5s max
//	err := retry.ExecuteFileSystem(ctx, operation)
//
// # Retryable Errors
//
// By default, all errors trigger retry. Configure specific retryable errors:
//
//	config.RetryableErrors = []error{
//	    syscall.ECONNREFUSED,
//	    io.ErrUnexpectedEOF,
//	}
//
// # Context Support
//
// Retries respect context cancellation and deadlines:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	err := retrier.Execute(ctx, operation)
//
// # Logging
//
// Retry attempts are logged with structured context including attempt number,
// delay duration, and error details for debugging transient failures.
package retry
