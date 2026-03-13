package server

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
)

// upgradeConnectionNhooyr upgrades an HTTP request to a WebSocket connection
// using the nhooyr.io/websocket library. This is the default implementation.
//
// Parameters:
//   - w: HTTP response writer for the upgrade
//   - r: HTTP request to upgrade
//
// Returns:
//   - WebSocketConn: Abstracted WebSocket connection interface
//   - error: Error if upgrade fails
func (s *RPCServer) upgradeConnectionNhooyr(w http.ResponseWriter, r *http.Request) (WebSocketConn, error) {
	opts := &websocket.AcceptOptions{
		// Enable compression for bandwidth savings
		CompressionMode: websocket.CompressionContextTakeover,
	}

	// Configure origin validation
	if s.config != nil && s.config.EnableDevMode {
		// In development mode, skip origin verification
		opts.InsecureSkipVerify = true
		logrus.WithField("devMode", true).Debug("WebSocket origin verification disabled (dev mode)")
	} else {
		// In production, use origin patterns from config or environment
		allowedOrigins := s.getAllowedOriginsForNhooyr()
		if len(allowedOrigins) > 0 {
			opts.OriginPatterns = allowedOrigins
		} else {
			// If no origins configured, allow same origin only (nhooyr default)
			logrus.Debug("WebSocket using same-origin policy (no allowed origins configured)")
		}
	}

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		logrus.WithError(err).Error("websocket upgrade failed")
		return nil, err
	}

	// Capture remote address for the connection wrapper
	remoteAddr := r.RemoteAddr
	if remoteAddr == "" {
		remoteAddr = "unknown"
	}

	return NewNhooyrWebSocketConnWithAddr(conn, &nhooyrAddr{addr: remoteAddr}), nil
}

// nhooyrAddr implements net.Addr for nhooyr connections
type nhooyrAddr struct {
	addr string
}

// Network returns the network type
func (a *nhooyrAddr) Network() string {
	return "tcp"
}

// String returns the address string
func (a *nhooyrAddr) String() string {
	return a.addr
}

// getAllowedOriginsForNhooyr returns origin patterns for nhooyr.io/websocket.
// nhooyr uses host patterns (not full URLs) for origin matching.
//
// Returns:
//   - []string: List of allowed host patterns (e.g., "localhost", "*.example.com")
func (s *RPCServer) getAllowedOriginsForNhooyr() []string {
	origins := os.Getenv("WEBSOCKET_ALLOWED_ORIGINS")
	if origins == "" {
		// Fall back to configuration-based origins if available
		if s.config != nil && len(s.config.AllowedOrigins) > 0 {
			return extractHostPatterns(s.config.AllowedOrigins)
		}
		// Allow localhost by default for development
		return []string{"localhost", "127.0.0.1"}
	}

	// Parse origins and extract host patterns
	return extractHostPatterns(strings.Split(origins, ","))
}

// extractHostPatterns converts full URLs to host patterns for nhooyr origin matching.
// For example, "http://localhost:8080" becomes "localhost".
//
// Parameters:
//   - origins: Slice of origin URLs or host patterns
//
// Returns:
//   - []string: Extracted host patterns
func extractHostPatterns(origins []string) []string {
	patterns := make([]string, 0, len(origins))
	seen := make(map[string]bool)

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}

		// If it looks like a URL, extract the host
		var host string
		if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
			// Remove protocol prefix
			host = strings.TrimPrefix(origin, "http://")
			host = strings.TrimPrefix(host, "https://")
			// Remove port if present
			if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
				// Check if this is actually a port (not IPv6)
				if !strings.Contains(host[colonIdx:], "]") {
					host = host[:colonIdx]
				}
			}
			// Remove path if present
			if slashIdx := strings.Index(host, "/"); slashIdx != -1 {
				host = host[:slashIdx]
			}
		} else {
			// Assume it's already a host pattern
			host = origin
			// Remove port if present
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}
		}

		if host != "" && !seen[host] {
			patterns = append(patterns, host)
			seen[host] = true
		}
	}

	return patterns
}

// CheckOriginAllowed tests if a given origin would be allowed for WebSocket connections.
// This method is primarily for testing purposes to verify origin validation behavior.
//
// Parameters:
//   - origin: Full origin URL to check (e.g., "https://example.com:8080")
//
// Returns:
//   - bool: true if origin would be allowed, false otherwise
func (s *RPCServer) CheckOriginAllowed(origin string) bool {
	// In dev mode, all origins are allowed
	if s.config != nil && s.config.EnableDevMode {
		return true
	}

	// Extract host from origin
	host := extractHostFromOrigin(origin)
	if host == "" {
		return false
	}

	// Get allowed origin patterns
	allowedPatterns := s.getAllowedOriginsForNhooyr()

	// Check if host matches any pattern
	for _, pattern := range allowedPatterns {
		if matchOriginPattern(host, pattern) {
			return true
		}
	}

	return false
}

// extractHostFromOrigin extracts the hostname from an origin URL.
func extractHostFromOrigin(origin string) string {
	if origin == "" {
		return ""
	}

	// Remove protocol prefix
	host := strings.TrimPrefix(origin, "http://")
	host = strings.TrimPrefix(host, "https://")

	// Remove port if present
	if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
		// Check if this is not IPv6
		if !strings.Contains(host[colonIdx:], "]") {
			host = host[:colonIdx]
		}
	}

	// Remove path if present
	if slashIdx := strings.Index(host, "/"); slashIdx != -1 {
		host = host[:slashIdx]
	}

	return host
}

// matchOriginPattern checks if a host matches a pattern.
// Supports simple glob patterns with * wildcard.
func matchOriginPattern(host, pattern string) bool {
	// Exact match
	if host == pattern {
		return true
	}

	// Simple wildcard matching (similar to filepath.Match)
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // Remove leading *
		return strings.HasSuffix(host, suffix)
	}

	return false
}
