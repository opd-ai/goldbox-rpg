// Package e2e provides end-to-end testing infrastructure for the GoldBox RPG server.
//
// This package contains test utilities, fixtures, and integration tests that verify
// complete server functionality through HTTP and WebSocket interfaces.
//
// # Test Client
//
// Client provides a comprehensive test client for JSON-RPC and WebSocket testing:
//
//	client := e2e.NewClient("http://localhost:8080")
//	defer client.Close()
//
//	// JSON-RPC call
//	resp, err := client.Call("createCharacter", map[string]interface{}{
//	    "name": "TestHero",
//	    "class": "fighter",
//	})
//
//	// WebSocket connection
//	err := client.ConnectWebSocket()
//	msg := <-client.WebSocketMessages()
//
// # Test Server
//
// ServerFixture manages test server lifecycle:
//
//	fixture := e2e.NewServerFixture()
//	err := fixture.Start()
//	defer fixture.Stop()
//
//	client := e2e.NewClient(fixture.BaseURL())
//
// # Test Fixtures
//
// Pre-configured test data for common scenarios:
//   - Character creation and progression
//   - Combat encounters and spellcasting
//   - Inventory management and equipment
//   - Quest acceptance and completion
//
// # Direction Constants
//
// Movement direction constants for position testing:
//
//	const (
//	    DirectionNorth = 0
//	    DirectionEast  = 1
//	    DirectionSouth = 2
//	    DirectionWest  = 3
//	)
//
// # Running Tests
//
// Execute E2E tests from the repository root:
//
//	go test ./test/e2e/... -v
//
// Tests require a built server binary and may spin up actual server instances.
package e2e
