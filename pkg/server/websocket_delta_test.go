package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateDelta_NoChanges(t *testing.T) {
	oldState := map[string]interface{}{
		"health": 100,
		"mana":   50,
		"name":   "Hero",
	}
	newState := map[string]interface{}{
		"health": 100,
		"mana":   50,
		"name":   "Hero",
	}

	delta := CalculateDelta(oldState, newState)
	assert.Empty(t, delta, "No changes should produce empty delta")
}

func TestCalculateDelta_SimpleChanges(t *testing.T) {
	oldState := map[string]interface{}{
		"health": 100,
		"mana":   50,
		"name":   "Hero",
	}
	newState := map[string]interface{}{
		"health": 80, // Changed
		"mana":   50,
		"name":   "Hero",
	}

	delta := CalculateDelta(oldState, newState)
	assert.Len(t, delta, 1, "Should have one change")
	assert.Equal(t, 80, delta["health"])
}

func TestCalculateDelta_NewFields(t *testing.T) {
	oldState := map[string]interface{}{
		"health": 100,
	}
	newState := map[string]interface{}{
		"health": 100,
		"mana":   50, // New field
	}

	delta := CalculateDelta(oldState, newState)
	assert.Len(t, delta, 1, "Should have one new field")
	assert.Equal(t, 50, delta["mana"])
}

func TestCalculateDelta_DeletedFields(t *testing.T) {
	oldState := map[string]interface{}{
		"health": 100,
		"mana":   50,
	}
	newState := map[string]interface{}{
		"health": 100,
		// mana deleted
	}

	delta := CalculateDelta(oldState, newState)
	assert.Len(t, delta, 1, "Should have one deleted field")
	assert.Nil(t, delta["mana"], "Deleted field should be nil")
}

func TestCalculateDelta_NestedChanges(t *testing.T) {
	oldState := map[string]interface{}{
		"player": map[string]interface{}{
			"health": 100,
			"mana":   50,
		},
	}
	newState := map[string]interface{}{
		"player": map[string]interface{}{
			"health": 80, // Changed
			"mana":   50,
		},
	}

	delta := CalculateDelta(oldState, newState)
	assert.Len(t, delta, 1, "Should have nested delta")

	playerDelta, ok := delta["player"].(map[string]interface{})
	require.True(t, ok, "player should be a map")
	assert.Equal(t, 80, playerDelta["health"])
}

func TestCalculateDelta_NilStates(t *testing.T) {
	t.Run("nil old state", func(t *testing.T) {
		newState := map[string]interface{}{"health": 100}
		delta := CalculateDelta(nil, newState)
		assert.Equal(t, newState, delta)
	})

	t.Run("nil new state", func(t *testing.T) {
		oldState := map[string]interface{}{"health": 100, "mana": 50}
		delta := CalculateDelta(oldState, nil)
		assert.Len(t, delta, 2)
		assert.Nil(t, delta["health"])
		assert.Nil(t, delta["mana"])
	})
}

func TestDeltaState_UpdateState(t *testing.T) {
	ds := NewDeltaState()

	// First update should return full state
	state1 := map[string]interface{}{"health": 100, "mana": 50}
	delta1 := ds.UpdateState(state1)
	assert.Equal(t, state1, delta1, "First update should return full state")

	// Second update should return only changes
	state2 := map[string]interface{}{"health": 80, "mana": 50}
	delta2 := ds.UpdateState(state2)
	assert.Len(t, delta2, 1, "Should only include changed field")
	assert.Equal(t, 80, delta2["health"])
}

func TestDeltaState_Reset(t *testing.T) {
	ds := NewDeltaState()

	// Add initial state
	ds.UpdateState(map[string]interface{}{"health": 100})

	// Reset
	ds.Reset()

	// Next update should return full state
	state := map[string]interface{}{"health": 80}
	delta := ds.UpdateState(state)
	assert.Equal(t, state, delta, "After reset, should return full state")
}

func TestCompressionStats(t *testing.T) {
	stats := NewCompressionStats()

	// Record full message
	stats.RecordFullMessage(1000)
	assert.Equal(t, uint64(1), stats.TotalMessages)
	assert.Equal(t, uint64(1), stats.FullMessages)

	// Record delta message
	stats.RecordDeltaMessage(200, 1000)
	assert.Equal(t, uint64(2), stats.TotalMessages)
	assert.Equal(t, uint64(1), stats.DeltaMessages)
	assert.Equal(t, uint64(800), stats.BytesSaved)

	// Get stats
	result := stats.GetStats()
	assert.Greater(t, result["savings_percent"].(float64), float64(0))
}

func TestEstimateMessageSize(t *testing.T) {
	data := map[string]interface{}{
		"health": 100,
		"mana":   50,
		"name":   "Hero",
	}

	size := EstimateMessageSize(data)
	assert.Greater(t, size, 0, "Size should be positive")

	// Verify against actual JSON size
	jsonBytes, _ := json.Marshal(data)
	assert.Equal(t, len(jsonBytes), size)
}

func TestDeltaMessageSequence(t *testing.T) {
	seq := &DeltaMessageSequence{}

	assert.Equal(t, uint64(0), seq.Current())
	assert.Equal(t, uint64(1), seq.Next())
	assert.Equal(t, uint64(1), seq.Current())
	assert.Equal(t, uint64(2), seq.Next())
	assert.Equal(t, uint64(2), seq.Current())
}

// Benchmark tests for delta compression

func BenchmarkCalculateDelta_SmallState(b *testing.B) {
	oldState := map[string]interface{}{
		"health": 100,
		"mana":   50,
		"x":      10,
		"y":      20,
	}
	newState := map[string]interface{}{
		"health": 80,
		"mana":   50,
		"x":      11,
		"y":      20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateDelta(oldState, newState)
	}
}

func BenchmarkCalculateDelta_LargeState(b *testing.B) {
	oldState := createLargeState(100)
	newState := createLargeState(100)
	// Make some changes
	newState["health"] = 80
	newState["field_50"] = "modified"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateDelta(oldState, newState)
	}
}

func BenchmarkWebSocketDelta_TypicalGameState(b *testing.B) {
	// Simulates typical game state updates
	ds := NewDeltaState()

	baseState := map[string]interface{}{
		"player": map[string]interface{}{
			"id":     "player-123",
			"name":   "Hero",
			"health": 100,
			"mana":   50,
			"x":      10,
			"y":      20,
			"level":  5,
			"exp":    1250,
		},
		"enemies": []interface{}{
			map[string]interface{}{"id": "e1", "health": 50, "x": 15, "y": 22},
			map[string]interface{}{"id": "e2", "health": 30, "x": 8, "y": 18},
		},
		"turn":   5,
		"combat": true,
	}

	// Initialize with base state
	ds.UpdateState(baseState)

	// Create update state (simulating a turn)
	updateState := deepCopyMap(baseState)
	updateState["player"].(map[string]interface{})["health"] = 85
	updateState["player"].(map[string]interface{})["x"] = 11
	updateState["turn"] = 6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ds.UpdateState(updateState)
		ds.Reset() // Reset to measure delta calculation consistently
		ds.UpdateState(baseState)
	}
}

func BenchmarkWebSocketDelta_MessageSizeComparison(b *testing.B) {
	ds := NewDeltaState()

	fullState := createLargeState(50)
	ds.UpdateState(fullState)

	// Simulate small change
	modifiedState := deepCopyMap(fullState)
	modifiedState["health"] = 80
	modifiedState["x"] = 15

	var fullSize, deltaSize int

	b.Run("FullState", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, _ := json.Marshal(fullState)
			fullSize = len(data)
		}
		b.ReportMetric(float64(fullSize), "bytes/msg")
	})

	b.Run("DeltaState", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			delta := CalculateDelta(fullState, modifiedState)
			data, _ := json.Marshal(delta)
			deltaSize = len(data)
		}
		b.ReportMetric(float64(deltaSize), "bytes/msg")
	})

	// Report savings
	if fullSize > 0 && deltaSize > 0 {
		savings := float64(fullSize-deltaSize) / float64(fullSize) * 100
		b.Logf("Delta compression savings: %.1f%% (full: %d bytes, delta: %d bytes)", savings, fullSize, deltaSize)
	}
}

func createLargeState(fields int) map[string]interface{} {
	state := make(map[string]interface{}, fields)
	state["health"] = 100
	state["mana"] = 50
	state["x"] = 10
	state["y"] = 20
	state["name"] = "TestPlayer"

	for i := 0; i < fields-5; i++ {
		state[string(rune('a'+i%26))+string(rune('0'+i/26))] = i * 10
	}
	return state
}
