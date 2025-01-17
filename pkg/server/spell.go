package server

import (
	"goldbox-rpg/pkg/game"
)

// processSpellCast handles the execution of a spell cast by a player.
// It validates the spell requirements and processes the effects based on the spell school.
//
// Parameters:
//   - caster: *game.Player - The player casting the spell
//   - spell: *game.Spell - The spell being cast
//   - targetID: string - ID of the target (player/monster/object)
//   - pos: game.Position - Position for location-based spells
//
// Returns:
//   - interface{} - The result of the spell cast, specific to each spell type
//   - error - Any validation or processing errors that occurred
//
// Errors:
//   - Returns validation errors from validateSpellCast
//   - May return errors from individual spell processing functions
//
// Related:
//   - validateSpellCast
//   - processEvocationSpell
//   - processEnchantmentSpell
//   - processIllusionSpell
//   - processGenericSpell
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
