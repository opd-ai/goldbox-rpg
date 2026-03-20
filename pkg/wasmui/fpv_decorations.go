//go:build js && wasm

package wasmui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Deterministic seed primes for decoration placement.
const (
	decoSeedPrimeA = 37
	decoSeedPrimeB = 53
	decoSeedPrimeC = 71
	decoSeedPrimeD = 23
)

// absInt returns the absolute value of an integer.
// Local to FPV decorations to avoid coupling with unrelated packages.
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawWallMossDetail overlays irregular green patches on stone walls for the natural theme.
// Renders at near (depth 0) and mid (depth 1) only; far depth is omitted.
func drawWallMossDetail(screen *ebiten.Image, x, y, w, h, depth, posX, posY int) {
	if w <= 4 || h <= 4 || depth >= 2 {
		return
	}
	seed := posX*decoSeedPrimeA + posY*decoSeedPrimeB
	mossColor := color.RGBA{R: 45, G: 90, B: 40, A: 80}
	count := 4
	if depth == 1 {
		count = 2
		mossColor.A = 50
	}
	for i := 0; i < count; i++ {
		s := seed + i*decoSeedPrimeC
		mx := x + (absInt(s*decoSeedPrimeA) % max(1, w-6))
		my := y + (absInt(s*decoSeedPrimeB) % max(1, h-4))
		mw := 4 + absInt(s*decoSeedPrimeD)%5
		mh := 3 + absInt(s*decoSeedPrimeC)%3
		drawRect(screen, mx, my, min(mw, w-(mx-x)), min(mh, h-(my-y)), mossColor)
	}
}

// drawWallBloodSplatter draws dark red irregular marks on walls for the horror theme.
// Renders at near depth only.
func drawWallBloodSplatter(screen *ebiten.Image, x, y, w, h, posX, posY int) {
	if w <= 4 || h <= 4 {
		return
	}
	seed := posX*decoSeedPrimeB + posY*decoSeedPrimeA
	bloodColor := color.RGBA{R: 120, G: 20, B: 15, A: 100}
	for i := 0; i < 3; i++ {
		s := seed + i*decoSeedPrimeD
		bx := x + (absInt(s*decoSeedPrimeA) % max(1, w-4))
		by := y + (absInt(s*decoSeedPrimeB) % max(1, h-6))
		drawRect(screen, bx, by, 2+absInt(s)%3, 3+absInt(s*decoSeedPrimeC)%4, bloodColor)
		// Drip line below some splats
		if i%2 == 0 {
			drawLine(screen, bx+1, by+3, bx+1, by+3+absInt(s)%6, bloodColor)
		}
	}
}

// drawWallRuneGlow draws small glowing sigil shapes on walls for the magical theme.
// Uses flickerFrame for subtle pulse animation.
func drawWallRuneGlow(screen *ebiten.Image, x, y, w, h int, accent color.RGBA, depth, posX, posY int) {
	if w <= 8 || h <= 8 || depth >= 2 {
		return
	}
	seed := posX*decoSeedPrimeC + posY*decoSeedPrimeD
	frame := flickerFrame()
	alpha := uint8(60 + frame*15)
	runeColor := color.RGBA{R: accent.R, G: accent.G, B: accent.B, A: alpha}
	count := 2
	if depth == 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		s := seed + i*decoSeedPrimeA
		rx := x + 4 + (absInt(s*decoSeedPrimeB) % max(1, w-12))
		ry := y + 4 + (absInt(s*decoSeedPrimeC) % max(1, h-12))
		if i%2 == 0 {
			// Circle sigil approximation
			drawRectOutline(screen, rx, ry, 6, 6, runeColor)
		} else {
			// Triangle sigil
			drawLine(screen, rx+3, ry, rx, ry+6, runeColor)
			drawLine(screen, rx+3, ry, rx+6, ry+6, runeColor)
			drawLine(screen, rx, ry+6, rx+6, ry+6, runeColor)
		}
	}
}

// drawWallBoneInlay draws small bone/skull outline shapes embedded in walls for the undead theme.
// Renders at near depth only.
func drawWallBoneInlay(screen *ebiten.Image, x, y, w, h, posX, posY int) {
	if w <= 8 || h <= 8 {
		return
	}
	seed := posX*decoSeedPrimeD + posY*decoSeedPrimeC
	boneColor := color.RGBA{R: 200, G: 195, B: 175, A: 100}
	for i := 0; i < 2; i++ {
		s := seed + i*decoSeedPrimeA
		bx := x + 3 + (absInt(s*decoSeedPrimeB) % max(1, w-10))
		by := y + 3 + (absInt(s*decoSeedPrimeC) % max(1, h-10))
		if i%2 == 0 {
			// Small skull: circle + jaw
			drawRectOutline(screen, bx, by, 5, 4, boneColor)
			drawLine(screen, bx+1, by+4, bx+4, by+4, boneColor)
		} else {
			// Crossed bones
			drawLine(screen, bx, by, bx+6, by+4, boneColor)
			drawLine(screen, bx+6, by, bx, by+4, boneColor)
		}
	}
}

// drawThemeWallOverlay dispatches theme-specific wall decorations based on palette theme.
// Called after base stone detail at near and mid depth.
func drawThemeWallOverlay(screen *ebiten.Image, x, y, w, h int, p *fpvParams, depth, posX, posY int) {
	switch p.palette.theme {
	case "natural":
		drawWallMossDetail(screen, x, y, w, h, depth, posX, posY)
		if depth == 0 {
			drawWallVineTendrils(screen, x, y, w, h, posX, posY)
		}
	case "horror":
		if depth == 0 {
			drawWallBloodSplatter(screen, x, y, w, h, posX, posY)
			drawWallHangingChains(screen, x, y, w, h, posX, posY)
		}
	case "magical":
		drawWallRuneGlow(screen, x, y, w, h, p.palette.accentColor, depth, posX, posY)
	case "undead":
		if depth == 0 {
			drawWallBoneInlay(screen, x, y, w, h, posX, posY)
			drawWallSkullShelf(screen, x, y, w, h, posX, posY)
		}
	case "classic":
		if depth == 0 {
			drawWallBannerShield(screen, x, y, w, h, p.palette.accentColor, posX, posY)
		}
	}
}

// drawWallBannerShield draws a wall-mounted shield/banner shape for classic theme at near depth.
func drawWallBannerShield(screen *ebiten.Image, x, y, w, h int, accent color.RGBA, posX, posY int) {
	if w <= 12 || h <= 16 {
		return
	}
	seed := posX*decoSeedPrimeA + posY*decoSeedPrimeC
	// Only draw on some walls (deterministic)
	if absInt(seed)%3 != 0 {
		return
	}
	cx := x + w/2
	cy := y + h/3
	// Shield body (rectangle)
	sw, sh := 8, 10
	shieldColor := color.RGBA{R: accent.R, G: accent.G, B: accent.B, A: 140}
	drawRect(screen, cx-sw/2, cy, sw, sh, shieldColor)
	drawRectOutline(screen, cx-sw/2, cy, sw, sh, brightenColor(shieldColor, 20))
	// Banner/pennant below (triangle pointing down)
	bannerColor := color.RGBA{R: 180, G: 40, B: 40, A: 120}
	bw := 6
	bh := 8
	by := cy + sh
	drawLine(screen, cx-bw/2, by, cx, by+bh, bannerColor)
	drawLine(screen, cx+bw/2, by, cx, by+bh, bannerColor)
	drawLine(screen, cx-bw/2, by, cx+bw/2, by, bannerColor)
}

// drawWallHangingChains draws vertical dotted lines representing hanging chains for horror theme.
func drawWallHangingChains(screen *ebiten.Image, x, y, w, h int, posX, posY int) {
	if w <= 8 || h <= 12 {
		return
	}
	seed := posX*decoSeedPrimeB + posY*decoSeedPrimeD
	chainColor := color.RGBA{R: 100, G: 95, B: 85, A: 140}
	count := 1 + absInt(seed)%2
	for i := 0; i < count; i++ {
		s := seed + i*decoSeedPrimeA
		cx := x + 4 + absInt(s*decoSeedPrimeC)%max(1, w-8)
		chainLen := h/2 + absInt(s*decoSeedPrimeD)%max(1, h/3)
		// Dotted vertical line (chain links)
		for cy := y; cy < y+chainLen; cy += 3 {
			drawRect(screen, cx, cy, 1, 2, chainColor)
		}
	}
}

// drawWallVineTendrils draws curved vine segments from ceiling downward for natural theme.
func drawWallVineTendrils(screen *ebiten.Image, x, y, w, h int, posX, posY int) {
	if w <= 8 || h <= 12 {
		return
	}
	seed := posX*decoSeedPrimeC + posY*decoSeedPrimeA
	vineColor := color.RGBA{R: 35, G: 80, B: 30, A: 120}
	count := 1 + absInt(seed)%2
	for i := 0; i < count; i++ {
		s := seed + i*decoSeedPrimeB
		vx := x + 3 + absInt(s*decoSeedPrimeD)%max(1, w-6)
		vineLen := h/3 + absInt(s*decoSeedPrimeA)%max(1, h/4)
		// Curved line segments (zigzag approximation)
		for vy := y; vy < y+vineLen; vy += 4 {
			dx := absInt((s+vy)*decoSeedPrimeC)%3 - 1 // -1, 0, or 1
			drawLine(screen, vx, vy, vx+dx, vy+4, vineColor)
			vx += dx
		}
	}
}

// drawWallSkullShelf draws a shelf with skull shapes for undead theme at near depth.
func drawWallSkullShelf(screen *ebiten.Image, x, y, w, h int, posX, posY int) {
	if w <= 12 || h <= 16 {
		return
	}
	seed := posX*decoSeedPrimeD + posY*decoSeedPrimeB
	// Only draw on some walls
	if absInt(seed)%3 != 0 {
		return
	}
	shelfY := y + h*2/5
	shelfColor := color.RGBA{R: 100, G: 95, B: 80, A: 160}
	// Shelf line
	drawLine(screen, x+2, shelfY, x+w-2, shelfY, shelfColor)
	// Small skull shapes on shelf
	skullColor := color.RGBA{R: 200, G: 195, B: 175, A: 140}
	count := max(1, (w-8)/8)
	for i := 0; i < count; i++ {
		sx := x + 4 + i*(w-8)/max(1, count)
		// Skull = small circle (square) + jaw line
		drawRect(screen, sx, shelfY-5, 4, 4, skullColor)
		drawLine(screen, sx+1, shelfY-1, sx+3, shelfY-1, skullColor)
	}
}

// drawFloorCheckerboard draws alternating light/dark floor tile shading at near depth.
func drawFloorCheckerboard(screen *ebiten.Image, p *fpvParams) {
	floorY := p.nearBottom
	floorH := p.vpY + p.vpHeight - p.nearBottom
	floorX := p.vpX + p.nearInset
	floorW := p.vpWidth - 2*p.nearInset
	if floorW <= 0 || floorH <= 0 {
		return
	}
	tileW := max(1, floorW/4)
	tileH := max(1, floorH/2)
	lightColor := color.RGBA{R: 255, G: 255, B: 255, A: 12}
	for row := 0; row < 2; row++ {
		for col := 0; col < 4; col++ {
			if (row+col)%2 == 0 {
				tx := floorX + col*tileW
				ty := floorY + row*tileH
				tw := min(tileW, floorX+floorW-tx)
				th := min(tileH, p.vpY+p.vpHeight-ty)
				if tw > 0 && th > 0 {
					drawRect(screen, tx, ty, tw, th, lightColor)
				}
			}
		}
	}
}

// drawFloorCracks draws deterministic hairline crack lines on the floor at near depth.
func drawFloorCracks(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	floorY := p.nearBottom
	floorH := p.vpY + p.vpHeight - p.nearBottom
	floorX := p.vpX + p.nearInset
	floorW := p.vpWidth - 2*p.nearInset
	if floorW <= 8 || floorH <= 4 {
		return
	}
	seed := posX*decoSeedPrimeB + posY*decoSeedPrimeA
	crackColor := color.RGBA{R: 0, G: 0, B: 0, A: 40}
	for i := 0; i < 3; i++ {
		s := seed + i*decoSeedPrimeC
		cx := floorX + (absInt(s*decoSeedPrimeA) % max(1, floorW))
		cy := floorY + (absInt(s*decoSeedPrimeB) % max(1, floorH))
		dx := 3 + absInt(s*decoSeedPrimeD)%6
		dy := 2 + absInt(s)%4
		drawLine(screen, cx, cy, cx+dx, cy+dy, crackColor)
	}
}

// drawFloorDebris draws small rubble dots on the floor at deterministic positions.
func drawFloorDebris(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	floorY := p.nearBottom
	floorH := p.vpY + p.vpHeight - p.nearBottom
	floorX := p.vpX + p.nearInset
	floorW := p.vpWidth - 2*p.nearInset
	if floorW <= 4 || floorH <= 4 {
		return
	}
	seed := posX*decoSeedPrimeC + posY*decoSeedPrimeD
	debrisColor := color.RGBA{R: 80, G: 75, B: 65, A: 80}
	for i := 0; i < 5; i++ {
		s := seed + i*decoSeedPrimeA
		dx := floorX + (absInt(s*decoSeedPrimeB) % max(1, floorW))
		dy := floorY + (absInt(s*decoSeedPrimeC) % max(1, floorH))
		drawRect(screen, dx, dy, 1+absInt(s)%2, 1+absInt(s*decoSeedPrimeD)%2, debrisColor)
	}
}

// drawFloorNaturalPatches draws irregular earth-tone patches for the natural theme floor.
func drawFloorNaturalPatches(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	floorY := p.nearBottom
	floorH := p.vpY + p.vpHeight - p.nearBottom
	floorX := p.vpX + p.nearInset
	floorW := p.vpWidth - 2*p.nearInset
	if floorW <= 4 || floorH <= 4 {
		return
	}
	seed := posX*decoSeedPrimeA + posY*decoSeedPrimeC
	earthColor := color.RGBA{R: 70, G: 60, B: 35, A: 35}
	for i := 0; i < 4; i++ {
		s := seed + i*decoSeedPrimeB
		px := floorX + (absInt(s*decoSeedPrimeC) % max(1, floorW-8))
		py := floorY + (absInt(s*decoSeedPrimeD) % max(1, floorH-4))
		pw := 6 + absInt(s*decoSeedPrimeA)%8
		ph := 3 + absInt(s)%4
		drawRect(screen, px, py, min(pw, floorX+floorW-px), min(ph, p.vpY+p.vpHeight-py), earthColor)
	}
}

// drawFloorBoneFragments draws scattered bone-like lines on the floor for the undead theme.
func drawFloorBoneFragments(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	floorY := p.nearBottom
	floorH := p.vpY + p.vpHeight - p.nearBottom
	floorX := p.vpX + p.nearInset
	floorW := p.vpWidth - 2*p.nearInset
	if floorW <= 4 || floorH <= 4 {
		return
	}
	seed := posX*decoSeedPrimeD + posY*decoSeedPrimeA
	boneColor := color.RGBA{R: 190, G: 185, B: 170, A: 70}
	for i := 0; i < 4; i++ {
		s := seed + i*decoSeedPrimeB
		bx := floorX + (absInt(s*decoSeedPrimeC) % max(1, floorW))
		by := floorY + (absInt(s*decoSeedPrimeD) % max(1, floorH))
		drawLine(screen, bx, by, bx+3+absInt(s)%4, by+absInt(s*decoSeedPrimeA)%3, boneColor)
	}
}

// drawFloorDetails dispatches floor detail rendering based on theme.
func drawFloorDetails(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	// Perspective floor tile pattern that recedes toward vanishing point
	drawFloorPerspectiveTiles(screen, p)
	switch p.palette.theme {
	case "natural":
		drawFloorNaturalPatches(screen, p, posX, posY)
		drawFloorMushroomDots(screen, p, posX, posY)
	case "undead":
		drawFloorBoneFragments(screen, p, posX, posY)
	case "horror":
		drawFloorPentagram(screen, p, posX, posY)
		drawFloorCheckerboard(screen, p)
	case "classic":
		drawFloorCheckerboard(screen, p)
		drawFloorMosaicSquares(screen, p, posX, posY)
	default:
		drawFloorCheckerboard(screen, p)
	}
	drawFloorCracks(screen, p, posX, posY)
	drawFloorDebris(screen, p, posX, posY)
}

// drawFloorPerspectiveTiles draws a trapezoidal grid that recedes toward the vanishing point.
func drawFloorPerspectiveTiles(screen *ebiten.Image, p *fpvParams) {
	gridColor := color.RGBA{R: 255, G: 255, B: 255, A: 8}
	// Draw horizontal lines at depth boundaries with perspective narrowing
	depths := []struct{ y, inset int }{
		{p.farBottom, p.farInset},
		{p.midBottom, p.midInset},
		{p.nearBottom, p.nearInset},
	}
	for _, d := range depths {
		lx := p.vpX + d.inset
		rx := p.vpX + p.vpWidth - d.inset
		drawLine(screen, lx, d.y, rx, d.y, gridColor)
	}
	// Draw converging vertical lines from bottom to vanishing point
	floorBottom := p.vpY + p.vpHeight
	lines := 5
	for i := 0; i <= lines; i++ {
		bx := p.vpX + i*(p.vpWidth)/lines
		drawLine(screen, bx, floorBottom, p.vanishX, p.vanishY, gridColor)
	}
}

// drawFloorPentagram draws a faint pentagram outline on the floor for horror theme.
func drawFloorPentagram(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	seed := posX*decoSeedPrimeA + posY*decoSeedPrimeD
	if absInt(seed)%4 != 0 {
		return // Only draw occasionally
	}
	floorY := p.nearBottom
	floorH := p.vpY + p.vpHeight - p.nearBottom
	floorX := p.vpX + p.nearInset
	floorW := p.vpWidth - 2*p.nearInset
	if floorW <= 16 || floorH <= 8 {
		return
	}
	pentColor := color.RGBA{R: 140, G: 30, B: 30, A: 60}
	cx := floorX + floorW/2
	cy := floorY + floorH/2
	r := min(floorW/4, floorH/2)
	// Circle outline (octagon approximation)
	drawRectOutline(screen, cx-r, cy-r, r*2, r*2, pentColor)
	// Star lines (simplified X + vertical)
	drawLine(screen, cx-r, cy, cx+r, cy, pentColor)
	drawLine(screen, cx, cy-r, cx, cy+r, pentColor)
	drawLine(screen, cx-r, cy-r, cx+r, cy+r, pentColor)
	drawLine(screen, cx+r, cy-r, cx-r, cy+r, pentColor)
}

// drawFloorMosaicSquares draws small mosaic tile squares for classic theme floor.
func drawFloorMosaicSquares(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	floorY := p.nearBottom
	floorH := p.vpY + p.vpHeight - p.nearBottom
	floorX := p.vpX + p.nearInset
	floorW := p.vpWidth - 2*p.nearInset
	if floorW <= 8 || floorH <= 4 {
		return
	}
	seed := posX*decoSeedPrimeC + posY*decoSeedPrimeA
	mosaicColor := color.RGBA{R: 80, G: 70, B: 110, A: 30}
	for i := 0; i < 4; i++ {
		s := seed + i*decoSeedPrimeB
		mx := floorX + absInt(s*decoSeedPrimeA)%max(1, floorW-4)
		my := floorY + absInt(s*decoSeedPrimeD)%max(1, floorH-4)
		drawRect(screen, mx, my, 3, 3, mosaicColor)
	}
}

// drawFloorMushroomDots draws small mushroom-like dots for natural theme floor.
func drawFloorMushroomDots(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	floorY := p.nearBottom
	floorH := p.vpY + p.vpHeight - p.nearBottom
	floorX := p.vpX + p.nearInset
	floorW := p.vpWidth - 2*p.nearInset
	if floorW <= 4 || floorH <= 4 {
		return
	}
	seed := posX*decoSeedPrimeB + posY*decoSeedPrimeD
	mushColor := color.RGBA{R: 160, G: 140, B: 90, A: 80}
	for i := 0; i < 3; i++ {
		s := seed + i*decoSeedPrimeC
		mx := floorX + absInt(s*decoSeedPrimeA)%max(1, floorW)
		my := floorY + absInt(s*decoSeedPrimeB)%max(1, floorH)
		// Cap (small dot)
		drawRect(screen, mx, my, 2, 1, mushColor)
		// Stem (1px below)
		drawRect(screen, mx, my+1, 1, 1, color.RGBA{R: 140, G: 130, B: 100, A: 60})
	}
}

// drawCobwebCorners draws small cobweb triangle shapes in the upper corners of the near viewport.
func drawCobwebCorners(screen *ebiten.Image, p *fpvParams) {
	webColor := color.RGBA{R: 180, G: 175, B: 160, A: 50}
	lx := p.vpX + p.nearInset
	rx := p.vpX + p.vpWidth - p.nearInset
	ty := p.nearTop
	sz := min(12, (p.vpWidth-2*p.nearInset)/8)
	if sz <= 2 {
		return
	}
	// Top-left cobweb
	drawLine(screen, lx, ty, lx+sz, ty, webColor)
	drawLine(screen, lx, ty, lx, ty+sz, webColor)
	drawLine(screen, lx, ty, lx+sz, ty+sz, webColor)
	// Top-right cobweb
	drawLine(screen, rx, ty, rx-sz, ty, webColor)
	drawLine(screen, rx, ty, rx, ty+sz, webColor)
	drawLine(screen, rx, ty, rx-sz, ty+sz, webColor)
}

// drawCeilingCrossbeams draws small filled squares at beam intersections.
func drawCeilingCrossbeams(screen *ebiten.Image, p *fpvParams) {
	crossColor := color.RGBA{R: 60, G: 55, B: 80, A: 80}
	sz := 3
	// Intersections at depth boundaries on quarter lines
	pts := []struct{ x, y int }{
		{p.vpX + p.vpWidth/4, p.nearTop},
		{p.vpX + p.vpWidth*3/4, p.nearTop},
		{p.vanishX, p.midTop},
	}
	for _, pt := range pts {
		drawRect(screen, pt.x-sz/2, pt.y-sz/2, sz, sz, crossColor)
	}
}

// drawCeilingHorrorDrips draws red-tinted dripping lines from the ceiling for the horror theme.
func drawCeilingHorrorDrips(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	seed := posX*decoSeedPrimeA + posY*decoSeedPrimeD
	dripColor := color.RGBA{R: 120, G: 30, B: 25, A: 140}
	nearW := p.vpWidth - 2*p.nearInset
	baseX := p.vpX + p.nearInset
	if nearW <= 0 {
		return
	}
	for i := 0; i < 3; i++ {
		offset := absInt(seed+i*dripOffsetPrime) % max(1, nearW)
		dx := baseX + offset
		dh := dripMinHeight + absInt(seed+i*dripHeightPrime)%dripHeightRange
		dy := p.nearTop
		drawLine(screen, dx, dy, dx, dy+dh, dripColor)
	}
}

// drawCeilingMagicSparks draws faint floating spark dots on the ceiling for the magical theme.
// Uses flickerFrame to toggle visibility for subtle animation.
func drawCeilingMagicSparks(screen *ebiten.Image, p *fpvParams, accent color.RGBA, posX, posY int) {
	frame := flickerFrame()
	seed := posX*decoSeedPrimeB + posY*decoSeedPrimeC
	sparkColor := color.RGBA{R: accent.R, G: accent.G, B: accent.B, A: 120}
	nearW := p.vpWidth - 2*p.nearInset
	baseX := p.vpX + p.nearInset
	if nearW <= 0 {
		return
	}
	for i := 0; i < 3; i++ {
		if (i+frame)%2 != 0 {
			continue // Toggle visibility per frame
		}
		s := seed + i*decoSeedPrimeA
		sx := baseX + (absInt(s*decoSeedPrimeD) % max(1, nearW))
		sy := p.nearTop + (absInt(s*decoSeedPrimeC) % max(1, (p.vanishY-p.nearTop)/2+1))
		drawRect(screen, sx, sy, 2, 2, sparkColor)
	}
}

// drawCeilingDetails dispatches ceiling detail rendering based on theme.
func drawCeilingDetails(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	drawCobwebCorners(screen, p)
	drawCeilingCrossbeams(screen, p)
	switch p.palette.theme {
	case "horror":
		drawCeilingHorrorDrips(screen, p, posX, posY)
	case "magical":
		drawCeilingMagicSparks(screen, p, p.palette.accentColor, posX, posY)
	case "natural":
		drawCeilingStalactites(screen, p, posX, posY)
	case "classic":
		drawCeilingChandelier(screen, p)
	}
}

// drawCeilingStalactites draws larger stalactite shapes hanging from the ceiling for natural theme.
func drawCeilingStalactites(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	seed := posX*dripSeedPrimeX + posY*dripSeedPrimeY
	stalColor := color.RGBA{R: 80, G: 75, B: 60, A: 180}
	stalLight := color.RGBA{R: 100, G: 95, B: 75, A: 140}
	nearW := p.vpWidth - 2*p.nearInset
	baseX := p.vpX + p.nearInset
	if nearW <= 0 {
		return
	}
	for i := 0; i < 5; i++ {
		offset := absInt(seed+i*dripOffsetPrime) % max(1, nearW)
		dx := baseX + offset
		dh := dripMinHeight + 2 + absInt(seed+i*dripHeightPrime)%(dripHeightRange+4)
		dy := p.nearTop
		// Wider base, narrowing to a point
		baseW := 2 + absInt(seed+i*decoSeedPrimeA)%3
		for sy := 0; sy < dh; sy++ {
			t := float32(sy) / float32(max(1, dh))
			sw := int(float32(baseW) * (1.0 - t))
			if sw > 0 {
				drawRect(screen, dx-sw/2, dy+sy, sw, 1, stalColor)
			}
		}
		// Highlight edge
		drawLine(screen, dx, dy, dx, dy+dh, stalLight)
	}
}

// drawCeilingChandelier draws a hanging lantern/chandelier shape at mid depth for classic theme.
func drawCeilingChandelier(screen *ebiten.Image, p *fpvParams) {
	cx := p.vanishX
	cy := p.midTop + 4
	// Chain from ceiling
	chainColor := color.RGBA{R: 120, G: 110, B: 90, A: 160}
	drawLine(screen, cx, p.midTop, cx, cy+4, chainColor)
	// Lantern body
	lanternColor := color.RGBA{R: 140, G: 120, B: 60, A: 180}
	lw, lh := 8, 6
	drawRect(screen, cx-lw/2, cy+4, lw, lh, lanternColor)
	drawRectOutline(screen, cx-lw/2, cy+4, lw, lh, brightenColor(lanternColor, 20))
	// Flame glow inside
	frame := flickerFrame()
	glowAlpha := uint8(60 + frame*15)
	glowColor := color.RGBA{R: 220, G: 180, B: 60, A: glowAlpha}
	drawRect(screen, cx-2, cy+5, 4, 4, glowColor)
}

// drawDoorRivets draws iron nail/rivet dots along iron bands on a closed door.
func drawDoorRivets(screen *ebiten.Image, dx, dy, dw, dh int, bandColor color.RGBA) {
	if dw <= 8 || dh <= 8 {
		return
	}
	// 4 rivets along each of the 2 iron bands
	for _, bandY := range []int{dy + dh/4, dy + dh*3/4} {
		for j := 0; j < 4; j++ {
			rx := dx + 4 + j*(dw-8)/3
			drawRect(screen, rx, bandY-1, 2, 2, bandColor)
		}
	}
}

// drawDoorKeyhole draws a keyhole shape below the door handle.
func drawDoorKeyhole(screen *ebiten.Image, dx, dy, dw, dh int) {
	if dw <= 8 || dh <= 12 {
		return
	}
	khColor := color.RGBA{R: 20, G: 18, B: 15, A: 220}
	kx := dx + dw*3/4
	ky := dy + dh/2 + 5
	// Circle part (approximated as small square)
	drawRect(screen, kx, ky, 3, 3, khColor)
	// Slot below
	drawRect(screen, kx+1, ky+3, 1, 3, khColor)
}

// drawDoorFrameShadow draws a 2px dark strip on left and top edges for depth illusion.
func drawDoorFrameShadow(screen *ebiten.Image, dx, dy, dw, dh int) {
	if dw <= 2 || dh <= 2 {
		return
	}
	shadow := color.RGBA{R: 0, G: 0, B: 0, A: 60}
	drawRect(screen, dx-2, dy-1, 2, dh+1, shadow)
	drawRect(screen, dx, dy-2, dw, 2, shadow)
}

// drawDoorMagicalGlow draws a faint accent-colored outline around the door frame.
func drawDoorMagicalGlow(screen *ebiten.Image, dx, dy, dw, dh int, accent color.RGBA) {
	if dw <= 4 || dh <= 4 {
		return
	}
	frame := flickerFrame()
	alpha := uint8(40 + frame*10)
	glow := color.RGBA{R: accent.R, G: accent.G, B: accent.B, A: alpha}
	drawRectOutline(screen, dx-1, dy-1, dw+2, dh+2, glow)
}

// drawAmbientOcclusion darkens strips where walls meet floor and ceiling.
func drawAmbientOcclusion(screen *ebiten.Image, p *fpvParams) {
	nearX := p.vpX + p.nearInset
	nearW := p.vpWidth - 2*p.nearInset
	if nearW <= 0 {
		return
	}
	// Floor-wall junction: darken 6px strip above floor line
	for i := 0; i < 6; i++ {
		a := uint8(25 - i*4)
		if a > 25 {
			a = 0
		}
		c := color.RGBA{R: 0, G: 0, B: 0, A: a}
		drawRect(screen, nearX, p.nearBottom-6+i, nearW, 1, c)
	}
	// Ceiling-wall junction: darken 6px strip below ceiling line
	for i := 0; i < 6; i++ {
		a := uint8(25 - i*4)
		if a > 25 {
			a = 0
		}
		drawRect(screen, nearX, p.nearTop+i, nearW, 1, color.RGBA{R: 0, G: 0, B: 0, A: a})
	}
}

// drawTorchLightCone draws a warm-tinted trapezoid below a torch on the floor.
func drawTorchLightCone(screen *ebiten.Image, cx, floorY int, palette fpvThemePalette) {
	coneColor := color.RGBA{R: palette.torchGlow.R, G: palette.torchGlow.G, B: palette.torchGlow.B / 2, A: 20}
	// Small warm trapezoid widening downward from torch toward floor
	for i := 0; i < 8; i++ {
		halfW := 2 + i
		y := floorY - 8 + i
		drawRect(screen, cx-halfW, y, halfW*2, 1, coneColor)
	}
}

// drawVignette darkens the viewport edges for CRT/dungeon immersion.
func drawVignette(screen *ebiten.Image, p *fpvParams) {
	borderSize := 8
	if p.vpWidth <= borderSize*2 || p.vpHeight <= borderSize*2 {
		return
	}
	for i := 0; i < borderSize; i++ {
		a := uint8(20 - i*2)
		if a > 20 {
			a = 0
		}
		c := color.RGBA{R: 0, G: 0, B: 0, A: a}
		// Top edge
		drawRect(screen, p.vpX, p.vpY+i, p.vpWidth, 1, c)
		// Bottom edge
		drawRect(screen, p.vpX, p.vpY+p.vpHeight-1-i, p.vpWidth, 1, c)
		// Left edge
		drawRect(screen, p.vpX+i, p.vpY, 1, p.vpHeight, c)
		// Right edge
		drawRect(screen, p.vpX+p.vpWidth-1-i, p.vpY, 1, p.vpHeight, c)
	}
}

// drawCorridorDepthHint draws a smaller receding trapezoid inside an opening
// to suggest continued hallway at far depth.
func drawCorridorDepthHint(screen *ebiten.Image, ox, oy, ow, oh int, openingColor color.RGBA) {
	if ow <= 6 || oh <= 6 {
		return
	}
	// Inner corridor walls: slightly brighter rectangles inset from opening edges
	inset := max(2, ow/6)
	innerColor := brightenColor(openingColor, 8)
	// Left inner wall
	drawRect(screen, ox, oy, inset, oh, innerColor)
	// Right inner wall
	drawRect(screen, ox+ow-inset, oy, inset, oh, innerColor)
	// Far back wall hint
	backInset := max(3, ow/4)
	drawLine(screen, ox+backInset, oy+oh/4, ox+ow-backInset, oy+oh/4, innerColor)
}
