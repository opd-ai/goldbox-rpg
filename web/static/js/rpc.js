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
    if (RPCClient.isDevelopment()) {
      console.group("RPCClient.constructor: Initializing");
    }

    try {
      super();
      if (RPCClient.isDevelopment()) {
        console.debug("RPCClient.constructor: Setting up base properties");
      }

      this.baseUrl = "./rpc";
      if (RPCClient.isDevelopment()) {
        console.info("RPCClient.constructor: Base URL set to", this.baseUrl);
      }

      this.ws = null;
      this.sessionId = null;
      if (RPCClient.isDevelopment()) {
        console.info(
          "RPCClient.constructor: WebSocket and session initialized to null",
        );
      }

      this.requestQueue = new Map();
      this.requestId = 1;
      if (RPCClient.isDevelopment()) {
        console.info("RPCClient.constructor: Request tracking initialized");
      }

      this.reconnectAttempts = 0;
      this.maxReconnectAttempts = 5;
      if (RPCClient.isDevelopment()) {
        console.info("RPCClient.constructor: Reconnect settings configured", {
          maxAttempts: this.maxReconnectAttempts,
        });
      }
    } catch (error) {
      console.error("RPCClient.constructor: Failed to initialize:", error);
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
      // Use secure WebSocket protocol for HTTPS origins
      const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${location.host}/rpc/ws`;
      
      console.debug(
        "RPCClient.connect: Using WebSocket URL",
        wsUrl,
      );

      this.ws = new WebSocket(wsUrl);
      console.info("RPCClient.connect: WebSocket instance created");

      this.setupWebSocket();
      console.info("RPCClient.connect: WebSocket handlers configured");

      await this.waitForConnection();
      console.info("RPCClient.connect: Connection established");

      this.reconnectAttempts = 0;
      console.info("RPCClient.connect: Reset reconnect attempts to 0");

      this.emit("connected");
      console.info("RPCClient.connect: Connected event emitted");
    } catch (error) {
      console.error("RPCClient.connect: Connection failed:", error);
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
      console.debug("RPCClient.setupWebSocket: Binding message handler");
      this.ws.onmessage = this.handleMessage.bind(this);

      console.debug("RPCClient.setupWebSocket: Binding close handler");
      this.ws.onclose = this.handleClose.bind(this);

      console.debug("RPCClient.setupWebSocket: Binding error handler");
      this.ws.onerror = this.handleError.bind(this);

      console.info("RPCClient.setupWebSocket: All handlers bound successfully");
    } catch (error) {
      console.error(
        "RPCClient.setupWebSocket: Failed to setup handlers:",
        error,
      );
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
      const response = JSON.parse(event.data);
      this.safeLog("info", "RPCClient.handleMessage: Parsed response", this.sanitizeForLogging(response));

      if (!response.id || !this.requestQueue.has(response.id)) {
        this.safeLog("warn", "RPCClient.handleMessage: No matching request found", {
          id: response.id,
        });
        return;
      }

      const { resolve, reject } = this.requestQueue.get(response.id);
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
    console.group("RPCClient.handleClose: Processing WebSocket close");
    try {
      console.info("RPCClient.handleClose: Connection closed", {
        code: event.code,
        reason: event.reason,
      });

      this.emit("disconnected");

      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        console.info("RPCClient.handleClose: Attempting reconnection");
        this.reconnectAttempts++;
        setTimeout(() => this.connect(), 1000 * this.reconnectAttempts);
      } else {
        console.error(
          "RPCClient.handleClose: Max reconnection attempts exceeded",
        );
      }
    } catch (error) {
      console.error("RPCClient.handleClose: Error handling close", error);
    } finally {
      console.groupEnd();
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
    console.group("RPCClient.handleError: Processing WebSocket error");
    try {
      console.error("RPCClient.handleError: WebSocket error occurred", error);
      this.emit("error", error);
    } catch (e) {
      console.error("RPCClient.handleError: Error handling WebSocket error", e);
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
    console.group("RPCClient.move: Processing move request");
    try {
      console.debug("RPCClient.move: Direction parameter", { direction });
      const result = await this.request("move", { direction });
      console.info("RPCClient.move: Move request completed", { result });
      return result;
    } catch (error) {
      console.error("RPCClient.move: Failed to process move request", error);
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
      console.debug("RPCClient.attack: Attack parameters", {
        targetId,
        weaponId,
      });
      const result = await this.request("attack", {
        target_id: targetId,
        weapon_id: weaponId,
      });
      console.info("RPCClient.attack: Attack request completed", { result });
      return result;
    } catch (error) {
      console.error(
        "RPCClient.attack: Failed to process attack request",
        error,
      );
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
      console.debug("RPCClient.castSpell: Spell parameters", {
        spellId,
        targetId,
        position,
      });
      const result = await this.request("castSpell", {
        spell_id: spellId,
        target_id: targetId,
        position,
      });
      console.info("RPCClient.castSpell: Spell cast completed", { result });
      return result;
    } catch (error) {
      console.error("RPCClient.castSpell: Failed to cast spell", error);
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
      console.debug("RPCClient.startCombat: Combat parameters", {
        participantIds,
      });
      const result = await this.request("startCombat", {
        participant_ids: participantIds,
      });
      console.info("RPCClient.startCombat: Combat started successfully", {
        result,
      });
      return result;
    } catch (error) {
      console.error("RPCClient.startCombat: Failed to start combat", error);
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
      console.debug("RPCClient.endTurn: Sending end turn request");
      const result = await this.request("endTurn");
      console.info("RPCClient.endTurn: Turn ended successfully", { result });
      return result;
    } catch (error) {
      console.error("RPCClient.endTurn: Failed to end turn", error);
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
      console.debug("RPCClient.getGameState: Making request");
      const result = await this.request("getGameState");
      console.info("RPCClient.getGameState: State retrieved successfully", {
        result,
      });
      return result;
    } catch (error) {
      console.error(
        "RPCClient.getGameState: Failed to retrieve game state",
        error,
      );
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
      this.safeLog("info", "RPCClient.joinGame: Session established", {
        hasSessionId: !!result.session_id,
      });
      this.sessionId = result.session_id;
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
    console.group("RPCClient.leaveGame: Processing leave game request");
    try {
      if (this.sessionId) {
        console.debug("RPCClient.leaveGame: Current session ID", {
          sessionId: this.sessionId,
        });
        await this.request("leaveGame");
        console.info("RPCClient.leaveGame: Successfully left game");
        this.sessionId = null;
        console.info("RPCClient.leaveGame: Session ID cleared");
      } else {
        console.warn("RPCClient.leaveGame: No active session to leave");
      }
    } catch (error) {
      console.error("RPCClient.leaveGame: Failed to leave game", error);
      throw error;
    } finally {
      console.groupEnd();
    }
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
}
