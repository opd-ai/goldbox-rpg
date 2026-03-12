//go:build js && wasm

// Package main provides the Ebitengine/WASM-based game UI client.
// This client connects to the existing JSON-RPC server via WebSocket
// and renders the game interface using Ebitengine.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"goldbox-rpg/pkg/wasmui"
)

func main() {
	game, err := wasmui.NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	ebiten.SetWindowSize(wasmui.ScreenWidth, wasmui.ScreenHeight)
	ebiten.SetWindowTitle("Gold Box RPG")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
