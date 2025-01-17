package game

import (
	"encoding/json"
	"fmt"
)

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

// FromJSON implements GameObject.
// Subtle: this method shadows the method (Character).FromJSON of Player.Character.
func (p *Player) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

// GetDescription implements GameObject.
// Subtle: this method shadows the method (Character).GetDescription of Player.Character.
func (p *Player) GetDescription() string {
	return p.Description
}

// GetHealth implements GameObject.
func (p *Player) GetHealth() int {
	return p.HP
}

// GetID implements GameObject.
// Subtle: this method shadows the method (Character).GetID of Player.Character.
func (p *Player) GetID() string {
	return p.ID
}

// GetName implements GameObject.
// Subtle: this method shadows the method (Character).GetName of Player.Character.
func (p *Player) GetName() string {
	return p.Name
}

// GetPosition implements GameObject.
// Subtle: this method shadows the method (Character).GetPosition of Player.Character.
func (p *Player) GetPosition() Position {
	return p.Position
}

// GetTags implements GameObject.
// Subtle: this method shadows the method (Character).GetTags of Player.Character.
func (p *Player) GetTags() []string {
	return p.GetTags()
}

// IsActive implements GameObject.
// Subtle: this method shadows the method (Character).IsActive of Player.Character.
func (p *Player) IsActive() bool {
	return p.IsActive()
}

// IsObstacle implements GameObject.
func (p *Player) IsObstacle() bool {
	// Players are considered obstacles for movement/pathing
	return true
}

// SetHealth implements GameObject.
func (p *Player) SetHealth(health int) {
	p.HP = health
	// Ensure health doesn't go below 0
	if p.HP < 0 {
		p.HP = 0
	}
	// Optional: Cap health at max health
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
}

// SetPosition implements GameObject.
// Subtle: this method shadows the method (Character).SetPosition of Player.Character.
func (p *Player) SetPosition(pos Position) error {
	// Basic position validation
	if pos.X < 0 || pos.Y < 0 {
		return fmt.Errorf("invalid position: coordinates cannot be negative")
	}
	p.Position = pos
	return nil
}

// ToJSON implements GameObject.
// Subtle: this method shadows the method (Character).ToJSON of Player.Character.
func (p *Player) ToJSON() ([]byte, error) {
	return json.Marshal(p)
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

// Add this method to Player
func (p *Player) GetStats() *Stats {
	return &Stats{
		Health:       float64(p.HP),
		MaxHealth:    float64(p.MaxHP),
		Strength:     float64(p.Strength),
		Dexterity:    float64(p.Dexterity),
		Intelligence: float64(p.Intelligence),
	}
}
