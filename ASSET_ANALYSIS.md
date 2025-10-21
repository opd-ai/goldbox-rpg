# GoldBox RPG Engine - Asset Analysis Report

**Generated:** 2025-10-21  
**Repository:** https://github.com/opd-ai/goldbox-rpg  
**Game Type:** Classic RPG (SSI Gold Box inspired)

## Executive Summary

This document provides a comprehensive analysis of the GoldBox RPG Engine codebase to identify all visual asset requirements and create a production-ready asset generation pipeline. The engine is a modern Go-based framework for turn-based RPG games inspired by the classic SSI Gold Box series, featuring character management, combat systems, and world interactions.

## Technology Stack

### Backend
- **Language:** Go 1.23.0+ with toolchain 1.23.2
- **Framework:** Native Go HTTP server with JSON-RPC 2.0 protocol
- **Real-time Communication:** Gorilla WebSocket v1.5.3
- **Data Format:** YAML v3.0.1 for game data configuration
- **Logging:** Sirupsen Logrus v1.9.3

### Frontend
- **Language:** TypeScript with ES2020 target
- **Build Tool:** ESBuild bundling
- **Rendering:** Canvas-based game rendering
- **Communication:** EventEmitter pattern for state management

### Asset Loading System
- **Location:** `src/rendering/GameRenderer.ts`
- **Method:** Asynchronous image loading with Promise-based error handling
- **Format Support:** PNG (primary), JPG, SVG
- **Fallback System:** Automatic fallback sprites when loading fails

## Asset Discovery

### Current Asset References

From codebase analysis (`src/rendering/GameRenderer.ts:142-147`):

```typescript
const spriteUrls = {
  terrain: './static/assets/sprites/terrain.png',
  characters: './static/assets/sprites/characters.png',
  effects: './static/assets/sprites/effects.png',
  ui: './static/assets/sprites/ui.png',
};
```

### Existing Assets

Located in `web/static/assets/sprites/`:
- `terrain.png` - Terrain tiles and map elements
- `terrain.jpg` - Alternative terrain format
- `characters.png` - Character sprites and portraits
- `effects.png` - Combat effects and spell visuals
- `ui.png` - User interface elements

## Asset Requirements by Category

### 1. Character Assets

**Source Analysis:**
- `pkg/game/character.go`: Character system with 6 attributes (Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma)
- `pkg/game/types.go`: CharacterClass enum (Fighter, Mage, Cleric, Thief, Ranger, Paladin)
- `data/spells/`: Spell data indicating magic users

**Required Assets:**

#### Character Portraits (Head and Shoulders)
- **Dimensions:** 128x128 pixels (estimated from UI mockups)
- **Format:** PNG with transparency
- **Categories:**
  - **Fighters:** Human Male/Female, Elf Male/Female, Dwarf Male/Female, Halfling Male/Female
  - **Mages:** Human Male/Female, Elf Male/Female, Dwarf Male/Female, Halfling Male/Female
  - **Clerics:** Human Male/Female, Elf Male/Female, Dwarf Male/Female, Halfling Male/Female
  - **Thieves:** Human Male/Female, Elf Male/Female, Dwarf Male/Female, Halfling Male/Female
  - **Rangers:** Human Male/Female, Elf Male/Female, Dwarf Male/Female, Halfling Male/Female
  - **Paladins:** Human Male/Female, Elf Male/Female, Dwarf Male/Female, Halfling Male/Female

**Total Character Portraits:** 48 (6 classes × 4 races × 2 genders)

#### Battle Sprites (Combat View)
- **Dimensions:** 64x64 pixels
- **Format:** PNG with transparency
- **Perspective:** Front-facing for tactical combat
- **Same categorization as portraits:** 48 battle sprites

### 2. Monster Assets

**Source Analysis:**
- `pkg/pcg/levels/rooms.go:127`: References elemental, sprite, wisp
- Classic D&D-inspired RPG implies standard monster types
- `pkg/game/combat.go`: Combat system supporting multiple enemies

**Required Monster Sprites:**

#### Common Monsters (64x64 pixels, PNG)
- **Undead:** Skeleton Warrior, Zombie, Ghoul, Wight, Vampire, Lich
- **Humanoids:** Goblin, Hobgoblin, Orc, Ogre, Troll
- **Dragons:** Red Dragon, Black Dragon, Blue Dragon, Green Dragon, White Dragon
- **Magical Creatures:** Elemental (Fire/Water/Earth/Air), Sprite, Wisp, Wraith
- **Beasts:** Dire Wolf, Giant Spider, Giant Rat, Bear, Wyvern
- **Demons:** Imp, Demon, Balor
- **Constructs:** Golem (Stone/Iron/Clay)

**Total Monster Sprites:** ~35 unique creatures

### 3. Item & Equipment Assets

**Source Analysis:**
- `data/items/items.yaml`: Comprehensive item list
- Equipment types: weapons, armor, shields, consumables, equipment

**Required Item Icons:**

#### Weapons (48x48 pixels, PNG)
- **Melee:** Sword, Dagger, Staff, Battle Axe, Mace, Warhammer, Flail, Spear, Halberd
- **Ranged:** Bow, Longbow, Crossbow, Sling
- **Magic:** Mage Staff (with crystal), Cleric's Mace (holy)

#### Armor (48x48 pixels, PNG)
- **Light:** Leather Armor, Padded Armor, Studded Leather
- **Medium:** Chain Mail, Scale Mail, Hide Armor
- **Heavy:** Plate Armor, Full Plate, Banded Mail
- **Accessories:** Shield (Small/Large), Helmet, Gauntlets, Boots

#### Consumables (48x48 pixels, PNG)
- **Potions:** Health Potion (Red), Mana Potion (Blue), Antidote (Green), Strength Potion (Orange)
- **Scrolls:** Spell Scroll, Map Scroll, Teleport Scroll
- **Food:** Trail Rations, Bread, Cheese, Dried Meat

#### Magic Items (48x48 pixels, PNG)
- **Jewelry:** Magic Ring, Amulet, Crown
- **Special:** Crystal Ball, Holy Symbol, Spellbook
- **Keys:** Iron Key, Gold Key, Skeleton Key

#### Equipment (48x48 pixels, PNG)
- Rope, Torch, Lantern, Backpack, Waterskin, Bedroll

**Total Item Icons:** ~60 unique items

### 4. Terrain & Environment Assets

**Source Analysis:**
- `pkg/game/tile.go`: Tile system with sprite coordinates
- `pkg/pcg/terrain/`: Procedural terrain generation
- `pkg/pcg/dungeon.go`: Dungeon generation with walls, floors, doors

**Required Terrain Tiles (32x32 pixels, PNG, Tileable):**

#### Dungeon Elements
- **Floors:** Stone Floor, Wooden Floor, Dirt Floor, Marble Floor
- **Walls:** Stone Wall (N/S/E/W/Corners), Brick Wall, Cave Wall
- **Doors:** Wooden Door (Closed/Open), Iron Door, Secret Door
- **Special:** Chest, Barrel, Crate, Furniture, Torch Holder, Statue

#### Outdoor Terrain
- **Ground:** Grass, Dirt, Sand, Snow, Stone Path
- **Water:** Water Tile (Center/Edge/Corner variations)
- **Nature:** Tree, Bush, Rock, Flower
- **Structures:** Building Wall, Roof Tile, Bridge

#### Special Tiles
- **Interactive:** Trap (visible/hidden), Lever, Button, Stairs (Up/Down)
- **Effects:** Lava, Ice, Poison Pool, Magic Circle

**Total Terrain Tiles:** ~80 unique tiles (with variations)

### 5. Combat Effects Assets

**Source Analysis:**
- `data/spells/`: Spell system with cantrips and leveled spells
- `pkg/game/effects.go`: Effect system for combat conditions
- Damage types: Physical, Fire, Poison, Frost, Lightning

**Required Effect Sprites (64x64 pixels, PNG with transparency):**

#### Spell Effects
- **Attack Spells:** Fireball, Lightning Bolt, Magic Missile, Ice Shard, Acid Splash
- **Area Effects:** Flame Strike, Cone of Cold, Lightning Storm, Meteor Swarm
- **Buff Effects:** Shield, Bless, Haste, Fly, Stoneskin
- **Debuff Effects:** Curse, Slow, Web, Hold Person, Fear
- **Healing:** Healing Light, Cure Wounds, Mass Heal
- **Utility:** Light, Darkness, Detect Magic, Teleport Circle

#### Combat Animations (Frame sequences, 32x32 pixels)
- **Weapon Impacts:** Sword Slash, Arrow Hit, Blunt Impact
- **Damage Types:** Fire Burst, Ice Crystal, Poison Cloud, Lightning Strike, Physical Hit
- **Status Effects:** Burning (3 frames), Frozen (3 frames), Poisoned (3 frames), Stunned (3 frames)
- **Explosion:** Small/Medium/Large (each 4 frames)

**Total Effect Sprites:** ~50 unique effects, ~100 animation frames

### 6. UI Elements Assets

**Source Analysis:**
- `src/ui/GameUI.ts`: UI component system
- `src/rendering/GameRenderer.ts`: Canvas-based rendering
- JSON-RPC methods require UI feedback

**Required UI Assets:**

#### Buttons & Controls (Various sizes, PNG with transparency)
- **Buttons:** Normal, Hover, Active, Disabled (each state)
- **Sizes:** Small (64x32), Medium (128x48), Large (192x64)
- **Styles:** Primary, Secondary, Danger, Success

#### Panels & Frames (Scalable 9-slice, PNG)
- **Dialog Box:** Stone/Wood/Metal frame with decorative corners
- **Inventory Panel:** Grid-based container with slots
- **Character Sheet:** Ornate frame for character display
- **Combat Log:** Scrollable text container

#### Icons (32x32 pixels, PNG with transparency)
- **Stats:** Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma
- **Combat:** Attack, Defense, Magic, Movement, Action Points
- **Status:** Health (Heart), Mana (Star), Experience (Trophy)
- **Actions:** Move, Attack, Cast Spell, Use Item, Wait, End Turn
- **Conditions:** Stunned, Burning, Frozen, Poisoned, Blessed, Cursed

#### Indicators (Various sizes, PNG)
- **Health Bar:** Empty/Partial/Full (Red gradient)
- **Mana Bar:** Empty/Partial/Full (Blue gradient)
- **Progress Bar:** Generic progress indicator
- **Selection:** Selection Circle, Targeting Reticle, Movement Range

#### Decorative Elements (PNG with transparency)
- **Borders:** Corner pieces, Edge pieces for medieval frames
- **Dividers:** Horizontal/Vertical dividers for panels
- **Backgrounds:** Parchment texture, Stone texture, Wood texture
- **Typography:** Game title logo, Section headers

**Total UI Elements:** ~100 assets (including variations)

## Visual Style Requirements

### Art Direction

**Style:** Fantasy RPG pixel art with modern sensibilities
- **Inspiration:** SSI Gold Box games, modern indie RPGs
- **Pixel Resolution:** Medium-detail pixel art (not ultra-low res)
- **Color Palette:** Rich, vibrant colors with medieval fantasy theme
- **Mood:** Heroic fantasy with dungeons and dragons aesthetic

### Color Palette

**Primary Colors:**
- **Stone/Dungeon:** #5A5A5A (dark gray), #8B8B8B (medium gray), #C0C0C0 (light stone)
- **Fantasy Blue:** #2E5090 (deep blue), #4A7DBF (medium blue), #7AA8E0 (light blue)
- **Medieval Red:** #8B2E2E (deep red), #BF4A4A (medium red), #E07A7A (light red)
- **Nature Green:** #2E5A2E (forest green), #4A8B4A (grass green), #7ABF7A (light green)
- **Gold/Yellow:** #8B7A2E (bronze), #BFA54A (gold), #E0C57A (light gold)

**Secondary Colors:**
- **Magic Purple:** #5A2E8B (deep purple), #7A4ABF (medium), #A67AE0 (light)
- **Fire Orange:** #8B4A2E (deep), #BF6A4A (medium), #E0987A (light)
- **Ice Cyan:** #2E8B8B (deep), #4ABFBF (medium), #7AE0E0 (light)

### Technical Specifications

#### Image Formats
- **Primary:** PNG-24 with alpha channel
- **Fallback:** PNG-8 for simple graphics
- **Not Used:** JPG (except for specific backgrounds without transparency)

#### Dimensions & Constraints
- **Character Portraits:** 128x128 pixels
- **Battle Sprites:** 64x64 pixels
- **Monster Sprites:** 64x64 pixels
- **Item Icons:** 48x48 pixels
- **Terrain Tiles:** 32x32 pixels (must be tileable)
- **Effect Sprites:** 64x64 pixels
- **UI Icons:** 32x32 pixels
- **Buttons:** 64x32 (small), 128x48 (medium), 192x64 (large)

#### File Size & Performance
- **Target:** < 50KB per sprite asset
- **Maximum:** 200KB for complex sprites
- **Optimization:** PNG compression with tools like pngquant
- **Mobile Consideration:** Assets should work on devices with limited memory

### Naming Conventions

#### Character Assets
```
portrait_{class}_{race}_{gender}.png
battle_{class}_{race}_{gender}.png

Examples:
portrait_fighter_human_male.png
battle_mage_elf_female.png
```

#### Monster Assets
```
monster_{type}_{variant}.png

Examples:
monster_skeleton_warrior.png
monster_dragon_red.png
```

#### Item Assets
```
item_{category}_{name}.png

Examples:
item_weapon_sword.png
item_armor_leather.png
item_potion_health.png
```

#### Terrain Assets
```
tile_{type}_{variant}.png

Examples:
tile_floor_stone.png
tile_wall_brick.png
tile_door_wood_open.png
```

#### Effect Assets
```
effect_{type}_{name}_{frame}.png

Examples:
effect_spell_fireball.png
effect_damage_fire_01.png
effect_status_burning_03.png
```

#### UI Assets
```
ui_{category}_{element}_{state}.png

Examples:
ui_button_primary_normal.png
ui_icon_strength.png
ui_panel_inventory.png
```

## Directory Structure

### Current Structure
```
web/static/assets/sprites/
├── README.md
├── characters.png
├── effects.png
├── terrain.png
├── terrain.jpg
└── ui.png
```

### Proposed Enhanced Structure
```
web/static/assets/
├── sprites/
│   ├── characters/
│   │   ├── portraits/
│   │   │   ├── fighters/
│   │   │   │   ├── portrait_fighter_human_male.png
│   │   │   │   ├── portrait_fighter_human_female.png
│   │   │   │   └── ...
│   │   │   ├── mages/
│   │   │   ├── clerics/
│   │   │   ├── thieves/
│   │   │   ├── rangers/
│   │   │   └── paladins/
│   │   └── battle/
│   │       ├── fighters/
│   │       ├── mages/
│   │       └── ...
│   ├── monsters/
│   │   ├── undead/
│   │   ├── humanoids/
│   │   ├── dragons/
│   │   ├── magical/
│   │   ├── beasts/
│   │   └── demons/
│   ├── items/
│   │   ├── weapons/
│   │   ├── armor/
│   │   ├── consumables/
│   │   ├── magic/
│   │   └── equipment/
│   ├── terrain/
│   │   ├── dungeon/
│   │   ├── outdoor/
│   │   └── special/
│   ├── effects/
│   │   ├── spells/
│   │   ├── combat/
│   │   └── status/
│   └── ui/
│       ├── buttons/
│       ├── panels/
│       ├── icons/
│       ├── indicators/
│       └── decorative/
└── spritesheets/
    ├── characters_sheet.png
    ├── monsters_sheet.png
    ├── items_sheet.png
    ├── terrain_sheet.png
    ├── effects_sheet.png
    └── ui_sheet.png
```

## Asset Loading Implementation

### Current Implementation
Location: `src/rendering/GameRenderer.ts`

```typescript
// Loads 4 sprite sheets
private async loadSprites(): Promise<void> {
  const spriteUrls = {
    terrain: './static/assets/sprites/terrain.png',
    characters: './static/assets/sprites/characters.png',
    effects: './static/assets/sprites/effects.png',
    ui: './static/assets/sprites/ui.png',
  };
  // Async loading with Promise.all
}
```

### Integration Requirements

1. **Sprite Sheet Approach:** Current system uses consolidated sprite sheets
2. **Individual Assets:** New structure supports both sprite sheets and individual assets
3. **Fallback System:** Must maintain fallback sprites when loading fails
4. **Progressive Loading:** Can load sprite sheets first, individual assets on demand

## Asset Count Summary

| Category | Count | Total Pixels | Estimated Size |
|----------|-------|--------------|----------------|
| Character Portraits | 48 | 128x128 | ~1.5 MB |
| Battle Sprites | 48 | 64x64 | ~400 KB |
| Monsters | 35 | 64x64 | ~300 KB |
| Items | 60 | 48x48 | ~350 KB |
| Terrain | 80 | 32x32 | ~300 KB |
| Effects | 50 + 100 frames | 64x64 & 32x32 | ~600 KB |
| UI Elements | 100 | Various | ~800 KB |
| **TOTAL** | **~521 assets** | - | **~4.25 MB** |

## Configuration Files Analysis

### Spell Data (`data/spells/*.yaml`)
- **Cantrips:** Light, Mage Hand, Prestidigitation
- **Level 1:** Various attack and utility spells
- **Level 2:** More powerful spells
- **Effect Sprites Needed:** Match spell names with visual effects

### Item Data (`data/items/items.yaml`)
- **11 Items Defined:** Sword, Bow, Dagger, Staff, Leather Armor, Chain Mail, Shield, Healing Potion, Rope, Torch, Trail Rations
- **Icon Requirements:** Each item needs corresponding icon

### PCG Templates (`data/pcg/*.yaml`)
- **Bootstrap Configuration:** World generation parameters
- **Item Templates:** Template-based item generation
- **Asset Implications:** Generated content needs visual representation

## Placeholder & TODO Analysis

**README Reference:** `web/static/assets/sprites/README.md`
> "Currently, this directory is prepared for game sprites and graphical assets that will be used by the TypeScript frontend rendering system."

**Planned Assets Listed:**
- Character sprites (player classes, NPCs)
- Item icons (weapons, armor, consumables)
- Terrain tiles (floor, walls, doors)
- Effect sprites (spells, combat effects)
- UI elements (buttons, panels, icons)

**Status:** All categories are in planning/preparation phase awaiting asset generation.

## Recommendations

### Priority 1 (Critical - Blocks Gameplay)
1. **Core Terrain Tiles** (Stone floor, walls, doors) - Enables dungeon exploration
2. **Basic Character Sprites** (One portrait per class) - Enables character representation
3. **Essential UI Elements** (Buttons, health bars, icons) - Enables interface interaction

### Priority 2 (High - Enhances Experience)
4. **All Character Portraits** (All class/race/gender combinations) - Character variety
5. **Common Monster Sprites** (10-15 most used enemies) - Combat variety
6. **Weapon & Armor Icons** (All equipment from items.yaml) - Inventory visualization

### Priority 3 (Medium - Adds Polish)
7. **Combat Effect Sprites** (Basic spell effects) - Visual feedback
8. **Complete Terrain Set** (All tile variations) - Environmental variety
9. **Additional UI Elements** (All interface components) - Professional polish

### Priority 4 (Low - Future Enhancement)
10. **All Monster Sprites** (Complete bestiary) - Maximum variety
11. **Animation Frames** (All combat animations) - Smooth transitions
12. **Decorative UI Elements** (Backgrounds, borders) - Aesthetic enhancement

## Technical Constraints

### Performance Considerations
- **Mobile Targets:** Assets must work on devices with 2GB RAM
- **Loading Time:** Initial load should be < 3 seconds on 4G connection
- **Memory Usage:** Total loaded sprites should stay under 50MB in memory
- **Sprite Sheets:** Consider sprite sheet consolidation for performance

### Browser Compatibility
- **PNG Support:** Universal across all modern browsers
- **Alpha Channel:** Required for character and effect sprites
- **Canvas Rendering:** Assets must work with HTML5 Canvas API
- **WebGL Future:** Consider WebGL compatibility for future enhancements

### Build System Integration
- **Existing Build:** `make build` compiles Go backend, npm builds frontend
- **Asset Pipeline:** Should integrate with existing Makefile
- **Deployment:** Docker container must include generated assets
- **CI/CD:** Assets should be versioned and cached in CI pipeline

## Code References

### Asset Loading Code
- **File:** `src/rendering/GameRenderer.ts`
- **Lines:** 140-199 (sprite loading system)
- **Method:** `loadSprites()`, `loadSpriteImage()`

### Tile System
- **File:** `pkg/game/tile.go`
- **Line 38:** `Sprite string` field for sprite identifier
- **Usage:** Maps tile types to sprite coordinates

### Map Rendering
- **File:** `pkg/game/map.go`
- **Lines 7-8:** `SpriteX` and `SpriteY` coordinates for tile sprites

### Terrain Generation
- **Files:** `pkg/pcg/terrain/*.go`
- **Usage:** Assigns sprite coordinates during procedural generation

## Next Steps

1. **Create Asset Generation Pipeline** (`game-assets.yaml`)
   - Define all asset categories hierarchically
   - Specify prompts for each asset
   - Set appropriate seed offsets
   - Include metadata for consistency

2. **Develop Generation Scripts**
   - `generate-all.sh` - Generate complete asset library
   - `generate-priority1.sh` - Generate critical assets first
   - `post-process.sh` - Resize, compress, optimize
   - `verify-assets.sh` - Validate generated assets

3. **Create Integration Guide** (`ASSET_INTEGRATION.md`)
   - Installation instructions for asset-generator tool
   - Pipeline execution procedures
   - Build system integration
   - Troubleshooting common issues

4. **Update Makefile**
   - Add `make assets` target
   - Add `make assets-preview` for dry-run
   - Add `make assets-clean` to reset
   - Integrate with existing build process

5. **Test & Validate**
   - Generate sample assets
   - Load in game renderer
   - Verify visual appearance
   - Check performance metrics
   - Iterate on prompts as needed

## Conclusion

The GoldBox RPG Engine requires approximately **521 visual assets** across 6 main categories (characters, monsters, items, terrain, effects, UI). The asset generation pipeline should produce pixel art style graphics in PNG format with specified dimensions per category. The proposed hierarchical structure organizes assets logically while maintaining compatibility with the existing sprite loading system in `GameRenderer.ts`.

**Total Estimated Pipeline Generation Time:** 4-6 hours for complete asset library (assuming 25-30 seconds per asset)

**Recommended Approach:** Incremental generation starting with Priority 1 (critical) assets, followed by Priority 2-4 as time and resources permit.

---

*This analysis serves as the foundation for creating the automated asset generation pipeline for the GoldBox RPG Engine.*
