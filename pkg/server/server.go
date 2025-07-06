package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// Session configuration constants are defined in constants.go

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
/*type RPCServer struct {
	webDir     string
	fileServer http.Handler
	state      *GameState
	eventSys   *game.EventSystem
	mu         sync.RWMutex
	timekeeper *TimeManager
	sessions   map[string]*PlayerSession
}*/

// RPCServer handles RPC requests and maintains game state.
type RPCServer struct {
	webDir       string
	fileServer   http.Handler
	state        *GameState
	eventSys     *game.EventSystem
	mu           sync.RWMutex
	timekeeper   *TimeManager
	sessions     map[string]*PlayerSession
	done         chan struct{}
	spellManager *game.SpellManager
	Addr         net.Addr // Address the server is listening on
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

	// Initialize spell manager
	spellsDir := "data/spells"
	spellManager := game.NewSpellManager(spellsDir)

	// Load spells from YAML files
	if err := spellManager.LoadSpells(); err != nil {
		logger.WithError(err).Warn("failed to load spells, continuing with empty spell database")
	} else {
		logger.WithField("spellCount", spellManager.GetSpellCount()).Info("loaded spells from YAML files")
	}

	// Create server with default world
	server := &RPCServer{
		webDir:     webDir,
		fileServer: http.FileServer(http.Dir(webDir)),
		state: &GameState{
			WorldState:  game.CreateDefaultWorld(), // Use default world
			TurnManager: NewTurnManager(),
			TimeManager: NewTimeManager(),
			Sessions:    make(map[string]*PlayerSession),
			Version:     1,
		},
		eventSys:     game.NewEventSystem(),
		sessions:     make(map[string]*PlayerSession),
		timekeeper:   NewTimeManager(),
		done:         make(chan struct{}),
		spellManager: spellManager,
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
//
// ADDED: ServeHTTP implements the http.Handler interface for processing HTTP requests.
// It handles both static file serving and JSON-RPC method calls with session management.
//
// Request routing:
// - WebSocket upgrade requests: Routed to HandleWebSocket
// - Static file requests: Served from configured web directory
// - JSON-RPC requests: Parsed and routed to appropriate method handlers
//
// Session management: Automatically creates or retrieves player sessions
// Error handling: Returns proper JSON-RPC error codes for various failure scenarios
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
		JSONRPC string          `json:"jsonrpc"`
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
//
// ADDED: handleMethod routes RPC method calls to their appropriate handler functions.
// It serves as the central dispatcher for all game-related RPC operations.
//
// Supported method categories:
// - Character actions: move, attack, castSpell, useItem
// - Combat management: startCombat, endTurn
// - Equipment: equipItem, unequipItem, getEquipment
// - Quest system: startQuest, completeQuest, failQuest, etc.
// - Spell queries: getSpell, getSpellsByLevel, etc.
// - Spatial queries: getObjectsInRange, getNearestObjects
// - Game state: getGameState, joinGame, leaveGame
//
// All handlers receive JSON-encoded parameters and return serializable results.
func (s *RPCServer) handleMethod(method RPCMethod, params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleMethod",
		"method":   method,
	})
	logger.Debug("entering handleMethod")

	var result interface{}
	var err error

	switch method {
	case MethodJoinGame:
		logger.Info("handling join game method")
		result, err = s.handleJoinGame(params)
	case MethodCreateCharacter:
		logger.Info("handling create character method")
		result, err = s.handleCreateCharacter(params)
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
	case MethodEquipItem:
		logger.Info("handling equip item method")
		result, err = s.handleEquipItem(params)
	case MethodUnequipItem:
		logger.Info("handling unequip item method")
		result, err = s.handleUnequipItem(params)
	case MethodGetEquipment:
		logger.Info("handling get equipment method")
		result, err = s.handleGetEquipment(params)
	case MethodStartQuest:
		logger.Info("handling start quest method")
		result, err = s.handleStartQuest(params)
	case MethodCompleteQuest:
		logger.Info("handling complete quest method")
		result, err = s.handleCompleteQuest(params)
	case MethodUpdateObjective:
		logger.Info("handling update objective method")
		result, err = s.handleUpdateObjective(params)
	case MethodFailQuest:
		logger.Info("handling fail quest method")
		result, err = s.handleFailQuest(params)
	case MethodGetQuest:
		logger.Info("handling get quest method")
		result, err = s.handleGetQuest(params)
	case MethodGetActiveQuests:
		logger.Info("handling get active quests method")
		result, err = s.handleGetActiveQuests(params)
	case MethodGetCompletedQuests:
		logger.Info("handling get completed quests method")
		result, err = s.handleGetCompletedQuests(params)
	case MethodGetQuestLog:
		logger.Info("handling get quest log method")
		result, err = s.handleGetQuestLog(params)
	case MethodGetSpell:
		logger.Info("handling get spell method")
		result, err = s.handleGetSpell(params)
	case MethodGetSpellsByLevel:
		logger.Info("handling get spells by level method")
		result, err = s.handleGetSpellsByLevel(params)
	case MethodGetSpellsBySchool:
		logger.Info("handling get spells by school method")
		result, err = s.handleGetSpellsBySchool(params)
	case MethodGetAllSpells:
		logger.Info("handling get all spells method")
		result, err = s.handleGetAllSpells(params)
	case MethodSearchSpells:
		logger.Info("handling search spells method")
		result, err = s.handleSearchSpells(params)
	case MethodGetObjectsInRange:
		logger.Info("handling get objects in range method")
		result, err = s.handleGetObjectsInRange(params)
	case MethodGetObjectsInRadius:
		logger.Info("handling get objects in radius method")
		result, err = s.handleGetObjectsInRadius(params)
	case MethodGetNearestObjects:
		logger.Info("handling get nearest objects method")
		result, err = s.handleGetNearestObjects(params)
	case MethodUseItem:
		logger.Info("handling use item method")
		result, err = s.handleUseItem(params)
	case MethodLeaveGame:
		logger.Info("handling leave game method")
		result, err = s.handleLeaveGame(params)
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
		JSONRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
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
		JSONRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
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

// Stop gracefully shuts down the RPC server by closing the done channel.
// This signals background goroutines and services to terminate cleanly.
//
// The done channel is used for coordinating shutdown across:
// - Session cleanup routines
// - Background processing tasks
// - Event system cleanup
//
// This method should be called before process termination to ensure clean shutdown.
func (s *RPCServer) Stop() {
	close(s.done)
}

// Serve starts the HTTP server on the provided listener and begins handling requests.
// It configures the HTTP server and starts listening for incoming connections.
//
// Parameters:
//   - listener: Network listener to accept connections on (e.g., TCP, Unix socket)
//
// Returns:
//   - error: nil on clean shutdown, error if server fails to start or encounters issues
//
// Server lifecycle:
// 1. Sets the server address from the listener
// 2. Creates HTTP server with RPCServer as handler
// 3. Starts serving requests until Stop() is called or error occurs
// 4. Handles graceful shutdown scenarios
//
// The server will continue running until Stop() is called or a fatal error occurs.
func (s *RPCServer) Serve(listener net.Listener) error {
	logger := logrus.WithFields(logrus.Fields{
		"function": "Serve",
		"address":  listener.Addr().String(),
	})
	s.Addr = listener.Addr()
	logger.Info("starting RPC server")

	srv := &http.Server{
		Handler: s,
	}

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Error("server failed")
		return err
	}

	logger.Info("RPC server stopped")
	return nil
}
