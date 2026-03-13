//go:build js && wasm

// Package wasmui provides the Ebitengine/WASM-based game UI.
package wasmui

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// pixelImage is a 1x1 white image used for efficient rectangle drawing.
// Reusing this image avoids per-frame allocations in drawRect.
var pixelImage *ebiten.Image

func init() {
	pixelImage = ebiten.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	pixelImage.Fill(color.White)
}

const (
	// UI layout constants
	ScreenWidth       = 800
	ScreenHeight      = 600
	charPanelWidth    = 200
	logPanelHeight    = 150
	actionPanelHeight = 100
	tileSize          = 32
)

// Game implements ebiten.Game interface for the Gold Box RPG UI.
type Game struct {
	// RPC communication
	rpcClient *RPCClient
	connected bool

	// Game state (protected by mu)
	mu               sync.RWMutex
	player           *PlayerState
	combat           *CombatState
	mode             UIMode
	sessionID        string
	currentAdventure *Adventure

	// UI state (protected by mu)
	logMessages    []LogMessage
	maxLogMessages int
	selectedAction string
	hoveredButton  string

	// Error display (protected by mu)
	lastError    string
	errorTimeout time.Time

	// Input state (only accessed from main goroutine)
	lastInputTime time.Time
	inputCooldown time.Duration

	// Screen dimensions (only accessed from main goroutine)
	screenWidth  int
	screenHeight int

	// Adventure selection screen
	adventureScreen *AdventureScreen
}

// NewGame creates and initializes a new Game instance.
func NewGame() (*Game, error) {
	g := &Game{
		rpcClient:       NewRPCClient(),
		maxLogMessages:  100,
		inputCooldown:   100 * time.Millisecond,
		screenWidth:     ScreenWidth,
		screenHeight:    ScreenHeight,
		mode:            ModeNormal,
		logMessages:     make([]LogMessage, 0),
		adventureScreen: NewAdventureScreen(),
	}

	// Set up RPC callbacks
	g.rpcClient.SetOnConnected(func() {
		g.mu.Lock()
		g.connected = true
		g.mu.Unlock()
		g.addLogMessage("Connected to server", MessageSystem)
	})

	g.rpcClient.SetOnDisconnect(func(reason string) {
		g.mu.Lock()
		g.connected = false
		g.mu.Unlock()
		g.addLogMessage(fmt.Sprintf("Disconnected: %s", reason), MessageError)
	})

	g.rpcClient.SetOnError(func(err error) {
		g.showError(err.Error())
	})

	// Connect to server (async in WASM)
	go g.connectAndJoin()

	return g, nil
}

// connectAndJoin handles the initial connection and game join.
func (g *Game) connectAndJoin() {
	g.addLogMessage("Connecting to server...", MessageSystem)

	if err := g.rpcClient.Connect(); err != nil {
		g.showError(fmt.Sprintf("Connection failed: %v", err))
		return
	}

	// Auto-join game
	result, err := g.rpcClient.JoinGame("Player1")
	if err != nil {
		g.showError(fmt.Sprintf("Failed to join game: %v", err))
		return
	}

	if result.Success {
		g.mu.Lock()
		g.sessionID = result.SessionID
		g.mu.Unlock()
		g.addLogMessage("Joined game successfully", MessageSystem)

		// Fetch initial game state
		g.refreshGameState()
	}
}

// extractPlayerState attempts to extract player state from the session data.
func extractPlayerState(sessions map[string]interface{}, sessionID string) *PlayerState {
	if sessions == nil || sessionID == "" {
		return nil
	}
	sessionData, ok := sessions[sessionID]
	if !ok {
		return nil
	}
	sessionMap, ok := sessionData.(map[string]interface{})
	if !ok {
		return nil
	}
	playerData, ok := sessionMap["player"]
	if !ok || playerData == nil {
		return nil
	}
	data, err := json.Marshal(playerData)
	if err != nil {
		return nil
	}
	var player PlayerState
	if err := json.Unmarshal(data, &player); err != nil {
		return nil
	}
	return &player
}

// extractCombatState attempts to extract combat state from the world data.
func extractCombatState(world interface{}) *CombatState {
	if world == nil {
		return nil
	}
	worldMap, ok := world.(map[string]interface{})
	if !ok {
		return nil
	}
	combatData, ok := worldMap["combat"]
	if !ok || combatData == nil {
		return nil
	}
	data, err := json.Marshal(combatData)
	if err != nil {
		return nil
	}
	var combat CombatState
	if err := json.Unmarshal(data, &combat); err != nil {
		return nil
	}
	return &combat
}

// refreshGameState fetches the current game state from the server.
func (g *Game) refreshGameState() {
	g.mu.RLock()
	connected := g.connected
	sessionID := g.sessionID
	g.mu.RUnlock()

	if !connected {
		return
	}

	stateResult, err := g.rpcClient.GetGameState()
	if err != nil {
		g.showError(fmt.Sprintf("Failed to get game state: %v", err))
		return
	}

	if player := extractPlayerState(stateResult.Sessions, sessionID); player != nil {
		g.mu.Lock()
		g.player = player
		g.mu.Unlock()
	}

	if combat := extractCombatState(stateResult.World); combat != nil {
		g.mu.Lock()
		g.combat = combat
		g.mu.Unlock()
	}
}

// Update implements ebiten.Game interface.
func (g *Game) Update() error {
	// Handle adventure selection mode
	g.mu.RLock()
	mode := g.mode
	g.mu.RUnlock()

	if mode == ModeAdventureSelect {
		g.adventureScreen.Update(g)
		return nil
	}

	// Handle keyboard input
	g.handleKeyboardInput()

	// Handle mouse input
	g.handleMouseInput()

	return nil
}

// handleKeyboardInput processes keyboard events.
func (g *Game) handleKeyboardInput() {
	// Check input cooldown
	if time.Since(g.lastInputTime) < g.inputCooldown {
		return
	}

	// Movement keys
	directions := map[ebiten.Key]string{
		ebiten.KeyW:          "north",
		ebiten.KeyS:          "south",
		ebiten.KeyA:          "west",
		ebiten.KeyD:          "east",
		ebiten.KeyQ:          "northwest",
		ebiten.KeyE:          "northeast",
		ebiten.KeyZ:          "southwest",
		ebiten.KeyC:          "southeast",
		ebiten.KeyArrowUp:    "north",
		ebiten.KeyArrowDown:  "south",
		ebiten.KeyArrowLeft:  "west",
		ebiten.KeyArrowRight: "east",
		ebiten.KeyNumpad8:    "north",
		ebiten.KeyNumpad2:    "south",
		ebiten.KeyNumpad4:    "west",
		ebiten.KeyNumpad6:    "east",
		ebiten.KeyNumpad7:    "northwest",
		ebiten.KeyNumpad9:    "northeast",
		ebiten.KeyNumpad1:    "southwest",
		ebiten.KeyNumpad3:    "southeast",
	}

	for key, direction := range directions {
		if inpututil.IsKeyJustPressed(key) {
			g.handleMove(direction)
			g.lastInputTime = time.Now()
			return
		}
	}

	// Action keys
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.handleEndTurn()
		g.lastInputTime = time.Now()
	}

	// Adventure selection (F1 key)
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.mu.Lock()
		g.mode = ModeAdventureSelect
		g.mu.Unlock()
		g.adventureScreen.RefreshAdventures(g)
		g.lastInputTime = time.Now()
	}
}

// handleMouseInput processes mouse events.
func (g *Game) handleMouseInput() {
	x, y := ebiten.CursorPosition()

	// Check button hover states
	g.hoveredButton = g.getButtonAtPosition(x, y)

	// Handle clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.handleClick(x, y)
	}
}

// getButtonAtPosition returns the button ID at the given position.
func (g *Game) getButtonAtPosition(x, y int) string {
	// Direction buttons (bottom-left)
	dirButtons := g.getDirectionButtonBounds()
	for name, bounds := range dirButtons {
		if x >= bounds.X && x < bounds.X+bounds.W &&
			y >= bounds.Y && y < bounds.Y+bounds.H {
			return "dir_" + name
		}
	}

	// Action buttons (bottom-center)
	actionButtons := g.getActionButtonBounds()
	for name, bounds := range actionButtons {
		if x >= bounds.X && x < bounds.X+bounds.W &&
			y >= bounds.Y && y < bounds.Y+bounds.H {
			return "action_" + name
		}
	}

	return ""
}

// Rect represents a rectangle for button bounds.
type Rect struct {
	X, Y, W, H int
}

// getDirectionButtonBounds returns the bounds for direction buttons.
func (g *Game) getDirectionButtonBounds() map[string]Rect {
	baseX := 10
	baseY := g.screenHeight - actionPanelHeight + 10
	btnSize := 28

	return map[string]Rect{
		"nw": {X: baseX, Y: baseY, W: btnSize, H: btnSize},
		"n":  {X: baseX + btnSize + 2, Y: baseY, W: btnSize, H: btnSize},
		"ne": {X: baseX + (btnSize+2)*2, Y: baseY, W: btnSize, H: btnSize},
		"w":  {X: baseX, Y: baseY + btnSize + 2, W: btnSize, H: btnSize},
		"e":  {X: baseX + (btnSize+2)*2, Y: baseY + btnSize + 2, W: btnSize, H: btnSize},
		"sw": {X: baseX, Y: baseY + (btnSize+2)*2, W: btnSize, H: btnSize},
		"s":  {X: baseX + btnSize + 2, Y: baseY + (btnSize+2)*2, W: btnSize, H: btnSize},
		"se": {X: baseX + (btnSize+2)*2, Y: baseY + (btnSize+2)*2, W: btnSize, H: btnSize},
	}
}

// getActionButtonBounds returns the bounds for action buttons.
func (g *Game) getActionButtonBounds() map[string]Rect {
	baseX := 120
	baseY := g.screenHeight - actionPanelHeight + 15
	btnWidth := 80
	btnHeight := 30
	spacing := 10

	return map[string]Rect{
		"attack":  {X: baseX, Y: baseY, W: btnWidth, H: btnHeight},
		"cast":    {X: baseX + btnWidth + spacing, Y: baseY, W: btnWidth, H: btnHeight},
		"item":    {X: baseX + (btnWidth+spacing)*2, Y: baseY, W: btnWidth, H: btnHeight},
		"endturn": {X: baseX + (btnWidth+spacing)*3, Y: baseY, W: btnWidth, H: btnHeight},
	}
}

// handleClick processes a mouse click at the given position.
func (g *Game) handleClick(x, y int) {
	button := g.getButtonAtPosition(x, y)
	if button == "" {
		return
	}

	// Direction buttons
	dirMap := map[string]string{
		"dir_n":  "north",
		"dir_s":  "south",
		"dir_e":  "east",
		"dir_w":  "west",
		"dir_ne": "northeast",
		"dir_nw": "northwest",
		"dir_se": "southeast",
		"dir_sw": "southwest",
	}

	if direction, ok := dirMap[button]; ok {
		g.handleMove(direction)
		return
	}

	// Action buttons
	switch button {
	case "action_attack":
		g.addLogMessage("Attack mode - select target", MessageInfo)
		g.selectedAction = "attack"
	case "action_cast":
		g.addLogMessage("Cast spell - select spell and target", MessageInfo)
		g.mu.Lock()
		g.selectedAction = "cast"
		g.mu.Unlock()
	case "action_item":
		g.addLogMessage("Use item - select item", MessageInfo)
		g.mu.Lock()
		g.selectedAction = "item"
		g.mu.Unlock()
	case "action_endturn":
		g.handleEndTurn()
	}
}

// handleMove sends a move command to the server.
func (g *Game) handleMove(direction string) {
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()

	if !connected {
		g.showError("Not connected to server")
		return
	}

	go func() {
		result, err := g.rpcClient.Move(direction)
		if err != nil {
			g.showError(fmt.Sprintf("Move failed: %v", err))
			return
		}

		if result.Success {
			g.addLogMessage(fmt.Sprintf("Moved %s", direction), MessageInfo)
			if result.NewPosition != nil {
				g.mu.Lock()
				if g.player != nil {
					g.player.Position = *result.NewPosition
				}
				g.mu.Unlock()
			}
		} else if result.Message != "" {
			g.addLogMessage(result.Message, MessageWarning)
		}
	}()
}

// handleEndTurn sends an end turn command to the server.
func (g *Game) handleEndTurn() {
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()

	if !connected {
		g.showError("Not connected to server")
		return
	}

	go func() {
		result, err := g.rpcClient.EndTurn()
		if err != nil {
			g.showError(fmt.Sprintf("End turn failed: %v", err))
			return
		}

		if result.Success {
			g.addLogMessage(fmt.Sprintf("Turn ended. Next: %s", result.NextTurn), MessageCombat)
		}
	}()
}

// Draw implements ebiten.Game interface.
func (g *Game) Draw(screen *ebiten.Image) {
	// Handle adventure selection mode
	g.mu.RLock()
	mode := g.mode
	g.mu.RUnlock()

	if mode == ModeAdventureSelect {
		g.adventureScreen.Draw(screen, g)
		return
	}

	// Clear background
	screen.Fill(color.RGBA{R: 30, G: 30, B: 40, A: 255})

	// Draw main game viewport
	g.drawViewport(screen)

	// Draw character panel (right side)
	g.drawCharacterPanel(screen)

	// Draw combat log (bottom)
	g.drawCombatLog(screen)

	// Draw action panel (bottom)
	g.drawActionPanel(screen)

	// Draw error message if any
	g.drawError(screen)

	// Draw connection status
	g.drawConnectionStatus(screen)
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

// drawCharacterPanel renders the character information panel.
func (g *Game) drawCharacterPanel(screen *ebiten.Image) {
	panelX := g.screenWidth - charPanelWidth
	panelY := 0
	panelHeight := g.screenHeight - actionPanelHeight

	// Panel background
	drawRect(screen, panelX, panelY, charPanelWidth, panelHeight, color.RGBA{R: 40, G: 40, B: 50, A: 255})
	drawRectOutline(screen, panelX, panelY, charPanelWidth, panelHeight, color.RGBA{R: 80, G: 80, B: 100, A: 255})

	// Title
	ebitenutil.DebugPrintAt(screen, "CHARACTER", panelX+60, panelY+10)

	// Get player and combat state with lock
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
		g.drawCombatInfo(screen, panelX, panelY+220, combat)
	}
}

// drawPlayerStats renders player character statistics.
func (g *Game) drawPlayerStats(screen *ebiten.Image, panelX, panelY int, player *PlayerState) {
	// Character name
	ebitenutil.DebugPrintAt(screen, player.Name, panelX+10, panelY+40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level %d %s", player.Level, player.Class), panelX+10, panelY+55)

	// HP bar
	g.drawHPBar(screen, panelX, panelY, player)

	// Attributes
	g.drawAttributes(screen, panelX, panelY, player.Attributes)

	// Position
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Pos: (%d, %d)", player.Position.X, player.Position.Y), panelX+10, panelY+185)
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
	ebitenutil.DebugPrintAt(screen, "ATTRIBUTES", panelX+50, panelY+110)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("STR: %d", attrs.Strength), panelX+10, panelY+130)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("DEX: %d", attrs.Dexterity), panelX+100, panelY+130)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("CON: %d", attrs.Constitution), panelX+10, panelY+145)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("INT: %d", attrs.Intelligence), panelX+100, panelY+145)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("WIS: %d", attrs.Wisdom), panelX+10, panelY+160)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("CHA: %d", attrs.Charisma), panelX+100, panelY+160)
}

// drawCombatInfo renders combat status and initiative order.
func (g *Game) drawCombatInfo(screen *ebiten.Image, panelX, combatY int, combat *CombatState) {
	ebitenutil.DebugPrintAt(screen, "COMBAT", panelX+70, combatY)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round: %d", combat.Round), panelX+10, combatY+20)
	if combat.CurrentTurn != "" {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Turn: %s", combat.CurrentTurn), panelX+10, combatY+35)
	}

	// Initiative order
	ebitenutil.DebugPrintAt(screen, "Initiative:", panelX+10, combatY+55)
	for i, entry := range combat.Initiative {
		if i >= 5 {
			break // Limit display
		}
		colorStr := ""
		if entry.IsPlayer {
			colorStr = "[P]"
		}
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("%d. %s%s (%d)", i+1, colorStr, entry.Name, entry.Initiative),
			panelX+10, combatY+70+i*15)
	}
}

// drawCombatLog renders the combat/game log panel.
func (g *Game) drawCombatLog(screen *ebiten.Image) {
	logX := 0
	logY := g.screenHeight - logPanelHeight - actionPanelHeight
	logWidth := g.screenWidth - charPanelWidth

	// Background
	drawRect(screen, logX, logY, logWidth, logPanelHeight, color.RGBA{R: 25, G: 25, B: 35, A: 255})
	drawRectOutline(screen, logX, logY, logWidth, logPanelHeight, color.RGBA{R: 60, G: 60, B: 80, A: 255})

	// Title
	ebitenutil.DebugPrintAt(screen, "COMBAT LOG", logX+10, logY+5)

	// Get a copy of messages for thread-safe rendering
	g.mu.RLock()
	messages := make([]LogMessage, len(g.logMessages))
	copy(messages, g.logMessages)
	g.mu.RUnlock()

	// Messages (show last N that fit)
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
		// Note: Using DebugPrintAt which doesn't support color.
		// For colored text, use ebitenutil.DrawText or text/v2 package.
		ebitenutil.DebugPrintAt(screen, msg.Text, logX+10, y)
	}
}

// drawActionPanel renders the action buttons panel.
func (g *Game) drawActionPanel(screen *ebiten.Image) {
	panelY := g.screenHeight - actionPanelHeight
	panelWidth := g.screenWidth

	// Background
	drawRect(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 35, G: 35, B: 45, A: 255})
	drawRectOutline(screen, 0, panelY, panelWidth, actionPanelHeight, color.RGBA{R: 70, G: 70, B: 90, A: 255})

	// Draw direction buttons
	dirBounds := g.getDirectionButtonBounds()
	dirSymbols := map[string]string{
		"nw": "↖", "n": "↑", "ne": "↗",
		"w": "←", "e": "→",
		"sw": "↙", "s": "↓", "se": "↘",
	}

	for name, bounds := range dirBounds {
		btnColor := color.RGBA{R: 60, G: 60, B: 80, A: 255}
		if g.hoveredButton == "dir_"+name {
			btnColor = color.RGBA{R: 80, G: 80, B: 120, A: 255}
		}
		drawRect(screen, bounds.X, bounds.Y, bounds.W, bounds.H, btnColor)
		drawRectOutline(screen, bounds.X, bounds.Y, bounds.W, bounds.H, color.RGBA{R: 100, G: 100, B: 140, A: 255})
		ebitenutil.DebugPrintAt(screen, dirSymbols[name], bounds.X+8, bounds.Y+6)
	}

	// Draw action buttons
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
}

// drawConnectionStatus shows the current connection state.
func (g *Game) drawConnectionStatus(screen *ebiten.Image) {
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()

	statusText := "Disconnected"
	statusColor := color.RGBA{R: 200, G: 50, B: 50, A: 255}

	if connected {
		statusText = "Connected"
		statusColor = color.RGBA{R: 50, G: 200, B: 50, A: 255}
	}

	x := g.screenWidth - 100
	y := 5
	drawRect(screen, x-5, y-2, 95, 16, statusColor)
	ebitenutil.DebugPrintAt(screen, statusText, x, y)
}

// drawError displays error messages.
func (g *Game) drawError(screen *ebiten.Image) {
	g.mu.RLock()
	lastError := g.lastError
	errorTimeout := g.errorTimeout
	g.mu.RUnlock()

	if lastError == "" || time.Now().After(errorTimeout) {
		g.mu.Lock()
		g.lastError = ""
		g.mu.Unlock()
		return
	}

	errX := g.screenWidth/2 - 150
	errY := 50
	errWidth := 300
	errHeight := 40

	drawRect(screen, errX, errY, errWidth, errHeight, color.RGBA{R: 150, G: 30, B: 30, A: 230})
	drawRectOutline(screen, errX, errY, errWidth, errHeight, color.RGBA{R: 255, G: 100, B: 100, A: 255})
	ebitenutil.DebugPrintAt(screen, lastError, errX+10, errY+12)
}

// Layout implements ebiten.Game interface.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

// Helper methods

// addLogMessage adds a message to the combat log (thread-safe).
func (g *Game) addLogMessage(text string, msgType MessageType) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.logMessages = append(g.logMessages, LogMessage{
		Text:      text,
		Type:      msgType,
		Timestamp: time.Now().Unix(),
	})

	// Trim old messages
	if len(g.logMessages) > g.maxLogMessages {
		g.logMessages = g.logMessages[len(g.logMessages)-g.maxLogMessages:]
	}
}

// showError displays an error message temporarily (thread-safe).
func (g *Game) showError(msg string) {
	g.mu.Lock()
	g.lastError = msg
	g.errorTimeout = time.Now().Add(5 * time.Second)
	g.mu.Unlock()

	g.addLogMessage("Error: "+msg, MessageError)
}

// Drawing helpers

// drawRect draws a filled rectangle using a cached 1x1 pixel image scaled to size.
// This avoids allocating a new ebiten.Image every frame for better WASM performance.
func drawRect(screen *ebiten.Image, x, y, w, h int, c color.Color) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w), float64(h))
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	screen.DrawImage(pixelImage, op)
}

// drawRectOutline draws a rectangle outline.
func drawRectOutline(screen *ebiten.Image, x, y, w, h int, c color.Color) {
	drawLine(screen, x, y, x+w, y, c)     // Top
	drawLine(screen, x, y+h, x+w, y+h, c) // Bottom
	drawLine(screen, x, y, x, y+h, c)     // Left
	drawLine(screen, x+w, y, x+w, y+h, c) // Right
}

// drawLine draws a line between two points.
func drawLine(screen *ebiten.Image, x1, y1, x2, y2 int, c color.Color) {
	ebitenutil.DrawLine(screen, float64(x1), float64(y1), float64(x2), float64(y2), c)
}
