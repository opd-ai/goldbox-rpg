package quests

import (
	"fmt"
	"math/rand"

	"goldbox-rpg/pkg/pcg"
)

// NarrativeEngine generates quest stories and dialogue
type NarrativeEngine struct {
	storyTemplates map[pcg.QuestType][]*StoryTemplate
	characterPool  []*NPCTemplate
}

// StoryTemplate defines narrative structure
type StoryTemplate struct {
	Theme      string   `yaml:"theme"`
	Setup      string   `yaml:"setup"`
	Motivation string   `yaml:"motivation"`
	Climax     string   `yaml:"climax"`
	Resolution string   `yaml:"resolution"`
	Characters []string `yaml:"characters"`
	Locations  []string `yaml:"locations"`
}

// NPCTemplate defines quest-giver characteristics
type NPCTemplate struct {
	Archetype   string   `yaml:"archetype"`
	Personality []string `yaml:"personality"`
	Motivations []string `yaml:"motivations"`
	Speech      []string `yaml:"speech_patterns"`
}

// QuestNarrative holds the complete story context
type QuestNarrative struct {
	Title         string `yaml:"title"`
	Description   string `yaml:"description"`
	QuestGiver    string `yaml:"quest_giver"`
	StartDialogue string `yaml:"start_dialogue"`
	EndDialogue   string `yaml:"end_dialogue"`
	Lore          string `yaml:"lore"`
}

// NewNarrativeEngine creates a new narrative engine
func NewNarrativeEngine() *NarrativeEngine {
	ne := &NarrativeEngine{
		storyTemplates: make(map[pcg.QuestType][]*StoryTemplate),
		characterPool:  make([]*NPCTemplate, 0),
	}

	// Initialize default templates
	ne.initializeDefaultTemplates()

	return ne
}

// GenerateQuestNarrative creates story context for a quest
func (ne *NarrativeEngine) GenerateQuestNarrative(questType pcg.QuestType, objectives []pcg.QuestObjective, params pcg.QuestParams, rng *rand.Rand) (*QuestNarrative, error) {
	templates, exists := ne.storyTemplates[questType]
	if !exists || len(templates) == 0 {
		return nil, fmt.Errorf("no story templates available for quest type: %s", questType)
	}

	// Select random template
	template := templates[rng.Intn(len(templates))]

	// Select quest giver
	questGiver := ne.selectQuestGiver(rng)

	// Generate title based on quest type and objectives
	title := ne.generateTitle(questType, objectives, rng)

	// Create description combining setup and objectives
	description := ne.generateDescription(template, objectives, rng)

	// Generate dialogue
	startDialogue := ne.generateStartDialogue(template, questGiver, rng)
	endDialogue := ne.generateEndDialogue(template, questGiver, rng)

	narrative := &QuestNarrative{
		Title:         title,
		Description:   description,
		QuestGiver:    questGiver.Archetype,
		StartDialogue: startDialogue,
		EndDialogue:   endDialogue,
		Lore:          template.Setup,
	}

	return narrative, nil
}

// selectQuestGiver chooses an appropriate NPC for the quest
func (ne *NarrativeEngine) selectQuestGiver(rng *rand.Rand) *NPCTemplate {
	if len(ne.characterPool) == 0 {
		// Return default quest giver if no templates available
		return &NPCTemplate{
			Archetype:   "Village Elder",
			Personality: []string{"wise", "concerned"},
			Motivations: []string{"protect village", "maintain peace"},
			Speech:      []string{"formal", "respectful"},
		}
	}

	return ne.characterPool[rng.Intn(len(ne.characterPool))]
}

// generateTitle creates an appropriate quest title
func (ne *NarrativeEngine) generateTitle(questType pcg.QuestType, objectives []pcg.QuestObjective, rng *rand.Rand) string {
	titlePrefixes := map[pcg.QuestType][]string{
		pcg.QuestTypeKill:     {"Eliminate", "Destroy", "Hunt", "Slay"},
		pcg.QuestTypeFetch:    {"Retrieve", "Collect", "Gather", "Find"},
		pcg.QuestTypeExplore:  {"Explore", "Discover", "Chart", "Scout"},
		pcg.QuestTypeDelivery: {"Deliver", "Transport", "Carry", "Bring"},
		pcg.QuestTypeEscort:   {"Escort", "Guard", "Protect", "Guide"},
		pcg.QuestTypeDefend:   {"Defend", "Protect", "Guard", "Hold"},
		pcg.QuestTypePuzzle:   {"Solve", "Unravel", "Decode", "Unlock"},
	}

	prefixes, exists := titlePrefixes[questType]
	if !exists || len(prefixes) == 0 {
		prefixes = []string{"Complete", "Accomplish"}
	}

	prefix := prefixes[rng.Intn(len(prefixes))]

	suffixes := []string{
		"the Threat", "the Challenge", "the Mission", "the Task",
		"the Problem", "the Request", "the Duty", "the Assignment",
	}

	suffix := suffixes[rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

// generateDescription creates a quest description
func (ne *NarrativeEngine) generateDescription(template *StoryTemplate, objectives []pcg.QuestObjective, rng *rand.Rand) string {
	baseDesc := template.Setup

	// Add objective-specific details
	if len(objectives) > 0 {
		mainObjective := objectives[0]
		baseDesc += fmt.Sprintf(" You must %s", mainObjective.Description)

		if len(objectives) > 1 {
			baseDesc += fmt.Sprintf(" and complete %d additional tasks", len(objectives)-1)
		}
		baseDesc += "."
	}

	return baseDesc
}

// generateStartDialogue creates quest start dialogue
func (ne *NarrativeEngine) generateStartDialogue(template *StoryTemplate, questGiver *NPCTemplate, rng *rand.Rand) string {
	greetings := []string{
		"Greetings, adventurer!",
		"Ah, you look capable.",
		"Thank goodness you're here.",
		"I've been hoping someone like you would come along.",
	}

	greeting := greetings[rng.Intn(len(greetings))]

	return fmt.Sprintf("%s %s %s", greeting, template.Motivation, "Will you help us?")
}

// generateEndDialogue creates quest completion dialogue
func (ne *NarrativeEngine) generateEndDialogue(template *StoryTemplate, questGiver *NPCTemplate, rng *rand.Rand) string {
	thanks := []string{
		"Excellent work!",
		"You've done it!",
		"Marvelous!",
		"I knew you could do it!",
	}

	thank := thanks[rng.Intn(len(thanks))]

	return fmt.Sprintf("%s %s Please accept this reward as thanks for your service.", thank, template.Resolution)
}

// initializeDefaultTemplates sets up basic story templates
func (ne *NarrativeEngine) initializeDefaultTemplates() {
	// Kill quest templates
	ne.storyTemplates[pcg.QuestTypeKill] = []*StoryTemplate{
		{
			Theme:      "Monster Threat",
			Setup:      "Dangerous creatures have been terrorizing the local area.",
			Motivation: "We need someone brave enough to deal with these monsters.",
			Climax:     "The final beast falls before your might.",
			Resolution: "Peace has been restored to the region.",
			Characters: []string{"worried_villager", "town_guard"},
			Locations:  []string{"village", "forest", "cave"},
		},
		{
			Theme:      "Bandit Problem",
			Setup:      "A group of bandits has been raiding trade routes.",
			Motivation: "Our merchants can't travel safely anymore.",
			Climax:     "The bandit leader is defeated.",
			Resolution: "Trade can resume safely once more.",
			Characters: []string{"merchant", "caravan_master"},
			Locations:  []string{"road", "hideout", "camp"},
		},
	}

	// Fetch quest templates
	ne.storyTemplates[pcg.QuestTypeFetch] = []*StoryTemplate{
		{
			Theme:      "Lost Artifact",
			Setup:      "An important artifact has gone missing from our collection.",
			Motivation: "We suspect it may have been taken to the old ruins.",
			Climax:     "You discover the artifact in the depths of the ruins.",
			Resolution: "The artifact has been safely returned.",
			Characters: []string{"scholar", "curator", "priest"},
			Locations:  []string{"library", "ruins", "temple"},
		},
		{
			Theme:      "Magical Components",
			Setup:      "We require rare materials for an important ritual.",
			Motivation: "These components can only be found in dangerous places.",
			Climax:     "You gather the last of the required materials.",
			Resolution: "The ritual can now proceed as planned.",
			Characters: []string{"wizard", "alchemist", "druid"},
			Locations:  []string{"tower", "swamp", "mountain"},
		},
	}

	// Explore quest templates
	ne.storyTemplates[pcg.QuestTypeExplore] = []*StoryTemplate{
		{
			Theme:      "Uncharted Territory",
			Setup:      "Strange reports have come from the unexplored regions.",
			Motivation: "We need scouts to investigate these claims.",
			Climax:     "You uncover the truth behind the mysterious reports.",
			Resolution: "Your findings will help us prepare for what lies ahead.",
			Characters: []string{"explorer", "cartographer", "scout"},
			Locations:  []string{"wilderness", "unknown_region", "frontier"},
		},
	}

	// Initialize character pool
	ne.characterPool = []*NPCTemplate{
		{
			Archetype:   "Village Elder",
			Personality: []string{"wise", "patient", "caring"},
			Motivations: []string{"protect village", "preserve traditions"},
			Speech:      []string{"formal", "measured"},
		},
		{
			Archetype:   "Town Guard Captain",
			Personality: []string{"duty-bound", "serious", "protective"},
			Motivations: []string{"maintain order", "protect citizens"},
			Speech:      []string{"direct", "military"},
		},
		{
			Archetype:   "Merchant",
			Personality: []string{"practical", "worried", "grateful"},
			Motivations: []string{"profit", "safe trade"},
			Speech:      []string{"business-like", "persuasive"},
		},
		{
			Archetype:   "Scholar",
			Personality: []string{"curious", "thoughtful", "academic"},
			Motivations: []string{"knowledge", "research"},
			Speech:      []string{"verbose", "precise"},
		},
	}
}
