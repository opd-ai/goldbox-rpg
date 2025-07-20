package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Metrics holds all Prometheus metrics for the GoldBox RPG server
type Metrics struct {
	// HTTP and RPC metrics
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestSize     *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec

	// WebSocket metrics
	activeConnections prometheus.Gauge
	wsConnections     *prometheus.CounterVec
	wsMessages        *prometheus.CounterVec

	// Game-specific metrics
	activeSessions prometheus.Gauge
	playerActions  *prometheus.CounterVec
	gameEvents     *prometheus.CounterVec

	// System metrics
	serverStartTime prometheus.Gauge
	healthChecks    *prometheus.CounterVec

	// Registry for all metrics
	registry *prometheus.Registry
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "goldbox_http_requests_total",
				Help: "Total number of HTTP requests processed by method and status",
			},
			[]string{"method", "endpoint", "status"},
		),

		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "goldbox_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),

		requestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "goldbox_http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8), // 100B to 100MB
			},
			[]string{"method", "endpoint"},
		),

		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "goldbox_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8), // 100B to 100MB
			},
			[]string{"method", "endpoint"},
		),

		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "goldbox_websocket_connections_active",
				Help: "Number of active WebSocket connections",
			},
		),

		wsConnections: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "goldbox_websocket_connections_total",
				Help: "Total number of WebSocket connections by type",
			},
			[]string{"type"}, // "connected", "disconnected", "failed"
		),

		wsMessages: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "goldbox_websocket_messages_total",
				Help: "Total number of WebSocket messages by direction and type",
			},
			[]string{"direction", "type"}, // direction: "inbound"/"outbound", type: event type
		),

		activeSessions: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "goldbox_player_sessions_active",
				Help: "Number of active player sessions",
			},
		),

		playerActions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "goldbox_player_actions_total",
				Help: "Total number of player actions by type",
			},
			[]string{"action_type", "status"}, // status: "success", "error"
		),

		gameEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "goldbox_game_events_total",
				Help: "Total number of game events by type",
			},
			[]string{"event_type"},
		),

		serverStartTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "goldbox_server_start_time_seconds",
				Help: "Unix timestamp when the server started",
			},
		),

		healthChecks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "goldbox_health_checks_total",
				Help: "Total number of health checks by name and status",
			},
			[]string{"check_name", "status"}, // status: "success", "failure"
		),

		registry: registry,
	}

	// Register all metrics with the registry
	m.registry.MustRegister(
		m.requestCount,
		m.requestDuration,
		m.requestSize,
		m.responseSize,
		m.activeConnections,
		m.wsConnections,
		m.wsMessages,
		m.activeSessions,
		m.playerActions,
		m.gameEvents,
		m.serverStartTime,
		m.healthChecks,
	)

	// Set server start time
	m.serverStartTime.SetToCurrentTime()

	return m
}

// GetHandler returns an HTTP handler for exposing metrics
func (m *Metrics) GetHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		Registry:          m.registry,
	})
}

// RecordHTTPRequest records metrics for an HTTP request
func (m *Metrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	status := strconv.Itoa(statusCode)

	m.requestCount.WithLabelValues(method, endpoint, status).Inc()
	m.requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

	if requestSize > 0 {
		m.requestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	}
	if responseSize > 0 {
		m.responseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
	}
}

// RecordWebSocketConnection records WebSocket connection events
func (m *Metrics) RecordWebSocketConnection(connectionType string) {
	m.wsConnections.WithLabelValues(connectionType).Inc()

	if connectionType == "connected" {
		m.activeConnections.Inc()
	} else if connectionType == "disconnected" {
		m.activeConnections.Dec()
	}
}

// RecordWebSocketMessage records WebSocket message events
func (m *Metrics) RecordWebSocketMessage(direction, messageType string) {
	m.wsMessages.WithLabelValues(direction, messageType).Inc()
}

// RecordPlayerAction records player action events
func (m *Metrics) RecordPlayerAction(actionType, status string) {
	m.playerActions.WithLabelValues(actionType, status).Inc()
}

// RecordGameEvent records game event occurrences
func (m *Metrics) RecordGameEvent(eventType string) {
	m.gameEvents.WithLabelValues(eventType).Inc()
}

// UpdateActiveSessions updates the active sessions gauge
func (m *Metrics) UpdateActiveSessions(count int) {
	m.activeSessions.Set(float64(count))
}

// RecordHealthCheck records health check results
func (m *Metrics) RecordHealthCheck(checkName, status string) {
	m.healthChecks.WithLabelValues(checkName, status).Inc()
}

// MetricsMiddleware provides HTTP middleware for recording request metrics
func (m *Metrics) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Capture response details
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Get request size
		var requestSize int64
		if r.ContentLength > 0 {
			requestSize = r.ContentLength
		}

		// Process request
		next.ServeHTTP(recorder, r)

		// Record metrics
		duration := time.Since(start)
		endpoint := sanitizeEndpoint(r.URL.Path)

		m.RecordHTTPRequest(
			r.Method,
			endpoint,
			recorder.statusCode,
			duration,
			requestSize,
			recorder.responseSize,
		)

		// Log request for debugging
		logrus.WithFields(logrus.Fields{
			"method":        r.Method,
			"endpoint":      endpoint,
			"status":        recorder.statusCode,
			"duration_ms":   duration.Milliseconds(),
			"request_size":  requestSize,
			"response_size": recorder.responseSize,
			"user_agent":    r.UserAgent(),
		}).Debug("HTTP request processed")
	})
}

// responseRecorder wraps http.ResponseWriter to capture response details
type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	size, err := r.ResponseWriter.Write(data)
	r.responseSize += int64(size)
	return size, err
}

// sanitizeEndpoint normalizes endpoint paths for metrics
func sanitizeEndpoint(path string) string {
	// Common endpoint patterns for the goldbox server
	switch path {
	case "/":
		return "root"
	case "/health":
		return "health"
	case "/ready":
		return "ready"
	case "/live":
		return "live"
	case "/metrics":
		return "metrics"
	case "/rpc":
		return "rpc"
	case "/ws":
		return "websocket"
	default:
		// For static files and other endpoints
		if len(path) > 20 {
			return "other"
		}
		return path
	}
}
