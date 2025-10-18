/**
 * Game rendering system with TypeScript support
 * Handles canvas-based rendering of terrain, objects, and effects
 */
import { BaseComponent } from '../core/BaseComponent';
import type { GameState as IGameState } from '../types/GameTypes';
import type { RenderOptions } from '../types/UITypes';
export interface GameRendererEvents {
    spritesLoaded: {
        count: number;
    };
    spriteLoadError: {
        sprite: string;
        error: Error;
    };
    renderComplete: {
        frameTime: number;
    };
    error: Error;
}
export interface Camera {
    x: number;
    y: number;
    zoom: number;
}
export interface SpriteInfo {
    readonly image: HTMLImageElement;
    readonly width: number;
    readonly height: number;
    readonly loaded: boolean;
}
export declare class GameRenderer extends BaseComponent {
    private readonly canvasLayers;
    private readonly contexts;
    private readonly tileSize;
    private readonly sprites;
    private readonly animations;
    private readonly camera;
    private readonly boundHandleResize;
    private lastFrameTime;
    private frameCount;
    constructor();
    /**
     * Initialize the renderer
     */
    protected onInitialize(): Promise<void>;
    /**
     * Clean up renderer resources
     */
    protected onCleanup(): Promise<void>;
    /**
     * Load sprite images asynchronously
     */
    loadSprites(): Promise<void>;
    /**
     * Load a single sprite image
     */
    private loadSpriteImage;
    /**
     * Create fallback sprites when image loading fails
     */
    useFallbackSprites(): void;
    /**
     * Handle window resize events
     */
    private handleResize;
    /**
     * Clear all canvas layers
     */
    clearLayers(): void;
    /**
     * Main render method
     */
    render(gameState: IGameState, options?: RenderOptions): void;
    /**
     * Render terrain tiles
     */
    private renderTerrain;
    /**
     * Render game objects
     */
    private renderObjects;
    /**
     * Render visual effects (placeholder for future implementation)
     * @unused - Reserved for future effects system implementation
     */
    private renderEffects;
    /**
     * Draw a sprite on the specified context
     */
    private drawSprite;
    /**
     * Check if a position is on screen
     */
    private isOnScreen;
    /**
     * Check if a map position is valid
     */
    private isValidMapPosition;
    /**
     * Update camera position
     */
    updateCamera(x: number, y: number, zoom?: number): void;
    /**
     * Get current camera state
     */
    getCamera(): Readonly<Camera>;
    /**
     * Update highlighted cells (for combat targeting)
     */
    updateHighlights(cells: string[]): void;
    /**
     * Get rendering performance stats
     */
    getPerformanceStats(): {
        frameTime: number;
        frameCount: number;
        fps: number;
    };
    /**
     * Check WebGL support
     */
    private checkWebGLSupport;
    /**
     * Get sprite coordinates for a tile type
     */
    private getTileSprite;
}
//# sourceMappingURL=GameRenderer.d.ts.map