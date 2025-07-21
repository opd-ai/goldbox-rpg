# Performance Monitoring & Profiling

The GoldBox RPG Engine includes comprehensive performance monitoring and profiling capabilities to help with production deployment and debugging.

## Features

### Memory Usage Monitoring
- **Heap allocation tracking**: Monitors current heap memory usage
- **Stack usage monitoring**: Tracks stack memory consumption  
- **Goroutine counting**: Monitors active goroutine count
- **Garbage collection metrics**: Tracks GC pause times and frequency

### CPU and Goroutine Profiling
The engine exposes standard Go pprof profiling endpoints when enabled:

- `/debug/pprof/` - Main profiling index
- `/debug/pprof/heap` - Heap memory profiling
- `/debug/pprof/goroutine` - Goroutine stack traces
- `/debug/pprof/profile` - CPU profiling (30-second sample)
- `/debug/pprof/block` - Blocking profiling
- `/debug/pprof/mutex` - Mutex contention profiling
- `/debug/pprof/allocs` - Memory allocation profiling
- `/debug/pprof/trace` - Execution trace (for `go tool trace`)

### Performance Metrics
All metrics are exposed in Prometheus format at `/metrics`:

```
# Memory metrics
goldbox_memory_usage_bytes{type="heap"}
goldbox_memory_usage_bytes{type="stack"}
goldbox_goroutines_active
goldbox_heap_objects_total
goldbox_stack_inuse_bytes

# GC metrics  
goldbox_gc_duration_seconds
```

### Alerting System
The engine includes configurable performance alerting with thresholds for:

- **Memory usage**: Alert when heap size exceeds threshold
- **Goroutine count**: Alert when goroutine count is excessive
- **GC pause time**: Alert when garbage collection pauses are too long
- **Available memory**: Alert when free memory is critically low

## Configuration

Configure performance monitoring via environment variables:

```bash
# Enable profiling endpoints (disabled by default for security)
ENABLE_PROFILING=true

# Performance metrics collection interval (default: 30s)
METRICS_INTERVAL=30s

# Enable performance alerting (default: true)
ALERTING_ENABLED=true

# Alerting check interval (default: 30s)  
ALERTING_INTERVAL=30s

# Development mode enables profiling automatically
ENABLE_DEV_MODE=true
```

## Usage Examples

### Viewing Memory Profile
```bash
# Get heap profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Get goroutine profile  
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### CPU Profiling
```bash
# Get 30-second CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile

# Get 60-second CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=60
```

### Metrics Collection
```bash
# View all metrics
curl http://localhost:8080/metrics

# View specific memory metrics
curl http://localhost:8080/metrics | grep goldbox_memory
```

### Execution Tracing
```bash
# Collect 5-second trace
curl http://localhost:8080/debug/pprof/trace?seconds=5 > trace.out

# Analyze trace
go tool trace trace.out
```

## Security Considerations

- **Profiling endpoints are disabled by default** in production for security
- Enable profiling only in development or secure production environments
- Profiling data can contain sensitive information about application internals
- Consider restricting access to profiling endpoints via firewall or reverse proxy
- Monitor profiling endpoint access in production logs

## Performance Impact

- **Metrics collection**: Minimal overhead (~1-2ms every 30s)
- **Memory profiling**: Low overhead when not actively profiling
- **CPU profiling**: ~5% overhead during active profiling
- **Goroutine profiling**: Minimal overhead
- **Execution tracing**: Higher overhead, use sparingly in production

## Troubleshooting

### High Memory Usage
1. Check heap profile: `go tool pprof http://localhost:8080/debug/pprof/heap`
2. Look for large allocations in the profile
3. Monitor `goldbox_memory_usage_bytes` metrics
4. Check for memory leaks in goroutines

### High CPU Usage  
1. Collect CPU profile: `go tool pprof http://localhost:8080/debug/pprof/profile`
2. Identify hot functions in the profile
3. Monitor request duration metrics
4. Check for infinite loops or inefficient algorithms

### Goroutine Leaks
1. Get goroutine profile: `go tool pprof http://localhost:8080/debug/pprof/goroutine`
2. Look for large numbers of similar goroutines
3. Monitor `goldbox_goroutines_active` metric over time
4. Check for missing goroutine cleanup

### Performance Alerts
Performance alerts are logged with structured fields:
```json
{
  "level": "warning",
  "metric": "heap_size_mb", 
  "value": 256,
  "threshold": 200,
  "timestamp": "2025-07-20T10:30:00Z",
  "message": "Heap size exceeds threshold: 256MB > 200MB"
}
```

Monitor application logs for these alerts to proactively address performance issues.
