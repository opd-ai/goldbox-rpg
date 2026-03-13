package server

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// handleAdventureList processes a request to list all available adventures.
// This handler returns lightweight summaries of all loaded adventure packs.
//
// Parameters:
//   - params: json.RawMessage (no required parameters)
//
// Returns:
//   - interface{}: List of adventure summaries with metadata
//   - error: Error if request fails
func (s *RPCServer) handleAdventureList(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleAdventureList",
	})
	logger.Debug("entering handleAdventureList")

	// Get adventure manager from server
	advMgr := s.getAdventureManager()
	if advMgr == nil {
		logger.Warn("adventure manager not initialized")
		return map[string]interface{}{
			"success":    true,
			"adventures": []interface{}{},
			"count":      0,
			"message":    "No adventures available",
		}, nil
	}

	summaries := advMgr.List()

	logger.WithField("count", len(summaries)).Debug("exiting handleAdventureList")

	return map[string]interface{}{
		"success":    true,
		"adventures": summaries,
		"count":      len(summaries),
	}, nil
}

// handleAdventureLoad processes a request to load a specific adventure by slug.
// This handler returns the full adventure data including maps, NPCs, and quests.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - slug: string - The slug identifier of the adventure to load
//
// Returns:
//   - interface{}: Full adventure data if found
//   - error: Error if request fails or adventure not found
func (s *RPCServer) handleAdventureLoad(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleAdventureLoad",
	})
	logger.Debug("entering handleAdventureLoad")

	var req struct {
		Slug string `json:"slug"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	// Get adventure manager from server
	advMgr := s.getAdventureManager()
	if advMgr == nil {
		return nil, fmt.Errorf("adventure manager not initialized")
	}

	adventure, err := advMgr.Get(req.Slug)
	if err != nil {
		logger.WithError(err).WithField("slug", req.Slug).Error("failed to load adventure")
		return nil, fmt.Errorf("adventure not found: %w", err)
	}

	logger.WithField("slug", req.Slug).Debug("exiting handleAdventureLoad")

	return map[string]interface{}{
		"success":   true,
		"adventure": adventure,
	}, nil
}
