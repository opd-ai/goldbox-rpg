/**
 * Main entry point for the GoldBox RPG Engine TypeScript client
 * Re-exports all core modules and provides initialization utilities
 */
// Core exports
export { EventEmitter, TypedEventEmitter } from './core/EventEmitter';
export { BaseComponent, BaseService, ComponentManager } from './core/BaseComponent';
// Utility exports
export { Logger, logger } from './utils/Logger';
export { ErrorHandler, createErrorHandler, GlobalErrorHandler } from './utils/ErrorHandler';
export { SpatialQueryManager } from './utils/SpatialQueryManager';
// Re-export for backward compatibility
export { logger as default } from './utils/Logger';
// Network exports
export { RPCClient, rpcClient } from './network/RPCClient';
// UI exports  
export { GameUI, gameUI } from './ui/GameUI';
// Game exports
export { GameState, gameState } from './game/GameState';
//# sourceMappingURL=index.js.map