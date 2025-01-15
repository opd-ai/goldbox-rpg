package game

// isValidPosition checks if a position is within world bounds
func isValidPosition(pos Position) bool {
	// Add your validation logic here
	// For example:
	return pos.X >= 0 && pos.Y >= 0 && pos.Level >= 0
}

// calculateLevel determines level based on experience points
func calculateLevel(exp int) int {
	// Implement D&D-style level progression
	// This is a simplified example:
	levels := []int{0, 2000, 4000, 8000, 16000, 32000, 64000}
	for level, requirement := range levels {
		if exp < requirement {
			return level
		}
	}
	return len(levels)
}

// calculateHealthGain determines HP increase on level up
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
