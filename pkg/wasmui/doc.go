// Package wasmui provides an Ebitengine/WASM-based game UI client
// for the Gold Box RPG Engine.
//
// This package implements the client-side game interface using Ebitengine,
// a 2D game library for Go that supports WebAssembly (WASM) builds.
// The UI communicates with the existing JSON-RPC server via WebSocket.
//
// # Architecture
//
// The wasmui package consists of several components:
//
//   - Game: Main Ebitengine game implementation handling rendering and input
//   - RPCClient: WebSocket-based JSON-RPC 2.0 client for server communication
//   - Types: Shared type definitions for game state and UI elements
//
// # Building for WASM
//
// To build the WASM binary:
//
//	GOOS=js GOARCH=wasm go build -o web/static/js/game.wasm ./cmd/wasm-ui
//
// Then copy the JavaScript glue file:
//
//	cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/static/js/
//
// # Usage
//
// The WASM binary is loaded by the HTML wrapper (web/wasm.html) which
// initializes the WebAssembly environment and starts the game.
//
// The game connects to the server at /rpc/ws and uses the same JSON-RPC
// protocol as the existing TypeScript client.
//
// # UI Features
//
//   - Character panel: Display player stats, HP, attributes
//   - Combat log: Scrolling game/combat message history
//   - Direction controls: 8-way movement via buttons or keyboard
//   - Action buttons: Attack, Cast Spell, Use Item, End Turn
//   - Viewport: Grid-based game world display
//
// # Keyboard Controls
//
//   - WASD/Arrow keys: Cardinal movement (N/S/E/W)
//   - QEZC: Diagonal movement (NW/NE/SW/SE)
//   - Numpad 1-9: Full 8-directional movement
//   - Space: End turn
package wasmui
