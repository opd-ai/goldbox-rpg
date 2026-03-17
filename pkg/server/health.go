package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"goldbox-rpg/pkg/resilience"

	"github.com/sirupsen/logrus"
)

// buildVersion is set at build time via ldflags, or derived from build info
var buildVersion = getVersion()

func getVersion() string {
	// Try to get version from build info (set via ldflags or VCS)
	if info, ok := debug.ReadBuildInfo(); ok {
		// Check for ldflags-set version in settings
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:min(7, len(setting.Value))]
			}
		}
		// Fall back to module version
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "1.0.0"
}

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
	checks             map[string]func(context.Context) error
	server             *RPCServer
	degradationManager *resilience.DegradationManager
	criticalSubsystems map[string]bool // Tracks which subsystems are critical
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(server *RPCServer) *HealthChecker {
	hc := &HealthChecker{
		checks:             make(map[string]func(context.Context) error),
		server:             server,
		degradationManager: resilience.GetGlobalDegradationManager(),
		criticalSubsystems: make(map[string]bool),
	}

	// Register critical health checks (failures affect service availability)
	hc.RegisterCheckWithCriticality("server", hc.checkServer, true)
	hc.RegisterCheckWithCriticality("game_state", hc.checkGameState, true)
	hc.RegisterCheckWithCriticality("spell_manager", hc.checkSpellManager, true)
	hc.RegisterCheckWithCriticality("event_system", hc.checkEventSystem, true)

	// Register non-critical health checks (failures cause degradation but service continues)
	hc.RegisterCheckWithCriticality("pcg_manager", hc.checkPCGManager, false)
	hc.RegisterCheckWithCriticality("validation_system", hc.checkValidationSystem, false)
	hc.RegisterCheckWithCriticality("circuit_breakers", hc.checkCircuitBreakers, false)
	hc.RegisterCheckWithCriticality("metrics_system", hc.checkMetricsSystem, false)
	hc.RegisterCheckWithCriticality("configuration", hc.checkConfiguration, true)
	hc.RegisterCheckWithCriticality("performance_monitor", hc.checkPerformanceMonitor, false)

	return hc
}

// RegisterCheckWithCriticality adds a health check and registers it with degradation manager
func (hc *HealthChecker) RegisterCheckWithCriticality(name string, check func(context.Context) error, critical bool) {
	hc.checks[name] = check
	hc.criticalSubsystems[name] = critical
	hc.degradationManager.RegisterSubsystem(name, critical)
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
		Version:   buildVersion,
	}

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

			// Update degradation manager
			hc.degradationManager.UpdateSubsystemStatus(name, false, err)

			// Record failed health check in metrics (best effort - metrics may be degraded)
			if hc.server.metrics != nil {
				hc.server.metrics.RecordHealthCheck(name, "failure")
			}

			logrus.WithFields(logrus.Fields{
				"check":    name,
				"duration": result.Duration,
				"error":    err,
				"critical": hc.criticalSubsystems[name],
			}).Error("health check failed")
		} else {
			// Update degradation manager
			hc.degradationManager.UpdateSubsystemStatus(name, true, nil)

			// Record successful health check in metrics (best effort)
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

	// Set overall status based on degradation level
	degradationLevel := hc.degradationManager.GetDegradationLevel()
	switch degradationLevel {
	case resilience.LevelFull:
		response.Status = HealthStatusHealthy
	case resilience.LevelDegraded:
		response.Status = HealthStatusDegraded
	case resilience.LevelMinimal:
		response.Status = HealthStatusUnhealthy
	default:
		response.Status = HealthStatusUnhealthy
	}

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

// checkServer verifies the RPC server instance is running and accepting connections.
func (hc *HealthChecker) checkServer(ctx context.Context) error {
	if hc.server == nil {
		return NewHealthCheckError("server", "instance", ErrServerNil)
	}

	// Check if server is accepting connections
	select {
	case <-hc.server.done:
		return NewHealthCheckError("server", "shutdown", ErrServerShuttingDown)
	default:
		// Server is running
	}

	return nil
}

// checkGameState verifies the game state is initialized and the world is loaded.
func (hc *HealthChecker) checkGameState(ctx context.Context) error {
	if hc.server == nil || hc.server.state == nil {
		return NewHealthCheckError("game_state", "initialization", ErrGameStateNil)
	}

	// Try to acquire a read lock to ensure state is accessible
	hc.server.mu.RLock()
	defer hc.server.mu.RUnlock()

	if hc.server.state.WorldState == nil {
		return NewHealthCheckError("world", "initialization", ErrWorldNil)
	}

	return nil
}

// checkSpellManager verifies the spell manager is initialized and spells are loaded.
func (hc *HealthChecker) checkSpellManager(ctx context.Context) error {
	if hc.server == nil || hc.server.spellManager == nil {
		return NewHealthCheckError("spell_manager", "initialization", ErrSpellManagerNil)
	}

	// Check if spells are loaded
	spellCount := hc.server.spellManager.GetSpellCount()
	if spellCount == 0 {
		return NewHealthCheckError("spell_manager", "spells_loaded", fmt.Errorf("no spells loaded"))
	}

	return nil
}

// checkEventSystem verifies the event system is initialized.
func (hc *HealthChecker) checkEventSystem(ctx context.Context) error {
	if hc.server == nil || hc.server.eventSys == nil {
		return NewHealthCheckError("event_system", "initialization", ErrEventSystemNil)
	}

	// Event system is functional if we can reach this point
	return nil
}

// checkPCGManager verifies the procedural content generation system is initialized
// and has a valid registry and metrics available.
func (hc *HealthChecker) checkPCGManager(ctx context.Context) error {
	if hc.server == nil || hc.server.pcgManager == nil {
		return NewHealthCheckError("pcg_manager", "initialization", ErrPCGManagerNil)
	}

	// Check if PCG manager has registry and generators
	registry := hc.server.pcgManager.GetRegistry()
	if registry == nil {
		return NewHealthCheckError("pcg_registry", "initialization", ErrPCGManagerNil)
	}

	// Check if metrics are available
	metrics := hc.server.pcgManager.GetMetrics()
	if metrics == nil {
		return NewHealthCheckError("pcg_metrics", "initialization", ErrPCGManagerNil)
	}

	// Get generation statistics to ensure the system is functional
	stats := hc.server.pcgManager.GetGenerationStatistics()
	if stats == nil {
		return NewHealthCheckError("pcg_stats", "retrieval", fmt.Errorf("unable to retrieve PCG statistics"))
	}

	return nil
}

// checkValidationSystem verifies the input validation system is initialized
// by running a simple test validation.
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

// checkCircuitBreakers verifies the circuit breaker manager is initialized
// and can retrieve statistics.
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

// checkMetricsSystem verifies the Prometheus metrics system is initialized.
func (hc *HealthChecker) checkMetricsSystem(ctx context.Context) error {
	if hc.server == nil || hc.server.metrics == nil {
		return fmt.Errorf("metrics system is not initialized")
	}

	// Metrics system is considered healthy if it exists
	// (It doesn't have validation methods, but the existence check is sufficient)
	return nil
}

// checkConfiguration verifies the server configuration is loaded and valid.
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

// checkPerformanceMonitor verifies the performance monitoring system is initialized.
func (hc *HealthChecker) checkPerformanceMonitor(ctx context.Context) error {
	if hc.server == nil || hc.server.perfMonitor == nil {
		return fmt.Errorf("performance monitor is not initialized")
	}

	// Performance monitor exists, system is healthy
	return nil
}
