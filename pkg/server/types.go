package server

import (
	"fmt"
	"sync/atomic"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/gorilla/websocket"
)

// RPCMethod represents a unique identifier for RPC methods in the system.
// It is a string type alias used to strongly type RPC method names and
// prevent errors from mistyped method strings.
type RPCMethod string

// MethodMove represents an RPC method for handling player movement actions in the game.
// This method allows a player character to change their position on the game map.
// Related methods: MethodEndTurn, MethodGetGameState
//
// Expected payload parameters:
// - position: Vec2D - Target destination coordinates
// - characterID: string - ID of the character being moved
//
// Returns:
// - error if movement is invalid or character cannot move
//
// Edge cases:
// - Movement blocked by obstacles/terrain
// RPCMethod constants are defined in constants.go
// - Character has insufficient movement points
// - Position is outside map bounds

// EventCombat constants are defined in constants.go
// EventCombatStart represents when combat begins in the game. This event is triggered
// when characters initiate or are forced into combat.
// Event number: 100 (base combat event number + iota)
// Related events: EventCombatEnd, EventTurnStart, EventTurnEnd

// StateUpdate represents an atomic change to the game state.
// It captures what changed, which entity was affected, and when the change occurred.
//
// Fields:
//   - UpdateType: String identifying the type of update (e.g. "MOVE", "DAMAGE")
//   - EntityID: Unique identifier for the affected game entity
//   - ChangeData: Map containing the specific changes/updates to apply.
//     Values can be of any type due to interface{}
//   - Timestamp: When this state update occurred
//
// StateUpdate is used by the game engine to track and apply changes to entities.
// Updates are processed in chronological order based on Timestamp.
//
// Related types:
//   - Entity: The game object being modified
//   - Game: Top level game state manager
type StateUpdate struct {
	UpdateType string                 `yaml:"update_type"`      // Type of update
	EntityID   string                 `yaml:"update_entity_id"` // Affected entity
	ChangeData map[string]interface{} `yaml:"update_data"`      // Update details
	Timestamp  time.Time              `yaml:"update_timestamp"` // When it occurred
}

// PlayerSession represents an active game session for a player, managing their connection state
// and activity tracking. It maintains the link between a player and their current game session.
//
// Fields:
//   - SessionID: A unique string identifier for this specific session
//   - Player: Pointer to the associated game.Player instance containing player data
//   - LastActive: Timestamp of the most recent player activity in this session
//   - Connected: Boolean flag indicating if the player is currently connected
//
// Related types:
//   - game.Player: The player entity associated with this session
type PlayerSession struct {
	SessionID   string          `yaml:"session_id"`  // Unique session identifier
	Player      *game.Player    `yaml:"player"`      // Associated player
	LastActive  time.Time       `yaml:"last_active"` // Last activity timestamp
	CreatedAt   time.Time       `yaml:"created_at"`  // Session creation timestamp
	Connected   bool            `yaml:"connected"`   // Connection status
	MessageChan chan []byte     `yaml:"-"`           // Channel for sending messages
	WSConn      *websocket.Conn `yaml:"-"`           // WebSocket connection
	inUse       int32           `yaml:"-"`           // Atomic counter for active usage (prevents cleanup)
}

// Update modifies the player session with the provided updates.
func (p *PlayerSession) Update(updateMap map[string]interface{}) error {
	if p == nil {
		return fmt.Errorf("cannot update nil PlayerSession")
	}

	for key, value := range updateMap {
		switch key {
		case "player":
			if playerData, ok := value.(map[string]interface{}); ok {
				p.Player.Update(playerData)
			}
		case "connected":
			if connected, ok := value.(bool); ok {
				p.Connected = connected
			}
		case "lastActive":
			if timestamp, ok := value.(time.Time); ok {
				p.LastActive = timestamp
			}
		case "sessionId":
			if sessionID, ok := value.(string); ok {
				p.SessionID = sessionID
			}
		}
	}

	return nil
}

// Clone creates a deep copy of the PlayerSession.
func (p *PlayerSession) Clone() *PlayerSession {
	if p == nil {
		return nil
	}

	clone := &PlayerSession{
		SessionID:   p.SessionID,
		Player:      p.Player.Clone(), // Assuming Player has a Clone method
		LastActive:  p.LastActive,
		CreatedAt:   p.CreatedAt,
		Connected:   p.Connected,
		MessageChan: make(chan []byte, 500), // Use consistent buffer size
		WSConn:      p.WSConn,               // Keep same connection
		inUse:       0,                      // Reset usage counter for clone
	}
	return clone
}

// PublicData returns a sanitized version of the PlayerSession for client consumption.
func (p *PlayerSession) PublicData() interface{} {
	return struct {
		SessionID  string      `json:"sessionId"`
		PlayerData interface{} `json:"player"`
		Connected  bool        `json:"connected"`
		LastActive time.Time   `json:"lastActive"`
	}{
		SessionID:  p.SessionID,
		PlayerData: p.Player.PublicData(),
		Connected:  p.Connected,
		LastActive: p.LastActive,
	}
}

// addRef atomically increments the usage counter to prevent cleanup
func (p *PlayerSession) addRef() {
	atomic.AddInt32(&p.inUse, 1)
}

// release atomically decrements the usage counter
func (p *PlayerSession) release() {
	atomic.AddInt32(&p.inUse, -1)
}

// isInUse atomically checks if the session is currently being used
func (p *PlayerSession) isInUse() bool {
	return atomic.LoadInt32(&p.inUse) > 0
}
