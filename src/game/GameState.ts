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
    initiative?: Array<{ id: string; name: string; initiative: number }>;
    currentTurn?: string;
  };
  session?: {
    id: string;
    status: 'connected' | 'disconnected' | 'error';
  };
}

export class GameState extends BaseComponent {
  private state: GameStateData = {};

  constructor() {
    super({
      name: 'GameState',
      enableEventEmission: true,
      enableErrorHandling: true,
      autoInitialize: false
    });
  }

  protected async onInitialize(): Promise<void> {
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

  protected async onCleanup(): Promise<void> {
    this.componentLogger.info('Cleaning up GameState');
    this.state = {};
    this.componentLogger.info('GameState cleanup complete');
  }

  // State getters
  getCharacter(): Character | undefined {
    return this.state.character;
  }

  getWorld(): any {
    return this.state.world;
  }

  getUIState(): GameUIState | undefined {
    return this.state.ui;
  }

  getCombatState(): any {
    return this.state.combat;
  }

  getSessionState(): any {
    return this.state.session;
  }

  getFullState(): GameStateData {
    return { ...this.state };
  }

  // State setters
  setCharacter(character: Character): void {
    this.state.character = character;
    this.emit('characterUpdated', character);
  }

  setWorld(world: any): void {
    this.state.world = world;
    this.emit('worldUpdated', world);
  }

  setUIState(uiState: Partial<GameUIState>): void {
    if (this.state.ui) {
      this.state.ui = {
        mode: uiState.mode ?? this.state.ui.mode,
        selectedTarget: uiState.selectedTarget ?? this.state.ui.selectedTarget,
        inventoryOpen: uiState.inventoryOpen ?? this.state.ui.inventoryOpen,
        spellbookOpen: uiState.spellbookOpen ?? this.state.ui.spellbookOpen,
        characterSheetOpen: uiState.characterSheetOpen ?? this.state.ui.characterSheetOpen
      };
    } else {
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

  setCombatState(combatState: any): void {
    this.state.combat = { ...this.state.combat, ...combatState };
    this.emit('combatStateUpdated', this.state.combat);
  }

  setSessionState(sessionState: any): void {
    this.state.session = { ...this.state.session, ...sessionState };
    this.emit('sessionStateUpdated', this.state.session);
  }

  // Utility methods
  isInCombat(): boolean {
    return this.state.combat?.inCombat || false;
  }

  isConnected(): boolean {
    return this.state.session?.status === 'connected';
  }

  updatePlayerPosition(position: Position): void {
    if (this.state.character) {
      // Create a new character object with updated position to respect readonly
      this.state.character = {
        ...this.state.character,
        position
      };
      this.emit('playerPositionUpdated', position);
    }
  }

  updateInitiativeOrder(initiative: Array<{ id: string; name: string; initiative: number }>): void {
    if (this.state.combat) {
      this.state.combat.initiative = initiative;
      this.emit('initiativeUpdated', initiative);
    }
  }
}

// Create and export a singleton instance
export const gameState = new GameState();
