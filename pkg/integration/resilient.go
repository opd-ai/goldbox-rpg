// Package integration provides integration between retry and circuit breaker patterns
// for comprehensive resilience in external dependency operations.
package integration

import (
	"context"

	"goldbox-rpg/pkg/resilience"
	"goldbox-rpg/pkg/retry"

	"github.com/sirupsen/logrus"
)

// ResilientExecutor combines circuit breaker and retry patterns for maximum resilience
type ResilientExecutor struct {
	circuitBreaker *resilience.CircuitBreaker
	retrier        *retry.Retrier
	logger         *logrus.Entry
}

// NewResilientExecutor creates a new executor combining circuit breaker and retry patterns
func NewResilientExecutor(cbConfig resilience.CircuitBreakerConfig, retryConfig retry.RetryConfig) *ResilientExecutor {
	return &ResilientExecutor{
		circuitBreaker: resilience.NewCircuitBreaker(cbConfig),
		retrier:        retry.NewRetrier(retryConfig),
		logger:         logrus.WithField("component", "ResilientExecutor"),
	}
}

// Execute runs an operation with both circuit breaker and retry protection
func (re *ResilientExecutor) Execute(ctx context.Context, operation func(context.Context) error) error {
	// Wrap the operation with circuit breaker protection first
	wrappedOperation := func(ctx context.Context) error {
		return re.circuitBreaker.Execute(ctx, operation)
	}

	// Then apply retry logic around the circuit breaker
	return re.retrier.Execute(ctx, wrappedOperation)
}

// GetStats returns statistics from both circuit breaker and retry operations
func (re *ResilientExecutor) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Add circuit breaker stats
	cbStats := re.circuitBreaker.GetStats()
	for key, value := range cbStats {
		stats["circuit_breaker_"+key] = value
	}

	return stats
}

// Predefined resilient executors for common operations
var (
	// FileSystemExecutor provides resilient file system operations
	FileSystemExecutor = NewResilientExecutor(
		resilience.FileSystemConfig,
		retry.FileSystemRetryConfig(),
	)

	// NetworkExecutor provides resilient network operations
	NetworkExecutor = NewResilientExecutor(
		resilience.WebSocketConfig,
		retry.NetworkRetryConfig(),
	)

	// ConfigLoaderExecutor provides resilient configuration loading
	ConfigLoaderExecutor = NewResilientExecutor(
		resilience.ConfigLoaderConfig,
		retry.DefaultRetryConfig(),
	)
)

// Convenience functions for common resilient operations

// ExecuteFileSystemOperation runs a file system operation with full resilience
func ExecuteFileSystemOperation(ctx context.Context, operation func(context.Context) error) error {
	return FileSystemExecutor.Execute(ctx, operation)
}

// ExecuteNetworkOperation runs a network operation with full resilience
func ExecuteNetworkOperation(ctx context.Context, operation func(context.Context) error) error {
	return NetworkExecutor.Execute(ctx, operation)
}

// ExecuteConfigOperation runs a configuration operation with full resilience
func ExecuteConfigOperation(ctx context.Context, operation func(context.Context) error) error {
	return ConfigLoaderExecutor.Execute(ctx, operation)
}

// CreateCustomExecutor creates a resilient executor with custom configuration
func CreateCustomExecutor(cbName string, cbConfig resilience.CircuitBreakerConfig, retryConfig retry.RetryConfig) *ResilientExecutor {
	// Ensure circuit breaker name is set
	cbConfig.Name = cbName
	return NewResilientExecutor(cbConfig, retryConfig)
}

// WithRetryDisabled creates a resilient executor that only uses circuit breaker
func WithRetryDisabled(cbConfig resilience.CircuitBreakerConfig) *ResilientExecutor {
	noRetryConfig := retry.RetryConfig{
		MaxAttempts:       1, // No retry, just one attempt
		InitialDelay:      0,
		MaxDelay:          0,
		BackoffMultiplier: 1.0,
		JitterMaxPercent:  0,
		RetryableErrors:   []error{},
	}
	return NewResilientExecutor(cbConfig, noRetryConfig)
}

// WithCircuitBreakerDisabled creates a resilient executor that only uses retry
func WithCircuitBreakerDisabled(retryConfig retry.RetryConfig) *ResilientExecutor {
	// Create a circuit breaker that never opens (very high threshold)
	alwaysClosedConfig := resilience.CircuitBreakerConfig{
		Name:        "disabled",
		MaxFailures: 999999, // Effectively never opens
		Timeout:     0,
		MaxRequests: 999999,
	}
	return NewResilientExecutor(alwaysClosedConfig, retryConfig)
}

// ExecuteResilient is a convenience function for ad-hoc resilient operations
func ExecuteResilient(ctx context.Context, operation func(context.Context) error, options ...func(*ResilientExecutor)) error {
	// Use default configuration
	executor := NewResilientExecutor(
		resilience.DefaultCircuitBreakerConfig("ad_hoc"),
		retry.DefaultRetryConfig(),
	)

	// Apply any customization options
	for _, option := range options {
		option(executor)
	}

	return executor.Execute(ctx, operation)
}

// ConfigureRetry is an option function to customize retry behavior
func ConfigureRetry(config retry.RetryConfig) func(*ResilientExecutor) {
	return func(re *ResilientExecutor) {
		re.retrier = retry.NewRetrier(config)
	}
}

// ConfigureCircuitBreaker is an option function to customize circuit breaker behavior
func ConfigureCircuitBreaker(config resilience.CircuitBreakerConfig) func(*ResilientExecutor) {
	return func(re *ResilientExecutor) {
		re.circuitBreaker = resilience.NewCircuitBreaker(config)
	}
}

// Example usage patterns:
//
// Basic resilient execution:
//   err := integration.ExecuteFileSystemOperation(ctx, func(ctx context.Context) error {
//       return os.WriteFile("test.txt", data, 0644)
//   })
//
// Custom resilient execution:
//   executor := integration.CreateCustomExecutor("my_service", myCircuitConfig, myRetryConfig)
//   err := executor.Execute(ctx, myOperation)
//
// Ad-hoc resilient execution with options:
//   err := integration.ExecuteResilient(ctx, myOperation,
//       integration.ConfigureRetry(customRetryConfig),
//       integration.ConfigureCircuitBreaker(customCBConfig),
//   )
