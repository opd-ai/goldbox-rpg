package game

import (
	"fmt"
	"time"
)

// Duration represents a game time duration
// Duration represents time duration in a game context, combining different time measurements.
// It can track duration in rounds, turns, and real-world time simultaneously.
//
// Fields:
//   - Rounds: Number of combat/game rounds the duration lasts
//   - Turns: Number of player/character turns the duration lasts
//   - RealTime: Actual real-world time duration (uses time.Duration)
//
// The zero value represents an instant/immediate duration with no lasting effect.
// All fields are optional and can be combined - e.g. "2 rounds and 30 seconds"
// Moved from: effects.go
type Duration struct {
	Rounds   int           `yaml:"duration_rounds"`
	Turns    int           `yaml:"duration_turns"`
	RealTime time.Duration `yaml:"duration_real"`
}

// NewDuration creates a new Duration with the specified parameters.
// Moved from: effects.go
func NewDuration(rounds, turns int, realTime time.Duration) Duration {
	return Duration{
		Rounds:   rounds,
		Turns:    turns,
		RealTime: realTime,
	}
}

// IsExpired checks if the duration has elapsed.
// For game purposes, this checks if all duration components are zero or negative.
// Moved from: effects.go
func (d Duration) IsExpired() bool {
	return d.Rounds <= 0 && d.Turns <= 0 && d.RealTime <= 0
}

// String returns a human-readable representation of the duration.
// Moved from: effects.go
func (d Duration) String() string {
	if d.Rounds > 0 {
		return fmt.Sprintf("%d rounds", d.Rounds)
	}
	if d.Turns > 0 {
		return fmt.Sprintf("%d turns", d.Turns)
	}
	if d.RealTime > 0 {
		return d.RealTime.String()
	}
	return "instant"
}
