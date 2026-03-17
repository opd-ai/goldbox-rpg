//go:build js && wasm

package wasmui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// selectionDelta computes the selection change from keyboard, touch, and mouse
// input for a list with n items. Returns the delta to apply to the current
// selection (typically -1, 0, or +1).
func (g *Game) selectionDelta(current, max int) int {
	// Keyboard up/down
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && current > 0 {
		return -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && current < max-1 {
		return 1
	}

	// Touch swipe
	if swiped, dir := g.touchState.HasSwipe(); swiped {
		switch dir {
		case GestureSwipeUp:
			if current > 0 {
				return -1
			}
		case GestureSwipeDown:
			if current < max-1 {
				return 1
			}
		}
	}

	// Mouse wheel
	_, wy := mouseWheelDelta()
	if wy < 0 && current < max-1 {
		return 1
	} else if wy > 0 && current > 0 {
		return -1
	}

	return 0
}

// inventoryTouchButtons handles touch taps on inventory overlay buttons.
// Returns true if a tap was handled (and the caller should return early).
func (g *Game) inventoryTouchButtons(items []ItemData, sel int) bool {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return false
	}

	// Close button (top-right)
	closeBtnX := ScreenWidth - overlayCloseBtnW - 10
	if PointInRect(tx, ty, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH) {
		g.closeInventory()
		return true
	}

	// Equip/Unequip button
	equipBtnX, equipBtnY, equipBtnW, equipBtnH := 350, 540, 120, 28
	if PointInRect(tx, ty, equipBtnX, equipBtnY, equipBtnW, equipBtnH) {
		if len(items) > 0 && sel < len(items) {
			g.toggleEquipItem(items[sel])
		}
		return true
	}

	// Use button
	useBtnX, useBtnY, useBtnW, useBtnH := 490, 540, 100, 28
	if PointInRect(tx, ty, useBtnX, useBtnY, useBtnW, useBtnH) {
		if len(items) > 0 && sel < len(items) {
			g.useItem(items[sel])
		}
		return true
	}

	return false
}

// inventoryTouchListSelect handles touch taps on inventory list items.
// Returns the newly selected index, or -1 if no item was tapped.
func (g *Game) inventoryTouchListSelect() int {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return -1
	}

	listX := 350
	listY := 75
	for i := 0; i < 15; i++ {
		y := listY + i*32
		if PointInRect(tx, ty, listX, y, 410, 28) {
			return i
		}
	}
	return -1
}

// spellbookTouchButtons handles touch taps on spellbook overlay buttons.
// Returns true if a tap was handled (and the caller should return early).
func (g *Game) spellbookTouchButtons(sel int) bool {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return false
	}

	// Close button (top-right)
	closeBtnX := ScreenWidth - overlayCloseBtnW - 10
	if tx >= closeBtnX && tx <= closeBtnX+overlayCloseBtnW && ty >= 10 && ty <= 10+overlayCloseBtnH {
		g.closeSpellbook()
		return true
	}

	// Cast button
	castBtnX, castBtnY, castBtnW, castBtnH := 200, 555, 120, 28
	if tx >= castBtnX && tx <= castBtnX+castBtnW && ty >= castBtnY && ty <= castBtnY+castBtnH {
		g.castSelectedSpell(sel)
		return true
	}

	// Filter button
	filterBtnX, filterBtnY, filterBtnW, filterBtnH := 480, 555, 120, 28
	if tx >= filterBtnX && tx <= filterBtnX+filterBtnW && ty >= filterBtnY && ty <= filterBtnY+filterBtnH {
		g.cycleSpellFilter()
		return true
	}

	return false
}

// spellbookTouchListSelect handles touch taps on spell list items.
// Returns the newly selected index, or -1 if no item was tapped.
func (g *Game) spellbookTouchListSelect(maxItems int) int {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return -1
	}

	listY := 70
	for i := 0; i < 16 && i < maxItems; i++ {
		y := listY + i*28
		if tx >= 50 && tx <= 750 && ty >= y && ty <= y+24 {
			return i
		}
	}
	return -1
}

// questLogTouchButtons handles touch taps on quest log overlay buttons.
// Returns true if a tap was handled (and the caller should return early).
func (g *Game) questLogTouchButtons(panelX, panelY, panelW int) bool {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return false
	}

	// Close button (top-right of panel)
	closeBtnX := panelX + panelW - overlayCloseBtnW - 10
	if tx >= closeBtnX && tx <= closeBtnX+overlayCloseBtnW && ty >= panelY+5 && ty <= panelY+5+overlayCloseBtnH {
		g.mu.Lock()
		g.overlays.ShowQuestLog = false
		g.mu.Unlock()
		return true
	}

	// Tab bar
	for i := 0; i < 3; i++ {
		tabX := panelX + 20 + i*120
		if tx >= tabX && tx <= tabX+100 && ty >= panelY+30 && ty <= panelY+52 {
			g.mu.Lock()
			g.questLogTab = i
			g.selectedQuest = 0
			g.mu.Unlock()
			return true
		}
	}

	return false
}

// questLogTouchListSelect handles touch taps on quest list items.
// Returns the newly selected index, or -1 if no item was tapped.
func (g *Game) questLogTouchListSelect(panelX, panelY, panelW, total int) int {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return -1
	}

	questY := panelY + 65
	for idx := 0; idx < total && idx < 8; idx++ {
		y := questY + idx*44
		if tx >= panelX+15 && tx <= panelX+panelW-15 && ty >= y && ty <= y+40 {
			return idx
		}
	}
	return -1
}

// adventureTouchButtons handles touch taps on adventure screen buttons.
// Returns true if a tap was handled (and the caller should return early).
func adventureTouchButtons(g *Game, s *AdventureScreen) bool {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return false
	}

	// Back button (top-right)
	if tx >= ScreenWidth-80 && tx <= ScreenWidth-10 && ty >= 5 && ty <= 30 {
		g.mu.Lock()
		g.mode = ModeNormal
		g.screenState = ScreenMainMenu
		g.mu.Unlock()
		return true
	}

	// Load button (bottom bar)
	if tx >= 200 && tx <= 310 && ty >= ScreenHeight-48 && ty <= ScreenHeight-20 {
		if len(s.adventures) > 0 {
			s.loadSelectedAdventure(g)
			return true
		}
	}

	return false
}

// adventureTouchListSelect handles touch taps on adventure list items.
// Returns the newly selected index, -2 if the selected item was tapped (load),
// or -1 if nothing was tapped.
func adventureTouchListSelect(g *Game, s *AdventureScreen) int {
	tapped, tx, ty := g.touchState.HasTap()
	if !tapped {
		return -1
	}

	listTop := 45
	listBottom := ScreenHeight - 60
	if tx >= 10 && tx <= 390 && ty >= listTop && ty <= listBottom {
		idx := (ty - listTop) / 30
		if idx >= 0 && idx < len(s.adventures) {
			if idx == s.selectedIndex {
				return -2 // Signal to load
			}
			return idx
		}
	}
	return -1
}
