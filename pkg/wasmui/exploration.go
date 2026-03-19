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
func (g *Game) drawViewport(screen *ebiten.Image) {
	viewportWidth := g.screenWidth - charPanelWidth
	viewportHeight := g.screenHeight - logPanelHeight - actionPanelHeight

	// Draw viewport background (dark dungeon ceiling/sky)
	drawRect(screen, 0, 0, viewportWidth, viewportHeight, color.RGBA{R: 10, G: 10, B: 20, A: 255})

	g.mu.RLock()
	player := g.player
	facing := g.playerFacing
	g.mu.RUnlock()

	if player == nil {
		drawColoredText(screen, "Waiting for game state...", viewportWidth/2-80, viewportHeight/2, ColorStatLabel)
		return
	}

	// Draw first-person view with depth slices
	g.drawFirstPersonView(screen, viewportWidth, viewportHeight, facing)

	// Draw facing direction indicator at bottom of viewport
	facingNames := []string{"North", "East", "South", "West"}
	facingText := fmt.Sprintf("Facing: %s", facingNames[facing])
	drawColoredText(screen, facingText, 10, viewportHeight-20, ColorGold)

	// Draw position info
	posText := fmt.Sprintf("Pos: %d, %d", player.Position.X, player.Position.Y)
	drawColoredText(screen, posText, 10, viewportHeight-40, ColorStatLabel)
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

	// Double-border: outer bright, inner dim
	drawRectOutline(screen, panelX, panelY, charPanelWidth, panelHeight, ColorPanelBorderHi)
	drawRectOutline(screen, panelX+2, panelY+2, charPanelWidth-4, panelHeight-4, ColorPanelBorder)

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
	drawColoredText(screen, player.Name, panelX+10, panelY+40, ColorPlayerName)
	drawColoredText(screen, fmt.Sprintf("Lv %d %s", player.Level, player.Class), panelX+10, panelY+55, ColorStatLabel)

	// HP bar (§9.1)
	g.drawHPBar(screen, panelX, panelY, player)

	// AP bar (§9.1)
	g.drawAPBar(screen, panelX, panelY+95, player)

	// Attributes
	g.drawAttributes(screen, panelX, panelY, player.Attributes)

	// Position (§9.5)
	drawColoredText(screen, fmt.Sprintf("Pos: (%d, %d)", player.Position.X, player.Position.Y), panelX+10, panelY+185, ColorStatLabel)

	// Active effects (§9.4)
	g.drawActiveEffects(screen, panelX, panelY+200, player.Effects)
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

// drawMinimap renders a simplified 100×80 overhead map in the character panel (§9.2).
func (g *Game) drawMinimap(screen *ebiten.Image, x, y int) {
	const mapW, mapH = 100, 80

	// Background (unexplored = black)
	drawRect(screen, x, y, mapW, mapH, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	drawRectOutline(screen, x, y, mapW, mapH, ColorPanelBorder)

	drawColoredText(screen, "MAP", x+36, y-14, ColorGold)

	g.mu.RLock()
	player := g.player
	g.mu.RUnlock()

	if player == nil {
		return
	}

	halfW, halfH := mapW/2, mapH/2

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

	// Double-border: outer bright, inner dim — Gold Box panel framing
	drawRectOutline(screen, logX, logY, logWidth, logPanelHeight, ColorPanelBorder)
	drawRectOutline(screen, logX+2, logY+2, logWidth-4, logPanelHeight-4, color.RGBA{R: 50, G: 45, B: 70, A: 255})

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
	drawRectOutline(screen, 0, panelY, panelWidth, actionPanelHeight, ColorPanelBorder)

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
	result, err := g.rpcClient.GetEquipment()
	if err != nil {
		g.showError(fmt.Sprintf("Failed to load inventory: %v", err))
		return
	}
	g.mu.Lock()
	g.inventoryItems = result.Inventory
	g.mu.Unlock()
}

func (g *Game) loadSpells() {
	result, err := g.rpcClient.GetAllSpells()
	if err != nil {
		g.showError(fmt.Sprintf("Failed to load spells: %v", err))
		return
	}
	g.mu.Lock()
	g.spellList = result.Spells
	g.mu.Unlock()
}

func (g *Game) loadQuestLog() {
	result, err := g.rpcClient.GetQuestLog()
	if err != nil {
		g.showError(fmt.Sprintf("Failed to load quest log: %v", err))
		return
	}
	g.mu.Lock()
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

	// Panel dimensions (centered in viewport area)
	viewportWidth := g.screenWidth - charPanelWidth
	viewportHeight := g.screenHeight - logPanelHeight - actionPanelHeight

	panelWidth := 400
	panelHeight := 200
	if len(overlay.Choices) > 0 {
		panelHeight += len(overlay.Choices) * 24 // Extra height for choices
	}

	panelX := (viewportWidth - panelWidth) / 2
	panelY := (viewportHeight - panelHeight) / 2

	// Draw semi-transparent backdrop
	drawRect(screen, 0, 0, viewportWidth, viewportHeight, color.RGBA{R: 0, G: 0, B: 0, A: 160})

	// Draw panel background
	drawRect(screen, panelX, panelY, panelWidth, panelHeight, ColorPanelBG)

	// Draw double border (Gold Box style)
	drawRectOutline(screen, panelX, panelY, panelWidth, panelHeight, ColorPanelBorderHi)
	drawRectOutline(screen, panelX+2, panelY+2, panelWidth-4, panelHeight-4, ColorPanelBorder)

	// Content area offsets
	contentX := panelX + 16
	contentY := panelY + 16
	contentWidth := panelWidth - 32

	// Draw portrait if available
	portraitSize := 64
	textOffsetX := 0
	if overlay.PortraitPath != "" {
		DrawSpriteWithFallback(screen, overlay.PortraitPath,
			contentX, contentY, portraitSize, portraitSize,
			color.RGBA{R: 80, G: 80, B: 100, A: 255}) // Fallback gray
		textOffsetX = portraitSize + 12
		contentWidth -= textOffsetX
	}

	// Draw title in gold
	if overlay.Title != "" {
		drawColoredText(screen, overlay.Title, contentX+textOffsetX, contentY, ColorGold)
		contentY += 24
	}

	// Draw body text in white (word wrap)
	if overlay.Text != "" {
		g.drawWrappedText(screen, overlay.Text, contentX+textOffsetX, contentY, contentWidth, ColorStatValue)
		contentY += 60 // Approximate space for wrapped text
	}

	// Draw choices if present
	if len(overlay.Choices) > 0 {
		contentY += 16 // Gap before choices
		for i, choice := range overlay.Choices {
			choiceColor := ColorStatLabel
			prefix := "  "
			if i == overlay.SelectedChoice {
				choiceColor = ColorGoldHi
				prefix = "> "
			}
			drawColoredText(screen, prefix+choice, contentX+textOffsetX, contentY+i*24, choiceColor)
		}
	}

	// Draw instruction at bottom
	instructionY := panelY + panelHeight - 24
	if len(overlay.Choices) > 0 {
		drawColoredText(screen, "[↑/↓] Navigate  [Enter] Select", panelX+80, instructionY, ColorStatLabel)
	} else {
		drawColoredText(screen, "[Enter] Continue", panelX+140, instructionY, ColorStatLabel)
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
