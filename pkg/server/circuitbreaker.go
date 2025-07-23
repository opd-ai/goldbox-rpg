// Package server integrates circuit breaker patterns into the server for protecting
// external dependencies and preventing cascade failures.
package server

import (
	"context"

	"goldbox-rpg/pkg/resilience"
)

// Re-export circuit breaker types for server package compatibility
type (
	CircuitBreakerState  = resilience.CircuitBreakerState
	CircuitBreakerConfig = resilience.CircuitBreakerConfig
	CircuitBreaker       = resilience.CircuitBreaker
)

// Re-export circuit breaker states
const (
	StateClosed   = resilience.StateClosed
	StateOpen     = resilience.StateOpen
	StateHalfOpen = resilience.StateHalfOpen
)

// Re-export circuit breaker errors
var ErrCircuitBreakerOpen = resilience.ErrCircuitBreakerOpen

// Re-export constructor functions
var (
	NewCircuitBreaker           = resilience.NewCircuitBreaker
	DefaultCircuitBreakerConfig = resilience.DefaultCircuitBreakerConfig
)

// Helper functions for server-specific circuit breaker operations

// GetCircuitBreakerManager returns the global circuit breaker manager instance
func GetCircuitBreakerManager() *resilience.CircuitBreakerManager {
	return resilience.GetGlobalCircuitBreakerManager()
}

// ExecuteWithServerCircuitBreaker executes a function with WebSocket circuit breaker protection
func ExecuteWithServerCircuitBreaker(ctx context.Context, fn func(context.Context) error) error {
	return resilience.ExecuteWithWebSocketCircuitBreaker(ctx, fn)
}
