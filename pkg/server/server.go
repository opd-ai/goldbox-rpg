package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
	"goldbox-rpg/pkg/pcg/items"
	"goldbox-rpg/pkg/pcg/quests"
	"goldbox-rpg/pkg/validation"
)

// JSON-RPC 2.0 error codes
const (
	// Standard JSON-RPC 2.0 error codes
	JSONRPCParseError     = -32700 // Invalid JSON was received by the server
	JSONRPCInvalidRequest = -32600 // The JSON sent is not a valid Request object
	JSONRPCMethodNotFound = -32601 // The method does not exist / is not available
	JSONRPCInvalidParams  = -32602 // Invalid method parameter(s)
	JSONRPCInternalError  = -32603 // Internal JSON-RPC error
)

// Custom error types for JSON-RPC error handling
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *JSONRPCError) Error() string {
	return e.Message
}

// NewJSONRPCError creates a new JSON-RPC error with the specified code and message
func NewJSONRPCError(code int, message string, data interface{}) *JSONRPCError {
	return &JSONRPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

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
	webDir        string
	fileServer    http.Handler
	state         *GameState
	eventSys      *game.EventSystem
	mu            sync.RWMutex
	timekeeper    *TimeManager
	sessions      map[string]*PlayerSession
	done          chan struct{}
	spellManager  *game.SpellManager
	pcgManager    *pcg.PCGManager            // Procedural content generation manager
	Addr          net.Addr                   // Address the server is listening on
	broadcaster   *WebSocketBroadcaster      // WebSocket event broadcaster
	config        *config.Config             // Server configuration
	validator     *validation.InputValidator // Input validation
	healthChecker *HealthChecker             // Health check system
	metrics       *Metrics                   // Prometheus metrics
	profiling     *ProfilingServer           // Performance profiling server
	perfMonitor   *PerformanceMonitor        // Performance metrics monitor
	perfAlerter   *PerformanceAlerter        // Performance alerting system
}

// NewRPCServer creates and initializes a new RPCServer instance with configuration.
// It sets up the core game systems including:
//   - World state management
//   - Turn-based gameplay handling
//   - Time tracking and management
//   - Player session tracking
//   - Input validation and security controls
//
// Parameters:
//   - webDir: string path to web directory for static files
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
//   - InputValidator: Validates and sanitizes user input
func NewRPCServer(webDir string) (*RPCServer, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "NewRPCServer",
		"webDir":   webDir,
	})
	logger.Debug("entering NewRPCServer")

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Error("failed to load configuration")
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create input validator with configured request size limit
	validator := validation.NewInputValidator(cfg.MaxRequestSize)

	// Initialize spell manager - find spells directory relative to project root
	wd, err := os.Getwd()
	if err != nil {
		logger.WithError(err).Error("failed to get working directory")
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Look for data/spells from current directory or walk up to find project root
	spellsDir := "data/spells"
	if _, err := os.Stat(spellsDir); os.IsNotExist(err) {
		// Try relative to project root (for tests running from pkg/server)
		spellsDir = "../../data/spells"
		if _, err := os.Stat(spellsDir); os.IsNotExist(err) {
			logger.WithFields(logrus.Fields{
				"workingDir": wd,
				"attempted1": "data/spells",
				"attempted2": "../../data/spells",
			}).Error("could not find spells directory")
			return nil, fmt.Errorf("spells directory not found from working directory: %s", wd)
		}
	}

	spellManager := game.NewSpellManager(spellsDir)

	// Load spells from YAML files
	if err := spellManager.LoadSpells(); err != nil {
		logger.WithError(err).Error("failed to load spells - server cannot start without spell data")
		return nil, err
	}
	logger.WithField("spellCount", spellManager.GetSpellCount()).Info("loaded spells from YAML files")
	// Initialize PCG manager
	pcgManager := pcg.NewPCGManager(game.CreateDefaultWorld(), logrus.StandardLogger())
	pcgManager.InitializeWithSeed(time.Now().UnixNano()) // Use current time as seed

	// Register available generators
	questGen := quests.NewObjectiveBasedGenerator()
	if err := pcgManager.GetRegistry().RegisterGenerator("objective_based", questGen); err != nil {
		logger.WithError(err).Error("failed to register quest generator")
		return nil, fmt.Errorf("failed to register quest generator: %w", err)
	}

	itemGen := items.NewTemplateBasedGenerator()
	if err := pcgManager.GetRegistry().RegisterGenerator("template_based", itemGen); err != nil {
		logger.WithError(err).Error("failed to register item generator")
		return nil, fmt.Errorf("failed to register item generator: %w", err)
	}

	// Call RegisterDefaultGenerators to complete initialization
	if err := pcgManager.RegisterDefaultGenerators(); err != nil {
		logger.WithError(err).Error("failed to register default generators")
		return nil, fmt.Errorf("failed to register default generators: %w", err)
	}

	logger.Info("initialized PCG manager with default generators")

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
		pcgManager:   pcgManager,
		config:       cfg,
		validator:    validator,
	}

	// Initialize metrics system
	server.metrics = NewMetrics()

	// Initialize health checker with server reference
	server.healthChecker = NewHealthChecker(server)
	// Initialize performance monitoring components
	profilingConfig := ProfilingConfig{
		Enabled: cfg.EnableProfiling || cfg.EnableDevMode, // Enable profiling in dev mode or when explicitly enabled
		Path:    "/debug/pprof",
	}
	server.profiling = NewProfilingServer(profilingConfig)

	// Create performance monitor with configured interval
	server.perfMonitor = NewPerformanceMonitor(server.metrics, cfg.MetricsInterval)

	// Create performance alerter with default thresholds if alerting is enabled
	if cfg.AlertingEnabled {
		alertHandler := &LogAlertHandler{}
		thresholds := DefaultAlertThresholds()
		thresholds.CheckInterval = cfg.AlertingInterval
		server.perfAlerter = NewPerformanceAlerter(thresholds, alertHandler, server.metrics)
	}

	// Initialize and start WebSocket broadcaster
	server.broadcaster = NewWebSocketBroadcaster(server)
	server.broadcaster.Start()

	// Start performance monitoring in background if enabled
	if server.perfMonitor != nil {
		go server.perfMonitor.Start()
	}
	if server.perfAlerter != nil {
		go server.perfAlerter.Start(context.Background())
	}

	server.startSessionCleanup()

	// Initialize resilience components for production stability
	// Rate limiter: Allow 100 requests per second with burst of 200 per IP
	server.rateLimiter = NewRateLimiter(
		rate.Limit(cfg.RateLimit), // requests per second
		cfg.RateBurst,             // burst size
		5*time.Minute,             // cleanup interval
	)

	// Request size limiter: Protect against large request attacks
	server.requestLimiter = NewRequestSizeLimiter(cfg.MaxRequestSize)

	// Connection pool for external service calls
	poolConfig := DefaultConnectionPoolConfig()
	server.connectionPool = NewConnectionPool(poolConfig)

	// Initialize circuit breakers for external dependencies
	server.circuitBreakers = make(map[string]*CircuitBreaker)

	// Add default circuit breakers for common external services
	// These can be used by game logic that needs to call external APIs
	server.circuitBreakers["auth"] = NewCircuitBreaker("auth-service", DefaultCircuitBreakerConfig())
	server.circuitBreakers["content"] = NewCircuitBreaker("content-service", DefaultCircuitBreakerConfig())
	server.circuitBreakers["analytics"] = NewCircuitBreaker("analytics-service", DefaultCircuitBreakerConfig())

	logger.WithField("server", server).Info("initialized new RPC server")
	logger.Debug("exiting NewRPCServer")
	return server, nil
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

	// Add request correlation ID for tracing
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	w.Header().Set("X-Request-ID", requestID)
	ctx := context.WithValue(r.Context(), requestIDKey, requestID)
	r = r.WithContext(ctx)

	// Handle observability endpoints first
	switch r.URL.Path {
	case "/health":
		if r.Method == http.MethodGet {
			// Apply metrics middleware to health endpoint too
			metricsHandler := s.metrics.MetricsMiddleware(http.HandlerFunc(s.healthChecker.HealthHandler))
			metricsHandler.ServeHTTP(w, r)
			return
		}
	case "/ready":
		if r.Method == http.MethodGet {
			s.healthChecker.ReadinessHandler(w, r)
			return
		}
	case "/live":
		if r.Method == http.MethodGet {
			s.healthChecker.LivenessHandler(w, r)
			return
		}
	case "/metrics":
		if r.Method == http.MethodGet {
			s.metrics.GetHandler().ServeHTTP(w, r)
			return
		}
	}

	// Handle profiling endpoints (only when enabled)
	if (s.config.EnableProfiling || s.config.EnableDevMode) && r.URL.Path == "/debug/pprof" {
		http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
		return
	}
	if (s.config.EnableProfiling || s.config.EnableDevMode) && len(r.URL.Path) > 12 && r.URL.Path[:12] == "/debug/pprof" {
		// Strip the path prefix and let the profiling server handle it
		r.URL.Path = r.URL.Path[0:] // Keep the full path for pprof
		s.profiling.server.Handler.ServeHTTP(w, r)
		return
	}

	// Apply metrics middleware for all other requests
	metricsHandler := s.metrics.MetricsMiddleware(http.HandlerFunc(s.handleRequest))
	metricsHandler.ServeHTTP(w, r)
}

// handleRequest processes the actual game requests after middleware
func (s *RPCServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithFields(logrus.Fields{
		"function":   "handleRequest",
		"method":     r.Method,
		"url":        r.URL.String(),
		"request_id": r.Context().Value(requestIDKey),
	})
	logger.Debug("entering handleRequest")

	session, err := s.getOrCreateSession(w, r)
	if err != nil {
		logger.WithError(err).Error("session creation failed")
		writeError(w, JSONRPCInternalError, "Internal error", nil)
		return
	}
	defer s.releaseSession(session)

	ctx := context.WithValue(r.Context(), sessionKey, session)
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
		writeError(w, JSONRPCParseError, "Parse error", nil)
		return
	}

	// Validate JSON-RPC request structure
	if req.JSONRPC != "2.0" {
		logger.Error("invalid JSON-RPC version")
		writeError(w, JSONRPCInvalidRequest, "Invalid Request", "JSON-RPC version must be 2.0")
		return
	}

	if req.Method == "" {
		logger.Error("missing method in request")
		writeError(w, JSONRPCInvalidRequest, "Invalid Request", "Method field is required")
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

		// Check if it's a custom JSON-RPC error
		if jsonRPCErr, ok := err.(*JSONRPCError); ok {
			writeError(w, jsonRPCErr.Code, jsonRPCErr.Message, jsonRPCErr.Data)
		} else {
			// Default to internal error for other errors
			writeError(w, JSONRPCInternalError, err.Error(), nil)
		}
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

	// Parse params into interface{} for validation
	var paramsInterface interface{}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &paramsInterface); err != nil {
			return nil, NewJSONRPCError(JSONRPCParseError, "Invalid parameters format", err.Error())
		}
	}

	// Validate input parameters with request size check
	requestSize := int64(len(params))
	if err := s.validator.ValidateRPCRequest(string(method), paramsInterface, requestSize); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid method parameters", err.Error())
	}

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
	case MethodGenerateContent:
		logger.Info("handling generate content method")
		result, err = s.handleGenerateContent(params)
	case MethodRegenerateTerrain:
		logger.Info("handling regenerate terrain method")
		result, err = s.handleRegenerateTerrain(params)
	case MethodGenerateItems:
		logger.Info("handling generate items method")
		result, err = s.handleGenerateItems(params)
	case MethodGenerateLevel:
		logger.Info("handling generate level method")
		result, err = s.handleGenerateLevel(params)
	case MethodGenerateQuest:
		logger.Info("handling generate quest method")
		result, err = s.handleGenerateQuest(params)
	case MethodGetPCGStats:
		logger.Info("handling get PCG stats method")
		result, err = s.handleGetPCGStats(params)
	case MethodValidateContent:
		logger.Info("handling validate content method")
		result, err = s.handleValidateContent(params)
	default:
		err = NewJSONRPCError(JSONRPCMethodNotFound, fmt.Sprintf("Method not found: %s", method), nil)
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
	logger.Info("starting RPC server with resilience and security middleware")

	// Build middleware stack for production resilience
	// Order is important: outermost middleware runs first
	handler := http.Handler(s)

	// Core functionality middleware (innermost)
	handler = s.withTimeout(s.config.RequestTimeout)(handler)

	// Resilience middleware
	handler = RequestSizeLimitMiddleware(s.requestLimiter)(handler)
	handler = RateLimitMiddleware(s.rateLimiter)(handler)

	// Security and observability middleware
	handler = CORSMiddleware(s.config.AllowedOrigins)(handler)
	handler = s.metrics.MetricsMiddleware(handler)
	handler = LoggingMiddleware(handler)
	handler = RequestIDMiddleware(handler)

	// Recovery middleware (outermost - catches all panics)
	handler = RecoveryMiddleware(handler)

	srv := &http.Server{
		Handler: handler,
	}

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Error("server failed")
		return err
	}

	logger.Info("RPC server stopped")
	return nil
}

// withRecovery wraps an HTTP handler with panic recovery middleware.
// It prevents panics from crashing the server and logs them for debugging.
func (s *RPCServer) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Generate request ID for correlation
				requestID := generateRequestID()

				logrus.WithFields(logrus.Fields{
					"panic":       err,
					"request_id":  requestID,
					"method":      r.Method,
					"url":         r.URL.String(),
					"remote_addr": r.RemoteAddr,
					"user_agent":  r.Header.Get("User-Agent"),
				}).Error("recovered from panic in HTTP handler")

				// Set correlation ID header
				w.Header().Set("X-Request-ID", requestID)

				// Return JSON-RPC error response
				writeError(w, JSONRPCInternalError, "Internal Server Error", map[string]interface{}{
					"request_id": requestID,
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// withTimeout wraps an HTTP handler with context timeout middleware.
// It ensures requests don't run indefinitely and provides graceful timeout handling.
func (s *RPCServer) withTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Add request ID for correlation
			requestID := generateRequestID()
			ctx = context.WithValue(ctx, requestIDKey, requestID)
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// generateRequestID creates a unique request ID for correlation
func generateRequestID() string {
	return uuid.New().String()
}

// Shutdown gracefully shuts down the server and all its components.
// It stops the rate limiter cleanup, closes connection pools, and performs
// other necessary cleanup operations.
func (s *RPCServer) Shutdown(ctx context.Context) error {
	logger := logrus.WithField("function", "Shutdown")
	logger.Info("beginning graceful server shutdown")

	// Stop rate limiter cleanup goroutine
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
		logger.Debug("stopped rate limiter cleanup")
	}

	// Close connection pool to free system resources
	if s.connectionPool != nil {
		s.connectionPool.Close()
		logger.Debug("closed connection pool")
	}

	// Stop performance monitoring
	if s.perfMonitor != nil {
		s.perfMonitor.Stop()
		logger.Debug("stopped performance monitor")
	}

	// Stop performance alerting
	if s.perfAlerter != nil {
		s.perfAlerter.Stop()
		logger.Debug("stopped performance alerter")
	}

	// Stop WebSocket broadcaster
	if s.broadcaster != nil {
		s.broadcaster.Stop()
		logger.Debug("stopped WebSocket broadcaster")
	}

	// Stop profiling server if running separately
	if s.profiling != nil && s.profiling.server != nil {
		if err := s.profiling.server.Shutdown(ctx); err != nil {
			logger.WithError(err).Warn("error shutting down profiling server")
		} else {
			logger.Debug("stopped profiling server")
		}
	}

	logger.Info("graceful server shutdown completed")
	return nil
}
