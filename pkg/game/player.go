package game

import "fmt"

// Player extends Character with player-specific functionality
// Contains all attributes and mechanics specific to player characters
type Player struct {
	Character   `yaml:",inline"` // Base character attributes
	Class       CharacterClass   `yaml:"player_class"`      // Character's chosen class
	Level       int              `yaml:"player_level"`      // Current experience level
	Experience  int              `yaml:"player_experience"` // Total experience points
	QuestLog    []Quest          `yaml:"player_quests"`     // Active and completed quests
	KnownSpells []Spell          `yaml:"player_spells"`     // Learned/available spells
}

// PlayerProgressData represents serializable player progress
type PlayerProgressData struct {
	CurrentLevel       int `yaml:"progress_level"`          // Current level
	ExperiencePoints   int `yaml:"progress_exp"`            // Total XP
	NextLevelThreshold int `yaml:"progress_next_level_exp"` // XP needed for next level
	CompletedQuests    int `yaml:"progress_quests_done"`    // Number of completed quests
	SpellsLearned      int `yaml:"progress_spells_known"`   // Number of known spells
}

// AddExperience safely adds experience points and handles level ups
func (p *Player) AddExperience(exp int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if exp < 0 {
		return fmt.Errorf("cannot add negative experience: %d", exp)
	}

	p.Experience += exp

	// Check for level up
	if newLevel := calculateLevel(p.Experience); newLevel > p.Level {
		return p.levelUp(newLevel)
	}

	return nil
}

func (p *Player) levelUp(newLevel int) error {
	oldLevel := p.Level
	p.Level = newLevel

	// Calculate and apply level up benefits
	healthGain := calculateHealthGain(p.Class, p.Constitution)
	p.MaxHP += healthGain
	p.HP += healthGain

	// Emit level up event (implementation depends on event system)
	emitLevelUpEvent(p.ID, oldLevel, newLevel)

	return nil
}
