// Package integration combines circuit breaker and retry patterns for comprehensive
// fault tolerance in the GoldBox RPG Engine.
//
// This package provides ResilientExecutor which layers retry logic on top of
// circuit breaker protection, giving operations the benefits of both mechanisms:
// automatic retries for transient failures and fast-fail for persistent outages.
//
// # Execution Flow
//
// When executing an operation:
//
//  1. Circuit breaker checks if the operation should proceed
//  2. If circuit is open, fails immediately with ErrCircuitBreakerOpen
//  3. If circuit allows, operation executes with retry protection
//  4. Retry handles transient failures with exponential backoff
//  5. Circuit breaker records success/failure for state management
//
// # Creating Executors
//
// Create a custom executor with specific configuration:
//
//	cbConfig := resilience.CircuitBreakerConfig{
//	    MaxFailures: 5,
//	    Timeout:     30 * time.Second,
//	}
//	retryConfig := retry.RetryConfig{
//	    MaxAttempts:  3,
//	    InitialDelay: 100 * time.Millisecond,
//	}
//	executor := integration.NewResilientExecutor("my-service", cbConfig, retryConfig)
//
// # Executing Operations
//
// Wrap operations with combined protection:
//
//	err := executor.Execute(ctx, func() error {
//	    return callExternalAPI()
//	})
//
// # Pre-configured Executors
//
// Global executors for common operation types:
//
//	// File system operations
//	err := integration.ExecuteFileSystemOperation(ctx, operation)
//
//	// Network/WebSocket operations
//	err := integration.ExecuteNetworkOperation(ctx, operation)
//
//	// Configuration loading
//	err := integration.ExecuteConfigOperation(ctx, operation)
//
// # Ad-hoc Execution
//
// For one-off operations with custom options:
//
//	err := integration.ExecuteResilient(ctx, "operation-name", operation,
//	    integration.ConfigureRetry(retryConfig),
//	    integration.ConfigureCircuitBreaker(cbConfig),
//	)
//
// # Disabling Mechanisms
//
// Run with only one protection mechanism:
//
//	// Retry only, no circuit breaker
//	err := integration.WithRetryDisabled(executor).Execute(ctx, operation)
//
//	// Circuit breaker only, no retry
//	err := integration.WithCircuitBreakerDisabled(executor).Execute(ctx, operation)
//
// # Statistics
//
// Query combined statistics from both mechanisms:
//
//	stats := executor.GetStats()
//	// Contains circuit breaker state and retry metrics
//
// # Testing
//
// Reset global executors between tests:
//
//	integration.ResetExecutorsForTesting()
package integration
