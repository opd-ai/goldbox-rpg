// Package wasmui provides the Ebitengine/WASM-based game UI client types.
// This file contains touch input and gesture recognition types.
package wasmui

import "math"

// Gesture detection thresholds (tick-based at ~60fps).
const (
	tapMaxDurationTicks       = 18 // ~300ms at 60fps
	tapMaxDistance            = 20 // pixels
	swipeMinDistance          = 50 // pixels
	swipeMaxDurationTicks     = 30 // ~500ms at 60fps
	pinchMinDistChange        = 30 // pixels
	longPressMinDurationTicks = 30 // ~500ms at 60fps
)

// GestureType represents recognized touch gestures.
type GestureType int

const (
	// GestureNone indicates no gesture detected.
	GestureNone GestureType = iota
	// GestureTap indicates a short touch with minimal movement.
	GestureTap
	// GestureSwipeUp indicates an upward swipe gesture.
	GestureSwipeUp
	// GestureSwipeDown indicates a downward swipe gesture.
	GestureSwipeDown
	// GestureSwipeLeft indicates a leftward swipe gesture.
	GestureSwipeLeft
	// GestureSwipeRight indicates a rightward swipe gesture.
	GestureSwipeRight
	// GesturePinchIn indicates a two-finger pinch-in (zoom out) gesture.
	GesturePinchIn
	// GesturePinchOut indicates a two-finger pinch-out (zoom in) gesture.
	GesturePinchOut
	// GestureLongPress indicates a sustained touch with minimal movement.
	GestureLongPress
)

// TouchPoint tracks state for a single touch contact.
type TouchPoint struct {
	ID        int
	StartX    int
	StartY    int
	CurrentX  int
	CurrentY  int
	StartTick int
}

// TouchState tracks multi-touch state for gesture recognition.
type TouchState struct {
	ActiveTouches    map[int]*TouchPoint
	LastGesture      GestureType
	LastGestureX     int
	LastGestureY     int
	InitialPinchDist float64
	CurrentTick      int
}

// NewTouchState creates a new touch state tracker.
func NewTouchState() *TouchState {
	return &TouchState{
		ActiveTouches: make(map[int]*TouchPoint),
	}
}

// classifyGesture determines the gesture type from a completed touch.
func (ts *TouchState) classifyGesture(tp *TouchPoint) {
	duration := ts.CurrentTick - tp.StartTick
	dx := tp.CurrentX - tp.StartX
	dy := tp.CurrentY - tp.StartY
	dist := math.Sqrt(float64(dx*dx + dy*dy))

	// Long press: small movement, long duration
	if dist < float64(tapMaxDistance) && duration >= longPressMinDurationTicks {
		ts.LastGesture = GestureLongPress
		ts.LastGestureX = tp.CurrentX
		ts.LastGestureY = tp.CurrentY
		return
	}

	// Tap: small movement, short duration
	if dist < float64(tapMaxDistance) && duration < tapMaxDurationTicks {
		ts.LastGesture = GestureTap
		ts.LastGestureX = tp.CurrentX
		ts.LastGestureY = tp.CurrentY
		return
	}

	// Swipe: significant directional movement within time window
	if dist >= float64(swipeMinDistance) && duration < swipeMaxDurationTicks {
		absDX := math.Abs(float64(dx))
		absDY := math.Abs(float64(dy))
		if absDX > absDY {
			if dx > 0 {
				ts.LastGesture = GestureSwipeRight
			} else {
				ts.LastGesture = GestureSwipeLeft
			}
		} else {
			if dy > 0 {
				ts.LastGesture = GestureSwipeDown
			} else {
				ts.LastGesture = GestureSwipeUp
			}
		}
		ts.LastGestureX = tp.StartX
		ts.LastGestureY = tp.StartY
	}
}

// CheckPinch returns the pinch gesture type if a two-finger pinch is active.
func (ts *TouchState) CheckPinch() GestureType {
	if len(ts.ActiveTouches) != 2 || ts.InitialPinchDist == 0 {
		return GestureNone
	}
	var points []*TouchPoint
	for _, tp := range ts.ActiveTouches {
		points = append(points, tp)
	}
	if len(points) != 2 {
		return GestureNone
	}
	currentDist := touchDistance(
		points[0].CurrentX, points[0].CurrentY,
		points[1].CurrentX, points[1].CurrentY,
	)
	delta := currentDist - ts.InitialPinchDist
	if delta > float64(pinchMinDistChange) {
		return GesturePinchOut
	}
	if delta < -float64(pinchMinDistChange) {
		return GesturePinchIn
	}
	return GestureNone
}

// HasTap returns true if a tap was detected this frame, with the position.
func (ts *TouchState) HasTap() (bool, int, int) {
	if ts.LastGesture == GestureTap {
		return true, ts.LastGestureX, ts.LastGestureY
	}
	return false, 0, 0
}

// HasSwipe returns true if a swipe was detected this frame, with the direction.
func (ts *TouchState) HasSwipe() (bool, GestureType) {
	switch ts.LastGesture {
	case GestureSwipeUp, GestureSwipeDown, GestureSwipeLeft, GestureSwipeRight:
		return true, ts.LastGesture
	}
	return false, GestureNone
}

// HasLongPress returns true if a long press was detected this frame.
func (ts *TouchState) HasLongPress() (bool, int, int) {
	if ts.LastGesture == GestureLongPress {
		return true, ts.LastGestureX, ts.LastGestureY
	}
	return false, 0, 0
}

// ActiveTouchCount returns the number of currently active touches.
func (ts *TouchState) ActiveTouchCount() int {
	return len(ts.ActiveTouches)
}

// SwipeDirection converts a swipe gesture to a movement direction string.
func SwipeDirection(gesture GestureType) string {
	switch gesture {
	case GestureSwipeUp:
		return "north"
	case GestureSwipeDown:
		return "south"
	case GestureSwipeLeft:
		return "west"
	case GestureSwipeRight:
		return "east"
	}
	return ""
}

// touchDistance calculates the Euclidean distance between two points.
func touchDistance(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Sqrt(dx*dx + dy*dy)
}

// PointInRect returns true if point (px, py) is inside the rectangle
// defined by (rx, ry, rw, rh) where (rx, ry) is the top-left corner and
// (rw, rh) are extents to the inclusive right/bottom edges (i.e. rx <= px <= rx+rw
// and ry <= py <= ry+rh). Callers should not treat rw/rh as half-open widths.
// Used for touch/click hit testing on UI buttons.
func PointInRect(px, py, rx, ry, rw, rh int) bool {
	return px >= rx && px <= rx+rw && py >= ry && py <= ry+rh
}
