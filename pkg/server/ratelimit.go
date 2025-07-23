package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"goldbox-rpg/pkg/config"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// RateLimiter manages per-IP rate limiting using token bucket algorithm.
// It tracks rate limiters for each client IP and provides automatic cleanup
// of inactive limiters to prevent memory leaks.
type RateLimiter struct {
	// limiters stores rate.Limiter instances keyed by client IP
	limiters map[string]*rateLimiterEntry
	// mu protects concurrent access to the limiters map
	mu sync.RWMutex
	// requestsPerSecond defines the sustained rate of requests allowed
	requestsPerSecond rate.Limit
	// burst defines the maximum number of requests allowed in a burst
	burst int
	// cleanupInterval defines how often to clean up expired limiters
	cleanupInterval time.Duration
	// maxAge defines how long a limiter can be idle before cleanup
	maxAge time.Duration
	// ctx for cancellation of background cleanup
	ctx context.Context
	// cancel function to stop background cleanup
	cancel context.CancelFunc
}

// rateLimiterEntry wraps a rate.Limiter with last access tracking
type rateLimiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// NewRateLimiter creates a new RateLimiter with the specified configuration.
// It starts a background goroutine to periodically clean up unused limiters.
//
// Parameters:
//   - cfg: Configuration containing rate limiting settings
//
// Returns:
//   - *RateLimiter: Configured rate limiter instance
func NewRateLimiter(cfg *config.Config) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		limiters:          make(map[string]*rateLimiterEntry),
		requestsPerSecond: rate.Limit(cfg.RateLimitRequestsPerSecond),
		burst:             cfg.RateLimitBurst,
		cleanupInterval:   cfg.RateLimitCleanupInterval,
		maxAge:            cfg.RateLimitCleanupInterval * 5, // Keep limiters for 5x cleanup interval
		ctx:               ctx,
		cancel:            cancel,
	}

	// Start background cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from the given IP address should be allowed.
// It creates a new rate limiter for unknown IPs and updates the last access time.
//
// Parameters:
//   - ip: Client IP address to check
//
// Returns:
//   - bool: true if the request should be allowed, false if rate limited
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		// Create new rate limiter for this IP
		entry = &rateLimiterEntry{
			limiter:    rate.NewLimiter(rl.requestsPerSecond, rl.burst),
			lastAccess: time.Now(),
		}
		rl.limiters[ip] = entry
	} else {
		// Update last access time
		entry.lastAccess = time.Now()
	}

	return entry.limiter.Allow()
}

// cleanupLoop runs in the background to remove expired rate limiters.
// This prevents memory leaks from clients that stop making requests.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

// cleanup removes rate limiters that haven't been used recently.
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	var removed int

	for ip, entry := range rl.limiters {
		if now.Sub(entry.lastAccess) > rl.maxAge {
			delete(rl.limiters, ip)
			removed++
		}
	}

	if removed > 0 {
		logrus.WithFields(logrus.Fields{
			"removed_limiters": removed,
			"active_limiters":  len(rl.limiters),
		}).Debug("cleaned up expired rate limiters")
	}
}

// Close stops the background cleanup goroutine and releases resources.
func (rl *RateLimiter) Close() {
	if rl.cancel != nil {
		rl.cancel()
	}
}

// Stats returns current statistics about the rate limiter.
type RateLimiterStats struct {
	ActiveLimiters int `json:"active_limiters"`
}

// GetStats returns current rate limiter statistics.
func (rl *RateLimiter) GetStats() RateLimiterStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return RateLimiterStats{
		ActiveLimiters: len(rl.limiters),
	}
}

// RateLimitingMiddleware creates HTTP middleware that enforces rate limiting per client IP.
// Requests that exceed the rate limit receive a 429 Too Many Requests response.
//
// Parameters:
//   - rateLimiter: Configured RateLimiter instance
//
// Returns:
//   - func(http.Handler) http.Handler: Middleware function
func RateLimitingMiddleware(rateLimiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting if rateLimiter is nil (disabled)
			if rateLimiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Get client IP address
			clientIP := getClientIP(r)

			// Check rate limit
			if !rateLimiter.Allow(clientIP) {
				// Get logger from context for consistent logging
				logger := getLoggerFromContext(r.Context())

				logger.WithFields(logrus.Fields{
					"client_ip": clientIP,
					"method":    r.Method,
					"path":      r.URL.Path,
				}).Warn("request rate limited")

				// Return 429 Too Many Requests
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			// Request allowed, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
