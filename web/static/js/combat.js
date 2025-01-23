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
    document.querySelectorAll(".action-btn").forEach((btn) => {
      btn.removeEventListener("click", this.handleActionButton);
    });
    this.highlightedCells.clear();
    this.renderer.updateHighlights(this.highlightedCells);
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
    if (!this.active || this.currentTurn !== this.gameState.player.id) return;

    // Clear previous state
    this.selectedAction = null;
    this.highlightedCells.clear();
    this.renderer.updateHighlights(this.highlightedCells);

    // Set new state
    this.selectedAction = action;
    await this.highlightValidTargets(action);
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
    if (!this.selectedSpell) return;
    const range = this.selectedSpell.range;
    const playerPos = this.gameState.player.position;

    this.gameState.world.objects.forEach((obj) => {
      if (this.isInRange(playerPos, obj.position, range)) {
        this.highlightedCells.add(obj.position);
      }
    });
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
    if (!this.selectedItem) return;
    const range = this.selectedItem.range;
    const playerPos = this.gameState.player.position;

    this.gameState.world.objects.forEach((obj) => {
      if (this.isInRange(playerPos, obj.position, range)) {
        this.highlightedCells.add(obj.position);
      }
    });
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
    const dx = Math.abs(to.x - from.x);
    const dy = Math.abs(to.y - from.y);
    return dx + dy <= range;
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
    const rect = event.target.getBoundingClientRect();
    const x = Math.floor((event.clientX - rect.left) / this.renderer.tileSize);
    const y = Math.floor((event.clientY - rect.top) / this.renderer.tileSize);
    return { x, y };
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
