// Package wasmui provides editor state management for the map editor.
// This file contains platform-independent types and logic for editor state.
package wasmui

import "errors"

// ErrNoMapLoaded is returned when an operation requires a loaded map but none exists.
var ErrNoMapLoaded = errors.New("no map loaded")

// EditorTool represents the currently selected editing tool.
type EditorTool int

const (
	// ToolPaint places tiles on the map.
	ToolPaint EditorTool = iota
	// ToolErase resets tiles to default (walkable, transparent).
	ToolErase
	// ToolFill flood-fills a region with the selected tile.
	ToolFill
	// ToolSelect selects a rectangular region for batch operations.
	ToolSelect
)

// String returns the display name for an editor tool.
func (t EditorTool) String() string {
	switch t {
	case ToolPaint:
		return "Paint"
	case ToolErase:
		return "Erase"
	case ToolFill:
		return "Fill"
	case ToolSelect:
		return "Select"
	default:
		return "Unknown"
	}
}

// TerrainEntry represents a selectable terrain type in the palette.
type TerrainEntry struct {
	Name        string `json:"name"`
	SpriteX     int    `json:"sprite_x"`
	SpriteY     int    `json:"sprite_y"`
	Walkable    bool   `json:"walkable"`
	Transparent bool   `json:"transparent"`
}

// DefaultTerrainPalette returns the built-in terrain palette for the editor.
func DefaultTerrainPalette() []TerrainEntry {
	return []TerrainEntry{
		{Name: "Grass", SpriteX: 0, SpriteY: 0, Walkable: true, Transparent: true},
		{Name: "Stone", SpriteX: 1, SpriteY: 0, Walkable: true, Transparent: true},
		{Name: "Water", SpriteX: 2, SpriteY: 0, Walkable: false, Transparent: true},
		{Name: "Wall", SpriteX: 3, SpriteY: 0, Walkable: false, Transparent: false},
		{Name: "Door", SpriteX: 4, SpriteY: 0, Walkable: true, Transparent: false},
		{Name: "Sand", SpriteX: 0, SpriteY: 1, Walkable: true, Transparent: true},
		{Name: "Dirt", SpriteX: 1, SpriteY: 1, Walkable: true, Transparent: true},
		{Name: "Tree", SpriteX: 2, SpriteY: 1, Walkable: false, Transparent: false},
		{Name: "Lava", SpriteX: 3, SpriteY: 1, Walkable: false, Transparent: true},
		{Name: "Ice", SpriteX: 4, SpriteY: 1, Walkable: true, Transparent: true},
	}
}

// EditorMapState holds the current state of the map being edited.
type EditorMapState struct {
	MapID  string      `json:"map_id"`
	Name   string      `json:"name"`
	Width  int         `json:"width"`
	Height int         `json:"height"`
	Tiles  [][]TileRef `json:"tiles"`
}

// TileRef represents a tile's configurable properties for the editor.
type TileRef struct {
	SpriteX     int  `json:"sprite_x"`
	SpriteY     int  `json:"sprite_y"`
	Walkable    bool `json:"walkable"`
	Transparent bool `json:"transparent"`
}

// NewEditorMapState creates a new empty map with the given dimensions.
func NewEditorMapState(name string, width, height int) *EditorMapState {
	tiles := make([][]TileRef, height)
	for y := range tiles {
		tiles[y] = make([]TileRef, width)
		for x := range tiles[y] {
			tiles[y][x] = TileRef{
				SpriteX:     0,
				SpriteY:     0,
				Walkable:    true,
				Transparent: true,
			}
		}
	}

	return &EditorMapState{
		Name:   name,
		Width:  width,
		Height: height,
		Tiles:  tiles,
	}
}

// GetTile returns the tile at the given coordinates, or nil if out of bounds.
func (m *EditorMapState) GetTile(x, y int) *TileRef {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return nil
	}
	return &m.Tiles[y][x]
}

// SetTile sets the tile at the given coordinates. Returns false if out of bounds.
func (m *EditorMapState) SetTile(x, y int, tile TileRef) bool {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return false
	}
	m.Tiles[y][x] = tile
	return true
}

// UndoEntry represents a single undoable tile change.
type UndoEntry struct {
	X       int
	Y       int
	OldTile TileRef
	NewTile TileRef
}

// UndoStack manages undo/redo state for the editor.
type UndoStack struct {
	stack    []UndoEntry
	position int
	maxSize  int
}

// NewUndoStack creates a new undo stack with the given maximum size.
func NewUndoStack(maxSize int) *UndoStack {
	return &UndoStack{
		stack:    make([]UndoEntry, 0, maxSize),
		position: -1,
		maxSize:  maxSize,
	}
}

// Push adds a new undo entry, discarding any future redo entries.
func (u *UndoStack) Push(entry UndoEntry) {
	// Discard entries after current position (redo history)
	if u.position < len(u.stack)-1 {
		u.stack = u.stack[:u.position+1]
	}

	// Enforce max size
	if len(u.stack) >= u.maxSize {
		u.stack = u.stack[1:]
		u.position--
	}

	u.stack = append(u.stack, entry)
	u.position = len(u.stack) - 1
}

// Undo returns the entry to undo, or nil if nothing to undo.
func (u *UndoStack) Undo() *UndoEntry {
	if u.position < 0 {
		return nil
	}
	entry := &u.stack[u.position]
	u.position--
	return entry
}

// Redo returns the entry to redo, or nil if nothing to redo.
func (u *UndoStack) Redo() *UndoEntry {
	if u.position >= len(u.stack)-1 {
		return nil
	}
	u.position++
	return &u.stack[u.position]
}

// CanUndo returns true if there are entries to undo.
func (u *UndoStack) CanUndo() bool {
	return u.position >= 0
}

// CanRedo returns true if there are entries to redo.
func (u *UndoStack) CanRedo() bool {
	return u.position < len(u.stack)-1
}

// Len returns the total number of undo entries.
func (u *UndoStack) Len() int {
	return len(u.stack)
}
