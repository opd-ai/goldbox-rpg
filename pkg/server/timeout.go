// Package server provides timeout and retry configuration utilities for server operations.
// This module extends the server package with configurable timeout and retry logic
// for handling transient failures and preventing timeouts in production environments.
package server

import (
	"context"
	"fmt"
	"time"

	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/retry"

	"github.com/sirupsen/logrus"
)

// TimeoutConfig holds timeout-related configuration for server operations
type TimeoutConfig struct {
	// RequestTimeout is the maximum duration for processing individual requests
	RequestTimeout time.Duration

	// SessionTimeout is the duration after which inactive sessions expire
	SessionTimeout time.Duration

	// CleanupInterval is how often cleanup operations run
	CleanupInterval time.Duration

	// RetryEnabled enables retry logic for transient failures
	RetryEnabled bool

	// RetryConfig holds the retry configuration parameters
	RetryConfig retry.RetryConfig
}

// NewTimeoutConfig creates a timeout configuration from application config
func NewTimeoutConfig(cfg *config.Config) *TimeoutConfig {
	var retryConfig retry.RetryConfig
	if cfg.RetryEnabled {
		retryConfig = retry.RetryConfig{
			MaxAttempts:       cfg.RetryMaxAttempts,
			InitialDelay:      cfg.RetryInitialDelay,
			MaxDelay:          cfg.RetryMaxDelay,
			BackoffMultiplier: cfg.RetryBackoffMultiplier,
			JitterMaxPercent:  cfg.RetryJitterPercent,
			RetryableErrors:   []error{context.DeadlineExceeded}, // Default retryable errors
		}
	} else {
		// Disabled retry configuration (only one attempt)
		retryConfig = retry.RetryConfig{
			MaxAttempts:       1,
			InitialDelay:      0,
			MaxDelay:          0,
			BackoffMultiplier: 1.0,
			JitterMaxPercent:  0,
			RetryableErrors:   []error{},
		}
	}

	return &TimeoutConfig{
		RequestTimeout:  cfg.RequestTimeout,
		SessionTimeout:  cfg.SessionTimeout,
		CleanupInterval: cfg.MetricsInterval, // Reuse metrics interval for cleanup
		RetryEnabled:    cfg.RetryEnabled,
		RetryConfig:     retryConfig,
	}
}

// ExecuteWithTimeout runs an operation with timeout and optional retry logic
func (tc *TimeoutConfig) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, operation func(context.Context) error) error {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if tc.RetryEnabled {
		// Use retry logic with timeout context
		retrier := retry.NewRetrier(tc.RetryConfig)
		return retrier.Execute(timeoutCtx, operation)
	} else {
		// Execute directly with timeout context
		return operation(timeoutCtx)
	}
}

// ExecuteWithRequestTimeout executes an operation with the configured request timeout
func (tc *TimeoutConfig) ExecuteWithRequestTimeout(ctx context.Context, operation func(context.Context) error) error {
	return tc.ExecuteWithTimeout(ctx, tc.RequestTimeout, operation)
}

// ExecuteWithCustomRetry executes an operation with custom retry configuration
func (tc *TimeoutConfig) ExecuteWithCustomRetry(ctx context.Context, retryConfig retry.RetryConfig, operation func(context.Context) error) error {
	retrier := retry.NewRetrier(retryConfig)
	return retrier.Execute(ctx, operation)
}

// LogTimeoutConfig logs the current timeout configuration for debugging
func (tc *TimeoutConfig) LogTimeoutConfig() {
	logger := logrus.WithField("component", "TimeoutConfig")

	logger.WithFields(logrus.Fields{
		"request_timeout":     tc.RequestTimeout,
		"session_timeout":     tc.SessionTimeout,
		"cleanup_interval":    tc.CleanupInterval,
		"retry_enabled":       tc.RetryEnabled,
		"retry_max_attempts":  tc.RetryConfig.MaxAttempts,
		"retry_initial_delay": tc.RetryConfig.InitialDelay,
		"retry_max_delay":     tc.RetryConfig.MaxDelay,
	}).Info("Timeout and retry configuration loaded")
}

// Validate checks that the timeout configuration values are reasonable
func (tc *TimeoutConfig) Validate() error {
	if tc.RequestTimeout < time.Second {
		return fmt.Errorf("request timeout must be at least 1 second, got %v", tc.RequestTimeout)
	}

	if tc.SessionTimeout < time.Minute {
		return fmt.Errorf("session timeout must be at least 1 minute, got %v", tc.SessionTimeout)
	}

	if tc.RetryEnabled {
		if tc.RetryConfig.MaxAttempts < 1 {
			return fmt.Errorf("retry max attempts must be at least 1 when retry is enabled")
		}

		if tc.RetryConfig.InitialDelay < 0 {
			return fmt.Errorf("retry initial delay must be non-negative")
		}

		if tc.RetryConfig.MaxDelay < tc.RetryConfig.InitialDelay {
			return fmt.Errorf("retry max delay must be greater than or equal to initial delay")
		}
	}

	return nil
}

// Global timeout configuration instance (initialized by server startup)
var globalTimeoutConfig *TimeoutConfig

// InitTimeoutConfig initializes the global timeout configuration
func InitTimeoutConfig(cfg *config.Config) error {
	timeoutConfig := NewTimeoutConfig(cfg)

	if err := timeoutConfig.Validate(); err != nil {
		return fmt.Errorf("invalid timeout configuration: %w", err)
	}

	globalTimeoutConfig = timeoutConfig
	timeoutConfig.LogTimeoutConfig()

	return nil
}

// GetTimeoutConfig returns the global timeout configuration
func GetTimeoutConfig() *TimeoutConfig {
	return globalTimeoutConfig
}

// ExecuteWithTimeout is a convenience function for timeout execution with global config
func ExecuteWithTimeout(ctx context.Context, timeout time.Duration, operation func(context.Context) error) error {
	if globalTimeoutConfig == nil {
		// Fallback to direct execution if not initialized
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return operation(timeoutCtx)
	}

	return globalTimeoutConfig.ExecuteWithTimeout(ctx, timeout, operation)
}

// ExecuteWithRequestTimeout is a convenience function using the configured request timeout
func ExecuteWithRequestTimeout(ctx context.Context, operation func(context.Context) error) error {
	if globalTimeoutConfig == nil {
		// Fallback to 30 second timeout if not initialized
		return ExecuteWithTimeout(ctx, 30*time.Second, operation)
	}

	return globalTimeoutConfig.ExecuteWithRequestTimeout(ctx, operation)
}
