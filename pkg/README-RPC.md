# Gold Box RPG JSON-RPC API Documentation

## Connection Details
- Base URL: `http://localhost:8080/rpc`
- Protocol: HTTP/1.1 
- Content-Type: `application/json`
- Method: POST

## Base Request Format
```json
{
    "jsonrpc": "2.0",
    "method": "methodName",
    "params": {},
    "id": 1
}
```

## Methods

### move
Moves a player character to a new position on the game map.

**Parameters:**
```json
{
    "session_id": string,
    "direction": "north" | "south" | "east" | "west"
}
```

**Response:**
```json
{
    "success": boolean,
    "position": {
        "x": number,
        "y": number
    }
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'move',
        params: {
            session_id: 'abc123',
            direction: 'north'
        },
        id: 1
    })
});
```

```go
// Go
type MoveParams struct {
    SessionID string         `json:"session_id"`
    Direction game.Direction `json:"direction"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "move",
    Params:  MoveParams{
        SessionID: "abc123",
        Direction: "north",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "move",
    "params": {
        "session_id": "abc123",
        "direction": "north"
    },
    "id": 1
  }'
```

### attack
Performs a combat attack action.

**Parameters:**
```json
{
    "session_id": string,
    "target_id": string,
    "weapon_id": string
}
```

**Response:**
```json
{
    "success": boolean,
    "damage": number
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'attack',
        params: {
            session_id: 'abc123',
            target_id: 'monster_1',
            weapon_id: 'sword_1'
        },
        id: 1
    })
});
```

```go
// Go
type AttackParams struct {
    SessionID string `json:"session_id"`
    TargetID  string `json:"target_id"`
    WeaponID  string `json:"weapon_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "attack",
    Params:  AttackParams{
        SessionID: "abc123",
        TargetID:  "monster_1",
        WeaponID:  "sword_1",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "attack",
    "params": {
        "session_id": "abc123",
        "target_id": "monster_1",
        "weapon_id": "sword_1"
    },
    "id": 1
  }'
```

### castSpell
Casts a spell on a target or location.

**Parameters:**
```json
{
    "session_id": string,
    "spell_id": string,
    "target_id": string,
    "position": {
        "x": number,
        "y": number
    }
}
```

**Response:**
```json
{
    "success": boolean,
    "spell_id": string
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'castSpell',
        params: {
            session_id: 'abc123',
            spell_id: 'fireball_1',
            target_id: 'monster_1',
            position: {x: 10, y: 15}
        },
        id: 1
    })
});
```

```go
// Go
type SpellCastParams struct {
    SessionID string        `json:"session_id"`
    SpellID   string        `json:"spell_id"`
    TargetID  string        `json:"target_id"`
    Position  game.Position `json:"position"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "castSpell",
    Params:  SpellCastParams{
        SessionID: "abc123",
        SpellID:   "fireball_1",
        TargetID:  "monster_1",
        Position:  game.Position{X: 10, Y: 15},
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "castSpell",
    "params": {
        "session_id": "abc123",
        "spell_id": "fireball_1",
        "target_id": "monster_1",
        "position": {"x": 10, "y": 15}
    },
    "id": 1
  }'
```

### applyEffect
Applies a status effect to a target entity.

**Parameters:**
```json
{
    "session_id": string,
    "effect_type": string,
    "target_id": string,
    "magnitude": number,
    "duration": number
}
```

**Response:**
```json
{
    "success": boolean,
    "effect_id": string
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'applyEffect',
        params: {
            session_id: 'abc123',
            effect_type: 'poison',
            target_id: 'monster_1',
            magnitude: 5,
            duration: 3
        },
        id: 1
    })
});
```

```go
// Go
type ApplyEffectParams struct {
    SessionID  string          `json:"session_id"`
    EffectType game.EffectType `json:"effect_type"`
    TargetID   string          `json:"target_id"`
    Magnitude  float64         `json:"magnitude"`
    Duration   game.Duration   `json:"duration"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "applyEffect",
    Params:  ApplyEffectParams{
        SessionID:  "abc123",
        EffectType: "poison",
        TargetID:   "monster_1",
        Magnitude:  5,
        Duration:   3,
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "applyEffect",
    "params": {
        "session_id": "abc123",
        "effect_type": "poison",
        "target_id": "monster_1",
        "magnitude": 5,
        "duration": 3
    },
    "id": 1
  }'
```

### startCombat
Initiates a combat encounter with specified participants.

**Parameters:**
```json
{
    "session_id": string,
    "participant_ids": string[]
}
```

**Response:**
```json
{
    "success": boolean,
    "initiative": string[],
    "first_turn": string
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'startCombat',
        params: {
            session_id: 'abc123',
            participant_ids: ['player_1', 'monster_1', 'monster_2']
        },
        id: 1
    })
});
```

```go
// Go
type StartCombatParams struct {
    SessionID    string   `json:"session_id"`
    Participants []string `json:"participant_ids"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "startCombat",
    Params:  StartCombatParams{
        SessionID:    "abc123",
        Participants: []string{"player_1", "monster_1", "monster_2"},
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "startCombat",
    "params": {
        "session_id": "abc123",
        "participant_ids": ["player_1", "monster_1", "monster_2"]
    },
    "id": 1
  }'
```

### endTurn
Ends the current player's turn in combat.

**Parameters:**
```json
{
    "session_id": string
}
```

**Response:**
```json
{
    "success": boolean,
    "next_turn": string
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'endTurn',
        params: {
            session_id: 'abc123'
        },
        id: 1
    })
});
```

```go
// Go
type EndTurnParams struct {
    SessionID string `json:"session_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "endTurn",
    Params:  EndTurnParams{
        SessionID: "abc123",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "endTurn",
    "params": {
        "session_id": "abc123"
    },
    "id": 1
  }'
```

### getGameState
Retrieves the current game state for a session.

**Parameters:**
```json
{
    "session_id": string
}
```

**Response:**
```json
{
    "player": {
        "position": {
            "x": number,
            "y": number
        },
        "stats": {
            "hp": number,
            "max_hp": number,
            "level": number
        },
        "effects": [],
        "inventory": [],
        "spells": [],
        "experience": number
    },
    "world": {
        "visible_objects": [],
        "current_time": string,
        "combat_state": null | {
            "active_combatants": string[],
            "round_count": number,
            "combat_zone": {
                "x": number,
                "y": number
            },
            "status_effects": {
                [key: string]: game.Effect[]
            }
        }
    }
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'getGameState',
        params: {
            session_id: 'abc123'
        },
        id: 1
    })
});
```

```go
// Go
type GameStateParams struct {
    SessionID string `json:"session_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "getGameState",
    Params:  GameStateParams{
        SessionID: "abc123",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getGameState",
    "params": {
        "session_id": "abc123"
    },
    "id": 1
  }'
```

### useItem
Uses an item from player's inventory.

**Parameters:**
```json
{
    "session_id": string,
    "item_id": string,
    "target_id": string
}
```

**Response:**
```json
{
    "success": boolean,
    "effect": string
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'useItem',
        params: {
            session_id: 'abc123',
            item_id: 'potion_1',
            target_id: 'player_1'
        },
        id: 1
    })
});
```

```go
// Go
type UseItemParams struct {
    SessionID string `json:"session_id"`
    ItemID    string `json:"item_id"`
    TargetID  string `json:"target_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "useItem",
    Params:  UseItemParams{
        SessionID: "abc123",
        ItemID:    "potion_1",
        TargetID:  "player_1",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "useItem",
    "params": {
        "session_id": "abc123",
        "item_id": "potion_1",
        "target_id": "player_1"
    },
    "id": 1
  }'
```

### joinGame
Creates a new game session.

**Parameters:**
```json
{
    "player_name": string
}
```

**Response:**
```json
{
    "success": boolean,
    "session_id": string
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'joinGame',
        params: {
            player_name: 'Alice'
        },
        id: 1
    })
});
```

```go
// Go
type JoinGameParams struct {
    PlayerName string `json:"player_name"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "joinGame",
    Params:  JoinGameParams{
        PlayerName: "Alice",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "joinGame",
    "params": {
        "player_name": "Alice"
    },
    "id": 1
  }'
```

### leaveGame
Ends a game session.

**Parameters:**
```json
{
    "session_id": string
}
```

**Response:**
```json
{
    "success": boolean
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'leaveGame',
        params: {
            session_id: 'abc123'
        },
        id: 1
    })
});
```

```go
// Go
type LeaveGameParams struct {
    SessionID string `json:"session_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "leaveGame",
    Params:  LeaveGameParams{
        SessionID: "abc123",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "leaveGame",
    "params": {
        "session_id": "abc123"
    },
    "id": 1
  }'
```

### equipItem
Equips an item from the player's inventory to a specific equipment slot.

**Parameters:**
```json
{
    "session_id": string,
    "item_id": string,
    "slot": string
}
```

**Response:**
```json
{
    "success": boolean,
    "message": string,
    "equipped_item": object,
    "previous_item": object (optional)
}
```

**Valid slot names:**
- "head" - Head armor/helmets
- "neck" - Amulets/necklaces  
- "chest" - Armor/robes
- "hands" - Gloves/gauntlets
- "rings" - Rings
- "legs" - Pants/leggings
- "feet" - Boots/shoes
- "weapon_main" or "main_hand" - Primary weapon
- "weapon_off" or "off_hand" - Shield/off-hand weapon

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'equipItem',
        params: {
            session_id: 'abc123',
            item_id: 'sword_001',
            slot: 'weapon_main'
        },
        id: 1
    })
});
```

```go
// Go
type EquipItemParams struct {
    SessionID string `json:"session_id"`
    ItemID    string `json:"item_id"`
    Slot      string `json:"slot"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "equipItem",
    Params:  EquipItemParams{
        SessionID: "abc123",
        ItemID:    "sword_001",
        Slot:      "weapon_main",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "equipItem",
    "params": {
        "session_id": "abc123",
        "item_id": "sword_001",
        "slot": "weapon_main"
    },
    "id": 1
  }'
```

### unequipItem
Removes an equipped item and returns it to the player's inventory.

**Parameters:**
```json
{
    "session_id": string,
    "slot": string
}
```

**Response:**
```json
{
    "success": boolean,
    "message": string,
    "unequipped_item": object
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'unequipItem',
        params: {
            session_id: 'abc123',
            slot: 'weapon_main'
        },
        id: 1
    })
});
```

```go
// Go
type UnequipItemParams struct {
    SessionID string `json:"session_id"`
    Slot      string `json:"slot"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "unequipItem",
    Params:  UnequipItemParams{
        SessionID: "abc123",
        Slot:      "weapon_main",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "unequipItem",
    "params": {
        "session_id": "abc123",
        "slot": "weapon_main"
    },
    "id": 1
  }'
```

### getEquipment
Returns all currently equipped items for a player.

**Parameters:**
```json
{
    "session_id": string
}
```

**Response:**
```json
{
    "success": boolean,
    "equipment": {
        "slot_name": {
            "id": string,
            "name": string,
            "type": string,
            "damage": string,
            "ac": number,
            "weight": number,
            "value": number,
            "properties": [string]
        }
    },
    "total_weight": number,
    "equipment_bonuses": {
        "stat_name": number
    }
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'getEquipment',
        params: {
            session_id: 'abc123'
        },
        id: 1
    })
});
```

```go
// Go
type GetEquipmentParams struct {
    SessionID string `json:"session_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "getEquipment",
    Params:  GetEquipmentParams{
        SessionID: "abc123",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getEquipment",
    "params": {
        "session_id": "abc123"
    },
    "id": 1
  }'
```

### createCharacter
Creates a new character with specified attributes and class.

**Parameters:**
```json
{
    "name": string,
    "class": "fighter" | "mage" | "cleric" | "thief" | "ranger" | "paladin",
    "attribute_method": "roll" | "pointbuy" | "standard" | "custom",
    "custom_attributes": {
        "strength": number,
        "dexterity": number,
        "constitution": number,
        "intelligence": number,
        "wisdom": number,
        "charisma": number
    },
    "starting_equipment": boolean,
    "starting_gold": number
}
```

**Response:**
```json
{
    "success": boolean,
    "character": {
        "name": string,
        "class": string,
        "level": number,
        "attributes": {
            "strength": number,
            "dexterity": number,
            "constitution": number,
            "intelligence": number,
            "wisdom": number,
            "charisma": number
        },
        "hit_points": number,
        "max_hit_points": number
    },
    "player": {
        "id": string,
        "character": object,
        "position": {
            "x": number,
            "y": number
        }
    },
    "session_id": string,
    "errors": string[],
    "warnings": string[],
    "creation_time": string,
    "generated_stats": object,
    "starting_items": object[]
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'createCharacter',
        params: {
            name: 'Aragorn',
            class: 'ranger',
            attribute_method: 'roll',
            starting_equipment: true,
            starting_gold: 100
        },
        id: 1
    })
});
```

```go
// Go
type CreateCharacterParams struct {
    Name              string         `json:"name"`
    Class             string         `json:"class"`
    AttributeMethod   string         `json:"attribute_method"`
    CustomAttributes  map[string]int `json:"custom_attributes,omitempty"`
    StartingEquipment bool           `json:"starting_equipment"`
    StartingGold      int            `json:"starting_gold"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "createCharacter",
    Params:  CreateCharacterParams{
        Name:              "Aragorn",
        Class:             "ranger",
        AttributeMethod:   "roll",
        StartingEquipment: true,
        StartingGold:      100,
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "createCharacter",
    "params": {
        "name": "Aragorn",
        "class": "ranger",
        "attribute_method": "roll",
        "starting_equipment": true,
        "starting_gold": 100
    },
    "id": 1
  }'
```

### leaveGame
Ends a game session.

**Parameters:**
```json
{
    "session_id": string
}
```

**Response:**
```json
{
    "success": boolean
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'leaveGame',
        params: {
            session_id: 'abc123'
        },
        id: 1
    })
});
```

```go
// Go
type LeaveGameParams struct {
    SessionID string `json:"session_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "leaveGame",
    Params:  LeaveGameParams{
        SessionID: "abc123",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "leaveGame",
    "params": {
        "session_id": "abc123"
    },
    "id": 1
  }'
```

### equipItem
Equips an item from the player's inventory to a specific equipment slot.

**Parameters:**
```json
{
    "session_id": string,
    "item_id": string,
    "slot": string
}
```

**Response:**
```json
{
    "success": boolean,
    "message": string,
    "equipped_item": object,
    "previous_item": object (optional)
}
```

**Valid slot names:**
- "head" - Head armor/helmets
- "neck" - Amulets/necklaces  
- "chest" - Armor/robes
- "hands" - Gloves/gauntlets
- "rings" - Rings
- "legs" - Pants/leggings
- "feet" - Boots/shoes
- "weapon_main" or "main_hand" - Primary weapon
- "weapon_off" or "off_hand" - Shield/off-hand weapon

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'equipItem',
        params: {
            session_id: 'abc123',
            item_id: 'sword_001',
            slot: 'weapon_main'
        },
        id: 1
    })
});
```

```go
// Go
type EquipItemParams struct {
    SessionID string `json:"session_id"`
    ItemID    string `json:"item_id"`
    Slot      string `json:"slot"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "equipItem",
    Params:  EquipItemParams{
        SessionID: "abc123",
        ItemID:    "sword_001",
        Slot:      "weapon_main",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "equipItem",
    "params": {
        "session_id": "abc123",
        "item_id": "sword_001",
        "slot": "weapon_main"
    },
    "id": 1
  }'
```

### unequipItem
Removes an equipped item and returns it to the player's inventory.

**Parameters:**
```json
{
    "session_id": string,
    "slot": string
}
```

**Response:**
```json
{
    "success": boolean,
    "message": string,
    "unequipped_item": object
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'unequipItem',
        params: {
            session_id: 'abc123',
            slot: 'weapon_main'
        },
        id: 1
    })
});
```

```go
// Go
type UnequipItemParams struct {
    SessionID string `json:"session_id"`
    Slot      string `json:"slot"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "unequipItem",
    Params:  UnequipItemParams{
        SessionID: "abc123",
        Slot:      "weapon_main",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "unequipItem",
    "params": {
        "session_id": "abc123",
        "slot": "weapon_main"
    },
    "id": 1
  }'
```

### getEquipment
Returns all currently equipped items for a player.

**Parameters:**
```json
{
    "session_id": string
}
```

**Response:**
```json
{
    "success": boolean,
    "equipment": {
        "slot_name": {
            "id": string,
            "name": string,
            "type": string,
            "damage": string,
            "ac": number,
            "weight": number,
            "value": number,
            "properties": [string]
        }
    },
    "total_weight": number,
    "equipment_bonuses": {
        "stat_name": number
    }
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'getEquipment',
        params: {
            session_id: 'abc123'
        },
        id: 1
    })
});
```

```go
// Go
type GetEquipmentParams struct {
    SessionID string `json:"session_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "getEquipment",
    Params:  GetEquipmentParams{
        SessionID: "abc123",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getEquipment",
    "params": {
        "session_id": "abc123"
    },
    "id": 1
  }'
```

### createCharacter
Creates a new character with specified attributes and class.

**Parameters:**
```json
{
    "name": string,
    "class": "fighter" | "mage" | "cleric" | "thief" | "ranger" | "paladin",
    "attribute_method": "roll" | "pointbuy" | "standard" | "custom",
    "custom_attributes": {
        "strength": number,
        "dexterity": number,
        "constitution": number,
        "intelligence": number,
        "wisdom": number,
        "charisma": number
    },
    "starting_equipment": boolean,
    "starting_gold": number
}
```

**Response:**
```json
{
    "success": boolean,
    "character": {
        "name": string,
        "class": string,
        "level": number,
        "attributes": {
            "strength": number,
            "dexterity": number,
            "constitution": number,
            "intelligence": number,
            "wisdom": number,
            "charisma": number
        },
        "hit_points": number,
        "max_hit_points": number
    },
    "player": {
        "id": string,
        "character": object,
        "position": {
            "x": number,
            "y": number
        }
    },
    "session_id": string,
    "errors": string[],
    "warnings": string[],
    "creation_time": string,
    "generated_stats": object,
    "starting_items": object[]
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'createCharacter',
        params: {
            name: 'Aragorn',
            class: 'ranger',
            attribute_method: 'roll',
            starting_equipment: true,
            starting_gold: 100
        },
        id: 1
    })
});
```

```go
// Go
type CreateCharacterParams struct {
    Name              string         `json:"name"`
    Class             string         `json:"class"`
    AttributeMethod   string         `json:"attribute_method"`
    CustomAttributes  map[string]int `json:"custom_attributes,omitempty"`
    StartingEquipment bool           `json:"starting_equipment"`
    StartingGold      int            `json:"starting_gold"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "createCharacter",
    Params:  CreateCharacterParams{
        Name:              "Aragorn",
        Class:             "ranger",
        AttributeMethod:   "roll",
        StartingEquipment: true,
        StartingGold:      100,
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "createCharacter",
    "params": {
        "name": "Aragorn",
        "class": "ranger",
        "attribute_method": "roll",
        "starting_equipment": true,
        "starting_gold": 100
    },
    "id": 1
  }'
```

### leaveGame
Ends a game session.

**Parameters:**
```json
{
    "session_id": string
}
```

**Response:**
```json
{
    "success": boolean
}
```

**Examples:**

```javascript
// JavaScript
const response = await fetch('/rpc', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'leaveGame',
        params: {
            session_id: 'abc123'
        },
        id: 1
    })
});
```

```go
// Go
type LeaveGameParams struct {
    SessionID string `json:"session_id"`
}

req := &JSONRPCRequest{
    JsonRPC: "2.0",
    Method:  "leaveGame",
    Params:  LeaveGameParams{
        SessionID: "abc123",
    },
    ID: 1,
}
```

```bash
# curl
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "leaveGame",
    "params": {
        "session_id": "abc123"
    },
    "id": 1
  }'
```

## Error Codes

| Code    | Message               | Description                           |
|---------|----------------------|---------------------------------------|
| -32700  | Parse error         | Invalid JSON                          |
| -32600  | Invalid request     | Invalid JSON-RPC request              |
| -32601  | Method not found    | Unknown method                        |
| -32602  | Invalid params      | Invalid method parameters             |
| -32603  | Internal error      | Internal server error                 |