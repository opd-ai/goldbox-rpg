//go:build js && wasm

package wasmui

import (
	"encoding/json"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// AdventureSummary represents adventure data from the server.
type AdventureSummary struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Theme       string `json:"theme"`
	MinLevel    int    `json:"min_level"`
	MaxLevel    int    `json:"max_level"`
	EstHours    string `json:"est_hours"`
	MapCount    int    `json:"map_count"`
	QuestCount  int    `json:"quest_count"`
}

// AdventureScreen handles adventure selection UI.
type AdventureScreen struct {
	adventures      []AdventureSummary
	selectedIndex   int
	loading         bool
	errorMsg        string
	errorExpiry     time.Time
	lastRefresh     time.Time
	refreshCooldown time.Duration
}

// NewAdventureScreen creates a new adventure selection screen.
func NewAdventureScreen() *AdventureScreen {
	return &AdventureScreen{
		adventures:      make([]AdventureSummary, 0),
		selectedIndex:   0,
		refreshCooldown: 2 * time.Second,
	}
}

// Update handles adventure screen input.
func (s *AdventureScreen) Update(g *Game) {
	// Handle up/down navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		if s.selectedIndex > 0 {
			s.selectedIndex--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		if s.selectedIndex < len(s.adventures)-1 {
			s.selectedIndex++
		}
	}

	// Handle selection with Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(s.adventures) > 0 {
		s.loadSelectedAdventure(g)
	}

	// Refresh list with R key
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if time.Since(s.lastRefresh) > s.refreshCooldown {
			s.RefreshAdventures(g)
		}
	}

	// Exit adventure screen with Escape
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.mode = ModeNormal
		g.mu.Unlock()
	}
}

// Draw renders the adventure selection screen.
func (s *AdventureScreen) Draw(screen *ebiten.Image, g *Game) {
	// Background
	screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Title
	ebitenutil.DebugPrintAt(screen, "=== ADVENTURE SELECTION ===", 280, 30)

	// Instructions
	ebitenutil.DebugPrintAt(screen, "UP/DOWN: Navigate | ENTER: Select | R: Refresh | ESC: Back", 180, 55)

	if s.loading {
		ebitenutil.DebugPrintAt(screen, "Loading adventures...", 340, 300)
		return
	}

	if s.errorMsg != "" && time.Now().Before(s.errorExpiry) {
		ebitenutil.DebugPrintAt(screen, s.errorMsg, 250, 100)
	}

	if len(s.adventures) == 0 {
		ebitenutil.DebugPrintAt(screen, "No adventures available.", 310, 300)
		ebitenutil.DebugPrintAt(screen, "Place adventure packs in data/adventures/", 250, 330)
		return
	}

	// Draw adventure list
	startY := 100
	for i, adv := range s.adventures {
		y := startY + (i * 60)
		if y > ScreenHeight-100 {
			break // Don't draw off screen
		}

		// Selection highlight
		if i == s.selectedIndex {
			drawRect(screen, 30, y-5, ScreenWidth-60, 55, color.RGBA{R: 60, G: 60, B: 100, A: 200})
			ebitenutil.DebugPrintAt(screen, ">", 20, y+10)
		}

		// Adventure info
		titleText := adv.Title
		if len(titleText) > 50 {
			titleText = titleText[:47] + "..."
		}
		ebitenutil.DebugPrintAt(screen, titleText, 40, y)

		// Level range and theme
		levelText := formatAdventureDetails(adv)
		ebitenutil.DebugPrintAt(screen, levelText, 60, y+15)

		// Description (truncated)
		desc := adv.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		ebitenutil.DebugPrintAt(screen, desc, 60, y+30)
	}

	// Footer with selected adventure details
	if s.selectedIndex < len(s.adventures) {
		adv := s.adventures[s.selectedIndex]
		footerY := ScreenHeight - 50
		mapQuestText := formatFooterText(adv)
		ebitenutil.DebugPrintAt(screen, mapQuestText, 50, footerY)
	}
}

// formatAdventureDetails formats level range and theme info.
func formatAdventureDetails(adv AdventureSummary) string {
	return "Levels " + itoa(adv.MinLevel) + "-" + itoa(adv.MaxLevel) + " | " + adv.Theme + " | " + adv.EstHours + " hours"
}

// formatFooterText formats map and quest count info.
func formatFooterText(adv AdventureSummary) string {
	return "Maps: " + itoa(adv.MapCount) + " | Quests: " + itoa(adv.QuestCount) + " | Slug: " + adv.Slug
}

// itoa converts int to string without strconv import (WASM-friendly).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// RefreshAdventures fetches the adventure list from the server.
func (s *AdventureScreen) RefreshAdventures(g *Game) {
	s.loading = true
	s.lastRefresh = time.Now()

	go func() {
		adventures, err := g.rpcClient.ListAdventures()
		if err != nil {
			s.errorMsg = "Failed to load adventures: " + err.Error()
			s.errorExpiry = time.Now().Add(5 * time.Second)
			s.loading = false
			return
		}

		s.adventures = adventures
		s.loading = false
		if s.selectedIndex >= len(s.adventures) {
			s.selectedIndex = 0
		}
	}()
}

// loadSelectedAdventure loads the currently selected adventure.
func (s *AdventureScreen) loadSelectedAdventure(g *Game) {
	if s.selectedIndex >= len(s.adventures) {
		return
	}

	adv := s.adventures[s.selectedIndex]
	s.loading = true

	go func() {
		adventure, err := g.rpcClient.LoadAdventure(adv.Slug)
		if err != nil {
			s.errorMsg = "Failed to load adventure: " + err.Error()
			s.errorExpiry = time.Now().Add(5 * time.Second)
			s.loading = false
			return
		}

		// Adventure loaded successfully - store in game state
		g.mu.Lock()
		g.currentAdventure = adventure
		g.mode = ModeNormal
		g.mu.Unlock()

		g.addLogMessage("Loaded adventure: "+adventure.Title, MessageSystem)
		s.loading = false
	}()
}

// ListAdventures calls the adventure.list RPC method.
func (c *RPCClient) ListAdventures() ([]AdventureSummary, error) {
	result, err := c.Call("adventure.list", nil)
	if err != nil {
		return nil, err
	}

	// Parse response
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, errorf("unexpected response format")
	}

	adventuresRaw, ok := resultMap["adventures"]
	if !ok {
		return []AdventureSummary{}, nil
	}

	// Re-marshal and unmarshal to get proper types
	adventuresJSON, err := json.Marshal(adventuresRaw)
	if err != nil {
		return nil, err
	}

	var adventures []AdventureSummary
	if err := json.Unmarshal(adventuresJSON, &adventures); err != nil {
		return nil, err
	}

	return adventures, nil
}

// Adventure represents the full adventure data from server.
type Adventure struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Theme       string `json:"theme"`
	MinLevel    int    `json:"min_level"`
	MaxLevel    int    `json:"max_level"`
	EstHours    string `json:"est_hours"`
}

// LoadAdventure calls the adventure.load RPC method.
func (c *RPCClient) LoadAdventure(slug string) (*Adventure, error) {
	result, err := c.Call("adventure.load", map[string]interface{}{"slug": slug})
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, errorf("unexpected response format")
	}

	adventureRaw, ok := resultMap["adventure"]
	if !ok {
		return nil, errorf("adventure not found in response")
	}

	// Re-marshal and unmarshal
	adventureJSON, err := json.Marshal(adventureRaw)
	if err != nil {
		return nil, err
	}

	var adventure Adventure
	if err := json.Unmarshal(adventureJSON, &adventure); err != nil {
		return nil, err
	}

	return &adventure, nil
}

// errorf creates a simple error.
func errorf(format string, args ...interface{}) error {
	return &simpleError{msg: format}
}

type simpleError struct {
	msg string
}

func (e *simpleError) Error() string {
	return e.msg
}
