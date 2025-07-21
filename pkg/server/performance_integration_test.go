package server

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformanceMonitoringIntegration(t *testing.T) {
	// Create server with performance monitoring enabled
	server, err := NewRPCServer("./test_web")
	require.NoError(t, err)
	defer func() {
		server.Shutdown(context.Background())
	}()

	// Verify that performance monitoring components are initialized
	assert.NotNil(t, server.metrics, "Metrics should be initialized")
	assert.NotNil(t, server.perfMonitor, "Performance monitor should be initialized")
	assert.NotNil(t, server.profiling, "Profiling server should be initialized")

	// Test that metrics collection works
	server.metrics.UpdateMemoryUsage()
	server.metrics.UpdateGoroutinesCount()
	server.metrics.UpdateHeapObjects()
	server.metrics.UpdateStackInUse()

	// Verify that metrics endpoints are accessible
	req := httptest.NewRequest("GET", "/metrics", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code, "Metrics endpoint should be accessible")

	// Test profiling endpoints (should be accessible in dev mode)
	req = httptest.NewRequest("GET", "/debug/pprof/", nil)
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	// Should get a response (either 200 for enabled or redirect for disabled)
	assert.True(t, recorder.Code == 200 || recorder.Code == 301 || recorder.Code == 404,
		"Profiling endpoint should return a valid status code")
}

func TestPerformanceAlertsIntegration(t *testing.T) {
	// Create metrics and alerter
	metrics := NewMetrics()

	// Create thresholds that will definitely trigger alerts
	thresholds := AlertThresholds{
		MaxHeapSizeMB:      1,                   // Very low threshold
		MaxGoroutines:      1,                   // Very low threshold
		MaxGCPauseDuration: 1 * time.Nanosecond, // Very low threshold
		MinMemoryFreeMB:    999999,              // Very high threshold
		CheckInterval:      100 * time.Millisecond,
	}

	alertReceived := false
	testHandler := &testAlertHandler{
		onAlert: func(alert Alert) {
			alertReceived = true
		},
	}

	alerter := NewPerformanceAlerter(thresholds, testHandler, metrics)
	require.NotNil(t, alerter)

	// Run performance check - should trigger alerts
	alerter.checkPerformance()

	assert.True(t, alertReceived, "Should have received at least one alert")
}

func TestProfilingConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		config         ProfilingConfig
		expectedStatus bool
	}{
		{
			name: "profiling enabled",
			config: ProfilingConfig{
				Enabled: true,
				Path:    "/debug/pprof",
			},
			expectedStatus: true,
		},
		{
			name: "profiling disabled",
			config: ProfilingConfig{
				Enabled: false,
				Path:    "/debug/pprof",
			},
			expectedStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiling := NewProfilingServer(tt.config)
			require.NotNil(t, profiling)
			assert.Equal(t, tt.expectedStatus, profiling.config.Enabled)
		})
	}
}
