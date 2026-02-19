// Package main provides a demonstration CLI application for the level generation system.
//
// This demo showcases the [goldbox-rpg/pkg/pcg/levels.RoomCorridorGenerator] functionality
// and serves as executable documentation for developers integrating procedural level
// generation into their applications.
//
// # Usage
//
// Run the demo directly:
//
//	go run ./pkg/pcg/levels/demo
//
// Or build and execute:
//
//	go build -o level-demo ./pkg/pcg/levels/demo
//	./level-demo
//
// # Output
//
// The demo generates a dungeon level with configurable parameters and displays:
//   - Level name and dimensions
//   - Generation properties (room count, difficulty, etc.)
//   - ASCII visualization of the first 20x20 tiles
//
// # Configuration
//
// The demo uses the following default parameters which can be modified in main.go:
//   - Seed: 42 (for reproducible generation)
//   - Difficulty: 8 (medium-high challenge)
//   - PlayerLevel: 10 (mid-game progression)
//   - MinRooms: 4, MaxRooms: 7
//   - RoomTypes: Combat, Treasure, Puzzle
//   - CorridorStyle: Straight corridors
//   - LevelTheme: Classic dungeon
//   - HasBoss: true (boss room included)
//   - SecretRooms: 1
//
// # Example Output
//
//	Generated Level: Classic Level
//	Dimensions: 60x60
//	Properties: map[has_boss:true secret_rooms:1]
//
//	Level Map (first 20x20 section):
//	####################
//	#..................#
//	#..................#
//	#....##########....#
//	...
//
// See [goldbox-rpg/pkg/pcg/levels] for the full level generation API.
package main
