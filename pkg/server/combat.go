// Package server implements the game server and combat system functionality
package server

import (
	"fmt"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// CombatState represents the current state of a combat encounter.
// It tracks participating entities, round count, combat area, and active effects.
type CombatState struct {
	// ActiveCombatants contains the IDs of all entities currently in combat
	ActiveCombatants []string `yaml:"combat_active_entities"`
	// RoundCount tracks the current combat round number
	RoundCount int `yaml:"combat_round_count"`
	// CombatZone defines the center position of the combat area
	CombatZone game.Position `yaml:"combat_zone_center"`
	// StatusEffects maps entity IDs to their active effects
	StatusEffects map[string][]game.Effect `yaml:"combat_status_effects"`
}

// TurnManager handles combat turn order and initiative tracking.
// It manages the flow of combat rounds and tracks allied groups.
type TurnManager struct {
	// CurrentRound represents the current combat round number
	CurrentRound int `yaml:"turn_current_round"`
	// Initiative holds entity IDs in their initiative order
	Initiative []string `yaml:"turn_initiative_order"`
	// CurrentIndex tracks the current actor's position in the initiative order
	CurrentIndex int `yaml:"turn_current_index"`
	// IsInCombat indicates whether combat is currently active
	IsInCombat bool `yaml:"turn_in_combat"`
	// CombatGroups maps entity IDs to their allied group members
	CombatGroups map[string][]string `yaml:"turn_combat_groups"`
	// DelayedActions holds actions to be executed at a later time
	DelayedActions []DelayedAction `yaml:"turn_delayed_actions"`
}

// DelayedAction represents a combat action that will be executed at a specific time.
type DelayedAction struct {
	// ActorID is the ID of the entity performing the action
	ActorID string `yaml:"action_actor_id"`
	// ActionType defines the type of action to be performed
	ActionType string `yaml:"action_type"`
	// Target specifies the position where the action will take effect
	Target game.Position `yaml:"action_target_pos"`
	// TriggerTime determines when the action should be executed
	TriggerTime game.GameTime `yaml:"action_trigger_time"`
	// Parameters contains additional data needed for the action
	Parameters []string `yaml:"action_parameters"`
}

// IsCurrentTurn checks if the given entity is the current actor in combat.
//
// Parameters:
//   - entityID: The ID of the entity to check
//
// Returns:
//   - bool: true if it's the entity's turn, false otherwise
func (tm *TurnManager) IsCurrentTurn(entityID string) bool {
	logrus.WithFields(logrus.Fields{
		"function": "IsCurrentTurn",
		"entityID": entityID,
	}).Debug("checking if entity has current turn")

	if !tm.IsInCombat || tm.CurrentIndex >= len(tm.Initiative) {
		logrus.WithFields(logrus.Fields{
			"function":      "IsCurrentTurn",
			"isInCombat":    tm.IsInCombat,
			"currentIndex":  tm.CurrentIndex,
			"initiativeLen": len(tm.Initiative),
		}).Debug("combat inactive or invalid index")
		return false
	}

	isCurrent := tm.Initiative[tm.CurrentIndex] == entityID
	logrus.WithFields(logrus.Fields{
		"function":  "IsCurrentTurn",
		"entityID":  entityID,
		"isCurrent": isCurrent,
	}).Debug("turn check complete")

	return isCurrent
}

// StartCombat initializes a new combat encounter with the given initiative order.
//
// Parameters:
//   - initiative: Ordered slice of entity IDs representing turn order
func (tm *TurnManager) StartCombat(initiative []string) {
	logrus.WithFields(logrus.Fields{
		"function":        "StartCombat",
		"initiativeCount": len(initiative),
	}).Debug("starting new combat")

	tm.IsInCombat = true
	tm.Initiative = initiative
	tm.CurrentIndex = 0
	tm.CurrentRound = 1

	logrus.WithFields(logrus.Fields{
		"function": "StartCombat",
		"round":    tm.CurrentRound,
	}).Info("combat started successfully")
}

// AdvanceTurn moves to the next entity in the initiative order.
// Increments the round counter when returning to the first entity.
//
// Returns:
//   - string: The ID of the next entity in the initiative order, or empty string if not in combat
func (tm *TurnManager) AdvanceTurn() string {
	logrus.WithFields(logrus.Fields{
		"function":   "AdvanceTurn",
		"isInCombat": tm.IsInCombat,
	}).Debug("checking combat state")

	if !tm.IsInCombat {
		logrus.WithFields(logrus.Fields{
			"function": "AdvanceTurn",
		}).Debug("not in combat, returning")
		return ""
	}

	prevIndex := tm.CurrentIndex
	tm.CurrentIndex = (tm.CurrentIndex + 1) % len(tm.Initiative)

	if tm.CurrentIndex == 0 {
		tm.CurrentRound++
		logrus.WithFields(logrus.Fields{
			"function": "AdvanceTurn",
			"round":    tm.CurrentRound,
		}).Info("new combat round started")
	}

	nextEntity := tm.Initiative[tm.CurrentIndex]
	logrus.WithFields(logrus.Fields{
		"function":   "AdvanceTurn",
		"prevIndex":  prevIndex,
		"nextIndex":  tm.CurrentIndex,
		"nextEntity": nextEntity,
	}).Debug("turn advanced")

	return nextEntity
}

// processDelayedActions checks and executes any delayed actions that are due.
// Removes executed actions from the pending actions list.
func (s *RPCServer) processDelayedActions() {
	logrus.WithFields(logrus.Fields{
		"function": "processDelayedActions",
	}).Debug("processing delayed actions")

	currentTime := s.state.TimeManager.CurrentTime
	totalActions := len(s.state.TurnManager.DelayedActions)

	logrus.WithFields(logrus.Fields{
		"function":    "processDelayedActions",
		"currentTime": currentTime,
		"actionCount": totalActions,
	}).Debug("checking delayed actions")

	for i := totalActions - 1; i >= 0; i-- {
		action := s.state.TurnManager.DelayedActions[i]

		logrus.WithFields(logrus.Fields{
			"function":    "processDelayedActions",
			"actionIndex": i,
			"actorID":     action.ActorID,
			"actionType":  action.ActionType,
			"triggerTime": action.TriggerTime,
		}).Debug("checking action timing")

		if isTimeToExecute(currentTime, action.TriggerTime) {
			logrus.WithFields(logrus.Fields{
				"function":   "processDelayedActions",
				"actorID":    action.ActorID,
				"actionType": action.ActionType,
			}).Info("executing delayed action")

			s.executeDelayedAction(action)
			s.state.TurnManager.DelayedActions = append(
				s.state.TurnManager.DelayedActions[:i],
				s.state.TurnManager.DelayedActions[i+1:]...,
			)
		}
	}

	logrus.WithFields(logrus.Fields{
		"function":         "processDelayedActions",
		"remainingActions": len(s.state.TurnManager.DelayedActions),
	}).Debug("finished processing delayed actions")
}

// checkCombatEnd determines if combat should end based on remaining hostile groups.
//
// Returns:
//   - bool: true if combat ended, false if it should continue
func (s *RPCServer) checkCombatEnd() bool {
	logrus.WithFields(logrus.Fields{
		"function":   "checkCombatEnd",
		"isInCombat": s.state.TurnManager.IsInCombat,
	}).Debug("checking if combat should end")

	if !s.state.TurnManager.IsInCombat {
		logrus.WithFields(logrus.Fields{
			"function": "checkCombatEnd",
		}).Debug("not in combat, returning")
		return false
	}

	hostileGroups := s.getHostileGroups()
	logrus.WithFields(logrus.Fields{
		"function":          "checkCombatEnd",
		"hostileGroupCount": len(hostileGroups),
	}).Debug("got hostile groups")

	if len(hostileGroups) <= 1 {
		logrus.WithFields(logrus.Fields{
			"function": "checkCombatEnd",
		}).Info("ending combat - only one or no hostile groups remain")
		s.endCombat()
		return true
	}

	logrus.WithFields(logrus.Fields{
		"function": "checkCombatEnd",
	}).Debug("combat continues")
	return false
}

// executeDelayedAction handles the execution of a delayed combat action.
// Implementation depends on the specific action type.
//
// Parameters:
//   - action: The DelayedAction to execute
func (s *RPCServer) executeDelayedAction(action DelayedAction) {
	// Implement the logic to execute the delayed action here
}

// getHostileGroups returns groups of allied entities in combat.
//
// Returns:
//   - [][]string: Slice of entity ID groups, where each group represents allied entities
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

// endCombat terminates the current combat encounter and emits a combat end event.
func (s *RPCServer) endCombat() {
	s.state.TurnManager.IsInCombat = false
	s.state.TurnManager.Initiative = nil
	s.state.TurnManager.CurrentIndex = 0

	s.eventSys.Emit(game.GameEvent{
		Type: EventCombatEnd,
		Data: map[string]interface{}{
			"rounds_completed": s.state.TurnManager.CurrentRound,
		},
	})
}

// applyDamage applies damage to a game object, handling death if applicable.
//
// Parameters:
//   - target: The GameObject receiving damage
//   - damage: Amount of damage to apply
//
// Returns:
//   - error: Error if target cannot receive damage
func (s *RPCServer) applyDamage(target game.GameObject, damage int) error {
	if char, ok := target.(*game.Character); ok {
		char.HP -= damage
		if char.HP < 0 {
			char.HP = 0
		}

		if char.HP == 0 {
			s.handleCharacterDeath(char)
		}
		return nil
	}
	return fmt.Errorf("target cannot receive damage")
}

// calculateWeaponDamage computes the total damage for a weapon attack.
//
// Parameters:
//   - weapon: The weapon being used
//   - attacker: The attacking player
//
// Returns:
//   - int: Total calculated damage
func calculateWeaponDamage(weapon *game.Item, attacker *game.Player) int {
	baseDamage := parseDamageString(weapon.Damage)
	strBonus := (attacker.Strength - 10) / 2
	return baseDamage + strBonus
}

// handleCharacterDeath processes a character's death, dropping inventory and emitting event.
//
// Parameters:
//   - character: The Character that died
func (s *RPCServer) handleCharacterDeath(character *game.Character) {
	character.SetActive(false)

	dropPosition := character.GetPosition()
	for _, item := range character.Inventory {
		s.state.WorldState.AddObject(CreateItemDrop(item, character, dropPosition))
	}
	character.Inventory = nil

	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventDeath,
		SourceID: character.GetID(),
		Data: map[string]interface{}{
			"position": dropPosition,
		},
	})
}

// CreateItemDrop creates a new item object when dropped from inventory.
//
// Parameters:
//   - item: The item being dropped
//   - char: The character dropping the item
//   - dropPosition: Where the item should be placed
//
// Returns:
//   - game.GameObject: The created item object
func CreateItemDrop(item game.Item, char *game.Character, dropPosition game.Position) game.GameObject {
	return &game.Item{
		ID:         fmt.Sprintf("drop_%s_%s", item.ID, char.GetName()),
		Name:       item.Name,
		Type:       item.Type,
		Damage:     item.Damage,
		AC:         item.AC,
		Weight:     item.Weight,
		Value:      item.Value,
		Properties: item.Properties,
		Position:   dropPosition,
	}
}

// processCombatAction handles weapon attacks during combat.
//
// Parameters:
//   - player: The attacking player
//   - targetID: ID of the attack target
//   - weaponID: ID of the weapon to use (optional)
//
// Returns:
//   - interface{}: Combat result containing success and damage
//   - error: Error if target is invalid or attack fails
func (s *RPCServer) processCombatAction(player *game.Player, targetID, weaponID string) (interface{}, error) {
	target, exists := s.state.WorldState.Objects[targetID]
	if !exists {
		return nil, fmt.Errorf("invalid target")
	}

	var weapon *game.Item
	if weaponID != "" {
		weapon = findInventoryItem(player.Inventory, weaponID)
		if weapon == nil && player.Equipment != nil {
			w := player.Equipment[game.SlotHands]
			weapon = &w
		}
	}

	damage := calculateWeaponDamage(weapon, player)

	if err := s.applyDamage(target, damage); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"damage":  damage,
	}, nil
}
