/**
 * Production-safe logging utility with TypeScript support
 * Provides structured logging with performance considerations and environment detection
 */

import type { LogLevel, LoggerConfig, LogEntry } from '../types/GameTypes';

interface LogQueueEntry {
  readonly level: LogLevel;
  readonly timestamp: number;
  readonly component?: string;
  readonly message: string;
  readonly args: readonly unknown[];
}

export class Logger {
  private readonly config: LoggerConfig;
  private readonly logQueue: LogQueueEntry[] = [];
  private readonly maxQueueSize: number = 100;
  
  // Bind console methods for performance
  private readonly consoleDebug: typeof console.debug;
  private readonly consoleInfo: typeof console.info;
  private readonly consoleWarn: typeof console.warn;
  private readonly consoleError: typeof console.error;
  private readonly consoleGroup: typeof console.group;
  private readonly consoleGroupCollapsed: typeof console.groupCollapsed;
  private readonly consoleGroupEnd: typeof console.groupEnd;

  constructor(_component: string = 'Logger') {
    this.config = this.createConfig();
    
    // Bind console methods for performance optimization
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
  private createConfig(): LoggerConfig {
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
  private detectDevelopmentEnvironment(): boolean {
    if (typeof window === 'undefined') {
      return false; // Server-side rendering or Node.js
    }

    const hostname = window.location.hostname;
    
    // Development indicators
    const developmentHosts = [
      'localhost',
      '127.0.0.1',
      '0.0.0.0'
    ];

    // Check for common development patterns
    const isDevelopmentHost = developmentHosts.includes(hostname);
    const isLocalIP = /^192\.168\.|^10\.|^172\.(1[6-9]|2\d|3[01])\./.test(hostname);
    const isVSCodeLocal = hostname.includes('vscode-local');
    const isCodespaces = hostname.includes('githubpreview') || hostname.includes('app.github.dev');
    const isGitpod = hostname.includes('gitpod.io');
    const isPreviewApp = hostname.includes('preview.app') || hostname.includes('netlify.app');
    
    return isDevelopmentHost || isLocalIP || isVSCodeLocal || isCodespaces || isGitpod || isPreviewApp;
  }

  /**
   * Gets enabled log levels based on environment
   */
  private getEnabledLevels(isDevelopment: boolean): ReadonlySet<LogLevel> {
    if (isDevelopment) {
      return new Set<LogLevel>(['debug', 'info', 'warn', 'error', 'group']);
    } else {
      // Production: only warnings and errors
      return new Set<LogLevel>(['warn', 'error']);
    }
  }

  /**
   * Adds entry to log queue for monitoring
   */
  private addToQueue(entry: LogQueueEntry): void {
    this.logQueue.push(entry);
    
    // Maintain queue size limit
    if (this.logQueue.length > this.maxQueueSize) {
      this.logQueue.shift();
    }
  }

  /**
   * Logs a debug message (development only)
   */
  debug(message: string, ...args: unknown[]): void {
    this.log('debug', message, ...args);
  }

  /**
   * Logs an info message (development only)
   */
  info(message: string, ...args: unknown[]): void {
    this.log('info', message, ...args);
  }

  /**
   * Logs a warning message (always enabled)
   */
  warn(message: string, ...args: unknown[]): void {
    this.log('warn', message, ...args);
  }

  /**
   * Logs an error message (always enabled)
   */
  error(message: string, ...args: unknown[]): void {
    this.log('error', message, ...args);
  }

  /**
   * Starts a console group (development only)
   */
  group(label: string, ...args: unknown[]): void {
    if (this.config.enabledLevels.has('group')) {
      this.consoleGroup(label, ...args);
      this.addToQueue({
        level: 'info',
        message: `GROUP: ${label}`,
        timestamp: Date.now(),
        args: [label, ...args]
      });
    }
  }

  /**
   * Starts a collapsed console group (development only)
   */
  groupCollapsed(label: string, ...args: unknown[]): void {
    if (this.config.enabledLevels.has('group')) {
      this.consoleGroupCollapsed(label, ...args);
      this.addToQueue({
        level: 'info',
        message: `GROUP_COLLAPSED: ${label}`,
        timestamp: Date.now(),
        args: [label, ...args]
      });
    }
  }

  /**
   * Ends a console group (development only)
   */
  groupEnd(): void {
    if (this.config.enabledLevels.has('group')) {
      this.consoleGroupEnd();
      this.addToQueue({
        level: 'info',
        message: 'GROUP_END',
        timestamp: Date.now(),
        args: []
      });
    }
  }

  /**
   * Conditionally logs based on environment and level
   */
  private log(level: LogLevel, message: string, ...args: unknown[]): void {
    if (!this.config.enabledLevels.has(level)) {
      return;
    }

    const timestamp = Date.now();
    const entry: LogQueueEntry = {
      level,
      message,
      timestamp,
      args
    };

    this.addToQueue(entry);

    // Log to console based on level
    switch (level) {
      case 'debug':
        this.consoleDebug(message, ...args);
        break;
      case 'info':
        this.consoleInfo(message, ...args);
        break;
      case 'warn':
        this.consoleWarn(message, ...args);
        break;
      case 'error':
        this.consoleError(message, ...args);
        break;
    }
  }

  /**
   * Enables debug logging temporarily (useful for production debugging)
   */
  enableDebug(duration: number = 60000): void {
    const originalLevels = this.config.enabledLevels;
    
    // Create new set with debug enabled
    const newLevels = new Set<LogLevel>([...Array.from(originalLevels), 'debug', 'info', 'group']);
    (this.config as { enabledLevels: ReadonlySet<LogLevel> }).enabledLevels = newLevels;

    this.info(`Debug logging enabled for ${duration}ms`);

    // Restore original levels after duration
    setTimeout(() => {
      (this.config as { enabledLevels: ReadonlySet<LogLevel> }).enabledLevels = originalLevels;
      this.info('Debug logging disabled');
    }, duration);
  }

  /**
   * Gets current logging configuration
   */
  getConfig(): Readonly<LoggerConfig> {
    return this.config;
  }

  /**
   * Gets recent log entries from the queue
   */
  getRecentLogs(count: number = 50): readonly LogEntry[] {
    return this.logQueue
      .slice(-count)
      .map(({ args, ...entry }) => entry); // Remove args for external access
  }

  /**
   * Clears the log queue
   */
  clearQueue(): void {
    this.logQueue.length = 0;
  }

  /**
   * Creates a child logger with a component prefix
   */
  createChildLogger(component: string): Logger {
    const childLogger = new Logger(component);
    
    // Override log method to include component prefix
    const originalLog = childLogger.log.bind(childLogger);
    (childLogger as any).log = (level: LogLevel, message: string, ...args: unknown[]) => {
      originalLog(level, `[${component}] ${message}`, ...args);
    };

    return childLogger;
  }

  /**
   * Formats a timestamp for logging
   */
  static formatTimestamp(timestamp: number): string {
    return new Date(timestamp).toISOString();
  }

  /**
   * Safely stringifies an object for logging
   */
  static safeStringify(obj: unknown): string {
    try {
      return JSON.stringify(obj, null, 2);
    } catch (error) {
      return `[Object: ${String(obj)}]`;
    }
  }
}

// Create and export global logger instance
export const logger = new Logger('Global');

// For backward compatibility with existing code
if (typeof window !== 'undefined' && !logger.getConfig().isDevelopment) {
  // In production, replace console methods with logger methods
  console.debug = logger.debug.bind(logger);
  console.info = logger.info.bind(logger);
  console.group = logger.group.bind(logger);
  console.groupEnd = logger.groupEnd.bind(logger);
  console.groupCollapsed = logger.groupCollapsed.bind(logger);
  // Leave warn and error as-is for production debugging
}

// Expose logger globally for manual use
if (typeof window !== 'undefined') {
  (window as any).logger = logger;
}
