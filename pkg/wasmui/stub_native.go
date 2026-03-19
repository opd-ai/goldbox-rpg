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

// SaveMapToDownload is a stub for native builds.
func (g *EditorGame) SaveMapToDownload() error {
	return fmt.Errorf("save only available in WASM builds")
}

// LoadMapFromFile is a stub for native builds.
func (g *EditorGame) LoadMapFromFile() {}

// NewMap is a stub for native builds.
func (g *EditorGame) NewMap(name string, width, height int) {}

// ExportToGameMap is a stub for native builds.
func (g *EditorGame) ExportToGameMap() ([]byte, error) {
	return nil, fmt.Errorf("export only available in WASM builds")
}

// QuestNode is a stub for native builds.
type QuestNode struct {
	ID          string
	Description string
	Required    int
	X           int
	Y           int
	Connections []string
}

// QuestEditorState is a stub for native builds.
type QuestEditorState struct {
	QuestID     string
	Title       string
	Description string
	Objectives  []*QuestNode
	Rewards     []QuestRewardEntry
}

// QuestRewardEntry is a stub for native builds.
type QuestRewardEntry struct {
	Type   string
	Value  int
	ItemID string
}

// QuestEditorGame is a stub for native builds.
type QuestEditorGame struct{}

// NewQuestEditorGame returns a stub quest editor for native builds.
func NewQuestEditorGame() *QuestEditorGame {
	return &QuestEditorGame{}
}

// NewQuestEditorGameWithClient returns a stub quest editor for native builds.
func NewQuestEditorGameWithClient(_ *RPCClient) *QuestEditorGame {
	return &QuestEditorGame{}
}

// QuestInputMode is a stub type for native builds.
type QuestInputMode int

const (
	// QuestModeNormal is the default mode.
	QuestModeNormal QuestInputMode = iota
	// QuestModeEditTitle is for editing the quest title.
	QuestModeEditTitle
	// QuestModeEditDescription is for editing node descriptions.
	QuestModeEditDescription
	// QuestModeConnect is for creating connections between nodes.
	QuestModeConnect
)
