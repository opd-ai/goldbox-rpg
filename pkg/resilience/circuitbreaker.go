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
	StateClosed:  "Closed",
	StateOpen:    "Open",
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
	logrus.WithFields(logrus.Fields{
		"function": "DefaultCircuitBreakerConfig",
		"package":  "resilience",
		"name":     name,
	}).Debug("entering DefaultCircuitBreakerConfig")

	config := CircuitBreakerConfig{
		Name:        name,
		MaxFailures: 5,
		Timeout:     30 * time.Second,
		MaxRequests: 3,
	}

	logrus.WithFields(logrus.Fields{
		"function":     "DefaultCircuitBreakerConfig",
		"package":      "resilience",
		"name":         name,
		"max_failures": config.MaxFailures,
		"timeout":      config.Timeout,
		"max_requests": config.MaxRequests,
	}).Debug("exiting DefaultCircuitBreakerConfig")

	return config
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
	logrus.WithFields(logrus.Fields{
		"function": "NewCircuitBreaker",
		"package":  "resilience",
		"name":     config.Name,
	}).Debug("entering NewCircuitBreaker")

	cb := &CircuitBreaker{
		config: config,
		state:  StateClosed,
		logger: logrus.WithField("circuit_breaker", config.Name),
	}

	logrus.WithFields(logrus.Fields{
		"function":      "NewCircuitBreaker",
		"package":       "resilience",
		"name":          config.Name,
		"initial_state": cb.state.String(),
	}).Info("circuit breaker created successfully")

	logrus.WithFields(logrus.Fields{
		"function": "NewCircuitBreaker",
		"package":  "resilience",
		"name":     config.Name,
	}).Debug("exiting NewCircuitBreaker")

	return cb
}

// ErrCircuitBreakerOpen is returned when the circuit breaker is open
var ErrCircuitBreakerOpen = errors.New("circuit breaker is open")

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	logrus.WithFields(logrus.Fields{
		"function": "Execute",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("entering Execute")

	// Check if we can execute the request
	if !cb.canExecute() {
		logrus.WithFields(logrus.Fields{
			"function": "Execute",
			"package":  "resilience",
			"name":     cb.config.Name,
			"state":    cb.state.String(),
			"reason":   "circuit breaker open",
		}).Warn("circuit breaker prevented execution")

		logrus.WithFields(logrus.Fields{
			"function": "Execute",
			"package":  "resilience",
			"name":     cb.config.Name,
		}).Debug("exiting Execute - blocked by circuit breaker")

		return fmt.Errorf("%w: %s", ErrCircuitBreakerOpen, cb.config.Name)
	}

	logrus.WithFields(logrus.Fields{
		"function": "Execute",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("circuit breaker allowing execution")

	// Track the request
	cb.beforeRequest()

	// Execute the function with context cancellation support
	done := make(chan error, 1)
	go func() {
		logrus.WithFields(logrus.Fields{
			"function": "Execute",
			"package":  "resilience",
			"name":     cb.config.Name,
		}).Debug("starting goroutine for protected function execution")

		defer func() {
			if r := recover(); r != nil {
				logrus.WithFields(logrus.Fields{
					"function": "Execute",
					"package":  "resilience",
					"name":     cb.config.Name,
					"panic":    r,
				}).Error("circuit breaker function panicked")
				done <- fmt.Errorf("function panicked: %v", r)
			}
		}()
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		logrus.WithFields(logrus.Fields{
			"function": "Execute",
			"package":  "resilience",
			"name":     cb.config.Name,
			"success":  err == nil,
		}).Debug("function execution completed")

		cb.afterRequest(err)

		logrus.WithFields(logrus.Fields{
			"function": "Execute",
			"package":  "resilience",
			"name":     cb.config.Name,
		}).Debug("exiting Execute - function completed")

		return err
	case <-ctx.Done():
		logrus.WithFields(logrus.Fields{
			"function": "Execute",
			"package":  "resilience",
			"name":     cb.config.Name,
			"error":    ctx.Err().Error(),
		}).Debug("function execution cancelled by context")

		cb.afterRequest(ctx.Err())

		logrus.WithFields(logrus.Fields{
			"function": "Execute",
			"package":  "resilience",
			"name":     cb.config.Name,
		}).Debug("exiting Execute - context cancelled")

		return ctx.Err()
	}
}

// canExecute determines if a request can be executed based on current state
func (cb *CircuitBreaker) canExecute() bool {
	logrus.WithFields(logrus.Fields{
		"function": "canExecute",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("entering canExecute")

	cb.mu.RLock()
	defer cb.mu.RUnlock()

	var result bool
	currentState := cb.state

	switch cb.state {
	case StateClosed:
		result = true
		logrus.WithFields(logrus.Fields{
			"function": "canExecute",
			"package":  "resilience",
			"name":     cb.config.Name,
			"state":    "closed",
		}).Debug("circuit breaker closed - allowing execution")
	case StateOpen:
		// Check if timeout has passed to transition to half-open
		if time.Since(cb.lastFailure) > cb.config.Timeout {
			result = true // Will transition to half-open in beforeRequest
			logrus.WithFields(logrus.Fields{
				"function":       "canExecute",
				"package":        "resilience",
				"name":           cb.config.Name,
				"state":          "open",
				"timeout_passed": true,
			}).Debug("circuit breaker open but timeout passed - allowing execution")
		} else {
			result = false
			logrus.WithFields(logrus.Fields{
				"function":           "canExecute",
				"package":            "resilience",
				"name":               cb.config.Name,
				"state":              "open",
				"time_since_failure": time.Since(cb.lastFailure),
				"timeout":            cb.config.Timeout,
			}).Debug("circuit breaker open and timeout not passed - blocking execution")
		}
	case StateHalfOpen:
		result = cb.requests < cb.config.MaxRequests
		logrus.WithFields(logrus.Fields{
			"function":     "canExecute",
			"package":      "resilience",
			"name":         cb.config.Name,
			"state":        "half-open",
			"requests":     cb.requests,
			"max_requests": cb.config.MaxRequests,
			"can_execute":  result,
		}).Debug("circuit breaker half-open - checking request limit")
	default:
		result = false
		logrus.WithFields(logrus.Fields{
			"function": "canExecute",
			"package":  "resilience",
			"name":     cb.config.Name,
			"state":    "unknown",
		}).Warn("circuit breaker in unknown state - blocking execution")
	}

	logrus.WithFields(logrus.Fields{
		"function":    "canExecute",
		"package":     "resilience",
		"name":        cb.config.Name,
		"state":       currentState.String(),
		"can_execute": result,
	}).Debug("exiting canExecute")

	return result
}

// beforeRequest is called before executing a request
func (cb *CircuitBreaker) beforeRequest() {
	logrus.WithFields(logrus.Fields{
		"function": "beforeRequest",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("entering beforeRequest")

	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state

	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.config.Timeout {
		logrus.WithFields(logrus.Fields{
			"function":           "beforeRequest",
			"package":            "resilience",
			"name":               cb.config.Name,
			"old_state":          oldState.String(),
			"new_state":          "HalfOpen",
			"time_since_failure": time.Since(cb.lastFailure),
			"timeout":            cb.config.Timeout,
		}).Info("circuit breaker transitioning to half-open state")
		cb.state = StateHalfOpen
		cb.requests = 0
	}

	if cb.state == StateHalfOpen {
		cb.requests++
		logrus.WithFields(logrus.Fields{
			"function":     "beforeRequest",
			"package":      "resilience",
			"name":         cb.config.Name,
			"state":        "half-open",
			"requests":     cb.requests,
			"max_requests": cb.config.MaxRequests,
		}).Debug("incremented half-open request counter")
	}

	logrus.WithFields(logrus.Fields{
		"function": "beforeRequest",
		"package":  "resilience",
		"name":     cb.config.Name,
		"state":    cb.state.String(),
	}).Debug("exiting beforeRequest")
}

// afterRequest is called after a request completes
func (cb *CircuitBreaker) afterRequest(err error) {
	logrus.WithFields(logrus.Fields{
		"function":  "afterRequest",
		"package":   "resilience",
		"name":      cb.config.Name,
		"has_error": err != nil,
	}).Debug("entering afterRequest")

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "afterRequest",
			"package":  "resilience",
			"name":     cb.config.Name,
			"error":    err.Error(),
		}).Debug("request failed - calling onFailure")
		cb.onFailure()
	} else {
		logrus.WithFields(logrus.Fields{
			"function": "afterRequest",
			"package":  "resilience",
			"name":     cb.config.Name,
		}).Debug("request succeeded - calling onSuccess")
		cb.onSuccess()
	}

	logrus.WithFields(logrus.Fields{
		"function": "afterRequest",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("exiting afterRequest")
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	logrus.WithFields(logrus.Fields{
		"function": "onFailure",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("entering onFailure")

	oldState := cb.state
	oldFailures := cb.failures

	cb.failures++
	cb.lastFailure = time.Now()

	logrus.WithFields(logrus.Fields{
		"function":     "onFailure",
		"package":      "resilience",
		"name":         cb.config.Name,
		"old_failures": oldFailures,
		"new_failures": cb.failures,
		"last_failure": cb.lastFailure,
	}).Debug("updated failure count and timestamp")

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			logrus.WithFields(logrus.Fields{
				"function":     "onFailure",
				"package":      "resilience",
				"name":         cb.config.Name,
				"failures":     cb.failures,
				"max_failures": cb.config.MaxFailures,
				"old_state":    oldState.String(),
				"new_state":    "Open",
			}).Warn("circuit breaker opening due to excessive failures")
			cb.state = StateOpen
		} else {
			logrus.WithFields(logrus.Fields{
				"function":        "onFailure",
				"package":         "resilience",
				"name":            cb.config.Name,
				"failures":        cb.failures,
				"max_failures":    cb.config.MaxFailures,
				"remaining_tries": cb.config.MaxFailures - cb.failures,
			}).Debug("failure recorded but staying closed")
		}
	case StateHalfOpen:
		logrus.WithFields(logrus.Fields{
			"function":  "onFailure",
			"package":   "resilience",
			"name":      cb.config.Name,
			"old_state": oldState.String(),
			"new_state": "Open",
		}).Info("circuit breaker returning to open state after half-open failure")
		cb.state = StateOpen
		cb.requests = 0
	}

	logrus.WithFields(logrus.Fields{
		"function": "onFailure",
		"package":  "resilience",
		"name":     cb.config.Name,
		"state":    cb.state.String(),
	}).Debug("exiting onFailure")
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	logrus.WithFields(logrus.Fields{
		"function": "onSuccess",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("entering onSuccess")

	oldState := cb.state
	oldFailures := cb.failures

	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		if cb.failures > 0 {
			logrus.WithFields(logrus.Fields{
				"function":     "onSuccess",
				"package":      "resilience",
				"name":         cb.config.Name,
				"old_failures": oldFailures,
				"new_failures": 0,
			}).Debug("resetting failure count after successful request")
		}
		cb.failures = 0
	case StateHalfOpen:
		logrus.WithFields(logrus.Fields{
			"function":     "onSuccess",
			"package":      "resilience",
			"name":         cb.config.Name,
			"requests":     cb.requests,
			"max_requests": cb.config.MaxRequests,
		}).Debug("successful request in half-open state")

		if cb.requests >= cb.config.MaxRequests {
			logrus.WithFields(logrus.Fields{
				"function":  "onSuccess",
				"package":   "resilience",
				"name":      cb.config.Name,
				"old_state": oldState.String(),
				"new_state": "Closed",
				"requests":  cb.requests,
			}).Info("circuit breaker closing after successful half-open test")
			cb.state = StateClosed
			cb.failures = 0
			cb.requests = 0
		}
	}

	logrus.WithFields(logrus.Fields{
		"function": "onSuccess",
		"package":  "resilience",
		"name":     cb.config.Name,
		"state":    cb.state.String(),
	}).Debug("exiting onSuccess")
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	logrus.WithFields(logrus.Fields{
		"function": "GetState",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("entering GetState")

	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state := cb.state

	logrus.WithFields(logrus.Fields{
		"function": "GetState",
		"package":  "resilience",
		"name":     cb.config.Name,
		"state":    state.String(),
	}).Debug("exiting GetState")

	return state
}

// GetStats returns current statistics for the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	logrus.WithFields(logrus.Fields{
		"function": "GetStats",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("entering GetStats")

	cb.mu.RLock()
	defer cb.mu.RUnlock()

	stats := map[string]interface{}{
		"name":         cb.config.Name,
		"state":        cb.state.String(),
		"failures":     cb.failures,
		"max_failures": cb.config.MaxFailures,
		"requests":     cb.requests,
		"max_requests": cb.config.MaxRequests,
		"last_failure": cb.lastFailure,
		"timeout":      cb.config.Timeout,
	}

	logrus.WithFields(logrus.Fields{
		"function": "GetStats",
		"package":  "resilience",
		"name":     cb.config.Name,
		"stats":    stats,
	}).Debug("exiting GetStats")

	return stats
}

// Reset forces the circuit breaker back to closed state
func (cb *CircuitBreaker) Reset() {
	logrus.WithFields(logrus.Fields{
		"function": "Reset",
		"package":  "resilience",
		"name":     cb.config.Name,
	}).Debug("entering Reset")

	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	oldFailures := cb.failures

	logrus.WithFields(logrus.Fields{
		"function":     "Reset",
		"package":      "resilience",
		"name":         cb.config.Name,
		"old_state":    oldState.String(),
		"old_failures": oldFailures,
	}).Info("circuit breaker manually reset")

	cb.state = StateClosed
	cb.failures = 0
	cb.requests = 0
	cb.lastFailure = time.Time{}

	logrus.WithFields(logrus.Fields{
		"function": "Reset",
		"package":  "resilience",
		"name":     cb.config.Name,
		"state":    cb.state.String(),
	}).Debug("exiting Reset")
}
