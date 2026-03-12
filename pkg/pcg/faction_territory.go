// Package pcg provides procedural content generation for faction territory systems.
package pcg

import (
	"context"
	"math"
	"math/rand"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// TerritoryGenerator creates geographically-aware faction territories.
// Generates territories based on faction power, biome compatibility, and world geography.
type TerritoryGenerator struct {
	logger *logrus.Logger
	rng    *rand.Rand
}

// NewTerritoryGenerator creates a new territory generator instance.
func NewTerritoryGenerator(logger *logrus.Logger) *TerritoryGenerator {
	if logger == nil {
		logger = logrus.New()
	}
	return &TerritoryGenerator{
		logger: logger,
	}
}

// TerritoryGenerationParams defines parameters for territory generation.
type TerritoryGenerationParams struct {
	WorldWidth       int            // World width in tiles
	WorldHeight      int            // World height in tiles
	Biomes           []BiomeType    // Available biome types
	FactionPowers    map[string]int // Faction ID -> power level
	TerritoryDensity float64        // 0.0-1.0, how packed territories should be
	BorderConflict   float64        // 0.0-1.0, how much overlap/contested zones
}

// TerritoryBorder represents a shared border between territories.
type TerritoryBorder struct {
	Territory1ID string                 `json:"territory1_id"`
	Territory2ID string                 `json:"territory2_id"`
	BorderLength int                    `json:"border_length"` // Length of shared border
	Contested    bool                   `json:"contested"`     // Whether border is disputed
	Fortified    bool                   `json:"fortified"`     // Whether border has defenses
	Properties   map[string]interface{} `json:"properties"`
}

// TerritoryInfluence tracks faction influence over a territory.
type TerritoryInfluence struct {
	TerritoryID string             `json:"territory_id"`
	Influences  map[string]float64 `json:"influences"` // Faction ID -> influence (0.0-1.0)
	Controller  string             `json:"controller"` // Current controlling faction
	Stability   float64            `json:"stability"`  // 0.0-1.0, how stable control is
}

// GenerateTerritories creates territories for a faction system based on geography.
func (tg *TerritoryGenerator) GenerateTerritories(
	ctx context.Context,
	factions []*Faction,
	params TerritoryGenerationParams,
	seed int64,
) ([]*Territory, []*TerritoryBorder, error) {
	tg.rng = rand.New(rand.NewSource(seed))

	tg.logger.WithFields(logrus.Fields{
		"faction_count": len(factions),
		"world_size":    params.WorldWidth * params.WorldHeight,
		"seed":          seed,
	}).Info("generating faction territories")

	territories := make([]*Territory, 0)
	borders := make([]*TerritoryBorder, 0)

	// Calculate total power for proportional territory distribution
	totalPower := 0
	for _, faction := range factions {
		totalPower += faction.Power
	}

	if totalPower == 0 {
		totalPower = len(factions) // Avoid division by zero
	}

	// Generate territories for each faction based on power
	usedPositions := make(map[game.Position]bool)

	for _, faction := range factions {
		factionTerritories := tg.generateFactionTerritories(
			faction, params, totalPower, usedPositions,
		)
		territories = append(territories, factionTerritories...)
	}

	// Generate borders between adjacent territories
	borders = tg.generateTerritoryBorders(territories, params)

	tg.logger.WithFields(logrus.Fields{
		"territories_generated": len(territories),
		"borders_generated":     len(borders),
	}).Info("territory generation completed")

	return territories, borders, nil
}

// generateFactionTerritories creates territories for a single faction.
func (tg *TerritoryGenerator) generateFactionTerritories(
	faction *Faction,
	params TerritoryGenerationParams,
	totalPower int,
	usedPositions map[game.Position]bool,
) []*Territory {
	// Calculate number of territories based on faction power
	baseTerritories := 2
	powerRatio := float64(faction.Power) / float64(totalPower)
	extraTerritories := int(powerRatio * 5) // Up to 5 extra territories for powerful factions
	territoryCount := baseTerritories + extraTerritories

	territories := make([]*Territory, 0, territoryCount)

	// Generate capital first (always present)
	capital := tg.generateTerritory(faction, TerritoryTypeCapital, params, usedPositions)
	if capital != nil {
		territories = append(territories, capital)
	}

	// Generate remaining territories
	for i := 1; i < territoryCount; i++ {
		territoryType := tg.selectTerritoryTypeByIndex(i, faction)
		territory := tg.generateTerritory(faction, territoryType, params, usedPositions)
		if territory != nil {
			territories = append(territories, territory)
		}
	}

	return territories
}

// generateTerritory creates a single territory with geographic awareness.
func (tg *TerritoryGenerator) generateTerritory(
	faction *Faction,
	territoryType TerritoryType,
	params TerritoryGenerationParams,
	usedPositions map[game.Position]bool,
) *Territory {
	// Find a valid position
	maxAttempts := 50
	var pos game.Position

	for attempt := 0; attempt < maxAttempts; attempt++ {
		pos = game.Position{
			X: tg.rng.Intn(params.WorldWidth),
			Y: tg.rng.Intn(params.WorldHeight),
		}

		// Check if position is not too close to existing territories
		if !tg.isPositionTooClose(pos, usedPositions, params) {
			break
		}

		if attempt == maxAttempts-1 {
			// Couldn't find a good position, use random
			pos = game.Position{
				X: tg.rng.Intn(params.WorldWidth),
				Y: tg.rng.Intn(params.WorldHeight),
			}
		}
	}

	usedPositions[pos] = true

	// Calculate territory properties based on type and faction
	size := tg.calculateTerritorySize(territoryType, faction.Power)
	population := tg.calculatePopulation(territoryType, size, faction.Wealth)
	defenses := tg.calculateDefenses(territoryType, faction.Military)

	return &Territory{
		ID:           generateTerritoryID(),
		Name:         tg.generateTerritoryName(faction.Name, territoryType),
		Type:         territoryType,
		ControllerID: faction.ID,
		Position:     pos,
		Size:         size,
		Population:   population,
		Defenses:     defenses,
		Resources:    tg.selectResources(territoryType, faction.Resources),
		Strategic:    tg.isStrategicLocation(territoryType, pos),
		Properties:   make(map[string]interface{}),
	}
}

// isPositionTooClose checks if a position is too close to existing territories.
func (tg *TerritoryGenerator) isPositionTooClose(
	pos game.Position,
	usedPositions map[game.Position]bool,
	params TerritoryGenerationParams,
) bool {
	minDistance := int(float64(params.WorldWidth) * (1.0 - params.TerritoryDensity) * 0.1)
	if minDistance < 5 {
		minDistance = 5
	}

	for existingPos := range usedPositions {
		dist := tg.distance(pos, existingPos)
		if dist < minDistance {
			return true
		}
	}
	return false
}

// distance calculates Manhattan distance between two positions.
func (tg *TerritoryGenerator) distance(p1, p2 game.Position) int {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// selectTerritoryTypeByIndex selects territory type based on generation order.
func (tg *TerritoryGenerator) selectTerritoryTypeByIndex(index int, faction *Faction) TerritoryType {
	// Distribution based on faction characteristics
	if index == 1 {
		// Second territory is usually a city
		return TerritoryTypeCity
	}

	// Military factions prefer fortresses
	if faction.Military > faction.Wealth && tg.rng.Float64() < 0.4 {
		return TerritoryTypeFortress
	}

	// Economic factions prefer trading posts
	if faction.Wealth > faction.Military && tg.rng.Float64() < 0.4 {
		return TerritoryTypeTradingPost
	}

	// Random selection for remaining
	types := []TerritoryType{
		TerritoryTypeCity,
		TerritoryTypeOutpost,
		TerritoryTypeFortress,
		TerritoryTypeTradingPost,
		TerritoryTypeResource,
	}
	return types[tg.rng.Intn(len(types))]
}

// calculateTerritorySize determines territory size based on type and power.
func (tg *TerritoryGenerator) calculateTerritorySize(territoryType TerritoryType, power int) int {
	baseSize := 20
	powerBonus := power * 3

	switch territoryType {
	case TerritoryTypeCapital:
		return baseSize + powerBonus + tg.rng.Intn(30)
	case TerritoryTypeCity:
		return baseSize + powerBonus/2 + tg.rng.Intn(20)
	case TerritoryTypeFortress:
		return baseSize/2 + tg.rng.Intn(15)
	case TerritoryTypeOutpost:
		return baseSize/3 + tg.rng.Intn(10)
	case TerritoryTypeTradingPost:
		return baseSize/2 + tg.rng.Intn(15)
	case TerritoryTypeResource:
		return baseSize + tg.rng.Intn(25)
	default:
		return baseSize + tg.rng.Intn(20)
	}
}

// calculatePopulation determines territory population based on type and size.
func (tg *TerritoryGenerator) calculatePopulation(territoryType TerritoryType, size, wealth int) int {
	basePopulation := size * 100
	wealthBonus := wealth * 50

	switch territoryType {
	case TerritoryTypeCapital:
		return basePopulation*3 + wealthBonus + tg.rng.Intn(5000)
	case TerritoryTypeCity:
		return basePopulation*2 + wealthBonus/2 + tg.rng.Intn(3000)
	case TerritoryTypeFortress:
		return basePopulation/4 + tg.rng.Intn(500) // Mostly military
	case TerritoryTypeOutpost:
		return basePopulation/5 + tg.rng.Intn(200)
	case TerritoryTypeTradingPost:
		return basePopulation + wealthBonus/2 + tg.rng.Intn(1000)
	case TerritoryTypeResource:
		return basePopulation/3 + tg.rng.Intn(800) // Workers
	default:
		return basePopulation + tg.rng.Intn(1000)
	}
}

// calculateDefenses determines territory defenses based on type and military.
func (tg *TerritoryGenerator) calculateDefenses(territoryType TerritoryType, military int) int {
	baseDefense := 3

	switch territoryType {
	case TerritoryTypeCapital:
		return baseDefense + military + tg.rng.Intn(5)
	case TerritoryTypeCity:
		return baseDefense + military/2 + tg.rng.Intn(3)
	case TerritoryTypeFortress:
		return baseDefense*2 + military + tg.rng.Intn(5) // Heavily defended
	case TerritoryTypeOutpost:
		return baseDefense + tg.rng.Intn(3)
	case TerritoryTypeTradingPost:
		return baseDefense/2 + tg.rng.Intn(2)
	case TerritoryTypeResource:
		return baseDefense + tg.rng.Intn(3)
	default:
		return baseDefense + tg.rng.Intn(3)
	}
}

// selectResources picks resources appropriate for the territory type.
func (tg *TerritoryGenerator) selectResources(
	territoryType TerritoryType,
	factionResources []ResourceType,
) []ResourceType {
	// Start with faction resources
	resources := make([]ResourceType, 0)
	if len(factionResources) > 0 {
		// Copy some faction resources
		copyCount := 1 + tg.rng.Intn(len(factionResources))
		if copyCount > len(factionResources) {
			copyCount = len(factionResources)
		}
		for i := 0; i < copyCount; i++ {
			resources = append(resources, factionResources[i])
		}
	}

	// Add territory-type specific resources
	switch territoryType {
	case TerritoryTypeResource:
		// Resource territories always have resources
		if len(resources) == 0 {
			resources = append(resources, ResourceFood, ResourceStone)
		}
	case TerritoryTypeTradingPost:
		resources = append(resources, ResourceGold)
	}

	return resources
}

// isStrategicLocation determines if a territory position is strategic.
func (tg *TerritoryGenerator) isStrategicLocation(territoryType TerritoryType, pos game.Position) bool {
	// Capitals and fortresses are always strategic
	if territoryType == TerritoryTypeCapital || territoryType == TerritoryTypeFortress {
		return true
	}
	// Random chance for other locations
	return tg.rng.Float64() < 0.2
}

// generateTerritoryName creates a name for a territory.
func (tg *TerritoryGenerator) generateTerritoryName(factionName string, territoryType TerritoryType) string {
	prefixes := []string{"North", "South", "East", "West", "Upper", "Lower", "Great", "Old", "New"}
	suffixes := map[TerritoryType][]string{
		TerritoryTypeCapital:     {"Capital", "Seat", "Throne", "Heart"},
		TerritoryTypeCity:        {"City", "Town", "Haven", "Port"},
		TerritoryTypeFortress:    {"Fortress", "Keep", "Stronghold", "Bastion"},
		TerritoryTypeOutpost:     {"Outpost", "Watch", "Post", "Camp"},
		TerritoryTypeTradingPost: {"Market", "Exchange", "Bazaar", "Trade Hub"},
		TerritoryTypeResource:    {"Mines", "Fields", "Quarry", "Grove"},
	}

	prefix := prefixes[tg.rng.Intn(len(prefixes))]
	suffix := "Territory"
	if s, ok := suffixes[territoryType]; ok {
		suffix = s[tg.rng.Intn(len(s))]
	}

	return prefix + " " + factionName + " " + suffix
}

// generateTerritoryBorders creates border relationships between territories.
func (tg *TerritoryGenerator) generateTerritoryBorders(
	territories []*Territory,
	params TerritoryGenerationParams,
) []*TerritoryBorder {
	borders := make([]*TerritoryBorder, 0)
	adjacencyThreshold := int(float64(params.WorldWidth) * 0.15)

	for i := 0; i < len(territories); i++ {
		for j := i + 1; j < len(territories); j++ {
			t1, t2 := territories[i], territories[j]
			dist := tg.distance(t1.Position, t2.Position)

			// Consider territories adjacent if close enough
			if dist <= adjacencyThreshold {
				border := &TerritoryBorder{
					Territory1ID: t1.ID,
					Territory2ID: t2.ID,
					BorderLength: tg.calculateBorderLength(t1, t2, dist),
					Contested:    tg.isBorderContested(t1, t2, params.BorderConflict),
					Fortified:    tg.isBorderFortified(t1, t2),
					Properties:   make(map[string]interface{}),
				}
				borders = append(borders, border)
			}
		}
	}

	return borders
}

// calculateBorderLength estimates border length based on territory sizes.
func (tg *TerritoryGenerator) calculateBorderLength(t1, t2 *Territory, distance int) int {
	// Larger territories have longer borders
	avgSize := (t1.Size + t2.Size) / 2
	proximityFactor := 1.0 - (float64(distance) / 100.0)
	if proximityFactor < 0.1 {
		proximityFactor = 0.1
	}
	return int(float64(avgSize) * proximityFactor * math.Sqrt(float64(avgSize)))
}

// isBorderContested determines if a border is disputed.
func (tg *TerritoryGenerator) isBorderContested(t1, t2 *Territory, conflictLevel float64) bool {
	// Different controllers means potential contest
	if t1.ControllerID != t2.ControllerID {
		return tg.rng.Float64() < conflictLevel
	}
	return false
}

// isBorderFortified determines if a border has fortifications.
func (tg *TerritoryGenerator) isBorderFortified(t1, t2 *Territory) bool {
	// Fortresses always have fortified borders
	if t1.Type == TerritoryTypeFortress || t2.Type == TerritoryTypeFortress {
		return true
	}
	// Different controllers more likely to fortify
	if t1.ControllerID != t2.ControllerID {
		return tg.rng.Float64() < 0.4
	}
	return tg.rng.Float64() < 0.1
}

// CalculateTerritoryInfluence computes faction influence over all territories.
func (tg *TerritoryGenerator) CalculateTerritoryInfluence(
	territories []*Territory,
	factions []*Faction,
) []*TerritoryInfluence {
	influences := make([]*TerritoryInfluence, len(territories))

	for i, territory := range territories {
		influence := &TerritoryInfluence{
			TerritoryID: territory.ID,
			Influences:  make(map[string]float64),
			Controller:  territory.ControllerID,
			Stability:   1.0, // Start with full stability
		}

		// Calculate each faction's influence
		for _, faction := range factions {
			if faction.ID == territory.ControllerID {
				influence.Influences[faction.ID] = 1.0
			} else {
				// Distance-based influence decay
				influenceLevel := tg.calculateFactionInfluence(territory, faction, territories)
				if influenceLevel > 0.05 {
					influence.Influences[faction.ID] = influenceLevel
				}
			}
		}

		// Calculate stability based on competing influences
		influence.Stability = tg.calculateStability(influence.Influences)
		influences[i] = influence
	}

	return influences
}

// calculateFactionInfluence computes a faction's influence on a territory.
func (tg *TerritoryGenerator) calculateFactionInfluence(
	territory *Territory,
	faction *Faction,
	allTerritories []*Territory,
) float64 {
	// Find closest territory controlled by this faction
	minDistance := math.MaxInt32
	for _, t := range allTerritories {
		if t.ControllerID == faction.ID {
			dist := tg.distance(territory.Position, t.Position)
			if dist < minDistance {
				minDistance = dist
			}
		}
	}

	if minDistance == math.MaxInt32 {
		return 0.0
	}

	// Influence decays with distance, boosted by faction power
	distanceFactor := 1.0 / (1.0 + float64(minDistance)*0.1)
	powerFactor := float64(faction.Influence) / 20.0 // Normalize to 0-0.5 range
	return distanceFactor * (0.5 + powerFactor)
}

// calculateStability computes territory stability from competing influences.
func (tg *TerritoryGenerator) calculateStability(influences map[string]float64) float64 {
	if len(influences) <= 1 {
		return 1.0
	}

	// Find top two influences
	var first, second float64
	for _, inf := range influences {
		if inf > first {
			second = first
			first = inf
		} else if inf > second {
			second = inf
		}
	}

	// Stability based on margin between top influences
	if first == 0 {
		return 1.0
	}
	margin := (first - second) / first
	return 0.5 + (margin * 0.5)
}

// generateTerritoryID creates a unique territory identifier.
func generateTerritoryID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 12)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}
	return "terr_" + string(id)
}
