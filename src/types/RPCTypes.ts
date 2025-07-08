/**
 * JSON-RPC 2.0 type definitions for GoldBox RPG Engine
 * Provides strict typing for RPC communication with the Go backend
 */

// Core JSON-RPC 2.0 types
export interface RPCRequest<TParams = unknown> {
  readonly jsonrpc: '2.0';
  readonly method: string;
  readonly params?: TParams;
  readonly id: string | number | null;
}

export interface RPCResponse<TResult = unknown> {
  readonly jsonrpc: '2.0';
  readonly result?: TResult;
  readonly error?: RPCError;
  readonly id: string | number | null;
}

export interface RPCError {
  readonly code: number;
  readonly message: string;
  readonly data?: unknown;
}

// Standard JSON-RPC error codes
export const enum RPCErrorCode {
  PARSE_ERROR = -32700,
  INVALID_REQUEST = -32600,
  METHOD_NOT_FOUND = -32601,
  INVALID_PARAMS = -32602,
  INTERNAL_ERROR = -32603,
  // Server error range: -32099 to -32000
  SERVER_ERROR_MIN = -32099,
  SERVER_ERROR_MAX = -32000,
}

// Game-specific RPC method parameters
export interface JoinGameParams {
  readonly player_name: string;
}

export interface MoveParams {
  readonly session_id: string;
  readonly direction: string;
}

export interface AttackParams {
  readonly session_id: string;
  readonly target_id: string;
  readonly weapon?: string;
}

export interface SpellParams {
  readonly session_id: string;
  readonly spell_id: string;
  readonly target_id?: string;
  readonly target_position?: {
    readonly x: number;
    readonly y: number;
  };
}

export interface StartCombatParams {
  readonly session_id: string;
  readonly enemy_ids: readonly string[];
}

export interface GetGameStateParams {
  readonly session_id: string;
}

export interface SpatialQueryParams {
  readonly session_id: string;
  readonly query_type: 'range' | 'radius' | 'nearest';
  readonly params: Readonly<{
    minX?: number;
    minY?: number;
    maxX?: number;
    maxY?: number;
    x?: number;
    y?: number;
    radius?: number;
    k?: number;
    object_type?: string;
  }>;
}

// Game-specific RPC method results
export interface JoinGameResult {
  readonly session_id: string;
  readonly player_id: string;
  readonly success: boolean;
}

export interface MoveResult {
  readonly success: boolean;
  readonly new_position?: {
    readonly x: number;
    readonly y: number;
  };
  readonly message?: string;
}

export interface AttackResult {
  readonly success: boolean;
  readonly damage?: number;
  readonly target_health?: number;
  readonly message: string;
}

export interface SpellResult {
  readonly success: boolean;
  readonly effects?: readonly string[];
  readonly message: string;
}

export interface GameStateResult {
  readonly player: unknown;
  readonly world: unknown;
  readonly combat: unknown;
  readonly timestamp: number;
}

export interface SpatialQueryResult {
  readonly objects: readonly unknown[];
  readonly count: number;
}

// Method name type union for type safety
export type RPCMethodName = 
  | 'joinGame'
  | 'move'
  | 'attack'
  | 'castSpell'
  | 'startCombat'
  | 'getGameState'
  | 'spatialQuery'
  | 'leaveGame';

// Mapping of method names to their parameter and result types
export interface RPCMethodMap {
  'joinGame': {
    params: JoinGameParams;
    result: JoinGameResult;
  };
  'move': {
    params: MoveParams;
    result: MoveResult;
  };
  'attack': {
    params: AttackParams;
    result: AttackResult;
  };
  'castSpell': {
    params: SpellParams;
    result: SpellResult;
  };
  'startCombat': {
    params: StartCombatParams;
    result: unknown;
  };
  'getGameState': {
    params: GetGameStateParams;
    result: GameStateResult;
  };
  'spatialQuery': {
    params: SpatialQueryParams;
    result: SpatialQueryResult;
  };
  'leaveGame': {
    params: { readonly session_id: string };
    result: { readonly success: boolean };
  };
}

// Type-safe RPC method call interface
export type TypedRPCCall = <T extends RPCMethodName>(
  method: T,
  params: RPCMethodMap[T]['params']
) => Promise<RPCMethodMap[T]['result']>;

// WebSocket message types
export interface WebSocketMessage {
  readonly type: 'rpc_request' | 'rpc_response' | 'state_update' | 'error';
  readonly data: unknown;
  readonly timestamp: number;
}

export interface StateUpdateMessage {
  readonly type: 'state_update';
  readonly data: {
    readonly player?: unknown;
    readonly world?: unknown;
    readonly combat?: unknown;
  };
  readonly timestamp: number;
}

// RPC client configuration
export interface RPCClientConfig {
  readonly baseUrl: string;
  readonly timeout: number;
  readonly maxReconnectAttempts: number;
  readonly reconnectBackoffBase: number;
  readonly reconnectBackoffMax: number;
  readonly enableLogging: boolean;
  readonly validateOrigin: boolean;
}

// Request tracking
export interface PendingRequest {
  readonly id: string | number;
  readonly method: string;
  readonly timestamp: number;
  readonly resolve: (value: unknown) => void;
  readonly reject: (reason: unknown) => void;
}

// Session management types
export interface SessionData {
  readonly session_id: string;
  readonly player_id?: string;
  readonly expiry: number;
  readonly created: number;
}

export interface SessionValidationResult {
  readonly valid: boolean;
  readonly expired: boolean;
  readonly error?: string;
}

// Session information
export interface SessionInfo {
  readonly sessionId: string;
  readonly expiresAt: Date;
  readonly isValid: boolean;
}

// RPC method names type
export type RPCMethod = keyof RPCMethodMap;
