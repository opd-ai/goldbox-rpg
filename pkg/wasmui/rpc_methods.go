//go:build js && wasm

// Package wasmui provides RPC wrapper methods for all gameplay JSON-RPC calls.
// Each method corresponds to a server-side handler defined in pkg/server/constants.go.
package wasmui

import "encoding/json"

// --- Character Actions (§13 methods 1-5) ---

// CastSpell sends a castSpell request.
func (c *RPCClient) CastSpell(spellID, targetID string, position *Position) (*CastSpellResult, error) {
	params := map[string]interface{}{
		"spell_id":  spellID,
		"target_id": targetID,
	}
	if position != nil {
		params["position"] = map[string]interface{}{"X": position.X, "Y": position.Y}
	}
	return rpcCall[CastSpellResult](c, "castSpell", params)
}

// UseItem sends a useItem request.
func (c *RPCClient) UseItem(itemID, targetID string) (*UseItemResult, error) {
	return rpcCall[UseItemResult](c, "useItem", map[string]interface{}{
		"item_id":   itemID,
		"target_id": targetID,
	})
}

// ApplyEffect sends an applyEffect request.
func (c *RPCClient) ApplyEffect(targetID, effectType string, duration, magnitude int) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "applyEffect", map[string]interface{}{
		"target_id":   targetID,
		"effect_type": effectType,
		"duration":    duration,
		"magnitude":   magnitude,
	})
}

// --- Combat Management (§13 methods 6-7) ---

// StartCombat sends a startCombat request.
func (c *RPCClient) StartCombat(enemyIDs []string) (*StartCombatResult, error) {
	return rpcCall[StartCombatResult](c, "startCombat", map[string]interface{}{
		"enemy_ids": enemyIDs,
	})
}

// --- Game State (§13 methods 8-11) ---

// LeaveGame sends a leaveGame request.
func (c *RPCClient) LeaveGame() (*GenericResult, error) {
	return rpcCall[GenericResult](c, "leaveGame", nil)
}

// CreateCharacter sends a createCharacter request.
func (c *RPCClient) CreateCharacter(name, class, method string, attributes *PlayerAttributes) (*CreateCharacterResult, error) {
	params := map[string]interface{}{
		"name":   name,
		"class":  class,
		"method": method,
	}
	if attributes != nil {
		params["attributes"] = map[string]interface{}{
			"strength":     attributes.Strength,
			"dexterity":    attributes.Dexterity,
			"constitution": attributes.Constitution,
			"intelligence": attributes.Intelligence,
			"wisdom":       attributes.Wisdom,
			"charisma":     attributes.Charisma,
		}
	}
	return rpcCall[CreateCharacterResult](c, "createCharacter", params)
}

// --- Equipment (§13 methods 12-14) ---

// EquipItem sends an equipItem request.
func (c *RPCClient) EquipItem(itemID, slot string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "equipItem", map[string]interface{}{
		"item_id": itemID,
		"slot":    slot,
	})
}

// UnequipItem sends an unequipItem request.
func (c *RPCClient) UnequipItem(slot string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "unequipItem", map[string]interface{}{
		"slot": slot,
	})
}

// GetEquipment sends a getEquipment request.
func (c *RPCClient) GetEquipment() (*GetEquipmentResult, error) {
	return rpcCall[GetEquipmentResult](c, "getEquipment", nil)
}

// --- Quest Management (§13 methods 15-22) ---

// StartQuest sends a startQuest request.
func (c *RPCClient) StartQuest(questID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "startQuest", map[string]interface{}{
		"quest_id": questID,
	})
}

// CompleteQuest sends a completeQuest request.
func (c *RPCClient) CompleteQuest(questID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "completeQuest", map[string]interface{}{
		"quest_id": questID,
	})
}

// UpdateObjective sends an updateObjective request.
func (c *RPCClient) UpdateObjective(questID string, objectiveIndex, progress int) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "updateObjective", map[string]interface{}{
		"quest_id":        questID,
		"objective_index": objectiveIndex,
		"progress":        progress,
	})
}

// FailQuest sends a failQuest request.
func (c *RPCClient) FailQuest(questID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "failQuest", map[string]interface{}{
		"quest_id": questID,
	})
}

// GetQuest sends a getQuest request for a single quest.
func (c *RPCClient) GetQuest(questID string) (*QuestData, error) {
	return rpcCall[QuestData](c, "getQuest", map[string]interface{}{
		"quest_id": questID,
	})
}

// GetActiveQuests sends a getActiveQuests request.
func (c *RPCClient) GetActiveQuests() (*QuestListResult, error) {
	return rpcCall[QuestListResult](c, "getActiveQuests", nil)
}

// GetCompletedQuests sends a getCompletedQuests request.
func (c *RPCClient) GetCompletedQuests() (*QuestListResult, error) {
	return rpcCall[QuestListResult](c, "getCompletedQuests", nil)
}

// GetQuestLog sends a getQuestLog request returning all quests.
func (c *RPCClient) GetQuestLog() (*QuestLogResult, error) {
	return rpcCall[QuestLogResult](c, "getQuestLog", nil)
}

// --- Spell System (§13 methods 23-27) ---

// GetSpell sends a getSpell request for a single spell.
func (c *RPCClient) GetSpell(spellID string) (*SpellData, error) {
	return rpcCall[SpellData](c, "getSpell", map[string]interface{}{
		"spell_id": spellID,
	})
}

// GetSpellsByLevel sends a getSpellsByLevel request.
func (c *RPCClient) GetSpellsByLevel(level int) (*SpellListResult, error) {
	return rpcCall[SpellListResult](c, "getSpellsByLevel", map[string]interface{}{
		"level": level,
	})
}

// GetSpellsBySchool sends a getSpellsBySchool request.
func (c *RPCClient) GetSpellsBySchool(school int) (*SpellListResult, error) {
	return rpcCall[SpellListResult](c, "getSpellsBySchool", map[string]interface{}{
		"school": school,
	})
}

// GetAllSpells sends a getAllSpells request.
func (c *RPCClient) GetAllSpells() (*SpellListResult, error) {
	return rpcCall[SpellListResult](c, "getAllSpells", nil)
}

// SearchSpells sends a searchSpells request.
func (c *RPCClient) SearchSpells(query string) (*SpellListResult, error) {
	return rpcCall[SpellListResult](c, "searchSpells", map[string]interface{}{
		"query": query,
	})
}

// --- Spatial Queries (§13 methods 28-30) ---

// GetObjectsInRange sends a getObjectsInRange request.
func (c *RPCClient) GetObjectsInRange(x, y, rangeVal int) (*SpatialResult, error) {
	return rpcCall[SpatialResult](c, "getObjectsInRange", map[string]interface{}{
		"x": x, "y": y, "range": rangeVal,
	})
}

// GetObjectsInRadius sends a getObjectsInRadius request.
func (c *RPCClient) GetObjectsInRadius(x, y, radius int) (*SpatialResult, error) {
	return rpcCall[SpatialResult](c, "getObjectsInRadius", map[string]interface{}{
		"x": x, "y": y, "radius": radius,
	})
}

// GetNearestObjects sends a getNearestObjects request.
func (c *RPCClient) GetNearestObjects(x, y, count int) (*SpatialResult, error) {
	return rpcCall[SpatialResult](c, "getNearestObjects", map[string]interface{}{
		"x": x, "y": y, "count": count,
	})
}

// --- Pathfinding (§13 method 31) ---

// FindPath sends a findPath request.
func (c *RPCClient) FindPath(fromX, fromY, toX, toY int) (*FindPathResult, error) {
	return rpcCall[FindPathResult](c, "findPath", map[string]interface{}{
		"from_x": fromX, "from_y": fromY,
		"to_x": toX, "to_y": toY,
	})
}

// --- Procedural Content Generation (§13 methods 32-38) ---

// GenerateContent sends a generateContent request.
func (c *RPCClient) GenerateContent(contentType string, seed int64) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "generateContent", map[string]interface{}{
		"type": contentType,
		"seed": seed,
	})
}

// RegenerateTerrain sends a regenerateTerrain request.
func (c *RPCClient) RegenerateTerrain(seed int64) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "regenerateTerrain", map[string]interface{}{
		"seed": seed,
	})
}

// GenerateItems sends a generateItems request.
func (c *RPCClient) GenerateItems(count int, seed int64) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "generateItems", map[string]interface{}{
		"count": count,
		"seed":  seed,
	})
}

// GenerateLevel sends a generateLevel request.
func (c *RPCClient) GenerateLevel(width, height int, seed int64) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "generateLevel", map[string]interface{}{
		"width": width, "height": height, "seed": seed,
	})
}

// GenerateQuest sends a generateQuest request.
func (c *RPCClient) GenerateQuest(seed int64) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "generateQuest", map[string]interface{}{
		"seed": seed,
	})
}

// GetPCGStats sends a getPCGStats request.
func (c *RPCClient) GetPCGStats() (*GenericResult, error) {
	return rpcCall[GenericResult](c, "getPCGStats", nil)
}

// ValidateContent sends a validateContent request.
func (c *RPCClient) ValidateContent(contentType, contentID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "validateContent", map[string]interface{}{
		"type": contentType,
		"id":   contentID,
	})
}

// --- Guild Management (§13 methods 39-50) ---

// CreateGuild sends a createGuild request.
func (c *RPCClient) CreateGuild(name string) (*GuildResult, error) {
	return rpcCall[GuildResult](c, "createGuild", map[string]interface{}{
		"name": name,
	})
}

// GetGuild sends a getGuild request.
func (c *RPCClient) GetGuild(guildID string) (*GuildData, error) {
	return rpcCall[GuildData](c, "getGuild", map[string]interface{}{
		"guild_id": guildID,
	})
}

// GetCharacterGuild sends a getCharacterGuild request.
func (c *RPCClient) GetCharacterGuild() (*GuildData, error) {
	return rpcCall[GuildData](c, "getCharacterGuild", nil)
}

// JoinGuild sends a joinGuild request.
func (c *RPCClient) JoinGuild(guildID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "joinGuild", map[string]interface{}{
		"guild_id": guildID,
	})
}

// LeaveGuild sends a leaveGuild request.
func (c *RPCClient) LeaveGuild() (*GenericResult, error) {
	return rpcCall[GenericResult](c, "leaveGuild", nil)
}

// KickGuildMember sends a kickGuildMember request.
func (c *RPCClient) KickGuildMember(memberID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "kickGuildMember", map[string]interface{}{
		"member_id": memberID,
	})
}

// PromoteGuildMember sends a promoteGuildMember request.
func (c *RPCClient) PromoteGuildMember(memberID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "promoteGuildMember", map[string]interface{}{
		"member_id": memberID,
	})
}

// DemoteGuildMember sends a demoteGuildMember request.
func (c *RPCClient) DemoteGuildMember(memberID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "demoteGuildMember", map[string]interface{}{
		"member_id": memberID,
	})
}

// GuildDeposit sends a guildDeposit request.
func (c *RPCClient) GuildDeposit(amount int) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "guildDeposit", map[string]interface{}{
		"amount": amount,
	})
}

// GuildWithdraw sends a guildWithdraw request.
func (c *RPCClient) GuildWithdraw(amount int) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "guildWithdraw", map[string]interface{}{
		"amount": amount,
	})
}

// ListGuilds sends a listGuilds request.
func (c *RPCClient) ListGuilds() (*GuildListResult, error) {
	return rpcCall[GuildListResult](c, "listGuilds", nil)
}

// TransferGuildLeader sends a transferGuildLeader request.
func (c *RPCClient) TransferGuildLeader(memberID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "transferGuildLeader", map[string]interface{}{
		"member_id": memberID,
	})
}

// --- Faction Diplomacy (§13 methods 51-60) ---

// GetFactionRelation sends a getFactionRelation request.
func (c *RPCClient) GetFactionRelation(factionID string) (*FactionRelation, error) {
	return rpcCall[FactionRelation](c, "getFactionRelation", map[string]interface{}{
		"faction_id": factionID,
	})
}

// GetFactionRelations sends a getFactionRelations request.
func (c *RPCClient) GetFactionRelations() (*FactionListResult, error) {
	return rpcCall[FactionListResult](c, "getFactionRelations", nil)
}

// DeclareWar sends a declareWar request.
func (c *RPCClient) DeclareWar(factionID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "declareWar", map[string]interface{}{
		"faction_id": factionID,
	})
}

// OfferPeace sends an offerPeace request.
func (c *RPCClient) OfferPeace(factionID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "offerPeace", map[string]interface{}{
		"faction_id": factionID,
	})
}

// AcceptPeace sends an acceptPeace request.
func (c *RPCClient) AcceptPeace(factionID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "acceptPeace", map[string]interface{}{
		"faction_id": factionID,
	})
}

// ProposeAlliance sends a proposeAlliance request.
func (c *RPCClient) ProposeAlliance(factionID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "proposeAlliance", map[string]interface{}{
		"faction_id": factionID,
	})
}

// AcceptAlliance sends an acceptAlliance request.
func (c *RPCClient) AcceptAlliance(factionID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "acceptAlliance", map[string]interface{}{
		"faction_id": factionID,
	})
}

// BreakAlliance sends a breakAlliance request.
func (c *RPCClient) BreakAlliance(factionID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "breakAlliance", map[string]interface{}{
		"faction_id": factionID,
	})
}

// SignTrade sends a signTrade request.
func (c *RPCClient) SignTrade(factionID string) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "signTrade", map[string]interface{}{
		"faction_id": factionID,
	})
}

// SendDiplomaticGift sends a sendDiplomaticGift request.
func (c *RPCClient) SendDiplomaticGift(factionID string, amount int) (*GenericResult, error) {
	return rpcCall[GenericResult](c, "sendDiplomaticGift", map[string]interface{}{
		"faction_id": factionID,
		"amount":     amount,
	})
}

// --- RPC Result types ---

// GenericResult represents a generic success/message response.
type GenericResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// CastSpellResult represents the result of a castSpell call.
type CastSpellResult struct {
	Success bool   `json:"success"`
	Damage  int    `json:"damage,omitempty"`
	Healing int    `json:"healing,omitempty"`
	Message string `json:"message"`
}

// UseItemResult represents the result of a useItem call.
type UseItemResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// StartCombatResult represents the result of a startCombat call.
type StartCombatResult struct {
	Success    bool              `json:"success"`
	Initiative []InitiativeEntry `json:"initiative,omitempty"`
	Message    string            `json:"message"`
}

// CreateCharacterResult represents the result of a createCharacter call.
type CreateCharacterResult struct {
	Success   bool         `json:"success"`
	SessionID string       `json:"session_id,omitempty"`
	Character *PlayerState `json:"character,omitempty"`
	Message   string       `json:"message,omitempty"`
}

// GetEquipmentResult represents the result of a getEquipment call.
type GetEquipmentResult struct {
	Success     bool           `json:"success"`
	Equipment   []EquippedItem `json:"equipment,omitempty"`
	Inventory   []ItemData     `json:"inventory,omitempty"`
	TotalWeight int            `json:"total_weight,omitempty"`
	WeightLimit int            `json:"weight_limit,omitempty"`
}

// QuestListResult represents a list of quests.
type QuestListResult struct {
	Success bool        `json:"success"`
	Quests  []QuestData `json:"quests,omitempty"`
	Count   int         `json:"count"`
}

// QuestLogResult represents the full quest log.
type QuestLogResult struct {
	Success         bool        `json:"success"`
	ActiveQuests    []QuestData `json:"active_quests,omitempty"`
	CompletedQuests []QuestData `json:"completed_quests,omitempty"`
	FailedQuests    []QuestData `json:"failed_quests,omitempty"`
}

// SpellListResult represents a list of spells.
type SpellListResult struct {
	Success bool        `json:"success"`
	Spells  []SpellData `json:"spells,omitempty"`
	Count   int         `json:"count"`
}

// SpatialResult represents the result of a spatial query.
type SpatialResult struct {
	Success bool          `json:"success"`
	Objects []interface{} `json:"objects,omitempty"`
	Count   int           `json:"count"`
}

// FindPathResult represents the result of a pathfinding query.
type FindPathResult struct {
	Success bool       `json:"success"`
	Path    []Position `json:"path,omitempty"`
	Cost    int        `json:"cost,omitempty"`
	Message string     `json:"message,omitempty"`
}

// GuildResult represents the result of a guild creation.
type GuildResult struct {
	Success bool       `json:"success"`
	Guild   *GuildData `json:"guild,omitempty"`
	Message string     `json:"message,omitempty"`
}

// GuildListResult represents a list of guilds.
type GuildListResult struct {
	Success bool        `json:"success"`
	Guilds  []GuildData `json:"guilds,omitempty"`
	Count   int         `json:"count"`
}

// FactionListResult represents a list of faction relations.
type FactionListResult struct {
	Success   bool              `json:"success"`
	Relations []FactionRelation `json:"relations,omitempty"`
	Count     int               `json:"count"`
}

// --- Helper: generic typed RPC call ---

// rpcCall sends an RPC request and unmarshals the result into type T.
func rpcCall[T any](c *RPCClient, method string, params map[string]interface{}) (*T, error) {
	result, err := c.Call(method, params)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
