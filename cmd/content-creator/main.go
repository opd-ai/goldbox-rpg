// Package main provides the content-creator CLI tool for creating spell and item YAML files.
// This tool provides template-driven creation with validation via pkg/validation.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"goldbox-rpg/pkg/cliutil"
	"goldbox-rpg/pkg/game"
)

// Config holds the command-line configuration for the content creator.
type Config struct {
	// OutputFile specifies the path to write the generated YAML.
	OutputFile string
	// ContentType specifies what to create: spell or item.
	ContentType string
	// Template specifies a template type for quick scaffolding.
	Template string
	// Interactive enables interactive mode for guided creation.
	Interactive bool
}

// parseFlags parses command-line flags and returns the configuration.
func parseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.OutputFile, "output", "", "output file path for generated YAML (default: stdout)")
	flag.StringVar(&cfg.OutputFile, "o", "", "output file path (shorthand)")
	flag.StringVar(&cfg.ContentType, "type", "", "content type to create: spell, item")
	flag.StringVar(&cfg.ContentType, "c", "", "content type (shorthand)")
	flag.StringVar(&cfg.Template, "template", "", "template type (depends on content type)")
	flag.StringVar(&cfg.Template, "t", "", "template type (shorthand)")
	flag.BoolVar(&cfg.Interactive, "interactive", false, "enable interactive mode")
	flag.BoolVar(&cfg.Interactive, "i", false, "interactive mode (shorthand)")
	flag.Usage = printUsage
	flag.Parse()
	return cfg
}

// printUsage prints the usage information for the content creator.
func printUsage() {
	fmt.Fprintf(os.Stderr, `Content Creator - Create spell and item YAML files for GoldBox RPG

Usage:
  content-creator -c TYPE [options]

Options:
  -c, --type TYPE        Content type: spell, item (required)
  -o, --output FILE      Write output to FILE instead of stdout
  -t, --template NAME    Use a template for quick scaffolding
  -i, --interactive      Enable interactive mode for guided creation
  -h, --help             Show this help message

Spell Templates (-c spell -t NAME):
  damage      - Offensive spell that deals damage
  healing     - Restorative spell that heals HP
  buff        - Spell that enhances abilities
  debuff      - Spell that weakens enemies
  utility     - Non-combat utility spell

Item Templates (-c item -t NAME):
  weapon      - Melee or ranged weapon
  armor       - Protective armor piece
  potion      - Consumable potion
  accessory   - Ring, amulet, or other accessory

Examples:
  # Create a damage spell from template
  content-creator -c spell -t damage -o fireball.yaml

  # Interactive spell creation
  content-creator -c spell -i -o new_spell.yaml

  # Create a weapon from template
  content-creator -c item -t weapon -o longsword.yaml

`)
}

// main is the entry point for the content creator application.
func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the content creator with the given configuration.
func run(cfg *Config) error {
	if cfg.ContentType == "" {
		printUsage()
		return nil
	}

	switch strings.ToLower(cfg.ContentType) {
	case "spell":
		return createSpell(cfg)
	case "item":
		return createItem(cfg)
	default:
		return fmt.Errorf("unknown content type: %s (use: spell, item)", cfg.ContentType)
	}
}

// createSpell handles spell content creation.
func createSpell(cfg *Config) error {
	var spell game.Spell

	// Start with template if specified
	if cfg.Template != "" {
		var err error
		spell, err = spellFromTemplate(cfg.Template)
		if err != nil {
			return err
		}
	}

	// Interactive mode
	if cfg.Interactive {
		var err error
		reader := bufio.NewReader(os.Stdin)
		spell, err = interactiveSpellWithReader(reader, spell)
		if err != nil {
			return err
		}
	} else if cfg.Template == "" {
		printUsage()
		return nil
	}

	// Validate
	if err := validateSpell(spell); err != nil {
		return fmt.Errorf("spell validation failed: %w", err)
	}

	return outputSpell(spell, cfg.OutputFile)
}

// spellFromTemplate creates a spell from a template.
func spellFromTemplate(templateType string) (game.Spell, error) {
	templates := map[string]game.Spell{
		"damage": {
			ID:          "spell_damage_001",
			Name:        "Firebolt",
			Level:       1,
			School:      game.SchoolEvocation,
			Range:       60,
			Duration:    0,
			Components:  []game.SpellComponent{game.ComponentVerbal, game.ComponentSomatic},
			Description: "Hurl a bolt of fire at a target, dealing fire damage.",
			DamageType:  "fire",
			DamageDice:  "1d10",
			AreaEffect:  false,
			SaveType:    "dexterity",
		},
		"healing": {
			ID:          "spell_heal_001",
			Name:        "Cure Wounds",
			Level:       1,
			School:      game.SchoolEvocation,
			Range:       0,
			Duration:    0,
			Components:  []game.SpellComponent{game.ComponentVerbal, game.ComponentSomatic},
			Description: "Restore hit points to a creature you touch.",
			HealingDice: "1d8+2",
			AreaEffect:  false,
		},
		"buff": {
			ID:             "spell_buff_001",
			Name:           "Bless",
			Level:          1,
			School:         game.SchoolEnchantment,
			Range:          30,
			Duration:       60,
			Components:     []game.SpellComponent{game.ComponentVerbal, game.ComponentSomatic, game.ComponentMaterial},
			Description:    "Bless up to three creatures, granting bonus to attack rolls and saving throws.",
			AreaEffect:     false,
			EffectKeywords: []string{"buff", "attack_bonus", "save_bonus"},
		},
		"debuff": {
			ID:             "spell_debuff_001",
			Name:           "Bane",
			Level:          1,
			School:         game.SchoolEnchantment,
			Range:          30,
			Duration:       60,
			Components:     []game.SpellComponent{game.ComponentVerbal, game.ComponentSomatic, game.ComponentMaterial},
			Description:    "Curse up to three creatures, penalizing their attack rolls and saving throws.",
			AreaEffect:     false,
			SaveType:       "charisma",
			EffectKeywords: []string{"debuff", "attack_penalty", "save_penalty"},
		},
		"utility": {
			ID:             "spell_utility_001",
			Name:           "Detect Magic",
			Level:          1,
			School:         game.SchoolDivination,
			Range:          0,
			Duration:       600,
			Components:     []game.SpellComponent{game.ComponentVerbal, game.ComponentSomatic},
			Description:    "Sense the presence of magic within 30 feet.",
			AreaEffect:     true,
			EffectKeywords: []string{"detection", "utility"},
		},
	}

	spell, ok := templates[strings.ToLower(templateType)]
	if !ok {
		return game.Spell{}, fmt.Errorf("unknown spell template: %s (valid: damage, healing, buff, debuff, utility)", templateType)
	}
	return spell, nil
}

// interactiveSpell guides spell creation interactively (deprecated wrapper).
func interactiveSpell(base game.Spell) (game.Spell, error) {
	reader := bufio.NewReader(os.Stdin)
	return interactiveSpellWithReader(reader, base)
}

// interactiveSpellWithReader guides spell creation interactively using provided reader.
func interactiveSpellWithReader(reader *bufio.Reader, base game.Spell) (game.Spell, error) {
	spell := base

	fmt.Println("\n=== Content Creator - Spell ===")
	fmt.Println()

	spell.ID = promptString(reader, "Spell ID", spell.ID, "spell_")
	spell.Name = promptString(reader, "Spell Name", spell.Name, "")
	spell.Level = promptInt(reader, "Spell Level (0-9)", spell.Level)
	spell.School = promptSpellSchool(reader, spell.School)
	spell.Range = promptInt(reader, "Range (0 for touch)", spell.Range)
	spell.Duration = promptInt(reader, "Duration (turns, 0 for instant)", spell.Duration)
	spell.Description = promptString(reader, "Description", spell.Description, "")
	spell.DamageType = promptString(reader, "Damage Type (or empty)", spell.DamageType, "")
	spell.DamageDice = promptString(reader, "Damage Dice (e.g., 2d6, or empty)", spell.DamageDice, "")
	spell.HealingDice = promptString(reader, "Healing Dice (e.g., 1d8+2, or empty)", spell.HealingDice, "")
	spell.SaveType = promptString(reader, "Save Type (or empty)", spell.SaveType, "")

	fmt.Print("Area Effect? (y/n) [n]: ")
	input, _ := reader.ReadString('\n')
	spell.AreaEffect = strings.ToLower(strings.TrimSpace(input)) == "y"

	return spell, nil
}

// promptSpellSchool prompts for a spell school selection.
func promptSpellSchool(reader *bufio.Reader, current game.SpellSchool) game.SpellSchool {
	schools := []string{"abjuration", "conjuration", "divination", "enchantment", "evocation", "illusion", "necromancy", "transmutation"}
	fmt.Println("Spell Schools: abjuration(0), conjuration(1), divination(2), enchantment(3),")
	fmt.Println("               evocation(4), illusion(5), necromancy(6), transmutation(7)")
	fmt.Printf("School [%d]: ", current)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return current
	}

	// Try as number first
	if num, err := strconv.Atoi(input); err == nil && num >= 0 && num <= 7 {
		return game.SpellSchool(num)
	}

	// Try as name
	input = strings.ToLower(input)
	for i, name := range schools {
		if name == input {
			return game.SpellSchool(i)
		}
	}

	fmt.Printf("Invalid school, using: %d\n", current)
	return current
}

// validateSpell validates a spell structure.
func validateSpell(spell game.Spell) error {
	if spell.ID == "" {
		return fmt.Errorf("spell ID is required")
	}
	if spell.Name == "" {
		return fmt.Errorf("spell name is required")
	}
	if spell.Level < 0 || spell.Level > 9 {
		return fmt.Errorf("spell level must be 0-9")
	}
	if spell.Range < 0 {
		return fmt.Errorf("spell range cannot be negative")
	}
	return nil
}

// outputSpell writes the spell to a file or stdout.
func outputSpell(spell game.Spell, outputFile string) error {
	return cliutil.OutputYAML(spell, cliutil.ContentInfo{
		Type: "Spell",
		Tool: "content-creator",
		ID:   spell.ID,
		Name: spell.Name,
	}, outputFile)
}

// createItem handles item content creation.
func createItem(cfg *Config) error {
	var item game.Item

	if cfg.Template != "" {
		var err error
		item, err = itemFromTemplate(cfg.Template)
		if err != nil {
			return err
		}
	}

	if cfg.Interactive {
		var err error
		reader := bufio.NewReader(os.Stdin)
		item, err = interactiveItemWithReader(reader, item)
		if err != nil {
			return err
		}
	} else if cfg.Template == "" {
		printUsage()
		return nil
	}

	if err := validateItem(item); err != nil {
		return fmt.Errorf("item validation failed: %w", err)
	}

	return outputItem(item, cfg.OutputFile)
}

// itemFromTemplate creates an item from a template.
func itemFromTemplate(templateType string) (game.Item, error) {
	templates := map[string]game.Item{
		"weapon": {
			ID:         "item_weapon_001",
			Name:       "Longsword",
			Type:       "weapon",
			Value:      15,
			Weight:     3,
			Damage:     "1d8",
			Properties: []string{"versatile", "slashing"},
		},
		"armor": {
			ID:         "item_armor_001",
			Name:       "Chain Mail",
			Type:       "armor",
			Value:      75,
			Weight:     55,
			AC:         16,
			Properties: []string{"heavy"},
		},
		"potion": {
			ID:         "item_potion_001",
			Name:       "Healing Potion",
			Type:       "consumable",
			Value:      50,
			Weight:     1,
			Properties: []string{"consumable", "healing", "2d4+2"},
		},
		"accessory": {
			ID:         "item_accessory_001",
			Name:       "Ring of Protection",
			Type:       "accessory",
			Value:      200,
			Weight:     0,
			Properties: []string{"ac_bonus_1", "save_bonus_1"},
		},
	}

	item, ok := templates[strings.ToLower(templateType)]
	if !ok {
		return game.Item{}, fmt.Errorf("unknown item template: %s (valid: weapon, armor, potion, accessory)", templateType)
	}
	return item, nil
}

// interactiveItem guides item creation interactively (deprecated wrapper).
func interactiveItem(base game.Item) (game.Item, error) {
	reader := bufio.NewReader(os.Stdin)
	return interactiveItemWithReader(reader, base)
}

// interactiveItemWithReader guides item creation interactively using provided reader.
func interactiveItemWithReader(reader *bufio.Reader, base game.Item) (game.Item, error) {
	item := base

	fmt.Println("\n=== Content Creator - Item ===")
	fmt.Println()

	item.ID = promptString(reader, "Item ID", item.ID, "item_")
	item.Name = promptString(reader, "Item Name", item.Name, "")
	item.Type = promptString(reader, "Item Type (weapon/armor/consumable/accessory/quest/misc)", item.Type, "")
	item.Value = promptInt(reader, "Value (gold)", item.Value)
	item.Weight = promptInt(reader, "Weight (lbs)", item.Weight)

	if item.Type == "weapon" {
		item.Damage = promptString(reader, "Damage Dice (e.g., 1d8)", item.Damage, "")
	}

	if item.Type == "armor" {
		item.AC = promptInt(reader, "Armor Class", item.AC)
	}

	// Properties
	fmt.Println("\n--- Item Properties ---")
	fmt.Println("Current properties:", item.Properties)
	fmt.Print("Add more properties? (comma-separated, or empty to keep): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		newProps := strings.Split(input, ",")
		for _, p := range newProps {
			p = strings.TrimSpace(p)
			if p != "" {
				item.Properties = append(item.Properties, p)
			}
		}
	}

	return item, nil
}

// validateItem validates an item structure.
func validateItem(item game.Item) error {
	if item.ID == "" {
		return fmt.Errorf("item ID is required")
	}
	if item.Name == "" {
		return fmt.Errorf("item name is required")
	}
	if item.Value < 0 {
		return fmt.Errorf("item value cannot be negative")
	}
	return nil
}

// outputItem writes the item to a file or stdout.
func outputItem(item game.Item, outputFile string) error {
	return cliutil.OutputYAML(item, cliutil.ContentInfo{
		Type: "Item",
		Tool: "content-creator",
		ID:   item.ID,
		Name: item.Name,
	}, outputFile)
}

// promptString prompts for a string value with a default.
func promptString(reader *bufio.Reader, prompt, defaultVal, prefix string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else if prefix != "" {
		fmt.Printf("%s [%s...]: ", prompt, prefix)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		if defaultVal != "" {
			return defaultVal
		}
		if prefix != "" {
			return prefix + "new"
		}
	}
	return input
}

// promptInt prompts for an integer value with a default.
func promptInt(reader *bufio.Reader, prompt string, defaultVal int) int {
	fmt.Printf("%s [%d]: ", prompt, defaultVal)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("Invalid number, using default: %d\n", defaultVal)
		return defaultVal
	}
	return val
}
