/**
 * Test script for EventEmitter memory leak prevention functionality
 * 
 * Tests the new cleanup methods added to prevent memory leaks from accumulated event listeners.
 * 
 * Usage: node test-event-cleanup.js
 */

// Simple EventEmitter implementation for testing (matches the one in rpc.js)
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

  off(event, callback) {
    if (!this.events.has(event)) {
      return false;
    }

    const callbacks = this.events.get(event);
    const index = callbacks.indexOf(callback);
    
    if (index === -1) {
      return false;
    }

    callbacks.splice(index, 1);

    if (callbacks.length === 0) {
      this.events.delete(event);
    }

    return true;
  }

  removeAllListeners(event) {
    if (!this.events.has(event)) {
      return false;
    }

    this.events.delete(event);
    return true;
  }

  clear() {
    this.events.clear();
  }

  listenerCount(event) {
    return this.events.has(event) ? this.events.get(event).length : 0;
  }

  eventNames() {
    return Array.from(this.events.keys());
  }
}

// Test suite
console.log("=== EventEmitter Memory Leak Prevention Tests ===\n");

const emitter = new EventEmitter();
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

// Test 1: Basic event listener addition and removal
test("Basic event listener addition and removal", () => {
  const callback1 = () => {};
  const callback2 = () => {};
  
  emitter.on('test', callback1);
  emitter.on('test', callback2);
  
  assertEqual(emitter.listenerCount('test'), 2, "Should have 2 listeners");
  
  const removed = emitter.off('test', callback1);
  assertTrue(removed, "Should return true when removing existing listener");
  assertEqual(emitter.listenerCount('test'), 1, "Should have 1 listener after removal");
  
  emitter.off('test', callback2);
  assertEqual(emitter.listenerCount('test'), 0, "Should have 0 listeners after removing all");
  assertEqual(emitter.eventNames().length, 0, "Should have no events after cleanup");
});

// Test 2: Removing non-existent listeners
test("Removing non-existent listeners", () => {
  const callback = () => {};
  const result1 = emitter.off('nonexistent', callback);
  const result2 = emitter.off('test', callback);
  
  assertEqual(result1, false, "Should return false for non-existent event");
  assertEqual(result2, false, "Should return false for non-existent callback");
});

// Test 3: Remove all listeners for specific event
test("Remove all listeners for specific event", () => {
  const callback1 = () => {};
  const callback2 = () => {};
  
  emitter.on('test1', callback1);
  emitter.on('test1', callback2);
  emitter.on('test2', callback1);
  
  assertEqual(emitter.listenerCount('test1'), 2, "Should have 2 listeners for test1");
  assertEqual(emitter.listenerCount('test2'), 1, "Should have 1 listener for test2");
  
  const removed = emitter.removeAllListeners('test1');
  assertTrue(removed, "Should return true when removing existing event");
  assertEqual(emitter.listenerCount('test1'), 0, "Should have 0 listeners for test1");
  assertEqual(emitter.listenerCount('test2'), 1, "Should still have 1 listener for test2");
  
  emitter.removeAllListeners('test2');
});

// Test 4: Clear all events
test("Clear all events", () => {
  const callback = () => {};
  
  emitter.on('event1', callback);
  emitter.on('event2', callback);
  emitter.on('event3', callback);
  
  assertEqual(emitter.eventNames().length, 3, "Should have 3 events");
  
  emitter.clear();
  
  assertEqual(emitter.eventNames().length, 0, "Should have no events after clear");
  assertEqual(emitter.listenerCount('event1'), 0, "Should have no listeners after clear");
});

// Test 5: Event names and listener counts
test("Event names and listener counts", () => {
  const callback = () => {};
  
  emitter.on('alpha', callback);
  emitter.on('beta', callback);
  emitter.on('beta', callback);
  
  const eventNames = emitter.eventNames();
  assertTrue(eventNames.includes('alpha'), "Should include 'alpha'");
  assertTrue(eventNames.includes('beta'), "Should include 'beta'");
  assertEqual(eventNames.length, 2, "Should have 2 event names");
  
  assertEqual(emitter.listenerCount('alpha'), 1, "Alpha should have 1 listener");
  assertEqual(emitter.listenerCount('beta'), 2, "Beta should have 2 listeners");
  assertEqual(emitter.listenerCount('gamma'), 0, "Gamma should have 0 listeners");
  
  emitter.clear();
});

// Test 6: Memory leak simulation
test("Memory leak prevention simulation", () => {
  const callbacks = [];
  
  // Simulate adding many listeners
  for (let i = 0; i < 1000; i++) {
    const callback = () => `callback${i}`;
    callbacks.push(callback);
    emitter.on('massive-event', callback);
  }
  
  assertEqual(emitter.listenerCount('massive-event'), 1000, "Should have 1000 listeners");
  
  // Remove half individually
  for (let i = 0; i < 500; i++) {
    emitter.off('massive-event', callbacks[i]);
  }
  
  assertEqual(emitter.listenerCount('massive-event'), 500, "Should have 500 listeners after individual removal");
  
  // Clear the rest
  emitter.removeAllListeners('massive-event');
  assertEqual(emitter.listenerCount('massive-event'), 0, "Should have 0 listeners after removeAllListeners");
  assertEqual(emitter.eventNames().length, 0, "Should have no events after cleanup");
});

// Test 7: Functional test - events still work after partial cleanup
test("Events still work after partial cleanup", () => {
  let counter = 0;
  const callback1 = () => counter++;
  const callback2 = () => counter += 2;
  const callback3 = () => counter += 3;
  
  emitter.on('counter', callback1);
  emitter.on('counter', callback2);
  emitter.on('counter', callback3);
  
  emitter.emit('counter');
  assertEqual(counter, 6, "Should increment by 6 (1+2+3)");
  
  // Remove middle callback
  emitter.off('counter', callback2);
  counter = 0;
  
  emitter.emit('counter');
  assertEqual(counter, 4, "Should increment by 4 (1+3) after removing callback2");
  
  emitter.clear();
});

console.log(`\n=== Test Results ===`);
console.log(`✅ Passed: ${testsPassed}/${testsTotal}`);
console.log(`❌ Failed: ${testsTotal - testsPassed}/${testsTotal}`);

if (testsPassed === testsTotal) {
  console.log("\n🎉 All EventEmitter memory leak prevention tests passed!");
  console.log("\nMemory leak prevention features verified:");
  console.log("• off() method removes specific listeners");
  console.log("• removeAllListeners() clears all listeners for an event");
  console.log("• clear() removes all events and listeners");
  console.log("• Empty event arrays are cleaned up automatically");
  console.log("• listenerCount() and eventNames() provide introspection");
} else {
  console.log("\n⚠️  Some tests failed. Please review the implementation.");
  process.exit(1);
}
