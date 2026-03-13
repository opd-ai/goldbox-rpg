package server

import (
	"context"
	"net"
)

// WebSocketConn defines an interface for WebSocket connections that abstracts
// library-specific implementations. This enables gradual migration from
// gorilla/websocket to nhooyr.io/websocket without breaking changes.
//
// Implementations should be thread-safe for concurrent read/write operations.
//
// Related types:
//   - gorillaWebSocketConn: Implementation using gorilla/websocket
//   - nhooyrWebSocketConn: Implementation using nhooyr.io/websocket (future)
type WebSocketConn interface {
	// ReadMessage reads a message from the WebSocket connection.
	// The context can be used to set deadlines or cancel the read operation.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//
	// Returns:
	//   - messageType: Type of message (1 = TextMessage, 2 = BinaryMessage)
	//   - p: Message payload bytes
	//   - err: Error if read failed or connection closed
	ReadMessage(ctx context.Context) (messageType int, p []byte, err error)

	// WriteMessage writes a message to the WebSocket connection.
	// The context can be used to set deadlines or cancel the write operation.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//   - messageType: Type of message (1 = TextMessage, 2 = BinaryMessage)
	//   - data: Message payload bytes to send
	//
	// Returns:
	//   - err: Error if write failed or connection closed
	WriteMessage(ctx context.Context, messageType int, data []byte) error

	// Close closes the WebSocket connection with a status code and reason.
	//
	// Parameters:
	//   - code: WebSocket close status code (e.g., 1000 for normal closure)
	//   - reason: Human-readable reason for closing
	//
	// Returns:
	//   - err: Error if close failed
	Close(code int, reason string) error

	// RemoteAddr returns the remote network address of the WebSocket connection.
	//
	// Returns:
	//   - net.Addr: Remote address of the connected client
	RemoteAddr() net.Addr
}

// WebSocket message type constants (compatible with both gorilla and nhooyr)
const (
	// TextMessage denotes a text data message (UTF-8 encoded)
	WebSocketTextMessage = 1
	// BinaryMessage denotes a binary data message
	WebSocketBinaryMessage = 2
)

// WebSocket close status codes (RFC 6455)
const (
	// CloseNormalClosure indicates a normal closure
	CloseNormalClosure = 1000
	// CloseGoingAway indicates the endpoint is going away
	CloseGoingAway = 1001
	// CloseProtocolError indicates a protocol error
	CloseProtocolError = 1002
	// CloseInvalidFramePayloadData indicates invalid frame payload data
	CloseInvalidFramePayloadData = 1007
	// ClosePolicyViolation indicates a policy violation
	ClosePolicyViolation = 1008
	// CloseMessageTooBig indicates the message is too big
	CloseMessageTooBig = 1009
	// CloseInternalServerErr indicates an internal server error
	CloseInternalServerErr = 1011
)
