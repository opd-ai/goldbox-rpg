package main

import (
"encoding/json"
"fmt"
)

// Simulate client's handleMessage behavior
type RPCResponse struct {
JSONRPC string      `json:"jsonrpc"`
Result  interface{} `json:"result,omitempty"`
Error   interface{} `json:"error,omitempty"`
ID      *int64      `json:"id"`
}

func isJSONRPCMessage(raw map[string]interface{}) bool {
v, ok := raw["jsonrpc"].(string)
return ok && v == "2.0"
}

func dispatchPendingResponse(resp *RPCResponse, pending map[int64]bool) bool {
if resp.ID == nil {
fmt.Println("Response has no ID (nil) - will NOT dispatch to pending")
return false
}

fmt.Printf("Response ID: %d\n", *resp.ID)
if _, ok := pending[*resp.ID]; ok {
fmt.Println("Found matching pending request!")
return true
}
fmt.Println("No pending request with this ID")
return false
}

func main() {
// Server sends session confirmation with id: 0
confirmationMsg := `{"jsonrpc":"2.0","result":{"session_id":"abc-123"},"id":0}`

fmt.Println("Server sends:", confirmationMsg)

var raw map[string]interface{}
json.Unmarshal([]byte(confirmationMsg), &raw)

if isJSONRPCMessage(raw) {
fmt.Println("Recognized as JSON-RPC message")

var resp RPCResponse
json.Unmarshal([]byte(confirmationMsg), &resp)

// Client's pending requests: start at 0, first Add(1) = 1
pending := make(map[int64]bool)
pending[1] = true // First request ID
pending[2] = true // Second request ID

fmt.Printf("Pending request IDs: %v\n", pending)
fmt.Println("\nTrying to dispatch session confirmation (id=0):")

dispatched := dispatchPendingResponse(&resp, pending)
if !dispatched {
fmt.Println("\nResult: Session confirmation will be treated as server notification")
if resp.Result != nil {
fmt.Printf("Notification callback would receive: %v\n", resp.Result)
}
}
}
}
