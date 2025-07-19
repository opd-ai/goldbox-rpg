package game

import (
	"github.com/google/uuid"
)

// NewUID generates a unique identifier string using UUID v4.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
//
// Returns a 36-character UUID string (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
func NewUID() string {
	return uuid.NewString()
}

// isValidPosition checks if a given Position is valid within game bounds.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
//
// Parameters:
//   - pos: Position struct containing X, Y coordinates and Level number
//   - width: map width (max X+1)
//   - height: map height (max Y+1)
//   - maxLevel: number of levels (max Level+1)
//
// Returns:
//   - bool: true if position is valid (within bounds), false otherwise
func isValidPosition(pos Position, width, height, maxLevel int) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.Level >= 0 &&
		pos.X < width && pos.Y < height && pos.Level < maxLevel
}

// calculateLevel determines the character level based on experience points using a D&D-style progression system.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
//
// Parameters:
//   - exp: The total experience points (must be non-negative integer)
//
// Returns:
//   - An integer representing the character level based on experience thresholds
//
// Level thresholds:
//
//	Level 1: 0-1999 XP
//	Level 2: 2000-3999 XP
//	Level 3: 4000-7999 XP
//	Level 4: 8000-15999 XP
//	Level 5: 16000-31999 XP
//	Level 6: 32000-63999 XP
//	Level 7: 64000+ XP
//
// Notes:
//   - Returns level 0 for negative experience values
//   - Returns max level (7) for experience values above highest threshold
func calculateLevel(exp int64) int {
	// Handle negative experience values
	if exp < 0 {
		return 0
	}

	// Implement D&D-style level progression
	// Level 1: 0-1999 XP, Level 2: 2000-3999 XP, etc.
	levels := []int64{0, 2000, 4000, 8000, 16000, 32000, 64000}

	// Find the highest threshold that the experience meets or exceeds
	currentLevel := 1
	for i := 1; i < len(levels); i++ {
		if exp >= levels[i] {
			currentLevel = i + 1
		} else {
			break
		}
	}

	return currentLevel
}

// calculateHealthGain calculates the health points gained when a character levels up
// based on their character class and constitution score.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
func calculateHealthGain(class CharacterClass, constitution int) int {
	baseGain := map[CharacterClass]int{
		ClassFighter: 10,
		ClassMage:    4,
		ClassCleric:  8,
		ClassThief:   6,
		ClassRanger:  8,
		ClassPaladin: 10,
	}

	conBonus := (constitution - 10) / 2
	return baseGain[class] + conBonus
}

// minFloat returns the smaller of two float64 values.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxFloat returns the larger of two float64 values.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// maxInt returns the larger of two int values.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the smaller of two int values.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// clampFloat restricts a value between a minimum and maximum bound.
// This utility function ensures that val is within the range [minVal, maxVal].
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
func clampFloat(val, minVal, maxVal float64) float64 {
	return maxFloat(minVal, minFloat(val, maxVal))
}

// calculateMaxActionPoints determines the maximum action points for a character based on their level and dexterity.
//
// Thread Safety:
//   - This function is thread-safe and does not modify shared state.
//
// Parameters:
//   - level: The character's current level (must be at least 1)
//   - dexterity: The character's dexterity score
//
// Returns:
//   - int: The maximum action points for the given level and dexterity
//
// Notes:
//   - If level < 1, the function clamps it to 1. This may mask logic errors in calling code.
//     Consider validating level before calling this function.
func calculateMaxActionPoints(level, dexterity int) int {
	if level < 1 {
		level = 1
	}

	// Start with base action points
	basePoints := ActionPointsPerTurn

	// Add 1 action point for each odd level beyond 1
	// Odd levels: 3, 5, 7, 9, etc.
	// At level 3: (3-1)/2 = 1 bonus point
	// At level 5: (5-1)/2 = 2 bonus points
	// At level 7: (7-1)/2 = 3 bonus points
	bonusPoints := 0
	if level >= 3 {
		bonusPoints = (level - 1) / 2
	}

	// Add dexterity bonus
	dexterityBonus := 0
	if dexterity > 14 {
		dexterityBonus = 1
	}

	return basePoints + bonusPoints + dexterityBonus
}
