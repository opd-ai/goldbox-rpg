//go:build js && wasm

package wasmui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// updateFromEbiten processes raw touch input from Ebiten and detects gestures.
// Call once per frame at the beginning of Update().
func (ts *TouchState) updateFromEbiten() {
	ts.LastGesture = GestureNone
	ts.CurrentTick++

	// Track newly pressed touches
	justPressed := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range justPressed {
		x, y := ebiten.TouchPosition(id)
		ts.ActiveTouches[int(id)] = &TouchPoint{
			ID:        int(id),
			StartX:    x,
			StartY:    y,
			CurrentX:  x,
			CurrentY:  y,
			StartTick: ts.CurrentTick,
		}
	}

	// Update positions for all active touches
	touchIDs := ebiten.AppendTouchIDs(nil)
	activeSet := make(map[int]bool)
	for _, id := range touchIDs {
		activeSet[int(id)] = true
		if tp, ok := ts.ActiveTouches[int(id)]; ok {
			tp.CurrentX, tp.CurrentY = ebiten.TouchPosition(id)
		}
	}

	// Track pinch distance when exactly two touches are active
	if len(ts.ActiveTouches) == 2 {
		var points []*TouchPoint
		for _, tp := range ts.ActiveTouches {
			points = append(points, tp)
		}
		if len(points) == 2 {
			dist := touchDistance(
				points[0].CurrentX, points[0].CurrentY,
				points[1].CurrentX, points[1].CurrentY,
			)
			if ts.InitialPinchDist == 0 {
				ts.InitialPinchDist = dist
			}
		}
	} else {
		ts.InitialPinchDist = 0
	}

	// Check for released touches and classify gestures
	for id, tp := range ts.ActiveTouches {
		if !activeSet[id] {
			ts.classifyGesture(tp)
			delete(ts.ActiveTouches, id)
		}
	}
}

// handleTouchInput processes touch events for the game.
// Touch taps are translated to clicks at the tap position.
func (g *Game) handleTouchInput() {
	if tapped, x, y := g.touchState.HasTap(); tapped {
		g.handleClick(x, y)
	}
}

// mouseWheelDelta returns the mouse wheel scroll delta this frame.
func mouseWheelDelta() (float64, float64) {
	return ebiten.Wheel()
}
