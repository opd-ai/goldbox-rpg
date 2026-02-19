// Package config provides configuration management for the GoldBox RPG Engine.
// It handles environment variable loading, validation, and provides secure defaults
// for production deployment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"goldbox-rpg/pkg/retry"

	"github.com/sirupsen/logrus"
)

// Config represents the server configuration with environment variable support.
// All configuration values can be set via environment variables or will use
// secure defaults appropriate for production deployment.
// Config is thread-safe; all field access should be done through getter methods
// when used concurrently, or by holding the mutex directly.
type Config struct {
	// mu provides thread-safe access to configuration fields when the Config
	// instance is shared across goroutines. Use RLock for reads and Lock for writes.
	mu sync.RWMutex `json:"-"`

	// ServerPort is the port the HTTP server will listen on
	ServerPort int `json:"server_port"`

	// WebDir is the directory containing static web files
	WebDir string `json:"web_dir"`

	// SessionTimeout is the duration after which inactive sessions expire
	SessionTimeout time.Duration `json:"session_timeout"`

	// LogLevel controls the logging verbosity (debug, info, warn, error)
	LogLevel string `json:"log_level"`

	// AllowedOrigins is a list of allowed WebSocket origins for CORS
	AllowedOrigins []string `json:"allowed_origins"`

	// MaxRequestSize is the maximum size of incoming requests in bytes
	MaxRequestSize int64 `json:"max_request_size"`

	// EnableDevMode enables development-friendly settings (broader CORS, verbose logging)
	EnableDevMode bool `json:"enable_dev_mode"`

	// RequestTimeout is the maximum duration for processing requests
	RequestTimeout time.Duration `json:"request_timeout"`

	// Performance monitoring configuration

	// EnableProfiling enables pprof profiling endpoints (/debug/pprof)
	EnableProfiling bool `json:"enable_profiling"`

	// ProfilingPort is the port for the profiling server (0 = disabled, same port as main server)
	ProfilingPort int `json:"profiling_port"`

	// MetricsInterval is how often performance metrics are collected
	MetricsInterval time.Duration `json:"metrics_interval"`

	// AlertingEnabled enables performance alerting
	AlertingEnabled bool `json:"alerting_enabled"`

	// AlertingInterval is how often performance alerts are checked
	AlertingInterval time.Duration `json:"alerting_interval"`

	// Rate limiting configuration

	// RateLimitEnabled enables rate limiting middleware
	RateLimitEnabled bool `json:"rate_limit_enabled"`

	// RateLimitRequestsPerSecond is the number of requests allowed per second per IP
	RateLimitRequestsPerSecond float64 `json:"rate_limit_requests_per_second"`

	// RateLimitBurst is the maximum number of requests allowed in a burst per IP
	RateLimitBurst int `json:"rate_limit_burst"`

	// RateLimitCleanupInterval is how often to clean up expired rate limiters
	RateLimitCleanupInterval time.Duration `json:"rate_limit_cleanup_interval"`

	// Retry configuration

	// RetryEnabled enables retry logic for transient failures
	RetryEnabled bool `json:"retry_enabled"`

	// RetryMaxAttempts is the maximum number of retry attempts (including initial attempt)
	RetryMaxAttempts int `json:"retry_max_attempts"`

	// RetryInitialDelay is the initial delay before the first retry
	RetryInitialDelay time.Duration `json:"retry_initial_delay"`

	// RetryMaxDelay is the maximum delay between retries
	RetryMaxDelay time.Duration `json:"retry_max_delay"`

	// RetryBackoffMultiplier is the multiplier for exponential backoff (typically 2.0)
	RetryBackoffMultiplier float64 `json:"retry_backoff_multiplier"`

	// RetryJitterPercent is the maximum percentage of jitter to add (0-100)
	RetryJitterPercent int `json:"retry_jitter_percent"`

	// Persistence configuration

	// DataDir is the directory where game state and character data is persisted
	DataDir string `json:"data_dir"`

	// AutoSaveInterval is how often game state is automatically saved to disk
	AutoSaveInterval time.Duration `json:"auto_save_interval"`

	// EnablePersistence enables automatic game state persistence
	EnablePersistence bool `json:"enable_persistence"`

	// Server lifecycle timeouts

	// BootstrapTimeout is the maximum duration for bootstrap game generation
	BootstrapTimeout time.Duration `json:"bootstrap_timeout"`

	// ShutdownTimeout is the maximum duration for graceful server shutdown
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// ShutdownGracePeriod is the grace period after shutdown before forcing exit
	ShutdownGracePeriod time.Duration `json:"shutdown_grace_period"`
}

// Load creates a new Config instance by reading from environment variables
// and applying secure defaults. It validates all configuration values and
// returns an error if any required values are missing or invalid.
func Load() (*Config, error) {
	logrus.WithFields(logrus.Fields{
		"function": "Load",
		"package":  "config",
	}).Debug("entering Load")

	config := &Config{
		// Secure defaults for production deployment
		ServerPort:     getEnvAsInt("SERVER_PORT", 8080),
		WebDir:         getEnvAsString("WEB_DIR", "./web"),
		SessionTimeout: getEnvAsDuration("SESSION_TIMEOUT", 30*time.Minute),
		LogLevel:       getEnvAsString("LOG_LEVEL", "info"),
		AllowedOrigins: getEnvAsStringSlice("ALLOWED_ORIGINS", []string{}),
		MaxRequestSize: getEnvAsInt64("MAX_REQUEST_SIZE", 1*1024*1024), // 1MB default
		EnableDevMode:  getEnvAsBool("ENABLE_DEV_MODE", true),          // Default to dev mode for easier setup
		RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 30*time.Second),

		// Performance monitoring defaults
		EnableProfiling:  getEnvAsBool("ENABLE_PROFILING", false),               // Disabled by default for security
		ProfilingPort:    getEnvAsInt("PROFILING_PORT", 0),                      // 0 = use same port as main server
		MetricsInterval:  getEnvAsDuration("METRICS_INTERVAL", 30*time.Second),  // Collect metrics every 30s
		AlertingEnabled:  getEnvAsBool("ALERTING_ENABLED", true),                // Enable alerting by default
		AlertingInterval: getEnvAsDuration("ALERTING_INTERVAL", 30*time.Second), // Check alerts every 30s

		// Rate limiting defaults
		RateLimitEnabled:           getEnvAsBool("RATE_LIMIT_ENABLED", false),                      // Disabled by default
		RateLimitRequestsPerSecond: getEnvAsFloat64("RATE_LIMIT_REQUESTS_PER_SECOND", 5),           // 5 requests per second default
		RateLimitBurst:             getEnvAsInt("RATE_LIMIT_BURST", 10),                            // 10 requests burst default
		RateLimitCleanupInterval:   getEnvAsDuration("RATE_LIMIT_CLEANUP_INTERVAL", 1*time.Minute), // 1 minute cleanup interval

		// Retry defaults
		RetryEnabled:           getEnvAsBool("RETRY_ENABLED", true),                           // Enabled by default
		RetryMaxAttempts:       getEnvAsInt("RETRY_MAX_ATTEMPTS", 3),                          // 3 attempts default
		RetryInitialDelay:      getEnvAsDuration("RETRY_INITIAL_DELAY", 100*time.Millisecond), // 100ms initial delay
		RetryMaxDelay:          getEnvAsDuration("RETRY_MAX_DELAY", 30*time.Second),           // 30s max delay
		RetryBackoffMultiplier: getEnvAsFloat64("RETRY_BACKOFF_MULTIPLIER", 2.0),              // 2.0 backoff multiplier
		RetryJitterPercent:     getEnvAsInt("RETRY_JITTER_PERCENT", 10),                       // 10% jitter

		// Persistence defaults
		DataDir:           getEnvAsString("DATA_DIR", "./data"),                   // ./data directory default
		AutoSaveInterval:  getEnvAsDuration("AUTO_SAVE_INTERVAL", 30*time.Second), // 30s auto-save interval
		EnablePersistence: getEnvAsBool("ENABLE_PERSISTENCE", true),               // Enabled by default

		// Server lifecycle timeout defaults
		BootstrapTimeout:    getEnvAsDuration("BOOTSTRAP_TIMEOUT", 60*time.Second),    // 60s bootstrap timeout
		ShutdownTimeout:     getEnvAsDuration("SHUTDOWN_TIMEOUT", 30*time.Second),     // 30s shutdown timeout
		ShutdownGracePeriod: getEnvAsDuration("SHUTDOWN_GRACE_PERIOD", 1*time.Second), // 1s grace period
	}

	logrus.WithFields(logrus.Fields{
		"function":    "Load",
		"package":     "config",
		"server_port": config.ServerPort,
		"dev_mode":    config.EnableDevMode,
		"log_level":   config.LogLevel,
	}).Debug("configuration loaded, starting validation")

	// Validate configuration
	if err := config.validate(); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "Load",
			"package":  "config",
			"error":    err,
		}).Error("configuration validation failed")
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":    "Load",
		"package":     "config",
		"server_port": config.ServerPort,
		"dev_mode":    config.EnableDevMode,
		"log_level":   config.LogLevel,
	}).Debug("exiting Load - configuration successfully loaded and validated")

	return config, nil
}

// validate checks that all configuration values are valid and consistent.
// validate performs comprehensive configuration validation with multiple checks.
// This method coordinates validation of all configuration sections including
// server settings, timeouts, rate limiting, and retry policies.
func (c *Config) validate() error {
	if err := c.validateServerSettings(); err != nil {
		return err
	}

	if err := c.validateTimeouts(); err != nil {
		return err
	}

	if err := c.validateSecuritySettings(); err != nil {
		return err
	}

	if err := c.validateRateLimitConfig(); err != nil {
		return err
	}

	if err := c.validateRetryConfig(); err != nil {
		return err
	}

	return nil
}

// validateServerSettings checks server port and log level configuration.
// Ensures the server port is within valid range (1-65535) and log level
// is one of the supported values (debug, info, warn, error).
func (c *Config) validateServerSettings() error {
	// Validate server port range
	if c.ServerPort < 1 || c.ServerPort > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.ServerPort)
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	found := false
	for _, level := range validLogLevels {
		if strings.ToLower(c.LogLevel) == level {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("log level must be one of %v, got %s", validLogLevels, c.LogLevel)
	}

	return nil
}

// validateTimeouts ensures timeout values meet minimum requirements.
// Session timeout must be at least 1 minute and request timeout must be
// at least 1 second to prevent performance issues.
func (c *Config) validateTimeouts() error {
	if c.SessionTimeout < time.Minute {
		return fmt.Errorf("session timeout must be at least 1 minute, got %v", c.SessionTimeout)
	}

	if c.RequestTimeout < time.Second {
		return fmt.Errorf("request timeout must be at least 1 second, got %v", c.RequestTimeout)
	}

	return nil
}

// validateSecuritySettings checks security-related configuration.
// Validates request size limits and ensures production mode has proper
// origin allowlist configuration for WebSocket security.
func (c *Config) validateSecuritySettings() error {
	// Validate request size
	if c.MaxRequestSize < 1024 { // 1KB minimum
		return fmt.Errorf("max request size must be at least 1024 bytes, got %d", c.MaxRequestSize)
	}

	// In production mode, require explicit origin allowlist
	if !c.EnableDevMode && len(c.AllowedOrigins) == 0 {
		return fmt.Errorf("allowed origins must be specified when dev mode is disabled")
	}

	return nil
}

// validateRateLimitConfig ensures rate limiting parameters are valid when enabled.
// Checks that requests per second and burst values are positive numbers
// to prevent division by zero and ensure meaningful rate limiting.
func (c *Config) validateRateLimitConfig() error {
	if c.RateLimitEnabled {
		if c.RateLimitRequestsPerSecond <= 0 {
			return fmt.Errorf("rate limit requests per second must be greater than 0 when rate limiting is enabled")
		}
		if c.RateLimitBurst <= 0 {
			return fmt.Errorf("rate limit burst must be greater than 0 when rate limiting is enabled")
		}
	}

	return nil
}

// validateRetryConfig ensures retry policy parameters are valid when enabled.
// Validates attempt counts, delay values, backoff multiplier, and jitter
// percentage to ensure retry behavior functions as expected.
func (c *Config) validateRetryConfig() error {
	if c.RetryEnabled {
		if c.RetryMaxAttempts < 1 {
			return fmt.Errorf("retry max attempts must be at least 1 when retry is enabled")
		}
		if c.RetryInitialDelay < 0 {
			return fmt.Errorf("retry initial delay must be non-negative when retry is enabled")
		}
		if c.RetryMaxDelay < c.RetryInitialDelay {
			return fmt.Errorf("retry max delay must be greater than or equal to initial delay when retry is enabled")
		}
		if c.RetryBackoffMultiplier <= 1.0 {
			return fmt.Errorf("retry backoff multiplier must be greater than 1.0 when retry is enabled")
		}
		if c.RetryJitterPercent < 0 || c.RetryJitterPercent > 100 {
			return fmt.Errorf("retry jitter percent must be between 0 and 100 when retry is enabled")
		}
	}

	return nil
}

// OriginAllowed checks if the given origin is allowed for WebSocket connections.
// In development mode, all origins are allowed. In production mode, only explicitly
// allowed origins are permitted. This method is thread-safe.
func (c *Config) OriginAllowed(origin string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// In development mode, allow all origins for convenience
	if c.EnableDevMode {
		return true
	}

	// In production mode, check against allowlist
	for _, allowed := range c.AllowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}

// GetRetryConfig creates a retry.RetryConfig from the current configuration.
// This converts the application-level retry settings into the format expected
// by the retry package. The returned configuration can be used directly with
// retry.NewRetrier() to create a retrier instance.
func (c *Config) GetRetryConfig() retry.RetryConfig {
	return retry.RetryConfig{
		MaxAttempts:       c.RetryMaxAttempts,
		InitialDelay:      c.RetryInitialDelay,
		MaxDelay:          c.RetryMaxDelay,
		BackoffMultiplier: c.RetryBackoffMultiplier,
		JitterMaxPercent:  c.RetryJitterPercent,
		RetryableErrors:   []error{}, // Will use default error classification
	}
}

// Helper functions for environment variable parsing with type safety and defaults

func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim whitespace
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}

func getEnvAsFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
