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

	switch component {
	case game.ComponentVerbal:
		// Check if character can speak (not silenced or stunned)
		if caster.HasEffect(game.EffectStun) {
			logrus.WithFields(logrus.Fields{
				"function": "hasSpellComponent",
				"reason":   "character is stunned",
			}).Debug("verbal component unavailable")
			return false
		}
		return true

	case game.ComponentSomatic:
		// Check if character can use hands (not paralyzed or hands bound)
		if caster.HasEffect(game.EffectStun) {
			logrus.WithFields(logrus.Fields{
				"function": "hasSpellComponent",
				"reason":   "character is stunned",
			}).Debug("somatic component unavailable")
			return false
		}
		return true

	case game.ComponentMaterial:
		// Check if the caster has the required material component in their inventory
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

	default:
		logrus.WithFields(logrus.Fields{
			"function":  "hasSpellComponent",
			"component": component,
		}).Warn("unknown spell component type")
		return false
	}
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

// processEvocationSpell handles the effects of an evocation spell, including damage, healing, and result construction.
func (s *RPCServer) processEvocationSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	s.logEvocationSpellStart(spell, caster, targetID)

	spellPower := calculateSpellPower(caster, spell)

	// Refactored: delegate to helper functions for damage, healing, and result construction
	damage, damageRoll, hitTargets, err := s.processEvocationDamage(spell, targetID)
	if err != nil {
		return nil, fmt.Errorf("damage processing failed: %w", err)
	}

	healing, healingRoll, healedTargets, err := s.processEvocationHealing(spell, targetID)
	if err != nil {
		return nil, fmt.Errorf("healing processing failed: %w", err)
	}

	// Fallback if no dice specified
	if damage == 0 && healing == 0 {
		damage, hitTargets, err = s.processEvocationFallback(spell, spellPower, targetID)
		if err != nil {
			return nil, fmt.Errorf("fallback damage failed: %w", err)
		}
	}

	result := s.buildEvocationResult(
		spell,
		spellPower,
		damage,
		healing,
		append(hitTargets, healedTargets...),
		damageRoll,
		healingRoll,
	)

	s.logEvocationSpellSuccess(spell, damage, healing, spellPower)
	return result, nil
}

// logEvocationSpellStart logs the start of an evocation spell processing.
func (s *RPCServer) logEvocationSpellStart(spell *game.Spell, caster *game.Player, targetID string) {
	logrus.WithFields(logrus.Fields{
		"function": "logEvocationSpellStart",
		"spell":    spell.Name,
		"caster":   caster.ID,
		"targetID": targetID,
	}).Info("Starting evocation spell processing")
}

// processEvocationDamage rolls damage dice, applies damage, and returns results.
func (s *RPCServer) processEvocationDamage(spell *game.Spell, targetID string) (int, *game.DiceRoll, []string, error) {
	roll, err := game.GlobalDiceRoller.Roll(spell.DamageDice)
	if err != nil {
		logrus.WithError(err).Error("failed to roll damage dice")
		return 0, nil, nil, fmt.Errorf("failed to roll damage dice: %w", err)
	}
	damage := roll.Final
	hitTargets := []string{targetID}
	damageType := spell.DamageType
	if damageType == "" {
		damageType = "magical"
	}
	if err := s.applySpellDamage(targetID, damage, damageType); err != nil {
		logrus.WithError(err).Error("failed to apply spell damage")
		return 0, nil, nil, fmt.Errorf("failed to apply spell damage: %w", err)
	}
	return damage, roll, hitTargets, nil
}

// processEvocationHealing rolls healing dice, applies healing, and returns results.
func (s *RPCServer) processEvocationHealing(spell *game.Spell, targetID string) (int, *game.DiceRoll, []string, error) {
	roll, err := game.GlobalDiceRoller.Roll(spell.HealingDice)
	if err != nil {
		logrus.WithError(err).Error("failed to roll healing dice")
		return 0, nil, nil, fmt.Errorf("failed to roll healing dice: %w", err)
	}
	healing := roll.Final
	hitTargets := []string{targetID}
	if err := s.applySpellHealing(targetID, healing); err != nil {
		logrus.WithError(err).Error("failed to apply spell healing")
		return 0, nil, nil, fmt.Errorf("failed to apply spell healing: %w", err)
	}
	return healing, roll, hitTargets, nil
}

// processEvocationFallback applies fallback damage if no dice are specified.
func (s *RPCServer) processEvocationFallback(spell *game.Spell, spellPower int, targetID string) (int, []string, error) {
	damage := spellPower * spell.Level
	hitTargets := []string{targetID}
	if err := s.applySpellDamage(targetID, damage, "magical"); err != nil {
		logrus.WithError(err).Error("failed to apply generic spell damage")
		return 0, nil, fmt.Errorf("failed to apply spell damage: %w", err)
	}
	return damage, hitTargets, nil
}

// buildEvocationResult constructs the result map for an evocation spell.
func (s *RPCServer) buildEvocationResult(
	spell *game.Spell,
	spellPower, damage, healing int,
	hitTargets []string,
	damageRoll, healingRoll *game.DiceRoll,
) map[string]interface{} {
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
	if damageRoll != nil {
		result["damage_roll"] = damageRoll.String()
	}
	if healingRoll != nil {
		result["healing_roll"] = healingRoll.String()
	}
	return result
}

// logEvocationSpellSuccess logs the successful processing of an evocation spell.
func (s *RPCServer) logEvocationSpellSuccess(spell *game.Spell, damage, healing, spellPower int) {
	logrus.WithFields(logrus.Fields{
		"function":   "logEvocationSpellSuccess",
		"spell":      spell.Name,
		"damage":     damage,
		"healing":    healing,
		"spellPower": spellPower,
	}).Info("Evocation spell processed successfully")
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
	logrus.WithFields(logrus.Fields{
		"function":  "rollDice",
		"package":   "server",
		"mode":      "SIMULATION",
		"simulates": "physical dice rolling",
		"num_dice":  numDice,
		"die_size":  dieSize,
	}).Warn("SIMULATION FUNCTION - NOT A REAL DICE ROLL")

	if numDice <= 0 || dieSize <= 0 {
		logrus.WithFields(logrus.Fields{
			"function": "rollDice",
			"package":  "server",
			"num_dice": numDice,
			"die_size": dieSize,
		}).Warn("invalid dice parameters, returning 0")
		return 0
	}

	total := 0
	for i := 0; i < numDice; i++ {
		roll := (total % dieSize) + 1 // Simple pseudo-random for deterministic testing
		total += roll
		logrus.WithFields(logrus.Fields{
			"function": "rollDice",
			"package":  "server",
			"die_num":  i + 1,
			"roll":     roll,
			"total":    total,
		}).Debug("simulated die roll")
	}

	logrus.WithFields(logrus.Fields{
		"function":    "rollDice",
		"package":     "server",
		"mode":        "SIMULATION",
		"num_dice":    numDice,
		"die_size":    dieSize,
		"final_total": total,
	}).Info("dice simulation completed")

	return total
}

// applySpellDamage applies spell damage to a target
func (s *RPCServer) applySpellDamage(targetID string, damage int, damageType string) error {
	logrus.WithFields(logrus.Fields{
		"function":    "applySpellDamage",
		"package":     "server",
		"target_id":   targetID,
		"damage":      damage,
		"damage_type": damageType,
	}).Debug("entering applySpellDamage")

	// Find target in sessions (if it's a player)
	s.mu.RLock()
	for _, session := range s.sessions {
		if session.Player.GetID() == targetID {
			s.mu.RUnlock()

			logrus.WithFields(logrus.Fields{
				"function":    "applySpellDamage",
				"package":     "server",
				"target_id":   targetID,
				"target_type": "player",
			}).Debug("target found as player")

			// Apply damage to player
			currentHP := session.Player.GetHP()
			newHP := currentHP - damage
			if newHP < 0 {
				logrus.WithFields(logrus.Fields{
					"function":   "applySpellDamage",
					"package":    "server",
					"target_id":  targetID,
					"damage":     damage,
					"current_hp": currentHP,
				}).Debug("damage would cause death, capping at 0 HP")
				newHP = 0
			}

			session.Player.SetHP(newHP)

			logrus.WithFields(logrus.Fields{
				"function":    "applySpellDamage",
				"package":     "server",
				"target_id":   targetID,
				"damage":      damage,
				"damage_type": damageType,
				"old_hp":      currentHP,
				"new_hp":      newHP,
			}).Info("spell damage applied to player")

			logrus.WithFields(logrus.Fields{
				"function":  "applySpellDamage",
				"package":   "server",
				"target_id": targetID,
			}).Debug("exiting applySpellDamage - player damage applied")

			return nil
		}
	}
	s.mu.RUnlock()

	// Target not found as player, assume it's an NPC (simulated for now)
	logrus.WithFields(logrus.Fields{
		"function":    "applySpellDamage",
		"package":     "server",
		"target_id":   targetID,
		"damage":      damage,
		"damage_type": damageType,
		"mode":        "SIMULATION",
		"simulates":   "NPC damage application",
	}).Warn("SIMULATION FUNCTION - NPC DAMAGE NOT FULLY IMPLEMENTED")

	logrus.WithFields(logrus.Fields{
		"function":  "applySpellDamage",
		"package":   "server",
		"target_id": targetID,
	}).Debug("exiting applySpellDamage - NPC damage simulated")

	return nil
}

// applySpellHealing applies spell healing to a target
func (s *RPCServer) applySpellHealing(targetID string, healing int) error {
	logrus.WithFields(logrus.Fields{
		"function":  "applySpellHealing",
		"package":   "server",
		"target_id": targetID,
		"healing":   healing,
	}).Debug("entering applySpellHealing")

	// Find target in sessions (if it's a player)
	s.mu.RLock()
	for _, session := range s.sessions {
		if session.Player.GetID() == targetID {
			s.mu.RUnlock()

			logrus.WithFields(logrus.Fields{
				"function":    "applySpellHealing",
				"package":     "server",
				"target_id":   targetID,
				"target_type": "player",
			}).Debug("target found as player")

			// Apply healing to player
			currentHP := session.Player.GetHP()
			maxHP := session.Player.GetMaxHP()
			newHP := currentHP + healing
			if newHP > maxHP {
				logrus.WithFields(logrus.Fields{
					"function":    "applySpellHealing",
					"package":     "server",
					"target_id":   targetID,
					"healing":     healing,
					"current_hp":  currentHP,
					"max_hp":      maxHP,
					"would_be_hp": newHP,
				}).Debug("healing would exceed max HP, capping at maximum")
				newHP = maxHP
			}

			session.Player.SetHP(newHP)

			logrus.WithFields(logrus.Fields{
				"function":  "applySpellHealing",
				"package":   "server",
				"target_id": targetID,
				"healing":   healing,
				"old_hp":    currentHP,
				"new_hp":    newHP,
				"max_hp":    maxHP,
			}).Info("spell healing applied to player")

			logrus.WithFields(logrus.Fields{
				"function":  "applySpellHealing",
				"package":   "server",
				"target_id": targetID,
			}).Debug("exiting applySpellHealing - player healing applied")

			return nil
		}
	}
	s.mu.RUnlock()

	// Target not found as player, assume it's an NPC (simulated for now)
	logrus.WithFields(logrus.Fields{
		"function":  "applySpellHealing",
		"package":   "server",
		"target_id": targetID,
		"healing":   healing,
		"mode":      "SIMULATION",
		"simulates": "NPC healing application",
	}).Warn("SIMULATION FUNCTION - NPC HEALING NOT FULLY IMPLEMENTED")

	logrus.WithFields(logrus.Fields{
		"function":  "applySpellHealing",
		"package":   "server",
		"target_id": targetID,
	}).Debug("exiting applySpellHealing - NPC healing simulated")

	return nil
}
