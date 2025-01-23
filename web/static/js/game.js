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
    super();
    this.rpc = rpcClient;
    this.player = null;
    this.world = null;
    this.combat = null;
    this.lastUpdate = 0;
    this.updateInterval = 100; // 10 updates per second
    this.initialized = false;
    this.updating = false;
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
    if (this.initialized) return;

    this.rpc.on("stateUpdate", this.handleStateUpdate.bind(this));
    await this.updateState();
    this.startUpdateLoop();
    this.initialized = true;
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
    if (this.updating) return;
    this.updating = true;
    try {
      const state = await this.rpc.getGameState();
      this.handleStateUpdate(state);
    } catch (error) {
      this.emit("error", error);
    } finally {
      this.updating = false;
    }
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
    const prevState = {
      player: this.player,
      world: this.world,
      combat: this.combat,
    };

    this.player = state.player;
    this.world = state.world;
    this.combat = state.combat;

    this.emit("stateChanged", {
      previous: prevState,
      current: state,
    });
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
    const update = async () => {
      const now = Date.now();
      if (now - this.lastUpdate >= this.updateInterval) {
        await this.updateState();
        this.lastUpdate = now;
      }
      requestAnimationFrame(update);
    };
    update();
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
    try {
      const result = await this.rpc.move(direction);
      if (result.success) {
        await this.updateState();
      }
      return result;
    } catch (error) {
      this.emit("error", error);
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
    try {
      const result = await this.rpc.attack(targetId, weaponId);
      if (result.success) {
        await this.updateState();
      }
      return result;
    } catch (error) {
      this.emit("error", error);
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
    try {
      const result = await this.rpc.castSpell(spellId, targetId, position);
      if (result.success) {
        await this.updateState();
      }
      return result;
    } catch (error) {
      this.emit("error", error);
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
    try {
      const result = await this.rpc.endTurn();
      if (result.success) {
        await this.updateState();
      }
      return result;
    } catch (error) {
      this.emit("error", error);
      return { success: false, error };
    }
  }
}
