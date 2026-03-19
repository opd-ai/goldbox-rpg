//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// updateExploration handles input for the exploration screen (§3.5).
func (g *Game) updateExploration() {
	// Check input cooldown
	if time.Since(g.lastInputTime) < g.inputCooldown {
		return
	}

	// Handle encounter overlay if visible (takes priority)
	if g.updateEncounterOverlay() {
		g.lastInputTime = time.Now()
		return
	}

	// Handle overlay toggle keys
	if g.handleExplorationOverlayKeys() {
		return
	}

	// Handle movement keys and touch
	if g.handleExplorationMovement() {
		return
	}

	// Space → end turn (if in combat context)
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.handleEndTurn()
		g.lastInputTime = time.Now()
	}

	// Handle mouse and touch tap input for exploration
	g.handleMouseInput()
}

// handleExplorationOverlayKeys processes overlay toggle hotkeys (I, Shift+S, J, G, Esc, F1).
// Returns true if an overlay was toggled.
func (g *Game) handleExplorationOverlayKeys() bool {
	// I → Inventory
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.mu.Lock()
		g.previousMode = g.mode
		g.mode = ModeInventory
		g.mu.Unlock()
		go g.loadInventory()
		g.lastInputTime = time.Now()
		return true
	}
	// Shift+S → Spellbook
	if inpututil.IsKeyJustPressed(ebiten.KeyS) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.mu.Lock()
		g.previousMode = g.mode
		g.mode = ModeSpellcasting
		g.mu.Unlock()
		go g.loadSpells()
		g.lastInputTime = time.Now()
		return true
	}
	// J → Quest Log
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		g.mu.Lock()
		g.overlays.ShowQuestLog = true
		g.mu.Unlock()
		go g.loadQuestLog()
		g.lastInputTime = time.Now()
		return true
	}
	// G → Guild Panel
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.mu.Lock()
		g.overlays.ShowGuildPanel = true
		g.mu.Unlock()
		go g.loadGuildData()
		g.lastInputTime = time.Now()
		return true
	}
	// Escape → Settings
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.overlays.ShowSettings = true
		g.mu.Unlock()
		g.lastInputTime = time.Now()
		return true
	}
	// F1 → Adventure Select
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.mu.Lock()
		g.mode = ModeAdventureSelect
		g.mu.Unlock()
		g.adventureScreen.RefreshAdventures(g)
		g.lastInputTime = time.Now()
		return true
	}
	return false
}

// handleExplorationMovement processes 8-directional movement keys, turning, and touch swipes.
// Returns true if movement or turning was processed.
func (g *Game) handleExplorationMovement() bool {
	// Q → Turn left (counter-clockwise)
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.mu.Lock()
		g.playerFacing = (g.playerFacing + 3) % 4 // -1 mod 4 = +3
		g.mu.Unlock()
		g.addLogMessage("Turned left", MessageInfo)
		g.lastInputTime = time.Now()
		return true
	}

	// E → Turn right (clockwise)
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.mu.Lock()
		g.playerFacing = (g.playerFacing + 1) % 4
		g.mu.Unlock()
		g.addLogMessage("Turned right", MessageInfo)
		g.lastInputTime = time.Now()
		return true
	}

	// Movement is relative to facing direction
	// W/Up = forward, S/Down = backward, A/Left = strafe left, D/Right = strafe right
	directions := map[ebiten.Key]int{
		ebiten.KeyW:          0, // forward
		ebiten.KeyArrowUp:    0,
		ebiten.KeyNumpad8:    0,
		ebiten.KeyS:          2, // backward (without shift)
		ebiten.KeyArrowDown:  2,
		ebiten.KeyNumpad2:    2,
		ebiten.KeyA:          3, // strafe left
		ebiten.KeyArrowLeft:  3,
		ebiten.KeyNumpad4:    3,
		ebiten.KeyD:          1, // strafe right
		ebiten.KeyArrowRight: 1,
		ebiten.KeyNumpad6:    1,
	}

	// S without shift is backward movement
	if inpututil.IsKeyJustPressed(ebiten.KeyS) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		// Shift+S goes to settings, don't process as movement
		return false
	}

	for key, relativeDir := range directions {
		if inpututil.IsKeyJustPressed(key) {
			// Convert relative direction to absolute direction based on facing
			g.mu.RLock()
			facing := g.playerFacing
			g.mu.RUnlock()
			absDir := (facing + relativeDir) % 4
			dirNames := []string{"north", "east", "south", "west"}
			g.handleMove(dirNames[absDir])
			g.lastInputTime = time.Now()
			return true
		}
	}

	// Touch swipe for directional movement
	if swiped, dir := g.touchState.HasSwipe(); swiped {
		if d := SwipeDirection(dir); d != "" {
			g.handleMove(d)
			g.lastInputTime = time.Now()
			return true
		}
	}

	return false
}

// drawExplorationScreen renders the full exploration UI (§3.5).
func (g *Game) drawExplorationScreen(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 30, G: 30, B: 40, A: 255})

	// Draw main game viewport
	g.drawViewport(screen)

	// Draw character panel (right side)
	g.drawCharacterPanel(screen)

	// Draw combat log (bottom-left)
	g.drawCombatLog(screen)

	// Draw action panel (bottom)
	g.drawActionPanel(screen)

	// Draw encounter overlay if visible (on top of everything)
	g.drawEncounterOverlay(screen)
}

// drawViewport renders the first-person dungeon view (Gold Box style).
// Step 16: Enforces 4:3 aspect ratio with letterboxing for authentic proportions.
// Step 17: Applies movement transition effect (50ms offset or flash).
func (g *Game) drawViewport(screen *ebiten.Image) {
	availWidth := g.screenWidth - charPanelWidth
	availHeight := g.screenHeight - logPanelHeight - actionPanelHeight

	// Calculate viewport with 4:3 aspect ratio and letterboxing
	vpWidth, vpHeight, vpX, vpY := calculateAspectRatioViewport(availWidth, availHeight)

	// Step 17: Calculate movement transition offset
	transOffsetX, transOffsetY := g.calculateMoveTransitionOffset()

	// Draw letterbox bars (black) around viewport
	drawRect(screen, 0, 0, availWidth, availHeight, color.RGBA{R: 5, G: 5, B: 10, A: 255})

	// Draw viewport background with transition offset
	drawRect(screen, vpX+transOffsetX, vpY+transOffsetY, vpWidth, vpHeight, color.RGBA{R: 10, G: 10, B: 20, A: 255})

	g.mu.RLock()
	player := g.player
	facing := g.playerFacing
	g.mu.RUnlock()

	if player == nil {
		drawColoredText(screen, "Waiting for game state...", vpX+vpWidth/2-80, vpY+vpHeight/2, ColorStatLabel)
		return
	}

	// Draw first-person view with depth slices (offset by viewport position and transition)
	g.drawFirstPersonViewAt(screen, vpX+transOffsetX, vpY+transOffsetY, vpWidth, vpHeight, facing)

	// Step 17: Draw brief flash overlay during transition
	if g.isMoveTransitionActive() {
		flashAlpha := g.getMoveTransitionFlashAlpha()
		if flashAlpha > 0 {
			drawRect(screen, vpX, vpY, vpWidth, vpHeight, color.RGBA{R: 200, G: 200, B: 255, A: flashAlpha})
		}
	}

	// Draw facing direction indicator at bottom of viewport
	facingNames := []string{"North", "East", "South", "West"}
	facingText := fmt.Sprintf("Facing: %s", facingNames[facing])
	drawColoredText(screen, facingText, vpX+10, vpY+vpHeight-20, ColorGold)

	// Draw position info
	posText := fmt.Sprintf("Pos: %d, %d", player.Position.X, player.Position.Y)
	drawColoredText(screen, posText, vpX+10, vpY+vpHeight-40, ColorStatLabel)
}

// calculateMoveTransitionOffset returns the viewport offset for movement transition.
// Step 17: Creates a brief directional shift effect during movement.
func (g *Game) calculateMoveTransitionOffset() (int, int) {
	g.mu.RLock()
	start := g.moveTransitionStart
	dir := g.moveTransitionDir
	dur := g.moveTransitionDur
	g.mu.RUnlock()

	if start.IsZero() || time.Since(start) > dur {
		return 0, 0
	}

	// Calculate progress (0.0 to 1.0)
	progress := float64(time.Since(start)) / float64(dur)
	// Ease out: start with offset, return to center
	offsetAmount := int((1.0 - progress) * 8) // Max 8 pixel offset

	switch dir {
	case "north", "forward":
		return 0, offsetAmount
	case "south", "backward":
		return 0, -offsetAmount
	case "east":
		return -offsetAmount, 0
	case "west":
		return offsetAmount, 0
	case "northeast":
		return -offsetAmount / 2, offsetAmount / 2
	case "northwest":
		return offsetAmount / 2, offsetAmount / 2
	case "southeast":
		return -offsetAmount / 2, -offsetAmount / 2
	case "southwest":
		return offsetAmount / 2, -offsetAmount / 2
	}
	return 0, 0
}

// isMoveTransitionActive returns true if a movement transition is in progress.
func (g *Game) isMoveTransitionActive() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return !g.moveTransitionStart.IsZero() && time.Since(g.moveTransitionStart) < g.moveTransitionDur
}

// getMoveTransitionFlashAlpha returns the alpha value for the flash overlay.
// Returns a value that peaks early and fades quickly for a subtle flash effect.
func (g *Game) getMoveTransitionFlashAlpha() uint8 {
	g.mu.RLock()
	start := g.moveTransitionStart
	dur := g.moveTransitionDur
	g.mu.RUnlock()

	if start.IsZero() {
		return 0
	}
	elapsed := time.Since(start)
	if elapsed > dur {
		return 0
	}

	// Peak at 25% of duration, fade out
	progress := float64(elapsed) / float64(dur)
	var alpha float64
	if progress < 0.25 {
		alpha = progress * 4 * 30 // Ramp up to 30
	} else {
		alpha = (1.0 - (progress-0.25)/0.75) * 30 // Fade from 30 to 0
	}
	return uint8(alpha)
}

// calculateAspectRatioViewport calculates viewport dimensions with 4:3 aspect ratio.
// Returns (width, height, x-offset, y-offset) for centered letterboxed viewport.
func calculateAspectRatioViewport(availW, availH int) (int, int, int, int) {
	targetAspect := float64(viewportBaseW) / float64(viewportBaseH) // 4:3 = 1.333

	// Calculate scale to fit while maintaining aspect ratio
	scaleW := float64(availW) / float64(viewportBaseW)
	scaleH := float64(availH) / float64(viewportBaseH)
	scale := min(scaleW, scaleH)

	// Round to integer multiple for clean pixel scaling
	if scale >= 2 {
		scale = float64(int(scale))
	}

	vpW := int(float64(viewportBaseW) * scale)
	vpH := int(float64(viewportBaseH) * scale)

	// Ensure aspect ratio is maintained
	if float64(vpW)/float64(vpH) > targetAspect {
		vpW = int(float64(vpH) * targetAspect)
	} else {
		vpH = int(float64(vpW) / targetAspect)
	}

	// Center in available space
	vpX := (availW - vpW) / 2
	vpY := (availH - vpH) / 2

	return vpW, vpH, vpX, vpY
}

// drawFirstPersonViewAt renders the first-person view at the specified position.
// Uses real map data from getVisibleTiles RPC when available.
func (g *Game) drawFirstPersonViewAt(screen *ebiten.Image, vpX, vpY, vpWidth, vpHeight, facing int) {
	// Color scheme for walls (EGA-inspired)
	wallColorFar := ColorPanelBorder    // dim purple-blue for distant walls
	wallColorMid := ColorStatValue      // brighter for mid-distance
	wallColorNear := ColorPanelBorderHi // brightest for near walls
	doorColor := ColorGold              // gold for door frames
	floorColor := color.RGBA{R: 60, G: 55, B: 70, A: 255}
	ceilingColor := color.RGBA{R: 30, G: 28, B: 42, A: 255}
	openingColor := color.RGBA{R: 20, G: 20, B: 30, A: 255}

	// Calculate perspective parameters
	vanishX := vpX + vpWidth/2
	vanishY := vpY + vpHeight/2

	// Draw floor and ceiling base
	floorTop := vpY + vpHeight/2
	drawRect(screen, vpX, floorTop, vpWidth, vpHeight/2, floorColor)
	drawRect(screen, vpX, vpY, vpWidth, vpHeight/2, ceilingColor)

	// Get cached visible tiles
	g.mu.RLock()
	tiles := g.visibleTiles
	g.mu.RUnlock()

	// Request visible tiles refresh if needed
	g.maybeRefreshVisibleTiles()

	// Depth insets for perspective (far, mid, near)
	farInset := vpWidth / 4
	midInset := vpWidth / 6
	nearInset := vpWidth / 10

	// Depth Y ranges
	farTop := vpY + vpHeight/4
	farBottom := vpY + vpHeight*3/4
	midTop := vpY + vpHeight/6
	midBottom := vpY + vpHeight*5/6
	nearTop := vpY + vpHeight/10
	nearBottom := vpY + vpHeight*9/10

	// Helper to check if a tile at (relX, depth) is a wall
	isWall := func(relX, depth int) bool {
		for _, t := range tiles {
			if t.RelativeX == relX && t.Depth == depth {
				return t.TileType == "wall"
			}
		}
		return true // Default to wall if unknown
	}

	// Helper to check if a tile is a door
	isDoor := func(relX, depth int) (bool, bool) { // returns (isDoor, isOpen)
		for _, t := range tiles {
			if t.RelativeX == relX && t.Depth == depth {
				if t.TileType == "door_open" {
					return true, true
				}
				if t.TileType == "door_closed" {
					return true, false
				}
			}
		}
		return false, false
	}

	// Draw far depth (depth=2)
	// Far left wall
	if isWall(-1, 2) {
		drawFilledTrapezoidAt(screen, vpX, vpY, vpX+farInset, farTop, vpHeight, farBottom-farTop, wallColorFar)
	}
	// Far right wall
	if isWall(1, 2) {
		drawFilledTrapezoidAt(screen, vpX+vpWidth-farInset, farTop, vpX+vpWidth, vpY, farBottom-farTop, vpHeight, wallColorFar)
	}
	// Far center - wall, door, or opening
	if isWall(0, 2) {
		// Solid wall in the center at far distance
		drawRect(screen, vpX+farInset, farTop, vpWidth-2*farInset, farBottom-farTop, wallColorFar)
	} else if isDoor, isOpen := isDoor(0, 2); isDoor {
		// Door at far distance
		drawRect(screen, vpX+farInset, farTop, vpWidth-2*farInset, farBottom-farTop, openingColor)
		doorWidth := (vpWidth - 2*farInset) / 3
		doorX := vanishX - doorWidth/2
		doorHeight := (farBottom - farTop) * 3 / 4
		doorY := farBottom - doorHeight
		drawRectOutline(screen, doorX-2, doorY-2, doorWidth+4, doorHeight+4, doorColor)
		if !isOpen {
			drawRect(screen, doorX+2, doorY+2, doorWidth-4, doorHeight-4,
				color.RGBA{R: 80, G: 60, B: 50, A: 255}) // Closed door
		}
	} else {
		// Open passage
		drawRect(screen, vpX+farInset, farTop, vpWidth-2*farInset, farBottom-farTop, openingColor)
	}

	// Draw mid depth (depth=1)
	if isWall(-1, 1) {
		drawVerticalGradient(screen, vpX+midInset-40, midTop, 40, midBottom-midTop, wallColorMid, wallColorFar)
	}
	if isWall(1, 1) {
		drawVerticalGradient(screen, vpX+vpWidth-midInset, midTop, 40, midBottom-midTop, wallColorMid, wallColorFar)
	}
	// Check center mid for wall blocking view
	if isWall(0, 1) {
		// Wall blocking passage at mid distance
		centerW := vpWidth - 2*midInset
		drawRect(screen, vpX+midInset, midTop, centerW, midBottom-midTop, wallColorMid)
	} else if isDoor, isOpen := isDoor(0, 1); isDoor {
		// Door at mid distance
		centerW := vpWidth - 2*midInset
		doorWidth := centerW / 2
		doorX := vpX + midInset + (centerW-doorWidth)/2
		doorHeight := (midBottom - midTop) * 3 / 4
		doorY := midBottom - doorHeight
		drawRectOutline(screen, doorX-3, doorY-3, doorWidth+6, doorHeight+6, doorColor)
		if !isOpen {
			drawRect(screen, doorX, doorY, doorWidth, doorHeight,
				color.RGBA{R: 90, G: 70, B: 55, A: 255})
		}
	}

	// Draw near depth (depth=0)
	if isWall(-1, 0) {
		drawRect(screen, vpX, nearTop, nearInset, nearBottom-nearTop, wallColorNear)
	}
	if isWall(1, 0) {
		drawRect(screen, vpX+vpWidth-nearInset, nearTop, nearInset, nearBottom-nearTop, wallColorNear)
	}
	// Check for wall or door directly ahead at near distance
	if isWall(0, 0) {
		// Wall right in front
		centerW := vpWidth - 2*nearInset
		drawRect(screen, vpX+nearInset, nearTop, centerW, nearBottom-nearTop, wallColorNear)
	} else if isDoor, isOpen := isDoor(0, 0); isDoor {
		// Door right in front
		centerW := vpWidth - 2*nearInset
		doorWidth := centerW * 2 / 3
		doorX := vpX + nearInset + (centerW-doorWidth)/2
		doorHeight := (nearBottom - nearTop) * 7 / 8
		doorY := nearBottom - doorHeight
		// Bold door frame
		drawRectOutline(screen, doorX-4, doorY-4, doorWidth+8, doorHeight+8, doorColor)
		drawRectOutline(screen, doorX-2, doorY-2, doorWidth+4, doorHeight+4, doorColor)
		if !isOpen {
			drawRect(screen, doorX, doorY, doorWidth, doorHeight,
				color.RGBA{R: 100, G: 80, B: 60, A: 255})
		}
	}

	// Draw corridor lines for depth perception
	lineColor := color.RGBA{R: 80, G: 70, B: 100, A: 128}
	drawLine(screen, vpX+nearInset, vpY+vpHeight, vanishX, vanishY, lineColor)
	drawLine(screen, vpX+vpWidth-nearInset, vpY+vpHeight, vanishX, vanishY, lineColor)
	drawLine(screen, vpX+nearInset, vpY, vanishX, vanishY, lineColor)
	drawLine(screen, vpX+vpWidth-nearInset, vpY, vanishX, vanishY, lineColor)
}

// maybeRefreshVisibleTiles requests new visible tiles if position/facing changed.
func (g *Game) maybeRefreshVisibleTiles() {
	g.mu.RLock()
	player := g.player
	facing := g.playerFacing
	lastPos := g.visibleTilesPos
	lastFacing := g.visibleTilesFace
	lastTime := g.visibleTilesTime
	g.mu.RUnlock()

	if player == nil {
		return
	}

	pos := player.Position

	// Only refresh if position or facing changed, or stale (>2 seconds old)
	needsRefresh := pos.X != lastPos.X || pos.Y != lastPos.Y ||
		pos.Level != lastPos.Level || facing != lastFacing ||
		time.Since(lastTime) > 2*time.Second

	if !needsRefresh {
		return
	}

	// Update position/facing before request to prevent multiple requests
	g.mu.Lock()
	g.visibleTilesPos = pos
	g.visibleTilesFace = facing
	g.visibleTilesTime = time.Now()
	g.mu.Unlock()

	// Async request for visible tiles
	go func() {
		result, err := g.rpcClient.GetVisibleTiles()
		if err != nil {
			return // Silently fail; will retry
		}
		if result != nil && result.Success {
			g.mu.Lock()
			g.visibleTiles = result.Tiles
			g.mu.Unlock()
		}
	}()
}

// drawFilledTrapezoidAt draws a filled trapezoid at absolute positions.
func drawFilledTrapezoidAt(screen *ebiten.Image, x1, y1, x2, y2, h1, h2 int, c color.RGBA) {
	strips := 20
	for i := 0; i < strips; i++ {
		t := float32(i) / float32(strips-1)
		sx := int(float32(x1)*(1-t) + float32(x2)*t)
		sy := int(float32(y1)*(1-t) + float32(y2)*t)
		sh := int(float32(h1)*(1-t) + float32(h2)*t)
		if sh > 0 {
			drawRect(screen, sx, sy, (x2-x1)/strips+1, sh, c)
		}
	}
}

// drawFirstPersonView renders the first-person dungeon corridor view.
// Uses pre-rendered depth slices approach: far (small), mid, near (large).
func (g *Game) drawFirstPersonView(screen *ebiten.Image, vpWidth, vpHeight, facing int) {
	// Color scheme for walls (EGA-inspired)
	wallColorFar := ColorPanelBorder    // dim purple-blue for distant walls
	wallColorMid := ColorStatValue      // brighter for mid-distance
	wallColorNear := ColorPanelBorderHi // brightest for near walls
	doorColor := ColorGold              // gold for door frames
	floorColor := color.RGBA{R: 60, G: 55, B: 70, A: 255}
	ceilingColor := color.RGBA{R: 30, G: 28, B: 42, A: 255}

	// Calculate perspective parameters
	// Vanishing point at center of viewport
	vanishX := vpWidth / 2
	vanishY := vpHeight / 2

	// Draw floor (gradient from near to far)
	floorTop := vpHeight / 2
	drawRect(screen, 0, floorTop, vpWidth, vpHeight-floorTop, floorColor)

	// Draw ceiling
	drawRect(screen, 0, 0, vpWidth, floorTop, ceilingColor)

	// Draw depth slices (far to near to ensure proper layering)
	// For now, draw a simple corridor view without actual map data
	// TODO: Query server for visible walls via getVisibleWalls RPC

	// Depth level 3 (far) - smallest opening in center
	farInset := vpWidth / 4
	farTop := vpHeight / 4
	farBottom := vpHeight * 3 / 4
	// Left wall (far)
	drawFilledTrapezoid(screen, 0, 0, farInset, farTop, vpHeight, farBottom-farTop, wallColorFar)
	// Right wall (far)
	drawFilledTrapezoid(screen, vpWidth-farInset, farTop, vpWidth, 0, farBottom-farTop, vpHeight, wallColorFar)
	// Far wall background (opening)
	drawRect(screen, farInset, farTop, vpWidth-2*farInset, farBottom-farTop,
		color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Depth level 2 (mid)
	midInset := vpWidth / 6
	midTop := vpHeight / 6
	midBottom := vpHeight * 5 / 6
	// Left wall (mid)
	drawVerticalGradient(screen, midInset-40, midTop, 40, midBottom-midTop, wallColorMid, wallColorFar)
	// Right wall (mid)
	drawVerticalGradient(screen, vpWidth-midInset, midTop, 40, midBottom-midTop, wallColorMid, wallColorFar)

	// Depth level 1 (near) - edges of view
	nearInset := vpWidth / 10
	nearTop := vpHeight / 10
	nearBottom := vpHeight * 9 / 10
	// Left wall (near)
	drawRect(screen, 0, nearTop, nearInset, nearBottom-nearTop, wallColorNear)
	// Right wall (near)
	drawRect(screen, vpWidth-nearInset, nearTop, nearInset, nearBottom-nearTop, wallColorNear)

	// Draw corridor lines for depth perception
	lineColor := color.RGBA{R: 80, G: 70, B: 100, A: 128}
	// Perspective lines on floor
	drawLine(screen, nearInset, vpHeight, vanishX, vanishY, lineColor)
	drawLine(screen, vpWidth-nearInset, vpHeight, vanishX, vanishY, lineColor)
	// Perspective lines on ceiling
	drawLine(screen, nearInset, 0, vanishX, vanishY, lineColor)
	drawLine(screen, vpWidth-nearInset, 0, vanishX, vanishY, lineColor)

	// Draw a placeholder door in the far wall
	doorWidth := (vpWidth - 2*farInset) / 3
	doorX := vanishX - doorWidth/2
	doorHeight := (farBottom - farTop) * 3 / 4
	doorY := farBottom - doorHeight
	// Door frame
	drawRectOutline(screen, doorX-2, doorY-2, doorWidth+4, doorHeight+4, doorColor)
	drawRectOutline(screen, doorX, doorY, doorWidth, doorHeight, wallColorMid)
	// Door interior (darker)
	drawRect(screen, doorX+2, doorY+2, doorWidth-4, doorHeight-4,
		color.RGBA{R: 40, G: 35, B: 50, A: 255})
}

// drawVerticalGradient draws a rectangle with a vertical color gradient.
func drawVerticalGradient(screen *ebiten.Image, x, y, w, h int, topColor, bottomColor color.RGBA) {
	if h <= 0 || w <= 0 {
		return
	}
	// Simple approach: draw several horizontal strips
	strips := 8
	stripH := h / strips
	for i := 0; i < strips; i++ {
		t := float32(i) / float32(strips-1)
		stripColor := color.RGBA{
			R: uint8(float32(topColor.R)*(1-t) + float32(bottomColor.R)*t),
			G: uint8(float32(topColor.G)*(1-t) + float32(bottomColor.G)*t),
			B: uint8(float32(topColor.B)*(1-t) + float32(bottomColor.B)*t),
			A: 255,
		}
		drawRect(screen, x, y+i*stripH, w, stripH, stripColor)
	}
}

// drawFilledTrapezoid draws a filled trapezoid shape (for perspective walls).
func drawFilledTrapezoid(screen *ebiten.Image, x1, y1, x2, y2, h1, h2 int, c color.RGBA) {
	// Simple approximation: draw vertical strips
	strips := 20
	for i := 0; i < strips; i++ {
		t := float32(i) / float32(strips-1)
		sx := int(float32(x1)*(1-t) + float32(x2)*t)
		sy := int(float32(y1)*(1-t) + float32(y2)*t)
		sh := int(float32(h1)*(1-t) + float32(h2)*t)
		if sh > 0 {
			drawRect(screen, sx, sy, (x2-x1)/strips+1, sh, c)
		}
	}
}

// drawViewportFloorTiles draws floor tiles across the viewport.
func (g *Game) drawViewportFloorTiles(screen *ebiten.Image, viewportWidth, viewportHeight int) {
	floorPath := TerrainTilePath("floor_stone", "dungeon")
	floorColor := color.RGBA{R: 60, G: 55, B: 70, A: 255}

	for y := 0; y < viewportHeight; y += tileSize {
		for x := 0; x < viewportWidth; x += tileSize {
			DrawSpriteWithFallback(screen, floorPath, x, y, tileSize, tileSize, floorColor)
		}
	}
}

// getPlayerSpritePath returns the sprite path for the player character.
func (g *Game) getPlayerSpritePath(player *PlayerState) string {
	class := player.Class
	if class == "" {
		class = "fighter"
	}
	// Default to human male for now; could extend with player race/gender fields
	return CharacterPortraitPath(class, "human", "male")
}

// drawCharacterPanel renders the character information panel (§9).
func (g *Game) drawCharacterPanel(screen *ebiten.Image) {
	panelX := g.screenWidth - charPanelWidth
	panelY := 0
	panelHeight := g.screenHeight - actionPanelHeight

	// Panel background — deep dark
	drawRect(screen, panelX, panelY, charPanelWidth, panelHeight, color.RGBA{R: 30, G: 28, B: 42, A: 255})

	// Bold Gold Box-style triple border
	drawBoldPanelBorder(screen, panelX, panelY, charPanelWidth, panelHeight)

	// Title in gold
	drawColoredText(screen, "CHARACTER", panelX+60, panelY+10, ColorGold)

	g.mu.RLock()
	player := g.player
	combat := g.combat
	g.mu.RUnlock()

	if player != nil {
		g.drawPlayerStats(screen, panelX, panelY, player)
	} else {
		drawColoredText(screen, "No character", panelX+50, panelY+80, ColorStatLabel)
	}

	// Combat info if in combat
	if combat != nil && combat.InCombat {
		g.drawCombatInfo(screen, panelX, panelY+250, combat)
	}

	// Minimap (§9.2) — 100×80 px simplified overhead view
	g.drawMinimap(screen, panelX+50, panelHeight-240)

	// Quest tracker at bottom of panel (§9.1)
	g.drawQuestTracker(screen, panelX, panelHeight-120)
}

// drawPlayerStats renders player character statistics.
func (g *Game) drawPlayerStats(screen *ebiten.Image, panelX, panelY int, player *PlayerState) {
	// Portrait (64x64) at top of panel, centered
	portraitX := panelX + (charPanelWidth-64)/2
	portraitY := panelY + 30
	portraitPath := g.getPlayerSpritePath(player)
	fallbackColor := g.getClassFallbackColor(player.Class)
	DrawSpriteWithFallback(screen, portraitPath, portraitX, portraitY, 64, 64, fallbackColor)

	// Name and class below portrait
	drawColoredText(screen, player.Name, panelX+10, panelY+105, ColorPlayerName)
	drawColoredText(screen, fmt.Sprintf("Lv %d %s", player.Level, player.Class), panelX+10, panelY+120, ColorStatLabel)

	// HP bar (§9.1) - adjusted y for portrait
	g.drawHPBar(screen, panelX, panelY+60, player)

	// AP bar (§9.1)
	g.drawAPBar(screen, panelX, panelY+160, player)

	// Attributes - adjusted y
	g.drawAttributes(screen, panelX, panelY+55, player.Attributes)

	// Position (§9.5)
	drawColoredText(screen, fmt.Sprintf("Pos: (%d, %d)", player.Position.X, player.Position.Y), panelX+10, panelY+240, ColorStatLabel)

	// Active effects (§9.4)
	g.drawActiveEffects(screen, panelX, panelY+255, player.Effects)

	// Effect immunities
	g.drawImmunities(screen, panelX, panelY+290, player.Immunities)
}

// getClassFallbackColor returns the fallback color for a character class portrait.
func (g *Game) getClassFallbackColor(class string) color.RGBA {
	switch class {
	case "fighter", "Fighter":
		return color.RGBA{R: 139, G: 46, B: 46, A: 255} // Medieval red
	case "mage", "Mage":
		return color.RGBA{R: 46, G: 80, B: 144, A: 255} // Deep blue
	case "cleric", "Cleric":
		return color.RGBA{R: 220, G: 220, B: 220, A: 255} // White
	case "thief", "Thief":
		return color.RGBA{R: 90, G: 90, B: 90, A: 255} // Gray
	case "ranger", "Ranger":
		return color.RGBA{R: 46, G: 139, B: 46, A: 255} // Green
	case "paladin", "Paladin":
		return color.RGBA{R: 191, G: 165, B: 74, A: 255} // Gold
	default:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255} // Default gray
	}
}

// drawHPBar renders the HP bar with color coding.
func (g *Game) drawHPBar(screen *ebiten.Image, panelX, panelY int, player *PlayerState) {
	drawColoredText(screen, "HP:", panelX+10, panelY+80, ColorStatLabel)
	hpBarWidth := charPanelWidth - 60
	hpBarX := panelX + 35
	hpBarY := panelY + 80
	drawRect(screen, hpBarX, hpBarY, hpBarWidth, 12, color.RGBA{R: 60, G: 20, B: 20, A: 255})
	if player.MaxHP > 0 {
		hpPercent := float64(player.HP) / float64(player.MaxHP)
		filledWidth := int(float64(hpBarWidth) * hpPercent)
		hpColor := hpBarColor(hpPercent)
		drawRect(screen, hpBarX, hpBarY, filledWidth, 12, hpColor)
	}
	drawColoredText(screen, fmt.Sprintf("%d/%d", player.HP, player.MaxHP), hpBarX+hpBarWidth+5, hpBarY, ColorStatValue)
}

// drawAPBar renders the AP bar as filled/empty dots (§9.1).
func (g *Game) drawAPBar(screen *ebiten.Image, panelX, y int, player *PlayerState) {
	drawColoredText(screen, "AP:", panelX+10, y, ColorStatLabel)
	ap := player.AP
	maxAP := player.MaxAP
	if maxAP == 0 {
		maxAP = 2 // default
	}
	dotStr := ""
	for i := 0; i < maxAP; i++ {
		if i < ap {
			dotStr += "@ "
		} else {
			dotStr += "o "
		}
	}
	apColor := ColorGoldHi
	if ap == 0 {
		apColor = ColorAPDepleted
	}
	drawColoredText(screen, dotStr+fmt.Sprintf("(%d/%d)", ap, maxAP), panelX+35, y, apColor)
}

// drawActiveEffects renders active effects on the character panel (§9.4).
func (g *Game) drawActiveEffects(screen *ebiten.Image, panelX, y int, effects []EffectData) {
	if len(effects) == 0 {
		return
	}
	drawColoredText(screen, "Effects:", panelX+10, y, ColorStatLabel)
	for i, eff := range effects {
		if i >= 3 {
			break
		}
		icon := EffectIcon(eff.Type)
		// Color effects by severity: debuffs red-ish, buffs green-ish
		effColor := ColorEffectDefault
		switch eff.Type {
		case "burning", "poison", "bleeding", "damage_over_time":
			effColor = ColorEffectDebuff
		case "stun", "root", "paralysis", "slow":
			effColor = ColorEffectControl
		case "regeneration", "heal_over_time", "stat_boost", "haste":
			effColor = ColorEffectBuff
		}
		drawColoredText(screen, fmt.Sprintf("%s %dt", icon, eff.Remaining), panelX+10+i*55, y+15, effColor)
	}
}

// drawImmunities renders effect immunities on the character panel.
// Immunities are shown with a distinct color to indicate protection.
func (g *Game) drawImmunities(screen *ebiten.Image, panelX, y int, immunities []string) {
	if len(immunities) == 0 {
		return
	}
	drawColoredText(screen, "Immunities:", panelX+10, y, ColorGold)
	for i, immunity := range immunities {
		if i >= 4 {
			break
		}
		// Get the immunity icon/text
		icon := getImmunityIcon(immunity)
		// Immunities use buff green color per Gold Box UI spec
		drawColoredText(screen, icon, panelX+10+i*45, y+15, ColorEffectBuff)
	}
}

// getImmunityIcon returns a short icon/text representation for an immunity type.
func getImmunityIcon(immunity string) string {
	switch immunity {
	case "burning", "fire":
		return "🔥X"
	case "poison":
		return "☠X"
	case "stun":
		return "⚡X"
	case "root":
		return "🌿X"
	case "bleeding":
		return "💧X"
	case "paralysis":
		return "⚡X"
	case "slow":
		return "🐢X"
	case "all":
		return "★"
	default:
		return immunity[:min(4, len(immunity))]
	}
}

// drawMinimap renders a simplified 100×80 overhead map in the character panel (§9.2).
// Shows explored tiles with fog of war - unexplored areas remain black.
func (g *Game) drawMinimap(screen *ebiten.Image, x, y int) {
	const mapW, mapH = 100, 80
	const tilePixels = 2 // Each tile is 2x2 pixels on minimap

	// Background (unexplored = black)
	drawRect(screen, x, y, mapW, mapH, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	// Bold double-border for Gold Box aesthetic
	drawRectOutline(screen, x, y, mapW, mapH, ColorPanelBorderHi)
	drawRectOutline(screen, x+1, y+1, mapW-2, mapH-2, ColorPanelBorder)

	drawColoredText(screen, "MAP", x+36, y-14, ColorGold)

	g.mu.RLock()
	player := g.player
	explored := g.exploredTiles
	level := 0
	if player != nil {
		level = player.Position.Level
	}
	g.mu.RUnlock()

	if player == nil {
		return
	}

	halfW, halfH := mapW/2, mapH/2
	playerX, playerY := player.Position.X, player.Position.Y

	// Draw explored tiles relative to player position
	// Visible range: ~10 tiles in each direction
	visibleRange := 20
	floorColor := color.RGBA{R: 60, G: 60, B: 70, A: 255} // Gray for explored floors

	for key := range explored {
		var tx, ty, tl int
		if _, err := fmt.Sscanf(key, "%d,%d,%d", &tx, &ty, &tl); err != nil {
			continue
		}
		if tl != level {
			continue // Only show current level
		}

		// Calculate offset from player
		dx := tx - playerX
		dy := ty - playerY

		// Skip if outside visible range
		if dx < -visibleRange/2 || dx > visibleRange/2 || dy < -visibleRange/2 || dy > visibleRange/2 {
			continue
		}

		// Calculate minimap position (scale: mapW/visibleRange per tile)
		scaleX := float64(mapW) / float64(visibleRange)
		scaleY := float64(mapH) / float64(visibleRange)
		mapX := x + halfW + int(float64(dx)*scaleX)
		mapY := y + halfH + int(float64(dy)*scaleY)

		// Draw tile (2x2 pixels) - floor
		if mapX >= x && mapX < x+mapW-tilePixels && mapY >= y && mapY < y+mapH-tilePixels {
			drawRect(screen, mapX, mapY, tilePixels, tilePixels, floorColor)
		}
	}

	// Player position (bright green dot, 3×3 at center)
	drawRect(screen, x+halfW-1, y+halfH-1, 3, 3, color.RGBA{R: 80, G: 255, B: 80, A: 255})

	// Position label
	drawColoredText(screen, fmt.Sprintf("(%d,%d)", player.Position.X, player.Position.Y), x+5, y+mapH-14, ColorStatLabel)
}

// drawQuestTracker draws the compact quest tracker at the bottom of the character panel (§7).
func (g *Game) drawQuestTracker(screen *ebiten.Image, panelX, y int) {
	g.mu.RLock()
	ql := g.questLog
	g.mu.RUnlock()

	drawColoredText(screen, "QUESTS", panelX+70, y, ColorGold)
	if ql == nil || len(ql.ActiveQuests) == 0 {
		drawColoredText(screen, "(none)", panelX+10, y+15, ColorStatLabel)
		return
	}
	count := 0
	for _, q := range ql.ActiveQuests {
		if count >= 3 {
			break
		}
		for _, obj := range q.Objectives {
			if !obj.Completed && count < 3 {
				drawColoredText(screen, fmt.Sprintf("- %s [%d/%d]", truncateText(obj.Description, 18), obj.Progress, obj.Required), panelX+10, y+15+count*15, ColorStatLabel)
				count++
			}
		}
	}
}

// hpBarColor returns the appropriate color for the HP bar based on percent.
func hpBarColor(hpPercent float64) color.RGBA {
	if hpPercent > 0.5 {
		return color.RGBA{R: 50, G: 200, B: 50, A: 255}
	} else if hpPercent > 0.25 {
		return color.RGBA{R: 200, G: 200, B: 50, A: 255}
	}
	return color.RGBA{R: 200, G: 50, B: 50, A: 255}
}

// drawAttributes renders the character attributes section.
func (g *Game) drawAttributes(screen *ebiten.Image, panelX, panelY int, attrs PlayerAttributes) {
	drawColoredText(screen, fmt.Sprintf("STR:%d", attrs.Strength), panelX+10, panelY+120, ColorStatLabel)
	drawColoredText(screen, fmt.Sprintf("DEX:%d", attrs.Dexterity), panelX+100, panelY+120, ColorStatLabel)
	drawColoredText(screen, fmt.Sprintf("CON:%d", attrs.Constitution), panelX+10, panelY+135, ColorStatLabel)
	drawColoredText(screen, fmt.Sprintf("INT:%d", attrs.Intelligence), panelX+100, panelY+135, ColorStatLabel)
	drawColoredText(screen, fmt.Sprintf("WIS:%d", attrs.Wisdom), panelX+10, panelY+150, ColorStatLabel)
	drawColoredText(screen, fmt.Sprintf("CHA:%d", attrs.Charisma), panelX+100, panelY+150, ColorStatLabel)
}

// drawCombatInfo renders combat status and initiative order.
func (g *Game) drawCombatInfo(screen *ebiten.Image, panelX, combatY int, combat *CombatState) {
	drawColoredText(screen, "COMBAT", panelX+70, combatY, ColorGold)
	drawColoredText(screen, fmt.Sprintf("Round: %d", combat.Round), panelX+10, combatY+20, ColorStatValue)
	if combat.CurrentTurn != "" {
		drawColoredText(screen, fmt.Sprintf("Turn: %s", combat.CurrentTurn), panelX+10, combatY+35, ColorStatLabel)
	}

	drawColoredText(screen, "Initiative:", panelX+10, combatY+55, ColorStatLabel)
	for i, entry := range combat.Initiative {
		if i >= 5 {
			break
		}
		marker := "  "
		if entry.ID == combat.CurrentTurn {
			marker = "> "
		}
		tag := ""
		if entry.IsPlayer {
			tag = "[P]"
		}
		nameColor := ColorEnemyName
		if entry.IsPlayer {
			nameColor = ColorPlayerName
		}
		drawColoredText(screen,
			fmt.Sprintf("%s%d. %s%s (%d)", marker, i+1, tag, entry.Name, entry.Initiative),
			panelX+10, combatY+70+i*15, nameColor)
	}
}

// drawCombatLog renders the combat/game log panel with Gold Box-style colored text.
func (g *Game) drawCombatLog(screen *ebiten.Image) {
	logX := 0
	logY := g.screenHeight - logPanelHeight - actionPanelHeight
	logWidth := g.screenWidth - charPanelWidth

	// Panel background — deep dark
	drawRect(screen, logX, logY, logWidth, logPanelHeight, ColorPanelBG)

	// Bold Gold Box-style panel border with shadow
	drawBoldPanelBorder(screen, logX, logY, logWidth, logPanelHeight)

	// Title in gold
	drawColoredText(screen, "MESSAGE LOG", logX+10, logY+5, ColorGold)

	g.mu.RLock()
	messages := make([]LogMessage, len(g.logMessages))
	copy(messages, g.logMessages)
	g.mu.RUnlock()

	maxVisible := (logPanelHeight - 25) / 15
	startIdx := 0
	if len(messages) > maxVisible {
		startIdx = len(messages) - maxVisible
	}

	for i, msg := range messages[startIdx:] {
		y := logY + 25 + i*15
		if y > logY+logPanelHeight-5 {
			break
		}
		drawColoredText(screen, msg.Text, logX+10, y, msg.Type.Color())
	}
}

// drawActionPanel renders the action buttons panel.
func (g *Game) drawActionPanel(screen *ebiten.Image) {
	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth

	drawRect(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 25, G: 23, B: 38, A: 255})
	// Bold Gold Box-style panel border
	drawBoldPanelBorder(screen, 0, panelY, panelWidth, actionPanelHeight)

	// Direction buttons
	dirBounds := g.getDirectionButtonBounds()
	dirSymbols := map[string]string{
		"nw": "NW", "n": "N", "ne": "NE",
		"w": "W", "e": "E",
		"sw": "SW", "s": "S", "se": "SE",
	}
	for name, bounds := range dirBounds {
		btnColor := color.RGBA{R: 60, G: 60, B: 80, A: 255}
		if g.hoveredButton == "dir_"+name {
			btnColor = color.RGBA{R: 80, G: 80, B: 120, A: 255}
		}
		drawRect(screen, bounds.X, bounds.Y, bounds.W, bounds.H, btnColor)
		drawRectOutline(screen, bounds.X, bounds.Y, bounds.W, bounds.H, ColorPanelBorder)
		drawColoredText(screen, dirSymbols[name], bounds.X+4, bounds.Y+6, ColorStatValue)
	}

	// Action buttons
	actionBounds := g.getActionButtonBounds()
	actionLabels := map[string]string{
		"attack":  "Attack",
		"cast":    "Cast",
		"item":    "Item",
		"endturn": "End Turn",
	}
	for name, bounds := range actionBounds {
		btnColor := color.RGBA{R: 60, G: 60, B: 80, A: 255}
		if g.hoveredButton == "action_"+name {
			btnColor = color.RGBA{R: 80, G: 80, B: 120, A: 255}
		}
		if g.selectedAction == name {
			btnColor = color.RGBA{R: 100, G: 80, B: 60, A: 255}
		}
		drawRect(screen, bounds.X, bounds.Y, bounds.W, bounds.H, btnColor)
		drawRectOutline(screen, bounds.X, bounds.Y, bounds.W, bounds.H, ColorPanelBorder)
		drawColoredText(screen, actionLabels[name], bounds.X+5, bounds.Y+8, ColorStatValue)
	}

	// Mode buttons (I/S/J/G shortcuts)
	modeX := g.screenWidth - charPanelWidth - 140
	modeY := panelY + 60
	modeLabels := []string{"[I]", "[S]", "[J]", "[G]"}
	for i, label := range modeLabels {
		drawColoredText(screen, label, modeX+i*35, modeY, ColorGold)
	}
}

// Data loaders (called from goroutines)

func (g *Game) loadInventory() {
	g.mu.Lock()
	g.loadingInv = true
	g.mu.Unlock()

	result, err := g.rpcClient.GetEquipment()

	g.mu.Lock()
	g.loadingInv = false
	if err != nil {
		g.mu.Unlock()
		g.showError(fmt.Sprintf("Failed to load inventory: %v", err))
		return
	}
	g.inventoryItems = result.Inventory
	g.mu.Unlock()
}

func (g *Game) loadSpells() {
	g.mu.Lock()
	g.loadingSpells = true
	g.mu.Unlock()

	result, err := g.rpcClient.GetAllSpells()

	g.mu.Lock()
	g.loadingSpells = false
	if err != nil {
		g.mu.Unlock()
		g.showError(fmt.Sprintf("Failed to load spells: %v", err))
		return
	}
	g.spellList = result.Spells
	g.mu.Unlock()
}

func (g *Game) loadQuestLog() {
	g.mu.Lock()
	g.loadingQuestLog = true
	g.mu.Unlock()

	result, err := g.rpcClient.GetQuestLog()

	g.mu.Lock()
	g.loadingQuestLog = false
	if err != nil {
		g.mu.Unlock()
		g.showError(fmt.Sprintf("Failed to load quest log: %v", err))
		return
	}
	g.questLog = result
	g.mu.Unlock()
}

func (g *Game) loadGuildData() {
	guild, err := g.rpcClient.GetCharacterGuild()
	if err != nil {
		// Not in a guild is normal
		g.mu.Lock()
		g.guildData = nil
		g.mu.Unlock()
		return
	}
	g.mu.Lock()
	g.guildData = guild
	g.mu.Unlock()

	// Also load factions
	factions, err := g.rpcClient.GetFactionRelations()
	if err == nil {
		g.mu.Lock()
		g.factionRelations = factions.Relations
		g.mu.Unlock()
	}
}

// ======================
// Encounter Overlay System (Gold Box style dialogue/encounter panel)
// ======================

// drawEncounterOverlay renders the encounter/dialogue overlay if visible.
// Draws a centered panel over the viewport with title, text, optional portrait, and choices.
func (g *Game) drawEncounterOverlay(screen *ebiten.Image) {
	g.mu.RLock()
	overlay := g.encounterOverlay
	g.mu.RUnlock()

	if !overlay.Visible {
		return
	}

	dims := g.calculateOverlayDimensions(overlay)
	g.drawOverlayBackdrop(screen, dims)
	g.drawOverlayContent(screen, overlay, dims)
}

// overlayDimensions holds computed dimensions for encounter overlay rendering.
type overlayDimensions struct {
	viewportW, viewportH int
	panelX, panelY       int
	panelW, panelH       int
	contentX, contentY   int
	contentW             int
	textOffsetX          int
}

// Portrait dimensions for NPC encounters (Gold Box style).
const (
	npcPortraitWidth  = 96
	npcPortraitHeight = 128
	portraitBorderW   = 4  // border thickness for portrait frame
	portraitMargin    = 12 // spacing between portrait and text
)

// calculateOverlayDimensions computes panel and content area dimensions.
func (g *Game) calculateOverlayDimensions(overlay EncounterOverlay) overlayDimensions {
	viewportW := g.screenWidth - charPanelWidth
	viewportH := g.screenHeight - logPanelHeight - actionPanelHeight
	panelW, panelH := 400, 200
	if overlay.PortraitPath != "" {
		// Enlarge panel to accommodate portrait
		panelH = max(panelH, npcPortraitHeight+50)
		panelW = max(panelW, 450)
	}
	if len(overlay.Choices) > 0 {
		panelH += len(overlay.Choices) * 24
	}
	panelX := (viewportW - panelW) / 2
	panelY := (viewportH - panelH) / 2
	contentX, contentY := panelX+16, panelY+16
	contentW := panelW - 32
	textOffsetX := 0
	if overlay.PortraitPath != "" {
		textOffsetX = npcPortraitWidth + portraitBorderW*2 + portraitMargin
		contentW -= textOffsetX
	}
	return overlayDimensions{viewportW, viewportH, panelX, panelY, panelW, panelH, contentX, contentY, contentW, textOffsetX}
}

// drawOverlayBackdrop renders the backdrop and panel background.
func (g *Game) drawOverlayBackdrop(screen *ebiten.Image, dims overlayDimensions) {
	drawRect(screen, 0, 0, dims.viewportW, dims.viewportH, color.RGBA{R: 0, G: 0, B: 0, A: 160})
	drawRect(screen, dims.panelX, dims.panelY, dims.panelW, dims.panelH, ColorPanelBG)
	// Bold Gold Box-style panel border
	drawBoldPanelBorder(screen, dims.panelX, dims.panelY, dims.panelW, dims.panelH)
}

// drawOverlayContent renders the content area of the encounter overlay.
func (g *Game) drawOverlayContent(screen *ebiten.Image, overlay EncounterOverlay, dims overlayDimensions) {
	contentY := dims.contentY
	if overlay.PortraitPath != "" {
		g.drawNPCPortrait(screen, overlay.PortraitPath, dims.contentX, dims.contentY)
	}

	if overlay.Title != "" {
		drawColoredText(screen, overlay.Title, dims.contentX+dims.textOffsetX, contentY, ColorGold)
		contentY += 24
	}

	if overlay.Text != "" {
		g.drawWrappedText(screen, overlay.Text, dims.contentX+dims.textOffsetX, contentY, dims.contentW, ColorStatValue)
		contentY += 60
	}

	g.drawOverlayChoices(screen, overlay, dims.contentX+dims.textOffsetX, contentY)
	g.drawOverlayInstructions(screen, overlay, dims.panelX, dims.panelY+dims.panelH-24)
}

// drawNPCPortrait renders an NPC portrait with Gold Box-style decorative border.
func (g *Game) drawNPCPortrait(screen *ebiten.Image, path string, x, y int) {
	// Draw decorative border frame
	frameX, frameY := x-portraitBorderW, y-portraitBorderW
	frameW, frameH := npcPortraitWidth+portraitBorderW*2, npcPortraitHeight+portraitBorderW*2

	// Outer bright edge
	drawRectOutline(screen, frameX, frameY, frameW, frameH, ColorPanelBorderHi)
	// Middle border
	drawRectOutline(screen, frameX+1, frameY+1, frameW-2, frameH-2, ColorGold)
	// Inner border
	drawRectOutline(screen, frameX+2, frameY+2, frameW-4, frameH-4, ColorPanelBorder)
	// Inner shadow
	drawRectOutline(screen, frameX+3, frameY+3, frameW-6, frameH-6, ColorPanelShadow)

	// Draw the portrait image (use adventure sprite loader for adventure NPC portraits)
	fallbackColor := color.RGBA{R: 60, G: 50, B: 80, A: 255}
	DrawAdventureSpriteWithFallback(screen, path, x, y, npcPortraitWidth, npcPortraitHeight, fallbackColor)
}

// drawOverlayChoices renders the choice list for encounter overlays.
func (g *Game) drawOverlayChoices(screen *ebiten.Image, overlay EncounterOverlay, x, y int) {
	if len(overlay.Choices) == 0 {
		return
	}
	y += 16
	for i, choice := range overlay.Choices {
		clr, prefix := ColorStatLabel, "  "
		if i == overlay.SelectedChoice {
			clr, prefix = ColorGoldHi, "> "
		}
		drawColoredText(screen, prefix+choice, x, y+i*24, clr)
	}
}

// drawOverlayInstructions renders the instruction text at the bottom.
func (g *Game) drawOverlayInstructions(screen *ebiten.Image, overlay EncounterOverlay, panelX, y int) {
	if len(overlay.Choices) > 0 {
		drawColoredText(screen, "[↑/↓] Navigate  [Enter] Select", panelX+80, y, ColorStatLabel)
	} else {
		drawColoredText(screen, "[Enter] Continue", panelX+140, y, ColorStatLabel)
	}
}

// drawWrappedText draws text with basic word wrapping.
func (g *Game) drawWrappedText(screen *ebiten.Image, text string, x, y, maxWidth int, c color.RGBA) {
	// Approximate character width (using debug font)
	charWidth := 6
	charsPerLine := maxWidth / charWidth
	if charsPerLine < 10 {
		charsPerLine = 10
	}

	words := splitWords(text)
	line := ""
	lineY := y

	for _, word := range words {
		testLine := line
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > charsPerLine && line != "" {
			drawColoredText(screen, line, x, lineY, c)
			lineY += 16
			line = word
		} else {
			line = testLine
		}
	}
	// Draw remaining line
	if line != "" {
		drawColoredText(screen, line, x, lineY, c)
	}
}

// splitWords splits text into words, handling basic whitespace.
func splitWords(text string) []string {
	var words []string
	var current string
	for _, r := range text {
		if r == ' ' || r == '\n' || r == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

// updateEncounterOverlay handles input for the encounter overlay.
// Returns true if the overlay consumed the input.
func (g *Game) updateEncounterOverlay() bool {
	g.mu.RLock()
	visible := g.encounterOverlay.Visible
	hasChoices := g.encounterOverlay.HasChoices()
	g.mu.RUnlock()

	if !visible {
		return false
	}

	// Handle choice navigation
	if hasChoices {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			g.mu.Lock()
			g.encounterOverlay.SelectPrev()
			g.mu.Unlock()
			return true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			g.mu.Lock()
			g.encounterOverlay.SelectNext()
			g.mu.Unlock()
			return true
		}
	}

	// Enter to dismiss or select choice
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.mu.Lock()
		choice := ""
		if hasChoices {
			choice = g.encounterOverlay.GetSelectedChoice()
		}
		g.encounterOverlay.Visible = false
		g.mu.Unlock()

		// If there was a choice selected, we could send it to the server
		// For now, just log it
		if choice != "" {
			g.addLogMessage(fmt.Sprintf("Selected: %s", choice), MessageInfo)
		}
		return true
	}

	// Escape to dismiss
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.encounterOverlay.Visible = false
		g.mu.Unlock()
		return true
	}

	return true // Overlay is visible, consume all input
}

// ShowEncounter displays an encounter overlay with the given content.
func (g *Game) ShowEncounter(title, text string, choices []string, portraitPath string) {
	g.mu.Lock()
	g.encounterOverlay = EncounterOverlay{
		Visible:        true,
		Title:          title,
		Text:           text,
		PortraitPath:   portraitPath,
		Choices:        choices,
		SelectedChoice: 0,
	}
	g.mu.Unlock()
}

// DismissEncounter hides the encounter overlay.
func (g *Game) DismissEncounter() {
	g.mu.Lock()
	g.encounterOverlay.Visible = false
	g.mu.Unlock()
}
