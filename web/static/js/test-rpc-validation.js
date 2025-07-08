/**
 * Test suite for JSON-RPC response validation
 */

// Mock minimal EventEmitter for testing
class TestEventEmitter {
  constructor() {
    this.listeners = {};
  }
  
  emit(event, data) {
    // Minimal implementation for testing
  }
}

// Mock RPCClient with validation methods for testing
class TestRPCClient extends TestEventEmitter {
  validateJSONRPCResponse(response) {
    if (!response || typeof response !== 'object') {
      return false;
    }
    
    // Check required jsonrpc field
    if (response.jsonrpc !== "2.0") {
      return false;
    }
    
    // Must have either result or error, but not both
    const hasResult = 'result' in response;
    const hasError = 'error' in response;
    
    if ((!hasResult && !hasError) || (hasResult && hasError)) {
      return false;
    }
    
    // Must have id field (can be null for notifications)
    if (!('id' in response)) {
      return false;
    }
    
    // Validate error format if present
    if (hasError) {
      if (!response.error || typeof response.error !== 'object') {
        return false;
      }
      if (typeof response.error.code !== 'number' || 
          typeof response.error.message !== 'string') {
        return false;
      }
    }
    
    return true;
  }

  validateMethodParameters(method, params) {
    if (!method || typeof method !== 'string') {
      throw new Error('Invalid method name');
    }
    
    if (params && typeof params !== 'object') {
      throw new Error('Parameters must be an object');
    }
    
    // Method-specific validation
    switch (method) {
      case 'move':
        if (!params.direction || !['up', 'down', 'left', 'right', 'n', 's', 'e', 'w', 'ne', 'nw', 'se', 'sw'].includes(params.direction)) {
          throw new Error('Invalid movement direction. Must be one of: up, down, left, right, n, s, e, w, ne, nw, se, sw');
        }
        break;
        
      case 'attack':
        if (!params.target_id && !params.targetId) {
          throw new Error('Attack requires target_id');
        }
        if (!params.weapon_id && !params.weaponId) {
          throw new Error('Attack requires weapon_id');
        }
        // Validate IDs are not empty strings or invalid values
        const targetId = params.target_id || params.targetId;
        const weaponId = params.weapon_id || params.weaponId;
        if (typeof targetId !== 'string' && typeof targetId !== 'number') {
          throw new Error('target_id must be a string or number');
        }
        if (typeof weaponId !== 'string' && typeof weaponId !== 'number') {
          throw new Error('weapon_id must be a string or number');
        }
        break;
        
      case 'castSpell':
        if (!params.spell_id && !params.spellId) {
          throw new Error('Spell casting requires spell_id');
        }
        if (!params.target_id && !params.targetId && !params.position) {
          throw new Error('Spell casting requires either target_id or position');
        }
        // Validate spell ID
        const spellId = params.spell_id || params.spellId;
        if (typeof spellId !== 'string' && typeof spellId !== 'number') {
          throw new Error('spell_id must be a string or number');
        }
        // Validate position if provided
        if (params.position) {
          if (typeof params.position !== 'object' || 
              typeof params.position.x !== 'number' || 
              typeof params.position.y !== 'number') {
            throw new Error('position must be an object with numeric x and y coordinates');
          }
        }
        break;
        
      case 'joinGame':
        if (!params.player_name && !params.playerName) {
          throw new Error('joinGame requires player_name');
        }
        const playerName = params.player_name || params.playerName;
        if (typeof playerName !== 'string' || playerName.trim().length === 0) {
          throw new Error('player_name must be a non-empty string');
        }
        if (playerName.length > 50) {
          throw new Error('player_name must be 50 characters or less');
        }
        // Basic sanitation check - no control characters
        if (/[\x00-\x1F\x7F]/.test(playerName)) {
          throw new Error('player_name contains invalid characters');
        }
        break;
        
      case 'startCombat':
        if (!params.participant_ids && !params.participantIds) {
          throw new Error('startCombat requires participant_ids');
        }
        const participantIds = params.participant_ids || params.participantIds;
        if (!Array.isArray(participantIds)) {
          throw new Error('participant_ids must be an array');
        }
        if (participantIds.length === 0) {
          throw new Error('participant_ids cannot be empty');
        }
        participantIds.forEach((id, index) => {
          if (typeof id !== 'string' && typeof id !== 'number') {
            throw new Error(`participant_ids[${index}] must be a string or number`);
          }
        });
        break;
        
      case 'getGameState':
      case 'endTurn':
      case 'leaveGame':
        // These methods don't require additional parameters beyond session_id
        break;
        
      default:
        // For unknown methods, just validate basic parameter structure
        if (params && typeof params !== 'object') {
          throw new Error('Parameters must be an object');
        }
        break;
    }
  }
}

// Test cases
function runValidationTests() {
  const client = new TestRPCClient();
  let passed = 0;
  let failed = 0;
  
  function test(name, response, expected) {
    const result = client.validateJSONRPCResponse(response);
    if (result === expected) {
      console.log(`✓ ${name}`);
      passed++;
    } else {
      console.log(`✗ ${name} - Expected ${expected}, got ${result}`);
      failed++;
    }
  }
  
  console.log('Running JSON-RPC Response Validation Tests...\n');
  
  // Valid response with result
  test('Valid response with result', {
    jsonrpc: "2.0",
    result: { success: true },
    id: 1
  }, true);
  
  // Valid response with error
  test('Valid response with error', {
    jsonrpc: "2.0",
    error: { code: -32600, message: "Invalid Request" },
    id: 1
  }, true);
  
  // Valid notification response (id can be null)
  test('Valid notification response', {
    jsonrpc: "2.0",
    result: { success: true },
    id: null
  }, true);
  
  // Invalid: missing jsonrpc
  test('Invalid: missing jsonrpc', {
    result: { success: true },
    id: 1
  }, false);
  
  // Invalid: wrong jsonrpc version
  test('Invalid: wrong jsonrpc version', {
    jsonrpc: "1.0",
    result: { success: true },
    id: 1
  }, false);
  
  // Invalid: both result and error
  test('Invalid: both result and error', {
    jsonrpc: "2.0",
    result: { success: true },
    error: { code: -32600, message: "Invalid Request" },
    id: 1
  }, false);
  
  // Invalid: neither result nor error
  test('Invalid: neither result nor error', {
    jsonrpc: "2.0",
    id: 1
  }, false);
  
  // Invalid: missing id
  test('Invalid: missing id', {
    jsonrpc: "2.0",
    result: { success: true }
  }, false);
  
  // Invalid: malformed error (missing code)
  test('Invalid: malformed error (missing code)', {
    jsonrpc: "2.0",
    error: { message: "Invalid Request" },
    id: 1
  }, false);
  
  // Invalid: malformed error (missing message)
  test('Invalid: malformed error (missing message)', {
    jsonrpc: "2.0",
    error: { code: -32600 },
    id: 1
  }, false);
  
  // Invalid: null response
  test('Invalid: null response', null, false);
  
  // Invalid: non-object response
  test('Invalid: non-object response', "invalid", false);
  
  console.log(`\nResults: ${passed} passed, ${failed} failed`);
  return failed === 0;
}

// Parameter validation tests
function runParameterValidationTests() {
  const client = new TestRPCClient();
  let passed = 0;
  let failed = 0;
  
  function test(name, method, params, shouldPass) {
    try {
      client.validateMethodParameters(method, params);
      if (shouldPass) {
        console.log(`✓ ${name}`);
        passed++;
      } else {
        console.log(`✗ ${name} - Expected validation to fail but it passed`);
        failed++;
      }
    } catch (error) {
      if (!shouldPass) {
        console.log(`✓ ${name} - Correctly rejected: ${error.message}`);
        passed++;
      } else {
        console.log(`✗ ${name} - Unexpected validation error: ${error.message}`);
        failed++;
      }
    }
  }
  
  console.log('\nRunning Parameter Validation Tests...\n');
  
  // Move validation tests
  test('Valid move - up', 'move', { direction: 'up' }, true);
  test('Valid move - ne', 'move', { direction: 'ne' }, true);
  test('Invalid move - bad direction', 'move', { direction: 'invalid' }, false);
  test('Invalid move - missing direction', 'move', {}, false);
  
  // Attack validation tests
  test('Valid attack', 'attack', { target_id: 'enemy1', weapon_id: 'sword' }, true);
  test('Valid attack - alt params', 'attack', { targetId: 123, weaponId: 456 }, true);
  test('Invalid attack - missing target', 'attack', { weapon_id: 'sword' }, false);
  test('Invalid attack - missing weapon', 'attack', { target_id: 'enemy1' }, false);
  test('Invalid attack - bad target type', 'attack', { target_id: {}, weapon_id: 'sword' }, false);
  
  // Spell casting validation tests
  test('Valid castSpell - with target', 'castSpell', { spell_id: 'fireball', target_id: 'enemy1' }, true);
  test('Valid castSpell - with position', 'castSpell', { spell_id: 'fireball', position: { x: 10, y: 20 } }, true);
  test('Invalid castSpell - missing spell_id', 'castSpell', { target_id: 'enemy1' }, false);
  test('Invalid castSpell - missing target and position', 'castSpell', { spell_id: 'fireball' }, false);
  test('Invalid castSpell - bad position', 'castSpell', { spell_id: 'fireball', position: { x: 'bad' } }, false);
  
  // Join game validation tests
  test('Valid joinGame', 'joinGame', { player_name: 'TestPlayer' }, true);
  test('Valid joinGame - alt param', 'joinGame', { playerName: 'TestPlayer' }, true);
  test('Invalid joinGame - empty name', 'joinGame', { player_name: '' }, false);
  test('Invalid joinGame - whitespace only', 'joinGame', { player_name: '   ' }, false);
  test('Invalid joinGame - too long', 'joinGame', { player_name: 'a'.repeat(51) }, false);
  test('Invalid joinGame - control chars', 'joinGame', { player_name: 'test\x00player' }, false);
  test('Invalid joinGame - missing name', 'joinGame', {}, false);
  
  // Start combat validation tests
  test('Valid startCombat', 'startCombat', { participant_ids: ['player1', 'enemy1'] }, true);
  test('Valid startCombat - alt param', 'startCombat', { participantIds: [123, 456] }, true);
  test('Invalid startCombat - empty array', 'startCombat', { participant_ids: [] }, false);
  test('Invalid startCombat - not array', 'startCombat', { participant_ids: 'notarray' }, false);
  test('Invalid startCombat - bad participant type', 'startCombat', { participant_ids: ['valid', {}] }, false);
  test('Invalid startCombat - missing param', 'startCombat', {}, false);
  
  // Methods without parameters
  test('Valid getGameState', 'getGameState', {}, true);
  test('Valid endTurn', 'endTurn', {}, true);
  test('Valid leaveGame', 'leaveGame', {}, true);
  
  // General validation tests
  test('Invalid method name - empty', '', { test: 'value' }, false);
  test('Invalid method name - not string', 123, { test: 'value' }, false);
  test('Invalid params - not object', 'testMethod', 'notobject', false);
  test('Valid unknown method', 'unknownMethod', { test: 'value' }, true);
  
  console.log(`\nParameter Validation Results: ${passed} passed, ${failed} failed`);
  return failed === 0;
}

// Run tests if this file is executed directly
if (typeof window !== 'undefined') {
  const responseTestsPassed = runValidationTests();
  const parameterTestsPassed = runParameterValidationTests();
  
  console.log('\n=== OVERALL TEST RESULTS ===');
  console.log(`Response validation: ${responseTestsPassed ? 'PASSED' : 'FAILED'}`);
  console.log(`Parameter validation: ${parameterTestsPassed ? 'PASSED' : 'FAILED'}`);
  console.log(`Overall: ${responseTestsPassed && parameterTestsPassed ? 'PASSED' : 'FAILED'}`);
}