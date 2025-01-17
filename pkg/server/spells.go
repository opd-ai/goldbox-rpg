package server

import (
	"fmt"
	"goldbox-rpg/pkg/game"
)

func (s *RPCServer) hasSpellComponent(caster *game.Player, component game.SpellComponent) bool {
	// Check if the caster has the required spell component in their inventory
	if component == game.ComponentMaterial {
		for _, item := range caster.Inventory {
			if item.Type == "SpellComponent" {
				return true
			}
		}
		return false
	}
	return false
}

func (s *RPCServer) validateSpellCast(caster *game.Player, spell *game.Spell) error {
	// Check level requirements
	if caster.Level < spell.Level {
		return fmt.Errorf("insufficient level to cast spell")
	}

	// Check components
	for _, component := range spell.Components {
		if !s.hasSpellComponent(caster, component) {
			return fmt.Errorf("missing required spell component: %v", component)
		}
	}

	return nil
}

func (s *RPCServer) processEvocationSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	// Implement damage/healing spells
	return map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}, nil
}

func (s *RPCServer) processEnchantmentSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	// Implement buff/debuff spells
	return map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}, nil
}

func (s *RPCServer) processIllusionSpell(spell *game.Spell, caster *game.Player, pos game.Position) (interface{}, error) {
	// Implement area effect spells
	return map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}, nil
}

func (s *RPCServer) processGenericSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	// Default spell processing
	return map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}, nil
}
