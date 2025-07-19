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
	s.logSpellCastStart(caster, spell, targetID)

	if err := s.validateSpellCastForCast(caster, spell); err != nil {
		s.logSpellValidationError(err)
		return nil, err
	}

	s.logSpellSchoolProcessing(spell)

	result, err := s.dispatchSpellBySchool(spell, caster, targetID, pos)
	if err != nil {
		s.logSpellProcessingError(err)
	}

	return result, err
}

// logSpellCastStart logs the start of a spell cast attempt.
func (s *RPCServer) logSpellCastStart(caster *game.Player, spell *game.Spell, targetID string) {
	logrus.WithFields(logrus.Fields{
		"function": "processSpellCast",
		"caster":   caster.ID,
		"spell":    spell.Name,
		"targetID": targetID,
	}).Debug("processing spell cast")
}

// validateSpellCastForCast validates the spell requirements for casting.
func (s *RPCServer) validateSpellCastForCast(caster *game.Player, spell *game.Spell) error {
	return s.validateSpellCast(caster, spell)
}

// logSpellValidationError logs a spell validation error.
func (s *RPCServer) logSpellValidationError(err error) {
	logrus.WithFields(logrus.Fields{
		"function": "processSpellCast",
		"error":    err.Error(),
	}).Error("spell validation failed")
}

// logSpellSchoolProcessing logs the spell school being processed.
func (s *RPCServer) logSpellSchoolProcessing(spell *game.Spell) {
	logrus.WithFields(logrus.Fields{
		"function":    "processSpellCast",
		"spellSchool": spell.School,
	}).Info("processing spell by school")
}

// dispatchSpellBySchool routes spell processing to the correct handler based on school.
func (s *RPCServer) dispatchSpellBySchool(spell *game.Spell, caster *game.Player, targetID string, pos game.Position) (interface{}, error) {
	switch spell.School {
	case game.SchoolEvocation:
		return s.processEvocationSpell(spell, caster, targetID)
	case game.SchoolEnchantment:
		return s.processEnchantmentSpell(spell, caster, targetID)
	case game.SchoolIllusion:
		return s.processIllusionSpell(spell, caster, pos)
	default:
		logrus.WithFields(logrus.Fields{
			"function":    "processSpellCast",
			"spellSchool": spell.School,
		}).Warn("unknown spell school, using generic processing")
		return s.processGenericSpell(spell, caster, targetID)
	}
}

// logSpellProcessingError logs an error that occurred during spell processing.
func (s *RPCServer) logSpellProcessingError(err error) {
	logrus.WithFields(logrus.Fields{
		"function": "processSpellCast",
		"error":    err.Error(),
	}).Error("spell processing failed")
}
