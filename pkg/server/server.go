package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"goldbox-rpg/pkg/game"
)

// RPCServer represents the main RPC server instance that handles game state and player sessions.
// It provides functionality for managing game state, player sessions, and event handling.
//
// Fields:
//   - state: Pointer to GameState that maintains the current game state
//   - eventSys: Pointer to game.EventSystem for handling game events
//   - mu: RWMutex for thread-safe access to server resources
//   - timekeeper: Pointer to TimeManager for managing game time and scheduling
//   - sessions: Map of player session IDs to PlayerSession objects
//
// Related types:
//   - GameState
//   - game.EventSystem
//   - TimeManager
//   - PlayerSession
type RPCServer struct {
	webDir     string
	fileServer http.Handler
	state      *GameState
	eventSys   *game.EventSystem
	mu         sync.RWMutex
	timekeeper *TimeManager
	sessions   map[string]*PlayerSession
}

// NewRPCServer creates and initializes a new RPCServer instance with default configuration.
// It sets up the core game systems including:
//   - World state management
//   - Turn-based gameplay handling
//   - Time tracking and management
//   - Player session tracking
//
// Returns:
//   - *RPCServer: A fully initialized server instance ready to handle RPC requests
//
// Related types:
//   - GameState: Contains the core game state
//   - TurnManager: Manages turn order and progression
//   - TimeManager: Handles in-game time tracking
//   - PlayerSession: Tracks individual player connections
//   - EventSystem: Handles game event dispatching
func NewRPCServer(webDir string) *RPCServer {
	return &RPCServer{
		webDir:     webDir,
		fileServer: http.FileServer(http.Dir(webDir)),
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

// ServeHTTP handles incoming JSON-RPC requests over HTTP, implementing the http.Handler interface.
// It processes POST requests only and expects a JSON-RPC 2.0 formatted request body.
//
// Parameters:
//   - w http.ResponseWriter: The response writer for sending the HTTP response
//   - r *http.Request: The incoming HTTP request containing the JSON-RPC payload
//
// The request body should contain a JSON object with:
//   - jsonrpc: String specifying the JSON-RPC version (must be "2.0")
//   - method: The RPC method name to invoke
//   - params: The parameters for the method (as raw JSON)
//   - id: Request identifier that will be echoed back in the response
//
// Error handling:
//   - Returns 405 Method Not Allowed if request is not POST
//   - Returns JSON-RPC error code -32700 for invalid JSON
//   - Returns JSON-RPC error code -32603 for internal errors during method execution
//
// Related:
//   - handleMethod: Processes the individual RPC method calls
//   - writeResponse: Formats and sends successful responses
//   - writeError: Formats and sends error responses
func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Upgrade") == "websocket" {
		s.HandleWebSocket(w, r)
		return
	}

	if r.Method != http.MethodPost {
		s.fileServer.ServeHTTP(w, r)
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

// handleMethod processes an RPC method call with the given parameters and returns the appropriate response.
// It uses a mutex to ensure thread-safe access to shared resources.
//
// Parameters:
//   - method: RPCMethod - The RPC method to be executed (e.g. MethodMove, MethodAttack, etc)
//   - params: json.RawMessage - The raw JSON parameters for the method call
//
// Returns:
//   - interface{} - The result of the method execution
//   - error - Any error that occurred during execution
//
// Error cases:
//   - Returns error if the method is not recognized
//
// Related methods:
//   - handleMove
//   - handleAttack
//   - handleCastSpell
//   - handleApplyEffect
//   - handleStartCombat
//   - handleEndTurn
//   - handleGetGameState
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

// writeResponse writes a JSON-RPC 2.0 compliant response to the http.ResponseWriter
//
// Parameters:
//   - w http.ResponseWriter: The response writer to write the JSON response to
//   - result interface{}: The result payload to include in the response
//   - id interface{}: The JSON-RPC request ID to correlate the response
//
// The function sets the Content-Type header to application/json and writes a JSON object
// containing the JSON-RPC version (2.0), the result, and the request ID.
//
// No error handling is currently implemented - errors from json.Encode are silently ignored.
// Consider adding error handling in production code.
//
// Related:
// - JSON-RPC 2.0 Specification: https://www.jsonrpc.org/specification
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

// writeError writes a JSON-RPC 2.0 error response to the provided http.ResponseWriter
//
// Parameters:
//   - w http.ResponseWriter: The response writer to write the error to
//   - code int: The error code to include in the response
//   - message string: The error message to include in the response
//   - data interface{}: Optional additional error data (will be omitted if nil)
//
// The function writes the error as a JSON object with the following structure:
//
//	{
//	  "jsonrpc": "2.0",
//	  "error": {
//	    "code": <code>,
//	    "message": <message>,
//	    "data": <data>  // Optional
//	  },
//	  "id": null
//	}
//
// The Content-Type header is set to application/json
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
