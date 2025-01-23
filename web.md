Project Path: web

Source Tree:

```
web
├── index.html
└── static
    ├── css
    │   ├── main.css
    │   ├── ui.css
    │   └── combat.css
    ├── js
    │   ├── ui.js
    │   ├── combat.js
    │   ├── game.js
    │   ├── rpc.js
    │   └── render.js
    └── assets

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/index.html`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Gold Box RPG</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="stylesheet" href="/static/css/main.css">
    <link rel="stylesheet" href="/static/css/combat.css">
    <link rel="stylesheet" href="/static/css/ui.css">
</head>
<body>
    <div id="game-container">
        <!-- Main viewport -->
        <div id="viewport-container">
            <canvas id="terrain-layer"></canvas>
            <canvas id="object-layer"></canvas>
            <canvas id="effect-layer"></canvas>
        </div>

        <!-- Character panel -->
        <div id="character-panel">
            <div id="portrait-container">
                <img id="character-portrait" src="" alt="Character Portrait">
                <div id="character-name"></div>
            </div>
            <div id="stats-container">
                <div class="stat-row">
                    <span>HP:</span>
                    <div class="stat-bar" id="hp-bar"></div>
                </div>
                <div class="stat-grid">
                    <div class="stat">STR: <span id="stat-str"></span></div>
                    <div class="stat">DEX: <span id="stat-dex"></span></div>
                    <div class="stat">CON: <span id="stat-con"></span></div>
                    <div class="stat">INT: <span id="stat-int"></span></div>
                    <div class="stat">WIS: <span id="stat-wis"></span></div>
                    <div class="stat">CHA: <span id="stat-cha"></span></div>
                </div>
            </div>
        </div>

        <!-- Combat log -->
        <div id="combat-log">
            <div id="log-content"></div>
        </div>

        <!-- Action panel -->
        <div id="action-panel">
            <div id="combat-actions">
                <button class="action-btn" data-action="attack">Attack</button>
                <button class="action-btn" data-action="cast">Cast Spell</button>
                <button class="action-btn" data-action="item">Use Item</button>
                <button class="action-btn" data-action="end">End Turn</button>
            </div>
            <div id="movement-controls">
                <div class="direction-grid">
                    <button class="dir-btn" data-dir="nw">↖</button>
                    <button class="dir-btn" data-dir="n">↑</button>
                    <button class="dir-btn" data-dir="ne">↗</button>
                    <button class="dir-btn" data-dir="w">←</button>
                    <button class="dir-btn" data-dir="wait">•</button>
                    <button class="dir-btn" data-dir="e">→</button>
                    <button class="dir-btn" data-dir="sw">↙</button>
                    <button class="dir-btn" data-dir="s">↓</button>
                    <button class="dir-btn" data-dir="se">↘</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Game scripts -->
    <script src="/static/js/rpc.js"></script>
    <script src="/static/js/game.js"></script>
    <script src="/static/js/render.js"></script>
    <script src="/static/js/combat.js"></script>
    <script src="/static/js/ui.js"></script>
</body>
</html>
```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/static/css/main.css`:

```css
:root {
    --gold-dark: #8B7355;
    --gold-light: #D4C391;
    --bg-dark: #2C2C2C;
    --bg-light: #454545;
    --text-primary: #D4C391;
    --text-secondary: #8B7355;
    --border-color: #8B7355;
}

body {
    margin: 0;
    padding: 0;
    background: var(--bg-dark);
    color: var(--text-primary);
    font-family: 'Courier New', monospace;
}

#game-container {
    display: grid;
    grid-template-columns: 3fr 1fr;
    grid-template-rows: auto 1fr auto;
    gap: 1rem;
    padding: 1rem;
    height: 100vh;
}

#viewport-container {
    grid-column: 1;
    grid-row: 1 / span 3;
    position: relative;
    border: 2px solid var(--border-color);
    background: var(--bg-light);
}

canvas {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    image-rendering: pixelated;
}
```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/static/css/ui.css`:

```css
:root {
    --gold-dark: #8B7355;
    --gold-light: #D4C391;
    --bg-dark: #2C2C2C;
    --bg-light: #454545;
    --text-primary: #D4C391;
    --text-secondary: #8B7355;
    --border-color: #8B7355;
}

body {
    margin: 0;
    padding: 0;
    background: var(--bg-dark);
    color: var(--text-primary);
    font-family: 'Courier New', monospace;
}

#game-container {
    display: grid;
    grid-template-columns: 3fr 1fr;
    grid-template-rows: auto 1fr auto;
    gap: 1rem;
    padding: 1rem;
    height: 100vh;
}

#viewport-container {
    grid-column: 1;
    grid-row: 1 / span 3;
    position: relative;
    border: 2px solid var(--border-color);
    background: var(--bg-light);
}

canvas {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    image-rendering: pixelated;
}
```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/static/css/combat.css`:

```css
#action-panel {
    grid-column: 2;
    grid-row: 3;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    border: 2px solid var(--border-color);
    background: var(--bg-light);
}

.action-btn {
    width: 100%;
    padding: 0.5rem;
    background: var(--bg-dark);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    cursor: pointer;
}

.action-btn:hover {
    background: var(--gold-dark);
}

.direction-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.25rem;
}

.dir-btn {
    aspect-ratio: 1;
    background: var(--bg-dark);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    cursor: pointer;
    font-size: 1.2rem;
}

.dir-btn:hover {
    background: var(--gold-dark);
}
```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/static/js/ui.js`:

```js
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
    console.group('constructor: Initializing GameUI');
    
    console.debug('constructor: Parameters:', { gameState, combatManager });
    
    if (!gameState || !combatManager) {
      console.error('constructor: Missing required dependencies');
      throw new Error('GameUI requires gameState and combatManager');
    }

    super();
    
    console.info('constructor: Setting up core dependencies');
    this.gameState = gameState;
    this.combatManager = combatManager;

    console.info('constructor: Initializing UI element references');
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

    console.info('constructor: Setting up event handlers and controls');
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
    console.group('setupEventListeners: Setting up UI event handlers');

    // Movement controls
    console.debug('setupEventListeners: Binding direction button events');
    this.elements.dirButtons.forEach((btn) => {
      if (!btn.dataset.dir) {
        console.warn('setupEventListeners: Direction button missing data-dir attribute');
      }
      btn.addEventListener("click", () => this.handleMove(btn.dataset.dir));
    });

    // Game state updates
    console.info('setupEventListeners: Registering state change listener');
    this.gameState.on("stateChanged", (state) => {
      if (!state) {
        console.error('setupEventListeners: Received invalid state update');
        return;
      }
      this.updateUI(state);
    });

    // Combat events
    console.info('setupEventListeners: Registering combat log listener');
    this.combatManager.on("updateCombatLog", (data) => {
      if (!data) {
        console.error('setupEventListeners: Received invalid combat data');
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
    console.group('setupKeyboardControls: Setting up keyboard event handlers');

    console.info('setupKeyboardControls: Initializing key mapping');
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

    console.info('setupKeyboardControls: Adding keydown event listener');
    document.addEventListener("keydown", (e) => {
      console.debug('setupKeyboardControls: Key pressed:', e.code);
      
      if (keyMap[e.code]) {
        e.preventDefault();
        console.info('setupKeyboardControls: Processing mapped key:', keyMap[e.code]);
        this.handleMove(keyMap[e.code]);
      } else {
        console.warn('setupKeyboardControls: Unmapped key pressed:', e.code);
      }
    });

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
    console.group('handleMove: Processing movement request');
    console.debug('handleMove: Direction:', direction);

    if (
      this.combatManager.active &&
      this.gameState.player.id !== this.combatManager.currentTurn
    ) {
      console.warn('handleMove: Movement blocked - not player turn in combat');
      console.groupEnd();
      return;
    }

    try {
      console.info('handleMove: Attempting to move player');
      await this.gameState.move(direction);
      console.info('handleMove: Movement successful');
    } catch (error) {
      console.error('handleMove: Movement failed:', error.message);
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
    console.group('updateUI: Updating interface elements');
    console.debug('updateUI: State received:', state);

    if (!state?.current?.player) {
      console.error('updateUI: Invalid state object received');
      console.groupEnd();
      return;
    }

    const { player } = state.current;

    // Update character info
    console.info('updateUI: Updating character portrait and name');
    const portraitPath = `./static/assets/portraits/${player.class.toLowerCase()}.png`;
    this.elements.portrait.src = portraitPath;
    this.elements.name.textContent = player.name;

    // Update stats
    console.info('updateUI: Updating character statistics');
    Object.entries(this.elements.stats).forEach(([stat, element]) => {
      if (!player[stat]) {
        console.warn(`updateUI: Missing stat value for ${stat}`);
      }
      element.textContent = player[stat];
    });

    // Update HP bar
    console.info('updateUI: Updating HP bar');
    const hpPercent = (player.hp / player.maxHp) * 100;
    if (hpPercent < 25) {
      console.warn('updateUI: Player HP critically low');
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
    console.group('logMessage: Adding new message to log');
    console.debug('logMessage: Parameters:', { message, type });

    const maxMessages = 100;
    const entry = document.createElement("div");
    entry.className = `log-entry log-${type}`;
    entry.textContent = message;

    // Check message count
    const currentCount = this.elements.logContent.children.length;
    if (currentCount >= maxMessages) {
      console.warn('logMessage: Max messages reached, removing oldest entries');
      while (this.elements.logContent.children.length >= maxMessages) {
        this.elements.logContent.removeChild(this.elements.logContent.firstChild);
      }
    }

    console.info('logMessage: Appending new message entry');
    this.elements.logContent.appendChild(entry);

    if (!this.elements.logContent) {
      console.error('logMessage: Log content element not found');
    } else {
      console.info('logMessage: Scrolling to latest message');
      this.elements.logContent.scrollTop = this.elements.logContent.scrollHeight;
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
    console.group('updateCombatLog: Processing combat log update');
    console.debug('updateCombatLog: Data received:', data);

    if (!data?.currentTurn || !data?.initiative) {
      console.error('updateCombatLog: Invalid combat data received');
      console.groupEnd();
      return;
    }

    const { currentTurn, initiative } = data;
    const isPlayerTurn = currentTurn === this.gameState.player.id;

    if (!this.gameState?.player?.id) {
      console.warn('updateCombatLog: Player state may be invalid');
    }

    console.info('updateCombatLog: Logging turn message');
    this.logMessage(`${isPlayerTurn ? "Your" : currentTurn + "'s"} turn`);

    console.info('updateCombatLog: Updating initiative order display');
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
    console.group('updateInitiativeOrder: Updating initiative display');
    console.debug('updateInitiativeOrder: Initiative array:', initiative);

    const initiativeList = document.createElement("div");
    initiativeList.className = "initiative-list";

    console.info('updateInitiativeOrder: Creating initiative list items');
    initiative.forEach((entityId) => {
      const entity = this.gameState.world.objects[entityId];
      if (!entity) {
        console.warn('updateInitiativeOrder: Entity not found for ID:', entityId);
        return;
      }

      const item = document.createElement("div");
      item.className = `initiative-item ${entityId === this.combatManager.currentTurn ? "active" : ""}`;
      item.textContent = entity.name;
      initiativeList.appendChild(item);
    });

    console.info('updateInitiativeOrder: Updating DOM');
    const oldList = document.querySelector(".initiative-list");
    if (oldList) {
      console.debug('updateInitiativeOrder: Replacing existing list');
      oldList.replaceWith(initiativeList);
    } else {
      const combatLog = document.getElementById("combat-log");
      if (!combatLog) {
        console.error('updateInitiativeOrder: Combat log element not found');
        console.groupEnd();
        return;
      }
      console.debug('updateInitiativeOrder: Creating new list');
      combatLog.prepend(initiativeList);
    }

    console.groupEnd();
  }
}

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/static/js/combat.js`:

```js
/**
 * @class CombatManager
 * @extends EventEmitter
 * @description Manages turn-based combat mechanics including initiative, actions, and targeting.
 * Handles combat flow, player actions, target selection, and UI updates during combat sequences.
 * 
 * @fires combatStarted - When a new combat sequence begins
 * @fires error - When combat operations fail
 * @fires updateCombatLog - When combat state changes require UI updates
 * 
 * @property {Object} gameState - Global game state reference
 * @property {Object} renderer - Rendering system for combat visuals
 * @property {boolean} active - Whether combat is currently in progress
 * @property {string|null} currentTurn - ID of entity whose turn is active
 * @property {Array} initiative - Ordered list of combatants' turn sequence
 * @property {Object|null} selectedAction - Currently chosen combat action
 * @property {Object|null} selectedTarget - Currently selected action target
 * @property {Set} highlightedCells - Grid cells currently highlighted for targeting
 * 
 * @param {Object} gameState - Game state containing world, player and combat data
 * @param {Object} renderer - Rendering system for combat visualizations
 * 
 * @throws {Error} If gameState or renderer are not provided
 * 
 * @example
 * const combat = new CombatManager(gameState, renderer);
 * await combat.startCombat([player, enemy1, enemy2]);
 */
class CombatManager extends EventEmitter {
  /**
   * Initializes a combat controller instance
   * @param {Object} gameState - The current state of the game
   * @param {Object} renderer - The renderer used to display the game
   * @constructor
   * @extends EventEmitter
   * @description Creates a new combat controller that manages combat state and flow.
   * Initializes combat-related properties like turn order, selected actions/targets,
   * and highlighted cells. Sets up event listeners for combat interactions.
   * @property {boolean} active - Whether combat is currently active
   * @property {Object|null} currentTurn - The entity whose turn it currently is
   * @property {Array} initiative - Array tracking turn order
   * @property {Object|null} selectedAction - Currently selected combat action
   * @property {Object|null} selectedTarget - Currently selected target
   * @property {Set} highlightedCells - Set of cells currently highlighted
   */
  constructor(gameState, renderer) {
    console.group('CombatManager.constructor');
    
    try {
      console.debug('CombatManager.constructor: params', { gameState, renderer });

      if (!gameState || !renderer) {
        console.error('CombatManager.constructor: missing required parameters');
        throw new Error('gameState and renderer are required');
      }

      super();

      this.gameState = gameState;
      this.renderer = renderer;
      console.info('CombatManager.constructor: initialized core dependencies');

      this.active = false; 
      this.currentTurn = null;
      this.initiative = [];
      this.selectedAction = null;
      this.selectedTarget = null;
      this.highlightedCells = new Set();
      console.info('CombatManager.constructor: initialized combat state');

      this.setupEventListeners();
      console.info('CombatManager.constructor: event listeners set up');

    } catch (err) {
      console.error('CombatManager.constructor:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Performs cleanup operations for the combat system:
   * - Removes event listeners from action buttons
   * - Clears highlighted cells
   * - Updates renderer with cleared highlights
   * 
   * This should be called when transitioning away from combat mode
   * or resetting the combat state.
   * 
   * @see CombatRenderer#updateHighlights
   */
  cleanup() {
    console.group('CombatManager.cleanup');
    
    try {
      console.debug('CombatManager.cleanup: starting cleanup process');

      document.querySelectorAll(".action-btn").forEach((btn) => {
        btn.removeEventListener("click", this.handleActionButton);
      });
      console.info('CombatManager.cleanup: removed action button event listeners');

      if (this.highlightedCells.size > 0) {
        console.warn('CombatManager.cleanup: clearing non-empty highlighted cells');
      }
      this.highlightedCells.clear();
      console.info('CombatManager.cleanup: cleared highlighted cells');

      this.renderer.updateHighlights(this.highlightedCells);
      console.info('CombatManager.cleanup: updated renderer highlights');

    } catch (err) {
      console.error('CombatManager.cleanup:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Sets up event listeners for combat-related interactions in the game.
   * Initializes click handlers for action buttons and terrain grid.
   * 
   * Attaches listeners to:
   * - Action buttons with class 'action-btn' that trigger handleActionButton()
   * - Terrain layer grid clicks that trigger handleGridClick()
   * 
   * @see handleActionButton
   * @see handleGridClick
   * @see getGridPosition
   */
  setupEventListeners() {
    console.group('CombatManager.setupEventListeners');
    
    try {
      console.debug('CombatManager.setupEventListeners: initializing event listeners');

      // Combat action buttons
      document.querySelectorAll(".action-btn").forEach((btn) => {
        btn.addEventListener("click", () => {
          console.info('CombatManager.setupEventListeners: action button clicked', btn.dataset.action);
          this.handleActionButton(btn.dataset.action);
        });
      });
      console.info('CombatManager.setupEventListeners: attached action button listeners');

      // Combat grid interaction
      const terrainLayer = document.getElementById("terrain-layer");
      if (!terrainLayer) {
        console.warn('CombatManager.setupEventListeners: terrain layer element not found');
      } else {
        terrainLayer.addEventListener("click", (e) => {
          const pos = this.getGridPosition(e);
          console.info('CombatManager.setupEventListeners: grid clicked at', pos);
          this.handleGridClick(pos);
        });
        console.info('CombatManager.setupEventListeners: attached grid click listener');
      }

    } catch (err) {
      console.error('CombatManager.setupEventListeners:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Initiates a combat sequence with the specified participants.
   * 
   * @param {Object[]} participants - Array of combat participants with their properties
   * @returns {Promise<void>} 
   * @fires combatStarted - When combat successfully starts with initiative order
   * @fires error - When there is an error starting combat
   * @throws {Error} When RPC call fails
   * 
   * The method:
   * 1. Makes RPC call to start combat
   * 2. If successful:
   *    - Sets combat as active
   *    - Stores initiative order
   *    - Sets first turn
   *    - Emits combatStarted event
   *    - Updates UI
   * 3. If error occurs, emits error event
   *
   * @see updateUI
   * @see gameState.rpc.startCombat
   */
  async startCombat(participants) {
    console.group('CombatManager.startCombat');
    
    try {
      console.debug('CombatManager.startCombat: params', { participants });

      const result = await this.gameState.rpc.startCombat(participants);
      console.info('CombatManager.startCombat: RPC call complete', result);

      if (result.success) {
        this.active = true;
        this.initiative = result.initiative; 
        this.currentTurn = result.first_turn;
        console.info('CombatManager.startCombat: combat state initialized');

        this.emit("combatStarted", result);
        console.info('CombatManager.startCombat: combatStarted event emitted');

        this.updateUI();
        console.info('CombatManager.startCombat: UI updated');
      } else {
        console.warn('CombatManager.startCombat: RPC call unsuccessful');
      }

    } catch (error) {
      console.error('CombatManager.startCombat:', error);
      this.emit("error", error);
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Handles the selection of an action button in combat
   * @param {Object} action - The combat action to be performed
   * @returns {Promise<void>} - Nothing
   * @throws {Error} Implicitly may throw if highlighting targets fails
   * @description
   * This method manages the combat action selection flow:
   * 1. Validates turn state
   * 2. Clears previous selection state
   * 3. Sets new selected action
   * 4. Highlights valid targets for the action
   * 
   * Only processes actions if combat is active and it's the player's turn
   * 
   * @see highlightValidTargets
   * @see renderer.updateHighlights
   */
  async handleActionButton(action) {
    console.group('CombatManager.handleActionButton');
    
    try {
      console.debug('CombatManager.handleActionButton: params', { action });

      if (!this.active || this.currentTurn !== this.gameState.player.id) {
        console.warn('CombatManager.handleActionButton: action blocked - inactive or not player turn');
        console.groupEnd();
        return;
      }

      // Clear previous state
      this.selectedAction = null;
      this.highlightedCells.clear();
      this.renderer.updateHighlights(this.highlightedCells);
      console.info('CombatManager.handleActionButton: cleared previous state');

      // Set new state
      this.selectedAction = action;
      console.info('CombatManager.handleActionButton: selected new action', action);

      this.highlightValidTargets(action);
      console.info('CombatManager.handleActionButton: highlighted valid targets');

    } catch (err) {
      console.error('CombatManager.handleActionButton:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Executes a combat action against a target
   * @param {string} action - The type of action to execute ('attack', 'cast', 'item', or 'end')
   * @param {Object} target - The target object containing id and position properties
   * @param {string} target.id - The unique identifier of the target
   * @param {Object} [target.position] - The position coordinates of the target (required for 'cast' action)
   * @returns {Promise<void>} - Resolves when the action is complete
   * @throws {Error} - If the action execution fails
   * @emits error - When an error occurs during execution
   * @see gameState.attack
   * @see gameState.castSpell  
   * @see gameState.useItem
   * @see gameState.endTurn
   * @see playActionAnimation
   * @see updateUI
   */
  async executeAction(action, target) {
    console.group('CombatManager.executeAction');
    
    try {
      console.debug('CombatManager.executeAction: params', { action, target });

      let result;
      switch (action) {
        case "attack":
          result = await this.gameState.attack(
            target.id,
            this.gameState.player.equipped.weapon,
          );
          console.info('CombatManager.executeAction: attack executed', result);
          break;
        case "cast":
          result = await this.gameState.castSpell(
            this.selectedSpell,
            target.id,
            target.position,
          );
          console.info('CombatManager.executeAction: spell cast', result);
          break;
        case "item":
          result = await this.gameState.useItem(this.selectedItem, target.id);
          console.info('CombatManager.executeAction: item used', result);
          break;
        case "end":
          result = await this.gameState.endTurn();
          console.info('CombatManager.executeAction: turn ended', result);
          break;
      }

      if (result.success) {
        await this.playActionAnimation(action, target, result);
        console.info('CombatManager.executeAction: animation played');
        this.updateUI();
        console.info('CombatManager.executeAction: UI updated');
      } else {
        console.warn('CombatManager.executeAction: action failed', result);
      }

    } catch (error) {
      console.error('CombatManager.executeAction:', error);
      this.emit("error", error);
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Plays an animation for a combat action based on the action type
   * 
   * @param {string} action - The type of action being performed ("attack" or "cast")
   * @param {Object} target - The target of the action, must have a position property
   * @param {Object} result - The result of the action
   * @param {boolean} result.hit - Whether the attack hit or missed (for attack actions)
   * @param {number} result.damage - The amount of damage dealt (for successful attacks)
   * @returns {Promise<void>} - Resolves when the animation completes
   * 
   * @see Renderer#playAttackAnimation
   * @see Renderer#playDamageNumber 
   * @see Renderer#playSpellAnimation
   */
  async playActionAnimation(action, target, result) {
    console.group('CombatManager.playActionAnimation');
    
    try {
      console.debug('CombatManager.playActionAnimation: params', { action, target, result });

      switch (action) {
        case "attack":
          console.info('CombatManager.playActionAnimation: playing attack animation');
          await this.renderer.playAttackAnimation(
            this.gameState.player.position,
            target.position,
            result.hit,
          );
          
          if (result.hit) {
            console.info('CombatManager.playActionAnimation: playing damage number animation');
            await this.renderer.playDamageNumber(target.position, result.damage);
          } else {
            console.warn('CombatManager.playActionAnimation: attack missed');
          }
          break;

        case "cast":
          console.info('CombatManager.playActionAnimation: playing spell animation');
          await this.renderer.playSpellAnimation(
            this.selectedSpell,
            target.position,
          );
          break;

        default:
          console.warn('CombatManager.playActionAnimation: unhandled action type', action);
      }

    } catch (err) {
      console.error('CombatManager.playActionAnimation:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Highlights valid target cells on the game grid based on the selected action type.
   * Clears any existing highlighted cells before applying new highlights.
   * 
   * @param {string} action - The type of action being taken ('attack', 'cast', or 'item')
   * @throws {Error} If an invalid action type is provided
   * @see {@link highlightAttackTargets}
   * @see {@link highlightSpellTargets} 
   * @see {@link highlightItemTargets}
   * @see {@link renderer.updateHighlights}
   */
  highlightValidTargets(action) {
    console.group('CombatManager.highlightValidTargets');
    
    try {
      console.debug('CombatManager.highlightValidTargets: params', { action });

      this.highlightedCells.clear();
      console.info('CombatManager.highlightValidTargets: cleared highlighted cells');

      switch (action) {
        case "attack":
          this.highlightAttackTargets();
          console.info('CombatManager.highlightValidTargets: highlighted attack targets');
          break;
        case "cast":
          this.highlightSpellTargets();
          console.info('CombatManager.highlightValidTargets: highlighted spell targets');
          break;
        case "item":
          this.highlightItemTargets();
          console.info('CombatManager.highlightValidTargets: highlighted item targets');
          break;
        default:
          console.warn('CombatManager.highlightValidTargets: unrecognized action type', action);
      }

      this.renderer.updateHighlights(this.highlightedCells);
      console.info('CombatManager.highlightValidTargets: updated renderer highlights');

    } catch (err) {
      console.error('CombatManager.highlightValidTargets:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Highlights valid attack targets within the player's weapon range.
   * Identifies objects that are:
   * 1. Not in the same faction as the player
   * 2. Within the equipped weapon's range from player position
   * 
   * Adds valid target positions to the highlightedCells collection.
   * 
   * @requires gameState.player - The current player object
   * @requires gameState.player.equipped.weapon - The player's equipped weapon
   * @requires gameState.world.objects - Collection of game objects
   * @requires highlightedCells - Set to store highlighted cell positions
   * @requires isInRange - Helper method to check if positions are within range
   * 
   * @see isInRange
   * @see gameState
   */
  highlightAttackTargets() {
    console.group('CombatManager.highlightAttackTargets');
    
    try {
      const range = this.gameState.player.equipped.weapon.range;
      const playerPos = this.gameState.player.position;
      
      console.debug('CombatManager.highlightAttackTargets: params', { range, playerPos });

      let targetCount = 0;
      this.gameState.world.objects.forEach((obj) => {
        if (obj.faction !== this.gameState.player.faction && 
            this.isInRange(playerPos, obj.position, range)) {
          this.highlightedCells.add(obj.position);
          targetCount++;
        }
      });

      if (targetCount === 0) {
        console.warn('CombatManager.highlightAttackTargets: no valid targets found in range');
      } else {
        console.info('CombatManager.highlightAttackTargets: highlighted', targetCount, 'target cells');
      }

    } catch (err) {
      console.error('CombatManager.highlightAttackTargets:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Highlights cells that are valid targets for the currently selected spell based on range
   * Adds cells within range of player's position to highlightedCells set
   * 
   * @returns {void}
   * @throws {Error} If selectedSpell is not defined
   * 
   * @requires gameState - Global game state object containing:
   *   - player.position {x: number, y: number} - Current player position
   *   - world.objects {Array} - Array of game objects with positions
   * @requires selectedSpell - Currently selected spell object with range property
   * @requires highlightedCells - Set to store highlighted cell positions
   * @requires isInRange - Helper function to check if positions are within range
   * 
   * @see isInRange
   */
  highlightSpellTargets() {
    console.group('CombatManager.highlightSpellTargets');
    
    try {
      if (!this.selectedSpell) {
        console.warn('CombatManager.highlightSpellTargets: no spell selected');
        return;
      }

      const range = this.selectedSpell.range;
      const playerPos = this.gameState.player.position;
      console.debug('CombatManager.highlightSpellTargets: params', { range, playerPos });

      let targetCount = 0;
      this.gameState.world.objects.forEach((obj) => {
        if (this.isInRange(playerPos, obj.position, range)) {
          this.highlightedCells.add(obj.position);
          targetCount++;
        }
      });

      if (targetCount === 0) {
        console.warn('CombatManager.highlightSpellTargets: no valid targets found in range');
      } else {
        console.info('CombatManager.highlightSpellTargets: highlighted', targetCount, 'target cells');
      }

    } catch (err) {
      console.error('CombatManager.highlightSpellTargets:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Highlights cells within range of the selected item based on player position.
   * Iterates through world objects and marks cells that are within the selected item's range
   * from the player's current position.
   * 
   * @requires {Object} this.selectedItem - The currently selected item with a range property
   * @requires {Object} this.gameState - The game state containing player and world data
   * @requires {Object} this.highlightedCells - Collection to store highlighted cell positions
   * @requires {Function} this.isInRange - Helper method to check if positions are within range
   * 
   * @see isInRange
   * @see GameState
   * 
   * @returns {void}
   * 
   * @example
   * combat.highlightItemTargets();
   */
  highlightItemTargets() {
    console.group('CombatManager.highlightItemTargets');
    
    try {
      if (!this.selectedItem) {
        console.warn('CombatManager.highlightItemTargets: no item selected');
        return;
      }

      const range = this.selectedItem.range;
      const playerPos = this.gameState.player.position;
      console.debug('CombatManager.highlightItemTargets: params', { range, playerPos });

      let targetCount = 0;
      this.gameState.world.objects.forEach((obj) => {
        if (this.isInRange(playerPos, obj.position, range)) {
          this.highlightedCells.add(obj.position);
          targetCount++;
        }
      });

      if (targetCount === 0) {
        console.warn('CombatManager.highlightItemTargets: no valid targets found in range');
      } else {
        console.info('CombatManager.highlightItemTargets: highlighted', targetCount, 'target cells');
      }

    } catch (err) {
      console.error('CombatManager.highlightItemTargets:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Determines if two points are within a specified Manhattan distance (L1 norm) range
   * @param {Object} from - Starting point coordinates
   * @param {number} from.x - X coordinate of starting point
   * @param {number} from.y - Y coordinate of starting point
   * @param {Object} to - Target point coordinates
   * @param {number} to.x - X coordinate of target point
   * @param {number} to.y - Y coordinate of target point 
   * @param {number} range - Maximum allowed Manhattan distance between points
   * @returns {boolean} True if points are within range, false otherwise
   *
   * Uses Manhattan distance (sum of x and y differences) rather than Euclidean distance.
   * All coordinates should be integers.
   * Range must be non-negative.
   */
  isInRange(from, to, range) {
    console.group('CombatManager.isInRange');
    
    try {
      console.debug('CombatManager.isInRange: params', { from, to, range });

      if (range < 0) {
        console.warn('CombatManager.isInRange: negative range provided');
      }

      const dx = Math.abs(to.x - from.x);
      const dy = Math.abs(to.y - from.y);
      const distance = dx + dy;

      console.info('CombatManager.isInRange: calculated distance', distance);

      return distance <= range;

    } catch (err) {
      console.error('CombatManager.isInRange:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Calculates the grid position from a mouse/touch event
   * @param {MouseEvent|TouchEvent} event - The DOM event object from the interaction
   * @returns {{x: number, y: number}} An object containing the calculated grid coordinates
   * @description Converts pixel coordinates from a mouse/touch event into grid coordinates
   * based on the renderer's tile size. The coordinates are zero-based.
   * @throws {TypeError} If event.target does not support getBoundingClientRect()
   */
  getGridPosition(event) {
    console.group('CombatManager.getGridPosition');
    
    try {
      console.debug('CombatManager.getGridPosition: params', { event });
      
      const rect = event.target.getBoundingClientRect();
      console.info('CombatManager.getGridPosition: got bounding rect', rect);

      if (event.clientX < rect.left || event.clientY < rect.top) {
        console.warn('CombatManager.getGridPosition: click outside grid bounds');
      }

      const x = Math.floor((event.clientX - rect.left) / this.renderer.tileSize);
      const y = Math.floor((event.clientY - rect.top) / this.renderer.tileSize);
      console.info('CombatManager.getGridPosition: calculated grid coords', { x, y });

      console.groupEnd();
      return { x, y };

    } catch (err) {
      console.error('CombatManager.getGridPosition:', err);
      console.groupEnd();
      throw err;
    }
  }

  /**
   * Updates the user interface elements of the combat system
   * - Disables action buttons when it's not the player's turn
   * - Updates the combat log with current turn and initiative information
   * 
   * @fires updateCombatLog - Emits event with current turn and initiative data
   * 
   * @example
   * combat.updateUI();
   * 
   * @see this.gameState.player
   * @see this.initiative
   */
  updateUI() {
    console.group('CombatManager.updateUI');
    
    try {
      console.debug('CombatManager.updateUI: starting UI update', {
        currentTurn: this.currentTurn,
        playerId: this.gameState.player.id
      });

      // Update turn indicator
      document.querySelectorAll(".action-btn").forEach((btn) => {
        const isPlayerTurn = this.currentTurn === this.gameState.player.id;
        btn.disabled = !isPlayerTurn;
      });
      console.info('CombatManager.updateUI: updated action button states');

      if (!this.initiative.length) {
        console.warn('CombatManager.updateUI: empty initiative order');
      }

      // Update combat log
      this.emit("updateCombatLog", {
        currentTurn: this.currentTurn,
        initiative: this.initiative,
      });
      console.info('CombatManager.updateUI: emitted combat log update');

    } catch (err) {
      console.error('CombatManager.updateUI:', err);
      throw err;
    } finally {
      console.groupEnd();
    }
  }
}

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/static/js/game.js`:

```js
/**
 * Represents the core game state and manages state updates and game actions.
 * Extends EventEmitter to provide event-based state change notifications.
 * 
 * Manages:
 * - Player state
 * - World state  
 * - Combat state
 * - State synchronization via RPC
 * - Game update loop
 * - Game actions (movement, combat, spells)
 *
 * Events emitted:
 * - stateChanged: When game state is updated
 * - error: When an error occurs during operations
 *
 * @class
 * @extends {EventEmitter} 
 * 
 * @property {Object} rpc - RPC client for server communication
 * @property {Object|null} player - Current player state 
 * @property {Object|null} world - Current game world state
 * @property {Object|null} combat - Current combat state
 * @property {number} lastUpdate - Timestamp of last state update
 * @property {number} updateInterval - MS between updates (default 100ms)
 * @property {boolean} initialized - Whether game is initialized
 * @property {boolean} updating - Update mutex flag
 *
 * @fires GameState#stateChanged
 * @fires GameState#error
 *
 * @see RpcClient
 */
class GameState extends EventEmitter {
  /**
   * Creates a new Game instance
   * 
   * @param {Object} rpcClient - The RPC client instance used for server communication
   * 
   * @property {Object} player - Stores the current player state
   * @property {Object} world - Stores the game world state
   * @property {Object} combat - Stores the current combat state
   * @property {number} lastUpdate - Timestamp of last game state update
   * @property {number} updateInterval - Milliseconds between updates (100ms = 10 updates/sec)
   * @property {boolean} initialized - Whether game has completed initialization
   * @property {boolean} updating - Whether an update is currently in progress
   * 
   * @extends EventEmitter
   */
  constructor(rpcClient) {
    console.group('GameState.constructor');
    console.debug('GameState.constructor: params', { rpcClient });

    super();
    this.rpc = rpcClient;
    if (!rpcClient) {
      console.warn('GameState.constructor: rpcClient not provided');
    }

    console.info('GameState.constructor: initializing state properties');
    this.player = null;
    this.world = null;
    this.combat = null;
    this.lastUpdate = 0;
    this.updateInterval = 100; // 10 updates per second
    this.initialized = false;
    this.updating = false;

    console.groupEnd();
  }

  /**
   * Initializes the game state and sets up event listeners and update loop.
   * Only initializes once - subsequent calls are ignored if already initialized.
   * 
   * Sets up:
   * - RPC state update listener
   * - Initial state fetch
   * - Game update loop
   * 
   * @async
   * @returns {Promise<void>}
   * 
   * @throws {Error} If state update or RPC setup fails
   * 
   * @see handleStateUpdate
   * @see updateState 
   * @see startUpdateLoop
   */
  async initialize() {
    console.group('GameState.initialize');
    console.debug('GameState.initialize: params', { initialized: this.initialized });

    if (this.initialized) {
      console.warn('GameState.initialize: already initialized');
      console.groupEnd();
      return;
    }

    console.info('GameState.initialize: setting up state update listener');
    this.rpc.on("stateUpdate", this.handleStateUpdate.bind(this));

    console.info('GameState.initialize: performing initial state update');
    try {
      await this.updateState();
    } catch (error) {
      console.error('GameState.initialize: failed to update state', error);
      throw error;
    }

    console.info('GameState.initialize: starting update loop');
    this.startUpdateLoop();

    console.info('GameState.initialize: setting initialized flag');
    this.initialized = true;

    console.groupEnd();
  }

  /**
   * Updates the game state by fetching the latest state via RPC and handling the update.
   * Uses a mutex flag to prevent concurrent updates.
   * 
   * @async
   * @emits {error} Emits an error event if the RPC call fails
   * @throws {Error} Propagates any errors from the RPC call
   * @returns {Promise<void>}
   * 
   * @see handleStateUpdate - Method called with the fetched state
   * @see rpc.getGameState - RPC method to fetch game state
   * 
   * State updates are synchronized using the updating flag to prevent
   * concurrent updates that could lead to race conditions.
   */
  async updateState() {
    console.group('GameState.updateState');
    console.debug('GameState.updateState: params', { updating: this.updating });
    
    if (this.updating) {
      console.warn('GameState.updateState: update already in progress');
      console.groupEnd();
      return;
    }

    this.updating = true;
    console.info('GameState.updateState: starting state update');
    
    try {
      const state = await this.rpc.getGameState();
      console.info('GameState.updateState: received new state', state);
      this.handleStateUpdate(state);
    } catch (error) {
      console.error('GameState.updateState: failed to update state', error);
      this.emit("error", error);
    } finally {
      this.updating = false;
      console.info('GameState.updateState: completed state update');
    }
    
    console.groupEnd();
  }

  /**
   * Updates the game state with new state data and emits a state change event
   * 
   * @param {Object} state - The new game state object
   * @param {Object} state.player - Updated player state
   * @param {Object} state.world - Updated world state
   * @param {Object} state.combat - Updated combat state
   * @fires stateChanged
   * 
   * @emits stateChanged - Emitted with object containing previous and current state
   * @property {Object} event.previous - Previous state before update
   * @property {Object} event.current - New state after update
   */
  handleStateUpdate(state) {
    console.group('GameState.handleStateUpdate');
    console.debug('GameState.handleStateUpdate: params', { state });

    if (!state) {
      console.warn('GameState.handleStateUpdate: received null/undefined state');
      console.groupEnd();
      return;
    }

    console.info('GameState.handleStateUpdate: saving previous state');
    const prevState = {
      player: this.player,
      world: this.world,
      combat: this.combat,
    };

    console.info('GameState.handleStateUpdate: updating state properties');
    this.player = state.player;
    this.world = state.world;
    this.combat = state.combat;

    console.info('GameState.handleStateUpdate: emitting stateChanged event');
    this.emit("stateChanged", {
      previous: prevState,
      current: state,
    });

    console.groupEnd();
  }

  /**
   * Starts the main game update loop using requestAnimationFrame
   * 
   * This method initiates a continuous loop that:
   * 1. Checks if enough time has elapsed since last update
   * 2. Updates game state if interval has passed
   * 3. Schedules next animation frame
   * 
   * The loop runs continuously until the game/component is destroyed.
   * Updates are throttled based on this.updateInterval to control 
   * update frequency.
   * 
   * Uses async/await to handle asynchronous state updates.
   * 
   * @see this.updateState - Called to update game state each interval
   * @see this.updateInterval - Time between updates in milliseconds
   * @see this.lastUpdate - Timestamp of last update
   */
  startUpdateLoop() {
    console.group('GameState.startUpdateLoop');
    console.debug('GameState.startUpdateLoop: params', { lastUpdate: this.lastUpdate, updateInterval: this.updateInterval });

    const update = async () => {
      console.group('GameState.startUpdateLoop.update');
      const now = Date.now();
      console.debug('GameState.startUpdateLoop.update: params', { now, lastUpdate: this.lastUpdate });

      if (now - this.lastUpdate >= this.updateInterval) {
        console.info('GameState.startUpdateLoop.update: executing state update');
        await this.updateState();
        this.lastUpdate = now;
        console.info('GameState.startUpdateLoop.update: updated lastUpdate timestamp', { lastUpdate: this.lastUpdate });
      } else {
        console.debug('GameState.startUpdateLoop.update: skipping update - interval not elapsed');
      }

      console.info('GameState.startUpdateLoop.update: scheduling next frame');
      requestAnimationFrame(update);
      console.groupEnd();
    };

    console.info('GameState.startUpdateLoop: starting update loop');
    update();
    console.groupEnd();
  }

  /**
   * Moves the player/entity in the specified direction
   * @param {string} direction - The direction to move ('up', 'down', 'left', 'right')
   * @returns {Promise<{success: boolean, error?: Error}>} Result object indicating if move was successful
   * @throws {Error} If RPC call fails or state update fails
   * @emits {error} If an error occurs during movement
   * @see {@link updateState}
   * @see {@link rpc.move}
   */
  async move(direction) {
    console.group('GameState.move');
    console.debug('GameState.move: params', { direction });

    try {
      console.info('GameState.move: executing move via RPC');
      const result = await this.rpc.move(direction);
      
      if (result.success) {
        console.info('GameState.move: move successful, updating state');
        await this.updateState();
      } else {
        console.warn('GameState.move: move was not successful', result);
      }

      console.groupEnd();
      return result;
    } catch (error) {
      console.error('GameState.move: error during move operation', error);
      this.emit("error", error);
      console.groupEnd();
      return { success: false, error };
    }
  }

  /**
   * Executes an attack action against a target using a specified weapon
   * 
   * @param {string|number} targetId - The unique identifier of the target to attack
   * @param {string|number} weaponId - The unique identifier of the weapon to use
   * @returns {Promise<Object>} A promise that resolves to an object containing:
   *   - success: {boolean} Whether the attack was successful
   *   - error: {Error} Error object if attack failed
   * @throws Will emit an "error" event if the RPC call fails
   * @see rpc.attack
   * @see updateState
   */
  async attack(targetId, weaponId) {
    console.group('GameState.attack');
    console.debug('GameState.attack: params', { targetId, weaponId });

    try {
      console.info('GameState.attack: executing attack via RPC');
      const result = await this.rpc.attack(targetId, weaponId);
      
      if (result.success) {
        console.info('GameState.attack: attack successful, updating state');
        await this.updateState();
      } else {
        console.warn('GameState.attack: attack was not successful', result);
      }

      console.groupEnd();
      return result;
    } catch (error) {
      console.error('GameState.attack: error during attack operation', error);
      this.emit("error", error);
      console.groupEnd();
      return { success: false, error };
    }
  }

  /**
   * Casts a spell on a target or at a position in the game
   * 
   * @async
   * @param {string|number} spellId - The unique identifier of the spell to cast
   * @param {string|number} targetId - The unique identifier of the target entity (optional if position is provided)
   * @param {Object} position - The x,y coordinates to cast the spell at (optional if targetId is provided)
   * @returns {Promise<Object>} Result object containing:
   *                           - success: boolean indicating if spell was cast successfully
   *                           - error: Error object if spell casting failed
   * @throws {Error} If RPC call fails, error is emitted via 'error' event
   * @see {@link RPC#castSpell} For the underlying RPC implementation
   */
  async castSpell(spellId, targetId, position) {
    console.group('GameState.castSpell');
    console.debug('GameState.castSpell: params', { spellId, targetId, position });

    try {
      console.info('GameState.castSpell: executing spell cast via RPC');
      const result = await this.rpc.castSpell(spellId, targetId, position);
      
      if (result.success) {
        console.info('GameState.castSpell: spell cast successful, updating state');
        await this.updateState();
      } else {
        console.warn('GameState.castSpell: spell cast was not successful', result);
      }

      console.groupEnd();
      return result;
    } catch (error) {
      console.error('GameState.castSpell: error during spell cast operation', error);
      this.emit("error", error);
      console.groupEnd();
      return { success: false, error };
    }
  }

  /**
   * Ends the current turn in the game by making an RPC call and updates the game state
   * if successful.
   * 
   * @async
   * @returns {Promise<Object>} A promise that resolves to an object containing:
   *   - success {boolean} - Whether the turn was ended successfully
   *   - error {Error} [optional] - Error object if the operation failed
   * 
   * @fires error - Emitted when an error occurs during the turn end operation
   * 
   * @throws {Error} - Any error that occurs during the RPC call or state update
   * will be caught, emitted as an 'error' event, and returned in the result object
   * 
   * @see {@link updateState} - Called to refresh game state after successful turn end
   * @see {@link rpc.endTurn} - The RPC method called to end the turn
   */
  async endTurn() {
    console.group('GameState.endTurn');
    console.debug('GameState.endTurn: params', {});

    try {
      console.info('GameState.endTurn: executing turn end via RPC');
      const result = await this.rpc.endTurn();
      
      if (result.success) {
        console.info('GameState.endTurn: turn end successful, updating state');
        await this.updateState();
      } else {
        console.warn('GameState.endTurn: turn end was not successful', result);
      }

      console.groupEnd();
      return result;
    } catch (error) {
      console.error('GameState.endTurn: error during turn end operation', error);
      this.emit("error", error);
      console.groupEnd();
      return { success: false, error };
    }
  }
}

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/static/js/rpc.js`:

```js
/**
 * EventEmitter class provides a basic pub/sub pattern implementation.
 * Allows registration and triggering of event handlers.
 *
 * @class
 * @classdesc Implementation of the observer pattern with event emission and handling
 * @example
 * const emitter = new EventEmitter();
 * emitter.on('event', data => console.log(data));
 * emitter.emit('event', 'hello'); // logs: hello
 *
 * @property {Map<string, Function[]>} events - Map storing event name to array of handlers
 *
 * @see {@link EventEmitter#on} - Method to register event handlers
 * @see {@link EventEmitter#emit} - Method to trigger events
 *
 * @notes
 * - All event handlers are stored in a Map for O(1) lookup
 * - Handlers for a single event are stored in arrays and executed in order
 * - No built-in error handling for invalid inputs
 * - Does not support removing event listeners
 * - Does not handle circular event dependencies
 */
class EventEmitter {
  /**
   * Initializes a new instance with an empty events Map.
   * Used to store event mappings for pub/sub functionality.
   * 
   * @constructor
   * @memberof RPC
   * @class
   */
  constructor() {
    this.events = new Map();
  }

  /**
   * Registers a callback function for a specific event
   * @param {string} event - The name of the event to listen for
   * @param {Function} callback - The function to execute when the event occurs
   * @description Creates an array of callbacks for the event if it doesn't exist,
   * then adds the new callback to that array
   */
  on(event, callback) {
    if (!this.events.has(event)) {
      this.events.set(event, []);
    }
    this.events.get(event).push(callback);
  }

  /**
   * Emits an event with the provided data to all registered callbacks for that event
   * 
   * @param {string} event - The name of the event to emit
   * @param {*} data - The data to pass to the event callbacks
   * @fires event
   * 
   * @example
   * // Emit a 'message' event with data
   * emit('message', {text: 'Hello'});
   * 
   * @notes
   * - If the event doesn't exist in the events Map, no callbacks will be executed
   * - Each registered callback is executed with the provided data
   * - Callbacks are executed synchronously in registration order
   * 
   * @see {@link this.events} - Map storing event callbacks
   */
  emit(event, data) {
    if (this.events.has(event)) {
      this.events.get(event).forEach((cb) => cb(data));
    }
  }
}

/**
 * A WebSocket-based RPC client that handles communication with a game server
 * implementing JSON-RPC 2.0 protocol.
 * 
 * @class
 * @extends {EventEmitter}
 * 
 * @description
 * Provides a high-level interface for making RPC calls to a game server with:
 * - Automatic WebSocket connection management and reconnection
 * - Request/response tracking with timeouts
 * - Session management
 * - Game-specific method wrappers (move, attack, spell casting, etc.)
 * 
 * @property {string} baseUrl - Base URL for RPC endpoint, defaults to "./rpc"
 * @property {WebSocket} ws - WebSocket connection instance
 * @property {string} sessionId - Unique session identifier for the current player
 * @property {Map<number, {resolve: Function, reject: Function}>} requestQueue - Pending request callbacks
 * @property {number} requestId - Auto-incrementing counter for generating unique request IDs
 * @property {number} reconnectAttempts - Number of connection retry attempts made
 * @property {number} maxReconnectAttempts - Maximum number of retry attempts allowed (default: 5)
 * 
 * @fires RPCClient#connected - When WebSocket connection is established
 * @fires RPCClient#disconnected - When WebSocket connection is lost
 * @fires RPCClient#error - When a WebSocket or request error occurs
 * 
 * @example
 * ```js
 * const rpc = new RPCClient();
 * await rpc.connect();
 * await rpc.joinGame("Player1");
 * const gameState = await rpc.getGameState();
 * ```
 * 
 * @see {@link https://www.jsonrpc.org/specification|JSON-RPC 2.0 Specification}
 * @see {@link WebSocket|WebSocket API}
 * 
 * @throws {Error} If WebSocket connection fails after maximum retry attempts
 * @throws {Error} If requests timeout or fail to send
 */
class RPCClient extends EventEmitter {
  /**
   * Creates a new RPC client instance with WebSocket capabilities
   * @class
   * @extends {EventEmitter} 
   * @description Initializes an RPC client that handles WebSocket connections and request queueing
   * @property {string} baseUrl - Base URL for RPC endpoint, defaults to "./rpc"
   * @property {WebSocket} ws - WebSocket connection instance
   * @property {string} sessionId - Unique session identifier 
   * @property {Map} requestQueue - Queue storing pending RPC requests
   * @property {number} requestId - Counter for generating unique request IDs
   * @property {number} reconnectAttempts - Number of connection retry attempts
   * @property {number} maxReconnectAttempts - Maximum number of retry attempts allowed
   * @throws {Error} If WebSocket connection fails after max retry attempts
   * @see {@link handleWebSocketMessage} For WebSocket message handling
   * @see {@link reconnect} For reconnection logic
   */
  constructor() {
    console.group('RPCClient.constructor: Initializing');
    
    try {
      super();
      console.debug('RPCClient.constructor: Setting up base properties');
      
      this.baseUrl = "./rpc";
      console.info('RPCClient.constructor: Base URL set to', this.baseUrl);
      
      this.ws = null;
      this.sessionId = null;
      console.info('RPCClient.constructor: WebSocket and session initialized to null');
      
      this.requestQueue = new Map();
      this.requestId = 1;
      console.info('RPCClient.constructor: Request tracking initialized');
      
      this.reconnectAttempts = 0;
      this.maxReconnectAttempts = 5;
      console.info('RPCClient.constructor: Reconnect settings configured', {
        maxAttempts: this.maxReconnectAttempts
      });
    } catch (error) {
      console.error('RPCClient.constructor: Failed to initialize:', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Establishes a WebSocket connection to the RPC server endpoint.
   * Sets up WebSocket event handlers and emits a 'connected' event on success.
   * 
   * @async
   * @fires connected
   * @throws {Error} If connection fails after max reconnection attempts
   * @see {@link setupWebSocket} For WebSocket event handling setup
   * @see {@link waitForConnection} For connection promise resolution
   * @see {@link handleConnectionError} For error handling
   * @returns {Promise<void>} Resolves when connection is established successfully
   */
  async connect() {
    console.group('RPCClient.connect: Establishing WebSocket connection');
    
    try {
      console.debug('RPCClient.connect: Using WebSocket URL', `ws://${location.host}/rpc/ws`);
      
      this.ws = new WebSocket(`ws://${location.host}/rpc/ws`);
      console.info('RPCClient.connect: WebSocket instance created');
      
      this.setupWebSocket();
      console.info('RPCClient.connect: WebSocket handlers configured');
      
      await this.waitForConnection();
      console.info('RPCClient.connect: Connection established');
      
      this.reconnectAttempts = 0;
      console.info('RPCClient.connect: Reset reconnect attempts to 0');
      
      this.emit("connected");
      console.info('RPCClient.connect: Connected event emitted');
      
    } catch (error) {
      console.error('RPCClient.connect: Connection failed:', error);
      this.handleConnectionError(error);
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Sends a JSON-RPC 2.0 request over WebSocket with timeout handling
   * 
   * @param {string} method - The RPC method name to call
   * @param {Object} [params={}] - Parameters to pass to the RPC method
   * @param {number} [timeout=5000] - Request timeout in milliseconds
   * @returns {Promise<*>} Promise that resolves with the RPC response result
   * @throws {Error} If the request times out or WebSocket send fails
   *
   * @description
   * This method handles sending JSON-RPC 2.0 formatted requests over a WebSocket connection.
   * It adds request ID tracking and session ID to each request.
   * The request is tracked in a queue and will reject if timeout is reached.
   * 
   * @example
   * ```js
   * // Send RPC request with 3 second timeout
   * const result = await rpc.request('methodName', {param: 'value'}, 3000);
   * ```
   */
  async request(method, params = {}, timeout = 5000) {
    console.group('RPCClient.request: Processing RPC request');
    
    try {
      const id = this.requestId++;
      console.debug('RPCClient.request: Request parameters', { method, params, timeout, id });

      const message = {
        jsonrpc: "2.0",
        method,
        params: { ...params, session_id: this.sessionId },
        id,
      };
      console.info('RPCClient.request: Formed JSON-RPC message', message);

      return new Promise((resolve, reject) => {
        const timeoutId = setTimeout(() => {
          console.warn('RPCClient.request: Request timed out', { method, id });
          this.requestQueue.delete(id);
          reject(new Error(`Request timeout: ${method}`));
        }, timeout);

        console.info('RPCClient.request: Adding to request queue', { id });
        this.requestQueue.set(id, {
          resolve: (result) => {
            console.info('RPCClient.request: Request resolved', { id, result });
            clearTimeout(timeoutId);
            resolve(result);
          },
          reject: (error) => {
            console.error('RPCClient.request: Request rejected', { id, error });
            clearTimeout(timeoutId);
            reject(error);
          },
        });

        try {
          console.debug('RPCClient.request: Sending WebSocket message');
          this.ws.send(JSON.stringify(message));
        } catch (error) {
          console.error('RPCClient.request: Failed to send message', error);
          clearTimeout(timeoutId);
          this.requestQueue.delete(id);
          reject(error);
        }
      });
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Sets up WebSocket event handlers for message handling, connection closure and errors
   * Binds the event handler methods to the current instance context using .bind(this)
   * 
   * Related:
   * @see handleMessage - Processes incoming WebSocket messages
   * @see handleClose - Handles WebSocket connection closures  
   * @see handleError - Handles WebSocket errors
   * 
   * @throws {Error} If WebSocket is not initialized or invalid
   */
  setupWebSocket() {
    console.group('RPCClient.setupWebSocket: Setting up WebSocket handlers');
    
    try {
      console.debug('RPCClient.setupWebSocket: Binding message handler');
      this.ws.onmessage = this.handleMessage.bind(this);
      
      console.debug('RPCClient.setupWebSocket: Binding close handler');
      this.ws.onclose = this.handleClose.bind(this);
      
      console.debug('RPCClient.setupWebSocket: Binding error handler');
      this.ws.onerror = this.handleError.bind(this);
      
      console.info('RPCClient.setupWebSocket: All handlers bound successfully');
    } catch (error) {
      console.error('RPCClient.setupWebSocket: Failed to setup handlers:', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Sends a move request to the server with the specified direction
   * 
   * @param {string} direction - The direction to move ('up'|'down'|'left'|'right')
   * @returns {Promise<Object>} Response from the server containing the result of the move
   * @throws {Error} If the server request fails or returns an error
   * @see request - Base RPC request method
   */
  async move(direction) {
    console.group('RPCClient.move: Processing move request');
    try {
      console.debug('RPCClient.move: Direction parameter', { direction });
      const result = await this.request("move", { direction });
      console.info('RPCClient.move: Move request completed', { result });
      return result;
    } catch (error) {
      console.error('RPCClient.move: Failed to process move request', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Initiates an attack action against a target using a specified weapon
   * 
   * @param {string|number} targetId - The unique identifier of the target to attack
   * @param {string|number} weaponId - The unique identifier of the weapon to use
   * @returns {Promise<Object>} The response from the server containing the attack result
   * @throws {Error} If the request fails or returns an error
   * @see request - The base request method used to send the RPC call
   */
  async attack(targetId, weaponId) {
    console.group('RPCClient.attack: Processing attack request');
    try {
      console.debug('RPCClient.attack: Attack parameters', { targetId, weaponId });
      const result = await this.request("attack", { target_id: targetId, weapon_id: weaponId });
      console.info('RPCClient.attack: Attack request completed', { result });
      return result;
    } catch (error) {
      console.error('RPCClient.attack: Failed to process attack request', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Sends a request to cast a spell at a target or position
   * 
   * @param {number|string} spellId - Unique identifier for the spell to cast
   * @param {number|string} targetId - Unique identifier of the target entity (optional if position is provided)
   * @param {Object} position - Target position coordinates (optional if targetId is provided)
   * @param {number} position.x - X coordinate
   * @param {number} position.y - Y coordinate
   * @returns {Promise<Object>} Response from the server with spell cast results
   * 
   * @throws Will throw an error if neither targetId nor position is provided
   * @throws Will throw an error if the spell casting request fails
   * 
   * @see SpellSystem.handleCastSpell - Server-side spell handling
   * @see Spell - Spell entity definition
   */
  async castSpell(spellId, targetId, position) {
    console.group('RPCClient.castSpell: Processing spell cast request');
    try {
      console.debug('RPCClient.castSpell: Spell parameters', { spellId, targetId, position });
      const result = await this.request("castSpell", {
        spell_id: spellId,
        target_id: targetId,
        position,
      });
      console.info('RPCClient.castSpell: Spell cast completed', { result });
      return result;
    } catch (error) {
      console.error('RPCClient.castSpell: Failed to cast spell', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Initiates a combat sequence with specified participants
   * 
   * @param {Array<string|number>} participantIds - Array of unique identifiers for combat participants
   * @returns {Promise<Object>} A promise that resolves to the combat session details
   * 
   * @throws {Error} If the RPC request fails or participant IDs are invalid
   * @see {@link request} For the underlying RPC implementation
   */
  async startCombat(participantIds) {
    console.group('RPCClient.startCombat: Processing combat start request');
    try {
      console.debug('RPCClient.startCombat: Combat parameters', { participantIds });
      const result = await this.request("startCombat", { participant_ids: participantIds });
      console.info('RPCClient.startCombat: Combat started successfully', { result });
      return result;
    } catch (error) {
      console.error('RPCClient.startCombat: Failed to start combat', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Ends the current player's turn in the game.
   * Makes an RPC request to the server to process turn completion.
   * 
   * @async
   * @returns {Promise<*>} Result from the server after ending the turn
   * @throws {Error} If the RPC request fails
   * @see request - The underlying RPC request method used
   */
  async endTurn() {
    console.group('RPCClient.endTurn: Processing end turn request');
    try {
      console.debug('RPCClient.endTurn: Sending end turn request');
      const result = await this.request("endTurn");
      console.info('RPCClient.endTurn: Turn ended successfully', { result });
      return result;
    } catch (error) {
      console.error('RPCClient.endTurn: Failed to end turn', error);
      throw error;  
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Retrieves the current game state from the server
   * 
   * @async
   * @returns {Promise<Object>} A promise that resolves with the game state object
   * @throws {Error} If the server request fails
   * @see request - The underlying RPC request method used
   */
  async getGameState() {
    console.group('RPCClient.getGameState: Fetching game state');
    try {
      console.debug('RPCClient.getGameState: Making request');
      const result = await this.request("getGameState");
      console.info('RPCClient.getGameState: State retrieved successfully', { result });
      return result;
    } catch (error) {
      console.error('RPCClient.getGameState: Failed to retrieve game state', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Joins a game session with the specified player name
   * 
   * @param {string} playerName - The name of the player joining the game. Must be non-empty.
   * @returns {Promise<Object>} A promise that resolves to the game join result containing:
   *                           - session_id: Unique identifier for the player session
   * @throws {Error} If the RPC request fails or returns an error
   * @see request - For the underlying RPC implementation
   */
  async joinGame(playerName) {
    console.group('RPCClient.joinGame: Processing join game request');
    try {
      console.debug('RPCClient.joinGame: Player name parameter', { playerName });
      const result = await this.request("joinGame", { player_name: playerName });
      console.info('RPCClient.joinGame: Session ID set', { sessionId: result.session_id });
      this.sessionId = result.session_id;
      return result;
    } catch (error) {
      console.error('RPCClient.joinGame: Failed to join game', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Leaves the current game session by making a request to the server
   * and clearing the local session ID.
   * 
   * @async
   * @returns {Promise<void>} A promise that resolves when the leave game request completes
   * @throws {Error} If the request to leave game fails
   * 
   * @remarks
   * - Only attempts to leave if there is an active session (sessionId exists)
   * - Cleans up the session state by setting sessionId to null after leaving
   * - Uses the request() method to make the server call
   * 
   * @see request
   */
  async leaveGame() {
    console.group('RPCClient.leaveGame: Processing leave game request');
    try {
      if (this.sessionId) {
        console.debug('RPCClient.leaveGame: Current session ID', { sessionId: this.sessionId });
        await this.request("leaveGame");
        console.info('RPCClient.leaveGame: Successfully left game');
        this.sessionId = null;
        console.info('RPCClient.leaveGame: Session ID cleared');
      } else {
        console.warn('RPCClient.leaveGame: No active session to leave');
      }
    } catch (error) {
      console.error('RPCClient.leaveGame: Failed to leave game', error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }
}

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/web/static/js/render.js`:

```js
class GameRenderer {
  constructor() {
    console.group('Constructor: GameRenderer initialization');
    console.debug('Constructor: Getting canvas elements');
    
    this.terrainLayer = document.getElementById("terrain-layer");
    this.objectLayer = document.getElementById("object-layer");
    this.effectLayer = document.getElementById("effect-layer");

    if (!this.terrainLayer || !this.objectLayer || !this.effectLayer) {
      console.error('Constructor: Failed to get one or more canvas elements');
      throw new Error('Canvas elements not found');
    }

    console.info('Constructor: Setting up canvas contexts');
    this.terrainCtx = this.terrainLayer.getContext("2d");
    this.objectCtx = this.objectLayer.getContext("2d");
    this.effectCtx = this.effectLayer.getContext("2d");

    console.info('Constructor: Initializing core properties');
    this.tileSize = 32;
    this.sprites = new Map();
    this.animations = new Map();

    this.camera = {
      x: 0,
      y: 0,
      zoom: 1,
    };

    console.debug('Constructor: Setting up resize event listener');
    window.addEventListener("resize", this.handleResize.bind(this));
    
    console.info('Constructor: Performing initial resize');
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
    console.group('loadSprites: Loading sprite assets');
    const spriteUrls = {
      terrain: "./static/assets/sprites/terrain.png",
      characters: "./static/assets/sprites/characters.png",
      effects: "./static/assets/sprites/effects.png",
      ui: "./static/assets/sprites/ui.png",
    };
    console.debug('loadSprites: Sprite URLs to load:', spriteUrls);

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
            console.error(`loadSprites: Failed to load sprite "${key}" from ${url}`);
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
    console.group('handleResize: Resizing canvas layers');
    
    const container = document.getElementById("viewport-container");
    if (!container) {
      console.error('handleResize: Viewport container not found');
      console.groupEnd();
      return;
    }

    const width = container.clientWidth;
    const height = container.clientHeight;
    console.debug('handleResize: Container dimensions', { width, height });

    if (width === 0 || height === 0) {
      console.warn('handleResize: Container has zero dimension');
    }

    [this.terrainLayer, this.objectLayer, this.effectLayer].forEach(
      (canvas) => {
        if (!canvas) {
          console.error('handleResize: Canvas layer is null');
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
    console.group('clearLayers: Clearing all canvas layers');
    console.debug('clearLayers: Canvas contexts:', [this.terrainCtx, this.objectCtx, this.effectCtx]);

    [this.terrainCtx, this.objectCtx, this.effectCtx].forEach((ctx) => {
      if (!ctx) {
        console.error('clearLayers: Missing context:', ctx);
        return;
      }
      console.info(`clearLayers: Clearing canvas of size ${ctx.canvas.width}x${ctx.canvas.height}`);
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
    console.group('drawSprite: Drawing sprite to canvas');
    console.debug('drawSprite: Parameters:', { spriteName, sx, sy, dx, dy, width, height });

    const sprite = this.sprites.get(spriteName);
    if (!sprite) {
      console.error('drawSprite: Sprite not found:', spriteName);
      console.groupEnd();
      return;
    }

    if (width !== this.tileSize || height !== this.tileSize) {
      console.warn('drawSprite: Non-standard tile dimensions used:', { width, height });
    }

    console.info('drawSprite: Drawing image with dimensions:', {
      sourceX: sx * this.tileSize,
      sourceY: sy * this.tileSize,
      sourceWidth: this.tileSize,
      sourceHeight: this.tileSize,
      destX: dx,
      destY: dy,
      destWidth: width,
      destHeight: height
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
    console.group('render: Rendering game state');
    console.debug('render: Game state:', gameState);

    if (!gameState) {
      console.error('render: Game state is null or undefined');
      console.groupEnd();
      return;
    }

    if (!gameState.world) {
      console.warn('render: World data is missing from game state');
    }

    this.clearLayers();
    console.info('render: Cleared all canvas layers');

    this.renderTerrain(gameState.world?.map);
    console.info('render: Rendered terrain layer');

    this.renderObjects(gameState.world?.objects);
    console.info('render: Rendered objects layer');

    this.renderEffects(gameState.world?.effects);
    console.info('render: Rendered effects layer');

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
    console.group('renderTerrain: Rendering terrain layer');
    console.debug('renderTerrain: Map data:', map);

    if (!map) {
      console.error('renderTerrain: Map is null or undefined');
      console.groupEnd();
      return;
    }

    const viewportWidth = Math.ceil(this.terrainLayer.width / this.tileSize);
    const viewportHeight = Math.ceil(this.terrainLayer.height / this.tileSize);
    console.info('renderTerrain: Viewport dimensions:', { viewportWidth, viewportHeight });

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
            console.warn('renderTerrain: Missing tile data at', { worldX, worldY });
            continue;
          }
          
          console.debug('renderTerrain: Drawing tile:', {
            worldX,
            worldY,
            spriteX: tile.spriteX,
            spriteY: tile.spriteY
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
          console.warn('renderTerrain: Tile position out of bounds:', { worldX, worldY });
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
    console.group('renderObjects: Rendering object layer');
    console.debug('renderObjects: Objects array:', objects);

    if (!objects) {
      console.warn('renderObjects: Objects array is null or undefined');
      console.groupEnd();
      return;
    }

    objects.forEach((obj) => {
      const screenX = (obj.x - this.camera.x) * this.tileSize;
      const screenY = (obj.y - this.camera.y) * this.tileSize;
      
      console.debug('renderObjects: Calculated screen coordinates:', { screenX, screenY });

      if (this.isOnScreen(screenX, screenY)) {
        console.info('renderObjects: Drawing object:', {
          x: obj.x,
          y: obj.y,
          spriteX: obj.spriteX,
          spriteY: obj.spriteY
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
        console.warn('renderObjects: Object outside viewport:', { screenX, screenY });
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
    console.group('renderEffects: Rendering effects layer');
    console.debug('renderEffects: Effects array:', effects);

    if (!effects) {
      console.warn('renderEffects: Effects array is null or undefined');
      console.groupEnd();
      return;
    }

    effects.forEach((effect) => {
      const screenX = (effect.x - this.camera.x) * this.tileSize;
      const screenY = (effect.y - this.camera.y) * this.tileSize;
      
      console.debug('renderEffects: Calculated screen coordinates:', { screenX, screenY });

      if (this.isOnScreen(screenX, screenY)) {
        console.info('renderEffects: Drawing effect:', {
          x: effect.x,
          y: effect.y,
          spriteX: effect.spriteX,
          spriteY: effect.spriteY
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
        console.warn('renderEffects: Effect outside viewport:', { screenX, screenY });
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
    console.group('updateHighlights: Updating highlight effects');
    console.debug('updateHighlights: Cells array:', cells);

    if (!cells) {
      console.error('updateHighlights: Cells array is null or undefined');
      console.groupEnd();
      return;
    }

    console.info('updateHighlights: Clearing effect layer');
    this.effectCtx.clearRect(
      0,
      0,
      this.effectLayer.width,
      this.effectLayer.height,
    );

    cells.forEach((pos) => {
      const screenX = (pos.x - this.camera.x) * this.tileSize;
      const screenY = (pos.y - this.camera.y) * this.tileSize;
      
      console.debug('updateHighlights: Calculated screen coordinates:', { screenX, screenY });

      if (this.isOnScreen(screenX, screenY)) {
        console.info('updateHighlights: Drawing highlight at:', { screenX, screenY });
        this.effectCtx.fillStyle = "rgba(255, 255, 0, 0.3)";
        this.effectCtx.fillRect(screenX, screenY, this.tileSize, this.tileSize);
      } else {
        console.warn('updateHighlights: Cell outside viewport:', { screenX, screenY });
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
    console.group('isOnScreen: Checking if coordinate is within viewport');
    console.debug('isOnScreen: Checking coordinates:', { x, y });

    if (x < -this.tileSize || y < -this.tileSize) {
      console.warn('isOnScreen: Coordinates below minimum bounds');
    }

    if (x > this.objectLayer.width || y > this.objectLayer.height) {
      console.warn('isOnScreen: Coordinates exceed viewport dimensions');
    }

    const result = (
      x >= -this.tileSize &&
      y >= -this.tileSize &&
      x <= this.objectLayer.width &&
      y <= this.objectLayer.height
    );

    console.info('isOnScreen: Visibility check result:', result);
    console.groupEnd();
    return result;
  }
}

```