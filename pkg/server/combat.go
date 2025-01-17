package server

import (
	"fmt"
	"goldbox-rpg/pkg/game"
)

// CombatState tracks active combat information
type CombatState struct {
	ActiveCombatants []string                 `yaml:"combat_active_entities"` // Entities in combat
	RoundCount       int                      `yaml:"combat_round_count"`     // Number of rounds
	CombatZone       game.Position            `yaml:"combat_zone_center"`     // Combat area center
	StatusEffects    map[string][]game.Effect `yaml:"combat_status_effects"`  // Active effects
}

// TurnManager handles combat turns and initiative ordering
type TurnManager struct {
	CurrentRound   int                 `yaml:"turn_current_round"`    // Active combat round
	Initiative     []string            `yaml:"turn_initiative_order"` // Turn order by entity ID
	CurrentIndex   int                 `yaml:"turn_current_index"`    // Current actor index
	IsInCombat     bool                `yaml:"turn_in_combat"`        // Combat state flag
	CombatGroups   map[string][]string `yaml:"turn_combat_groups"`    // Allied entities
	DelayedActions []DelayedAction     `yaml:"turn_delayed_actions"`  // Pending actions
}

// DelayedAction represents a pending combat action
type DelayedAction struct {
	ActorID     string        `yaml:"action_actor_id"`     // Entity performing action
	ActionType  string        `yaml:"action_type"`         // Type of action
	Target      game.Position `yaml:"action_target_pos"`   // Target location
	TriggerTime game.GameTime `yaml:"action_trigger_time"` // When to execute
	Parameters  []string      `yaml:"action_parameters"`   // Additional data
}

func (tm *TurnManager) IsCurrentTurn(entityID string) bool {
	if !tm.IsInCombat || tm.CurrentIndex >= len(tm.Initiative) {
		return false
	}
	return tm.Initiative[tm.CurrentIndex] == entityID
}

func (tm *TurnManager) StartCombat(initiative []string) {
	tm.IsInCombat = true
	tm.Initiative = initiative
	tm.CurrentIndex = 0
	tm.CurrentRound = 1
}

func (tm *TurnManager) AdvanceTurn() string {
	if !tm.IsInCombat {
		return ""
	}

	tm.CurrentIndex = (tm.CurrentIndex + 1) % len(tm.Initiative)
	if tm.CurrentIndex == 0 {
		tm.CurrentRound++
	}

	return tm.Initiative[tm.CurrentIndex]
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

func (s *RPCServer) executeDelayedAction(action DelayedAction) {
	// Implement the logic to execute the delayed action here
	// For example, apply the action's effect to the target
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

func (s *RPCServer) applyDamage(target game.GameObject, damage int) error {
	if char, ok := target.(*game.Character); ok {
		char.HP -= damage
		if char.HP < 0 {
			char.HP = 0
		}

		// Check for death
		if char.HP == 0 {
			s.handleCharacterDeath(char)
		}
		return nil
	}
	return fmt.Errorf("target cannot receive damage")
}

func calculateWeaponDamage(weapon *game.Item, attacker *game.Player) int {
	// Basic damage calculation
	baseDamage := parseDamageString(weapon.Damage)
	strBonus := (attacker.Strength - 10) / 2
	return baseDamage + strBonus
}

func (s *RPCServer) handleCharacterDeath(character *game.Character) {
	// Set character as inactive
	character.SetActive(false)

	// Drop inventory items
	dropPosition := character.GetPosition()
	for _, item := range character.Inventory {
		s.state.WorldState.AddObject(CreateItemDrop(item, character, dropPosition))
	}
	character.Inventory = nil

	// Emit death event
	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventDeath,
		SourceID: character.GetID(),
		Data: map[string]interface{}{
			"position": dropPosition,
		},
	})
}

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
func (s *RPCServer) processCombatAction(player *game.Player, targetID, weaponID string) (interface{}, error) {
	target, exists := s.state.WorldState.Objects[targetID]
	if !exists {
		return nil, fmt.Errorf("invalid target")
	}

	// Get weapon from inventory or equipped items
	var weapon *game.Item
	if weaponID != "" {
		weapon = findInventoryItem(player.Inventory, weaponID)
		if weapon == nil && player.Equipment != nil {
			w := player.Equipment[game.SlotHands]
			weapon = &w
		}
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
