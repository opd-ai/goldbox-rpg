package main

import (
	"encoding/json"
	"fmt"
	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/server"
)

func main() {
	// Create a server instance
	srv := server.NewRPCServer(":8080")

	// Set up a small world for testing
	srv.state.WorldState = game.NewWorldWithSize(5, 5, 10) // 5x5 world

	// Create a test player at position (2,2)
	player := &game.Player{
		Character: game.Character{
			ID:       "test_player",
			Name:     "Test Player",
			Position: game.Position{X: 2, Y: 2},
		},
	}

	// Create a session
	srv.sessions["test_session"] = &server.PlayerSession{
		SessionID: "test_session",
		Player:    player,
		Connected: true,
	}

	// Add player to world
	srv.state.WorldState.AddObject(player)

	// Test movement beyond boundaries
	fmt.Printf("Testing boundary enforcement in a 5x5 world (coordinates 0-4)\n")
	fmt.Printf("Player starting position: %+v\n", player.GetPosition())

	// Try to move beyond north boundary (Y from 2 to 3 to 4 to 5 - should be clamped at 4)
	for i := 0; i < 5; i++ {
		moveParams := map[string]interface{}{
			"session_id": "test_session",
			"direction":  "north",
		}
		paramsJSON, _ := json.Marshal(moveParams)

		result, err := srv.handleMove(paramsJSON)
		if err != nil {
			fmt.Printf("Move %d North failed: %v\n", i+1, err)
		} else {
			fmt.Printf("Move %d North succeeded: %+v\n", i+1, result)
		}
		fmt.Printf("Player position after move %d: %+v\n", i+1, player.GetPosition())
	}

	// Try to move beyond south boundary
	fmt.Printf("\nNow testing south boundary...\n")
	for i := 0; i < 10; i++ {
		moveParams := map[string]interface{}{
			"session_id": "test_session",
			"direction":  "south",
		}
		paramsJSON, _ := json.Marshal(moveParams)

		result, err := srv.handleMove(paramsJSON)
		if err != nil {
			fmt.Printf("Move %d South failed: %v\n", i+1, err)
		} else {
			fmt.Printf("Move %d South succeeded: %+v\n", i+1, result)
		}
		fmt.Printf("Player position after move %d: %+v\n", i+1, player.GetPosition())
	}
}
