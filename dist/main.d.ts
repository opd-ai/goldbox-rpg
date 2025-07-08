/**
 * Main application entry point for GoldBox RPG Engine
 * Initializes all systems and provides the global game interface
 */
import { ComponentManager } from './index';
/**
 * Main application class that coordinates all game systems
 */
declare class GoldBoxRPGApp {
    private readonly logger;
    private initialized;
    /**
     * Initialize the application
     */
    initialize(): Promise<void>;
    /**
     * Cleanup and shutdown the application
     */
    cleanup(): Promise<void>;
    /**
     * Get the component manager for accessing game systems
     */
    getComponentManager(): ComponentManager;
    /**
     * Check if application is initialized
     */
    isInitialized(): boolean;
}
declare const app: GoldBoxRPGApp;
export { app as GoldBoxRPG };
//# sourceMappingURL=main.d.ts.map