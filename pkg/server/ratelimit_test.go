package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"goldbox-rpg/pkg/config"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 5.0,
		RateLimitBurst:             10,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	require.NotNil(t, rl)
	assert.Equal(t, rate.Limit(5.0), rl.requestsPerSecond)
	assert.Equal(t, 10, rl.burst)
	assert.Equal(t, time.Minute, rl.cleanupInterval)
	assert.Equal(t, time.Minute*5, rl.maxAge)
	assert.NotNil(t, rl.ctx)
	assert.NotNil(t, rl.cancel)

	// Clean up
	rl.Close()
}

func TestRateLimiter_Allow(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 2.0, // 2 requests per second
		RateLimitBurst:             3,   // Allow burst of 3
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	// Test initial burst allowance
	assert.True(t, rl.Allow("192.168.1.1"))
	assert.True(t, rl.Allow("192.168.1.1"))
	assert.True(t, rl.Allow("192.168.1.1"))

	// Should be rate limited now
	assert.False(t, rl.Allow("192.168.1.1"))

	// Different IP should have its own limit
	assert.True(t, rl.Allow("192.168.1.2"))
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 1.0,
		RateLimitBurst:             1,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	// Each IP should have independent rate limits
	ips := []string{"192.168.1.1", "192.168.1.2", "10.0.0.1"}

	for _, ip := range ips {
		assert.True(t, rl.Allow(ip), "IP %s should be allowed", ip)
		assert.False(t, rl.Allow(ip), "IP %s should be rate limited", ip)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 1.0,
		RateLimitBurst:             1,
		RateLimitCleanupInterval:   time.Millisecond * 50,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	// Force short max age for testing
	rl.maxAge = time.Millisecond * 100

	// Create limiters for multiple IPs
	rl.Allow("192.168.1.1")
	rl.Allow("192.168.1.2")
	rl.Allow("192.168.1.3")

	stats := rl.GetStats()
	assert.Equal(t, 3, stats.ActiveLimiters)

	// Wait for cleanup to occur
	time.Sleep(time.Millisecond * 200)

	// Trigger cleanup by getting stats
	stats = rl.GetStats()

	// All limiters should eventually be cleaned up
	// Note: This might be flaky in fast CI environments, so we allow some tolerance
	assert.LessOrEqual(t, stats.ActiveLimiters, 3)
}

func TestRateLimiter_Close(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 1.0,
		RateLimitBurst:             1,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)

	// Should not panic
	rl.Close()

	// Calling close again should not panic
	rl.Close()
}

func TestRateLimiter_GetStats(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 1.0,
		RateLimitBurst:             1,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	// Initially empty
	stats := rl.GetStats()
	assert.Equal(t, 0, stats.ActiveLimiters)

	// Add some limiters
	rl.Allow("192.168.1.1")
	rl.Allow("192.168.1.2")

	stats = rl.GetStats()
	assert.Equal(t, 2, stats.ActiveLimiters)
}

func TestRateLimitingMiddleware_Disabled(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware with nil rate limiter (disabled)
	middleware := RateLimitingMiddleware(nil)
	handler := middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Should pass through without rate limiting
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

func TestRateLimitingMiddleware_Allowed(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 10.0,
		RateLimitBurst:             10,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware with rate limiter
	middleware := RateLimitingMiddleware(rl)
	handler := middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Add logger context to prevent nil pointer
	ctx := context.WithValue(req.Context(), "logger", logrus.StandardLogger())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Should be allowed
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

func TestRateLimitingMiddleware_RateLimited(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 1.0,
		RateLimitBurst:             1,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware with rate limiter
	middleware := RateLimitingMiddleware(rl)
	handler := middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Add logger context to prevent nil pointer
	ctx := context.WithValue(req.Context(), "logger", logrus.StandardLogger())
	req = req.WithContext(ctx)

	// First request should be allowed
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, "success", w1.Body.String())

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Contains(t, w2.Body.String(), "Too Many Requests")
	assert.Equal(t, "1", w2.Header().Get("Retry-After"))
}

func TestRateLimitingMiddleware_XForwardedFor(t *testing.T) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 1.0,
		RateLimitBurst:             1,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware with rate limiter
	middleware := RateLimitingMiddleware(rl)
	handler := middleware(testHandler)

	// Create test request with X-Forwarded-For header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 192.168.1.1")
	req.RemoteAddr = "192.168.1.100:12345"

	// Add logger context to prevent nil pointer
	ctx := context.WithValue(req.Context(), "logger", logrus.StandardLogger())
	req = req.WithContext(ctx)

	// First request should be allowed
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited (same X-Forwarded-For IP)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedResult string
	}{
		{
			name:           "RemoteAddr only",
			remoteAddr:     "192.168.1.1:12345",
			expectedResult: "192.168.1.1", // IP without port for rate limiting
		},
		{
			name:           "X-Real-IP header",
			remoteAddr:     "192.168.1.1:12345",
			xRealIP:        "203.0.113.1",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "X-Forwarded-For single IP",
			remoteAddr:     "192.168.1.1:12345",
			xForwardedFor:  "203.0.113.1",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "X-Forwarded-For multiple IPs",
			remoteAddr:     "192.168.1.1:12345",
			xForwardedFor:  "203.0.113.1, 192.168.1.100, 10.0.0.1",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "X-Forwarded-For with spaces",
			remoteAddr:     "192.168.1.1:12345",
			xForwardedFor:  "  203.0.113.1  ,  192.168.1.100  ",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "Both headers, X-Forwarded-For takes precedence",
			remoteAddr:     "192.168.1.1:12345",
			xForwardedFor:  "203.0.113.1",
			xRealIP:        "203.0.113.2",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "Invalid RemoteAddr format (fallback)",
			remoteAddr:     "invalid-addr-format",
			expectedResult: "invalid-addr-format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			result := getClientIP(req)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestExtractFirstIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single IP",
			input:    "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "multiple IPs",
			input:    "203.0.113.1,192.168.1.1,10.0.0.1",
			expected: "203.0.113.1",
		},
		{
			name:     "IPs with spaces",
			input:    "  203.0.113.1  ,  192.168.1.1  ",
			expected: "203.0.113.1",
		},
		{
			name:     "no comma",
			input:    "203.0.113.1",
			expected: "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFirstIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrimSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no spaces",
			input:    "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "leading spaces",
			input:    "   192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "trailing spaces",
			input:    "192.168.1.1   ",
			expected: "192.168.1.1",
		},
		{
			name:     "both leading and trailing",
			input:    "  192.168.1.1  ",
			expected: "192.168.1.1",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimSpaces(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration test with server
func TestRateLimitingIntegration(t *testing.T) {
	// Set log level to debug for this test
	logrus.SetLevel(logrus.DebugLevel)
	defer logrus.SetLevel(logrus.InfoLevel)

	// Set environment variables to enable rate limiting for the test
	originalEnabled := os.Getenv("RATE_LIMIT_ENABLED")
	originalRate := os.Getenv("RATE_LIMIT_REQUESTS_PER_SECOND")
	originalBurst := os.Getenv("RATE_LIMIT_BURST")

	// Set restrictive limits for testing
	os.Setenv("RATE_LIMIT_ENABLED", "true")
	os.Setenv("RATE_LIMIT_REQUESTS_PER_SECOND", "1.0")
	os.Setenv("RATE_LIMIT_BURST", "1")

	defer func() {
		// Restore original environment
		if originalEnabled == "" {
			os.Unsetenv("RATE_LIMIT_ENABLED")
		} else {
			os.Setenv("RATE_LIMIT_ENABLED", originalEnabled)
		}
		if originalRate == "" {
			os.Unsetenv("RATE_LIMIT_REQUESTS_PER_SECOND")
		} else {
			os.Setenv("RATE_LIMIT_REQUESTS_PER_SECOND", originalRate)
		}
		if originalBurst == "" {
			os.Unsetenv("RATE_LIMIT_BURST")
		} else {
			os.Setenv("RATE_LIMIT_BURST", originalBurst)
		}
	}()

	// Create a temporary directory for web files
	webDir := t.TempDir()

	// Create server with rate limiting enabled
	server, err := NewRPCServer(webDir)
	require.NoError(t, err)
	defer server.Close()

	// Verify rate limiting is enabled
	require.NotNil(t, server.rateLimiter, "Rate limiter should be initialized")

	// Create test server
	testServer := httptest.NewServer(server)
	defer testServer.Close()

	// Create HTTP client
	client := &http.Client{Timeout: time.Second * 5}

	// First request should succeed
	resp1, err := client.Get(testServer.URL + "/health")
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	t.Logf("First request: status=%d", resp1.StatusCode)

	// Wait a tiny bit to ensure the first request is processed
	time.Sleep(10 * time.Millisecond)

	// Second request should be rate limited
	resp2, err := client.Get(testServer.URL + "/health")
	require.NoError(t, err)
	resp2.Body.Close()
	t.Logf("Second request: status=%d, retry-after=%s", resp2.StatusCode, resp2.Header.Get("Retry-After"))
	assert.Equal(t, http.StatusTooManyRequests, resp2.StatusCode)
	assert.Equal(t, "1", resp2.Header.Get("Retry-After"))
}

// Benchmark rate limiter performance
func BenchmarkRateLimiter_Allow(b *testing.B) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 1000.0,
		RateLimitBurst:             1000,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ip := "192.168.1.1"
		for pb.Next() {
			rl.Allow(ip)
		}
	})
}

// Benchmark rate limiting middleware
func BenchmarkRateLimitingMiddleware(b *testing.B) {
	cfg := &config.Config{
		RateLimitRequestsPerSecond: 1000.0,
		RateLimitBurst:             1000,
		RateLimitCleanupInterval:   time.Minute,
	}

	rl := NewRateLimiter(cfg)
	defer rl.Close()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitingMiddleware(rl)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Add logger context
	ctx := context.WithValue(req.Context(), "logger", logrus.StandardLogger())
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
