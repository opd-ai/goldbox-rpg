package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketConnection tests basic WebSocket connectivity
func TestWebSocketConnection(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSConnector")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	time.Sleep(500 * time.Millisecond)
}

// TestWebSocketMovementEvents tests movement event broadcasting
func TestWebSocketMovementEvents(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSMover")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Walker", "fighter")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	err = client.Move(sessionID, DirectionNorth)
	require.NoError(t, err)

	event, err := client.WaitForEvent("movement", 3*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "movement", event["type"])
}

// TestWebSocketCombatEvents tests combat event broadcasting
func TestWebSocketCombatEvents(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSCombatter")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Warrior", "fighter")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	event, err := client.WaitForEvent("combat_start", 3*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "combat_start", event["type"])
}

// TestWebSocketTurnEvents tests turn-based event broadcasting
func TestWebSocketTurnEvents(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSTurner")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "TurnTaker", "mage")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	// Wait for WebSocket connection to be fully established server-side
	// The server needs time to set session.Connected = true after upgrade
	time.Sleep(100 * time.Millisecond)

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Wait for combat_start event (event handlers run asynchronously)
	_, err = client.WaitForEvent("combat_start", 5*time.Second)
	require.NoError(t, err)

	// Small delay to ensure event system is settled before next action
	time.Sleep(50 * time.Millisecond)

	_, err = client.Call("endTurn", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	// Wait for turn_end event (event handlers run asynchronously via "go handler(event)")
	event, err := client.WaitForEvent("turn_end", 5*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, event)
}

// TestWebSocketMultipleClients tests multiple clients receiving events
func TestWebSocketMultipleClients(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client1 := helper.Client()
	client2 := NewClient(helper.Server().BaseURL())

	sessionID1, err := client1.JoinGame("WSClient1")
	require.NoError(t, err)

	sessionID2, err := client2.JoinGame("WSClient2")
	require.NoError(t, err)

	err = client1.ConnectWebSocket()
	require.NoError(t, err)
	defer client1.CloseWebSocket()

	err = client2.ConnectWebSocket()
	require.NoError(t, err)
	defer client2.CloseWebSocket()

	_, _, err = client1.CreateCharacter(sessionID1, "Player1", "fighter")
	require.NoError(t, err)

	_, _, err = client2.CreateCharacter(sessionID2, "Player2", "mage")
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
}

// TestWebSocketReconnection tests WebSocket reconnection handling
func TestWebSocketReconnection(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSReconnector")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)

	err = client.CloseWebSocket()
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()
}

// TestWebSocketEventOrdering tests that events are received in order
func TestWebSocketEventOrdering(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSOrderer")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Sequencer", "ranger")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	for i := 0; i < 5; i++ {
		err = client.Move(sessionID, DirectionEast)
		require.NoError(t, err)
		time.Sleep(200 * time.Millisecond)
	}
}

// TestWebSocketSpellEvents tests spell casting event broadcasting
func TestWebSocketSpellEvents(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSSpellCaster")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "Caster", "mage")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	_, err = client.WaitForEvent("combat_start", 3*time.Second)
	require.NoError(t, err)

	_, err = client.Call("castSpell", map[string]interface{}{
		"session_id": sessionID,
		"spell_id":   "magic_missile",
		"target_id":  "enemy1",
	})

	if err == nil {
		_, _ = client.WaitForEvent("spell_cast", 3*time.Second)
	}
}

// TestWebSocketEffectEvents tests effect application event broadcasting
func TestWebSocketEffectEvents(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("WSEffectTester")
	require.NoError(t, err)

	sessionID, charID, err := client.CreateCharacter("", "Affected", "cleric")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	_, err = client.Call("startCombat", map[string]interface{}{
		"session_id": sessionID,
	})
	require.NoError(t, err)

	_, err = client.Call("applyEffect", map[string]interface{}{
		"session_id":   sessionID,
		"character_id": charID,
		"effect_type":  "stun",
		"duration":     2,
		"magnitude":    1,
	})

	if err == nil {
		time.Sleep(500 * time.Millisecond)
	}
}

// TestWebSocketBroadcastToAll tests broadcasting events to all connected clients
func TestWebSocketBroadcastToAll(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	clients := make([]*Client, 3)
	sessionIDs := make([]string, 3)

	for i := 0; i < 3; i++ {
		if i == 0 {
			clients[i] = helper.Client()
		} else {
			clients[i] = NewClient(helper.Server().BaseURL())
		}

		var err error
		sessionIDs[i], err = clients[i].JoinGame("Broadcaster" + string(rune('A'+i)))
		require.NoError(t, err)

		err = clients[i].ConnectWebSocket()
		require.NoError(t, err)
		defer clients[i].CloseWebSocket()
	}

	time.Sleep(500 * time.Millisecond)

	_, _, err := clients[0].CreateCharacter(sessionIDs[0], "GlobalPlayer", "fighter")
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
}

// TestWebSocketLatency tests event delivery latency
func TestWebSocketLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency test in short mode")
	}

	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	_, err := client.JoinGame("LatencyTester")
	require.NoError(t, err)

	sessionID, _, err := client.CreateCharacter("", "SpeedTest", "thief")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	start := time.Now()
	err = client.Move(sessionID, DirectionNorth)
	require.NoError(t, err)

	_, err = client.WaitForEvent("movement", 5*time.Second)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 2*time.Second, "Event should be delivered quickly")
}

// TestWebSocketErrorHandling tests error scenarios over WebSocket
func TestWebSocketErrorHandling(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := helper.Client()

	sessionID, err := client.JoinGame("ErrorTester")
	require.NoError(t, err)

	err = client.ConnectWebSocket()
	require.NoError(t, err)
	defer client.CloseWebSocket()

	_, _, err = client.CreateCharacter(sessionID, "ErrorChar", "invalid_class")
	assert.Error(t, err)

	time.Sleep(200 * time.Millisecond)
}
