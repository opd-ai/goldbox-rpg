// Package main provides the map-editor CLI tool for creating ASCII-based tile maps.
// This tool allows users to create and edit game maps that export to the GameMap format.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"goldbox-rpg/pkg/game"
)

// Config holds the command-line configuration for the map editor.
type Config struct {
	// OutputFile specifies the path to write the generated JSON map.
	OutputFile string
	// Width specifies the map width in tiles.
	Width int
	// Height specifies the map height in tiles.
	Height int
	// LoadFile specifies an existing map to load and edit.
	LoadFile string
	// Template specifies a template type for quick scaffolding.
	Template string
}

// tileChar maps tile types to ASCII characters for display.
var tileChar = map[string]rune{
	"floor":    '.',
	"wall":     '#',
	"water":    '~',
	"door":     '+',
	"stairs":   '>',
	"tree":     'T',
	"grass":    ',',
	"sand":     ':',
	"rock":     'o',
	"lava":     '*',
	"chest":    '$',
	"entrance": '@',
}

// charToTile maps ASCII characters to tile properties.
var charToTile = map[rune]game.MapTile{
	'.': {SpriteX: 0, SpriteY: 0, Walkable: true, Transparent: true},   // floor
	'#': {SpriteX: 1, SpriteY: 0, Walkable: false, Transparent: false}, // wall
	'~': {SpriteX: 2, SpriteY: 0, Walkable: false, Transparent: true},  // water
	'+': {SpriteX: 3, SpriteY: 0, Walkable: true, Transparent: false},  // door
	'>': {SpriteX: 4, SpriteY: 0, Walkable: true, Transparent: true},   // stairs
	'T': {SpriteX: 5, SpriteY: 0, Walkable: false, Transparent: false}, // tree
	',': {SpriteX: 0, SpriteY: 1, Walkable: true, Transparent: true},   // grass
	':': {SpriteX: 1, SpriteY: 1, Walkable: true, Transparent: true},   // sand
	'o': {SpriteX: 2, SpriteY: 1, Walkable: false, Transparent: true},  // rock
	'*': {SpriteX: 3, SpriteY: 1, Walkable: false, Transparent: true},  // lava
	'$': {SpriteX: 4, SpriteY: 1, Walkable: true, Transparent: true},   // chest
	'@': {SpriteX: 5, SpriteY: 1, Walkable: true, Transparent: true},   // entrance
}

// parseFlags parses command-line flags and returns the configuration.
func parseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.OutputFile, "output", "", "output file path for generated JSON (default: stdout)")
	flag.StringVar(&cfg.OutputFile, "o", "", "output file path (shorthand)")
	flag.IntVar(&cfg.Width, "width", 20, "map width in tiles")
	flag.IntVar(&cfg.Width, "w", 20, "map width (shorthand)")
	flag.IntVar(&cfg.Height, "height", 15, "map height in tiles")
	flag.IntVar(&cfg.Height, "h", 15, "map height (shorthand)")
	flag.StringVar(&cfg.LoadFile, "load", "", "load existing map file for editing")
	flag.StringVar(&cfg.LoadFile, "l", "", "load file (shorthand)")
	flag.StringVar(&cfg.Template, "template", "", "map template: dungeon, outdoor, cave, town")
	flag.StringVar(&cfg.Template, "t", "", "template type (shorthand)")
	flag.Usage = printUsage
	flag.Parse()
	return cfg
}

// printUsage prints the usage information for the map editor.
func printUsage() {
	fmt.Fprintf(os.Stderr, `Map Editor - Create tile maps for GoldBox RPG

Usage:
  map-editor [options]

Options:
  -o, --output FILE      Write output to FILE instead of stdout
  -w, --width N          Map width in tiles (default: 20)
  -h, --height N         Map height in tiles (default: 15)
  -l, --load FILE        Load existing map file for editing
  -t, --template TYPE    Use map template: dungeon, outdoor, cave, town
      --help             Show this help message

Tile Legend:
  . = floor (walkable)     # = wall (solid)
  ~ = water (impassable)   + = door (walkable)
  > = stairs (walkable)    T = tree (solid)
  , = grass (walkable)     : = sand (walkable)
  o = rock (solid)         * = lava (impassable)
  $ = chest (walkable)     @ = entrance (walkable)

Editor Commands (during interactive mode):
  [y,x]=CHAR    Set tile at (x,y) to character (e.g., 5,3=#)
  fill=CHAR     Fill entire map with character
  rect=x1,y1,x2,y2,CHAR  Draw rectangle outline
  solid=x1,y1,x2,y2,CHAR Fill solid rectangle
  save          Save and exit
  show          Display current map
  help          Show commands
  quit          Quit without saving

Examples:
  # Create new 30x20 map interactively
  map-editor -w 30 -h 20 -o dungeon.json

  # Use dungeon template
  map-editor -t dungeon -o dungeon.json

  # Edit existing map
  map-editor -l dungeon.json -o dungeon_v2.json

`)
}

// main is the entry point for the map editor application.
func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the map editor with the given configuration.
func run(cfg *Config) error {
	var gameMap *game.GameMap

	// Load existing map or create new one
	if cfg.LoadFile != "" {
		var err error
		gameMap, err = loadMap(cfg.LoadFile)
		if err != nil {
			return fmt.Errorf("failed to load map: %w", err)
		}
	} else if cfg.Template != "" {
		var err error
		gameMap, err = createFromTemplate(cfg.Template, cfg.Width, cfg.Height)
		if err != nil {
			return err
		}
	} else {
		gameMap = createEmptyMap(cfg.Width, cfg.Height)
	}

	// Interactive editing
	if err := interactiveEdit(gameMap); err != nil {
		return err
	}

	return outputMap(gameMap, cfg.OutputFile)
}

// createEmptyMap creates a new empty map filled with floor tiles.
func createEmptyMap(width, height int) *game.GameMap {
	tiles := make([][]game.MapTile, height)
	for y := 0; y < height; y++ {
		tiles[y] = make([]game.MapTile, width)
		for x := 0; x < width; x++ {
			tiles[y][x] = charToTile['.']
		}
	}
	return &game.GameMap{Width: width, Height: height, Tiles: tiles}
}

// createFromTemplate creates a map from a template.
func createFromTemplate(templateType string, width, height int) (*game.GameMap, error) {
	gameMap := createEmptyMap(width, height)

	switch strings.ToLower(templateType) {
	case "dungeon":
		applyDungeonTemplate(gameMap)
	case "outdoor":
		applyOutdoorTemplate(gameMap)
	case "cave":
		applyCaveTemplate(gameMap)
	case "town":
		applyTownTemplate(gameMap)
	default:
		return nil, fmt.Errorf("unknown template: %s (valid: dungeon, outdoor, cave, town)", templateType)
	}

	return gameMap, nil
}

// applyDungeonTemplate applies a dungeon-style template to the map.
func applyDungeonTemplate(m *game.GameMap) {
	// Fill with walls
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Tiles[y][x] = charToTile['#']
		}
	}
	// Carve out a central room
	roomX, roomY := 3, 3
	roomW, roomH := m.Width-6, m.Height-6
	for y := roomY; y < roomY+roomH; y++ {
		for x := roomX; x < roomX+roomW; x++ {
			m.Tiles[y][x] = charToTile['.']
		}
	}
	// Add entrance
	m.Tiles[m.Height/2][roomX-1] = charToTile['+']
	m.Tiles[m.Height/2][roomX-2] = charToTile['@']
	// Add stairs
	m.Tiles[m.Height/2][roomX+roomW] = charToTile['+']
	m.Tiles[m.Height/2][roomX+roomW+1] = charToTile['>']
}

// applyOutdoorTemplate applies an outdoor-style template to the map.
func applyOutdoorTemplate(m *game.GameMap) {
	// Fill with grass
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Tiles[y][x] = charToTile[',']
		}
	}
	// Add some trees around edges
	for x := 0; x < m.Width; x++ {
		if x%3 == 0 {
			m.Tiles[0][x] = charToTile['T']
			m.Tiles[m.Height-1][x] = charToTile['T']
		}
	}
	// Add a path
	for x := 0; x < m.Width; x++ {
		m.Tiles[m.Height/2][x] = charToTile['.']
	}
	// Add entrance
	m.Tiles[m.Height/2][0] = charToTile['@']
}

// applyCaveTemplate applies a cave-style template to the map.
func applyCaveTemplate(m *game.GameMap) {
	// Fill with rocks
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Tiles[y][x] = charToTile['o']
		}
	}
	// Carve irregular cave
	centerX, centerY := m.Width/2, m.Height/2
	for y := 2; y < m.Height-2; y++ {
		for x := 2; x < m.Width-2; x++ {
			// Simple cave shape
			dx := x - centerX
			dy := y - centerY
			if dx*dx+dy*dy < (m.Width/3)*(m.Height/3) {
				m.Tiles[y][x] = charToTile['.']
			}
		}
	}
	// Add entrance
	m.Tiles[centerY][1] = charToTile['.']
	m.Tiles[centerY][0] = charToTile['@']
	// Add water pool
	m.Tiles[centerY+2][centerX] = charToTile['~']
	m.Tiles[centerY+2][centerX+1] = charToTile['~']
}

// applyTownTemplate applies a town-style template to the map.
func applyTownTemplate(m *game.GameMap) {
	// Fill with grass
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Tiles[y][x] = charToTile[',']
		}
	}
	// Add roads
	for x := 0; x < m.Width; x++ {
		m.Tiles[m.Height/2][x] = charToTile['.']
	}
	for y := 0; y < m.Height; y++ {
		m.Tiles[y][m.Width/2] = charToTile['.']
	}
	// Add a building
	buildX, buildY := m.Width/4, m.Height/4
	for y := buildY; y < buildY+4; y++ {
		for x := buildX; x < buildX+5; x++ {
			m.Tiles[y][x] = charToTile['#']
		}
	}
	m.Tiles[buildY+3][buildX+2] = charToTile['+']
	// Add entrance
	m.Tiles[m.Height/2][0] = charToTile['@']
}

// loadMap loads a map from a JSON file.
func loadMap(filename string) (*game.GameMap, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var gameMap game.GameMap
	if err := json.Unmarshal(data, &gameMap); err != nil {
		return nil, err
	}
	return &gameMap, nil
}

// interactiveEdit provides interactive map editing.
func interactiveEdit(m *game.GameMap) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Map Editor - Interactive Mode ===")
	printHelp()
	displayMap(m)

	for {
		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		parts := strings.SplitN(input, "=", 2)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "save", "exit":
			return nil
		case "quit", "q":
			fmt.Println("Quitting without saving.")
			os.Exit(0)
		case "show", "display":
			displayMap(m)
		case "help", "?":
			printHelp()
		case "fill":
			if len(parts) < 2 {
				fmt.Println("Usage: fill=CHAR")
				continue
			}
			fillMap(m, rune(parts[1][0]))
			displayMap(m)
		case "rect":
			if len(parts) < 2 {
				fmt.Println("Usage: rect=x1,y1,x2,y2,CHAR")
				continue
			}
			if err := drawRect(m, parts[1], false); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				displayMap(m)
			}
		case "solid":
			if len(parts) < 2 {
				fmt.Println("Usage: solid=x1,y1,x2,y2,CHAR")
				continue
			}
			if err := drawRect(m, parts[1], true); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				displayMap(m)
			}
		default:
			// Try to parse as coordinate: y,x=CHAR
			if len(parts) == 2 {
				if err := setTile(m, parts[0], parts[1]); err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					displayMap(m)
				}
			} else {
				fmt.Println("Unknown command. Type 'help' for help.")
			}
		}
	}
}

// displayMap prints the map as ASCII art.
func displayMap(m *game.GameMap) {
	fmt.Printf("\nMap: %dx%d\n", m.Width, m.Height)

	// Column headers
	fmt.Print("   ")
	for x := 0; x < m.Width; x++ {
		fmt.Printf("%d", x%10)
	}
	fmt.Println()

	for y := 0; y < m.Height; y++ {
		fmt.Printf("%2d ", y)
		for x := 0; x < m.Width; x++ {
			tile := m.Tiles[y][x]
			char := tileToChar(tile)
			fmt.Printf("%c", char)
		}
		fmt.Println()
	}
}

// tileToChar converts a tile to its ASCII representation.
func tileToChar(tile game.MapTile) rune {
	for char, t := range charToTile {
		if t.SpriteX == tile.SpriteX && t.SpriteY == tile.SpriteY {
			return char
		}
	}
	return '?'
}

// printHelp prints available commands.
func printHelp() {
	fmt.Println("\nCommands:")
	fmt.Println("  y,x=CHAR    Set tile (e.g., 5,3=#)")
	fmt.Println("  fill=CHAR   Fill entire map")
	fmt.Println("  rect=x1,y1,x2,y2,CHAR  Draw rectangle")
	fmt.Println("  solid=x1,y1,x2,y2,CHAR Fill rectangle")
	fmt.Println("  show        Display map")
	fmt.Println("  save        Save and exit")
	fmt.Println("  quit        Quit without saving")
	fmt.Println("  help        Show this help")
	fmt.Println("\nTiles: . # ~ + > T , : o * $ @")
}

// fillMap fills the entire map with a tile type.
func fillMap(m *game.GameMap, char rune) {
	tile, ok := charToTile[char]
	if !ok {
		fmt.Printf("Unknown tile character: %c\n", char)
		return
	}
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Tiles[y][x] = tile
		}
	}
}

// setTile sets a single tile at coordinates.
func setTile(m *game.GameMap, coords, char string) error {
	parts := strings.Split(coords, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid coordinates, use: y,x")
	}

	y, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return fmt.Errorf("invalid y coordinate")
	}
	x, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return fmt.Errorf("invalid x coordinate")
	}

	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return fmt.Errorf("coordinates out of bounds")
	}

	tile, ok := charToTile[rune(char[0])]
	if !ok {
		return fmt.Errorf("unknown tile character: %s", char)
	}

	m.Tiles[y][x] = tile
	return nil
}

// drawRect draws a rectangle (outline or filled).
func drawRect(m *game.GameMap, params string, filled bool) error {
	parts := strings.Split(params, ",")
	if len(parts) != 5 {
		return fmt.Errorf("need 5 parameters: x1,y1,x2,y2,CHAR")
	}

	x1, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	y1, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	x2, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
	y2, _ := strconv.Atoi(strings.TrimSpace(parts[3]))
	char := rune(parts[4][0])

	tile, ok := charToTile[char]
	if !ok {
		return fmt.Errorf("unknown tile character: %c", char)
	}

	// Ensure coordinates are in bounds
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 >= m.Width {
		x2 = m.Width - 1
	}
	if y2 >= m.Height {
		y2 = m.Height - 1
	}

	if filled {
		for y := y1; y <= y2; y++ {
			for x := x1; x <= x2; x++ {
				m.Tiles[y][x] = tile
			}
		}
	} else {
		// Draw outline
		for x := x1; x <= x2; x++ {
			m.Tiles[y1][x] = tile
			m.Tiles[y2][x] = tile
		}
		for y := y1; y <= y2; y++ {
			m.Tiles[y][x1] = tile
			m.Tiles[y][x2] = tile
		}
	}

	return nil
}

// outputMap writes the map to a file or stdout.
func outputMap(m *game.GameMap, outputFile string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	if outputFile == "" {
		fmt.Println(string(data))
		return nil
	}

	if err := os.WriteFile(outputFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Map saved to: %s\n", outputFile)
	return nil
}
