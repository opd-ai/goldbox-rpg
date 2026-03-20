# Multiplayer Implementation Plan

Generated: 2026-03-20

## 1. Overview

### Goals

Add first-class cooperative PvE multiplayer support to the Gold Box RPG Engine, allowing 2–6 players to explore, quest, and fight together in shared game worlds. The implementation leverages the existing session system, WebSocket infrastructure, and event-driven architecture — introducing room-based game isolation above the current single `GameState` without breaking singleplayer.

### Scope

| Feature | Priority | Phase |
|---------|----------|-------|
| Room-based game isolation | Required | 1 |
| Shared world exploration | Required | 2 |
| Cooperative turn-based combat | Required | 3 |
| Optional PvP (arena/duel) | Optional | 4 |

### Non-Goals

- Massively multiplayer (>6 players per room)
- Persistent server-side worlds across restarts (beyond existing `pkg/persistence/` capabilities)
- Voice chat or non-game social features
- Cross-server federation

### Backward Compatibility

Singleplayer remains the zero-configuration default. When no room methods are called, the server creates an implicit default room (`"default"`) that behaves identically to the current single `GameState`. All existing RPC methods (`move`, `attack`, `castSpell`, `endTurn`, etc.) continue to work unchanged — the room context is inferred from the session.

---

## 2. Architecture Changes

### Current Architecture

The server uses a single `GameState` (`pkg/server/state.go:35`) containing one `WorldState`, one `TurnManager`, one `TimeManager`, and a shared `Sessions` map. All players share the same world via `World.Objects` and `World.Players` (`pkg/game/world.go:14`). The `RPCServer` (`pkg/server/server.go:93`) holds a single `state *GameState` and broadcasts events to all sessions via `WebSocketBroadcaster.broadcastToAll()` (`pkg/server/websocket.go:701`).

```
┌─────────────────────────────────────────┐
│                RPCServer                │
│  state: *GameState (single instance)    │
│  sessions: map[string]*PlayerSession    │
│  broadcaster: *WebSocketBroadcaster     │
└─────────────────────────────────────────┘
```

### Proposed Architecture

Introduce a `RoomManager` that owns multiple `GameRoom` instances, each wrapping its own `GameState`. The `RPCServer.state` field becomes a convenience reference to the default room. Sessions are assigned to rooms; broadcasts are scoped per-room using the `EditorBroadcaster.broadcastToMapEditors()` pattern (`pkg/server/websocket_editor.go:139`).

```
┌──────────────────────────────────────────────────────────┐
│                       RPCServer                          │
│  roomManager: *RoomManager                               │
│  sessions: map[string]*PlayerSession  (global registry)  │
│  broadcaster: *WebSocketBroadcaster   (room-scoped)      │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  GameRoom     │  │  GameRoom    │  │  GameRoom    │   │
│  │  "default"    │  │  "room-abc"  │  │  "room-xyz"  │   │
│  │  ┌─────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │   │
│  │  │GameState│  │  │  │GameState│  │  │  │GameState│  │   │
│  │  └─────────┘  │  │  └────────┘  │  │  └────────┘  │   │
│  │  sessions: {} │  │  sessions:{}│  │  sessions:{}│   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
└──────────────────────────────────────────────────────────┘
```

### New Types

All new types follow the project convention of `sync.RWMutex` fields with `yaml:"-"` tags.

#### `GameRoom` (new file: `pkg/server/room.go`)

```go
type GameRoom struct {
    mu          sync.RWMutex            `yaml:"-"`
    RoomID      string                  `yaml:"room_id"`
    Name        string                  `yaml:"room_name"`
    HostID      string                  `yaml:"room_host"`       // Session ID of room creator
    State       *GameState              `yaml:"room_state"`
    Sessions    map[string]*PlayerSession `yaml:"room_sessions"`
    MaxPlayers  int                     `yaml:"room_max_players"`
    Password    string                  `yaml:"-"`               // Optional room password (never serialized)
    CreatedAt   time.Time               `yaml:"room_created_at"`
    Adventure   string                  `yaml:"room_adventure"`  // Loaded adventure pack ID
    Settings    RoomSettings            `yaml:"room_settings"`
}
```

#### `RoomSettings` (in `pkg/server/room.go`)

```go
type RoomSettings struct {
    PvPEnabled    bool          `yaml:"settings_pvp_enabled"`
    TurnTimeout   time.Duration `yaml:"settings_turn_timeout"`
    MaxPlayers    int           `yaml:"settings_max_players"`
    AllowMidJoin  bool          `yaml:"settings_allow_mid_join"`
    DifficultyMod float64       `yaml:"settings_difficulty_mod"` // 1.0 = normal
}
```

#### `RoomManager` (new file: `pkg/server/room_manager.go`)

```go
type RoomManager struct {
    mu      sync.RWMutex          `yaml:"-"`
    rooms   map[string]*GameRoom  `yaml:"rooms"`
    server  *RPCServer            `yaml:"-"`
    maxRooms int                  `yaml:"max_rooms"`
}
```

### Types Modified (Not Replaced)

| Existing Type | File | Change |
|--------------|------|--------|
| `RPCServer` | `pkg/server/server.go:93` | Add `roomManager *RoomManager` field; keep `state *GameState` as alias to default room |
| `PlayerSession` | `pkg/server/session.go:113` | Add `RoomID string yaml:"room_id"` field to track room assignment |
| `WebSocketBroadcaster` | `pkg/server/websocket.go:588` | Add `broadcastToRoom(roomID string, message interface{})` method alongside existing `broadcastToAll()` |
| `GameState` | `pkg/server/state.go:35` | No structural change; instances are now owned per-room instead of singleton |

---

## 3. Server-Side Changes

### 3.1 Room Lifecycle

New file: `pkg/server/room.go`

Room lifecycle methods are registered in the `methodRegistry` (`pkg/server/server.go`) following the existing `HandlerFunc` pattern (`func(json.RawMessage) (interface{}, error)`).

**Create Room:**
```go
func (s *RPCServer) handleCreateRoom(params json.RawMessage) (interface{}, error) {
    // 1. Parse request (name, max_players, password, adventure_id)
    // 2. Validate session via getSessionForMove(sessionID)
    // 3. defer s.releaseSession(session)
    // 4. Generate room ID via uuid.New().String()
    // 5. Create GameState with NewGameState() and NewTimeManager()
    // 6. Register room in RoomManager
    // 7. Assign session to room (session.RoomID = roomID)
    // 8. Emit event: eventSys.Emit(GameEvent{Type: EventRoomCreated, ...})
    // 9. Return {success: true, room_id: string, room_name: string}
}
```

**Join Room:**
```go
func (s *RPCServer) handleJoinRoom(params json.RawMessage) (interface{}, error) {
    // 1. Parse request (room_id, password)
    // 2. Validate session via getSessionForMove(sessionID)
    // 3. defer s.releaseSession(session)
    // 4. Look up room in RoomManager; check password if set
    // 5. Check MaxPlayers limit and AllowMidJoin setting
    // 6. Add player to room's GameState via room.State.AddPlayer(session)
    // 7. Update session.RoomID
    // 8. Emit EventPlayerJoinedRoom to room participants
    // 9. Return {success: true, room_id: string, players: []PlayerInfo}
}
```

**Leave Room:**
```go
func (s *RPCServer) handleLeaveRoom(params json.RawMessage) (interface{}, error) {
    // 1. Validate session
    // 2. Remove player from room's GameState
    // 3. Clear session.RoomID (or reassign to "default")
    // 4. If host leaves, promote next player or destroy room
    // 5. Emit EventPlayerLeftRoom to remaining participants
    // 6. Return {success: true}
}
```

**List Rooms:**
```go
func (s *RPCServer) handleListRooms(params json.RawMessage) (interface{}, error) {
    // 1. No session validation required (lobby endpoint)
    // 2. Iterate RoomManager.rooms under RLock
    // 3. Return {rooms: []RoomInfo} with public metadata (no passwords)
}
```

**Method Registration** (added to `pkg/server/server.go` alongside existing registrations):
```go
s.methodRegistry[MethodCreateRoom] = s.handleCreateRoom
s.methodRegistry[MethodJoinRoom]   = s.handleJoinRoom
s.methodRegistry[MethodLeaveRoom]  = s.handleLeaveRoom
s.methodRegistry[MethodListRooms]  = s.handleListRooms
s.methodRegistry[MethodGetRoom]    = s.handleGetRoom
s.methodRegistry[MethodKickPlayer] = s.handleKickPlayer
```

### 3.2 Scoped Broadcasting

The current `broadcastToAll()` (`pkg/server/websocket.go:701`) iterates all sessions. For multiplayer, broadcasts must be scoped to room participants, reusing the pattern from `EditorBroadcaster.broadcastToMapEditors()` (`pkg/server/websocket_editor.go:139`).

**New method on `WebSocketBroadcaster`:**
```go
// broadcastToRoom sends a message to all sessions in a specific room.
// Pattern: snapshot sessions under RLock, filter by RoomID, send under WSWriteMu.
func (wb *WebSocketBroadcaster) broadcastToRoom(roomID string, message interface{}) {
    wb.mu.RLock()
    sessions := make([]*PlayerSession, 0)
    for _, session := range wb.server.sessions {
        if session.RoomID == roomID && session.Connected && session.WSConn != nil {
            sessions = append(sessions, session)
        }
    }
    wb.mu.RUnlock()

    for _, session := range sessions {
        writeWSJSON(session.WSConn, session, message) // uses session.WSWriteMu internally
    }
}
```

**Migration path for existing handlers:** The `handleEvent()` method in `WebSocketBroadcaster` is updated to extract the room ID from the `GameEvent.Data["room_id"]` field and call `broadcastToRoom()` instead of `broadcastToAll()`. When `room_id` is empty or `"default"`, behavior is identical to current `broadcastToAll()`.

### 3.3 Session-to-Room Resolution

All existing handlers (e.g., `handleMove`, `handleAttack`, `handleCastSpell`) use `getSessionForMove()` (`pkg/server/handlers.go:161`) to retrieve the session. The room context is resolved from `session.RoomID`:

```go
// getRoomForSession resolves the GameRoom for a session's current room assignment.
func (s *RPCServer) getRoomForSession(session *PlayerSession) (*GameRoom, error) {
    roomID := session.RoomID
    if roomID == "" {
        roomID = "default"
    }
    return s.roomManager.GetRoom(roomID)
}
```

Existing handlers are modified minimally — replacing `s.state` references with `room.State` where `room` is obtained from `getRoomForSession(session)`. This keeps the handler logic unchanged while scoping operations to the correct room.

### 3.4 Turn Management for Multiple Human Players

The existing `TurnManager` (`pkg/server/combat.go:61`) already supports multiple entity IDs in `Initiative []string` and tracks the current actor via `CurrentIndex`. The key changes for multiplayer combat:

1. **Initiative includes all human players**: When `StartCombat()` is called, all players in the room's `World.Players` are added to the initiative roll alongside NPCs/enemies. The existing `CombatGroups map[string][]string` (`pkg/server/combat.go:61`) groups allied players together.

2. **Turn notifications become player-specific**: The `EventTurnStart` broadcast includes the `active_player_id` so each client knows whose turn it is. Non-active players see a "Waiting for [PlayerName]'s turn" indicator.

3. **Turn timeout enforcement**: The existing `turnTimer` and `turnDuration` (`DefaultTurnDuration = 10 * time.Second` at `pkg/server/combat.go:15`) apply to human players. On timeout, the turn auto-advances via `endTurn()`. For multiplayer, `turnDuration` should be configurable per room via `RoomSettings.TurnTimeout` (default 30 seconds for human players, 10 seconds for AI).

### 3.5 Game State Isolation

Each `GameRoom` owns its own `GameState` instance with independent:
- `WorldState *game.World` — separate `Objects`, `Players`, `NPCs`, `SpatialIndex`
- `TurnManager` — independent combat/initiative state
- `TimeManager` — independent game clock

The `RPCServer` retains global-scope resources that are shared across rooms:
- `spellManager *game.SpellManager` — read-only spell data
- `adventureManager *game.AdventureManager` — read-only adventure packs
- `pcgManager *pcg.PCGManager` — procedural generation (stateless per call)
- `metrics *Metrics` — global server metrics
- `rateLimiter *RateLimiter` — global rate limiting
- `config *config.Config` — server configuration

---

## 4. Client-Side Changes

### 4.1 Room Selection / Lobby Screen

**New `ScreenState` value** (added to `pkg/wasmui/types_ui.go:103`):
```go
const (
    ScreenSplash      ScreenState = iota // 0
    ScreenMainMenu                       // 1
    ScreenExploration                    // 2
    ScreenVictory                        // 3
    ScreenDefeat                         // 4
    ScreenLobby                          // 5 — NEW: Room selection screen
)
```

**New file: `pkg/wasmui/lobby_screen.go`**

The lobby screen displays:
- List of available rooms (from `room.list` RPC) with player count, adventure name, and host
- "Create Room" button that opens a room creation dialog
- "Join Room" button for selected room (prompts for password if protected)
- "Quick Play" button that joins the default singleplayer room
- Transition: `ScreenMainMenu` → `ScreenLobby` (if multiplayer selected) → `ScreenExploration`

```go
func (g *Game) drawLobbyScreen(screen *ebiten.Image) {
    // Draw room list panel using Gold Box UI palette (ColorPanelBorder, ColorGold)
    // Each room entry: [Name] [Players: 2/4] [Adventure: Mines of Madness] [Host: PlayerName]
    // Selected room highlighted with ColorHighlight
    // Action buttons at bottom using drawColoredText()
}

func (g *Game) handleLobbyInput() {
    // Arrow keys navigate room list
    // Enter joins selected room via room.join RPC
    // 'C' creates new room via room.create RPC
    // Escape returns to ScreenMainMenu
}
```

### 4.2 Rendering Other Players

**Modified file: `pkg/wasmui/exploration.go`**

In the exploration view, other players in the same room appear as character sprites on the tile map. The `GameStateResult` (`pkg/wasmui/rpc_client_wasm.go`) already includes `World.Players` data — currently unused beyond the local player.

```go
func (g *Game) drawOtherPlayers(screen *ebiten.Image, vpX, vpY, vpW, vpH int) {
    // Iterate g.otherPlayers (populated from gameState.World.Players minus own player)
    // For each, calculate screen position relative to viewport
    // Draw character sprite at position using existing sprite system
    // Draw player name above sprite using drawColoredText() with ColorPlayerName
}
```

**Modified file: `pkg/wasmui/combat_screen.go`**

In combat view, other human players appear in the initiative panel with distinct styling:
- Own character: highlighted with `ColorGold`
- Allied players: displayed with `ColorPlayerName` (green)
- Enemies: displayed with `ColorEnemyName` (red)

### 4.3 Turn Indication in Combat

**Modified file: `pkg/wasmui/combat_screen.go`**

The client distinguishes active turn ownership:

```go
func (g *Game) updateTurnState(event map[string]interface{}) {
    activePlayerID := event["active_player_id"].(string)
    g.mu.Lock()
    defer g.mu.Unlock()
    if activePlayerID == g.player.ID {
        g.combat.IsMyTurn = true
        g.addLogMessage("Your turn!", MessageSystem)
    } else {
        g.combat.IsMyTurn = false
        playerName := event["active_player_name"].(string)
        g.addLogMessage(fmt.Sprintf("Waiting for %s's turn...", playerName), MessageSystem)
    }
}
```

When `IsMyTurn == false`, combat action buttons are grayed out and input is blocked. A "Waiting..." overlay displays the active player's name.

### 4.4 New WASM RPC Methods

**Modified file: `pkg/wasmui/rpc_methods.go`**

```go
func (c *RPCClient) CreateRoom(name string, maxPlayers int, adventure string) (*CreateRoomResult, error) {
    result, err := c.Call("room.create", map[string]interface{}{
        "name": name, "max_players": maxPlayers, "adventure": adventure,
    })
    // Parse and return CreateRoomResult
}

func (c *RPCClient) JoinRoom(roomID, password string) (*JoinRoomResult, error) { ... }
func (c *RPCClient) LeaveRoom() error { ... }
func (c *RPCClient) ListRooms() ([]RoomInfo, error) { ... }
```

---

## 5. New RPC Methods

All methods follow the existing `validate session → process logic → emit event → return response` pattern from `pkg/server/handlers.go`. Method constants are added to `pkg/server/constants.go`.

### Method Constants

```go
const (
    MethodCreateRoom RPCMethod = "room.create"
    MethodJoinRoom   RPCMethod = "room.join"
    MethodLeaveRoom  RPCMethod = "room.leave"
    MethodListRooms  RPCMethod = "room.list"
    MethodGetRoom    RPCMethod = "room.get"
    MethodKickPlayer RPCMethod = "room.kick"
)
```

### Event Constants

```go
const (
    EventRoomCreated     game.EventType = 200 + iota // 200
    EventRoomDestroyed                                // 201
    EventPlayerJoinedRoom                             // 202
    EventPlayerLeftRoom                               // 203
    EventPlayerKicked                                 // 204
)
```

### Method Reference Table

| Method | Params | Response | Auth | Description |
|--------|--------|----------|------|-------------|
| `room.create` | `{name: string, max_players?: int, password?: string, adventure?: string, settings?: RoomSettings}` | `{success: bool, room_id: string, room_name: string}` | Session required | Create a new game room. Creator becomes host. |
| `room.join` | `{room_id: string, password?: string}` | `{success: bool, room_id: string, players: []PlayerInfo}` | Session required | Join an existing room. Fails if full or password wrong. |
| `room.leave` | `{}` | `{success: bool}` | Session required | Leave current room. Host leaving promotes next player or destroys room. |
| `room.list` | `{include_full?: bool}` | `{rooms: []RoomInfo}` | None | List available rooms with metadata. Passwords are never included. |
| `room.get` | `{room_id: string}` | `{room: RoomInfo, players: []PlayerInfo}` | None | Get detailed room information. |
| `room.kick` | `{player_id: string, reason?: string}` | `{success: bool}` | Host only | Remove a player from the room. |

### Data Schemas

**`RoomInfo`** (returned by `room.list` and `room.get`):
```json
{
    "room_id": "uuid-string",
    "name": "Dungeon Crawlers",
    "host_name": "PlayerOne",
    "player_count": 2,
    "max_players": 4,
    "adventure": "mines_of_madness",
    "has_password": true,
    "in_combat": false,
    "created_at": "2026-03-20T04:00:00Z"
}
```

**`PlayerInfo`** (returned in room join/get responses):
```json
{
    "player_id": "uuid-string",
    "name": "PlayerOne",
    "class": "Fighter",
    "level": 5,
    "is_host": true,
    "is_connected": true
}
```

---

## 6. Shared World Considerations

### Multiple Players in World State

The existing `World` struct (`pkg/game/world.go:14`) already supports multiple players via `Players map[string]*Player` and `Objects map[string]GameObject`. Multiple human players are added to both maps when they join a room:

```go
// In GameRoom.AddPlayer():
room.State.WorldState.mu.Lock()
room.State.WorldState.Players[player.GetID()] = player
room.State.WorldState.Objects[player.GetID()] = player
room.State.WorldState.mu.Unlock()
```

### Spatial Index Queries

The `SpatialIndex` (`pkg/game/spatial_index.go`) already supports multiple objects at arbitrary positions. Key multiplayer considerations:

- **`GetObjectsInRadius(center, radius)`** — Returns all entities including other players. Used for spell AoE targeting; must include allied players for friendly-fire checks when PvP is disabled.
- **`GetObjectsAt(pos)`** — Returns all objects at a tile. Used to prevent two players from occupying the same tile (collision).
- **`UpdateObjectPosition(objectID, newPos)`** — Updates spatial index atomically. Protected by `World.mu` RWMutex.

### Movement Collision

When a player moves, `World.ValidateMove()` (`pkg/game/world.go:337`) checks for obstacles via `GetObjectsAt(newPos)`. For multiplayer, this must also check for other players:

```go
func (w *World) ValidateMove(player *Player, newPos Position) error {
    // Existing: bounds check, obstacle check
    // New: check if another player occupies newPos
    objects := w.GetObjectsAt(newPos)
    for _, obj := range objects {
        if _, isPlayer := obj.(*Player); isPlayer && obj.GetID() != player.GetID() {
            return fmt.Errorf("tile occupied by another player")
        }
    }
    return nil
}
```

### Concurrent Position Updates

Multiple players may attempt to move simultaneously. The `World.mu` RWMutex (`pkg/game/world.go:14`) serializes position updates. The write lock is held during `UpdateObjectPosition()`, preventing two players from moving to the same tile concurrently.

---

## 7. Combat Multiplayer

### Initiative with Multiple Human Players

The existing `TurnManager.Initiative []string` (`pkg/server/combat.go:61`) holds entity IDs in turn order. For multiplayer combat, human player IDs are interspersed with NPC/enemy IDs based on initiative rolls:

```
Initiative: ["player-alice", "goblin-1", "player-bob", "goblin-2", "goblin-3"]
```

`TurnManager.IsCurrentTurn(entityID)` (`pkg/server/combat.go:247`) remains unchanged — it checks `Initiative[CurrentIndex] == entityID`. Only the current player's actions are accepted; other players receive `ErrNotYourTurn` from `validateCombatConstraints()` (`pkg/server/handlers.go:180`).

### Turn Flow for Human Players

```
1. TurnManager.AdvanceTurn() sets CurrentIndex to next entity
2. If entity is human player:
   a. Emit EventTurnStart with {active_player_id, active_player_name}
   b. broadcastToRoom() sends to all clients in room
   c. Active player's client enables combat actions (IsMyTurn = true)
   d. Other players' clients show "Waiting for [Name]" overlay
   e. turnTimer starts (configurable via RoomSettings.TurnTimeout)
   f. On timeout or explicit endTurn: advance to next entity
3. If entity is NPC:
   a. AI processes action immediately
   b. Results broadcast to room
   c. Advance to next entity
```

### Disconnection Mid-Combat

When a player disconnects during their combat turn:

1. **Immediate**: `HandleWebSocket()` (`pkg/server/websocket.go:234`) cleanup sets `session.Connected = false`
2. **Turn timeout**: The existing `turnTimer` fires after `turnDuration`. The disconnected player's turn is skipped via `endTurn()`.
3. **Grace period**: The player's character remains in combat for `reconnectGracePeriod` (default: 2 minutes). If they reconnect within this window, they resume play.
4. **AI takeover**: After the grace period, the player's character switches to defensive AI behavior — using basic attacks on nearest enemies, no spell casting, no movement away from current position.
5. **Session preserved**: The `PlayerSession` is not cleaned up during the 30-minute `sessionCleanupInterval` (`pkg/server/session.go:221`) as long as the room exists, allowing reconnection.

### CombatGroups for Allied Players

The existing `CombatGroups map[string][]string` groups allied entities. All human players in a room share a combat group:

```go
combatGroups := map[string][]string{
    "players": {"player-alice", "player-bob"},
    "enemies": {"goblin-1", "goblin-2", "goblin-3"},
}
```

This prevents friendly-fire in PvE mode. When `RoomSettings.PvPEnabled` is true, each player gets their own combat group.

### Turn Timeout Configuration

| Context | Default Timeout | Configurable Via |
|---------|----------------|------------------|
| Singleplayer | 10 seconds (`DefaultTurnDuration`) | N/A |
| Multiplayer (human) | 30 seconds | `RoomSettings.TurnTimeout` |
| Multiplayer (AI) | 2 seconds | Server config |

---

## 8. Persistence & Reconnection

### Saving Multiplayer State

Room state is persisted using the existing `pkg/persistence/` system. Each room saves independently:

```
data/saves/rooms/{room_id}/
    state.yaml      — GameState (world, time, turns)
    sessions.yaml   — Player session data (without WebSocket connections)
    room.yaml       — Room metadata and settings
```

The existing `fileStore` interface on `RPCServer` (`pkg/server/server.go:93`) provides `Save(path, data)` and `Load(path, data)` methods. Room persistence follows the same pattern as the existing auto-save system (`autoSaveCancel` on RPCServer).

### Player Disconnect/Reconnect

**Disconnect handling:**
1. WebSocket `onclose` triggers cleanup in `HandleWebSocket()` (`pkg/server/websocket.go:234`)
2. `session.Connected = false`, `session.WSConn = nil`
3. Session remains in room's `Sessions` map (not removed)
4. Room broadcasts `EventPlayerLeftRoom` with `{temporary: true}` to indicate disconnection (not voluntary leave)
5. Other players see "[PlayerName] disconnected" in message log

**Reconnect handling:**
1. Player reconnects WebSocket → `HandleWebSocket()` associates new connection
2. Player calls `joinGame` with existing `session_id` from cookie → `tryAttachToExistingSession()` (`pkg/server/handlers.go`) finds existing session
3. Session's `RoomID` is still set → player is automatically placed back in their room
4. Room broadcasts `EventPlayerJoinedRoom` with `{reconnected: true}`
5. Client receives full `GameState` via `getGameState` RPC to resynchronize
6. Other players see "[PlayerName] reconnected" in message log

### Session Migration

If a player's cookie is lost but they remember their room, they can call `room.join` with the room ID. The server creates a new session but preserves their character data if the room host allows mid-join. The character's position, inventory, and quest state are maintained in `World.Players` keyed by character name rather than session ID.

---

## 9. Migration Path

### Phase 1: Room Infrastructure (Estimated: 2–3 weeks)

**Goal:** Room creation, joining, leaving, and listing work. Singleplayer unchanged.

**Deliverables:**
- [ ] `pkg/server/room.go` — `GameRoom` struct, room lifecycle methods
- [ ] `pkg/server/room_manager.go` — `RoomManager` struct, room CRUD operations
- [ ] `pkg/server/constants.go` — New `RPCMethod` and `EventType` constants for rooms
- [ ] `pkg/server/server.go` — Add `roomManager` field, create default room on startup
- [ ] `pkg/server/session.go` — Add `RoomID` field to `PlayerSession`
- [ ] `pkg/server/handlers.go` — Register room methods in `methodRegistry`
- [ ] Tests: Room creation, join/leave lifecycle, max player enforcement

**Acceptance Criteria:**
- `room.create` / `room.join` / `room.leave` / `room.list` RPCs function correctly
- Default room created automatically on server startup
- Existing singleplayer tests pass without modification
- `go test -race ./pkg/server/... -run TestRoom` passes

### Phase 2: Shared Exploration (Estimated: 2–3 weeks)

**Goal:** Multiple players see each other in the game world and move independently.

**Deliverables:**
- [ ] `pkg/server/websocket.go` — `broadcastToRoom()` method on `WebSocketBroadcaster`
- [ ] `pkg/server/handlers.go` — Modify `handleMove` to use room-scoped state and broadcast
- [ ] `pkg/game/world.go` — Multi-player collision checks in `ValidateMove()`
- [ ] `pkg/wasmui/lobby_screen.go` — Lobby UI screen
- [ ] `pkg/wasmui/types_ui.go` — `ScreenLobby` constant
- [ ] `pkg/wasmui/exploration.go` — Render other players on map
- [ ] `pkg/wasmui/rpc_methods.go` — Room RPC client methods

**Acceptance Criteria:**
- Two clients in same room see each other's movement in real time
- Movement events scoped to room (other rooms don't receive them)
- Lobby screen lists rooms and allows creation/joining
- `go test -race ./pkg/server/... -run TestMultiPlayer` passes

### Phase 3: Cooperative Combat (Estimated: 3–4 weeks)

**Goal:** Multiple human players participate in turn-based combat with shared initiative.

**Deliverables:**
- [ ] `pkg/server/combat.go` — Multi-player initiative ordering, per-player turn timeouts
- [ ] `pkg/server/handlers.go` — Room-scoped combat actions (attack, spell, endTurn)
- [ ] `pkg/wasmui/combat_screen.go` — Turn indication (own turn vs. waiting), allied player rendering
- [ ] `pkg/server/room.go` — Disconnect handling, AI takeover after grace period
- [ ] `pkg/wasmui/game.go` — `IsMyTurn` state management, input blocking during other turns

**Acceptance Criteria:**
- Two players alternate turns in shared combat
- Non-active player cannot take actions (receives error)
- Disconnected player's turn is skipped after timeout
- Reconnected player resumes mid-combat
- `go test -race ./pkg/server/... -run TestCombatMulti` passes

### Phase 4: Optional PvP (Estimated: 2–3 weeks)

**Goal:** Players can optionally engage in PvP combat (arena/duel).

**Deliverables:**
- [ ] `pkg/server/room.go` — `RoomSettings.PvPEnabled` enforcement
- [ ] `pkg/server/combat.go` — Separate combat groups per player when PvP enabled
- [ ] `pkg/server/handlers.go` — PvP-specific attack validation (friendly-fire rules)
- [ ] `pkg/wasmui/combat_screen.go` — PvP targeting UI (allow selecting allied players)

**Acceptance Criteria:**
- PvP rooms allow player-vs-player attacks
- Non-PvP rooms reject attacks targeting allied players
- PvP flag visible in room list and room settings
- `go test -race ./pkg/server/... -run TestPvP` passes

---

## 10. Testing Strategy

### E2E Test Extensions

Extend existing patterns from `test/e2e/websocket_test.go`:

```go
// TestMultiPlayerRoomLifecycle tests room creation, joining, and leaving.
func TestMultiPlayerRoomLifecycle(t *testing.T) {
    // Pattern: mirrors TestWebSocketMultipleClients (line 127)
    client1 := helper.Client()
    client2 := helper.NewClient()

    // Client 1 creates room
    room, err := client1.Call("room.create", map[string]interface{}{
        "name": "Test Room", "max_players": 4,
    })
    require.NoError(t, err)
    roomID := room.(map[string]interface{})["room_id"].(string)

    // Client 2 joins room
    _, err = client2.Call("room.join", map[string]interface{}{"room_id": roomID})
    require.NoError(t, err)

    // Client 1 receives join event
    event := client1.WaitForEvent("player_joined", 3*time.Second)
    require.NotNil(t, event)
}

// TestMultiPlayerMovementBroadcast tests room-scoped movement events.
func TestMultiPlayerMovementBroadcast(t *testing.T) {
    // Pattern: mirrors TestWebSocketMovementEvents (line 29)
    // Create room, join 2 clients, move client1, verify client2 receives event
    // Verify a client in a DIFFERENT room does NOT receive the event
}

// TestMultiPlayerCombatTurns tests alternating turns between human players.
func TestMultiPlayerCombatTurns(t *testing.T) {
    // Pattern: mirrors TestWebSocketTurnEvents (line 83)
    // Create room with 2 players, start combat
    // Verify correct player receives turn_start events
    // Verify non-active player's actions are rejected
}
```

### Race Condition Testing

All multiplayer tests must pass with the race detector:
```bash
go test -race ./pkg/server/... -run "TestRoom|TestMultiPlayer" -count=1
go test -race ./test/e2e/... -run "TestMultiPlayer" -count=1
```

Critical race scenarios to test:
- Two players joining the same room simultaneously
- Two players moving to the same tile simultaneously
- Player disconnecting while another player is mid-action
- Room destruction while a player is joining
- Multiple rooms running combat simultaneously

### Load Testing

Validate server performance under multiplayer load:
- **Concurrent rooms**: 50 rooms with 4 players each (200 sessions)
- **Message throughput**: Measure broadcast latency per room with 6 players
- **Memory**: Profile per-room `GameState` memory footprint
- **Cleanup**: Verify empty rooms are destroyed and resources freed

```bash
go test -bench BenchmarkRoomBroadcast -benchmem ./pkg/server/...
go test -bench BenchmarkConcurrentRooms -benchmem ./pkg/server/...
```

### Browser Playtests

Extend `test/browser/browser_playtest_test.go` with multi-tab tests using chromedp:
- Open two browser tabs, each connecting to the same room
- Verify both tabs render the game world
- Verify movement in one tab is visible in the other

---

## 11. Thread Safety

### Locking Strategy

All multiplayer state follows the project's `sync.RWMutex` convention with `yaml:"-"` tags.

**Lock Hierarchy** (acquire in this order to prevent deadlocks):

```
1. RoomManager.mu          — protects rooms map
2. GameRoom.mu             — protects room state and session list
3. GameState.stateMu       — protects world/turn/time state (existing)
4. GameState.worldMu       — protects World operations (existing)
5. GameState.turnMu        — protects TurnManager operations (existing)
6. GameState.sessionMu     — protects per-room sessions (existing)
7. PlayerSession.WSWriteMu — protects WebSocket writes (existing)
```

Never acquire a higher-numbered lock while holding a lower-numbered one. Never hold `RoomManager.mu` while calling into `GameState` methods.

**Key Locking Patterns:**

| Operation | Lock(s) Held | Type |
|-----------|-------------|------|
| List rooms | `RoomManager.mu.RLock()` | Read |
| Create room | `RoomManager.mu.Lock()` | Write |
| Join room | `RoomManager.mu.RLock()` → `GameRoom.mu.Lock()` | Read + Write |
| Player move | `GameRoom.mu.RLock()` → `GameState.worldMu.Lock()` | Read + Write |
| Broadcast to room | `GameRoom.mu.RLock()` → per-session `WSWriteMu.Lock()` | Read + Write |
| Room cleanup | `RoomManager.mu.Lock()` → `GameRoom.mu.Lock()` | Write + Write |

**WebSocket Write Safety:**

All WebSocket writes continue to use the `writeWSJSON()` helper (`pkg/server/websocket.go:329`) which acquires `session.WSWriteMu` before calling `conn.WriteJSON()`. This prevents races between room-scoped broadcasts and direct RPC responses, following the same pattern documented in the codebase.

**Session Reference Counting:**

The existing `addRef()` / `release()` / `isInUse()` pattern (`pkg/server/session.go`) applies to room operations. When processing a room join/leave, the handler calls `getSessionForMove()` which increments the ref count, and defers `releaseSession()`. The session cleanup goroutine (`startSessionCleanup()` at `pkg/server/session.go:221`) checks `isInUse()` before removal, preventing cleanup of sessions actively participating in room operations.

**Atomic Operations:**

- `PlayerSession.inUse` (existing): `atomic.Int32` for reference counting without mutex
- `RPCClient.connected` (existing): `atomic.Bool` for connection state
- `RPCClient.requestID` (existing): `atomic.Int64` for unique request IDs
- `GameState.cacheVersion` (existing): `atomic.Int32` for optimistic cache invalidation

No new atomic fields are needed; multiplayer state uses mutex-based protection consistent with the rest of the codebase.

---

## Summary

| Aspect | Approach |
|--------|----------|
| Game isolation | `RoomManager` → `GameRoom` → `GameState` (per-room instances) |
| Singleplayer compat | Implicit `"default"` room, zero config change |
| Broadcasting | `broadcastToRoom()` reusing `EditorBroadcaster` pattern |
| Turn management | Existing `TurnManager.Initiative` with human player IDs |
| Disconnect handling | Grace period → AI takeover → session preserved |
| New RPC methods | 6 room methods following `HandlerFunc` pattern |
| New files | `room.go`, `room_manager.go`, `lobby_screen.go` |
| Thread safety | Hierarchical locking, existing `sync.RWMutex` convention |
| External deps | None (uses existing WebSocket, event, session infra) |
| Estimated effort | 9–13 weeks across 4 phases |

**Total New Constants**: 6 `RPCMethod` + 5 `EventType`
**Total New Types**: 4 (`GameRoom`, `RoomSettings`, `RoomManager`, plus `ScreenLobby`)
**Total Modified Types**: 4 (`RPCServer`, `PlayerSession`, `WebSocketBroadcaster`, `GameState`)
**Total New Files**: 3 server-side + 1 client-side
**Total Modified Files**: ~12 across `pkg/server/` and `pkg/wasmui/`
