# GoldBox RPG Asset Generation Pipeline

This directory contains the asset generation pipeline for the GoldBox RPG Engine using `asset-generator`. The pipeline generates 2D game assets from the existing game metadata and procedural content system without requiring real-time AI content generation.

## Pipeline Structure

- `characters.yaml` - Character class portraits and variations
- `spells.yaml` - Spell effect visuals and icons
- `items.yaml` - Equipment and item assets from PCG templates
- `environments.yaml` - Terrain and environment assets
- `ui.yaml` - User interface elements and icons
- `config/` - Asset generator configuration

## Quick Start

```bash
# Install asset-generator (see main project README for setup)
curl -sSL https://github.com/opd-ai/asset-generator/releases/latest/download/asset-generator-linux-amd64 -o asset-generator
chmod +x asset-generator
sudo mv asset-generator /usr/local/bin/

# Configure for your SwarmUI instance
asset-generator config set api-url http://localhost:7801

# Generate all game assets
make assets

# Generate specific asset categories
make assets-characters
make assets-spells
make assets-items
make assets-environments
```

## Asset Categories

### Characters
- Class portraits (Fighter, Mage, Cleric, Thief, Ranger, Paladin)
- Character variations (race, gender, equipment)
- NPC archetypes from PCG system

### Spells
- School-based spell effect visuals
- Cantrip animations
- Spell level progression effects

### Items
- Weapons from PCG templates
- Armor sets
- Consumables and potions
- Treasure and artifacts

### Environments
- Terrain tiles (floor, wall, door, water, lava, pit, stairs)
- Biome variations
- Dungeon environments

### UI Elements
- Action buttons
- Status indicators
- Inventory icons
- Combat interface elements

## Integration with Game Data

The asset pipeline reads from:
- `/data/spells/` - Spell definitions for generating spell effects
- `/data/pcg/` - PCG templates for item and environment generation
- Game constants from `/pkg/game/constants.go` for character classes and types

Generated assets are organized to match the game's data structure for easy integration.