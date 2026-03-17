package main

import (
"encoding/json"
"fmt"
)

type RPCRequestServer struct {
JSONRPC string                 `json:"jsonrpc"`
Method  string                 `json:"method"`
Params  map[string]interface{} `json:"params"`
ID      interface{}            `json:"id"` // Server accepts interface{}
}

type RPCResponseClient struct {
JSONRPC string      `json:"jsonrpc"`
Result  interface{} `json:"result,omitempty"`
Error   interface{} `json:"error,omitempty"`
ID      *int64      `json:"id"` // Client expects *int64
}

func main() {
// Simulate: Client sends ID as int64(1)
clientReq := map[string]interface{}{
"jsonrpc": "2.0",
"method":  "test",
"id":      int64(1),
}

// Marshal to JSON (simulating network transmission)
jsonBytes, _ := json.Marshal(clientReq)
fmt.Println("Client request JSON:", string(jsonBytes))

// Server receives and unmarshals to RPCRequestServer (ID as interface{})
var serverReq RPCRequestServer
json.Unmarshal(jsonBytes, &serverReq)
fmt.Printf("Server received ID type: %T, value: %v\n", serverReq.ID, serverReq.ID)

// Server sends response with the same ID (now possibly float64 after JSON round-trip)
serverResp := map[string]interface{}{
"jsonrpc": "2.0",
"result":  "test result",
"id":      serverReq.ID, // This is the ID as interface{} (could be float64)
}

// Marshal server response to JSON
respJSON, _ := json.Marshal(serverResp)
fmt.Println("Server response JSON:", string(respJSON))

// Client receives and unmarshals to RPCResponseClient (ID as *int64)
var clientResp RPCResponseClient
json.Unmarshal(respJSON, &clientResp)
fmt.Printf("Client received ID type: %T, value: %v\n", clientResp.ID, clientResp.ID)
if clientResp.ID != nil {
fmt.Printf("Client ID value is: %d\n", *clientResp.ID)
}

// Test: what if the server's interface{} contains float64?
fmt.Println("\n=== Testing float64 ID handling ===")
serverResp2 := map[string]interface{}{
"jsonrpc": "2.0",
"result":  "test result",
"id":      float64(1), // Simulate after JSON round-trip
}
respJSON2, _ := json.Marshal(serverResp2)
fmt.Println("Response with float64 ID:", string(respJSON2))

var clientResp2 RPCResponseClient
json.Unmarshal(respJSON2, &clientResp2)
fmt.Printf("Client received ID type: %T, value: %v\n", clientResp2.ID, clientResp2.ID)
if clientResp2.ID != nil {
fmt.Printf("Client ID value is: %d\n", *clientResp2.ID)
}
}
