package game

import (
	"fmt"
)

// Player extends Character with player-specific functionality
// Contains all attributes and mechanics specific to player characters
// Player represents a playable character in the game with additional attributes beyond
// the base Character type. It tracks progression elements like level, experience,
// quests and learned spells.
//
// The Player struct embeds the Character type to inherit basic attributes while adding
// RPG-specific fields for character advancement and gameplay mechanics.
//
// Fields:
//   - Character: Base character attributes (embedded)
//   - Class: The character's chosen class that determines available abilities
//   - Level: Current experience level of the player (1 or greater)
//   - Experience: Total experience points accumulated
//   - QuestLog: Slice of active and completed quests
//   - KnownSpells: Slice of spells the player has learned and can cast
//
// Related types:
//   - Character: Base character attributes
//   - CharacterClass: Available character classes
//   - Quest: Quest structure
//   - Spell: Spell structure
type Player struct {
	Character   `yaml:",inline"` // Base character attributes
	Class       CharacterClass   `yaml:"player_class"`      // Character's chosen class
	Level       int              `yaml:"player_level"`      // Current experience level
	Experience  int              `yaml:"player_experience"` // Total experience points
	QuestLog    []Quest          `yaml:"player_quests"`     // Active and completed quests
	KnownSpells []Spell          `yaml:"player_spells"`     // Learned/available spells
}

// Update updates the player's data based on the provided map of attributes.
// It safely updates player fields while maintaining data consistency.
//
// Parameters:
//   - playerData: Map containing field names and their new values
//
// Fields that can be updated:
//   - "class": CharacterClass
//   - "level": int
//   - "experience": int
//   - "hp": int
//   - "max_hp": int
//   - "strength": int
//   - "constitution": int
//   - "dexterity": int
//   - "intelligence": int
func (p *Player) Update(playerData map[string]interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if class, ok := playerData["class"].(CharacterClass); ok {
		p.Class = class
	}
	if level, ok := playerData["level"].(int); ok {
		p.Level = level
	}
	if exp, ok := playerData["experience"].(int); ok {
		p.Experience = exp
	}
	if hp, ok := playerData["hp"].(int); ok {
		p.HP = hp
	}
	if maxHP, ok := playerData["max_hp"].(int); ok {
		p.MaxHP = maxHP
	}
	if str, ok := playerData["strength"].(int); ok {
		p.Strength = str
	}
	if con, ok := playerData["constitution"].(int); ok {
		p.Constitution = con
	}
	if dex, ok := playerData["dexterity"].(int); ok {
		p.Dexterity = dex
	}
	if intel, ok := playerData["intelligence"].(int); ok {
		p.Intelligence = intel
	}
}

// Clone creates and returns a deep copy of the Player.
// This is useful for creating separate instances of a player for different sessions
// while preserving the original player data.
//
// Returns:
//   - *Player: A pointer to a new Player instance with copied data
func (p *Player) Clone() *Player {
	clone := &Player{
		Class:      p.Class,
		Level:      p.Level,
		Experience: p.Experience,
	}

	// Clone base Character data
	clone.Character = *p.Character.Clone()

	// Deep copy QuestLog
	clone.QuestLog = make([]Quest, len(p.QuestLog))
	copy(clone.QuestLog, p.QuestLog)

	// Deep copy KnownSpells
	clone.KnownSpells = make([]Spell, len(p.KnownSpells))
	copy(clone.KnownSpells, p.KnownSpells)

	return clone
}

// PublicData returns a struct containing non-sensitive player information that can be
// shared with other players or game systems. This includes basic character info
// and visible stats while excluding progression and private data.
//
// Returns:
//   - map[string]interface{}: A map containing the player's basic shareable info
func (p *Player) PublicData() map[string]interface{} {
	return map[string]interface{}{
		"name":         p.Name,
		"class":        p.Class,
		"hp":           p.HP,
		"max_hp":       p.MaxHP,
		"strength":     p.Strength,
		"constitution": p.Constitution,
	}
}

// PlayerProgressData represents the current progress and achievements of a player in the game.
// It keeps track of various metrics like level, experience points, and accomplishments.
//
// Fields:
//   - CurrentLevel: The player's current level in the game (must be >= 1)
//   - ExperiencePoints: Total accumulated experience points
//   - NextLevelThreshold: Experience points required to advance to next level
//   - CompletedQuests: Number of quests the player has finished
//   - SpellsLearned: Number of spells the player has mastered
//
// Related types:
//   - Use with Player struct to track overall player state
//   - Experience points calculation handled by LevelingSystem
type PlayerProgressData struct {
	CurrentLevel       int `yaml:"progress_level"`          // Current level
	ExperiencePoints   int `yaml:"progress_exp"`            // Total XP
	NextLevelThreshold int `yaml:"progress_next_level_exp"` // XP needed for next level
	CompletedQuests    int `yaml:"progress_quests_done"`    // Number of completed quests
	SpellsLearned      int `yaml:"progress_spells_known"`   // Number of known spells
}

// AddExperience safely adds experience points and handles level ups
// AddExperience adds the specified amount of experience points to the player and handles leveling up.
// It is thread-safe through mutex locking.
//
// Parameters:
//   - exp: Amount of experience points to add (must be non-negative)
//
// Returns:
//   - error: Returns nil on success, error if exp is negative or if levelUp fails
//
// Errors:
//   - Returns error if exp is negative
//   - Returns error from levelUp if leveling up fails
//
// Related:
//   - calculateLevel(): Used to determine if player should level up
//   - levelUp(): Called when experience gain triggers a level increase
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

// levelUp increases the player's level to the specified new level and applies corresponding stat increases.
// It updates the player's maximum and current HP based on their class and constitution,
// and emits a level up event to notify the game system.
//
// Parameters:
//   - newLevel: The target level to advance the player to (must be greater than current level)
//
// Returns:
//   - error: Returns nil if successful, or an error if the level up could not be completed
//
// Related:
//   - calculateHealthGain() - Calculates HP increase on level up
//   - emitLevelUpEvent() - Broadcasts level up event to game systems
//
// Note: This method does not validate if the new level is valid (greater than current).
// Caller must ensure proper level progression.
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

// Add this method to Player
// GetStats returns a copy of the player's current stats converted to float64 values.
// It creates a new Stats struct containing the player's health, max health,
// strength, dexterity and intelligence values.
//
// Returns:
//   - *Stats: A pointer to a new Stats struct containing the converted stat values
//
// Related types:
//   - Stats struct
func (p *Player) GetStats() *Stats {
	return &Stats{
		Health:       float64(p.HP),
		Mana:         float64(p.Intelligence),
		Strength:     float64(p.Strength),
		Dexterity:    float64(p.Dexterity),
		Intelligence: float64(p.Intelligence),
		MaxHealth:    float64(p.MaxHP),
		MaxMana:      float64(p.Intelligence),
		Defense:      0,
		Speed:        0,
	}
}
