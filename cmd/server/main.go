package main

import (
	"goldbox-rpg/pkg/server"
	"log"
	"net/http"
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

	// Start server on port 8080
	log.Printf("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
