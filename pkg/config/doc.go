// Package config provides configuration management for the GoldBox RPG Engine.
//
// This package handles environment variable loading with type-safe parsing,
// applies secure production defaults, and performs extensive validation of
// all configuration values.
//
// # Loading Configuration
//
// Configuration is loaded from environment variables with the GOLDBOX_ prefix:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Environment Variables
//
// Server settings:
//   - GOLDBOX_PORT: HTTP port (default: 8080)
//   - GOLDBOX_WEB_DIR: Static file directory (default: "web/static")
//   - GOLDBOX_LOG_LEVEL: Logging verbosity (default: "info")
//
// Timeouts:
//   - GOLDBOX_SESSION_TIMEOUT: Session inactivity timeout (default: 30m)
//   - GOLDBOX_REQUEST_TIMEOUT: HTTP request timeout (default: 30s)
//
// Security:
//   - GOLDBOX_DEV_MODE: Enable development mode (default: false)
//   - GOLDBOX_ALLOWED_ORIGINS: CORS allowed origins (comma-separated)
//   - GOLDBOX_MAX_REQUEST_SIZE: Maximum request body size (default: 1MB)
//
// Rate limiting:
//   - GOLDBOX_RATE_LIMIT: Requests per second (default: 10)
//   - GOLDBOX_RATE_BURST: Burst allowance (default: 20)
//
// Retry policy:
//   - GOLDBOX_MAX_RETRY_ATTEMPTS: Maximum retries (default: 3)
//   - GOLDBOX_RETRY_INITIAL_DELAY: First retry delay (default: 100ms)
//   - GOLDBOX_RETRY_MAX_DELAY: Maximum retry delay (default: 5s)
//   - GOLDBOX_RETRY_BACKOFF_MULTIPLIER: Backoff factor (default: 2.0)
//
// Persistence:
//   - GOLDBOX_DATA_DIR: Data storage directory (default: "data")
//   - GOLDBOX_AUTO_SAVE_INTERVAL: Auto-save frequency (default: 5m)
//
// # Validation
//
// All configuration values are validated on load:
//   - Port must be in valid range (1-65535)
//   - Timeouts must meet minimum requirements
//   - Rate limit values must be positive
//   - Retry configuration must be sensible
//
// # CORS Support
//
// Use IsOriginAllowed to check WebSocket origins:
//
//	if cfg.IsOriginAllowed(origin) {
//	    // Allow connection
//	}
//
// In development mode (DevMode=true), all origins are allowed.
//
// # Retry Configuration
//
// GetRetryConfig returns a retry.RetryConfig that can be used directly
// with the retry package:
//
//	retryConfig := cfg.GetRetryConfig()
//	retrier := retry.NewRetrier(retryConfig)
package config
