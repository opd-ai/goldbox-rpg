# Visual Editor Guide

This guide covers the browser-based visual editors for content creation in GoldBox RPG Engine.

## Overview

The GoldBox RPG Engine provides two visual editors accessible through your web browser:

1. **Map Editor** (`/editor.html`) - Visual map creation and editing
2. **Quest Builder** (`/quest-builder.html`) - Visual quest chain creation

Both editors provide real-time collaboration features via WebSocket and require no command-line knowledge.

## Map Editor

### Accessing the Map Editor

Start the server and navigate to:
```
http://localhost:8080/editor.html
```

### Features

#### Create New Map
- **Endpoint**: `editor.createMap`
- **Purpose**: Initialize a new blank map with specified dimensions
- **Parameters**:
  - `session_id` (string): Your game session ID
  - `map_id` (string): Unique identifier for the new map
  - `width` (int): Map width in tiles
  - `height` (int): Map height in tiles
  - `default_terrain` (string, optional): Default terrain type (grass, stone, water, etc.)

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "editor.createMap",
  "params": {
    "session_id": "your-session-id",
    "map_id": "dungeon_level_1",
    "width": 50,
    "height": 50,
    "default_terrain": "stone"
  },
  "id": 1
}
```

#### Update Tiles
- **Endpoint**: `editor.updateTile`
- **Purpose**: Modify individual tiles (terrain, objects, NPCs)
- **Parameters**:
  - `session_id` (string): Your game session ID
  - `map_id` (string): Map to modify
  - `x` (int): Tile X coordinate
  - `y` (int): Tile Y coordinate
  - `terrain` (string, optional): New terrain type
  - `object_id` (string, optional): Place object on tile
  - `passable` (bool, optional): Override tile passability

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "editor.updateTile",
  "params": {
    "session_id": "your-session-id",
    "map_id": "dungeon_level_1",
    "x": 10,
    "y": 15,
    "terrain": "water",
    "passable": false
  },
  "id": 2
}
```

#### Save Map
- **Endpoint**: `editor.saveMap`
- **Purpose**: Persist map to storage (YAML format)
- **Parameters**:
  - `session_id` (string): Your game session ID
  - `map_id` (string): Map to save
  - `filepath` (string): Save location (relative to data directory)

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "editor.saveMap",
  "params": {
    "session_id": "your-session-id",
    "map_id": "dungeon_level_1",
    "filepath": "maps/dungeon_level_1.yaml"
  },
  "id": 3
}
```

#### Load Map
- **Endpoint**: `editor.loadMap`
- **Purpose**: Load existing map from storage
- **Parameters**:
  - `session_id` (string): Your game session ID
  - `filepath` (string): Map file to load

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "editor.loadMap",
  "params": {
    "session_id": "your-session-id",
    "filepath": "maps/dungeon_level_1.yaml"
  },
  "id": 4
}
```

### Real-Time Collaboration

The map editor supports multiple simultaneous editors via WebSocket events:

#### WebSocket Endpoint
```
ws://localhost:8080/ws
```

After connecting, you'll receive editor-specific events:

##### Tile Update Event
```json
{
  "type": "tile_update",
  "map_id": "dungeon_level_1",
  "timestamp": "2026-03-16T18:00:00Z",
  "data": {
    "x": 10,
    "y": 15,
    "terrain": "water",
    "updated_by": "other-session-id"
  }
}
```

##### Map Created Event
```json
{
  "type": "map_created",
  "map_id": "dungeon_level_2",
  "timestamp": "2026-03-16T18:01:00Z",
  "data": {
    "width": 40,
    "height": 40
  }
}
```

##### Map Saved Event
```json
{
  "type": "map_saved",
  "map_id": "dungeon_level_1",
  "timestamp": "2026-03-16T18:02:00Z",
  "data": {
    "filepath": "maps/dungeon_level_1.yaml"
  }
}
```

### Keyboard Shortcuts

*(To be implemented in future versions)*
- `Ctrl+S`: Save current map
- `Ctrl+Z`: Undo last edit
- `Ctrl+Y`: Redo last undo
- `G`: Switch to grass terrain
- `W`: Switch to water terrain
- `S`: Switch to stone terrain
- `D`: Switch to dirt terrain

## Quest Builder

### Accessing the Quest Builder

Start the server and navigate to:
```
http://localhost:8080/quest-builder.html
```

### Features

#### Create Quest
- **Endpoint**: `questEditor.create`
- **Purpose**: Define a new quest with objectives and rewards
- **Parameters**:
  - `session_id` (string): Your game session ID
  - `quest_id` (string): Unique quest identifier
  - `title` (string): Quest display name
  - `description` (string): Quest narrative
  - `giver_npc` (string): NPC who gives the quest
  - `level_requirement` (int, optional): Minimum character level
  - `objectives` (array): List of quest objectives
  - `rewards` (object): Quest completion rewards

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "questEditor.create",
  "params": {
    "session_id": "your-session-id",
    "quest_id": "slay_goblin_king",
    "title": "Slay the Goblin King",
    "description": "The goblin king terrorizes the village. Venture into his lair and defeat him.",
    "giver_npc": "village_elder",
    "level_requirement": 5,
    "objectives": [
      {
        "type": "kill",
        "target": "goblin_king",
        "count": 1,
        "description": "Defeat the Goblin King"
      }
    ],
    "rewards": {
      "experience": 500,
      "gold": 100,
      "items": ["magic_sword"]
    }
  },
  "id": 1
}
```

#### Get Quest
- **Endpoint**: `questEditor.get`
- **Purpose**: Retrieve quest details for editing
- **Parameters**:
  - `session_id` (string): Your game session ID
  - `quest_id` (string): Quest to retrieve

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "questEditor.get",
  "params": {
    "session_id": "your-session-id",
    "quest_id": "slay_goblin_king"
  },
  "id": 2
}
```

#### Update Quest
- **Endpoint**: `questEditor.update`
- **Purpose**: Modify existing quest
- **Parameters**: Same as `questEditor.create`

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "questEditor.update",
  "params": {
    "session_id": "your-session-id",
    "quest_id": "slay_goblin_king",
    "title": "Slay the Goblin King (Revised)",
    "rewards": {
      "experience": 750,
      "gold": 150,
      "items": ["magic_sword", "healing_potion"]
    }
  },
  "id": 3
}
```

#### Delete Quest
- **Endpoint**: `questEditor.delete`
- **Purpose**: Remove quest from game
- **Parameters**:
  - `session_id` (string): Your game session ID
  - `quest_id` (string): Quest to delete

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "questEditor.delete",
  "params": {
    "session_id": "your-session-id",
    "quest_id": "slay_goblin_king"
  },
  "id": 4
}
```

#### List Quests
- **Endpoint**: `questEditor.list`
- **Purpose**: Get all available quests
- **Parameters**:
  - `session_id` (string): Your game session ID
  - `status` (string, optional): Filter by status (available, active, completed)

**Example RPC Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "questEditor.list",
  "params": {
    "session_id": "your-session-id"
  },
  "id": 5
}
```

### Quest Objective Types

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `kill` | Defeat specific enemy | `target`, `count` |
| `collect` | Gather items | `item_id`, `count` |
| `talk_to` | Speak with NPC | `npc_id` |
| `explore` | Visit location | `location_id` |
| `escort` | Protect NPC | `npc_id`, `destination` |
| `deliver` | Bring item to NPC | `item_id`, `npc_id` |

### Reward Types

```yaml
rewards:
  experience: 500        # XP points
  gold: 100             # Currency
  items:                # Item IDs
    - magic_sword
    - healing_potion
  reputation:           # Faction standing
    merchants_guild: 10
```

## Session Management

Both editors require a valid game session. Obtain a session by calling:

```json
{
  "jsonrpc": "2.0",
  "method": "joinGame",
  "params": {
    "player_name": "ContentCreator"
  },
  "id": 1
}
```

Response includes `session_id` to use in all editor operations.

## Error Handling

All editor endpoints return standard JSON-RPC 2.0 errors:

| Code | Message | Cause |
|------|---------|-------|
| -32600 | Invalid Request | Malformed JSON |
| -32601 | Method not found | Unknown endpoint |
| -32602 | Invalid params | Missing/wrong parameters |
| -32603 | Internal error | Server-side failure |
| -32000 | Invalid session | Session expired or invalid |
| -32001 | Permission denied | Unauthorized operation |
| -32002 | Resource not found | Map/quest doesn't exist |

**Example Error Response**:
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "Invalid session",
    "data": "session expired after 30 minutes of inactivity"
  },
  "id": 1
}
```

## CLI Alternatives

For scripting and automation, equivalent CLI tools are available:

### Map Editor CLI
```bash
# Create new map
./bin/map-editor create --id dungeon_1 --width 50 --height 50

# Load and edit map
./bin/map-editor edit --file maps/dungeon_1.yaml

# Validate map
./bin/map-editor validate --file maps/dungeon_1.yaml
```

### Quest Builder CLI
```bash
# Create quest from template
./bin/quest-builder create --template kill_quest --output quest.yaml

# Validate quest
./bin/quest-builder validate --file quest.yaml

# List quest templates
./bin/quest-builder templates
```

See individual tool documentation for complete command reference:
- `./bin/map-editor --help`
- `./bin/quest-builder --help`

## Best Practices

### Map Design
1. **Start small**: Begin with 20x20 maps for testing, expand later
2. **Plan pathways**: Ensure player can navigate from start to exit
3. **Variety**: Mix terrain types for visual interest
4. **Performance**: Maps >100x100 may impact client rendering
5. **Save often**: Use `editor.saveMap` frequently to prevent data loss

### Quest Design
1. **Clear objectives**: Use descriptive objective text
2. **Balanced rewards**: XP ~= 100 × level_requirement
3. **Quest chains**: Use prerequisites for story progression
4. **Testing**: Verify all objectives are achievable
5. **Feedback**: Include objective progress updates

## Troubleshooting

### Editor doesn't load
- **Check server**: Ensure server running on http://localhost:8080
- **Browser console**: Open DevTools (F12) to see JavaScript errors
- **WebSocket**: Verify `/ws` endpoint accessible

### Changes not saving
- **Session timeout**: Sessions expire after 30 minutes (configurable)
- **File permissions**: Check data directory is writable
- **Validation errors**: Review server logs for save failures

### Real-time updates missing
- **WebSocket connection**: Check network tab for active ws:// connection
- **Firewall**: Ensure WebSocket traffic allowed
- **Different map**: Collaboration only visible for same `map_id`

## API Reference

Complete JSON-RPC API documentation available at:
- http://localhost:8080/api/docs
- API Specification: http://localhost:8080/api/openapi.yaml

For programmatic access, see:
- `pkg/README-RPC.md` - Complete RPC method reference
- `pkg/server/constants.go` - All available RPC methods
- `api/openapi.yaml` - OpenAPI 3.0 specification

## Examples

### Complete Map Creation Workflow

1. Join game and get session:
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"joinGame","params":{"player_name":"MapBuilder"},"id":1}'
```

2. Create new map:
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"editor.createMap","params":{"session_id":"SESSION_HERE","map_id":"test_map","width":20,"height":20},"id":2}'
```

3. Add some water tiles:
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"editor.updateTile","params":{"session_id":"SESSION_HERE","map_id":"test_map","x":5,"y":5,"terrain":"water"},"id":3}'
```

4. Save map:
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"editor.saveMap","params":{"session_id":"SESSION_HERE","map_id":"test_map","filepath":"maps/test_map.yaml"},"id":4}'
```

### Complete Quest Creation Workflow

1. Use existing session or join game
2. Create quest:
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "questEditor.create",
    "params": {
      "session_id": "SESSION_HERE",
      "quest_id": "first_quest",
      "title": "Rats in the Cellar",
      "description": "Clear the inn cellar of giant rats",
      "giver_npc": "innkeeper",
      "objectives": [{
        "type": "kill",
        "target": "giant_rat",
        "count": 5
      }],
      "rewards": {
        "experience": 100,
        "gold": 50
      }
    },
    "id": 1
  }'
```

## See Also

- [RPC API Documentation](../pkg/README-RPC.md)
- [Server Configuration Guide](../SECURITY.md#production-deployment)
- [Asset Generation Guide](../ASSET_INTEGRATION.md)
- [CLI Tools Documentation](../cmd/README.md)
