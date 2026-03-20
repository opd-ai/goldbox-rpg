//go:build js && wasm

package wasmui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// fpvThemePalette defines the color palette for a dungeon theme.
// Each palette provides colors for walls, floor, ceiling, doors, and accents.
type fpvThemePalette struct {
	wallColorFar  color.RGBA
	wallColorMid  color.RGBA
	wallColorNear color.RGBA
	floorColor    color.RGBA
	ceilingColor  color.RGBA
	doorColor     color.RGBA
	trimColor     color.RGBA
	torchFlame    color.RGBA
	torchGlow     color.RGBA
	accentColor   color.RGBA
	openingColor  color.RGBA
	theme         string
}

// getThemePalette returns the FPV color palette for the given dungeon theme.
// Falls back to "classic" if the theme is unrecognized.
func getThemePalette(theme string) fpvThemePalette {
	switch theme {
	case "horror":
		return fpvThemePalette{
			wallColorFar:  color.RGBA{R: 80, G: 30, B: 30, A: 255},
			wallColorMid:  color.RGBA{R: 120, G: 50, B: 50, A: 255},
			wallColorNear: color.RGBA{R: 160, G: 70, B: 70, A: 255},
			floorColor:    color.RGBA{R: 50, G: 35, B: 35, A: 255},
			ceilingColor:  color.RGBA{R: 25, G: 15, B: 15, A: 255},
			doorColor:     color.RGBA{R: 140, G: 80, B: 60, A: 255},
			trimColor:     color.RGBA{R: 60, G: 25, B: 25, A: 255},
			torchFlame:    color.RGBA{R: 200, G: 80, B: 30, A: 220},
			torchGlow:     color.RGBA{R: 180, G: 60, B: 20, A: 60},
			accentColor:   color.RGBA{R: 200, G: 50, B: 50, A: 255},
			openingColor:  color.RGBA{R: 15, G: 8, B: 8, A: 255},
			theme:         "horror",
		}
	case "natural":
		return fpvThemePalette{
			wallColorFar:  color.RGBA{R: 60, G: 80, B: 50, A: 255},
			wallColorMid:  color.RGBA{R: 90, G: 120, B: 70, A: 255},
			wallColorNear: color.RGBA{R: 120, G: 160, B: 90, A: 255},
			floorColor:    color.RGBA{R: 55, G: 65, B: 40, A: 255},
			ceilingColor:  color.RGBA{R: 30, G: 35, B: 22, A: 255},
			doorColor:     color.RGBA{R: 140, G: 120, B: 60, A: 255},
			trimColor:     color.RGBA{R: 50, G: 60, B: 35, A: 255},
			torchFlame:    color.RGBA{R: 200, G: 180, B: 50, A: 220},
			torchGlow:     color.RGBA{R: 150, G: 140, B: 40, A: 60},
			accentColor:   color.RGBA{R: 100, G: 180, B: 60, A: 255},
			openingColor:  color.RGBA{R: 12, G: 15, B: 8, A: 255},
			theme:         "natural",
		}
	case "undead":
		return fpvThemePalette{
			wallColorFar:  color.RGBA{R: 80, G: 85, B: 70, A: 255},
			wallColorMid:  color.RGBA{R: 130, G: 135, B: 110, A: 255},
			wallColorNear: color.RGBA{R: 180, G: 185, B: 160, A: 255},
			floorColor:    color.RGBA{R: 55, G: 60, B: 50, A: 255},
			ceilingColor:  color.RGBA{R: 28, G: 30, B: 25, A: 255},
			doorColor:     color.RGBA{R: 140, G: 145, B: 100, A: 255},
			trimColor:     color.RGBA{R: 60, G: 65, B: 50, A: 255},
			torchFlame:    color.RGBA{R: 120, G: 200, B: 100, A: 220},
			torchGlow:     color.RGBA{R: 80, G: 150, B: 60, A: 60},
			accentColor:   color.RGBA{R: 100, G: 200, B: 80, A: 255},
			openingColor:  color.RGBA{R: 10, G: 12, B: 8, A: 255},
			theme:         "undead",
		}
	case "magical":
		return fpvThemePalette{
			wallColorFar:  color.RGBA{R: 70, G: 50, B: 120, A: 255},
			wallColorMid:  color.RGBA{R: 110, G: 80, B: 180, A: 255},
			wallColorNear: color.RGBA{R: 150, G: 120, B: 230, A: 255},
			floorColor:    color.RGBA{R: 45, G: 40, B: 70, A: 255},
			ceilingColor:  color.RGBA{R: 22, G: 20, B: 40, A: 255},
			doorColor:     color.RGBA{R: 140, G: 120, B: 200, A: 255},
			trimColor:     color.RGBA{R: 50, G: 40, B: 80, A: 255},
			torchFlame:    color.RGBA{R: 150, G: 100, B: 255, A: 220},
			torchGlow:     color.RGBA{R: 100, G: 60, B: 200, A: 60},
			accentColor:   color.RGBA{R: 100, G: 220, B: 255, A: 255},
			openingColor:  color.RGBA{R: 10, G: 8, B: 20, A: 255},
			theme:         "magical",
		}
	default: // classic
		return fpvThemePalette{
			wallColorFar:  ColorPanelBorder,
			wallColorMid:  ColorStatValue,
			wallColorNear: ColorPanelBorderHi,
			floorColor:    color.RGBA{R: 60, G: 55, B: 70, A: 255},
			ceilingColor:  color.RGBA{R: 30, G: 28, B: 42, A: 255},
			doorColor:     ColorGold,
			trimColor:     color.RGBA{R: 50, G: 40, B: 70, A: 255},
			torchFlame:    color.RGBA{R: 200, G: 120, B: 30, A: 220},
			torchGlow:     color.RGBA{R: 180, G: 120, B: 30, A: 60},
			accentColor:   ColorGold,
			openingColor:  color.RGBA{R: 20, G: 20, B: 30, A: 255},
			theme:         "classic",
		}
	}
}

// flickerFrame returns the current torch flicker frame (0-3) based on wall clock.
// Uses time division to avoid storing animation state.
func flickerFrame() int {
	return int(time.Now().UnixMilli()/150) % 4
}

// flickerFlameColor returns a flame color varying by flicker frame.
// Cycles between orange, yellow, and bright yellow EGA shades.

// clampChannel adjusts a uint8 color channel by delta in int space and clamps to [0, 255].
func clampChannel(base uint8, delta int) uint8 {
	v := int(base) + delta
	if v > 255 {
		v = 255
	} else if v < 0 {
		v = 0
	}
	return uint8(v)
}

func flickerFlameColor(base color.RGBA, frame int) color.RGBA {
	switch frame % 4 {
	case 0:
		return base
	case 1:
		return color.RGBA{
			R: clampChannel(base.R, 40),
			G: clampChannel(base.G, 60),
			B: base.B,
			A: base.A,
		}
	case 2:
		return color.RGBA{
			R: clampChannel(base.R, 20),
			G: clampChannel(base.G, 30),
			B: clampChannel(base.B, 20),
			A: base.A,
		}
	default:
		return color.RGBA{
			R: base.R,
			G: clampChannel(base.G, 40),
			B: base.B,
			A: base.A,
		}
	}
}

// drawTorchFlicker draws an animated torch sconce with frame-to-frame flicker.
// cx, cy: center position for the torch bracket, palette provides theme colors.
func drawTorchFlicker(screen *ebiten.Image, cx, cy int, palette fpvThemePalette) {
	frame := flickerFrame()
	bracketColor := color.RGBA{R: 130, G: 110, B: 90, A: 255}
	// Bracket and shaft
	drawRectOutline(screen, cx-3, cy, 6, 10, bracketColor)
	drawRect(screen, cx-1, cy-8, 2, 10, bracketColor)
	// Glow on wall (varies with frame)
	glowAlpha := uint8(40 + frame*10)
	glow := color.RGBA{R: palette.torchGlow.R, G: palette.torchGlow.G, B: palette.torchGlow.B, A: glowAlpha}
	drawRect(screen, cx-8, cy-20, 16, 24, glow)
	// Flame layers with flicker offset
	yOff := -1 // ±1px vertical wobble
	if frame%2 == 1 {
		yOff = 1
	}
	flame := flickerFlameColor(palette.torchFlame, frame)
	drawRect(screen, cx-3, cy-14+yOff, 6, 7, flame)
	mid := flickerFlameColor(palette.torchFlame, frame+1)
	drawRect(screen, cx-2, cy-16+yOff, 4, 7, mid)
	tip := color.RGBA{R: 255, G: 240, B: 150, A: 180}
	drawRect(screen, cx-1, cy-17+yOff, 2, 5, tip)
}

// drawDepthFog overlays semi-transparent dark rectangles for distance darkening.
// Creates atmospheric perspective: far areas appear hazier than near areas.
func drawDepthFog(screen *ebiten.Image, p *fpvParams) {
	// Far depth fog (stronger)
	farX := p.vpX + p.farInset
	farW := p.vpWidth - 2*p.farInset
	farH := p.farBottom - p.farTop
	if farW > 0 && farH > 0 {
		drawRect(screen, farX, p.farTop, farW, farH, color.RGBA{R: 0, G: 0, B: 0, A: 55})
	}
	// Mid depth fog (lighter)
	midX := p.vpX + p.midInset
	midW := p.vpWidth - 2*p.midInset
	midH := p.midBottom - p.midTop
	if midW > 0 && midH > 0 {
		drawRect(screen, midX, p.midTop, midW, midH, color.RGBA{R: 0, G: 0, B: 0, A: 20})
	}
}

// drawStairsNear draws ascending steps at near depth (4-5 visible steps).
func drawStairsNear(screen *ebiten.Image, x, y, w, h int, baseColor color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	steps := 5
	stepH := h / steps
	if stepH == 0 {
		return
	}
	light := brightenColor(baseColor, 20)
	dark := color.RGBA{
		R: uint8(max(0, int(baseColor.R)-25)),
		G: uint8(max(0, int(baseColor.G)-25)),
		B: uint8(max(0, int(baseColor.B)-20)),
		A: 255,
	}
	for i := 0; i < steps; i++ {
		sy := y + h - (i+1)*stepH
		sw := w - i*(w/(steps*2))
		sx := x + (w-sw)/2
		if i%2 == 0 {
			drawRect(screen, sx, sy, sw, stepH, light)
		} else {
			drawRect(screen, sx, sy, sw, stepH, dark)
		}
		drawLine(screen, sx, sy, sx+sw, sy, baseColor) // step edge
	}
}

// drawStairsMid draws ascending steps at mid depth (2-3 visible steps).
func drawStairsMid(screen *ebiten.Image, x, y, w, h int, baseColor color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	steps := 3
	stepH := h / steps
	if stepH == 0 {
		return
	}
	light := brightenColor(baseColor, 15)
	dark := color.RGBA{
		R: uint8(max(0, int(baseColor.R)-20)),
		G: uint8(max(0, int(baseColor.G)-20)),
		B: uint8(max(0, int(baseColor.B)-15)),
		A: 255,
	}
	for i := 0; i < steps; i++ {
		sy := y + h - (i+1)*stepH
		clr := light
		if i%2 == 1 {
			clr = dark
		}
		drawRect(screen, x, sy, w, stepH, clr)
		drawLine(screen, x, sy, x+w, sy, baseColor)
	}
}

// drawStairsFar draws a stairs indicator at far depth (arrow hint).
func drawStairsFar(screen *ebiten.Image, x, y, w, h int, baseColor color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	drawRect(screen, x, y, w, h, baseColor)
	// Single step line in center
	drawLine(screen, x, y+h/2, x+w, y+h/2, brightenColor(baseColor, 25))
	// Upward arrow hint
	cx := x + w/2
	cy := y + h/3
	lineC := brightenColor(baseColor, 40)
	drawLine(screen, cx, cy+h/6, cx, cy, lineC)
	drawLine(screen, cx-w/6, cy+h/8, cx, cy, lineC)
	drawLine(screen, cx+w/6, cy+h/8, cx, cy, lineC)
}

// Deterministic seed primes for ceiling drip placement.
const (
	dripSeedPrimeX  = 31
	dripSeedPrimeY  = 17
	dripOffsetPrime = 47
	dripHeightPrime = 13
	dripMinHeight   = 4
	dripHeightRange = 6
)

// drawCeilingDrips draws stalactite hints on the ceiling for natural/cave themes.
// Uses player position as a deterministic seed for stable placement.
func drawCeilingDrips(screen *ebiten.Image, p *fpvParams, posX, posY int) {
	seed := posX*dripSeedPrimeX + posY*dripSeedPrimeY
	dripColor := color.RGBA{R: 60, G: 55, B: 50, A: 180}
	nearW := p.vpWidth - 2*p.nearInset
	baseX := p.vpX + p.nearInset
	// Draw 3 small stalactite triangles at deterministic positions
	for i := 0; i < 3; i++ {
		offset := ((seed + i*dripOffsetPrime) % max(1, nearW))
		dx := baseX + offset
		dh := dripMinHeight + (seed+i*dripHeightPrime)%dripHeightRange
		dy := p.nearTop
		drawLine(screen, dx, dy, dx, dy+dh, dripColor)
		drawLine(screen, dx-1, dy, dx, dy+dh, dripColor)
		drawLine(screen, dx+1, dy, dx, dy+dh, dripColor)
	}
}

// drawWallEdgeHighlightLeft draws a bright strip on the right edge of a left wall.
func drawWallEdgeHighlightLeft(screen *ebiten.Image, x, y, w, h int, baseColor color.RGBA) {
	if w <= 2 || h <= 0 {
		return
	}
	bright := brightenColor(baseColor, 18)
	drawRect(screen, x+w-2, y, 2, h, bright)
}

// drawWallEdgeHighlightRight draws a bright strip on the left edge of a right wall.
func drawWallEdgeHighlightRight(screen *ebiten.Image, x, y, w, h int, baseColor color.RGBA) {
	if w <= 2 || h <= 0 {
		return
	}
	bright := brightenColor(baseColor, 18)
	drawRect(screen, x, y, 2, h, bright)
}

// drawWallEdgeHighlightCenter draws a bright strip on the top edge of a center wall.
func drawWallEdgeHighlightCenter(screen *ebiten.Image, x, y, w, h int, baseColor color.RGBA) {
	if w <= 0 || h <= 2 {
		return
	}
	bright := brightenColor(baseColor, 18)
	drawRect(screen, x, y, w, 2, bright)
}

// drawViewportBorderFrame draws a themed decorative border around the viewport.
// Adds corner ornaments and an inner border line.
func drawViewportBorderFrame(screen *ebiten.Image, vpX, vpY, vpW, vpH int, palette fpvThemePalette) {
	// Thin inner border
	drawRectOutline(screen, vpX-1, vpY-1, vpW+2, vpH+2, palette.trimColor)
	// Outer frame
	drawRectOutline(screen, vpX-2, vpY-2, vpW+4, vpH+4, palette.accentColor)
	// Corner ornaments (L-shaped marks)
	drawCornerOrnament(screen, vpX-2, vpY-2, 1, 1, palette.accentColor)
	drawCornerOrnament(screen, vpX+vpW+1, vpY-2, -1, 1, palette.accentColor)
	drawCornerOrnament(screen, vpX-2, vpY+vpH+1, 1, -1, palette.accentColor)
	drawCornerOrnament(screen, vpX+vpW+1, vpY+vpH+1, -1, -1, palette.accentColor)
}

// drawCornerOrnament draws a small L-shaped decorative mark at a corner.
// dirX/dirY: 1 or -1 to indicate which direction the L extends.
func drawCornerOrnament(screen *ebiten.Image, x, y, dirX, dirY int, c color.RGBA) {
	drawLine(screen, x, y, x+dirX*6, y, c)
	drawLine(screen, x, y, x, y+dirY*6, c)
}

// drawFloorWaterHint draws wavy blue lines to indicate water on the floor.
func drawFloorWaterHint(screen *ebiten.Image, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	waterColor := color.RGBA{R: 50, G: 80, B: 180, A: 120}
	frame := flickerFrame()
	// 2-3 wavy lines made of short alternating segments
	for row := 0; row < 3; row++ {
		ly := y + h/4 + row*(h/4)
		for seg := 0; seg < w; seg += 6 {
			yOff := 0
			if (seg/6+row+frame)%2 == 0 {
				yOff = 1
			}
			sx := x + seg
			sw := min(6, w-seg)
			drawLine(screen, sx, ly+yOff, sx+sw, ly+yOff, waterColor)
		}
	}
}

// drawFloorTrapHint draws a crosshatch X pattern to indicate a trap on the floor.
func drawFloorTrapHint(screen *ebiten.Image, x, y, w, h int) {
	if w <= 4 || h <= 4 {
		return
	}
	trapColor := color.RGBA{R: 180, G: 80, B: 50, A: 150}
	cx := x + w/2
	cy := y + h/2
	sz := min(w, h) / 4
	drawLine(screen, cx-sz, cy-sz, cx+sz, cy+sz, trapColor)
	drawLine(screen, cx+sz, cy-sz, cx-sz, cy+sz, trapColor)
}

// drawFloorChestHint draws a small rectangular silhouette for a chest/object.
func drawFloorChestHint(screen *ebiten.Image, x, y, w, h int) {
	if w <= 4 || h <= 4 {
		return
	}
	chestColor := color.RGBA{R: 160, G: 130, B: 60, A: 160}
	cx := x + w/2
	cy := y + h*2/3
	cw := min(w/4, 12)
	ch := min(h/6, 8)
	drawRect(screen, cx-cw/2, cy-ch/2, cw, ch, chestColor)
	drawRectOutline(screen, cx-cw/2, cy-ch/2, cw, ch, brightenColor(chestColor, 20))
}
