// Package wasmui provides the Ebitengine/WASM-based game UI client types.
package wasmui

import "image/color"

// Position represents a 2D coordinate.
// Uses uppercase field names to match server's default JSON encoding.
type Position struct {
	X      int `json:"X"`
	Y      int `json:"Y"`
	Level  int `json:"Level,omitempty"`
	Facing int `json:"Facing,omitempty"`
}

// PlayerAttributes represents the six core D&D-style attributes.
// Field names match the server's public data format.
type PlayerAttributes struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// PlayerState represents the current state of a player character.
// Field names match the server's session PublicData format.
type PlayerState struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Position   Position         `json:"position"`
	HP         int              `json:"hp"`
	MaxHP      int              `json:"max_hp"`
	AP         int              `json:"ap"`
	MaxAP      int              `json:"max_ap"`
	Level      int              `json:"level"`
	Experience int              `json:"experience"`
	Class      string           `json:"class"`
	Attributes PlayerAttributes `json:"attributes"`
	Appearance *Appearance      `json:"appearance,omitempty"`
	Effects    []EffectData     `json:"effects,omitempty"`
	Equipment  []EquippedItem   `json:"equipment,omitempty"`
}

// Appearance holds cosmetic and biographical character properties.
// Mirrors game.Appearance for the WASM UI layer.
type Appearance struct {
	SkinTone            int    `json:"skin_tone,omitempty"`
	HairStyle           string `json:"hair_style,omitempty"`
	HairColor           string `json:"hair_color,omitempty"`
	BodyType            int    `json:"body_type,omitempty"`
	GenderExpression    string `json:"gender_expression,omitempty"`
	Pronouns            string `json:"pronouns,omitempty"`
	RomanticOrientation string `json:"romantic_orientation,omitempty"`
}

// EquippedItem represents an item in an equipment slot.
type EquippedItem struct {
	Slot   string `json:"slot"`
	ItemID string `json:"item_id"`
	Name   string `json:"name"`
}

// EffectData represents an active effect on a character.
type EffectData struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Duration      int    `json:"duration"`
	Remaining     int    `json:"remaining"`
	Magnitude     int    `json:"magnitude"`
	Source        string `json:"source"`
	EffectKeyword string `json:"effect_keyword,omitempty"`
}

// ItemData represents an item in the player's inventory.
type ItemData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Slot        string `json:"slot,omitempty"`
	Damage      string `json:"damage,omitempty"`
	Defense     int    `json:"defense,omitempty"`
	Weight      int    `json:"weight,omitempty"`
	Value       int    `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
	Consumable  bool   `json:"consumable,omitempty"`
	Proficiency string `json:"proficiency,omitempty"`
	Equipped    bool   `json:"equipped,omitempty"`
}

// SpellData represents a spell known by the player.
type SpellData struct {
	ID          string   `json:"spell_id"`
	Name        string   `json:"name"`
	Level       int      `json:"spell_level"`
	School      int      `json:"spell_school"`
	SchoolName  string   `json:"school_name,omitempty"`
	Range       string   `json:"range,omitempty"`
	Duration    string   `json:"duration,omitempty"`
	Components  string   `json:"components,omitempty"`
	Description string   `json:"description,omitempty"`
	DamageType  string   `json:"damage_type,omitempty"`
	DamageDice  string   `json:"damage_dice,omitempty"`
	HealingDice string   `json:"healing_dice,omitempty"`
	AreaEffect  bool     `json:"area_effect,omitempty"`
	SaveType    string   `json:"save_type,omitempty"`
	Keywords    []string `json:"effect_keywords,omitempty"`
}

// QuestData represents a quest in the quest log.
type QuestData struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      string           `json:"status"`
	Objectives  []QuestObjective `json:"objectives,omitempty"`
	Rewards     []QuestReward    `json:"rewards,omitempty"`
}

// QuestObjective represents a single objective in a quest.
type QuestObjective struct {
	Description string `json:"description"`
	Progress    int    `json:"progress"`
	Required    int    `json:"required"`
	Completed   bool   `json:"completed"`
}

// QuestReward represents a reward for completing a quest.
type QuestReward struct {
	Type   string `json:"type"`
	Value  int    `json:"value"`
	ItemID string `json:"item_id,omitempty"`
	Name   string `json:"name,omitempty"`
}

// GuildData represents guild information.
type GuildData struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Level      int           `json:"level"`
	Experience int           `json:"experience"`
	NextLevel  int           `json:"next_level"`
	Treasury   int           `json:"treasury"`
	LeaderID   string        `json:"leader_id"`
	LeaderName string        `json:"leader_name"`
	MemberCnt  int           `json:"member_count"`
	Members    []GuildMember `json:"members,omitempty"`
	Perks      []GuildPerk   `json:"perks,omitempty"`
	Ranks      []string      `json:"ranks,omitempty"`
}

// GuildMember represents a member of a guild.
type GuildMember struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Rank         int    `json:"rank"`
	RankName     string `json:"rank_name"`
	Contribution int    `json:"contribution"`
}

// GuildPerk represents a guild perk.
type GuildPerk struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	LevelReq    int    `json:"level_req"`
}

// FactionRelation represents a diplomatic relation with a faction.
type FactionRelation struct {
	FactionID      string `json:"faction_id"`
	FactionName    string `json:"faction_name"`
	State          string `json:"state"`
	Opinion        int    `json:"opinion"`
	Trust          int    `json:"trust"`
	TradeTreaty    bool   `json:"trade_treaty"`
	MilitaryAccess bool   `json:"military_access"`
	DefensivePact  bool   `json:"defensive_pact"`
}

// CombatState represents the current combat state.
type CombatState struct {
	Active      bool              `json:"active"`
	CurrentTurn string            `json:"currentTurn"`
	Initiative  []InitiativeEntry `json:"initiative"`
	Round       int               `json:"round"`
	InCombat    bool              `json:"inCombat"`
}

// InitiativeEntry represents an entry in the combat initiative order.
type InitiativeEntry struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Initiative int    `json:"initiative"`
	IsPlayer   bool   `json:"isPlayer"`
	HP         int    `json:"hp,omitempty"`
	MaxHP      int    `json:"max_hp,omitempty"`
}

// GameStateData represents the complete game state from server.
type GameStateData struct {
	Player  *PlayerState `json:"player"`
	Combat  *CombatState `json:"combat"`
	World   interface{}  `json:"world"`
	Session *SessionData `json:"session"`
}

// SessionData represents the current session information.
type SessionData struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// LogMessage represents a message in the combat/game log.
type LogMessage struct {
	Text      string
	Type      MessageType
	Timestamp int64
}

// MessageType defines the type of log message for styling.
type MessageType int

const (
	MessageInfo MessageType = iota
	MessageWarning
	MessageError
	MessageCombat
	MessageSystem
)

// Color returns the appropriate color for the message type.
func (mt MessageType) Color() color.RGBA {
	switch mt {
	case MessageWarning:
		return color.RGBA{R: 255, G: 200, B: 0, A: 255}
	case MessageError:
		return color.RGBA{R: 255, G: 100, B: 100, A: 255}
	case MessageCombat:
		return color.RGBA{R: 200, G: 150, B: 255, A: 255}
	case MessageSystem:
		return color.RGBA{R: 150, G: 200, B: 255, A: 255}
	default:
		return color.RGBA{R: 220, G: 220, B: 220, A: 255}
	}
}

// UIMode represents the current UI mode.
type UIMode int

const (
	ModeNormal UIMode = iota
	ModeCombat
	ModeInventory
	ModeSpellcasting
	ModeAdventureSelect
	ModeCharacterCreation
)

// ScreenState tracks the current screen within ModeNormal.
type ScreenState int

const (
	ScreenSplash ScreenState = iota
	ScreenMainMenu
	ScreenExploration
	ScreenVictory
	ScreenDefeat
)

// OverlayState tracks which overlay panels are open.
// These are drawn on top of the current mode per §2 note.
type OverlayState struct {
	ShowQuestLog   bool
	ShowGuildPanel bool
	ShowSettings   bool
}

// CharCreationStep tracks the current step within character creation.
type CharCreationStep int

const (
	CharStepName CharCreationStep = iota
	CharStepClass
	CharStepAttributes
	CharStepReview
)

// AttributeMethod determines how attributes are generated.
type AttributeMethod int

const (
	AttrMethodRoll AttributeMethod = iota
	AttrMethodStandard
	AttrMethodPointBuy
	AttrMethodCustom
)

// String returns the display name of the attribute method.
func (m AttributeMethod) String() string {
	switch m {
	case AttrMethodRoll:
		return "Roll"
	case AttrMethodStandard:
		return "Standard"
	case AttrMethodPointBuy:
		return "Point Buy"
	case AttrMethodCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// CharCreationState holds temporary state for the character creation flow.
type CharCreationState struct {
	Step           CharCreationStep
	Name           string
	SelectedClass  int // index into ClassInfoList
	AttrMethod     AttributeMethod
	Attributes     PlayerAttributes
	SelectedAttr   int // index 0-5 for attribute selection in point-buy / adjustment
	PointBuyPoints int
}

// GetAttr returns the attribute value at index i (0=STR..5=CHA).
func (c *CharCreationState) GetAttr(i int) int {
	switch i {
	case 0:
		return c.Attributes.Strength
	case 1:
		return c.Attributes.Dexterity
	case 2:
		return c.Attributes.Constitution
	case 3:
		return c.Attributes.Intelligence
	case 4:
		return c.Attributes.Wisdom
	case 5:
		return c.Attributes.Charisma
	}
	return 0
}

// SetAttr sets the attribute value at index i (0=STR..5=CHA).
func (c *CharCreationState) SetAttr(i, val int) {
	switch i {
	case 0:
		c.Attributes.Strength = val
	case 1:
		c.Attributes.Dexterity = val
	case 2:
		c.Attributes.Constitution = val
	case 3:
		c.Attributes.Intelligence = val
	case 4:
		c.Attributes.Wisdom = val
	case 5:
		c.Attributes.Charisma = val
	}
}

// SetStandardArray loads the standard array into the attributes.
func (c *CharCreationState) SetStandardArray() {
	c.Attributes.Strength = StandardArray[0]
	c.Attributes.Dexterity = StandardArray[1]
	c.Attributes.Constitution = StandardArray[2]
	c.Attributes.Intelligence = StandardArray[3]
	c.Attributes.Wisdom = StandardArray[4]
	c.Attributes.Charisma = StandardArray[5]
}

// ResetAttributes sets all attribute values to the given value.
func (c *CharCreationState) ResetAttributes(val int) {
	c.Attributes.Strength = val
	c.Attributes.Dexterity = val
	c.Attributes.Constitution = val
	c.Attributes.Intelligence = val
	c.Attributes.Wisdom = val
	c.Attributes.Charisma = val
}

// ClassInfo describes a character class for the creation screen.
type ClassInfo struct {
	Name                string
	Description         string
	HitDice             string
	WeaponProficiencies []string
	ArmorProficiencies  []string
	ShieldProficiency   bool
	Restrictions        string
}

// ClassInfoList is the ordered list of selectable classes per §4.
var ClassInfoList = []ClassInfo{
	{Name: "Fighter", Description: "Melee combat specialist with high HP", HitDice: "d10", WeaponProficiencies: []string{"sword", "axe", "mace", "bow", "dagger", "spear", "wand", "staff"}, ArmorProficiencies: []string{"light", "medium", "heavy"}, ShieldProficiency: true},
	{Name: "Mage", Description: "Arcane spellcaster with powerful magic", HitDice: "d4", WeaponProficiencies: []string{"staff", "dagger", "wand"}, ArmorProficiencies: []string{}, ShieldProficiency: false, Restrictions: "No armor, no shields"},
	{Name: "Cleric", Description: "Divine healer and support caster", HitDice: "d8", WeaponProficiencies: []string{"mace", "staff", "dagger"}, ArmorProficiencies: []string{"light", "medium", "heavy"}, ShieldProficiency: true, Restrictions: "No edged weapons"},
	{Name: "Thief", Description: "Stealthy rogue with trap expertise", HitDice: "d6", WeaponProficiencies: []string{"dagger", "sword", "bow"}, ArmorProficiencies: []string{"light"}, ShieldProficiency: false},
	{Name: "Ranger", Description: "Wilderness warrior with tracking skills", HitDice: "d8", WeaponProficiencies: []string{"bow", "sword", "dagger", "spear"}, ArmorProficiencies: []string{"light", "medium"}, ShieldProficiency: true},
	{Name: "Paladin", Description: "Holy warrior combining arms and faith", HitDice: "d10", WeaponProficiencies: []string{"sword", "mace", "spear", "bow", "dagger"}, ArmorProficiencies: []string{"light", "medium", "heavy"}, ShieldProficiency: true},
}

// CombatAction represents the currently selected combat action.
type CombatAction int

const (
	CombatActionNone CombatAction = iota
	CombatActionMove
	CombatActionAttack
	CombatActionCast
	CombatActionItem
	CombatActionDefend
	CombatActionFlee
)

// String returns the display name of the combat action.
func (a CombatAction) String() string {
	switch a {
	case CombatActionNone:
		return "None"
	case CombatActionMove:
		return "Move"
	case CombatActionAttack:
		return "Attack"
	case CombatActionCast:
		return "Cast"
	case CombatActionItem:
		return "Item"
	case CombatActionDefend:
		return "Defend"
	case CombatActionFlee:
		return "Flee"
	default:
		return "Unknown"
	}
}

// VictoryData holds statistics displayed on the victory screen.
type VictoryData struct {
	AdventureTitle  string
	TimePlayed      string
	QuestsComplete  int
	QuestsTotal     int
	EnemiesDefeated int
	GoldEarned      int
	XPEarned        int
	LevelFrom       int
	LevelTo         int
}

// DefeatData holds information displayed on the defeat screen.
type DefeatData struct {
	LastLocation string
	CauseOfDeath string
}

// SpellSchoolName returns the display name for a spell school integer.
func SpellSchoolName(school int) string {
	names := []string{
		"Abjuration", "Conjuration", "Divination", "Enchantment",
		"Evocation", "Illusion", "Necromancy", "Transmutation",
	}
	if school >= 0 && school < len(names) {
		return names[school]
	}
	return "Unknown"
}

// EffectIcon returns the display icon for an effect type string.
func EffectIcon(effectType string) string {
	icons := map[string]string{
		"burning":          "Fire",
		"poison":           "Pois",
		"bleeding":         "Bld",
		"stun":             "Stun",
		"root":             "Root",
		"stat_boost":       "Bst+",
		"stat_penalty":     "Bst-",
		"haste":            "Hst",
		"slow":             "Slow",
		"regeneration":     "Rgen",
		"paralysis":        "Para",
		"heal_over_time":   "HoT",
		"damage_over_time": "DoT",
	}
	if icon, ok := icons[effectType]; ok {
		return icon
	}
	return "Eff"
}

// AttributeModifier calculates the D&D-style modifier for an attribute score.
func AttributeModifier(score int) int {
	return (score - 10) / 2
}

// StandardArray is the fixed attribute array per §4 Step 3.
var StandardArray = [6]int{15, 14, 13, 12, 10, 8}

// PointBuyCost returns the point cost for a given attribute score.
func PointBuyCost(score int) int {
	if score < 8 || score > 15 {
		return -1
	}
	costs := map[int]int{8: 0, 9: 1, 10: 2, 11: 3, 12: 4, 13: 5, 14: 7, 15: 9}
	return costs[score]
}

// PointBuyTotal is the total points available for point-buy.
const PointBuyTotal = 27
