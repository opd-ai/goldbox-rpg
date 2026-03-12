package game

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Adventure represents a complete, self-contained adventure data pack.
// Each adventure contains maps, encounters, NPCs, items, quest chains,
// and metadata describing the adventure's theme and difficulty.
type Adventure struct {
	ID          string               `yaml:"adventure_id"          json:"id"`
	Slug        string               `yaml:"adventure_slug"        json:"slug"`
	Title       string               `yaml:"adventure_title"       json:"title"`
	Description string               `yaml:"adventure_description" json:"description"`
	Synopsis    string               `yaml:"adventure_synopsis"    json:"synopsis"`
	Theme       string               `yaml:"adventure_theme"       json:"theme"`
	MinLevel    int                  `yaml:"adventure_min_level"   json:"min_level"`
	MaxLevel    int                  `yaml:"adventure_max_level"   json:"max_level"`
	EstHours    string               `yaml:"adventure_est_hours"   json:"est_hours"`
	Author      string               `yaml:"adventure_author"      json:"author"`
	Version     string               `yaml:"adventure_version"     json:"version"`
	Maps        []AdventureMap       `yaml:"adventure_maps"        json:"maps"`
	NPCs        []AdventureNPC       `yaml:"adventure_npcs"        json:"npcs"`
	Items       []AdventureItem      `yaml:"adventure_items"       json:"items"`
	Encounters  []AdventureEncounter `yaml:"adventure_encounters" json:"encounters"`
	QuestChain  []AdventureQuest     `yaml:"adventure_quests"      json:"quests"`
}

// AdventureMap represents a map within an adventure.
type AdventureMap struct {
	ID     string `yaml:"map_id"     json:"id"`
	Name   string `yaml:"map_name"   json:"name"`
	Width  int    `yaml:"map_width"  json:"width"`
	Height int    `yaml:"map_height" json:"height"`
	File   string `yaml:"map_file"   json:"file"`
}

// AdventureNPC represents a non-player character within an adventure.
type AdventureNPC struct {
	ID          string `yaml:"npc_id"          json:"id"`
	Name        string `yaml:"npc_name"        json:"name"`
	Role        string `yaml:"npc_role"        json:"role"`
	Description string `yaml:"npc_description" json:"description"`
	Level       int    `yaml:"npc_level"       json:"level"`
	HP          int    `yaml:"npc_hp"          json:"hp"`
	Hostile     bool   `yaml:"npc_hostile"     json:"hostile"`
	Dialogue    string `yaml:"npc_dialogue"    json:"dialogue,omitempty"`
}

// AdventureItem represents a unique item defined by an adventure.
type AdventureItem struct {
	ID          string `yaml:"item_id"          json:"id"`
	Name        string `yaml:"item_name"        json:"name"`
	Type        string `yaml:"item_type"        json:"type"`
	Description string `yaml:"item_description" json:"description"`
	Rarity      string `yaml:"item_rarity"      json:"rarity"`
	Value       int    `yaml:"item_value"       json:"value"`
}

// AdventureEncounter represents a combat encounter within an adventure.
type AdventureEncounter struct {
	ID          string   `yaml:"encounter_id"          json:"id"`
	Name        string   `yaml:"encounter_name"        json:"name"`
	Description string   `yaml:"encounter_description" json:"description"`
	MapID       string   `yaml:"encounter_map_id"      json:"map_id"`
	MinLevel    int      `yaml:"encounter_min_level"    json:"min_level"`
	MaxLevel    int      `yaml:"encounter_max_level"    json:"max_level"`
	Enemies     []string `yaml:"encounter_enemies"      json:"enemies"`
	Rewards     []string `yaml:"encounter_rewards"      json:"rewards"`
}

// AdventureQuest represents a quest within an adventure's quest chain.
type AdventureQuest struct {
	ID          string           `yaml:"quest_id"          json:"id"`
	Title       string           `yaml:"quest_title"       json:"title"`
	Description string           `yaml:"quest_description" json:"description"`
	Order       int              `yaml:"quest_order"       json:"order"`
	Objectives  []QuestObjective `yaml:"quest_objectives"  json:"objectives"`
	Rewards     []QuestReward    `yaml:"quest_rewards"     json:"rewards"`
	NextQuest   string           `yaml:"quest_next"        json:"next_quest,omitempty"`
}

// AdventureSummary is a lightweight representation for listing adventures.
type AdventureSummary struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Theme       string `json:"theme"`
	MinLevel    int    `json:"min_level"`
	MaxLevel    int    `json:"max_level"`
	EstHours    string `json:"est_hours"`
	MapCount    int    `json:"map_count"`
	QuestCount  int    `json:"quest_count"`
}

// Summary returns a lightweight summary of the adventure for listing.
func (a *Adventure) Summary() AdventureSummary {
	return AdventureSummary{
		ID:          a.ID,
		Slug:        a.Slug,
		Title:       a.Title,
		Description: a.Description,
		Theme:       a.Theme,
		MinLevel:    a.MinLevel,
		MaxLevel:    a.MaxLevel,
		EstHours:    a.EstHours,
		MapCount:    len(a.Maps),
		QuestCount:  len(a.QuestChain),
	}
}

// Validate checks that the adventure data is well-formed.
func (a *Adventure) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("adventure missing ID")
	}
	if a.Slug == "" {
		return fmt.Errorf("adventure %q missing slug", a.ID)
	}
	if a.Title == "" {
		return fmt.Errorf("adventure %q missing title", a.ID)
	}
	if a.MinLevel < 1 {
		return fmt.Errorf("adventure %q min_level must be >= 1", a.ID)
	}
	if a.MaxLevel < a.MinLevel {
		return fmt.Errorf("adventure %q max_level must be >= min_level", a.ID)
	}
	return nil
}

// AdventureManager loads, caches, and provides access to adventure data packs.
type AdventureManager struct {
	mu         sync.RWMutex
	adventures map[string]*Adventure // keyed by slug
	dataDir    string                // root directory for adventure packs
}

// NewAdventureManager creates a new AdventureManager.
// dataDir should point to the data/adventures/ directory.
func NewAdventureManager(dataDir string) *AdventureManager {
	return &AdventureManager{
		adventures: make(map[string]*Adventure),
		dataDir:    dataDir,
	}
}

// LoadAll scans the data directory for adventure YAML files and loads them.
// Each adventure is expected to be in data/adventures/<slug>/adventure.yaml.
func (m *AdventureManager) LoadAll() error {
	logger := logrus.WithFields(logrus.Fields{
		"function": "LoadAll",
		"package":  "game",
		"dataDir":  m.dataDir,
	})
	logger.Debug("scanning for adventures")

	if _, err := os.Stat(m.dataDir); os.IsNotExist(err) {
		logger.Info("adventures directory does not exist, skipping")
		return nil
	}

	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		return fmt.Errorf("failed to read adventures directory: %w", err)
	}

	loaded := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		advFile := filepath.Join(m.dataDir, slug, "adventure.yaml")
		if _, err := os.Stat(advFile); os.IsNotExist(err) {
			logger.WithField("slug", slug).Debug("no adventure.yaml found, skipping")
			continue
		}

		adv, err := loadAdventureFile(advFile)
		if err != nil {
			logger.WithError(err).WithField("slug", slug).Warn("failed to load adventure")
			continue
		}

		if err := adv.Validate(); err != nil {
			logger.WithError(err).WithField("slug", slug).Warn("adventure validation failed")
			continue
		}

		// Ensure slug matches directory name
		adv.Slug = slug

		m.mu.Lock()
		m.adventures[slug] = adv
		m.mu.Unlock()
		loaded++
		logger.WithField("slug", slug).Info("loaded adventure")
	}

	logger.WithField("count", loaded).Info("adventure loading complete")
	return nil
}

// loadAdventureFile reads and parses a single adventure YAML file.
func loadAdventureFile(path string) (*Adventure, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var adv Adventure
	if err := yaml.Unmarshal(data, &adv); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return &adv, nil
}

// List returns summaries of all loaded adventures.
func (m *AdventureManager) List() []AdventureSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]AdventureSummary, 0, len(m.adventures))
	for _, adv := range m.adventures {
		summaries = append(summaries, adv.Summary())
	}
	return summaries
}

// Get returns a full adventure by slug.
func (m *AdventureManager) Get(slug string) (*Adventure, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adv, ok := m.adventures[slug]
	if !ok {
		return nil, fmt.Errorf("adventure %q not found", slug)
	}
	return adv, nil
}

// Count returns the number of loaded adventures.
func (m *AdventureManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.adventures)
}
