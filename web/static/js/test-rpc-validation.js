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

// Mock RPCClient with just the validation method for testing
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

// Run tests if this file is executed directly
if (typeof window !== 'undefined') {
  runValidationTests();
}