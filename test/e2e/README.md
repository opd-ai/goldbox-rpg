# End-to-End (E2E) Integration Tests

This directory contains end-to-end integration tests for the GoldBox RPG Engine. These tests validate complete user workflows and system behavior by running a real server instance and making HTTP/WebSocket API calls.

## Current Status

⚠️ **Work in Progress**: The E2E test framework has been implemented but requires additional configuration to properly start test server instances. The framework is ready for use once the server startup issues are resolved.

**Known Issues:**
- Test server instances not starting properly due to path/configuration issues
- Health check endpoints timing out
- Server log files not being written (process may not be starting)

**Next Steps:**
- Debug test server startup process
- Verify server binary can run with test configuration
- Add better error logging and diagnostics
- Test with simplified server configuration

## Overview

E2E tests differ from unit tests in that they:
- Start a real server instance with isolated test data
- Make actual HTTP and WebSocket connections
- Test complete workflows from start to finish
- Validate integration between all system components
- Test error scenarios and edge cases

## Test Structure

### Core Components

1. **client.go** - HTTP and WebSocket client for E2E tests
   - JSON-RPC 2.0 request/response handling
   - WebSocket connection management
   - Helper methods for common API calls

2. **server.go** - Test server lifecycle management
   - Starts isolated server instances for each test suite
   - Manages temporary data directories
   - Handles graceful shutdown and cleanup

3. **fixtures.go** - Test data and assertion helpers
   - Character fixtures and factories
   - Custom assertions for game-specific data
   - Test helper utilities

### Test Suites

- **session_test.go** - Session management workflows
  - Session creation and validation
  - Concurrent session handling
  - Session timeout behavior
  - Multi-client scenarios

- **character_test.go** - Character creation and management
  - Character creation with different classes
  - Attribute validation
  - Character state verification
  - Error handling

- **persistence_test.go** - Data persistence validation
  - Auto-save functionality
  - Multi-session persistence
  - File integrity checks

## Running E2E Tests

### Prerequisites

1. Build the server binary:
   ```bash
   make build
   ```

2. Ensure you have required dependencies:
   ```bash
   go mod download
   ```

### Run All E2E Tests

```bash
# Run from project root
go test ./test/e2e/... -v

# Or using the Makefile target
make test-e2e
```

### Run Specific Test Suite

```bash
# Run only session tests
go test ./test/e2e/ -v -run TestSession

# Run only character tests
go test ./test/e2e/ -v -run TestCharacter

# Run specific test case
go test ./test/e2e/ -v -run TestSessionWorkflow/join_game_creates_session
```

### Run with Race Detector

```bash
go test ./test/e2e/... -v -race
```

### Run with Coverage

```bash
go test ./test/e2e/... -v -coverprofile=e2e_coverage.out
go tool cover -html=e2e_coverage.out
```

## Test Configuration

E2E tests start server instances with the following configuration:

- **Port**: Random available port (to avoid conflicts)
- **Data Directory**: Temporary directory (cleaned up after tests)
- **Web Directory**: Temporary directory with minimal static files
- **Auto-save Interval**: 5 seconds (faster than production for testing)
- **Session Timeout**: 30 seconds
- **Dev Mode**: Enabled (allows all WebSocket origins)
- **Log Level**: Info

## Writing New E2E Tests

### Basic Test Structure

```go
func TestMyFeature(t *testing.T) {
    // Create test helper (starts server automatically)
    helper := NewTestHelper(t)
    defer helper.Cleanup()

    // Get client
    client := helper.Client()

    // Create session and character if needed
    sessionID, charID := helper.CreateSession()

    // Test your feature
    result, err := client.Call("my_method", map[string]interface{}{
        "session_id": sessionID,
        "param": "value",
    })
    require.NoError(t, err)
    
    // Make assertions
    assert.Equal(t, expectedValue, result["field"])
}
```

### Table-Driven Tests

```go
func TestMyFeatureScenarios(t *testing.T) {
    helper := NewTestHelper(t)
    defer helper.Cleanup()

    testCases := []struct {
        name          string
        input         string
        expectError   bool
        errorContains string
    }{
        {
            name:        "valid_input",
            input:       "test",
            expectError: false,
        },
        {
            name:          "invalid_input",
            input:         "",
            expectError:   true,
            errorContains: "required",
        },
    }

    client := helper.Client()

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test logic here
        })
    }
}
```

### Testing WebSocket Events

```go
func TestWebSocketEvent(t *testing.T) {
    helper := NewTestHelper(t)
    defer helper.Cleanup()

    client := helper.Client()

    // Connect to WebSocket
    err := client.ConnectWebSocket()
    require.NoError(t, err)
    defer client.CloseWebSocket()

    // Trigger action that generates event
    sessionID, _ := helper.CreateSession()

    // Wait for event
    event, err := client.WaitForEvent("player_joined", 5*time.Second)
    require.NoError(t, err)

    // Verify event structure
    AssertWebSocketEvent(t, event, "player_joined")
}
```

## Best Practices

1. **Isolation**: Each test should be independent and not rely on others
2. **Cleanup**: Always use `defer helper.Cleanup()` to clean up resources
3. **Timeouts**: Use reasonable timeouts for async operations
4. **Assertions**: Use the provided assertion helpers for consistent error messages
5. **Logging**: Use `t.Logf()` for debugging information
6. **Fixtures**: Use fixture functions to generate test data
7. **Error Messages**: Provide descriptive failure messages in assertions

## Debugging Failed Tests

### View Server Logs

```go
func TestMyFeature(t *testing.T) {
    helper := NewTestHelper(t)
    defer func() {
        // Print logs on failure
        if t.Failed() {
            logs, _ := helper.Server().GetLogContents()
            t.Logf("Server logs:\n%s", logs)
        }
        helper.Cleanup()
    }()
    
    // Test code...
}
```

### Increase Timeout for Debugging

```go
// Temporarily increase timeouts when debugging
event, err := client.WaitForEvent("my_event", 60*time.Second)
```

### Run Single Test

```bash
go test ./test/e2e/ -v -run TestMyFeature/my_specific_case
```

## CI Integration

E2E tests are integrated into the CI pipeline and run on:
- Every pull request
- Commits to main branch
- Before deployments

CI configuration: `.github/workflows/ci.yml`

## Test Coverage

E2E tests complement unit tests by:
- Validating API contracts
- Testing integration between components
- Verifying real-world workflows
- Catching issues that unit tests miss

Current E2E coverage focuses on:
- ✅ Session management
- ✅ Character creation
- ✅ Basic persistence
- 🚧 Combat system (future)
- 🚧 WebSocket events (future)
- 🚧 PCG integration (future)

## Future Enhancements

Planned improvements to E2E test suite:

1. **Combat Workflow Tests**
   - Complete combat rounds
   - Movement and positioning
   - Attack actions
   - Spell casting
   - Effect application

2. **WebSocket Integration Tests**
   - Real-time event broadcasting
   - Multiple client subscriptions
   - Event ordering and consistency

3. **PCG Integration Tests**
   - Content generation workflows
   - Deterministic seeding validation
   - Performance benchmarks

4. **Error Recovery Tests**
   - Network failure scenarios
   - Server restart recovery
   - Invalid state handling

5. **Performance Tests**
   - Load testing with multiple concurrent sessions
   - Stress testing system limits
   - Resource usage monitoring

## Troubleshooting

### Port Already in Use

E2E tests use random ports, but if you encounter port conflicts:
```bash
# Check for running test servers
ps aux | grep goldbox
kill <pid>
```

### Test Data Not Cleaned Up

If tests fail and leave data behind:
```bash
# Clean up temp directories
rm -rf /tmp/goldbox-e2e-*
```

### Build Errors

Ensure server binary is built:
```bash
make build
ls -la bin/server
```

### Timeout Errors

If tests timeout:
1. Check server logs for errors
2. Verify server started successfully
3. Check network connectivity
4. Increase timeout if needed for slow systems

## Contributing

When adding new E2E tests:

1. Follow existing patterns and conventions
2. Add documentation for complex test scenarios
3. Update this README with new test suites
4. Ensure tests pass locally before submitting PR
5. Verify tests pass in CI environment

## References

- [Testing Guide](../../docs/TESTING.md)
- [API Documentation](../../pkg/README-RPC.md)
- [Development Guide](../../CONTRIBUTING.md)
