//go:build js && wasm

// Package main provides a static WASM demo of the Gold Box RPG Ebitengine client.
//
// This entry point is designed for GitHub Pages deployment where no backend
// server is available. It uses wasmui.NewGame() which gracefully handles the
// disconnected state — the WebSocket connection will fail, and the UI renders
// in offline mode showing the game interface with a "Disconnected" status indicator.
//
// Approach: Offline-tolerant wrapper (Strategy A from the issue).
// The existing wasmui code already sets connected=false on WS failure and
// displays the full game UI (viewport grid, character panel, combat log,
// action buttons) regardless of connection state.
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
	ebiten.SetWindowTitle("Gold Box RPG — Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
