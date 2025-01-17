package game

// EquipmentSlot represents the different slots where equipment/items can be equipped on a character.
// This type is used as an enum to identify valid equipment positions (e.g. weapon slot, armor slot, etc).
type EquipmentSlot int

const (
	SlotHead EquipmentSlot = iota
	SlotNeck
	SlotChest
	SlotHands
	SlotRings
	SlotLegs
	SlotFeet
	SlotWeaponMain
	SlotWeaponOff
)

// String returns a human-readable string representation of an EquipmentSlot.
// This method maps the numeric equipment slot enum value to its corresponding
// string name from a fixed array of slot names.
//
// Returns:
//   - string: The name of the equipment slot (one of: Head, Neck, Chest, Hands,
//     Rings, Legs, Feet, MainHand, OffHand)
//
// Note: This method will panic if the EquipmentSlot value is outside the valid
// range (0-8) as it directly indexes into a fixed array.
func (es EquipmentSlot) String() string {
	return [...]string{
		"Head",
		"Neck",
		"Chest",
		"Hands",
		"Rings",
		"Legs",
		"Feet",
		"MainHand",
		"OffHand",
	}[es]
}

// EquipmentSlotConfig defines the configuration for an equipment slot in the game.
// It specifies what types of items can be equipped and any special requirements.
//
// Fields:
//   - Slot: The type of equipment slot (e.g. weapon, armor, etc)
//   - Name: Human readable display name for the equipment slot
//   - Description: Detailed description of what items can be equipped in this slot
//   - AllowedTypes: List of item type IDs that can be equipped in this slot
//   - Restricted: If true, additional requirements must be met to use this slot
//
// Related types:
//   - EquipmentSlot (enum type for slot categories)
//   - Item (for equippable items)
type EquipmentSlotConfig struct {
	Slot         EquipmentSlot `yaml:"slot_type"`        // Type of equipment slot
	Name         string        `yaml:"slot_name"`        // Display name for the slot
	Description  string        `yaml:"slot_description"` // Description of what can be equipped
	AllowedTypes []string      `yaml:"allowed_types"`    // Types of items that can be equipped
	Restricted   bool          `yaml:"slot_restricted"`  // Whether slot has special requirements
}

// EquipmentSet represents a character's complete set of equipped items across different slots.
// This struct maintains the relationship between a character and their equipped items.
//
// Fields:
//   - CharacterID: Unique identifier string for the character who owns this equipment set
//   - Slots: Map containing the configuration for each equipment slot, keyed by EquipmentSlot type
//
// The Slots map allows for flexible equipment configurations while enforcing slot-specific
// validation rules defined in EquipmentSlotConfig.
//
// Related types:
//   - EquipmentSlot: Enum defining valid equipment slot types
//   - EquipmentSlotConfig: Configuration for individual equipment slots
type EquipmentSet struct {
	CharacterID string                                `yaml:"character_id"`    // ID of character owning the equipment
	Slots       map[EquipmentSlot]EquipmentSlotConfig `yaml:"equipment_slots"` // Map of all equipment slots
}
