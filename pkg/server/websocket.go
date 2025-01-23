package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	wsConn := &wsConnection{conn: conn}

	// Main message loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
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
			s.sendWSError(wsConn, -32700, "Parse error", nil, req.ID)
			continue
		}

		// Handle the RPC method
		result, err := s.handleMethod(req.Method, req.Params)
		if err != nil {
			s.sendWSError(wsConn, -32603, err.Error(), nil, req.ID)
			continue
		}

		// Send successful response
		s.sendWSResponse(wsConn, result, req.ID)
	}
}

func (s *RPCServer) sendWSResponse(wsConn *wsConnection, result, id interface{}) {
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
		log.Printf("WebSocket write error: %v", err)
	}
}

func (s *RPCServer) sendWSError(wsConn *wsConnection, code int, message string, data interface{}, id interface{}) {
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
		log.Printf("WebSocket write error: %v", err)
	}
}
