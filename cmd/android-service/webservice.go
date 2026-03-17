// FILENAME: webservice.go
// PURPOSE: Go HTTP server with dual network binding for Android embedding.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const defaultListenPort = 8080

type statusResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Host      string `json:"host"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello from Go service")
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		resp := statusResponse{
			Status:    "running",
			Timestamp: time.Now().Format(time.RFC3339),
			Host:      hostname,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode status response: %v", err)
		}
	})

	mux.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		ip := getLANIP()
		fmt.Fprintln(w, ip)
	})

	// Configure bind address and port with safe defaults.
	bindAddr := os.Getenv("GOLDBOX_BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}

	port := defaultListenPort
	if portStr := os.Getenv("GOLDBOX_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err != nil {
			log.Printf("Invalid GOLDBOX_PORT %q, using default %d", portStr, defaultListenPort)
		} else if p <= 0 || p > 65535 {
			log.Printf("GOLDBOX_PORT out of range (%d), using default %d", p, defaultListenPort)
		} else {
			port = p
		}
	}

	addr := fmt.Sprintf("%s:%d", bindAddr, port)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	log.Printf("Starting Go web service on %s\n", addr)
	if bindAddr == "0.0.0.0" {
		if ip := getLANIP(); ip != "" {
			log.Printf("LAN access:  http://%s:%d\n", ip, port)
		}
	}
	log.Printf("Local access: http://127.0.0.1:%d\n", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v\n", err)
	}
	log.Println("Server stopped.")
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
