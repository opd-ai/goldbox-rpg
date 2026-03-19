// Package wasmui provides the Ebitengine/WASM-based game UI client types.
// This file contains RPC request/response types for server communication.
package wasmui

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
	TargetType  string   `json:"target_type,omitempty"` // "self", "single", "area", "cone"
	AreaRadius  int      `json:"area_radius,omitempty"` // Radius in tiles for area spells
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
