package server

import (
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
)

// TestMetrics_RecordWebSocketConnection tests WebSocket connection recording
func TestMetrics_RecordWebSocketConnection(t *testing.T) {
	metrics := NewMetrics()

	tests := []struct {
		name           string
		connectionType string
	}{
		{name: "record connected", connectionType: "connected"},
		{name: "record disconnected", connectionType: "disconnected"},
		{name: "record other type", connectionType: "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				metrics.RecordWebSocketConnection(tt.connectionType)
			})
		})
	}
}

// TestMetrics_RecordWebSocketMessage tests WebSocket message recording
func TestMetrics_RecordWebSocketMessage(t *testing.T) {
	metrics := NewMetrics()

	tests := []struct {
		name        string
		direction   string
		messageType string
	}{
		{name: "incoming text", direction: "incoming", messageType: "text"},
		{name: "outgoing binary", direction: "outgoing", messageType: "binary"},
		{name: "incoming rpc", direction: "incoming", messageType: "rpc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				metrics.RecordWebSocketMessage(tt.direction, tt.messageType)
			})
		})
	}
}

// TestMetrics_RecordPlayerAction tests player action recording
func TestMetrics_RecordPlayerAction(t *testing.T) {
	metrics := NewMetrics()

	tests := []struct {
		name       string
		actionType string
		status     string
	}{
		{name: "successful move", actionType: "move", status: "success"},
		{name: "failed attack", actionType: "attack", status: "failed"},
		{name: "successful cast", actionType: "cast_spell", status: "success"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				metrics.RecordPlayerAction(tt.actionType, tt.status)
			})
		})
	}
}

// TestMetrics_RecordGameEvent tests game event recording
func TestMetrics_RecordGameEvent(t *testing.T) {
	metrics := NewMetrics()

	tests := []struct {
		name      string
		eventType string
	}{
		{name: "combat event", eventType: "combat_start"},
		{name: "movement event", eventType: "player_move"},
		{name: "item event", eventType: "item_pickup"},
		{name: "quest event", eventType: "quest_complete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				metrics.RecordGameEvent(tt.eventType)
			})
		})
	}
}

// TestIsTimeToExecute_Coverage tests time-based execution check
func TestIsTimeToExecute_Coverage(t *testing.T) {
	tests := []struct {
		name         string
		triggerTicks int64
		currentTicks int64
		expected     bool
	}{
		{
			name:         "current time equals trigger",
			triggerTicks: 100,
			currentTicks: 100,
			expected:     true,
		},
		{
			name:         "current time past trigger",
			triggerTicks: 50,
			currentTicks: 100,
			expected:     true,
		},
		{
			name:         "current time before trigger",
			triggerTicks: 150,
			currentTicks: 100,
			expected:     false,
		},
		{
			name:         "zero trigger time",
			triggerTicks: 0,
			currentTicks: 100,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := game.GameTime{GameTicks: tt.currentTicks}
			trigger := game.GameTime{GameTicks: tt.triggerTicks}
			result := isTimeToExecute(current, trigger)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFindInventoryItem_Coverage tests inventory item lookup
func TestFindInventoryItem_Coverage(t *testing.T) {
	tests := []struct {
		name   string
		items  []game.Item
		itemID string
		found  bool
	}{
		{
			name: "item found",
			items: []game.Item{
				{ID: "sword", Name: "Sword"},
				{ID: "shield", Name: "Shield"},
				{ID: "potion", Name: "Potion"},
			},
			itemID: "shield",
			found:  true,
		},
		{
			name: "item not found",
			items: []game.Item{
				{ID: "sword", Name: "Sword"},
				{ID: "shield", Name: "Shield"},
			},
			itemID: "armor",
			found:  false,
		},
		{
			name:   "empty inventory",
			items:  []game.Item{},
			itemID: "sword",
			found:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findInventoryItem(tt.items, tt.itemID)
			if tt.found {
				assert.NotNil(t, result)
				assert.Equal(t, tt.itemID, result.ID)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
