package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/retry"
)

func TestNewTimeoutConfig(t *testing.T) {
	cfg := &config.Config{
		RequestTimeout:         30 * time.Second,
		SessionTimeout:         30 * time.Minute,
		MetricsInterval:        60 * time.Second,
		RetryEnabled:           true,
		RetryMaxAttempts:       3,
		RetryInitialDelay:      100 * time.Millisecond,
		RetryMaxDelay:          30 * time.Second,
		RetryBackoffMultiplier: 2.0,
		RetryJitterPercent:     10,
	}

	timeoutConfig := NewTimeoutConfig(cfg)

	if timeoutConfig == nil {
		t.Error("Expected non-nil timeout config")
	}

	if timeoutConfig.RequestTimeout != cfg.RequestTimeout {
		t.Errorf("Expected RequestTimeout %v, got %v", cfg.RequestTimeout, timeoutConfig.RequestTimeout)
	}

	if timeoutConfig.SessionTimeout != cfg.SessionTimeout {
		t.Errorf("Expected SessionTimeout %v, got %v", cfg.SessionTimeout, timeoutConfig.SessionTimeout)
	}

	if !timeoutConfig.RetryEnabled {
		t.Error("Expected retry enabled")
	}

	if timeoutConfig.RetryConfig.MaxAttempts != cfg.RetryMaxAttempts {
		t.Errorf("Expected MaxAttempts %d, got %d", cfg.RetryMaxAttempts, timeoutConfig.RetryConfig.MaxAttempts)
	}
}

func TestNewTimeoutConfigRetryDisabled(t *testing.T) {
	cfg := &config.Config{
		RequestTimeout:  30 * time.Second,
		SessionTimeout:  30 * time.Minute,
		MetricsInterval: 60 * time.Second,
		RetryEnabled:    false, // Disabled
	}

	timeoutConfig := NewTimeoutConfig(cfg)

	if timeoutConfig.RetryEnabled {
		t.Error("Expected retry disabled")
	}

	if timeoutConfig.RetryConfig.MaxAttempts != 1 {
		t.Errorf("Expected MaxAttempts 1 for disabled retry, got %d", timeoutConfig.RetryConfig.MaxAttempts)
	}
}

func TestTimeoutConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *TimeoutConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &TimeoutConfig{
				RequestTimeout:  30 * time.Second,
				SessionTimeout:  30 * time.Minute,
				CleanupInterval: 60 * time.Second,
				RetryEnabled:    true,
				RetryConfig: retry.RetryConfig{
					MaxAttempts:       3,
					InitialDelay:      100 * time.Millisecond,
					MaxDelay:          30 * time.Second,
					BackoffMultiplier: 2.0,
					JitterMaxPercent:  10,
				},
			},
			wantErr: false,
		},
		{
			name: "request timeout too short",
			config: &TimeoutConfig{
				RequestTimeout:  500 * time.Millisecond, // Too short
				SessionTimeout:  30 * time.Minute,
				CleanupInterval: 60 * time.Second,
				RetryEnabled:    false,
			},
			wantErr: true,
		},
		{
			name: "session timeout too short",
			config: &TimeoutConfig{
				RequestTimeout:  30 * time.Second,
				SessionTimeout:  30 * time.Second, // Too short
				CleanupInterval: 60 * time.Second,
				RetryEnabled:    false,
			},
			wantErr: true,
		},
		{
			name: "invalid retry config",
			config: &TimeoutConfig{
				RequestTimeout:  30 * time.Second,
				SessionTimeout:  30 * time.Minute,
				CleanupInterval: 60 * time.Second,
				RetryEnabled:    true,
				RetryConfig: retry.RetryConfig{
					MaxAttempts:       0, // Invalid
					InitialDelay:      100 * time.Millisecond,
					MaxDelay:          30 * time.Second,
					BackoffMultiplier: 2.0,
					JitterMaxPercent:  10,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimeoutConfigExecuteWithTimeout(t *testing.T) {
	timeoutConfig := &TimeoutConfig{
		RequestTimeout:  30 * time.Second,
		SessionTimeout:  30 * time.Minute,
		CleanupInterval: 60 * time.Second,
		RetryEnabled:    false, // No retry for this test
		RetryConfig: retry.RetryConfig{
			MaxAttempts: 1,
		},
	}

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		return nil
	}

	err := timeoutConfig.ExecuteWithTimeout(ctx, 1*time.Second, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestTimeoutConfigExecuteWithTimeoutRetryEnabled(t *testing.T) {
	timeoutConfig := &TimeoutConfig{
		RequestTimeout:  30 * time.Second,
		SessionTimeout:  30 * time.Minute,
		CleanupInterval: 60 * time.Second,
		RetryEnabled:    true,
		RetryConfig: retry.RetryConfig{
			MaxAttempts:       3,
			InitialDelay:      1 * time.Millisecond,
			MaxDelay:          10 * time.Millisecond,
			BackoffMultiplier: 2.0,
			JitterMaxPercent:  0,
			RetryableErrors:   []error{},
		},
	}

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := timeoutConfig.ExecuteWithTimeout(ctx, 1*time.Second, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls with retry, got %d", callCount)
	}
}

func TestTimeoutConfigExecuteWithRequestTimeout(t *testing.T) {
	timeoutConfig := &TimeoutConfig{
		RequestTimeout:  50 * time.Millisecond,
		SessionTimeout:  30 * time.Minute,
		CleanupInterval: 60 * time.Second,
		RetryEnabled:    false,
		RetryConfig: retry.RetryConfig{
			MaxAttempts: 1,
		},
	}

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		return nil
	}

	err := timeoutConfig.ExecuteWithRequestTimeout(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestTimeoutConfigExecuteWithCustomRetry(t *testing.T) {
	timeoutConfig := &TimeoutConfig{
		RequestTimeout:  30 * time.Second,
		SessionTimeout:  30 * time.Minute,
		CleanupInterval: 60 * time.Second,
		RetryEnabled:    false, // Global retry disabled, but using custom
	}

	customRetryConfig := retry.RetryConfig{
		MaxAttempts:       2,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          5 * time.Millisecond,
		BackoffMultiplier: 1.5,
		JitterMaxPercent:  0,
		RetryableErrors:   []error{},
	}

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := timeoutConfig.ExecuteWithCustomRetry(ctx, customRetryConfig, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls with custom retry, got %d", callCount)
	}
}

func TestInitTimeoutConfig(t *testing.T) {
	// Store original global config
	originalConfig := globalTimeoutConfig
	defer func() {
		globalTimeoutConfig = originalConfig
	}()

	cfg := &config.Config{
		RequestTimeout:         30 * time.Second,
		SessionTimeout:         30 * time.Minute,
		MetricsInterval:        60 * time.Second,
		RetryEnabled:           true,
		RetryMaxAttempts:       3,
		RetryInitialDelay:      100 * time.Millisecond,
		RetryMaxDelay:          30 * time.Second,
		RetryBackoffMultiplier: 2.0,
		RetryJitterPercent:     10,
	}

	err := InitTimeoutConfig(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if globalTimeoutConfig == nil {
		t.Error("Expected global timeout config to be initialized")
	}

	if globalTimeoutConfig.RequestTimeout != cfg.RequestTimeout {
		t.Errorf("Expected RequestTimeout %v, got %v", cfg.RequestTimeout, globalTimeoutConfig.RequestTimeout)
	}
}

func TestInitTimeoutConfigValidationError(t *testing.T) {
	// Store original global config
	originalConfig := globalTimeoutConfig
	defer func() {
		globalTimeoutConfig = originalConfig
	}()

	cfg := &config.Config{
		RequestTimeout:  500 * time.Millisecond, // Too short, should cause validation error
		SessionTimeout:  30 * time.Minute,
		MetricsInterval: 60 * time.Second,
		RetryEnabled:    false,
	}

	err := InitTimeoutConfig(cfg)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	// Global config should not be updated on validation failure
	if globalTimeoutConfig != originalConfig {
		t.Error("Expected global timeout config to remain unchanged after validation error")
	}
}

func TestGetTimeoutConfig(t *testing.T) {
	// Store original global config
	originalConfig := globalTimeoutConfig
	defer func() {
		globalTimeoutConfig = originalConfig
	}()

	// Test when not initialized
	globalTimeoutConfig = nil
	config := GetTimeoutConfig()
	if config != nil {
		t.Error("Expected nil config when not initialized")
	}

	// Test when initialized
	testConfig := &TimeoutConfig{
		RequestTimeout: 30 * time.Second,
	}
	globalTimeoutConfig = testConfig

	config = GetTimeoutConfig()
	if config != testConfig {
		t.Error("Expected to get the same config instance")
	}
}

func TestExecuteWithTimeoutGlobalFunction(t *testing.T) {
	// Store original global config
	originalConfig := globalTimeoutConfig
	defer func() {
		globalTimeoutConfig = originalConfig
	}()

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		return nil
	}

	// Test fallback when global config is nil
	globalTimeoutConfig = nil
	err := ExecuteWithTimeout(ctx, 1*time.Second, operation)
	if err != nil {
		t.Errorf("Expected no error with fallback, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call with fallback, got %d", callCount)
	}

	// Test with global config initialized
	globalTimeoutConfig = &TimeoutConfig{
		RequestTimeout:  30 * time.Second,
		SessionTimeout:  30 * time.Minute,
		CleanupInterval: 60 * time.Second,
		RetryEnabled:    false,
		RetryConfig: retry.RetryConfig{
			MaxAttempts: 1,
		},
	}

	callCount = 0 // Reset
	err = ExecuteWithTimeout(ctx, 1*time.Second, operation)
	if err != nil {
		t.Errorf("Expected no error with global config, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call with global config, got %d", callCount)
	}
}

func TestExecuteWithRequestTimeoutGlobalFunction(t *testing.T) {
	// Store original global config
	originalConfig := globalTimeoutConfig
	defer func() {
		globalTimeoutConfig = originalConfig
	}()

	ctx := context.Background()
	callCount := 0

	operation := func(ctx context.Context) error {
		callCount++
		return nil
	}

	// Test fallback when global config is nil
	globalTimeoutConfig = nil
	err := ExecuteWithRequestTimeout(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error with fallback, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call with fallback, got %d", callCount)
	}

	// Test with global config initialized
	globalTimeoutConfig = &TimeoutConfig{
		RequestTimeout:  50 * time.Millisecond,
		SessionTimeout:  30 * time.Minute,
		CleanupInterval: 60 * time.Second,
		RetryEnabled:    false,
		RetryConfig: retry.RetryConfig{
			MaxAttempts: 1,
		},
	}

	callCount = 0 // Reset
	err = ExecuteWithRequestTimeout(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error with global config, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call with global config, got %d", callCount)
	}
}
