# PCG Performance Metrics System

This document describes the Performance Metrics and Monitoring system implemented for the GoldBox RPG Engine's Procedural Content Generation (PCG) module.

## Overview

The metrics system provides comprehensive tracking of PCG performance statistics, including generation timing, error rates, and cache performance. It is designed to be thread-safe and integrated seamlessly with the existing PCG architecture.

## Features

- **Thread-safe performance tracking** for all content types
- **Automatic timing measurement** integrated into PCG manager
- **Error counting** for failed generation attempts
- **Cache hit/miss ratio tracking** for optimization insights
- **Rolling average calculations** for timing statistics
- **Real-time statistics access** through simple APIs

## Usage

### Basic Usage

```go
// The metrics system is automatically included when creating a PCG manager
pcgManager := pcg.NewPCGManager(world, logger)

// Access metrics
metrics := pcgManager.GetMetrics()

// Get comprehensive statistics
stats := pcgManager.GetGenerationStatistics()
```

### Manual Metrics Recording

```go
// Record successful generation
metrics.RecordGeneration(pcg.ContentTypeTerrain, 150*time.Millisecond)

// Record generation error
metrics.RecordError(pcg.ContentTypeItems)

// Record cache activity
metrics.RecordCacheHit()
metrics.RecordCacheMiss()
```

### Accessing Statistics

```go
// Get all statistics
allStats := metrics.GetStats()

// Get specific metrics
terrainCount := metrics.GetGenerationCount(pcg.ContentTypeTerrain)
avgTiming := metrics.GetAverageTiming(pcg.ContentTypeTerrain)
errorCount := metrics.GetErrorCount(pcg.ContentTypeTerrain)
hitRatio := metrics.GetCacheHitRatio()
```

### Resetting Metrics

```go
// Clear all metrics data
pcgManager.ResetMetrics()
// or
metrics.Reset()
```

## Tracked Content Types

The system tracks metrics for all PCG content types:
- `ContentTypeTerrain` - Terrain generation
- `ContentTypeItems` - Item generation
- `ContentTypeLevels` - Level/dungeon generation
- `ContentTypeQuests` - Quest generation
- `ContentTypeNPCs` - NPC generation
- `ContentTypeEvents` - Event generation

## Automatic Integration

The metrics system is automatically integrated into PCG manager generation methods:

- **GenerateTerrainForLevel()** - Tracks timing and errors for terrain generation
- **GenerateItemsForLocation()** - Tracks timing and errors for item generation
- Additional generation methods can be easily instrumented

## Performance Considerations

- All metrics operations use read-write mutexes for optimal concurrent performance
- Rolling averages are calculated incrementally to avoid storing large datasets
- Memory footprint is minimal, storing only aggregated statistics

## Demo

Run the included demo to see the metrics system in action:

```bash
cd pkg/pcg/demo
go run metrics_demo.go
```

## Testing

The metrics system includes comprehensive unit tests covering:
- Basic functionality
- Concurrent access
- Edge cases
- Performance characteristics

Run tests with:
```bash
go test ./pkg/pcg -v
```

## Integration with Game Systems

The metrics are exposed through the PCG manager's statistics API, making them available to:
- Server endpoints for monitoring dashboards
- Game administration tools
- Performance optimization analysis
- Automated scaling decisions

## Future Enhancements

Potential future improvements include:
- Histogram-based timing distribution tracking
- Automatic performance alerts and thresholds
- Metrics export to external monitoring systems
- Detailed memory usage tracking
- Generation quality scoring integration
