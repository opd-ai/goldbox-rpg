package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

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
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid movement parameters", err.Error())
	}

	session, err := s.getSessionSafely(req.SessionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function":  "handleMove",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}
	defer s.releaseSession(session) // Ensure session is released when handler completes

	// Check if currently in combat - if so, validate turn order and action points
	if s.state.TurnManager.IsInCombat {
		if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
			logrus.WithFields(logrus.Fields{
				"function": "handleMove",
				"playerID": session.Player.GetID(),
			}).Warn("player attempted to move when not their turn")
			return nil, fmt.Errorf("not your turn")
		}

		// Check if player has enough action points for movement
		if session.Player.GetActionPoints() < game.ActionCostMove {
			logrus.WithFields(logrus.Fields{
				"function":   "handleMove",
				"playerID":   session.Player.GetID(),
				"currentAP":  session.Player.GetActionPoints(),
				"requiredAP": game.ActionCostMove,
			}).Warn("player attempted to move without enough action points")
			return nil, fmt.Errorf("insufficient action points for movement (need %d, have %d)",
				game.ActionCostMove, session.Player.GetActionPoints())
		}
	}

	player := session.Player
	currentPos := player.GetPosition()
	newPos := calculateNewPosition(currentPos, req.Direction, s.state.WorldState.Width, s.state.WorldState.Height)

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

	// Consume action points BEFORE making any state changes if in combat
	if s.state.TurnManager.IsInCombat {
		if !player.ConsumeActionPoints(game.ActionCostMove) {
			logrus.WithFields(logrus.Fields{
				"function": "handleMove",
				"playerID": player.GetID(),
			}).Error("failed to consume action points before movement")
			return nil, fmt.Errorf("action point consumption failed")
		}
		logrus.WithFields(logrus.Fields{
			"function":    "handleMove",
			"playerID":    player.GetID(),
			"consumedAP":  game.ActionCostMove,
			"remainingAP": player.GetActionPoints(),
		}).Info("consumed action points for movement")
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
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid attack parameters", err.Error())
	}

	session, err := s.getSessionSafely(req.SessionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function":  "handleAttack",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}
	defer s.releaseSession(session) // Ensure session is released when handler completes

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

	// Check if player has enough action points for attack
	if session.Player.GetActionPoints() < game.ActionCostAttack {
		logrus.WithFields(logrus.Fields{
			"function":   "handleAttack",
			"playerID":   session.Player.GetID(),
			"currentAP":  session.Player.GetActionPoints(),
			"requiredAP": game.ActionCostAttack,
		}).Warn("player attempted to attack without enough action points")
		return nil, fmt.Errorf("insufficient action points for attack (need %d, have %d)",
			game.ActionCostAttack, session.Player.GetActionPoints())
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

	// Consume action points after successful attack
	if !session.Player.ConsumeActionPoints(game.ActionCostAttack) {
		// This should not happen due to earlier validation, but safety check
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
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid spell parameters", err.Error())
	}

	session, err := s.getSessionSafely(req.SessionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function":  "handleCastSpell",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}
	defer s.releaseSession(session) // Ensure session is released when handler completes

	// Check if currently in combat (spells can also be cast outside combat)
	if s.state.TurnManager.IsInCombat {
		// If in combat, validate turn order
		if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
			logrus.WithFields(logrus.Fields{
				"function": "handleCastSpell",
				"playerID": session.Player.GetID(),
			}).Warn("player attempted to cast spell when not their turn")
			return nil, fmt.Errorf("not your turn")
		}

		// Check if player has enough action points for spell casting
		if session.Player.GetActionPoints() < game.ActionCostSpell {
			logrus.WithFields(logrus.Fields{
				"function":   "handleCastSpell",
				"playerID":   session.Player.GetID(),
				"currentAP":  session.Player.GetActionPoints(),
				"requiredAP": game.ActionCostSpell,
			}).Warn("player attempted to cast spell without enough action points")
			return nil, fmt.Errorf("insufficient action points for spell casting (need %d, have %d)",
				game.ActionCostSpell, session.Player.GetActionPoints())
		}
	}

	player := session.Player
	spell, err := s.spellManager.GetSpell(req.SpellID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleCastSpell",
			"spellID":  req.SpellID,
			"playerID": player.GetID(),
		}).Warn("spell not found in spell database")
		return nil, fmt.Errorf("spell not found: %s", req.SpellID)
	}

	// Check if player knows this spell
	if !player.KnowsSpell(req.SpellID) {
		logrus.WithFields(logrus.Fields{
			"function": "handleCastSpell",
			"playerID": player.GetID(),
			"spellID":  req.SpellID,
		}).Warn("player does not know this spell")
		return nil, fmt.Errorf("you do not know this spell: %s", spell.Name)
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

	// Consume action points if in combat
	if s.state.TurnManager.IsInCombat {
		if !player.ConsumeActionPoints(game.ActionCostSpell) {
			// This should not happen due to earlier validation, but safety check
			logrus.WithFields(logrus.Fields{
				"function": "handleCastSpell",
				"playerID": player.GetID(),
			}).Error("failed to consume action points after spell validation")
			return nil, fmt.Errorf("action point consumption failed")
		}
		logrus.WithFields(logrus.Fields{
			"function":    "handleCastSpell",
			"playerID":    player.GetID(),
			"consumedAP":  game.ActionCostSpell,
			"remainingAP": player.GetActionPoints(),
		}).Info("consumed action points for spell casting")
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
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid combat parameters", err.Error())
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
	if err := s.state.TurnManager.StartCombat(initiative); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleStartCombat",
			"error":    err.Error(),
		}).Error("failed to start combat")
		return nil, fmt.Errorf("failed to start combat: %w", err)
	}

	// Initialize action points for all combat participants
	s.mu.RLock()
	for _, participantID := range initiative {
		for _, session := range s.sessions {
			if session.Player.GetID() == participantID {
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
	s.mu.RUnlock()

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
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid turn parameters", err.Error())
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleEndTurn",
			"sessionID": req.SessionID,
		}).Warn("invalid session ID")
		return nil, fmt.Errorf("invalid session")
	}
	defer s.releaseSession(session) // Ensure session is released when handler completes

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

	// Restore action points for the next player
	if nextTurn != "" {
		s.mu.RLock()
		for _, nextSession := range s.sessions {
			if nextSession.Player.GetID() == nextTurn {
				nextSession.Player.RestoreActionPoints()
				logrus.WithFields(logrus.Fields{
					"function":     "handleEndTurn",
					"nextPlayerID": nextTurn,
					"restoredAP":   nextSession.Player.GetActionPoints(),
				}).Info("restored action points for next player")
				break
			}
		}
		s.mu.RUnlock()
	}

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

	logger.Debug("exiting handleGetGameState")
	return state, nil
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

	session, exists := s.getSession(req.SessionID)
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

func (s *RPCServer) handleJoinGame(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleJoinGame",
	}).Debug("entering handleJoinGame")

	var req struct {
		PlayerName string `json:"player_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleJoinGame",
			"error":    err.Error(),
		}).Error("failed to unmarshal join parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid join parameters", err.Error())
	}

	if req.PlayerName == "" {
		logrus.WithFields(logrus.Fields{
			"function": "handleJoinGame",
		}).Warn("empty player name")
		return nil, fmt.Errorf("player name is required")
	}

	// Create new session
	s.mu.Lock()
	sessionID := uuid.New().String()
	session := &PlayerSession{
		SessionID:   sessionID,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}
	s.sessions[sessionID] = session
	s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function":    "handleJoinGame",
		"sessionID":   sessionID,
		"player_name": req.PlayerName,
	}).Info("created new session for player")

	// Initialize player in session
	s.state.AddPlayer(session)

	logrus.WithFields(logrus.Fields{
		"function": "handleJoinGame",
	}).Debug("exiting handleJoinGame")

	return map[string]interface{}{
		"success":    true,
		"session_id": session.SessionID,
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
	logrus.WithFields(logrus.Fields{
		"function": "handleCreateCharacter",
	}).Debug("entering handleCreateCharacter")

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
		logrus.WithFields(logrus.Fields{
			"function": "handleCreateCharacter",
			"errors":   result.Errors,
		}).Error("character creation failed")
		return map[string]interface{}{
			"success":  false,
			"errors":   result.Errors,
			"warnings": result.Warnings,
		}, nil
	}

	session := s.createAndRegisterSession(result.PlayerData)

	logrus.WithFields(logrus.Fields{
		"function":      "handleCreateCharacter",
		"sessionID":     session.SessionID,
		"characterName": req.Name,
		"class":         req.Class,
	}).Info("character created successfully")

	return map[string]interface{}{
		"success":         true,
		"character":       result.Character,
		"player":          result.PlayerData,
		"session_id":      session.SessionID,
		"errors":          result.Errors,
		"warnings":        result.Warnings,
		"creation_time":   result.CreationTime,
		"generated_stats": result.GeneratedStats,
		"starting_items":  result.StartingItems,
	}, nil
}

// createCharacterRequest defines the structure for a character creation request.
type createCharacterRequest struct {
	Name              string         `json:"name"`
	Class             string         `json:"class"`
	AttributeMethod   string         `json:"attribute_method"`
	CustomAttributes  map[string]int `json:"custom_attributes,omitempty"`
	StartingEquipment bool           `json:"starting_equipment"`
	StartingGold      int            `json:"starting_gold"`
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

	characterClass, exists := classMap[req.Class]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function": "buildCharacterConfig",
			"class":    req.Class,
		}).Error("invalid character class")
		return nil, fmt.Errorf("invalid character class: %s", req.Class)
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
		AttributeMethod:   req.AttributeMethod,
		CustomAttributes:  req.CustomAttributes,
		StartingEquipment: req.StartingEquipment,
		StartingGold:      req.StartingGold,
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
	return session
}

// Equipment management handlers
func (s *RPCServer) handleEquipItem(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleEquipItem",
	}).Debug("entering handleEquipItem")

	var req struct {
		SessionID string `json:"session_id"`
		ItemID    string `json:"item_id"`
		Slot      string `json:"slot"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleEquipItem",
			"error":    err.Error(),
		}).Error("failed to unmarshal equip item parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid equip item parameters", err.Error())
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	player := session.Player

	// Parse slot name to EquipmentSlot
	slot, err := parseEquipmentSlot(req.Slot)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleEquipItem",
			"slot":     req.Slot,
		}).Error("invalid equipment slot")
		return nil, fmt.Errorf("invalid equipment slot: %s", req.Slot)
	}

	// Check if there's a previously equipped item
	var previousItem *game.Item
	if prevEquipped, exists := player.GetEquippedItem(slot); exists {
		previousItem = prevEquipped
	}

	// Equip the item
	if err := player.EquipItem(req.ItemID, slot); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleEquipItem",
			"itemID":   req.ItemID,
			"slot":     req.Slot,
			"error":    err.Error(),
		}).Error("failed to equip item")
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, nil
	}

	// Get the newly equipped item
	equippedItem, _ := player.GetEquippedItem(slot)

	logrus.WithFields(logrus.Fields{
		"function":     "handleEquipItem",
		"sessionID":    req.SessionID,
		"itemID":       req.ItemID,
		"slot":         req.Slot,
		"equippedItem": equippedItem.Name,
	}).Info("item equipped successfully")

	response := map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Successfully equipped %s", equippedItem.Name),
		"equipped_item": equippedItem,
	}

	if previousItem != nil {
		response["previous_item"] = previousItem
	}

	return response, nil
}

// handleUnequipItem removes an equipped item and returns it to the player's inventory.
//
// Parameters (JSON):
//   - session_id: string - Player session identifier
//   - slot: string - Name of the equipment slot to unequip
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if unequipping was successful
//   - message: string describing the result
//   - unequipped_item: object containing details of the unequipped item
//
// Errors:
//   - "invalid session" if session is not found or inactive
//   - "invalid slot" if slot name is not recognized
//   - "no item equipped" if the specified slot is empty
func (s *RPCServer) handleUnequipItem(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleUnequipItem",
	}).Debug("entering handleUnequipItem")

	var req struct {
		SessionID string `json:"session_id"`
		Slot      string `json:"slot"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUnequipItem",
			"error":    err.Error(),
		}).Error("failed to unmarshal unequip item parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid unequip item parameters", err.Error())
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	player := session.Player

	// Parse slot name to EquipmentSlot
	slot, err := parseEquipmentSlot(req.Slot)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUnequipItem",
			"slot":     req.Slot,
		}).Error("invalid equipment slot")
		return nil, fmt.Errorf("invalid equipment slot: %s", req.Slot)
	}

	// Unequip the item
	unequippedItem, err := player.UnequipItem(slot)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUnequipItem",
			"slot":     req.Slot,
			"error":    err.Error(),
		}).Error("failed to unequip item")
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, nil
	}

	logrus.WithFields(logrus.Fields{
		"function":       "handleUnequipItem",
		"sessionID":      req.SessionID,
		"slot":           req.Slot,
		"unequippedItem": unequippedItem.Name,
	}).Info("item unequipped successfully")

	return map[string]interface{}{
		"success":         true,
		"message":         fmt.Sprintf("Successfully unequipped %s", unequippedItem.Name),
		"unequipped_item": unequippedItem,
	}, nil
}

// handleGetEquipment returns all currently equipped items for a player.
//
// Parameters (JSON):
//   - session_id: string - Player session identifier
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - equipment: map of slot names to equipped item objects
//   - total_weight: int total weight of all equipped items
//   - equipment_bonuses: map of stat bonuses from equipment
//
// Errors:
//   - "invalid session" if session is not found or inactive
func (s *RPCServer) handleGetEquipment(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetEquipment",
	}).Debug("entering handleGetEquipment")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetEquipment",
			"error":    err.Error(),
		}).Error("failed to unmarshal get equipment parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid get equipment parameters", err.Error())
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	player := session.Player

	// Get all equipped items
	equippedItems := player.GetAllEquippedItems()

	// Convert equipment slots to string keys for JSON response
	equipment := make(map[string]game.Item)
	totalWeight := 0
	for slot, item := range equippedItems {
		slotName := equipmentSlotToString(slot)
		equipment[slotName] = item
		totalWeight += item.Weight
	}

	// Get equipment bonuses
	bonuses := player.CalculateEquipmentBonuses()

	logrus.WithFields(logrus.Fields{
		"function":    "handleGetEquipment",
		"sessionID":   req.SessionID,
		"numItems":    len(equipment),
		"totalWeight": totalWeight,
	}).Info("equipment retrieved successfully")

	return map[string]interface{}{
		"success":           true,
		"equipment":         equipment,
		"total_weight":      totalWeight,
		"equipment_bonuses": bonuses,
	}, nil
}

// parseEquipmentSlot converts a string slot name to an EquipmentSlot enum value
func parseEquipmentSlot(slotName string) (game.EquipmentSlot, error) {
	slotMap := map[string]game.EquipmentSlot{
		"head":        game.SlotHead,
		"neck":        game.SlotNeck,
		"chest":       game.SlotChest,
		"hands":       game.SlotHands,
		"rings":       game.SlotRings,
		"legs":        game.SlotLegs,
		"feet":        game.SlotFeet,
		"weapon_main": game.SlotWeaponMain,
		"weapon_off":  game.SlotWeaponOff,
		"main_hand":   game.SlotWeaponMain, // Alternative naming
		"off_hand":    game.SlotWeaponOff,  // Alternative naming
	}

	if slot, exists := slotMap[slotName]; exists {
		return slot, nil
	}

	return game.SlotHead, fmt.Errorf("unknown equipment slot: %s", slotName)
}

// equipmentSlotToString converts an EquipmentSlot enum value to a string
func equipmentSlotToString(slot game.EquipmentSlot) string {
	slotNames := map[game.EquipmentSlot]string{
		game.SlotHead:       "head",
		game.SlotNeck:       "neck",
		game.SlotChest:      "chest",
		game.SlotHands:      "hands",
		game.SlotRings:      "rings",
		game.SlotLegs:       "legs",
		game.SlotFeet:       "feet",
		game.SlotWeaponMain: "weapon_main",
		game.SlotWeaponOff:  "weapon_off",
	}

	if name, exists := slotNames[slot]; exists {
		return name
	}

	return "unknown"
}

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

// handleStartQuest processes a request to start a new quest for a player.
// This handler validates the quest data and adds it to the player's quest log.
//
// Parameters:
//   - params: json.RawMessage containing the start quest request with:
//   - session_id: string - The session ID of the requesting player
//   - quest: Quest object - The quest data to start
//
// Returns:
//   - interface{}: Success response with quest ID if quest started successfully
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest validation failures
//   - Quest already exists in player's quest log
func (s *RPCServer) handleStartQuest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleStartQuest",
	})
	logger.Debug("entering handleStartQuest")

	var req struct {
		SessionID string     `json:"session_id"`
		Quest     game.Quest `json:"quest"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleStartQuest",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleStartQuest",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Start quest for player
	if err := session.Player.StartQuest(req.Quest); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleStartQuest",
			"quest_id": req.Quest.ID,
		}).Error("failed to start quest")
		return nil, fmt.Errorf("failed to start quest: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"function": "handleStartQuest",
		"quest_id": req.Quest.ID,
	}).Debug("exiting handleStartQuest")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.Quest.ID,
		"message":  "Quest started successfully",
	}, nil
}

// handleCompleteQuest processes a request to complete a quest for a player.
// This handler validates quest completion criteria and processes rewards.
//
// Parameters:
//   - params: json.RawMessage containing the complete quest request with:
//   - session_id: string - The session ID of the requesting player
//   - quest_id: string - The ID of the quest to complete
//
// Returns:
//   - interface{}: Success response with rewards if quest completed successfully
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest not found or not completable
//   - Quest objectives not fulfilled
func (s *RPCServer) handleCompleteQuest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleCompleteQuest",
	})
	logger.Debug("entering handleCompleteQuest")

	var req struct {
		SessionID string `json:"session_id"`
		QuestID   string `json:"quest_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleCompleteQuest",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleCompleteQuest",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Complete quest for player
	rewards, err := session.Player.CompleteQuest(req.QuestID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleCompleteQuest",
			"quest_id": req.QuestID,
		}).Error("failed to complete quest")
		return nil, fmt.Errorf("failed to complete quest: %w", err)
	}

	// Process rewards and apply them to player
	for _, reward := range rewards {
		switch reward.Type {
		case "exp":
			if err := session.Player.AddExperience(int64(reward.Value)); err != nil {
				logger.WithError(err).WithFields(logrus.Fields{
					"function":    "handleCompleteQuest",
					"quest_id":    req.QuestID,
					"reward_type": "exp",
					"value":       reward.Value,
				}).Error("failed to apply experience reward")
				return nil, fmt.Errorf("failed to apply experience reward: %w", err)
			}
			logger.WithFields(logrus.Fields{
				"function":  "handleCompleteQuest",
				"quest_id":  req.QuestID,
				"exp_added": reward.Value,
			}).Info("applied experience reward")

		case "gold":
			// Update the character's gold amount safely
			// Note: Character.Gold is a simple int field that can be safely updated
			// by modifying the Player.Character struct through the session
			previousGold := session.Player.Character.Gold
			session.Player.Character.Gold += reward.Value
			logger.WithFields(logrus.Fields{
				"function":      "handleCompleteQuest",
				"quest_id":      req.QuestID,
				"gold_added":    reward.Value,
				"previous_gold": previousGold,
				"new_gold":      session.Player.Character.Gold,
			}).Info("applied gold reward")

		case "item":
			if reward.ItemID != "" {
				// Create an item to add to inventory
				item := game.Item{
					ID:   reward.ItemID,
					Name: reward.ItemID, // Basic implementation - could be enhanced with item lookup
					Type: "quest_reward",
				}
				if err := session.Player.Character.AddItemToInventory(item); err != nil {
					logger.WithError(err).WithFields(logrus.Fields{
						"function":    "handleCompleteQuest",
						"quest_id":    req.QuestID,
						"reward_type": "item",
						"item_id":     reward.ItemID,
					}).Error("failed to apply item reward")
					return nil, fmt.Errorf("failed to apply item reward: %w", err)
				}
				logger.WithFields(logrus.Fields{
					"function": "handleCompleteQuest",
					"quest_id": req.QuestID,
					"item_id":  reward.ItemID,
				}).Info("applied item reward")
			}

		default:
			logger.WithFields(logrus.Fields{
				"function":    "handleCompleteQuest",
				"quest_id":    req.QuestID,
				"reward_type": reward.Type,
			}).Warn("unknown reward type, skipping")
		}
	}

	logger.WithFields(logrus.Fields{
		"function":     "handleCompleteQuest",
		"quest_id":     req.QuestID,
		"reward_count": len(rewards),
	}).Info("quest completed and all rewards applied")

	logger.WithFields(logrus.Fields{
		"function": "handleCompleteQuest",
		"quest_id": req.QuestID,
	}).Debug("exiting handleCompleteQuest")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.QuestID,
		"rewards":  rewards,
		"message":  "Quest completed successfully",
	}, nil
}

// handleUpdateObjective processes a request to update quest objective progress.
// This handler validates the objective update and tracks completion.
//
// Parameters:
//   - params: json.RawMessage containing the update objective request with:
//   - session_id: string - The session ID of the requesting player
//   - quest_id: string - The ID of the quest containing the objective
//   - objective_index: int - The index of the objective to update (0-based)
//   - progress: int - The new progress value for the objective
//
// Returns:
//   - interface{}: Success response with updated objective status
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest not found or not active
//   - Invalid objective index
func (s *RPCServer) handleUpdateObjective(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleUpdateObjective",
	})
	logger.Debug("entering handleUpdateObjective")

	var req struct {
		SessionID      string `json:"session_id"`
		QuestID        string `json:"quest_id"`
		ObjectiveIndex int    `json:"objective_index"`
		Progress       int    `json:"progress"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleUpdateObjective",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleUpdateObjective",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Update quest objective for player
	if err := session.Player.UpdateQuestObjective(req.QuestID, req.ObjectiveIndex, req.Progress); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":        "handleUpdateObjective",
			"quest_id":        req.QuestID,
			"objective_index": req.ObjectiveIndex,
		}).Error("failed to update quest objective")
		return nil, fmt.Errorf("failed to update quest objective: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"function":        "handleUpdateObjective",
		"quest_id":        req.QuestID,
		"objective_index": req.ObjectiveIndex,
		"progress":        req.Progress,
	}).Debug("exiting handleUpdateObjective")

	return map[string]interface{}{
		"success":         true,
		"quest_id":        req.QuestID,
		"objective_index": req.ObjectiveIndex,
		"progress":        req.Progress,
		"message":         "Quest objective updated successfully",
	}, nil
}

// handleFailQuest processes a request to fail a quest for a player.
// This handler marks the quest as failed, preventing completion.
//
// Parameters:
//   - params: json.RawMessage containing the fail quest request with:
//   - session_id: string - The session ID of the requesting player
//   - quest_id: string - The ID of the quest to fail
//
// Returns:
//   - interface{}: Success response confirming quest failure
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest not found or already completed/failed
func (s *RPCServer) handleFailQuest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleFailQuest",
	})
	logger.Debug("entering handleFailQuest")

	var req struct {
		SessionID string `json:"session_id"`
		QuestID   string `json:"quest_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleFailQuest",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleFailQuest",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Fail quest for player
	if err := session.Player.FailQuest(req.QuestID); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleFailQuest",
			"quest_id": req.QuestID,
		}).Error("failed to fail quest")
		return nil, fmt.Errorf("failed to fail quest: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"function": "handleFailQuest",
		"quest_id": req.QuestID,
	}).Debug("exiting handleFailQuest")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.QuestID,
		"message":  "Quest failed successfully",
	}, nil
}

// handleGetQuest processes a request to retrieve a specific quest from a player's quest log.
// This handler returns quest details including objectives and current status.
//
// Parameters:
//   - params: json.RawMessage containing the get quest request with:
//   - session_id: string - The session ID of the requesting player
//   - quest_id: string - The ID of the quest to retrieve
//
// Returns:
//   - interface{}: Quest data with full details
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest not found in player's quest log
func (s *RPCServer) handleGetQuest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetQuest",
	})
	logger.Debug("entering handleGetQuest")

	var req struct {
		SessionID string `json:"session_id"`
		QuestID   string `json:"quest_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetQuest",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleGetQuest",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Get quest from player
	quest, err := session.Player.GetQuest(req.QuestID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetQuest",
			"quest_id": req.QuestID,
		}).Error("failed to get quest")
		return nil, fmt.Errorf("failed to get quest: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"function": "handleGetQuest",
		"quest_id": req.QuestID,
	}).Debug("exiting handleGetQuest")

	return map[string]interface{}{
		"success": true,
		"quest":   quest,
	}, nil
}

// handleGetActiveQuests processes a request to retrieve all active quests for a player.
// This handler returns a list of quests that are currently in progress.
//
// Parameters:
//   - params: json.RawMessage containing the get active quests request with:
//   - session_id: string - The session ID of the requesting player
//
// Returns:
//   - interface{}: Array of active quest data
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
func (s *RPCServer) handleGetActiveQuests(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetActiveQuests",
	})
	logger.Debug("entering handleGetActiveQuests")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetActiveQuests",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleGetActiveQuests",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Get active quests from player
	activeQuests := session.Player.GetActiveQuests()

	logger.WithFields(logrus.Fields{
		"function":    "handleGetActiveQuests",
		"quest_count": len(activeQuests),
	}).Debug("exiting handleGetActiveQuests")

	return map[string]interface{}{
		"success":       true,
		"active_quests": activeQuests,
		"count":         len(activeQuests),
	}, nil
}

// handleGetCompletedQuests processes a request to retrieve all completed quests for a player.
// This handler returns a list of quests that have been successfully finished.
//
// Parameters:
//   - params: json.RawMessage containing the get completed quests request with:
//   - session_id: string - The session ID of the requesting player
//
// Returns:
//   - interface{}: Array of completed quest data
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
func (s *RPCServer) handleGetCompletedQuests(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetCompletedQuests",
	})
	logger.Debug("entering handleGetCompletedQuests")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetCompletedQuests",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleGetCompletedQuests",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Get completed quests from player
	completedQuests := session.Player.GetCompletedQuests()

	logger.WithFields(logrus.Fields{
		"function":    "handleGetCompletedQuests",
		"quest_count": len(completedQuests),
	}).Debug("exiting handleGetCompletedQuests")

	return map[string]interface{}{
		"success":          true,
		"completed_quests": completedQuests,
		"count":            len(completedQuests),
	}, nil
}

// handleGetQuestLog processes a request to retrieve the complete quest log for a player.
// This handler returns all quests regardless of status (active, completed, failed).
//
// Parameters:
//   - params: json.RawMessage containing the get quest log request with:
//   - session_id: string - The session ID of the requesting player
//
// Returns:
//   - interface{}: Complete quest log with all quest data
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
func (s *RPCServer) handleGetQuestLog(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetQuestLog",
	})
	logger.Debug("entering handleGetQuestLog")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetQuestLog",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleGetQuestLog",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Get complete quest log from player
	questLog := session.Player.GetQuestLog()

	logger.WithFields(logrus.Fields{
		"function":    "handleGetQuestLog",
		"quest_count": len(questLog),
	}).Debug("exiting handleGetQuestLog")

	return map[string]interface{}{
		"success":   true,
		"quest_log": questLog,
		"count":     len(questLog),
	}, nil
}

// Spell management handlers

// handleGetSpell retrieves a specific spell by ID from the spell database.
//
// Parameters (JSON):
//   - spell_id: string - The unique identifier of the spell to retrieve
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - spell: object containing the spell data
//
// Errors:
//   - "invalid spell ID" if spell_id is empty or not provided
//   - "spell not found" if the spell doesn't exist in the database
func (s *RPCServer) handleGetSpell(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpell",
	}).Debug("entering handleGetSpell")

	var req struct {
		SpellID string `json:"spell_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetSpell",
			"error":    err.Error(),
		}).Error("failed to unmarshal get spell parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid get spell parameters", err.Error())
	}

	if req.SpellID == "" {
		return nil, fmt.Errorf("spell ID cannot be empty")
	}

	spell, err := s.spellManager.GetSpell(req.SpellID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetSpell",
			"spellID":  req.SpellID,
			"error":    err.Error(),
		}).Error("spell not found")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpell",
		"spellID":  req.SpellID,
	}).Info("spell retrieved successfully")

	return map[string]interface{}{
		"success": true,
		"spell":   spell,
	}, nil
}

// handleGetSpellsByLevel retrieves all spells of a specific level.
//
// Parameters (JSON):
//   - level: int - The spell level to filter by (0 for cantrips, 1+ for leveled spells)
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - spells: array of spell objects
//   - count: int number of spells found
func (s *RPCServer) handleGetSpellsByLevel(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpellsByLevel",
	}).Debug("entering handleGetSpellsByLevel")

	var req struct {
		Level int `json:"level"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetSpellsByLevel",
			"error":    err.Error(),
		}).Error("failed to unmarshal get spells by level parameters")
		return nil, fmt.Errorf("invalid get spells by level parameters")
	}

	if req.Level < 0 {
		return nil, fmt.Errorf("spell level cannot be negative")
	}

	spells := s.spellManager.GetSpellsByLevel(req.Level)

	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpellsByLevel",
		"level":    req.Level,
		"count":    len(spells),
	}).Info("spells retrieved by level")

	return map[string]interface{}{
		"success": true,
		"spells":  spells,
		"count":   len(spells),
		"level":   req.Level,
	}, nil
}

// handleGetSpellsBySchool retrieves all spells of a specific magic school.
//
// Parameters (JSON):
//   - school: string - The magic school name (e.g., "Evocation", "Illusion")
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - spells: array of spell objects
//   - count: int number of spells found
//   - school: string the school name searched
func (s *RPCServer) handleGetSpellsBySchool(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpellsBySchool",
	}).Debug("entering handleGetSpellsBySchool")

	var req struct {
		School string `json:"school"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetSpellsBySchool",
			"error":    err.Error(),
		}).Error("failed to unmarshal get spells by school parameters")
		return nil, fmt.Errorf("invalid get spells by school parameters")
	}

	if req.School == "" {
		return nil, fmt.Errorf("school cannot be empty")
	}

	school := game.ParseSpellSchool(req.School)
	spells := s.spellManager.GetSpellsBySchool(school)

	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpellsBySchool",
		"school":   req.School,
		"count":    len(spells),
	}).Info("spells retrieved by school")

	return map[string]interface{}{
		"success": true,
		"spells":  spells,
		"count":   len(spells),
		"school":  req.School,
	}, nil
}

// handleGetAllSpells retrieves all spells in the spell database.
//
// Parameters: None required
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - spells: array of all spell objects
//   - count: int total number of spells
//   - by_level: map of spell counts by level
func (s *RPCServer) handleGetAllSpells(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetAllSpells",
	}).Debug("entering handleGetAllSpells")

	spells := s.spellManager.GetAllSpells()
	countsByLevel := s.spellManager.GetSpellCountByLevel()

	logrus.WithFields(logrus.Fields{
		"function": "handleGetAllSpells",
		"count":    len(spells),
	}).Info("all spells retrieved")

	return map[string]interface{}{
		"success":  true,
		"spells":   spells,
		"count":    len(spells),
		"by_level": countsByLevel,
	}, nil
}

// handleSearchSpells searches for spells by name, description, or keywords.
//
// Parameters (JSON):
//   - query: string - The search query to match against spell names, descriptions, and keywords
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if search was successful
//   - spells: array of matching spell objects
//   - count: int number of spells found
//   - query: string the search query used
func (s *RPCServer) handleSearchSpells(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleSearchSpells",
	}).Debug("entering handleSearchSpells")

	var req struct {
		Query string `json:"query"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleSearchSpells",
			"error":    err.Error(),
		}).Error("failed to unmarshal search spells parameters")
		return nil, fmt.Errorf("invalid search spells parameters")
	}

	if req.Query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	spells := s.spellManager.SearchSpells(req.Query)

	logrus.WithFields(logrus.Fields{
		"function": "handleSearchSpells",
		"query":    req.Query,
		"count":    len(spells),
	}).Info("spell search completed")

	return map[string]interface{}{
		"success": true,
		"spells":  spells,
		"count":   len(spells),
		"query":   req.Query,
	}, nil
}

// handleGetObjectsInRange processes a spatial query request for objects within a rectangular area.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - min_x: int minimum X coordinate of query rectangle
//   - min_y: int minimum Y coordinate of query rectangle
//   - max_x: int maximum X coordinate of query rectangle
//   - max_y: int maximum Y coordinate of query rectangle
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if query was successful
//   - objects: Array of game objects within the specified range
//   - count: Number of objects found
//   - error: Possible errors:
//   - "invalid range query parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
func (s *RPCServer) handleGetObjectsInRange(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetObjectsInRange",
	}).Debug("entering range query handler")

	var req struct {
		SessionID string `json:"session_id"`
		MinX      int    `json:"min_x"`
		MinY      int    `json:"min_y"`
		MaxX      int    `json:"max_x"`
		MaxY      int    `json:"max_y"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).Error("failed to unmarshal range query parameters")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid range query parameters",
		}, nil
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithField("sessionID", req.SessionID).Warn("range query attempted with invalid session")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid session",
		}, nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"sessionID": req.SessionID,
		"playerID":  session.Player.GetID(),
		"minX":      req.MinX,
		"minY":      req.MinY,
		"maxX":      req.MaxX,
		"maxY":      req.MaxY,
	})

	rect := game.Rectangle{
		MinX: req.MinX,
		MinY: req.MinY,
		MaxX: req.MaxX,
		MaxY: req.MaxY,
	}

	objects := s.state.WorldState.GetObjectsInRange(rect)
	logger.WithField("objectCount", len(objects)).Info("range query completed")

	return map[string]interface{}{
		"success": true,
		"objects": objects,
		"count":   len(objects),
	}, nil
}

// handleGetObjectsInRadius processes a spatial query request for objects within a circular area.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - center_x: int X coordinate of circle center
//   - center_y: int Y coordinate of query center
//   - radius: float64 radius of the search circle
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if query was successful
//   - objects: Array of game objects within the specified radius
//   - count: Number of objects found
//   - error: Possible errors:
//   - "invalid radius query parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
func (s *RPCServer) handleGetObjectsInRadius(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetObjectsInRadius",
	}).Debug("entering radius query handler")

	var req struct {
		SessionID string  `json:"session_id"`
		CenterX   int     `json:"center_x"`
		CenterY   int     `json:"center_y"`
		Radius    float64 `json:"radius"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).Error("failed to unmarshal radius query parameters")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid radius query parameters",
		}, nil
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithField("sessionID", req.SessionID).Warn("radius query attempted with invalid session")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid session",
		}, nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"sessionID": req.SessionID,
		"playerID":  session.Player.GetID(),
		"centerX":   req.CenterX,
		"centerY":   req.CenterY,
		"radius":    req.Radius,
	})

	center := game.Position{X: req.CenterX, Y: req.CenterY}
	objects := s.state.WorldState.GetObjectsInRadius(center, req.Radius)
	logger.WithField("objectCount", len(objects)).Info("radius query completed")

	return map[string]interface{}{
		"success": true,
		"objects": objects,
		"count":   len(objects),
	}, nil
}

// handleGetNearestObjects processes a spatial query request for the k nearest objects to a position.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - center_x: int X coordinate of query center
//   - center_y: int Y coordinate of query center
//   - k: int maximum number of nearest objects to return
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if query was successful
//   - objects: Array of k nearest game objects
//   - count: Number of objects found (may be less than k)
//   - error: Possible errors:
//   - "invalid nearest query parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
func (s *RPCServer) handleGetNearestObjects(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetNearestObjects",
	}).Debug("entering nearest objects query handler")

	var req struct {
		SessionID string `json:"session_id"`
		CenterX   int    `json:"center_x"`
		CenterY   int    `json:"center_y"`
		K         int    `json:"k"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).Error("failed to unmarshal nearest query parameters")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid nearest query parameters",
		}, nil
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithField("sessionID", req.SessionID).Warn("nearest query attempted with invalid session")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid session",
		}, nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"sessionID": req.SessionID,
		"playerID":  session.Player.GetID(),
		"centerX":   req.CenterX,
		"centerY":   req.CenterY,
		"k":         req.K,
	})

	center := game.Position{X: req.CenterX, Y: req.CenterY}
	objects := s.state.WorldState.GetNearestObjects(center, req.K)
	logger.WithField("objectCount", len(objects)).Info("nearest objects query completed")

	return map[string]interface{}{
		"success": true,
		"objects": objects,
		"count":   len(objects),
	}, nil
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

	var req struct {
		SessionID string `json:"session_id"`
		ItemID    string `json:"item_id"`
		TargetID  string `json:"target_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUseItem",
			"error":    err.Error(),
		}).Error("failed to unmarshal use item parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid use item parameters", err.Error())
	}

	if req.SessionID == "" {
		logrus.WithFields(logrus.Fields{
			"function": "handleUseItem",
		}).Warn("empty session ID")
		return nil, ErrInvalidSession
	}

	if req.ItemID == "" {
		logrus.WithFields(logrus.Fields{
			"function": "handleUseItem",
		}).Warn("empty item ID")
		return nil, fmt.Errorf("item ID is required")
	}

	s.mu.RLock()
	session, exists := s.sessions[req.SessionID]
	s.mu.RUnlock()

	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleUseItem",
			"sessionID": req.SessionID,
		}).Warn("session not found")
		return nil, ErrInvalidSession
	}

	if session.Player == nil {
		logrus.WithFields(logrus.Fields{
			"function":  "handleUseItem",
			"sessionID": req.SessionID,
		}).Warn("no player associated with session")
		return nil, fmt.Errorf("no player in session")
	}

	// Check if currently in combat (items can also be used outside combat)
	if s.state.TurnManager.IsInCombat {
		// If in combat, validate turn order
		if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
			logrus.WithFields(logrus.Fields{
				"function": "handleUseItem",
				"playerID": session.Player.GetID(),
			}).Warn("player attempted to use item when not their turn")
			return nil, fmt.Errorf("not your turn")
		}
	}

	logrus.WithFields(logrus.Fields{
		"function":  "handleUseItem",
		"sessionID": req.SessionID,
		"itemID":    req.ItemID,
		"targetID":  req.TargetID,
	}).Info("using item from inventory")

	// Find the item in the player's inventory
	item := findInventoryItem(session.Player.Character.Inventory, req.ItemID)
	if item == nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUseItem",
			"itemID":   req.ItemID,
		}).Error("failed to find item in inventory")
		return map[string]interface{}{
			"success": false,
			"effect":  fmt.Sprintf("Item %s not found in inventory", req.ItemID),
		}, nil
	}

	// For now, implement basic item usage (can be expanded later)
	effect := fmt.Sprintf("Used %s", item.Name)
	if req.TargetID != "" {
		effect = fmt.Sprintf("Used %s on %s", item.Name, req.TargetID)
	}

	// If the item is consumable, remove it from inventory
	if item.Type == "consumable" {
		// Remove one quantity of the item
		logrus.WithFields(logrus.Fields{
			"function": "handleUseItem",
			"itemID":   req.ItemID,
		}).Info("removing consumable item from inventory")
		// Note: This is a simplified implementation. In a full implementation,
		// you would handle item quantities and removal properly.
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleUseItem",
		"effect":   effect,
	}).Info("item used successfully")

	return map[string]interface{}{
		"success": true,
		"effect":  effect,
	}, nil
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

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleLeaveGame",
			"error":    err.Error(),
		}).Error("failed to unmarshal leave game parameters")
		return nil, fmt.Errorf("invalid leave game parameters")
	}

	if req.SessionID == "" {
		logrus.WithFields(logrus.Fields{
			"function": "handleLeaveGame",
		}).Warn("empty session ID")
		return nil, ErrInvalidSession
	}

	s.mu.Lock()
	session, exists := s.sessions[req.SessionID]
	if exists {
		// Close WebSocket connection if it exists
		if session.WSConn != nil {
			if err := session.WSConn.Close(); err != nil {
				logrus.WithFields(logrus.Fields{
					"function":  "handleLeaveGame",
					"sessionID": req.SessionID,
					"error":     err.Error(),
				}).Warn("failed to close WebSocket connection")
			}
		}

		// Close message channel
		if session.MessageChan != nil {
			close(session.MessageChan)
		}

		// Remove player from game state
		if session.Player != nil {
			// Remove player from world state objects
			if s.state.WorldState != nil && s.state.WorldState.Objects != nil {
				delete(s.state.WorldState.Objects, session.Player.GetID())
			}
		}

		// Remove session from sessions map
		delete(s.sessions, req.SessionID)

		logrus.WithFields(logrus.Fields{
			"function":  "handleLeaveGame",
			"sessionID": req.SessionID,
		}).Info("player left game and session removed")
	}
	s.mu.Unlock()

	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleLeaveGame",
			"sessionID": req.SessionID,
		}).Warn("session not found")
		return nil, ErrInvalidSession
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// PCG (Procedural Content Generation) handlers

// handleGenerateContent generates procedural content on demand
func (s *RPCServer) handleGenerateContent(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGenerateContent",
	}).Debug("entering handleGenerateContent")

	var req struct {
		SessionID   string                 `json:"session_id"`
		ContentType string                 `json:"content_type"`
		LocationID  string                 `json:"location_id"`
		Difficulty  int                    `json:"difficulty"`
		Constraints map[string]interface{} `json:"constraints"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGenerateContent",
			"error":    err.Error(),
		}).Error("failed to unmarshal content generation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid content generation parameters", err.Error())
	}
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	if req.ContentType == "" {
		return nil, fmt.Errorf("content_type parameter required")
	}

	if req.LocationID == "" {
		return nil, fmt.Errorf("location_id parameter required")
	}

	if req.Difficulty == 0 {
		req.Difficulty = 5 // Default difficulty
	}

	ctx := context.Background()

	// Generate content based on type using the available PCGManager methods
	var content interface{}
	switch pcg.ContentType(req.ContentType) {
	case pcg.ContentTypeTerrain:
		content, err = s.pcgManager.GenerateTerrainForLevel(ctx, req.LocationID, 50, 50, pcg.BiomeDungeon, req.Difficulty)
	case pcg.ContentTypeItems:
		content, err = s.pcgManager.GenerateItemsForLocation(ctx, req.LocationID, 3, pcg.RarityCommon, pcg.RarityRare, req.Difficulty)
	case pcg.ContentTypeLevels:
		content, err = s.pcgManager.GenerateDungeonLevel(ctx, req.LocationID, 5, 15, pcg.ThemeClassic, req.Difficulty)
	case pcg.ContentTypeQuests:
		content, err = s.pcgManager.GenerateQuestForArea(ctx, req.LocationID, pcg.QuestTypeFetch, req.Difficulty)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", req.ContentType)
	}

	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":    "handleGenerateContent",
		"sessionID":   req.SessionID,
		"contentType": req.ContentType,
		"locationID":  req.LocationID,
		"difficulty":  req.Difficulty,
	}).Info("content generated successfully")

	return map[string]interface{}{
		"success":      true,
		"content_type": req.ContentType,
		"location_id":  req.LocationID,
		"content":      content,
		"difficulty":   req.Difficulty,
	}, nil
}

// handleRegenerateTerrain regenerates terrain for a specific area
func (s *RPCServer) handleRegenerateTerrain(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleRegenerateTerrain",
	}).Debug("entering handleRegenerateTerrain")

	var req struct {
		SessionID    string  `json:"session_id"`
		LocationID   string  `json:"location_id"`
		Width        int     `json:"width"`
		Height       int     `json:"height"`
		BiomeType    string  `json:"biome_type"`
		Density      float64 `json:"density"`
		WaterLevel   float64 `json:"water_level"`
		Connectivity string  `json:"connectivity"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleRegenerateTerrain",
			"error":    err.Error(),
		}).Error("failed to unmarshal terrain regeneration parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid terrain parameters", err.Error())
	}
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	if req.LocationID == "" {
		return nil, fmt.Errorf("location_id parameter required")
	}

	// Set defaults
	if req.Width == 0 {
		req.Width = 50
	}
	if req.Height == 0 {
		req.Height = 50
	}
	if req.BiomeType == "" {
		req.BiomeType = "forest"
	}
	if req.Density == 0 {
		req.Density = 0.5
	}
	if req.WaterLevel == 0 {
		req.WaterLevel = 0.3
	}
	if req.Connectivity == "" {
		req.Connectivity = "moderate"
	}

	ctx := context.Background()

	// Convert biome type string to PCG BiomeType
	biomeType := pcg.BiomeType(req.BiomeType)

	gameMap, err := s.pcgManager.GenerateTerrainForLevel(ctx, req.LocationID, req.Width, req.Height, biomeType, 5)
	if err != nil {
		return nil, fmt.Errorf("terrain generation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":   "handleRegenerateTerrain",
		"sessionID":  req.SessionID,
		"locationID": req.LocationID,
		"width":      req.Width,
		"height":     req.Height,
		"biomeType":  req.BiomeType,
	}).Info("terrain regenerated successfully")

	return map[string]interface{}{
		"success":     true,
		"location_id": req.LocationID,
		"terrain":     gameMap,
		"width":       req.Width,
		"height":      req.Height,
		"biome_type":  req.BiomeType,
	}, nil
}

// handleGenerateItems generates items for a location
func (s *RPCServer) handleGenerateItems(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGenerateItems",
	}).Debug("entering handleGenerateItems")

	var req struct {
		SessionID   string   `json:"session_id"`
		LocationID  string   `json:"location_id"`
		Count       int      `json:"count"`
		MinRarity   string   `json:"min_rarity"`
		MaxRarity   string   `json:"max_rarity"`
		PlayerLevel int      `json:"player_level"`
		ItemTypes   []string `json:"item_types"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGenerateItems",
			"error":    err.Error(),
		}).Error("failed to unmarshal item generation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid item generation parameters", err.Error())
	}
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	if req.LocationID == "" {
		return nil, fmt.Errorf("location_id parameter required")
	}

	// Set defaults
	if req.Count == 0 {
		req.Count = 3
	}
	if req.MinRarity == "" {
		req.MinRarity = "common"
	}
	if req.MaxRarity == "" {
		req.MaxRarity = "rare"
	}
	if req.PlayerLevel == 0 {
		req.PlayerLevel = 5
	}

	ctx := context.Background()

	// Convert rarity strings to PCG RarityTier
	minRarity := pcg.RarityTier(req.MinRarity)
	maxRarity := pcg.RarityTier(req.MaxRarity)

	items, err := s.pcgManager.GenerateItemsForLocation(ctx, req.LocationID, req.Count, minRarity, maxRarity, req.PlayerLevel)
	if err != nil {
		return nil, fmt.Errorf("item generation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":       "handleGenerateItems",
		"sessionID":      req.SessionID,
		"locationID":     req.LocationID,
		"count":          req.Count,
		"playerLevel":    req.PlayerLevel,
		"itemsGenerated": len(items),
	}).Info("items generated successfully")

	return map[string]interface{}{
		"success":     true,
		"location_id": req.LocationID,
		"items":       items,
		"count":       len(items),
	}, nil
}

// handleGenerateLevel generates a complete level/dungeon
func (s *RPCServer) handleGenerateLevel(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGenerateLevel",
	}).Debug("entering handleGenerateLevel")

	var req struct {
		SessionID     string `json:"session_id"`
		Width         int    `json:"width"`
		Height        int    `json:"height"`
		RoomCount     int    `json:"room_count"`
		Theme         string `json:"theme"`
		Difficulty    int    `json:"difficulty"`
		CorridorStyle string `json:"corridor_style"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGenerateLevel",
			"error":    err.Error(),
		}).Error("failed to unmarshal level generation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid level generation parameters", err.Error())
	}
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	// Set defaults
	if req.Width == 0 {
		req.Width = 50
	}
	if req.Height == 0 {
		req.Height = 50
	}
	if req.RoomCount == 0 {
		req.RoomCount = 8
	}
	if req.Theme == "" {
		req.Theme = "classic"
	}
	if req.Difficulty == 0 {
		req.Difficulty = 5
	}
	if req.CorridorStyle == "" {
		req.CorridorStyle = "straight"
	}

	ctx := context.Background()

	// Convert theme string to PCG LevelTheme
	theme := pcg.LevelTheme(req.Theme)

	level, err := s.pcgManager.GenerateDungeonLevel(ctx, "generated_level", 5, req.RoomCount, theme, req.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("level generation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":   "handleGenerateLevel",
		"sessionID":  req.SessionID,
		"width":      req.Width,
		"height":     req.Height,
		"roomCount":  req.RoomCount,
		"theme":      req.Theme,
		"difficulty": req.Difficulty,
	}).Info("level generated successfully")

	return map[string]interface{}{
		"success":        true,
		"level":          level,
		"width":          req.Width,
		"height":         req.Height,
		"room_count":     req.RoomCount,
		"theme":          req.Theme,
		"difficulty":     req.Difficulty,
		"corridor_style": req.CorridorStyle,
	}, nil
}

// handleGenerateQuest generates a procedural quest
func (s *RPCServer) handleGenerateQuest(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGenerateQuest",
	}).Debug("entering handleGenerateQuest")

	var req struct {
		SessionID     string `json:"session_id"`
		QuestType     string `json:"quest_type"`
		Difficulty    int    `json:"difficulty"`
		MinObjectives int    `json:"min_objectives"`
		MaxObjectives int    `json:"max_objectives"`
		RewardTier    string `json:"reward_tier"`
		NarrativeType string `json:"narrative_type"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGenerateQuest",
			"error":    err.Error(),
		}).Error("failed to unmarshal quest generation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid quest generation parameters", err.Error())
	}
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	// Set defaults
	if req.QuestType == "" {
		req.QuestType = "fetch"
	}
	if req.Difficulty == 0 {
		req.Difficulty = 5
	}
	if req.MinObjectives == 0 {
		req.MinObjectives = 1
	}
	if req.MaxObjectives == 0 {
		req.MaxObjectives = 3
	}
	if req.RewardTier == "" {
		req.RewardTier = "common"
	}
	if req.NarrativeType == "" {
		req.NarrativeType = "linear"
	}

	ctx := context.Background()

	// Convert quest type string to PCG QuestType
	questType := pcg.QuestType(req.QuestType)

	quest, err := s.pcgManager.GenerateQuestForArea(ctx, "generated_quest_area", questType, req.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("quest generation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":       "handleGenerateQuest",
		"sessionID":      req.SessionID,
		"questType":      req.QuestType,
		"difficulty":     req.Difficulty,
		"objectiveCount": len(quest.Objectives),
	}).Info("quest generated successfully")

	return map[string]interface{}{
		"success":        true,
		"quest":          quest,
		"quest_type":     req.QuestType,
		"difficulty":     req.Difficulty,
		"min_objectives": req.MinObjectives,
		"max_objectives": req.MaxObjectives,
		"reward_tier":    req.RewardTier,
		"narrative_type": req.NarrativeType,
	}, nil
}

// handleGetPCGStats returns statistics about the PCG system
func (s *RPCServer) handleGetPCGStats(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetPCGStats",
	}).Debug("entering handleGetPCGStats")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetPCGStats",
			"error":    err.Error(),
		}).Error("failed to unmarshal PCG stats parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid PCG stats parameters", err.Error())
	}

	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	// Get PCG statistics
	stats := s.pcgManager.GetGenerationStatistics()

	logrus.WithFields(logrus.Fields{
		"function":  "handleGetPCGStats",
		"sessionID": req.SessionID,
	}).Info("PCG stats retrieved successfully")

	return map[string]interface{}{
		"success": true,
		"stats":   stats,
	}, nil
}

// handleValidateContent validates generated content
func (s *RPCServer) handleValidateContent(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleValidateContent",
	}).Debug("entering handleValidateContent")

	var req struct {
		SessionID   string      `json:"session_id"`
		ContentType string      `json:"content_type"`
		Content     interface{} `json:"content"`
		Strict      bool        `json:"strict"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleValidateContent",
			"error":    err.Error(),
		}).Error("failed to unmarshal content validation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid content validation parameters", err.Error())
	}

	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	if req.ContentType == "" {
		return nil, fmt.Errorf("content_type parameter required")
	}

	if req.Content == nil {
		return nil, fmt.Errorf("content parameter required")
	}

	// Validate content using PCG validator with type information
	validationResult, err := s.pcgManager.ValidateGeneratedContentWithType(req.Content, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":         "handleValidateContent",
		"sessionID":        req.SessionID,
		"contentType":      req.ContentType,
		"validationResult": validationResult.IsValid(),
	}).Info("content validated successfully")

	return map[string]interface{}{
		"success":      true,
		"valid":        validationResult.IsValid(),
		"errors":       validationResult.Errors,
		"warnings":     validationResult.Warnings,
		"content_type": req.ContentType,
		"strict":       req.Strict,
	}, nil
}
