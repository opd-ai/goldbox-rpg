package server

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"goldbox-rpg/pkg/resilience"
)

func TestNewCircuitBreaker(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	cb := NewCircuitBreaker(config)

	if cb.GetStats()["name"] != "test" {
		t.Errorf("Expected name 'test', got %s", cb.GetStats()["name"])
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected initial state Closed, got %s", cb.GetState())
	}
}

func TestCircuitBreakerStates(t *testing.T) {
	tests := []struct {
		state    CircuitBreakerState
		expected string
	}{
		{StateClosed, "Closed"},
		{StateOpen, "Open"},
		{StateHalfOpen, "HalfOpen"},
	}

	for _, test := range tests {
		if test.state.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.state.String())
		}
	}
}

func TestCircuitBreakerIntegration(t *testing.T) {
	// Test that circuit breaker integration works through server package
	manager := GetCircuitBreakerManager()
	if manager == nil {
		t.Error("Expected non-nil circuit breaker manager")
	}

	// Test server circuit breaker execution
	ctx := context.Background()
	err := ExecuteWithServerCircuitBreaker(ctx, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test with error
	expectedErr := errors.New("test error")
	err = ExecuteWithServerCircuitBreaker(ctx, func(ctx context.Context) error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Expected test error, got %v", err)
	}
}

func TestCircuitBreakerManagerIntegration(t *testing.T) {
	manager := GetCircuitBreakerManager()

	// Create a circuit breaker through the manager
	config := resilience.CircuitBreakerConfig{
		Name:        "test_integration",
		MaxFailures: 2,
		Timeout:     100 * time.Millisecond,
		MaxRequests: 1,
	}

	cb := manager.GetOrCreate("test_integration", &config)
	if cb == nil {
		t.Error("Expected non-nil circuit breaker")
	}

	ctx := context.Background()

	// Test failure threshold
	for i := 0; i < config.MaxFailures; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return errors.New("test error")
		})
		if err == nil {
			t.Error("Expected error")
		}
	}

	if cb.GetState() != StateOpen {
		t.Errorf("Expected state Open after failures, got %s", cb.GetState())
	}

	// Test circuit open behavior
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	if !errors.Is(err, ErrCircuitBreakerOpen) {
		t.Errorf("Expected circuit breaker open error, got %v", err)
	}
}

func TestCircuitBreakerStats(t *testing.T) {
	config := DefaultCircuitBreakerConfig("stats_test")
	cb := NewCircuitBreaker(config)

	stats := cb.GetStats()
	expectedFields := []string{"name", "state", "failures", "max_failures", "requests", "max_requests", "last_failure", "timeout"}

	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected stats field %s", field)
		}
	}

	if stats["name"] != "stats_test" {
		t.Errorf("Expected name 'stats_test', got %v", stats["name"])
	}

	if stats["state"] != "Closed" {
		t.Errorf("Expected state 'Closed', got %v", stats["state"])
	}
}

func TestCircuitBreakerConcurrentAccess(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:        "concurrent_test",
		MaxFailures: 10,
		Timeout:     100 * time.Millisecond,
		MaxRequests: 5,
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	// Launch concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Mix of success and failure operations
			var err error
			if id%3 == 0 {
				err = cb.Execute(ctx, func(ctx context.Context) error {
					return errors.New("test error")
				})
			} else {
				err = cb.Execute(ctx, func(ctx context.Context) error {
					time.Sleep(1 * time.Millisecond) // Small delay
					return nil
				})
			}

			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	// Count results
	total := 0
	for range results {
		total++
	}

	if total != numGoroutines {
		t.Errorf("Expected %d operations, got %d", numGoroutines, total)
	}
}
