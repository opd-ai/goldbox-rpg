"use strict";
var GoldBoxRPG = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

  // src/main.ts
  var main_exports = {};
  __export(main_exports, {
    GoldBoxRPG: () => app
  });

  // src/core/EventEmitter.ts
  var EventEmitter = class {
    constructor() {
      this.events = /* @__PURE__ */ new Map();
    }
    /**
     * Register an event listener
     * @param event - Event name to listen for
     * @param callback - Function to call when event is emitted
     * @returns Unsubscriber function to remove the listener
     */
    on(event, callback) {
      if (!this.events.has(event)) {
        this.events.set(event, /* @__PURE__ */ new Set());
      }
      const listeners = this.events.get(event);
      listeners.add(callback);
      return () => {
        listeners.delete(callback);
        if (listeners.size === 0) {
          this.events.delete(event);
        }
      };
    }
    /**
     * Emit an event to all registered listeners
     * @param event - Event name to emit
     * @param data - Data to pass to event listeners
     */
    emit(event, data) {
      const listeners = this.events.get(event);
      if (!listeners) {
        return;
      }
      const listenersCopy = Array.from(listeners);
      for (const callback of listenersCopy) {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in event listener for '${event}':`, error);
        }
      }
    }
    /**
     * Remove a specific event listener
     * @param event - Event name
     * @param callback - Specific callback to remove
     * @returns True if listener was found and removed, false otherwise
     */
    off(event, callback) {
      const listeners = this.events.get(event);
      if (!listeners) {
        return false;
      }
      const removed = listeners.delete(callback);
      if (listeners.size === 0) {
        this.events.delete(event);
      }
      return removed;
    }
    /**
     * Remove all listeners for a specific event, or all events if no event specified
     * @param event - Optional specific event to clear, or undefined to clear all
     */
    removeAllListeners(event) {
      if (event !== void 0) {
        this.events.delete(event);
      } else {
        this.events.clear();
      }
    }
    /**
     * Clear all events and listeners (memory leak prevention)
     */
    clear() {
      this.events.clear();
    }
    /**
     * Get the number of listeners for a specific event
     * @param event - Event name to count listeners for
     * @returns Number of registered listeners
     */
    listenerCount(event) {
      const listeners = this.events.get(event);
      return listeners ? listeners.size : 0;
    }
    /**
     * Get an array of all event names that have listeners
     * @returns Array of event names
     */
    eventNames() {
      return Array.from(this.events.keys());
    }
    /**
     * Check if there are any listeners for a specific event
     * @param event - Event name to check
     * @returns True if there are listeners, false otherwise
     */
    hasListeners(event) {
      return this.listenerCount(event) > 0;
    }
    /**
     * Get debugging information about current event listeners
     * @returns Object with event listener statistics
     */
    getDebugInfo() {
      const events = {};
      let totalListeners = 0;
      for (const [event, listeners] of Array.from(this.events.entries())) {
        events[event] = listeners.size;
        totalListeners += listeners.size;
      }
      return {
        totalEvents: this.events.size,
        totalListeners,
        events
      };
    }
  };
  var TypedEventEmitter = class {
    constructor() {
      this.emitter = new EventEmitter();
    }
    /**
     * Register a typed event listener
     */
    on(event, callback) {
      return this.emitter.on(event, callback);
    }
    /**
     * Emit a typed event
     */
    emit(event, data) {
      this.emitter.emit(event, data);
    }
    /**
     * Remove a typed event listener
     */
    off(event, callback) {
      return this.emitter.off(event, callback);
    }
    /**
     * Remove all listeners for an event
     */
    removeAllListeners(event) {
      this.emitter.removeAllListeners(event);
    }
    /**
     * Clear all events
     */
    clear() {
      this.emitter.clear();
    }
    /**
     * Get listener count for an event
     */
    listenerCount(event) {
      return this.emitter.listenerCount(event);
    }
    /**
     * Get all event names
     */
    eventNames() {
      return this.emitter.eventNames();
    }
    /**
     * Get debug information
     */
    getDebugInfo() {
      return this.emitter.getDebugInfo();
    }
  };

  // src/utils/Logger.ts
  var Logger = class _Logger {
    constructor(_component = "Logger") {
      this.logQueue = [];
      this.maxQueueSize = 100;
      this.config = this.createConfig();
      this.consoleDebug = console.debug.bind(console);
      this.consoleInfo = console.info.bind(console);
      this.consoleWarn = console.warn.bind(console);
      this.consoleError = console.error.bind(console);
      this.consoleGroup = console.group.bind(console);
      this.consoleGroupCollapsed = console.groupCollapsed.bind(console);
      this.consoleGroupEnd = console.groupEnd.bind(console);
    }
    /**
     * Creates logger configuration based on environment detection
     */
    createConfig() {
      const isDevelopment = this.detectDevelopmentEnvironment();
      return {
        isDevelopment,
        enabledLevels: this.getEnabledLevels(isDevelopment),
        maxQueueSize: this.maxQueueSize
      };
    }
    /**
     * Detects if running in development environment
     */
    detectDevelopmentEnvironment() {
      if (typeof window === "undefined") {
        return false;
      }
      const hostname = window.location.hostname;
      const developmentHosts = [
        "localhost",
        "127.0.0.1",
        "0.0.0.0"
      ];
      const isDevelopmentHost = developmentHosts.includes(hostname);
      const isLocalIP = /^192\.168\.|^10\.|^172\.(1[6-9]|2\d|3[01])\./.test(hostname);
      const isVSCodeLocal = hostname.includes("vscode-local");
      const isCodespaces = hostname.includes("githubpreview") || hostname.includes("app.github.dev");
      const isGitpod = hostname.includes("gitpod.io");
      const isPreviewApp = hostname.includes("preview.app") || hostname.includes("netlify.app");
      return isDevelopmentHost || isLocalIP || isVSCodeLocal || isCodespaces || isGitpod || isPreviewApp;
    }
    /**
     * Gets enabled log levels based on environment
     */
    getEnabledLevels(isDevelopment) {
      if (isDevelopment) {
        return /* @__PURE__ */ new Set(["debug", "info", "warn", "error", "group"]);
      } else {
        return /* @__PURE__ */ new Set(["warn", "error"]);
      }
    }
    /**
     * Adds entry to log queue for monitoring
     */
    addToQueue(entry) {
      this.logQueue.push(entry);
      if (this.logQueue.length > this.maxQueueSize) {
        this.logQueue.shift();
      }
    }
    /**
     * Logs a debug message (development only)
     */
    debug(message, ...args) {
      this.log("debug", message, ...args);
    }
    /**
     * Logs an info message (development only)
     */
    info(message, ...args) {
      this.log("info", message, ...args);
    }
    /**
     * Logs a warning message (always enabled)
     */
    warn(message, ...args) {
      this.log("warn", message, ...args);
    }
    /**
     * Logs an error message (always enabled)
     */
    error(message, ...args) {
      this.log("error", message, ...args);
    }
    /**
     * Starts a console group (development only)
     */
    group(label, ...args) {
      if (this.config.enabledLevels.has("group")) {
        this.consoleGroup(label, ...args);
        this.addToQueue({
          level: "info",
          message: `GROUP: ${label}`,
          timestamp: Date.now(),
          args: [label, ...args]
        });
      }
    }
    /**
     * Starts a collapsed console group (development only)
     */
    groupCollapsed(label, ...args) {
      if (this.config.enabledLevels.has("group")) {
        this.consoleGroupCollapsed(label, ...args);
        this.addToQueue({
          level: "info",
          message: `GROUP_COLLAPSED: ${label}`,
          timestamp: Date.now(),
          args: [label, ...args]
        });
      }
    }
    /**
     * Ends a console group (development only)
     */
    groupEnd() {
      if (this.config.enabledLevels.has("group")) {
        this.consoleGroupEnd();
        this.addToQueue({
          level: "info",
          message: "GROUP_END",
          timestamp: Date.now(),
          args: []
        });
      }
    }
    /**
     * Conditionally logs based on environment and level
     */
    log(level, message, ...args) {
      if (!this.config.enabledLevels.has(level)) {
        return;
      }
      const timestamp = Date.now();
      const entry = {
        level,
        message,
        timestamp,
        args
      };
      this.addToQueue(entry);
      switch (level) {
        case "debug":
          this.consoleDebug(message, ...args);
          break;
        case "info":
          this.consoleInfo(message, ...args);
          break;
        case "warn":
          this.consoleWarn(message, ...args);
          break;
        case "error":
          this.consoleError(message, ...args);
          break;
      }
    }
    /**
     * Enables debug logging temporarily (useful for production debugging)
     */
    enableDebug(duration = 6e4) {
      const originalLevels = this.config.enabledLevels;
      const newLevels = /* @__PURE__ */ new Set([...Array.from(originalLevels), "debug", "info", "group"]);
      this.config.enabledLevels = newLevels;
      this.info(`Debug logging enabled for ${duration}ms`);
      setTimeout(() => {
        this.config.enabledLevels = originalLevels;
        this.info("Debug logging disabled");
      }, duration);
    }
    /**
     * Gets current logging configuration
     */
    getConfig() {
      return this.config;
    }
    /**
     * Gets recent log entries from the queue
     */
    getRecentLogs(count = 50) {
      return this.logQueue.slice(-count).map(({ args, ...entry }) => entry);
    }
    /**
     * Clears the log queue
     */
    clearQueue() {
      this.logQueue.length = 0;
    }
    /**
     * Creates a child logger with a component prefix
     */
    createChildLogger(component) {
      const childLogger = new _Logger(component);
      const originalLog = childLogger.log.bind(childLogger);
      childLogger.log = (level, message, ...args) => {
        originalLog(level, `[${component}] ${message}`, ...args);
      };
      return childLogger;
    }
    /**
     * Formats a timestamp for logging
     */
    static formatTimestamp(timestamp) {
      return new Date(timestamp).toISOString();
    }
    /**
     * Safely stringifies an object for logging
     */
    static safeStringify(obj) {
      try {
        return JSON.stringify(obj, null, 2);
      } catch (error) {
        return `[Object: ${String(obj)}]`;
      }
    }
  };
  var logger = new Logger("Global");
  if (typeof window !== "undefined" && !logger.getConfig().isDevelopment) {
    console.debug = logger.debug.bind(logger);
    console.info = logger.info.bind(logger);
    console.group = logger.group.bind(logger);
    console.groupEnd = logger.groupEnd.bind(logger);
    console.groupCollapsed = logger.groupCollapsed.bind(logger);
  }
  if (typeof window !== "undefined") {
    window.logger = logger;
  }

  // src/utils/ErrorHandler.ts
  var ErrorHandler = class {
    constructor(options) {
      this.component = options.component;
      this.eventEmitter = options.eventEmitter;
      this.userMessageCallback = options.userMessageCallback;
      this.enableStackTrace = options.enableStackTrace ?? true;
      this.enableMetadataLogging = options.enableMetadataLogging ?? true;
      this.componentLogger = logger.createChildLogger(this.component);
    }
    /**
     * Handles recoverable errors that should not stop execution
     * Logs error, emits event, and optionally shows user message
     */
    handleRecoverableError(error, context, userMessage, metadata = {}) {
      const errorObj = this.normalizeError(error);
      const errorContext = this.createErrorContext(context, metadata);
      this.logError(errorObj, errorContext, "warn");
      if (this.eventEmitter) {
        this.eventEmitter.emit("error", {
          error: errorObj,
          context: errorContext,
          recoverable: true,
          userMessage
        });
      }
      if (userMessage && this.userMessageCallback) {
        this.userMessageCallback(userMessage, "warning");
      }
    }
    /**
     * Handles critical errors that should stop execution
     * Logs error and throws it to stop execution flow
     */
    handleCriticalError(error, context, metadata = {}) {
      const errorObj = this.normalizeError(error);
      const errorContext = this.createErrorContext(context, metadata);
      this.logError(errorObj, errorContext, "error");
      if (this.eventEmitter) {
        this.eventEmitter.emit("error", {
          error: errorObj,
          context: errorContext,
          recoverable: false
        });
      }
      if (this.userMessageCallback) {
        this.userMessageCallback(
          `Critical error in ${this.component}: ${errorObj.message}`,
          "error"
        );
      }
      throw errorObj;
    }
    /**
     * Handles initialization errors with cleanup
     * Logs error, attempts cleanup, and throws
     */
    handleInitializationError(error, context, cleanupFn, metadata = {}) {
      const errorObj = this.normalizeError(error);
      const errorContext = this.createErrorContext(context, metadata);
      this.componentLogger.error(
        `Initialization failed in ${context}:`,
        errorObj,
        errorContext
      );
      if (cleanupFn) {
        try {
          cleanupFn();
          this.componentLogger.info("Cleanup completed after initialization failure");
        } catch (cleanupError) {
          this.componentLogger.error("Cleanup failed:", cleanupError);
        }
      }
      if (this.eventEmitter) {
        this.eventEmitter.emit("initializationError", {
          error: errorObj,
          context: errorContext
        });
      }
      throw errorObj;
    }
    /**
     * Wraps async operations with standardized error handling
     */
    wrapAsync(asyncFn, context, options = {}) {
      return async (...args) => {
        try {
          return await asyncFn(...args);
        } catch (error) {
          const normalizedError = this.normalizeError(error);
          if (options.onError) {
            try {
              options.onError(normalizedError);
            } catch (handlerError) {
              this.componentLogger.error("Error in custom error handler:", handlerError);
            }
          }
          if (options.critical) {
            this.handleCriticalError(normalizedError, context, options.metadata);
          } else {
            this.handleRecoverableError(
              normalizedError,
              context,
              options.userMessage,
              options.metadata
            );
            throw normalizedError;
          }
        }
      };
    }
    /**
     * Creates a safe wrapper for synchronous operations
     */
    wrapSync(syncFn, context, options = {}) {
      return (...args) => {
        try {
          return syncFn(...args);
        } catch (error) {
          const normalizedError = this.normalizeError(error);
          if (options.critical) {
            this.handleCriticalError(normalizedError, context, options.metadata);
          } else {
            this.handleRecoverableError(
              normalizedError,
              context,
              options.userMessage,
              options.metadata
            );
            if (options.defaultValue !== void 0) {
              return options.defaultValue;
            }
            throw normalizedError;
          }
        }
      };
    }
    /**
     * Normalizes different error types to Error objects
     */
    normalizeError(error) {
      if (error instanceof Error) {
        return error;
      }
      if (typeof error === "string") {
        return new Error(error);
      }
      if (error && typeof error === "object" && "message" in error) {
        return new Error(String(error.message));
      }
      return new Error(`Unknown error: ${String(error)}`);
    }
    /**
     * Creates error context with metadata
     */
    createErrorContext(method, metadata = {}) {
      const stackTrace = this.enableStackTrace ? new Error().stack : void 0;
      return {
        method,
        timestamp: Date.now(),
        metadata: this.enableMetadataLogging ? metadata : void 0,
        stackTrace
      };
    }
    /**
     * Logs error with appropriate level and formatting
     */
    logError(error, context, level = "error") {
      const logMessage = `${this.component}.${context.method}: ${error.message}`;
      const logData = {
        error: {
          name: error.name,
          message: error.message,
          stack: this.enableStackTrace ? error.stack : void 0
        },
        context
      };
      if (level === "error") {
        this.componentLogger.error(logMessage, logData);
      } else {
        this.componentLogger.warn(logMessage, logData);
      }
    }
    /**
     * Gets the component name this error handler is associated with
     */
    getComponent() {
      return this.component;
    }
    /**
     * Checks if error handling is configured for user messages
     */
    hasUserMessageHandler() {
      return this.userMessageCallback !== void 0;
    }
    /**
     * Checks if error handling is configured for event emission
     */
    hasEventEmitter() {
      return this.eventEmitter !== void 0;
    }
  };
  var GlobalErrorHandler = class {
    static {
      this.handlers = /* @__PURE__ */ new Map();
    }
    /**
     * Gets or creates an error handler for a component
     */
    static getHandler(component, options) {
      if (!this.handlers.has(component)) {
        this.handlers.set(component, new ErrorHandler({
          component,
          ...options
        }));
      }
      return this.handlers.get(component);
    }
    /**
     * Sets up global error handlers for unhandled errors
     */
    static setupGlobalHandlers() {
      if (typeof window !== "undefined") {
        window.addEventListener("unhandledrejection", (event) => {
          const handler = this.getHandler("GlobalPromiseRejection");
          handler.handleRecoverableError(
            event.reason,
            "unhandledPromiseRejection",
            "An unexpected error occurred",
            { promise: event.promise }
          );
        });
        window.addEventListener("error", (event) => {
          const handler = this.getHandler("GlobalError");
          handler.handleRecoverableError(
            event.error || new Error(event.message),
            "uncaughtError",
            "An unexpected error occurred",
            {
              filename: event.filename,
              lineno: event.lineno,
              colno: event.colno
            }
          );
        });
      }
    }
    /**
     * Clears all cached error handlers
     */
    static clearHandlers() {
      this.handlers.clear();
    }
  };

  // src/core/BaseComponent.ts
  var ComponentManager = class {
    constructor() {
      this.components = /* @__PURE__ */ new Map();
      this.logger = logger.createChildLogger("ComponentManager");
    }
    /**
     * Register a component
     */
    register(component) {
      if (this.components.has(component.name)) {
        throw new Error(`Component ${component.name} is already registered`);
      }
      this.components.set(component.name, component);
      this.logger.debug(`Registered component: ${component.name}`);
    }
    /**
     * Get a component by name
     */
    get(name) {
      return this.components.get(name);
    }
    /**
     * Initialize all registered components
     */
    async initializeAll() {
      this.logger.info("Initializing all components");
      const initPromises = Array.from(this.components.values()).map(
        (component) => component.initialize()
      );
      await Promise.all(initPromises);
      this.logger.info("All components initialized");
    }
    /**
     * Cleanup all registered components
     */
    async cleanupAll() {
      this.logger.info("Cleaning up all components");
      const cleanupPromises = Array.from(this.components.values()).map(
        (component) => component.cleanup()
      );
      await Promise.all(cleanupPromises);
      this.components.clear();
      this.logger.info("All components cleaned up");
    }
    /**
     * Update all components
     */
    updateAll(deltaTime) {
      for (const component of Array.from(this.components.values())) {
        component.update(deltaTime);
      }
    }
    /**
     * Get all registered component names
     */
    getComponentNames() {
      return Array.from(this.components.keys());
    }
    /**
     * Get component count
     */
    getComponentCount() {
      return this.components.size;
    }
  };

  // src/network/RPCClient.ts
  var RPCClient = class extends TypedEventEmitter {
    constructor(config = {}) {
      super();
      this.clientLogger = logger.createChildLogger("RPCClient");
      this.requestQueue = /* @__PURE__ */ new Map();
      this.ws = null;
      this.sessionId = null;
      this.sessionExpiry = null;
      this.requestId = 1;
      this.reconnectAttempts = 0;
      this.reconnectTimer = null;
      this.isConnecting = false;
      this.isDestroyed = false;
      this.config = {
        baseUrl: "./rpc",
        maxReconnectAttempts: 5,
        connectionTimeout: 1e4,
        requestTimeout: 3e4,
        reconnectDelay: 1e3,
        enableLogging: true,
        ...config
      };
      this.clientLogger.info("RPC Client initialized", { config: this.config });
    }
    /**
     * Connect to the RPC server via WebSocket
     */
    async connect() {
      if (this.isDestroyed) {
        throw new Error("Cannot connect destroyed RPC client");
      }
      if (this.isConnected() || this.isConnecting) {
        this.clientLogger.warn("Already connected or connecting");
        return;
      }
      this.isConnecting = true;
      this.clientLogger.info("Establishing WebSocket connection");
      try {
        await this.establishConnection();
        this.reconnectAttempts = 0;
        this.isConnecting = false;
        this.emit("connected", void 0);
        this.clientLogger.info("Successfully connected to RPC server");
      } catch (error) {
        this.isConnecting = false;
        this.clientLogger.error("Failed to connect:", error);
        await this.handleConnectionError(error);
        throw error;
      }
    }
    /**
     * Disconnect from the RPC server
     */
    disconnect(reason = "Client disconnect") {
      this.clientLogger.info("Disconnecting RPC client", { reason });
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
      if (this.ws) {
        this.ws.close(1e3, reason);
        this.ws = null;
      }
      this.sessionId = null;
      this.sessionExpiry = null;
      this.isConnecting = false;
      this.rejectPendingRequests(new Error(`Connection closed: ${reason}`));
      this.emit("disconnected", { reason });
    }
    /**
     * Destroy the RPC client and clean up resources
     */
    destroy() {
      this.isDestroyed = true;
      this.disconnect("Client destroyed");
      this.removeAllListeners();
      this.clientLogger.info("RPC client destroyed");
    }
    /**
     * Check if currently connected to server
     */
    isConnected() {
      return this.ws?.readyState === WebSocket.OPEN;
    }
    /**
     * Get current session information
     */
    getSession() {
      if (!this.sessionId || !this.sessionExpiry) {
        return null;
      }
      return {
        sessionId: this.sessionId,
        expiresAt: new Date(this.sessionExpiry),
        isValid: Date.now() < this.sessionExpiry
      };
    }
    /**
     * Send an RPC request to the server
     */
    async call(method, params, timeout) {
      if (this.isDestroyed) {
        throw new Error("Cannot call method on destroyed RPC client");
      }
      if (!this.isConnected()) {
        throw new Error("Not connected to RPC server");
      }
      const id = this.requestId++;
      const baseParams = params || {};
      const requestParams = this.sessionId ? { ...baseParams, sessionId: this.sessionId } : baseParams;
      const request = {
        jsonrpc: "2.0",
        method,
        params: requestParams,
        id
      };
      const requestTimeout = timeout || this.config.requestTimeout;
      return new Promise((resolve, reject) => {
        const timeoutId = setTimeout(() => {
          this.requestQueue.delete(id);
          reject(new Error(`Request timeout after ${requestTimeout}ms`));
        }, requestTimeout);
        this.requestQueue.set(id, {
          resolve: (value) => {
            clearTimeout(timeoutId);
            resolve(value);
          },
          reject: (error) => {
            clearTimeout(timeoutId);
            reject(error);
          },
          timestamp: Date.now(),
          method,
          timeout: requestTimeout
        });
        try {
          this.ws.send(JSON.stringify(request));
          this.clientLogger.debug("Sent RPC request", { method, id });
        } catch (error) {
          this.requestQueue.delete(id);
          clearTimeout(timeoutId);
          reject(error);
        }
      });
    }
    /**
     * Establish WebSocket connection with proper error handling
     */
    async establishConnection() {
      const protocol = location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${protocol}//${location.host}/rpc/ws`;
      this.clientLogger.debug("Connecting to WebSocket", { wsUrl });
      this.ws = new WebSocket(wsUrl);
      this.setupWebSocketHandlers();
      return this.waitForConnection();
    }
    /**
     * Set up WebSocket event handlers
     */
    setupWebSocketHandlers() {
      if (!this.ws)
        return;
      this.ws.addEventListener("open", () => {
        this.clientLogger.info("WebSocket connection opened");
      });
      this.ws.addEventListener("message", (event) => {
        this.handleMessage(event.data);
      });
      this.ws.addEventListener("close", (event) => {
        this.clientLogger.info("WebSocket connection closed", {
          code: event.code,
          reason: event.reason
        });
        if (!this.isDestroyed && event.code !== 1e3) {
          this.handleConnectionError(new Error(`Connection closed unexpectedly: ${event.reason}`));
        }
      });
      this.ws.addEventListener("error", (event) => {
        this.clientLogger.error("WebSocket error:", event);
        this.emit("error", { error: new Error("WebSocket error") });
      });
    }
    /**
     * Wait for WebSocket connection to be established
     */
    waitForConnection() {
      return new Promise((resolve, reject) => {
        if (!this.ws) {
          reject(new Error("No WebSocket instance"));
          return;
        }
        if (this.ws.readyState === WebSocket.OPEN) {
          resolve();
          return;
        }
        const timeout = setTimeout(() => {
          reject(new Error(`Connection timeout after ${this.config.connectionTimeout}ms`));
        }, this.config.connectionTimeout);
        const openHandler = () => {
          clearTimeout(timeout);
          resolve();
        };
        const errorHandler = () => {
          clearTimeout(timeout);
          reject(new Error("WebSocket connection failed"));
        };
        this.ws.addEventListener("open", openHandler, { once: true });
        this.ws.addEventListener("error", errorHandler, { once: true });
      });
    }
    /**
     * Handle incoming WebSocket messages
     */
    handleMessage(data) {
      try {
        const message = JSON.parse(data);
        this.clientLogger.debug("Received RPC message", { id: message.id });
        if ("id" in message && message.id !== null) {
          this.handleResponse(message);
        } else {
          this.handleNotification(message);
        }
      } catch (error) {
        this.clientLogger.error("Failed to parse message:", error);
      }
    }
    /**
     * Handle RPC response messages
     */
    handleResponse(response) {
      const pendingRequest = this.requestQueue.get(response.id);
      if (!pendingRequest) {
        this.clientLogger.warn("Received response for unknown request", { id: response.id });
        return;
      }
      this.requestQueue.delete(response.id);
      if ("error" in response && response.error) {
        const error = new Error(response.error.message);
        error.code = response.error.code;
        error.data = response.error.data;
        pendingRequest.reject(error);
      } else {
        pendingRequest.resolve(response.result);
      }
    }
    /**
     * Handle server notifications
     */
    handleNotification(notification) {
      this.emit("message", { data: notification.result });
      if (typeof notification.result === "object" && notification.result) {
        const result = notification.result;
        if (result.type === "sessionExpired") {
          this.sessionId = null;
          this.sessionExpiry = null;
          this.emit("sessionExpired", void 0);
        }
      }
    }
    /**
     * Handle connection errors with exponential backoff
     */
    async handleConnectionError(error) {
      this.reconnectAttempts++;
      this.clientLogger.warn(`Connection error (attempt ${this.reconnectAttempts}):`, error);
      if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
        this.clientLogger.error("Max reconnection attempts reached");
        this.emit("error", { error: new Error("Max reconnection attempts exceeded") });
        return;
      }
      const delay = this.config.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      this.clientLogger.info(`Reconnecting in ${delay}ms...`);
      this.emit("reconnecting", { attempt: this.reconnectAttempts });
      this.reconnectTimer = window.setTimeout(async () => {
        try {
          await this.connect();
        } catch (reconnectError) {
          this.clientLogger.error("Reconnection failed:", reconnectError);
        }
      }, delay);
    }
    /**
     * Reject all pending requests
     */
    rejectPendingRequests(error) {
      for (const [, request] of this.requestQueue.entries()) {
        request.reject(error);
      }
      this.requestQueue.clear();
    }
  };
  var rpcClient = new RPCClient();

  // src/ui/GameUI.ts
  var GameUI = class extends TypedEventEmitter {
    constructor() {
      super();
      this.uiLogger = logger.createChildLogger("GameUI");
      this.elements = null;
      this.isInitialized = false;
      this.keyboardHandlers = /* @__PURE__ */ new Map();
      this.uiLogger.info("GameUI initialized");
    }
    /**
     * Initialize the UI by finding DOM elements and setting up event handlers
     */
    async initialize() {
      if (this.isInitialized) {
        this.uiLogger.warn("UI already initialized");
        return;
      }
      try {
        this.uiLogger.info("Initializing Game UI...");
        this.elements = this.findUIElements();
        this.validateElements();
        this.setupEventListeners();
        this.setupKeyboardControls();
        this.isInitialized = true;
        this.uiLogger.info("Game UI initialized successfully");
      } catch (error) {
        this.uiLogger.error("Failed to initialize UI:", error);
        throw error;
      }
    }
    /**
     * Clean up UI resources and event listeners
     */
    cleanup() {
      if (!this.isInitialized) {
        return;
      }
      this.uiLogger.info("Cleaning up Game UI...");
      for (const [, handler] of this.keyboardHandlers.entries()) {
        document.removeEventListener("keydown", handler);
      }
      this.keyboardHandlers.clear();
      this.removeAllListeners();
      this.elements = null;
      this.isInitialized = false;
      this.uiLogger.info("Game UI cleanup completed");
    }
    /**
     * Update the UI with current game state
     */
    updateUI(state) {
      if (!this.isInitialized || !this.elements) {
        this.uiLogger.warn("Cannot update UI - not initialized");
        return;
      }
      try {
        if (state.player) {
          this.updatePlayerInfo(state.player);
        }
        if (state.combat) {
          this.updateCombatInfo(state.combat);
        }
        this.emit("updateUI", { state });
      } catch (error) {
        this.uiLogger.error("Failed to update UI:", error);
      }
    }
    /**
     * Add a message to the game log
     */
    logMessage(message, type = "info") {
      if (!this.isInitialized || !this.elements?.logContent) {
        this.uiLogger.warn("Cannot log message - UI not initialized");
        return;
      }
      try {
        const timestamp = (/* @__PURE__ */ new Date()).toLocaleTimeString();
        const messageElement = document.createElement("div");
        messageElement.className = `log-entry log-${type}`;
        messageElement.innerHTML = `<span class="log-time">[${timestamp}]</span> ${message}`;
        this.elements.logContent.appendChild(messageElement);
        this.elements.logContent.scrollTop = this.elements.logContent.scrollHeight;
        const maxEntries = 100;
        const entries = this.elements.logContent.children;
        while (entries.length > maxEntries) {
          entries[0].remove();
        }
        this.emit("logMessage", { message, type });
      } catch (error) {
        this.uiLogger.error("Failed to log message:", error);
      }
    }
    /**
     * Update combat log with new information
     */
    updateCombatLog(data) {
      if (data.message) {
        this.logMessage(data.message, data.type || "combat");
      }
      if (data.initiative) {
        this.updateInitiativeOrder(data.initiative);
      }
    }
    /**
     * Update initiative order display
     */
    updateInitiativeOrder(initiative) {
      if (!this.elements?.initiativeList) {
        return;
      }
      try {
        this.elements.initiativeList.innerHTML = "";
        const sorted = [...initiative].sort((a, b) => b.initiative - a.initiative);
        sorted.forEach((entry, index) => {
          const entryElement = document.createElement("div");
          entryElement.className = `initiative-entry ${entry.isPlayer ? "player" : "npc"}`;
          entryElement.innerHTML = `
          <span class="initiative-order">${index + 1}.</span>
          <span class="initiative-name">${entry.name}</span>
          <span class="initiative-score">${entry.initiative}</span>
        `;
          this.elements.initiativeList.appendChild(entryElement);
        });
      } catch (error) {
        this.uiLogger.error("Failed to update initiative order:", error);
      }
    }
    /**
     * Find and return all required UI elements
     */
    findUIElements() {
      const elements = {
        // Character display elements
        portrait: document.getElementById("character-portrait"),
        name: document.getElementById("character-name"),
        // Stat elements
        stats: {
          str: document.getElementById("stat-str"),
          dex: document.getElementById("stat-dex"),
          con: document.getElementById("stat-con"),
          int: document.getElementById("stat-int"),
          wis: document.getElementById("stat-wis"),
          cha: document.getElementById("stat-cha")
        },
        // Health bar
        hpBar: document.getElementById("hp-bar"),
        hpText: document.getElementById("hp-text"),
        // Log elements
        logContent: document.getElementById("log-content"),
        // Combat elements
        initiativeList: document.getElementById("initiative-list"),
        // Control buttons
        actionButtons: {
          attack: document.getElementById("btn-attack"),
          defend: document.getElementById("btn-defend"),
          cast: document.getElementById("btn-cast"),
          item: document.getElementById("btn-item")
        },
        // Direction buttons
        directionButtons: {
          north: document.getElementById("btn-north"),
          south: document.getElementById("btn-south"),
          east: document.getElementById("btn-east"),
          west: document.getElementById("btn-west"),
          northeast: document.getElementById("btn-northeast"),
          northwest: document.getElementById("btn-northwest"),
          southeast: document.getElementById("btn-southeast"),
          southwest: document.getElementById("btn-southwest")
        }
      };
      return elements;
    }
    /**
     * Validate that all required UI elements were found
     */
    validateElements() {
      if (!this.elements) {
        throw new Error("UI elements not initialized");
      }
      const missingElements = [];
      if (!this.elements.logContent)
        missingElements.push("log-content");
      if (!this.elements.hpBar)
        missingElements.push("hp-bar");
      if (missingElements.length > 0) {
        throw new Error(`Missing required UI elements: ${missingElements.join(", ")}`);
      }
    }
    /**
     * Set up event listeners for UI interactions
     */
    setupEventListeners() {
      if (!this.elements)
        return;
      Object.entries(this.elements.actionButtons).forEach(([action, button]) => {
        if (button) {
          button.addEventListener("click", () => {
            this.emit("action", { action });
            this.uiLogger.debug("Action button clicked", { action });
          });
        }
      });
      Object.entries(this.elements.directionButtons).forEach(([direction, button]) => {
        if (button) {
          button.addEventListener("click", () => {
            this.emit("move", { direction });
            this.uiLogger.debug("Direction button clicked", { direction });
          });
        }
      });
    }
    /**
     * Set up keyboard controls for game navigation
     */
    setupKeyboardControls() {
      const keyMap = {
        "ArrowUp": "north",
        "ArrowDown": "south",
        "ArrowLeft": "west",
        "ArrowRight": "east",
        "w": "north",
        "s": "south",
        "a": "west",
        "d": "east",
        "q": "northwest",
        "e": "northeast",
        "z": "southwest",
        "c": "southeast"
      };
      const keyboardHandler = (event) => {
        const key = event.key.toLowerCase();
        const direction = keyMap[event.key] || keyMap[key];
        if (direction) {
          event.preventDefault();
          this.emit("move", { direction });
          this.uiLogger.debug("Keyboard movement", { key, direction });
        }
      };
      document.addEventListener("keydown", keyboardHandler);
      this.keyboardHandlers.set("movement", keyboardHandler);
    }
    /**
     * Update player information display
     */
    updatePlayerInfo(player) {
      if (!this.elements)
        return;
      try {
        if (player.name && this.elements.name) {
          this.elements.name.textContent = player.name;
        }
        if (player.attributes && this.elements.stats) {
          const stats = this.elements.stats;
          if (stats.str)
            stats.str.textContent = player.attributes.strength.toString();
          if (stats.dex)
            stats.dex.textContent = player.attributes.dexterity.toString();
          if (stats.con)
            stats.con.textContent = player.attributes.constitution.toString();
          if (stats.int)
            stats.int.textContent = player.attributes.intelligence.toString();
          if (stats.wis)
            stats.wis.textContent = player.attributes.wisdom.toString();
          if (stats.cha)
            stats.cha.textContent = player.attributes.charisma.toString();
        }
        if (player.hp && this.elements.hpBar) {
          const percentage = player.hp.current / player.hp.max * 100;
          this.elements.hpBar.style.width = `${percentage}%`;
          if (this.elements.hpText) {
            this.elements.hpText.textContent = `${player.hp.current}/${player.hp.max}`;
          }
        }
      } catch (error) {
        this.uiLogger.error("Failed to update player info:", error);
      }
    }
    /**
     * Update combat information display
     */
    updateCombatInfo(combat) {
      if (!this.elements)
        return;
      try {
        if (combat.initiative.length > 0) {
          this.updateInitiativeOrder(combat.initiative);
        }
        if (combat.active) {
          this.logMessage(`Combat Round ${combat.round}`, "combat");
          if (combat.currentTurn) {
            this.logMessage(`${combat.currentTurn}'s turn`, "info");
          }
        } else {
          this.logMessage("Combat ended", "combat");
        }
      } catch (error) {
        this.uiLogger.error("Failed to update combat info:", error);
      }
    }
  };
  var gameUI = new GameUI();

  // src/game/GameState.ts
  var GameState = class extends TypedEventEmitter {
    constructor() {
      super();
      this.stateLogger = logger.createChildLogger("GameState");
      this._state = null;
      this._initialized = false;
      this.stateLogger.info("GameState manager created");
    }
    /**
     * Initialize the game state
     */
    async initialize() {
      if (this._initialized) {
        this.stateLogger.warn("Game state already initialized");
        return;
      }
      try {
        this.stateLogger.info("Initializing game state...");
        this._state = this.createDefaultState();
        this._initialized = true;
        this.stateLogger.info("Game state initialized successfully");
        this.emit("stateChanged", { state: this._state });
      } catch (error) {
        this.stateLogger.error("Failed to initialize game state:", error);
        this.emit("error", { error });
        throw error;
      }
    }
    /**
     * Get current game state
     */
    get state() {
      return this._state;
    }
    /**
     * Check if game state is initialized
     */
    get initialized() {
      return this._initialized;
    }
    /**
     * Update the entire game state
     */
    updateState(newState) {
      if (!this._initialized || !this._state) {
        this.stateLogger.warn("Cannot update state - not initialized");
        return;
      }
      try {
        this._state = { ...this._state, ...newState, lastUpdate: Date.now() };
        this.stateLogger.debug("Game state updated");
        this.emit("stateChanged", { state: this._state });
      } catch (error) {
        this.stateLogger.error("Failed to update state:", error);
        this.emit("error", { error });
      }
    }
    /**
     * Update player state
     */
    updatePlayer(playerUpdates) {
      if (!this._initialized || !this._state || !this._state.player) {
        this.stateLogger.warn("Cannot update player - not initialized or no player");
        return;
      }
      try {
        const updatedPlayer = { ...this._state.player, ...playerUpdates };
        this._state = {
          ...this._state,
          player: updatedPlayer,
          lastUpdate: Date.now()
        };
        this.stateLogger.debug("Player state updated");
        this.emit("playerChanged", { player: updatedPlayer });
        this.emit("stateChanged", { state: this._state });
      } catch (error) {
        this.stateLogger.error("Failed to update player:", error);
        this.emit("error", { error });
      }
    }
    /**
     * Update combat state
     */
    updateCombat(combatUpdates) {
      if (!this._initialized || !this._state) {
        this.stateLogger.warn("Cannot update combat - not initialized");
        return;
      }
      try {
        const updatedCombat = this._state.combat ? { ...this._state.combat, ...combatUpdates } : combatUpdates;
        this._state = {
          ...this._state,
          combat: updatedCombat,
          lastUpdate: Date.now()
        };
        this.stateLogger.debug("Combat state updated");
        this.emit("combatChanged", { combat: updatedCombat });
        this.emit("stateChanged", { state: this._state });
      } catch (error) {
        this.stateLogger.error("Failed to update combat:", error);
        this.emit("error", { error });
      }
    }
    /**
     * Reset the game state to default
     */
    reset() {
      try {
        this.stateLogger.info("Resetting game state...");
        this._state = this.createDefaultState();
        this.emit("stateChanged", { state: this._state });
        this.stateLogger.info("Game state reset successfully");
      } catch (error) {
        this.stateLogger.error("Failed to reset state:", error);
        this.emit("error", { error });
      }
    }
    /**
     * Clean up resources
     */
    cleanup() {
      this.stateLogger.info("Cleaning up game state...");
      this._state = null;
      this._initialized = false;
      this.removeAllListeners();
      this.stateLogger.info("Game state cleanup completed");
    }
    /**
     * Create default game state
     */
    createDefaultState() {
      const defaultWorld = {
        map: {
          width: 20,
          height: 20,
          tiles: [],
          objects: []
        },
        objects: [],
        regions: []
      };
      const defaultPlayer = {
        id: "player-1",
        name: "Hero",
        position: { x: 5, y: 5 },
        health: 20,
        maxHealth: 20,
        level: 1,
        experience: 0,
        class: "Fighter",
        attributes: {
          strength: 15,
          dexterity: 14,
          constitution: 16,
          intelligence: 12,
          wisdom: 13,
          charisma: 11
        },
        equipment: {
          accessories: []
        }
      };
      const defaultCombat = {
        active: false,
        currentTurn: null,
        initiative: [],
        round: 0
      };
      return {
        player: defaultPlayer,
        world: defaultWorld,
        combat: defaultCombat,
        initialized: true,
        lastUpdate: Date.now()
      };
    }
  };
  var gameState = new GameState();

  // src/main.ts
  GlobalErrorHandler.setupGlobalHandlers();
  var componentManager = new ComponentManager();
  var GoldBoxRPGApp = class {
    constructor() {
      this.logger = logger.createChildLogger("GoldBoxRPGApp");
      this.initialized = false;
    }
    /**
     * Initialize the application
     */
    async initialize() {
      if (this.initialized) {
        this.logger.warn("Application already initialized");
        return;
      }
      try {
        this.logger.info("Initializing GoldBox RPG Engine...");
        await componentManager.initializeAll();
        this.initialized = true;
        this.logger.info("GoldBox RPG Engine initialized successfully");
        if (typeof window !== "undefined") {
          window.dispatchEvent(new CustomEvent("goldbox-ready", {
            detail: { app: this }
          }));
        }
      } catch (error) {
        this.logger.error("Failed to initialize application:", error);
        throw error;
      }
    }
    /**
     * Cleanup and shutdown the application
     */
    async cleanup() {
      if (!this.initialized) {
        return;
      }
      try {
        this.logger.info("Shutting down GoldBox RPG Engine...");
        await componentManager.cleanupAll();
        this.initialized = false;
        this.logger.info("GoldBox RPG Engine shut down successfully");
      } catch (error) {
        this.logger.error("Error during application cleanup:", error);
      }
    }
    /**
     * Get the component manager for accessing game systems
     */
    getComponentManager() {
      return componentManager;
    }
    /**
     * Check if application is initialized
     */
    isInitialized() {
      return this.initialized;
    }
  };
  var app = new GoldBoxRPGApp();
  if (typeof document !== "undefined") {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", () => {
        app.initialize().catch((error) => {
          console.error("Failed to auto-initialize application:", error);
        });
      });
    } else {
      app.initialize().catch((error) => {
        console.error("Failed to auto-initialize application:", error);
      });
    }
  }
  if (typeof window !== "undefined") {
    window.addEventListener("beforeunload", () => {
      app.cleanup().catch((error) => {
        console.error("Error during cleanup:", error);
      });
    });
  }
  if (typeof window !== "undefined") {
    window.GoldBoxRPG = {
      app,
      logger,
      componentManager,
      // Expose other utilities for migration period
      ErrorHandler: GlobalErrorHandler
    };
  }
  return __toCommonJS(main_exports);
})();
