package pcg

import (
	"context"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

func TestNewWorldGenerator(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with nil logger",
			logger: nil,
		},
		{
			name:   "with provided logger",
			logger: logrus.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := NewWorldGenerator(tt.logger)

			if wg == nil {
				t.Fatal("NewWorldGenerator returned nil")
			}

			if wg.version != "1.0.0" {
				t.Errorf("expected version 1.0.0, got %s", wg.version)
			}

			if wg.logger == nil {
				t.Error("logger should not be nil")
			}

			if wg.rng == nil {
				t.Error("rng should not be nil")
			}
		})
	}
}

func TestWorldGenerator_GetType(t *testing.T) {
	wg := NewWorldGenerator(nil)

	if wg.GetType() != ContentTypeTerrain {
		t.Errorf("expected ContentTypeTerrain, got %s", wg.GetType())
	}
}

func TestWorldGenerator_GetVersion(t *testing.T) {
	wg := NewWorldGenerator(nil)

	if wg.GetVersion() != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", wg.GetVersion())
	}
}

func TestWorldGenerator_Validate(t *testing.T) {
	wg := NewWorldGenerator(nil)

	tests := []struct {
		name      string
		params    GenerationParams
		wantError bool
	}{
		{
			name: "valid parameters",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"world_params": WorldParams{
						WorldWidth:        200,
						WorldHeight:       200,
						RegionCount:       9,
						SettlementCount:   15,
						LandmarkCount:     5,
						Climate:           ClimateTemperate,
						PopulationDensity: 1.0,
						MagicLevel:        5,
						DangerLevel:       3,
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing world_params",
			params: GenerationParams{
				Seed:        12345,
				Constraints: map[string]interface{}{},
			},
			wantError: true,
		},
		{
			name: "world width too small",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"world_params": WorldParams{
						WorldWidth:        10, // Too small
						WorldHeight:       200,
						RegionCount:       9,
						SettlementCount:   15,
						PopulationDensity: 1.0,
						MagicLevel:        5,
						DangerLevel:       3,
					},
				},
			},
			wantError: true,
		},
		{
			name: "world width too large",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"world_params": WorldParams{
						WorldWidth:        2000, // Too large
						WorldHeight:       200,
						RegionCount:       9,
						SettlementCount:   15,
						PopulationDensity: 1.0,
						MagicLevel:        5,
						DangerLevel:       3,
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid population density",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"world_params": WorldParams{
						WorldWidth:        200,
						WorldHeight:       200,
						RegionCount:       9,
						SettlementCount:   15,
						PopulationDensity: -0.5, // Invalid
						MagicLevel:        5,
						DangerLevel:       3,
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid magic level",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"world_params": WorldParams{
						WorldWidth:        200,
						WorldHeight:       200,
						RegionCount:       9,
						SettlementCount:   15,
						PopulationDensity: 1.0,
						MagicLevel:        15, // Too high
						DangerLevel:       3,
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := wg.Validate(tt.params)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.wantError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestWorldGenerator_Generate(t *testing.T) {
	wg := NewWorldGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		params    GenerationParams
		wantError bool
	}{
		{
			name: "successful generation",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				WorldState:  &game.World{},
				Timeout:     30 * time.Second,
				Constraints: map[string]interface{}{
					"world_params": WorldParams{
						WorldWidth:        100,
						WorldHeight:       100,
						RegionCount:       4,
						SettlementCount:   8,
						LandmarkCount:     3,
						Climate:           ClimateTemperate,
						Connectivity:      ConnectivityModerate,
						PopulationDensity: 1.0,
						MagicLevel:        5,
						DangerLevel:       3,
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid parameters",
			params: GenerationParams{
				Seed:        12345,
				Constraints: map[string]interface{}{},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := wg.Generate(ctx, tt.params)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			world, ok := result.(*GeneratedWorld)
			if !ok {
				t.Fatal("result is not a GeneratedWorld")
			}

			// Validate world structure
			if world.ID == "" {
				t.Error("world ID should not be empty")
			}

			if world.Name == "" {
				t.Error("world name should not be empty")
			}

			worldParams := tt.params.Constraints["world_params"].(WorldParams)

			if world.Width != worldParams.WorldWidth {
				t.Errorf("expected width %d, got %d", worldParams.WorldWidth, world.Width)
			}

			if world.Height != worldParams.WorldHeight {
				t.Errorf("expected height %d, got %d", worldParams.WorldHeight, world.Height)
			}

			if len(world.Regions) != worldParams.RegionCount {
				t.Errorf("expected %d regions, got %d", worldParams.RegionCount, len(world.Regions))
			}

			if len(world.Settlements) != worldParams.SettlementCount {
				t.Errorf("expected %d settlements, got %d", worldParams.SettlementCount, len(world.Settlements))
			}

			if len(world.Landmarks) != worldParams.LandmarkCount {
				t.Errorf("expected %d landmarks, got %d", worldParams.LandmarkCount, len(world.Landmarks))
			}

			// Validate that travel paths exist
			if len(world.TravelPaths) == 0 {
				t.Error("expected travel paths to be generated")
			}
		})
	}
}

func TestWorldGenerator_GenerateRegions(t *testing.T) {
	wg := NewWorldGenerator(nil)
	wg.rng.Seed(12345) // Deterministic for testing

	world := &GeneratedWorld{
		Width:  100,
		Height: 100,
	}

	params := WorldParams{
		RegionCount: 4,
		Climate:     ClimateTemperate,
		DangerLevel: 3,
	}

	err := wg.generateRegions(world, params)
	if err != nil {
		t.Fatalf("generateRegions failed: %v", err)
	}

	if len(world.Regions) != 4 {
		t.Errorf("expected 4 regions, got %d", len(world.Regions))
	}

	// Check that regions have valid properties
	for i, region := range world.Regions {
		if region.ID == "" {
			t.Errorf("region %d has empty ID", i)
		}

		if region.Name == "" {
			t.Errorf("region %d has empty name", i)
		}

		if region.Bounds.Width <= 0 || region.Bounds.Height <= 0 {
			t.Errorf("region %d has invalid bounds: %+v", i, region.Bounds)
		}

		if region.Population <= 0 {
			t.Errorf("region %d has invalid population: %d", i, region.Population)
		}
	}
}

func TestWorldGenerator_GenerateSettlements(t *testing.T) {
	wg := NewWorldGenerator(nil)
	wg.rng.Seed(12345) // Deterministic for testing

	world := &GeneratedWorld{
		Width:  100,
		Height: 100,
		Regions: []*Region{
			{
				ID:     "region_0",
				Bounds: Rectangle{X: 0, Y: 0, Width: 50, Height: 50},
			},
			{
				ID:     "region_1",
				Bounds: Rectangle{X: 50, Y: 0, Width: 50, Height: 50},
			},
		},
		Settlements: make([]*Settlement, 0),
	}

	params := WorldParams{
		SettlementCount:   5,
		PopulationDensity: 1.0,
		DangerLevel:       3,
	}

	err := wg.generateSettlements(world, params)
	if err != nil {
		t.Fatalf("generateSettlements failed: %v", err)
	}

	if len(world.Settlements) != 5 {
		t.Errorf("expected 5 settlements, got %d", len(world.Settlements))
	}

	// Check that settlements have valid properties
	for i, settlement := range world.Settlements {
		if settlement.ID == "" {
			t.Errorf("settlement %d has empty ID", i)
		}

		if settlement.Name == "" {
			t.Errorf("settlement %d has empty name", i)
		}

		if settlement.Population <= 0 {
			t.Errorf("settlement %d has invalid population: %d", i, settlement.Population)
		}

		if settlement.Position.X < 0 || settlement.Position.X >= 100 ||
			settlement.Position.Y < 0 || settlement.Position.Y >= 100 {
			t.Errorf("settlement %d has position outside world bounds: %+v", i, settlement.Position)
		}
	}
}

func TestWorldGenerator_GenerateTravelNetwork(t *testing.T) {
	wg := NewWorldGenerator(nil)
	wg.rng.Seed(12345) // Deterministic for testing

	settlements := []*Settlement{
		{
			ID:          "settlement_0",
			Name:        "Town A",
			Position:    game.Position{X: 10, Y: 10},
			Connections: make([]string, 0),
		},
		{
			ID:          "settlement_1",
			Name:        "Town B",
			Position:    game.Position{X: 50, Y: 50},
			Connections: make([]string, 0),
		},
		{
			ID:          "settlement_2",
			Name:        "Town C",
			Position:    game.Position{X: 80, Y: 20},
			Connections: make([]string, 0),
		},
	}

	world := &GeneratedWorld{
		Width:       100,
		Height:      100,
		Settlements: settlements,
		TravelPaths: make([]*TravelPath, 0),
	}

	params := WorldParams{
		Connectivity: ConnectivityModerate,
	}

	err := wg.generateTravelNetwork(world, params)
	if err != nil {
		t.Fatalf("generateTravelNetwork failed: %v", err)
	}

	if len(world.TravelPaths) == 0 {
		t.Error("expected travel paths to be generated")
	}

	// Check that travel paths have valid properties
	for i, path := range world.TravelPaths {
		if path.ID == "" {
			t.Errorf("travel path %d has empty ID", i)
		}

		if path.From == "" || path.To == "" {
			t.Errorf("travel path %d has empty from/to: %s -> %s", i, path.From, path.To)
		}

		if path.TravelTime <= 0 {
			t.Errorf("travel path %d has invalid travel time: %d", i, path.TravelTime)
		}

		if len(path.Points) < 2 {
			t.Errorf("travel path %d should have at least 2 points", i)
		}
	}

	// Check that settlements are connected
	totalConnections := 0
	for _, settlement := range world.Settlements {
		totalConnections += len(settlement.Connections)
	}

	if totalConnections == 0 {
		t.Error("settlements should have connections after travel network generation")
	}
}

func TestWorldGenerator_Deterministic(t *testing.T) {
	// Test that the same seed produces the same world
	params := GenerationParams{
		Seed:        54321,
		Difficulty:  5,
		PlayerLevel: 3,
		WorldState:  &game.World{},
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"world_params": WorldParams{
				WorldWidth:        80,
				WorldHeight:       80,
				RegionCount:       4,
				SettlementCount:   6,
				LandmarkCount:     2,
				Climate:           ClimateTemperate,
				PopulationDensity: 1.0,
				MagicLevel:        5,
				DangerLevel:       3,
			},
		},
	}

	wg1 := NewWorldGenerator(nil)
	wg2 := NewWorldGenerator(nil)

	ctx := context.Background()

	result1, err1 := wg1.Generate(ctx, params)
	if err1 != nil {
		t.Fatalf("first generation failed: %v", err1)
	}

	result2, err2 := wg2.Generate(ctx, params)
	if err2 != nil {
		t.Fatalf("second generation failed: %v", err2)
	}

	world1 := result1.(*GeneratedWorld)
	world2 := result2.(*GeneratedWorld)

	// Compare basic properties
	if world1.Name != world2.Name {
		t.Errorf("world names should be identical with same seed: %s != %s", world1.Name, world2.Name)
	}

	if len(world1.Settlements) != len(world2.Settlements) {
		t.Errorf("settlement counts should be identical: %d != %d", len(world1.Settlements), len(world2.Settlements))
	}

	// Compare first settlement positions (should be identical with same seed)
	if len(world1.Settlements) > 0 && len(world2.Settlements) > 0 {
		pos1 := world1.Settlements[0].Position
		pos2 := world2.Settlements[0].Position

		if pos1.X != pos2.X || pos1.Y != pos2.Y {
			t.Errorf("first settlement positions should be identical: %+v != %+v", pos1, pos2)
		}
	}
}

func TestWorldGenerator_HelperMethods(t *testing.T) {
	wg := NewWorldGenerator(nil)
	wg.rng.Seed(12345)

	// Test name generation methods
	worldName := wg.generateWorldName(ClimateTemperate)
	if worldName == "" {
		t.Error("world name should not be empty")
	}

	regionName := wg.generateRegionName(0)
	if regionName == "" {
		t.Error("region name should not be empty")
	}

	settlementName := wg.generateSettlementName()
	if settlementName == "" {
		t.Error("settlement name should not be empty")
	}

	landmarkName := wg.generateLandmarkName()
	if landmarkName == "" {
		t.Error("landmark name should not be empty")
	}

	// Test distance calculation
	pos1 := game.Position{X: 0, Y: 0}
	pos2 := game.Position{X: 3, Y: 4}
	distance := wg.calculateDistance(pos1, pos2)
	if distance != 5 { // 3-4-5 triangle
		t.Errorf("expected distance 5, got %d", distance)
	}

	// Test biome selection
	biome := wg.chooseBiome(ClimateTemperate)
	validBiomes := []BiomeType{BiomeForest, BiomeMountain, BiomeCoastal, BiomeUrban}
	found := false
	for _, valid := range validBiomes {
		if biome == valid {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("invalid biome for temperate climate: %s", biome)
	}
}

// Benchmark tests

func BenchmarkWorldGenerator_Generate(b *testing.B) {
	wg := NewWorldGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		WorldState:  &game.World{},
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"world_params": WorldParams{
				WorldWidth:        100,
				WorldHeight:       100,
				RegionCount:       9,
				SettlementCount:   20,
				LandmarkCount:     5,
				Climate:           ClimateTemperate,
				PopulationDensity: 1.0,
				MagicLevel:        5,
				DangerLevel:       3,
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Use different seed for each iteration
		params.Seed = int64(i)
		_, err := wg.Generate(ctx, params)
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}
}

func BenchmarkWorldGenerator_GenerateSmallWorld(b *testing.B) {
	wg := NewWorldGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		WorldState:  &game.World{},
		Timeout:     30 * time.Second,
		Constraints: map[string]interface{}{
			"world_params": WorldParams{
				WorldWidth:        50,
				WorldHeight:       50,
				RegionCount:       4,
				SettlementCount:   8,
				LandmarkCount:     2,
				Climate:           ClimateTemperate,
				PopulationDensity: 1.0,
				MagicLevel:        5,
				DangerLevel:       3,
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		params.Seed = int64(i)
		_, err := wg.Generate(ctx, params)
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}
}
