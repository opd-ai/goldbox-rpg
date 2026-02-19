// Package resilience provides circuit breaker patterns for external dependencies
// to prevent cascade failures and improve system resilience.
package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	// Configure structured logging with caller context
	logrus.SetReportCaller(true)
}

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

// circuitBreakerStateNames provides O(1) lookup for state string representation
var circuitBreakerStateNames = [...]string{
	StateClosed:   "Closed",
	StateOpen:     "Open",
	StateHalfOpen: "HalfOpen",
}

// String returns the string representation of the circuit breaker state.
// Uses bounds-checked array lookup for efficiency.
func (s CircuitBreakerState) String() string {
	if s >= 0 && int(s) < len(circuitBreakerStateNames) {
		return circuitBreakerStateNames[s]
	}
	return "Unknown"
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
	cb := &CircuitBreaker{
		config: config,
		state:  StateClosed,
		logger: logrus.WithField("circuit_breaker", config.Name),
	}

	logrus.WithFields(logrus.Fields{
		"function":      "NewCircuitBreaker",
		"name":          config.Name,
		"initial_state": cb.state.String(),
	}).Info("circuit breaker created successfully")

	return cb
}

// ErrCircuitBreakerOpen is returned when the circuit breaker is open
var ErrCircuitBreakerOpen = errors.New("circuit breaker is open")

// Execute runs the given function with circuit breaker protection.
// The function is executed synchronously in the calling goroutine for performance.
// Panics in the wrapped function are recovered and returned as errors.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check context before attempting execution
	if err := ctx.Err(); err != nil {
		cb.afterRequest(err)
		return err
	}

	// Check if we can execute the request
	if !cb.canExecute() {
		logrus.WithFields(logrus.Fields{
			"name":  cb.config.Name,
			"state": cb.state.String(),
		}).Warn("circuit breaker prevented execution")
		return fmt.Errorf("%w: %s", ErrCircuitBreakerOpen, cb.config.Name)
	}

	// Track the request
	cb.beforeRequest()

	// Execute synchronously with panic recovery
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithFields(logrus.Fields{
					"name":  cb.config.Name,
					"panic": r,
				}).Error("circuit breaker function panicked")
				err = fmt.Errorf("function panicked: %v", r)
			}
		}()
		err = fn(ctx)
	}()

	cb.afterRequest(err)
	return err
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
		return time.Since(cb.lastFailure) > cb.config.Timeout
	case StateHalfOpen:
		return cb.requests < cb.config.MaxRequests
	default:
		logrus.WithFields(logrus.Fields{
			"name":  cb.config.Name,
			"state": cb.state,
		}).Warn("circuit breaker in unknown state")
		return false
	}
}

// beforeRequest is called before executing a request
func (cb *CircuitBreaker) beforeRequest() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.config.Timeout {
		logrus.WithFields(logrus.Fields{
			"name":      cb.config.Name,
			"old_state": StateOpen.String(),
			"new_state": StateHalfOpen.String(),
		}).Info("circuit breaker transitioning to half-open state")
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

// onFailure handles a failed request (must be called with mutex held)
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			logrus.WithFields(logrus.Fields{
				"name":         cb.config.Name,
				"failures":     cb.failures,
				"max_failures": cb.config.MaxFailures,
			}).Warn("circuit breaker opening due to excessive failures")
			cb.state = StateOpen
		}
	case StateHalfOpen:
		logrus.WithFields(logrus.Fields{
			"name": cb.config.Name,
		}).Info("circuit breaker returning to open state after half-open failure")
		cb.state = StateOpen
		cb.requests = 0
	}
}

// onSuccess handles a successful request (must be called with mutex held)
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
	case StateHalfOpen:
		if cb.requests >= cb.config.MaxRequests {
			logrus.WithFields(logrus.Fields{
				"name":     cb.config.Name,
				"requests": cb.requests,
			}).Info("circuit breaker closing after successful half-open test")
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

	oldState := cb.state

	logrus.WithFields(logrus.Fields{
		"name":      cb.config.Name,
		"old_state": oldState.String(),
	}).Info("circuit breaker manually reset")

	cb.state = StateClosed
	cb.failures = 0
	cb.requests = 0
	cb.lastFailure = time.Time{}
}
