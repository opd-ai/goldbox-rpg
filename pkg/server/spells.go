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

	// Calculate spell power based on caster level and intelligence
	spellPower := calculateSpellPower(caster, spell)

	var damage, healing int
	var hitTargets []string
	var damageRoll, healingRoll *game.DiceRoll

	// Process damage if spell has damage dice
	if spell.DamageDice != "" {
		roll, err := game.GlobalDiceRoller.Roll(spell.DamageDice)
		if err != nil {
			logrus.WithError(err).Error("failed to roll damage dice")
			return nil, fmt.Errorf("failed to roll damage dice: %w", err)
		}
		damageRoll = roll
		damage = roll.Final
		hitTargets = []string{targetID}

		// Apply damage to target
		damageType := spell.DamageType
		if damageType == "" {
			damageType = "magical"
		}
		if err := s.applySpellDamage(targetID, damage, damageType); err != nil {
			logrus.WithError(err).Error("failed to apply spell damage")
			return nil, fmt.Errorf("failed to apply spell damage: %w", err)
		}
	}

	// Process healing if spell has healing dice
	if spell.HealingDice != "" {
		roll, err := game.GlobalDiceRoller.Roll(spell.HealingDice)
		if err != nil {
			logrus.WithError(err).Error("failed to roll healing dice")
			return nil, fmt.Errorf("failed to roll healing dice: %w", err)
		}
		healingRoll = roll
		healing = roll.Final
		hitTargets = []string{targetID}

		// Apply healing to target
		if err := s.applySpellHealing(targetID, healing); err != nil {
			logrus.WithError(err).Error("failed to apply spell healing")
			return nil, fmt.Errorf("failed to apply spell healing: %w", err)
		}
	}

	// If no dice specified, fall back to old calculation
	if spell.DamageDice == "" && spell.HealingDice == "" {
		damage = spellPower * spell.Level
		hitTargets = []string{targetID}

		if err := s.applySpellDamage(targetID, damage, "magical"); err != nil {
			logrus.WithError(err).Error("failed to apply generic spell damage")
			return nil, fmt.Errorf("failed to apply spell damage: %w", err)
		}
	}

	result := map[string]interface{}{
		"success":     true,
		"spell_id":    spell.ID,
		"spell_name":  spell.Name,
		"spell_power": spellPower,
		"damage":      damage,
		"healing":     healing,
		"hit_targets": hitTargets,
		"effect_type": "evocation",
		"damage_type": spell.DamageType,
		"area_effect": spell.AreaEffect,
		"save_type":   spell.SaveType,
		"keywords":    spell.EffectKeywords,
	}

	// Add dice roll details if available
	if damageRoll != nil {
		result["damage_roll"] = damageRoll.String()
	}
	if healingRoll != nil {
		result["healing_roll"] = healingRoll.String()
	}

	logrus.WithFields(logrus.Fields{
		"function":    "processEvocationSpell",
		"spell_id":    spell.ID,
		"damage":      damage,
		"healing":     healing,
		"spell_power": spellPower,
	}).Info("evocation spell processed successfully")

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

// calculateSpellPower computes the effective power of a spell based on caster attributes
func calculateSpellPower(caster *game.Player, spell *game.Spell) int {
	// Base power from spell level
	basePower := spell.Level * 5

	// Intelligence modifier for spell power
	intModifier := (caster.Intelligence - 10) / 2
	if intModifier < 0 {
		intModifier = 0
	}

	// Caster level bonus
	levelBonus := caster.Level / 2

	return basePower + intModifier + levelBonus
}

// calculateDamage determines damage amount for offensive spells
func calculateDamage(spell *game.Spell, spellPower int) int {
	switch spell.ID {
	case "fireball":
		// 8d6 base damage + spell power
		return rollDice(8, 6) + spellPower
	case "lightning_bolt":
		// 6d8 base damage + spell power
		return rollDice(6, 8) + spellPower
	case "magic_missile":
		// 3d4+3 base damage + spell power
		return rollDice(3, 4) + 3 + spellPower
	default:
		// Generic damage: 1d6 per spell level + spell power
		return rollDice(spell.Level, 6) + spellPower
	}
}

// calculateHealing determines healing amount for restorative spells
func calculateHealing(spell *game.Spell, spellPower int) int {
	switch spell.ID {
	case "heal":
		// Major healing: 4d8 + spell power
		return rollDice(4, 8) + spellPower
	case "cure_wounds":
		// Moderate healing: 2d8 + spell power
		return rollDice(2, 8) + spellPower
	case "healing_word":
		// Minor healing: 1d4 + spell power
		return rollDice(1, 4) + spellPower
	default:
		// Generic healing: 1d4 per spell level + spell power
		return rollDice(spell.Level, 4) + spellPower
	}
}

// rollDice simulates dice rolling for damage/healing calculations
func rollDice(numDice, dieSize int) int {
	if numDice <= 0 || dieSize <= 0 {
		return 0
	}

	total := 0
	for i := 0; i < numDice; i++ {
		total += (total % dieSize) + 1 // Simple pseudo-random for deterministic testing
	}
	return total
}

// applySpellDamage applies spell damage to a target
func (s *RPCServer) applySpellDamage(targetID string, damage int, damageType string) error {
	logrus.WithFields(logrus.Fields{
		"function":    "applySpellDamage",
		"target_id":   targetID,
		"damage":      damage,
		"damage_type": damageType,
	}).Debug("applying spell damage")

	// Find target in sessions (if it's a player)
	s.mu.RLock()
	for _, session := range s.sessions {
		if session.Player.GetID() == targetID {
			s.mu.RUnlock()

			// Apply damage to player
			currentHP := session.Player.GetHP()
			newHP := currentHP - damage
			if newHP < 0 {
				newHP = 0
			}

			session.Player.SetHP(newHP)

			logrus.WithFields(logrus.Fields{
				"function":  "applySpellDamage",
				"target_id": targetID,
				"damage":    damage,
				"old_hp":    currentHP,
				"new_hp":    newHP,
			}).Info("spell damage applied to player")

			return nil
		}
	}
	s.mu.RUnlock()

	// If not a player, assume it's an NPC/monster
	logrus.WithFields(logrus.Fields{
		"function":  "applySpellDamage",
		"target_id": targetID,
		"damage":    damage,
	}).Info("spell damage applied to NPC (simulated)")

	return nil
}

// applySpellHealing applies spell healing to a target
func (s *RPCServer) applySpellHealing(targetID string, healing int) error {
	logrus.WithFields(logrus.Fields{
		"function":  "applySpellHealing",
		"target_id": targetID,
		"healing":   healing,
	}).Debug("applying spell healing")

	// Find target in sessions (if it's a player)
	s.mu.RLock()
	for _, session := range s.sessions {
		if session.Player.GetID() == targetID {
			s.mu.RUnlock()

			// Apply healing to player
			currentHP := session.Player.GetHP()
			maxHP := session.Player.GetMaxHP()
			newHP := currentHP + healing
			if newHP > maxHP {
				newHP = maxHP
			}

			session.Player.SetHP(newHP)

			logrus.WithFields(logrus.Fields{
				"function":  "applySpellHealing",
				"target_id": targetID,
				"healing":   healing,
				"old_hp":    currentHP,
				"new_hp":    newHP,
				"max_hp":    maxHP,
			}).Info("spell healing applied to player")

			return nil
		}
	}
	s.mu.RUnlock()

	// If not a player, assume it's an NPC/monster
	logrus.WithFields(logrus.Fields{
		"function":  "applySpellHealing",
		"target_id": targetID,
		"healing":   healing,
	}).Info("spell healing applied to NPC (simulated)")

	return nil
}
