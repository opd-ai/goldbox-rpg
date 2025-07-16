package items

import (
	"fmt"
	"math/rand"

	"goldbox-rpg/pkg/game"

	"goldbox-rpg/pkg/pcg"
)

// EnchantmentSystem manages procedural enchantments
type EnchantmentSystem struct {
	enchantments map[string]*pcg.EnchantmentTemplate
	schools      map[string][]string // Magic school -> enchantment list
	rng          *rand.Rand
}

// NewEnchantmentSystem creates a new enchantment system
func NewEnchantmentSystem() *EnchantmentSystem {
	es := &EnchantmentSystem{
		enchantments: make(map[string]*pcg.EnchantmentTemplate),
		schools:      make(map[string][]string),
	}

	// Load default enchantments
	es.loadDefaultEnchantments()

	return es
}

// SetSeed sets the random seed for enchantment generation
func (es *EnchantmentSystem) SetSeed(seed int64) {
	es.rng = rand.New(rand.NewSource(seed))
}

// ApplyEnchantments adds procedural enchantments to an item
func (es *EnchantmentSystem) ApplyEnchantments(item *game.Item, rarity pcg.RarityTier, playerLevel int, rng *rand.Rand) error {
	if rng == nil {
		return fmt.Errorf("random generator is nil")
	}

	if es.rng == nil {
		es.rng = rng
	}

	// Get rarity modifier to determine enchantment parameters
	registry := NewItemTemplateRegistry()
	registry.LoadDefaultTemplates()
	modifier := registry.GetRarityModifier(rarity)

	// Check if we should add enchantments
	if es.rng.Float64() > modifier.EnchantmentChance {
		return nil // No enchantments for this item
	}

	// Determine number of enchantments
	numEnchantments := 1
	if modifier.MaxEnchantments > 1 {
		numEnchantments = 1 + es.rng.Intn(modifier.MaxEnchantments)
	}

	// Get available enchantments for this item type
	availableEnchants := es.GetAvailableEnchantments(item.Type, 1, playerLevel)
	if len(availableEnchants) == 0 {
		return nil // No available enchantments
	}

	// Apply enchantments
	appliedEnchants := make(map[string]bool) // Track to avoid duplicates
	for i := 0; i < numEnchantments && len(appliedEnchants) < len(availableEnchants); i++ {
		// Select random enchantment
		enchant := availableEnchants[es.rng.Intn(len(availableEnchants))]

		// Skip if already applied
		if appliedEnchants[enchant.Name] {
			continue
		}

		// Apply enchantment effects
		if err := es.applyEnchantmentToItem(item, enchant, playerLevel); err != nil {
			continue // Skip failed enchantments
		}

		appliedEnchants[enchant.Name] = true

		// Update item name if this is the first enchantment
		if len(appliedEnchants) == 1 {
			item.Name = enchant.Name + " " + item.Name
		}
	}

	// Update item value based on enchantments
	enchantmentMultiplier := 1.0 + (0.5 * float64(len(appliedEnchants)))
	item.Value = int(float64(item.Value) * enchantmentMultiplier)

	return nil
}

// GetAvailableEnchantments returns enchantments valid for item type
func (es *EnchantmentSystem) GetAvailableEnchantments(itemType string, minLevel, maxLevel int) []*pcg.EnchantmentTemplate {
	var available []*pcg.EnchantmentTemplate

	for _, enchant := range es.enchantments {
		// Check level requirements
		if enchant.MinLevel > maxLevel || enchant.MaxLevel < minLevel {
			continue
		}

		// Check item type compatibility
		if es.isEnchantmentCompatible(enchant, itemType) {
			available = append(available, enchant)
		}
	}

	return available
}

// applyEnchantmentToItem applies a specific enchantment to an item
func (es *EnchantmentSystem) applyEnchantmentToItem(item *game.Item, enchant *pcg.EnchantmentTemplate, playerLevel int) error {
	// Apply effects based on enchantment type
	switch enchant.Type {
	case "weapon_bonus":
		return es.applyWeaponBonus(item, enchant, playerLevel)
	case "armor_bonus":
		return es.applyArmorBonus(item, enchant, playerLevel)
	case "damage_type":
		return es.applyDamageType(item, enchant, playerLevel)
	case "resistance":
		return es.applyResistance(item, enchant, playerLevel)
	default:
		// Generic property addition
		return es.applyGenericEnchantment(item, enchant, playerLevel)
	}
}

// applyWeaponBonus applies weapon enhancement enchantments
func (es *EnchantmentSystem) applyWeaponBonus(item *game.Item, enchant *pcg.EnchantmentTemplate, playerLevel int) error {
	if item.Type != "weapon" {
		return fmt.Errorf("weapon bonus applied to non-weapon")
	}

	// Add weapon enhancement property
	bonus := 1 + (playerLevel / 5) // +1 per 5 levels
	if bonus > 5 {
		bonus = 5 // Cap at +5
	}

	enchantProp := fmt.Sprintf("+%d enhancement", bonus)
	item.Properties = append(item.Properties, enchantProp)

	return nil
}

// applyArmorBonus applies armor enhancement enchantments
func (es *EnchantmentSystem) applyArmorBonus(item *game.Item, enchant *pcg.EnchantmentTemplate, playerLevel int) error {
	if item.Type != "armor" {
		return fmt.Errorf("armor bonus applied to non-armor")
	}

	// Increase AC
	bonus := 1 + (playerLevel / 4) // +1 per 4 levels
	if bonus > 3 {
		bonus = 3 // Cap at +3
	}

	item.AC += bonus
	enchantProp := fmt.Sprintf("+%d armor", bonus)
	item.Properties = append(item.Properties, enchantProp)

	return nil
}

// applyDamageType applies elemental damage enchantments
func (es *EnchantmentSystem) applyDamageType(item *game.Item, enchant *pcg.EnchantmentTemplate, playerLevel int) error {
	if item.Type != "weapon" {
		return fmt.Errorf("damage type applied to non-weapon")
	}

	elements := []string{"fire", "cold", "lightning", "acid"}
	element := elements[es.rng.Intn(len(elements))]

	damageDice := 1 + (playerLevel / 3) // Additional dice every 3 levels
	if damageDice > 6 {
		damageDice = 6
	}

	enchantProp := fmt.Sprintf("+%dd6 %s", damageDice, element)
	item.Properties = append(item.Properties, enchantProp)

	return nil
}

// applyResistance applies resistance enchantments
func (es *EnchantmentSystem) applyResistance(item *game.Item, enchant *pcg.EnchantmentTemplate, playerLevel int) error {
	if item.Type != "armor" {
		return fmt.Errorf("resistance applied to non-armor")
	}

	resistances := []string{"fire", "cold", "lightning", "acid", "necrotic"}
	resistance := resistances[es.rng.Intn(len(resistances))]

	enchantProp := fmt.Sprintf("resistance %s", resistance)
	item.Properties = append(item.Properties, enchantProp)

	return nil
}

// applyGenericEnchantment applies generic property enchantments
func (es *EnchantmentSystem) applyGenericEnchantment(item *game.Item, enchant *pcg.EnchantmentTemplate, playerLevel int) error {
	// Add enchantment name as property
	item.Properties = append(item.Properties, enchant.Name)
	return nil
}

// isEnchantmentCompatible checks if enchantment works with item type
func (es *EnchantmentSystem) isEnchantmentCompatible(enchant *pcg.EnchantmentTemplate, itemType string) bool {
	switch enchant.Type {
	case "weapon_bonus", "damage_type":
		return itemType == "weapon"
	case "armor_bonus", "resistance":
		return itemType == "armor"
	case "utility", "generic":
		return true // Compatible with all items
	default:
		return true // Default to compatible
	}
}

// loadDefaultEnchantments loads built-in enchantment templates
func (es *EnchantmentSystem) loadDefaultEnchantments() {
	// Weapon enchantments
	es.enchantments["enhancement"] = &pcg.EnchantmentTemplate{
		Name:     "Enhancement",
		Type:     "weapon_bonus",
		MinLevel: 1,
		MaxLevel: 20,
		Effects:  []game.Effect{}, // Effects would be defined if we had the Effect system
		Restrictions: map[string]interface{}{
			"item_types": []string{"weapon"},
		},
	}

	es.enchantments["flaming"] = &pcg.EnchantmentTemplate{
		Name:     "Flaming",
		Type:     "damage_type",
		MinLevel: 3,
		MaxLevel: 20,
		Effects:  []game.Effect{},
		Restrictions: map[string]interface{}{
			"item_types": []string{"weapon"},
			"elements":   []string{"fire"},
		},
	}

	es.enchantments["frost"] = &pcg.EnchantmentTemplate{
		Name:     "Frost",
		Type:     "damage_type",
		MinLevel: 3,
		MaxLevel: 20,
		Effects:  []game.Effect{},
		Restrictions: map[string]interface{}{
			"item_types": []string{"weapon"},
			"elements":   []string{"cold"},
		},
	}

	es.enchantments["shock"] = &pcg.EnchantmentTemplate{
		Name:     "Shock",
		Type:     "damage_type",
		MinLevel: 3,
		MaxLevel: 20,
		Effects:  []game.Effect{},
		Restrictions: map[string]interface{}{
			"item_types": []string{"weapon"},
			"elements":   []string{"lightning"},
		},
	}

	// Armor enchantments
	es.enchantments["protection"] = &pcg.EnchantmentTemplate{
		Name:     "Protection",
		Type:     "armor_bonus",
		MinLevel: 1,
		MaxLevel: 20,
		Effects:  []game.Effect{},
		Restrictions: map[string]interface{}{
			"item_types": []string{"armor"},
		},
	}

	es.enchantments["fire_resistance"] = &pcg.EnchantmentTemplate{
		Name:     "Fire Resistance",
		Type:     "resistance",
		MinLevel: 5,
		MaxLevel: 20,
		Effects:  []game.Effect{},
		Restrictions: map[string]interface{}{
			"item_types": []string{"armor"},
			"elements":   []string{"fire"},
		},
	}

	// Set up magic schools
	es.schools["evocation"] = []string{"flaming", "frost", "shock"}
	es.schools["abjuration"] = []string{"protection", "fire_resistance"}
	es.schools["transmutation"] = []string{"enhancement"}
}
