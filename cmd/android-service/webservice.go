// FILENAME: webservice.go
// PURPOSE: Go HTTP server with dual network binding for Android embedding.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const listenPort = 8080

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
		hostname, _ := os.Hostname()
		resp := statusResponse{
			Status:    "running",
			Timestamp: time.Now().Format(time.RFC3339),
			Host:      hostname,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		ip := getLANIP()
		fmt.Fprintln(w, ip)
	})

	addr := fmt.Sprintf("0.0.0.0:%d", listenPort)
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
		server.Close()
	}()

	log.Printf("Starting Go web service on %s\n", addr)
	if ip := getLANIP(); ip != "" {
		log.Printf("LAN access: http://%s:%d\n", ip, listenPort)
	}
	log.Printf("Local access: http://127.0.0.1:%d\n", listenPort)

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
