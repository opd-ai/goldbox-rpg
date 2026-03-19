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

// rpcErrConnectionLost is a custom JSON-RPC error code (server-defined range
// -32000 to -32099) used when the WebSocket connection drops while requests
// are still pending. Callers receive this instead of a 30-second timeout.
const rpcErrConnectionLost = -32003

// sessionIDMu guards concurrent access to RPCClient.sessionID.
var sessionIDMu sync.RWMutex

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

	// Reconnection state
	reconnecting atomic.Bool
	reconnectMax int
}

// NewRPCClient creates a new RPC client instance.
func NewRPCClient() *RPCClient {
	return &RPCClient{
		pending:      make(map[int64]*PendingRequest),
		reconnectMax: 3,
	}
}

// Connect establishes a WebSocket connection to the server.
func (c *RPCClient) Connect() error {
	wsURL := c.buildWebSocketURL()
	ws := js.Global().Get("WebSocket").New(wsURL)
	c.ws = ws

	connectDone := make(chan error, 1)
	var connectSent sync.Once

	c.setupWebSocketHandlers(ws, &connectSent, connectDone)

	return c.waitForConnection(connectDone)
}

// buildWebSocketURL constructs the WebSocket URL from the current page location.
func (c *RPCClient) buildWebSocketURL() string {
	location := js.Global().Get("location")
	protocol := "ws:"
	if location.Get("protocol").String() == "https:" {
		protocol = "wss:"
	}
	host := location.Get("host").String()
	return fmt.Sprintf("%s//%s/rpc/ws", protocol, host)
}

// setupWebSocketHandlers configures the WebSocket event callbacks.
func (c *RPCClient) setupWebSocketHandlers(ws js.Value, connectSent *sync.Once, connectDone chan error) {
	ws.Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c.connected.Store(true)
		if c.onConnected != nil {
			c.onConnected()
		}
		connectSent.Do(func() {
			connectDone <- nil
		})
		return nil
	}))

	ws.Set("onclose", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c.connected.Store(false)
		reason := "connection closed"
		if len(args) > 0 {
			reason = args[0].Get("reason").String()
		}
		connectSent.Do(func() {
			connectDone <- fmt.Errorf("connection closed: %s", reason)
		})
		if c.onDisconnect != nil {
			c.onDisconnect(reason)
		}
		// Automatic reconnection with exponential back-off per §1.
		go c.autoReconnect()
		return nil
	}))

	ws.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		connectSent.Do(func() {
			connectDone <- fmt.Errorf("WebSocket connection error")
		})
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
}

// waitForConnection blocks until the WebSocket connects or times out.
func (c *RPCClient) waitForConnection(connectDone chan error) error {
	select {
	case err := <-connectDone:
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("connection timeout")
	}
}

// Disconnect closes the WebSocket connection.
func (c *RPCClient) Disconnect() {
	c.reconnecting.Store(true) // prevent auto-reconnect on intentional disconnect
	if c.ws.Truthy() && c.connected.Load() {
		c.ws.Call("close", 1000, "client disconnect")
	}
	c.connected.Store(false)
}

// autoReconnect attempts to reconnect with exponential back-off (2s, 4s, 6s).
func (c *RPCClient) autoReconnect() {
	if c.reconnecting.Load() {
		return
	}
	c.reconnecting.Store(true)
	defer c.reconnecting.Store(false)

	// Fail all pending requests immediately so callers don't block for 30s.
	c.drainPendingRequests()

	// Clear stale session; the server will issue a new one on reconnect.
	sessionIDMu.Lock()
	c.sessionID = ""
	sessionIDMu.Unlock()

	backoffs := []time.Duration{2 * time.Second, 4 * time.Second, 6 * time.Second}
	for attempt, delay := range backoffs {
		if c.connected.Load() {
			return
		}
		time.Sleep(delay)
		if err := c.Connect(); err != nil {
			if c.onError != nil {
				c.onError(fmt.Errorf("reconnect attempt %d failed: %v", attempt+1, err))
			}
			continue
		}
		// Reconnected — session_id will be captured from the server's
		// confirmation message by handleJSONRPCMessage/captureSessionID.
		return
	}
}

// drainPendingRequests fails all outstanding requests with a connection-lost error
// so that blocked callers return immediately instead of waiting for a 30s timeout.
func (c *RPCClient) drainPendingRequests() {
	c.pendingMu.Lock()
	for id, pending := range c.pending {
		select {
		case pending.ResponseChan <- &RPCResponse{
			Error: &RPCError{Code: rpcErrConnectionLost, Message: "connection lost"},
		}:
		default:
		}
		delete(c.pending, id)
	}
	c.pendingMu.Unlock()
}

// IsConnected returns true if connected to the server.
func (c *RPCClient) IsConnected() bool {
	return c.connected.Load()
}

// GetSessionID returns the current session ID.
func (c *RPCClient) GetSessionID() string {
	sessionIDMu.RLock()
	defer sessionIDMu.RUnlock()
	return c.sessionID
}

// SetSessionID sets the session ID for subsequent requests.
func (c *RPCClient) SetSessionID(id string) {
	sessionIDMu.Lock()
	defer sessionIDMu.Unlock()
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
	sessionIDMu.RLock()
	sid := c.sessionID
	sessionIDMu.RUnlock()

	if sid != "" && params != nil {
		params["session_id"] = sid
	} else if sid != "" {
		params = map[string]interface{}{"session_id": sid}
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
// It detects JSON-RPC 2.0 messages vs plain JSON game events.
func (c *RPCClient) handleMessage(data string) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		c.reportError(fmt.Errorf("failed to parse incoming message: %w", err))
		return
	}

	if c.isJSONRPCMessage(raw) {
		c.handleJSONRPCMessage(data)
		return
	}

	// Non-JSON-RPC payload (e.g., game events broadcast as plain JSON).
	if c.onMessage != nil {
		c.onMessage(raw)
	}
}

// isJSONRPCMessage checks if the message is a JSON-RPC 2.0 message.
func (c *RPCClient) isJSONRPCMessage(raw map[string]interface{}) bool {
	v, ok := raw["jsonrpc"].(string)
	return ok && v == "2.0"
}

// handleJSONRPCMessage processes a JSON-RPC 2.0 response or notification.
func (c *RPCClient) handleJSONRPCMessage(data string) {
	var resp RPCResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		c.reportError(fmt.Errorf("failed to parse JSON-RPC response: %w", err))
		return
	}

	if c.dispatchPendingResponse(&resp) {
		return
	}

	// Capture session_id from the server's initial confirmation message
	// (sent with id:0 on connect, before any client requests).
	c.captureSessionID(&resp)

	// Treat as server notification (no ID field)
	if c.onMessage != nil && resp.Result != nil {
		c.onMessage(resp.Result)
	}
}

// captureSessionID extracts and stores a session_id from a JSON-RPC response
// that was not matched to any pending request (e.g. the server's initial
// session confirmation message).
func (c *RPCClient) captureSessionID(resp *RPCResponse) {
	if resp.Result == nil {
		return
	}
	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		return
	}
	if sid, ok := resultMap["session_id"].(string); ok && sid != "" {
		sessionIDMu.Lock()
		c.sessionID = sid
		sessionIDMu.Unlock()
	}
}

// dispatchPendingResponse routes a response to its waiting request, returns true if dispatched.
func (c *RPCClient) dispatchPendingResponse(resp *RPCResponse) bool {
	if resp.ID == nil {
		return false
	}

	c.pendingMu.RLock()
	pending, ok := c.pending[*resp.ID]
	c.pendingMu.RUnlock()

	if ok {
		// Non-blocking send: if the caller already timed out or the channel
		// was drained during reconnect, drop the response instead of blocking.
		select {
		case pending.ResponseChan <- resp:
		default:
		}
		return true
	}
	return false
}

// reportError sends an error to the error callback if set.
func (c *RPCClient) reportError(err error) {
	if c.onError != nil {
		c.onError(err)
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
		sessionIDMu.Lock()
		c.sessionID = joinResult.SessionID
		sessionIDMu.Unlock()
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

// GetCombatModifiers retrieves cover and flanking information for a target.
func (c *RPCClient) GetCombatModifiers(targetID string) (*CombatModifiers, error) {
	result, err := c.Call("getCombatModifiers", map[string]interface{}{
		"target_id": targetID,
	})
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var modifiers CombatModifiers
	if err := json.Unmarshal(data, &modifiers); err != nil {
		return nil, err
	}

	return &modifiers, nil
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
	NewPosition *Position `json:"position,omitempty"`
	Message     string    `json:"message,omitempty"`
}

// AttackResult represents the result of an attack call.
type AttackResult struct {
	Success      bool   `json:"success"`
	Hit          bool   `json:"hit"`
	Damage       int    `json:"damage,omitempty"`
	TargetHealth int    `json:"target_health,omitempty"`
	Message      string `json:"message"`
	// Extended fields for rich combat narration (Gold Box style)
	AttackRoll   int    `json:"attack_roll,omitempty"`   // The d20 attack roll result
	TargetAC     int    `json:"target_ac,omitempty"`     // Target's armor class
	IsCritical   bool   `json:"is_critical,omitempty"`   // True if critical hit
	AttackerName string `json:"attacker_name,omitempty"` // Name of attacker for narration
	TargetName   string `json:"target_name,omitempty"`   // Name of target for narration
	WeaponName   string `json:"weapon_name,omitempty"`   // Name of weapon used
}

// GameStateResult represents the result of a getGameState call.
// Fields match the server's GetState response structure.
type GameStateResult struct {
	// Fields directly populated from the server's getGameState response.
	World    interface{}            `json:"world"`
	Time     interface{}            `json:"time"`
	Turns    interface{}            `json:"turns"`
	Sessions map[string]interface{} `json:"sessions"`
	Version  int64                  `json:"version"`

	// Top-level player data added by handleGetGameState for the requesting session.
	Player map[string]interface{} `json:"player,omitempty"`
}

// EndTurnResult represents the result of an endTurn call.
type EndTurnResult struct {
	Success  bool   `json:"success"`
	NextTurn string `json:"next_turn"`
}
