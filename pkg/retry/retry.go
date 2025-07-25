// Package retry provides configurable retry mechanisms with exponential backoff
// for transient failures. It integrates with circuit breakers and respects
// context deadlines to provide resilient operation handling.
package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// RetryConfig holds configuration for retry operations
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts (including initial attempt)
	MaxAttempts int

	// InitialDelay is the initial delay before the first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// BackoffMultiplier is the multiplier for exponential backoff (typically 2.0)
	BackoffMultiplier float64

	// JitterMaxPercent is the maximum percentage of jitter to add (0-100)
	JitterMaxPercent int

	// RetryableErrors are error types that should trigger a retry
	RetryableErrors []error
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterMaxPercent:  10,
		RetryableErrors:   []error{context.DeadlineExceeded},
	}
}

// NetworkRetryConfig returns retry configuration optimized for network operations
func NetworkRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       5,
		InitialDelay:      200 * time.Millisecond,
		MaxDelay:          60 * time.Second,
		BackoffMultiplier: 2.0,
		JitterMaxPercent:  15,
		RetryableErrors:   []error{context.DeadlineExceeded},
	}
}

// FileSystemRetryConfig returns retry configuration optimized for file system operations
func FileSystemRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      50 * time.Millisecond,
		MaxDelay:          5 * time.Second,
		BackoffMultiplier: 1.5,
		JitterMaxPercent:  5,
		RetryableErrors:   []error{context.DeadlineExceeded},
	}
}

// Retrier provides retry functionality with exponential backoff
type Retrier struct {
	config RetryConfig
	logger *logrus.Entry
}

// NewRetrier creates a new retrier with the given configuration
func NewRetrier(config RetryConfig) *Retrier {
	return &Retrier{
		config: config,
		logger: logrus.WithField("component", "Retrier"),
	}
}

// Execute runs the given function with retry logic and exponential backoff
func (r *Retrier) Execute(ctx context.Context, operation func(context.Context) error) error {
	return r.ExecuteWithResult(ctx, func(ctx context.Context) (interface{}, error) {
		err := operation(ctx)
		return nil, err
	})
}

// ExecuteWithResult runs the given function with retry logic and returns both result and error
func (r *Retrier) ExecuteWithResult(ctx context.Context, operation func(context.Context) (interface{}, error)) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		logger := r.createAttemptLogger(attempt)

		if err := r.validateContext(ctx, logger); err != nil {
			return err
		}

		if err := r.executeOperation(ctx, operation, logger, attempt, &lastErr); err != nil {
			return err
		}

		if lastErr == nil {
			return nil
		}

		if r.shouldStopRetrying(attempt, lastErr, logger) {
			break
		}

		if err := r.waitForRetry(ctx, attempt, logger); err != nil {
			return err
		}
	}

	return r.createFinalError(lastErr)
}

// createAttemptLogger creates a logger with attempt context
func (r *Retrier) createAttemptLogger(attempt int) *logrus.Entry {
	return r.logger.WithFields(logrus.Fields{
		"attempt":      attempt,
		"max_attempts": r.config.MaxAttempts,
	})
}

// validateContext checks if the context is still valid before attempting operation
func (r *Retrier) validateContext(ctx context.Context, logger *logrus.Entry) error {
	if ctx.Err() != nil {
		logger.Debug("Context cancelled before retry attempt")
		return ctx.Err()
	}
	return nil
}

// executeOperation executes the operation and handles success/failure logging
func (r *Retrier) executeOperation(ctx context.Context, operation func(context.Context) (interface{}, error), logger *logrus.Entry, attempt int, lastErr *error) error {
	logger.Debug("Executing operation attempt")

	_, err := operation(ctx)
	*lastErr = err

	if err == nil {
		if attempt > 1 {
			logger.WithField("total_attempts", attempt).Info("Operation succeeded after retry")
		}
		return nil
	}

	logger.WithError(err).Debug("Operation failed")
	return nil
}

// shouldStopRetrying determines if retry attempts should stop
func (r *Retrier) shouldStopRetrying(attempt int, lastErr error, logger *logrus.Entry) bool {
	if attempt == r.config.MaxAttempts {
		logger.WithError(lastErr).Warn("All retry attempts exhausted")
		return true
	}

	if !r.isRetryable(lastErr) {
		logger.WithError(lastErr).Debug("Error is not retryable, stopping")
		return true
	}

	return false
}

// waitForRetry handles the delay between retry attempts with context cancellation support
func (r *Retrier) waitForRetry(ctx context.Context, attempt int, logger *logrus.Entry) error {
	delay := r.calculateDelay(attempt)
	logger.WithField("delay", delay).Debug("Waiting before retry")

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		logger.Debug("Context cancelled during retry delay")
		return ctx.Err()
	}
}

// createFinalError wraps the last error with retry context
func (r *Retrier) createFinalError(lastErr error) error {
	return fmt.Errorf("operation failed after %d attempts: %w", r.config.MaxAttempts, lastErr)
}

// isRetryable checks if an error should trigger a retry
func (r *Retrier) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check against configured retryable errors
	for _, retryableErr := range r.config.RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	// Default retryable conditions - most errors are retryable by default
	// unless they are explicitly non-retryable
	return true
}

// calculateDelay calculates the delay for a given attempt with exponential backoff and jitter
func (r *Retrier) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff: InitialDelay * BackoffMultiplier^(attempt-1)
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.BackoffMultiplier, float64(attempt-1))

	// Apply maximum delay limit
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if r.config.JitterMaxPercent > 0 {
		jitterRange := delay * float64(r.config.JitterMaxPercent) / 100.0
		jitter := (rand.Float64() - 0.5) * 2 * jitterRange // Random between -jitterRange and +jitterRange
		delay += jitter

		// Ensure delay is never negative
		if delay < 0 {
			delay = float64(r.config.InitialDelay)
		}
	}

	return time.Duration(delay)
}

// Helper functions for error classification

// isTimeoutError checks if an error is timeout-related
func isTimeoutError(err error) bool {
	type timeout interface {
		Timeout() bool
	}

	if timeout, ok := err.(timeout); ok {
		return timeout.Timeout()
	}

	return errors.Is(err, context.DeadlineExceeded)
}

// Global retry instances for common use cases
var (
	DefaultRetrier    = NewRetrier(DefaultRetryConfig())
	NetworkRetrier    = NewRetrier(NetworkRetryConfig())
	FileSystemRetrier = NewRetrier(FileSystemRetryConfig())
)

// Convenience functions for common operations

// Execute runs an operation with default retry configuration
func Execute(ctx context.Context, operation func(context.Context) error) error {
	return DefaultRetrier.Execute(ctx, operation)
}

// ExecuteNetwork runs an operation with network-optimized retry configuration
func ExecuteNetwork(ctx context.Context, operation func(context.Context) error) error {
	return NetworkRetrier.Execute(ctx, operation)
}

// ExecuteFileSystem runs an operation with file system-optimized retry configuration
func ExecuteFileSystem(ctx context.Context, operation func(context.Context) error) error {
	return FileSystemRetrier.Execute(ctx, operation)
}
