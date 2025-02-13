class GameRenderer {
  constructor() {
    console.group("Constructor: GameRenderer initialization");
    console.debug("Constructor: Getting canvas elements");

    this.terrainLayer = document.getElementById("terrain-layer");
    this.objectLayer = document.getElementById("object-layer");
    this.effectLayer = document.getElementById("effect-layer");

    if (!this.terrainLayer || !this.objectLayer || !this.effectLayer) {
      console.error("Constructor: Failed to get one or more canvas elements");
      throw new Error("Canvas elements not found");
    }

    console.info("Constructor: Setting up canvas contexts");
    this.terrainCtx = this.terrainLayer.getContext("2d");
    this.objectCtx = this.objectLayer.getContext("2d");
    this.effectCtx = this.effectLayer.getContext("2d");

    console.info("Constructor: Initializing core properties");
    this.tileSize = 32;
    this.sprites = new Map();
    this.animations = new Map();

    this.camera = {
      x: 0,
      y: 0,
      zoom: 1,
    };

    console.debug("Constructor: Setting up resize event listener");
    window.addEventListener("resize", this.handleResize.bind(this));

    console.info("Constructor: Performing initial resize");
    this.handleResize();

    console.groupEnd();
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
    console.group("loadSprites: Loading sprite assets");
    const spriteUrls = {
      terrain: "./static/assets/sprites/terrain.png",
      characters: "./static/assets/sprites/characters.png",
      effects: "./static/assets/sprites/effects.png",
      ui: "./static/assets/sprites/ui.png",
    };
    console.debug("loadSprites: Sprite URLs to load:", spriteUrls);

    for (const [key, url] of Object.entries(spriteUrls)) {
      console.info(`loadSprites: Loading sprite "${key}" from ${url}`);
      try {
        const img = new Image();
        img.src = url;
        await new Promise((resolve, reject) => {
          img.onload = () => {
            console.info(`loadSprites: Successfully loaded sprite "${key}"`);
            resolve();
          };
          img.onerror = () => {
            console.error(
              `loadSprites: Failed to load sprite "${key}" from ${url}`,
            );
            reject(new Error(`Failed to load sprite: ${url}`));
          };
        });
        this.sprites.set(key, img);
      } catch (error) {
        console.error(`loadSprites: Error loading sprite "${key}":`, error);
        throw error;
      }
    }

    console.info(`loadSprites: Completed loading ${this.sprites.size} sprites`);
    console.groupEnd();
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
    console.group("handleResize: Resizing canvas layers");

    const container = document.getElementById("viewport-container");
    if (!container) {
      console.error("handleResize: Viewport container not found");
      console.groupEnd();
      return;
    }

    const width = container.clientWidth;
    const height = container.clientHeight;
    console.debug("handleResize: Container dimensions", { width, height });

    if (width === 0 || height === 0) {
      console.warn("handleResize: Container has zero dimension");
    }

    [this.terrainLayer, this.objectLayer, this.effectLayer].forEach(
      (canvas) => {
        if (!canvas) {
          console.error("handleResize: Canvas layer is null");
          return;
        }
        console.info(`handleResize: Resizing canvas to ${width}x${height}`);
        canvas.width = width;
        canvas.height = height;
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
      },
    );

    console.groupEnd();
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
    console.group("clearLayers: Clearing all canvas layers");
    console.debug("clearLayers: Canvas contexts:", [
      this.terrainCtx,
      this.objectCtx,
      this.effectCtx,
    ]);

    [this.terrainCtx, this.objectCtx, this.effectCtx].forEach((ctx) => {
      if (!ctx) {
        console.error("clearLayers: Missing context:", ctx);
        return;
      }
      console.info(
        `clearLayers: Clearing canvas of size ${ctx.canvas.width}x${ctx.canvas.height}`,
      );
      ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
    });

    console.groupEnd();
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
    console.group("drawSprite: Drawing sprite to canvas");
    console.debug("drawSprite: Parameters:", {
      spriteName,
      sx,
      sy,
      dx,
      dy,
      width,
      height,
    });

    const sprite = this.sprites.get(spriteName);
    if (!sprite) {
      console.error("drawSprite: Sprite not found:", spriteName);
      console.groupEnd();
      return;
    }

    if (width !== this.tileSize || height !== this.tileSize) {
      console.warn("drawSprite: Non-standard tile dimensions used:", {
        width,
        height,
      });
    }

    console.info("drawSprite: Drawing image with dimensions:", {
      sourceX: sx * this.tileSize,
      sourceY: sy * this.tileSize,
      sourceWidth: this.tileSize,
      sourceHeight: this.tileSize,
      destX: dx,
      destY: dy,
      destWidth: width,
      destHeight: height,
    });

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

    console.groupEnd();
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
    console.group("render: Rendering game state");
    console.debug("render: Game state:", gameState);

    if (!gameState) {
      console.error("render: Game state is null or undefined");
      console.groupEnd();
      return;
    }

    if (!gameState.world) {
      console.warn("render: World data is missing from game state");
    }

    this.clearLayers();
    console.info("render: Cleared all canvas layers");

    this.renderTerrain(gameState.world?.map);
    console.info("render: Rendered terrain layer");

    this.renderObjects(gameState.world?.objects);
    console.info("render: Rendered objects layer");

    this.renderEffects(gameState.world?.effects);
    console.info("render: Rendered effects layer");

    console.groupEnd();
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
    console.group("renderTerrain: Rendering terrain layer");
    console.debug("renderTerrain: Map data:", map);

    if (!map) {
      console.error("renderTerrain: Map is null or undefined");
      console.groupEnd();
      return;
    }

    const viewportWidth = Math.ceil(this.terrainLayer.width / this.tileSize);
    const viewportHeight = Math.ceil(this.terrainLayer.height / this.tileSize);
    console.info("renderTerrain: Viewport dimensions:", {
      viewportWidth,
      viewportHeight,
    });

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
          if (!tile) {
            console.warn("renderTerrain: Missing tile data at", {
              worldX,
              worldY,
            });
            continue;
          }

          console.debug("renderTerrain: Drawing tile:", {
            worldX,
            worldY,
            spriteX: tile.spriteX,
            spriteY: tile.spriteY,
          });

          this.drawSprite(
            this.terrainCtx,
            "terrain",
            tile.spriteX,
            tile.spriteY,
            x * this.tileSize,
            y * this.tileSize,
          );
        } else {
          console.warn("renderTerrain: Tile position out of bounds:", {
            worldX,
            worldY,
          });
        }
      }
    }

    console.groupEnd();
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
    console.group("renderObjects: Rendering object layer");
    console.debug("renderObjects: Objects input:", objects);
  
    if (!objects) {
      console.warn("renderObjects: Objects is null or undefined");
      console.groupEnd();
      return;
    }
  
    // Convert single object to array or ensure objects is an array
    const objectsArray = Array.isArray(objects) ? objects : [objects];
  
    objectsArray.forEach((obj) => {
      const screenX = (obj.x - this.camera.x) * this.tileSize;
      const screenY = (obj.y - this.camera.y) * this.tileSize;
  
      console.debug("renderObjects: Calculated screen coordinates:", {
        screenX,
        screenY,
      });
  
      if (this.isOnScreen(screenX, screenY)) {
        console.info("renderObjects: Drawing object:", {
          x: obj.x,
          y: obj.y,
          spriteX: obj.spriteX,
          spriteY: obj.spriteY,
        });
  
        this.drawSprite(
          this.objectCtx,
          "characters",
          obj.spriteX,
          obj.spriteY,
          screenX,
          screenY,
        );
      } else {
        console.warn("renderObjects: Object outside viewport:", {
          screenX,
          screenY,
        });
      }
    });

    console.groupEnd();
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
    console.group("renderEffects: Rendering effects layer");
    console.debug("renderEffects: Effects array:", effects);

    if (!effects) {
      console.warn("renderEffects: Effects array is null or undefined");
      console.groupEnd();
      return;
    }

    effects.forEach((effect) => {
      const screenX = (effect.x - this.camera.x) * this.tileSize;
      const screenY = (effect.y - this.camera.y) * this.tileSize;

      console.debug("renderEffects: Calculated screen coordinates:", {
        screenX,
        screenY,
      });

      if (this.isOnScreen(screenX, screenY)) {
        console.info("renderEffects: Drawing effect:", {
          x: effect.x,
          y: effect.y,
          spriteX: effect.spriteX,
          spriteY: effect.spriteY,
        });

        this.drawSprite(
          this.effectCtx,
          "effects",
          effect.spriteX,
          effect.spriteY,
          screenX,
          screenY,
        );
      } else {
        console.warn("renderEffects: Effect outside viewport:", {
          screenX,
          screenY,
        });
      }
    });

    console.groupEnd();
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
    console.group("updateHighlights: Updating highlight effects");
    console.debug("updateHighlights: Cells array:", cells);

    if (!cells) {
      console.error("updateHighlights: Cells array is null or undefined");
      console.groupEnd();
      return;
    }

    console.info("updateHighlights: Clearing effect layer");
    this.effectCtx.clearRect(
      0,
      0,
      this.effectLayer.width,
      this.effectLayer.height,
    );

    cells.forEach((pos) => {
      const screenX = (pos.x - this.camera.x) * this.tileSize;
      const screenY = (pos.y - this.camera.y) * this.tileSize;

      console.debug("updateHighlights: Calculated screen coordinates:", {
        screenX,
        screenY,
      });

      if (this.isOnScreen(screenX, screenY)) {
        console.info("updateHighlights: Drawing highlight at:", {
          screenX,
          screenY,
        });
        this.effectCtx.fillStyle = "rgba(255, 255, 0, 0.3)";
        this.effectCtx.fillRect(screenX, screenY, this.tileSize, this.tileSize);
      } else {
        console.warn("updateHighlights: Cell outside viewport:", {
          screenX,
          screenY,
        });
      }
    });

    console.groupEnd();
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
    console.group("isOnScreen: Checking if coordinate is within viewport");
    console.debug("isOnScreen: Checking coordinates:", { x, y });

    if (x < -this.tileSize || y < -this.tileSize) {
      console.warn("isOnScreen: Coordinates below minimum bounds");
    }

    if (x > this.objectLayer.width || y > this.objectLayer.height) {
      console.warn("isOnScreen: Coordinates exceed viewport dimensions");
    }

    const result =
      x >= -this.tileSize &&
      y >= -this.tileSize &&
      x <= this.objectLayer.width &&
      y <= this.objectLayer.height;

    console.info("isOnScreen: Visibility check result:", result);
    console.groupEnd();
    return result;
  }
}
