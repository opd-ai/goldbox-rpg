package game

import (
	"container/heap"
	"math"
)

// PathFinder implements A* pathfinding algorithm using the existing spatial index.
// It finds the shortest path between two positions, respecting tile walkability.
//
// Fields:
//   - World: Reference to the game world for tile walkability checks
//
// Related types:
//   - Position: Start and end coordinates for pathfinding
//   - World: Game world containing map data
//   - Tile: Map tiles with walkability information
//
// Usage:
//
//	pf := &PathFinder{World: gameWorld}
//	path, found := pf.FindPath(start, end)
//	if found {
//	  // Navigate along path
//	}
type PathFinder struct {
	World *World
}

// NewPathFinder creates a new pathfinding instance for the given world.
func NewPathFinder(world *World) *PathFinder {
	return &PathFinder{World: world}
}

// FindPath uses A* algorithm to find the shortest walkable path.
// Returns the path as a slice of positions and true if a path exists.
// Returns nil, false if no valid path can be found.
func (pf *PathFinder) FindPath(start, end Position) ([]Position, bool) {
	if start.Level != end.Level {
		return nil, false
	}

	if !pf.isWalkable(end.X, end.Y) {
		return nil, false
	}

	openSet := &priorityQueue{}
	heap.Init(openSet)

	startNode := &pathNode{
		pos:  start,
		g:    0,
		h:    pf.heuristic(start, end),
		f:    pf.heuristic(start, end),
		prev: nil,
	}
	heap.Push(openSet, startNode)

	closed := make(map[Position]bool)
	gScore := make(map[Position]float64)
	gScore[start] = 0

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*pathNode)

		if current.pos.X == end.X && current.pos.Y == end.Y {
			return pf.reconstructPath(current), true
		}

		closed[current.pos] = true

		for _, neighbor := range pf.getNeighbors(current.pos) {
			if closed[neighbor] {
				continue
			}

			tentativeG := current.g + 1.0

			if score, exists := gScore[neighbor]; exists && tentativeG >= score {
				continue
			}

			neighborNode := &pathNode{
				pos:  neighbor,
				g:    tentativeG,
				h:    pf.heuristic(neighbor, end),
				f:    tentativeG + pf.heuristic(neighbor, end),
				prev: current,
			}

			gScore[neighbor] = tentativeG
			heap.Push(openSet, neighborNode)
		}
	}

	return nil, false
}

// CanReach checks if a path exists between two positions without computing the full path.
func (pf *PathFinder) CanReach(start, end Position) bool {
	_, found := pf.FindPath(start, end)
	return found
}

// heuristic estimates the cost from pos to goal using Manhattan distance.
func (pf *PathFinder) heuristic(pos, goal Position) float64 {
	dx := math.Abs(float64(pos.X - goal.X))
	dy := math.Abs(float64(pos.Y - goal.Y))
	return dx + dy
}

// getNeighbors returns walkable adjacent positions (4-directional movement).
func (pf *PathFinder) getNeighbors(pos Position) []Position {
	neighbors := make([]Position, 0, 4)
	directions := []struct{ dx, dy int }{
		{0, -1}, // North
		{1, 0},  // East
		{0, 1},  // South
		{-1, 0}, // West
	}

	for _, dir := range directions {
		nx, ny := pos.X+dir.dx, pos.Y+dir.dy
		if pf.isWalkable(nx, ny) {
			neighbors = append(neighbors, Position{
				X:      nx,
				Y:      ny,
				Level:  pos.Level,
				Facing: pos.Facing,
			})
		}
	}

	return neighbors
}

// isWalkable checks if a tile at the given coordinates is walkable.
func (pf *PathFinder) isWalkable(x, y int) bool {
	if pf.World == nil || len(pf.World.Levels) == 0 {
		return false
	}

	// Use first level for now (Level 0)
	// In future, can be extended to specify which level index
	level := &pf.World.Levels[0]

	if x < 0 || y < 0 || x >= level.Width || y >= level.Height {
		return false
	}

	if len(level.Tiles) <= y || len(level.Tiles[y]) <= x {
		return false
	}

	return level.Tiles[y][x].Walkable
}

// reconstructPath builds the final path by following prev pointers.
func (pf *PathFinder) reconstructPath(node *pathNode) []Position {
	path := make([]Position, 0)
	for node != nil {
		path = append([]Position{node.pos}, path...)
		node = node.prev
	}
	return path
}

// pathNode represents a node in the A* search space.
type pathNode struct {
	pos  Position
	g    float64   // Cost from start
	h    float64   // Heuristic to goal
	f    float64   // Total estimated cost
	prev *pathNode // Previous node in path
}

// priorityQueue implements heap.Interface for A* open set.
// It orders path nodes by their total estimated cost (f-score) for priority queue operations.
type priorityQueue []*pathNode

// Len returns the number of nodes in the priority queue.
// This is part of the heap.Interface implementation.
func (pq priorityQueue) Len() int { return len(pq) }

// Less compares two nodes by their f-score for heap ordering.
// Returns true if node at index i has a lower f-score than node at index j.
func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].f < pq[j].f
}

// Swap exchanges the nodes at indices i and j in the priority queue.
// This is part of the heap.Interface implementation.
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// Push adds a pathNode to the priority queue.
// The element must be a *pathNode.
func (pq *priorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*pathNode))
}

// Pop removes and returns the highest priority (lowest f-score) node from the queue.
// Returns the removed element as interface{}.
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
