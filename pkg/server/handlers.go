package server

import (
	"encoding/json"
	"fmt"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// handleMove processes a player movement request in the game world.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - direction: game.Direction enum indicating movement direction
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if move was successful
//   - position: Updated position coordinates
//   - error: Possible errors:
//   - "invalid movement parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
//   - Validation errors from WorldState.ValidateMove
//   - Position setting errors from Player.SetPosition
//
// Related:
//   - game.Direction
//   - game.GameEvent
//   - game.EventMovement
//   - RPCServer.sessions
//   - WorldState.ValidateMove
//   - Player.SetPosition
//   - Player.GetPosition
func (s *RPCServer) handleMove(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleMove",
	}).Debug("entering handleMove")

	var req struct {
		SessionID string         `json:"session_id"`
		Direction game.Direction `json:"direction"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleMove",
			"error":    err.Error(),
		}).Error("failed to unmarshal movement parameters")
		return nil, fmt.Errorf("invalid movement parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleMove",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}

	player := session.Player
	currentPos := player.GetPosition()
	newPos := calculateNewPosition(currentPos, req.Direction)

	logrus.WithFields(logrus.Fields{
		"function": "handleMove",
		"playerID": player.GetID(),
		"from":     currentPos,
		"to":       newPos,
	}).Info("validating player move")

	if err := s.state.WorldState.ValidateMove(player, newPos); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleMove",
			"error":    err.Error(),
		}).Error("move validation failed")
		return nil, err
	}

	if err := player.SetPosition(newPos); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleMove",
			"error":    err.Error(),
		}).Error("failed to set player position")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleMove",
		"playerID": player.GetID(),
	}).Info("emitting movement event")

	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventMovement,
		SourceID: player.GetID(),
		Data: map[string]interface{}{
			"old_position": currentPos,
			"new_position": newPos,
		},
	})

	logrus.WithFields(logrus.Fields{
		"function": "handleMove",
	}).Debug("exiting handleMove")

	return map[string]interface{}{
		"success":  true,
		"position": newPos,
	}, nil
}

// handleAttack processes an attack action during combat in the RPG game.
//
// Parameters:
//   - params: json.RawMessage containing the attack request with:
//   - session_id: string identifier for the player session
//   - target_id: string identifier for the attack target
//   - weapon_id: string identifier for the weapon being used
//
// Returns:
//   - interface{}: The result of the combat action if successful
//   - error: Error if the attack is invalid due to:
//   - Invalid JSON parameters
//   - Invalid session
//   - Not being in combat
//   - Not being the player's turn
//   - Combat action processing errors
//
// Related:
//   - TurnManager.IsInCombat
//   - TurnManager.IsCurrentTurn
//   - processCombatAction
func (s *RPCServer) handleAttack(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleAttack",
	}).Debug("entering handleAttack")

	var req struct {
		SessionID string `json:"session_id"`
		TargetID  string `json:"target_id"`
		WeaponID  string `json:"weapon_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
			"error":    err.Error(),
		}).Error("failed to unmarshal attack parameters")
		return nil, fmt.Errorf("invalid attack parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleAttack",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}

	if !s.state.TurnManager.IsInCombat {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
		}).Warn("attempted attack while not in combat")
		return nil, fmt.Errorf("not in combat")
	}

	if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
			"playerID": session.Player.GetID(),
		}).Warn("player attempted attack when not their turn")
		return nil, fmt.Errorf("not your turn")
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleAttack",
		"playerID": session.Player.GetID(),
		"targetID": req.TargetID,
		"weaponID": req.WeaponID,
	}).Info("processing combat action")

	result, err := s.processCombatAction(session.Player, req.TargetID, req.WeaponID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
			"error":    err.Error(),
		}).Error("combat action failed")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleAttack",
	}).Debug("exiting handleAttack")

	return result, nil
}

// handleCastSpell processes a spell casting request from a client.
// It validates the spell parameters, checks if the spell exists in player's known spells,
// and executes the spell casting logic.
//
// Parameters:
//   - params: Raw JSON message containing:
//   - session_id: Unique identifier for the player session
//   - spell_id: Identifier of the spell to cast
//   - target_id: ID of the target entity (if applicable)
//   - position: Target position for area spells (optional)
//
// Returns:
//   - interface{}: Result of the spell cast operation
//   - error: Error if:
//   - Invalid JSON parameters
//   - Invalid session ID
//   - Spell not found in player's known spells
//   - Spell casting fails (via processSpellCast)
//
// Related:
//   - processSpellCast: Handles the actual spell casting logic
//   - findSpell: Searches for a spell in player's known spells
func (s *RPCServer) handleCastSpell(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleCastSpell",
	}).Debug("entering handleCastSpell")

	var req struct {
		SessionID string        `json:"session_id"`
		SpellID   string        `json:"spell_id"`
		TargetID  string        `json:"target_id"`
		Position  game.Position `json:"position,omitempty"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleCastSpell",
			"error":    err.Error(),
		}).Error("failed to unmarshal spell parameters")
		return nil, fmt.Errorf("invalid spell parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleCastSpell",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}

	player := session.Player
	spell := findSpell(player.KnownSpells, req.SpellID)
	if spell == nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleCastSpell",
			"spellID":  req.SpellID,
			"playerID": player.GetID(),
		}).Warn("spell not found in player's known spells")
		return nil, fmt.Errorf("spell not found")
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleCastSpell",
		"spellID":  req.SpellID,
		"targetID": req.TargetID,
		"playerID": player.GetID(),
	}).Info("attempting to cast spell")

	result, err := s.processSpellCast(player, spell, req.TargetID, req.Position)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleCastSpell",
			"error":    err.Error(),
			"spellID":  req.SpellID,
		}).Error("spell cast failed")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleCastSpell",
	}).Debug("exiting handleCastSpell")

	return result, nil
}

// handleStartCombat initiates a new combat session with the specified participants.
//
// Parameters:
//   - params: Raw JSON message containing:
//   - session_id: Unique identifier for the game session
//   - participant_ids: Array of string IDs for the combat participants
//
// Returns:
//   - interface{}: Map containing:
//   - success: Boolean indicating successful combat start
//   - initiative: Ordered array of participant IDs based on initiative rolls
//   - first_turn: ID of the participant who goes first
//   - error: Error if:
//   - Invalid JSON parameters provided
//   - Combat is already in progress for this session
//
// Related:
//   - TurnManager.StartCombat(): Handles the actual combat state management
//   - rollInitiative(): Determines turn order for participants
func (s *RPCServer) handleStartCombat(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleStartCombat",
	}).Debug("entering handleStartCombat")

	var req struct {
		SessionID    string   `json:"session_id"`
		Participants []string `json:"participant_ids"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleStartCombat",
			"error":    err.Error(),
		}).Error("failed to unmarshal combat parameters")
		return nil, fmt.Errorf("invalid combat parameters")
	}

	if s.state.TurnManager.IsInCombat {
		logrus.WithFields(logrus.Fields{
			"function": "handleStartCombat",
		}).Warn("attempted to start combat while already in combat")
		return nil, fmt.Errorf("combat already in progress")
	}

	logrus.WithFields(logrus.Fields{
		"function":     "handleStartCombat",
		"participants": len(req.Participants),
	}).Info("rolling initiative for combat participants")

	initiative := s.rollInitiative(req.Participants)
	s.state.TurnManager.StartCombat(initiative)

	logrus.WithFields(logrus.Fields{
		"function":  "handleStartCombat",
		"firstTurn": initiative[0],
	}).Info("combat started successfully")

	logrus.WithFields(logrus.Fields{
		"function": "handleStartCombat",
	}).Debug("exiting handleStartCombat")

	return map[string]interface{}{
		"success":    true,
		"initiative": initiative,
		"first_turn": initiative[0],
	}, nil
}

// handleEndTurn processes a request to end the current player's turn in combat.
//
// Params:
//   - params: json.RawMessage containing a session_id field
//
// Returns:
//   - interface{}: A map containing "success" (bool) and "next_turn" with the next player's ID
//   - error: If session is invalid, not in combat, not player's turn, or invalid parameters
//
// Errors:
//   - "invalid turn parameters": If params cannot be unmarshaled
//   - "invalid session": If session ID does not exist
//   - "not in combat": If TurnManager.IsInCombat is false
//   - "not your turn": If current turn does not belong to requesting player
//
// Related:
//   - TurnManager.AdvanceTurn()
//   - processEndTurnEffects()
//   - processEndRound()
func (s *RPCServer) handleEndTurn(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleEndTurn",
	}).Debug("entering handleEndTurn")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleEndTurn",
			"error":    err.Error(),
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid turn parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleEndTurn",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}

	if !s.state.TurnManager.IsInCombat {
		logrus.WithFields(logrus.Fields{
			"function": "handleEndTurn",
		}).Warn("attempted to end turn while not in combat")
		return nil, fmt.Errorf("not in combat")
	}

	if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
		logrus.WithFields(logrus.Fields{
			"function": "handleEndTurn",
			"playerID": session.Player.GetID(),
		}).Warn("player attempted to end turn when not their turn")
		return nil, fmt.Errorf("not your turn")
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleEndTurn",
		"playerID": session.Player.GetID(),
	}).Info("processing end of turn effects")
	s.processEndTurnEffects(session.Player)

	nextTurn := s.state.TurnManager.AdvanceTurn()
	logrus.WithFields(logrus.Fields{
		"function": "handleEndTurn",
		"nextTurn": nextTurn,
	}).Info("advanced to next turn")

	if s.state.TurnManager.CurrentIndex == 0 {
		logrus.WithFields(logrus.Fields{
			"function": "handleEndTurn",
		}).Info("processing end of round")
		s.processEndRound()
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleEndTurn",
	}).Debug("exiting handleEndTurn")

	return map[string]interface{}{
		"success":   true,
		"next_turn": nextTurn,
	}, nil
}

// handleGetGameState processes a request to retrieve the current game state for a given session.
// The method returns a comprehensive snapshot of the player's state and visible world elements.
//
// Parameters:
//   - params: json.RawMessage containing the session_id parameter
//
// Returns:
//   - interface{}: A map containing two main sections:
//   - player: Contains position, stats, active effects, inventory, spells and experience
//   - world: Contains visible objects, current game time and combat state if any
//   - error: Returns error if:
//   - Session ID is invalid or not found
//   - Request parameters cannot be unmarshaled
//
// Related:
//   - Player.GetPosition()
//   - Player.GetStats()
//   - TimeManager.CurrentTime
//   - getVisibleObjects()
//   - getActiveEffects()
//   - getCombatStateIfActive()
func (s *RPCServer) handleGetGameState(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetGameState",
	}).Debug("entering handleGetGameState")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetGameState",
			"error":    err.Error(),
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid state request parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleGetGameState",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}

	player := session.Player
	logrus.WithFields(logrus.Fields{
		"function": "handleGetGameState",
		"playerID": player.GetID(),
	}).Info("retrieving game state for player")

	visibleObjects := s.getVisibleObjects(player)
	activeEffects := s.getActiveEffects(player)
	combatState := s.getCombatStateIfActive(player)

	logrus.WithFields(logrus.Fields{
		"function": "handleGetGameState",
		"objects":  len(visibleObjects),
		"effects":  len(activeEffects),
	}).Info("collected state data")

	logrus.WithFields(logrus.Fields{
		"function": "handleGetGameState",
	}).Debug("exiting handleGetGameState")

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

// handleApplyEffect processes a request to apply an effect to a target entity in the game world.
//
// Parameters:
// - params: json.RawMessage containing the request parameters:
//   - session_id: string identifier for the player session
//   - effect_type: game.EffectType enum specifying the type of effect
//   - target_id: string identifier for the target entity
//   - magnitude: float64 indicating the strength/amount of the effect
//   - duration: game.Duration specifying how long the effect lasts
//
// Returns:
// - interface{}: A map containing:
//   - success: bool indicating if effect was applied
//   - effect_id: string identifier for the created effect
//
// - error: Error if request fails due to:
//   - Invalid JSON parameters
//   - Invalid session ID
//   - Invalid target ID
//   - Target not implementing EffectHolder interface
//   - Effect application failure
//
// Related types:
// - game.Effect
// - game.EffectHolder
// - game.EffectType
// - game.Duration
func (s *RPCServer) handleApplyEffect(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleApplyEffect",
	}).Debug("entering handleApplyEffect")

	var req struct {
		SessionID  string          `json:"session_id"`
		EffectType game.EffectType `json:"effect_type"`
		TargetID   string          `json:"target_id"`
		Magnitude  float64         `json:"magnitude"`
		Duration   game.Duration   `json:"duration"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleApplyEffect",
			"error":    err.Error(),
		}).Error("failed to unmarshal effect parameters")
		return nil, fmt.Errorf("invalid effect parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleApplyEffect",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}

	// Create and apply the effect
	effect := game.NewEffect(req.EffectType, req.Duration, req.Magnitude)
	effect.SourceID = session.Player.GetID()

	logrus.WithFields(logrus.Fields{
		"function":   "handleApplyEffect",
		"effectType": req.EffectType,
		"targetID":   req.TargetID,
	}).Info("creating new effect")

	target, exists := s.state.WorldState.Objects[req.TargetID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function": "handleApplyEffect",
			"targetID": req.TargetID,
		}).Warn("invalid target ID")
		return nil, fmt.Errorf("invalid target")
	}

	effectHolder, ok := target.(game.EffectHolder)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"function": "handleApplyEffect",
			"targetID": req.TargetID,
		}).Warn("target cannot receive effects")
		return nil, fmt.Errorf("target cannot receive effects")
	}

	if err := effectHolder.AddEffect(effect); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleApplyEffect",
			"error":    err.Error(),
		}).Error("failed to add effect")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleApplyEffect",
		"effectID": effect.ID,
	}).Info("effect successfully applied")

	logrus.WithFields(logrus.Fields{
		"function": "handleApplyEffect",
	}).Debug("exiting handleApplyEffect")

	return map[string]interface{}{
		"success":   true,
		"effect_id": effect.ID,
	}, nil
}
