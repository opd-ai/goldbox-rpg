//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ======================
// Inventory Screen (§6)
// ======================

// Overlay button layout constants for touch-friendly close/action buttons.
const (
	overlayCloseBtnW = 60
	overlayCloseBtnH = 28
)

// updateInventory handles input for the inventory/equipment screen.
func (g *Game) updateInventory() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.closeInventory()
		return
	}

	g.mu.RLock()
	items := g.inventoryItems
	sel := g.selectedItem
	g.mu.RUnlock()

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if sel > 0 {
			g.mu.Lock()
			g.selectedItem--
			g.mu.Unlock()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if sel < len(items)-1 {
			g.mu.Lock()
			g.selectedItem++
			g.mu.Unlock()
		}
	}

	// Touch swipe or mouse wheel to scroll inventory list
	if swiped, dir := g.touchState.HasSwipe(); swiped {
		switch dir {
		case GestureSwipeUp:
			if sel > 0 {
				g.mu.Lock()
				g.selectedItem--
				g.mu.Unlock()
			}
		case GestureSwipeDown:
			if sel < len(items)-1 {
				g.mu.Lock()
				g.selectedItem++
				g.mu.Unlock()
			}
		}
	}
	_, wy := mouseWheelDelta()
	if wy < 0 && sel < len(items)-1 {
		g.mu.Lock()
		g.selectedItem++
		g.mu.Unlock()
	} else if wy > 0 && sel > 0 {
		g.mu.Lock()
		g.selectedItem--
		g.mu.Unlock()
	}

	// Touch tap on inventory items and action buttons
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		// Close button (top-right)
		closeBtnX := ScreenWidth - overlayCloseBtnW - 10
		if PointInRect(tx, ty, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH) {
			g.closeInventory()
			return
		}

		// Equip/Unequip button
		equipBtnX, equipBtnY, equipBtnW, equipBtnH := 350, 540, 120, 28
		if PointInRect(tx, ty, equipBtnX, equipBtnY, equipBtnW, equipBtnH) {
			if len(items) > 0 && sel < len(items) {
				g.toggleEquipItem(items[sel])
			}
			return
		}

		// Use button
		useBtnX, useBtnY, useBtnW, useBtnH := 490, 540, 100, 28
		if PointInRect(tx, ty, useBtnX, useBtnY, useBtnW, useBtnH) {
			if len(items) > 0 && sel < len(items) {
				g.useItem(items[sel])
			}
			return
		}

		// Item list tap
		listX := 350
		listY := 75
		for i := range items {
			if i >= 15 {
				break
			}
			y := listY + i*32
			if PointInRect(tx, ty, listX, y, 410, 28) {
				g.mu.Lock()
				g.selectedItem = i
				g.mu.Unlock()
				break
			}
		}
	}

	// Enter → equip/use item
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(items) > 0 && sel < len(items) {
		g.toggleEquipItem(items[sel])
	}

	// U → use consumable
	if inpututil.IsKeyJustPressed(ebiten.KeyU) && len(items) > 0 && sel < len(items) {
		g.useItem(items[sel])
	}
}

// closeInventory returns from the inventory screen to the previous mode.
func (g *Game) closeInventory() {
	g.mu.Lock()
	g.mode = g.previousMode
	if g.mode == ModeInventory {
		g.mode = ModeNormal // safety fallback
	}
	g.mu.Unlock()
}

// toggleEquipItem equips or unequips the given item.
func (g *Game) toggleEquipItem(item ItemData) {
	go func() {
		if item.Equipped {
			_, err := g.rpcClient.UnequipItem(item.ID)
			if err != nil {
				g.showError(fmt.Sprintf("Unequip failed: %v", err))
			} else {
				g.addLogMessage(fmt.Sprintf("Unequipped %s", item.Name), MessageInfo)
				g.loadInventory()
			}
		} else {
			_, err := g.rpcClient.EquipItem(item.ID, item.Slot)
			if err != nil {
				g.showError(fmt.Sprintf("Equip failed: %v", err))
			} else {
				g.addLogMessage(fmt.Sprintf("Equipped %s", item.Name), MessageInfo)
				g.loadInventory()
			}
		}
	}()
}

// useItem uses the given consumable item.
func (g *Game) useItem(item ItemData) {
	go func() {
		_, err := g.rpcClient.UseItem(item.ID, "")
		if err != nil {
			g.showError(fmt.Sprintf("Use item failed: %v", err))
		} else {
			g.addLogMessage(fmt.Sprintf("Used %s", item.Name), MessageInfo)
			g.loadInventory()
		}
	}()
}

// drawInventoryScreen renders the inventory/equipment panel (§6).
func (g *Game) drawInventoryScreen(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 30, G: 30, B: 40, A: 255})

	ebitenutil.DebugPrintAt(screen, "INVENTORY & EQUIPMENT", 280, 15)

	// Close button (top-right)
	closeBtnX := ScreenWidth - overlayCloseBtnW - 10
	drawRect(screen, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 120, G: 50, B: 50, A: 255})
	drawRectOutline(screen, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 200, G: 80, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "Close", closeBtnX+8, 16)

	ebitenutil.DebugPrintAt(screen, "[I/Esc] Close  |  Enter: Equip/Unequip  |  U: Use", 170, 560)

	g.mu.RLock()
	items := g.inventoryItems
	sel := g.selectedItem
	player := g.player
	g.mu.RUnlock()

	// Equipment slots (left side, §6.1)
	g.drawEquipmentSlots(screen, 30, 50, player)

	// Inventory list (right side)
	listX := 350
	listY := 50
	ebitenutil.DebugPrintAt(screen, "INVENTORY", listX+80, listY)
	listY += 25

	if len(items) == 0 {
		ebitenutil.DebugPrintAt(screen, "(empty)", listX+80, listY)
	}

	for i, item := range items {
		if i >= 15 {
			break
		}
		y := listY + i*32
		bgColor := color.RGBA{R: 40, G: 40, B: 55, A: 255}
		if i == sel {
			bgColor = color.RGBA{R: 60, G: 50, B: 80, A: 255}
		}

		drawRect(screen, listX, y, 410, 28, bgColor)
		drawRectOutline(screen, listX, y, 410, 28, color.RGBA{R: 80, G: 80, B: 100, A: 255})

		equipTag := ""
		if item.Equipped {
			equipTag = "[E] "
		}
		marker := "  "
		if i == sel {
			marker = "> "
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%s%s", marker, equipTag, item.Name), listX+5, y+6)
		ebitenutil.DebugPrintAt(screen, item.Type, listX+300, y+6)
	}

	// Item detail panel (bottom)
	if sel < len(items) && len(items) > 0 {
		g.drawItemDetail(screen, items[sel])
	}

	// Touch action buttons
	equipLabel := "Equip"
	if sel < len(items) && len(items) > 0 && items[sel].Equipped {
		equipLabel = "Unequip"
	}
	drawRect(screen, 350, 540, 120, 28, color.RGBA{R: 50, G: 80, B: 50, A: 255})
	drawRectOutline(screen, 350, 540, 120, 28, color.RGBA{R: 80, G: 140, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, equipLabel, 370, 546)

	drawRect(screen, 490, 540, 100, 28, color.RGBA{R: 50, G: 50, B: 80, A: 255})
	drawRectOutline(screen, 490, 540, 100, 28, color.RGBA{R: 80, G: 80, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "Use", 520, 546)
}

// drawEquipmentSlots renders the character equipment slots.
func (g *Game) drawEquipmentSlots(screen *ebiten.Image, x, y int, player *PlayerState) {
	ebitenutil.DebugPrintAt(screen, "EQUIPPED", x+80, y)
	y += 25

	slots := []string{"Head", "Neck", "Chest", "Hands", "Rings", "Legs", "Feet", "WeaponMain", "WeaponOff"}

	g.mu.RLock()
	equippedMap := make(map[string]string)
	for _, item := range g.inventoryItems {
		if item.Equipped {
			equippedMap[item.Slot] = item.Name
		}
	}
	g.mu.RUnlock()

	for i, slot := range slots {
		sy := y + i*30
		drawRect(screen, x, sy, 280, 26, color.RGBA{R: 45, G: 45, B: 60, A: 255})
		drawRectOutline(screen, x, sy, 280, 26, color.RGBA{R: 70, G: 70, B: 90, A: 255})

		itemName := "(empty)"
		if name, ok := equippedMap[strings.ToLower(slot)]; ok {
			itemName = name
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%-10s: %s", slot, itemName), x+5, sy+6)
	}
}

// drawItemDetail renders the selected item's details.
func (g *Game) drawItemDetail(screen *ebiten.Image, item ItemData) {
	y := 530
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s  |  %s  |  Slot: %s  |  Weight: %d",
		item.Name, item.Type, item.Slot, item.Weight), 30, y)
}

// ======================
// Spellbook Screen (§3.8)
// ======================

// updateSpellbook handles input for the spellbook screen.
func (g *Game) updateSpellbook() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.closeSpellbook()
		return
	}

	g.mu.RLock()
	sel := g.selectedSpell
	g.mu.RUnlock()

	// Use the filtered spell list for all navigation bounds so selection
	// stays consistent with the visible/castable spells.
	filtered := g.filteredSpells()
	filteredLen := len(filtered)

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if sel > 0 {
			g.mu.Lock()
			g.selectedSpell--
			g.mu.Unlock()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if sel < filteredLen-1 {
			g.mu.Lock()
			g.selectedSpell++
			g.mu.Unlock()
		}
	}

	// Touch swipe or mouse wheel to scroll spell list
	if swiped, dir := g.touchState.HasSwipe(); swiped {
		switch dir {
		case GestureSwipeUp:
			if sel > 0 {
				g.mu.Lock()
				g.selectedSpell--
				g.mu.Unlock()
			}
		case GestureSwipeDown:
			if sel < filteredLen-1 {
				g.mu.Lock()
				g.selectedSpell++
				g.mu.Unlock()
			}
		}
	}
	_, wy := mouseWheelDelta()
	if wy < 0 && sel < filteredLen-1 {
		g.mu.Lock()
		g.selectedSpell++
		g.mu.Unlock()
	} else if wy > 0 && sel > 0 {
		g.mu.Lock()
		g.selectedSpell--
		g.mu.Unlock()
	}

	// Touch tap on spell list items and action buttons
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		// Close button (top-right)
		closeBtnX := ScreenWidth - overlayCloseBtnW - 10
		if tx >= closeBtnX && tx <= closeBtnX+overlayCloseBtnW && ty >= 10 && ty <= 10+overlayCloseBtnH {
			g.closeSpellbook()
			return
		}

		// Cast button
		castBtnX, castBtnY, castBtnW, castBtnH := 200, 555, 120, 28
		if tx >= castBtnX && tx <= castBtnX+castBtnW && ty >= castBtnY && ty <= castBtnY+castBtnH {
			g.castSelectedSpell(sel)
			return
		}

		// Filter button
		filterBtnX, filterBtnY, filterBtnW, filterBtnH := 480, 555, 120, 28
		if tx >= filterBtnX && tx <= filterBtnX+filterBtnW && ty >= filterBtnY && ty <= filterBtnY+filterBtnH {
			g.cycleSpellFilter()
			return
		}

		// Spell list items
		listY := 70
		for i := range filtered {
			if i >= 16 {
				break
			}
			y := listY + i*28
			if tx >= 50 && tx <= 750 && ty >= y && ty <= y+24 {
				g.mu.Lock()
				g.selectedSpell = i
				g.mu.Unlock()
				break
			}
		}
	}

	// Tab → cycle level filter
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.cycleSpellFilter()
	}

	// Enter → cast spell
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.castSelectedSpell(sel)
	}
}

// closeSpellbook returns from the spellbook screen to the previous mode.
func (g *Game) closeSpellbook() {
	g.mu.Lock()
	g.mode = g.previousMode
	if g.mode == ModeSpellcasting {
		g.mode = ModeNormal // safety fallback
	}
	g.mu.Unlock()
}

// cycleSpellFilter cycles through spell level filters.
func (g *Game) cycleSpellFilter() {
	g.mu.Lock()
	g.spellFilter++
	if g.spellFilter > 9 {
		g.spellFilter = -1
	}
	g.selectedSpell = 0
	g.mu.Unlock()
}

// castSelectedSpell casts the spell at the given index in the filtered list.
func (g *Game) castSelectedSpell(sel int) {
	filtered := g.filteredSpells()
	if sel < len(filtered) {
		spell := filtered[sel]
		go func() {
			_, err := g.rpcClient.CastSpell(spell.ID, "", nil)
			if err != nil {
				g.showError(fmt.Sprintf("Cast failed: %v", err))
			} else {
				g.addLogMessage(fmt.Sprintf("Cast %s!", spell.Name), MessageCombat)
				g.closeSpellbook()
			}
		}()
	}
}

// drawSpellbookScreen renders the spellbook panel (§3.8).
func (g *Game) drawSpellbookScreen(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 25, G: 25, B: 45, A: 255})

	ebitenutil.DebugPrintAt(screen, "SPELLBOOK", 340, 15)

	// Close button (top-right)
	closeBtnX := ScreenWidth - overlayCloseBtnW - 10
	drawRect(screen, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 120, G: 50, B: 50, A: 255})
	drawRectOutline(screen, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 200, G: 80, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "Close", closeBtnX+8, 16)

	g.mu.RLock()
	filter := g.spellFilter
	sel := g.selectedSpell
	g.mu.RUnlock()

	// Filter indicator
	filterText := "All Levels"
	if filter >= 0 {
		filterText = fmt.Sprintf("Level %d", filter)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Filter: %s  [Tab to change]", filterText), 50, 45)

	// Spell list
	filtered := g.filteredSpells()
	listY := 70
	for i, spell := range filtered {
		if i >= 16 {
			break
		}
		y := listY + i*28
		bgColor := color.RGBA{R: 35, G: 35, B: 55, A: 255}
		if i == sel {
			bgColor = color.RGBA{R: 55, G: 45, B: 80, A: 255}
		}

		drawRect(screen, 50, y, 700, 24, bgColor)
		drawRectOutline(screen, 50, y, 700, 24, color.RGBA{R: 70, G: 70, B: 100, A: 255})

		marker := "  "
		if i == sel {
			marker = "> "
		}
		schoolName := SpellSchoolName(spell.School)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%sLv%d %-20s %s", marker, spell.Level, spell.Name, schoolName), 55, y+4)
	}

	// Spell detail
	if sel < len(filtered) && len(filtered) > 0 {
		g.drawSpellDetail(screen, filtered[sel])
	}

	// Touch action buttons
	drawRect(screen, 200, 555, 120, 28, color.RGBA{R: 50, G: 80, B: 50, A: 255})
	drawRectOutline(screen, 200, 555, 120, 28, color.RGBA{R: 80, G: 140, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "Cast", 240, 561)

	drawRect(screen, 480, 555, 120, 28, color.RGBA{R: 50, G: 50, B: 80, A: 255})
	drawRectOutline(screen, 480, 555, 120, 28, color.RGBA{R: 80, G: 80, B: 140, A: 255})
	ebitenutil.DebugPrintAt(screen, "Filter", 515, 561)

	ebitenutil.DebugPrintAt(screen, "[Esc] Close  |  Enter: Cast  |  Tab: Filter Level", 200, 585)
}

func (g *Game) filteredSpells() []SpellData {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.spellFilter < 0 {
		return g.spellList
	}
	var result []SpellData
	for _, s := range g.spellList {
		if s.Level == g.spellFilter {
			result = append(result, s)
		}
	}
	return result
}

func (g *Game) drawSpellDetail(screen *ebiten.Image, spell SpellData) {
	y := 530
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s  |  Lv%d  |  %s  |  Range: %s  |  %s",
		spell.Name, spell.Level, SpellSchoolName(spell.School), spell.Range, spell.Description), 50, y)
}

// ======================
// Quest Log Overlay (§7)
// ======================

// updateQuestLogOverlay handles input for the quest log overlay.
func (g *Game) updateQuestLogOverlay() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		g.mu.Lock()
		g.overlays.ShowQuestLog = false
		g.mu.Unlock()
		return
	}

	// Tab → cycle quest log tabs: Active / Completed / Failed per §7
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.mu.Lock()
		g.questLogTab = (g.questLogTab + 1) % 3
		g.selectedQuest = 0
		g.mu.Unlock()
		return
	}

	g.mu.RLock()
	ql := g.questLog
	sel := g.selectedQuest
	tab := g.questLogTab
	g.mu.RUnlock()

	if ql == nil {
		return
	}

	var total int
	switch tab {
	case 0:
		total = len(ql.ActiveQuests)
	case 1:
		total = len(ql.CompletedQuests)
	case 2:
		total = len(ql.FailedQuests)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && sel > 0 {
		g.mu.Lock()
		g.selectedQuest--
		g.mu.Unlock()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && sel < total-1 {
		g.mu.Lock()
		g.selectedQuest++
		g.mu.Unlock()
	}

	// Touch swipe or mouse wheel for quest list navigation
	if swiped, dir := g.touchState.HasSwipe(); swiped {
		switch dir {
		case GestureSwipeUp:
			if sel > 0 {
				g.mu.Lock()
				g.selectedQuest--
				g.mu.Unlock()
			}
		case GestureSwipeDown:
			if sel < total-1 {
				g.mu.Lock()
				g.selectedQuest++
				g.mu.Unlock()
			}
		}
	}
	_, wy := mouseWheelDelta()
	if wy < 0 && sel < total-1 {
		g.mu.Lock()
		g.selectedQuest++
		g.mu.Unlock()
	} else if wy > 0 && sel > 0 {
		g.mu.Lock()
		g.selectedQuest--
		g.mu.Unlock()
	}

	// Touch tap on quest log tab bar, quest items, and close button
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		panelX := 80
		panelY := 60
		panelW := g.screenWidth - 160

		// Close button (top-right of panel)
		closeBtnX := panelX + panelW - overlayCloseBtnW - 10
		if tx >= closeBtnX && tx <= closeBtnX+overlayCloseBtnW && ty >= panelY+5 && ty <= panelY+5+overlayCloseBtnH {
			g.mu.Lock()
			g.overlays.ShowQuestLog = false
			g.mu.Unlock()
			return
		}

		// Tab bar
		for i := 0; i < 3; i++ {
			tabX := panelX + 20 + i*120
			if tx >= tabX && tx <= tabX+100 && ty >= panelY+30 && ty <= panelY+52 {
				g.mu.Lock()
				g.questLogTab = i
				g.selectedQuest = 0
				g.mu.Unlock()
				return
			}
		}

		// Quest list items
		questY := panelY + 65
		for idx := 0; idx < total; idx++ {
			if idx >= 8 {
				break
			}
			y := questY + idx*44
			if tx >= panelX+15 && tx <= panelX+panelW-15 && ty >= y && ty <= y+40 {
				g.mu.Lock()
				g.selectedQuest = idx
				g.mu.Unlock()
				return
			}
		}
	}
}

// drawQuestLogOverlay renders the quest log overlay (§7).
func (g *Game) drawQuestLogOverlay(screen *ebiten.Image) {
	// Dim background
	drawRect(screen, 0, 0, g.screenWidth, g.screenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 160})

	// Overlay panel
	panelX, panelY := 80, 60
	panelW, panelH := g.screenWidth-160, g.screenHeight-120
	drawRect(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 35, G: 35, B: 50, A: 245})
	drawRectOutline(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 100, G: 100, B: 150, A: 255})

	ebitenutil.DebugPrintAt(screen, "QUEST LOG", panelX+panelW/2-30, panelY+10)

	// Close button (top-right of panel)
	closeBtnX := panelX + panelW - overlayCloseBtnW - 10
	drawRect(screen, closeBtnX, panelY+5, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 120, G: 50, B: 50, A: 255})
	drawRectOutline(screen, closeBtnX, panelY+5, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 200, G: 80, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "Close", closeBtnX+8, panelY+11)

	g.mu.RLock()
	ql := g.questLog
	sel := g.selectedQuest
	tab := g.questLogTab
	g.mu.RUnlock()

	if ql == nil {
		ebitenutil.DebugPrintAt(screen, "Loading...", panelX+20, panelY+40)
		return
	}

	// Tab bar: Active / Completed / Failed per §7
	tabLabels := []string{"Active", "Completed", "Failed"}
	for i, label := range tabLabels {
		tx := panelX + 20 + i*120
		tbg := color.RGBA{R: 50, G: 50, B: 70, A: 255}
		if i == tab {
			tbg = color.RGBA{R: 70, G: 60, B: 100, A: 255}
		}
		drawRect(screen, tx, panelY+30, 100, 22, tbg)
		drawRectOutline(screen, tx, panelY+30, 100, 22, color.RGBA{R: 90, G: 90, B: 130, A: 255})
		ebitenutil.DebugPrintAt(screen, label, tx+10, panelY+34)
	}

	y := panelY + 65

	var quests []QuestData
	switch tab {
	case 0:
		quests = ql.ActiveQuests
	case 1:
		quests = ql.CompletedQuests
	case 2:
		quests = ql.FailedQuests
	}

	if len(quests) == 0 {
		ebitenutil.DebugPrintAt(screen, "(no quests)", panelX+20, y)
	} else {
		for idx, q := range quests {
			if idx >= 8 {
				break
			}
			bgColor := color.RGBA{R: 40, G: 40, B: 60, A: 255}
			if idx == sel {
				bgColor = color.RGBA{R: 60, G: 50, B: 80, A: 255}
			}
			drawRect(screen, panelX+15, y, panelW-30, 40, bgColor)
			drawRectOutline(screen, panelX+15, y, panelW-30, 40, color.RGBA{R: 70, G: 70, B: 100, A: 255})

			marker := "  "
			if idx == sel {
				marker = "> "
			}
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%s", marker, q.Title), panelX+20, y+4)
			// Show objectives for active quests
			if tab == 0 {
				for i, obj := range q.Objectives {
					if i >= 2 {
						break
					}
					check := "[ ]"
					if obj.Completed {
						check = "[x]"
					}
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  %s %s (%d/%d)", check, truncateText(obj.Description, 40), obj.Progress, obj.Required), panelX+30, y+18+i*14)
				}
			}

			y += 44
		}
	}

	ebitenutil.DebugPrintAt(screen, "[J/Esc] Close  |  Tab: Switch  |  Up/Down: Navigate", panelX+panelW/2-140, panelY+panelH-25)
}

// ======================
// Guild Panel Overlay (§8)
// ======================

// updateGuildPanelOverlay handles input for the guild/faction overlay.
func (g *Game) updateGuildPanelOverlay() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.mu.Lock()
		g.overlays.ShowGuildPanel = false
		g.mu.Unlock()
		return
	}

	// Tab → switch sub-tab (Guild / Members / Factions)
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.mu.Lock()
		g.guildTab = (g.guildTab + 1) % 3
		g.mu.Unlock()
	}

	// Touch tap on guild tab bar and close button
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		panelX := 80
		panelY := 60
		panelW := g.screenWidth - 160

		// Close button (top-right of panel)
		closeBtnX := panelX + panelW - overlayCloseBtnW - 10
		if tx >= closeBtnX && tx <= closeBtnX+overlayCloseBtnW && ty >= panelY+5 && ty <= panelY+5+overlayCloseBtnH {
			g.mu.Lock()
			g.overlays.ShowGuildPanel = false
			g.mu.Unlock()
			return
		}

		// Tab bar
		for i := 0; i < 3; i++ {
			tabX := panelX + 20 + i*120
			if tx >= tabX && tx <= tabX+100 && ty >= panelY+10 && ty <= panelY+35 {
				g.mu.Lock()
				g.guildTab = i
				g.mu.Unlock()
				return
			}
		}
	}
}

// drawGuildPanelOverlay renders the guild/faction overlay (§8).
func (g *Game) drawGuildPanelOverlay(screen *ebiten.Image) {
	drawRect(screen, 0, 0, g.screenWidth, g.screenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 160})

	panelX, panelY := 80, 60
	panelW, panelH := g.screenWidth-160, g.screenHeight-120
	drawRect(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 35, G: 35, B: 50, A: 245})
	drawRectOutline(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 100, G: 100, B: 150, A: 255})

	// Close button (top-right of panel)
	closeBtnX := panelX + panelW - overlayCloseBtnW - 10
	drawRect(screen, closeBtnX, panelY+5, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 120, G: 50, B: 50, A: 255})
	drawRectOutline(screen, closeBtnX, panelY+5, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 200, G: 80, B: 80, A: 255})
	ebitenutil.DebugPrintAt(screen, "Close", closeBtnX+8, panelY+11)

	g.mu.RLock()
	tab := g.guildTab
	guild := g.guildData
	factions := g.factionRelations
	g.mu.RUnlock()

	// Tab bar
	tabs := []string{"Guild", "Members", "Factions"}
	for i, t := range tabs {
		x := panelX + 20 + i*120
		bgColor := color.RGBA{R: 50, G: 50, B: 70, A: 255}
		if i == tab {
			bgColor = color.RGBA{R: 70, G: 60, B: 100, A: 255}
		}
		drawRect(screen, x, panelY+10, 100, 25, bgColor)
		drawRectOutline(screen, x, panelY+10, 100, 25, color.RGBA{R: 90, G: 90, B: 130, A: 255})
		ebitenutil.DebugPrintAt(screen, t, x+20, panelY+16)
	}

	contentY := panelY + 50
	switch tab {
	case 0:
		g.drawGuildInfo(screen, panelX, contentY, panelW, guild)
	case 1:
		g.drawGuildMembers(screen, panelX, contentY, panelW, guild)
	case 2:
		g.drawFactionRelations(screen, panelX, contentY, panelW, factions)
	}

	ebitenutil.DebugPrintAt(screen, "[G/Esc] Close  |  Tab: Switch Panel", panelX+panelW/2-100, panelY+panelH-25)
}

func (g *Game) drawGuildInfo(screen *ebiten.Image, panelX, y, panelW int, guild *GuildData) {
	if guild == nil {
		ebitenutil.DebugPrintAt(screen, "Not in a guild. Join or create one!", panelX+20, y)
		return
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Guild: %s", guild.Name), panelX+20, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level: %d  |  Members: %d  |  Treasury: %d gold",
		guild.Level, guild.MemberCnt, guild.Treasury), panelX+20, y+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Leader: %s", guild.LeaderID), panelX+20, y+40)

	// Perks
	if len(guild.Perks) > 0 {
		ebitenutil.DebugPrintAt(screen, "Perks:", panelX+20, y+70)
		for i, perk := range guild.Perks {
			if i >= 5 {
				break
			}
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  - %s (Req Lv%d): %s",
				perk.Name, perk.LevelReq, perk.Description), panelX+30, y+85+i*15)
		}
	}
}

func (g *Game) drawGuildMembers(screen *ebiten.Image, panelX, y, panelW int, guild *GuildData) {
	if guild == nil {
		ebitenutil.DebugPrintAt(screen, "Not in a guild", panelX+20, y)
		return
	}
	ebitenutil.DebugPrintAt(screen, "MEMBERS", panelX+20, y)
	y += 25
	for i, m := range guild.Members {
		if i >= 12 {
			break
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  %-15s  %-10s  Contrib: %d",
			m.Name, m.RankName, m.Contribution), panelX+20, y)
		y += 18
	}
}

func (g *Game) drawFactionRelations(screen *ebiten.Image, panelX, y, panelW int, factions []FactionRelation) {
	ebitenutil.DebugPrintAt(screen, "FACTION RELATIONS", panelX+20, y)
	y += 25

	if len(factions) == 0 {
		ebitenutil.DebugPrintAt(screen, "No known factions", panelX+20, y)
		return
	}

	for i, f := range factions {
		if i >= 10 {
			break
		}
		statusColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
		switch f.State {
		case "allied":
			statusColor = color.RGBA{R: 80, G: 200, B: 80, A: 255}
		case "war":
			statusColor = color.RGBA{R: 200, G: 80, B: 80, A: 255}
		case "peace":
			statusColor = color.RGBA{R: 200, G: 200, B: 80, A: 255}
		}
		_ = statusColor

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  %-15s  Opinion: %d  State: %s",
			f.FactionName, f.Opinion, f.State), panelX+20, y)
		y += 18
	}
}

// ======================
// Settings Overlay (§12)
// ======================

// updateSettingsOverlay handles input for the settings/accessibility overlay.
func (g *Game) updateSettingsOverlay() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.Lock()
		g.overlays.ShowSettings = false
		g.mu.Unlock()
		return
	}

	// Touch tap on close area (Esc close text at bottom)
	if tapped, tx, ty := g.touchState.HasTap(); tapped {
		panelX := 150
		panelY := 100
		panelW := g.screenWidth - 300
		panelH := g.screenHeight - 200
		// Tap outside the settings panel to close
		if tx < panelX || tx > panelX+panelW || ty < panelY || ty > panelY+panelH {
			g.mu.Lock()
			g.overlays.ShowSettings = false
			g.mu.Unlock()
		}
	}
}

// drawSettingsOverlay renders the settings/accessibility overlay (§12).
func (g *Game) drawSettingsOverlay(screen *ebiten.Image) {
	drawRect(screen, 0, 0, g.screenWidth, g.screenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 160})

	panelX, panelY := 150, 100
	panelW, panelH := g.screenWidth-300, g.screenHeight-200
	drawRect(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 40, G: 40, B: 55, A: 245})
	drawRectOutline(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 100, G: 100, B: 150, A: 255})

	ebitenutil.DebugPrintAt(screen, "SETTINGS", panelX+panelW/2-25, panelY+15)

	y := panelY + 50
	settings := []string{
		"[1] Audio Volume:        ████░░ 60%",
		"[2] Music Volume:        ███░░░ 50%",
		"[3] Font Size:           Normal",
		"[4] High Contrast:       Off",
		"[5] Screen Reader:       Off",
		"[6] Auto-Save Interval:  5 min",
		"[7] Key Bindings:        Default",
	}

	for _, s := range settings {
		ebitenutil.DebugPrintAt(screen, s, panelX+40, y)
		y += 30
	}

	ebitenutil.DebugPrintAt(screen, "[Esc] Close", panelX+panelW/2-30, panelY+panelH-30)
}
