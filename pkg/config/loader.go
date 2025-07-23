package config

import (
	"context"
	"os"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/resilience"

	"gopkg.in/yaml.v3"
)

// LoadItems loads item definitions from a YAML file and returns them as a slice of game.Item.
// This function is protected by a circuit breaker to prevent cascade failures from file system issues.
//
// Parameters:
//   - filename: Path to the YAML file containing item definitions
//
// Returns:
//   - []game.Item: Slice of parsed item objects
//   - error: File read, YAML parsing, or circuit breaker errors if any occurred
//
// The function reads the entire file contents and unmarshals them as YAML into a slice
// of game.Item structs. It handles three main error cases:
//  1. Circuit breaker is open (too many recent failures)
//  2. File read errors (missing file, permissions, etc)
//  3. YAML parsing errors (invalid format, missing required fields)
//
// Related types:
//   - game.Item: The target struct for item definitions
func LoadItems(filename string) ([]game.Item, error) {
	var items []game.Item
	ctx := context.Background()

	err := resilience.ExecuteWithConfigLoaderCircuitBreaker(ctx, func(ctx context.Context) error {
		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(data, &items); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}
