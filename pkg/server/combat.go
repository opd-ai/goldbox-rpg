// Package server implements the game server and combat system functionality
package server

import (
	"fmt"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// DefaultTurnDuration defines the default time limit for combat turns.
// Players have this amount of time to complete their actions before the turn automatically ends.
var DefaultTurnDuration = 10 * time.Second

// CombatState represents the current state of an active combat encounter.
// It tracks all participating entities, combat progression, and environmental effects.
//
// Fields:
//   - ActiveCombatants: IDs of all entities currently participating in combat
//   - RoundCount: Current combat round number (increments each full initiative cycle)
//   - CombatZone: Center position defining the combat area boundaries
//   - StatusEffects: Map of entity IDs to their currently active effects
//
// Combat lifecycle:
// 1. Combat starts with participants entering initiative order
// 2. Round counter tracks combat progression
// 3. Status effects are managed per-entity
// 4. Combat zone defines spatial boundaries for the encounter
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

// TurnManager handles combat turn sequencing and initiative tracking.
// It manages the flow of combat rounds, turn timeouts, and coordinated group actions.
//
// Core responsibilities:
// - Initiative order management and turn progression
// - Combat round tracking and state transitions
// - Turn timer enforcement with configurable durations
// - Allied group coordination for team-based combat
// - Delayed action scheduling and execution
//
// Fields:
//   - CurrentRound: Current combat round number
//   - Initiative: Ordered list of entity IDs by initiative roll
//   - CurrentIndex: Position in initiative order (whose turn it is)
//   - IsInCombat: Flag indicating active combat state
//   - CombatGroups: Maps entity IDs to their allied group members
//   - DelayedActions: Queue of actions scheduled for future execution
//   - turnTimer: Internal timer for enforcing turn time limits
//   - turnDuration: Configurable duration for each turn
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
	turnTimer      *time.Timer     // Timer for turn timeouts
	turnDuration   time.Duration   // Duration for turn timeouts
}

// NewTurnManager creates and initializes a new TurnManager instance.
// It sets up the turn management system with default values and empty state.
//
// Returns:
//   - *TurnManager: Fully initialized turn manager ready for combat
//
// Initial state:
// - CurrentRound: 0 (no combat started)
// - Initiative: Empty slice (no participants)
// - CurrentIndex: 0 (first position)
// - IsInCombat: false (not in combat)
// - Empty combat groups and delayed actions
// - Default turn duration configuration
func NewTurnManager() *TurnManager {
	return &TurnManager{
		CurrentRound:   0,
		Initiative:     []string{},
		CurrentIndex:   0,
		IsInCombat:     false,
		CombatGroups:   make(map[string][]string),
		DelayedActions: make([]DelayedAction, 0),
		turnTimer:      nil, // Initialize as nil, will be set when combat starts
		turnDuration:   DefaultTurnDuration,
	}
}

// Update applies the provided updates to the TurnManager.
//
// Parameters:
//   - turnUpdates: Map of field names to their new values
//
// Returns:
//   - any: Updated TurnManager instance
func (tm *TurnManager) Update(turnUpdates map[string]interface{}) error {
	logrus.WithFields(logrus.Fields{
		"function": "Update",
	}).Debug("updating turn manager state")

	// Update fields if present in updates map
	if round, ok := turnUpdates["current_round"].(int); ok {
		tm.CurrentRound = round
	}

	if initiative, ok := turnUpdates["initiative_order"].([]string); ok {
		// Always validate initiative order to prevent corruption
		if err := tm.validateInitiativeOrder(initiative); err != nil {
			logrus.WithFields(logrus.Fields{
				"function": "Update",
				"error":    err.Error(),
			}).Error("invalid initiative order in update")
			return fmt.Errorf("failed to update initiative: %w", err)
		}
		tm.Initiative = initiative
	}

	if index, ok := turnUpdates["current_index"].(int); ok {
		tm.CurrentIndex = index
	}

	if inCombat, ok := turnUpdates["in_combat"].(bool); ok {
		tm.IsInCombat = inCombat
	}

	if groups, ok := turnUpdates["combat_groups"].(map[string][]string); ok {
		tm.CombatGroups = groups
	}

	if actions, ok := turnUpdates["delayed_actions"].([]DelayedAction); ok {
		tm.DelayedActions = actions
	}

	logrus.WithFields(logrus.Fields{
		"function": "Update",
	}).Debug("turn manager state updated")

	return nil
}

// Clone creates and returns a deep copy of the TurnManager
func (tm *TurnManager) Clone() *TurnManager {
	// Create new TurnManager
	clone := &TurnManager{
		CurrentRound:   tm.CurrentRound,
		CurrentIndex:   tm.CurrentIndex,
		IsInCombat:     tm.IsInCombat,
		Initiative:     make([]string, len(tm.Initiative)),
		CombatGroups:   make(map[string][]string),
		DelayedActions: make([]DelayedAction, len(tm.DelayedActions)),
	}

	// Copy initiative slice
	copy(clone.Initiative, tm.Initiative)

	// Deep copy combat groups map
	for k, v := range tm.CombatGroups {
		groupCopy := make([]string, len(v))
		copy(groupCopy, v)
		clone.CombatGroups[k] = groupCopy
	}

	// Copy delayed actions
	copy(clone.DelayedActions, tm.DelayedActions)

	return clone
}

// Serialize returns a map representation of the TurnManager state.
func (tm *TurnManager) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"current_round":    tm.CurrentRound,
		"initiative_order": tm.Initiative,
		"current_index":    tm.CurrentIndex,
		"in_combat":        tm.IsInCombat,
		"combat_groups":    tm.CombatGroups,
		"delayed_actions":  tm.DelayedActions,
	}
}

// DelayedAction represents a combat action scheduled for future execution.
// It enables complex combat mechanics like spell casting times and triggered abilities.
//
// Use cases:
// - Spell casting with casting times (e.g., 3-second fireball)
// - Triggered abilities that activate later
// - Environmental effects with delays
// - Coordinated group actions
//
// Fields:
//   - ActorID: Entity performing the action
//   - ActionType: Type of action (spell, attack, movement, etc.)
//   - Target: Position where action takes effect
//   - TriggerTime: Game time when action should execute
//   - Parameters: Additional action-specific data
//
// The action queue processes DelayedActions when their TriggerTime is reached.
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
func (tm *TurnManager) StartCombat(initiative []string) error {
	logrus.WithFields(logrus.Fields{
		"function":        "StartCombat",
		"initiativeCount": len(initiative),
	}).Debug("starting new combat")

	// Validate initiative order
	if err := tm.validateInitiativeOrder(initiative); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "StartCombat",
			"error":    err.Error(),
		}).Error("invalid initiative order")
		return fmt.Errorf("failed to start combat: %w", err)
	}

	tm.IsInCombat = true
	tm.Initiative = initiative
	tm.CurrentIndex = 0
	tm.CurrentRound = 1
	tm.startTurnTimer()

	logrus.WithFields(logrus.Fields{
		"function": "StartCombat",
		"round":    tm.CurrentRound,
	}).Info("combat started successfully")
	return nil
}

func (tm *TurnManager) startTurnTimer() {
	if tm.turnTimer != nil {
		tm.turnTimer.Stop()
	}
	tm.turnTimer = time.AfterFunc(tm.turnDuration, tm.endTurn)
}

func (tm *TurnManager) endTurn() {
	// Check if initiative is valid before accessing it
	if len(tm.Initiative) == 0 || tm.CurrentIndex >= len(tm.Initiative) {
		logrus.WithFields(logrus.Fields{
			"function":      "endTurn",
			"currentIndex":  tm.CurrentIndex,
			"initiativeLen": len(tm.Initiative),
		}).Error("invalid initiative state during endTurn")
		return
	}

	currentActor := tm.Initiative[tm.CurrentIndex]

	// Check if actor took action
	actorHasAction := false
	for _, action := range tm.DelayedActions {
		if action.ActorID == currentActor {
			actorHasAction = true
			break
		}
	}

	if !actorHasAction {
		if err := tm.moveToTopOfInitiative(currentActor); err != nil {
			logrus.WithFields(logrus.Fields{
				"function":     "endTurn",
				"currentActor": currentActor,
				"error":        err.Error(),
			}).Error("failed to move actor to top of initiative")
			// Continue without reordering if validation fails
		}
	}

	// Process delayed actions
	tm.processDelayedActions()

	// Advance turn
	tm.CurrentIndex = (tm.CurrentIndex + 1) % len(tm.Initiative)
	if tm.CurrentIndex == 0 {
		tm.CurrentRound++
	}

	if tm.IsInCombat {
		tm.startTurnTimer()
	}
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

	// Check if initiative is valid before accessing it
	if len(tm.Initiative) == 0 {
		logrus.WithFields(logrus.Fields{
			"function": "AdvanceTurn",
		}).Error("initiative is empty during AdvanceTurn")
		return ""
	}

	// Ensure CurrentIndex is within bounds
	if tm.CurrentIndex >= len(tm.Initiative) {
		logrus.WithFields(logrus.Fields{
			"function":      "AdvanceTurn",
			"currentIndex":  tm.CurrentIndex,
			"initiativeLen": len(tm.Initiative),
		}).Error("CurrentIndex out of bounds, resetting to 0")
		tm.CurrentIndex = 0
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
	logrus.WithFields(logrus.Fields{
		"function": "getHostileGroups",
	}).Debug("getting hostile groups")

	groups := make([][]string, 0)
	processed := make(map[string]bool)

	logrus.WithFields(logrus.Fields{
		"function":    "getHostileGroups",
		"groupsCount": len(s.state.TurnManager.CombatGroups),
	}).Debug("processing combat groups")

	for id := range s.state.TurnManager.CombatGroups {
		if !processed[id] {
			group := s.state.TurnManager.CombatGroups[id]
			groups = append(groups, group)
			for _, memberID := range group {
				processed[memberID] = true
			}

			logrus.WithFields(logrus.Fields{
				"function":     "getHostileGroups",
				"groupLeader":  id,
				"membersCount": len(group),
			}).Debug("processed group")
		}
	}

	logrus.WithFields(logrus.Fields{
		"function":          "getHostileGroups",
		"hostileGroupCount": len(groups),
	}).Info("hostile groups identified")

	return groups
}

// endCombat terminates the current combat encounter and emits a combat end event.
func (s *RPCServer) endCombat() {
	logrus.WithFields(logrus.Fields{
		"function": "endCombat",
	}).Debug("ending combat")

	// Stop the turn timer if it's running
	if s.state.TurnManager.turnTimer != nil {
		s.state.TurnManager.turnTimer.Stop()
		s.state.TurnManager.turnTimer = nil
	}

	s.state.TurnManager.IsInCombat = false
	s.state.TurnManager.Initiative = nil
	s.state.TurnManager.CurrentIndex = 0

	logrus.WithFields(logrus.Fields{
		"function": "endCombat",
		"rounds":   s.state.TurnManager.CurrentRound,
	}).Info("combat ended")

	s.eventSys.Emit(game.GameEvent{
		Type: EventCombatEnd,
		Data: map[string]interface{}{
			"rounds_completed": s.state.TurnManager.CurrentRound,
		},
	})

	logrus.WithFields(logrus.Fields{
		"function": "endCombat",
	}).Debug("combat cleanup complete")
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
	logrus.WithFields(logrus.Fields{
		"function": "applyDamage",
		"damage":   damage,
		"targetID": target.GetID(),
	}).Debug("applying damage to target")

	// Handle both Character and Player types
	var char *game.Character
	if player, ok := target.(*game.Player); ok {
		char = &player.Character
	} else if character, ok := target.(*game.Character); ok {
		char = character
	} else {
		err := fmt.Errorf("target cannot receive damage")
		logrus.WithFields(logrus.Fields{
			"function": "applyDamage",
			"error":    err.Error(),
		}).Error("invalid target type")
		return err
	}

	oldHP := char.HP
	char.HP -= damage

	if char.HP < 0 {
		logrus.WithFields(logrus.Fields{
			"function": "applyDamage",
			"charID":   char.GetID(),
		}).Debug("clamping HP to 0")
		char.HP = 0
	}

	logrus.WithFields(logrus.Fields{
		"function": "applyDamage",
		"charID":   char.GetID(),
		"oldHP":    oldHP,
		"newHP":    char.HP,
		"damage":   damage,
	}).Info("damage applied to character")

	if char.HP == 0 {
		logrus.WithFields(logrus.Fields{
			"function": "applyDamage",
			"charID":   char.GetID(),
		}).Info("character died from damage")
		s.handleCharacterDeath(char)
	}
	return nil
}

// ADDED: calculateWeaponDamage computes total damage output for a weapon attack.
// It combines base weapon damage with attacker bonuses and modifiers.
//
// Damage calculation:
// 1. Base weapon damage (from weapon item properties)
// 2. Strength bonus for melee weapons
// 3. Dexterity bonus for ranged weapons
// 4. Weapon proficiency bonuses
// 5. Additional enchantments or effects
//
// Parameters:
//   - weapon: Weapon item being used (nil for unarmed attacks)
//   - attacker: Player character making the attack
//
// Returns:
//   - int: Total calculated damage value
//
// Special cases: Handles unarmed attacks when weapon is nil
func calculateWeaponDamage(weapon *game.Item, attacker *game.Player) int {
	// Handle nil weapon (unarmed attack)
	if weapon == nil {
		logrus.WithFields(logrus.Fields{
			"function":   "calculateWeaponDamage",
			"weaponID":   "unarmed",
			"attackerID": attacker.GetID(),
		}).Debug("calculating unarmed damage")

		// Unarmed attack: 1 + Strength bonus
		strBonus := (attacker.Strength - 10) / 2
		unarmedDamage := 1 + strBonus
		if unarmedDamage < 1 {
			unarmedDamage = 1 // Minimum 1 damage
		}

		logrus.WithFields(logrus.Fields{
			"function":    "calculateWeaponDamage",
			"baseDamage":  1,
			"strBonus":    strBonus,
			"totalDamage": unarmedDamage,
		}).Info("unarmed damage calculation completed")

		return unarmedDamage
	}

	logrus.WithFields(logrus.Fields{
		"function":   "calculateWeaponDamage",
		"weaponID":   weapon.ID,
		"attackerID": attacker.GetID(),
	}).Debug("calculating weapon damage")

	baseDamage := parseDamageString(weapon.Damage)
	strBonus := (attacker.Strength - 10) / 2

	logrus.WithFields(logrus.Fields{
		"function":    "calculateWeaponDamage",
		"baseDamage":  baseDamage,
		"strBonus":    strBonus,
		"totalDamage": baseDamage + strBonus,
	}).Info("damage calculation completed")

	return baseDamage + strBonus
}

// handleCharacterDeath processes a character's death, dropping inventory and emitting event.
//
// Parameters:
//   - character: The Character that died
func (s *RPCServer) handleCharacterDeath(character *game.Character) {
	logrus.WithFields(logrus.Fields{
		"function":    "handleCharacterDeath",
		"characterID": character.GetID(),
	}).Debug("handling character death")

	character.SetActive(false)
	dropPosition := character.GetPosition()

	logrus.WithFields(logrus.Fields{
		"function":     "handleCharacterDeath",
		"characterID":  character.GetID(),
		"dropPosition": dropPosition,
		"itemCount":    len(character.Inventory),
	}).Info("processing inventory drops")

	for _, item := range character.Inventory {
		logrus.WithFields(logrus.Fields{
			"function": "handleCharacterDeath",
			"itemID":   item.ID,
		}).Debug("dropping item")
		s.state.WorldState.AddObject(CreateItemDrop(item, character, dropPosition))
	}
	character.Inventory = nil

	logrus.WithFields(logrus.Fields{
		"function":    "handleCharacterDeath",
		"characterID": character.GetID(),
	}).Info("emitting death event")

	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventDeath,
		SourceID: character.GetID(),
		Data: map[string]interface{}{
			"position": dropPosition,
		},
	})

	logrus.WithFields(logrus.Fields{
		"function": "handleCharacterDeath",
	}).Debug("character death handling complete")
}

// CreateItemDrop creates a new item GameObject when an item is dropped from inventory.
// It handles the transition from inventory item to world object with proper positioning.
//
// Creation process:
// 1. Creates new GameObject wrapper around the item
// 2. Sets the drop position in the game world
// 3. Assigns unique identifier for world tracking
// 4. Configures item as interactive world object
//
// Parameters:
//   - item: The game.Item being dropped from inventory
//   - char: The character who is dropping the item
//   - dropPosition: World position where item should appear
//
// Returns:
//   - game.GameObject: New item object ready for world placement
//
// The returned object can be added to the world's object collection for player interaction.
func CreateItemDrop(item game.Item, char *game.Character, dropPosition game.Position) game.GameObject {
	logrus.WithFields(logrus.Fields{
		"function":     "CreateItemDrop",
		"itemID":       item.ID,
		"characterID":  char.GetID(),
		"dropPosition": dropPosition,
	}).Debug("creating new item drop")

	droppedItem := &game.Item{
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

	logrus.WithFields(logrus.Fields{
		"function":    "CreateItemDrop",
		"droppedID":   droppedItem.ID,
		"droppedName": droppedItem.Name,
	}).Info("item drop created")

	return droppedItem
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
	logrus.WithFields(logrus.Fields{
		"function": "processCombatAction",
		"playerID": player.GetID(),
		"targetID": targetID,
		"weaponID": weaponID,
	}).Debug("processing combat action")

	target, exists := s.state.WorldState.Objects[targetID]
	if !exists {
		err := fmt.Errorf("invalid target")
		logrus.WithFields(logrus.Fields{
			"function": "processCombatAction",
			"error":    err.Error(),
		}).Error("target not found")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "processCombatAction",
		"targetID": targetID,
	}).Debug("found valid target")

	var weapon *game.Item
	if weaponID != "" {
		weapon = findInventoryItem(player.Inventory, weaponID)
		if weapon == nil && player.Equipment != nil {
			logrus.WithFields(logrus.Fields{
				"function": "processCombatAction",
			}).Debug("checking equipped weapon")
			w := player.Equipment[game.SlotHands]
			weapon = &w
		}
	}

	damage := calculateWeaponDamage(weapon, player)
	logrus.WithFields(logrus.Fields{
		"function": "processCombatAction",
		"damage":   damage,
	}).Info("calculated weapon damage")

	if err := s.applyDamage(target, damage); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "processCombatAction",
			"error":    err.Error(),
		}).Error("failed to apply damage")
		return nil, err
	}

	result := map[string]interface{}{
		"success": true,
		"damage":  damage,
	}

	logrus.WithFields(logrus.Fields{
		"function": "processCombatAction",
		"damage":   damage,
	}).Debug("combat action completed successfully")

	return result, nil
}

// QueueAction adds a delayed action to the turn manager's queue.
func (tm *TurnManager) QueueAction(action DelayedAction) error {
	logger := logrus.WithFields(logrus.Fields{
		"function": "QueueAction",
		"actorID":  action.ActorID,
	})

	if !tm.IsCurrentTurn(action.ActorID) {
		logger.Warn("attempt to queue action on wrong turn")
		return fmt.Errorf("not actor's turn")
	}

	action.TriggerTime = game.GameTime{
		RealTime:  time.Now(),
		GameTicks: tm.getCurrentGameTicks(),
		TimeScale: 1.0,
	}

	logger.WithField("triggerTime", action.TriggerTime).Debug("queueing delayed action")
	tm.DelayedActions = append(tm.DelayedActions, action)
	return nil
}

func (tm *TurnManager) moveToTopOfInitiative(entityID string) error {
	// Find group members
	group := append([]string{entityID}, tm.CombatGroups[entityID]...)

	// Create new initiative order
	newOrder := make([]string, 0, len(tm.Initiative))
	newOrder = append(newOrder, group...)

	for _, id := range tm.Initiative {
		inGroup := false
		for _, gid := range group {
			if id == gid {
				inGroup = true
				break
			}
		}
		if !inGroup {
			newOrder = append(newOrder, id)
		}
	}

	// Validate the new order before applying it
	if err := tm.validateInitiativeOrder(newOrder); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "moveToTopOfInitiative",
			"entityID": entityID,
			"error":    err.Error(),
		}).Error("failed to reorder initiative")
		return fmt.Errorf("initiative reorder failed: %w", err)
	}

	tm.Initiative = newOrder
	tm.CurrentIndex = 0
	return nil
}

func (tm *TurnManager) processDelayedActions() {
	currentTime := game.GameTime{
		RealTime:  time.Now(),
		GameTicks: tm.getCurrentGameTicks(),
	}

	remainingActions := make([]DelayedAction, 0)
	for _, action := range tm.DelayedActions {
		if currentTime.IsSameTurn(action.TriggerTime) {
			logrus.WithField("action", action).Debug("processing delayed action")
		} else {
			remainingActions = append(remainingActions, action)
		}
	}
	tm.DelayedActions = remainingActions
}

func (tm *TurnManager) getCurrentGameTicks() int64 {
	return int64(tm.CurrentRound*6+tm.CurrentIndex) * 10
}

// EndCombat terminates the current combat encounter and cleans up timers.
func (tm *TurnManager) EndCombat() {
	logrus.WithFields(logrus.Fields{
		"function": "EndCombat",
	}).Debug("ending combat via TurnManager")

	// Stop the turn timer if it's running
	if tm.turnTimer != nil {
		tm.turnTimer.Stop()
		tm.turnTimer = nil
	}

	tm.IsInCombat = false
	tm.Initiative = nil
	tm.CurrentIndex = 0

	logrus.WithFields(logrus.Fields{
		"function": "EndCombat",
		"rounds":   tm.CurrentRound,
	}).Info("combat ended via TurnManager")
}

// validateInitiativeOrder ensures the initiative slice contains valid entity IDs without duplicates.
//
// Parameters:
//   - initiative: The initiative order slice to validate
//
// Returns:
//   - error: Error if validation fails, nil if valid
func (tm *TurnManager) validateInitiativeOrder(initiative []string) error {
	if len(initiative) == 0 {
		return fmt.Errorf("initiative order cannot be empty when starting combat")
	}

	// Check for duplicate entity IDs
	seen := make(map[string]bool)
	for _, entityID := range initiative {
		if entityID == "" {
			return fmt.Errorf("initiative order contains empty entity ID")
		}
		if seen[entityID] {
			return fmt.Errorf("initiative order contains duplicate entity ID: %s", entityID)
		}
		seen[entityID] = true
	}

	return nil
}
