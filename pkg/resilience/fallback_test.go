package resilience

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFallbackStrategies(t *testing.T) {
	fs := NewFallbackStrategies()
	assert.NotNil(t, fs)
	assert.NotNil(t, fs.logger)
}

func TestMetricsNoOp(t *testing.T) {
	fs := NewFallbackStrategies()
	ctx := context.Background()

	result, err := fs.MetricsNoOp(ctx)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestValidationLogOnly(t *testing.T) {
	fs := NewFallbackStrategies()
	ctx := context.Background()

	result, err := fs.ValidationLogOnly(ctx)

	assert.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestDefaultErrorResponse(t *testing.T) {
	fs := NewFallbackStrategies()
	ctx := context.Background()
	message := "Service temporarily unavailable"

	fallback := fs.DefaultErrorResponse(message)
	result, err := fallback(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, message, resultMap["error"])
	assert.Equal(t, true, resultMap["degraded"])
}

func TestCachedDataFallback_WithData(t *testing.T) {
	fs := NewFallbackStrategies()
	ctx := context.Background()
	cachedData := map[string]string{"key": "value"}

	fallback := fs.CachedDataFallback(cachedData)
	result, err := fallback(ctx)

	assert.NoError(t, err)
	assert.Equal(t, cachedData, result)
}

func TestCachedDataFallback_NoData(t *testing.T) {
	fs := NewFallbackStrategies()
	ctx := context.Background()

	fallback := fs.CachedDataFallback(nil)
	result, err := fallback(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no cached data available")
}

func TestNoOpSuccess(t *testing.T) {
	fs := NewFallbackStrategies()
	ctx := context.Background()

	result, err := fs.NoOpSuccess(ctx)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetGlobalFallbackStrategies(t *testing.T) {
	fs1 := GetGlobalFallbackStrategies()
	fs2 := GetGlobalFallbackStrategies()

	assert.NotNil(t, fs1)
	assert.NotNil(t, fs2)
	assert.Same(t, fs1, fs2, "should return same instance")
}

func TestPCGCached(t *testing.T) {
	fs := NewFallbackStrategies()
	ctx := context.Background()

	result, err := fs.PCGCached(ctx)

	// Currently returns error since cache not implemented
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cache not implemented")
}
