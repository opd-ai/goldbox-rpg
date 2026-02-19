package e2e

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixtures provides test data and helper functions for E2E tests

// CharacterClass represents available character classes
var CharacterClasses = []string{
	"fighter",
	"mage",
	"cleric",
	"thief",
	"ranger",
	"paladin",
}

// CharacterNames provides sample character names for testing
var CharacterNames = []string{
	"Aldric",
	"Brianna",
	"Cedric",
	"Diana",
	"Eldrin",
	"Fiona",
	"Gareth",
	"Helena",
}

// Direction constants
const (
	DirectionNorth     = 0
	DirectionNorthEast = 1
	DirectionEast      = 2
	DirectionSouthEast = 3
	DirectionSouth     = 4
	DirectionSouthWest = 5
	DirectionWest      = 6
	DirectionNorthWest = 7
)

// RandomCharacterName returns a random character name
func RandomCharacterName() string {
	return CharacterNames[rand.Intn(len(CharacterNames))]
}

// RandomCharacterClass returns a random character class
func RandomCharacterClass() string {
	return CharacterClasses[rand.Intn(len(CharacterClasses))]
}

// AssertSessionID asserts that a session ID is valid
func AssertSessionID(t *testing.T, sessionID string) {
	require.NotEmpty(t, sessionID, "session ID should not be empty")
	require.Len(t, sessionID, 36, "session ID should be a UUID (36 characters)")
}

// AssertCharacterID asserts that a character ID is valid
func AssertCharacterID(t *testing.T, charID string) {
	require.NotEmpty(t, charID, "character ID should not be empty")
	require.Len(t, charID, 36, "character ID should be a UUID (36 characters)")
}

// AssertGameState asserts that a game state response is valid
func AssertGameState(t *testing.T, state map[string]interface{}) {
	require.NotNil(t, state, "game state should not be nil")

	// Check for expected fields
	assert.Contains(t, state, "world", "game state should contain world")
	assert.Contains(t, state, "player", "game state should contain player")
	assert.Contains(t, state, "turn", "game state should contain turn")
}

// AssertCharacterState asserts that a character state is valid
func AssertCharacterState(t *testing.T, char map[string]interface{}, expectedName, expectedClass string) {
	require.NotNil(t, char, "character should not be nil")

	// Check ID
	charID, ok := char["id"].(string)
	require.True(t, ok, "character should have string ID")
	AssertCharacterID(t, charID)

	// Check name
	name, ok := char["name"].(string)
	require.True(t, ok, "character should have string name")
	if expectedName != "" {
		assert.Equal(t, expectedName, name, "character name should match")
	}

	// Check class
	class, ok := char["class"].(string)
	require.True(t, ok, "character should have string class")
	if expectedClass != "" {
		assert.Equal(t, expectedClass, class, "character class should match")
	}

	// Check attributes
	assert.Contains(t, char, "strength", "character should have strength")
	assert.Contains(t, char, "dexterity", "character should have dexterity")
	assert.Contains(t, char, "constitution", "character should have constitution")
	assert.Contains(t, char, "intelligence", "character should have intelligence")
	assert.Contains(t, char, "wisdom", "character should have wisdom")
	assert.Contains(t, char, "charisma", "character should have charisma")

	// Check HP
	assert.Contains(t, char, "current_hp", "character should have current HP")
	assert.Contains(t, char, "max_hp", "character should have max HP")

	currentHP, ok := char["current_hp"].(float64)
	require.True(t, ok, "current HP should be a number")
	assert.Greater(t, currentHP, float64(0), "current HP should be positive")

	maxHP, ok := char["max_hp"].(float64)
	require.True(t, ok, "max HP should be a number")
	assert.Greater(t, maxHP, float64(0), "max HP should be positive")

	// Check level
	assert.Contains(t, char, "level", "character should have level")
	level, ok := char["level"].(float64)
	require.True(t, ok, "level should be a number")
	assert.GreaterOrEqual(t, level, float64(1), "level should be at least 1")
}

// AssertPosition asserts that a position is valid
func AssertPosition(t *testing.T, pos map[string]interface{}, expectedX, expectedY int) {
	require.NotNil(t, pos, "position should not be nil")

	x, ok := pos["x"].(float64)
	require.True(t, ok, "position should have numeric x")

	y, ok := pos["y"].(float64)
	require.True(t, ok, "position should have numeric y")

	if expectedX >= 0 {
		assert.Equal(t, float64(expectedX), x, "x coordinate should match")
	}

	if expectedY >= 0 {
		assert.Equal(t, float64(expectedY), y, "y coordinate should match")
	}
}

// AssertWebSocketEvent asserts that a WebSocket event has expected structure
func AssertWebSocketEvent(t *testing.T, event map[string]interface{}, expectedType string) {
	require.NotNil(t, event, "event should not be nil")

	eventType, ok := event["type"].(string)
	require.True(t, ok, "event should have string type")

	if expectedType != "" {
		assert.Equal(t, expectedType, eventType, "event type should match")
	}

	assert.Contains(t, event, "data", "event should have data field")
	assert.Contains(t, event, "timestamp", "event should have timestamp")
}

// CreateTestSession creates a test session and character
func CreateTestSession(t *testing.T, client *Client) (sessionID, charID string) {
	// Join game
	var err error
	sessionID, err = client.JoinGame(RandomCharacterName())
	require.NoError(t, err, "should join game successfully")
	AssertSessionID(t, sessionID)

	// Create character
	charID, err = client.CreateCharacter(sessionID, RandomCharacterName(), RandomCharacterClass())
	require.NoError(t, err, "should create character successfully")
	AssertCharacterID(t, charID)

	return sessionID, charID
}

// WaitForServerStart waits for server to start and returns a client
func WaitForServerStart(t *testing.T, server *TestServer) *Client {
	client := NewClient(server.BaseURL())
	err := client.WaitForHealth(30 * time.Second)
	require.NoError(t, err, "server should be healthy")
	return client
}

// TestHelper provides common test setup and teardown
type TestHelper struct {
	t      *testing.T
	server *TestServer
	client *Client
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	// Create and start test server
	server, err := NewTestServer()
	require.NoError(t, err, "should create test server")

	err = server.Start()
	require.NoError(t, err, "should start test server")

	// Create client
	client := NewClient(server.BaseURL())

	return &TestHelper{
		t:      t,
		server: server,
		client: client,
	}
}

// Cleanup cleans up test resources
func (th *TestHelper) Cleanup() {
	if th.client != nil {
		th.client.Close()
	}
	if th.server != nil {
		th.server.Stop()
	}
}

// Server returns the test server
func (th *TestHelper) Server() *TestServer {
	return th.server
}

// Client returns the test client
func (th *TestHelper) Client() *Client {
	return th.client
}

// CreateSession creates a test session with a character
func (th *TestHelper) CreateSession() (sessionID, charID string) {
	return CreateTestSession(th.t, th.client)
}

// ErrorContains asserts that an error contains a specific message
func ErrorContains(t *testing.T, err error, contains string) {
	require.Error(t, err, "expected an error")
	assert.Contains(t, err.Error(), contains, fmt.Sprintf("error should contain '%s'", contains))
}
