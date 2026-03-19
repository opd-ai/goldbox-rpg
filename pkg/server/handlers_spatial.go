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

// VisibleTile represents a tile visible from the player's first-person view.
type VisibleTile struct {
	RelativeX int    `json:"rel_x"` // -1 (left), 0 (center), 1 (right)
	Depth     int    `json:"depth"` // 0 = near, 1 = mid, 2 = far
	TileType  string `json:"type"`  // wall, floor, door_open, door_closed
	Walkable  bool   `json:"walkable"`
}

// handleGetVisibleTiles processes a request for visible tiles in the player's first-person view.
// Returns a 3-deep view cone based on player position and facing direction.
//
// Parameters:
//   - params: json.RawMessage containing:
//   - session_id: string identifier for the player session
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if query was successful
//   - tiles: Array of VisibleTile representing the view cone
//   - facing: Current facing direction
//   - position: Current player position
func (s *RPCServer) handleGetVisibleTiles(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetVisibleTiles",
	}).Debug("entering visible tiles handler")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithError(err).Error("failed to unmarshal visible tiles parameters")
		return map[string]interface{}{
			"success": false,
			"error":   "invalid parameters",
		}, nil
	}

	session, exists := s.getSession(req.SessionID)
	if !exists {
		logrus.WithField("sessionID", req.SessionID).Warn("visible tiles attempted with invalid session")
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

	pos := session.Player.GetPosition()
	facing := pos.Facing
	levelIdx := pos.Level

	logger := logrus.WithFields(logrus.Fields{
		"sessionID": req.SessionID,
		"playerID":  session.Player.GetID(),
		"posX":      pos.X,
		"posY":      pos.Y,
		"facing":    facing,
		"level":     levelIdx,
	})

	// Get the level tiles
	s.mu.RLock()
	world := s.state.WorldState
	s.mu.RUnlock()

	if world == nil || len(world.Levels) == 0 {
		logger.Warn("no world or levels available")
		return map[string]interface{}{
			"success": false,
			"error":   "no map data available",
		}, nil
	}

	// Ensure level index is valid
	if levelIdx < 0 || levelIdx >= len(world.Levels) {
		levelIdx = 0
	}
	level := &world.Levels[levelIdx]

	// Calculate direction vectors based on facing
	var dx, dy int
	switch game.Direction(facing) {
	case game.DirectionNorth:
		dx, dy = 0, -1
	case game.DirectionSouth:
		dx, dy = 0, 1
	case game.DirectionEast:
		dx, dy = 1, 0
	case game.DirectionWest:
		dx, dy = -1, 0
	default:
		dx, dy = 0, -1 // Default north
	}

	// Calculate perpendicular direction for left/right
	var ldx, ldy int // Left offset
	switch game.Direction(facing) {
	case game.DirectionNorth:
		ldx, ldy = -1, 0
	case game.DirectionSouth:
		ldx, ldy = 1, 0
	case game.DirectionEast:
		ldx, ldy = 0, -1
	case game.DirectionWest:
		ldx, ldy = 0, 1
	default:
		ldx, ldy = -1, 0
	}

	tiles := make([]VisibleTile, 0, 9) // 3 depths * 3 positions

	// Generate view cone tiles for depths 0, 1, 2
	for depth := 0; depth <= 2; depth++ {
		// Center tile at this depth
		cx := pos.X + dx*(depth+1)
		cy := pos.Y + dy*(depth+1)

		// Left tile at this depth
		lx := cx + ldx
		ly := cy + ldy

		// Right tile at this depth
		rx := cx - ldx
		ry := cy - ldy

		// Add left tile
		tiles = append(tiles, getTileInfo(level, lx, ly, -1, depth))
		// Add center tile
		tiles = append(tiles, getTileInfo(level, cx, cy, 0, depth))
		// Add right tile
		tiles = append(tiles, getTileInfo(level, rx, ry, 1, depth))
	}

	logger.WithField("tileCount", len(tiles)).Info("visible tiles query completed")

	return map[string]interface{}{
		"success": true,
		"tiles":   tiles,
		"facing":  facing,
		"position": map[string]int{
			"x":     pos.X,
			"y":     pos.Y,
			"level": pos.Level,
		},
	}, nil
}

// getTileInfo extracts tile information for the visible tiles response.
func getTileInfo(level *game.Level, x, y, relX, depth int) VisibleTile {
	// Out of bounds = wall
	if x < 0 || y < 0 || x >= level.Width || y >= level.Height {
		return VisibleTile{
			RelativeX: relX,
			Depth:     depth,
			TileType:  "wall",
			Walkable:  false,
		}
	}

	tile := &level.Tiles[y][x]
	tileType := "floor"

	switch tile.Type {
	case game.TileWall:
		tileType = "wall"
	case game.TileDoor:
		// Check if door is open via properties
		if open, ok := tile.Properties["open"].(bool); ok && open {
			tileType = "door_open"
		} else {
			tileType = "door_closed"
		}
	case game.TileFloor:
		tileType = "floor"
	case game.TileStairs:
		tileType = "stairs"
	default:
		tileType = "floor"
	}

	return VisibleTile{
		RelativeX: relX,
		Depth:     depth,
		TileType:  tileType,
		Walkable:  tile.Walkable,
	}
}
