package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
	"goldbox-rpg/pkg/server"
)

func main() {
	cfg := loadAndConfigureSystem()

	// Check if zero-configuration bootstrap is needed
	dataDir := cfg.DataDir
	if !pcg.DetectConfigurationPresence(dataDir) {
		logrus.Info("No existing configuration detected, initializing zero-configuration bootstrap")

		if err := initializeBootstrapGame(cfg, dataDir); err != nil {
			logrus.WithError(err).Fatal("Failed to initialize bootstrap game")
		}

		logrus.Info("Zero-configuration bootstrap completed successfully")
	}

	srv, listener := initializeServer(cfg)
	executeServerLifecycle(cfg, srv, listener)
}

// initializeBootstrapGame creates a complete game using zero-configuration bootstrap
func initializeBootstrapGame(cfg *config.Config, dataDir string) error {
	// Create a basic world instance for PCG
	world := game.NewWorld()

	// Use default bootstrap configuration
	bootstrapConfig := pcg.DefaultBootstrapConfig()
	bootstrapConfig.DataDirectory = dataDir

	// Initialize bootstrap system
	bootstrap := pcg.NewBootstrap(bootstrapConfig, world, logrus.StandardLogger())

	// Generate complete game
	ctx, cancel := context.WithTimeout(context.Background(), cfg.BootstrapTimeout)
	defer cancel()

	_, err := bootstrap.GenerateCompleteGame(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap game generation failed: %w", err)
	}

	return nil
}

// loadAndConfigureSystem loads configuration and sets up logging.
func loadAndConfigureSystem() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	configureLogging(cfg.LogLevel)
	logStartupInfo(cfg)
	return cfg
}

// configureLogging sets up the logging system based on configuration.
func configureLogging(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
}

// logStartupInfo logs server startup information.
func logStartupInfo(cfg *config.Config) {
	logrus.WithFields(logrus.Fields{
		"port":           cfg.ServerPort,
		"webDir":         cfg.WebDir,
		"sessionTimeout": cfg.SessionTimeout,
		"logLevel":       cfg.LogLevel,
		"devMode":        cfg.EnableDevMode,
	}).Info("Starting GoldBox RPG Engine server")
}

// initializeServer creates the server and network listener.
func initializeServer(cfg *config.Config) (*server.RPCServer, net.Listener) {
	srv, err := server.NewRPCServer(cfg.WebDir)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize server")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ServerPort))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to start listener")
	}

	return srv, listener
}

// executeServerLifecycle handles the complete server lifecycle including startup and shutdown.
func executeServerLifecycle(cfg *config.Config, srv *server.RPCServer, listener net.Listener) {
	sigChan, errChan := setupShutdownHandling()
	startServerAsync(srv, listener, errChan)
	waitForShutdownSignal(sigChan, errChan)
	performGracefulShutdown(cfg, listener, srv)
}

// setupShutdownHandling creates channels for graceful shutdown signal handling.
func setupShutdownHandling() (chan os.Signal, chan error) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error, 1)
	return sigChan, errChan
}

// startServerAsync starts the server in a background goroutine with panic recovery.
// If the server panics, the error is captured and sent to errChan to trigger
// graceful shutdown rather than crashing the entire process.
func startServerAsync(srv *server.RPCServer, listener net.Listener, errChan chan error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Server goroutine panicked")
				errChan <- fmt.Errorf("server panicked: %v", r)
			}
		}()
		logrus.WithField("address", listener.Addr()).Info("Server listening")
		if err := srv.Serve(listener); err != nil {
			errChan <- fmt.Errorf("server failed: %w", err)
		}
	}()
}

// waitForShutdownSignal waits for either a shutdown signal or server error.
func waitForShutdownSignal(sigChan chan os.Signal, errChan chan error) {
	select {
	case sig := <-sigChan:
		logrus.WithField("signal", sig).Info("Received shutdown signal")
	case err := <-errChan:
		logrus.WithError(err).Error("Server error")
	}
}

// performGracefulShutdown handles the graceful server shutdown process.
func performGracefulShutdown(cfg *config.Config, listener net.Listener, srv *server.RPCServer) {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	logrus.Info("Shutting down server gracefully...")

	// Save game state before shutting down if persistence is enabled
	if cfg.EnablePersistence {
		logrus.Info("Saving game state before shutdown...")
		// Access the server's internal state to save it
		// We'll add a SaveState method to RPCServer
		if err := srv.SaveState(); err != nil {
			logrus.WithError(err).Error("Failed to save game state during shutdown")
		} else {
			logrus.Info("Game state saved successfully")
		}
	}

	if err := listener.Close(); err != nil {
		logrus.WithError(err).Warn("Error closing listener")
	}

	select {
	case <-shutdownCtx.Done():
		logrus.Warn("Shutdown timeout exceeded, forcing exit")
	case <-time.After(cfg.ShutdownGracePeriod):
		logrus.Info("Server shutdown completed")
	}
}
