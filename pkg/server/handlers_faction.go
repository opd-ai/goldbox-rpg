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

// factionOp is the signature for simple faction operations.
type factionOp func(faction1ID, faction2ID string) error

// executeFactionOp handles common pattern for simple faction operations.
func (s *RPCServer) executeFactionOp(params json.RawMessage, opName, successMsg string, op factionOp) (interface{}, error) {
	logrus.WithField("function", opName).Debug("entering " + opName)

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	if err := op(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed: "+opName, err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": successMsg,
	}, nil
}

// executeFactionOpWithRelation handles faction operations that require relation initialization.
func (s *RPCServer) executeFactionOpWithRelation(params json.RawMessage, opName, successMsg string, op factionOp) (interface{}, error) {
	logrus.WithField("function", opName).Debug("entering " + opName)

	var req factionActionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	if err := s.ensureFactionRelation(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, err
	}

	if err := op(req.Faction1ID, req.Faction2ID); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed: "+opName, err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": successMsg,
	}, nil
}

// ensureFactionRelation initializes a relation if it doesn't exist.
func (s *RPCServer) ensureFactionRelation(faction1ID, faction2ID string) error {
	_, err := s.diplomacyManager.GetRelation(faction1ID, faction2ID)
	if err != nil {
		_, err = s.diplomacyManager.InitializeRelation(faction1ID, faction2ID)
		if err != nil {
			return NewJSONRPCError(JSONRPCInternalError, "failed to initialize relation", err.Error())
		}
	}
	return nil
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
	return s.executeFactionOp(params, "handleOfferPeace", "peace offer sent", s.diplomacyManager.OfferPeace)
}

// handleAcceptPeace accepts a peace offer and ends the war.
func (s *RPCServer) handleAcceptPeace(params json.RawMessage) (interface{}, error) {
	return s.executeFactionOp(params, "handleAcceptPeace", "peace accepted", s.diplomacyManager.AcceptPeace)
}

// handleProposeAlliance sends an alliance proposal between factions.
func (s *RPCServer) handleProposeAlliance(params json.RawMessage) (interface{}, error) {
	return s.executeFactionOpWithRelation(params, "handleProposeAlliance", "alliance proposed", s.diplomacyManager.ProposeAlliance)
}

// handleAcceptAlliance accepts an alliance proposal.
func (s *RPCServer) handleAcceptAlliance(params json.RawMessage) (interface{}, error) {
	return s.executeFactionOp(params, "handleAcceptAlliance", "alliance accepted", s.diplomacyManager.AcceptAlliance)
}

// handleBreakAlliance ends an alliance between factions.
func (s *RPCServer) handleBreakAlliance(params json.RawMessage) (interface{}, error) {
	return s.executeFactionOp(params, "handleBreakAlliance", "alliance broken", s.diplomacyManager.BreakAlliance)
}

// handleSignTrade establishes a trade agreement between factions.
func (s *RPCServer) handleSignTrade(params json.RawMessage) (interface{}, error) {
	return s.executeFactionOpWithRelation(params, "handleSignTrade", "trade agreement signed", s.diplomacyManager.SignTradeAgreement)
}

// handleSendDiplomaticGift sends a gift to improve diplomatic relations.
func (s *RPCServer) handleSendDiplomaticGift(params json.RawMessage) (interface{}, error) {
	logrus.WithField("function", "handleSendDiplomaticGift").Debug("entering handleSendDiplomaticGift")

	var req diplomaticGiftRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "invalid parameters", err.Error())
	}

	if err := s.ensureFactionRelation(req.SenderID, req.ReceiverID); err != nil {
		return nil, err
	}

	if err := s.diplomacyManager.SendGift(req.SenderID, req.ReceiverID, req.Value); err != nil {
		return nil, NewJSONRPCError(JSONRPCInternalError, "failed to send gift", err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": "gift sent",
	}, nil
}
