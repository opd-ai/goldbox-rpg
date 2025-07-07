package game

import (
	"testing"
	"time"
)

func TestNewDuration(t *testing.T) {
	tests := []struct {
		name     string
		rounds   int
		turns    int
		realTime time.Duration
		expected Duration
	}{
		{
			name:     "all values positive",
			rounds:   5,
			turns:    3,
			realTime: 30 * time.Second,
			expected: Duration{Rounds: 5, Turns: 3, RealTime: 30 * time.Second},
		},
		{
			name:     "zero values",
			rounds:   0,
			turns:    0,
			realTime: 0,
			expected: Duration{Rounds: 0, Turns: 0, RealTime: 0},
		},
		{
			name:     "only rounds",
			rounds:   10,
			turns:    0,
			realTime: 0,
			expected: Duration{Rounds: 10, Turns: 0, RealTime: 0},
		},
		{
			name:     "only turns",
			rounds:   0,
			turns:    7,
			realTime: 0,
			expected: Duration{Rounds: 0, Turns: 7, RealTime: 0},
		},
		{
			name:     "only real time",
			rounds:   0,
			turns:    0,
			realTime: 2 * time.Minute,
			expected: Duration{Rounds: 0, Turns: 0, RealTime: 2 * time.Minute},
		},
		{
			name:     "negative values",
			rounds:   -1,
			turns:    -5,
			realTime: -10 * time.Second,
			expected: Duration{Rounds: -1, Turns: -5, RealTime: -10 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewDuration(tt.rounds, tt.turns, tt.realTime)
			if result != tt.expected {
				t.Errorf("NewDuration(%d, %d, %v) = %v, want %v",
					tt.rounds, tt.turns, tt.realTime, result, tt.expected)
			}
		})
	}
}

func TestDuration_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		expected bool
	}{
		{
			name:     "all positive - not expired",
			duration: Duration{Rounds: 1, Turns: 1, RealTime: 1 * time.Second},
			expected: false,
		},
		{
			name:     "all zero - expired",
			duration: Duration{Rounds: 0, Turns: 0, RealTime: 0},
			expected: true,
		},
		{
			name:     "all negative - expired",
			duration: Duration{Rounds: -1, Turns: -1, RealTime: -1 * time.Second},
			expected: true,
		},
		{
			name:     "mixed positive and zero - not expired",
			duration: Duration{Rounds: 1, Turns: 0, RealTime: 0},
			expected: false,
		},
		{
			name:     "mixed zero and positive - not expired",
			duration: Duration{Rounds: 0, Turns: 1, RealTime: 0},
			expected: false,
		},
		{
			name:     "real time positive only - not expired",
			duration: Duration{Rounds: 0, Turns: 0, RealTime: 1 * time.Millisecond},
			expected: false,
		},
		{
			name:     "mixed negative and positive - not expired",
			duration: Duration{Rounds: -1, Turns: 1, RealTime: 0},
			expected: false,
		},
		{
			name:     "mixed negative and zero - not expired (positive real time)",
			duration: Duration{Rounds: -1, Turns: 0, RealTime: 1 * time.Second},
			expected: false,
		},
		{
			name:     "large positive values - not expired",
			duration: Duration{Rounds: 100, Turns: 50, RealTime: 1 * time.Hour},
			expected: false,
		},
		{
			name:     "large negative values - expired",
			duration: Duration{Rounds: -100, Turns: -50, RealTime: -1 * time.Hour},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.duration.IsExpired()
			if result != tt.expected {
				t.Errorf("Duration{%d, %d, %v}.IsExpired() = %v, want %v",
					tt.duration.Rounds, tt.duration.Turns, tt.duration.RealTime, result, tt.expected)
			}
		})
	}
}

func TestDuration_String(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		expected string
	}{
		{
			name:     "rounds only",
			duration: Duration{Rounds: 5, Turns: 0, RealTime: 0},
			expected: "5 rounds",
		},
		{
			name:     "single round",
			duration: Duration{Rounds: 1, Turns: 0, RealTime: 0},
			expected: "1 rounds",
		},
		{
			name:     "turns only",
			duration: Duration{Rounds: 0, Turns: 3, RealTime: 0},
			expected: "3 turns",
		},
		{
			name:     "single turn",
			duration: Duration{Rounds: 0, Turns: 1, RealTime: 0},
			expected: "1 turns",
		},
		{
			name:     "real time only - seconds",
			duration: Duration{Rounds: 0, Turns: 0, RealTime: 30 * time.Second},
			expected: "30s",
		},
		{
			name:     "real time only - minutes",
			duration: Duration{Rounds: 0, Turns: 0, RealTime: 2 * time.Minute},
			expected: "2m0s",
		},
		{
			name:     "real time only - complex",
			duration: Duration{Rounds: 0, Turns: 0, RealTime: 1*time.Hour + 30*time.Minute + 45*time.Second},
			expected: "1h30m45s",
		},
		{
			name:     "instant - all zero",
			duration: Duration{Rounds: 0, Turns: 0, RealTime: 0},
			expected: "instant",
		},
		{
			name:     "rounds priority over turns",
			duration: Duration{Rounds: 2, Turns: 5, RealTime: 0},
			expected: "2 rounds",
		},
		{
			name:     "rounds priority over real time",
			duration: Duration{Rounds: 3, Turns: 0, RealTime: 10 * time.Second},
			expected: "3 rounds",
		},
		{
			name:     "turns priority over real time",
			duration: Duration{Rounds: 0, Turns: 4, RealTime: 15 * time.Second},
			expected: "4 turns",
		},
		{
			name:     "rounds priority over all",
			duration: Duration{Rounds: 1, Turns: 2, RealTime: 20 * time.Second},
			expected: "1 rounds",
		},
		{
			name:     "negative rounds - shows instant",
			duration: Duration{Rounds: -2, Turns: 0, RealTime: 0},
			expected: "instant",
		},
		{
			name:     "negative turns - shows instant",
			duration: Duration{Rounds: 0, Turns: -3, RealTime: 0},
			expected: "instant",
		},
		{
			name:     "negative real time - shows instant",
			duration: Duration{Rounds: 0, Turns: 0, RealTime: -5 * time.Second},
			expected: "instant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.duration.String()
			if result != tt.expected {
				t.Errorf("Duration{%d, %d, %v}.String() = %q, want %q",
					tt.duration.Rounds, tt.duration.Turns, tt.duration.RealTime, result, tt.expected)
			}
		})
	}
}

// TestDuration_ZeroValue tests the zero value behavior
func TestDuration_ZeroValue(t *testing.T) {
	var d Duration

	// Zero value should be expired (instant)
	if !d.IsExpired() {
		t.Error("Zero value Duration should be expired")
	}

	// Zero value should return "instant"
	if d.String() != "instant" {
		t.Errorf("Zero value Duration.String() = %q, want %q", d.String(), "instant")
	}
}

// TestDuration_RealTimeFormats tests various time.Duration formats
func TestDuration_RealTimeFormats(t *testing.T) {
	tests := []struct {
		name     string
		realTime time.Duration
		expected string
	}{
		{
			name:     "nanoseconds",
			realTime: 500 * time.Nanosecond,
			expected: "500ns",
		},
		{
			name:     "microseconds",
			realTime: 750 * time.Microsecond,
			expected: "750µs",
		},
		{
			name:     "milliseconds",
			realTime: 250 * time.Millisecond,
			expected: "250ms",
		},
		{
			name:     "fractional seconds",
			realTime: 1500 * time.Millisecond,
			expected: "1.5s",
		},
		{
			name:     "hours and minutes",
			realTime: 3*time.Hour + 45*time.Minute,
			expected: "3h45m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := Duration{Rounds: 0, Turns: 0, RealTime: tt.realTime}
			result := duration.String()
			if result != tt.expected {
				t.Errorf("Duration with RealTime %v.String() = %q, want %q",
					tt.realTime, result, tt.expected)
			}
		})
	}
}

// TestDuration_EdgeCases tests edge cases and boundary conditions
func TestDuration_EdgeCases(t *testing.T) {
	t.Run("very large values", func(t *testing.T) {
		d := Duration{
			Rounds:   1000000,
			Turns:    999999,
			RealTime: 24 * time.Hour,
		}

		// Should prioritize rounds
		expected := "1000000 rounds"
		if result := d.String(); result != expected {
			t.Errorf("Large Duration.String() = %q, want %q", result, expected)
		}

		// Should not be expired
		if d.IsExpired() {
			t.Error("Large positive Duration should not be expired")
		}
	})

	t.Run("mixed positive and negative", func(t *testing.T) {
		d := Duration{
			Rounds:   -1,
			Turns:    5,
			RealTime: -10 * time.Second,
		}

		// Should not be expired (turns is positive)
		if d.IsExpired() {
			t.Error("Duration with positive turns should not be expired")
		}

		// Should prioritize turns (rounds is negative, so skipped)
		expected := "5 turns"
		if result := d.String(); result != expected {
			t.Errorf("Mixed Duration.String() = %q, want %q", result, expected)
		}
	})
}
