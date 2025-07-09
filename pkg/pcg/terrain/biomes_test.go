package terrain

import (
	"testing"

	"goldbox-rpg/pkg/pcg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBiomeDefinition(t *testing.T) {
	tests := []struct {
		name    string
		biome   pcg.BiomeType
		wantErr bool
	}{
		{
			name:    "valid cave biome",
			biome:   pcg.BiomeCave,
			wantErr: false,
		},
		{
			name:    "valid dungeon biome",
			biome:   pcg.BiomeDungeon,
			wantErr: false,
		},
		{
			name:    "valid forest biome",
			biome:   pcg.BiomeForest,
			wantErr: false,
		},
		{
			name:    "invalid biome",
			biome:   pcg.BiomeType("nonexistent"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := GetBiomeDefinition(tt.biome)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, def)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, def)
			assert.Equal(t, tt.biome, def.Type)
			assert.Greater(t, def.DefaultDensity, 0.0)
			assert.LessOrEqual(t, def.DefaultDensity, 1.0)
			assert.NotEmpty(t, def.TileDistribution)
		})
	}
}

func TestApplyBiomeModifications(t *testing.T) {
	tests := []struct {
		name   string
		biome  pcg.BiomeType
		params pcg.TerrainParams
	}{
		{
			name:  "cave biome with empty params",
			biome: pcg.BiomeCave,
			params: pcg.TerrainParams{
				GenerationParams: pcg.GenerationParams{
					Seed: 12345,
				},
			},
		},
		{
			name:  "dungeon biome with existing density",
			biome: pcg.BiomeDungeon,
			params: pcg.TerrainParams{
				GenerationParams: pcg.GenerationParams{
					Seed: 12345,
				},
				Density: 0.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalDensity := tt.params.Density
			err := ApplyBiomeModifications(&tt.params, tt.biome)

			require.NoError(t, err)
			assert.Equal(t, tt.biome, tt.params.BiomeType)

			if originalDensity == 0 {
				assert.Greater(t, tt.params.Density, 0.0)
			} else {
				assert.Equal(t, originalDensity, tt.params.Density)
			}
		})
	}
}

func TestGetBiomeFeatures(t *testing.T) {
	features, err := GetBiomeFeatures(pcg.BiomeCave)
	require.NoError(t, err)
	assert.NotEmpty(t, features)
	assert.Contains(t, features, pcg.FeatureStalactites)

	features, err = GetBiomeFeatures(pcg.BiomeDungeon)
	require.NoError(t, err)
	assert.NotEmpty(t, features)
	assert.Contains(t, features, pcg.FeatureSecretDoors)

	_, err = GetBiomeFeatures(pcg.BiomeType("invalid"))
	assert.Error(t, err)
}

func TestGetBiomeTileDistribution(t *testing.T) {
	dist, err := GetBiomeTileDistribution(pcg.BiomeCave)
	require.NoError(t, err)
	assert.NotEmpty(t, dist)

	// Check that probabilities sum to approximately 1.0
	var sum float64
	for _, prob := range dist {
		sum += prob
		assert.GreaterOrEqual(t, prob, 0.0)
		assert.LessOrEqual(t, prob, 1.0)
	}
	assert.InDelta(t, 1.0, sum, 0.01)

	// Test modification doesn't affect original
	dist["test"] = 0.5
	dist2, err := GetBiomeTileDistribution(pcg.BiomeCave)
	require.NoError(t, err)
	_, exists := dist2["test"]
	assert.False(t, exists)
}

func TestBiomeDefinitionCompleteness(t *testing.T) {
	// Test that all defined biomes have complete definitions
	biomes := []pcg.BiomeType{
		pcg.BiomeCave,
		pcg.BiomeDungeon,
		pcg.BiomeForest,
		pcg.BiomeMountain,
		pcg.BiomeSwamp,
		pcg.BiomeDesert,
	}

	for _, biome := range biomes {
		t.Run(string(biome), func(t *testing.T) {
			def, err := GetBiomeDefinition(biome)
			require.NoError(t, err)

			assert.Equal(t, biome, def.Type)
			assert.Greater(t, def.DefaultDensity, 0.0)
			assert.LessOrEqual(t, def.DefaultDensity, 1.0)
			assert.Len(t, def.WaterLevelRange, 2)
			assert.Len(t, def.RoughnessRange, 2)
			assert.NotEmpty(t, def.TileDistribution)
			assert.NotEmpty(t, def.Features)

			// Validate ranges
			assert.LessOrEqual(t, def.WaterLevelRange[0], def.WaterLevelRange[1])
			assert.LessOrEqual(t, def.RoughnessRange[0], def.RoughnessRange[1])
		})
	}
}
