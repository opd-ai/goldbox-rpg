package terrain

import (
	"fmt"

	"goldbox-rpg/pkg/pcg"
)

// BiomeDefinition defines the characteristics of a biome
type BiomeDefinition struct {
	Type              pcg.BiomeType         `yaml:"type"`
	DefaultDensity    float64               `yaml:"default_density"`
	WaterLevelRange   [2]float64            `yaml:"water_level_range"`
	RoughnessRange    [2]float64            `yaml:"roughness_range"`
	ConnectivityLevel pcg.ConnectivityLevel `yaml:"connectivity_level"`
	Features          []pcg.TerrainFeature  `yaml:"features"`
	TileDistribution  map[string]float64    `yaml:"tile_distribution"`
}

// biomeDefinitions holds the default biome configurations
var biomeDefinitions = map[pcg.BiomeType]*BiomeDefinition{
	pcg.BiomeCave: {
		Type:              pcg.BiomeCave,
		DefaultDensity:    0.45,
		WaterLevelRange:   [2]float64{0.0, 0.1},
		RoughnessRange:    [2]float64{0.6, 0.8},
		ConnectivityLevel: pcg.ConnectivityModerate,
		Features:          []pcg.TerrainFeature{pcg.FeatureStalactites, pcg.FeatureUndergroundRiver},
		TileDistribution: map[string]float64{
			"wall":  0.45,
			"floor": 0.50,
			"water": 0.03,
			"deep":  0.02,
		},
	},
	pcg.BiomeDungeon: {
		Type:              pcg.BiomeDungeon,
		DefaultDensity:    0.40,
		WaterLevelRange:   [2]float64{0.0, 0.05},
		RoughnessRange:    [2]float64{0.3, 0.5},
		ConnectivityLevel: pcg.ConnectivityHigh,
		Features:          []pcg.TerrainFeature{pcg.FeatureSecretDoors, pcg.FeatureTraps},
		TileDistribution: map[string]float64{
			"wall":   0.40,
			"floor":  0.55,
			"door":   0.03,
			"secret": 0.02,
		},
	},
	pcg.BiomeForest: {
		Type:              pcg.BiomeForest,
		DefaultDensity:    0.35,
		WaterLevelRange:   [2]float64{0.05, 0.15},
		RoughnessRange:    [2]float64{0.4, 0.7},
		ConnectivityLevel: pcg.ConnectivityModerate,
		Features:          []pcg.TerrainFeature{pcg.FeatureTrees, pcg.FeatureStreams},
		TileDistribution: map[string]float64{
			"trees": 0.35,
			"grass": 0.50,
			"water": 0.10,
			"rocks": 0.05,
		},
	},
	pcg.BiomeMountain: {
		Type:              pcg.BiomeMountain,
		DefaultDensity:    0.60,
		WaterLevelRange:   [2]float64{0.0, 0.05},
		RoughnessRange:    [2]float64{0.7, 0.9},
		ConnectivityLevel: pcg.ConnectivityLow,
		Features:          []pcg.TerrainFeature{pcg.FeatureCliffs, pcg.FeatureCrevasses},
		TileDistribution: map[string]float64{
			"rock": 0.60,
			"path": 0.25,
			"snow": 0.10,
			"ice":  0.05,
		},
	},
	pcg.BiomeSwamp: {
		Type:              pcg.BiomeSwamp,
		DefaultDensity:    0.30,
		WaterLevelRange:   [2]float64{0.25, 0.40},
		RoughnessRange:    [2]float64{0.2, 0.4},
		ConnectivityLevel: pcg.ConnectivityLow,
		Features:          []pcg.TerrainFeature{pcg.FeatureBogs, pcg.FeatureVines},
		TileDistribution: map[string]float64{
			"mud":   0.30,
			"water": 0.35,
			"reeds": 0.25,
			"solid": 0.10,
		},
	},
	pcg.BiomeDesert: {
		Type:              pcg.BiomeDesert,
		DefaultDensity:    0.15,
		WaterLevelRange:   [2]float64{0.0, 0.02},
		RoughnessRange:    [2]float64{0.1, 0.3},
		ConnectivityLevel: pcg.ConnectivityHigh,
		Features:          []pcg.TerrainFeature{pcg.FeatureDunes, pcg.FeatureOasis},
		TileDistribution: map[string]float64{
			"sand":   0.85,
			"rock":   0.10,
			"water":  0.02,
			"cactus": 0.03,
		},
	},
}

// GetBiomeDefinition returns the definition for a specific biome
func GetBiomeDefinition(biome pcg.BiomeType) (*BiomeDefinition, error) {
	def, exists := biomeDefinitions[biome]
	if !exists {
		return nil, fmt.Errorf("unknown biome type: %s", biome)
	}

	// Return a copy to prevent modification
	result := *def
	return &result, nil
}

// ApplyBiomeModifications modifies generation parameters based on biome
func ApplyBiomeModifications(params *pcg.TerrainParams, biome pcg.BiomeType) error {
	def, err := GetBiomeDefinition(biome)
	if err != nil {
		return err
	}

	// Apply biome-specific modifications
	if params.Density == 0 {
		params.Density = def.DefaultDensity
	}

	params.BiomeType = biome
	params.Connectivity = def.ConnectivityLevel

	// Adjust water level if not specified
	if params.WaterLevel == 0 {
		params.WaterLevel = (def.WaterLevelRange[0] + def.WaterLevelRange[1]) / 2
	}

	return nil
}

// GetBiomeFeatures returns the available features for a biome
func GetBiomeFeatures(biome pcg.BiomeType) ([]pcg.TerrainFeature, error) {
	def, err := GetBiomeDefinition(biome)
	if err != nil {
		return nil, err
	}

	return def.Features, nil
}

// GetBiomeTileDistribution returns the tile distribution for a biome
func GetBiomeTileDistribution(biome pcg.BiomeType) (map[string]float64, error) {
	def, err := GetBiomeDefinition(biome)
	if err != nil {
		return nil, err
	}

	// Return a copy to prevent modification
	result := make(map[string]float64)
	for k, v := range def.TileDistribution {
		result[k] = v
	}

	return result, nil
}
