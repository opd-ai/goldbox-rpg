/**
 * UI Manager for GoldBox RPG Engine
 * Handles game interface elements, user input, and display updates
 */
import { BaseComponent } from '../core/BaseComponent';
import type { MessageType } from '../types/UITypes';
import type { PlayerAttributes, CombatState } from '../types/GameTypes';
/**
 * Main UI manager class that coordinates all game interface elements
 */
export declare class GameUI extends BaseComponent {
    private elements;
    private keyboardHandlers;
    constructor();
    /**
     * Initialize the UI by finding DOM elements and setting up event handlers
     */
    protected onInitialize(): Promise<void>;
    /**
     * Clean up UI resources and event listeners
     */
    protected onCleanup(): Promise<void>;
    /**
     * Update the UI with current game state
     */
    updateUI(state: {
        player?: {
            name?: string;
            attributes?: PlayerAttributes;
            hp?: {
                current: number;
                max: number;
            };
            position?: {
                x: number;
                y: number;
            };
        };
        combat?: CombatState;
    }): void;
    /**
     * Add a message to the game log
     */
    logMessage(message: string, type?: MessageType): void;
    /**
     * Update combat log with new information
     */
    updateCombatLog(data: {
        message?: string;
        type?: MessageType;
        initiative?: Array<{
            id: string;
            name: string;
            initiative: number;
            isPlayer: boolean;
        }>;
    }): void;
    /**
     * Update initiative order display
     */
    private updateInitiativeOrder;
    /**
     * Find and return all required UI elements
     */
    private findUIElements;
    /**
     * Validate that all required UI elements were found
     */
    private validateElements;
    /**
     * Set up event listeners for UI interactions
     */
    private setupEventListeners;
    /**
     * Set up keyboard controls for game navigation
     */
    private setupKeyboardControls;
    /**
     * Update player information display
     */
    private updatePlayerInfo;
    /**
     * Update combat information display
     */
    private updateCombatInfo;
}
export declare const gameUI: GameUI;
//# sourceMappingURL=GameUI.d.ts.map