package server

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/sirupsen/logrus"
)

// ProfilingConfig holds configuration for profiling endpoints
type ProfilingConfig struct {
	Enabled bool
	Path    string
}

// ProfilingServer provides HTTP endpoints for CPU and memory profiling
type ProfilingServer struct {
	server *http.Server
	config ProfilingConfig
}

// NewProfilingServer creates a new profiling server instance
func NewProfilingServer(config ProfilingConfig) *ProfilingServer {
	mux := http.NewServeMux()

	// Add pprof endpoints for performance monitoring
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Add custom endpoints for specific profiling data
	mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	mux.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	mux.HandleFunc("/debug/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)

	return &ProfilingServer{
		server: &http.Server{
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		config: config,
	}
}

// StartProfiling starts the profiling server on the specified address
func (ps *ProfilingServer) StartProfiling(addr string) error {
	if !ps.config.Enabled {
		logrus.Info("Profiling is disabled")
		return nil
	}

	ps.server.Addr = addr

	logrus.WithFields(logrus.Fields{
		"address": addr,
		"path":    ps.config.Path,
	}).Info("Starting profiling server")

	return ps.server.ListenAndServe()
}

// Shutdown gracefully shuts down the profiling server
func (ps *ProfilingServer) Shutdown(ctx context.Context) error {
	if !ps.config.Enabled {
		return nil
	}

	logrus.Info("Shutting down profiling server")
	return ps.server.Shutdown(ctx)
}

// PerformanceMonitor provides periodic collection of performance metrics
type PerformanceMonitor struct {
	metrics  *Metrics
	interval time.Duration
	stopChan chan struct{}
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(metrics *Metrics, interval time.Duration) *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:  metrics,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start begins periodic collection of performance metrics
func (pm *PerformanceMonitor) Start() {
	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()

	logrus.WithField("interval", pm.interval).Info("Starting performance monitoring")

	for {
		select {
		case <-ticker.C:
			pm.collectMetrics()
		case <-pm.stopChan:
			logrus.Info("Stopping performance monitoring")
			return
		}
	}
}

// Stop stops the performance monitoring
func (pm *PerformanceMonitor) Stop() {
	close(pm.stopChan)
}

// collectMetrics collects all performance metrics
func (pm *PerformanceMonitor) collectMetrics() {
	// Update memory usage metrics
	pm.metrics.UpdateMemoryUsage()

	// Update goroutines count
	pm.metrics.UpdateGoroutinesCount()

	// Update heap objects count
	pm.metrics.UpdateHeapObjects()

	// Update stack usage
	pm.metrics.UpdateStackInUse()
}
