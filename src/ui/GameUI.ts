/**
 * UI Manager for GoldBox RPG Engine
 * Handles game interface elements, user input, and display updates
 */

import { BaseComponent } from '../core/BaseComponent';
import type { 
  GameUIElements,
  MessageType
} from '../types/UITypes';
import type { PlayerAttributes, CombatState } from '../types/GameTypes';

/**
 * Main UI manager class that coordinates all game interface elements
 */
export class GameUI extends BaseComponent {
  private elements: GameUIElements | null = null;
  private keyboardHandlers = new Map<string, (event: KeyboardEvent) => void>();

  constructor() {
    super({ name: 'GameUI' });
    this.componentLogger.info('GameUI created');
  }

  /**
   * Initialize the UI by finding DOM elements and setting up event handlers
   */
  protected async onInitialize(): Promise<void> {
    this.componentLogger.info('Initializing Game UI...');
    
    // Find and validate required DOM elements
    this.elements = this.findUIElements();
    this.validateElements();
    
    // Set up event listeners
    this.setupEventListeners();
    this.setupKeyboardControls();
    
    this.componentLogger.info('Game UI initialized successfully');
  }

  /**
   * Clean up UI resources and event listeners
   */
  protected async onCleanup(): Promise<void> {
    this.componentLogger.info('Cleaning up Game UI...');

    // Remove keyboard event handlers
    for (const [, handler] of this.keyboardHandlers.entries()) {
      document.removeEventListener('keydown', handler);
    }
    this.keyboardHandlers.clear();

    this.elements = null;
    
    this.componentLogger.info('Game UI cleanup completed');
  }

  /**
   * Update the UI with current game state
   */
  updateUI(state: {
    player?: {
      name?: string;
      attributes?: PlayerAttributes;
      hp?: { current: number; max: number };
      position?: { x: number; y: number };
    };
    combat?: CombatState;
  }): void {
    if (!this.initialized || !this.elements) {
      this.componentLogger.warn('Cannot update UI - not initialized');
      return;
    }

    try {
      // Update player information
      if (state.player) {
        this.updatePlayerInfo(state.player);
      }

      // Update combat state
      if (state.combat) {
        this.updateCombatInfo(state.combat);
      }

      this.emit('updateUI', { state });
      
    } catch (error) {
      this.componentLogger.error('Failed to update UI:', error);
    }
  }

  /**
   * Add a message to the game log
   */
  logMessage(message: string, type: MessageType = 'info'): void {
    if (!this.initialized || !this.elements?.logContent) {
      this.componentLogger.warn('Cannot log message - UI not initialized');
      return;
    }

    try {
      const timestamp = new Date().toLocaleTimeString();
      const messageElement = document.createElement('div');
      messageElement.className = `log-entry log-${type}`;
      messageElement.innerHTML = `<span class="log-time">[${timestamp}]</span> ${message}`;

      this.elements.logContent.appendChild(messageElement);
      
      // Auto-scroll to bottom
      this.elements.logContent.scrollTop = this.elements.logContent.scrollHeight;

      // Limit log entries to prevent memory issues
      const maxEntries = 100;
      const entries = this.elements.logContent.children;
      while (entries.length > maxEntries) {
        entries[0].remove();
      }

      this.emit('logMessage', { message, type });
      
    } catch (error) {
      this.componentLogger.error('Failed to log message:', error);
    }
  }

  /**
   * Update combat log with new information
   */
  updateCombatLog(data: {
    message?: string;
    type?: MessageType;
    initiative?: Array<{ id: string; name: string; initiative: number; isPlayer: boolean }>;
  }): void {
    if (data.message) {
      this.logMessage(data.message, data.type || 'combat');
    }

    if (data.initiative) {
      this.updateInitiativeOrder(data.initiative);
    }
  }

  /**
   * Update initiative order display
   */
  private updateInitiativeOrder(initiative: readonly { 
    id: string; 
    name: string; 
    initiative: number; 
    isPlayer: boolean 
  }[]): void {
    if (!this.elements?.initiativeList) {
      return;
    }

    try {
      // Clear existing initiative display
      this.elements.initiativeList.innerHTML = '';

      // Sort by initiative (highest first) - create mutable copy
      const sorted = [...initiative].sort((a, b) => b.initiative - a.initiative);

      // Create initiative entries
      sorted.forEach((entry, index) => {
        const entryElement = document.createElement('div');
        entryElement.className = `initiative-entry ${entry.isPlayer ? 'player' : 'npc'}`;
        entryElement.innerHTML = `
          <span class="initiative-order">${index + 1}.</span>
          <span class="initiative-name">${entry.name}</span>
          <span class="initiative-score">${entry.initiative}</span>
        `;
        
        this.elements!.initiativeList!.appendChild(entryElement);
      });
      
    } catch (error) {
      this.componentLogger.error('Failed to update initiative order:', error);
    }
  }

  /**
   * Find and return all required UI elements
   */
  private findUIElements(): GameUIElements {
    const elements: GameUIElements = {
      // Character display elements
      portrait: document.getElementById('character-portrait') as HTMLImageElement,
      name: document.getElementById('character-name') as HTMLElement,
      
      // Stat elements
      stats: {
        str: document.getElementById('stat-str') as HTMLElement,
        dex: document.getElementById('stat-dex') as HTMLElement,
        con: document.getElementById('stat-con') as HTMLElement,
        int: document.getElementById('stat-int') as HTMLElement,
        wis: document.getElementById('stat-wis') as HTMLElement,
        cha: document.getElementById('stat-cha') as HTMLElement,
      },
      
      // Health bar
      hpBar: document.getElementById('hp-bar') as HTMLElement,
      hpText: document.getElementById('hp-text') as HTMLElement,
      
      // Log elements
      logContent: document.getElementById('log-content') as HTMLElement,
      
      // Combat elements
      initiativeList: document.getElementById('initiative-list') as HTMLElement,
      
      // Control buttons
      actionButtons: {
        attack: document.getElementById('btn-attack') as HTMLButtonElement,
        defend: document.getElementById('btn-defend') as HTMLButtonElement,
        cast: document.getElementById('btn-cast') as HTMLButtonElement,
        item: document.getElementById('btn-item') as HTMLButtonElement,
      },
      
      // Direction buttons
      directionButtons: {
        north: document.getElementById('btn-north') as HTMLButtonElement,
        south: document.getElementById('btn-south') as HTMLButtonElement,
        east: document.getElementById('btn-east') as HTMLButtonElement,
        west: document.getElementById('btn-west') as HTMLButtonElement,
        northeast: document.getElementById('btn-northeast') as HTMLButtonElement,
        northwest: document.getElementById('btn-northwest') as HTMLButtonElement,
        southeast: document.getElementById('btn-southeast') as HTMLButtonElement,
        southwest: document.getElementById('btn-southwest') as HTMLButtonElement,
      }
    };

    return elements;
  }

  /**
   * Validate that all required UI elements were found
   */
  private validateElements(): void {
    if (!this.elements) {
      throw new Error('UI elements not initialized');
    }

    const missingElements: string[] = [];

    // Check critical elements
    if (!this.elements.logContent) missingElements.push('log-content');
    if (!this.elements.hpBar) missingElements.push('hp-bar');

    if (missingElements.length > 0) {
      throw new Error(`Missing required UI elements: ${missingElements.join(', ')}`);
    }
  }

  /**
   * Set up event listeners for UI interactions
   */
  private setupEventListeners(): void {
    if (!this.elements) return;

    // Action button handlers
    Object.entries(this.elements.actionButtons).forEach(([action, button]) => {
      if (button) {
        button.addEventListener('click', () => {
          this.emit('action', { action });
          this.componentLogger.debug('Action button clicked', { action });
        });
      }
    });

    // Direction button handlers
    Object.entries(this.elements.directionButtons).forEach(([direction, button]) => {
      if (button) {
        button.addEventListener('click', () => {
          this.emit('move', { direction });
          this.componentLogger.debug('Direction button clicked', { direction });
        });
      }
    });
  }

  /**
   * Set up keyboard controls for game navigation
   */
  private setupKeyboardControls(): void {
    const keyMap: Record<string, string> = {
      'ArrowUp': 'north',
      'ArrowDown': 'south',
      'ArrowLeft': 'west',
      'ArrowRight': 'east',
      'w': 'north',
      's': 'south',
      'a': 'west',
      'd': 'east',
      'q': 'northwest',
      'e': 'northeast',
      'z': 'southwest',
      'c': 'southeast'
    };

    const keyboardHandler = (event: KeyboardEvent) => {
      const key = event.key.toLowerCase();
      const direction = keyMap[event.key] || keyMap[key];
      
      if (direction) {
        event.preventDefault();
        this.emit('move', { direction });
        this.componentLogger.debug('Keyboard movement', { key, direction });
      }
    };

    document.addEventListener('keydown', keyboardHandler);
    this.keyboardHandlers.set('movement', keyboardHandler);
  }

  /**
   * Update player information display
   */
  private updatePlayerInfo(player: {
    name?: string;
    attributes?: PlayerAttributes;
    hp?: { current: number; max: number };
    position?: { x: number; y: number };
  }): void {
    if (!this.elements) return;

    try {
      // Update character name
      if (player.name && this.elements.name) {
        this.elements.name.textContent = player.name;
      }

      // Update attributes
      if (player.attributes && this.elements.stats) {
        const stats = this.elements.stats;
        if (stats.str) stats.str.textContent = player.attributes.strength.toString();
        if (stats.dex) stats.dex.textContent = player.attributes.dexterity.toString();
        if (stats.con) stats.con.textContent = player.attributes.constitution.toString();
        if (stats.int) stats.int.textContent = player.attributes.intelligence.toString();
        if (stats.wis) stats.wis.textContent = player.attributes.wisdom.toString();
        if (stats.cha) stats.cha.textContent = player.attributes.charisma.toString();
      }

      // Update HP bar
      if (player.hp && this.elements.hpBar) {
        const percentage = (player.hp.current / player.hp.max) * 100;
        this.elements.hpBar.style.width = `${percentage}%`;
        
        if (this.elements.hpText) {
          this.elements.hpText.textContent = `${player.hp.current}/${player.hp.max}`;
        }
      }
      
    } catch (error) {
      this.componentLogger.error('Failed to update player info:', error);
    }
  }

  /**
   * Update combat information display
   */
  private updateCombatInfo(combat: CombatState): void {
    if (!this.elements) return;

    try {
      // Update initiative display
      if (combat.initiative.length > 0) {
        this.updateInitiativeOrder(combat.initiative);
      }

      // Log combat state changes
      if (combat.active) {
        this.logMessage(`Combat Round ${combat.round}`, 'combat');
        if (combat.currentTurn) {
          this.logMessage(`${combat.currentTurn}'s turn`, 'info');
        }
      } else {
        this.logMessage('Combat ended', 'combat');
      }
      
    } catch (error) {
      this.componentLogger.error('Failed to update combat info:', error);
    }
  }
}

// Export singleton instance
export const gameUI = new GameUI();
