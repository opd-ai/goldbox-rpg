//go:build js && wasm

// Package wasmui provides the JSON-RPC client for WASM UI communication with the server.
package wasmui

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"
)

// RPCRequest represents a JSON-RPC 2.0 request.
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int64       `json:"id"`
}

// RPCResponse represents a JSON-RPC 2.0 response.
// ID is a pointer to handle both null and absent ID (notifications).
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      *int64      `json:"id"`
}

// RPCError represents a JSON-RPC 2.0 error.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// PendingRequest tracks a request awaiting response.
type PendingRequest struct {
	ResponseChan chan *RPCResponse
	Timestamp    time.Time
}

// RPCClient handles WebSocket communication with the JSON-RPC server.
type RPCClient struct {
	ws           js.Value
	connected    atomic.Bool
	sessionID    string
	requestID    atomic.Int64
	pendingMu    sync.RWMutex
	pending      map[int64]*PendingRequest
	onConnected  func()
	onDisconnect func(reason string)
	onError      func(err error)
	onMessage    func(data interface{})
}

// NewRPCClient creates a new RPC client instance.
func NewRPCClient() *RPCClient {
	return &RPCClient{
		pending: make(map[int64]*PendingRequest),
	}
}

// Connect establishes a WebSocket connection to the server.
func (c *RPCClient) Connect() error {
	// Get the current location to build WebSocket URL
	location := js.Global().Get("location")
	protocol := "ws:"
	if location.Get("protocol").String() == "https:" {
		protocol = "wss:"
	}
	host := location.Get("host").String()
	wsURL := fmt.Sprintf("%s//%s/rpc/ws", protocol, host)

	// Create WebSocket connection
	ws := js.Global().Get("WebSocket").New(wsURL)
	c.ws = ws

	// Set up event handlers
	connectDone := make(chan error, 1)

	ws.Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c.connected.Store(true)
		if c.onConnected != nil {
			c.onConnected()
		}
		connectDone <- nil
		return nil
	}))

	ws.Set("onclose", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c.connected.Store(false)
		reason := "connection closed"
		if len(args) > 0 {
			reason = args[0].Get("reason").String()
		}
		if c.onDisconnect != nil {
			c.onDisconnect(reason)
		}
		return nil
	}))

	ws.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if c.onError != nil {
			c.onError(fmt.Errorf("WebSocket error"))
		}
		return nil
	}))

	ws.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			data := args[0].Get("data").String()
			go c.handleMessage(data)
		}
		return nil
	}))

	// Wait for connection with timeout
	select {
	case err := <-connectDone:
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("connection timeout")
	}
}

// Disconnect closes the WebSocket connection.
func (c *RPCClient) Disconnect() {
	if c.ws.Truthy() && c.connected.Load() {
		c.ws.Call("close", 1000, "client disconnect")
	}
	c.connected.Store(false)
}

// IsConnected returns true if connected to the server.
func (c *RPCClient) IsConnected() bool {
	return c.connected.Load()
}

// GetSessionID returns the current session ID.
func (c *RPCClient) GetSessionID() string {
	return c.sessionID
}

// SetSessionID sets the session ID for subsequent requests.
func (c *RPCClient) SetSessionID(id string) {
	c.sessionID = id
}

// SetOnConnected sets the callback for successful connection.
func (c *RPCClient) SetOnConnected(fn func()) {
	c.onConnected = fn
}

// SetOnDisconnect sets the callback for disconnection.
func (c *RPCClient) SetOnDisconnect(fn func(reason string)) {
	c.onDisconnect = fn
}

// SetOnError sets the callback for errors.
func (c *RPCClient) SetOnError(fn func(err error)) {
	c.onError = fn
}

// SetOnMessage sets the callback for server notifications.
func (c *RPCClient) SetOnMessage(fn func(data interface{})) {
	c.onMessage = fn
}

// Call sends an RPC request and waits for the response.
func (c *RPCClient) Call(method string, params map[string]interface{}) (interface{}, error) {
	if !c.connected.Load() {
		return nil, fmt.Errorf("not connected")
	}

	// Add session ID to params if available
	if c.sessionID != "" && params != nil {
		params["session_id"] = c.sessionID
	} else if c.sessionID != "" {
		params = map[string]interface{}{"session_id": c.sessionID}
	}

	id := c.requestID.Add(1)
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create pending request
	pending := &PendingRequest{
		ResponseChan: make(chan *RPCResponse, 1),
		Timestamp:    time.Now(),
	}

	c.pendingMu.Lock()
	c.pending[id] = pending
	c.pendingMu.Unlock()

	// Send request
	c.ws.Call("send", string(data))

	// Wait for response with timeout
	select {
	case resp := <-pending.ResponseChan:
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()

		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Result, nil
	case <-time.After(30 * time.Second):
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("request timeout")
	}
}

// handleMessage processes incoming WebSocket messages.
func (c *RPCClient) handleMessage(data string) {
	var resp RPCResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		if c.onError != nil {
			c.onError(fmt.Errorf("failed to parse response: %w", err))
		}
		return
	}

	// Check if this is a response to a pending request
	// Per JSON-RPC 2.0, responses have an ID, notifications don't have an ID field
	if resp.ID != nil {
		c.pendingMu.RLock()
		pending, ok := c.pending[*resp.ID]
		c.pendingMu.RUnlock()

		if ok {
			pending.ResponseChan <- &resp
			return
		}
	}

	// Otherwise, treat as server notification (no ID field)
	if c.onMessage != nil && resp.Result != nil {
		c.onMessage(resp.Result)
	}
}

// JoinGame sends a joinGame request to the server.
func (c *RPCClient) JoinGame(playerName string) (*JoinGameResult, error) {
	result, err := c.Call("joinGame", map[string]interface{}{
		"player_name": playerName,
	})
	if err != nil {
		return nil, err
	}

	// Parse result
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var joinResult JoinGameResult
	if err := json.Unmarshal(data, &joinResult); err != nil {
		return nil, err
	}

	// Store session ID
	if joinResult.Success {
		c.sessionID = joinResult.SessionID
	}

	return &joinResult, nil
}

// Move sends a move request to the server.
func (c *RPCClient) Move(direction string) (*MoveResult, error) {
	result, err := c.Call("move", map[string]interface{}{
		"direction": direction,
	})
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var moveResult MoveResult
	if err := json.Unmarshal(data, &moveResult); err != nil {
		return nil, err
	}

	return &moveResult, nil
}

// Attack sends an attack request to the server.
func (c *RPCClient) Attack(targetID, weaponID string) (*AttackResult, error) {
	result, err := c.Call("attack", map[string]interface{}{
		"target_id": targetID,
		"weapon_id": weaponID,
	})
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var attackResult AttackResult
	if err := json.Unmarshal(data, &attackResult); err != nil {
		return nil, err
	}

	return &attackResult, nil
}

// GetGameState retrieves the current game state from the server.
func (c *RPCClient) GetGameState() (*GameStateResult, error) {
	result, err := c.Call("getGameState", nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var stateResult GameStateResult
	if err := json.Unmarshal(data, &stateResult); err != nil {
		return nil, err
	}

	return &stateResult, nil
}

// EndTurn sends an end turn request to the server.
func (c *RPCClient) EndTurn() (*EndTurnResult, error) {
	result, err := c.Call("endTurn", nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var endResult EndTurnResult
	if err := json.Unmarshal(data, &endResult); err != nil {
		return nil, err
	}

	return &endResult, nil
}

// RPC Result types

// JoinGameResult represents the result of a joinGame call.
type JoinGameResult struct {
	SessionID string `json:"session_id"`
	PlayerID  string `json:"player_id"`
	Success   bool   `json:"success"`
}

// MoveResult represents the result of a move call.
type MoveResult struct {
	Success     bool      `json:"success"`
	NewPosition *Position `json:"new_position,omitempty"`
	Message     string    `json:"message,omitempty"`
}

// AttackResult represents the result of an attack call.
type AttackResult struct {
	Success      bool   `json:"success"`
	Damage       int    `json:"damage,omitempty"`
	TargetHealth int    `json:"target_health,omitempty"`
	Message      string `json:"message"`
}

// GameStateResult represents the result of a getGameState call.
type GameStateResult struct {
	Player    interface{} `json:"player"`
	World     interface{} `json:"world"`
	Combat    interface{} `json:"combat"`
	Timestamp int64       `json:"timestamp"`
}

// EndTurnResult represents the result of an endTurn call.
type EndTurnResult struct {
	Success  bool   `json:"success"`
	NextTurn string `json:"next_turn"`
}
