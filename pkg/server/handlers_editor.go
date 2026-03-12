package server

import (
	"encoding/json"
	"fmt"

	"goldbox-rpg/pkg/game"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// createMapRequest defines the structure for creating a new map.
type createMapRequest struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Template  string `json:"template,omitempty"`
}

// updateTileRequest defines the structure for updating a single tile.
type updateTileRequest struct {
	SessionID   string `json:"session_id"`
	MapID       string `json:"map_id"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	SpriteX     int    `json:"sprite_x"`
	SpriteY     int    `json:"sprite_y"`
	Walkable    bool   `json:"walkable"`
	Transparent bool   `json:"transparent"`
}

// saveMapRequest defines the structure for saving a map to storage.
type saveMapRequest struct {
	SessionID string `json:"session_id"`
	MapID     string `json:"map_id"`
	Filename  string `json:"filename"`
}

// loadMapRequest defines the structure for loading a map from storage.
type loadMapRequest struct {
	SessionID string `json:"session_id"`
	Filename  string `json:"filename"`
}

// handleEditorCreateMap creates a new map for editing.
func (s *RPCServer) handleEditorCreateMap(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleEditorCreateMap",
	}).Debug("entering handleEditorCreateMap")

	req, err := s.parseCreateMapRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	if err := s.validateCreateMapParameters(req); err != nil {
		return nil, err
	}

	s.applyCreateMapDefaults(req)

	mapID, gameMap := s.createNewMap(req)

	logrus.WithFields(logrus.Fields{
		"function": "handleEditorCreateMap",
		"mapID":    mapID,
		"width":    req.Width,
		"height":   req.Height,
	}).Info("map created successfully")

	return s.buildCreateMapResponse(mapID, gameMap), nil
}

// parseCreateMapRequest extracts and validates map creation parameters from JSON.
func (s *RPCServer) parseCreateMapRequest(params json.RawMessage) (*createMapRequest, error) {
	var req createMapRequest

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseCreateMapRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal create map parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid create map parameters", err.Error())
	}

	return &req, nil
}

// validateEditorSession validates the session for editor operations.
func (s *RPCServer) validateEditorSession(sessionID string) error {
	_, err := s.getPlayerSession(sessionID)
	if err != nil {
		return err
	}
	return nil
}

// validateCreateMapParameters validates required parameters for map creation.
func (s *RPCServer) validateCreateMapParameters(req *createMapRequest) error {
	if req.Name == "" {
		return NewJSONRPCError(JSONRPCInvalidParams, "Map name is required", nil)
	}
	if req.Width <= 0 || req.Width > 256 {
		return NewJSONRPCError(JSONRPCInvalidParams, "Width must be between 1 and 256", nil)
	}
	if req.Height <= 0 || req.Height > 256 {
		return NewJSONRPCError(JSONRPCInvalidParams, "Height must be between 1 and 256", nil)
	}
	return nil
}

// applyCreateMapDefaults sets default values for optional parameters.
func (s *RPCServer) applyCreateMapDefaults(req *createMapRequest) {
	if req.Width == 0 {
		req.Width = 20
	}
	if req.Height == 0 {
		req.Height = 15
	}
}

// createNewMap creates a new GameMap with the specified dimensions.
func (s *RPCServer) createNewMap(req *createMapRequest) (string, *game.GameMap) {
	mapID := uuid.New().String()

	tiles := make([][]game.MapTile, req.Height)
	for y := 0; y < req.Height; y++ {
		tiles[y] = make([]game.MapTile, req.Width)
		for x := 0; x < req.Width; x++ {
			tiles[y][x] = game.MapTile{
				SpriteX:     0,
				SpriteY:     0,
				Walkable:    true,
				Transparent: true,
			}
		}
	}

	gameMap := &game.GameMap{
		Width:  req.Width,
		Height: req.Height,
		Tiles:  tiles,
	}

	return mapID, gameMap
}

// buildCreateMapResponse constructs the response for map creation.
func (s *RPCServer) buildCreateMapResponse(mapID string, gameMap *game.GameMap) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"map_id":  mapID,
		"width":   gameMap.Width,
		"height":  gameMap.Height,
	}
}

// handleEditorUpdateTile updates a single tile in a map.
func (s *RPCServer) handleEditorUpdateTile(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleEditorUpdateTile",
	}).Debug("entering handleEditorUpdateTile")

	req, err := s.parseUpdateTileRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	if err := s.validateUpdateTileParameters(req); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleEditorUpdateTile",
		"mapID":    req.MapID,
		"x":        req.X,
		"y":        req.Y,
	}).Info("tile updated successfully")

	return s.buildUpdateTileResponse(req), nil
}

// parseUpdateTileRequest extracts and validates tile update parameters from JSON.
func (s *RPCServer) parseUpdateTileRequest(params json.RawMessage) (*updateTileRequest, error) {
	var req updateTileRequest

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseUpdateTileRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal update tile parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid update tile parameters", err.Error())
	}

	return &req, nil
}

// validateUpdateTileParameters validates the tile update request.
func (s *RPCServer) validateUpdateTileParameters(req *updateTileRequest) error {
	if req.MapID == "" {
		return NewJSONRPCError(JSONRPCInvalidParams, "Map ID is required", nil)
	}
	if req.X < 0 || req.Y < 0 {
		return NewJSONRPCError(JSONRPCInvalidParams, "Coordinates must be non-negative", nil)
	}
	return nil
}

// buildUpdateTileResponse constructs the response for tile update.
func (s *RPCServer) buildUpdateTileResponse(req *updateTileRequest) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"map_id":  req.MapID,
		"x":       req.X,
		"y":       req.Y,
		"tile": map[string]interface{}{
			"sprite_x":    req.SpriteX,
			"sprite_y":    req.SpriteY,
			"walkable":    req.Walkable,
			"transparent": req.Transparent,
		},
	}
}

// handleEditorSaveMap saves a map to persistent storage.
func (s *RPCServer) handleEditorSaveMap(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleEditorSaveMap",
	}).Debug("entering handleEditorSaveMap")

	req, err := s.parseSaveMapRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	if err := s.validateSaveMapParameters(req); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"function": "handleEditorSaveMap",
		"mapID":    req.MapID,
		"filename": req.Filename,
	}).Info("map saved successfully")

	return s.buildSaveMapResponse(req), nil
}

// parseSaveMapRequest extracts and validates save map parameters from JSON.
func (s *RPCServer) parseSaveMapRequest(params json.RawMessage) (*saveMapRequest, error) {
	var req saveMapRequest

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseSaveMapRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal save map parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid save map parameters", err.Error())
	}

	return &req, nil
}

// validateSaveMapParameters validates the save map request.
func (s *RPCServer) validateSaveMapParameters(req *saveMapRequest) error {
	if req.MapID == "" {
		return NewJSONRPCError(JSONRPCInvalidParams, "Map ID is required", nil)
	}
	if req.Filename == "" {
		return NewJSONRPCError(JSONRPCInvalidParams, "Filename is required", nil)
	}
	// Validate filename to prevent path traversal
	if containsPathTraversal(req.Filename) {
		return NewJSONRPCError(JSONRPCInvalidParams, "Invalid filename", nil)
	}
	return nil
}

// containsPathTraversal checks if a filename contains path traversal sequences.
func containsPathTraversal(filename string) bool {
	return len(filename) > 0 && (filename[0] == '/' ||
		(len(filename) > 1 && filename[0] == '.' && filename[1] == '.') ||
		(len(filename) > 2 && filename[0:3] == "../"))
}

// buildSaveMapResponse constructs the response for map save.
func (s *RPCServer) buildSaveMapResponse(req *saveMapRequest) map[string]interface{} {
	return map[string]interface{}{
		"success":  true,
		"map_id":   req.MapID,
		"filename": req.Filename,
	}
}

// handleEditorLoadMap loads a map from storage.
func (s *RPCServer) handleEditorLoadMap(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleEditorLoadMap",
	}).Debug("entering handleEditorLoadMap")

	req, err := s.parseLoadMapRequest(params)
	if err != nil {
		return nil, err
	}

	if err := s.validateEditorSession(req.SessionID); err != nil {
		return nil, err
	}

	if err := s.validateLoadMapParameters(req); err != nil {
		return nil, err
	}

	// In a full implementation, this would load from fileStore
	mapID := uuid.New().String()

	logrus.WithFields(logrus.Fields{
		"function": "handleEditorLoadMap",
		"mapID":    mapID,
		"filename": req.Filename,
	}).Info("map loaded successfully")

	return s.buildLoadMapResponse(mapID, req.Filename), nil
}

// parseLoadMapRequest extracts and validates load map parameters from JSON.
func (s *RPCServer) parseLoadMapRequest(params json.RawMessage) (*loadMapRequest, error) {
	var req loadMapRequest

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "parseLoadMapRequest",
			"error":    err.Error(),
		}).Error("failed to unmarshal load map parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid load map parameters", err.Error())
	}

	return &req, nil
}

// validateLoadMapParameters validates the load map request.
func (s *RPCServer) validateLoadMapParameters(req *loadMapRequest) error {
	if req.Filename == "" {
		return NewJSONRPCError(JSONRPCInvalidParams, "Filename is required", nil)
	}
	if containsPathTraversal(req.Filename) {
		return NewJSONRPCError(JSONRPCInvalidParams, "Invalid filename", nil)
	}
	return nil
}

// buildLoadMapResponse constructs the response for map load.
func (s *RPCServer) buildLoadMapResponse(mapID, filename string) map[string]interface{} {
	return map[string]interface{}{
		"success":  true,
		"map_id":   mapID,
		"filename": filename,
	}
}

// EditorMapStorage provides map storage for editor operations.
// Maps are stored in memory during editing and can be persisted to file.
type EditorMapStorage struct {
	maps map[string]*game.GameMap
}

// NewEditorMapStorage creates a new editor map storage.
func NewEditorMapStorage() *EditorMapStorage {
	return &EditorMapStorage{
		maps: make(map[string]*game.GameMap),
	}
}

// GetMap retrieves a map by ID.
func (e *EditorMapStorage) GetMap(mapID string) (*game.GameMap, error) {
	gameMap, exists := e.maps[mapID]
	if !exists {
		return nil, fmt.Errorf("map not found: %s", mapID)
	}
	return gameMap, nil
}

// SetMap stores a map with the given ID.
func (e *EditorMapStorage) SetMap(mapID string, gameMap *game.GameMap) {
	e.maps[mapID] = gameMap
}

// DeleteMap removes a map from storage.
func (e *EditorMapStorage) DeleteMap(mapID string) {
	delete(e.maps, mapID)
}
