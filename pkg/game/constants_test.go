package game

import (
	"testing"
)

// TestDirectionConstants_Values tests that direction constants have expected values
func TestDirectionConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		expected  Direction
	}{
		{"North", DirectionNorth, 0},
		{"East", DirectionEast, 1},
		{"South", DirectionSouth, 2},
		{"West", DirectionWest, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.direction != tt.expected {
				t.Errorf("Direction %s = %d, want %d", tt.name, tt.direction, tt.expected)
			}
		})
	}
}

// TestDirectionConstants_LegacyCompatibility tests backward compatibility constants
func TestDirectionConstants_LegacyCompatibility_MatchNewConstants(t *testing.T) {
	tests := []struct {
		name   string
		legacy Direction
		modern Direction
	}{
		{"North compatibility", North, DirectionNorth},
		{"East compatibility", East, DirectionEast},
		{"South compatibility", South, DirectionSouth},
		{"West compatibility", West, DirectionWest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.legacy != tt.modern {
				t.Errorf("%s: legacy constant %d != modern constant %d", tt.name, tt.legacy, tt.modern)
			}
		})
	}
}

// TestTileTypeConstants_Values tests tile type constant values
func TestTileTypeConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name     string
		tileType TileType
		expected TileType
	}{
		{"Floor", TileFloor, 0},
		{"Wall", TileWall, 1},
		{"Door", TileDoor, 2},
		{"Water", TileWater, 3},
		{"Lava", TileLava, 4},
		{"Pit", TilePit, 5},
		{"Stairs", TileStairs, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tileType != tt.expected {
				t.Errorf("TileType %s = %d, want %d", tt.name, tt.tileType, tt.expected)
			}
		})
	}
}

// TestEffectTypeConstants_StringValues tests effect type string constants
func TestEffectTypeConstants_StringValues_HaveExpectedValues(t *testing.T) {
	tests := []struct {
		name       string
		effectType EffectType
		expected   string
	}{
		{"DamageOverTime", EffectDamageOverTime, "damage_over_time"},
		{"HealOverTime", EffectHealOverTime, "heal_over_time"},
		{"Poison", EffectPoison, "poison"},
		{"Burning", EffectBurning, "burning"},
		{"Bleeding", EffectBleeding, "bleeding"},
		{"Stun", EffectStun, "stun"},
		{"Root", EffectRoot, "root"},
		{"StatBoost", EffectStatBoost, "stat_boost"},
		{"StatPenalty", EffectStatPenalty, "stat_penalty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.effectType) != tt.expected {
				t.Errorf("EffectType %s = %q, want %q", tt.name, string(tt.effectType), tt.expected)
			}
		})
	}
}

// TestDamageTypeConstants_StringValues tests damage type string constants
func TestDamageTypeConstants_StringValues_HaveExpectedValues(t *testing.T) {
	tests := []struct {
		name       string
		damageType DamageType
		expected   string
	}{
		{"Physical", DamagePhysical, "physical"},
		{"Fire", DamageFire, "fire"},
		{"Poison", DamagePoison, "poison"},
		{"Frost", DamageFrost, "frost"},
		{"Lightning", DamageLightning, "lightning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.damageType) != tt.expected {
				t.Errorf("DamageType %s = %q, want %q", tt.name, string(tt.damageType), tt.expected)
			}
		})
	}
}

// TestDispelTypeConstants_StringValues tests dispel type string constants
func TestDispelTypeConstants_StringValues_HaveExpectedValues(t *testing.T) {
	tests := []struct {
		name       string
		dispelType DispelType
		expected   string
	}{
		{"Magic", DispelMagic, "magic"},
		{"Curse", DispelCurse, "curse"},
		{"Poison", DispelPoison, "poison"},
		{"Disease", DispelDisease, "disease"},
		{"All", DispelAll, "all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.dispelType) != tt.expected {
				t.Errorf("DispelType %s = %q, want %q", tt.name, string(tt.dispelType), tt.expected)
			}
		})
	}
}

// TestImmunityTypeConstants_Values tests immunity type constant values
func TestImmunityTypeConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name         string
		immunityType ImmunityType
		expected     ImmunityType
	}{
		{"None", ImmunityNone, 0},
		{"Partial", ImmunityPartial, 1},
		{"Complete", ImmunityComplete, 2},
		{"Reflect", ImmunityReflect, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.immunityType != tt.expected {
				t.Errorf("ImmunityType %s = %d, want %d", tt.name, tt.immunityType, tt.expected)
			}
		})
	}
}

// TestDispelPriorityConstants_Values tests dispel priority constant values
func TestDispelPriorityConstants_Values_AreInAscendingOrder(t *testing.T) {
	tests := []struct {
		name           string
		dispelPriority DispelPriority
		expected       DispelPriority
	}{
		{"Lowest", DispelPriorityLowest, 0},
		{"Low", DispelPriorityLow, 25},
		{"Normal", DispelPriorityNormal, 50},
		{"High", DispelPriorityHigh, 75},
		{"Highest", DispelPriorityHighest, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dispelPriority != tt.expected {
				t.Errorf("DispelPriority %s = %d, want %d", tt.name, tt.dispelPriority, tt.expected)
			}
		})
	}
}

// TestDispelPriorityConstants_Ordering tests that priorities are properly ordered
func TestDispelPriorityConstants_Ordering_IsAscending(t *testing.T) {
	priorities := []DispelPriority{
		DispelPriorityLowest,
		DispelPriorityLow,
		DispelPriorityNormal,
		DispelPriorityHigh,
		DispelPriorityHighest,
	}

	for i := 1; i < len(priorities); i++ {
		if priorities[i-1] >= priorities[i] {
			t.Errorf("Priority order broken: %d >= %d", priorities[i-1], priorities[i])
		}
	}
}

// TestModOpTypeConstants_StringValues tests modifier operation type string constants
func TestModOpTypeConstants_StringValues_HaveExpectedValues(t *testing.T) {
	tests := []struct {
		name      string
		modOpType ModOpType
		expected  string
	}{
		{"Add", ModAdd, "add"},
		{"Multiply", ModMultiply, "multiply"},
		{"Set", ModSet, "set"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.modOpType) != tt.expected {
				t.Errorf("ModOpType %s = %q, want %q", tt.name, string(tt.modOpType), tt.expected)
			}
		})
	}
}

// TestEquipmentSlotConstants_Values tests equipment slot constant values
func TestEquipmentSlotConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name     string
		slot     EquipmentSlot
		expected EquipmentSlot
	}{
		{"Head", SlotHead, 0},
		{"Neck", SlotNeck, 1},
		{"Chest", SlotChest, 2},
		{"Hands", SlotHands, 3},
		{"Rings", SlotRings, 4},
		{"Legs", SlotLegs, 5},
		{"Feet", SlotFeet, 6},
		{"WeaponMain", SlotWeaponMain, 7},
		{"WeaponOff", SlotWeaponOff, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.slot != tt.expected {
				t.Errorf("EquipmentSlot %s = %d, want %d", tt.name, tt.slot, tt.expected)
			}
		})
	}
}

// TestCharacterClassConstants_Values tests character class constant values
func TestCharacterClassConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name     string
		class    CharacterClass
		expected CharacterClass
	}{
		{"Fighter", ClassFighter, 0},
		{"Mage", ClassMage, 1},
		{"Cleric", ClassCleric, 2},
		{"Thief", ClassThief, 3},
		{"Ranger", ClassRanger, 4},
		{"Paladin", ClassPaladin, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.class != tt.expected {
				t.Errorf("CharacterClass %s = %d, want %d", tt.name, tt.class, tt.expected)
			}
		})
	}
}

// TestSpellSchoolConstants_Values tests spell school constant values
func TestSpellSchoolConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name     string
		school   SpellSchool
		expected SpellSchool
	}{
		{"Abjuration", SchoolAbjuration, 0},
		{"Conjuration", SchoolConjuration, 1},
		{"Divination", SchoolDivination, 2},
		{"Enchantment", SchoolEnchantment, 3},
		{"Evocation", SchoolEvocation, 4},
		{"Illusion", SchoolIllusion, 5},
		{"Necromancy", SchoolNecromancy, 6},
		{"Transmutation", SchoolTransmutation, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.school != tt.expected {
				t.Errorf("SpellSchool %s = %d, want %d", tt.name, tt.school, tt.expected)
			}
		})
	}
}

// TestSpellComponentConstants_Values tests spell component constant values
func TestSpellComponentConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name      string
		component SpellComponent
		expected  SpellComponent
	}{
		{"Verbal", ComponentVerbal, 0},
		{"Somatic", ComponentSomatic, 1},
		{"Material", ComponentMaterial, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.component != tt.expected {
				t.Errorf("SpellComponent %s = %d, want %d", tt.name, tt.component, tt.expected)
			}
		})
	}
}

// TestQuestStatusConstants_Values tests quest status constant values
func TestQuestStatusConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name     string
		status   QuestStatus
		expected QuestStatus
	}{
		{"NotStarted", QuestNotStarted, 0},
		{"Active", QuestActive, 1},
		{"Completed", QuestCompleted, 2},
		{"Failed", QuestFailed, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != tt.expected {
				t.Errorf("QuestStatus %s = %d, want %d", tt.name, tt.status, tt.expected)
			}
		})
	}
}

// TestEventTypeConstants_Values tests event type constant values
func TestEventTypeConstants_Values_AreSequentialFromZero(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		expected  EventType
	}{
		{"LevelUp", EventLevelUp, 0},
		{"Damage", EventDamage, 1},
		{"Death", EventDeath, 2},
		{"ItemPickup", EventItemPickup, 3},
		{"ItemDrop", EventItemDrop, 4},
		{"Movement", EventMovement, 5},
		{"SpellCast", EventSpellCast, 6},
		{"QuestUpdate", EventQuestUpdate, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.eventType != tt.expected {
				t.Errorf("EventType %s = %d, want %d", tt.name, tt.eventType, tt.expected)
			}
		})
	}
}

// TestItemTypeConstants_StringValues tests item type string constants
func TestItemTypeConstants_StringValues_HaveExpectedValues(t *testing.T) {
	tests := []struct {
		name     string
		itemType string
		expected string
	}{
		{"Weapon", ItemTypeWeapon, "weapon"},
		{"Armor", ItemTypeArmor, "armor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.itemType != tt.expected {
				t.Errorf("ItemType %s = %q, want %q", tt.name, tt.itemType, tt.expected)
			}
		})
	}
}

// TestDefaultWorldConstants_Values tests default world dimension constants
func TestDefaultWorldConstants_Values_HaveExpectedValues(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"DefaultWorldWidth", DefaultWorldWidth, 10},
		{"DefaultWorldHeight", DefaultWorldHeight, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.expected)
			}
		})
	}
}

// TestActionPointConstants_Values tests action point constant values
func TestActionPointConstants_Values_HaveExpectedValues(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"ActionPointsPerTurn", ActionPointsPerTurn, 2},
		{"ActionCostMove", ActionCostMove, 1},
		{"ActionCostAttack", ActionCostAttack, 1},
		{"ActionCostSpell", ActionCostSpell, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.expected)
			}
		})
	}
}

// TestActionPointConstants_LogicalRelationships tests logical relationships between action point constants
func TestActionPointConstants_LogicalRelationships_AreValid(t *testing.T) {
	// Test that individual action costs don't exceed total points per turn
	actionCosts := []int{ActionCostMove, ActionCostAttack, ActionCostSpell}

	for _, cost := range actionCosts {
		if cost > ActionPointsPerTurn {
			t.Errorf("Action cost %d exceeds ActionPointsPerTurn %d", cost, ActionPointsPerTurn)
		}
		if cost <= 0 {
			t.Errorf("Action cost %d should be positive", cost)
		}
	}

	// Test that you can perform basic actions within turn limit
	moveAndAttack := ActionCostMove + ActionCostAttack
	if moveAndAttack > ActionPointsPerTurn {
		t.Errorf("Move + Attack cost (%d) exceeds ActionPointsPerTurn (%d)", moveAndAttack, ActionPointsPerTurn)
	}
}

// TestConstants_TypeConsistency tests that related constants use consistent types
func TestConstants_TypeConsistency_IsCorrect(t *testing.T) {
	t.Run("Direction constants have same underlying type", func(t *testing.T) {
		// Test that all direction constants can be compared
		directions := []Direction{DirectionNorth, DirectionEast, DirectionSouth, DirectionWest}
		for i := 0; i < len(directions)-1; i++ {
			if directions[i] >= directions[i+1] {
				t.Errorf("Direction constants not in expected order: %d >= %d", directions[i], directions[i+1])
			}
		}
	})

	t.Run("String-based constants are non-empty", func(t *testing.T) {
		stringConstants := []string{
			string(EffectDamageOverTime),
			string(DamagePhysical),
			string(DispelMagic),
			string(ModAdd),
			ItemTypeWeapon,
			ItemTypeArmor,
		}

		for _, constant := range stringConstants {
			if constant == "" {
				t.Error("String constant should not be empty")
			}
		}
	})
}
