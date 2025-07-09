package quests

import (
	"context"
	"fmt"
	"math/rand"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// ObjectiveBasedGenerator creates quests using objective templates
type ObjectiveBasedGenerator struct {
	version            string
	objectiveTemplates map[pcg.QuestType][]*ObjectiveTemplate
	narrativeEngine    *NarrativeEngine
}

// ObjectiveTemplate defines the structure of quest objectives
type ObjectiveTemplate struct {
	Type         string   `yaml:"type"`
	Description  string   `yaml:"description"`
	Requirements []string `yaml:"requirements"`
	Targets      []string `yaml:"targets"`
	Quantities   [2]int   `yaml:"quantities"`
	Rewards      []string `yaml:"rewards"`
}

// NewObjectiveBasedGenerator creates a new objective-based quest generator
func NewObjectiveBasedGenerator() *ObjectiveBasedGenerator {
	obg := &ObjectiveBasedGenerator{
		version:            "1.0.0",
		objectiveTemplates: make(map[pcg.QuestType][]*ObjectiveTemplate),
		narrativeEngine:    NewNarrativeEngine(),
	}

	// Initialize default templates
	obg.initializeDefaultTemplates()

	return obg
}

// GetType returns the content type this generator produces
func (obg *ObjectiveBasedGenerator) GetType() pcg.ContentType {
	return pcg.ContentTypeQuests
}

// GetVersion returns the generator version for compatibility checking
func (obg *ObjectiveBasedGenerator) GetVersion() string {
	return obg.version
}

// Validate checks if the provided parameters are valid for this generator
func (obg *ObjectiveBasedGenerator) Validate(params pcg.GenerationParams) error {
	if params.Difficulty < 1 || params.Difficulty > 20 {
		return fmt.Errorf("difficulty must be between 1 and 20")
	}

	// Check constraint validity if provided
	if minObj, ok := params.Constraints["min_objectives"].(int); ok {
		if minObj < 1 {
			return fmt.Errorf("min_objectives must be at least 1")
		}
	}

	if maxObj, ok := params.Constraints["max_objectives"].(int); ok {
		if minObj, ok := params.Constraints["min_objectives"].(int); ok {
			if maxObj < minObj {
				return fmt.Errorf("max_objectives must be >= min_objectives")
			}
		}
	}

	return nil
}

// Generate implements the Generator interface
func (obg *ObjectiveBasedGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
	if err := obg.Validate(params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Create quest parameters from generation parameters
	questParams := pcg.QuestParams{
		GenerationParams: params,
		QuestType:        pcg.QuestTypeFetch, // Default quest type
		MinObjectives:    1,
		MaxObjectives:    3,
		RewardTier:       pcg.RarityCommon,
		Narrative:        pcg.NarrativeLinear,
		RequiredItems:    []string{},
		ForbiddenItems:   []string{},
	}

	// Extract quest type from constraints if provided
	if questTypeStr, ok := params.Constraints["quest_type"].(string); ok {
		questParams.QuestType = pcg.QuestType(questTypeStr)
	}

	// Extract objective counts from constraints if provided
	if minObj, ok := params.Constraints["min_objectives"].(int); ok {
		questParams.MinObjectives = minObj
	}
	if maxObj, ok := params.Constraints["max_objectives"].(int); ok {
		questParams.MaxObjectives = maxObj
	}

	// Extract reward tier from constraints if provided
	if rewardTierStr, ok := params.Constraints["reward_tier"].(string); ok {
		questParams.RewardTier = pcg.RarityTier(rewardTierStr)
	}

	return obg.GenerateQuest(ctx, questParams.QuestType, questParams)
}

// GenerateQuest creates a quest with objectives and narrative
func (obg *ObjectiveBasedGenerator) GenerateQuest(ctx context.Context, questType pcg.QuestType, params pcg.QuestParams) (*game.Quest, error) {
	// Create deterministic random generator from seed
	rng := rand.New(rand.NewSource(params.Seed))

	// Generate quest ID
	questID := fmt.Sprintf("quest_%d_%s", params.Seed, questType)

	// Generate objectives
	objectiveCount := rng.Intn(params.MaxObjectives-params.MinObjectives+1) + params.MinObjectives
	objectives, err := obg.generateObjectives(ctx, questType, objectiveCount, params, rng)
	if err != nil {
		return nil, fmt.Errorf("failed to generate objectives: %w", err)
	}

	// Generate narrative context
	narrative, err := obg.narrativeEngine.GenerateQuestNarrative(questType, objectives, params, rng)
	if err != nil {
		return nil, fmt.Errorf("failed to generate narrative: %w", err)
	}

	// Generate rewards
	rewards, err := obg.generateRewards(params.Difficulty, params.RewardTier, rng)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rewards: %w", err)
	}

	// Convert objectives to game format
	gameObjectives := make([]game.QuestObjective, len(objectives))
	for i, obj := range objectives {
		gameObjectives[i] = game.QuestObjective{
			Description: obj.Description,
			Progress:    0,
			Required:    obj.Quantity,
			Completed:   false,
		}
	}

	// Create the quest
	quest := &game.Quest{
		ID:          questID,
		Title:       narrative.Title,
		Description: narrative.Description,
		Status:      game.QuestNotStarted,
		Objectives:  gameObjectives,
		Rewards:     rewards,
	}

	return quest, nil
}

// GenerateQuestChain creates a series of connected quests
func (obg *ObjectiveBasedGenerator) GenerateQuestChain(ctx context.Context, chainLength int, params pcg.QuestParams) ([]*game.Quest, error) {
	if chainLength < 1 {
		return nil, fmt.Errorf("chain length must be at least 1")
	}

	quests := make([]*game.Quest, 0, chainLength)
	rng := rand.New(rand.NewSource(params.Seed))

	for i := 0; i < chainLength; i++ {
		// Create modified parameters for this quest in the chain
		chainParams := params
		chainParams.Seed = rng.Int63() // Derive new seed for each quest

		// Scale difficulty slightly for later quests in chain
		chainParams.Difficulty += i / 2
		if chainParams.Difficulty > 20 {
			chainParams.Difficulty = 20
		}

		quest, err := obg.GenerateQuest(ctx, params.QuestType, chainParams)
		if err != nil {
			return nil, fmt.Errorf("failed to generate quest %d in chain: %w", i+1, err)
		}

		// Modify quest title to indicate chain position
		if chainLength > 1 {
			quest.Title = fmt.Sprintf("%s (Part %d)", quest.Title, i+1)
		}

		quests = append(quests, quest)
	}

	return quests, nil
}

// GenerateObjectives creates quest objectives based on available content
func (obg *ObjectiveBasedGenerator) GenerateObjectives(ctx context.Context, world *game.World, params pcg.QuestParams) ([]pcg.QuestObjective, error) {
	rng := rand.New(rand.NewSource(params.Seed))
	objectiveCount := rng.Intn(params.MaxObjectives-params.MinObjectives+1) + params.MinObjectives

	return obg.generateObjectives(ctx, params.QuestType, objectiveCount, params, rng)
}

// generateObjectives creates specific quest objectives
func (obg *ObjectiveBasedGenerator) generateObjectives(ctx context.Context, questType pcg.QuestType, count int, params pcg.QuestParams, rng *rand.Rand) ([]pcg.QuestObjective, error) {
	templates, exists := obg.objectiveTemplates[questType]
	if !exists || len(templates) == 0 {
		return nil, fmt.Errorf("no objective templates available for quest type: %s", questType)
	}

	objectives := make([]pcg.QuestObjective, 0, count)

	for i := 0; i < count; i++ {
		// Select random template
		template := templates[rng.Intn(len(templates))]

		// Determine quantity based on difficulty and template
		minQty, maxQty := template.Quantities[0], template.Quantities[1]
		if maxQty <= minQty {
			maxQty = minQty + 1
		}

		// Scale quantity based on difficulty
		difficultyScale := float64(params.Difficulty) / 10.0
		scaledMax := int(float64(maxQty) * difficultyScale)
		if scaledMax < minQty {
			scaledMax = minQty
		}

		quantity := rng.Intn(scaledMax-minQty+1) + minQty

		// Select target
		target := ""
		if len(template.Targets) > 0 {
			target = template.Targets[rng.Intn(len(template.Targets))]
		}

		objective := pcg.QuestObjective{
			ID:          fmt.Sprintf("obj_%d_%d", params.Seed, i),
			Type:        template.Type,
			Description: template.Description,
			Target:      target,
			Quantity:    quantity,
			Progress:    0,
			Complete:    false,
			Optional:    i >= 1 && rng.Float32() < 0.3, // 30% chance for optional objectives after first
			Conditions:  make(map[string]interface{}),
		}

		objectives = append(objectives, objective)
	}

	return objectives, nil
}

// generateRewards creates appropriate rewards for quest completion
func (obg *ObjectiveBasedGenerator) generateRewards(difficulty int, tier pcg.RarityTier, rng *rand.Rand) ([]game.QuestReward, error) {
	rewards := make([]game.QuestReward, 0, 3)

	// Always include experience reward
	expReward := game.QuestReward{
		Type:  "exp",
		Value: difficulty * 100 * (rng.Intn(3) + 1), // 100-300 exp per difficulty level
	}
	rewards = append(rewards, expReward)

	// Add gold reward with 80% probability
	if rng.Float32() < 0.8 {
		goldReward := game.QuestReward{
			Type:  "gold",
			Value: difficulty * 25 * (rng.Intn(4) + 1), // 25-100 gold per difficulty level
		}
		rewards = append(rewards, goldReward)
	}

	// Add item reward based on tier and difficulty
	if difficulty >= 3 && rng.Float32() < 0.6 {
		itemReward := game.QuestReward{
			Type:   "item",
			Value:  1,
			ItemID: fmt.Sprintf("quest_item_%s_%d", tier, rng.Intn(1000)),
		}
		rewards = append(rewards, itemReward)
	}

	return rewards, nil
}

// initializeDefaultTemplates sets up basic objective templates for each quest type
func (obg *ObjectiveBasedGenerator) initializeDefaultTemplates() {
	// Kill quest templates
	obg.objectiveTemplates[pcg.QuestTypeKill] = []*ObjectiveTemplate{
		{
			Type:         "kill",
			Description:  "Defeat the dangerous creatures",
			Requirements: []string{"combat"},
			Targets:      []string{"goblin", "orc", "skeleton", "wolf"},
			Quantities:   [2]int{3, 8},
			Rewards:      []string{"exp", "gold"},
		},
		{
			Type:         "kill_boss",
			Description:  "Slay the powerful enemy leader",
			Requirements: []string{"combat", "tactics"},
			Targets:      []string{"orc_chief", "goblin_king", "dark_wizard"},
			Quantities:   [2]int{1, 1},
			Rewards:      []string{"exp", "gold", "item"},
		},
	}

	// Fetch quest templates
	obg.objectiveTemplates[pcg.QuestTypeFetch] = []*ObjectiveTemplate{
		{
			Type:         "collect",
			Description:  "Gather the required materials",
			Requirements: []string{"exploration"},
			Targets:      []string{"herb", "crystal", "scroll", "key"},
			Quantities:   [2]int{5, 15},
			Rewards:      []string{"exp", "gold"},
		},
		{
			Type:         "retrieve",
			Description:  "Recover the lost artifact",
			Requirements: []string{"exploration", "combat"},
			Targets:      []string{"ancient_relic", "magic_tome", "royal_crown"},
			Quantities:   [2]int{1, 1},
			Rewards:      []string{"exp", "gold", "item"},
		},
	}

	// Explore quest templates
	obg.objectiveTemplates[pcg.QuestTypeExplore] = []*ObjectiveTemplate{
		{
			Type:         "discover",
			Description:  "Explore the uncharted territory",
			Requirements: []string{"movement"},
			Targets:      []string{"cave", "ruins", "forest", "mountain"},
			Quantities:   [2]int{1, 3},
			Rewards:      []string{"exp", "gold"},
		},
		{
			Type:         "map",
			Description:  "Chart the area completely",
			Requirements: []string{"movement", "observation"},
			Targets:      []string{"dungeon_level", "wilderness_area"},
			Quantities:   [2]int{80, 100}, // Percentage
			Rewards:      []string{"exp", "gold"},
		},
	}

	// Delivery quest templates
	obg.objectiveTemplates[pcg.QuestTypeDelivery] = []*ObjectiveTemplate{
		{
			Type:         "deliver",
			Description:  "Transport the package safely",
			Requirements: []string{"movement"},
			Targets:      []string{"merchant", "guard", "scholar", "noble"},
			Quantities:   [2]int{1, 3},
			Rewards:      []string{"exp", "gold"},
		},
	}

	// Escort quest templates
	obg.objectiveTemplates[pcg.QuestTypeEscort] = []*ObjectiveTemplate{
		{
			Type:         "escort",
			Description:  "Guide the traveler to safety",
			Requirements: []string{"movement", "protection"},
			Targets:      []string{"merchant", "diplomat", "pilgrim"},
			Quantities:   [2]int{1, 1},
			Rewards:      []string{"exp", "gold"},
		},
	}
}
