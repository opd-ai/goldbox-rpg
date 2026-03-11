package server

import (
"context"
"encoding/json"
"fmt"
"testing"
"time"

"goldbox-rpg/pkg/game"
)

// BenchmarkJSONRPC_Parse benchmarks JSON-RPC request parsing
func BenchmarkJSONRPC_Parse(b *testing.B) {
reqJSON := []byte(`{
"jsonrpc": "2.0",
"method": "move",
"params": {"session_id": "test", "direction": 0},
"id": "bench"
}`)

b.ResetTimer()
for i := 0; i < b.N; i++ {
var req JSONRPCRequest
_ = json.Unmarshal(reqJSON, &req)
}
}

// BenchmarkJSONRPC_Marshal benchmarks JSON response serialization
func BenchmarkJSONRPC_Marshal(b *testing.B) {
response := map[string]interface{}{
"jsonrpc": "2.0",
"result": map[string]interface{}{
"success":  true,
"position": map[string]int{"x": 5, "y": 5},
},
"id": "bench",
}

b.ResetTimer()
for i := 0; i < b.N; i++ {
_, _ = json.Marshal(response)
}
}

// BenchmarkServer_SessionLookup benchmarks session retrieval
func BenchmarkServer_SessionLookup_Bench(b *testing.B) {
server, err := NewRPCServer(b.TempDir())
if err != nil {
b.Fatalf("Failed to create server: %v", err)
}
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
defer func() { _ = server.Shutdown(ctx) }()

// Create test sessions
for i := 0; i < 100; i++ {
sessionID := fmt.Sprintf("session_%d", i)
player := &game.Player{
Character: game.Character{
ID:   fmt.Sprintf("player_%d", i),
Name: "Test Player",
},
}
session := &PlayerSession{
SessionID:  sessionID,
Player:     player,
CreatedAt:  time.Now(),
LastActive: time.Now(),
}
server.mu.Lock()
server.sessions[sessionID] = session
server.mu.Unlock()
}

b.ResetTimer()
for i := 0; i < b.N; i++ {
sessionID := fmt.Sprintf("session_%d", i%100)
server.mu.RLock()
_ = server.sessions[sessionID]
server.mu.RUnlock()
}
}
