package items

import (
	"os"
	"path/filepath"
	"testing"

	"goldbox-rpg/pkg/pcg"
)

// Test_PCG_Template_YAML_Loading_Regression ensures that LoadFromFile
// properly loads custom templates from YAML files instead of ignoring
// the configPath parameter and only loading defaults.
// This is a regression test for the bug documented in AUDIT.md.
func Test_PCG_Template_YAML_Loading_Regression(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	yamlPath := filepath.Join(tempDir, "test_templates.yaml")

	// Create a test YAML file with item templates
	yamlContent := `templates:
  test_weapon:
    base_type: "test_weapon"
    name_parts: ["Test", "Custom", "YAML"]
    stat_ranges:
      damage:
        min: 10
        max: 15
        scaling: 0.2
      value:
        min: 100
        max: 200
        scaling: 1.0
    properties: ["testing", "yaml_loaded"]
    materials: ["custom_material", "test_alloy"]
    rarities: ["common", "uncommon"]
`

	err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	// Create registry and attempt to load from file
	registry := NewItemTemplateRegistry()

	// This should load the custom template from YAML but currently doesn't
	err = registry.LoadFromFile(yamlPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify that the custom template was loaded (this should fail with current implementation)
	template, err := registry.GetTemplate("test_weapon", pcg.RarityCommon)
	if err != nil {
		t.Errorf("Expected custom template to be loaded, but got error: %v", err)
	}

	if template != nil {
		// Verify the template has the custom properties from YAML
		if template.BaseType != "test_weapon" {
			t.Errorf("Expected BaseType 'test_weapon', got '%s'", template.BaseType)
		}

		// Check for custom name parts that would only come from YAML
		foundCustomName := false
		for _, namePart := range template.NameParts {
			if namePart == "YAML" {
				foundCustomName = true
				break
			}
		}
		if !foundCustomName {
			t.Errorf("Expected to find 'YAML' in name parts, indicating template was loaded from file")
		}

		// Check for custom properties that would only come from YAML
		foundCustomProperty := false
		for _, prop := range template.Properties {
			if prop == "yaml_loaded" {
				foundCustomProperty = true
				break
			}
		}
		if !foundCustomProperty {
			t.Errorf("Expected to find 'yaml_loaded' property, indicating template was loaded from file")
		}
	}
}
