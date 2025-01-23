package server

import (
	"fmt"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

func (s *RPCServer) hasSpellComponent(caster *game.Player, component game.SpellComponent) bool {
	logrus.WithFields(logrus.Fields{
		"function":  "hasSpellComponent",
		"caster_id": caster.ID,
		"component": component,
	}).Debug("checking spell component")

	// Check if the caster has the required spell component in their inventory
	if component == game.ComponentMaterial {
		for _, item := range caster.Inventory {
			if item.Type == "SpellComponent" {
				logrus.WithFields(logrus.Fields{
					"function": "hasSpellComponent",
				}).Debug("found required spell component")
				return true
			}
		}
		logrus.WithFields(logrus.Fields{
			"function":  "hasSpellComponent",
			"component": component,
		}).Warn("spell component not found in inventory")
		return false
	}

	logrus.WithFields(logrus.Fields{
		"function":  "hasSpellComponent",
		"component": component,
	}).Debug("non-material component check completed")
	return false
}

func (s *RPCServer) validateSpellCast(caster *game.Player, spell *game.Spell) error {
	logrus.WithFields(logrus.Fields{
		"function":  "validateSpellCast",
		"caster_id": caster.ID,
		"spell_id":  spell.ID,
	}).Debug("validating spell cast")

	// Check level requirements
	if caster.Level < spell.Level {
		logrus.WithFields(logrus.Fields{
			"function":       "validateSpellCast",
			"caster_level":   caster.Level,
			"required_level": spell.Level,
		}).Warn("insufficient level to cast spell")
		return fmt.Errorf("insufficient level to cast spell")
	}

	// Check components
	for _, component := range spell.Components {
		if !s.hasSpellComponent(caster, component) {
			logrus.WithFields(logrus.Fields{
				"function":  "validateSpellCast",
				"component": component,
			}).Warn("missing required spell component")
			return fmt.Errorf("missing required spell component: %v", component)
		}
	}

	logrus.WithFields(logrus.Fields{
		"function": "validateSpellCast",
	}).Debug("spell cast validation successful")
	return nil
}

func (s *RPCServer) processEvocationSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "processEvocationSpell",
		"spell_id": spell.ID,
		"caster":   caster.ID,
		"target":   targetID,
	}).Debug("processing evocation spell")

	// Implement damage/healing spells
	result := map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}

	logrus.WithFields(logrus.Fields{
		"function": "processEvocationSpell",
		"spell_id": spell.ID,
	}).Debug("evocation spell processed")

	return result, nil
}

func (s *RPCServer) processEnchantmentSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "processEnchantmentSpell",
		"spell_id": spell.ID,
		"caster":   caster.ID,
		"target":   targetID,
	}).Debug("processing enchantment spell")

	// Implement buff/debuff spells
	result := map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}

	logrus.WithFields(logrus.Fields{
		"function": "processEnchantmentSpell",
		"spell_id": spell.ID,
	}).Debug("enchantment spell processed")

	return result, nil
}

func (s *RPCServer) processIllusionSpell(spell *game.Spell, caster *game.Player, pos game.Position) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "processIllusionSpell",
		"spell_id": spell.ID,
		"caster":   caster.ID,
		"position": pos,
	}).Debug("processing illusion spell")

	// Implement area effect spells
	result := map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}

	logrus.WithFields(logrus.Fields{
		"function": "processIllusionSpell",
		"spell_id": spell.ID,
	}).Debug("illusion spell processed")

	return result, nil
}

func (s *RPCServer) processGenericSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "processGenericSpell",
		"spell_id": spell.ID,
		"caster":   caster.ID,
		"target":   targetID,
	}).Debug("processing generic spell")

	// Default spell processing
	result := map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}

	logrus.WithFields(logrus.Fields{
		"function": "processGenericSpell",
		"spell_id": spell.ID,
	}).Debug("generic spell processed")

	return result, nil
}
