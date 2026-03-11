# Graceful Degradation Guide

## Overview

The GoldBox RPG Engine implements graceful degradation to ensure service availability even when non-critical subsystems fail. This guide explains how the degradation system works and how to use it.

## Degradation Levels

The system supports three degradation levels:

### 1. Full Operation (Level: `Full`)
- **Status**: All systems operational
- **Behavior**: Normal operation with all features enabled
- **HTTP Status**: 200 OK
- **Health Endpoint**: Returns `"status": "healthy"`

### 2. Degraded Operation (Level: `Degraded`)
- **Status**: One or more non-critical systems unavailable
- **Behavior**: Core functionality works, some features may be limited
- **HTTP Status**: 200 OK (still accepting traffic)
- **Health Endpoint**: Returns `"status": "degraded"`
- **Example**: Metrics collection disabled, PCG using cached content

### 3. Minimal Operation (Level: `Minimal`)
- **Status**: One or more critical systems unavailable
- **Behavior**: Service severely limited or unavailable
- **HTTP Status**: 503 Service Unavailable
- **Health Endpoint**: Returns `"status": "unhealthy"`
- **Example**: Database unavailable, game state corrupted

## Subsystem Classification

### Critical Subsystems
Failures in critical subsystems trigger `Minimal` degradation level:
- `server` - HTTP/RPC server infrastructure
- `game_state` - Core game state management
- `spell_manager` - Spell system
- `event_system` - Event processing
- `configuration` - System configuration

### Non-Critical Subsystems
Failures in non-critical subsystems trigger `Degraded` level:
- `pcg_manager` - Procedural content generation
- `validation_system` - Input validation
- `circuit_breakers` - Circuit breaker management
- `metrics_system` - Prometheus metrics
- `performance_monitor` - Performance monitoring

## Using the Degradation Manager

### Registering a Subsystem

```go
import "goldbox-rpg/pkg/resilience"

dm := resilience.GetGlobalDegradationManager()

// Register a non-critical subsystem
dm.RegisterSubsystem("analytics", false)

// Register a critical subsystem
dm.RegisterSubsystem("database", true)
```

### Updating Subsystem Status

```go
// Mark subsystem as unhealthy
err := doOperation()
if err != nil {
    dm.UpdateSubsystemStatus("analytics", false, err)
}

// Mark subsystem as healthy
dm.UpdateSubsystemStatus("analytics", true, nil)
```

### Checking Degradation Level

```go
level := dm.GetDegradationLevel()

switch level {
case resilience.LevelFull:
    // All systems operational
case resilience.LevelDegraded:
    // Some non-critical systems down
case resilience.LevelMinimal:
    // Critical systems down
}
```

### Using Fallback Strategies

```go
dm := resilience.GetGlobalDegradationManager()
fs := resilience.GetGlobalFallbackStrategies()

// Execute with fallback
result, err := dm.ExecuteWithFallback(
    ctx,
    "metrics",
    func(ctx context.Context) (interface{}, error) {
        // Primary operation
        return collectMetrics(ctx)
    },
    fs.MetricsNoOp, // Fallback if primary fails
)

if errors.Is(err, resilience.ErrFallbackUsed) {
    // Fallback was used, service continues in degraded mode
}
```

## Available Fallback Strategies

### Metrics No-Op
Continues serving requests without collecting metrics:
```go
fs.MetricsNoOp(ctx)
```

### Validation Log-Only
Logs validation failures but allows requests:
```go
fs.ValidationLogOnly(ctx)
```

### Cached Data
Returns cached data when primary source fails:
```go
fallback := fs.CachedDataFallback(cachedData)
```

### No-Op Success
Returns success without performing the operation:
```go
fs.NoOpSuccess(ctx)
```

### Default Error Response
Provides a generic error response:
```go
fallback := fs.DefaultErrorResponse("Service temporarily unavailable")
```

## Health Check Integration

The health check system automatically integrates with the degradation manager:

### Health Endpoint Response

```json
{
  "status": "degraded",
  "timestamp": "2026-03-11T10:00:00Z",
  "duration": "15ms",
  "checks": [
    {
      "name": "server",
      "status": "healthy",
      "duration": "2ms"
    },
    {
      "name": "metrics_system",
      "status": "unhealthy",
      "duration": "5ms",
      "error": "metrics collection failed"
    }
  ],
  "version": "1.0.0"
}
```

### Health Check Behavior

- **All checks pass**: Returns `healthy` status (200 OK)
- **Non-critical check fails**: Returns `degraded` status (200 OK, still accepting traffic)
- **Critical check fails**: Returns `unhealthy` status (503 Service Unavailable)

## Best Practices

### 1. Design for Degradation
Plan fallback behaviors for each subsystem:
```go
// Good: Graceful degradation
metrics, err := collectMetrics()
if err != nil {
    log.Warn("Metrics unavailable, continuing without metrics")
    // Continue serving requests
}

// Bad: Hard failure
metrics := collectMetrics() // panics if metrics unavailable
```

### 2. Update Status Proactively
Update subsystem status as soon as you detect issues:
```go
conn, err := connectToDatabase()
if err != nil {
    dm.UpdateSubsystemStatus("database", false, err)
    return err
}
dm.UpdateSubsystemStatus("database", true, nil)
```

### 3. Use Appropriate Fallbacks
Choose fallbacks based on the operation:
- **Metrics/Analytics**: No-op (skip the operation)
- **Content Generation**: Cached data (return pre-generated content)
- **Validation**: Log-only (log but allow requests)

### 4. Monitor Degradation Events
Log degradation level changes for monitoring:
```go
// Degradation manager automatically logs level changes
// Add additional monitoring:
if dm.GetDegradationLevel() != resilience.LevelFull {
    alerting.NotifyOps("Service degraded")
}
```

## Testing Degradation

### Unit Tests

```go
func TestGracefulDegradation(t *testing.T) {
    dm := resilience.NewDegradationManager()
    dm.RegisterSubsystem("test_service", false)
    
    // Simulate failure
    dm.UpdateSubsystemStatus("test_service", false, errors.New("test error"))
    
    assert.Equal(t, resilience.LevelDegraded, dm.GetDegradationLevel())
}
```

### Integration Tests

Test that the service continues operating when non-critical systems fail:
```go
// Disable metrics
disableMetrics()

// Service should still accept requests
resp := makeRequest("/api/game/move")
assert.Equal(t, 200, resp.StatusCode)

// Health should report degraded
health := makeRequest("/health")
assert.Contains(t, health.Body, "degraded")
```

## Operational Considerations

### Monitoring

Monitor these metrics:
- Degradation level changes
- Subsystem health status
- Fallback usage frequency
- Failed primary operations

### Alerting

Configure alerts for:
- **Warning**: Service enters degraded mode
- **Critical**: Service enters minimal mode
- **Recovery**: Service returns to full operation

### Recovery

When subsystems recover:
1. Degradation manager automatically detects health restoration
2. Health checks transition back to healthy status
3. Degradation level upgrades automatically
4. Normal operation resumes

## Example: PCG Degradation

```go
// In handleGenerateTerrain
result, err := dm.ExecuteWithFallback(
    ctx,
    "pcg_manager",
    func(ctx context.Context) (interface{}, error) {
        // Primary: Generate new terrain
        return s.pcgManager.GenerateTerrainForLevel(ctx, locationID, width, height, biome, difficulty)
    },
    func(ctx context.Context) (interface{}, error) {
        // Fallback: Return cached or default terrain
        return getDefaultTerrain(locationID), nil
    },
)

if errors.Is(err, resilience.ErrFallbackUsed) {
    logrus.Warn("Using cached terrain due to PCG failure")
}
```

## Troubleshooting

### Service Stuck in Degraded Mode

1. Check subsystem status: `GET /health`
2. Review logs for subsystem errors
3. Manually verify subsystem health
4. Restart failed subsystem if needed

### Frequent Degradation Flapping

1. Review circuit breaker thresholds
2. Increase subsystem timeouts
3. Add retry logic before marking unhealthy
4. Investigate root cause of intermittent failures

## Related Documentation

- [Error Handling Guide](./ERROR_HANDLING.md)
- [Circuit Breaker Documentation](../pkg/resilience/doc.go)
- [Health Check API](./README-RPC.md#health-endpoints)
