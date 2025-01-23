class GameRenderer {
  constructor() {
    this.terrainLayer = document.getElementById("terrain-layer");
    this.objectLayer = document.getElementById("object-layer");
    this.effectLayer = document.getElementById("effect-layer");

    this.terrainCtx = this.terrainLayer.getContext("2d");
    this.objectCtx = this.objectLayer.getContext("2d");
    this.effectCtx = this.effectLayer.getContext("2d");

    this.tileSize = 32;
    this.sprites = new Map();
    this.animations = new Map();

    this.camera = {
      x: 0,
      y: 0,
      zoom: 1,
    };

    window.addEventListener("resize", this.handleResize.bind(this));
    this.handleResize();
  }

  /**
   * Asynchronously loads sprite images from predefined URLs and stores them in the sprites Map.
   * 
   * Loads terrain, characters, effects and UI sprite sheets from the static assets directory.
   * Each sprite is loaded as an Image object and stored with its corresponding key in this.sprites.
   * 
   * @async
   * @returns {Promise<void>} Resolves when all sprites are loaded successfully
   * @throws {Error} Throws an error if any sprite fails to load, with details about which sprite failed
   * 
   * @example
   * await renderer.loadSprites();
   * 
   * @see {@link Image} for the browser's Image object implementation
   * @see {@link Map#set} for how sprites are stored
   */
  async loadSprites() {
    const spriteUrls = {
      terrain: "./static/assets/sprites/terrain.png",
      characters: "./static/assets/sprites/characters.png",
      effects: "./static/assets/sprites/effects.png",
      ui: "./static/assets/sprites/ui.png",
    };

    for (const [key, url] of Object.entries(spriteUrls)) {
      try {
        const img = new Image();
        img.src = url;
        await new Promise((resolve, reject) => {
          img.onload = resolve;
          img.onerror = () =>
            reject(new Error(`Failed to load sprite: ${url}`));
        });
        this.sprites.set(key, img);
      } catch (error) {
        console.error(`Failed to load sprite ${key}:`, error);
        throw error;
      }
    }
  }

  /**
   * Handles resizing of all canvas layers to match the container dimensions.
   * Updates the width and height of terrain, object and effect layers when viewport container is resized.
   * 
   * @returns {void}
   * 
   * @example
   * // Typical usage within a resize event listener
   * window.addEventListener('resize', () => this.handleResize());
   * 
   * @see terrainLayer Canvas element for terrain rendering
   * @see objectLayer Canvas element for game objects rendering 
   * @see effectLayer Canvas element for effects rendering
   *
   * @notes
   * - Requires a DOM element with id "viewport-container" to exist
   * - All canvas layers must be initialized before calling this method
   * - Both the canvas dimensions and CSS dimensions are updated to prevent scaling issues
   */
  handleResize() {
    const container = document.getElementById("viewport-container");
    const width = container.clientWidth;
    const height = container.clientHeight;

    [this.terrainLayer, this.objectLayer, this.effectLayer].forEach(
      (canvas) => {
        canvas.width = width;
        canvas.height = height;
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
      },
    );
  }

  /**
   * Clears all canvas layers by erasing their contents.
   * Iterates through terrain, object and effect contexts and clears their entire canvas area.
   * 
   * @method clearLayers
   * @memberof Renderer
   * @instance
   * 
   * @description
   * This method clears the contents of three canvas layers:
   * - Terrain layer (background/floor tiles)
   * - Object layer (items, entities, etc)
   * - Effect layer (animations, particles, etc)
   * Each canvas is cleared by calling clearRect() with dimensions matching the canvas size.
   * 
   * @returns {void}
   * 
   * @see {@link Renderer#terrainCtx}
   * @see {@link Renderer#objectCtx} 
   * @see {@link Renderer#effectCtx}
   */
  clearLayers() {
    [this.terrainCtx, this.objectCtx, this.effectCtx].forEach((ctx) => {
      ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
    });
  }

  /**
   * Draws a sprite from a spritesheet onto the canvas context at the specified position
   * 
   * @param {CanvasRenderingContext2D} ctx - The canvas 2D rendering context to draw on
   * @param {string} spriteName - The name/key of the sprite in the sprites Map
   * @param {number} sx - The x coordinate of the sprite in the spritesheet (in tiles)
   * @param {number} sy - The y coordinate of the sprite in the spritesheet (in tiles) 
   * @param {number} dx - The x coordinate on the canvas to draw the sprite
   * @param {number} dy - The y coordinate on the canvas to draw the sprite
   * @param {number} [width=this.tileSize] - The width to draw the sprite (defaults to tileSize)
   * @param {number} [height=this.tileSize] - The height to draw the sprite (defaults to tileSize)
   * @returns {void}
   * 
   * @throws {undefined} If sprite is not found in the sprites Map, silently returns
   * 
   * @see {@link this.sprites} - Map containing loaded sprite images
   * @see {@link this.tileSize} - Size of a single tile in pixels
   */
  drawSprite(
    ctx,
    spriteName,
    sx,
    sy,
    dx,
    dy,
    width = this.tileSize,
    height = this.tileSize,
  ) {
    const sprite = this.sprites.get(spriteName);
    if (!sprite) return;

    ctx.drawImage(
      sprite,
      sx * this.tileSize,
      sy * this.tileSize,
      this.tileSize,
      this.tileSize,
      dx,
      dy,
      width,
      height,
    );
  }

  /**
   * Renders the current game state by clearing and redrawing all visual layers
   * 
   * @param {Object} gameState - The complete game state object to render
   * @param {Object} [gameState.world] - The world state containing map and object data
   * @param {Array} [gameState.world.map] - 2D array representing the terrain/map tiles
   * @param {Array} [gameState.world.objects] - Array of game objects to render
   * @param {Array} [gameState.world.effects] - Array of visual effects to render
   * 
   * @returns {void}
   * 
   * @see clearLayers
   * @see renderTerrain
   * @see renderObjects  
   * @see renderEffects
   */
  render(gameState) {
    this.clearLayers();
    this.renderTerrain(gameState.world?.map);
    this.renderObjects(gameState.world?.objects);
    this.renderEffects(gameState.world?.effects);
  }

  /**
   * Renders the terrain tiles within the viewport based on the camera position
   * 
   * @param {Object} map - The map object containing terrain data
   * @param {number} map.width - Width of the entire map in tiles
   * @param {number} map.height - Height of the entire map in tiles
   * @param {Function} map.getTile - Function that returns tile data for a given x,y coordinate
   * 
   * @returns {void}
   * 
   * @remarks
   * - Iterates through viewport tiles and renders visible terrain
   * - Handles edge cases by checking map bounds
   * - Early returns if map is null/undefined
   * - Uses this.camera position to determine visible area
   * - Draws terrain sprites using this.drawSprite()
   * 
   * @see drawSprite
   */
  renderTerrain(map) {
    if (!map) return;

    const viewportWidth = Math.ceil(this.terrainLayer.width / this.tileSize);
    const viewportHeight = Math.ceil(this.terrainLayer.height / this.tileSize);

    for (let y = 0; y < viewportHeight; y++) {
      for (let x = 0; x < viewportWidth; x++) {
        const worldX = x + Math.floor(this.camera.x);
        const worldY = y + Math.floor(this.camera.y);

        if (
          worldX >= 0 &&
          worldX < map.width &&
          worldY >= 0 &&
          worldY < map.height
        ) {
          const tile = map.getTile(worldX, worldY);
          this.drawSprite(
            this.terrainCtx,
            "terrain",
            tile.spriteX,
            tile.spriteY,
            x * this.tileSize,
            y * this.tileSize,
          );
        }
      }
    }
  }

  /**
   * Renders game objects to the canvas based on camera position
   * 
   * @param {Object[]} objects - Array of game objects to render
   * @param {number} objects[].x - X coordinate of object in world space
   * @param {number} objects[].y - Y coordinate of object in world space  
   * @param {number} objects[].spriteX - X coordinate of sprite in sprite sheet
   * @param {number} objects[].spriteY - Y coordinate of sprite in sprite sheet
   * 
   * @returns {void}
   * 
   * @throws {TypeError} If objects parameter is not an array
   * 
   * @example
   * renderer.renderObjects([
   *   {x: 10, y: 20, spriteX: 0, spriteY: 32}
   * ]);
   * 
   * @see {@link drawSprite}
   * @see {@link isOnScreen}
   */
  renderObjects(objects) {
    if (!objects) return;

    objects.forEach((obj) => {
      const screenX = (obj.x - this.camera.x) * this.tileSize;
      const screenY = (obj.y - this.camera.y) * this.tileSize;

      if (this.isOnScreen(screenX, screenY)) {
        this.drawSprite(
          this.objectCtx,
          "characters",
          obj.spriteX,
          obj.spriteY,
          screenX,
          screenY,
        );
      }
    });
  }


  /**
   * Renders visual effects on the game canvas
   * @param {Array<Object>} effects - Array of effect objects to render
   * @param {number} effects[].x - X coordinate of the effect in world space
   * @param {number} effects[].y - Y coordinate of the effect in world space 
   * @param {number} effects[].spriteX - X coordinate of the effect sprite in the sprite sheet
   * @param {number} effects[].spriteY - Y coordinate of the effect sprite in the sprite sheet
   * @returns {void}
   * 
   * Each effect object in the array must have:
   * - x,y coordinates in world space
   * - spriteX,spriteY coordinates for the sprite sheet position
   * 
   * Effects are only rendered if they are within the visible screen area
   * as determined by isOnScreen(). If effects array is null/undefined,
   * function returns early.
   * 
   * @see isOnScreen
   * @see drawSprite
   */
  renderEffects(effects) {
    if (!effects) return;

    effects.forEach((effect) => {
      const screenX = (effect.x - this.camera.x) * this.tileSize;
      const screenY = (effect.y - this.camera.y) * this.tileSize;

      if (this.isOnScreen(screenX, screenY)) {
        this.drawSprite(
          this.effectCtx,
          "effects",
          effect.spriteX,
          effect.spriteY,
          screenX,
          screenY,
        );
      }
    });
  }

  /**
   * Updates visual highlights on the effect layer canvas for specified grid cells
   * 
   * @param {Array<{x: number, y: number}>} cells - Array of cell positions to highlight
   *                                                Each cell object must have x,y coordinates
   * 
   * @description
   * Clears the existing effect layer canvas and draws semi-transparent yellow 
   * highlight rectangles for each cell position. Only cells within the visible 
   * screen area are rendered.
   * 
   * @requires this.effectCtx - Canvas 2D rendering context for effects layer
   * @requires this.camera - Camera position object with x,y coordinates
   * @requires this.tileSize - Size of each grid tile in pixels
   * @requires this.isOnScreen() - Method to check if position is in viewport
   * 
   * @see isOnScreen
   */
  updateHighlights(cells) {
    this.effectCtx.clearRect(
      0,
      0,
      this.effectLayer.width,
      this.effectLayer.height,
    );

    cells.forEach((pos) => {
      const screenX = (pos.x - this.camera.x) * this.tileSize;
      const screenY = (pos.y - this.camera.y) * this.tileSize;

      if (this.isOnScreen(screenX, screenY)) {
        this.effectCtx.fillStyle = "rgba(255, 255, 0, 0.3)";
        this.effectCtx.fillRect(screenX, screenY, this.tileSize, this.tileSize);
      }
    });
  }

  /**
   * Checks if a given coordinate point is within the visible screen bounds
   * 
   * @param {number} x - The x coordinate to check
   * @param {number} y - The y coordinate to check
   * @returns {boolean} True if the point is within screen bounds, false otherwise
   * 
   * @remarks
   * - Considers points within a tile size outside the bounds to be "on screen" 
   * - Uses this.tileSize and this.objectLayer dimensions to determine boundaries
   * - Coordinates can be negative (up to -tileSize)
   */
  isOnScreen(x, y) {
    return (
      x >= -this.tileSize &&
      y >= -this.tileSize &&
      x <= this.objectLayer.width &&
      y <= this.objectLayer.height
    );
  }
}
