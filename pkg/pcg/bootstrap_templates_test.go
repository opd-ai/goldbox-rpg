package pcg

import (
	"os"
	"path/filepath"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadBootstrapTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	pcgDir := filepath.Join(tempDir, "pcg")
	err := os.MkdirAll(pcgDir, 0o755)
	require.NoError(t, err)

	// Create a test template file
	templateContent := `
default:
  game_length: "medium"
  complexity_level: "standard"
  genre_variant: "classic_fantasy"
  max_players: 4
  starting_level: 1
  world_seed: 0
  enable_quick_start: true
  data_directory: "data"

test_template:
  game_length: "long"
  complexity_level: "advanced"
  genre_variant: "grimdark"
  max_players: 6
  starting_level: 3
  world_seed: 12345
  enable_quick_start: false
  data_directory: "test_data"

simple_game:
  game_length: "short"
  complexity_level: "simple"
  genre_variant: "low_fantasy"
  max_players: 2
  starting_level: 1
  world_seed: 999
  enable_quick_start: true
  data_directory: "simple"
`

	templatesPath := filepath.Join(pcgDir, "bootstrap_templates.yaml")
	err = os.WriteFile(templatesPath, []byte(templateContent), 0o644)
	require.NoError(t, err)

	t.Run("Load existing template", func(t *testing.T) {
		config, err := LoadBootstrapTemplate("test_template", tempDir)
		require.NoError(t, err)

		assert.Equal(t, GameLengthLong, config.GameLength)
		assert.Equal(t, ComplexityAdvanced, config.ComplexityLevel)
		assert.Equal(t, GenreGrimdark, config.GenreVariant)
		assert.Equal(t, 6, config.MaxPlayers)
		assert.Equal(t, 3, config.StartingLevel)
		assert.Equal(t, int64(12345), config.WorldSeed)
		assert.False(t, config.EnableQuickStart)
		assert.Equal(t, "test_data", config.DataDirectory)
	})

	t.Run("Load non-existent template falls back to default", func(t *testing.T) {
		config, err := LoadBootstrapTemplate("non_existent", tempDir)
		require.NoError(t, err)

		// Should get the "default" template from the file
		assert.Equal(t, GameLengthMedium, config.GameLength)
		assert.Equal(t, ComplexityStandard, config.ComplexityLevel)
		assert.Equal(t, GenreClassicFantasy, config.GenreVariant)
		assert.Equal(t, 4, config.MaxPlayers)
		assert.Equal(t, 1, config.StartingLevel)
		assert.Equal(t, int64(0), config.WorldSeed)
		assert.True(t, config.EnableQuickStart)
		assert.Equal(t, "data", config.DataDirectory)
	})

	t.Run("Load another template", func(t *testing.T) {
		config, err := LoadBootstrapTemplate("simple_game", tempDir)
		require.NoError(t, err)

		assert.Equal(t, GameLengthShort, config.GameLength)
		assert.Equal(t, ComplexitySimple, config.ComplexityLevel)
		assert.Equal(t, GenreLowFantasy, config.GenreVariant)
		assert.Equal(t, 2, config.MaxPlayers)
		assert.Equal(t, 1, config.StartingLevel)
		assert.Equal(t, int64(999), config.WorldSeed)
		assert.True(t, config.EnableQuickStart)
		assert.Equal(t, "simple", config.DataDirectory)
	})
}

func TestLoadBootstrapTemplate_NoFile(t *testing.T) {
	// Test with a directory that doesn't have a templates file
	tempDir := t.TempDir()

	config, err := LoadBootstrapTemplate("any_template", tempDir)
	require.NoError(t, err)

	// Should get the hardcoded default config
	defaultConfig := DefaultBootstrapConfig()
	assert.Equal(t, defaultConfig.GameLength, config.GameLength)
	assert.Equal(t, defaultConfig.ComplexityLevel, config.ComplexityLevel)
	assert.Equal(t, defaultConfig.GenreVariant, config.GenreVariant)
	assert.Equal(t, defaultConfig.MaxPlayers, config.MaxPlayers)
	assert.Equal(t, defaultConfig.StartingLevel, config.StartingLevel)
	assert.Equal(t, defaultConfig.WorldSeed, config.WorldSeed)
	assert.Equal(t, defaultConfig.EnableQuickStart, config.EnableQuickStart)
	assert.Equal(t, defaultConfig.DataDirectory, config.DataDirectory)
}

func TestLoadBootstrapTemplate_InvalidYAML(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	pcgDir := filepath.Join(tempDir, "pcg")
	err := os.MkdirAll(pcgDir, 0o755)
	require.NoError(t, err)

	// Create an invalid YAML file
	invalidYAML := `
invalid: yaml: content: with: malformed: structure
  - this is not valid yaml
    badly indented
`

	templatesPath := filepath.Join(pcgDir, "bootstrap_templates.yaml")
	err = os.WriteFile(templatesPath, []byte(invalidYAML), 0o644)
	require.NoError(t, err)

	config, err := LoadBootstrapTemplate("any_template", tempDir)

	// Should return default config and an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse templates file")

	// Should still return a valid default config
	defaultConfig := DefaultBootstrapConfig()
	assert.Equal(t, defaultConfig.GameLength, config.GameLength)
}

func TestLoadBootstrapTemplate_NoDefaultTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	pcgDir := filepath.Join(tempDir, "pcg")
	err := os.MkdirAll(pcgDir, 0o755)
	require.NoError(t, err)

	// Create a template file without a "default" template
	templateContent := `
custom_only:
  game_length: "short"
  complexity_level: "simple"
  genre_variant: "classic_fantasy"
  max_players: 1
  starting_level: 5
  world_seed: 111
  enable_quick_start: false
  data_directory: "custom"
`

	templatesPath := filepath.Join(pcgDir, "bootstrap_templates.yaml")
	err = os.WriteFile(templatesPath, []byte(templateContent), 0o644)
	require.NoError(t, err)

	config, err := LoadBootstrapTemplate("non_existent", tempDir)
	require.NoError(t, err)

	// Should get the hardcoded default config since no "default" template exists
	defaultConfig := DefaultBootstrapConfig()
	assert.Equal(t, defaultConfig.GameLength, config.GameLength)
	assert.Equal(t, defaultConfig.ComplexityLevel, config.ComplexityLevel)
	assert.Equal(t, defaultConfig.GenreVariant, config.GenreVariant)
}

func TestListAvailableTemplates(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	pcgDir := filepath.Join(tempDir, "pcg")
	err := os.MkdirAll(pcgDir, 0o755)
	require.NoError(t, err)

	// Create a test template file
	templateContent := `
default:
  game_length: "medium"
  complexity_level: "standard"
  genre_variant: "classic_fantasy"
  max_players: 4
  starting_level: 1
  world_seed: 0
  enable_quick_start: true
  data_directory: "data"

epic_campaign:
  game_length: "long"
  complexity_level: "advanced"
  genre_variant: "classic_fantasy"
  max_players: 6
  starting_level: 1
  world_seed: 0
  enable_quick_start: false
  data_directory: "data"

quick_adventure:
  game_length: "short"
  complexity_level: "simple"
  genre_variant: "classic_fantasy"
  max_players: 4
  starting_level: 1
  world_seed: 0
  enable_quick_start: true
  data_directory: "data"
`

	templatesPath := filepath.Join(pcgDir, "bootstrap_templates.yaml")
	err = os.WriteFile(templatesPath, []byte(templateContent), 0o644)
	require.NoError(t, err)

	t.Run("List existing templates", func(t *testing.T) {
		templates, err := ListAvailableTemplates(tempDir)
		require.NoError(t, err)

		assert.Len(t, templates, 3)
		assert.Contains(t, templates, "default")
		assert.Contains(t, templates, "epic_campaign")
		assert.Contains(t, templates, "quick_adventure")
	})
}

func TestListAvailableTemplates_NoFile(t *testing.T) {
	// Test with a directory that doesn't have a templates file
	tempDir := t.TempDir()

	templates, err := ListAvailableTemplates(tempDir)
	require.NoError(t, err)
	assert.Empty(t, templates)
}

func TestListAvailableTemplates_InvalidYAML(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	pcgDir := filepath.Join(tempDir, "pcg")
	err := os.MkdirAll(pcgDir, 0o755)
	require.NoError(t, err)

	// Create an invalid YAML file
	invalidYAML := `invalid yaml content`

	templatesPath := filepath.Join(pcgDir, "bootstrap_templates.yaml")
	err = os.WriteFile(templatesPath, []byte(invalidYAML), 0o644)
	require.NoError(t, err)

	templates, err := ListAvailableTemplates(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse templates file")
	assert.Nil(t, templates)
}

func TestBootstrapTemplateIntegration(t *testing.T) {
	// Integration test: Load a template and use it to create a bootstrap instance
	tempDir := t.TempDir()
	pcgDir := filepath.Join(tempDir, "pcg")
	err := os.MkdirAll(pcgDir, 0o755)
	require.NoError(t, err)

	// Create a test template file with a specific configuration
	templateContent := `
integration_test:
  game_length: "short"
  complexity_level: "simple"
  genre_variant: "high_magic"
  max_players: 3
  starting_level: 2
  world_seed: 54321
  enable_quick_start: false
  data_directory: "integration_test"
`

	templatesPath := filepath.Join(pcgDir, "bootstrap_templates.yaml")
	err = os.WriteFile(templatesPath, []byte(templateContent), 0o644)
	require.NoError(t, err)

	// Load the template
	config, err := LoadBootstrapTemplate("integration_test", tempDir)
	require.NoError(t, err)

	// Verify the configuration was loaded correctly
	assert.Equal(t, GameLengthShort, config.GameLength)
	assert.Equal(t, ComplexitySimple, config.ComplexityLevel)
	assert.Equal(t, GenreHighMagic, config.GenreVariant)
	assert.Equal(t, 3, config.MaxPlayers)
	assert.Equal(t, 2, config.StartingLevel)
	assert.Equal(t, int64(54321), config.WorldSeed)
	assert.False(t, config.EnableQuickStart)
	assert.Equal(t, "integration_test", config.DataDirectory)

	// Create a bootstrap instance using the loaded config
	world := game.NewWorld()
	bootstrap := NewBootstrap(config, world, nil)

	// Verify the bootstrap was configured correctly
	assert.NotNil(t, bootstrap)
	assert.Equal(t, config, bootstrap.config)
}
