package server

import (
	"os"
	"testing"
	"time"

	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/validation"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONRPCError(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		data    interface{}
	}{
		{
			name:    "parse error",
			code:    JSONRPCParseError,
			message: "Parse error",
			data:    nil,
		},
		{
			name:    "invalid request with data",
			code:    JSONRPCInvalidRequest,
			message: "Invalid request",
			data:    map[string]string{"detail": "missing field"},
		},
		{
			name:    "method not found",
			code:    JSONRPCMethodNotFound,
			message: "Method not found",
			data:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewJSONRPCError(tt.code, tt.message, tt.data)
			assert.NotNil(t, err)
			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.data, err.Data)
			assert.Equal(t, tt.message, err.Error())
		})
	}
}

func TestJSONRPCError_Error(t *testing.T) {
	err := &JSONRPCError{
		Code:    -32600,
		Message: "Test error message",
		Data:    nil,
	}

	assert.Equal(t, "Test error message", err.Error())
}

func TestLoadServerConfiguration(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name    string
		setup   func()
		cleanup func()
		wantErr bool
	}{
		{
			name: "successful configuration load",
			setup: func() {
				os.Setenv("GOLDBOX_PORT", "8080")
				os.Setenv("GOLDBOX_LOG_LEVEL", "info")
			},
			cleanup: func() {
				os.Unsetenv("GOLDBOX_PORT")
				os.Unsetenv("GOLDBOX_LOG_LEVEL")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			cfg, validator, err := loadServerConfiguration(logger)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.NotNil(t, validator)
		})
	}
}

func TestInitializeSpellManager(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	for i := 0; i < 5; i++ {
		if _, err := os.Stat("data/spells"); err == nil {
			break
		}
		os.Chdir("..")
	}

	t.Run("initialize spell manager successfully", func(t *testing.T) {
		spellManager, err := initializeSpellManager(logger)
		if err != nil {
			t.Logf("Spell manager initialization skipped: %v", err)
			return
		}
		assert.NoError(t, err)
		assert.NotNil(t, spellManager)
		assert.Greater(t, spellManager.GetSpellCount(), 0)
	})
}

func TestSetupPCGManager(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("setup PCG manager with default generators", func(t *testing.T) {
		pcgManager, err := setupPCGManager(logger)
		assert.NoError(t, err)
		assert.NotNil(t, pcgManager)
		assert.NotNil(t, pcgManager.GetRegistry())
	})
}

func TestCreateServerInstance(t *testing.T) {
	webDir := "./web"
	cfg := &config.Config{
		ServerPort:     8080,
		LogLevel:       "info",
		MaxRequestSize: 1048576,
		SessionTimeout: 30 * time.Minute,
		AllowedOrigins: []string{"*"},
	}
	validator := validation.NewInputValidator(cfg.MaxRequestSize)
	spellManager := game.NewSpellManager("data/spells")

	logger := logrus.NewEntry(logrus.New())
	pcgManager, err := setupPCGManager(logger)
	require.NoError(t, err)

	server := createServerInstance(webDir, cfg, validator, spellManager, pcgManager)

	assert.NotNil(t, server)
	assert.Equal(t, webDir, server.webDir)
	assert.NotNil(t, server.fileServer)
	assert.NotNil(t, server.state)
	assert.NotNil(t, server.eventSys)
	assert.NotNil(t, server.sessions)
	assert.NotNil(t, server.timekeeper)
	assert.NotNil(t, server.done)
	assert.NotNil(t, server.spellManager)
	assert.NotNil(t, server.pcgManager)
	assert.NotNil(t, server.config)
	assert.NotNil(t, server.validator)
	assert.NotNil(t, server.state.WorldState)
	assert.NotNil(t, server.state.TurnManager)
	assert.NotNil(t, server.state.TimeManager)
	assert.NotNil(t, server.state.Sessions)
	assert.Equal(t, 1, server.state.Version)
}

func TestConfigurePerformanceMonitoring(t *testing.T) {
	webDir := "./web"
	cfg := &config.Config{
		ServerPort:       8080,
		LogLevel:         "info",
		MaxRequestSize:   1048576,
		SessionTimeout:   30 * time.Minute,
		EnableProfiling:  true,
		EnableDevMode:    true,
		AlertingEnabled:  true,
		MetricsInterval:  10 * time.Second,
		AlertingInterval: 30 * time.Second,
		AllowedOrigins:   []string{"*"},
	}
	validator := validation.NewInputValidator(cfg.MaxRequestSize)
	spellManager := game.NewSpellManager("data/spells")

	logger := logrus.NewEntry(logrus.New())
	pcgManager, err := setupPCGManager(logger)
	require.NoError(t, err)

	tests := []struct {
		name      string
		alerting  bool
		profiling bool
	}{
		{
			name:      "with alerting and profiling enabled",
			alerting:  true,
			profiling: true,
		},
		{
			name:      "without alerting",
			alerting:  false,
			profiling: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCfg := &config.Config{
				ServerPort:       cfg.ServerPort,
				WebDir:           cfg.WebDir,
				SessionTimeout:   cfg.SessionTimeout,
				LogLevel:         cfg.LogLevel,
				MaxRequestSize:   cfg.MaxRequestSize,
				EnableDevMode:    cfg.EnableDevMode,
				RequestTimeout:   cfg.RequestTimeout,
				AlertingEnabled:  tt.alerting,
				AlertingInterval: cfg.AlertingInterval,
				EnableProfiling:  tt.profiling,
				AllowedOrigins:   cfg.AllowedOrigins,
			}

			testServer := createServerInstance(webDir, testCfg, validator, spellManager, pcgManager)
			configurePerformanceMonitoring(testServer, testCfg)

			assert.NotNil(t, testServer.metrics)
			assert.NotNil(t, testServer.healthChecker)
			assert.NotNil(t, testServer.profiling)
			assert.NotNil(t, testServer.perfMonitor)

			if tt.alerting {
				assert.NotNil(t, testServer.perfAlerter)
			}
		})
	}
}

func TestRPCServer_StateManagement(t *testing.T) {
	webDir := "./web"
	cfg := &config.Config{
		ServerPort:     8080,
		LogLevel:       "info",
		MaxRequestSize: 1048576,
		SessionTimeout: 30 * time.Minute,
		AllowedOrigins: []string{"*"},
	}
	validator := validation.NewInputValidator(cfg.MaxRequestSize)
	spellManager := game.NewSpellManager("data/spells")

	logger := logrus.NewEntry(logrus.New())
	pcgManager, err := setupPCGManager(logger)
	require.NoError(t, err)

	server := createServerInstance(webDir, cfg, validator, spellManager, pcgManager)

	t.Run("verify initial game state", func(t *testing.T) {
		assert.NotNil(t, server.state)
		assert.NotNil(t, server.state.WorldState)
		assert.NotNil(t, server.state.TurnManager)
		assert.NotNil(t, server.state.TimeManager)
		assert.NotNil(t, server.state.Sessions)
		assert.Equal(t, 1, server.state.Version)
	})

	t.Run("access state thread-safely", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 5; i++ {
			go func() {
				server.mu.RLock()
				_ = server.state.Version
				server.mu.RUnlock()
				done <- true
			}()
		}

		for i := 0; i < 5; i++ {
			go func(val int) {
				server.mu.Lock()
				server.state.Version = val
				server.mu.Unlock()
				done <- true
			}(i + 2)
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		assert.GreaterOrEqual(t, server.state.Version, 1)
	})
}

func TestRPCServer_EventSystem(t *testing.T) {
	webDir := "./web"
	cfg := &config.Config{
		ServerPort:     8080,
		LogLevel:       "info",
		MaxRequestSize: 1048576,
		SessionTimeout: 30 * time.Minute,
		AllowedOrigins: []string{"*"},
	}
	validator := validation.NewInputValidator(cfg.MaxRequestSize)
	spellManager := game.NewSpellManager("data/spells")

	logger := logrus.NewEntry(logrus.New())
	pcgManager, err := setupPCGManager(logger)
	require.NoError(t, err)

	server := createServerInstance(webDir, cfg, validator, spellManager, pcgManager)

	t.Run("event system is initialized", func(t *testing.T) {
		assert.NotNil(t, server.eventSys)
	})

	t.Run("can emit and subscribe to events", func(t *testing.T) {
		eventReceived := make(chan bool, 1)

		handler := func(event game.GameEvent) {
			eventReceived <- true
		}

		server.eventSys.Subscribe(game.EventDamage, handler)

		event := game.GameEvent{
			Type:      game.EventDamage,
			SourceID:  "test-source",
			TargetID:  "test-target",
			Timestamp: time.Now().Unix(),
			Data:      make(map[string]interface{}),
		}

		server.eventSys.Emit(event)

		select {
		case <-eventReceived:
		case <-time.After(1 * time.Second):
			t.Fatal("event was not received within timeout")
		}
	})
}

func TestRPCServer_PCGManager(t *testing.T) {
	webDir := "./web"
	cfg := &config.Config{
		ServerPort:     8080,
		LogLevel:       "info",
		MaxRequestSize: 1048576,
		SessionTimeout: 30 * time.Minute,
		AllowedOrigins: []string{"*"},
	}
	validator := validation.NewInputValidator(cfg.MaxRequestSize)
	spellManager := game.NewSpellManager("data/spells")

	logger := logrus.NewEntry(logrus.New())
	pcgManager, err := setupPCGManager(logger)
	require.NoError(t, err)

	server := createServerInstance(webDir, cfg, validator, spellManager, pcgManager)

	t.Run("PCG manager is initialized and accessible", func(t *testing.T) {
		assert.NotNil(t, server.pcgManager)
		assert.NotNil(t, server.pcgManager.GetRegistry())

		stats := server.pcgManager.GetGenerationStatistics()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "available_generators")
	})
}
