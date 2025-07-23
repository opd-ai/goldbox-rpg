package server

import (
	"context"
	"errors"
	"testing"

	"goldbox-rpg/pkg/resilience"
)

func TestNewCircuitBreakerManager(t *testing.T) {
	cbm := NewCircuitBreakerManager()

	if cbm == nil {
		t.Error("Expected non-nil circuit breaker manager")
	}

	stats := cbm.GetAllStats()
	if len(stats) != 0 {
		t.Errorf("Expected empty stats initially, got %d items", len(stats))
	}
}

func TestCircuitBreakerManagerServerIntegration(t *testing.T) {
	cbm := NewCircuitBreakerManager()

	// Test creating a circuit breaker with predefined config
	config := FileSystemConfig
	cb := cbm.GetOrCreate("test_filesystem", &config)
	if cb == nil {
		t.Error("Expected non-nil circuit breaker")
	}

	// Test getting an existing circuit breaker
	cb2 := cbm.GetOrCreate("test_filesystem", nil)
	if cb != cb2 {
		t.Error("Expected same circuit breaker instance")
	}

	// Test removing circuit breaker
	cbm.Remove("test_filesystem")
	_, exists := cbm.Get("test_filesystem")
	if exists {
		t.Error("Expected circuit breaker to not exist after removal")
	}
}

func TestPredefinedConfigurations(t *testing.T) {
	tests := []struct {
		name   string
		config resilience.CircuitBreakerConfig
	}{
		{"FileSystemConfig", FileSystemConfig},
		{"WebSocketConfig", WebSocketConfig},
		{"ConfigLoaderConfig", ConfigLoaderConfig},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.config.Name == "" {
				t.Error("Expected non-empty name")
			}

			if test.config.MaxFailures <= 0 {
				t.Errorf("Expected positive MaxFailures, got %d", test.config.MaxFailures)
			}

			if test.config.Timeout <= 0 {
				t.Errorf("Expected positive Timeout, got %v", test.config.Timeout)
			}

			if test.config.MaxRequests <= 0 {
				t.Errorf("Expected positive MaxRequests, got %d", test.config.MaxRequests)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func(context.Context, func(context.Context) error) error
	}{
		{"ExecuteWithFileSystemCircuitBreaker", ExecuteWithFileSystemCircuitBreaker},
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

func TestCircuitBreakerManagerStats(t *testing.T) {
	cbm := NewCircuitBreakerManager()

	// Create some circuit breakers
	cbm.GetOrCreate("test1", &FileSystemConfig)
	cbm.GetOrCreate("test2", &WebSocketConfig)

	stats := cbm.GetAllStats()
	if len(stats) != 2 {
		t.Errorf("Expected 2 stats entries, got %d", len(stats))
	}

	// Verify stats content
	for name, stat := range stats {
		statMap, ok := stat.(map[string]interface{})
		if !ok {
			t.Errorf("Expected stats to be map[string]interface{}, got %T", stat)
			continue
		}

		if statMap["state"] != "Closed" {
			t.Errorf("Expected initial state Closed for %s, got %v", name, statMap["state"])
		}
	}
}

func TestCircuitBreakerManagerConcurrency(t *testing.T) {
	cbm := NewCircuitBreakerManager()
	ctx := context.Background()

	// Test concurrent creation and access
	const numGoroutines = 10
	const numOperations = 50

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// Mix of operations
				switch j % 3 {
				case 0:
					// Create/get circuit breaker
					cb := cbm.GetOrCreate("concurrent_test", &FileSystemConfig)
					if cb == nil {
						t.Errorf("Got nil circuit breaker in goroutine %d", id)
						return
					}

				case 1:
					// Get stats
					stats := cbm.GetAllStats()
					if stats == nil {
						t.Errorf("Got nil stats in goroutine %d", id)
						return
					}

				case 2:
					// Execute operation
					cb := cbm.GetOrCreate("concurrent_test", &FileSystemConfig)
					cb.Execute(ctx, func(ctx context.Context) error {
						return nil
					})
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final state
	stats := cbm.GetAllStats()
	if len(stats) == 0 {
		t.Error("Expected at least one circuit breaker after concurrent operations")
	}
}
