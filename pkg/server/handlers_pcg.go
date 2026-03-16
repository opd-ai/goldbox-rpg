package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
)

// contentGenerationRequest defines the structure for content generation requests.
type contentGenerationRequest struct {
	SessionID   string                 `json:"session_id"`
	ContentType string                 `json:"content_type"`
	LocationID  string                 `json:"location_id"`
	Difficulty  int                    `json:"difficulty"`
	Constraints map[string]interface{} `json:"constraints"`
}

// handleGenerateContent generates procedural content on demand.
func (s *RPCServer) handleGenerateContent(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGenerateContent",
	}).Debug("entering handleGenerateContent")

	req, err := s.parseContentGenerationRequest(params)
	if err != nil {
		return nil, err
	}

	session, err := s.validateContentGenerationSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	if err := s.validateContentGenerationParameters(req); err != nil {
		return nil, err
	}

	s.applyContentGenerationDefaults(req)

	content, err := s.executeContentGeneration(req)
	if err != nil {
		return nil, err
	}

	s.logContentGenerationSuccess(req)

	return s.buildContentGenerationResponse(req, content), nil
}

// parseContentGenerationRequest extracts and validates content generation parameters from JSON.
func (s *RPCServer) parseContentGenerationRequest(params json.RawMessage) (*contentGenerationRequest, error) {
	var req contentGenerationRequest

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseContentGenerationRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal content generation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid content generation parameters", err.Error())
	}

	return &req, nil
}

// validateContentGenerationSession retrieves and validates the player session for content generation.
func (s *RPCServer) validateContentGenerationSession(sessionID string) (*PlayerSession, error) {
	session, err := s.getPlayerSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// validateContentGenerationParameters checks that required content generation parameters are present.
func (s *RPCServer) validateContentGenerationParameters(req *contentGenerationRequest) error {
	if req.ContentType == "" {
		return fmt.Errorf("content_type parameter required")
	}

	if req.LocationID == "" {
		return fmt.Errorf("location_id parameter required")
	}

	return nil
}

// applyContentGenerationDefaults sets default values for optional content generation parameters.
func (s *RPCServer) applyContentGenerationDefaults(req *contentGenerationRequest) {
	if req.Difficulty == 0 {
		req.Difficulty = 5 // Default difficulty
	}
}

// executeContentGeneration performs the actual content generation based on content type.
func (s *RPCServer) executeContentGeneration(req *contentGenerationRequest) (interface{}, error) {
	ctx := context.Background()
	var content interface{}
	var err error

	switch pcg.ContentType(strings.ToLower(req.ContentType)) {
	case pcg.ContentTypeTerrain:
		content, err = s.pcgManager.GenerateTerrainForLevel(ctx, req.LocationID, 50, 50, pcg.BiomeDungeon, req.Difficulty)
	case pcg.ContentTypeItems:
		content, err = s.pcgManager.GenerateItemsForLocation(ctx, req.LocationID, 3, pcg.RarityCommon, pcg.RarityRare, req.Difficulty)
	case pcg.ContentTypeLevels:
		content, err = s.pcgManager.GenerateDungeonLevel(ctx, req.LocationID, 5, 15, pcg.ThemeClassic, req.Difficulty)
	case pcg.ContentTypeQuests:
		content, err = s.pcgManager.GenerateQuestForArea(ctx, req.LocationID, pcg.QuestTypeFetch, req.Difficulty)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", req.ContentType)
	}

	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return content, nil
}

// logContentGenerationSuccess logs successful content generation with relevant details.
func (s *RPCServer) logContentGenerationSuccess(req *contentGenerationRequest) {
	logrus.WithFields(logrus.Fields{
		"function":    "executeContentGeneration",
		"sessionID":   req.SessionID,
		"contentType": req.ContentType,
		"locationID":  req.LocationID,
		"difficulty":  req.Difficulty,
	}).Info("content generated successfully")
}

// buildContentGenerationResponse constructs the response map for successful content generation.
func (s *RPCServer) buildContentGenerationResponse(req *contentGenerationRequest, content interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success":      true,
		"content_type": req.ContentType,
		"location_id":  req.LocationID,
		"content":      content,
		"difficulty":   req.Difficulty,
	}
}

// terrainRegenerationRequest defines the structure for terrain regeneration requests.
type terrainRegenerationRequest struct {
	SessionID    string  `json:"session_id"`
	LocationID   string  `json:"location_id"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	BiomeType    string  `json:"biome_type"`
	Density      float64 `json:"density"`
	WaterLevel   float64 `json:"water_level"`
	Connectivity string  `json:"connectivity"`
}

// parseTerrainRegenerationRequest extracts and validates terrain regeneration parameters from JSON.
func (s *RPCServer) parseTerrainRegenerationRequest(params json.RawMessage) (*terrainRegenerationRequest, error) {
	var req terrainRegenerationRequest

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseTerrainRegenerationRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal terrain regeneration parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid terrain parameters", err.Error())
	}

	return &req, nil
}

// validateTerrainRegenerationRequest validates required parameters and session.
func (s *RPCServer) validateTerrainRegenerationRequest(req *terrainRegenerationRequest) error {
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return err
	}
	_ = session // Suppress unused variable warning

	if req.LocationID == "" {
		return fmt.Errorf("location_id parameter required")
	}

	return nil
}

// applyTerrainRegenerationDefaults sets default values for empty request fields.
func (s *RPCServer) applyTerrainRegenerationDefaults(req *terrainRegenerationRequest) {
	if req.Width == 0 {
		req.Width = 50
	}
	if req.Height == 0 {
		req.Height = 50
	}
	if req.BiomeType == "" {
		req.BiomeType = "forest"
	}
	if req.Density == 0 {
		req.Density = 0.5
	}
	if req.WaterLevel == 0 {
		req.WaterLevel = 0.3
	}
	if req.Connectivity == "" {
		req.Connectivity = "moderate"
	}
}

// executeTerrainGeneration performs the actual terrain generation using the PCG manager.
func (s *RPCServer) executeTerrainGeneration(req *terrainRegenerationRequest) (interface{}, error) {
	ctx := context.Background()
	biomeType := pcg.BiomeType(strings.ToLower(req.BiomeType))

	gameMap, err := s.pcgManager.GenerateTerrainForLevel(ctx, req.LocationID, req.Width, req.Height, biomeType, 5)
	if err != nil {
		return nil, fmt.Errorf("terrain generation failed: %w", err)
	}

	return gameMap, nil
}

// logTerrainRegenerationSuccess logs successful terrain generation with relevant details.
func (s *RPCServer) logTerrainRegenerationSuccess(req *terrainRegenerationRequest) {
	logrus.WithFields(logrus.Fields{
		"function":   "executeTerrainGeneration",
		"sessionID":  req.SessionID,
		"locationID": req.LocationID,
		"width":      req.Width,
		"height":     req.Height,
		"biomeType":  req.BiomeType,
	}).Info("terrain regenerated successfully")
}

// buildTerrainRegenerationResponse constructs the response map for successful terrain generation.
func (s *RPCServer) buildTerrainRegenerationResponse(req *terrainRegenerationRequest, terrain interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success":     true,
		"location_id": req.LocationID,
		"terrain":     terrain,
		"width":       req.Width,
		"height":      req.Height,
		"biome_type":  req.BiomeType,
	}
}

// handleRegenerateTerrain regenerates terrain for a specific area.
func (s *RPCServer) handleRegenerateTerrain(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleRegenerateTerrain",
	}).Debug("entering handleRegenerateTerrain")

	req, err := s.parseTerrainRegenerationRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.validateTerrainRegenerationRequest(req); err != nil {
		return nil, err
	}

	s.applyTerrainRegenerationDefaults(req)

	terrain, err := s.executeTerrainGeneration(req)
	if err != nil {
		return nil, err
	}

	s.logTerrainRegenerationSuccess(req)

	return s.buildTerrainRegenerationResponse(req, terrain), nil
}

// handleGenerateItems generates items for a location.
func (s *RPCServer) handleGenerateItems(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGenerateItems",
	}).Debug("entering handleGenerateItems")

	var req struct {
		SessionID   string   `json:"session_id"`
		LocationID  string   `json:"location_id"`
		Count       int      `json:"count"`
		MinRarity   string   `json:"min_rarity"`
		MaxRarity   string   `json:"max_rarity"`
		PlayerLevel int      `json:"player_level"`
		ItemTypes   []string `json:"item_types"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGenerateItems",
			"error":    err.Error(),
		}).Error("failed to unmarshal item generation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid item generation parameters", err.Error())
	}
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	if req.LocationID == "" {
		return nil, fmt.Errorf("location_id parameter required")
	}

	// Set defaults
	if req.Count == 0 {
		req.Count = 3
	}
	if req.MinRarity == "" {
		req.MinRarity = "common"
	}
	if req.MaxRarity == "" {
		req.MaxRarity = "rare"
	}
	if req.PlayerLevel == 0 {
		req.PlayerLevel = 5
	}

	ctx := context.Background()

	// Convert rarity strings to PCG RarityTier
	minRarity := pcg.RarityTier(strings.ToLower(req.MinRarity))
	maxRarity := pcg.RarityTier(strings.ToLower(req.MaxRarity))

	items, err := s.pcgManager.GenerateItemsForLocation(ctx, req.LocationID, req.Count, minRarity, maxRarity, req.PlayerLevel)
	if err != nil {
		return nil, fmt.Errorf("item generation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":       "handleGenerateItems",
		"sessionID":      req.SessionID,
		"locationID":     req.LocationID,
		"count":          req.Count,
		"playerLevel":    req.PlayerLevel,
		"itemsGenerated": len(items),
	}).Info("items generated successfully")

	return map[string]interface{}{
		"success":     true,
		"location_id": req.LocationID,
		"items":       items,
		"count":       len(items),
	}, nil
}

// levelGenerationRequest represents the request structure for level generation.
type levelGenerationRequest struct {
	SessionID     string `json:"session_id"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	RoomCount     int    `json:"room_count"`
	Theme         string `json:"theme"`
	Difficulty    int    `json:"difficulty"`
	CorridorStyle string `json:"corridor_style"`
}

// parseLevelGenerationRequest unmarshals and validates the level generation request parameters.
func (s *RPCServer) parseLevelGenerationRequest(params json.RawMessage) (*levelGenerationRequest, error) {
	var req levelGenerationRequest
	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseLevelGenerationRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal level generation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid level generation parameters", err.Error())
	}
	return &req, nil
}

// validateLevelGenerationSession retrieves and validates the player session for level generation.
func (s *RPCServer) validateLevelGenerationSession(sessionID string) error {
	session, err := s.getPlayerSession(sessionID)
	if err != nil {
		return err
	}
	_ = session // Suppress unused variable warning
	return nil
}

// applyLevelGenerationDefaults sets default values for level generation parameters.
func (s *RPCServer) applyLevelGenerationDefaults(req *levelGenerationRequest) {
	if req.Width == 0 {
		req.Width = 50
	}
	if req.Height == 0 {
		req.Height = 50
	}
	if req.RoomCount == 0 {
		req.RoomCount = 8
	}
	if req.Theme == "" {
		req.Theme = "classic"
	}
	if req.Difficulty == 0 {
		req.Difficulty = 5
	}
	if req.CorridorStyle == "" {
		req.CorridorStyle = "straight"
	}
}

// executeLevelGeneration performs the actual level generation using PCG manager.
func (s *RPCServer) executeLevelGeneration(req *levelGenerationRequest) (interface{}, error) {
	ctx := context.Background()
	theme := pcg.LevelTheme(strings.ToLower(req.Theme))

	level, err := s.pcgManager.GenerateDungeonLevel(ctx, "generated_level", 5, req.RoomCount, theme, req.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("level generation failed: %w", err)
	}

	return level, nil
}

// buildLevelGenerationResponse constructs the success response for level generation.
func (s *RPCServer) buildLevelGenerationResponse(req *levelGenerationRequest, level interface{}) map[string]interface{} {
	logrus.WithFields(logrus.Fields{
		"function":   "handleGenerateLevel",
		"sessionID":  req.SessionID,
		"width":      req.Width,
		"height":     req.Height,
		"roomCount":  req.RoomCount,
		"theme":      req.Theme,
		"difficulty": req.Difficulty,
	}).Info("level generated successfully")

	return map[string]interface{}{
		"success":        true,
		"level":          level,
		"width":          req.Width,
		"height":         req.Height,
		"room_count":     req.RoomCount,
		"theme":          req.Theme,
		"difficulty":     req.Difficulty,
		"corridor_style": req.CorridorStyle,
	}
}

// handleGenerateLevel generates a complete level/dungeon.
func (s *RPCServer) handleGenerateLevel(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGenerateLevel",
	}).Debug("entering handleGenerateLevel")

	req, err := s.parseLevelGenerationRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.validateLevelGenerationSession(req.SessionID); err != nil {
		return nil, err
	}

	s.applyLevelGenerationDefaults(req)

	level, err := s.executeLevelGeneration(req)
	if err != nil {
		return nil, err
	}

	return s.buildLevelGenerationResponse(req, level), nil
}

// generateQuestRequest represents the request structure for quest generation.
type generateQuestRequest struct {
	SessionID     string `json:"session_id"`
	QuestType     string `json:"quest_type"`
	Difficulty    int    `json:"difficulty"`
	MinObjectives int    `json:"min_objectives"`
	MaxObjectives int    `json:"max_objectives"`
	RewardTier    string `json:"reward_tier"`
	NarrativeType string `json:"narrative_type"`
}

// handleGenerateQuest generates a procedural quest.
func (s *RPCServer) handleGenerateQuest(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGenerateQuest",
	}).Debug("entering handleGenerateQuest")

	req, err := s.parseQuestGenerationRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.validateQuestGenerationSession(req.SessionID); err != nil {
		return nil, err
	}

	s.applyQuestGenerationDefaults(req)

	quest, err := s.executeQuestGeneration(req)
	if err != nil {
		return nil, err
	}

	s.logQuestGenerationSuccess(req, quest)

	return s.buildQuestGenerationResponse(req, quest), nil
}

// parseQuestGenerationRequest parses and validates the JSON request parameters.
func (s *RPCServer) parseQuestGenerationRequest(params json.RawMessage) (*generateQuestRequest, error) {
	var req generateQuestRequest

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseQuestGenerationRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal quest generation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid quest generation parameters", err.Error())
	}

	return &req, nil
}

// validateQuestGenerationSession validates the session ID and retrieves the session.
func (s *RPCServer) validateQuestGenerationSession(sessionID string) error {
	session, err := s.getPlayerSession(sessionID)
	if err != nil {
		return err
	}
	_ = session // Session is valid but not used in current implementation
	return nil
}

// applyQuestGenerationDefaults sets default values for empty request fields.
func (s *RPCServer) applyQuestGenerationDefaults(req *generateQuestRequest) {
	if req.QuestType == "" {
		req.QuestType = "fetch"
	}
	if req.Difficulty == 0 {
		req.Difficulty = 5
	}
	if req.MinObjectives == 0 {
		req.MinObjectives = 1
	}
	if req.MaxObjectives == 0 {
		req.MaxObjectives = 3
	}
	if req.RewardTier == "" {
		req.RewardTier = "common"
	}
	if req.NarrativeType == "" {
		req.NarrativeType = "linear"
	}
}

// executeQuestGeneration performs the actual quest generation using the PCG manager.
func (s *RPCServer) executeQuestGeneration(req *generateQuestRequest) (*game.Quest, error) {
	ctx := context.Background()
	questType := pcg.QuestType(strings.ToLower(req.QuestType))

	quest, err := s.pcgManager.GenerateQuestForArea(ctx, "generated_quest_area", questType, req.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("quest generation failed: %w", err)
	}

	return quest, nil
}

// logQuestGenerationSuccess logs successful quest generation with relevant details.
func (s *RPCServer) logQuestGenerationSuccess(req *generateQuestRequest, quest *game.Quest) {
	logrus.WithFields(logrus.Fields{
		"function":       "executeQuestGeneration",
		"sessionID":      req.SessionID,
		"questType":      req.QuestType,
		"difficulty":     req.Difficulty,
		"objectiveCount": len(quest.Objectives),
	}).Info("quest generated successfully")
}

// buildQuestGenerationResponse constructs the response map for successful quest generation.
func (s *RPCServer) buildQuestGenerationResponse(req *generateQuestRequest, quest *game.Quest) map[string]interface{} {
	return map[string]interface{}{
		"success":        true,
		"quest":          quest,
		"quest_type":     req.QuestType,
		"difficulty":     req.Difficulty,
		"min_objectives": req.MinObjectives,
		"max_objectives": req.MaxObjectives,
		"reward_tier":    req.RewardTier,
		"narrative_type": req.NarrativeType,
	}
}

// handleGetPCGStats returns statistics about the PCG system.
func (s *RPCServer) handleGetPCGStats(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetPCGStats",
	}).Debug("entering handleGetPCGStats")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetPCGStats",
			"error":    err.Error(),
		}).Error("failed to unmarshal PCG stats parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid PCG stats parameters", err.Error())
	}

	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	// Get PCG statistics
	stats := s.pcgManager.GetGenerationStatistics()

	logrus.WithFields(logrus.Fields{
		"function":  "handleGetPCGStats",
		"sessionID": req.SessionID,
	}).Info("PCG stats retrieved successfully")

	return map[string]interface{}{
		"success": true,
		"stats":   stats,
	}, nil
}

// handleValidateContent validates generated content.
func (s *RPCServer) handleValidateContent(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleValidateContent",
	}).Debug("entering handleValidateContent")

	var req struct {
		SessionID   string      `json:"session_id"`
		ContentType string      `json:"content_type"`
		Content     interface{} `json:"content"`
		Strict      bool        `json:"strict"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleValidateContent",
			"error":    err.Error(),
		}).Error("failed to unmarshal content validation parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid content validation parameters", err.Error())
	}

	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}
	_ = session // Suppress unused variable warning

	if req.ContentType == "" {
		return nil, fmt.Errorf("content_type parameter required")
	}

	if req.Content == nil {
		return nil, fmt.Errorf("content parameter required")
	}

	// Validate content using PCG validator with type information
	validationResult, err := s.pcgManager.ValidateGeneratedContentWithType(req.Content, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function":         "handleValidateContent",
		"sessionID":        req.SessionID,
		"contentType":      req.ContentType,
		"validationResult": validationResult.IsValid(),
	}).Info("content validated successfully")

	return map[string]interface{}{
		"success":      true,
		"valid":        validationResult.IsValid(),
		"errors":       validationResult.Errors,
		"warnings":     validationResult.Warnings,
		"content_type": req.ContentType,
		"strict":       req.Strict,
	}, nil
}
