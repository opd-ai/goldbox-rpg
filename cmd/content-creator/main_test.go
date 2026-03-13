package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpellFromTemplate(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		wantErr    bool
		wantName   string
		wantLevel  int
		wantSchool game.SpellSchool
	}{
		{
			name:       "damage template",
			template:   "damage",
			wantErr:    false,
			wantName:   "Firebolt",
			wantLevel:  1,
			wantSchool: game.SchoolEvocation,
		},
		{
			name:       "healing template",
			template:   "healing",
			wantErr:    false,
			wantName:   "Cure Wounds",
			wantLevel:  1,
			wantSchool: game.SchoolEvocation,
		},
		{
			name:       "buff template",
			template:   "buff",
			wantErr:    false,
			wantName:   "Bless",
			wantLevel:  1,
			wantSchool: game.SchoolEnchantment,
		},
		{
			name:       "debuff template",
			template:   "debuff",
			wantErr:    false,
			wantName:   "Bane",
			wantLevel:  1,
			wantSchool: game.SchoolEnchantment,
		},
		{
			name:       "utility template",
			template:   "utility",
			wantErr:    false,
			wantName:   "Detect Magic",
			wantLevel:  1,
			wantSchool: game.SchoolDivination,
		},
		{
			name:     "invalid template",
			template: "invalid",
			wantErr:  true,
		},
		{
			name:       "case insensitive",
			template:   "DAMAGE",
			wantErr:    false,
			wantName:   "Firebolt",
			wantLevel:  1,
			wantSchool: game.SchoolEvocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell, err := spellFromTemplate(tt.template)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, spell.Name)
			assert.Equal(t, tt.wantLevel, spell.Level)
			assert.Equal(t, tt.wantSchool, spell.School)
		})
	}
}

func TestItemFromTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
		wantName string
		wantType string
	}{
		{
			name:     "weapon template",
			template: "weapon",
			wantErr:  false,
			wantName: "Longsword",
			wantType: "weapon",
		},
		{
			name:     "armor template",
			template: "armor",
			wantErr:  false,
			wantName: "Chain Mail",
			wantType: "armor",
		},
		{
			name:     "potion template",
			template: "potion",
			wantErr:  false,
			wantName: "Healing Potion",
			wantType: "consumable",
		},
		{
			name:     "accessory template",
			template: "accessory",
			wantErr:  false,
			wantName: "Ring of Protection",
			wantType: "accessory",
		},
		{
			name:     "invalid template",
			template: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := itemFromTemplate(tt.template)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, item.Name)
			assert.Equal(t, tt.wantType, item.Type)
		})
	}
}

func TestValidateSpell(t *testing.T) {
	tests := []struct {
		name    string
		spell   game.Spell
		wantErr bool
	}{
		{
			name: "valid spell",
			spell: game.Spell{
				ID:    "spell_test",
				Name:  "Test Spell",
				Level: 1,
				Range: 30,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			spell: game.Spell{
				Name:  "Test Spell",
				Level: 1,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			spell: game.Spell{
				ID:    "spell_test",
				Level: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid level too low",
			spell: game.Spell{
				ID:    "spell_test",
				Name:  "Test Spell",
				Level: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid level too high",
			spell: game.Spell{
				ID:    "spell_test",
				Name:  "Test Spell",
				Level: 10,
			},
			wantErr: true,
		},
		{
			name: "negative range",
			spell: game.Spell{
				ID:    "spell_test",
				Name:  "Test Spell",
				Level: 1,
				Range: -5,
			},
			wantErr: true,
		},
		{
			name: "cantrip level 0",
			spell: game.Spell{
				ID:    "spell_cantrip",
				Name:  "Test Cantrip",
				Level: 0,
			},
			wantErr: false,
		},
		{
			name: "max level 9",
			spell: game.Spell{
				ID:    "spell_wish",
				Name:  "Wish",
				Level: 9,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSpell(tt.spell)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateItem(t *testing.T) {
	tests := []struct {
		name    string
		item    game.Item
		wantErr bool
	}{
		{
			name: "valid item",
			item: game.Item{
				ID:    "item_test",
				Name:  "Test Item",
				Value: 10,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			item: game.Item{
				Name:  "Test Item",
				Value: 10,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			item: game.Item{
				ID:    "item_test",
				Value: 10,
			},
			wantErr: true,
		},
		{
			name: "negative value",
			item: game.Item{
				ID:    "item_test",
				Name:  "Test Item",
				Value: -5,
			},
			wantErr: true,
		},
		{
			name: "zero value is valid",
			item: game.Item{
				ID:    "item_test",
				Name:  "Test Item",
				Value: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateItem(tt.item)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOutputSpell(t *testing.T) {
	spell := game.Spell{
		ID:          "spell_test",
		Name:        "Test Spell",
		Level:       1,
		School:      game.SchoolEvocation,
		Description: "A test spell",
	}

	// Test writing to file
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test_spell.yaml")

	err := outputSpell(spell, outputFile)
	require.NoError(t, err)

	// Verify file was created and contains expected content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "spell_test")
	assert.Contains(t, string(content), "Test Spell")
}

func TestOutputItem(t *testing.T) {
	item := game.Item{
		ID:    "item_test",
		Name:  "Test Item",
		Type:  "weapon",
		Value: 50,
	}

	// Test writing to file
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test_item.yaml")

	err := outputItem(item, outputFile)
	require.NoError(t, err)

	// Verify file was created and contains expected content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "item_test")
	assert.Contains(t, string(content), "Test Item")
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
			name: "invalid content type",
			cfg: &Config{
				ContentType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "spell template only",
			cfg: &Config{
				ContentType: "spell",
				Template:    "damage",
				OutputFile:  filepath.Join(tmpDir, "spell.yaml"),
			},
			wantErr: false,
		},
		{
			name: "item template only",
			cfg: &Config{
				ContentType: "item",
				Template:    "weapon",
				OutputFile:  filepath.Join(tmpDir, "item.yaml"),
			},
			wantErr: false,
		},
		{
			name: "spell no template non-interactive shows usage",
			cfg: &Config{
				ContentType: "spell",
			},
			wantErr: false,
		},
		{
			name: "item no template non-interactive shows usage",
			cfg: &Config{
				ContentType: "item",
			},
			wantErr: false,
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
			input:      "Custom Value\n",
			prompt:     "Test Prompt",
			defaultVal: "default",
			prefix:     "",
			want:       "Custom Value",
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
			prefix:     "spell_",
			want:       "spell_new",
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

func TestPromptSpellSchool(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		current game.SpellSchool
		want    game.SpellSchool
	}{
		{
			name:    "empty returns current",
			input:   "\n",
			current: game.SchoolEvocation,
			want:    game.SchoolEvocation,
		},
		{
			name:    "numeric input 0",
			input:   "0\n",
			current: game.SchoolEvocation,
			want:    game.SpellSchool(0), // Abjuration
		},
		{
			name:    "numeric input 4",
			input:   "4\n",
			current: game.SpellSchool(0),
			want:    game.SchoolEvocation, // 4 = evocation
		},
		{
			name:    "string input evocation",
			input:   "evocation\n",
			current: game.SpellSchool(0),
			want:    game.SchoolEvocation,
		},
		{
			name:    "string input divination",
			input:   "divination\n",
			current: game.SpellSchool(0),
			want:    game.SchoolDivination,
		},
		{
			name:    "invalid input returns current",
			input:   "invalid\n",
			current: game.SchoolEnchantment,
			want:    game.SchoolEnchantment,
		},
		{
			name:    "out of range number returns current",
			input:   "99\n",
			current: game.SchoolNecromancy,
			want:    game.SchoolNecromancy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			got := promptSpellSchool(reader, tt.current)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateSpell(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid template",
			cfg: &Config{
				ContentType: "spell",
				Template:    "damage",
				OutputFile:  filepath.Join(tmpDir, "spell1.yaml"),
			},
			wantErr: false,
		},
		{
			name: "invalid template",
			cfg: &Config{
				ContentType: "spell",
				Template:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "no template no interactive",
			cfg: &Config{
				ContentType: "spell",
			},
			wantErr: false, // shows usage
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createSpell(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateItem(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid template",
			cfg: &Config{
				ContentType: "item",
				Template:    "weapon",
				OutputFile:  filepath.Join(tmpDir, "item1.yaml"),
			},
			wantErr: false,
		},
		{
			name: "invalid template",
			cfg: &Config{
				ContentType: "item",
				Template:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "no template no interactive",
			cfg: &Config{
				ContentType: "item",
			},
			wantErr: false, // shows usage
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createItem(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Just verify printUsage doesn't panic
	assert.NotPanics(t, func() {
		printUsage()
	})
}

func TestOutputSpellStdout(t *testing.T) {
	spell := game.Spell{
		ID:    "spell_stdout",
		Name:  "Stdout Spell",
		Level: 0,
	}
	// Empty file outputs to stdout - should not error
	err := outputSpell(spell, "")
	assert.NoError(t, err)
}

func TestOutputItemStdout(t *testing.T) {
	item := game.Item{
		ID:    "item_stdout",
		Name:  "Stdout Item",
		Value: 0,
	}
	// Empty file outputs to stdout - should not error
	err := outputItem(item, "")
	assert.NoError(t, err)
}

func TestAllSpellTemplates(t *testing.T) {
	templates := []string{"damage", "healing", "buff", "debuff", "utility"}
	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			spell, err := spellFromTemplate(tmpl)
			require.NoError(t, err)
			assert.NotEmpty(t, spell.ID)
			assert.NotEmpty(t, spell.Name)
			err = validateSpell(spell)
			assert.NoError(t, err)
		})
	}
}

func TestAllItemTemplates(t *testing.T) {
	templates := []string{"weapon", "armor", "potion", "accessory"}
	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			item, err := itemFromTemplate(tmpl)
			require.NoError(t, err)
			assert.NotEmpty(t, item.ID)
			assert.NotEmpty(t, item.Name)
			err = validateItem(item)
			assert.NoError(t, err)
		})
	}
}
