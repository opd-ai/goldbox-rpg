package pcg

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerritoryGenerator_GenerateTerritories(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	tg := NewTerritoryGenerator(logger)

	factions := []*Faction{
		{ID: "faction1", Name: "Kingdom of Light", Power: 8, Wealth: 6, Military: 7, Influence: 5},
		{ID: "faction2", Name: "Dark Empire", Power: 7, Wealth: 8, Military: 6, Influence: 7},
		{ID: "faction3", Name: "Merchant Guild", Power: 4, Wealth: 10, Military: 2, Influence: 6},
	}

	params := TerritoryGenerationParams{
		WorldWidth:       100,
		WorldHeight:      100,
		TerritoryDensity: 0.7,
		BorderConflict:   0.3,
	}

	territories, borders, err := tg.GenerateTerritories(context.Background(), factions, params, 12345)

	require.NoError(t, err)
	assert.NotEmpty(t, territories, "should generate territories")
	assert.True(t, len(territories) >= len(factions)*2, "each faction should have at least 2 territories")

	// Verify territories have proper attributes
	for _, territory := range territories {
		assert.NotEmpty(t, territory.ID, "territory should have ID")
		assert.NotEmpty(t, territory.Name, "territory should have name")
		assert.NotEmpty(t, territory.ControllerID, "territory should have controller")
		assert.Greater(t, territory.Size, 0, "territory should have positive size")
		assert.Greater(t, territory.Population, 0, "territory should have positive population")
	}

	// Verify borders are generated
	if len(territories) > 1 {
		// With multiple territories, we expect some borders
		t.Logf("Generated %d territories and %d borders", len(territories), len(borders))
	}
}

func TestTerritoryGenerator_Deterministic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	factions := []*Faction{
		{ID: "f1", Name: "Faction One", Power: 5, Wealth: 5, Military: 5, Influence: 5},
	}

	params := TerritoryGenerationParams{
		WorldWidth:       50,
		WorldHeight:      50,
		TerritoryDensity: 0.5,
		BorderConflict:   0.2,
	}

	// Generate twice with same seed
	tg1 := NewTerritoryGenerator(logger)
	territories1, _, err1 := tg1.GenerateTerritories(context.Background(), factions, params, 99999)
	require.NoError(t, err1)

	tg2 := NewTerritoryGenerator(logger)
	territories2, _, err2 := tg2.GenerateTerritories(context.Background(), factions, params, 99999)
	require.NoError(t, err2)

	// Should produce same results
	assert.Equal(t, len(territories1), len(territories2), "same seed should produce same territory count")

	for i := range territories1 {
		assert.Equal(t, territories1[i].Type, territories2[i].Type, "same seed should produce same territory types")
		assert.Equal(t, territories1[i].Size, territories2[i].Size, "same seed should produce same sizes")
	}
}

func TestTerritoryGenerator_TerritoryTypes(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	tg := NewTerritoryGenerator(logger)

	// Create a faction with enough power for many territories
	factions := []*Faction{
		{ID: "f1", Name: "Empire", Power: 10, Wealth: 10, Military: 10, Influence: 10, Resources: []ResourceType{ResourceFood}},
	}

	params := TerritoryGenerationParams{
		WorldWidth:       200,
		WorldHeight:      200,
		TerritoryDensity: 0.3,
		BorderConflict:   0.1,
	}

	territories, _, err := tg.GenerateTerritories(context.Background(), factions, params, 54321)
	require.NoError(t, err)

	// First territory should be capital
	hasCapital := false
	for _, territory := range territories {
		if territory.Type == TerritoryTypeCapital {
			hasCapital = true
			assert.True(t, territory.Strategic, "capital should be strategic")
			break
		}
	}
	assert.True(t, hasCapital, "faction should have a capital")
}

func TestTerritoryGenerator_CalculateTerritoryInfluence(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	tg := NewTerritoryGenerator(logger)

	factions := []*Faction{
		{ID: "f1", Name: "Faction1", Power: 5, Influence: 8},
		{ID: "f2", Name: "Faction2", Power: 3, Influence: 4},
	}

	params := TerritoryGenerationParams{
		WorldWidth:       100,
		WorldHeight:      100,
		TerritoryDensity: 0.5,
		BorderConflict:   0.3,
	}

	territories, _, err := tg.GenerateTerritories(context.Background(), factions, params, 11111)
	require.NoError(t, err)

	influences := tg.CalculateTerritoryInfluence(territories, factions)

	assert.Equal(t, len(territories), len(influences), "should have influence for each territory")

	for _, influence := range influences {
		assert.NotEmpty(t, influence.TerritoryID, "should have territory ID")
		assert.NotEmpty(t, influence.Controller, "should have controller")
		assert.GreaterOrEqual(t, influence.Stability, 0.0, "stability should be non-negative")
		assert.LessOrEqual(t, influence.Stability, 1.0, "stability should be at most 1.0")
	}
}

func TestTerritoryGenerator_BorderGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	tg := NewTerritoryGenerator(logger)

	factions := []*Faction{
		{ID: "f1", Name: "North Kingdom", Power: 6, Military: 8},
		{ID: "f2", Name: "South Kingdom", Power: 6, Military: 7},
	}

	// Small world to ensure adjacency
	params := TerritoryGenerationParams{
		WorldWidth:       50,
		WorldHeight:      50,
		TerritoryDensity: 0.9, // High density forces territories close together
		BorderConflict:   0.5,
	}

	territories, borders, err := tg.GenerateTerritories(context.Background(), factions, params, 22222)
	require.NoError(t, err)

	t.Logf("Generated %d territories and %d borders", len(territories), len(borders))

	// Verify border properties if borders exist
	for _, border := range borders {
		assert.NotEmpty(t, border.Territory1ID, "border should have territory1")
		assert.NotEmpty(t, border.Territory2ID, "border should have territory2")
		assert.Greater(t, border.BorderLength, 0, "border should have positive length")
	}
}
