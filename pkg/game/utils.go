package game

import (
	"crypto/rand"
	"encoding/hex"
)

// NewUID generates a unique identifier string by creating a random 8-byte sequence
// and encoding it as a hexadecimal string.
//
// Returns a 16-character hexadecimal string representing the random bytes.
//
// Note: This function uses crypto/rand for secure random number generation.
// The probability of collision is low but not zero. For cryptographic purposes or
// when absolute uniqueness is required, consider using UUID instead.
//
// Related: encoding/hex.EncodeToString()
func NewUID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// isValidPosition checks if a given Position is valid within game bounds.
//
// Parameters:
//   - pos: Position struct containing X, Y coordinates and Level number
//
// Returns:
//   - bool: true if position is valid (non-negative coordinates), false otherwise
//
// Note: Currently only checks for non-negative values. May need to add upper bounds
// checking based on map/level size constraints.
//
// Related:
//   - Position struct
func isValidPosition(pos Position) bool {
	// Add your validation logic here
	// For example:
	return pos.X >= 0 && pos.Y >= 0 && pos.Level >= 0
}

// calculateLevel determines the character level based on experience points using a D&D-style progression system.
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
// This is a simple utility function for comparing floating point numbers.
// Note: This function handles basic float comparison with no special cases for NaN or Inf.
// Moved from: effectmanager.go
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxFloat returns the larger of two float64 values.
// This is a simple utility function for comparing floating point numbers.
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// maxInt returns the larger of two int values.
// Moved from: spatial_index.go
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the smaller of two int values.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// clampFloat restricts a value between a minimum and maximum bound.
// This utility function ensures that val is within the range [minVal, maxVal].
func clampFloat(val, minVal, maxVal float64) float64 {
	return maxFloat(minVal, minFloat(val, maxVal))
}

// calculateMaxActionPoints determines the maximum action points for a character based on their level and dexterity.
// Characters start with 2 action points at level 1 and gain an additional action point
// at odd levels (1, 3, 5, 7, 9, etc.), providing tactical progression as they advance.
// Additionally, characters with dexterity > 14 gain 1 bonus action point.
//
// Parameters:
//   - level: The character's current level (must be at least 1)
//   - dexterity: The character's dexterity score
//
// Returns:
//   - int: The maximum action points for the given level and dexterity
//
// Level progression:
//   - Level 1: 2 action points (base)
//   - Level 3: 3 action points (+1 for odd level)
//   - Level 5: 4 action points (+1 for odd level)
//   - Level 7: 5 action points (+1 for odd level)
//   - etc.
//
// Dexterity bonus:
//   - Dexterity > 14: +1 action point
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
