package server

import (
	"encoding/json"
	"fmt"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// handleGetSpell retrieves a specific spell by ID from the spell database.
//
// Parameters (JSON):
//   - spell_id: string - The unique identifier of the spell to retrieve
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - spell: object containing the spell data
//
// Errors:
//   - "invalid spell ID" if spell_id is empty or not provided
//   - "spell not found" if the spell doesn't exist in the database
func (s *RPCServer) handleGetSpell(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpell",
	}).Debug("entering handleGetSpell")

	var req struct {
		SpellID string `json:"spell_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetSpell",
			"error":    err.Error(),
		}).Error("failed to unmarshal get spell parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid get spell parameters", err.Error())
	}

	if req.SpellID == "" {
		return nil, fmt.Errorf("spell ID cannot be empty")
	}

	spell, err := s.spellManager.GetSpell(req.SpellID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetSpell",
			"spellID":  req.SpellID,
			"error":    err.Error(),
		}).Error("spell not found")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpell",
		"spellID":  req.SpellID,
	}).Info("spell retrieved successfully")

	return map[string]interface{}{
		"success": true,
		"spell":   spell,
	}, nil
}

// handleGetSpellsByLevel retrieves all spells of a specific level.
//
// Parameters (JSON):
//   - level: int - The spell level to filter by (0 for cantrips, 1+ for leveled spells)
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - spells: array of spell objects
//   - count: int number of spells found
func (s *RPCServer) handleGetSpellsByLevel(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpellsByLevel",
	}).Debug("entering handleGetSpellsByLevel")

	var req struct {
		Level int `json:"level"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetSpellsByLevel",
			"error":    err.Error(),
		}).Error("failed to unmarshal get spells by level parameters")
		return nil, fmt.Errorf("invalid get spells by level parameters")
	}

	spells := s.spellManager.GetSpellsByLevel(req.Level)

	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpellsByLevel",
		"level":    req.Level,
		"count":    len(spells),
	}).Info("spells retrieved by level")

	return map[string]interface{}{
		"success": true,
		"spells":  spells,
		"count":   len(spells),
		"level":   req.Level,
	}, nil
}

// handleGetSpellsBySchool retrieves all spells of a specific school.
//
// Parameters (JSON):
//   - school: string - The spell school to filter by (e.g., "evocation", "abjuration")
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - spells: array of spell objects
//   - count: int number of spells found
//   - school: string the school name searched
func (s *RPCServer) handleGetSpellsBySchool(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpellsBySchool",
	}).Debug("entering handleGetSpellsBySchool")

	var req struct {
		School string `json:"school"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetSpellsBySchool",
			"error":    err.Error(),
		}).Error("failed to unmarshal get spells by school parameters")
		return nil, fmt.Errorf("invalid get spells by school parameters")
	}

	if req.School == "" {
		return nil, fmt.Errorf("school cannot be empty")
	}

	school := game.ParseSpellSchool(req.School)
	spells := s.spellManager.GetSpellsBySchool(school)

	logrus.WithFields(logrus.Fields{
		"function": "handleGetSpellsBySchool",
		"school":   req.School,
		"count":    len(spells),
	}).Info("spells retrieved by school")

	return map[string]interface{}{
		"success": true,
		"spells":  spells,
		"count":   len(spells),
		"school":  req.School,
	}, nil
}

// handleGetAllSpells retrieves all spells in the spell database.
//
// Parameters: None required
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - spells: array of all spell objects
//   - count: int total number of spells
//   - by_level: map of spell counts by level
func (s *RPCServer) handleGetAllSpells(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetAllSpells",
	}).Debug("entering handleGetAllSpells")

	spells := s.spellManager.GetAllSpells()
	countsByLevel := s.spellManager.GetSpellCountByLevel()

	logrus.WithFields(logrus.Fields{
		"function": "handleGetAllSpells",
		"count":    len(spells),
	}).Info("all spells retrieved")

	return map[string]interface{}{
		"success":  true,
		"spells":   spells,
		"count":    len(spells),
		"by_level": countsByLevel,
	}, nil
}

// handleSearchSpells searches for spells by name, description, or keywords.
//
// Parameters (JSON):
//   - query: string - The search query to match against spell names, descriptions, and keywords
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if search was successful
//   - spells: array of matching spell objects
//   - count: int number of spells found
//   - query: string the search query used
func (s *RPCServer) handleSearchSpells(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleSearchSpells",
	}).Debug("entering handleSearchSpells")

	var req struct {
		Query string `json:"query"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleSearchSpells",
			"error":    err.Error(),
		}).Error("failed to unmarshal search spells parameters")
		return nil, fmt.Errorf("invalid search spells parameters")
	}

	if req.Query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	spells := s.spellManager.SearchSpells(req.Query)

	logrus.WithFields(logrus.Fields{
		"function": "handleSearchSpells",
		"query":    req.Query,
		"count":    len(spells),
	}).Info("spell search completed")

	return map[string]interface{}{
		"success": true,
		"spells":  spells,
		"count":   len(spells),
		"query":   req.Query,
	}, nil
}
