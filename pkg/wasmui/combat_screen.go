//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// updateCombat handles input during combat mode (§5).
func (g *Game) updateCombat() {
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

// handleCombatHotkeys processes combat action hotkeys (M, A, C, U).
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

// handleCombatTouchTap processes touch taps on combat action bar buttons.
func (g *Game) handleCombatTouchTap() {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return
	}

	panelY := g.screenHeight - actionPanelHeight
	btnWidth := 100
	btnHeight := 35
	startX := 20

	// Action buttons: Move, Attack, Cast, UseItem
	combatActions := []CombatAction{CombatActionMove, CombatActionAttack, CombatActionCast, CombatActionItem}
	for i, ca := range combatActions {
		x := startX + i*(btnWidth+10)
		y := panelY + 20
		if tx >= x && tx <= x+btnWidth && ty >= y && ty <= y+btnHeight {
			g.mu.Lock()
			g.combatAction = ca
			g.mu.Unlock()
			g.executeCombatAction(ca)
			return
		}
	}

	// End Turn button
	endX := startX + 4*(btnWidth+10) + 20
	endY := panelY + 20
	endW := btnWidth + 10
	if tx >= endX && tx <= endX+endW && ty >= endY && ty <= endY+btnHeight {
		g.handleEndTurn()
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

	// Draw movement range highlighting when in move mode
	g.mu.RLock()
	combatAction := g.combatAction
	playerAP := 0
	if g.player != nil {
		playerAP = g.player.AP
	}
	g.mu.RUnlock()

	if combatAction == CombatActionMove && playerAP > 0 {
		// Calculate center of grid (where player is)
		centerTileX := gridWidth / (2 * tileSize)
		centerTileY := gridHeight / (2 * tileSize)

		// Get positions occupied by enemies (to exclude from movement range)
		occupiedPositions := g.getOccupiedPositions(tileSize)

		moveRange := g.getMovementRange(centerTileX, centerTileY, playerAP, gridWidth/tileSize, gridHeight/tileSize, occupiedPositions)

		// Draw blue tint on reachable tiles
		moveHighlightColor := color.RGBA{R: 74, G: 125, B: 191, A: 80}
		for _, pos := range moveRange {
			tx := pos.X * tileSize
			ty := pos.Y * tileSize
			if tx >= 0 && tx < gridWidth-tileSize && ty >= 0 && ty < gridHeight-tileSize {
				drawRect(screen, tx, ty, tileSize, tileSize, moveHighlightColor)
			}
		}
	}

	// Draw attack range highlighting when in attack mode
	if combatAction == CombatActionAttack {
		centerTileX := gridWidth / (2 * tileSize)
		centerTileY := gridHeight / (2 * tileSize)

		// Get weapon range based on equipped weapon
		weaponRange := g.getEquippedWeaponRange()

		// Get enemy positions to highlight valid targets
		enemyPositions := g.getOccupiedPositions(tileSize)

		attackRange := g.getAttackRange(centerTileX, centerTileY, weaponRange, gridWidth/tileSize, gridHeight/tileSize)

		// Draw red tint on attackable tiles
		attackHighlightColor := color.RGBA{R: 191, G: 74, B: 74, A: 80}
		for _, pos := range attackRange {
			tx := pos.X * tileSize
			ty := pos.Y * tileSize
			if tx >= 0 && tx < gridWidth-tileSize && ty >= 0 && ty < gridHeight-tileSize {
				drawRect(screen, tx, ty, tileSize, tileSize, attackHighlightColor)

				// If this tile has an enemy, add pulsing gold border to indicate valid target
				if enemyPositions[pos] {
					g.drawPulsingBorder(screen, tx, ty, tileSize, tileSize)
				}
			}
		}
	}

	// Clean up expired flash effects
	g.cleanupExpiredFlashes()
	g.cleanupExpiredSpellEffects()

	g.mu.RLock()
	player := g.player
	combat := g.combat
	g.mu.RUnlock()

	// Determine if it's the player's turn for active tile highlight
	isPlayerTurn := combat != nil && combat.IsPlayerTurn

	// Draw player token
	if player != nil {
		px := (gridWidth / 2) - (tileSize / 2)
		py := (gridHeight / 2) - (tileSize / 2)

		// Draw pulsing gold border if it's player's turn
		if isPlayerTurn {
			g.drawPulsingBorder(screen, px-2, py-2, tileSize+2, tileSize+2)
		}

		// Use player sprite based on class
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
	}

	// Draw enemy indicators from initiative
	if combat != nil {
		enemyIdx := 0
		for _, entry := range combat.Initiative {
			if !entry.IsPlayer {
				ex := 100 + enemyIdx*tileSize*2
				ey := 50
				if ex < gridWidth-tileSize {
					// Draw pulsing gold border if it's this enemy's turn
					isEnemyTurn := entry.ID == combat.CurrentTurn
					if isEnemyTurn {
						g.drawPulsingBorder(screen, ex-2, ey-2, tileSize+2, tileSize+2)
					}

					// Use monster sprite with fallback to red rect
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
				}
				enemyIdx++
			}
		}
	}

	// Draw spell effects overlay
	g.drawSpellEffects(screen, 0, 0, tileSize)

	// Combat round indicator
	if combat != nil {
		drawColoredText(screen, fmt.Sprintf("Round %d", combat.Round), 10, 5, ColorGold)
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

// drawInitiativePanel renders the initiative tracker on the right side (§5.1).
func (g *Game) drawInitiativePanel(screen *ebiten.Image) {
	panelX := g.screenWidth - charPanelWidth
	panelHeight := g.screenHeight - actionPanelHeight

	// Panel background — deep dark
	drawRect(screen, panelX, 0, charPanelWidth, panelHeight, color.RGBA{R: 30, G: 28, B: 42, A: 255})

	// Bold Gold Box-style triple border
	drawBoldPanelBorder(screen, panelX, 0, charPanelWidth, panelHeight)

	// Title in gold
	drawColoredText(screen, "INITIATIVE", panelX+50, 10, ColorGold)

	g.mu.RLock()
	combat := g.combat
	player := g.player
	g.mu.RUnlock()

	if combat == nil {
		drawColoredText(screen, "No combat", panelX+50, 40, ColorStatLabel)
		return
	}

	// Current turn indicator
	drawColoredText(screen, fmt.Sprintf("Round: %d", combat.Round), panelX+10, 35, ColorStatValue)
	if combat.CurrentTurn != "" {
		drawColoredText(screen, fmt.Sprintf("Turn: %s", truncateText(combat.CurrentTurn, 15)), panelX+10, 50, ColorStatLabel)
	}

	// Separator line
	drawLine(screen, panelX+10, 67, panelX+charPanelWidth-10, 67, ColorPanelBorder)

	// Initiative list
	y := 75
	for i, entry := range combat.Initiative {
		if i >= 10 {
			break
		}
		marker := "  "
		if entry.ID == combat.CurrentTurn {
			marker = "> "
		}

		// Color by allegiance: green for player, red for enemy
		// Current turn gets a brighter highlight
		nameColor := ColorEnemyName
		if entry.IsPlayer {
			nameColor = ColorPlayerName
		}
		if entry.ID == combat.CurrentTurn {
			// Brighten the current turn name
			nameColor = brightenColor(nameColor, 60)
			// Draw a subtle highlight bar behind current turn
			drawRect(screen, panelX+5, y-1, charPanelWidth-10, 18,
				color.RGBA{R: 40, G: 38, B: 60, A: 255})
		}

		label := fmt.Sprintf("%s%s", marker, truncateText(entry.Name, 12))
		drawColoredText(screen, label, panelX+10, y, nameColor)

		// HP bar for each combatant
		if entry.MaxHP > 0 {
			barX := panelX + 120
			barW := 60
			pct := float64(entry.HP) / float64(entry.MaxHP)
			drawRect(screen, barX, y, barW, 10, color.RGBA{R: 60, G: 20, B: 20, A: 255})
			drawRect(screen, barX, y, int(pct*float64(barW)), 10, hpBarColor(pct))
		}

		y += 20
	}

	// Player stats summary at bottom
	if player != nil {
		y = panelHeight - 100
		drawColoredText(screen, player.Name, panelX+10, y, ColorPlayerName)
		g.drawHPBar(screen, panelX, y-65, player)
		g.drawAPBar(screen, panelX, y+20, player)
	}
}

// drawCombatActionBar renders the bottom action bar for combat (§5.2).
func (g *Game) drawCombatActionBar(screen *ebiten.Image) {
	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth

	drawRect(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 25, G: 23, B: 38, A: 255})
	drawRectOutline(screen, 0, panelY, panelWidth, actionPanelHeight, ColorPanelBorder)

	g.mu.RLock()
	currentAction := g.combatAction
	combat := g.combat
	player := g.player
	g.mu.RUnlock()

	// Get current AP for affordability check
	currentAP := 0
	if player != nil {
		currentAP = player.AP
	}

	// Turn indicator per §5 — "YOUR TURN" / "Waiting..." with proper color
	turnLabel := "Waiting..."
	turnColor := color.RGBA{R: 180, G: 140, B: 60, A: 255}
	if combat != nil && combat.IsPlayerTurn {
		turnLabel = "YOUR TURN"
		turnColor = color.RGBA{R: 80, G: 220, B: 80, A: 255}
	}
	drawColoredText(screen, turnLabel, panelWidth-120, panelY+5, turnColor)

	// Action buttons per §5 Action Panel: Move / Attack / Cast / UseItem / EndTurn
	// Each action has an AP cost
	actions := []struct {
		label  string
		action CombatAction
		key    string
		cost   int    // AP cost (0 means variable)
		costTx string // cost text to display
	}{
		{"Move", CombatActionMove, "M", 1, "1"},
		{"Attack", CombatActionAttack, "A", 1, "1"},
		{"Cast", CombatActionCast, "C", 0, "1-3"}, // Variable cost for spells
		{"UseItem", CombatActionItem, "U", 1, "1"},
	}

	btnWidth := 100
	btnHeight := 35
	startX := 20
	for i, a := range actions {
		x := startX + i*(btnWidth+10)
		y := panelY + 20

		// Check if action is affordable
		canAfford := a.cost == 0 || currentAP >= a.cost

		btnColor := color.RGBA{R: 45, G: 40, B: 65, A: 255}
		if !canAfford {
			// Dim unaffordable actions
			btnColor = color.RGBA{R: 30, G: 28, B: 42, A: 255}
		} else if currentAction == a.action {
			btnColor = color.RGBA{R: 100, G: 80, B: 60, A: 255}
		}
		if g.hoveredButton == "combat_"+a.label && canAfford {
			btnColor = color.RGBA{R: 65, G: 58, B: 95, A: 255}
		}

		drawRect(screen, x, y, btnWidth, btnHeight, btnColor)
		drawRectOutline(screen, x, y, btnWidth, btnHeight, ColorPanelBorder)

		// Button text with AP cost in parentheses: "[M] Move (1)"
		btnText := fmt.Sprintf("[%s] %s (%s)", a.key, a.label, a.costTx)
		textColor := ColorStatValue
		if !canAfford {
			// Gray out text for unaffordable actions
			textColor = color.RGBA{R: 80, G: 80, B: 100, A: 255}
		} else if currentAction == a.action {
			textColor = ColorGoldHi
		}
		drawColoredText(screen, btnText, x+3, y+10, textColor)
	}

	// End Turn button (no AP cost)
	endX := startX + 4*(btnWidth+10) + 20
	endY := panelY + 20
	endColor := color.RGBA{R: 65, G: 45, B: 45, A: 255}
	drawRect(screen, endX, endY, btnWidth+10, btnHeight, endColor)
	drawRectOutline(screen, endX, endY, btnWidth+10, btnHeight, color.RGBA{R: 130, G: 90, B: 90, A: 255})
	drawColoredText(screen, "[Space] End Turn", endX+5, endY+10, ColorStatValue)

	// Status line showing current action
	drawColoredText(screen, fmt.Sprintf("Action: %s", currentAction), 20, panelY+60, ColorStatLabel)
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
		// Even on error, provide narration
		g.addLogMessage(fmt.Sprintf("%s attacks %s...", attackerName, targetName), MessageCombat)
		g.addLogMessage(fmt.Sprintf("  Attack failed: %v", err), MessageError)
		return
	}

	if result.Success {
		// Rich Gold Box narration: "Fighter attacks Goblin — HIT for 7 damage!"
		if result.Damage > 0 {
			g.addLogMessage(fmt.Sprintf("%s attacks %s -- HIT for %d damage!", attackerName, targetName, result.Damage), MessageCombat)
			// Add damage flash effect on the target (red flash)
			g.addDamageFlash(targetID, ColorEnemyName)
			// Show remaining target HP if available
			if result.TargetHealth >= 0 {
				g.addLogMessage(fmt.Sprintf("  %s: %d HP remaining", targetName, result.TargetHealth), MessageInfo)
			}
		} else {
			g.addLogMessage(fmt.Sprintf("%s attacks %s -- HIT!", attackerName, targetName), MessageCombat)
			// Still add flash for hit without damage (e.g., resistance)
			g.addDamageFlash(targetID, ColorEnemyName)
		}
	} else {
		// Miss narration
		g.addLogMessage(fmt.Sprintf("%s attacks %s -- MISS", attackerName, targetName), MessageCombat)
	}

	if result.Message != "" {
		g.addLogMessage(fmt.Sprintf("  %s", result.Message), MessageInfo)
	}
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

		// Calculate screen position
		screenX := gridX + effect.TargetPos.X*tileSize + tileSize/2
		screenY := gridY + effect.TargetPos.Y*tileSize + tileSize/2

		// Get effect color based on spell school
		effectColor := SpellSchoolColor(effect.SpellSchool)

		// Draw expanding circle effect as fallback
		radius := effect.GetRadius()
		alpha := effect.GetAlpha()

		// Adjust color alpha
		c := color.RGBA{
			R: effectColor.R,
			G: effectColor.G,
			B: effectColor.B,
			A: uint8(float32(effectColor.A) * alpha),
		}

		// Draw effect as multiple concentric circles
		g.drawSpellCircle(screen, screenX, screenY, radius, c)

		// Draw inner brighter core
		coreColor := color.RGBA{
			R: uint8(min(int(effectColor.R)+80, 255)),
			G: uint8(min(int(effectColor.G)+80, 255)),
			B: uint8(min(int(effectColor.B)+80, 255)),
			A: uint8(float32(200) * alpha),
		}
		g.drawSpellCircle(screen, screenX, screenY, radius*0.5, coreColor)
	}
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

// cycleTarget cycles through available targets in the initiative list.
func (g *Game) cycleTarget(delta int) {
	g.mu.RLock()
	combat := g.combat
	g.mu.RUnlock()
	if combat == nil {
		return
	}

	// Count enemies
	enemies := make([]string, 0)
	for _, entry := range combat.Initiative {
		if !entry.IsPlayer {
			enemies = append(enemies, entry.Name)
		}
	}
	if len(enemies) == 0 {
		return
	}

	g.addLogMessage(fmt.Sprintf("Target: %s", enemies[0]), MessageCombat)
}
