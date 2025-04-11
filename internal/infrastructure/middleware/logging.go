package middleware

import (
	"net/http"
	"time"

	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"go.uber.org/zap"
)

// RequestLogger is a middleware that logs HTTP requests
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a new context with trace ID
		ctx := log.WithTraceID(r.Context())

		// Extract trace ID for logging
		traceID, _ := ctx.Value(log.TraceIDKey).(string)

		// Create a logger with request information
		logger := log.GetLogger().With(
			zap.String("trace_id", traceID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)

		// Add logger to context
		ctx = log.WithLogger(ctx, logger)

		// Log the request
		log.InfoContext(ctx, "Request started")

		// Create a response wrapper to capture the status code
		rw := newResponseWriter(w)

		// Call the next handler with the enhanced context
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Calculate request duration
		duration := time.Since(start)

		// Log the response
		log.InfoContext(ctx, "Request completed",
			zap.Int("status", rw.status),
			zap.Duration("duration", duration),
			zap.Int("size", rw.size),
		)
	})
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// newResponseWriter creates a new responseWriter
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK, // Default status is 200 OK
	}
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Now let's prepare for connect-go interceptors
// We'll create a placeholder for now and implement it fully when connect-go is integrated

// ConnectInterceptor returns a connect-go interceptor for logging
// This is a placeholder for future implementation
/*
func ConnectInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()

			// Create a new context with trace ID if not already present
			if _, ok := ctx.Value(log.TraceIDKey).(string); !ok {
				ctx = log.WithTraceID(ctx)
			}

			// Extract trace ID for logging
			traceID, _ := ctx.Value(log.TraceIDKey).(string)

			// Create a logger with request information
			logger := log.GetLogger().With(
				zap.String("trace_id", traceID),
				zap.String("procedure", req.Spec().Procedure),
			)

			// Add logger to context
			ctx = log.WithLogger(ctx, logger)

			// Log the request
			log.InfoContext(ctx, "RPC request started")

			// Call the next handler
			res, err := next(ctx, req)

			// Calculate request duration
			duration := time.Since(start)

			// Determine status code
			statusCode := connect.CodeOf(err)

			// Log fields
			fields := []zap.Field{
				zap.String("status", statusCode.String()),
				zap.Duration("duration", duration),
			}

			// Add error information if present
			if err != nil {
				fields = append(fields, log.FormatError(err))
			}

			// Log the response
			log.InfoContext(ctx, "RPC request completed", fields...)

			return res, err
		}
	}
}
*/
