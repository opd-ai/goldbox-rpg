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

func TestCircuitBreakerManagerRemove(t *testing.T) {
	cbm := NewCircuitBreakerManager()

	// Create circuit breakers
	cbm.GetOrCreate("cb1", nil)
	cbm.GetOrCreate("cb2", nil)
	cbm.GetOrCreate("cb3", nil)

	// Verify they exist
	if _, exists := cbm.Get("cb1"); !exists {
		t.Error("Expected cb1 to exist")
	}
	if _, exists := cbm.Get("cb2"); !exists {
		t.Error("Expected cb2 to exist")
	}

	// Remove one
	cbm.Remove("cb2")

	// Verify cb2 is removed but others remain
	if _, exists := cbm.Get("cb1"); !exists {
		t.Error("Expected cb1 to still exist after removing cb2")
	}
	if _, exists := cbm.Get("cb2"); exists {
		t.Error("Expected cb2 to be removed")
	}
	if _, exists := cbm.Get("cb3"); !exists {
		t.Error("Expected cb3 to still exist after removing cb2")
	}

	// Remove non-existent breaker (should not panic)
	cbm.Remove("nonexistent")

	// Verify state unchanged
	if _, exists := cbm.Get("cb1"); !exists {
		t.Error("Expected cb1 to remain after removing non-existent")
	}
}

func TestCircuitBreakerManagerGetBreakerNames(t *testing.T) {
	cbm := NewCircuitBreakerManager()

	// Test empty manager
	names := cbm.GetBreakerNames()
	if len(names) != 0 {
		t.Errorf("Expected empty names slice, got %v", names)
	}

	// Add some breakers
	cbm.GetOrCreate("alpha", nil)
	cbm.GetOrCreate("beta", nil)
	cbm.GetOrCreate("gamma", nil)

	names = cbm.GetBreakerNames()
	if len(names) != 3 {
		t.Errorf("Expected 3 breaker names, got %d", len(names))
	}

	// Check all names are present (order not guaranteed)
	expected := map[string]bool{"alpha": false, "beta": false, "gamma": false}
	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		} else {
			t.Errorf("Unexpected breaker name: %s", name)
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("Expected breaker name '%s' not found", name)
		}
	}

	// Remove one and verify
	cbm.Remove("beta")
	names = cbm.GetBreakerNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 breaker names after removal, got %d", len(names))
	}

	for _, name := range names {
		if name == "beta" {
			t.Error("Expected 'beta' to be removed from names")
		}
	}
}

func TestCircuitBreakerManagerResetAll(t *testing.T) {
	cbm := NewCircuitBreakerManager()
	ctx := context.Background()

	// Create breakers with low failure threshold
	config := CircuitBreakerConfig{
		Name:        "test",
		MaxFailures: 1,
		Timeout:     10 * time.Second, // Long timeout to ensure they stay open
		MaxRequests: 1,
	}

	cb1 := cbm.GetOrCreate("cb1", &config)
	cb2 := cbm.GetOrCreate("cb2", &config)

	// Trigger failures to open both breakers
	failFunc := func(ctx context.Context) error {
		return errors.New("test failure")
	}

	cb1.Execute(ctx, failFunc)
	cb2.Execute(ctx, failFunc)

	// Verify both are open
	if cb1.GetState() != StateOpen {
		t.Errorf("Expected cb1 to be open, got %s", cb1.GetState())
	}
	if cb2.GetState() != StateOpen {
		t.Errorf("Expected cb2 to be open, got %s", cb2.GetState())
	}

	// Reset all
	cbm.ResetAll()

	// Verify both are now closed
	if cb1.GetState() != StateClosed {
		t.Errorf("Expected cb1 to be closed after ResetAll, got %s", cb1.GetState())
	}
	if cb2.GetState() != StateClosed {
		t.Errorf("Expected cb2 to be closed after ResetAll, got %s", cb2.GetState())
	}

	// Verify they can execute again
	successFunc := func(ctx context.Context) error {
		return nil
	}
	if err := cb1.Execute(ctx, successFunc); err != nil {
		t.Errorf("Expected cb1 to execute after reset, got %v", err)
	}
	if err := cb2.Execute(ctx, successFunc); err != nil {
		t.Errorf("Expected cb2 to execute after reset, got %v", err)
	}
}

func TestCircuitBreakerManagerGetAllStats(t *testing.T) {
	cbm := NewCircuitBreakerManager()

	// Test empty manager
	stats := cbm.GetAllStats()
	if len(stats) != 0 {
		t.Errorf("Expected empty stats map, got %v", stats)
	}

	// Add breakers
	cbm.GetOrCreate("stats1", nil)
	cbm.GetOrCreate("stats2", nil)

	stats = cbm.GetAllStats()
	if len(stats) != 2 {
		t.Errorf("Expected 2 stats entries, got %d", len(stats))
	}

	// Verify stats contain expected keys
	for _, name := range []string{"stats1", "stats2"} {
		if _, ok := stats[name]; !ok {
			t.Errorf("Expected stats for '%s'", name)
		}
		statMap, ok := stats[name].(map[string]interface{})
		if !ok {
			t.Errorf("Expected stats[%s] to be map[string]interface{}", name)
			continue
		}
		if statMap["name"] != name {
			t.Errorf("Expected stats name '%s', got %v", name, statMap["name"])
		}
	}
}

func TestHelperFunctionsErrorPaths(t *testing.T) {
	// Test error propagation with fresh circuit breakers by using unique names
	// This avoids issues with global state from other tests

	t.Run("ErrorPropagation", func(t *testing.T) {
		ctx := context.Background()
		testErr := errors.New("specific test error")

		tests := []struct {
			name string
			fn   func(context.Context, func(context.Context) error) error
		}{
			{"FileSystem", ExecuteWithFileSystemCircuitBreaker},
			{"WebSocket", ExecuteWithWebSocketCircuitBreaker},
			{"ConfigLoader", ExecuteWithConfigLoaderCircuitBreaker},
		}

		// Reset global manager breakers before testing error propagation
		cbm := GetGlobalCircuitBreakerManager()
		cbm.ResetAll()

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				// Execute with error
				err := test.fn(ctx, func(ctx context.Context) error {
					return testErr
				})

				// Error should be propagated (unless circuit is already open)
				if err != testErr && !errors.Is(err, ErrCircuitBreakerOpen) {
					t.Errorf("Expected error %v or circuit breaker open, got %v", testErr, err)
				}
			})
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		// Reset breakers before context cancellation test
		cbm := GetGlobalCircuitBreakerManager()
		cbm.ResetAll()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Use a custom circuit breaker to avoid global state issues
		config := CircuitBreakerConfig{
			Name:        "context_test",
			MaxFailures: 10,
			Timeout:     30 * time.Second,
			MaxRequests: 5,
		}
		cb := NewCircuitBreaker(config)

		err := cb.Execute(ctx, func(ctx context.Context) error {
			// Check if context was cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(10 * time.Millisecond)
				return nil
			}
		})

		// Should return context error
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}

func TestCircuitBreakerManagerConcurrentAccess(t *testing.T) {
	cbm := NewCircuitBreakerManager()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 20

	// Concurrent operations on manager
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			name := "concurrent_cb"

			// Mix of manager operations
			switch id % 5 {
			case 0:
				cbm.GetOrCreate(name, nil)
			case 1:
				cbm.Get(name)
			case 2:
				cbm.GetBreakerNames()
			case 3:
				cbm.GetAllStats()
			case 4:
				cb, exists := cbm.Get(name)
				if exists {
					cb.Execute(ctx, func(ctx context.Context) error {
						return nil
					})
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify manager is still functional
	names := cbm.GetBreakerNames()
	if len(names) == 0 {
		t.Log("No breakers created during concurrent test (possible race)")
	}
}

func TestCircuitBreakerGetMethod(t *testing.T) {
	cbm := NewCircuitBreakerManager()

	// Test Get on non-existent breaker
	cb, exists := cbm.Get("nonexistent")
	if exists {
		t.Error("Expected non-existent breaker to not exist")
	}
	if cb != nil {
		t.Error("Expected nil circuit breaker for non-existent name")
	}

	// Create a breaker
	cbm.GetOrCreate("existing", nil)

	// Test Get on existing breaker
	cb, exists = cbm.Get("existing")
	if !exists {
		t.Error("Expected existing breaker to exist")
	}
	if cb == nil {
		t.Error("Expected non-nil circuit breaker for existing name")
	}

	// Verify it's the same instance
	cb2 := cbm.GetOrCreate("existing", nil)
	if cb != cb2 {
		t.Error("Expected same circuit breaker instance")
	}
}

func TestPredefinedConfigs(t *testing.T) {
	// Verify predefined configs have sensible values
	configs := []struct {
		name   string
		config CircuitBreakerConfig
	}{
		{"FileSystemConfig", FileSystemConfig},
		{"WebSocketConfig", WebSocketConfig},
		{"ConfigLoaderConfig", ConfigLoaderConfig},
	}

	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.config.Name == "" {
				t.Error("Expected non-empty name")
			}
			if tc.config.MaxFailures <= 0 {
				t.Error("Expected positive MaxFailures")
			}
			if tc.config.Timeout <= 0 {
				t.Error("Expected positive Timeout")
			}
			if tc.config.MaxRequests <= 0 {
				t.Error("Expected positive MaxRequests")
			}
		})
	}
}
