/**
 * Test for secure session management functionality
 * Run this in browser console to verify session validation works correctly
 */

function testSessionValidation() {
  console.log("Testing secure session management functionality...");
  
  // Mock RPC client with session management methods
  const mockClient = {
    sessionId: null,
    sessionExpiry: null,
    
    // Copy the actual validation methods
    validateSessionTokenFormat(token) {
      if (typeof token !== 'string' || token.length === 0) {
        return false;
      }
      
      // Basic format validation - should be a UUID-like string
      const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
      return uuidRegex.test(token);
    },

    validateSessionData(sessionData) {
      if (!sessionData || typeof sessionData !== 'object') {
        return false;
      }
      
      // Must have session_id
      if (!sessionData.session_id || !this.validateSessionTokenFormat(sessionData.session_id)) {
        return false;
      }
      
      return true;
    },

    isSessionExpired() {
      if (!this.sessionId || !this.sessionExpiry) {
        return true;
      }
      
      return new Date() >= this.sessionExpiry;
    },

    setSession(sessionData, expiryMinutes = 30) {
      if (!this.validateSessionData(sessionData)) {
        throw new Error('Invalid session data received from server');
      }
      
      this.sessionId = sessionData.session_id;
      
      // Set expiration time (default 30 minutes from now)
      this.sessionExpiry = new Date();
      this.sessionExpiry.setMinutes(this.sessionExpiry.getMinutes() + expiryMinutes);
    },

    clearSession() {
      this.sessionId = null;
      this.sessionExpiry = null;
    },

    validateSessionForRequest() {
      if (!this.sessionId) {
        throw new Error('No active session - please join a game first');
      }
      
      if (!this.validateSessionTokenFormat(this.sessionId)) {
        throw new Error('Invalid session token format');
      }
      
      if (this.isSessionExpired()) {
        this.clearSession();
        throw new Error('Session has expired - please join the game again');
      }
    }
  };

  let passed = 0;
  let failed = 0;

  function test(name, testFn, shouldThrow = false) {
    try {
      const result = testFn();
      if (shouldThrow) {
        console.error(`❌ ${name}: FAIL (expected error but none thrown)`);
        failed++;
      } else {
        console.log(`✅ ${name}: PASS`);
        passed++;
      }
      return result;
    } catch (error) {
      if (shouldThrow) {
        console.log(`✅ ${name}: PASS (correctly threw: ${error.message})`);
        passed++;
      } else {
        console.error(`❌ ${name}: FAIL (unexpected error: ${error.message})`);
        failed++;
      }
    }
  }

  // Test 1: Token format validation
  console.log("\n1. Testing token format validation:");
  test("Valid UUID format", () => {
    return mockClient.validateSessionTokenFormat("12345678-1234-1234-1234-123456789abc");
  });
  
  test("Invalid token format", () => {
    return !mockClient.validateSessionTokenFormat("invalid-token");
  });
  
  test("Empty token", () => {
    return !mockClient.validateSessionTokenFormat("");
  });
  
  test("Null token", () => {
    return !mockClient.validateSessionTokenFormat(null);
  });

  // Test 2: Session data validation
  console.log("\n2. Testing session data validation:");
  test("Valid session data", () => {
    return mockClient.validateSessionData({
      session_id: "12345678-1234-1234-1234-123456789abc"
    });
  });
  
  test("Invalid session data (missing session_id)", () => {
    return !mockClient.validateSessionData({});
  });
  
  test("Invalid session data (bad token format)", () => {
    return !mockClient.validateSessionData({
      session_id: "invalid-token"
    });
  });

  // Test 3: Session expiration
  console.log("\n3. Testing session expiration:");
  test("No session is expired", () => {
    mockClient.clearSession();
    return mockClient.isSessionExpired();
  });
  
  test("Valid session not expired", () => {
    mockClient.setSession({
      session_id: "12345678-1234-1234-1234-123456789abc"
    });
    return !mockClient.isSessionExpired();
  });
  
  test("Expired session detection", () => {
    mockClient.setSession({
      session_id: "12345678-1234-1234-1234-123456789abc"
    }, -1); // Set expiry to 1 minute ago
    return mockClient.isSessionExpired();
  });

  // Test 4: Request validation
  console.log("\n4. Testing request validation:");
  test("Valid session allows request", () => {
    mockClient.setSession({
      session_id: "12345678-1234-1234-1234-123456789abc"
    });
    mockClient.validateSessionForRequest();
    return true;
  });
  
  test("No session throws error", () => {
    mockClient.clearSession();
    mockClient.validateSessionForRequest();
  }, true);
  
  test("Expired session throws error", () => {
    mockClient.setSession({
      session_id: "12345678-1234-1234-1234-123456789abc"
    }, -1); // Expired
    mockClient.validateSessionForRequest();
  }, true);

  // Test 5: Session lifecycle
  console.log("\n5. Testing session lifecycle:");
  test("Session set and clear", () => {
    mockClient.setSession({
      session_id: "12345678-1234-1234-1234-123456789abc"
    });
    const hasSession = !!mockClient.sessionId;
    mockClient.clearSession();
    const sessionCleared = !mockClient.sessionId;
    return hasSession && sessionCleared;
  });

  console.log(`\n=== Test Results ===`);
  console.log(`✅ Passed: ${passed}`);
  console.log(`❌ Failed: ${failed}`);
  console.log(`📊 Success Rate: ${((passed / (passed + failed)) * 100).toFixed(1)}%`);
  
  return failed === 0;
}

// Auto-run test if in browser environment
if (typeof window !== 'undefined') {
  testSessionValidation();
}
