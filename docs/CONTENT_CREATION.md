# Content Creation Guide

This guide explains how to create and modify game content for the GoldBox RPG Engine.

## Overview

The GoldBox RPG Engine uses YAML files for game content configuration. Content types include:
- **Spells** - Magic abilities for characters
- **Items** - Weapons, armor, and consumables
- **Quests** - Story-driven objectives
- **PCG Templates** - Procedural content generation parameters

## CLI Tools

Three command-line tools are available for content creation:

### Quest Builder (`cmd/quest-builder`)

Interactive quest creation tool with 5 templates:
- **Fetch** - Retrieve item quests
- **Kill** - Combat-focused quests
- **Escort** - Protection missions
- **Explore** - Discovery quests
- **Puzzle** - Problem-solving challenges

```bash
go run cmd/quest-builder/main.go --help
go run cmd/quest-builder/main.go --template fetch --output data/quests/my_quest.yaml
```

### Map Editor (`cmd/map-editor`)

ASCII-based tile map editor with 4 templates:
- **Dungeon** - Underground areas
- **Outdoor** - Open terrain
- **Cave** - Natural caverns
- **Town** - Settlement maps

```bash
go run cmd/map-editor/main.go --help
go run cmd/map-editor/main.go --template dungeon --width 40 --height 30
```

### Content Creator (`cmd/content-creator`)

Template-driven spell and item creation:

```bash
go run cmd/content-creator/main.go --help
go run cmd/content-creator/main.go --type spell --output data/spells/custom.yaml
go run cmd/content-creator/main.go --type item --output data/items/custom.yaml
```

## YAML Schema Reference

### Spell Schema

Location: `data/spells/*.yaml`

```yaml
spells:
  - spell_id: "unique_spell_id"      # Unique identifier (snake_case)
    spell_name: "Display Name"        # Human-readable name
    spell_level: 0                    # 0 = cantrip, 1-9 = spell levels
    spell_school: 5                   # School enum (see below)
    spell_range: 30                   # Range in feet
    spell_duration: 60                # Duration in rounds
    damage_type: "fire"               # Optional: fire, cold, lightning, etc.
    damage_dice: "2d6"                # Optional: damage dice
    healing_dice: "1d8"               # Optional: healing dice
    spell_components:                 # Component requirements
      - 0                             # 0 = verbal, 1 = somatic, 2 = material
      - 1
    spell_description: "Description"  # Spell effect description
```

**Spell Schools:**
| Value | School |
|-------|--------|
| 0 | Abjuration |
| 1 | Conjuration |
| 2 | Divination |
| 3 | Enchantment |
| 4 | Evocation |
| 5 | Illusion |
| 6 | Necromancy |
| 7 | Transmutation |
| 8 | Universal |

### Item Schema

Location: `data/items/items.yaml`

```yaml
- item_id: "unique_item_id"           # Unique identifier (snake_case)
  name: "Display Name"                # Human-readable name
  description: "Item description"      # Flavor text
  type: "weapon"                      # weapon, armor, consumable, misc
  value: 100                          # Gold value
  weight: 5                           # Weight in pounds
  
  # Weapon-specific fields
  weapon_type: "melee"                # melee, ranged, thrown
  damage: "1d8"                       # Damage dice
  range: 5                            # Range in feet (for ranged)
  
  # Armor-specific fields
  armor_class: 16                     # Base AC
  ac_bonus: 2                         # AC bonus (shields)
  
  # Consumable-specific fields
  uses: 3                             # Number of uses
  effect: "healing"                   # Effect type
```

### Quest Schema

Location: `data/quests/*.yaml`

```yaml
quests:
  - quest_id: "unique_quest_id"       # Unique identifier
    quest_name: "Quest Name"           # Display name
    quest_description: "Description"   # Story text
    quest_type: "fetch"               # fetch, kill, escort, explore, puzzle
    difficulty: 5                     # 1-20 difficulty rating
    objectives:
      - id: "obj_1"
        type: "collect"               # collect, kill, reach, interact
        target: "item_id"             # Target entity ID
        quantity: 3                   # Required count
        description: "Collect 3 gems"
    rewards:
      experience: 500
      gold: 100
      items:
        - "reward_item_id"
    prerequisites:
      - "previous_quest_id"           # Quests that must be completed first
```

### PCG Parameters

Location: `data/pcg/*.yaml`

```yaml
terrain:
  seed: 12345                         # Random seed for reproducibility
  width: 100                          # Map width in tiles
  height: 100                         # Map height in tiles
  biome: "forest"                     # forest, mountain, desert, swamp, cave
  density: 0.6                        # Feature density (0.0-1.0)
  features:
    - "water"
    - "trees"
    - "roads"

dungeon:
  seed: 54321
  min_rooms: 5
  max_rooms: 15
  room_min_size: 5
  room_max_size: 15
  corridor_style: "winding"           # straight, winding, maze
  theme: "classic"                    # classic, horror, natural, mechanical
```

## Example Content

### Example: Creating a Fire Spell

```yaml
# data/spells/fire_spells.yaml
spells:
  - spell_id: "fireball"
    spell_name: "Fireball"
    spell_level: 3
    spell_school: 4                   # Evocation
    spell_range: 150
    spell_duration: 0                 # Instantaneous
    damage_type: "fire"
    damage_dice: "8d6"
    spell_components:
      - 0                             # Verbal
      - 1                             # Somatic
      - 2                             # Material (bat guano)
    spell_description: "A bright streak flashes from your pointing finger to a point you choose within range and then blossoms with a low roar into an explosion of flame."
```

### Example: Creating a Magical Sword

```yaml
# data/items/magic_weapons.yaml
- item_id: "flaming_sword"
  name: "Flaming Sword"
  description: "A steel longsword wreathed in magical flames."
  type: "weapon"
  weapon_type: "melee"
  damage: "1d8+1d6"                   # Base + fire damage
  value: 2500
  weight: 3
  properties:
    magical: true
    damage_type: "fire"
    bonus_attack: 1
    bonus_damage: 1
```

### Example: Creating a Fetch Quest

```yaml
# data/quests/guild_missions.yaml
quests:
  - quest_id: "herbs_for_healer"
    quest_name: "Herbs for the Healer"
    quest_description: "The village healer needs rare herbs from the forest."
    quest_type: "fetch"
    difficulty: 3
    objectives:
      - id: "collect_moonpetal"
        type: "collect"
        target: "moonpetal_herb"
        quantity: 5
        description: "Gather 5 Moonpetal flowers from the forest clearing"
      - id: "return_to_healer"
        type: "interact"
        target: "npc_healer"
        description: "Return the herbs to the village healer"
    rewards:
      experience: 200
      gold: 50
      items:
        - "healing_potion"
```

## Validation

Content is validated when loaded by the game. Use the validator demo to check your content:

```bash
go run cmd/validator-demo/main.go --file data/spells/custom.yaml
```

Common validation errors:
- **Missing required fields** - Check schema for required fields
- **Invalid enum values** - Use correct numeric values for enums
- **Duplicate IDs** - Ensure all IDs are unique within content type

## File Organization

```
data/
├── items/
│   └── items.yaml           # All items in single file
├── spells/
│   ├── cantrips.yaml        # Level 0 spells
│   ├── level1.yaml          # Level 1 spells
│   └── ...                  # Levels 2-9
├── quests/
│   └── *.yaml               # Quest definitions
└── pcg/
    └── *.yaml               # PCG templates
```

## Best Practices

1. **Use descriptive IDs** - IDs should clearly identify the content (e.g., `healing_potion_greater` not `item_042`)

2. **Balance values** - Reference existing content for appropriate damage, value, and difficulty numbers

3. **Provide descriptions** - All content should have meaningful descriptions for player immersion

4. **Test incrementally** - Add content in small batches and test after each addition

5. **Backup before editing** - Keep copies of working YAML files before making changes

6. **Follow naming conventions** - Use snake_case for IDs, Title Case for display names

## Troubleshooting

### YAML Syntax Errors
Use a YAML linter or online validator to check syntax. Common issues:
- Incorrect indentation (use 2 spaces, not tabs)
- Missing colons after keys
- Unquoted special characters

### Content Not Loading
- Check file is in correct `data/` subdirectory
- Verify file extension is `.yaml` not `.yml`
- Check server logs for loading errors

### Balance Issues
- Reference `pkg/game/combat.go` for damage calculations
- Check `pkg/game/spell_manager.go` for spell effect processing
- Review existing content in `data/` for comparable values

## Additional Resources

- [ASSET_INTEGRATION.md](../ASSET_INTEGRATION.md) - Visual asset pipeline
- [pkg/README-RPC.md](../pkg/README-RPC.md) - API documentation
- [pkg/pcg/README.md](../pkg/pcg/README.md) - PCG system details
