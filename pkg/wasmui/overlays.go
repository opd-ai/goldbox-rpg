//go:build js && wasm

package wasmui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

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

// drawLoadingIndicator draws an animated "Loading..." text.
// Item 22: Loading State Indicators - visual feedback during async data fetches.
func drawLoadingIndicator(screen *ebiten.Image, x, y int, label string) {
	// Animate dots: cycles through "", ".", "..", "..." every 300ms
	dots := int(time.Now().UnixMilli()/300) % 4
	dotStr := strings.Repeat(".", dots)
	text := fmt.Sprintf("%s%s", label, dotStr)
	drawColoredText(screen, text, x, y, ColorGold)
}

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

	// Handle list selection via keyboard, touch swipe, and mouse wheel
	if delta := g.selectionDelta(sel, len(items)); delta != 0 {
		g.mu.Lock()
		g.selectedItem += delta
		g.mu.Unlock()
		sel += delta
	}

	// Handle touch on action buttons
	if g.inventoryTouchButtons(items, sel) {
		return
	}

	// Handle touch on list items
	if newSel := g.inventoryTouchListSelect(); newSel >= 0 && newSel < len(items) {
		g.mu.Lock()
		g.selectedItem = newSel
		g.mu.Unlock()
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
			_, err := g.rpcClient.UnequipItem(item.Slot) // fix: server expects slot name, not item ID
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

	g.drawInventoryHeader(screen)

	g.mu.RLock()
	items := g.inventoryItems
	sel := g.selectedItem
	player := g.player
	loadingInv := g.loadingInv
	g.mu.RUnlock()

	g.drawEquipmentSlots(screen, 30, 50, player)

	if loadingInv {
		drawLoadingIndicator(screen, 410, 125, "Loading inventory")
		return
	}

	g.drawInventoryList(screen, items, sel)
	g.drawInventoryFooter(screen, items, sel)
}

// drawInventoryHeader renders the title and close button.
func (g *Game) drawInventoryHeader(screen *ebiten.Image) {
	drawColoredText(screen, "INVENTORY & EQUIPMENT", 280, 15, ColorGold)
	closeBtnX := ScreenWidth - overlayCloseBtnW - 10
	drawRect(screen, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 120, G: 50, B: 50, A: 255})
	drawRectOutline(screen, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 200, G: 80, B: 80, A: 255})
	drawColoredText(screen, "Close", closeBtnX+8, 16, ColorStatValue)
	drawColoredText(screen, "[I/Esc] Close  |  Enter: Equip/Unequip  |  U: Use", 170, 560, ColorStatLabel)
}

// drawInventoryList renders the scrollable item list.
func (g *Game) drawInventoryList(screen *ebiten.Image, items []InventoryItem, sel int) {
	listX, listY := 350, 50
	drawColoredText(screen, "INVENTORY", listX+80, listY, ColorGold)
	listY += 25

	if len(items) == 0 {
		drawColoredText(screen, "(empty)", listX+80, listY, ColorStatLabel)
		return
	}

	for i, item := range items {
		if i >= 15 {
			break
		}
		g.drawInventoryItem(screen, item, i, sel, listX, listY+i*32)
	}
}

// drawInventoryItem renders a single inventory item row.
func (g *Game) drawInventoryItem(screen *ebiten.Image, item InventoryItem, idx, sel, x, y int) {
	bgColor := color.RGBA{R: 40, G: 40, B: 55, A: 255}
	if idx == sel {
		bgColor = color.RGBA{R: 60, G: 50, B: 80, A: 255}
	}
	drawRect(screen, x, y, 410, 30, bgColor)
	drawRectOutline(screen, x, y, 410, 30, color.RGBA{R: 80, G: 80, B: 100, A: 255})

	iconPath := ItemIconPath(item.Type, item.Name)
	DrawSpriteWithFallback(screen, iconPath, x+4, y+3, 24, 24, getItemFallbackColor(item.Type))

	equipTag, marker, nameClr := "", "  ", ColorStatValue
	if item.Equipped {
		equipTag = "[E] "
	}
	if idx == sel {
		marker, nameClr = "> ", ColorGoldHi
	}
	drawColoredText(screen, fmt.Sprintf("%s%s%s", marker, equipTag, item.Name), x+34, y+8, nameClr)
	drawColoredText(screen, item.Type, x+310, y+8, ColorStatLabel)
}

// drawInventoryFooter renders item detail panel and action buttons.
func (g *Game) drawInventoryFooter(screen *ebiten.Image, items []InventoryItem, sel int) {
	if sel < len(items) && len(items) > 0 {
		g.drawItemDetail(screen, items[sel])
	}

	equipLabel := "Equip"
	if sel < len(items) && len(items) > 0 && items[sel].Equipped {
		equipLabel = "Unequip"
	}
	drawRect(screen, 350, 540, 120, 28, color.RGBA{R: 50, G: 80, B: 50, A: 255})
	drawRectOutline(screen, 350, 540, 120, 28, color.RGBA{R: 80, G: 140, B: 80, A: 255})
	drawColoredText(screen, equipLabel, 370, 546, ColorStatValue)

	drawRect(screen, 490, 540, 100, 28, color.RGBA{R: 50, G: 50, B: 80, A: 255})
	drawRectOutline(screen, 490, 540, 100, 28, color.RGBA{R: 80, G: 80, B: 140, A: 255})
	drawColoredText(screen, "Use", 520, 546, ColorStatValue)
}

// getItemFallbackColor returns the fallback color for item icons based on item type.
func getItemFallbackColor(itemType string) color.RGBA {
	switch strings.ToLower(itemType) {
	case "weapon", "weapons", "sword", "axe", "mace", "dagger":
		return color.RGBA{R: 128, G: 128, B: 128, A: 255} // Gray for weapons
	case "armor", "armour", "chest", "helm", "shield":
		return color.RGBA{R: 80, G: 100, B: 180, A: 255} // Blue for armor
	case "consumable", "potion", "scroll", "food":
		return color.RGBA{R: 80, G: 180, B: 80, A: 255} // Green for consumables
	case "ring", "amulet", "jewelry", "accessory":
		return color.RGBA{R: 180, G: 150, B: 80, A: 255} // Gold for jewelry
	case "wand", "staff", "magic":
		return color.RGBA{R: 150, G: 80, B: 180, A: 255} // Purple for magic items
	default:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255} // Default gray
	}
}

// drawEquipmentSlots renders the character equipment slots.
func (g *Game) drawEquipmentSlots(screen *ebiten.Image, x, y int, player *PlayerState) {
	drawColoredText(screen, "EQUIPPED", x+80, y, ColorGold)
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
		drawColoredText(screen, fmt.Sprintf("%-10s: %s", slot, itemName), x+5, sy+6, ColorStatValue)
	}
}

// drawItemDetail renders the selected item's details.
func (g *Game) drawItemDetail(screen *ebiten.Image, item ItemData) {
	y := 530
	drawColoredText(screen, fmt.Sprintf("%s  |  %s  |  Slot: %s  |  Weight: %d",
		item.Name, item.Type, item.Slot, item.Weight), 30, y, ColorStatLabel)
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

	// Use the filtered spell list for all navigation bounds
	filtered := g.filteredSpells()
	filteredLen := len(filtered)

	// Handle list selection via keyboard, touch swipe, and mouse wheel
	if delta := g.selectionDelta(sel, filteredLen); delta != 0 {
		g.mu.Lock()
		g.selectedSpell += delta
		g.mu.Unlock()
		sel += delta
	}

	// Handle touch on action buttons
	if g.spellbookTouchButtons(sel) {
		return
	}

	// Handle touch on list items
	if newSel := g.spellbookTouchListSelect(filteredLen); newSel >= 0 {
		g.mu.Lock()
		g.selectedSpell = newSel
		g.mu.Unlock()
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
// If the spell requires targeting (single or area), enters targeting mode instead.
func (g *Game) castSelectedSpell(sel int) {
	filtered := g.filteredSpells()
	if sel >= len(filtered) {
		return
	}
	spell := filtered[sel]

	// Check if spell requires targeting
	switch spell.TargetType {
	case "single", "area", "cone":
		// Enter targeting mode
		g.mu.Lock()
		g.pendingSpell = &spell
		g.spellTargetMode = true
		g.spellTargetPos = Position{X: 5, Y: 5} // Default center
		g.spellTargetID = ""
		g.mu.Unlock()

		g.addLogMessage(fmt.Sprintf("Select target for %s...", spell.Name), MessageInfo)
		g.closeSpellbook()
		return
	}

	// Self-targeting or no targeting required - cast immediately
	g.executeCastSpell(&spell, "", nil)
}

// executeCastSpell sends the spell cast RPC and handles the result.
func (g *Game) executeCastSpell(spell *SpellData, targetID string, targetPos *Position) {
	spellSchool := SpellSchoolName(spell.School)
	go func() {
		result, err := g.rpcClient.CastSpell(spell.ID, targetID, targetPos)
		if err != nil {
			g.addLogMessage(fmt.Sprintf("Casting %s... FAILED", spell.Name), MessageError)
			g.showError(fmt.Sprintf("Cast failed: %v", err))
		} else {
			// Rich Gold Box-style spell narration
			g.addLogMessage(fmt.Sprintf("Cast %s!", spell.Name), MessageCombat)

			// Add visual spell effect at target position
			effectPos := Position{X: 5, Y: 5} // Default
			if targetPos != nil {
				effectPos = *targetPos
			}
			g.addSpellEffect(spell.ID, spellSchool, effectPos)

			if result != nil {
				if result.Damage > 0 {
					g.addLogMessage(fmt.Sprintf("  %s deals %d damage!", spell.Name, result.Damage), MessageCombat)
				}
				if result.Healing > 0 {
					g.addLogMessage(fmt.Sprintf("  %s heals %d HP!", spell.Name, result.Healing), MessageCombat)
				}
				if result.Message != "" {
					g.addLogMessage(fmt.Sprintf("  %s", result.Message), MessageInfo)
				}
			}
		}
	}()
}

// cancelSpellTargeting exits spell targeting mode without casting.
func (g *Game) cancelSpellTargeting() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.pendingSpell != nil {
		g.addLogMessage(fmt.Sprintf("Cancelled %s", g.pendingSpell.Name), MessageInfo)
	}
	g.pendingSpell = nil
	g.spellTargetMode = false
	g.spellTargetPos = Position{}
	g.spellTargetID = ""
}

// confirmSpellTarget casts the pending spell on the selected target.
func (g *Game) confirmSpellTarget() {
	g.mu.Lock()
	spell := g.pendingSpell
	targetID := g.spellTargetID
	targetPos := g.spellTargetPos
	g.pendingSpell = nil
	g.spellTargetMode = false
	g.mu.Unlock()

	if spell == nil {
		return
	}

	// For area spells, use position; for single target, use ID
	var pos *Position
	if spell.TargetType == "area" || spell.TargetType == "cone" {
		pos = &targetPos
	}
	g.executeCastSpell(spell, targetID, pos)
}

// drawSpellbookScreen renders the spellbook panel (§3.8).
func (g *Game) drawSpellbookScreen(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 25, G: 25, B: 45, A: 255})

	drawColoredText(screen, "SPELLBOOK", 340, 15, ColorGold)

	// Close button (top-right)
	closeBtnX := ScreenWidth - overlayCloseBtnW - 10
	drawRect(screen, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 120, G: 50, B: 50, A: 255})
	drawRectOutline(screen, closeBtnX, 10, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 200, G: 80, B: 80, A: 255})
	drawColoredText(screen, "Close", closeBtnX+8, 16, ColorStatValue)

	g.mu.RLock()
	filter := g.spellFilter
	sel := g.selectedSpell
	loadingSpells := g.loadingSpells
	g.mu.RUnlock()

	// Filter indicator
	filterText := "All Levels"
	if filter >= 0 {
		filterText = fmt.Sprintf("Level %d", filter)
	}
	drawColoredText(screen, fmt.Sprintf("Filter: %s  [Tab to change]", filterText), 50, 45, ColorStatLabel)

	// Show loading indicator if fetching data
	if loadingSpells {
		drawLoadingIndicator(screen, 300, 250, "Loading spells")
		return
	}

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
		spellClr := ColorStatValue
		if i == sel {
			marker = "> "
			spellClr = ColorGoldHi
		}
		schoolName := SpellSchoolName(spell.School)
		drawColoredText(screen, fmt.Sprintf("%sLv%d %-20s %s", marker, spell.Level, spell.Name, schoolName), 55, y+4, spellClr)
	}

	// Spell detail
	if sel < len(filtered) && len(filtered) > 0 {
		g.drawSpellDetail(screen, filtered[sel])
	}

	// Touch action buttons
	drawRect(screen, 200, 555, 120, 28, color.RGBA{R: 50, G: 80, B: 50, A: 255})
	drawRectOutline(screen, 200, 555, 120, 28, color.RGBA{R: 80, G: 140, B: 80, A: 255})
	drawColoredText(screen, "Cast", 240, 561, ColorStatValue)

	drawRect(screen, 480, 555, 120, 28, color.RGBA{R: 50, G: 50, B: 80, A: 255})
	drawRectOutline(screen, 480, 555, 120, 28, color.RGBA{R: 80, G: 80, B: 140, A: 255})
	drawColoredText(screen, "Filter", 515, 561, ColorStatValue)

	drawColoredText(screen, "[Esc] Close  |  Enter: Cast  |  Tab: Filter Level", 200, 585, ColorStatLabel)
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
	drawColoredText(screen, fmt.Sprintf("%s  |  Lv%d  |  %s  |  Range: %s  |  %s",
		spell.Name, spell.Level, SpellSchoolName(spell.School), spell.Range, spell.Description), 50, y, ColorStatLabel)
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

	total := g.questLogTotalForTab(ql, tab)

	// Handle list selection via keyboard, touch swipe, and mouse wheel
	if delta := g.selectionDelta(sel, total); delta != 0 {
		g.mu.Lock()
		g.selectedQuest += delta
		g.mu.Unlock()
	}

	// Handle touch on action buttons (close, tabs)
	panelX := 80
	panelY := 60
	panelW := g.screenWidth - 160
	if g.questLogTouchButtons(panelX, panelY, panelW) {
		return
	}

	// Handle touch on quest list items
	if newSel := g.questLogTouchListSelect(panelX, panelY, panelW, total); newSel >= 0 {
		g.mu.Lock()
		g.selectedQuest = newSel
		g.mu.Unlock()
	}
}

// questLogTotalForTab returns the number of quests in the given tab.
func (g *Game) questLogTotalForTab(ql *QuestLogResult, tab int) int {
	switch tab {
	case 0:
		return len(ql.ActiveQuests)
	case 1:
		return len(ql.CompletedQuests)
	case 2:
		return len(ql.FailedQuests)
	}
	return 0
}

// questLogGetQuestsForTab returns the quests for the currently selected tab.
func questLogGetQuestsForTab(ql *QuestLogResult, tab int) []QuestData {
	switch tab {
	case 0:
		return ql.ActiveQuests
	case 1:
		return ql.CompletedQuests
	case 2:
		return ql.FailedQuests
	default:
		return nil
	}
}

// drawQuestLogTabs renders the tab bar for the quest log.
func drawQuestLogTabs(screen *ebiten.Image, panelX, panelY, selectedTab int) {
	tabLabels := []string{"Active", "Completed", "Failed"}
	for i, label := range tabLabels {
		tx := panelX + 20 + i*120
		tbg := color.RGBA{R: 50, G: 50, B: 70, A: 255}
		tabClr := ColorStatValue
		if i == selectedTab {
			tbg = color.RGBA{R: 70, G: 60, B: 100, A: 255}
			tabClr = ColorGoldHi
		}
		drawRect(screen, tx, panelY+30, 100, 22, tbg)
		drawRectOutline(screen, tx, panelY+30, 100, 22, color.RGBA{R: 90, G: 90, B: 130, A: 255})
		drawColoredText(screen, label, tx+10, panelY+34, tabClr)
	}
}

// drawQuestLogItem renders a single quest list item.
func drawQuestLogItem(screen *ebiten.Image, q QuestData, idx, sel, panelX, panelW, y, tab int) int {
	bgColor := color.RGBA{R: 40, G: 40, B: 60, A: 255}
	if idx == sel {
		bgColor = color.RGBA{R: 60, G: 50, B: 80, A: 255}
	}
	drawRect(screen, panelX+15, y, panelW-30, 40, bgColor)
	drawRectOutline(screen, panelX+15, y, panelW-30, 40, color.RGBA{R: 70, G: 70, B: 100, A: 255})

	marker := "  "
	titleClr := ColorStatValue
	if idx == sel {
		marker = "> "
		titleClr = ColorGoldHi
	}
	drawColoredText(screen, fmt.Sprintf("%s%s", marker, q.Title), panelX+20, y+4, titleClr)

	// Show objectives for active quests
	if tab == 0 {
		for i, obj := range q.Objectives {
			if i >= 2 {
				break
			}
			check := "[ ]"
			objClr := ColorStatLabel
			if obj.Completed {
				check = "[x]"
				objClr = ColorPlayerName
			}
			drawColoredText(screen, fmt.Sprintf("  %s %s (%d/%d)", check, truncateText(obj.Description, 40), obj.Progress, obj.Required), panelX+30, y+18+i*14, objClr)
		}
	}
	return y + 44
}

// drawQuestLogOverlay renders the quest log overlay (§7).
func (g *Game) drawQuestLogOverlay(screen *ebiten.Image) {
	// Dim background
	drawRect(screen, 0, 0, g.screenWidth, g.screenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 160})

	// Overlay panel
	panelX, panelY := 80, 60
	panelW, panelH := g.screenWidth-160, g.screenHeight-120
	drawRect(screen, panelX, panelY, panelW, panelH, color.RGBA{R: 35, G: 35, B: 50, A: 245})
	drawRectOutline(screen, panelX, panelY, panelW, panelH, ColorPanelBorder)
	drawColoredText(screen, "QUEST LOG", panelX+panelW/2-30, panelY+10, ColorGold)

	// Close button (top-right of panel)
	closeBtnX := panelX + panelW - overlayCloseBtnW - 10
	drawRect(screen, closeBtnX, panelY+5, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 120, G: 50, B: 50, A: 255})
	drawRectOutline(screen, closeBtnX, panelY+5, overlayCloseBtnW, overlayCloseBtnH, color.RGBA{R: 200, G: 80, B: 80, A: 255})
	drawColoredText(screen, "Close", closeBtnX+8, panelY+11, ColorStatValue)

	g.mu.RLock()
	ql := g.questLog
	sel := g.selectedQuest
	tab := g.questLogTab
	loadingQL := g.loadingQuestLog
	g.mu.RUnlock()

	// Show animated loading indicator
	if ql == nil || loadingQL {
		drawLoadingIndicator(screen, panelX+panelW/2-60, panelY+100, "Loading quests")
		return
	}

	drawQuestLogTabs(screen, panelX, panelY, tab)

	y := panelY + 65
	quests := questLogGetQuestsForTab(ql, tab)

	if len(quests) == 0 {
		drawColoredText(screen, "(no quests)", panelX+20, y, ColorStatLabel)
	} else {
		for idx, q := range quests {
			if idx >= 8 {
				break
			}
			y = drawQuestLogItem(screen, q, idx, sel, panelX, panelW, y, tab)
		}
	}

	drawColoredText(screen, "[J/Esc] Close  |  Tab: Switch  |  Up/Down: Navigate", panelX+panelW/2-140, panelY+panelH-25, ColorStatLabel)
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

	// Handle mouse click for treasury buttons
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		g.handleGuildTreasuryClick(mx, my)
	}

	// Touch tap on guild tab bar, close button, and treasury buttons
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

		// Treasury buttons (only on Guild tab)
		g.handleGuildTreasuryClick(tx, ty)
	}
}

// handleGuildTreasuryClick processes clicks on treasury buttons.
func (g *Game) handleGuildTreasuryClick(x, y int) {
	g.mu.RLock()
	tab := g.guildTab
	guild := g.guildData
	g.mu.RUnlock()

	// Only handle on Guild tab with valid guild
	if tab != 0 || guild == nil {
		return
	}

	panelX := 80
	panelY := 60
	contentY := panelY + 50
	treasuryY := contentY + 70
	amountY := treasuryY + 25

	// Amount adjustment buttons
	// -10 button
	if x >= panelX+150 && x <= panelX+190 && y >= amountY-3 && y <= amountY+17 {
		g.mu.Lock()
		g.treasuryAmount = max(10, g.treasuryAmount-10)
		g.mu.Unlock()
		return
	}
	// +10 button
	if x >= panelX+195 && x <= panelX+235 && y >= amountY-3 && y <= amountY+17 {
		g.mu.Lock()
		g.treasuryAmount += 10
		g.mu.Unlock()
		return
	}
	// +100 button
	if x >= panelX+240 && x <= panelX+290 && y >= amountY-3 && y <= amountY+17 {
		g.mu.Lock()
		g.treasuryAmount += 100
		g.mu.Unlock()
		return
	}

	buttonY := amountY + 30
	// Deposit button
	if x >= panelX+20 && x <= panelX+100 && y >= buttonY && y <= buttonY+25 {
		g.doGuildDeposit()
		return
	}
	// Withdraw button
	if x >= panelX+110 && x <= panelX+190 && y >= buttonY && y <= buttonY+25 {
		g.doGuildWithdraw()
		return
	}
}

// doGuildDeposit deposits gold into the guild treasury.
func (g *Game) doGuildDeposit() {
	g.mu.RLock()
	guild := g.guildData
	amount := g.treasuryAmount
	g.mu.RUnlock()

	if guild == nil {
		return
	}

	go func() {
		result, err := g.rpcClient.GuildDeposit(guild.ID, amount)
		if err != nil {
			g.showError(fmt.Sprintf("Deposit failed: %v", err))
			return
		}
		if result != nil && result.Success {
			g.addLogMessage(fmt.Sprintf("Deposited %d gold to guild treasury", amount), MessageInfo)
			// Refresh guild data
			g.refreshGuildData()
		}
	}()
}

// doGuildWithdraw withdraws gold from the guild treasury.
func (g *Game) doGuildWithdraw() {
	g.mu.RLock()
	guild := g.guildData
	amount := g.treasuryAmount
	g.mu.RUnlock()

	if guild == nil {
		return
	}

	go func() {
		result, err := g.rpcClient.GuildWithdraw(guild.ID, amount)
		if err != nil {
			g.showError(fmt.Sprintf("Withdraw failed: %v", err))
			return
		}
		if result != nil && result.Success {
			g.addLogMessage(fmt.Sprintf("Withdrew %d gold from guild treasury", amount), MessageInfo)
			// Refresh guild data
			g.refreshGuildData()
		}
	}()
}

// refreshGuildData fetches updated guild data from server.
func (g *Game) refreshGuildData() {
	guild, err := g.rpcClient.GetCharacterGuild()
	if err == nil && guild != nil {
		g.mu.Lock()
		g.guildData = guild
		g.mu.Unlock()
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

	// Treasury controls (Step 14: Guild Treasury UI)
	g.mu.RLock()
	amount := g.treasuryAmount
	g.mu.RUnlock()

	treasuryY := y + 70
	drawColoredText(screen, "TREASURY OPERATIONS", panelX+20, treasuryY, ColorGold)

	// Amount selector
	amountY := treasuryY + 25
	drawColoredText(screen, fmt.Sprintf("Amount: %d gold", amount), panelX+20, amountY, ColorStatValue)

	// Amount adjustment buttons
	g.drawTreasuryButton(screen, "-10", panelX+150, amountY-3, 40, 20)
	g.drawTreasuryButton(screen, "+10", panelX+195, amountY-3, 40, 20)
	g.drawTreasuryButton(screen, "+100", panelX+240, amountY-3, 50, 20)

	// Deposit/Withdraw buttons
	buttonY := amountY + 30
	g.drawTreasuryButton(screen, "Deposit", panelX+20, buttonY, 80, 25)
	g.drawTreasuryButton(screen, "Withdraw", panelX+110, buttonY, 80, 25)

	// Perks section (moved down)
	perksY := buttonY + 45
	if len(guild.Perks) > 0 {
		ebitenutil.DebugPrintAt(screen, "Perks:", panelX+20, perksY)
		for i, perk := range guild.Perks {
			if i >= 5 {
				break
			}
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  - %s (Req Lv%d): %s",
				perk.Name, perk.LevelReq, perk.Description), panelX+30, perksY+15+i*15)
		}
	}
}

// drawTreasuryButton draws a styled button for treasury operations.
func (g *Game) drawTreasuryButton(screen *ebiten.Image, label string, x, y, w, h int) {
	bgColor := color.RGBA{R: 60, G: 50, B: 80, A: 255}
	borderColor := color.RGBA{R: 120, G: 100, B: 150, A: 255}
	drawRect(screen, x, y, w, h, bgColor)
	drawRectOutline(screen, x, y, w, h, borderColor)
	// Center text in button
	textX := x + (w-len(label)*6)/2
	textY := y + (h-12)/2
	drawColoredText(screen, label, textX, textY, ColorStatValue)
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
	drawColoredText(screen, "FACTION RELATIONS", panelX+20, y, ColorGold)
	y += 25

	if len(factions) == 0 {
		ebitenutil.DebugPrintAt(screen, "No known factions", panelX+20, y)
		return
	}

	for i, f := range factions {
		if i >= 8 { // Limit to fit panel
			break
		}
		// Draw faction name
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%-12s", f.FactionName), panelX+20, y)

		// Draw state with color coding
		stateColor := getStateColor(f.State)
		drawColoredText(screen, f.State, panelX+110, y, stateColor)

		// Draw opinion bar (100px wide, centered at 0)
		barX := panelX + 180
		barW := 100
		barH := 8
		drawOpinionBar(screen, barX, y+2, barW, barH, f.Opinion)

		// Draw trust bar below opinion (optional field - check if non-zero)
		if f.Trust != 0 {
			drawTrustBar(screen, barX, y+12, barW, barH, f.Trust)
		}

		y += 25
	}
}

// getStateColor returns the appropriate color for a diplomatic state.
func getStateColor(state string) color.RGBA {
	switch state {
	case "war", "War":
		return ColorEnemyName // Red
	case "hostile", "Hostile":
		return color.RGBA{R: 255, G: 150, B: 50, A: 255} // Orange
	case "neutral", "Neutral":
		return color.RGBA{R: 150, G: 150, B: 150, A: 255} // Gray
	case "friendly", "Friendly":
		return color.RGBA{R: 80, G: 200, B: 80, A: 255} // Green
	case "allied", "Allied":
		return ColorGold // Gold
	case "peace", "Peace":
		return color.RGBA{R: 200, G: 200, B: 80, A: 255} // Yellow
	default:
		return ColorStatLabel
	}
}

// drawOpinionBar draws a horizontal bar for faction opinion (-100 to +100).
// Negative values fill red from center-left, positive fill green from center-right.
func drawOpinionBar(screen *ebiten.Image, x, y, w, h, opinion int) {
	// Background bar
	drawRect(screen, x, y, w, h, color.RGBA{R: 40, G: 40, B: 40, A: 200})

	// Center marker
	centerX := x + w/2
	drawRect(screen, centerX-1, y, 2, h, color.RGBA{R: 100, G: 100, B: 100, A: 255})

	// Clamp opinion to -100 to +100
	if opinion > 100 {
		opinion = 100
	} else if opinion < -100 {
		opinion = -100
	}

	// Calculate fill width (half bar = 100 units)
	fillW := (opinion * w / 2) / 100
	if fillW < 0 {
		// Negative: fill from center-left
		fillColor := color.RGBA{R: 200, G: 60, B: 60, A: 255}
		drawRect(screen, centerX+fillW, y+1, -fillW, h-2, fillColor)
	} else if fillW > 0 {
		// Positive: fill from center-right
		fillColor := color.RGBA{R: 60, G: 200, B: 60, A: 255}
		drawRect(screen, centerX, y+1, fillW, h-2, fillColor)
	}
}

// drawTrustBar draws a horizontal bar for faction trust (0 to 100).
func drawTrustBar(screen *ebiten.Image, x, y, w, h, trust int) {
	// Background bar
	drawRect(screen, x, y, w, h, color.RGBA{R: 40, G: 40, B: 40, A: 200})

	// Clamp trust to 0 to 100
	if trust > 100 {
		trust = 100
	} else if trust < 0 {
		trust = 0
	}

	// Calculate fill width
	fillW := trust * w / 100
	if fillW > 0 {
		// Blue for trust
		fillColor := color.RGBA{R: 60, G: 120, B: 200, A: 255}
		drawRect(screen, x+1, y+1, fillW-2, h-2, fillColor)
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
