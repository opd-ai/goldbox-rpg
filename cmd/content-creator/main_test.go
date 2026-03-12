package main

import (
	"os"
	"path/filepath"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpellFromTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		wantErr     bool
		wantName    string
		wantLevel   int
		wantSchool  game.SpellSchool
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
