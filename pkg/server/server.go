package server

import (
	"encoding/json"
	"fmt"
	"goldbox-rpg/pkg/game"
	"net/http"
	"sync"
)

// RPCServer handles all game server functionality
type RPCServer struct {
	state      *GameState
	eventSys   *game.EventSystem
	mu         sync.RWMutex
	timekeeper *TimeManager
	sessions   map[string]*PlayerSession
}

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

func writeResponse(w http.ResponseWriter, result interface{}, id interface{}) {
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
