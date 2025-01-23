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
    super();
    this.baseUrl = "./rpc";
    this.ws = null;
    this.sessionId = null;
    this.requestQueue = new Map();
    this.requestId = 1;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
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
    this.ws.onmessage = this.handleMessage.bind(this);
    this.ws.onclose = this.handleClose.bind(this);
    this.ws.onerror = this.handleError.bind(this);
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
    return this.request("move", { direction });
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
    return this.request("attack", { target_id: targetId, weapon_id: weaponId });
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
    return this.request("castSpell", {
      spell_id: spellId,
      target_id: targetId,
      position,
    });
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
    return this.request("startCombat", { participant_ids: participantIds });
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
    return this.request("endTurn");
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
    return this.request("getGameState");
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
    const result = await this.request("joinGame", { player_name: playerName });
    this.sessionId = result.session_id;
    return result;
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
    if (this.sessionId) {
      await this.request("leaveGame");
      this.sessionId = null;
    }
  }
}
