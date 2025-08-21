# AI-Generated Visual Resources Plan

## Project Overview

GoldBox RPG Engine is a modern Go-based turn-based RPG game inspired by the classic SSI Gold Box series. The project requires comprehensive visual assets to support its tile-based 2D rendering system with canvas-based frontend. The game features classic RPG mechanics including six character classes (Fighter, Mage, Cleric, Thief, Ranger, Paladin), tactical grid-based combat, spell effects, and equipment systems.

**Visual Requirements Analysis:**
- **Game Type**: 2D tile-based turn-based RPG with top-down perspective
- **Rendering System**: Canvas 2D with 128x128 pixel tile-based rendering (high resolution)
- **Art Style**: Classic RPG aesthetic suitable for detailed artwork and high-resolution sprites
- **Technical Constraints**: 128x128 pixel base tiles, PNG format with transparency support
- **Existing Infrastructure**: Canvas layer system (terrain, objects, effects) with sprite sheet support

## Asset Categories

### Character Sprites
- **Purpose**: Represent the six playable character classes and NPCs in the game world
- **Specifications**: 
  - Base size: 128x128 pixels per sprite
  - Format: PNG with alpha transparency
  - Sprite sheet layout with multiple frames per character
  - 4-directional facing sprites (North, East, South, West)
- **AI Generation Notes**: 
  - Style: "High-resolution fantasy RPG artwork, 128x128 resolution, top-down perspective, detailed character design"
  - Prompt template: "Fantasy [CLASS] character sprite, detailed artwork, 128x128 pixels, top-down view, clear silhouette, medieval fantasy setting, facing [DIRECTION]"
  - Color palette: Rich and detailed colors suitable for high-resolution fantasy setting
- **Integration Points**: 
  - `pkg/game/character.go` (Character struct)
  - `pkg/game/classes.go` (CharacterClass enum)
  - `src/rendering/GameRenderer.ts` (character rendering)
  - File path: `/web/static/assets/sprites/characters.png`

### Terrain Tiles
- **Purpose**: Environment tiles for map construction including floors, walls, doors, and special terrain
- **Specifications**:
  - Base size: 128x128 pixels per tile
  - Format: PNG with optional transparency
  - Seamless tiling capability for floor types
  - Support for TileType enum: floor, wall, door, water, lava, pit, stairs
- **AI Generation Notes**:
  - Style: "High-resolution medieval fantasy dungeon tileset, detailed textures, seamless tiling, realistic stone textures"
  - Prompt template: "Fantasy dungeon [TILE_TYPE] tile, 128x128 detailed artwork, top-down view, seamless edges, medieval stone architecture"
  - Consistency: Unified lighting direction (top-left), consistent color temperature
- **Integration Points**:
  - `pkg/game/map.go` (MapTile struct)
  - `pkg/game/constants.go` (TileType constants)
  - `src/rendering/GameRenderer.ts` (terrain rendering)
  - File path: `/web/static/assets/sprites/terrain.png`

### Item Icons
- **Purpose**: Visual representation of weapons, armor, consumables, and treasure
- **Specifications**:
  - Base size: 128x128 pixels for world items, 96x96 for inventory icons
  - Format: PNG with alpha transparency
  - Clear iconography readable at various sizes
  - Categories: weapons, armor, consumables, quest items, treasure
- **AI Generation Notes**:
  - Style: "High-resolution fantasy RPG item icons, detailed artwork, clear symbolism, inventory style"
  - Prompt template: "Fantasy [ITEM_TYPE] icon, 128x128 detailed artwork, clear design, medieval fantasy style, [ITEM_NAME]"
  - Emphasis: High contrast, distinctive silhouettes, consistent perspective
- **Integration Points**:
  - `pkg/game/item.go` (Item struct)
  - `pkg/game/equipment.go` (equipment system)
  - `data/items/items.yaml` (item definitions)
  - File path: `/web/static/assets/sprites/items.png`

### Spell Effects
- **Purpose**: Visual effects for spell casting, combat actions, and status effects
- **Specifications**:
  - Variable sizes: 128x128 base, up to 256x256 for area effects
  - Format: PNG with alpha transparency for overlay effects
  - Animated sequences supported via sprite sheets
  - Effect types: damage over time, healing, elemental damage, status conditions
- **AI Generation Notes**:
  - Style: "High-resolution magical spell effects, detailed glowing particles, energy bursts, fantasy game VFX"
  - Prompt template: "Fantasy [SPELL_TYPE] spell effect, detailed magical energy, glowing particles, [ELEMENT_TYPE] magic, transparent background, 128x128 resolution"
  - Animation: Frame-by-frame sequences for dynamic effects
- **Integration Points**:
  - `pkg/game/effects.go` (effect system)
  - `data/spells/` YAML files (spell definitions)
  - `src/rendering/GameRenderer.ts` (effects layer)
  - File path: `/web/static/assets/sprites/effects.png`

### UI Components
- **Purpose**: User interface elements including buttons, panels, frames, and HUD elements
- **Specifications**:
  - Variable sizes based on component needs (minimum 128x128 for buttons)
  - Format: PNG with alpha transparency
  - 9-slice compatible borders for scalable panels
  - Medieval fantasy theme matching game aesthetic
- **AI Generation Notes**:
  - Style: "High-resolution medieval fantasy UI elements, ornate borders, detailed scroll-like textures, gold accents"
  - Prompt template: "Fantasy game UI [COMPONENT], detailed medieval scroll texture, ornate borders, gold trim, parchment background, high resolution"
  - Consistency: Warm color palette (browns, golds), readable typography support
- **Integration Points**:
  - `web/index.html` (UI structure)
  - `web/static/css/ui.css` (UI styling)
  - `src/ui/` components
  - File path: `/web/static/assets/sprites/ui.png`

### Status Icons
- **Purpose**: Small icons representing character states, effects, and conditions
- **Specifications**:
  - Base size: 64x64 pixels for status bars
  - Format: PNG with alpha transparency
  - High contrast for visibility at various sizes
  - Categories: buffs, debuffs, conditions, class abilities
- **AI Generation Notes**:
  - Style: "High-resolution RPG status effect icons, detailed symbols, high contrast, clear meaning"
  - Prompt template: "Fantasy RPG [EFFECT_TYPE] status icon, 64x64 pixels, detailed symbol, high contrast, clear design"
  - Symbolism: Universally recognizable gaming iconography
- **Integration Points**:
  - `pkg/game/effects.go` (status effect system)
  - `src/ui/GameUI.ts` (status display)
  - File path: `/web/static/assets/sprites/status.png`

## Directory Structure

```
web/static/assets/
├── sprites/
│   ├── characters.png          # Character class sprites and NPCs
│   ├── terrain.png            # Environment tiles and terrain
│   ├── items.png              # Weapons, armor, consumables
│   ├── effects.png            # Spell effects and animations
│   ├── ui.png                 # Interface components and frames
│   ├── status.png             # Status effect icons
│   ├── portraits/             # Character portraits (future)
│   │   ├── fighter.png
│   │   ├── mage.png
│   │   └── ...
│   └── variations/            # Theme and resolution variants
│       ├── hd/                # High-resolution variants
│       └── themes/            # Alternative art styles
├── audio/                     # Sound effects and music (future)
├── fonts/                     # Custom fonts (future)
└── data/                      # Asset metadata and configurations
    ├── sprite_mappings.json   # Sprite coordinate mappings
    ├── animation_data.json    # Animation frame sequences
    └── asset_manifest.json    # Complete asset inventory
```

## Code Integration

### Asset Path Management

```go
// pkg/game/assets.go
package game

import "embed"

// AssetConfig defines the structure for managing game assets
type AssetConfig struct {
    BasePath    string                    `yaml:"base_path"`
    SpritePaths map[string]string        `yaml:"sprite_paths"`
    Characters  map[CharacterClass]CharacterAssets `yaml:"characters"`
    Items       map[string]ItemAssets    `yaml:"items"`
    Effects     map[EffectType]EffectAssets `yaml:"effects"`
}

// CharacterAssets defines sprite coordinates for character classes
type CharacterAssets struct {
    IdleSprites     [4]SpriteCoord `yaml:"idle_sprites"`      // N, E, S, W
    WalkSprites     [4][]SpriteCoord `yaml:"walk_sprites"`    // Animation frames
    Portrait        SpriteCoord    `yaml:"portrait"`
    ClassIcon       SpriteCoord    `yaml:"class_icon"`
}

// ItemAssets defines sprite information for items
type ItemAssets struct {
    WorldSprite     SpriteCoord `yaml:"world_sprite"`    // When dropped in world
    InventoryIcon   SpriteCoord `yaml:"inventory_icon"`  // In inventory/equipment
    EquippedOverlay SpriteCoord `yaml:"equipped_overlay,omitempty"` // Visual equipment overlay
}

// EffectAssets defines visual effect sprite data
type EffectAssets struct {
    AnimationFrames []SpriteCoord `yaml:"animation_frames"`
    Duration        int           `yaml:"duration_ms"`
    LoopCount       int           `yaml:"loop_count"`      // -1 for infinite
    BlendMode       string        `yaml:"blend_mode"`      // "normal", "additive", "multiply"
}

// SpriteCoord represents a sprite's position in a sprite sheet
type SpriteCoord struct {
    X      int `yaml:"x"`      // X coordinate in sprite sheet
    Y      int `yaml:"y"`      // Y coordinate in sprite sheet
    Width  int `yaml:"width"`  // Sprite width (default 128)
    Height int `yaml:"height"` // Sprite height (default 128)
}

// Constants for asset paths
const (
    AssetBasePath = "/web/static/assets"
    SpritePath    = AssetBasePath + "/sprites"
)

//go:embed web/static/assets/sprites/*
var embeddedSprites embed.FS

// AssetManager provides centralized asset management
type AssetManager struct {
    config     *AssetConfig
    spriteMaps map[string]map[string]SpriteCoord
    useEmbedded bool
}

// NewAssetManager creates a new asset manager instance
func NewAssetManager(configPath string, useEmbedded bool) (*AssetManager, error) {
    // Implementation details...
    return &AssetManager{
        useEmbedded: useEmbedded,
        spriteMaps:  make(map[string]map[string]SpriteCoord),
    }, nil
}

// GetCharacterSprite returns sprite coordinates for a character class and facing
func (am *AssetManager) GetCharacterSprite(class CharacterClass, facing Direction, frame int) (SpriteCoord, error) {
    // Implementation details...
    return SpriteCoord{}, nil
}

// GetItemSprite returns sprite coordinates for an item
func (am *AssetManager) GetItemSprite(itemID string, context string) (SpriteCoord, error) {
    // Implementation details...
    return SpriteCoord{}, nil
}
```

### Asset Loading Pattern

```typescript
// src/core/AssetLoader.ts
export interface AssetManifest {
  version: string;
  sprites: Record<string, SpriteSheetInfo>;
  animations: Record<string, AnimationData>;
  metadata: AssetMetadata;
}

export interface SpriteSheetInfo {
  path: string;
  tileSize: number; // 128 for high resolution
  columns: number;
  rows: number;
  mapping: Record<string, SpriteCoord>;
}

export interface AnimationData {
  frames: SpriteCoord[];
  duration: number;
  loop: boolean;
  blendMode?: string;
}

export interface SpriteCoord {
  x: number;
  y: number;
  width?: number;
  height?: number;
}

export class AssetLoader extends BaseComponent {
  private manifest: AssetManifest | null = null;
  private loadedSheets = new Map<string, HTMLImageElement>();
  private loadingPromises = new Map<string, Promise<HTMLImageElement>>();

  async loadManifest(path: string = '/static/assets/data/asset_manifest.json'): Promise<void> {
    try {
      const response = await fetch(path);
      if (!response.ok) {
        throw new Error(`Failed to load asset manifest: ${response.statusText}`);
      }
      this.manifest = await response.json();
      this.componentLogger.info('Asset manifest loaded', this.manifest);
    } catch (error) {
      this.componentLogger.error('Failed to load asset manifest', error);
      throw error;
    }
  }

  async loadSpriteSheet(sheetName: string): Promise<HTMLImageElement> {
    if (this.loadedSheets.has(sheetName)) {
      return this.loadedSheets.get(sheetName)!;
    }

    if (this.loadingPromises.has(sheetName)) {
      return this.loadingPromises.get(sheetName)!;
    }

    if (!this.manifest) {
      throw new Error('Asset manifest not loaded');
    }

    const sheetInfo = this.manifest.sprites[sheetName];
    if (!sheetInfo) {
      throw new Error(`Sprite sheet not found: ${sheetName}`);
    }

    const loadPromise = new Promise<HTMLImageElement>((resolve, reject) => {
      const img = new Image();
      img.onload = () => {
        this.loadedSheets.set(sheetName, img);
        this.loadingPromises.delete(sheetName);
        resolve(img);
      };
      img.onerror = () => {
        this.loadingPromises.delete(sheetName);
        reject(new Error(`Failed to load sprite sheet: ${sheetInfo.path}`));
      };
      img.src = sheetInfo.path;
    });

    this.loadingPromises.set(sheetName, loadPromise);
    return loadPromise;
  }

  getSpriteCoord(sheetName: string, spriteId: string): SpriteCoord | null {
    if (!this.manifest) return null;
    
    const sheetInfo = this.manifest.sprites[sheetName];
    if (!sheetInfo) return null;
    
    return sheetInfo.mapping[spriteId] || null;
  }

  getAnimationData(animationId: string): AnimationData | null {
    if (!this.manifest) return null;
    return this.manifest.animations[animationId] || null;
  }
}

// Global asset loader instance
export const assetLoader = new AssetLoader();
```

### Enhanced Renderer Integration

```typescript
// src/rendering/GameRenderer.ts (enhanced)
export class GameRenderer extends BaseComponent {
  private assetLoader: AssetLoader;
  private readonly tileSize: number = 128; // Quadrupled from 32 to 128
  
  constructor() {
    super({ name: 'GameRenderer' });
    this.assetLoader = assetLoader;
    // ... existing initialization
  }

  async loadSprites(): Promise<void> {
    this.componentLogger.group('Loading sprite assets via AssetLoader');
    
    try {
      await this.assetLoader.loadManifest();
      
      const requiredSheets = ['characters', 'terrain', 'items', 'effects', 'ui', 'status'];
      const loadPromises = requiredSheets.map(sheet => this.assetLoader.loadSpriteSheet(sheet));
      
      await Promise.all(loadPromises);
      this.componentLogger.info(`Loaded ${requiredSheets.length} sprite sheets`);
      
      this.emit('spritesLoaded', { count: requiredSheets.length });
    } catch (error) {
      this.componentLogger.error('Failed to load sprites via AssetLoader', error);
      this.useFallbackSprites();
    } finally {
      this.componentLogger.groupEnd();
    }
  }

  private drawCharacter(character: Character, screenX: number, screenY: number): void {
    const spriteCoord = this.assetLoader.getSpriteCoord('characters', 
      `${character.class}_${character.facing}_idle`);
    
    if (spriteCoord) {
      this.drawSpriteFromCoord(this.contexts.objects, 'characters', spriteCoord, screenX, screenY);
    } else {
      this.componentLogger.warn(`Sprite not found for character: ${character.class}`);
    }
  }

  private drawSpriteFromCoord(
    ctx: CanvasRenderingContext2D,
    sheetName: string,
    coord: SpriteCoord,
    destX: number,
    destY: number
  ): void {
    const sheet = this.assetLoader.loadedSheets.get(sheetName);
    if (!sheet) {
      this.componentLogger.warn(`Sprite sheet not loaded: ${sheetName}`);
      return;
    }

    try {
      ctx.drawImage(
        sheet,
        coord.x,
        coord.y,
        coord.width || this.tileSize, // 128px
        coord.height || this.tileSize, // 128px
        destX,
        destY,
        coord.width || this.tileSize, // 128px
        coord.height || this.tileSize // 128px
      );
    } catch (error) {
      this.componentLogger.error(`Failed to draw sprite from ${sheetName}`, error);
    }
  }
}
```

## Extensibility Considerations

### Asset Variant System
The asset system is designed to support multiple variants and themes:

- **Resolution Variants**: Support for different pixel densities (1x, 2x, 4x scaling)
- **Theme Variants**: Multiple art styles (classic, modern, minimalist) stored in theme subdirectories
- **Seasonal Content**: Holiday or event-specific asset overrides
- **Modding Support**: External asset pack loading via JSON manifests
- **Localization**: Region-specific assets for cultural adaptation

### Hot-Reload Development System
```typescript
// Development-only hot-reload capability
if (process.env.NODE_ENV === 'development') {
  const assetWatcher = new AssetWatcher();
  assetWatcher.on('assetChanged', async (assetPath: string) => {
    await assetLoader.reloadAsset(assetPath);
    gameRenderer.invalidateCache();
  });
}
```

### Asset Validation Framework
```go
// pkg/game/validation/assets.go
type AssetValidator struct {
    requiredAssets map[string]AssetRequirement
}

type AssetRequirement struct {
    Path        string
    MinSize     Point
    MaxSize     Point
    Format      string
    HasAlpha    bool
    Required    bool
}

func (av *AssetValidator) ValidateAssetPack(packPath string) error {
    // Validate all assets meet requirements
    // Check for missing assets
    // Verify image formats and dimensions
    // Test sprite sheet coordinate mappings
    return nil
}
```

### Performance Optimization
- **Sprite Atlasing**: Automatic packing of individual sprites into optimized sprite sheets
- **Lazy Loading**: Load assets only when needed for current game area
- **Caching Strategy**: Smart cache invalidation and preloading based on player location
- **Compression**: WebP format support with PNG fallback for better loading times

## Generation Workflow

### 1. Asset Planning Phase
- Review game requirements and create detailed asset specifications
- Define consistent art style guidelines and color palettes
- Create reference images and style guides for AI generation
- Establish naming conventions and file organization

### 2. AI Generation Process
- **Batch Generation**: Use consistent prompts across related assets
- **Style Transfer**: Maintain visual consistency using reference images
- **Iteration Cycles**: Generate multiple variants and select best options
- **Quality Control**: Review each asset against specifications

### 3. Post-Processing Pipeline
- **Optimization**: Reduce file sizes while maintaining quality
- **Format Standardization**: Convert to appropriate formats (PNG, WebP)
- **Sprite Sheet Assembly**: Pack individual sprites into optimized sheets
- **Metadata Generation**: Create sprite coordinate mappings and animation data

### 4. Integration Testing
- **Asset Loading Tests**: Verify all assets load correctly
- **Visual Regression Tests**: Ensure new assets don't break existing functionality
- **Performance Benchmarks**: Test loading times and memory usage
- **Cross-Browser Compatibility**: Verify assets work across different browsers

### 5. Version Control and Deployment
- **Asset Versioning**: Tag asset releases with semantic versioning
- **CDN Integration**: Deploy optimized assets to content delivery network
- **Rollback Capability**: Maintain previous asset versions for quick rollback
- **Documentation**: Update asset manifests and developer documentation

## Quality Standards

### Technical Requirements
- **File Formats**: PNG preferred (transparency support), WebP for optimization
- **Dimensions**: Strict adherence to 128x128 pixel base grid (quadrupled resolution)
- **Color Depth**: 24-bit color with 8-bit alpha channel
- **File Size**: Individual sprites <40KB, sprite sheets <2MB (adjusted for higher resolution)
- **Compression**: Lossless PNG compression, optimized WebP variants

### Visual Consistency
- **Art Style**: Cohesive detailed artwork or high-resolution sprite style throughout
- **Color Palette**: Rich, harmonious color scheme with detailed shading and highlights
- **Lighting**: Unified light source direction (top-left) across all assets
- **Perspective**: Consistent top-down orthographic projection
- **Character Proportions**: Standardized character sizing and proportions for 128x128 resolution

### Accessibility
- **Contrast Ratios**: Meet WCAG AA standards for visual elements
- **Color Blind Support**: Avoid relying solely on color for important information
- **Readability**: Ensure icons and text are clear at target display sizes
- **Alternative Text**: Provide descriptive metadata for screen readers

### Performance Criteria
- **Loading Time**: Complete asset loading under 5 seconds on average connections (adjusted for larger files)
- **Memory Usage**: Total asset memory footprint under 200MB (quadrupled for higher resolution)
- **Cache Efficiency**: Optimal browser caching with appropriate HTTP headers
- **Rendering Performance**: Maintain 60fps during normal gameplay with higher resolution assets

### Asset Validation Checklist
- [ ] Correct dimensions and format specifications met
- [ ] Visual style consistent with established art direction
- [ ] Transparent backgrounds properly implemented where required
- [ ] Sprite sheet coordinates correctly mapped in metadata
- [ ] Animation frames properly sequenced and timed
- [ ] File sizes optimized without quality loss
- [ ] Cross-browser compatibility verified
- [ ] Integration with game systems tested and functional

This comprehensive asset plan provides a scalable foundation for integrating AI-generated visual content into the GoldBox RPG Engine while maintaining technical excellence and visual consistency throughout the development process.
