// Package resilience provides graceful degradation strategies for non-critical dependencies
package resilience

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DegradationLevel represents the current operational state
type DegradationLevel int

const (
	// LevelFull - all systems operational
	LevelFull DegradationLevel = iota
	// LevelDegraded - some non-critical systems unavailable, core functionality works
	LevelDegraded
	// LevelMinimal - only critical systems operational
	LevelMinimal
)

var degradationLevelNames = [...]string{
	LevelFull:     "Full",
	LevelDegraded: "Degraded",
	LevelMinimal:  "Minimal",
}

// String returns the human-readable name of the degradation level.
func (d DegradationLevel) String() string {
	if d >= 0 && int(d) < len(degradationLevelNames) {
		return degradationLevelNames[d]
	}
	return "Unknown"
}

// SubsystemStatus represents the health of an individual subsystem
type SubsystemStatus struct {
	Name      string
	Healthy   bool
	Critical  bool // If true, failure affects overall service availability
	Error     error
	LastCheck time.Time
}

// DegradationManager coordinates graceful degradation across subsystems
type DegradationManager struct {
	mu         sync.RWMutex
	subsystems map[string]*SubsystemStatus
	level      DegradationLevel
	logger     *logrus.Entry
}

// NewDegradationManager creates a new degradation manager
func NewDegradationManager() *DegradationManager {
	return &DegradationManager{
		subsystems: make(map[string]*SubsystemStatus),
		level:      LevelFull,
		logger:     logrus.WithField("component", "DegradationManager"),
	}
}

// RegisterSubsystem adds a subsystem to the degradation manager
func (dm *DegradationManager) RegisterSubsystem(name string, critical bool) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.subsystems[name] = &SubsystemStatus{
		Name:      name,
		Healthy:   true,
		Critical:  critical,
		LastCheck: time.Now(),
	}

	dm.logger.WithFields(logrus.Fields{
		"subsystem": name,
		"critical":  critical,
	}).Debug("Registered subsystem")
}

// UpdateSubsystemStatus updates the health status of a subsystem
func (dm *DegradationManager) UpdateSubsystemStatus(name string, healthy bool, err error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	status, exists := dm.subsystems[name]
	if !exists {
		dm.logger.WithField("subsystem", name).Warn("Attempted to update unknown subsystem")
		return
	}

	previousHealth := status.Healthy
	status.Healthy = healthy
	status.Error = err
	status.LastCheck = time.Now()

	if previousHealth != healthy {
		logLevel := logrus.InfoLevel
		if !healthy && status.Critical {
			logLevel = logrus.ErrorLevel
		} else if !healthy {
			logLevel = logrus.WarnLevel
		}

		dm.logger.WithFields(logrus.Fields{
			"subsystem": name,
			"healthy":   healthy,
			"critical":  status.Critical,
			"error":     err,
		}).Log(logLevel, "Subsystem status changed")
	}

	dm.recalculateDegradationLevel()
}

// recalculateDegradationLevel determines the overall degradation level
func (dm *DegradationManager) recalculateDegradationLevel() {
	criticalFailures := 0
	nonCriticalFailures := 0

	for _, status := range dm.subsystems {
		if !status.Healthy {
			if status.Critical {
				criticalFailures++
			} else {
				nonCriticalFailures++
			}
		}
	}

	previousLevel := dm.level

	if criticalFailures > 0 {
		dm.level = LevelMinimal
	} else if nonCriticalFailures > 0 {
		dm.level = LevelDegraded
	} else {
		dm.level = LevelFull
	}

	if previousLevel != dm.level {
		dm.logger.WithFields(logrus.Fields{
			"previous_level":        previousLevel.String(),
			"new_level":             dm.level.String(),
			"critical_failures":     criticalFailures,
			"non_critical_failures": nonCriticalFailures,
		}).Warn("Degradation level changed")
	}
}

// GetDegradationLevel returns the current degradation level
func (dm *DegradationManager) GetDegradationLevel() DegradationLevel {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.level
}

// GetSubsystemStatus returns the status of a specific subsystem
func (dm *DegradationManager) GetSubsystemStatus(name string) (*SubsystemStatus, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	status, exists := dm.subsystems[name]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external mutation
	statusCopy := *status
	return &statusCopy, true
}

// GetAllStatuses returns the status of all subsystems
func (dm *DegradationManager) GetAllStatuses() map[string]SubsystemStatus {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	statuses := make(map[string]SubsystemStatus, len(dm.subsystems))
	for name, status := range dm.subsystems {
		statuses[name] = *status
	}

	return statuses
}

// IsSubsystemHealthy checks if a subsystem is healthy
func (dm *DegradationManager) IsSubsystemHealthy(name string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	status, exists := dm.subsystems[name]
	return exists && status.Healthy
}

// Global degradation manager instance
var (
	globalDegradationManager *DegradationManager
	globalDegradationOnce    sync.Once
)

func initGlobalDegradationManager() {
	globalDegradationManager = NewDegradationManager()
}

// GetGlobalDegradationManager returns the global degradation manager instance
func GetGlobalDegradationManager() *DegradationManager {
	globalDegradationOnce.Do(initGlobalDegradationManager)
	return globalDegradationManager
}

// FallbackStrategy represents a fallback behavior for a subsystem
type FallbackStrategy func(ctx context.Context) (interface{}, error)

// ErrFallbackUsed indicates that a fallback strategy was employed
var ErrFallbackUsed = errors.New("fallback strategy used due to subsystem failure")

// ExecuteWithFallback executes an operation with a fallback strategy if the primary fails
func (dm *DegradationManager) ExecuteWithFallback(
	ctx context.Context,
	subsystemName string,
	primary func(context.Context) (interface{}, error),
	fallback FallbackStrategy,
) (interface{}, error) {
	// Try primary operation
	result, err := primary(ctx)
	if err == nil {
		dm.UpdateSubsystemStatus(subsystemName, true, nil)
		return result, nil
	}

	// Mark subsystem as unhealthy
	dm.UpdateSubsystemStatus(subsystemName, false, err)

	// Log the failure
	dm.logger.WithFields(logrus.Fields{
		"subsystem": subsystemName,
		"error":     err,
	}).Warn("Primary operation failed, using fallback")

	// Try fallback
	if fallback != nil {
		result, fallbackErr := fallback(ctx)
		if fallbackErr != nil {
			return nil, fallbackErr
		}
		return result, ErrFallbackUsed
	}

	return nil, err
}
