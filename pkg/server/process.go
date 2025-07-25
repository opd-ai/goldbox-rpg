package server

import (
	"fmt"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// processEffectTick handles the execution of a single effect tick in the game state.
// It routes effect processing to appropriate handlers based on effect type.
//
// This function manages effect processing for:
// - Damage over time effects (poison, burning, bleeding)
// - Healing over time effects (regeneration)
// - Stat modification effects (buffs, debuffs)
//
// Parameters:
//   - effect: The effect to process (must not be nil)
//
// Returns:
//   - error: nil on success, error if effect is nil or type unsupported
//
// Processing flow:
// 1. Validates effect is not nil
// 2. Determines effect type
// 3. Routes to specific effect handler
// 4. Logs processing results
func (gs *GameState) processEffectTick(effect *game.Effect) error {
	if err := gs.validateEffectNotNil(effect); err != nil {
		return err
	}

	switch effect.Type {
	case game.EffectDamageOverTime:
		return gs.handleDamageOverTimeEffect(effect)
	case game.EffectHealOverTime:
		return gs.handleHealingOverTimeEffect(effect)
	case game.EffectStatBoost, game.EffectStatPenalty:
		return gs.handleStatModificationEffect(effect)
	default:
		logrus.WithFields(logrus.Fields{
			"function": "processEffectTick",
			"effectID": effect.ID,
			"type":     effect.Type,
		}).Warn("unsupported effect type")
		return fmt.Errorf("unsupported effect type: %v", effect.Type)
	}
}

// validateEffectNotNil checks that the provided effect is not nil.
func (gs *GameState) validateEffectNotNil(effect *game.Effect) error {
	if effect == nil {
		logrus.WithFields(logrus.Fields{
			"function": "processEffectTick",
		}).Error("nil effect provided")
		return fmt.Errorf("effect is nil")
	}
	return nil
}

// handleDamageOverTimeEffect processes a damage over time effect.
func (gs *GameState) handleDamageOverTimeEffect(effect *game.Effect) error {
	err := gs.processDamageEffect(effect)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "processEffectTick",
			"effectID": effect.ID,
			"type":     effect.Type,
			"error":    err,
		}).Error("failed to process damage effect")
	}
	return err
}

// handleHealingOverTimeEffect processes a healing over time effect.
func (gs *GameState) handleHealingOverTimeEffect(effect *game.Effect) error {
	err := gs.processHealEffect(effect)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "processEffectTick",
			"effectID": effect.ID,
			"type":     effect.Type,
			"error":    err,
		}).Error("failed to process healing effect")
	}
	return err
}

// handleStatModificationEffect processes a stat boost or penalty effect.
func (gs *GameState) handleStatModificationEffect(effect *game.Effect) error {
	err := gs.processStatEffect(effect)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "processEffectTick",
			"effectID": effect.ID,
			"type":     effect.Type,
			"error":    err,
		}).Error("failed to process stat modification effect")
	}
	return err
}

// ADDED: processDamageEffect applies damage over time effects to target characters.
// It handles HP reduction and ensures characters don't go below 0 HP.
//
// Processing steps:
// 1. Validates target exists in world state
// 2. Ensures target is a Character type
// 3. Applies damage based on effect magnitude
// 4. Clamps HP to minimum value of 0
// 5. Logs damage application results
//
// Parameters:
//   - effect: Damage effect containing target ID and magnitude
//
// Returns:
//   - error: nil on success, error if target invalid or not found
//
// Effect handling: Only processes Character objects, ignores other entity types
func (gs *GameState) processDamageEffect(effect *game.Effect) error {
	logger := logrus.WithFields(logrus.Fields{
		"function": "processDamageEffect",
	})
	logger.Debug("processing damage effect")

	target, exists := gs.WorldState.Objects[effect.TargetID]
	if !exists {
		logger.WithField("targetID", effect.TargetID).Error("invalid effect target")
		return fmt.Errorf("invalid effect target")
	}

	if char, ok := target.(*game.Character); ok {
		damage := int(effect.Magnitude)
		char.HP -= damage
		if char.HP < 0 {
			char.HP = 0
			logger.WithFields(logrus.Fields{
				"targetID": effect.TargetID,
				"damage":   damage,
			}).Warn("character HP reduced to 0")
		} else {
			logger.WithFields(logrus.Fields{
				"targetID":    effect.TargetID,
				"damage":      damage,
				"remainingHP": char.HP,
			}).Info("applied damage to character")
		}
		return nil
	}

	logger.WithField("targetID", effect.TargetID).Error("target cannot receive damage")
	return fmt.Errorf("target cannot receive damage")
}

// processHealEffect applies a healing effect to a target character in the game world.
// It increases the target's HP by the effect magnitude, up to their max HP.
//
// Parameters:
//   - effect: *game.Effect - The healing effect to process, must contain:
//   - TargetID: ID of the character to heal
//   - Magnitude: Amount of HP to heal
//
// Returns:
//   - error: Returns nil on success, or an error if:
//   - Target does not exist in world state
//   - Target is not a Character type
//
// Related:
//   - game.Character
//   - game.Effect
//   - GameState.WorldState
//
// ADDED: processHealEffect applies healing over time effects to target characters.
// It restores HP while respecting maximum HP limits and logs healing results.
//
// Processing steps:
// 1. Validates target exists in world state
// 2. Ensures target is a Character type
// 3. Applies healing based on effect magnitude
// 4. Clamps HP to character's maximum HP value
// 5. Logs healing amount and HP changes
//
// Parameters:
//   - effect: Healing effect containing target ID and magnitude
//
// Returns:
//   - error: nil on success, error if target invalid or not found
//
// Healing mechanics: Only affects Character objects, respects MaxHP boundaries
func (gs *GameState) processHealEffect(effect *game.Effect) error {
	logger := logrus.WithFields(logrus.Fields{
		"function": "processHealEffect",
	})
	logger.Debug("processing heal effect")

	target, exists := gs.WorldState.Objects[effect.TargetID]
	if !exists {
		logger.WithField("targetID", effect.TargetID).Error("invalid effect target")
		return fmt.Errorf("invalid effect target")
	}

	if char, ok := target.(*game.Character); ok {
		healAmount := int(effect.Magnitude)
		oldHP := char.HP
		char.HP = min(char.HP+healAmount, char.MaxHP)
		logger.WithFields(logrus.Fields{
			"targetID":   effect.TargetID,
			"healAmount": healAmount,
			"oldHP":      oldHP,
			"newHP":      char.HP,
		}).Info("healed character")
		return nil
	}

	logger.WithField("targetID", effect.TargetID).Error("target cannot be healed")
	return fmt.Errorf("target cannot be healed")
}

// ProcessStatEffect applies a stat modification effect to a character target.
//
// Parameters:
//   - effect: *game.Effect - Contains the target ID, stat to modify, and magnitude
//     of the modification. Must have valid StatAffected and Magnitude fields.
//
// Returns:
//
//	error - Returns nil if successful, or an error if:
//	- Target ID doesn't exist in WorldState
//	- Target is not a Character type
//	- StatAffected is not a valid stat name
//
// StatAffected must be one of: strength, dexterity, constitution, intelligence,
// wisdom, charisma
//
// Related types:
//   - game.Effect
//   - game.Character
//
// ADDED: processStatEffect applies stat modification effects to target characters.
// It handles both stat boosts and penalties by modifying character attributes.
//
// Supported stats: strength, dexterity, constitution, intelligence, wisdom, charisma
// Effect types: EffectStatBoost (positive) and EffectStatPenalty (negative)
//
// Processing steps:
// 1. Validates target exists and is a Character
// 2. Determines effect sign (boost vs penalty)
// 3. Applies magnitude to specified stat
// 4. Logs stat modification results
//
// Note: Stat modifications are applied directly to Character fields
func (gs *GameState) processStatEffect(effect *game.Effect) error {
	logger := logrus.WithFields(logrus.Fields{
		"function": "processStatEffect",
	})
	logger.Debug("processing stat effect")

	target, exists := gs.WorldState.Objects[effect.TargetID]
	if !exists {
		logger.WithField("targetID", effect.TargetID).Error("invalid effect target")
		return fmt.Errorf("invalid effect target")
	}

	if char, ok := target.(*game.Character); ok {
		magnitude := int(effect.Magnitude)
		logger.WithFields(logrus.Fields{
			"function":  "processStatEffect",
			"targetID":  effect.TargetID,
			"stat":      effect.StatAffected,
			"magnitude": magnitude,
		}).Info("applying stat modification")

		switch effect.StatAffected {
		case "strength":
			char.Strength += magnitude
		case "dexterity":
			char.Dexterity += magnitude
		case "constitution":
			char.Constitution += magnitude
		case "intelligence":
			char.Intelligence += magnitude
		case "wisdom":
			char.Wisdom += magnitude
		case "charisma":
			char.Charisma += magnitude
		default:
			logger.WithField("stat", effect.StatAffected).Error("unknown stat type")
			return fmt.Errorf("unknown stat type: %s", effect.StatAffected)
		}
		return nil
	}

	logger.WithField("targetID", effect.TargetID).Error("target cannot receive stat effects")
	return fmt.Errorf("target cannot receive stat effects")
}
