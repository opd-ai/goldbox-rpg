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
	if p == nil {
		return nil
	}

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

// StartQuest adds a new quest to the player's quest log and marks it as active.
// This method is thread-safe and validates that the quest doesn't already exist.
//
// Parameters:
//   - quest: The Quest object to add to the player's quest log
//
// Returns:
//   - error: Returns error if quest is invalid or already exists in quest log
//
// The method performs the following validations:
// - Quest ID must not be empty
// - Quest must not already exist in player's quest log
// - Quest status is automatically set to QuestActive
func (p *Player) StartQuest(quest Quest) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if quest.ID == "" {
		return fmt.Errorf("quest ID cannot be empty")
	}

	// Check if quest already exists
	for _, existingQuest := range p.QuestLog {
		if existingQuest.ID == quest.ID {
			return fmt.Errorf("quest %s already exists in quest log", quest.ID)
		}
	}

	// Set quest as active and add to quest log
	quest.Status = QuestActive
	p.QuestLog = append(p.QuestLog, quest)

	return nil
}

// CompleteQuest marks a quest as completed and processes its rewards.
// This method finds the quest by ID, validates it can be completed, and processes rewards.
//
// Parameters:
//   - questID: The unique identifier of the quest to complete
//
// Returns:
//   - []QuestReward: Slice of rewards granted for completing the quest
//   - error: Returns error if quest not found, already completed, or cannot be completed
//
// The method performs the following operations:
// - Validates quest exists and is active
// - Checks all objectives are completed
// - Marks quest as completed
// - Returns quest rewards for processing
func (p *Player) CompleteQuest(questID string) ([]QuestReward, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if questID == "" {
		return nil, fmt.Errorf("quest ID cannot be empty")
	}

	// Find quest in quest log
	for i, quest := range p.QuestLog {
		if quest.ID == questID {
			if quest.Status == QuestCompleted {
				return nil, fmt.Errorf("quest %s is already completed", questID)
			}
			if quest.Status != QuestActive {
				return nil, fmt.Errorf("quest %s is not active", questID)
			}

			// Check if all objectives are completed
			for _, objective := range quest.Objectives {
				if !objective.Completed {
					return nil, fmt.Errorf("quest %s cannot be completed: objective '%s' is not finished", questID, objective.Description)
				}
			}

			// Mark quest as completed
			p.QuestLog[i].Status = QuestCompleted

			return quest.Rewards, nil
		}
	}

	return nil, fmt.Errorf("quest %s not found in quest log", questID)
}

// UpdateQuestObjective updates the progress of a specific objective within a quest.
// This method is thread-safe and handles objective completion automatically.
//
// Parameters:
//   - questID: The unique identifier of the quest containing the objective
//   - objectiveIndex: The index of the objective to update (0-based)
//   - progress: The new progress value for the objective
//
// Returns:
//   - error: Returns error if quest not found, objective index invalid, or quest not active
//
// The method performs the following operations:
// - Validates quest exists and is active
// - Validates objective index is within bounds
// - Updates objective progress and completion status
// - Progress cannot exceed the required amount
func (p *Player) UpdateQuestObjective(questID string, objectiveIndex int, progress int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if questID == "" {
		return fmt.Errorf("quest ID cannot be empty")
	}
	if progress < 0 {
		return fmt.Errorf("progress cannot be negative")
	}

	// Find quest in quest log
	for i, quest := range p.QuestLog {
		if quest.ID == questID {
			if quest.Status != QuestActive {
				return fmt.Errorf("quest %s is not active", questID)
			}

			// Validate objective index
			if objectiveIndex < 0 || objectiveIndex >= len(quest.Objectives) {
				return fmt.Errorf("objective index %d is out of bounds for quest %s", objectiveIndex, questID)
			}

			// Update objective progress
			objective := &p.QuestLog[i].Objectives[objectiveIndex]
			if progress >= objective.Required {
				objective.Progress = objective.Required
				objective.Completed = true
			} else {
				objective.Progress = progress
				objective.Completed = false
			}

			return nil
		}
	}

	return fmt.Errorf("quest %s not found in quest log", questID)
}

// FailQuest marks a quest as failed, preventing completion but keeping it in the log.
// This method is thread-safe and handles quest state transitions.
//
// Parameters:
//   - questID: The unique identifier of the quest to fail
//
// Returns:
//   - error: Returns error if quest not found or already completed/failed
//
// Failed quests remain in the quest log for reference but cannot be completed.
func (p *Player) FailQuest(questID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if questID == "" {
		return fmt.Errorf("quest ID cannot be empty")
	}

	// Find quest in quest log
	for i, quest := range p.QuestLog {
		if quest.ID == questID {
			if quest.Status == QuestCompleted {
				return fmt.Errorf("quest %s is already completed and cannot be failed", questID)
			}
			if quest.Status == QuestFailed {
				return fmt.Errorf("quest %s is already failed", questID)
			}

			// Mark quest as failed
			p.QuestLog[i].Status = QuestFailed
			return nil
		}
	}

	return fmt.Errorf("quest %s not found in quest log", questID)
}

// GetQuest retrieves a specific quest from the player's quest log by ID.
// This method is thread-safe and returns a copy of the quest to prevent external modification.
//
// Parameters:
//   - questID: The unique identifier of the quest to retrieve
//
// Returns:
//   - *Quest: Pointer to a copy of the quest, or nil if not found
//   - error: Returns error if quest ID is empty or quest not found
func (p *Player) GetQuest(questID string) (*Quest, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if questID == "" {
		return nil, fmt.Errorf("quest ID cannot be empty")
	}

	for _, quest := range p.QuestLog {
		if quest.ID == questID {
			// Return a copy to prevent external modification
			questCopy := quest
			return &questCopy, nil
		}
	}

	return nil, fmt.Errorf("quest %s not found in quest log", questID)
}

// GetActiveQuests returns all quests that are currently active.
// This method is thread-safe and returns copies of quests to prevent external modification.
//
// Returns:
//   - []Quest: Slice containing copies of all active quests
//
// Active quests are those with status QuestActive that can still be progressed and completed.
func (p *Player) GetActiveQuests() []Quest {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var activeQuests []Quest
	for _, quest := range p.QuestLog {
		if quest.Status == QuestActive {
			activeQuests = append(activeQuests, quest)
		}
	}

	return activeQuests
}

// GetCompletedQuests returns all quests that have been completed.
// This method is thread-safe and returns copies of quests to prevent external modification.
//
// Returns:
//   - []Quest: Slice containing copies of all completed quests
//
// Completed quests show the player's progression history and achievements.
func (p *Player) GetCompletedQuests() []Quest {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var completedQuests []Quest
	for _, quest := range p.QuestLog {
		if quest.Status == QuestCompleted {
			completedQuests = append(completedQuests, quest)
		}
	}

	return completedQuests
}

// GetQuestLog returns a copy of the player's complete quest log.
// This method is thread-safe and returns copies to prevent external modification.
//
// Returns:
//   - []Quest: Slice containing copies of all quests in the quest log
//
// The quest log includes quests of all statuses: active, completed, and failed.
func (p *Player) GetQuestLog() []Quest {
	p.mu.RLock()
	defer p.mu.RUnlock()

	questLog := make([]Quest, len(p.QuestLog))
	copy(questLog, p.QuestLog)
	return questLog
}
