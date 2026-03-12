package wasmui

import (
	"testing"
)

func TestEditorToolString(t *testing.T) {
	tests := []struct {
		name string
		tool EditorTool
		want string
	}{
		{name: "paint tool", tool: ToolPaint, want: "Paint"},
		{name: "erase tool", tool: ToolErase, want: "Erase"},
		{name: "fill tool", tool: ToolFill, want: "Fill"},
		{name: "select tool", tool: ToolSelect, want: "Select"},
		{name: "unknown tool", tool: EditorTool(99), want: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tool.String(); got != tt.want {
				t.Errorf("EditorTool.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultTerrainPalette(t *testing.T) {
	palette := DefaultTerrainPalette()
	if len(palette) == 0 {
		t.Fatal("DefaultTerrainPalette() returned empty palette")
	}

	// Check first entry
	if palette[0].Name != "Grass" {
		t.Errorf("first terrain name = %q, want %q", palette[0].Name, "Grass")
	}

	// Ensure all entries have names
	for i, entry := range palette {
		if entry.Name == "" {
			t.Errorf("palette[%d] has empty name", i)
		}
	}

	// Check that at least one non-walkable entry exists
	hasNonWalkable := false
	for _, entry := range palette {
		if !entry.Walkable {
			hasNonWalkable = true
			break
		}
	}
	if !hasNonWalkable {
		t.Error("palette should contain at least one non-walkable terrain")
	}
}

func TestNewEditorMapState(t *testing.T) {
	tests := []struct {
		name   string
		mName  string
		width  int
		height int
	}{
		{name: "small map", mName: "Test", width: 10, height: 8},
		{name: "large map", mName: "Big", width: 128, height: 128},
		{name: "single tile", mName: "Tiny", width: 1, height: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewEditorMapState(tt.mName, tt.width, tt.height)
			if m.Name != tt.mName {
				t.Errorf("Name = %q, want %q", m.Name, tt.mName)
			}
			if m.Width != tt.width {
				t.Errorf("Width = %d, want %d", m.Width, tt.width)
			}
			if m.Height != tt.height {
				t.Errorf("Height = %d, want %d", m.Height, tt.height)
			}
			if len(m.Tiles) != tt.height {
				t.Errorf("Tiles rows = %d, want %d", len(m.Tiles), tt.height)
			}
			for y, row := range m.Tiles {
				if len(row) != tt.width {
					t.Errorf("Tiles[%d] cols = %d, want %d", y, len(row), tt.width)
				}
			}
			// Check default tile is walkable and transparent
			tile := m.GetTile(0, 0)
			if tile == nil {
				t.Fatal("GetTile(0,0) returned nil")
			}
			if !tile.Walkable || !tile.Transparent {
				t.Error("default tile should be walkable and transparent")
			}
		})
	}
}

func TestEditorMapState_GetTile(t *testing.T) {
	m := NewEditorMapState("Test", 10, 8)

	tests := []struct {
		name    string
		x, y    int
		wantNil bool
	}{
		{name: "valid origin", x: 0, y: 0, wantNil: false},
		{name: "valid middle", x: 5, y: 4, wantNil: false},
		{name: "valid max", x: 9, y: 7, wantNil: false},
		{name: "negative x", x: -1, y: 0, wantNil: true},
		{name: "negative y", x: 0, y: -1, wantNil: true},
		{name: "x out of bounds", x: 10, y: 0, wantNil: true},
		{name: "y out of bounds", x: 0, y: 8, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile := m.GetTile(tt.x, tt.y)
			if tt.wantNil && tile != nil {
				t.Error("expected nil tile")
			}
			if !tt.wantNil && tile == nil {
				t.Error("expected non-nil tile")
			}
		})
	}
}

func TestEditorMapState_SetTile(t *testing.T) {
	m := NewEditorMapState("Test", 10, 8)
	wall := TileRef{SpriteX: 3, SpriteY: 0, Walkable: false, Transparent: false}

	// Valid set
	if !m.SetTile(5, 3, wall) {
		t.Error("SetTile should succeed for valid coordinates")
	}
	tile := m.GetTile(5, 3)
	if tile.SpriteX != 3 || tile.Walkable || tile.Transparent {
		t.Error("SetTile did not apply correctly")
	}

	// Out of bounds
	if m.SetTile(-1, 0, wall) {
		t.Error("SetTile should fail for negative x")
	}
	if m.SetTile(0, -1, wall) {
		t.Error("SetTile should fail for negative y")
	}
	if m.SetTile(10, 0, wall) {
		t.Error("SetTile should fail for x >= width")
	}
	if m.SetTile(0, 8, wall) {
		t.Error("SetTile should fail for y >= height")
	}
}

func TestUndoStack(t *testing.T) {
	t.Run("basic push and undo", func(t *testing.T) {
		stack := NewUndoStack(100)
		if stack.CanUndo() {
			t.Error("empty stack should not be undoable")
		}
		if stack.CanRedo() {
			t.Error("empty stack should not be redoable")
		}

		entry := UndoEntry{
			X:       5,
			Y:       3,
			OldTile: TileRef{SpriteX: 0, Walkable: true, Transparent: true},
			NewTile: TileRef{SpriteX: 3, Walkable: false, Transparent: false},
		}
		stack.Push(entry)

		if !stack.CanUndo() {
			t.Error("should be undoable after push")
		}
		if stack.Len() != 1 {
			t.Errorf("Len() = %d, want 1", stack.Len())
		}

		undone := stack.Undo()
		if undone == nil {
			t.Fatal("Undo returned nil")
		}
		if undone.X != 5 || undone.Y != 3 {
			t.Error("Undo returned wrong entry")
		}

		if stack.CanUndo() {
			t.Error("should not be undoable after undoing all")
		}
		if !stack.CanRedo() {
			t.Error("should be redoable after undo")
		}
	})

	t.Run("redo after undo", func(t *testing.T) {
		stack := NewUndoStack(100)
		entry := UndoEntry{X: 1, Y: 2}
		stack.Push(entry)
		stack.Undo()

		redone := stack.Redo()
		if redone == nil {
			t.Fatal("Redo returned nil")
		}
		if redone.X != 1 || redone.Y != 2 {
			t.Error("Redo returned wrong entry")
		}
	})

	t.Run("push discards redo history", func(t *testing.T) {
		stack := NewUndoStack(100)
		stack.Push(UndoEntry{X: 1, Y: 1})
		stack.Push(UndoEntry{X: 2, Y: 2})
		stack.Push(UndoEntry{X: 3, Y: 3})

		// Undo two entries
		stack.Undo()
		stack.Undo()

		// Push new entry — should discard X:2,Y:2 and X:3,Y:3
		stack.Push(UndoEntry{X: 4, Y: 4})

		if stack.CanRedo() {
			t.Error("redo should not be available after new push")
		}
		if stack.Len() != 2 {
			t.Errorf("Len() = %d, want 2", stack.Len())
		}
	})

	t.Run("max size enforcement", func(t *testing.T) {
		stack := NewUndoStack(3)
		stack.Push(UndoEntry{X: 1})
		stack.Push(UndoEntry{X: 2})
		stack.Push(UndoEntry{X: 3})
		stack.Push(UndoEntry{X: 4}) // should evict X:1

		if stack.Len() != 3 {
			t.Errorf("Len() = %d, want 3", stack.Len())
		}

		// Undo all — should get 4, 3, 2
		e1 := stack.Undo()
		e2 := stack.Undo()
		e3 := stack.Undo()

		if e1.X != 4 || e2.X != 3 || e3.X != 2 {
			t.Errorf("undo order wrong: got %d, %d, %d", e1.X, e2.X, e3.X)
		}

		if stack.CanUndo() {
			t.Error("should not be undoable, oldest entry was evicted")
		}
	})

	t.Run("undo on empty returns nil", func(t *testing.T) {
		stack := NewUndoStack(10)
		if stack.Undo() != nil {
			t.Error("Undo on empty stack should return nil")
		}
	})

	t.Run("redo on empty returns nil", func(t *testing.T) {
		stack := NewUndoStack(10)
		if stack.Redo() != nil {
			t.Error("Redo on empty stack should return nil")
		}
	})
}
