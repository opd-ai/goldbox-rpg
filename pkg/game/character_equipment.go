package game

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Equipment Management Methods - extracted from character.go for maintainability

// EquipItem equips an item from the character's inventory to the specified slot.
//
// Parameters:
//   - itemID: The unique identifier of the item to equip
//   - slot: The equipment slot to equip the item to
//
// Returns:
//   - error: Returns nil on success, or an error if the item cannot be equipped
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) EquipItem(itemID string, slot EquipmentSlot) error {
	c.mu.Lock()

	itemIndex, itemToEquip, err := c.findItemInInventory(itemID)
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("character %s failed to find item %s for equipping: %w", c.ID, itemID, err)
	}

	if err := c.validateItemCanBeEquipped(itemToEquip, slot); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("character %s cannot equip item %s to slot %d: %w", c.ID, itemID, slot, err)
	}

	if err := c.handleSlotConflict(slot); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("character %s failed to handle slot %d conflict: %w", c.ID, slot, err)
	}

	c.equipItemToSlot(itemToEquip, slot)
	c.removeItemFromInventoryByIndex(itemIndex)

	c.mu.Unlock()

	// Apply resistance bonuses from equipment (after releasing lock)
	c.ApplyEquipmentResistances()

	return nil
}

// findItemInInventory searches for the item by ID and returns its index and value.
// Returns an error if not found.
func (c *Character) findItemInInventory(itemID string) (int, Item, error) {
	for i, item := range c.Inventory {
		if item.ID == itemID {
			return i, item, nil
		}
	}
	return -1, Item{}, NewInventoryError(itemID, "find", ErrItemNotFound)
}

// validateItemCanBeEquipped checks if the item can be equipped in the specified slot.
// Returns an error if validation fails.
func (c *Character) validateItemCanBeEquipped(item Item, slot EquipmentSlot) error {
	if !c.canEquipItemInSlot(item, slot) {
		return NewInventoryError(item.ID, "equip", ErrCannotEquipItem)
	}
	return nil
}

// handleSlotConflict unequips any existing item in the slot, if present.
// Returns an error if unequipping fails.
func (c *Character) handleSlotConflict(slot EquipmentSlot) error {
	if existingItem, exists := c.Equipment[slot]; exists {
		if _, err := c.unequipItemFromSlot(slot); err != nil {
			return fmt.Errorf("failed to unequip existing item %s: %w", existingItem.Name, err)
		}
	}
	return nil
}

// equipItemToSlot assigns the item to the specified equipment slot.
func (c *Character) equipItemToSlot(item Item, slot EquipmentSlot) {
	c.Equipment[slot] = item
}

// removeItemFromInventoryByIndex removes the item at the specified index from inventory.
func (c *Character) removeItemFromInventoryByIndex(index int) {
	c.Inventory = append(c.Inventory[:index], c.Inventory[index+1:]...)
}

// UnequipItem removes an item from the specified equipment slot and adds it to the character's inventory.
//
// Parameters:
//   - slot: The equipment slot to unequip
//
// Returns:
//   - *Item: Pointer to the unequipped item, or nil if slot was empty
//   - error: Returns nil on success, or an error if the operation fails
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) UnequipItem(slot EquipmentSlot) (*Item, error) {
	c.mu.Lock()
	item, err := c.unequipItemFromSlot(slot)
	c.mu.Unlock()

	if err == nil {
		// Apply resistance bonuses from equipment (after releasing lock)
		c.ApplyEquipmentResistances()
	}

	return item, err
}

// unequipItemFromSlot is the internal implementation of unequipping an item (requires lock to be held)
func (c *Character) unequipItemFromSlot(slot EquipmentSlot) (*Item, error) {
	// Check if there's an item equipped in this slot
	equippedItem, exists := c.Equipment[slot]
	if !exists {
		return nil, fmt.Errorf("no item equipped in slot %s", slot.String())
	}

	// Add the item back to inventory
	c.Inventory = append(c.Inventory, equippedItem)

	// Remove from equipment slot
	delete(c.Equipment, slot)

	return &equippedItem, nil
}

// CanEquipItem checks if the character can equip the specified item in the given slot.
// This performs all validation checks without actually equipping the item.
//
// Parameters:
//   - itemID: The unique identifier of the item to check
//   - slot: The equipment slot to check compatibility with
//
// Returns:
//   - bool: true if the item can be equipped, false otherwise
//   - error: Returns nil if check was successful, or an error if validation fails
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) CanEquipItem(itemID string, slot EquipmentSlot) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Find the item in inventory
	var itemToCheck Item
	found := false
	for _, item := range c.Inventory {
		if item.ID == itemID {
			itemToCheck = item
			found = true
			break
		}
	}

	if !found {
		return false, NewInventoryError(itemID, "check", ErrItemNotFound)
	}

	return c.canEquipItemInSlot(itemToCheck, slot), nil
}

// canEquipItemInSlot is the internal validation logic for equipment compatibility
func (c *Character) canEquipItemInSlot(item Item, slot EquipmentSlot) bool {
	if !c.isItemTypeValidForSlot(item, slot) {
		return false
	}

	proficiencies := GetClassProficiencies(c.Class)

	if c.isWeaponSlot(slot) {
		return c.canEquipWeaponInSlot(item, slot, proficiencies)
	}

	if c.isArmorSlot(slot) {
		return c.canEquipArmorInSlot(item, proficiencies)
	}

	return true
}

// isItemTypeValidForSlot checks if the item type is valid for the specified equipment slot.
// It returns true if the item can be placed in the slot based on type compatibility.
func (c *Character) isItemTypeValidForSlot(item Item, slot EquipmentSlot) bool {
	slotValidTypes := c.getSlotValidTypes()

	validTypes, exists := slotValidTypes[slot]
	if !exists {
		return false
	}

	for _, validType := range validTypes {
		if item.Type == validType {
			return true
		}
	}
	return false
}

// getSlotValidTypes returns the mapping of equipment slots to their valid item types.
// This defines which item types can be equipped in each slot.
func (c *Character) getSlotValidTypes() map[EquipmentSlot][]string {
	return map[EquipmentSlot][]string{
		SlotHead:       {"helmet", "hat", "crown", "circlet"},
		SlotNeck:       {"amulet", "necklace", "pendant"},
		SlotChest:      {"armor", "robe", "shirt", "vest"},
		SlotHands:      {"gloves", "gauntlets", "bracers"},
		SlotRings:      {"ring"},
		SlotLegs:       {"pants", "leggings", "greaves"},
		SlotFeet:       {"boots", "shoes", "sandals"},
		SlotWeaponMain: {"weapon", "sword", "axe", "staff", "bow", "dagger", "mace", "spear", "hammer", "wand"},
		SlotWeaponOff:  {"shield", "weapon", "dagger", "orb"},
	}
}

// isWeaponSlot checks if the given slot is a weapon slot.
// It returns true for main hand and off-hand weapon slots.
func (c *Character) isWeaponSlot(slot EquipmentSlot) bool {
	return slot == SlotWeaponMain || slot == SlotWeaponOff
}

// isArmorSlot checks if the given slot is an armor slot.
// It returns true for head, chest, hands, legs, and feet slots.
func (c *Character) isArmorSlot(slot EquipmentSlot) bool {
	return slot == SlotHead || slot == SlotChest || slot == SlotHands || slot == SlotLegs || slot == SlotFeet
}

// canEquipWeaponInSlot validates if a character can equip a weapon in the specified slot.
// It checks weapon proficiencies and special shield handling for off-hand slots.
func (c *Character) canEquipWeaponInSlot(item Item, slot EquipmentSlot, proficiencies ClassProficiencies) bool {
	// Special handling for shields in off-hand slot
	if slot == SlotWeaponOff && item.Type == "shield" {
		return proficiencies.ShieldProficient
	}

	// Allow generic "weapon" type if character has any weapon proficiencies
	if item.Type == "weapon" && len(proficiencies.WeaponTypes) > 0 {
		return true
	}

	// Check for specific weapon type match
	for _, allowedWeapon := range proficiencies.WeaponTypes {
		if item.Type == allowedWeapon {
			return true
		}
	}

	return false
}

// canEquipArmorInSlot validates if a character can equip armor in the specified slot.
// It checks armor proficiencies and determines armor type based on item properties.
func (c *Character) canEquipArmorInSlot(item Item, proficiencies ClassProficiencies) bool {
	if !c.isArmorItem(item) {
		return true // Non-armor items don't require armor proficiency
	}

	armorType := determineArmorType(item)

	for _, allowedArmor := range proficiencies.ArmorTypes {
		if armorType == allowedArmor {
			return true
		}
	}
	return false
}

// isArmorItem checks if the item is considered armor that requires proficiency.
// It returns true for items that are classified as armor types.
func (c *Character) isArmorItem(item Item) bool {
	return item.Type == "armor" || item.Type == "helmet" || item.Type == "gauntlets" || item.Type == "greaves"
}

// GetEquippedItem returns the item equipped in the specified slot.
//
// Parameters:
//   - slot: The equipment slot to check
//
// Returns:
//   - *Item: Pointer to the equipped item, or nil if slot is empty
//   - bool: true if an item is equipped in the slot, false otherwise
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetEquippedItem(slot EquipmentSlot) (*Item, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, exists := c.Equipment[slot]; exists {
		return &item, true
	}
	return nil, false
}

// GetAllEquippedItems returns a copy of all currently equipped items.
//
// Returns:
//   - map[EquipmentSlot]Item: A map containing all equipped items by slot
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetAllEquippedItems() map[EquipmentSlot]Item {
	c.mu.RLock()
	defer c.mu.RUnlock()

	equippedItems := make(map[EquipmentSlot]Item)
	for slot, item := range c.Equipment {
		equippedItems[slot] = item
	}
	return equippedItems
}

// GetEquipmentSlots returns all available equipment slots for this character.
//
// Returns:
//   - []EquipmentSlot: Slice containing all valid equipment slot types
func (c *Character) GetEquipmentSlots() []EquipmentSlot {
	return []EquipmentSlot{
		SlotHead, SlotNeck, SlotChest, SlotHands, SlotRings,
		SlotLegs, SlotFeet, SlotWeaponMain, SlotWeaponOff,
	}
}

// CalculateEquipmentBonuses calculates the total stat bonuses from all equipped items.
// This examines item properties for stat modifiers and returns the cumulative effect.
//
// Returns:
//   - map[string]int: Map of stat names to their total bonus values
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) CalculateEquipmentBonuses() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bonuses := make(map[string]int)

	for _, item := range c.Equipment {
		c.applyPropertyBonuses(item, bonuses)
		c.applyArmorClassBonus(item, bonuses)
	}

	return bonuses
}

// applyPropertyBonuses parses item properties for stat bonuses and updates the bonuses map.
func (c *Character) applyPropertyBonuses(item Item, bonuses map[string]int) {
	for _, property := range item.Properties {
		if len(property) > 2 {
			stat, value, ok := parseStatProperty(property)
			if ok && stat != "" {
				bonuses[stat] += value
			}
		}
	}
}

// applyArmorClassBonus adds AC bonus from armor items to the bonuses map.
func (c *Character) applyArmorClassBonus(item Item, bonuses map[string]int) {
	if item.Type == "armor" && item.AC > 0 {
		bonuses["armor_class"] += item.AC - 10 // Base AC is 10
	}
}

// parseStatProperty parses a property string like "strength+2" or "dexterity-10".
// Returns the stat name, the signed value, and true if parsing succeeded.
func parseStatProperty(property string) (string, int, bool) {
	signPos := -1
	for i := len(property) - 1; i >= 0; i-- {
		if property[i] == '+' || property[i] == '-' {
			signPos = i
			break
		}
	}
	if signPos > 0 && signPos < len(property)-1 {
		stat := property[:signPos]
		sign := 1
		if property[signPos] == '-' {
			sign = -1
		}
		var modifier int
		_, err := fmt.Sscanf(property[signPos+1:], "%d", &modifier)
		if err == nil {
			return stat, sign * modifier, true
		}
		logrus.WithField("property", property).Debug("failed to parse stat modifier")
	}
	return "", 0, false
}

// resistancePropertyMap maps resistance property names to their corresponding EffectType.
var resistancePropertyMap = map[string]EffectType{
	"fire_resistance":      EffectBurning,
	"poison_resistance":    EffectPoison,
	"frost_resistance":     EffectFrozen,
	"lightning_resistance": EffectShocked,
	"burning_resistance":   EffectBurning,
	"cold_resistance":      EffectFrozen,
	"electric_resistance":  EffectShocked,
}

// parseResistanceProperty parses a resistance property string like "fire_resistance+0.3".
// Returns the EffectType, the signed float value, and true if parsing succeeded.
// Valid values are in the range [-1.0, 1.0] where positive values grant resistance.
func parseResistanceProperty(property string) (EffectType, float64, bool) {
	signPos := -1
	for i := len(property) - 1; i >= 0; i-- {
		if property[i] == '+' || property[i] == '-' {
			signPos = i
			break
		}
	}
	if signPos <= 0 || signPos >= len(property)-1 {
		return "", 0, false
	}

	resistanceName := property[:signPos]
	effectType, exists := resistancePropertyMap[resistanceName]
	if !exists {
		return "", 0, false
	}

	sign := 1.0
	if property[signPos] == '-' {
		sign = -1.0
	}

	var modifier float64
	_, err := fmt.Sscanf(property[signPos+1:], "%f", &modifier)
	if err != nil {
		logrus.WithField("property", property).Debug("failed to parse resistance modifier")
		return "", 0, false
	}

	return effectType, sign * modifier, true
}

// CalculateEquipmentResistances calculates the total resistance bonuses from all equipped items.
// Returns a map of EffectType to the cumulative resistance value.
//
// Thread safety: This method is thread-safe using read mutex locking.
func (c *Character) CalculateEquipmentResistances() map[EffectType]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resistances := make(map[EffectType]float64)

	for _, item := range c.Equipment {
		for _, property := range item.Properties {
			effectType, value, ok := parseResistanceProperty(property)
			if ok {
				resistances[effectType] += value
			}
		}
	}

	// Clamp resistances to valid range [0.0, 1.0]
	for effectType, value := range resistances {
		if value < 0 {
			resistances[effectType] = 0
		} else if value > 1.0 {
			resistances[effectType] = 1.0
		}
	}

	return resistances
}

// ApplyEquipmentResistances recalculates and applies all equipment resistance bonuses
// to the character's EffectManager. This should be called after equipping or unequipping items.
//
// Thread safety: Acquires read lock for equipment iteration, then releases before
// calling EffectManager methods to avoid lock ordering issues.
func (c *Character) ApplyEquipmentResistances() {
	resistances := c.CalculateEquipmentResistances()

	if c.EffectManager == nil {
		return
	}

	// Clear existing equipment-based resistances by resetting to 0
	for _, effectType := range resistancePropertyMap {
		c.EffectManager.SetResistance(effectType, 0)
	}

	// Apply new resistances from equipment
	for effectType, value := range resistances {
		c.EffectManager.SetResistance(effectType, value)
	}
}

// determineArmorType determines the armor type (light, medium, heavy) based on item properties
func determineArmorType(item Item) string {
	// Check item properties for armor type indicators
	for _, property := range item.Properties {
		switch property {
		case "light", "light_armor":
			return "light"
		case "medium", "medium_armor":
			return "medium"
		case "heavy", "heavy_armor":
			return "heavy"
		}
	}

	// Default classification based on item type and name
	itemName := strings.ToLower(item.Name)
	switch {
	case strings.Contains(itemName, "leather") || strings.Contains(itemName, "cloth") || strings.Contains(itemName, "robe"):
		return "light"
	case strings.Contains(itemName, "chain") || strings.Contains(itemName, "scale") || strings.Contains(itemName, "studded"):
		return "medium"
	case strings.Contains(itemName, "plate") || strings.Contains(itemName, "full") || strings.Contains(itemName, "heavy"):
		return "heavy"
	default:
		// Default to light for unspecified armor
		return "light"
	}
}
