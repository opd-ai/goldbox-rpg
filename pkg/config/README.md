# Configuration Package

This package provides configuration management and environment variable handling for the GoldBox RPG Engine.

## Overview

The configuration package centralizes all configuration management for the engine, providing a unified interface for loading settings from environment variables with secure defaults. Configuration is automatically validated on load.

## Features

- **Environment Variable Support**: Automatic loading from environment variables
- **Default Values**: Sensible, secure defaults for all configuration options
- **Built-in Validation**: Configuration validation with detailed error reporting
- **Type Safety**: Strongly typed configuration structure
- **Production-Ready Defaults**: Secure settings appropriate for production deployment

## Configuration Structure

The `Config` struct is a flat configuration with the following sections:

```go
type Config struct {
    // Server settings
    ServerPort     int           // HTTP server port (env: SERVER_PORT, default: 8080)
    WebDir         string        // Static web files directory (env: WEB_DIR, default: "./web")
    SessionTimeout time.Duration // Inactive session expiry (env: SESSION_TIMEOUT, default: 30m)
    LogLevel       string        // Logging verbosity: debug, info, warn, error (env: LOG_LEVEL, default: "info")
    AllowedOrigins []string      // WebSocket CORS origins (env: ALLOWED_ORIGINS, default: [])
    MaxRequestSize int64         // Maximum request size in bytes (env: MAX_REQUEST_SIZE, default: 1MB)
    EnableDevMode  bool          // Enable development mode (env: ENABLE_DEV_MODE, default: true)
    RequestTimeout time.Duration // Maximum request processing time (env: REQUEST_TIMEOUT, default: 30s)

    // Performance monitoring
    EnableProfiling  bool          // Enable pprof endpoints (env: ENABLE_PROFILING, default: false)
    ProfilingPort    int           // Profiling server port (env: PROFILING_PORT, default: 0)
    MetricsInterval  time.Duration // Metrics collection interval (env: METRICS_INTERVAL, default: 30s)
    AlertingEnabled  bool          // Enable performance alerting (env: ALERTING_ENABLED, default: true)
    AlertingInterval time.Duration // Alert check interval (env: ALERTING_INTERVAL, default: 30s)

    // Rate limiting
    RateLimitEnabled           bool          // Enable rate limiting (env: RATE_LIMIT_ENABLED, default: false)
    RateLimitRequestsPerSecond float64       // Requests per second per IP (env: RATE_LIMIT_REQUESTS_PER_SECOND, default: 5)
    RateLimitBurst             int           // Maximum burst requests (env: RATE_LIMIT_BURST, default: 10)
    RateLimitCleanupInterval   time.Duration // Cleanup interval (env: RATE_LIMIT_CLEANUP_INTERVAL, default: 1m)

    // Retry settings
    RetryEnabled           bool          // Enable retry logic (env: RETRY_ENABLED, default: true)
    RetryMaxAttempts       int           // Maximum retry attempts (env: RETRY_MAX_ATTEMPTS, default: 3)
    RetryInitialDelay      time.Duration // Initial retry delay (env: RETRY_INITIAL_DELAY, default: 100ms)
    RetryMaxDelay          time.Duration // Maximum retry delay (env: RETRY_MAX_DELAY, default: 30s)
    RetryBackoffMultiplier float64       // Exponential backoff multiplier (env: RETRY_BACKOFF_MULTIPLIER, default: 2.0)
    RetryJitterPercent     int           // Jitter percentage 0-100 (env: RETRY_JITTER_PERCENT, default: 10)

    // Persistence
    DataDir           string        // Game state directory (env: DATA_DIR, default: "./data")
    AutoSaveInterval  time.Duration // Auto-save interval (env: AUTO_SAVE_INTERVAL, default: 30s)
    EnablePersistence bool          // Enable persistence (env: ENABLE_PERSISTENCE, default: true)
}
```

## Usage

### Basic Configuration Loading

```go
import "goldbox-rpg/pkg/config"

// Load configuration from environment variables with defaults
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load configuration:", err)
}

// Access configuration values
fmt.Printf("Server will run on port: %d\n", cfg.ServerPort)
fmt.Printf("Session timeout: %v\n", cfg.SessionTimeout)
fmt.Printf("Dev mode enabled: %v\n", cfg.EnableDevMode)
```

### Environment Variable Configuration

```bash
# Set environment variables
export SERVER_PORT=9000
export LOG_LEVEL=debug
export SESSION_TIMEOUT=45m
export ALLOWED_ORIGINS="https://game.example.com,https://admin.example.com"
export ENABLE_DEV_MODE=false

# Run application (will use environment variables)
./bin/server
```

### Origin Validation

```go
// Check if an origin is allowed for WebSocket connections
if cfg.OriginAllowed("https://example.com") {
    // Origin is allowed
}

// In dev mode, all origins are allowed
// In production mode (ENABLE_DEV_MODE=false), only explicitly listed origins are allowed
```

### Retry Configuration Integration

```go
import (
    "goldbox-rpg/pkg/config"
    "goldbox-rpg/pkg/retry"
)

// Get retry configuration for use with pkg/retry
retryConfig := cfg.GetRetryConfig()

// Create a retrier directly with the returned config
retrier := retry.NewRetrier(retryConfig)

// The returned retry.RetryConfig contains:
// - MaxAttempts
// - InitialDelay
// - MaxDelay
// - BackoffMultiplier
// - JitterMaxPercent
// - RetryableErrors (empty by default)
```

## Built-in Validation

Configuration is automatically validated during `Load()`. The following checks are performed:

### Server Settings
- Server port must be between 1 and 65535
- Log level must be one of: `debug`, `info`, `warn`, `error`

### Timeouts
- Session timeout must be at least 1 minute
- Request timeout must be at least 1 second

### Security Settings
- Max request size must be at least 1024 bytes (1KB)
- In production mode (`EnableDevMode=false`), `AllowedOrigins` must not be empty

### Rate Limiting (when enabled)
- Requests per second must be greater than 0
- Burst must be greater than 0

### Retry (when enabled)
- Max attempts must be at least 1
- Initial delay must be non-negative
- Max delay must be >= initial delay
- Backoff multiplier must be greater than 1.0
- Jitter percent must be between 0 and 100

## Environment Variables Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `SERVER_PORT` | int | 8080 | HTTP server port |
| `WEB_DIR` | string | "./web" | Static files directory |
| `SESSION_TIMEOUT` | duration | 30m | Session expiry time |
| `LOG_LEVEL` | string | "info" | Log level |
| `ALLOWED_ORIGINS` | string | "" | Comma-separated origins |
| `MAX_REQUEST_SIZE` | int64 | 1048576 | Max request bytes |
| `ENABLE_DEV_MODE` | bool | true | Development mode |
| `REQUEST_TIMEOUT` | duration | 30s | Request timeout |
| `ENABLE_PROFILING` | bool | false | Enable pprof |
| `PROFILING_PORT` | int | 0 | Profiling port |
| `METRICS_INTERVAL` | duration | 30s | Metrics interval |
| `ALERTING_ENABLED` | bool | true | Enable alerting |
| `ALERTING_INTERVAL` | duration | 30s | Alert check interval |
| `RATE_LIMIT_ENABLED` | bool | false | Enable rate limiting |
| `RATE_LIMIT_REQUESTS_PER_SECOND` | float64 | 5 | Requests/sec/IP |
| `RATE_LIMIT_BURST` | int | 10 | Max burst requests |
| `RATE_LIMIT_CLEANUP_INTERVAL` | duration | 1m | Rate limiter cleanup |
| `RETRY_ENABLED` | bool | true | Enable retry logic |
| `RETRY_MAX_ATTEMPTS` | int | 3 | Max retry attempts |
| `RETRY_INITIAL_DELAY` | duration | 100ms | Initial retry delay |
| `RETRY_MAX_DELAY` | duration | 30s | Max retry delay |
| `RETRY_BACKOFF_MULTIPLIER` | float64 | 2.0 | Backoff multiplier |
| `RETRY_JITTER_PERCENT` | int | 10 | Jitter percentage |
| `DATA_DIR` | string | "./data" | Data directory |
| `AUTO_SAVE_INTERVAL` | duration | 30s | Auto-save interval |
| `ENABLE_PERSISTENCE` | bool | true | Enable persistence |

## Production Configuration Example

```bash
# Production environment variables
export ENABLE_DEV_MODE=false
export ALLOWED_ORIGINS="https://yourdomain.com,https://www.yourdomain.com"
export LOG_LEVEL=warn
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_REQUESTS_PER_SECOND=10
export RATE_LIMIT_BURST=20
export DATA_DIR=/var/lib/goldbox/data
```

## Testing

```go
func TestConfigurationLoading(t *testing.T) {
    // Test environment variable override
    os.Setenv("SERVER_PORT", "9000")
    defer os.Unsetenv("SERVER_PORT")
    
    cfg, err := config.Load()
    assert.NoError(t, err)
    assert.Equal(t, 9000, cfg.ServerPort)
}

func TestConfigurationValidation(t *testing.T) {
    // Invalid port triggers validation error
    os.Setenv("SERVER_PORT", "70000")
    defer os.Unsetenv("SERVER_PORT")
    
    _, err := config.Load()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "server port")
}
```

## Dependencies

- `os`: Environment variable access
- `strconv`: Type conversion for environment variables
- `time`: Duration parsing and handling
- `github.com/sirupsen/logrus`: Structured logging

Last Updated: 2026-02-19
