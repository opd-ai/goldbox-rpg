/**
 * Integration test for GoldBox RPG Engine TypeScript migration
 * Tests basic initialization and integration of all core components
 */

// Import components from the compiled output
import { 
  logger,
  ComponentManager,
  rpcClient,
  gameUI,
  gameState
} from '../dist/index.js';

/**
 * Basic integration test
 */
async function runIntegrationTest() {
  const testLogger = logger.createChildLogger('IntegrationTest');
  
  try {
    testLogger.info('Starting integration test...');
    
    // Test 1: Component Manager
    const componentManager = new ComponentManager();
    testLogger.info('✓ ComponentManager created successfully');
    
    // Test 2: Register components
    componentManager.register(gameState);
    componentManager.register(gameUI);
    testLogger.info('✓ Components registered successfully');
    
    // Test 3: Initialize components
    await componentManager.initializeAll();
    testLogger.info('✓ All components initialized successfully');
    
    // Test 4: Check component states
    if (gameState.initialized) {
      testLogger.info('✓ GameState is initialized');
    } else {
      throw new Error('GameState failed to initialize');
    }
    
    if (gameUI.initialized) {
      testLogger.info('✓ GameUI is initialized');
    } else {
      throw new Error('GameUI failed to initialize');
    }
    
    // Test 5: Test event emission
    let eventReceived = false;
    gameState.on('test-event', () => {
      eventReceived = true;
    });
    
    gameState.emit('test-event', { test: true });
    
    if (eventReceived) {
      testLogger.info('✓ Event system working correctly');
    } else {
      throw new Error('Event system failed');
    }
    
    // Test 6: Test state management
    gameState.setUIState({ mode: 'combat', selectedTarget: null, inventoryOpen: true, spellbookOpen: false, characterSheetOpen: false });
    const uiState = gameState.getUIState();
    
    if (uiState && uiState.mode === 'combat' && uiState.inventoryOpen === true) {
      testLogger.info('✓ State management working correctly');
    } else {
      throw new Error('State management failed');
    }
    
    // Test 7: Cleanup
    await componentManager.cleanupAll();
    testLogger.info('✓ All components cleaned up successfully');
    
    testLogger.info('🎉 All integration tests passed!');
    
  } catch (error) {
    testLogger.error('❌ Integration test failed:', error);
    throw error;
  }
}

// Run the test if this file is executed directly
if (typeof window === 'undefined') {
  // Node.js environment
  runIntegrationTest().catch(console.error);
} else {
  // Browser environment - expose test function
  window.runIntegrationTest = runIntegrationTest;
  console.log('Integration test loaded. Run window.runIntegrationTest() to execute.');
}

export { runIntegrationTest };
