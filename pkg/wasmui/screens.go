//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
		}
	}
}

func (g *Game) drawSplash(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 15, G: 15, B: 25, A: 255})

	// Title
	ebitenutil.DebugPrintAt(screen, "GOLDBOX RPG ENGINE", 310, 180)
	ebitenutil.DebugPrintAt(screen, "==================", 310, 195)

	// Status
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()

	if connected {
		drawRect(screen, 250, 325, 300, 16, color.RGBA{R: 50, G: 200, B: 50, A: 255})
		ebitenutil.DebugPrintAt(screen, "Connected", 370, 300)
		ebitenutil.DebugPrintAt(screen, "Press any key to continue", 300, 400)
	} else {
		drawRect(screen, 250, 325, 150, 16, color.RGBA{R: 100, G: 100, B: 200, A: 255})
		ebitenutil.DebugPrintAt(screen, "Connecting...", 360, 300)
	}

	// Version
	ebitenutil.DebugPrintAt(screen, "v1.0  -  opd-ai  -  MIT License", 290, 560)
}

// --- Main Menu (§3.2) ---

const (
	menuNewGame   = 0
	menuContinue  = 1
	menuSettings  = 2
	menuQuit      = 3
	menuItemCount = 4
)

func (g *Game) updateMainMenu() {
	// Arrow keys for menu navigation
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

	// F1 → New Game shortcut
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.mu.Lock()
		g.menuIndex = menuNewGame
		g.mu.Unlock()
		g.activateMenuItem(menuNewGame)
		return
	}

	// Escape → Quit
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.activateMenuItem(menuQuit)
		return
	}

	// Enter to select
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.mu.RLock()
		idx := g.menuIndex
		g.mu.RUnlock()
		g.activateMenuItem(idx)
		return
	}

	// Mouse clicks on menu items
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		btnX, btnW := 300, 200
		btnH := 40
		for i := 0; i < menuItemCount; i++ {
			btnY := 230 + i*50
			if x >= btnX && x <= btnX+btnW && y >= btnY && y <= btnY+btnH {
				g.mu.Lock()
				g.menuIndex = i
				g.mu.Unlock()
				g.activateMenuItem(i)
				return
			}
		}
	}

	// Touch taps on menu items
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		btnX, btnW := 300, 200
		btnH := 40
		for i := 0; i < menuItemCount; i++ {
			btnY := 230 + i*50
			if tx >= btnX && tx <= btnX+btnW && ty >= btnY && ty <= btnY+btnH {
				g.mu.Lock()
				g.menuIndex = i
				g.mu.Unlock()
				g.activateMenuItem(i)
				return
			}
		}
	}
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
			_, _ = g.rpcClient.LeaveGame()
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
	ebitenutil.DebugPrintAt(screen, "GOLDBOX RPG ENGINE", 310, 120)

	// Menu items
	labels := []string{"New Game", "Continue", "Settings", "Quit"}
	g.mu.RLock()
	idx := g.menuIndex
	hasPlayer := g.player != nil
	g.mu.RUnlock()

	for i, label := range labels {
		btnX, btnY, btnW, btnH := 300, 230+i*50, 200, 40

		btnColor := color.RGBA{R: 50, G: 50, B: 70, A: 255}
		textColor := label

		// Gray out Continue if no session
		if i == menuContinue && !hasPlayer {
			btnColor = color.RGBA{R: 40, G: 40, B: 50, A: 255}
			textColor = "(no save)"
		}

		if i == idx {
			btnColor = color.RGBA{R: 80, G: 80, B: 120, A: 255}
		}

		drawRect(screen, btnX, btnY, btnW, btnH, btnColor)
		drawRectOutline(screen, btnX, btnY, btnW, btnH, color.RGBA{R: 100, G: 100, B: 140, A: 255})
		ebitenutil.DebugPrintAt(screen, textColor, btnX+10, btnY+12)

		// Key hints
		if i == menuNewGame {
			ebitenutil.DebugPrintAt(screen, "F1", btnX+btnW+10, btnY+12)
		}
	}

	// Connection indicator
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()
	statusText := "Disconnected"
	dotColor := color.RGBA{R: 200, G: 50, B: 50, A: 255}
	if connected {
		statusText = "Connected"
		dotColor = color.RGBA{R: 50, G: 200, B: 50, A: 255}
	}
	drawRect(screen, 340, 570, 12, 12, dotColor)
	ebitenutil.DebugPrintAt(screen, "Connection: "+statusText, 360, 570)
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

	ebitenutil.DebugPrintAt(screen, "*** VICTORY! ***", 330, 60)

	g.mu.RLock()
	v := g.victoryData
	g.mu.RUnlock()

	if v != nil {
		ebitenutil.DebugPrintAt(screen, "Adventure: "+v.AdventureTitle, 280, 120)
		ebitenutil.DebugPrintAt(screen, "Completed!", 350, 145)
		ebitenutil.DebugPrintAt(screen, "--- SUMMARY ---", 330, 185)
		ebitenutil.DebugPrintAt(screen, "Time Played:     "+v.TimePlayed, 280, 215)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Quests Complete: %d/%d", v.QuestsComplete, v.QuestsTotal), 280, 235)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Enemies Defeated: %d", v.EnemiesDefeated), 280, 255)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Gold Earned:     %d", v.GoldEarned), 280, 275)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("XP Earned:       %d", v.XPEarned), 280, 295)
		if v.LevelFrom != v.LevelTo {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level: %d -> %d", v.LevelFrom, v.LevelTo), 280, 315)
		}
	}

	// Buttons
	drawRect(screen, 280, 430, 240, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 280, 430, 240, 35, color.RGBA{R: 100, G: 100, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "Return to Menu", 330, 440)

	drawRect(screen, 280, 475, 240, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 280, 475, 240, 35, color.RGBA{R: 100, G: 100, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "Next Adventure", 330, 485)
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

	ebitenutil.DebugPrintAt(screen, "*** DEFEAT ***", 340, 80)
	ebitenutil.DebugPrintAt(screen, "Your adventurer has fallen...", 290, 140)

	g.mu.RLock()
	d := g.defeatData
	g.mu.RUnlock()

	if d != nil {
		ebitenutil.DebugPrintAt(screen, "Last Location: "+d.LastLocation, 280, 200)
		ebitenutil.DebugPrintAt(screen, "Cause: "+d.CauseOfDeath, 280, 225)
	}

	// Buttons
	drawRect(screen, 280, 380, 240, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 280, 380, 240, 35, color.RGBA{R: 100, G: 100, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "Return to Menu", 330, 390)

	drawRect(screen, 280, 425, 240, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 280, 425, 240, 35, color.RGBA{R: 100, G: 100, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "Try Again", 345, 435)
}

// returnToMenu performs the return-to-menu flow per §11.
func (g *Game) returnToMenu() {
	go func() {
		_, _ = g.rpcClient.LeaveGame()
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
