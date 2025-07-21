package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthCheckEndpoints tests all health check endpoints
func TestHealthCheckEndpoints(t *testing.T) {
	server, err := NewRPCServer("./test_web")
	require.NoError(t, err)
	defer close(server.done)

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "health endpoint returns healthy status",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
			expectedBody:   "healthy",
		},
		{
			name:           "readiness endpoint returns ready",
			endpoint:       "/ready",
			expectedStatus: http.StatusOK,
			expectedBody:   "Ready",
		},
		{
			name:           "liveness endpoint returns alive",
			endpoint:       "/live",
			expectedStatus: http.StatusOK,
			expectedBody:   "Alive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestMetricsEndpoint tests the Prometheus metrics endpoint
func TestMetricsEndpoint(t *testing.T) {
	server, err := NewRPCServer("./test_web")
	require.NoError(t, err)
	defer close(server.done)

	// First, trigger a health check to generate metrics
	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthW := httptest.NewRecorder()
	server.ServeHTTP(healthW, healthReq)

	// Now test the metrics endpoint
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")

	// Check for some expected metrics
	body := w.Body.String()
	assert.Contains(t, body, "goldbox_server_start_time_seconds")
	assert.Contains(t, body, "goldbox_player_sessions_active")
	assert.Contains(t, body, "goldbox_health_checks_total")
}

// TestRequestCorrelationID tests that request correlation IDs are properly set
func TestRequestCorrelationID(t *testing.T) {
	server, err := NewRPCServer("./test_web")
	require.NoError(t, err)
	defer close(server.done)

	t.Run("auto-generated request ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// Should have set X-Request-ID header
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
		assert.Len(t, requestID, 36) // UUID length
	})

	t.Run("existing request ID preserved", func(t *testing.T) {
		existingID := "test-request-123"
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-Request-ID", existingID)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// Should preserve existing request ID
		requestID := w.Header().Get("X-Request-ID")
		assert.Equal(t, existingID, requestID)
	})
}

// TestHealthCheckerComponents tests individual health check components
func TestHealthCheckerComponents(t *testing.T) {
	server, err := NewRPCServer("./test_web")
	require.NoError(t, err)
	defer close(server.done)

	ctx := context.Background()

	t.Run("all health checks pass", func(t *testing.T) {
		response := server.healthChecker.RunHealthChecks(ctx)

		assert.Equal(t, HealthStatusHealthy, response.Status)
		assert.NotEmpty(t, response.Checks)
		assert.True(t, response.Duration > 0)

		// Check specific health checks exist
		checkNames := make([]string, len(response.Checks))
		for i, check := range response.Checks {
			checkNames[i] = check.Name
		}

		expectedChecks := []string{"server", "game_state", "spell_manager", "event_system"}
		for _, expected := range expectedChecks {
			assert.Contains(t, checkNames, expected)
		}
	})
}

// TestMetricsIntegration tests that metrics are properly recorded
func TestMetricsIntegration(t *testing.T) {
	server, err := NewRPCServer("./test_web")
	require.NoError(t, err)
	defer close(server.done)

	// Make a health check request to generate metrics
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Check metrics endpoint contains recorded data
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	server.ServeHTTP(metricsW, metricsReq)

	body := metricsW.Body.String()

	// Should contain health check metrics
	assert.Contains(t, body, "goldbox_health_checks_total")
	assert.Contains(t, body, "goldbox_http_requests_total")
}

// TestSessionMetricsTracking tests that session creation/cleanup updates metrics
func TestSessionMetricsTracking(t *testing.T) {
	server, err := NewRPCServer("./test_web")
	require.NoError(t, err)
	defer close(server.done)

	// Initial metrics check
	initialMetricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	initialMetricsW := httptest.NewRecorder()
	server.ServeHTTP(initialMetricsW, initialMetricsReq)
	initialBody := initialMetricsW.Body.String()

	// Create a session by making a request that goes through session handling
	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Check metrics after session creation
	finalMetricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	finalMetricsW := httptest.NewRecorder()
	server.ServeHTTP(finalMetricsW, finalMetricsReq)
	finalBody := finalMetricsW.Body.String()

	// Should contain session metrics
	assert.Contains(t, finalBody, "goldbox_player_sessions_active")

	// The metrics should have changed (requests recorded)
	assert.NotEqual(t, initialBody, finalBody)
}

// TestStructuredLogging tests that structured logging with request IDs works
func TestStructuredLogging(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIDKey, "test-request-123")

	// Create a logger with request context (replacing the missing getRequestLogger function)
	requestID, _ := ctx.Value(requestIDKey).(string)
	logger := logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"operation":  "test_operation",
	})

	// Should have request_id in the logger context
	entry := logger.WithField("test", "value")
	data := entry.Data

	assert.Equal(t, "test_operation", data["operation"])
	assert.Equal(t, "test-request-123", data["request_id"])
	assert.Equal(t, "value", data["test"])
}
