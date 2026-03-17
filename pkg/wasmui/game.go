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
	screenState      ScreenState
	sessionID        string
	currentAdventure *Adventure

	// Overlay state (protected by mu) — per §2 note
	overlays OverlayState

	// Character creation state (protected by mu)
	charCreation CharCreationState

	// Combat targeting state (protected by mu)
	combatAction CombatAction

	// Previous mode for overlay returns (protected by mu)
	previousMode UIMode

	// Victory/Defeat data (protected by mu)
	victoryData *VictoryData
	defeatData  *DefeatData

	// UI state (protected by mu)
	logMessages    []LogMessage
	maxLogMessages int
	selectedAction string
	hoveredButton  string
	menuIndex      int // current menu selection index

	// Inventory/spell state (protected by mu)
	inventoryItems []ItemData
	spellList      []SpellData
	questLog       *QuestLogResult
	selectedItem   int
	selectedSpell  int
	selectedQuest  int
	questLogTab    int // 0=Active, 1=Completed, 2=Failed (§7)
	spellFilter    int // -1 = all, 0-9 = level filter
	spellSearch    string

	// Guild/faction state (protected by mu)
	guildData        *GuildData
	factionRelations []FactionRelation
	guildTab         int // 0=Guild, 1=Members, 2=Factions

	// Error display (protected by mu)
	lastError    string
	errorTimeout time.Time

	// Periodic state refresh (protected by mu)
	lastRefresh     time.Time
	refreshInterval time.Duration

	// Input state (only accessed from main goroutine)
	lastInputTime time.Time
	inputCooldown time.Duration

	// Text input state (only accessed from main goroutine)
	textInputActive bool
	textInputBuffer string

	// Screen dimensions (only accessed from main goroutine)
	screenWidth  int
	screenHeight int

	// Touch input state (only accessed from main goroutine)
	touchState *TouchState

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
		screenState:     ScreenSplash,
		logMessages:     make([]LogMessage, 0),
		adventureScreen: NewAdventureScreen(),
		refreshInterval: 5 * time.Second,
		spellFilter:     -1,
		menuIndex:       0,
		touchState:      NewTouchState(),
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

// connectAndJoin handles the initial connection and game join with retry logic.
func (g *Game) connectAndJoin() {
	g.addLogMessage("Connecting to server...", MessageSystem)

	// Retry connection up to 3 times
	var connectErr error
	for attempt := 1; attempt <= 3; attempt++ {
		connectErr = g.rpcClient.Connect()
		if connectErr == nil {
			break
		}
		g.addLogMessage(fmt.Sprintf("Connection attempt %d failed: %v", attempt, connectErr), MessageWarning)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}
	if connectErr != nil {
		g.showError(fmt.Sprintf("Connection failed after retries: %v", connectErr))
		return
	}

	// Auto-join game
	result, err := g.rpcClient.JoinGame("Player1")
	if err != nil {
		g.showError(fmt.Sprintf("Failed to join game: %v", err))
		return
	}

	if !result.Success {
		g.showError("Server rejected join request")
		return
	}

	g.mu.Lock()
	g.sessionID = result.SessionID
	g.mu.Unlock()
	g.addLogMessage(fmt.Sprintf("Joined game (session: %s)", result.SessionID[:8]), MessageSystem)

	// Transition from Splash to MainMenu
	g.mu.Lock()
	g.screenState = ScreenMainMenu
	g.mu.Unlock()

	// Fetch initial game state once (no character exists yet at this point,
	// so retrying is unnecessary and causes delays that can interfere with
	// adventure loading if the user navigates quickly)
	g.refreshGameState()
}

// extractPlayerState attempts to extract player state from the top-level player data
// returned by the server's handleGetGameState. The server returns:
//
//	player: { session_id, connected, name, character: { id, name, class, level, ... position: {X, Y, Level} } }
func extractPlayerState(playerData, sessions map[string]interface{}, sessionID string) *PlayerState {
	// Primary path: use top-level player data from handleGetGameState
	if playerData != nil {
		if ps := extractFromPlayerData(playerData); ps != nil {
			return ps
		}
	}

	// Fallback: try sessions map
	if sessions != nil && sessionID != "" {
		if sessionData, ok := sessions[sessionID]; ok {
			if sessionMap, ok := sessionData.(map[string]interface{}); ok {
				if pd, ok := sessionMap["player"].(map[string]interface{}); ok {
					return extractFromFlatPlayerData(pd)
				}
			}
		}
	}

	return nil
}

// extractFromPlayerData extracts PlayerState from the server's buildPlayerStateData structure.
// Structure: { session_id, connected, name, character: { id, name, class, level, current_hp, max_hp, ... position: {X,Y,Level} } }
func extractFromPlayerData(data map[string]interface{}) *PlayerState {
	charData, ok := data["character"].(map[string]interface{})
	if !ok {
		return nil
	}

	player := &PlayerState{}

	if id, ok := charData["id"].(string); ok {
		player.ID = id
	}
	if name, ok := charData["name"].(string); ok {
		player.Name = name
	} else if name, ok := data["name"].(string); ok {
		player.Name = name
	}
	if class, ok := charData["class"].(string); ok {
		player.Class = class
	}
	player.Level = jsonInt(charData, "level")
	player.HP = jsonInt(charData, "current_hp")
	player.MaxHP = jsonInt(charData, "max_hp")
	player.Experience = jsonInt(charData, "experience")

	// Extract position
	if posData, ok := charData["position"].(map[string]interface{}); ok {
		player.Position.X = jsonInt(posData, "X")
		player.Position.Y = jsonInt(posData, "Y")
		player.Position.Level = jsonInt(posData, "Level")
	}

	// Extract attributes
	player.Attributes = PlayerAttributes{
		Strength:     jsonInt(charData, "strength"),
		Dexterity:    jsonInt(charData, "dexterity"),
		Constitution: jsonInt(charData, "constitution"),
		Intelligence: jsonInt(charData, "intelligence"),
		Wisdom:       jsonInt(charData, "wisdom"),
		Charisma:     jsonInt(charData, "charisma"),
	}

	// Only return if we got meaningful data
	if player.Name == "" && player.ID == "" {
		return nil
	}

	return player
}

// extractFromFlatPlayerData extracts PlayerState from the sessions path Player.PublicData() format.
// Structure: { name, class (int), hp, max_hp, strength, constitution }
func extractFromFlatPlayerData(data map[string]interface{}) *PlayerState {
	player := &PlayerState{}

	if name, ok := data["name"].(string); ok {
		player.Name = name
	}
	player.HP = jsonInt(data, "hp")
	player.MaxHP = jsonInt(data, "max_hp")
	player.Attributes.Strength = jsonInt(data, "strength")
	player.Attributes.Constitution = jsonInt(data, "constitution")

	if player.Name == "" {
		return nil
	}

	return player
}

// jsonInt extracts an integer from a map, handling JSON number types (float64).
func jsonInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case int64:
			return int(n)
		}
	}
	return 0
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

	if player := extractPlayerState(stateResult.Player, stateResult.Sessions, sessionID); player != nil {
		g.mu.Lock()
		g.player = player
		g.mu.Unlock()
		g.addLogMessage(fmt.Sprintf("Player loaded: %s (Level %d %s)", player.Name, player.Level, player.Class), MessageSystem)
	} else {
		g.addLogMessage("Warning: game state received but player data not found", MessageWarning)
	}

	if combat := extractCombatState(stateResult.World); combat != nil {
		g.mu.Lock()
		g.combat = combat
		g.mu.Unlock()
	}
}

// Update implements ebiten.Game interface.
func (g *Game) Update() error {
	// Update touch state for gesture detection (must be first)
	g.touchState.updateFromEbiten()

	g.mu.RLock()
	mode := g.mode
	screen := g.screenState
	overlays := g.overlays
	g.mu.RUnlock()

	// Handle overlays first — they intercept input when shown
	if overlays.ShowSettings {
		g.updateSettingsOverlay()
		return nil
	}
	if overlays.ShowQuestLog {
		g.updateQuestLogOverlay()
		return nil
	}
	if overlays.ShowGuildPanel {
		g.updateGuildPanelOverlay()
		return nil
	}

	// Route by mode
	switch mode {
	case ModeAdventureSelect:
		g.adventureScreen.Update(g)
	case ModeCharacterCreation:
		g.updateCharacterCreation()
	case ModeCombat:
		g.updateCombat()
	case ModeInventory:
		g.updateInventory()
	case ModeSpellcasting:
		g.updateSpellbook()
	case ModeNormal:
		switch screen {
		case ScreenSplash:
			g.updateSplash()
		case ScreenMainMenu:
			g.updateMainMenu()
		case ScreenVictory:
			g.updateVictory()
		case ScreenDefeat:
			g.updateDefeat()
		case ScreenExploration:
			g.updateExploration()
		}
	}

	// Periodic state refresh (only when in exploration or combat)
	if mode == ModeNormal && screen == ScreenExploration || mode == ModeCombat {
		g.mu.RLock()
		connected := g.connected
		lastRefresh := g.lastRefresh
		refreshInterval := g.refreshInterval
		g.mu.RUnlock()

		if connected && time.Since(lastRefresh) > refreshInterval {
			g.mu.Lock()
			g.lastRefresh = time.Now()
			g.mu.Unlock()
			go g.refreshGameState()
		}
	}

	return nil
}

// handleMouseInput processes mouse and touch events.
func (g *Game) handleMouseInput() {
	x, y := ebiten.CursorPosition()

	// Check button hover states
	g.hoveredButton = g.getButtonAtPosition(x, y)

	// Handle mouse clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.handleClick(x, y)
	}

	// Handle touch taps as clicks
	g.handleTouchInput()
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
	g.mu.RLock()
	mode := g.mode
	screenState := g.screenState
	overlays := g.overlays
	g.mu.RUnlock()

	// Route drawing by mode
	switch mode {
	case ModeAdventureSelect:
		g.adventureScreen.Draw(screen, g)
	case ModeCharacterCreation:
		g.drawCharacterCreation(screen)
	case ModeCombat:
		g.drawCombatScreen(screen)
	case ModeInventory:
		g.drawInventoryScreen(screen)
	case ModeSpellcasting:
		g.drawSpellbookScreen(screen)
	case ModeNormal:
		switch screenState {
		case ScreenSplash:
			g.drawSplash(screen)
		case ScreenMainMenu:
			g.drawMainMenu(screen)
		case ScreenVictory:
			g.drawVictory(screen)
		case ScreenDefeat:
			g.drawDefeat(screen)
		case ScreenExploration:
			g.drawExplorationScreen(screen)
		}
	}

	// Draw overlays on top
	if overlays.ShowQuestLog {
		g.drawQuestLogOverlay(screen)
	}
	if overlays.ShowGuildPanel {
		g.drawGuildPanelOverlay(screen)
	}
	if overlays.ShowSettings {
		g.drawSettingsOverlay(screen)
	}

	// Always draw error overlay and connection status
	g.drawError(screen)
	g.drawConnectionStatus(screen)
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
// Returns fixed 800×600 logical canvas per §1 spec.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
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
