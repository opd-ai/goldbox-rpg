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

  async loadSprites() {
    const spriteUrls = {
      terrain: "./static/assets/sprites/terrain.png",
      characters: "./static/assets/sprites/characters.png",
      effects: "./static/assets/sprites/effects.png",
      ui: "./static/assets/sprites/ui.png",
    };

    for (const [key, url] of Object.entries(spriteUrls)) {
      const img = new Image();
      img.src = url;
      await new Promise((resolve) => (img.onload = resolve));
      this.sprites.set(key, img);
    }
  }

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

  clearLayers() {
    [this.terrainCtx, this.objectCtx, this.effectCtx].forEach((ctx) => {
      ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
    });
  }

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

  render(gameState) {
    this.clearLayers();
    this.renderTerrain(gameState.world?.map);
    this.renderObjects(gameState.world?.objects);
    this.renderEffects(gameState.world?.effects);
  }

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

  // Add these methods
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

  isOnScreen(x, y) {
    return (
      x >= -this.tileSize &&
      y >= -this.tileSize &&
      x <= this.objectLayer.width &&
      y <= this.objectLayer.height
    );
  }
}
