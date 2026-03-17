package wasmui

import (
	"testing"
)

func TestTouchDistance(t *testing.T) {
	tests := []struct {
		name             string
		x1, y1, x2, y2  int
		expected         float64
	}{
		{"same point", 0, 0, 0, 0, 0},
		{"horizontal", 0, 0, 3, 0, 3},
		{"vertical", 0, 0, 0, 4, 4},
		{"diagonal 3-4-5", 0, 0, 3, 4, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := touchDistance(tt.x1, tt.y1, tt.x2, tt.y2)
			if got != tt.expected {
				t.Errorf("touchDistance(%d,%d,%d,%d) = %f, want %f",
					tt.x1, tt.y1, tt.x2, tt.y2, got, tt.expected)
			}
		})
	}
}

func TestNewTouchState(t *testing.T) {
	ts := NewTouchState()
	if ts == nil {
		t.Fatal("NewTouchState() returned nil")
	}
	if len(ts.ActiveTouches) != 0 {
		t.Error("new touch state should have no active touches")
	}
	if ts.LastGesture != GestureNone {
		t.Errorf("new touch state gesture = %d, want GestureNone", ts.LastGesture)
	}
}

func TestClassifyGestureTap(t *testing.T) {
	ts := NewTouchState()
	ts.CurrentTick = 100
	tp := &TouchPoint{
		ID:        1,
		StartX:    100,
		StartY:    200,
		CurrentX:  105,
		CurrentY:  203,
		StartTick: 90, // 10 ticks < tapMaxDurationTicks (18)
	}
	ts.classifyGesture(tp)
	if ts.LastGesture != GestureTap {
		t.Errorf("expected GestureTap, got %d", ts.LastGesture)
	}
	if ts.LastGestureX != 105 || ts.LastGestureY != 203 {
		t.Errorf("gesture position = (%d, %d), want (105, 203)",
			ts.LastGestureX, ts.LastGestureY)
	}
}

func TestClassifyGestureSwipe(t *testing.T) {
	tests := []struct {
		name     string
		dx, dy   int
		expected GestureType
	}{
		{"swipe right", 60, 10, GestureSwipeRight},
		{"swipe left", -60, 10, GestureSwipeLeft},
		{"swipe down", 10, 60, GestureSwipeDown},
		{"swipe up", 10, -60, GestureSwipeUp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTouchState()
			ts.CurrentTick = 120
			tp := &TouchPoint{
				ID:        1,
				StartX:    100,
				StartY:    200,
				CurrentX:  100 + tt.dx,
				CurrentY:  200 + tt.dy,
				StartTick: 100, // 20 ticks < swipeMaxDurationTicks (30)
			}
			ts.classifyGesture(tp)
			if ts.LastGesture != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, ts.LastGesture)
			}
		})
	}
}

func TestClassifyGestureLongPress(t *testing.T) {
	ts := NewTouchState()
	ts.CurrentTick = 150
	tp := &TouchPoint{
		ID:        1,
		StartX:    100,
		StartY:    200,
		CurrentX:  102,
		CurrentY:  201,
		StartTick: 100, // 50 ticks > longPressMinDurationTicks (30)
	}
	ts.classifyGesture(tp)
	if ts.LastGesture != GestureLongPress {
		t.Errorf("expected GestureLongPress, got %d", ts.LastGesture)
	}
	if ts.LastGestureX != 102 || ts.LastGestureY != 201 {
		t.Errorf("long press position = (%d, %d), want (102, 201)",
			ts.LastGestureX, ts.LastGestureY)
	}
}

func TestClassifyGestureNone(t *testing.T) {
	ts := NewTouchState()
	ts.CurrentTick = 200
	// Medium movement, medium duration — doesn't match any gesture
	tp := &TouchPoint{
		ID:        1,
		StartX:    100,
		StartY:    200,
		CurrentX:  125, // 25px: too far for tap, too short for swipe
		CurrentY:  200,
		StartTick: 160, // 40 ticks: too long for tap and swipe
	}
	ts.classifyGesture(tp)
	if ts.LastGesture != GestureNone {
		t.Errorf("expected GestureNone, got %d", ts.LastGesture)
	}
}

func TestHasTap(t *testing.T) {
	ts := NewTouchState()

	ok, _, _ := ts.HasTap()
	if ok {
		t.Error("HasTap should be false when no gesture")
	}

	ts.LastGesture = GestureTap
	ts.LastGestureX = 50
	ts.LastGestureY = 100
	ok, x, y := ts.HasTap()
	if !ok {
		t.Error("HasTap should be true")
	}
	if x != 50 || y != 100 {
		t.Errorf("tap position = (%d, %d), want (50, 100)", x, y)
	}
}

func TestHasSwipe(t *testing.T) {
	ts := NewTouchState()

	ok, _ := ts.HasSwipe()
	if ok {
		t.Error("HasSwipe should be false when no gesture")
	}

	swipes := []GestureType{GestureSwipeUp, GestureSwipeDown, GestureSwipeLeft, GestureSwipeRight}
	for _, g := range swipes {
		ts.LastGesture = g
		ok, dir := ts.HasSwipe()
		if !ok {
			t.Errorf("HasSwipe should be true for gesture %d", g)
		}
		if dir != g {
			t.Errorf("swipe direction = %d, want %d", dir, g)
		}
	}

	// Non-swipe gestures should return false
	ts.LastGesture = GestureTap
	ok, _ = ts.HasSwipe()
	if ok {
		t.Error("HasSwipe should be false for GestureTap")
	}
}

func TestHasLongPress(t *testing.T) {
	ts := NewTouchState()

	ok, _, _ := ts.HasLongPress()
	if ok {
		t.Error("HasLongPress should be false when no gesture")
	}

	ts.LastGesture = GestureLongPress
	ts.LastGestureX = 200
	ts.LastGestureY = 300
	ok, x, y := ts.HasLongPress()
	if !ok {
		t.Error("HasLongPress should be true")
	}
	if x != 200 || y != 300 {
		t.Errorf("long press position = (%d, %d), want (200, 300)", x, y)
	}
}

func TestSwipeDirection(t *testing.T) {
	tests := []struct {
		gesture  GestureType
		expected string
	}{
		{GestureSwipeUp, "north"},
		{GestureSwipeDown, "south"},
		{GestureSwipeLeft, "west"},
		{GestureSwipeRight, "east"},
		{GestureNone, ""},
		{GestureTap, ""},
		{GestureLongPress, ""},
	}
	for _, tt := range tests {
		got := SwipeDirection(tt.gesture)
		if got != tt.expected {
			t.Errorf("SwipeDirection(%d) = %q, want %q", tt.gesture, got, tt.expected)
		}
	}
}

func TestActiveTouchCount(t *testing.T) {
	ts := NewTouchState()
	if ts.ActiveTouchCount() != 0 {
		t.Error("empty state should have 0 active touches")
	}

	ts.ActiveTouches[1] = &TouchPoint{ID: 1}
	if ts.ActiveTouchCount() != 1 {
		t.Errorf("active count = %d, want 1", ts.ActiveTouchCount())
	}

	ts.ActiveTouches[2] = &TouchPoint{ID: 2}
	if ts.ActiveTouchCount() != 2 {
		t.Errorf("active count = %d, want 2", ts.ActiveTouchCount())
	}
}

func TestCheckPinch(t *testing.T) {
	t.Run("no touches", func(t *testing.T) {
		ts := NewTouchState()
		if ts.CheckPinch() != GestureNone {
			t.Error("expected GestureNone with no touches")
		}
	})

	t.Run("one touch", func(t *testing.T) {
		ts := NewTouchState()
		ts.ActiveTouches[1] = &TouchPoint{CurrentX: 100, CurrentY: 100}
		ts.InitialPinchDist = 100
		if ts.CheckPinch() != GestureNone {
			t.Error("expected GestureNone with one touch")
		}
	})

	t.Run("pinch in", func(t *testing.T) {
		ts := NewTouchState()
		ts.ActiveTouches[1] = &TouchPoint{CurrentX: 100, CurrentY: 100}
		ts.ActiveTouches[2] = &TouchPoint{CurrentX: 120, CurrentY: 100}
		ts.InitialPinchDist = 100 // started far apart, now 20px apart
		if ts.CheckPinch() != GesturePinchIn {
			t.Errorf("expected GesturePinchIn, got %d", ts.CheckPinch())
		}
	})

	t.Run("pinch out", func(t *testing.T) {
		ts := NewTouchState()
		ts.ActiveTouches[1] = &TouchPoint{CurrentX: 50, CurrentY: 100}
		ts.ActiveTouches[2] = &TouchPoint{CurrentX: 200, CurrentY: 100}
		ts.InitialPinchDist = 50 // started close, now 150px apart
		if ts.CheckPinch() != GesturePinchOut {
			t.Errorf("expected GesturePinchOut, got %d", ts.CheckPinch())
		}
	})

	t.Run("no significant change", func(t *testing.T) {
		ts := NewTouchState()
		ts.ActiveTouches[1] = &TouchPoint{CurrentX: 100, CurrentY: 100}
		ts.ActiveTouches[2] = &TouchPoint{CurrentX: 200, CurrentY: 100}
		ts.InitialPinchDist = 95 // close to current dist of 100
		if ts.CheckPinch() != GestureNone {
			t.Errorf("expected GestureNone for small pinch delta, got %d", ts.CheckPinch())
		}
	})
}

func TestGestureTypeConstants(t *testing.T) {
	// Verify gesture types are distinct
	gestures := []GestureType{
		GestureNone, GestureTap, GestureSwipeUp, GestureSwipeDown,
		GestureSwipeLeft, GestureSwipeRight, GesturePinchIn,
		GesturePinchOut, GestureLongPress,
	}
	seen := make(map[GestureType]bool)
	for _, g := range gestures {
		if seen[g] {
			t.Errorf("duplicate gesture type value: %d", g)
		}
		seen[g] = true
	}
}

func TestMultipleTouchTracking(t *testing.T) {
	ts := NewTouchState()

	// Simulate adding three touches
	ts.ActiveTouches[0] = &TouchPoint{ID: 0, StartX: 10, StartY: 10, CurrentX: 10, CurrentY: 10, StartTick: 0}
	ts.ActiveTouches[1] = &TouchPoint{ID: 1, StartX: 50, StartY: 50, CurrentX: 50, CurrentY: 50, StartTick: 0}
	ts.ActiveTouches[2] = &TouchPoint{ID: 2, StartX: 90, StartY: 90, CurrentX: 90, CurrentY: 90, StartTick: 0}

	if ts.ActiveTouchCount() != 3 {
		t.Errorf("active count = %d, want 3", ts.ActiveTouchCount())
	}

	// Remove one touch and classify
	ts.CurrentTick = 10
	tp := ts.ActiveTouches[1]
	ts.classifyGesture(tp)
	delete(ts.ActiveTouches, 1)

	if ts.ActiveTouchCount() != 2 {
		t.Errorf("active count after removal = %d, want 2", ts.ActiveTouchCount())
	}
	if ts.LastGesture != GestureTap {
		t.Errorf("expected GestureTap for short stationary touch, got %d", ts.LastGesture)
	}
}
