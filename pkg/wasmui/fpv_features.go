//go:build js && wasm

package wasmui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawFeatureAtDepth renders architectural features from VisibleTile.Feature at the given depth.
// Only center tile (relX=0) features are rendered. Depth controls detail level.
func drawFeatureAtDepth(screen *ebiten.Image, p *fpvParams, depth int) {
	ct := p.getTile(0, depth)
	if ct == nil || ct.Feature == "" {
		return
	}
	// Don't draw features on walls (features appear on floor tiles)
	if ct.TileType == "wall" {
		return
	}

	// Get the rendering area for this depth
	var fx, fy, fw, fh int
	switch depth {
	case 0:
		fx = p.vpX + p.nearInset
		fw = p.vpWidth - 2*p.nearInset
		fy = p.nearTop
		fh = p.nearBottom - p.nearTop
	case 1:
		fx = p.vpX + p.midInset
		fw = p.vpWidth - 2*p.midInset
		fy = p.midTop
		fh = p.midBottom - p.midTop
	case 2:
		fx = p.vpX + p.farInset
		fw = p.vpWidth - 2*p.farInset
		fy = p.farTop
		fh = p.farBottom - p.farTop
	default:
		return
	}

	if fw <= 4 || fh <= 4 {
		return
	}

	switch ct.Feature {
	case "pillar":
		drawPillar(screen, fx, fy, fw, fh, p.palette, depth)
	case "altar":
		drawAltar(screen, fx, fy, fw, fh, p.palette, depth)
	case "fountain":
		drawFountain(screen, fx, fy, fw, fh, p.palette, depth)
	case "archway":
		drawArchway(screen, fx, fy, fw, fh, p.palette, depth)
	case "alcove":
		drawAlcove(screen, fx, fy, fw, fh, p.palette, depth)
	case "rubble":
		drawRubblePile(screen, fx, fy, fw, fh, p.palette, depth)
	}

	// Room-type-specific overlays
	drawRoomTypeOverlay(screen, ct.RoomType, fx, fy, fw, fh, p.palette, depth)
}

// drawPillar draws a vertical column with capital and base blocks.
func drawPillar(screen *ebiten.Image, x, y, w, h int, palette fpvThemePalette, depth int) {
	cx := x + w/2
	floorY := y + h

	var pillarW, capH int
	switch depth {
	case 0: // Near — full detail
		pillarW = max(6, w/10)
		capH = max(4, h/12)
	case 1: // Mid — simplified
		pillarW = max(4, w/12)
		capH = max(2, h/14)
	default: // Far — hint only
		pillarW = max(2, w/16)
		drawRect(screen, cx-pillarW/2, y+h/4, pillarW, h/2, brightenColor(palette.wallColorFar, 10))
		return
	}

	pillarColor := brightenColor(palette.wallColorNear, 15)
	capColor := brightenColor(palette.wallColorNear, 25)
	baseColor := color.RGBA{
		R: uint8(max(0, int(palette.wallColorNear.R)-10)),
		G: uint8(max(0, int(palette.wallColorNear.G)-10)),
		B: uint8(max(0, int(palette.wallColorNear.B)-8)),
		A: 255,
	}

	pillarH := h - 2*capH
	py := y + capH

	// Capital block (top)
	drawRect(screen, cx-pillarW/2-2, y, pillarW+4, capH, capColor)
	// Shaft
	drawRect(screen, cx-pillarW/2, py, pillarW, pillarH, pillarColor)
	// Base block (bottom)
	drawRect(screen, cx-pillarW/2-2, floorY-capH, pillarW+4, capH, baseColor)

	// Edge highlight on shaft (near only)
	if depth == 0 {
		drawRect(screen, cx-pillarW/2, py, 1, pillarH, brightenColor(pillarColor, 12))
	}
}

// drawAltar draws a low rectangular block on the floor with a highlight edge.
func drawAltar(screen *ebiten.Image, x, y, w, h int, palette fpvThemePalette, depth int) {
	floorY := y + h

	if depth >= 2 {
		// Far — small rectangle hint
		aw := max(4, w/5)
		ah := max(2, h/8)
		drawRect(screen, x+w/2-aw/2, floorY-ah, aw, ah, brightenColor(palette.wallColorFar, 10))
		return
	}

	var altarW, altarH int
	if depth == 0 {
		altarW = max(12, w/4)
		altarH = max(8, h/5)
	} else {
		altarW = max(8, w/5)
		altarH = max(4, h/6)
	}

	ax := x + w/2 - altarW/2
	ay := floorY - altarH

	stoneColor := brightenColor(palette.wallColorNear, 20)
	drawRect(screen, ax, ay, altarW, altarH, stoneColor)
	// Top highlight edge
	drawLine(screen, ax, ay, ax+altarW, ay, brightenColor(stoneColor, 15))
	// Dark base
	drawRect(screen, ax+1, floorY-2, altarW-2, 2, color.RGBA{
		R: uint8(max(0, int(stoneColor.R)-30)),
		G: uint8(max(0, int(stoneColor.G)-30)),
		B: uint8(max(0, int(stoneColor.B)-25)),
		A: 255,
	})

	// Candle on altar (near only)
	if depth == 0 {
		candleColor := color.RGBA{R: 200, G: 190, B: 160, A: 255}
		drawRect(screen, ax+altarW/2-1, ay-6, 2, 6, candleColor)
		frame := flickerFrame()
		flameColor := flickerFlameColor(palette.torchFlame, frame)
		drawRect(screen, ax+altarW/2-1, ay-9, 2, 3, flameColor)
	}
}

// drawFountain draws a basin shape with water-colored fill at near depth.
func drawFountain(screen *ebiten.Image, x, y, w, h int, palette fpvThemePalette, depth int) {
	floorY := y + h

	if depth >= 2 {
		bw := max(4, w/6)
		drawRect(screen, x+w/2-bw/2, floorY-3, bw, 3, color.RGBA{R: 50, G: 80, B: 140, A: 120})
		return
	}

	var basinW, basinH, pedestalH int
	if depth == 0 {
		basinW = max(16, w/3)
		basinH = max(6, h/8)
		pedestalH = max(8, h/5)
	} else {
		basinW = max(10, w/4)
		basinH = max(4, h/10)
		pedestalH = max(6, h/6)
	}

	bx := x + w/2 - basinW/2
	by := floorY - pedestalH

	// Pedestal
	pedestalW := basinW * 2 / 3
	px := x + w/2 - pedestalW/2
	pedestalColor := brightenColor(palette.wallColorNear, 10)
	drawRect(screen, px, by, pedestalW, pedestalH, pedestalColor)

	// Basin (wider than pedestal)
	basinColor := color.RGBA{R: 140, G: 135, B: 120, A: 255}
	drawRect(screen, bx, by-basinH, basinW, basinH, basinColor)
	drawRectOutline(screen, bx, by-basinH, basinW, basinH, brightenColor(basinColor, 15))

	// Water fill inside basin
	waterColor := color.RGBA{R: 40, G: 80, B: 160, A: 150}
	drawRect(screen, bx+2, by-basinH+2, basinW-4, basinH-3, waterColor)

	// Animated water ripple (near only)
	if depth == 0 {
		frame := flickerFrame()
		rippleColor := color.RGBA{R: 80, G: 130, B: 200, A: uint8(80 + frame*15)}
		rippleW := basinW/3 + frame*2
		drawLine(screen, x+w/2-rippleW/2, by-basinH/2, x+w/2+rippleW/2, by-basinH/2, rippleColor)
	}
}

// drawArchway draws a curved archway top on open passages.
func drawArchway(screen *ebiten.Image, x, y, w, h int, palette fpvThemePalette, depth int) {
	if depth >= 2 {
		// Far — subtle curved line hint
		drawLine(screen, x+w/3, y, x+w/2, y-2, brightenColor(palette.wallColorFar, 15))
		drawLine(screen, x+w/2, y-2, x+w*2/3, y, brightenColor(palette.wallColorFar, 15))
		return
	}

	archColor := brightenColor(palette.wallColorNear, 20)
	archH := max(4, h/8)
	if depth == 1 {
		archH = max(2, h/10)
	}

	// Semicircle arch using stacked horizontal lines
	for i := 0; i < archH; i++ {
		t := float32(i) / float32(max(1, archH-1))
		halfW := int(float32(w/2) * (1.0 - t*t))
		cx := x + w/2
		ay := y - i
		if halfW > 0 {
			drawLine(screen, cx-halfW, ay, cx+halfW, ay, archColor)
		}
	}

	// Keystone at top center (near only)
	if depth == 0 && archH > 4 {
		ksColor := brightenColor(archColor, 10)
		drawRect(screen, x+w/2-2, y-archH-2, 4, 4, ksColor)
	}
}

// drawAlcove draws a recessed dark rectangle in a side wall.
func drawAlcove(screen *ebiten.Image, x, y, w, h int, palette fpvThemePalette, depth int) {
	if depth >= 2 {
		// Far — dark rectangle hint
		aw := max(4, w/5)
		ah := max(4, h/4)
		drawRect(screen, x+w/2-aw/2, y+h/3, aw, ah, color.RGBA{R: 10, G: 8, B: 15, A: 180})
		return
	}

	var alcoveW, alcoveH int
	if depth == 0 {
		alcoveW = max(12, w/4)
		alcoveH = max(16, h/3)
	} else {
		alcoveW = max(8, w/5)
		alcoveH = max(10, h/4)
	}

	ax := x + w/2 - alcoveW/2
	ay := y + (h-alcoveH)/2

	// Dark recessed interior
	recessColor := color.RGBA{R: 10, G: 8, B: 15, A: 200}
	drawRect(screen, ax, ay, alcoveW, alcoveH, recessColor)
	// Frame border
	borderColor := color.RGBA{
		R: uint8(max(0, int(palette.wallColorNear.R)-20)),
		G: uint8(max(0, int(palette.wallColorNear.G)-20)),
		B: uint8(max(0, int(palette.wallColorNear.B)-15)),
		A: 255,
	}
	drawRectOutline(screen, ax, ay, alcoveW, alcoveH, borderColor)

	// Shadow inside (near only)
	if depth == 0 && alcoveW > 4 && alcoveH > 4 {
		shadowColor := color.RGBA{R: 0, G: 0, B: 0, A: 40}
		drawRect(screen, ax+1, ay+1, 2, alcoveH-2, shadowColor)
		drawRect(screen, ax+1, ay+1, alcoveW-2, 2, shadowColor)
	}
}

// drawRubblePile draws irregular stacked rectangles on the floor.
func drawRubblePile(screen *ebiten.Image, x, y, w, h int, palette fpvThemePalette, depth int) {
	floorY := y + h

	if depth >= 2 {
		// Far — tiny dark dots
		rubbleColor := color.RGBA{
			R: uint8(max(0, int(palette.floorColor.R)-15)),
			G: uint8(max(0, int(palette.floorColor.G)-15)),
			B: uint8(max(0, int(palette.floorColor.B)-10)),
			A: 180,
		}
		drawRect(screen, x+w/2-2, floorY-3, 4, 3, rubbleColor)
		return
	}

	rubbleColor := color.RGBA{
		R: uint8(max(0, int(palette.wallColorNear.R)-20)),
		G: uint8(max(0, int(palette.wallColorNear.G)-20)),
		B: uint8(max(0, int(palette.wallColorNear.B)-15)),
		A: 220,
	}
	rubbleLight := brightenColor(rubbleColor, 15)

	cx := x + w/2
	var count int
	if depth == 0 {
		count = 5
	} else {
		count = 3
	}

	// Stack irregular blocks from bottom
	for i := 0; i < count; i++ {
		seed := (i + 1) * decoSeedPrimeA
		bw := 3 + absInt(seed*decoSeedPrimeB)%6
		bh := 2 + absInt(seed*decoSeedPrimeC)%4
		bx := cx - bw/2 + absInt(seed*decoSeedPrimeD)%5 - 2
		by := floorY - (i+1)*bh
		c := rubbleColor
		if i%2 == 0 {
			c = rubbleLight
		}
		drawRect(screen, bx, by, bw, bh, c)
	}
}

// drawRoomTypeOverlay adds room-type-specific visual accents.
func drawRoomTypeOverlay(screen *ebiten.Image, roomType string, x, y, w, h int, palette fpvThemePalette, depth int) {
	if roomType == "" {
		return
	}

	switch roomType {
	case "treasure":
		drawTreasureGlint(screen, x, y, w, h, palette, depth)
	case "shop":
		drawShopShelfLines(screen, x, y, w, h, palette, depth)
	}
}

// drawTreasureGlint draws small gold-colored glint rectangles for treasury rooms.
func drawTreasureGlint(screen *ebiten.Image, x, y, w, h int, palette fpvThemePalette, depth int) {
	if depth >= 2 {
		return
	}
	floorY := y + h
	glintColor := color.RGBA{R: 220, G: 200, B: 80, A: uint8(80 + flickerFrame()*15)}

	count := 3
	if depth == 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		seed := (i + 1) * decoSeedPrimeC
		gx := x + 4 + absInt(seed*decoSeedPrimeA)%max(1, w-8)
		gy := floorY - 4 - absInt(seed*decoSeedPrimeB)%max(1, h/4)
		drawRect(screen, gx, gy, 2, 2, glintColor)
	}
}

// drawShopShelfLines draws horizontal shelf lines on walls for shop rooms.
func drawShopShelfLines(screen *ebiten.Image, x, y, w, h int, palette fpvThemePalette, depth int) {
	if depth >= 2 {
		return
	}
	shelfColor := brightenColor(palette.wallColorNear, 10)
	shelves := 2
	if depth == 1 {
		shelves = 1
	}
	for i := 0; i < shelves; i++ {
		sy := y + h*(i+1)/(shelves+1)
		drawLine(screen, x+2, sy, x+w-2, sy, shelfColor)
		// Small item blocks on shelf (near only)
		if depth == 0 {
			for j := 0; j < 3; j++ {
				ix := x + 4 + j*(w-8)/3
				drawRect(screen, ix, sy-3, 3, 3, palette.accentColor)
			}
		}
	}
}
