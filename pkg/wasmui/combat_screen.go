//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// updateCombat handles input during combat mode (§5).
func (g *Game) updateCombat() {
	// Handle spell targeting mode separately
	if g.isSpellTargetMode() {
		g.handleSpellTargeting()
		return
	}

	// Escape → back to exploration (exit combat view)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.combatAction = CombatActionNone
		g.mu.Unlock()
		return
	}

	// Tab → cycle through targets
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			g.cycleTarget(-1)
		} else {
			g.cycleTarget(1)
		}
		return
	}

	// Handle combat action hotkeys
	if g.handleCombatHotkeys() {
		return
	}

	// Space / Enter → end turn per §5
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.handleEndTurn()
		return
	}

	// Handle movement in move mode
	if g.handleCombatMovement() {
		return
	}

	// Mouse and touch input
	g.handleMouseInput()

	// Handle touch taps on combat UI
	g.handleCombatTouchTap()
}

// handleCombatHotkeys processes combat action hotkeys (M, A, C, U, D, F).
// Game improvement #1: Added D (Defend) and F (Flee) hotkeys for command menu consistency.
// Returns true if a hotkey was pressed.
func (g *Game) handleCombatHotkeys() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.mu.Lock()
		g.combatAction = CombatActionMove
		g.mu.Unlock()
		g.addLogMessage("Move mode - click tile or use movement keys", MessageCombat)
		return true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.mu.Lock()
		g.combatAction = CombatActionAttack
		g.mu.Unlock()
		g.addLogMessage("Attack mode - select target", MessageCombat)
		return true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.mu.Lock()
		g.combatAction = CombatActionCast
		g.mu.Unlock()
		g.executeCombatAction(CombatActionCast)
		return true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		g.mu.Lock()
		g.combatAction = CombatActionItem
		g.mu.Unlock()
		g.executeCombatAction(CombatActionItem)
		return true
	}
	// Game improvement #1: D → Defend (take defensive stance)
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.mu.Lock()
		g.combatAction = CombatActionDefend
		g.mu.Unlock()
		g.executeCombatAction(CombatActionDefend)
		return true
	}
	// Game improvement #1: F → Flee (attempt to escape combat)
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.mu.Lock()
		g.combatAction = CombatActionFlee
		g.mu.Unlock()
		g.executeCombatAction(CombatActionFlee)
		return true
	}
	return false
}

// handleCombatMovement processes 8-directional movement keys and touch swipes.
// Returns true if movement was processed.
func (g *Game) handleCombatMovement() bool {
	g.mu.RLock()
	action := g.combatAction
	g.mu.RUnlock()

	if action != CombatActionMove {
		return false
	}

	combatDirs := map[ebiten.Key]string{
		ebiten.KeyW: "north", ebiten.KeyArrowUp: "north", ebiten.KeyNumpad8: "north",
		ebiten.KeyS: "south", ebiten.KeyArrowDown: "south", ebiten.KeyNumpad2: "south",
		ebiten.KeyA: "west", ebiten.KeyArrowLeft: "west", ebiten.KeyNumpad4: "west",
		ebiten.KeyD: "east", ebiten.KeyArrowRight: "east", ebiten.KeyNumpad6: "east",
		ebiten.KeyQ: "northwest", ebiten.KeyNumpad7: "northwest",
		ebiten.KeyE: "northeast", ebiten.KeyNumpad9: "northeast",
		ebiten.KeyZ: "southwest", ebiten.KeyNumpad1: "southwest",
		ebiten.KeyNumpad3: "southeast",
	}

	for key, dir := range combatDirs {
		if inpututil.IsKeyJustPressed(key) {
			g.handleMove(dir)
			return true
		}
	}

	// Touch swipe for movement
	if swiped, dir := g.touchState.HasSwipe(); swiped {
		if d := SwipeDirection(dir); d != "" {
			g.handleMove(d)
			return true
		}
	}

	return false
}

// handleCombatTouchTap processes touch taps on the combat command menu.
func (g *Game) handleCombatTouchTap() {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return
	}

	// Combat command menu is drawn at panelY+45, height 45 (mirrors drawCommandMenu constants).
	panelY := g.screenHeight - actionPanelHeight
	const (
		menuYOffset = 45 // Offset from panelY where command menu starts
		menuHeight  = 45 // Total height of command menu panel (matches drawCommandMenu)
		titleWidth  = 75 // Width reserved for "COMMANDS:" label
		borderPad   = 5  // Border padding inside the panel
	)

	menuTop := panelY + menuYOffset
	if ty < menuTop || ty >= menuTop+menuHeight {
		return
	}

	// Compute per-command width to determine which command was tapped,
	// mirroring the layout in drawCommandMenu.
	panelWidth := g.screenWidth - charPanelWidth
	availWidth := panelWidth - titleWidth - borderPad*2
	if availWidth <= 0 {
		return
	}

	g.mu.RLock()
	currentAP := 0
	if g.player != nil {
		currentAP = g.player.AP
	}
	g.mu.RUnlock()

	commands := combatCommands(currentAP)
	cmdCount := len(commands)
	if cmdCount == 0 {
		return
	}

	cmdWidth := calcCmdWidth(availWidth, cmdCount)
	startX := titleWidth + borderPad

	for i, cmd := range commands {
		cmdX := startX + i*cmdWidth
		if tx < cmdX || tx >= cmdX+cmdWidth {
			continue
		}
		if !cmd.Available {
			return
		}
		switch {
		case cmd.Key == "Space":
			g.handleEndTurn()
		case cmd.Action != CombatActionNone:
			g.mu.Lock()
			g.combatAction = cmd.Action
			g.mu.Unlock()
			g.executeCombatAction(cmd.Action)
		}
		return
	}
}

// drawCombatScreen renders the combat interface (§5).
func (g *Game) drawCombatScreen(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 18, G: 15, B: 25, A: 255})

	// Combat grid viewport (left side)
	g.drawCombatGrid(screen)

	// Initiative tracker (right panel, §5.1)
	g.drawInitiativePanel(screen)

	// Combat log (bottom-left)
	g.drawCombatLog(screen)

	// Combat action bar (bottom)
	g.drawCombatActionBar(screen)
}

// drawCombatGrid renders the tactical grid.
func (g *Game) drawCombatGrid(screen *ebiten.Image) {
	gridWidth := g.screenWidth - charPanelWidth
	gridHeight := g.screenHeight - logPanelHeight - actionPanelHeight

	// Draw floor and grid lines
	g.drawCombatFloor(screen, gridWidth, gridHeight)

	// Draw movement/attack range highlighting
	g.drawCombatHighlights(screen, gridWidth, gridHeight)

	// Clean up expired flash effects and popups
	g.cleanupExpiredFlashes()
	g.cleanupExpiredSpellEffects()
	g.cleanupExpiredPopups()

	// Draw player and enemy entities
	g.drawCombatEntities(screen, gridWidth, gridHeight)

	// Draw spell effects overlay
	g.drawSpellEffects(screen, 0, 0, tileSize)

	// Draw damage number popups (on top of entities)
	g.drawDamagePopups(screen)

	// Draw turn change flash (screen border pulse for player turn)
	g.drawTurnChangeFlash(screen)

	// Combat round indicator
	g.mu.RLock()
	combat := g.combat
	combatAction := g.combatAction
	modifiers := g.targetModifiers
	g.mu.RUnlock()
	if combat != nil {
		drawColoredText(screen, fmt.Sprintf("Round %d", combat.Round), 10, 5, ColorGold)
	}

	// Draw cover/flanking indicators when in attack mode
	if combatAction == CombatActionAttack && modifiers != nil {
		g.drawCombatModifiers(screen, modifiers, gridWidth)
	}
}

// drawCombatFloor draws the background, floor tiles, and grid lines.
func (g *Game) drawCombatFloor(screen *ebiten.Image, gridWidth, gridHeight int) {
	drawRect(screen, 0, 0, gridWidth, gridHeight, color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Draw floor tiles for combat grid
	floorPath := TerrainTilePath("floor_stone", "dungeon")
	floorColor := color.RGBA{R: 50, G: 45, B: 60, A: 255}
	for y := 0; y < gridHeight; y += tileSize {
		for x := 0; x < gridWidth; x += tileSize {
			DrawSpriteWithFallback(screen, floorPath, x, y, tileSize, tileSize, floorColor)
		}
	}

	// Draw grid lines
	gridColor := color.RGBA{R: 40, G: 40, B: 55, A: 128}
	for x := 0; x < gridWidth; x += tileSize {
		drawLine(screen, x, 0, x, gridHeight, gridColor)
	}
	for y := 0; y < gridHeight; y += tileSize {
		drawLine(screen, 0, y, gridWidth, y, gridColor)
	}
}

// drawCombatHighlights draws movement and attack range highlights.
func (g *Game) drawCombatHighlights(screen *ebiten.Image, gridWidth, gridHeight int) {
	g.mu.RLock()
	combatAction := g.combatAction
	playerAP := 0
	if g.player != nil {
		playerAP = g.player.AP
	}
	g.mu.RUnlock()

	centerTileX := gridWidth / (2 * tileSize)
	centerTileY := gridHeight / (2 * tileSize)
	tilesWide := gridWidth / tileSize
	tilesHigh := gridHeight / tileSize

	if combatAction == CombatActionMove && playerAP > 0 {
		// Draw opportunity attack warning zones before movement highlights
		g.drawOpportunityZones(screen, tilesWide, tilesHigh, gridWidth, gridHeight)
		g.drawMovementHighlights(screen, centerTileX, centerTileY, playerAP, tilesWide, tilesHigh, gridWidth, gridHeight)
	}

	if combatAction == CombatActionAttack {
		g.drawAttackHighlights(screen, centerTileX, centerTileY, tilesWide, tilesHigh, gridWidth, gridHeight)
	}

	// Draw spell targeting highlights when in spell target mode
	if g.isSpellTargetMode() {
		g.drawSpellTargetHighlights(screen, gridWidth, gridHeight)
	}
}

// drawMovementHighlights draws blue tint on reachable tiles during move mode.
// Also shows cover indicators on tiles adjacent to obstacles/enemies.
func (g *Game) drawMovementHighlights(screen *ebiten.Image, centerX, centerY, ap, tilesW, tilesH, gridW, gridH int) {
	occupiedPositions := g.getOccupiedPositions(tileSize)
	moveRange := g.getMovementRange(centerX, centerY, ap, tilesW, tilesH, occupiedPositions)

	// Get tiles that provide cover (adjacent to obstacles/enemies)
	coverTiles := g.getCoverProvidingTiles(occupiedPositions, tilesW, tilesH)

	moveHighlightColor := color.RGBA{R: 74, G: 125, B: 191, A: 80}
	for _, pos := range moveRange {
		tx := pos.X * tileSize
		ty := pos.Y * tileSize
		if tx >= 0 && tx < gridW-tileSize && ty >= 0 && ty < gridH-tileSize {
			drawRect(screen, tx, ty, tileSize, tileSize, moveHighlightColor)

			// Draw cover indicator if this tile provides cover
			if coverTiles[pos] {
				g.drawCoverIcon(screen, tx, ty)
			}
		}
	}
}

// getCoverProvidingTiles returns tiles adjacent to obstacles that provide cover.
// Tiles next to enemies/obstacles provide tactical cover advantage.
func (g *Game) getCoverProvidingTiles(occupiedPositions map[Position]bool, maxX, maxY int) map[Position]bool {
	coverTiles := make(map[Position]bool)

	// Directions for adjacent tiles (8-directional)
	dirs := []struct{ dx, dy int }{
		{-1, -1},
		{0, -1},
		{1, -1},
		{-1, 0},
		{1, 0},
		{-1, 1},
		{0, 1},
		{1, 1},
	}

	// For each occupied position (enemy), mark adjacent tiles as providing cover
	for pos := range occupiedPositions {
		for _, d := range dirs {
			adjX := pos.X + d.dx
			adjY := pos.Y + d.dy

			// Check bounds
			if adjX < 0 || adjX >= maxX || adjY < 0 || adjY >= maxY {
				continue
			}

			adjPos := Position{X: adjX, Y: adjY}
			// Don't mark occupied tiles as cover tiles
			if !occupiedPositions[adjPos] {
				coverTiles[adjPos] = true
			}
		}
	}

	return coverTiles
}

// drawCoverIcon draws a small shield icon indicating cover availability.
func (g *Game) drawCoverIcon(screen *ebiten.Image, x, y int) {
	// Draw small shield icon in corner of tile (8x8 pixels)
	iconX := x + tileSize - 12
	iconY := y + 2
	iconSize := 10

	// Shield shape: slightly darker blue with gold outline
	shieldColor := color.RGBA{R: 50, G: 80, B: 140, A: 200}
	outlineColor := ColorGold

	// Simple shield rectangle
	drawRect(screen, iconX, iconY, iconSize, iconSize, shieldColor)
	drawRectOutline(screen, iconX, iconY, iconSize, iconSize, outlineColor)

	// Small "C" indicator for cover
	drawColoredText(screen, "C", iconX+2, iconY, color.RGBA{R: 255, G: 255, B: 255, A: 255})
}

// drawAttackHighlights draws red tint on attackable tiles during attack mode.
func (g *Game) drawAttackHighlights(screen *ebiten.Image, centerX, centerY, tilesW, tilesH, gridW, gridH int) {
	weaponRange := g.getEquippedWeaponRange()
	enemyPositions := g.getOccupiedPositions(tileSize)
	attackRange := g.getAttackRange(centerX, centerY, weaponRange, tilesW, tilesH)

	attackHighlightColor := color.RGBA{R: 191, G: 74, B: 74, A: 80}
	for _, pos := range attackRange {
		tx := pos.X * tileSize
		ty := pos.Y * tileSize
		if tx >= 0 && tx < gridW-tileSize && ty >= 0 && ty < gridH-tileSize {
			drawRect(screen, tx, ty, tileSize, tileSize, attackHighlightColor)

			// If this tile has an enemy, add pulsing gold border to indicate valid target
			if enemyPositions[pos] {
				g.drawPulsingBorder(screen, tx, ty, tileSize, tileSize)
			}
		}
	}
}

// drawCombatModifiers displays cover and flanking indicators during attack targeting.
// Shows cover type on the target and "FLANK" text if flanking bonus applies.
func (g *Game) drawCombatModifiers(screen *ebiten.Image, mods *CombatModifiers, gridWidth int) {
	// Display modifier panel in top-right of combat grid
	panelX := gridWidth - 150
	panelY := 5
	panelW := 140
	panelH := 50

	// Background panel
	drawRect(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 30, G: 30, B: 45, A: 220})
	drawRectOutline(screen, panelX, panelY, panelW, panelH, ColorGold)

	// Cover indicator
	coverText := "Cover: None"
	coverColor := ColorStatLabel
	switch mods.CoverType {
	case "half":
		coverText = fmt.Sprintf("Cover: Half (+%d AC)", mods.CoverBonus)
		coverColor = ColorEffectControl
	case "three_quarters":
		coverText = fmt.Sprintf("Cover: 3/4 (+%d AC)", mods.CoverBonus)
		coverColor = color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
	case "full":
		coverText = "Cover: FULL (Immune)"
		coverColor = ColorEnemyName
	}
	drawColoredText(screen, coverText, panelX+5, panelY+8, coverColor)

	// Flanking indicator
	if mods.IsFlanking {
		flankText := fmt.Sprintf("FLANK! (+%d Attack)", mods.FlankingBonus)
		drawColoredText(screen, flankText, panelX+5, panelY+28, ColorEffectBuff)
	} else {
		drawColoredText(screen, "No Flanking", panelX+5, panelY+28, ColorStatLabel)
	}
}

// drawCombatEntities draws player and enemy tokens on the combat grid.
func (g *Game) drawCombatEntities(screen *ebiten.Image, gridWidth, gridHeight int) {
	g.mu.RLock()
	player := g.player
	combat := g.combat
	g.mu.RUnlock()

	isPlayerTurn := combat != nil && combat.IsPlayerTurn

	// Draw player token
	if player != nil {
		px := (gridWidth / 2) - (tileSize / 2)
		py := (gridHeight / 2) - (tileSize / 2)
		g.drawPlayerToken(screen, player, px, py, isPlayerTurn)
	}

	// Draw enemy tokens from initiative
	if combat != nil {
		g.drawEnemyTokens(screen, combat, gridWidth)
	}
}

// getEntityScreenPos returns the screen coordinates for an entity in combat.
// Returns (x, y, found) where found indicates if the entity was located.
func (g *Game) getEntityScreenPos(entityID string) (int, int, bool) {
	g.mu.RLock()
	player := g.player
	combat := g.combat
	gridWidth := g.screenWidth - charPanelWidth
	gridHeight := g.screenHeight - logPanelHeight - actionPanelHeight
	g.mu.RUnlock()

	// Check if it's the player
	if player != nil && player.ID == entityID {
		px := (gridWidth / 2) - (tileSize / 2)
		py := (gridHeight / 2) - (tileSize / 2)
		return px + tileSize/2, py, true
	}

	// Check enemies in initiative
	if combat != nil {
		enemyIdx := 0
		for _, entry := range combat.Initiative {
			if !entry.IsPlayer {
				if entry.ID == entityID {
					ex := 100 + enemyIdx*tileSize*2
					ey := 50
					return ex + tileSize/2, ey, true
				}
				enemyIdx++
			}
		}
	}
	return 0, 0, false
}

// drawPlayerToken renders the player character on the combat grid.
func (g *Game) drawPlayerToken(screen *ebiten.Image, player *PlayerState, px, py int, isPlayerTurn bool) {
	if isPlayerTurn {
		g.drawPulsingBorder(screen, px-2, py-2, tileSize+2, tileSize+2)
	}

	spritePath := g.getPlayerSpritePath(player)
	DrawSpriteWithFallback(screen, spritePath, px, py, tileSize-2, tileSize-2,
		color.RGBA{R: 80, G: 200, B: 80, A: 255})

	// Show "P" indicator while sprite loads
	initSpriteCache()
	if !spriteCache.IsCached(spritePath) {
		drawColoredText(screen, "P", px+10, py+8, ColorPlayerName)
	}

	// Draw damage/heal flash overlay for player
	if flash := g.getFlashForEntity(player.ID); flash != nil {
		g.drawFlashOverlay(screen, px, py, tileSize-2, tileSize-2, flash)
	}

	// Draw effect indicators above the token
	if len(player.Effects) > 0 {
		g.drawEffectIndicators(screen, player.Effects, px, py-12, tileSize)
	}
}

// drawEnemyTokens renders enemy characters on the combat grid.
func (g *Game) drawEnemyTokens(screen *ebiten.Image, combat *CombatState, gridWidth int) {
	enemyIdx := 0
	for _, entry := range combat.Initiative {
		if !entry.IsPlayer {
			ex := 100 + enemyIdx*tileSize*2
			ey := 50
			if ex < gridWidth-tileSize {
				g.drawSingleEnemyToken(screen, entry, ex, ey, combat.CurrentTurn)
			}
			enemyIdx++
		}
	}
}

// drawSingleEnemyToken renders one enemy token with turn indicator and flash effects.
func (g *Game) drawSingleEnemyToken(screen *ebiten.Image, entry InitiativeEntry, ex, ey int, currentTurn string) {
	isEnemyTurn := entry.ID == currentTurn
	if isEnemyTurn {
		g.drawPulsingBorder(screen, ex-2, ey-2, tileSize+2, tileSize+2)
	}

	monsterPath := MonsterSpritePath(entry.Name)
	DrawSpriteWithFallback(screen, monsterPath, ex, ey, tileSize-2, tileSize-2,
		color.RGBA{R: 200, G: 80, B: 80, A: 255})

	// Show "E" indicator while sprite loads
	initSpriteCache()
	if !spriteCache.IsCached(monsterPath) {
		drawColoredText(screen, "E", ex+10, ey+8, ColorEnemyName)
	}

	// Draw damage/heal flash overlay for enemy
	if flash := g.getFlashForEntity(entry.ID); flash != nil {
		g.drawFlashOverlay(screen, ex, ey, tileSize-2, tileSize-2, flash)
	}

	// Draw effect indicators above the enemy token
	if len(entry.Effects) > 0 {
		g.drawEffectIndicators(screen, entry.Effects, ex, ey-12, tileSize)
	}
}

// drawPulsingBorder draws a pulsing gold border around an entity tile.
// Used to indicate the current turn character in combat.
func (g *Game) drawPulsingBorder(screen *ebiten.Image, x, y, w, h int) {
	// Calculate pulse alpha using sine wave (~1 Hz pulse rate)
	// time.Now().UnixMilli() gives milliseconds; 1000ms = 1 full cycle
	pulsePhase := float64(time.Now().UnixMilli()%1000) / 1000.0 * 2 * 3.14159
	alpha := 0.5 + 0.5*((1+pulseFloat64Sin(pulsePhase))/2) // Range: 0.5 to 1.0

	borderColor := color.RGBA{
		R: ColorGoldHi.R,
		G: ColorGoldHi.G,
		B: ColorGoldHi.B,
		A: uint8(alpha * 255),
	}

	// Draw 2px thick border
	drawRectOutline(screen, x, y, w, h, borderColor)
	drawRectOutline(screen, x+1, y+1, w-2, h-2, borderColor)
}

// pulseFloat64Sin returns sin(x) for float64, used for pulsing effects.
func pulseFloat64Sin(x float64) float64 {
	// Simple Taylor series approximation for sin
	// Sufficient for visual pulsing effects
	x = x - float64(int(x/(2*3.14159)))*2*3.14159 // Normalize to [0, 2π)
	if x > 3.14159 {
		x -= 2 * 3.14159
	}
	// sin(x) ≈ x - x³/6 + x⁵/120 for small x
	x3 := x * x * x
	x5 := x3 * x * x
	return x - x3/6 + x5/120
}

// drawFlashOverlay renders a semi-transparent colored overlay for damage/heal effects.
func (g *Game) drawFlashOverlay(screen *ebiten.Image, x, y, w, h int, flash *DamageFlash) {
	alpha := flash.Alpha()
	if alpha <= 0 {
		return
	}
	// Create flash color with calculated alpha
	flashColor := color.RGBA{
		R: flash.Color.R,
		G: flash.Color.G,
		B: flash.Color.B,
		A: uint8(alpha * 255),
	}
	drawRect(screen, x, y, w, h, flashColor)
}

// drawEffectIndicators renders small colored squares above a token for each active effect.
// Shows up to 4 effect icons; displays "+" if there are more.
func (g *Game) drawEffectIndicators(screen *ebiten.Image, effects []EffectData, x, y, maxWidth int) {
	if len(effects) == 0 {
		return
	}

	iconSize := 8
	iconSpacing := 2
	maxIcons := 4
	startX := x

	for i, effect := range effects {
		if i >= maxIcons {
			// Draw overflow indicator
			drawColoredText(screen, "+", startX, y, ColorGold)
			break
		}

		effectColor := getEffectColor(effect.Type)
		drawRect(screen, startX, y, iconSize, iconSize, effectColor)
		// Draw a small outline for visibility
		drawRectOutline(screen, startX, y, iconSize, iconSize, color.RGBA{R: 255, G: 255, B: 255, A: 128})
		startX += iconSize + iconSpacing
	}
}

// getEffectColor returns the appropriate color for an effect type.
// Uses Gold Box palette: debuffs=red, CC=yellow, buffs=green.
func getEffectColor(effectType string) color.RGBA {
	switch effectType {
	case "damage", "damage_over_time", "dot", "burning", "bleeding", "poison":
		return ColorEffectDebuff
	case "stun", "root", "slow", "control", "held", "paralyzed", "paralysis", "cc":
		return ColorEffectControl
	case "heal", "heal_over_time", "hot", "buff", "boost", "blessing", "regeneration", "stat_boost", "haste":
		return ColorEffectBuff
	default:
		return ColorEffectDefault
	}
}

// drawInitiativePanel renders the initiative tracker on the right side (§5.1).
func (g *Game) drawInitiativePanel(screen *ebiten.Image) {
	panelX := g.screenWidth - charPanelWidth
	panelHeight := g.screenHeight - actionPanelHeight

	g.drawInitiativePanelBackground(screen, panelX, panelHeight)

	g.mu.RLock()
	combat := g.combat
	player := g.player
	g.mu.RUnlock()

	if combat == nil {
		drawColoredText(screen, "No combat", panelX+50, 40, ColorStatLabel)
		return
	}

	g.drawCombatRoundInfo(screen, combat, panelX)
	g.drawInitiativeList(screen, combat, panelX)
	g.drawPlayerStatsSummary(screen, player, panelX, panelHeight)
}

// drawInitiativePanelBackground renders the panel background and border.
func (g *Game) drawInitiativePanelBackground(screen *ebiten.Image, panelX, panelHeight int) {
	drawRect(screen, panelX, 0, charPanelWidth, panelHeight, color.RGBA{R: 30, G: 28, B: 42, A: 255})
	drawBoldPanelBorder(screen, panelX, 0, charPanelWidth, panelHeight)
	// Gold Box-style centered header
	drawPanelHeader(screen, panelX, 0, charPanelWidth, "INITIATIVE")
}

// drawCombatRoundInfo displays the current round and turn information.
func (g *Game) drawCombatRoundInfo(screen *ebiten.Image, combat *CombatState, panelX int) {
	drawColoredText(screen, fmt.Sprintf("Round: %d", combat.Round), panelX+10, 35, ColorStatValue)
	if combat.CurrentTurn != "" {
		drawColoredText(screen, fmt.Sprintf("Turn: %s", truncateText(combat.CurrentTurn, 15)), panelX+10, 50, ColorStatLabel)
	}
	drawLine(screen, panelX+10, 67, panelX+charPanelWidth-10, 67, ColorPanelBorder)
}

// drawInitiativeList renders the initiative order for combatants.
func (g *Game) drawInitiativeList(screen *ebiten.Image, combat *CombatState, panelX int) {
	y := 75
	nextTurnID := g.findNextTurnID(combat)

	for i, entry := range combat.Initiative {
		if i >= 10 {
			break
		}
		isNext := entry.ID == nextTurnID && entry.ID != combat.CurrentTurn
		g.drawInitiativeEntry(screen, entry, combat.CurrentTurn, isNext, panelX, y)
		y += 20
	}
}

// findNextTurnID determines which combatant will act after the current turn.
func (g *Game) findNextTurnID(combat *CombatState) string {
	if len(combat.Initiative) == 0 {
		return ""
	}
	for i, entry := range combat.Initiative {
		if entry.ID == combat.CurrentTurn && i+1 < len(combat.Initiative) {
			return combat.Initiative[i+1].ID
		}
	}
	return combat.Initiative[0].ID // Wrap around
}

// drawInitiativeEntry renders a single combatant in the initiative list.
func (g *Game) drawInitiativeEntry(screen *ebiten.Image, entry InitiativeEntry, currentTurn string, isNext bool, panelX, y int) {
	marker := g.getInitiativeMarker(entry.ID, currentTurn, isNext)
	nameColor := g.getInitiativeNameColor(entry, currentTurn)
	g.drawInitiativeBackground(screen, entry.ID, currentTurn, isNext, panelX, y)

	// Draw behavior icon for NPCs (before name)
	behaviorIcon := ""
	if !entry.IsPlayer && entry.BehaviorType != "" {
		behaviorIcon = getBehaviorIcon(entry.BehaviorType)
	}

	// Draw name with marker and behavior icon
	nameText := fmt.Sprintf("%s%s%s", marker, behaviorIcon, truncateText(entry.Name, 9))
	drawColoredText(screen, nameText, panelX+10, y, nameColor)

	// Draw initiative value, morale, and HP bar
	g.drawInitiativeStats(screen, entry, panelX, y)
}

// getBehaviorIcon returns an icon character for an NPC behavior type.
func getBehaviorIcon(behaviorType string) string {
	icons := map[string]string{
		"aggressive": "!",  // Aggressive - attacks on sight
		"guard":      "G ", // Guard - defends position
		"patrol":     "P ", // Patrol - moves along route
		"coward":     "F ", // Coward/Flee - runs when injured
		"support":    "S ", // Support - heals/buffs allies
		"ranged":     "R ", // Ranged - prefers distance
		"berserker":  "B ", // Berserker - reckless attacks
		"tactical":   "T ", // Tactical - uses cover/flanking
	}
	if icon, ok := icons[behaviorType]; ok {
		return icon
	}
	return ""
}

// getInitiativeMarker returns the turn indicator symbol.
func (g *Game) getInitiativeMarker(entryID, currentTurn string, isNext bool) string {
	if entryID == currentTurn {
		return "> "
	}
	if isNext {
		return "* "
	}
	return "  "
}

// getInitiativeNameColor returns the appropriate color for an initiative entry name.
func (g *Game) getInitiativeNameColor(entry InitiativeEntry, currentTurn string) color.RGBA {
	nameColor := ColorEnemyName
	if entry.IsPlayer {
		nameColor = ColorPlayerName
	}
	if entry.ID == currentTurn {
		nameColor = brightenColor(nameColor, 60)
	}
	return nameColor
}

// drawInitiativeBackground draws the background highlight for initiative entries.
func (g *Game) drawInitiativeBackground(screen *ebiten.Image, entryID, currentTurn string, isNext bool, panelX, y int) {
	if entryID == currentTurn {
		drawRect(screen, panelX+5, y-1, charPanelWidth-10, 18, color.RGBA{R: 70, G: 60, B: 20, A: 255})
	} else if isNext {
		drawRect(screen, panelX+5, y-1, charPanelWidth-10, 18, color.RGBA{R: 40, G: 45, B: 55, A: 255})
	}
}

// drawInitiativeStats draws initiative value, morale indicator, and HP bar.
func (g *Game) drawInitiativeStats(screen *ebiten.Image, entry InitiativeEntry, panelX, y int) {
	if entry.Initiative > 0 {
		drawColoredText(screen, fmt.Sprintf("%d", entry.Initiative), panelX+charPanelWidth-90, y, ColorStatLabel)
	}
	if !entry.IsPlayer && entry.MoraleState != "" {
		if moraleIcon, moraleColor := getMoraleIndicator(entry.MoraleState); moraleIcon != "" {
			drawColoredText(screen, moraleIcon, panelX+charPanelWidth-75, y, moraleColor)
		}
	}
	if entry.MaxHP > 0 {
		barX, barW := panelX+charPanelWidth-60, 55
		pct := float64(entry.HP) / float64(entry.MaxHP)
		drawRect(screen, barX, y, barW, 10, color.RGBA{R: 60, G: 20, B: 20, A: 255})
		drawRect(screen, barX, y, int(pct*float64(barW)), 10, hpBarColor(pct))
	}
}

// drawPlayerStatsSummary renders the player's stats at the bottom of the panel.
func (g *Game) drawPlayerStatsSummary(screen *ebiten.Image, player *PlayerState, panelX, panelHeight int) {
	if player == nil {
		return
	}
	y := panelHeight - 100
	drawColoredText(screen, player.Name, panelX+10, y, ColorPlayerName)
	g.drawHPBar(screen, panelX, y-65, player)
	g.drawAPBar(screen, panelX, y+20, player)
}

// drawCombatActionBar renders the bottom action bar for combat (§5.2).
// Game improvement #1: Authentic Gold Box-style command menu with AP indicators.
func (g *Game) drawCombatActionBar(screen *ebiten.Image) {
	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth

	// Panel background with deeper, more authentic Gold Box color
	drawRect(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 22, G: 20, B: 32, A: 255})
	// Bold Gold Box-style panel border
	drawBoldPanelBorder(screen, 0, panelY, panelWidth, actionPanelHeight)

	g.mu.RLock()
	currentAction := g.combatAction
	combat := g.combat
	player := g.player
	g.mu.RUnlock()

	// Get current AP for affordability check and display
	currentAP := 0
	maxAP := 3 // Default max AP
	if player != nil {
		currentAP = player.AP
		if player.MaxAP > 0 {
			maxAP = player.MaxAP
		}
	}

	// Turn indicator per §5 — "YOUR TURN" / "Waiting..." with proper color
	turnLabel := "Waiting..."
	turnColor := color.RGBA{R: 180, G: 140, B: 60, A: 255}
	if combat != nil && combat.IsPlayerTurn {
		turnLabel = "YOUR TURN"
		turnColor = color.RGBA{R: 80, G: 220, B: 80, A: 255}
	}
	drawColoredText(screen, turnLabel, 10, panelY+8, turnColor)

	// AP indicator next to turn label
	drawAPIndicator(screen, 120, panelY+8, currentAP, maxAP)

	// Round indicator
	if combat != nil {
		roundText := fmt.Sprintf("Round %d", combat.Round)
		drawColoredText(screen, roundText, 260, panelY+8, ColorGold)
	}

	// Draw Gold Box-style command menu
	g.drawCombatCommandMenu(screen)

	// Current action status line
	actionStatus := "Select action..."
	if currentAction != CombatActionNone {
		actionStatus = fmt.Sprintf("Action: %s - confirm target or press again to cancel", currentAction)
	}
	drawColoredText(screen, actionStatus, 10, panelY+30, ColorStatLabel)
}

// executeCombatAction dispatches the selected combat action via RPC.
func (g *Game) executeCombatAction(action CombatAction) {
	switch action {
	case CombatActionMove:
		g.addLogMessage("Move mode - click tile or use movement keys", MessageCombat)
	case CombatActionAttack:
		// Enter attack mode; player must explicitly select and confirm a target.
		g.addLogMessage("Attack mode - select target (Tab to cycle)", MessageCombat)
	case CombatActionCast:
		g.mu.Lock()
		g.previousMode = g.mode
		g.mode = ModeSpellcasting
		g.mu.Unlock()
		go g.loadSpells()
	case CombatActionItem:
		g.mu.Lock()
		g.previousMode = g.mode
		g.mode = ModeInventory
		g.mu.Unlock()
		go g.loadInventory()
	case CombatActionDefend:
		g.addLogMessage("Defending this turn — AC bonus active", MessageCombat)
		g.handleEndTurn()
	case CombatActionFlee:
		g.addLogMessage("Attempting to flee...", MessageCombat)
		go func() {
			result, err := g.rpcClient.Move("flee")
			if err != nil {
				g.showError(fmt.Sprintf("Flee failed: %v", err))
				return
			}
			if result.Success {
				g.mu.Lock()
				g.mode = ModeNormal
				g.screenState = ScreenExploration
				g.mu.Unlock()
				g.addLogMessage("Fled from combat!", MessageCombat)
			} else {
				g.addLogMessage("Cannot flee — enemies block your escape!", MessageWarning)
			}
		}()
	}
}

// executeAttack performs an attack via RPC and narrates the result in Gold Box style.
func (g *Game) executeAttack(attackerName, targetID, targetName string) {
	result, err := g.rpcClient.Attack(targetID, "")
	if err != nil {
		g.addLogMessage(fmt.Sprintf("%s attacks %s...", attackerName, targetName), MessageCombat)
		g.addLogMessage(fmt.Sprintf("  Attack failed: %v", err), MessageError)
		return
	}

	narrationAttacker, narrationTarget := g.getNarrationNames(attackerName, targetName, result)
	rollInfo := g.buildRollInfo(result)

	// Server message takes priority; fallback to client-side construction
	if result.Message != "" {
		g.addLogMessage(result.Message+rollInfo, MessageCombat)
	} else {
		g.addLogMessage(g.buildAttackNarration(result, narrationAttacker, narrationTarget, rollInfo), MessageCombat)
	}

	// Visual effects on hit
	if result.Success || result.Hit {
		g.addDamageFlash(targetID, ColorEnemyName)

		// Add floating damage popup
		if result.Damage > 0 {
			if x, y, found := g.getEntityScreenPos(targetID); found {
				g.addDamagePopup(x, y, result.Damage, false, result.IsCritical)
			}
		}

		if result.TargetHealth >= 0 {
			g.addLogMessage(fmt.Sprintf("  %s: %d HP remaining", narrationTarget, result.TargetHealth), MessageInfo)
		}
	}
}

// getNarrationNames returns the names to use for attack narration, preferring server-provided names.
func (g *Game) getNarrationNames(clientAttacker, clientTarget string, result *AttackResult) (string, string) {
	attacker := clientAttacker
	if result.AttackerName != "" {
		attacker = result.AttackerName
	}
	target := clientTarget
	if result.TargetName != "" {
		target = result.TargetName
	}
	return attacker, target
}

// buildRollInfo constructs the roll details string for attack narration.
func (g *Game) buildRollInfo(result *AttackResult) string {
	if result.AttackRoll > 0 && result.TargetAC > 0 {
		return fmt.Sprintf(" (%d vs AC %d)", result.AttackRoll, result.TargetAC)
	}
	return ""
}

// buildAttackNarration constructs the attack message for legacy servers that don't provide a message.
func (g *Game) buildAttackNarration(result *AttackResult, attacker, target, rollInfo string) string {
	if !result.Success && !result.Hit {
		return fmt.Sprintf("%s attacks %s — MISS%s", attacker, target, rollInfo)
	}
	if result.IsCritical {
		return fmt.Sprintf("%s CRITICAL HIT on %s for %d damage!!%s", attacker, target, result.Damage, rollInfo)
	}
	if result.Damage > 0 {
		return fmt.Sprintf("%s hits %s for %d damage%s", attacker, target, result.Damage, rollInfo)
	}
	return fmt.Sprintf("%s hits %s%s", attacker, target, rollInfo)
}

// addDamageFlash adds a visual flash effect for an entity (damage=red, heal=green).
func (g *Game) addDamageFlash(entityID string, flashColor color.RGBA) {
	flash := DamageFlash{
		EntityID:  entityID,
		StartTime: time.Now(),
		Duration:  200 * time.Millisecond,
		Color:     flashColor,
	}
	g.mu.Lock()
	g.damageFlashes = append(g.damageFlashes, flash)
	g.mu.Unlock()
}

// addHealFlash adds a green flash effect when an entity is healed.
func (g *Game) addHealFlash(entityID string) {
	g.addDamageFlash(entityID, ColorPlayerName) // Green for healing
}

// addDamagePopup adds a floating damage number at the specified screen position.
func (g *Game) addDamagePopup(x, y, amount int, isHeal, isCrit bool) {
	popup := DamagePopup{
		X:         x,
		Y:         y,
		Amount:    amount,
		IsHeal:    isHeal,
		IsCrit:    isCrit,
		StartTime: time.Now(),
		Duration:  800 * time.Millisecond,
	}
	g.mu.Lock()
	g.damagePopups = append(g.damagePopups, popup)
	g.mu.Unlock()
}

// cleanupExpiredFlashes removes flash effects that have finished.
func (g *Game) cleanupExpiredFlashes() {
	g.mu.Lock()
	defer g.mu.Unlock()
	active := g.damageFlashes[:0]
	for _, f := range g.damageFlashes {
		if f.IsActive() {
			active = append(active, f)
		}
	}
	g.damageFlashes = active
}

// cleanupExpiredPopups removes damage popups that have finished.
func (g *Game) cleanupExpiredPopups() {
	g.mu.Lock()
	defer g.mu.Unlock()
	active := g.damagePopups[:0]
	for _, p := range g.damagePopups {
		if p.IsActive() {
			active = append(active, p)
		}
	}
	g.damagePopups = active
}

// getFlashForEntity returns the flash effect for an entity, if any active.
func (g *Game) getFlashForEntity(entityID string) *DamageFlash {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for i := range g.damageFlashes {
		if g.damageFlashes[i].EntityID == entityID && g.damageFlashes[i].IsActive() {
			return &g.damageFlashes[i]
		}
	}
	return nil
}

// addSpellEffect adds a visual spell effect at the given position.
func (g *Game) addSpellEffect(spellID, school string, targetPos Position) {
	effect := SpellEffect{
		SpellID:     spellID,
		SpellSchool: school,
		TargetPos:   targetPos,
		StartTime:   time.Now(),
		Duration:    400 * time.Millisecond,
		TotalFrames: 4,
	}
	g.mu.Lock()
	g.spellEffects = append(g.spellEffects, effect)
	g.mu.Unlock()
}

// cleanupExpiredSpellEffects removes spell effects that have finished.
func (g *Game) cleanupExpiredSpellEffects() {
	g.mu.Lock()
	defer g.mu.Unlock()
	active := g.spellEffects[:0]
	for _, e := range g.spellEffects {
		if e.IsActive() {
			active = append(active, e)
		}
	}
	g.spellEffects = active
}

// drawSpellEffects renders all active spell effects on the combat grid.
func (g *Game) drawSpellEffects(screen *ebiten.Image, gridX, gridY, tileSize int) {
	g.mu.RLock()
	effects := make([]SpellEffect, len(g.spellEffects))
	copy(effects, g.spellEffects)
	g.mu.RUnlock()

	for _, effect := range effects {
		if !effect.IsActive() {
			continue
		}

		// Calculate screen position (centered on tile)
		screenX := gridX + effect.TargetPos.X*tileSize + tileSize/2
		screenY := gridY + effect.TargetPos.Y*tileSize + tileSize/2

		// Get effect alpha for fade-out
		alpha := effect.GetAlpha()

		// Try to use sprite-based effect if available
		spritePath := SpellEffectPath(effect.SpellID)
		if spriteCache != nil && spriteCache.IsCached(spritePath) {
			if sprite := spriteCache.Get(spritePath); sprite != nil {
				g.drawSpellSprite(screen, sprite, screenX, screenY, tileSize, alpha)
				continue
			}
		}

		// Trigger async load for future frames
		if spriteCache != nil {
			spriteCache.Get(spritePath)
		}

		// Fallback to procedural expanding circle effect
		effectColor := SpellSchoolColor(effect.SpellSchool)
		g.drawProceduralSpellEffect(screen, screenX, screenY, effect.GetRadius(), effectColor, alpha)
	}
}

// drawSpellSprite draws a spell effect sprite at the given position with alpha.
func (g *Game) drawSpellSprite(screen, sprite *ebiten.Image, cx, cy, tileSize int, alpha float32) {
	sw, sh := sprite.Bounds().Dx(), sprite.Bounds().Dy()

	// Scale sprite to fit tile size
	scaleX := float64(tileSize) / float64(sw)
	scaleY := float64(tileSize) / float64(sh)

	// Center on target position
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scaleX, scaleY)
	opts.GeoM.Translate(float64(cx-tileSize/2), float64(cy-tileSize/2))
	opts.ColorScale.ScaleAlpha(alpha)

	screen.DrawImage(sprite, opts)
}

// drawProceduralSpellEffect draws the fallback expanding circle effect.
func (g *Game) drawProceduralSpellEffect(screen *ebiten.Image, cx, cy int, radius float64, effectColor color.RGBA, alpha float32) {
	// Adjust color alpha
	c := color.RGBA{
		R: effectColor.R,
		G: effectColor.G,
		B: effectColor.B,
		A: uint8(float32(effectColor.A) * alpha),
	}

	// Draw effect as expanding circle
	g.drawSpellCircle(screen, cx, cy, radius, c)

	// Draw inner brighter core
	coreColor := color.RGBA{
		R: uint8(min(int(effectColor.R)+80, 255)),
		G: uint8(min(int(effectColor.G)+80, 255)),
		B: uint8(min(int(effectColor.B)+80, 255)),
		A: uint8(float32(200) * alpha),
	}
	g.drawSpellCircle(screen, cx, cy, radius*0.5, coreColor)
}

// drawSpellCircle draws an approximated circle using filled rectangles.
func (g *Game) drawSpellCircle(screen *ebiten.Image, cx, cy int, radius float64, c color.RGBA) {
	// Draw filled circle using concentric rings of rectangles
	r := int(radius)
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			// Check if point is within circle
			dist := float64(dx*dx+dy*dy) / (radius * radius)
			if dist <= 1.0 {
				// Alpha falloff at edges
				edgeAlpha := 1.0 - dist*0.5
				pixelColor := color.RGBA{
					R: c.R,
					G: c.G,
					B: c.B,
					A: uint8(float64(c.A) * edgeAlpha),
				}
				drawRect(screen, cx+dx, cy+dy, 1, 1, pixelColor)
			}
		}
	}
}

// drawDamagePopups renders floating damage numbers above combat grid.
func (g *Game) drawDamagePopups(screen *ebiten.Image) {
	g.mu.RLock()
	popups := make([]DamagePopup, len(g.damagePopups))
	copy(popups, g.damagePopups)
	g.mu.RUnlock()

	for _, popup := range popups {
		if !popup.IsActive() {
			continue
		}

		// Calculate position with float upward
		yOffset := popup.YOffset()
		x := popup.X
		y := popup.Y - yOffset

		// Format text
		text := fmt.Sprintf("%d", popup.Amount)
		if popup.IsCrit {
			text = "!" + text + "!"
		}

		// Choose color based on damage/heal and apply alpha
		alpha := popup.Alpha()
		var textColor color.RGBA
		if popup.IsHeal {
			textColor = color.RGBA{R: 60, G: 220, B: 60, A: uint8(alpha * 255)} // Green for healing
		} else if popup.IsCrit {
			textColor = color.RGBA{R: 255, G: 200, B: 50, A: uint8(alpha * 255)} // Gold for crits
		} else {
			textColor = color.RGBA{R: 220, G: 60, B: 60, A: uint8(alpha * 255)} // Red for damage
		}

		// Draw text (centered on x position)
		drawColoredText(screen, text, x-len(text)*3, y, textColor)
	}
}

// drawTurnChangeFlash draws a screen border pulse when the turn changes to player.
func (g *Game) drawTurnChangeFlash(screen *ebiten.Image) {
	g.mu.RLock()
	flashStart := g.turnChangeFlash
	flashDur := g.turnChangeFlashDur
	g.mu.RUnlock()

	if flashStart.IsZero() {
		return
	}

	elapsed := time.Since(flashStart)
	if elapsed > flashDur {
		return
	}

	// Calculate alpha: peak at 25%, then fade out
	progress := float64(elapsed) / float64(flashDur)
	var alpha float64
	if progress < 0.25 {
		alpha = progress * 4 * 150 // Ramp up to 150
	} else {
		alpha = (1.0 - (progress-0.25)/0.75) * 150 // Fade from 150 to 0
	}

	// Draw glowing border around screen
	borderColor := color.RGBA{R: 255, G: 215, B: 0, A: uint8(alpha)} // Gold color
	borderThickness := 4

	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Top border
	drawRect(screen, 0, 0, w, borderThickness, borderColor)
	// Bottom border
	drawRect(screen, 0, h-borderThickness, w, borderThickness, borderColor)
	// Left border
	drawRect(screen, 0, 0, borderThickness, h, borderColor)
	// Right border
	drawRect(screen, w-borderThickness, 0, borderThickness, h, borderColor)
}

// getMovementRange calculates all tiles reachable with the given AP.
// Uses Manhattan distance: range = AP * 2 tiles.
// Excludes the player's current position and tiles occupied by enemies/walls.
func (g *Game) getMovementRange(centerX, centerY, ap, maxX, maxY int, occupied map[Position]bool) []Position {
	if ap <= 0 {
		return nil
	}

	moveRange := ap * 2 // Each AP allows 2 tiles of movement
	var result []Position

	// Generate all positions within Manhattan distance
	for dx := -moveRange; dx <= moveRange; dx++ {
		for dy := -moveRange; dy <= moveRange; dy++ {
			// Skip center (player position)
			if dx == 0 && dy == 0 {
				continue
			}

			// Check Manhattan distance
			dist := intAbs(dx) + intAbs(dy)
			if dist > moveRange {
				continue
			}

			tx := centerX + dx
			ty := centerY + dy

			// Check bounds
			if tx < 0 || tx >= maxX || ty < 0 || ty >= maxY {
				continue
			}

			// Skip tiles occupied by enemies or walls
			pos := Position{X: tx, Y: ty}
			if occupied != nil && occupied[pos] {
				continue
			}

			result = append(result, pos)
		}
	}

	return result
}

// intAbs returns the absolute value of an integer.
func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// getOccupiedPositions returns a map of tile positions occupied by enemies.
// Uses the same enemy placement logic as drawCombatGrid for consistency.
func (g *Game) getOccupiedPositions(tileSize int) map[Position]bool {
	g.mu.RLock()
	combat := g.combat
	g.mu.RUnlock()

	occupied := make(map[Position]bool)

	if combat == nil {
		return occupied
	}

	// Calculate enemy positions using the same logic as drawCombatGrid
	// Enemies are placed at: ex = 100 + enemyIdx*tileSize*2, ey = 50
	enemyIdx := 0
	for _, entry := range combat.Initiative {
		if !entry.IsPlayer {
			ex := 100 + enemyIdx*tileSize*2
			ey := 50
			// Convert pixel position to tile position
			tileX := ex / tileSize
			tileY := ey / tileSize
			occupied[Position{X: tileX, Y: tileY}] = true
			enemyIdx++
		}
	}

	return occupied
}

// getEnemyThreatZones calculates tiles adjacent to enemies that would trigger opportunity attacks.
// In D&D terms, moving out of an enemy's threatened squares can provoke attacks of opportunity.
func (g *Game) getEnemyThreatZones(tileSize int) map[Position]bool {
	g.mu.RLock()
	combat := g.combat
	g.mu.RUnlock()

	threatZones := make(map[Position]bool)

	if combat == nil {
		return threatZones
	}

	// Directions for adjacent tiles (8-directional for melee threat)
	dirs := []struct{ dx, dy int }{
		{-1, -1},
		{0, -1},
		{1, -1},
		{-1, 0},
		{1, 0},
		{-1, 1},
		{0, 1},
		{1, 1},
	}

	// Calculate enemy positions and their threat zones
	enemyIdx := 0
	for _, entry := range combat.Initiative {
		if !entry.IsPlayer {
			ex := 100 + enemyIdx*tileSize*2
			ey := 50
			// Convert pixel position to tile position
			enemyTileX := ex / tileSize
			enemyTileY := ey / tileSize

			// Mark all adjacent tiles as threatened
			for _, d := range dirs {
				adjX := enemyTileX + d.dx
				adjY := enemyTileY + d.dy

				// Only mark if within reasonable bounds
				if adjX >= 0 && adjY >= 0 {
					threatZones[Position{X: adjX, Y: adjY}] = true
				}
			}
			enemyIdx++
		}
	}

	return threatZones
}

// drawOpportunityZones draws warning indicators on tiles that would provoke opportunity attacks.
func (g *Game) drawOpportunityZones(screen *ebiten.Image, tilesW, tilesH, gridW, gridH int) {
	threatZones := g.getEnemyThreatZones(tileSize)
	occupiedPositions := g.getOccupiedPositions(tileSize)

	// Orange-ish warning color with transparency
	threatColor := color.RGBA{R: 255, G: 140, B: 0, A: 40}
	warningIconColor := color.RGBA{R: 255, G: 180, B: 0, A: 200}

	for pos := range threatZones {
		// Don't draw on occupied tiles (enemies themselves)
		if occupiedPositions[pos] {
			continue
		}

		tx := pos.X * tileSize
		ty := pos.Y * tileSize

		if tx >= 0 && tx < gridW && ty >= 0 && ty < gridH {
			// Draw warning tint
			drawRect(screen, tx, ty, tileSize, tileSize, threatColor)

			// Draw small "!" warning icon in corner
			iconX := tx + 2
			iconY := ty + 2
			drawColoredText(screen, "!", iconX, iconY, warningIconColor)
		}
	}
}

// getAttackRange calculates all tiles within attack range.
// Uses Chebyshev distance (8-directional) for melee weapons.
// weaponRange: 1 = melee (adjacent 8 tiles), 2+ = ranged weapons.
func (g *Game) getAttackRange(centerX, centerY, weaponRange, maxX, maxY int) []Position {
	if weaponRange <= 0 {
		return nil
	}

	var result []Position

	// Generate all positions within Chebyshev distance (king moves)
	for dx := -weaponRange; dx <= weaponRange; dx++ {
		for dy := -weaponRange; dy <= weaponRange; dy++ {
			// Skip center (player position)
			if dx == 0 && dy == 0 {
				continue
			}

			tx := centerX + dx
			ty := centerY + dy

			// Check bounds
			if tx < 0 || tx >= maxX || ty < 0 || ty >= maxY {
				continue
			}

			result = append(result, Position{X: tx, Y: ty})
		}
	}

	return result
}

// getEquippedWeaponRange returns the attack range based on equipped weapon.
// Returns 1 for melee weapons (adjacent tiles) and 3+ for ranged weapons.
func (g *Game) getEquippedWeaponRange() int {
	g.mu.RLock()
	player := g.player
	inventoryItems := g.inventoryItems
	g.mu.RUnlock()

	if player == nil {
		return 1 // Default to melee
	}

	// Check equipped weapon from inventory
	for _, item := range inventoryItems {
		if item.Equipped && isWeapon(item.Type) {
			return getWeaponRangeFromType(item.Type, item.Name)
		}
	}

	// Check equipment slots on player (fallback)
	for _, equip := range player.Equipment {
		if equip.Slot == "weapon" || equip.Slot == "main_hand" {
			// Heuristic based on weapon name
			return getWeaponRangeFromName(equip.Name)
		}
	}

	return 1 // Default to melee
}

// isWeapon returns true if the item type is a weapon.
func isWeapon(itemType string) bool {
	switch itemType {
	case "weapon", "melee", "ranged", "sword", "axe", "bow", "crossbow", "dagger", "mace", "spear", "staff", "wand":
		return true
	}
	return false
}

// getWeaponRangeFromType returns weapon range based on item type.
func getWeaponRangeFromType(itemType, name string) int {
	// Ranged weapons have greater range
	switch itemType {
	case "ranged", "bow", "crossbow":
		return 5 // Standard ranged weapon range
	case "wand", "staff":
		return 3 // Magic implements have medium range
	}

	// Check name for ranged indicators
	return getWeaponRangeFromName(name)
}

// getWeaponRangeFromName returns weapon range based on weapon name heuristics.
func getWeaponRangeFromName(name string) int {
	// Check for ranged weapon keywords
	lowerName := strings.ToLower(name)
	if strings.Contains(lowerName, "bow") ||
		strings.Contains(lowerName, "crossbow") ||
		strings.Contains(lowerName, "longbow") ||
		strings.Contains(lowerName, "shortbow") {
		return 6 // Bows have longest range
	}
	if strings.Contains(lowerName, "javelin") ||
		strings.Contains(lowerName, "throwing") ||
		strings.Contains(lowerName, "sling") {
		return 4 // Thrown weapons have medium range
	}
	if strings.Contains(lowerName, "wand") ||
		strings.Contains(lowerName, "staff") {
		return 3 // Magic implements
	}
	if strings.Contains(lowerName, "spear") ||
		strings.Contains(lowerName, "polearm") ||
		strings.Contains(lowerName, "halberd") {
		return 2 // Reach weapons
	}

	return 1 // Default melee (adjacent 8 tiles)
}

// cycleTarget cycles through available targets in the initiative list.
func (g *Game) cycleTarget(delta int) {
	g.mu.RLock()
	combat := g.combat
	currentIdx := g.targetIndex
	g.mu.RUnlock()
	if combat == nil {
		return
	}

	// Build list of enemy targets with their IDs
	type targetInfo struct {
		ID   string
		Name string
	}
	enemies := make([]targetInfo, 0)
	for _, entry := range combat.Initiative {
		if !entry.IsPlayer {
			enemies = append(enemies, targetInfo{ID: entry.ID, Name: entry.Name})
		}
	}
	if len(enemies) == 0 {
		return
	}

	// Cycle through targets
	newIdx := currentIdx + delta
	if newIdx < 0 {
		newIdx = len(enemies) - 1
	} else if newIdx >= len(enemies) {
		newIdx = 0
	}

	target := enemies[newIdx]

	// Update target state
	g.mu.Lock()
	g.targetIndex = newIdx
	g.targetID = target.ID
	g.mu.Unlock()

	// Fetch combat modifiers for the new target asynchronously
	go g.fetchCombatModifiers(target.ID)

	g.addLogMessage(fmt.Sprintf("Target: %s", target.Name), MessageCombat)
}

// fetchCombatModifiers retrieves cover/flanking info for the given target.
func (g *Game) fetchCombatModifiers(targetID string) {
	if g.rpcClient == nil {
		return
	}

	modifiers, err := g.rpcClient.GetCombatModifiers(targetID)
	if err != nil {
		// Silently ignore errors - modifiers are optional enhancement
		return
	}

	g.mu.Lock()
	// Only update if this is still the current target
	if g.targetID == targetID {
		g.targetModifiers = modifiers
	}
	g.mu.Unlock()
}

// getMoraleIndicator returns the icon and color for an NPC morale state.
// Steadfast: no icon, Shaken: yellow "!", Broken: red "!!", Panicked: flee icon
func getMoraleIndicator(moraleState string) (string, color.RGBA) {
	switch moraleState {
	case "Steadfast", "steadfast":
		// No icon for steadfast (normal fighting state)
		return "", ColorGold
	case "Shaken", "shaken":
		// Yellow warning icon
		return "!", ColorEffectControl
	case "Broken", "broken":
		// Red double warning
		return "!!", ColorEnemyName
	case "Panicked", "panicked":
		// Flee/skull icon (using text symbol)
		return "X!", color.RGBA{R: 255, G: 50, B: 50, A: 255}
	default:
		return "", ColorStatLabel
	}
}

// announceTurnTransition checks for round/turn changes and announces them.
// Item 17: Turn Transition Announcement - announces round starts and turn changes.
func (g *Game) announceTurnTransition(combat *CombatState) {
	if combat == nil || !combat.Active {
		return
	}

	g.mu.Lock()
	lastRound := g.lastAnnouncedRound
	lastTurn := g.lastAnnouncedTurn
	g.mu.Unlock()

	// Check for round change
	if combat.Round > lastRound && combat.Round > 0 {
		g.addLogMessage(fmt.Sprintf("--- ROUND %d BEGINS ---", combat.Round), MessageSystem)
		g.mu.Lock()
		g.lastAnnouncedRound = combat.Round
		g.mu.Unlock()
	}

	// Check for turn change
	if combat.CurrentTurn != "" && combat.CurrentTurn != lastTurn {
		// Find the current combatant's name
		turnName := combat.CurrentTurn
		isPlayer := false
		for _, entry := range combat.Initiative {
			if entry.ID == combat.CurrentTurn {
				turnName = entry.Name
				isPlayer = entry.IsPlayer
				break
			}
		}

		if isPlayer {
			g.addLogMessage("-- YOUR TURN --", MessageSystem)
			// Trigger visual flash for player turn
			g.mu.Lock()
			g.turnChangeFlash = time.Now()
			g.mu.Unlock()
		} else {
			g.addLogMessage(fmt.Sprintf("%s's turn", turnName), MessageCombat)
		}

		g.mu.Lock()
		g.lastAnnouncedTurn = combat.CurrentTurn
		g.mu.Unlock()
	}
}

// --- Spell Targeting System ---

// isSpellTargetMode returns true if spell targeting is active.
func (g *Game) isSpellTargetMode() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.spellTargetMode && g.pendingSpell != nil
}

// handleSpellTargeting processes input during spell target selection.
func (g *Game) handleSpellTargeting() {
	// Escape → cancel spell targeting
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.cancelSpellTargeting()
		return
	}

	// Enter → confirm target
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.confirmSpellTarget()
		return
	}

	// Tab → cycle through targets (for single-target spells)
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.mu.RLock()
		spell := g.pendingSpell
		g.mu.RUnlock()
		if spell != nil && spell.TargetType == "single" {
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				g.cycleSpellTarget(-1)
			} else {
				g.cycleSpellTarget(1)
			}
		}
		return
	}

	// Arrow keys → move area target cursor (for area spells)
	g.handleSpellTargetMovement()

	// Mouse click → select target
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.handleSpellTargetClick()
	}
}

// cycleSpellTarget cycles through valid targets for single-target spells.
func (g *Game) cycleSpellTarget(delta int) {
	g.mu.RLock()
	combat := g.combat
	currentID := g.spellTargetID
	g.mu.RUnlock()

	if combat == nil {
		return
	}

	// Build list of valid targets with calculated positions
	type targetInfo struct {
		entry InitiativeEntry
		pos   Position
	}
	var validTargets []targetInfo
	enemyIdx := 0
	for _, entry := range combat.Initiative {
		if !entry.IsPlayer {
			// Calculate position using same logic as getOccupiedPositions
			ex := 100 + enemyIdx*tileSize*2
			ey := 50
			pos := Position{X: ex / tileSize, Y: ey / tileSize}
			validTargets = append(validTargets, targetInfo{entry: entry, pos: pos})
			enemyIdx++
		}
	}

	if len(validTargets) == 0 {
		return
	}

	// Find current index
	currentIdx := -1
	for i, t := range validTargets {
		if t.entry.ID == currentID {
			currentIdx = i
			break
		}
	}

	// Calculate new index
	newIdx := currentIdx + delta
	if newIdx < 0 {
		newIdx = len(validTargets) - 1
	} else if newIdx >= len(validTargets) {
		newIdx = 0
	}

	target := validTargets[newIdx]
	g.mu.Lock()
	g.spellTargetID = target.entry.ID
	g.spellTargetPos = target.pos
	g.mu.Unlock()
}

// handleSpellTargetMovement handles arrow key movement for area spell targeting.
func (g *Game) handleSpellTargetMovement() {
	g.mu.RLock()
	spell := g.pendingSpell
	pos := g.spellTargetPos
	g.mu.RUnlock()

	if spell == nil || spell.TargetType == "single" {
		return
	}

	dx, dy := 0, 0
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		dy = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		dy = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		dx = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		dx = 1
	}

	if dx != 0 || dy != 0 {
		newX := pos.X + dx
		newY := pos.Y + dy
		// Clamp to grid bounds
		if newX >= 0 && newX < 10 && newY >= 0 && newY < 10 {
			g.mu.Lock()
			g.spellTargetPos = Position{X: newX, Y: newY}
			g.mu.Unlock()
		}
	}
}

// handleSpellTargetClick handles mouse click for target selection.
func (g *Game) handleSpellTargetClick() {
	mx, my := ebiten.CursorPosition()
	gridX := mx / tileSize
	gridY := my / tileSize

	if gridX < 0 || gridX >= 10 || gridY < 0 || gridY >= 10 {
		return
	}

	g.mu.RLock()
	spell := g.pendingSpell
	combat := g.combat
	g.mu.RUnlock()

	if spell == nil {
		return
	}

	// For single-target, check if clicked on a valid target
	if spell.TargetType == "single" && combat != nil {
		// Calculate enemy positions using same logic as getOccupiedPositions
		enemyIdx := 0
		for _, entry := range combat.Initiative {
			if !entry.IsPlayer {
				ex := 100 + enemyIdx*tileSize*2
				ey := 50
				tileX := ex / tileSize
				tileY := ey / tileSize
				if tileX == gridX && tileY == gridY {
					g.mu.Lock()
					g.spellTargetID = entry.ID
					g.spellTargetPos = Position{X: gridX, Y: gridY}
					g.mu.Unlock()
					g.confirmSpellTarget()
					return
				}
				enemyIdx++
			}
		}
	} else {
		// For area spells, set position and confirm
		g.mu.Lock()
		g.spellTargetPos = Position{X: gridX, Y: gridY}
		g.mu.Unlock()
		g.confirmSpellTarget()
	}
}

// getCombatInitiative returns the current combat initiative list.
func (g *Game) getCombatInitiative() []InitiativeEntry {
	if g.combat != nil {
		return g.combat.Initiative
	}
	return nil
}

// drawSpellTargetHighlights draws the spell targeting overlay on the combat grid.
func (g *Game) drawSpellTargetHighlights(screen *ebiten.Image, gridW, gridH int) {
	g.mu.RLock()
	spell := g.pendingSpell
	targetPos := g.spellTargetPos
	combat := g.combat
	spellTargetID := g.spellTargetID
	g.mu.RUnlock()

	if spell == nil {
		return
	}

	switch spell.TargetType {
	case "single":
		g.drawSingleTargetHighlights(screen, combat, spellTargetID, gridW, gridH)
	case "area":
		g.drawAreaTargetHighlights(screen, spell, targetPos, gridW, gridH)
	case "cone":
		g.drawConeTargetHighlight(screen, targetPos, gridW, gridH)
	}

	g.drawSpellTargetPanel(screen, spell, gridW)
}

// drawSingleTargetHighlights draws highlights for single-target spells on enemies.
func (g *Game) drawSingleTargetHighlights(screen *ebiten.Image, combat *CombatState, selectedID string, gridW, gridH int) {
	if combat == nil {
		return
	}
	rangeColor := color.RGBA{R: 60, G: 100, B: 200, A: 80}
	enemyIdx := 0
	for _, entry := range combat.Initiative {
		if entry.IsPlayer {
			continue
		}
		ex := 100 + enemyIdx*tileSize*2
		ey := 50
		tx := (ex / tileSize) * tileSize
		ty := (ey / tileSize) * tileSize
		if tx >= 0 && tx < gridW-tileSize && ty >= 0 && ty < gridH-tileSize {
			drawRect(screen, tx, ty, tileSize, tileSize, rangeColor)
			if entry.ID == selectedID {
				g.drawPulsingBorder(screen, tx, ty, tileSize, tileSize)
			}
		}
		enemyIdx++
	}
}

// drawAreaTargetHighlights draws highlights for area-of-effect spells.
func (g *Game) drawAreaTargetHighlights(screen *ebiten.Image, spell *SpellData, targetPos Position, gridW, gridH int) {
	rangeColor := color.RGBA{R: 60, G: 100, B: 200, A: 80}
	targetColor := color.RGBA{R: 100, G: 200, B: 255, A: 120}
	radius := spell.AreaRadius
	if radius < 1 {
		radius = 1
	}
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			ax, ay := targetPos.X+dx, targetPos.Y+dy
			if ax >= 0 && ax < 10 && ay >= 0 && ay < 10 {
				tx, ty := ax*tileSize, ay*tileSize
				if tx >= 0 && tx < gridW-tileSize && ty >= 0 && ty < gridH-tileSize {
					drawRect(screen, tx, ty, tileSize, tileSize, rangeColor)
				}
			}
		}
	}
	cx, cy := targetPos.X*tileSize, targetPos.Y*tileSize
	if cx >= 0 && cx < gridW-tileSize && cy >= 0 && cy < gridH-tileSize {
		drawRect(screen, cx, cy, tileSize, tileSize, targetColor)
		g.drawPulsingBorder(screen, cx, cy, tileSize, tileSize)
	}
}

// drawConeTargetHighlight draws the highlight for cone-type spell targeting.
func (g *Game) drawConeTargetHighlight(screen *ebiten.Image, targetPos Position, gridW, gridH int) {
	targetColor := color.RGBA{R: 100, G: 200, B: 255, A: 120}
	cx, cy := targetPos.X*tileSize, targetPos.Y*tileSize
	if cx >= 0 && cx < gridW-tileSize && cy >= 0 && cy < gridH-tileSize {
		drawRect(screen, cx, cy, tileSize, tileSize, targetColor)
		g.drawPulsingBorder(screen, cx, cy, tileSize, tileSize)
	}
}

// drawSpellTargetPanel draws the spell targeting info panel.
func (g *Game) drawSpellTargetPanel(screen *ebiten.Image, spell *SpellData, gridW int) {
	panelX := gridW - 180
	panelY := 5
	panelW := 170
	panelH := 60

	// Background
	drawRect(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 30, G: 30, B: 60, A: 220})
	drawRectOutline(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 100, G: 150, B: 255, A: 255})

	// Spell name
	drawColoredText(screen, spell.Name, panelX+5, panelY+8, ColorGold)

	// Target type hint
	var hint string
	switch spell.TargetType {
	case "single":
		hint = "Tab: Cycle | Click: Select"
	case "area":
		hint = "Arrows: Move | Click: Cast"
	case "cone":
		hint = "Click: Target Direction"
	}
	drawColoredText(screen, hint, panelX+5, panelY+26, ColorStatLabel)

	// Controls
	drawColoredText(screen, "Enter: Cast | Esc: Cancel", panelX+5, panelY+44, ColorStatLabel)
}
