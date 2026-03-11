package terrain

import (
	"context"
	"testing"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// BenchmarkCellularAutomata_Small benchmarks small terrain generation (50x50)
func BenchmarkCellularAutomata_Small(b *testing.B) {
	generator := NewCellularAutomataGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          50,
			"height":         50,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(ctx, params)
	}
}

// BenchmarkCellularAutomata_Medium benchmarks medium terrain generation (100x100)
func BenchmarkCellularAutomata_Medium(b *testing.B) {
	generator := NewCellularAutomataGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          100,
			"height":         100,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(ctx, params)
	}
}

// BenchmarkCellularAutomata_Large benchmarks large terrain generation (200x200)
func BenchmarkCellularAutomata_Large(b *testing.B) {
	generator := NewCellularAutomataGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          200,
			"height":         200,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(ctx, params)
	}
}

// BenchmarkMaze_Small benchmarks small maze generation (30x30)
func BenchmarkMaze_Small(b *testing.B) {
	generator := NewMazeGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          30,
			"height":         30,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(ctx, params)
	}
}

// BenchmarkMaze_Medium benchmarks medium maze generation (50x50)
func BenchmarkMaze_Medium(b *testing.B) {
	generator := NewMazeGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          50,
			"height":         50,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(ctx, params)
	}
}

// BenchmarkMaze_Large benchmarks large maze generation (100x100)
func BenchmarkMaze_Large(b *testing.B) {
	generator := NewMazeGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          100,
			"height":         100,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(ctx, params)
	}
}

// BenchmarkTerrainGeneration_DifferentBiomes benchmarks generation across different biomes
func BenchmarkTerrainGeneration_DifferentBiomes(b *testing.B) {
	generator := NewCellularAutomataGenerator()
	biomes := []pcg.BiomeType{
		pcg.BiomeDungeon,
		pcg.BiomeForest,
		pcg.BiomeCave,
		pcg.BiomeMountain,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		biome := biomes[i%len(biomes)]
		params := pcg.GenerationParams{
			Seed: 12345,
			Constraints: map[string]interface{}{
				"width":          100,
				"height":         100,
				"biome":          biome,
				"difficulty":     5,
				"connectivity":   pcg.ConnectivityModerate,
				"terrain_params": pcg.TerrainParams{},
			},
		}
		_, _ = generator.Generate(ctx, params)
	}
}

// BenchmarkTileAccess_Sequential benchmarks sequential tile access on generated terrain
func BenchmarkTileAccess_Sequential(b *testing.B) {
	generator := NewCellularAutomataGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          100,
			"height":         100,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	ctx := context.Background()
	result, err := generator.Generate(ctx, params)
	if err != nil {
		b.Fatalf("Failed to generate terrain: %v", err)
	}

	gameMap, ok := result.(*game.GameMap)
	if !ok {
		b.Fatal("Result is not a GameMap")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < gameMap.Height; y++ {
			for x := 0; x < gameMap.Width; x++ {
				_ = gameMap.Tiles[y][x]
			}
		}
	}
}

// BenchmarkTileAccess_Random benchmarks random tile access on generated terrain
func BenchmarkTileAccess_Random(b *testing.B) {
	generator := NewCellularAutomataGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          100,
			"height":         100,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	ctx := context.Background()
	result, err := generator.Generate(ctx, params)
	if err != nil {
		b.Fatalf("Failed to generate terrain: %v", err)
	}

	gameMap, ok := result.(*game.GameMap)
	if !ok {
		b.Fatal("Result is not a GameMap")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % gameMap.Width
		y := (i / gameMap.Width) % gameMap.Height
		_ = gameMap.Tiles[y][x]
	}
}

// BenchmarkCellularAutomata_WithCancellation benchmarks generation with context cancellation overhead
func BenchmarkCellularAutomata_WithCancellation(b *testing.B) {
	generator := NewCellularAutomataGenerator()
	params := pcg.GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"width":          100,
			"height":         100,
			"biome":          pcg.BiomeDungeon,
			"difficulty":     5,
			"connectivity":   pcg.ConnectivityModerate,
			"terrain_params": pcg.TerrainParams{},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		_, _ = generator.Generate(ctx, params)
		cancel()
	}
}
