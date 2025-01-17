package server

import (
	"encoding/json"
	"fmt"
	"sort"

	"goldbox-rpg/pkg/game"

	"golang.org/x/exp/rand"
)

// Additional RPC methods
func (s *RPCServer) handleCastSpell(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string        `json:"session_id"`
		SpellID   string        `json:"spell_id"`
		TargetID  string        `json:"target_id"`
		Position  game.Position `json:"position,omitempty"` // For area spells
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid spell parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Validate spell casting
	player := session.Player
	spell := findSpell(player.KnownSpells, req.SpellID)
	if spell == nil {
		return nil, fmt.Errorf("spell not known")
	}

	// Validate turn if in combat
	if s.state.TurnManager.IsInCombat && !s.state.TurnManager.IsCurrentTurn(player.GetID()) {
		return nil, fmt.Errorf("not your turn")
	}

	// Process spell effects
	result, err := s.processSpellCast(player, spell, req.TargetID, req.Position)
	if err != nil {
		return nil, err
	}

	// Emit spell cast event
	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventSpellCast,
		SourceID: player.GetID(),
		TargetID: req.TargetID,
		Data: map[string]interface{}{
			"spell_id": req.SpellID,
			"position": req.Position,
			"effects":  result,
		},
	})

	return result, nil
}

func (s *RPCServer) handleApplyEffect(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID  string          `json:"session_id"`
		EffectType game.EffectType `json:"effect_type"`
		TargetID   string          `json:"target_id"`
		Magnitude  float64         `json:"magnitude"`
		Duration   game.Duration   `json:"duration"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid effect parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Create and apply the effect
	effect := game.NewEffect(req.EffectType, req.Duration, req.Magnitude)
	effect.SourceID = session.Player.GetID()

	target, exists := s.state.WorldState.Objects[req.TargetID]
	if !exists {
		return nil, fmt.Errorf("invalid target")
	}

	effectHolder, ok := target.(game.EffectHolder)
	if !ok {
		return nil, fmt.Errorf("target cannot receive effects")
	}

	if err := effectHolder.AddEffect(effect); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success":   true,
		"effect_id": effect.ID,
	}, nil
}

func (s *RPCServer) handleStartCombat(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID    string   `json:"session_id"`
		Participants []string `json:"participant_ids"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid combat parameters")
	}

	if s.state.TurnManager.IsInCombat {
		return nil, fmt.Errorf("combat already in progress")
	}

	// Roll initiative for all participants
	initiative := s.rollInitiative(req.Participants)
	s.state.TurnManager.StartCombat(initiative)

	// Emit combat start event
	s.eventSys.Emit(game.GameEvent{
		Type: game.EventCombatStart,
		Data: map[string]interface{}{
			"participants": req.Participants,
			"initiative":   initiative,
		},
	})

	return map[string]interface{}{
		"success":    true,
		"initiative": initiative,
		"first_turn": initiative[0],
	}, nil
}

func (s *RPCServer) handleEndTurn(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid turn parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	if !s.state.TurnManager.IsInCombat {
		return nil, fmt.Errorf("not in combat")
	}

	if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
		return nil, fmt.Errorf("not your turn")
	}

	// Process end of turn effects
	s.processEndTurnEffects(session.Player)

	// Advance to next turn
	nextTurn := s.state.TurnManager.AdvanceTurn()

	// Check for round end
	if s.state.TurnManager.CurrentIndex == 0 {
		s.processEndRound()
	}

	return map[string]interface{}{
		"success":   true,
		"next_turn": nextTurn,
	}, nil
}

func (s *RPCServer) handleGetGameState(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid state request parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Get visible game state for player
	player := session.Player
	visibleObjects := s.getVisibleObjects(player)
	activeEffects := s.getActiveEffects(player)
	combatState := s.getCombatStateIfActive(player)

	return map[string]interface{}{
		"player": map[string]interface{}{
			"position":   player.GetPosition(),
			"stats":      player.GetStats(),
			"effects":    activeEffects,
			"inventory":  player.Inventory,
			"spells":     player.KnownSpells,
			"experience": player.Experience,
		},
		"world": map[string]interface{}{
			"visible_objects": visibleObjects,
			"current_time":    s.state.TimeManager.CurrentTime,
			"combat_state":    combatState,
		},
	}, nil
}

func (s *RPCServer) handleUseItem(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
		ItemID    string `json:"item_id"`
		TargetID  string `json:"target_id,omitempty"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid item parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Validate item ownership and usage
	item := findInventoryItem(session.Player.Inventory, req.ItemID)
	if item == nil {
		return nil, fmt.Errorf("item not found in inventory")
	}

	// Process item usage
	result, err := s.processItemUse(session.Player, item, req.TargetID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Helper functions for RPC methods
func (s *RPCServer) rollInitiative(participants []string) []string {
	type initiativeRoll struct {
		entityID string
		roll     int
	}

	rolls := make([]initiativeRoll, len(participants))
	for i, id := range participants {
		if obj, exists := s.state.WorldState.Objects[id]; exists {
			// Base roll on dexterity if character
			if char, ok := obj.(*game.Character); ok {
				rolls[i] = initiativeRoll{
					entityID: id,
					roll:     rand.Intn(20) + 1 + (char.Dexterity-10)/2,
				}
			} else {
				rolls[i] = initiativeRoll{
					entityID: id,
					roll:     rand.Intn(20) + 1,
				}
			}
		}
	}

	// Sort by initiative roll
	sort.Slice(rolls, func(i, j int) bool {
		return rolls[i].roll > rolls[j].roll
	})

	// Extract sorted IDs
	result := make([]string, len(rolls))
	for i, roll := range rolls {
		result[i] = roll.entityID
	}

	return result
}

func (s *RPCServer) processEndTurnEffects(character game.GameObject) {
	if holder, ok := character.(game.EffectHolder); ok {
		for _, effect := range holder.GetEffects() {
			if effect.ShouldTick(s.state.TimeManager.CurrentTime.RealTime) {
				s.state.processEffectTick(effect)
			}
		}
	}
}

func (s *RPCServer) processEndRound() {
	s.state.TurnManager.RoundCount++
	s.processDelayedActions()
	s.checkCombatEnd()
}

func findSpell(spells []game.Spell, spellID string) *game.Spell {
	for i := range spells {
		if spells[i].ID == spellID {
			return &spells[i]
		}
	}
	return nil
}

func findInventoryItem(inventory []game.Item, itemID string) *game.Item {
	for i := range inventory {
		if inventory[i].ID == itemID {
			return &inventory[i]
		}
	}
	return nil
}
