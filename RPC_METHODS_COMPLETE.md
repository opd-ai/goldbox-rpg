# Gold Box RPG JSON-RPC API - Complete Method Reference

## Summary

This document provides a comprehensive list of all 38 RPC methods available in the Gold Box RPG JSON-RPC API, organized by category, with complete parameter and response specifications.

---

## Common Patterns

### Request Format
All RPC requests follow JSON-RPC 2.0 specification:
```json
{
  "jsonrpc": "2.0",
  "method": "methodName",
  "params": { /* method-specific parameters */ },
  "id": 1
}
```

### Response Format
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": { /* response data */ },
  "error": null
}
```

### Common Response Fields
- `success` (boolean): Indicates success/failure
- `message` (string): Human-readable message
- `error` (string): Error description on failure
- All responses follow this structure: `{ "jsonrpc": "2.0", "id": <id>, "result": {...} }`

### Session Management
- Most methods require `session_id` parameter (string)
- `session_id` obtained from `joinGame` or `createCharacter`
- Sessions maintain player state and identity

---

## Category 1: Character Actions (5 Methods)

### 1. move
**Description**: Move player character to adjacent position
**RPC Method**: `move`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "direction": "north|south|east|west|n|s|e|w"
}
```
**Response**:
```json
{
  "success": boolean,
  "position": {
    "x": number,
    "y": number
  }
}
```
**Constraints**:
- Direction supports both full names and abbreviations
- Validates movement boundaries
- Consumes action points during combat (ActionCostMove)
- Cannot move if not player's turn during combat

---

### 2. attack
**Description**: Perform combat attack with weapon
**RPC Method**: `attack`
**Requires Session**: Yes
**Requires Combat**: Yes (checked but optional)
**Parameters**:
```json
{
  "session_id": "string",
  "target_id": "string",
  "weapon_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "damage": number,
  "attack_roll": number
}
```
**Constraints**:
- Must have sufficient action points (ActionCostAttack)
- Can only attack during combat turn
- Target must exist in world state
- Weapon must be available

---

### 3. castSpell
**Description**: Cast spell on target or location
**RPC Method**: `castSpell`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "spell_id": "string",
  "target_id": "string",
  "position": {
    "x": number,
    "y": number
  }
}
```
**Response**:
```json
{
  "success": boolean,
  "spell_id": "string",
  "effect": { /* spell effect details */ }
}
```
**Constraints**:
- Player must know the spell
- Spell must exist in spell database
- Consumes action points during combat (ActionCostSpell)
- Position is optional (for area spells)
- Supports both single-target and area-effect spells

---

### 4. useItem
**Description**: Use item from player inventory
**RPC Method**: `useItem`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "item_id": "string",
  "target_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "effect": "string describing the effect"
}
```
**Constraints**:
- Item must exist in player inventory
- target_id is optional (some items don't target)
- Consumable items removed from inventory after use
- Turn validation applies during combat

---

### 5. applyEffect
**Description**: Apply status effect to target entity
**RPC Method**: `applyEffect`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "effect_type": "string",
  "target_id": "string",
  "magnitude": number,
  "duration": number
}
```
**Response**:
```json
{
  "success": boolean,
  "effect_id": "string"
}
```
**Constraints**:
- Target must exist and implement EffectHolder interface
- Effect types: poison, stun, heal, buff, etc.
- Duration in game time units
- Magnitude represents strength/amount

---

## Category 2: Combat Management (2 Methods)

### 6. startCombat
**Description**: Initiate combat encounter with participants
**RPC Method**: `startCombat`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "participant_ids": ["string", "string", ...]
}
```
**Response**:
```json
{
  "success": boolean,
  "initiative": ["participant1_id", "participant2_id", ...],
  "first_turn": "participant_id"
}
```
**Constraints**:
- Cannot start combat if already in combat
- Initializes action points for all participants
- Initiative determined by d20 roll + DEX modifier
- Participants appear in initiative order

---

### 7. endTurn
**Description**: End current player's turn in combat
**RPC Method**: `endTurn`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "next_turn": "participant_id"
}
```
**Constraints**:
- Must be in combat
- Must be player's turn
- Restores action points for next participant
- Triggers end-of-turn effects
- Detects round completion (when index wraps to 0)

---

## Category 3: Game State Management (4 Methods)

### 8. joinGame
**Description**: Create new player session
**RPC Method**: `joinGame`
**Requires Session**: No
**Parameters**:
```json
{
  "player_name": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "session_id": "string"
}
```
**Constraints**:
- player_name required and non-empty
- Generates unique UUID session_id
- Creates PlayerSession and adds to world state

---

### 9. leaveGame
**Description**: End game session and cleanup
**RPC Method**: `leaveGame`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string"
}
```
**Response**:
```json
{
  "success": boolean
}
```
**Constraints**:
- Closes WebSocket connection if present
- Closes message channel
- Removes player from world state objects
- Removes session from sessions map

---

### 10. getGameState
**Description**: Retrieve current game state for session
**RPC Method**: `getGameState`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string"
}
```
**Response**:
```json
{
  "player": {
    "position": {"x": number, "y": number},
    "stats": {
      "hp": number,
      "max_hp": number,
      "level": number,
      "experience": number
    },
    "effects": [],
    "inventory": [],
    "spells": [],
    "experience": number
  },
  "world": {
    "visible_objects": [],
    "current_time": "string",
    "combat_state": {
      "active_combatants": ["string"],
      "round_count": number,
      "combat_zone": {"x": number, "y": number},
      "status_effects": {
        "[key:string]": [/* Effect objects */]
      }
    }
  }
}
```
**Constraints**:
- Returns comprehensive game state snapshot
- Includes player position, stats, inventory, spells
- Includes visible world objects and current time
- Combat state only present if in combat

---

### 11. createCharacter
**Description**: Create new player character with customization
**RPC Method**: `createCharacter`
**Requires Session**: No
**Parameters**:
```json
{
  "name": "string",
  "class": "fighter|mage|cleric|thief|ranger|paladin",
  "attribute_method": "roll|pointbuy|standard|custom",
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
**Response**:
```json
{
  "success": boolean,
  "session_id": "string",
  "character": { /* character data */ },
  "player": { /* player data */ },
  "errors": ["string"],
  "warnings": ["string"],
  "creation_time": "string",
  "generated_stats": { /* stats by method */ },
  "starting_items": [{ /* items */ }]
}
```
**Constraints**:
- Character class must be valid
- Attribute method determines stat generation
- custom_attributes required if method is 'custom'
- Default gold varies by class (fighter=100, mage=50, etc.)
- Creates new session automatically

---

## Category 4: Equipment Management (3 Methods)

### 12. equipItem
**Description**: Equip item to equipment slot
**RPC Method**: `equipItem`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "item_id": "string",
  "slot": "head|neck|chest|hands|rings|legs|feet|weapon_main|weapon_off|main_hand|off_hand"
}
```
**Response**:
```json
{
  "success": boolean,
  "message": "string",
  "equipped_item": { /* item object */ },
  "previous_item": { /* item object */ } or null
}
```
**Constraints**:
- Slot names support both weapon_main/main_hand and weapon_off/off_hand
- Item must exist in inventory
- Replaces previous item in slot
- Returns previous item if any

---

### 13. unequipItem
**Description**: Remove equipped item and return to inventory
**RPC Method**: `unequipItem`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "slot": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "message": "string",
  "unequipped_item": { /* item object */ }
}
```
**Constraints**:
- Item must be equipped in specified slot
- Item returned to player inventory

---

### 14. getEquipment
**Description**: Get all currently equipped items
**RPC Method**: `getEquipment`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "equipment": {
    "slot_name": { /* item object */ },
    ...
  },
  "total_weight": number,
  "equipment_bonuses": { /* stat bonuses */ }
}
```
**Constraints**:
- Lists all equipped items by slot
- Calculates total weight
- Includes equipment bonuses to stats

---

## Category 5: Quest Management (8 Methods)

### 15. startQuest
**Description**: Begin a quest for player
**RPC Method**: `startQuest`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "quest": { /* quest object */ }
}
```
**Response**:
```json
{
  "success": boolean,
  "quest_id": "string",
  "message": "Quest started successfully"
}
```
**Constraints**:
- Quest must be valid
- Player cannot have same quest already

---

### 16. completeQuest
**Description**: Complete active quest and process rewards
**RPC Method**: `completeQuest`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "quest_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "quest_id": "string",
  "rewards": [
    {
      "type": "exp|gold|item",
      "value": number,
      "item_id": "string"
    }
  ],
  "message": "Quest completed successfully"
}
```
**Constraints**:
- Quest must be active
- Objectives must be fulfilled
- Supports multiple reward types: exp, gold, item

---

### 17. failQuest
**Description**: Mark quest as failed
**RPC Method**: `failQuest`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "quest_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "quest_id": "string",
  "message": "Quest failed successfully"
}
```
**Constraints**:
- Quest must not already be completed/failed

---

### 18. updateObjective
**Description**: Update quest objective progress
**RPC Method**: `updateObjective`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "quest_id": "string",
  "objective_index": number,
  "progress": number
}
```
**Response**:
```json
{
  "success": boolean,
  "quest_id": "string",
  "objective_index": number,
  "progress": number,
  "message": "Quest objective updated successfully"
}
```
**Constraints**:
- objective_index is 0-based
- Progress value validated against requirements

---

### 19. getQuest
**Description**: Retrieve specific quest details
**RPC Method**: `getQuest`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "quest_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "quest": { /* quest object with all details */ }
}
```
**Constraints**:
- Quest must exist in player's quest log

---

### 20. getActiveQuests
**Description**: Get all active quests for player
**RPC Method**: `getActiveQuests`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "active_quests": [{ /* quest objects */ }],
  "count": number
}
```
**Constraints**:
- Returns only in-progress quests

---

### 21. getCompletedQuests
**Description**: Get all completed quests for player
**RPC Method**: `getCompletedQuests`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "completed_quests": [{ /* quest objects */ }],
  "count": number
}
```
**Constraints**:
- Returns only successfully completed quests

---

### 22. getQuestLog
**Description**: Get complete quest log (all quests)
**RPC Method**: `getQuestLog`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "quest_log": [{ /* quest objects */ }],
  "count": number
}
```
**Constraints**:
- Returns all quests regardless of status

---

## Category 6: Spell System (5 Methods)

### 23. getSpell
**Description**: Get spell by ID from database
**RPC Method**: `getSpell`
**Requires Session**: No
**Read-only**: Yes
**Parameters**:
```json
{
  "spell_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "spell": {
    "id": "string",
    "name": "string",
    "level": number,
    "school": "string",
    "casting_time": "string",
    "range": "string",
    "duration": "string",
    "description": "string"
  }
}
```
**Constraints**:
- spell_id required and non-empty
- Returns null if not found

---

### 24. getSpellsByLevel
**Description**: Get all spells of specific level
**RPC Method**: `getSpellsByLevel`
**Requires Session**: No
**Read-only**: Yes
**Parameters**:
```json
{
  "level": number
}
```
**Response**:
```json
{
  "success": boolean,
  "spells": [{ /* spell objects */ }],
  "count": number,
  "level": number
}
```
**Constraints**:
- Level must be >= 0 (0 for cantrips)

---

### 25. getSpellsBySchool
**Description**: Get all spells of magic school
**RPC Method**: `getSpellsBySchool`
**Requires Session**: No
**Read-only**: Yes
**Parameters**:
```json
{
  "school": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "spells": [{ /* spell objects */ }],
  "count": number,
  "school": "string"
}
```
**Constraints**:
- School name required and non-empty
- Examples: Evocation, Illusion, Divination, etc.

---

### 26. getAllSpells
**Description**: Get all spells in database
**RPC Method**: `getAllSpells`
**Requires Session**: No
**Read-only**: Yes
**Parameters**: (none)
```json
{}
```
**Response**:
```json
{
  "success": boolean,
  "spells": [{ /* all spell objects */ }],
  "count": number,
  "by_level": {
    "0": number,
    "1": number,
    ...
  }
}
```

---

### 27. searchSpells
**Description**: Search spells by name, description, keywords
**RPC Method**: `searchSpells`
**Requires Session**: No
**Read-only**: Yes
**Parameters**:
```json
{
  "query": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "spells": [{ /* matching spell objects */ }],
  "count": number,
  "query": "string"
}
```
**Constraints**:
- Query required and non-empty
- Searches spell name, description, keywords

---

## Category 7: Spatial Queries (3 Methods)

### 28. getObjectsInRange
**Description**: Get objects within rectangular area
**RPC Method**: `getObjectsInRange`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "min_x": number,
  "min_y": number,
  "max_x": number,
  "max_y": number
}
```
**Response**:
```json
{
  "success": boolean,
  "objects": [{ /* world objects */ }],
  "count": number
}
```
**Constraints**:
- Returns all objects within bounding box
- Supports efficient spatial indexing

---

### 29. getObjectsInRadius
**Description**: Get objects within circular area
**RPC Method**: `getObjectsInRadius`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "center_x": number,
  "center_y": number,
  "radius": number
}
```
**Response**:
```json
{
  "success": boolean,
  "objects": [{ /* world objects */ }],
  "count": number
}
```
**Constraints**:
- Radius must be >= 0
- Euclidean distance calculation

---

### 30. getNearestObjects
**Description**: Get k nearest objects to position
**RPC Method**: `getNearestObjects`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "center_x": number,
  "center_y": number,
  "k": number
}
```
**Response**:
```json
{
  "success": boolean,
  "objects": [{ /* k nearest objects */ }],
  "count": number
}
```
**Constraints**:
- k >= 1
- Returns up to k objects, fewer if less exist
- Sorted by distance (nearest first)

---

## Category 8: Procedural Content Generation (8 Methods)

### 31. generateContent
**Description**: Generate procedural content on demand
**RPC Method**: `generateContent`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "content_type": "terrain|items|levels|quests",
  "location_id": "string",
  "difficulty": number,
  "constraints": { /* content-specific constraints */ }
}
```
**Response**:
```json
{
  "success": boolean,
  "content_type": "string",
  "location_id": "string",
  "content": { /* generated content */ },
  "difficulty": number
}
```
**Constraints**:
- content_type required (terrain, items, levels, quests)
- location_id required
- difficulty default 5, range 1-20

---

### 32. regenerateTerrain
**Description**: Generate or regenerate terrain for area
**RPC Method**: `regenerateTerrain`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "location_id": "string",
  "width": number,
  "height": number,
  "biome_type": "string",
  "density": number,
  "water_level": number,
  "connectivity": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "location_id": "string",
  "terrain": { /* terrain map */ },
  "width": number,
  "height": number,
  "biome_type": "string"
}
```
**Defaults**:
- width: 50 (min 10)
- height: 50 (min 10)
- biome_type: "forest"
- density: 0.5 (0-1)
- water_level: 0.3 (0-1)
- connectivity: "moderate"

---

### 33. generateItems
**Description**: Generate items for location
**RPC Method**: `generateItems`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "location_id": "string",
  "count": number,
  "min_rarity": "string",
  "max_rarity": "string",
  "player_level": number,
  "item_types": ["string"]
}
```
**Response**:
```json
{
  "success": boolean,
  "location_id": "string",
  "items": [{ /* item objects */ }],
  "count": number
}
```
**Defaults**:
- count: 3 (min 1)
- min_rarity: "common"
- max_rarity: "rare"
- player_level: 5 (min 1)

---

### 34. generateLevel
**Description**: Generate complete dungeon level
**RPC Method**: `generateLevel`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "width": number,
  "height": number,
  "room_count": number,
  "theme": "string",
  "difficulty": number,
  "corridor_style": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "level": { /* level map/structure */ },
  "width": number,
  "height": number,
  "room_count": number,
  "theme": "string",
  "difficulty": number,
  "corridor_style": "string"
}
```
**Defaults**:
- width: 50 (min 20)
- height: 50 (min 20)
- room_count: 8 (min 1)
- theme: "classic"
- difficulty: 5 (min 1)
- corridor_style: "straight"

---

### 35. generateQuest
**Description**: Generate procedural quest
**RPC Method**: `generateQuest`
**Requires Session**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "quest_type": "string",
  "difficulty": number,
  "min_objectives": number,
  "max_objectives": number,
  "reward_tier": "string",
  "narrative_type": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "quest": { /* quest object */ },
  "quest_type": "string",
  "difficulty": number,
  "min_objectives": number,
  "max_objectives": number,
  "reward_tier": "string",
  "narrative_type": "string"
}
```
**Defaults**:
- quest_type: "fetch"
- difficulty: 5 (min 1)
- min_objectives: 1 (min 1)
- max_objectives: 3 (min 1)
- reward_tier: "common"
- narrative_type: "linear"

---

### 36. getPCGStats
**Description**: Get PCG system statistics
**RPC Method**: `getPCGStats`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string"
}
```
**Response**:
```json
{
  "success": boolean,
  "stats": {
    "total_generations": number,
    "successful_generations": number,
    "failed_generations": number,
    "average_generation_time": number,
    "content_counts": {
      "terrain": number,
      "items": number,
      "levels": number,
      "quests": number
    }
  }
}
```

---

### 37. validateContent
**Description**: Validate generated content
**RPC Method**: `validateContent`
**Requires Session**: Yes
**Read-only**: Yes
**Parameters**:
```json
{
  "session_id": "string",
  "content_type": "string",
  "content": { /* content to validate */ },
  "strict": boolean
}
```
**Response**:
```json
{
  "success": boolean,
  "valid": boolean,
  "errors": ["string"],
  "warnings": ["string"],
  "content_type": "string",
  "strict": boolean
}
```
**Constraints**:
- content_type required
- content required
- strict mode enforces all validations

---

## Error Codes and Validation

### JSON-RPC Error Codes
- `-32700`: Parse error
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

### Common Validation Rules

**Session Validation**:
- session_id must be non-empty string
- Session must exist in sessions map
- Player must be associated with session

**Direction Validation**:
- Accepts: "north", "south", "east", "west", "n", "s", "e", "w"
- Case-insensitive

**Position Validation**:
- X and Y coordinates within world bounds
- Integer values required

**Equipment Slots**:
- Valid slots: head, neck, chest, hands, rings, legs, feet, weapon_main, weapon_off
- Alternative names: main_hand (weapon_main), off_hand (weapon_off)

**Action Costs During Combat**:
- Move: ActionCostMove (1 AP)
- Attack: ActionCostAttack (1 AP)
- Spell: ActionCostSpell (1 AP)

**Difficulty Ranges**:
- PCG: 1-20 (default 5)
- Min/Max constraints enforced per method

---

## Data Types and Enums

### CharacterClass
- fighter
- mage
- cleric
- thief
- ranger
- paladin

### AttributeMethod
- roll (4d6 drop lowest)
- pointbuy (standard point buy system)
- standard (array: 15, 14, 13, 12, 10, 8)
- custom (provided values)

### EffectType
- poison
- stun
- heal
- buff
- debuff
- damage

### ContentType
- terrain
- items
- levels
- quests

### Rarity Tier
- common
- uncommon
- rare
- very_rare
- legendary

---

## Common Response Patterns

### Success Response
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "success": true,
    "message": "Operation completed successfully",
    "data": { /* operation-specific data */ }
  }
}
```

### Error Response
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "details": "session_id is required"
    }
  }
}
```

---

## Performance Considerations

1. **Spatial Queries**: Use appropriate query type
   - getObjectsInRange: Efficient for rectangular areas
   - getObjectsInRadius: Better for circular searches
   - getNearestObjects: Optimal for k-NN searches

2. **PCG Generation**: Expensive operations
   - generateContent: Most flexible, slowest
   - regenerateTerrain: Moderate cost
   - generateItems: Low cost
   - generateLevel: High cost
   - generateQuest: Moderate cost

3. **Read-only Methods**: Can be called frequently
   - All spell queries
   - All spatial queries
   - All quest getters

4. **State-modifying Methods**: Rate limit as needed
   - Combat actions
   - Equipment changes
   - PCG generation

---

## Summary Statistics

**Total Methods**: 38
- Character Actions: 5
- Combat Management: 2
- Game State: 4
- Equipment: 3
- Quest Management: 8
- Spell System: 5
- Spatial Queries: 3
- PCG: 8

**Methods Requiring Session**: 32
**Read-only Methods**: 18
**Methods Supporting Constraints**: 8 (PCG methods)

