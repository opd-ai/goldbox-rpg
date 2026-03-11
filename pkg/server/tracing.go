package server

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// TraceContext holds tracing information for a request
type TraceContext struct {
	CorrelationID string
	SessionID     string
	StartTime     time.Time
	Method        string
	Path          string
}

// NewTraceContext creates a new trace context from an HTTP request context
func NewTraceContext(ctx context.Context) *TraceContext {
	return &TraceContext{
		CorrelationID: GetCorrelationID(ctx),
		SessionID:     GetSessionID(ctx),
		StartTime:     time.Now(),
	}
}

// GetLogger retrieves the context-aware logger from the request context.
// The logger automatically includes correlation_id and other request metadata.
// Falls back to standard logger if not found in context.
func GetLogger(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value(loggerKey).(*logrus.Entry); ok {
		return logger
	}
	// Fallback to standard logger
	return logrus.NewEntry(logrus.StandardLogger())
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LogWithTrace logs a message with full trace context
func LogWithTrace(ctx context.Context, level logrus.Level, message string, fields logrus.Fields) {
	logger := GetLogger(ctx)

	// Add trace context to fields
	trace := NewTraceContext(ctx)
	if trace.CorrelationID != "" && fields["correlation_id"] == nil {
		fields["correlation_id"] = trace.CorrelationID
	}
	if trace.SessionID != "" && fields["session_id"] == nil {
		fields["session_id"] = trace.SessionID
	}

	logger.WithFields(fields).Log(level, message)
}

// TraceOperation logs the start and end of an operation with elapsed time
func TraceOperation(ctx context.Context, operation string, fn func() error) error {
	logger := GetLogger(ctx)
	start := time.Now()

	logger.WithFields(logrus.Fields{
		"operation": operation,
	}).Debug("operation started")

	err := fn()
	elapsed := time.Since(start)

	fields := logrus.Fields{
		"operation":    operation,
		"elapsed_ms":   elapsed.Milliseconds(),
		"elapsed_nano": elapsed.Nanoseconds(),
	}

	if err != nil {
		fields["error"] = err.Error()
		logger.WithFields(fields).Error("operation failed")
	} else {
		logger.WithFields(fields).Debug("operation completed")
	}

	return err
}

// AddTraceFields adds common trace fields to a logger entry
func AddTraceFields(ctx context.Context, logger *logrus.Entry) *logrus.Entry {
	trace := NewTraceContext(ctx)

	fields := logrus.Fields{}
	if trace.CorrelationID != "" {
		fields["correlation_id"] = trace.CorrelationID
	}
	if trace.SessionID != "" {
		fields["session_id"] = trace.SessionID
	}

	if len(fields) > 0 {
		return logger.WithFields(fields)
	}
	return logger
}

// MeasureDuration returns a function that logs the elapsed time when called
func MeasureDuration(ctx context.Context, operation string) func() {
	logger := GetLogger(ctx)
	start := time.Now()

	return func() {
		elapsed := time.Since(start)
		logger.WithFields(logrus.Fields{
			"operation":  operation,
			"elapsed_ms": elapsed.Milliseconds(),
		}).Debug("operation duration")
	}
}
