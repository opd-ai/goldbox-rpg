package server

import (
	"encoding/json"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// handleGetObjectsInRange processes a spatial query request for objects within a rectangular area.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - min_x: int minimum X coordinate of query rectangle
//   - min_y: int minimum Y coordinate of query rectangle
//   - max_x: int maximum X coordinate of query rectangle
//   - max_y: int maximum Y coordinate of query rectangle
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if query was successful
//   - objects: Array of game objects within the specified range
//   - count: Number of objects found
//   - error: Possible errors:
//   - "invalid range query parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
func (s *RPCServer) handleGetObjectsInRange(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetObjectsInRange",
	}).Debug("entering range query handler")

	var req struct {
		SessionID string `json:"session_id"`
		MinX      int    `json:"min_x"`
		MinY      int    `json:"min_y"`
		MaxX      int    `json:"max_x"`
		MaxY      int    `json:"max_y"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).Error("failed to unmarshal range query parameters")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid range query parameters",
		}, nil
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithField("sessionID", req.SessionID).Warn("range query attempted with invalid session")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid session",
		}, nil
	}
	defer s.releaseSession(session)

	if session.Player == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "session has no player",
		}, nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"sessionID": req.SessionID,
		"playerID":  session.Player.GetID(),
		"minX":      req.MinX,
		"minY":      req.MinY,
		"maxX":      req.MaxX,
		"maxY":      req.MaxY,
	})

	rect := game.Rectangle{
		MinX: req.MinX,
		MinY: req.MinY,
		MaxX: req.MaxX,
		MaxY: req.MaxY,
	}

	objects := s.state.WorldState.GetObjectsInRange(rect)
	logger.WithField("objectCount", len(objects)).Info("range query completed")

	return map[string]interface{}{
		"success": true,
		"objects": objects,
		"count":   len(objects),
	}, nil
}

// handleGetObjectsInRadius processes a spatial query request for objects within a circular area.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - center_x: int X coordinate of circle center
//   - center_y: int Y coordinate of query center
//   - radius: float64 radius of the search circle
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if query was successful
//   - objects: Array of game objects within the specified radius
//   - count: Number of objects found
//   - error: Possible errors:
//   - "invalid radius query parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
func (s *RPCServer) handleGetObjectsInRadius(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetObjectsInRadius",
	}).Debug("entering radius query handler")

	var req struct {
		SessionID string  `json:"session_id"`
		CenterX   int     `json:"center_x"`
		CenterY   int     `json:"center_y"`
		Radius    float64 `json:"radius"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).Error("failed to unmarshal radius query parameters")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid radius query parameters",
		}, nil
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithField("sessionID", req.SessionID).Warn("radius query attempted with invalid session")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid session",
		}, nil
	}
	defer s.releaseSession(session)

	if session.Player == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "session has no player",
		}, nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"sessionID": req.SessionID,
		"playerID":  session.Player.GetID(),
		"centerX":   req.CenterX,
		"centerY":   req.CenterY,
		"radius":    req.Radius,
	})

	center := game.Position{X: req.CenterX, Y: req.CenterY}
	objects := s.state.WorldState.GetObjectsInRadius(center, req.Radius)
	logger.WithField("objectCount", len(objects)).Info("radius query completed")

	return map[string]interface{}{
		"success": true,
		"objects": objects,
		"count":   len(objects),
	}, nil
}

// handleGetNearestObjects processes a spatial query request for the K nearest objects.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - center_x: int X coordinate of query center
//   - center_y: int Y coordinate of query center
//   - k: int number of nearest objects to return
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if query was successful
//   - objects: Array of K nearest game objects
//   - count: Number of objects found
//   - error: Possible errors:
//   - "invalid nearest query parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
func (s *RPCServer) handleGetNearestObjects(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetNearestObjects",
	}).Debug("entering nearest objects query handler")

	var req struct {
		SessionID string `json:"session_id"`
		CenterX   int    `json:"center_x"`
		CenterY   int    `json:"center_y"`
		K         int    `json:"k"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).Error("failed to unmarshal nearest query parameters")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid nearest query parameters",
		}, nil
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithField("sessionID", req.SessionID).Warn("nearest query attempted with invalid session")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid session",
		}, nil
	}
	defer s.releaseSession(session)

	if session.Player == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "session has no player",
		}, nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"sessionID": req.SessionID,
		"playerID":  session.Player.GetID(),
		"centerX":   req.CenterX,
		"centerY":   req.CenterY,
		"k":         req.K,
	})

	center := game.Position{X: req.CenterX, Y: req.CenterY}
	objects := s.state.WorldState.GetNearestObjects(center, req.K)
	logger.WithField("objectCount", len(objects)).Info("nearest objects query completed")

	return map[string]interface{}{
		"success": true,
		"objects": objects,
		"count":   len(objects),
	}, nil
}

// handleFindPath processes a pathfinding request using A* algorithm.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//   - start_x: int X coordinate of path start
//   - start_y: int Y coordinate of path start
//   - end_x: int X coordinate of path destination
//   - end_y: int Y coordinate of path destination
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if pathfinding succeeded
//   - path: Array of position objects representing the path
//   - path_length: Number of steps in the path
//   - found: bool indicating if a valid path was found
//   - error: Possible errors:
//   - "invalid pathfinding parameters" if JSON unmarshaling fails
//   - "invalid session" if session ID not found
func (s *RPCServer) handleFindPath(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleFindPath",
	}).Debug("entering pathfinding handler")

	var req struct {
		SessionID string `json:"session_id"`
		StartX    int    `json:"start_x"`
		StartY    int    `json:"start_y"`
		EndX      int    `json:"end_x"`
		EndY      int    `json:"end_y"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).Error("failed to unmarshal pathfinding parameters")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid pathfinding parameters",
		}, nil
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithField("sessionID", req.SessionID).Warn("pathfinding attempted with invalid session")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid session",
		}, nil
	}
	defer s.releaseSession(session)

	if session.Player == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "session has no player",
		}, nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"sessionID": req.SessionID,
		"playerID":  session.Player.GetID(),
		"startX":    req.StartX,
		"startY":    req.StartY,
		"endX":      req.EndX,
		"endY":      req.EndY,
	})

	start := game.Position{X: req.StartX, Y: req.StartY}
	end := game.Position{X: req.EndX, Y: req.EndY}

	pathfinder := game.NewPathFinder(s.state.WorldState)
	path, found := pathfinder.FindPath(start, end)

	// Convert path to serializable format
	var pathData []map[string]int
	for _, pos := range path {
		pathData = append(pathData, map[string]int{
			"x":     pos.X,
			"y":     pos.Y,
			"level": pos.Level,
		})
	}

	logger.WithFields(logrus.Fields{
		"found":      found,
		"pathLength": len(path),
	}).Info("pathfinding completed")

	return map[string]interface{}{
		"success":     true,
		"path":        pathData,
		"path_length": len(path),
		"found":       found,
	}, nil
}
