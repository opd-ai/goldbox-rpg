//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// updateCharacterCreation handles input for the character creation flow (§4).
func (g *Game) updateCharacterCreation() {
	g.mu.RLock()
	step := g.charCreation.Step
	g.mu.RUnlock()

	switch step {
	case CharStepName:
		g.updateCharCreationName()
	case CharStepClass:
		g.updateCharCreationClass()
	case CharStepAttributes:
		g.updateCharCreationAttributes()
	case CharStepReview:
		g.updateCharCreationReview()
	}
}

// drawCharacterCreation renders the character creation screen (§4).
func (g *Game) drawCharacterCreation(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 25, G: 25, B: 40, A: 255})

	ebitenutil.DebugPrintAt(screen, "CHARACTER CREATION", 300, 20)

	g.mu.RLock()
	step := g.charCreation.Step
	g.mu.RUnlock()

	steps := []string{"Name", "Class", "Attributes", "Review"}
	for i, s := range steps {
		x := 150 + i*150
		marker := "  "
		if CharCreationStep(i) == step {
			marker = "> "
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%d. %s", marker, i+1, s), x, 50)
	}

	drawLine(screen, 50, 70, 750, 70, color.RGBA{R: 80, G: 80, B: 120, A: 255})

	switch step {
	case CharStepName:
		g.drawCharCreationName(screen)
	case CharStepClass:
		g.drawCharCreationClass(screen)
	case CharStepAttributes:
		g.drawCharCreationAttributes(screen)
	case CharStepReview:
		g.drawCharCreationReview(screen)
	}

	ebitenutil.DebugPrintAt(screen, "Esc: Cancel  |  Enter: Next  |  Backspace: Back", 220, 570)
}

// --- Step 1: Name ---

func (g *Game) updateCharCreationName() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.mode = ModeNormal
		g.screenState = ScreenMainMenu
		g.charCreation = CharCreationState{}
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.mu.RLock()
		name := g.charCreation.Name
		g.mu.RUnlock()
		if name != "" {
			g.mu.Lock()
			g.charCreation.Step = CharStepClass
			g.mu.Unlock()
		}
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.mu.Lock()
		if len(g.charCreation.Name) > 0 {
			g.charCreation.Name = g.charCreation.Name[:len(g.charCreation.Name)-1]
		}
		g.mu.Unlock()
		return
	}

	runes := ebiten.AppendInputChars(nil)
	if len(runes) > 0 {
		g.mu.Lock()
		for _, r := range runes {
			if len(g.charCreation.Name) < 20 {
				g.charCreation.Name += string(r)
			}
		}
		g.mu.Unlock()
	}
}

func (g *Game) drawCharCreationName(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "Enter your character's name:", 250, 120)

	g.mu.RLock()
	name := g.charCreation.Name
	g.mu.RUnlock()

	boxX, boxY := 250, 150
	boxW, boxH := 300, 30
	drawRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 50, G: 50, B: 70, A: 255})
	drawRectOutline(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 120, G: 120, B: 180, A: 255})

	ebitenutil.DebugPrintAt(screen, name+"_", boxX+10, boxY+8)
	ebitenutil.DebugPrintAt(screen, "(1-20 characters)", 280, 190)
}

// --- Step 2: Class ---

func (g *Game) updateCharCreationClass() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.mode = ModeNormal
		g.screenState = ScreenMainMenu
		g.charCreation = CharCreationState{}
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.mu.Lock()
		g.charCreation.Step = CharStepName
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.mu.Lock()
		g.charCreation.Step = CharStepAttributes
		if g.charCreation.AttrMethod == 0 {
			g.charCreation.AttrMethod = AttrMethodStandard
			g.charCreation.SetStandardArray()
		}
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.mu.Lock()
		if g.charCreation.SelectedClass > 0 {
			g.charCreation.SelectedClass--
		}
		g.mu.Unlock()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.mu.Lock()
		if g.charCreation.SelectedClass < len(ClassInfoList)-1 {
			g.charCreation.SelectedClass++
		}
		g.mu.Unlock()
	}
}

func (g *Game) drawCharCreationClass(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "Select your class:", 250, 90)

	g.mu.RLock()
	selectedIdx := g.charCreation.SelectedClass
	g.mu.RUnlock()

	for i, cls := range ClassInfoList {
		y := 120 + i*60
		bgColor := color.RGBA{R: 40, G: 40, B: 55, A: 255}
		if i == selectedIdx {
			bgColor = color.RGBA{R: 60, G: 50, B: 80, A: 255}
		}

		drawRect(screen, 100, y, 600, 52, bgColor)
		drawRectOutline(screen, 100, y, 600, 52, color.RGBA{R: 80, G: 80, B: 120, A: 255})

		marker := "  "
		if i == selectedIdx {
			marker = "> "
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%s (%s)", marker, cls.Name, cls.HitDice), 110, y+5)
		ebitenutil.DebugPrintAt(screen, cls.Description, 130, y+20)

		wpns := ""
		for j, w := range cls.WeaponProficiencies {
			if j > 0 {
				wpns += ", "
			}
			wpns += w
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Weapons: %s", truncateText(wpns, 50)), 130, y+35)
	}
}

// --- Step 3: Attributes ---

func (g *Game) updateCharCreationAttributes() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.mode = ModeNormal
		g.screenState = ScreenMainMenu
		g.charCreation = CharCreationState{}
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.mu.Lock()
		g.charCreation.Step = CharStepClass
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.mu.Lock()
		g.charCreation.Step = CharStepReview
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.mu.Lock()
		if g.charCreation.SelectedAttr > 0 {
			g.charCreation.SelectedAttr--
		}
		g.mu.Unlock()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.mu.Lock()
		if g.charCreation.SelectedAttr < 5 {
			g.charCreation.SelectedAttr++
		}
		g.mu.Unlock()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.mu.Lock()
		idx := g.charCreation.SelectedAttr
		val := g.charCreation.GetAttr(idx)
		if val > 8 {
			g.charCreation.SetAttr(idx, val-1)
		}
		g.mu.Unlock()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.mu.Lock()
		idx := g.charCreation.SelectedAttr
		val := g.charCreation.GetAttr(idx)
		if val < 18 {
			g.charCreation.SetAttr(idx, val+1)
		}
		g.mu.Unlock()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.mu.Lock()
		switch g.charCreation.AttrMethod {
		case AttrMethodStandard:
			g.charCreation.AttrMethod = AttrMethodPointBuy
			g.charCreation.ResetAttributes(8)
		case AttrMethodPointBuy:
			g.charCreation.AttrMethod = AttrMethodRoll
			g.charCreation.ResetAttributes(10)
		default:
			g.charCreation.AttrMethod = AttrMethodStandard
			g.charCreation.SetStandardArray()
		}
		g.mu.Unlock()
	}
}

func (g *Game) drawCharCreationAttributes(screen *ebiten.Image) {
	g.mu.RLock()
	method := g.charCreation.AttrMethod
	cc := g.charCreation
	selAttr := cc.SelectedAttr
	g.mu.RUnlock()

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Attribute Allocation [Tab to switch: %s]", method), 180, 90)

	attrNames := []string{"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma"}

	for i, name := range attrNames {
		y := 120 + i*40
		bgColor := color.RGBA{R: 40, G: 40, B: 55, A: 255}
		if i == selAttr {
			bgColor = color.RGBA{R: 60, G: 50, B: 80, A: 255}
		}

		drawRect(screen, 200, y, 400, 32, bgColor)
		drawRectOutline(screen, 200, y, 400, 32, color.RGBA{R: 80, G: 80, B: 120, A: 255})

		val := cc.GetAttr(i)
		mod := AttributeModifier(val)

		marker := "  "
		if i == selAttr {
			marker = "> "
		}

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%-14s", marker, name), 210, y+8)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("< %2d > (%+d)", val, mod), 400, y+8)
	}

	if method == AttrMethodPointBuy {
		total := 0
		for i := 0; i < 6; i++ {
			total += PointBuyCost(cc.GetAttr(i))
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Points: %d / %d", total, PointBuyTotal), 300, 380)
	}

	ebitenutil.DebugPrintAt(screen, "Up/Down: Select  |  Left/Right: Adjust  |  Tab: Method", 150, 420)
}

// --- Step 4: Review ---

func (g *Game) updateCharCreationReview() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.mode = ModeNormal
		g.screenState = ScreenMainMenu
		g.charCreation = CharCreationState{}
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.mu.Lock()
		g.charCreation.Step = CharStepAttributes
		g.mu.Unlock()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.submitCharacterCreation()
		return
	}
}

func (g *Game) drawCharCreationReview(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "REVIEW YOUR CHARACTER", 280, 90)

	g.mu.RLock()
	cc := g.charCreation
	g.mu.RUnlock()

	y := 120
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Name:  %s", cc.Name), 250, y)
	y += 25

	className := ""
	if cc.SelectedClass < len(ClassInfoList) {
		className = ClassInfoList[cc.SelectedClass].Name
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Class: %s", className), 250, y)
	y += 40

	attrNames := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	ebitenutil.DebugPrintAt(screen, "Attributes:", 250, y)
	y += 20
	for i, name := range attrNames {
		val := cc.GetAttr(i)
		mod := AttributeModifier(val)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  %s: %2d (%+d)", name, val, mod), 270, y)
		y += 18
	}

	y += 20
	drawRect(screen, 280, y, 240, 35, color.RGBA{R: 50, G: 100, B: 50, A: 255})
	drawRectOutline(screen, 280, y, 240, 35, color.RGBA{R: 80, G: 180, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "[Enter] Create Character", 300, y+10)
}

// submitCharacterCreation sends the character creation request via RPC.
func (g *Game) submitCharacterCreation() {
	g.mu.RLock()
	cc := g.charCreation
	g.mu.RUnlock()

	className := ""
	if cc.SelectedClass < len(ClassInfoList) {
		className = ClassInfoList[cc.SelectedClass].Name
	}

	go func() {
		attrs := cc.Attributes
		result, err := g.rpcClient.CreateCharacter(cc.Name, className, cc.AttrMethod.String(), &attrs)
		if err != nil {
			g.showError(fmt.Sprintf("Character creation failed: %v", err))
			return
		}
		if result.Success {
			g.addLogMessage(fmt.Sprintf("Character '%s' created!", cc.Name), MessageSystem)
			g.mu.Lock()
			g.mode = ModeNormal
			g.screenState = ScreenExploration
			g.charCreation = CharCreationState{}
			g.mu.Unlock()
			go g.refreshGameState()
		} else {
			g.showError("Character creation failed: " + result.Message)
		}
	}()
}
