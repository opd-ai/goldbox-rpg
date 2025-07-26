package pcg

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// FactionGenerator creates political entities, relationships, and conflicts
// Generates cohesive faction systems that influence world politics and economics
type FactionGenerator struct {
	version string
	logger  *logrus.Logger
	rng     *rand.Rand
}

// NewFactionGenerator creates a new faction generator instance
func NewFactionGenerator(logger *logrus.Logger) *FactionGenerator {
	if logger == nil {
		logger = logrus.New()
	}

	return &FactionGenerator{
		version: "1.0.0",
		logger:  logger,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Generate creates faction systems based on the provided parameters
// Returns GeneratedFactionSystem containing all political entities and relationships
func (fg *FactionGenerator) Generate(ctx context.Context, params GenerationParams) (interface{}, error) {
	if err := fg.Validate(params); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Use seed for deterministic generation
	rng := rand.New(rand.NewSource(params.Seed))
	fg.rng = rng

	factionParams, ok := params.Constraints["faction_params"].(FactionParams)
	if !ok {
		// Use default parameters
		factionParams = FactionParams{
			GenerationParams: params,
			FactionCount:     rng.Intn(8) + 3, // 3-10 factions
			MinPower:         1,
			MaxPower:         10,
			ConflictLevel:    0.3,
			EconomicFocus:    0.5,
			MilitaryFocus:    0.4,
			CulturalFocus:    0.3,
		}
	}

	fg.logger.WithFields(logrus.Fields{
		"seed":           params.Seed,
		"faction_count":  factionParams.FactionCount,
		"conflict_level": factionParams.ConflictLevel,
	}).Info("generating faction system")

	start := time.Now()

	// Generate the faction system
	factionSystem, err := fg.generateFactionSystem(ctx, factionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate faction system: %w", err)
	}

	duration := time.Since(start)
	fg.logger.WithFields(logrus.Fields{
		"duration":      duration,
		"factions":      len(factionSystem.Factions),
		"relationships": len(factionSystem.Relationships),
	}).Info("faction system generation completed")

	return factionSystem, nil
}

// GetType returns the content type for faction generation
func (fg *FactionGenerator) GetType() ContentType {
	return ContentTypeFactions
}

// GetVersion returns the generator version
func (fg *FactionGenerator) GetVersion() string {
	return fg.version
}

// Validate checks if the provided parameters are valid for faction generation
func (fg *FactionGenerator) Validate(params GenerationParams) error {
	if params.Seed == 0 {
		return fmt.Errorf("seed cannot be zero")
	}

	if params.Difficulty < 1 || params.Difficulty > 20 {
		return fmt.Errorf("difficulty must be between 1 and 20, got %d", params.Difficulty)
	}

	// Validate faction-specific constraints if provided
	if factionConstraints, exists := params.Constraints["faction_params"]; exists {
		if factionParams, ok := factionConstraints.(FactionParams); ok {
			if factionParams.FactionCount < 1 || factionParams.FactionCount > 20 {
				return fmt.Errorf("faction count must be between 1 and 20, got %d", factionParams.FactionCount)
			}

			if factionParams.ConflictLevel < 0.0 || factionParams.ConflictLevel > 1.0 {
				return fmt.Errorf("conflict level must be between 0.0 and 1.0, got %f", factionParams.ConflictLevel)
			}
		}
	}

	return nil
}

// generateFactionSystem creates a complete faction system with relationships
func (fg *FactionGenerator) generateFactionSystem(ctx context.Context, params FactionParams) (*GeneratedFactionSystem, error) {
	system := &GeneratedFactionSystem{
		ID:            fg.generateID("faction_system"),
		Name:          fg.generateSystemName(),
		Factions:      make([]*Faction, 0, params.FactionCount),
		Relationships: make([]*FactionRelationship, 0),
		Territories:   make([]*Territory, 0),
		TradeDeals:    make([]*TradeDeal, 0),
		Conflicts:     make([]*Conflict, 0),
		Metadata:      make(map[string]interface{}),
		Generated:     time.Now(),
	}

	// Generate individual factions
	for i := 0; i < params.FactionCount; i++ {
		faction, err := fg.generateFaction(ctx, i, params)
		if err != nil {
			return nil, fmt.Errorf("failed to generate faction %d: %w", i, err)
		}
		system.Factions = append(system.Factions, faction)
	}

	// Generate relationships between factions
	if err := fg.generateRelationships(ctx, system, params); err != nil {
		return nil, fmt.Errorf("failed to generate relationships: %w", err)
	}

	// Generate territories and assign to factions
	if err := fg.generateTerritories(ctx, system, params); err != nil {
		return nil, fmt.Errorf("failed to generate territories: %w", err)
	}

	// Generate trade deals based on economic interests
	if err := fg.generateTradeDeals(ctx, system, params); err != nil {
		return nil, fmt.Errorf("failed to generate trade deals: %w", err)
	}

	// Generate active conflicts based on relationships
	if err := fg.generateConflicts(ctx, system, params); err != nil {
		return nil, fmt.Errorf("failed to generate conflicts: %w", err)
	}

	// Add metadata about the generation
	system.Metadata["generation_seed"] = params.Seed
	system.Metadata["conflict_level"] = params.ConflictLevel
	system.Metadata["economic_focus"] = params.EconomicFocus
	system.Metadata["military_focus"] = params.MilitaryFocus

	return system, nil
}

// generateFaction creates a single faction with characteristics and goals
func (fg *FactionGenerator) generateFaction(ctx context.Context, index int, params FactionParams) (*Faction, error) {
	faction := &Faction{
		ID:         fg.generateID("faction"),
		Name:       fg.generateFactionName(),
		Type:       fg.selectFactionType(),
		Government: fg.selectGovernmentType(),
		Ideology:   fg.generateIdeology(),
		Power:      fg.rng.Intn(params.MaxPower-params.MinPower+1) + params.MinPower,
		Wealth:     fg.rng.Intn(params.MaxPower-params.MinPower+1) + params.MinPower,
		Military:   fg.rng.Intn(params.MaxPower-params.MinPower+1) + params.MinPower,
		Influence:  fg.rng.Intn(params.MaxPower-params.MinPower+1) + params.MinPower,
		Stability:  fg.rng.Float64(),
		Goals:      fg.generateFactionGoals(),
		Resources:  fg.generateControlledResources(),
		Leaders:    fg.generateLeadership(),
		Properties: make(map[string]interface{}),
	}

	// Add faction-specific properties based on type
	fg.addFactionTypeProperties(faction)

	// Adjust attributes based on faction focus areas
	fg.adjustFactionAttributes(faction, params)

	fg.logger.WithFields(logrus.Fields{
		"faction_id":   faction.ID,
		"faction_name": faction.Name,
		"faction_type": faction.Type,
		"power":        faction.Power,
	}).Debug("generated faction")

	return faction, nil
}

// generateRelationships creates diplomatic relationships between all factions
func (fg *FactionGenerator) generateRelationships(ctx context.Context, system *GeneratedFactionSystem, params FactionParams) error {
	factions := system.Factions

	for i := 0; i < len(factions); i++ {
		for j := i + 1; j < len(factions); j++ {
			relationship := fg.generateRelationship(factions[i], factions[j], params)
			system.Relationships = append(system.Relationships, relationship)
		}
	}

	// Ensure relationship consistency (if A likes B, B should have some opinion of A)
	fg.balanceRelationships(system.Relationships)

	return nil
}

// generateRelationship creates a diplomatic relationship between two factions
func (fg *FactionGenerator) generateRelationship(faction1, faction2 *Faction, params FactionParams) *FactionRelationship {
	// Base relationship influenced by faction types and ideologies
	baseRelation := fg.calculateBaseRelation(faction1, faction2)

	// Add random variation
	variation := (fg.rng.Float64() - 0.5) * 0.4 // ±0.2 variation
	finalRelation := math.Max(-1.0, math.Min(1.0, baseRelation+variation))

	status := fg.determineRelationshipStatus(finalRelation)

	return &FactionRelationship{
		ID:          fg.generateID("relationship"),
		Faction1ID:  faction1.ID,
		Faction2ID:  faction2.ID,
		Status:      status,
		Opinion:     finalRelation,
		TrustLevel:  fg.rng.Float64(),
		TradeLevel:  fg.calculateTradeLevel(finalRelation, faction1, faction2),
		Hostility:   fg.calculateHostility(finalRelation, params.ConflictLevel),
		History:     fg.generateRelationshipHistory(),
		LastChanged: time.Now().AddDate(0, 0, -fg.rng.Intn(365)),
		Properties:  make(map[string]interface{}),
	}
}

// Helper methods for faction generation

func (fg *FactionGenerator) generateSystemName() string {
	prefixes := []string{"The", "Greater", "United", "Free", "Allied", "Independent"}
	middles := []string{"Kingdoms", "Realms", "States", "Territories", "Provinces", "Lands"}
	suffixes := []string{"Coalition", "Alliance", "Federation", "Union", "League", "Assembly"}

	if fg.rng.Float32() < 0.3 {
		return fmt.Sprintf("%s %s",
			prefixes[fg.rng.Intn(len(prefixes))],
			middles[fg.rng.Intn(len(middles))])
	}

	return fmt.Sprintf("%s %s %s",
		prefixes[fg.rng.Intn(len(prefixes))],
		middles[fg.rng.Intn(len(middles))],
		suffixes[fg.rng.Intn(len(suffixes))])
}

func (fg *FactionGenerator) generateFactionName() string {
	prefixes := []string{"Order of", "House", "Clan", "Guild of", "Brotherhood of", "Circle of", "Council of"}
	names := []string{"Iron", "Gold", "Silver", "Storm", "Shadow", "Light", "Fire", "Stone", "Steel", "Crystal"}
	suffixes := []string{"Keepers", "Guardians", "Warriors", "Merchants", "Scholars", "Mages", "Knights", "Rangers"}

	prefix := prefixes[fg.rng.Intn(len(prefixes))]
	name := names[fg.rng.Intn(len(names))]
	suffix := suffixes[fg.rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s %s", prefix, name, suffix)
}

func (fg *FactionGenerator) selectFactionType() FactionType {
	types := []FactionType{
		FactionTypeMilitary, FactionTypeEconomic, FactionTypeReligious,
		FactionTypeCriminal, FactionTypeScholarly, FactionTypePolitical,
		FactionTypeMercenary, FactionTypeMagical,
	}
	return types[fg.rng.Intn(len(types))]
}

func (fg *FactionGenerator) selectGovernmentType() GovernmentType {
	types := []GovernmentType{
		GovernmentMonarchy, GovernmentRepublic, GovernmentTheocracy,
		GovernmentMilitary, GovernmentTribal, GovernmentAnarchy,
	}
	return types[fg.rng.Intn(len(types))]
}

func (fg *FactionGenerator) generateIdeology() string {
	ideologies := []string{
		"Expansionist", "Isolationist", "Traditionalist", "Progressive",
		"Militaristic", "Pacifist", "Mercantile", "Religious",
		"Scholarly", "Pragmatic", "Idealistic", "Nationalist",
	}
	return ideologies[fg.rng.Intn(len(ideologies))]
}

func (fg *FactionGenerator) generateFactionGoals() []string {
	possibleGoals := []string{
		"territorial_expansion", "economic_dominance", "religious_conversion",
		"knowledge_acquisition", "political_influence", "military_supremacy",
		"trade_monopoly", "cultural_preservation", "magical_research",
		"resource_control", "alliance_building", "enemy_destruction",
	}

	goalCount := fg.rng.Intn(3) + 2 // 2-4 goals
	goals := make([]string, 0, goalCount)

	// Select unique goals
	selected := make(map[int]bool)
	for len(goals) < goalCount {
		index := fg.rng.Intn(len(possibleGoals))
		if !selected[index] {
			goals = append(goals, possibleGoals[index])
			selected[index] = true
		}
	}

	return goals
}

func (fg *FactionGenerator) generateControlledResources() []ResourceType {
	resources := []ResourceType{
		ResourceGold, ResourceIron, ResourceFood,
		ResourceWood, ResourceStone, ResourceMagicite,
	}

	resourceCount := fg.rng.Intn(4) + 1 // 1-4 resources
	controlled := make([]ResourceType, 0, resourceCount)

	selected := make(map[int]bool)
	for len(controlled) < resourceCount {
		index := fg.rng.Intn(len(resources))
		if !selected[index] {
			controlled = append(controlled, resources[index])
			selected[index] = true
		}
	}

	return controlled
}

func (fg *FactionGenerator) generateLeadership() []*FactionLeader {
	leaderCount := fg.rng.Intn(3) + 1 // 1-3 leaders
	leaders := make([]*FactionLeader, 0, leaderCount)

	titles := []string{"Lord", "Commander", "High Priest", "Guildmaster", "Archmage", "Chief", "President", "General"}
	traits := []string{"Ambitious", "Cunning", "Wise", "Ruthless", "Charismatic", "Strategic", "Diplomatic", "Fierce"}

	for i := 0; i < leaderCount; i++ {
		leader := &FactionLeader{
			ID:         fg.generateID("leader"),
			Name:       fg.generateLeaderName(),
			Title:      titles[fg.rng.Intn(len(titles))],
			Age:        fg.rng.Intn(40) + 25, // 25-65 years old
			Traits:     []string{traits[fg.rng.Intn(len(traits))]},
			Loyalty:    fg.rng.Float64(),
			Competence: fg.rng.Float64(),
			Influence:  fg.rng.Float64(),
			Properties: make(map[string]interface{}),
		}
		leaders = append(leaders, leader)
	}

	return leaders
}

func (fg *FactionGenerator) generateLeaderName() string {
	firstNames := []string{"Aldric", "Vera", "Marcus", "Elena", "Gareth", "Lyra", "Theron", "Mira"}
	lastNames := []string{"Ironforge", "Goldweaver", "Stormwind", "Shadowbane", "Lightbringer", "Darkbane"}

	return fmt.Sprintf("%s %s",
		firstNames[fg.rng.Intn(len(firstNames))],
		lastNames[fg.rng.Intn(len(lastNames))])
}

// generateID creates a unique identifier with a prefix
func (fg *FactionGenerator) generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, fg.rng.Int63())
}

// Helper methods for generating missing functionality

func (fg *FactionGenerator) addFactionTypeProperties(faction *Faction) {
	// Add type-specific properties based on faction type
	switch faction.Type {
	case FactionTypeMilitary:
		faction.Properties["unit_types"] = []string{"infantry", "cavalry", "archers"}
		faction.Military += 2 // Military factions get bonus military power
	case FactionTypeEconomic:
		faction.Properties["trade_routes"] = fg.rng.Intn(5) + 3
		faction.Wealth += 2 // Economic factions get bonus wealth
	case FactionTypeReligious:
		faction.Properties["temples"] = fg.rng.Intn(3) + 1
		faction.Influence += 2 // Religious factions get bonus influence
	case FactionTypeMagical:
		faction.Properties["mage_towers"] = fg.rng.Intn(2) + 1
		faction.Properties["magical_knowledge"] = fg.rng.Intn(100) + 50
	}
}

func (fg *FactionGenerator) adjustFactionAttributes(faction *Faction, params FactionParams) {
	// Adjust attributes based on focus parameters
	if params.MilitaryFocus > 0.5 {
		faction.Military = int(float64(faction.Military) * (1.0 + params.MilitaryFocus))
	}
	if params.EconomicFocus > 0.5 {
		faction.Wealth = int(float64(faction.Wealth) * (1.0 + params.EconomicFocus))
	}
}

func (fg *FactionGenerator) calculateBaseRelation(faction1, faction2 *Faction) float64 {
	// Calculate base relationship based on faction characteristics
	relation := 0.0

	// Same government types tend to get along better
	if faction1.Government == faction2.Government {
		relation += 0.2
	}

	// Same ideology creates affinity
	if faction1.Ideology == faction2.Ideology {
		relation += 0.3
	}

	// Military factions tend to be suspicious of each other
	if faction1.Type == FactionTypeMilitary && faction2.Type == FactionTypeMilitary {
		relation -= 0.2
	}

	// Economic factions often cooperate
	if faction1.Type == FactionTypeEconomic && faction2.Type == FactionTypeEconomic {
		relation += 0.1
	}

	// Criminal factions are generally disliked
	if faction1.Type == FactionTypeCriminal || faction2.Type == FactionTypeCriminal {
		relation -= 0.3
	}

	return relation
}

func (fg *FactionGenerator) determineRelationshipStatus(opinion float64) RelationshipStatus {
	if opinion >= 0.7 {
		return RelationStatusAllied
	} else if opinion >= 0.3 {
		return RelationStatusFriendly
	} else if opinion >= -0.3 {
		return RelationStatusNeutral
	} else if opinion >= -0.7 {
		return RelationStatusTense
	} else if opinion >= -0.9 {
		return RelationStatusHostile
	}
	return RelationStatusWar
}

func (fg *FactionGenerator) calculateTradeLevel(opinion float64, faction1, faction2 *Faction) float64 {
	baseTradeLevel := (opinion + 1.0) / 2.0 // Convert -1:1 to 0:1

	// Economic factions trade more
	if faction1.Type == FactionTypeEconomic || faction2.Type == FactionTypeEconomic {
		baseTradeLevel *= 1.3
	}

	// Criminal factions trade less legitimately
	if faction1.Type == FactionTypeCriminal || faction2.Type == FactionTypeCriminal {
		baseTradeLevel *= 0.5
	}

	return math.Min(1.0, baseTradeLevel)
}

func (fg *FactionGenerator) calculateHostility(opinion float64, conflictLevel float64) float64 {
	// Convert opinion to hostility (inverse relationship)
	hostility := (1.0 - opinion) / 2.0

	// Scale by overall conflict level
	hostility *= conflictLevel

	return math.Max(0.0, math.Min(1.0, hostility))
}

func (fg *FactionGenerator) generateRelationshipHistory() []string {
	historyEvents := []string{
		"trade_agreement", "border_dispute", "alliance_formation", "betrayal",
		"military_cooperation", "diplomatic_incident", "succession_crisis",
		"resource_conflict", "territorial_exchange", "marriage_alliance",
	}

	eventCount := fg.rng.Intn(3) + 1 // 1-3 historical events
	history := make([]string, 0, eventCount)

	for i := 0; i < eventCount; i++ {
		event := historyEvents[fg.rng.Intn(len(historyEvents))]
		history = append(history, event)
	}

	return history
}

func (fg *FactionGenerator) balanceRelationships(relationships []*FactionRelationship) {
	// Ensure relationship consistency and realistic political dynamics
	for _, rel := range relationships {
		// Add some mutual opinion influence (relationships aren't perfectly one-sided)
		if fg.rng.Float64() < 0.3 {
			// Small chance for asymmetric relationships
			variation := (fg.rng.Float64() - 0.5) * 0.2
			rel.Opinion = math.Max(-1.0, math.Min(1.0, rel.Opinion+variation))
		}
	}
}

// Placeholder methods for complete implementation
func (fg *FactionGenerator) generateTerritories(ctx context.Context, system *GeneratedFactionSystem, params FactionParams) error {
	// TODO: Implement territory generation based on faction power and world geography
	// For now, create basic territories for each faction
	for _, faction := range system.Factions {
		territoryCount := fg.rng.Intn(3) + 1 // 1-3 territories per faction
		for j := 0; j < territoryCount; j++ {
			territory := &Territory{
				ID:           fg.generateID("territory"),
				Name:         fmt.Sprintf("%s Territory %d", faction.Name, j+1),
				Type:         fg.selectTerritoryType(),
				ControllerID: faction.ID,
				Position:     game.Position{X: fg.rng.Intn(100), Y: fg.rng.Intn(100)},
				Size:         fg.rng.Intn(50) + 10,
				Population:   fg.rng.Intn(10000) + 1000,
				Defenses:     faction.Military + fg.rng.Intn(5),
				Resources:    faction.Resources,
				Strategic:    fg.rng.Float64() < 0.3,
				Properties:   make(map[string]interface{}),
			}
			system.Territories = append(system.Territories, territory)
		}
	}
	return nil
}

func (fg *FactionGenerator) generateTradeDeals(ctx context.Context, system *GeneratedFactionSystem, params FactionParams) error {
	// Generate trade deals between friendly factions
	for _, rel := range system.Relationships {
		if rel.TradeLevel > 0.5 && fg.rng.Float64() < params.EconomicFocus {
			deal := &TradeDeal{
				ID:         fg.generateID("trade"),
				Name:       fmt.Sprintf("Trade Agreement %s", fg.generateID("agreement")),
				Faction1ID: rel.Faction1ID,
				Faction2ID: rel.Faction2ID,
				Resource1:  fg.selectTradeResource(),
				Resource2:  fg.selectTradeResource(),
				Volume1:    fg.rng.Intn(1000) + 100,
				Volume2:    fg.rng.Intn(1000) + 100,
				Duration:   fg.rng.Intn(365) + 90, // 90-455 days
				Profit1:    fg.rng.Intn(500) + 50,
				Profit2:    fg.rng.Intn(500) + 50,
				Active:     true,
				Properties: make(map[string]interface{}),
			}
			system.TradeDeals = append(system.TradeDeals, deal)
		}
	}
	return nil
}

func (fg *FactionGenerator) generateConflicts(ctx context.Context, system *GeneratedFactionSystem, params FactionParams) error {
	// Generate active conflicts based on hostile relationships
	for _, rel := range system.Relationships {
		if rel.Hostility > 0.7 && fg.rng.Float64() < params.ConflictLevel {
			conflict := &Conflict{
				ID:         fg.generateID("conflict"),
				Name:       fmt.Sprintf("Conflict of %s", fg.generateConflictName()),
				Type:       fg.selectConflictType(),
				Factions:   []string{rel.Faction1ID, rel.Faction2ID},
				Cause:      fg.generateConflictCause(),
				Intensity:  rel.Hostility,
				Duration:   fg.rng.Intn(180) + 30, // 30-210 days
				Territory:  "", // Would be populated if territory system is more advanced
				Resolution: "",
				Active:     true,
				Properties: make(map[string]interface{}),
			}
			system.Conflicts = append(system.Conflicts, conflict)
		}
	}
	return nil
}

// Additional helper methods

func (fg *FactionGenerator) selectTerritoryType() TerritoryType {
	types := []TerritoryType{
		TerritoryTypeCapital, TerritoryTypeCity, TerritoryTypeOutpost,
		TerritoryTypeFortress, TerritoryTypeTradingPost, TerritoryTypeResource,
	}
	return types[fg.rng.Intn(len(types))]
}

func (fg *FactionGenerator) selectTradeResource() ResourceType {
	resources := []ResourceType{
		ResourceIron, ResourceGold, ResourceGems, ResourceWood,
		ResourceStone, ResourceMagicite, ResourceFood, ResourceWater,
	}
	return resources[fg.rng.Intn(len(resources))]
}

func (fg *FactionGenerator) selectConflictType() ConflictType {
	types := []ConflictType{
		ConflictTypeTrade, ConflictTypeTerritory, ConflictTypeReligious,
		ConflictTypeResource, ConflictTypeSuccession, ConflictTypeRevenge,
	}
	return types[fg.rng.Intn(len(types))]
}

func (fg *FactionGenerator) generateConflictName() string {
	names := []string{
		"the Iron Crown", "Disputed Territories", "Sacred Relics",
		"Trade Route Control", "Succession Rights", "Ancient Grievances",
		"Resource Claims", "Border Demarcation", "Religious Differences",
	}
	return names[fg.rng.Intn(len(names))]
}

func (fg *FactionGenerator) generateConflictCause() string {
	causes := []string{
		"territorial_expansion", "resource_competition", "succession_dispute",
		"trade_route_control", "religious_differences", "historical_grievance",
		"border_incident", "diplomatic_insult", "alliance_betrayal",
	}
	return causes[fg.rng.Intn(len(causes))]
}
