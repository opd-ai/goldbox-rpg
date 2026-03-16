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
	// Escape → back to exploration
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.mode = ModeNormal
		g.screenState = ScreenExploration
		g.mu.Unlock()
		return
	}

	// Tab → cycle through targets
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.cycleTarget(1)
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.cycleTarget(-1)
		return
	}

	// Number keys → quick-select action
	numberActions := map[ebiten.Key]CombatAction{
		ebiten.KeyDigit1: CombatActionAttack,
		ebiten.KeyDigit2: CombatActionCast,
		ebiten.KeyDigit3: CombatActionItem,
		ebiten.KeyDigit4: CombatActionDefend,
		ebiten.KeyDigit5: CombatActionFlee,
	}
	for key, action := range numberActions {
		if inpututil.IsKeyJustPressed(key) {
			g.mu.Lock()
			g.combatAction = action
			g.mu.Unlock()
			g.executeCombatAction(action)
			return
		}
	}

	// Space → end turn
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.handleEndTurn()
		return
	}

	// Enter → confirm current action
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.mu.RLock()
		action := g.combatAction
		g.mu.RUnlock()
		if action != CombatActionNone {
			g.executeCombatAction(action)
		}
		return
	}

	// Movement in combat grid (§5.3)
	directions := map[ebiten.Key]string{
		ebiten.KeyW: "north", ebiten.KeyS: "south",
		ebiten.KeyA: "west", ebiten.KeyD: "east",
	}
	for key, dir := range directions {
		if inpututil.IsKeyJustPressed(key) {
			g.handleMove(dir)
			return
		}
	}

	// Mouse input
	g.handleMouseInput()
}

// drawCombatScreen renders the combat interface (§5).
func (g *Game) drawCombatScreen(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 25, G: 20, B: 30, A: 255})

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

	// Draw grid lines
	gridColor := color.RGBA{R: 40, G: 40, B: 55, A: 255}
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
		drawRect(screen, px, py, tileSize-2, tileSize-2, color.RGBA{R: 80, G: 200, B: 80, A: 255})
		ebitenutil.DebugPrintAt(screen, "P", px+10, py+8)
	}

	// Draw enemy indicators from initiative
	if combat != nil {
		enemyIdx := 0
		for _, entry := range combat.Initiative {
			if !entry.IsPlayer {
				ex := 100 + enemyIdx*tileSize*2
				ey := 50
				if ex < gridWidth-tileSize {
					drawRect(screen, ex, ey, tileSize-2, tileSize-2, color.RGBA{R: 200, G: 80, B: 80, A: 255})
					ebitenutil.DebugPrintAt(screen, "E", ex+10, ey+8)
				}
				enemyIdx++
			}
		}
	}

	// Combat round indicator
	if combat != nil {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round %d", combat.Round), 10, 5)
	}
}

// drawInitiativePanel renders the initiative tracker on the right side (§5.1).
func (g *Game) drawInitiativePanel(screen *ebiten.Image) {
	panelX := g.screenWidth - charPanelWidth
	panelHeight := g.screenHeight - actionPanelHeight

	drawRect(screen, panelX, 0, charPanelWidth, panelHeight, color.RGBA{R: 40, G: 35, B: 50, A: 255})
	drawRectOutline(screen, panelX, 0, charPanelWidth, panelHeight, color.RGBA{R: 80, G: 70, B: 100, A: 255})

	ebitenutil.DebugPrintAt(screen, "INITIATIVE", panelX+50, 10)

	g.mu.RLock()
	combat := g.combat
	player := g.player
	g.mu.RUnlock()

	if combat == nil {
		ebitenutil.DebugPrintAt(screen, "No combat", panelX+50, 40)
		return
	}

	// Current turn indicator
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round: %d", combat.Round), panelX+10, 35)
	if combat.CurrentTurn != "" {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Turn: %s", truncateText(combat.CurrentTurn, 15)), panelX+10, 50)
	}

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
		nameColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}
		if entry.IsPlayer {
			nameColor = color.RGBA{R: 100, G: 200, B: 100, A: 255}
		}
		_ = nameColor // would use with text/v2; DebugPrintAt doesn't support color

		label := fmt.Sprintf("%s%s", marker, truncateText(entry.Name, 12))
		ebitenutil.DebugPrintAt(screen, label, panelX+10, y)

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
		ebitenutil.DebugPrintAt(screen, player.Name, panelX+10, y)
		g.drawHPBar(screen, panelX, y-65, player)
		g.drawAPBar(screen, panelX, y+20, player)
	}
}

// drawCombatActionBar renders the bottom action bar for combat (§5.2).
func (g *Game) drawCombatActionBar(screen *ebiten.Image) {
	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth

	drawRect(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 35, G: 30, B: 45, A: 255})
	drawRectOutline(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 70, G: 60, B: 90, A: 255})

	g.mu.RLock()
	currentAction := g.combatAction
	g.mu.RUnlock()

	// Action buttons
	actions := []struct {
		label  string
		action CombatAction
		key    string
	}{
		{"Attack", CombatActionAttack, "1"},
		{"Cast", CombatActionCast, "2"},
		{"Item", CombatActionItem, "3"},
		{"Defend", CombatActionDefend, "4"},
		{"Flee", CombatActionFlee, "5"},
	}

	btnWidth := 100
	btnHeight := 35
	startX := 20
	for i, a := range actions {
		x := startX + i*(btnWidth+10)
		y := panelY + 10

		btnColor := color.RGBA{R: 60, G: 55, B: 80, A: 255}
		if currentAction == a.action {
			btnColor = color.RGBA{R: 100, G: 80, B: 60, A: 255}
		}
		if g.hoveredButton == "combat_"+a.label {
			btnColor = color.RGBA{R: 80, G: 75, B: 110, A: 255}
		}

		drawRect(screen, x, y, btnWidth, btnHeight, btnColor)
		drawRectOutline(screen, x, y, btnWidth, btnHeight, color.RGBA{R: 100, G: 90, B: 130, A: 255})
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[%s] %s", a.key, a.label), x+5, y+10)
	}

	// End Turn button
	endX := startX + 5*(btnWidth+10) + 20
	endY := panelY + 10
	endColor := color.RGBA{R: 80, G: 60, B: 60, A: 255}
	drawRect(screen, endX, endY, btnWidth, btnHeight, endColor)
	drawRectOutline(screen, endX, endY, btnWidth, btnHeight, color.RGBA{R: 130, G: 90, B: 90, A: 255})
	ebitenutil.DebugPrintAt(screen, "[Space] End", endX+5, endY+10)

	// Status line
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Action: %s", currentAction), 20, panelY+55)
}

// executeCombatAction dispatches the selected combat action via RPC.
func (g *Game) executeCombatAction(action CombatAction) {
	switch action {
	case CombatActionAttack:
		g.addLogMessage("Select attack target...", MessageCombat)
	case CombatActionCast:
		g.mu.Lock()
		g.mode = ModeSpellcasting
		g.mu.Unlock()
		go g.loadSpells()
	case CombatActionItem:
		g.mu.Lock()
		g.mode = ModeInventory
		g.mu.Unlock()
		go g.loadInventory()
	case CombatActionDefend:
		g.addLogMessage("Defending this turn", MessageCombat)
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
				g.addLogMessage("Cannot flee!", MessageCombat)
			}
		}()
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
