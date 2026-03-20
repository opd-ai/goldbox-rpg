package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ErrInvalidSession is
var ErrInvalidSession = errors.New("invalid session")

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
	logrus.WithField("function", "handleMove").Debug("entering handleMove")

	result, err := s.executeMoveAction(params)
	s.recordActionMetrics("move", err)
	if err != nil {
		return nil, err
	}

	logrus.WithField("function", "handleMove").Debug("exiting handleMove")
	return result, nil
}

// executeMoveAction performs the core movement logic.
func (s *RPCServer) executeMoveAction(params json.RawMessage) (map[string]interface{}, error) {
	req, err := s.parseMoveRequest(params)
	if err != nil {
		return nil, err
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, err
	}
	defer s.releaseSession(session)

	if err := s.validateCombatConstraints(session.Player); err != nil {
		return nil, err
	}

	newPos, err := s.calculateAndValidateNewPosition(session.Player, req.Direction)
	if err != nil {
		return nil, err
	}

	if err := s.consumeMovementActionPoints(session.Player); err != nil {
		return nil, err
	}

	if err := s.executePlayerMovement(session.Player, newPos); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"success":  true,
		"position": newPos,
	}

	// Check for tile interactions (Gold Box-style environmental feedback)
	interaction, interactionText := s.state.GetInteractionText(newPos.Level, newPos.X, newPos.Y)
	if interactionText != "" {
		result["interaction"] = interaction
		result["interaction_text"] = interactionText
	}

	return result, nil
}

// recordActionMetrics records player action metrics for success or failure.
func (s *RPCServer) recordActionMetrics(action string, err error) {
	if s.metrics == nil {
		return
	}
	status := "success"
	if err != nil {
		status = "failed"
	}
	s.metrics.RecordPlayerAction(action, status)
}

// parseMoveRequest extracts and validates movement request parameters from JSON.
// Supports both integer direction values and string direction names ("north", "south", "east", "west").
func (s *RPCServer) parseMoveRequest(params json.RawMessage) (*struct {
	SessionID string         `json:"session_id"`
	Direction game.Direction `json:"direction"`
}, error,
) {
	// First try parsing with Direction as integer
	var req struct {
		SessionID string         `json:"session_id"`
		Direction game.Direction `json:"direction"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		// Try parsing with direction as a string (sent by WASM client)
		var strReq struct {
			SessionID string `json:"session_id"`
			Direction string `json:"direction"`
		}
		if err2 := json.Unmarshal(params, &strReq); err2 != nil {
			logrus.WithFields(logrus.Fields{
				"function": "parseMoveRequest",
				"error":    err.Error(),
			}).Error("failed to unmarshal movement parameters")
			return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid movement parameters", err.Error())
		}

		// Convert string direction to Direction enum
		dir, ok := parseDirectionString(strReq.Direction)
		if !ok {
			return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid direction", fmt.Sprintf("unknown direction: %s", strReq.Direction))
		}
		req.SessionID = strReq.SessionID
		req.Direction = dir
	}

	return &req, nil
}

// parseDirectionString converts a string direction name to a game.Direction value.
func parseDirectionString(dir string) (game.Direction, bool) {
	switch strings.ToLower(strings.TrimSpace(dir)) {
	case "north", "n":
		return game.DirectionNorth, true
	case "east", "e":
		return game.DirectionEast, true
	case "south", "s":
		return game.DirectionSouth, true
	case "west", "w":
		return game.DirectionWest, true
	default:
		return 0, false
	}
}

// getSessionForMove retrieves and validates the player session for movement.
func (s *RPCServer) getSessionForMove(sessionID string) (*PlayerSession, error) {
	session, err := s.getSessionSafely(sessionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function":  "getSessionForMove",
			"sessionID": sessionID,
		}).Warn("invalid session ID")
		return nil, NewSessionError(sessionID, "getSessionForMove", ErrInvalidSession)
	}
	// fix: guard against nil Player to prevent nil pointer dereference
	if session.Player == nil {
		s.releaseSession(session)
		return nil, NewSessionError(sessionID, "getSessionForMove", fmt.Errorf("session has no player"))
	}
	return session, nil
}

// validateCombatConstraints checks turn order and action point requirements during combat.
// Also validates that the player is not stunned (cannot act) or rooted (cannot move).
func (s *RPCServer) validateCombatConstraints(player *game.Player) error {
	if !s.state.TurnManager.IsInCombat {
		return nil
	}

	// Check for stun effect - stunned characters cannot perform any actions
	if player.HasEffect(game.EffectStun) {
		logrus.WithFields(logrus.Fields{
			"function": "validateCombatConstraints",
			"playerID": player.GetID(),
		}).Warn("player attempted to move while stunned")
		return NewValidationError("move", "stunned", player.GetID(), errors.New("cannot act while stunned"))
	}

	// Check for root effect - rooted characters cannot move
	if player.HasEffect(game.EffectRoot) {
		logrus.WithFields(logrus.Fields{
			"function": "validateCombatConstraints",
			"playerID": player.GetID(),
		}).Warn("player attempted to move while rooted")
		return NewValidationError("move", "rooted", player.GetID(), errors.New("cannot move while rooted"))
	}

	if !s.state.TurnManager.IsCurrentTurn(player.GetID()) {
		logrus.WithFields(logrus.Fields{
			"function": "validateCombatConstraints",
			"playerID": player.GetID(),
		}).Warn("player attempted to move when not their turn")
		return NewValidationError("move", "turn_order", player.GetID(), errors.New("not your turn"))
	}

	if player.GetActionPoints() < game.ActionCostMove {
		logrus.WithFields(logrus.Fields{
			"function":   "validateCombatConstraints",
			"playerID":   player.GetID(),
			"currentAP":  player.GetActionPoints(),
			"requiredAP": game.ActionCostMove,
		}).Warn("player attempted to move without enough action points")
		return NewValidationError("move", "action_points", player.GetActionPoints(),
			fmt.Errorf("insufficient action points for movement (need %d, have %d)",
				game.ActionCostMove, player.GetActionPoints()))
	}

	return nil
}

// calculateAndValidateNewPosition computes the target position and validates the move.
func (s *RPCServer) calculateAndValidateNewPosition(player *game.Player, direction game.Direction) (game.Position, error) {
	currentPos := player.GetPosition()
	newPos := calculateNewPosition(currentPos, direction, s.state.WorldState.Width, s.state.WorldState.Height)

	logrus.WithFields(logrus.Fields{
		"function": "calculateAndValidateNewPosition",
		"playerID": player.GetID(),
		"from":     currentPos,
		"to":       newPos,
	}).Info("validating player move")

	if err := s.state.WorldState.ValidateMove(player, newPos); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "calculateAndValidateNewPosition",
			"error":    err.Error(),
		}).Error("move validation failed")
		return game.Position{}, fmt.Errorf("move validation failed for player %s from %v to %v: %w", player.GetID(), currentPos, newPos, err)
	}

	return newPos, nil
}

// consumeMovementActionPoints deducts action points for movement during combat.
func (s *RPCServer) consumeMovementActionPoints(player *game.Player) error {
	if !s.state.TurnManager.IsInCombat {
		return nil
	}

	if !player.ConsumeActionPoints(game.ActionCostMove) {
		logrus.WithFields(logrus.Fields{
			"function": "consumeMovementActionPoints",
			"playerID": player.GetID(),
		}).Error("failed to consume action points before movement")
		return NewValidationError("move", "action_points", player.GetActionPoints(),
			fmt.Errorf("action point consumption failed for player %s", player.GetID()))
	}

	logrus.WithFields(logrus.Fields{
		"function":    "consumeMovementActionPoints",
		"playerID":    player.GetID(),
		"consumedAP":  game.ActionCostMove,
		"remainingAP": player.GetActionPoints(),
	}).Info("consumed action points for movement")

	return nil
}

// executePlayerMovement updates player position and emits movement event.
func (s *RPCServer) executePlayerMovement(player *game.Player, newPos game.Position) error {
	currentPos := player.GetPosition()

	if err := player.SetPosition(newPos); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "executePlayerMovement",
			"error":    err.Error(),
		}).Error("failed to set player position")
		return fmt.Errorf("failed to set player %s position to %v: %w", player.GetID(), newPos, err)
	}

	logrus.WithFields(logrus.Fields{
		"function": "executePlayerMovement",
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

	return nil
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

	req, err := s.parseAttackRequest(params)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordPlayerAction("attack", "failed")
		}
		return nil, err
	}

	session, err := s.validateAttackSession(req.SessionID)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordPlayerAction("attack", "failed")
		}
		return nil, err
	}
	defer s.releaseSession(session)

	if err := s.validateAttackCombatState(session); err != nil {
		if s.metrics != nil {
			s.metrics.RecordPlayerAction("attack", "failed")
		}
		return nil, err
	}

	result, err := s.executeAttack(session, req.TargetID, req.WeaponID)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordPlayerAction("attack", "failed")
		}
		return nil, err
	}

	if s.metrics != nil {
		s.metrics.RecordPlayerAction("attack", "success")
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleAttack",
	}).Debug("exiting handleAttack")

	return result, nil
}

// attackRequest holds the parsed attack request parameters.
type attackRequest struct {
	SessionID string `json:"session_id"`
	TargetID  string `json:"target_id"`
	WeaponID  string `json:"weapon_id"`
}

// parseAttackRequest parses and validates the attack request.
func (s *RPCServer) parseAttackRequest(params json.RawMessage) (*attackRequest, error) {
	var req attackRequest
	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
			"error":    err.Error(),
		}).Error("failed to unmarshal attack parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid attack parameters", err.Error())
	}
	return &req, nil
}

// validateAttackSession validates the session for attack request.
func (s *RPCServer) validateAttackSession(sessionID string) (*PlayerSession, error) {
	session, err := s.getSessionSafely(sessionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function":  "handleAttack",
			"sessionID": sessionID,
		}).Warn("invalid session ID")
		return nil, NewSessionError(sessionID, "handleAttack", ErrInvalidSession)
	}
	// fix: guard against nil Player to prevent nil pointer dereference
	if session.Player == nil {
		s.releaseSession(session)
		return nil, NewSessionError(sessionID, "handleAttack", fmt.Errorf("session has no player"))
	}
	return session, nil
}

// validateAttackCombatState validates combat state requirements for attacking.
// Checks that player is in combat, it's their turn, they have enough AP, and not stunned.
func (s *RPCServer) validateAttackCombatState(session *PlayerSession) error {
	if !s.state.TurnManager.IsInCombat {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
		}).Warn("attempted attack while not in combat")
		return NewValidationError("attack", "combat_state", s.state.TurnManager.IsInCombat, errors.New("not in combat"))
	}

	// Check for stun effect - stunned characters cannot attack
	if session.Player.HasEffect(game.EffectStun) {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
			"playerID": session.Player.GetID(),
		}).Warn("player attempted to attack while stunned")
		return NewValidationError("attack", "stunned", session.Player.GetID(), errors.New("cannot act while stunned"))
	}

	if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
			"playerID": session.Player.GetID(),
		}).Warn("player attempted attack when not their turn")
		return NewValidationError("attack", "turn_order", session.Player.GetID(), errors.New("not your turn"))
	}

	if session.Player.GetActionPoints() < game.ActionCostAttack {
		logrus.WithFields(logrus.Fields{
			"function":   "handleAttack",
			"playerID":   session.Player.GetID(),
			"currentAP":  session.Player.GetActionPoints(),
			"requiredAP": game.ActionCostAttack,
		}).Warn("player attempted to attack without enough action points")
		return NewValidationError("attack", "action_points", session.Player.GetActionPoints(),
			fmt.Errorf("insufficient action points for attack (need %d, have %d)",
				game.ActionCostAttack, session.Player.GetActionPoints()))
	}

	return nil
}

// executeAttack performs the attack action and consumes action points.
func (s *RPCServer) executeAttack(session *PlayerSession, targetID, weaponID string) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleAttack",
		"playerID": session.Player.GetID(),
		"targetID": targetID,
		"weaponID": weaponID,
	}).Info("processing combat action")

	result, err := s.processCombatAction(session.Player, targetID, weaponID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
			"error":    err.Error(),
		}).Error("combat action failed")
		return nil, fmt.Errorf("combat action failed for player %s attacking %s: %w",
			session.Player.GetID(), targetID, err)
	}

	if !session.Player.ConsumeActionPoints(game.ActionCostAttack) {
		logrus.WithFields(logrus.Fields{
			"function": "handleAttack",
			"playerID": session.Player.GetID(),
		}).Error("failed to consume action points after attack validation")
		return nil, fmt.Errorf("action point consumption failed")
	}

	logrus.WithFields(logrus.Fields{
		"function":    "handleAttack",
		"playerID":    session.Player.GetID(),
		"consumedAP":  game.ActionCostAttack,
		"remainingAP": session.Player.GetActionPoints(),
	}).Info("consumed action points for attack")

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
	logrus.WithField("function", "handleCastSpell").Debug("entering handleCastSpell")

	result, err := s.executeCastSpellAction(params)
	s.recordActionMetrics("cast_spell", err)
	if err != nil {
		return nil, err
	}

	logrus.WithField("function", "handleCastSpell").Debug("exiting handleCastSpell")
	return result, nil
}

// executeCastSpellAction performs the core spell casting logic.
func (s *RPCServer) executeCastSpellAction(params json.RawMessage) (interface{}, error) {
	req, err := s.parseCastSpellRequest(params)
	if err != nil {
		return nil, err
	}

	session, err := s.validateSpellCastSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	defer s.releaseSession(session)

	if err := s.validateCombatConstraintsForSpell(session.Player); err != nil {
		return nil, err
	}

	spell, err := s.validatePlayerSpellKnowledge(session.Player, req.SpellID)
	if err != nil {
		return nil, err
	}

	// Check and consume spell slot (cantrips don't use slots)
	if spell.Level > 0 {
		if !session.Player.HasSpellSlot(spell.Level) {
			return nil, NewJSONRPCError(JSONRPCInvalidParams,
				fmt.Sprintf("No level %d spell slots remaining", spell.Level), nil)
		}
		if err := session.Player.UseSpellSlot(spell.Level); err != nil {
			return nil, NewJSONRPCError(JSONRPCInvalidParams, err.Error(), nil)
		}
	}

	result, err := s.executeSpellCast(session.Player, spell, req.TargetID, req.Position)
	if err != nil {
		return nil, err
	}

	// Emit spell cast event for real-time updates
	if s.eventSys != nil {
		s.eventSys.Emit(game.GameEvent{
			Type:     game.EventSpellCast,
			SourceID: session.Player.GetID(),
			TargetID: req.TargetID,
			Data: map[string]interface{}{
				"spell_id":    req.SpellID,
				"spell_name":  spell.Name,
				"spell_level": spell.Level,
				"caster_name": session.Player.Name,
			},
		})
	}

	if err := s.consumeSpellCastActionPoints(session.Player); err != nil {
		return nil, err
	}

	return result, nil
}

// parseCastSpellRequest extracts and validates spell casting parameters from JSON.
func (s *RPCServer) parseCastSpellRequest(params json.RawMessage) (*struct {
	SessionID string        `json:"session_id"`
	SpellID   string        `json:"spell_id"`
	TargetID  string        `json:"target_id"`
	Position  game.Position `json:"position,omitempty"`
}, error,
) {
	var req struct {
		SessionID string        `json:"session_id"`
		SpellID   string        `json:"spell_id"`
		TargetID  string        `json:"target_id"`
		Position  game.Position `json:"position,omitempty"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseCastSpellRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal spell parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid spell parameters", err.Error())
	}

	return &req, nil
}

// validateSpellCastSession retrieves and validates the player session for spell casting.
func (s *RPCServer) validateSpellCastSession(sessionID string) (*PlayerSession, error) {
	session, err := s.getSessionSafely(sessionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function":  "validateSpellCastSession",
			"sessionID": sessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}
	// fix: guard against nil Player to prevent nil pointer dereference
	if session.Player == nil {
		s.releaseSession(session)
		return nil, fmt.Errorf("session has no player")
	}
	return session, nil
}

// validateCombatConstraintsForSpell checks combat turn order and action points for spell casting.
// Also validates that the player is not stunned.
func (s *RPCServer) validateCombatConstraintsForSpell(player *game.Player) error {
	// Check for stun effect - stunned characters cannot cast spells (even outside combat)
	if player.HasEffect(game.EffectStun) {
		logrus.WithFields(logrus.Fields{
			"function": "validateCombatConstraintsForSpell",
			"playerID": player.GetID(),
		}).Warn("player attempted to cast spell while stunned")
		return fmt.Errorf("cannot act while stunned")
	}

	// Check if currently in combat (spells can also be cast outside combat)
	if !s.state.TurnManager.IsInCombat {
		return nil // No combat constraints when not in combat
	}

	// If in combat, validate turn order
	if !s.state.TurnManager.IsCurrentTurn(player.GetID()) {
		logrus.WithFields(logrus.Fields{
			"function": "validateCombatConstraintsForSpell",
			"playerID": player.GetID(),
		}).Warn("player attempted to cast spell when not their turn")
		return fmt.Errorf("not your turn")
	}

	// Check if player has enough action points for spell casting
	if player.GetActionPoints() < game.ActionCostSpell {
		logrus.WithFields(logrus.Fields{
			"function":   "validateCombatConstraintsForSpell",
			"playerID":   player.GetID(),
			"currentAP":  player.GetActionPoints(),
			"requiredAP": game.ActionCostSpell,
		}).Warn("player attempted to cast spell without enough action points")
		return fmt.Errorf("insufficient action points for spell casting (need %d, have %d)",
			game.ActionCostSpell, player.GetActionPoints())
	}

	return nil
}

// validatePlayerSpellKnowledge checks if the player knows the requested spell.
func (s *RPCServer) validatePlayerSpellKnowledge(player *game.Player, spellID string) (*game.Spell, error) {
	spell, err := s.spellManager.GetSpell(spellID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "validatePlayerSpellKnowledge",
			"spellID":  spellID,
			"playerID": player.GetID(),
		}).Warn("spell not found in spell database")
		return nil, fmt.Errorf("spell not found: %s", spellID)
	}

	// Check if player knows this spell
	if !player.KnowsSpell(spellID) {
		logrus.WithFields(logrus.Fields{
			"function": "validatePlayerSpellKnowledge",
			"playerID": player.GetID(),
			"spellID":  spellID,
		}).Warn("player does not know this spell")
		return nil, fmt.Errorf("you do not know this spell: %s", spell.Name)
	}

	return spell, nil
}

// executeSpellCast performs the actual spell casting operation.
func (s *RPCServer) executeSpellCast(player *game.Player, spell *game.Spell, targetID string, position game.Position) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "executeSpellCast",
		"spellID":  spell.ID,
		"targetID": targetID,
		"playerID": player.GetID(),
	}).Info("attempting to cast spell")

	result, err := s.processSpellCast(player, spell, targetID, position)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "executeSpellCast",
			"error":    err.Error(),
			"spellID":  spell.ID,
		}).Error("spell cast failed")
		return nil, err
	}

	return result, nil
}

// consumeSpellCastActionPoints deducts action points after successful spell casting.
func (s *RPCServer) consumeSpellCastActionPoints(player *game.Player) error {
	// Consume action points if in combat
	if !s.state.TurnManager.IsInCombat {
		return nil // No action point consumption when not in combat
	}

	if !player.ConsumeActionPoints(game.ActionCostSpell) {
		// This should not happen due to earlier validation, but safety check
		logrus.WithFields(logrus.Fields{
			"function": "consumeSpellCastActionPoints",
			"playerID": player.GetID(),
		}).Error("failed to consume action points after spell validation")
		return fmt.Errorf("action point consumption failed")
	}

	logrus.WithFields(logrus.Fields{
		"function":    "consumeSpellCastActionPoints",
		"playerID":    player.GetID(),
		"consumedAP":  game.ActionCostSpell,
		"remainingAP": player.GetActionPoints(),
	}).Info("consumed action points for spell casting")

	return nil
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

	req, err := s.parseStartCombatRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.validateCombatNotActive(); err != nil {
		return nil, err
	}

	participants := s.resolveCombatParticipants(req.SessionID, req.Participants)
	initiative, err := s.initiateCombat(participants)
	if err != nil {
		return nil, err
	}

	s.initializeParticipantActionPoints(initiative)
	s.emitCombatStartEvent(req.SessionID, initiative)

	logrus.WithFields(logrus.Fields{
		"function": "handleStartCombat",
	}).Debug("exiting handleStartCombat")

	return s.buildCombatStartResponse(initiative), nil
}

// startCombatRequest holds the parsed start combat request parameters.
type startCombatRequest struct {
	SessionID    string   `json:"session_id"`
	Participants []string `json:"participant_ids"`
}

// parseStartCombatRequest parses and validates the start combat request.
func (s *RPCServer) parseStartCombatRequest(params json.RawMessage) (*startCombatRequest, error) {
	var req startCombatRequest
	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleStartCombat",
			"error":    err.Error(),
		}).Error("failed to unmarshal combat parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid combat parameters", err.Error())
	}
	return &req, nil
}

// validateCombatNotActive checks if combat is not already in progress.
func (s *RPCServer) validateCombatNotActive() error {
	if s.state.TurnManager.IsInCombat {
		logrus.WithFields(logrus.Fields{
			"function": "handleStartCombat",
		}).Warn("attempted to start combat while already in combat")
		return fmt.Errorf("combat already in progress")
	}
	return nil
}

// resolveCombatParticipants determines the combat participants.
// If no participants are provided, auto-populates with the session's player.
func (s *RPCServer) resolveCombatParticipants(sessionID string, participants []string) []string {
	if len(participants) > 0 {
		return participants
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if session, exists := s.sessions[sessionID]; exists && session.Player != nil {
		logrus.WithFields(logrus.Fields{
			"function":   "handleStartCombat",
			"session_id": sessionID,
			"player_id":  session.Player.GetID(),
		}).Info("auto-populated combat participants with session player")
		return []string{session.Player.GetID()}
	}
	return participants
}

// initiateCombat rolls initiative and starts combat via TurnManager.
func (s *RPCServer) initiateCombat(participants []string) ([]string, error) {
	logrus.WithFields(logrus.Fields{
		"function":     "handleStartCombat",
		"participants": len(participants),
	}).Info("rolling initiative for combat participants")

	// Register NPCs with the morale system for combat
	s.registerNPCsForCombat(participants)

	initiative := s.rollInitiative(participants)
	if err := s.state.TurnManager.StartCombat(initiative); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleStartCombat",
			"error":    err.Error(),
		}).Error("failed to start combat")
		return nil, fmt.Errorf("failed to start combat: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":  "handleStartCombat",
		"firstTurn": initiative[0],
	}).Info("combat started successfully")

	return initiative, nil
}

// registerNPCsForCombat registers all NPCs in the participant list with the morale system.
func (s *RPCServer) registerNPCsForCombat(participants []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, entityID := range participants {
		// Skip if this is a player
		isPlayer := false
		for _, session := range s.sessions {
			if session.Player != nil && session.Player.GetID() == entityID {
				isPlayer = true
				break
			}
		}
		if isPlayer {
			continue
		}

		// Register NPC with morale system (default faction "enemies", not a leader, 80 initial morale)
		// The morale system will determine the morale state based on the score
		s.moraleSystem.RegisterNPC(entityID, "enemies", false, 80)

		logrus.WithFields(logrus.Fields{
			"function": "registerNPCsForCombat",
			"npcID":    entityID,
		}).Debug("registered NPC with morale system")
	}
}

// initializeParticipantActionPoints restores action points for all combat participants.
func (s *RPCServer) initializeParticipantActionPoints(initiative []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, participantID := range initiative {
		for _, session := range s.sessions {
			if session.Player != nil && session.Player.GetID() == participantID {
				session.Player.RestoreActionPoints()
				logrus.WithFields(logrus.Fields{
					"function":      "handleStartCombat",
					"participantID": participantID,
					"actionPoints":  session.Player.GetActionPoints(),
				}).Info("initialized action points for combat participant")
				break
			}
		}
	}
}

// emitCombatStartEvent emits the combat_start event for WebSocket subscribers.
func (s *RPCServer) emitCombatStartEvent(sessionID string, initiative []string) {
	s.eventSys.Emit(game.GameEvent{
		Type:     EventCombatStart,
		SourceID: sessionID,
		Data: map[string]interface{}{
			"initiative":   initiative,
			"first_turn":   initiative[0],
			"participants": len(initiative),
		},
	})
}

// buildCombatStartResponse builds the response for a successful combat start.
func (s *RPCServer) buildCombatStartResponse(initiative []string) map[string]interface{} {
	// Build full initiative entries with HP, MaxHP, Name, IsPlayer, and MoraleState
	initiativeEntries := s.buildInitiativeEntries(initiative)

	return map[string]interface{}{
		"success":    true,
		"initiative": initiativeEntries,
		"first_turn": initiative[0],
		"combat_state": map[string]interface{}{
			"is_in_combat":      s.state.TurnManager.IsInCombat,
			"current_index":     s.state.TurnManager.CurrentIndex,
			"initiative_order":  initiative,
			"current_round":     s.state.TurnManager.CurrentRound,
			"active_combatants": initiative,
		},
	}
}

// buildInitiativeEntries creates full initiative entry data including HP, MaxHP, name, morale state, and effects.
func (s *RPCServer) buildInitiativeEntries(initiative []string) []map[string]interface{} {
	entries := make([]map[string]interface{}, 0, len(initiative))

	for i, entityID := range initiative {
		entry := map[string]interface{}{
			"id":         entityID,
			"name":       entityID,            // Default to ID as name
			"initiative": len(initiative) - i, // Higher initiative = earlier in order
			"is_player":  false,
		}

		// Try to find entity in world objects to get HP, name, and effects
		if obj, exists := s.state.WorldState.Objects[entityID]; exists {
			// Get name from object
			if named, ok := obj.(interface{ GetName() string }); ok {
				entry["name"] = named.GetName()
			}

			// Check if it's a Character (has HP and effects)
			if char, ok := obj.(*game.Character); ok {
				entry["hp"] = char.HP
				entry["max_hp"] = char.MaxHP
				entry["effects"] = s.buildEffectData(char)
			}
		}

		// Check sessions to see if this is a player
		s.mu.RLock()
		for _, session := range s.sessions {
			if session.Player != nil && session.Player.GetID() == entityID {
				entry["is_player"] = true
				entry["name"] = session.Player.Name
				entry["hp"] = session.Player.HP
				entry["max_hp"] = session.Player.MaxHP
				entry["effects"] = s.buildEffectData(&session.Player.Character)
				break
			}
		}
		s.mu.RUnlock()

		// Get morale state for NPCs (not players)
		if !entry["is_player"].(bool) {
			moraleState := s.moraleSystem.GetMoraleState(entityID)
			entry["morale_state"] = moraleStateToString(moraleState)
		}

		entries = append(entries, entry)
	}

	return entries
}

// buildEffectData creates effect data from a character's effects for JSON serialization.
func (s *RPCServer) buildEffectData(char *game.Character) []map[string]interface{} {
	if char.EffectManager == nil {
		return nil
	}

	effects := char.EffectManager.GetEffects()
	if len(effects) == 0 {
		return nil
	}

	effectData := make([]map[string]interface{}, 0, len(effects))
	for _, effect := range effects {
		effectData = append(effectData, map[string]interface{}{
			"id":        effect.ID,
			"name":      effect.Name,
			"type":      string(effect.Type),
			"duration":  effect.Duration.Turns,
			"magnitude": int(effect.Magnitude),
			"source_id": effect.SourceID,
			"is_active": effect.IsActive,
			"stacks":    effect.Stacks,
		})
	}
	return effectData
}

// moraleStateToString converts a MoraleState to its string representation.
func moraleStateToString(state game.MoraleState) string {
	switch state {
	case game.MoraleSteadfast:
		return "Steadfast"
	case game.MoraleShaken:
		return "Shaken"
	case game.MoraleBroken:
		return "Broken"
	case game.MoralePanicked:
		return "Panicked"
	default:
		return "Steadfast"
	}
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

	sessionID, err := s.parseEndTurnRequest(params)
	if err != nil {
		return nil, err
	}

	session, err := s.validateEndTurnSession(sessionID)
	if err != nil {
		return nil, err
	}
	defer s.releaseSession(session)

	if err := s.validateEndTurnCombatState(session); err != nil {
		return nil, err
	}

	nextTurn := s.processEndTurn(session)
	s.emitTurnEndEvent(session.Player.GetID(), nextTurn)

	logrus.WithFields(logrus.Fields{
		"function": "handleEndTurn",
	}).Debug("exiting handleEndTurn")

	return map[string]interface{}{
		"success":   true,
		"next_turn": nextTurn,
	}, nil
}

// parseEndTurnRequest parses the end turn request parameters.
func (s *RPCServer) parseEndTurnRequest(params json.RawMessage) (string, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleEndTurn",
			"error":    err.Error(),
		}).Error("failed to unmarshal request parameters")
		return "", NewJSONRPCError(JSONRPCInvalidParams, "Invalid turn parameters", err.Error())
	}
	return req.SessionID, nil
}

// validateEndTurnSession validates the session for end turn request.
func (s *RPCServer) validateEndTurnSession(sessionID string) (*PlayerSession, error) {
	session, exists := s.getSession(sessionID)
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleEndTurn",
			"sessionID": sessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}
	// fix: guard against nil Player to prevent nil pointer dereference
	if session.Player == nil {
		s.releaseSession(session)
		return nil, fmt.Errorf("session has no player")
	}
	return session, nil
}

// validateEndTurnCombatState validates combat state requirements for ending a turn.
func (s *RPCServer) validateEndTurnCombatState(session *PlayerSession) error {
	if !s.state.TurnManager.IsInCombat {
		logrus.WithFields(logrus.Fields{
			"function": "handleEndTurn",
		}).Warn("attempted to end turn while not in combat")
		return fmt.Errorf("not in combat")
	}

	if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
		logrus.WithFields(logrus.Fields{
			"function": "handleEndTurn",
			"playerID": session.Player.GetID(),
		}).Warn("player attempted to end turn when not their turn")
		return fmt.Errorf("not your turn")
	}
	return nil
}

// processEndTurn handles the end of turn logic and returns the next player's turn.
func (s *RPCServer) processEndTurn(session *PlayerSession) string {
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

	s.restoreNextPlayerActionPoints(nextTurn)
	s.checkAndProcessEndRound()

	return nextTurn
}

// restoreNextPlayerActionPoints restores action points for the next player.
func (s *RPCServer) restoreNextPlayerActionPoints(nextTurn string) {
	if nextTurn == "" {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, nextSession := range s.sessions {
		if nextSession.Player != nil && nextSession.Player.GetID() == nextTurn {
			nextSession.Player.RestoreActionPoints()
			logrus.WithFields(logrus.Fields{
				"function":     "handleEndTurn",
				"nextPlayerID": nextTurn,
				"restoredAP":   nextSession.Player.GetActionPoints(),
			}).Info("restored action points for next player")
			break
		}
	}
}

// checkAndProcessEndRound checks if the round has ended and processes it.
func (s *RPCServer) checkAndProcessEndRound() {
	if s.state.TurnManager.CurrentIndex == 0 {
		logrus.WithFields(logrus.Fields{
			"function": "handleEndTurn",
		}).Info("processing end of round")
		s.processEndRound()
	}
}

// emitTurnEndEvent emits the turn_end event for WebSocket subscribers.
func (s *RPCServer) emitTurnEndEvent(playerID, nextTurn string) {
	s.eventSys.Emit(game.GameEvent{
		Type:     EventTurnEnd,
		SourceID: playerID,
		Data: map[string]interface{}{
			"player_id":     playerID,
			"next_turn":     nextTurn,
			"current_round": s.state.TurnManager.CurrentRound,
		},
	})
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
/*func (s *RPCServer) handleGetGameState(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetGameState",
	})
	logger.Debug("entering handleGetGameState")

	// 1. Validate params
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).Error("failed to unmarshal parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid parameters", err.Error())
	}

	// 2. Validate session
	if req.SessionID == "" {
		logger.Warn("invalid session ID")
		return nil, ErrInvalidSession
	}

	// 3. Validate server state
	if s.state == nil {
		logger.Error("game state not initialized")
		return nil, fmt.Errorf("server state not initialized")
	}

	// 4. Get and validate session
	session, exists := s.getSession(req.SessionID)

	if !exists {
		logger.WithField("sessionID", req.SessionID).Warn("session not found")
		return nil, ErrInvalidSession
	}

	// 5. Get game state
	session.LastActive = time.Now()
	state := s.state.GetState()

	// 6. Validate response
	if state == nil {
		logger.Error("failed to get game state")
		return nil, fmt.Errorf("internal server error")
	}

	logger.Debug("exiting handleGetGameState")
	return state, nil
}*/

// handleRest processes a rest action, restoring spell slots and recovering HP.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if rest was successful
//   - slots_restored: bool indicating if spell slots were restored
//   - hp_restored: int amount of HP recovered
//   - message: string describing the result
//   - error: Error if session is invalid or player is in combat
func (s *RPCServer) handleRest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleRest",
	})
	logger.Debug("entering handleRest")

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).Error("failed to unmarshal parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid parameters", err.Error())
	}

	session, err := s.getSessionSafely(req.SessionID)
	if err != nil {
		logger.WithField("sessionID", req.SessionID).Warn("session not found")
		return nil, ErrInvalidSession
	}
	defer s.releaseSession(session)

	// Check if player is in combat - cannot rest during combat
	if s.state != nil && s.state.TurnManager.IsInCombat {
		logger.Warn("cannot rest during combat")
		return map[string]interface{}{
			"success": false,
			"message": "Cannot rest during combat",
		}, nil
	}

	// Restore spell slots and HP
	var hpRestored int
	if session.Player != nil {
		session.Player.RestoreSpellSlots()
		logger.Info("spell slots restored")

		// Restore HP based on character level (1 HP per level, minimum 1)
		hpRestored = calculateRestHP(session.Player)
		if hpRestored > 0 {
			currentHP := session.Player.GetHP()
			maxHP := session.Player.GetMaxHP()
			newHP := currentHP + hpRestored
			if newHP > maxHP {
				hpRestored = maxHP - currentHP
				newHP = maxHP
			}
			session.Player.SetHP(newHP)
			logger.WithFields(logrus.Fields{
				"hp_restored": hpRestored,
				"new_hp":      newHP,
			}).Info("HP restored from rest")
		}
	}

	logger.Debug("exiting handleRest")
	return map[string]interface{}{
		"success":        true,
		"slots_restored": true,
		"hp_restored":    hpRestored,
		"message":        formatRestMessage(hpRestored),
	}, nil
}

// calculateRestHP computes the HP restored from resting based on character level.
// Uses D&D-style resting rules: 1 HP per character level, minimum of 1 HP.
func calculateRestHP(player *game.Player) int {
	if player == nil {
		return 0
	}
	level := player.GetLevel()
	if level < 1 {
		level = 1
	}
	return level
}

// formatRestMessage creates a descriptive message for the rest action results.
func formatRestMessage(hpRestored int) string {
	if hpRestored > 0 {
		return fmt.Sprintf("You rest and recover your spell slots and %d hit points.", hpRestored)
	}
	return "You rest and recover your spell slots."
}

// handleGetCombatModifiers returns combat modifiers (cover and flanking) for an attack.
// This is used by the UI to display tactical information during attack targeting.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - target_id: string identifier for the potential attack target
//
// Returns:
//   - interface{}: Map containing cover_type, cover_bonus, is_flanking, flanking_bonus
//   - error: Error if parameters are invalid or session doesn't exist
func (s *RPCServer) handleGetCombatModifiers(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetCombatModifiers",
	})
	logger.Debug("entering handleGetCombatModifiers")

	var req struct {
		SessionID string `json:"session_id"`
		TargetID  string `json:"target_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).Error("failed to unmarshal parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid parameters", err.Error())
	}

	if req.SessionID == "" {
		logger.Warn("missing session_id")
		return nil, ErrInvalidSession
	}
	if req.TargetID == "" {
		logger.Warn("missing target_id")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "target_id is required", nil)
	}

	session, err := s.getSessionSafely(req.SessionID)
	if err != nil {
		logger.WithError(err).Warn("session not found")
		return nil, err
	}
	defer s.releaseSession(session)

	if session.Player == nil {
		logger.Warn("no player in session")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "no player in session", nil)
	}

	// Get target from world (WorldState handles its own locking internally)
	target, exists := s.state.WorldState.Objects[req.TargetID]
	world := s.state.WorldState

	if !exists || target == nil {
		logger.WithField("targetID", req.TargetID).Warn("target not found")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "target not found", nil)
	}

	attackerPos := session.Player.GetPosition()
	defenderPos := target.GetPosition()

	combatMods := game.NewCombatModifiers(world)
	coverType := combatMods.CalculateCover(attackerPos, defenderPos)
	coverBonus := game.CoverBonus(coverType)

	// Get ally positions for flanking calculation
	allyPositions := s.getPlayerAllyPositions(session.Player.GetID())
	isFlanking, flankingBonus := combatMods.CalculateFlanking(attackerPos, defenderPos, allyPositions)

	result := map[string]interface{}{
		"cover_type":     coverTypeToString(coverType),
		"cover_bonus":    coverBonus,
		"is_flanking":    isFlanking,
		"flanking_bonus": flankingBonus,
		"attacker_pos":   map[string]int{"x": attackerPos.X, "y": attackerPos.Y},
		"defender_pos":   map[string]int{"x": defenderPos.X, "y": defenderPos.Y},
	}

	logger.WithFields(logrus.Fields{
		"cover":    result["cover_type"],
		"flanking": isFlanking,
	}).Debug("exiting handleGetCombatModifiers")

	return result, nil
}

// coverTypeToString converts a CoverType enum to a human-readable string.
func coverTypeToString(ct game.CoverType) string {
	switch ct {
	case game.CoverHalf:
		return "half"
	case game.CoverThreeQuarters:
		return "three_quarters"
	case game.CoverFull:
		return "full"
	default:
		return "none"
	}
}

// getPlayerAllyPositions returns positions of all allies of the given player.
// For flanking purposes, considers other player characters as allies.
func (s *RPCServer) getPlayerAllyPositions(playerID string) []game.Position {
	positions := make([]game.Position, 0)

	for id, obj := range s.state.WorldState.Objects {
		if id == playerID {
			continue
		}
		// Check if object has "player" tag (indicates a player character)
		tags := obj.GetTags()
		isPlayerChar := false
		for _, tag := range tags {
			if tag == "player" {
				isPlayerChar = true
				break
			}
		}
		if isPlayerChar {
			positions = append(positions, obj.GetPosition())
		}
	}

	return positions
}

// handleGetGameState returns the current game state for a session.
// It validates the session and returns world, player, and combat information.
func (s *RPCServer) handleGetGameState(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetGameState",
	})
	logger.Debug("entering handleGetGameState")

	// 1. Validate params
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).Error("failed to unmarshal parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid parameters", err.Error())
	}

	// 2. Check session safely
	session, err := s.getSessionSafely(req.SessionID)
	if err != nil {
		logger.WithField("sessionID", req.SessionID).Warn("session not found")
		return nil, ErrInvalidSession
	}
	defer s.releaseSession(session) // Ensure session is released when handler completes

	// 3. Get game state (uses its own internal locking)
	state := s.state.GetState()
	if state == nil {
		logger.Error("failed to get game state")
		return nil, fmt.Errorf("internal server error")
	}

	// 4. Add player data with character information for the requesting session
	state["player"] = s.buildPlayerStateData(session)

	logger.Debug("exiting handleGetGameState")
	return state, nil
}

// buildPlayerStateData constructs player data including character info for game state response.
func (s *RPCServer) buildPlayerStateData(session *PlayerSession) map[string]interface{} {
	playerData := map[string]interface{}{
		"session_id": session.SessionID,
		"connected":  session.Connected,
	}

	if session.Player != nil {
		playerData["name"] = session.Player.Name

		// Player embeds Character directly, so we access its fields
		char := &session.Player.Character
		pos := char.GetPosition()
		playerData["character"] = map[string]interface{}{
			"id":           char.ID,
			"name":         char.Name,
			"class":        char.Class.String(),
			"level":        char.Level,
			"strength":     char.Strength,
			"dexterity":    char.Dexterity,
			"constitution": char.Constitution,
			"intelligence": char.Intelligence,
			"wisdom":       char.Wisdom,
			"charisma":     char.Charisma,
			"current_hp":   char.HP,
			"max_hp":       char.MaxHP,
			"armor_class":  char.ArmorClass,
			"experience":   char.Experience,
			"gold":         char.Gold,
			"position": map[string]interface{}{
				"X":     pos.X,
				"Y":     pos.Y,
				"Level": pos.Level,
			},
			"immunities":  char.GetImmunities(),
			"spell_slots": session.Player.GetSpellSlots(),
			"used_slots":  session.Player.GetUsedSlots(),
		}

		// Include dungeon theme from current level properties if available
		if theme := s.state.GetLevelTheme(pos.Level); theme != "" {
			playerData["dungeon_theme"] = theme
		}
	}

	return playerData
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
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid effect parameters", err.Error())
	}

	if req.EffectType == "" {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "effect_type is required", nil)
	}
	if req.TargetID == "" {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "target_id is required", nil)
	}
	if req.Magnitude == 0 {
		req.Magnitude = 1.0
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleApplyEffect",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}

	defer s.releaseSession(session)

	// fix: guard against nil Player to prevent nil pointer dereference
	if session.Player == nil {
		return nil, fmt.Errorf("session has no player")
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

// handleJoinGame creates a new player session for the given player name.
// When called via WebSocket, the request params are enriched with the
// WebSocket connection's session_id. If that session already exists, the
// handler attaches the new player to it instead of creating a separate
// session, so that all subsequent WebSocket requests (which carry the
// same enriched session_id) can find the player.
func (s *RPCServer) handleJoinGame(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleJoinGame").Debug("entering handleJoinGame")

	req, err := s.parseJoinGameRequest(params)
	if err != nil {
		return nil, err
	}

	player, err := s.createDefaultPlayer(req.PlayerName)
	if err != nil {
		return nil, err
	}

	// Try to attach to existing session if session_id provided
	if req.SessionID != "" {
		if result, ok := s.tryAttachToExistingSession(req.SessionID, player); ok {
			return result, nil
		}
	}

	// Create new session
	return s.createNewPlayerSession(player, req.PlayerName)
}

// joinGameRequest holds parsed join game request parameters.
type joinGameRequest struct {
	PlayerName string `json:"player_name"`
	SessionID  string `json:"session_id"`
}

// parseJoinGameRequest parses and validates join game request parameters.
func (s *RPCServer) parseJoinGameRequest(params json.RawMessage) (*joinGameRequest, error) {
	var req joinGameRequest
	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithField("function", "parseJoinGameRequest").WithError(err).Error("failed to unmarshal")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid join parameters", err.Error())
	}
	if req.PlayerName == "" {
		logrus.WithField("function", "parseJoinGameRequest").Warn("empty player name")
		return nil, fmt.Errorf("player name is required")
	}
	return &req, nil
}

// createDefaultPlayer creates a default fighter character for a new player.
func (s *RPCServer) createDefaultPlayer(playerName string) (*game.Player, error) {
	creator := game.NewCharacterCreator()
	result := creator.CreateCharacter(game.CharacterCreationConfig{
		Name:              playerName,
		Class:             game.ClassFighter,
		AttributeMethod:   "standard",
		StartingEquipment: true,
		StartingGold:      100,
	})
	if !result.Success || result.PlayerData == nil {
		logrus.WithField("function", "createDefaultPlayer").WithField("errors", result.Errors).Error("creation failed")
		return nil, fmt.Errorf("failed to create player character")
	}
	return result.PlayerData, nil
}

// tryAttachToExistingSession attempts to attach a player to an existing session.
// Returns the result and true if successful, or nil and false if session not found.
func (s *RPCServer) tryAttachToExistingSession(sessionID string, player *game.Player) (interface{}, bool) {
	s.mu.Lock()
	existing, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return nil, false
	}

	// Prevent overwriting an existing player
	if existing.Player != nil {
		s.mu.Unlock()
		logrus.WithField("function", "tryAttachToExistingSession").WithField("sessionID", sessionID).
			Warn("session already has a player")
		return nil, false
	}

	existing.Player = player
	existing.Player.InitializeSpellSlots()
	existing.Connected = true
	existing.LastActive = time.Now()
	s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function":    "tryAttachToExistingSession",
		"sessionID":   sessionID,
		"player_name": player.Name,
	}).Info("attached player to existing session")

	s.state.AddPlayer(existing)

	return map[string]interface{}{
		"success":    true,
		"session_id": existing.SessionID,
		"player_id":  player.GetID(),
	}, true
}

// createNewPlayerSession creates a new session for a player.
func (s *RPCServer) createNewPlayerSession(player *game.Player, playerName string) (interface{}, error) {
	s.mu.Lock()
	sessionID := uuid.New().String()
	player.InitializeSpellSlots()
	session := &PlayerSession{
		SessionID:   sessionID,
		Player:      player,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		Connected:   true,
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}
	s.sessions[sessionID] = session
	s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function":    "createNewPlayerSession",
		"sessionID":   sessionID,
		"player_name": playerName,
	}).Info("created new session for player")

	s.state.AddPlayer(session)

	return map[string]interface{}{
		"success":    true,
		"session_id": session.SessionID,
		"player_id":  player.GetID(),
	}, nil
}

// handleCreateCharacter processes a character creation request and creates a new character.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - name: string - Character name
//   - class: string - Character class ("fighter", "mage", "cleric", "thief", "ranger", "paladin")
//   - attribute_method: string - Attribute generation method ("roll", "pointbuy", "standard", "custom")
//   - custom_attributes: map[string]int - Custom attribute values (optional)
//   - starting_equipment: bool - Whether to include starting equipment
//   - starting_gold: int - Starting gold amount (optional)
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if creation was successful
//   - character: Created character data
//   - player: Created player data
//   - session_id: Session ID for the new character
//   - errors: List of any error messages
//   - warnings: List of any warning messages
//
// Errors:
//   - "invalid character creation parameters" if JSON unmarshaling fails
//   - Character creation validation errors from CharacterCreator
//   - Session creation errors
func (s *RPCServer) handleCreateCharacter(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleCreateCharacter").Debug("entering handleCreateCharacter")

	req, err := s.parseCharacterCreationRequest(params)
	if err != nil {
		return nil, err
	}

	config, err := s.buildCharacterConfig(req)
	if err != nil {
		return nil, err
	}

	result := s.createNewCharacter(config)
	if !result.Success {
		logrus.WithField("function", "handleCreateCharacter").WithField("errors", result.Errors).Error("creation failed")
		return map[string]interface{}{
			"success":  false,
			"errors":   result.Errors,
			"warnings": result.Warnings,
		}, nil
	}

	session := s.attachOrCreateSession(req.SessionID, result.PlayerData)

	logrus.WithFields(logrus.Fields{
		"function":      "handleCreateCharacter",
		"sessionID":     session.SessionID,
		"characterName": req.Name,
		"class":         req.Class,
	}).Info("character created successfully")

	return buildCharacterResponse(result, session), nil
}

// buildCharacterResponse constructs the response for character creation.
func buildCharacterResponse(result *game.CharacterCreationResult, session *PlayerSession) map[string]interface{} {
	char := result.Character
	characterData := buildCharacterData(char)

	return map[string]interface{}{
		// New keys
		"success":         true,
		"character":       characterData,
		"player":          result.PlayerData,
		"session_id":      session.SessionID,
		"errors":          result.Errors,
		"warnings":        result.Warnings,
		"creation_time":   result.CreationTime,
		"generated_stats": result.GeneratedStats,
		"starting_items":  result.StartingItems,
		// Legacy aliases
		"Success":        true,
		"Character":      characterData,
		"Player":         result.PlayerData,
		"SessionID":      session.SessionID,
		"Errors":         result.Errors,
		"Warnings":       result.Warnings,
		"CreationTime":   result.CreationTime,
		"GeneratedStats": result.GeneratedStats,
		"StartingItems":  result.StartingItems,
	}
}

// buildCharacterData constructs the character data map with both new and legacy keys.
func buildCharacterData(char *game.Character) map[string]interface{} {
	pos := char.GetPosition()
	attrs := map[string]interface{}{
		"strength":     char.Strength,
		"dexterity":    char.Dexterity,
		"constitution": char.Constitution,
		"intelligence": char.Intelligence,
		"wisdom":       char.Wisdom,
		"charisma":     char.Charisma,
	}
	attrsLegacy := map[string]interface{}{
		"Strength":     char.Strength,
		"Dexterity":    char.Dexterity,
		"Constitution": char.Constitution,
		"Intelligence": char.Intelligence,
		"Wisdom":       char.Wisdom,
		"Charisma":     char.Charisma,
	}
	posMap := map[string]interface{}{"X": pos.X, "Y": pos.Y, "Level": pos.Level}

	data := map[string]interface{}{
		// New lower_snake_case keys
		"id": char.ID, "name": char.Name, "class": char.Class.String(),
		"level": char.Level, "hp": char.HP, "max_hp": char.MaxHP,
		"ap": char.ActionPoints, "max_ap": char.MaxActionPoints,
		"experience": char.Experience, "position": posMap, "attributes": attrs,
		// Legacy struct-style/camel-case aliases
		"ID": char.ID, "Name": char.Name, "Class": char.Class.String(),
		"Level": char.Level, "HP": char.HP, "MaxHP": char.MaxHP,
		"AP": char.ActionPoints, "MaxAP": char.MaxActionPoints,
		"Experience": char.Experience, "Position": posMap, "Attributes": attrsLegacy,
	}
	if char.Appearance != (game.Appearance{}) {
		data["appearance"] = char.Appearance
		data["Appearance"] = char.Appearance
	}
	return data
}

// createCharacterRequest defines the structure for a character creation request.
type createCharacterRequest struct {
	SessionID         string         `json:"session_id"`
	Name              string         `json:"name"`
	Class             string         `json:"class"`
	AttributeMethod   string         `json:"attribute_method"`
	CustomAttributes  map[string]int `json:"custom_attributes,omitempty"`
	StartingEquipment bool           `json:"starting_equipment"`
	StartingGold      int            `json:"starting_gold"`
	// Cosmetic / identity fields — all optional
	SkinTone            int    `json:"skin_tone,omitempty"`
	HairStyle           string `json:"hair_style,omitempty"`
	HairColor           string `json:"hair_color,omitempty"`
	BodyType            int    `json:"body_type,omitempty"`
	GenderExpression    string `json:"gender_expression,omitempty"`
	Pronouns            string `json:"pronouns,omitempty"`
	RomanticOrientation string `json:"romantic_orientation,omitempty"`
}

// parseCharacterCreationRequest unmarshals the raw JSON into a createCharacterRequest struct.
func (s *RPCServer) parseCharacterCreationRequest(params json.RawMessage) (*createCharacterRequest, error) {
	var req createCharacterRequest
	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseCharacterCreationRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal character creation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid character creation parameters", err.Error())
	}
	return &req, nil
}

// buildCharacterConfig creates the character configuration from the request.
func (s *RPCServer) buildCharacterConfig(req *createCharacterRequest) (*game.CharacterCreationConfig, error) {
	classMap := map[string]game.CharacterClass{
		"fighter": game.ClassFighter,
		"mage":    game.ClassMage,
		"cleric":  game.ClassCleric,
		"thief":   game.ClassThief,
		"ranger":  game.ClassRanger,
		"paladin": game.ClassPaladin,
	}

	if req.Class == "" {
		req.Class = "fighter"
	}

	characterClass, exists := classMap[strings.ToLower(req.Class)]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function": "buildCharacterConfig",
			"class":    req.Class,
		}).Error("invalid character class")
		return nil, fmt.Errorf("invalid character class: %s", req.Class)
	}

	if req.AttributeMethod == "" {
		req.AttributeMethod = "standard"
	}

	if req.Name == "" {
		req.Name = "Adventurer"
	}

	if req.StartingGold == 0 {
		defaultGold := map[game.CharacterClass]int{
			game.ClassFighter: 100,
			game.ClassMage:    50,
			game.ClassCleric:  75,
			game.ClassThief:   80,
			game.ClassRanger:  90,
			game.ClassPaladin: 120,
		}
		req.StartingGold = defaultGold[characterClass]
	}

	return &game.CharacterCreationConfig{
		Name:              req.Name,
		Class:             characterClass,
		AttributeMethod:   strings.ToLower(req.AttributeMethod),
		CustomAttributes:  req.CustomAttributes,
		StartingEquipment: req.StartingEquipment || req.StartingGold > 0, // Default to true for new characters
		StartingGold:      req.StartingGold,
		Appearance: game.Appearance{
			SkinTone:            req.SkinTone,
			HairStyle:           req.HairStyle,
			HairColor:           req.HairColor,
			BodyType:            game.BodyType(req.BodyType),
			GenderExpression:    req.GenderExpression,
			Pronouns:            req.Pronouns,
			RomanticOrientation: req.RomanticOrientation,
		},
	}, nil
}

// createNewCharacter uses the CharacterCreator to create a new character based on the config.
func (s *RPCServer) createNewCharacter(config *game.CharacterCreationConfig) *game.CharacterCreationResult {
	creator := game.NewCharacterCreator()
	result := creator.CreateCharacter(*config)
	return &result
}

// createAndRegisterSession creates a new player session and registers it with the server.
func (s *RPCServer) createAndRegisterSession(playerData *game.Player) *PlayerSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	var sessionID string
	for {
		sessionID = game.NewUID()
		if _, exists := s.sessions[sessionID]; !exists {
			break
		}
		logrus.WithFields(logrus.Fields{
			"function":  "createAndRegisterSession",
			"sessionID": sessionID,
		}).Warn("session ID collision detected, generating new ID")
	}

	session := &PlayerSession{
		SessionID:   sessionID,
		Player:      playerData,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   false,
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}

	s.sessions[sessionID] = session

	// Add player to world state so they appear in WorldState.Objects
	if s.state != nil {
		s.state.AddPlayer(session)
	}

	return session
}

// attachOrCreateSession attaches playerData to an existing session identified
// by sessionID when one exists, or creates a new session otherwise.
// This ensures that WebSocket-enriched requests reuse the connection session
// instead of orphaning the player in an unreachable session.
func (s *RPCServer) attachOrCreateSession(sessionID string, playerData *game.Player) *PlayerSession {
	if sessionID != "" {
		s.mu.Lock()
		if existing, ok := s.sessions[sessionID]; ok {
			// If the session already has a different player attached, do not overwrite it.
			if existing.Player != nil && existing.Player != playerData {
				logrus.WithFields(logrus.Fields{
					"function":  "attachOrCreateSession",
					"sessionID": sessionID,
				}).Warn("attempt to attach a different player to an existing session; keeping original player")
				s.mu.Unlock()
				return existing
			}

			// Determine if this is the first time a player is being attached to this session.
			needAdd := existing.Player == nil

			existing.Player = playerData
			existing.Connected = true
			existing.LastActive = time.Now()
			s.mu.Unlock()

			// Only add the player to world state the first time it is attached to this session.
			if needAdd && s.state != nil {
				s.state.AddPlayer(existing)
			}

			logrus.WithFields(logrus.Fields{
				"function":  "attachOrCreateSession",
				"sessionID": sessionID,
			}).Info("attached player to existing session")
			return existing
		}
		s.mu.Unlock()
	}
	return s.createAndRegisterSession(playerData)
}

// Equipment management handlers
// getPlayerSession retrieves a player session by session ID with validation
func (s *RPCServer) getPlayerSession(sessionID string) (*PlayerSession, error) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	if session.Player == nil {
		return nil, fmt.Errorf("session has no associated player")
	}

	return session, nil
}

// useItemRequest defines the structure for a use item request.
type useItemRequest struct {
	SessionID string `json:"session_id"`
	ItemID    string `json:"item_id"`
	TargetID  string `json:"target_id"`
}

// parseAndValidateUseItemRequest parses and validates the use item request.
func (s *RPCServer) parseAndValidateUseItemRequest(params json.RawMessage) (*useItemRequest, error) {
	var req useItemRequest
	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseAndValidateUseItemRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal use item parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid use item parameters", err.Error())
	}

	if req.SessionID == "" {
		logrus.WithFields(logrus.Fields{
			"function": "parseAndValidateUseItemRequest",
		}).Warn("empty session ID")
		return nil, ErrInvalidSession
	}

	if req.ItemID == "" {
		logrus.WithFields(logrus.Fields{
			"function": "parseAndValidateUseItemRequest",
		}).Warn("empty item ID")
		return nil, fmt.Errorf("item ID is required")
	}

	return &req, nil
}

// validateCombatTurnForItemUse checks if the player can use an item during combat.
func (s *RPCServer) validateCombatTurnForItemUse(player *game.Player) error {
	if s.state.TurnManager.IsInCombat {
		if !s.state.TurnManager.IsCurrentTurn(player.GetID()) {
			logrus.WithFields(logrus.Fields{
				"function": "validateCombatTurnForItemUse",
				"playerID": player.GetID(),
			}).Warn("player attempted to use item when not their turn")
			return fmt.Errorf("not your turn")
		}
	}
	return nil
}

// executeItemUsage contains the core logic for using an item.
func (s *RPCServer) executeItemUsage(player *game.Player, itemID, targetID string) (string, error) {
	item := findInventoryItem(player.Character.Inventory, itemID)
	if item == nil {
		logrus.WithFields(logrus.Fields{
			"function": "executeItemUsage",
			"itemID":   itemID,
		}).Error("failed to find item in inventory")
		return "", fmt.Errorf("item %s not found in inventory", itemID)
	}

	effect := fmt.Sprintf("Used %s", item.Name)
	if targetID != "" {
		effect = fmt.Sprintf("Used %s on %s", item.Name, targetID)
	}

	if item.Type == "consumable" {
		logrus.WithFields(logrus.Fields{
			"function": "executeItemUsage",
			"itemID":   itemID,
		}).Info("removing consumable item from inventory")
		// This is a simplified implementation. In a full implementation,
		// you would handle item quantities and removal properly.
	}

	return effect, nil
}

// handleUseItem processes a request to use an item from the player's inventory.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - item_id: string identifier for the item to use
//   - target_id: string identifier for the target (player, NPC, etc.)
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if item use was successful
//   - effect: string describing the effect of using the item
//   - error: Possible errors:
//   - "invalid use item parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
//   - Item-specific validation errors
//
// Related:
//   - game.Item
//   - game.Inventory
//   - PlayerSession
func (s *RPCServer) handleUseItem(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleUseItem",
	}).Debug("entering handleUseItem")

	req, err := s.parseAndValidateUseItemRequest(params)
	if err != nil {
		s.recordActionMetrics("use_item", err)
		return nil, err
	}

	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		s.recordActionMetrics("use_item", err)
		return nil, err
	}

	if err := s.validateCombatTurnForItemUse(session.Player); err != nil {
		s.recordActionMetrics("use_item", err)
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function":  "handleUseItem",
		"sessionID": req.SessionID,
		"itemID":    req.ItemID,
		"targetID":  req.TargetID,
	}).Info("using item from inventory")

	result, err := s.executeItemUsage(session.Player, req.ItemID, req.TargetID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUseItem",
			"error":    err,
		}).Error("failed to use item")
		s.recordActionMetrics("use_item", err)
		return nil, err
	}

	s.recordActionMetrics("use_item", nil)
	logrus.WithFields(logrus.Fields{
		"function": "handleUseItem",
		"effect":   result,
	}).Info("item used successfully")
	return map[string]interface{}{"success": true, "effect": result}, nil
}

// findInventoryItem searches for an item in the player's inventory by its ID.
func (s *RPCServer) findInventoryItem(player *game.Player, itemID string) (*game.Item, bool) {
	for _, item := range player.Character.Inventory {
		if item.ID == itemID {
			return &item, true
		}
	}
	return nil, false
}

// parseLeaveGameRequest validates and unmarshals leave game request parameters.
func (s *RPCServer) parseLeaveGameRequest(params json.RawMessage) (string, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseLeaveGameRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal leave game parameters")
		return "", fmt.Errorf("invalid leave game parameters")
	}

	if req.SessionID == "" {
		logrus.WithFields(logrus.Fields{
			"function": "parseLeaveGameRequest",
		}).Warn("empty session ID")
		return "", ErrInvalidSession
	}

	return req.SessionID, nil
}

// cleanupSessionConnections handles cleanup of WebSocket connections and channels for a session.
func (s *RPCServer) cleanupSessionConnections(session *PlayerSession, sessionID string) {
	// Close WebSocket connection if it exists
	if session.WSConn != nil {
		if err := session.WSConn.CloseNow(); err != nil {
			logrus.WithFields(logrus.Fields{
				"function":  "cleanupSessionConnections",
				"sessionID": sessionID,
				"error":     err.Error(),
			}).Warn("failed to close WebSocket connection")
		}
	}

	// Close message channel safely (prevents double-close panic)
	session.closeMessageChannel()
}

// removePlayerFromGameState removes player from world state objects.
func (s *RPCServer) removePlayerFromGameState(session *PlayerSession) {
	if session.Player != nil {
		// Remove player from world state objects
		if s.state.WorldState != nil && s.state.WorldState.Objects != nil {
			delete(s.state.WorldState.Objects, session.Player.GetID())
		}
	}
}

// executeSessionCleanup performs the complete cleanup process for a player session.
func (s *RPCServer) executeSessionCleanup(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "executeSessionCleanup",
			"sessionID": sessionID,
		}).Warn("session not found")
		return ErrInvalidSession
	}

	// Cleanup connections and channels
	s.cleanupSessionConnections(session, sessionID)

	// Remove player from game state
	s.removePlayerFromGameState(session)

	// Remove session from sessions map
	delete(s.sessions, sessionID)

	logrus.WithFields(logrus.Fields{
		"function":  "executeSessionCleanup",
		"sessionID": sessionID,
	}).Info("player left game and session removed")

	return nil
}

// handleLeaveGame processes a request to leave the game and end the session.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session to end
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if leave operation was successful
//   - error: Possible errors:
//   - "invalid leave game parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
//
// Related:
//   - PlayerSession
//   - RPCServer.sessions
func (s *RPCServer) handleLeaveGame(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleLeaveGame",
	}).Debug("entering handleLeaveGame")

	sessionID, err := s.parseLeaveGameRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.executeSessionCleanup(sessionID); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}
