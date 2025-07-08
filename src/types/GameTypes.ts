/**
 * Core type definitions for GoldBox RPG Engine
 * Enhanced from existing types.d.ts with strict type safety
 */

// Basic game coordinate and position types
export interface Position {
  readonly x: number;
  readonly y: number;
}

export interface Size {
  readonly width: number;
  readonly height: number;
}

export interface Rectangle extends Position, Size {}

export interface Bounds {
  readonly minX: number;
  readonly minY: number;
  readonly maxX: number;
  readonly maxY: number;
}

// Game state interfaces with immutability
export interface PlayerAttributes {
  readonly strength: number;
  readonly dexterity: number;
  readonly constitution: number;
  readonly intelligence: number;
  readonly wisdom: number;
  readonly charisma: number;
}

export interface Equipment {
  readonly weapon?: Item;
  readonly armor?: Item;
  readonly shield?: Item;
  readonly accessories: readonly Item[];
}

export interface Item {
  readonly id: string;
  readonly name: string;
  readonly type: ItemType;
  readonly properties: Readonly<ItemProperties>;
  readonly quantity: number;
}

export type ItemType = 'weapon' | 'armor' | 'shield' | 'accessory' | 'consumable' | 'misc';

export interface ItemProperties {
  readonly [key: string]: unknown;
}

export interface PlayerState {
  readonly id: string;
  readonly name: string;
  readonly position: Position;
  readonly health: number;
  readonly maxHealth: number;
  readonly level: number;
  readonly experience: number;
  readonly class: string;
  readonly attributes: PlayerAttributes;
  readonly equipment: Equipment;
}

export interface GameObject {
  readonly id: string;
  readonly type: GameObjectType;
  readonly position: Position;
  readonly properties: Readonly<Record<string, unknown>>;
}

export type GameObjectType = 'static' | 'dynamic' | 'interactive' | 'decoration';

export interface Tile {
  readonly x: number;
  readonly y: number;
  readonly type: TileType;
  readonly passable: boolean;
  readonly properties: Readonly<Record<string, unknown>>;
}

export type TileType = 'floor' | 'wall' | 'door' | 'stairs' | 'water' | 'void';

export interface GameMap {
  readonly width: number;
  readonly height: number;
  readonly tiles: readonly Tile[];
  readonly objects: readonly GameObject[];
}

export interface InitiativeEntry {
  readonly id: string;
  readonly name: string;
  readonly initiative: number;
  readonly isPlayer: boolean;
}

export interface CombatState {
  readonly active: boolean;
  readonly currentTurn: string | null;
  readonly initiative: readonly InitiativeEntry[];
  readonly round: number;
}

export interface WorldState {
  readonly map: GameMap;
  readonly objects: readonly GameObject[];
  readonly regions: readonly Rectangle[];
}

export interface GameState {
  readonly player: PlayerState | null;
  readonly world: WorldState | null;
  readonly combat: CombatState | null;
  readonly initialized: boolean;
  readonly lastUpdate: number;
}

// Movement and direction types
export type MoveDirection = 'n' | 's' | 'e' | 'w' | 'ne' | 'nw' | 'se' | 'sw' | 
                           'up' | 'down' | 'left' | 'right';

// Event system types
export type EventCallback<T = unknown> = (data: T) => void;

export interface EventMap {
  readonly [event: string]: EventCallback;
}

// Validation and error handling
export type ValidationResult<T> = {
  readonly success: true;
  readonly data: T;
} | {
  readonly success: false;
  readonly error: string;
};

export interface ErrorMetadata {
  readonly [key: string]: unknown;
}

// Logger configuration
export type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'group';

export interface LogEntry {
  readonly level: LogLevel;
  readonly timestamp: number;
  readonly component?: string;
  readonly message: string;
}

export interface LoggerConfig {
  readonly isDevelopment: boolean;
  readonly enabledLevels: ReadonlySet<LogLevel>;
  readonly maxQueueSize: number;
}

// Spatial query types
export interface SpatialQuery {
  readonly type: 'range' | 'radius' | 'nearest';
  readonly params: Readonly<Record<string, unknown>>;
}

export interface SpatialResult {
  readonly objects: readonly GameObject[];
  readonly count: number;
}

// UI and rendering types
export interface RenderOptions {
  readonly clearLayers?: boolean;
  readonly updateOnly?: string[];
}

export interface UIElements {
  readonly [elementId: string]: HTMLElement;
}

// Configuration interfaces
export interface GameConfig {
  readonly updateInterval: number;
  readonly maxReconnectAttempts: number;
  readonly sessionTimeout: number;
  readonly wsUrl: string;
}

// WebSocket and connection types
export interface WebSocketConfig {
  readonly url: string;
  readonly protocols?: string[];
  readonly timeout: number;
  readonly maxReconnectAttempts: number;
}

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'error';

// Type guards and utility types
export type DeepReadonly<T> = {
  readonly [P in keyof T]: T[P] extends object ? DeepReadonly<T[P]> : T[P];
};

export type PartialExcept<T, K extends keyof T> = Partial<T> & Pick<T, K>;

export type RequireAtLeastOne<T, Keys extends keyof T = keyof T> =
  Pick<T, Exclude<keyof T, Keys>> & 
  {
    [K in Keys]-?: Required<Pick<T, K>> & Partial<Pick<T, Exclude<Keys, K>>>;
  }[Keys];

// Component lifecycle interface
export interface ComponentLifecycle {
  initialize(): Promise<void> | void;
  cleanup(): Promise<void> | void;
}

// Generic service interface
export interface Service extends ComponentLifecycle {
  readonly name: string;
  readonly initialized: boolean;
}

// Character type alias for backward compatibility
export type Character = PlayerState;
