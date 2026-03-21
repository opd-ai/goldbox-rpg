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

// handleExplorationOverlayKeys processes overlay toggle hotkeys (I, C, Shift+S, J, G, M, Esc, F1).
// C opens the spellbook (primary shortcut); Shift+S is retained for compatibility.
// Also handles number keys 1-6 for party member selection.
// Returns true if an overlay was toggled or party member was selected.
func (g *Game) handleExplorationOverlayKeys() bool {
	// Number keys 1-6 → Select party member
	if g.handlePartyMemberSelection() {
		return true
	}

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
	// Game improvement #1: C opens the spellbook (primary shortcut, matches command menu).
	// C → Spellbook (primary shortcut, matches command menu)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.mu.Lock()
		g.previousMode = g.mode
		g.mode = ModeSpellcasting
		g.mu.Unlock()
		go g.loadSpells()
		g.lastInputTime = time.Now()
		return true
	}
	// Shift+S → Spellbook (legacy shortcut for compatibility)
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
	// F → Faction Relations Panel (opens guild panel on Factions tab)
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.mu.Lock()
		g.overlays.ShowGuildPanel = true
		g.guildTab = 2 // Factions tab
		g.mu.Unlock()
		go g.loadGuildData()
		g.addLogMessage("Viewing faction relations", MessageSystem)
		g.lastInputTime = time.Now()
		return true
	}
	// M → Minimap overlay
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.mu.Lock()
		g.overlays.ShowMinimap = !g.overlays.ShowMinimap
		g.mu.Unlock()
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
	// Ctrl+L → Cycle message log filter
	if inpututil.IsKeyJustPressed(ebiten.KeyL) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		g.mu.Lock()
		g.logFilterIndex = (g.logFilterIndex + 1) % len(logFilterModes)
		filterName := logFilterModes[g.logFilterIndex].Name
		g.mu.Unlock()
		g.addLogMessage(fmt.Sprintf("Log filter: %s", filterName), MessageSystem)
		g.lastInputTime = time.Now()
		return true
	}
	return false
}

// handlePartyMemberSelection checks for number keys 1-6 to select party members.
// Returns true if a party member was selected.
func (g *Game) handlePartyMemberSelection() bool {
	keys := []ebiten.Key{
		ebiten.Key1, ebiten.Key2, ebiten.Key3,
		ebiten.Key4, ebiten.Key5, ebiten.Key6,
	}

	for i, key := range keys {
		if inpututil.IsKeyJustPressed(key) {
			g.mu.Lock()
			// Get total party size
			allMembers := g.getAllPartyMembers(g.partyMembers, g.player)
			if i < len(allMembers) {
				g.selectedPartyMember = i
				g.addLogMessageLocked(fmt.Sprintf("Selected: %s", allMembers[i].Name), MessageSystem)
			}
			g.mu.Unlock()
			g.lastInputTime = time.Now()
			return true
		}
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

	// Draw themed viewport border frame
	g.mu.RLock()
	theme := g.dungeonTheme
	g.mu.RUnlock()
	if theme == "" {
		theme = "classic"
	}
	drawViewportBorderFrame(screen, vpX, vpY, vpWidth, vpHeight, getThemePalette(theme))

	// Step 17: Draw brief flash overlay and direction indicator during transition
	if g.isMoveTransitionActive() {
		flashAlpha := g.getMoveTransitionFlashAlpha()
		if flashAlpha > 0 {
			drawRect(screen, vpX, vpY, vpWidth, vpHeight, color.RGBA{R: 200, G: 200, B: 255, A: flashAlpha})
		}
		// Draw movement direction indicator at bottom center
		g.drawMoveDirectionIndicator(screen, vpX+vpWidth/2, vpY+vpHeight-30)
	}

	// Draw compass rose in upper-right corner of viewport
	drawCompassRose(screen, vpX+vpWidth-60, vpY+10, facing)

	// Draw position info at bottom-left
	posText := fmt.Sprintf("Pos: %d, %d", player.Position.X, player.Position.Y)
	drawColoredText(screen, posText, vpX+10, vpY+vpHeight-40, ColorStatLabel)
}

// drawCompassRose draws a Gold Box-style compass indicator showing current facing.
// facing: 0=North, 1=East, 2=South, 3=West
func drawCompassRose(screen *ebiten.Image, x, y, facing int) {
	const size = 50
	centerX := x + size/2
	centerY := y + size/2

	// Draw gray circular background
	drawRect(screen, x, y, size, size, color.RGBA{R: 40, G: 40, B: 50, A: 200})
	drawRectOutline(screen, x, y, size, size, ColorPanelBorder)

	// Cardinal directions: N=0, E=1, S=2, W=3
	// Positions for letters (relative to center)
	positions := []struct {
		dx, dy int
		letter string
	}{
		{0, -18, "N"}, // North at top
		{18, 0, "E"},  // East at right
		{0, 18, "S"},  // South at bottom
		{-18, 0, "W"}, // West at left
	}

	for i, pos := range positions {
		clr := ColorStatLabel
		if i == facing {
			clr = ColorGold
		}
		// Center the letter at position
		drawColoredText(screen, pos.letter, centerX+pos.dx-4, centerY+pos.dy-6, clr)
	}

	// Draw center dot
	dotColor := ColorPanelBorderHi
	drawRect(screen, centerX-2, centerY-2, 4, 4, dotColor)

	// Draw indicator line from center toward facing direction
	lineLen := 12
	dx, dy := 0, 0
	switch facing {
	case 0:
		dy = -lineLen // North
	case 1:
		dx = lineLen // East
	case 2:
		dy = lineLen // South
	case 3:
		dx = -lineLen // West
	}
	drawLine(screen, centerX, centerY, centerX+dx, centerY+dy, ColorGold)
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

// drawMoveDirectionIndicator draws a simple arrow showing movement direction.
func (g *Game) drawMoveDirectionIndicator(screen *ebiten.Image, cx, cy int) {
	g.mu.RLock()
	dir := g.moveTransitionDir
	g.mu.RUnlock()

	// Arrow dimensions
	const arrowLen = 16
	const arrowWidth = 8

	arrowColor := color.RGBA{R: 255, G: 255, B: 100, A: 180}

	// Calculate arrow direction offsets
	var dx, dy int
	switch dir {
	case "north", "forward":
		dy = -arrowLen
	case "south", "backward":
		dy = arrowLen
	case "east":
		dx = arrowLen
	case "west":
		dx = -arrowLen
	case "northeast":
		dx, dy = arrowLen/2, -arrowLen/2
	case "northwest":
		dx, dy = -arrowLen/2, -arrowLen/2
	case "southeast":
		dx, dy = arrowLen/2, arrowLen/2
	case "southwest":
		dx, dy = -arrowLen/2, arrowLen/2
	default:
		return // No direction indicator for turn
	}

	// Draw simple line arrow from center to direction
	drawRect(screen, cx-2, cy-2, 5, 5, arrowColor)                                               // Center dot
	drawRect(screen, cx+dx/2-1, cy+dy/2-1, 3, 3, arrowColor)                                     // Mid point
	drawRect(screen, cx+dx-arrowWidth/2, cy+dy-arrowWidth/2, arrowWidth, arrowWidth, arrowColor) // Arrow head
}

// getMoveTransitionFlashAlpha returns the alpha value for the flash overlay.
// Returns a value that peaks early and fades quickly for a more noticeable flash effect.
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
		alpha = progress * 4 * 50 // Ramp up to 50 for more noticeable feedback
	} else {
		alpha = (1.0 - (progress-0.25)/0.75) * 50 // Fade from 50 to 0
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

// fpvParams holds parameters for first-person view rendering.
type fpvParams struct {
	vpX, vpY, vpWidth, vpHeight int
	vanishX, vanishY            int
	// Depth insets for perspective
	farInset, midInset, nearInset int
	// Depth Y ranges
	farTop, farBottom   int
	midTop, midBottom   int
	nearTop, nearBottom int
	// Colors
	wallColorFar, wallColorMid, wallColorNear color.RGBA
	doorColor, openingColor                   color.RGBA
	// Theme palette for atmosphere rendering
	palette fpvThemePalette
	// Tile helpers
	tiles []VisibleTile
	// Player position for deterministic decoration seeding
	posX, posY int
}

// newFPVParams creates rendering parameters for first-person view.
func newFPVParams(vpX, vpY, vpWidth, vpHeight int, tiles []VisibleTile, palette fpvThemePalette) *fpvParams {
	return &fpvParams{
		vpX: vpX, vpY: vpY, vpWidth: vpWidth, vpHeight: vpHeight,
		vanishX: vpX + vpWidth/2, vanishY: vpY + vpHeight/2,
		// Depth insets
		farInset: vpWidth / 4, midInset: vpWidth / 6, nearInset: vpWidth / 10,
		// Depth Y ranges
		farTop: vpY + vpHeight/4, farBottom: vpY + vpHeight*3/4,
		midTop: vpY + vpHeight/6, midBottom: vpY + vpHeight*5/6,
		nearTop: vpY + vpHeight/10, nearBottom: vpY + vpHeight*9/10,
		// Colors from theme palette
		wallColorFar: palette.wallColorFar, wallColorMid: palette.wallColorMid, wallColorNear: palette.wallColorNear,
		doorColor: palette.doorColor, openingColor: palette.openingColor,
		palette: palette,
		tiles:   tiles,
	}
}

// isWall checks if a tile at (relX, depth) is a wall.
func (p *fpvParams) isWall(relX, depth int) bool {
	for _, t := range p.tiles {
		if t.RelativeX == relX && t.Depth == depth {
			return t.TileType == "wall"
		}
	}
	return true // Default to wall if unknown
}

// isDoor checks if a tile is a door and whether it's open.
func (p *fpvParams) isDoor(relX, depth int) (bool, bool) {
	for _, t := range p.tiles {
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

// isStairs checks if a tile at (relX, depth) is stairs.
func (p *fpvParams) isStairs(relX, depth int) bool {
	for _, t := range p.tiles {
		if t.RelativeX == relX && t.Depth == depth {
			return t.TileType == "stairs"
		}
	}
	return false
}

// getTile returns the VisibleTile at (relX, depth), or nil if not found.
func (p *fpvParams) getTile(relX, depth int) *VisibleTile {
	for i := range p.tiles {
		if p.tiles[i].RelativeX == relX && p.tiles[i].Depth == depth {
			return &p.tiles[i]
		}
	}
	return nil
}

// drawFirstPersonViewAt renders the first-person view at the specified position.
// Uses real map data from getVisibleTiles RPC when available.
func (g *Game) drawFirstPersonViewAt(screen *ebiten.Image, vpX, vpY, vpWidth, vpHeight, facing int) {
	// Get current theme palette
	g.mu.RLock()
	theme := g.dungeonTheme
	tiles := g.visibleTiles
	player := g.player
	var posX, posY int
	if player != nil {
		posX = player.Position.X
		posY = player.Position.Y
	}
	g.mu.RUnlock()

	if theme == "" {
		theme = "classic"
	}
	palette := getThemePalette(theme)

	// Draw floor and ceiling base
	floorTop := vpY + vpHeight/2
	drawRect(screen, vpX, floorTop, vpWidth, vpHeight/2, palette.floorColor)
	drawRect(screen, vpX, vpY, vpWidth, vpHeight/2, palette.ceilingColor)

	// Request visible tiles refresh if needed
	g.maybeRefreshVisibleTiles()

	// Create rendering parameters with theme palette
	p := newFPVParams(vpX, vpY, vpWidth, vpHeight, tiles, palette)

	// Set player position for deterministic decoration seeding
	p.posX = posX
	p.posY = posY

	// Draw perspective grids (behind walls)
	g.drawFloorGrid(screen, p)
	g.drawCeilingBeams(screen, p)

	// Draw floor and ceiling decorative details
	drawFloorDetails(screen, p, p.posX, p.posY)
	drawCeilingDetails(screen, p, p.posX, p.posY)

	// Draw each depth layer (back to front)
	g.drawFarDepthLayer(screen, p)
	g.drawMidDepthLayer(screen, p)
	g.drawNearDepthLayer(screen, p)

	// Draw ambient occlusion at wall-floor/ceiling junctions
	drawAmbientOcclusion(screen, p)

	// Draw depth fog for atmospheric perspective
	drawDepthFog(screen, p)

	// Draw ceiling drips for natural/cave themes
	if palette.theme == "natural" && player != nil {
		drawCeilingDrips(screen, p, p.posX, p.posY)
	}

	// Draw corridor lines for depth perception
	g.drawCorridorLines(screen, p)

	// Draw vignette for dungeon atmosphere
	drawVignette(screen, p)
}

// drawFarDepthLayer renders the far depth layer (depth=2).
func (g *Game) drawFarDepthLayer(screen *ebiten.Image, p *fpvParams) {
	// Far left wall
	if p.isWall(-1, 2) {
		drawFilledTrapezoidAt(screen, p.vpX, p.vpY, p.vpX+p.farInset, p.farTop,
			p.vpHeight, p.farBottom-p.farTop, p.wallColorFar)
	}
	// Far right wall
	if p.isWall(1, 2) {
		drawFilledTrapezoidAt(screen, p.vpX+p.vpWidth-p.farInset, p.farTop, p.vpX+p.vpWidth, p.vpY,
			p.farBottom-p.farTop, p.vpHeight, p.wallColorFar)
	}
	// Far center - wall, door, or opening
	g.drawFarCenterTile(screen, p)
	// Draw architectural features on center tile at far depth (hints only)
	drawFeatureAtDepth(screen, p, 2)
}

// drawFarCenterTile renders the center tile at far depth.
func (g *Game) drawFarCenterTile(screen *ebiten.Image, p *fpvParams) {
	cx := p.vpX + p.farInset
	cw := p.vpWidth - 2*p.farInset
	ch := p.farBottom - p.farTop
	if p.isWall(0, 2) {
		drawRect(screen, cx, p.farTop, cw, ch, p.wallColorFar)
		drawWallStoneDetail(screen, cx, p.farTop, cw, ch, p.wallColorFar, 2)
		drawWallBaseTrim(screen, cx, p.farBottom, cw, p.wallColorFar)
		drawWallEdgeHighlightCenter(screen, cx, p.farTop, cw, ch, p.wallColorFar)
		return
	}
	if p.isStairs(0, 2) {
		drawStairsFar(screen, cx, p.farTop, cw, ch, p.wallColorFar)
		return
	}
	if isDoor, isOpen := p.isDoor(0, 2); isDoor {
		drawRect(screen, cx, p.farTop, cw, ch, p.openingColor)
		doorWidth := cw / 3
		doorX := p.vanishX - doorWidth/2
		doorHeight := ch * 3 / 4
		doorY := p.farBottom - doorHeight
		if isOpen {
			drawOpenDoorDetail(screen, doorX, doorY, doorWidth, doorHeight, p.doorColor, 2)
		} else {
			drawClosedDoorDetail(screen, doorX, doorY, doorWidth, doorHeight, p.doorColor, 2)
		}
		return
	}
	// Open passage
	drawRect(screen, cx, p.farTop, cw, ch, p.openingColor)
	drawCorridorDepthHint(screen, cx, p.farTop, cw, ch, p.openingColor)
}

// drawMidDepthLayer renders the mid depth layer (depth=1).
func (g *Game) drawMidDepthLayer(screen *ebiten.Image, p *fpvParams) {
	midH := p.midBottom - p.midTop
	// Side walls
	if p.isWall(-1, 1) {
		lx := p.vpX + p.midInset - 40
		drawVerticalGradient(screen, lx, p.midTop, 40, midH, p.wallColorMid, p.wallColorFar)
		drawWallStoneDetailSeeded(screen, lx, p.midTop, 40, midH, p.wallColorMid, 1, p.posX, p.posY+1)
		drawThemeWallOverlay(screen, lx, p.midTop, 40, midH, p, 1, p.posX, p.posY)
		drawWallBaseTrim(screen, lx, p.midBottom, 40, p.wallColorMid)
		drawWallEdgeHighlightLeft(screen, lx, p.midTop, 40, midH, p.wallColorMid)
	}
	if p.isWall(1, 1) {
		rx := p.vpX + p.vpWidth - p.midInset
		drawVerticalGradient(screen, rx, p.midTop, 40, midH, p.wallColorMid, p.wallColorFar)
		drawWallStoneDetailSeeded(screen, rx, p.midTop, 40, midH, p.wallColorMid, 1, p.posX+1, p.posY+1)
		drawThemeWallOverlay(screen, rx, p.midTop, 40, midH, p, 1, p.posX, p.posY)
		drawWallBaseTrim(screen, rx, p.midBottom, 40, p.wallColorMid)
		drawWallEdgeHighlightRight(screen, rx, p.midTop, 40, midH, p.wallColorMid)
	}
	// Center - wall or door
	g.drawMidCenterTile(screen, p)
	// Draw architectural features on center tile at mid depth
	drawFeatureAtDepth(screen, p, 1)
}

// drawMidCenterTile renders the center tile at mid depth.
func (g *Game) drawMidCenterTile(screen *ebiten.Image, p *fpvParams) {
	cx := p.vpX + p.midInset
	centerW := p.vpWidth - 2*p.midInset
	ch := p.midBottom - p.midTop
	if p.isWall(0, 1) {
		drawRect(screen, cx, p.midTop, centerW, ch, p.wallColorMid)
		drawWallStoneDetailSeeded(screen, cx, p.midTop, centerW, ch, p.wallColorMid, 1, p.posX, p.posY+3)
		drawThemeWallOverlay(screen, cx, p.midTop, centerW, ch, p, 1, p.posX, p.posY)
		drawWallBaseTrim(screen, cx, p.midBottom, centerW, p.wallColorMid)
		drawWallEdgeHighlightCenter(screen, cx, p.midTop, centerW, ch, p.wallColorMid)
		return
	}
	if p.isStairs(0, 1) {
		drawStairsMid(screen, cx, p.midTop, centerW, ch, p.wallColorMid)
		return
	}
	if isDoor, isOpen := p.isDoor(0, 1); isDoor {
		doorWidth := centerW / 2
		doorX := cx + (centerW-doorWidth)/2
		doorHeight := ch * 3 / 4
		doorY := p.midBottom - doorHeight
		if isOpen {
			drawOpenDoorDetail(screen, doorX, doorY, doorWidth, doorHeight, p.doorColor, 1)
		} else {
			drawClosedDoorDetail(screen, doorX, doorY, doorWidth, doorHeight, p.doorColor, 1)
		}
	}
}

// drawNearDepthLayer renders the near depth layer (depth=0).
func (g *Game) drawNearDepthLayer(screen *ebiten.Image, p *fpvParams) {
	nearH := p.nearBottom - p.nearTop
	// Side walls
	if p.isWall(-1, 0) {
		drawRect(screen, p.vpX, p.nearTop, p.nearInset, nearH, p.wallColorNear)
		drawWallStoneDetailSeeded(screen, p.vpX, p.nearTop, p.nearInset, nearH, p.wallColorNear, 0, p.posX, p.posY)
		drawThemeWallOverlay(screen, p.vpX, p.nearTop, p.nearInset, nearH, p, 0, p.posX, p.posY)
		drawWallBaseTrim(screen, p.vpX, p.nearBottom, p.nearInset, p.wallColorNear)
		drawWallEdgeHighlightLeft(screen, p.vpX, p.nearTop, p.nearInset, nearH, p.wallColorNear)
		// Render torch: always shown on side walls (default dungeon lighting)
		drawTorchFlicker(screen, p.vpX+p.nearInset/2, p.nearTop+nearH/2, p.palette)
		drawTorchLightCone(screen, p.vpX+p.nearInset/2, p.nearBottom, p.palette)
	}
	if p.isWall(1, 0) {
		rx := p.vpX + p.vpWidth - p.nearInset
		drawRect(screen, rx, p.nearTop, p.nearInset, nearH, p.wallColorNear)
		drawWallStoneDetailSeeded(screen, rx, p.nearTop, p.nearInset, nearH, p.wallColorNear, 0, p.posX+1, p.posY)
		drawThemeWallOverlay(screen, rx, p.nearTop, p.nearInset, nearH, p, 0, p.posX, p.posY)
		drawWallBaseTrim(screen, rx, p.nearBottom, p.nearInset, p.wallColorNear)
		drawWallEdgeHighlightRight(screen, rx, p.nearTop, p.nearInset, nearH, p.wallColorNear)
		// Render torch: always shown on side walls (default dungeon lighting)
		drawTorchFlicker(screen, rx+p.nearInset/2, p.nearTop+nearH/2, p.palette)
		drawTorchLightCone(screen, rx+p.nearInset/2, p.nearBottom, p.palette)
	}
	// Center - wall or door
	g.drawNearCenterTile(screen, p)
	// Draw architectural features on center tile at near depth
	drawFeatureAtDepth(screen, p, 0)
}

// drawNearCenterTile renders the center tile at near depth.
func (g *Game) drawNearCenterTile(screen *ebiten.Image, p *fpvParams) {
	cx := p.vpX + p.nearInset
	centerW := p.vpWidth - 2*p.nearInset
	ch := p.nearBottom - p.nearTop
	if p.isWall(0, 0) {
		drawRect(screen, cx, p.nearTop, centerW, ch, p.wallColorNear)
		drawWallStoneDetailSeeded(screen, cx, p.nearTop, centerW, ch, p.wallColorNear, 0, p.posX, p.posY+2)
		drawThemeWallOverlay(screen, cx, p.nearTop, centerW, ch, p, 0, p.posX, p.posY)
		drawWallBaseTrim(screen, cx, p.nearBottom, centerW, p.wallColorNear)
		drawWallEdgeHighlightCenter(screen, cx, p.nearTop, centerW, ch, p.wallColorNear)
		return
	}
	if p.isStairs(0, 0) {
		drawStairsNear(screen, cx, p.nearTop, centerW, ch, p.wallColorNear)
		return
	}
	if isDoor, isOpen := p.isDoor(0, 0); isDoor {
		doorWidth := centerW * 2 / 3
		doorX := cx + (centerW-doorWidth)/2
		doorHeight := ch * 7 / 8
		doorY := p.nearBottom - doorHeight
		drawDoorFrameShadow(screen, doorX, doorY, doorWidth, doorHeight)
		if isOpen {
			drawOpenDoorDetail(screen, doorX, doorY, doorWidth, doorHeight, p.doorColor, 0)
		} else {
			drawClosedDoorDetail(screen, doorX, doorY, doorWidth, doorHeight, p.doorColor, 0)
		}
		if p.palette.theme == "magical" {
			drawDoorMagicalGlow(screen, doorX, doorY, doorWidth, doorHeight, p.palette.accentColor)
		}
		// Flanking torches when tile has torch data
		ct := p.getTile(0, 0)
		if ct != nil && ct.HasTorch {
			torchY := doorY + doorHeight/3
			drawTorchFlicker(screen, doorX-8, torchY, p.palette)
			drawTorchFlicker(screen, doorX+doorWidth+8, torchY, p.palette)
		}
	}
}

// drawCorridorLines draws perspective lines for depth perception.
func (g *Game) drawCorridorLines(screen *ebiten.Image, p *fpvParams) {
	lineColor := color.RGBA{R: 80, G: 70, B: 100, A: 128}
	drawLine(screen, p.vpX+p.nearInset, p.vpY+p.vpHeight, p.vanishX, p.vanishY, lineColor)
	drawLine(screen, p.vpX+p.vpWidth-p.nearInset, p.vpY+p.vpHeight, p.vanishX, p.vanishY, lineColor)
	drawLine(screen, p.vpX+p.nearInset, p.vpY, p.vanishX, p.vanishY, lineColor)
	drawLine(screen, p.vpX+p.vpWidth-p.nearInset, p.vpY, p.vanishX, p.vanishY, lineColor)
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
			// Update dungeon theme from visible tiles response if present
			if result.Theme != "" {
				g.dungeonTheme = result.Theme
			}
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

// drawWallStoneDetail draws procedural stone/brick patterns on a rectangular wall.
// depth: 0=near (individual blocks), 1=mid (larger blocks), 2=far (hints only).
// Uses posX/posY seed from fpvParams to vary block size and mortar pattern per tile.
func drawWallStoneDetail(screen *ebiten.Image, x, y, w, h int, baseColor color.RGBA, depth int) {
	drawWallStoneDetailSeeded(screen, x, y, w, h, baseColor, depth, 0, 0)
}

// drawWallStoneDetailSeeded draws procedural stone patterns with per-tile variation.
func drawWallStoneDetailSeeded(screen *ebiten.Image, x, y, w, h int, baseColor color.RGBA, depth, posX, posY int) {
	if w <= 0 || h <= 0 {
		return
	}
	mortarColor := color.RGBA{
		R: uint8(max(0, int(baseColor.R)-30)),
		G: uint8(max(0, int(baseColor.G)-30)),
		B: uint8(max(0, int(baseColor.B)-25)),
		A: 100,
	}
	// Seed-based variation for block sizes
	seed := posX*decoSeedPrimeA + posY*decoSeedPrimeB
	switch depth {
	case 0: // Near — individual stone blocks with varied joints
		blockH := 12 + absInt(seed*decoSeedPrimeC)%5  // 12-16
		jointW := 24 + absInt(seed*decoSeedPrimeD)%10 // 24-33
		for row := 0; row*blockH < h; row++ {
			ly := y + row*blockH
			drawLine(screen, x, ly, x+w, ly, mortarColor)
			offset := 0
			if row%2 == 1 {
				offset = jointW/2 + absInt(seed+row*decoSeedPrimeA)%4
			}
			for jx := offset; jx < w; jx += jointW {
				drawLine(screen, x+jx, ly, x+jx, min(ly+blockH, y+h), mortarColor)
			}
		}
		// Occasional highlight stone (lighter block) for visual variety
		if w > 20 && h > 20 {
			hlSeed := seed * decoSeedPrimeC
			hlX := x + 4 + absInt(hlSeed)%(max(1, w-12))
			hlY := y + 4 + absInt(hlSeed*decoSeedPrimeA)%(max(1, h-12))
			hlColor := color.RGBA{R: uint8(min(255, int(baseColor.R)+10)), G: uint8(min(255, int(baseColor.G)+10)), B: uint8(min(255, int(baseColor.B)+8)), A: 40}
			drawRect(screen, hlX, hlY, min(8, w-(hlX-x)), min(6, h-(hlY-y)), hlColor)
		}
	case 1: // Mid — simplified horizontal mortar lines with slight variation
		blockH := 18 + absInt(seed)%5
		for row := 0; row*blockH < h; row++ {
			drawLine(screen, x, y+row*blockH, x+w, y+row*blockH, mortarColor)
		}
	case 2: // Far — subtle horizontal hints only
		if h > 6 {
			drawLine(screen, x, y+h/3, x+w, y+h/3, mortarColor)
			drawLine(screen, x, y+h*2/3, x+w, y+h*2/3, mortarColor)
		}
	}
}

// drawFloorGrid draws perspective-correct floor tile grid lines.
func (g *Game) drawFloorGrid(screen *ebiten.Image, p *fpvParams) {
	gridColor := color.RGBA{R: 80, G: 70, B: 100, A: 64}
	floorBottom := p.vpY + p.vpHeight
	// Horizontal lines at depth boundaries
	drawLine(screen, p.vpX, p.farBottom, p.vpX+p.vpWidth, p.farBottom, gridColor)
	drawLine(screen, p.vpX, p.midBottom, p.vpX+p.vpWidth, p.midBottom, gridColor)
	drawLine(screen, p.vpX, p.nearBottom, p.vpX+p.vpWidth, p.nearBottom, gridColor)
	// Converging vertical lines from bottom edge to vanishing point
	drawLine(screen, p.vpX+p.vpWidth/4, floorBottom, p.vanishX, p.vanishY, gridColor)
	drawLine(screen, p.vanishX, floorBottom, p.vanishX, p.vanishY, gridColor)
	drawLine(screen, p.vpX+p.vpWidth*3/4, floorBottom, p.vanishX, p.vanishY, gridColor)
}

// drawCeilingBeams draws perspective beam lines on the ceiling half.
func (g *Game) drawCeilingBeams(screen *ebiten.Image, p *fpvParams) {
	beamColor := color.RGBA{R: 50, G: 45, B: 70, A: 64}
	// Horizontal lines at depth boundaries
	drawLine(screen, p.vpX, p.farTop, p.vpX+p.vpWidth, p.farTop, beamColor)
	drawLine(screen, p.vpX, p.midTop, p.vpX+p.vpWidth, p.midTop, beamColor)
	drawLine(screen, p.vpX, p.nearTop, p.vpX+p.vpWidth, p.nearTop, beamColor)
	// Converging lines from top edge to vanishing point
	drawLine(screen, p.vpX+p.vpWidth/4, p.vpY, p.vanishX, p.vanishY, beamColor)
	drawLine(screen, p.vanishX, p.vpY, p.vanishX, p.vanishY, beamColor)
	drawLine(screen, p.vpX+p.vpWidth*3/4, p.vpY, p.vanishX, p.vanishY, beamColor)
}

// drawClosedDoorDetail draws a closed door with wood plank lines and iron banding.
// depth: 0=near (full detail), 1=mid (simplified), 2=far (outline only).
func drawClosedDoorDetail(screen *ebiten.Image, dx, dy, dw, dh int, doorColor color.RGBA, depth int) {
	// Door arch at near and mid depth
	if depth <= 1 {
		drawDoorArch(screen, dx, dy, dw, doorColor, depth)
	}
	// Door frame
	drawRectOutline(screen, dx, dy, dw, dh, doorColor)
	if depth >= 2 {
		return // Far depth: frame only
	}
	// Wood background (mid and near only)
	woodColor := color.RGBA{R: 100, G: 80, B: 60, A: 255}
	drawRect(screen, dx+1, dy+1, max(0, dw-2), max(0, dh-2), woodColor)
	// Wood plank lines — direction varies by position seed
	plankColor := color.RGBA{R: 75, G: 60, B: 45, A: 180}
	planks := 4
	if depth == 1 {
		planks = 3
	}
	// Vertical planks with slight grain variation
	for i := 1; i < planks; i++ {
		px := dx + dw*i/planks
		drawLine(screen, px, dy+2, px, dy+dh-2, plankColor)
		// Subtle grain lines within planks (near only)
		if depth == 0 && dh > 20 {
			grainColor := color.RGBA{R: 85, G: 68, B: 50, A: 100}
			gx := dx + dw*i/planks - dw/(planks*2)
			for gy := dy + 4; gy < dy+dh-4; gy += 6 + (i*3)%4 {
				drawLine(screen, gx, gy, gx+2, gy+3, grainColor)
			}
		}
	}
	// Iron banding (horizontal lines)
	bandColor := color.RGBA{R: 80, G: 80, B: 90, A: 200}
	drawLine(screen, dx+2, dy+dh/4, dx+dw-2, dy+dh/4, bandColor)
	drawLine(screen, dx+2, dy+dh*3/4, dx+dw-2, dy+dh*3/4, bandColor)
	if depth == 0 {
		// Door handle (small filled rectangle)
		drawRect(screen, dx+dw*3/4, dy+dh/2-2, 4, 4, doorColor)
		// Keyhole below handle
		drawDoorKeyhole(screen, dx, dy, dw, dh)
		// Iron rivets along bands
		drawDoorRivets(screen, dx, dy, dw, dh, bandColor)
		// Arched top approximation (angled lines)
		archColor := brightenColor(doorColor, 20)
		drawLine(screen, dx, dy, dx+dw/2, dy-dh/10, archColor)
		drawLine(screen, dx+dw/2, dy-dh/10, dx+dw, dy, archColor)
	}
}

// drawDoorArch draws a semicircular arch above a door frame using stacked horizontal lines.
func drawDoorArch(screen *ebiten.Image, dx, dy, dw int, doorColor color.RGBA, depth int) {
	archColor := brightenColor(doorColor, 15)
	archH := max(4, dw/6)
	if depth == 1 {
		archH = max(2, dw/8)
	}
	// Approximate semicircle with horizontal lines of decreasing width
	for i := 0; i < archH; i++ {
		t := float32(i) / float32(max(1, archH-1))
		halfW := int(float32(dw/2) * (1.0 - t*t))
		cx := dx + dw/2
		ay := dy - i - 1
		if halfW > 0 {
			drawLine(screen, cx-halfW, ay, cx+halfW, ay, archColor)
		}
	}
}

// drawOpenDoorDetail draws an open door with frame edges and recessed interior.
// depth: 0=near (full detail), 1=mid (simplified), 2=far (outline only).
func drawOpenDoorDetail(screen *ebiten.Image, dx, dy, dw, dh int, doorColor color.RGBA, depth int) {
	// Door arch at near and mid depth
	if depth <= 1 {
		drawDoorArch(screen, dx, dy, dw, doorColor, depth)
	}
	// Door frame
	drawRectOutline(screen, dx, dy, dw, dh, doorColor)
	if depth >= 2 {
		return // Far depth: frame only
	}
	// Dark recessed interior (mid and near only)
	drawRect(screen, dx+1, dy+1, max(0, dw-2), max(0, dh-2), color.RGBA{R: 15, G: 12, B: 20, A: 255})
	if dw > 2 && dh > 2 {
		frameColor := color.RGBA{
			R: uint8(max(0, int(doorColor.R)-20)),
			G: uint8(max(0, int(doorColor.G)-20)),
			B: uint8(max(0, int(doorColor.B)-15)),
			A: 255,
		}
		drawRectOutline(screen, dx+1, dy+1, dw-2, dh-2, frameColor)
	}
	// Door panel sliver on left side (the open door)
	if depth <= 1 {
		sliverW := max(2, dw/8)
		drawRect(screen, dx, dy, sliverW, dh, color.RGBA{R: 100, G: 80, B: 60, A: 255})
		drawLine(screen, dx+sliverW, dy, dx+sliverW, dy+dh, doorColor)
	}
}

// drawTorchSconce draws a small torch/sconce decoration on a side wall.
// cx, cy: center position for the torch bracket.
func drawTorchSconce(screen *ebiten.Image, cx, cy int) {
	bracketColor := color.RGBA{R: 130, G: 110, B: 90, A: 255}
	// Bracket
	drawRectOutline(screen, cx-3, cy, 6, 10, bracketColor)
	// Shaft
	drawRect(screen, cx-1, cy-8, 2, 10, bracketColor)
	// Flame layers (orange → yellow → bright core)
	drawRect(screen, cx-3, cy-14, 6, 7, color.RGBA{R: 200, G: 120, B: 30, A: 220})
	drawRect(screen, cx-2, cy-16, 4, 7, color.RGBA{R: 240, G: 200, B: 50, A: 200})
	drawRect(screen, cx-1, cy-17, 2, 5, color.RGBA{R: 255, G: 240, B: 150, A: 180})
}

// drawWallBaseTrim draws a thin dark stripe at the wall-floor boundary.
func drawWallBaseTrim(screen *ebiten.Image, x, bottom, w int, wallColor color.RGBA) {
	trimColor := color.RGBA{
		R: uint8(max(0, int(wallColor.R)-40)),
		G: uint8(max(0, int(wallColor.G)-40)),
		B: uint8(max(0, int(wallColor.B)-35)),
		A: 255,
	}
	drawRect(screen, x, bottom-2, w, 3, trimColor)
}

// drawFirstPersonView renders the first-person dungeon corridor view.
// Uses pre-rendered depth slices approach: far (small), mid, near (large).
// Deprecated: This function is not used - see drawFirstPersonViewAt which uses server-backed tiles.
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
	// Note: This draws a simple corridor view without actual map data.
	// For server-backed tiles, use drawFirstPersonViewAt instead.

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
// Supports multi-character party display with vertical roster.
//
// Layout uses a flowing cursorY approach so that elements are placed
// sequentially from top to bottom.  Each section checks whether it
// fits within the remaining panel height before drawing, preventing
// overlapping UI elements (e.g. combat info vs minimap vs quest tracker).
func (g *Game) drawCharacterPanel(screen *ebiten.Image) {
	panelX := g.screenWidth - charPanelWidth
	panelY := 0
	panelHeight := g.screenHeight - actionPanelHeight

	// Panel background — sprite or fallback to deep dark
	drawPanelBackground(screen, panelX, panelY, charPanelWidth, panelHeight, "character")

	// Bold Gold Box-style triple border
	drawBoldPanelBorder(screen, panelX, panelY, charPanelWidth, panelHeight)

	// Gold Box-style centered header
	drawPanelHeader(screen, panelX, panelY, charPanelWidth, "PARTY")

	g.mu.RLock()
	player := g.player
	partyMembers := g.partyMembers
	selectedIdx := g.selectedPartyMember
	combat := g.combat
	g.mu.RUnlock()

	// Approximate heights used to decide whether a section fits.
	const (
		playerStatsHeight   = 340 // single-player fallback stats (portrait + attrs + effects, conservatively estimated)
		memberDetailsHeight = 170 // portrait + attrs + effects + immunities, conservatively estimated
		combatInfoHeight    = 160 // title + round/turn + up to 5 initiative entries, conservatively estimated
		minimapTitleH       = 14  // "MAP" title drawn above the map area
		minimapMapH         = 80  // minimap content area
		minimapSpacing      = 5   // vertical spacing after minimap
		minimapHeight       = minimapTitleH + minimapMapH + minimapSpacing
		minimapXOffset      = 50 // horizontal offset within character panel
		questTrackerHeight  = 60 // title + up to 3 objective lines
	)

	// cursorY tracks the next available y position as we stack elements.
	cursorY := panelY + 25

	// Draw party roster at top of panel
	rosterHeight := g.drawPartyRoster(screen, panelX, cursorY, partyMembers, player, selectedIdx)
	cursorY += rosterHeight + 5

	// Draw selected member's full details below roster (if room)
	selectedPlayer := g.getSelectedPartyMember(partyMembers, player, selectedIdx)
	if selectedPlayer != nil {
		if cursorY+memberDetailsHeight <= panelHeight {
			g.drawSelectedMemberDetails(screen, panelX, cursorY, selectedPlayer)
			cursorY += memberDetailsHeight
		}
	} else if player != nil {
		// Fallback to single player if no party
		g.drawPlayerStats(screen, panelX, panelY, player)
		cursorY = panelY + playerStatsHeight // approximate height consumed by drawPlayerStats
	} else {
		drawColoredText(screen, "No character", panelX+50, panelY+80, ColorStatLabel)
	}

	// Combat info if in combat (only when there is enough room)
	if combat != nil && combat.InCombat {
		if cursorY+combatInfoHeight <= panelHeight {
			g.drawCombatInfo(screen, panelX, cursorY, combat)
			cursorY += combatInfoHeight
		}
	}

	// Minimap (§9.2) — 100×80 px simplified overhead view
	if cursorY+minimapHeight <= panelHeight {
		g.drawMinimap(screen, panelX+minimapXOffset, cursorY+minimapTitleH) // offset leaves room for "MAP" title drawn at y-14
		cursorY += minimapHeight
	}

	// Quest tracker at bottom of panel (§9.1)
	if cursorY+questTrackerHeight <= panelHeight {
		g.drawQuestTracker(screen, panelX, cursorY)
	}
}

// drawPartyRoster renders the vertical party member list.
// Each entry shows name, class, and HP bar. Selected member has gold border.
// Returns the total height used by the roster.
func (g *Game) drawPartyRoster(screen *ebiten.Image, panelX, panelY int, partyMembers []PlayerState, player *PlayerState, selectedIdx int) int {
	const entryHeight = 50
	const maxMembers = 6

	// Combine current player with party members for display
	allMembers := g.getAllPartyMembers(partyMembers, player)

	if len(allMembers) == 0 {
		return 0
	}

	y := panelY
	for i := 0; i < len(allMembers) && i < maxMembers; i++ {
		member := &allMembers[i]
		isSelected := i == selectedIdx

		// Entry background
		bgColor := color.RGBA{R: 35, G: 32, B: 48, A: 255}
		if isSelected {
			bgColor = color.RGBA{R: 45, G: 42, B: 60, A: 255} // Slightly brighter for selected
		}
		drawRect(screen, panelX+5, y, charPanelWidth-10, entryHeight-2, bgColor)

		// Selection indicator (gold border for selected)
		if isSelected {
			drawRectOutline(screen, panelX+5, y, charPanelWidth-10, entryHeight-2, ColorGold)
			drawRectOutline(screen, panelX+6, y+1, charPanelWidth-12, entryHeight-4, ColorGoldHi)
		} else {
			drawRectOutline(screen, panelX+5, y, charPanelWidth-10, entryHeight-2, ColorPanelBorder)
		}

		// Number key indicator (1-6)
		keyColor := ColorStatLabel
		if isSelected {
			keyColor = ColorGold
		}
		drawColoredText(screen, fmt.Sprintf("%d", i+1), panelX+10, y+4, keyColor)

		// Name with appropriate color
		nameColor := ColorPlayerName
		if member.HP <= 0 {
			nameColor = ColorEnemyName // Red for downed members
		}
		drawColoredText(screen, truncateName(member.Name, 10), panelX+25, y+4, nameColor)

		// Class abbreviation
		classAbbrev := getClassAbbreviation(member.Class)
		drawColoredText(screen, classAbbrev, panelX+charPanelWidth-45, y+4, ColorStatLabel)

		// HP bar (compact version)
		hpBarWidth := charPanelWidth - 35
		hpBarY := y + 22
		drawRect(screen, panelX+10, hpBarY, hpBarWidth, 10, color.RGBA{R: 60, G: 20, B: 20, A: 255})
		if member.MaxHP > 0 {
			hpPercent := float64(member.HP) / float64(member.MaxHP)
			filledWidth := int(float64(hpBarWidth) * hpPercent)
			hpColor := hpBarColor(hpPercent)
			drawRect(screen, panelX+10, hpBarY, filledWidth, 10, hpColor)
		}

		// HP text
		hpText := fmt.Sprintf("%d/%d", member.HP, member.MaxHP)
		drawColoredText(screen, hpText, panelX+10, hpBarY+12, ColorStatValue)

		// Status icons for effects (compact, max 3)
		if len(member.Effects) > 0 {
			g.drawCompactEffectIcons(screen, panelX+hpBarWidth-30, hpBarY+12, member.Effects)
		}

		y += entryHeight
	}

	return y - panelY
}

// getAllPartyMembers combines the current player with party members for display.
func (g *Game) getAllPartyMembers(partyMembers []PlayerState, player *PlayerState) []PlayerState {
	var allMembers []PlayerState

	// If party members exist, use them
	if len(partyMembers) > 0 {
		allMembers = append(allMembers, partyMembers...)
	}

	// If player exists and isn't already in party, add at front
	if player != nil {
		found := false
		for _, m := range allMembers {
			if m.ID == player.ID {
				found = true
				break
			}
		}
		if !found {
			// Prepend player to list
			allMembers = append([]PlayerState{*player}, allMembers...)
		}
	}

	return allMembers
}

// getSelectedPartyMember returns the currently selected party member.
func (g *Game) getSelectedPartyMember(partyMembers []PlayerState, player *PlayerState, selectedIdx int) *PlayerState {
	allMembers := g.getAllPartyMembers(partyMembers, player)
	if selectedIdx >= 0 && selectedIdx < len(allMembers) {
		return &allMembers[selectedIdx]
	}
	return nil
}

// drawSelectedMemberDetails renders the full stats of the selected party member.
func (g *Game) drawSelectedMemberDetails(screen *ebiten.Image, panelX, panelY int, member *PlayerState) {
	// Separator line
	drawRect(screen, panelX+10, panelY, charPanelWidth-20, 2, ColorPanelBorder)
	panelY += 10

	// Draw a smaller portrait for selected member
	const portraitSize = 64
	portraitX := panelX + 10
	portraitY := panelY
	portraitPath := getCharacterPortraitPath(member.Class)
	fallbackColor := getClassColor(member.Class)

	// Portrait border
	drawRectOutline(screen, portraitX-2, portraitY-2, portraitSize+4, portraitSize+4, ColorGold)
	DrawSpriteWithFallback(screen, portraitPath, portraitX, portraitY, portraitSize, portraitSize, fallbackColor)

	// Name and level to right of portrait
	drawColoredText(screen, member.Name, portraitX+portraitSize+10, panelY, ColorPlayerName)
	drawColoredText(screen, fmt.Sprintf("Lv %d %s", member.Level, member.Class), portraitX+portraitSize+10, panelY+15, ColorStatLabel)

	// Position
	drawColoredText(screen, fmt.Sprintf("Pos: (%d,%d)", member.Position.X, member.Position.Y), portraitX+portraitSize+10, panelY+30, ColorStatLabel)

	// Attributes below portrait (compact grid)
	attrY := panelY + portraitSize + 10
	attrs := []struct {
		name  string
		value int
	}{
		{"STR", member.Attributes.Strength},
		{"DEX", member.Attributes.Dexterity},
		{"CON", member.Attributes.Constitution},
		{"INT", member.Attributes.Intelligence},
		{"WIS", member.Attributes.Wisdom},
		{"CHA", member.Attributes.Charisma},
	}

	// Draw 3 columns x 2 rows
	colWidth := (charPanelWidth - 20) / 3
	for i, attr := range attrs {
		col := i % 3
		row := i / 3
		x := panelX + 10 + col*colWidth
		y := attrY + row*15
		drawColoredText(screen, fmt.Sprintf("%s:%d", attr.name, attr.value), x, y, ColorStatValue)
	}

	// Active effects (below attributes)
	effectY := attrY + 35
	g.drawActiveEffects(screen, panelX, effectY, member.Effects)

	// Immunities
	g.drawImmunities(screen, panelX, effectY+35, member.Immunities)
}

// drawCompactEffectIcons draws small effect icons for the party roster.
func (g *Game) drawCompactEffectIcons(screen *ebiten.Image, x, y int, effects []EffectData) {
	for i, eff := range effects {
		if i >= 3 {
			drawColoredText(screen, "+", x+i*12, y, ColorStatLabel)
			break
		}
		icon := EffectIcon(eff.Type)
		effColor := getEffectColor(eff.Type)
		drawColoredText(screen, icon, x+i*12, y, effColor)
	}
}

// getClassAbbreviation returns a 3-letter abbreviation for a class name.
func getClassAbbreviation(class string) string {
	switch class {
	case "fighter", "Fighter":
		return "FTR"
	case "mage", "Mage":
		return "MAG"
	case "cleric", "Cleric":
		return "CLR"
	case "thief", "Thief":
		return "THF"
	case "ranger", "Ranger":
		return "RNG"
	case "paladin", "Paladin":
		return "PAL"
	default:
		if len(class) >= 3 {
			return class[:3]
		}
		return class
	}
}

// truncateName truncates a name to fit in the UI.
func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-1] + "."
}

// getCharacterPortraitPath returns the portrait path for a character.
func getCharacterPortraitPath(class string) string {
	return CharacterPortraitPath(class, "human", "male")
}

// getClassColor returns a fallback color for a character class.
func getClassColor(class string) color.RGBA {
	switch class {
	case "fighter", "Fighter":
		return color.RGBA{R: 139, G: 46, B: 46, A: 255}
	case "mage", "Mage":
		return color.RGBA{R: 46, G: 80, B: 144, A: 255}
	case "cleric", "Cleric":
		return color.RGBA{R: 220, G: 220, B: 220, A: 255}
	case "thief", "Thief":
		return color.RGBA{R: 90, G: 90, B: 90, A: 255}
	case "ranger", "Ranger":
		return color.RGBA{R: 46, G: 139, B: 46, A: 255}
	case "paladin", "Paladin":
		return color.RGBA{R: 191, G: 165, B: 74, A: 255}
	default:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}
}

// drawPlayerStats renders player character statistics.
func (g *Game) drawPlayerStats(screen *ebiten.Image, panelX, panelY int, player *PlayerState) {
	// Portrait (96x96) at top of panel with Gold Box decorative border
	const portraitSize = 96
	portraitX := panelX + (charPanelWidth-portraitSize)/2
	portraitY := panelY + 30
	portraitPath := g.getPlayerSpritePath(player)
	fallbackColor := g.getClassFallbackColor(player.Class)

	// Draw decorative border around portrait (Gold Box style)
	borderX, borderY := portraitX-4, portraitY-4
	borderW, borderH := portraitSize+8, portraitSize+8
	drawRectOutline(screen, borderX, borderY, borderW, borderH, ColorPanelBorderHi)
	drawRectOutline(screen, borderX+1, borderY+1, borderW-2, borderH-2, ColorGold)
	drawRectOutline(screen, borderX+2, borderY+2, borderW-4, borderH-4, ColorPanelBorder)

	// Draw the portrait
	DrawSpriteWithFallback(screen, portraitPath, portraitX, portraitY, portraitSize, portraitSize, fallbackColor)

	// Name and class below portrait (adjusted for larger portrait)
	drawColoredText(screen, player.Name, panelX+10, panelY+140, ColorPlayerName)
	drawColoredText(screen, fmt.Sprintf("Lv %d %s", player.Level, player.Class), panelX+10, panelY+155, ColorStatLabel)

	// HP bar (§9.1) - adjusted y for larger portrait
	g.drawHPBar(screen, panelX, panelY+95, player)

	// AP bar (§9.1)
	g.drawAPBar(screen, panelX, panelY+195, player)

	// Attributes - adjusted y
	g.drawAttributes(screen, panelX, panelY+90, player.Attributes)

	// Position (§9.5)
	drawColoredText(screen, fmt.Sprintf("Pos: (%d, %d)", player.Position.X, player.Position.Y), panelX+10, panelY+275, ColorStatLabel)

	// Active effects (§9.4)
	g.drawActiveEffects(screen, panelX, panelY+290, player.Effects)

	// Effect immunities
	g.drawImmunities(screen, panelX, panelY+325, player.Immunities)
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

// drawAttributes renders the character attributes section with modifiers.
func (g *Game) drawAttributes(screen *ebiten.Image, panelX, panelY int, attrs PlayerAttributes) {
	attrList := []struct {
		name string
		val  int
		x    int
		y    int
	}{
		{"STR", attrs.Strength, panelX + 10, panelY + 120},
		{"DEX", attrs.Dexterity, panelX + 100, panelY + 120},
		{"CON", attrs.Constitution, panelX + 10, panelY + 135},
		{"INT", attrs.Intelligence, panelX + 100, panelY + 135},
		{"WIS", attrs.Wisdom, panelX + 10, panelY + 150},
		{"CHA", attrs.Charisma, panelX + 100, panelY + 150},
	}

	for _, attr := range attrList {
		mod := AttributeModifier(attr.val)
		modStr := fmt.Sprintf("%+d", mod)
		// Choose color based on modifier value
		modColor := ColorStatValue
		if mod > 0 {
			modColor = color.RGBA{R: 60, G: 180, B: 60, A: 255} // Green for positive
		} else if mod < 0 {
			modColor = color.RGBA{R: 180, G: 60, B: 60, A: 255} // Red for negative
		}
		// Draw attribute name and score
		drawColoredText(screen, fmt.Sprintf("%s:%d", attr.name, attr.val), attr.x, attr.y, ColorStatLabel)
		// Draw modifier in color
		drawColoredText(screen, modStr, attr.x+45, attr.y, modColor)
	}
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

	// Panel background — sprite or fallback to deep dark
	drawPanelBackground(screen, logX, logY, logWidth, logPanelHeight, "combat_log")

	// Bold Gold Box-style panel border with shadow
	drawBoldPanelBorder(screen, logX, logY, logWidth, logPanelHeight)

	// Get filter state for header and filtering
	g.mu.RLock()
	filterIdx := g.logFilterIndex
	messages := make([]LogMessage, len(g.logMessages))
	copy(messages, g.logMessages)
	scrollOffset := g.logScrollOffset
	g.mu.RUnlock()

	// Gold Box-style centered header with filter indicator
	filterMode := logFilterModes[filterIdx]
	headerText := "MESSAGE LOG"
	if filterMode.Name != "All" {
		headerText = fmt.Sprintf("MESSAGE LOG [%s]", filterMode.Name)
	}
	drawPanelHeader(screen, logX, logY, logWidth, headerText)

	// Filter messages based on current filter mode
	var filtered []LogMessage
	if len(filterMode.FilterTypes) == 0 {
		// "All" mode - show everything
		filtered = messages
	} else {
		for _, msg := range messages {
			for _, ft := range filterMode.FilterTypes {
				if msg.Type == ft {
					filtered = append(filtered, msg)
					break
				}
			}
		}
	}

	maxVisible := (logPanelHeight - 30) / 15 // Adjust for header height

	// Calculate visible range with scroll offset (based on filtered messages)
	endIdx := len(filtered) - scrollOffset
	startIdx := endIdx - maxVisible
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}

	for i, msg := range filtered[startIdx:endIdx] {
		y := logY + 25 + i*15
		if y > logY+logPanelHeight-5 {
			break
		}

		// Add round prefix for combat messages
		displayText := msg.Text
		xOffset := 0
		if msg.CombatRound > 0 {
			prefix := fmt.Sprintf("[R%d] ", msg.CombatRound)
			// Draw prefix in muted color
			prefixColor := color.RGBA{R: 120, G: 120, B: 140, A: 255}
			drawColoredText(screen, prefix, logX+10, y, prefixColor)
			xOffset = len(prefix) * 7 // Approximate character width
		}
		drawColoredText(screen, displayText, logX+10+xOffset, y, msg.Type.Color())
	}

	// Timestamp tooltip on hover
	mx, my := ebiten.CursorPosition()
	if mx >= logX && mx < logX+logWidth && my >= logY+25 && my < logY+logPanelHeight-5 {
		// Calculate which message line is hovered
		hoveredIdx := (my - logY - 25) / 15
		msgIdx := startIdx + hoveredIdx
		if msgIdx >= 0 && msgIdx < len(filtered) {
			msg := filtered[msgIdx]
			if msg.Timestamp > 0 {
				// Format timestamp as HH:MM:SS
				timeStr := time.Unix(0, msg.Timestamp).Format("15:04:05")
				// Draw tooltip background
				tooltipX := mx + 10
				tooltipY := my - 5
				tooltipW := len(timeStr)*7 + 8
				tooltipH := 16
				// Ensure tooltip doesn't go off-screen
				if tooltipX+tooltipW > logX+logWidth {
					tooltipX = logX + logWidth - tooltipW
				}
				drawRect(screen, tooltipX, tooltipY, tooltipW, tooltipH, color.RGBA{R: 40, G: 40, B: 60, A: 240})
				drawRectOutline(screen, tooltipX, tooltipY, tooltipW, tooltipH, ColorPanelBorder)
				drawColoredText(screen, timeStr, tooltipX+4, tooltipY+3, ColorStatValue)
			}
		}
	}

	// Draw scroll indicators if there's content above or below
	if scrollOffset > 0 {
		// Down arrow indicator (more recent messages below)
		drawColoredText(screen, "v v v", logX+logWidth/2-15, logY+logPanelHeight-15, ColorStatLabel)
	}
	if startIdx > 0 {
		// Up arrow indicator (older messages above)
		drawColoredText(screen, "^ ^ ^", logX+logWidth/2-15, logY+18, ColorStatLabel)
	}
}

// drawActionPanel renders the action buttons panel.
// Game improvement #1: Authentic Gold Box-style command menu with prominent keyboard shortcuts.
func (g *Game) drawActionPanel(screen *ebiten.Image) {
	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth

	// Panel background with deeper, more authentic Gold Box color
	drawRect(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 22, G: 20, B: 32, A: 255})
	// Bold Gold Box-style panel border
	drawBoldPanelBorder(screen, 0, panelY, panelWidth, actionPanelHeight)

	// Draw directional control pad on the left side (aligned with hit-test bounds)
	g.drawDirectionalControls(screen, 10, panelY+10)

	// Draw Gold Box-style command menu
	g.drawExplorationCommandMenu(screen)

	// Draw facing indicator next to direction pad
	g.mu.RLock()
	facing := g.playerFacing
	g.mu.RUnlock()
	facingLabels := []string{"N", "E", "S", "W"}
	facingText := fmt.Sprintf("Facing: %s", facingLabels[facing])
	drawColoredText(screen, facingText, 115, panelY+8, ColorGold)

	// Turn controls hint
	drawColoredText(screen, "[Q] Turn Left  [E] Turn Right", 115, panelY+22, ColorStatLabel)
}

// Data loaders (called from goroutines)

func (g *Game) loadInventory() {
	g.mu.Lock()
	g.loadingInv = true
	// Capture previous item IDs for comparison
	var previousItemIDs []string
	for _, item := range g.inventoryItems {
		previousItemIDs = append(previousItemIDs, item.ID)
	}
	g.mu.Unlock()

	result, err := g.rpcClient.GetEquipment()

	g.mu.Lock()
	g.loadingInv = false
	if err != nil {
		g.mu.Unlock()
		g.showError(fmt.Sprintf("Failed to load inventory: %v", err))
		return
	}

	// Build set of previous IDs for fast lookup
	prevIDs := make(map[string]bool)
	for _, id := range previousItemIDs {
		prevIDs[id] = true
	}

	// Log any new items found
	for _, item := range result.Inventory {
		if !prevIDs[item.ID] && len(previousItemIDs) > 0 {
			g.addLogMessageLocked(fmt.Sprintf("Found: %s", item.Name), MessageLoot)
		}
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
	// Capture previous quest state for comparison
	var previousActive, previousCompleted, previousFailed []string
	if g.questLog != nil {
		for _, q := range g.questLog.ActiveQuests {
			previousActive = append(previousActive, q.ID)
		}
		for _, q := range g.questLog.CompletedQuests {
			previousCompleted = append(previousCompleted, q.ID)
		}
		for _, q := range g.questLog.FailedQuests {
			previousFailed = append(previousFailed, q.ID)
		}
	}
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

	// Log quest state changes
	g.announceQuestChanges(result, previousActive, previousCompleted, previousFailed)
}

// announceQuestChanges logs any quest status changes.
func (g *Game) announceQuestChanges(result *QuestLogResult, prevActive, prevCompleted, prevFailed []string) {
	if result == nil {
		return
	}

	// Helper to check if ID is in slice
	contains := func(slice []string, id string) bool {
		for _, s := range slice {
			if s == id {
				return true
			}
		}
		return false
	}

	// Check for newly activated quests
	for _, q := range result.ActiveQuests {
		if !contains(prevActive, q.ID) && !contains(prevCompleted, q.ID) && !contains(prevFailed, q.ID) {
			g.addLogMessage(fmt.Sprintf("Quest Started: %s", q.Title), MessageQuest)
		}
	}

	// Check for newly completed quests
	for _, q := range result.CompletedQuests {
		if contains(prevActive, q.ID) {
			g.addLogMessage(fmt.Sprintf("*** Quest Complete: %s ***", q.Title), MessageQuest)
		}
	}

	// Check for newly failed quests
	for _, q := range result.FailedQuests {
		if contains(prevActive, q.ID) {
			g.addLogMessage(fmt.Sprintf("Quest Failed: %s", q.Title), MessageWarning)
		}
	}
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
			g.addLogMessage(fmt.Sprintf("Selected: %s", choice), MessageInteract)
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
