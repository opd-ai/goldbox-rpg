package server

import (
	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
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
	logrus.WithFields(logrus.Fields{
		"function": "processSpellCast",
		"caster":   caster.ID,
		"spell":    spell.Name,
		"targetID": targetID,
	}).Debug("processing spell cast")

	// Validate spell requirements
	if err := s.validateSpellCast(caster, spell); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "processSpellCast",
			"error":    err.Error(),
		}).Error("spell validation failed")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function":    "processSpellCast",
		"spellSchool": spell.School,
	}).Info("processing spell by school")

	// Process spell effects based on type
	var result interface{}
	var err error
	switch spell.School {
	case game.SchoolEvocation:
		result, err = s.processEvocationSpell(spell, caster, targetID)
	case game.SchoolEnchantment:
		result, err = s.processEnchantmentSpell(spell, caster, targetID)
	case game.SchoolIllusion:
		result, err = s.processIllusionSpell(spell, caster, pos)
	default:
		logrus.WithFields(logrus.Fields{
			"function":    "processSpellCast",
			"spellSchool": spell.School,
		}).Warn("unknown spell school, using generic processing")
		result, err = s.processGenericSpell(spell, caster, targetID)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "processSpellCast",
			"error":    err.Error(),
		}).Error("spell processing failed")
	}

	return result, err
}
