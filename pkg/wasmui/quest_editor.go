//go:build js && wasm

// Package wasmui provides WebAssembly UI components for the GoldBox RPG Engine.
package wasmui

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	questEditorWidth  = 1024
	questEditorHeight = 768
	nodeWidth         = 180
	nodeHeight        = 80
	nodePadding       = 20
	objectiveHeight   = 24
)

// QuestNode represents a draggable quest objective in the visual editor.
type QuestNode struct {
	ID          string
	Description string
	Required    int
	X           int
	Y           int
	Connections []string
}

// QuestEditorState holds the current state of the quest being edited.
type QuestEditorState struct {
	QuestID     string
	Title       string
	Description string
	Objectives  []*QuestNode
	Rewards     []QuestRewardEntry
}

// QuestRewardEntry represents a reward in the quest editor.
type QuestRewardEntry struct {
	Type   string
	Value  int
	ItemID string
}

// QuestEditorGame implements ebiten.Game for the visual quest chain builder.
type QuestEditorGame struct {
	mu sync.RWMutex

	state          *QuestEditorState
	selectedNode   int
	dragging       bool
	dragOffsetX    int
	dragOffsetY    int
	inputMode      QuestInputMode
	inputBuffer    string
	inputTarget    int
	statusMessage  string
	statusTimeout  time.Time
	cursorX        int
	cursorY        int
	scrollY        int
	screenWidth    int
	screenHeight   int
	dirty          bool
	connectingFrom int
}

// QuestInputMode represents the current input mode in the quest editor.
type QuestInputMode int

const (
	// QuestModeNormal is the default mode for node manipulation.
	QuestModeNormal QuestInputMode = iota
	// QuestModeEditTitle is for editing the quest title.
	QuestModeEditTitle
	// QuestModeEditDescription is for editing node descriptions.
	QuestModeEditDescription
	// QuestModeConnect is for creating connections between nodes.
	QuestModeConnect
)

// NewQuestEditorGame creates a new quest editor instance.
func NewQuestEditorGame() *QuestEditorGame {
	return &QuestEditorGame{
		state: &QuestEditorState{
			Title:       "New Quest",
			Description: "Quest description...",
			Objectives:  make([]*QuestNode, 0),
			Rewards:     make([]QuestRewardEntry, 0),
		},
		selectedNode:   -1,
		inputMode:      QuestModeNormal,
		connectingFrom: -1,
		screenWidth:    questEditorWidth,
		screenHeight:   questEditorHeight,
	}
}

// Update implements ebiten.Game for the quest editor.
func (g *QuestEditorGame) Update() error {
	g.handleKeyboardInput()
	g.handleMouseInput()
	return nil
}

// handleKeyboardInput processes keyboard input for the quest editor.
func (g *QuestEditorGame) handleKeyboardInput() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.inputMode != QuestModeNormal {
		g.handleTextInput()
		return
	}

	g.handleShortcutKeys()
	g.handleScrollKeys()
}

// handleShortcutKeys processes single-key shortcuts and Ctrl combinations.
func (g *QuestEditorGame) handleShortcutKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.addNewObjective()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.deleteSelectedNode()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.startConnecting()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.inputMode = QuestModeEditTitle
		g.inputBuffer = g.state.Title
		g.setStatusLocked("Editing title (Enter to confirm, Esc to cancel)")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.addReward()
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.saveQuest()
	}
}

// handleScrollKeys adjusts the scroll position based on arrow key input.
func (g *QuestEditorGame) handleScrollKeys() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.scrollY = max(0, g.scrollY-5)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.scrollY += 5
	}
}

// handleTextInput processes text input when in edit mode.
func (g *QuestEditorGame) handleTextInput() {
	// Handle escape to cancel
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.inputMode = QuestModeNormal
		g.inputBuffer = ""
		g.setStatusLocked("Edit cancelled")
		return
	}

	// Handle enter to confirm
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		switch g.inputMode {
		case QuestModeEditTitle:
			g.state.Title = g.inputBuffer
			g.dirty = true
		case QuestModeEditDescription:
			if g.inputTarget >= 0 && g.inputTarget < len(g.state.Objectives) {
				g.state.Objectives[g.inputTarget].Description = g.inputBuffer
				g.dirty = true
			}
		}
		g.inputMode = QuestModeNormal
		g.inputBuffer = ""
		g.setStatusLocked("Saved")
		return
	}

	// Handle backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.inputBuffer) > 0 {
		g.inputBuffer = g.inputBuffer[:len(g.inputBuffer)-1]
		return
	}

	// Handle character input
	for _, char := range ebiten.AppendInputChars(nil) {
		if len(g.inputBuffer) < 100 {
			g.inputBuffer += string(char)
		}
	}
}

// handleMouseInput processes mouse events for the quest editor.
func (g *QuestEditorGame) handleMouseInput() {
	mx, my := ebiten.CursorPosition()
	g.mu.Lock()
	g.cursorX = mx
	g.cursorY = my
	g.mu.Unlock()

	// Handle node dragging
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.handleDrag(mx, my)
	} else {
		g.mu.Lock()
		g.dragging = false
		g.mu.Unlock()
	}

	// Handle node selection on click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.handleNodeClick(mx, my)
	}

	// Handle right-click for context actions
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.handleRightClick(mx, my)
	}
}

// handleNodeClick selects a node or deselects if clicking empty space.
func (g *QuestEditorGame) handleNodeClick(mx, my int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	adjustedY := my + g.scrollY - toolbarHeight

	// Check for connection mode
	if g.inputMode == QuestModeConnect {
		nodeIdx := g.findNodeAt(mx, adjustedY)
		if nodeIdx >= 0 && nodeIdx != g.connectingFrom {
			g.createConnection(g.connectingFrom, nodeIdx)
		}
		g.inputMode = QuestModeNormal
		g.connectingFrom = -1
		return
	}

	nodeIdx := g.findNodeAt(mx, adjustedY)
	if nodeIdx >= 0 {
		g.selectedNode = nodeIdx
		g.dragging = true
		node := g.state.Objectives[nodeIdx]
		g.dragOffsetX = mx - node.X
		g.dragOffsetY = adjustedY - node.Y
		g.setStatusLocked(fmt.Sprintf("Selected: %s", node.Description))
	} else {
		g.selectedNode = -1
	}
}

// handleDrag updates node position during drag.
func (g *QuestEditorGame) handleDrag(mx, my int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.dragging || g.selectedNode < 0 {
		return
	}

	adjustedY := my + g.scrollY - toolbarHeight
	node := g.state.Objectives[g.selectedNode]
	node.X = mx - g.dragOffsetX
	node.Y = adjustedY - g.dragOffsetY
	g.dirty = true
}

// handleRightClick handles right-click context actions.
func (g *QuestEditorGame) handleRightClick(mx, my int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	adjustedY := my + g.scrollY - toolbarHeight
	nodeIdx := g.findNodeAt(mx, adjustedY)

	if nodeIdx >= 0 {
		g.selectedNode = nodeIdx
		g.inputMode = QuestModeEditDescription
		g.inputTarget = nodeIdx
		g.inputBuffer = g.state.Objectives[nodeIdx].Description
		g.setStatusLocked("Editing description (Enter to confirm)")
	}
}

// findNodeAt returns the index of the node at position (x, y), or -1.
func (g *QuestEditorGame) findNodeAt(x, y int) int {
	for i, node := range g.state.Objectives {
		if x >= node.X && x <= node.X+nodeWidth &&
			y >= node.Y && y <= node.Y+nodeHeight {
			return i
		}
	}
	return -1
}

// addNewObjective adds a new objective node to the quest.
func (g *QuestEditorGame) addNewObjective() {
	id := fmt.Sprintf("obj_%d", len(g.state.Objectives)+1)
	x := nodePadding + (len(g.state.Objectives)%3)*(nodeWidth+nodePadding)
	y := nodePadding + (len(g.state.Objectives)/3)*(nodeHeight+nodePadding*2)

	node := &QuestNode{
		ID:          id,
		Description: "New objective",
		Required:    1,
		X:           x,
		Y:           y,
		Connections: make([]string, 0),
	}

	g.state.Objectives = append(g.state.Objectives, node)
	g.selectedNode = len(g.state.Objectives) - 1
	g.dirty = true
	g.setStatusLocked("Added new objective (right-click to edit)")
}

// deleteSelectedNode removes the currently selected node.
func (g *QuestEditorGame) deleteSelectedNode() {
	if g.selectedNode < 0 || g.selectedNode >= len(g.state.Objectives) {
		return
	}

	deletedID := g.state.Objectives[g.selectedNode].ID

	// Remove connections to this node
	for _, node := range g.state.Objectives {
		newConns := make([]string, 0)
		for _, conn := range node.Connections {
			if conn != deletedID {
				newConns = append(newConns, conn)
			}
		}
		node.Connections = newConns
	}

	// Remove the node
	g.state.Objectives = append(
		g.state.Objectives[:g.selectedNode],
		g.state.Objectives[g.selectedNode+1:]...,
	)

	g.selectedNode = -1
	g.dirty = true
	g.setStatusLocked("Objective deleted")
}

// startConnecting enters connection mode to link two nodes.
func (g *QuestEditorGame) startConnecting() {
	if g.selectedNode < 0 {
		g.setStatusLocked("Select a node first, then press C to connect")
		return
	}
	g.inputMode = QuestModeConnect
	g.connectingFrom = g.selectedNode
	g.setStatusLocked("Click target node to connect")
}

// createConnection links two nodes together.
func (g *QuestEditorGame) createConnection(fromIdx, toIdx int) {
	if fromIdx < 0 || fromIdx >= len(g.state.Objectives) {
		return
	}
	if toIdx < 0 || toIdx >= len(g.state.Objectives) {
		return
	}

	fromNode := g.state.Objectives[fromIdx]
	toNode := g.state.Objectives[toIdx]

	// Check for existing connection
	for _, conn := range fromNode.Connections {
		if conn == toNode.ID {
			g.setStatusLocked("Connection already exists")
			return
		}
	}

	fromNode.Connections = append(fromNode.Connections, toNode.ID)
	g.dirty = true
	g.setStatusLocked(fmt.Sprintf("Connected %s → %s", fromNode.ID, toNode.ID))
}

// addReward adds a new reward to the quest.
func (g *QuestEditorGame) addReward() {
	reward := QuestRewardEntry{
		Type:  "gold",
		Value: 100,
	}
	g.state.Rewards = append(g.state.Rewards, reward)
	g.dirty = true
	g.setStatusLocked("Added reward: 100 gold")
}

// saveQuest saves the quest (placeholder for WebSocket integration).
func (g *QuestEditorGame) saveQuest() {
	g.dirty = false
	g.setStatusLocked(fmt.Sprintf("Saved: %s (%d objectives)", g.state.Title, len(g.state.Objectives)))
}

// setStatusLocked sets status message (caller must hold lock).
func (g *QuestEditorGame) setStatusLocked(msg string) {
	g.statusMessage = msg
	g.statusTimeout = time.Now().Add(3 * time.Second)
}

// Draw implements ebiten.Game for the quest editor.
func (g *QuestEditorGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 25, G: 30, B: 40, A: 255})

	g.drawQuestHeader(screen)
	g.drawConnections(screen)
	g.drawNodes(screen)
	g.drawRewards(screen)
	g.drawStatusBar(screen)
	g.drawHelpPanel(screen)
}

// drawQuestHeader draws the quest title and description area.
func (g *QuestEditorGame) drawQuestHeader(screen *ebiten.Image) {
	g.mu.RLock()
	title := g.state.Title
	description := g.state.Description
	inputMode := g.inputMode
	inputBuffer := g.inputBuffer
	dirty := g.dirty
	g.mu.RUnlock()

	// Header background
	drawFilledRect(screen, 0, 0, g.screenWidth, toolbarHeight,
		color.RGBA{R: 45, G: 50, B: 65, A: 255})

	// Title
	displayTitle := title
	if dirty {
		displayTitle += " *"
	}
	if inputMode == QuestModeEditTitle {
		displayTitle = "> " + inputBuffer + "_"
	}
	ebitenutil.DebugPrintAt(screen, "Quest: "+displayTitle, 10, 8)

	// Description hint
	descHint := description
	if len(descHint) > 60 {
		descHint = descHint[:57] + "..."
	}
	ebitenutil.DebugPrintAt(screen, descHint, 10, 22)
}

// drawConnections draws lines between connected nodes.
func (g *QuestEditorGame) drawConnections(screen *ebiten.Image) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodeMap := make(map[string]*QuestNode)
	for _, node := range g.state.Objectives {
		nodeMap[node.ID] = node
	}

	for _, node := range g.state.Objectives {
		fromX := node.X + nodeWidth/2
		fromY := node.Y + nodeHeight - g.scrollY + toolbarHeight

		for _, connID := range node.Connections {
			target, ok := nodeMap[connID]
			if !ok {
				continue
			}
			toX := target.X + nodeWidth/2
			toY := target.Y - g.scrollY + toolbarHeight

			// Draw line (simple vertical/horizontal segments)
			midY := (fromY + toY) / 2
			drawFilledRect(screen, fromX-1, fromY, 2, midY-fromY,
				color.RGBA{R: 100, G: 180, B: 255, A: 200})
			drawFilledRect(screen, min(fromX, toX), midY-1, abs(toX-fromX), 2,
				color.RGBA{R: 100, G: 180, B: 255, A: 200})
			drawFilledRect(screen, toX-1, midY, 2, toY-midY,
				color.RGBA{R: 100, G: 180, B: 255, A: 200})
		}
	}

	// Draw connection preview line if in connect mode
	if g.inputMode == QuestModeConnect && g.connectingFrom >= 0 {
		fromNode := g.state.Objectives[g.connectingFrom]
		fromX := fromNode.X + nodeWidth/2
		fromY := fromNode.Y + nodeHeight/2 - g.scrollY + toolbarHeight
		toX := g.cursorX
		toY := g.cursorY

		drawFilledRect(screen, min(fromX, toX), min(fromY, toY),
			abs(toX-fromX)+2, abs(toY-fromY)+2,
			color.RGBA{R: 255, G: 200, B: 100, A: 150})
	}
}

// drawNodes draws all objective nodes.
func (g *QuestEditorGame) drawNodes(screen *ebiten.Image) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for i, node := range g.state.Objectives {
		x := node.X
		y := node.Y - g.scrollY + toolbarHeight

		// Skip off-screen nodes
		if y+nodeHeight < toolbarHeight || y > g.screenHeight-statusBarHeight {
			continue
		}

		// Node background
		bgColor := color.RGBA{R: 55, G: 65, B: 85, A: 255}
		if i == g.selectedNode {
			bgColor = color.RGBA{R: 70, G: 100, B: 150, A: 255}
		}
		drawFilledRect(screen, x, y, nodeWidth, nodeHeight, bgColor)

		// Node border
		borderColor := color.RGBA{R: 80, G: 90, B: 110, A: 255}
		if i == g.selectedNode {
			borderColor = color.RGBA{R: 100, G: 150, B: 220, A: 255}
		}
		drawFilledRect(screen, x, y, nodeWidth, 2, borderColor)
		drawFilledRect(screen, x, y+nodeHeight-2, nodeWidth, 2, borderColor)
		drawFilledRect(screen, x, y, 2, nodeHeight, borderColor)
		drawFilledRect(screen, x+nodeWidth-2, y, 2, nodeHeight, borderColor)

		// Node ID
		ebitenutil.DebugPrintAt(screen, node.ID, x+5, y+5)

		// Node description (truncated)
		desc := node.Description
		if len(desc) > 20 {
			desc = desc[:17] + "..."
		}
		ebitenutil.DebugPrintAt(screen, desc, x+5, y+25)

		// Required count
		reqText := fmt.Sprintf("Required: %d", node.Required)
		ebitenutil.DebugPrintAt(screen, reqText, x+5, y+45)

		// Connection count
		connText := fmt.Sprintf("→ %d", len(node.Connections))
		ebitenutil.DebugPrintAt(screen, connText, x+nodeWidth-30, y+5)
	}
}

// drawRewards draws the rewards panel.
func (g *QuestEditorGame) drawRewards(screen *ebiten.Image) {
	g.mu.RLock()
	rewards := g.state.Rewards
	g.mu.RUnlock()

	x := g.screenWidth - 200
	y := toolbarHeight + 10

	// Rewards header
	ebitenutil.DebugPrintAt(screen, "Rewards (R to add):", x, y)
	y += 20

	for i, reward := range rewards {
		text := fmt.Sprintf("%d. %s: %d", i+1, reward.Type, reward.Value)
		if reward.ItemID != "" {
			text += " (" + reward.ItemID + ")"
		}
		ebitenutil.DebugPrintAt(screen, text, x, y)
		y += 16
	}
}

// drawStatusBar draws the bottom status bar.
func (g *QuestEditorGame) drawStatusBar(screen *ebiten.Image) {
	y := g.screenHeight - statusBarHeight
	drawFilledRect(screen, 0, y, g.screenWidth, statusBarHeight,
		color.RGBA{R: 35, G: 40, B: 55, A: 255})

	g.mu.RLock()
	statusMsg := g.statusMessage
	statusTimeout := g.statusTimeout
	objCount := len(g.state.Objectives)
	g.mu.RUnlock()

	// Objective count
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Objectives: %d", objCount), 10, y+5)

	// Status message
	if time.Now().Before(statusTimeout) {
		ebitenutil.DebugPrintAt(screen, statusMsg, 150, y+5)
	}
}

// drawHelpPanel draws keyboard shortcuts help.
func (g *QuestEditorGame) drawHelpPanel(screen *ebiten.Image) {
	x := g.screenWidth - 200
	y := g.screenHeight - 150

	help := []string{
		"[N] New objective",
		"[Del] Delete selected",
		"[C] Connect nodes",
		"[T] Edit title",
		"[R] Add reward",
		"[Ctrl+S] Save",
		"Right-click: Edit",
	}

	for i, line := range help {
		ebitenutil.DebugPrintAt(screen, line, x, y+i*14)
	}
}

// Layout implements ebiten.Game.
func (g *QuestEditorGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.mu.Lock()
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	g.mu.Unlock()
	return outsideWidth, outsideHeight
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
