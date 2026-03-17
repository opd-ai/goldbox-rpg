//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	// EditorScreenWidth is the default width of the editor window.
	EditorScreenWidth = 1024
	// EditorScreenHeight is the default height of the editor window.
	EditorScreenHeight = 768
	// editorTileSize is the pixel size of tiles in the editor view.
	editorTileSize = 32
	// palettePanelWidth is the width of the palette sidebar.
	palettePanelWidth = 200
	// toolbarHeight is the height of the top toolbar.
	toolbarHeight = 40
	// statusBarHeight is the height of the bottom status bar.
	statusBarHeight = 24
	// paletteEntryHeight is the height of each palette entry.
	paletteEntryHeight = 28
)

// EditorGame implements ebiten.Game for the map editor.
type EditorGame struct {
	mu sync.RWMutex

	// Map state
	mapState *EditorMapState
	mapName  string

	// Editor state
	currentTool   EditorTool
	selectedTile  int
	palette       []TerrainEntry
	undoStack     *UndoStack
	cursorX       int
	cursorY       int
	isDragging    bool
	lastInputTime time.Time
	dirty         bool
	statusMessage string
	statusTimeout time.Time

	// Camera/viewport
	cameraX int
	cameraY int

	// Screen dimensions
	screenWidth  int
	screenHeight int
}

// NewEditorGame creates a new editor game instance.
func NewEditorGame() *EditorGame {
	palette := DefaultTerrainPalette()
	mapState := NewEditorMapState("Untitled", 20, 15)

	return &EditorGame{
		mapState:     mapState,
		mapName:      "Untitled",
		currentTool:  ToolPaint,
		selectedTile: 0,
		palette:      palette,
		undoStack:    NewUndoStack(1000),
		screenWidth:  EditorScreenWidth,
		screenHeight: EditorScreenHeight,
	}
}

// Update implements ebiten.Game.
func (g *EditorGame) Update() error {
	g.handleKeyboardInput()
	g.handleMouseInput()
	return nil
}

// handleKeyboardInput processes editor keyboard shortcuts.
func (g *EditorGame) handleKeyboardInput() {
	if g.handleFileShortcuts() {
		return
	}
	g.handleToolShortcuts()
	g.handleTerrainShortcuts()
	g.handleUndoRedo()
	g.handleCameraMovement()
}

// handleFileShortcuts processes file-related keyboard shortcuts.
// Returns true if a file operation was triggered, indicating other shortcuts should be skipped.
func (g *EditorGame) handleFileShortcuts() bool {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		_ = g.SaveMapToDownload()
		return true
	}
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyO) {
		g.LoadMapFromFile()
		return true
	}
	return false
}

// handleToolShortcuts processes tool selection keyboard shortcuts.
func (g *EditorGame) handleToolShortcuts() {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.setTool(ToolPaint)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.setTool(ToolErase)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.setTool(ToolFill)
	}
}

// handleTerrainShortcuts processes quick terrain selection keyboard shortcuts.
// G=Grass, W=Water, S=Stone, D=Dirt per docs/EDITOR_GUIDE.md §Keyboard Shortcuts.
func (g *EditorGame) handleTerrainShortcuts() {
	terrainKeys := []struct {
		key   ebiten.Key
		index int
		name  string
	}{
		{ebiten.KeyG, 0, "Grass"}, // Index 0 in DefaultTerrainPalette
		{ebiten.KeyW, 2, "Water"}, // Index 2
		{ebiten.KeyS, 1, "Stone"}, // Index 1
		{ebiten.KeyD, 6, "Dirt"},  // Index 6
	}

	for _, tk := range terrainKeys {
		if inpututil.IsKeyJustPressed(tk.key) {
			g.mu.Lock()
			if tk.index < len(g.palette) {
				g.selectedTile = tk.index
				g.mu.Unlock()
				g.setStatus(fmt.Sprintf("Selected: %s", tk.name))
			} else {
				g.mu.Unlock()
			}
			return
		}
	}
}

// handleUndoRedo processes undo/redo keyboard shortcuts.
func (g *EditorGame) handleUndoRedo() {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			g.redo()
		} else {
			g.undo()
		}
	}
	// Ctrl+Y is an alternative redo shortcut
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyY) {
		g.redo()
	}
}

// handleCameraMovement processes camera pan keyboard input.
func (g *EditorGame) handleCameraMovement() {
	scrollSpeed := 4
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.cameraX -= scrollSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.cameraX += scrollSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.cameraY -= scrollSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.cameraY += scrollSpeed
	}
}

// handleMouseInput processes mouse events in the editor.
func (g *EditorGame) handleMouseInput() {
	mx, my := ebiten.CursorPosition()

	// Update cursor position in tile coordinates
	mapAreaX := palettePanelWidth
	mapAreaY := toolbarHeight
	tileX := (mx - mapAreaX + g.cameraX) / editorTileSize
	tileY := (my - mapAreaY + g.cameraY) / editorTileSize

	g.mu.Lock()
	g.cursorX = tileX
	g.cursorY = tileY
	g.mu.Unlock()

	// Handle palette clicks
	if mx < palettePanelWidth && my >= toolbarHeight {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handlePaletteClick(mx, my)
		}
		return
	}

	// Handle toolbar clicks
	if my < toolbarHeight {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleToolbarClick(mx, my)
		}
		return
	}

	// Handle map area interactions
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if mx >= mapAreaX {
			g.applyToolAt(tileX, tileY)
		}
	}
}

// handlePaletteClick handles clicks in the terrain palette.
func (g *EditorGame) handlePaletteClick(mx, my int) {
	paletteY := my - toolbarHeight
	index := paletteY / paletteEntryHeight
	if index >= 0 && index < len(g.palette) {
		g.mu.Lock()
		g.selectedTile = index
		g.mu.Unlock()
		g.setStatus(fmt.Sprintf("Selected: %s", g.palette[index].Name))
	}
}

// handleToolbarClick handles clicks in the toolbar area.
func (g *EditorGame) handleToolbarClick(mx, _ int) {
	// Tool buttons: Paint(0-60), Erase(65-125), Fill(130-190)
	btnWidth := 60
	spacing := 5
	tools := []EditorTool{ToolPaint, ToolErase, ToolFill, ToolSelect}

	for i, tool := range tools {
		x := spacing + i*(btnWidth+spacing)
		if mx >= x && mx < x+btnWidth {
			g.setTool(tool)
			return
		}
	}
}

// applyToolAt applies the current tool at the specified tile coordinates.
func (g *EditorGame) applyToolAt(tileX, tileY int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.mapState == nil {
		return
	}

	oldTile := g.mapState.GetTile(tileX, tileY)
	if oldTile == nil {
		return
	}

	var newTile TileRef

	switch g.currentTool {
	case ToolPaint:
		if g.selectedTile >= 0 && g.selectedTile < len(g.palette) {
			entry := g.palette[g.selectedTile]
			newTile = TileRef{
				SpriteX:     entry.SpriteX,
				SpriteY:     entry.SpriteY,
				Walkable:    entry.Walkable,
				Transparent: entry.Transparent,
			}
		}
	case ToolErase:
		newTile = TileRef{
			SpriteX:     0,
			SpriteY:     0,
			Walkable:    true,
			Transparent: true,
		}
	default:
		return
	}

	// Skip if tile is already the same
	if *oldTile == newTile {
		return
	}

	oldCopy := *oldTile
	g.undoStack.Push(UndoEntry{
		X:       tileX,
		Y:       tileY,
		OldTile: oldCopy,
		NewTile: newTile,
	})

	g.mapState.SetTile(tileX, tileY, newTile)
	g.dirty = true
}

// setTool changes the currently selected tool.
func (g *EditorGame) setTool(tool EditorTool) {
	g.mu.Lock()
	g.currentTool = tool
	g.mu.Unlock()
	g.setStatus(fmt.Sprintf("Tool: %s", tool))
}

// undo reverts the last tile change.
func (g *EditorGame) undo() {
	g.mu.Lock()
	defer g.mu.Unlock()

	entry := g.undoStack.Undo()
	if entry == nil {
		return
	}
	g.mapState.SetTile(entry.X, entry.Y, entry.OldTile)
	g.dirty = true
}

// redo reapplies the last undone change.
func (g *EditorGame) redo() {
	g.mu.Lock()
	defer g.mu.Unlock()

	entry := g.undoStack.Redo()
	if entry == nil {
		return
	}
	g.mapState.SetTile(entry.X, entry.Y, entry.NewTile)
	g.dirty = true
}

// setStatus sets a temporary status message.
func (g *EditorGame) setStatus(msg string) {
	g.mu.Lock()
	g.statusMessage = msg
	g.statusTimeout = time.Now().Add(3 * time.Second)
	g.mu.Unlock()
}

// Draw implements ebiten.Game.
func (g *EditorGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 30, G: 30, B: 40, A: 255})

	g.drawToolbar(screen)
	g.drawPalette(screen)
	g.drawMapArea(screen)
	g.drawStatusBar(screen)
}

// drawToolbar draws the top toolbar with tool buttons.
func (g *EditorGame) drawToolbar(screen *ebiten.Image) {
	// Toolbar background
	drawFilledRect(screen, 0, 0, g.screenWidth, toolbarHeight,
		color.RGBA{R: 50, G: 50, B: 65, A: 255})

	tools := []struct {
		tool EditorTool
		name string
	}{
		{ToolPaint, "Paint"},
		{ToolErase, "Erase"},
		{ToolFill, "Fill"},
		{ToolSelect, "Select"},
	}

	g.mu.RLock()
	currentTool := g.currentTool
	g.mu.RUnlock()

	btnWidth := 60
	spacing := 5
	for i, t := range tools {
		x := spacing + i*(btnWidth+spacing)
		bg := color.RGBA{R: 70, G: 70, B: 85, A: 255}
		if t.tool == currentTool {
			bg = color.RGBA{R: 100, G: 130, B: 200, A: 255}
		}
		drawFilledRect(screen, x, 5, btnWidth, 30, bg)
		ebitenutil.DebugPrintAt(screen, t.name, x+8, 12)
	}

	// Map name
	g.mu.RLock()
	name := g.mapName
	dirty := g.dirty
	g.mu.RUnlock()

	label := name
	if dirty {
		label += " *"
	}
	ebitenutil.DebugPrintAt(screen, label, g.screenWidth/2-40, 12)
}

// drawPalette draws the terrain palette sidebar.
func (g *EditorGame) drawPalette(screen *ebiten.Image) {
	// Palette background
	drawFilledRect(screen, 0, toolbarHeight, palettePanelWidth,
		g.screenHeight-toolbarHeight-statusBarHeight,
		color.RGBA{R: 40, G: 40, B: 55, A: 255})

	g.mu.RLock()
	selectedTile := g.selectedTile
	g.mu.RUnlock()

	for i, entry := range g.palette {
		y := toolbarHeight + i*paletteEntryHeight
		bg := color.RGBA{R: 50, G: 50, B: 65, A: 255}
		if i == selectedTile {
			bg = color.RGBA{R: 80, G: 110, B: 180, A: 255}
		}
		drawFilledRect(screen, 2, y+2, palettePanelWidth-4, paletteEntryHeight-4, bg)

		// Draw terrain color preview
		tileColor := terrainColor(entry.SpriteX, entry.SpriteY)
		drawFilledRect(screen, 6, y+6, 16, 16, tileColor)

		// Draw terrain name and properties
		props := ""
		if !entry.Walkable {
			props += " [X]"
		}
		if !entry.Transparent {
			props += " [O]"
		}
		ebitenutil.DebugPrintAt(screen, entry.Name+props, 28, y+8)
	}
}

// drawMapArea draws the tile map editing area.
func (g *EditorGame) drawMapArea(screen *ebiten.Image) {
	g.mu.RLock()
	mapState := g.mapState
	cursorX := g.cursorX
	cursorY := g.cursorY
	g.mu.RUnlock()

	if mapState == nil {
		return
	}

	offsetX := palettePanelWidth - g.cameraX
	offsetY := toolbarHeight - g.cameraY

	// Draw tiles
	for y := 0; y < mapState.Height; y++ {
		for x := 0; x < mapState.Width; x++ {
			px := offsetX + x*editorTileSize
			py := offsetY + y*editorTileSize

			// Skip off-screen tiles
			if px+editorTileSize < palettePanelWidth || px > g.screenWidth {
				continue
			}
			if py+editorTileSize < toolbarHeight || py > g.screenHeight-statusBarHeight {
				continue
			}

			tile := &mapState.Tiles[y][x]
			tileColor := terrainColor(tile.SpriteX, tile.SpriteY)
			drawFilledRect(screen, px, py, editorTileSize-1, editorTileSize-1, tileColor)

			// Draw grid lines
			drawFilledRect(screen, px+editorTileSize-1, py, 1, editorTileSize,
				color.RGBA{R: 60, G: 60, B: 70, A: 128})
			drawFilledRect(screen, px, py+editorTileSize-1, editorTileSize, 1,
				color.RGBA{R: 60, G: 60, B: 70, A: 128})
		}
	}

	// Draw cursor highlight
	if cursorX >= 0 && cursorX < mapState.Width && cursorY >= 0 && cursorY < mapState.Height {
		cx := offsetX + cursorX*editorTileSize
		cy := offsetY + cursorY*editorTileSize
		drawFilledRect(screen, cx, cy, editorTileSize-1, editorTileSize-1,
			color.RGBA{R: 255, G: 255, B: 255, A: 60})
	}
}

// drawStatusBar draws the bottom status bar.
func (g *EditorGame) drawStatusBar(screen *ebiten.Image) {
	y := g.screenHeight - statusBarHeight
	drawFilledRect(screen, 0, y, g.screenWidth, statusBarHeight,
		color.RGBA{R: 35, G: 35, B: 48, A: 255})

	g.mu.RLock()
	cursorX := g.cursorX
	cursorY := g.cursorY
	mapState := g.mapState
	statusMsg := g.statusMessage
	statusTimeout := g.statusTimeout
	g.mu.RUnlock()

	// Cursor position
	posText := fmt.Sprintf("(%d, %d)", cursorX, cursorY)
	ebitenutil.DebugPrintAt(screen, posText, 5, y+5)

	// Map size
	if mapState != nil {
		sizeText := fmt.Sprintf("%dx%d", mapState.Width, mapState.Height)
		ebitenutil.DebugPrintAt(screen, sizeText, 100, y+5)
	}

	// Status message
	if time.Now().Before(statusTimeout) {
		ebitenutil.DebugPrintAt(screen, statusMsg, 200, y+5)
	}
}

// Layout implements ebiten.Game.
func (g *EditorGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

// terrainColor returns a representative color for a terrain sprite index.
func terrainColor(spriteX, spriteY int) color.RGBA {
	colors := [][]color.RGBA{
		// Row 0: Grass, Stone, Water, Wall, Door
		{
			{R: 80, G: 160, B: 60, A: 255},   // Grass
			{R: 140, G: 140, B: 140, A: 255}, // Stone
			{R: 60, G: 100, B: 200, A: 255},  // Water
			{R: 100, G: 80, B: 60, A: 255},   // Wall
			{R: 160, G: 120, B: 80, A: 255},  // Door
		},
		// Row 1: Sand, Dirt, Tree, Lava, Ice
		{
			{R: 220, G: 200, B: 140, A: 255}, // Sand
			{R: 140, G: 100, B: 60, A: 255},  // Dirt
			{R: 40, G: 120, B: 40, A: 255},   // Tree
			{R: 220, G: 80, B: 20, A: 255},   // Lava
			{R: 180, G: 220, B: 240, A: 255}, // Ice
		},
	}

	if spriteY >= 0 && spriteY < len(colors) {
		row := colors[spriteY]
		if spriteX >= 0 && spriteX < len(row) {
			return row[spriteX]
		}
	}

	return color.RGBA{R: 80, G: 80, B: 80, A: 255}
}

// drawFilledRect draws a filled rectangle on the screen.
func drawFilledRect(screen *ebiten.Image, x, y, w, h int, c color.RGBA) {
	if pixelImage == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w), float64(h))
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	screen.DrawImage(pixelImage, op)
}
