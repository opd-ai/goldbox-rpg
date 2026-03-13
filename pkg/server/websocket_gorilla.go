package server

import (
	"context"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// gorillaWebSocketConn wraps a gorilla/websocket connection to implement
// the WebSocketConn interface. This allows gradual migration to other
// WebSocket libraries while maintaining backward compatibility.
//
// Thread Safety: gorilla/websocket allows concurrent reads and writes,
// but this wrapper adds context support for cancellation.
type gorillaWebSocketConn struct {
	conn *websocket.Conn
}

// NewGorillaWebSocketConn creates a new WebSocketConn adapter from a gorilla websocket connection.
//
// Parameters:
//   - conn: The underlying gorilla websocket connection
//
// Returns:
//   - WebSocketConn: Adapter implementing the common WebSocket interface
func NewGorillaWebSocketConn(conn *websocket.Conn) WebSocketConn {
	return &gorillaWebSocketConn{conn: conn}
}

// ReadMessage reads a message from the WebSocket connection.
// The context is used to set read deadlines for cancellation support.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//
// Returns:
//   - messageType: Type of message (1 = TextMessage, 2 = BinaryMessage)
//   - p: Message payload bytes
//   - err: Error if read failed or connection closed
func (g *gorillaWebSocketConn) ReadMessage(ctx context.Context) (messageType int, p []byte, err error) {
	// Set read deadline from context if available
	if deadline, ok := ctx.Deadline(); ok {
		g.conn.SetReadDeadline(deadline)
		defer g.conn.SetReadDeadline(time.Time{}) // Clear deadline after read
	}

	// Check for context cancellation before reading
	select {
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	default:
	}

	return g.conn.ReadMessage()
}

// WriteMessage writes a message to the WebSocket connection.
// The context is used to set write deadlines for cancellation support.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//   - messageType: Type of message (1 = TextMessage, 2 = BinaryMessage)
//   - data: Message payload bytes to send
//
// Returns:
//   - err: Error if write failed or connection closed
func (g *gorillaWebSocketConn) WriteMessage(ctx context.Context, messageType int, data []byte) error {
	// Set write deadline from context if available
	if deadline, ok := ctx.Deadline(); ok {
		g.conn.SetWriteDeadline(deadline)
		defer g.conn.SetWriteDeadline(time.Time{}) // Clear deadline after write
	}

	// Check for context cancellation before writing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return g.conn.WriteMessage(messageType, data)
}

// Close closes the WebSocket connection with a status code and reason.
// It sends a close control message before closing the underlying connection.
//
// Parameters:
//   - code: WebSocket close status code (e.g., 1000 for normal closure)
//   - reason: Human-readable reason for closing
//
// Returns:
//   - err: Error if close failed
func (g *gorillaWebSocketConn) Close(code int, reason string) error {
	// Send close message with status code and reason
	msg := websocket.FormatCloseMessage(code, reason)
	g.conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
	return g.conn.Close()
}

// RemoteAddr returns the remote network address of the WebSocket connection.
//
// Returns:
//   - net.Addr: Remote address of the connected client
func (g *gorillaWebSocketConn) RemoteAddr() net.Addr {
	return g.conn.RemoteAddr()
}

// GetUnderlyingConn returns the underlying gorilla websocket connection.
// This can be used when direct access to gorilla-specific features is needed.
//
// Returns:
//   - *websocket.Conn: The underlying gorilla websocket connection
func (g *gorillaWebSocketConn) GetUnderlyingConn() *websocket.Conn {
	return g.conn
}
