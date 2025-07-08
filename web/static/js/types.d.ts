/**
 * Type definitions for GoldBox RPG Engine
 * This file provides TypeScript type definitions for the game's core data structures
 * and interfaces, serving as a foundation for future TypeScript migration.
 */

// Basic game coordinate and position types
export interface Position {
  x: number;
  y: number;
}

export interface Size {
  width: number;
  height: number;
}

export interface Rectangle extends Position, Size {}

// Game state interfaces
export interface PlayerState {
  id: string;
  name: string;
  position: Position;
  health: number;
  maxHealth: number;
  level: number;
  experience: number;
  class: string;
  attributes: PlayerAttributes;
  equipment: Equipment;
}

export interface PlayerAttributes {
  strength: number;
  dexterity: number;
  constitution: number;
  intelligence: number;
  wisdom: number;
  charisma: number;
}

export interface Equipment {
  weapon?: Item;
  armor?: Item;
  shield?: Item;
  accessories: Item[];
}

export interface Item {
  id: string;
  name: string;
  type: ItemType;
  properties: ItemProperties;
  quantity: number;
}

export type ItemType = 'weapon' | 'armor' | 'shield' | 'accessory' | 'consumable' | 'misc';

export interface ItemProperties {
  damage?: number;
  defense?: number;
  weight?: number;
  value?: number;
  durability?: number;
  maxDurability?: number;
  [key: string]: any;
}

export interface WorldState {
  map: GameMap;
  players: { [id: string]: PlayerState };
  objects: GameObject[];
  dimensions: Size;
}

export interface GameObject {
  id: string;
  type: GameObjectType;
  position: Position;
  sprite?: string;
  properties: { [key: string]: any };
}

export type GameObjectType = 'static' | 'dynamic' | 'interactive' | 'decoration';

export interface GameMap {
  width: number;
  height: number;
  tiles: Tile[][];
  getTile(x: number, y: number): Tile | null;
}

export interface Tile {
  type: TileType;
  passable: boolean;
  sprite?: string;
  properties?: { [key: string]: any };
}

export type TileType = 'floor' | 'wall' | 'door' | 'stairs' | 'water' | 'void';

export interface CombatState {
  active: boolean;
  currentTurn?: string;
  initiative: InitiativeEntry[];
  round: number;
}

export interface InitiativeEntry {
  id: string;
  name: string;
  initiative: number;
  hasActed: boolean;
}

export interface GameState {
  player: PlayerState | null;
  world: WorldState | null;
  combat: CombatState | null;
  lastUpdate: number;
  initialized: boolean;
}

// RPC-related interfaces
export interface RPCRequest<TParams = any> {
  jsonrpc: '2.0';
  method: string;
  params?: TParams;
  id: number;
}

export interface RPCResponse<TResult = any> {
  jsonrpc: '2.0';
  result?: TResult;
  error?: RPCError;
  id: number | null;
}

export interface RPCError {
  code: number;
  message: string;
  data?: any;
}

// Method-specific parameter interfaces
export interface MoveParams {
  direction: MoveDirection;
}

export type MoveDirection = 'n' | 's' | 'e' | 'w' | 'ne' | 'nw' | 'se' | 'sw' | 
                           'up' | 'down' | 'left' | 'right';

export interface AttackParams {
  target_id: string;
  weapon_id: string;
}

export interface SpellParams {
  spell_id: string;
  target_id?: string;
  position?: Position;
  level?: number;
}

export interface JoinGameParams {
  player_name: string;
  character_class?: string;
}

export interface StartCombatParams {
  participants: string[];
}

// Event system interfaces
export interface GameEvents {
  'stateChanged': (state: GameState) => void;
  'error': (error: Error) => void;
  'playerMoved': (player: PlayerState, oldPosition: Position) => void;
  'combatStarted': (combat: CombatState) => void;
  'combatEnded': () => void;
  'playerJoined': (player: PlayerState) => void;
  'playerLeft': (playerId: string) => void;
}

// UI and rendering interfaces
export interface RenderOptions {
  showGrid?: boolean;
  showCoordinates?: boolean;
  highlightPlayer?: boolean;
  scale?: number;
}

export interface UIElements {
  portrait: HTMLElement | null;
  playerName: HTMLElement | null;
  healthBar: HTMLElement | null;
  logContent: HTMLElement | null;
  actionButtons: NodeListOf<HTMLButtonElement>;
  directionButtons: NodeListOf<HTMLButtonElement>;
}

// Configuration interfaces
export interface GameConfig {
  tileSize: number;
  updateInterval: number;
  maxLogEntries: number;
  debugMode: boolean;
  apiUrl: string;
  websocketUrl: string;
}

// Validation and error handling
export type ValidationResult<T> = {
  success: true;
  data: T;
} | {
  success: false;
  error: string;
};

export interface ErrorMetadata {
  context?: string;
  timestamp?: number;
  userAgent?: string;
  url?: string;
  [key: string]: any;
}

// Spatial query interfaces
export interface SpatialQuery {
  center: Position;
  radius?: number;
  bounds?: Rectangle;
  objectTypes?: GameObjectType[];
}

export interface SpatialResult {
  objects: GameObject[];
  players: PlayerState[];
  tiles: { position: Position; tile: Tile }[];
}

// Logger interfaces
export type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'group';

export interface LoggerConfig {
  isDevelopment: boolean;
  enabledLevels: LogLevel[];
  hostname: string;
  port: string;
  maxQueueSize: number;
}

// WebSocket-related interfaces
export interface WebSocketConfig {
  url: string;
  reconnectAttempts: number;
  reconnectDelay: number;
  maxReconnectDelay: number;
  timeout: number;
}

export interface ConnectionState {
  connected: boolean;
  connecting: boolean;
  reconnectAttempts: number;
  lastError?: Error;
}
