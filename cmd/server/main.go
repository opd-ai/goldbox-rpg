package main

import (
	"goldbox-rpg/pkg/server"
	"log"
	"net"
	"os"
	"path/filepath"
)

func main() {
	// Get absolute path to web directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	webDir := filepath.Join(wd, "web")

	// Create new server instance
	server := server.NewRPCServer(webDir)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}

	// Start server on port 8080
	log.Printf("Starting server on :8080...")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
