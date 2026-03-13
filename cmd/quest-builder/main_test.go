package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goldbox-rpg/pkg/cliutil"
	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFromTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
		wantID   string
		wantObjs int
	}{
		{
			name:     "fetch template",
			template: "fetch",
			wantErr:  false,
			wantID:   "quest_fetch_001",
			wantObjs: 3,
		},
		{
			name:     "kill template",
			template: "kill",
			wantErr:  false,
			wantID:   "quest_kill_001",
			wantObjs: 2,
		},
		{
			name:     "escort template",
			template: "escort",
			wantErr:  false,
			wantID:   "quest_escort_001",
			wantObjs: 3,
		},
		{
			name:     "explore template",
			template: "explore",
			wantErr:  false,
			wantID:   "quest_explore_001",
			wantObjs: 2,
		},
		{
			name:     "puzzle template",
			template: "puzzle",
			wantErr:  false,
			wantID:   "quest_puzzle_001",
			wantObjs: 2,
		},
		{
			name:     "invalid template",
			template: "invalid",
			wantErr:  true,
		},
		{
			name:     "case insensitive",
			template: "FETCH",
			wantErr:  false,
			wantID:   "quest_fetch_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quest, err := createFromTemplate(tt.template)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, quest.ID)
			if tt.wantObjs > 0 {
				assert.Len(t, quest.Objectives, tt.wantObjs)
			}
		})
	}
}

func TestValidateQuest(t *testing.T) {
	tests := []struct {
		name    string
		quest   game.Quest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid quest",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test objective", Required: 1},
				},
				Rewards: []game.QuestReward{
					{Type: "gold", Value: 100},
				},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			quest: game.Quest{
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test objective", Required: 1},
				},
			},
			wantErr: true,
			errMsg:  "quest ID is required",
		},
		{
			name: "missing title",
			quest: game.Quest{
				ID: "quest_test",
				Objectives: []game.QuestObjective{
					{Description: "Test objective", Required: 1},
				},
			},
			wantErr: true,
			errMsg:  "quest title is required",
		},
		{
			name: "no objectives",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
			},
			wantErr: true,
			errMsg:  "at least one objective is required",
		},
		{
			name: "objective without description",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "", Required: 1},
				},
			},
			wantErr: true,
			errMsg:  "objective 1 has no description",
		},
		{
			name: "objective with zero required",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test", Required: 0},
				},
			},
			wantErr: true,
			errMsg:  "objective 1 has invalid required count",
		},
		{
			name: "invalid reward type",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test", Required: 1},
				},
				Rewards: []game.QuestReward{
					{Type: "invalid", Value: 100},
				},
			},
			wantErr: true,
			errMsg:  "reward 1 has invalid type",
		},
		{
			name: "reward with zero value",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test", Required: 1},
				},
				Rewards: []game.QuestReward{
					{Type: "gold", Value: 0},
				},
			},
			wantErr: true,
			errMsg:  "reward 1 has invalid value",
		},
		{
			name: "item reward without item ID",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test", Required: 1},
				},
				Rewards: []game.QuestReward{
					{Type: "item", Value: 1, ItemID: ""},
				},
			},
			wantErr: true,
			errMsg:  "reward 1 (item) requires an item ID",
		},
		{
			name: "valid item reward",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test", Required: 1},
				},
				Rewards: []game.QuestReward{
					{Type: "item", Value: 1, ItemID: "item_sword"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid exp reward",
			quest: game.Quest{
				ID:    "quest_test",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test", Required: 1},
				},
				Rewards: []game.QuestReward{
					{Type: "exp", Value: 500},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQuest(tt.quest)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOutputQuest(t *testing.T) {
	quest := game.Quest{
		ID:          "quest_test_output",
		Title:       "Test Output Quest",
		Description: "A quest for testing output",
		Objectives: []game.QuestObjective{
			{Description: "Test objective", Required: 1},
		},
		Rewards: []game.QuestReward{
			{Type: "gold", Value: 100},
		},
	}

	// Test writing to file
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test_quest.yaml")

	err := outputQuest(quest, outputFile)
	require.NoError(t, err)

	// Verify file was created and contains expected content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "quest_test_output")
	assert.Contains(t, string(content), "Test Output Quest")
	assert.Contains(t, string(content), "Test objective")
}

func TestRun(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "empty config shows usage",
			cfg:     &Config{},
			wantErr: false,
		},
		{
			name: "template only",
			cfg: &Config{
				Template:   "fetch",
				OutputFile: filepath.Join(tmpDir, "run_quest.yaml"),
			},
			wantErr: false,
		},
		{
			name: "invalid template",
			cfg: &Config{
				Template: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQuestTemplateContent(t *testing.T) {
	// Verify all templates have valid structure
	templates := []string{"fetch", "kill", "escort", "explore", "puzzle"}

	for _, template := range templates {
		t.Run(template, func(t *testing.T) {
			quest, err := createFromTemplate(template)
			require.NoError(t, err)

			// Each template should have valid required fields
			assert.NotEmpty(t, quest.ID, "template %s should have ID", template)
			assert.NotEmpty(t, quest.Title, "template %s should have title", template)
			assert.NotEmpty(t, quest.Description, "template %s should have description", template)
			assert.NotEmpty(t, quest.Objectives, "template %s should have objectives", template)
			assert.NotEmpty(t, quest.Rewards, "template %s should have rewards", template)

			// Should pass validation
			err = validateQuest(quest)
			assert.NoError(t, err, "template %s should be valid", template)
		})
	}
}

func TestQuestObjectiveValidation(t *testing.T) {
	// Test multiple objectives with one invalid
	quest := game.Quest{
		ID:    "quest_multi_obj",
		Title: "Multi Objective Quest",
		Objectives: []game.QuestObjective{
			{Description: "Valid objective 1", Required: 1},
			{Description: "", Required: 1}, // Invalid - no description
			{Description: "Valid objective 3", Required: 1},
		},
	}

	err := validateQuest(quest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "objective 2")
}

func TestQuestRewardValidation(t *testing.T) {
	// Test multiple rewards with one invalid
	quest := game.Quest{
		ID:    "quest_multi_reward",
		Title: "Multi Reward Quest",
		Objectives: []game.QuestObjective{
			{Description: "Test", Required: 1},
		},
		Rewards: []game.QuestReward{
			{Type: "gold", Value: 100},
			{Type: "exp", Value: -5}, // Invalid - negative value
		},
	}

	err := validateQuest(quest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reward 2")
}

func TestOutputQuestToStdout(t *testing.T) {
	quest := game.Quest{
		ID:          "quest_stdout",
		Title:       "Stdout Quest",
		Description: "A quest for testing stdout output",
		Objectives: []game.QuestObjective{
			{Description: "Test objective", Required: 1},
		},
		Rewards: []game.QuestReward{
			{Type: "gold", Value: 50},
		},
	}

	// Test writing to stdout (empty file path) - should not error
	err := outputQuest(quest, "")
	assert.NoError(t, err)
}

func TestRunWithStdoutOutput(t *testing.T) {
	// Test run with template but no output file (writes to stdout)
	cfg := &Config{
		Template:   "explore",
		OutputFile: "", // stdout
	}

	err := run(cfg)
	assert.NoError(t, err)
}

func TestRunAllTemplates(t *testing.T) {
	templates := []string{"fetch", "kill", "escort", "explore", "puzzle"}
	tmpDir := t.TempDir()

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			cfg := &Config{
				Template:   tmpl,
				OutputFile: filepath.Join(tmpDir, tmpl+"_quest.yaml"),
			}

			err := run(cfg)
			assert.NoError(t, err)

			// Verify file was created
			_, err = os.Stat(cfg.OutputFile)
			assert.NoError(t, err)
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Just verify printUsage doesn't panic
	printUsage()
}

func TestPreviewServer(t *testing.T) {
	t.Run("create and broadcast", func(t *testing.T) {
		// Create a preview server on a random port
		ps := cliutil.NewPreviewServer(0, previewHTML, ".")
		assert.NotNil(t, ps)

		// Test broadcasting to empty client list (should not panic)
		quest := game.Quest{
			ID:          "test_quest",
			Title:       "Test Quest",
			Description: "A test quest",
			Objectives: []game.QuestObjective{
				{Description: "Test objective", Required: 1},
			},
		}
		data, _ := json.Marshal(quest)
		ps.Broadcast(data)
	})

	t.Run("add and remove client", func(t *testing.T) {
		ps := cliutil.NewPreviewServer(0, previewHTML, ".")
		assert.NotNil(t, ps)
		// Note: We can't easily test with real websocket connections here
		// but we verify the server struct is properly initialized
		assert.Equal(t, 0, ps.Port())
	})
}

func TestPreviewHTMLEmbedded(t *testing.T) {
	// Verify the preview.html file is properly embedded
	content, err := previewHTML.ReadFile("preview.html")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Quest Builder - Live Preview")
	assert.Contains(t, string(content), "WebSocket")
}

func TestPromptString(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		prompt     string
		defaultVal string
		prefix     string
		want       string
	}{
		{
			name:       "user provides input",
			input:      "My Custom Value\n",
			prompt:     "Test Prompt",
			defaultVal: "default",
			prefix:     "",
			want:       "My Custom Value",
		},
		{
			name:       "empty input returns default",
			input:      "\n",
			prompt:     "Test Prompt",
			defaultVal: "default_value",
			prefix:     "",
			want:       "default_value",
		},
		{
			name:       "empty input with prefix",
			input:      "\n",
			prompt:     "Test Prompt",
			defaultVal: "",
			prefix:     "quest_",
			want:       "quest_new",
		},
		{
			name:       "whitespace trimmed",
			input:      "  trimmed  \n",
			prompt:     "Test Prompt",
			defaultVal: "",
			prefix:     "",
			want:       "trimmed",
		},
		{
			name:       "no default or prefix",
			input:      "value\n",
			prompt:     "Test Prompt",
			defaultVal: "",
			prefix:     "",
			want:       "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			got := promptString(reader, tt.prompt, tt.defaultVal, tt.prefix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPromptInt(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		prompt     string
		defaultVal int
		want       int
	}{
		{
			name:       "valid integer input",
			input:      "42\n",
			prompt:     "Enter number",
			defaultVal: 10,
			want:       42,
		},
		{
			name:       "empty returns default",
			input:      "\n",
			prompt:     "Enter number",
			defaultVal: 99,
			want:       99,
		},
		{
			name:       "invalid returns default",
			input:      "not a number\n",
			prompt:     "Enter number",
			defaultVal: 50,
			want:       50,
		},
		{
			name:       "negative number",
			input:      "-5\n",
			prompt:     "Enter number",
			defaultVal: 10,
			want:       -5,
		},
		{
			name:       "zero",
			input:      "0\n",
			prompt:     "Enter number",
			defaultVal: 100,
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			got := promptInt(reader, tt.prompt, tt.defaultVal)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPromptObjectives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		existing []game.QuestObjective
		wantLen  int
	}{
		{
			name:     "keep existing objectives",
			input:    "n\n",
			existing: []game.QuestObjective{{Description: "Existing", Required: 1}},
			wantLen:  1,
		},
		{
			name:     "add new objectives",
			input:    "y\nFirst objective\n5\nSecond objective\n3\n\n",
			existing: []game.QuestObjective{},
			wantLen:  2,
		},
		{
			name:     "replace existing objectives",
			input:    "y\nNew objective\n2\n\n",
			existing: []game.QuestObjective{{Description: "Old", Required: 1}},
			wantLen:  1,
		},
		{
			name:     "must add at least one if none exist",
			input:    "n\nRequired objective\n1\n\n",
			existing: []game.QuestObjective{},
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			got := promptObjectives(reader, tt.existing)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestPromptRewards(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		existing []game.QuestReward
		wantLen  int
	}{
		{
			name:     "keep existing rewards",
			input:    "n\n",
			existing: []game.QuestReward{{Type: "gold", Value: 100}},
			wantLen:  1,
		},
		{
			name:     "add gold and exp rewards",
			input:    "y\ngold\n100\nexp\n50\n\n",
			existing: []game.QuestReward{},
			wantLen:  2,
		},
		{
			name:     "add item reward",
			input:    "y\nitem\n1\nitem_sword\n\n",
			existing: []game.QuestReward{},
			wantLen:  1,
		},
		{
			name:     "invalid type retries",
			input:    "y\ninvalid\ngold\n50\n\n",
			existing: []game.QuestReward{},
			wantLen:  1,
		},
		{
			name:     "empty type finishes",
			input:    "y\n\n",
			existing: []game.QuestReward{},
			wantLen:  0,
		},
		{
			name:     "no existing with no input",
			input:    "n\n\n",
			existing: []game.QuestReward{},
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			got := promptRewards(reader, tt.existing)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestPromptObjectivesWithPreview(t *testing.T) {
	t.Run("broadcasts update", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader("n\n"))
		quest := &game.Quest{
			ID:    "test",
			Title: "Test",
			Objectives: []game.QuestObjective{
				{Description: "Existing", Required: 1},
			},
		}
		// nil preview should not panic
		got := promptObjectivesWithPreview(reader, quest.Objectives, quest, nil)
		assert.Len(t, got, 1)
	})
}

func TestPromptRewardsWithPreview(t *testing.T) {
	t.Run("broadcasts update", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader("n\n"))
		quest := &game.Quest{
			ID:    "test",
			Title: "Test",
			Rewards: []game.QuestReward{
				{Type: "gold", Value: 100},
			},
		}
		// nil preview should not panic
		got := promptRewardsWithPreview(reader, quest.Rewards, quest, nil)
		assert.Len(t, got, 1)
	})
}

func TestValidateQuestBasicFields(t *testing.T) {
	tests := []struct {
		name    string
		quest   game.Quest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid basic fields",
			quest:   game.Quest{ID: "q1", Title: "Title"},
			wantErr: false,
		},
		{
			name:    "empty ID",
			quest:   game.Quest{ID: "", Title: "Title"},
			wantErr: true,
			errMsg:  "quest ID is required",
		},
		{
			name:    "empty title",
			quest:   game.Quest{ID: "q1", Title: ""},
			wantErr: true,
			errMsg:  "quest title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQuestBasicFields(tt.quest)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateQuestObjectivesList(t *testing.T) {
	tests := []struct {
		name       string
		objectives []game.QuestObjective
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid objectives",
			objectives: []game.QuestObjective{
				{Description: "Obj 1", Required: 1},
				{Description: "Obj 2", Required: 5},
			},
			wantErr: false,
		},
		{
			name:       "empty list",
			objectives: []game.QuestObjective{},
			wantErr:    true,
			errMsg:     "at least one objective is required",
		},
		{
			name: "empty description",
			objectives: []game.QuestObjective{
				{Description: "", Required: 1},
			},
			wantErr: true,
			errMsg:  "objective 1 has no description",
		},
		{
			name: "zero required",
			objectives: []game.QuestObjective{
				{Description: "Test", Required: 0},
			},
			wantErr: true,
			errMsg:  "objective 1 has invalid required count",
		},
		{
			name: "negative required",
			objectives: []game.QuestObjective{
				{Description: "Test", Required: -1},
			},
			wantErr: true,
			errMsg:  "objective 1 has invalid required count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQuestObjectivesList(tt.objectives)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateQuestRewardsList(t *testing.T) {
	tests := []struct {
		name    string
		rewards []game.QuestReward
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid rewards",
			rewards: []game.QuestReward{
				{Type: "gold", Value: 100},
				{Type: "exp", Value: 50},
				{Type: "item", Value: 1, ItemID: "item_sword"},
			},
			wantErr: false,
		},
		{
			name:    "empty list is valid",
			rewards: []game.QuestReward{},
			wantErr: false,
		},
		{
			name: "invalid type",
			rewards: []game.QuestReward{
				{Type: "invalid", Value: 100},
			},
			wantErr: true,
			errMsg:  "reward 1 has invalid type",
		},
		{
			name: "zero value",
			rewards: []game.QuestReward{
				{Type: "gold", Value: 0},
			},
			wantErr: true,
			errMsg:  "reward 1 has invalid value",
		},
		{
			name: "item without ID",
			rewards: []game.QuestReward{
				{Type: "item", Value: 1, ItemID: ""},
			},
			wantErr: true,
			errMsg:  "reward 1 (item) requires an item ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQuestRewardsList(tt.rewards)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRewardDisplayFormat(t *testing.T) {
	// Test that rewards display properly in interactive mode by verifying
	// the data structure used for display
	rewards := []game.QuestReward{
		{Type: "gold", Value: 100},
		{Type: "item", Value: 2, ItemID: "item_potion"},
	}

	// Verify gold reward doesn't have ItemID in display context
	assert.Equal(t, "gold", rewards[0].Type)
	assert.Equal(t, 100, rewards[0].Value)
	assert.Empty(t, rewards[0].ItemID)

	// Verify item reward has ItemID
	assert.Equal(t, "item", rewards[1].Type)
	assert.Equal(t, 2, rewards[1].Value)
	assert.Equal(t, "item_potion", rewards[1].ItemID)
}

func TestObjectiveDisplayFormat(t *testing.T) {
	// Test objective structure for display
	obj := game.QuestObjective{
		Description: "Defeat enemies",
		Progress:    3,
		Required:    10,
		Completed:   false,
	}

	assert.Equal(t, "Defeat enemies", obj.Description)
	assert.Equal(t, 10, obj.Required)
	assert.Equal(t, 3, obj.Progress)
	assert.False(t, obj.Completed)
}
