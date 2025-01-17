package config

import (
	"os"

	"goldbox-rpg/pkg/game" // Updated from internal to pkg

	"gopkg.in/yaml.v3"
)

// LoadItems loads item definitions from a YAML file and returns them as a slice of game.Item.
//
// Parameters:
//   - filename: Path to the YAML file containing item definitions
//
// Returns:
//   - []game.Item: Slice of parsed item objects
//   - error: File read or YAML parsing errors if any occurred
//
// The function reads the entire file contents and unmarshals them as YAML into a slice
// of game.Item structs. It handles two main error cases:
//  1. File read errors (missing file, permissions, etc)
//  2. YAML parsing errors (invalid format, missing required fields)
//
// Related types:
//   - game.Item: The target struct for item definitions
func LoadItems(filename string) ([]game.Item, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var items []game.Item
	if err := yaml.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}
