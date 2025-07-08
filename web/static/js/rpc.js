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

  /**
   * Removes a specific callback from an event
   *
   * @param {string} event - The name of the event
   * @param {Function} callback - The specific callback function to remove
   * @returns {boolean} - True if the callback was found and removed, false otherwise
   *
   * @example
   * // Remove a specific callback
   * const myCallback = (data) => console.log(data);
   * eventEmitter.on('test', myCallback);
   * eventEmitter.off('test', myCallback); // Returns true
   *
   * @notes
   * - Uses strict equality (===) to match the callback function
   * - If the event has no remaining callbacks after removal, the event entry is cleaned up
   */
  off(event, callback) {
    if (!this.events.has(event)) {
      return false;
    }

    const callbacks = this.events.get(event);
    const index = callbacks.indexOf(callback);
    
    if (index === -1) {
      return false;
    }

    callbacks.splice(index, 1);

    // Clean up empty event arrays to prevent memory leaks
    if (callbacks.length === 0) {
      this.events.delete(event);
    }

    return true;
  }

  /**
   * Removes all callbacks for a specific event
   *
   * @param {string} event - The name of the event to clear
   * @returns {boolean} - True if the event existed and was cleared, false otherwise
   *
   * @example
   * // Clear all callbacks for an event
   * eventEmitter.removeAllListeners('test');
   *
   * @notes
   * - Completely removes the event from the events Map
   * - Prevents memory leaks from accumulated event listeners
   */
  removeAllListeners(event) {
    if (!this.events.has(event)) {
      return false;
    }

    this.events.delete(event);
    return true;
  }

  /**
   * Removes all events and callbacks, completely clearing the EventEmitter
   *
   * @example
   * // Clear everything
   * eventEmitter.clear();
   *
   * @notes
   * - Use this method when disposing of an EventEmitter instance
   * - Essential for preventing memory leaks in long-running applications
   */
  clear() {
    this.events.clear();
  }

  /**
   * Gets the number of listeners for a specific event
   *
   * @param {string} event - The name of the event
   * @returns {number} - The number of callbacks registered for the event
   *
   * @example
   * // Check listener count
   * const count = eventEmitter.listenerCount('test');
   */
  listenerCount(event) {
    return this.events.has(event) ? this.events.get(event).length : 0;
  }

  /**
   * Gets all event names that have registered listeners
   *
   * @returns {string[]} - Array of event names
   *
   * @example
   * // Get all active events
   * const events = eventEmitter.eventNames();
   */
  eventNames() {
    return Array.from(this.events.keys());
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
    if (RPCClient.isDevelopment()) {
      console.group("RPCClient.constructor: Initializing");
    }

    try {
      super();
      this.safeLog("debug", "RPCClient.constructor: Setting up base properties");

      this.baseUrl = "./rpc";
      this.safeLog("info", "RPCClient.constructor: Base URL set to", { baseUrl: this.baseUrl });

      this.ws = null;
      this.sessionId = null;
      this.sessionExpiry = null;
      this.safeLog("info", "RPCClient.constructor: WebSocket and session initialized to null");

      this.requestQueue = new Map();
      this.requestId = 1;
      this.safeLog("info", "RPCClient.constructor: Request tracking initialized");

      this.reconnectAttempts = 0;
      this.maxReconnectAttempts = 5;
      this.safeLog("info", "RPCClient.constructor: Reconnect settings configured", {
        maxAttempts: this.maxReconnectAttempts,
      });

      this.sessionExpiry = null; // Initialize session expiry
    } catch (error) {
      this.safeLog("error", "RPCClient.constructor: Failed to initialize:", error);
      throw error;
    } finally {
      if (RPCClient.isDevelopment()) {
        console.groupEnd();
      }
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
    console.group("RPCClient.connect: Establishing WebSocket connection");

    try {
      // Validate origin before attempting connection (CORS protection)
      this.validateOrigin();
      
      // Use secure WebSocket protocol for HTTPS origins
      const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${location.host}/rpc/ws`;
      
      this.safeLog("debug", "RPCClient.connect: Using WebSocket URL", { wsUrl });

      this.ws = new WebSocket(wsUrl);
      this.safeLog("info", "RPCClient.connect: WebSocket instance created");

      this.setupWebSocket();
      this.safeLog("info", "RPCClient.connect: WebSocket handlers configured");

      await this.waitForConnection();
      this.safeLog("info", "RPCClient.connect: Connection established");

      this.reconnectAttempts = 0;
      this.safeLog("info", "RPCClient.connect: Reset reconnect attempts to 0");

      this.emit("connected");
      this.safeLog("info", "RPCClient.connect: Connected event emitted");
    } catch (error) {
      this.safeLog("error", "RPCClient.connect: Connection failed:", error);
      this.handleConnectionError(error);
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Waits for WebSocket connection to be established
   *
   * @returns {Promise<void>} Resolves when connection is ready, rejects on error
   * @throws {Error} If connection fails or times out
   * @private
   */
  waitForConnection() {
    return new Promise((resolve, reject) => {
      if (this.ws.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      this.ws.onopen = () => resolve();
      this.ws.onerror = () => reject(new Error("WebSocket connection failed"));
    });
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
    if (this.isDevelopment()) {
      console.group("RPCClient.request: Processing RPC request");
    }

    try {
      // Validate session before making request (except for joinGame)
      if (method !== 'joinGame' && this.sessionId) {
        this.validateSessionForRequest();
      }

      const id = this.requestId++;
      this.safeLog("debug", "RPCClient.request: Request parameters", {
        method,
        params: this.sanitizeForLogging(params),
        timeout,
        id,
      });

      const message = {
        jsonrpc: "2.0",
        method,
        params: { ...params, session_id: this.sessionId },
        id,
      };
      this.safeLog("info", "RPCClient.request: Formed JSON-RPC message", this.sanitizeForLogging(message));

      return new Promise((resolve, reject) => {
        const timeoutId = setTimeout(() => {
          this.safeLog("warn", "RPCClient.request: Request timed out", { method, id });
          this.requestQueue.delete(id);
          reject(new Error(`Request timeout: ${method}`));
        }, timeout);

        this.safeLog("info", "RPCClient.request: Adding to request queue", { id });
        this.requestQueue.set(id, {
          originalId: id,  // Store original ID for validation
          method: method,  // Store method for debugging
          timestamp: Date.now(),  // Store timestamp for monitoring
          resolve: (result) => {
            this.safeLog("info", "RPCClient.request: Request resolved", { 
              id, 
              result: this.sanitizeForLogging(result) 
            });
            clearTimeout(timeoutId);
            resolve(result);
          },
          reject: (error) => {
            this.safeLog("error", "RPCClient.request: Request rejected", { id, error });
            clearTimeout(timeoutId);
            reject(error);
          },
        });

        try {
          this.safeLog("debug", "RPCClient.request: Sending WebSocket message");
          this.ws.send(JSON.stringify(message));
        } catch (error) {
          this.safeLog("error", "RPCClient.request: Failed to send message", error);
          clearTimeout(timeoutId);
          this.requestQueue.delete(id);
          reject(error);
        }
      });
    } finally {
      if (this.isDevelopment()) {
        console.groupEnd();
      }
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
    console.group("RPCClient.setupWebSocket: Setting up WebSocket handlers");

    try {
      this.safeLog("debug", "RPCClient.setupWebSocket: Binding message handler");
      this.ws.onmessage = this.handleMessage.bind(this);

      this.safeLog("debug", "RPCClient.setupWebSocket: Binding close handler");
      this.ws.onclose = this.handleClose.bind(this);

      this.safeLog("debug", "RPCClient.setupWebSocket: Binding error handler");
      this.ws.onerror = this.handleError.bind(this);

      this.safeLog("info", "RPCClient.setupWebSocket: All handlers bound successfully");
    } catch (error) {
      this.safeLog("error", "RPCClient.setupWebSocket: Failed to setup handlers:", error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  /**
   * Processes incoming WebSocket messages and handles JSON-RPC responses
   *
   * @param {MessageEvent} event - The WebSocket message event containing response data
   * @throws {Error} If message parsing fails or response format is invalid
   * @fires error - If response contains an error
   * @see requestQueue - For tracking and resolving pending requests
   */
  handleMessage(event) {
    if (this.isDevelopment()) {
      console.group("RPCClient.handleMessage: Processing WebSocket message");
    }

    try {
      this.safeLog("debug", "RPCClient.handleMessage: Parsing message data");
      
      // Parse and validate JSON-RPC response format
      let response;
      try {
        response = JSON.parse(event.data);
        if (!this.validateJSONRPCResponse(response)) {
          throw new Error('Invalid JSON-RPC response format');
        }
      } catch (parseError) {
        this.safeLog("error", "RPCClient.handleMessage: Invalid response format", {
          error: parseError.message,
          rawData: event.data.substring(0, 200) // Log first 200 chars for debugging
        });
        this.emit('error', { 
          type: 'VALIDATION_ERROR', 
          message: parseError.message,
          rawData: event.data.substring(0, 200)
        });
        return;
      }

      this.safeLog("info", "RPCClient.handleMessage: Parsed response", this.sanitizeForLogging(response));

      // Check if we have a pending request for this response ID
      if (!response.id || !this.requestQueue.has(response.id)) {
        this.safeLog("warn", "RPCClient.handleMessage: No matching request found", {
          id: response.id,
        });
        this.emit('error', { 
          type: 'NO_MATCHING_REQUEST', 
          responseId: response.id,
          message: 'Received response for unknown request ID'
        });
        return;
      }

      // Enhanced ID validation: verify response ID matches original request
      const pendingRequest = this.requestQueue.get(response.id);
      if (!pendingRequest || pendingRequest.originalId !== response.id) {
        this.safeLog("error", "RPCClient.handleMessage: Response ID mismatch detected", {
          responseId: response.id,
          expectedId: pendingRequest ? pendingRequest.originalId : 'unknown',
          method: pendingRequest ? pendingRequest.method : 'unknown'
        });
        this.emit('error', { 
          type: 'ID_MISMATCH', 
          responseId: response.id,
          expectedId: pendingRequest ? pendingRequest.originalId : null,
          message: 'Response ID does not match original request ID - possible spoofing attack'
        });
        return;
      }

      const { resolve, reject } = pendingRequest;
      this.requestQueue.delete(response.id);

      if (response.error) {
        this.safeLog("error", "RPCClient.handleMessage: Error in response", response.error);
        reject(response.error);
        this.emit("error", response.error);
      } else {
        this.safeLog("info", "RPCClient.handleMessage: Success response", {
          result: this.sanitizeForLogging(response.result),
        });
        resolve(response.result);
      }
    } catch (error) {
      this.safeLog("error", "RPCClient.handleMessage: Failed to process message", error);
      this.emit("error", error);
    } finally {
      if (this.isDevelopment()) {
        console.groupEnd();
      }
    }
  }

  /**
   * Handles WebSocket connection closure events
   * Attempts to reconnect if maximum attempts not exceeded
   *
   * @param {CloseEvent} event - The WebSocket close event
   * @fires disconnected When the connection is closed
   * @see reconnect For the reconnection attempt logic
   */
  handleClose(event) {
    if (this.isDevelopment()) {
      console.group("RPCClient.handleClose: Processing WebSocket close");
    }
    try {
      this.safeLog("info", "RPCClient.handleClose: Connection closed", {
        code: event.code,
        reason: event.reason,
      });

      this.emit("disconnected");

      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        
        // Calculate exponential backoff delay with jitter
        const delay = this.calculateReconnectionDelay(this.reconnectAttempts);
        
        this.safeLog("info", "RPCClient.handleClose: Attempting reconnection", {
          attempt: this.reconnectAttempts,
          maxAttempts: this.maxReconnectAttempts,
          delayMs: delay
        });
        
        setTimeout(() => this.connect(), delay);
      } else {
        this.safeLog("error", "RPCClient.handleClose: Max reconnection attempts exceeded");
      }
    } catch (error) {
      this.safeLog("error", "RPCClient.handleClose: Error handling close", error);
    } finally {
      if (this.isDevelopment()) {
        console.groupEnd();
      }
    }
  }

  /**
   * Handles WebSocket error events by emitting them
   *
   * @param {Event} error - The WebSocket error event
   * @fires error When a WebSocket error occurs
   * @see handleConnectionError For connection-specific error handling
   */
  handleError(error) {
    if (this.isDevelopment()) {
      console.group("RPCClient.handleError: Processing WebSocket error");
    }
    try {
      this.safeLog("error", "RPCClient.handleError: WebSocket error occurred", error);
      this.emit("error", error);
    } catch (e) {
      this.safeLog("error", "RPCClient.handleError: Error handling WebSocket error", e);
    } finally {
      if (this.isDevelopment()) {
        console.groupEnd();
      }
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
    console.group("RPCClient.move: Processing move request");
    try {
      this.safeLog("debug", "RPCClient.move: Direction parameter", { direction });
      const result = await this.request("move", { direction });
      this.safeLog("info", "RPCClient.move: Move request completed", { result });
      return result;
    } catch (error) {
      this.safeLog("error", "RPCClient.move: Failed to process move request", error);
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
    console.group("RPCClient.attack: Processing attack request");
    try {
      this.safeLog("debug", "RPCClient.attack: Attack parameters", {
        targetId,
        weaponId,
      });
      const result = await this.request("attack", {
        target_id: targetId,
        weapon_id: weaponId,
      });
      this.safeLog("info", "RPCClient.attack: Attack request completed", { result: this.sanitizeForLogging(result) });
      return result;
    } catch (error) {
      this.safeLog("error", "RPCClient.attack: Failed to process attack request", error);
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
    console.group("RPCClient.castSpell: Processing spell cast request");
    try {
      this.safeLog("debug", "RPCClient.castSpell: Spell parameters", {
        spellId,
        targetId,
        position,
      });
      const result = await this.request("castSpell", {
        spell_id: spellId,
        target_id: targetId,
        position,
      });
      this.safeLog("info", "RPCClient.castSpell: Spell cast completed", { result });
      return result;
    } catch (error) {
      this.safeLog("error", "RPCClient.castSpell: Failed to cast spell", error);
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
    console.group("RPCClient.startCombat: Processing combat start request");
    try {
      this.safeLog("debug", "RPCClient.startCombat: Combat parameters", {
        participantIds,
      });
      const result = await this.request("startCombat", {
        participant_ids: participantIds,
      });
      this.safeLog("info", "RPCClient.startCombat: Combat started successfully", {
        result,
      });
      return result;
    } catch (error) {
      this.safeLog("error", "RPCClient.startCombat: Failed to start combat", error);
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
    console.group("RPCClient.endTurn: Processing end turn request");
    try {
      this.safeLog("debug", "RPCClient.endTurn: Sending end turn request");
      const result = await this.request("endTurn");
      this.safeLog("info", "RPCClient.endTurn: Turn ended successfully", { result });
      return result;
    } catch (error) {
      this.safeLog("error", "RPCClient.endTurn: Failed to end turn", error);
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
    console.group("RPCClient.getGameState: Fetching game state");
    try {
      this.safeLog("debug", "RPCClient.getGameState: Making request");
      const result = await this.request("getGameState");
      this.safeLog("info", "RPCClient.getGameState: State retrieved successfully", {
        result,
      });
      return result;
    } catch (error) {
      this.safeLog("error", "RPCClient.getGameState: Failed to retrieve game state", error);
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
    if (this.isDevelopment()) {
      console.group("RPCClient.joinGame: Processing join game request");
    }
    try {
      this.safeLog("debug", "RPCClient.joinGame: Player name parameter", {
        playerName,
      });
      const result = await this.request("joinGame", {
        player_name: playerName,
      });
      
      // Use secure session management
      this.setSession(result);
      return result;
    } catch (error) {
      this.safeLog("error", "RPCClient.joinGame: Failed to join game", error);
      throw error;
    } finally {
      if (this.isDevelopment()) {
        console.groupEnd();
      }
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
    if (this.isDevelopment()) {
      console.group("RPCClient.leaveGame: Processing leave game request");
    }
    try {
      if (this.sessionId) {
        this.safeLog("debug", "RPCClient.leaveGame: Leaving current session");
        await this.request("leaveGame");
        this.safeLog("info", "RPCClient.leaveGame: Successfully left game");
        this.clearSession();
      } else {
        this.safeLog("warn", "RPCClient.leaveGame: No active session to leave");
      }
    } catch (error) {
      this.safeLog("error", "RPCClient.leaveGame: Failed to leave game", error);
      throw error;
    } finally {
      if (this.isDevelopment()) {
        console.groupEnd();
      }
    }
  }

  /**
   * Validates a session token structure and format
   * @param {string} token - The session token to validate
   * @returns {boolean} True if the token format is valid
   * @private
   */
  validateSessionTokenFormat(token) {
    if (typeof token !== 'string' || token.length === 0) {
      return false;
    }
    
    // Basic format validation - should be a UUID-like string
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    return uuidRegex.test(token);
  }

  /**
   * Validates session data structure and token
   * @param {Object} sessionData - Session data from server
   * @returns {boolean} True if session data is valid
   * @private
   */
  validateSessionData(sessionData) {
    if (!sessionData || typeof sessionData !== 'object') {
      return false;
    }
    
    // Must have session_id
    if (!sessionData.session_id || !this.validateSessionTokenFormat(sessionData.session_id)) {
      return false;
    }
    
    return true;
  }

  /**
   * Checks if the current session has expired
   * @returns {boolean} True if session is expired or invalid
   * @private
   */
  isSessionExpired() {
    if (!this.sessionId || !this.sessionExpiry) {
      return true;
    }
    
    return new Date() >= this.sessionExpiry;
  }

  /**
   * Validates session before making requests
   * @throws {Error} If session is invalid or expired
   * @private
   */
  validateSessionForRequest() {
    if (!this.sessionId) {
      throw new Error('No active session - please join a game first');
    }
    
    if (!this.validateSessionTokenFormat(this.sessionId)) {
      throw new Error('Invalid session token format');
    }
    
    if (this.isSessionExpired()) {
      this.clearSession();
      throw new Error('Session has expired - please join the game again');
    }
  }

  /**
   * Sets session data with validation and expiration tracking
   * @param {Object} sessionData - Session data from server
   * @param {string} sessionData.session_id - The session token
   * @param {number} [expiryMinutes=30] - Session expiry time in minutes
   * @throws {Error} If session data is invalid
   * @private
   */
  setSession(sessionData, expiryMinutes = 30) {
    if (!this.validateSessionData(sessionData)) {
      throw new Error('Invalid session data received from server');
    }
    
    this.sessionId = sessionData.session_id;
    
    // Set expiration time (default 30 minutes from now)
    this.sessionExpiry = new Date();
    this.sessionExpiry.setMinutes(this.sessionExpiry.getMinutes() + expiryMinutes);
    
    this.safeLog("info", "Session established", {
      hasSessionId: !!this.sessionId,
      expiresAt: this.sessionExpiry.toISOString()
    });
  }

  /**
   * Clears session data and expiration
   * @private
   */
  clearSession() {
    this.sessionId = null;
    this.sessionExpiry = null;
    this.safeLog("info", "Session cleared");
  }

  /**
   * Determines if we're in development mode for logging purposes
   * @returns {boolean} True if in development mode, false for production
   * @static
   */
  static isDevelopment() {
    // Check multiple indicators for development environment
    return (
      window.location.hostname === 'localhost' ||
      window.location.hostname === '127.0.0.1' ||
      window.location.hostname.includes('dev') ||
      window.location.port === '8080' || // Common dev port
      (typeof process !== 'undefined' && process.env?.NODE_ENV === 'development')
    );
  }

  /**
   * Determines if we're in development mode for logging purposes (instance method)
   * @returns {boolean} True if in development mode, false for production
   * @private
   */
  isDevelopment() {
    return RPCClient.isDevelopment();
  }

  /**
   * Sanitizes sensitive data from objects for safe logging
   * @param {*} data - Data to sanitize
   * @returns {*} Sanitized data with sensitive fields redacted
   * @private
   */
  sanitizeForLogging(data) {
    if (!this.isDevelopment()) {
      if (typeof data === 'object' && data !== null) {
        const sanitized = { ...data };
        
        // Redact sensitive fields
        if (sanitized.session_id) {
          sanitized.session_id = '[REDACTED]';
        }
        if (sanitized.sessionId) {
          sanitized.sessionId = '[REDACTED]';
        }
        if (sanitized.params && sanitized.params.session_id) {
          sanitized.params = { ...sanitized.params, session_id: '[REDACTED]' };
        }
        
        // Redact result data that might contain sensitive info
        if (sanitized.result && typeof sanitized.result === 'object') {
          const result = { ...sanitized.result };
          if (result.session_id) result.session_id = '[REDACTED]';
          if (result.player_data) result.player_data = '[REDACTED]';
          sanitized.result = result;
        }
        
        return sanitized;
      }
    }
    return data;
  }

  /**
   * Safe console logging that redacts sensitive data in production
   * @param {string} level - Log level (debug, info, warn, error)
   * @param {string} message - Log message
   * @param {*} data - Data to log (will be sanitized)
   * @private
   */
  safeLog(level, message, data = null) {
    if (!this.isDevelopment() && (level === 'debug' || level === 'info')) {
      // Suppress debug/info logs in production
      return;
    }
    
    const sanitizedData = data ? this.sanitizeForLogging(data) : null;
    
    if (sanitizedData) {
      console[level](message, sanitizedData);
    } else {
      console[level](message);
    }
  }

  /**
   * Calculates reconnection delay using exponential backoff with jitter
   * @param {number} attempt - Current reconnection attempt number
   * @returns {number} Delay in milliseconds before next reconnection attempt
   * @private
   */
  calculateReconnectionDelay(attempt) {
    const baseDelay = 1000; // 1 second base delay
    const maxDelay = 30000; // 30 seconds maximum delay
    
    // Exponential backoff: delay = baseDelay * 2^attempt
    const exponentialDelay = baseDelay * Math.pow(2, attempt - 1);
    
    // Cap at maximum delay
    const cappedDelay = Math.min(exponentialDelay, maxDelay);
    
    // Add jitter: ±10% random variation to prevent thundering herd
    const jitterRange = 0.1 * cappedDelay;
    const jitter = (Math.random() - 0.5) * 2 * jitterRange;
    
    return Math.round(cappedDelay + jitter);
  }

  /**
   * Properly cleans up the RPCClient instance to prevent memory leaks
   *
   * @description Closes WebSocket connection, clears all event listeners,
   * cancels pending reconnection attempts, and clears session data
   *
   * @example
   * // Clean up when done with the client
   * rpcClient.cleanup();
   *
   * @notes
   * - Should be called when the RPCClient is no longer needed
   * - Prevents memory leaks from accumulated event listeners
   * - Stops any pending reconnection attempts
   * - After calling cleanup, the client should not be used again
   */
  cleanup() {
    try {
      this.safeLog("info", "RPCClient.cleanup: Starting cleanup process");

      // Clear reconnection timeout if active
      if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout);
        this.reconnectTimeout = null;
        this.safeLog("debug", "RPCClient.cleanup: Cleared reconnection timeout");
      }

      // Close WebSocket connection if open
      if (this.ws && this.ws.readyState !== WebSocket.CLOSED) {
        this.safeLog("debug", "RPCClient.cleanup: Closing WebSocket connection");
        this.ws.close(1000, "Client cleanup");
        this.ws = null;
      }

      // Clear all event listeners to prevent memory leaks
      this.safeLog("debug", "RPCClient.cleanup: Clearing all event listeners");
      this.clear();

      // Clear session data
      this.clearSession();

      // Reset reconnection state
      this.reconnectAttempts = 0;

      this.safeLog("info", "RPCClient.cleanup: Cleanup completed successfully");
    } catch (error) {
      this.safeLog("error", "RPCClient.cleanup: Error during cleanup", error);
    }
  }

  /**
   * Validates the current origin against a configurable allowlist of authorized origins
   *
   * @returns {boolean} True if the origin is authorized, false otherwise
   * @throws {Error} If the current origin is not in the allowlist
   *
   * @description Provides CORS protection by validating the current hostname
   * against a predefined list of authorized origins. This prevents unauthorized
   * access from malicious sites hosting the client code.
   *
   * @example
   * // Will validate current hostname against allowlist
   * rpcClient.validateOrigin(); // May throw if unauthorized
   *
   * @notes
   * - In development mode, localhost and common dev hostnames are automatically allowed
   * - Configure AUTHORIZED_ORIGINS for production environments
   * - This prevents cross-site request forgery via unauthorized hosting
   */
  validateOrigin() {
    const currentOrigin = location.hostname.toLowerCase();
    
    // Development mode: allow common development origins
    if (this.isDevelopment()) {
      const devOrigins = [
        'localhost',
        '127.0.0.1',
        '0.0.0.0',
        'vscode-local', // VS Code development
        'goldbox-rpg' // Codespace hostnames
      ];
      
      // Check for exact matches or proper subdomain patterns
      const isDevOrigin = devOrigins.some(devOrigin => {
        // Exact match
        if (currentOrigin === devOrigin) return true;
        
        // Subdomain match (but not suffix match)
        if (currentOrigin.endsWith('.' + devOrigin)) return true;
        
        return false;
      });
      
      // Check for cloud development platforms
      const isCloudDev = currentOrigin.includes('github.dev') ||
                        currentOrigin.includes('gitpod.io') ||
                        currentOrigin.includes('preview.app');
      
      if (isDevOrigin || isCloudDev) {
        this.safeLog("debug", "RPCClient.validateOrigin: Development origin allowed", { 
          origin: currentOrigin 
        });
        return true;
      }
    }

    // Production mode: strict allowlist validation
    // Configure this array for your production deployment
    const authorizedOrigins = [
      // Add your production domains here
      'your-game-domain.com',
      'app.your-game-domain.com',
      'game.your-domain.com'
      // Note: This is intentionally restrictive for security
    ];

    const isAuthorized = authorizedOrigins.includes(currentOrigin);
    
    if (!isAuthorized) {
      const errorMsg = `Unauthorized origin: ${currentOrigin}. This client is not authorized to connect from this domain.`;
      this.safeLog("error", "RPCClient.validateOrigin: Unauthorized origin detected", {
        currentOrigin,
        authorizedOrigins: this.isDevelopment() ? authorizedOrigins : '[REDACTED]'
      });
      throw new Error(errorMsg);
    }

    this.safeLog("info", "RPCClient.validateOrigin: Origin authorized", { 
      origin: currentOrigin 
    });
    return true;
  }
}
