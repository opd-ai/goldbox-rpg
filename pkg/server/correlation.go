package server

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CorrelationIDMiddleware adds correlation IDs to all HTTP requests for distributed tracing.
// Correlation IDs persist across service boundaries and help trace requests through the system.
//
// Header precedence:
//  1. X-Correlation-ID (if present in request)
//  2. X-Request-ID (fallback for compatibility)
//  3. Generated UUID (if neither header exists)
//
// The correlation ID is:
//   - Added to response headers (X-Correlation-ID)
//   - Stored in request context
//   - Automatically included in all log entries via context-aware logger
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing correlation ID
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			// Fallback to X-Request-ID for compatibility
			correlationID = r.Header.Get("X-Request-ID")
		}
		if correlationID == "" {
			// Generate new correlation ID
			correlationID = uuid.New().String()
		}

		// Add correlation ID to response headers for client-side tracing
		w.Header().Set("X-Correlation-ID", correlationID)

		// Store correlation ID in context
		ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)
		r = r.WithContext(ctx)

		// Create logger with correlation ID for this request
		logger := logrus.WithField("correlation_id", correlationID)

		// Add additional request metadata to logger
		logger = logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_agent": r.UserAgent(),
			"remote_ip":  getClientIP(r),
		})

		// Store logger in context for use by handlers
		ctx = context.WithValue(ctx, loggerKey, logger)
		r = r.WithContext(ctx)

		logger.Debug("processing request")

		next.ServeHTTP(w, r)
	})
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}
