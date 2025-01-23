package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "NewRPCServer",
		"webDir":   webDir,
	})
	logger.Debug("entering NewRPCServer")

	server := &RPCServer{
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

	server.startSessionCleanup()

	logger.WithField("server", server).Info("initialized new RPC server")
	logger.Debug("exiting NewRPCServer")
	return server
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "ServeHTTP",
		"method":   r.Method,
		"url":      r.URL.String(),
	})
	logger.Debug("entering ServeHTTP")

	session, err := s.getOrCreateSession(w, r)
	if err != nil {
		logger.WithError(err).Error("session creation failed")
		writeError(w, -32603, "Internal error", nil)
		return
	}

	ctx := context.WithValue(r.Context(), "session", session)
	r = r.WithContext(ctx)

	if r.Header.Get("Upgrade") == "websocket" {
		s.HandleWebSocket(w, r)
		return
	}

	if r.Method != http.MethodPost {
		logger.Info("serving static file")
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
		logger.WithError(err).Error("failed to decode request body")
		writeError(w, -32700, "Parse error", nil)
		return
	}

	logger.WithFields(logrus.Fields{
		"rpcMethod": req.Method,
		"requestId": req.ID,
	}).Info("handling RPC method")

	// Handle the RPC method
	result, err := s.handleMethod(req.Method, req.Params)
	if err != nil {
		logger.WithError(err).Error("method handler failed")
		writeError(w, -32603, err.Error(), nil)
		return
	}

	// Write successful response
	writeResponse(w, result, req.ID)
	logger.Debug("exiting ServeHTTP")
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleMethod",
		"method":   method,
	})
	logger.Debug("entering handleMethod")

	s.mu.Lock()
	defer s.mu.Unlock()

	var result interface{}
	var err error

	switch method {
	case MethodMove:
		logger.Info("handling move method")
		result, err = s.handleMove(params)
	case MethodAttack:
		logger.Info("handling attack method")
		result, err = s.handleAttack(params)
	case MethodCastSpell:
		logger.Info("handling cast spell method")
		result, err = s.handleCastSpell(params)
	case MethodApplyEffect:
		logger.Info("handling apply effect method")
		result, err = s.handleApplyEffect(params)
	case MethodStartCombat:
		logger.Info("handling start combat method")
		result, err = s.handleStartCombat(params)
	case MethodEndTurn:
		logger.Info("handling end turn method")
		result, err = s.handleEndTurn(params)
	case MethodGetGameState:
		logger.Info("handling get game state method")
		result, err = s.handleGetGameState(params)
	default:
		err = fmt.Errorf("unknown method: %s", method)
		logger.WithError(err).Error("unknown method")
		return nil, err
	}

	if err != nil {
		logger.WithError(err).Error("method handler failed")
		return nil, err
	}

	logger.WithField("result", result).Debug("exiting handleMethod")
	return result, nil
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "writeResponse",
	})
	logger.Debug("entering writeResponse")

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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("failed to encode response")
		return
	}

	logger.WithFields(logrus.Fields{
		"response": response,
	}).Info("wrote response")
	logger.Debug("exiting writeResponse")
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "writeError",
		"code":     code,
		"message":  message,
	})
	logger.Debug("entering writeError")

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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("failed to encode error response")
		return
	}

	logger.WithFields(logrus.Fields{
		"response": response,
	}).Info("wrote error response")
	logger.Debug("exiting writeError")
}
