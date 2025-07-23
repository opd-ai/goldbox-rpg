// Package config provides configuration management for the GoldBox RPG Engine.
// It handles environment variable loading, validation, and provides secure defaults
// for production deployment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the server configuration with environment variable support.
// All configuration values can be set via environment variables or will use
// secure defaults appropriate for production deployment.
type Config struct {
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

	// Rate limiting and resilience configuration

	// RateLimit is the number of requests allowed per second per IP
	RateLimit float64 `json:"rate_limit"`

	// RateBurst is the maximum burst size for rate limiting
	RateBurst int `json:"rate_burst"`
}

// Load creates a new Config instance by reading from environment variables
// and applying secure defaults. It validates all configuration values and
// returns an error if any required values are missing or invalid.
func Load() (*Config, error) {
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

		// Rate limiting defaults - conservative for production safety
		RateLimit: getEnvAsFloat64("RATE_LIMIT", 100.0), // 100 requests per second per IP
		RateBurst: getEnvAsInt("RATE_BURST", 200),       // Allow burst of 200 requests
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validate checks that all configuration values are valid and consistent.
func (c *Config) validate() error {
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

	// Validate timeouts
	if c.SessionTimeout < time.Minute {
		return fmt.Errorf("session timeout must be at least 1 minute, got %v", c.SessionTimeout)
	}

	if c.RequestTimeout < time.Second {
		return fmt.Errorf("request timeout must be at least 1 second, got %v", c.RequestTimeout)
	}

	// Validate request size
	if c.MaxRequestSize < 1024 { // 1KB minimum
		return fmt.Errorf("max request size must be at least 1024 bytes, got %d", c.MaxRequestSize)
	}

	// In production mode, require explicit origin allowlist
	if !c.EnableDevMode && len(c.AllowedOrigins) == 0 {
		return fmt.Errorf("allowed origins must be specified when dev mode is disabled")
	}

	// Validate rate limiting settings
	if c.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be greater than 0, got %f", c.RateLimit)
	}

	if c.RateBurst < 0 {
		return fmt.Errorf("rate burst size must be at least 0, got %d", c.RateBurst)
	}

	return nil
}

// IsOriginAllowed checks if the given origin is allowed for WebSocket connections.
// In development mode, all origins are allowed. In production mode, only explicitly
// allowed origins are permitted.
func (c *Config) IsOriginAllowed(origin string) bool {
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
