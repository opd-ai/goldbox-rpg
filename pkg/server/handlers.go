package server

import (
	"encoding/json"
	"fmt"
	"goldbox-rpg/pkg/game"
)

// Movement handler
func (s *RPCServer) handleMove(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string         `json:"session_id"`
		Direction game.Direction `json:"direction"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid movement parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	player := session.Player
	currentPos := player.GetPosition()
	newPos := calculateNewPosition(currentPos, req.Direction)

	if err := s.state.WorldState.ValidateMove(player, newPos); err != nil {
		return nil, err
	}

	if err := player.SetPosition(newPos); err != nil {
		return nil, err
	}

	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventMovement,
		SourceID: player.GetID(),
		Data: map[string]interface{}{
			"old_position": currentPos,
			"new_position": newPos,
		},
	})

	return map[string]interface{}{
		"success":  true,
		"position": newPos,
	}, nil
}

// Combat action handler
func (s *RPCServer) handleAttack(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
		TargetID  string `json:"target_id"`
		WeaponID  string `json:"weapon_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid attack parameters")
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

	result, err := s.processCombatAction(session.Player, req.TargetID, req.WeaponID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *RPCServer) handleCastSpell(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string        `json:"session_id"`
		SpellID   string        `json:"spell_id"`
		TargetID  string        `json:"target_id"`
		Position  game.Position `json:"position,omitempty"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid spell parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	player := session.Player
	spell := findSpell(player.KnownSpells, req.SpellID)
	if spell == nil {
		return nil, fmt.Errorf("spell not found")
	}

	result, err := s.processSpellCast(player, spell, req.TargetID, req.Position)
	if err != nil {
		return nil, err
	}

	return result, nil
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

	initiative := s.rollInitiative(req.Participants)
	s.state.TurnManager.StartCombat(initiative)

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

	s.processEndTurnEffects(session.Player)
	nextTurn := s.state.TurnManager.AdvanceTurn()

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
