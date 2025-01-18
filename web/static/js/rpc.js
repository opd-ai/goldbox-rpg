class EventEmitter {
  constructor() {
    this.events = new Map();
  }

  on(event, callback) {
    if (!this.events.has(event)) {
      this.events.set(event, []);
    }
    this.events.get(event).push(callback);
  }

  emit(event, data) {
    if (this.events.has(event)) {
      this.events.get(event).forEach((cb) => cb(data));
    }
  }
}

class RPCClient extends EventEmitter {
  constructor() {
    super();
    this.baseUrl = "./rpc";
    this.ws = null;
    this.sessionId = null;
    this.requestQueue = new Map();
    this.requestId = 1;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  async connect() {
    try {
      this.ws = new WebSocket(`ws://${location.host}/rpc/ws`);
      this.setupWebSocket();
      await this.waitForConnection();
      this.reconnectAttempts = 0;
      this.emit("connected");
    } catch (error) {
      this.handleConnectionError(error);
    }
  }

  async request(method, params = {}, timeout = 5000) {
    const id = this.requestId++;
    const message = {
      jsonrpc: "2.0",
      method,
      params: { ...params, session_id: this.sessionId },
      id,
    };

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.requestQueue.delete(id);
        reject(new Error(`Request timeout: ${method}`));
      }, timeout);

      this.requestQueue.set(id, {
        resolve: (result) => {
          clearTimeout(timeoutId);
          resolve(result);
        },
        reject: (error) => {
          clearTimeout(timeoutId);
          reject(error);
        },
      });

      try {
        this.ws.send(JSON.stringify(message));
      } catch (error) {
        clearTimeout(timeoutId);
        this.requestQueue.delete(id);
        reject(error);
      }
    });
  }

  setupWebSocket() {
    this.ws.onmessage = this.handleMessage.bind(this);
    this.ws.onclose = this.handleClose.bind(this);
    this.ws.onerror = this.handleError.bind(this);
  }

  // RPC Methods from README-RPC.md
  async move(direction) {
    return this.request("move", { direction });
  }

  async attack(targetId, weaponId) {
    return this.request("attack", { target_id: targetId, weapon_id: weaponId });
  }

  async castSpell(spellId, targetId, position) {
    return this.request("castSpell", {
      spell_id: spellId,
      target_id: targetId,
      position,
    });
  }

  async startCombat(participantIds) {
    return this.request("startCombat", { participant_ids: participantIds });
  }

  async endTurn() {
    return this.request("endTurn");
  }

  async getGameState() {
    return this.request("getGameState");
  }

  async joinGame(playerName) {
    const result = await this.request("joinGame", { player_name: playerName });
    this.sessionId = result.session_id;
    return result;
  }

  async leaveGame() {
    if (this.sessionId) {
      await this.request("leaveGame");
      this.sessionId = null;
    }
  }
}
