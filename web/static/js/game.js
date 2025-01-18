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
  }

  async initialize() {
    if (this.initialized) return;

    this.rpc.on("stateUpdate", this.handleStateUpdate.bind(this));
    await this.updateState();
    this.startUpdateLoop();
    this.initialized = true;
  }

  async updateState() {
    try {
      const state = await this.rpc.getGameState();
      this.handleStateUpdate(state);
    } catch (error) {
      this.emit("error", error);
    }
  }

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

  // Game action methods
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
