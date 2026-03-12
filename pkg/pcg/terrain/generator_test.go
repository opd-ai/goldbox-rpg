package terrain

import (
	"context"
	"testing"

	"goldbox-rpg/pkg/pcg"
)

func TestCellularAutomataGenerator_GetType(t *testing.T) {
	gen := NewCellularAutomataGenerator()
	if gen.GetType() != pcg.ContentTypeTerrain {
		t.Errorf("GetType() = %v, want %v", gen.GetType(), pcg.ContentTypeTerrain)
	}
}

func TestCellularAutomataGenerator_GetVersion(t *testing.T) {
	gen := NewCellularAutomataGenerator()
	version := gen.GetVersion()
	if version == "" {
		t.Error("GetVersion() returned empty string")
	}
}

func TestCellularAutomataGenerator_Validate(t *testing.T) {
	gen := NewCellularAutomataGenerator()

	// Test with valid params including terrain_params
	terrainParams := pcg.TerrainParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        12345,
			Difficulty:  1,
			PlayerLevel: 5,
		},
		BiomeType: pcg.BiomeForest,
		Density:   0.45,
	}
	params := pcg.GenerationParams{
		Seed:        12345,
		Difficulty:  1,
		PlayerLevel: 5,
		Constraints: map[string]interface{}{
			"terrain_params": terrainParams,
		},
	}

	err := gen.Validate(params)
	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestCellularAutomataGenerator_Validate_MissingParams(t *testing.T) {
	gen := NewCellularAutomataGenerator()

	// Test without terrain_params - should error
	params := pcg.GenerationParams{
		Seed:       12345,
		Difficulty: 1,
	}

	err := gen.Validate(params)
	if err == nil {
		t.Error("Validate() expected error for missing terrain_params")
	}
}

func TestCellularAutomataGenerator_ValidateConnectivity(t *testing.T) {
	gen := NewCellularAutomataGenerator()

	// Generate a small map to test connectivity
	params := pcg.TerrainParams{
		GenerationParams: pcg.GenerationParams{
			Seed:       12345,
			Difficulty: 1,
		},
		Density: 0.45,
	}

	ctx := context.Background()
	gameMap, err := gen.GenerateTerrain(ctx, 20, 20, params)
	if err != nil {
		t.Fatalf("GenerateTerrain() error: %v", err)
	}

	// Test connectivity validation
	isValid := gen.ValidateConnectivity(gameMap)
	// Result doesn't matter, just that it runs without panic
	_ = isValid
}

func TestCellularAutomataGenerator_Generate(t *testing.T) {
	gen := NewCellularAutomataGenerator()

	terrainParams := pcg.TerrainParams{
		GenerationParams: pcg.GenerationParams{
			Seed:       12345,
			Difficulty: 1,
		},
		Density: 0.45,
	}
	params := pcg.GenerationParams{
		Seed:       12345,
		Difficulty: 1,
		Constraints: map[string]interface{}{
			"terrain_params": terrainParams,
			"width":          20,
			"height":         20,
		},
	}

	ctx := context.Background()
	result, err := gen.Generate(ctx, params)
	if err != nil {
		t.Errorf("Generate() error: %v", err)
		return
	}
	if result == nil {
		t.Error("Generate() returned nil result")
	}
}

func TestCellularAutomataGenerator_GenerateTerrain(t *testing.T) {
	gen := NewCellularAutomataGenerator()

	params := pcg.TerrainParams{
		GenerationParams: pcg.GenerationParams{
			Seed:       12345,
			Difficulty: 1,
		},
		Density: 0.45,
	}

	ctx := context.Background()
	gameMap, err := gen.GenerateTerrain(ctx, 20, 20, params)
	if err != nil {
		t.Errorf("GenerateTerrain() error: %v", err)
		return
	}
	if gameMap == nil {
		t.Error("GenerateTerrain() returned nil")
		return
	}
	if gameMap.Width != 20 {
		t.Errorf("Width = %d, want 20", gameMap.Width)
	}
	if gameMap.Height != 20 {
		t.Errorf("Height = %d, want 20", gameMap.Height)
	}
}

func TestCellularAutomataGenerator_GenerateBiome(t *testing.T) {
	gen := NewCellularAutomataGenerator()

	params := pcg.TerrainParams{
		GenerationParams: pcg.GenerationParams{
			Seed:       12345,
			Difficulty: 1,
		},
		BiomeType: pcg.BiomeForest,
		Density:   0.45,
	}

	bounds := pcg.Rectangle{Width: 30, Height: 30}
	ctx := context.Background()
	gameMap, err := gen.GenerateBiome(ctx, pcg.BiomeForest, bounds, params)
	if err != nil {
		t.Errorf("GenerateBiome() error: %v", err)
		return
	}
	if gameMap == nil {
		t.Error("GenerateBiome() returned nil")
		return
	}
	if gameMap.Width != 30 {
		t.Errorf("Width = %d, want 30", gameMap.Width)
	}
}
