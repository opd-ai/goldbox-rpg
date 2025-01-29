package server

import (
	"fmt"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// processEffectTick handles the processing of a single effect tick in the game state.
// It determines the effect type and routes the processing to the appropriate handler.
//
// Parameters:
//   - effect: *game.Effect - The effect to process. Must not be nil.
//
// Returns:
//   - error: Returns nil on success, or an error if:
//   - The effect parameter is nil
//   - The effect type is unknown/unsupported
//
// Related:
//   - processDamageEffect
//   - processHealEffect
//   - processStatEffect
//
// Handles effect types:
//   - EffectDamageOverTime
//   - EffectHealOverTime
//   - EffectStatBoost
//   - EffectStatPenalty
func (gs *GameState) processEffectTick(effect *game.Effect) error {
	logger := logrus.WithFields(logrus.Fields{
		"function": "processEffectTick",
	})
	logger.Debug("processing effect tick")

	if effect == nil {
		logger.Error("nil effect provided")
		return fmt.Errorf("nil effect")
	}

	logger.WithFields(logrus.Fields{
		"effectType": effect.Type,
		"targetID":   effect.TargetID,
	}).Info("processing effect")

	var err error
	switch effect.Type {
	case game.EffectDamageOverTime:
		err = gs.processDamageEffect(effect)
	case game.EffectHealOverTime:
		err = gs.processHealEffect(effect)
	case game.EffectStatBoost, game.EffectStatPenalty:
		err = gs.processStatEffect(effect)
	default:
		logger.WithField("effectType", effect.Type).Warn("unknown effect type")
		return fmt.Errorf("unknown effect type: %s", effect.Type)
	}

	if err != nil {
		logger.WithError(err).Error("failed to process effect")
	}
	return err
}

// processDamageEffect applies damage to a target character based on the provided effect.
// It locates the target in the world state and reduces their HP by the effect magnitude.
//
// Parameters:
//   - effect: *game.Effect - Contains target ID and damage magnitude to apply
//
// Returns:
//   - error - Returns nil if damage was successfully applied, or an error if:
//   - Target ID does not exist in world state
//   - Target is not a Character type that can receive damage
//
// Related:
//   - game.Character
//   - game.Effect
//   - GameState.WorldState
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
