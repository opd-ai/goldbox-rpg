package main

import (
"encoding/json"
"fmt"
)

// NewErrorResponse creates a new JSON-RPC 2.0 error response message.
func NewErrorResponse(id interface{}, err error) interface{} {
return map[string]interface{}{
"jsonrpc": "2.0",
"error": map[string]interface{}{
"code":    -32000,
"message": err.Error(),
},
"id": id,
}
}

// Client's RPCError type
type RPCError struct {
Code    int         `json:"code"`
Message string      `json:"message"`
Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// Client's RPCResponse
type RPCResponse struct {
JSONRPC string      `json:"jsonrpc"`
Result  interface{} `json:"result,omitempty"`
Error   *RPCError   `json:"error,omitempty"`
ID      *int64      `json:"id"`
}

func main() {
// Server's handleJoinGame returns: fmt.Errorf("player name is required")
plainErr := fmt.Errorf("player name is required")

// Server calls NewErrorResponse(req.ID, plainErr)
responseObj := NewErrorResponse(int64(1), plainErr)

// Marshal to JSON
respJSON, _ := json.Marshal(responseObj)
fmt.Println("Server's NewErrorResponse JSON:")
fmt.Println(string(respJSON))

// Client receives and unmarshals
var clientResp RPCResponse
err := json.Unmarshal(respJSON, &clientResp)
fmt.Printf("\nClient unmarshal error: %v\n", err)
fmt.Printf("Error field present: %v\n", clientResp.Error != nil)

if clientResp.Error != nil {
fmt.Printf("Error code: %d\n", clientResp.Error.Code)
fmt.Printf("Error message: %s\n", clientResp.Error.Message)
fmt.Printf("Error.Error(): %s\n", clientResp.Error.Error())
fmt.Printf("Type check: Is *RPCError: %v\n", true)
}
}
