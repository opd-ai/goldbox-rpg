// FILENAME: webservice.go
// PURPOSE: Full Gold Box RPG server for Android embedding, serving the WASM interface.
package main

import (
	"context"
	"fmt"
	"log"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	configureLogging(cfg.LogLevel)

	// Bootstrap game data when no existing configuration is present.
	if !pcg.DetectConfigurationPresence(cfg.DataDir) {
		logrus.Info("No existing configuration detected, running zero-configuration bootstrap")
		if err := bootstrapGame(cfg); err != nil {
			logrus.WithError(err).Fatal("Bootstrap failed")
		}
		logrus.Info("Bootstrap completed successfully")
	}

	srv, err := server.NewRPCServer(cfg.WebDir)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize server")
	}

	bindAddr := getBindAddress()
	addr := fmt.Sprintf("%s:%d", bindAddr, cfg.ServerPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to start listener")
	}

	setupGracefulShutdown(srv, listener, getShutdownTimeout(cfg.ShutdownTimeout))
	logServerStartup(bindAddr, cfg.ServerPort)

	if err := srv.Serve(listener); err != nil {
		logrus.WithError(err).Fatal("Server error")
	}
	logrus.Info("Server stopped.")
}

// getBindAddress returns the server bind address from environment or default.
func getBindAddress() string {
	if addr := os.Getenv("GOLDBOX_BIND_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1"
}

// getShutdownTimeout returns the shutdown timeout, defaulting to 5 seconds.
func getShutdownTimeout(configured time.Duration) time.Duration {
	if configured <= 0 {
		return 5 * time.Second
	}
	return configured
}

// setupGracefulShutdown sets up signal handling for graceful shutdown.
func setupGracefulShutdown(srv *server.RPCServer, listener net.Listener, timeout time.Duration) {
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logrus.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logrus.WithError(err).Warn("Error during server shutdown")
		}
		if err := listener.Close(); err != nil {
			logrus.WithError(err).Warn("Error closing listener")
		}
	}()
}

// logServerStartup logs the server startup message including LAN IP if binding to all interfaces.
func logServerStartup(bindAddr string, port int) {
	logrus.WithField("address", fmt.Sprintf("%s:%d", bindAddr, port)).Info("Starting Gold Box RPG server")
	if bindAddr == "0.0.0.0" {
		if ip := getLANIP(); ip != "" {
			logrus.Infof("LAN access:  http://%s:%d", ip, port)
		}
	}
	logrus.Infof("Local access: http://127.0.0.1:%d", port)
}

// configureLogging sets the logrus level from a string.
func configureLogging(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logrus.SetLevel(lvl)
}

// bootstrapGame generates initial game data using the PCG system.
func bootstrapGame(cfg *config.Config) error {
	world := game.NewWorld()
	bc := pcg.DefaultBootstrapConfig()
	bc.DataDirectory = cfg.DataDir
	bootstrap := pcg.NewBootstrap(bc, world, logrus.StandardLogger())

	ctx, cancel := context.WithTimeout(context.Background(), cfg.BootstrapTimeout)
	defer cancel()

	_, err := bootstrap.GenerateCompleteGame(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap game generation failed: %w", err)
	}
	return nil
}

func getLANIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip != nil {
				return ip.String()
			}
		}
	}
	return ""
}
