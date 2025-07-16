package pcg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerationMetrics(t *testing.T) {
	metrics := NewGenerationMetrics()

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.GenerationCounts)
	assert.NotNil(t, metrics.AverageTimings)
	assert.NotNil(t, metrics.ErrorCounts)
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(0), metrics.CacheMisses)
	assert.Equal(t, int64(0), metrics.TotalGenerations)
}

func TestRecordGeneration(t *testing.T) {
	metrics := NewGenerationMetrics()

	// Test first terrain generation
	metrics.RecordGeneration(ContentTypeTerrain, 100*time.Millisecond)
	assert.Equal(t, int64(1), metrics.TotalGenerations)
	assert.Equal(t, int64(1), metrics.GetGenerationCount(ContentTypeTerrain))
	assert.Equal(t, 100*time.Millisecond, metrics.GetAverageTiming(ContentTypeTerrain))

	// Test items generation (different content type)
	metrics.RecordGeneration(ContentTypeItems, 50*time.Millisecond)
	assert.Equal(t, int64(2), metrics.TotalGenerations)
	assert.Equal(t, int64(1), metrics.GetGenerationCount(ContentTypeItems))
	assert.Equal(t, 50*time.Millisecond, metrics.GetAverageTiming(ContentTypeItems))

	// Test second terrain generation (same content type)
	metrics.RecordGeneration(ContentTypeTerrain, 200*time.Millisecond)
	assert.Equal(t, int64(3), metrics.TotalGenerations)
	assert.Equal(t, int64(2), metrics.GetGenerationCount(ContentTypeTerrain))
	// Average should be (100 + 200) / 2 = 150ms
	assert.Equal(t, 150*time.Millisecond, metrics.GetAverageTiming(ContentTypeTerrain))
}

func TestRecordGenerationAveraging(t *testing.T) {
	metrics := NewGenerationMetrics()

	// Record first generation
	metrics.RecordGeneration(ContentTypeTerrain, 100*time.Millisecond)
	assert.Equal(t, 100*time.Millisecond, metrics.GetAverageTiming(ContentTypeTerrain))

	// Record second generation
	metrics.RecordGeneration(ContentTypeTerrain, 200*time.Millisecond)
	assert.Equal(t, 150*time.Millisecond, metrics.GetAverageTiming(ContentTypeTerrain))

	// Record third generation
	metrics.RecordGeneration(ContentTypeTerrain, 300*time.Millisecond)
	assert.Equal(t, 200*time.Millisecond, metrics.GetAverageTiming(ContentTypeTerrain))

	assert.Equal(t, int64(3), metrics.GetGenerationCount(ContentTypeTerrain))
	assert.Equal(t, int64(3), metrics.TotalGenerations)
}

func TestRecordError(t *testing.T) {
	metrics := NewGenerationMetrics()

	metrics.RecordError(ContentTypeTerrain)
	assert.Equal(t, int64(1), metrics.GetErrorCount(ContentTypeTerrain))

	metrics.RecordError(ContentTypeTerrain)
	assert.Equal(t, int64(2), metrics.GetErrorCount(ContentTypeTerrain))

	metrics.RecordError(ContentTypeItems)
	assert.Equal(t, int64(1), metrics.GetErrorCount(ContentTypeItems))
	assert.Equal(t, int64(2), metrics.GetErrorCount(ContentTypeTerrain))
}

func TestCacheMetrics(t *testing.T) {
	metrics := NewGenerationMetrics()

	// Test initial state
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(0), metrics.CacheMisses)
	assert.Equal(t, 0.0, metrics.GetCacheHitRatio())

	// Record cache hits and misses
	metrics.RecordCacheHit()
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()

	assert.Equal(t, int64(2), metrics.CacheHits)
	assert.Equal(t, int64(1), metrics.CacheMisses)
	// Use assert.InDelta for floating point comparison with tolerance
	assert.InDelta(t, 66.66666666666667, metrics.GetCacheHitRatio(), 0.0001)

	// Add more misses
	metrics.RecordCacheMiss()
	assert.Equal(t, 50.0, metrics.GetCacheHitRatio())
}

func TestGetStats(t *testing.T) {
	metrics := NewGenerationMetrics()

	// Record some data
	metrics.RecordGeneration(ContentTypeTerrain, 100*time.Millisecond)
	metrics.RecordGeneration(ContentTypeItems, 50*time.Millisecond)
	metrics.RecordError(ContentTypeLevels)
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()

	stats := metrics.GetStats()

	require.Contains(t, stats, "generation_counts")
	require.Contains(t, stats, "average_timings")
	require.Contains(t, stats, "error_counts")
	require.Contains(t, stats, "cache_hits")
	require.Contains(t, stats, "cache_misses")
	require.Contains(t, stats, "total_generations")

	assert.Equal(t, int64(2), stats["total_generations"])
	assert.Equal(t, int64(1), stats["cache_hits"])
	assert.Equal(t, int64(1), stats["cache_misses"])

	generationCounts := stats["generation_counts"].(map[ContentType]int64)
	assert.Equal(t, int64(1), generationCounts[ContentTypeTerrain])
	assert.Equal(t, int64(1), generationCounts[ContentTypeItems])

	averageTimings := stats["average_timings"].(map[ContentType]time.Duration)
	assert.Equal(t, 100*time.Millisecond, averageTimings[ContentTypeTerrain])
	assert.Equal(t, 50*time.Millisecond, averageTimings[ContentTypeItems])

	errorCounts := stats["error_counts"].(map[ContentType]int64)
	assert.Equal(t, int64(1), errorCounts[ContentTypeLevels])
}

func TestReset(t *testing.T) {
	metrics := NewGenerationMetrics()

	// Record some data
	metrics.RecordGeneration(ContentTypeTerrain, 100*time.Millisecond)
	metrics.RecordError(ContentTypeItems)
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()

	// Verify data exists
	assert.Equal(t, int64(1), metrics.TotalGenerations)
	assert.Equal(t, int64(1), metrics.CacheHits)
	assert.Equal(t, int64(1), metrics.CacheMisses)

	// Reset metrics
	metrics.Reset()

	// Verify everything is cleared
	assert.Equal(t, int64(0), metrics.TotalGenerations)
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(0), metrics.CacheMisses)
	assert.Equal(t, int64(0), metrics.GetGenerationCount(ContentTypeTerrain))
	assert.Equal(t, int64(0), metrics.GetErrorCount(ContentTypeItems))
	assert.Equal(t, time.Duration(0), metrics.GetAverageTiming(ContentTypeTerrain))
}

func TestConcurrency(t *testing.T) {
	metrics := NewGenerationMetrics()
	done := make(chan bool)

	// Start multiple goroutines recording metrics
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				contentType := ContentType([]ContentType{
					ContentTypeTerrain,
					ContentTypeItems,
					ContentTypeLevels,
					ContentTypeQuests,
				}[j%4])

				metrics.RecordGeneration(contentType, time.Duration(j)*time.Millisecond)
				metrics.RecordError(contentType)
				metrics.RecordCacheHit()
				metrics.RecordCacheMiss()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify metrics were recorded (should be 10 * 100 = 1000 total)
	assert.Equal(t, int64(1000), metrics.TotalGenerations)
	assert.Equal(t, int64(1000), metrics.CacheHits)
	assert.Equal(t, int64(1000), metrics.CacheMisses)

	// Each content type should have 250 generations and errors
	for _, contentType := range []ContentType{
		ContentTypeTerrain,
		ContentTypeItems,
		ContentTypeLevels,
		ContentTypeQuests,
	} {
		assert.Equal(t, int64(250), metrics.GetGenerationCount(contentType))
		assert.Equal(t, int64(250), metrics.GetErrorCount(contentType))
	}

	// Cache hit ratio should be 50%
	assert.Equal(t, 50.0, metrics.GetCacheHitRatio())
}

func TestGettersWithEmptyMetrics(t *testing.T) {
	metrics := NewGenerationMetrics()

	assert.Equal(t, int64(0), metrics.GetGenerationCount(ContentTypeTerrain))
	assert.Equal(t, time.Duration(0), metrics.GetAverageTiming(ContentTypeTerrain))
	assert.Equal(t, int64(0), metrics.GetErrorCount(ContentTypeTerrain))
	assert.Equal(t, 0.0, metrics.GetCacheHitRatio())
}

func TestMultipleContentTypes(t *testing.T) {
	metrics := NewGenerationMetrics()

	contentTypes := []ContentType{
		ContentTypeTerrain,
		ContentTypeItems,
		ContentTypeLevels,
		ContentTypeQuests,
		ContentTypeNPCs,
		ContentTypeEvents,
	}

	expectedCounts := make(map[ContentType]int64)
	expectedDurations := make(map[ContentType]time.Duration)

	// Record different amounts for each content type
	for i, contentType := range contentTypes {
		count := int64(i + 1)
		duration := time.Duration((i+1)*10) * time.Millisecond

		for j := int64(0); j < count; j++ {
			metrics.RecordGeneration(contentType, duration)
		}

		expectedCounts[contentType] = count
		expectedDurations[contentType] = duration
	}

	// Verify each content type has the correct metrics
	for _, contentType := range contentTypes {
		assert.Equal(t, expectedCounts[contentType], metrics.GetGenerationCount(contentType))
		assert.Equal(t, expectedDurations[contentType], metrics.GetAverageTiming(contentType))
	}

	// Verify total
	expectedTotal := int64(1 + 2 + 3 + 4 + 5 + 6) // Sum of counts
	assert.Equal(t, expectedTotal, metrics.TotalGenerations)
}
