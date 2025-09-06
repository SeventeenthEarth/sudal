package middleware

import (
	"net/http"
	"strings"

	"github.com/seventeenthearth/sudal/internal/infrastructure/apispec"
	log "github.com/seventeenthearth/sudal/internal/service/logger"
	"go.uber.org/zap"
)

// ProtocolFilterMiddleware creates a middleware that filters requests by protocol for specified paths
// It blocks HTTP/JSON requests and only allows gRPC and gRPC-Web protocols
func ProtocolFilterMiddleware(grpcOnlyPaths []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this path should be gRPC-only
			if shouldRestrictToGRPC(r.URL.Path, grpcOnlyPaths) {
				if !isGRPCRequest(r) {
					log.WarnContext(r.Context(), "HTTP/JSON request blocked for gRPC-only endpoint",
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.String("content_type", r.Header.Get("Content-Type")),
						zap.String("user_agent", r.UserAgent()),
						zap.String("protocol_detected", detectGRPCProtocol(r)),
					)

					// Return 404 Not Found to hide the existence of the endpoint for non-gRPC clients
					// This is a security best practice for gRPC-only endpoints
					http.NotFound(w, r)
					return
				}

				log.InfoContext(r.Context(), "gRPC request allowed",
					zap.String("path", r.URL.Path),
					zap.String("protocol", detectGRPCProtocol(r)),
				)
			}

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// shouldRestrictToGRPC checks if the given path should be restricted to gRPC only
func shouldRestrictToGRPC(requestPath string, grpcOnlyPaths []string) bool {
	for _, path := range grpcOnlyPaths {
		if strings.HasPrefix(requestPath, path) {
			return true
		}
	}
	return false
}

// isGRPCRequest determines if the incoming request is a gRPC request
// It checks for gRPC-specific headers and content types
// Note: Connect protocol requests are intentionally NOT treated as gRPC
func isGRPCRequest(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")

	// Check for standard gRPC protocols
	if isStandardGRPCContentType(contentType) {
		return true
	}

	// Check for gRPC-specific headers
	if hasGRPCHeaders(r) {
		return true
	}

	return false
}

// isStandardGRPCContentType checks if the content type indicates gRPC or gRPC-Web
func isStandardGRPCContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/grpc") ||
		strings.HasPrefix(contentType, "application/grpc-web")
}

// hasGRPCHeaders checks for gRPC-specific headers
func hasGRPCHeaders(r *http.Request) bool {
	// TE header with "trailers" is required for gRPC over HTTP/2
	if te := r.Header.Get("TE"); strings.Contains(te, "trailers") {
		return true
	}

	// gRPC-Web specific header
	if r.Header.Get("X-Grpc-Web") != "" {
		return true
	}

	return false
}

// detectGRPCProtocol detects which protocol is being used
func detectGRPCProtocol(r *http.Request) string {
	contentType := r.Header.Get("Content-Type")

	// Check for Connect protocol (not allowed on gRPC-only paths)
	if isConnectProtocol(r, contentType) {
		if strings.HasPrefix(contentType, "application/connect+") {
			return "connect-streaming"
		}
		return "connect"
	}

	// Standard gRPC protocols
	if strings.HasPrefix(contentType, "application/grpc-web") {
		return "grpc-web"
	}

	if strings.HasPrefix(contentType, "application/grpc") {
		return "grpc"
	}

	if r.ProtoMajor == 2 && r.Header.Get("TE") != "" {
		return "grpc-http2"
	}

	return "unknown"
}

// isConnectProtocol checks if the request is using Connect protocol
func isConnectProtocol(r *http.Request, contentType string) bool {
	// Check for Connect-Protocol-Version header (most reliable)
	if r.Header.Get("Connect-Protocol-Version") != "" {
		return true
	}

	// Check for Connect streaming content types
	if strings.HasPrefix(contentType, "application/connect+") {
		return true
	}

	// Check for Connect-specific headers with standard content types
	return r.Header.Get("Connect-Accept-Encoding") != "" ||
		r.Header.Get("Connect-Content-Encoding") != "" ||
		r.Header.Get("Connect-Timeout-Ms") != ""
}

// GetGRPCOnlyPaths returns the list of paths that should be restricted to gRPC only
func GetGRPCOnlyPaths() []string {
	return []string{
		apispec.HealthServiceBase, // health.v1.HealthService/Check
		apispec.UserServiceBase,   // user.v1.UserService/* (all methods)
		apispec.QuizServiceBase,   // quiz.v1.QuizService/* (all methods)
	}
}
