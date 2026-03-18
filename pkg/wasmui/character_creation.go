//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// prevCharCreationStep tracks the previous character creation step so we can
// perform one-time actions (such as hiding the soft keyboard) on transitions.
var prevCharCreationStep CharCreationStep = CharStepName

// updateCharacterCreation handles input for the character creation flow (§4).
func (g *Game) updateCharacterCreation() {
	g.mu.RLock()
	step := g.charCreation.Step
	g.mu.RUnlock()

	// Dismiss the soft keyboard once when transitioning away from the name
	// entry step, instead of every frame.
	if prevCharCreationStep == CharStepName && step != CharStepName {
		hideSoftKeyboard()
	}
	prevCharCreationStep = step

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

// handleCharCreationEscape handles escape key press during character creation.
// Returns true if escape was pressed and handled.
func (g *Game) handleCharCreationEscape() bool {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return false
	}
	g.mu.Lock()
	g.mode = ModeAdventureSelect
	g.charCreation = CharCreationState{}
	g.mu.Unlock()
	g.adventureScreen.RefreshAdventures(g)
	return true
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

// Name text-box layout (shared with softkeyboard.go positioning).
const (
	nameBoxX = 250
	nameBoxY = 150
	nameBoxW = 300
	nameBoxH = 30

	// "Next" button sits to the right of the name box.
	nameNextBtnX = 560
	nameNextBtnY = nameBoxY
	nameNextBtnW = 80
	nameNextBtnH = nameBoxH
)

func (g *Game) updateCharCreationName() {
	if g.handleCharCreationEscape() {
		hideSoftKeyboard()
		return
	}

	// Show transparent soft keyboard overlay for mobile touch input.
	g.mu.RLock()
	currentName := g.charCreation.Name
	g.mu.RUnlock()
	showSoftKeyboard(currentName)

	// Check for Enter from either the soft keyboard or a physical keyboard.
	if softKeyboardEnterPressed() || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.tryAdvanceFromName()
		return
	}

	// Touch tap on "Next" button for mobile navigation.
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		if tx >= nameNextBtnX && tx <= nameNextBtnX+nameNextBtnW &&
			ty >= nameNextBtnY && ty <= nameNextBtnY+nameNextBtnH {
			g.tryAdvanceFromName()
		}
	}

	// When the soft keyboard has focus, it is the authoritative input source.
	if isSoftKeyboardFocused() {
		g.syncNameFromSoftKeyboard()
		return
	}

	// Physical keyboard: backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.mu.Lock()
		if len(g.charCreation.Name) > 0 {
			_, size := utf8.DecodeLastRuneInString(g.charCreation.Name)
			if size > 0 {
				g.charCreation.Name = g.charCreation.Name[:len(g.charCreation.Name)-size]
			}
		}
		name := g.charCreation.Name
		g.mu.Unlock()
		setSoftKeyboardValue(name)
		return
	}

	// Physical keyboard: character input
	runes := ebiten.AppendInputChars(nil)
	if len(runes) > 0 {
		g.mu.Lock()
		for _, r := range runes {
			if utf8.RuneCountInString(g.charCreation.Name) < 30 {
				g.charCreation.Name += string(r)
			}
		}
		name := g.charCreation.Name
		g.mu.Unlock()
		setSoftKeyboardValue(name)
	}
}

// tryAdvanceFromName syncs the soft keyboard value, then advances to the
// class step if the name is non-empty.
func (g *Game) tryAdvanceFromName() {
	g.syncNameFromSoftKeyboard()
	g.mu.RLock()
	name := g.charCreation.Name
	g.mu.RUnlock()
	if name != "" {
		hideSoftKeyboard()
		g.mu.Lock()
		g.charCreation.Step = CharStepClass
		g.mu.Unlock()
	}
}

// syncNameFromSoftKeyboard reads the current value of the soft keyboard input
// and updates charCreation.Name, enforcing the 30-character (rune) limit.
func (g *Game) syncNameFromSoftKeyboard() {
	val := softKeyboardValue()
	if utf8.RuneCountInString(val) > 30 {
		runes := []rune(val)
		val = string(runes[:30])
		setSoftKeyboardValue(val)
	}
	g.mu.Lock()
	g.charCreation.Name = val
	g.mu.Unlock()
}

func (g *Game) drawCharCreationName(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "Enter your character's name:", nameBoxX, 120)

	g.mu.RLock()
	name := g.charCreation.Name
	g.mu.RUnlock()

	drawRect(screen, nameBoxX, nameBoxY, nameBoxW, nameBoxH, color.RGBA{R: 50, G: 50, B: 70, A: 255})
	drawRectOutline(screen, nameBoxX, nameBoxY, nameBoxW, nameBoxH, color.RGBA{R: 120, G: 120, B: 180, A: 255})

	ebitenutil.DebugPrintAt(screen, name+"_", nameBoxX+10, nameBoxY+8)
	ebitenutil.DebugPrintAt(screen, "(1-30 characters)", 280, 190)

	// "Next" button for touch navigation
	drawRect(screen, nameNextBtnX, nameNextBtnY, nameNextBtnW, nameNextBtnH, color.RGBA{R: 50, G: 100, B: 50, A: 255})
	drawRectOutline(screen, nameNextBtnX, nameNextBtnY, nameNextBtnW, nameNextBtnH, color.RGBA{R: 80, G: 180, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "Next >", nameNextBtnX+10, nameNextBtnY+8)

	// Hint for touch-capable devices when the input is empty and unfocused
	if hasTouchSupport() && name == "" && !isSoftKeyboardFocused() {
		ebitenutil.DebugPrintAt(screen, "Tap the box above to type", 270, 210)
	}
}

// --- Step 2: Class ---

func (g *Game) updateCharCreationClass() {
	if g.handleCharCreationEscape() {
		return
	}
	if g.handleClassStepBack() {
		return
	}
	if g.handleClassStepAdvance() {
		return
	}
	g.handleClassKeyboardNav()
	if g.handleClassTouchTap() {
		return
	}
	g.handleClassTouchSwipe()
}

// handleClassStepBack returns to the name step on backspace.
func (g *Game) handleClassStepBack() bool {
	if !inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		return false
	}
	g.mu.Lock()
	g.charCreation.Step = CharStepName
	g.mu.Unlock()
	return true
}

// handleClassStepAdvance moves to the attributes step on enter.
func (g *Game) handleClassStepAdvance() bool {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return false
	}
	g.advanceToAttributesStep()
	return true
}

// advanceToAttributesStep initializes attribute method and moves to attributes step.
func (g *Game) advanceToAttributesStep() {
	g.mu.Lock()
	g.charCreation.Step = CharStepAttributes
	if g.charCreation.AttrMethod == 0 {
		g.charCreation.AttrMethod = AttrMethodStandard
		g.charCreation.SetStandardArray()
	}
	g.mu.Unlock()
}

// handleClassKeyboardNav handles up/down arrow and W/S key navigation.
func (g *Game) handleClassKeyboardNav() {
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

// handleClassTouchTap handles touch taps on class list items.
// Tapping the selected class advances to attributes; tapping another selects it.
func (g *Game) handleClassTouchTap() bool {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return false
	}
	for i := range ClassInfoList {
		y := 120 + i*60
		if tx >= 100 && tx <= 700 && ty >= y && ty <= y+52 {
			g.mu.Lock()
			if g.charCreation.SelectedClass == i {
				g.mu.Unlock()
				g.advanceToAttributesStep()
				return true
			}
			g.charCreation.SelectedClass = i
			g.mu.Unlock()
			return false
		}
	}
	return false
}

// handleClassTouchSwipe handles swipe gestures for class list navigation.
func (g *Game) handleClassTouchSwipe() {
	swiped, dir := g.touchState.HasSwipe()
	if !swiped {
		return
	}
	g.mu.Lock()
	switch dir {
	case GestureSwipeUp:
		if g.charCreation.SelectedClass > 0 {
			g.charCreation.SelectedClass--
		}
	case GestureSwipeDown:
		if g.charCreation.SelectedClass < len(ClassInfoList)-1 {
			g.charCreation.SelectedClass++
		}
	}
	g.mu.Unlock()
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

// handleAttrSelection handles up/down arrow key presses for attribute selection.
func (g *Game) handleAttrSelection() {
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
}

// handleAttrAdjustment handles left/right arrow key presses for attribute value adjustment.
func (g *Game) handleAttrAdjustment() {
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
}

// cycleAttrMethod cycles through attribute allocation methods on Tab key press.
func (g *Game) cycleAttrMethod() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	switch g.charCreation.AttrMethod {
	case AttrMethodRoll:
		g.charCreation.AttrMethod = AttrMethodStandard
		g.charCreation.SetStandardArray()
	case AttrMethodStandard:
		g.charCreation.AttrMethod = AttrMethodPointBuy
		g.charCreation.ResetAttributes(8)
		g.charCreation.PointBuyPoints = 27
	case AttrMethodPointBuy:
		g.charCreation.AttrMethod = AttrMethodCustom
		g.charCreation.ResetAttributes(10)
	case AttrMethodCustom:
		g.charCreation.AttrMethod = AttrMethodRoll
		g.charCreation.ResetAttributes(10)
	default:
		g.charCreation.AttrMethod = AttrMethodStandard
		g.charCreation.SetStandardArray()
	}
}

// Attribute step touch button layout constants.
const (
	attrDecrBtnX = 390 // "<" button X
	attrIncrBtnX = 455 // ">" button X
	attrBtnW     = 28
	attrBtnH     = 28

	attrBackBtnX = 200
	attrBackBtnY = 430
	attrBackBtnW = 100
	attrBackBtnH = 32

	attrNextBtnX = 320
	attrNextBtnY = 430
	attrNextBtnW = 100
	attrNextBtnH = 32

	attrMethodBtnX = 460
	attrMethodBtnY = 430
	attrMethodBtnW = 120
	attrMethodBtnH = 32
)

// handleAttrTouchNavigation handles touch input for navigation buttons on attribute step.
// Returns true if a navigation action was performed.
func (g *Game) handleAttrTouchNavigation(tx, ty int) bool {
	// Back button
	if tx >= attrBackBtnX && tx <= attrBackBtnX+attrBackBtnW && ty >= attrBackBtnY && ty <= attrBackBtnY+attrBackBtnH {
		g.mu.Lock()
		g.charCreation.Step = CharStepClass
		g.mu.Unlock()
		return true
	}
	// Next button
	if tx >= attrNextBtnX && tx <= attrNextBtnX+attrNextBtnW && ty >= attrNextBtnY && ty <= attrNextBtnY+attrNextBtnH {
		g.mu.Lock()
		g.charCreation.Step = CharStepReview
		g.mu.Unlock()
		return true
	}
	// Method button
	if tx >= attrMethodBtnX && tx <= attrMethodBtnX+attrMethodBtnW && ty >= attrMethodBtnY && ty <= attrMethodBtnY+attrMethodBtnH {
		g.cycleAttrMethodDirect()
		return true
	}
	return false
}

// handleAttrTouchAdjust handles touch input for attribute +/- buttons and row selection.
func (g *Game) handleAttrTouchAdjust(tx, ty int) {
	for i := 0; i < 6; i++ {
		y := 120 + i*40
		// Decrement "<" button
		if tx >= attrDecrBtnX && tx <= attrDecrBtnX+attrBtnW && ty >= y+2 && ty <= y+2+attrBtnH {
			g.mu.Lock()
			g.charCreation.SelectedAttr = i
			val := g.charCreation.GetAttr(i)
			if val > 8 {
				g.charCreation.SetAttr(i, val-1)
			}
			g.mu.Unlock()
			return
		}
		// Increment ">" button
		if tx >= attrIncrBtnX && tx <= attrIncrBtnX+attrBtnW && ty >= y+2 && ty <= y+2+attrBtnH {
			g.mu.Lock()
			g.charCreation.SelectedAttr = i
			val := g.charCreation.GetAttr(i)
			if val < 18 {
				g.charCreation.SetAttr(i, val+1)
			}
			g.mu.Unlock()
			return
		}
		// Row selection (rest of the row)
		if tx >= 200 && tx <= 600 && ty >= y && ty <= y+32 {
			g.mu.Lock()
			g.charCreation.SelectedAttr = i
			g.mu.Unlock()
			break
		}
	}
}

func (g *Game) updateCharCreationAttributes() {
	if g.handleCharCreationEscape() {
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

	g.handleAttrSelection()
	g.handleAttrAdjustment()
	g.cycleAttrMethod()

	// Touch tap on attribute rows, +/- buttons, and navigation buttons
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		if g.handleAttrTouchNavigation(tx, ty) {
			return
		}
		g.handleAttrTouchAdjust(tx, ty)
	}
}

// cycleAttrMethodDirect cycles through attribute allocation methods directly (for touch).
func (g *Game) cycleAttrMethodDirect() {
	g.cycleAttrMethod()
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

		// Touch-friendly "<" and ">" buttons for adjusting values
		decrColor := color.RGBA{R: 80, G: 50, B: 50, A: 255}
		incrColor := color.RGBA{R: 50, G: 80, B: 50, A: 255}
		drawRect(screen, attrDecrBtnX, y+2, attrBtnW, attrBtnH, decrColor)
		drawRectOutline(screen, attrDecrBtnX, y+2, attrBtnW, attrBtnH, color.RGBA{R: 140, G: 80, B: 80, A: 255})
		ebitenutil.DebugPrintAt(screen, "<", attrDecrBtnX+10, y+8)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%2d", val), attrDecrBtnX+attrBtnW+6, y+8)

		drawRect(screen, attrIncrBtnX, y+2, attrBtnW, attrBtnH, incrColor)
		drawRectOutline(screen, attrIncrBtnX, y+2, attrBtnW, attrBtnH, color.RGBA{R: 80, G: 140, B: 80, A: 255})
		ebitenutil.DebugPrintAt(screen, ">", attrIncrBtnX+10, y+8)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("(%+d)", mod), attrIncrBtnX+attrBtnW+10, y+8)
	}

	if method == AttrMethodPointBuy {
		total := 0
		for i := 0; i < 6; i++ {
			total += PointBuyCost(cc.GetAttr(i))
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Points: %d / %d", total, PointBuyTotal), 300, 380)
	}

	// Touch navigation buttons
	drawRect(screen, attrBackBtnX, attrBackBtnY, attrBackBtnW, attrBackBtnH, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, attrBackBtnX, attrBackBtnY, attrBackBtnW, attrBackBtnH, color.RGBA{R: 100, G: 100, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "< Back", attrBackBtnX+15, attrBackBtnY+8)

	drawRect(screen, attrNextBtnX, attrNextBtnY, attrNextBtnW, attrNextBtnH, color.RGBA{R: 50, G: 100, B: 50, A: 255})
	drawRectOutline(screen, attrNextBtnX, attrNextBtnY, attrNextBtnW, attrNextBtnH, color.RGBA{R: 80, G: 180, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "Next >", attrNextBtnX+15, attrNextBtnY+8)

	drawRect(screen, attrMethodBtnX, attrMethodBtnY, attrMethodBtnW, attrMethodBtnH, color.RGBA{R: 50, G: 50, B: 80, A: 255})
	drawRectOutline(screen, attrMethodBtnX, attrMethodBtnY, attrMethodBtnW, attrMethodBtnH, color.RGBA{R: 80, G: 80, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "Method", attrMethodBtnX+25, attrMethodBtnY+8)

	ebitenutil.DebugPrintAt(screen, "Up/Down: Select  |  Left/Right: Adjust  |  Tab: Method", 150, 470)
}

// --- Step 4: Review ---

func (g *Game) updateCharCreationReview() {
	if g.handleCharCreationEscape() {
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

	// Touch tap on Back / Create buttons
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		g.mu.RLock()
		cc := g.charCreation
		g.mu.RUnlock()
		btnY := reviewButtonY(cc)
		// Back button: (250, btnY, 120, 35)
		if tx >= 250 && tx <= 370 && ty >= btnY && ty <= btnY+35 {
			g.mu.Lock()
			g.charCreation.Step = CharStepAttributes
			g.mu.Unlock()
			return
		}
		// Create button: (400, btnY, 150, 35)
		if tx >= 400 && tx <= 550 && ty >= btnY && ty <= btnY+35 {
			g.submitCharacterCreation()
			return
		}
	}
}

func (g *Game) drawCharCreationReview(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "CHARACTER SUMMARY", 300, 90)

	g.mu.RLock()
	cc := g.charCreation
	g.mu.RUnlock()

	y := 120
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Name:   %s", cc.Name), 250, y)
	y += 22

	var classInfo *ClassInfo
	className := ""
	if cc.SelectedClass < len(ClassInfoList) {
		classInfo = &ClassInfoList[cc.SelectedClass]
		className = classInfo.Name
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Class:  %s", className), 250, y)
	y += 22
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Method: %s", cc.AttrMethod), 250, y)
	y += 30

	// Attributes in two columns per §4 Step 4
	attrNames := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	for i := 0; i < 6; i += 2 {
		v1 := cc.GetAttr(i)
		v2 := cc.GetAttr(i + 1)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s: %2d (%+d)    %s: %2d (%+d)",
			attrNames[i], v1, AttributeModifier(v1),
			attrNames[i+1], v2, AttributeModifier(v2)), 250, y)
		y += 18
	}
	y += 15

	// HP and AC preview per §4 Step 4
	conMod := AttributeModifier(cc.Attributes.Constitution)
	dexMod := AttributeModifier(cc.Attributes.Dexterity)
	hpPreview := 0
	hitDice := ""
	if classInfo != nil {
		hitDice = classInfo.HitDice
		switch hitDice {
		case "d4":
			hpPreview = 4 + conMod
		case "d6":
			hpPreview = 6 + conMod
		case "d8":
			hpPreview = 8 + conMod
		case "d10":
			hpPreview = 10 + conMod
		}
		if hpPreview < 1 {
			hpPreview = 1
		}
	}
	acPreview := 10 + dexMod
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %d           AC: %d", hpPreview, acPreview), 250, y)
	y += 18
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Hit Dice: %s", hitDice), 250, y)
	y += 25

	// Proficiencies per §4 Step 4
	if classInfo != nil {
		wpns := ""
		for j, w := range classInfo.WeaponProficiencies {
			if j > 0 {
				wpns += ", "
			}
			wpns += w
		}
		ebitenutil.DebugPrintAt(screen, "Weapons: "+truncateText(wpns, 45), 250, y)
		y += 16
		armors := ""
		for j, a := range classInfo.ArmorProficiencies {
			if j > 0 {
				armors += ", "
			}
			armors += a
		}
		if armors == "" {
			armors = "none"
		}
		ebitenutil.DebugPrintAt(screen, "Armor:   "+armors, 250, y)
		y += 16
		shieldStr := "No"
		if classInfo.ShieldProficiency {
			shieldStr = "Yes"
		}
		ebitenutil.DebugPrintAt(screen, "Shield:  "+shieldStr, 250, y)
		y += 25
	}

	// Buttons per §4 Step 4
	drawRect(screen, 250, y, 120, 35, color.RGBA{R: 60, G: 60, B: 80, A: 255})
	drawRectOutline(screen, 250, y, 120, 35, color.RGBA{R: 100, G: 100, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "[Bksp] Back", 260, y+10)

	drawRect(screen, 400, y, 150, 35, color.RGBA{R: 50, G: 100, B: 50, A: 255})
	drawRectOutline(screen, 400, y, 150, 35, color.RGBA{R: 80, G: 180, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "[Enter] Create", 415, y+10)
}

// reviewButtonY computes the Y position of the Back/Create buttons on the
// review screen, mirroring the layout in drawCharCreationReview.
func reviewButtonY(cc CharCreationState) int {
	y := 120
	y += 22 // Name
	y += 22 // Class
	y += 30 // Method

	// Attributes: 3 rows (2 per row)
	for i := 0; i < 6; i += 2 {
		y += 18
	}
	y += 15

	// HP, Hit Dice
	y += 18
	y += 25

	// Proficiencies (only if class info present)
	if cc.SelectedClass < len(ClassInfoList) {
		y += 16      // Weapons
		y += 16      // Armor
		y += 16 + 25 // Shield + gap
	}
	return y
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
