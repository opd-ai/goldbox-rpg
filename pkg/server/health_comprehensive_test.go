package server

import (
	"context"
	"testing"
	"time"
)

// Test_HealthChecker_Comprehensive_Coverage verifies that the health checker
// provides comprehensive coverage of all major subsystems (regression test for bug fix)
func Test_HealthChecker_Comprehensive_Coverage(t *testing.T) {
	// Create a minimal RPCServer for testing
	server := &RPCServer{}

	// Create health checker
	hc := NewHealthChecker(server)

	// Run health checks to see what's covered
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response := hc.RunHealthChecks(ctx)

	// Count the number of checks
	checkCount := len(response.Checks)

	// After bug fix, we should have at least 10 comprehensive checks
	if checkCount < 10 {
		t.Errorf("Expected at least 10 comprehensive health checks after bug fix, got %d", checkCount)
	}

	// Verify the specific comprehensive checks that should exist after bug fix
	expectedChecks := map[string]bool{
		"server":              false,
		"game_state":          false,
		"spell_manager":       false,
		"event_system":        false,
		"pcg_manager":         false,
		"validation_system":   false,
		"circuit_breakers":    false,
		"metrics_system":      false,
		"configuration":       false,
		"performance_monitor": false,
	}

	for _, check := range response.Checks {
		if _, exists := expectedChecks[check.Name]; exists {
			expectedChecks[check.Name] = true
		}
	}

	// All comprehensive checks should be present
	for checkName, found := range expectedChecks {
		if !found {
			t.Errorf("Expected comprehensive health check '%s' was not found", checkName)
		}
	}

	// Verify that comprehensive health checks are now implemented
	implementedChecks := []string{
		"pcg_manager",
		"validation_system",
		"circuit_breakers",
		"metrics_system",
		"configuration",
		"performance_monitor",
	}

	for _, expectedCheck := range implementedChecks {
		found := false
		for _, check := range response.Checks {
			if check.Name == expectedCheck {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Comprehensive health check '%s' should be implemented but was not found", expectedCheck)
		}
	}

	// This test verifies the bug is fixed: health checker now has comprehensive coverage
	t.Logf("Health checker has %d comprehensive checks, confirming bug fix", checkCount)
	t.Logf("Successfully implemented %d comprehensive subsystem checks", len(implementedChecks))
}
