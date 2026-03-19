package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// EditorEventType represents types of editor-specific events.
type EditorEventType string

// Editor event type constants
const (
	EditorEventTileUpdate EditorEventType = "tile_update"
	EditorEventMapCreated EditorEventType = "map_created"
	EditorEventMapLoaded  EditorEventType = "map_loaded"
	EditorEventMapSaved   EditorEventType = "map_saved"
	EditorEventCursorMove EditorEventType = "cursor_move"
	EditorEventSelectTool EditorEventType = "select_tool"
	EditorEventUndoRedo   EditorEventType = "undo_redo"
)

// EditorMessage represents a message in the editor WebSocket protocol.
type EditorMessage struct {
	Type      EditorEventType        `json:"type"`
	MapID     string                 `json:"map_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// TileUpdateData represents data for a tile update message.
type TileUpdateData struct {
	X           int  `json:"x"`
	Y           int  `json:"y"`
	SpriteX     int  `json:"sprite_x"`
	SpriteY     int  `json:"sprite_y"`
	Walkable    bool `json:"walkable"`
	Transparent bool `json:"transparent"`
}

// EditorSession represents an active editor session with map state.
type EditorSession struct {
	SessionID   string
	MapID       string
	CurrentMap  *game.GameMap
	WSConn      WebSocketConn
	mu          sync.Mutex
	subscribers map[string]*EditorSession
}

// EditorBroadcaster manages WebSocket connections for editor clients.
type EditorBroadcaster struct {
	server   *RPCServer
	sessions map[string]*EditorSession
	mu       sync.RWMutex
	active   bool
}

// NewEditorBroadcaster creates a new editor broadcaster instance.
func NewEditorBroadcaster(server *RPCServer) *EditorBroadcaster {
	return &EditorBroadcaster{
		server:   server,
		sessions: make(map[string]*EditorSession),
		active:   false,
	}
}

// Start activates the editor broadcaster.
func (eb *EditorBroadcaster) Start() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.active {
		return
	}

	eb.active = true
	logrus.Info("Editor broadcaster started")
}

// Stop deactivates the editor broadcaster.
func (eb *EditorBroadcaster) Stop() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.active = false
	logrus.Info("Editor broadcaster stopped")
}

// RegisterSession registers a new editor session.
func (eb *EditorBroadcaster) RegisterSession(session *EditorSession) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.sessions[session.SessionID] = session
	logrus.WithField("sessionID", session.SessionID).Debug("Editor session registered")
}

// UnregisterSession removes an editor session.
func (eb *EditorBroadcaster) UnregisterSession(sessionID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	delete(eb.sessions, sessionID)
	logrus.WithField("sessionID", sessionID).Debug("Editor session unregistered")
}

// BroadcastTileUpdate broadcasts a tile update to all sessions editing the same map.
func (eb *EditorBroadcaster) BroadcastTileUpdate(mapID, sourceSessionID string, data TileUpdateData) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if !eb.active {
		return
	}

	message := EditorMessage{
		Type:      EditorEventTileUpdate,
		MapID:     mapID,
		SessionID: sourceSessionID,
		Data: map[string]interface{}{
			"x":           data.X,
			"y":           data.Y,
			"sprite_x":    data.SpriteX,
			"sprite_y":    data.SpriteY,
			"walkable":    data.Walkable,
			"transparent": data.Transparent,
		},
	}

	eb.broadcastToMapEditors(mapID, sourceSessionID, message)
}

// broadcastToMapEditors sends a message to all sessions editing a specific map.
func (eb *EditorBroadcaster) broadcastToMapEditors(mapID, excludeSession string, message EditorMessage) {
	// Create a snapshot of sessions under read lock to avoid concurrent map access
	eb.mu.RLock()
	sessions := make([]*EditorSession, 0, len(eb.sessions))
	sessionIDs := make([]string, 0, len(eb.sessions))
	for sessionID, session := range eb.sessions {
		if session.MapID == mapID && sessionID != excludeSession && session.WSConn != nil {
			sessions = append(sessions, session)
			sessionIDs = append(sessionIDs, sessionID)
		}
	}
	eb.mu.RUnlock()

	// Iterate over snapshot without holding lock
	for i, session := range sessions {
		session.mu.Lock()
		err := session.WSConn.WriteJSON(message)
		session.mu.Unlock()

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"sessionID": sessionIDs[i],
				"error":     err.Error(),
			}).Error("Failed to send editor message")
		}
	}
}

// HandleEditorWebSocket manages WebSocket connections for editor clients.
func (s *RPCServer) HandleEditorWebSocket(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("function", "HandleEditorWebSocket")

	playerSession, ok := s.getSessionFromContext(w, r, "HandleEditorWebSocket")
	if !ok {
		return
	}

	conn, err := s.upgradeConnection(w, r)
	if err != nil {
		return
	}

	editorSession := &EditorSession{
		SessionID:   playerSession.SessionID,
		WSConn:      conn,
		subscribers: make(map[string]*EditorSession),
	}

	defer func() {
		conn.CloseNow()
		if s.editorBroadcaster != nil {
			s.editorBroadcaster.UnregisterSession(editorSession.SessionID)
		}
	}()

	if s.editorBroadcaster != nil {
		s.editorBroadcaster.RegisterSession(editorSession)
	}

	// Send connection confirmation
	confirmMsg := EditorMessage{
		Type:      EditorEventType("connected"),
		SessionID: playerSession.SessionID,
		Data: map[string]interface{}{
			"status": "connected",
		},
	}
	if err := conn.WriteJSON(confirmMsg); err != nil {
		logger.WithError(err).Error("Failed to send confirmation")
		return
	}

	// Handle incoming messages
	s.handleEditorMessages(conn, editorSession)
}

// handleEditorMessages processes incoming WebSocket messages for editor sessions.
func (s *RPCServer) handleEditorMessages(conn WebSocketConn, session *EditorSession) {
	logger := logrus.WithField("sessionID", session.SessionID)

	for {
		var msg EditorMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			// Log error and break out of loop - connection closed
			logger.WithError(err).Debug("WebSocket read error (connection likely closed)")
			break
		}

		logger.WithField("type", msg.Type).Debug("Received editor message")

		switch msg.Type {
		case EditorEventTileUpdate:
			s.handleEditorTileUpdate(session, msg)
		case EditorEventSelectTool:
			// Tool selection is client-side only, but we can broadcast cursor position
			s.handleEditorToolSelect(session, msg)
		case EditorEventCursorMove:
			s.handleEditorCursorMove(session, msg)
		default:
			logger.WithField("type", msg.Type).Warn("Unknown editor message type")
		}
	}
}

// handleEditorTileUpdate processes tile update messages.
func (s *RPCServer) handleEditorTileUpdate(session *EditorSession, msg EditorMessage) {
	logger := logrus.WithFields(logrus.Fields{
		"function":  "handleEditorTileUpdate",
		"sessionID": session.SessionID,
		"mapID":     msg.MapID,
	})

	// Extract tile update data
	x, _ := msg.Data["x"].(float64)
	y, _ := msg.Data["y"].(float64)
	spriteX, _ := msg.Data["sprite_x"].(float64)
	spriteY, _ := msg.Data["sprite_y"].(float64)
	walkable, _ := msg.Data["walkable"].(bool)
	transparent, _ := msg.Data["transparent"].(bool)

	data := TileUpdateData{
		X:           int(x),
		Y:           int(y),
		SpriteX:     int(spriteX),
		SpriteY:     int(spriteY),
		Walkable:    walkable,
		Transparent: transparent,
	}

	// Update the map in the session
	if session.CurrentMap != nil {
		tile := session.CurrentMap.GetTile(data.X, data.Y)
		if tile != nil {
			tile.SpriteX = data.SpriteX
			tile.SpriteY = data.SpriteY
			tile.Walkable = data.Walkable
			tile.Transparent = data.Transparent
		}
	}

	// Broadcast to other editors of the same map
	if s.editorBroadcaster != nil {
		s.editorBroadcaster.BroadcastTileUpdate(msg.MapID, session.SessionID, data)
	}

	logger.WithFields(logrus.Fields{
		"x": data.X,
		"y": data.Y,
	}).Debug("Tile update processed")
}

// handleEditorToolSelect handles tool selection messages.
func (s *RPCServer) handleEditorToolSelect(session *EditorSession, msg EditorMessage) {
	// Tool selection is primarily client-side
	// We just log for debugging purposes
	tool, _ := msg.Data["tool"].(string)
	logrus.WithFields(logrus.Fields{
		"sessionID": session.SessionID,
		"tool":      tool,
	}).Debug("Tool selected")
}

// handleEditorCursorMove handles cursor movement messages for collaborative editing.
func (s *RPCServer) handleEditorCursorMove(session *EditorSession, msg EditorMessage) {
	// Optional: broadcast cursor position to other editors for collaboration
	// This is a nice-to-have feature for multi-user editing
	x, _ := msg.Data["x"].(float64)
	y, _ := msg.Data["y"].(float64)

	logrus.WithFields(logrus.Fields{
		"sessionID": session.SessionID,
		"x":         int(x),
		"y":         int(y),
	}).Debug("Cursor moved")
}

// EditorWebSocketMessage represents the JSON structure for editor WebSocket messages.
type EditorWebSocketMessage struct {
	Type      string          `json:"type"`
	MapID     string          `json:"map_id,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// SendEditorUpdate sends an update to the editor session.
func (es *EditorSession) SendEditorUpdate(eventType EditorEventType, data map[string]interface{}) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.WSConn == nil {
		return nil
	}

	msg := EditorMessage{
		Type:      eventType,
		MapID:     es.MapID,
		SessionID: es.SessionID,
		Data:      data,
	}

	return es.WSConn.WriteJSON(msg)
}
