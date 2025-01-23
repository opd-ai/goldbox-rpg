class CombatManager extends EventEmitter {
  constructor(gameState, renderer) {
    super();
    this.gameState = gameState;
    this.renderer = renderer;
    this.active = false;
    this.currentTurn = null;
    this.initiative = [];
    this.selectedAction = null;
    this.selectedTarget = null;
    this.highlightedCells = new Set();

    this.setupEventListeners();
  }

  cleanup() {
    document.querySelectorAll(".action-btn").forEach((btn) => {
      btn.removeEventListener("click", this.handleActionButton);
    });
    this.highlightedCells.clear();
    this.renderer.updateHighlights(this.highlightedCells);
  }

  setupEventListeners() {
    // Combat action buttons
    document.querySelectorAll(".action-btn").forEach((btn) => {
      btn.addEventListener("click", () =>
        this.handleActionButton(btn.dataset.action),
      );
    });

    // Combat grid interaction
    document
      .getElementById("terrain-layer")
      .addEventListener("click", (e) =>
        this.handleGridClick(this.getGridPosition(e)),
      );
  }

  async startCombat(participants) {
    try {
      const result = await this.gameState.rpc.startCombat(participants);
      if (result.success) {
        this.active = true;
        this.initiative = result.initiative;
        this.currentTurn = result.first_turn;
        this.emit("combatStarted", result);
        this.updateUI();
      }
    } catch (error) {
      this.emit("error", error);
    }
  }

  async handleActionButton(action) {
    if (!this.active || this.currentTurn !== this.gameState.player.id) return;

    // Clear previous state
    this.selectedAction = null;
    this.highlightedCells.clear();
    this.renderer.updateHighlights(this.highlightedCells);

    // Set new state
    this.selectedAction = action;
    await this.highlightValidTargets(action);
  }

  async executeAction(action, target) {
    try {
      let result;
      switch (action) {
        case "attack":
          result = await this.gameState.attack(
            target.id,
            this.gameState.player.equipped.weapon,
          );
          break;
        case "cast":
          result = await this.gameState.castSpell(
            this.selectedSpell,
            target.id,
            target.position,
          );
          break;
        case "item":
          result = await this.gameState.useItem(this.selectedItem, target.id);
          break;
        case "end":
          result = await this.gameState.endTurn();
          break;
      }

      if (result.success) {
        await this.playActionAnimation(action, target, result);
        this.updateUI();
      }
    } catch (error) {
      this.emit("error", error);
    }
  }

  async playActionAnimation(action, target, result) {
    switch (action) {
      case "attack":
        await this.renderer.playAttackAnimation(
          this.gameState.player.position,
          target.position,
          result.hit,
        );
        if (result.hit) {
          await this.renderer.playDamageNumber(target.position, result.damage);
        }
        break;
      case "cast":
        await this.renderer.playSpellAnimation(
          this.selectedSpell,
          target.position,
        );
        break;
    }
  }

  highlightValidTargets(action) {
    this.highlightedCells.clear();

    switch (action) {
      case "attack":
        this.highlightAttackTargets();
        break;
      case "cast":
        this.highlightSpellTargets();
        break;
      case "item":
        this.highlightItemTargets();
        break;
    }

    this.renderer.updateHighlights(this.highlightedCells);
  }

  highlightAttackTargets() {
    const range = this.gameState.player.equipped.weapon.range;
    const playerPos = this.gameState.player.position;

    this.gameState.world.objects.forEach((obj) => {
      if (
        obj.faction !== this.gameState.player.faction &&
        this.isInRange(playerPos, obj.position, range)
      ) {
        this.highlightedCells.add(obj.position);
      }
    });
  }

  highlightSpellTargets() {
    if (!this.selectedSpell) return;
    const range = this.selectedSpell.range;
    const playerPos = this.gameState.player.position;

    this.gameState.world.objects.forEach((obj) => {
      if (this.isInRange(playerPos, obj.position, range)) {
        this.highlightedCells.add(obj.position);
      }
    });
  }

  highlightItemTargets() {
    if (!this.selectedItem) return;
    const range = this.selectedItem.range;
    const playerPos = this.gameState.player.position;

    this.gameState.world.objects.forEach((obj) => {
      if (this.isInRange(playerPos, obj.position, range)) {
        this.highlightedCells.add(obj.position);
      }
    });
  }

  isInRange(from, to, range) {
    const dx = Math.abs(to.x - from.x);
    const dy = Math.abs(to.y - from.y);
    return dx + dy <= range;
  }

  getGridPosition(event) {
    const rect = event.target.getBoundingClientRect();
    const x = Math.floor((event.clientX - rect.left) / this.renderer.tileSize);
    const y = Math.floor((event.clientY - rect.top) / this.renderer.tileSize);
    return { x, y };
  }

  updateUI() {
    // Update turn indicator
    document.querySelectorAll(".action-btn").forEach((btn) => {
      btn.disabled = this.currentTurn !== this.gameState.player.id;
    });

    // Update combat log
    this.emit("updateCombatLog", {
      currentTurn: this.currentTurn,
      initiative: this.initiative,
    });
  }
}
