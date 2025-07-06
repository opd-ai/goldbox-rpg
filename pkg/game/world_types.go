package game

import "time"

// Level represents a game level/map with its dimensions, layout and properties.
// A level contains a 2D grid of Tiles and can be loaded from YAML configuration.
//
// Fields:
//   - ID: Unique string identifier for the level
//   - Name: Human readable display name for the level
//   - Width: Level width in number of tiles (must be > 0)
//   - Height: Level height in number of tiles (must be > 0)
//   - Tiles: 2D slice containing the level's tile grid, dimensions must match Width x Height
//   - Properties: Map of custom level attributes for game-specific data
//
// Related types:
//   - Tile: Individual map tile type used in the Tiles grid
//
// Usage:
//
//	level := &Level{
//	  ID: "level1",
//	  Name: "Tutorial Level",
//	  Width: 10,
//	  Height: 10,
//	  Tiles: make([][]Tile, height),
//	  Properties: make(map[string]interface{}),
//	}
type Level struct {
	ID         string                 `yaml:"level_id"`         // Unique level identifier
	Name       string                 `yaml:"level_name"`       // Display name of the level
	Width      int                    `yaml:"level_width"`      // Width in tiles
	Height     int                    `yaml:"level_height"`     // Height in tiles
	Tiles      [][]Tile               `yaml:"level_tiles"`      // 2D grid of map tiles
	Properties map[string]interface{} `yaml:"level_properties"` // Custom level attributes
}

// GameTime represents the in-game time system and manages game time progression
// Handles conversion between real time and game time using a configurable scale factor.
//
// Fields:
//   - RealTime: System time when game time was last updated
//   - GameTicks: Counter tracking elapsed game time units
//   - TimeScale: Multiplier for converting real time to game time (1.0 = realtime)
//
// Usage:
//
//	gameTime := &GameTime{
//	  RealTime: time.Now(),
//	  GameTicks: 0,
//	  TimeScale: 2.0, // Game time passes 2x faster than real time
//	}
//
// Related types:
//   - Level: Game levels track time for events and updates
//   - NPC: NPCs use game time for behavior and schedules
type GameTime struct {
	RealTime  time.Time `yaml:"time_real"`  // Actual system time
	GameTicks int64     `yaml:"time_ticks"` // Internal game time counter
	TimeScale float64   `yaml:"time_scale"` // Game/real time ratio
}

// GetCombatTurn returns the current combat round and turn index.
func (gt *GameTime) GetCombatTurn() (round, index int) {
	ticksPerTurn := int64(10) // 10 second turns
	totalTurns := gt.GameTicks / ticksPerTurn
	round = int(totalTurns / 6) // 6 turns per round
	index = int(totalTurns % 6)
	return
}

// IsSameTurn checks if this GameTime represents the same combat turn as another.
func (gt *GameTime) IsSameTurn(other GameTime) bool {
	r1, i1 := gt.GetCombatTurn()
	r2, i2 := other.GetCombatTurn()
	return r1 == r2 && i1 == i2
}

// NPC represents a non-player character in the game world
// Extends the base Character type with AI behaviors and interaction capabilities
//
// Fields:
//   - Character: Embedded base character attributes (health, stats, inventory etc)
//   - Behavior: AI behavior pattern ID determining how NPC acts (e.g. "guard", "merchant")
//   - Faction: Group allegiance affecting NPC relationships and interactions
//   - Dialog: Available conversation options when player interacts with NPC
//   - LootTable: Items that may be dropped when NPC dies
//
// Related types:
//   - Character: Base type providing core character functionality
//   - DialogEntry: Defines conversation nodes and options
//   - LootEntry: Defines droppable items and probabilities
//
// Usage:
//
//	npc := &NPC{
//	  Character: Character{Name: "Guard"},
//	  Behavior: "patrol",
//	  Faction: "town_guard",
//	  Dialog: []DialogEntry{...},
//	  LootTable: []LootEntry{...},
//	}
type NPC struct {
	Character `yaml:",inline"` // Base character attributes
	Behavior  string           `yaml:"npc_behavior"`   // AI behavior pattern
	Faction   string           `yaml:"npc_faction"`    // Allegiance group
	Dialog    []DialogEntry    `yaml:"npc_dialog"`     // Conversation options
	LootTable []LootEntry      `yaml:"npc_loot_table"` // Droppable items
}

// DialogEntry represents a single dialog interaction node in the game's conversation system.
// It contains the text spoken by an NPC, possible player responses, and conditions that must
// be met for this dialog to be available.
//
// Fields:
//   - ID: A unique string identifier for this dialog entry
//   - Text: The actual dialog text spoken by the NPC
//   - Responses: A slice of DialogResponse objects representing possible player choices
//   - Conditions: A slice of DialogCondition objects that must be satisfied for this dialog to appear
//
// Related types:
//   - DialogResponse: Represents a player's response option
//   - DialogCondition: Defines requirements that must be met
//
// Usage:
// Dialog entries are typically loaded from YAML configuration files and used by the
// dialog system to present NPC conversations to the player.
type DialogEntry struct {
	ID         string            `yaml:"dialog_id"`         // Unique dialog identifier
	Text       string            `yaml:"dialog_text"`       // NPC's spoken text
	Responses  []DialogResponse  `yaml:"dialog_responses"`  // Player response options
	Conditions []DialogCondition `yaml:"dialog_conditions"` // Requirements to show dialog
}

// DialogResponse represents a player conversation choice
// DialogResponse represents a player's response option in a dialog system.
// It contains the text shown to the player, the ID of the next dialog to trigger,
// and any associated game action to execute when this response is chosen.
//
// Fields:
//   - Text: The response text shown to the player as a dialog choice
//   - NextDialog: ID reference to the next dialog that should be triggered when this response is selected
//   - Action: Optional action identifier that will be executed when this response is chosen
//
// This struct is typically used as part of a larger Dialog structure to create branching conversations.
// The NextDialog field enables creating dialog trees by linking responses to subsequent dialog nodes.
type DialogResponse struct {
	Text       string `yaml:"response_text"`        // Player's response text
	NextDialog string `yaml:"response_next_dialog"` // Following dialog ID
	Action     string `yaml:"response_action"`      // Triggered action
}

// DialogCondition represents requirements for dialog options
// DialogCondition represents a condition that must be met for a dialog option or event to occur.
// It consists of a condition type and an associated value that needs to be satisfied.
//
// Fields:
//   - Type: The type of condition to check (e.g. "quest_complete", "has_item", etc.)
//   - Value: The required value or state for the condition to be met. Can be of any type
//     depending on the condition type.
//
// The specific validation and handling of conditions depends on the condition type.
// Custom condition types can be defined by implementing appropriate handlers.
type DialogCondition struct {
	Type  string      `yaml:"condition_type"`  // Type of condition
	Value interface{} `yaml:"condition_value"` // Required value/state
}

// LootEntry represents a single item drop configuration in the game's loot system.
// It defines the probability and quantity range for a specific item that can be obtained.
//
// Fields:
//   - ItemID: Unique identifier string for the item that can be dropped
//   - Chance: Float value between 0.0 and 1.0 representing drop probability percentage
//   - MinQuantity: Minimum number of items that can drop (must be >= 0)
//   - MaxQuantity: Maximum number of items that can drop (must be >= MinQuantity)
//
// Related types:
//   - Item - The actual item definition this entry references
//   - LootTable - Collection of LootEntry that defines all possible drops
type LootEntry struct {
	ItemID      string  `yaml:"loot_item_id"`      // Item identifier
	Chance      float64 `yaml:"loot_chance"`       // Drop probability
	MinQuantity int     `yaml:"loot_min_quantity"` // Minimum amount
	MaxQuantity int     `yaml:"loot_max_quantity"` // Maximum amount
}
