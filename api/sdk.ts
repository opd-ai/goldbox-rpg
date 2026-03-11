/**
 * GoldBox RPG TypeScript SDK
 * Auto-generated from OpenAPI specification
 * Provides type-safe JSON-RPC 2.0 client for the GoldBox RPG engine
 */

import type { components, paths } from './client';

type RPCRequest = components['schemas']['RPCRequest'];
type RPCResponse = components['schemas']['RPCResponse'];
type RPCErrorType = components['schemas']['RPCError'];

export interface ClientConfig {
  baseURL: string;
  headers?: Record<string, string>;
  timeout?: number;
}

export interface RPCCallOptions {
  timeout?: number;
  headers?: Record<string, string>;
}

export class RPCClientError extends Error {
  code: number;
  data?: unknown;

  constructor(code: number, message: string, data?: unknown) {
    super(message);
    this.name = 'RPCClientError';
    this.code = code;
    this.data = data;
  }
}

export class GoldBoxRPGClient {
  private baseURL: string;
  private headers: Record<string, string>;
  private timeout: number;
  private requestId = 0;

  constructor(config: ClientConfig) {
    this.baseURL = config.baseURL;
    this.headers = {
      'Content-Type': 'application/json',
      ...config.headers,
    };
    this.timeout = config.timeout || 30000;
  }

  private async call<TResult = unknown>(
    method: string,
    params?: Record<string, unknown>,
    options?: RPCCallOptions
  ): Promise<TResult> {
    const id = ++this.requestId;
    const request = {
      jsonrpc: '2.0' as const,
      method,
      params: params || ({} as Record<string, never>),
      id,
    };

    const controller = new AbortController();
    const timeout = setTimeout(
      () => controller.abort(),
      options?.timeout || this.timeout
    );

    try {
      const response = await fetch(`${this.baseURL}/rpc`, {
        method: 'POST',
        headers: { ...this.headers, ...options?.headers },
        body: JSON.stringify(request),
        signal: controller.signal,
      });

      clearTimeout(timeout);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = (await response.json()) as RPCResponse;

      if ('error' in data && data.error) {
        const err = data.error as RPCErrorType;
        throw new RPCClientError(err.code, err.message, err.data);
      }

      if ('result' in data) {
        return data.result as TResult;
      }

      throw new Error('Invalid RPC response: missing result and error');
    } catch (error) {
      clearTimeout(timeout);
      throw error;
    }
  }

  // === Character Actions ===

  async move(params: {
    session_id: string;
    target_x: number;
    target_y: number;
  }): Promise<{ success: boolean; message: string; new_position: { x: number; y: number } }> {
    return this.call('move', params);
  }

  async attack(params: {
    session_id: string;
    target_id: string;
  }): Promise<{ success: boolean; damage: number; message: string }> {
    return this.call('attack', params);
  }

  async castSpell(params: {
    session_id: string;
    spell_id: string;
    target_id?: string;
    position?: { x: number; y: number };
  }): Promise<{ success: boolean; message: string; effects: unknown[] }> {
    return this.call('castSpell', params);
  }

  async useItem(params: {
    session_id: string;
    item_id: string;
    target_id?: string;
  }): Promise<{ success: boolean; message: string; effects: unknown[] }> {
    return this.call('useItem', params);
  }

  async applyEffect(params: {
    session_id: string;
    effect: unknown;
    target_id: string;
  }): Promise<{ success: boolean; message: string }> {
    return this.call('applyEffect', params);
  }

  // === Combat Management ===

  async startCombat(params: {
    session_id: string;
    enemy_ids: string[];
  }): Promise<{ success: boolean; combat_id: string; turn_order: string[] }> {
    return this.call('startCombat', params);
  }

  async endTurn(params: {
    session_id: string;
  }): Promise<{ success: boolean; next_character_id: string }> {
    return this.call('endTurn', params);
  }

  // === Game State ===

  async getGameState(params: {
    session_id: string;
  }): Promise<{ success: boolean; game_state: unknown }> {
    return this.call('getGameState', params);
  }

  async joinGame(params: {
    session_id: string;
    player_id: string;
  }): Promise<{ success: boolean; message: string }> {
    return this.call('joinGame', params);
  }

  async leaveGame(params: {
    session_id: string;
    player_id: string;
  }): Promise<{ success: boolean; message: string }> {
    return this.call('leaveGame', params);
  }

  async createCharacter(params: {
    session_id: string;
    name: string;
    character_class: string;
    stats: Record<string, number>;
    method?: string;
  }): Promise<{ success: boolean; character: unknown }> {
    return this.call('createCharacter', params);
  }

  // === Equipment ===

  async equipItem(params: {
    session_id: string;
    item_id: string;
    slot: string;
  }): Promise<{
    success: boolean;
    message: string;
    equipped_item: unknown;
    previous_item: unknown | null;
  }> {
    return this.call('equipItem', params);
  }

  async unequipItem(params: {
    session_id: string;
    slot: string;
  }): Promise<{ success: boolean; message: string; unequipped_item: unknown }> {
    return this.call('unequipItem', params);
  }

  async getEquipment(params: {
    session_id: string;
  }): Promise<{
    success: boolean;
    equipment: Record<string, unknown>;
    total_weight: number;
    equipment_bonuses: unknown;
  }> {
    return this.call('getEquipment', params);
  }

  // === Quest Management ===

  async startQuest(params: {
    session_id: string;
    quest: unknown;
  }): Promise<{ success: boolean; message: string }> {
    return this.call('startQuest', params);
  }

  async completeQuest(params: {
    session_id: string;
    quest_id: string;
  }): Promise<{ success: boolean; message: string; rewards: unknown }> {
    return this.call('completeQuest', params);
  }

  async updateObjective(params: {
    session_id: string;
    quest_id: string;
    objective_id: string;
    progress: number;
  }): Promise<{ success: boolean; message: string }> {
    return this.call('updateObjective', params);
  }

  async failQuest(params: {
    session_id: string;
    quest_id: string;
    reason?: string;
  }): Promise<{ success: boolean; message: string }> {
    return this.call('failQuest', params);
  }

  async getQuest(params: {
    session_id: string;
    quest_id: string;
  }): Promise<{ success: boolean; quest: unknown }> {
    return this.call('getQuest', params);
  }

  async getActiveQuests(params: {
    session_id: string;
  }): Promise<{ success: boolean; quests: unknown[] }> {
    return this.call('getActiveQuests', params);
  }

  async getCompletedQuests(params: {
    session_id: string;
  }): Promise<{ success: boolean; quests: unknown[] }> {
    return this.call('getCompletedQuests', params);
  }

  async getQuestLog(params: {
    session_id: string;
  }): Promise<{ success: boolean; quest_log: unknown }> {
    return this.call('getQuestLog', params);
  }

  // === Spell System ===

  async getSpell(params: {
    spell_id: string;
  }): Promise<{ success: boolean; spell: unknown }> {
    return this.call('getSpell', params);
  }

  async getSpellsByLevel(params: {
    level: number;
  }): Promise<{ success: boolean; spells: unknown[] }> {
    return this.call('getSpellsByLevel', params);
  }

  async getSpellsBySchool(params: {
    school: string;
  }): Promise<{ success: boolean; spells: unknown[] }> {
    return this.call('getSpellsBySchool', params);
  }

  async getAllSpells(): Promise<{ success: boolean; spells: unknown[] }> {
    return this.call('getAllSpells');
  }

  async searchSpells(params: {
    query: string;
    filters?: Record<string, unknown>;
  }): Promise<{ success: boolean; spells: unknown[] }> {
    return this.call('searchSpells', params);
  }

  // === Spatial Queries ===

  async getObjectsInRange(params: {
    session_id: string;
    min_x: number;
    min_y: number;
    max_x: number;
    max_y: number;
  }): Promise<{ success: boolean; objects: unknown[] }> {
    return this.call('getObjectsInRange', params);
  }

  async getObjectsInRadius(params: {
    session_id: string;
    center_x: number;
    center_y: number;
    radius: number;
  }): Promise<{ success: boolean; objects: unknown[] }> {
    return this.call('getObjectsInRadius', params);
  }

  async getNearestObjects(params: {
    session_id: string;
    center_x: number;
    center_y: number;
    k: number;
  }): Promise<{ success: boolean; objects: unknown[] }> {
    return this.call('getNearestObjects', params);
  }

  // === Procedural Content Generation ===

  async generateContent(params: {
    session_id: string;
    content_type: string;
    options?: Record<string, unknown>;
  }): Promise<{ success: boolean; content: unknown }> {
    return this.call('generateContent', params);
  }

  async regenerateTerrain(params: {
    session_id: string;
    seed?: number;
    biome?: string;
  }): Promise<{ success: boolean; terrain: unknown }> {
    return this.call('regenerateTerrain', params);
  }

  async generateItems(params: {
    session_id: string;
    count: number;
    item_type?: string;
  }): Promise<{ success: boolean; items: unknown[] }> {
    return this.call('generateItems', params);
  }

  async generateLevel(params: {
    session_id: string;
    level_type?: string;
    difficulty?: number;
  }): Promise<{ success: boolean; level: unknown }> {
    return this.call('generateLevel', params);
  }

  async generateQuest(params: {
    session_id: string;
    quest_type?: string;
  }): Promise<{ success: boolean; quest: unknown }> {
    return this.call('generateQuest', params);
  }

  async getPCGStats(params: {
    session_id: string;
  }): Promise<{ success: boolean; stats: unknown }> {
    return this.call('getPCGStats', params);
  }

  async validateContent(params: {
    session_id: string;
    content: unknown;
  }): Promise<{ success: boolean; valid: boolean; errors: string[] }> {
    return this.call('validateContent', params);
  }
}

// WebSocket client for real-time events
export interface EventHandler {
  (event: unknown): void;
}

export class GoldBoxRPGWebSocketClient {
  private ws: WebSocket | null = null;
  private eventHandlers: Map<string, EventHandler[]> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(private wsURL: string) {}

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.wsURL);

        this.ws.onopen = () => {
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onerror = (error) => {
          reject(error);
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };

        this.ws.onclose = () => {
          this.handleClose();
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  on(eventType: string, handler: EventHandler): void {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, []);
    }
    this.eventHandlers.get(eventType)!.push(handler);
  }

  off(eventType: string, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index !== -1) {
        handlers.splice(index, 1);
      }
    }
  }

  private handleMessage(data: string): void {
    try {
      const event = JSON.parse(data);
      const eventType = event.type || 'unknown';
      const handlers = this.eventHandlers.get(eventType);
      if (handlers) {
        handlers.forEach((handler) => handler(event));
      }
      const allHandlers = this.eventHandlers.get('*');
      if (allHandlers) {
        allHandlers.forEach((handler) => handler(event));
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  private handleClose(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        this.connect().catch(() => {
          // Retry failed, will retry again
        });
      }, this.reconnectDelay * this.reconnectAttempts);
    }
  }
}

export default GoldBoxRPGClient;
