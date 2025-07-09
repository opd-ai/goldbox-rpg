package items

import (
	"context"
	"fmt"
	"math/rand"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// TemplateBasedGenerator generates items using template system
type TemplateBasedGenerator struct {
	version   string
	templates map[string]*pcg.ItemTemplate
	registry  *ItemTemplateRegistry
	enchants  *EnchantmentSystem
	rng       *rand.Rand
}

// NewTemplateBasedGenerator creates a new template-based item generator
func NewTemplateBasedGenerator() *TemplateBasedGenerator {
	tbg := &TemplateBasedGenerator{
		version:   "1.0.0",
		templates: make(map[string]*pcg.ItemTemplate),
		registry:  NewItemTemplateRegistry(),
		enchants:  NewEnchantmentSystem(),
	}

	// Load default templates
	if err := tbg.registry.LoadDefaultTemplates(); err != nil {
		// Log error but continue - this is handled in actual usage
	}

	return tbg
}

// SetSeed sets the random seed for deterministic generation
func (tbg *TemplateBasedGenerator) SetSeed(seed int64) {
	tbg.rng = rand.New(rand.NewSource(seed))
	tbg.enchants.SetSeed(seed + 1) // Offset for enchantment system
}

// LoadTemplates loads item templates from YAML configuration
func (tbg *TemplateBasedGenerator) LoadTemplates(configPath string) error {
	return tbg.registry.LoadFromFile(configPath)
}

// Generate implements the Generator interface
func (tbg *TemplateBasedGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
	// Initialize RNG if not set
	if tbg.rng == nil {
		tbg.SetSeed(params.Seed)
	}

	// Create default item parameters
	itemParams := pcg.ItemParams{
		GenerationParams: params,
		MinRarity:        pcg.RarityCommon,
		MaxRarity:        pcg.RarityRare,
		ItemTypes:        []string{"sword", "bow", "armor", "potion"},
		EnchantmentRate:  0.3,
		UniqueChance:     0.1,
		LevelScaling:     true,
	}

	// Generate random item with constraints
	rarity := tbg.selectRandomRarity(itemParams.MinRarity, itemParams.MaxRarity)
	itemType := itemParams.ItemTypes[tbg.rng.Intn(len(itemParams.ItemTypes))]

	template, err := tbg.registry.GetTemplate(itemType, rarity)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	item, err := tbg.GenerateItem(ctx, *template, itemParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate item: %w", err)
	}

	return item, nil
}

// GenerateItem creates a single item from template
func (tbg *TemplateBasedGenerator) GenerateItem(ctx context.Context, template pcg.ItemTemplate, params pcg.ItemParams) (*game.Item, error) {
	if tbg.rng == nil {
		return nil, fmt.Errorf("random generator not initialized")
	}

	// Select rarity from params range
	rarity := tbg.selectRandomRarity(params.MinRarity, params.MaxRarity)

	// Create base item
	item := &game.Item{
		ID:   generateItemID(),
		Type: template.BaseType,
	}

	// Generate procedural name
	item.Name = GenerateItemName(&template, rarity, tbg.rng)

	// Roll stats within template ranges
	if err := tbg.applyStatRanges(item, template.StatRanges, params.PlayerLevel); err != nil {
		return nil, fmt.Errorf("failed to apply stat ranges: %w", err)
	}

	// Set base properties
	item.Properties = make([]string, len(template.Properties))
	copy(item.Properties, template.Properties)

	// Apply level scaling and rarity modifications
	if err := tbg.applyRarityModifications(item, rarity, &template); err != nil {
		return nil, fmt.Errorf("failed to apply rarity modifications: %w", err)
	}

	// Add enchantments based on rarity and rate
	if tbg.rng.Float64() < params.EnchantmentRate {
		if err := tbg.enchants.ApplyEnchantments(item, rarity, params.PlayerLevel, tbg.rng); err != nil {
			return nil, fmt.Errorf("failed to apply enchantments: %w", err)
		}
	}

	// Set appropriate value and weight
	tbg.calculateValueAndWeight(item, &template, rarity)

	return item, nil
}

// GenerateItemSet creates a collection of related items
func (tbg *TemplateBasedGenerator) GenerateItemSet(ctx context.Context, setType pcg.ItemSetType, params pcg.ItemParams) ([]*game.Item, error) {
	if tbg.rng == nil {
		return nil, fmt.Errorf("random generator not initialized")
	}

	quantity := tbg.getDefaultSetSize(setType)

	items := make([]*game.Item, 0, quantity)
	itemTypes := tbg.getItemTypesForSet(setType)

	for i := 0; i < quantity; i++ {
		// Select item type for this set
		itemType := itemTypes[tbg.rng.Intn(len(itemTypes))]

		// Select rarity from range
		rarity := tbg.selectRandomRarity(params.MinRarity, params.MaxRarity)

		// Get template and generate item
		template, err := tbg.registry.GetTemplate(itemType, rarity)
		if err != nil {
			continue // Skip items we can't generate
		}

		item, err := tbg.GenerateItem(ctx, *template, params)
		if err != nil {
			continue // Skip failed items
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("failed to generate any items for set type: %s", setType)
	}

	return items, nil
}

// applyStatRanges applies template stat ranges to item
func (tbg *TemplateBasedGenerator) applyStatRanges(item *game.Item, ranges map[string]pcg.StatRange, playerLevel int) error {
	for statName, statRange := range ranges {
		// Calculate base value within range
		baseValue := statRange.Min + tbg.rng.Intn(statRange.Max-statRange.Min+1)

		// Apply level scaling
		scaledValue := int(float64(baseValue) * (1.0 + statRange.Scaling*float64(playerLevel)))

		// Apply stat to item based on stat name
		switch statName {
		case "damage":
			item.Damage = fmt.Sprintf("1d%d", scaledValue)
		case "ac":
			item.AC = scaledValue
		case "weight":
			item.Weight = scaledValue
		case "value":
			item.Value = scaledValue
		default:
			// Add as property for unknown stats
			item.Properties = append(item.Properties, fmt.Sprintf("%s:%d", statName, scaledValue))
		}
	}

	return nil
}

// applyRarityModifications applies rarity-based modifications to item
func (tbg *TemplateBasedGenerator) applyRarityModifications(item *game.Item, rarity pcg.RarityTier, template *pcg.ItemTemplate) error {
	modifier := tbg.registry.GetRarityModifier(rarity)

	// Apply stat multipliers
	if item.AC > 0 {
		item.AC = int(float64(item.AC) * modifier.StatMultiplier)
	}

	// Parse and modify damage
	if item.Damage != "" {
		if newDamage := tbg.scaleDamage(item.Damage, modifier.StatMultiplier); newDamage != "" {
			item.Damage = newDamage
		}
	}

	return nil
}

// calculateValueAndWeight sets final value and weight based on rarity
func (tbg *TemplateBasedGenerator) calculateValueAndWeight(item *game.Item, template *pcg.ItemTemplate, rarity pcg.RarityTier) {
	modifier := tbg.registry.GetRarityModifier(rarity)

	// Base value calculation if not set by stat ranges
	if item.Value <= 0 {
		item.Value = 10 // Default base value
	}

	// Apply rarity value multiplier
	item.Value = int(float64(item.Value) * modifier.ValueMultiplier)

	// Base weight calculation if not set by stat ranges
	if item.Weight <= 0 {
		switch item.Type {
		case "weapon":
			item.Weight = 3
		case "armor":
			item.Weight = 10
		case "consumable":
			item.Weight = 1
		default:
			item.Weight = 2
		}
	}
}

// scaleDamage scales damage strings by multiplier
func (tbg *TemplateBasedGenerator) scaleDamage(damage string, multiplier float64) string {
	// Simple scaling - could be more sophisticated
	if multiplier > 1.0 {
		return damage + "+1" // Add bonus damage for higher rarities
	}
	return damage
}

// getDefaultSetSize returns default number of items for set type
func (tbg *TemplateBasedGenerator) getDefaultSetSize(setType pcg.ItemSetType) int {
	switch setType {
	case pcg.ItemSetArmor:
		return 3 // helmet, armor, boots
	case pcg.ItemSetWeapons:
		return 2 // primary and secondary weapon
	case pcg.ItemSetJewelry:
		return 2 // ring and amulet
	case pcg.ItemSetConsumab:
		return 5 // various potions
	default:
		return 3
	}
}

// getItemTypesForSet returns appropriate item types for set
func (tbg *TemplateBasedGenerator) getItemTypesForSet(setType pcg.ItemSetType) []string {
	switch setType {
	case pcg.ItemSetArmor:
		return []string{"armor", "shield", "helmet"}
	case pcg.ItemSetWeapons:
		return []string{"sword", "bow", "dagger", "staff"}
	case pcg.ItemSetJewelry:
		return []string{"ring", "amulet", "bracelet"}
	case pcg.ItemSetConsumab:
		return []string{"potion", "scroll", "elixir"}
	case pcg.ItemSetTools:
		return []string{"tool", "kit", "instrument"}
	default:
		return []string{"misc"}
	}
}

// GetType returns the content type this generator produces
func (tbg *TemplateBasedGenerator) GetType() pcg.ContentType {
	return pcg.ContentTypeItems
}

// GetVersion returns the generator version for compatibility checking
func (tbg *TemplateBasedGenerator) GetVersion() string {
	return tbg.version
}

// Validate checks if the provided parameters are valid for this generator
func (tbg *TemplateBasedGenerator) Validate(params pcg.GenerationParams) error {
	if params.Seed == 0 {
		return fmt.Errorf("seed cannot be zero")
	}

	if params.PlayerLevel < 1 || params.PlayerLevel > 20 {
		return fmt.Errorf("player level must be between 1 and 20")
	}

	if params.Difficulty < 1 || params.Difficulty > 20 {
		return fmt.Errorf("difficulty must be between 1 and 20")
	}

	return nil
}

// generateItemID creates a unique identifier for items
func generateItemID() string {
	return fmt.Sprintf("item_%d", rand.Int63())
}

// selectRandomRarity selects a random rarity within the given range
func (tbg *TemplateBasedGenerator) selectRandomRarity(minRarity, maxRarity pcg.RarityTier) pcg.RarityTier {
	rarities := []pcg.RarityTier{
		pcg.RarityCommon,
		pcg.RarityUncommon,
		pcg.RarityRare,
		pcg.RarityEpic,
		pcg.RarityLegendary,
		pcg.RarityArtifact,
	}

	// Find indices of min and max rarities
	minIndex := 0
	maxIndex := len(rarities) - 1

	for i, rarity := range rarities {
		if rarity == minRarity {
			minIndex = i
		}
		if rarity == maxRarity {
			maxIndex = i
		}
	}

	// Ensure valid range
	if maxIndex < minIndex {
		maxIndex = minIndex
	}

	// Select random rarity within range
	selectedIndex := minIndex + tbg.rng.Intn(maxIndex-minIndex+1)
	return rarities[selectedIndex]
}
