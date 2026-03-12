package game

import (
	"testing"
)

// TestPathFinderFindPath tests A* pathfinding across various scenarios.
func TestPathFinderFindPath(t *testing.T) {
	tests := []struct {
		name      string
		setupMap  func() *World
		start     Position
		end       Position
		wantFound bool
		wantLen   int
	}{
		{
			name: "simple straight path",
			setupMap: func() *World {
				return createTestWorld(5, 5, [][]bool{
					{true, true, true, true, true},
					{true, true, true, true, true},
					{true, true, true, true, true},
					{true, true, true, true, true},
					{true, true, true, true, true},
				})
			},
			start:     Position{X: 0, Y: 0, Level: 0},
			end:       Position{X: 4, Y: 0, Level: 0},
			wantFound: true,
			wantLen:   5,
		},
		{
			name: "path around obstacle",
			setupMap: func() *World {
				return createTestWorld(5, 5, [][]bool{
					{true, true, true, true, true},
					{true, false, false, false, true},
					{true, true, true, false, true},
					{true, true, true, false, true},
					{true, true, true, true, true},
				})
			},
			start:     Position{X: 0, Y: 0, Level: 0},
			end:       Position{X: 4, Y: 0, Level: 0},
			wantFound: true,
			wantLen:   5, // Straight path across top row
		},
		{
			name: "no path exists",
			setupMap: func() *World {
				return createTestWorld(5, 5, [][]bool{
					{true, false, true, true, true},
					{true, false, true, true, true},
					{true, false, true, true, true},
					{true, false, true, true, true},
					{true, false, true, true, true},
				})
			},
			start:     Position{X: 0, Y: 0, Level: 0},
			end:       Position{X: 4, Y: 0, Level: 0},
			wantFound: false,
			wantLen:   0,
		},
		{
			name: "start equals end",
			setupMap: func() *World {
				return createTestWorld(3, 3, [][]bool{
					{true, true, true},
					{true, true, true},
					{true, true, true},
				})
			},
			start:     Position{X: 1, Y: 1, Level: 0},
			end:       Position{X: 1, Y: 1, Level: 0},
			wantFound: true,
			wantLen:   1,
		},
		{
			name: "destination not walkable",
			setupMap: func() *World {
				return createTestWorld(3, 3, [][]bool{
					{true, true, true},
					{true, true, false},
					{true, true, true},
				})
			},
			start:     Position{X: 0, Y: 0, Level: 0},
			end:       Position{X: 2, Y: 1, Level: 0},
			wantFound: false,
			wantLen:   0,
		},
		{
			name: "maze path",
			setupMap: func() *World {
				return createTestWorld(7, 7, [][]bool{
					{true, false, true, true, true, true, true},
					{true, false, true, false, false, false, true},
					{true, true, true, false, true, true, true},
					{false, false, false, false, true, false, false},
					{true, true, true, true, true, true, true},
					{true, false, false, false, false, false, true},
					{true, true, true, true, true, true, true},
				})
			},
			start:     Position{X: 0, Y: 0, Level: 0},
			end:       Position{X: 6, Y: 6, Level: 0},
			wantFound: true,
			wantLen:   21,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := tt.setupMap()
			pf := NewPathFinder(world)

			path, found := pf.FindPath(tt.start, tt.end)

			if found != tt.wantFound {
				t.Errorf("FindPath() found = %v, want %v", found, tt.wantFound)
				return
			}

			if found {
				if len(path) != tt.wantLen {
					t.Errorf("FindPath() path length = %d, want %d", len(path), tt.wantLen)
				}

				if len(path) > 0 {
					if path[0] != tt.start {
						t.Errorf("Path start = %v, want %v", path[0], tt.start)
					}
					if path[len(path)-1].X != tt.end.X || path[len(path)-1].Y != tt.end.Y {
						t.Errorf("Path end = %v, want position (%d,%d)", path[len(path)-1], tt.end.X, tt.end.Y)
					}
				}

				// Verify path continuity
				for i := 0; i < len(path)-1; i++ {
					dx := abs(path[i+1].X - path[i].X)
					dy := abs(path[i+1].Y - path[i].Y)
					if dx+dy != 1 {
						t.Errorf("Path not continuous at index %d: (%d,%d) -> (%d,%d)",
							i, path[i].X, path[i].Y, path[i+1].X, path[i+1].Y)
					}
				}
			}
		})
	}
}

// TestPathFinderCanReach tests reachability checks.
func TestPathFinderCanReach(t *testing.T) {
	tests := []struct {
		name          string
		setupMap      func() *World
		start         Position
		end           Position
		wantReachable bool
	}{
		{
			name: "reachable position",
			setupMap: func() *World {
				return createTestWorld(5, 5, [][]bool{
					{true, true, true, true, true},
					{true, true, true, true, true},
					{true, true, true, true, true},
					{true, true, true, true, true},
					{true, true, true, true, true},
				})
			},
			start:         Position{X: 0, Y: 0, Level: 0},
			end:           Position{X: 4, Y: 4, Level: 0},
			wantReachable: true,
		},
		{
			name: "unreachable position",
			setupMap: func() *World {
				return createTestWorld(5, 5, [][]bool{
					{true, false, true, true, true},
					{true, false, true, true, true},
					{true, false, true, true, true},
					{true, false, true, true, true},
					{true, false, true, true, true},
				})
			},
			start:         Position{X: 0, Y: 0, Level: 0},
			end:           Position{X: 4, Y: 4, Level: 0},
			wantReachable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := tt.setupMap()
			pf := NewPathFinder(world)

			reachable := pf.CanReach(tt.start, tt.end)

			if reachable != tt.wantReachable {
				t.Errorf("CanReach() = %v, want %v", reachable, tt.wantReachable)
			}
		})
	}
}

// TestPathFinderDifferentLevels tests that pathfinding fails across different levels.
func TestPathFinderDifferentLevels(t *testing.T) {
	world := createTestWorld(3, 3, [][]bool{
		{true, true, true},
		{true, true, true},
		{true, true, true},
	})

	pf := NewPathFinder(world)

	start := Position{X: 0, Y: 0, Level: 0}
	end := Position{X: 2, Y: 2, Level: 1}

	path, found := pf.FindPath(start, end)

	if found {
		t.Error("FindPath() should not find path across different levels")
	}
	if path != nil {
		t.Error("FindPath() should return nil path for different levels")
	}
}

// TestPathFinderNilWorld tests handling of nil world.
func TestPathFinderNilWorld(t *testing.T) {
	pf := &PathFinder{World: nil}

	start := Position{X: 0, Y: 0, Level: 0}
	end := Position{X: 1, Y: 1, Level: 0}

	path, found := pf.FindPath(start, end)

	if found {
		t.Error("FindPath() with nil world should not find path")
	}
	if path != nil {
		t.Error("FindPath() with nil world should return nil path")
	}
}

// TestPathFinderOptimality tests that A* finds optimal (shortest) paths.
func TestPathFinderOptimality(t *testing.T) {
	world := createTestWorld(5, 5, [][]bool{
		{true, true, true, true, true},
		{true, false, true, false, true},
		{true, true, true, true, true},
		{true, false, true, false, true},
		{true, true, true, true, true},
	})

	pf := NewPathFinder(world)

	start := Position{X: 0, Y: 2, Level: 0}
	end := Position{X: 4, Y: 2, Level: 0}

	path, found := pf.FindPath(start, end)

	if !found {
		t.Fatal("FindPath() should find path")
	}

	// Optimal path length is 5 (straight across middle row)
	if len(path) != 5 {
		t.Errorf("FindPath() path length = %d, want 5 (optimal)", len(path))
	}
}

// createTestWorld creates a test world with the given walkability grid.
func createTestWorld(width, height int, walkable [][]bool) *World {
	tiles := make([][]Tile, height)
	for y := 0; y < height; y++ {
		tiles[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			tiles[y][x] = Tile{
				Walkable: walkable[y][x],
			}
		}
	}

	level := Level{
		ID:     "test",
		Width:  width,
		Height: height,
		Tiles:  tiles,
	}

	return &World{
		Levels: []Level{level},
	}
}
