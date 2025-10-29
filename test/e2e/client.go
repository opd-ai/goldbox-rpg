package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Client is an E2E test client for the GoldBox RPG server
// It provides methods for JSON-RPC calls and WebSocket communication
type Client struct {
	baseURL    string
	httpClient *http.Client
	wsConn     *websocket.Conn
	wsMessages chan map[string]interface{}
	wsErrors   chan error
	wsCloseCh  chan struct{}
	wsMutex    sync.Mutex
	idCounter  int
	log        *logrus.Logger
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int         `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
	ID      int                    `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{}                  `json:"data,omitempty"`
}

// NewClient creates a new E2E test client
func NewClient(baseURL string) *Client {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		wsMessages: make(chan map[string]interface{}, 100),
		wsErrors:   make(chan error, 10),
		wsCloseCh:  make(chan struct{}),
		log:        logger,
	}
}

// Call makes a JSON-RPC call to the server
func (c *Client) Call(method string, params interface{}) (map[string]interface{}, error) {
	c.idCounter++
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.idCounter,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.log.Debugf("Calling %s with params: %v", method, params)

	resp, err := c.httpClient.Post(
		c.baseURL+"/rpc",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response JSONRPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// ConnectWebSocket connects to the WebSocket endpoint
func (c *Client) ConnectWebSocket() error {
	c.wsMutex.Lock()
	defer c.wsMutex.Unlock()

	if c.wsConn != nil {
		return fmt.Errorf("WebSocket already connected")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	// Convert http/https to ws/wss
	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}
	wsURL := fmt.Sprintf("%s://%s/ws", wsScheme, u.Host)

	c.log.Debugf("Connecting to WebSocket: %s", wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.wsConn = conn

	// Start message reader goroutine
	go c.readWebSocketMessages()

	return nil
}

// readWebSocketMessages reads messages from the WebSocket connection
func (c *Client) readWebSocketMessages() {
	defer func() {
		close(c.wsMessages)
		close(c.wsErrors)
	}()

	for {
		select {
		case <-c.wsCloseCh:
			return
		default:
			var msg map[string]interface{}
			if err := c.wsConn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					c.wsErrors <- fmt.Errorf("WebSocket read error: %w", err)
				}
				return
			}
			c.wsMessages <- msg
		}
	}
}

// WaitForEvent waits for a WebSocket event with the given type
func (c *Client) WaitForEvent(eventType string, timeout time.Duration) (map[string]interface{}, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case msg := <-c.wsMessages:
			if msg["type"] == eventType {
				return msg, nil
			}
			// Put the message back if it's not the one we're looking for
			// This is simplified - in production would need a better queue
		case err := <-c.wsErrors:
			return nil, err
		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for event %s", eventType)
		}
	}
}

// GetNextEvent returns the next WebSocket event
func (c *Client) GetNextEvent(timeout time.Duration) (map[string]interface{}, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case msg := <-c.wsMessages:
		return msg, nil
	case err := <-c.wsErrors:
		return nil, err
	case <-timer.C:
		return nil, fmt.Errorf("timeout waiting for event")
	}
}

// CloseWebSocket closes the WebSocket connection
func (c *Client) CloseWebSocket() error {
	c.wsMutex.Lock()
	defer c.wsMutex.Unlock()

	if c.wsConn == nil {
		return nil
	}

	close(c.wsCloseCh)
	
	err := c.wsConn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		c.log.Warnf("Failed to send close message: %v", err)
	}

	if err := c.wsConn.Close(); err != nil {
		return fmt.Errorf("failed to close WebSocket: %w", err)
	}

	c.wsConn = nil
	return nil
}

// Close closes all connections
func (c *Client) Close() error {
	if c.wsConn != nil {
		return c.CloseWebSocket()
	}
	return nil
}

// WaitForHealth waits for the server to be healthy
func (c *Client) WaitForHealth(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		resp, err := c.httpClient.Get(c.baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	return fmt.Errorf("server did not become healthy within %v", timeout)
}

// Helper methods for common RPC calls

// JoinGame joins a game session
func (c *Client) JoinGame(playerName string) (string, error) {
	params := map[string]interface{}{
		"player_name": playerName,
	}
	
	result, err := c.Call("join_game", params)
	if err != nil {
		return "", err
	}
	
	sessionID, ok := result["session_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid session_id in response")
	}
	
	return sessionID, nil
}

// CreateCharacter creates a new character
func (c *Client) CreateCharacter(sessionID, name string, class string) (string, error) {
	params := map[string]interface{}{
		"session_id": sessionID,
		"name":       name,
		"class":      class,
		"method":     "standard_array",
	}
	
	result, err := c.Call("create_character", params)
	if err != nil {
		return "", err
	}
	
	charID, ok := result["character_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid character_id in response")
	}
	
	return charID, nil
}

// Move moves the character in a direction
func (c *Client) Move(sessionID string, direction int) error {
	params := map[string]interface{}{
		"session_id": sessionID,
		"direction":  direction,
	}
	
	_, err := c.Call("move", params)
	return err
}

// GetGameState retrieves the current game state
func (c *Client) GetGameState(sessionID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"session_id": sessionID,
	}
	
	return c.Call("get_game_state", params)
}
