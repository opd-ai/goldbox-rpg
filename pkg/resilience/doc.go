// Package resilience implements the circuit breaker pattern for fault tolerance.
//
// This package protects external dependencies and prevents cascade failures by
// enabling fast-fail behavior when services become unavailable, with automatic
// recovery testing when conditions improve.
//
// # Circuit Breaker Pattern
//
// A circuit breaker operates in three states:
//
//   - Closed: Normal operation, all requests pass through
//   - Open: Service failing, requests fail immediately (fast-fail)
//   - HalfOpen: Testing recovery with limited requests
//
// State transitions:
//
//	Closed → Open: After MaxFailures consecutive failures
//	Open → HalfOpen: After Timeout period expires
//	HalfOpen → Closed: After successful test requests
//	HalfOpen → Open: If test requests fail
//
// # Creating Circuit Breakers
//
// Create a circuit breaker with custom configuration:
//
//	config := resilience.CircuitBreakerConfig{
//	    MaxFailures: 5,           // Open after 5 failures
//	    Timeout:     30*time.Second, // Wait 30s before testing
//	    MaxRequests: 3,           // Allow 3 test requests in half-open
//	}
//	cb := resilience.NewCircuitBreaker("external-api", config)
//
// # Executing Protected Operations
//
// Wrap operations with circuit breaker protection:
//
//	err := cb.Execute(ctx, func() error {
//	    return callExternalService()
//	})
//	if errors.Is(err, resilience.ErrCircuitBreakerOpen) {
//	    // Service is down, handle gracefully
//	}
//
// # Managing Multiple Breakers
//
// Use CircuitBreakerManager for multiple dependencies:
//
//	manager := resilience.NewCircuitBreakerManager()
//	cb := manager.GetOrCreate("database", config)
//	stats := manager.GetAllStats()
//
// # Pre-configured Breakers
//
// Global convenience functions with sensible defaults:
//
//	// File system operations (3 failures, 5s timeout)
//	err := resilience.ExecuteWithFileSystemBreaker(ctx, func() error { ... })
//
//	// WebSocket operations (5 failures, 10s timeout)
//	err := resilience.ExecuteWithWebSocketBreaker(ctx, func() error { ... })
//
//	// Config loading (3 failures, 30s timeout)
//	err := resilience.ExecuteWithConfigLoaderBreaker(ctx, func() error { ... })
//
// # Monitoring
//
// Query circuit breaker state and statistics:
//
//	state := cb.GetState()       // StateClosed, StateOpen, or StateHalfOpen
//	stats := cb.GetStats()       // Failure counts, request counts, timestamps
//
// # Thread Safety
//
// All circuit breaker operations are thread-safe via internal mutex protection.
// Multiple goroutines can safely execute through the same breaker.
package resilience
