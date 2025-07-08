/**
 * Game rendering system with TypeScript support
 * Handles canvas-based rendering of terrain, objects, and effects
 */

import { BaseComponent } from '../core/BaseComponent';
import type { 
  GameState as IGameState,
  GameObject,
  GameMap,
  TileType
} from '../types/GameTypes';
import type { 
  CanvasLayers,
  CanvasContexts,
  RenderOptions
} from '../types/UITypes';

export interface GameRendererEvents {
  spritesLoaded: { count: number };
  spriteLoadError: { sprite: string; error: Error };
  renderComplete: { frameTime: number };
  error: Error;
}

export interface Camera {
  x: number;
  y: number;
  zoom: number;
}

export interface SpriteInfo {
  readonly image: HTMLImageElement;
  readonly width: number;
  readonly height: number;
  readonly loaded: boolean;
}

export class GameRenderer extends BaseComponent {
  // Canvas elements and contexts
  private readonly canvasLayers: CanvasLayers;
  private readonly contexts: CanvasContexts;
  
  // Rendering properties
  private readonly tileSize: number = 32;
  private readonly sprites = new Map<string, SpriteInfo>();
  private readonly animations = new Map<string, any>(); // TODO: Define animation types
  
  // Camera system
  private readonly camera: Camera = {
    x: 0,
    y: 0,
    zoom: 1,
  };

  // Resize handling
  private readonly boundHandleResize: () => void;
  
  // Performance tracking
  private lastFrameTime: number = 0;
  private frameCount: number = 0;

  constructor() {
    super({ name: 'GameRenderer' });
    
    this.componentLogger.debug('Getting canvas elements');
    
    // Get canvas elements
    const terrainLayer = document.getElementById('terrain-layer') as HTMLCanvasElement;
    const objectLayer = document.getElementById('object-layer') as HTMLCanvasElement;
    const effectLayer = document.getElementById('effect-layer') as HTMLCanvasElement;

    if (!terrainLayer || !objectLayer || !effectLayer) {
      throw new Error('Canvas elements not found');
    }

    this.canvasLayers = {
      terrain: terrainLayer,
      objects: objectLayer,
      effects: effectLayer
    };

    this.componentLogger.info('Setting up canvas contexts');
    
    // Get 2D contexts
    const terrainCtx = terrainLayer.getContext('2d');
    const objectCtx = objectLayer.getContext('2d');
    const effectCtx = effectLayer.getContext('2d');

    if (!terrainCtx || !objectCtx || !effectCtx) {
      throw new Error('Canvas 2D rendering contexts not available');
    }

    this.contexts = {
      terrain: terrainCtx,
      objects: objectCtx,
      effects: effectCtx
    };

    // Check for WebGL support as fallback information
    this.checkWebGLSupport();

    this.componentLogger.info('Initializing core properties');

    // Set up resize handling
    this.boundHandleResize = this.handleResize.bind(this);
    window.addEventListener('resize', this.boundHandleResize);

    this.componentLogger.info('Performing initial resize');
    this.handleResize();
  }

  /**
   * Initialize the renderer
   */
  protected async onInitialize(): Promise<void> {
    try {
      await this.loadSprites();
      this.componentLogger.info('Renderer initialized successfully');
    } catch (error) {
      this.componentLogger.error('Failed to initialize renderer', error);
      throw error;
    }
  }

  /**
   * Clean up renderer resources
   */
  protected async onCleanup(): Promise<void> {
    window.removeEventListener('resize', this.boundHandleResize);
    this.sprites.clear();
    this.animations.clear();
    this.componentLogger.info('Renderer cleaned up');
  }

  /**
   * Load sprite images asynchronously
   */
  async loadSprites(): Promise<void> {
    this.componentLogger.group('Loading sprite assets');
    
    const spriteUrls = {
      terrain: './static/assets/sprites/terrain.png',
      characters: './static/assets/sprites/characters.png',
      effects: './static/assets/sprites/effects.png',
      ui: './static/assets/sprites/ui.png',
    };

    this.componentLogger.debug('Sprite URLs to load', spriteUrls);

    const loadPromises = Object.entries(spriteUrls).map(async ([key, url]) => {
      this.componentLogger.info(`Loading sprite "${key}" from ${url}`);
      
      try {
        const img = new Image();
        const spriteInfo = await this.loadSpriteImage(img, url);
        this.sprites.set(key, spriteInfo);
        this.componentLogger.info(`Successfully loaded sprite "${key}"`);
      } catch (error) {
        this.componentLogger.error(`Failed to load sprite "${key}" from ${url}`, error);
        const errorObj = error instanceof Error ? error : new Error(String(error));
        this.emit('spriteLoadError', { sprite: key, error: errorObj });
        throw new Error(`Failed to load sprite: ${url}`);
      }
    });

    try {
      await Promise.all(loadPromises);
      this.componentLogger.info(`Completed loading ${this.sprites.size} sprites`);
      this.emit('spritesLoaded', { count: this.sprites.size });
    } catch (error) {
      this.componentLogger.error('Failed to load all sprites', error);
      this.useFallbackSprites();
    } finally {
      this.componentLogger.groupEnd();
    }
  }

  /**
   * Load a single sprite image
   */
  private loadSpriteImage(img: HTMLImageElement, url: string): Promise<SpriteInfo> {
    return new Promise((resolve, reject) => {
      img.onload = () => {
        resolve({
          image: img,
          width: img.naturalWidth,
          height: img.naturalHeight,
          loaded: true
        });
      };
      
      img.onerror = () => {
        reject(new Error(`Failed to load sprite: ${url}`));
      };
      
      img.src = url;
    });
  }

  /**
   * Create fallback sprites when image loading fails
   */
  useFallbackSprites(): void {
    this.componentLogger.group('Creating fallback sprites');

    const fallbackSprites = {
      terrain: { color: '#8B4513', size: 32 }, // Brown for terrain
      characters: { color: '#FFD700', size: 32 }, // Gold for characters
      effects: { color: '#FF69B4', size: 32 }, // Pink for effects
      ui: { color: '#708090', size: 32 }, // Slate gray for UI
    };

    for (const [key, config] of Object.entries(fallbackSprites)) {
      this.componentLogger.info(`Creating fallback sprite for "${key}"`);

      const canvas = document.createElement('canvas');
      canvas.width = config.size;
      canvas.height = config.size;
      const ctx = canvas.getContext('2d');

      if (ctx) {
        // Draw a simple colored rectangle with border
        ctx.fillStyle = config.color;
        ctx.fillRect(0, 0, config.size, config.size);
        ctx.strokeStyle = '#000000';
        ctx.lineWidth = 2;
        ctx.strokeRect(0, 0, config.size, config.size);

        // Add a text label
        ctx.fillStyle = '#FFFFFF';
        ctx.font = '10px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(key.charAt(0).toUpperCase(), config.size / 2, config.size / 2 + 3);

        // Create a temporary image from canvas
        const img = new Image();
        img.src = canvas.toDataURL();
        
        this.sprites.set(key, {
          image: img,
          width: config.size,
          height: config.size,
          loaded: true
        });
      }
    }

    this.componentLogger.info(`Created ${Object.keys(fallbackSprites).length} fallback sprites`);
    this.componentLogger.groupEnd();
  }

  /**
   * Handle window resize events
   */
  private handleResize(): void {
    this.componentLogger.group('Resizing canvas layers');

    const container = document.getElementById('viewport-container');
    if (!container) {
      this.componentLogger.error('Viewport container not found');
      this.componentLogger.groupEnd();
      return;
    }

    const width = container.clientWidth;
    const height = container.clientHeight;
    this.componentLogger.debug('Container dimensions', { width, height });

    if (width === 0 || height === 0) {
      this.componentLogger.warn('Container has zero dimension');
    }

    // Resize all canvas layers
    Object.values(this.canvasLayers).forEach((canvas) => {
      if (canvas) {
        this.componentLogger.debug(`Resizing canvas to ${width}x${height}`);
        canvas.width = width;
        canvas.height = height;
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
      } else {
        this.componentLogger.error('Canvas layer is null');
      }
    });

    this.componentLogger.groupEnd();
  }

  /**
   * Clear all canvas layers
   */
  clearLayers(): void {
    this.componentLogger.group('Clearing all canvas layers');
    
    Object.entries(this.contexts).forEach(([name, ctx]) => {
      if (ctx) {
        this.componentLogger.debug(`Clearing ${name} canvas of size ${ctx.canvas.width}x${ctx.canvas.height}`);
        ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
      } else {
        this.componentLogger.error(`Missing ${name} context`);
      }
    });

    this.componentLogger.groupEnd();
  }

  /**
   * Main render method
   */
  render(gameState: IGameState, options: RenderOptions = {}): void {
    const startTime = performance.now();
    
    this.componentLogger.group('Rendering game state');
    this.componentLogger.debug('Game state', gameState);

    if (!gameState) {
      this.componentLogger.error('Game state is null or undefined');
      this.componentLogger.groupEnd();
      return;
    }

    if (!gameState.world) {
      this.componentLogger.warn('World data is missing from game state');
    }

    // Clear layers unless specified otherwise
    if (options.clearLayers !== false) {
      this.clearLayers();
    }

    // Render layers in order
    this.renderTerrain(gameState.world?.map);
    this.renderObjects(gameState.world?.objects);

    const frameTime = performance.now() - startTime;
    this.lastFrameTime = frameTime;
    this.frameCount++;

    this.componentLogger.debug(`Render completed in ${frameTime.toFixed(2)}ms`);
    this.emit('renderComplete', { frameTime });
    this.componentLogger.groupEnd();
  }

  /**
   * Render terrain tiles
   */
  private renderTerrain(map?: GameMap): void {
    this.componentLogger.group('Rendering terrain layer');
    this.componentLogger.debug('Map data', map);

    if (!map) {
      this.componentLogger.error('Map is null or undefined');
      this.componentLogger.groupEnd();
      return;
    }

    const viewportWidth = Math.ceil(this.canvasLayers.terrain.width / this.tileSize);
    const viewportHeight = Math.ceil(this.canvasLayers.terrain.height / this.tileSize);
    
    this.componentLogger.info('Viewport dimensions', { viewportWidth, viewportHeight });

    for (let y = 0; y < viewportHeight; y++) {
      for (let x = 0; x < viewportWidth; x++) {
        const worldX = x + Math.floor(this.camera.x);
        const worldY = y + Math.floor(this.camera.y);

        if (this.isValidMapPosition(map, worldX, worldY)) {
          // Find tile at the given position
          const tile = map.tiles.find(t => t.x === worldX && t.y === worldY);

          if (tile) {
            const tileSprite = this.getTileSprite(tile.type);
            
            this.componentLogger.debug('Drawing tile', {
              worldX,
              worldY,
              tileType: tile.type,
              spriteX: tileSprite.spriteX,
              spriteY: tileSprite.spriteY,
            });

            this.drawSprite(
              this.contexts.terrain,
              'terrain',
              tileSprite.spriteX,
              tileSprite.spriteY,
              x * this.tileSize,
              y * this.tileSize,
            );
          } else {
            this.componentLogger.warn('Missing tile data at', { worldX, worldY });
          }
        } else {
          this.componentLogger.debug('Tile position out of bounds', { worldX, worldY });
        }
      }
    }

    this.componentLogger.groupEnd();
  }

  /**
   * Render game objects
   */
  private renderObjects(objects?: readonly GameObject[]): void {
    this.componentLogger.group('Rendering object layer');
    this.componentLogger.debug('Objects input', objects);

    if (!objects) {
      this.componentLogger.warn('Objects is null or undefined');
      this.componentLogger.groupEnd();
      return;
    }

    // Convert single object to array or ensure objects is an array
    const objectsArray = Array.isArray(objects) ? objects : [objects];

    objectsArray.forEach((obj) => {
      const screenX = (obj.x - this.camera.x) * this.tileSize;
      const screenY = (obj.y - this.camera.y) * this.tileSize;

      this.componentLogger.debug('Calculated screen coordinates', { screenX, screenY });

      if (this.isOnScreen(screenX, screenY)) {
        this.componentLogger.info('Drawing object', {
          x: obj.x,
          y: obj.y,
          spriteX: obj.spriteX,
          spriteY: obj.spriteY,
        });

        this.drawSprite(
          this.contexts.objects,
          'characters',
          obj.spriteX || 0,
          obj.spriteY || 0,
          screenX,
          screenY,
        );
      } else {
        this.componentLogger.debug('Object outside viewport', { screenX, screenY });
      }
    });

    this.componentLogger.groupEnd();
  }

  /**
   * Render visual effects (placeholder for future implementation)
   * @unused - Reserved for future effects system implementation
   */
  // @ts-ignore: TS6133 - Method reserved for future implementation
  private renderEffects(effects?: readonly any[]): void {
    this.componentLogger.group('Rendering effects layer');
    this.componentLogger.debug('Effects array', effects);

    if (!effects) {
      this.componentLogger.warn('Effects array is null or undefined');
      this.componentLogger.groupEnd();
      return;
    }

    effects.forEach((effect) => {
      const screenX = (effect.x - this.camera.x) * this.tileSize;
      const screenY = (effect.y - this.camera.y) * this.tileSize;

      this.componentLogger.debug('Calculated screen coordinates', { screenX, screenY });

      if (this.isOnScreen(screenX, screenY)) {
        this.componentLogger.info('Drawing effect', {
          x: effect.x,
          y: effect.y,
          spriteX: effect.spriteX,
          spriteY: effect.spriteY,
        });

        this.drawSprite(
          this.contexts.effects,
          'effects',
          effect.spriteX || 0,
          effect.spriteY || 0,
          screenX,
          screenY,
        );
      } else {
        this.componentLogger.debug('Effect outside viewport', { screenX, screenY });
      }
    });

    this.componentLogger.groupEnd();
  }

  /**
   * Draw a sprite on the specified context
   */
  private drawSprite(
    ctx: CanvasRenderingContext2D,
    spriteName: string,
    spriteX: number,
    spriteY: number,
    destX: number,
    destY: number,
  ): void {
    const spriteInfo = this.sprites.get(spriteName);
    
    if (!spriteInfo || !spriteInfo.loaded) {
      this.componentLogger.warn(`Sprite "${spriteName}" not loaded or not found`);
      return;
    }

    try {
      ctx.drawImage(
        spriteInfo.image,
        spriteX * this.tileSize,
        spriteY * this.tileSize,
        this.tileSize,
        this.tileSize,
        destX,
        destY,
        this.tileSize,
        this.tileSize,
      );
    } catch (error) {
      this.componentLogger.error(`Failed to draw sprite "${spriteName}"`, error);
    }
  }

  /**
   * Check if a position is on screen
   */
  private isOnScreen(x: number, y: number): boolean {
    return (
      x >= -this.tileSize &&
      y >= -this.tileSize &&
      x <= this.canvasLayers.objects.width &&
      y <= this.canvasLayers.objects.height
    );
  }

  /**
   * Check if a map position is valid
   */
  private isValidMapPosition(map: GameMap, x: number, y: number): boolean {
    return (
      x >= 0 &&
      x < (map.width || 0) &&
      y >= 0 &&
      y < (map.height || 0)
    );
  }

  /**
   * Update camera position
   */
  updateCamera(x: number, y: number, zoom?: number): void {
    this.camera.x = x;
    this.camera.y = y;
    if (zoom !== undefined) {
      this.camera.zoom = Math.max(0.1, Math.min(5.0, zoom));
    }
    
    this.componentLogger.debug('Camera updated', this.camera);
  }

  /**
   * Get current camera state
   */
  getCamera(): Readonly<Camera> {
    return { ...this.camera };
  }

  /**
   * Update highlighted cells (for combat targeting)
   */
  updateHighlights(cells: string[]): void {
    // Implementation for highlighting cells
    // This would typically modify the effect layer
    this.componentLogger.debug('Updating highlights', { cells });
  }

  /**
   * Get rendering performance stats
   */
  getPerformanceStats(): { frameTime: number; frameCount: number; fps: number } {
    const fps = this.frameCount > 0 ? 1000 / this.lastFrameTime : 0;
    return {
      frameTime: this.lastFrameTime,
      frameCount: this.frameCount,
      fps: Math.round(fps * 100) / 100
    };
  }

  /**
   * Check WebGL support
   */
  private checkWebGLSupport(): void {
    try {
      const testCanvas = document.createElement('canvas');
      const webglCtx = testCanvas.getContext('webgl') || testCanvas.getContext('experimental-webgl');
      
      if (webglCtx) {
        this.componentLogger.info('WebGL support detected (available as potential fallback)');
      } else {
        this.componentLogger.warn('No WebGL support detected - limited to Canvas 2D rendering');
      }
    } catch (webglError) {
      this.componentLogger.warn('WebGL detection failed', webglError);
    }
  }

  /**
   * Get sprite coordinates for a tile type
   */
  private getTileSprite(tileType: TileType): { spriteX: number; spriteY: number } {
    // Default sprite mapping for tile types
    // These coordinates should match your actual sprite sheet layout
    const tileSprites: Record<TileType, { spriteX: number; spriteY: number }> = {
      floor: { spriteX: 0, spriteY: 0 },
      wall: { spriteX: 32, spriteY: 0 },
      door: { spriteX: 64, spriteY: 0 },
      stairs: { spriteX: 96, spriteY: 0 },
      water: { spriteX: 0, spriteY: 32 },
      void: { spriteX: 32, spriteY: 32 }
    };
    
    return tileSprites[tileType] || tileSprites.floor;
  }
}
