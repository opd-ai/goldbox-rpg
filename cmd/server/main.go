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
	"goldbox-rpg/pkg/server"
)

func main() {
	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Configure logging based on config
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	logrus.WithFields(logrus.Fields{
		"port":           cfg.ServerPort,
		"webDir":         cfg.WebDir,
		"sessionTimeout": cfg.SessionTimeout,
		"logLevel":       cfg.LogLevel,
		"devMode":        cfg.EnableDevMode,
	}).Info("Starting GoldBox RPG Engine server")

	// Create new server instance with configuration
	srv, err := server.NewRPCServer(cfg.WebDir)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize server")
	}

	// Create listener with configured port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ServerPort))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to start listener")
	}

	// Set up graceful shutdown handling
	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		logrus.WithField("address", listener.Addr()).Info("Server listening")
		if err := srv.Serve(listener); err != nil {
			errChan <- fmt.Errorf("server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		logrus.WithField("signal", sig).Info("Received shutdown signal")
	case err := <-errChan:
		logrus.WithError(err).Error("Server error")
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logrus.Info("Shutting down server gracefully...")

	// Close the listener to stop accepting new connections
	if err := listener.Close(); err != nil {
		logrus.WithError(err).Warn("Error closing listener")
	}

	// Wait for shutdown or timeout
	select {
	case <-shutdownCtx.Done():
		logrus.Warn("Shutdown timeout exceeded, forcing exit")
	case <-time.After(1 * time.Second):
		logrus.Info("Server shutdown completed")
	}
}
