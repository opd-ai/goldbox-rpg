import { BaseComponent } from '../core/BaseComponent';
export class GameState extends BaseComponent {
    constructor() {
        super({
            name: 'GameState',
            enableEventEmission: true,
            enableErrorHandling: true,
            autoInitialize: false
        });
        this.state = {};
    }
    async onInitialize() {
        this.componentLogger.info('Initializing GameState');
        this.state = {
            world: {},
            ui: {
                mode: 'normal',
                selectedTarget: null,
                inventoryOpen: false,
                spellbookOpen: false,
                characterSheetOpen: false
            },
            combat: {
                inCombat: false
            },
            session: {
                id: '',
                status: 'disconnected'
            }
        };
        this.emit('gameStateInitialized', this.state);
        this.componentLogger.info('GameState initialized successfully');
    }
    async onCleanup() {
        this.componentLogger.info('Cleaning up GameState');
        this.state = {};
        this.componentLogger.info('GameState cleanup complete');
    }
    // State getters
    getCharacter() {
        return this.state.character;
    }
    getWorld() {
        return this.state.world;
    }
    getUIState() {
        return this.state.ui;
    }
    getCombatState() {
        return this.state.combat;
    }
    getSessionState() {
        return this.state.session;
    }
    getFullState() {
        return { ...this.state };
    }
    // State setters
    setCharacter(character) {
        this.state.character = character;
        this.emit('characterUpdated', character);
    }
    setWorld(world) {
        this.state.world = world;
        this.emit('worldUpdated', world);
    }
    setUIState(uiState) {
        if (this.state.ui) {
            this.state.ui = {
                mode: uiState.mode ?? this.state.ui.mode,
                selectedTarget: uiState.selectedTarget ?? this.state.ui.selectedTarget,
                inventoryOpen: uiState.inventoryOpen ?? this.state.ui.inventoryOpen,
                spellbookOpen: uiState.spellbookOpen ?? this.state.ui.spellbookOpen,
                characterSheetOpen: uiState.characterSheetOpen ?? this.state.ui.characterSheetOpen
            };
        }
        else {
            this.state.ui = {
                mode: uiState.mode ?? 'normal',
                selectedTarget: uiState.selectedTarget ?? null,
                inventoryOpen: uiState.inventoryOpen ?? false,
                spellbookOpen: uiState.spellbookOpen ?? false,
                characterSheetOpen: uiState.characterSheetOpen ?? false
            };
        }
        this.emit('uiStateUpdated', this.state.ui);
    }
    setCombatState(combatState) {
        this.state.combat = { ...this.state.combat, ...combatState };
        this.emit('combatStateUpdated', this.state.combat);
    }
    setSessionState(sessionState) {
        this.state.session = { ...this.state.session, ...sessionState };
        this.emit('sessionStateUpdated', this.state.session);
    }
    // Utility methods
    isInCombat() {
        return this.state.combat?.inCombat || false;
    }
    isConnected() {
        return this.state.session?.status === 'connected';
    }
    updatePlayerPosition(position) {
        if (this.state.character) {
            // Create a new character object with updated position to respect readonly
            this.state.character = {
                ...this.state.character,
                position
            };
            this.emit('playerPositionUpdated', position);
        }
    }
    updateInitiativeOrder(initiative) {
        if (this.state.combat) {
            this.state.combat.initiative = initiative;
            this.emit('initiativeUpdated', initiative);
        }
    }
}
// Create and export a singleton instance
export const gameState = new GameState();
//# sourceMappingURL=GameState.js.map