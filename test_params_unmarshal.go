package main

import (
"encoding/json"
"fmt"
)

type ServerRPCRequest struct {
JSONRPC string                 `json:"jsonrpc"`
Method  string                 `json:"method"`
Params  map[string]interface{} `json:"params"`
ID      interface{}            `json:"id"`
}

type ClientRPCRequest struct {
JSONRPC string      `json:"jsonrpc"`
Method  string      `json:"method"`
Params  interface{} `json:"params,omitempty"`
ID      int64       `json:"id"`
}

func main() {
fmt.Println("=== Test 1: Nil params enriched with session_id ===")
clientParams := map[string]interface{}{"session_id": "session-123"}

req := map[string]interface{}{
"jsonrpc": "2.0",
"method":  "getGameState",
"params":  clientParams,
"id":      int64(1),
}

jsonBytes, _ := json.Marshal(req)
fmt.Println("JSON:", string(jsonBytes))

var serverReq ServerRPCRequest
json.Unmarshal(jsonBytes, &serverReq)
fmt.Printf("Params type: %T\n", serverReq.Params)
fmt.Printf("Params: %v\n", serverReq.Params)
fmt.Printf("session_id: %v\n", serverReq.Params["session_id"])

fmt.Println("\n=== Test 2: Client RPC request structure ===")
clientReq := ClientRPCRequest{
JSONRPC: "2.0",
Method:  "getGameState",
Params:  map[string]interface{}{"session_id": "session-123"},
ID:      1,
}

clientJSON, _ := json.Marshal(clientReq)
fmt.Println("Client sends:", string(clientJSON))

var parsed ServerRPCRequest
json.Unmarshal(clientJSON, &parsed)
fmt.Printf("Server receives Params: %v\n", parsed.Params)

fmt.Println("\n=== Test 3: Can json.Unmarshal float64 into *int64? ===")
respJSON := `{"jsonrpc":"2.0","result":"ok","id":1.0}`
var resp struct {
JSONRPC string `json:"jsonrpc"`
Result  string `json:"result"`
ID      *int64 `json:"id"`
}
err := json.Unmarshal([]byte(respJSON), &resp)
fmt.Printf("Error: %v\n", err)
if resp.ID != nil {
fmt.Printf("ID value: %d\n", *resp.ID)
}
}
