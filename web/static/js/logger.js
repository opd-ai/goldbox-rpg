/**
 * Production-safe logging utility that respects environment settings
 * and provides structured logging with performance considerations.
 * 
 * Automatically detects environment and adjusts logging levels accordingly.
 * In production, only errors and warnings are logged to reduce noise.
 */
class Logger {
  constructor() {
    this.isDevelopment = this.detectDevelopmentEnvironment();
    this.enabledLevels = this.getEnabledLevels();
    this.logQueue = [];
    this.maxQueueSize = 100;
    
    // Bind console methods for performance
    this.originalConsole = {
      debug: console.debug.bind(console),
      info: console.info.bind(console),
      warn: console.warn.bind(console),
      error: console.error.bind(console),
      group: console.group.bind(console),
      groupEnd: console.groupEnd.bind(console),
      groupCollapsed: console.groupCollapsed.bind(console)
    };
  }

  /**
   * Detects if running in development environment
   * @returns {boolean} True if in development mode
   */
  detectDevelopmentEnvironment() {
    // Check various indicators of development environment
    return (
      location.hostname === 'localhost' ||
      location.hostname === '127.0.0.1' ||
      location.hostname.endsWith('.local') ||
      location.hostname.endsWith('.dev') ||
      location.hostname.endsWith('.github.dev') ||
      location.hostname.endsWith('.gitpod.io') ||
      location.port !== '' ||
      localStorage.getItem('debug') === 'true' ||
      window.location.search.includes('debug=true')
    );
  }

  /**
   * Gets enabled log levels based on environment
   * @returns {Set<string>} Set of enabled log levels
   */
  getEnabledLevels() {
    if (this.isDevelopment) {
      return new Set(['debug', 'info', 'warn', 'error', 'group']);
    } else {
      // Production: only warnings and errors
      return new Set(['warn', 'error']);
    }
  }

  /**
   * Logs a debug message (development only)
   * @param {...any} args - Arguments to log
   */
  debug(...args) {
    if (this.enabledLevels.has('debug')) {
      this.originalConsole.debug(...args);
    }
  }

  /**
   * Logs an info message (development only)
   * @param {...any} args - Arguments to log
   */
  info(...args) {
    if (this.enabledLevels.has('info')) {
      this.originalConsole.info(...args);
    }
  }

  /**
   * Logs a warning message (always enabled)
   * @param {...any} args - Arguments to log
   */
  warn(...args) {
    if (this.enabledLevels.has('warn')) {
      this.originalConsole.warn(...args);
    }
  }

  /**
   * Logs an error message (always enabled)
   * @param {...any} args - Arguments to log
   */
  error(...args) {
    if (this.enabledLevels.has('error')) {
      this.originalConsole.error(...args);
    }
  }

  /**
   * Starts a console group (development only)
   * @param {...any} args - Arguments for group label
   */
  group(...args) {
    if (this.enabledLevels.has('group')) {
      this.originalConsole.group(...args);
    }
  }

  /**
   * Starts a collapsed console group (development only)
   * @param {...any} args - Arguments for group label
   */
  groupCollapsed(...args) {
    if (this.enabledLevels.has('group')) {
      this.originalConsole.groupCollapsed(...args);
    }
  }

  /**
   * Ends a console group (development only)
   */
  groupEnd() {
    if (this.enabledLevels.has('group')) {
      this.originalConsole.groupEnd();
    }
  }

  /**
   * Conditionally logs based on environment
   * @param {string} level - Log level ('debug', 'info', 'warn', 'error')
   * @param {...any} args - Arguments to log
   */
  log(level, ...args) {
    if (this.enabledLevels.has(level) && this[level]) {
      this[level](...args);
    }
  }

  /**
   * Enables debug logging temporarily (useful for production debugging)
   * @param {number} [duration=60000] - Duration in milliseconds (default 1 minute)
   */
  enableDebug(duration = 60000) {
    const originalLevels = new Set(this.enabledLevels);
    this.enabledLevels.add('debug');
    this.enabledLevels.add('info');
    this.enabledLevels.add('group');
    
    this.info('Debug logging enabled for', duration, 'ms');
    
    setTimeout(() => {
      this.enabledLevels = originalLevels;
      this.info('Debug logging disabled, returning to production mode');
    }, duration);
  }

  /**
   * Gets current logging configuration
   * @returns {Object} Current configuration
   */
  getConfig() {
    return {
      isDevelopment: this.isDevelopment,
      enabledLevels: Array.from(this.enabledLevels),
      hostname: location.hostname,
      port: location.port
    };
  }
}

// Create global logger instance
const logger = new Logger();

// For backward compatibility, augment console with environment-aware logging
// This allows existing code to work without modification
if (!logger.isDevelopment) {
  // In production, replace console methods with logger methods
  console.debug = logger.debug.bind(logger);
  console.info = logger.info.bind(logger);
  console.group = logger.group.bind(logger);
  console.groupEnd = logger.groupEnd.bind(logger);
  console.groupCollapsed = logger.groupCollapsed.bind(logger);
  // Leave warn and error as-is for production debugging
}

// Expose logger globally for manual use
window.logger = logger;
