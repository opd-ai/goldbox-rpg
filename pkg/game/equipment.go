package game

// EquipmentSlot represents character equipment locations
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

// EquipmentSlotConfig represents serializable configuration for equipment slots
type EquipmentSlotConfig struct {
	Slot         EquipmentSlot `yaml:"slot_type"`        // Type of equipment slot
	Name         string        `yaml:"slot_name"`        // Display name for the slot
	Description  string        `yaml:"slot_description"` // Description of what can be equipped
	AllowedTypes []string      `yaml:"allowed_types"`    // Types of items that can be equipped
	Restricted   bool          `yaml:"slot_restricted"`  // Whether slot has special requirements
}

// EquipmentSet represents a complete set of equipment slots for serialization
type EquipmentSet struct {
	CharacterID string                                `yaml:"character_id"`    // ID of character owning the equipment
	Slots       map[EquipmentSlot]EquipmentSlotConfig `yaml:"equipment_slots"` // Map of all equipment slots
}
