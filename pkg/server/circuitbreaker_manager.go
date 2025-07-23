// Package server integrates circuit breaker manager functionality from the
// resilience package for coordinating multiple circuit breakers.
package server

import (
	"context"

	"goldbox-rpg/pkg/resilience"
)

// Re-export circuit breaker manager type
type CircuitBreakerManager = resilience.CircuitBreakerManager

// Re-export constructor functions
var NewCircuitBreakerManager = resilience.NewCircuitBreakerManager

// Re-export predefined configurations
var (
	FileSystemConfig   = resilience.FileSystemConfig
	WebSocketConfig    = resilience.WebSocketConfig
	ConfigLoaderConfig = resilience.ConfigLoaderConfig
)

// Helper functions for server-specific circuit breaker operations

// ExecuteWithFileSystemCircuitBreaker executes a function with file system circuit breaker protection
func ExecuteWithFileSystemCircuitBreaker(ctx context.Context, fn func(context.Context) error) error {
	return resilience.ExecuteWithFileSystemCircuitBreaker(ctx, fn)
}

// ExecuteWithConfigLoaderCircuitBreaker executes a function with config loader circuit breaker protection
func ExecuteWithConfigLoaderCircuitBreaker(ctx context.Context, fn func(context.Context) error) error {
	return resilience.ExecuteWithConfigLoaderCircuitBreaker(ctx, fn)
}
