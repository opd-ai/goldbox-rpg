package items

import (
	"fmt"
	"math/rand"

	"goldbox-rpg/pkg/pcg"
)

// ItemTemplateRegistry manages available item templates
type ItemTemplateRegistry struct {
	templates       map[string]*pcg.ItemTemplate
	rarityModifiers map[pcg.RarityTier]RarityModifier
}

// RarityModifier defines how rarity affects item generation
type RarityModifier struct {
	StatMultiplier    float64  `yaml:"stat_multiplier"`
	EnchantmentChance float64  `yaml:"enchantment_chance"`
	MaxEnchantments   int      `yaml:"max_enchantments"`
	ValueMultiplier   float64  `yaml:"value_multiplier"`
	NamePrefixes      []string `yaml:"name_prefixes"`
	NameSuffixes      []string `yaml:"name_suffixes"`
}

// NewItemTemplateRegistry creates a new item template registry
func NewItemTemplateRegistry() *ItemTemplateRegistry {
	return &ItemTemplateRegistry{
		templates:       make(map[string]*pcg.ItemTemplate),
		rarityModifiers: make(map[pcg.RarityTier]RarityModifier),
	}
}

// LoadDefaultTemplates loads built-in item templates
func (itr *ItemTemplateRegistry) LoadDefaultTemplates() error {
	// Define default weapon templates
	swordTemplate := &pcg.ItemTemplate{
		BaseType:  "weapon",
		NameParts: []string{"Blade", "Sword", "Saber", "Falchion"},
		StatRanges: map[string]pcg.StatRange{
			"damage": {Min: 6, Max: 8, Scaling: 0.1},
			"value":  {Min: 10, Max: 50, Scaling: 0.5},
			"weight": {Min: 2, Max: 4, Scaling: 0.0},
		},
		Properties: []string{"slashing", "martial"},
		Materials:  []string{"iron", "steel", "mithril", "adamantine"},
		Rarities:   []pcg.RarityTier{pcg.RarityCommon, pcg.RarityUncommon, pcg.RarityRare, pcg.RarityEpic, pcg.RarityLegendary},
	}
	itr.templates["sword"] = swordTemplate

	bowTemplate := &pcg.ItemTemplate{
		BaseType:  "weapon",
		NameParts: []string{"Bow", "Longbow", "Shortbow", "Recurve"},
		StatRanges: map[string]pcg.StatRange{
			"damage": {Min: 6, Max: 6, Scaling: 0.1},
			"range":  {Min: 80, Max: 150, Scaling: 1.0},
			"value":  {Min: 15, Max: 75, Scaling: 0.5},
			"weight": {Min: 1, Max: 3, Scaling: 0.0},
		},
		Properties: []string{"ranged", "martial", "ammunition"},
		Materials:  []string{"wood", "yew", "ironwood", "dragonbone"},
		Rarities:   []pcg.RarityTier{pcg.RarityCommon, pcg.RarityUncommon, pcg.RarityRare, pcg.RarityEpic},
	}
	itr.templates["bow"] = bowTemplate

	// Define default armor templates
	leatherTemplate := &pcg.ItemTemplate{
		BaseType:  "armor",
		NameParts: []string{"Leather", "Hide", "Studded"},
		StatRanges: map[string]pcg.StatRange{
			"ac":     {Min: 11, Max: 12, Scaling: 0.05},
			"value":  {Min: 10, Max: 40, Scaling: 0.3},
			"weight": {Min: 8, Max: 12, Scaling: 0.0},
		},
		Properties: []string{"light"},
		Materials:  []string{"leather", "studded_leather", "dragonskin"},
		Rarities:   []pcg.RarityTier{pcg.RarityCommon, pcg.RarityUncommon, pcg.RarityRare},
	}
	itr.templates["armor"] = leatherTemplate

	// Define consumable templates
	potionTemplate := &pcg.ItemTemplate{
		BaseType:  "consumable",
		NameParts: []string{"Potion", "Elixir", "Draught"},
		StatRanges: map[string]pcg.StatRange{
			"healing": {Min: 8, Max: 16, Scaling: 0.5},
			"value":   {Min: 25, Max: 100, Scaling: 0.8},
			"weight":  {Min: 1, Max: 1, Scaling: 0.0},
		},
		Properties: []string{"consumable", "magical"},
		Materials:  []string{"glass", "crystal", "vial"},
		Rarities:   []pcg.RarityTier{pcg.RarityCommon, pcg.RarityUncommon, pcg.RarityRare},
	}
	itr.templates["potion"] = potionTemplate

	// Load default rarity modifiers
	itr.loadDefaultRarityModifiers()

	return nil
}

// LoadFromFile loads templates from YAML file
func (itr *ItemTemplateRegistry) LoadFromFile(configPath string) error {
	// TODO: Implement YAML file loading
	// For now, just ensure default templates are loaded
	return itr.LoadDefaultTemplates()
}

// GetTemplate retrieves template by base type and rarity
func (itr *ItemTemplateRegistry) GetTemplate(baseType string, rarity pcg.RarityTier) (*pcg.ItemTemplate, error) {
	template, exists := itr.templates[baseType]
	if !exists {
		return nil, fmt.Errorf("template not found for type: %s", baseType)
	}

	// Check if the template supports this rarity
	supported := false
	for _, supportedRarity := range template.Rarities {
		if supportedRarity == rarity {
			supported = true
			break
		}
	}

	if !supported {
		// Return template anyway but with common rarity as fallback
		templateCopy := *template
		return &templateCopy, nil
	}

	// Return a copy of the template
	templateCopy := *template
	return &templateCopy, nil
}

// GetRarityModifier returns rarity modifier for given tier
func (itr *ItemTemplateRegistry) GetRarityModifier(rarity pcg.RarityTier) RarityModifier {
	if modifier, exists := itr.rarityModifiers[rarity]; exists {
		return modifier
	}
	// Return common rarity as fallback
	return itr.rarityModifiers[pcg.RarityCommon]
}

// loadDefaultRarityModifiers loads built-in rarity modifiers
func (itr *ItemTemplateRegistry) loadDefaultRarityModifiers() {
	itr.rarityModifiers[pcg.RarityCommon] = RarityModifier{
		StatMultiplier:    1.0,
		EnchantmentChance: 0.0,
		MaxEnchantments:   0,
		ValueMultiplier:   1.0,
		NamePrefixes:      []string{},
		NameSuffixes:      []string{},
	}

	itr.rarityModifiers[pcg.RarityUncommon] = RarityModifier{
		StatMultiplier:    1.1,
		EnchantmentChance: 0.3,
		MaxEnchantments:   1,
		ValueMultiplier:   2.0,
		NamePrefixes:      []string{"Fine", "Quality"},
		NameSuffixes:      []string{},
	}

	itr.rarityModifiers[pcg.RarityRare] = RarityModifier{
		StatMultiplier:    1.25,
		EnchantmentChance: 0.6,
		MaxEnchantments:   2,
		ValueMultiplier:   5.0,
		NamePrefixes:      []string{"Superior", "Masterwork"},
		NameSuffixes:      []string{"of Power"},
	}

	itr.rarityModifiers[pcg.RarityEpic] = RarityModifier{
		StatMultiplier:    1.5,
		EnchantmentChance: 0.8,
		MaxEnchantments:   3,
		ValueMultiplier:   10.0,
		NamePrefixes:      []string{"Epic", "Heroic"},
		NameSuffixes:      []string{"of the Champion", "of Might"},
	}

	itr.rarityModifiers[pcg.RarityLegendary] = RarityModifier{
		StatMultiplier:    2.0,
		EnchantmentChance: 1.0,
		MaxEnchantments:   4,
		ValueMultiplier:   25.0,
		NamePrefixes:      []string{"Legendary", "Mythic"},
		NameSuffixes:      []string{"of Legend", "of the Gods"},
	}

	itr.rarityModifiers[pcg.RarityArtifact] = RarityModifier{
		StatMultiplier:    3.0,
		EnchantmentChance: 1.0,
		MaxEnchantments:   5,
		ValueMultiplier:   100.0,
		NamePrefixes:      []string{"Artifact", "Primordial"},
		NameSuffixes:      []string{"of Creation", "of the Ancients"},
	}
}

// GenerateItemName creates procedural item names
func GenerateItemName(template *pcg.ItemTemplate, rarity pcg.RarityTier, rng *rand.Rand) string {
	registry := NewItemTemplateRegistry()
	registry.LoadDefaultTemplates()

	modifier := registry.GetRarityModifier(rarity)

	// Start with base name
	baseName := template.NameParts[rng.Intn(len(template.NameParts))]

	// Add material qualifier
	if len(template.Materials) > 0 {
		material := template.Materials[rng.Intn(len(template.Materials))]
		baseName = material + " " + baseName
	}

	// Add rarity prefix
	if len(modifier.NamePrefixes) > 0 && rng.Float64() < 0.7 {
		prefix := modifier.NamePrefixes[rng.Intn(len(modifier.NamePrefixes))]
		baseName = prefix + " " + baseName
	}

	// Add rarity suffix
	if len(modifier.NameSuffixes) > 0 && rng.Float64() < 0.5 {
		suffix := modifier.NameSuffixes[rng.Intn(len(modifier.NameSuffixes))]
		baseName = baseName + " " + suffix
	}

	return baseName
}
