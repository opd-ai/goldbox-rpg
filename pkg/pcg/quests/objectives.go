package quests

import (
	"fmt"
	"math/rand"

	"goldbox-rpg/pkg/pcg"
)

// ObjectiveGenerator creates specific quest objectives.
// ObjectiveGenerator is stateless and does not require world state - it generates
// objectives using procedural generation contexts and hardcoded location/item pools.
// Future implementations may add world-aware methods that query actual game state.
type ObjectiveGenerator struct{}

// NewObjectiveGenerator creates an objective generator.
// The generator is stateless and can be safely reused across multiple generation calls.
func NewObjectiveGenerator() *ObjectiveGenerator {
	return &ObjectiveGenerator{}
}

// validateGenerationContext validates that the generation context is not nil.
func validateGenerationContext(genCtx *pcg.GenerationContext) error {
	if genCtx == nil {
		return fmt.Errorf("generation context cannot be nil")
	}
	return nil
}

// validateContextAndSelectType validates context, checks value range, selects from types, and returns RNG.
// Returns the selected type string and the RNG from the context.
func validateContextAndSelectType(genCtx *pcg.GenerationContext, value, min, max int, paramName string, typeSelector func(int) []string) (string, error) {
	if err := validateGenerationContext(genCtx); err != nil {
		return "", err
	}

	if value < min || value > max {
		return "", fmt.Errorf("%s must be between %d and %d, got %d", paramName, min, max, value)
	}

	types := typeSelector(value)
	if len(types) == 0 {
		return "", fmt.Errorf("no types available for %s %d", paramName, value)
	}

	rng := genCtx.RNG
	selectedType := types[rng.Intn(len(types))]
	return selectedType, nil
}

// selectRandomLocation selects a random location from available locations.
func (og *ObjectiveGenerator) selectRandomLocation(rng *rand.Rand, minLocations int) (string, error) {
	locations := og.getAvailableLocations()
	if len(locations) < minLocations {
		return "", fmt.Errorf("need at least %d location(s), got %d", minLocations, len(locations))
	}
	return locations[rng.Intn(len(locations))], nil
}

// GenerateKillObjective creates kill/defeat objectives
func (og *ObjectiveGenerator) GenerateKillObjective(difficulty int, genCtx *pcg.GenerationContext) (*pcg.QuestObjective, error) {
	enemyType, err := validateContextAndSelectType(genCtx, difficulty, 1, 10, "difficulty", og.selectEnemyTypesForDifficulty)
	if err != nil {
		return nil, err
	}

	rng := genCtx.RNG

	// Determine quantity based on challenge rating
	minQuantity := max(1, difficulty/2)
	maxQuantity := difficulty + 2
	quantity := minQuantity + rng.Intn(maxQuantity-minQuantity+1)

	// Choose location from available areas
	location, err := og.selectRandomLocation(rng, 1)
	if err != nil {
		return nil, fmt.Errorf("no locations available for kill objective")
	}

	objective := &pcg.QuestObjective{
		ID:          fmt.Sprintf("obj_%d", rng.Int63()),
		Type:        "kill",
		Description: fmt.Sprintf("Defeat %d %s in %s", quantity, enemyType, location),
		Target:      enemyType,
		Quantity:    quantity,
		Progress:    0,
		Complete:    false,
		Optional:    false,
		Conditions:  map[string]interface{}{"location": location},
	}

	return objective, nil
}

// GenerateFetchObjective creates item retrieval objectives
func (og *ObjectiveGenerator) GenerateFetchObjective(playerLevel int, genCtx *pcg.GenerationContext) (*pcg.QuestObjective, error) {
	itemType, err := validateContextAndSelectType(genCtx, playerLevel, 1, 20, "player level", og.selectItemTypesForLevel)
	if err != nil {
		return nil, err
	}

	rng := genCtx.RNG

	// Determine quantity (usually 1 for fetch quests, but can be more for common items)
	quantity := 1
	if isCommonItem(itemType) {
		quantity = 1 + rng.Intn(3) // 1-3 for common items
	}

	// Choose pickup and delivery locations
	pickupLocation, err := og.selectRandomLocation(rng, 2)
	if err != nil {
		return nil, fmt.Errorf("need at least 2 locations for fetch objective")
	}

	deliveryLocation, err := og.selectRandomLocation(rng, 2)
	if err != nil {
		return nil, err
	}
	// Ensure different locations
	for deliveryLocation == pickupLocation {
		deliveryLocation, _ = og.selectRandomLocation(rng, 2)
	}

	objective := &pcg.QuestObjective{
		ID:          fmt.Sprintf("obj_%d", rng.Int63()),
		Type:        "fetch",
		Description: fmt.Sprintf("Retrieve %d %s from %s and deliver to %s", quantity, itemType, pickupLocation, deliveryLocation),
		Target:      itemType,
		Quantity:    quantity,
		Progress:    0,
		Complete:    false,
		Optional:    false,
		Conditions:  map[string]interface{}{"pickup": pickupLocation, "delivery": deliveryLocation},
	}

	return objective, nil
}

// GenerateExploreObjective creates exploration objectives.
// The objective targets procedurally selected areas from a hardcoded pool.
// Future implementations may accept world state to query actual unexplored areas.
func (og *ObjectiveGenerator) GenerateExploreObjective(genCtx *pcg.GenerationContext) (*pcg.QuestObjective, error) {
	if genCtx == nil {
		return nil, fmt.Errorf("generation context cannot be nil")
	}
	// Identify unexplored or partially explored areas
	unexploredAreas := og.getUnexploredAreas()
	if len(unexploredAreas) == 0 {
		return nil, fmt.Errorf("no unexplored areas available")
	}

	rng := genCtx.RNG
	targetArea := unexploredAreas[rng.Intn(len(unexploredAreas))]

	// Set discovery requirements (area percentage, landmarks)
	requiredPercentage := 70 + rng.Intn(31) // 70-100%

	// Add optional sub-objectives
	var subObjectives []string
	if rng.Float32() < 0.3 { // 30% chance for hidden areas
		subObjectives = append(subObjectives, "Find hidden areas")
	}
	if rng.Float32() < 0.2 { // 20% chance for secrets
		subObjectives = append(subObjectives, "Discover secrets")
	}

	description := fmt.Sprintf("Explore %d%% of %s", requiredPercentage, targetArea)
	if len(subObjectives) > 0 {
		description += fmt.Sprintf(" and %s", subObjectives[0])
	}

	objective := &pcg.QuestObjective{
		ID:          fmt.Sprintf("obj_%d", rng.Int63()),
		Type:        "explore",
		Description: description,
		Target:      targetArea,
		Quantity:    requiredPercentage,
		Progress:    0,
		Complete:    false,
		Optional:    false,
		Conditions:  map[string]interface{}{"area": targetArea, "percentage": requiredPercentage},
	}

	return objective, nil
}

// Helper functions

func (og *ObjectiveGenerator) selectEnemyTypesForDifficulty(difficulty int) []string {
	// Define enemy types by difficulty level
	enemyMap := map[int][]string{
		1:  {"Rat", "Goblin", "Skeleton"},
		2:  {"Orc", "Wolf", "Zombie"},
		3:  {"Hobgoblin", "Bear", "Ghoul"},
		4:  {"Ogre", "Wight", "Owlbear"},
		5:  {"Troll", "Wraith", "Manticore"},
		6:  {"Hill Giant", "Spectre", "Chimera"},
		7:  {"Stone Giant", "Vampire", "Roc"},
		8:  {"Fire Giant", "Lich", "Dragon"},
		9:  {"Storm Giant", "Demon", "Ancient Dragon"},
		10: {"Titan", "Archdevil", "Legendary Dragon"},
	}

	var enemies []string
	// Include enemies from current difficulty and slightly lower
	for level := max(1, difficulty-1); level <= difficulty; level++ {
		if levelEnemies, exists := enemyMap[level]; exists {
			enemies = append(enemies, levelEnemies...)
		}
	}

	return enemies
}

func (og *ObjectiveGenerator) selectItemTypesForLevel(playerLevel int) []string {
	// Define item types by level ranges
	if playerLevel <= 3 {
		return []string{"Iron Sword", "Leather Armor", "Health Potion", "Rope", "Torch"}
	} else if playerLevel <= 6 {
		return []string{"Steel Sword", "Chain Mail", "Greater Health Potion", "Magic Scroll", "Silver Amulet"}
	} else if playerLevel <= 10 {
		return []string{"Enchanted Sword", "Plate Armor", "Elixir", "Spell Component", "Gem"}
	} else if playerLevel <= 15 {
		return []string{"Magic Weapon", "Enchanted Armor", "Rare Potion", "Ancient Scroll", "Crystal"}
	} else {
		return []string{"Legendary Weapon", "Artifact Armor", "Divine Elixir", "Lost Tome", "Divine Relic"}
	}
}

// getAvailableLocations returns a pool of procedural location names.
// These hardcoded locations are used as a template pool for quest generation.
// The generator selects from this pool deterministically based on the RNG seed.
// For world-aware location queries, integrate with the game.World instance directly.
func (og *ObjectiveGenerator) getAvailableLocations() []string {
	return []string{
		"Dark Cave",
		"Abandoned Mine",
		"Ancient Ruins",
		"Haunted Forest",
		"Forgotten Temple",
		"Underground Cavern",
		"Crumbling Tower",
		"Mystic Grove",
		"Desolate Wasteland",
		"Hidden Valley",
	}
}

// getUnexploredAreas returns a pool of exploration area names.
// These hardcoded areas serve as procedural templates for explore quests.
// The generator selects from this pool deterministically based on the RNG seed.
// For actual world exploration state queries, integrate with game.World directly.
func (og *ObjectiveGenerator) getUnexploredAreas() []string {
	return []string{
		"Northern Wilderness",
		"Eastern Marshlands",
		"Southern Desert",
		"Western Mountains",
		"Deep Forest",
		"Underground Tunnels",
		"Floating Islands",
		"Shadow Realm",
	}
}

func isCommonItem(itemType string) bool {
	commonItems := map[string]bool{
		"Health Potion": true,
		"Rope":          true,
		"Torch":         true,
		"Rations":       true,
		"Arrow":         true,
	}
	return commonItems[itemType]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
