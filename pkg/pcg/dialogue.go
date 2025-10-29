package pcg

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/mb-14/gomarkov"
	"github.com/sirupsen/logrus"

	"goldbox-rpg/pkg/game"
)

// DialogueGenerator creates procedural dialogue trees for NPCs using template-based
// generation supplemented with Markov chain text enhancement for more natural
// conversational flow.
//
// Design Approach:
// - Template-based system ensures consistent dialogue patterns and game mechanics integration
// - Markov chain enhancement adds natural language variation to avoid repetitive text
// - Context-aware generation considers NPC personality, faction relationships, and quest states
// - Thread-safe implementation supports concurrent dialogue generation
//
// Library Choice Rationale:
// - Uses github.com/mb-14/gomarkov for Markov chain text generation
// - Note: Library has 369 stars but last updated 2 years ago, doesn't meet our 6-month criteria
// - However, it's specifically recommended in PLAN.md and is focused/stable for this use case
// - Simple API with JSON serialization support makes it suitable for our deterministic requirements
type DialogueGenerator struct {
	mu              sync.RWMutex
	rng             *rand.Rand
	markovChains    map[string]*gomarkov.Chain // Per-personality/context chains
	dialogTemplates map[DialogType][]DialogTemplate
	greetings       []string
	farewells       []string
	logger          *logrus.Logger
}

// DialogType represents different categories of dialogue interactions
type DialogType string

const (
	DialogTypeGreeting    DialogType = "greeting"
	DialogTypeFarewell    DialogType = "farewell"
	DialogTypeQuest       DialogType = "quest"
	DialogTypeTrade       DialogType = "trade"
	DialogTypeInformation DialogType = "information"
	DialogTypeRumor       DialogType = "rumor"
	DialogTypeThreat      DialogType = "threat"
	DialogTypeFlirtation  DialogType = "flirtation"
	DialogTypeInsult      DialogType = "insult"
	DialogTypeCompliment  DialogType = "compliment"
	DialogTypeNegotiation DialogType = "negotiation"
	DialogTypeConfession  DialogType = "confession"
)

// DialogTone represents the emotional tone of dialogue
type DialogTone string

const (
	DialogToneFriendly   DialogTone = "friendly"
	DialogToneNeutral    DialogTone = "neutral"
	DialogToneHostile    DialogTone = "hostile"
	DialogToneMysterious DialogTone = "mysterious"
	DialogTonePlayful    DialogTone = "playful"
	DialogToneFormal     DialogTone = "formal"
	DialogToneCasual     DialogTone = "casual"
	DialogToneIntimate   DialogTone = "intimate"
)

// DialogTemplate defines a structure for generating consistent dialogue patterns
type DialogTemplate struct {
	Pattern    string            // Template with placeholders like {name}, {quest}
	Responses  []ResponsePattern // Possible player response templates
	Conditions []string          // Context conditions where this template applies
	Tone       DialogTone        // Emotional tone of the dialogue
	MinWords   int               // Minimum words for Markov enhancement
	MaxWords   int               // Maximum words for Markov enhancement
}

// ResponsePattern defines possible player response structures
type ResponsePattern struct {
	Text       string // Response template with placeholders
	NextDialog string // ID pattern for next dialogue
	Action     string // Game action to trigger
	Conditions []string
}

// DialogParams specifies parameters for dialogue generation
type DialogParams struct {
	NPC              *game.NPC          `yaml:"npc"`              // Target NPC for dialogue
	PlayerCharacter  *game.Character    `yaml:"player"`           // Player character for context
	Context          DialogContext      `yaml:"context"`          // Conversation context
	DialogType       DialogType         `yaml:"dialog_type"`      // Type of dialogue to generate
	Tone             DialogTone         `yaml:"tone"`             // Desired emotional tone
	MaxDepth         int                `yaml:"max_depth"`        // Maximum dialogue tree depth
	ResponseCount    int                `yaml:"response_count"`   // Number of response options
	UseMarkov        bool               `yaml:"use_markov"`       // Enable Markov chain enhancement
	MarkovChainOrder int                `yaml:"markov_order"`     // Markov chain order (1-3)
	Personality      PersonalityProfile `yaml:"personality"`      // NPC personality for context
	FactionStanding  map[string]int     `yaml:"faction_standing"` // Player's standing with factions
}

// DialogContext provides conversation context for generation
type DialogContext struct {
	Location        string                 `yaml:"location"`         // Where conversation takes place
	TimeOfDay       string                 `yaml:"time_of_day"`      // Time context
	ActiveQuests    []string               `yaml:"active_quests"`    // Player's active quests
	CompletedQuests []string               `yaml:"completed_quests"` // Player's completed quests
	Reputation      int                    `yaml:"reputation"`       // Player's general reputation
	LastInteraction time.Time              `yaml:"last_interaction"` // When last spoke to this NPC
	Inventory       []string               `yaml:"inventory"`        // Player's notable items
	Relationships   map[string]interface{} `yaml:"relationships"`    // NPC relationship states
}

// GeneratedDialogue represents a complete dialogue tree
type GeneratedDialogue struct {
	RootEntry   *game.DialogEntry   `yaml:"root_entry"`   // Starting dialogue node
	AllEntries  []*game.DialogEntry `yaml:"all_entries"`  // All dialogue nodes in tree
	NPCPersona  PersonalityProfile  `yaml:"npc_persona"`  // Generated NPC personality
	GeneratedAt time.Time           `yaml:"generated_at"` // When dialogue was created
	ContextUsed DialogContext       `yaml:"context_used"` // Context that influenced generation
	MarkovUsed  bool                `yaml:"markov_used"`  // Whether Markov enhancement was applied
	TotalNodes  int                 `yaml:"total_nodes"`  // Number of dialogue nodes created
	MaxDepth    int                 `yaml:"max_depth"`    // Actual maximum depth achieved
}

// NewDialogueGenerator creates a new dialogue generator with predefined templates and Markov chains
func NewDialogueGenerator(logger *logrus.Logger) *DialogueGenerator {
	if logger == nil {
		logger = logrus.New()
	}

	generator := &DialogueGenerator{
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
		markovChains:    make(map[string]*gomarkov.Chain),
		dialogTemplates: make(map[DialogType][]DialogTemplate),
		logger:          logger,
	}

	generator.initializeTemplates()
	generator.initializeMarkovChains()

	return generator
}

// Generate creates a dialogue tree based on the provided parameters
func (dg *DialogueGenerator) Generate(ctx context.Context, params GenerationParams) (interface{}, error) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	// Set up RNG with seed for deterministic generation
	if params.Seed != 0 {
		dg.rng = rand.New(rand.NewSource(params.Seed))
	}

	// Extract dialogue-specific parameters
	dialogParams, ok := params.Constraints["dialogue_params"].(DialogParams)
	if !ok {
		// Create default parameters
		dialogParams = DialogParams{
			DialogType:       DialogTypeGreeting,
			Tone:             DialogToneFriendly,
			MaxDepth:         3,
			ResponseCount:    3,
			UseMarkov:        true,
			MarkovChainOrder: 2,
		}
	}

	startTime := time.Now()
	dg.logger.WithFields(logrus.Fields{
		"npc_id":      params.Constraints["npc_id"],
		"dialog_type": dialogParams.DialogType,
		"tone":        dialogParams.Tone,
		"max_depth":   dialogParams.MaxDepth,
		"use_markov":  dialogParams.UseMarkov,
		"seed":        params.Seed,
	}).Info("generating dialogue tree")

	// Generate the dialogue tree
	dialogue, err := dg.generateDialogueTree(ctx, dialogParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dialogue tree: %w", err)
	}

	duration := time.Since(startTime)
	dg.logger.WithFields(logrus.Fields{
		"npc_id":      params.Constraints["npc_id"],
		"nodes":       dialogue.TotalNodes,
		"max_depth":   dialogue.MaxDepth,
		"markov_used": dialogue.MarkovUsed,
		"duration":    duration,
	}).Info("dialogue generation completed")

	return dialogue, nil
}

// GetType returns the content type this generator produces
func (dg *DialogueGenerator) GetType() ContentType {
	return ContentTypeDialogue
}

// GetVersion returns the generator version for compatibility checking
func (dg *DialogueGenerator) GetVersion() string {
	return "1.0.0"
}

// Validate checks if the provided parameters are valid for dialogue generation
func (dg *DialogueGenerator) Validate(params GenerationParams) error {
	if params.Seed < 0 {
		return fmt.Errorf("seed must be non-negative, got %d", params.Seed)
	}

	// Check for dialogue-specific parameters
	if dialogParams, ok := params.Constraints["dialogue_params"].(DialogParams); ok {
		if dialogParams.MaxDepth < 1 || dialogParams.MaxDepth > 10 {
			return fmt.Errorf("max_depth must be between 1 and 10, got %d", dialogParams.MaxDepth)
		}

		if dialogParams.ResponseCount < 1 || dialogParams.ResponseCount > 8 {
			return fmt.Errorf("response_count must be between 1 and 8, got %d", dialogParams.ResponseCount)
		}

		if dialogParams.MarkovChainOrder < 1 || dialogParams.MarkovChainOrder > 4 {
			return fmt.Errorf("markov_chain_order must be between 1 and 4, got %d", dialogParams.MarkovChainOrder)
		}
	}

	return nil
}

// generateDialogueTree creates a complete dialogue tree with branching conversations
func (dg *DialogueGenerator) generateDialogueTree(ctx context.Context, params DialogParams) (*GeneratedDialogue, error) {
	// Create root dialogue entry
	rootEntry, err := dg.generateDialogueEntry(ctx, params, 0, "root")
	if err != nil {
		return nil, fmt.Errorf("failed to generate root dialogue: %w", err)
	}

	// Generate all dialogue nodes
	allEntries := []*game.DialogEntry{rootEntry}
	maxDepth := 0

	// BFS to generate dialogue tree
	queue := []struct {
		entry *game.DialogEntry
		depth int
	}{{rootEntry, 0}}

	for len(queue) > 0 && len(allEntries) < 50 { // Prevent excessive generation
		current := queue[0]
		queue = queue[1:]

		if current.depth > maxDepth {
			maxDepth = current.depth
		}

		// Generate responses and their follow-up dialogues
		for i, response := range current.entry.Responses {
			if response.NextDialog != "" && current.depth < params.MaxDepth {
				nextParams := params
				nextParams.DialogType = dg.inferDialogTypeFromResponse(response.Text)

				nextEntry, err := dg.generateDialogueEntry(ctx, nextParams, current.depth+1,
					fmt.Sprintf("%s_r%d", current.entry.ID, i))
				if err != nil {
					dg.logger.WithError(err).Warn("failed to generate follow-up dialogue")
					continue
				}

				allEntries = append(allEntries, nextEntry)
				queue = append(queue, struct {
					entry *game.DialogEntry
					depth int
				}{nextEntry, current.depth + 1})

				// Update response to point to generated dialogue
				current.entry.Responses[i].NextDialog = nextEntry.ID
			}
		}
	}

	return &GeneratedDialogue{
		RootEntry:   rootEntry,
		AllEntries:  allEntries,
		NPCPersona:  params.Personality,
		GeneratedAt: time.Now(),
		ContextUsed: params.Context,
		MarkovUsed:  params.UseMarkov,
		TotalNodes:  len(allEntries),
		MaxDepth:    maxDepth,
	}, nil
}

// generateDialogueEntry creates a single dialogue entry with responses
func (dg *DialogueGenerator) generateDialogueEntry(ctx context.Context, params DialogParams, depth int, idSuffix string) (*game.DialogEntry, error) {
	// Select appropriate template
	templates := dg.dialogTemplates[params.DialogType]
	if len(templates) == 0 {
		templates = dg.dialogTemplates[DialogTypeGreeting] // fallback
	}

	template := templates[dg.rng.Intn(len(templates))]

	// Generate dialogue text
	dialogText, err := dg.generateDialogueText(template, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dialogue text: %w", err)
	}

	// Generate responses
	responses, err := dg.generateResponses(template, params, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to generate responses: %w", err)
	}

	entry := &game.DialogEntry{
		ID:         fmt.Sprintf("dialog_%s_%s", strings.ReplaceAll(string(params.DialogType), " ", "_"), idSuffix),
		Text:       dialogText,
		Responses:  responses,
		Conditions: []game.DialogCondition{}, // Would be populated based on context
	}

	return entry, nil
}

// generateDialogueText creates dialogue text using templates and optional Markov enhancement
func (dg *DialogueGenerator) generateDialogueText(template DialogTemplate, params DialogParams) (string, error) {
	// Start with template-based text
	baseText := dg.fillTemplate(template.Pattern, params)

	// Apply Markov enhancement if enabled
	if params.UseMarkov && len(strings.Fields(baseText)) >= template.MinWords {
		enhanced, err := dg.enhanceWithMarkov(baseText, params)
		if err != nil {
			dg.logger.WithError(err).Debug("Markov enhancement failed, using template text")
			return baseText, nil
		}
		return enhanced, nil
	}

	return baseText, nil
}

// generateResponses creates player response options for a dialogue
func (dg *DialogueGenerator) generateResponses(template DialogTemplate, params DialogParams, depth int) ([]game.DialogResponse, error) {
	var responses []game.DialogResponse

	// Limit response count based on depth to prevent explosion
	maxResponses := params.ResponseCount
	if depth > 2 {
		maxResponses = min(maxResponses, 2)
	}

	responseCount := min(len(template.Responses), maxResponses)
	if responseCount == 0 {
		responseCount = 1 // Always have at least one response
	}

	for i := 0; i < responseCount; i++ {
		var responsePattern ResponsePattern

		if i < len(template.Responses) {
			responsePattern = template.Responses[i]
		} else {
			// Generate generic response
			responsePattern = ResponsePattern{
				Text:       "I understand.",
				NextDialog: "",
				Action:     "",
			}
		}

		responseText := dg.fillTemplate(responsePattern.Text, params)

		response := game.DialogResponse{
			Text:       responseText,
			NextDialog: responsePattern.NextDialog,
			Action:     responsePattern.Action,
		}

		responses = append(responses, response)
	}

	// Always add a farewell option at deeper levels
	if depth > 0 {
		responses = append(responses, game.DialogResponse{
			Text:       dg.generateFarewell(params),
			NextDialog: "",
			Action:     "end_conversation",
		})
	}

	return responses, nil
}

// fillTemplate replaces placeholders in templates with contextual information
func (dg *DialogueGenerator) fillTemplate(template string, params DialogParams) string {
	text := template

	// Replace common placeholders
	if params.NPC != nil {
		text = strings.ReplaceAll(text, "{npc_name}", params.NPC.Name)
	}
	if params.PlayerCharacter != nil {
		text = strings.ReplaceAll(text, "{player_name}", params.PlayerCharacter.Name)
	}

	text = strings.ReplaceAll(text, "{location}", params.Context.Location)
	text = strings.ReplaceAll(text, "{time}", params.Context.TimeOfDay)

	// Add personality-based modifications
	text = dg.applyPersonalityToText(text, params.Personality)

	return text
}

// applyPersonalityToText modifies text based on NPC personality traits
func (dg *DialogueGenerator) applyPersonalityToText(text string, personality PersonalityProfile) string {
	// Apply speech pattern formality
	if personality.Speech.Formality == "formal" {
		text = dg.makeTextFormal(text)
	} else if personality.Speech.Formality == "crude" {
		text = dg.makeTextCasual(text)
	}

	// Add catchphrase if present
	if personality.Speech.Catchphrase != "" && dg.rng.Float64() < 0.3 {
		text = text + " " + personality.Speech.Catchphrase
	}

	// Apply trait-based modifications
	for _, trait := range personality.Traits {
		if trait.Intensity > 0.7 {
			text = dg.applyTraitToText(text, trait.Name)
		}
	}

	return text
}

// makeTextFormal converts casual text to more formal speech
func (dg *DialogueGenerator) makeTextFormal(text string) string {
	// Simple formality transformations
	text = strings.ReplaceAll(text, "you're", "you are")
	text = strings.ReplaceAll(text, "I'm", "I am")
	text = strings.ReplaceAll(text, "can't", "cannot")
	text = strings.ReplaceAll(text, "won't", "will not")
	text = strings.ReplaceAll(text, "don't", "do not")
	text = strings.ReplaceAll(text, "isn't", "is not")
	return text
}

// makeTextCasual converts formal text to more casual speech
func (dg *DialogueGenerator) makeTextCasual(text string) string {
	// Simple casualness transformations
	text = strings.ReplaceAll(text, "you are", "you're")
	text = strings.ReplaceAll(text, "I am", "I'm")
	text = strings.ReplaceAll(text, "cannot", "can't")
	text = strings.ReplaceAll(text, "do not", "don't")
	text = strings.ReplaceAll(text, "is not", "isn't")
	return text
}

// applyTraitToText modifies text based on a specific personality trait
func (dg *DialogueGenerator) applyTraitToText(text, traitName string) string {
	lowerTrait := strings.ToLower(traitName)

	if strings.Contains(lowerTrait, "arrogant") {
		text = "Obviously, " + strings.ToLower(text[:1]) + text[1:]
	} else if strings.Contains(lowerTrait, "nervous") {
		text = text + "... if you don't mind me saying."
	} else if strings.Contains(lowerTrait, "cheerful") {
		text = text + "!"
	}

	return text
}

// enhanceWithMarkov uses Markov chains to add variation to template-based text
func (dg *DialogueGenerator) enhanceWithMarkov(baseText string, params DialogParams) (string, error) {
	personalityKey := dg.getPersonalityKey(params.Personality)

	chain, exists := dg.markovChains[personalityKey]
	if !exists {
		return baseText, fmt.Errorf("no Markov chain available for personality: %s", personalityKey)
	}

	words := strings.Fields(baseText)
	if len(words) < 2 {
		return baseText, nil
	}
	// Use first words as seed and generate continuation
	seedWords := words[:min(len(words)/2, params.MarkovChainOrder)]

	generated, err := chain.Generate(seedWords)
	if err != nil {
		return baseText, fmt.Errorf("Markov generation failed: %w", err)
	}

	// Combine original concept with Markov-generated continuation
	enhanced := strings.Join(words[:len(seedWords)], " ") + " " + generated

	// Ensure reasonable length
	enhancedWords := strings.Fields(enhanced)
	if len(enhancedWords) > 30 {
		enhanced = strings.Join(enhancedWords[:30], " ")
	}

	return enhanced, nil
}

// Helper methods

func (dg *DialogueGenerator) initializeTemplates() {
	// Initialize dialogue templates for different types
	dg.dialogTemplates[DialogTypeGreeting] = []DialogTemplate{
		{
			Pattern: "Good {time}, {player_name}. What brings you to {location}?",
			Responses: []ResponsePattern{
				{Text: "I'm just passing through.", NextDialog: "casual_chat", Action: ""},
				{Text: "I'm looking for work.", NextDialog: "quest_offer", Action: ""},
				{Text: "What's the latest news?", NextDialog: "information", Action: ""},
			},
			Tone:     DialogToneFriendly,
			MinWords: 5,
			MaxWords: 15,
		},
		{
			Pattern: "Well, well... {player_name}. I wasn't expecting to see you here.",
			Responses: []ResponsePattern{
				{Text: "Surprised to see me?", NextDialog: "mysterious_chat", Action: ""},
				{Text: "I go where I please.", NextDialog: "defiant_response", Action: ""},
			},
			Tone:     DialogToneMysterious,
			MinWords: 6,
			MaxWords: 12,
		},
	}

	dg.dialogTemplates[DialogTypeQuest] = []DialogTemplate{
		{
			Pattern: "I have a task that might interest someone of your... capabilities, {player_name}.",
			Responses: []ResponsePattern{
				{Text: "Tell me more.", NextDialog: "quest_details", Action: ""},
				{Text: "What's the reward?", NextDialog: "quest_reward", Action: ""},
				{Text: "I'm not interested.", NextDialog: "", Action: "decline_quest"},
			},
			Tone:     DialogToneFormal,
			MinWords: 8,
			MaxWords: 20,
		},
	}

	// Add more template types...
	dg.initializeAdditionalTemplates()
}

func (dg *DialogueGenerator) initializeAdditionalTemplates() {
	// Trade dialogue
	dg.dialogTemplates[DialogTypeTrade] = []DialogTemplate{
		{
			Pattern: "Welcome to my shop, {player_name}. What can I get for you today?",
			Responses: []ResponsePattern{
				{Text: "Show me your wares.", NextDialog: "", Action: "open_shop"},
				{Text: "I need supplies.", NextDialog: "trade_supplies", Action: ""},
				{Text: "Maybe later.", NextDialog: "", Action: "end_conversation"},
			},
			Tone:     DialogToneCasual,
			MinWords: 6,
			MaxWords: 15,
		},
	}

	// Information dialogue
	dg.dialogTemplates[DialogTypeInformation] = []DialogTemplate{
		{
			Pattern: "Information? Well, I hear many things in my line of work...",
			Responses: []ResponsePattern{
				{Text: "Tell me about recent events.", NextDialog: "recent_news", Action: ""},
				{Text: "Any rumors worth knowing?", NextDialog: "rumors", Action: ""},
				{Text: "What about this area?", NextDialog: "local_info", Action: ""},
			},
			Tone:     DialogToneNeutral,
			MinWords: 7,
			MaxWords: 18,
		},
	}
}

func (dg *DialogueGenerator) initializeMarkovChains() {
	// Create Markov chains for different personality types
	// In a real implementation, these would be trained on personality-specific text corpora

	personalities := []string{"friendly", "hostile", "mysterious", "formal", "casual"}

	for _, personality := range personalities {
		chain := gomarkov.NewChain(2) // Order 2 Markov chain

		// Add training data based on personality
		// This is simplified - real implementation would load from files or databases
		dg.trainMarkovChain(chain, personality)

		dg.markovChains[personality] = chain
	}
}

func (dg *DialogueGenerator) trainMarkovChain(chain *gomarkov.Chain, personality string) {
	// Simplified training data - real implementation would use much larger corpora
	trainingData := dg.getTrainingDataForPersonality(personality)

	for _, sentence := range trainingData {
		words := strings.Fields(sentence)
		if len(words) > 2 {
			chain.Add(words)
		}
	}
}

func (dg *DialogueGenerator) getTrainingDataForPersonality(personality string) []string {
	// Simplified training data sets for different personalities
	switch personality {
	case "friendly":
		return []string{
			"I hope you're having a wonderful day",
			"It's always nice to meet new people",
			"I'd be happy to help you with that",
			"Please let me know if you need anything",
			"Thank you so much for your kindness",
		}
	case "hostile":
		return []string{
			"I don't have time for this nonsense",
			"You'd better watch your step around here",
			"I suggest you move along quickly",
			"Don't test my patience any further",
			"You're not welcome in these parts",
		}
	case "mysterious":
		return []string{
			"Some secrets are better left buried",
			"The shadows hold many interesting tales",
			"Not everything is as it appears to be",
			"There are forces at work beyond your understanding",
			"Time will reveal what must be known",
		}
	default:
		return []string{
			"I understand your concern about this matter",
			"Let me provide you with the information you need",
			"This situation requires careful consideration",
			"We must follow proper procedures in this case",
			"I believe we can reach a satisfactory agreement",
		}
	}
}

func (dg *DialogueGenerator) getPersonalityKey(personality PersonalityProfile) string {
	// Analyze personality traits to determine appropriate Markov chain
	friendliness := dg.getTraitIntensity(personality, "friendly")
	aggression := dg.getTraitIntensity(personality, "aggressive")
	formality := dg.getTraitIntensity(personality, "formal")

	if friendliness > 0.7 {
		return "friendly"
	} else if aggression > 0.7 {
		return "hostile"
	} else if formality > 0.7 {
		return "formal"
	} else if dg.hasTraitType(personality, "mysterious") {
		return "mysterious"
	}
	return "casual"
}

// getTraitIntensity finds the intensity of a specific trait type in the personality
func (dg *DialogueGenerator) getTraitIntensity(personality PersonalityProfile, traitType string) float64 {
	for _, trait := range personality.Traits {
		if strings.Contains(strings.ToLower(trait.Name), traitType) {
			return trait.Intensity
		}
	}
	return 0.0
}

// hasTraitType checks if personality has a trait containing the specified type
func (dg *DialogueGenerator) hasTraitType(personality PersonalityProfile, traitType string) bool {
	for _, trait := range personality.Traits {
		if strings.Contains(strings.ToLower(trait.Name), traitType) {
			return true
		}
	}
	return false
}

func (dg *DialogueGenerator) inferDialogTypeFromResponse(responseText string) DialogType {
	// Simple heuristics to determine next dialogue type based on response
	lowerText := strings.ToLower(responseText)

	if strings.Contains(lowerText, "quest") || strings.Contains(lowerText, "task") {
		return DialogTypeQuest
	} else if strings.Contains(lowerText, "trade") || strings.Contains(lowerText, "buy") {
		return DialogTypeTrade
	} else if strings.Contains(lowerText, "news") || strings.Contains(lowerText, "rumor") {
		return DialogTypeInformation
	}

	return DialogTypeInformation // Default fallback
}

func (dg *DialogueGenerator) generateFarewell(params DialogParams) string {
	farewells := []string{
		"Until next time, {player_name}.",
		"Safe travels, {player_name}.",
		"Farewell for now.",
		"May fortune smile upon you.",
		"I'll be here if you need me.",
	}

	farewell := farewells[dg.rng.Intn(len(farewells))]
	return dg.fillTemplate(farewell, params)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
