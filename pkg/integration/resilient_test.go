package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"goldbox-rpg/pkg/resilience"
	"goldbox-rpg/pkg/retry"
)

func TestNewResilientExecutor(t *testing.T) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("test")
	retryConfig := retry.DefaultRetryConfig()

	executor := NewResilientExecutor(cbConfig, retryConfig)

	if executor == nil {
		t.Error("Expected non-nil executor")
	}

	if executor.circuitBreaker == nil {
		t.Error("Expected non-nil circuit breaker")
	}

	if executor.retrier == nil {
		t.Error("Expected non-nil retrier")
	}
}

func TestResilientExecutorSuccess(t *testing.T) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("test")
	retryConfig := retry.DefaultRetryConfig()
	executor := NewResilientExecutor(cbConfig, retryConfig)

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		return nil
	}

	err := executor.Execute(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestResilientExecutorRetrySuccess(t *testing.T) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("test")
	retryConfig := retry.RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		JitterMaxPercent:  0,
		RetryableErrors:   []error{},
	}
	executor := NewResilientExecutor(cbConfig, retryConfig)

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := executor.Execute(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestResilientExecutorCircuitBreakerOpen(t *testing.T) {
	cbConfig := resilience.CircuitBreakerConfig{
		Name:        "test",
		MaxFailures: 1,
		Timeout:     100 * time.Millisecond,
		MaxRequests: 1,
	}
	retryConfig := retry.RetryConfig{
		MaxAttempts:       5,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		JitterMaxPercent:  0,
		RetryableErrors:   []error{},
	}
	executor := NewResilientExecutor(cbConfig, retryConfig)

	ctx := context.Background()
	failureErr := errors.New("persistent failure")

	// First, cause the circuit breaker to open
	operation := func(ctx context.Context) error {
		return failureErr
	}

	err := executor.Execute(ctx, operation)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Subsequent calls should fail fast due to open circuit breaker
	err2 := executor.Execute(ctx, operation)
	if err2 == nil {
		t.Error("Expected error due to open circuit breaker, got nil")
	}
}

func TestResilientExecutorGetStats(t *testing.T) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("test")
	retryConfig := retry.DefaultRetryConfig()
	executor := NewResilientExecutor(cbConfig, retryConfig)

	stats := executor.GetStats()
	if stats == nil {
		t.Error("Expected non-nil stats")
	}

	// Check for circuit breaker stats
	if _, exists := stats["circuit_breaker_name"]; !exists {
		t.Error("Expected circuit_breaker_name in stats")
	}

	if _, exists := stats["circuit_breaker_state"]; !exists {
		t.Error("Expected circuit_breaker_state in stats")
	}
}

func TestPredefinedExecutors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		executor *ResilientExecutor
	}{
		{"FileSystemExecutor", FileSystemExecutor},
		{"NetworkExecutor", NetworkExecutor},
		{"ConfigLoaderExecutor", ConfigLoaderExecutor},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.executor == nil {
				t.Error("Expected non-nil predefined executor")
			}

			// Test basic operation
			callCount := 0
			operation := func(ctx context.Context) error {
				callCount++
				return nil
			}

			err := test.executor.Execute(ctx, operation)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if callCount != 1 {
				t.Errorf("Expected 1 call, got %d", callCount)
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		function func(context.Context, func(context.Context) error) error
	}{
		{"ExecuteFileSystemOperation", ExecuteFileSystemOperation},
		{"ExecuteNetworkOperation", ExecuteNetworkOperation},
		{"ExecuteConfigOperation", ExecuteConfigOperation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			callCount := 0
			operation := func(ctx context.Context) error {
				callCount++
				return nil
			}

			err := test.function(ctx, operation)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if callCount != 1 {
				t.Errorf("Expected 1 call, got %d", callCount)
			}
		})
	}
}

func TestCreateCustomExecutor(t *testing.T) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("custom")
	retryConfig := retry.DefaultRetryConfig()

	executor := CreateCustomExecutor("custom_test", cbConfig, retryConfig)
	if executor == nil {
		t.Error("Expected non-nil custom executor")
	}

	stats := executor.GetStats()
	if stats["circuit_breaker_name"] != "custom_test" {
		t.Errorf("Expected circuit breaker name 'custom_test', got %v", stats["circuit_breaker_name"])
	}
}

func TestWithRetryDisabled(t *testing.T) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("test")
	executor := WithRetryDisabled(cbConfig)

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		return errors.New("failure")
	}

	err := executor.Execute(ctx, operation)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Should only be called once (no retry)
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retry), got %d", callCount)
	}
}

func TestWithCircuitBreakerDisabled(t *testing.T) {
	retryConfig := retry.RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		JitterMaxPercent:  0,
		RetryableErrors:   []error{},
	}
	executor := WithCircuitBreakerDisabled(retryConfig)

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := executor.Execute(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should retry (circuit breaker effectively disabled)
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestExecuteResilient(t *testing.T) {
	ctx := context.Background()

	// Test with default configuration
	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return nil
	}

	err := ExecuteResilient(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestExecuteResilientWithOptions(t *testing.T) {
	ctx := context.Background()

	customRetryConfig := retry.RetryConfig{
		MaxAttempts:       2,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          5 * time.Millisecond,
		BackoffMultiplier: 1.5,
		JitterMaxPercent:  0,
		RetryableErrors:   []error{},
	}

	customCBConfig := resilience.CircuitBreakerConfig{
		Name:        "custom",
		MaxFailures: 10,
		Timeout:     100 * time.Millisecond,
		MaxRequests: 5,
	}

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := ExecuteResilient(ctx, operation,
		ConfigureRetry(customRetryConfig),
		ConfigureCircuitBreaker(customCBConfig),
	)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestResilientExecutorContextCancellation(t *testing.T) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("test")
	retryConfig := retry.RetryConfig{
		MaxAttempts:       5,
		InitialDelay:      50 * time.Millisecond,
		MaxDelay:          200 * time.Millisecond,
		BackoffMultiplier: 2.0,
		JitterMaxPercent:  0,
		RetryableErrors:   []error{},
	}
	executor := NewResilientExecutor(cbConfig, retryConfig)

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return errors.New("failure")
	}

	// Cancel context after first attempt
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := executor.Execute(ctx, operation)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}

	// Should only be called once before cancellation
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestResilientExecutorConcurrency(t *testing.T) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("test")
	retryConfig := retry.DefaultRetryConfig()
	executor := NewResilientExecutor(cbConfig, retryConfig)

	ctx := context.Background()
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Launch concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			operation := func(ctx context.Context) error {
				return nil // Always succeed
			}

			err := executor.Execute(ctx, operation)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Expected no error from goroutine, got %v", err)
		}
	}
}

// Benchmark tests
func BenchmarkResilientExecutorSuccess(b *testing.B) {
	cbConfig := resilience.DefaultCircuitBreakerConfig("test")
	retryConfig := retry.DefaultRetryConfig()
	executor := NewResilientExecutor(cbConfig, retryConfig)
	ctx := context.Background()

	operation := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.Execute(ctx, operation)
	}
}

func BenchmarkConvenienceFunction(b *testing.B) {
	ctx := context.Background()
	operation := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExecuteFileSystemOperation(ctx, operation)
	}
}
