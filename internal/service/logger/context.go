package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type contextKey string

const (
	// LoggerKey is the key used to store the logger in the context
	LoggerKey contextKey = "logger"
	// TraceIDKey is the key used to store the trace ID in the context
	TraceIDKey contextKey = "trace_id"
	// UserIDKey is the key used to store the user ID in the context
	UserIDKey contextKey = "user_id"
)

// GenerateTraceID generates a random trace ID
func GenerateTraceID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// If random generation fails, use a timestamp-based fallback
		return fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context) context.Context {
	traceID := GenerateTraceID()
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// FromContext retrieves the logger from the context
// If no logger is found, it returns the global logger
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return GetLogger()
	}

	if logger, ok := ctx.Value(LoggerKey).(*zap.Logger); ok && logger != nil {
		return logger
	}

	// Create a new logger with trace ID if available
	logger := GetLogger()
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		logger = logger.With(zap.String("trace_id", traceID))
	}

	// Add user ID if available
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		logger = logger.With(zap.String("user_id", userID))
	}

	return logger
}

// DebugContext logs a debug message with context
func DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

// InfoContext logs an info message with context
func InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Info(msg, fields...)
}

// WarnContext logs a warning message with context
func WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Warn(msg, fields...)
}

// ErrorContext logs an error message with context
func ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Error(msg, fields...)
}

// FatalContext logs a fatal message with context and exits
func FatalContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Fatal(msg, fields...)
}
