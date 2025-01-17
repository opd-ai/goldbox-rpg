# game
--
    import "github.com/opd-ai/goldbox-rpg/pkg/game"

pkg/game/effect.go

## Usage

```go
const (
	// Effect Types
	EffectDamageOverTime EffectType = "damage_over_time"
	EffectHealOverTime   EffectType = "heal_over_time"
	EffectPoison         EffectType = "poison"
	EffectBurning        EffectType = "burning"
	EffectBleeding       EffectType = "bleeding"
	EffectStun           EffectType = "stun"
	EffectRoot           EffectType = "root"
	EffectStatBoost      EffectType = "stat_boost"
	EffectStatPenalty    EffectType = "stat_penalty"

	// Damage Types
	DamagePhysical  DamageType = "physical"
	DamageFire      DamageType = "fire"
	DamagePoison    DamageType = "poison"
	DamageFrost     DamageType = "frost"
	DamageLightning DamageType = "lightning"

	// Dispel Types
	DispelMagic   DispelType = "magic"
	DispelCurse   DispelType = "curse"
	DispelPoison  DispelType = "poison"
	DispelDisease DispelType = "disease"
	DispelAll     DispelType = "all"

	// Immunity Types
	ImmunityNone ImmunityType = iota
	ImmunityPartial
	ImmunityComplete
	ImmunityReflect

	// Dispel Priorities
	DispelPriorityLowest  DispelPriority = 0
	DispelPriorityLow     DispelPriority = 25
	DispelPriorityNormal  DispelPriority = 50
	DispelPriorityHigh    DispelPriority = 75
	DispelPriorityHighest DispelPriority = 100
)
```
Constants

```go
const (
	ItemTypeWeapon = "weapon"
	ItemTypeArmor  = "armor"
)
```
ItemType constants

#### func  ExampleEffectDispel

```go
func ExampleEffectDispel()
```
Example usage:

#### func  NewUID

```go
func NewUID() string
```
NewUID generates a unique identifier for game entities

#### func  SetLogger

```go
func SetLogger(l *log.Logger)
```
SetLogger allows changing the default logger

#### type Character

```go
type Character struct {
	ID          string   `yaml:"char_id"`          // Unique identifier
	Name        string   `yaml:"char_name"`        // Character's name
	Description string   `yaml:"char_description"` // Character's description
	Position    Position `yaml:"char_position"`    // Current location in game world

	// Attributes
	Strength     int `yaml:"attr_strength"`     // Physical power
	Dexterity    int `yaml:"attr_dexterity"`    // Agility and reflexes
	Constitution int `yaml:"attr_constitution"` // Health and stamina
	Intelligence int `yaml:"attr_intelligence"` // Learning and reasoning
	Wisdom       int `yaml:"attr_wisdom"`       // Intuition and perception
	Charisma     int `yaml:"attr_charisma"`     // Leadership and personality

	// Combat stats
	HP         int `yaml:"combat_current_hp"`  // Current hit points
	MaxHP      int `yaml:"combat_max_hp"`      // Maximum hit points
	ArmorClass int `yaml:"combat_armor_class"` // Defense rating
	THAC0      int `yaml:"combat_thac0"`       // To Hit Armor Class 0

	// Equipment and inventory
	Equipment map[EquipmentSlot]Item `yaml:"char_equipment"` // Equipped items by slot
	Inventory []Item                 `yaml:"char_inventory"` // Carried items
	Gold      int                    `yaml:"char_gold"`      // Currency amount
}
```

Character represents the base attributes for both Players and NPCs Contains all
attributes, stats, and equipment for game entities

#### func (*Character) FromJSON

```go
func (c *Character) FromJSON(data []byte) error
```

#### func (*Character) GetDescription

```go
func (c *Character) GetDescription() string
```

#### func (*Character) GetHealth

```go
func (c *Character) GetHealth() int
```

#### func (*Character) GetID

```go
func (c *Character) GetID() string
```
Implement GameObject interface methods

#### func (*Character) GetName

```go
func (c *Character) GetName() string
```

#### func (*Character) GetPosition

```go
func (c *Character) GetPosition() Position
```

#### func (*Character) GetTags

```go
func (c *Character) GetTags() []string
```

#### func (*Character) IsActive

```go
func (c *Character) IsActive() bool
```

#### func (*Character) IsObstacle

```go
func (c *Character) IsObstacle() bool
```

#### func (*Character) SetHealth

```go
func (c *Character) SetHealth(health int)
```

#### func (*Character) SetPosition

```go
func (c *Character) SetPosition(pos Position) error
```

#### func (*Character) ToJSON

```go
func (c *Character) ToJSON() ([]byte, error)
```

#### type CharacterClass

```go
type CharacterClass int
```

CharacterClass represents available character classes

```go
const (
	ClassFighter CharacterClass = iota
	ClassMage
	ClassCleric
	ClassThief
	ClassRanger
	ClassPaladin
)
```

#### func (CharacterClass) String

```go
func (cc CharacterClass) String() string
```

#### type ClassConfig

```go
type ClassConfig struct {
	Type         CharacterClass `yaml:"class_type"`        // The class enumeration value
	Name         string         `yaml:"class_name"`        // Display name of the class
	Description  string         `yaml:"class_description"` // Class description and background
	HitDice      string         `yaml:"class_hit_dice"`    // Hit points per level (e.g., "1d10")
	BaseSkills   []string       `yaml:"class_base_skills"` // Default skills for the class
	Abilities    []string       `yaml:"class_abilities"`   // Special class abilities
	Requirements struct {
		MinStr int `yaml:"min_strength"`     // Minimum strength requirement
		MinDex int `yaml:"min_dexterity"`    // Minimum dexterity requirement
		MinCon int `yaml:"min_constitution"` // Minimum constitution requirement
		MinInt int `yaml:"min_intelligence"` // Minimum intelligence requirement
		MinWis int `yaml:"min_wisdom"`       // Minimum wisdom requirement
		MinCha int `yaml:"min_charisma"`     // Minimum charisma requirement
	} `yaml:"class_requirements"` // Minimum stat requirements
}
```

ClassConfig represents the configuration for a character class Contains all
metadata and attributes for a specific class

#### type ClassProficiencies

```go
type ClassProficiencies struct {
	Class            CharacterClass `yaml:"class_type"`             // Associated character class
	WeaponTypes      []string       `yaml:"allowed_weapons"`        // Allowed weapon types
	ArmorTypes       []string       `yaml:"allowed_armor"`          // Allowed armor types
	ShieldProficient bool           `yaml:"can_use_shields"`        // Whether class can use shields
	Restrictions     []string       `yaml:"equipment_restrictions"` // Special equipment restrictions
}
```

ClassProficiencies represents weapon and armor proficiencies for a class

#### type DamageEffect

```go
type DamageEffect struct {
	Effect         *Effect    `yaml:",inline"` // Change to pointer
	DamageType     DamageType `yaml:"damage_type"`
	BaseDamage     float64    `yaml:"base_damage"`
	DamageScale    float64    `yaml:"damage_scale"`
	PenetrationPct float64    `yaml:"penetration_pct"`
}
```

DamageEffect represents effects that deal damage

#### func  AsDamageEffect

```go
func AsDamageEffect(e *Effect) (*DamageEffect, bool)
```
Add method to check if Effect is DamageEffect

#### func  CreateBleedingEffect

```go
func CreateBleedingEffect(baseDamage float64, duration time.Duration) *DamageEffect
```

#### func  CreateBurningEffect

```go
func CreateBurningEffect(baseDamage float64, duration time.Duration) *DamageEffect
```

#### func  CreatePoisonEffect

```go
func CreatePoisonEffect(baseDamage float64, duration time.Duration) *DamageEffect
```
Status effect creation functions

#### func  CreatePoisonEffectWithDispel

```go
func CreatePoisonEffectWithDispel(baseDamage float64, duration time.Duration) *DamageEffect
```
Example effect creation with dispel info

#### func  ToDamageEffect

```go
func ToDamageEffect(e *Effect) (*DamageEffect, bool)
```
Helper method to check and convert Effect to DamageEffect

#### func (*DamageEffect) GetEffect

```go
func (de *DamageEffect) GetEffect() *Effect
```
Add methods to properly access Effect fields

#### func (*DamageEffect) GetEffectType

```go
func (de *DamageEffect) GetEffectType() EffectType
```
Implement EffectTyper for DamageEffect

#### func (*DamageEffect) ToEffect

```go
func (de *DamageEffect) ToEffect() *Effect
```
Helper method to convert DamageEffect to Effect

#### type DamageType

```go
type DamageType string
```

Core types

#### type DialogCondition

```go
type DialogCondition struct {
	Type  string      `yaml:"condition_type"`  // Type of condition
	Value interface{} `yaml:"condition_value"` // Required value/state
}
```

DialogCondition represents requirements for dialog options

#### type DialogEntry

```go
type DialogEntry struct {
	ID         string            `yaml:"dialog_id"`         // Unique dialog identifier
	Text       string            `yaml:"dialog_text"`       // NPC's spoken text
	Responses  []DialogResponse  `yaml:"dialog_responses"`  // Player response options
	Conditions []DialogCondition `yaml:"dialog_conditions"` // Requirements to show dialog
}
```

DialogEntry represents a conversation node

#### type DialogResponse

```go
type DialogResponse struct {
	Text       string `yaml:"response_text"`        // Player's response text
	NextDialog string `yaml:"response_next_dialog"` // Following dialog ID
	Action     string `yaml:"response_action"`      // Triggered action
}
```

DialogResponse represents a player conversation choice

#### type Direction

```go
type Direction int
```

Direction represents cardinal directions in the game world

```go
const (
	North Direction = iota
	East
	South
	West
)
```

#### type DirectionConfig

```go
type DirectionConfig struct {
	Value       Direction `yaml:"direction_value"` // Numeric value of the direction
	Name        string    `yaml:"direction_name"`  // String representation (North, East, etc.)
	DegreeAngle int       `yaml:"direction_angle"` // Angle in degrees (0, 90, 180, 270)
}
```

DirectionConfig represents a serializable direction configuration

#### type DispelInfo

```go
type DispelInfo struct {
	Priority  DispelPriority `yaml:"dispel_priority"`
	Types     []DispelType   `yaml:"dispel_types"`
	Removable bool           `yaml:"dispel_removable"`
}
```

DispelInfo contains metadata about effect dispelling

#### type DispelPriority

```go
type DispelPriority int
```

Core types

#### type DispelType

```go
type DispelType string
```

Core types

#### type Duration

```go
type Duration struct {
	Rounds   int           `yaml:"duration_rounds"`
	Turns    int           `yaml:"duration_turns"`
	RealTime time.Duration `yaml:"duration_real"`
}
```

Duration represents a game time duration

#### type Effect

```go
type Effect struct {
	ID          string     `yaml:"effect_id"`
	Type        EffectType `yaml:"effect_type"`
	Name        string     `yaml:"effect_name"`
	Description string     `yaml:"effect_desc"`

	StartTime time.Time `yaml:"effect_start"`
	Duration  Duration  `yaml:"effect_duration"`
	TickRate  Duration  `yaml:"effect_tick_rate"`

	Magnitude  float64    `yaml:"effect_magnitude"`
	DamageType DamageType `yaml:"damage_type,omitempty"`

	SourceID   string `yaml:"effect_source"`
	SourceType string `yaml:"effect_source_type"`

	IsActive bool     `yaml:"effect_active"`
	Stacks   int      `yaml:"effect_stacks"`
	Tags     []string `yaml:"effect_tags"`

	DispelInfo DispelInfo `yaml:"dispel_info"`
	Modifiers  []Modifier `yaml:"effect_modifiers"`
}
```

Effect represents a game effect

#### func  CreateDamageEffect

```go
func CreateDamageEffect(effectType EffectType, damageType DamageType, damage float64, duration time.Duration) *Effect
```

#### func  NewEffect

```go
func NewEffect(effectType EffectType, duration Duration, magnitude float64) *Effect
```
Effect creation helpers

#### func  NewEffectWithDispel

```go
func NewEffectWithDispel(effectType EffectType, duration Duration, magnitude float64, dispelInfo DispelInfo) *Effect
```
Helper function to create effect with dispel info

#### func (*Effect) GetEffectType

```go
func (e *Effect) GetEffectType() EffectType
```
Implement EffectTyper for Effect

#### func (*Effect) IsExpired

```go
func (e *Effect) IsExpired(currentTime time.Time) bool
```
Add to Effect type in effects.go

#### func (*Effect) ShouldTick

```go
func (e *Effect) ShouldTick(currentTime time.Time) bool
```

#### type EffectHolder

```go
type EffectHolder interface {
	// Effect management
	AddEffect(effect *Effect) error
	RemoveEffect(effectID string) error
	HasEffect(effectType EffectType) bool
	GetEffects() []*Effect

	// Stats that can be modified by effects
	GetStats() *Stats
	SetStats(*Stats)

	// Base stats before effects
	GetBaseStats() *Stats
}
```

EffectHolder represents an entity that can have effects applied

#### type EffectManager

```go
type EffectManager struct {
}
```

EffectManager handles effect application and management

#### func  NewEffectManager

```go
func NewEffectManager(baseStats *Stats) *EffectManager
```
NewEffectManager creates a new effect manager

#### func (*EffectManager) AddImmunity

```go
func (em *EffectManager) AddImmunity(effectType EffectType, immunity ImmunityData)
```
AddImmunity adds or updates an immunity

#### func (*EffectManager) ApplyEffect

```go
func (em *EffectManager) ApplyEffect(effect *Effect) error
```
Update ApplyEffect to check immunities

#### func (*EffectManager) CheckImmunity

```go
func (em *EffectManager) CheckImmunity(effectType EffectType) *ImmunityData
```
CheckImmunity returns immunity status for an effect type

#### func (*EffectManager) DispelEffects

```go
func (em *EffectManager) DispelEffects(dispelType DispelType, count int) []string
```
DispelEffects removes effects based on type and count

#### func (*EffectManager) RemoveEffect

```go
func (em *EffectManager) RemoveEffect(effectID string) error
```
RemoveEffect removes an effect by ID

#### func (*EffectManager) UpdateEffects

```go
func (em *EffectManager) UpdateEffects(currentTime time.Time)
```
UpdateEffects processes all active effects

#### type EffectType

```go
type EffectType string
```

Core types

#### func (EffectType) AllowsStacking

```go
func (et EffectType) AllowsStacking() bool
```
Method to check if effect type allows stacking

#### type EffectTyper

```go
type EffectTyper interface {
	GetEffectType() EffectType
}
```

EffectTyper interface for getting effect type

#### type EquipmentSet

```go
type EquipmentSet struct {
	CharacterID string                                `yaml:"character_id"`    // ID of character owning the equipment
	Slots       map[EquipmentSlot]EquipmentSlotConfig `yaml:"equipment_slots"` // Map of all equipment slots
}
```

EquipmentSet represents a complete set of equipment slots for serialization

#### type EquipmentSlot

```go
type EquipmentSlot int
```

EquipmentSlot represents character equipment locations

```go
const (
	SlotHead EquipmentSlot = iota
	SlotNeck
	SlotChest
	SlotHands
	SlotRings
	SlotLegs
	SlotFeet
	SlotWeaponMain
	SlotWeaponOff
)
```

#### func (EquipmentSlot) String

```go
func (es EquipmentSlot) String() string
```

#### type EquipmentSlotConfig

```go
type EquipmentSlotConfig struct {
	Slot         EquipmentSlot `yaml:"slot_type"`        // Type of equipment slot
	Name         string        `yaml:"slot_name"`        // Display name for the slot
	Description  string        `yaml:"slot_description"` // Description of what can be equipped
	AllowedTypes []string      `yaml:"allowed_types"`    // Types of items that can be equipped
	Restricted   bool          `yaml:"slot_restricted"`  // Whether slot has special requirements
}
```

EquipmentSlotConfig represents serializable configuration for equipment slots

#### type EventHandler

```go
type EventHandler func(event GameEvent)
```

EventHandler represents a function that handles game events

#### type EventSystem

```go
type EventSystem struct {
}
```

EventSystem manages game event subscriptions and dispatching Provides
thread-safe event handling infrastructure

#### func  NewEventSystem

```go
func NewEventSystem() *EventSystem
```
NewEventSystem creates a new event system

#### func (*EventSystem) Emit

```go
func (es *EventSystem) Emit(event GameEvent)
```
Emit sends an event to all registered handlers

#### func (*EventSystem) Subscribe

```go
func (es *EventSystem) Subscribe(eventType EventType, handler EventHandler)
```
Subscribe registers a handler for a specific event type

#### type EventSystemConfig

```go
type EventSystemConfig struct {
	RegisteredTypes []EventType       `yaml:"registered_event_types"` // List of registered event types
	HandlerCount    map[EventType]int `yaml:"handler_counts"`         // Number of handlers per type
	AsyncHandling   bool              `yaml:"async_handling"`         // Whether events are handled asynchronously
}
```

EventSystemConfig represents serializable configuration for the event system

#### type EventType

```go
type EventType int
```

EventType represents different types of game events

```go
const (
	EventLevelUp EventType = iota
	EventCombatStart
	EventCombatEnd
	EventDamage
	EventDeath
	EventItemPickup
	EventItemDrop
	EventMovement
	EventSpellCast
	EventQuestUpdate
)
```

#### type GameEvent

```go
type GameEvent struct {
	Type      EventType              `yaml:"event_type"`      // Type of the event
	SourceID  string                 `yaml:"source_id"`       // ID of the event originator
	TargetID  string                 `yaml:"target_id"`       // ID of the event target
	Data      map[string]interface{} `yaml:"event_data"`      // Additional event data
	Timestamp int64                  `yaml:"event_timestamp"` // When the event occurred
}
```

GameEvent represents an event in the game Contains all metadata and payload for
event processing

#### type GameObject

```go
type GameObject interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetPosition() Position
	SetPosition(Position) error
	IsActive() bool
	GetTags() []string
	ToJSON() ([]byte, error)
	FromJSON([]byte) error
	GetHealth() int
	SetHealth(int)
	IsObstacle() bool
}
```

GameObject defines the interface for all interactive entities

#### type GameTime

```go
type GameTime struct {
	RealTime  time.Time `yaml:"time_real"`  // Actual system time
	GameTicks int64     `yaml:"time_ticks"` // Internal game time counter
	TimeScale float64   `yaml:"time_scale"` // Game/real time ratio
}
```

GameTime represents the in-game time system Manages game time progression and
real-time conversion

#### type ImmunityData

```go
type ImmunityData struct {
	Type       ImmunityType
	Duration   time.Duration
	Resistance float64
	ExpiresAt  time.Time
}
```

ImmunityData represents immunity information

#### type ImmunityType

```go
type ImmunityType int
```

Core types

#### type Item

```go
type Item struct {
	ID         string   `yaml:"item_id"`                    // Unique identifier for the item
	Name       string   `yaml:"item_name"`                  // Display name of the item
	Type       string   `yaml:"item_type"`                  // Category of item (weapon, armor, etc.)
	Damage     string   `yaml:"item_damage,omitempty"`      // Damage specification for weapons
	AC         int      `yaml:"item_armor_class,omitempty"` // Armor class for defensive items
	Weight     int      `yaml:"item_weight"`                // Weight in game units
	Value      int      `yaml:"item_value"`                 // Monetary value in game currency
	Properties []string `yaml:"item_properties,omitempty"`  // Special properties or effects
}
```

Item represents a game item with its properties Contains all attributes that
define an item's behavior and characteristics

#### type Level

```go
type Level struct {
	ID         string                 `yaml:"level_id"`         // Unique level identifier
	Name       string                 `yaml:"level_name"`       // Display name of the level
	Width      int                    `yaml:"level_width"`      // Width in tiles
	Height     int                    `yaml:"level_height"`     // Height in tiles
	Tiles      [][]Tile               `yaml:"level_tiles"`      // 2D grid of map tiles
	Properties map[string]interface{} `yaml:"level_properties"` // Custom level attributes
}
```

Level represents a game map/dungeon level Contains all data needed to render and
interact with a game area

#### type LootEntry

```go
type LootEntry struct {
	ItemID      string  `yaml:"loot_item_id"`      // Item identifier
	Chance      float64 `yaml:"loot_chance"`       // Drop probability
	MinQuantity int     `yaml:"loot_min_quantity"` // Minimum amount
	MaxQuantity int     `yaml:"loot_max_quantity"` // Maximum amount
}
```

LootEntry represents an item that can be dropped by an NPC

#### type ModOpType

```go
type ModOpType string
```


```go
const (
	ModAdd      ModOpType = "add"
	ModMultiply ModOpType = "multiply"
	ModSet      ModOpType = "set"
)
```

#### type Modifier

```go
type Modifier struct {
	Stat      string    `yaml:"mod_stat"`
	Value     float64   `yaml:"mod_value"`
	Operation ModOpType `yaml:"mod_operation"`
}
```


#### type NPC

```go
type NPC struct {
	Character `yaml:",inline"` // Base character attributes
	Behavior  string           `yaml:"npc_behavior"`   // AI behavior pattern
	Faction   string           `yaml:"npc_faction"`    // Allegiance group
	Dialog    []DialogEntry    `yaml:"npc_dialog"`     // Conversation options
	LootTable []LootEntry      `yaml:"npc_loot_table"` // Droppable items
}
```

NPC represents non-player characters Extends Character with AI and interaction
capabilities

#### type Player

```go
type Player struct {
	Character   `yaml:",inline"` // Base character attributes
	Class       CharacterClass   `yaml:"player_class"`      // Character's chosen class
	Level       int              `yaml:"player_level"`      // Current experience level
	Experience  int              `yaml:"player_experience"` // Total experience points
	QuestLog    []Quest          `yaml:"player_quests"`     // Active and completed quests
	KnownSpells []Spell          `yaml:"player_spells"`     // Learned/available spells
}
```

Player extends Character with player-specific functionality Contains all
attributes and mechanics specific to player characters

#### func (*Player) AddExperience

```go
func (p *Player) AddExperience(exp int) error
```
AddExperience safely adds experience points and handles level ups

#### func (*Player) FromJSON

```go
func (p *Player) FromJSON(data []byte) error
```
FromJSON implements GameObject. Subtle: this method shadows the method
(Character).FromJSON of Player.Character.

#### func (*Player) GetDescription

```go
func (p *Player) GetDescription() string
```
GetDescription implements GameObject. Subtle: this method shadows the method
(Character).GetDescription of Player.Character.

#### func (*Player) GetHealth

```go
func (p *Player) GetHealth() int
```
GetHealth implements GameObject.

#### func (*Player) GetID

```go
func (p *Player) GetID() string
```
GetID implements GameObject. Subtle: this method shadows the method
(Character).GetID of Player.Character.

#### func (*Player) GetName

```go
func (p *Player) GetName() string
```
GetName implements GameObject. Subtle: this method shadows the method
(Character).GetName of Player.Character.

#### func (*Player) GetPosition

```go
func (p *Player) GetPosition() Position
```
GetPosition implements GameObject. Subtle: this method shadows the method
(Character).GetPosition of Player.Character.

#### func (*Player) GetStats

```go
func (p *Player) GetStats() *Stats
```
Add this method to Player

#### func (*Player) GetTags

```go
func (p *Player) GetTags() []string
```
GetTags implements GameObject. Subtle: this method shadows the method
(Character).GetTags of Player.Character.

#### func (*Player) IsActive

```go
func (p *Player) IsActive() bool
```
IsActive implements GameObject. Subtle: this method shadows the method
(Character).IsActive of Player.Character.

#### func (*Player) IsObstacle

```go
func (p *Player) IsObstacle() bool
```
IsObstacle implements GameObject.

#### func (*Player) SetHealth

```go
func (p *Player) SetHealth(health int)
```
SetHealth implements GameObject.

#### func (*Player) SetPosition

```go
func (p *Player) SetPosition(pos Position) error
```
SetPosition implements GameObject. Subtle: this method shadows the method
(Character).SetPosition of Player.Character.

#### func (*Player) ToJSON

```go
func (p *Player) ToJSON() ([]byte, error)
```
ToJSON implements GameObject. Subtle: this method shadows the method
(Character).ToJSON of Player.Character.

#### type PlayerProgressData

```go
type PlayerProgressData struct {
	CurrentLevel       int `yaml:"progress_level"`          // Current level
	ExperiencePoints   int `yaml:"progress_exp"`            // Total XP
	NextLevelThreshold int `yaml:"progress_next_level_exp"` // XP needed for next level
	CompletedQuests    int `yaml:"progress_quests_done"`    // Number of completed quests
	SpellsLearned      int `yaml:"progress_spells_known"`   // Number of known spells
}
```

PlayerProgressData represents serializable player progress

#### type Position

```go
type Position struct {
	X      int       `yaml:"position_x"`      // X coordinate on the map grid
	Y      int       `yaml:"position_y"`      // Y coordinate on the map grid
	Level  int       `yaml:"position_level"`  // Current dungeon/map level
	Facing Direction `yaml:"position_facing"` // Direction the entity is facing
}
```

Position represents a location in the game world Contains coordinates and facing
direction for precise positioning

#### type Quest

```go
type Quest struct {
	ID          string           `yaml:"quest_id"`          // Unique quest identifier
	Title       string           `yaml:"quest_title"`       // Display title of the quest
	Description string           `yaml:"quest_description"` // Detailed quest description
	Status      QuestStatus      `yaml:"quest_status"`      // Current quest state
	Objectives  []QuestObjective `yaml:"quest_objectives"`  // List of quest goals
	Rewards     []QuestReward    `yaml:"quest_rewards"`     // Rewards for completion
}
```

Quest represents a game quest/mission Contains all information about a quest
including objectives and rewards

#### type QuestObjective

```go
type QuestObjective struct {
	Description string `yaml:"objective_description"` // What needs to be done
	Progress    int    `yaml:"objective_progress"`    // Current completion amount
	Required    int    `yaml:"objective_required"`    // Amount needed for completion
	Completed   bool   `yaml:"objective_completed"`   // Whether objective is done
}
```

QuestObjective represents a single goal or task within a quest

#### type QuestProgress

```go
type QuestProgress struct {
	QuestID            string `yaml:"progress_quest_id"`        // Associated quest ID
	ObjectivesComplete int    `yaml:"progress_objectives_done"` // Number of completed objectives
	TimeSpent          int    `yaml:"progress_time_spent"`      // Time spent on quest
	Attempts           int    `yaml:"progress_attempts"`        // Number of attempts
}
```

QuestProgress tracks overall quest completion metrics

#### type QuestReward

```go
type QuestReward struct {
	Type   string `yaml:"reward_type"`    // Type of reward (gold, item, exp)
	Value  int    `yaml:"reward_value"`   // Quantity or amount of reward
	ItemID string `yaml:"reward_item_id"` // Reference to reward item if applicable
}
```

QuestReward represents a reward given upon quest completion

#### type QuestStatus

```go
type QuestStatus int
```

QuestStatus represents the current state of a quest

```go
const (
	QuestNotStarted QuestStatus = iota
	QuestActive
	QuestCompleted
	QuestFailed
)
```

#### type RGB

```go
type RGB struct {
	R uint8 `yaml:"color_red"`   // Red component
	G uint8 `yaml:"color_green"` // Green component
	B uint8 `yaml:"color_blue"`  // Blue component
}
```

RGB represents a color in RGB format Each component ranges from 0-255

#### type Spell

```go
type Spell struct {
	ID          string           `yaml:"spell_id"`          // Unique identifier for the spell
	Name        string           `yaml:"spell_name"`        // Display name of the spell
	Level       int              `yaml:"spell_level"`       // Required caster level for the spell
	School      SpellSchool      `yaml:"spell_school"`      // Magic school classification
	Range       int              `yaml:"spell_range"`       // Range in game units
	Duration    int              `yaml:"spell_duration"`    // Duration in game turns
	Components  []SpellComponent `yaml:"spell_components"`  // Required components for casting
	Description string           `yaml:"spell_description"` // Full spell description and effects
}
```

Spell represents a magical ability that can be cast by characters. Contains all
the core attributes and metadata needed to define a spell effect.

#### type SpellComponent

```go
type SpellComponent int
```

SpellComponent represents the physical or verbal components required to cast a
spell

```go
const (
	ComponentVerbal SpellComponent = iota
	ComponentSomatic
	ComponentMaterial
)
```

#### type SpellSchool

```go
type SpellSchool int
```

SpellSchool represents the different schools of magic available in the game

```go
const (
	SchoolAbjuration SpellSchool = iota
	SchoolConjuration
	SchoolDivination
	SchoolEnchantment
	SchoolEvocation
	SchoolIllusion
	SchoolNecromancy
	SchoolTransmutation
)
```

#### type Stats

```go
type Stats struct {
	Health       float64
	Mana         float64
	Strength     float64
	Dexterity    float64
	Intelligence float64

	// Calculated stats
	MaxHealth float64
	MaxMana   float64
	Defense   float64
	Speed     float64
}
```

Stats represents an entity's modifiable attributes

#### func  NewDefaultStats

```go
func NewDefaultStats() *Stats
```

#### func (*Stats) Clone

```go
func (s *Stats) Clone() *Stats
```
Stats Clone method

#### type Tile

```go
type Tile struct {
	Type        TileType               `yaml:"tile_type"`        // Base type of the tile
	Walkable    bool                   `yaml:"tile_walkable"`    // Whether entities can move through
	Transparent bool                   `yaml:"tile_transparent"` // Whether light passes through
	Properties  map[string]interface{} `yaml:"tile_properties"`  // Custom property map

	// Visual properties
	Sprite string `yaml:"tile_sprite"` // Sprite/texture identifier
	Color  RGB    `yaml:"tile_color"`  // Base color tint

	// Special properties
	BlocksSight bool   `yaml:"tile_blocks_sight"` // Whether blocks line of sight
	Dangerous   bool   `yaml:"tile_dangerous"`    // Whether causes damage
	DamageType  string `yaml:"tile_damage_type"`  // Type of damage dealt
	Damage      int    `yaml:"tile_damage"`       // Amount of damage per turn
}
```

Tile represents a single map cell Contains all properties that define a tile's
behavior and appearance

#### func  NewFloorTile

```go
func NewFloorTile() Tile
```
Common tile factory functions

#### func  NewWallTile

```go
func NewWallTile() Tile
```

#### type TileType

```go
type TileType int
```

TileType represents different types of map tiles

```go
const (
	TileFloor TileType = iota
	TileWall
	TileDoor
	TileWater
	TileLava
	TilePit
	TileStairs
)
```

#### type World

```go
type World struct {
	Levels      []Level               `yaml:"world_levels"`       // All game levels/maps
	CurrentTime GameTime              `yaml:"world_current_time"` // Current game time
	Objects     map[string]GameObject `yaml:"world_objects"`      // All game objects by ID
	Players     map[string]*Player    `yaml:"world_players"`      // Active players by ID
	NPCs        map[string]*NPC       `yaml:"world_npcs"`         // Non-player characters by ID
	SpatialGrid map[Position][]string `yaml:"world_spatial_grid"` // Spatial index of objects
	Width       int                   `yaml:"world_width"`        // Width of the world
	Height      int                   `yaml:"world_height"`       // Height of the world
}
```

World manages the game state and all game objects Contains the complete state of
the game world including all entities and maps

#### func  NewWorld

```go
func NewWorld() *World
```
NewWorld creates a new game world instance

#### func (*World) AddObject

```go
func (w *World) AddObject(obj GameObject) error
```
AddObject safely adds a GameObject to the world

#### func (*World) GetObjectsAt

```go
func (w *World) GetObjectsAt(pos Position) []GameObject
```
GetObjectsAt returns all objects at a given position

#### func (*World) ValidateMove

```go
func (w *World) ValidateMove(player *Player, newPos Position) error
```
ValidateMove checks if the move is valid for the given player and position

#### type WorldConfig

```go
type WorldConfig struct {
	MaxPlayers      int      `yaml:"config_max_players"`      // Maximum allowed players
	MaxLevel        int      `yaml:"config_max_level"`        // Maximum character level
	StartingLevel   string   `yaml:"config_starting_level"`   // Initial player level ID
	EnabledFeatures []string `yaml:"config_enabled_features"` // Enabled world features
}
```

WorldConfig represents world configuration settings

#### type WorldState

```go
type WorldState struct {
	WorldVersion string     `yaml:"world_version"`       // World data version
	LastSaved    GameTime   `yaml:"world_last_saved"`    // Last save timestamp
	ActiveLevels []string   `yaml:"world_active_levels"` // Currently active level IDs
	Statistics   WorldStats `yaml:"world_stats"`         // World statistics
}
```

WorldState represents the serializable state of the world Used for
saving/loading game state

#### type WorldStats

```go
type WorldStats struct {
	TotalPlayers  int `yaml:"stat_total_players"`  // Total number of players
	ActiveNPCs    int `yaml:"stat_active_npcs"`    // Current active NPCs
	LoadedObjects int `yaml:"stat_loaded_objects"` // Total loaded objects
	ActiveQuests  int `yaml:"stat_active_quests"`  // Current active quests
	WorldAge      int `yaml:"stat_world_age"`      // Time since world creation
}
```

WorldStats tracks various world statistics
