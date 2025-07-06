package game

// Quest represents a game quest with its properties and progress tracking.
// A quest consists of a unique identifier, title, description, current status,
// objectives that need to be completed, and rewards granted upon completion.
//
// Fields:
//   - ID: Unique string identifier for the quest
//   - Title: Display name shown to the player
//   - Description: Detailed explanation of the quest's story and goals
//   - Status: Current state of the quest (see QuestStatus type)
//   - Objectives: Slice of QuestObjective containing individual goals
//   - Rewards: Slice of QuestReward given when quest is complete
//
// Related types:
//   - QuestStatus: Enum defining possible quest states
//   - QuestObjective: Individual goals that must be completed
//   - QuestReward: Items/experience granted on completion
type Quest struct {
	ID          string           `yaml:"quest_id"`          // Unique quest identifier
	Title       string           `yaml:"quest_title"`       // Display title of the quest
	Description string           `yaml:"quest_description"` // Detailed quest description
	Status      QuestStatus      `yaml:"quest_status"`      // Current quest state
	Objectives  []QuestObjective `yaml:"quest_objectives"`  // List of quest goals
	Rewards     []QuestReward    `yaml:"quest_rewards"`     // Rewards for completion
}

// QuestStatus represents the current state of a quest in the game.
// It is implemented as an integer enumeration to track quest progression.
//
// QuestStatus values indicate whether a quest is:
// - Not started/available
// - In progress/active
// - Completed/finished
// - Failed/abandoned
//
// Related types:
// - Quest struct: Contains the QuestStatus field
// - QuestLog: Manages multiple quests and their statuses
type QuestStatus int

// QuestStatus constants are defined in constants.go
// QuestNotStarted indicates that a quest has not yet been started by the player.
// This is the initial state of any quest when first created or discovered.
// Related: QuestActive, QuestCompleted, QuestFailed.

// QuestObjective represents a specific task or goal within a quest that needs to be completed.
// It tracks the progress towards completion and maintains the completion status.
//
// Fields:
//   - Description: String describing what needs to be accomplished
//   - Progress: Current amount of progress made towards completion (must be >= 0)
//   - Required: Total amount needed to complete the objective (must be > 0)
//   - Completed: Boolean flag indicating if the objective is finished
//
// The Progress field should never exceed Required. When Progress equals or exceeds
// Required, Completed should be set to true.
//
// Related types:
//   - Quest (parent type containing objectives)
type QuestObjective struct {
	Description string `yaml:"objective_description"` // What needs to be done
	Progress    int    `yaml:"objective_progress"`    // Current completion amount
	Required    int    `yaml:"objective_required"`    // Amount needed for completion
	Completed   bool   `yaml:"objective_completed"`   // Whether objective is done
}

// QuestReward represents a reward that can be awarded to a player for completing a quest.
// It supports different types of rewards like gold, items, or experience points.
//
// Fields:
//   - Type: The type of the reward, must be one of: "gold", "item", "exp"
//   - Value: The quantity of the reward to give (amount of gold/exp, or number of items)
//   - ItemID: Optional reference ID for item rewards, required only when Type is "item"
//
// The reward is typically processed by the reward system which handles validation
// and distribution to players. See RewardSystem.ProcessReward() for implementation details.
type QuestReward struct {
	Type   string `yaml:"reward_type"`    // Type of reward (gold, item, exp)
	Value  int    `yaml:"reward_value"`   // Quantity or amount of reward
	ItemID string `yaml:"reward_item_id"` // Reference to reward item if applicable
}

// QuestProgress tracks the player's progression status for a specific quest.
// It maintains metrics like completion status, time investment and retry attempts.
//
// Fields:
//   - QuestID: Unique identifier string for the associated quest
//   - ObjectivesComplete: Number of objectives completed in the quest (non-negative integer)
//   - TimeSpent: Total time spent on quest in seconds (non-negative integer)
//   - Attempts: Number of times player has attempted the quest (non-negative integer)
//
// The struct is serializable via YAML for persistence.
// Related types:
//   - Quest (for quest definition details)
//   - QuestObjective (for individual objective tracking)
type QuestProgress struct {
	QuestID            string `yaml:"progress_quest_id"`        // Associated quest ID
	ObjectivesComplete int    `yaml:"progress_objectives_done"` // Number of completed objectives
	TimeSpent          int    `yaml:"progress_time_spent"`      // Time spent on quest
	Attempts           int    `yaml:"progress_attempts"`        // Number of attempts
}
