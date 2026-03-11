# Gold Box RPG JSON-RPC API - Quick Reference

## All 38 RPC Methods - Quick Lookup Table

### Category 1: Character Actions (5)
| # | Method | Session | Combat | Params | Response |
|---|--------|---------|--------|--------|----------|
| 1 | `move` | ✓ | - | session_id, direction | success, position |
| 2 | `attack` | ✓ | - | session_id, target_id, weapon_id | success, damage |
| 3 | `castSpell` | ✓ | - | session_id, spell_id, target_id, position | success, spell_id |
| 4 | `useItem` | ✓ | - | session_id, item_id, target_id | success, effect |
| 5 | `applyEffect` | ✓ | - | session_id, effect_type, target_id, magnitude, duration | success, effect_id |

### Category 2: Combat Management (2)
| # | Method | Session | Combat | Params | Response |
|---|--------|---------|--------|--------|----------|
| 6 | `startCombat` | ✓ | - | session_id, participant_ids | success, initiative, first_turn |
| 7 | `endTurn` | ✓ | ✓ | session_id | success, next_turn |

### Category 3: Game State (4)
| # | Method | Session | Combat | Params | Response |
|---|--------|---------|--------|--------|----------|
| 8 | `joinGame` | - | - | player_name | success, session_id |
| 9 | `leaveGame` | ✓ | - | session_id | success |
| 10 | `getGameState` | ✓ | - | session_id | success, player, world |
| 11 | `createCharacter` | - | - | name, class, attribute_method | success, session_id, character |

### Category 4: Equipment (3)
| # | Method | Session | Combat | Params | Response |
|---|--------|---------|--------|--------|----------|
| 12 | `equipItem` | ✓ | - | session_id, item_id, slot | success, equipped_item, previous_item |
| 13 | `unequipItem` | ✓ | - | session_id, slot | success, unequipped_item |
| 14 | `getEquipment` | ✓ | - | session_id | success, equipment, total_weight, bonuses |

### Category 5: Quest Management (8)
| # | Method | Session | Combat | Params | Response |
|---|--------|---------|--------|--------|----------|
| 15 | `startQuest` | ✓ | - | session_id, quest | success, quest_id |
| 16 | `completeQuest` | ✓ | - | session_id, quest_id | success, quest_id, rewards |
| 17 | `failQuest` | ✓ | - | session_id, quest_id | success, quest_id |
| 18 | `updateObjective` | ✓ | - | session_id, quest_id, objective_index, progress | success, quest_id, objective_index, progress |
| 19 | `getQuest` | ✓ | - | session_id, quest_id | success, quest |
| 20 | `getActiveQuests` | ✓ | - | session_id | success, active_quests, count |
| 21 | `getCompletedQuests` | ✓ | - | session_id | success, completed_quests, count |
| 22 | `getQuestLog` | ✓ | - | session_id | success, quest_log, count |

### Category 6: Spell System (5)
| # | Method | Session | Combat | Params | Response |
|---|--------|---------|--------|--------|----------|
| 23 | `getSpell` | - | - | spell_id | success, spell |
| 24 | `getSpellsByLevel` | - | - | level | success, spells, count |
| 25 | `getSpellsBySchool` | - | - | school | success, spells, count |
| 26 | `getAllSpells` | - | - | (none) | success, spells, count, by_level |
| 27 | `searchSpells` | - | - | query | success, spells, count |

### Category 7: Spatial Queries (3)
| # | Method | Session | Combat | Params | Response |
|---|--------|---------|--------|--------|----------|
| 28 | `getObjectsInRange` | ✓ | - | session_id, min_x, min_y, max_x, max_y | success, objects, count |
| 29 | `getObjectsInRadius` | ✓ | - | session_id, center_x, center_y, radius | success, objects, count |
| 30 | `getNearestObjects` | ✓ | - | session_id, center_x, center_y, k | success, objects, count |

### Category 8: Procedural Content Generation (8)
| # | Method | Session | Combat | Params | Response |
|---|--------|---------|--------|--------|----------|
| 31 | `generateContent` | ✓ | - | session_id, content_type, location_id, difficulty | success, content_type, location_id, content, difficulty |
| 32 | `regenerateTerrain` | ✓ | - | session_id, location_id, width, height, biome_type | success, location_id, terrain, width, height |
| 33 | `generateItems` | ✓ | - | session_id, location_id, count, min_rarity, max_rarity | success, location_id, items, count |
| 34 | `generateLevel` | ✓ | - | session_id, width, height, room_count, theme, difficulty | success, level, width, height, room_count, theme, difficulty |
| 35 | `generateQuest` | ✓ | - | session_id, quest_type, difficulty | success, quest, quest_type, difficulty |
| 36 | `getPCGStats` | ✓ | - | session_id | success, stats |
| 37 | `validateContent` | ✓ | - | session_id, content_type, content, strict | success, valid, errors, warnings |

---

## Parameter Reference by Type

### String Parameters
```
session_id, direction, player_name, name, item_id, slot, quest_id
spell_id, school, query, location_id, content_type, biome_type
corridor_style, quest_type, narrative_type, reward_tier, target_id
weapon_id, effect_type, effect_id
```

### Numeric Parameters
```
difficulty (1-20)          // PCG methods
player_level (≥ 1)         // Item generation
count (≥ 1)                // Item/object counts
k (≥ 1)                    // Nearest objects limit
width, height (≥ 10-20)    // Terrain/level dimensions
room_count (≥ 1)           // Level generation
level (≥ 0)                // Spell level
radius (≥ 0)               // Spatial query
magnitude, duration        // Effects
min_x, min_y, max_x, max_y // Range queries
center_x, center_y         // Radius/nearest queries
objective_index (≥ 0)      // Quest objectives
progress                   // Objective progress
```

### Enumerated Parameters
```
direction: north, south, east, west, n, s, e, w
class: fighter, mage, cleric, thief, ranger, paladin
attribute_method: roll, pointbuy, standard, custom
content_type: terrain, items, levels, quests
effect_type: poison, stun, heal, buff, debuff, damage
biome_type: forest, mountain, desert, swamp, etc.
rarity: common, uncommon, rare, very_rare, legendary
quest_type: fetch, defend, explore, discover, etc.
theme: classic, dungeon, castle, underground, etc.

Equipment Slots:
  head, neck, chest, hands, rings, legs, feet
  weapon_main, weapon_off (or main_hand, off_hand)
```

---

## Key Patterns

### Request Format
```json
{
  "jsonrpc": "2.0",
  "method": "methodName",
  "params": { /* parameters */ },
  "id": 1
}
```

### Response Format (Success)
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "success": true,
    ...
  }
}
```

### Response Format (Error)
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": { "details": "..." }
  }
}
```

---

## Common Validations

| Type | Validation | Default |
|------|-----------|---------|
| session_id | Required (non-empty), must exist | N/A |
| direction | Must match enum | N/A |
| difficulty | 1-20 | 5 |
| width/height | ≥ 10-20 depending on method | 50 |
| level | ≥ 0 | N/A |
| radius | ≥ 0 | N/A |
| k | ≥ 1 | N/A |
| count | ≥ 1 | 3 |
| player_name | Non-empty string | N/A |
| quest_type | Valid enum | "fetch" |
| reward_tier | Valid enum | "common" |

---

## Statistics

```
Total Methods: 38
├─ Character Actions: 5
├─ Combat Management: 2
├─ Game State: 4
├─ Equipment: 3
├─ Quest Management: 8
├─ Spell System: 5
├─ Spatial Queries: 3
└─ PCG: 8

Methods by Session Requirement:
├─ Require Session: 32
└─ No Session: 6

Methods by IO Type:
├─ Read-only: 18
└─ State-modifying: 20

Methods with Constraints:
└─ PCG Methods: 8 (difficulty, dimensions, counts)
```

---

## Common Error Messages

```
"invalid session" - session_id not found
"invalid movement parameters" - Direction/position invalid
"not in combat" - Combat-only method called outside combat
"not your turn" - Attempted action when not player's turn
"insufficient action points" - Not enough AP for action
"spell not found" - Spell ID doesn't exist in database
"player does not know this spell" - Spell not in player's spellbook
"item not found in inventory" - Item ID doesn't exist in inventory
"invalid equipment slot" - Slot name not recognized
"quest not found" - Quest ID not in player's quest log
"session has no associated player" - Session missing player data
"invalid target" - Target ID doesn't exist in world
```

---

## Method Call Examples

### Minimal Examples
```bash
# Move character
curl -X POST /rpc -d '{"jsonrpc":"2.0","method":"move","params":{"session_id":"ABC","direction":"north"},"id":1}'

# Get game state
curl -X POST /rpc -d '{"jsonrpc":"2.0","method":"getGameState","params":{"session_id":"ABC"},"id":1}'

# Complete quest
curl -X POST /rpc -d '{"jsonrpc":"2.0","method":"completeQuest","params":{"session_id":"ABC","quest_id":"Q1"},"id":1}'

# Get spell
curl -X POST /rpc -d '{"jsonrpc":"2.0","method":"getSpell","params":{"spell_id":"fireball"},"id":1}'

# Generate terrain
curl -X POST /rpc -d '{"jsonrpc":"2.0","method":"regenerateTerrain","params":{"session_id":"ABC","location_id":"L1","width":50,"height":50},"id":1}'
```

