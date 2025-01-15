package config

import (
	"os"

	"goldbox-rpg/pkg/game" // Updated from internal to pkg

	"gopkg.in/yaml.v3"
)

// LoadItems loads items from a YAML file
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
