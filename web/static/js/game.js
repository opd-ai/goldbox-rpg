class GameState extends EventEmitter {
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
