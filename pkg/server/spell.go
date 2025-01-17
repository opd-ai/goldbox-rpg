package server

import (
	"goldbox-rpg/pkg/game"
)

func (s *RPCServer) processSpellCast(caster *game.Player, spell *game.Spell, targetID string, pos game.Position) (interface{}, error) {
	// Validate spell requirements
	if err := s.validateSpellCast(caster, spell); err != nil {
		return nil, err
	}

	// Process spell effects based on type
	switch spell.School {
	case game.SchoolEvocation:
		return s.processEvocationSpell(spell, caster, targetID)
	case game.SchoolEnchantment:
		return s.processEnchantmentSpell(spell, caster, targetID)
	case game.SchoolIllusion:
		return s.processIllusionSpell(spell, caster, pos)
	default:
		return s.processGenericSpell(spell, caster, targetID)
	}
}
