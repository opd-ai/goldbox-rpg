package items

import (
	"os"
	"path/filepath"
	"testing"

	"goldbox-rpg/pkg/pcg"
)

func TestLoadFromFile_FileNotFound(t *testing.T) {
	registry := NewItemTemplateRegistry()

	// Try to load from non-existent file
	err := registry.LoadFromFile("/nonexistent/path/templates.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, but got nil")
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	yamlPath := filepath.Join(tempDir, "invalid.yaml")

	// Create invalid YAML content
	invalidYAML := `invalid: yaml: content:
  - missing: proper
    structure
`

	err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	registry := NewItemTemplateRegistry()
	err = registry.LoadFromFile(yamlPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, but got nil")
	}
}

func TestLoadFromFile_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	yamlPath := filepath.Join(tempDir, "empty.yaml")

	// Create empty file
	err := os.WriteFile(yamlPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	registry := NewItemTemplateRegistry()
	err = registry.LoadFromFile(yamlPath)
	if err != nil {
		t.Errorf("Empty file should fall back to defaults, but got error: %v", err)
	}

	// Should have loaded default templates as fallback
	if len(registry.templates) == 0 {
		t.Error("Expected default templates to be loaded as fallback")
	}
}

func TestLoadFromFile_RarityModifiersOnly(t *testing.T) {
	tempDir := t.TempDir()
	yamlPath := filepath.Join(tempDir, "modifiers_only.yaml")

	yamlContent := `rarity_modifiers:
  test_common:
    stat_multiplier: 1.5
    enchantment_chance: 0.2
    max_enchantments: 1
    value_multiplier: 2.0
    name_prefixes: ["Test"]
    name_suffixes: ["Modifier"]
`

	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	registry := NewItemTemplateRegistry()
	err = registry.LoadFromFile(yamlPath)
	if err != nil {
		t.Errorf("LoadFromFile failed: %v", err)
	}

	// Should have loaded default templates as fallback since no templates in file
	if len(registry.templates) == 0 {
		t.Error("Expected default templates to be loaded as fallback")
	}

	// Should have custom rarity modifier
	modifier := registry.GetRarityModifier(pcg.RarityTier("test_common"))
	if modifier.StatMultiplier != 1.5 {
		t.Errorf("Expected custom rarity modifier to be loaded, got stat multiplier %f", modifier.StatMultiplier)
	}
}
