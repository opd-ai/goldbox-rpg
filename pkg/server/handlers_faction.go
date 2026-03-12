// Package server provides faction diplomacy RPC handlers.
package server

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// Faction diplomacy request types

type factionRelationRequest struct {
	Faction1ID string `json:"faction1_id"`
	Faction2ID string `json:"faction2_id"`
}

type factionActionRequest struct {
	SessionID  string `json:"session_id"`
	Faction1ID string `json:"faction1_id"`
	Faction2ID string `json:"faction2_id"`
	Reason     string `json:"reason,omitempty"`
}

type diplomaticGiftRequest struct {
	SessionID  string `json:"session_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Value      int    `json:"value"`
}

// handleGetFactionRelation retrieves the diplomatic relation between two factions.
func (s *RPCServer) handleGetFactionRelation(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleGetFactionRelation").Debug("entering handleGetFactionRelation")

	var req factionRelationRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	relation, err := s.diplomacyManager.GetRelation(req.Faction1ID, req.Faction2ID)
	if err != nil {
		// Try to initialize if not found
		relation, err = s.diplomacyManager.InitializeRelation(req.Faction1ID, req.Faction2ID)
		if err != nil {
			return nil, NewJSONRPCError(JSONRPCInternalError, "failed to get relation", err.Error())
		}
	}

	return map[string]interface{}{
		"success":  true,
		"relation": relation,
	}, nil
}

// handleGetFactionRelations retrieves all diplomatic relations for a faction.
func (s *RPCServer) handleGetFactionRelations(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleGetFactionRelations").Debug("entering handleGetFactionRelations")

	var req struct {
		FactionID string `json:"faction_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	relations := s.diplomacyManager.GetFactionRelations(req.FactionID)

	return map[string]interface{}{
		"success":   true,
		"relations": relations,
		"count":     len(relations),
	}, nil
}

// handleDeclareWar initiates war between two factions.
func (s *RPCServer) handleDeclareWar(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleDeclareWar").Debug("entering handleDeclareWar")

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	// Ensure relation exists
	_, err := s.diplomacyManager.GetRelation(req.Faction1ID, req.Faction2ID)
	if err != nil {
		_, err = s.diplomacyManager.InitializeRelation(req.Faction1ID, req.Faction2ID)
		if err != nil {
			return nil, NewJSONRPCError(JSONRPCInternalError, "failed to initialize relation", err.Error())
		}
	}

	if err := s.diplomacyManager.DeclareWar(req.Faction1ID, req.Faction2ID, req.Reason); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to declare war", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "war declared",
	}, nil
}

// handleOfferPeace sends a peace offer from one faction to another.
func (s *RPCServer) handleOfferPeace(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleOfferPeace").Debug("entering handleOfferPeace")

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	if err := s.diplomacyManager.OfferPeace(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to offer peace", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "peace offer sent",
	}, nil
}

// handleAcceptPeace accepts a peace offer and ends the war.
func (s *RPCServer) handleAcceptPeace(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleAcceptPeace").Debug("entering handleAcceptPeace")

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	if err := s.diplomacyManager.AcceptPeace(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to accept peace", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "peace accepted",
	}, nil
}

// handleProposeAlliance sends an alliance proposal between factions.
func (s *RPCServer) handleProposeAlliance(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleProposeAlliance").Debug("entering handleProposeAlliance")

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	// Ensure relation exists
	_, err := s.diplomacyManager.GetRelation(req.Faction1ID, req.Faction2ID)
	if err != nil {
		_, err = s.diplomacyManager.InitializeRelation(req.Faction1ID, req.Faction2ID)
		if err != nil {
			return nil, NewJSONRPCError(JSONRPCInternalError, "failed to initialize relation", err.Error())
		}
	}

	if err := s.diplomacyManager.ProposeAlliance(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to propose alliance", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "alliance proposed",
	}, nil
}

// handleAcceptAlliance accepts an alliance proposal.
func (s *RPCServer) handleAcceptAlliance(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleAcceptAlliance").Debug("entering handleAcceptAlliance")

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	if err := s.diplomacyManager.AcceptAlliance(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to accept alliance", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "alliance accepted",
	}, nil
}

// handleBreakAlliance ends an alliance between factions.
func (s *RPCServer) handleBreakAlliance(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleBreakAlliance").Debug("entering handleBreakAlliance")

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	if err := s.diplomacyManager.BreakAlliance(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to break alliance", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "alliance broken",
	}, nil
}

// handleSignTrade establishes a trade agreement between factions.
func (s *RPCServer) handleSignTrade(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleSignTrade").Debug("entering handleSignTrade")

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	// Ensure relation exists
	_, err := s.diplomacyManager.GetRelation(req.Faction1ID, req.Faction2ID)
	if err != nil {
		_, err = s.diplomacyManager.InitializeRelation(req.Faction1ID, req.Faction2ID)
		if err != nil {
			return nil, NewJSONRPCError(JSONRPCInternalError, "failed to initialize relation", err.Error())
		}
	}

	if err := s.diplomacyManager.SignTradeAgreement(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to sign trade agreement", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "trade agreement signed",
	}, nil
}

// handleSendDiplomaticGift sends a gift to improve diplomatic relations.
func (s *RPCServer) handleSendDiplomaticGift(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleSendDiplomaticGift").Debug("entering handleSendDiplomaticGift")

	var req diplomaticGiftRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	// Ensure relation exists
	_, err := s.diplomacyManager.GetRelation(req.SenderID, req.ReceiverID)
	if err != nil {
		_, err = s.diplomacyManager.InitializeRelation(req.SenderID, req.ReceiverID)
		if err != nil {
			return nil, NewJSONRPCError(JSONRPCInternalError, "failed to initialize relation", err.Error())
		}
	}

	if err := s.diplomacyManager.SendGift(req.SenderID, req.ReceiverID, req.Value); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to send gift", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "gift sent",
	}, nil
}
