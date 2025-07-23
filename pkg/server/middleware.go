package server

import (
	"context"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request correlation IDs
	RequestIDKey ContextKey = "request_id"
	// SessionIDKey is the context key for session IDs
	SessionIDKey ContextKey = "session_id"
)

// RequestIDMiddleware adds request correlation IDs to all HTTP requests
// If a request already has an X-Request-ID header, it uses that value.
// Otherwise, it generates a new UUID for the request.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = uuid.New().String()
		}

		// Add request ID to response headers for tracing
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add request ID to all log entries for this request
		logger := logrus.WithField("request_id", requestID)

		// Store logger in context for use by handlers
		ctx = context.WithValue(ctx, "logger", logger)
		r = r.WithContext(ctx)

		logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_agent": r.UserAgent(),
			"remote_ip":  getClientIP(r),
		}).Debug("processing request")

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware provides structured logging for HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get logger from context (set by RequestIDMiddleware)
		logger := getLoggerFromContext(r.Context())

		// Create a response writer wrapper to capture status code
		wrapper := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		// Log the completed request
		logger.WithFields(logrus.Fields{
			"status_code": wrapper.statusCode,
			"method":      r.Method,
			"path":        r.URL.Path,
		}).Info("request completed")
	})
}

// RecoveryMiddleware recovers from panics and logs them with request context
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := getLoggerFromContext(r.Context())

				logger.WithFields(logrus.Fields{
					"panic":  err,
					"method": r.Method,
					"path":   r.URL.Path,
				}).Error("recovered from panic")

				// Return 500 error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing headers
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

// getLoggerFromContext retrieves the logger from the request context
func getLoggerFromContext(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value("logger").(*logrus.Entry); ok {
		return logger
	}
	// Fallback to standard logger
	return logrus.NewEntry(logrus.StandardLogger())
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetSessionID retrieves the session ID from the context
func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check for forwarded headers first
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if firstIP := extractFirstIP(ip); firstIP != "" {
			return firstIP
		}
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Extract just the IP from RemoteAddr (which includes port)
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	// Fallback if SplitHostPort fails
	return r.RemoteAddr
}

// extractFirstIP extracts the first IP from a comma-separated list
func extractFirstIP(ips string) string {
	if ips == "" {
		return ""
	}
	// Split by comma and take the first non-empty entry
	for i := 0; i < len(ips); i++ {
		if ips[i] == ',' {
			return trimSpaces(ips[:i])
		}
	}
	return trimSpaces(ips)
}

// trimSpaces removes leading and trailing spaces
func trimSpaces(s string) string {
	start := 0
	end := len(s)

	// Remove leading spaces
	for start < end && s[start] == ' ' {
		start++
	}

	// Remove trailing spaces
	for end > start && s[end-1] == ' ' {
		end--
	}

	return s[start:end]
}

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false // No origins allowed
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
