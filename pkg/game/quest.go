package game

// Quest represents a game quest/mission
// Contains all information about a quest including objectives and rewards
type Quest struct {
	ID          string           `yaml:"quest_id"`          // Unique quest identifier
	Title       string           `yaml:"quest_title"`       // Display title of the quest
	Description string           `yaml:"quest_description"` // Detailed quest description
	Status      QuestStatus      `yaml:"quest_status"`      // Current quest state
	Objectives  []QuestObjective `yaml:"quest_objectives"`  // List of quest goals
	Rewards     []QuestReward    `yaml:"quest_rewards"`     // Rewards for completion
}

// QuestStatus represents the current state of a quest
type QuestStatus int

const (
	QuestNotStarted QuestStatus = iota
	QuestActive
	QuestCompleted
	QuestFailed
)

// QuestObjective represents a single goal or task within a quest
type QuestObjective struct {
	Description string `yaml:"objective_description"` // What needs to be done
	Progress    int    `yaml:"objective_progress"`    // Current completion amount
	Required    int    `yaml:"objective_required"`    // Amount needed for completion
	Completed   bool   `yaml:"objective_completed"`   // Whether objective is done
}

// QuestReward represents a reward given upon quest completion
type QuestReward struct {
	Type   string `yaml:"reward_type"`    // Type of reward (gold, item, exp)
	Value  int    `yaml:"reward_value"`   // Quantity or amount of reward
	ItemID string `yaml:"reward_item_id"` // Reference to reward item if applicable
}

// QuestProgress tracks overall quest completion metrics
type QuestProgress struct {
	QuestID            string `yaml:"progress_quest_id"`        // Associated quest ID
	ObjectivesComplete int    `yaml:"progress_objectives_done"` // Number of completed objectives
	TimeSpent          int    `yaml:"progress_time_spent"`      // Time spent on quest
	Attempts           int    `yaml:"progress_attempts"`        // Number of attempts
}
