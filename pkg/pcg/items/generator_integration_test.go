package items

import (
	"context"
	"testing"

	"goldbox-rpg/pkg/pcg"
)

func TestGenerator_LoadFromFile_Integration(t *testing.T) {
	// Create a generator and load custom templates
	gen := NewTemplateBasedGenerator()
	gen.SetSeed(12345)
	
	// Load custom templates from our example file
	err := gen.LoadTemplates("/home/user/go/src/github.com/opd-ai/goldbox-rpg/data/pcg/items/templates.yaml")
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	// Test generating an item using a custom template
	ctx := context.Background()
	
	// Try to get the custom template
	template, err := gen.registry.GetTemplate("custom_sword", pcg.RarityRare)
	if err != nil {
		t.Fatalf("Failed to get custom template: %v", err)
	}

	// Generate an item from the custom template
	params := pcg.ItemParams{
		GenerationParams: pcg.GenerationParams{
			PlayerLevel: 10,
		},
		MinRarity:       pcg.RarityRare,
		MaxRarity:       pcg.RarityRare,
		EnchantmentRate: 0.5,
	}

	item, err := gen.GenerateItem(ctx, *template, params)
	if err != nil {
		t.Fatalf("Failed to generate item from custom template: %v", err)
	}

	// Verify the item has properties from the custom template
	if item.Type != "weapon" {
		t.Errorf("Expected item type 'weapon', got '%s'", item.Type)
	}

	// Check for custom properties
	hasCustomProperty := false
	for _, prop := range item.Properties {
		if prop == "custom" {
			hasCustomProperty = true
			break
		}
	}
	if !hasCustomProperty {
		t.Error("Generated item should have 'custom' property from template")
	}

	// Verify the item name contains parts from the template
	expectedNameParts := []string{"Blade", "Edge", "Cutter"}
	foundNamePart := false
	for _, part := range expectedNameParts {
		if contains(item.Name, part) {
			foundNamePart = true
			break
		}
	}
	if !foundNamePart {
		t.Errorf("Item name '%s' should contain one of the custom name parts", item.Name)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
		 (s[:len(substr)] == substr || 
		  s[len(s)-len(substr):] == substr ||
		  indexOf(s, substr) != -1)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
