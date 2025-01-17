# server
--
    import "github.com/opd-ai/goldbox-rpg/pkg/server"

Package server implements the game server and combat system functionality

## Usage

```go
const (
	EventCombatStart game.EventType = 100 + iota
	EventCombatEnd
	EventTurnStart
	EventTurnEnd
	EventMovement
)
```
EventCombatStart represents when combat begins in the game. This event is
triggered when characters initiate or are forced into combat. Event number: 100
(base combat event number + iota) Related events: EventCombatEnd,
EventTurnStart, EventTurnEnd

#### func  CreateItemDrop

```go
func CreateItemDrop(item game.Item, char *game.Character, dropPosition game.Position) game.GameObject
```
CreateItemDrop creates a new item object when dropped from inventory.

Parameters:

    - item: The item being dropped
    - char: The character dropping the item
    - dropPosition: Where the item should be placed

Returns:

    - game.GameObject: The created item object

#### type CombatState

```go
type CombatState struct {
	// ActiveCombatants contains the IDs of all entities currently in combat
	ActiveCombatants []string `yaml:"combat_active_entities"`
	// RoundCount tracks the current combat round number
	RoundCount int `yaml:"combat_round_count"`
	// CombatZone defines the center position of the combat area
	CombatZone game.Position `yaml:"combat_zone_center"`
	// StatusEffects maps entity IDs to their active effects
	StatusEffects map[string][]game.Effect `yaml:"combat_status_effects"`
}
```

CombatState represents the current state of a combat encounter. It tracks
participating entities, round count, combat area, and active effects.

#### type DelayedAction

```go
type DelayedAction struct {
	// ActorID is the ID of the entity performing the action
	ActorID string `yaml:"action_actor_id"`
	// ActionType defines the type of action to be performed
	ActionType string `yaml:"action_type"`
	// Target specifies the position where the action will take effect
	Target game.Position `yaml:"action_target_pos"`
	// TriggerTime determines when the action should be executed
	TriggerTime game.GameTime `yaml:"action_trigger_time"`
	// Parameters contains additional data needed for the action
	Parameters []string `yaml:"action_parameters"`
}
```

DelayedAction represents a combat action that will be executed at a specific
time.

#### type GameState

```go
type GameState struct {
	WorldState  *game.World               `yaml:"state_world"`    // Current world state
	TurnManager *TurnManager              `yaml:"state_turns"`    // Turn management
	TimeManager *TimeManager              `yaml:"state_time"`     // Time tracking
	Sessions    map[string]*PlayerSession `yaml:"state_sessions"` // Active player sessions
}
```

GameState represents the core game state container managing all dynamic game
elements. It provides thread-safe access to the world state, turn sequencing,
time tracking, and player session management.

Fields:

    - WorldState: Holds the current state of the game world including entities, items, etc
    - TurnManager: Manages turn order and action resolution for game entities
    - TimeManager: Tracks game time progression and scheduling
    - Sessions: Maps session IDs to active PlayerSession objects
    - mu: Provides thread-safe access to state
    - updates: Channel for broadcasting state changes to listeners

Thread Safety: All public methods are protected by mutex to ensure thread-safe
concurrent access. The updates channel allows for non-blocking notifications of
state changes.

Related Types:

    - game.World
    - TurnManager
    - TimeManager
    - PlayerSession

#### type PlayerSession

```go
type PlayerSession struct {
	SessionID  string       `yaml:"session_id"`  // Unique session identifier
	Player     *game.Player `yaml:"player"`      // Associated player
	LastActive time.Time    `yaml:"last_active"` // Last activity timestamp
	Connected  bool         `yaml:"connected"`   // Connection status
}
```

PlayerSession represents an active game session for a player, managing their
connection state and activity tracking. It maintains the link between a player
and their current game session.

Fields:

    - SessionID: A unique string identifier for this specific session
    - Player: Pointer to the associated game.Player instance containing player data
    - LastActive: Timestamp of the most recent player activity in this session
    - Connected: Boolean flag indicating if the player is currently connected

Related types:

    - game.Player: The player entity associated with this session

#### type RPCMethod

```go
type RPCMethod string
```

RPCMethod represents a unique identifier for RPC methods in the system. It is a
string type alias used to strongly type RPC method names and prevent errors from
mistyped method strings.

```go
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
```
MethodMove represents an RPC method for handling player movement actions in the
game. This method allows a player character to change their position on the game
map. Related methods: MethodEndTurn, MethodGetGameState

Expected payload parameters: - position: Vec2D - Target destination coordinates
- characterID: string - ID of the character being moved

Returns: - error if movement is invalid or character cannot move

Edge cases: - Movement blocked by obstacles/terrain - Character has insufficient
movement points - Position is outside map bounds

#### type RPCServer

```go
type RPCServer struct {
}
```

RPCServer represents the main RPC server instance that handles game state and
player sessions. It provides functionality for managing game state, player
sessions, and event handling.

Fields:

    - state: Pointer to GameState that maintains the current game state
    - eventSys: Pointer to game.EventSystem for handling game events
    - mu: RWMutex for thread-safe access to server resources
    - timekeeper: Pointer to TimeManager for managing game time and scheduling
    - sessions: Map of player session IDs to PlayerSession objects

Related types:

    - GameState
    - game.EventSystem
    - TimeManager
    - PlayerSession

#### func  NewRPCServer

```go
func NewRPCServer() *RPCServer
```
NewRPCServer creates and initializes a new RPCServer instance with default
configuration. It sets up the core game systems including:

    - World state management
    - Turn-based gameplay handling
    - Time tracking and management
    - Player session tracking

Returns:

    - *RPCServer: A fully initialized server instance ready to handle RPC requests

Related types:

    - GameState: Contains the core game state
    - TurnManager: Manages turn order and progression
    - TimeManager: Handles in-game time tracking
    - PlayerSession: Tracks individual player connections
    - EventSystem: Handles game event dispatching

#### func (*RPCServer) ServeHTTP

```go
func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request)
```
ServeHTTP handles incoming JSON-RPC requests over HTTP, implementing the
http.Handler interface. It processes POST requests only and expects a JSON-RPC
2.0 formatted request body.

Parameters:

    - w http.ResponseWriter: The response writer for sending the HTTP response
    - r *http.Request: The incoming HTTP request containing the JSON-RPC payload

The request body should contain a JSON object with:

    - jsonrpc: String specifying the JSON-RPC version (must be "2.0")
    - method: The RPC method name to invoke
    - params: The parameters for the method (as raw JSON)
    - id: Request identifier that will be echoed back in the response

Error handling:

    - Returns 405 Method Not Allowed if request is not POST
    - Returns JSON-RPC error code -32700 for invalid JSON
    - Returns JSON-RPC error code -32603 for internal errors during method execution

Related:

    - handleMethod: Processes the individual RPC method calls
    - writeResponse: Formats and sends successful responses
    - writeError: Formats and sends error responses

#### type ScheduledEvent

```go
type ScheduledEvent struct {
	EventID     string        `yaml:"event_id"`           // Event identifier
	EventType   string        `yaml:"event_type"`         // Type of event
	TriggerTime game.GameTime `yaml:"event_trigger_time"` // When to trigger
	Parameters  []string      `yaml:"event_parameters"`   // Event data
	Repeating   bool          `yaml:"event_is_repeating"` // Whether it repeats
}
```

ScheduledEvent represents a future event that will be triggered at a specific
game time. It is used to schedule in-game events like monster spawns, weather
changes, or quest updates.

Fields:

    - EventID: Unique string identifier for the event
    - EventType: Category/type of the event (e.g. "spawn", "weather", etc)
    - TriggerTime: The game.GameTime when this event should execute
    - Parameters: Additional string data needed for the event execution
    - Repeating: If true, the event will reschedule itself after triggering

Related types:

    - game.GameTime: Represents the in-game time when event triggers

#### type ScriptContext

```go
type ScriptContext struct {
	ScriptID     string                 `yaml:"script_id"`            // Script identifier
	Variables    map[string]interface{} `yaml:"script_variables"`     // Script state
	LastExecuted time.Time              `yaml:"script_last_executed"` // Last run timestamp
	IsActive     bool                   `yaml:"script_is_active"`     // Execution state
}
```

ScriptContext represents the execution state and variables of a running script
in the game. It maintains context between script executions including variables
and timing.

Fields:

    - ScriptID: Unique identifier string for the script
    - Variables: Map storing script state variables and their values
    - LastExecuted: Timestamp of when the script was last run
    - IsActive: Boolean flag indicating if script is currently executing

Related types:

    - Server.Scripts (map[string]*ScriptContext)
    - ScriptEngine interface

Thread-safety: This struct should be protected by a mutex when accessed
concurrently

#### type StateUpdate

```go
type StateUpdate struct {
	UpdateType string                 `yaml:"update_type"`      // Type of update
	EntityID   string                 `yaml:"update_entity_id"` // Affected entity
	ChangeData map[string]interface{} `yaml:"update_data"`      // Update details
	Timestamp  time.Time              `yaml:"update_timestamp"` // When it occurred
}
```

StateUpdate represents an atomic change to the game state. It captures what
changed, which entity was affected, and when the change occurred.

Fields:

    - UpdateType: String identifying the type of update (e.g. "MOVE", "DAMAGE")
    - EntityID: Unique identifier for the affected game entity
    - ChangeData: Map containing the specific changes/updates to apply.
      Values can be of any type due to interface{}
    - Timestamp: When this state update occurred

StateUpdate is used by the game engine to track and apply changes to entities.
Updates are processed in chronological order based on Timestamp.

Related types:

    - Entity: The game object being modified
    - Game: Top level game state manager

#### type TimeManager

```go
type TimeManager struct {
	CurrentTime     game.GameTime    `yaml:"time_current"`          // Current game time
	TimeScale       float64          `yaml:"time_scale"`            // Time progression rate
	LastTick        time.Time        `yaml:"time_last_tick"`        // Last update time
	ScheduledEvents []ScheduledEvent `yaml:"time_scheduled_events"` // Pending events
}
```

TimeManager handles game time progression and scheduled event management. It
maintains the current game time, controls time progression speed, and manages a
queue of scheduled future events.

Fields:

    - CurrentTime: The current in-game time represented as a GameTime struct
    - TimeScale: Multiplier that controls how fast game time progresses relative to real time (e.g. 2.0 = twice as fast)
    - LastTick: Real-world timestamp of the most recent time update
    - ScheduledEvents: Slice of pending events to be triggered at specific game times

Related types:

    - game.GameTime - Represents a point in game time
    - ScheduledEvent - Defines a future event to occur at a specific game time

#### func  NewTimeManager

```go
func NewTimeManager() *TimeManager
```
NewTimeManager creates and initializes a new TimeManager instance.

The TimeManager handles game time tracking, time scaling, and scheduled event
management. It maintains the current game time, real time mapping, and a list of
scheduled events.

Returns:

    - *TimeManager: A new TimeManager instance initialized with:
    - Current time set to now
    - Game ticks starting at 0
    - Default time scale of 1.0
    - Empty scheduled events list

Related types:

    - game.GameTime
    - ScheduledEvent

#### type TurnManager

```go
type TurnManager struct {
	// CurrentRound represents the current combat round number
	CurrentRound int `yaml:"turn_current_round"`
	// Initiative holds entity IDs in their initiative order
	Initiative []string `yaml:"turn_initiative_order"`
	// CurrentIndex tracks the current actor's position in the initiative order
	CurrentIndex int `yaml:"turn_current_index"`
	// IsInCombat indicates whether combat is currently active
	IsInCombat bool `yaml:"turn_in_combat"`
	// CombatGroups maps entity IDs to their allied group members
	CombatGroups map[string][]string `yaml:"turn_combat_groups"`
	// DelayedActions holds actions to be executed at a later time
	DelayedActions []DelayedAction `yaml:"turn_delayed_actions"`
}
```

TurnManager handles combat turn order and initiative tracking. It manages the
flow of combat rounds and tracks allied groups.

#### func (*TurnManager) AdvanceTurn

```go
func (tm *TurnManager) AdvanceTurn() string
```
AdvanceTurn moves to the next entity in the initiative order. Increments the
round counter when returning to the first entity.

Returns:

    - string: The ID of the next entity in the initiative order, or empty string if not in combat

#### func (*TurnManager) IsCurrentTurn

```go
func (tm *TurnManager) IsCurrentTurn(entityID string) bool
```
IsCurrentTurn checks if the given entity is the current actor in combat.

Parameters:

    - entityID: The ID of the entity to check

Returns:

    - bool: true if it's the entity's turn, false otherwise

#### func (*TurnManager) StartCombat

```go
func (tm *TurnManager) StartCombat(initiative []string)
```
StartCombat initializes a new combat encounter with the given initiative order.

Parameters:

    - initiative: Ordered slice of entity IDs representing turn order
