package server

import (
	"fmt"
	"time"

	"goldbox-rpg/pkg/game"
)

func NewTimeManager() *TimeManager {
	return &TimeManager{
		CurrentTime: game.GameTime{
			RealTime:  time.Now(),
			GameTicks: 0,
			TimeScale: 1.0,
		},
		TimeScale:       1.0,
		LastTick:        time.Now(),
		ScheduledEvents: make([]ScheduledEvent, 0),
	}
}

// Add these methods to GameState
func (gs *GameState) processEffectTick(effect *game.Effect) error {
	if effect == nil {
		return fmt.Errorf("nil effect")
	}

	switch effect.Type {
	case game.EffectDamageOverTime:
		return gs.processDamageEffect(effect)
	case game.EffectHealOverTime:
		return gs.processHealEffect(effect)
	case game.EffectStatBoost, game.EffectStatPenalty:
		return gs.processStatEffect(effect)
	default:
		return fmt.Errorf("unknown effect type: %s", effect.Type)
	}
}

// Add to RPCServer
func (s *RPCServer) processItemUse(player *game.Player, item *game.Item, targetID string) (interface{}, error) {
	switch item.Type {
	case game.ItemTypeWeapon:
		return s.processWeaponUse(player, item, targetID)
	case game.ItemTypeArmor:
		return s.processArmorUse(player, item)
	default:
		return s.processConsumableUse(player, item, targetID)
	}
}

func (s *RPCServer) processDelayedActions() {
	currentTime := s.state.TimeManager.CurrentTime

	for i := len(s.state.TurnManager.DelayedActions) - 1; i >= 0; i-- {
		action := s.state.TurnManager.DelayedActions[i]
		if isTimeToExecute(currentTime, action.TriggerTime) {
			s.executeDelayedAction(action)
			// Remove executed action
			s.state.TurnManager.DelayedActions = append(
				s.state.TurnManager.DelayedActions[:i],
				s.state.TurnManager.DelayedActions[i+1:]...,
			)
		}
	}
}

func (s *RPCServer) checkCombatEnd() bool {
	if !s.state.TurnManager.IsInCombat {
		return false
	}

	// Check if combat should end
	hostileGroups := s.getHostileGroups()
	if len(hostileGroups) <= 1 {
		s.endCombat()
		return true
	}
	return false
}

// Add helper methods
func (s *RPCServer) processWeaponUse(player *game.Player, weapon *game.Item, targetID string) (interface{}, error) {
	target, exists := s.state.WorldState.Objects[targetID]
	if !exists {
		return nil, fmt.Errorf("invalid target")
	}

	// Calculate damage
	damage := calculateWeaponDamage(weapon, player)

	// Apply damage to target
	if err := s.applyDamage(target, damage); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"damage":  damage,
	}, nil
}

func (s *RPCServer) processArmorUse(player *game.Player, armor *game.Item) (interface{}, error) {
	// Equip armor
	slot := determineArmorSlot(armor)
	if err := s.equipItem(player, armor, slot); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"slot":    slot,
	}, nil
}

func (s *RPCServer) processConsumableUse(player *game.Player, item *game.Item, targetID string) (interface{}, error) {
	// Apply item effects
	effects, err := s.applyItemEffects(player, item, targetID)
	if err != nil {
		return nil, err
	}

	// Remove consumed item
	s.removeItemFromInventory(player, item)

	return map[string]interface{}{
		"success": true,
		"effects": effects,
	}, nil
}

func (s *RPCServer) getHostileGroups() [][]string {
	groups := make([][]string, 0)
	processed := make(map[string]bool)

	for id := range s.state.TurnManager.CombatGroups {
		if !processed[id] {
			group := s.state.TurnManager.CombatGroups[id]
			groups = append(groups, group)
			for _, memberID := range group {
				processed[memberID] = true
			}
		}
	}

	return groups
}

func (s *RPCServer) endCombat() {
	s.state.TurnManager.IsInCombat = false
	s.state.TurnManager.Initiative = nil
	s.state.TurnManager.CurrentIndex = 0

	// Emit combat end event
	s.eventSys.Emit(game.GameEvent{
		Type: EventCombatEnd,
		Data: map[string]interface{}{
			"rounds_completed": s.state.TurnManager.CurrentRound,
		},
	})
}

func isTimeToExecute(current, trigger game.GameTime) bool {
	return current.GameTicks >= trigger.GameTicks
}

func calculateWeaponDamage(weapon *game.Item, attacker *game.Player) int {
	// Basic damage calculation
	baseDamage := parseDamageString(weapon.Damage)
	strBonus := (attacker.Strength - 10) / 2
	return baseDamage + strBonus
}

func determineArmorSlot(armor *game.Item) game.EquipmentSlot {
	// Determine appropriate slot based on armor type
	switch armor.Type {
	case "helmet":
		return game.SlotHead
	case "chest":
		return game.SlotChest
	case "gloves":
		return game.SlotHands
	case "boots":
		return game.SlotFeet
	default:
		return game.SlotChest
	}
}
