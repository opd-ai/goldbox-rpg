package pcg

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/game"
)

func TestNewPCGManager(t *testing.T) {
	tests := []struct {
		name    string
		logger  *logrus.Logger
		wantNil bool
	}{
		{
			name:    "create manager with logger",
			logger:  logrus.New(),
			wantNil: false,
		},
		{
			name:    "create manager without logger",
			logger:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			manager := NewPCGManager(world, tt.logger)

			assert.NotNil(t, manager)
			assert.NotNil(t, manager.registry)
			assert.NotNil(t, manager.factory)
			assert.NotNil(t, manager.validator)
			assert.NotNil(t, manager.logger)
			assert.NotNil(t, manager.seedManager)
			assert.NotNil(t, manager.metrics)
			assert.NotNil(t, manager.qualityMetrics)
		})
	}
}

func TestPCGManager_InitializeWithSeed(t *testing.T) {
	tests := []struct {
		name string
		seed int64
	}{
		{name: "positive seed", seed: 12345},
		{name: "zero seed", seed: 0},
		{name: "negative seed", seed: -999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			manager := NewPCGManager(world, logrus.New())

			manager.InitializeWithSeed(tt.seed)
			assert.NotNil(t, manager.seedManager)
		})
	}
}

func TestPCGManager_RegisterDefaultGenerators(t *testing.T) {
	world := game.NewWorld()
	manager := NewPCGManager(world, logrus.New())

	err := manager.RegisterDefaultGenerators()
	assert.NoError(t, err)
}

func TestPCGManager_GenerateTerrainForLevel(t *testing.T) {
	tests := []struct {
		name       string
		levelID    string
		width      int
		height     int
		biome      BiomeType
		difficulty int
		timeout    time.Duration
		wantErr    bool
	}{
		{
			name:       "forest level",
			levelID:    "level-1",
			width:      50,
			height:     50,
			biome:      BiomeForest,
			difficulty: 1,
			timeout:    10 * time.Second,
			wantErr:    false,
		},
		{
			name:       "dungeon level",
			levelID:    "level-2",
			width:      40,
			height:     40,
			biome:      BiomeDungeon,
			difficulty: 2,
			timeout:    10 * time.Second,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			manager := NewPCGManager(world, logrus.New())
			manager.InitializeWithSeed(12345)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			gameMap, err := manager.GenerateTerrainForLevel(ctx, tt.levelID, tt.width, tt.height, tt.biome, tt.difficulty)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gameMap)
			}
		})
	}
}

func TestPCGManager_GenerateItemsForLocation(t *testing.T) {
	tests := []struct {
		name        string
		locationID  string
		itemCount   int
		minRarity   RarityTier
		maxRarity   RarityTier
		playerLevel int
		timeout     time.Duration
		wantErr     bool
	}{
		{
			name:        "common items",
			locationID:  "loc-1",
			itemCount:   5,
			minRarity:   RarityCommon,
			maxRarity:   RarityUncommon,
			playerLevel: 1,
			timeout:     10 * time.Second,
			wantErr:     false,
		},
		{
			name:        "rare items",
			locationID:  "loc-2",
			itemCount:   3,
			minRarity:   RarityRare,
			maxRarity:   RarityEpic,
			playerLevel: 5,
			timeout:     10 * time.Second,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			manager := NewPCGManager(world, logrus.New())
			manager.InitializeWithSeed(54321)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			items, err := manager.GenerateItemsForLocation(ctx, tt.locationID, tt.itemCount, tt.minRarity, tt.maxRarity, tt.playerLevel)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, items)
			}
		})
	}
}

func TestPCGManager_GenerateDungeonLevel(t *testing.T) {
	tests := []struct {
		name       string
		levelID    string
		minRooms   int
		maxRooms   int
		theme      LevelTheme
		difficulty int
		timeout    time.Duration
		wantErr    bool
	}{
		{
			name:       "small dungeon",
			levelID:    "dungeon-1",
			minRooms:   3,
			maxRooms:   5,
			theme:      ThemeClassic,
			difficulty: 1,
			timeout:    15 * time.Second,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			manager := NewPCGManager(world, logrus.New())
			manager.InitializeWithSeed(99999)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			level, err := manager.GenerateDungeonLevel(ctx, tt.levelID, tt.minRooms, tt.maxRooms, tt.theme, tt.difficulty)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, level)
			}
		})
	}
}

func TestPCGManager_GenerateQuestForArea(t *testing.T) {
	tests := []struct {
		name        string
		areaID      string
		questType   QuestType
		playerLevel int
		timeout     time.Duration
		wantErr     bool
	}{
		{
			name:        "kill quest",
			areaID:      "area-1",
			questType:   QuestTypeKill,
			playerLevel: 1,
			timeout:     15 * time.Second,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			manager := NewPCGManager(world, logrus.New())
			manager.InitializeWithSeed(777)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			quest, err := manager.GenerateQuestForArea(ctx, tt.areaID, tt.questType, tt.playerLevel)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, quest)
			}
		})
	}
}

func TestPCGManager_ValidateGeneratedContent(t *testing.T) {
	tests := []struct {
		name    string
		content interface{}
		wantErr bool
	}{
		{
			name:    "validate game map",
			content: &game.GameMap{},
			wantErr: false,
		},
		{
			name:    "validate item",
			content: &game.Item{},
			wantErr: false,
		},
		{
			name:    "validate level",
			content: &game.Level{},
			wantErr: false,
		},
		{
			name:    "validate quest",
			content: &game.Quest{},
			wantErr: false,
		},
		{
			name:    "unsupported type",
			content: "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			manager := NewPCGManager(world, logrus.New())

			result, err := manager.ValidateGeneratedContent(tt.content)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestPCGManager_GetMetrics(t *testing.T) {
	world := game.NewWorld()
	manager := NewPCGManager(world, logrus.New())

	metrics := manager.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestPCGManager_GetQualityMetrics(t *testing.T) {
	world := game.NewWorld()
	manager := NewPCGManager(world, logrus.New())

	qualityMetrics := manager.GetQualityMetrics()
	assert.NotNil(t, qualityMetrics)
}

func TestPCGManager_GetRegistry(t *testing.T) {
	world := game.NewWorld()
	manager := NewPCGManager(world, logrus.New())

	registry := manager.GetRegistry()
	assert.NotNil(t, registry)
}

func TestPCGManager_ConcurrentGeneration(t *testing.T) {
	world := game.NewWorld()
	manager := NewPCGManager(world, logrus.New())
	manager.InitializeWithSeed(99999)

	goroutineCount := 5
	generationsEach := 2
	done := make(chan bool, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			for j := 0; j < generationsEach; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				_, err := manager.GenerateTerrainForLevel(ctx, "concurrent-level", 30, 30, BiomeForest, 1)
				cancel()
				require.NoError(t, err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutineCount; i++ {
		<-done
	}
}

func TestPCGManager_ContextCancellation(t *testing.T) {
	world := game.NewWorld()
	manager := NewPCGManager(world, logrus.New())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := manager.GenerateTerrainForLevel(ctx, "cancel-test", 100, 100, BiomeForest, 1)
	assert.Error(t, err)
}
