package game

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// SpellManager handles loading, saving, and managing spells from YAML files
type SpellManager struct {
	spellsDir string
	spells    map[string]*Spell // Map of spell ID to spell
}

// NewSpellManager creates a new SpellManager instance
func NewSpellManager(spellsDir string) *SpellManager {
	return &SpellManager{
		spellsDir: spellsDir,
		spells:    make(map[string]*Spell),
	}
}

// LoadSpells loads all spell files from the spells directory
func (sm *SpellManager) LoadSpells() error {
	if _, err := os.Stat(sm.spellsDir); os.IsNotExist(err) {
		return fmt.Errorf("spells directory does not exist: %s", sm.spellsDir)
	}

	files, err := ioutil.ReadDir(sm.spellsDir)
	if err != nil {
		return fmt.Errorf("failed to read spells directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		filePath := filepath.Join(sm.spellsDir, file.Name())
		if err := sm.loadSpellFile(filePath); err != nil {
			return fmt.Errorf("failed to load spell file %s: %w", file.Name(), err)
		}
	}

	return nil
}

// loadSpellFile loads spells from a single YAML file
func (sm *SpellManager) loadSpellFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var collection SpellCollection
	if err := yaml.Unmarshal(data, &collection); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	for i := range collection.Spells {
		spell := &collection.Spells[i]

		// Validate spell
		if err := sm.validateSpell(spell); err != nil {
			return fmt.Errorf("invalid spell %s: %w", spell.ID, err)
		}

		sm.spells[spell.ID] = spell
	}

	return nil
}

// validateSpell ensures a spell has valid data
func (sm *SpellManager) validateSpell(spell *Spell) error {
	if spell.ID == "" {
		return fmt.Errorf("spell ID cannot be empty")
	}
	if spell.Name == "" {
		return fmt.Errorf("spell name cannot be empty")
	}
	if spell.Level < 0 {
		return fmt.Errorf("spell level cannot be negative")
	}
	if spell.Range < 0 {
		return fmt.Errorf("spell range cannot be negative")
	}
	if spell.Duration < 0 {
		return fmt.Errorf("spell duration cannot be negative")
	}
	return nil
}

// SaveSpell saves a single spell to a YAML file
func (sm *SpellManager) SaveSpell(spell *Spell, filename string) error {
	if err := sm.validateSpell(spell); err != nil {
		return fmt.Errorf("invalid spell: %w", err)
	}

	collection := SpellCollection{
		Spells: []Spell{*spell},
	}

	data, err := yaml.Marshal(collection)
	if err != nil {
		return fmt.Errorf("failed to marshal spell to YAML: %w", err)
	}

	filePath := filepath.Join(sm.spellsDir, filename)
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write spell file: %w", err)
	}

	// Add to memory
	sm.spells[spell.ID] = spell

	return nil
}

// SaveSpellsByLevel saves spells grouped by level to separate files
func (sm *SpellManager) SaveSpellsByLevel() error {
	// Group spells by level
	spellsByLevel := make(map[int][]Spell)

	for _, spell := range sm.spells {
		spellsByLevel[spell.Level] = append(spellsByLevel[spell.Level], *spell)
	}

	// Save each level to a separate file
	for level, spells := range spellsByLevel {
		// Sort spells by name for consistent output
		sort.Slice(spells, func(i, j int) bool {
			return spells[i].Name < spells[j].Name
		})

		collection := SpellCollection{Spells: spells}
		data, err := yaml.Marshal(collection)
		if err != nil {
			return fmt.Errorf("failed to marshal level %d spells: %w", level, err)
		}

		var filename string
		if level == 0 {
			filename = "cantrips.yaml"
		} else {
			filename = fmt.Sprintf("level%d.yaml", level)
		}

		filePath := filepath.Join(sm.spellsDir, filename)
		if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write level %d spells: %w", level, err)
		}
	}

	return nil
}

// GetSpell retrieves a spell by ID
func (sm *SpellManager) GetSpell(spellID string) (*Spell, error) {
	spell, exists := sm.spells[spellID]
	if !exists {
		return nil, fmt.Errorf("spell not found: %s", spellID)
	}

	// Return a copy to prevent external modification
	spellCopy := *spell
	return &spellCopy, nil
}

// GetSpellsByLevel returns all spells of a specific level
func (sm *SpellManager) GetSpellsByLevel(level int) []*Spell {
	var result []*Spell

	for _, spell := range sm.spells {
		if spell.Level == level {
			spellCopy := *spell
			result = append(result, &spellCopy)
		}
	}

	// Sort by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// GetSpellsBySchool returns all spells of a specific school
func (sm *SpellManager) GetSpellsBySchool(school SpellSchool) []*Spell {
	var result []*Spell

	for _, spell := range sm.spells {
		if spell.School == school {
			spellCopy := *spell
			result = append(result, &spellCopy)
		}
	}

	// Sort by level, then name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Level != result[j].Level {
			return result[i].Level < result[j].Level
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// GetAllSpells returns all loaded spells
func (sm *SpellManager) GetAllSpells() []*Spell {
	var result []*Spell

	for _, spell := range sm.spells {
		spellCopy := *spell
		result = append(result, &spellCopy)
	}

	// Sort by level, then name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Level != result[j].Level {
			return result[i].Level < result[j].Level
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// AddSpell adds a new spell to the manager
func (sm *SpellManager) AddSpell(spell *Spell) error {
	if err := sm.validateSpell(spell); err != nil {
		return fmt.Errorf("invalid spell: %w", err)
	}

	if _, exists := sm.spells[spell.ID]; exists {
		return fmt.Errorf("spell already exists: %s", spell.ID)
	}

	sm.spells[spell.ID] = spell
	return nil
}

// UpdateSpell updates an existing spell
func (sm *SpellManager) UpdateSpell(spell *Spell) error {
	if err := sm.validateSpell(spell); err != nil {
		return fmt.Errorf("invalid spell: %w", err)
	}

	if _, exists := sm.spells[spell.ID]; !exists {
		return fmt.Errorf("spell not found: %s", spell.ID)
	}

	sm.spells[spell.ID] = spell
	return nil
}

// RemoveSpell removes a spell from the manager
func (sm *SpellManager) RemoveSpell(spellID string) error {
	if _, exists := sm.spells[spellID]; !exists {
		return fmt.Errorf("spell not found: %s", spellID)
	}

	delete(sm.spells, spellID)
	return nil
}

// GetSpellCount returns the total number of loaded spells
func (sm *SpellManager) GetSpellCount() int {
	return len(sm.spells)
}

// GetSpellCountByLevel returns the number of spells at each level
func (sm *SpellManager) GetSpellCountByLevel() map[int]int {
	counts := make(map[int]int)

	for _, spell := range sm.spells {
		counts[spell.Level]++
	}

	return counts
}

// SearchSpells searches for spells by name or keywords
func (sm *SpellManager) SearchSpells(query string) []*Spell {
	var result []*Spell
	query = strings.ToLower(query)

	for _, spell := range sm.spells {
		// Search in name
		if strings.Contains(strings.ToLower(spell.Name), query) {
			spellCopy := *spell
			result = append(result, &spellCopy)
			continue
		}

		// Search in description
		if strings.Contains(strings.ToLower(spell.Description), query) {
			spellCopy := *spell
			result = append(result, &spellCopy)
			continue
		}

		// Search in keywords
		for _, keyword := range spell.EffectKeywords {
			if strings.Contains(strings.ToLower(keyword), query) {
				spellCopy := *spell
				result = append(result, &spellCopy)
				break
			}
		}
	}

	// Sort by level, then name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Level != result[j].Level {
			return result[i].Level < result[j].Level
		}
		return result[i].Name < result[j].Name
	})

	return result
}
