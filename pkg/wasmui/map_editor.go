//go:build js && wasm

// Package wasmui provides map editor save/load functionality for WASM builds.
// This file contains browser-specific file I/O operations for the map editor.
package wasmui

import (
	"encoding/json"
	"syscall/js"
)

// SaveMapToDownload triggers a browser download of the current map as JSON.
func (g *EditorGame) SaveMapToDownload() error {
	g.mu.RLock()
	mapState := g.mapState
	mapName := g.mapName
	g.mu.RUnlock()

	if mapState == nil {
		return ErrNoMapLoaded
	}

	data, err := json.MarshalIndent(mapState, "", "  ")
	if err != nil {
		return err
	}

	triggerBrowserDownload(mapName+".json", data)

	g.mu.Lock()
	g.dirty = false
	g.mu.Unlock()

	g.setStatus("Map saved: " + mapName + ".json")
	return nil
}

// triggerBrowserDownload initiates a file download in the browser.
func triggerBrowserDownload(filename string, data []byte) {
	doc := js.Global().Get("document")
	blob := js.Global().Get("Blob").New(
		[]interface{}{string(data)},
		map[string]interface{}{"type": "application/json"},
	)

	url := js.Global().Get("URL").Call("createObjectURL", blob)

	a := doc.Call("createElement", "a")
	a.Set("href", url)
	a.Set("download", filename)
	a.Call("click")

	js.Global().Get("URL").Call("revokeObjectURL", url)
}

// LoadMapFromFile opens a file picker and loads the selected map JSON.
func (g *EditorGame) LoadMapFromFile() {
	doc := js.Global().Get("document")
	input := doc.Call("createElement", "input")
	input.Set("type", "file")
	input.Set("accept", ".json")

	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := input.Get("files")
		if files.Length() == 0 {
			return nil
		}

		file := files.Index(0)
		reader := js.Global().Get("FileReader").New()

		onload := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			result := reader.Get("result").String()
			g.loadMapFromJSON([]byte(result), file.Get("name").String())
			return nil
		})
		reader.Set("onload", onload)
		reader.Call("readAsText", file)

		return nil
	})

	input.Set("onchange", callback)
	input.Call("click")
}

// loadMapFromJSON parses JSON data and loads it as the current map.
func (g *EditorGame) loadMapFromJSON(data []byte, filename string) {
	var mapState EditorMapState
	if err := json.Unmarshal(data, &mapState); err != nil {
		g.setStatus("Error loading map: " + err.Error())
		return
	}

	g.mu.Lock()
	g.mapState = &mapState
	g.mapName = mapState.Name
	if g.mapName == "" {
		g.mapName = filename
	}
	g.dirty = false
	g.undoStack = NewUndoStack(1000)
	g.cameraX = 0
	g.cameraY = 0
	g.mu.Unlock()

	g.setStatus("Loaded: " + g.mapName)
}

// NewMap creates a new empty map with the specified dimensions.
func (g *EditorGame) NewMap(name string, width, height int) {
	g.mu.Lock()
	g.mapState = NewEditorMapState(name, width, height)
	g.mapName = name
	g.dirty = false
	g.undoStack = NewUndoStack(1000)
	g.cameraX = 0
	g.cameraY = 0
	g.mu.Unlock()

	g.setStatus("New map: " + name)
}

// ExportToGameMap converts the editor map to a game-compatible format JSON string.
func (g *EditorGame) ExportToGameMap() ([]byte, error) {
	g.mu.RLock()
	mapState := g.mapState
	g.mu.RUnlock()

	if mapState == nil {
		return nil, ErrNoMapLoaded
	}

	// Convert to the game's MapTile format
	gameMap := struct {
		Width  int `json:"width"`
		Height int `json:"height"`
		Tiles  [][]struct {
			SpriteX     int  `json:"spriteX"`
			SpriteY     int  `json:"spriteY"`
			Walkable    bool `json:"walkable"`
			Transparent bool `json:"transparent"`
		} `json:"tiles"`
	}{
		Width:  mapState.Width,
		Height: mapState.Height,
	}

	gameMap.Tiles = make([][]struct {
		SpriteX     int  `json:"spriteX"`
		SpriteY     int  `json:"spriteY"`
		Walkable    bool `json:"walkable"`
		Transparent bool `json:"transparent"`
	}, mapState.Height)

	for y := 0; y < mapState.Height; y++ {
		gameMap.Tiles[y] = make([]struct {
			SpriteX     int  `json:"spriteX"`
			SpriteY     int  `json:"spriteY"`
			Walkable    bool `json:"walkable"`
			Transparent bool `json:"transparent"`
		}, mapState.Width)
		for x := 0; x < mapState.Width; x++ {
			tile := mapState.Tiles[y][x]
			gameMap.Tiles[y][x] = struct {
				SpriteX     int  `json:"spriteX"`
				SpriteY     int  `json:"spriteY"`
				Walkable    bool `json:"walkable"`
				Transparent bool `json:"transparent"`
			}{
				SpriteX:     tile.SpriteX,
				SpriteY:     tile.SpriteY,
				Walkable:    tile.Walkable,
				Transparent: tile.Transparent,
			}
		}
	}

	return json.MarshalIndent(gameMap, "", "  ")
}
