package game

import (
	"fmt"
	"math"
	"sync"
)

// SpatialIndex provides efficient spatial queries for game objects
// Implements an R-tree-like spatial data structure optimized for 2D game worlds
type SpatialIndex struct {
	mu       sync.RWMutex
	root     *SpatialNode
	cellSize int
	bounds   Rectangle
}

// SpatialNode represents a node in the spatial index tree
type SpatialNode struct {
	bounds   Rectangle
	objects  []GameObject
	children []*SpatialNode
	isLeaf   bool
}

// Rectangle represents a bounding box for spatial queries
type Rectangle struct {
	MinX, MinY, MaxX, MaxY int
}

// Circle represents a circular query area
type Circle struct {
	CenterX, CenterY int
	Radius           float64
}

// NewSpatialIndex creates a new spatial index with specified bounds and cell size
func NewSpatialIndex(width, height, cellSize int) *SpatialIndex {
	return &SpatialIndex{
		cellSize: cellSize,
		bounds: Rectangle{
			MinX: 0, MinY: 0,
			MaxX: width, MaxY: height,
		},
		root: &SpatialNode{
			bounds:  Rectangle{MinX: 0, MinY: 0, MaxX: width, MaxY: height},
			isLeaf:  true,
			objects: make([]GameObject, 0),
		},
	}
}

// Insert adds a game object to the spatial index
func (si *SpatialIndex) Insert(obj GameObject) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	pos := obj.GetPosition()
	if !si.contains(si.bounds, pos) {
		return fmt.Errorf("object position %v is outside spatial index bounds", pos)
	}

	return si.insertNode(si.root, obj)
}

// Remove removes a game object from the spatial index
func (si *SpatialIndex) Remove(objectID string) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	return si.removeNode(si.root, objectID)
}

// GetObjectsInRange returns all objects within a rectangular area
func (si *SpatialIndex) GetObjectsInRange(rect Rectangle) []GameObject {
	si.mu.RLock()
	defer si.mu.RUnlock()

	var result []GameObject
	si.queryNode(si.root, rect, &result)
	return result
}

// GetObjectsInRadius returns all objects within a circular area
func (si *SpatialIndex) GetObjectsInRadius(center Position, radius float64) []GameObject {
	si.mu.RLock()
	defer si.mu.RUnlock()

	// Convert circle to bounding rectangle for initial filtering
	rect := Rectangle{
		MinX: center.X - int(math.Ceil(radius)),
		MinY: center.Y - int(math.Ceil(radius)),
		MaxX: center.X + int(math.Ceil(radius)),
		MaxY: center.Y + int(math.Ceil(radius)),
	}

	var candidates []GameObject
	si.queryNode(si.root, rect, &candidates)

	// Filter candidates by actual circular distance
	var result []GameObject
	for _, obj := range candidates {
		objPos := obj.GetPosition()
		distance := si.distance(center, objPos)
		if distance <= radius {
			result = append(result, obj)
		}
	}

	return result
}

// GetNearestObjects returns the k nearest objects to a given position
func (si *SpatialIndex) GetNearestObjects(center Position, k int) []GameObject {
	si.mu.RLock()
	defer si.mu.RUnlock()

	// Start with a small radius and expand as needed
	radius := float64(si.cellSize)
	maxRadius := float64(max(si.bounds.MaxX-si.bounds.MinX, si.bounds.MaxY-si.bounds.MinY))

	for radius <= maxRadius {
		objects := si.GetObjectsInRadius(center, radius)
		if len(objects) >= k {
			// Sort by distance and return k nearest
			si.sortByDistance(objects, center)
			if len(objects) > k {
				return objects[:k]
			}
			return objects
		}
		radius *= 2
	}

	// If we still don't have enough, return all objects sorted by distance
	var allObjects []GameObject
	si.queryNode(si.root, si.bounds, &allObjects)
	si.sortByDistance(allObjects, center)
	if len(allObjects) > k {
		return allObjects[:k]
	}
	return allObjects
}

// GetObjectsAt returns all objects at an exact position (optimized for single-cell queries)
func (si *SpatialIndex) GetObjectsAt(pos Position) []GameObject {
	rect := Rectangle{
		MinX: pos.X, MinY: pos.Y,
		MaxX: pos.X, MaxY: pos.Y,
	}
	return si.GetObjectsInRange(rect)
}

// Update moves an object to a new position in the spatial index
func (si *SpatialIndex) Update(objectID string, newPos Position) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	// Find and remove the object
	var obj GameObject
	if err := si.removeNodeWithObject(si.root, objectID, &obj); err != nil {
		return fmt.Errorf("object %s not found for update: %w", objectID, err)
	}

	// Re-insert at new position
	return si.insertNode(si.root, obj)
}

// Clear removes all objects from the spatial index
func (si *SpatialIndex) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.root = &SpatialNode{
		bounds:  si.bounds,
		isLeaf:  true,
		objects: make([]GameObject, 0),
	}
}

// GetStats returns statistics about the spatial index
func (si *SpatialIndex) GetStats() SpatialIndexStats {
	si.mu.RLock()
	defer si.mu.RUnlock()

	stats := SpatialIndexStats{}
	si.collectStats(si.root, &stats, 0)
	return stats
}

// SpatialIndexStats provides performance and structure information
type SpatialIndexStats struct {
	TotalObjects      int
	TotalNodes        int
	MaxDepth          int
	LeafNodes         int
	AvgObjectsPerLeaf float64
}

// Private helper methods

func (si *SpatialIndex) insertNode(node *SpatialNode, obj GameObject) error {
	pos := obj.GetPosition()

	if !si.contains(node.bounds, pos) {
		return fmt.Errorf("object position %v outside node bounds %v", pos, node.bounds)
	}

	if node.isLeaf {
		node.objects = append(node.objects, obj)

		// Split if node becomes too full
		if len(node.objects) > 8 && si.canSplit(node.bounds) {
			si.splitNode(node)
		}
		return nil
	}

	// Find best child node
	for _, child := range node.children {
		if si.contains(child.bounds, pos) {
			return si.insertNode(child, obj)
		}
	}

	return fmt.Errorf("no suitable child node found for position %v", pos)
}

func (si *SpatialIndex) removeNode(node *SpatialNode, objectID string) error {
	if node.isLeaf {
		for i, obj := range node.objects {
			if obj.GetID() == objectID {
				// Remove object by swapping with last element
				node.objects[i] = node.objects[len(node.objects)-1]
				node.objects = node.objects[:len(node.objects)-1]
				return nil
			}
		}
		return fmt.Errorf("object %s not found", objectID)
	}

	// Recursively search children
	for _, child := range node.children {
		if err := si.removeNode(child, objectID); err == nil {
			return nil
		}
	}

	return fmt.Errorf("object %s not found in any child", objectID)
}

func (si *SpatialIndex) removeNodeWithObject(node *SpatialNode, objectID string, obj *GameObject) error {
	if node.isLeaf {
		for i, o := range node.objects {
			if o.GetID() == objectID {
				*obj = o
				node.objects[i] = node.objects[len(node.objects)-1]
				node.objects = node.objects[:len(node.objects)-1]
				return nil
			}
		}
		return fmt.Errorf("object %s not found", objectID)
	}

	for _, child := range node.children {
		if err := si.removeNodeWithObject(child, objectID, obj); err == nil {
			return nil
		}
	}

	return fmt.Errorf("object %s not found in any child", objectID)
}

func (si *SpatialIndex) queryNode(node *SpatialNode, rect Rectangle, result *[]GameObject) {
	if !si.intersects(node.bounds, rect) {
		return
	}

	if node.isLeaf {
		for _, obj := range node.objects {
			pos := obj.GetPosition()
			if si.contains(rect, pos) {
				*result = append(*result, obj)
			}
		}
		return
	}

	for _, child := range node.children {
		si.queryNode(child, rect, result)
	}
}

func (si *SpatialIndex) splitNode(node *SpatialNode) {
	if !node.isLeaf || len(node.objects) <= 1 {
		return
	}

	bounds := node.bounds
	midX := (bounds.MinX + bounds.MaxX) / 2
	midY := (bounds.MinY + bounds.MaxY) / 2

	// Create four child nodes
	node.children = []*SpatialNode{
		{bounds: Rectangle{bounds.MinX, bounds.MinY, midX, midY}, isLeaf: true, objects: make([]GameObject, 0)},
		{bounds: Rectangle{midX, bounds.MinY, bounds.MaxX, midY}, isLeaf: true, objects: make([]GameObject, 0)},
		{bounds: Rectangle{bounds.MinX, midY, midX, bounds.MaxY}, isLeaf: true, objects: make([]GameObject, 0)},
		{bounds: Rectangle{midX, midY, bounds.MaxX, bounds.MaxY}, isLeaf: true, objects: make([]GameObject, 0)},
	}

	// Redistribute objects to children
	for _, obj := range node.objects {
		pos := obj.GetPosition()
		for _, child := range node.children {
			if si.contains(child.bounds, pos) {
				child.objects = append(child.objects, obj)
				break
			}
		}
	}

	// Clear parent objects and mark as non-leaf
	node.objects = nil
	node.isLeaf = false
}

func (si *SpatialIndex) canSplit(bounds Rectangle) bool {
	width := bounds.MaxX - bounds.MinX
	height := bounds.MaxY - bounds.MinY
	return width > si.cellSize && height > si.cellSize
}

func (si *SpatialIndex) contains(rect Rectangle, pos Position) bool {
	return pos.X >= rect.MinX && pos.X <= rect.MaxX &&
		pos.Y >= rect.MinY && pos.Y <= rect.MaxY
}

func (si *SpatialIndex) intersects(rect1, rect2 Rectangle) bool {
	return rect1.MinX <= rect2.MaxX && rect1.MaxX >= rect2.MinX &&
		rect1.MinY <= rect2.MaxY && rect1.MaxY >= rect2.MinY
}

func (si *SpatialIndex) distance(pos1, pos2 Position) float64 {
	dx := float64(pos1.X - pos2.X)
	dy := float64(pos1.Y - pos2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

func (si *SpatialIndex) sortByDistance(objects []GameObject, center Position) {
	for i := 0; i < len(objects)-1; i++ {
		for j := i + 1; j < len(objects); j++ {
			dist1 := si.distance(center, objects[i].GetPosition())
			dist2 := si.distance(center, objects[j].GetPosition())
			if dist1 > dist2 {
				objects[i], objects[j] = objects[j], objects[i]
			}
		}
	}
}

func (si *SpatialIndex) collectStats(node *SpatialNode, stats *SpatialIndexStats, depth int) {
	stats.TotalNodes++
	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	if node.isLeaf {
		stats.LeafNodes++
		stats.TotalObjects += len(node.objects)
	} else {
		for _, child := range node.children {
			si.collectStats(child, stats, depth+1)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
