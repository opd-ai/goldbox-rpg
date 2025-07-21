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

// PerformanceAlerter monitors system performance and triggers alerts
type PerformanceAlerter struct {
	thresholds AlertThresholds
	handler    AlertHandler
	metrics    *Metrics
	stopChan   chan struct{}
}

// NewPerformanceAlerter creates a new performance alerter
func NewPerformanceAlerter(thresholds AlertThresholds, handler AlertHandler, metrics *Metrics) *PerformanceAlerter {
	return &PerformanceAlerter{
		thresholds: thresholds,
		handler:    handler,
		metrics:    metrics,
		stopChan:   make(chan struct{}),
	}
}

// Start begins monitoring and alerting
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

// Stop stops the performance alerter
func (pa *PerformanceAlerter) Stop() {
	close(pa.stopChan)
}

// checkPerformance performs all performance checks
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

	// Check available memory
	heapAllocMB := int64(memStats.HeapAlloc / 1024 / 1024)
	heapSysMB := int64(memStats.HeapSys / 1024 / 1024)
	freeMemoryMB := heapSysMB - heapAllocMB

	if freeMemoryMB < pa.thresholds.MinMemoryFreeMB {
		pa.handler.HandleAlert(Alert{
			Level:     AlertLevelCritical,
			Message:   fmt.Sprintf("Free memory below threshold: %dMB < %dMB", freeMemoryMB, pa.thresholds.MinMemoryFreeMB),
			Metric:    "free_memory_mb",
			Value:     freeMemoryMB,
			Threshold: pa.thresholds.MinMemoryFreeMB,
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
