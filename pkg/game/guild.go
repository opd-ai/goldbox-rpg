// Package game provides guild membership mechanics for player organizations.
package game

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Guild represents a player organization with membership, ranks, and perks.
type Guild struct {
	mu          sync.RWMutex            `yaml:"-"`
	ID          string                  `yaml:"guild_id"`
	Name        string                  `yaml:"guild_name"`
	Description string                  `yaml:"guild_description"`
	Founded     time.Time               `yaml:"guild_founded"`
	FactionID   string                  `yaml:"guild_faction_id"` // Associated faction
	LeaderID    string                  `yaml:"guild_leader_id"`  // Character ID of leader
	Members     map[string]*GuildMember `yaml:"guild_members"`    // Character ID -> membership
	Ranks       []*GuildRank            `yaml:"guild_ranks"`
	Treasury    int                     `yaml:"guild_treasury"` // Gold in guild bank
	Experience  int64                   `yaml:"guild_experience"`
	Level       int                     `yaml:"guild_level"`
	HallID      string                  `yaml:"guild_hall_id"` // Location/territory ID
	Perks       []GuildPerk             `yaml:"guild_perks"`
	Properties  map[string]interface{}  `yaml:"guild_properties"`
}

// GuildMember represents a character's membership in a guild.
type GuildMember struct {
	CharacterID  string    `yaml:"member_character_id"`
	JoinDate     time.Time `yaml:"member_join_date"`
	RankIndex    int       `yaml:"member_rank_index"`   // Index into Guild.Ranks
	Contribution int64     `yaml:"member_contribution"` // Total contribution to guild
	LastActive   time.Time `yaml:"member_last_active"`
	Notes        string    `yaml:"member_notes"`
}

// GuildRank defines a rank within a guild hierarchy.
type GuildRank struct {
	Name            string          `yaml:"rank_name"`
	Level           int             `yaml:"rank_level"`            // 0 = lowest, higher = more authority
	Permissions     GuildPermission `yaml:"rank_permissions"`      // Bitwise permissions
	MinContribution int64           `yaml:"rank_min_contribution"` // Required contribution for promotion
}

// GuildPermission defines what actions a guild member can perform.
type GuildPermission uint32

const (
	PermissionNone           GuildPermission = 0
	PermissionInvite         GuildPermission = 1 << iota // Can invite new members
	PermissionKick                                       // Can remove members
	PermissionPromote                                    // Can promote members
	PermissionDemote                                     // Can demote members
	PermissionWithdraw                                   // Can withdraw from treasury
	PermissionDeposit                                    // Can deposit to treasury
	PermissionEditMotD                                   // Can edit message of the day
	PermissionManageQuests                               // Can accept guild quests
	PermissionManageAlliance                             // Can manage alliances
	PermissionAll            GuildPermission = 0xFFFFFFFF
)

// GuildPerk represents special benefits available to guild members.
type GuildPerk string

const (
	PerkSharedStorage   GuildPerk = "shared_storage"
	PerkBonusExperience GuildPerk = "bonus_experience"
	PerkDiscountShop    GuildPerk = "discount_shop"
	PerkBonusReputation GuildPerk = "bonus_reputation"
	PerkGuildQuests     GuildPerk = "guild_quests"
	PerkFastTravel      GuildPerk = "fast_travel"
	PerkCraftingBonus   GuildPerk = "crafting_bonus"
)

// GuildManager handles all guild operations and memberships.
type GuildManager struct {
	mu              sync.RWMutex
	guilds          map[string]*Guild // Guild ID -> Guild
	characterGuilds map[string]string // Character ID -> Guild ID
	logger          *logrus.Logger
}

// NewGuildManager creates a new guild management system.
func NewGuildManager(logger *logrus.Logger) *GuildManager {
	if logger == nil {
		logger = logrus.New()
	}
	return &GuildManager{
		guilds:          make(map[string]*Guild),
		characterGuilds: make(map[string]string),
		logger:          logger,
	}
}

// Guild operation errors.
var (
	ErrGuildNotFound     = errors.New("guild not found")
	ErrMemberNotFound    = errors.New("member not found")
	ErrAlreadyInGuild    = errors.New("character is already in a guild")
	ErrNotInGuild        = errors.New("character is not in a guild")
	ErrInsufficientRank  = errors.New("insufficient rank for this action")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrCannotKickLeader  = errors.New("cannot kick the guild leader")
	ErrGuildNameTaken    = errors.New("guild name is already taken")
	ErrInvalidRankIndex  = errors.New("invalid rank index")
)

// CreateGuild creates a new guild with the founder as leader.
func (gm *GuildManager) CreateGuild(name, description, founderID string) (*Guild, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if founder is already in a guild
	if _, exists := gm.characterGuilds[founderID]; exists {
		return nil, ErrAlreadyInGuild
	}

	// Check name uniqueness
	for _, g := range gm.guilds {
		if g.Name == name {
			return nil, ErrGuildNameTaken
		}
	}

	guild := &Guild{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Founded:     time.Now(),
		LeaderID:    founderID,
		Members:     make(map[string]*GuildMember),
		Ranks:       defaultGuildRanks(),
		Treasury:    0,
		Experience:  0,
		Level:       1,
		Perks:       []GuildPerk{PerkSharedStorage},
		Properties:  make(map[string]interface{}),
	}

	// Add founder as leader (highest rank)
	guild.Members[founderID] = &GuildMember{
		CharacterID:  founderID,
		JoinDate:     time.Now(),
		RankIndex:    len(guild.Ranks) - 1, // Highest rank
		Contribution: 0,
		LastActive:   time.Now(),
	}

	gm.guilds[guild.ID] = guild
	gm.characterGuilds[founderID] = guild.ID

	gm.logger.WithFields(logrus.Fields{
		"guild_id":   guild.ID,
		"guild_name": name,
		"founder_id": founderID,
	}).Info("guild created")

	return guild, nil
}

// defaultGuildRanks returns the default rank structure for new guilds.
func defaultGuildRanks() []*GuildRank {
	return []*GuildRank{
		{
			Name:        "Initiate",
			Level:       0,
			Permissions: PermissionDeposit,
		},
		{
			Name:            "Member",
			Level:           1,
			Permissions:     PermissionDeposit | PermissionInvite,
			MinContribution: 100,
		},
		{
			Name:            "Veteran",
			Level:           2,
			Permissions:     PermissionDeposit | PermissionInvite | PermissionEditMotD,
			MinContribution: 500,
		},
		{
			Name:            "Officer",
			Level:           3,
			Permissions:     PermissionDeposit | PermissionInvite | PermissionKick | PermissionPromote | PermissionWithdraw | PermissionManageQuests,
			MinContribution: 2000,
		},
		{
			Name:        "Guild Master",
			Level:       4,
			Permissions: PermissionAll,
		},
	}
}

// GetGuild retrieves a guild by ID.
func (gm *GuildManager) GetGuild(guildID string) (*Guild, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	guild, exists := gm.guilds[guildID]
	if !exists {
		return nil, ErrGuildNotFound
	}
	return guild, nil
}

// GetCharacterGuild gets the guild a character belongs to.
func (gm *GuildManager) GetCharacterGuild(characterID string) (*Guild, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	guildID, exists := gm.characterGuilds[characterID]
	if !exists {
		return nil, ErrNotInGuild
	}

	guild, exists := gm.guilds[guildID]
	if !exists {
		return nil, ErrGuildNotFound
	}
	return guild, nil
}

// JoinGuild adds a character to a guild.
func (gm *GuildManager) JoinGuild(guildID, characterID, inviterID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if character is already in a guild
	if _, exists := gm.characterGuilds[characterID]; exists {
		return ErrAlreadyInGuild
	}

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	// Check inviter has permission
	if inviterID != "" {
		if err := gm.checkPermissionUnlocked(guild, inviterID, PermissionInvite); err != nil {
			return err
		}
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	guild.Members[characterID] = &GuildMember{
		CharacterID:  characterID,
		JoinDate:     time.Now(),
		RankIndex:    0, // Start at lowest rank
		Contribution: 0,
		LastActive:   time.Now(),
	}

	gm.characterGuilds[characterID] = guildID

	gm.logger.WithFields(logrus.Fields{
		"guild_id":     guildID,
		"character_id": characterID,
		"inviter_id":   inviterID,
	}).Info("character joined guild")

	return nil
}

// LeaveGuild removes a character from their guild.
func (gm *GuildManager) LeaveGuild(characterID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	guildID, exists := gm.characterGuilds[characterID]
	if !exists {
		return ErrNotInGuild
	}

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	// Check if character is the leader
	if guild.LeaderID == characterID {
		// Transfer leadership or disband
		if len(guild.Members) > 1 {
			// Find next highest rank member
			guild.LeaderID = gm.findNewLeader(guild, characterID)
		}
	}

	delete(guild.Members, characterID)
	delete(gm.characterGuilds, characterID)

	// Disband if empty
	if len(guild.Members) == 0 {
		delete(gm.guilds, guildID)
		gm.logger.WithField("guild_id", guildID).Info("guild disbanded (no members)")
	}

	gm.logger.WithFields(logrus.Fields{
		"guild_id":     guildID,
		"character_id": characterID,
	}).Info("character left guild")

	return nil
}

// findNewLeader selects a new leader when current leader leaves.
func (gm *GuildManager) findNewLeader(guild *Guild, excludeID string) string {
	var bestCandidate string
	bestRank := -1

	for id, member := range guild.Members {
		if id == excludeID {
			continue
		}
		if member.RankIndex > bestRank {
			bestRank = member.RankIndex
			bestCandidate = id
		}
	}
	return bestCandidate
}

// KickMember removes a member from the guild.
func (gm *GuildManager) KickMember(guildID, kickerID, targetID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	// Cannot kick the leader
	if targetID == guild.LeaderID {
		return ErrCannotKickLeader
	}

	// Check permission
	if err := gm.checkPermissionUnlocked(guild, kickerID, PermissionKick); err != nil {
		return err
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	if _, exists := guild.Members[targetID]; !exists {
		return ErrMemberNotFound
	}

	delete(guild.Members, targetID)
	delete(gm.characterGuilds, targetID)

	gm.logger.WithFields(logrus.Fields{
		"guild_id":  guildID,
		"kicker_id": kickerID,
		"target_id": targetID,
	}).Info("member kicked from guild")

	return nil
}

// PromoteMember increases a member's rank.
func (gm *GuildManager) PromoteMember(guildID, promoterID, targetID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	if err := gm.checkPermissionUnlocked(guild, promoterID, PermissionPromote); err != nil {
		return err
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	member, exists := guild.Members[targetID]
	if !exists {
		return ErrMemberNotFound
	}

	// Cannot promote beyond promoter's rank (unless leader)
	promoter := guild.Members[promoterID]
	maxRank := promoter.RankIndex
	if promoterID == guild.LeaderID {
		maxRank = len(guild.Ranks) - 1
	}

	if member.RankIndex >= maxRank-1 {
		return ErrInsufficientRank
	}

	member.RankIndex++

	gm.logger.WithFields(logrus.Fields{
		"guild_id": guildID,
		"promoter": promoterID,
		"target":   targetID,
		"new_rank": guild.Ranks[member.RankIndex].Name,
	}).Info("member promoted")

	return nil
}

// DemoteMember decreases a member's rank.
func (gm *GuildManager) DemoteMember(guildID, demoterID, targetID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	if err := gm.checkPermissionUnlocked(guild, demoterID, PermissionDemote); err != nil {
		return err
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	member, exists := guild.Members[targetID]
	if !exists {
		return ErrMemberNotFound
	}

	if member.RankIndex <= 0 {
		return ErrInvalidRankIndex
	}

	// Cannot demote someone of equal or higher rank (unless leader)
	demoter := guild.Members[demoterID]
	if demoterID != guild.LeaderID && member.RankIndex >= demoter.RankIndex {
		return ErrInsufficientRank
	}

	member.RankIndex--

	gm.logger.WithFields(logrus.Fields{
		"guild_id": guildID,
		"demoter":  demoterID,
		"target":   targetID,
		"new_rank": guild.Ranks[member.RankIndex].Name,
	}).Info("member demoted")

	return nil
}

// Deposit adds gold to the guild treasury.
func (gm *GuildManager) Deposit(guildID, characterID string, amount int) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	if err := gm.checkPermissionUnlocked(guild, characterID, PermissionDeposit); err != nil {
		return err
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	guild.Treasury += amount

	// Add to contribution
	if member, exists := guild.Members[characterID]; exists {
		member.Contribution += int64(amount)
		member.LastActive = time.Now()
	}

	gm.logger.WithFields(logrus.Fields{
		"guild_id":     guildID,
		"character_id": characterID,
		"amount":       amount,
	}).Info("gold deposited to guild treasury")

	return nil
}

// Withdraw removes gold from the guild treasury.
func (gm *GuildManager) Withdraw(guildID, characterID string, amount int) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	if err := gm.checkPermissionUnlocked(guild, characterID, PermissionWithdraw); err != nil {
		return err
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	if guild.Treasury < amount {
		return ErrInsufficientFunds
	}

	guild.Treasury -= amount

	gm.logger.WithFields(logrus.Fields{
		"guild_id":     guildID,
		"character_id": characterID,
		"amount":       amount,
	}).Info("gold withdrawn from guild treasury")

	return nil
}

// AddExperience adds experience to the guild for leveling.
func (gm *GuildManager) AddExperience(guildID string, amount int64) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	guild.Experience += amount

	// Check for level up
	requiredExp := gm.experienceForLevel(guild.Level + 1)
	if guild.Experience >= requiredExp {
		guild.Level++
		gm.unlockPerks(guild)

		gm.logger.WithFields(logrus.Fields{
			"guild_id":  guildID,
			"new_level": guild.Level,
		}).Info("guild leveled up")
	}

	return nil
}

// experienceForLevel calculates required experience for a guild level.
func (gm *GuildManager) experienceForLevel(level int) int64 {
	return int64(level * level * 1000)
}

// unlockPerks grants new perks based on guild level.
func (gm *GuildManager) unlockPerks(guild *Guild) {
	perksByLevel := map[int]GuildPerk{
		2:  PerkBonusExperience,
		3:  PerkDiscountShop,
		5:  PerkGuildQuests,
		7:  PerkBonusReputation,
		10: PerkFastTravel,
		15: PerkCraftingBonus,
	}

	if perk, exists := perksByLevel[guild.Level]; exists {
		guild.Perks = append(guild.Perks, perk)
	}
}

// HasPerk checks if a guild has a specific perk.
// HasPerk checks if the guild has unlocked the specified perk.
// Returns true if the perk is in the guild's active perks list.
func (g *Guild) HasPerk(perk GuildPerk) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, p := range g.Perks {
		if p == perk {
			return true
		}
	}
	return false
}

// GetMemberRank returns the rank of a guild member.
func (g *Guild) GetMemberRank(characterID string) (*GuildRank, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	member, exists := g.Members[characterID]
	if !exists {
		return nil, ErrMemberNotFound
	}

	if member.RankIndex < 0 || member.RankIndex >= len(g.Ranks) {
		return nil, ErrInvalidRankIndex
	}

	return g.Ranks[member.RankIndex], nil
}

// MemberCount returns the number of members in the guild.
func (g *Guild) MemberCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Members)
}

// checkPermissionUnlocked verifies a member has the required permission.
// Caller must hold gm.mu lock.
func (gm *GuildManager) checkPermissionUnlocked(guild *Guild, characterID string, perm GuildPermission) error {
	guild.mu.RLock()
	defer guild.mu.RUnlock()

	member, exists := guild.Members[characterID]
	if !exists {
		return ErrMemberNotFound
	}

	if member.RankIndex < 0 || member.RankIndex >= len(guild.Ranks) {
		return ErrInvalidRankIndex
	}

	rank := guild.Ranks[member.RankIndex]
	if rank.Permissions&perm == 0 {
		return ErrInsufficientRank
	}

	return nil
}

// ListGuilds returns all guilds (for admin or discovery).
func (gm *GuildManager) ListGuilds() []*Guild {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	guilds := make([]*Guild, 0, len(gm.guilds))
	for _, g := range gm.guilds {
		guilds = append(guilds, g)
	}
	return guilds
}

// TransferLeadership transfers guild leadership to another member.
func (gm *GuildManager) TransferLeadership(guildID, currentLeaderID, newLeaderID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	guild, exists := gm.guilds[guildID]
	if !exists {
		return ErrGuildNotFound
	}

	if guild.LeaderID != currentLeaderID {
		return ErrInsufficientRank
	}

	guild.mu.Lock()
	defer guild.mu.Unlock()

	if _, exists := guild.Members[newLeaderID]; !exists {
		return ErrMemberNotFound
	}

	// Promote new leader to highest rank
	guild.Members[newLeaderID].RankIndex = len(guild.Ranks) - 1
	guild.LeaderID = newLeaderID

	gm.logger.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"old_leader": currentLeaderID,
		"new_leader": newLeaderID,
	}).Info("guild leadership transferred")

	return nil
}
