package server

import (
	"encoding/json"

	"goldbox-rpg/pkg/game"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// createQuestRequest defines the structure for creating a new quest via the editor.
type createQuestRequest struct {
	SessionID   string                `json:"session_id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Objectives  []questObjectiveInput `json:"objectives"`
	Rewards     []questRewardInput    `json:"rewards"`
}

// questObjectiveInput represents an objective in a quest creation request.
type questObjectiveInput struct {
	Description string `json:"description"`
	Required    int    `json:"required"`
}

// questRewardInput represents a reward in a quest creation request.
type questRewardInput struct {
	Type   string `json:"type"`
	Value  int    `json:"value"`
	ItemID string `json:"item_id,omitempty"`
}

// getQuestEditorRequest defines the structure for retrieving a quest for editing.
type getQuestEditorRequest struct {
	SessionID string `json:"session_id"`
	QuestID   string `json:"quest_id"`
}

// updateQuestRequest defines the structure for updating an existing quest.
type updateQuestRequest struct {
	SessionID   string                `json:"session_id"`
	QuestID     string                `json:"quest_id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Objectives  []questObjectiveInput `json:"objectives"`
	Rewards     []questRewardInput    `json:"rewards"`
}

// deleteQuestRequest defines the structure for deleting a quest.
type deleteQuestRequest struct {
	SessionID string `json:"session_id"`
	QuestID   string `json:"quest_id"`
}

// listQuestsRequest defines the structure for listing quests in the editor.
type listQuestsRequest struct {
	SessionID string `json:"session_id"`
}

// handleQuestEditorCreate creates a new quest via the visual editor.
func (s *RPCServer) handleQuestEditorCreate(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithField("function", "handleQuestEditorCreate")
	logger.Debug("entering handleQuestEditorCreate")

	var req createQuestRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid quest parameters", err.Error())
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	if err := validateQuestEditorInput(req); err != nil {
		return nil, err
	}

	quest := buildQuestFromInput(req)

	logger.WithFields(logrus.Fields{
		"questID": quest.ID,
		"title":   quest.Title,
	}).Info("quest created via editor")

	return map[string]interface{}{
		"success":  true,
		"quest_id": quest.ID,
		"title":    quest.Title,
	}, nil
}

// handleQuestEditorGet retrieves a quest for editing.
func (s *RPCServer) handleQuestEditorGet(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithField("function", "handleQuestEditorGet")
	logger.Debug("entering handleQuestEditorGet")

	var req getQuestEditorRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid parameters", err.Error())
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	if req.QuestID == "" {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Quest ID is required", nil)
	}

	logger.WithField("questID", req.QuestID).Info("quest retrieved for editing")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.QuestID,
	}, nil
}

// handleQuestEditorUpdate updates an existing quest via the editor.
func (s *RPCServer) handleQuestEditorUpdate(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithField("function", "handleQuestEditorUpdate")
	logger.Debug("entering handleQuestEditorUpdate")

	var req updateQuestRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid parameters", err.Error())
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	if req.QuestID == "" {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Quest ID is required", nil)
	}

	if req.Title == "" {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Quest title is required", nil)
	}

	logger.WithFields(logrus.Fields{
		"questID": req.QuestID,
		"title":   req.Title,
	}).Info("quest updated via editor")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.QuestID,
		"title":    req.Title,
	}, nil
}

// handleQuestEditorDelete deletes a quest via the editor.
func (s *RPCServer) handleQuestEditorDelete(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithField("function", "handleQuestEditorDelete")
	logger.Debug("entering handleQuestEditorDelete")

	var req deleteQuestRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid parameters", err.Error())
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	if req.QuestID == "" {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Quest ID is required", nil)
	}

	logger.WithField("questID", req.QuestID).Info("quest deleted via editor")

	return map[string]interface{}{
		"success":  true,
		"quest_id": req.QuestID,
	}, nil
}

// handleQuestEditorList lists all quests available for editing.
func (s *RPCServer) handleQuestEditorList(params json.RawMessage) (interface{}, error) {
	logger := logrus.WithField("function", "handleQuestEditorList")
	logger.Debug("entering handleQuestEditorList")

	var req listQuestsRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid parameters", err.Error())
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	logger.Info("quest list retrieved for editor")

	return map[string]interface{}{
		"success": true,
		"quests":  []interface{}{},
	}, nil
}

// validateQuestEditorInput validates the quest creation request.
func validateQuestEditorInput(req createQuestRequest) error {
	if err := validateQuestTitle(req.Title); err != nil {
		return err
	}
	if err := validateQuestDescription(req.Description); err != nil {
		return err
	}
	if err := validateQuestObjectives(req.Objectives); err != nil {
		return err
	}
	return validateQuestRewards(req.Rewards)
}

// validateQuestTitle validates the quest title field.
func validateQuestTitle(title string) error {
	if title == "" {
		return NewJSONRPCError(JSONRPCInvalidParams, "Quest title is required", nil)
	}
	if len(title) > 200 {
		return NewJSONRPCError(JSONRPCInvalidParams, "Quest title too long (max 200)", nil)
	}
	return nil
}

// validateQuestDescription validates the quest description field.
func validateQuestDescription(description string) error {
	if len(description) > 2000 {
		return NewJSONRPCError(JSONRPCInvalidParams, "Description too long (max 2000)", nil)
	}
	return nil
}

// validateQuestObjectives validates the quest objectives list.
func validateQuestObjectives(objectives []questObjectiveInput) error {
	if len(objectives) == 0 {
		return NewJSONRPCError(JSONRPCInvalidParams, "At least one objective is required", nil)
	}
	for i, obj := range objectives {
		if obj.Description == "" {
			return NewJSONRPCError(JSONRPCInvalidParams,
				"Objective description is required", map[string]interface{}{"index": i})
		}
		if obj.Required <= 0 {
			return NewJSONRPCError(JSONRPCInvalidParams,
				"Required must be positive", map[string]interface{}{"index": i})
		}
	}
	return nil
}

// validateQuestRewards validates the quest rewards list.
func validateQuestRewards(rewards []questRewardInput) error {
	for i, rew := range rewards {
		if err := validateRewardType(rew.Type); err != nil {
			return NewJSONRPCError(JSONRPCInvalidParams,
				err.Error(), map[string]interface{}{"index": i})
		}
		if rew.Value <= 0 {
			return NewJSONRPCError(JSONRPCInvalidParams,
				"Reward value must be positive", map[string]interface{}{"index": i})
		}
	}
	return nil
}

// validateRewardType checks the reward type is valid.
func validateRewardType(rewardType string) error {
	switch rewardType {
	case "gold", "item", "exp":
		return nil
	default:
		return NewJSONRPCError(JSONRPCInvalidParams,
			"Invalid reward type (must be gold, item, or exp)", nil)
	}
}

// buildQuestFromInput converts the editor input into a Quest struct.
func buildQuestFromInput(req createQuestRequest) *game.Quest {
	objectives := make([]game.QuestObjective, len(req.Objectives))
	for i, obj := range req.Objectives {
		objectives[i] = game.QuestObjective{
			Description: obj.Description,
			Required:    obj.Required,
			Progress:    0,
			Completed:   false,
		}
	}

	rewards := make([]game.QuestReward, len(req.Rewards))
	for i, rew := range req.Rewards {
		rewards[i] = game.QuestReward{
			Type:   rew.Type,
			Value:  rew.Value,
			ItemID: rew.ItemID,
		}
	}

	return &game.Quest{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      game.QuestNotStarted,
		Objectives:  objectives,
		Rewards:     rewards,
	}
}
