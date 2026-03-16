package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// BenchmarkConcurrentClients measures latency and throughput with 100+ concurrent WebSocket clients.
// This benchmark establishes baseline SLIs (Service Level Indicators) for network optimization decisions.
//
// Metrics measured:
// - Message latency (p50, p95, p99)
// - Throughput (messages per second)
// - Concurrent connection handling
func BenchmarkConcurrentClients(b *testing.B) {
	// Create server
	server, err := NewRPCServer(b.TempDir())
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() { _ = server.Shutdown(ctx) }()

	// Create test HTTP server
	ts := httptest.NewServer(server.createTestMux())
	defer ts.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + ts.URL[4:] + "/ws"

	clientCounts := []int{10, 50, 100}

	for _, numClients := range clientCounts {
		b.Run(fmt.Sprintf("clients_%d", numClients), func(b *testing.B) {
			var latencies []time.Duration
			var latencyMu sync.Mutex
			var successCount int64
			var failCount int64

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup

				for c := 0; c < numClients; c++ {
					wg.Add(1)
					go func(clientID int) {
						defer wg.Done()

						// Create session first via HTTP
						sessionID, err := createBenchSession(ts.URL, clientID)
						if err != nil {
							atomic.AddInt64(&failCount, 1)
							return
						}

						// Connect via WebSocket with session cookie
						conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
							"Cookie": []string{fmt.Sprintf("session_id=%s", sessionID)},
						})
						if err != nil {
							atomic.AddInt64(&failCount, 1)
							return
						}
						defer conn.Close()

						// Read session confirmation
						var confirmResp map[string]interface{}
						if err := conn.ReadJSON(&confirmResp); err != nil {
							atomic.AddInt64(&failCount, 1)
							return
						}

						// Send RPC request and measure latency
						start := time.Now()
						req := map[string]interface{}{
							"jsonrpc": "2.0",
							"method":  "getGameState",
							"params":  map[string]interface{}{"session_id": sessionID},
							"id":      1,
						}

						if err := conn.WriteJSON(req); err != nil {
							atomic.AddInt64(&failCount, 1)
							return
						}

						var resp map[string]interface{}
						if err := conn.ReadJSON(&resp); err != nil {
							atomic.AddInt64(&failCount, 1)
							return
						}
						latency := time.Since(start)

						latencyMu.Lock()
						latencies = append(latencies, latency)
						latencyMu.Unlock()

						atomic.AddInt64(&successCount, 1)
					}(c)
				}

				wg.Wait()
			}

			// Report metrics
			if len(latencies) > 0 {
				sort.Slice(latencies, func(i, j int) bool {
					return latencies[i] < latencies[j]
				})

				p50 := latencies[len(latencies)*50/100]
				p95 := latencies[len(latencies)*95/100]
				p99 := latencies[len(latencies)*99/100]

				b.ReportMetric(float64(p50.Microseconds()), "p50_us")
				b.ReportMetric(float64(p95.Microseconds()), "p95_us")
				b.ReportMetric(float64(p99.Microseconds()), "p99_us")
				b.ReportMetric(float64(successCount), "success_count")
				b.ReportMetric(float64(failCount), "fail_count")
			}
		})
	}
}

// BenchmarkWebSocketLatency measures raw WebSocket message latency without concurrent load.
func BenchmarkWebSocketLatency(b *testing.B) {
	bc, cleanup := setupBenchServer(b)
	defer cleanup()

	req := makeGameStateRequest(bc.sessionID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := bc.conn.WriteJSON(req); err != nil {
			b.Fatalf("Write failed: %v", err)
		}

		var resp map[string]interface{}
		if err := bc.conn.ReadJSON(&resp); err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}

// BenchmarkWebSocketThroughput measures maximum message throughput on a single connection.
func BenchmarkWebSocketThroughput(b *testing.B) {
	bc, cleanup := setupBenchServer(b)
	defer cleanup()

	req := makeGameStateRequest(bc.sessionID)

	start := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := bc.conn.WriteJSON(req); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
		var resp map[string]interface{}
		if err := bc.conn.ReadJSON(&resp); err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
	duration := time.Since(start)

	b.ReportMetric(float64(b.N)/duration.Seconds(), "msgs/sec")
}

// BenchmarkMessagePayloadSize measures bandwidth usage per message type.
func BenchmarkMessagePayloadSize(b *testing.B) {
	tests := []struct {
		name    string
		method  string
		params  map[string]interface{}
		wantMax int // Maximum expected bytes
	}{
		{
			name:    "getGameState",
			method:  "getGameState",
			params:  map[string]interface{}{"session_id": "test"},
			wantMax: 10240, // 10KB
		},
		{
			name:    "getCharacter",
			method:  "getCharacter",
			params:  map[string]interface{}{"session_id": "test"},
			wantMax: 4096, // 4KB
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			req := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  tt.method,
				"params":  tt.params,
				"id":      1,
			}

			reqBytes, _ := json.Marshal(req)
			b.ReportMetric(float64(len(reqBytes)), "req_bytes")
		})
	}
}

// BenchmarkConnectionEstablishment measures WebSocket connection setup time.
func BenchmarkConnectionEstablishment(b *testing.B) {
	server, err := NewRPCServer(b.TempDir())
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() { _ = server.Shutdown(ctx) }()

	ts := httptest.NewServer(server.createTestMux())
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:] + "/ws"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create session
		sessionID, err := createBenchSession(ts.URL, i)
		if err != nil {
			b.Fatalf("Failed to create session: %v", err)
		}

		// Connect
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
			"Cookie": []string{fmt.Sprintf("session_id=%s", sessionID)},
		})
		if err != nil {
			b.Fatalf("Failed to connect: %v", err)
		}

		// Read confirmation
		var confirmResp map[string]interface{}
		if err := conn.ReadJSON(&confirmResp); err != nil {
			b.Fatalf("Failed to read confirmation: %v", err)
		}

		conn.Close()
	}
}

// createTestMux creates a simple mux for benchmark testing.
func (s *RPCServer) createTestMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Session creation endpoint
	mux.HandleFunc("/api/session", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Create session directly without HTTP context
		session, err := s.getOrCreateSession(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"session_id": session.SessionID})
	})

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Extract session ID from cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "No session cookie", http.StatusUnauthorized)
			return
		}

		s.mu.RLock()
		session, exists := s.sessions[cookie.Value]
		s.mu.RUnlock()

		if !exists {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		// Add session to context
		ctx := context.WithValue(r.Context(), sessionKey, session)
		r = r.WithContext(ctx)

		s.HandleWebSocket(w, r)
	})

	// RPC endpoint
	mux.HandleFunc("/rpc", s.handleRPCBench)

	return mux
}

// handleRPCBench handles RPC requests for benchmarking.
func (s *RPCServer) handleRPCBench(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONRPCError(w, nil, JSONRPCParseError, "Parse error")
		return
	}

	paramsJSON, _ := json.Marshal(req.Params)
	result, err := s.handleMethod(RPCMethod(req.Method), paramsJSON)
	if err != nil {
		writeJSONRPCError(w, req.ID, -32000, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  result,
		"id":      req.ID,
	})
}

// createBenchSession creates a session for benchmarking via HTTP.
func createBenchSession(baseURL string, clientID int) (string, error) {
	resp, err := http.Post(baseURL+"/api/session", "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("session creation failed: %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["session_id"], nil
}

// writeJSONRPCError writes a JSON-RPC error response.
func writeJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
		"id": id,
	})
}

// benchServerSetup creates a test server with WebSocket support and returns cleanup function.
// This consolidates the repeated setup pattern across benchmark tests.
type benchContext struct {
	server    *RPCServer
	ts        *httptest.Server
	wsURL     string
	sessionID string
	conn      *websocket.Conn
	cancel    context.CancelFunc
}

// setupBenchServer creates a test server and WebSocket connection for benchmarks.
// Call cleanup() when done to release resources.
func setupBenchServer(b *testing.B) (*benchContext, func()) {
	server, err := NewRPCServer(b.TempDir())
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	ts := httptest.NewServer(server.createTestMux())
	wsURL := "ws" + ts.URL[4:] + "/ws"

	sessionID, err := createBenchSession(ts.URL, 0)
	if err != nil {
		cancel()
		ts.Close()
		b.Fatalf("Failed to create session: %v", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
		"Cookie": []string{fmt.Sprintf("session_id=%s", sessionID)},
	})
	if err != nil {
		cancel()
		ts.Close()
		b.Fatalf("Failed to connect: %v", err)
	}

	// Read session confirmation
	var confirmResp map[string]interface{}
	if err := conn.ReadJSON(&confirmResp); err != nil {
		conn.Close()
		cancel()
		ts.Close()
		b.Fatalf("Failed to read confirmation: %v", err)
	}

	bc := &benchContext{
		server:    server,
		ts:        ts,
		wsURL:     wsURL,
		sessionID: sessionID,
		conn:      conn,
		cancel:    cancel,
	}

	cleanup := func() {
		conn.Close()
		ts.Close()
		_ = server.Shutdown(ctx)
		cancel()
	}

	return bc, cleanup
}

// makeGameStateRequest returns a standard getGameState request for benchmarks.
func makeGameStateRequest(sessionID string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getGameState",
		"params":  map[string]interface{}{"session_id": sessionID},
		"id":      1,
	}
}

// TestConcurrentClients_P95Latency validates that p95 latency is under 100ms with 100 clients.
// This is the acceptance criteria specified in AUDIT.md.
// This test requires significant resources and may timeout on CI; run with LOAD_TEST=1 to enable.
func TestConcurrentClients_P95Latency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	// Skip unless explicitly enabled due to resource requirements
	if os.Getenv("LOAD_TEST") == "" {
		t.Skip("Skipping load test (set LOAD_TEST=1 to enable)")
	}

	server, err := NewRPCServer(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	ts := httptest.NewServer(server.createTestMux())
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:] + "/ws"

	numClients := 100
	requestsPerClient := 5

	var latencies []time.Duration
	var latencyMu sync.Mutex
	var wg sync.WaitGroup

	// Use a dialer with timeouts to prevent hanging
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	// Create a channel to signal all clients to stop
	stopCh := make(chan struct{})

	for c := 0; c < numClients; c++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			sessionID, err := createBenchSession(ts.URL, clientID)
			if err != nil {
				return
			}

			conn, _, err := dialer.Dial(wsURL, http.Header{
				"Cookie": []string{fmt.Sprintf("session_id=%s", sessionID)},
			})
			if err != nil {
				return
			}
			defer conn.Close()

			// Set read/write deadlines to prevent hanging
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			// Read confirmation
			var confirmResp map[string]interface{}
			if err := conn.ReadJSON(&confirmResp); err != nil {
				return
			}

			// Send multiple requests
			for r := 0; r < requestsPerClient; r++ {
				select {
				case <-stopCh:
					return
				case <-ctx.Done():
					return
				default:
				}

				// Reset deadlines for each request
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

				start := time.Now()
				req := map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "getGameState",
					"params":  map[string]interface{}{"session_id": sessionID},
					"id":      r,
				}

				if err := conn.WriteJSON(req); err != nil {
					continue
				}

				var resp map[string]interface{}
				if err := conn.ReadJSON(&resp); err != nil {
					continue
				}

				latency := time.Since(start)
				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()
			}
		}(c)
	}

	// Wait with timeout to prevent hanging
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines finished
	case <-time.After(30 * time.Second):
		t.Log("Test timed out waiting for clients")
		close(stopCh) // Signal all clients to stop
	case <-ctx.Done():
		t.Log("Context cancelled")
		close(stopCh)
	}

	if len(latencies) == 0 {
		t.Fatal("No successful requests recorded")
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p95Index := len(latencies) * 95 / 100
	p95 := latencies[p95Index]

	t.Logf("Latency results (n=%d):", len(latencies))
	t.Logf("  p50: %v", latencies[len(latencies)*50/100])
	t.Logf("  p95: %v", p95)
	t.Logf("  p99: %v", latencies[len(latencies)*99/100])

	// AUDIT.md specifies: 100 clients with <100ms p95 latency
	maxP95 := 100 * time.Millisecond
	if p95 > maxP95 {
		t.Errorf("P95 latency %v exceeds target %v", p95, maxP95)
	}
}

// =============================================================================
// WebSocket Library Benchmarks
// =============================================================================
//
// These benchmarks test WebSocket performance. The server uses nhooyr.io/websocket
// while the benchmark client uses gorilla/websocket (any WebSocket client works).
//
// Running benchmarks:
//   go test ./pkg/server/... -bench=BenchmarkWebSocket -benchmem
// =============================================================================

// BenchmarkWebSocketRoundTrip measures single-message round-trip time.
// This is the primary latency benchmark for library comparison.
func BenchmarkWebSocketRoundTrip(b *testing.B) {
	bc, cleanup := setupBenchServer(b)
	defer cleanup()

	req := makeGameStateRequest(bc.sessionID)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := bc.conn.WriteJSON(req); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
		var resp map[string]interface{}
		if err := bc.conn.ReadJSON(&resp); err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}

// BenchmarkWebSocketWriteOnly measures write-only throughput.
// Useful for comparing message serialization performance.
func BenchmarkWebSocketWriteOnly(b *testing.B) {
	bc, cleanup := setupBenchServer(b)
	defer cleanup()

	req := makeGameStateRequest(bc.sessionID)
	data, _ := json.Marshal(req)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := bc.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
		// Drain response to prevent buffer buildup
		bc.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		bc.conn.ReadMessage()
	}
}

// BenchmarkWebSocketConcurrentWrites measures concurrent write performance.
// Tests for contention in write path.
func BenchmarkWebSocketConcurrentWrites(b *testing.B) {
	server, err := NewRPCServer(b.TempDir())
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() { _ = server.Shutdown(ctx) }()

	ts := httptest.NewServer(server.createTestMux())
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:] + "/ws"

	// Create 10 connections for concurrent testing
	const numConns = 10
	conns := make([]*websocket.Conn, numConns)

	for i := 0; i < numConns; i++ {
		sessionID, err := createBenchSession(ts.URL, i)
		if err != nil {
			b.Fatalf("Failed to create session %d: %v", i, err)
		}

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
			"Cookie": []string{fmt.Sprintf("session_id=%s", sessionID)},
		})
		if err != nil {
			b.Fatalf("Failed to connect %d: %v", i, err)
		}
		conns[i] = conn

		var confirmResp map[string]interface{}
		if err := conn.ReadJSON(&confirmResp); err != nil {
			b.Fatalf("Failed to read confirmation %d: %v", i, err)
		}
	}
	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getGameState",
		"params":  map[string]interface{}{"session_id": "test"},
		"id":      1,
	}

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		for j := 0; j < numConns; j++ {
			wg.Add(1)
			go func(conn *websocket.Conn) {
				defer wg.Done()
				conn.WriteJSON(req)
				var resp map[string]interface{}
				conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				conn.ReadJSON(&resp)
			}(conns[j])
		}
		wg.Wait()
	}

	b.ReportMetric(float64(numConns), "conns")
}

// BenchmarkWebSocketMessageSizes measures performance across different message sizes.
// Helps identify optimal payload sizes for the application.
func BenchmarkWebSocketMessageSizes(b *testing.B) {
	bc, cleanup := setupBenchServer(b)
	defer cleanup()

	sizes := []int{64, 256, 1024, 4096, 16384}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// Create payload of target size
			payload := make([]byte, size)
			for i := range payload {
				payload[i] = byte('A' + (i % 26))
			}

			b.ResetTimer()
			b.ReportAllocs()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				if err := bc.conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
					b.Fatalf("Write failed: %v", err)
				}
				// Drain response
				bc.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				bc.conn.ReadMessage()
			}
		})
	}
}

// BenchmarkWebSocketConnectionSetup measures connection establishment overhead.
// Compares the upgrade handshake performance between libraries.
func BenchmarkWebSocketConnectionSetup(b *testing.B) {
	server, err := NewRPCServer(b.TempDir())
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() { _ = server.Shutdown(ctx) }()

	ts := httptest.NewServer(server.createTestMux())
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:] + "/ws"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sessionID, err := createBenchSession(ts.URL, i)
		if err != nil {
			b.Fatalf("Failed to create session: %v", err)
		}

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
			"Cookie": []string{fmt.Sprintf("session_id=%s", sessionID)},
		})
		if err != nil {
			b.Fatalf("Failed to connect: %v", err)
		}

		var confirmResp map[string]interface{}
		if err := conn.ReadJSON(&confirmResp); err != nil {
			b.Fatalf("Failed to read confirmation: %v", err)
		}

		conn.Close()
	}
}

// BenchmarkWebSocketMemoryProfile measures memory allocation patterns.
// Useful for identifying memory pressure differences between libraries.
func BenchmarkWebSocketMemoryProfile(b *testing.B) {
	bc, cleanup := setupBenchServer(b)
	defer cleanup()

	req := makeGameStateRequest(bc.sessionID)

	b.ResetTimer()
	b.ReportAllocs()

	// Run 100 iterations per benchmark iteration to get stable memory allocation numbers
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			if err := bc.conn.WriteJSON(req); err != nil {
				b.Fatalf("Write failed: %v", err)
			}
			var resp map[string]interface{}
			if err := bc.conn.ReadJSON(&resp); err != nil {
				b.Fatalf("Read failed: %v", err)
			}
		}
	}

	b.ReportMetric(100, "ops/iter")
}
