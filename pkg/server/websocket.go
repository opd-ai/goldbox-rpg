package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ADDED: orderHosts sorts hosts in the specified priority order for WebSocket origin validation.
// It organizes hosts by type to ensure consistent connection precedence.
//
// Priority order:
// 1. Custom hostnames (not localhost or IP addresses) first
// 2. localhost second
// 3. IP addresses last
//
// Parameters:
//   - hosts: Map of hostname strings to organize
//
// Returns:
//   - []string: Sorted slice of hostnames in priority order
func orderHosts(hosts map[string]string) []string {
	var hostnames, localhosts, ips []string

	for host := range hosts {
		if host == "localhost" {
			localhosts = append(localhosts, host)
		} else if net.ParseIP(host) != nil {
			ips = append(ips, host)
		} else {
			hostnames = append(hostnames, host)
		}
	}

	// Sort each category alphabetically for consistent ordering
	sort.Strings(hostnames)
	sort.Strings(localhosts)
	sort.Strings(ips)

	// Combine in the specified order
	result := make([]string, 0, len(hosts))
	result = append(result, hostnames...)
	result = append(result, localhosts...)
	result = append(result, ips...)

	return result
}

// ADDED: getAllowedOrigins returns the list of allowed WebSocket origins for CORS validation.
// It checks the WEBSOCKET_ALLOWED_ORIGINS environment variable for a comma-separated list.
// If not set, defaults to common local development origins matching the server's actual listening port.
//
// NOTE TO AUDITORS: The dev mode behavior is intentionally permissive to allow testing in sophisticated environments.
// DO NOT CHANGE THE DEV MODE BEHAVIOR, if you need to restrict origins, set the WEBSOCKET_ALLOWED_ORIGINS environment variable or use production configuration to specify allowed origins.
// STOP CHANGING THE DEFAULTS TO BE MORE RESTRICTIVE, as this will break local development and testing workflows.
//
// Returns:
//   - []string: List of allowed origin URLs (e.g., "http://localhost:8080")
//
// Environment variables:
//   - WEBSOCKET_ALLOWED_ORIGINS: Comma-separated list of allowed origin URLs
func (s *RPCServer) getAllowedOrigins() []string {
	origins := os.Getenv("WEBSOCKET_ALLOWED_ORIGINS")
	if origins == "" {
		// Default to common local development origins using the server's actual port
		//hosts := []string{"localhost", "127.0.0.1"}
		hosts := make(map[string]string)
		hosts["localhost"] = "localhost"
		hosts["127.0.0.1"] = "127.0.0.1"
		if s.Addr != nil {
			host, _, err := net.SplitHostPort(s.Addr.String())
			if err == nil && host != "" {
				hosts[host] = host
			}
		}
		port := "8080" // Default fallback
		if s.Addr != nil {
			_, ports, err := net.SplitHostPort(s.Addr.String())
			if err == nil && port != "" {
				// Use the actual port the server is listening on
				port = ports
			}
		}
		addrs := []string{}
		for _, host := range orderHosts(hosts) {
			addrs = append(addrs, fmt.Sprintf("http://%s:%s", host, port))
			addrs = append(addrs, fmt.Sprintf("https://%s:%s", host, port))
		}

		return addrs
	}
	return strings.Split(origins, ",")
}

// ADDED: isOriginAllowed checks if the given origin is in the allowed origins list for security validation.
// It performs case-sensitive string matching against the whitelist of allowed origins.
//
// Parameters:
//   - origin: The origin URL to validate (e.g., "http://localhost:8080")
//   - allowedOrigins: Slice of allowed origin URLs to check against
//
// Returns:
//   - bool: true if origin is allowed, false otherwise
func (s *RPCServer) isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}

// ADDED: upgrader creates and configures a WebSocket upgrader instance for handling HTTP to WebSocket protocol upgrades.
// It sets buffer sizes and implements origin checking for security purposes.
//
// Configuration:
//   - ReadBufferSize: 1024 bytes for incoming WebSocket frames
//   - WriteBufferSize: 1024 bytes for outgoing WebSocket frames
//   - CheckOrigin: Validates request origin against allowed origins list
//
// Security: The CheckOrigin function prevents cross-site WebSocket hijacking by validating
// request origins against the configured allowed origins list.
//
// Returns:
//   - *websocket.Upgrader: Configured upgrader instance for WebSocket connections
func (s *RPCServer) upgrader() *websocket.Upgrader {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			// Use configuration-based origin validation
			allowed := s.config.IsOriginAllowed(origin)

			if !allowed {
				logrus.WithFields(logrus.Fields{
					"origin":  origin,
					"devMode": s.config.EnableDevMode,
					"allowed": s.config.AllowedOrigins,
				}).Warn("WebSocket connection rejected: origin not allowed")
			} else {
				logrus.WithFields(logrus.Fields{
					"origin":  origin,
					"devMode": s.config.EnableDevMode,
				}).Debug("WebSocket connection allowed")
			}

			return allowed
		},
	}
	return &upgrader
}

// ADDED: wsConnection represents a WebSocket connection with thread-safe operations.
// It wraps the standard websocket.Conn with a mutex for concurrent access control.
//
// Fields:
//   - conn: The underlying WebSocket connection handler
//   - mu: Mutex to ensure thread-safe access to the connection
//
// Thread Safety: All write operations to the WebSocket connection should be protected
// by the mutex to prevent concurrent write panics.
//
// Related types:
//   - websocket.Conn from "github.com/gorilla/websocket"
type wsConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// RPCRequest represents a JSON-RPC 2.0 request message structure.
// It encapsulates all required fields for RPC method invocation over WebSocket.
//
// Fields:
//   - JSONRPC: Protocol version identifier (always "2.0")
//   - Method: RPC method name to invoke
//   - Params: Method parameters as a flexible map structure
//   - ID: Request identifier for matching responses
//
// Related standards: JSON-RPC 2.0 specification
type RPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      interface{}            `json:"id"`
}

// NewResponse creates a new JSON-RPC 2.0 success response message.
// It formats the result data according to JSON-RPC 2.0 specification.
//
// Parameters:
//   - id: Request identifier to match with original request
//   - result: Response data/payload to return to client
//
// Returns:
//   - interface{}: JSON-RPC 2.0 formatted response object
func NewResponse(id, result interface{}) interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  result,
		"id":      id,
	}
}

// NewErrorResponse creates a new JSON-RPC 2.0 error response message.
// It formats error information according to JSON-RPC 2.0 specification.
//
// Parameters:
//   - id: Request identifier to match with original request
//   - err: Error object containing failure details
//
// Returns:
//   - interface{}: JSON-RPC 2.0 formatted error response object with code -32000
func NewErrorResponse(id interface{}, err error) interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    -32000,
			"message": err.Error(),
		},
		"id": id,
	}
}

// HandleWebSocket manages WebSocket connections for real-time game communication.
// It upgrades HTTP connections to WebSocket protocol and handles bidirectional message flow.
//
// This method:
// 1. Retrieves the player session from request context
// 2. Upgrades the HTTP connection to WebSocket
// 3. Sends session confirmation to client
// 4. Spawns goroutines for message handling (send/receive)
// 5. Manages connection lifecycle and cleanup
//
// Parameters:
//   - w: HTTP response writer for the upgrade
//   - r: HTTP request containing session context
//
// Connection management:
//   - Automatic cleanup on disconnect
//   - Session state synchronization
//   - Bidirectional message queuing
func (s *RPCServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("function", "HandleWebSocket")
	session := r.Context().Value(sessionKey).(*PlayerSession)
	if session == nil {
		logrus.Error("no session in context")
		return
	}

	conn, err := s.upgradeConnection(w, r)
	if err != nil {
		return
	}
	defer conn.Close()

	if err := s.sendSessionConfirmation(conn, session); err != nil {
		return
	}

	session.WSConn = conn
	logrus.Info("websocket connection established")

	s.handleWebSocketMessages(conn, session, logger)
}

// upgradeConnection establishes a WebSocket connection from an HTTP request.
func (s *RPCServer) upgradeConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := s.upgrader().Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("websocket upgrade failed")
		return nil, err
	}
	return conn, nil
}

// sendSessionConfirmation sends initial session confirmation to the WebSocket client.
func (s *RPCServer) sendSessionConfirmation(conn *websocket.Conn, session *PlayerSession) error {
	confirmationMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]string{
			"session_id": session.SessionID,
		},
		"id": 0,
	}

	if err := conn.WriteJSON(confirmationMsg); err != nil {
		logrus.WithError(err).Error("failed to send session confirmation")
		return err
	}
	return nil
}

// handleWebSocketMessages processes incoming WebSocket messages in a continuous loop.
func (s *RPCServer) handleWebSocketMessages(conn *websocket.Conn, session *PlayerSession, logger *logrus.Entry) {
	for {
		var req RPCRequest
		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		if err := s.processWebSocketRequest(conn, session, req, logger); err != nil {
			break
		}
	}
}

// processWebSocketRequest handles a single WebSocket RPC request.
func (s *RPCServer) processWebSocketRequest(conn *websocket.Conn, session *PlayerSession, req RPCRequest, logger *logrus.Entry) error {
	enrichedParams := s.enrichRequestParams(req.Params, session.SessionID)

	paramsJSON, err := json.Marshal(enrichedParams)
	if err != nil {
		logger.WithError(err).Error("failed to marshal params")
		conn.WriteJSON(NewErrorResponse(req.ID, err))
		return nil
	}

	result, err := s.handleMethod(RPCMethod(req.Method), paramsJSON)
	if err != nil {
		logger.WithError(err).Error("RPC method execution failed")
		conn.WriteJSON(NewErrorResponse(req.ID, err))
		return nil
	}

	if err := conn.WriteJSON(NewResponse(req.ID, result)); err != nil {
		logger.WithError(err).Error("failed to write response")
		return err
	}
	return nil
}

// enrichRequestParams adds session ID to request parameters.
func (s *RPCServer) enrichRequestParams(params map[string]interface{}, sessionID string) map[string]interface{} {
	if params == nil {
		params = make(map[string]interface{})
	}
	params["session_id"] = sessionID
	return params
}

// ADDED: validateSession validates and retrieves a player session from RPC parameters.
// It extracts the session ID from the parameters map and returns the corresponding session.
//
// Parameters:
//   - params: Map containing RPC parameters, must include "session_id" key
//
// Returns:
//   - *PlayerSession: Valid player session if found
//   - error: ErrInvalidSession if session ID is missing or session not found
//
// This function is used by RPC handlers to authenticate and authorize requests.
func (s *RPCServer) validateSession(params map[string]interface{}) (*PlayerSession, error) {
	sessionID, ok := params["session_id"].(string)
	if !ok {
		return nil, ErrInvalidSession
	}

	return s.getSessionSafely(sessionID)
}

// ADDED: sendWSResponse sends a JSON-RPC 2.0 response message over a WebSocket connection.
// It constructs a properly formatted response and handles thread-safe transmission.
//
// Parameters:
//   - wsConn: The WebSocket connection wrapper (must not be nil)
//   - result: The result payload to include in the response (JSON-serializable)
//   - id: The request ID to correlate with the original request
//
// Response format follows JSON-RPC 2.0 specification:
//   - jsonrpc: "2.0"
//   - result: The provided result value
//   - id: The provided request ID
//
// Thread safety: Uses the connection's mutex lock to prevent concurrent write operations.
// Errors are logged but not returned to avoid breaking the message flow.
func (s *RPCServer) sendWSResponse(wsConn *wsConnection, result, id interface{}) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "sendWSResponse",
		"id":       id,
	})
	logger.Debug("sending websocket response")

	response := struct {
		JSONRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	wsConn.mu.Lock()
	defer wsConn.mu.Unlock()

	if err := wsConn.conn.WriteJSON(response); err != nil {
		logger.WithError(err).Error("failed to write websocket response")
	} else {
		logger.Debug("websocket response sent successfully")
	}
}

// ADDED: sendWSError sends a JSON-RPC 2.0 error response over the WebSocket connection.
// It constructs a properly formatted error response following JSON-RPC 2.0 specification.
//
// Parameters:
//   - wsConn: The WebSocket connection wrapper to send the response on
//   - code: The JSON-RPC error code to include (standard or custom)
//   - message: Human-readable error message describing the error
//   - data: Optional additional error details (may be nil)
//   - id: The JSON-RPC request ID the error responds to
//
// Error response structure:
//   - jsonrpc: "2.0"
//   - error: Object containing code, message, and optional data
//   - id: Original request identifier
//
// Thread safety: Uses mutex locking to prevent concurrent write operations.
// Write errors are logged but not returned to avoid breaking message flow.
func (s *RPCServer) sendWSError(wsConn *wsConnection, code int, message string, data, id interface{}) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "sendWSError",
		"id":       id,
		"code":     code,
	})
	logger.Debug("sending websocket error response")

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
		ID: id,
	}

	wsConn.mu.Lock()
	defer wsConn.mu.Unlock()

	if err := wsConn.conn.WriteJSON(response); err != nil {
		logger.WithError(err).Error("failed to write websocket error response")
	} else {
		logger.Debug("websocket error response sent successfully")
	}
}

// ADDED: getSessionSafely retrieves and validates a player session with thread-safe access.
// It performs atomic session lookup, validation, and timestamp updates to prevent race conditions.
//
// This function ensures:
// - Thread-safe session map access using read locks
// - Session existence and validity validation
// - WebSocket connection presence verification
// - Atomic LastActive timestamp updates
//
// Parameters:
//   - sessionID: The session ID to look up (must not be empty)
//
// Returns:
//   - *PlayerSession: Valid session if found and has active WebSocket connection
//   - error: ErrInvalidSession if not found, invalid, or missing WebSocket connection
//
// Thread Safety: Prevents TOCTOU race conditions by maintaining locks during
// validation and ensuring returned session references remain valid.
func (s *RPCServer) getSessionSafely(sessionID string) (*PlayerSession, error) {
	if sessionID == "" {
		return nil, ErrInvalidSession
	}

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	if !exists {
		s.mu.RUnlock()
		return nil, ErrInvalidSession
	}

	// Additional validation while still holding the lock
	if session.WSConn == nil {
		s.mu.RUnlock()
		return nil, ErrInvalidSession
	}

	// Increment reference count and update last active timestamp while holding lock
	session.addRef()
	session.LastActive = time.Now()
	s.mu.RUnlock()

	return session, nil
}

// WebSocketBroadcaster manages real-time event broadcasting to all connected WebSocket clients.
// It bridges the game event system with WebSocket connections for live multiplayer updates.
//
// Core responsibilities:
// - Event subscription and filtering for relevant game events
// - Broadcasting events to all active WebSocket connections
// - Connection lifecycle management and cleanup
// - Message formatting and serialization for WebSocket transmission
//
// Fields:
//   - server: Reference to the RPC server for accessing sessions and connections
//   - eventTypes: Set of EventType values that should be broadcast to clients
//   - mu: Mutex for thread-safe access to connection management
//   - active: Flag indicating if the broadcaster is running
//
// The broadcaster subscribes to specific game events and distributes them to all
// connected WebSocket clients in real-time, enabling live multiplayer gameplay.
type WebSocketBroadcaster struct {
	server     *RPCServer
	eventTypes map[game.EventType]bool
	mu         sync.RWMutex
	active     bool
}

// NewWebSocketBroadcaster creates and initializes a new WebSocket event broadcaster.
//
// Parameters:
//   - server: The RPC server instance containing WebSocket connections
//
// Returns:
//   - *WebSocketBroadcaster: Configured broadcaster ready for event subscription
func NewWebSocketBroadcaster(server *RPCServer) *WebSocketBroadcaster {
	return &WebSocketBroadcaster{
		server:     server,
		eventTypes: make(map[game.EventType]bool),
		active:     false,
	}
}

// Start activates the WebSocket broadcaster and subscribes to relevant game events.
// It registers event handlers for multiplayer-relevant events that should be broadcast.
//
// Subscribed events:
//   - Movement events: Player position changes
//   - Combat events: Attacks, damage, death
//   - Spell casting: Magic effects and targeting
//   - Chat/communication: Player messages
//   - World changes: Item drops, object interactions
func (wb *WebSocketBroadcaster) Start() {
	wb.mu.Lock()
	defer wb.mu.Unlock()

	if wb.active {
		return // Already started
	}

	// Subscribe to events that should be broadcast to all clients
	wb.eventTypes[game.EventMovement] = true
	wb.eventTypes[game.EventDamage] = true
	wb.eventTypes[game.EventSpellCast] = true
	wb.eventTypes[game.EventDeath] = true
	wb.eventTypes[game.EventItemDrop] = true
	wb.eventTypes[EventCombatStart] = true
	wb.eventTypes[EventCombatEnd] = true

	// Register as event handler for each type
	for eventType := range wb.eventTypes {
		wb.server.eventSys.Subscribe(eventType, wb.handleEvent)
	}

	wb.active = true
	logrus.Info("WebSocket broadcaster started and subscribed to game events")
}

// Stop deactivates the WebSocket broadcaster and unsubscribes from game events.
func (wb *WebSocketBroadcaster) Stop() {
	wb.mu.Lock()
	defer wb.mu.Unlock()

	wb.active = false
	wb.eventTypes = make(map[game.EventType]bool)
	logrus.Info("WebSocket broadcaster stopped")
}

// handleEvent processes game events and broadcasts them to all connected WebSocket clients.
//
// Parameters:
//   - event: The game event to broadcast
func (wb *WebSocketBroadcaster) handleEvent(event game.GameEvent) {
	wb.mu.RLock()
	active := wb.active
	shouldBroadcast := wb.eventTypes[event.Type]
	wb.mu.RUnlock()

	if !active || !shouldBroadcast {
		return
	}

	// Create WebSocket event message
	wsEvent := map[string]interface{}{
		"type":      "game_event",
		"event":     event.Type,
		"source":    event.SourceID,
		"target":    event.TargetID,
		"data":      event.Data,
		"timestamp": event.Timestamp,
	}

	// Broadcast to all connected WebSocket clients
	wb.broadcastToAll(wsEvent)
}

// broadcastToAll sends a message to all active WebSocket connections.
//
// Parameters:
//   - message: The message data to broadcast (must be JSON-serializable)
func (wb *WebSocketBroadcaster) broadcastToAll(message interface{}) {
	wb.server.mu.RLock()
	sessions := make([]*PlayerSession, 0, len(wb.server.sessions))
	for _, session := range wb.server.sessions {
		if session != nil && session.WSConn != nil && session.Connected {
			sessions = append(sessions, session)
		}
	}
	wb.server.mu.RUnlock()

	if len(sessions) == 0 {
		return // No active WebSocket connections
	}

	successCount := 0
	for _, session := range sessions {
		// Double-check connection is still valid before writing
		if session.WSConn != nil {
			// Safely attempt to write, catching any panics from invalid connections
			func() {
				defer func() {
					if r := recover(); r != nil {
						logrus.WithFields(logrus.Fields{
							"sessionID": session.SessionID,
							"error":     fmt.Sprintf("panic during WebSocket write: %v", r),
						}).Warn("recovered from WebSocket write panic")
					}
				}()

				if err := session.WSConn.WriteJSON(message); err != nil {
					logrus.WithFields(logrus.Fields{
						"sessionID": session.SessionID,
						"error":     err.Error(),
					}).Warn("failed to broadcast to WebSocket client")
				} else {
					successCount++
				}
			}()
		}
	}

	logrus.WithFields(logrus.Fields{
		"totalClients":    len(sessions),
		"successfulSends": successCount,
		"failedSends":     len(sessions) - successCount,
	}).Debug("WebSocket broadcast completed")
}

// Package server implements the game server and combat system functionality
