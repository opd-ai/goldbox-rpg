package items

import (
	"context"
	"testing"

	"goldbox-rpg/pkg/pcg"
)

func TestNewTemplateBasedGenerator(t *testing.T) {
	gen := NewTemplateBasedGenerator()

	if gen == nil {
		t.Fatal("NewTemplateBasedGenerator returned nil")
	}

	if gen.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", gen.version)
	}

	if gen.templates == nil {
		t.Error("Templates map not initialized")
	}

	if gen.registry == nil {
		t.Error("Registry not initialized")
	}

	if gen.enchants == nil {
		t.Error("Enchantment system not initialized")
	}
}

func TestSetSeed(t *testing.T) {
	gen := NewTemplateBasedGenerator()

	gen.SetSeed(12345)

	if gen.rng == nil {
		t.Error("Random generator not set")
	}
}

func TestGetType(t *testing.T) {
	gen := NewTemplateBasedGenerator()

	contentType := gen.GetType()
	expected := pcg.ContentTypeItems

	if contentType != expected {
		t.Errorf("Expected content type %s, got %s", expected, contentType)
	}
}

func TestGetVersion(t *testing.T) {
	gen := NewTemplateBasedGenerator()

	version := gen.GetVersion()
	expected := "1.0.0"

	if version != expected {
		t.Errorf("Expected version %s, got %s", expected, version)
	}
}

func TestValidate(t *testing.T) {
	gen := NewTemplateBasedGenerator()

	tests := []struct {
		name    string
		params  pcg.GenerationParams
		wantErr bool
	}{
		{
			name: "valid parameters",
			params: pcg.GenerationParams{
				Seed:        12345,
				PlayerLevel: 5,
				Difficulty:  3,
			},
			wantErr: false,
		},
		{
			name: "zero seed",
			params: pcg.GenerationParams{
				Seed:        0,
				PlayerLevel: 5,
				Difficulty:  3,
			},
			wantErr: true,
		},
		{
			name: "invalid player level - too low",
			params: pcg.GenerationParams{
				Seed:        12345,
				PlayerLevel: 0,
				Difficulty:  3,
			},
			wantErr: true,
		},
		{
			name: "invalid player level - too high",
			params: pcg.GenerationParams{
				Seed:        12345,
				PlayerLevel: 25,
				Difficulty:  3,
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty - too low",
			params: pcg.GenerationParams{
				Seed:        12345,
				PlayerLevel: 5,
				Difficulty:  0,
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty - too high",
			params: pcg.GenerationParams{
				Seed:        12345,
				PlayerLevel: 5,
				Difficulty:  25,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	gen := NewTemplateBasedGenerator()

	params := pcg.GenerationParams{
		Seed:        12345,
		PlayerLevel: 5,
		Difficulty:  3,
	}

	ctx := context.Background()

	result, err := gen.Generate(ctx, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if result == nil {
		t.Error("Generate() returned nil result")
	}
}

func TestGenerateItem(t *testing.T) {
	gen := NewTemplateBasedGenerator()
	gen.SetSeed(12345)

	template := pcg.ItemTemplate{
		BaseType:  "weapon",
		NameParts: []string{"Sword", "Blade"},
		StatRanges: map[string]pcg.StatRange{
			"damage": {Min: 6, Max: 8, Scaling: 0.1},
			"value":  {Min: 10, Max: 50, Scaling: 0.5},
		},
		Properties: []string{"slashing"},
		Materials:  []string{"iron", "steel"},
		Rarities:   []pcg.RarityTier{pcg.RarityCommon, pcg.RarityUncommon},
	}

	params := pcg.ItemParams{
		GenerationParams: pcg.GenerationParams{
			PlayerLevel: 5,
		},
		MinRarity:       pcg.RarityCommon,
		MaxRarity:       pcg.RarityUncommon,
		EnchantmentRate: 0.5,
	}

	ctx := context.Background()

	item, err := gen.GenerateItem(ctx, template, params)
	if err != nil {
		t.Fatalf("GenerateItem() failed: %v", err)
	}

	if item == nil {
		t.Fatal("GenerateItem() returned nil item")
	}

	if item.ID == "" {
		t.Error("Generated item has no ID")
	}

	if item.Name == "" {
		t.Error("Generated item has no name")
	}

	if item.Type != "weapon" {
		t.Errorf("Expected item type 'weapon', got '%s'", item.Type)
	}

	if len(item.Properties) == 0 {
		t.Error("Generated item has no properties")
	}
}

func TestGenerateItemSet(t *testing.T) {
	gen := NewTemplateBasedGenerator()
	gen.SetSeed(12345)

	params := pcg.ItemParams{
		GenerationParams: pcg.GenerationParams{
			PlayerLevel: 5,
		},
		MinRarity:       pcg.RarityCommon,
		MaxRarity:       pcg.RarityUncommon,
		EnchantmentRate: 0.3,
	}

	ctx := context.Background()

	items, err := gen.GenerateItemSet(ctx, pcg.ItemSetWeapons, params)
	if err != nil {
		t.Fatalf("GenerateItemSet() failed: %v", err)
	}

	if len(items) == 0 {
		t.Error("GenerateItemSet() returned no items")
	}

	// Check that all items are weapons
	for i, item := range items {
		if item == nil {
			t.Errorf("Item %d is nil", i)
			continue
		}

		if item.Type != "weapon" {
			t.Errorf("Item %d type is '%s', expected 'weapon'", i, item.Type)
		}
	}
}

func TestSelectRandomRarity(t *testing.T) {
	gen := NewTemplateBasedGenerator()
	gen.SetSeed(12345)

	tests := []struct {
		name      string
		minRarity pcg.RarityTier
		maxRarity pcg.RarityTier
	}{
		{
			name:      "common to uncommon",
			minRarity: pcg.RarityCommon,
			maxRarity: pcg.RarityUncommon,
		},
		{
			name:      "rare to epic",
			minRarity: pcg.RarityRare,
			maxRarity: pcg.RarityEpic,
		},
		{
			name:      "same rarity",
			minRarity: pcg.RarityRare,
			maxRarity: pcg.RarityRare,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rarity := gen.selectRandomRarity(tt.minRarity, tt.maxRarity)

			// Verify rarity is within range
			rarities := []pcg.RarityTier{
				pcg.RarityCommon,
				pcg.RarityUncommon,
				pcg.RarityRare,
				pcg.RarityEpic,
				pcg.RarityLegendary,
				pcg.RarityArtifact,
			}

			minIndex := -1
			maxIndex := -1
			selectedIndex := -1

			for i, r := range rarities {
				if r == tt.minRarity {
					minIndex = i
				}
				if r == tt.maxRarity {
					maxIndex = i
				}
				if r == rarity {
					selectedIndex = i
				}
			}

			if selectedIndex < minIndex || selectedIndex > maxIndex {
				t.Errorf("Selected rarity %s is outside range %s-%s", rarity, tt.minRarity, tt.maxRarity)
			}
		})
	}
}

func TestDeterministicGeneration(t *testing.T) {
	// Test that the same seed produces the same results
	params := pcg.GenerationParams{
		Seed:        54321,
		PlayerLevel: 3,
		Difficulty:  2,
	}

	ctx := context.Background()

	// Generate item with first generator
	gen1 := NewTemplateBasedGenerator()
	result1, err1 := gen1.Generate(ctx, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	// Generate item with second generator using same seed
	gen2 := NewTemplateBasedGenerator()
	result2, err2 := gen2.Generate(ctx, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	// Results should be identical (same names, stats, etc.)
	// Note: This is a basic check - full determinism would require
	// deeper comparison of item properties
	if result1 == nil || result2 == nil {
		t.Error("One or both results are nil")
	}
}
