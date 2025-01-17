package server

import (
	"goldbox-rpg/pkg/game"
	"time"
)

// RPCMethod represents available RPC methods
type RPCMethod string

const (
	MethodMove         RPCMethod = "move"
	MethodAttack       RPCMethod = "attack"
	MethodCastSpell    RPCMethod = "castSpell"
	MethodUseItem      RPCMethod = "useItem"
	MethodApplyEffect  RPCMethod = "applyEffect"
	MethodStartCombat  RPCMethod = "startCombat"
	MethodEndTurn      RPCMethod = "endTurn"
	MethodGetGameState RPCMethod = "getGameState"
	MethodJoinGame     RPCMethod = "joinGame"
	MethodLeaveGame    RPCMethod = "leaveGame"
)

// Additional EventType constants
const (
	EventCombatStart game.EventType = 100 + iota
	EventCombatEnd
	EventTurnStart
	EventTurnEnd
	EventMovement
)

// StateUpdate represents a game state change notification
type StateUpdate struct {
	UpdateType string                 `yaml:"update_type"`      // Type of update
	EntityID   string                 `yaml:"update_entity_id"` // Affected entity
	ChangeData map[string]interface{} `yaml:"update_data"`      // Update details
	Timestamp  time.Time              `yaml:"update_timestamp"` // When it occurred
}

// PlayerSession represents an active player connection
type PlayerSession struct {
	SessionID  string       `yaml:"session_id"`  // Unique session identifier
	Player     *game.Player `yaml:"player"`      // Associated player
	LastActive time.Time    `yaml:"last_active"` // Last activity timestamp
	Connected  bool         `yaml:"connected"`   // Connection status
}
