package main

import (
	"log"
	"net/http"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
)

func main() {
	// Create a new RPC server
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")

	// Register the GameService
	s.RegisterService(new(GameService), "")

	// Handle RPC endpoint
	http.Handle("/rpc", s)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

type GameService struct{}

func (s *GameService) Ping(r *http.Request, args *struct{}, reply *string) error {
	*reply = "PONG"
	return nil
}
