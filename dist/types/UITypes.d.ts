/**
 * UI and DOM-related type definitions for GoldBox RPG Engine
 */
export type EventCallback<T = unknown> = (data: T) => void;
export type EventUnsubscriber = () => void;
export interface EventEmitterInterface {
    on<T = unknown>(event: string, callback: EventCallback<T>): EventUnsubscriber;
    emit<T = unknown>(event: string, data?: T): void;
    off(event: string, callback: EventCallback): boolean;
    removeAllListeners(event?: string): void;
    clear(): void;
    listenerCount(event: string): number;
    eventNames(): string[];
}
export interface UIComponent {
    readonly name: string;
    readonly element: HTMLElement;
    initialize(): Promise<void> | void;
    update(data: unknown): void;
    cleanup(): void;
}
export interface UIElements {
    readonly [elementId: string]: HTMLElement | null;
}
export interface GameUIElements {
    readonly portrait: HTMLImageElement;
    readonly name: HTMLElement;
    readonly stats: {
        readonly str: HTMLElement;
        readonly dex: HTMLElement;
        readonly con: HTMLElement;
        readonly int: HTMLElement;
        readonly wis: HTMLElement;
        readonly cha: HTMLElement;
    };
    readonly hpBar: HTMLElement;
    readonly hpText: HTMLElement;
    readonly logContent: HTMLElement;
    readonly initiativeList: HTMLElement;
    readonly actionButtons: {
        readonly attack: HTMLButtonElement;
        readonly defend: HTMLButtonElement;
        readonly cast: HTMLButtonElement;
        readonly item: HTMLButtonElement;
    };
    readonly directionButtons: {
        readonly north: HTMLButtonElement;
        readonly south: HTMLButtonElement;
        readonly east: HTMLButtonElement;
        readonly west: HTMLButtonElement;
        readonly northeast: HTMLButtonElement;
        readonly northwest: HTMLButtonElement;
        readonly southeast: HTMLButtonElement;
        readonly southwest: HTMLButtonElement;
    };
}
export type KeyboardDirection = 'ArrowUp' | 'ArrowDown' | 'ArrowLeft' | 'ArrowRight' | 'KeyW' | 'KeyA' | 'KeyS' | 'KeyD' | 'Numpad8' | 'Numpad2' | 'Numpad4' | 'Numpad6' | 'Numpad7' | 'Numpad9' | 'Numpad1' | 'Numpad3';
export interface KeyboardEventMap {
    readonly [key: string]: () => void;
}
export interface CanvasLayers {
    readonly terrain: HTMLCanvasElement;
    readonly objects: HTMLCanvasElement;
    readonly effects: HTMLCanvasElement;
}
export interface CanvasContexts {
    readonly terrain: CanvasRenderingContext2D;
    readonly objects: CanvasRenderingContext2D;
    readonly effects: CanvasRenderingContext2D;
}
export interface SpriteMap {
    readonly [spriteName: string]: HTMLImageElement;
}
export interface RenderOptions {
    readonly clearLayers?: boolean;
    readonly updateOnly?: readonly string[];
    readonly viewport?: {
        readonly x: number;
        readonly y: number;
        readonly width: number;
        readonly height: number;
    };
}
export interface CombatUIState {
    readonly active: boolean;
    readonly currentTurn: string | null;
    readonly selectedAction: string | null;
    readonly selectedTarget: string | null;
    readonly highlightedCells: ReadonlySet<string>;
}
export interface ActionButton {
    readonly id: string;
    readonly label: string;
    readonly enabled: boolean;
    readonly action: () => void;
}
export type MessageType = 'info' | 'warning' | 'error' | 'combat' | 'system';
export interface GameMessage {
    readonly id: string;
    readonly type: MessageType;
    readonly content: string;
    readonly timestamp: number;
    readonly metadata?: Readonly<Record<string, unknown>>;
}
export interface LogEntry {
    readonly level: 'debug' | 'info' | 'warn' | 'error';
    readonly message: string;
    readonly timestamp: number;
    readonly component?: string;
    readonly metadata?: Readonly<Record<string, unknown>>;
}
export interface UIEventMap {
    'move': {
        direction: string;
    };
    'attack': {
        targetId: string;
    };
    'castSpell': {
        spellId: string;
        targetId?: string;
    };
    'selectAction': {
        actionId: string;
    };
    'selectTarget': {
        targetId: string;
    };
    'logMessage': GameMessage;
    'updateUI': unknown;
    'resize': {
        width: number;
        height: number;
    };
}
export interface ErrorDisplayOptions {
    readonly title?: string;
    readonly message: string;
    readonly type: 'error' | 'warning' | 'info';
    readonly duration?: number;
    readonly dismissible?: boolean;
}
export interface Viewport {
    readonly x: number;
    readonly y: number;
    readonly width: number;
    readonly height: number;
    readonly scale: number;
}
export interface CameraTarget {
    readonly x: number;
    readonly y: number;
    readonly smooth?: boolean;
    readonly duration?: number;
}
export interface AnimationFrame {
    readonly timestamp: number;
    readonly deltaTime: number;
}
export interface TransitionOptions {
    readonly duration: number;
    readonly easing?: 'linear' | 'ease-in' | 'ease-out' | 'ease-in-out';
    readonly onComplete?: () => void;
}
export interface TouchPoint {
    readonly id: number;
    readonly x: number;
    readonly y: number;
    readonly pressure?: number;
}
export interface GestureEvent {
    readonly type: 'tap' | 'drag' | 'pinch' | 'swipe';
    readonly touches: readonly TouchPoint[];
    readonly deltaX?: number;
    readonly deltaY?: number;
    readonly scale?: number;
}
export interface AccessibilityOptions {
    readonly enableScreenReader: boolean;
    readonly enableKeyboardNavigation: boolean;
    readonly enableHighContrast: boolean;
    readonly fontSize: 'small' | 'medium' | 'large';
}
export interface ThemeColors {
    readonly primary: string;
    readonly secondary: string;
    readonly background: string;
    readonly text: string;
    readonly error: string;
    readonly warning: string;
    readonly success: string;
}
export interface UITheme {
    readonly name: string;
    readonly colors: ThemeColors;
    readonly fonts: {
        readonly primary: string;
        readonly monospace: string;
    };
    readonly spacing: {
        readonly small: number;
        readonly medium: number;
        readonly large: number;
    };
}
export interface GameUIState {
    readonly mode: 'normal' | 'combat' | 'inventory' | 'spellcasting';
    readonly selectedTarget: string | null;
    readonly inventoryOpen: boolean;
    readonly spellbookOpen: boolean;
    readonly characterSheetOpen: boolean;
}
//# sourceMappingURL=UITypes.d.ts.map