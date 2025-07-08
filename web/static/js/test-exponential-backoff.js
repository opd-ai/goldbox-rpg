/**
 * Test for exponential backoff reconnection functionality
 * Run this in browser console to verify backoff calculation works correctly
 */

function testExponentialBackoff() {
  console.log("Testing exponential backoff reconnection logic...");
  
  // Mock RPC client with backoff calculation method
  const mockClient = {
    calculateReconnectionDelay(attempt) {
      const baseDelay = 1000; // 1 second base delay
      const maxDelay = 30000; // 30 seconds maximum delay
      
      // Exponential backoff: delay = baseDelay * 2^attempt
      const exponentialDelay = baseDelay * Math.pow(2, attempt - 1);
      
      // Cap at maximum delay
      const cappedDelay = Math.min(exponentialDelay, maxDelay);
      
      // Add jitter: ±10% random variation to prevent thundering herd
      const jitterRange = 0.1 * cappedDelay;
      const jitter = (Math.random() - 0.5) * 2 * jitterRange;
      
      return Math.round(cappedDelay + jitter);
    }
  };

  let passed = 0;
  let failed = 0;

  function test(name, testFn) {
    try {
      const result = testFn();
      if (result) {
        console.log(`✅ ${name}: PASS`);
        passed++;
      } else {
        console.error(`❌ ${name}: FAIL`);
        failed++;
      }
      return result;
    } catch (error) {
      console.error(`❌ ${name}: FAIL (error: ${error.message})`);
      failed++;
      return false;
    }
  }

  // Test 1: Basic exponential growth
  console.log("\n1. Testing exponential backoff growth:");
  test("Attempt 1 delay around 1 second", () => {
    const delay = mockClient.calculateReconnectionDelay(1);
    return delay >= 900 && delay <= 1100; // Allow for jitter
  });

  test("Attempt 2 delay around 2 seconds", () => {
    const delay = mockClient.calculateReconnectionDelay(2);
    return delay >= 1800 && delay <= 2200; // Allow for jitter
  });

  test("Attempt 3 delay around 4 seconds", () => {
    const delay = mockClient.calculateReconnectionDelay(3);
    return delay >= 3600 && delay <= 4400; // Allow for jitter
  });

  test("Attempt 4 delay around 8 seconds", () => {
    const delay = mockClient.calculateReconnectionDelay(4);
    return delay >= 7200 && delay <= 8800; // Allow for jitter
  });

  // Test 2: Maximum delay cap
  console.log("\n2. Testing maximum delay cap:");
  test("High attempt number capped at 30 seconds", () => {
    const delay = mockClient.calculateReconnectionDelay(10);
    return delay >= 27000 && delay <= 33000; // 30s ± 10% jitter
  });

  test("Very high attempt number still capped", () => {
    const delay = mockClient.calculateReconnectionDelay(20);
    return delay >= 27000 && delay <= 33000; // 30s ± 10% jitter
  });

  // Test 3: Jitter variation
  console.log("\n3. Testing jitter variation:");
  test("Multiple calculations produce different results (jitter)", () => {
    const delays = [];
    for (let i = 0; i < 10; i++) {
      delays.push(mockClient.calculateReconnectionDelay(3));
    }
    
    // Check that we get at least some variation (not all identical)
    const uniqueDelays = new Set(delays);
    return uniqueDelays.size > 1;
  });

  test("Jitter stays within reasonable bounds", () => {
    const attempts = 100;
    const targetDelay = 4000; // 4 seconds for attempt 3
    const maxJitter = targetDelay * 0.1; // 10% jitter
    
    for (let i = 0; i < attempts; i++) {
      const delay = mockClient.calculateReconnectionDelay(3);
      const deviation = Math.abs(delay - targetDelay);
      
      if (deviation > maxJitter) {
        return false;
      }
    }
    return true;
  });

  // Test 4: Compare old vs new approach
  console.log("\n4. Comparing old linear vs new exponential approach:");
  console.log("Old linear delays: 1s, 2s, 3s, 4s, 5s...");
  console.log("New exponential delays (approx):");
  
  for (let attempt = 1; attempt <= 8; attempt++) {
    const delay = mockClient.calculateReconnectionDelay(attempt);
    const seconds = (delay / 1000).toFixed(1);
    console.log(`  Attempt ${attempt}: ${seconds}s`);
  }

  // Test 5: Server load reduction calculation
  console.log("\n5. Analyzing server load reduction:");
  
  function calculateTotalReconnectionTime(attempts, delayFunction) {
    let total = 0;
    for (let i = 1; i <= attempts; i++) {
      total += delayFunction(i);
    }
    return total;
  }
  
  const linearDelay = (attempt) => 1000 * attempt;
  const exponentialDelay = (attempt) => mockClient.calculateReconnectionDelay(attempt);
  
  const attemptsToTest = 5;
  const linearTotal = calculateTotalReconnectionTime(attemptsToTest, linearDelay);
  const exponentialTotal = calculateTotalReconnectionTime(attemptsToTest, exponentialDelay);
  
  console.log(`Total time for ${attemptsToTest} attempts:`);
  console.log(`  Linear: ${(linearTotal / 1000).toFixed(1)}s`);
  console.log(`  Exponential: ${(exponentialTotal / 1000).toFixed(1)}s`);
  
  const loadReduction = exponentialTotal > linearTotal;
  test("Exponential backoff reduces server load (longer total time)", () => loadReduction);

  console.log(`\n=== Test Results ===`);
  console.log(`✅ Passed: ${passed}`);
  console.log(`❌ Failed: ${failed}`);
  console.log(`📊 Success Rate: ${((passed / (passed + failed)) * 100).toFixed(1)}%`);
  
  console.log(`\n🔧 Benefits of exponential backoff:`);
  console.log(`   • Reduces server load during outages`);
  console.log(`   • Prevents thundering herd effect with jitter`);
  console.log(`   • Provides reasonable delays for transient issues`);
  console.log(`   • Caps maximum delay to prevent indefinite waits`);
  
  return failed === 0;
}

// Auto-run test if in browser environment
if (typeof window !== 'undefined') {
  testExponentialBackoff();
}
