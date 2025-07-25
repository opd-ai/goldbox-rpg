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
	rateLimiter   *RateLimiter               // Rate limiting system
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
//
// loadServerConfiguration loads and validates the server configuration from environment.
func loadServerConfiguration(logger *logrus.Entry) (*config.Config, *validation.InputValidator, error) {
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Error("failed to load configuration")
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	validator := validation.NewInputValidator(cfg.MaxRequestSize)
	return cfg, validator, nil
}

// initializeSpellManager creates and initializes the spell manager with spell data.
func initializeSpellManager(logger *logrus.Entry) (*game.SpellManager, error) {
	wd, err := os.Getwd()
	if err != nil {
		logger.WithError(err).Error("failed to get working directory")
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	spellsDir := "data/spells"
	if _, err := os.Stat(spellsDir); os.IsNotExist(err) {
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
	if err := spellManager.LoadSpells(); err != nil {
		logger.WithError(err).Error("failed to load spells - server cannot start without spell data")
		return nil, err
	}

	logger.WithField("spellCount", spellManager.GetSpellCount()).Info("loaded spells from YAML files")
	return spellManager, nil
}

// setupPCGManager initializes and configures the PCG manager with default generators.
func setupPCGManager(logger *logrus.Entry) (*pcg.PCGManager, error) {
	pcgManager := pcg.NewPCGManager(game.CreateDefaultWorld(), logrus.StandardLogger())
	pcgManager.InitializeWithSeed(time.Now().UnixNano())

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

	if err := pcgManager.RegisterDefaultGenerators(); err != nil {
		logger.WithError(err).Error("failed to register default generators")
		return nil, fmt.Errorf("failed to register default generators: %w", err)
	}

	logger.Info("initialized PCG manager with default generators")
	return pcgManager, nil
}

// createServerInstance constructs the main server instance with core components.
func createServerInstance(webDir string, cfg *config.Config, validator *validation.InputValidator, spellManager *game.SpellManager, pcgManager *pcg.PCGManager) *RPCServer {
	return &RPCServer{
		webDir:     webDir,
		fileServer: http.FileServer(http.Dir(webDir)),
		state: &GameState{
			WorldState:  game.CreateDefaultWorld(),
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
}

// configurePerformanceMonitoring sets up metrics, profiling, and performance monitoring components.
func configurePerformanceMonitoring(server *RPCServer, cfg *config.Config) {
	server.metrics = NewMetrics()
	server.healthChecker = NewHealthChecker(server)

	profilingConfig := ProfilingConfig{
		Enabled: cfg.EnableProfiling || cfg.EnableDevMode,
		Path:    "/debug/pprof",
	}
	server.profiling = NewProfilingServer(profilingConfig)
	server.perfMonitor = NewPerformanceMonitor(server.metrics, cfg.MetricsInterval)

	if cfg.AlertingEnabled {
		alertHandler := &LogAlertHandler{}
		thresholds := DefaultAlertThresholds()
		thresholds.CheckInterval = cfg.AlertingInterval
		server.perfAlerter = NewPerformanceAlerter(thresholds, alertHandler, server.metrics)
	}
}

// initializeNetworkComponents sets up WebSocket broadcasting and rate limiting.
func initializeNetworkComponents(server *RPCServer, cfg *config.Config, logger *logrus.Entry) {
	server.broadcaster = NewWebSocketBroadcaster(server)
	server.broadcaster.Start()

	if cfg.RateLimitEnabled {
		server.rateLimiter = NewRateLimiter(cfg)
		logger.WithFields(logrus.Fields{
			"requests_per_second": cfg.RateLimitRequestsPerSecond,
			"burst":               cfg.RateLimitBurst,
			"cleanup_interval":    cfg.RateLimitCleanupInterval,
		}).Info("rate limiting enabled")
	} else {
		logger.Info("rate limiting disabled")
	}
}

func NewRPCServer(webDir string) (*RPCServer, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "NewRPCServer",
		"webDir":   webDir,
	})
	logger.Debug("entering NewRPCServer")

	cfg, validator, err := loadServerConfiguration(logger)
	if err != nil {
		return nil, err
	}

	spellManager, err := initializeSpellManager(logger)
	if err != nil {
		return nil, err
	}

	pcgManager, err := setupPCGManager(logger)
	if err != nil {
		return nil, err
	}

	server := createServerInstance(webDir, cfg, validator, spellManager, pcgManager)
	configurePerformanceMonitoring(server, cfg)
	initializeNetworkComponents(server, cfg, logger)

	if server.perfMonitor != nil {
		go server.perfMonitor.Start()
	}
	if server.perfAlerter != nil {
		go server.perfAlerter.Start(context.Background())
	}

	server.startSessionCleanup()

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
	// Build and apply the full middleware chain for all requests
	// This ensures correlation IDs, logging, and recovery are applied consistently
	handler := RequestIDMiddleware(
		LoggingMiddleware(
			RecoveryMiddleware(
				http.HandlerFunc(s.serveHTTPWithMiddleware))))

	handler.ServeHTTP(w, r)
}

// checkRateLimit applies rate limiting to the request and returns true if the request should be allowed.
// If rate limited, it writes the appropriate error response and returns false.
func (s *RPCServer) checkRateLimit(w http.ResponseWriter, r *http.Request) bool {
	if s.rateLimiter == nil {
		return true
	}

	clientIP := getClientIP(r)
	if !s.rateLimiter.Allow(clientIP) {
		requestID := GetRequestID(r.Context())

		logrus.WithFields(logrus.Fields{
			"client_ip":  clientIP,
			"method":     r.Method,
			"path":       r.URL.Path,
			"request_id": requestID,
		}).Warn("request rate limited")

		w.Header().Set("Retry-After", "1")
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return false
	}
	return true
}

// handleObservabilityEndpoints processes health, readiness, liveness, and metrics endpoints.
// Returns true if the request was handled, false if it should continue to other handlers.
func (s *RPCServer) handleObservabilityEndpoints(w http.ResponseWriter, r *http.Request) bool {
	switch r.URL.Path {
	case "/health":
		if r.Method == http.MethodGet {
			// Apply metrics middleware to health endpoint too
			metricsHandler := s.metrics.MetricsMiddleware(http.HandlerFunc(s.healthChecker.HealthHandler))
			metricsHandler.ServeHTTP(w, r)
			return true
		}
	case "/ready":
		if r.Method == http.MethodGet {
			s.healthChecker.ReadinessHandler(w, r)
			return true
		}
	case "/live":
		if r.Method == http.MethodGet {
			s.healthChecker.LivenessHandler(w, r)
			return true
		}
	case "/metrics":
		if r.Method == http.MethodGet {
			s.metrics.GetHandler().ServeHTTP(w, r)
			return true
		}
	}
	return false
}

// handleProfilingEndpoints processes debug profiling endpoints when profiling is enabled.
// Returns true if the request was handled, false if it should continue to other handlers.
func (s *RPCServer) handleProfilingEndpoints(w http.ResponseWriter, r *http.Request) bool {
	if !(s.config.EnableProfiling || s.config.EnableDevMode) {
		return false
	}

	if r.URL.Path == "/debug/pprof" {
		http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
		return true
	}

	if len(r.URL.Path) > 12 && r.URL.Path[:12] == "/debug/pprof" {
		// Strip the path prefix and let the profiling server handle it
		r.URL.Path = r.URL.Path[0:] // Keep the full path for pprof
		s.profiling.server.Handler.ServeHTTP(w, r)
		return true
	}

	return false
}

// serveHTTPWithMiddleware handles requests after middleware has been applied
func (s *RPCServer) serveHTTPWithMiddleware(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithFields(logrus.Fields{
		"function":   "serveHTTPWithMiddleware",
		"method":     r.Method,
		"url":        r.URL.String(),
		"request_id": GetRequestID(r.Context()),
	})
	logger.Debug("entering serveHTTPWithMiddleware")

	// Apply rate limiting after middleware (so we have request ID for logging)
	if !s.checkRateLimit(w, r) {
		return
	}

	// Handle observability endpoints first
	if s.handleObservabilityEndpoints(w, r) {
		return
	}

	// Handle profiling endpoints (only when enabled)
	if s.handleProfilingEndpoints(w, r) {
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
		"request_id": GetRequestID(r.Context()),
	})
	logger.Debug("entering handleRequest")

	r, err := s.setupSessionContext(w, r, logger)
	if err != nil {
		return
	}

	if s.handleNonPOSTRequests(w, r, logger) {
		return
	}

	rpcRequest, err := s.parseJSONRPCRequest(r, logger)
	if err != nil {
		s.writeJSONRPCError(w, err, logger)
		return
	}

	if err := s.validateJSONRPCRequest(rpcRequest, logger); err != nil {
		s.writeJSONRPCError(w, err, logger)
		return
	}

	s.processRPCMethod(w, rpcRequest, logger)
	logger.Debug("exiting ServeHTTP")
}

// setupSessionContext creates and configures the session context for the request
func (s *RPCServer) setupSessionContext(w http.ResponseWriter, r *http.Request, logger *logrus.Entry) (*http.Request, error) {
	session, err := s.getOrCreateSession(w, r)
	if err != nil {
		logger.WithError(err).Error("session creation failed")
		writeError(w, JSONRPCInternalError, "Internal error", nil)
		return nil, err
	}
	defer s.releaseSession(session)

	ctx := context.WithValue(r.Context(), sessionKey, session)
	return r.WithContext(ctx), nil
}

// handleNonPOSTRequests processes WebSocket upgrades and static file requests
func (s *RPCServer) handleNonPOSTRequests(w http.ResponseWriter, r *http.Request, logger *logrus.Entry) bool {
	if r.Header.Get("Upgrade") == "websocket" {
		s.HandleWebSocket(w, r)
		return true
	}

	if r.Method != http.MethodPost {
		logger.Info("serving static file")
		s.fileServer.ServeHTTP(w, r)
		return true
	}

	return false
}

// JSONRPCRequest represents a parsed JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  RPCMethod       `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// parseJSONRPCRequest decodes and parses the JSON-RPC request from the request body
func (s *RPCServer) parseJSONRPCRequest(r *http.Request, logger *logrus.Entry) (*JSONRPCRequest, error) {
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode request body")
		return nil, &JSONRPCError{
			Code:    JSONRPCParseError,
			Message: "Parse error",
			Data:    nil,
		}
	}
	return &req, nil
}

// validateJSONRPCRequest validates the structure and required fields of a JSON-RPC request
func (s *RPCServer) validateJSONRPCRequest(req *JSONRPCRequest, logger *logrus.Entry) error {
	if req.JSONRPC != "2.0" {
		logger.Error("invalid JSON-RPC version")
		return &JSONRPCError{
			Code:    JSONRPCInvalidRequest,
			Message: "Invalid Request",
			Data:    "JSON-RPC version must be 2.0",
		}
	}

	if req.Method == "" {
		logger.Error("missing method in request")
		return &JSONRPCError{
			Code:    JSONRPCInvalidRequest,
			Message: "Invalid Request",
			Data:    "Method field is required",
		}
	}

	return nil
}

// writeJSONRPCError writes a JSON-RPC error response using the provided error
func (s *RPCServer) writeJSONRPCError(w http.ResponseWriter, err error, logger *logrus.Entry) {
	if jsonRPCErr, ok := err.(*JSONRPCError); ok {
		writeError(w, jsonRPCErr.Code, jsonRPCErr.Message, jsonRPCErr.Data)
	} else {
		writeError(w, JSONRPCInternalError, err.Error(), nil)
	}
}

// processRPCMethod handles the execution of an RPC method and writes the response
func (s *RPCServer) processRPCMethod(w http.ResponseWriter, req *JSONRPCRequest, logger *logrus.Entry) {
	logger.WithFields(logrus.Fields{
		"rpcMethod": req.Method,
		"requestId": req.ID,
	}).Info("handling RPC method")

	result, err := s.handleMethod(req.Method, req.Params)
	if err != nil {
		logger.WithError(err).Error("method handler failed")
		s.writeJSONRPCError(w, err, logger)
		return
	}

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

// Shutdown gracefully shuts down the RPCServer and all its components.
// It accepts a context for controlling shutdown timeout and cancellation.
//
// The shutdown process includes:
//   - Stopping the profiling server if running
//   - Closing the done channel to signal all background goroutines
//   - Gracefully shutting down performance monitoring components
//
// Parameters:
//   - ctx: context.Context for controlling shutdown timeout and cancellation
//
// Returns:
//   - error: nil on successful shutdown, error if any component fails to shut down gracefully
func (s *RPCServer) Shutdown(ctx context.Context) error {
	var shutdownErr error

	// Shutdown profiling server if it exists
	if s.profiling != nil {
		if err := s.profiling.Shutdown(ctx); err != nil {
			shutdownErr = err
		}
	}

	// Stop all background operations
	s.Stop()

	return shutdownErr
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
	logger.Info("starting RPC server with comprehensive middleware chain")

	// Build middleware chain: RequestID -> Logging -> Recovery -> Timeout -> Server
	handler := RequestIDMiddleware(
		LoggingMiddleware(
			s.withRecovery(
				s.withTimeout(s.config.RequestTimeout)(s))))

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
				// Get request ID from context (set by RequestIDMiddleware)
				requestID := GetRequestID(r.Context())
				if requestID == "" {
					requestID = uuid.New().String() // Fallback if middleware isn't used
				}

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

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Close performs cleanup of server resources including rate limiters and performance monitors.
// This should be called during server shutdown to prevent resource leaks.
func (s *RPCServer) Close() error {
	logger := logrus.WithField("function", "Close")
	logger.Info("shutting down server resources")

	// Stop rate limiter cleanup goroutine
	if s.rateLimiter != nil {
		s.rateLimiter.Close()
		logger.Debug("rate limiter closed")
	}

	// Stop performance monitoring
	if s.perfMonitor != nil {
		s.perfMonitor.Stop()
		logger.Debug("performance monitor stopped")
	}

	// Stop performance alerting
	if s.perfAlerter != nil {
		s.perfAlerter.Stop()
		logger.Debug("performance alerter stopped")
	}

	// Stop WebSocket broadcaster
	if s.broadcaster != nil {
		s.broadcaster.Stop()
		logger.Debug("websocket broadcaster stopped")
	}

	logger.Info("server shutdown complete")
	return nil
}
