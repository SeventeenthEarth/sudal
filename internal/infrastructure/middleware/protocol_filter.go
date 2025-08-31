package middleware

import (
	"net/http"
	"strings"

	"github.com/seventeenthearth/sudal/internal/infrastructure/apispec"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
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
func isGRPCRequest(r *http.Request) bool {
	// Check Content-Type header for gRPC protocols
	contentType := r.Header.Get("Content-Type")

	// gRPC over HTTP/2 (standard gRPC)
	if strings.HasPrefix(contentType, "application/grpc") {
		return true
	}

	// gRPC-Web protocol
	if strings.HasPrefix(contentType, "application/grpc-web") {
		return true
	}

	// Connect protocol detection (connect-go)
	// Check for Connect-Protocol-Version header (most reliable way to detect Connect requests)
	if r.Header.Get("Connect-Protocol-Version") != "" {
		return true
	}

	// Connect protocol streaming RPCs use special content types
	if strings.HasPrefix(contentType, "application/connect+") {
		return true
	}

	// Connect protocol unary RPCs can use standard content types with special headers
	// Check if this is a Connect unary RPC by looking for Connect-specific headers
	if r.Method == "POST" && isGRPCServicePath(r.URL.Path) {
		// Connect unary RPCs may use application/json or application/proto
		// but will have other Connect-specific indicators
		if contentType == "application/json" || contentType == "application/proto" ||
			strings.HasPrefix(contentType, "application/json") || 
			strings.HasPrefix(contentType, "application/proto") {
			// Check for Connect-specific headers or patterns
			// Connect clients may send requests with these content types
			// We need to allow them if they're targeting gRPC service paths
			userAgent := r.UserAgent()
			if strings.Contains(userAgent, "connect") {
				return true
			}
			// If it's a POST to a gRPC service path with proto/json content,
			// and has Connect headers, it's a Connect request
			if r.Header.Get("Connect-Accept-Encoding") != "" ||
				r.Header.Get("Connect-Content-Encoding") != "" ||
				r.Header.Get("Connect-Timeout-Ms") != "" {
				return true
			}
		}
	}

	// Check for gRPC-specific headers
	// TE header with "trailers" is required for gRPC over HTTP/2
	if te := r.Header.Get("TE"); strings.Contains(te, "trailers") {
		return true
	}

	// Check for gRPC-Web specific headers
	if r.Header.Get("X-Grpc-Web") != "" {
		return true
	}

	// Check for HTTP/2 with gRPC content type patterns
	// Connect-go may use different content types but still be gRPC
	if r.ProtoMajor == 2 {
		// For HTTP/2, check if it's likely a gRPC request based on other indicators
		userAgent := r.UserAgent()
		if strings.Contains(userAgent, "grpc") || strings.Contains(userAgent, "connect") {
			return true
		}

		// Check for gRPC method patterns (POST to service paths)
		if r.Method == "POST" && isGRPCServicePath(r.URL.Path) {
			// Additional check: if it's HTTP/2 POST to a service path with binary content
			// or no explicit JSON content type, it's likely gRPC
			if !strings.Contains(contentType, "application/json") {
				return true
			}
		}
	}

	return false
}

// isGRPCServicePath checks if the path looks like a gRPC service path
func isGRPCServicePath(path string) bool {
	// gRPC service paths typically follow the pattern: /package.service/method
	// For our services: /health.v1.HealthService/Check, /user.v1.UserService/RegisterUser, etc.
	return strings.Contains(path, ".") && strings.Count(path, "/") >= 2
}

// detectGRPCProtocol detects which gRPC protocol is being used
func detectGRPCProtocol(r *http.Request) string {
	contentType := r.Header.Get("Content-Type")

	// Check for Connect protocol first (most specific)
	if r.Header.Get("Connect-Protocol-Version") != "" {
		return "connect"
	}

	if strings.HasPrefix(contentType, "application/connect+") {
		return "connect-streaming"
	}

	if strings.HasPrefix(contentType, "application/grpc-web") {
		return "grpc-web"
	}

	if strings.HasPrefix(contentType, "application/grpc") {
		return "grpc"
	}

	if r.ProtoMajor == 2 && r.Header.Get("TE") != "" {
		return "grpc-http2"
	}

	// Check for Connect unary with standard content types
	if (contentType == "application/json" || contentType == "application/proto") &&
		(r.Header.Get("Connect-Accept-Encoding") != "" ||
			r.Header.Get("Connect-Content-Encoding") != "" ||
			r.Header.Get("Connect-Timeout-Ms") != "") {
		return "connect-unary"
	}

	return "unknown"
}

// GetGRPCOnlyPaths returns the list of paths that should be restricted to gRPC only
func GetGRPCOnlyPaths() []string {
	return []string{
		apispec.HealthServiceBase, // health.v1.HealthService/Check
		apispec.UserServiceBase,   // user.v1.UserService/* (all methods)
	}
}
