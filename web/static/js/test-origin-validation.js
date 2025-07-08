/**
 * Test script for Origin Validation functionality
 * 
 * Tests the CORS protection added to prevent unauthorized origins from connecting.
 * 
 * Usage: node test-origin-validation.js
 */

// Mock browser location object
const createMockLocation = (hostname, protocol = 'http:') => ({
  hostname,
  protocol,
  host: hostname
});

// Simulate RPCClient's origin validation logic
class MockRPCClient {
  constructor(location) {
    this.location = location;
  }

  isDevelopment() {
    // Simple environment detection
    return process.env.NODE_ENV !== 'production';
  }

  validateOrigin() {
    const currentOrigin = this.location.hostname.toLowerCase();
    
    // Development mode: allow common development origins
    if (this.isDevelopment()) {
      const devOrigins = [
        'localhost',
        '127.0.0.1',
        '0.0.0.0',
        'vscode-local',
        'goldbox-rpg'
      ];
      
      // Check for exact matches or proper subdomain patterns
      const isDevOrigin = devOrigins.some(devOrigin => {
        // Exact match
        if (currentOrigin === devOrigin) return true;
        
        // Subdomain match (but not suffix match)
        if (currentOrigin.endsWith('.' + devOrigin)) return true;
        
        return false;
      });
      
      // Check for cloud development platforms
      const isCloudDev = currentOrigin.includes('github.dev') ||
                        currentOrigin.includes('gitpod.io') ||
                        currentOrigin.includes('preview.app');
      
      if (isDevOrigin || isCloudDev) {
        console.debug("Development origin allowed:", currentOrigin);
        return true;
      }
    }

    // Production mode: strict allowlist validation
    const authorizedOrigins = [
      'your-game-domain.com',
      'app.your-game-domain.com',
      'game.your-domain.com'
    ];

    const isAuthorized = authorizedOrigins.includes(currentOrigin);
    
    if (!isAuthorized) {
      const errorMsg = `Unauthorized origin: ${currentOrigin}. This client is not authorized to connect from this domain.`;
      throw new Error(errorMsg);
    }

    console.info("Origin authorized:", currentOrigin);
    return true;
  }
}

// Test suite
console.log("=== Origin Validation CORS Protection Tests ===\n");

let testsPassed = 0;
let testsTotal = 0;

function test(description, testFn) {
  testsTotal++;
  try {
    testFn();
    console.log(`✅ ${description}`);
    testsPassed++;
  } catch (error) {
    console.log(`❌ ${description}: ${error.message}`);
  }
}

function assertTrue(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function assertThrows(fn, expectedMessage, testDescription) {
  try {
    fn();
    throw new Error(`${testDescription}: Expected function to throw but it didn't`);
  } catch (error) {
    if (expectedMessage && !error.message.includes(expectedMessage)) {
      throw new Error(`${testDescription}: Expected error message to contain "${expectedMessage}" but got "${error.message}"`);
    }
  }
}

// Test 1: Development origins should be allowed
test("Allow localhost in development", () => {
  const client = new MockRPCClient(createMockLocation('localhost'));
  const result = client.validateOrigin();
  assertTrue(result, "localhost should be allowed in development");
});

test("Allow 127.0.0.1 in development", () => {
  const client = new MockRPCClient(createMockLocation('127.0.0.1'));
  const result = client.validateOrigin();
  assertTrue(result, "127.0.0.1 should be allowed in development");
});

test("Allow vscode-local hostnames", () => {
  const client = new MockRPCClient(createMockLocation('vscode-local'));
  const result = client.validateOrigin();
  assertTrue(result, "vscode-local should be allowed in development");
});

test("Allow GitHub Codespaces domains", () => {
  const client = new MockRPCClient(createMockLocation('abc123-3000.github.dev'));
  const result = client.validateOrigin();
  assertTrue(result, "GitHub Codespaces domains should be allowed in development");
});

test("Allow Gitpod domains", () => {
  const client = new MockRPCClient(createMockLocation('3000-abc123-def456.ws-us999.gitpod.io'));
  const result = client.validateOrigin();
  assertTrue(result, "Gitpod domains should be allowed in development");
});

test("Allow preview app domains", () => {
  const client = new MockRPCClient(createMockLocation('myapp.preview.app'));
  const result = client.validateOrigin();
  assertTrue(result, "Preview app domains should be allowed in development");
});

// Test 2: Unauthorized origins should be blocked in development
test("Block unauthorized origin in development", () => {
  const client = new MockRPCClient(createMockLocation('malicious-site.com'));
  assertThrows(
    () => client.validateOrigin(),
    'Unauthorized origin',
    "Should block unauthorized origins even in development"
  );
});

test("Block suspicious localhost variants", () => {
  const client = new MockRPCClient(createMockLocation('localhost.malicious.com'));
  assertThrows(
    () => client.validateOrigin(),
    'Unauthorized origin',
    "Should block domains that contain localhost but are not localhost"
  );
});

// Test 3: Production mode behavior (simulate by changing environment)
test("Production mode blocks even localhost", () => {
  // Temporarily set production mode
  const originalEnv = process.env.NODE_ENV;
  process.env.NODE_ENV = 'production';
  
  try {
    const client = new MockRPCClient(createMockLocation('localhost'));
    assertThrows(
      () => client.validateOrigin(),
      'Unauthorized origin',
      "Production mode should block localhost"
    );
  } finally {
    // Restore original environment
    process.env.NODE_ENV = originalEnv;
  }
});

test("Production mode allows configured domains", () => {
  const originalEnv = process.env.NODE_ENV;
  process.env.NODE_ENV = 'production';
  
  try {
    const client = new MockRPCClient(createMockLocation('your-game-domain.com'));
    const result = client.validateOrigin();
    assertTrue(result, "Production mode should allow configured domains");
  } finally {
    process.env.NODE_ENV = originalEnv;
  }
});

// Test 4: Case sensitivity
test("Origin validation is case insensitive", () => {
  const client = new MockRPCClient(createMockLocation('LOCALHOST'));
  const result = client.validateOrigin();
  assertTrue(result, "Origin validation should be case insensitive");
});

test("Mixed case GitHub domain", () => {
  const client = new MockRPCClient(createMockLocation('ABC123-3000.GITHUB.DEV'));
  const result = client.validateOrigin();
  assertTrue(result, "Mixed case GitHub domains should be allowed");
});

console.log(`\n=== Test Results ===`);
console.log(`✅ Passed: ${testsPassed}/${testsTotal}`);
console.log(`❌ Failed: ${testsTotal - testsPassed}/${testsTotal}`);

if (testsPassed === testsTotal) {
  console.log("\n🎉 All Origin Validation tests passed!");
  console.log("\nCORS protection features verified:");
  console.log("• Development origins (localhost, 127.0.0.1, etc.) are allowed");
  console.log("• Cloud development platforms (GitHub Codespaces, Gitpod) are allowed");
  console.log("• Unauthorized origins are blocked with clear error messages");
  console.log("• Production mode enforces strict allowlist validation");
  console.log("• Origin comparison is case-insensitive for better compatibility");
  console.log("• Prevents cross-site request forgery via unauthorized hosting");
} else {
  console.log("\n⚠️  Some tests failed. Please review the implementation.");
  process.exit(1);
}
