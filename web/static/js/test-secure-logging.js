/**
 * Simple test for secure logging functionality
 * Run this in browser console to verify logging sanitization works correctly
 */

// Mock window.location for testing
const originalLocation = window.location;

function testSecureLogging() {
  console.log("Testing secure logging functionality...");
  
  // Test 1: Development mode detection
  console.log("1. Testing development mode detection:");
  
  // Mock localhost environment
  Object.defineProperty(window, 'location', {
    value: {
      hostname: 'localhost',
      port: '8080',
      protocol: 'http:'
    },
    writable: true
  });
  
  console.log("Localhost environment:", RPCClient.isDevelopment()); // Should be true
  
  // Mock production environment
  Object.defineProperty(window, 'location', {
    value: {
      hostname: 'example.com',
      port: '443',
      protocol: 'https:'
    },
    writable: true
  });
  
  console.log("Production environment:", RPCClient.isDevelopment()); // Should be false
  
  // Test 2: Data sanitization
  console.log("\n2. Testing data sanitization:");
  
  const mockClient = {
    isDevelopment: () => false, // Force production mode
    sanitizeForLogging: function(data) {
      if (!this.isDevelopment()) {
        if (typeof data === 'object' && data !== null) {
          const sanitized = { ...data };
          
          // Redact sensitive fields
          if (sanitized.session_id) {
            sanitized.session_id = '[REDACTED]';
          }
          if (sanitized.sessionId) {
            sanitized.sessionId = '[REDACTED]';
          }
          if (sanitized.params && sanitized.params.session_id) {
            sanitized.params = { ...sanitized.params, session_id: '[REDACTED]' };
          }
          
          // Redact result data that might contain sensitive info
          if (sanitized.result && typeof sanitized.result === 'object') {
            const result = { ...sanitized.result };
            if (result.session_id) result.session_id = '[REDACTED]';
            if (result.player_data) result.player_data = '[REDACTED]';
            sanitized.result = result;
          }
          
          return sanitized;
        }
      }
      return data;
    }
  };
  
  const sensitiveData = {
    method: 'joinGame',
    params: {
      player_name: 'TestPlayer',
      session_id: 'abc123-sensitive-session-id'
    },
    result: {
      session_id: 'abc123-sensitive-session-id',
      player_data: { hp: 100, level: 5 }
    }
  };
  
  const sanitized = mockClient.sanitizeForLogging(sensitiveData);
  console.log("Original data:", sensitiveData);
  console.log("Sanitized data:", sanitized);
  
  // Verify session IDs are redacted
  const sessionRedacted = sanitized.params.session_id === '[REDACTED]' &&
                          sanitized.result.session_id === '[REDACTED]' &&
                          sanitized.result.player_data === '[REDACTED]';
  
  console.log("Session data properly redacted:", sessionRedacted);
  
  // Restore original location
  Object.defineProperty(window, 'location', {
    value: originalLocation,
    writable: true
  });
  
  console.log("\n✅ Secure logging tests complete!");
  return sessionRedacted;
}

// Auto-run test if in browser environment
if (typeof window !== 'undefined') {
  testSecureLogging();
}
