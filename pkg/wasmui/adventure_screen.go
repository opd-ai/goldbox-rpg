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

	// Exit adventure screen with Escape → return to MainMenu
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.mode = ModeNormal
		g.screenState = ScreenMainMenu
		g.mu.Unlock()
	}

	// Mouse click on list items (§3.3)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		listTop := 45
		listBottom := ScreenHeight - 60
		if x >= 10 && x <= 390 && y >= listTop && y <= listBottom {
			idx := (y - listTop) / 30
			if idx >= 0 && idx < len(s.adventures) {
				s.selectedIndex = idx
			}
		}
	}

	// Touch tap on list items
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		listTop := 45
		listBottom := ScreenHeight - 60
		if tx >= 10 && tx <= 390 && ty >= listTop && ty <= listBottom {
			idx := (ty - listTop) / 30
			if idx >= 0 && idx < len(s.adventures) {
				s.selectedIndex = idx
			}
		}
	}

	// Touch swipe for list navigation
	if swiped, dir := g.touchState.HasSwipe(); swiped {
		switch dir {
		case GestureSwipeUp:
			if s.selectedIndex > 0 {
				s.selectedIndex--
			}
		case GestureSwipeDown:
			if s.selectedIndex < len(s.adventures)-1 {
				s.selectedIndex++
			}
		}
	}
}

// Draw renders the adventure selection screen with list + detail panels (§3.3).
func (s *AdventureScreen) Draw(screen *ebiten.Image, g *Game) {
	screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Header bar
	drawRect(screen, 0, 0, ScreenWidth, 35, color.RGBA{R: 35, G: 35, B: 50, A: 255})
	ebitenutil.DebugPrintAt(screen, "ADVENTURE SELECT", 20, 10)
	ebitenutil.DebugPrintAt(screen, "[ESC]", ScreenWidth-60, 10)

	if s.loading {
		ebitenutil.DebugPrintAt(screen, "Loading adventures...", 340, 300)
		return
	}

	s.drawErrorMessage(screen)

	if len(s.adventures) == 0 {
		ebitenutil.DebugPrintAt(screen, "No adventures available.", 310, 300)
		ebitenutil.DebugPrintAt(screen, "Place adventure packs in data/adventures/", 250, 330)
		return
	}

	// Divider line between panels
	drawLine(screen, 400, 40, 400, ScreenHeight-60, color.RGBA{R: 80, G: 80, B: 100, A: 255})

	// Left panel — adventure list (0-400 × 40-520)
	listTop := 45
	for i, adv := range s.adventures {
		y := listTop + i*30
		if y > ScreenHeight-80 {
			break
		}
		bgColor := color.RGBA{R: 30, G: 30, B: 45, A: 255}
		if i == s.selectedIndex {
			bgColor = color.RGBA{R: 60, G: 50, B: 80, A: 255}
		}
		drawRect(screen, 10, y, 380, 26, bgColor)

		marker := "  "
		if i == s.selectedIndex {
			marker = "> "
		}
		ebitenutil.DebugPrintAt(screen, marker+truncateText(adv.Title, 40), 15, y+5)
	}

	// Right panel — detail (400-800 × 40-520)
	if s.selectedIndex < len(s.adventures) {
		adv := s.adventures[s.selectedIndex]
		dx := 415
		dy := 50
		ebitenutil.DebugPrintAt(screen, "Title: "+adv.Title, dx, dy)
		dy += 20
		ebitenutil.DebugPrintAt(screen, "Theme: "+adv.Theme, dx, dy)
		dy += 20
		ebitenutil.DebugPrintAt(screen, "Level: "+itoa(adv.MinLevel)+"-"+itoa(adv.MaxLevel), dx, dy)
		dy += 20
		ebitenutil.DebugPrintAt(screen, "Maps:  "+itoa(adv.MapCount), dx, dy)
		dy += 20
		ebitenutil.DebugPrintAt(screen, "Quests: "+itoa(adv.QuestCount), dx, dy)
		dy += 20
		ebitenutil.DebugPrintAt(screen, "Est:   "+adv.EstHours+" hrs", dx, dy)
		dy += 30
		ebitenutil.DebugPrintAt(screen, "Description:", dx, dy)
		dy += 18
		// Wrap description text into lines
		desc := adv.Description
		for len(desc) > 0 {
			lineLen := 42
			if len(desc) < lineLen {
				lineLen = len(desc)
			}
			ebitenutil.DebugPrintAt(screen, desc[:lineLen], dx, dy)
			desc = desc[lineLen:]
			dy += 15
			if dy > ScreenHeight-80 {
				break
			}
		}
	}

	// Footer
	drawRect(screen, 0, ScreenHeight-50, ScreenWidth, 50, color.RGBA{R: 35, G: 35, B: 50, A: 255})
	ebitenutil.DebugPrintAt(screen, "[Enter] Load   [R] Refresh   [Esc] Back", 240, ScreenHeight-35)
}

// drawErrorMessage draws any active error message.
func (s *AdventureScreen) drawErrorMessage(screen *ebiten.Image) {
	if s.errorMsg != "" && time.Now().Before(s.errorExpiry) {
		ebitenutil.DebugPrintAt(screen, s.errorMsg, 250, 100)
	}
}

// truncateText truncates text to maxLen characters with ellipsis.
func truncateText(text string, maxLen int) string {
	if len(text) > maxLen {
		return text[:maxLen-3] + "..."
	}
	return text
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

		// Adventure loaded successfully - transition to CharacterCreation
		g.mu.Lock()
		g.currentAdventure = adventure
		g.mode = ModeCharacterCreation
		g.charCreation = CharCreationState{Step: CharStepName}
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

// Error implements the error interface for simpleError.
func (e *simpleError) Error() string {
	return e.msg
}
