package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// ErrInvalidSession is returned when a session ID is invalid or not found
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
		return nil, fmt.Errorf("invalid parameters")
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
	s.mu.RLock()
	session, exists := s.sessions[req.SessionID]
	s.mu.RUnlock()

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
		return nil, fmt.Errorf("invalid parameters")
	}

	// 2. Check session with read lock
	s.mu.RLock()
	session, exists := s.sessions[req.SessionID]
	s.mu.RUnlock()

	if !exists {
		logger.WithField("sessionID", req.SessionID).Warn("session not found")
		return nil, ErrInvalidSession
	}

	// 3. Update last active time with write lock
	s.mu.Lock()
	session.LastActive = time.Now()
	s.mu.Unlock()

	// 4. Get game state (uses its own internal locking)
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

func (s *RPCServer) handleJoinGame(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleJoinGame",
	}).Debug("entering handleJoinGame")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleJoinGame",
			"error":    err.Error(),
		}).Error("failed to unmarshal join parameters")
		return nil, fmt.Errorf("invalid join parameters")
	}

	if req.SessionID == "" {
		logrus.WithFields(logrus.Fields{
			"function": "handleJoinGame",
		}).Warn("empty session ID")
		return nil, ErrInvalidSession
	}

	s.mu.RLock()
	session, exists := s.sessions[req.SessionID]
	s.mu.RUnlock()

	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":  "handleJoinGame",
			"sessionID": req.SessionID,
		}).Warn("session not found")
		return nil, ErrInvalidSession
	}

	logrus.WithFields(logrus.Fields{
		"function":  "handleJoinGame",
		"sessionID": req.SessionID,
	}).Info("adding player to game state")

	// Initialize player in session
	s.state.AddPlayer(session)

	logrus.WithFields(logrus.Fields{
		"function": "handleJoinGame",
	}).Debug("exiting handleJoinGame")

	return map[string]interface{}{
		"player_id": session.SessionID,
		"state":     s.state.GetState(),
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

	var req struct {
		Name             string         `json:"name"`
		Class            string         `json:"class"`
		AttributeMethod  string         `json:"attribute_method"`
		CustomAttributes map[string]int `json:"custom_attributes,omitempty"`
		StartingEquipment bool          `json:"starting_equipment"`
		StartingGold     int            `json:"starting_gold"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleCreateCharacter",
			"error":    err.Error(),
		}).Error("failed to unmarshal character creation parameters")
		return nil, fmt.Errorf("invalid character creation parameters")
	}

	// Convert string class to CharacterClass enum
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
			"function": "handleCreateCharacter",
			"class":    req.Class,
		}).Error("invalid character class")
		return nil, fmt.Errorf("invalid character class: %s", req.Class)
	}

	// Set default starting gold if not specified
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

	// Create character creation config
	config := game.CharacterCreationConfig{
		Name:              req.Name,
		Class:             characterClass,
		AttributeMethod:   req.AttributeMethod,
		CustomAttributes:  req.CustomAttributes,
		StartingEquipment: req.StartingEquipment,
		StartingGold:      req.StartingGold,
	}

	// Create character creator and generate character
	creator := game.NewCharacterCreator()
	result := creator.CreateCharacter(config)

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

	// Create a new session for this character
	sessionID := game.NewUID()
	session := &PlayerSession{
		SessionID:   sessionID,
		Player:      result.PlayerData,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   false,
		MessageChan: make(chan []byte, 100),
	}

	// Store session
	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function":    "handleCreateCharacter",
		"sessionID":   sessionID,
		"characterName": req.Name,
		"class":       req.Class,
	}).Info("character created successfully")

	return map[string]interface{}{
		"success":    true,
		"character":  result.Character,
		"player":     result.PlayerData,
		"session_id": sessionID,
		"errors":     result.Errors,
		"warnings":   result.Warnings,
		"creation_time": result.CreationTime,
		"generated_stats": result.GeneratedStats,
		"starting_items": result.StartingItems,
	}, nil
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
		return nil, fmt.Errorf("invalid equip item parameters")
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
		"function":      "handleEquipItem",
		"sessionID":     req.SessionID,
		"itemID":        req.ItemID,
		"slot":          req.Slot,
		"equippedItem":  equippedItem.Name,
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
//     - success: bool indicating if unequipping was successful
//     - message: string describing the result
//     - unequipped_item: object containing details of the unequipped item
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
		return nil, fmt.Errorf("invalid unequip item parameters")
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
		"function":        "handleUnequipItem",
		"sessionID":       req.SessionID,
		"slot":            req.Slot,
		"unequippedItem":  unequippedItem.Name,
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
//     - success: bool indicating if retrieval was successful
//     - equipment: map of slot names to equipped item objects
//     - total_weight: int total weight of all equipped items
//     - equipment_bonuses: map of stat bonuses from equipment
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
		return nil, fmt.Errorf("invalid get equipment parameters")
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
		"function":      "handleGetEquipment",
		"sessionID":     req.SessionID,
		"numItems":      len(equipment),
		"totalWeight":   totalWeight,
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
