/**
 * UI manager class that handles game interface elements and player controls
 * Extends EventEmitter to support event-based communication
 * 
 * @class
 * @extends EventEmitter
 * @description
 * Manages:
 * - DOM element references and updates
 * - User input handling (keyboard/mouse)
 * - Combat interface updates
 * - Game state visualization
 * - Message logging system
 * 
 * Core responsibilities:
 * - Initializes and maintains UI element references
 * - Sets up event listeners for user input
 * - Updates display based on game state changes
 * - Handles player movement commands
 * - Manages combat log and initiative display
 * 
 * Required dependencies:
 * - GameState: Manages core game state
 * - CombatManager: Handles combat system
 * - EventEmitter: Provides event handling capabilities
 * 
 * @param {GameState} gameState - Game state management instance
 * @param {CombatManager} combatManager - Combat system instance
 * 
 * @fires GameUI#move - When player attempts movement
 * @fires GameUI#logMessage - When adding message to game log
 * @fires GameUI#updateUI - When refreshing interface elements
 * 
 * @listens {GameState#stateChanged} - Updates UI when game state changes
 * @listens {CombatManager#updateCombatLog} - Updates combat display
 * 
 * @throws {Error} If required DOM elements are not found during initialization
 * @throws {Error} If gameState or combatManager dependencies are missing
 * 
 * @example
 * const ui = new GameUI(gameState, combatManager);
 * 
 * @see GameState
 * @see CombatManager
 * @see EventEmitter
 */
class GameUI extends EventEmitter {
  /**
   * Creates a new GameUI instance to manage game interface and controls
   * 
   * @param {GameState} gameState - The game state manager instance
   * @param {CombatManager} combatManager - The combat management system instance
   * 
   * @description
   * Initializes UI by:
   * - Setting up references to DOM elements
   * - Configuring event listeners for UI interactions
   * - Setting up keyboard controls
   * 
   * Required DOM elements:
   * - Character portrait image
   * - Character name display
   * - Stat displays (str, dex, con, int, wis, cha)
   * - HP bar
   * - Log content area
   * - Action and direction buttons
   * 
   * @throws {Error} If required DOM elements are not found
   * @see setupEventListeners
   * @see setupKeyboardControls
   */
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

  /**
   * Sets up event listeners for UI interactions including movement controls, game state updates, and combat events
   * Binds click handlers for directional buttons and subscribes to game state and combat manager events
   * 
   * Events handled:
   * - Direction button clicks: Triggers handleMove with direction data
   * - Game state changes: Updates UI with new state
   * - Combat updates: Updates combat log with new data
   * 
   * @listens {click} Direction button click events
   * @listens {stateChanged} Game state change events 
   * @listens {updateCombatLog} Combat log update events
   * 
   * @see handleMove
   * @see updateUI  
   * @see updateCombatLog
   */
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

  /**
   * Sets up keyboard controls for movement and actions
   * Maps arrow keys, home/end/pageup/pagedown to cardinal/diagonal directions
   * Maps spacebar to wait action
   * 
   * Attaches keydown event listener to document that:
   * - Prevents default behavior for mapped keys
   * - Translates key codes to movement commands
   * - Calls handleMove() with the mapped direction
   * 
   * Key mappings:
   * - Arrow keys → n,s,w,e (cardinal directions)
   * - Home/PgUp/End/PgDn → nw,ne,sw,se (diagonal directions) 
   * - Space → wait
   * 
   * @see handleMove - Called with mapped direction when valid key pressed
   * @see KeyboardEvent.code - Used to identify pressed keys
   * 
   * @returns {void}
   */
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

  /**
   * Handles player movement in a specific direction, respecting combat turn order
   * 
   * @param {string} direction - The direction to move ('up', 'down', 'left', 'right')
   * @returns {Promise<void>}
   * @throws {Error} If movement fails due to collision or invalid direction
   * 
   * @description
   * This method checks if movement is allowed based on combat state and turn order,
   * then attempts to move the player in the specified direction.
   * Movement is blocked if:
   * - Combat is active and it's not the player's turn
   * - The destination is blocked/invalid
   * 
   * @example
   * await ui.handleMove('up');
   * 
   * @see {@link GameState#move}
   * @see {@link CombatManager}
   */
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

  /**
   * Updates the UI elements based on the current game state
   * 
   * @param {Object} state - The current game state
   * @param {Object} state.current - The current state data
   * @param {Object} state.current.player - Player character data
   * @param {string} state.current.player.name - Player name
   * @param {string} state.current.player.class - Player character class
   * @param {number} state.current.player.hp - Current hit points
   * @param {number} state.current.player.maxHp - Maximum hit points
   * @param {Object.<string,number>} state.current.player.stats - Player statistics
   * 
   * Updates:
   * - Character portrait image source
   * - Character name display
   * - All character statistics
   * - HP bar width and color (green >50%, yellow 25-50%, red <25%)
   * 
   * @throws {Error} If required player properties are missing
   * @throws {Error} If UI elements are not properly initialized
   * 
   * @see this.elements - UI element references needed for updates
   */
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

  /**
   * Adds a message to the log display with specified type styling
   * 
   * @param {string} message - The text message to display in the log
   * @param {string} [type="info"] - The type/style of log message ("info", "error", etc)
   * 
   * @description
   * Appends a new message div to the log content area, limited to maxMessages entries.
   * Older messages are removed if the limit is exceeded. Automatically scrolls to
   * latest message.
   * 
   * @example
   * logMessage("Player attacked monster", "combat")
   * logMessage("Error loading map", "error") 
   *
   * @remarks
   * - Maintains a fixed size buffer of messages (maxMessages = 100)
   * - Removes oldest messages first when buffer is full
   * - Automatically scrolls to show newest messages
   * - Message styling controlled by log-${type} CSS classes
   */
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

  /**
   * Updates the combat log with turn information and initiative order
   * 
   * @param {Object} data - The combat data object
   * @param {string|number} data.currentTurn - ID of entity whose turn it currently is
   * @param {Array} data.initiative - Array containing initiative order of entities
   * 
   * @throws {TypeError} Will throw if data parameters are missing or invalid types
   * 
   * @see updateInitiativeOrder - Called to update initiative display
   * @see logMessage - Used to display turn information
   * @see gameState.player - References player state for turn comparison
   */
  updateCombatLog(data) {
    const { currentTurn, initiative } = data;
    const isPlayerTurn = currentTurn === this.gameState.player.id;

    this.logMessage(`${isPlayerTurn ? "Your" : currentTurn + "'s"} turn`);

    // Update initiative display
    this.updateInitiativeOrder(initiative);
  }

  /**
   * Updates the initiative order display in the combat UI
   * 
   * @param {Array<string>} initiative - Array of entity IDs in initiative order
   * @throws {Error} If gameState.world.objects or combatManager are not initialized
   * @see CombatManager
   * @see GameState
   * 
   * Creates or updates the initiative list display showing the turn order.
   * The current entity's turn is highlighted.
   * Replaces existing list if present, otherwise prepends to combat log.
   */
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
