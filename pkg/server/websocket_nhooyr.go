package server

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/coder/websocket"
)

// nhooyrWebSocketConn wraps a nhooyr.io/websocket connection to implement
// the WebSocketConn interface. This is the default WebSocket implementation
// for the server, providing modern, actively-maintained WebSocket support.
//
// Thread Safety: nhooyr.io/websocket is designed for concurrent use and
// supports context-based cancellation natively.
type nhooyrWebSocketConn struct {
	conn *websocket.Conn
}

// NewNhooyrWebSocketConn creates a new WebSocketConn adapter from a nhooyr.io websocket connection.
//
// Parameters:
//   - conn: The underlying nhooyr websocket connection
//
// Returns:
//   - WebSocketConn: Adapter implementing the common WebSocket interface
func NewNhooyrWebSocketConn(conn *websocket.Conn) WebSocketConn {
	return &nhooyrWebSocketConn{conn: conn}
}

// ReadMessage reads a message from the WebSocket connection.
// Uses nhooyr's native context support for cancellation.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//
// Returns:
//   - messageType: Type of message (1 = TextMessage, 2 = BinaryMessage)
//   - p: Message payload bytes
//   - err: Error if read failed or connection closed
func (n *nhooyrWebSocketConn) ReadMessage(ctx context.Context) (messageType int, p []byte, err error) {
	msgType, data, err := n.conn.Read(ctx)
	if err != nil {
		return 0, nil, err
	}
	// Convert nhooyr message type to common type
	return int(msgType), data, nil
}

// WriteMessage writes a message to the WebSocket connection.
// Uses nhooyr's native context support for cancellation.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//   - messageType: Type of message (1 = TextMessage, 2 = BinaryMessage)
//   - data: Message payload bytes to send
//
// Returns:
//   - err: Error if write failed or connection closed
func (n *nhooyrWebSocketConn) WriteMessage(ctx context.Context, messageType int, data []byte) error {
	return n.conn.Write(ctx, websocket.MessageType(messageType), data)
}

// WriteJSON marshals v to JSON and writes it as a text message.
// Uses a write timeout to prevent indefinite blocking on stalled connections.
//
// Parameters:
//   - v: Value to marshal to JSON and send
//
// Returns:
//   - err: Error if marshal or write failed
func (n *nhooyrWebSocketConn) WriteJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return n.conn.Write(ctx, websocket.MessageText, data)
}

// ReadJSON reads a JSON message and unmarshals it into v.
//
// Parameters:
//   - v: Pointer to value to unmarshal into
//
// Returns:
//   - err: Error if read or unmarshal failed
func (n *nhooyrWebSocketConn) ReadJSON(v interface{}) error {
	_, data, err := n.conn.Read(context.Background())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Close closes the WebSocket connection with a status code and reason.
// Uses nhooyr's CloseNow for immediate closure after sending close frame.
//
// Parameters:
//   - code: WebSocket close status code (e.g., 1000 for normal closure)
//   - reason: Human-readable reason for closing
//
// Returns:
//   - err: Error if close failed
func (n *nhooyrWebSocketConn) Close(code int, reason string) error {
	return n.conn.Close(websocket.StatusCode(code), reason)
}

// CloseNow closes the WebSocket connection with a normal closure status.
// This is a convenience method for simple connection cleanup.
//
// Returns:
//   - err: Error if close failed
func (n *nhooyrWebSocketConn) CloseNow() error {
	return n.Close(CloseNormalClosure, "")
}

// RemoteAddr returns the remote network address of the WebSocket connection.
// Note: nhooyr.io/websocket doesn't directly expose RemoteAddr, so we
// need to track it separately during connection upgrade.
//
// Returns:
//   - net.Addr: Remote address of the connected client (may be nil)
func (n *nhooyrWebSocketConn) RemoteAddr() net.Addr {
	// nhooyr.io/websocket doesn't directly expose RemoteAddr
	// This needs to be captured during the HTTP upgrade
	return nil
}

// GetUnderlyingConn returns the underlying nhooyr websocket connection.
// This can be used when direct access to nhooyr-specific features is needed.
//
// Returns:
//   - *websocket.Conn: The underlying nhooyr websocket connection
func (n *nhooyrWebSocketConn) GetUnderlyingConn() *websocket.Conn {
	return n.conn
}

// nhooyrWebSocketConnWithAddr wraps nhooyrWebSocketConn with address tracking.
// This is needed because nhooyr.io/websocket doesn't expose RemoteAddr directly.
type nhooyrWebSocketConnWithAddr struct {
	nhooyrWebSocketConn
	remoteAddr net.Addr
}

// NewNhooyrWebSocketConnWithAddr creates a new WebSocketConn adapter with address tracking.
//
// Parameters:
//   - conn: The underlying nhooyr websocket connection
//   - remoteAddr: The remote address captured during HTTP upgrade
//
// Returns:
//   - WebSocketConn: Adapter implementing the common WebSocket interface with address
func NewNhooyrWebSocketConnWithAddr(conn *websocket.Conn, remoteAddr net.Addr) WebSocketConn {
	return &nhooyrWebSocketConnWithAddr{
		nhooyrWebSocketConn: nhooyrWebSocketConn{conn: conn},
		remoteAddr:          remoteAddr,
	}
}

// RemoteAddr returns the remote network address captured during HTTP upgrade.
//
// Returns:
//   - net.Addr: Remote address of the connected client
func (n *nhooyrWebSocketConnWithAddr) RemoteAddr() net.Addr {
	return n.remoteAddr
}
