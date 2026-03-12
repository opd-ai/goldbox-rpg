// Package server provides guild and faction management RPC handlers.
package server

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// Guild request/response types

type createGuildRequest struct {
	SessionID   string `json:"session_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type guildIDRequest struct {
	SessionID string `json:"session_id"`
	GuildID   string `json:"guild_id"`
}

type joinGuildRequest struct {
	SessionID   string `json:"session_id"`
	GuildID     string `json:"guild_id"`
	CharacterID string `json:"character_id"`
	InviterID   string `json:"inviter_id,omitempty"`
}

type guildMemberRequest struct {
	SessionID string `json:"session_id"`
	GuildID   string `json:"guild_id"`
	TargetID  string `json:"target_id"`
}

type guildTransactionRequest struct {
	SessionID string `json:"session_id"`
	GuildID   string `json:"guild_id"`
	Amount    int    `json:"amount"`
}

type transferLeaderRequest struct {
	SessionID   string `json:"session_id"`
	GuildID     string `json:"guild_id"`
	NewLeaderID string `json:"new_leader_id"`
}

// handleCreateGuild creates a new guild with the session's character as founder.
func (s *RPCServer) handleCreateGuild(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleCreateGuild").Debug("entering handleCreateGuild")

	var req createGuildRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	characterID := session.Player.GetID()
	guild, err := s.guildManager.CreateGuild(req.Name, req.Description, characterID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to create guild", err.Error())
	}

	return map[string]interface{}{
		"success":  true,
		"guild_id": guild.ID,
		"guild":    guild,
	}, nil
}

// handleGetGuild retrieves guild information by ID.
func (s *RPCServer) handleGetGuild(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleGetGuild").Debug("entering handleGetGuild")

	var req guildIDRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	guild, err := s.guildManager.GetGuild(req.GuildID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "guild not found", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"guild":   guild,
	}, nil
}

// handleGetCharacterGuild retrieves the guild for the session's character.
func (s *RPCServer) handleGetCharacterGuild(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleGetCharacterGuild").Debug("entering handleGetCharacterGuild")

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	characterID := session.Player.GetID()
	guild, err := s.guildManager.GetCharacterGuild(characterID)
	if err != nil {
		return map[string]interface{}{
			"success": true,
			"guild":   nil,
			"message": "character is not in a guild",
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"guild":   guild,
	}, nil
}

// handleJoinGuild adds the session's character to a guild.
func (s *RPCServer) handleJoinGuild(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleJoinGuild").Debug("entering handleJoinGuild")

	var req joinGuildRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	characterID := session.Player.GetID()
	if err := s.guildManager.JoinGuild(req.GuildID, characterID, req.InviterID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to join guild", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "successfully joined guild",
	}, nil
}

// handleLeaveGuild removes the session's character from their guild.
func (s *RPCServer) handleLeaveGuild(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleLeaveGuild").Debug("entering handleLeaveGuild")

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	characterID := session.Player.GetID()
	if err := s.guildManager.LeaveGuild(characterID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to leave guild", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "successfully left guild",
	}, nil
}

// handleKickGuildMember removes a member from a guild.
func (s *RPCServer) handleKickGuildMember(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleKickGuildMember").Debug("entering handleKickGuildMember")

	var req guildMemberRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	kickerID := session.Player.GetID()
	if err := s.guildManager.KickMember(req.GuildID, kickerID, req.TargetID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to kick member", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "member kicked from guild",
	}, nil
}

// handlePromoteGuildMember promotes a guild member.
func (s *RPCServer) handlePromoteGuildMember(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handlePromoteGuildMember").Debug("entering handlePromoteGuildMember")

	var req guildMemberRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	promoterID := session.Player.GetID()
	if err := s.guildManager.PromoteMember(req.GuildID, promoterID, req.TargetID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to promote member", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "member promoted",
	}, nil
}

// handleDemoteGuildMember demotes a guild member.
func (s *RPCServer) handleDemoteGuildMember(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleDemoteGuildMember").Debug("entering handleDemoteGuildMember")

	var req guildMemberRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	demoterID := session.Player.GetID()
	if err := s.guildManager.DemoteMember(req.GuildID, demoterID, req.TargetID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to demote member", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "member demoted",
	}, nil
}

// handleGuildDeposit deposits gold into the guild treasury.
func (s *RPCServer) handleGuildDeposit(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleGuildDeposit").Debug("entering handleGuildDeposit")

	var req guildTransactionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	characterID := session.Player.GetID()
	if err := s.guildManager.Deposit(req.GuildID, characterID, req.Amount); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to deposit", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "deposit successful",
	}, nil
}

// handleGuildWithdraw withdraws gold from the guild treasury.
func (s *RPCServer) handleGuildWithdraw(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleGuildWithdraw").Debug("entering handleGuildWithdraw")

	var req guildTransactionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	characterID := session.Player.GetID()
	if err := s.guildManager.Withdraw(req.GuildID, characterID, req.Amount); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to withdraw", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "withdrawal successful",
	}, nil
}

// handleListGuilds returns all guilds in the system.
func (s *RPCServer) handleListGuilds(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleListGuilds").Debug("entering handleListGuilds")

	guilds := s.guildManager.ListGuilds()

	return map[string]interface{}{
		"success": true,
		"guilds":  guilds,
		"count":   len(guilds),
	}, nil
}

// handleTransferGuildLeader transfers guild leadership to another member.
func (s *RPCServer) handleTransferGuildLeader(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleTransferGuildLeader").Debug("entering handleTransferGuildLeader")

	var req transferLeaderRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	session, err := s.getSessionForMove(req.SessionID)
	if err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "invalid session", err.Error())
	}
	defer s.releaseSession(session)

	currentLeaderID := session.Player.GetID()
	if err := s.guildManager.TransferLeadership(req.GuildID, currentLeaderID, req.NewLeaderID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to transfer leadership", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "leadership transferred",
	}, nil
}
