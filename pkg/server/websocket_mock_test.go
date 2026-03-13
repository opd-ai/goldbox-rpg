package server

import (
	"context"
	"net"
)

// mockWebSocketConn implements WebSocketConn for testing purposes.
// It provides a no-op implementation that can be used in unit tests
// where a real WebSocket connection is not needed.
type mockWebSocketConn struct {
	messages []interface{} // Captured messages from WriteJSON
}

// newMockWebSocketConn creates a new mock WebSocket connection for testing.
func newMockWebSocketConn() *mockWebSocketConn {
	return &mockWebSocketConn{
		messages: make([]interface{}, 0),
	}
}

// ReadMessage implements WebSocketConn.ReadMessage.
func (m *mockWebSocketConn) ReadMessage(ctx context.Context) (messageType int, p []byte, err error) {
	return 0, nil, nil
}

// WriteMessage implements WebSocketConn.WriteMessage.
func (m *mockWebSocketConn) WriteMessage(ctx context.Context, messageType int, data []byte) error {
	return nil
}

// WriteJSON implements WebSocketConn.WriteJSON.
func (m *mockWebSocketConn) WriteJSON(v interface{}) error {
	m.messages = append(m.messages, v)
	return nil
}

// ReadJSON implements WebSocketConn.ReadJSON.
func (m *mockWebSocketConn) ReadJSON(v interface{}) error {
	return nil
}

// Close implements WebSocketConn.Close.
func (m *mockWebSocketConn) Close(code int, reason string) error {
	return nil
}

// CloseNow implements WebSocketConn.CloseNow.
func (m *mockWebSocketConn) CloseNow() error {
	return nil
}

// RemoteAddr implements WebSocketConn.RemoteAddr.
func (m *mockWebSocketConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
}

// GetMessages returns all captured messages from WriteJSON calls.
func (m *mockWebSocketConn) GetMessages() []interface{} {
	return m.messages
}

// ClearMessages clears all captured messages.
func (m *mockWebSocketConn) ClearMessages() {
	m.messages = make([]interface{}, 0)
}
