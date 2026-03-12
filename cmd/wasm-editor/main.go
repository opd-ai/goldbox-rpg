//go:build js && wasm

// Package main provides the Ebitengine/WASM-based map editor client.
// This client provides a browser-based GUI for creating and editing
// game maps using the Gold Box RPG tile system.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"goldbox-rpg/pkg/wasmui"
)

func main() {
	editor := wasmui.NewEditorGame()

	ebiten.SetWindowSize(wasmui.EditorScreenWidth, wasmui.EditorScreenHeight)
	ebiten.SetWindowTitle("Gold Box RPG - Map Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(editor); err != nil {
		log.Fatal(err)
	}
}
