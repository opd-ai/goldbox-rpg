package server

import (
	"encoding/json"
	"reflect"
	"sync"
)

// DeltaState tracks the last known state per WebSocket connection for delta compression.
// It enables sending only changed fields to reduce bandwidth consumption.
//
// Fields:
//   - lastState: The last known complete state for this connection
//   - mu: Mutex protecting concurrent access to lastState
type DeltaState struct {
	lastState map[string]interface{}
	mu        sync.RWMutex
}

// NewDeltaState creates a new delta state tracker for a WebSocket connection.
func NewDeltaState() *DeltaState {
	return &DeltaState{
		lastState: make(map[string]interface{}),
	}
}

// GetLastState returns a copy of the last known state.
func (d *DeltaState) GetLastState() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.lastState == nil {
		return nil
	}

	copy := make(map[string]interface{}, len(d.lastState))
	for k, v := range d.lastState {
		copy[k] = v
	}
	return copy
}

// UpdateState stores the new state and returns the delta (changes only).
// If there was no previous state, returns the full state.
func (d *DeltaState) UpdateState(newState map[string]interface{}) map[string]interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.lastState == nil || len(d.lastState) == 0 {
		// First state update - send full state
		d.lastState = deepCopyMap(newState)
		return newState
	}

	// Calculate delta
	delta := CalculateDelta(d.lastState, newState)

	// Store new state
	d.lastState = deepCopyMap(newState)

	return delta
}

// Reset clears the stored state, forcing a full state send on next update.
func (d *DeltaState) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lastState = make(map[string]interface{})
}

// CalculateDelta computes the difference between old and new state maps.
// It returns a map containing only the fields that have changed or are new.
// Deleted fields are represented as null values.
//
// Parameters:
//   - oldState: The previous state map
//   - newState: The current state map
//
// Returns:
//   - map[string]interface{}: A delta containing only changed/new/deleted fields
func CalculateDelta(oldState, newState map[string]interface{}) map[string]interface{} {
	if oldState == nil {
		return newState
	}
	if newState == nil {
		// All fields deleted - return deletion markers
		delta := make(map[string]interface{}, len(oldState))
		for k := range oldState {
			delta[k] = nil
		}
		return delta
	}

	delta := make(map[string]interface{})

	// Check for new or changed fields
	for key, newVal := range newState {
		oldVal, exists := oldState[key]
		if !exists {
			// New field
			delta[key] = newVal
			continue
		}

		// Check if value has changed
		if !deepEqual(oldVal, newVal) {
			// Handle nested maps recursively
			if oldMap, ok := oldVal.(map[string]interface{}); ok {
				if newMap, ok := newVal.(map[string]interface{}); ok {
					nestedDelta := CalculateDelta(oldMap, newMap)
					if len(nestedDelta) > 0 {
						delta[key] = nestedDelta
					}
					continue
				}
			}
			delta[key] = newVal
		}
	}

	// Check for deleted fields
	for key := range oldState {
		if _, exists := newState[key]; !exists {
			delta[key] = nil // Mark as deleted
		}
	}

	return delta
}

// deepEqual compares two values for equality, handling maps and slices.
func deepEqual(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Use JSON serialization for reliable deep comparison
	aJSON, errA := json.Marshal(a)
	bJSON, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return reflect.DeepEqual(a, b)
	}
	return string(aJSON) == string(bJSON)
}

// deepCopyMap creates a deep copy of a map[string]interface{}.
func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}

	// Use JSON round-trip for reliable deep copy
	data, err := json.Marshal(src)
	if err != nil {
		// Fallback to shallow copy
		copy := make(map[string]interface{}, len(src))
		for k, v := range src {
			copy[k] = v
		}
		return copy
	}

	var copy map[string]interface{}
	if err := json.Unmarshal(data, &copy); err != nil {
		// Fallback to shallow copy
		copy = make(map[string]interface{}, len(src))
		for k, v := range src {
			copy[k] = v
		}
	}
	return copy
}

// DeltaMessage wraps a delta update message for WebSocket transmission.
// It includes metadata to help clients determine how to apply the update.
type DeltaMessage struct {
	// Type indicates the message type (e.g., "state_delta", "full_state")
	Type string `json:"type"`
	// IsDelta is true if this message contains only changed fields
	IsDelta bool `json:"is_delta"`
	// Sequence is a monotonically increasing counter for ordering
	Sequence uint64 `json:"sequence"`
	// Data contains the state or delta data
	Data map[string]interface{} `json:"data"`
}

// DeltaMessageSequence tracks message sequence numbers per connection.
type DeltaMessageSequence struct {
	counter uint64
	mu      sync.Mutex
}

// Next returns the next sequence number.
func (d *DeltaMessageSequence) Next() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.counter++
	return d.counter
}

// Current returns the current sequence number without incrementing.
func (d *DeltaMessageSequence) Current() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.counter
}

// EncodeDeltaMessage creates a DeltaMessage from state data.
func EncodeDeltaMessage(msgType string, isDelta bool, seq uint64, data map[string]interface{}) *DeltaMessage {
	return &DeltaMessage{
		Type:     msgType,
		IsDelta:  isDelta,
		Sequence: seq,
		Data:     data,
	}
}

// EstimateMessageSize returns an approximate size of the message in bytes.
// Useful for bandwidth estimation and compression decisions.
func EstimateMessageSize(data map[string]interface{}) int {
	bytes, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	return len(bytes)
}

// CompressionStats tracks compression effectiveness.
type CompressionStats struct {
	TotalMessages   uint64
	FullMessages    uint64
	DeltaMessages   uint64
	TotalBytesFull  uint64
	TotalBytesDelta uint64
	BytesSaved      uint64
	mu              sync.RWMutex
}

// NewCompressionStats creates a new compression statistics tracker.
func NewCompressionStats() *CompressionStats {
	return &CompressionStats{}
}

// RecordFullMessage records a full state message was sent.
func (c *CompressionStats) RecordFullMessage(size int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TotalMessages++
	c.FullMessages++
	c.TotalBytesFull += uint64(size)
}

// RecordDeltaMessage records a delta message was sent with savings.
func (c *CompressionStats) RecordDeltaMessage(deltaSize, fullSize int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TotalMessages++
	c.DeltaMessages++
	c.TotalBytesDelta += uint64(deltaSize)
	if fullSize > deltaSize {
		c.BytesSaved += uint64(fullSize - deltaSize)
	}
}

// GetStats returns a snapshot of compression statistics.
func (c *CompressionStats) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	savingsPercent := float64(0)
	totalBytes := c.TotalBytesFull + c.TotalBytesDelta
	if totalBytes > 0 && c.BytesSaved > 0 {
		savingsPercent = float64(c.BytesSaved) / float64(totalBytes+c.BytesSaved) * 100
	}

	return map[string]interface{}{
		"total_messages":    c.TotalMessages,
		"full_messages":     c.FullMessages,
		"delta_messages":    c.DeltaMessages,
		"total_bytes_full":  c.TotalBytesFull,
		"total_bytes_delta": c.TotalBytesDelta,
		"bytes_saved":       c.BytesSaved,
		"savings_percent":   savingsPercent,
	}
}

// Reset clears all compression statistics.
func (c *CompressionStats) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TotalMessages = 0
	c.FullMessages = 0
	c.DeltaMessages = 0
	c.TotalBytesFull = 0
	c.TotalBytesDelta = 0
	c.BytesSaved = 0
}
