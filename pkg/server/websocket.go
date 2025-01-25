package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

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
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

// HandleWebSocket handles incoming WebSocket connections and implements a JSON-RPC protocol.
//
// It upgrades the HTTP connection to WebSocket, maintains the connection, and processes
// incoming JSON-RPC messages in a loop until the connection is closed.
//
// Parameters:
//   - w http.ResponseWriter: The HTTP response writer
//   - r *http.Request: The incoming HTTP request to upgrade
//
// Notable behaviors:
//   - Implements JSON-RPC 2.0 protocol over WebSocket
//   - Handles WebSocket connection upgrades
//   - Processes incoming messages in an infinite loop until connection closes
//   - Parses JSON-RPC requests with method, params and ID
//   - Forwards requests to handleMethod for processing
//   - Sends back JSON-RPC responses/errors
//
// Error handling:
//   - Returns immediately if WebSocket upgrade fails
//   - Handles unexpected WebSocket close errors
//   - Sends JSON-RPC error responses for parse errors (-32700)
//   - Sends JSON-RPC error responses for internal errors (-32603)
//
// Related:
//   - handleMethod() - Processes individual RPC method calls
//   - sendWSError() - Sends error responses
//   - sendWSResponse() - Sends success responses
//   - wsConnection - WebSocket connection wrapper
func (s *RPCServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "HandleWebSocket",
	})
	logger.Debug("entering websocket handler")
	///

	session := r.Context().Value("session").(*PlayerSession)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("websocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Associate WebSocket with session
	session.WSConn = conn
	session.LastActive = time.Now()

	// Send session confirmation
	if err := conn.WriteJSON(map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]string{
			"session_id": session.SessionID,
		},
		"id": 0,
	}); err != nil {
		logrus.Error("failed to send session confirmation:", err)
		return
	}

	//
	wsConn := &wsConnection{conn: conn}
	logger.Info("websocket connection established")

	// Main message loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.WithError(err).Error("unexpected websocket closure")
			} else {
				logger.WithError(err).Info("websocket connection closed")
			}
			break
		}

		// Parse the RPC request
		var req struct {
			JsonRPC string          `json:"jsonrpc"`
			Method  RPCMethod       `json:"method"`
			Params  json.RawMessage `json:"params"`
			ID      interface{}     `json:"id"`
		}

		if err := json.Unmarshal(message, &req); err != nil {
			logger.WithError(err).Warn("failed to parse JSON-RPC request")
			s.sendWSError(wsConn, -32700, "Parse error", nil, req.ID)
			continue
		}

		logger.WithFields(logrus.Fields{
			"method": req.Method,
			"id":     req.ID,
		}).Debug("processing RPC request")

		// Handle the RPC method
		result, err := s.handleMethod(req.Method, req.Params)
		if err != nil {
			logger.WithError(err).Error("RPC method execution failed")
			s.sendWSError(wsConn, -32603, err.Error(), nil, req.ID)
			continue
		}

		logger.WithFields(logrus.Fields{
			"method": req.Method,
			"id":     req.ID,
		}).Debug("sending RPC response")

		// Send successful response
		s.sendWSResponse(wsConn, result, req.ID)
	}

	logger.Debug("exiting websocket handler")
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
func (s *RPCServer) sendWSError(wsConn *wsConnection, code int, message string, data interface{}, id interface{}) {
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
