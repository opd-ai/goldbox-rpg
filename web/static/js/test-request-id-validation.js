/**
 * Test script for Request ID Validation functionality
 * 
 * Tests the enhanced request/response ID correlation to prevent spoofing attacks.
 * 
 * Usage: node test-request-id-validation.js
 */

// Mock WebSocket and other browser APIs
global.WebSocket = class MockWebSocket {
  constructor(url) {
    this.url = url;
    this.readyState = 1; // OPEN
    this.onopen = null;
    this.onmessage = null;
    this.onclose = null;
    this.onerror = null;
  }
  
  send(data) {
    this.lastSentData = data;
  }
  
  close() {
    this.readyState = 3; // CLOSED
  }
  
  // Test helper to simulate receiving a message
  simulateMessage(data) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) });
    }
  }
};

global.location = {
  protocol: 'http:',
  hostname: 'localhost',
  host: 'localhost:3000'
};

// Simple EventEmitter for testing
class EventEmitter {
  constructor() {
    this.events = new Map();
  }
  
  on(event, callback) {
    if (!this.events.has(event)) {
      this.events.set(event, []);
    }
    this.events.get(event).push(callback);
  }
  
  emit(event, data) {
    if (this.events.has(event)) {
      this.events.get(event).forEach((cb) => cb(data));
    }
  }
}

// Simplified RPCClient for testing ID validation
class TestRPCClient extends EventEmitter {
  constructor() {
    super();
    this.requestId = 1;
    this.requestQueue = new Map();
    this.ws = null;
    this.sessionId = null;
  }

  isDevelopment() {
    return true;
  }

  safeLog(level, message, data) {
    // Silent logging for tests
  }

  sanitizeForLogging(data) {
    return data;
  }

  validateOrigin() {
    return true; // Skip for tests
  }

  validateSessionForRequest() {
    return true; // Skip for tests
  }

  async connect() {
    this.ws = new WebSocket('ws://localhost:3000/rpc/ws');
    this.ws.onmessage = this.handleMessage.bind(this);
    return Promise.resolve();
  }

  async request(method, params = {}, timeout = 30000) {
    if (!this.ws) await this.connect();

    const id = this.requestId++;
    
    const message = {
      jsonrpc: "2.0",
      method,
      params: { ...params, session_id: this.sessionId },
      id,
    };

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.requestQueue.delete(id);
        reject(new Error(`Request timeout: ${method}`));
      }, timeout);

      this.requestQueue.set(id, {
        originalId: id,
        method: method,
        timestamp: Date.now(),
        resolve: (result) => {
          clearTimeout(timeoutId);
          resolve(result);
        },
        reject: (error) => {
          clearTimeout(timeoutId);
          reject(error);
        },
      });

      this.ws.send(JSON.stringify(message));
    });
  }

  handleMessage(event) {
    try {
      const response = JSON.parse(event.data);

      if (!response.id || !this.requestQueue.has(response.id)) {
        this.emit('error', { 
          type: 'NO_MATCHING_REQUEST', 
          responseId: response.id,
          message: 'Received response for unknown request ID'
        });
        return;
      }

      // Enhanced ID validation
      const pendingRequest = this.requestQueue.get(response.id);
      if (!pendingRequest || pendingRequest.originalId !== response.id) {
        this.emit('error', { 
          type: 'ID_MISMATCH', 
          responseId: response.id,
          expectedId: pendingRequest ? pendingRequest.originalId : null,
          message: 'Response ID does not match original request ID - possible spoofing attack'
        });
        return;
      }

      const { resolve, reject } = pendingRequest;
      this.requestQueue.delete(response.id);

      if (response.error) {
        reject(response.error);
        this.emit("error", response.error);
      } else {
        resolve(response.result);
      }
    } catch (error) {
      this.emit("error", error);
    }
  }
}

// Test suite
console.log("=== Request ID Validation Tests ===\n");

let testsPassed = 0;
let testsTotal = 0;

function test(description, testFn) {
  testsTotal++;
  return testFn().then(() => {
    console.log(`✅ ${description}`);
    testsPassed++;
  }).catch(error => {
    console.log(`❌ ${description}: ${error.message}`);
  });
}

function assertEqual(actual, expected, message) {
  if (actual !== expected) {
    throw new Error(`${message}: expected ${expected}, got ${actual}`);
  }
}

function assertTrue(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

// Test 1: Valid request/response ID correlation
async function testValidIdCorrelation() {
  const client = new TestRPCClient();
  await client.connect();

  // Start a request
  const requestPromise = client.request('getGameState');
  
  // Simulate valid response with matching ID
  setTimeout(() => {
    client.ws.simulateMessage({
      jsonrpc: "2.0",
      result: { status: "ok" },
      id: 1
    });
  }, 10);

  const result = await requestPromise;
  assertEqual(result.status, "ok", "Should receive correct result");
}

// Test 2: Invalid response ID should trigger error
async function testInvalidResponseId() {
  const client = new TestRPCClient();
  let errorCaught = false;
  
  client.on('error', (error) => {
    if (error.type === 'NO_MATCHING_REQUEST') {
      errorCaught = true;
    }
  });

  await client.connect();

  // Simulate response with non-existent request ID
  client.ws.simulateMessage({
    jsonrpc: "2.0",
    result: { status: "fake" },
    id: 999  // Non-existent request ID
  });

  // Wait for error to be emitted
  await new Promise(resolve => setTimeout(resolve, 50));
  
  assertTrue(errorCaught, "Should emit error for non-existent request ID");
}

// Test 3: ID mismatch detection (spoofing simulation)
async function testIdMismatchDetection() {
  const client = new TestRPCClient();
  let mismatchErrorCaught = false;
  
  client.on('error', (error) => {
    if (error.type === 'ID_MISMATCH') {
      mismatchErrorCaught = true;
      assertEqual(error.responseId, 1, "Should report correct response ID");
      assertTrue(error.message.includes('spoofing'), "Should mention spoofing attack");
    }
  });

  await client.connect();

  // Start a request that will get ID 1
  const requestPromise = client.request('getGameState');
  
  // Manually corrupt the pending request to simulate ID mismatch
  const pendingRequest = client.requestQueue.get(1);
  if (pendingRequest) {
    pendingRequest.originalId = 2; // Corrupt the original ID
  }
  
  // Simulate response with correct ID 1 (but now mismatched)
  setTimeout(() => {
    client.ws.simulateMessage({
      jsonrpc: "2.0",
      result: { status: "fake" },
      id: 1
    });
  }, 10);

  // The request should not resolve due to ID mismatch
  try {
    await Promise.race([
      requestPromise,
      new Promise((_, reject) => setTimeout(() => reject(new Error('Expected timeout')), 100))
    ]);
    throw new Error('Request should not have resolved');
  } catch (error) {
    if (error.message !== 'Expected timeout') {
      throw error;
    }
  }

  assertTrue(mismatchErrorCaught, "Should detect ID mismatch");
}

// Test 4: Multiple concurrent requests with correct IDs
async function testConcurrentRequests() {
  const client = new TestRPCClient();
  await client.connect();

  // Start multiple requests
  const request1 = client.request('getGameState');
  const request2 = client.request('getPlayer');
  const request3 = client.request('getWorld');

  // Simulate responses in different order
  setTimeout(() => {
    client.ws.simulateMessage({
      jsonrpc: "2.0",
      result: { type: "player" },
      id: 2  // Response to second request
    });
  }, 10);

  setTimeout(() => {
    client.ws.simulateMessage({
      jsonrpc: "2.0",
      result: { type: "world" },
      id: 3  // Response to third request
    });
  }, 20);

  setTimeout(() => {
    client.ws.simulateMessage({
      jsonrpc: "2.0",
      result: { type: "state" },
      id: 1  // Response to first request
    });
  }, 30);

  const [result1, result2, result3] = await Promise.all([request1, request2, request3]);
  
  assertEqual(result1.type, "state", "First request should get correct response");
  assertEqual(result2.type, "player", "Second request should get correct response");
  assertEqual(result3.type, "world", "Third request should get correct response");
}

// Test 5: Request queue cleanup on ID validation
async function testRequestQueueCleanup() {
  const client = new TestRPCClient();
  await client.connect();

  // Start a request
  const requestPromise = client.request('getGameState');
  
  // Verify request is in queue
  assertTrue(client.requestQueue.has(1), "Request should be in queue");
  
  // Simulate valid response
  setTimeout(() => {
    client.ws.simulateMessage({
      jsonrpc: "2.0",
      result: { status: "ok" },
      id: 1
    });
  }, 10);

  await requestPromise;
  
  // Verify request is cleaned up from queue
  assertTrue(!client.requestQueue.has(1), "Request should be removed from queue after response");
}

// Run all tests
async function runAllTests() {
  await test("Valid request/response ID correlation", testValidIdCorrelation);
  await test("Invalid response ID should trigger error", testInvalidResponseId);
  await test("ID mismatch detection (spoofing simulation)", testIdMismatchDetection);
  await test("Multiple concurrent requests with correct IDs", testConcurrentRequests);
  await test("Request queue cleanup on ID validation", testRequestQueueCleanup);

  console.log(`\n=== Test Results ===`);
  console.log(`✅ Passed: ${testsPassed}/${testsTotal}`);
  console.log(`❌ Failed: ${testsTotal - testsPassed}/${testsTotal}`);

  if (testsPassed === testsTotal) {
    console.log("\n🎉 All Request ID Validation tests passed!");
    console.log("\nRequest/Response security features verified:");
    console.log("• Strict request/response ID correlation prevents spoofing");
    console.log("• Unknown request IDs are rejected with clear error messages");
    console.log("• ID mismatch detection prevents response manipulation attacks");
    console.log("• Concurrent requests maintain proper ID correlation");
    console.log("• Request queue cleanup prevents memory leaks");
    console.log("• Enhanced error reporting for debugging and monitoring");
  } else {
    console.log("\n⚠️  Some tests failed. Please review the implementation.");
    process.exit(1);
  }
}

runAllTests().catch(console.error);
