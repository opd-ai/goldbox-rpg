package server

import (
	"strings"
	"testing"

	"goldbox-rpg/pkg/game"
)

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestRPCServer_hasSpellComponent tests the hasSpellComponent method
func TestRPCServer_hasSpellComponent(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name      string
		caster    *game.Player
		component game.SpellComponent
		expected  bool
	}{
		{
			name:      "MaterialComponent_ComponentFound",
			component: game.ComponentMaterial,
			caster: &game.Player{
				Character: game.Character{
					ID: "test-player",
					Inventory: []game.Item{
						{Type: "SpellComponent", Name: "Crystal"},
						{Type: "Weapon", Name: "Sword"},
					},
				},
			},
			expected: true,
		},
		{
			name:      "MaterialComponent_ComponentNotFound",
			component: game.ComponentMaterial,
			caster: &game.Player{
				Character: game.Character{
					ID: "test-player",
					Inventory: []game.Item{
						{Type: "Weapon", Name: "Sword"},
						{Type: "Armor", Name: "Shield"},
					},
				},
			},
			expected: false,
		},
		{
			name:      "MaterialComponent_EmptyInventory",
			component: game.ComponentMaterial,
			caster: &game.Player{
				Character: game.Character{
					ID:        "test-player",
					Inventory: []game.Item{},
				},
			},
			expected: false,
		},
		{
			name:      "VerbalComponent_AlwaysReturnsFalse",
			component: game.ComponentVerbal,
			caster: &game.Player{
				Character: game.Character{
					ID: "test-player",
				},
			},
			expected: false,
		},
		{
			name:      "SomaticComponent_AlwaysReturnsFalse",
			component: game.ComponentSomatic,
			caster: &game.Player{
				Character: game.Character{
					ID: "test-player",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.hasSpellComponent(tt.caster, tt.component)
			if result != tt.expected {
				t.Errorf("hasSpellComponent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestRPCServer_validateSpellCast tests the validateSpellCast method
func TestRPCServer_validateSpellCast(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name        string
		caster      *game.Player
		spell       *game.Spell
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidSpellCast_NoComponents",
			caster: &game.Player{
				Character: game.Character{ID: "test-player"},
				Level:     5,
			},
			spell: &game.Spell{
				ID:         "fireball",
				Level:      3,
				Components: []game.SpellComponent{},
			},
			expectError: false,
		},
		{
			name: "InvalidSpellCast_InsufficientLevel",
			caster: &game.Player{
				Character: game.Character{ID: "test-player"},
				Level:     2,
			},
			spell: &game.Spell{
				ID:         "fireball",
				Level:      5,
				Components: []game.SpellComponent{},
			},
			expectError: true,
			errorMsg:    "insufficient level to cast spell",
		},
		{
			name: "ValidSpellCast_WithMaterialComponent",
			caster: &game.Player{
				Character: game.Character{
					ID: "test-player",
					Inventory: []game.Item{
						{Type: "SpellComponent", Name: "Crystal"},
					},
				},
				Level: 5,
			},
			spell: &game.Spell{
				ID:         "magic_missile",
				Level:      1,
				Components: []game.SpellComponent{game.ComponentMaterial},
			},
			expectError: false,
		},
		{
			name: "InvalidSpellCast_MissingMaterialComponent",
			caster: &game.Player{
				Character: game.Character{
					ID: "test-player",
					Inventory: []game.Item{
						{Type: "Weapon", Name: "Sword"},
					},
				},
				Level: 5,
			},
			spell: &game.Spell{
				ID:         "magic_missile",
				Level:      1,
				Components: []game.SpellComponent{game.ComponentMaterial},
			},
			expectError: true,
			errorMsg:    "missing required spell component",
		},
		{
			name: "InvalidSpellCast_MissingVerbalComponent",
			caster: &game.Player{
				Character: game.Character{ID: "test-player"},
				Level:     5,
			},
			spell: &game.Spell{
				ID:         "healing_word",
				Level:      1,
				Components: []game.SpellComponent{game.ComponentVerbal},
			},
			expectError: true,
			errorMsg:    "missing required spell component",
		},
		{
			name: "ValidSpellCast_ExactLevel",
			caster: &game.Player{
				Character: game.Character{ID: "test-player"},
				Level:     3,
			},
			spell: &game.Spell{
				ID:         "fireball",
				Level:      3,
				Components: []game.SpellComponent{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateSpellCast(tt.caster, tt.spell)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateSpellCast() expected error but got nil")
					return
				}
				if tt.errorMsg != "" {
					errMsg := err.Error()
					if errMsg != tt.errorMsg && !containsString(errMsg, tt.errorMsg) {
						t.Errorf("validateSpellCast() error = %v, want error containing %v", errMsg, tt.errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("validateSpellCast() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestRPCServer_processEvocationSpell tests the processEvocationSpell method
func TestRPCServer_processEvocationSpell(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name     string
		spell    *game.Spell
		caster   *game.Player
		targetID string
	}{
		{
			name: "EvocationSpell_ValidInput",
			spell: &game.Spell{
				ID:   "fireball",
				Name: "Fireball",
			},
			caster: &game.Player{
				Character: game.Character{ID: "caster-1"},
			},
			targetID: "target-1",
		},
		{
			name: "EvocationSpell_EmptyTargetID",
			spell: &game.Spell{
				ID:   "magic_missile",
				Name: "Magic Missile",
			},
			caster: &game.Player{
				Character: game.Character{ID: "caster-2"},
			},
			targetID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.processEvocationSpell(tt.spell, tt.caster, tt.targetID)

			if err != nil {
				t.Errorf("processEvocationSpell() unexpected error = %v", err)
				return
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Errorf("processEvocationSpell() result is not a map[string]interface{}")
				return
			}

			// Check that success is true
			if success, exists := resultMap["success"]; !exists || success != true {
				t.Errorf("processEvocationSpell() success = %v, want true", success)
			}

			// Check that spell_id matches
			if spellID, exists := resultMap["spell_id"]; !exists || spellID != tt.spell.ID {
				t.Errorf("processEvocationSpell() spell_id = %v, want %v", spellID, tt.spell.ID)
			}
		})
	}
}

// TestRPCServer_processEnchantmentSpell tests the processEnchantmentSpell method
func TestRPCServer_processEnchantmentSpell(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name     string
		spell    *game.Spell
		caster   *game.Player
		targetID string
	}{
		{
			name: "EnchantmentSpell_ValidInput",
			spell: &game.Spell{
				ID:   "bless",
				Name: "Bless",
			},
			caster: &game.Player{
				Character: game.Character{ID: "caster-1"},
			},
			targetID: "target-1",
		},
		{
			name: "EnchantmentSpell_SelfTarget",
			spell: &game.Spell{
				ID:   "shield",
				Name: "Shield",
			},
			caster: &game.Player{
				Character: game.Character{ID: "caster-1"},
			},
			targetID: "caster-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.processEnchantmentSpell(tt.spell, tt.caster, tt.targetID)

			if err != nil {
				t.Errorf("processEnchantmentSpell() unexpected error = %v", err)
				return
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Errorf("processEnchantmentSpell() result is not a map[string]interface{}")
				return
			}

			// Check that success is true
			if success, exists := resultMap["success"]; !exists || success != true {
				t.Errorf("processEnchantmentSpell() success = %v, want true", success)
			}

			// Check that spell_id matches
			if spellID, exists := resultMap["spell_id"]; !exists || spellID != tt.spell.ID {
				t.Errorf("processEnchantmentSpell() spell_id = %v, want %v", spellID, tt.spell.ID)
			}
		})
	}
}

// TestRPCServer_processIllusionSpell tests the processIllusionSpell method
func TestRPCServer_processIllusionSpell(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name   string
		spell  *game.Spell
		caster *game.Player
		pos    game.Position
	}{
		{
			name: "IllusionSpell_ValidInput",
			spell: &game.Spell{
				ID:   "fog_cloud",
				Name: "Fog Cloud",
			},
			caster: &game.Player{
				Character: game.Character{ID: "caster-1"},
			},
			pos: game.Position{X: 10, Y: 15},
		},
		{
			name: "IllusionSpell_ZeroPosition",
			spell: &game.Spell{
				ID:   "mirror_image",
				Name: "Mirror Image",
			},
			caster: &game.Player{
				Character: game.Character{ID: "caster-2"},
			},
			pos: game.Position{X: 0, Y: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.processIllusionSpell(tt.spell, tt.caster, tt.pos)

			if err != nil {
				t.Errorf("processIllusionSpell() unexpected error = %v", err)
				return
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Errorf("processIllusionSpell() result is not a map[string]interface{}")
				return
			}

			// Check that success is true
			if success, exists := resultMap["success"]; !exists || success != true {
				t.Errorf("processIllusionSpell() success = %v, want true", success)
			}

			// Check that spell_id matches
			if spellID, exists := resultMap["spell_id"]; !exists || spellID != tt.spell.ID {
				t.Errorf("processIllusionSpell() spell_id = %v, want %v", spellID, tt.spell.ID)
			}
		})
	}
}

// TestRPCServer_processGenericSpell tests the processGenericSpell method
func TestRPCServer_processGenericSpell(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name     string
		spell    *game.Spell
		caster   *game.Player
		targetID string
	}{
		{
			name: "GenericSpell_ValidInput",
			spell: &game.Spell{
				ID:   "unknown_spell",
				Name: "Unknown Spell",
			},
			caster: &game.Player{
				Character: game.Character{ID: "caster-1"},
			},
			targetID: "target-1",
		},
		{
			name: "GenericSpell_NoTarget",
			spell: &game.Spell{
				ID:   "cantrip",
				Name: "Cantrip",
			},
			caster: &game.Player{
				Character: game.Character{ID: "caster-2"},
			},
			targetID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.processGenericSpell(tt.spell, tt.caster, tt.targetID)

			if err != nil {
				t.Errorf("processGenericSpell() unexpected error = %v", err)
				return
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Errorf("processGenericSpell() result is not a map[string]interface{}")
				return
			}

			// Check that success is true
			if success, exists := resultMap["success"]; !exists || success != true {
				t.Errorf("processGenericSpell() success = %v, want true", success)
			}

			// Check that spell_id matches
			if spellID, exists := resultMap["spell_id"]; !exists || spellID != tt.spell.ID {
				t.Errorf("processGenericSpell() spell_id = %v, want %v", spellID, tt.spell.ID)
			}
		})
	}
}

// TestRPCServer_SpellProcessing_TableDriven is a comprehensive table-driven test
// covering multiple spell processing scenarios
func TestRPCServer_SpellProcessing_TableDriven(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name        string
		setup       func() (*game.Spell, *game.Player, string, game.Position)
		testFunc    string // Which function to test: "evocation", "enchantment", "illusion", "generic"
		expectError bool
	}{
		{
			name: "AllSpellTypes_SuccessfulProcessing",
			setup: func() (*game.Spell, *game.Player, string, game.Position) {
				spell := &game.Spell{ID: "test_spell", Name: "Test Spell"}
				player := &game.Player{Character: game.Character{ID: "test_player"}}
				return spell, player, "target", game.Position{X: 5, Y: 5}
			},
			testFunc:    "evocation",
			expectError: false,
		},
		{
			name: "NilSpell_HandledGracefully",
			setup: func() (*game.Spell, *game.Player, string, game.Position) {
				player := &game.Player{Character: game.Character{ID: "test_player"}}
				return nil, player, "target", game.Position{X: 5, Y: 5}
			},
			testFunc:    "generic",
			expectError: true, // This should cause a panic/error due to nil spell
		},
		{
			name: "NilPlayer_HandledGracefully",
			setup: func() (*game.Spell, *game.Player, string, game.Position) {
				spell := &game.Spell{ID: "test_spell", Name: "Test Spell"}
				return spell, nil, "target", game.Position{X: 5, Y: 5}
			},
			testFunc:    "enchantment",
			expectError: true, // This should cause a panic/error due to nil player
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell, player, targetID, pos := tt.setup()

			var result interface{}
			var err error
			var panicked bool

			// Capture panics
			defer func() {
				if r := recover(); r != nil {
					panicked = true
					if !tt.expectError {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			switch tt.testFunc {
			case "evocation":
				result, err = server.processEvocationSpell(spell, player, targetID)
			case "enchantment":
				result, err = server.processEnchantmentSpell(spell, player, targetID)
			case "illusion":
				result, err = server.processIllusionSpell(spell, player, pos)
			case "generic":
				result, err = server.processGenericSpell(spell, player, targetID)
			default:
				t.Fatalf("Unknown test function: %s", tt.testFunc)
			}

			if tt.expectError {
				if err == nil && !panicked {
					t.Errorf("Expected error or panic but got neither")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if panicked {
					t.Errorf("Unexpected panic")
				}

				// For non-error cases, verify basic result structure
				if result != nil {
					if resultMap, ok := result.(map[string]interface{}); ok {
						if success, exists := resultMap["success"]; !exists || success != true {
							t.Errorf("Expected success=true in result")
						}
					}
				}
			}
		})
	}
}

// TestRPCServer_ComponentValidation_EdgeCases tests edge cases for component validation
func TestRPCServer_ComponentValidation_EdgeCases(t *testing.T) {
	server := &RPCServer{}

	t.Run("MultipleSpellComponents_InInventory", func(t *testing.T) {
		caster := &game.Player{
			Character: game.Character{
				ID: "test-player",
				Inventory: []game.Item{
					{Type: "SpellComponent", Name: "Crystal"},
					{Type: "SpellComponent", Name: "Herb"},
					{Type: "Weapon", Name: "Staff"},
				},
			},
		}

		result := server.hasSpellComponent(caster, game.ComponentMaterial)
		if !result {
			t.Errorf("hasSpellComponent() = false, want true when multiple spell components exist")
		}
	})

	t.Run("CaseInsensitive_ComponentType", func(t *testing.T) {
		caster := &game.Player{
			Character: game.Character{
				ID: "test-player",
				Inventory: []game.Item{
					{Type: "spellcomponent", Name: "Crystal"}, // lowercase
				},
			},
		}

		result := server.hasSpellComponent(caster, game.ComponentMaterial)
		if result {
			t.Errorf("hasSpellComponent() = true, want false for case-sensitive type matching")
		}
	})

	t.Run("ComplexSpell_MultipleComponents", func(t *testing.T) {
		caster := &game.Player{
			Character: game.Character{
				ID: "test-player",
				Inventory: []game.Item{
					{Type: "SpellComponent", Name: "Crystal"},
				},
			},
			Level: 10,
		}

		spell := &game.Spell{
			ID:    "complex_spell",
			Level: 5,
			Components: []game.SpellComponent{
				game.ComponentVerbal,
				game.ComponentSomatic,
				game.ComponentMaterial,
			},
		}

		err := server.validateSpellCast(caster, spell)
		if err == nil {
			t.Errorf("validateSpellCast() expected error for missing verbal/somatic components")
		}
	})
}
