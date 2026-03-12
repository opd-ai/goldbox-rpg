//go:build !js || !wasm

// Package wasmui provides stubs for native builds.
// The actual implementation is in game.go and rpc_client_wasm.go,
// which are only compiled for WASM targets.
package wasmui

import (
	"fmt"
)

// Game is a stub for native builds.
// The actual implementation is only available in WASM builds.
type Game struct{}

// NewGame returns an error on native builds.
// Use WASM build for actual game functionality.
func NewGame() (*Game, error) {
	return nil, fmt.Errorf("wasmui package only supports WASM builds (GOOS=js GOARCH=wasm)")
}

// RPCClient is a stub for native builds.
type RPCClient struct{}

// NewRPCClient returns a stub client for native builds.
func NewRPCClient() *RPCClient {
	return &RPCClient{}
}

// EditorGame is a stub for native builds.
// The actual implementation is only available in WASM builds.
type EditorGame struct{}

// NewEditorGame returns a stub editor for native builds.
func NewEditorGame() *EditorGame {
	return &EditorGame{}
}
