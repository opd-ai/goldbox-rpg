package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// HandleWebSocket upgrades an HTTP connection to WebSocket and handles the connection
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
