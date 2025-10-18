/**
 * JSON-RPC 2.0 type definitions for GoldBox RPG Engine
 * Provides strict typing for RPC communication with the Go backend
 */
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
export declare const enum RPCErrorCode {
    PARSE_ERROR = -32700,
    INVALID_REQUEST = -32600,
    METHOD_NOT_FOUND = -32601,
    INVALID_PARAMS = -32602,
    INTERNAL_ERROR = -32603,
    SERVER_ERROR_MIN = -32099,
    SERVER_ERROR_MAX = -32000
}
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
    readonly weapon_id: string;
}
export interface SpellParams {
    readonly session_id: string;
    readonly spell_id: string;
    readonly target_id?: string;
    readonly position?: {
        readonly x: number;
        readonly y: number;
        readonly level: number;
        readonly facing: string;
    };
}
export interface StartCombatParams {
    readonly session_id: string;
    readonly participant_ids: readonly string[];
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
export interface UseItemParams {
    readonly session_id: string;
    readonly item_id: string;
    readonly target_id: string;
}
export interface EndTurnParams {
    readonly session_id: string;
}
export interface EquipItemParams {
    readonly session_id: string;
    readonly item_id: string;
    readonly slot: string;
}
export interface UnequipItemParams {
    readonly session_id: string;
    readonly slot: string;
}
export interface GetEquipmentParams {
    readonly session_id: string;
}
export declare const enum CharacterClass {
    Fighter = "fighter",
    Mage = "mage",
    Cleric = "cleric",
    Thief = "thief",
    Ranger = "ranger",
    Paladin = "paladin"
}
export declare const enum AttributeMethod {
    Roll = "roll",
    PointBuy = "pointbuy",
    Standard = "standard",
    Custom = "custom"
}
export declare const enum EffectType {
    DamageOverTime = "damage_over_time",
    HealOverTime = "heal_over_time",
    Poison = "poison",
    Burning = "burning",
    Bleeding = "bleeding",
    Stun = "stun",
    Root = "root",
    StatBoost = "stat_boost",
    StatPenalty = "stat_penalty"
}
export interface Duration {
    readonly rounds?: number;
    readonly turns?: number;
    readonly real_time?: number;
}
export interface CreateCharacterParams {
    readonly name: string;
    readonly class: CharacterClass;
    readonly attribute_method: AttributeMethod;
    readonly custom_attributes?: Record<string, number>;
    readonly starting_equipment: boolean;
    readonly starting_gold?: number;
}
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
    readonly success: boolean;
    readonly objects?: readonly unknown[];
    readonly count: number;
}
export interface CreateCharacterResult {
    readonly success: boolean;
    readonly character?: unknown;
    readonly player?: unknown;
    readonly session_id?: string;
    readonly errors?: readonly string[];
    readonly warnings?: readonly string[];
    readonly creation_time?: string;
}
export interface UseItemResult {
    readonly success: boolean;
    readonly effect: string;
}
export interface EndTurnResult {
    readonly success: boolean;
    readonly next_turn: string;
}
export interface EquipItemResult {
    readonly success: boolean;
    readonly equipped_item?: unknown;
    readonly previous_item?: unknown;
}
export interface UnequipItemResult {
    readonly success: boolean;
    readonly unequipped_item?: unknown;
}
export interface GetEquipmentResult {
    readonly success: boolean;
    readonly equipment?: Record<string, unknown>;
    readonly total_weight?: number;
}
export interface ApplyEffectParams {
    readonly session_id: string;
    readonly effect_type: EffectType;
    readonly target_id: string;
    readonly magnitude: number;
    readonly duration: Duration;
}
export interface ApplyEffectResult {
    readonly success: boolean;
    readonly effect_id: string;
}
export interface GetObjectsInRangeParams {
    readonly session_id: string;
    readonly min_x: number;
    readonly min_y: number;
    readonly max_x: number;
    readonly max_y: number;
}
export interface GetObjectsInRadiusParams {
    readonly session_id: string;
    readonly center_x: number;
    readonly center_y: number;
    readonly radius: number;
}
export interface GetNearestObjectsParams {
    readonly session_id: string;
    readonly center_x: number;
    readonly center_y: number;
    readonly k: number;
}
export interface GetSpellParams {
    readonly spell_id: string;
}
export interface GetSpellsByLevelParams {
    readonly level: number;
}
export interface GetSpellsBySchoolParams {
    readonly school: string;
}
export interface GetAllSpellsParams {
}
export interface SearchSpellsParams {
    readonly query: string;
}
export interface SpellDetailResult {
    readonly success: boolean;
    readonly spell?: unknown;
}
export interface SpellListResult {
    readonly success: boolean;
    readonly spells?: readonly unknown[];
    readonly count: number;
}
export declare const enum QuestStatus {
    NotStarted = 0,
    Active = 1,
    Completed = 2,
    Failed = 3
}
export interface QuestObjective {
    readonly description: string;
    readonly progress: number;
    readonly required: number;
    readonly completed: boolean;
}
export interface QuestReward {
    readonly type: string;
    readonly value: number;
    readonly item_id?: string;
}
export interface QuestProgress {
    readonly quest_id: string;
    readonly objectives_complete: number;
    readonly time_spent: number;
    readonly attempts: number;
}
export interface Quest {
    readonly id: string;
    readonly title: string;
    readonly description: string;
    readonly status: QuestStatus;
    readonly objectives: QuestObjective[];
    readonly rewards: QuestReward[];
}
export interface StartQuestParams {
    readonly session_id: string;
    readonly quest: Quest;
}
export interface CompleteQuestParams {
    readonly session_id: string;
    readonly quest_id: string;
}
export interface UpdateObjectiveParams {
    readonly session_id: string;
    readonly quest_id: string;
    readonly objective_index: number;
    readonly progress: number;
}
export interface FailQuestParams {
    readonly session_id: string;
    readonly quest_id: string;
}
export interface GetQuestParams {
    readonly session_id: string;
    readonly quest_id: string;
}
export interface GetActiveQuestsParams {
    readonly session_id: string;
}
export interface GetCompletedQuestsParams {
    readonly session_id: string;
}
export interface GetQuestLogParams {
    readonly session_id: string;
}
export interface StartQuestResult {
    readonly success: boolean;
    readonly quest_id: string;
    readonly error?: string;
}
export interface CompleteQuestResult {
    readonly success: boolean;
    readonly quest_id: string;
    readonly error?: string;
}
export interface UpdateObjectiveResult {
    readonly success: boolean;
    readonly error?: string;
}
export interface FailQuestResult {
    readonly success: boolean;
    readonly quest_id: string;
    readonly error?: string;
}
export interface GetQuestResult {
    readonly success: boolean;
    readonly quest?: Quest;
    readonly error?: string;
}
export interface GetActiveQuestsResult {
    readonly success: boolean;
    readonly active_quests: Quest[];
    readonly count: number;
    readonly error?: string;
}
export interface GetCompletedQuestsResult {
    readonly success: boolean;
    readonly completed_quests: Quest[];
    readonly count: number;
    readonly error?: string;
}
export interface GetQuestLogResult {
    readonly success: boolean;
    readonly quest_log: Quest[];
    readonly count: number;
    readonly error?: string;
}
export type RPCMethodName = 'joinGame' | 'move' | 'attack' | 'castSpell' | 'startCombat' | 'getGameState' | 'getObjectsInRange' | 'getObjectsInRadius' | 'getNearestObjects' | 'leaveGame' | 'createCharacter' | 'useItem' | 'endTurn' | 'applyEffect' | 'equipItem' | 'unequipItem' | 'getEquipment' | 'getSpell' | 'getSpellsByLevel' | 'getSpellsBySchool' | 'getAllSpells' | 'searchSpells' | 'startQuest' | 'completeQuest' | 'updateObjective' | 'failQuest' | 'getQuest' | 'getActiveQuests' | 'getCompletedQuests' | 'getQuestLog';
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
    'getObjectsInRange': {
        params: GetObjectsInRangeParams;
        result: SpatialQueryResult;
    };
    'getObjectsInRadius': {
        params: GetObjectsInRadiusParams;
        result: SpatialQueryResult;
    };
    'getNearestObjects': {
        params: GetNearestObjectsParams;
        result: SpatialQueryResult;
    };
    'leaveGame': {
        params: {
            readonly session_id: string;
        };
        result: {
            readonly success: boolean;
        };
    };
    'createCharacter': {
        params: CreateCharacterParams;
        result: CreateCharacterResult;
    };
    'useItem': {
        params: UseItemParams;
        result: UseItemResult;
    };
    'endTurn': {
        params: EndTurnParams;
        result: EndTurnResult;
    };
    'applyEffect': {
        params: ApplyEffectParams;
        result: ApplyEffectResult;
    };
    'equipItem': {
        params: EquipItemParams;
        result: EquipItemResult;
    };
    'unequipItem': {
        params: UnequipItemParams;
        result: UnequipItemResult;
    };
    'getEquipment': {
        params: GetEquipmentParams;
        result: GetEquipmentResult;
    };
    'getSpell': {
        params: GetSpellParams;
        result: SpellDetailResult;
    };
    'getSpellsByLevel': {
        params: GetSpellsByLevelParams;
        result: SpellListResult;
    };
    'getSpellsBySchool': {
        params: GetSpellsBySchoolParams;
        result: SpellListResult;
    };
    'getAllSpells': {
        params: GetAllSpellsParams;
        result: SpellListResult;
    };
    'searchSpells': {
        params: SearchSpellsParams;
        result: SpellListResult;
    };
    'startQuest': {
        params: StartQuestParams;
        result: StartQuestResult;
    };
    'completeQuest': {
        params: CompleteQuestParams;
        result: CompleteQuestResult;
    };
    'updateObjective': {
        params: UpdateObjectiveParams;
        result: UpdateObjectiveResult;
    };
    'failQuest': {
        params: FailQuestParams;
        result: FailQuestResult;
    };
    'getQuest': {
        params: GetQuestParams;
        result: GetQuestResult;
    };
    'getActiveQuests': {
        params: GetActiveQuestsParams;
        result: GetActiveQuestsResult;
    };
    'getCompletedQuests': {
        params: GetCompletedQuestsParams;
        result: GetCompletedQuestsResult;
    };
    'getQuestLog': {
        params: GetQuestLogParams;
        result: GetQuestLogResult;
    };
}
export type TypedRPCCall = <T extends RPCMethodName>(method: T, params: RPCMethodMap[T]['params']) => Promise<RPCMethodMap[T]['result']>;
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
export interface RPCClientConfig {
    readonly baseUrl: string;
    readonly timeout: number;
    readonly maxReconnectAttempts: number;
    readonly reconnectBackoffBase: number;
    readonly reconnectBackoffMax: number;
    readonly enableLogging: boolean;
    readonly validateOrigin: boolean;
}
export interface PendingRequest {
    readonly id: string | number;
    readonly method: string;
    readonly timestamp: number;
    readonly resolve: (value: unknown) => void;
    readonly reject: (reason: unknown) => void;
}
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
export interface SessionInfo {
    readonly sessionId: string;
    readonly expiresAt: Date;
    readonly isValid: boolean;
}
export type RPCMethod = keyof RPCMethodMap;
//# sourceMappingURL=RPCTypes.d.ts.map