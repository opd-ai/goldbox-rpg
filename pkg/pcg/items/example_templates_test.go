package items

import (
	"testing"

	"goldbox-rpg/pkg/pcg"
)

func TestExampleTemplatesFile(t *testing.T) {
	registry := NewItemTemplateRegistry()

	// Load the example templates file
	err := registry.LoadFromFile("/home/user/go/src/github.com/opd-ai/goldbox-rpg/data/pcg/items/templates.yaml")
	if err != nil {
		t.Fatalf("Failed to load example templates file: %v", err)
	}

	// Test that custom templates were loaded
	customSword, err := registry.GetTemplate("custom_sword", pcg.RarityCommon)
	if err != nil {
		t.Errorf("Custom sword template not found: %v", err)
	} else {
		if customSword.BaseType != "weapon" {
			t.Errorf("Expected custom_sword base type 'weapon', got '%s'", customSword.BaseType)
		}

		// Check for custom properties
		hasCustomProperty := false
		for _, prop := range customSword.Properties {
			if prop == "custom" {
				hasCustomProperty = true
				break
			}
		}
		if !hasCustomProperty {
			t.Error("Expected custom_sword to have 'custom' property")
		}
	}

	// Test custom rarity modifier
	uncommonModifier := registry.GetRarityModifier(pcg.RarityUncommon)
	if uncommonModifier.StatMultiplier != 1.15 {
		t.Errorf("Expected custom uncommon stat multiplier 1.15, got %f", uncommonModifier.StatMultiplier)
	}
}
