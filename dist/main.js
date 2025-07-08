/**
 * Main application entry point for GoldBox RPG Engine
 * Initializes all systems and provides the global game interface
 */
import { logger, GlobalErrorHandler, ComponentManager, rpcClient, gameUI, gameState } from './index';
// Setup global error handling
GlobalErrorHandler.setupGlobalHandlers();
// Create component manager
const componentManager = new ComponentManager();
/**
 * Main application class that coordinates all game systems
 */
class GoldBoxRPGApp {
    constructor() {
        this.logger = logger.createChildLogger('GoldBoxRPGApp');
        this.initialized = false;
    }
    /**
     * Initialize the application
     */
    async initialize() {
        if (this.initialized) {
            this.logger.warn('Application already initialized');
            return;
        }
        try {
            this.logger.info('Initializing GoldBox RPG Engine...');
            // Register core components
            componentManager.register(gameState);
            componentManager.register(gameUI);
            // Initialize all components
            await componentManager.initializeAll();
            // Connect RPC client
            await rpcClient.connect();
            this.initialized = true;
            this.logger.info('GoldBox RPG Engine initialized successfully');
            // Emit ready event for legacy JavaScript integration
            if (typeof window !== 'undefined') {
                window.dispatchEvent(new CustomEvent('goldbox-ready', {
                    detail: { app: this, rpcClient, gameUI, gameState }
                }));
            }
        }
        catch (error) {
            this.logger.error('Failed to initialize application:', error);
            throw error;
        }
    }
    /**
     * Cleanup and shutdown the application
     */
    async cleanup() {
        if (!this.initialized) {
            return;
        }
        try {
            this.logger.info('Shutting down GoldBox RPG Engine...');
            await componentManager.cleanupAll();
            this.initialized = false;
            this.logger.info('GoldBox RPG Engine shut down successfully');
        }
        catch (error) {
            this.logger.error('Error during application cleanup:', error);
        }
    }
    /**
     * Get the component manager for accessing game systems
     */
    getComponentManager() {
        return componentManager;
    }
    /**
     * Check if application is initialized
     */
    isInitialized() {
        return this.initialized;
    }
}
// Create global application instance
const app = new GoldBoxRPGApp();
// Auto-initialize when DOM is ready
if (typeof document !== 'undefined') {
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            app.initialize().catch(error => {
                console.error('Failed to auto-initialize application:', error);
            });
        });
    }
    else {
        // DOM already loaded
        app.initialize().catch(error => {
            console.error('Failed to auto-initialize application:', error);
        });
    }
}
// Cleanup on page unload
if (typeof window !== 'undefined') {
    window.addEventListener('beforeunload', () => {
        app.cleanup().catch(error => {
            console.error('Error during cleanup:', error);
        });
    });
}
// Export for global access
export { app as GoldBoxRPG };
// For legacy JavaScript compatibility
if (typeof window !== 'undefined') {
    window.GoldBoxRPG = {
        app,
        logger,
        componentManager,
        // Expose other utilities for migration period
        ErrorHandler: GlobalErrorHandler,
    };
}
//# sourceMappingURL=main.js.map