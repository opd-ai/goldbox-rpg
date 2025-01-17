package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"goldbox-rpg/pkg/game"
)

// RPCServer handles all game server functionality
type RPCServer struct {
	state      *GameState
	eventSys   *game.EventSystem
	mu         sync.RWMutex
	timekeeper *TimeManager

	// Session management
	sessions map[string]*PlayerSession
}

// RPCMethod represents available RPC methods
type RPCMethod string

const (
	MethodMove         RPCMethod = "move"
	MethodAttack       RPCMethod = "attack"
	MethodCastSpell    RPCMethod = "castSpell"
	MethodUseItem      RPCMethod = "useItem"
	MethodApplyEffect  RPCMethod = "applyEffect"
	MethodStartCombat  RPCMethod = "startCombat"
	MethodEndTurn      RPCMethod = "endTurn"
	MethodGetGameState RPCMethod = "getGameState"
	MethodJoinGame     RPCMethod = "joinGame"
	MethodLeaveGame    RPCMethod = "leaveGame"
)

// NewRPCServer creates a new game server instance
func NewRPCServer() *RPCServer {
	return &RPCServer{
		state: &GameState{
			WorldState:  game.NewWorld(),
			TurnManager: &TurnManager{},
			TimeManager: NewTimeManager(),
			Sessions:    make(map[string]*PlayerSession),
		},
		eventSys:   game.NewEventSystem(),
		sessions:   make(map[string]*PlayerSession),
		timekeeper: NewTimeManager(),
	}
}

// ServeHTTP implements http.Handler
func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JsonRPC string          `json:"jsonrpc"`
		Method  RPCMethod       `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      interface{}     `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, -32700, "Parse error", nil)
		return
	}

	// Handle the RPC method
	result, err := s.handleMethod(req.Method, req.Params)
	if err != nil {
		writeError(w, -32603, err.Error(), nil)
		return
	}

	// Write successful response
	writeResponse(w, result, req.ID)
}

// handleMethod processes individual RPC methods
func (s *RPCServer) handleMethod(method RPCMethod, params json.RawMessage) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch method {
	case MethodMove:
		return s.handleMove(params)
	case MethodAttack:
		return s.handleAttack(params)
	case MethodCastSpell:
		return s.handleCastSpell(params)
	case MethodApplyEffect:
		return s.handleApplyEffect(params)
	case MethodStartCombat:
		return s.handleStartCombat(params)
	case MethodEndTurn:
		return s.handleEndTurn(params)
	case MethodGetGameState:
		return s.handleGetGameState(params)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// handleMove processes movement requests
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
	newPos := s.calculateNewPosition(currentPos, req.Direction)

	if err := s.state.WorldState.ValidateMove(player, newPos); err != nil {
		return nil, err
	}

	if err := player.SetPosition(newPos); err != nil {
		return nil, err
	}

	// Emit movement event
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

// calculateNewPosition calculates the new position based on the current position and direction
func (s *RPCServer) calculateNewPosition(currentPos game.Position, direction game.Direction) game.Position {
	newPos := currentPos
	switch direction {
	case game.North:
		newPos.Y--
	case game.South:
		newPos.Y++
	case game.East:
		newPos.X++
	case game.West:
		newPos.X--
	}
	return newPos
}

// processCombatAction processes the combat action for the given player, target, and weapon
func (s *RPCServer) processCombatAction(player *game.Player, targetID, weaponID string) (interface{}, error) {
	target, exists := s.state.WorldState.Objects[targetID]
	if !exists {
		return nil, fmt.Errorf("invalid target")
	}

	weapon := findInventoryItem(player.Inventory, weaponID)
	if weapon == nil {
		return nil, fmt.Errorf("weapon not found in inventory")
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

// handleAttack processes combat actions
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

	// Validate combat state
	if !s.state.TurnManager.IsInCombat {
		return nil, fmt.Errorf("not in combat")
	}

	if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
		return nil, fmt.Errorf("not your turn")
	}

	// Process attack
	result, err := s.processCombatAction(session.Player, req.TargetID, req.WeaponID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// applyDamage applies damage to the target
func (s *RPCServer) applyDamage(target game.GameObject, damage int) error {
	newHealth := target.GetHealth() - damage
	if newHealth <= 0 {
		newHealth = 0
	}
	target.SetHealth(newHealth)
	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventDeath,
		SourceID: target.GetID(),
	})
	return nil
}

// Helper functions
func writeResponse(w http.ResponseWriter, result, id interface{}) {
	response := struct {
		JsonRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JsonRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeError(w http.ResponseWriter, code int, message string, data interface{}) {
	response := struct {
		JsonRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}{
		JsonRPC: "2.0",
		Error: struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		}{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
