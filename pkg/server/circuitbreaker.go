// Package server provides circuit breaker patterns for external dependencies
// to prevent cascade failures and improve system resilience.
package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CircuitBreakerState represents the current state of a circuit breaker
type CircuitBreakerState int

const (
	// StateClosed - circuit breaker is closed, allowing requests through
	StateClosed CircuitBreakerState = iota
	// StateOpen - circuit breaker is open, failing fast
	StateOpen
	// StateHalfOpen - circuit breaker is testing if service has recovered
	StateHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "Closed"
	case StateOpen:
		return "Open"
	case StateHalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// Name is the identifier for this circuit breaker
	Name string

	// MaxFailures is the number of failures before opening the circuit
	MaxFailures int

	// Timeout is how long to wait before transitioning from Open to HalfOpen
	Timeout time.Duration

	// MaxRequests is the maximum number of requests allowed in HalfOpen state
	MaxRequests int
}

// DefaultCircuitBreakerConfig returns a sensible default configuration
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:        name,
		MaxFailures: 5,
		Timeout:     30 * time.Second,
		MaxRequests: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern for protecting external dependencies
type CircuitBreaker struct {
	config      CircuitBreakerConfig
	mu          sync.RWMutex
	state       CircuitBreakerState
	failures    int
	requests    int
	lastFailure time.Time
	logger      *logrus.Entry
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		logger: logrus.WithField("circuit_breaker", config.Name),
	}
}

// ErrCircuitBreakerOpen is returned when the circuit breaker is open
var ErrCircuitBreakerOpen = errors.New("circuit breaker is open")

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check if we can execute the request
	if !cb.canExecute() {
		cb.logger.Debug("Circuit breaker prevented execution")
		return fmt.Errorf("%w: %s", ErrCircuitBreakerOpen, cb.config.Name)
	}

	// Track the request
	cb.beforeRequest()

	// Execute the function with context cancellation support
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				cb.logger.WithField("panic", r).Error("Circuit breaker function panicked")
				done <- fmt.Errorf("function panicked: %v", r)
			}
		}()
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		cb.afterRequest(err)
		return err
	case <-ctx.Done():
		cb.afterRequest(ctx.Err())
		return ctx.Err()
	}
}

// canExecute determines if a request can be executed based on current state
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed to transition to half-open
		if time.Since(cb.lastFailure) > cb.config.Timeout {
			return true // Will transition to half-open in beforeRequest
		}
		return false
	case StateHalfOpen:
		return cb.requests < cb.config.MaxRequests
	default:
		return false
	}
}

// beforeRequest is called before executing a request
func (cb *CircuitBreaker) beforeRequest() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.config.Timeout {
		cb.logger.Info("Circuit breaker transitioning to half-open state")
		cb.state = StateHalfOpen
		cb.requests = 0
	}

	if cb.state == StateHalfOpen {
		cb.requests++
	}
}

// afterRequest is called after a request completes
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.logger.WithFields(logrus.Fields{
				"failures": cb.failures,
				"max":      cb.config.MaxFailures,
			}).Warn("Circuit breaker opening due to failures")
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.logger.Info("Circuit breaker returning to open state after half-open failure")
		cb.state = StateOpen
		cb.requests = 0
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
	case StateHalfOpen:
		if cb.requests >= cb.config.MaxRequests {
			cb.logger.Info("Circuit breaker closing after successful half-open test")
			cb.state = StateClosed
			cb.failures = 0
			cb.requests = 0
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns current statistics for the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":         cb.config.Name,
		"state":        cb.state.String(),
		"failures":     cb.failures,
		"max_failures": cb.config.MaxFailures,
		"requests":     cb.requests,
		"max_requests": cb.config.MaxRequests,
		"last_failure": cb.lastFailure,
		"timeout":      cb.config.Timeout,
	}
}

// Reset forces the circuit breaker back to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.logger.Info("Circuit breaker manually reset")
	cb.state = StateClosed
	cb.failures = 0
	cb.requests = 0
	cb.lastFailure = time.Time{}
}
