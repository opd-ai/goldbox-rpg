package pcg

import (
	"sync"
	"time"
)

// GenerationMetrics tracks performance statistics
type GenerationMetrics struct {
	mu               sync.RWMutex
	GenerationCounts map[ContentType]int64         `json:"generation_counts"`
	AverageTimings   map[ContentType]time.Duration `json:"average_timings"`
	ErrorCounts      map[ContentType]int64         `json:"error_counts"`
	CacheHits        int64                         `json:"cache_hits"`
	CacheMisses      int64                         `json:"cache_misses"`
	TotalGenerations int64                         `json:"total_generations"`
}

// NewGenerationMetrics creates a new metrics tracker
func NewGenerationMetrics() *GenerationMetrics {
	return &GenerationMetrics{
		GenerationCounts: make(map[ContentType]int64),
		AverageTimings:   make(map[ContentType]time.Duration),
		ErrorCounts:      make(map[ContentType]int64),
	}
}

// RecordGeneration records a successful generation
func (gm *GenerationMetrics) RecordGeneration(contentType ContentType, duration time.Duration) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.GenerationCounts[contentType]++
	gm.TotalGenerations++

	// Update rolling average
	if current, exists := gm.AverageTimings[contentType]; exists {
		count := gm.GenerationCounts[contentType]
		gm.AverageTimings[contentType] = (current*time.Duration(count-1) + duration) / time.Duration(count)
	} else {
		gm.AverageTimings[contentType] = duration
	}
}

// RecordError records a generation error
func (gm *GenerationMetrics) RecordError(contentType ContentType) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.ErrorCounts[contentType]++
}

// RecordCacheHit records a cache hit
func (gm *GenerationMetrics) RecordCacheHit() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.CacheHits++
}

// RecordCacheMiss records a cache miss
func (gm *GenerationMetrics) RecordCacheMiss() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.CacheMisses++
}

// GetStats returns current performance statistics
func (gm *GenerationMetrics) GetStats() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return map[string]interface{}{
		"generation_counts": gm.GenerationCounts,
		"average_timings":   gm.AverageTimings,
		"error_counts":      gm.ErrorCounts,
		"cache_hits":        gm.CacheHits,
		"cache_misses":      gm.CacheMisses,
		"total_generations": gm.TotalGenerations,
	}
}

// GetGenerationCount returns the total generation count for a content type
func (gm *GenerationMetrics) GetGenerationCount(contentType ContentType) int64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return gm.GenerationCounts[contentType]
}

// GetAverageTiming returns the average generation time for a content type
func (gm *GenerationMetrics) GetAverageTiming(contentType ContentType) time.Duration {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return gm.AverageTimings[contentType]
}

// GetErrorCount returns the total error count for a content type
func (gm *GenerationMetrics) GetErrorCount(contentType ContentType) int64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return gm.ErrorCounts[contentType]
}

// GetCacheHitRatio returns the cache hit ratio as a percentage
func (gm *GenerationMetrics) GetCacheHitRatio() float64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	total := gm.CacheHits + gm.CacheMisses
	if total == 0 {
		return 0.0
	}

	return float64(gm.CacheHits) / float64(total) * 100.0
}

// Reset clears all metrics data
func (gm *GenerationMetrics) Reset() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.GenerationCounts = make(map[ContentType]int64)
	gm.AverageTimings = make(map[ContentType]time.Duration)
	gm.ErrorCounts = make(map[ContentType]int64)
	gm.CacheHits = 0
	gm.CacheMisses = 0
	gm.TotalGenerations = 0
}
