package server

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfilingServer(t *testing.T) {
	tests := []struct {
		name     string
		config   ProfilingConfig
		endpoint string
		expected int
	}{
		{
			name: "profiling enabled - pprof index",
			config: ProfilingConfig{
				Enabled: true,
				Path:    "/debug/pprof",
			},
			endpoint: "/debug/pprof/",
			expected: http.StatusOK,
		},
		{
			name: "profiling enabled - heap profile",
			config: ProfilingConfig{
				Enabled: true,
				Path:    "/debug/pprof",
			},
			endpoint: "/debug/pprof/heap",
			expected: http.StatusOK,
		},
		{
			name: "profiling enabled - goroutine profile",
			config: ProfilingConfig{
				Enabled: true,
				Path:    "/debug/pprof",
			},
			endpoint: "/debug/pprof/goroutine",
			expected: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewProfilingServer(tt.config)
			require.NotNil(t, ps)

			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			recorder := httptest.NewRecorder()

			ps.server.Handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expected, recorder.Code)
		})
	}
}

func TestPerformanceMonitor(t *testing.T) {
	metrics := NewMetrics()
	monitor := NewPerformanceMonitor(metrics, 100*time.Millisecond)

	require.NotNil(t, monitor)
	assert.Equal(t, 100*time.Millisecond, monitor.interval)

	// Test that metrics collection works
	monitor.collectMetrics()

	// Verify that metrics were updated by checking if they're not nil
	// Since we can't easily verify exact values, we just ensure the collection doesn't panic
	assert.NotNil(t, monitor.metrics)
}

func TestPerformanceAlerter(t *testing.T) {
	tests := []struct {
		name        string
		thresholds  AlertThresholds
		expectAlert bool
	}{
		{
			name: "thresholds not exceeded",
			thresholds: AlertThresholds{
				MaxHeapSizeMB:      1024,            // High threshold
				MaxGoroutines:      10000,           // High threshold
				MaxGCPauseDuration: 1 * time.Second, // High threshold
				MinMemoryFreeMB:    1,               // Low threshold
				CheckInterval:      1 * time.Second,
			},
			expectAlert: false,
		},
		{
			name: "goroutines threshold exceeded",
			thresholds: AlertThresholds{
				MaxHeapSizeMB:      1024,
				MaxGoroutines:      1, // Very low threshold - should trigger
				MaxGCPauseDuration: 1 * time.Second,
				MinMemoryFreeMB:    1,
				CheckInterval:      1 * time.Second,
			},
			expectAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()

			// Create a test alert handler to capture alerts
			alertCaptured := false
			testHandler := &testAlertHandler{
				onAlert: func(alert Alert) {
					alertCaptured = true
				},
			}

			alerter := NewPerformanceAlerter(tt.thresholds, testHandler, metrics)
			require.NotNil(t, alerter)

			// Run a single performance check
			alerter.checkPerformance()

			assert.Equal(t, tt.expectAlert, alertCaptured)
		})
	}
}

func TestDefaultAlertThresholds(t *testing.T) {
	thresholds := DefaultAlertThresholds()

	assert.Greater(t, thresholds.MaxHeapSizeMB, int64(0))
	assert.Greater(t, thresholds.MaxGoroutines, 0)
	assert.Greater(t, thresholds.MaxGCPauseDuration, time.Duration(0))
	assert.Greater(t, thresholds.MaxResponseTime, time.Duration(0))
	assert.Greater(t, thresholds.MinMemoryFreeMB, int64(0))
	assert.Greater(t, thresholds.CheckInterval, time.Duration(0))
}

func TestLogAlertHandler(t *testing.T) {
	handler := &LogAlertHandler{}

	alert := Alert{
		Level:     AlertLevelWarning,
		Message:   "Test alert",
		Metric:    "test_metric",
		Value:     100,
		Threshold: 50,
		Timestamp: time.Now(),
	}

	// This should not panic
	assert.NotPanics(t, func() {
		handler.HandleAlert(alert)
	})
}

func TestMetricsUpdateMethods(t *testing.T) {
	metrics := NewMetrics()

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "UpdateMemoryUsage",
			fn:   func() { metrics.UpdateMemoryUsage() },
		},
		{
			name: "UpdateGoroutinesCount",
			fn:   func() { metrics.UpdateGoroutinesCount() },
		},
		{
			name: "UpdateHeapObjects",
			fn:   func() { metrics.UpdateHeapObjects() },
		},
		{
			name: "UpdateStackInUse",
			fn:   func() { metrics.UpdateStackInUse() },
		},
		{
			name: "RecordGCDuration",
			fn:   func() { metrics.RecordGCDuration(10 * time.Millisecond) },
		},
		{
			name: "UpdateCPUUsage",
			fn:   func() { metrics.UpdateCPUUsage(100 * time.Millisecond) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These should not panic
			assert.NotPanics(t, tt.fn)
		})
	}
}

func TestAlertLevelString(t *testing.T) {
	tests := []struct {
		level    AlertLevel
		expected string
	}{
		{AlertLevelInfo, "info"},
		{AlertLevelWarning, "warning"},
		{AlertLevelCritical, "critical"},
		{AlertLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

// testAlertHandler is a mock alert handler for testing
type testAlertHandler struct {
	onAlert func(Alert)
}

func (tah *testAlertHandler) HandleAlert(alert Alert) {
	if tah.onAlert != nil {
		tah.onAlert(alert)
	}
}

// Benchmark tests for performance monitoring
func BenchmarkMetricsUpdate(b *testing.B) {
	metrics := NewMetrics()

	b.Run("UpdateMemoryUsage", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metrics.UpdateMemoryUsage()
		}
	})

	b.Run("UpdateGoroutinesCount", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metrics.UpdateGoroutinesCount()
		}
	})
}

func BenchmarkPerformanceCheck(b *testing.B) {
	metrics := NewMetrics()
	alerter := NewPerformanceAlerter(DefaultAlertThresholds(), &LogAlertHandler{}, metrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alerter.checkPerformance()
	}
}

// Helper function to trigger GC for testing
func triggerGC() {
	runtime.GC()
	runtime.GC() // Call twice to ensure completion
}
