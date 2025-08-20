# Configuration Package

This package provides comprehensive configuration management and environment variable handling for the GoldBox RPG Engine.

## Overview

The configuration package centralizes all configuration management for the engine, providing a unified interface for loading settings from environment variables, configuration files, and default values. It supports runtime configuration updates and validation.

## Features

- **Environment Variable Support**: Automatic loading from environment variables
- **Configuration Files**: YAML and JSON configuration file support
- **Default Values**: Sensible defaults for all configuration options
- **Validation**: Configuration validation with detailed error reporting
- **Hot Reload**: Runtime configuration updates (where safe)
- **Type Safety**: Strongly typed configuration structures
- **Documentation**: Self-documenting configuration options

## Configuration Structure

### Server Configuration

```go
type ServerConfig struct {
    Port               int           `yaml:"port" env:"GOLDBOX_PORT" default:"8080"`
    Host               string        `yaml:"host" env:"GOLDBOX_HOST" default:"localhost"`
    ReadTimeout        time.Duration `yaml:"read_timeout" env:"GOLDBOX_READ_TIMEOUT" default:"30s"`
    WriteTimeout       time.Duration `yaml:"write_timeout" env:"GOLDBOX_WRITE_TIMEOUT" default:"30s"`
    MaxRequestSize     int64         `yaml:"max_request_size" env:"GOLDBOX_MAX_REQUEST_SIZE" default:"1048576"`
    AllowedOrigins     []string      `yaml:"allowed_origins" env:"GOLDBOX_ALLOWED_ORIGINS"`
    EnableMetrics      bool          `yaml:"enable_metrics" env:"GOLDBOX_ENABLE_METRICS" default:"true"`
    MetricsPath        string        `yaml:"metrics_path" env:"GOLDBOX_METRICS_PATH" default:"/metrics"`
}
```

### Game Configuration

```go
type GameConfig struct {
    SessionTimeout     time.Duration `yaml:"session_timeout" env:"GOLDBOX_SESSION_TIMEOUT" default:"30m"`
    MaxPlayers         int           `yaml:"max_players" env:"GOLDBOX_MAX_PLAYERS" default:"100"`
    WorldUpdateRate    time.Duration `yaml:"world_update_rate" env:"GOLDBOX_WORLD_UPDATE_RATE" default:"100ms"`
    SaveInterval       time.Duration `yaml:"save_interval" env:"GOLDBOX_SAVE_INTERVAL" default:"5m"`
    EnablePCG          bool          `yaml:"enable_pcg" env:"GOLDBOX_ENABLE_PCG" default:"true"`
    PCGSeed            int64         `yaml:"pcg_seed" env:"GOLDBOX_PCG_SEED" default:"0"`
}
```

### Database Configuration

```go
type DatabaseConfig struct {
    Type            string        `yaml:"type" env:"GOLDBOX_DB_TYPE" default:"memory"`
    ConnectionString string       `yaml:"connection_string" env:"GOLDBOX_DB_CONNECTION"`
    MaxConnections   int          `yaml:"max_connections" env:"GOLDBOX_DB_MAX_CONNECTIONS" default:"10"`
    ConnTimeout      time.Duration `yaml:"connection_timeout" env:"GOLDBOX_DB_CONN_TIMEOUT" default:"10s"`
    QueryTimeout     time.Duration `yaml:"query_timeout" env:"GOLDBOX_DB_QUERY_TIMEOUT" default:"30s"`
}
```

### Logging Configuration

```go
type LoggingConfig struct {
    Level      string `yaml:"level" env:"GOLDBOX_LOG_LEVEL" default:"info"`
    Format     string `yaml:"format" env:"GOLDBOX_LOG_FORMAT" default:"json"`
    Output     string `yaml:"output" env:"GOLDBOX_LOG_OUTPUT" default:"stdout"`
    EnableFile bool   `yaml:"enable_file" env:"GOLDBOX_LOG_ENABLE_FILE" default:"false"`
    FilePath   string `yaml:"file_path" env:"GOLDBOX_LOG_FILE_PATH" default:"goldbox.log"`
}
```

## Usage

### Basic Configuration Loading

```go
import "goldbox-rpg/pkg/config"

// Load configuration with defaults
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load configuration:", err)
}

// Access configuration values
fmt.Printf("Server will run on port: %d\n", cfg.Server.Port)
fmt.Printf("Session timeout: %v\n", cfg.Game.SessionTimeout)
```

### Configuration from File

```go
// Load from specific configuration file
cfg, err := config.LoadFromFile("config/production.yaml")
if err != nil {
    log.Fatal("Failed to load config file:", err)
}

// Load with multiple sources (environment variables override file values)
cfg, err := config.LoadFromFileWithEnv("config/base.yaml")
if err != nil {
    log.Fatal("Failed to load configuration:", err)
}
```

### Environment Variable Configuration

```bash
# Set environment variables
export GOLDBOX_PORT=9000
export GOLDBOX_LOG_LEVEL=debug
export GOLDBOX_SESSION_TIMEOUT=45m
export GOLDBOX_ALLOWED_ORIGINS="https://game.example.com,https://admin.example.com"

# Run application (will use environment variables)
./bin/server
```

## Configuration Files

### YAML Configuration Example

```yaml
# config/production.yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"
  max_request_size: 2097152  # 2MB
  allowed_origins:
    - "https://yourdomain.com"
    - "https://www.yourdomain.com"
  enable_metrics: true
  metrics_path: "/metrics"

game:
  session_timeout: "30m"
  max_players: 500
  world_update_rate: "50ms"
  save_interval: "2m"
  enable_pcg: true
  pcg_seed: 12345

database:
  type: "postgresql"
  connection_string: "postgres://user:pass@localhost/goldbox?sslmode=require"
  max_connections: 25
  connection_timeout: "10s"
  query_timeout: "30s"

logging:
  level: "warn"
  format: "json"
  output: "file"
  enable_file: true
  file_path: "/var/log/goldbox/server.log"

resilience:
  circuit_breaker:
    failure_threshold: 5
    recovery_timeout: "30s"
    max_requests: 3
  
  retry:
    max_attempts: 3
    initial_delay: "100ms"
    max_delay: "5s"
    backoff_factor: 2.0

validation:
  max_request_size: 1048576  # 1MB
  enable_strict_mode: true
  custom_validators:
    - "character_creation"
    - "spell_casting"
    - "item_trading"
```

## Configuration Validation

### Built-in Validation

```go
// Validate configuration after loading
if err := cfg.Validate(); err != nil {
    log.Fatal("Configuration validation failed:", err)
}

// Example validation errors:
// - Port out of valid range (1-65535)
// - Invalid duration formats
// - Missing required database connection string
// - Invalid log levels
```

### Custom Validation

```go
// Register custom validator
config.RegisterValidator("game", func(cfg *config.GameConfig) error {
    if cfg.MaxPlayers < 1 {
        return errors.New("max_players must be greater than 0")
    }
    
    if cfg.SessionTimeout < time.Minute {
        return errors.New("session_timeout must be at least 1 minute")
    }
    
    return nil
})
```

## Runtime Configuration Updates

### Hot Reload Support

```go
// Watch for configuration changes
watcher, err := config.NewConfigWatcher("config/production.yaml")
if err != nil {
    log.Fatal("Failed to create config watcher:", err)
}

// Handle configuration updates
watcher.OnChange(func(newConfig *config.Config) {
    // Update safe configuration values
    server.UpdateLoggingLevel(newConfig.Logging.Level)
    server.UpdateSessionTimeout(newConfig.Game.SessionTimeout)
    
    log.Info("Configuration updated successfully")
})
```

### Safe vs Unsafe Updates

**Safe Updates (can be applied at runtime):**
- Logging levels and formats
- Session timeouts
- PCG settings
- Monitoring configurations

**Unsafe Updates (require restart):**
- Server port and host
- Database connection settings
- Core game mechanics settings

## Environment-Specific Configurations

### Development Configuration

```yaml
# config/development.yaml
server:
  port: 8080
  host: "localhost"
  allowed_origins: ["*"]  # Allow all origins for development

game:
  session_timeout: "5m"   # Shorter timeout for development
  enable_pcg: true
  pcg_seed: 0             # Random seed

logging:
  level: "debug"
  format: "text"          # Human-readable for development
  output: "stdout"

database:
  type: "memory"          # In-memory database for development
```

### Production Configuration

```yaml
# config/production.yaml
server:
  port: 8080
  host: "0.0.0.0"
  allowed_origins:
    - "https://yourdomain.com"
  max_request_size: 1048576

game:
  session_timeout: "30m"
  max_players: 1000
  save_interval: "1m"

logging:
  level: "warn"
  format: "json"
  enable_file: true
  file_path: "/var/log/goldbox/server.log"

database:
  type: "postgresql"
  connection_string: "${DATABASE_URL}"  # From environment
  max_connections: 50
```

## Integration with Other Packages

### Server Integration

```go
// In server initialization
func NewServer(configPath string) (*Server, error) {
    cfg, err := config.LoadFromFile(configPath)
    if err != nil {
        return nil, err
    }
    
    server := &Server{
        config: cfg,
        port:   cfg.Server.Port,
        host:   cfg.Server.Host,
    }
    
    // Configure logging
    if err := server.configureLogging(cfg.Logging); err != nil {
        return nil, err
    }
    
    return server, nil
}
```

### Game Engine Integration

```go
// Configure game engine with loaded configuration
func (s *Server) initializeGameEngine() error {
    gameEngine := game.NewEngine(game.Config{
        SessionTimeout:  s.config.Game.SessionTimeout,
        MaxPlayers:     s.config.Game.MaxPlayers,
        EnablePCG:      s.config.Game.EnablePCG,
        PCGSeed:        s.config.Game.PCGSeed,
    })
    
    s.gameEngine = gameEngine
    return nil
}
```

## Testing

### Configuration Testing

```go
func TestConfigurationLoading(t *testing.T) {
    // Test default configuration
    cfg := config.NewDefault()
    assert.Equal(t, 8080, cfg.Server.Port)
    assert.Equal(t, 30*time.Minute, cfg.Game.SessionTimeout)
    
    // Test environment variable override
    os.Setenv("GOLDBOX_PORT", "9000")
    defer os.Unsetenv("GOLDBOX_PORT")
    
    cfg, err := config.Load()
    assert.NoError(t, err)
    assert.Equal(t, 9000, cfg.Server.Port)
}
```

### Configuration Validation Testing

```go
func TestConfigurationValidation(t *testing.T) {
    cfg := config.NewDefault()
    
    // Test valid configuration
    err := cfg.Validate()
    assert.NoError(t, err)
    
    // Test invalid configuration
    cfg.Server.Port = -1
    err = cfg.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "port")
}
```

## Dependencies

- `gopkg.in/yaml.v3`: YAML configuration file support
- `encoding/json`: JSON configuration file support
- `os`: Environment variable access
- `time`: Duration parsing and handling
- `github.com/sirupsen/logrus`: Logging configuration

## Best Practices

1. **Use Environment Variables**: For deployment-specific settings
2. **Configuration Files**: For complex, structured configuration
3. **Validation**: Always validate configuration after loading
4. **Documentation**: Document all configuration options
5. **Defaults**: Provide sensible defaults for all options
6. **Security**: Never commit sensitive configuration to version control

Last Updated: 2025-08-20
