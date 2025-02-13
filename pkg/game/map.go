package game

import "encoding/json"

type MapTile struct {
	SpriteX     int  `json:"spriteX"`
	SpriteY     int  `json:"spriteY"`
	Walkable    bool `json:"walkable"`
	Transparent bool `json:"transparent"`
}

type GameMap struct {
	Width  int         `json:"width"`
	Height int         `json:"height"`
	Tiles  [][]MapTile `json:"tiles"`
}

func (m *GameMap) GetTile(x, y int) *MapTile {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return nil
	}
	return &m.Tiles[y][x]
}

func (m *GameMap) MarshalJSON() ([]byte, error) {
	type Alias GameMap
	return json.Marshal(&struct {
		*Alias
		GetTile string `json:"getTile"`
	}{
		Alias:   (*Alias)(m),
		GetTile: "function(x,y) { return this.tiles[y][x]; }",
	})
}
