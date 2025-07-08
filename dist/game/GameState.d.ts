import { BaseComponent } from '../core/BaseComponent';
import { Character, Position } from '../types/GameTypes';
import { GameUIState } from '../types/UITypes';
export interface GameStateData {
    character?: Character;
    world?: {
        currentMap?: string;
        mapData?: any;
    };
    ui?: GameUIState;
    combat?: {
        inCombat: boolean;
        initiative?: Array<{
            id: string;
            name: string;
            initiative: number;
        }>;
        currentTurn?: string;
    };
    session?: {
        id: string;
        status: 'connected' | 'disconnected' | 'error';
    };
}
export declare class GameState extends BaseComponent {
    private state;
    constructor();
    protected onInitialize(): Promise<void>;
    protected onCleanup(): Promise<void>;
    getCharacter(): Character | undefined;
    getWorld(): any;
    getUIState(): GameUIState | undefined;
    getCombatState(): any;
    getSessionState(): any;
    getFullState(): GameStateData;
    setCharacter(character: Character): void;
    setWorld(world: any): void;
    setUIState(uiState: Partial<GameUIState>): void;
    setCombatState(combatState: any): void;
    setSessionState(sessionState: any): void;
    isInCombat(): boolean;
    isConnected(): boolean;
    updatePlayerPosition(position: Position): void;
    updateInitiativeOrder(initiative: Array<{
        id: string;
        name: string;
        initiative: number;
    }>): void;
}
export declare const gameState: GameState;
//# sourceMappingURL=GameState.d.ts.map