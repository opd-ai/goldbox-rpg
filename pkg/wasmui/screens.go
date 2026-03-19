//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sirupsen/logrus"
)

// --- Splash Screen (§3.1) ---

func (g *Game) updateSplash() {
	// Any key press, click, or touch after connection → MainMenu
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()

	if connected {
		// Check for any key press or mouse click
		for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
			if inpututil.IsKeyJustPressed(k) {
				g.mu.Lock()
				g.screenState = ScreenMainMenu
				g.mu.Unlock()
				return
			}
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.mu.Lock()
			g.screenState = ScreenMainMenu
			g.mu.Unlock()
			return
		}
		// Touch tap also advances
		if tapped, _, _ := g.touchState.HasTap(); tapped {
			g.mu.Lock()
			g.screenState = ScreenMainMenu
			g.mu.Unlock()
			return
		}
	}
}

func (g *Game) drawSplash(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 15, G: 15, B: 25, A: 255})

	// Title
	drawColoredText(screen, "GOLDBOX RPG ENGINE", 310, 180, ColorGold)
	drawColoredText(screen, "==================", 310, 195, ColorGold)

	// Status
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()

	if connected {
		drawRect(screen, 250, 325, 300, 16, color.RGBA{R: 50, G: 200, B: 50, A: 255})
		drawColoredText(screen, "Connected", 370, 300, ColorPlayerName)
		drawColoredText(screen, "Tap or press any key to continue", 270, 400, ColorStatValue)
	} else {
		drawRect(screen, 250, 325, 150, 16, color.RGBA{R: 100, G: 100, B: 200, A: 255})
		drawColoredText(screen, "Connecting...", 360, 300, ColorStatLabel)
	}

	// Version
	drawColoredText(screen, "v1.0  -  opd-ai  -  MIT License", 290, 560, ColorStatLabel)
}

// --- Main Menu (§3.2) ---

const (
	menuNewGame   = 0
	menuContinue  = 1
	menuSettings  = 2
	menuQuit      = 3
	menuItemCount = 4

	// Menu button layout
	menuBtnX       = 300
	menuBtnW       = 200
	menuBtnH       = 40
	menuBtnY       = 230
	menuBtnSpacing = 50
)

func (g *Game) updateMainMenu() {
	g.handleMainMenuNavigation()
	g.handleMainMenuShortcuts()
	g.handleMainMenuInput()
}

// handleMainMenuNavigation processes arrow key navigation.
func (g *Game) handleMainMenuNavigation() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.mu.Lock()
		if g.menuIndex > 0 {
			g.menuIndex--
		}
		g.mu.Unlock()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.mu.Lock()
		if g.menuIndex < menuItemCount-1 {
			g.menuIndex++
		}
		g.mu.Unlock()
	}
}

// handleMainMenuShortcuts processes keyboard shortcuts (F1, Escape, Enter).
func (g *Game) handleMainMenuShortcuts() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.mu.Lock()
		g.menuIndex = menuNewGame
		g.mu.Unlock()
		g.activateMenuItem(menuNewGame)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.activateMenuItem(menuQuit)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.mu.RLock()
		idx := g.menuIndex
		g.mu.RUnlock()
		g.activateMenuItem(idx)
	}
}

// handleMainMenuInput processes mouse and touch input on menu items.
func (g *Game) handleMainMenuInput() {
	var x, y int
	var hasInput bool

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y = ebiten.CursorPosition()
		hasInput = true
	} else if tapped, tx, ty := g.touchState.HasTap(); tapped {
		x, y = tx, ty
		hasInput = true
	}

	if !hasInput {
		return
	}

	if idx := hitTestMenuButton(x, y); idx >= 0 {
		g.mu.Lock()
		g.menuIndex = idx
		g.mu.Unlock()
		g.activateMenuItem(idx)
	}
}

// hitTestMenuButton returns the menu item index at (x,y), or -1 if none.
func hitTestMenuButton(x, y int) int {
	if x < menuBtnX || x > menuBtnX+menuBtnW {
		return -1
	}
	for i := 0; i < menuItemCount; i++ {
		btnY := menuBtnY + i*menuBtnSpacing
		if y >= btnY && y <= btnY+menuBtnH {
			return i
		}
	}
	return -1
}

func (g *Game) activateMenuItem(idx int) {
	switch idx {
	case menuNewGame:
		g.mu.Lock()
		g.mode = ModeAdventureSelect
		g.mu.Unlock()
		g.adventureScreen.RefreshAdventures(g)
	case menuContinue:
		// Resume saved session if available
		g.mu.RLock()
		hasPlayer := g.player != nil
		g.mu.RUnlock()
		if hasPlayer {
			g.mu.Lock()
			g.screenState = ScreenExploration
			g.mu.Unlock()
		}
	case menuSettings:
		g.mu.Lock()
		g.overlays.ShowSettings = true
		g.mu.Unlock()
	case menuQuit:
		go func() {
			if _, err := g.rpcClient.LeaveGame(); err != nil {
				logrus.WithField("error", err).Warn("failed to leave game during quit")
			}
			g.mu.Lock()
			g.player = nil
			g.combat = nil
			g.sessionID = ""
			g.currentAdventure = nil
			g.screenState = ScreenSplash
			g.mu.Unlock()
		}()
	}
}

func (g *Game) drawMainMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Title
	drawColoredText(screen, "GOLDBOX RPG ENGINE", 310, 120, ColorGold)

	// Menu items
	labels := []string{"New Game", "Continue", "Settings", "Quit"}
	g.mu.RLock()
	idx := g.menuIndex
	hasPlayer := g.player != nil
	g.mu.RUnlock()

	for i, label := range labels {
		btnX, btnY, btnW, btnH := 300, 230+i*50, 200, 40

		btnColor := color.RGBA{R: 50, G: 50, B: 70, A: 255}
		textStr := label
		textClr := ColorStatValue

		// Gray out Continue if no session
		if i == menuContinue && !hasPlayer {
			btnColor = color.RGBA{R: 40, G: 40, B: 50, A: 255}
			textStr = "(no save)"
			textClr = ColorStatLabel
		}

		if i == idx {
			btnColor = color.RGBA{R: 80, G: 80, B: 120, A: 255}
			textClr = ColorGoldHi
		}

		drawRect(screen, btnX, btnY, btnW, btnH, btnColor)
		drawRectOutline(screen, btnX, btnY, btnW, btnH, ColorPanelBorder)
		drawColoredText(screen, textStr, btnX+10, btnY+12, textClr)

		// Key hints
		if i == menuNewGame {
			drawColoredText(screen, "F1", btnX+btnW+10, btnY+12, ColorStatLabel)
		}
	}

	// Connection indicator
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()
	statusText := "Disconnected"
	dotColor := color.RGBA{R: 200, G: 50, B: 50, A: 255}
	statusClr := ColorEnemyName
	if connected {
		statusText = "Connected"
		dotColor = color.RGBA{R: 50, G: 200, B: 50, A: 255}
		statusClr = ColorPlayerName
	}
	drawRect(screen, 340, 570, 12, 12, dotColor)
	drawColoredText(screen, "Connection: "+statusText, 360, 570, statusClr)
}

// --- Victory Screen (§11) ---

func (g *Game) updateVictory() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		// "Return to Menu" button
		if x >= 280 && x <= 520 && y >= 430 && y <= 465 {
			g.returnToMenu()
			return
		}
		// "Next Adventure" button
		if x >= 280 && x <= 520 && y >= 475 && y <= 510 {
			g.mu.Lock()
			g.mode = ModeAdventureSelect
			g.mu.Unlock()
			g.adventureScreen.RefreshAdventures(g)
			return
		}
		// Default: Enter → return to menu
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.returnToMenu()
		}
	}

	// Touch tap on victory screen buttons
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		if tx >= 280 && tx <= 520 && ty >= 430 && ty <= 465 {
			g.returnToMenu()
			return
		}
		if tx >= 280 && tx <= 520 && ty >= 475 && ty <= 510 {
			g.mu.Lock()
			g.mode = ModeAdventureSelect
			g.mu.Unlock()
			g.adventureScreen.RefreshAdventures(g)
		}
	}
}

func (g *Game) drawVictory(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 20, G: 25, B: 15, A: 255})

	drawColoredText(screen, "*** VICTORY! ***", 330, 60, ColorGoldHi)

	g.mu.RLock()
	v := g.victoryData
	g.mu.RUnlock()

	if v != nil {
		drawColoredText(screen, "Adventure: "+v.AdventureTitle, 280, 120, ColorGold)
		drawColoredText(screen, "Completed!", 350, 145, ColorPlayerName)
		drawColoredText(screen, "--- SUMMARY ---", 330, 185, ColorGold)
		drawColoredText(screen, "Time Played:     "+v.TimePlayed, 280, 215, ColorStatValue)
		drawColoredText(screen, fmt.Sprintf("Quests Complete: %d/%d", v.QuestsComplete, v.QuestsTotal), 280, 235, ColorStatValue)
		drawColoredText(screen, fmt.Sprintf("Enemies Defeated: %d", v.EnemiesDefeated), 280, 255, ColorStatValue)
		drawColoredText(screen, fmt.Sprintf("Gold Earned:     %d", v.GoldEarned), 280, 275, ColorStatValue)
		drawColoredText(screen, fmt.Sprintf("XP Earned:       %d", v.XPEarned), 280, 295, ColorStatValue)
		if v.LevelFrom != v.LevelTo {
			drawColoredText(screen, fmt.Sprintf("Level: %d -> %d", v.LevelFrom, v.LevelTo), 280, 315, ColorGoldHi)
		}
	}

	// Buttons
	drawRect(screen, 280, 430, 240, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 280, 430, 240, 35, ColorPanelBorder)
	drawColoredText(screen, "Return to Menu", 330, 440, ColorStatValue)

	drawRect(screen, 280, 475, 240, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 280, 475, 240, 35, ColorPanelBorder)
	drawColoredText(screen, "Next Adventure", 330, 485, ColorStatValue)
}

// --- Defeat Screen (§11) ---

func (g *Game) updateDefeat() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		// "Return to Menu" button
		if x >= 280 && x <= 520 && y >= 380 && y <= 415 {
			g.returnToMenu()
			return
		}
		// "Try Again" button
		if x >= 280 && x <= 520 && y >= 425 && y <= 460 {
			// Reload last state
			g.mu.Lock()
			g.screenState = ScreenExploration
			g.mu.Unlock()
			go g.refreshGameState()
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.returnToMenu()
		}
	}

	// Touch tap on defeat screen buttons
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		if tx >= 280 && tx <= 520 && ty >= 380 && ty <= 415 {
			g.returnToMenu()
			return
		}
		if tx >= 280 && tx <= 520 && ty >= 425 && ty <= 460 {
			g.mu.Lock()
			g.screenState = ScreenExploration
			g.mu.Unlock()
			go g.refreshGameState()
		}
	}
}

func (g *Game) drawDefeat(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 30, G: 15, B: 15, A: 255})

	drawColoredText(screen, "*** DEFEAT ***", 340, 80, ColorEnemyName)
	drawColoredText(screen, "Your adventurer has fallen...", 290, 140, ColorStatLabel)

	g.mu.RLock()
	d := g.defeatData
	g.mu.RUnlock()

	if d != nil {
		drawColoredText(screen, "Last Location: "+d.LastLocation, 280, 200, ColorStatValue)
		drawColoredText(screen, "Cause: "+d.CauseOfDeath, 280, 225, ColorStatValue)
	}

	// Buttons
	drawRect(screen, 280, 380, 240, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 280, 380, 240, 35, ColorPanelBorder)
	drawColoredText(screen, "Return to Menu", 330, 390, ColorStatValue)

	drawRect(screen, 280, 425, 240, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 280, 425, 240, 35, ColorPanelBorder)
	drawColoredText(screen, "Try Again", 345, 435, ColorStatValue)
}

// returnToMenu performs the return-to-menu flow per §11.
func (g *Game) returnToMenu() {
	go func() {
		if _, err := g.rpcClient.LeaveGame(); err != nil {
			logrus.WithField("error", err).Warn("failed to leave game during return to menu")
		}
		g.mu.Lock()
		g.player = nil
		g.combat = nil
		g.currentAdventure = nil
		g.victoryData = nil
		g.defeatData = nil
		g.mode = ModeNormal
		g.screenState = ScreenMainMenu
		g.mu.Unlock()
	}()
}
