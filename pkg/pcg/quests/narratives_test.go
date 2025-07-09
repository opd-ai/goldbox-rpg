package quests

import (
	"math/rand"
	"strings"
	"testing"

	"goldbox-rpg/pkg/pcg"
)

func TestNewNarrativeEngine(t *testing.T) {
	engine := NewNarrativeEngine()

	if engine == nil {
		t.Fatal("expected narrative engine to be created")
	}

	if len(engine.storyTemplates) == 0 {
		t.Error("expected story templates to be initialized")
	}

	if len(engine.characterPool) == 0 {
		t.Error("expected character pool to be initialized")
	}
}

func TestNarrativeEngine_GenerateQuestNarrative(t *testing.T) {
	engine := NewNarrativeEngine()
	rng := rand.New(rand.NewSource(12345))

	objectives := []pcg.QuestObjective{
		{
			ID:          "obj_1",
			Type:        "kill",
			Description: "Defeat 5 goblins",
			Target:      "goblin",
			Quantity:    5,
			Progress:    0,
			Complete:    false,
			Optional:    false,
			Conditions:  map[string]interface{}{},
		},
	}

	params := pcg.QuestParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        12345,
			Difficulty:  5,
			PlayerLevel: 3,
		},
		QuestType:     pcg.QuestTypeKill,
		MinObjectives: 1,
		MaxObjectives: 2,
		RewardTier:    pcg.RarityCommon,
		Narrative:     pcg.NarrativeLinear,
	}

	narrative, err := engine.GenerateQuestNarrative(pcg.QuestTypeKill, objectives, params, rng)
	if err != nil {
		t.Fatalf("GenerateQuestNarrative() error = %v", err)
	}

	if narrative.Title == "" {
		t.Error("expected narrative title to be set")
	}

	if narrative.Description == "" {
		t.Error("expected narrative description to be set")
	}

	if narrative.QuestGiver == "" {
		t.Error("expected quest giver to be set")
	}

	if narrative.StartDialogue == "" {
		t.Error("expected start dialogue to be set")
	}

	if narrative.EndDialogue == "" {
		t.Error("expected end dialogue to be set")
	}

	if narrative.Lore == "" {
		t.Error("expected lore to be set")
	}
}

func TestNarrativeEngine_GenerateQuestNarrativeUnsupportedType(t *testing.T) {
	engine := NewNarrativeEngine()
	rng := rand.New(rand.NewSource(12345))

	objectives := []pcg.QuestObjective{}
	params := pcg.QuestParams{
		QuestType: pcg.QuestType("unsupported_type"),
	}

	_, err := engine.GenerateQuestNarrative("unsupported_type", objectives, params, rng)
	if err == nil {
		t.Error("expected error for unsupported quest type")
	}
}

func TestNarrativeEngine_SelectQuestGiver(t *testing.T) {
	engine := NewNarrativeEngine()
	rng := rand.New(rand.NewSource(12345))

	questGiver := engine.selectQuestGiver(rng)

	if questGiver == nil {
		t.Fatal("expected quest giver to be selected")
	}

	if questGiver.Archetype == "" {
		t.Error("expected quest giver archetype to be set")
	}

	if len(questGiver.Personality) == 0 {
		t.Error("expected quest giver to have personality traits")
	}

	if len(questGiver.Motivations) == 0 {
		t.Error("expected quest giver to have motivations")
	}

	if len(questGiver.Speech) == 0 {
		t.Error("expected quest giver to have speech patterns")
	}
}

func TestNarrativeEngine_GenerateTitle(t *testing.T) {
	engine := NewNarrativeEngine()
	rng := rand.New(rand.NewSource(12345))

	objectives := []pcg.QuestObjective{
		{
			Type:        "kill",
			Description: "Defeat enemies",
		},
	}

	tests := []struct {
		name      string
		questType pcg.QuestType
	}{
		{"kill quest", pcg.QuestTypeKill},
		{"fetch quest", pcg.QuestTypeFetch},
		{"explore quest", pcg.QuestTypeExplore},
		{"delivery quest", pcg.QuestTypeDelivery},
		{"escort quest", pcg.QuestTypeEscort},
		{"defend quest", pcg.QuestTypeDefend},
		{"puzzle quest", pcg.QuestTypePuzzle},
		{"unknown quest", pcg.QuestType("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := engine.generateTitle(tt.questType, objectives, rng)

			if title == "" {
				t.Error("expected title to be generated")
			}

			// Title should contain at least two words
			words := len(strings.Fields(title))
			if words < 2 {
				t.Errorf("expected title to have at least 2 words, got %d", words)
			}
		})
	}
}

func TestNarrativeEngine_GenerateDescription(t *testing.T) {
	engine := NewNarrativeEngine()
	rng := rand.New(rand.NewSource(12345))

	template := &StoryTemplate{
		Setup: "A dangerous threat has appeared.",
	}

	objectives := []pcg.QuestObjective{
		{
			Description: "defeat the monsters",
		},
		{
			Description: "collect the reward",
		},
	}

	tests := []struct {
		name       string
		objectives []pcg.QuestObjective
	}{
		{"single objective", objectives[:1]},
		{"multiple objectives", objectives},
		{"no objectives", []pcg.QuestObjective{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := engine.generateDescription(template, tt.objectives, rng)

			if description == "" {
				t.Error("expected description to be generated")
			}

			// Should always contain the setup
			if !strings.Contains(description, template.Setup) {
				t.Error("expected description to contain template setup")
			}

			// Should mention objectives if present
			if len(tt.objectives) > 0 {
				if !strings.Contains(description, "You must") {
					t.Error("expected description to mention objectives")
				}
			}
		})
	}
}

func TestNarrativeEngine_GenerateDialogue(t *testing.T) {
	engine := NewNarrativeEngine()
	rng := rand.New(rand.NewSource(12345))

	template := &StoryTemplate{
		Motivation: "We need your help.",
		Resolution: "Thank you for your service.",
	}

	questGiver := &NPCTemplate{
		Archetype: "Village Elder",
	}

	startDialogue := engine.generateStartDialogue(template, questGiver, rng)
	if startDialogue == "" {
		t.Error("expected start dialogue to be generated")
	}

	endDialogue := engine.generateEndDialogue(template, questGiver, rng)
	if endDialogue == "" {
		t.Error("expected end dialogue to be generated")
	}

	// Start dialogue should contain motivation
	if !strings.Contains(startDialogue, template.Motivation) {
		t.Error("expected start dialogue to contain motivation")
	}

	// End dialogue should contain resolution
	if !strings.Contains(endDialogue, template.Resolution) {
		t.Error("expected end dialogue to contain resolution")
	}
}

func TestNarrativeEngine_TemplateInitialization(t *testing.T) {
	engine := NewNarrativeEngine()

	// Check that all major quest types have templates
	expectedTypes := []pcg.QuestType{
		pcg.QuestTypeKill,
		pcg.QuestTypeFetch,
		pcg.QuestTypeExplore,
	}

	for _, questType := range expectedTypes {
		templates, exists := engine.storyTemplates[questType]
		if !exists {
			t.Errorf("expected templates for quest type %s", questType)
			continue
		}

		if len(templates) == 0 {
			t.Errorf("expected at least one template for quest type %s", questType)
			continue
		}

		// Verify template structure
		for i, template := range templates {
			if template.Theme == "" {
				t.Errorf("template %d for %s: expected theme to be set", i, questType)
			}
			if template.Setup == "" {
				t.Errorf("template %d for %s: expected setup to be set", i, questType)
			}
			if template.Motivation == "" {
				t.Errorf("template %d for %s: expected motivation to be set", i, questType)
			}
			if template.Resolution == "" {
				t.Errorf("template %d for %s: expected resolution to be set", i, questType)
			}
		}
	}
}

// TestDeterministicNarrative verifies that the same seed produces the same narrative
func TestDeterministicNarrative(t *testing.T) {
	engine := NewNarrativeEngine()

	objectives := []pcg.QuestObjective{
		{
			Description: "defeat the enemies",
		},
	}

	params := pcg.QuestParams{
		QuestType: pcg.QuestTypeKill,
	}

	// Generate narrative twice with same seed
	rng1 := rand.New(rand.NewSource(12345))
	narrative1, err1 := engine.GenerateQuestNarrative(pcg.QuestTypeKill, objectives, params, rng1)
	if err1 != nil {
		t.Fatalf("first generation error = %v", err1)
	}

	rng2 := rand.New(rand.NewSource(12345))
	narrative2, err2 := engine.GenerateQuestNarrative(pcg.QuestTypeKill, objectives, params, rng2)
	if err2 != nil {
		t.Fatalf("second generation error = %v", err2)
	}

	// Verify deterministic generation
	if narrative1.Title != narrative2.Title {
		t.Errorf("expected same title, got '%s' and '%s'", narrative1.Title, narrative2.Title)
	}

	if narrative1.QuestGiver != narrative2.QuestGiver {
		t.Errorf("expected same quest giver, got '%s' and '%s'", narrative1.QuestGiver, narrative2.QuestGiver)
	}

	if narrative1.StartDialogue != narrative2.StartDialogue {
		t.Errorf("expected same start dialogue, got '%s' and '%s'", narrative1.StartDialogue, narrative2.StartDialogue)
	}
}
