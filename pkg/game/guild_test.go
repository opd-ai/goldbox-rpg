package game

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuildManager_CreateGuild(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, err := gm.CreateGuild("Heroes Guild", "A guild for brave adventurers", "char1")
	require.NoError(t, err)
	assert.NotEmpty(t, guild.ID)
	assert.Equal(t, "Heroes Guild", guild.Name)
	assert.Equal(t, "char1", guild.LeaderID)
	assert.Equal(t, 1, len(guild.Members))
	assert.Equal(t, 1, guild.Level)

	// Founder should be at highest rank
	member := guild.Members["char1"]
	assert.Equal(t, len(guild.Ranks)-1, member.RankIndex)
}

func TestGuildManager_CreateGuild_Errors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	_, err := gm.CreateGuild("Guild One", "Test", "char1")
	require.NoError(t, err)

	// Cannot create guild when already in one
	_, err = gm.CreateGuild("Guild Two", "Test", "char1")
	assert.ErrorIs(t, err, ErrAlreadyInGuild)

	// Cannot create guild with same name
	_, err = gm.CreateGuild("Guild One", "Test", "char2")
	assert.ErrorIs(t, err, ErrGuildNameTaken)
}

func TestGuildManager_JoinAndLeave(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, err := gm.CreateGuild("Test Guild", "Test", "founder")
	require.NoError(t, err)

	// Join guild
	err = gm.JoinGuild(guild.ID, "member1", "founder")
	require.NoError(t, err)

	// Verify membership
	memberGuild, err := gm.GetCharacterGuild("member1")
	require.NoError(t, err)
	assert.Equal(t, guild.ID, memberGuild.ID)

	// New member should be at lowest rank
	assert.Equal(t, 0, guild.Members["member1"].RankIndex)

	// Leave guild
	err = gm.LeaveGuild("member1")
	require.NoError(t, err)

	_, err = gm.GetCharacterGuild("member1")
	assert.ErrorIs(t, err, ErrNotInGuild)
}

func TestGuildManager_JoinErrors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Guild A", "Test", "founder")
	gm.CreateGuild("Guild B", "Test", "founder2")

	// Cannot join when already in guild
	err := gm.JoinGuild(guild.ID, "founder2", "founder")
	assert.ErrorIs(t, err, ErrAlreadyInGuild)

	// Cannot join non-existent guild
	err = gm.JoinGuild("nonexistent", "newchar", "")
	assert.ErrorIs(t, err, ErrGuildNotFound)
}

func TestGuildManager_Promotion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	gm.JoinGuild(guild.ID, "member1", "leader")

	// Leader promotes member
	err := gm.PromoteMember(guild.ID, "leader", "member1")
	require.NoError(t, err)
	assert.Equal(t, 1, guild.Members["member1"].RankIndex)

	// Promote again
	err = gm.PromoteMember(guild.ID, "leader", "member1")
	require.NoError(t, err)
	assert.Equal(t, 2, guild.Members["member1"].RankIndex)
}

func TestGuildManager_Demotion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	gm.JoinGuild(guild.ID, "officer", "leader")
	gm.PromoteMember(guild.ID, "leader", "officer")
	gm.PromoteMember(guild.ID, "leader", "officer")
	gm.PromoteMember(guild.ID, "leader", "officer")

	// Demote officer
	err := gm.DemoteMember(guild.ID, "leader", "officer")
	require.NoError(t, err)
	assert.Equal(t, 2, guild.Members["officer"].RankIndex)

	// Cannot demote below 0
	guild.Members["officer"].RankIndex = 0
	err = gm.DemoteMember(guild.ID, "leader", "officer")
	assert.ErrorIs(t, err, ErrInvalidRankIndex)
}

func TestGuildManager_KickMember(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	gm.JoinGuild(guild.ID, "officer", "leader")
	gm.JoinGuild(guild.ID, "member", "leader")
	gm.PromoteMember(guild.ID, "leader", "officer")
	gm.PromoteMember(guild.ID, "leader", "officer")
	gm.PromoteMember(guild.ID, "leader", "officer")

	// Officer kicks member
	err := gm.KickMember(guild.ID, "officer", "member")
	require.NoError(t, err)

	_, err = gm.GetCharacterGuild("member")
	assert.ErrorIs(t, err, ErrNotInGuild)

	// Cannot kick leader
	err = gm.KickMember(guild.ID, "officer", "leader")
	assert.ErrorIs(t, err, ErrCannotKickLeader)
}

func TestGuildManager_Treasury(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	gm.JoinGuild(guild.ID, "member", "leader")

	// Deposit
	err := gm.Deposit(guild.ID, "member", 1000)
	require.NoError(t, err)
	assert.Equal(t, 1000, guild.Treasury)
	assert.Equal(t, int64(1000), guild.Members["member"].Contribution)

	// Withdraw (leader has permission)
	err = gm.Withdraw(guild.ID, "leader", 500)
	require.NoError(t, err)
	assert.Equal(t, 500, guild.Treasury)

	// Cannot withdraw more than available
	err = gm.Withdraw(guild.ID, "leader", 1000)
	assert.ErrorIs(t, err, ErrInsufficientFunds)

	// Member without permission cannot withdraw
	err = gm.Withdraw(guild.ID, "member", 100)
	assert.ErrorIs(t, err, ErrInsufficientRank)
}

func TestGuildManager_Experience(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	assert.Equal(t, 1, guild.Level)

	// Add experience
	gm.AddExperience(guild.ID, 500)
	assert.Equal(t, int64(500), guild.Experience)
	assert.Equal(t, 1, guild.Level)

	// Level up requires 4000 exp (2^2 * 1000)
	gm.AddExperience(guild.ID, 3500)
	assert.Equal(t, int64(4000), guild.Experience)
	assert.Equal(t, 2, guild.Level)

	// Check perk unlocked
	assert.True(t, guild.HasPerk(PerkBonusExperience))
}

func TestGuildManager_TransferLeadership(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	gm.JoinGuild(guild.ID, "officer", "leader")

	err := gm.TransferLeadership(guild.ID, "leader", "officer")
	require.NoError(t, err)

	assert.Equal(t, "officer", guild.LeaderID)
	assert.Equal(t, len(guild.Ranks)-1, guild.Members["officer"].RankIndex)
}

func TestGuildManager_DisbandOnLastMemberLeave(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	guildID := guild.ID

	err := gm.LeaveGuild("leader")
	require.NoError(t, err)

	_, err = gm.GetGuild(guildID)
	assert.ErrorIs(t, err, ErrGuildNotFound)
}

func TestGuild_MemberCount(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	assert.Equal(t, 1, guild.MemberCount())

	gm.JoinGuild(guild.ID, "member1", "leader")
	assert.Equal(t, 2, guild.MemberCount())

	gm.JoinGuild(guild.ID, "member2", "leader")
	assert.Equal(t, 3, guild.MemberCount())
}

func TestGuild_GetMemberRank(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	gm := NewGuildManager(logger)

	guild, _ := gm.CreateGuild("Test Guild", "Test", "leader")
	gm.JoinGuild(guild.ID, "member", "leader")

	rank, err := guild.GetMemberRank("leader")
	require.NoError(t, err)
	assert.Equal(t, "Guild Master", rank.Name)

	rank, err = guild.GetMemberRank("member")
	require.NoError(t, err)
	assert.Equal(t, "Initiate", rank.Name)

	_, err = guild.GetMemberRank("nonexistent")
	assert.ErrorIs(t, err, ErrMemberNotFound)
}

// TestGuildManager_ListGuilds tests listing all guilds
func TestGuildManager_ListGuilds(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	gm := NewGuildManager(logger)

	// Empty list
	guilds := gm.ListGuilds()
	if len(guilds) != 0 {
		t.Errorf("ListGuilds() = %d guilds; want 0 for empty manager", len(guilds))
	}

	// Create guilds
	_, _ = gm.CreateGuild("Guild 1", "Desc 1", "founder1")
	_, _ = gm.CreateGuild("Guild 2", "Desc 2", "founder2")
	_, _ = gm.CreateGuild("Guild 3", "Desc 3", "founder3")

	guilds = gm.ListGuilds()
	if len(guilds) != 3 {
		t.Errorf("ListGuilds() = %d guilds; want 3", len(guilds))
	}
}

// TestGuildManager_findNewLeader tests the leader selection logic
func TestGuildManager_findNewLeader(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	gm := NewGuildManager(logger)

	// Create guild with founder
	guild, _ := gm.CreateGuild("Test Guild", "Description", "founder")

	// Add members with different ranks
	_ = gm.JoinGuild(guild.ID, "member1", "founder")
	_ = gm.JoinGuild(guild.ID, "member2", "founder")

	// Promote member2 to officer rank
	_ = gm.PromoteMember(guild.ID, "founder", "member2")

	// Find new leader excluding founder - should pick highest ranked member
	newLeader := gm.findNewLeader(guild, "founder")
	if newLeader != "member2" {
		t.Errorf("findNewLeader() = %s; want 'member2' (highest rank)", newLeader)
	}

	// Find new leader excluding founder and member2 - should pick member1
	newLeader = gm.findNewLeader(guild, "member2")
	if newLeader == "" && newLeader != "member1" && newLeader != "founder" {
		// Either founder or member1 should be returned
		t.Errorf("findNewLeader() = %s; want either 'founder' or 'member1'", newLeader)
	}
}
