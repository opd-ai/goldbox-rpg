package server

import (
	"goldbox-rpg/pkg/game"
	"time"
)

// Context key type for context values
type contextKey string

// Context keys
const (
	sessionKey   contextKey = "session"
	requestIDKey contextKey = "request_id"
)

// Session and server configuration constants
// Moved from: server.go
const (
	sessionCleanupInterval = 5 * time.Minute
	sessionTimeout         = 30 * time.Minute
)

// Session configuration constants
// MessageChanBufferSize defines the buffer size for session message channels
// Increased from 100 to provide better buffering while preventing unbounded growth
// MessageSendTimeout defines the timeout for non-blocking message sends
// Prevents goroutines from blocking indefinitely on full channels
// Moved from: session.go
const (
	MessageChanBufferSize = 500
	MessageSendTimeout    = 50 * time.Millisecond
)

// RPCMethod constants define the available RPC methods for the game server.
// These methods handle various game actions and state queries.
// - Character has insufficient movement points
// - Position is outside map bounds
// Moved from: types.go
const (
	MethodMove            RPCMethod = "move"
	MethodAttack          RPCMethod = "attack"
	MethodCastSpell       RPCMethod = "castSpell"
	MethodUseItem         RPCMethod = "useItem"
	MethodApplyEffect     RPCMethod = "applyEffect"
	MethodStartCombat     RPCMethod = "startCombat"
	MethodEndTurn         RPCMethod = "endTurn"
	MethodGetGameState    RPCMethod = "getGameState"
	MethodJoinGame        RPCMethod = "joinGame"
	MethodLeaveGame       RPCMethod = "leaveGame"
	MethodCreateCharacter RPCMethod = "createCharacter"

	// Equipment management methods
	MethodEquipItem    RPCMethod = "equipItem"
	MethodUnequipItem  RPCMethod = "unequipItem"
	MethodGetEquipment RPCMethod = "getEquipment"

	// Quest management methods
	MethodStartQuest         RPCMethod = "startQuest"
	MethodCompleteQuest      RPCMethod = "completeQuest"
	MethodUpdateObjective    RPCMethod = "updateObjective"
	MethodFailQuest          RPCMethod = "failQuest"
	MethodGetQuest           RPCMethod = "getQuest"
	MethodGetActiveQuests    RPCMethod = "getActiveQuests"
	MethodGetCompletedQuests RPCMethod = "getCompletedQuests"
	MethodGetQuestLog        RPCMethod = "getQuestLog"

	// Spell management methods
	MethodGetSpell          RPCMethod = "getSpell"
	MethodGetSpellsByLevel  RPCMethod = "getSpellsByLevel"
	MethodGetSpellsBySchool RPCMethod = "getSpellsBySchool"
	MethodGetAllSpells      RPCMethod = "getAllSpells"
	MethodSearchSpells      RPCMethod = "searchSpells"

	// Spatial query methods for efficient object retrieval
	MethodGetObjectsInRange  RPCMethod = "getObjectsInRange"
	MethodGetObjectsInRadius RPCMethod = "getObjectsInRadius"
	MethodGetNearestObjects  RPCMethod = "getNearestObjects"

	// PCG (Procedural Content Generation) methods
	MethodGenerateContent   RPCMethod = "generateContent"
	MethodRegenerateTerrain RPCMethod = "regenerateTerrain"
	MethodGenerateItems     RPCMethod = "generateItems"
	MethodGenerateLevel     RPCMethod = "generateLevel"
	MethodGenerateQuest     RPCMethod = "generateQuest"
	MethodGetPCGStats       RPCMethod = "getPCGStats"
	MethodValidateContent   RPCMethod = "validateContent"
)

// EventCombatStart represents when combat begins in the game. This event is triggered
// when characters initiate or are forced into combat.
// Event number: 100 (base combat event number + iota)
// Related events: EventCombatEnd, EventTurnStart, EventTurnEnd
// Moved from: types.go
const (
	EventCombatStart game.EventType = 100 + iota
	EventCombatEnd
	EventTurnStart
	EventTurnEnd
	EventMovement
)
