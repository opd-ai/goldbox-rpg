package server

import (
	"fmt"
	"regexp"
	"strconv"
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

func (s *RPCServer) processDelayedActions() {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

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

func (s *RPCServer) applyItemEffects(player *game.Player, item *game.Item, targetID string) ([]game.Effect, error) {
	// Implement the logic to apply item effects here
	// For example, apply the effects to the target
	effects := []game.Effect{}
	// Add logic to apply effects based on item properties
	return effects, nil
}

func (gs *GameState) processDamageEffect(effect *game.Effect) error {
	// Implement the logic for processing damage over time effects here
	// For example, apply the damage to the target
	return nil
}

func (gs *GameState) processHealEffect(effect *game.Effect) error {
	// Implement the logic for processing heal over time effects here
	// For example, apply the healing to the target
	return nil
}

func (gs *GameState) processStatEffect(effect *game.Effect) error {
	// Implement the logic for processing stat effects here
	// For example, apply the stat boost or penalty to the target
	return nil
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

func (s *RPCServer) executeDelayedAction(action DelayedAction) {
	// Implement the logic to execute the delayed action here
	// For example, apply the action's effect to the target
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

func (s *RPCServer) removeItemFromInventory(player *game.Player, item *game.Item) {
	// Implement the logic to remove the item from the player's inventory
	for i, invItem := range player.Inventory {
		if &invItem == item {
			player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
			break
		}
	}
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

func parseDamageString(damage string) int {
	// Regular expression to match dice notation: XdY+Z
	re := regexp.MustCompile(`^(\d+)?d(\d+)(?:\+(\d+))?$`)

	// If it's just a number, return it
	if num, err := strconv.Atoi(damage); err == nil {
		return num
	}

	matches := re.FindStringSubmatch(damage)
	if matches == nil {
		return 0 // Invalid format
	}

	// Parse components
	numDice := 1
	if matches[1] != "" {
		numDice, _ = strconv.Atoi(matches[1])
	}

	dieSize, _ := strconv.Atoi(matches[2])

	modifier := 0
	if matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	// Calculate average damage
	// Average roll on a die is (1 + size) / 2
	averageDamage := int(float64(numDice) * (float64(dieSize) + 1) / 2)
	return averageDamage + modifier
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

func (s *RPCServer) equipItem(player *game.Player, item *game.Item, slot game.EquipmentSlot) error {
	// Implement the logic to equip the item to the specified slot
	player.Equipment[slot] = *item
	return nil
}
