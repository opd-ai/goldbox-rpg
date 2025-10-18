/**
 * Production-safe logging utility with TypeScript support
 * Provides structured logging with performance considerations and environment detection
 */
import type { LoggerConfig, LogEntry } from '../types/GameTypes';
export declare class Logger {
    private readonly config;
    private readonly logQueue;
    private readonly maxQueueSize;
    private readonly consoleDebug;
    private readonly consoleInfo;
    private readonly consoleWarn;
    private readonly consoleError;
    private readonly consoleGroup;
    private readonly consoleGroupCollapsed;
    private readonly consoleGroupEnd;
    constructor(_component?: string);
    /**
     * Creates logger configuration based on environment detection
     */
    private createConfig;
    /**
     * Detects if running in development environment
     */
    private detectDevelopmentEnvironment;
    /**
     * Gets enabled log levels based on environment
     */
    private getEnabledLevels;
    /**
     * Adds entry to log queue for monitoring
     */
    private addToQueue;
    /**
     * Logs a debug message (development only)
     */
    debug(message: string, ...args: unknown[]): void;
    /**
     * Logs an info message (development only)
     */
    info(message: string, ...args: unknown[]): void;
    /**
     * Logs a warning message (always enabled)
     */
    warn(message: string, ...args: unknown[]): void;
    /**
     * Logs an error message (always enabled)
     */
    error(message: string, ...args: unknown[]): void;
    /**
     * Starts a console group (development only)
     */
    group(label: string, ...args: unknown[]): void;
    /**
     * Starts a collapsed console group (development only)
     */
    groupCollapsed(label: string, ...args: unknown[]): void;
    /**
     * Ends a console group (development only)
     */
    groupEnd(): void;
    /**
     * Conditionally logs based on environment and level
     */
    private log;
    /**
     * Enables debug logging temporarily (useful for production debugging)
     */
    enableDebug(duration?: number): void;
    /**
     * Gets current logging configuration
     */
    getConfig(): Readonly<LoggerConfig>;
    /**
     * Gets recent log entries from the queue
     */
    getRecentLogs(count?: number): readonly LogEntry[];
    /**
     * Clears the log queue
     */
    clearQueue(): void;
    /**
     * Creates a child logger with a component prefix
     */
    createChildLogger(component: string): Logger;
    /**
     * Formats a timestamp for logging
     */
    static formatTimestamp(timestamp: number): string;
    /**
     * Safely stringifies an object for logging
     */
    static safeStringify(obj: unknown): string;
}
export declare const logger: Logger;
//# sourceMappingURL=Logger.d.ts.map