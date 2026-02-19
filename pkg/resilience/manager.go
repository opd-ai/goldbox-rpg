// Package resilience provides a circuit breaker manager for coordinating multiple
// circuit breakers across different external dependencies.
package resilience

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CircuitBreakerManager manages multiple circuit breakers for different external dependencies
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	logger   *logrus.Entry
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logrus.WithField("component", "CircuitBreakerManager"),
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one with the given configuration
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	// Use provided config or create default
	var cbConfig CircuitBreakerConfig
	if config != nil {
		cbConfig = *config
		cbConfig.Name = name // Ensure name matches
	} else {
		cbConfig = DefaultCircuitBreakerConfig(name)
	}

	cb := NewCircuitBreaker(cbConfig)
	cbm.breakers[name] = cb

	cbm.logger.WithField("circuit_breaker", name).Info("Created new circuit breaker")
	return cb
}

// Get retrieves an existing circuit breaker by name
func (cbm *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	cb, exists := cbm.breakers[name]
	return cb, exists
}

// Remove removes a circuit breaker from the manager
func (cbm *CircuitBreakerManager) Remove(name string) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	delete(cbm.breakers, name)
	cbm.logger.WithField("circuit_breaker", name).Info("Removed circuit breaker")
}

// GetAllStats returns statistics for all managed circuit breakers
func (cbm *CircuitBreakerManager) GetAllStats() map[string]interface{} {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	stats := make(map[string]interface{})
	for name, cb := range cbm.breakers {
		stats[name] = cb.GetStats()
	}

	return stats
}

// ResetAll resets all circuit breakers to closed state
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	for name, cb := range cbm.breakers {
		cb.Reset()
		cbm.logger.WithField("circuit_breaker", name).Info("Reset circuit breaker")
	}
}

// GetBreakerNames returns a list of all circuit breaker names
func (cbm *CircuitBreakerManager) GetBreakerNames() []string {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	names := make([]string, 0, len(cbm.breakers))
	for name := range cbm.breakers {
		names = append(names, name)
	}

	return names
}

// Predefined circuit breaker configurations for common dependencies
var (
	// FileSystemConfig provides circuit breaker configuration for file system operations
	FileSystemConfig = CircuitBreakerConfig{
		Name:        "filesystem",
		MaxFailures: 3,
		Timeout:     10 * time.Second,
		MaxRequests: 2,
	}

	// WebSocketConfig provides circuit breaker configuration for WebSocket operations
	WebSocketConfig = CircuitBreakerConfig{
		Name:        "websocket",
		MaxFailures: 5,
		Timeout:     30 * time.Second,
		MaxRequests: 3,
	}

	// ConfigLoaderConfig provides circuit breaker configuration for configuration loading
	ConfigLoaderConfig = CircuitBreakerConfig{
		Name:        "config_loader",
		MaxFailures: 2,
		Timeout:     15 * time.Second,
		MaxRequests: 1,
	}
)

// Global circuit breaker manager instance with thread-safe initialization
var (
	globalCircuitBreakerManager *CircuitBreakerManager
	globalManagerOnce           sync.Once
)

// initGlobalManager initializes the global circuit breaker manager.
// Called via sync.Once to ensure thread-safe singleton initialization.
func initGlobalManager() {
	globalCircuitBreakerManager = NewCircuitBreakerManager()
}

// GetGlobalCircuitBreakerManager returns the global circuit breaker manager instance.
// Uses sync.Once to guarantee thread-safe initialization even under concurrent access.
func GetGlobalCircuitBreakerManager() *CircuitBreakerManager {
	globalManagerOnce.Do(initGlobalManager)
	return globalCircuitBreakerManager
}

// Helper functions for common operations

// ExecuteWithFileSystemCircuitBreaker executes a function with file system circuit breaker protection
func ExecuteWithFileSystemCircuitBreaker(ctx context.Context, fn func(context.Context) error) error {
	cb := GetGlobalCircuitBreakerManager().GetOrCreate("filesystem", &FileSystemConfig)
	return cb.Execute(ctx, fn)
}

// ExecuteWithWebSocketCircuitBreaker executes a function with WebSocket circuit breaker protection
func ExecuteWithWebSocketCircuitBreaker(ctx context.Context, fn func(context.Context) error) error {
	cb := GetGlobalCircuitBreakerManager().GetOrCreate("websocket", &WebSocketConfig)
	return cb.Execute(ctx, fn)
}

// ExecuteWithConfigLoaderCircuitBreaker executes a function with config loader circuit breaker protection
func ExecuteWithConfigLoaderCircuitBreaker(ctx context.Context, fn func(context.Context) error) error {
	cb := GetGlobalCircuitBreakerManager().GetOrCreate("config_loader", &ConfigLoaderConfig)
	return cb.Execute(ctx, fn)
}
