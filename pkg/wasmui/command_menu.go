//go:build js && wasm

// Game improvement #1: Authentic Gold Box-style command menu system.
// This implements the classic SSI Gold Box command interface with prominent
// single-letter keyboard shortcuts highlighted in gold. Commands are context-
// sensitive and change based on the current game mode (exploration vs combat).
// CommandDef, explorationCommands, and combatCommands live in command_menu_defs.go.

package wasmui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawCommandMenu renders a Gold Box-style command menu panel.
// The menu displays horizontally-arranged commands with highlighted keyboard shortcuts.
//
// Game improvement #1: Authentic Gold Box command interface styling.
func drawCommandMenu(screen *ebiten.Image, x, y, width int, commands []CommandDef, selectedAction CombatAction) {
	const (
		menuHeight    = 45 // Total height of command menu panel
		borderPadding = 5  // Padding inside the panel border
	)

	// Panel background with Gold Box-style deep dark color
	bgColor := color.RGBA{R: 22, G: 20, B: 32, A: 255}
	drawRect(screen, x, y, width, menuHeight, bgColor)

	// Double-pixel bold border (Gold Box authentic style)
	drawBoldPanelBorder(screen, x, y, width, menuHeight)

	// "COMMANDS" title at left edge (Gold Box style header)
	titleColor := ColorGold
	drawColoredText(screen, "COMMANDS:", x+borderPadding+2, y+6, titleColor)

	// Calculate command positions
	titleWidth := 75 // Approximate width of "COMMANDS:" text
	availWidth := width - titleWidth - borderPadding*2
	cmdCount := len(commands)
	if cmdCount == 0 || availWidth <= 0 {
		// Not enough space to render commands; bail out to avoid drawing off-panel.
		return
	}

	cmdWidth := calcCmdWidth(availWidth, cmdCount)
	startX := x + titleWidth + borderPadding

	// Draw each command and detect if any combat actions are present.
	hasCombatActions := false
	for i, cmd := range commands {
		cmdX := startX + i*cmdWidth
		drawCommand(screen, cmdX, y+4, cmd, selectedAction)
		if cmd.Action != CombatActionNone {
			hasCombatActions = true
		}
	}

	// Status line at bottom showing current mode.
	// "Tab: Cycle Target" is only relevant when combat actions are available.
	if hasCombatActions {
		statusY := y + menuHeight - 14
		drawColoredText(screen, "Tab: Cycle Target", x+width-130, statusY, ColorStatLabel)
	}
}

// cmdCharWidth is the pixel width per character in the debug font.
const cmdCharWidth = 6

// cmdMenuMinWidth is the minimum pixel width allocated per command in the menu.
const cmdMenuMinWidth = 65

// calcCmdWidth returns the per-command pixel width for a given available width and count.
// cmdMenuMinWidth is only applied when it will not cause commands to overflow availWidth.
func calcCmdWidth(availWidth, cmdCount int) int {
	w := availWidth / cmdCount
	if cmdMenuMinWidth*cmdCount <= availWidth && w < cmdMenuMinWidth {
		w = cmdMenuMinWidth
	}
	// Ensure we never return a zero-width command slot, which would break layout and hit-testing
	if w < 1 {
		w = 1
	}
	return w
}

// drawCommand renders a single command entry with highlighted keyboard shortcut.
//
// Format: "[K] Label" or "[Key] Label" where K is highlighted in gold
// Unavailable commands are dimmed.
func drawCommand(screen *ebiten.Image, x, y int, cmd CommandDef, selectedAction CombatAction) {
	// Determine colors based on availability and selection state
	var keyColor, textColor color.RGBA

	if !cmd.Available {
		// Dimmed colors for unavailable commands
		keyColor = color.RGBA{R: 80, G: 70, B: 50, A: 255}
		textColor = color.RGBA{R: 70, G: 70, B: 80, A: 255}
	} else if cmd.Action != CombatActionNone && cmd.Action == selectedAction {
		// Bright colors for selected action
		keyColor = ColorGoldHi
		textColor = ColorGoldHi
	} else {
		// Normal colors
		keyColor = ColorGold
		textColor = ColorStatValue
	}

	// Draw the command in "[K] Label" format with key in gold
	// Handle multi-character keys (like "Space" or "W/↑")
	bracketColor := textColor
	if cmd.Available && (cmd.Action == CombatActionNone || cmd.Action == selectedAction) {
		bracketColor = ColorStatLabel // Subtle brackets
	}

	// Build display text
	currentX := x

	// Opening bracket
	drawColoredText(screen, "[", currentX, y, bracketColor)
	currentX += 6

	// Key (highlighted in gold)
	drawColoredText(screen, cmd.Key, currentX, y, keyColor)
	currentX += len([]rune(cmd.Key))*cmdCharWidth + 1 // rune-aware width for multi-byte keys

	// Closing bracket and label
	closeBracket := "] "
	drawColoredText(screen, closeBracket, currentX, y, bracketColor)
	currentX += 12

	drawColoredText(screen, cmd.Label, currentX, y, textColor)

	// AP cost indicator for combat commands (if cost > 0)
	if cmd.APCost > 0 && cmd.Action != CombatActionNone {
		currentX += len([]rune(cmd.Label))*cmdCharWidth + 2 // rune-aware width
		costText := fmt.Sprintf("(%d)", cmd.APCost)
		costColor := ColorStatLabel
		if !cmd.Available {
			costColor = ColorAPDepleted
		}
		drawColoredText(screen, costText, currentX, y, costColor)
	}
}

// drawExplorationCommandMenu renders the command menu for exploration mode.
//
// Game improvement #1: Gold Box authentic exploration commands.
func (g *Game) drawExplorationCommandMenu(screen *ebiten.Image) {
	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth - charPanelWidth

	// Get exploration commands
	commands := explorationCommands()

	// Offset the command menu past the directional control pad so the
	// menu background does not paint over the bottom row of direction
	// buttons.  The direction pad extends to approximately x=102
	// (baseX 10 + padWidth 88 + 4 border padding).
	const dpadClearance = 108
	drawCommandMenu(screen, dpadClearance, panelY+45, panelWidth-dpadClearance, commands, CombatActionNone)
}

// drawCombatCommandMenu renders the command menu for combat mode.
//
// Game improvement #1: Gold Box authentic combat commands.
func (g *Game) drawCombatCommandMenu(screen *ebiten.Image) {
	g.mu.RLock()
	currentAP := 0
	if g.player != nil {
		currentAP = g.player.AP
	}
	currentAction := g.combatAction
	g.mu.RUnlock()

	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth - charPanelWidth

	// Get combat commands based on current AP
	commands := combatCommands(currentAP)

	// Draw the command menu
	drawCommandMenu(screen, 0, panelY+45, panelWidth, commands, currentAction)
}

// drawDirectionalControls renders the 8-directional movement pad.
// This is a Gold Box-style directional control layout for exploration.
//
// Game improvement #1: Compact directional controls integrated with command system.
func (g *Game) drawDirectionalControls(screen *ebiten.Image, x, y int) {
	const (
		btnSize   = 28  // Size of each direction button
		btnGap    = 2   // Gap between buttons
		gridWidth = 3   // 3x3 grid
		padWidth  = btnSize*gridWidth + btnGap*(gridWidth-1)
	)

	// Draw subtle background for the control pad
	bgColor := color.RGBA{R: 28, G: 26, B: 38, A: 255}
	drawRect(screen, x-2, y-2, padWidth+4, padWidth+4, bgColor)
	drawRectOutline(screen, x-2, y-2, padWidth+4, padWidth+4, ColorPanelBorder)

	// Direction buttons in 3x3 grid layout
	// Row 0: NW, N, NE
	// Row 1: W, (center), E
	// Row 2: SW, S, SE
	type dirBtn struct {
		row, col int
		key      string
		symbol   string
	}

	directions := []dirBtn{
		{0, 0, "nw", "NW"},
		{0, 1, "n", "N"},
		{0, 2, "ne", "NE"},
		{1, 0, "w", "W"},
		// Center is empty
		{1, 2, "e", "E"},
		{2, 0, "sw", "SW"},
		{2, 1, "s", "S"},
		{2, 2, "se", "SE"},
	}

	for _, d := range directions {
		btnX := x + d.col*(btnSize+btnGap)
		btnY := y + d.row*(btnSize+btnGap)

		// Button color - highlight on hover
		btnColor := color.RGBA{R: 50, G: 48, B: 65, A: 255}
		if g.hoveredButton == "dir_"+d.key {
			btnColor = color.RGBA{R: 70, G: 68, B: 95, A: 255}
		}

		drawRect(screen, btnX, btnY, btnSize, btnSize, btnColor)
		drawRectOutline(screen, btnX, btnY, btnSize, btnSize, ColorPanelBorder)

		// Direction symbol centered in button
		textX := btnX + (btnSize-len(d.symbol)*6)/2
		textY := btnY + (btnSize-12)/2
		drawColoredText(screen, d.symbol, textX, textY, ColorStatValue)
	}

	// Draw center dot (indicates player position)
	centerX := x + 1*(btnSize+btnGap) + btnSize/2 - 3
	centerY := y + 1*(btnSize+btnGap) + btnSize/2 - 3
	drawRect(screen, centerX, centerY, 6, 6, ColorGold)
}

// drawAPIndicator renders the Action Points display for combat.
// Shows current AP out of max AP with visual bar.
//
// Game improvement #1: Clear AP visibility for combat decisions.
func drawAPIndicator(screen *ebiten.Image, x, y, currentAP, maxAP int) {
	const (
		barWidth  = 100
		barHeight = 12
	)

	// Label
	drawColoredText(screen, "AP:", x, y, ColorGold)

	// Background bar
	barX := x + 25
	drawRect(screen, barX, y, barWidth, barHeight, color.RGBA{R: 40, G: 35, B: 55, A: 255})

	// Filled portion based on current AP
	if maxAP > 0 && currentAP > 0 {
		fillWidth := (barWidth * currentAP) / maxAP
		if fillWidth > barWidth {
			fillWidth = barWidth
		}

		// Color based on AP level: green (3+), yellow (2), red (1), empty (0)
		var apColor color.RGBA
		switch {
		case currentAP >= 3:
			apColor = ColorEffectBuff // Green
		case currentAP == 2:
			apColor = ColorEffectControl // Yellow
		case currentAP == 1:
			apColor = color.RGBA{R: 200, G: 100, B: 50, A: 255} // Orange
		default:
			apColor = ColorAPDepleted // Red
		}

		drawRect(screen, barX, y, fillWidth, barHeight, apColor)
	}

	// Border
	drawRectOutline(screen, barX, y, barWidth, barHeight, ColorPanelBorder)

	// Numeric display
	apText := fmt.Sprintf("%d/%d", currentAP, maxAP)
	textX := barX + (barWidth-len(apText)*6)/2
	drawColoredText(screen, apText, textX, y, ColorStatValue)
}
