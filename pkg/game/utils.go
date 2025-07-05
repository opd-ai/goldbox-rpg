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
func calculateLevel(exp int) int {
	// Handle negative experience values
	if exp < 0 {
		return 0
	}

	// Implement D&D-style level progression
	// Level 1: 0-1999 XP, Level 2: 2000-3999 XP, etc.
	levels := []int{0, 2000, 4000, 8000, 16000, 32000, 64000}

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
