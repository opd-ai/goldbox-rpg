package config

import (
	"context"
	"os"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/integration"

	"gopkg.in/yaml.v3"
)

// LoadItems loads item definitions from a YAML file and returns them as a slice of game.Item.
// This function is protected by both circuit breaker and retry patterns to prevent cascade
// failures and handle transient file system issues.
//
// Parameters:
//   - filename: Path to the YAML file containing item definitions
//
// Returns:
//   - []game.Item: Slice of parsed item objects
//   - error: File read, YAML parsing, circuit breaker, or retry errors if any occurred
//
// The function reads the entire file contents and unmarshals them as YAML into a slice
// of game.Item structs. It handles error cases with automatic retry and circuit breaker protection:
//  1. Circuit breaker is open (too many recent failures)
//  2. File read errors (missing file, permissions, etc) with retry
//  3. YAML parsing errors (invalid format, missing required fields)
//
// Related types:
//   - game.Item: The target struct for item definitions
func LoadItems(filename string) ([]game.Item, error) {
	var items []game.Item
	ctx := context.Background()

	err := integration.ExecuteConfigOperation(ctx, func(ctx context.Context) error {
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
