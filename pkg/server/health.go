package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// HealthStatus represents the overall health status of the server
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// CheckResult represents the result of a single health check
type CheckResult struct {
	Name     string        `json:"name"`
	Status   HealthStatus  `json:"status"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
	Details  interface{}   `json:"details,omitempty"`
}

// HealthResponse represents the complete health check response
type HealthResponse struct {
	Status    HealthStatus  `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Checks    []CheckResult `json:"checks"`
	Version   string        `json:"version,omitempty"`
}

// HealthChecker manages health checks for various system components
type HealthChecker struct {
	checks map[string]func(context.Context) error
	server *RPCServer
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(server *RPCServer) *HealthChecker {
	hc := &HealthChecker{
		checks: make(map[string]func(context.Context) error),
		server: server,
	}

	// Register default health checks
	hc.RegisterCheck("server", hc.checkServer)
	hc.RegisterCheck("game_state", hc.checkGameState)
	hc.RegisterCheck("spell_manager", hc.checkSpellManager)
	hc.RegisterCheck("event_system", hc.checkEventSystem)

	// Register comprehensive health checks
	hc.RegisterCheck("pcg_manager", hc.checkPCGManager)
	hc.RegisterCheck("validation_system", hc.checkValidationSystem)
	hc.RegisterCheck("circuit_breakers", hc.checkCircuitBreakers)
	hc.RegisterCheck("metrics_system", hc.checkMetricsSystem)
	hc.RegisterCheck("configuration", hc.checkConfiguration)
	hc.RegisterCheck("performance_monitor", hc.checkPerformanceMonitor)

	return hc
}

// RegisterCheck adds a new health check with the given name
func (hc *HealthChecker) RegisterCheck(name string, check func(context.Context) error) {
	hc.checks[name] = check
}

// RunHealthChecks executes all registered health checks and returns the results
func (hc *HealthChecker) RunHealthChecks(ctx context.Context) HealthResponse {
	start := time.Now()
	response := HealthResponse{
		Timestamp: start,
		Checks:    make([]CheckResult, 0, len(hc.checks)),
		Version:   "1.0.0", // TODO: Get from build info
	}

	overallStatus := HealthStatusHealthy

	for name, check := range hc.checks {
		checkStart := time.Now()
		result := CheckResult{
			Name:     name,
			Duration: 0,
			Status:   HealthStatusHealthy,
		}

		// Run the check with timeout
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := check(checkCtx)
		cancel()

		result.Duration = time.Since(checkStart)

		if err != nil {
			result.Status = HealthStatusUnhealthy
			result.Error = err.Error()
			overallStatus = HealthStatusUnhealthy

			// Record failed health check in metrics
			if hc.server.metrics != nil {
				hc.server.metrics.RecordHealthCheck(name, "failure")
			}

			logrus.WithFields(logrus.Fields{
				"check":    name,
				"duration": result.Duration,
				"error":    err,
			}).Error("health check failed")
		} else {
			// Record successful health check in metrics
			if hc.server.metrics != nil {
				hc.server.metrics.RecordHealthCheck(name, "success")
			}

			logrus.WithFields(logrus.Fields{
				"check":    name,
				"duration": result.Duration,
			}).Debug("health check passed")
		}

		response.Checks = append(response.Checks, result)
	}

	response.Status = overallStatus
	response.Duration = time.Since(start)

	return response
}

// HTTP handler for health checks
func (hc *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Add request correlation ID if available
	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
		ctx = context.WithValue(ctx, "request_id", reqID)
	}

	response := hc.RunHealthChecks(ctx)

	// Set appropriate HTTP status based on health
	var httpStatus int
	switch response.Status {
	case HealthStatusHealthy:
		httpStatus = http.StatusOK
	case HealthStatusDegraded:
		httpStatus = http.StatusOK // Still accepting traffic
	case HealthStatusUnhealthy:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.WithError(err).Error("failed to encode health response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Readiness handler for Kubernetes-style probes
func (hc *HealthChecker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response := hc.RunHealthChecks(ctx)

	// For readiness, we're more strict - any unhealthy check fails readiness
	if response.Status == HealthStatusUnhealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Ready"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// Liveness handler for basic server availability
func (hc *HealthChecker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	// Basic liveness check - just verify server is responding
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alive"))
}

// Default health check implementations

func (hc *HealthChecker) checkServer(ctx context.Context) error {
	if hc.server == nil {
		return fmt.Errorf("server instance is nil")
	}

	// Check if server is accepting connections
	select {
	case <-hc.server.done:
		return fmt.Errorf("server is shutting down")
	default:
		// Server is running
	}

	return nil
}

func (hc *HealthChecker) checkGameState(ctx context.Context) error {
	if hc.server == nil || hc.server.state == nil {
		return fmt.Errorf("game state is not initialized")
	}

	// Try to acquire a read lock to ensure state is accessible
	hc.server.mu.RLock()
	defer hc.server.mu.RUnlock()

	if hc.server.state.WorldState == nil {
		return fmt.Errorf("world state is nil")
	}

	return nil
}

func (hc *HealthChecker) checkSpellManager(ctx context.Context) error {
	if hc.server == nil || hc.server.spellManager == nil {
		return fmt.Errorf("spell manager is not initialized")
	}

	// Check if spells are loaded
	spellCount := hc.server.spellManager.GetSpellCount()
	if spellCount == 0 {
		return fmt.Errorf("no spells loaded")
	}

	return nil
}

func (hc *HealthChecker) checkEventSystem(ctx context.Context) error {
	if hc.server == nil || hc.server.eventSys == nil {
		return fmt.Errorf("event system is not initialized")
	}

	// Event system is functional if we can reach this point
	return nil
}

func (hc *HealthChecker) checkPCGManager(ctx context.Context) error {
	if hc.server == nil || hc.server.pcgManager == nil {
		return fmt.Errorf("PCG manager is not initialized")
	}

	// Check if PCG manager has registry and generators
	registry := hc.server.pcgManager.GetRegistry()
	if registry == nil {
		return fmt.Errorf("PCG registry is not initialized")
	}

	// Check if metrics are available
	metrics := hc.server.pcgManager.GetMetrics()
	if metrics == nil {
		return fmt.Errorf("PCG metrics are not initialized")
	}

	// Get generation statistics to ensure the system is functional
	stats := hc.server.pcgManager.GetGenerationStatistics()
	if stats == nil {
		return fmt.Errorf("unable to retrieve PCG statistics")
	}

	return nil
}

func (hc *HealthChecker) checkValidationSystem(ctx context.Context) error {
	if hc.server == nil || hc.server.validator == nil {
		return fmt.Errorf("validation system is not initialized")
	}

	// Simple check - just ensure the validator exists and is functional
	// We test with a simple request that should succeed
	err := hc.server.validator.ValidateRPCRequest("getSpells", map[string]interface{}{
		"session_id": "550e8400-e29b-41d4-a716-446655440000", // Valid UUID format
	}, 100)

	// getSpells should be a valid method with minimal validation requirements
	if err != nil {
		return fmt.Errorf("validation system test failed: %v", err)
	}

	return nil
}

func (hc *HealthChecker) checkCircuitBreakers(ctx context.Context) error {
	// Use the global circuit breaker manager
	cbManager := GetCircuitBreakerManager()
	if cbManager == nil {
		return fmt.Errorf("circuit breaker manager is not initialized")
	}

	// Get stats to ensure it's functional
	stats := cbManager.GetAllStats()
	if stats == nil {
		return fmt.Errorf("unable to retrieve circuit breaker statistics")
	}

	return nil
}

func (hc *HealthChecker) checkMetricsSystem(ctx context.Context) error {
	if hc.server == nil || hc.server.metrics == nil {
		return fmt.Errorf("metrics system is not initialized")
	}

	// Metrics system is considered healthy if it exists
	// (It doesn't have validation methods, but the existence check is sufficient)
	return nil
}

func (hc *HealthChecker) checkConfiguration(ctx context.Context) error {
	if hc.server == nil || hc.server.config == nil {
		return fmt.Errorf("configuration is not initialized")
	}

	// Check that basic configuration values are set
	if hc.server.config.ServerPort == 0 {
		return fmt.Errorf("server port not configured")
	}

	return nil
}

func (hc *HealthChecker) checkPerformanceMonitor(ctx context.Context) error {
	if hc.server == nil || hc.server.perfMonitor == nil {
		return fmt.Errorf("performance monitor is not initialized")
	}

	// Performance monitor exists, system is healthy
	return nil
}
