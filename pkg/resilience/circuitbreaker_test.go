package resilience

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
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
		{CircuitBreakerState(99), "Unknown"},
	}

	for _, test := range tests {
		if test.state.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.state.String())
		}
	}
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")

	if config.Name != "test" {
		t.Errorf("Expected name 'test', got %s", config.Name)
	}

	if config.MaxFailures != 5 {
		t.Errorf("Expected MaxFailures 5, got %d", config.MaxFailures)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", config.Timeout)
	}

	if config.MaxRequests != 3 {
		t.Errorf("Expected MaxRequests 3, got %d", config.MaxRequests)
	}
}

func TestCircuitBreakerExecuteSuccess(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	successFunc := func(ctx context.Context) error {
		return nil
	}

	err := cb.Execute(ctx, successFunc)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state Closed after success, got %s", cb.GetState())
	}
}

func TestCircuitBreakerExecuteFailure(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:        "test",
		MaxFailures: 2,
		Timeout:     100 * time.Millisecond,
		MaxRequests: 1,
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	expectedErr := errors.New("test error")
	failureFunc := func(ctx context.Context) error {
		return expectedErr
	}

	// Execute failures up to threshold
	for i := 0; i < config.MaxFailures; i++ {
		err := cb.Execute(ctx, failureFunc)
		if err != expectedErr {
			t.Errorf("Expected test error, got %v", err)
		}

		if i < config.MaxFailures-1 && cb.GetState() != StateClosed {
			t.Errorf("Expected state Closed before threshold, got %s", cb.GetState())
		}
	}

	if cb.GetState() != StateOpen {
		t.Errorf("Expected state Open after threshold, got %s", cb.GetState())
	}

	// Next execution should fail fast
	err := cb.Execute(ctx, failureFunc)
	if !errors.Is(err, ErrCircuitBreakerOpen) {
		t.Errorf("Expected circuit breaker open error, got %v", err)
	}
}

func TestCircuitBreakerHalfOpenTransition(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:        "test",
		MaxFailures: 1,
		Timeout:     50 * time.Millisecond,
		MaxRequests: 2,
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	// Trigger failure to open circuit
	failureFunc := func(ctx context.Context) error {
		return errors.New("test error")
	}

	err := cb.Execute(ctx, failureFunc)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if cb.GetState() != StateOpen {
		t.Errorf("Expected state Open, got %s", cb.GetState())
	}

	// Wait for timeout
	time.Sleep(config.Timeout + 10*time.Millisecond)

	// Next execution should transition to half-open
	successFunc := func(ctx context.Context) error {
		return nil
	}

	err = cb.Execute(ctx, successFunc)
	if err != nil {
		t.Errorf("Expected no error in half-open, got %v", err)
	}

	// Should be half-open after first request
	stats := cb.GetStats()
	if stats["state"] != "HalfOpen" {
		t.Errorf("Expected state HalfOpen, got %s", stats["state"])
	}
}

func TestCircuitBreakerManagerIntegration(t *testing.T) {
	cbm := NewCircuitBreakerManager()

	// Test creating a circuit breaker with default config
	cb1 := cbm.GetOrCreate("test1", nil)
	if cb1 == nil {
		t.Error("Expected non-nil circuit breaker")
	}

	stats := cb1.GetStats()
	if stats["name"] != "test1" {
		t.Errorf("Expected name 'test1', got %s", stats["name"])
	}

	// Test getting an existing circuit breaker
	cb2 := cbm.GetOrCreate("test1", nil)
	if cb1 != cb2 {
		t.Error("Expected same circuit breaker instance")
	}

	// Test creating a new circuit breaker with custom config
	customConfig := CircuitBreakerConfig{
		Name:        "custom",
		MaxFailures: 10,
		Timeout:     time.Minute,
		MaxRequests: 5,
	}

	cb3 := cbm.GetOrCreate("test2", &customConfig)
	stats3 := cb3.GetStats()
	if stats3["max_failures"] != 10 {
		t.Errorf("Expected MaxFailures 10, got %v", stats3["max_failures"])
	}

	if stats3["name"] != "test2" {
		t.Errorf("Expected name 'test2', got %s", stats3["name"])
	}
}

func TestGlobalCircuitBreakerManager(t *testing.T) {
	// Test that global instance is available
	cbm := GetGlobalCircuitBreakerManager()
	if cbm == nil {
		t.Error("Expected non-nil global circuit breaker manager")
	}

	// Test that it's the same instance
	cbm2 := GetGlobalCircuitBreakerManager()
	if cbm != cbm2 {
		t.Error("Expected same global instance")
	}
}

func TestHelperFunctions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func(context.Context, func(context.Context) error) error
	}{
		{"ExecuteWithFileSystemCircuitBreaker", ExecuteWithFileSystemCircuitBreaker},
		{"ExecuteWithWebSocketCircuitBreaker", ExecuteWithWebSocketCircuitBreaker},
		{"ExecuteWithConfigLoaderCircuitBreaker", ExecuteWithConfigLoaderCircuitBreaker},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test successful execution
			err := test.fn(ctx, func(ctx context.Context) error {
				return nil
			})
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Test error execution
			expectedErr := errors.New("test error")
			err = test.fn(ctx, func(ctx context.Context) error {
				return expectedErr
			})

			if err != expectedErr {
				t.Errorf("Expected test error, got %v", err)
			}
		})
	}
}

func TestCircuitBreakerConcurrentAccess(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:        "test",
		MaxFailures: 10,
		Timeout:     100 * time.Millisecond,
		MaxRequests: 5,
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 20
	successCount := make(chan int, numGoroutines)

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

			if err == nil {
				successCount <- 1
			} else {
				successCount <- 0
			}
		}(i)
	}

	wg.Wait()
	close(successCount)

	// Count successes
	total := 0
	successes := 0
	for count := range successCount {
		total++
		successes += count
	}

	if total != numGoroutines {
		t.Errorf("Expected %d operations, got %d", numGoroutines, total)
	}

	// Should have some successes (exact count depends on timing and circuit state)
	if successes == 0 {
		t.Error("Expected some successful operations")
	}

	t.Logf("Concurrent test completed: %d/%d successful operations", successes, total)
}
