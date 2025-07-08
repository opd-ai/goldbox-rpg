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
    console.group("constructor: Initializing GameUI");

    console.debug("constructor: Parameters:", { gameState, combatManager });

    if (!gameState || !combatManager) {
      console.error("constructor: Missing required dependencies");
      throw new Error("GameUI requires gameState and combatManager");
    }

    super();

    console.info("constructor: Setting up core dependencies");
    this.gameState = gameState;
    this.combatManager = combatManager;

    console.info("constructor: Initializing UI element references");
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

    // Check if all elements were found
    Object.entries(this.elements).forEach(([key, element]) => {
      if (!element || (element instanceof NodeList && element.length === 0)) {
        console.warn(`constructor: UI element "${key}" not found`);
      }
    });

    console.info("constructor: Setting up event handlers and controls");
    this.setupEventListeners();
    this.setupKeyboardControls();

    console.groupEnd();
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
    console.group("setupEventListeners: Setting up UI event handlers");

    // Movement controls
    console.debug("setupEventListeners: Binding direction button events");
    this.elements.dirButtons.forEach((btn) => {
      if (!btn.dataset.dir) {
        console.warn(
          "setupEventListeners: Direction button missing data-dir attribute",
        );
      }
      btn.addEventListener("click", () => this.handleMove(btn.dataset.dir));
    });

    // Game state updates
    console.info("setupEventListeners: Registering state change listener");
    this.gameState.on("stateChanged", (state) => {
      if (!state) {
        console.error("setupEventListeners: Received invalid state update");
        return;
      }
      this.updateUI(state);
    });

    // Combat events
    console.info("setupEventListeners: Registering combat log listener");
    this.combatManager.on("updateCombatLog", (data) => {
      if (!data) {
        console.error("setupEventListeners: Received invalid combat data");
        return;
      }
      this.updateCombatLog(data);
    });

    console.groupEnd();
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
    console.group("setupKeyboardControls: Setting up keyboard event handlers");

    console.info("setupKeyboardControls: Initializing key mapping");
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

    console.info("setupKeyboardControls: Adding keydown event listener");
    this.boundKeydownHandler = (e) => {
      console.debug("setupKeyboardControls: Key pressed:", e.code);

      if (keyMap[e.code]) {
        e.preventDefault();
        console.info(
          "setupKeyboardControls: Processing mapped key:",
          keyMap[e.code],
        );
        this.handleMove(keyMap[e.code]);
      } else {
        console.warn("setupKeyboardControls: Unmapped key pressed:", e.code);
      }
    };
    document.addEventListener("keydown", this.boundKeydownHandler);

    console.groupEnd();
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
    console.group("handleMove: Processing movement request");
    console.debug("handleMove: Direction:", direction);

    if (
      this.combatManager.active &&
      this.gameState.player.id !== this.combatManager.currentTurn
    ) {
      console.warn("handleMove: Movement blocked - not player turn in combat");
      console.groupEnd();
      return;
    }

    try {
      console.info("handleMove: Attempting to move player");
      await this.gameState.move(direction);
      console.info("handleMove: Movement successful");
    } catch (error) {
      console.error("handleMove: Movement failed:", error.message);
      this.logMessage(`Move failed: ${error.message}`, "error");
    }

    console.groupEnd();
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
    console.group("updateUI: Updating interface elements");
    console.debug("updateUI: State received:", state);

    if (!state?.current?.player) {
      console.error("updateUI: Invalid state object received");
      console.groupEnd();
      return;
    }

    const { player } = state.current;

    // Update character info
    console.info("updateUI: Updating character portrait and name");
    const portraitPath = `./static/assets/portraits/${player.class.toLowerCase()}.png`;
    this.elements.portrait.src = portraitPath;
    this.elements.name.textContent = player.name;

    // Update stats
    console.info("updateUI: Updating character statistics");
    Object.entries(this.elements.stats).forEach(([stat, element]) => {
      if (!player[stat]) {
        console.warn(`updateUI: Missing stat value for ${stat}`);
      }
      element.textContent = player[stat];
    });

    // Update HP bar
    console.info("updateUI: Updating HP bar");
    const hpPercent = (player.hp / player.maxHp) * 100;
    if (hpPercent < 25) {
      console.warn("updateUI: Player HP critically low");
    }
    this.elements.hpBar.style.width = `${hpPercent}%`;
    this.elements.hpBar.style.backgroundColor =
      hpPercent < 25 ? "red" : hpPercent < 50 ? "yellow" : "green";

    console.groupEnd();
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
    console.group("logMessage: Adding new message to log");
    console.debug("logMessage: Parameters:", { message, type });

    const maxMessages = 100;
    const entry = document.createElement("div");
    entry.className = `log-entry log-${type}`;
    entry.textContent = message;

    // Check message count
    const currentCount = this.elements.logContent.children.length;
    if (currentCount >= maxMessages) {
      console.warn("logMessage: Max messages reached, removing oldest entries");
      while (this.elements.logContent.children.length >= maxMessages) {
        this.elements.logContent.removeChild(
          this.elements.logContent.firstChild,
        );
      }
    }

    console.info("logMessage: Appending new message entry");
    this.elements.logContent.appendChild(entry);

    if (!this.elements.logContent) {
      console.error("logMessage: Log content element not found");
    } else {
      console.info("logMessage: Scrolling to latest message");
      this.elements.logContent.scrollTop =
        this.elements.logContent.scrollHeight;
    }

    console.groupEnd();
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
    console.group("updateCombatLog: Processing combat log update");
    console.debug("updateCombatLog: Data received:", data);

    if (!data?.currentTurn || !data?.initiative) {
      console.error("updateCombatLog: Invalid combat data received");
      console.groupEnd();
      return;
    }

    const { currentTurn, initiative } = data;
    const isPlayerTurn = currentTurn === this.gameState.player.id;

    if (!this.gameState?.player?.id) {
      console.warn("updateCombatLog: Player state may be invalid");
    }

    console.info("updateCombatLog: Logging turn message");
    this.logMessage(`${isPlayerTurn ? "Your" : currentTurn + "'s"} turn`);

    console.info("updateCombatLog: Updating initiative order display");
    this.updateInitiativeOrder(initiative);

    console.groupEnd();
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
    console.group("updateInitiativeOrder: Updating initiative display");
    console.debug("updateInitiativeOrder: Initiative array:", initiative);

    const initiativeList = document.createElement("div");
    initiativeList.className = "initiative-list";

    console.info("updateInitiativeOrder: Creating initiative list items");
    initiative.forEach((entityId) => {
      const entity = this.gameState.world.objects[entityId];
      if (!entity) {
        console.warn(
          "updateInitiativeOrder: Entity not found for ID:",
          entityId,
        );
        return;
      }

      const item = document.createElement("div");
      item.className = `initiative-item ${entityId === this.combatManager.currentTurn ? "active" : ""}`;
      item.textContent = entity.name;
      initiativeList.appendChild(item);
    });

    console.info("updateInitiativeOrder: Updating DOM");
    const oldList = document.querySelector(".initiative-list");
    if (oldList) {
      console.debug("updateInitiativeOrder: Replacing existing list");
      oldList.replaceWith(initiativeList);
    } else {
      const combatLog = document.getElementById("combat-log");
      if (!combatLog) {
        console.error("updateInitiativeOrder: Combat log element not found");
        console.groupEnd();
        return;
      }
      console.debug("updateInitiativeOrder: Creating new list");
      combatLog.prepend(initiativeList);
    }

    console.groupEnd();
  }

  /**
   * Cleans up event listeners and resources to prevent memory leaks
   * Should be called when the UI manager is no longer needed
   */
  cleanup() {
    console.group("UIManager.cleanup");
    try {
      console.debug("UIManager.cleanup: removing keydown event listener");
      if (this.boundKeydownHandler) {
        document.removeEventListener("keydown", this.boundKeydownHandler);
        this.boundKeydownHandler = null;
      }
      console.info("UIManager.cleanup: cleanup completed");
    } catch (err) {
      console.error("UIManager.cleanup:", err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }
}
