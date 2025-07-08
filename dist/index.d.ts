/**
 * Main entry point for the GoldBox RPG Engine TypeScript client
 * Re-exports all core modules and provides initialization utilities
 */
export type * from './types/GameTypes';
export type { EventEmitterInterface, UIComponent, GameUIElements, KeyboardDirection, KeyboardEventMap, CanvasLayers, CanvasContexts, SpriteMap, CombatUIState, ActionButton, MessageType, GameMessage, UIEventMap, ErrorDisplayOptions, Viewport, CameraTarget } from './types/UITypes';
export type * from './types/RPCTypes';
export { EventEmitter, TypedEventEmitter } from './core/EventEmitter';
export { BaseComponent, BaseService, ComponentManager } from './core/BaseComponent';
export { Logger, logger } from './utils/Logger';
export { ErrorHandler, createErrorHandler, GlobalErrorHandler } from './utils/ErrorHandler';
export { SpatialQueryManager } from './utils/SpatialQueryManager';
export { logger as default } from './utils/Logger';
export { RPCClient, rpcClient } from './network/RPCClient';
export { GameUI, gameUI } from './ui/GameUI';
export { GameState, gameState } from './game/GameState';
//# sourceMappingURL=index.d.ts.map