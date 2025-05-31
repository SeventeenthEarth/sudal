package log

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	globalLogger *zap.Logger
	once         sync.Once
)

// LogLevel represents the severity level of a log message
type LogLevel string

// Log levels
const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// ParseLevel converts a string level to a LogLevel
func ParseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel // Default to info level
	}
}

// zapLevel converts LogLevel to zapcore.Level
func zapLevel(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Init initializes the global logger with the specified log level
func Init(level LogLevel) {
	once.Do(func() {
		// Create encoder configuration
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// Create JSON encoder
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

		// Create core with stdout output
		core := zapcore.NewCore(
			jsonEncoder,
			zapcore.AddSync(os.Stdout),
			zapLevel(level),
		)

		// Create logger with caller and stacktrace options
		globalLogger = zap.New(
			core,
			zap.AddCaller(),
			zap.AddCallerSkip(1), // Skip the logger wrapper
			zap.AddStacktrace(zapcore.ErrorLevel),
		)

		// Log initialization
		globalLogger.Info("Logger initialized", zap.String("level", string(level)))
	})
}

// GetLogger returns the global logger instance
// If the logger hasn't been initialized, it initializes with InfoLevel
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		Init(InfoLevel)
	}
	return globalLogger
}

// Debug logs a message at debug level
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs a message at info level
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a message at warn level
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs a message at error level
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1)
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// With creates a child logger with the given fields
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.Sync()
}

// FormatError formats an error with its stack trace if available
func FormatError(err error) zap.Field {
	if err == nil {
		return zap.Skip()
	}

	// Check if the error implements stacktracer protocol (from pkg/errors)
	type stackTracer interface {
		StackTrace() []uintptr
	}

	if st, ok := err.(stackTracer); ok {
		return zap.String("error", fmt.Sprintf("%+v", st))
	}

	return zap.Error(err)
}
