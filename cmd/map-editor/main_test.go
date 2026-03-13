package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"goldbox-rpg/pkg/cliutil"
	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEmptyMap(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "small map", width: 10, height: 10},
		{name: "large map", width: 50, height: 30},
		{name: "narrow map", width: 5, height: 20},
		{name: "wide map", width: 30, height: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createEmptyMap(tt.width, tt.height)

			assert.Equal(t, tt.width, m.Width)
			assert.Equal(t, tt.height, m.Height)
			assert.Len(t, m.Tiles, tt.height)

			// Verify all tiles are floor (walkable)
			for y := 0; y < tt.height; y++ {
				assert.Len(t, m.Tiles[y], tt.width)
				for x := 0; x < tt.width; x++ {
					assert.True(t, m.Tiles[y][x].Walkable, "tile at (%d,%d) should be walkable", x, y)
				}
			}
		})
	}
}

func TestCreateFromTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{name: "dungeon template", template: "dungeon", wantErr: false},
		{name: "outdoor template", template: "outdoor", wantErr: false},
		{name: "cave template", template: "cave", wantErr: false},
		{name: "town template", template: "town", wantErr: false},
		{name: "invalid template", template: "invalid", wantErr: true},
		{name: "case insensitive", template: "DUNGEON", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := createFromTemplate(tt.template, 20, 15)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, m)
			assert.Equal(t, 20, m.Width)
			assert.Equal(t, 15, m.Height)
		})
	}
}

func TestApplyDungeonTemplate(t *testing.T) {
	m := createEmptyMap(20, 15)
	applyDungeonTemplate(m)

	// Verify some walls exist (edges should be walls)
	assert.False(t, m.Tiles[0][0].Walkable, "corner should be wall")

	// Verify there's walkable space in the center
	centerY, centerX := m.Height/2, m.Width/2
	assert.True(t, m.Tiles[centerY][centerX].Walkable, "center should be walkable")
}

func TestApplyOutdoorTemplate(t *testing.T) {
	m := createEmptyMap(20, 15)
	applyOutdoorTemplate(m)

	// Verify middle row is a path (walkable)
	centerY := m.Height / 2
	for x := 0; x < m.Width; x++ {
		assert.True(t, m.Tiles[centerY][x].Walkable, "path at (%d,%d) should be walkable", x, centerY)
	}
}

func TestApplyCaveTemplate(t *testing.T) {
	m := createEmptyMap(20, 15)
	applyCaveTemplate(m)

	// Verify some walkable space exists
	hasWalkable := false
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.Tiles[y][x].Walkable {
				hasWalkable = true
				break
			}
		}
	}
	assert.True(t, hasWalkable, "cave should have walkable areas")
}

func TestApplyTownTemplate(t *testing.T) {
	m := createEmptyMap(20, 15)
	applyTownTemplate(m)

	// Verify roads exist (crossroads at center)
	centerY, centerX := m.Height/2, m.Width/2
	assert.True(t, m.Tiles[centerY][centerX].Walkable, "crossroads should be walkable")
}

func TestSetTile(t *testing.T) {
	tests := []struct {
		name    string
		coords  string
		char    string
		wantErr bool
	}{
		{name: "valid floor", coords: "5,5", char: ".", wantErr: false},
		{name: "valid wall", coords: "0,0", char: "#", wantErr: false},
		{name: "valid water", coords: "7,3", char: "~", wantErr: false},
		{name: "invalid coords format", coords: "5", char: ".", wantErr: true},
		{name: "invalid y coord", coords: "abc,5", char: ".", wantErr: true},
		{name: "invalid x coord", coords: "5,abc", char: ".", wantErr: true},
		{name: "out of bounds y", coords: "100,5", char: ".", wantErr: true},
		{name: "out of bounds x", coords: "5,100", char: ".", wantErr: true},
		{name: "negative y", coords: "-1,5", char: ".", wantErr: true},
		{name: "unknown tile char", coords: "5,5", char: "Z", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createEmptyMap(20, 15)
			err := setTile(m, tt.coords, tt.char)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFillMap(t *testing.T) {
	m := createEmptyMap(10, 10)
	fillMap(m, '#')

	// Verify all tiles are walls
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			assert.False(t, m.Tiles[y][x].Walkable, "tile at (%d,%d) should be wall", x, y)
		}
	}
}

func TestDrawRect(t *testing.T) {
	tests := []struct {
		name    string
		params  string
		filled  bool
		wantErr bool
	}{
		{name: "outline rectangle", params: "2,2,8,8,#", filled: false, wantErr: false},
		{name: "filled rectangle", params: "2,2,8,8,#", filled: true, wantErr: false},
		{name: "invalid params count", params: "2,2,8,#", filled: false, wantErr: true},
		{name: "unknown char", params: "2,2,8,8,Z", filled: false, wantErr: true},
		{name: "negative coords clamped", params: "-1,-1,5,5,#", filled: false, wantErr: false},
		{name: "out of bounds clamped", params: "0,0,100,100,#", filled: false, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createEmptyMap(20, 15)
			err := drawRect(m, tt.params, tt.filled)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTileToChar(t *testing.T) {
	tests := []struct {
		name string
		tile game.MapTile
		want rune
	}{
		{name: "floor", tile: charToTile['.'], want: '.'},
		{name: "wall", tile: charToTile['#'], want: '#'},
		{name: "water", tile: charToTile['~'], want: '~'},
		{name: "door", tile: charToTile['+'], want: '+'},
		{name: "unknown tile", tile: game.MapTile{SpriteX: 99, SpriteY: 99}, want: '?'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tileToChar(tt.tile)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadMap(t *testing.T) {
	// Create a temp map file
	tmpDir := t.TempDir()
	mapFile := filepath.Join(tmpDir, "test_map.json")

	testMap := &game.GameMap{
		Width:  10,
		Height: 10,
		Tiles:  createEmptyMap(10, 10).Tiles,
	}
	data, err := json.Marshal(testMap)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(mapFile, data, 0o644))

	// Test loading
	loaded, err := loadMap(mapFile)
	require.NoError(t, err)
	assert.Equal(t, 10, loaded.Width)
	assert.Equal(t, 10, loaded.Height)

	// Test loading non-existent file
	_, err = loadMap(filepath.Join(tmpDir, "nonexistent.json"))
	assert.Error(t, err)

	// Test loading invalid JSON
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	require.NoError(t, os.WriteFile(invalidFile, []byte("not json"), 0o644))
	_, err = loadMap(invalidFile)
	assert.Error(t, err)
}

func TestOutputMap(t *testing.T) {
	m := createEmptyMap(10, 10)

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output_map.json")

	err := outputMap(m, outputFile)
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	// Verify valid JSON
	var loaded game.GameMap
	err = json.Unmarshal(content, &loaded)
	require.NoError(t, err)
	assert.Equal(t, 10, loaded.Width)
}

func TestCharToTileMapping(t *testing.T) {
	// Verify all documented tile characters are mapped
	expectedChars := []rune{'.', '#', '~', '+', '>', 'T', ',', ':', 'o', '*', '$', '@'}

	for _, char := range expectedChars {
		_, ok := charToTile[char]
		assert.True(t, ok, "character %c should be mapped", char)
	}
}

func TestTileCharMapping(t *testing.T) {
	// Verify tileChar map has expected entries
	expectedTypes := []string{"floor", "wall", "water", "door", "stairs", "tree", "grass", "sand", "rock", "lava", "chest", "entrance"}

	for _, tileType := range expectedTypes {
		_, ok := tileChar[tileType]
		assert.True(t, ok, "tile type %s should be mapped", tileType)
	}
}

func TestPreviewServer(t *testing.T) {
	t.Run("create and broadcast", func(t *testing.T) {
		// Create a preview server on a random port
		ps := cliutil.NewPreviewServer(0, previewHTML, ".")
		assert.NotNil(t, ps)

		// Test broadcasting to empty client list (should not panic)
		m := createEmptyMap(5, 5)
		data, _ := json.Marshal(m)
		ps.Broadcast(data)
	})
}

func TestExecuteCommandReturnsModified(t *testing.T) {
	m := createEmptyMap(10, 10)

	tests := []struct {
		name         string
		cmd          string
		parts        []string
		wantModified bool
	}{
		{name: "save", cmd: "save", parts: []string{"save"}, wantModified: false},
		{name: "quit", cmd: "quit", parts: []string{"quit"}, wantModified: false},
		{name: "show", cmd: "show", parts: []string{"show"}, wantModified: false},
		{name: "help", cmd: "help", parts: []string{"help"}, wantModified: false},
		{name: "fill with char", cmd: "fill", parts: []string{"fill", "#"}, wantModified: true},
		{name: "fill without char", cmd: "fill", parts: []string{"fill"}, wantModified: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, modified := executeCommand(m, tt.cmd, tt.parts)
			assert.Equal(t, tt.wantModified, modified)
		})
	}
}

func TestPreviewHTMLEmbedded(t *testing.T) {
	// Verify the preview.html file is properly embedded
	content, err := previewHTML.ReadFile("preview.html")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Map Editor - Live Preview")
	assert.Contains(t, string(content), "WebSocket")
}
