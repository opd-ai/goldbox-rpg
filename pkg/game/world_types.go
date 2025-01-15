package game

import "time"

// Level represents a game map/dungeon level
// Contains all data needed to render and interact with a game area
type Level struct {
	ID         string                 `yaml:"level_id"`         // Unique level identifier
	Name       string                 `yaml:"level_name"`       // Display name of the level
	Width      int                    `yaml:"level_width"`      // Width in tiles
	Height     int                    `yaml:"level_height"`     // Height in tiles
	Tiles      [][]Tile               `yaml:"level_tiles"`      // 2D grid of map tiles
	Properties map[string]interface{} `yaml:"level_properties"` // Custom level attributes
}

// GameTime represents the in-game time system
// Manages game time progression and real-time conversion
type GameTime struct {
	RealTime  time.Time `yaml:"time_real"`  // Actual system time
	GameTicks int64     `yaml:"time_ticks"` // Internal game time counter
	TimeScale float64   `yaml:"time_scale"` // Game/real time ratio
}

// NPC represents non-player characters
// Extends Character with AI and interaction capabilities
type NPC struct {
	Character `yaml:",inline"` // Base character attributes
	Behavior  string           `yaml:"npc_behavior"`   // AI behavior pattern
	Faction   string           `yaml:"npc_faction"`    // Allegiance group
	Dialog    []DialogEntry    `yaml:"npc_dialog"`     // Conversation options
	LootTable []LootEntry      `yaml:"npc_loot_table"` // Droppable items
}

// DialogEntry represents a conversation node
type DialogEntry struct {
	ID         string            `yaml:"dialog_id"`         // Unique dialog identifier
	Text       string            `yaml:"dialog_text"`       // NPC's spoken text
	Responses  []DialogResponse  `yaml:"dialog_responses"`  // Player response options
	Conditions []DialogCondition `yaml:"dialog_conditions"` // Requirements to show dialog
}

// DialogResponse represents a player conversation choice
type DialogResponse struct {
	Text       string `yaml:"response_text"`        // Player's response text
	NextDialog string `yaml:"response_next_dialog"` // Following dialog ID
	Action     string `yaml:"response_action"`      // Triggered action
}

// DialogCondition represents requirements for dialog options
type DialogCondition struct {
	Type  string      `yaml:"condition_type"`  // Type of condition
	Value interface{} `yaml:"condition_value"` // Required value/state
}

// LootEntry represents an item that can be dropped by an NPC
type LootEntry struct {
	ItemID      string  `yaml:"loot_item_id"`      // Item identifier
	Chance      float64 `yaml:"loot_chance"`       // Drop probability
	MinQuantity int     `yaml:"loot_min_quantity"` // Minimum amount
	MaxQuantity int     `yaml:"loot_max_quantity"` // Maximum amount
}
