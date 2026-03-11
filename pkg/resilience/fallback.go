// Package resilience provides fallback strategies for graceful degradation
package resilience

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// FallbackStrategies provides common fallback behaviors for subsystems
type FallbackStrategies struct {
	logger *logrus.Entry
}

// NewFallbackStrategies creates a new fallback strategies instance
func NewFallbackStrategies() *FallbackStrategies {
	return &FallbackStrategies{
		logger: logrus.WithField("component", "FallbackStrategies"),
	}
}

// MetricsNoOp returns a no-op fallback for metrics collection
// When metrics system fails, we continue serving requests without collecting metrics
func (fs *FallbackStrategies) MetricsNoOp(ctx context.Context) (interface{}, error) {
	fs.logger.Debug("Using metrics no-op fallback")
	return nil, nil
}

// PCGCached returns cached PCG content when generation fails
// This would integrate with a cache layer (placeholder for now)
func (fs *FallbackStrategies) PCGCached(ctx context.Context) (interface{}, error) {
	fs.logger.Warn("PCG generation failed, would use cached content if available")
	// In a real implementation, this would return cached terrain/items/quests
	return nil, fmt.Errorf("cache not implemented yet")
}

// ValidationLogOnly returns a fallback that logs validation failures but allows requests
// This is useful for non-critical validation that shouldn't block requests
func (fs *FallbackStrategies) ValidationLogOnly(ctx context.Context) (interface{}, error) {
	fs.logger.Warn("Validation system unavailable, using log-only mode")
	// In a real implementation, this would still accept the request but log the validation skip
	return true, nil // Return true to indicate request is allowed
}

// DefaultErrorResponse provides a generic error response when primary system fails
func (fs *FallbackStrategies) DefaultErrorResponse(message string) FallbackStrategy {
	return func(ctx context.Context) (interface{}, error) {
		fs.logger.WithField("message", message).Debug("Using default error response")
		return map[string]interface{}{
			"error":    message,
			"degraded": true,
		}, nil
	}
}

// CachedDataFallback provides cached data when primary data source fails
func (fs *FallbackStrategies) CachedDataFallback(cachedData interface{}) FallbackStrategy {
	return func(ctx context.Context) (interface{}, error) {
		fs.logger.Debug("Using cached data fallback")
		if cachedData == nil {
			return nil, fmt.Errorf("no cached data available")
		}
		return cachedData, nil
	}
}

// NoOpSuccess returns success without performing the operation
// Useful for non-critical operations like analytics, logging to external systems, etc.
func (fs *FallbackStrategies) NoOpSuccess(ctx context.Context) (interface{}, error) {
	fs.logger.Debug("Using no-op success fallback")
	return nil, nil
}

// Global fallback strategies instance
var (
	globalFallbackStrategies *FallbackStrategies
)

// GetGlobalFallbackStrategies returns the global fallback strategies instance
func GetGlobalFallbackStrategies() *FallbackStrategies {
	if globalFallbackStrategies == nil {
		globalFallbackStrategies = NewFallbackStrategies()
	}
	return globalFallbackStrategies
}
