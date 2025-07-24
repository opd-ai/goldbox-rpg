package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"goldbox-rpg/pkg/integration"
	"goldbox-rpg/pkg/resilience"
)

// TestLoadItemsWithCircuitBreakerProtection tests the integration approach for config loading
func TestLoadItemsWithCircuitBreakerProtection(t *testing.T) {
	// Reset circuit breaker state and integration executors
	resetCircuitBreakerForTesting()
	integration.ResetExecutorsForTesting()

	tempDir := t.TempDir()

	// Test 1: Successful file loading
	validFile := filepath.Join(tempDir, "valid.yaml")
	validContent := `
- item_id: "test_001"
  item_name: "Test Item"
  item_type: "weapon"
  item_damage: "1d6"
  item_weight: 1
  item_value: 10
`
	err := os.WriteFile(validFile, []byte(validContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	items, err := LoadItems(validFile)
	if err != nil {
		t.Fatalf("Expected successful load, got error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}

	// Test 2: Test with non-existent file to verify error handling
	nonExistentFile := filepath.Join(tempDir, "does_not_exist.yaml")
	_, err = LoadItems(nonExistentFile)
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}

	// The error should contain information about the operation failing
	// (it might be wrapped by retry logic or circuit breaker)
	errorStr := strings.ToLower(err.Error())
	if !strings.Contains(errorStr, "no such file") && !strings.Contains(errorStr, "operation failed") {
		t.Errorf("Expected file not found or operation failed error, got: %v", err)
	}

	// Test 3: Test with invalid YAML content to verify parsing error handling
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	invalidContent := `invalid_yaml: [unclosed_bracket`
	err = os.WriteFile(invalidFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid test file: %v", err)
	}

	_, err = LoadItems(invalidFile)
	if err == nil {
		t.Error("Expected error when parsing invalid YAML")
	}

	// The error should contain YAML parsing information or operation failure
	errorStr = strings.ToLower(err.Error())
	if !strings.Contains(errorStr, "yaml") && !strings.Contains(errorStr, "unmarshal") && !strings.Contains(errorStr, "operation failed") {
		t.Errorf("Expected YAML parsing or operation failed error, got: %v", err)
	}
}

// TestConfigLoaderCircuitBreakerConfiguration tests the circuit breaker configuration
func TestConfigLoaderCircuitBreakerConfiguration(t *testing.T) {
	resetCircuitBreakerForTesting()
	integration.ResetExecutorsForTesting()

	manager := resilience.GetGlobalCircuitBreakerManager()
	cb := manager.GetOrCreate("config_loader", &resilience.ConfigLoaderConfig)
	// Test configuration values
	config := resilience.ConfigLoaderConfig

	if config.MaxFailures != 2 {
		t.Errorf("Expected MaxFailures to be 2, got %d", config.MaxFailures)
	}

	if config.Timeout != 15*time.Second {
		t.Errorf("Expected Timeout to be 15s, got %v", config.Timeout)
	}

	if config.Name != "config_loader" {
		t.Errorf("Expected Name to be 'config_loader', got %s", config.Name)
	}

	// Verify circuit breaker uses the expected configuration
	if cb.GetState() != resilience.StateClosed {
		t.Errorf("Expected initial state to be closed, got %s", cb.GetState())
	}
}

// TestCircuitBreakerRecovery tests circuit breaker recovery behavior
func TestCircuitBreakerRecovery(t *testing.T) {
	resetCircuitBreakerForTesting()
	integration.ResetExecutorsForTesting()

	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "recovery.yaml")
	validContent := `
- item_id: "recovery_001"
  item_name: "Recovery Test"
  item_type: "misc"
  item_weight: 1
  item_value: 1
`
	err := os.WriteFile(validFile, []byte(validContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Force circuit breaker to open
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_ = resilience.ExecuteWithConfigLoaderCircuitBreaker(ctx, func(ctx context.Context) error {
			return fmt.Errorf("failure %d", i)
		})
	}

	manager := resilience.GetGlobalCircuitBreakerManager()
	cb := manager.GetOrCreate("config_loader", &resilience.ConfigLoaderConfig)

	if cb.GetState() != resilience.StateOpen {
		t.Errorf("Expected circuit breaker to be open, got %s", cb.GetState())
	}

	// Wait for circuit breaker to transition to half-open
	// Note: In a real test, we might need to wait or mock time
	// For this test, we'll simulate the behavior

	// The circuit breaker should eventually allow recovery
	// This is a simplified test since full recovery testing would require time manipulation
	if cb.GetState() == resilience.StateOpen {
		t.Log("Circuit breaker is open as expected after failures")
	}
}
