package pcg

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// QuestGeneratorImpl implements the QuestGenerator interface for procedural quest creation
// Creates engaging quests with varied objectives, balanced rewards, and meaningful narrative context
type QuestGeneratorImpl struct {
	version string
	logger  *logrus.Logger
	rng     *rand.Rand
}

// NewQuestGenerator creates a new quest generator instance
func NewQuestGenerator(logger *logrus.Logger) *QuestGeneratorImpl {
	if logger == nil {
		logger = logrus.New()
	}

	return &QuestGeneratorImpl{
		version: "1.0.0",
		logger:  logger,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Generate creates quests based on the provided parameters
// Returns generated quests with complete objectives and balanced rewards
func (qg *QuestGeneratorImpl) Generate(ctx context.Context, params GenerationParams) (interface{}, error) {
	if err := qg.Validate(params); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Use seed for deterministic generation
	rng := rand.New(rand.NewSource(params.Seed))
	qg.rng = rng

	questParams, ok := params.Constraints["quest_params"].(QuestParams)
	if !ok {
		// Use default parameters
		questParams = QuestParams{
			GenerationParams: params,
			QuestType:        QuestTypeFetch, // Default to fetch quest
			MinObjectives:    1,
			MaxObjectives:    3,
			RewardTier:       RarityCommon,
			Narrative:        NarrativeLinear,
		}
	}

	// Generate single quest
	quest, err := qg.GenerateQuest(ctx, questParams.QuestType, questParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quest: %w", err)
	}

	qg.logger.WithFields(logrus.Fields{
		"quest_id":   quest.ID,
		"quest_type": questParams.QuestType,
		"objectives": len(quest.Objectives),
		"rewards":    len(quest.Rewards),
	}).Info("Generated quest successfully")

	return quest, nil
}

// GenerateQuest creates a single quest with the specified type and parameters
func (qg *QuestGeneratorImpl) GenerateQuest(ctx context.Context, questType QuestType, params QuestParams) (*game.Quest, error) {
	// Validate context
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Generate unique quest ID
	questID := qg.generateQuestID(questType)

	// Generate quest title and description based on type
	title, description := qg.generateQuestNarrative(questType, params)

	// Generate objectives
	objectives, err := qg.GenerateObjectives(ctx, params.WorldState, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate objectives: %w", err)
	}

	// Convert QuestObjective to game.QuestObjective
	gameObjectives := make([]game.QuestObjective, len(objectives))
	for i, obj := range objectives {
		gameObjectives[i] = game.QuestObjective{
			Description: obj.Description,
			Progress:    obj.Progress,
			Required:    obj.Quantity,
			Completed:   obj.Complete,
		}
	}

	// Generate rewards
	rewards := qg.generateRewards(questType, params)

	quest := &game.Quest{
		ID:          questID,
		Title:       title,
		Description: description,
		Status:      game.QuestNotStarted,
		Objectives:  gameObjectives,
		Rewards:     rewards,
	}

	return quest, nil
}

// GenerateQuestChain creates a series of connected quests
func (qg *QuestGeneratorImpl) GenerateQuestChain(ctx context.Context, chainLength int, params QuestParams) ([]*game.Quest, error) {
	if chainLength <= 0 {
		return nil, fmt.Errorf("chain length must be positive, got %d", chainLength)
	}

	quests := make([]*game.Quest, 0, chainLength)
	baseParams := params

	for i := 0; i < chainLength; i++ {
		// Vary quest types in chain for diversity
		questType := qg.selectQuestTypeForChain(i, chainLength)

		// Adjust difficulty and rewards based on position in chain
		chainParams := baseParams
		chainParams.Difficulty = baseParams.Difficulty + i
		chainParams.PlayerLevel = baseParams.PlayerLevel + (i / 2) // Gradual level scaling

		// Modify seed for each quest while maintaining determinism
		chainParams.Seed = baseParams.Seed + int64(i*1000)

		quest, err := qg.GenerateQuest(ctx, questType, chainParams)
		if err != nil {
			return nil, fmt.Errorf("failed to generate quest %d in chain: %w", i+1, err)
		}

		// Modify quest title to indicate chain position
		if chainLength > 1 {
			quest.Title = fmt.Sprintf("%s (Part %d)", quest.Title, i+1)
		}

		quests = append(quests, quest)
	}

	qg.logger.WithFields(logrus.Fields{
		"chain_length": chainLength,
		"base_seed":    baseParams.Seed,
	}).Info("Generated quest chain successfully")

	return quests, nil
}

// GenerateObjectives creates quest objectives based on available content
func (qg *QuestGeneratorImpl) GenerateObjectives(ctx context.Context, world *game.World, params QuestParams) ([]QuestObjective, error) {
	objectiveCount := qg.rng.Intn(params.MaxObjectives-params.MinObjectives+1) + params.MinObjectives
	objectives := make([]QuestObjective, 0, objectiveCount)

	for i := 0; i < objectiveCount; i++ {
		objType := qg.selectObjectiveType(params.QuestType)
		objective := qg.generateSingleObjective(objType, params, i)
		objectives = append(objectives, objective)
	}

	return objectives, nil
}

// GetType returns the content type this generator produces
func (qg *QuestGeneratorImpl) GetType() ContentType {
	return ContentTypeQuests
}

// GetVersion returns the generator version for compatibility checking
func (qg *QuestGeneratorImpl) GetVersion() string {
	return qg.version
}

// Validate checks if the provided parameters are valid for this generator
func (qg *QuestGeneratorImpl) Validate(params GenerationParams) error {
	if params.Seed == 0 {
		return fmt.Errorf("seed must be non-zero")
	}

	if params.Difficulty < 1 || params.Difficulty > 20 {
		return fmt.Errorf("difficulty must be between 1 and 20, got %d", params.Difficulty)
	}

	if params.PlayerLevel < 1 || params.PlayerLevel > 20 {
		return fmt.Errorf("player level must be between 1 and 20, got %d", params.PlayerLevel)
	}

	// Validate quest-specific parameters if provided
	if questParams, ok := params.Constraints["quest_params"].(QuestParams); ok {
		if err := qg.validateQuestParams(questParams); err != nil {
			return fmt.Errorf("invalid quest parameters: %w", err)
		}
	}

	return nil
}

// validateQuestParams validates quest-specific parameters
func (qg *QuestGeneratorImpl) validateQuestParams(params QuestParams) error {
	if params.MinObjectives < 1 {
		return fmt.Errorf("min objectives must be at least 1, got %d", params.MinObjectives)
	}

	if params.MaxObjectives < params.MinObjectives {
		return fmt.Errorf("max objectives (%d) must be >= min objectives (%d)", params.MaxObjectives, params.MinObjectives)
	}

	if params.MaxObjectives > 10 {
		return fmt.Errorf("max objectives cannot exceed 10, got %d", params.MaxObjectives)
	}

	return nil
}

// generateQuestID creates a unique identifier for the quest
func (qg *QuestGeneratorImpl) generateQuestID(questType QuestType) string {
	timestamp := time.Now().Unix()
	randomSuffix := qg.rng.Intn(10000)
	return fmt.Sprintf("quest_%s_%d_%04d", questType, timestamp, randomSuffix)
}

// generateQuestNarrative creates title and description based on quest type
func (qg *QuestGeneratorImpl) generateQuestNarrative(questType QuestType, params QuestParams) (string, string) {
	templates := qg.getQuestTemplates(questType)
	template := templates[qg.rng.Intn(len(templates))]

	// Apply narrative style modifications
	switch params.Narrative {
	case NarrativeBranching:
		template.Description += " Multiple paths to completion are available."
	case NarrativeOpen:
		template.Description += " Approach this challenge however you see fit."
	case NarrativeEpisodic:
		template.Description += " This is part of a larger ongoing story."
	}

	return template.Title, template.Description
}

// questTemplate represents a template for quest narrative generation
type questTemplate struct {
	Title       string
	Description string
}

// getQuestTemplates returns narrative templates for each quest type
func (qg *QuestGeneratorImpl) getQuestTemplates(questType QuestType) []questTemplate {
	switch questType {
	case QuestTypeFetch:
		return []questTemplate{
			{
				Title:       "The Missing Artifact",
				Description: "A valuable artifact has gone missing and needs to be recovered. Search the area and bring it back safely.",
			},
			{
				Title:       "Gathering Supplies",
				Description: "Local merchants need specific items gathered from the wilderness. Collect the required materials and return them.",
			},
			{
				Title:       "The Lost Heirloom",
				Description: "A family heirloom has been lost in dangerous territory. Retrieve it and return it to its rightful owners.",
			},
		}
	case QuestTypeKill:
		return []questTemplate{
			{
				Title:       "Monster Extermination",
				Description: "Dangerous creatures threaten the local area. Eliminate the specified targets to restore safety.",
			},
			{
				Title:       "Bandit Elimination",
				Description: "A group of bandits has been terrorizing travelers. Hunt them down and put an end to their activities.",
			},
			{
				Title:       "The Corrupted Beast",
				Description: "A once-peaceful creature has been corrupted by dark magic. Put it out of its misery to restore balance.",
			},
		}
	case QuestTypeEscort:
		return []questTemplate{
			{
				Title:       "Safe Passage",
				Description: "An important individual needs safe escort through dangerous territory. Protect them during the journey.",
			},
			{
				Title:       "Merchant Caravan",
				Description: "A merchant caravan requires protection from bandits and monsters. Ensure they reach their destination safely.",
			},
			{
				Title:       "The Diplomatic Mission",
				Description: "An ambassador needs protection while traveling to negotiate peace. Guard them against any threats.",
			},
		}
	case QuestTypeExplore:
		return []questTemplate{
			{
				Title:       "Uncharted Territory",
				Description: "An unexplored region needs to be mapped and surveyed. Document the area and report your findings.",
			},
			{
				Title:       "The Ancient Ruins",
				Description: "Mysterious ruins have been discovered nearby. Explore them and uncover their secrets.",
			},
			{
				Title:       "Scouting Mission",
				Description: "Intelligence is needed about enemy movements in the area. Scout the region and gather information.",
			},
		}
	case QuestTypeDefend:
		return []questTemplate{
			{
				Title:       "Hold the Line",
				Description: "Enemy forces are approaching the settlement. Organize defenses and repel the attack.",
			},
			{
				Title:       "Protecting the Innocent",
				Description: "Civilians are in danger from an imminent threat. Establish a defensive perimeter and keep them safe.",
			},
			{
				Title:       "The Last Stand",
				Description: "The final battle approaches. Make your stand and protect everything you hold dear.",
			},
		}
	case QuestTypePuzzle:
		return []questTemplate{
			{
				Title:       "The Ancient Riddle",
				Description: "An ancient puzzle blocks progress deeper into mysterious ruins. Solve the riddle to proceed.",
			},
			{
				Title:       "The Locked Door",
				Description: "A complex mechanism bars the way forward. Decipher the pattern and unlock the passage.",
			},
			{
				Title:       "The Scholar's Challenge",
				Description: "A learned sage has posed an intellectual challenge. Use wit and wisdom to find the solution.",
			},
		}
	case QuestTypeDelivery:
		return []questTemplate{
			{
				Title:       "Urgent Message",
				Description: "Time-sensitive information must be delivered to its destination. Ensure the message arrives intact.",
			},
			{
				Title:       "Supply Run",
				Description: "Critical supplies need to be transported to an outpost. Deliver them before they're desperately needed.",
			},
			{
				Title:       "The Secret Package",
				Description: "A mysterious package requires discrete delivery. Transport it safely without asking questions.",
			},
		}
	case QuestTypeSurvival:
		return []questTemplate{
			{
				Title:       "Against the Elements",
				Description: "Harsh conditions threaten survival in the wilderness. Endure the challenges and emerge victorious.",
			},
			{
				Title:       "The Gauntlet",
				Description: "Navigate through a series of deadly traps and hazards. Only the skilled and careful will survive.",
			},
			{
				Title:       "Endurance Test",
				Description: "Prove your resilience by surviving in hostile territory for a specified duration.",
			},
		}
	case QuestTypeStory:
		return []questTemplate{
			{
				Title:       "The Hero's Journey",
				Description: "Embark on an epic adventure that will test your courage, wisdom, and strength. The fate of many depends on your choices.",
			},
			{
				Title:       "Unraveling the Mystery",
				Description: "Strange events have been occurring in the region. Investigate the truth behind the mysterious happenings.",
			},
			{
				Title:       "The Path of Destiny",
				Description: "Ancient prophecies speak of a chosen one. Discover if you are the one destined to fulfill this role.",
			},
		}
	default:
		return []questTemplate{
			{
				Title:       "Unknown Task",
				Description: "A mysterious task awaits completion. The details will become clear as you progress.",
			},
		}
	}
}

// selectQuestTypeForChain chooses appropriate quest types for quest chains
func (qg *QuestGeneratorImpl) selectQuestTypeForChain(position, totalLength int) QuestType {
	questTypes := []QuestType{
		QuestTypeFetch, QuestTypeKill, QuestTypeEscort,
		QuestTypeExplore, QuestTypeDefend, QuestTypeDelivery,
	}

	// First quest should be engaging but not too difficult
	if position == 0 {
		simpleTypes := []QuestType{QuestTypeFetch, QuestTypeDelivery, QuestTypeExplore}
		return simpleTypes[qg.rng.Intn(len(simpleTypes))]
	}

	// Final quest should be climactic
	if position == totalLength-1 {
		climacticTypes := []QuestType{QuestTypeKill, QuestTypeDefend, QuestTypeStory}
		return climacticTypes[qg.rng.Intn(len(climacticTypes))]
	}

	// Middle quests can be any type
	return questTypes[qg.rng.Intn(len(questTypes))]
}

// selectObjectiveType chooses objective types appropriate for the quest type
func (qg *QuestGeneratorImpl) selectObjectiveType(questType QuestType) string {
	switch questType {
	case QuestTypeFetch:
		types := []string{"collect", "retrieve", "gather"}
		return types[qg.rng.Intn(len(types))]
	case QuestTypeKill:
		types := []string{"eliminate", "defeat", "slay"}
		return types[qg.rng.Intn(len(types))]
	case QuestTypeEscort:
		types := []string{"protect", "escort", "guard"}
		return types[qg.rng.Intn(len(types))]
	case QuestTypeExplore:
		types := []string{"explore", "map", "discover"}
		return types[qg.rng.Intn(len(types))]
	case QuestTypeDefend:
		types := []string{"defend", "protect", "hold"}
		return types[qg.rng.Intn(len(types))]
	case QuestTypePuzzle:
		types := []string{"solve", "decipher", "unlock"}
		return types[qg.rng.Intn(len(types))]
	case QuestTypeDelivery:
		types := []string{"deliver", "transport", "carry"}
		return types[qg.rng.Intn(len(types))]
	case QuestTypeSurvival:
		types := []string{"survive", "endure", "withstand"}
		return types[qg.rng.Intn(len(types))]
	case QuestTypeStory:
		types := []string{"investigate", "discover", "uncover"}
		return types[qg.rng.Intn(len(types))]
	default:
		return "complete"
	}
}

// generateSingleObjective creates a single quest objective
func (qg *QuestGeneratorImpl) generateSingleObjective(objType string, params QuestParams, index int) QuestObjective {
	// Generate quantity based on difficulty and player level
	quantity := qg.calculateObjectiveQuantity(params, objType)

	// Generate target based on objective type
	target := qg.generateObjectiveTarget(objType, params)

	// Create description
	description := qg.generateObjectiveDescription(objType, target, quantity)

	return QuestObjective{
		ID:          fmt.Sprintf("obj_%d_%s", index+1, objType),
		Type:        objType,
		Description: description,
		Target:      target,
		Quantity:    quantity,
		Progress:    0,
		Complete:    false,
		Optional:    qg.shouldBeOptional(index, params),
		Conditions:  make(map[string]interface{}),
	}
}

// calculateObjectiveQuantity determines how many items/enemies/etc are needed
func (qg *QuestGeneratorImpl) calculateObjectiveQuantity(params QuestParams, objType string) int {
	baseQuantities := map[string]int{
		"collect":     3,
		"eliminate":   2,
		"escort":      1,
		"explore":     1,
		"defend":      1,
		"deliver":     1,
		"solve":       1,
		"survive":     1,
		"investigate": 1,
	}

	baseQuantity, exists := baseQuantities[objType]
	if !exists {
		baseQuantity = 1
	}

	// Scale based on difficulty and player level
	scalingFactor := 1.0 + (float64(params.Difficulty-1) * 0.1) + (float64(params.PlayerLevel-1) * 0.05)
	scaledQuantity := int(float64(baseQuantity) * scalingFactor)

	// Add some randomization
	variation := qg.rng.Intn(scaledQuantity/2 + 1)
	finalQuantity := scaledQuantity + variation

	// Ensure minimum of 1
	if finalQuantity < 1 {
		finalQuantity = 1
	}

	return finalQuantity
}

// generateObjectiveTarget creates appropriate targets for objectives
func (qg *QuestGeneratorImpl) generateObjectiveTarget(objType string, params QuestParams) string {
	switch objType {
	case "collect", "retrieve", "gather":
		items := []string{"ancient coins", "magical herbs", "crystal shards", "rare gems", "lost scrolls"}
		return items[qg.rng.Intn(len(items))]
	case "eliminate", "defeat", "slay":
		enemies := []string{"bandits", "goblins", "wolves", "corrupted beasts", "undead"}
		return enemies[qg.rng.Intn(len(enemies))]
	case "protect", "escort", "guard":
		targets := []string{"merchant", "diplomat", "scholar", "refugee", "noble"}
		return targets[qg.rng.Intn(len(targets))]
	case "explore", "map", "discover":
		locations := []string{"ancient ruins", "hidden caves", "forbidden forest", "mountain pass", "abandoned village"}
		return locations[qg.rng.Intn(len(locations))]
	case "defend", "hold":
		places := []string{"village", "outpost", "bridge", "gate", "sanctuary"}
		return places[qg.rng.Intn(len(places))]
	case "deliver", "transport", "carry":
		items := []string{"important message", "medical supplies", "ancient artifact", "diplomatic papers", "sacred relic"}
		return items[qg.rng.Intn(len(items))]
	case "solve", "decipher", "unlock":
		puzzles := []string{"ancient riddle", "magical lock", "cipher text", "mystic pattern", "forgotten language"}
		return puzzles[qg.rng.Intn(len(puzzles))]
	default:
		return "unknown target"
	}
}

// generateObjectiveDescription creates human-readable objective descriptions
func (qg *QuestGeneratorImpl) generateObjectiveDescription(objType, target string, quantity int) string {
	switch objType {
	case "collect", "retrieve", "gather":
		if quantity == 1 {
			return fmt.Sprintf("Collect %s", target)
		}
		return fmt.Sprintf("Collect %d %s", quantity, target)
	case "eliminate", "defeat", "slay":
		if quantity == 1 {
			return fmt.Sprintf("Defeat the %s", strings.TrimSuffix(target, "s"))
		}
		return fmt.Sprintf("Defeat %d %s", quantity, target)
	case "protect", "escort", "guard":
		return fmt.Sprintf("Safely escort the %s", target)
	case "explore", "map", "discover":
		return fmt.Sprintf("Explore the %s", target)
	case "defend", "hold":
		return fmt.Sprintf("Defend the %s", target)
	case "deliver", "transport", "carry":
		return fmt.Sprintf("Deliver the %s", target)
	case "solve", "decipher", "unlock":
		return fmt.Sprintf("Solve the %s", target)
	default:
		return fmt.Sprintf("Complete objective involving %s", target)
	}
}

// shouldBeOptional determines if an objective should be optional
func (qg *QuestGeneratorImpl) shouldBeOptional(index int, params QuestParams) bool {
	// First objective is never optional
	if index == 0 {
		return false
	}

	// 20% chance for objectives beyond the first to be optional
	return qg.rng.Float64() < 0.2
}

// generateRewards creates appropriate rewards for the quest
func (qg *QuestGeneratorImpl) generateRewards(questType QuestType, params QuestParams) []game.QuestReward {
	rewards := make([]game.QuestReward, 0, 3)

	// Always include experience reward
	expReward := qg.calculateExperienceReward(params)
	rewards = append(rewards, game.QuestReward{
		Type:  "exp",
		Value: expReward,
	})

	// Usually include gold reward
	if qg.rng.Float64() < 0.8 {
		goldReward := qg.calculateGoldReward(params)
		rewards = append(rewards, game.QuestReward{
			Type:  "gold",
			Value: goldReward,
		})
	}

	// Chance for item reward based on quest type and reward tier
	if qg.shouldIncludeItemReward(questType, params) {
		itemReward := qg.generateItemReward(params)
		rewards = append(rewards, itemReward)
	}

	return rewards
}

// calculateExperienceReward determines experience points for quest completion
func (qg *QuestGeneratorImpl) calculateExperienceReward(params QuestParams) int {
	baseExp := 100
	difficultyMultiplier := float64(params.Difficulty) * 0.5
	levelScaling := float64(params.PlayerLevel) * 1.2

	totalExp := int(float64(baseExp) * (1.0 + difficultyMultiplier + levelScaling))

	// Add some randomization
	variation := qg.rng.Intn(totalExp/4 + 1)
	return totalExp + variation
}

// calculateGoldReward determines gold reward for quest completion
func (qg *QuestGeneratorImpl) calculateGoldReward(params QuestParams) int {
	baseGold := 50
	difficultyMultiplier := float64(params.Difficulty) * 0.3
	levelScaling := float64(params.PlayerLevel) * 0.8

	totalGold := int(float64(baseGold) * (1.0 + difficultyMultiplier + levelScaling))

	// Add some randomization
	variation := qg.rng.Intn(totalGold/3 + 1)
	return totalGold + variation
}

// shouldIncludeItemReward determines if quest should have item rewards
func (qg *QuestGeneratorImpl) shouldIncludeItemReward(questType QuestType, params QuestParams) bool {
	// Higher tier quests more likely to have item rewards
	baseProbability := 0.3

	switch params.RewardTier {
	case RarityUncommon:
		baseProbability = 0.5
	case RarityRare:
		baseProbability = 0.7
	case RarityEpic:
		baseProbability = 0.85
	case RarityLegendary:
		baseProbability = 0.95
	}

	// Some quest types more likely to have items
	switch questType {
	case QuestTypeFetch, QuestTypeKill:
		baseProbability += 0.2
	case QuestTypeStory:
		baseProbability += 0.3
	}

	return qg.rng.Float64() < baseProbability
}

// generateItemReward creates an item reward for the quest
func (qg *QuestGeneratorImpl) generateItemReward(params QuestParams) game.QuestReward {
	// Generate a generic item ID based on reward tier
	itemTypes := []string{"sword", "armor", "ring", "potion", "scroll"}
	itemType := itemTypes[qg.rng.Intn(len(itemTypes))]

	tierSuffix := string(params.RewardTier)
	itemID := fmt.Sprintf("%s_%s_%d", itemType, tierSuffix, qg.rng.Intn(1000))

	return game.QuestReward{
		Type:   "item",
		Value:  1,
		ItemID: itemID,
	}
}
