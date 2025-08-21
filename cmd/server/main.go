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
	dataDir := "data" // Default data directory
	if !pcg.DetectConfigurationPresence(dataDir) {
		logrus.Info("No existing configuration detected, initializing zero-configuration bootstrap")

		if err := initializeBootstrapGame(dataDir); err != nil {
			logrus.WithError(err).Fatal("Failed to initialize bootstrap game")
		}

		logrus.Info("Zero-configuration bootstrap completed successfully")
	}

	srv, listener := initializeServer(cfg)
	executeServerLifecycle(srv, listener)
}

// initializeBootstrapGame creates a complete game using zero-configuration bootstrap
func initializeBootstrapGame(dataDir string) error {
	// Create a basic world instance for PCG
	world := game.NewWorld()

	// Use default bootstrap configuration
	bootstrapConfig := pcg.DefaultBootstrapConfig()
	bootstrapConfig.DataDirectory = dataDir

	// Initialize bootstrap system
	bootstrap := pcg.NewBootstrap(bootstrapConfig, world, logrus.StandardLogger())

	// Generate complete game
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
func executeServerLifecycle(srv *server.RPCServer, listener net.Listener) {
	sigChan, errChan := setupShutdownHandling()
	startServerAsync(srv, listener, errChan)
	waitForShutdownSignal(sigChan, errChan)
	performGracefulShutdown(listener)
}

// setupShutdownHandling creates channels for graceful shutdown signal handling.
func setupShutdownHandling() (chan os.Signal, chan error) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error, 1)
	return sigChan, errChan
}

// startServerAsync starts the server in a background goroutine.
func startServerAsync(srv *server.RPCServer, listener net.Listener, errChan chan error) {
	go func() {
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
func performGracefulShutdown(listener net.Listener) {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logrus.Info("Shutting down server gracefully...")

	if err := listener.Close(); err != nil {
		logrus.WithError(err).Warn("Error closing listener")
	}

	select {
	case <-shutdownCtx.Done():
		logrus.Warn("Shutdown timeout exceeded, forcing exit")
	case <-time.After(1 * time.Second):
		logrus.Info("Server shutdown completed")
	}
}
