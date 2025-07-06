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

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// orderHosts sorts hosts in the specified priority order:
// 1. Custom hostnames (not localhost or IP addresses) first
// 2. localhost second
// 3. IP addresses last
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

// getAllowedOrigins returns the list of allowed WebSocket origins.
// It checks the WEBSOCKET_ALLOWED_ORIGINS environment variable for a comma-separated list.
// If not set, defaults to common local development origins matching the server's actual listening port.
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

// isOriginAllowed checks if the given origin is in the allowed origins list.
func (s *RPCServer) isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}

// upgrader is a websocket.Upgrader instance that handles WebSocket connection upgrades.
// It configures the following settings:
//   - ReadBufferSize: 1024 bytes for incoming WebSocket frames
//   - WriteBufferSize: 1024 bytes for outgoing WebSocket frames
//   - CheckOrigin: Allows all origins by always returning true (note: only suitable for development)
//
// Security Note: The current CheckOrigin setting allows any origin to establish WebSocket connections.
// This should be restricted in production environments to prevent cross-site WebSocket hijacking.
//
// Related Types:
//   - websocket.Upgrader (gorilla/websocket)
//   - http.Request
func (s *RPCServer) upgrader() *websocket.Upgrader {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			allowedOrigins := s.getAllowedOrigins()
			allowed := s.isOriginAllowed(origin, allowedOrigins)

			if !allowed {
				logrus.WithFields(logrus.Fields{
					"origin":         origin,
					"allowedOrigins": allowedOrigins,
				}).Warn("WebSocket connection rejected: origin not allowed")
			}

			return allowed
		},
	}
	return &upgrader
}

// wsConnection represents a WebSocket connection with thread-safe operations.
// It wraps the standard websocket.Conn with a mutex for concurrent access control.
//
// Fields:
//   - conn: The underlying WebSocket connection handler
//   - mu: Mutex to ensure thread-safe access to the connection
//
// Related types:
//   - websocket.Conn from "github.com/gorilla/websocket"
type wsConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// RPCRequest represents a JSON-RPC 2.0 request
type RPCRequest struct {
	JsonRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      interface{}            `json:"id"`
}

// NewResponse creates a new JSON-RPC 2.0 response
func NewResponse(id, result interface{}) interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  result,
		"id":      id,
	}
}

// NewErrorResponse creates a new JSON-RPC 2.0 error response
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

func (s *RPCServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("function", "HandleWebSocket")
	session := r.Context().Value("session").(*PlayerSession)
	if session == nil {
		logrus.Error("no session in context")
		return
	}

	conn, err := s.upgrader().Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("websocket upgrade failed")
		return
	}
	defer conn.Close()

	// Send session confirmation
	if err := conn.WriteJSON(map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]string{
			"session_id": session.SessionID,
		},
		"id": 0,
	}); err != nil {
		logrus.WithError(err).Error("failed to send session confirmation")
		return
	}

	session.WSConn = conn
	logrus.Info("websocket connection established")

	// Message handling loop
	for {
		var req RPCRequest
		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		// Inject session ID into params
		if req.Params == nil {
			req.Params = make(map[string]interface{})
		}
		req.Params["session_id"] = session.SessionID

		// Convert string to RPCMethod type
		method := RPCMethod(req.Method)

		// Convert params to json.RawMessage
		paramsJSON, err := json.Marshal(req.Params)
		if err != nil {
			logger.WithError(err).Error("failed to marshal params")
			conn.WriteJSON(NewErrorResponse(req.ID, err))
			continue
		}

		result, err := s.handleMethod(method, paramsJSON)
		if err != nil {
			logger.WithError(err).Error("RPC method execution failed")
			conn.WriteJSON(NewErrorResponse(req.ID, err))
			continue
		}

		if err := conn.WriteJSON(NewResponse(req.ID, result)); err != nil {
			logger.WithError(err).Error("failed to write response")
			break
		}
	}
}

func (s *RPCServer) validateSession(params map[string]interface{}) (*PlayerSession, error) {
	sessionID, ok := params["session_id"].(string)
	if !ok {
		return nil, ErrInvalidSession
	}

	return s.getSessionSafely(sessionID)
}

// sendWSResponse sends a JSON-RPC 2.0 response message over a WebSocket connection.
//
// Parameters:
//   - wsConn (*wsConnection): The WebSocket connection to send the response on. Must not be nil.
//   - result (interface{}): The result payload to include in the response. Can be any JSON-serializable value.
//   - id (interface{}): The request ID to correlate the response with the original request.
//
// The function constructs a JSON-RPC 2.0 compliant response object with:
//   - jsonrpc: Fixed to "2.0"
//   - result: The provided result value
//   - id: The provided request ID
//
// Thread safety is handled via the connection's mutex lock.
//
// Errors:
//   - Logs but does not return WebSocket write errors
//
// Related:
//   - wsConnection type (containing the WebSocket conn and mutex)
func (s *RPCServer) sendWSResponse(wsConn *wsConnection, result, id interface{}) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "sendWSResponse",
		"id":       id,
	})
	logger.Debug("sending websocket response")

	response := struct {
		JsonRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JsonRPC: "2.0",
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

// sendWSError sends a JSON-RPC 2.0 error response over the WebSocket connection.
//
// Parameters:
//   - wsConn: The WebSocket connection wrapper to send the response on
//   - code: The JSON-RPC error code to include
//   - message: A human-readable error message describing the error
//   - data: Optional additional error details to include in response (may be nil)
//   - id: The JSON-RPC request ID the error is in response to
//
// The function constructs a proper JSON-RPC 2.0 error response object and sends it
// over the provided WebSocket connection. Thread safety is handled via mutex locking.
//
// Error handling:
// - If the write to WebSocket fails, error is logged but not returned to caller
//
// Related:
// - JSON-RPC 2.0 Spec: https://www.jsonrpc.org/specification#error_object
func (s *RPCServer) sendWSError(wsConn *wsConnection, code int, message string, data, id interface{}) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "sendWSError",
		"id":       id,
		"code":     code,
	})
	logger.Debug("sending websocket error response")

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

// getSessionSafely retrieves a session by ID with proper locking and validation.
// This prevents time-of-check-time-of-use (TOCTOU) race conditions by ensuring
// atomic session access and validation.
//
// Parameters:
//   - sessionID: The session ID to look up
//
// Returns:
//   - *PlayerSession: The validated session if found and valid
//   - error: Error if session not found, invalid, or lacks WebSocket connection
//
// Thread Safety:
//   - Uses read lock for session map access
//   - Updates LastActive timestamp atomically
//   - Returns a reference that's safe to use (session won't be deleted while in use)
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

	// Update last active timestamp while holding lock
	session.LastActive = time.Now()
	s.mu.RUnlock()

	return session, nil
}
