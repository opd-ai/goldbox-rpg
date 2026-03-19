//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

	g.mu.RLock()
	player := g.player
	combat := g.combat
	g.mu.RUnlock()

	// Draw player token
	if player != nil {
		px := (gridWidth / 2) - (tileSize / 2)
		py := (gridHeight / 2) - (tileSize / 2)

		// Use player sprite based on class
		spritePath := g.getPlayerSpritePath(player)
		DrawSpriteWithFallback(screen, spritePath, px, py, tileSize-2, tileSize-2,
			color.RGBA{R: 80, G: 200, B: 80, A: 255})

		// Show "P" indicator while sprite loads
		initSpriteCache()
		if !spriteCache.IsCached(spritePath) {
			ebitenutil.DebugPrintAt(screen, "P", px+10, py+8)
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
					// Use monster sprite with fallback to red rect
					monsterPath := MonsterSpritePath(entry.Name)
					DrawSpriteWithFallback(screen, monsterPath, ex, ey, tileSize-2, tileSize-2,
						color.RGBA{R: 200, G: 80, B: 80, A: 255})

					// Show "E" indicator while sprite loads
					initSpriteCache()
					if !spriteCache.IsCached(monsterPath) {
						ebitenutil.DebugPrintAt(screen, "E", ex+10, ey+8)
					}
				}
				enemyIdx++
			}
		}
	}

	// Combat round indicator
	if combat != nil {
		drawColoredText(screen, fmt.Sprintf("Round %d", combat.Round), 10, 5, ColorGold)
	}
}

// drawInitiativePanel renders the initiative tracker on the right side (§5.1).
func (g *Game) drawInitiativePanel(screen *ebiten.Image) {
	panelX := g.screenWidth - charPanelWidth
	panelHeight := g.screenHeight - actionPanelHeight

	// Panel background — deep dark
	drawRect(screen, panelX, 0, charPanelWidth, panelHeight, color.RGBA{R: 30, G: 28, B: 42, A: 255})

	// Double-border: outer bright, inner dim
	drawRectOutline(screen, panelX, 0, charPanelWidth, panelHeight, ColorPanelBorder)
	drawRectOutline(screen, panelX+2, 2, charPanelWidth-4, panelHeight-4, color.RGBA{R: 50, G: 45, B: 70, A: 255})

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
	g.mu.RUnlock()

	// Turn indicator per §5 — "YOUR TURN" / "Waiting..." with proper color
	turnLabel := "Waiting..."
	turnColor := color.RGBA{R: 180, G: 140, B: 60, A: 255}
	if combat != nil && combat.IsPlayerTurn {
		turnLabel = "YOUR TURN"
		turnColor = color.RGBA{R: 80, G: 220, B: 80, A: 255}
	}
	drawColoredText(screen, turnLabel, panelWidth-120, panelY+5, turnColor)

	// Action buttons per §5 Action Panel: Move / Attack / Cast / UseItem / EndTurn
	actions := []struct {
		label  string
		action CombatAction
		key    string
	}{
		{"Move", CombatActionMove, "M"},
		{"Attack", CombatActionAttack, "A"},
		{"Cast", CombatActionCast, "C"},
		{"UseItem", CombatActionItem, "U"},
	}

	btnWidth := 100
	btnHeight := 35
	startX := 20
	for i, a := range actions {
		x := startX + i*(btnWidth+10)
		y := panelY + 20

		btnColor := color.RGBA{R: 45, G: 40, B: 65, A: 255}
		if currentAction == a.action {
			btnColor = color.RGBA{R: 100, G: 80, B: 60, A: 255}
		}
		if g.hoveredButton == "combat_"+a.label {
			btnColor = color.RGBA{R: 65, G: 58, B: 95, A: 255}
		}

		drawRect(screen, x, y, btnWidth, btnHeight, btnColor)
		drawRectOutline(screen, x, y, btnWidth, btnHeight, ColorPanelBorder)

		// Highlight the hotkey letter in gold
		btnText := fmt.Sprintf("[%s] %s", a.key, a.label)
		textColor := ColorStatValue
		if currentAction == a.action {
			textColor = ColorGoldHi
		}
		drawColoredText(screen, btnText, x+5, y+10, textColor)
	}

	// End Turn button
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
		g.addLogMessage("Attack mode - select target (Tab to cycle)", MessageCombat)
		// Attempt to attack the first available enemy target
		g.mu.RLock()
		combat := g.combat
		player := g.player
		g.mu.RUnlock()
		if combat != nil && player != nil {
			// Find first enemy target
			for _, entry := range combat.Initiative {
				if !entry.IsPlayer {
					go g.executeAttack(player.Name, entry.ID, entry.Name)
					break
				}
			}
		}
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
			// Show remaining target HP if available
			if result.TargetHealth >= 0 {
				g.addLogMessage(fmt.Sprintf("  %s: %d HP remaining", targetName, result.TargetHealth), MessageInfo)
			}
		} else {
			g.addLogMessage(fmt.Sprintf("%s attacks %s -- HIT!", attackerName, targetName), MessageCombat)
		}
	} else {
		// Miss narration
		g.addLogMessage(fmt.Sprintf("%s attacks %s -- MISS", attackerName, targetName), MessageCombat)
	}

	if result.Message != "" {
		g.addLogMessage(fmt.Sprintf("  %s", result.Message), MessageInfo)
	}
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
