package utils

import (
	"container/heap"
	"math"

	"goldbox-rpg/pkg/game"
)

// PathfindingResult represents the result of pathfinding
type PathfindingResult struct {
	Path     []game.Position `json:"path"`
	Found    bool            `json:"found"`
	Distance int             `json:"distance"`
}

// Node represents a node in the A* pathfinding algorithm.
// It stores position, cost values, and parent reference for path reconstruction.
//
// Fields:
//   - Position: The grid position this node represents
//   - G: Actual cost from start to this node (accumulated movement cost)
//   - H: Heuristic estimate from this node to goal (Manhattan distance)
//   - F: Total estimated cost (G + H), used for priority queue ordering
//   - Parent: Reference to the previous node in the path, used to reconstruct
//     the final path once the goal is reached
//   - Index: Internal field used by the priority queue implementation for
//     efficient heap operations; not intended for external use
type Node struct {
	Position game.Position
	G        int   // Cost from start to this node
	H        int   // Heuristic cost from this node to goal
	F        int   // Total cost (G + H)
	Parent   *Node // Parent node for path reconstruction
	Index    int   // Index in the priority queue (internal use)
}

// PriorityQueue implements a priority queue for A* pathfinding
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].F < pq[j].F
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*Node)
	node.Index = n
	*pq = append(*pq, node)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	node.Index = -1
	*pq = old[0 : n-1]
	return node
}

// AStarPathfind finds optimal path using A* algorithm
func AStarPathfind(gameMap *game.GameMap, start, goal game.Position) *PathfindingResult {
	// Check if start and goal are valid
	if !isValidPosition(gameMap, start) || !isValidPosition(gameMap, goal) {
		return &PathfindingResult{Found: false}
	}

	// Check if start and goal are walkable
	if !gameMap.Tiles[start.Y][start.X].Walkable || !gameMap.Tiles[goal.Y][goal.X].Walkable {
		return &PathfindingResult{Found: false}
	}

	// Initialize open and closed sets
	openSet := &PriorityQueue{}
	heap.Init(openSet)

	closedSet := make(map[game.Position]bool)
	nodeMap := make(map[game.Position]*Node)

	// Create start node
	startNode := &Node{
		Position: start,
		G:        0,
		H:        manhattanDistance(start, goal),
		Parent:   nil,
	}
	startNode.F = startNode.G + startNode.H

	heap.Push(openSet, startNode)
	nodeMap[start] = startNode

	for openSet.Len() > 0 {
		// Get node with lowest F cost
		current := heap.Pop(openSet).(*Node)

		// Check if we reached the goal
		if current.Position == goal {
			path := reconstructPath(current)
			return &PathfindingResult{
				Path:     path,
				Found:    true,
				Distance: len(path) - 1,
			}
		}

		// Add current to closed set
		closedSet[current.Position] = true

		// Check all neighbors
		neighbors := getNeighbors(gameMap, current.Position)
		for _, neighborPos := range neighbors {
			if closedSet[neighborPos] {
				continue
			}

			tentativeG := current.G + 1 // Cost to move to neighbor

			// Check if this neighbor is already in open set
			neighborNode, exists := nodeMap[neighborPos]
			if !exists {
				// Create new node
				neighborNode = &Node{
					Position: neighborPos,
					G:        tentativeG,
					H:        manhattanDistance(neighborPos, goal),
					Parent:   current,
				}
				neighborNode.F = neighborNode.G + neighborNode.H
				heap.Push(openSet, neighborNode)
				nodeMap[neighborPos] = neighborNode
			} else if tentativeG < neighborNode.G {
				// This path to neighbor is better than previous one
				neighborNode.G = tentativeG
				neighborNode.F = neighborNode.G + neighborNode.H
				neighborNode.Parent = current
				heap.Fix(openSet, neighborNode.Index)
			}
		}
	}

	// No path found
	return &PathfindingResult{Found: false}
}

// FloodFill finds all connected walkable areas
func FloodFill(gameMap *game.GameMap, start game.Position) []game.Position {
	if !isValidPosition(gameMap, start) || !gameMap.Tiles[start.Y][start.X].Walkable {
		return nil
	}

	visited := make(map[game.Position]bool)
	var result []game.Position
	var stack []game.Position

	stack = append(stack, start)

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}

		if !isValidPosition(gameMap, current) || !gameMap.Tiles[current.Y][current.X].Walkable {
			continue
		}

		visited[current] = true
		result = append(result, current)

		// Add 4-connected neighbors
		neighbors := []game.Position{
			{X: current.X + 1, Y: current.Y},
			{X: current.X - 1, Y: current.Y},
			{X: current.X, Y: current.Y + 1},
			{X: current.X, Y: current.Y - 1},
		}

		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				stack = append(stack, neighbor)
			}
		}
	}

	return result
}

// ValidateConnectivity checks if all walkable areas are connected
func ValidateConnectivity(gameMap *game.GameMap) bool {
	// Find all walkable tiles
	var walkableTiles []game.Position
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			if gameMap.Tiles[y][x].Walkable {
				walkableTiles = append(walkableTiles, game.Position{X: x, Y: y})
			}
		}
	}

	if len(walkableTiles) == 0 {
		return true // No walkable tiles, technically connected
	}

	// Use flood fill from first walkable tile
	reachable := FloodFill(gameMap, walkableTiles[0])

	// Check if all walkable tiles are reachable
	return len(reachable) == len(walkableTiles)
}

// Helper functions

// isValidPosition checks if a position is within map bounds
func isValidPosition(gameMap *game.GameMap, pos game.Position) bool {
	return pos.X >= 0 && pos.X < gameMap.Width && pos.Y >= 0 && pos.Y < gameMap.Height
}

// manhattanDistance calculates Manhattan distance between two positions
func manhattanDistance(a, b game.Position) int {
	return int(math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y)))
}

// getNeighbors returns valid walkable neighbors of a position
func getNeighbors(gameMap *game.GameMap, pos game.Position) []game.Position {
	neighbors := []game.Position{
		{X: pos.X + 1, Y: pos.Y},
		{X: pos.X - 1, Y: pos.Y},
		{X: pos.X, Y: pos.Y + 1},
		{X: pos.X, Y: pos.Y - 1},
	}

	var validNeighbors []game.Position
	for _, neighbor := range neighbors {
		if isValidPosition(gameMap, neighbor) && gameMap.Tiles[neighbor.Y][neighbor.X].Walkable {
			validNeighbors = append(validNeighbors, neighbor)
		}
	}

	return validNeighbors
}

// reconstructPath builds the path from goal back to start
func reconstructPath(node *Node) []game.Position {
	var path []game.Position
	current := node

	for current != nil {
		path = append([]game.Position{current.Position}, path...)
		current = current.Parent
	}

	return path
}
