package server

import (
	"encoding/json"
	"fmt"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// handleStartQuest processes a request to start a new quest for a player.
// This handler validates the quest data and adds it to the player's quest log.
//
// Parameters:
//   - params: json.RawMessage containing the start quest request with:
//   - session_id: string - The session ID of the requesting player
//   - quest: Quest object - The quest data to start
//
// Returns:
//   - interface{}: Success response with quest ID if quest started successfully
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest validation failures
//   - Quest already exists in player's quest log
func (s *RPCServer) handleStartQuest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleStartQuest",
	})
	logger.Debug("entering handleStartQuest")

	var req struct {
		SessionID string     `json:"session_id"`
		Quest     game.Quest `json:"quest"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleStartQuest",
		}).Error("failed to unmarshal request parameters")
		s.recordActionMetrics("start_quest", err)
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleStartQuest",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		s.recordActionMetrics("start_quest", err)
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Start quest for player
	if err := session.Player.StartQuest(req.Quest); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleStartQuest",
			"quest_id": req.Quest.ID,
		}).Error("failed to start quest")
		s.recordActionMetrics("start_quest", err)
		return nil, fmt.Errorf("failed to start quest: %w", err)
	}

	s.recordActionMetrics("start_quest", nil)
	logger.WithFields(logrus.Fields{
		"function": "handleStartQuest",
		"quest_id": req.Quest.ID,
	}).Debug("exiting handleStartQuest")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.Quest.ID,
		"message":  "Quest started successfully",
	}, nil
}

// handleCompleteQuest processes a request to complete a quest for a player.
// This handler validates quest completion criteria and processes rewards.
//
// Parameters:
//   - params: json.RawMessage containing the complete quest request with:
//   - session_id: string - The session ID of the requesting player
//   - quest_id: string - The ID of the quest to complete
//
// Returns:
//   - interface{}: Success response with rewards if quest completed successfully
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest not found or not completable
//   - Quest objectives not fulfilled
func (s *RPCServer) handleCompleteQuest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleCompleteQuest",
	})
	logger.Debug("entering handleCompleteQuest")

	req, err := s.parseCompleteQuestRequest(params)
	if err != nil {
		s.recordActionMetrics("complete_quest", err)
		return nil, err
	}

	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithField("session_id", req.SessionID).Error("failed to get player session")
		s.recordActionMetrics("complete_quest", err)
		return nil, fmt.Errorf("session error: %w", err)
	}

	rewards, err := session.Player.CompleteQuest(req.QuestID)
	if err != nil {
		logger.WithError(err).WithField("quest_id", req.QuestID).Error("failed to complete quest")
		s.recordActionMetrics("complete_quest", err)
		return nil, fmt.Errorf("failed to complete quest: %w", err)
	}

	if err := s.applyQuestRewards(session.Player, req.QuestID, rewards); err != nil {
		s.recordActionMetrics("complete_quest", err)
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"quest_id":     req.QuestID,
		"reward_count": len(rewards),
	}).Info("quest completed and all rewards applied")

	s.recordActionMetrics("complete_quest", nil)
	logger.WithField("quest_id", req.QuestID).Debug("exiting handleCompleteQuest")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.QuestID,
		"rewards":  rewards,
		"message":  "Quest completed successfully",
	}, nil
}

// completeQuestRequest defines the structure for a complete quest request.
type completeQuestRequest struct {
	SessionID string `json:"session_id"`
	QuestID   string `json:"quest_id"`
}

// parseCompleteQuestRequest parses the JSON request for completing a quest.
func (s *RPCServer) parseCompleteQuestRequest(params json.RawMessage) (*completeQuestRequest, error) {
	var req completeQuestRequest
	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function": "parseCompleteQuestRequest",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}
	return &req, nil
}

// applyQuestRewards processes and applies all rewards for a completed quest.
func (s *RPCServer) applyQuestRewards(player *game.Player, questID string, rewards []game.QuestReward) error {
	for _, reward := range rewards {
		var err error
		switch reward.Type {
		case "exp":
			err = s.applyExperienceReward(player, questID, reward)
		case "gold":
			s.applyGoldReward(player, questID, reward)
		case "item":
			err = s.applyItemReward(player, questID, reward)
		default:
			logrus.WithFields(logrus.Fields{
				"function":    "applyQuestRewards",
				"quest_id":    questID,
				"reward_type": reward.Type,
			}).Warn("unknown reward type, skipping")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// applyExperienceReward applies an experience reward to the player.
func (s *RPCServer) applyExperienceReward(player *game.Player, questID string, reward game.QuestReward) error {
	logger := logrus.WithFields(logrus.Fields{
		"function":    "applyExperienceReward",
		"quest_id":    questID,
		"reward_type": "exp",
		"value":       reward.Value,
	})

	if err := player.AddExperience(int64(reward.Value)); err != nil {
		logger.WithError(err).Error("failed to apply experience reward")
		return fmt.Errorf("failed to apply experience reward: %w", err)
	}
	logger.Info("applied experience reward")
	return nil
}

// applyGoldReward applies a gold reward to the player.
func (s *RPCServer) applyGoldReward(player *game.Player, questID string, reward game.QuestReward) {
	previousGold := player.Character.Gold
	player.Character.Gold += reward.Value
	logrus.WithFields(logrus.Fields{
		"function":      "applyGoldReward",
		"quest_id":      questID,
		"gold_added":    reward.Value,
		"previous_gold": previousGold,
		"new_gold":      player.Character.Gold,
	}).Info("applied gold reward")
}

// applyItemReward applies an item reward to the player.
func (s *RPCServer) applyItemReward(player *game.Player, questID string, reward game.QuestReward) error {
	if reward.ItemID == "" {
		return nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"function":    "applyItemReward",
		"quest_id":    questID,
		"reward_type": "item",
		"item_id":     reward.ItemID,
	})

	item := game.Item{
		ID:   reward.ItemID,
		Name: reward.ItemID, // Basic implementation - could be enhanced with item lookup
		Type: "quest_reward",
	}
	if err := player.Character.AddItemToInventory(item); err != nil {
		logger.WithError(err).Error("failed to apply item reward")
		return fmt.Errorf("failed to apply item reward: %w", err)
	}
	logger.Info("applied item reward")
	return nil
}

// handleUpdateObjective processes a request to update quest objective progress.
// This handler validates the objective update and tracks completion.
//
// Parameters:
//   - params: json.RawMessage containing the update objective request with:
//   - session_id: string - The session ID of the requesting player
//   - quest_id: string - The ID of the quest containing the objective
//   - objective_index: int - The index of the objective to update (0-based)
//   - progress: int - The new progress value for the objective
//
// Returns:
//   - interface{}: Success response with updated objective status
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest not found or not active
//   - Invalid objective index
func (s *RPCServer) handleUpdateObjective(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleUpdateObjective",
	})
	logger.Debug("entering handleUpdateObjective")

	var req struct {
		SessionID      string `json:"session_id"`
		QuestID        string `json:"quest_id"`
		ObjectiveIndex int    `json:"objective_index"`
		Progress       int    `json:"progress"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleUpdateObjective",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleUpdateObjective",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Update quest objective for player
	if err := session.Player.UpdateQuestObjective(req.QuestID, req.ObjectiveIndex, req.Progress); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":        "handleUpdateObjective",
			"quest_id":        req.QuestID,
			"objective_index": req.ObjectiveIndex,
		}).Error("failed to update quest objective")
		return nil, fmt.Errorf("failed to update quest objective: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"function":        "handleUpdateObjective",
		"quest_id":        req.QuestID,
		"objective_index": req.ObjectiveIndex,
		"progress":        req.Progress,
	}).Debug("exiting handleUpdateObjective")

	return map[string]interface{}{
		"success":         true,
		"quest_id":        req.QuestID,
		"objective_index": req.ObjectiveIndex,
		"progress":        req.Progress,
		"message":         "Quest objective updated successfully",
	}, nil
}

// handleFailQuest processes a request to fail a quest for a player.
// This handler marks the quest as failed, preventing completion.
//
// Parameters:
//   - params: json.RawMessage containing the fail quest request with:
//   - session_id: string - The session ID of the requesting player
//   - quest_id: string - The ID of the quest to fail
//
// Returns:
//   - interface{}: Success response confirming quest failure
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest not found or already completed/failed
func (s *RPCServer) handleFailQuest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleFailQuest",
	})
	logger.Debug("entering handleFailQuest")

	var req struct {
		SessionID string `json:"session_id"`
		QuestID   string `json:"quest_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleFailQuest",
		}).Error("failed to unmarshal request parameters")
		s.recordActionMetrics("fail_quest", err)
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleFailQuest",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		s.recordActionMetrics("fail_quest", err)
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Fail quest for player
	if err := session.Player.FailQuest(req.QuestID); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleFailQuest",
			"quest_id": req.QuestID,
		}).Error("failed to fail quest")
		s.recordActionMetrics("fail_quest", err)
		return nil, fmt.Errorf("failed to fail quest: %w", err)
	}

	s.recordActionMetrics("fail_quest", nil)
	logger.WithFields(logrus.Fields{
		"function": "handleFailQuest",
		"quest_id": req.QuestID,
	}).Debug("exiting handleFailQuest")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.QuestID,
		"message":  "Quest failed successfully",
	}, nil
}

// handleGetQuest processes a request to retrieve a specific quest from a player's quest log.
// This handler returns quest details including objectives and current status.
//
// Parameters:
//   - params: json.RawMessage containing the get quest request with:
//   - session_id: string - The session ID of the requesting player
//   - quest_id: string - The ID of the quest to retrieve
//
// Returns:
//   - interface{}: Quest data with full details
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
//   - Quest not found in player's quest log
func (s *RPCServer) handleGetQuest(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetQuest",
	})
	logger.Debug("entering handleGetQuest")

	var req struct {
		SessionID string `json:"session_id"`
		QuestID   string `json:"quest_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetQuest",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleGetQuest",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Get quest from player
	quest, err := session.Player.GetQuest(req.QuestID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetQuest",
			"quest_id": req.QuestID,
		}).Error("failed to get quest")
		return nil, fmt.Errorf("failed to get quest: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"function": "handleGetQuest",
		"quest_id": req.QuestID,
	}).Debug("exiting handleGetQuest")

	return map[string]interface{}{
		"success": true,
		"quest":   quest,
	}, nil
}

// handleGetActiveQuests processes a request to retrieve all active quests for a player.
// This handler returns a list of quests that are currently in progress.
//
// Parameters:
//   - params: json.RawMessage containing the get active quests request with:
//   - session_id: string - The session ID of the requesting player
//
// Returns:
//   - interface{}: Array of active quest data
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
func (s *RPCServer) handleGetActiveQuests(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetActiveQuests",
	})
	logger.Debug("entering handleGetActiveQuests")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetActiveQuests",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleGetActiveQuests",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Get active quests from player
	activeQuests := session.Player.GetActiveQuests()

	logger.WithFields(logrus.Fields{
		"function":    "handleGetActiveQuests",
		"quest_count": len(activeQuests),
	}).Debug("exiting handleGetActiveQuests")

	return map[string]interface{}{
		"success":       true,
		"active_quests": activeQuests,
		"count":         len(activeQuests),
	}, nil
}

// handleGetCompletedQuests processes a request to retrieve all completed quests for a player.
// This handler returns a list of quests that have been successfully finished.
//
// Parameters:
//   - params: json.RawMessage containing the get completed quests request with:
//   - session_id: string - The session ID of the requesting player
//
// Returns:
//   - interface{}: Array of completed quest data
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
func (s *RPCServer) handleGetCompletedQuests(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetCompletedQuests",
	})
	logger.Debug("entering handleGetCompletedQuests")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetCompletedQuests",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleGetCompletedQuests",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Get completed quests from player
	completedQuests := session.Player.GetCompletedQuests()

	logger.WithFields(logrus.Fields{
		"function":    "handleGetCompletedQuests",
		"quest_count": len(completedQuests),
	}).Debug("exiting handleGetCompletedQuests")

	return map[string]interface{}{
		"success":          true,
		"completed_quests": completedQuests,
		"count":            len(completedQuests),
	}, nil
}

// handleGetQuestLog processes a request to retrieve the complete quest log for a player.
// This handler returns all quests regardless of status (active, completed, failed).
//
// Parameters:
//   - params: json.RawMessage containing the get quest log request with:
//   - session_id: string - The session ID of the requesting player
//
// Returns:
//   - interface{}: Complete quest log with all quest data
//   - error: Error if request fails due to:
//   - Invalid request parameters
//   - Session not found or inactive
func (s *RPCServer) handleGetQuestLog(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function": "handleGetQuestLog",
	})
	logger.Debug("entering handleGetQuestLog")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function": "handleGetQuestLog",
		}).Error("failed to unmarshal request parameters")
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"function":   "handleGetQuestLog",
			"session_id": req.SessionID,
		}).Error("failed to get player session")
		return nil, fmt.Errorf("session error: %w", err)
	}

	// Get complete quest log from player
	questLog := session.Player.GetQuestLog()

	logger.WithFields(logrus.Fields{
		"function":    "handleGetQuestLog",
		"quest_count": len(questLog),
	}).Debug("exiting handleGetQuestLog")

	return map[string]interface{}{
		"success":   true,
		"quest_log": questLog,
		"count":     len(questLog),
	}, nil
}
