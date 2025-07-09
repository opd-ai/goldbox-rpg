package items

import (
	"math/rand"
	"testing"
	"time"

	"goldbox-rpg/pkg/pcg"
)

func TestNewItemTemplateRegistry(t *testing.T) {
	registry := NewItemTemplateRegistry()

	if registry == nil {
		t.Fatal("NewItemTemplateRegistry returned nil")
	}

	if registry.templates == nil {
		t.Error("Templates map not initialized")
	}

	if registry.rarityModifiers == nil {
		t.Error("Rarity modifiers map not initialized")
	}
}

func TestLoadDefaultTemplates(t *testing.T) {
	registry := NewItemTemplateRegistry()

	err := registry.LoadDefaultTemplates()
	if err != nil {
		t.Fatalf("LoadDefaultTemplates failed: %v", err)
	}

	// Check that some expected templates are loaded
	expectedTemplates := []string{"sword", "bow", "armor", "potion"}

	for _, templateName := range expectedTemplates {
		if _, exists := registry.templates[templateName]; !exists {
			t.Errorf("Expected template '%s' not found", templateName)
		}
	}

	// Check that rarity modifiers are loaded
	expectedRarities := []pcg.RarityTier{
		pcg.RarityCommon,
		pcg.RarityUncommon,
		pcg.RarityRare,
		pcg.RarityEpic,
		pcg.RarityLegendary,
		pcg.RarityArtifact,
	}

	for _, rarity := range expectedRarities {
		if _, exists := registry.rarityModifiers[rarity]; !exists {
			t.Errorf("Expected rarity modifier for '%s' not found", rarity)
		}
	}
}

func TestGetTemplate(t *testing.T) {
	registry := NewItemTemplateRegistry()
	registry.LoadDefaultTemplates()

	tests := []struct {
		name      string
		baseType  string
		rarity    pcg.RarityTier
		expectErr bool
	}{
		{
			name:      "valid sword template",
			baseType:  "sword",
			rarity:    pcg.RarityCommon,
			expectErr: false,
		},
		{
			name:      "valid bow template",
			baseType:  "bow",
			rarity:    pcg.RarityRare,
			expectErr: false,
		},
		{
			name:      "invalid template",
			baseType:  "nonexistent",
			rarity:    pcg.RarityCommon,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := registry.GetTemplate(tt.baseType, tt.rarity)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if template == nil {
				t.Error("Template is nil")
				return
			}

			if template.BaseType != "weapon" && template.BaseType != "armor" && template.BaseType != "consumable" {
				t.Errorf("Unexpected base type: %s", template.BaseType)
			}
		})
	}
}

func TestGetRarityModifier(t *testing.T) {
	registry := NewItemTemplateRegistry()
	registry.LoadDefaultTemplates()

	tests := []struct {
		name   string
		rarity pcg.RarityTier
	}{
		{"common", pcg.RarityCommon},
		{"uncommon", pcg.RarityUncommon},
		{"rare", pcg.RarityRare},
		{"epic", pcg.RarityEpic},
		{"legendary", pcg.RarityLegendary},
		{"artifact", pcg.RarityArtifact},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := registry.GetRarityModifier(tt.rarity)

			// Basic validation of modifier values
			if modifier.StatMultiplier <= 0 {
				t.Error("StatMultiplier should be positive")
			}

			if modifier.EnchantmentChance < 0 || modifier.EnchantmentChance > 1 {
				t.Error("EnchantmentChance should be between 0 and 1")
			}

			if modifier.MaxEnchantments < 0 {
				t.Error("MaxEnchantments should be non-negative")
			}

			if modifier.ValueMultiplier <= 0 {
				t.Error("ValueMultiplier should be positive")
			}
		})
	}
}

func TestGenerateItemName(t *testing.T) {
	template := &pcg.ItemTemplate{
		BaseType:  "weapon",
		NameParts: []string{"Sword", "Blade"},
		Materials: []string{"iron", "steel"},
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	tests := []struct {
		name   string
		rarity pcg.RarityTier
	}{
		{"common rarity", pcg.RarityCommon},
		{"uncommon rarity", pcg.RarityUncommon},
		{"rare rarity", pcg.RarityRare},
		{"epic rarity", pcg.RarityEpic},
		{"legendary rarity", pcg.RarityLegendary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := GenerateItemName(template, tt.rarity, rng)

			if name == "" {
				t.Error("Generated name is empty")
			}

			// Check that name contains a base name part
			foundBaseName := false
			for _, baseName := range template.NameParts {
				if containsSubstring(name, baseName) {
					foundBaseName = true
					break
				}
			}

			if !foundBaseName {
				t.Errorf("Generated name '%s' doesn't contain expected base name", name)
			}

			// Check that name contains a material
			foundMaterial := false
			for _, material := range template.Materials {
				if containsSubstring(name, material) {
					foundMaterial = true
					break
				}
			}

			if !foundMaterial {
				t.Errorf("Generated name '%s' doesn't contain expected material", name)
			}
		})
	}
}

func TestRarityModifierProgression(t *testing.T) {
	registry := NewItemTemplateRegistry()
	registry.LoadDefaultTemplates()

	rarities := []pcg.RarityTier{
		pcg.RarityCommon,
		pcg.RarityUncommon,
		pcg.RarityRare,
		pcg.RarityEpic,
		pcg.RarityLegendary,
		pcg.RarityArtifact,
	}

	// Check that rarity modifiers increase with rarity
	prevStatMultiplier := 0.0
	prevValueMultiplier := 0.0

	for _, rarity := range rarities {
		modifier := registry.GetRarityModifier(rarity)

		if modifier.StatMultiplier < prevStatMultiplier {
			t.Errorf("StatMultiplier should increase with rarity, but %s (%f) < previous (%f)",
				rarity, modifier.StatMultiplier, prevStatMultiplier)
		}

		if modifier.ValueMultiplier < prevValueMultiplier {
			t.Errorf("ValueMultiplier should increase with rarity, but %s (%f) < previous (%f)",
				rarity, modifier.ValueMultiplier, prevValueMultiplier)
		}

		prevStatMultiplier = modifier.StatMultiplier
		prevValueMultiplier = modifier.ValueMultiplier
	}
}

func TestTemplateStructure(t *testing.T) {
	registry := NewItemTemplateRegistry()
	registry.LoadDefaultTemplates()

	for templateName, template := range registry.templates {
		t.Run(templateName, func(t *testing.T) {
			if template.BaseType == "" {
				t.Error("BaseType is empty")
			}

			if len(template.NameParts) == 0 {
				t.Error("NameParts is empty")
			}

			if len(template.StatRanges) == 0 {
				t.Error("StatRanges is empty")
			}

			if len(template.Materials) == 0 {
				t.Error("Materials is empty")
			}

			if len(template.Rarities) == 0 {
				t.Error("Rarities is empty")
			}

			// Validate stat ranges
			for statName, statRange := range template.StatRanges {
				if statRange.Min > statRange.Max {
					t.Errorf("StatRange for %s has Min > Max: %d > %d", statName, statRange.Min, statRange.Max)
				}

				if statRange.Min < 0 {
					t.Errorf("StatRange for %s has negative Min: %d", statName, statRange.Min)
				}

				if statRange.Scaling < 0 {
					t.Errorf("StatRange for %s has negative Scaling: %f", statName, statRange.Scaling)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsSubstring(str, substr string) bool {
	// Simple case-sensitive check for now
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
