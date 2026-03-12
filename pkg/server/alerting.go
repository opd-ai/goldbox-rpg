package server

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// AlertThresholds defines configurable thresholds for performance alerts
type AlertThresholds struct {
	// Memory thresholds
	MaxHeapSizeMB      int64         `yaml:"max_heap_size_mb" default:"512"`
	MaxGoroutines      int           `yaml:"max_goroutines" default:"1000"`
	MaxGCPauseDuration time.Duration `yaml:"max_gc_pause_duration" default:"100ms"`

	// Performance thresholds
	MaxResponseTime time.Duration `yaml:"max_response_time" default:"5s"`
	MinMemoryFreeMB int64         `yaml:"min_memory_free_mb" default:"50"`

	// Health check intervals
	CheckInterval time.Duration `yaml:"check_interval" default:"30s"`
}

// AlertLevel represents the severity of an alert
type AlertLevel int

const (
	AlertLevelInfo AlertLevel = iota
	AlertLevelWarning
	AlertLevelCritical
)

// String returns the string representation of an alert level
func (al AlertLevel) String() string {
	switch al {
	case AlertLevelInfo:
		return "info"
	case AlertLevelWarning:
		return "warning"
	case AlertLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// Alert represents a performance alert
type Alert struct {
	Level     AlertLevel
	Message   string
	Metric    string
	Value     interface{}
	Threshold interface{}
	Timestamp time.Time
}

// AlertHandler defines how alerts should be handled
type AlertHandler interface {
	HandleAlert(alert Alert)
}

// LogAlertHandler logs alerts using logrus
type LogAlertHandler struct{}

// HandleAlert implements AlertHandler for logging
func (lah *LogAlertHandler) HandleAlert(alert Alert) {
	logger := logrus.WithFields(logrus.Fields{
		"level":     alert.Level.String(),
		"metric":    alert.Metric,
		"value":     alert.Value,
		"threshold": alert.Threshold,
		"timestamp": alert.Timestamp,
	})

	switch alert.Level {
	case AlertLevelInfo:
		logger.Info(alert.Message)
	case AlertLevelWarning:
		logger.Warn(alert.Message)
	case AlertLevelCritical:
		logger.Error(alert.Message)
	}
}

// PerformanceAlerter monitors system performance and triggers alerts.
// It periodically checks memory usage, goroutine count, and GC metrics against
// configured thresholds and reports violations through the AlertHandler.
//
// Fields:
//   - thresholds: Configurable alert thresholds for various metrics
//   - handler: AlertHandler implementation for processing alerts
//   - metrics: Server metrics for performance data
//   - stopChan: Channel for graceful shutdown
type PerformanceAlerter struct {
	thresholds AlertThresholds
	handler    AlertHandler
	metrics    *Metrics
	stopChan   chan struct{}
}

// NewPerformanceAlerter creates a new performance alerter with the specified
// thresholds, handler, and metrics source. It initializes the stop channel
// for graceful shutdown support.
//
// Parameters:
//   - thresholds: AlertThresholds defining warning and critical levels
//   - handler: AlertHandler to process generated alerts
//   - metrics: Server metrics instance for performance data
//
// Returns:
//   - *PerformanceAlerter: Configured alerter ready to start monitoring
func NewPerformanceAlerter(thresholds AlertThresholds, handler AlertHandler, metrics *Metrics) *PerformanceAlerter {
	return &PerformanceAlerter{
		thresholds: thresholds,
		handler:    handler,
		metrics:    metrics,
		stopChan:   make(chan struct{}),
	}
}

// Start begins the monitoring loop that periodically checks performance metrics.
// The loop runs until Stop() is called or the context is cancelled.
// It logs startup and shutdown events for observability.
//
// Parameters:
//   - ctx: Context for cancellation support
func (pa *PerformanceAlerter) Start(ctx context.Context) {
	ticker := time.NewTicker(pa.thresholds.CheckInterval)
	defer ticker.Stop()

	logrus.WithField("interval", pa.thresholds.CheckInterval).Info("Starting performance alerting")

	for {
		select {
		case <-ticker.C:
			pa.checkPerformance()
		case <-pa.stopChan:
			logrus.Info("Stopping performance alerting")
			return
		case <-ctx.Done():
			logrus.Info("Context cancelled, stopping performance alerting")
			return
		}
	}
}

// Stop signals the performance alerter to stop monitoring.
// It closes the stop channel which causes the Start() loop to exit.
func (pa *PerformanceAlerter) Stop() {
	close(pa.stopChan)
}

// checkPerformance performs all configured performance checks.
// It reads runtime memory stats and compares them against thresholds,
// generating alerts for any violations detected.
func (pa *PerformanceAlerter) checkPerformance() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Check heap size
	heapSizeMB := int64(memStats.HeapSys / 1024 / 1024)
	if heapSizeMB > pa.thresholds.MaxHeapSizeMB {
		pa.handler.HandleAlert(Alert{
			Level:     AlertLevelWarning,
			Message:   fmt.Sprintf("Heap size exceeds threshold: %dMB > %dMB", heapSizeMB, pa.thresholds.MaxHeapSizeMB),
			Metric:    "heap_size_mb",
			Value:     heapSizeMB,
			Threshold: pa.thresholds.MaxHeapSizeMB,
			Timestamp: time.Now(),
		})
	}

	// Check goroutines count
	goroutines := runtime.NumGoroutine()
	if goroutines > pa.thresholds.MaxGoroutines {
		pa.handler.HandleAlert(Alert{
			Level:     AlertLevelWarning,
			Message:   fmt.Sprintf("Goroutines count exceeds threshold: %d > %d", goroutines, pa.thresholds.MaxGoroutines),
			Metric:    "goroutines_count",
			Value:     goroutines,
			Threshold: pa.thresholds.MaxGoroutines,
			Timestamp: time.Now(),
		})
	}

	// Check GC pause time (using last pause)
	if memStats.NumGC > 0 {
		gcPause := time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256])
		if gcPause > pa.thresholds.MaxGCPauseDuration {
			pa.handler.HandleAlert(Alert{
				Level:     AlertLevelWarning,
				Message:   fmt.Sprintf("GC pause duration exceeds threshold: %v > %v", gcPause, pa.thresholds.MaxGCPauseDuration),
				Metric:    "gc_pause_duration",
				Value:     gcPause,
				Threshold: pa.thresholds.MaxGCPauseDuration,
				Timestamp: time.Now(),
			})
		}
	}

	// Check heap size (not free memory - heap grows as needed)
	// Monitor heap allocation to detect memory leaks, not "free" heap space
	heapAllocMB := int64(memStats.HeapAlloc / 1024 / 1024)

	// Alert if heap allocation exceeds threshold (potential memory leak)
	// Use MaxHeapSizeMB instead of MinMemoryFreeMB for this check
	if heapAllocMB > pa.thresholds.MaxHeapSizeMB {
		pa.handler.HandleAlert(Alert{
			Level:     AlertLevelCritical,
			Message:   fmt.Sprintf("Heap allocation exceeds threshold: %dMB > %dMB", heapAllocMB, pa.thresholds.MaxHeapSizeMB),
			Metric:    "heap_alloc_mb",
			Value:     heapAllocMB,
			Threshold: pa.thresholds.MaxHeapSizeMB,
			Timestamp: time.Now(),
		})
	}
}

// DefaultAlertThresholds returns reasonable default thresholds
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		MaxHeapSizeMB:      512,
		MaxGoroutines:      1000,
		MaxGCPauseDuration: 100 * time.Millisecond,
		MaxResponseTime:    5 * time.Second,
		MinMemoryFreeMB:    50,
		CheckInterval:      30 * time.Second,
	}
}
