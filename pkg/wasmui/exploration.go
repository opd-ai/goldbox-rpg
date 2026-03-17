//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// updateExploration handles input for the exploration screen (§3.5).
func (g *Game) updateExploration() {
	// Check input cooldown
	if time.Since(g.lastInputTime) < g.inputCooldown {
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

// handleExplorationMovement processes 8-directional movement keys and touch swipes.
// Returns true if movement was processed.
func (g *Game) handleExplorationMovement() bool {
	directions := map[ebiten.Key]string{
		ebiten.KeyW:          "north",
		ebiten.KeyArrowUp:    "north",
		ebiten.KeyNumpad8:    "north",
		ebiten.KeyArrowDown:  "south",
		ebiten.KeyNumpad2:    "south",
		ebiten.KeyA:          "west",
		ebiten.KeyArrowLeft:  "west",
		ebiten.KeyNumpad4:    "west",
		ebiten.KeyD:          "east",
		ebiten.KeyArrowRight: "east",
		ebiten.KeyNumpad6:    "east",
		ebiten.KeyQ:          "northwest",
		ebiten.KeyNumpad7:    "northwest",
		ebiten.KeyE:          "northeast",
		ebiten.KeyNumpad9:    "northeast",
		ebiten.KeyZ:          "southwest",
		ebiten.KeyNumpad1:    "southwest",
		ebiten.KeyC:          "southeast",
		ebiten.KeyNumpad3:    "southeast",
	}

	// S without shift is south movement
	if inpututil.IsKeyJustPressed(ebiten.KeyS) && !ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.handleMove("south")
		g.lastInputTime = time.Now()
		return true
	}

	for key, direction := range directions {
		if inpututil.IsKeyJustPressed(key) {
			g.handleMove(direction)
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
}

// drawViewport renders the main game view.
func (g *Game) drawViewport(screen *ebiten.Image) {
	viewportWidth := g.screenWidth - charPanelWidth
	viewportHeight := g.screenHeight - logPanelHeight - actionPanelHeight

	// Draw viewport background
	drawRect(screen, 0, 0, viewportWidth, viewportHeight, color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Draw grid for reference
	gridColor := color.RGBA{R: 50, G: 50, B: 60, A: 255}
	for x := 0; x < viewportWidth; x += tileSize {
		drawLine(screen, x, 0, x, viewportHeight, gridColor)
	}
	for y := 0; y < viewportHeight; y += tileSize {
		drawLine(screen, 0, y, viewportWidth, y, gridColor)
	}

	// Draw player if available
	g.mu.RLock()
	player := g.player
	g.mu.RUnlock()

	if player != nil {
		playerX := (viewportWidth / 2) - (tileSize / 2)
		playerY := (viewportHeight / 2) - (tileSize / 2)
		drawRect(screen, playerX, playerY, tileSize-2, tileSize-2, color.RGBA{R: 100, G: 200, B: 100, A: 255})

		// Draw player indicator
		ebitenutil.DebugPrintAt(screen, "P", playerX+10, playerY+8)
	} else {
		// Draw placeholder
		ebitenutil.DebugPrintAt(screen, "Waiting for game state...", viewportWidth/2-80, viewportHeight/2)
	}
}

// drawCharacterPanel renders the character information panel (§9).
func (g *Game) drawCharacterPanel(screen *ebiten.Image) {
	panelX := g.screenWidth - charPanelWidth
	panelY := 0
	panelHeight := g.screenHeight - actionPanelHeight

	// Panel background
	drawRect(screen, panelX, panelY, charPanelWidth, panelHeight, color.RGBA{R: 40, G: 40, B: 50, A: 255})
	drawRectOutline(screen, panelX, panelY, charPanelWidth, panelHeight, color.RGBA{R: 80, G: 80, B: 100, A: 255})

	// Title
	ebitenutil.DebugPrintAt(screen, "CHARACTER", panelX+60, panelY+10)

	g.mu.RLock()
	player := g.player
	combat := g.combat
	g.mu.RUnlock()

	if player != nil {
		g.drawPlayerStats(screen, panelX, panelY, player)
	} else {
		ebitenutil.DebugPrintAt(screen, "No character", panelX+50, panelY+80)
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
	ebitenutil.DebugPrintAt(screen, player.Name, panelX+10, panelY+40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Lv %d %s", player.Level, player.Class), panelX+10, panelY+55)

	// HP bar (§9.1)
	g.drawHPBar(screen, panelX, panelY, player)

	// AP bar (§9.1)
	g.drawAPBar(screen, panelX, panelY+95, player)

	// Attributes
	g.drawAttributes(screen, panelX, panelY, player.Attributes)

	// Position (§9.5)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Pos: (%d, %d)", player.Position.X, player.Position.Y), panelX+10, panelY+185)

	// Active effects (§9.4)
	g.drawActiveEffects(screen, panelX, panelY+200, player.Effects)
}

// drawHPBar renders the HP bar with color coding.
func (g *Game) drawHPBar(screen *ebiten.Image, panelX, panelY int, player *PlayerState) {
	ebitenutil.DebugPrintAt(screen, "HP:", panelX+10, panelY+80)
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
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/%d", player.HP, player.MaxHP), hpBarX+hpBarWidth+5, hpBarY)
}

// drawAPBar renders the AP bar as filled/empty dots (§9.1).
func (g *Game) drawAPBar(screen *ebiten.Image, panelX, y int, player *PlayerState) {
	ebitenutil.DebugPrintAt(screen, "AP:", panelX+10, y)
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
	ebitenutil.DebugPrintAt(screen, dotStr+fmt.Sprintf("(%d/%d)", ap, maxAP), panelX+35, y)
}

// drawActiveEffects renders active effects on the character panel (§9.4).
func (g *Game) drawActiveEffects(screen *ebiten.Image, panelX, y int, effects []EffectData) {
	if len(effects) == 0 {
		return
	}
	ebitenutil.DebugPrintAt(screen, "Effects:", panelX+10, y)
	for i, eff := range effects {
		if i >= 3 {
			break
		}
		icon := EffectIcon(eff.Type)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s %dt", icon, eff.Remaining), panelX+10+i*55, y+15)
	}
}

// drawMinimap renders a simplified 100×80 overhead map in the character panel (§9.2).
func (g *Game) drawMinimap(screen *ebiten.Image, x, y int) {
	const mapW, mapH = 100, 80

	// Background (unexplored = black)
	drawRect(screen, x, y, mapW, mapH, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	drawRectOutline(screen, x, y, mapW, mapH, color.RGBA{R: 80, G: 80, B: 100, A: 255})

	ebitenutil.DebugPrintAt(screen, "MAP", x+36, y-14)

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
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("(%d,%d)", player.Position.X, player.Position.Y), x+5, y+mapH-14)
}

// drawQuestTracker draws the compact quest tracker at the bottom of the character panel (§7).
func (g *Game) drawQuestTracker(screen *ebiten.Image, panelX, y int) {
	g.mu.RLock()
	ql := g.questLog
	g.mu.RUnlock()

	ebitenutil.DebugPrintAt(screen, "QUESTS", panelX+70, y)
	if ql == nil || len(ql.ActiveQuests) == 0 {
		ebitenutil.DebugPrintAt(screen, "(none)", panelX+10, y+15)
		return
	}
	count := 0
	for _, q := range ql.ActiveQuests {
		if count >= 3 {
			break
		}
		for _, obj := range q.Objectives {
			if !obj.Completed && count < 3 {
				ebitenutil.DebugPrintAt(screen, fmt.Sprintf("- %s [%d/%d]", truncateText(obj.Description, 18), obj.Progress, obj.Required), panelX+10, y+15+count*15)
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
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("STR:%d", attrs.Strength), panelX+10, panelY+120)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("DEX:%d", attrs.Dexterity), panelX+100, panelY+120)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("CON:%d", attrs.Constitution), panelX+10, panelY+135)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("INT:%d", attrs.Intelligence), panelX+100, panelY+135)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("WIS:%d", attrs.Wisdom), panelX+10, panelY+150)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("CHA:%d", attrs.Charisma), panelX+100, panelY+150)
}

// drawCombatInfo renders combat status and initiative order.
func (g *Game) drawCombatInfo(screen *ebiten.Image, panelX, combatY int, combat *CombatState) {
	ebitenutil.DebugPrintAt(screen, "COMBAT", panelX+70, combatY)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round: %d", combat.Round), panelX+10, combatY+20)
	if combat.CurrentTurn != "" {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Turn: %s", combat.CurrentTurn), panelX+10, combatY+35)
	}

	ebitenutil.DebugPrintAt(screen, "Initiative:", panelX+10, combatY+55)
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
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("%s%d. %s%s (%d)", marker, i+1, tag, entry.Name, entry.Initiative),
			panelX+10, combatY+70+i*15)
	}
}

// drawCombatLog renders the combat/game log panel.
func (g *Game) drawCombatLog(screen *ebiten.Image) {
	logX := 0
	logY := g.screenHeight - logPanelHeight - actionPanelHeight
	logWidth := g.screenWidth - charPanelWidth

	drawRect(screen, logX, logY, logWidth, logPanelHeight, color.RGBA{R: 25, G: 25, B: 35, A: 255})
	drawRectOutline(screen, logX, logY, logWidth, logPanelHeight, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "COMBAT LOG", logX+10, logY+5)

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
		ebitenutil.DebugPrintAt(screen, msg.Text, logX+10, y)
	}
}

// drawActionPanel renders the action buttons panel.
func (g *Game) drawActionPanel(screen *ebiten.Image) {
	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth

	drawRect(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 35, G: 35, B: 45, A: 255})
	drawRectOutline(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 70, G: 70, B: 90, A: 255})

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
		drawRectOutline(screen, bounds.X, bounds.Y, bounds.W, bounds.H, color.RGBA{R: 100, G: 100, B: 140, A: 255})
		ebitenutil.DebugPrintAt(screen, dirSymbols[name], bounds.X+4, bounds.Y+6)
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
		drawRectOutline(screen, bounds.X, bounds.Y, bounds.W, bounds.H, color.RGBA{R: 100, G: 100, B: 140, A: 255})
		ebitenutil.DebugPrintAt(screen, actionLabels[name], bounds.X+5, bounds.Y+8)
	}

	// Mode buttons (I/S/J/G shortcuts)
	modeX := g.screenWidth - charPanelWidth - 140
	modeY := panelY + 60
	modeLabels := []string{"[I]", "[S]", "[J]", "[G]"}
	for i, label := range modeLabels {
		ebitenutil.DebugPrintAt(screen, label, modeX+i*35, modeY)
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
