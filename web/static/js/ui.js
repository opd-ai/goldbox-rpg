class GameUI extends EventEmitter {
  constructor(gameState, combatManager) {
    super();
    this.gameState = gameState;
    this.combatManager = combatManager;
    this.elements = {
      portrait: document.getElementById("character-portrait"),
      name: document.getElementById("character-name"),
      stats: {
        str: document.getElementById("stat-str"),
        dex: document.getElementById("stat-dex"),
        con: document.getElementById("stat-con"),
        int: document.getElementById("stat-int"),
        wis: document.getElementById("stat-wis"),
        cha: document.getElementById("stat-cha"),
      },
      hpBar: document.getElementById("hp-bar"),
      logContent: document.getElementById("log-content"),
      actionButtons: document.querySelectorAll(".action-btn"),
      dirButtons: document.querySelectorAll(".dir-btn"),
    };

    this.setupEventListeners();
    this.setupKeyboardControls();
  }

  setupEventListeners() {
    // Movement controls
    this.elements.dirButtons.forEach((btn) => {
      btn.addEventListener("click", () => this.handleMove(btn.dataset.dir));
    });

    // Game state updates
    this.gameState.on("stateChanged", (state) => this.updateUI(state));

    // Combat events
    this.combatManager.on("updateCombatLog", (data) =>
      this.updateCombatLog(data),
    );
  }

  setupKeyboardControls() {
    const keyMap = {
      ArrowUp: "n",
      ArrowDown: "s",
      ArrowLeft: "w",
      ArrowRight: "e",
      Home: "nw",
      PageUp: "ne",
      End: "sw",
      PageDown: "se",
      Space: "wait",
    };

    document.addEventListener("keydown", (e) => {
      if (keyMap[e.code]) {
        e.preventDefault();
        this.handleMove(keyMap[e.code]);
      }
    });
  }

  async handleMove(direction) {
    if (
      this.combatManager.active &&
      this.gameState.player.id !== this.combatManager.currentTurn
    ) {
      return;
    }

    try {
      await this.gameState.move(direction);
    } catch (error) {
      this.logMessage(`Move failed: ${error.message}`, "error");
    }
  }

  updateUI(state) {
    const { player } = state.current;

    // Update character info
    this.elements.portrait.src = `./static/assets/portraits/${player.class.toLowerCase()}.png`;
    this.elements.name.textContent = player.name;

    // Update stats
    Object.entries(this.elements.stats).forEach(([stat, element]) => {
      element.textContent = player[stat];
    });

    // Update HP bar
    const hpPercent = (player.hp / player.maxHp) * 100;
    this.elements.hpBar.style.width = `${hpPercent}%`;
    this.elements.hpBar.style.backgroundColor =
      hpPercent < 25 ? "red" : hpPercent < 50 ? "yellow" : "green";
  }

  logMessage(message, type = "info") {
    const maxMessages = 100;
    const entry = document.createElement("div");
    entry.className = `log-entry log-${type}`;
    entry.textContent = message;

    // Remove old messages first to prevent unnecessary reflows
    while (this.elements.logContent.children.length >= maxMessages) {
      this.elements.logContent.removeChild(this.elements.logContent.firstChild);
    }

    this.elements.logContent.appendChild(entry);
    this.elements.logContent.scrollTop = this.elements.logContent.scrollHeight;
  }

  updateCombatLog(data) {
    const { currentTurn, initiative } = data;
    const isPlayerTurn = currentTurn === this.gameState.player.id;

    this.logMessage(`${isPlayerTurn ? "Your" : currentTurn + "'s"} turn`);

    // Update initiative display
    this.updateInitiativeOrder(initiative);
  }

  // Add implementation for updateInitiativeOrder
  updateInitiativeOrder(initiative) {
    const initiativeList = document.createElement("div");
    initiativeList.className = "initiative-list";

    initiative.forEach((entityId) => {
      const entity = this.gameState.world.objects[entityId];
      const item = document.createElement("div");
      item.className = `initiative-item ${entityId === this.combatManager.currentTurn ? "active" : ""}`;
      item.textContent = entity.name;
      initiativeList.appendChild(item);
    });

    const oldList = document.querySelector(".initiative-list");
    if (oldList) {
      oldList.replaceWith(initiativeList);
    } else {
      document.getElementById("combat-log").prepend(initiativeList);
    }
  }
}
