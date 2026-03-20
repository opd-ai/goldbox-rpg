package game

import (
	"fmt"

	"github.com/sirupsen/logrus"
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
	Character   `yaml:",inline"` // Base character attributes (includes Class)
	Level       int              `yaml:"player_level"`      // Current experience level
	Experience  int64            `yaml:"player_experience"` // Total experience points (int64 to prevent overflow)
	QuestLog    []Quest          `yaml:"player_quests"`     // Active and completed quests
	KnownSpells []Spell          `yaml:"player_spells"`     // Learned/available spells
	SpellSlots  map[int]int      `yaml:"spell_slots"`       // Available spell slots per level (level -> count)
	UsedSlots   map[int]int      `yaml:"used_slots"`        // Used spell slots per level (level -> count)
}

// GetHP returns the player's current hit points.
// This method is thread-safe.
//
// Returns:
//   - int: The player's current HP
func (p *Player) GetHP() int {
	logrus.WithFields(logrus.Fields{
		"function":  "GetHP",
		"package":   "game",
		"player_id": p.ID,
	}).Debug("entering GetHP")

	p.mu.RLock()
	defer p.mu.RUnlock()

	hp := p.HP

	logrus.WithFields(logrus.Fields{
		"function":  "GetHP",
		"package":   "game",
		"player_id": p.ID,
		"hp":        hp,
	}).Debug("exiting GetHP")

	return hp
}

// SetHP sets the player's current hit points.
// This method is thread-safe and ensures HP doesn't exceed MaxHP or go below 0.
//
// Parameters:
//   - hp: The new HP value to set
func (p *Player) SetHP(hp int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if hp < 0 {
		p.HP = 0
	} else if hp > p.MaxHP {
		p.HP = p.MaxHP
	} else {
		p.HP = hp
	}
}

// GetMaxHP returns the player's maximum hit points.
// This method is thread-safe.
//
// Returns:
//   - int: The player's maximum HP
func (p *Player) GetMaxHP() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.MaxHP
}

// Update updates the player's data based on the provided map of attributes.
// It safely updates both Player-specific and underlying Character fields while maintaining data consistency.
//
// Parameters:
//   - playerData: Map containing field names and their new values
//
// Character fields that can be updated:
//   - "name": string (Character name)
//   - "description": string (Character description)
//   - "class": CharacterClass (Character class)
//   - "position_x": int (X coordinate)
//   - "position_y": int (Y coordinate)
//   - "position_level": int (Dungeon/map level)
//   - "position_facing": Direction (Facing direction)
//   - "strength": int (Strength attribute)
//   - "constitution": int (Constitution attribute)
//   - "dexterity": int (Dexterity attribute)
//   - "intelligence": int (Intelligence attribute)
//   - "wisdom": int (Wisdom attribute)
//   - "charisma": int (Charisma attribute)
//   - "hp": int (Current hit points)
//   - "max_hp": int (Maximum hit points)
//   - "armor_class": int (Armor class rating)
//   - "thac0": int (To Hit Armor Class 0)
//   - "gold": int (Currency amount)
//
// Player-specific fields that can be updated:
//   - "level": int (Player level)
//   - "experience": int (Experience points)
func (p *Player) Update(playerData map[string]interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.updateCharacterFields(playerData)
	p.updatePositionComponents(playerData)
	p.updateAttributes(playerData)
	p.updateCombatStats(playerData)
	p.updateEconomicData(playerData)
	p.updatePlayerSpecificFields(playerData)
}

// updateCharacterFields updates basic character identification and classification data.
func (p *Player) updateCharacterFields(playerData map[string]interface{}) {
	if name, ok := playerData["name"].(string); ok {
		p.Character.Name = name
	}
	if description, ok := playerData["description"].(string); ok {
		p.Character.Description = description
	}
	if class, ok := playerData["class"].(CharacterClass); ok {
		p.Character.Class = class
	}
}

// updatePositionComponents updates individual position components including coordinates and facing direction.
func (p *Player) updatePositionComponents(playerData map[string]interface{}) {
	if x, ok := playerData["position_x"].(int); ok {
		p.Character.Position.X = x
	}
	if y, ok := playerData["position_y"].(int); ok {
		p.Character.Position.Y = y
	}
	if level, ok := playerData["position_level"].(int); ok {
		p.Character.Position.Level = level
	}
	if facing, ok := playerData["position_facing"].(Direction); ok {
		p.Character.Position.Facing = facing
	}
}

// updateAttributes updates the six core character attributes (strength, dexterity, constitution, intelligence, wisdom, charisma).
func (p *Player) updateAttributes(playerData map[string]interface{}) {
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
	if wisdom, ok := playerData["wisdom"].(int); ok {
		p.Character.Wisdom = wisdom
	}
	if charisma, ok := playerData["charisma"].(int); ok {
		p.Character.Charisma = charisma
	}
}

// updateCombatStats updates combat-related statistics including health points, armor class, and THAC0.
func (p *Player) updateCombatStats(playerData map[string]interface{}) {
	if hp, ok := playerData["hp"].(int); ok {
		p.HP = hp
	}
	if maxHP, ok := playerData["max_hp"].(int); ok {
		p.MaxHP = maxHP
	}
	if ac, ok := playerData["armor_class"].(int); ok {
		p.Character.ArmorClass = ac
	}
	if thac0, ok := playerData["thac0"].(int); ok {
		p.Character.THAC0 = thac0
	}
}

// updateEconomicData updates the character's gold and other economic resources.
func (p *Player) updateEconomicData(playerData map[string]interface{}) {
	if gold, ok := playerData["gold"].(int); ok {
		p.Character.Gold = gold
	}
}

// updatePlayerSpecificFields updates player-specific data including level and experience with type compatibility handling.
func (p *Player) updatePlayerSpecificFields(playerData map[string]interface{}) {
	if level, ok := playerData["level"].(int); ok {
		p.Level = level
	}
	if exp, ok := playerData["experience"].(int64); ok {
		p.Experience = exp
	} else if exp, ok := playerData["experience"].(int); ok {
		// Handle backwards compatibility with int values
		p.Experience = int64(exp)
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

	// Deep copy SpellSlots
	if p.SpellSlots != nil {
		clone.SpellSlots = make(map[int]int, len(p.SpellSlots))
		for k, v := range p.SpellSlots {
			clone.SpellSlots[k] = v
		}
	}

	// Deep copy UsedSlots
	if p.UsedSlots != nil {
		clone.UsedSlots = make(map[int]int, len(p.UsedSlots))
		for k, v := range p.UsedSlots {
			clone.UsedSlots[k] = v
		}
	}

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
		"class":        p.Character.Class,
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
func (p *Player) AddExperience(exp int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if exp < 0 {
		return fmt.Errorf("cannot add negative experience: %d", exp)
	}

	// Check for integer overflow before adding experience
	if p.Experience > 0 && exp > 0 && p.Experience > (1<<63-1)-exp {
		return fmt.Errorf("experience addition would cause overflow: current=%d, adding=%d", p.Experience, exp)
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
// recalculates THAC0 based on class progression rates, and emits a level up event.
//
// Parameters:
//   - newLevel: The target level to advance the player to (must be greater than current level)
//
// Returns:
//   - error: Returns nil if successful, or an error if the level up could not be completed
//
// Related:
//   - calculateHealthGain() - Calculates HP increase on level up
//   - calculateTHAC0() - Calculates THAC0 based on class and level
//   - emitLevelUpEvent() - Broadcasts level up event to game systems
//
// Note: This method does not validate if the new level is valid (greater than current).
// Caller must ensure proper level progression.
func (p *Player) levelUp(newLevel int) error {
	oldLevel := p.Level
	p.Level = newLevel

	// Calculate and apply level up benefits
	healthGain := calculateHealthGain(p.Character.Class, p.Constitution)
	p.MaxHP += healthGain
	p.HP += healthGain

	// Update action points based on new level and dexterity
	newMaxActionPoints := calculateMaxActionPoints(newLevel, p.Character.Dexterity)
	p.Character.MaxActionPoints = newMaxActionPoints
	p.Character.ActionPoints = newMaxActionPoints // Restore to full on level up

	// Recalculate THAC0 based on class and new level
	p.Character.THAC0 = calculateTHAC0(p.Character.Class, newLevel)

	// Emit level up event (implementation depends on event system)
	emitLevelUpEvent(p.ID, oldLevel, newLevel)

	return nil
}

// calculateTHAC0 returns the THAC0 (To Hit Armor Class 0) value for a character
// based on their class and level. THAC0 starts at 20 and decreases as characters
// level up, with different classes improving at different rates.
//
// Improvement rates per level (based on classic D&D rules):
//   - Fighter, Paladin, Ranger: 1 point per level (best combat progression)
//   - Cleric: 2 points per 3 levels (~0.66 per level)
//   - Thief: 1 point per 2 levels (0.5 per level)
//   - Mage: 1 point per 3 levels (~0.33 per level, worst combat progression)
//
// Parameters:
//   - class: The character's class
//   - level: The character's current level
//
// Returns:
//   - int: The calculated THAC0 value (minimum of 1)
func calculateTHAC0(class CharacterClass, level int) int {
	const baseTHAC0 = 20

	// Calculate improvement based on class
	var improvement int
	switch class {
	case ClassFighter, ClassPaladin, ClassRanger:
		// Warriors improve by 1 per level
		improvement = level - 1
	case ClassCleric:
		// Clerics improve by 2 every 3 levels
		improvement = ((level - 1) * 2) / 3
	case ClassThief:
		// Thieves improve by 1 every 2 levels
		improvement = (level - 1) / 2
	case ClassMage:
		// Mages improve by 1 every 3 levels
		improvement = (level - 1) / 3
	default:
		// Default to cleric progression for unknown classes
		improvement = ((level - 1) * 2) / 3
	}

	thac0 := baseTHAC0 - improvement

	// THAC0 cannot go below 1
	if thac0 < 1 {
		thac0 = 1
	}

	return thac0
}

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
func (p *Player) UpdateQuestObjective(questID string, objectiveIndex, progress int) error {
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

	// Return a copy of all quests (active and completed)
	result := make([]Quest, len(p.QuestLog))
	copy(result, p.QuestLog)
	return result
}

// KnowsSpell checks if the player has learned a specific spell
// Returns true if the spell is in the player's KnownSpells list
func (p *Player) KnowsSpell(spellID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, spell := range p.KnownSpells {
		if spell.ID == spellID {
			return true
		}
	}
	return false
}

// LearnSpell adds a new spell to the player's known spells if they don't already know it
// Returns an error if the player cannot learn the spell due to class or level restrictions
func (p *Player) LearnSpell(spell Spell) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if player already knows the spell
	for _, knownSpell := range p.KnownSpells {
		if knownSpell.ID == spell.ID {
			return fmt.Errorf("player already knows spell: %s", spell.ID)
		}
	}

	// Check if player's class can cast spells
	if !p.canCastSpells() {
		return fmt.Errorf("class %s cannot cast spells", p.Class.String())
	}

	// Check if player's level is sufficient for the spell
	if p.Level < spell.Level {
		return fmt.Errorf("player level %d insufficient for spell level %d", p.Level, spell.Level)
	}

	// Add the spell to known spells
	p.KnownSpells = append(p.KnownSpells, spell)
	return nil
}

// canCastSpells determines if the player's class can cast spells
// Based on D&D-style classes where only certain classes are spellcasters
func (p *Player) canCastSpells() bool {
	switch p.Class {
	case ClassMage:
		return true
	case ClassCleric:
		return true
	case ClassPaladin:
		return p.Level >= 9 // Paladins get spells at level 9
	case ClassRanger:
		return p.Level >= 8 // Rangers get spells at level 8
	default:
		return false
	}
}

// InitializeSpellSlots sets up the spell slots based on player's class and level.
// This follows D&D 5e-style spell slot progression.
// Thread-safe.
func (p *Player) InitializeSpellSlots() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.SpellSlots = make(map[int]int)
	p.UsedSlots = make(map[int]int)

	if !p.canCastSpells() {
		return
	}

	// Calculate caster level based on class
	casterLevel := p.getCasterLevel()
	if casterLevel <= 0 {
		return
	}

	// Set spell slots based on caster level (simplified D&D 5e progression)
	p.SpellSlots = getSpellSlotsForLevel(casterLevel)
}

// getCasterLevel returns the effective caster level for spell slot calculation.
// Full casters (Mage, Cleric) use their full level.
// Half casters (Paladin, Ranger) use half their level (rounded down).
func (p *Player) getCasterLevel() int {
	switch p.Class {
	case ClassMage, ClassCleric:
		return p.Level
	case ClassPaladin:
		if p.Level >= 9 {
			return (p.Level - 8) / 2
		}
		return 0
	case ClassRanger:
		if p.Level >= 8 {
			return (p.Level - 7) / 2
		}
		return 0
	default:
		return 0
	}
}

// getSpellSlotsForLevel returns the spell slots available at a given caster level.
// Based on simplified D&D 5e spell slot progression.
// spellSlotTable defines spell slots by caster level. Each entry is a map from spell level to slot count.
var spellSlotTable = map[int]map[int]int{
	1:  {1: 2},
	2:  {1: 3},
	3:  {1: 4, 2: 2},
	4:  {1: 4, 2: 3},
	5:  {1: 4, 2: 3, 3: 2},
	6:  {1: 4, 2: 3, 3: 3},
	7:  {1: 4, 2: 3, 3: 3, 4: 1},
	8:  {1: 4, 2: 3, 3: 3, 4: 2},
	9:  {1: 4, 2: 3, 3: 3, 4: 3, 5: 1},
	10: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2},
}

// getSpellSlotsForLevel returns spell slots available for a given caster level.
func getSpellSlotsForLevel(casterLevel int) map[int]int {
	if casterLevel < 1 {
		return make(map[int]int)
	}
	// Clamp to max defined level
	if casterLevel > 10 {
		casterLevel = 10
	}
	// Copy table entry to avoid mutation
	base := spellSlotTable[casterLevel]
	result := make(map[int]int, len(base))
	for k, v := range base {
		result[k] = v
	}
	return result
}

// GetSpellSlots returns a copy of the spell slots map.
// Thread-safe.
func (p *Player) GetSpellSlots() map[int]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.SpellSlots == nil {
		return make(map[int]int)
	}

	result := make(map[int]int, len(p.SpellSlots))
	for k, v := range p.SpellSlots {
		result[k] = v
	}
	return result
}

// GetUsedSlots returns a copy of the used slots map.
// Thread-safe.
func (p *Player) GetUsedSlots() map[int]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.UsedSlots == nil {
		return make(map[int]int)
	}

	result := make(map[int]int, len(p.UsedSlots))
	for k, v := range p.UsedSlots {
		result[k] = v
	}
	return result
}

// HasSpellSlot checks if the player has an available spell slot at the given level.
// Thread-safe.
func (p *Player) HasSpellSlot(spellLevel int) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.SpellSlots == nil {
		return false
	}

	available := p.SpellSlots[spellLevel]
	used := 0
	if p.UsedSlots != nil {
		used = p.UsedSlots[spellLevel]
	}

	return available > used
}

// UseSpellSlot consumes a spell slot at the given level.
// Returns an error if no slot is available.
// Thread-safe.
func (p *Player) UseSpellSlot(spellLevel int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Cantrips (level 0) don't use slots
	if spellLevel == 0 {
		return nil
	}

	if p.SpellSlots == nil {
		return fmt.Errorf("no spell slots available")
	}

	available := p.SpellSlots[spellLevel]
	used := 0
	if p.UsedSlots != nil {
		used = p.UsedSlots[spellLevel]
	}

	if available <= used {
		return fmt.Errorf("no level %d spell slots remaining", spellLevel)
	}

	if p.UsedSlots == nil {
		p.UsedSlots = make(map[int]int)
	}
	p.UsedSlots[spellLevel]++

	return nil
}

// RestoreSpellSlots restores all used spell slots (e.g., after a long rest).
// Thread-safe.
func (p *Player) RestoreSpellSlots() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.UsedSlots = make(map[int]int)
}

// GetAvailableSlots returns the number of available (unused) slots at a given level.
// Thread-safe.
func (p *Player) GetAvailableSlots(spellLevel int) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.SpellSlots == nil {
		return 0
	}

	available := p.SpellSlots[spellLevel]
	used := 0
	if p.UsedSlots != nil {
		used = p.UsedSlots[spellLevel]
	}

	return available - used
}
