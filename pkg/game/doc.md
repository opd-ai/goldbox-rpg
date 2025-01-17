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
Constants EffectDamageOverTime represents an effect that deals damage to a
target over a period of time. It is commonly used for effects like poison,
burning, or bleeding that deal periodic damage. Related effects: EffectPoison,
EffectBurning, EffectBleeding Related damage types: DamagePhysical, DamageFire,
DamagePoison

```go
const (
	ItemTypeWeapon = "weapon"
	ItemTypeArmor  = "armor"
)
```
ItemType constants ItemTypeWeapon represents a weapon item type constant used
for categorizing items in the game inventory and equipment system. This type is
used when creating or identifying weapon items.

#### func  ExampleEffectDispel

```go
func ExampleEffectDispel()
```
Example usage: ExampleEffectDispel demonstrates how to create, apply and dispel
effects in the game. It shows: - Creating a poison effect with damage and
duration - Creating a curse effect with dispel properties - Applying effects to
an effect manager - Dispelling effects based on priority

This example illustrates the dispel system workflow: 1. Create effects with
dispel properties 2. Apply them to an effect manager 3. Selectively remove
effects using dispel type and count

Related types: - Effect: Base effect interface - EffectManager: Manages active
effects - DispelInfo: Controls how effects can be dispelled - DispelType:
Categories of dispel effects (curse, magic, etc)

The example does not handle errors from ApplyEffect() for simplicity. In
production code, these errors should be properly handled.

#### func  NewUID

```go
func NewUID() string
```
NewUID generates a unique identifier string by creating a random 8-byte sequence
and encoding it as a hexadecimal string.

Returns a 16-character hexadecimal string representing the random bytes.

Note: This function uses crypto/rand for secure random number generation. The
probability of collision is low but not zero. For cryptographic purposes or when
absolute uniqueness is required, consider using UUID instead.

Related: encoding/hex.EncodeToString()

#### func  SetLogger

```go
func SetLogger(l *log.Logger)
```
SetLogger allows changing the default logger SetLogger sets the package-level
logger instance used for logging throughout the game. It allows injection of a
custom logger for different logging implementations.

Parameters:

    - l: pointer to a log.Logger instance to be used for logging. Must not be nil.

Related:

    - logger (package-level variable)

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
attributes, stats, and equipment for game entities Character represents a
playable character or NPC in the game world. It encapsulates all attributes,
stats, and inventory management for characters.

Key features: - Thread-safe with sync.RWMutex protection - Complete attribute
system (Strength, Dexterity etc) - Combat stats tracking (HP, AC, THAC0) -
Equipment and inventory management - Position tracking in game world - Tagging
system for special attributes

The Character struct uses YAML tags for persistence and serialization. All
numeric fields use int type for simplicity and compatibility.

Related types: - Position: Represents location in game world - Item: Represents
equipment and inventory items - EquipmentSlot: Equipment slot enumeration

Thread safety: All operations that modify Character fields should hold mu.Lock()
Read operations should hold mu.RLock()

#### func (*Character) FromJSON

```go
func (c *Character) FromJSON(data []byte) error
```
FromJSON unmarshals a JSON byte array into the Character struct. This method is
thread-safe as it uses a mutex lock.

Parameters:

    - data []byte: JSON encoded byte array containing character data

Returns:

    - error: Returns any error that occurred during unmarshaling

Related:

    - Character.ToJSON() for serialization
    - json.Unmarshal() from encoding/json package

#### func (*Character) GetDescription

```go
func (c *Character) GetDescription() string
```
GetDescription returns the character's description as a string. This method is
thread-safe as it uses a read lock when accessing the description field.
Returns:

    - string: The character's description text

Related:

    - Character struct
    - Character.SetDescription()

#### func (*Character) GetHealth

```go
func (c *Character) GetHealth() int
```
GetHealth returns the current hit points (HP) of the Character.

Returns:

    - int: The current health points value

Related:

    - Character.HP field
    - Character.SetHealth (if exists)

#### func (*Character) GetID

```go
func (c *Character) GetID() string
```
Implement GameObject interface methods GetID returns the unique identifier
string for this Character instance. It uses a read lock to safely access the ID
field in a concurrent context. Returns the character's unique ID string.
Related: Character struct, ID field

#### func (*Character) GetName

```go
func (c *Character) GetName() string
```
GetName returns the name of the Character.

This method is thread-safe and uses a read lock to safely access the character's
name.

Returns:

    - string: The name of the character

Related:

    - Character struct

#### func (*Character) GetPosition

```go
func (c *Character) GetPosition() Position
```
GetPosition returns the current position of the Character. This method is
thread-safe and uses read locking to protect concurrent access. Returns a
Position struct containing the character's x,y coordinates. Related types: -
Position struct

#### func (*Character) GetTags

```go
func (c *Character) GetTags() []string
```
GetTags returns a copy of the character's tags list.

This method provides thread-safe access to the character's tags by using a read
lock. A new slice containing copies of all tags is returned to prevent external
modifications to the character's internal state.

Returns:

    []string - A new slice containing copies of all the character's tags

Related:

    Character.AddTag() - For adding new tags
    Character.RemoveTag() - For removing existing tags

#### func (*Character) IsActive

```go
func (c *Character) IsActive() bool
```
IsActive returns the current active state of the Character. This method is
concurrent-safe through use of a read lock.

Returns:

    - bool: true if the character is active, false otherwise

Thread-safety: This method uses RLock/RUnlock for concurrent access

#### func (*Character) IsObstacle

```go
func (c *Character) IsObstacle() bool
```
IsObstacle indicates if this Character should be treated as an obstacle for
movement/pathing. In the current implementation, all Characters are always
considered obstacles.

Returns:

    - bool: Always returns true since Characters are obstacles by default

Related:

    - Used by pathing and collision detection systems

#### func (*Character) SetActive

```go
func (c *Character) SetActive(active bool)
```
SetActive sets the active state of the character. Thread-safe method that
controls whether the character is active in the game.

Parameters:

    - active: bool - The desired active state for the character

Thread safety:

    Uses mutex locking to ensure thread-safe access to the active state

Related:

    - Character struct (contains the active field being modified)

#### func (*Character) SetHealth

```go
func (c *Character) SetHealth(health int)
```
SetHealth updates the character's current health points (HP) with the provided
value. The health value will be constrained between 0 and the character's
maximum HP.

Parameters:

    - health: The new health value to set (integer)

Edge cases handled:

    - Health below 0 is capped at 0
    - Health above MaxHP is capped at MaxHP

Related fields:

    - Character.HP
    - Character.MaxHP

#### func (*Character) SetPosition

```go
func (c *Character) SetPosition(pos Position) error
```
SetPosition updates the character's position to the specified coordinates after
validation.

Parameters:

    - pos Position: The new position coordinates to set

Returns:

    - error: nil if successful, error if position is invalid

Errors:

    - Returns error if position fails validation check

Thread Safety:

    - Method is thread-safe using mutex locking

Related:

    - isValidPosition() - Helper function that validates position coordinates

#### func (*Character) ToJSON

```go
func (c *Character) ToJSON() ([]byte, error)
```
ToJSON serializes the Character struct to JSON format with thread safety.

This method acquires a read lock on the character to ensure safe concurrent
access during serialization.

Returns:

    - []byte: The JSON encoded representation of the Character
    - error: Any error that occurred during marshaling

Related:

    - FromJSON() for deserialization
    - json.Marshal() from encoding/json

#### type CharacterClass

```go
type CharacterClass int
```

CharacterClass represents available character classes CharacterClass represents
the character's role or profession in the game. It is implemented as an
enumerated type using integers for efficient storage and comparison operations.

The specific class values and their gameplay implications should be defined as
constants using this type. Each class may have different abilities, starting
stats, and progression paths.

Related types: - Character struct (which likely contains this as a field) - Any
class-specific ability or skill types

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
String returns the string representation of a CharacterClass. It converts the
CharacterClass enum value to its corresponding human-readable name.

Returns:

    string: The name of the character class as a string ("Fighter", "Mage", etc.)

Notable Cases:

    - Assumes valid enum values within array bounds
    - Will panic if given an invalid CharacterClass value

Related Types:

    - CharacterClass type (enum)

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

ClassConfig represents the configuration for a character class in the game. It
defines all the attributes, requirements and abilities that make up a class.

Fields:

    - Type: The enumerated class type identifier (CharacterClass)
    - Name: Human-readable display name of the class
    - Description: Full text description and background lore of the class
    - HitDice: Dice expression for calculating hit points per level (e.g. "1d10")
    - BaseSkills: List of default skills granted by the class
    - Abilities: List of special abilities granted by the class
    - Requirements: Struct containing minimum ability score requirements:
    - MinStr: Minimum Strength score needed
    - MinDex: Minimum Dexterity score needed
    - MinCon: Minimum Constitution score needed
    - MinInt: Minimum Intelligence score needed
    - MinWis: Minimum Wisdom score needed
    - MinCha: Minimum Charisma score needed

This struct is typically populated from YAML configuration files. The
Requirements struct ensures characters meet minimum ability scores before being
allowed to take levels in this class.

Related types:

    - CharacterClass enum type (defines valid class types)
    - Character struct (uses ClassConfig for character creation)

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
ClassProficiencies defines what equipment and items a character class can use.
It specifies allowed weapons, armor types and any special restrictions.

Fields:

    - Class: The character class these proficiencies apply to
    - WeaponTypes: List of weapon categories this class can use (e.g. "sword", "bow")
    - ArmorTypes: List of armor categories this class can wear (e.g. "light", "heavy")
    - ShieldProficient: Whether the class is trained in shield usage
    - Restrictions: Any special limitations on equipment usage

Related types:

    - CharacterClass: The class enum these proficiencies are linked to

Example:

    Fighter proficiencies would allow all weapons and armor types with shield use
    Mage proficiencies would be limited to staves/wands and light armor with no shields

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

DamageEffect represents a damage-dealing effect in the game system. It extends
the base Effect struct with damage-specific attributes.

Fields:

    - Effect: Pointer to the base Effect struct containing common effect properties
    - DamageType: The type of damage dealt (e.g. physical, magical, etc)
    - BaseDamage: The base amount of damage before scaling
    - DamageScale: Multiplier applied to the base damage
    - PenetrationPct: Percentage of target's defense that is ignored

Related types:

    - Effect: Base effect type this extends
    - DamageType: Enum of possible damage types

Example usage:

    damageEffect := &DamageEffect{
      Effect: &Effect{},
      DamageType: Physical,
      BaseDamage: 10.0,
      DamageScale: 1.5,
      PenetrationPct: 0.25,
    }

#### func  AsDamageEffect

```go
func AsDamageEffect(e *Effect) (*DamageEffect, bool)
```
Add method to check if Effect is DamageEffect AsDamageEffect attempts to convert
a generic Effect into a DamageEffect.

Parameters:

    - e: A pointer to the Effect to convert

Returns:

    - *DamageEffect: A pointer to the created DamageEffect if conversion was successful
    - bool: True if conversion was successful, false otherwise

The function will only convert effects of type EffectPoison, EffectBurning, or
EffectBleeding. For all other effect types, it returns nil and false.

The resulting DamageEffect will: - Inherit the base Effect properties - Use the
original Effect's DamageType and Magnitude - Have DamageScale and PenetrationPct
set to 0

Related types:

    - Effect
    - DamageEffect
    - EffectType (EffectPoison, EffectBurning, EffectBleeding)

#### func  CreateBleedingEffect

```go
func CreateBleedingEffect(baseDamage float64, duration time.Duration) *DamageEffect
```
CreateBleedingEffect creates a new bleeding damage effect that deals continuous
physical damage over time

Parameters:

    - baseDamage: Base amount of physical damage dealt per tick (float64, must be >= 0)
    - duration: How long the bleeding effect lasts (time.Duration)

Returns:

    *DamageEffect - A configured bleeding damage effect that:
    - Deals physical damage over time
    - Ignores 50% of armor via penetration
    - Scales at 1.0x base damage

Related:

    - DamageEffect struct
    - NewEffect() - Base effect constructor
    - EffectBleeding constant
    - DamagePhysical constant

#### func  CreateBurningEffect

```go
func CreateBurningEffect(baseDamage float64, duration time.Duration) *DamageEffect
```
CreateBurningEffect creates a new fire-based damage effect that deals damage
over time

Parameters:

    - baseDamage: Base damage per tick (float64) that will be dealt
    - duration: How long the burning effect lasts (time.Duration)

Returns:

    *DamageEffect - A configured burning damage effect with:
    - Fire damage type
    - 20% damage scaling multiplier
    - No armor penetration
    - Real-time based duration tracking

Related:

    - DamageEffect
    - EffectBurning constant
    - DamageFire constant

#### func  CreatePoisonEffect

```go
func CreatePoisonEffect(baseDamage float64, duration time.Duration) *DamageEffect
```

#### func  CreatePoisonEffectWithDispel

```go
func CreatePoisonEffectWithDispel(baseDamage float64, duration time.Duration) *DamageEffect
```
CreatePoisonEffectWithDispel creates a poison damage effect that can be
dispelled. It extends the basic poison effect by adding dispel information.

Parameters:

    - baseDamage: The base damage per tick (must be >= 0)
    - duration: How long the poison effect lasts (must be > 0)

Returns:

    A DamageEffect pointer configured as a dispellable poison effect

Related:

    - CreatePoisonEffect - Base poison effect creator
    - DispelInfo - Structure defining dispel properties
    - DamageEffect - Base damage effect type

The effect can be removed by poison or magic dispel types with normal priority.

#### func  ToDamageEffect

```go
func ToDamageEffect(e *Effect) (*DamageEffect, bool)
```
Helper method to check and convert Effect to DamageEffect ToDamageEffect
attempts to convert a generic Effect to a DamageEffect.

Parameters:

    - e *Effect: The effect to convert. Must not be nil.

Returns:

    - *DamageEffect: The converted damage effect if successful, nil otherwise
    - bool: true if conversion was successful, false if effect type is not convertible

The function only converts poison, burning and bleeding effect types. All other
effect types will return nil and false.

Related types:

    - Effect
    - DamageEffect
    - EffectType (EffectPoison, EffectBurning, EffectBleeding)

#### func (*DamageEffect) GetEffect

```go
func (de *DamageEffect) GetEffect() *Effect
```
GetEffect returns the Effect object associated with this DamageEffect. This is
an accessor method that provides access to the underlying Effect field.

Returns:

    - *Effect: A pointer to the Effect object contained within this DamageEffect

Related types:

    - Effect type
    - DamageEffect type

#### func (*DamageEffect) GetEffectType

```go
func (de *DamageEffect) GetEffectType() EffectType
```
Implement EffectTyper for DamageEffect GetEffectType returns the type of this
DamageEffect

Returns:

    - EffectType: The type of effect this DamageEffect represents

Related:

    - EffectType interface
    - Effect.Type field

#### func (*DamageEffect) ToEffect

```go
func (de *DamageEffect) ToEffect() *Effect
```
Helper method to convert DamageEffect to Effect ToEffect converts a DamageEffect
to an Effect by returning the underlying Effect field. This method allows
DamageEffect to be used as an Effect type.

Returns:

    - *Effect: The underlying Effect pointer contained in the DamageEffect struct

Related Types:

    - Effect
    - DamageEffect

#### type DamageType

```go
type DamageType string
```

Core types EffectType represents a type of effect that can be applied to a game
entity in the RPG system. It is implemented as a string to allow for easy
extensibility and readable effect definitions.

Common effect types might include: - Damage - Healing - Status - Buff/Debuff

Related types: - DamageType - DispelType - ImmunityType

#### type DialogCondition

```go
type DialogCondition struct {
	Type  string      `yaml:"condition_type"`  // Type of condition
	Value interface{} `yaml:"condition_value"` // Required value/state
}
```

DialogCondition represents requirements for dialog options DialogCondition
represents a condition that must be met for a dialog option or event to occur.
It consists of a condition type and an associated value that needs to be
satisfied.

Fields:

    - Type: The type of condition to check (e.g. "quest_complete", "has_item", etc.)
    - Value: The required value or state for the condition to be met. Can be of any type
      depending on the condition type.

The specific validation and handling of conditions depends on the condition
type. Custom condition types can be defined by implementing appropriate
handlers.

#### type DialogEntry

```go
type DialogEntry struct {
	ID         string            `yaml:"dialog_id"`         // Unique dialog identifier
	Text       string            `yaml:"dialog_text"`       // NPC's spoken text
	Responses  []DialogResponse  `yaml:"dialog_responses"`  // Player response options
	Conditions []DialogCondition `yaml:"dialog_conditions"` // Requirements to show dialog
}
```

DialogEntry represents a single dialog interaction node in the game's
conversation system. It contains the text spoken by an NPC, possible player
responses, and conditions that must be met for this dialog to be available.

Fields:

    - ID: A unique string identifier for this dialog entry
    - Text: The actual dialog text spoken by the NPC
    - Responses: A slice of DialogResponse objects representing possible player choices
    - Conditions: A slice of DialogCondition objects that must be satisfied for this dialog to appear

Related types:

    - DialogResponse: Represents a player's response option
    - DialogCondition: Defines requirements that must be met

Usage: Dialog entries are typically loaded from YAML configuration files and
used by the dialog system to present NPC conversations to the player.

#### type DialogResponse

```go
type DialogResponse struct {
	Text       string `yaml:"response_text"`        // Player's response text
	NextDialog string `yaml:"response_next_dialog"` // Following dialog ID
	Action     string `yaml:"response_action"`      // Triggered action
}
```

DialogResponse represents a player conversation choice DialogResponse represents
a player's response option in a dialog system. It contains the text shown to the
player, the ID of the next dialog to trigger, and any associated game action to
execute when this response is chosen.

Fields:

    - Text: The response text shown to the player as a dialog choice
    - NextDialog: ID reference to the next dialog that should be triggered when this response is selected
    - Action: Optional action identifier that will be executed when this response is chosen

This struct is typically used as part of a larger Dialog structure to create
branching conversations. The NextDialog field enables creating dialog trees by
linking responses to subsequent dialog nodes.

#### type Direction

```go
type Direction int
```

Direction represents a cardinal direction in 2D space. It is implemented as an
integer type to allow for efficient direction comparisons and calculations.

```go
const (
	North Direction = iota // North direction (0 degrees)
	East                   // East direction (90 degrees)
	South                  // South direction (180 degrees)
	West                   // West direction (270 degrees)
)
```
Direction constants represent the four cardinal directions. These values are
used throughout the game for movement, facing, and orientation. The values
increment clockwise starting from North (0).

#### type DirectionConfig

```go
type DirectionConfig struct {
	Value       Direction `yaml:"direction_value"` // Numeric value of the direction
	Name        string    `yaml:"direction_name"`  // String representation (North, East, etc.)
	DegreeAngle int       `yaml:"direction_angle"` // Angle in degrees (0, 90, 180, 270)
}
```

DirectionConfig represents the configuration for a directional value in the game
system. It encapsulates direction-related properties including numeric values,
names and angular measurements.

Fields:

    - Value: Direction type representing the numeric/enum value of the direction
    - Name: String name of the direction (e.g. "North", "East")
    - DegreeAngle: Integer angle in degrees, must be one of: 0, 90, 180, 270

The DirectionConfig struct is typically loaded from YAML configuration files and
used to define cardinal directions in the game world.

Related types:

    - Direction (enum type)

#### type DispelInfo

```go
type DispelInfo struct {
	Priority  DispelPriority `yaml:"dispel_priority"`
	Types     []DispelType   `yaml:"dispel_types"`
	Removable bool           `yaml:"dispel_removable"`
}
```

DispelInfo contains metadata about how a game effect can be dispelled or
removed.

Fields:

    - Priority: Determines the order in which effects are dispelled (higher priority = dispelled first)
    - Types: List of dispel types that can remove this effect (e.g. magic, poison, curse)
    - Removable: Whether the effect can be removed at all

Related types:

    - DispelPriority: Priority level constants (0-100)
    - DispelType: Type of dispel (magic, curse, poison, etc)
    - Effect: Contains DispelInfo as a field

Example usage:

    info := DispelInfo{
        Priority: DispelPriorityNormal,
        Types: []DispelType{DispelMagic},
        Removable: true,
    }

#### type DispelPriority

```go
type DispelPriority int
```

Core types EffectType represents a type of effect that can be applied to a game
entity in the RPG system. It is implemented as a string to allow for easy
extensibility and readable effect definitions.

Common effect types might include: - Damage - Healing - Status - Buff/Debuff

Related types: - DamageType - DispelType - ImmunityType

#### type DispelType

```go
type DispelType string
```

Core types EffectType represents a type of effect that can be applied to a game
entity in the RPG system. It is implemented as a string to allow for easy
extensibility and readable effect definitions.

Common effect types might include: - Damage - Healing - Status - Buff/Debuff

Related types: - DamageType - DispelType - ImmunityType

#### type Duration

```go
type Duration struct {
	Rounds   int           `yaml:"duration_rounds"`
	Turns    int           `yaml:"duration_turns"`
	RealTime time.Duration `yaml:"duration_real"`
}
```

Duration represents a game time duration Duration represents time duration in a
game context, combining different time measurements. It can track duration in
rounds, turns, and real-world time simultaneously.

Fields:

    - Rounds: Number of combat/game rounds the duration lasts
    - Turns: Number of player/character turns the duration lasts
    - RealTime: Actual real-world time duration (uses time.Duration)

The zero value represents an instant/immediate duration with no lasting effect.
All fields are optional and can be combined - e.g. "2 rounds and 30 seconds"

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

	TargetID     string `yaml:"effect_target"`
	StatAffected string `yaml:"effect_stat_affected"`

	IsActive bool     `yaml:"effect_active"`
	Stacks   int      `yaml:"effect_stacks"`
	Tags     []string `yaml:"effect_tags"`

	DispelInfo DispelInfo `yaml:"dispel_info"`
	Modifiers  []Modifier `yaml:"effect_modifiers"`
}
```

Effect represents a game effect Effect represents a game effect that can be
applied to entities, modifying their stats or behavior over time. It contains
all the information needed to track, apply and manage status effects in the
game.

Fields:

    - ID: Unique identifier for the effect
    - Type: Category/type of the effect (e.g. buff, debuff, dot)
    - Name: Display name of the effect
    - Description: Detailed description of what the effect does
    - StartTime: When the effect was applied
    - Duration: How long the effect lasts
    - TickRate: How often the effect triggers/updates
    - Magnitude: Strength/value of the effect
    - DamageType: Type of damage if effect deals damage
    - SourceID: ID of entity that applied the effect
    - SourceType: Type of entity that applied the effect
    - TargetID: ID of entity the effect is applied to
    - StatAffected: Which stat the effect modifies
    - IsActive: Whether effect is currently active
    - Stacks: Number of times effect has stacked
    - Tags: Labels for categorizing/filtering effects
    - DispelInfo: Rules for removing/dispelling the effect
    - Modifiers: List of stat/attribute modifications

Related types:

    - EffectType: Type definition for effect categories
    - Duration: Custom time duration type
    - DamageType: Enumeration of damage types
    - DispelInfo: Rules for dispelling effects
    - Modifier: Definition of stat modifications

#### func  CreateDamageEffect

```go
func CreateDamageEffect(effectType EffectType, damageType DamageType, damage float64, duration time.Duration) *Effect
```
CreateDamageEffect creates a new damage-dealing Effect with the specified
parameters.

Parameters:

    - effectType: The type of effect being created (e.g. poison, bleed, etc)
    - damageType: The type of damage this effect deals (e.g. physical, magic, etc)
    - damage: Amount of damage dealt per tick (must be >= 0)
    - duration: How long the effect lasts in real time

Returns:

    A new *Effect configured to deal periodic damage of the specified type

The effect will tick once per second (defined in TickRate). Related types:

    - Effect
    - EffectType
    - DamageType
    - Duration

#### func  NewEffect

```go
func NewEffect(effectType EffectType, duration Duration, magnitude float64) *Effect
```
NewEffect creates a new Effect instance with the specified type, duration and
magnitude.

Parameters:

    - effectType: The type of effect to create (EffectType)
    - duration: How long the effect lasts (Duration struct with Rounds, Turns, RealTime)
    - magnitude: The strength/amount of the effect (float64)

Returns:

    - *Effect: A pointer to the newly created Effect instance with default values

The effect is initialized with:

    - A new unique ID
    - Active status
    - 1 stack
    - Default dispel info (lowest priority, not removable)
    - Empty slices for tags and modifiers
    - Current time as start time
    - Empty strings for name, description and other string fields

Related types:

    - Effect struct
    - EffectType type
    - Duration struct
    - DispelInfo struct

#### func  NewEffectWithDispel

```go
func NewEffectWithDispel(effectType EffectType, duration Duration, magnitude float64, dispelInfo DispelInfo) *Effect
```
Helper function to create effect with dispel info NewEffectWithDispel creates a
new Effect with dispel information. It extends NewEffect by adding dispel
information to handle dispelling/cleansing mechanics.

Parameters:

    - effectType: The type of effect to create (e.g. buff, debuff, etc)
    - duration: How long the effect lasts
    - magnitude: The strength/intensity of the effect
    - dispelInfo: Information about how this effect can be dispelled/cleansed

Returns:

    A pointer to the newly created Effect with dispel information

Related:

    - NewEffect - Base effect creation function
    - Effect - The main effect struct
    - DispelInfo - Struct containing dispel rules and mechanics

#### func (*Effect) GetEffectType

```go
func (e *Effect) GetEffectType() EffectType
```
Implement EffectTyper for Effect GetEffectType returns the type of the Effect.

Returns:

    - EffectType: The type classification of this effect.

Related types:

    - EffectType: An enumeration defining different effect categories
    - Effect: The parent struct containing effect data

#### func (*Effect) IsExpired

```go
func (e *Effect) IsExpired(currentTime time.Time) bool
```
IsExpired checks if the effect has expired based on either real time duration or
number of rounds.

Parameters:

    - currentTime time.Time: The current time to check against the effect's start time

Returns:

    - bool: true if the effect has expired, false otherwise

Notes: - For real-time based effects (Duration.RealTime > 0), checks if
currentTime is after startTime + duration - For round-based effects
(Duration.Rounds > 0), currently returns false (TODO: implementation needed) -
If neither duration type is set, effect never expires (returns false)

Related: - Duration struct containing RealTime and Rounds fields - Effect struct
containing StartTime and Duration fields

#### func (*Effect) ShouldTick

```go
func (e *Effect) ShouldTick(currentTime time.Time) bool
```
ShouldTick determines if the effect should trigger based on its tick rate. It
checks if enough real time has elapsed since the effect started for the next
tick to occur.

Parameters:

    - currentTime time.Time: The current timestamp to check against

Returns:

    - bool: true if the effect should tick, false otherwise

Edge cases:

    - Returns false if TickRate.RealTime is 0 to prevent infinite ticking
    - Uses modulo operation to determine regular intervals based on TickRate.RealTime

Related:

    - Effect.StartTime field
    - Effect.TickRate struct

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

EffectHolder represents an entity that can have effects applied EffectHolder
defines an interface for entities that can have effects applied to them. An
effect holder maintains both current stats (which include effect modifications)
and base stats (original values before effects).

Implementations must handle: - Effect management (add/remove/query) - Current
stats that can be modified by effects - Base stats that represent original
unmodified values

Related types: - Effect: Represents a single effect that can be applied - Stats:
Contains the actual stat values - EffectType: Enumeration of possible effect
types

#### type EffectManager

```go
type EffectManager struct {
}
```

EffectManager handles all temporary and permanent effects applied to an entity
in the game. It manages active effects, base and current stats, immunities,
resistances, and healing modifiers.

The manager maintains thread-safe access to its data structures through a mutex.

Fields:

    - activeEffects: Maps effect IDs to Effect instances currently applied
    - baseStats: Original unmodified stats of the entity
    - currentStats: Current stats after applying all active effects
    - immunities: Permanent immunity data mapped by effect type
    - tempImmunities: Temporary immunity data mapped by effect type
    - resistances: Damage/effect resistance multipliers (0-1) mapped by effect type
    - healingModifier: Multiplier affecting all healing received (1.0 = normal healing)

Related types:

    - Effect: Represents a single effect instance
    - Stats: Contains all modifiable entity statistics
    - EffectType: Enumeration of possible effect types
    - ImmunityData: Contains immunity duration and source information

#### func  NewEffectManager

```go
func NewEffectManager(baseStats *Stats) *EffectManager
```
NewEffectManager creates and initializes a new EffectManager instance.

Parameters:

    - baseStats: A pointer to Stats representing the base statistics that will be modified by effects.
      Must not be nil as it serves as the foundation for all stat calculations.

Returns:

    - *EffectManager: A new EffectManager instance with initialized maps for active effects,
      immunities, temporary immunities, and resistances. The current stats are initialized
      as a clone of the base stats.

Related types:

    - Stats: Base statistical values
    - Effect: Individual effect instances
    - EffectType: Types of effects that can be applied
    - ImmunityData: Immunity information for effect types

Note:

    - Automatically initializes default immunities via initializeDefaultImmunities()
    - All maps are initialized as empty but non-nil

#### func (*EffectManager) AddImmunity

```go
func (em *EffectManager) AddImmunity(effectType EffectType, immunity ImmunityData)
```
AddImmunity adds an immunity to a specific effect type to the EffectManager. If
the immunity has a duration > 0, it is added as a temporary immunity that will
expire after the specified duration. Otherwise, it is added as a permanent
immunity.

Parameters:

    - effectType: The type of effect to become immune to
    - immunity: ImmunityData struct containing duration and other immunity properties

The immunity is stored in either tempImmunities or immunities map based on
duration. If duration > 0, ExpiresAt is calculated as current time + duration.

Thread-safe through mutex locking.

Related:

    - ImmunityData struct
    - EffectType type

#### func (*EffectManager) ApplyEffect

```go
func (em *EffectManager) ApplyEffect(effect *Effect) error
```
ApplyEffect attempts to apply the given effect to the target, taking into
account any immunities. It handles different types of immunities including
complete immunity, reflection, and partial resistance.

Parameters:

    - effect: *Effect - The effect to be applied, containing type and magnitude information

Returns:

    - error: Returns an error if the effect cannot be applied (immunity/reflection) or if internal application fails

Error cases:

    - Returns error if target has complete immunity to the effect type
    - Returns error if effect is reflected
    - Panics if an unknown immunity type is encountered

Related:

    - CheckImmunity() - Called internally to determine immunity status
    - applyEffectInternal() - Called to handle actual effect application

#### func (*EffectManager) CheckImmunity

```go
func (em *EffectManager) CheckImmunity(effectType EffectType) *ImmunityData
```
CheckImmunity checks if there is an active immunity against the given effect
type. It first checks temporary immunities, then permanent immunities.

Parameters:

    - effectType: The type of effect to check immunity against

Returns:

    - *ImmunityData: Contains immunity details including:
    - Type: The type of immunity (temporary, permanent, or none)
    - Duration: How long the immunity lasts (0 for permanent)
    - Resistance: Resistance level against the effect (0-100)
    - ExpiresAt: When the immunity expires (empty for permanent)

Thread-safety: This method is thread-safe as it uses a read lock when accessing
the immunity maps.

Notable behaviors: - Automatically cleans up expired temporary immunities when
encountered - Returns a default ImmunityData with ImmunityNone if no immunity
exists - Temporary immunities take precedence over permanent ones

Related types: - ImmunityData - EffectType

#### func (*EffectManager) DispelEffects

```go
func (em *EffectManager) DispelEffects(dispelType DispelType, count int) []string
```
DispelEffects removes a specified number of active effects of a given dispel
type from the entity. It handles effect removal based on their dispel priority,
with higher priority effects being removed first.

Parameters:

    - dispelType: The type of dispel to apply (e.g., magic, curse, etc.). Using DispelAll will target all dispellable effects
    - count: Maximum number of effects to remove. Must be >= 0

Returns:

    - []string: Slice containing the IDs of all removed effects

Notable behaviors:

    - Thread-safe due to mutex locking
    - Only removes effects marked as removable
    - Automatically recalculates stats if any effects were removed
    - If count exceeds available effects, removes all eligible effects

Related types:

    - DispelType: Enum defining different types of dispel
    - DispelPriority: Defines removal priority of effects
    - Effect.DispelInfo: Contains dispel-related properties of an effect

#### func (*EffectManager) RemoveEffect

```go
func (em *EffectManager) RemoveEffect(effectID string) error
```
RemoveEffect removes an effect by ID RemoveEffect deactivates and removes an
effect from the active effects list by its ID.

Parameters:

    - effectID: string - The unique identifier of the effect to remove

Returns:

    - error: Returns nil if effect was successfully removed, or an error if effect was not found

Notable behavior: - Locks the EffectManager mutex during operation to ensure
thread safety - Sets effect's IsActive flag to false before removal - Triggers
recalculation of stats after removing the effect - Returns error if effect ID
does not exist in activeEffects map

Related: - recalculateStats() - Called after effect removal to update stats

#### func (*EffectManager) UpdateEffects

```go
func (em *EffectManager) UpdateEffects(currentTime time.Time)
```
UpdateEffects processes all active effects UpdateEffects processes and maintains
active effects based on the current time. It handles effect expiration, periodic
effect ticks, and stat recalculation.

Parameters:

    - currentTime time.Time: The current game time to check effects against

The method performs the following: - Removes expired effects from activeEffects
- Triggers periodic effect ticks when appropriate - Recalculates stats if any
effects were removed

Thread-safety: Uses mutex locking to safely modify shared state

Related: - EffectManager.processEffectTick() - EffectManager.recalculateStats()

#### type EffectType

```go
type EffectType string
```

Core types EffectType represents a type of effect that can be applied to a game
entity in the RPG system. It is implemented as a string to allow for easy
extensibility and readable effect definitions.

Common effect types might include: - Damage - Healing - Status - Buff/Debuff

Related types: - DamageType - DispelType - ImmunityType

#### func (EffectType) AllowsStacking

```go
func (et EffectType) AllowsStacking() bool
```
AllowsStacking determines whether effects of this type can stack with each
other. This method controls which effect types can have multiple instances
active at once on the same target.

Returns:

    - true if effects of this type can stack (EffectDamageOverTime, EffectHealOverTime, EffectStatBoost)
    - false for all other effect types

Related types:

    - EffectType: The enum type this method belongs to
    - Effect: The main effect struct that uses this stacking behavior

#### type EffectTyper

```go
type EffectTyper interface {
	GetEffectType() EffectType
}
```

EffectTyper is an interface that defines a contract for types that have an
associated effect type. It provides a common way to identify and categorize
different types of effects in the game.

Returns:

    - EffectType: The type classification of the effect

Related types:

    - EffectType: The enumeration of possible effect types

#### type EquipmentSet

```go
type EquipmentSet struct {
	CharacterID string                                `yaml:"character_id"`    // ID of character owning the equipment
	Slots       map[EquipmentSlot]EquipmentSlotConfig `yaml:"equipment_slots"` // Map of all equipment slots
}
```

EquipmentSet represents a character's complete set of equipped items across
different slots. This struct maintains the relationship between a character and
their equipped items.

Fields:

    - CharacterID: Unique identifier string for the character who owns this equipment set
    - Slots: Map containing the configuration for each equipment slot, keyed by EquipmentSlot type

The Slots map allows for flexible equipment configurations while enforcing
slot-specific validation rules defined in EquipmentSlotConfig.

Related types:

    - EquipmentSlot: Enum defining valid equipment slot types
    - EquipmentSlotConfig: Configuration for individual equipment slots

#### type EquipmentSlot

```go
type EquipmentSlot int
```

EquipmentSlot represents the different slots where equipment/items can be
equipped on a character. This type is used as an enum to identify valid
equipment positions (e.g. weapon slot, armor slot, etc).

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
String returns a human-readable string representation of an EquipmentSlot. This
method maps the numeric equipment slot enum value to its corresponding string
name from a fixed array of slot names.

Returns:

    - string: The name of the equipment slot (one of: Head, Neck, Chest, Hands,
      Rings, Legs, Feet, MainHand, OffHand)

Note: This method will panic if the EquipmentSlot value is outside the valid
range (0-8) as it directly indexes into a fixed array.

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

EquipmentSlotConfig defines the configuration for an equipment slot in the game.
It specifies what types of items can be equipped and any special requirements.

Fields:

    - Slot: The type of equipment slot (e.g. weapon, armor, etc)
    - Name: Human readable display name for the equipment slot
    - Description: Detailed description of what items can be equipped in this slot
    - AllowedTypes: List of item type IDs that can be equipped in this slot
    - Restricted: If true, additional requirements must be met to use this slot

Related types:

    - EquipmentSlot (enum type for slot categories)
    - Item (for equippable items)

#### type EventHandler

```go
type EventHandler func(event GameEvent)
```

EventHandler is a function type that handles game events in the game system. It
takes a GameEvent parameter and processes it according to the specific event
handling logic.

Parameters:

    - event GameEvent: The game event to be handled

Note: EventHandler functions are typically used as callbacks registered to
handle specific types of game events in an event-driven architecture.

Related types:

    - GameEvent (defined elsewhere in the codebase)

#### type EventSystem

```go
type EventSystem struct {
}
```

EventSystem manages event handling and dispatching in the game. It provides a
thread-safe way to register handlers for different event types and dispatch
events to all registered handlers.

Fields:

    - mu: sync.RWMutex for ensuring thread-safe access to handlers
    - handlers: Map storing event handlers organized by EventType

Thread Safety: All methods on EventSystem are thread-safe and can be called
concurrently from multiple goroutines.

Related Types:

    - EventType: Type definition for different kinds of game events
    - EventHandler: Interface for handling dispatched events

#### func  NewEventSystem

```go
func NewEventSystem() *EventSystem
```
NewEventSystem creates and initializes a new event system. It initializes an
empty map of event handlers that can be registered to handle different event
types.

Returns:

    - *EventSystem: A pointer to the newly created event system with an initialized
      empty handlers map.

Related types: - EventType: The type used to identify different kinds of events
- EventHandler: Function type for handling specific events

#### func (*EventSystem) Emit

```go
func (es *EventSystem) Emit(event GameEvent)
```
Emit asynchronously distributes a game event to all registered handlers for that
event type. It safely accesses the handlers map using a read lock to prevent
concurrent map access issues.

Parameters:

    - event GameEvent: The game event to be processed. Must contain a valid Type field that
      matches registered handler types.

Thread-safety:

    - Uses RWMutex to safely access handlers map
    - Handlers are executed concurrently in separate goroutines

Related types:

    - GameEvent interface
    - EventHandler func type
    - EventType enum

#### func (*EventSystem) Subscribe

```go
func (es *EventSystem) Subscribe(eventType EventType, handler EventHandler)
```
Subscribe registers a new event handler for a specific event type. The handler
will be called when events of the specified type are published.

Parameters:

    - eventType: The type of event to subscribe to
    - handler: The event handler function to be called when events occur

Thread safety: This method is thread-safe as it uses mutex locking.

Related:

    - EventType
    - EventHandler
    - EventSystem.Publish

#### type EventSystemConfig

```go
type EventSystemConfig struct {
	RegisteredTypes []EventType       `yaml:"registered_event_types"` // List of registered event types
	HandlerCount    map[EventType]int `yaml:"handler_counts"`         // Number of handlers per type
	AsyncHandling   bool              `yaml:"async_handling"`         // Whether events are handled asynchronously
}
```

EventSystemConfig defines the configuration settings for the event handling
system. It manages event type registration, handler tracking, and processing
behavior.

Fields:

    - RegisteredTypes: Slice of EventType that are registered in the system.
    - HandlerCount: Map tracking number of handlers registered for each EventType.
      A count of 0 indicates no handlers are registered for that type.
    - AsyncHandling: Boolean flag determining if events are processed asynchronously.
      When true, events are handled in separate goroutines.
      When false, events are handled synchronously in the calling goroutine.

The config should be initialized before registering any event handlers.
AsyncHandling should be used with caution as it may affect event ordering.

Related:

    - EventType: Type definition for supported event types
    - EventHandler: Interface for event handler implementations

#### type EventType

```go
type EventType int
```

EventType represents different types of game events EventType represents the
type of an event in the game. It is implemented as an integer enum to allow for
efficient comparison and switching. The specific event type values should be
defined as constants using this type.

Related types:

    - Event interface (if exists)
    - Any concrete event types that use this enum

```go
const (
	EventLevelUp EventType = iota
	EventDamage
	EventDeath
	EventItemPickup
	EventItemDrop
	EventMovement
	EventSpellCast
	EventQuestUpdate
)
```
EventLevelUp represents a character gaining a level. This event is triggered
when a character accumulates enough experience points to advance to the next
level. The event carries information about: - The character that leveled up -
The new level achieved - Any stat increases or new abilities gained

Related events: - EventDamage: May contribute to experience gain -
EventQuestUpdate: Quests may require reaching certain levels

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

Contains all metadata and payload for event processing GameEvent represents an
occurrence or action within the game system that needs to be tracked or handled.
It contains information about what happened, who/what was involved, and when it
occurred.

Fields:

    - Type: The category/classification of the event (EventType)
    - SourceID: Unique identifier for the entity that triggered/caused the event
    - TargetID: Unique identifier for the entity that the event affects/targets
    - Data: Additional contextual information about the event as key-value pairs
    - Timestamp: Unix timestamp (in seconds) when the event occurred

The GameEvent struct is used throughout the event system to standardize how game
occurrences are represented and processed. Events can represent things like
combat actions, item usage, movement, etc.

Related types:

    - EventType: Enumeration of possible event categories

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

GameObject represents a base interface for all game objects in the RPG system.
It defines the core functionality that every game object must implement.

Core capabilities include: - Unique identification (GetID) - Basic properties
(name, description, position) - State management (active status, health) -
Tag-based classification - JSON serialization/deserialization - Collision
detection (obstacle status)

Related types: - Position: Represents the object's location in the game world

Implementation note: All game objects should implement this interface to ensure
consistent behavior across the game system. This enables uniform handling of
different object types in the game loop and collision detection systems.

The interface is designed to be extensible - additional specialized interfaces
can embed GameObject to add more specific functionality while maintaining
compatibility with base game systems.

#### type GameTime

```go
type GameTime struct {
	RealTime  time.Time `yaml:"time_real"`  // Actual system time
	GameTicks int64     `yaml:"time_ticks"` // Internal game time counter
	TimeScale float64   `yaml:"time_scale"` // Game/real time ratio
}
```

GameTime represents the in-game time system and manages game time progression
Handles conversion between real time and game time using a configurable scale
factor.

Fields:

    - RealTime: System time when game time was last updated
    - GameTicks: Counter tracking elapsed game time units
    - TimeScale: Multiplier for converting real time to game time (1.0 = realtime)

Usage:

    gameTime := &GameTime{
      RealTime: time.Now(),
      GameTicks: 0,
      TimeScale: 2.0, // Game time passes 2x faster than real time
    }

Related types:

    - Level: Game levels track time for events and updates
    - NPC: NPCs use game time for behavior and schedules

#### type ImmunityData

```go
type ImmunityData struct {
	Type       ImmunityType
	Duration   time.Duration
	Resistance float64
	ExpiresAt  time.Time
}
```

ImmunityData represents immunity effects that can be applied to game entities.
It tracks the type, duration, resistance level and expiration time of
immunities.

Fields:

    - Type: The type/category of immunity effect (ImmunityType)
    - Duration: How long the immunity lasts (time.Duration)
    - Resistance: A value between 0-1 representing immunity strength
    - ExpiresAt: Timestamp when immunity effect ends

Related types:

    - ImmunityType: Enumeration of possible immunity types

The immunity effect expires when current time exceeds ExpiresAt. Resistance of
1.0 means complete immunity, while 0.0 means no immunity.

#### type ImmunityType

```go
type ImmunityType int
```

Core types EffectType represents a type of effect that can be applied to a game
entity in the RPG system. It is implemented as a string to allow for easy
extensibility and readable effect definitions.

Common effect types might include: - Damage - Healing - Status - Buff/Debuff

Related types: - DamageType - DispelType - ImmunityType

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
	Position   Position `yaml:"item_position,omitempty"`    // Current location in game world
}
```

Item represents a game item with its properties Contains all attributes that
define an item's behavior and characteristics Item represents a game item with
various attributes and properties. It is used to define objects that players can
interact with in the game world.

Fields:

    - ID (string): Unique identifier used to reference the item in the game
    - Name (string): Human-readable display name of the item
    - Type (string): Category classification (e.g. "weapon", "armor", "potion")
    - Damage (string): Optional damage specification for weapons (e.g. "1d6")
    - AC (int): Optional armor class value for defensive equipment
    - Weight (int): Weight of the item in game units
    - Value (int): Worth of the item in game currency
    - Properties ([]string): Optional list of special effects or attributes
    - Position (Position): Optional current location in the game world

The Item struct is serializable to/from YAML format using the specified tags.
Related types:

    - Position: Represents location coordinates in the game world

#### func (*Item) FromJSON

```go
func (i *Item) FromJSON(data []byte) error
```
FromJSON implements GameObject. FromJSON deserializes JSON data into an Item
struct.

Parameters:

    - data []byte: Raw JSON bytes to deserialize

Returns:

    - error: Returns an error if JSON unmarshaling fails

Related:

    - Item.ToJSON() for the inverse serialization operation

#### func (*Item) GetDescription

```go
func (i *Item) GetDescription() string
```
GetDescription implements GameObject. GetDescription returns a formatted string
representation of the item combining its Name and Type properties.

Returns a string in the format "Name (Type)"

Related types: - Item struct

#### func (*Item) GetHealth

```go
func (i *Item) GetHealth() int
```
GetHealth implements GameObject. GetHealth returns the health value of an Item.
Since items don't inherently have health in this implementation, it always
returns 0. This method satisfies an interface but has no practical effect for
basic Item objects. Returns:

    - int: Always returns 0 for base items

Related types:

    - Item struct

#### func (*Item) GetID

```go
func (i *Item) GetID() string
```
GetID implements GameObject. GetID returns the unique identifier string for this
Item. This method provides access to the private ID field. Returns a string
representing the item's unique identifier. Related: Item struct

#### func (*Item) GetName

```go
func (i *Item) GetName() string
```
GetName implements GameObject. GetName returns the name of the item

Returns:

    - string: The name property of the Item struct

#### func (*Item) GetPosition

```go
func (i *Item) GetPosition() Position
```
GetPosition implements GameObject. GetPosition returns the current position of
this item in the game world. If the item's position has not been explicitly set,
returns an empty Position struct. Returns:

    - Position: The x,y coordinates of the item

Related types:

    - Position struct

#### func (*Item) GetTags

```go
func (i *Item) GetTags() []string
```
GetTags implements GameObject. GetTags returns the Properties field of an Item,
which contains string tags/attributes associated with this item. The returned
slice can be empty if no properties are set.

Returns:

    - []string: A slice of strings representing the item's properties/tags

Related:

    - Item struct
    - Properties field

#### func (*Item) IsActive

```go
func (i *Item) IsActive() bool
```
IsActive implements GameObject. IsActive returns true to indicate if the item is
currently active in the game This method is used to indicate that items are
always usable / active by default

Returns:

    - bool: Always returns true as items are considered active by default

#### func (*Item) IsObstacle

```go
func (i *Item) IsObstacle() bool
```
IsObstacle implements GameObject. IsObstacle checks if this item blocks movement
in the game world. This is used by the movement system to determine if a
character can pass through the item. Returns false as items by default do not
obstruct movement. Related: Entity.IsObstacle(), Character.CanMoveTo()

#### func (*Item) SetHealth

```go
func (i *Item) SetHealth(health int)
```
SetHealth implements GameObject. SetHealth is a placeholder method that takes a
health value but performs no operation, as items in this game do not have health
attributes.

Parameters:

    - health: integer value representing health points (not used)

This is a no-op method maintained for compatibility with interfaces or future
use.

#### func (*Item) SetPosition

```go
func (i *Item) SetPosition(pos Position) error
```
SetPosition implements GameObject. SetPosition is a stub method that satisfies
the Entity interface but does not track position for Items. Currently items
don't maintain position state - this may change in future implementations.

Parameters:

    - pos: Position - The position to set (unused)

Returns:

    - error - Always returns nil since position is not tracked

Related types:

    - Position struct
    - Entity interface

#### func (*Item) ToJSON

```go
func (i *Item) ToJSON() ([]byte, error)
```
ToJSON implements GameObject. ToJSON serializes the Item struct into a JSON byte
array.

Returns:

    - []byte: The JSON representation of the Item
    - error: Error if marshaling fails

This method uses the standard encoding/json package for marshaling. Related
types: Item struct

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

Level represents a game level/map with its dimensions, layout and properties. A
level contains a 2D grid of Tiles and can be loaded from YAML configuration.

Fields:

    - ID: Unique string identifier for the level
    - Name: Human readable display name for the level
    - Width: Level width in number of tiles (must be > 0)
    - Height: Level height in number of tiles (must be > 0)
    - Tiles: 2D slice containing the level's tile grid, dimensions must match Width x Height
    - Properties: Map of custom level attributes for game-specific data

Related types:

    - Tile: Individual map tile type used in the Tiles grid

Usage:

    level := &Level{
      ID: "level1",
      Name: "Tutorial Level",
      Width: 10,
      Height: 10,
      Tiles: make([][]Tile, height),
      Properties: make(map[string]interface{}),
    }

#### type LootEntry

```go
type LootEntry struct {
	ItemID      string  `yaml:"loot_item_id"`      // Item identifier
	Chance      float64 `yaml:"loot_chance"`       // Drop probability
	MinQuantity int     `yaml:"loot_min_quantity"` // Minimum amount
	MaxQuantity int     `yaml:"loot_max_quantity"` // Maximum amount
}
```

LootEntry represents a single item drop configuration in the game's loot system.
It defines the probability and quantity range for a specific item that can be
obtained.

Fields:

    - ItemID: Unique identifier string for the item that can be dropped
    - Chance: Float value between 0.0 and 1.0 representing drop probability percentage
    - MinQuantity: Minimum number of items that can drop (must be >= 0)
    - MaxQuantity: Maximum number of items that can drop (must be >= MinQuantity)

Related types:

    - Item - The actual item definition this entry references
    - LootTable - Collection of LootEntry that defines all possible drops

#### type ModOpType

```go
type ModOpType string
```

ModOpType represents the type of modification operation that can be applied to
game attributes. It is implemented as a string type to allow for extensible
operation types while maintaining type safety through constant definitions.

```go
const (
	ModAdd      ModOpType = "add"      // Adds the modifier value to the base stat
	ModMultiply ModOpType = "multiply" // Multiplies the base stat by the modifier value
	ModSet      ModOpType = "set"      // Sets the stat directly to the modifier value
)
```
ModOpType constants define supported mathematical operations for modifying
stats/attributes. These are used by the Modifier type to specify how a stat
value should be changed.

Supported operations: - ModAdd: Adds the modifier value to the base stat -
ModMultiply: Multiplies the base stat by the modifier value - ModSet: Sets the
stat directly to the modifier value, ignoring the base value

Related types: - Modifier: Uses these operations to define stat modifications -
Effect: Contains Modifiers that use these operations

#### type Modifier

```go
type Modifier struct {
	Stat      string    `yaml:"mod_stat"`
	Value     float64   `yaml:"mod_value"`
	Operation ModOpType `yaml:"mod_operation"`
}
```

Modifier represents a modification to a game statistic or attribute. It defines
how a specific stat should be modified through a mathematical operation.

Fields:

    - Stat: The name/identifier of the stat being modified
    - Value: The numeric value to apply in the modification
    - Operation: The type of mathematical operation to perform (e.g. add, multiply)

Related types:

    - ModOpType: Enum defining valid modification operations

Usage example:

    mod := Modifier{
      Stat: "health",
      Value: 10,
      Operation: ModAdd,
    }

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

NPC represents a non-player character in the game world Extends the base
Character type with AI behaviors and interaction capabilities

Fields:

    - Character: Embedded base character attributes (health, stats, inventory etc)
    - Behavior: AI behavior pattern ID determining how NPC acts (e.g. "guard", "merchant")
    - Faction: Group allegiance affecting NPC relationships and interactions
    - Dialog: Available conversation options when player interacts with NPC
    - LootTable: Items that may be dropped when NPC dies

Related types:

    - Character: Base type providing core character functionality
    - DialogEntry: Defines conversation nodes and options
    - LootEntry: Defines droppable items and probabilities

Usage:

    npc := &NPC{
      Character: Character{Name: "Guard"},
      Behavior: "patrol",
      Faction: "town_guard",
      Dialog: []DialogEntry{...},
      LootTable: []LootEntry{...},
    }

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
attributes and mechanics specific to player characters Player represents a
playable character in the game with additional attributes beyond the base
Character type. It tracks progression elements like level, experience, quests
and learned spells.

The Player struct embeds the Character type to inherit basic attributes while
adding RPG-specific fields for character advancement and gameplay mechanics.

Fields:

    - Character: Base character attributes (embedded)
    - Class: The character's chosen class that determines available abilities
    - Level: Current experience level of the player (1 or greater)
    - Experience: Total experience points accumulated
    - QuestLog: Slice of active and completed quests
    - KnownSpells: Slice of spells the player has learned and can cast

Related types:

    - Character: Base character attributes
    - CharacterClass: Available character classes
    - Quest: Quest structure
    - Spell: Spell structure

#### func (*Player) AddExperience

```go
func (p *Player) AddExperience(exp int) error
```
AddExperience safely adds experience points and handles level ups AddExperience
adds the specified amount of experience points to the player and handles
leveling up. It is thread-safe through mutex locking.

Parameters:

    - exp: Amount of experience points to add (must be non-negative)

Returns:

    - error: Returns nil on success, error if exp is negative or if levelUp fails

Errors:

    - Returns error if exp is negative
    - Returns error from levelUp if leveling up fails

Related:

    - calculateLevel(): Used to determine if player should level up
    - levelUp(): Called when experience gain triggers a level increase

#### func (*Player) GetStats

```go
func (p *Player) GetStats() *Stats
```
Add this method to Player GetStats returns a copy of the player's current stats
converted to float64 values. It creates a new Stats struct containing the
player's health, max health, strength, dexterity and intelligence values.

Returns:

    - *Stats: A pointer to a new Stats struct containing the converted stat values

Related types:

    - Stats struct

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

PlayerProgressData represents the current progress and achievements of a player
in the game. It keeps track of various metrics like level, experience points,
and accomplishments.

Fields:

    - CurrentLevel: The player's current level in the game (must be >= 1)
    - ExperiencePoints: Total accumulated experience points
    - NextLevelThreshold: Experience points required to advance to next level
    - CompletedQuests: Number of quests the player has finished
    - SpellsLearned: Number of spells the player has mastered

Related types:

    - Use with Player struct to track overall player state
    - Experience points calculation handled by LevelingSystem

#### type Position

```go
type Position struct {
	X      int       `yaml:"position_x"`      // X coordinate on the map grid
	Y      int       `yaml:"position_y"`      // Y coordinate on the map grid
	Level  int       `yaml:"position_level"`  // Current dungeon/map level
	Facing Direction `yaml:"position_facing"` // Direction the entity is facing
}
```

Position represents the location and orientation of an entity in the game world.
It tracks both the 2D grid coordinates and vertical level for 3D positioning, as
well as which direction the entity is facing.

Fields:

    - X: Horizontal position on the map grid (integer)
    - Y: Vertical position on the map grid (integer)
    - Level: Current depth/floor number in the dungeon (integer)
    - Facing: Direction the entity is oriented (Direction enum)

Related types:

    - Direction: Used for the Facing field to indicate orientation

The Position struct uses YAML tags for serialization/deserialization

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

Quest represents a game quest with its properties and progress tracking. A quest
consists of a unique identifier, title, description, current status, objectives
that need to be completed, and rewards granted upon completion.

Fields:

    - ID: Unique string identifier for the quest
    - Title: Display name shown to the player
    - Description: Detailed explanation of the quest's story and goals
    - Status: Current state of the quest (see QuestStatus type)
    - Objectives: Slice of QuestObjective containing individual goals
    - Rewards: Slice of QuestReward given when quest is complete

Related types:

    - QuestStatus: Enum defining possible quest states
    - QuestObjective: Individual goals that must be completed
    - QuestReward: Items/experience granted on completion

#### type QuestObjective

```go
type QuestObjective struct {
	Description string `yaml:"objective_description"` // What needs to be done
	Progress    int    `yaml:"objective_progress"`    // Current completion amount
	Required    int    `yaml:"objective_required"`    // Amount needed for completion
	Completed   bool   `yaml:"objective_completed"`   // Whether objective is done
}
```

QuestObjective represents a specific task or goal within a quest that needs to
be completed. It tracks the progress towards completion and maintains the
completion status.

Fields:

    - Description: String describing what needs to be accomplished
    - Progress: Current amount of progress made towards completion (must be >= 0)
    - Required: Total amount needed to complete the objective (must be > 0)
    - Completed: Boolean flag indicating if the objective is finished

The Progress field should never exceed Required. When Progress equals or exceeds
Required, Completed should be set to true.

Related types:

    - Quest (parent type containing objectives)

#### type QuestProgress

```go
type QuestProgress struct {
	QuestID            string `yaml:"progress_quest_id"`        // Associated quest ID
	ObjectivesComplete int    `yaml:"progress_objectives_done"` // Number of completed objectives
	TimeSpent          int    `yaml:"progress_time_spent"`      // Time spent on quest
	Attempts           int    `yaml:"progress_attempts"`        // Number of attempts
}
```

QuestProgress tracks the player's progression status for a specific quest. It
maintains metrics like completion status, time investment and retry attempts.

Fields:

    - QuestID: Unique identifier string for the associated quest
    - ObjectivesComplete: Number of objectives completed in the quest (non-negative integer)
    - TimeSpent: Total time spent on quest in seconds (non-negative integer)
    - Attempts: Number of times player has attempted the quest (non-negative integer)

The struct is serializable via YAML for persistence. Related types:

    - Quest (for quest definition details)
    - QuestObjective (for individual objective tracking)

#### type QuestReward

```go
type QuestReward struct {
	Type   string `yaml:"reward_type"`    // Type of reward (gold, item, exp)
	Value  int    `yaml:"reward_value"`   // Quantity or amount of reward
	ItemID string `yaml:"reward_item_id"` // Reference to reward item if applicable
}
```

QuestReward represents a reward that can be awarded to a player for completing a
quest. It supports different types of rewards like gold, items, or experience
points.

Fields:

    - Type: The type of the reward, must be one of: "gold", "item", "exp"
    - Value: The quantity of the reward to give (amount of gold/exp, or number of items)
    - ItemID: Optional reference ID for item rewards, required only when Type is "item"

The reward is typically processed by the reward system which handles validation
and distribution to players. See RewardSystem.ProcessReward() for implementation
details.

#### type QuestStatus

```go
type QuestStatus int
```

QuestStatus represents the current state of a quest in the game. It is
implemented as an integer enumeration to track quest progression.

QuestStatus values indicate whether a quest is: - Not started/available - In
progress/active - Completed/finished - Failed/abandoned

Related types: - Quest struct: Contains the QuestStatus field - QuestLog:
Manages multiple quests and their statuses

```go
const (
	QuestNotStarted QuestStatus = iota
	QuestActive
	QuestCompleted
	QuestFailed
)
```
QuestNotStarted indicates that a quest has not yet been started by the player.
This is the initial state of any quest when first created or discovered.
Related: QuestActive, QuestCompleted, QuestFailed.

#### type RGB

```go
type RGB struct {
	R uint8 `yaml:"color_red"`   // Red component
	G uint8 `yaml:"color_green"` // Green component
	B uint8 `yaml:"color_blue"`  // Blue component
}
```

RGB represents a color in RGB format Each component ranges from 0-255 RGB
represents a color in the RGB color space with 8-bit components. Each component
(R,G,B) ranges from 0-255.

The struct is designed to be YAML serializable with custom field tags.

This is used throughout the game engine for defining colors of tiles, sprites
and other visual elements.

Related types:

    - Tile - Uses RGB for foreground/background colors

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

Spell represents a magical ability that can be cast in the game. It contains all
the necessary information about a spell's properties and effects.

Fields:

    - ID: Unique string identifier for the spell
    - Name: Display name shown to players
    - Level: Required caster level (must be >= 0)
    - School: Magic school classification (e.g. Abjuration, Evocation)
    - Range: Distance in game units the spell can reach (must be >= 0)
    - Duration: Number of game turns the spell effects last (must be >= 0)
    - Components: Required components needed to cast the spell
    - Description: Detailed text describing the spell's effects and usage

Related types:

    - SpellSchool: Enum defining valid magic schools
    - SpellComponent: Struct defining spell component requirements

#### type SpellComponent

```go
type SpellComponent int
```

SpellComponent represents a component of a spell in the game. It is implemented
as an integer type that can be used to classify different aspects or parts of a
spell, such as verbal, somatic, or material components.

Related types:

    - Spell (not shown in provided code)

```go
const (
	ComponentVerbal SpellComponent = iota
	ComponentSomatic
	ComponentMaterial
)
```
ComponentVerbal represents the verbal component required for casting spells. It
indicates that the spell requires specific words or phrases to be spoken to be
successfully cast. This is one of the fundamental spell components alongside
Somatic and Material components.

Related constants: - ComponentSomatic - ComponentMaterial

#### type SpellSchool

```go
type SpellSchool int
```

SpellSchool represents the different schools of magic available in the game
SpellSchool represents the school/category of magic that a spell belongs to. It
is implemented as an integer type for efficient storage and comparison. The
specific values are defined as constants representing different magical
disciplines like Abjuration, Conjuration, Evocation etc.

Related types: - Spell struct - Contains SpellSchool as one of its properties -
SpellEffect interface - Implemented by specific spell effects

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
SchoolAbjuration represents the school of Abjuration magic in the game world.
Abjuration spells are protective in nature, creating barriers, negating harmful
effects, or banishing creatures to other planes of existence.

This is one of the eight classical schools of magic defined in the game system.

Related constants: - SchoolConjuration - SchoolDivination - SchoolEnchantment -
SchoolEvocation - SchoolIllusion - SchoolNecromancy - SchoolTransmutation

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

Stats represents an entity's modifiable attributes Stats represents a
character's base and derived statistics in the game. It contains both primary
attributes that can be directly modified and secondary (calculated) attributes
that are derived from the primary ones.

Primary attributes:

    - Health: Current health points
    - Mana: Current mana points
    - Strength: Physical power and carrying capacity
    - Dexterity: Agility and precision
    - Intelligence: Mental capability and magical aptitude

Calculated attributes:

    - MaxHealth: Maximum possible health points
    - MaxMana: Maximum possible mana points
    - Defense: Damage reduction capability
    - Speed: Movement and action speed

The Stats struct is used throughout the game systems including: - Combat
calculations - Character progression - Status effect application - Equipment
bonuses

#### func  NewDefaultStats

```go
func NewDefaultStats() *Stats
```
NewDefaultStats creates and returns a new Stats structure initialized with
default values. It sets baseline stats that are commonly used as a starting
point for new game entities.

Returns:

    - *Stats: A pointer to a new Stats structure with the following default values:
      Health: 100, Mana: 100, Strength: 10, Dexterity: 10, Intelligence: 10,
      MaxHealth: 100, MaxMana: 100, Defense: 10, Speed: 10

Related types:

    - Stats struct: The base structure containing all stat fields

#### func (*Stats) Clone

```go
func (s *Stats) Clone() *Stats
```
Clone creates and returns a deep copy of a Stats object Clone duplicates all
stat values into a new Stats instance.

Returns:

    - *Stats: A new Stats instance with identical values to the original

Notable behavior: - Creates a completely independent copy of the Stats object -
All fields are copied by value since they are primitive types

Related types: - Stats struct: The base structure containing all stat fields -
NewDefaultStats(): Factory method for creating Stats objects

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

Tile represents a single cell in the game map. It encapsulates all properties
that define a tile's behavior, appearance, and interaction capabilities within
the game world.

Related types: - TileType: Defines the base classification of the tile - RGB:
Defines the color properties

Fields: - Type: Base classification of the tile (floor, wall, etc.) - Walkable:
Determines if entities can traverse this tile - Transparent: Controls if light
can pass through the tile - Properties: Custom key-value store for additional
tile attributes - Sprite: Identifier for the tile's visual representation -
Color: Base RGB color tint for rendering - BlocksSight: Specifically controls
line of sight behavior - Dangerous: Indicates if the tile can cause damage -
DamageType: Classification of damage (e.g., "fire", "poison") - Damage: Integer
amount of damage dealt per turn if dangerous

Note: Properties map allows for dynamic extension of tile attributes without
modifying the core structure.

#### func  NewFloorTile

```go
func NewFloorTile() Tile
```
Common tile factory functions NewFloorTile creates and returns a new floor tile
with default properties. The floor tile is walkable and transparent with a light
gray color (RGB: 200,200,200). Returns a Tile struct configured as a basic floor
tile with: - Type: TileFloor - Walkable: true - Transparent: true - Empty
properties map - Light gray color

Related types: - Tile struct - TileFloor constant

#### func  NewWallTile

```go
func NewWallTile() Tile
```
NewWallTile creates and returns a new wall tile with default properties. It
initializes an impassable, opaque wall with gray coloring that blocks line of
sight.

Returns:

    - Tile: A new wall tile instance with the following default properties:
    - Type: TileWall
    - Walkable: false (cannot be walked through)
    - Transparent: false (blocks vision)
    - Properties: empty map for custom properties
    - Sprite: empty string (no sprite assigned)
    - Color: gray RGB(128,128,128)
    - BlocksSight: true (blocks line of sight)
    - Dangerous: false (does not cause damage)
    - DamageType: empty string (no damage type)
    - Damage: 0 (no damage value)

Related types:

    - Tile
    - RGB
    - TileWall (constant)

#### type TileType

```go
type TileType int
```

TileType represents the type of a tile in the game world. It is implemented as
an integer enum to efficiently store and compare different tile types.

```go
const (
	TileFloor  TileType = iota // Basic floor tile that can be walked on
	TileWall                   // Solid wall that blocks movement and sight
	TileDoor                   // Door that can be opened/closed
	TileWater                  // Water tile that may affect movement
	TileLava                   // Dangerous lava tile that causes damage
	TilePit                    // Pit that entities may fall into
	TileStairs                 // Stairs for level transitions
)
```
TileType constants represent different types of tiles in the game world. Each
constant is assigned a unique integer value through iota.

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
