package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"goldbox-rpg/pkg/integration"
	"goldbox-rpg/pkg/resilience"
)

// resetCircuitBreakerForTesting resets the circuit breaker state for testing
func resetCircuitBreakerForTesting() {
	manager := resilience.GetGlobalCircuitBreakerManager()
	// Remove the existing config_loader circuit breaker to reset its state
	manager.Remove("config_loader")

	// Reset the integration executors to ensure clean state
	integration.ResetExecutorsForTesting()
}

// TestLoadItems_ValidYAMLFile tests successful loading of a valid YAML file
func TestLoadItems_ValidYAMLFile(t *testing.T) {
	resetCircuitBreakerForTesting()

	// Create a temporary directory for test files
	tempDir := t.TempDir()
	validYAMLFile := filepath.Join(tempDir, "valid_items.yaml")

	// Create valid YAML content
	validYAMLContent := `
- item_id: "sword_001"
  item_name: "Iron Sword"
  item_type: "weapon"
  item_damage: "1d8"
  item_weight: 3
  item_value: 50
  item_properties:
    - "sharp"
    - "metal"

- item_id: "armor_001"
  item_name: "Leather Armor"
  item_type: "armor"
  item_armor_class: 2
  item_weight: 10
  item_value: 100
`

	// Write test file
	err := os.WriteFile(validYAMLFile, []byte(validYAMLContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test LoadItems function
	items, err := LoadItems(validYAMLFile)
	if err != nil {
		t.Fatalf("LoadItems failed: %v", err)
	}

	// Verify the result
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Check first item (sword)
	sword := items[0]
	if sword.ID != "sword_001" {
		t.Errorf("Expected sword ID 'sword_001', got '%s'", sword.ID)
	}
	if sword.Name != "Iron Sword" {
		t.Errorf("Expected sword name 'Iron Sword', got '%s'", sword.Name)
	}
	if sword.Type != "weapon" {
		t.Errorf("Expected sword type 'weapon', got '%s'", sword.Type)
	}
	if sword.Damage != "1d8" {
		t.Errorf("Expected sword damage '1d8', got '%s'", sword.Damage)
	}
	if sword.Weight != 3 {
		t.Errorf("Expected sword weight 3, got %d", sword.Weight)
	}
	if sword.Value != 50 {
		t.Errorf("Expected sword value 50, got %d", sword.Value)
	}
	if len(sword.Properties) != 2 {
		t.Errorf("Expected 2 properties for sword, got %d", len(sword.Properties))
	}

	// Check second item (armor)
	armor := items[1]
	if armor.ID != "armor_001" {
		t.Errorf("Expected armor ID 'armor_001', got '%s'", armor.ID)
	}
	if armor.Name != "Leather Armor" {
		t.Errorf("Expected armor name 'Leather Armor', got '%s'", armor.Name)
	}
	if armor.Type != "armor" {
		t.Errorf("Expected armor type 'armor', got '%s'", armor.Type)
	}
	if armor.AC != 2 {
		t.Errorf("Expected armor AC 2, got %d", armor.AC)
	}
	if armor.Weight != 10 {
		t.Errorf("Expected armor weight 10, got %d", armor.Weight)
	}
	if armor.Value != 100 {
		t.Errorf("Expected armor value 100, got %d", armor.Value)
	}
}

// TestLoadItems_EmptyYAMLFile tests loading an empty YAML file
func TestLoadItems_EmptyYAMLFile(t *testing.T) {
	resetCircuitBreakerForTesting()

	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.yaml")

	// Create empty file
	err := os.WriteFile(emptyFile, []byte(""), 0o644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}

	items, err := LoadItems(emptyFile)
	if err != nil {
		t.Fatalf("LoadItems failed on empty file: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items from empty file, got %d", len(items))
	}
}

// TestLoadItems_EmptyArrayYAML tests loading a YAML file with an empty array
func TestLoadItems_EmptyArrayYAML(t *testing.T) {
	resetCircuitBreakerForTesting()

	tempDir := t.TempDir()
	emptyArrayFile := filepath.Join(tempDir, "empty_array.yaml")

	// Create file with empty array
	emptyArrayContent := "[]"
	err := os.WriteFile(emptyArrayFile, []byte(emptyArrayContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create empty array test file: %v", err)
	}

	items, err := LoadItems(emptyArrayFile)
	if err != nil {
		t.Fatalf("LoadItems failed on empty array file: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items from empty array file, got %d", len(items))
	}
}

// TestLoadItems_FileNotFound tests error handling when file doesn't exist
func TestLoadItems_FileNotFound(t *testing.T) {
	resetCircuitBreakerForTesting()

	nonExistentFile := "this_file_does_not_exist.yaml"

	items, err := LoadItems(nonExistentFile)

	// Should return an error
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Should return nil items on error
	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}
}

// TestLoadItems_InvalidYAMLSyntax tests error handling for malformed YAML
func TestLoadItems_InvalidYAMLSyntax(t *testing.T) {
	resetCircuitBreakerForTesting()

	tempDir := t.TempDir()
	invalidYAMLFile := filepath.Join(tempDir, "invalid.yaml")

	// Create file with invalid YAML syntax
	invalidYAMLContent := `
- item_id: "sword_001"
  item_name: "Iron Sword
  item_type: "weapon"  # Missing closing quote above
  invalid_indent:
wrong_nesting
`

	err := os.WriteFile(invalidYAMLFile, []byte(invalidYAMLContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create invalid YAML test file: %v", err)
	}

	items, err := LoadItems(invalidYAMLFile)

	// Should return an error
	if err == nil {
		t.Error("Expected error for invalid YAML syntax, got nil")
	}

	// Should return nil items on error
	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}
}

// TestLoadItems_PartiallyValidYAML tests loading YAML with some missing fields
func TestLoadItems_PartiallyValidYAML(t *testing.T) {
	resetCircuitBreakerForTesting()

	tempDir := t.TempDir()
	partialYAMLFile := filepath.Join(tempDir, "partial.yaml")

	// Create YAML with minimal required fields
	partialYAMLContent := `
- item_id: "minimal_001"
  item_name: "Minimal Item"
  item_type: "misc"
  item_weight: 1
  item_value: 1

- item_id: "partial_002"
  item_name: "Partial Item"
  item_type: "weapon"
  item_damage: "1d4"
  item_weight: 2
  item_value: 25
  item_properties:
    - "light"
`

	err := os.WriteFile(partialYAMLFile, []byte(partialYAMLContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create partial YAML test file: %v", err)
	}

	items, err := LoadItems(partialYAMLFile)
	if err != nil {
		t.Fatalf("LoadItems failed on partial YAML: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Check that omitted fields have default values
	minimal := items[0]
	if minimal.Damage != "" {
		t.Errorf("Expected empty damage for minimal item, got '%s'", minimal.Damage)
	}
	if minimal.AC != 0 {
		t.Errorf("Expected AC 0 for minimal item, got %d", minimal.AC)
	}
	if len(minimal.Properties) != 0 {
		t.Errorf("Expected 0 properties for minimal item, got %d", len(minimal.Properties))
	}

	// Check that provided fields are correctly parsed
	partial := items[1]
	if partial.Damage != "1d4" {
		t.Errorf("Expected damage '1d4' for partial item, got '%s'", partial.Damage)
	}
	if len(partial.Properties) != 1 {
		t.Errorf("Expected 1 property for partial item, got %d", len(partial.Properties))
	}
	if partial.Properties[0] != "light" {
		t.Errorf("Expected property 'light', got '%s'", partial.Properties[0])
	}
}

// TestLoadItems_PermissionDenied tests error handling for permission issues
func TestLoadItems_PermissionDenied(t *testing.T) {
	resetCircuitBreakerForTesting()

	// This test may not work on all systems, so we'll skip it if we can't create the scenario
	tempDir := t.TempDir()
	restrictedFile := filepath.Join(tempDir, "restricted.yaml")

	// Create file and then remove read permissions
	err := os.WriteFile(restrictedFile, []byte("- item_id: test"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create restricted test file: %v", err)
	}

	// Remove read permissions
	err = os.Chmod(restrictedFile, 0o000)
	if err != nil {
		t.Skip("Cannot modify file permissions on this system")
	}

	// Restore permissions after test
	defer func() {
		os.Chmod(restrictedFile, 0o644)
	}()

	items, err := LoadItems(restrictedFile)

	// Should return an error
	if err == nil {
		t.Error("Expected error for permission denied, got nil")
	}

	// Should return nil items on error
	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}
}

// TestLoadItems_TableDriven uses table-driven test approach for multiple scenarios
func TestLoadItems_TableDriven(t *testing.T) {
	resetCircuitBreakerForTesting()

	tempDir := t.TempDir()

	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		expectCount int
		description string
	}{
		{
			name: "Single valid item",
			yamlContent: `
- item_id: "test_001"
  item_name: "Test Item"
  item_type: "test"
  item_weight: 1
  item_value: 10
`,
			expectError: false,
			expectCount: 1,
			description: "Should successfully load single valid item",
		},
		{
			name: "Multiple valid items",
			yamlContent: `
- item_id: "item1"
  item_name: "Item One"
  item_type: "type1"
  item_weight: 1
  item_value: 10

- item_id: "item2"
  item_name: "Item Two"
  item_type: "type2"
  item_weight: 2
  item_value: 20

- item_id: "item3"
  item_name: "Item Three"
  item_type: "type3"
  item_weight: 3
  item_value: 30
`,
			expectError: false,
			expectCount: 3,
			description: "Should successfully load multiple valid items",
		},
		{
			name: "Invalid YAML structure",
			yamlContent: `
not_an_array: true
invalid: structure
`,
			expectError: true,
			expectCount: 0,
			description: "Should fail on invalid YAML structure",
		},
		{
			name: "Mixed valid and invalid items",
			yamlContent: `
- item_id: "valid_001"
  item_name: "Valid Item"
  item_type: "valid"
  item_weight: 1
  item_value: 10

- this is clearly invalid yaml syntax [
`,
			expectError: true,
			expectCount: 0,
			description: "Should fail when YAML has syntax errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, "test_"+tt.name+".yaml")
			err := os.WriteFile(testFile, []byte(tt.yamlContent), 0o644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Run LoadItems
			items, err := LoadItems(testFile)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check item count
			if len(items) != tt.expectCount {
				t.Errorf("Expected %d items, got %d", tt.expectCount, len(items))
			}
		})
	}
}

// TestLoadItems_LargeFile tests performance with a larger YAML file
func TestLoadItems_LargeFile(t *testing.T) {
	resetCircuitBreakerForTesting()

	tempDir := t.TempDir()
	largeFile := filepath.Join(tempDir, "large.yaml")

	// Generate a large YAML file with many items
	var yamlBuilder []byte
	itemCount := 100

	for i := 0; i < itemCount; i++ {
		itemYAML := fmt.Sprintf(`
- item_id: "item_%03d"
  item_name: "Generated Item %d"
  item_type: "generated"
  item_weight: %d
  item_value: %d
`, i, i, i%10+1, i*10)
		yamlBuilder = append(yamlBuilder, []byte(itemYAML)...)
	}

	err := os.WriteFile(largeFile, yamlBuilder, 0o644)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	items, err := LoadItems(largeFile)
	if err != nil {
		t.Fatalf("LoadItems failed on large file: %v", err)
	}

	if len(items) != itemCount {
		t.Errorf("Expected %d items in large file, got %d", itemCount, len(items))
	}

	// Verify first and last items
	if items[0].ID != "item_000" {
		t.Errorf("Expected first item ID 'item_000', got '%s'", items[0].ID)
	}
	if items[itemCount-1].ID != "item_099" {
		t.Errorf("Expected last item ID 'item_099', got '%s'", items[itemCount-1].ID)
	}
}
